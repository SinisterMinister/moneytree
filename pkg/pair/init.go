package pair

import "github.com/spf13/viper"

func init() {
	// Divide the funds into this many equal trades and use that trade size as the maximum trade size
	viper.SetDefault("maxOpenPairs", 8)

	// Set the expected return per trade pair
	viper.SetDefault("targetReturn", 0.001)

	// Force system to submit only market maker orders. Otherwise it will use taker orders for the first order
	viper.SetDefault("forceMakerOrders", false)
}
