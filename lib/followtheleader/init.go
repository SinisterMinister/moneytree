package followtheleader

import "github.com/spf13/viper"

func init() {
	viper.SetDefault("followtheleader.orderTTL", "5m")
	viper.SetDefault("followtheleader.waitAfterCancelStalledPair", "5s")
}
