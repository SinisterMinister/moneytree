package chaser

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/go-playground/log/v7"
	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/candle"
	"github.com/sinisterminister/currencytrader/types/fees"
	"github.com/sinisterminister/currencytrader/types/order"
	"github.com/sinisterminister/moneytree/lib/notifier"
	"github.com/sinisterminister/moneytree/lib/orderpair"
	"github.com/sinisterminister/moneytree/lib/trix"
	"github.com/spf13/viper"
)

type Direction string

var (
	Upward   Direction = "UPWARD"
	Downward Direction = "DOWNWARD"
)

type Processor struct{}

var (
	db        *sql.DB
	trader    types.Trader
	market    types.Market
	pairSvc   *orderpair.Service
	stopChan  <-chan bool
	direction Direction
)

func (p *Processor) Process(db *sql.DB, trader types.Trader, market types.Market, stop <-chan bool) {
	// Store the local vars
	storeLocalVars(db, trader, market, stop)

	for {

		// Load the next pair to execute
		pair := nextPair()

		// Execute the order pair
		doneChan := pair.Execute(stop)

		// Start the pair bail out watchers
		go bailOnDirectionChange(pair)

		// Wait for the order to be complete
		select {
		case <-stop: // Bail out if stopped
			return
		case <-doneChan:
		}

		// Nothing left to do but process again
		log.WithField("market", market.Name()).Info("market process cycle complete")
		<-time.NewTimer(viper.GetDuration("moneytree.cycleDelay")).C
	}
}

func storeLocalVars(d *sql.DB, t types.Trader, m types.Market, stop <-chan bool) {
	// Populate the local variables
	db = d
	trader = t
	market = m
	stopChan = stop

	// Start the order pair service
	svc, err := orderpair.NewService(db, trader, market)
	if err != nil {
		log.WithError(err).Fatal("could not start order pair service")
	}
	pairSvc = svc
}

func nextPair() *orderpair.OrderPair {
	// Try to recover the pair first
	pair, ok := recoverRunningPair()
	if !ok {
		// Get the direction the next order should go
		direction = nextPairDirection()

		// Build pair based on that direction
		var err error
		pair, err = buildPair(direction)
		if err != nil {
			log.WithError(err).Fatal("could not build the order pair")
		}
	}
	return pair
}

func recoverRunningPair() (*orderpair.OrderPair, bool) {
	pair, err := pairSvc.LoadMostRecentPair()
	if err != nil {
		log.WithError(err).Error("could not load most recent pair")
		return nil, false
	}

	return pair, true
}

func nextPairDirection() Direction {
	// Get the most recent open pair
	pair, err := pairSvc.LoadMostRecentOpenPair()
	if err != nil {
		log.WithError(err).Error("could not load most recent open pair")
		return currentMarketDirection()
	}
	if pair.FirstRequest().Side() == order.Buy {
		return Upward
	}
	return Downward
}

func currentMarketDirection() Direction {
	if isMarketUpwardTrending() {
		return Upward
	}
	return Downward
}

func isMarketUpwardTrending() bool {
	// Get trix values
	candles, err := market.Candles(candle.FiveMinutes, time.Now().Add(-4*time.Hour), time.Now())
	if err != nil {
		log.WithError(err).Error("could not fetch candles to calculate market direction")

		// Default to downward if we can't check
		return false
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
		log.F("market", market.Name()),
		log.F("trix", ma),
		log.F("osc", osc),
	).Info("trix value computed")

	return osc > 0
}

func buildPair(dir Direction) (pair *orderpair.OrderPair, err error) {
	switch dir {
	case Upward:
		pair, err = buildUpwardPair()
	case Downward:
		pair, err = buildDownwardPair()
	}
	return
}

