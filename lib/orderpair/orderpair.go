package orderpair

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-playground/log/v7"
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/fees"
	"github.com/sinisterminister/currencytrader/types/order"
	"github.com/spf13/viper"
)

type Status string

var (
	Open     Status = "OPEN"
	Success  Status = "SUCCESS"
	Failed   Status = "FAILED"
	Canceled Status = "CANCELED"
	Broken   Status = "BROKEN"
	Reversed Status = "REVERSED"
)

type OrderPair struct {
	svc           *Service
	mutex         sync.RWMutex
	uuid          uuid.UUID
	firstRequest  types.OrderRequest
	firstOrder    types.Order
	secondRequest types.OrderRequest
	secondOrder   types.Order
	running       bool
	failed        bool
	startHold     chan bool
	done          chan bool
	stop          <-chan bool
	createdAt     time.Time
	endedAt       time.Time
	status        Status
	reversalOrder types.Order
}

func (o *OrderPair) Execute(stop <-chan bool) <-chan bool {
	o.mutex.Lock()
	o.stop = stop
	o.mutex.Unlock()

	// Start workflow
	go o.executeWorkflow()

	// Wait for the orders to start
	<-o.startHold

	return o.done
}

func (o *OrderPair) IsDone() bool {
	select {
	case <-o.done:
		return true
	default:
		return false
	}
}

func (o *OrderPair) Done() <-chan bool {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	return o.done
}

func (o *OrderPair) UUID() uuid.UUID {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	return o.uuid
}

func (o *OrderPair) CreatedAt() time.Time {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	return o.createdAt
}

func (o *OrderPair) EndedAt() time.Time {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	return o.endedAt
}

func (o *OrderPair) FirstOrder() types.Order {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	return o.firstOrder
}

func (o *OrderPair) SecondOrder() types.Order {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	return o.secondOrder
}

func (o *OrderPair) FirstRequest() types.OrderRequest {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	return o.firstRequest
}

func (o *OrderPair) SecondRequest() types.OrderRequest {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	return o.secondRequest
}

func (o *OrderPair) BuyRequest() types.OrderRequest {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	if o.firstRequest.Side() == order.Buy {
		return o.firstRequest
	}
	return o.secondRequest
}

func (o *OrderPair) SellRequest() types.OrderRequest {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	if o.firstRequest.Side() != order.Buy {
		return o.firstRequest
	}
	return o.secondRequest
}

func (o *OrderPair) Spread() decimal.Decimal {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	return o.spread()
}

func (o *OrderPair) Cancel() error {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	// Don't cancel the pair if the first order is already complete
	if o.firstOrder.IsDone() {
		// Close the done channel if necessary to allow the next cycle to progress
		select {
		case <-o.done:
		default:
			close(o.done)
		}
		return o.svc.Save(o.ToDAO())
	}

	// Cancel the first order
	err := o.svc.trader.OrderSvc().CancelOrder(o.firstOrder)
	if err != nil {
		return err
	}

	// Mark order as canceled
	o.status = Canceled

	// Save and return
	return o.svc.Save(o.ToDAO())
}

func (o *OrderPair) ToDAO() OrderPairDAO {
	var fid, sid string
	if o.firstOrder != nil {
		fid = o.firstOrder.ID()
	}

	if o.secondOrder != nil {
		sid = o.secondOrder.ID()
	}

	var done bool
	select {
	case <-o.done:
		done = true
	default:
		done = false
	}
	var firstOrder, secondOrder, reversalOrder types.OrderDTO
	if o.firstOrder != nil {
		firstOrder = o.firstOrder.ToDTO()
	}
	if o.secondOrder != nil {
		secondOrder = o.secondOrder.ToDTO()
	}
	if o.reversalOrder != nil {
		reversalOrder = o.reversalOrder.ToDTO()
	}

	return OrderPairDAO{
		Uuid:          o.uuid.String(),
		FirstRequest:  o.firstRequest.ToDTO(),
		SecondRequest: o.secondRequest.ToDTO(),
		FirstOrderID:  fid,
		FirstOrder:    firstOrder,
		SecondOrderID: sid,
		SecondOrder:   secondOrder,
		Done:          done,
		Failed:        o.failed,
		Status:        o.status,
		CreatedAt:     o.createdAt,
		EndedAt:       o.endedAt,
		ReversalOrder: reversalOrder,
	}
}

