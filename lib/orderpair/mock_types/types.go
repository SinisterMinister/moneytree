// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/sinisterminister/currencytrader/types (interfaces: Market,Trader,AccountSvc,Fees)

// Package mock_types is a generated GoMock package.
package mock_types

import (
	gomock "github.com/golang/mock/gomock"
	decimal "github.com/shopspring/decimal"
	types "github.com/sinisterminister/currencytrader/types"
	reflect "reflect"
	time "time"
)

// MockMarket is a mock of Market interface
type MockMarket struct {
	ctrl     *gomock.Controller
	recorder *MockMarketMockRecorder
}

// MockMarketMockRecorder is the mock recorder for MockMarket
type MockMarketMockRecorder struct {
	mock *MockMarket
}

// NewMockMarket creates a new mock instance
func NewMockMarket(ctrl *gomock.Controller) *MockMarket {
	mock := &MockMarket{ctrl: ctrl}
	mock.recorder = &MockMarketMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockMarket) EXPECT() *MockMarketMockRecorder {
	return m.recorder
}

// AttemptOrder mocks base method
func (m *MockMarket) AttemptOrder(arg0 types.OrderType, arg1 types.OrderSide, arg2, arg3 decimal.Decimal) (types.Order, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AttemptOrder", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(types.Order)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AttemptOrder indicates an expected call of AttemptOrder
func (mr *MockMarketMockRecorder) AttemptOrder(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AttemptOrder", reflect.TypeOf((*MockMarket)(nil).AttemptOrder), arg0, arg1, arg2, arg3)
}

// BaseCurrency mocks base method
func (m *MockMarket) BaseCurrency() types.Currency {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BaseCurrency")
	ret0, _ := ret[0].(types.Currency)
	return ret0
}

// BaseCurrency indicates an expected call of BaseCurrency
func (mr *MockMarketMockRecorder) BaseCurrency() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BaseCurrency", reflect.TypeOf((*MockMarket)(nil).BaseCurrency))
}

// Candles mocks base method
func (m *MockMarket) Candles(arg0 types.CandleInterval, arg1, arg2 time.Time) ([]types.Candle, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Candles", arg0, arg1, arg2)
	ret0, _ := ret[0].([]types.Candle)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Candles indicates an expected call of Candles
func (mr *MockMarketMockRecorder) Candles(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Candles", reflect.TypeOf((*MockMarket)(nil).Candles), arg0, arg1, arg2)
}

// MaxPrice mocks base method
func (m *MockMarket) MaxPrice() decimal.Decimal {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MaxPrice")
	ret0, _ := ret[0].(decimal.Decimal)
	return ret0
}

// MaxPrice indicates an expected call of MaxPrice
func (mr *MockMarketMockRecorder) MaxPrice() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MaxPrice", reflect.TypeOf((*MockMarket)(nil).MaxPrice))
}

// MaxQuantity mocks base method
func (m *MockMarket) MaxQuantity() decimal.Decimal {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MaxQuantity")
	ret0, _ := ret[0].(decimal.Decimal)
	return ret0
}

// MaxQuantity indicates an expected call of MaxQuantity
func (mr *MockMarketMockRecorder) MaxQuantity() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MaxQuantity", reflect.TypeOf((*MockMarket)(nil).MaxQuantity))
}

// MinPrice mocks base method
func (m *MockMarket) MinPrice() decimal.Decimal {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MinPrice")
	ret0, _ := ret[0].(decimal.Decimal)
	return ret0
}

// MinPrice indicates an expected call of MinPrice
func (mr *MockMarketMockRecorder) MinPrice() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MinPrice", reflect.TypeOf((*MockMarket)(nil).MinPrice))
}

// MinQuantity mocks base method
func (m *MockMarket) MinQuantity() decimal.Decimal {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MinQuantity")
	ret0, _ := ret[0].(decimal.Decimal)
	return ret0
}

// MinQuantity indicates an expected call of MinQuantity
func (mr *MockMarketMockRecorder) MinQuantity() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MinQuantity", reflect.TypeOf((*MockMarket)(nil).MinQuantity))
}

// Name mocks base method
func (m *MockMarket) Name() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Name")
	ret0, _ := ret[0].(string)
	return ret0
}

// Name indicates an expected call of Name
func (mr *MockMarketMockRecorder) Name() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Name", reflect.TypeOf((*MockMarket)(nil).Name))
}

// PriceIncrement mocks base method
func (m *MockMarket) PriceIncrement() decimal.Decimal {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PriceIncrement")
	ret0, _ := ret[0].(decimal.Decimal)
	return ret0
}

// PriceIncrement indicates an expected call of PriceIncrement
func (mr *MockMarketMockRecorder) PriceIncrement() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PriceIncrement", reflect.TypeOf((*MockMarket)(nil).PriceIncrement))
}

