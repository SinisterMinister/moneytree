package orderpair

import (
	"time"

	"github.com/go-playground/log"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/candle"
	"github.com/sinisterminister/moneytree/lib/trix"
)

type OrderPair struct {
	market types.Market

	firstOrderDTO types.OrderDTO
	firstOrder    types.Order

	secondOrderDTO types.OrderDTO
	secondOrder    types.Order

	done chan bool
}

func New(market types.Market) (orderPair *OrderPair) {
	orderPair = &OrderPair{market: market, done: make(chan bool)}
	return orderPair
}

func (o *OrderPair) Execute() <-chan bool {
	go o.executeWorkflow()
	return o.done
}

func (o *OrderPair) executeWorkflow() {
	// Construct DTOs

	// Place first order

	// Wait for order to complete, bailing if it misses

	// If order missed, send false over done channel before closing

	// Place second order

	// Wait for it to complete, timing out after a configured amount of time

	// If timed out, send false over done channel before closing

	// If successful, send true over done channel before closing
}

func (o *OrderPair) constructOrderDTOs() {

}

func (o *OrderPair) isMarketUpwardTrending() (bool, error) {
	// Get trix values
	candles, err := o.market.Candles(candle.OneMinute, time.Now().Add(-60*time.Minute), time.Now())
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
		log.F("market", o.market.Name()),
		log.F("trix", ma),
		log.F("osc", osc),
	).Info("trix value computed")

	return osc > 0, nil
}
