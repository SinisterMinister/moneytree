package server

import (
	"github.com/go-playground/log"
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

	viper.SetDefault("postgres.host", "localhost")
	viper.SetDefault("postgres.port", "5432")
	viper.SetDefault("postgres.user", "postgres")
	viper.SetDefault("postgres.pass", "postgres")
	viper.SetDefault("postgres.database", "moneytree")
}
