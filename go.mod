module github.com/sinisterminister/moneytree

go 1.15

require (
	github.com/go-playground/log/v7 v7.0.2
	github.com/golang/mock v1.4.3
	github.com/golang/protobuf v1.4.3
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40
	github.com/lib/pq v1.9.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/satori/go.uuid v1.2.0
	github.com/shopspring/decimal v1.2.0
	github.com/sinisterminister/currencytrader v0.7.3
	github.com/sinisterminister/go-coinbasepro/v2 v2.1.0
	github.com/spf13/cobra v1.1.1
	github.com/spf13/viper v1.7.1
	google.golang.org/grpc v1.35.0
	google.golang.org/protobuf v1.25.0
	gopkg.in/DATA-DOG/go-sqlmock.v1 v1.3.0 // indirect
)

// replace github.com/sinisterminister/currencytrader => ../currencytrader
