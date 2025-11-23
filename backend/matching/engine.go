package matching

import (
	"container/heap"
	"sync"
	"time"
	"trading/models"
)

// OrderBook maintains buy and sell orders for a single stock
type OrderBook struct {
	Stock      string
	BuyOrders  *MaxHeap // Higher prices first
	SellOrders *MinHeap // Lower prices first
	mu         sync.RWMutex
}

// MatchingEngine coordinates order matching across all stocks
type MatchingEngine struct {
	orderBooks map[string]*OrderBook
	mu         sync.RWMutex
}

// NewMatchingEngine creates a new matching engine instance
func NewMatchingEngine() *MatchingEngine {
	return &MatchingEngine{
		orderBooks: make(map[string]*OrderBook),
	}
}

// getOrCreateBook returns existing order book or creates new one
func (me *MatchingEngine) getOrCreateBook(stock string) *OrderBook {
	me.mu.Lock()
	defer me.mu.Unlock()

	book, exists := me.orderBooks[stock]
	if !exists {
		book = &OrderBook{
			Stock:      stock,
			BuyOrders:  &MaxHeap{},
			SellOrders: &MinHeap{},
		}
		heap.Init(book.BuyOrders)
		heap.Init(book.SellOrders)
		me.orderBooks[stock] = book
	}
	return book
}

// PlaceOrder attempts to match order and returns executed trades
func (me *MatchingEngine) PlaceOrder(order *models.Order) []*models.Trade {
	book := me.getOrCreateBook(order.Stock)
	book.mu.Lock()
	defer book.mu.Unlock()

	var trades []*models.Trade

	if order.Type == "BUY" {
		trades = me.matchBuyOrder(order, book)
	} else {
		trades = me.matchSellOrder(order, book)
	}

	return trades
}

// matchBuyOrder matches buy order against sell orders
func (me *MatchingEngine) matchBuyOrder(order *models.Order, book *OrderBook) []*models.Trade {
	var trades []*models.Trade

	// Match with existing sell orders
	for order.Quantity > 0 && book.SellOrders.Len() > 0 {
		bestSell := (*book.SellOrders)[0]

		// Can only match if sell price <= buy price
		if bestSell.Price <= order.Price {
			tradeQty := min(order.Quantity, bestSell.Quantity)

			trade := &models.Trade{
				Stock:       order.Stock,
				BuyerID:     order.UserID,
				SellerID:    bestSell.UserID,
				BuyOrderID:  order.ID,
				SellOrderID: bestSell.ID,
				Price:       bestSell.Price, // Price from resting order
				Quantity:    tradeQty,
				ExecutedAt:  time.Now(),
			}
			trades = append(trades, trade)

			order.Quantity -= tradeQty
			bestSell.Quantity -= tradeQty

			if bestSell.Quantity == 0 {
				heap.Pop(book.SellOrders)
			}
		} else {
			break // No more matches possible
		}
	}

	// Add remaining quantity to order book
	if order.Quantity > 0 {
		heap.Push(book.BuyOrders, order)
	}

	return trades
}

// matchSellOrder matches sell order against buy orders
func (me *MatchingEngine) matchSellOrder(order *models.Order, book *OrderBook) []*models.Trade {
	var trades []*models.Trade

	// Match with existing buy orders
	for order.Quantity > 0 && book.BuyOrders.Len() > 0 {
		bestBuy := (*book.BuyOrders)[0]

		// Can only match if buy price >= sell price
		if bestBuy.Price >= order.Price {
			tradeQty := min(order.Quantity, bestBuy.Quantity)

			trade := &models.Trade{
				Stock:       order.Stock,
				BuyerID:     bestBuy.UserID,
				SellerID:    order.UserID,
				BuyOrderID:  bestBuy.ID,
				SellOrderID: order.ID,
				Price:       bestBuy.Price, // Price from resting order
				Quantity:    tradeQty,
				ExecutedAt:  time.Now(),
			}
			trades = append(trades, trade)

			order.Quantity -= tradeQty
			bestBuy.Quantity -= tradeQty

			if bestBuy.Quantity == 0 {
				heap.Pop(book.BuyOrders)
			}
		} else {
			break // No more matches possible
		}
	}

	// Add remaining quantity to order book
	if order.Quantity > 0 {
		heap.Push(book.SellOrders, order)
	}

	return trades
}

// GetOrderBook returns current order book state
func (me *MatchingEngine) GetOrderBook(stock string) models.OrderBookResponse {
	book := me.getOrCreateBook(stock)
	book.mu.RLock()
	defer book.mu.RUnlock()

	resp := models.OrderBookResponse{
		Stock: stock,
		Bids:  []models.PriceLevel{},
		Asks:  []models.PriceLevel{},
	}

	// Aggregate buy orders by price
	bidMap := make(map[float64]int)
	for _, order := range *book.BuyOrders {
		bidMap[order.Price] += order.Quantity
	}
	for price, qty := range bidMap {
		resp.Bids = append(resp.Bids, models.PriceLevel{
			Price:    price,
			Quantity: qty,
		})
	}

	// Aggregate sell orders by price
	askMap := make(map[float64]int)
	for _, order := range *book.SellOrders {
		askMap[order.Price] += order.Quantity
	}
	for price, qty := range askMap {
		resp.Asks = append(resp.Asks, models.PriceLevel{
			Price:    price,
			Quantity: qty,
		})
	}

	return resp
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
