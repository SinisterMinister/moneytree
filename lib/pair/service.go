package pair

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/go-playground/log"
	uuid "github.com/satori/go.uuid"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/order"
)

// Service manages the order pairs in the system
type Service struct {
	trader types.Trader
	market types.Market
	db     *sql.DB

	mutex sync.RWMutex
	pairs map[uuid.UUID]*OrderPair
}

// NewService creates a Service for use. Will initialize the database if it hasn't been already.
func NewService(db *sql.DB, trader types.Trader, market types.Market, stop <-chan bool) (svc *Service, err error) {
	svc = &Service{
		db:     db,
		trader: trader,
		market: market,
		pairs:  make(map[uuid.UUID]*OrderPair),
	}
	err = svc.initializeDB()

	return
}

func (svc *Service) New(first types.OrderRequest, second types.OrderRequest) (orderPair *OrderPair, err error) {
	id := uuid.NewV1()

	orderPair = &OrderPair{
		svc:           svc,
		uuid:          id,
		done:          make(chan bool),
		stop:          make(chan bool),
		firstRequest:  first,
		secondRequest: second,
		createdAt:     time.Now(),
		status:        New,
	}

	// Validate DTOs
	err = orderPair.validate()
	if err != nil {
		return nil, err
	}

	// Cache pair
	svc.mutex.Lock()
	svc.pairs[id] = orderPair
	svc.mutex.Unlock()

	return orderPair, nil
}

func (svc *Service) NewFromDAO(dao OrderPairDAO) (*OrderPair, error) {
	id, err := uuid.FromString(dao.Uuid)
	if err != nil {
		return nil, fmt.Errorf("could not parse order pair ID: %w", err)
	}

	// Lock up the mutex while we create the pair
	svc.mutex.Lock()
	defer svc.mutex.Unlock()

	// Try to get the cached pair
	orderPair, ok := svc.pairs[id]

	// Return the cached pair if it exists. We assume the live object is more up to date than the database.
	if !ok {
		log.Debugf("could not find pair %s in cache. building new instance", id.String())

		// Setup the done channel
		done := make(chan bool)
		if dao.Done {
			close(done)
		}

		// Setup the pair
		orderPair = &OrderPair{
			svc:           svc,
			uuid:          id,
			done:          done,
			createdAt:     dao.CreatedAt,
			endedAt:       dao.EndedAt,
			status:        dao.Status,
			firstRequest:  order.NewRequestFromDTO(svc.market, dao.FirstRequest),
			secondRequest: order.NewRequestFromDTO(svc.market, dao.SecondRequest),
		}

		// Load the first order if it's been placed
		if dao.FirstOrder.ID != "" {
			order, err := svc.trader.OrderSvc().Order(svc.market, dao.FirstOrder.ID)
			if err != nil {
				return nil, fmt.Errorf("could not load first order: %w", err)
			}
			orderPair.firstOrder = order
		}

		// Load the second order if it's been placed
		if dao.SecondOrder.ID != "" {
			order, err := svc.trader.OrderSvc().Order(svc.market, dao.SecondOrder.ID)
			if err != nil {
				return nil, fmt.Errorf("could not load second order: %w", err)
			}
			orderPair.secondOrder = order
		}

		// Load the reversal order if it's been placed
		if dao.ReversalOrder.ID != "" {
			order, err := svc.trader.OrderSvc().Order(svc.market, dao.ReversalOrder.ID)
			if err != nil {
				return nil, fmt.Errorf("could not load reversal order: %w", err)
			}
			orderPair.reversalOrder = order
		}

		// Save the pair with the latest data
		svc.Save(orderPair.ToDAO())

		// Cache the pair
		svc.pairs[id] = orderPair
	}

	return orderPair, nil
}

func (svc *Service) Save(dao OrderPairDAO) (err error) {
	log.WithField("dao", dao).Debug("saving order pair")
	_, err = svc.db.Exec("INSERT INTO orderpairs (uuid, data) VALUES ($1, $2) ON CONFLICT (uuid) DO UPDATE SET data = $2;", dao.Uuid, dao)
	if err != nil {
		err = fmt.Errorf("could not insert into database: %w", err)
	}
	return
}

func (svc *Service) Load(id string) (pair *OrderPair, err error) {
	dao := OrderPairDAO{}
	err = svc.db.QueryRow("SELECT data FROM orderpairs WHERE uuid = $1;", id).Scan(&dao)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, fmt.Errorf("could not load order pair from database: %w", err)
		}
	}

	pair, err = svc.NewFromDAO(dao)
	return
}

func (svc *Service) LoadMostRecentPair() (pair *OrderPair, err error) {
	dao := OrderPairDAO{}
	err = svc.db.QueryRow("SELECT data FROM orderpairs ORDER BY data->>'createdAt' DESC LIMIT 1").Scan(&dao)
	if err != nil {
		return nil, fmt.Errorf("could not load order pair from database: %w", err)
	}

	pair, err = svc.NewFromDAO(dao)
	return
}

func (svc *Service) LoadOpenPairs() (pairs []*OrderPair, err error) {
	pairs = []*OrderPair{}
	rows, err := svc.db.Query("SELECT data FROM orderpairs WHERE data->>'status' = 'OPEN' ORDER BY data->>'createdAt'")
	if err != nil {
		return nil, fmt.Errorf("could not load open order pairs from database: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		dao := OrderPairDAO{}
		err = rows.Scan(&dao)
		if err != nil {
			return nil, fmt.Errorf("could not load open order pair from database: %w", err)
		}

		// Load the pair
		pair, err := svc.NewFromDAO(dao)
		if err != nil {
			return nil, fmt.Errorf("could not load open order: %w", err)
		}
		// Add to return
		pairs = append(pairs, pair)

		// Throttle calls to API
		<-time.NewTimer(time.Second).C
	}
	return
}

func (svc *Service) initializeDB() error {
	_, err := svc.db.Exec("CREATE TABLE IF NOT EXISTS orderpairs (uuid char(36) primary key, data JSONB);")
	if err != nil {
		return err
	}
	return nil
}
