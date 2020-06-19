package marketwatcher

import (
	"github.com/go-playground/log/v7"
	"github.com/sinisterminister/currencytrader/types"
)

type MarketWatcher struct {
	stop   <-chan bool
	market types.Market
}

func New(stop <-chan bool, mkt types.Market) MarketWatcher {
	mw := MarketWatcher{stop, mkt}
	go mw.start()
	return mw
}

func (mw *MarketWatcher) start() {
	// Get the ticker stream from the market
	stream := mw.market.TickerStream(mw.stop)

	// Watch the stream and log any data sent over it
	for {
		// Bail out on stop
		select {
		case <-mw.stop:
			return
		default:
		}

		select {
		//Backup bailout
		case <-mw.stop:
			return

		// Data received
		case data := <-stream:
			if data != nil {
				log.WithField("data", data.ToDTO()).Infof("stream data received for %s market", mw.market.Name())
			} else {
				log.Info("empty stream data received")
			}
		}
	}
}