func buildDownwardPair() (*orderpair.OrderPair, error) {
	// Get the currencies
	quoteCurrency := market.QuoteCurrency()
	baseCurrency := market.BaseCurrency()

	// Get the ticker for the current prices
	ticker, err := market.Ticker()
	if err != nil {
		return nil, err
	}

	// Determine prices using the spread
	spread, err := spread()
	if err != nil {
		return nil, err
	}

	// Prepare the spread to be applied
	spread = decimal.NewFromFloat(1).Add(spread)

	// Set the prices
	askPrice := ticker.Ask()
	bidPrice := askPrice.Sub(askPrice.Mul(spread).Sub(askPrice)).Round(int32(quoteCurrency.Precision()))

	// Set the sizes
	size, err := size(ticker)
	if err != nil {
		return nil, err
	}
	bidSize := size.Round(int32(baseCurrency.Precision()))
	askSize := size.Div(decimal.NewFromFloat(2)).Mul(bidPrice).Div(askPrice).Add(size.Div(decimal.NewFromFloat(2))).Round(int32(baseCurrency.Precision()))

	// Build the order requests
	askReq := order.NewRequest(market, order.Limit, order.Sell, askSize, askPrice)
	bidReq := order.NewRequest(market, order.Limit, order.Buy, bidSize, bidPrice)
	log.WithFields(
		log.F("askSize", askSize.String()),
		log.F("askPrice", askPrice.String()),
		log.F("bidSize", bidSize.String()),
		log.F("bidPrice", bidPrice.String()),
	).Info("downward trending order data")

	// Create order pair
	op, err := pairSvc.New(askReq, bidReq)
	if err != nil {
		return nil, fmt.Errorf("could not create order pair: %w", err)
	}
	return op, nil
}

func buildUpwardPair() (*orderpair.OrderPair, error) {
	// Get the currencies
	quoteCurrency := market.QuoteCurrency()
	baseCurrency := market.BaseCurrency()

	// Get the ticker for the current prices
	ticker, err := market.Ticker()
	if err != nil {
		return nil, err
	}

	// Determine prices using the spread
	spread, err := spread()
	if err != nil {
		return nil, err
	}

	// Prepare the spread to be applied
	spread = decimal.NewFromFloat(1).Add(spread)

	// Set the prices
	bidPrice := ticker.Bid()
	askPrice := bidPrice.Mul(spread).Round(int32(quoteCurrency.Precision()))

	// Set the sizes
	size, err := size(ticker)
	if err != nil {
		return nil, err
	}
	bidSize := size.Round(int32(baseCurrency.Precision()))
	askSize := size.Div(decimal.NewFromFloat(2)).Mul(bidPrice).Div(askPrice).Add(size.Div(decimal.NewFromFloat(2))).Round(int32(baseCurrency.Precision()))

	// Build the order requests
	askReq := order.NewRequest(market, order.Limit, order.Sell, askSize, askPrice)
	bidReq := order.NewRequest(market, order.Limit, order.Buy, bidSize, bidPrice)
	log.WithFields(
		log.F("askSize", askSize.String()),
		log.F("askPrice", askPrice.String()),
		log.F("bidSize", bidSize.String()),
		log.F("bidPrice", bidPrice.String()),
	).Info("upward trending order data")

	// Create order pair
	op, err := pairSvc.New(bidReq, askReq)
	if err != nil {
		return nil, fmt.Errorf("could not create order pair: %w", err)
	}
	return op, nil
}

func size(ticker types.Ticker) (decimal.Decimal, error) {
	// Determine order size from average volume
	size, err := market.AverageTradeVolume()
	if err != nil {
		return size, err
	}

	// Get wallets
	baseWallet := market.BaseCurrency().Wallet()
	quoteWallet := market.QuoteCurrency().Wallet()

	// Get the maximum trade size by wallet
	baseMax := baseWallet.Available().Div(decimal.NewFromFloat(viper.GetFloat64("chaser.maxTradesFundsRatio")))
	quoteMax := quoteWallet.Available().Div(decimal.NewFromFloat(viper.GetFloat64("chaser.maxTradesFundsRatio"))).Div(ticker.Bid())

	// Normalize the size to available funds
	if size == decimal.Zero {
		size = decimal.Min(baseMax, quoteMax)
	}
	return decimal.Min(size, baseMax, quoteMax), nil
}

func spread() (decimal.Decimal, error) {
	var f types.Fees
	if viper.GetBool("disableFees") == true {
		f = fees.ZeroFee()
	} else {
		// Get the fees
		var err error
		f, err = trader.AccountSvc().Fees()
		if err != nil {
			log.WithError(err).Error("failed to get fees")
			return decimal.Zero, err
		}
	}

	// Set the profit target
	target := decimal.NewFromFloat(viper.GetFloat64("chaser.targetReturn"))

	// Add the taker fees twice for the two orders
	rate := f.TakerRate().Add(f.TakerRate())

	// Calculate spread
	spread := target.Add(rate)

	return spread, nil
}

