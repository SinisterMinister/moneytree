package server

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/go-playground/log/v7"
	"github.com/heptiolabs/healthcheck"
	"github.com/shopspring/decimal"
	"github.com/sinisterminister/go-coinbasepro/v2"
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	"github.com/sinisterminister/currencytrader"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/candle"
	"github.com/sinisterminister/currencytrader/types/order"
	"github.com/sinisterminister/currencytrader/types/provider/coinbase"
	coinbaseclient "github.com/sinisterminister/currencytrader/types/provider/coinbase/client"
	"github.com/sinisterminister/moneytree/pkg/pair"
	"github.com/sinisterminister/moneytree/pkg/proto"

	// Load up postgres driver
	_ "github.com/lib/pq"
)

var (
	market types.Market
	trader types.Trader
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
	provider := coinbase.New(killSwitch, client, 5, 10)

	// Get an instance of the trader
	trader = currencytrader.New(provider)
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
	market, err = trader.MarketSvc().Market(btc, usd)
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
	log.Infof("received place %s pair request", in.Direction)
	orderPair, err := pair.BuildSpreadBasedPair(s.pairSvc, pair.Direction(in.Direction))
	if err != nil {
		log.WithError(err).Error("could not build pair")
		return nil, err
	}

	log.Info("looking for a colliding pair")
	openPair, err := s.pairSvc.GetCollidingOpenPair(orderPair)
	if err != nil {
		return nil, err
	}

	// Use the colliding pair instead of the provided one
	if openPair != nil {
		if openPair.FirstOrder().Status() != order.Filled {
			// Update the first order to get any missed fills
			openPair.FirstOrder().Refresh()

			// Reverse if there are no fills
			if openPair.FirstOrder().Filled().Equal(decimal.Zero) {
				log.Infof("found overlapping open pair that missed %s; canceling prior pair", openPair.UUID().String())
				err = openPair.Cancel()
				if err != nil {
					log.WithError(err).Errorf("could not cancel stale overlapping pair")
					return nil, err
				}
			} else {
				// Order must be partially filled
				orderPair = openPair
				log.Infof("found overlapping open pair; resuming %s", orderPair.UUID().String())
			}
		} else {
			orderPair = openPair
			log.Infof("found overlapping open pair; resuming %s", orderPair.UUID().String())
		}
	} else {
		log.Infof("no overlapping pair found; using new pair %s", orderPair.UUID().String())
	}

	// Try to make room if we're placing a new order
	if orderPair != openPair {
		err = s.pairSvc.MakeRoom(orderPair.FirstRequest().Price(), pair.Direction(in.Direction))
		if err != nil {
			return nil, err
		}
	}

	// Execute the pair
	err = orderPair.Execute()
	if err != nil {
		return nil, err
	}

	return &proto.PlacePairResponse{Pair: &proto.Pair{
		Uuid: orderPair.UUID().String(),
	}}, nil
}

func (s *Server) GetCandles(ctx context.Context, in *proto.GetCandlesRequest) (*proto.CandleCollection, error) {
	log.Debug("Received get candles request")

	// Deserialize the interval
	var interval types.CandleInterval
	switch in.Duration {
	case proto.GetCandlesRequest_ONE_MINUTE:
		interval = candle.OneMinute
	case proto.GetCandlesRequest_FIVE_MINUTES:
		interval = candle.FiveMinutes
	case proto.GetCandlesRequest_FIFTEEN_MINUTES:
		interval = candle.FifteenMinutes
	case proto.GetCandlesRequest_ONE_HOUR:
		interval = candle.OneHour
	case proto.GetCandlesRequest_TWELVE_HOURS:
		interval = candle.TwelveHours
	case proto.GetCandlesRequest_TWENTY_FOUR_HOURS:
		interval = candle.OneDay
	}

	// Deserialize the times
	start := time.Unix(in.StartTime, 0)
	end := time.Unix(in.EndTime, 0)

	// Fetch the candles
	log.WithFields(log.F("interval", interval), log.F("start", start), log.F("end", end)).Debug("fetching candles")
	candles, err := market.Candles(interval, start, end)
	if err != nil {
		log.WithError(err).Error("could not fetch candles")
		return nil, err
	}

	// Serialize the candles
	protoCandles := []*proto.Candle{}
	for _, candle := range candles {
		protoCandles = append(protoCandles, &proto.Candle{
			Ts:     candle.Timestamp().Unix(),
			Open:   candle.Open().String(),
			Close:  candle.Close().String(),
			High:   candle.High().String(),
			Low:    candle.Low().String(),
			Volume: candle.Volume().String(),
		})
	}
	return &proto.CandleCollection{Candles: protoCandles}, nil
}

func (s *Server) GetOpenPairs(ctx context.Context, in *proto.NullRequest) (*proto.PairCollection, error) {
	log.Debug("Received get open pairs request")
	openPairs, err := s.pairSvc.LoadOpenPairs()
	if err != nil {
		return nil, err
	}

	retPairs := []*proto.Pair{}
	for _, pair := range openPairs {
		var buyOrder, sellOrder *proto.Order
		if pair.BuyOrder() != nil {
			buyOrder = &proto.Order{
				Side:     "BUY",
				Price:    pair.BuyRequest().Price().String(),
				Quantity: pair.BuyRequest().Quantity().String(),
				Filled:   pair.BuyOrder().Filled().String(),
				Status:   string(pair.BuyOrder().Status()),
			}
		} else {
			buyOrder = &proto.Order{
				Side:     "BUY",
				Price:    pair.BuyRequest().Price().String(),
				Quantity: pair.BuyRequest().Quantity().String(),
			}
		}

		if pair.SellOrder() != nil {
			sellOrder = &proto.Order{
				Side:     "SELL",
				Price:    pair.SellRequest().Price().String(),
				Quantity: pair.SellRequest().Quantity().String(),
				Filled:   pair.SellOrder().Filled().String(),
				Status:   string(pair.SellOrder().Status()),
			}

		} else {
			sellOrder = &proto.Order{
				Side:     "SELL",
				Price:    pair.SellRequest().Price().String(),
				Quantity: pair.SellRequest().Quantity().String(),
			}
		}

		retPairs = append(retPairs, &proto.Pair{
			Uuid:          pair.UUID().String(),
			Created:       pair.CreatedAt().Unix(),
			Ended:         pair.EndedAt().Unix(),
			Direction:     string(pair.Direction()),
			Done:          pair.IsDone(),
			Status:        string(pair.Status()),
			StatusDetails: pair.StatusDetails(),
			BuyOrder:      buyOrder,
			SellOrder:     sellOrder,
		})
	}
	return &proto.PairCollection{
		Pairs: retPairs,
	}, nil
}

func (s *Server) RefreshPair(ctx context.Context, in *proto.PairRequest) (*proto.Pair, error) {
	log.Infof("received get refresh pair request for %s", in.Uuid)
	// Load the pair from the database
	op, err := s.pairSvc.Load(in.Uuid)
	if err != nil {
		return nil, err
	}

	// Refresh the orders
	if op.FirstOrder() != nil {
		op.FirstOrder().Refresh()
	}
	if op.SecondOrder() != nil {
		op.SecondOrder().Refresh()
	}
	if op.ReversalOrder() != nil {
		op.ReversalOrder().Refresh()
	}

	// Save the refreshed order
	err = op.Save()
	if err != nil {
		return nil, err
	}

	// Return the pair
	return createProtoPair(op), nil
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
