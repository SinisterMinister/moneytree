package followtheleader

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
	baseSize  decimal.Decimal
)

func (p *Processor) Process(db *sql.DB, trader types.Trader, market types.Market, stop <-chan bool) {
	// Store the local vars
	storeLocalVars(db, trader, market, stop)

	if viper.GetBool("followtheleader.refreshDatabasePairs") {
		// Refresh database pairs
		go refreshDatabasePairs()
	}

	// Set the initial direction
	direction = currentMarketDirection()

	for {
		// Restore the open orders
		go restoreDoneOpenOrders()

		// Make room for the next order if necessary
		err := makeRoom()
		if err != nil {
			log.WithError(err).Fatal("could not make room for new orders")
		}

		// Load the next pair to execute
		pair := nextPair()

		// Log the direction
		log.Infof("order direction is %s", direction)

		// Execute the order pair
		doneChan := pair.Execute(stop)

		// Start the pair bail out watchers
		go bailOnDirectionChange(pair)
		go bailOnMiss(pair)
		go bailOnPass(pair)

		// Wait for the order to be complete
		select {
		case <-stop: // Bail out if stopped
			return
		case <-doneChan:
		}

		// Nothing left to do but process again
		log.WithField("market", market.Name()).Info("market process cycle complete")
		<-time.NewTimer(viper.GetDuration("followtheleader.cycleDelay")).C
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

func restoreDoneOpenOrders() {
	// Load the open orders
	orders, err := pairSvc.LoadOpenPairs()
	if err != nil {
		log.WithError(err).Warn("could not restore open orders")
		return
	}

	// Execute all done open orders in the background
	for _, o := range orders {
		if o.IsDone() {
			o.Execute(stopChan)
		}
	}
}

func refreshDatabasePairs() {
	err := pairSvc.RefreshDatabasePairs()
	if err != nil {
		log.WithError(err).Error("error encountered while refreshing database pairs")
	}
}

func nextPair() *orderpair.OrderPair {
	// Try to recover the pair first
	pair, ok := recoverRunningPair()
	if ok { // Determine direction from pair
		if pair.FirstRequest().Side() == order.Buy {
			direction = Upward
		} else {
			direction = Downward
		}
		return pair
	}

	// Build pair based on the current direction
	var err error
	pair, err = buildPair(direction)
	if err != nil {
		log.WithError(err).Fatal("could not build the order pair")
	}

	// Use colliding open order if exists
	collidingPair, err := pairSvc.ResumeCollidingOpenPair(pair)
	if err != nil {
		log.WithError(err).Warn("could not load colliding open pair")
	}
	if collidingPair != nil {
		log.Infof("using open pair %s", pair.UUID().String())
		pair = collidingPair
	}
	return pair
}

func recoverRunningPair() (*orderpair.OrderPair, bool) {
	pair, err := pairSvc.LoadMostRecentRunningPair()
	if err != nil {
		log.WithError(err).Warn("could not load most recent running pair")
		return nil, false
	}

	return pair, true
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

	// Get spread
	spread, err := spread()
	if err != nil {
		return nil, err
	}

	// Set the sell price
	var sellPrice decimal.Decimal

	// Force taker orders
	if viper.GetBool("followtheleader.forceTakerOrders") {
		sellPrice = ticker.Bid()
	} else {
		sellPrice = ticker.Ask()
	}

	// Determine buy price
	buyPrice := sellPrice.Sub(sellPrice.Mul(spread)).Round(int32(quoteCurrency.Precision()))

	// Set the base size
	size, err := size(ticker)
	if err != nil {
		return nil, err
	}

	// Set buy size to base size
	buySize := size.Round(int32(baseCurrency.Precision()))

	// Determine sell size so that both currencies gain
	orderFee, err := getFees()
	if err != nil {
		return nil, fmt.Errorf("could not load fees: %w", err)
	}

	// Get the fees
	fee1 := orderFee.MakerRate()
	fee2 := orderFee.TakerRate()

	// Get the target return
	target := decimal.NewFromFloat(viper.GetFloat64("followtheleader.targetReturn"))

	// Setup the numbers we need
	two := decimal.NewFromFloat(2)

	// 2a - 2ab - abt - 2abg
	n := two.Mul(buySize).Sub(two.Mul(buySize).Mul(buyPrice)).Sub(buySize.Mul(buyPrice).Mul(target)).Sub(two.Mul(buySize).Mul(buyPrice).Mul(fee2))
	// t + 2f + 2 - 2d
	d := target.Add(two.Mul(fee1)).Add(two).Sub(two.Mul(sellPrice))

	// Set sell size
	sellSize := n.Div(d)

	// Round to correct precision
	sellSize = sellSize.Round(int32(baseCurrency.Precision()))

	// Build the order requests
	sellReq := order.NewRequest(market, order.Limit, order.Sell, sellSize, sellPrice)
	buyReq := order.NewRequest(market, order.Limit, order.Buy, buySize, buyPrice)
	log.WithFields(
		log.F("sellSize", sellSize.String()),
		log.F("sellPrice", sellPrice.String()),
		log.F("buySize", buySize.String()),
		log.F("buyPrice", buyPrice.String()),
	).Info("downward trending order data")

	// Create order pair
	op, err := pairSvc.New(sellReq, buyReq)
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

	// Set the prices
	var buyPrice decimal.Decimal

	// Force taker orders
	if viper.GetBool("followtheleader.forceTakerOrders") {
		buyPrice = ticker.Ask()
	} else {
		buyPrice = ticker.Bid()
	}

	// Set the ask price from the bid price
	sellPrice := buyPrice.Add(buyPrice.Mul(spread)).Round(int32(quoteCurrency.Precision()))

	// Set the sizes
	size, err := size(ticker)
	if err != nil {
		return nil, err
	}

	// Set buy size to base size
	buySize := size.Round(int32(baseCurrency.Precision()))

	// Determine sell size so that both currencies gain
	orderFee, err := getFees()
	if err != nil {
		return nil, fmt.Errorf("could not load fees: %w", err)
	}

	// Get the fees
	fee1 := orderFee.TakerRate()
	fee2 := orderFee.MakerRate()

	// Get the target return
	target := decimal.NewFromFloat(viper.GetFloat64("followtheleader.targetReturn"))

	// Setup the numbers we need
	two := decimal.NewFromFloat(2)

	// 2a - 2ab - abt - 2abg
	n := two.Mul(buySize).Sub(two.Mul(buySize).Mul(buyPrice)).Sub(buySize.Mul(buyPrice).Mul(target)).Sub(two.Mul(buySize).Mul(buyPrice).Mul(fee2))
	// t + 2f + 2 - 2d
	d := target.Add(two.Mul(fee1)).Add(two).Sub(two.Mul(sellPrice))

	// Set sell size
	sellSize := n.Div(d)

	// Round to correct precision
	sellSize = sellSize.Round(int32(baseCurrency.Precision()))

	// Build the order requests
	sellReq := order.NewRequest(market, order.Limit, order.Sell, sellSize, sellPrice)
	buyReq := order.NewRequest(market, order.Limit, order.Buy, buySize, buyPrice)
	log.WithFields(
		log.F("sellSize", sellSize.String()),
		log.F("sellPrice", sellPrice.String()),
		log.F("buySize", buySize.String()),
		log.F("buyPrice", buyPrice.String()),
	).Info("upward trending order data")

	// Create order pair
	op, err := pairSvc.New(buyReq, sellReq)
	if err != nil {
		return nil, fmt.Errorf("could not create order pair: %w", err)
	}
	return op, nil
}

func size(ticker types.Ticker) (decimal.Decimal, error) {
	if !baseSize.IsZero() {
		return baseSize, nil
	}

	// Get the max order size ration from max number of open orders plus 1 to add a buffer
	ratio := decimal.NewFromFloat(viper.GetFloat64("followtheleader.maxOpenOrders")).Add(decimal.NewFromFloat(1))

	// Determine order size from average volume
	size, err := market.AverageTradeVolume()
	if err != nil {
		return decimal.Zero, err
	}

	// Get wallets
	baseWallet := market.BaseCurrency().Wallet()
	quoteWallet := market.QuoteCurrency().Wallet()

	// Get the maximum trade size by the ratio
	baseMax := baseWallet.Available().Div(ratio)
	quoteMax := quoteWallet.Available().Div(ratio).Div(ticker.Bid())

	// Normalize the size to available funds
	if size.Equal(decimal.Zero) {
		size = decimal.Min(baseMax, quoteMax)
	}

	// Set the base size
	baseSize = decimal.Min(size, baseMax, quoteMax)
	return baseSize, nil
}

func getFees() (f types.Fees, err error) {
	// Allow disabling of fees to let the system work the raw algorithm
	if viper.GetBool("disableFees") == true {
		f = fees.ZeroFee()
	} else {
		// Get the fees
		var err error
		f, err = trader.AccountSvc().Fees()
		if err != nil {
			log.WithError(err).Error("failed to get fees")
			return fees.ZeroFee(), err
		}
	}
	return
}

func spread() (decimal.Decimal, error) {
	f, err := getFees()
	if err != nil {
		log.WithError(err).Error("failed to get fees")
		return decimal.Zero, err
	}

	// Set the profit target
	target := decimal.NewFromFloat(viper.GetFloat64("followtheleader.targetReturn"))

	// Add the taker and maker fees for the orders. We add both as while the first order can be
	// either a maker or a taker depending on configuration and/or timing, the second order is
	// always a maker order due to the nature of the application.
	rate := f.TakerRate().Add(f.MakerRate())

	// Calculate spread
	spread := target.Add(rate)

	return spread, nil
}

func bailOnDirectionChange(pair *orderpair.OrderPair) {
	var minPrice, maxPrice, bailPrice decimal.Decimal
	var cancel bool

	// Get the reversal spread from config
	reversalSpread := decimal.NewFromFloat(viper.GetFloat64("followtheleader.reversalSpread"))

	// Get the price ticker stream
	stop := make(chan bool)
	ticker := market.TickerStream(stop)

	for {
		// Watch for price changes and act accordingly
		select {
		// Bail out if we're exiting
		case <-stopChan:
			close(stop)
			return

		// Process new ticks
		case tick := <-ticker:
			switch direction {
			case Downward:
				if tick.Price().LessThan(minPrice) || minPrice.IsZero() {
					minPrice = tick.Price()

					// Set price based on reversal spread
					bailPrice = minPrice.Add(minPrice.Mul(reversalSpread))
					log.Infof("new price minimum. setting bail price to %s", bailPrice)
				}
				if tick.Price().GreaterThanOrEqual(bailPrice) {
					// Cancel the order
					cancel = true
					direction = Upward
				}
			case Upward:
				if tick.Price().GreaterThan(maxPrice) || maxPrice.IsZero() {
					maxPrice = tick.Price()

					// Set price based on reversal spread
					bailPrice = maxPrice.Sub(maxPrice.Mul(reversalSpread))
					log.Infof("new price maximum. setting bail price to %s", bailPrice)
				}
				if tick.Price().LessThanOrEqual(bailPrice) {
					// Cancel the order
					cancel = true
					direction = Downward
				}
			default:
				log.Error("invalid direction for bail price")
			}
		}
		// Bail out when time to cancel
		if cancel {
			// Cancel the order
			log.Infof("price direction changed. price passed %s. canceling order", bailPrice.StringFixed(2))
			err := pair.Cancel()
			if err != nil {
				log.WithError(err).Error("could not cancel order")
			}

			// Stop the ticker stream
			close(stop)
			return
		}
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
				log.Warn("first order passed")

				// Cancel the order
				err := pair.Cancel()
				if err != nil {
					log.WithError(err).Error("could not cancel order")
				}
				brk = true
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
				log.Warn("first order missed")

				// Cancel the order
				err := pair.Cancel()
				if err != nil {
					log.WithError(err).Error("could not cancel order")
				}
				brk = true
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

// makeRoom will cancel the oldest outstanding pair if there isn't enough room for a new order
func makeRoom() error {
	// Get the maximum number of open orders
	maxOpen := viper.GetInt("followtheleader.maxOpenOrders")

	// Get open orders
	pairs, err := pairSvc.LoadOpenPairs()
	if err != nil {
		return fmt.Errorf("could not load open pairs: %w", err)
	}

	// Cancel enough orders so that there's enough room for one more
	if len(pairs) >= maxOpen+1 {
		toCancel := len(pairs) - maxOpen
		oldPairs := pairs[:toCancel]
		for _, pair := range oldPairs {
			log.Info("canceling pair %s to make room for new pairs", pair.ToDAO().Uuid)
			err := pair.CancelAndTakeLosses()
			if err != nil {
				return fmt.Errorf("could not cancel pair: %w", err)
			}
		}
	}

	return nil
}
