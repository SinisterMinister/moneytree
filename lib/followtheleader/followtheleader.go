package followtheleader

import (
	"fmt"
	"time"

	"github.com/go-playground/log/v7"
	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/candle"
	"github.com/sinisterminister/currencytrader/types/order"
	"github.com/sinisterminister/moneytree/lib/orderpair"
	"github.com/sinisterminister/moneytree/lib/trix"
	"github.com/spf13/viper"
)

type Processor struct {
	trader types.Trader
	market types.Market
	leader *orderpair.OrderPair
}

func New(trader types.Trader, market types.Market) *Processor {
	return &Processor{trader, market, nil}
}

func (p *Processor) Process(stop <-chan bool) (done <-chan bool, err error) {
	// Build closed return channel
	ret := make(chan bool)
	done = ret

	// Run the process
	go p.run(stop, ret)

	return done, err
}

func (p *Processor) run(stop <-chan bool, done chan<- bool) {
	var (
		orderPair      *orderpair.OrderPair
		upwardTrending bool
		err            error
	)

	// Follow the leader if there is one
	if p.leader != nil {
		upwardTrending = p.leader.SecondOrder().Request().Side() == order.Buy
	} else {
		upwardTrending, err = p.isMarketUpwardTrending()
		if err != nil {
			log.WithError(err).Error("could not get trend data")
			// Close the done channel
			close(done)
			return
		}
	}

	// Build the order pair
	if upwardTrending {
		orderPair, err = p.buildUpwardTrendingPair()
	} else {
		orderPair, err = p.buildDownwardTrendingPair()
	}
	if err != nil {
		log.WithError(err).Error("could not build pair")
		// Close the done channel
		close(done)
		return
	}

	// Execute the order
	orderDone := orderPair.Execute(stop)

	// Create timer to bail on stale orders
	timer := time.NewTimer(viper.GetDuration("followtheleader.markOrderStaleAfter"))

	// Wait for orders to process or timeout
	select {
	case <-timer.C:
		switch orderPair.FirstOrder().Status() {
		case order.Pending:
			err := orderPair.Cancel()
			if err != nil {
				log.WithError(err).Error("could not cancel stalled order")
			}

		// Not sure what's up with this order so fall through to use filled
		case order.Unknown:
			fallthrough
		case order.Updated:
			// This order is partially filled
			if orderPair.FirstOrder().Filled().Equal(decimal.Zero) {
				break
			}

			fallthrough
		case order.Partial:
			// Cancel the first order and let the pair self-heal
			err := p.trader.OrderSvc().CancelOrder(orderPair.FirstOrder())
			if err != nil {
				log.WithError(err).Error("could not cancel stalled order")
			}
			// Give the order some time to process
			wait := time.NewTimer(viper.GetDuration("followtheleader.waitAfterCancelStalledPair"))
			<-wait.C
			fallthrough
		case order.Canceled:
			fallthrough
		case order.Expired:
			fallthrough
		case order.Rejected:
			fallthrough
		case order.Filled:
			switch orderPair.SecondOrder().Status() {
			// Assume that this order should be ignored
			case order.Canceled:
			case order.Expired:
			case order.Rejected:

			// Move on if second order is also filled
			case order.Filled:

			// Not really sure what's going on so fall through to cancel just in case
			case order.Unknown:
				fallthrough

			// Not really sure what's going on so cancel just in case
			case order.Updated:
				err := p.trader.OrderSvc().CancelOrder(orderPair.SecondOrder())
				if err != nil {
					log.WithError(err).Error("could not cancel second order")
				}

				// break if we haven't filled anything, get out
				if orderPair.SecondOrder().Filled().Equal(decimal.Zero) {
					break
				}
				fallthrough
			// Partial case so fall through to make leader
			case order.Pending:
				fallthrough

			// Mark as leader
			case order.Partial:
				p.leader = orderPair
			}
		}
	case <-orderDone:
		// Order has complete. Nothing to do
	}

	// Close the done channel
	close(done)
}

func (p *Processor) isMarketUpwardTrending() (bool, error) {
	// Get trix values
	candles, err := p.market.Candles(candle.OneMinute, time.Now().Add(-60*time.Minute), time.Now())
	if err != nil {
		log.WithError(err).Error("unable to fetch candle data")
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

func (p *Processor) getSpread() (decimal.Decimal, error) {
	// Get the fees
	fees, err := p.trader.AccountSvc().Fees()
	if err != nil {
		log.WithError(err).Error("failed to get fees")
		return decimal.Zero, err
	}

	// Set the profit target
	target := decimal.NewFromFloat(0.005)

	// Add the taker fees twice for the two orders
	rate := fees.TakerRate().Add(fees.TakerRate())

	// Calculate spread
	spread := target.Add(rate)

	return spread, nil
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
	bidPrice := ticker.Bid().Add(p.market.QuoteCurrency().Increment())
	askPrice := bidPrice.Mul(spread)
	askSize := size.Mul(bidPrice).Div(askPrice).Add(size)
	bidReq := order.NewRequest(p.market, order.Limit, order.Buy, size, bidPrice)
	askReq := order.NewRequest(p.market, order.Limit, order.Sell, askSize, askPrice)

	// Create order pair
	op, err := orderpair.New(p.trader, p.market, bidReq, askReq)
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
	spread = decimal.NewFromFloat(1).Sub(spread)

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
	askPrice := ticker.Ask().Sub(p.market.QuoteCurrency().Increment())
	bidPrice := askPrice.Mul(spread)
	askReq := order.NewRequest(p.market, order.Limit, order.Sell, size, askPrice)
	bidReq := order.NewRequest(p.market, order.Limit, order.Buy, size, bidPrice)

	// Create order pair
	op, err := orderpair.New(p.trader, p.market, askReq, bidReq)
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
