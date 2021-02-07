package pair

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

type OrderPair struct {
	svc *Service

	mtx     sync.RWMutex
	runner  sync.Once
	ready   chan bool
	stop    chan bool
	execErr error

	uuid          uuid.UUID
	createdAt     time.Time
	endedAt       time.Time
	direction     Direction
	done          chan bool
	status        Status
	statusDetails string

	firstRequest    types.OrderRequest
	secondRequest   types.OrderRequest
	reversalRequest types.OrderRequest

	firstOrder    types.Order
	secondOrder   types.Order
	reversalOrder types.Order
}

func (o *OrderPair) IsDone() bool {
	o.mtx.RLock()
	defer o.mtx.RUnlock()

	return o.isDone()
}

func (o *OrderPair) Done() <-chan bool {
	o.mtx.RLock()
	defer o.mtx.RUnlock()

	return o.done
}

func (o *OrderPair) Status() Status {
	o.mtx.RLock()
	defer o.mtx.RUnlock()

	return o.status
}

func (o *OrderPair) StatusDetails() string {
	o.mtx.RLock()
	defer o.mtx.RUnlock()

	return o.statusDetails
}

func (o *OrderPair) UUID() uuid.UUID {
	o.mtx.RLock()
	defer o.mtx.RUnlock()

	return o.uuid
}

func (o *OrderPair) CreatedAt() time.Time {
	o.mtx.RLock()
	defer o.mtx.RUnlock()

	return o.createdAt
}

func (o *OrderPair) EndedAt() time.Time {
	o.mtx.RLock()
	defer o.mtx.RUnlock()

	return o.endedAt
}

func (o *OrderPair) Direction() Direction {
	o.mtx.RLock()
	defer o.mtx.RUnlock()

	return o.direction
}

func (o *OrderPair) FirstOrder() types.Order {
	o.mtx.RLock()
	defer o.mtx.RUnlock()

	return o.firstOrder
}

func (o *OrderPair) SecondOrder() types.Order {
	o.mtx.RLock()
	defer o.mtx.RUnlock()

	return o.secondOrder
}

func (o *OrderPair) ReversalOrder() types.Order {
	o.mtx.RLock()
	defer o.mtx.RUnlock()

	return o.reversalOrder
}

func (o *OrderPair) BuyOrder() types.Order {
	o.mtx.RLock()
	defer o.mtx.RUnlock()

	return o.buyOrder()
}

func (o *OrderPair) SellOrder() types.Order {
	o.mtx.RLock()
	defer o.mtx.RUnlock()

	return o.sellOrder()
}

func (o *OrderPair) FirstRequest() types.OrderRequest {
	o.mtx.RLock()
	defer o.mtx.RUnlock()

	return o.firstRequest
}

func (o *OrderPair) SecondRequest() types.OrderRequest {
	o.mtx.RLock()
	defer o.mtx.RUnlock()

	return o.secondRequest
}

func (o *OrderPair) ReversalRequest() types.OrderRequest {
	o.mtx.RLock()
	defer o.mtx.RUnlock()

	return o.reversalRequest
}

func (o *OrderPair) BuyRequest() types.OrderRequest {
	o.mtx.RLock()
	defer o.mtx.RUnlock()

	return o.buyRequest()
}

func (o *OrderPair) SellRequest() types.OrderRequest {
	o.mtx.RLock()
	defer o.mtx.RUnlock()

	return o.sellRequest()
}

