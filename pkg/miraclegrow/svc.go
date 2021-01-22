package miraclegrow

import (
	"time"

	"github.com/go-playground/log/v7"
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

func (svc *Service) Grow(stop <-chan bool) (err error) {
	ticker := time.NewTicker(svc.updateFrequency)
	for {

		select {
		// Bail out immediately
		case <-stop:
			return
		default:
		}
		select {
		case <-ticker.C:
			log.Infof("I'M GROWIN SO HARD RIGHT NOW")
		case <-stop:
		}
	}
	return
}