func bailOnDirectionChange(pair *orderpair.OrderPair) {
	var (
		notify <-chan bool
	)
	price := bailPrice(pair)
	switch direction {
	case Upward:
		// Start a price notifier for us to cancel if the rises below
		notify = notifier.NewPriceBelowNotifier(stopChan, market, price).Receive()
		log.Debugf("waiting for price to fall below %s", price.String())
	case Downward:
		notify = notifier.NewPriceAboveNotifier(stopChan, market, price).Receive()
		log.Debugf("waiting for price to rise above %s", price.String())
	}

	select {
	case <-notify: // Price went to low, time to bail and transition to opposite state
		// Cancel the order
		log.Info("price direction changed. canceling order")
		err := pair.Cancel()
		if err != nil {
			log.WithError(err).Error("could not cancel order")
		}

	case <-pair.Done(): // Order completed successfully, nothing to do here
	case <-stopChan: // Bail on stop
	}
}

func bailOnPass(pair *orderpair.OrderPair) {
	stop := make(chan bool)
	tickerStream := market.TickerStream(stop)
	for {
		brk := false
		select {
		case <-pair.Done():
			close(stop)
			return
		case tick := <-tickerStream:
			// Bail if the order passed
			if pair.IsPassedOrder(tick.Price()) {
				brk = true
				log.Warn("first order partially filled but price passed second order")
			}
		case <-pair.FirstOrder().Done():
			// Order is complete, time to move on
			brk = true
		}

		// I want to break free...
		if brk {
			break
		}
	}
	// Close ticker stream
	close(stop)
}

func bailOnMiss(pair *orderpair.OrderPair) {
	stop := make(chan bool)
	tickerStream := market.TickerStream(stop)
	for {
		brk := false
		select {
		case <-pair.Done():
			close(stop)
			return
		case tick := <-tickerStream:
			// Bail if the order missed
			if pair.IsMissedOrder(tick.Price()) && pair.FirstOrder().Filled().Equals(decimal.Zero) {
				brk = true
				log.Warn("first order missed")
			}
		case <-pair.FirstOrder().Done():
			// Order is complete, time to move on
			brk = true
		}

		// I want to break free...
		if brk {
			break
		}
	}
	// Close ticker stream
	close(stop)
	return
}

func bailPrice(pair *orderpair.OrderPair) (price decimal.Decimal) {
	var err error
	switch direction {
	case Upward:
		// Try to get bail price from pair service
		price, err = pairSvc.LowestOpenBuyFirstPrice()
		if err != nil || price == decimal.Zero {
			log.WithError(err).Warn("could not find bail price from open orders. bailing to reversal spread")
		}
		req := pair.FirstRequest()
		if price == decimal.Zero {
			log.Warn("bail price zero. falling back to reversal spread")
			failSpread := pair.Spread().Mul(decimal.NewFromFloat(viper.GetFloat64("chaser.reversalSpread")))
			price = req.Price().Sub(req.Price().Mul(failSpread))
		} else {
			// Build price from target spread
			price = req.Price().Sub(req.Price().Mul(pair.Spread()))
		}
		log.Debugf("order bail price is %s", price.String())
	case Downward:
		// Try to get bail price from pair service
		price, err = pairSvc.HighestOpenSellFirstPrice()
		// If price is zero, use reversal as base
		if err != nil {
			log.WithError(err).Warn("could not find bail price from open orders. falling back to reversal spread")
		}
		req := pair.FirstRequest()
		if price == decimal.Zero {
			log.Warn("bail price zero. falling back to reversal spread")
			failSpread := pair.Spread().Mul(decimal.NewFromFloat(viper.GetFloat64("chaser.reversalSpread")))
			price = req.Price().Add(req.Price().Mul(failSpread))
		} else {
			// Build price from target spread
			price = req.Price().Add(req.Price().Mul(pair.Spread()))
		}
		log.Debugf("order bail price is %s", price.String())
	}
	return
}
