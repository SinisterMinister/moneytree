package pair

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/shopspring/decimal"

	"github.com/go-playground/log/v7"
	uuid "github.com/satori/go.uuid"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/order"
	"github.com/spf13/viper"
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
func NewService(db *sql.DB, trader types.Trader, market types.Market) (svc *Service, err error) {
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
	id := uuid.NewV4()
	dir := Upward

	if first.Side() == order.Sell {
		dir = Downward
	}

	orderPair = &OrderPair{
		svc:           svc,
		uuid:          id,
		done:          make(chan bool),
		stop:          make(chan bool),
		ready:         make(chan bool),
		firstRequest:  first,
		secondRequest: second,
		createdAt:     time.Now(),
		status:        New,
		direction:     dir,
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
		log.Infof("could not find pair %s in cache. building new instance", id.String())

		// Setup the done channel
		done := make(chan bool)
		if dao.Done {
			close(done)
		}

		// Setup the pair
		orderPair = &OrderPair{
			svc:             svc,
			uuid:            id,
			createdAt:       dao.CreatedAt,
			endedAt:         dao.EndedAt,
			direction:       dao.Direction,
			done:            done,
			ready:           make(chan bool),
			status:          dao.Status,
			statusDetails:   dao.StatusDetails,
			firstRequest:    order.NewRequestFromDTO(svc.market, dao.FirstRequest),
			secondRequest:   order.NewRequestFromDTO(svc.market, dao.SecondRequest),
			reversalRequest: order.NewRequestFromDTO(svc.market, dao.ReversalRequest),
		}

		if orderPair.Direction() == "" {
			if dao.FirstRequest.Side == order.Buy {
				orderPair.direction = Upward
			} else {
				orderPair.direction = Downward
			}

		}

		// Load the first order if it's been placed
		if dao.FirstOrder.ID != "" {
			if dao.FirstOrder.Status != order.Canceled {
				order, err := svc.trader.OrderSvc().Order(svc.market, dao.FirstOrder.ID)
				if err != nil {
					return nil, fmt.Errorf("could not load first order: %w", err)
				}
				orderPair.firstOrder = order
			} else {
				orderPair.firstOrder = svc.trader.OrderSvc().OrderFromDTO(dao.FirstOrder)
			}
		}

		// Load the second order if it's been placed
		if dao.SecondOrder.ID != "" {
			if dao.SecondOrder.Status != order.Canceled {
				order, err := svc.trader.OrderSvc().Order(svc.market, dao.SecondOrder.ID)
				if err != nil {
					return nil, fmt.Errorf("could not load second order: %w", err)
				}
				orderPair.secondOrder = order
			} else {
				orderPair.secondOrder = svc.trader.OrderSvc().OrderFromDTO(dao.SecondOrder)
			}

		}

		// Load the reversal order if it's been placed
		if dao.ReversalOrder.ID != "" {
			if dao.ReversalOrder.Status != order.Canceled {
				order, err := svc.trader.OrderSvc().Order(svc.market, dao.ReversalOrder.ID)
				if err != nil {
					return nil, fmt.Errorf("could not load reversal order: %w", err)
				}
				orderPair.reversalOrder = order
			} else {
				orderPair.reversalOrder = svc.trader.OrderSvc().OrderFromDTO(dao.ReversalOrder)
			}
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
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("pair %s was not found in the database", id)
		} else {
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
			log.WithError(err).Warnf("could not load open order %s; marking as broken", dao.Uuid)
			dao.StatusDetails = err.Error()
			dao.Status = Broken
			svc.Save(dao)
			continue
		}
		// Add to return
		pairs = append(pairs, pair)
	}
	return
}

func (svc *Service) GetCollidingOpenPair(newPair *OrderPair) (pair *OrderPair, err error) {
	// Get the pairs from cache
	pairs, err := svc.LoadOpenPairs()
	if err != nil {
		return
	}

	// Search the pairs for a colliding one
	for _, p := range pairs {
		// Only check the ones going in the same direction
		if newPair.Direction() == p.Direction() {

			buyPrice := newPair.BuyRequest().Price()
			sellPrice := newPair.SellRequest().Price()

			lower := p.BuyRequest().Price()
			upper := p.SellRequest().Price()

			// If the buy or the sell of the new order is between the buy and sell of the order, it's colliding
			if (buyPrice.GreaterThanOrEqual(lower) && buyPrice.LessThanOrEqual(upper)) ||
				(sellPrice.GreaterThanOrEqual(lower) && sellPrice.LessThanOrEqual(upper)) {

				// Return colliding pair
				pair = p

				// Save pair
				pair.Save()

				return
			}
		}
	}
	return
}

