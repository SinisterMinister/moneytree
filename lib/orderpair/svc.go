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
)

type Service struct {
	trader types.Trader
	market types.Market
	db     *sql.DB

	mutex     sync.RWMutex
	openPairs map[string]*OrderPair
}

func NewService(db *sql.DB, trader types.Trader, market types.Market) (svc *Service, err error) {
	svc = &Service{
		db:     db,
		trader: trader,
		market: market,
	}
	err = svc.setupDB()
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

func (svc *Service) RegisterOpenPair(pair *OrderPair) {
	svc.mutex.Lock()
	defer svc.mutex.Unlock()

	// Add pair to open pairs
	svc.openPairs[pair.UUID().String()] = pair

	// Remove the pair from open when it closes out
	go func(pair *OrderPair) {
		// Wait for the pair to close
		<-pair.Done()

		// Wait for the first order if it exists
		if pair.FirstOrder() != nil {
			<-pair.FirstOrder().Done()
		}

		// Wait for the second order if it exists
		if pair.SecondOrder() != nil {
			<-pair.SecondOrder().Done()
		}

		// Remove the order from open orders
		svc.mutex.Lock()
		delete(svc.openPairs, pair.UUID().String())
		svc.mutex.Unlock()
	}(pair)
}

func (svc *Service) OpenPair(uuid string) (*OrderPair, bool) {
	svc.mutex.RLock()
	defer svc.mutex.RUnlock()
	pair, ok := svc.openPairs[uuid]

	return pair, ok
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

		// Use existing open pair if already loaded
		var pair *OrderPair
		if openPair, ok := svc.OpenPair(dao.Uuid); ok {
			pair = openPair
		} else {
			// Throttle calls to API
			<-time.NewTimer(time.Second).C

			// Load the pair from the database and API
			pair, err = svc.NewFromDAO(dao)
			if err != nil {
				return nil, fmt.Errorf("could not load open order: %w", err)
			}

			// Add pair to open pairs
			svc.RegisterOpenPair(pair)
		}

		// Add to return
		pairs = append(pairs, pair)
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

	return orderPair, nil
}

func (svc *Service) NewFromDAO(dao OrderPairDAO) (orderPair *OrderPair, err error) {
	id, err := uuid.FromString(dao.Uuid)
	if err != nil {
		return nil, fmt.Errorf("could not parse order pair ID: %w", err)
	}
	done := make(chan bool)
	if dao.Done {
		close(done)
	}

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

	if dao.FirstOrder.ID != "" {
		order, err := svc.trader.OrderSvc().Order(svc.market, dao.FirstOrder.ID)
		if err != nil {
			if strings.Contains(err.Error(), "NotFound") {
				log.Warnf("could not load first order %s, closing as failed", dao.FirstOrder.ID)
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

	if dao.SecondOrder.ID != "" {
		order, err := svc.trader.OrderSvc().Order(svc.market, dao.SecondOrder.ID)
		if err != nil {
			if strings.Contains(err.Error(), "NotFound") {
				log.Warnf("could not load second order %s, ignoring", dao.SecondOrder.ID)
			} else {
				return nil, err
			}
		}
		orderPair.secondOrder = order
	}

	if dao.ReversalOrder.ID != "" {
		order, err := svc.trader.OrderSvc().Order(svc.market, dao.ReversalOrder.ID)
		if err != nil {
			if strings.Contains(err.Error(), "NotFound") {
				log.Warnf("could not load reversal order %s, ignoring", dao.ReversalOrder.ID)
			} else {
				return nil, err
			}
		}
		orderPair.reversalOrder = order
	}
	svc.Save(orderPair.ToDAO())
	return orderPair, nil
}

func (svc *Service) RefreshDatabasePairs() error {
	rows, err := svc.db.Query("SELECT data FROM orderpairs ORDER BY data->>'createdAt' DESC")
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
		<-time.NewTimer(time.Second).C
	}
	return nil
}

func (svc *Service) CollidingOpenPair(newPair *OrderPair) (pair *OrderPair, err error) {
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
				return
			}
		}
	}
	return
}
