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

	// Connect to live
	client.UpdateConfig(&coinbasepro.ClientConfig{
		BaseURL:    "https://api.pro.coinbase.com",
		Key:        "7705b8bf56cb95c6e8957049e09c4b6d",
		Passphrase: "throwback",
		Secret:     "RQ8ml29SdZtjPBBwCbehRRQqZZaK8MSDlWeYJe5L2EJ3SaFU1U+Ter8agEeB7wsDC3oofcjgSWaEZ0dj0pweHw==",
	})

	// Setup sandbox websocket url
	viper.Set("coinbase.websocket.url", "wss://ws-feed.pro.coinbase.com")

	// Start up a coinbase provider
	provider := coinbase.New(killSwitch, client)

	// Get an instance of the trader
	trader := currencytrader.New(provider)
	trader.Start()

	Get the currencies to use
	btc, err := trader.WalletSvc().Currency("BTC")
	if err != nil {
		log.WithError(err).Fatal("could not get BTC")
	}
	usd, err := trader.WalletSvc().Currency("USD")
	if err != nil {
		log.WithError(err).Fatal("could not get USD")
	}
	usdc, err := trader.WalletSvc().Currency("USDC")
	if err != nil {
		log.WithError(err).Fatal("could not get USDC")
	}
	eth, err := trader.WalletSvc().Currency("ETH")
	if err != nil {
		log.WithError(err).Fatal("could not get ETH")
	}
	ltc, err := trader.WalletSvc().Currency("LTC")
	if err != nil {
		log.WithError(err).Fatal("could not get LTC")
	}
	xrp, err := trader.WalletSvc().Currency("XRP")
	if err != nil {
		log.WithError(err).Fatal("could not get XRP")
	}

	// Start a new moneytree
	moneytree.New(killSwitch, trader, btc, usd, usdc, eth, ltc, xrp)

	// Watch all currencies
	// currencies, _ := trader.WalletSvc().Currencies()
	// moneytree.New(killSwitch, trader, currencies...)

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
