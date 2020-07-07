package main

import (
	"os"
	"os/signal"

	"github.com/sinisterminister/moneytree/lib/moneytree"

	"github.com/go-playground/log/v7"
	"github.com/go-playground/log/v7/handlers/console"
	"github.com/preichenberger/go-coinbasepro/v2"
	"github.com/sinisterminister/currencytrader"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/provider/coinbase"
	coinbaseclient "github.com/sinisterminister/currencytrader/types/provider/coinbase/client"
	"github.com/spf13/viper"
)

func main() {
	// Setup the console logger
	log.AddHandler(console.New(true), log.InfoLevel, log.WarnLevel, log.ErrorLevel, log.NoticeLevel, log.FatalLevel, log.AlertLevel, log.PanicLevel)

	// Setup the kill switch
	killSwitch := make(chan bool)

	// Setup a coinbase client
	client := coinbaseclient.NewClient()

	// Connect to live
	client.UpdateConfig(&coinbasepro.ClientConfig{
		BaseURL:    viper.GetString("coinbase.baseUrl"),
		Key:        viper.GetString("coinbase.key"),
		Passphrase: viper.GetString("coinbase.passphrase"),
		Secret:     viper.GetString("coinbase.secret"),
	})

	// Start up a coinbase provider
	provider := coinbase.New(killSwitch, client)

	// Get an instance of the trader
	trader := currencytrader.New(provider)
	trader.Start()

	// Prepare the currencies
	currencies := []types.Currency{}
	symbols := viper.GetStringSlice("symbols")

	// Load the currencies
	for _, s := range symbols {
		cur, err := trader.AccountSvc().Currency(s)
		if err != nil {
			log.WithError(err).Error("could not get %s", s)
			continue
		}
		currencies = append(currencies, cur)
	}

	// Start a new moneytree
	moneytree.New(killSwitch, trader, currencies...)

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
