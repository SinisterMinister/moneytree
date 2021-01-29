package trix

import "github.com/sinisterminister/moneytree/pkg/ewma"

func GetTrixIndicator(prices []float64, periods float64) (ma float64, oscillator float64) {
	singleSmoothedValues := []float64{}
	doubleSmoothedValues := []float64{}
	tripleSmoothedValues := []float64{}

	singleSmoothed := ewma.NewMovingAverage(periods)
	doubleSmoothed := ewma.NewMovingAverage(periods)
	tripleSmoothed := ewma.NewMovingAverage(periods)

	// Calculate the single smoothed moving average values
	for _, price := range prices {
		singleSmoothed.Add(price)
		if singleSmoothed.Value() != 0.0 {
			singleSmoothedValues = append(singleSmoothedValues, singleSmoothed.Value())
		}
	}

	// Calculate the double smoothed moving average values
	for _, s := range singleSmoothedValues {
		doubleSmoothed.Add(s)
		if doubleSmoothed.Value() != 0.0 {
			doubleSmoothedValues = append(doubleSmoothedValues, doubleSmoothed.Value())
		}
	}

	// Calculate the triple smoothed moving average values
	for _, s := range doubleSmoothedValues {
		tripleSmoothed.Add(s)
		if tripleSmoothed.Value() != 0.0 {
			tripleSmoothedValues = append(tripleSmoothedValues, tripleSmoothed.Value())
		}
	}

	ma = tripleSmoothed.Value()
	originalValue := tripleSmoothedValues[len(tripleSmoothedValues)-2]
	oscillator = (ma - originalValue) / originalValue

	return ma, oscillator
}
