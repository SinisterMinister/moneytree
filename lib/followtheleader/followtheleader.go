package followtheleader

import (
	"time"

	"github.com/go-playground/log/v7"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/candle"
	"github.com/sinisterminister/moneytree/lib/trix"
)

type Processor struct{}

func (p *Processor) ProcessMarket(stop <-chan bool, market types.Market) (done <-chan bool, err error) {
	// Get trix values
	candles, err := market.Candles(candle.OneMinute, time.Now().Add(-60*time.Minute), time.Now())
	if err != nil {
		return done, err
	}

	prices := []float64{}

	// Build price slice
	for _, candle := range candles {
		price, _ := candle.Close().Float64()
		prices = append(prices, price)
	}

	// Get trix indicator
	ma, osc := trix.GetTrixIndicator(prices)
	log.WithFields(log.F("market", market.Name()), log.F("trix", ma), log.F("osc", osc)).Info("trix value computed")

	// Get wallets

	baseWallet := market.BaseCurrency().Wallet()
	quoteWallet := market.QuoteCurrency().Wallet()

	log.WithFields(log.F("total", baseWallet.Total()), log.F("available", baseWallet.Available())).Infof("wallet for %s", baseWallet.Currency().Name())
	log.WithFields(log.F("total", quoteWallet.Total()), log.F("available", quoteWallet.Available())).Infof("wallet for %s", quoteWallet.Currency().Name())
	// Build closed return channel
	ret := make(chan bool)
	close(ret)

	return ret, err
}