func (svc *Service) MakeRoom(startingPrice decimal.Decimal, direction Direction) error {
	// Get open pairs for direction
	pairs := []*OrderPair{}
	openPairs, err := svc.LoadOpenPairs()
	if err != nil {
		return fmt.Errorf("could not load open pairs to make room: %w", err)
	}
	for _, pair := range openPairs {
		if pair.Direction() == direction && pair.Status() == Open {
			pairs = append(pairs, pair)
		}
	}

	// Get the max open pairs
	log.Debug("getting max open pairs")
	max, err := svc.getMaxOpenPairs(startingPrice, direction)
	if err != nil {
		return fmt.Errorf("could not get max open pairs: %w", err)
	}

	// Bail if there's already enough room
	if len(pairs) < max {
		return nil
	}

	// Make room for new orders
	log.Debug("making room for new orders")
	switch viper.GetString("makeRoomStrategy") {
	case "newest":
		for len(pairs)+1 >= max {
			// Find the newest pair
			newest := pairs[0]
			var idx int
			for i, pair := range pairs {
				if pair.CreatedAt().After(newest.CreatedAt()) {
					newest = pair
					idx = i
				}
			}

			// Cancel newest pair
			if newest.Status() == Open {
				log.Infof("%s: canceling pair to make room", newest.UUID().String())
				err = newest.Cancel()
				if err != nil {
					return fmt.Errorf("could not cancel newest pair to make room: %w", err)
				}
			}
			// Wait for the pair to make room
			<-newest.Done()

			// Remove the pair from the slice
			pairs = append(pairs[:idx], pairs[idx+1:]...)

			// Wait for consistency
			<-time.NewTicker(time.Second).C

			// Reset max
			max, err = svc.getMaxOpenPairs(startingPrice, direction)
			if err != nil {
				return fmt.Errorf("could not get max open pairs: %w", err)
			}
		}
	case "oldest":
		for len(pairs)+1 >= max {
			// Find the oldest pair
			oldest := pairs[0]
			var idx int
			for i, pair := range pairs {
				if pair.CreatedAt().Before(oldest.CreatedAt()) {
					oldest = pair
					idx = i
				}
			}

			// Cancel oldest pair
			if oldest.Status() == Open {
				log.Infof("%s: canceling pair to make room", oldest.UUID().String())
				err = oldest.Cancel()
				if err != nil {
					return fmt.Errorf("could not cancel oldest pair to make room: %w", err)
				}
			}
			// Wait for the pair to make room
			<-oldest.Done()

			// Remove the pair from the slice
			pairs = append(pairs[:idx], pairs[idx+1:]...)

			// Wait for consistency
			<-time.NewTicker(time.Second).C

			// Reset max
			max, err = svc.getMaxOpenPairs(startingPrice, direction)
			if err != nil {
				return fmt.Errorf("could not get max open pairs: %w", err)
			}
		}
	default:
		return fmt.Errorf("unknown strategy '%s'", viper.GetString("makeRoomStrategy"))
	}

	return nil
}

func (svc *Service) initializeDB() error {
	_, err := svc.db.Exec("CREATE TABLE IF NOT EXISTS orderpairs (uuid char(36) primary key, data JSONB);")
	if err != nil {
		return err
	}
	return nil
}

func (svc *Service) getMaxOpenPairs(price decimal.Decimal, direction Direction) (max int, err error) {
	// Get the max order size from max number of open orders plus 1 to add a buffer
	maxOpenPairs := viper.GetInt("maxOpenPairs")

	// Get the number of pairs for direction
	pairs := 0
	openPairs, err := svc.LoadOpenPairs()
	if err != nil {
		return 0, fmt.Errorf("could not load open pairs to make room: %w", err)
	}
	for _, pair := range openPairs {
		if pair.Direction() == direction && pair.Status() == Open {
			pairs++
		}
	}

	// return max if there are more pairs than should be
	if maxOpenPairs < pairs {
		return maxOpenPairs, nil
	}

	// Figure out the max based on how much balance is available
	var size decimal.Decimal
	it := 0
	quoteWallet := svc.market.QuoteCurrency().Wallet()
	baseWallet := svc.market.BaseCurrency().Wallet()

	if direction == Upward {
		for svc.market.MinQuantity().GreaterThan(size) {
			// Make sure we have enough money for the max order size
			ratio := decimal.NewFromInt(int64(maxOpenPairs - pairs - it))
			if ratio.LessThanOrEqual(decimal.Zero) {
				break
			}
			size = quoteWallet.Available().Div(price).Div(ratio)
			log.Debugf("min funds %s size %s", svc.market.MinQuantity(), size)
			it++
		}
	} else {
		for svc.market.MinQuantity().GreaterThan(size) {
			// Make sure we have enough money for the max order size
			ratio := decimal.NewFromInt(int64(maxOpenPairs - pairs - it))
			if ratio.LessThanOrEqual(decimal.Zero) {
				break
			}
			size = baseWallet.Available().Div(ratio)
			log.Debugf("min qty %s size %s", svc.market.MinQuantity(), size)
			it++
		}
	}

	max = maxOpenPairs - it
	log.Debugf("max open pairs %d", max)
	return
}
