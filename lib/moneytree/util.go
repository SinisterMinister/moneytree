package moneytree

import (
	"fmt"

	"github.com/spf13/viper"
)

func getConnectionString() string {
	// Build the connection string
	user := viper.GetString("postgres.user")
	pass := viper.GetString("postgres.password")
	host := viper.GetString("postgres.host")
	port := viper.GetString("postgres.port")
	database := viper.GetString("postgres.database")
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, database)
}
