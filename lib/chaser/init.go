package chaser

import (
	"github.com/go-playground/log"
	"github.com/spf13/viper"
)

func init() {
	viper.SetDefault("chaser.orderTTL", "5m")
	viper.SetDefault("chaser.failSpreadPercentage", 0.25)
	viper.SetDefault("chaser.waitAfterCancelStalledPair", "5s")

	// Divide the funds into this many equal trades and use that trade size as the maximum trade size
	viper.SetDefault("chaser.maxTradesFundsRatio", 4)

	// Set the expected return per trade pair
	viper.SetDefault("chaser.targetReturn", 0.001)

	log.WithDefaultFields(log.F("source", "chaser"))
	viper.SetDefault("chaser.cycleDelay", "5s")
}
