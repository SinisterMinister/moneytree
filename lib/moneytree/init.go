package moneytree

import "github.com/spf13/viper"

func init() {
	viper.SetDefault("postgres.host", "localhost")
	viper.SetDefault("postgres.port", "5432")
	viper.SetDefault("postgres.user", "postgres")
	viper.SetDefault("postgres.pass", "postgres")
	viper.SetDefault("postgres.database", "moneytree")
}
