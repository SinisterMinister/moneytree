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
	Missed   Status = "MISSED"
	Broken   Status = "BROKEN"
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

	if o.firstOrder.IsDone() {
		// Close the done channel if necessary
		select {
		case <-o.done:
		default:
			close(o.done)
		}
		return o.svc.Save(o.ToDAO())
	}
	o.status = Canceled

	// Cancel the first order
	return o.svc.trader.OrderSvc().CancelOrder(o.firstOrder)
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
	var firstOrder, secondOrder types.OrderDTO
	if o.firstOrder != nil {
		firstOrder = o.firstOrder.ToDTO()
	}
	if o.secondOrder != nil {
		secondOrder = o.secondOrder.ToDTO()
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
	}
}

func (o *OrderPair) Save() error {
	o.mutex.RLock()
	dao := o.ToDAO()
	o.mutex.RUnlock()

	return o.svc.Save(dao)
}

func (o *OrderPair) executeWorkflow() {
	var err error

	// Bail out if already running
	o.mutex.Lock()
	if o.running {
		o.mutex.Unlock()
		return
	}
	// Mark workflow as running
	o.running = true
	o.mutex.Unlock()

	// Attempt to place first order
	err = o.placeFirstOrder()

	// Release start hold
	o.releaseStartHold()

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
	o.Save()

	// Place the second order. Retry if it fails
	for {
		err = o.placeSecondOrder()
		_, isSkipped := err.(*SkipSecondOrderError)
		if err != nil && !isSkipped {
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

	// End the workflow
	o.endWorkflow()
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

	// Save the order pair
	err := o.Save()
	if err != nil {
		log.WithError(err).Error("could not save order pair")
	}

	// If the orders are still open, launch routines to save when they close and update
	if !o.firstOrder.IsDone() {
		go func() {
			o.mutex.RLock()
			ord := o.firstOrder
			o.mutex.RUnlock()

			<-ord.Done()
			o.Save()
		}()
	}

	if o.secondOrder != nil && !o.secondOrder.IsDone() {
		go func() {
			o.mutex.RLock()
			ord := o.secondOrder
			o.mutex.RUnlock()

			<-ord.Done()
			o.Save()
		}()
	}
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
		log.Infof("first order done processing. status is %s", ord.Status())

		// Make sure the order completed successfully
		switch ord.Status() {
		case order.Filled:
			o.mutex.Lock()
			o.status = Open
			o.mutex.Unlock()
		case order.Canceled:
			o.mutex.Lock()
			o.status = Canceled
			o.mutex.Unlock()
		case order.Partial:
			log.Warn("first order status was partial when closed. attempting to cancel")
			o.mutex.Lock()
			o.status = Open
			err = o.svc.trader.OrderSvc().CancelOrder(ord)
			if err != nil {
				log.WithError(err).Warn("could not cancel order")
			}
			o.mutex.Unlock()
		case order.Unknown:
			log.Warn("first order status was UNKNOWN when closed. attempting to cancel")
			o.mutex.Lock()
			o.status = Canceled
			err = o.svc.trader.OrderSvc().CancelOrder(ord)
			if err != nil {
				log.WithError(err).Warn("could not cancel order")
				o.status = Failed
			}
			o.mutex.Unlock()
		case order.Pending:
			log.Warn("first order status was pending when closed. attempting to cancel")
			o.mutex.Lock()
			o.status = Canceled
			err = o.svc.trader.OrderSvc().CancelOrder(ord)
			if err != nil {
				log.WithError(err).Warn("could not cancel order")
				o.status = Failed
			}
			o.mutex.Unlock()
		case order.Rejected:
			log.Warn("first order status was rejected when closed. attempting to cancel")
			o.mutex.Lock()
			o.status = Failed
			o.mutex.Unlock()

		case order.Updated:
			log.Warn("first order status was updated when closed. attempting to cancel")
			o.mutex.Lock()
			o.status = Canceled
			err = o.svc.trader.OrderSvc().CancelOrder(ord)
			if err != nil {
				log.WithError(err).Warn("could not cancel order")
				o.status = Failed
			}
			o.mutex.Unlock()
		case order.Expired:
			log.Warn("first order status was expired when closed. attempting to cancel")
			o.mutex.Lock()
			o.status = Failed
			o.mutex.Unlock()
		}
	}
}

func (o *OrderPair) placeSecondOrder() (err error) {
	// Lock the pair while we work
	o.mutex.Lock()
	defer o.mutex.Unlock()

	// Don't execute the order if it already exists
	if o.secondOrder != nil {
		return
	}

	// Bail if fill amount is zero
	if o.firstOrder.Filled().Equal(decimal.Zero) {
		return &SkipSecondOrderError{}
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
	orderStop := o.stop
	ord := o.secondOrder
	o.mutex.RUnlock()

	if o.secondOrder == nil {
		log.Debug("no second order to wait for. bailing")
		return
	}

	select {
	case <-orderStop:
	case <-ord.Done():
		log.Info("second order done processing")

		// Make sure the order completed successfully
		switch ord.Status() {
		case order.Filled:
			o.mutex.Lock()
			o.status = Success
			o.mutex.Unlock()

		case order.Canceled:
			log.Warn("second order status was CANCELED when closed")
			o.mutex.Lock()
			o.status = Broken
			o.mutex.Unlock()

		case order.Partial:
			log.Warn("second order status was PARTIAL when closed")
			o.mutex.Lock()
			o.status = Broken
			o.mutex.Unlock()

		case order.Unknown:
			log.Warn("first order status was UNKNOWN when closed")
			o.mutex.Lock()
			o.status = Broken
			o.mutex.Unlock()

		case order.Pending:
			log.Warn("first order status was PENDING when closed")
			o.mutex.Lock()
			o.status = Broken
			o.mutex.Unlock()

		case order.Rejected:
			log.Warn("first order status was REJECTED when closed")
			o.mutex.Lock()
			o.status = Broken
			o.mutex.Unlock()

		case order.Updated:
			log.Warn("first order status was UPDATED when closed")
			o.mutex.Lock()
			o.status = Broken
			o.mutex.Unlock()

		case order.Expired:
			log.Warn("first order status was EXPIRED when closed")
			o.mutex.Lock()
			o.status = Broken
			o.mutex.Unlock()
		}
	}
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

func (o *OrderPair) recalculateSecondOrderSizeFromFilled() {
	// Determine the ratio from the first to the second
	ratio := o.secondRequest.Quantity().Div(o.firstRequest.Quantity())

	// Calculate the new size
	size := o.firstOrder.Filled().Mul(ratio).Round(int32(o.svc.market.QuoteCurrency().Precision()))

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
	baseFee := o.BuyRequest().Quantity().Mul(rates.TakerRate())
	quoteFee := o.SellRequest().Price().Mul(o.SellRequest().Quantity()).Mul(rates.TakerRate())

	// Make sure we're not losing currency
	if baseRes.LessThanOrEqual(baseFee) {
		return fmt.Errorf("not making more of base currency after fees, %w, %s, %s", &LosingPropositionError{o}, baseRes.String(), baseFee.String())
	}
	if quoteRes.LessThanOrEqual(quoteFee) {
		return fmt.Errorf("not making more of quote currency after fees, %w, %s, %s", &LosingPropositionError{o}, quoteRes.String(), quoteFee.String())
	}

	return nil
}
