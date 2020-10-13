package main

import (
	"github.com/go-playground/log/v7"
	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/moneytree/")
	viper.AddConfigPath("$HOME/.moneytree")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		log.WithError(err).Panic("fatal error loading config file")
	}

	// Force system to submit only market maker orders. Otherwise it will use taker orders for the first order
	viper.SetDefault("moneytree.forceMakerOrders", false)
}
