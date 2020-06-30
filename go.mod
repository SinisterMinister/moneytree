module github.com/sinisterminister/moneytree

go 1.14

require (
	github.com/go-playground/ansi v2.1.0+incompatible // indirect
	github.com/go-playground/log v6.3.0+incompatible
	github.com/go-playground/log/v7 v7.0.2
	github.com/preichenberger/go-coinbasepro/v2 v2.0.5
	github.com/sinisterminister/currencytrader v0.0.0
	github.com/spf13/viper v1.6.2
)

replace github.com/sinisterminister/currencytrader => ../currencytrader
