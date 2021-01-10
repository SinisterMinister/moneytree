package server

import (
	"os"

	"github.com/go-playground/log/v7"
	"github.com/go-playground/log/v7/handlers/console"
	"github.com/go-playground/log/v7/handlers/json"
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

	// Setup json logging for containers
	if _, err := os.Stat("/.dockerenv"); err == nil {
		// Setup the console logger
		log.AddHandler(json.New(os.Stdout), log.InfoLevel, log.WarnLevel, log.ErrorLevel, log.NoticeLevel, log.FatalLevel, log.AlertLevel, log.PanicLevel)
		if viper.GetBool("debug") {
			log.AddHandler(json.New(os.Stdout), log.DebugLevel)
		}
	} else {
		// Setup the console logger
		log.AddHandler(console.New(true), log.InfoLevel, log.WarnLevel, log.ErrorLevel, log.NoticeLevel, log.FatalLevel, log.AlertLevel, log.PanicLevel)
		if viper.GetBool("debug") {
			log.AddHandler(console.New(true), log.DebugLevel)
		}
	}
}
