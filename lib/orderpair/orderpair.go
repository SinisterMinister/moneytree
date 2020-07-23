package orderpair

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"

	"github.com/go-playground/log/v7"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/fees"
	"github.com/sinisterminister/currencytrader/types/order"
	"github.com/spf13/viper"
)

type OrderPair struct {
	trader types.Trader
	market types.Market
	db     *sql.DB

	mutex         sync.RWMutex
	uuid          uuid.UUID
	firstRequest  types.OrderRequest
	firstOrder    types.Order
	secondRequest types.OrderRequest
	secondOrder   types.Order
	running       bool
	failed        bool
	startHold     chan bool
	done          chan bool
	stop          <-chan bool
}

func New(db *sql.DB, trader types.Trader, market types.Market, first types.OrderRequest, second types.OrderRequest) (orderPair *OrderPair, err error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("could not create time based UUID: %w", err)
	}

	orderPair = &OrderPair{
		db:            db,
		uuid:          id,
		trader:        trader,
		market:        market,
		done:          make(chan bool),
		startHold:     make(chan bool),
		firstRequest:  first,
		secondRequest: second,
	}

	// Validate DTOs
	err = orderPair.validate()
	if err != nil {
		return nil, err
	}

	return orderPair, nil
}

func NewFromDAO(db *sql.DB, trader types.Trader, market types.Market, dao OrderPairDAO) (orderPair *OrderPair, err error) {
	id, err := uuid.Parse(dao.Uuid)
	if err != nil {
		return nil, fmt.Errorf("could not parse order pair ID: %w", err)
	}

	orderPair = &OrderPair{
		db:            db,
		uuid:          id,
		trader:        trader,
		market:        market,
		done:          make(chan bool),
		startHold:     make(chan bool),
		firstRequest:  order.NewRequestFromDTO(market, dao.FirstRequest),
		secondRequest: order.NewRequestFromDTO(market, dao.SecondRequest),
	}

	if dao.FirstOrderID != "" {
		order, err := trader.OrderSvc().Order(market, dao.FirstOrderID)
		if err != nil {
			if strings.Contains(err.Error(), "NotFound") {
				log.Warnf("could not load first order %s, closing as failed", dao.FirstOrderID)
				orderPair.failed = true
				close(orderPair.done)
				orderPair.Save()
				return orderPair, nil
			}
			return nil, err
		}
		orderPair.firstOrder = order
	}

	if dao.SecondOrderID != "" {
		order, err := trader.OrderSvc().Order(market, dao.SecondOrderID)
		if err != nil {
			if strings.Contains(err.Error(), "NotFound") {
				log.Warnf("could not load second order %s, ignoring", dao.SecondOrderID)
				orderPair.Save()
				return orderPair, nil
			}
			return nil, err
		}
		orderPair.secondOrder = order
	}
	return
}

func SetupDB(db *sql.DB) error {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS orderpairs (uuid char(36) primary key, data JSONB);")
	if err != nil {
		return err
	}
	return nil
}

func Load(db *sql.DB, trader types.Trader, market types.Market, id string) (pair *OrderPair, err error) {
	dao := OrderPairDAO{}
	err = db.QueryRow("SELECT data FROM orderpairs WHERE uuid = $1;", id).Scan(&dao)
	if err != nil {
		return nil, fmt.Errorf("could not load order pair from database: %w", err)
	}

	pair, err = NewFromDAO(db, trader, market, dao)
	return
}

func LoadOpenPairs(db *sql.DB, trader types.Trader, market types.Market) (pairs []*OrderPair, err error) {
	pairs = []*OrderPair{}
	rows, err := db.Query("SELECT data FROM orderpairs WHERE (data->>'done')::boolean = FALSE;")
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
		pair, err := NewFromDAO(db, trader, market, dao)
		if err != nil {
			return nil, fmt.Errorf("could not load open order: %w", err)
		}
		pairs = append(pairs, pair)
	}
	return
}