func (o *OrderPair) Save() error {
	o.mutex.RLock()
	dao := o.ToDAO()
	o.mutex.RUnlock()

	return o.svc.Save(dao)
}

func (o *OrderPair) IsMissedOrder(price decimal.Decimal) bool {
	if o.firstRequest.Side() == order.Buy {
		if price.GreaterThan(o.missPrice()) {
			return true
		}
	} else {
		if price.LessThan(o.missPrice()) {
			return true
		}
	}
	return false
}

func (o *OrderPair) IsPassedOrder(price decimal.Decimal) bool {
	if o.firstRequest.Side() == order.Buy {
		if price.GreaterThan(o.secondRequest.Price()) {
			return true
		}
	} else {
		if price.LessThan(o.secondRequest.Price()) {
			return true
		}
	}
	return false
}

func (o *OrderPair) CancelAndTakeLosses() error {
	// Skip if the order is an incompatible state
	o.mutex.RLock()
	if o.status != Canceled && o.status != Broken {
		// Nothing to do here
		o.mutex.RUnlock()
		return nil
	}
	o.mutex.RUnlock()

	// First cancel the pair
	err := o.Cancel()
	if err != nil {
		return fmt.Errorf("could not cancel pair: %w", err)
	}

	// Lock the mutex before we begin work
	o.mutex.Lock()
	defer o.mutex.Unlock()

	// Wait for the pair to mark itself done
	log.Infof("canceled pair %s to end workflow. waiting for it to complete", o.uuid)
	<-o.done

	// Wait for the first order to be done
	if o.firstOrder != nil && !o.firstOrder.IsDone() {
		log.Infof("waiting for first order to complete")
		<-o.firstOrder.Done()
	}

	// Cancel second order if it's been placed and still running
	if o.secondOrder != nil && !o.secondOrder.IsDone() {

		log.Infof("canceling second order to reverse it")
		err := o.svc.trader.OrderSvc().CancelOrder(o.secondOrder)
		if err != nil {
			log.WithError(err).Error("error when canceling second order: %s", err)
		}
	}

	// Wait for the second order to be done
	if o.secondOrder != nil && !o.secondOrder.IsDone() {
		log.Infof("waiting for second order to complete")
		<-o.secondOrder.Done()
	}

	// If the first order was filled, we need to reverse the trade
	if !o.firstOrder.Filled().IsZero() {
		log.Infof("begin reversing order pair")
		// Get the current ticker
		ticker, err := o.svc.market.Ticker()
		if err != nil {
			return fmt.Errorf("could not get ticker: %w", err)
		}
		log.Infof("using price %s for reversal order", ticker.Ask().StringFixed(2))

		// Get the second order filled amount
		var filled decimal.Decimal
		if o.secondOrder != nil {
			filled = o.secondOrder.Filled()
		}

		// Determine the size of the order
		size := o.firstRequest.Quantity().Sub(filled)
		log.Infof("use quantity %s for reversal order", size.StringFixed(8))

		// Build reversal order
		req := order.NewRequest(o.svc.market, o.secondRequest.Type(),
			o.secondRequest.Side(), size, ticker.Ask(), false)

		// Place the reversal order
		log.Info("placing reversal order")
		order, err := o.svc.market.AttemptOrder(req)
		if err != nil {
			return fmt.Errorf("could not place reversal order: %w", err)
		}

		// Set the reversal order
		o.reversalOrder = order
		o.svc.Save(o.ToDAO())

		// Wait for the reversal order to be filled
		log.Info("waiting for reversal order to complete")
		<-order.Done()

		// Set the status to reversed
		log.Infof("setting status for pair %s to REVERSED", o.uuid)
		o.status = Reversed
	}

	// Set the end date to now
	o.endedAt = time.Now()

	// Save the pair
	err = o.svc.Save(o.ToDAO())
	if err != nil {
		return fmt.Errorf("could not save order pair: %w", err)
	}

	return nil
}

