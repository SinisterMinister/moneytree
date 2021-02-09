package server

import (
	"github.com/sinisterminister/moneytree/pkg/pair"
	"github.com/sinisterminister/moneytree/pkg/proto"
)

func createProtoPair(op *pair.OrderPair) *proto.Pair {
	var buyOrder, sellOrder *proto.Order
	if op.BuyOrder() != nil {
		buyOrder = &proto.Order{
			Side:     "BUY",
			Price:    op.BuyRequest().Price().String(),
			Quantity: op.BuyRequest().Quantity().String(),
			Filled:   op.BuyOrder().Filled().String(),
			Status:   string(op.BuyOrder().Status()),
		}
	} else {
		buyOrder = &proto.Order{
			Side:     "BUY",
			Price:    op.BuyRequest().Price().String(),
			Quantity: op.BuyRequest().Quantity().String(),
		}
	}

	if op.SellOrder() != nil {
		sellOrder = &proto.Order{
			Side:     "SELL",
			Price:    op.SellRequest().Price().String(),
			Quantity: op.SellRequest().Quantity().String(),
			Filled:   op.SellOrder().Filled().String(),
			Status:   string(op.SellOrder().Status()),
		}

	} else {
		sellOrder = &proto.Order{
			Side:     "SELL",
			Price:    op.SellRequest().Price().String(),
			Quantity: op.SellRequest().Quantity().String(),
		}
	}

	return &proto.Pair{
		Uuid:          op.UUID().String(),
		Created:       op.CreatedAt().Unix(),
		Ended:         op.EndedAt().Unix(),
		Direction:     string(op.Direction()),
		Done:          op.IsDone(),
		Status:        string(op.Status()),
		StatusDetails: op.StatusDetails(),
		BuyOrder:      buyOrder,
		SellOrder:     sellOrder,
	}
}
