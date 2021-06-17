package miraclegrow

import (
	"time"

	"github.com/go-playground/log/v7"
	"github.com/sinisterminister/moneytree/pkg/pair"
)

func (svc *Service) BackfillTheHoles(stop <-chan bool) (err error) {
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
			log.Infof("filling up the holes")

			// Place the upward pair
			svc.placePair(pair.Upward)

			// Place the downward pair
			svc.placePair(pair.Downward)

			ticker.Reset(svc.updateFrequency)
		case <-stop:
			// Try to bail out if necessary
			return
		}
	}
}
