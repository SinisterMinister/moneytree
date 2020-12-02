package orderpair

import "github.com/spf13/viper"

func init() {
	viper.SetDefault("orderpair.missPercentage", 0.05)
	viper.SetDefault("orderpair.criticalPairRefreshInterval", "1m")
}
