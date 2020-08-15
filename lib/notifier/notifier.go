package notifier

import (
	"github.com/go-playground/log/v7"
	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types"
)

type Notifier interface {
	Receive() <-chan bool
}

func NewPriceAboveNotifier(stop <-chan bool, market types.Market, price decimal.Decimal) *PriceAbove {
	n := &PriceAbove{
		observerChannel:  make(chan bool),
		triggered:        false,
		incomingRequests: make(chan chan chan bool),
		market:           market,
		price:            price,
	}
	go n.runner(stop)
	return n
}

type PriceAbove struct {
	observerChannel  chan bool
	triggered        bool
	incomingRequests chan chan chan bool
	market           types.Market
	price            decimal.Decimal
}

func (n *PriceAbove) Receive() <-chan bool {
	ch := make(chan chan bool)

	// Register the receiver
	n.incomingRequests <- ch

	return <-ch
}

func (n *PriceAbove) runner(stop <-chan bool) {
	// Start a ticker stream
	tickerStop := make(chan bool)
	stream := n.market.TickerStream(tickerStop)

	for {
		select {

		// Bail out on stop
		case <-stop:
			// Stop the stream
			close(tickerStop)
			return

		// Handle ticker stream
		case ticker := <-stream:
			if ticker.Price().GreaterThan(n.price) {
				log.Debugf("notifying that price went above %s", n.price.String())
				// Stop the stream
				close(tickerStop)

				// Notify all the observers
				close(n.observerChannel)

				// Bail out
				return
			}

		case ch := <-n.incomingRequests:
			ch <- n.observerChannel
		}
	}
}

func NewPriceBelowNotifier(stop <-chan bool, market types.Market, price decimal.Decimal) *PriceBelow {
	n := &PriceBelow{
		observerChannel:  make(chan bool),
		triggered:        false,
		incomingRequests: make(chan chan chan bool),
		market:           market,
		price:            price,
	}
	go n.runner(stop)
	return n
}

type PriceBelow struct {
	observerChannel  chan bool
	triggered        bool
	incomingRequests chan chan chan bool
	market           types.Market
	price            decimal.Decimal
}

func (n *PriceBelow) Receive() <-chan bool {
	ch := make(chan chan bool)

	// Register the receiver
	n.incomingRequests <- ch

	return <-ch
}

func (n *PriceBelow) runner(stop <-chan bool) {
	// Start a ticker stream
	tickerStop := make(chan bool)
	stream := n.market.TickerStream(tickerStop)

	for {
		select {

		// Bail out on stop
		case <-stop:
			// Stop the stream
			close(tickerStop)
			return

		// Handle ticker stream
		case ticker := <-stream:
			if ticker.Price().LessThan(n.price) {
				log.Debugf("notifying that price fell below %s", n.price.String())
				// Stop the stream
				close(tickerStop)

				// Notify all the observers
				close(n.observerChannel)

				// Bail out
				return
			}

		case ch := <-n.incomingRequests:
			ch <- n.observerChannel
		}
	}
}
