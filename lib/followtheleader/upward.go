package followtheleader

import (
	"fmt"
	"sync"

	"github.com/go-playground/log/v7"
	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types/order"
	"github.com/sinisterminister/moneytree/lib/notifier"
	"github.com/sinisterminister/moneytree/lib/orderpair"
	"github.com/sinisterminister/moneytree/lib/state"
	"github.com/spf13/viper"
)

type UpwardTrending struct {
	processor *Processor

	mutex     sync.Mutex
	doneChan  chan bool
	active    bool
	orderPair *orderpair.OrderPair
}

func (s *UpwardTrending) Activate(stop <-chan bool, manager *state.Manager) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.active {
		s.setupDoneChan()

		go s.run(stop, manager)
	}
	s.active = true
}

func (s *UpwardTrending) AllowedFrom() []state.State {
	return []state.State{&DownwardTrending{}}
}
func (s *UpwardTrending) Done() <-chan bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.setupDoneChan()
	return s.doneChan
}

func (s *UpwardTrending) setupDoneChan() {
	if s.doneChan == nil {
		// Build the done chan
		log.Info("setting up done chan")
		s.doneChan = make(chan bool)
	}
}

func (s *UpwardTrending) Resume(stop <-chan bool, manager *state.Manager) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Wait for the order to complete
	log.Info("resuming upward trending state processing")
	s.active = true
	go s.wait(stop, manager)
}

func (s *UpwardTrending) run(stop <-chan bool, manager *state.Manager) {
	// Build the order pair
	orderPair, err := s.buildPair()
	if err != nil {
		log.WithError(err).Error("could not build order pair")
		return
	}
	s.orderPair = orderPair

	// Execute the order
	orderPair.Execute(stop)
	log.Info("order pair execution started")

	// Wait for the order to complete
	s.wait(stop, manager)
}

func (s *UpwardTrending) wait(stop <-chan bool, manager *state.Manager) {
	// Start a price notifier for us to cancel if the rises below
	req := s.orderPair.FirstRequest()
	belowPrice := req.Price().Sub(req.Price().Mul(decimal.NewFromFloat(viper.GetFloat64("followtheleader.reversalSpread"))))
	belowNotifier := notifier.NewPriceBelowNotifier(stop, s.processor.market, belowPrice).Receive()
	log.Debugf("waiting for pair to complete or price to fall below %s", belowPrice.String())

	select {
	case <-belowNotifier: // Price went to low, time to bail and transition to opposite state
		// Cancel the order
		log.Info("price went too far below bid. canceling order")
		err := s.orderPair.Cancel()
		if err != nil {
			log.WithError(err).Error("could not cancel order")
		}

		// Transition to an upward trending state
		log.Info("transitioning to downward trending state")
		manager.TransitionTo(&DownwardTrending{processor: s.processor})
		return

	case <-s.orderPair.Done(): // Order completed successfully, nothing to do here
	case <-stop: // Bail on stop
	}

	// Close the done channel
	close(s.doneChan)
}

func (s *UpwardTrending) buildPair() (*orderpair.OrderPair, error) {
	// Get the currencies
	quoteCurrency := s.processor.market.QuoteCurrency()
	baseCurrency := s.processor.market.BaseCurrency()

	// Get the ticker for the current prices
	ticker, err := s.processor.market.Ticker()
	if err != nil {
		return nil, err
	}

	// Determine prices using the spread
	spread, err := s.processor.getSpread()
	if err != nil {
		return nil, err
	}

	// Prepare the spread to be applied
	spread = decimal.NewFromFloat(1).Add(spread)

	// Set the prices
	bidPrice := ticker.Bid()
	askPrice := bidPrice.Mul(spread).Round(int32(quoteCurrency.Precision()))

	// Set the sizes
	size, err := s.processor.getSize(ticker)
	if err != nil {
		return nil, err
	}
	bidSize := size.Round(int32(baseCurrency.Precision()))
	askSize := size.Div(decimal.NewFromFloat(2)).Mul(bidPrice).Div(askPrice).Add(size.Div(decimal.NewFromFloat(2))).Round(int32(baseCurrency.Precision()))

	// Build the order requests
	askReq := order.NewRequest(s.processor.market, order.Limit, order.Sell, askSize, askPrice)
	bidReq := order.NewRequest(s.processor.market, order.Limit, order.Buy, bidSize, bidPrice)
	log.WithFields(
		log.F("askSize", askSize.String()),
		log.F("askPrice", askPrice.String()),
		log.F("bidSize", bidSize.String()),
		log.F("bidPrice", bidPrice.String()),
	).Info("upward trending order data")

	// Create order pair
	op, err := orderpair.New(s.processor.db, s.processor.trader, s.processor.market, bidReq, askReq)
	if err != nil {
		return nil, fmt.Errorf("could not create order pair: %w", err)
	}
	return op, nil
}
