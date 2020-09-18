package orderpair

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sinisterminister/currencytrader/types"
)

type OrderPairDAO struct {
	Uuid            string                `json:"uuid"`
	FirstRequest    types.OrderRequestDTO `json:"firstRequest"`
	SecondRequest   types.OrderRequestDTO `json:"secondRequest"`
	FirstOrderID    string                `json:"firstOrderID"`
	FirstOrder      types.OrderDTO        `json:"firstOrder"`
	SecondOrderID   string                `json:"secondOrderID"`
	SecondOrder     types.OrderDTO        `json:"secondOrder"`
	Done            bool                  `json:"done"`
	Failed          bool                  `json:"failed"`
	CreatedAt       time.Time             `json:"createdAt"`
	EndedAt         time.Time             `json:"endedAt"`
	Status          Status                `json:"status"`
	ReversalOrderId string                `json:"reversalOrderId"`
	ReversalOrder   types.OrderDTO        `json:"reversalOrder"`
}

func (o OrderPairDAO) Value() (driver.Value, error) {
	return json.Marshal(o)
}

func (o *OrderPairDAO) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &o)
}
