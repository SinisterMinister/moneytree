package orderpair

import "fmt"

type LosingPropositionError struct {
	orderPair *OrderPair
}

func (err *LosingPropositionError) Error() string {
	return fmt.Sprintf("order pair would lose currency if executed")
}

type SameSideError struct {
	orderPair *OrderPair
}

func (err *SameSideError) Error() string {
	return fmt.Sprintf("orders are both of the same side (%s)", err.orderPair.firstRequest.Side())
}

type SkipSecondOrderError struct {
	orderPair *OrderPair
}

func (err *SkipSecondOrderError) Error() string {
	return fmt.Sprintf("first order was not filled, skipping second")
}
