package pair

import (
	"time"

	"github.com/sinisterminister/currencytrader/types"
)

type OrderPairDAO struct {
	Uuid          string    `json:"uuid"`
	CreatedAt     time.Time `json:"createdAt"`
	EndedAt       time.Time `json:"endedAt"`
	Direction     Direction `json:"direction"`
	Done          bool      `json:"done"`
	Status        Status    `json:"status"`
	StatusDetails string    `json:"statusDetails"`

	FirstRequest types.OrderRequestDTO `json:"firstRequest"`
	FirstOrder   types.OrderDTO        `json:"firstOrder"`

	SecondRequest types.OrderRequestDTO `json:"secondRequest"`
	SecondOrder   types.OrderDTO        `json:"secondOrder"`

	ReversalRequest types.OrderRequestDTO `json:"reversalRequest"`
	ReversalOrder   types.OrderDTO        `json:"reversalOrder"`
}
