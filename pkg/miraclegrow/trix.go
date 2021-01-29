package miraclegrow

import (
	"context"
	"strconv"
	"time"

	"github.com/go-playground/log/v7"
	"github.com/shopspring/decimal"
	"github.com/sinisterminister/moneytree/pkg/pair"
	"github.com/sinisterminister/moneytree/pkg/proto"
	"github.com/sinisterminister/moneytree/pkg/trix"
)

func (svc *Service) TrixR5Kids(stop <-chan bool) (err error) {
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
			// TURN TRIX BABY
			log.Infof("pouring a bowl")
			err = svc.turnTrix()
			if err != nil {
				log.WithError(err).Error("something happened while pouring")
			}
			ticker.Reset(svc.updateFrequency)
		case <-stop:
			// Try to bail out if necessary
			return
		}
	}
}

func (svc *Service) turnTrix() (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	currentPrice, movingAverage, oscillator, err := svc.getCurrentTrixIndicators(ctx)
	if err != nil {
		log.WithError(err).Error("could not get trix indicators")
		return
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
	return
}

func (svc *Service) getCurrentTrixIndicators(ctx context.Context) (currentPrice decimal.Decimal, movingAvg decimal.Decimal, oscillator decimal.Decimal, err error) {
	log.Infof("calculate trix moving average and oscillator")
	candles, err := svc.moneytree.GetCandles(ctx, &proto.GetCandlesRequest{Duration: proto.GetCandlesRequest_FIVE_MINUTES, StartTime: time.Now().Add(-3 * time.Hour).Unix(), EndTime: time.Now().Unix()})
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

	ma, osc := trix.GetTrixIndicator(rawValues, 6)
	movingAvg = decimal.NewFromFloat(ma)
	oscillator = decimal.NewFromFloat(osc)
	log.Infof("cp: %s ma: %s osc: %s", currentPrice, movingAvg, oscillator)
	return
}
