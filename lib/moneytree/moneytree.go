package moneytree

import (
	"github.com/go-playground/log/v7"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/moneytree/lib/followtheleader"
	"github.com/sinisterminister/moneytree/lib/marketwatcher"
)

type Moneytree struct {
	currencies []types.Currency
	markets    map[string]types.Market
	stop       <-chan bool
	trader     types.Trader
}

func New(stop <-chan bool, trader types.Trader, currencies ...types.Currency) (Moneytree, error) {

	m := Moneytree{trader: trader, currencies: currencies}

	log.Info("loading markets")
	m.loadMarkets()

	log.Info("starting market watchers")
	m.startMarketWatchers()

	log.Info("Moneytree started...")
	return m, nil
}

func (m *Moneytree) loadMarkets() {
	// Build set of potential markets
	markets := make(map[string]types.Market)
	for _, c := range m.currencies {
		for _, cc := range m.currencies {
			if c != cc {
				mkt, err := m.trader.MarketSvc().Market(c, cc)
				if err != nil {
					continue
				}
				markets[mkt.Name()] = mkt
			}
		}
	}
	m.markets = markets
}

func (m *Moneytree) startMarketWatchers() {
	// Start the MarketWatchers
	for _, mkt := range m.markets {
		marketwatcher.New(m.stop, mkt, followtheleader.New(m.trader))
	}
}