// QuantityStepSize mocks base method
func (m *MockMarket) QuantityStepSize() decimal.Decimal {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QuantityStepSize")
	ret0, _ := ret[0].(decimal.Decimal)
	return ret0
}

// QuantityStepSize indicates an expected call of QuantityStepSize
func (mr *MockMarketMockRecorder) QuantityStepSize() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QuantityStepSize", reflect.TypeOf((*MockMarket)(nil).QuantityStepSize))
}

// QuoteCurrency mocks base method
func (m *MockMarket) QuoteCurrency() types.Currency {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QuoteCurrency")
	ret0, _ := ret[0].(types.Currency)
	return ret0
}

// QuoteCurrency indicates an expected call of QuoteCurrency
func (mr *MockMarketMockRecorder) QuoteCurrency() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QuoteCurrency", reflect.TypeOf((*MockMarket)(nil).QuoteCurrency))
}

// Ticker mocks base method
func (m *MockMarket) Ticker() (types.Ticker, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Ticker")
	ret0, _ := ret[0].(types.Ticker)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Ticker indicates an expected call of Ticker
func (mr *MockMarketMockRecorder) Ticker() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ticker", reflect.TypeOf((*MockMarket)(nil).Ticker))
}

// TickerStream mocks base method
func (m *MockMarket) TickerStream(arg0 <-chan bool) <-chan types.Ticker {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TickerStream", arg0)
	ret0, _ := ret[0].(<-chan types.Ticker)
	return ret0
}

// TickerStream indicates an expected call of TickerStream
func (mr *MockMarketMockRecorder) TickerStream(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TickerStream", reflect.TypeOf((*MockMarket)(nil).TickerStream), arg0)
}

// ToDTO mocks base method
func (m *MockMarket) ToDTO() types.MarketDTO {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ToDTO")
	ret0, _ := ret[0].(types.MarketDTO)
	return ret0
}

// ToDTO indicates an expected call of ToDTO
func (mr *MockMarketMockRecorder) ToDTO() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ToDTO", reflect.TypeOf((*MockMarket)(nil).ToDTO))
}

// MockTrader is a mock of Trader interface
type MockTrader struct {
	ctrl     *gomock.Controller
	recorder *MockTraderMockRecorder
}

// MockTraderMockRecorder is the mock recorder for MockTrader
type MockTraderMockRecorder struct {
	mock *MockTrader
}

// NewMockTrader creates a new mock instance
func NewMockTrader(ctrl *gomock.Controller) *MockTrader {
	mock := &MockTrader{ctrl: ctrl}
	mock.recorder = &MockTraderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockTrader) EXPECT() *MockTraderMockRecorder {
	return m.recorder
}

// AccountSvc mocks base method
func (m *MockTrader) AccountSvc() types.AccountSvc {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AccountSvc")
	ret0, _ := ret[0].(types.AccountSvc)
	return ret0
}

// AccountSvc indicates an expected call of AccountSvc
func (mr *MockTraderMockRecorder) AccountSvc() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AccountSvc", reflect.TypeOf((*MockTrader)(nil).AccountSvc))
}

// MarketSvc mocks base method
func (m *MockTrader) MarketSvc() types.MarketSvc {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MarketSvc")
	ret0, _ := ret[0].(types.MarketSvc)
	return ret0
}

// MarketSvc indicates an expected call of MarketSvc
func (mr *MockTraderMockRecorder) MarketSvc() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MarketSvc", reflect.TypeOf((*MockTrader)(nil).MarketSvc))
}

// OrderSvc mocks base method
func (m *MockTrader) OrderSvc() types.OrderSvc {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "OrderSvc")
	ret0, _ := ret[0].(types.OrderSvc)
	return ret0
}

// OrderSvc indicates an expected call of OrderSvc
func (mr *MockTraderMockRecorder) OrderSvc() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OrderSvc", reflect.TypeOf((*MockTrader)(nil).OrderSvc))
}

// Start mocks base method
func (m *MockTrader) Start() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Start")
}

// Start indicates an expected call of Start
func (mr *MockTraderMockRecorder) Start() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockTrader)(nil).Start))
}

// Stop mocks base method
func (m *MockTrader) Stop() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Stop")
}

// Stop indicates an expected call of Stop
func (mr *MockTraderMockRecorder) Stop() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockTrader)(nil).Stop))
}

// TickerSvc mocks base method
func (m *MockTrader) TickerSvc() types.TickerSvc {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TickerSvc")
	ret0, _ := ret[0].(types.TickerSvc)
	return ret0
}

// TickerSvc indicates an expected call of TickerSvc
func (mr *MockTraderMockRecorder) TickerSvc() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TickerSvc", reflect.TypeOf((*MockTrader)(nil).TickerSvc))
}

