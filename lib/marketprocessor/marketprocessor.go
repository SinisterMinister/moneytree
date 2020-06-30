package marketprocessor

import "github.com/sinisterminister/currencytrader/types"

type Processor interface {
	ProcessMarket(stop <-chan bool, market types.Market) (done <-chan bool, err error)
}