func (o *OrderPair) executeWorkflow() {
	var err error

	// Bail out if already running
	o.mutex.Lock()
	if o.running {
		o.mutex.Unlock()
		log.Debug("order already executing. bailing on new execution")
		return
	}
	// Mark workflow as running
	o.running = true
	o.mutex.Unlock()
	// Attempt to place first order
	err = o.placeFirstOrder()

	// Release start hold
	o.releaseStartHold()

	// Save the order to the database
	o.Save()

	// Handle any errors
	if err != nil {
		log.WithError(err).Errorf("could not place first order")

		// Mark pair as failed
		o.mutex.Lock()
		o.status = Failed
		o.mutex.Unlock()

		// End the workflow
		o.endWorkflow()
		return
	}

	// Wait for the first order to finish
	o.waitForFirstOrder()

	// Save the first order to the database
	o.Save()

	// Place the second order. Retry if it fails
	for {
		err = o.placeSecondOrder()

		// Save the order to the database
		o.Save()

		// Handle errors
		if err != nil {
			_, isSkipped := err.(*SkipSecondOrderError)
			if isSkipped {
				break
			}
			log.WithError(err).Errorf("could not place second order")

			// Don't spam the order
			<-time.NewTimer(5 * time.Second).C
			log.Info("retrying second order")
		} else {
			break
		}
	}

	// Wait for the second order to complete
	o.waitForSecondOrder()

	// Handle any recoverable failures
	o.recoverFromFailures()

	// End the workflow
	o.endWorkflow()

	// Save the order pair
	err = o.Save()
	if err != nil {
		log.WithError(err).Error("could not save order pair")
	}
}

func (o *OrderPair) releaseStartHold() {
	// Lock the pair while we work
	o.mutex.Lock()
	defer o.mutex.Unlock()

	// Close the channel
	select {
	case <-o.startHold:
	default:
		close(o.startHold)
	}
}

func (o *OrderPair) endWorkflow() {
	// Lock the pair while we work
	o.mutex.Lock()
	defer o.mutex.Unlock()

	// Close the done channel
	select {
	case <-o.done:
	default:
		close(o.done)
	}

	// Record the timestamp
	o.endedAt = time.Now()

	// launch routines to save when they close and update
	go func() {
		o.mutex.RLock()
		ord := o.firstOrder
		o.mutex.RUnlock()
		if ord == nil {
			return
		}

		<-ord.Done()

		o.mutex.Lock()
		freshOrd, err := o.svc.trader.OrderSvc().Order(o.svc.market, ord.ID())
		if err != nil {
			log.WithError(err).Warn("could not load fresh order data to save for first order")
		} else {
			o.firstOrder = freshOrd
		}
		o.mutex.Unlock()

		o.Save()
	}()
	go func() {
		o.mutex.RLock()
		ord := o.secondOrder
		o.mutex.RUnlock()
		if ord == nil {
			return
		}

		<-ord.Done()

		o.mutex.Lock()
		freshOrd, err := o.svc.trader.OrderSvc().Order(o.svc.market, ord.ID())
		if err != nil {
			log.WithError(err).Warn("could not load fresh order data to save for second order")
		} else {
			o.secondOrder = freshOrd
		}
		o.mutex.Unlock()

		o.Save()
	}()
}

func (o *OrderPair) placeFirstOrder() (err error) {
	// Lock the pair while we work
	o.mutex.Lock()
	defer o.mutex.Unlock()

	// Don't execute the order if it already exists
	if o.firstOrder != nil {
		return
	}

	r0 := o.firstRequest.ToDTO()
	log.WithFields(log.F("side", r0.Side), log.F("price", r0.Price), log.F("quantity", r0.Quantity)).Info("placing first order")

	// Place first order
	o.firstOrder, err = o.svc.market.AttemptOrder(o.firstRequest)
	return
}

