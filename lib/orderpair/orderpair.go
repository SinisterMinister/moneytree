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
	// Only launch routine if not running already
	if !o.running {
		go o.executeWorkflow()
	}
	o.running = true
	o.stop = stop
	o.mutex.Unlock()

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

	// Place the first order
	if o.firstOrder == nil {
		err = o.placeFirstOrder()
		if err != nil {
			log.WithError(err).Error("could not place first order")
			if o.status == Open {
				o.status = Failed
			}

			// End the workflow
			o.endWorkflow()
			close(o.startHold)
			return
		}

		// Save the order
		err = o.Save()
		if err != nil {
			log.WithError(err).Error("could not save order")
		}
	}

	// release start hold
	close(o.startHold)

	// Wait for order to complete. If it fails, keep going in case partial fill
	err = o.waitForOrder()
	if err != nil {
		log.WithError(err).Warn("first order failed")

		// Cancel the order if it's still open
		if !o.firstOrder.IsDone() {
			err = o.svc.trader.OrderSvc().CancelOrder(o.firstOrder)
			if err != nil {
				log.WithError(err).Infof("could not cancel first order: %w", err)
			}
		}
	}

	// Continue with second order in case first partially filled
	if o.secondOrder == nil {
		// Place second order
		err = o.placeSecondOrder()
		if err != nil {
			log.WithError(err).Warn("second order failed")
			o.status = Failed

			// End the workflow
			o.endWorkflow()
			return
		}
		log.Info("second order placed")

		// Save the order
		err = o.Save()
		if err != nil {
			log.WithError(err).Error("could not save order pair")
		}
	}

	// Wait for second order to complete
	log.Info("waiting on second order")
	<-o.secondOrder.Done()
	log.Info("second order done processing")

	if o.secondOrder.Status() == order.Filled {
		o.status = Success
	}

	// End the workflow
	o.endWorkflow()
}

func (o *OrderPair) endWorkflow() {
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
			<-o.firstOrder.Done()
			o.Save()
		}()
	}
	if !o.secondOrder.IsDone() {
		go func() {
			<-o.secondOrder.Done()
			o.Save()
		}()
	}
}

func (o *OrderPair) placeFirstOrder() (err error) {
	r0 := o.firstRequest.ToDTO()
	log.WithFields(log.F("side", r0.Side), log.F("price", r0.Price), log.F("quantity", r0.Quantity)).Info("placing first order")

	// Place first order
	o.mutex.Lock()
	o.firstOrder, err = o.svc.market.AttemptOrder(o.firstRequest)
	o.mutex.Unlock()
	return
}

func (o *OrderPair) placeSecondOrder() (err error) {
	// Bail if fill amount is zero
	if o.firstOrder.Filled().Equal(decimal.Zero) {
		return fmt.Errorf("first order was not filled, skipping second")
	}

	// Place second order
	o.mutex.Lock()
	o.recalculateSecondOrderSizeFromFilled()
	r1 := o.secondRequest.ToDTO()
	log.WithFields(log.F("side", r1.Side), log.F("price", r1.Price), log.F("quantity", r1.Quantity)).Info("placing second order")
	o.secondOrder, err = o.svc.market.AttemptOrder(o.secondRequest)
	o.mutex.Unlock()
	return
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

func (o *OrderPair) waitForOrder() (err error) {
	o.mutex.RLock()
	orderStop := o.stop
	o.mutex.RUnlock()

	select {
	case <-orderStop:
		return fmt.Errorf("stop channel closed")
	case <-o.firstOrder.Done():
		log.Info("first order done processing")

		// Make sure the order completed successfully
		if o.firstOrder.Status() != order.Filled && o.firstOrder.Filled().Equals(decimal.Zero) {
			err = fmt.Errorf("first order did not complete successfully")
		} else {
			// Make sure we're still marked as opened if we're still working
			o.status = Open
		}
	}
	return
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
