package moneytree

import (
	"database/sql"
	"net/http"

	"github.com/go-playground/log/v7"
	"github.com/heptiolabs/healthcheck"

	// Load up postgres driver
	_ "github.com/lib/pq"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/moneytree/lib/chaser"
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
	if err != nil {
		log.WithError(err).Fatal("could not connect to database")
	}

	log.Info("loading markets")
	m.loadMarkets()

	log.Info("starting market processors")
	m.startMarketProcessors()

	log.Info("starting healthcheck endoints")
	m.startHealthcheck()

	log.Info("moneytree started")
	return m, nil
}

func (m *Moneytree) connectToDatabase() error {
	db, err := sql.Open("postgres", getConnectionString())
	if err != nil {
		return err
	}
	m.db = db

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

func (m *Moneytree) startMarketProcessors() {
	// Start the processors
	for _, mkt := range m.markets {
		processor := &chaser.Processor{}
		go processor.Process(m.db, m.trader, mkt, m.stop)
	}
}

func (m *Moneytree) startHealthcheck() {
	// Create a healthcheck.Handler
	health := healthcheck.NewHandler()

	// Our app is not happy if we've got more than 100 goroutines running.
	health.AddLivenessCheck("goroutine-threshold", healthcheck.GoroutineCountCheck(256))

	// Expose the /live and /ready endpoints over HTTP (on port 8086)
	go http.ListenAndServe("0.0.0.0:8086", health)
}
