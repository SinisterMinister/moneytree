package orderpair

import (
	"database/sql"
	"fmt"
	"strings"
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
		return nil, fmt.Errorf("could not load order pair from database: %w", err)
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

func (svc *Service) LoadMostRecentOpenPair() (pair *OrderPair, err error) {
	dao := OrderPairDAO{}
	err = svc.db.QueryRow("SELECT data FROM orderpairs WHERE data->>'status' = 'OPEN' AND (data->>'done')::boolean = FALSE ORDER BY data->>'createdAt' DESC LIMIT 1").Scan(&dao)
	if err != nil {
		return nil, fmt.Errorf("could not load order pair from database: %w", err)
	}

	pair, err = svc.NewFromDAO(dao)
	return
}

func (svc *Service) LoadOpenPairs() (pairs []*OrderPair, err error) {
	pairs = []*OrderPair{}
	rows, err := svc.db.Query("SELECT data FROM orderpairs WHERE data->>'status' = 'OPEN'")
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
		pair, err := svc.NewFromDAO(dao)
		if err != nil {
			return nil, fmt.Errorf("could not load open order: %w", err)
		}
		pairs = append(pairs, pair)
	}
	return
}

func (svc *Service) LowestOpenBuyFirstPrice() (price decimal.Decimal, err error) {
	dao := OrderPairDAO{}
	err = svc.db.QueryRow("SELECT data FROM orderpairs WHERE data->>'status' = 'OPEN' AND data->'firstRequest'->>'side' = 'BUY' AND data->'firstOrder'->>'status' = 'FILLED' ORDER BY (data->'firstRequest'->>'price')::decimal DESC LIMIT 1").Scan(&dao)
	if err != nil {
		return decimal.Zero, fmt.Errorf("could not load order pair from database: %w", err)
	}
	price = dao.FirstRequest.Price

	return
}

func (svc *Service) HighestOpenSellFirstPrice() (price decimal.Decimal, err error) {
	dao := OrderPairDAO{}
	err = svc.db.QueryRow("SELECT data FROM orderpairs WHERE data->>'status' = 'OPEN' AND (data->'firstRequest'->>'side') = 'SELL' AND (data->'firstOrder'->>'status') = 'FILLED' ORDER BY (data->'firstRequest'->>'price')::decimal LIMIT 1").Scan(&dao)
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

	if dao.FirstOrderID != "" {
		order, err := svc.trader.OrderSvc().Order(svc.market, dao.FirstOrderID)
		if err != nil {
			if strings.Contains(err.Error(), "NotFound") {
				log.Warnf("could not load first order %s, closing as failed", dao.FirstOrderID)
				orderPair.failed = true
				select {
				case <-orderPair.done:
				default:
					close(orderPair.done)
				}
				svc.Save(orderPair.ToDAO())
				return orderPair, nil
			}
			return nil, err
		}
		orderPair.firstOrder = order
	}

	if dao.SecondOrderID != "" {
		order, err := svc.trader.OrderSvc().Order(svc.market, dao.SecondOrderID)
		if err != nil {
			if strings.Contains(err.Error(), "NotFound") {
				log.Warnf("could not load second order %s, ignoring", dao.SecondOrderID)
				svc.Save(orderPair.ToDAO())
				return orderPair, nil
			}
			return nil, err
		}
		orderPair.secondOrder = order
	}
	return
}
