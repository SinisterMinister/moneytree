package notifier

import (
	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types"
)

type Notifier interface {
	Receive() <-chan bool
}

func NewPriceAboveNotifier(stop <-chan bool, market types.Market, price decimal.Decimal) *PriceAbove {
	n := &PriceAbove{
		observers:         []chan bool{},
		triggered:         false,
		incomingReceivers: make(chan chan bool),
	}
	n.runner(stop)
	return n
}

type PriceAbove struct {
	observers         []chan bool
	triggered         bool
	incomingReceivers chan chan bool
	market            types.Market
	price             decimal.Decimal
}

func (n *PriceAbove) Receive() <-chan bool {
	ch := make(chan bool)

	// Register the receiver
	n.incomingReceivers <- ch

	return ch
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
				// Stop the stream
				close(tickerStop)

				// Notify all the observers by closing their channels
				for _, ch := range n.observers {
					close(ch)
				}

				// Bail out
				return
			}

		case ch := <-n.incomingReceivers:
			n.observers = append(n.observers, ch)
		}
	}
}

func NewPriceBelowNotifier(stop <-chan bool, market types.Market, price decimal.Decimal) *PriceBelow {
	n := &PriceBelow{
		observers:         []chan bool{},
		triggered:         false,
		incomingReceivers: make(chan chan bool),
	}
	n.runner(stop)
	return n
}

type PriceBelow struct {
	observers         []chan bool
	triggered         bool
	incomingReceivers chan chan bool
	market            types.Market
	price             decimal.Decimal
}

func (n *PriceBelow) Receive() <-chan bool {
	ch := make(chan bool)

	// Register the receiver
	n.incomingReceivers <- ch

	return ch
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
				// Stop the stream
				close(tickerStop)

				// Notify all the observers by closing their channels
				for _, ch := range n.observers {
					close(ch)
				}

				// Bail out
				return
			}

		case ch := <-n.incomingReceivers:
			n.observers = append(n.observers, ch)
		}
	}
}
