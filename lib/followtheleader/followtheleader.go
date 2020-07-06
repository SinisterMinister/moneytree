package followtheleader

import (
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
	return &Processor{trader, market, struct{}{}}
}

func (p *Processor) Process(stop <-chan bool) (done <-chan bool, err error) {
	// Build closed return channel
	ret := make(chan bool)
	done = ret

	// Run the process
	go p.run(ret)

	return done, err
}

func (p *Processor) run(done chan<- bool) {
	var (
		orderPair      *orderpair.OrderPair
		upwardTrending bool
	)

	// Follow the leader if there is one
	if p.leader != nil {
		upwardTrending = p.leader.SecondOrder().Request().Side() == order.Buy
	} else {
		upwardTrending = p.isMarketUpwardTrending()
	}

	// Build the order pair
	if upwardTrending {
		orderPair, err := p.buildUpwardTrendingPair()
	} else {
		orderPair, err := p.buildDownwardTrendingPair()
	}

	// Execute the order
	orderDone := orderPair.Execute()

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
			if orderPair.FirstOrder().Filled().GreaterThan(decimal.Zero) {
				fallthrough
			}
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

				// If we've filled anything, fall through to make leader
				if orderPair.SecondOrder().Filled().GreaterThan(decimal.Zero) {
					fallthrough
				}

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
		return nil, err
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

	// Get wallets
	baseWallet := market.BaseCurrency().Wallet()
	log.WithFields(
		log.F("total", baseWallet.Total()),
		log.F("available", baseWallet.Available()),
	).Infof("wallet for %s", baseWallet.Currency().Name())

	quoteWallet := market.QuoteCurrency().Wallet()
	log.WithFields(
		log.F("total", quoteWallet.Total()),
		log.F("available", quoteWallet.Available()),
	).Infof("wallet for %s", quoteWallet.Currency().Name())

	// Grab the current ticker
	ticker, err := p.market.Ticker()
	if err != nil {
		log.WithError(err).Error("failed to get ticker")
		close(done)
		return
	}
}

func (p *Processor) buildDownwardTrendingPair() (*orderpair.OrderPair, error) {

}
