package followtheleader

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/go-playground/log/v7"
	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/candle"
	"github.com/sinisterminister/currencytrader/types/fees"
	"github.com/sinisterminister/currencytrader/types/order"
	"github.com/sinisterminister/moneytree/lib/orderpair"
	"github.com/sinisterminister/moneytree/lib/trix"
	"github.com/spf13/viper"
)

type Processor struct {
	db     *sql.DB
	trader types.Trader
	market types.Market
	leader *orderpair.OrderPair
}

func New(db *sql.DB, trader types.Trader, market types.Market) *Processor {
	return &Processor{db, trader, market, nil}
}

func (p *Processor) Process(stop <-chan bool) (done <-chan bool, err error) {
	// Build closed return channel
	ret := make(chan bool)
	done = ret

	// Run the process
	go p.run(stop, ret)
	return
}

func (p *Processor) run(stop <-chan bool, done chan bool) {
	// Build the order pair
	orderPair, err := p.buildOrderPair()
	if err != nil {
		log.WithError(err).Error("unable to build order pair")
		close(done)
	}

	// Execute the order pair
	err = p.executeOrderPair(stop, orderPair)
	if err != nil {
		log.WithError(err).Error("error while executing order pair")
	}

	// Close the done channel
	close(done)
}

func (p *Processor) buildOrderPair() (orderPair *orderpair.OrderPair, err error) {
	var upwardTrending bool

	// Follow the leader if there is one
	if p.leader != nil && !p.leader.SecondOrder().IsDone() {
		upwardTrending = p.leader.SecondOrder().Request().Side() != order.Sell
	} else {
		upwardTrending, err = p.isMarketUpwardTrending()
		if err != nil {
			return nil, fmt.Errorf("could not get trend data: %w", err)
		}
	}
	log.WithFields(log.F("market", p.market.Name()), log.F("upward", upwardTrending)).Info("got market trend direction")

	// Build the order pair
	if upwardTrending {
		orderPair, err = p.buildUpwardTrendingPair()
	} else {
		orderPair, err = p.buildDownwardTrendingPair()
	}
	if err != nil {
		return
	}
	log.WithFields(log.F("first", orderPair.FirstRequest().ToDTO()), log.F("second", orderPair.SecondRequest().ToDTO())).Debug("order pair created")
	return
}

func (p *Processor) executeOrderPair(stop <-chan bool, orderPair *orderpair.OrderPair) (err error) {
	// Execute the order
	orderDone := orderPair.Execute(stop)
	log.Info("order pair execution started")

	// Create timer to bail on stale orders
	timer := time.NewTimer(viper.GetDuration("followtheleader.orderTTL"))

	// If there's a leader, we wait for it to complete instead of the timer
	if p.leader != nil && !p.leader.SecondOrder().IsDone() {
		done := p.leader.SecondOrder().Done()
		select {
		case <-orderDone:
			// Next order
			return
		case <-done:
			log.Info("second order completed for leader")
			p.leader = nil
		}
	}

	// Wait for the order to be complete or for it to timeout
	select {
	case <-timer.C:
		// This order has gone stale but may not need to be made leader
		p.rotateLeader(orderPair)
	case <-orderDone:
		// Order has complete. Nothing to do
	}
	return
}

func (p *Processor) isMarketUpwardTrending() (bool, error) {
	// Get trix values
	candles, err := p.market.Candles(candle.FiveMinutes, time.Now().Add(-4*time.Hour), time.Now())
	if err != nil {
		return false, err
	}

	// Build price slice
	prices := []float64{}
	for _, candle := range candles {
		price, _ := candle.Close().Float64()
		prices = append(prices, price)
	}

	// Get trix indicator
	ma, osc := trix.GetTrixIndicator(prices)
	log.WithFields(
		log.F("market", p.market.Name()),
		log.F("trix", ma),
		log.F("osc", osc),
	).Info("trix value computed")

	return osc > 0, nil
}

func (p *Processor) buildUpwardTrendingPair() (*orderpair.OrderPair, error) {
	// Determine prices using the spread
	spread, err := p.getSpread()
	if err != nil {
		return nil, err
	}
	spread = decimal.NewFromFloat(1).Add(spread)

	// Get the ticker for the current prices
	ticker, err := p.market.Ticker()
	if err != nil {
		return nil, err
	}

	// Get the size
	size, err := p.getSize(ticker)
	if err != nil {
		return nil, err
	}

	// Set the bid price to price + 1 increment
	// bidPrice := ticker.Bid()
	bidPrice := ticker.Bid().Add(p.market.QuoteCurrency().Increment().Mul(decimal.NewFromFloat(2))).Round(int32(p.market.QuoteCurrency().Precision()))
	bidSize := size.Round(int32(p.market.BaseCurrency().Precision()))
	askPrice := bidPrice.Mul(spread).Round(int32(p.market.QuoteCurrency().Precision()))
	askSize := size.Div(decimal.NewFromFloat(2)).Mul(bidPrice).Div(askPrice).Add(size.Div(decimal.NewFromFloat(2))).Round(int32(p.market.BaseCurrency().Precision()))
	bidReq := order.NewRequest(p.market, order.Limit, order.Buy, bidSize, bidPrice)
	askReq := order.NewRequest(p.market, order.Limit, order.Sell, askSize, askPrice)

	// Create order pair
	op, err := orderpair.New(p.db, p.trader, p.market, bidReq, askReq)
	if err != nil {
		return nil, fmt.Errorf("could not create order pair: %w", err)
	}
	return op, nil
}

