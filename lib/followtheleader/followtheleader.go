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

	// Build closed return channel
	ret := make(chan bool)
	close(ret)

	return ret, err
}
