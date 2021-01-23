module github.com/sinisterminister/moneytree

go 1.14

require (
	github.com/go-playground/ansi v2.1.0+incompatible // indirect
	github.com/go-playground/log v6.3.0+incompatible
	github.com/go-playground/log/v7 v7.0.2
	github.com/golang/mock v1.4.3
	github.com/golang/protobuf v1.4.3
	github.com/google/uuid v1.2.0 // indirect
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40
	github.com/lib/pq v1.9.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/satori/go.uuid v1.2.0
	github.com/shopspring/decimal v1.2.0
	github.com/sinisterminister/currencytrader v0.3.8
	github.com/sinisterminister/go-coinbasepro/v2 v2.1.0
	github.com/sinisterminister/miraclegrow v0.0.0-20210122224630-799d7deca2fc
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	golang.org/x/sys v0.0.0-20210122235752-a8b976e07c7b // indirect
	google.golang.org/genproto v0.0.0-20210122163508-8081c04a3579 // indirect
	google.golang.org/grpc v1.35.0
	google.golang.org/protobuf v1.25.0
)

// replace github.com/sinisterminister/currencytrader => ../currencytrader