func (o *OrderPair) Save() (err error) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	dao := o.ToDAO()
	log.WithField("dao", dao).Debug("saving order pair")
	_, err = o.db.Exec("INSERT INTO orderpairs (uuid, data) VALUES ($1, $2) ON CONFLICT (uuid) DO UPDATE SET data = $2;", o.uuid, dao)
	if err != nil {
		err = fmt.Errorf("could not insert into database: %w", err)
	}
	return
}

func (o *OrderPair) Execute(stop <-chan bool) <-chan bool {
	o.mutex.Lock()
	// Only launch routine if not running already
	if !o.running {
		go o.executeWorkflow()
	}
	o.running = true
	o.stop = stop
	o.mutex.Unlock()

	// Wait for the orders to start
	<-o.startHold

	return o.done
}

func (o *OrderPair) IsDone() bool {
	select {
	case <-o.done:
		return true
	default:
		return false
	}
}

func (o *OrderPair) FirstOrder() types.Order {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	return o.firstOrder
}

func (o *OrderPair) SecondOrder() types.Order {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	return o.secondOrder
}

func (o *OrderPair) FirstRequest() types.OrderRequest {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	return o.firstRequest
}

func (o *OrderPair) SecondRequest() types.OrderRequest {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	return o.secondRequest
}

func (o *OrderPair) BuyRequest() types.OrderRequest {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	if o.firstRequest.Side() == order.Buy {
		return o.firstRequest
	}
	return o.secondRequest
}

func (o *OrderPair) SellRequest() types.OrderRequest {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	if o.firstRequest.Side() != order.Buy {
		return o.firstRequest
	}
	return o.secondRequest
}

func (o *OrderPair) Cancel() error {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	return o.trader.OrderSvc().CancelOrder(o.firstOrder)
}

func (o *OrderPair) ToDAO() *OrderPairDAO {
	var fid, sid string
	if o.firstOrder != nil {
		fid = o.firstOrder.ID()
	}

	if o.secondOrder != nil {
		sid = o.secondOrder.ID()
	}

	var done bool
	select {
	case <-o.done:
		done = true
	default:
		done = false
	}

	return &OrderPairDAO{
		Uuid:          o.uuid.String(),
		FirstRequest:  o.firstRequest.ToDTO(),
		SecondRequest: o.secondRequest.ToDTO(),
		FirstOrderID:  fid,
		SecondOrderID: sid,
		Done:          done,
	}
}

func (o *OrderPair) executeWorkflow() {
	var err error

	// Place the first order
	if o.firstOrder == nil {
		err = o.placeFirstOrder()
		if err != nil {
			log.WithError(err).Error("could not place first order")
			close(o.done)
			return
		}

		// Save the order
		err = o.Save()
		if err != nil {
			log.WithError(err).Error("could not save order")
		}
	}

	// release start hold
	close(o.startHold)

	// Wait for order to complete. If it fails, keep going in case partial fill
	err = o.waitForFirstOrder()
	if err != nil {
		log.WithError(err).Warn("first order failed")
	}

	if o.secondOrder == nil {
		// Place second order
		err = o.placeSecondOrder()
		if err != nil {
			log.WithError(err).Warn("second order failed")
			select {
			case <-o.done:
			default:
				close(o.done)
			}
			return
		}
		log.Info("second order placed")

		// Save the order
		err = o.Save()
		if err != nil {
			log.WithError(err).Error("could not save order")
		}
	}

	// Wait for second order to complete
	log.Info("waiting on second order")
	<-o.secondOrder.Done()
	log.Info("second order done processing")

	// Signal completion
	close(o.done)

	// Save the order
	err = o.Save()
	if err != nil {
		log.WithError(err).Error("could not save order")
	}
}

func (o *OrderPair) placeFirstOrder() (err error) {
	r0 := o.firstRequest.ToDTO()
	log.WithFields(log.F("side", r0.Side), log.F("price", r0.Price), log.F("quantity", r0.Quantity)).Info("placing first order")

	// Place first order
	o.mutex.Lock()
	o.firstOrder, err = o.market.AttemptOrder(o.firstRequest)
	o.mutex.Unlock()
	return
}

