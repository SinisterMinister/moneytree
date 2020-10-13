package followtheleader

import (
	"github.com/spf13/viper"
)

func init() {
	// Spread distance the price must change to reverse the order direction
	viper.SetDefault("followtheleader.reversalSpread", 0.00075)

	// Divide the funds into this many equal trades and use that trade size as the maximum trade size
	viper.SetDefault("followtheleader.maxOpenOrders", 8)

	// Set the expected return per trade pair
	viper.SetDefault("followtheleader.targetReturn", 0.001)

	// Set the delay between each order cycle
	viper.SetDefault("followtheleader.cycleDelay", "5s")

	// Don't refresh pairs by default
	viper.SetDefault("followtheleader.refreshDatabasePairs", false)
}
