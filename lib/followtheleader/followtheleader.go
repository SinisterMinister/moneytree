package followtheleader

import (
	"time"

	"github.com/go-playground/log/v7"
	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/candle"
	"github.com/sinisterminister/moneytree/lib/trix"
)

type Processor struct {
	trader types.Trader
	market types.Market
	leader struct{}
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
	market := p.market

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

	// Get the fees
	fees, err := p.trader.AccountSvc().Fees()
	if err != nil {
		log.WithError(err).Error("failed to get fees")
		close(done)
		return
	}

	// Set the profit target
	target := decimal.NewFromFloat(0.005)

	// Add the taker fees twice for the two orders
	fees := fees.TakerRate().Add(fees.TakerRate())

	// Calculate spread
	spread := target.Add(fees)

	log.WithFields(
		log.F("maker", fees.MakerRate()),
		log.F("taker", fees.TakerRate()),
		log.F("volume", fees.Volume()),
		log.F("target", target),
		log.F("spread", spread),
	).Info("fees and spread")

	// Calculate prices and amounts

	// Build order pair

	// Execute order pair

	// Wait for orders to process or timeout

	// Store stuck orders
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
