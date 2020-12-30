package pair

import "github.com/spf13/viper"

func init() {
	viper.SetDefault("orderpair.criticalPairRefreshInterval", "1m")
}
