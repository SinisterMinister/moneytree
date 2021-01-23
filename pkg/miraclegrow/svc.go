package miraclegrow

import (
	"context"
	"time"

	"github.com/go-playground/log/v7"
	"github.com/sinisterminister/moneytree/pkg/pair"
	"github.com/sinisterminister/moneytree/pkg/proto"
	"google.golang.org/grpc"
)

type Service struct {
	moneytree       proto.MoneytreeClient
	updateFrequency time.Duration
}

func NewService(address string, updateFrequency time.Duration) (svc *Service) {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	// Setup the service
	svc = &Service{proto.NewMoneytreeClient(conn), updateFrequency}
	return
}

func (svc *Service) MakeItGrow(stop <-chan bool) (err error) {
	ticker := time.NewTicker(svc.updateFrequency)
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
		log.Fatalf("could not greet: %v", err)
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
		// We need to place both
		svc.placePair(pair.Upward)
		svc.placePair(pair.Downward)
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