// MockAccountSvc is a mock of AccountSvc interface
type MockAccountSvc struct {
	ctrl     *gomock.Controller
	recorder *MockAccountSvcMockRecorder
}

// MockAccountSvcMockRecorder is the mock recorder for MockAccountSvc
type MockAccountSvcMockRecorder struct {
	mock *MockAccountSvc
}

// NewMockAccountSvc creates a new mock instance
func NewMockAccountSvc(ctrl *gomock.Controller) *MockAccountSvc {
	mock := &MockAccountSvc{ctrl: ctrl}
	mock.recorder = &MockAccountSvcMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockAccountSvc) EXPECT() *MockAccountSvcMockRecorder {
	return m.recorder
}

// Currencies mocks base method
func (m *MockAccountSvc) Currencies() ([]types.Currency, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Currencies")
	ret0, _ := ret[0].([]types.Currency)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Currencies indicates an expected call of Currencies
func (mr *MockAccountSvcMockRecorder) Currencies() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Currencies", reflect.TypeOf((*MockAccountSvc)(nil).Currencies))
}

// Currency mocks base method
func (m *MockAccountSvc) Currency(arg0 string) (types.Currency, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Currency", arg0)
	ret0, _ := ret[0].(types.Currency)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Currency indicates an expected call of Currency
func (mr *MockAccountSvcMockRecorder) Currency(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Currency", reflect.TypeOf((*MockAccountSvc)(nil).Currency), arg0)
}

// Fees mocks base method
func (m *MockAccountSvc) Fees() (types.Fees, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Fees")
	ret0, _ := ret[0].(types.Fees)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Fees indicates an expected call of Fees
func (mr *MockAccountSvcMockRecorder) Fees() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Fees", reflect.TypeOf((*MockAccountSvc)(nil).Fees))
}

// Wallet mocks base method
func (m *MockAccountSvc) Wallet(arg0 types.Currency) (types.Wallet, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Wallet", arg0)
	ret0, _ := ret[0].(types.Wallet)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Wallet indicates an expected call of Wallet
func (mr *MockAccountSvcMockRecorder) Wallet(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Wallet", reflect.TypeOf((*MockAccountSvc)(nil).Wallet), arg0)
}

// Wallets mocks base method
func (m *MockAccountSvc) Wallets() ([]types.Wallet, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Wallets")
	ret0, _ := ret[0].([]types.Wallet)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Wallets indicates an expected call of Wallets
func (mr *MockAccountSvcMockRecorder) Wallets() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Wallets", reflect.TypeOf((*MockAccountSvc)(nil).Wallets))
}

// MockFees is a mock of Fees interface
type MockFees struct {
	ctrl     *gomock.Controller
	recorder *MockFeesMockRecorder
}

// MockFeesMockRecorder is the mock recorder for MockFees
type MockFeesMockRecorder struct {
	mock *MockFees
}

// NewMockFees creates a new mock instance
func NewMockFees(ctrl *gomock.Controller) *MockFees {
	mock := &MockFees{ctrl: ctrl}
	mock.recorder = &MockFeesMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockFees) EXPECT() *MockFeesMockRecorder {
	return m.recorder
}

// MakerRate mocks base method
func (m *MockFees) MakerRate() decimal.Decimal {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MakerRate")
	ret0, _ := ret[0].(decimal.Decimal)
	return ret0
}

// MakerRate indicates an expected call of MakerRate
func (mr *MockFeesMockRecorder) MakerRate() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MakerRate", reflect.TypeOf((*MockFees)(nil).MakerRate))
}

// TakerRate mocks base method
func (m *MockFees) TakerRate() decimal.Decimal {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TakerRate")
	ret0, _ := ret[0].(decimal.Decimal)
	return ret0
}

// TakerRate indicates an expected call of TakerRate
func (mr *MockFeesMockRecorder) TakerRate() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TakerRate", reflect.TypeOf((*MockFees)(nil).TakerRate))
}

// ToDTO mocks base method
func (m *MockFees) ToDTO() types.FeesDTO {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ToDTO")
	ret0, _ := ret[0].(types.FeesDTO)
	return ret0
}

// ToDTO indicates an expected call of ToDTO
func (mr *MockFeesMockRecorder) ToDTO() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ToDTO", reflect.TypeOf((*MockFees)(nil).ToDTO))
}

// Volume mocks base method
func (m *MockFees) Volume() decimal.Decimal {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Volume")
	ret0, _ := ret[0].(decimal.Decimal)
	return ret0
}

// Volume indicates an expected call of Volume
func (mr *MockFeesMockRecorder) Volume() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Volume", reflect.TypeOf((*MockFees)(nil).Volume))
}
