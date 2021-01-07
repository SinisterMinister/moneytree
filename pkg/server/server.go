package server

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"

	"github.com/go-playground/log/v7"
	"github.com/heptiolabs/healthcheck"
	"github.com/preichenberger/go-coinbasepro/v2"
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	"github.com/sinisterminister/currencytrader"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/provider/coinbase"
	coinbaseclient "github.com/sinisterminister/currencytrader/types/provider/coinbase/client"
	"github.com/sinisterminister/moneytree/pkg/pair"
	"github.com/sinisterminister/moneytree/pkg/proto"

	// Load up postgres driver
	_ "github.com/lib/pq"
)

func NewServer(port string) error {
	// Setup the kill switch
	killSwitch := make(chan bool)

	// Setup a coinbase client
	client := coinbaseclient.NewClient()

	// Connect to live
	client.UpdateConfig(&coinbasepro.ClientConfig{
		BaseURL:    viper.GetString("coinbase.baseUrl"),
		Key:        viper.GetString("coinbase.key"),
		Passphrase: viper.GetString("coinbase.passphrase"),
		Secret:     viper.GetString("coinbase.secret"),
	})

	// Start up a coinbase provider
	provider := coinbase.New(killSwitch, client)

	// Get an instance of the trader
	trader := currencytrader.New(provider)
	trader.Start()

	// Setup the market
	btc, err := trader.AccountSvc().Currency("BTC")
	if err != nil {
		log.WithError(err).Fatal("could not load BTC")
	}
	usd, err := trader.AccountSvc().Currency("USD")
	if err != nil {
		log.WithError(err).Fatal("could not load USD")
	}
	market, err := trader.MarketSvc().Market(btc, usd)
	if err != nil {
		log.WithError(err).Fatal("could not load market")
	}

	// Setup the server
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.WithError(err).Fatal("could not listen for connections")
		return err
	}
	s := grpc.NewServer()
	svr := &Server{}

	// Initialize server
	err = svr.init(trader, market)
	if err != nil {
		log.WithError(err).Fatal("could not initialize the server")
	}

	proto.RegisterMoneytreeServer(s, svr)

	if err := s.Serve(listener); err != nil {
		return err
	}
	return nil
}

type Server struct {
	proto.UnimplementedMoneytreeServer
	db *sql.DB

	pairSvc *pair.Service
}

func (s *Server) PlacePair(ctx context.Context, in *proto.PlacePairRequest) (*proto.PlacePairResponse, error) {
	log.Infof("received place pair request. building %s pair", in.Direction)
	orderPair, err := pair.BuildSpreadBasedPair(s.pairSvc, pair.Direction(in.Direction))
	if err != nil {
		return nil, err
	}

	log.Info("looking for a colliding pair")
	openPair, err := s.pairSvc.GetCollidingOpenPair(orderPair)
	if err != nil {
		return nil, err
	}

	// Use the colliding pair instead of the provided one
	if openPair != nil {
		log.Infof("found overlapping open pair; resuming %s", openPair.UUID().String())
		orderPair = openPair
	} else {
		s.pairSvc.MakeRoom(pair.Direction(in.Direction))
	}

	// Execute the pair
	orderPair.Execute()

	return &proto.PlacePairResponse{Pair: &proto.Pair{
		Uuid: orderPair.UUID().String(),
	}}, nil
}

func (s *Server) GetOpenPairs(ctx context.Context, in *proto.NullRequest) (*proto.PairCollection, error) {
	log.Info("Received get open pairs request")
	return &proto.PairCollection{}, nil
}

func (s *Server) init(trader types.Trader, market types.Market) (err error) {
	err = s.connectToDatabase()
	if err != nil {
		return
	}

	s.startHealthcheckHandler()
	s.pairSvc, err = pair.NewService(s.db, trader, market)
	if err != nil {
		return
	}

	// Load the open pairs
	pairs, err := s.pairSvc.LoadOpenPairs()
	if err != nil {
		return
	}

	// Kick off the execution process
	for _, pair := range pairs {
		pair.Execute()
	}
	return
}

func (s *Server) startHealthcheckHandler() {
	// Create a healthcheck.Handler
	health := healthcheck.NewHandler()

	// Our app is not happy if we've got more than 512 goroutines running.
	health.AddLivenessCheck("goroutine-threshold", healthcheck.GoroutineCountCheck(512))

	// Expose the /live and /ready endpoints over HTTP (on port 8086)
	go http.ListenAndServe("0.0.0.0:8086", health)
}

func (s *Server) connectToDatabase() error {
	log.Info("connecting to database")
	db, err := sql.Open("postgres", getDBConnectionString())
	if err != nil {
		return err
	}
	s.db = db

	return nil
}

func getDBConnectionString() string {
	// Build the connection string
	user := viper.GetString("postgres.user")
	pass := viper.GetString("postgres.password")
	host := viper.GetString("postgres.host")
	port := viper.GetString("postgres.port")
	database := viper.GetString("postgres.database")
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, database)
}
