module github.com/sinisterminister/moneytree

go 1.14

require (
	github.com/go-playground/log v6.3.0+incompatible
	github.com/go-playground/log/v7 v7.0.2
	github.com/golang/mock v1.4.3
	github.com/preichenberger/go-coinbasepro/v2 v2.0.5
	github.com/shopspring/decimal v0.0.0-20190905144223-a36b5d85f337
	github.com/sinisterminister/currencytrader v0.0.0
	github.com/spf13/viper v1.6.2
)

replace github.com/sinisterminister/currencytrader => ../currencytrader