func (o *OrderPair) ToDAO() OrderPairDAO {
	// Setup local vars using the order
	var (
		done                                         bool
		firstOrder, secondOrder, reversalOrder       types.OrderDTO
		firstRequest, secondRequest, reversalRequest types.OrderRequestDTO
	)

	// Set the value of done based on if the channel is open or not
	select {
	case <-o.done:
		done = true
	default:
		done = false
	}

	// Populate vars if they're set
	if o.firstOrder != nil {
		firstOrder = o.firstOrder.ToDTO()
	}
	if o.secondOrder != nil {
		secondOrder = o.secondOrder.ToDTO()
	}
	if o.reversalOrder != nil {
		reversalOrder = o.reversalOrder.ToDTO()
	}
	if o.firstRequest != nil {
		firstRequest = o.firstRequest.ToDTO()
	}
	if o.secondRequest != nil {
		secondRequest = o.secondRequest.ToDTO()
	}
	if o.reversalRequest != nil {
		reversalRequest = o.reversalRequest.ToDTO()
	}

	return OrderPairDAO{
		Uuid:            o.uuid.String(),
		FirstRequest:    firstRequest,
		SecondRequest:   secondRequest,
		ReversalRequest: reversalRequest,
		FirstOrder:      firstOrder,
		SecondOrder:     secondOrder,
		ReversalOrder:   reversalOrder,
		Done:            done,
		Direction:       o.direction,
		Status:          o.status,
		CreatedAt:       o.createdAt,
		EndedAt:         o.endedAt,
		StatusDetails:   o.statusDetails,
	}
}

func (o *OrderPair) Save() error {
	o.mtx.RLock()
	dao := o.ToDAO()
	o.mtx.RUnlock()

	return o.svc.Save(dao)
}

// Execute triggers the order pair execution
func (o *OrderPair) Execute() error {
	o.svc.mutex.RLock()
	ready := o.ready
	o.svc.mutex.RUnlock()

	// Only execute once
	o.runner.Do(func() {
		// Run execution in a separate goroutine
		log.Infof("%s: executing pair", o.UUID().String())
		go o.execute()
	})

	// Wait for the execution to finish starting
	<-ready

	// Return any errors
	o.svc.mutex.RLock()
	defer o.svc.mutex.RUnlock()
	return o.execErr
}

func (o *OrderPair) Cancel() (err error) {
	// If first order exists and is still open, cancel it
	if o.FirstOrder() != nil && !o.FirstOrder().IsDone() {
		log.Infof("%s: canceling first order", o.UUID().String())
		err = o.svc.trader.OrderSvc().CancelOrder(o.FirstOrder())
		if err != nil {
			if o.SecondOrder() == nil {
				return fmt.Errorf("%s: could not cancel first order - %w", o.UUID().String(), err)
			}
			log.WithError(err).Errorf("%s: could not cancel first order", o.UUID().String())
		}
	}

	// If second order exists and is still open, cancel it
	if o.SecondOrder() != nil && !o.SecondOrder().IsDone() {
		log.Infof("%s: canceling second order", o.UUID().String())
		err = o.svc.trader.OrderSvc().CancelOrder(o.SecondOrder())
		if err != nil {
			return fmt.Errorf("%s: could not cancel second order - %w", o.UUID().String(), err)
		}
	}

	// Wait for both orders to be done
	if o.FirstOrder() != nil && !o.FirstOrder().IsDone() {
		log.Infof("%s: waiting on first order to cancel", o.UUID().String())
		<-o.FirstOrder().Done()
	}
	if o.SecondOrder() != nil && !o.SecondOrder().IsDone() {
		log.Infof("%s: waiting on second order to cancel", o.UUID().String())
		<-o.SecondOrder().Done()
	}

	// Wait a sec just in case a reversal needs to happen
	<-time.Tick(time.Second)

	if o.ReversalOrder() != nil && !o.ReversalOrder().IsDone() {
		log.Infof("%s: waiting on reversal order to close", o.UUID().String())
		<-o.ReversalOrder().Done()
	}

	return
}

// ###########################
// ###   Private Methods   ###
// ###########################

