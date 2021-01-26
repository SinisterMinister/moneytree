package miraclegrow

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/go-playground/log/v7"
	"github.com/heptiolabs/healthcheck"
	"github.com/shopspring/decimal"
	"github.com/sinisterminister/moneytree/pkg/pair"
	"github.com/sinisterminister/moneytree/pkg/proto"
	"github.com/sinisterminister/moneytree/pkg/trix"
	"google.golang.org/grpc"
)

type Service struct {
	moneytree       proto.MoneytreeClient
	updateFrequency time.Duration
}

func NewService(address string, updateFrequency time.Duration) (svc *Service) {
	// Set up a connection to the server.
	log.Infof("connecting to %s...", address)
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	// Setup the service
	svc = &Service{proto.NewMoneytreeClient(conn), updateFrequency}
	return
}

func (svc *Service) MakeItGrow(stop <-chan bool) (err error) {
	svc.startHealthcheckHandler()
	ticker := time.NewTimer(1)
	for {
		select {
		case <-stop:
			// Bail out immediately
			return
		default:
			// Continue on
		}
		select {
		case <-ticker.C:
			// WATER THE MONEYTREE
			log.Infof("water the moneytree")
			err = svc.startWatering()
			if err != nil {
				return err
			}
			ticker.Reset(svc.updateFrequency)
		case <-stop:
			// Try to bail out if necessary
			return
		}
	}
}

func (svc *Service) startWatering() (err error) {
	log.Infof("get all the open pairs")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	pairs, err := svc.moneytree.GetOpenPairs(ctx, &proto.NullRequest{})
	if err != nil {
		log.WithError(err).Error("could not get open pairs")
		return
	}

	// Get the the direction with the least number of pairs
	var upCount, downCount int
	for _, p := range pairs.GetPairs() {
		switch pair.Direction(p.Direction) {
		case pair.Upward:
			if p.BuyOrder.Status == "FILLED" {
				upCount++
			}
		case pair.Downward:
			if p.SellOrder.Status == "FILLED" {
				downCount++
			}
		}
	}
	log.Infof("pair counts - total: %d, up: %d, down: %d", len(pairs.GetPairs()), upCount, downCount)
	if upCount > downCount {
		svc.placePair(pair.Downward)
	} else if upCount < downCount {
		svc.placePair(pair.Upward)
	} else {
		// We'll place the pair based on the trix indicators
		currentPrice, movingAverage, oscillator, err := svc.getCurrentTrixIndicators(ctx)
		if err != nil {
			log.WithError(err).Error("could not get trix indicators")
		}

		// If the current price is above the moving average, going up
		if currentPrice.GreaterThan(movingAverage) {
			// Make sure gaining momentum
			if oscillator.GreaterThanOrEqual(decimal.Zero) {
				// Place the upward pair
				svc.placePair(pair.Upward)
			}
		}

		// If the current price is below the moving average, going down
		if currentPrice.LessThan(movingAverage) {
			// Make sure losing momentum
			if oscillator.LessThanOrEqual(decimal.Zero) {
				// Place the downward pair
				svc.placePair(pair.Downward)
			}
		}
	}

	return
}

func (svc *Service) placePair(direction pair.Direction) (err error) {
	log.Infof("placing %s pair", direction)
	// Place the pair based on the direction
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	response, err := svc.moneytree.PlacePair(ctx, &proto.PlacePairRequest{Direction: string(direction)})
	if err != nil {
		log.Fatalf("could not place pair: %v", err)
		return
	}
	log.Infof("Pair returned: %s %s", direction, response.GetPair().Uuid)
	return
}

func (svc *Service) startHealthcheckHandler() {
	// Create a healthcheck.Handler
	health := healthcheck.NewHandler()

	// Our app is not happy if we've got more than 512 goroutines running.
	health.AddLivenessCheck("goroutine-threshold", healthcheck.GoroutineCountCheck(512))

	// Expose the /live and /ready endpoints over HTTP (on port 8086)
	go http.ListenAndServe("0.0.0.0:8086", health)
}

func (svc *Service) getCurrentTrixIndicators(ctx context.Context) (currentPrice decimal.Decimal, movingAvg decimal.Decimal, oscillator decimal.Decimal, err error) {
	log.Infof("calculate trix moving average and oscillator")
	candles, err := svc.moneytree.GetCandles(ctx, &proto.GetCandlesRequest{Duration: proto.GetCandlesRequest_ONE_MINUTE, StartTime: time.Now().Add(-1 * time.Hour).Unix(), EndTime: time.Now().Unix()})
	if err != nil {
		log.WithError(err).Error("could not get candles")
		return
	}

	// Set the current price
	currentPrice, err = decimal.NewFromString(candles.GetCandles()[0].Close)
	if err != nil {
		log.WithError(err).Error("could not parse current price")
		return
	}

	// Convert the candle values to floats so we can use them while reversing the sort
	rawValues := []float64{}
	for _, can := range candles.GetCandles() {
		val, e := strconv.ParseFloat(can.Close, 64)
		if e != nil {
			return
		}
		rawValues = append([]float64{val}, rawValues...)
	}

	ma, osc := trix.GetTrixIndicator(rawValues)
	movingAvg = decimal.NewFromFloat(ma)
	oscillator = decimal.NewFromFloat(osc)
	log.Infof("cp: %s ma: %s osc: %s", currentPrice, movingAvg, oscillator)
	return
}
