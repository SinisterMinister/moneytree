package orderpair

import (
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/order"
)

type OrderPair struct {
	market types.Market

	firstRequest types.OrderRequestDTO
	firstOrder   types.Order

	secondRequest types.OrderRequestDTO
	secondOrder   types.Order

	done chan bool
}

func New(market types.Market, first types.OrderRequestDTO, second types.OrderRequestDTO) (orderPair *OrderPair, err error) {
	orderPair = &OrderPair{
		market:        market,
		done:          make(chan bool),
		firstRequest:  first,
		secondRequest: second,
	}

	// Validate DTOs
	err := orderPair.validate()
	if err != nil {
		return nil, err
	}

	return orderPair, nil
}

func (o *OrderPair) Execute() <-chan bool {
	go o.executeWorkflow()
	return o.done
}

func (o *OrderPair) Direction() Direction { return o.direction }

func (o *OrderPair) executeWorkflow() {
	// Place first order

	// Wait for order to complete, bailing if it misses

	// If order missed, send false over done channel before closing

	// Place second order

	// Wait for it to complete, timing out after a configured amount of time

	// If timed out, send false over done channel before closing

	// If successful, send true over done channel before closing
}

func (o *OrderPair) validate() error {
	// Make sure it's a BUY/SELL pair
	if o.firstRequest.Side == o.secondRequest.Side {
		return &SameSideError{o}
	}

	// Figure out the net result of the trades against our currency balance
	baseRes := decimal.Zero
	quoteRes := decimal.Zero
	if o.firstRequest.Side == order.Buy {
		baseRes = o.firstRequest.Quantity.Sub(o.secondRequest.Quantity)
		quoteRes = o.secondRequest.Price.Mul(o.secondRequest.Quantity).Sub(o.firstRequest.Price.Mul(o.firstRequest.Quantity)
	} else {
		baseRes = o.secondRequest.Quantity.Sub(o.firstRequest.Quantity)
		quoteRes = o.firstRequest.Price.Mul(o.firstRequest.Quantity).Sub(o.secondRequest.Price.Mul(o.secondRequest.Quantity)
	}

	// Make sure we're not losing currency
	if baseRes.LessThanOrEqual(decimal.Zero) && quoteRes.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("not making more of both currencies, %w", &LosingPropositionError{o})
	}

	// Get the fee rates
	rates, err := p.trader.AccountSvc().Fees()
	if err != nil {
		return err
	}

	// Determin the fees
	if o.firstRequest.Side == order.Buy {
		baseFee = o.firstRequest.Quantity.Mul(rates.TakerRate())
		quoteFee = o.secondRequest.Price.Mul(o.secondRequest.Quantity).Mul(rates.TakerRate())
	} else {
		baseFee = o.secondRequest.Quantity.Mul(rates.TakerRate())
		quoteFee = o.firstRequest.Price.Mul(o.firstRequest.Quantity).Mul(rates.TakerRate())
	}

	// Make sure we're making money
	if baseRes.LessThanOrEqual(baseFee) && quoteRes.LessThanOrEqual(quoteFee) {
		return fmt.Errorf("losing money after fees, %w", &LosingPropositionError{o})
	}

	return nil
}