func (o *OrderPair) execute() {
	var err error

	// Save the pair first
	err = o.Save()
	if err != nil {
		log.WithError(err).Errorf("%s: could not save the pair", o.UUID().String())
	}

	// Execute first request
	err = o.executeFirstRequest()
	if err != nil {
		log.WithError(err).Errorf("%s: could not execute first request", o.UUID().String())
		o.setStatus(Failed)
		o.setStatusDetails(err)
		o.setExecErr(err)
		close(o.done)
		close(o.ready)
		o.setEndedAt()

		// Save the pair
		err = o.Save()
		if err != nil {
			log.WithError(err).Errorf("%s: could not save the pair", o.UUID().String())
		}
		return
	}

	// Mark the pair as ready
	close(o.ready)
	o.setStatus(Open)

	// Save the pair
	err = o.Save()
	if err != nil {
		log.WithError(err).Errorf("%s: could not save the pair", o.UUID().String())
	}

	// Handle first order
	err = o.handleFirstOrder()
	if err != nil {
		log.WithError(err).Warnf("%s: error handling first order", o.UUID().String())

		// Save the pair
		err = o.Save()
		if err != nil {
			log.WithError(err).Errorf("%s: could not save the pair", o.UUID().String())
		}

		if o.Status() == Canceled && o.FirstOrder() != nil && o.FirstOrder().Filled().GreaterThan(decimal.Zero) {
			log.Errorf("%s: reversing pair", o.UUID().String())
			o.setStatus(Reversed)
			err = o.buildReversalRequest()
			if err != nil {
				log.WithError(err).Errorf("%s: could not build reverse request", o.UUID().String())
				o.setStatus(Broken)
				o.setStatusDetails(err)
				o.setEndedAt()
				close(o.done)

				// Save the pair
				err = o.Save()
				if err != nil {
					log.WithError(err).Errorf("%s: could not save the pair", o.UUID().String())
				}
				return
			}

			// Save the pair
			err = o.Save()
			if err != nil {
				log.WithError(err).Errorf("%s: could not save the pair", o.UUID().String())
			}

			err = o.executeReversalRequest()
			if err != nil {
				log.WithError(err).Errorf("%s: could not reverse pair", o.UUID().String())
				o.setStatus(Broken)
				o.setStatusDetails(err)
				o.setEndedAt()
				close(o.done)

				// Save the pair
				err = o.Save()
				if err != nil {
					log.WithError(err).Errorf("%s: could not save the pair", o.UUID().String())
				}
				return
			}

			// Save the pair
			err = o.Save()
			if err != nil {
				log.WithError(err).Errorf("%s: could not save the pair", o.UUID().String())
			}

			// Wait for the reversal order to complete
			<-o.ReversalOrder().Done()
			log.Infof("%s: reversal order complete", o.UUID().String())

			// Give the system a second to get consistent
			<-time.Tick(time.Second)

			// Load the reversal fees
			o.ReversalOrder().Refresh()

			// Save the pair
			err = o.Save()
			if err != nil {
				log.WithError(err).Errorf("%s: could not save the pair", o.UUID().String())
			}
		}

		o.setEndedAt()
		close(o.done)

		// Save the pair
		err = o.Save()
		if err != nil {
			log.WithError(err).Errorf("%s: could not save the pair", o.UUID().String())
		}
		return
	}

	// Recalculate the second order if necessary
	if !o.FirstOrder().Filled().Equal(o.FirstRequest().Quantity()) {
		o.recalculateSecondOrderSizeFromFilled()
	}

	// Execute second request
	err = o.executeSecondRequest()
	if err != nil {
		log.WithError(err).Errorf("%s: could not execute second request", o.UUID().String())
		o.setStatus(Broken)
		o.setStatusDetails(err)
		o.setEndedAt()
		close(o.done)

		// Save the pair
		err = o.Save()
		if err != nil {
			log.WithError(err).Errorf("%s: could not save the pair", o.UUID().String())
		}
		return
	}

	// Save the pair
	err = o.Save()
	if err != nil {
		log.WithError(err).Errorf("%s: could not save the pair", o.UUID().String())
	}

	// Handle second order
	err = o.handleSecondOrder()
	if err != nil {
		log.WithError(err).Warnf("%s: error handling second order", o.UUID().String())

		// Save the pair
		err = o.Save()
		if err != nil {
			log.WithError(err).Errorf("%s: could not save the pair", o.UUID().String())
		}

		// Reverse the pair if the status has been set to reversed
		if o.Status() == Reversed {
			log.Warnf("%s: reversing pair", o.UUID().String())
			err = o.buildReversalRequest()
			if err != nil {
				log.WithError(err).Errorf("%s: could not build reverse request", o.UUID().String())
				o.setStatus(Broken)
				o.setStatusDetails(err)
				o.setEndedAt()
				close(o.done)

				// Save the pair
				err = o.Save()
				if err != nil {
					log.WithError(err).Errorf("%s: could not save the pair", o.UUID().String())
				}
				return
			}

			// Save the pair
			err = o.Save()
			if err != nil {
				log.WithError(err).Errorf("%s: could not save the pair", o.UUID().String())
			}

			err = o.executeReversalRequest()
			if err != nil {
				log.WithError(err).Errorf("%s: could not reverse pair", o.UUID().String())
				o.setStatus(Broken)
				o.setStatusDetails(err)
				o.setEndedAt()
				close(o.done)

				// Save the pair
				err = o.Save()
				if err != nil {
					log.WithError(err).Errorf("%s: could not save the pair", o.UUID().String())
				}
				return
			}

			// Save the pair
			err = o.Save()
			if err != nil {
				log.WithError(err).Errorf("%s: could not save the pair", o.UUID().String())
			}

			// Wait for the reversal order to complete
			<-o.ReversalOrder().Done()
			log.Infof("%s: reversal order complete", o.UUID().String())

			// Give the system a second to get consistent
			<-time.Tick(time.Second)

			// Get the fees
			o.ReversalOrder().Refresh()

			// Save the pair
			err = o.Save()
			if err != nil {
				log.WithError(err).Errorf("%s: could not save the pair", o.UUID().String())
			}
		}
	}

	o.setEndedAt()
	close(o.done)

	// Save the pair
	err = o.Save()
	if err != nil {
		log.WithError(err).Errorf("%s: could not save the pair", o.UUID().String())
	}
}

