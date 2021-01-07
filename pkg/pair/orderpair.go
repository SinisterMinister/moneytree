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

	mtx    sync.RWMutex
	runner sync.Once
	stop   chan bool

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
	}
}

func (o *OrderPair) Save() error {
	o.mtx.RLock()
	dao := o.ToDAO()
	o.mtx.RUnlock()

	return o.svc.Save(dao)
}

// Execute triggers the order pair execution
func (o *OrderPair) Execute() {
	// Only execute once
	o.runner.Do(func() {
		// Run execution in a separate goroutine
		log.Infof("%s: executing pair", o.UUID().String())
		go o.execute()
	})
}

func (o *OrderPair) Cancel() (err error) {
	// If first order exists and is still open, cancel it
	if o.FirstOrder() != nil && !o.FirstOrder().IsDone() {
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
		err = o.svc.trader.OrderSvc().CancelOrder(o.SecondOrder())
		if err != nil {
			return fmt.Errorf("%s: could not cancel second order - %w", o.UUID().String(), err)
		}
	}

	// Wait for both orders to be done
	if o.FirstOrder() != nil {
		<-o.FirstOrder().Done()
	}
	if o.SecondOrder() != nil {
		<-o.SecondOrder().Done()
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

		// Save the pair
		err = o.Save()
		if err != nil {
			log.WithError(err).Errorf("%s: could not save the pair", o.UUID().String())
		}
		return
	}

	// Save the pair
	o.setStatus(Open)
	err = o.Save()
	if err != nil {
		log.WithError(err).Errorf("%s: could not save the pair", o.UUID().String())
	}

	// Handle first order
	err = o.handleFirstOrder()
	if err != nil {
		log.WithError(err).Errorf("%s: error handling first order", o.UUID().String())
		o.setStatusDetails(err)

		// Save the pair
		err = o.Save()
		if err != nil {
			log.WithError(err).Errorf("%s: could not save the pair", o.UUID().String())
		}

		if o.FirstOrder().Filled().GreaterThan(decimal.Zero) && o.Status() == Canceled {
			log.Errorf("%s: reversing pair", o.UUID().String())
			err = o.buildReversalRequest()
			if err != nil {
				log.WithError(err).Errorf("%s: could not build reverse request", o.UUID().String())
				o.setStatusDetails(err)

				// Save the pair
				err = o.Save()
				if err != nil {
					log.WithError(err).Errorf("%s: could not save the pair", o.UUID().String())
				}
				return
			}
			o.Save()

			err = o.executeReversalRequest()
			if err != nil {
				log.WithError(err).Errorf("%s: could not reverse pair", o.UUID().String())
				o.setStatusDetails(err)

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

			// Save the pair
			err = o.Save()
			if err != nil {
				log.WithError(err).Errorf("%s: could not save the pair", o.UUID().String())
			}
		}
		return
	}

	// Execute second request
	err = o.executeSecondRequest()
	if err != nil {
		log.WithError(err).Errorf("%s: could not execute second request", o.UUID().String())
		o.setStatusDetails(err)

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
		log.WithError(err).Errorf("%s: error handling second order", o.UUID().String())
		o.setStatusDetails(err)

		// Save the pair
		err = o.Save()
		if err != nil {
			log.WithError(err).Errorf("%s: could not save the pair", o.UUID().String())
		}

		// Reverse the pair if the status has been set to reversed
		if o.Status() == Reversed {
			log.Errorf("%s: reversing pair", o.UUID().String())
			err = o.buildReversalRequest()
			if err != nil {
				log.WithError(err).Errorf("%s: could not build reverse request", o.UUID().String())
				o.setStatusDetails(err)

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
				o.setStatusDetails(err)

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

			// Save the pair
			err = o.Save()
			if err != nil {
				log.WithError(err).Errorf("%s: could not save the pair", o.UUID().String())
			}
		}
	}

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

	log.Infof("%s: placing reversal order - %s %s @ %s", o.uuid.String(), o.reversalRequest.Side(), o.reversalRequest.Quantity(), o.reversalRequest.Price())
	o.reversalOrder, err = o.svc.market.AttemptOrder(o.reversalRequest)
	return
}

func (o *OrderPair) handleFirstOrder() (err error) {
	// Wait for the first order to close
	<-o.FirstOrder().Done()
	log.Infof("%s: first order complete", o.UUID().String())

	// Handle first order outcome
	switch o.FirstOrder().Status() {
	case order.Canceled:
		// Check to see if it partially filled
		if o.FirstOrder().Filled().GreaterThan(decimal.Zero) && o.Status() != Canceled {
			log.Warnf("%s: recalculating second order size since first order was partially filled", o.UUID().String())
			o.recalculateSecondOrderSizeFromFilled()

			// Continue on
			break
		}

		// Mark pair as failed and bail
		err = fmt.Errorf("first order was canceled. stopping pair")
		o.mtx.Lock()
		o.status = Canceled
		o.endedAt = time.Now()
		o.statusDetails = err.Error()
		close(o.done)
		o.mtx.Unlock()

	case order.Filled:
		// Continue on
		break

	default:
		err = fmt.Errorf("first order returned unexpectedly with status %s", o.FirstOrder().Status())
		// Mark pair as failed and bail
		o.mtx.Lock()
		o.status = Broken
		o.endedAt = time.Now()
		o.statusDetails = err.Error()
		close(o.done)
		o.mtx.Unlock()
	}

	return
}

func (o *OrderPair) handleSecondOrder() (err error) {
	// Wait for the second order to close
	<-o.SecondOrder().Done()
	log.Infof("%s: second order complete", o.UUID().String())

	// Handle second order outcome
	o.mtx.Lock()
	switch o.secondOrder.Status() {
	case order.Canceled:
		err = fmt.Errorf("second order was canceled. setting status to %s and reversing", Reversed)
		// Mark pair as broken
		o.status = Reversed
		o.statusDetails = err.Error()
		close(o.done)

	case order.Filled:
		// Mark pair as success
		o.status = Success
		close(o.done)

	default:
		err = fmt.Errorf("second order returned unexpectedly with status %s", o.SecondOrder().Status())
		// Mark pair as broken
		o.status = Broken
		o.statusDetails = err.Error()
		close(o.done)
	}

	o.endedAt = time.Now()
	o.mtx.Unlock()

	return
}

func (o *OrderPair) recalculateSecondOrderSizeFromFilled() {
	// Determine the ratio from the first to the second
	ratio := o.SecondRequest().Quantity().Div(o.firstRequest.Quantity())

	// Calculate the new size
	size := o.FirstOrder().Filled().Mul(ratio).Round(int32(o.svc.market.BaseCurrency().Precision()))

	// Build updated DTO
	dto := o.SecondRequest().ToDTO()
	dto.Quantity = size

	// Set the new request
	o.mtx.Lock()
	o.secondRequest = order.NewRequestFromDTO(o.svc.market, dto)
	o.mtx.Unlock()
}

func (o *OrderPair) buildReversalRequest() error {
	// Get the current ticker
	ticker, err := o.svc.market.Ticker()
	if err != nil {
		return fmt.Errorf("could not get ticker: %w", err)
	}
	log.Infof("%s: using price %s for reversal order", o.UUID().String(), ticker.Ask().StringFixed(2))

	// Get the second order filled amount
	var filled decimal.Decimal
	if o.SecondOrder() != nil {
		filled = o.SecondOrder().Filled()
	}

	// Determine the size of the order
	size := o.FirstRequest().Quantity().Sub(filled)
	log.Infof("%s: use quantity %s for reversal order", o.UUID().String(), size.StringFixed(8))

	// Build reversal order
	req := order.NewRequest(o.svc.market, o.SecondRequest().Type(),
		o.SecondRequest().Side(), size, ticker.Ask(), false)

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
		if viper.GetBool("moneytree.forceMakerOrders") {
			baseFee = o.buyRequest().Quantity().Mul(rates.MakerRate())
		} else {
			baseFee = o.buyRequest().Quantity().Mul(rates.TakerRate())
		}
		quoteFee = o.sellRequest().Price().Mul(o.sellRequest().Quantity()).Mul(rates.MakerRate())
	} else {
		if viper.GetBool("moneytree.forceMakerOrders") {
			quoteFee = o.sellRequest().Price().Mul(o.sellRequest().Quantity()).Mul(rates.MakerRate())
		} else {
			quoteFee = o.sellRequest().Price().Mul(o.sellRequest().Quantity()).Mul(rates.TakerRate())
		}
		baseFee = o.buyRequest().Quantity().Mul(rates.MakerRate())
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