func (o *OrderPair) waitForFirstOrder() {
	var err error

	o.mutex.RLock()
	orderStop := o.stop
	ord := o.firstOrder
	o.mutex.RUnlock()

	select {
	case <-orderStop:
	case <-ord.Done():
		// Load the order from the API to get the latest data in case the status is out of sync
		ord, err = o.svc.trader.OrderSvc().Order(o.svc.market, ord.ID())
		if err != nil {
			log.WithError(err).Error("could not load fresh order. falling back to pair order")
			ord = o.firstOrder
		} else {
			// Store the updated order to the pair
			o.mutex.Lock()
			o.firstOrder = ord
			o.mutex.Unlock()
		}
		log.Infof("first order done processing. status is %s", ord.Status())

		// Make sure the order is really done
		switch ord.Status() {
		case order.Partial:
			log.Warn("first order status was PARTIAL when closed. attempting to cancel")
			err = o.svc.trader.OrderSvc().CancelOrder(ord)

		case order.Unknown:
			log.Warn("first order status was UNKNOWN when closed. attempting to cancel")
			err = o.svc.trader.OrderSvc().CancelOrder(ord)

		case order.Pending:
			log.Warn("first order status was PENDING when closed. attempting to cancel")
			err = o.svc.trader.OrderSvc().CancelOrder(ord)

		case order.Updated:
			log.Warn("first order status was UPDATED when closed. attempting to cancel")
			err = o.svc.trader.OrderSvc().CancelOrder(ord)
		}
		if err != nil {
			log.WithError(err).Warn("could not cancel order")
		}

		// The order is only open if something was filled
		o.mutex.Lock()
		if !ord.Filled().IsZero() {
			o.status = Open
		} else {
			o.status = Broken
		}
		o.mutex.Unlock()
	}
}

func (o *OrderPair) placeSecondOrder() (err error) {
	// Lock the pair while we work
	o.mutex.Lock()
	defer o.mutex.Unlock()

	// Bail if pair is no longer open
	if o.status != Open {
		log.Warnf("pair status is %s. skipping second order", o.status)
		return &SkipSecondOrderError{}
	}

	// Don't execute the order if it already exists
	if o.secondOrder != nil {
		log.Warn("second order exists. skipping.")
		return
	}

	// Place second order
	o.recalculateSecondOrderSizeFromFilled()
	r1 := o.secondRequest.ToDTO()
	log.WithFields(log.F("side", r1.Side), log.F("price", r1.Price), log.F("quantity", r1.Quantity)).Info("placing second order")
	o.secondOrder, err = o.svc.market.AttemptOrder(o.secondRequest)
	return
}

func (o *OrderPair) waitForSecondOrder() {
	o.mutex.RLock()
	ord := o.secondOrder
	o.mutex.RUnlock()

	if ord == nil {
		log.Debug("no second order to wait for. bailing")
		return
	}

	// Wait for the order to complete
	<-ord.Done()
	log.Info("second order done processing")

	// Load the order from the API to get the latest data in case the status is out of sync
	ord, err := o.svc.trader.OrderSvc().Order(o.svc.market, ord.ID())
	if err != nil {
		log.WithError(err).Error("could not load fresh order; falling back to pair order")
		ord = o.secondOrder
	}

	// Make sure the order completed successfully
	o.mutex.Lock()
	if ord.Status() == order.Filled {
		o.status = Success
	} else {
		log.Warnf("second order status was %s when closed", ord.Status())
		o.status = Broken
	}
	o.mutex.Unlock()

	// Save the pair
	o.Save()
}

func (o *OrderPair) maxSpread() decimal.Decimal {
	return o.spread().Mul(decimal.NewFromFloat(viper.GetFloat64("orderpair.missPercentage")))
}

func (o *OrderPair) spread() decimal.Decimal {
	return o.firstRequest.Price().Sub(o.secondRequest.Price()).Div(o.firstRequest.Price()).Abs()
}

func (o *OrderPair) missPrice() decimal.Decimal {
	if o.firstRequest.Side() == order.Buy {
		return o.firstRequest.Price().Mul(o.maxSpread()).Add(o.firstRequest.Price())
	} else {
		return o.firstRequest.Price().Sub(o.firstRequest.Price().Mul(o.maxSpread()))
	}
}

func (o *OrderPair) recalculateSecondOrderSizeFromFilled() {
	// Determine the ratio from the first to the second
	ratio := o.secondRequest.Quantity().Div(o.firstRequest.Quantity())

	// Calculate the new size
	size := o.firstOrder.Filled().Mul(ratio).Round(int32(o.svc.market.BaseCurrency().Precision()))

	// Build updated DTO
	dto := o.secondRequest.ToDTO()
	dto.Quantity = size

	// Set the new request
	o.secondRequest = order.NewRequestFromDTO(o.svc.market, dto)
}

