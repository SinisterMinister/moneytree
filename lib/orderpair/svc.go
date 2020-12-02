package orderpair

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-playground/log/v7"
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/order"
	"github.com/spf13/viper"
)

type Service struct {
	trader types.Trader
	market types.Market
	db     *sql.DB

	mutex sync.RWMutex
	pairs map[uuid.UUID]*OrderPair
}

func NewService(db *sql.DB, trader types.Trader, market types.Market, stop <-chan bool) (svc *Service, err error) {
	svc = &Service{
		db:     db,
		trader: trader,
		market: market,
		pairs:  make(map[uuid.UUID]*OrderPair),
	}
	err = svc.setupDB()

	// Start the pair janitor
	go svc.pairJanitor(stop)
	return
}

func (svc *Service) setupDB() error {
	_, err := svc.db.Exec("CREATE TABLE IF NOT EXISTS orderpairs (uuid char(36) primary key, data JSONB);")
	if err != nil {
		return err
	}
	return nil
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

func (svc *Service) LoadMostRecentRunningPair() (pair *OrderPair, err error) {
	dao := OrderPairDAO{}
	err = svc.db.QueryRow("SELECT data FROM orderpairs WHERE data->>'status' = 'OPEN' AND (data->>'done')::boolean = FALSE ORDER BY data->>'createdAt' DESC LIMIT 1").Scan(&dao)
	if err != nil {
		return nil, fmt.Errorf("could not load order pair from database: %w", err)
	}

	pair, err = svc.NewFromDAO(dao)
	return
}

func (svc *Service) LoadMostRecentOpenPair() (pair *OrderPair, err error) {
	// Load up the open pairs from cache
	pairs, err := svc.LoadOpenPairs()
	if err != nil {
		return
	}

	// Find the newest one
	for _, p := range pairs {
		if pair == nil {
			pair = p
			continue
		}

		if pair.CreatedAt().Before(p.CreatedAt()) {
			pair = p
		}
	}

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

func (svc *Service) LowestOpenBuyFirstPrice() (price decimal.Decimal, err error) {
	dao := OrderPairDAO{}
	err = svc.db.QueryRow("SELECT data FROM orderpairs WHERE data->>'status' = 'OPEN' AND data->'firstRequest'->>'side' = 'BUY' AND (data->>'done')::boolean = TRUE ORDER BY data->'firstRequest'->>'price' LIMIT 1").Scan(&dao)
	if err != nil {
		return decimal.Zero, fmt.Errorf("could not load order pair from database: %w", err)
	}
	price = dao.FirstRequest.Price

	return
}

func (svc *Service) HighestOpenSellFirstPrice() (price decimal.Decimal, err error) {
	dao := OrderPairDAO{}
	err = svc.db.QueryRow("SELECT data FROM orderpairs WHERE data->>'status' = 'OPEN' AND data->'firstRequest'->>'side' = 'SELL' AND (data->>'done')::boolean = TRUE ORDER BY data->'firstRequest'->>'price' DESC LIMIT 1").Scan(&dao)
	if err != nil {
		return decimal.Zero, fmt.Errorf("could not load order pair from database: %w", err)
	}
	price = dao.FirstRequest.Price

	return
}

func (svc *Service) Save(dao OrderPairDAO) (err error) {
	log.WithField("dao", dao).Debug("saving order pair")
	_, err = svc.db.Exec("INSERT INTO orderpairs (uuid, data) VALUES ($1, $2) ON CONFLICT (uuid) DO UPDATE SET data = $2;", dao.Uuid, dao)
	if err != nil {
		err = fmt.Errorf("could not insert into database: %w", err)
	}
	return
}

func (svc *Service) New(first types.OrderRequest, second types.OrderRequest) (orderPair *OrderPair, err error) {
	id := uuid.NewV1()
	orderPair = &OrderPair{
		svc:           svc,
		uuid:          id,
		done:          make(chan bool),
		startHold:     make(chan bool),
		firstRequest:  first,
		secondRequest: second,
		createdAt:     time.Now(),
		status:        Open,
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

	// Lock up the mutex while we create the  pair
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
			startHold:     make(chan bool),
			firstRequest:  order.NewRequestFromDTO(svc.market, dao.FirstRequest),
			secondRequest: order.NewRequestFromDTO(svc.market, dao.SecondRequest),
			createdAt:     dao.CreatedAt,
			endedAt:       dao.EndedAt,
			status:        dao.Status,
		}
	}

	// Load the first order if it's been placed
	if dao.FirstOrder.ID != "" {
		order, err := svc.trader.OrderSvc().Order(svc.market, dao.FirstOrder.ID)
		if err != nil {
			if strings.Contains(err.Error(), "NotFound") {
				log.Errorf("could not load first order %s, closing as failed", dao.FirstOrder.ID)
				orderPair.failed = true
				orderPair.status = Failed
				select {
				case <-orderPair.done:
				default:
					close(orderPair.done)
				}
			} else {
				return nil, err
			}
		}
		orderPair.firstOrder = order
	}

	// Load the second order if it's been placed
	if dao.SecondOrder.ID != "" {
		order, err := svc.trader.OrderSvc().Order(svc.market, dao.SecondOrder.ID)
		if err != nil {
			if strings.Contains(err.Error(), "NotFound") {
				log.Errorf("could not load second order %s, ignoring", dao.SecondOrder.ID)
			} else {
				return nil, err
			}
		}
		orderPair.secondOrder = order
	}

	// Load the reversal order if it's been placed
	if dao.ReversalOrder.ID != "" {
		order, err := svc.trader.OrderSvc().Order(svc.market, dao.ReversalOrder.ID)
		if err != nil {
			if strings.Contains(err.Error(), "NotFound") {
				log.Errorf("could not load reversal order %s, ignoring", dao.ReversalOrder.ID)
			} else {
				return nil, err
			}
		}
		orderPair.reversalOrder = order
	}

	// Save the pair with the latest data
	svc.Save(orderPair.ToDAO())

	// Cache the pair
	svc.pairs[id] = orderPair

	return orderPair, nil
}

func (svc *Service) ResumeCollidingOpenPair(newPair *OrderPair) (pair *OrderPair, err error) {
	// Get the pairs from cache
	pairs, err := svc.LoadOpenPairs()
	if err != nil {
		return
	}

	// Search the pairs for a colliding one
	for _, p := range pairs {
		// Only check the ones going in the same direction
		if newPair.FirstRequest().Side() == p.FirstRequest().Side() {
			buyPrice := newPair.BuyRequest().Price()
			sellPrice := newPair.SellRequest().Price()
			lower := p.BuyRequest().Price()
			upper := p.SellRequest().Price()

			// If the buy or the sell of the new order is between the buy and sell of the order, it's colliding
			if (buyPrice.GreaterThanOrEqual(lower) && buyPrice.LessThanOrEqual(upper)) ||
				(sellPrice.GreaterThanOrEqual(lower) && sellPrice.LessThanOrEqual(upper)) {

				// Return colliding pair
				pair = p
				pair.done = make(chan bool)

				// Save pair
				pair.Save()

				return
			}
		}
	}
	return
}

func (svc *Service) refreshUnfinishedPairs() error {
	rows, err := svc.db.Query("SELECT data FROM orderpairs WHERE data->>'status' NOT IN ('OPEN', 'FAILED') and (data->'firstOrder'->>'status' in ('PENDING', 'UNKNOWN', 'PARTIAL') or data->'secondOrder'->>'status' in ('PENDING', 'UNKNOWN', 'PARTIAL') or data->'reversalOrder'->>'status' in ('PENDING', 'UNKNOWN', 'PARTIAL')) ORDER BY data->>'createdAt' DESC")
	if err != nil {
		return fmt.Errorf("could not load order pairs from database: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		dao := OrderPairDAO{}
		err = rows.Scan(&dao)
		if err != nil {
			return fmt.Errorf("could not load order pair from database: %w", err)
		}
		_, err := svc.NewFromDAO(dao)
		if err != nil {
			return fmt.Errorf("could not load order: %w", err)
		}

		// Throttle calls to API
		<-time.NewTimer(5 * time.Second).C
	}
	return nil
}

func (svc *Service) pairJanitor(stop <-chan bool) {
	for {
		// Bail out if we're stopping
		select {
		case <-stop:
			return
		default:
		}

		svc.refreshUnfinishedPairs()

		// Time to wait between refreshes
		<-time.NewTimer(viper.GetDuration("orderpair.criticalPairRefreshInterval")).C
	}
}
