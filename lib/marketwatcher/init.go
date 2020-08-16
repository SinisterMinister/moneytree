package marketwatcher

import "github.com/spf13/viper"

func init() {
	viper.SetDefault("marketwatcher.marketCycleDelay", "5s")
}
