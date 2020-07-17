package orderpair

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/sinisterminister/currencytrader/types"
)

type OrderPairDAO struct {
	Uuid          string                `json:"uuid"`
	FirstRequest  types.OrderRequestDTO `json:"firstRequest"`
	SecondRequest types.OrderRequestDTO `json:"secondRequest"`
	FirstOrderID  string                `json:"firstOrderID"`
	SecondOrderID string                `json:"secondOrderID"`
	Done          bool                  `json:"done"`
	Failed        bool                  `json:"failed"`
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