func (o *OrderPair) executeFirstRequest() (err error) {
	o.mtx.Lock()
	defer o.mtx.Unlock()

	// Don't execute the order if it already exists
	if o.firstOrder != nil {
		log.Infof("%s: first order already placed", o.uuid.String())
		return
	}

	log.Infof("%s: placing first order - %s %s @ %s", o.uuid.String(), o.firstRequest.Side(), o.firstRequest.Quantity(), o.firstRequest.Price())
	o.firstOrder, err = o.svc.market.AttemptOrder(o.firstRequest)
	return
}

func (o *OrderPair) executeSecondRequest() (err error) {
	o.mtx.Lock()
	defer o.mtx.Unlock()

	// Don't execute the order if it already exists
	if o.secondOrder != nil {
		log.Infof("%s: second order already placed", o.uuid.String())
		return
	}

	log.Infof("%s: placing second order - %s %s @ %s", o.uuid.String(), o.secondRequest.Side(), o.secondRequest.Quantity(), o.secondRequest.Price())
	o.secondOrder, err = o.svc.market.AttemptOrder(o.secondRequest)
	return
}

func (o *OrderPair) executeReversalRequest() (err error) {
	o.mtx.Lock()
	defer o.mtx.Unlock()

	// Don't execute the order if it already exists
	if o.reversalOrder != nil {
		log.Infof("%s: reversal order already placed", o.uuid.String())
		return
	}

	log.Infof("%s: placing reversal order - %s %s", o.uuid.String(), o.reversalRequest.Side(), o.reversalRequest.Funds())
	o.reversalOrder, err = o.svc.market.AttemptOrder(o.reversalRequest)
	return
}

func (o *OrderPair) handleFirstOrder() (err error) {
	// Wait for the first order to close
	<-o.FirstOrder().Done()
	log.Infof("%s: first order complete", o.UUID().String())

	// Give the system a second to get consistent
	<-time.Tick(time.Second)

	// Refresh the order to make sure we have the fees
	err = o.FirstOrder().Refresh()
	if err != nil {
		// Something failed when trying to refresh. Trust nothing.
		return
	}

	// Handle first order outcome
	switch o.FirstOrder().Status() {
	case order.Canceled:
		// Mark pair as failed and bail
		err = fmt.Errorf("first order was canceled")
		o.setStatus(Canceled)
		o.setStatusDetails(err)

		return

	case order.Filled:
		// Continue on
		return

	case order.Pending:
		fallthrough
	case order.Partial:
		// Somehow the order was marked done when not fully updated or filled. We need to
		// poll the refresh method a few times to see if it finishes or not.
		var count time.Duration = 1

		// Retry refreshes
		for count < 10 {
			// Backoff on refreshes slowly
			<-time.Tick(time.Second * count)
			o.FirstOrder().Refresh()
			if o.FirstOrder().Status() == order.Filled {
				// We're good to move on
				return
			}
			if o.FirstOrder().Status() == order.Canceled {
				// Check to see if it partially filled
				if o.FirstOrder().Filled().GreaterThan(decimal.Zero) && o.Status() != Canceled {
					// Continue on
					return
				}

				// Mark pair as failed and bail
				err = fmt.Errorf("first order was canceled")
				o.setStatus(Canceled)
				o.setStatusDetails(err)
				return
			}
			count++
		}
		fallthrough

	default:
		err = fmt.Errorf("first order returned unexpectedly with status %s", o.FirstOrder().Status())
		// Mark pair as broken
		o.setStatus(Broken)
		o.setStatusDetails(err)
	}

	return
}

