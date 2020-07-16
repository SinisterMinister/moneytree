package moneytree

import (
	"database/sql"
	"fmt"

	"github.com/go-playground/log/v7"

	// Load up postgres driver
	_ "github.com/lib/pq"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/moneytree/lib/followtheleader"
	"github.com/sinisterminister/moneytree/lib/marketwatcher"
	"github.com/sinisterminister/moneytree/lib/orderpair"
)

type Moneytree struct {
	currencies []types.Currency
	markets    map[string]types.Market
	stop       <-chan bool
	trader     types.Trader
	db         *sql.DB
}

func New(stop <-chan bool, trader types.Trader, currencies ...types.Currency) (Moneytree, error) {
	m := Moneytree{trader: trader, currencies: currencies}

	log.Info("starting database connection")
	err := m.connectToDatabase()

	log.Info("loading markets")
	m.loadMarkets()

	log.Info("starting market watchers")
	m.startMarketWatchers()

	if err != nil {
		log.WithError(err).Fatal("could not connect to database")
	}

	log.Info("moneytree started")
	return m, nil
}

func (m *Moneytree) connectToDatabase() error {
	db, err := sql.Open("postgres", getConnectionString())
	if err != nil {
		return err
	}
	m.db = db

	// Setup the orderpair db
	log.Info("setting up order pair database")
	err = orderpair.SetupDB(db)
	if err != nil {
		return fmt.Errorf("could not setup order pair database: %w", err)
	}
	return nil
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
		marketwatcher.New(m.stop, mkt, followtheleader.New(m.db, m.trader, mkt))
	}
}
