package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

// DiskWriteHandler handles disk I/O intensive write operations
type DiskWriteHandler struct {
	db  *sql.DB
	sem chan struct{} // semaphore to limit concurrency
}

// NewDiskWriteHandler creates a new disk write handler
func NewDiskWriteHandler(db *sql.DB) *DiskWriteHandler {
	return &DiskWriteHandler{
		db:  db,
		sem: make(chan struct{}, 100),
	}
}

func (h *DiskWriteHandler) acquire() {
	h.sem <- struct{}{}
}

func (h *DiskWriteHandler) release() {
	<-h.sem
}

// TradeInsertRequest represents the request body for inserting trades
type TradeInsertRequest struct {
	Stock    string  `json:"stock"`
	BuyerID  int     `json:"buyer_id"`
	SellerID int     `json:"seller_id"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}

// InsertTrade handles POST /api/disk/trade
func (h *DiskWriteHandler) InsertTrade(w http.ResponseWriter, r *http.Request) {
	h.acquire()
	defer h.release()

	var req TradeInsertRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.BuyerID == 0 {
		req.BuyerID = 1
	}
	if req.SellerID == 0 {
		req.SellerID = 2
	}
	if req.Stock == "" {
		req.Stock = "TEST"
	}
	if req.Price == 0 {
		req.Price = 100.0
	}
	if req.Quantity == 0 {
		req.Quantity = 10
	}

	var tradeID int
	err := h.db.QueryRow(`
		INSERT INTO trades (stock_symbol, buyer_id, seller_id, buy_order_id, sell_order_id, price, quantity, executed_at)
		VALUES ($1, $2, $3, 1, 1, $4, $5, $6)
		RETURNING id
	`, req.Stock, req.BuyerID, req.SellerID, req.Price, req.Quantity, time.Now()).Scan(&tradeID)

	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "created",
		"trade_id": tradeID,
	})
}

// BulkInsertTrades handles POST /api/disk/trades/bulk
func (h *DiskWriteHandler) BulkInsertTrades(w http.ResponseWriter, r *http.Request) {
	h.acquire()
	defer h.release()

	var req struct {
		Count int `json:"count"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Count = 10
	}
	if req.Count <= 0 {
		req.Count = 10
	}
	if req.Count > 100 {
		req.Count = 100
	}

	tx, err := h.db.Begin()
	if err != nil {
		http.Error(w, "Transaction error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO trades (stock_symbol, buyer_id, seller_id, buy_order_id, sell_order_id, price, quantity, executed_at)
		VALUES ($1, $2, $3, 1, 1, $4, $5, NOW())
	`)
	if err != nil {
		http.Error(w, "Prepare error", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	for i := 0; i < req.Count; i++ {
		stock := "BULK_" + string(rune('A'+i%26))
		_, err := stmt.Exec(stock, 1, 2, 100.0+float64(i), 10+i)
		if err != nil {
			http.Error(w, "Insert error", http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Commit error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "created",
		"inserted": req.Count,
	})
}

// UpdateBalance handles POST /api/disk/balance
func (h *DiskWriteHandler) UpdateBalance(w http.ResponseWriter, r *http.Request) {
	h.acquire()
	defer h.release()

	var req struct {
		UserID int     `json:"user_id"`
		Amount float64 `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.UserID <= 0 {
		req.UserID = 1
	}

	result, err := h.db.Exec(`
		UPDATE users SET balance = balance + $1 WHERE id = $2
	`, req.Amount, req.UserID)

	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	rows, _ := result.RowsAffected()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":        "updated",
		"rows_affected": rows,
	})
}
