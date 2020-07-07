package followtheleader

import "github.com/spf13/viper"

func init() {
	viper.SetDefault("followtheleader.orderTTL", "5m")
	viper.SetDefault("followtheleader.waitAfterCancelStalledPair", "5s")

	// Divide the funds into this many equal trades and use that trade size as the maximum trade size
	viper.SetDefault("followtheleader.maxTradesFundsRatio", 4)
}
