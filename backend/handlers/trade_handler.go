package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"trading/cache"
	"trading/models"

	"github.com/gorilla/mux"
)

// TradeHandler handles trade-related HTTP requests with caching
type TradeHandler struct {
	db       *sql.DB
	cache    *cache.LRUCache
	cacheTTL time.Duration
}

// NewTradeHandler creates a new trade handler with cache
func NewTradeHandler(db *sql.DB, cacheSize int, cacheTTLSeconds int) *TradeHandler {
	return &TradeHandler{
		db:       db,
		cache:    cache.NewLRUCache(cacheSize),
		cacheTTL: time.Duration(cacheTTLSeconds) * time.Second,
	}
}

// GetTrades handles GET /api/trades/{stock}
// Query params: hours (default 24), limit (default 1000)
func (h *TradeHandler) GetTrades(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stock := vars["stock"]

	// Parse query parameters
	hours := 24
	if hoursStr := r.URL.Query().Get("hours"); hoursStr != "" {
		if parsed, err := strconv.Atoi(hoursStr); err == nil {
			hours = parsed
		}
	}

	limit := 1000
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil {
			limit = parsed
		}
	}

	// Create cache key
	cacheKey := fmt.Sprintf("trades:%s:%d:%d", stock, hours, limit)

	// Check cache first
	if cachedData, found := h.cache.Get(cacheKey); found {
		// Cache hit - return from memory
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		w.Header().Set("X-Cache-Key", cacheKey)
		json.NewEncoder(w).Encode(cachedData)
		return
	}

	// Cache miss - query database
	trades, err := h.queryTrades(stock, hours, limit)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Store in cache
	h.cache.Set(cacheKey, trades, h.cacheTTL)

	// Return response with cache miss header
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	w.Header().Set("X-Cache-Key", cacheKey)
	json.NewEncoder(w).Encode(trades)
}

// queryTrades retrieves trade history for a stock from database
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

// GetRecentTrades handles GET /api/trades/recent
func (h *TradeHandler) GetRecentTrades(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil {
			limit = parsed
		}
	}

	cacheKey := fmt.Sprintf("recent_trades:%d", limit)

	// Check cache
	if cachedData, found := h.cache.Get(cacheKey); found {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		json.NewEncoder(w).Encode(cachedData)
		return
	}

	// Query database
	query := `
		SELECT 
			id, stock_symbol, buyer_id, seller_id, 
			buy_order_id, sell_order_id, price, quantity, executed_at
		FROM trades
		ORDER BY executed_at DESC
		LIMIT $1
	`

	rows, err := h.db.Query(query, limit)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var trades []models.Trade
	for rows.Next() {
		var t models.Trade
		err := rows.Scan(
			&t.ID, &t.Stock, &t.BuyerID, &t.SellerID,
			&t.BuyOrderID, &t.SellOrderID, &t.Price, &t.Quantity, &t.ExecutedAt,
		)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		trades = append(trades, t)
	}

	// Cache result
	h.cache.Set(cacheKey, trades, h.cacheTTL)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	json.NewEncoder(w).Encode(trades)
}

// InvalidateCache clears all cached trade data
func (h *TradeHandler) InvalidateCache() {
	h.cache.Clear()
}

// InvalidateStockCache clears cached data for a specific stock
func (h *TradeHandler) InvalidateStockCache(stock string) {
	// This is a simple implementation
	// In production, you'd track which keys belong to which stocks
	h.cache.Clear()
}

// GetCacheStats handles GET /api/cache/stats
func (h *TradeHandler) GetCacheStats(w http.ResponseWriter, r *http.Request) {
	hits, misses, hitRate := h.cache.GetStats()

	stats := map[string]interface{}{
		"cache_hits":     hits,
		"cache_misses":   misses,
		"cache_hit_rate": fmt.Sprintf("%.2f%%", hitRate),
		"cache_size":     h.cache.Size(),
		"cache_ttl_sec":  int(h.cacheTTL.Seconds()),
		"timestamp":      time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// ResetCacheStats handles POST /api/cache/reset-stats
func (h *TradeHandler) ResetCacheStats(w http.ResponseWriter, r *http.Request) {
	h.cache.ResetStats()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message":   "Cache statistics reset successfully",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// ClearCache handles POST /api/cache/clear
func (h *TradeHandler) ClearCache(w http.ResponseWriter, r *http.Request) {
	h.cache.Clear()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message":   "Cache cleared successfully",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}