func (o *OrderPair) placeSecondOrder() (err error) {
	// Bail if fill amount is zero
	if o.firstOrder.Filled().Equal(decimal.Zero) {
		return fmt.Errorf("first order was not filled, skipping second")
	}

	// Place second order
	o.mutex.Lock()
	o.recalculateSecondOrderSizeFromFilled()
	r1 := o.secondRequest.ToDTO()
	log.WithFields(log.F("side", r1.Side), log.F("price", r1.Price), log.F("quantity", r1.Quantity)).Info("placing second order")
	o.secondOrder, err = o.market.AttemptOrder(o.secondRequest)
	o.mutex.Unlock()
	return
}

func (o *OrderPair) waitForFirstOrder() (err error) {
	stop := make(chan bool)
	tickerStream := o.market.TickerStream(stop)
	for {
		brk := false
		select {
		case <-o.stop:
			close(stop)
			return fmt.Errorf("stop channel closed")
		case tick := <-tickerStream:
			// Bail if the order missed
			spread := o.firstRequest.Price().Sub(tick.Ask()).Div(o.firstRequest.Price()).Abs()
			if spread.GreaterThan(decimal.NewFromFloat(viper.GetFloat64("orderpair.missDistance"))) && o.firstOrder.Filled().Equals(decimal.Zero) {
				close(stop)
				return fmt.Errorf("first order missed")
			}
		case <-o.firstOrder.Done():
			log.Info("first order done processing")
			// Order is complete, time to move on
			brk = true
		}

		// I want to break free...
		if brk {
			break
		}
	}
	// Close ticker stream
	close(stop)
	return
}

func (o *OrderPair) recalculateSecondOrderSizeFromFilled() {
	// Determine the ratio from the first to the second
	ratio := o.secondRequest.Quantity().Div(o.firstRequest.Quantity())

	// Calculate the new size
	size := o.firstOrder.Filled().Mul(ratio).Round(int32(o.market.QuoteCurrency().Precision()))

	// Build updated DTO
	dto := o.secondRequest.ToDTO()
	dto.Quantity = size

	// Set the new request
	o.secondRequest = order.NewRequestFromDTO(o.market, dto)
}

func (o *OrderPair) validate() error {
	// Make sure it's a BUY/SELL pair
	if o.firstRequest.Side() == o.secondRequest.Side() {
		return &SameSideError{o}
	}

	// Figure out the net result of the trades against our currency balance
	baseRes := o.BuyRequest().Quantity().Sub(o.SellRequest().Quantity())
	quoteRes := o.SellRequest().Price().Mul(o.SellRequest().Quantity()).Sub(o.BuyRequest().Price().Mul(o.BuyRequest().Quantity()))

	// Make sure we're not losing currency
	if baseRes.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("not making more of base currency, %w, %s", &LosingPropositionError{o}, baseRes.String())
	}
	if quoteRes.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("not making more of quote currency, %w, %s", &LosingPropositionError{o}, quoteRes.String())
	}

	// Get the fee rates
	rates, err := o.trader.AccountSvc().Fees()
	if err != nil {
		return err
	}
	if viper.GetBool("disableFees") == true {
		rates = fees.ZeroFee()
	}

	// Determin the fees
	baseFee := o.BuyRequest().Quantity().Mul(rates.TakerRate())
	quoteFee := o.SellRequest().Price().Mul(o.SellRequest().Quantity()).Mul(rates.TakerRate())

	// Make sure we're not losing currency
	if baseRes.LessThanOrEqual(baseFee) {
		return fmt.Errorf("not making more of base currency after fees, %w, %s, %s", &LosingPropositionError{o}, baseRes.String(), baseFee.String())
	}
	if quoteRes.LessThanOrEqual(quoteFee) {
		return fmt.Errorf("not making more of quote currency after fees, %w, %s, %s", &LosingPropositionError{o}, quoteRes.String(), quoteFee.String())
	}

	return nil
}