func (o *OrderPair) handleSecondOrder() (err error) {
	// Wait for the second order to close
	<-o.SecondOrder().Done()
	log.Infof("%s: second order complete", o.UUID().String())

	// Give the system a second to get consistent
	<-time.Tick(time.Second)

	// Refresh the order to get the fees
	err = o.SecondOrder().Refresh()
	if err != nil {
		// Something failed when trying to refresh. Trust nothing.
		return
	}

	// Handle second order outcome
	switch o.SecondOrder().Status() {
	case order.Canceled:
		err = fmt.Errorf("second order was canceled. setting status to %s and reversing", Reversed)
		// Mark pair as reversed
		o.setStatus(Reversed)
		o.setStatusDetails(err)

	case order.Filled:
		// Mark pair as success
		o.setStatus(Success)

	case order.Pending:
		fallthrough

	case order.Partial:
		// Somehow the order was marked done when not fully updated or filled. We need to
		// poll the refresh method a few times to see if it finishes or not.
		var count time.Duration = 1

		// Retry refreshes
		for count < 10 {
			// Backoff on refreshes slowly
			<-time.Tick(time.Second * count)
			o.SecondOrder().Refresh()
			if o.SecondOrder().Status() == order.Filled {
				// Mark pair as success
				o.setStatus(Success)
				return
			}
			if o.SecondOrder().Status() == order.Canceled {
				// Mark pair as reversed and bail
				err = fmt.Errorf("second order was canceled. setting status to %s and reversing", Reversed)
				o.setStatus(Reversed)
				o.setStatusDetails(err)
				return
			}
			count++
		}
		fallthrough

	default:
		err = fmt.Errorf("second order returned unexpectedly with status %s", o.SecondOrder().Status())
		// Mark pair as broken
		o.setStatus(Broken)
		o.setStatusDetails(err)
	}

	return
}

func (o *OrderPair) recalculateSecondOrderSizeFromFilled() {
	// Determine the ratio from the first to the second
	ratio := o.SecondRequest().Quantity().Div(o.firstRequest.Quantity())

	// Calculate the new size
	size := o.FirstOrder().Filled().Mul(ratio).RoundBank(int32(o.svc.market.BaseCurrency().Precision()))

	// Build updated DTO
	dto := o.SecondRequest().ToDTO()
	dto.Quantity = size

	// Set the new request
	o.mtx.Lock()
	o.secondRequest = order.NewRequestFromDTO(o.svc.market, dto)
	o.mtx.Unlock()
}

func (o *OrderPair) buildReversalRequest() error {
	var (
		req            types.OrderRequest
		remains, funds decimal.Decimal
	)

	// Get fee rates
	rates, err := getFees(o.svc.trader)
	if err != nil {
		log.WithError(err).Warnf("%s: could not get fee rates to predict loss", o.UUID().String())
	}

	// Get how much cash went out in the buy order
	if o.BuyOrder() != nil {
		remains = remains.Sub(o.BuyOrder().Filled().Mul(o.BuyOrder().Request().Price()))
		_, fee := o.BuyOrder().Fees()

		// Capture the fees
		remains = remains.Sub(fee)
	}

	// Get how much cash came back with the sell order
	if o.SellOrder() != nil {
		remains = remains.Add(o.SellOrder().Filled().Mul(o.SellOrder().Request().Price()))
		_, fee := o.SellOrder().Fees()

		// Capture the fees
		remains = remains.Sub(fee)
	}

	// Get how much cash remains to be filled
	one := decimal.NewFromInt(1)
	// Build the request
	if remains.IsPositive() {
		funds = remains.Div(one.Sub(rates.TakerRate())).RoundBank(int32(o.svc.market.QuoteCurrency().Precision()))
		req = order.NewRequest(o.svc.market, order.Market, order.Buy, decimal.Zero, decimal.Zero, funds, false)
	} else {
		funds = remains.Div(rates.TakerRate().Add(one)).RoundBank(int32(o.svc.market.QuoteCurrency().Precision()))
		req = order.NewRequest(o.svc.market, order.Market, order.Sell, decimal.Zero, decimal.Zero, funds.Abs(), false)
	}

	// Add to pair
	o.mtx.Lock()
	o.reversalRequest = req
	o.mtx.Unlock()

	return nil
}

