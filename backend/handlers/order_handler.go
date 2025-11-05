package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"trading/matching"
	"trading/models"

	"github.com/gorilla/mux"
)

// OrderHandler handles order-related HTTP requests
type OrderHandler struct {
	db             *sql.DB
	matchingEngine *matching.MatchingEngine
	orderBookCache struct {
		lastUpdate map[string]time.Time
		mu         sync.RWMutex
	}
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(db *sql.DB, engine *matching.MatchingEngine) *OrderHandler {
	h := &OrderHandler{
		db:             db,
		matchingEngine: engine,
	}
	h.orderBookCache.lastUpdate = make(map[string]time.Time)
	return h
}

// PlaceOrder handles POST /api/order/place
func (h *OrderHandler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	var req models.PlaceOrderRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Validate order
	if req.Type != "BUY" && req.Type != "SELL" {
		http.Error(w, "Invalid order type", http.StatusBadRequest)
		return
	}
	if req.Price <= 0 || req.Quantity <= 0 {
		http.Error(w, "Invalid price or quantity", http.StatusBadRequest)
		return
	}

	// Start database transaction
	tx, err := h.db.Begin()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Validate user has sufficient balance/holdings
	if err := h.validateOrder(tx, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Insert order into database
	orderID, err := h.createOrder(tx, &req)
	if err != nil {
		http.Error(w, "Failed to create order", http.StatusInternalServerError)
		return
	}

	// Create order object for matching engine
	order := &models.Order{
		ID:        orderID,
		UserID:    req.UserID,
		Stock:     req.Stock,
		Type:      req.Type,
		Price:     req.Price,
		Quantity:  req.Quantity,
		Timestamp: time.Now(),
	}

	// Match order using in-memory engine
	trades := h.matchingEngine.PlaceOrder(order)

	// Process each trade in database
	if err := h.processTrades(tx, trades); err != nil {
		log.Printf("Failed to process trades: %v", err)
		http.Error(w, "Failed to process trades", http.StatusInternalServerError)
		return
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	// Invalidate order book cache
	h.invalidateCache(req.Stock)

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.PlaceOrderResponse{
		Status:    "success",
		OrderID:   orderID,
		Trades:    trades,
		Remaining: order.Quantity,
	})
}

// GetOrderBook handles GET /api/orderbook/{stock}
func (h *OrderHandler) GetOrderBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stock := vars["stock"]

	// Check cache
	h.orderBookCache.mu.RLock()
	lastUpdate, exists := h.orderBookCache.lastUpdate[stock]
	h.orderBookCache.mu.RUnlock()

	// If cache is fresh (< 100ms), use cached data
	if exists && time.Since(lastUpdate) < 100*time.Millisecond {
		orderBook := h.matchingEngine.GetOrderBook(stock)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		json.NewEncoder(w).Encode(orderBook)
		return
	}

	// Get from matching engine
	orderBook := h.matchingEngine.GetOrderBook(stock)

	// Update cache
	h.orderBookCache.mu.Lock()
	h.orderBookCache.lastUpdate[stock] = time.Now()
	h.orderBookCache.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	json.NewEncoder(w).Encode(orderBook)
}

// CancelOrder handles DELETE /api/order/{id}
func (h *OrderHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var orderID int
	fmt.Sscanf(vars["id"], "%d", &orderID)

	result, err := h.db.Exec(`
        UPDATE orders 
        SET status = 'CANCELLED'
        WHERE id = $1 AND status = 'PENDING'
    `, orderID)

	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Order not found or already filled", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "cancelled"})
}

// Helper methods

func (h *OrderHandler) validateOrder(tx *sql.Tx, req *models.PlaceOrderRequest) error {
	if req.Type == "BUY" {
		var balance float64
		err := tx.QueryRow("SELECT balance FROM users WHERE id = $1", req.UserID).Scan(&balance)
		if err != nil {
			return fmt.Errorf("user not found")
		}
		if balance < req.Price*float64(req.Quantity) {
			return fmt.Errorf("insufficient balance")
		}
	} else {
		var holdings int
		err := tx.QueryRow(
			"SELECT COALESCE(quantity, 0) FROM portfolios WHERE user_id = $1 AND stock_symbol = $2",
			req.UserID, req.Stock,
		).Scan(&holdings)
		if err != nil || holdings < req.Quantity {
			return fmt.Errorf("insufficient holdings")
		}
	}
	return nil
}

func (h *OrderHandler) createOrder(tx *sql.Tx, req *models.PlaceOrderRequest) (int, error) {
	var orderID int
	err := tx.QueryRow(`
        INSERT INTO orders (user_id, stock_symbol, order_type, price, quantity, status, created_at)
        VALUES ($1, $2, $3, $4, $5, 'PENDING', NOW())
        RETURNING id
    `, req.UserID, req.Stock, req.Type, req.Price, req.Quantity).Scan(&orderID)
	return orderID, err
}

func (h *OrderHandler) processTrades(tx *sql.Tx, trades []*models.Trade) error {
	for _, trade := range trades {
		// Insert trade record
		_, err := tx.Exec(`
            INSERT INTO trades (stock_symbol, buyer_id, seller_id, buy_order_id, sell_order_id, price, quantity, executed_at)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        `, trade.Stock, trade.BuyerID, trade.SellerID, trade.BuyOrderID, trade.SellOrderID,
			trade.Price, trade.Quantity, trade.ExecutedAt)
		if err != nil {
			return err
		}

		// Update portfolios and balances
		if err := h.updateBuyerPortfolio(tx, trade); err != nil {
			return err
		}
		if err := h.updateSellerPortfolio(tx, trade); err != nil {
			return err
		}
		if err := h.updateOrderStatus(tx, trade); err != nil {
			return err
		}
	}
	return nil
}

func (h *OrderHandler) updateBuyerPortfolio(tx *sql.Tx, trade *models.Trade) error {
	// Add stock to buyer
	_, err := tx.Exec(`
        INSERT INTO portfolios (user_id, stock_symbol, quantity, avg_buy_price)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (user_id, stock_symbol)
        DO UPDATE SET 
            quantity = portfolios.quantity + $3,
            avg_buy_price = ((portfolios.avg_buy_price * portfolios.quantity) + ($4 * $3)) / (portfolios.quantity + $3)
    `, trade.BuyerID, trade.Stock, trade.Quantity, trade.Price)
	if err != nil {
		return err
	}

	// Deduct payment from buyer
	_, err = tx.Exec(`
        UPDATE users SET balance = balance - $1 WHERE id = $2
    `, trade.Price*float64(trade.Quantity), trade.BuyerID)
	return err
}

func (h *OrderHandler) updateSellerPortfolio(tx *sql.Tx, trade *models.Trade) error {
	// Remove stock from seller
	_, err := tx.Exec(`
        UPDATE portfolios SET quantity = quantity - $1
        WHERE user_id = $2 AND stock_symbol = $3
    `, trade.Quantity, trade.SellerID, trade.Stock)
	if err != nil {
		return err
	}

	// Add payment to seller
	_, err = tx.Exec(`
        UPDATE users SET balance = balance + $1 WHERE id = $2
    `, trade.Price*float64(trade.Quantity), trade.SellerID)
	return err
}

func (h *OrderHandler) updateOrderStatus(tx *sql.Tx, trade *models.Trade) error {
	_, err := tx.Exec(`
        UPDATE orders 
        SET status = 'FILLED', filled_quantity = filled_quantity + $1
        WHERE id = $2
    `, trade.Quantity, trade.BuyOrderID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
        UPDATE orders 
        SET status = 'FILLED', filled_quantity = filled_quantity + $1
        WHERE id = $2
    `, trade.Quantity, trade.SellOrderID)
	return err
}

func (h *OrderHandler) invalidateCache(stock string) {
	h.orderBookCache.mu.Lock()
	delete(h.orderBookCache.lastUpdate, stock)
	h.orderBookCache.mu.Unlock()
}
