package marketwatcher

import (
	"time"

	"github.com/go-playground/log/v7"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/moneytree/lib/marketprocessor"
)

type MarketWatcher struct {
	stop      <-chan bool
	market    types.Market
	processor marketprocessor.Processor
}

func New(stop <-chan bool, mkt types.Market, processor marketprocessor.Processor) MarketWatcher {
	mw := MarketWatcher{stop, mkt, processor}
	go mw.watchMarket()
	return mw
}

func (mw *MarketWatcher) watchMarket() {
	// Make phat stacks
	for {
		// Try receive op of stop to bail on close instead of random channel receive
		select {
		case <-mw.stop:
			return
		default:
		}

		done, err := mw.processor.ProcessMarket(mw.stop, mw.market)
		if err != nil {
			log.WithError(err).Error("error processing market")
			return
		}

		// Wait for the processor to complete
		select {
		case <-mw.stop:
			return
		case <-done:
			log.WithField("market", mw.market.Name()).Info("market process cycle complete")

			// Wait a sec
			<-time.NewTicker(time.Second).C
		}
	}
}
