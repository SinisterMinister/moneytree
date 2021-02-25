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
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	pairs, err := svc.moneytree.GetOpenPairs(ctx, &proto.NullRequest{})
	if err != nil {
		log.WithError(err).Error("could not get open pairs")
		return
	}
	currentPrice, movingAverage, fiveMinuteOscillator, err := svc.getFiveMinuteTrixIndicators(ctx)
	if err != nil {
		log.WithError(err).Error("could not get 5min trix indicators")
		return
	}

	// Get one minute trix oscillator
	_, _, oneMinuteOscillator, err := svc.getOneMinuteTrixIndicators(ctx)
	if err != nil {
		log.WithError(err).Error("could not get 1min trix indicators")
		return
	}

	// If the current price is above the moving average, going up
	if currentPrice.GreaterThan(movingAverage) &&
		// Make sure gaining momentum in both 1m and 5m intervals
		fiveMinuteOscillator.GreaterThan(decimal.Zero) && oneMinuteOscillator.GreaterThan(decimal.Zero) {

		// Place the upward pair
		svc.placePair(pair.Upward)

	} else if currentPrice.LessThan(movingAverage) &&
		// Make sure losing momentum
		fiveMinuteOscillator.LessThan(decimal.Zero) && oneMinuteOscillator.LessThan(decimal.Zero) {

		// Place the downward pair
		svc.placePair(pair.Downward)
	} else {
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
		} else {
			svc.placePair(pair.Upward)
		}
	}
	return
}

func (svc *Service) getFiveMinuteTrixIndicators(ctx context.Context) (currentPrice decimal.Decimal, movingAvg decimal.Decimal, oscillator decimal.Decimal, err error) {
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

	ma, osc := trix.GetTrixIndicator(rawValues, 5)
	movingAvg = decimal.NewFromFloat(ma)
	oscillator = decimal.NewFromFloat(osc)
	log.Infof("5m trix cp: %s ma: %s osc: %s", currentPrice, movingAvg, oscillator)
	return
}

func (svc *Service) getOneMinuteTrixIndicators(ctx context.Context) (currentPrice decimal.Decimal, movingAvg decimal.Decimal, oscillator decimal.Decimal, err error) {
	log.Infof("calculate trix moving average and oscillator")
	candles, err := svc.moneytree.GetCandles(ctx, &proto.GetCandlesRequest{Duration: proto.GetCandlesRequest_ONE_MINUTE, StartTime: time.Now().Add(-3 * time.Hour).Unix(), EndTime: time.Now().Unix()})
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

	ma, osc := trix.GetTrixIndicator(rawValues, 5)
	movingAvg = decimal.NewFromFloat(ma)
	oscillator = decimal.NewFromFloat(osc)
	log.Infof("1m trix cp: %s ma: %s osc: %s", currentPrice, movingAvg, oscillator)
	return
}
