package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"trading/models"

	"github.com/gorilla/mux"
)

// TradeHandler handles trade-related HTTP requests
type TradeHandler struct {
	db *sql.DB
}

// NewTradeHandler creates a new trade handler
func NewTradeHandler(db *sql.DB) *TradeHandler {
	return &TradeHandler{db: db}
}

// GetTrades handles GET /api/trades/{stock}
// Query params: hours (default 24), limit (default 1000)
func (h *TradeHandler) GetTrades(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stock := vars["stock"]

	// Parse query parameters
	hours := 24
	if h := r.URL.Query().Get("hours"); h != "" {
		fmt.Sscanf(h, "%d", &hours)
	}

	limit := 1000
	if l := r.URL.Query().Get("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}

	// Query trades from database
	trades, err := h.queryTrades(stock, hours, limit)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trades)
}

// queryTrades retrieves trade history for a stock
func (h *TradeHandler) queryTrades(stock string, hours, limit int) ([]models.Trade, error) {
	query := `
        SELECT 
            t.id,
            t.stock_symbol,
            t.buyer_id,
            t.seller_id,
            t.buy_order_id,
            t.sell_order_id,
            t.price,
            t.quantity,
            t.executed_at
        FROM trades t
        WHERE t.stock_symbol = $1 
          AND t.executed_at > NOW() - ($2 || ' hours')::INTERVAL
        ORDER BY t.executed_at DESC
        LIMIT $3
    `

	rows, err := h.db.Query(query, stock, fmt.Sprintf("%d", hours), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trades []models.Trade
	for rows.Next() {
		var t models.Trade
		err := rows.Scan(
			&t.ID,
			&t.Stock,
			&t.BuyerID,
			&t.SellerID,
			&t.BuyOrderID,
			&t.SellOrderID,
			&t.Price,
			&t.Quantity,
			&t.ExecutedAt,
		)
		if err != nil {
			return nil, err
		}
		trades = append(trades, t)
	}

	return trades, nil
}
