package orderpair

import "github.com/spf13/viper"

func init() {
	viper.SetDefault("orderpair.missDistance", 0.001)
	viper.SetDefault("followtheleader.waitAfterCancelStalledPair", "5s")
}
