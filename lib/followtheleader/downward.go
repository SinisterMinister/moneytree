package followtheleader

import (
	"fmt"

	"github.com/go-playground/log/v7"
	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types/order"
	"github.com/sinisterminister/moneytree/lib/notifier"
	"github.com/sinisterminister/moneytree/lib/orderpair"
	"github.com/sinisterminister/moneytree/lib/state"
	"github.com/spf13/viper"
)

type DownwardTrending struct {
	processor *Processor
}

func (s *DownwardTrending) Activate(stop <-chan bool, manager *state.Manager) {
	for {
		// Build the order pair
		orderPair, err := s.buildPair()
		if err != nil {
			log.WithError(err).Error("could not build order pair")
		}

		// Execute the order
		orderDone := orderPair.Execute(stop)
		log.Info("order pair execution started")

		// Start a price notifier for us to cancel if the rises above
		req := orderPair.FirstRequest()
		abovePrice := req.Price().Add(req.Price().Mul(decimal.NewFromFloat(viper.GetFloat64("followtheleader.reversalSpread"))))
		aboveNotifier := notifier.NewPriceAboveNotifier(stop, s.processor.market, abovePrice).Receive()

		select {
		case <-stop: // Bail on stop
			return

		case <-aboveNotifier: // Price went to high, time to bail and transition to opposite state
			// Cancel the order
			log.Info("price went too far above ask. canceling order")
			err := orderPair.Cancel()
			if err != nil {
				log.WithError(err).Error("could not cancel order")
			}

			// Transition to an upward trending state
			log.Info("transitioning to upward trending state")
			manager.TransitionTo(&UpwardTrending{})
			return

		case <-orderDone: // Order completed successfully, nothing to do here
		}
	}
}

func (s *DownwardTrending) AllowedFrom() []state.State {
	return []state.State{&UpwardTrending{}}
}

func (s *DownwardTrending) buildPair() (*orderpair.OrderPair, error) {
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
	askPrice := ticker.Ask()
	bidPrice := askPrice.Sub(askPrice.Mul(spread).Sub(askPrice)).Round(int32(quoteCurrency.Precision()))

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
	).Info("downward trending order sizes")

	// Create order pair
	op, err := orderpair.New(s.processor.db, s.processor.trader, s.processor.market, askReq, bidReq)
	if err != nil {
		return nil, fmt.Errorf("could not create order pair: %w", err)
	}
	return op, nil
}