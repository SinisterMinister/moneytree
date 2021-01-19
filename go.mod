module github.com/sinisterminister/moneytree

go 1.14

require (
	github.com/go-playground/log/v7 v7.0.2
	github.com/golang/mock v1.4.3
	github.com/golang/protobuf v1.4.3
	github.com/google/uuid v1.1.5 // indirect
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40
	github.com/lib/pq v1.9.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/prometheus/client_golang v1.9.0 // indirect
	github.com/prometheus/procfs v0.3.0 // indirect
	github.com/satori/go.uuid v1.2.0
	github.com/shopspring/decimal v1.2.0
	github.com/sinisterminister/currencytrader v0.3.3
	github.com/sinisterminister/go-coinbasepro/v2 v2.1.0
	github.com/spf13/cobra v1.1.1
	github.com/spf13/viper v1.7.1
	golang.org/x/net v0.0.0-20201224014010-6772e930b67b // indirect
	google.golang.org/genproto v0.0.0-20210114201628-6edceaf6022f // indirect
	google.golang.org/grpc v1.35.0
	google.golang.org/protobuf v1.25.0
)

// replace github.com/sinisterminister/currencytrader => ../currencytrader
