package followtheleader

import (
	"database/sql"
	"time"

	"github.com/go-playground/log/v7"
	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/candle"
	"github.com/sinisterminister/currencytrader/types/fees"
	"github.com/sinisterminister/currencytrader/types/order"
	"github.com/sinisterminister/moneytree/lib/orderpair"
	"github.com/sinisterminister/moneytree/lib/state"
	"github.com/sinisterminister/moneytree/lib/trix"
	"github.com/spf13/viper"
)

type Processor struct {
	db           *sql.DB
	trader       types.Trader
	market       types.Market
	stateManager *state.Manager
	stop         <-chan bool
}

func New(db *sql.DB, trader types.Trader, market types.Market, stop <-chan bool) *Processor {
	manager := state.NewManager(stop)
	return &Processor{db, trader, market, manager, stop}
}

func (p *Processor) Recover() {
	log.Info("recovering open order pairs")

	// Load running order pairs
	pairs, err := orderpair.LoadOpenPairs(p.db, p.trader, p.market)
	if err != nil {
		log.WithError(err).Errorf("could not load order pairs: %w", err)
		return
	}

	open := []*orderpair.OrderPair{}
	for _, pair := range pairs {
		// Cancel pairs missed pairs that may be open
		if pair.FirstOrder() != nil && pair.FirstOrder().Filled().Equal(decimal.Zero) {
			err = pair.Cancel()
			if err != nil {
				log.WithField("pair", pair.ToDAO()).WithError(err).Errorf("could not cancel order pair: %w", err)
			}
		}

		// Drop closed pairs
		if !pair.IsDone() {
			pair.Execute(p.stop)
			open = append(open, pair)
		}
	}

	// Load most recent pair to see if it should set the direction
	pair, err := orderpair.LoadMostRecentPair(p.db, p.trader, p.market)
	if err != nil {
		log.WithError(err).Error("could not load most recent pair")
		return
	}

	// If this isn't complete, it becomes our initial state
	if !pair.IsDone() {
		log.Debugf("last pair still open, resuming")
		if pair.BuyRequest().Side() == order.Buy {
			p.stateManager.Resume(&UpwardTrending{processor: p, orderPair: pair})
		} else {
			p.stateManager.Resume(&DownwardTrending{processor: p, orderPair: pair})
		}
	}
}

func (p *Processor) Process() (done <-chan bool, err error) {
	// Build closed return channel
	ret := make(chan bool)
	done = ret

	// Run the process
	go p.process(ret)
	return
}

func (p *Processor) process(done chan bool) {
	// If there's no defined initial state, create one
	if p.stateManager.CurrentState() == nil {
		// Determine if the market is upward trending
		upward, err := p.isMarketUpwardTrending()
		if err != nil {
			log.WithError(err).Error("unable to determine market direction")
			close(done)
		}

		// If the market is upward trending, set the state as such
		if upward {
			p.stateManager.TransitionTo(&UpwardTrending{processor: p})
		} else {
			p.stateManager.TransitionTo(&DownwardTrending{processor: p})
		}
	}

	// Wait for the state to process
	log.Infof("waiting for %T to process", p.stateManager.CurrentState())
	<-p.stateManager.CurrentState().Done()

	// Close the done channel
	close(done)
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
