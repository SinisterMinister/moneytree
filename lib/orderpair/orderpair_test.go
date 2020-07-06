package orderpair

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/order"
	"github.com/sinisterminister/moneytree/lib/orderpair/mock_types"
)

func buildStubs(ctrl *gomock.Controller) (types.Trader, types.Market) {
	trader := mock_types.NewMockTrader(ctrl)
	market := mock_types.NewMockMarket(ctrl)
	accountSvc := mock_types.NewMockAccountSvc(ctrl)
	fees := mock_types.NewMockFees(ctrl)

	trader.EXPECT().AccountSvc().Return(accountSvc).AnyTimes()
	accountSvc.EXPECT().Fees().Return(fees, nil).AnyTimes()
	fees.EXPECT().MakerRate().Return(decimal.NewFromFloat(0.005)).AnyTimes()
	fees.EXPECT().TakerRate().Return(decimal.NewFromFloat(0.005)).AnyTimes()

	return trader, market
}

type testValidateHarness struct {
	scenario string
	trader   types.Trader
	market   types.Market
	first    types.OrderRequestDTO
	second   types.OrderRequestDTO
}

func TestValidate_HappyPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	var tests = []testValidateHarness{
		validate_HappyPath_Upward(ctrl),
	}

	for _, tt := range tests {
		t.Run(tt.scenario, func(t *testing.T) {
			_, err := New(tt.trader, tt.market, tt.first, tt.second)
			if err != nil {
				t.Errorf("failed to create order pair: %w", err)
			}
		})
	}
}

func validate_HappyPath_Upward(ctrl *gomock.Controller) testValidateHarness {
	trader, market := buildStubs(ctrl)

	first := types.OrderRequestDTO{
		Price:    decimal.NewFromFloat(100),
		Quantity: decimal.NewFromFloat(100),
		Side:     order.Buy,
		Type:     order.Limit,
		Market:   types.MarketDTO{},
	}

	second := types.OrderRequestDTO{
		Price:    decimal.NewFromFloat(200),
		Quantity: decimal.NewFromFloat(99),
		Side:     order.Sell,
		Type:     order.Limit,
		Market:   types.MarketDTO{},
	}

	return testValidateHarness{"upward trending happy path", trader, market, first, second}
}