func (p *Processor) buildDownwardTrendingPair() (*orderpair.OrderPair, error) {
	// Determine prices using the spread
	spread, err := p.getSpread()
	if err != nil {
		return nil, err
	}

	// Prepare the spread to be applied
	spread = decimal.NewFromFloat(1).Add(spread)

	// Get the ticker for the current prices
	ticker, err := p.market.Ticker()
	if err != nil {
		return nil, err
	}

	// Get the size
	size, err := p.getSize(ticker)
	if err != nil {
		return nil, err
	}

	// Set the ask price to price - 1 increment
	// askPrice := ticker.Ask()
	askPrice := ticker.Ask().Sub(p.market.QuoteCurrency().Increment().Mul(decimal.NewFromFloat(2))).Round(int32(p.market.QuoteCurrency().Precision()))
	bidSize := size.Round(int32(p.market.BaseCurrency().Precision()))
	bidPrice := askPrice.Sub(askPrice.Mul(spread).Sub(askPrice)).Round(int32(p.market.QuoteCurrency().Precision()))
	askSize := size.Div(decimal.NewFromFloat(2)).Mul(bidPrice).Div(askPrice).Add(size.Div(decimal.NewFromFloat(2))).Round(int32(p.market.BaseCurrency().Precision()))
	askReq := order.NewRequest(p.market, order.Limit, order.Sell, askSize, askPrice)
	bidReq := order.NewRequest(p.market, order.Limit, order.Buy, bidSize, bidPrice)
	log.WithFields(log.F("askSize", askSize.String()), log.F("askPrice", askPrice.String()), log.F("bidSize", bidSize.String()), log.F("bidPrice", bidPrice.String())).Info("downward trending order sizes")

	// Create order pair
	op, err := orderpair.New(p.db, p.trader, p.market, askReq, bidReq)
	if err != nil {
		return nil, fmt.Errorf("could not create order pair: %w", err)
	}
	return op, nil
}

func (p *Processor) getSize(ticker types.Ticker) (decimal.Decimal, error) {
	// Determine order size from average volume
	size, err := p.market.AverageTradeVolume()
	if err != nil {
		return size, err
	}

	// Get wallets
	baseWallet := p.market.BaseCurrency().Wallet()
	quoteWallet := p.market.QuoteCurrency().Wallet()

	// Get the maximum trade size by wallet
	baseMax := baseWallet.Available().Div(decimal.NewFromFloat(viper.GetFloat64("followtheleader.maxTradesFundsRatio")))
	quoteMax := quoteWallet.Available().Div(decimal.NewFromFloat(viper.GetFloat64("followtheleader.maxTradesFundsRatio"))).Div(ticker.Bid())

	// Normalize the size to available funds
	if size == decimal.Zero {
		size = decimal.Min(baseMax, quoteMax)
	}
	return decimal.Min(size, baseMax, quoteMax), nil
}

func (p *Processor) getSpread() (decimal.Decimal, error) {
	// Get the fees
	f, err := p.trader.AccountSvc().Fees()
	if err != nil {
		log.WithError(err).Error("failed to get fees")
		return decimal.Zero, err
	}
	if viper.GetBool("disableFees") == true {
		f = fees.ZeroFee()
	}

	// Set the profit target
	target := decimal.NewFromFloat(viper.GetFloat64("followtheleader.targetReturn"))

	// Add the taker fees twice for the two orders
	rate := f.TakerRate().Add(f.TakerRate())

	// Calculate spread
	spread := target.Add(rate)

	return spread, nil
}

func (p *Processor) rotateLeader(op *orderpair.OrderPair) {
	log.Info("rotating the leader")
	// If first order is still open, cancel it
	if !op.FirstOrder().IsDone() {
		err := p.trader.OrderSvc().CancelOrder(op.FirstOrder())
		if err != nil {
			log.WithError(err).Warn("could not cancel stalled order")
		}
		// Give the order some time to process
		wait := time.NewTimer(viper.GetDuration("followtheleader.waitAfterCancelStalledPair"))
		<-wait.C
	}

	if op.SecondOrder() == nil {
		log.Warn("second order wasn't executed")
		return
	}

	// If the first order was filled at all and the second order is still open, it's leader
	if !op.FirstOrder().Filled().Equal(decimal.Zero) && !op.SecondOrder().IsDone() {
		p.leader = op
	}
}