func (o *OrderPair) validate() error {
	// Make sure it's a BUY/SELL pair
	if o.firstRequest.Side() == o.secondRequest.Side() {
		return &SameSideError{o}
	}

	// Figure out the net result of the trades against our currency balance
	baseRes := o.BuyRequest().Quantity().Sub(o.SellRequest().Quantity())
	quoteRes := o.SellRequest().Price().Mul(o.SellRequest().Quantity()).Sub(o.BuyRequest().Price().Mul(o.BuyRequest().Quantity()))

	// Make sure we're not losing currency
	if baseRes.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("not making more of base currency, %w, %s", &LosingPropositionError{o}, baseRes.String())
	}
	if quoteRes.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("not making more of quote currency, %w, %s", &LosingPropositionError{o}, quoteRes.String())
	}

	// Get the fee rates
	rates, err := o.svc.trader.AccountSvc().Fees()
	if err != nil {
		return err
	}
	if viper.GetBool("disableFees") == true {
		rates = fees.ZeroFee()
	}

	// Determin the fees
	var baseFee, quoteFee decimal.Decimal
	if o.FirstRequest().Side() == order.Buy {
		baseFee = o.BuyRequest().Quantity().Mul(rates.TakerRate())
		quoteFee = o.SellRequest().Price().Mul(o.SellRequest().Quantity()).Mul(rates.MakerRate())
	} else {
		baseFee = o.BuyRequest().Quantity().Mul(rates.MakerRate())
		quoteFee = o.SellRequest().Price().Mul(o.SellRequest().Quantity()).Mul(rates.TakerRate())
	}

	// Make sure we're not losing currency
	if baseRes.LessThanOrEqual(baseFee) {
		return fmt.Errorf("not making more of base currency after fees, %w, %s, %s", &LosingPropositionError{o}, baseRes.String(), baseFee.String())
	}
	if quoteRes.LessThanOrEqual(quoteFee) {
		return fmt.Errorf("not making more of quote currency after fees, %w, %s, %s", &LosingPropositionError{o}, quoteRes.String(), quoteFee.String())
	}

	return nil
}

func (o *OrderPair) recoverFromFailures() {
	// Setup local vars to work with
	o.mutex.RLock()
	firstOrder := o.firstOrder
	secondOrder := o.secondOrder
	secondRequest := o.secondRequest
	status := o.status
	o.mutex.RUnlock()

	switch {
	// Check to see if there's anything to recover. Skip if first order wasn't filled or second order was filled correctly
	case status != Broken || firstOrder == nil || firstOrder.Filled().IsZero():
		// Nothing to recover
		return

	// Check if order was successful but just got marked as failed
	case secondOrder != nil && secondOrder.Filled().Equal(secondRequest.Quantity()):
		// This order is successful as the second order request is completely filled
		o.mutex.Lock()
		o.status = Success
		o.mutex.Unlock()
		return

	// Check to see if second order was placed
	case secondOrder == nil:
		// Second order wasn't placed. Reopen the pair
		o.mutex.Lock()
		o.status = Open
		o.mutex.Unlock()
		return

	case secondOrder != nil && secondOrder.Status() == order.Pending:
		// Reload the order to make sure it's still pending
		ord, err := o.svc.trader.OrderSvc().Order(secondOrder.Market(), secondOrder.ID())
		if err == nil {
			if ord.Status() == order.Pending {
				// Second order still pending. Reopen the pair
				o.mutex.Lock()
				o.status = Open
				o.mutex.Unlock()
				return
			}
		}
		log.WithError(err).Error("could not load pending second order for recovery")
		fallthrough

	// Second order wasn't fully filled. Reverse remaining order
	case secondOrder != nil && secondOrder.Filled().LessThan(secondRequest.Quantity()):
		err := o.CancelAndTakeLosses()
		if err != nil {
			log.WithError(err).Error("could not reverse order")
		}
	}
}
