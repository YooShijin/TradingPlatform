package models

import "time"

// Order represents a buy or sell order
type Order struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Stock     string    `json:"stock"`
	Type      string    `json:"type"` // "BUY" or "SELL"
	Price     float64   `json:"price"`
	Quantity  int       `json:"quantity"`
	Timestamp time.Time `json:"timestamp"`
}

// Trade represents an executed trade between buyer and seller
type Trade struct {
	ID          int       `json:"id"`
	Stock       string    `json:"stock"`
	BuyerID     int       `json:"buyer_id"`
	SellerID    int       `json:"seller_id"`
	BuyOrderID  int       `json:"buy_order_id"`
	Type        string    `json:"type"` // "BUY" or "SELL"
	SellOrderID int       `json:"sell_order_id"`
	Price       float64   `json:"price"`
	Quantity    int       `json:"quantity"`
	ExecutedAt  time.Time `json:"executed_at"`
}

// OrderBookResponse represents aggregated order book data
type OrderBookResponse struct {
	Stock string       `json:"stock"`
	Bids  []PriceLevel `json:"bids"` // Buy orders
	Asks  []PriceLevel `json:"asks"` // Sell orders
}

// PriceLevel represents aggregated quantity at a price point
type PriceLevel struct {
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}

// PortfolioStats represents user's portfolio information
type PortfolioStats struct {
	UserID     int       `json:"user_id"`
	Balance    float64   `json:"balance"`
	Holdings   []Holding `json:"holdings"`
	TotalValue float64   `json:"total_value"`
	TotalPL    float64   `json:"total_profit_loss"`
}

// Holding represents a stock position in portfolio
type Holding struct {
	Stock        string  `json:"stock"`
	Quantity     int     `json:"quantity"`
	AvgBuyPrice  float64 `json:"avg_buy_price"`
	CurrentPrice float64 `json:"current_price"`
	ProfitLoss   float64 `json:"profit_loss"`
}

// PlaceOrderRequest represents incoming order request
type PlaceOrderRequest struct {
	UserID   int     `json:"user_id"`
	Stock    string  `json:"stock"`
	Type     string  `json:"type"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}

// PlaceOrderResponse represents order placement result
type PlaceOrderResponse struct {
	Status    string   `json:"status"`
	OrderID   int      `json:"order_id"`
	Trades    []*Trade `json:"trades"`
	Remaining int      `json:"remaining"`
}
