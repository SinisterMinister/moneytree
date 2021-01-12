package pair

import (
	"fmt"

	"github.com/go-playground/log"
	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/fees"
	"github.com/sinisterminister/currencytrader/types/order"
	"github.com/spf13/viper"
)

func BuildSpreadBasedPair(svc *Service, dir Direction) (pair *OrderPair, err error) {
	// Get the currencies
	quoteCurrency := svc.market.QuoteCurrency()
	baseCurrency := svc.market.BaseCurrency()

	// Get the ticker for the current prices
	ticker, err := svc.market.Ticker()
	if err != nil {
		return nil, err
	}

	// Determine sell size so that both currencies gain
	orderFee, err := getFees(svc.trader)
	if err != nil {
		return nil, fmt.Errorf("could not load fees: %w", err)
	}

	// Set the profit target
	targetReturn := decimal.NewFromFloat(viper.GetFloat64("targetReturn"))

	// Set the prices and fees
	var buyPrice, sellPrice, fee1, fee2 decimal.Decimal
	if dir == Upward {
		fee2 = orderFee.MakerRate()

		// Force maker orders
		if viper.GetBool("forceMakerOrders") {
			fee1 = orderFee.MakerRate()
			buyPrice = ticker.Bid()
		} else {
			fee1 = orderFee.TakerRate()
			buyPrice = ticker.Ask()
		}

		// Determine the fee rate
		rate := fee1.Add(fee2)

		// Calculate spread
		spread := targetReturn.Add(rate)

		// Determine sell price from spread
		sellPrice = buyPrice.Add(buyPrice.Mul(spread)).Round(int32(quoteCurrency.Precision()))
	} else {
		fee1 = orderFee.MakerRate()

		// Force maker orders
		if viper.GetBool("forceMakerOrders") {
			fee2 = orderFee.MakerRate()
			sellPrice = ticker.Bid()
		} else {
			fee2 = orderFee.TakerRate()
			sellPrice = ticker.Ask()
		}

		// Determine the fee rate
		rate := fee1.Add(fee2)

		// Calculate spread
		spread := targetReturn.Add(rate)

		// Determine buy price from spread
		buyPrice = sellPrice.Sub(sellPrice.Mul(spread)).Round(int32(quoteCurrency.Precision()))
	}

	// Set the sizes
	size, err := size(svc, ticker, dir)
	if err != nil {
		return nil, err
	}

	// Set buy size to base size
	buySize := size.Round(int32(baseCurrency.Precision()))

	// Setup the numbers we need
	two := decimal.NewFromFloat(2)

	// 2a - 2ab - abt - 2abg
	n := two.Mul(buySize).Sub(two.Mul(buySize).Mul(buyPrice)).Sub(buySize.Mul(buyPrice).Mul(targetReturn)).Sub(two.Mul(buySize).Mul(buyPrice).Mul(fee2))
	// t + 2f + 2 - 2d
	d := targetReturn.Add(two.Mul(fee1)).Add(two).Sub(two.Mul(sellPrice))

	// Set sell size
	sellSize := n.Div(d)

	// Round to correct precision
	sellSize = sellSize.Round(int32(baseCurrency.Precision()))

	// Create order pair
	var op *OrderPair
	var sellReq, buyReq types.OrderRequest
	if dir == Upward {
		sellReq = order.NewRequest(svc.market, order.Limit, order.Sell, sellSize, sellPrice, false)
		buyReq = order.NewRequest(svc.market, order.Limit, order.Buy, buySize, buyPrice, viper.GetBool("forceMakerOrders"))
		op, err = svc.New(buyReq, sellReq)
	} else {
		sellReq = order.NewRequest(svc.market, order.Limit, order.Sell, sellSize, sellPrice, viper.GetBool("forceMakerOrders"))
		buyReq = order.NewRequest(svc.market, order.Limit, order.Buy, buySize, buyPrice, false)
		op, err = svc.New(sellReq, buyReq)
	}
	log.WithFields(
		log.F("sellSize", sellSize.String()),
		log.F("sellPrice", sellPrice.String()),
		log.F("buySize", buySize.String()),
		log.F("buyPrice", buyPrice.String()),
	).Infof("%s trending order data", dir)

	if err != nil {
		return nil, fmt.Errorf("could not create order pair: %w", err)
	}
	return op, nil
}

func getFees(trader types.Trader) (f types.Fees, err error) {
	// Allow disabling of fees to let the system work the raw algorithm
	if viper.GetBool("disableFees") == true {
		f = fees.ZeroFee()
	} else {
		// Get the fees
		var err error
		f, err = trader.AccountSvc().Fees()
		if err != nil {
			log.WithError(err).Error("failed to get fees")
			return fees.ZeroFee(), err
		}
	}
	return
}

func size(svc *Service, ticker types.Ticker, dir Direction) (decimal.Decimal, error) {
	// Get the max order size from max number of open orders plus 1 to add a buffer
	maxOpenPairs := decimal.NewFromFloat(viper.GetFloat64("maxOpenPairs")).Add(decimal.NewFromFloat(1))

	// Load the open pairs
	openPairs, err := svc.LoadOpenPairs()
	if err != nil {
		return decimal.Zero, err
	}

	// Get the number of pairs of the same direction
	count := 0
	for _, p := range openPairs {
		if p.Direction() == dir {
			count++
		}
	}

	ratio := maxOpenPairs.Sub(decimal.NewFromInt(int64(count)))

	var size decimal.Decimal

	if dir == Upward {
		quoteWallet := svc.market.QuoteCurrency().Wallet()
		size = quoteWallet.Available().Div(ticker.Bid()).Div(ratio)
	} else {
		baseWallet := svc.market.BaseCurrency().Wallet()
		size = baseWallet.Available().Div(ratio)
	}

	// Set the base size
	return size, nil
}