func (o *OrderPair) isDone() bool {
	select {
	case <-o.done:
		return true
	default:
		return false
	}
}

func (o *OrderPair) buyRequest() types.OrderRequest {
	if o.firstRequest.Side() == order.Buy {
		return o.firstRequest
	}
	return o.secondRequest
}

func (o *OrderPair) sellRequest() types.OrderRequest {
	if o.firstRequest.Side() != order.Buy {
		return o.firstRequest
	}
	return o.secondRequest
}

func (o *OrderPair) buyOrder() types.Order {
	if o.firstRequest.Side() == order.Buy {
		return o.firstOrder
	}
	return o.secondOrder
}

func (o *OrderPair) sellOrder() types.Order {
	if o.firstRequest.Side() == order.Sell {
		return o.firstOrder
	}
	return o.secondOrder
}

func (o *OrderPair) setStatusDetails(err error) {
	o.mtx.Lock()
	defer o.mtx.Unlock()

	o.statusDetails = err.Error()
}

func (o *OrderPair) setStatus(status Status) {
	o.mtx.Lock()
	defer o.mtx.Unlock()

	o.status = status
}

func (o *OrderPair) setEndedAt() {
	o.mtx.Lock()
	defer o.mtx.Unlock()

	o.endedAt = time.Now()
}

func (o *OrderPair) setExecErr(err error) {
	o.mtx.Lock()
	defer o.mtx.Unlock()

	o.execErr = err
}

func (o *OrderPair) validate() error {
	o.mtx.RLock()
	defer o.mtx.RUnlock()

	// Make sure it's a BUY/SELL pair
	if o.firstRequest.Side() == o.secondRequest.Side() {
		return &SameSideError{o}
	}

	// Figure out the net result of the trades against our currency balance
	baseRes := o.buyRequest().Quantity().Sub(o.sellRequest().Quantity())
	quoteRes := o.sellRequest().Price().Mul(o.sellRequest().Quantity()).Sub(o.buyRequest().Price().Mul(o.buyRequest().Quantity()))

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
	if o.direction == Upward {
		if viper.GetBool("forceMakerOrders") {
			baseFee = o.buyRequest().Quantity().Mul(rates.MakerRate().Mul(o.buyRequest().Price()))
		} else {
			baseFee = o.buyRequest().Quantity().Mul(rates.TakerRate().Mul(o.buyRequest().Price()))
		}
		quoteFee = o.sellRequest().Price().Mul(o.sellRequest().Quantity().Mul(rates.MakerRate()))
	} else {
		if viper.GetBool("forceMakerOrders") {
			quoteFee = o.sellRequest().Price().Mul(o.sellRequest().Quantity()).Mul(rates.MakerRate())
		} else {
			quoteFee = o.sellRequest().Price().Mul(o.sellRequest().Quantity()).Mul(rates.TakerRate())
		}
		baseFee = o.buyRequest().Quantity().Mul(o.buyRequest().Price()).Mul(rates.MakerRate())
	}

	// Make sure we're not losing currency
	if quoteRes.LessThanOrEqual(quoteFee.Add(baseFee)) {
		return fmt.Errorf("not making more of quote currency after fees, %w, %s, %s", &LosingPropositionError{o}, quoteRes.String(), quoteFee.Add(baseFee))
	}

	return nil
}
