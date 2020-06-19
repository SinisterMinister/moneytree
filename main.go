package main

import (
	"os"
	"os/signal"

	"github.com/sinisterminister/moneytree/lib/moneytree"

	"github.com/go-playground/log/v7"
	"github.com/go-playground/log/v7/handlers/console"
	"github.com/preichenberger/go-coinbasepro/v2"
	"github.com/sinisterminister/currencytrader"
	"github.com/sinisterminister/currencytrader/types/provider/coinbase"
	"github.com/spf13/viper"
)

func main() {
	// Setup the console logger
	log.AddHandler(console.New(true), log.InfoLevel, log.WarnLevel, log.ErrorLevel, log.NoticeLevel, log.FatalLevel, log.AlertLevel, log.PanicLevel)

	// Setup the kill switch
	killSwitch := make(chan bool)

	// Setup a coinbase client
	client := coinbasepro.NewClient()

	// Connect to sandbox
	client.UpdateConfig(&coinbasepro.ClientConfig{
		BaseURL:    "https://api-public.sandbox.pro.coinbase.com",
		Key:        "db983743c2fa020a17502a111657b551",
		Passphrase: "throwback",
		Secret:     "SrHvi/n9HAcEoe/JXsaZlfok4O/hXULiK4OhoANFN5GS0odp5ciho1w1jmMXlQ40Br8G8GU6WGRPClmbQnUyEQ==",
	})

	// Setup sandbox websocket url
	viper.Set("coinbase.websocket.url", "wss://ws-feed-public.sandbox.pro.coinbase.com")

	// Start up a coinbase provider
	provider := coinbase.New(killSwitch, client)

	// Get an instance of the trader
	trader := currencytrader.New(provider)
	trader.Start()

	// Get the currencies to use
	btc, err := trader.WalletSvc().Currency("BTC")
	if err != nil {
		log.WithError(err).Fatal("could not get BTC")
	}
	usdc, err := trader.WalletSvc().Currency("USD")
	if err != nil {
		log.WithError(err).Fatal("could not get USD")
	}

	// Start a new moneytree
	moneytree.New(killSwitch, trader, btc, usdc)

	// Intercept the interrupt signal and pass it along
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Wait for the interrupt
	<-interrupt

	// Let the user know what happened
	log.Warn("Received an interrupt signal! Shutting down!")

	// Shutdown the trader
	trader.Stop()

	// Kill the provider
	close(killSwitch)
}
