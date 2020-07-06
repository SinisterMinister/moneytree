package orderpair

import (
	"fmt"
	"sync"

	"github.com/go-playground/log/v7"
	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/order"
)

type OrderPair struct {
	trader types.Trader
	market types.Market

	mutex         sync.Mutex
	firstRequest  types.OrderRequestDTO
	firstOrder    types.Order
	secondRequest types.OrderRequestDTO
	secondOrder   types.Order
	running       bool
	done          chan bool
	stop          chan bool
}

func New(trader types.Trader, market types.Market, first types.OrderRequestDTO, second types.OrderRequestDTO) (orderPair *OrderPair, err error) {
	orderPair = &OrderPair{
		trader:        trader,
		market:        market,
		done:          make(chan bool),
		firstRequest:  first,
		secondRequest: second,
	}

	// Validate DTOs
	err = orderPair.validate()
	if err != nil {
		return nil, err
	}

	return orderPair, nil
}

func (o *OrderPair) Execute(stop <-chan bool) <-chan bool {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	// Only launch routine if not running already
	if !o.running {
		go o.executeWorkflow()
	}
	o.running = true
	o.stop = stop

	return o.done
}

func (o *OrderPair) FirstOrder() types.Order {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	return o.firstOrder
}

func (o *OrderPair) SecondOrder() types.Order {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	return o.SecondOrder
}

func (o *OrderPair) Cancel() error {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	return o.trader.OrderSvc().CancelOrder(o.firstOrder)
}

func (o *OrderPair) executeWorkflow() {
	// Place first order
	first, err := o.trader.OrderSvc().AttemptOrder(o.market, o.firstRequest.Type, o.firstRequest.Side, o.firstRequest.Price, o.firstRequest.Quantity)
	if err != nil {
		log.WithError(err).Error("could not place first order")
		close(o.done)
		return
	}

	// TODO: Wait for order to complete, bailing if it misses

	// Bail if fill amount is zero
	if o.firstOrder.Filled().Equal(decimal.Zero) {
		log.Warn("first order was not filled, skipping second")
		close(o.done)
		return
	}

	// Place second order
	second, err := o.trader.OrderSvc().AttemptOrder(o.market, o.secondRequest.Type, o.secondRequest.Side, o.secondRequest.Price, o.secondRequest.Quantity)
	if err != nil {
		log.WithError(err).Error("could not place second order")
		close(o.done)
		return
	}

	// TODO: Wait for it to complete

	// Signal completion
	close(o.done)
}

func (o *OrderPair) validate() error {
	// Make sure it's a BUY/SELL pair
	if o.firstRequest.Side == o.secondRequest.Side {
		return &SameSideError{o}
	}

	// Figure out the net result of the trades against our currency balance
	var baseRes decimal.Decimal
	var quoteRes decimal.Decimal
	if o.firstRequest.Side == order.Buy {
		baseRes = o.firstRequest.Quantity.Sub(o.secondRequest.Quantity)
		quoteRes = o.secondRequest.Price.Mul(o.secondRequest.Quantity).Sub(o.firstRequest.Price.Mul(o.firstRequest.Quantity))
	} else {
		baseRes = o.secondRequest.Quantity.Sub(o.firstRequest.Quantity)
		quoteRes = o.firstRequest.Price.Mul(o.firstRequest.Quantity).Sub(o.secondRequest.Price.Mul(o.secondRequest.Quantity))
	}

	// Make sure we're not losing currency
	if baseRes.LessThanOrEqual(decimal.Zero) || quoteRes.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("not making more of both currencies, %w", &LosingPropositionError{o})
	}

	// Get the fee rates
	rates, err := o.trader.AccountSvc().Fees()
	if err != nil {
		return err
	}

	// Determin the fees
	var baseFee decimal.Decimal
	var quoteFee decimal.Decimal
	if o.firstRequest.Side == order.Buy {
		baseFee = o.firstRequest.Quantity.Mul(rates.TakerRate())
		quoteFee = o.secondRequest.Price.Mul(o.secondRequest.Quantity).Mul(rates.TakerRate())
	} else {
		baseFee = o.secondRequest.Quantity.Mul(rates.TakerRate())
		quoteFee = o.firstRequest.Price.Mul(o.firstRequest.Quantity).Mul(rates.TakerRate())
	}

	// Make sure we're making money
	if baseRes.LessThanOrEqual(baseFee) || quoteRes.LessThanOrEqual(quoteFee) {
		return fmt.Errorf("not making money after fees, %w", &LosingPropositionError{o})
	}

	return nil
}
