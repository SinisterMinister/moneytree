package orderpair

import (
	"errors"
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

	market.EXPECT().ToDTO().Return(types.MarketDTO{}).AnyTimes()
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
	first    types.OrderRequest
	second   types.OrderRequest
}

func TestValidate_HappyPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	var tests = []testValidateHarness{
		validate_HappyPath_Upward(ctrl),
		validate_HappyPath_Downward(ctrl),
	}

	for _, tt := range tests {
		t.Run(tt.scenario, func(t *testing.T) {
			op := &OrderPair{
				trader:        tt.trader,
				market:        tt.market,
				firstRequest:  tt.first,
				secondRequest: tt.second,
			}
			err := op.validate()
			if err != nil {
				t.Errorf("failed to create order pair: %w", err)
			}
		})
	}
}

func validate_HappyPath_Upward(ctrl *gomock.Controller) testValidateHarness {
	scenario := "upward trending happy path"
	trader, market := buildStubs(ctrl)

	first := order.NewRequest(market, order.Limit, order.Buy, decimal.NewFromFloat(100), decimal.NewFromFloat(100))
	second := order.NewRequest(market, order.Limit, order.Sell, decimal.NewFromFloat(99), decimal.NewFromFloat(200))

	return testValidateHarness{scenario, trader, market, first, second}
}

func validate_HappyPath_Downward(ctrl *gomock.Controller) testValidateHarness {
	scenario := "downward trending happy path"
	trader, market := buildStubs(ctrl)

	first := order.NewRequest(market, order.Limit, order.Sell, decimal.NewFromFloat(99), decimal.NewFromFloat(200))
	second := order.NewRequest(market, order.Limit, order.Buy, decimal.NewFromFloat(100), decimal.NewFromFloat(100))

	return testValidateHarness{scenario, trader, market, first, second}
}

func TestValidate_LosingProposition(t *testing.T) {
	ctrl := gomock.NewController(t)
	var tests = []testValidateHarness{
		validate_LosingProposition_LossOfBaseCurrency(ctrl),
		validate_LosingProposition_LossOfQuoteCurrency(ctrl),
		validate_LosingProposition_LossOfBaseCurrencyFromFees(ctrl),
	}

	for _, tt := range tests {
		t.Run(tt.scenario, func(t *testing.T) {
			op := &OrderPair{
				trader:        tt.trader,
				market:        tt.market,
				firstRequest:  tt.first,
				secondRequest: tt.second,
			}
			err := op.validate()
			var expected *LosingPropositionError
			if !errors.As(err, &expected) {
				t.Errorf("exected LosingPropositionError, got %T", err)
			}
		})
	}
}

func validate_LosingProposition_LossOfBaseCurrency(ctrl *gomock.Controller) testValidateHarness {
	scenario := "prevent losing base currency"
	trader, market := buildStubs(ctrl)

	first := order.NewRequest(market, order.Limit, order.Buy, decimal.NewFromFloat(100), decimal.NewFromFloat(100))
	second := order.NewRequest(market, order.Limit, order.Sell, decimal.NewFromFloat(101), decimal.NewFromFloat(200))

	return testValidateHarness{scenario, trader, market, first, second}
}

func validate_LosingProposition_LossOfQuoteCurrency(ctrl *gomock.Controller) testValidateHarness {
	scenario := "prevent losing quote currency"
	trader, market := buildStubs(ctrl)

	first := order.NewRequest(market, order.Limit, order.Buy, decimal.NewFromFloat(100), decimal.NewFromFloat(100))
	second := order.NewRequest(market, order.Limit, order.Sell, decimal.NewFromFloat(99), decimal.NewFromFloat(99))

	return testValidateHarness{scenario, trader, market, first, second}
}

func validate_LosingProposition_LossOfBaseCurrencyFromFees(ctrl *gomock.Controller) testValidateHarness {
	scenario := "prevent losing base currency to fees"
	trader, market := buildStubs(ctrl)

	first := order.NewRequest(market, order.Limit, order.Buy, decimal.NewFromFloat(100), decimal.NewFromFloat(100))
	second := order.NewRequest(market, order.Limit, order.Sell, decimal.NewFromFloat(100), decimal.NewFromFloat(200))

	return testValidateHarness{scenario, trader, market, first, second}
}

func validate_LosingProposition_LossOfQuoteCurrencyFromFees(ctrl *gomock.Controller) testValidateHarness {
	scenario := "prevent losing quote currency to fees"
	trader, market := buildStubs(ctrl)

	first := order.NewRequest(market, order.Limit, order.Buy, decimal.NewFromFloat(100), decimal.NewFromFloat(100))
	second := order.NewRequest(market, order.Limit, order.Sell, decimal.NewFromFloat(99), decimal.NewFromFloat(100))

	return testValidateHarness{scenario, trader, market, first, second}
}

func TestValidate_SameSide(t *testing.T) {
	ctrl := gomock.NewController(t)
	trader, market := buildStubs(ctrl)

	first := order.NewRequest(market, order.Limit, order.Buy, decimal.NewFromFloat(100), decimal.NewFromFloat(100))
	second := order.NewRequest(market, order.Limit, order.Buy, decimal.NewFromFloat(101), decimal.NewFromFloat(200))

	op := &OrderPair{
		trader:        trader,
		market:        market,
		firstRequest:  first,
		secondRequest: second,
	}
	err := op.validate()
	var expected *SameSideError
	if !errors.As(err, &expected) {
		t.Errorf("exected SameSideError, got %s", err)
	}
}
