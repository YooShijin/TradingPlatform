package main

import (
	"container/heap"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
)

// ============ Data Structures ============

type Order struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Stock     string    `json:"stock"`
	Type      string    `json:"type"` // "BUY" or "SELL"
	Price     float64   `json:"price"`
	Quantity  int       `json:"quantity"`
	Timestamp time.Time `json:"timestamp"`
}

type Trade struct {
	ID          int       `json:"id"`
	Stock       string    `json:"stock"`
	BuyerID     int       `json:"buyer_id"`
	SellerID    int       `json:"seller_id"`
	BuyOrderID  int       `json:"buy_order_id"`
	SellOrderID int       `json:"sell_order_id"`
	Price       float64   `json:"price"`
	Quantity    int       `json:"quantity"`
	ExecutedAt  time.Time `json:"executed_at"`
}

type OrderBook struct {
	Stock      string
	BuyOrders  *MaxHeap
	SellOrders *MinHeap
	mu         sync.RWMutex
}

type OrderBookResponse struct {
	Stock string       `json:"stock"`
	Bids  []PriceLevel `json:"bids"`
	Asks  []PriceLevel `json:"asks"`
}

type PriceLevel struct {
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}

type PortfolioStats struct {
	UserID     int       `json:"user_id"`
	Balance    float64   `json:"balance"`
	Holdings   []Holding `json:"holdings"`
	TotalValue float64   `json:"total_value"`
	TotalPL    float64   `json:"total_profit_loss"`
}

type Holding struct {
	Stock        string  `json:"stock"`
	Quantity     int     `json:"quantity"`
	AvgBuyPrice  float64 `json:"avg_buy_price"`
	CurrentPrice float64 `json:"current_price"`
	ProfitLoss   float64 `json:"profit_loss"`
}

// ============ Priority Queues ============

type MaxHeap []*Order

func (h MaxHeap) Len() int { return len(h) }
func (h MaxHeap) Less(i, j int) bool {
	// Higher price first, then earlier timestamp
	if h[i].Price != h[j].Price {
		return h[i].Price > h[j].Price
	}
	return h[i].Timestamp.Before(h[j].Timestamp)
}
func (h MaxHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h *MaxHeap) Push(x interface{}) {
	*h = append(*h, x.(*Order))
}

func (h *MaxHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

type MinHeap []*Order

func (h MinHeap) Len() int { return len(h) }
func (h MinHeap) Less(i, j int) bool {
	// Lower price first, then earlier timestamp
	if h[i].Price != h[j].Price {
		return h[i].Price < h[j].Price
	}
	return h[i].Timestamp.Before(h[j].Timestamp)
}
func (h MinHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h *MinHeap) Push(x interface{}) {
	*h = append(*h, x.(*Order))
}
func (h *MinHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// ============ Matching Engine ============

type MatchingEngine struct {
	orderBooks map[string]*OrderBook
	mu         sync.RWMutex
}

func NewMatchingEngine() *MatchingEngine {
	return &MatchingEngine{
		orderBooks: make(map[string]*OrderBook),
	}
}

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

func (me *MatchingEngine) PlaceOrder(order *Order) []*Trade {
	book := me.getOrCreateBook(order.Stock)
	book.mu.Lock()
	defer book.mu.Unlock()

	var trades []*Trade

	if order.Type == "BUY" {
		// Try to match with sell orders
		for order.Quantity > 0 && book.SellOrders.Len() > 0 {
			bestSell := (*book.SellOrders)[0]

			if bestSell.Price <= order.Price {
				tradeQty := min(order.Quantity, bestSell.Quantity)

				trade := &Trade{
					Stock:       order.Stock,
					BuyerID:     order.UserID,
					SellerID:    bestSell.UserID,
					BuyOrderID:  order.ID,
					SellOrderID: bestSell.ID,
					Price:       bestSell.Price,
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
				break
			}
		}

		// Add remaining to order book
		if order.Quantity > 0 {
			heap.Push(book.BuyOrders, order)
		}
	} else { // SELL
		// Try to match with buy orders
		for order.Quantity > 0 && book.BuyOrders.Len() > 0 {
			bestBuy := (*book.BuyOrders)[0]

			if bestBuy.Price >= order.Price {
				tradeQty := min(order.Quantity, bestBuy.Quantity)

				trade := &Trade{
					Stock:       order.Stock,
					BuyerID:     bestBuy.UserID,
					SellerID:    order.UserID,
					BuyOrderID:  bestBuy.ID,
					SellOrderID: order.ID,
					Price:       bestBuy.Price,
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
				break
			}
		}

		// Add remaining to order book
		if order.Quantity > 0 {
			heap.Push(book.SellOrders, order)
		}
	}

	return trades
}

func (me *MatchingEngine) GetOrderBook(stock string) OrderBookResponse {
	book := me.getOrCreateBook(stock)
	book.mu.RLock()
	defer book.mu.RUnlock()

	resp := OrderBookResponse{
		Stock: stock,
		Bids:  []PriceLevel{},
		Asks:  []PriceLevel{},
	}

	// Aggregate buy orders by price
	bidMap := make(map[float64]int)
	for _, order := range *book.BuyOrders {
		bidMap[order.Price] += order.Quantity
	}
	for price, qty := range bidMap {
		resp.Bids = append(resp.Bids, PriceLevel{Price: price, Quantity: qty})
	}

	// Aggregate sell orders by price
	askMap := make(map[float64]int)
	for _, order := range *book.SellOrders {
		askMap[order.Price] += order.Quantity
	}
	for price, qty := range askMap {
		resp.Asks = append(resp.Asks, PriceLevel{Price: price, Quantity: qty})
	}

	return resp
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ============ Server ============

type Server struct {
	db             *sql.DB
	matchingEngine *MatchingEngine
	orderBookCache struct {
		books      map[string]*OrderBook
		lastUpdate map[string]time.Time
		mu         sync.RWMutex
	}
	wsClients map[*websocket.Conn]bool
	wsMu      sync.Mutex
	upgrader  websocket.Upgrader
}

func NewServer(dbURL string) (*Server, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}

	server := &Server{
		db:             db,
		matchingEngine: NewMatchingEngine(),
		wsClients:      make(map[*websocket.Conn]bool),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}

	server.orderBookCache.books = make(map[string]*OrderBook)
	server.orderBookCache.lastUpdate = make(map[string]time.Time)

	return server, nil
}

// ============ HTTP Handlers ============

func (s *Server) handlePlaceOrder(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID   int     `json:"user_id"`
		Stock    string  `json:"stock"`
		Type     string  `json:"type"`
		Price    float64 `json:"price"`
		Quantity int     `json:"quantity"`
	}

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
	tx, err := s.db.Begin()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Validate user balance/holdings
	if req.Type == "BUY" {
		var balance float64
		err := tx.QueryRow("SELECT balance FROM users WHERE id = $1", req.UserID).Scan(&balance)
		if err != nil || balance < req.Price*float64(req.Quantity) {
			http.Error(w, "Insufficient balance", http.StatusBadRequest)
			return
		}
	} else {
		var holdings int
		err := tx.QueryRow(
			"SELECT COALESCE(quantity, 0) FROM portfolios WHERE user_id = $1 AND stock_symbol = $2",
			req.UserID, req.Stock,
		).Scan(&holdings)
		if err != nil || holdings < req.Quantity {
			http.Error(w, "Insufficient holdings", http.StatusBadRequest)
			return
		}
	}

	// Insert order into database
	var orderID int
	err = tx.QueryRow(`
        INSERT INTO orders (user_id, stock_symbol, order_type, price, quantity, status, created_at)
        VALUES ($1, $2, $3, $4, $5, 'PENDING', NOW())
        RETURNING id
    `, req.UserID, req.Stock, req.Type, req.Price, req.Quantity).Scan(&orderID)

	if err != nil {
		http.Error(w, "Failed to create order", http.StatusInternalServerError)
		return
	}

	// Create order object
	order := &Order{
		ID:        orderID,
		UserID:    req.UserID,
		Stock:     req.Stock,
		Type:      req.Type,
		Price:     req.Price,
		Quantity:  req.Quantity,
		Timestamp: time.Now(),
	}

	// Match order using in-memory engine
	trades := s.matchingEngine.PlaceOrder(order)

	// Process each trade
	for _, trade := range trades {
		// Insert trade record
		_, err := tx.Exec(`
            INSERT INTO trades (stock_symbol, buyer_id, seller_id, buy_order_id, sell_order_id, price, quantity, executed_at)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        `, trade.Stock, trade.BuyerID, trade.SellerID, trade.BuyOrderID, trade.SellOrderID,
			trade.Price, trade.Quantity, trade.ExecutedAt)

		if err != nil {
			log.Printf("Failed to insert trade: %v", err)
			continue
		}

		// Update buyer portfolio (add stock)
		_, err = tx.Exec(`
            INSERT INTO portfolios (user_id, stock_symbol, quantity, avg_buy_price)
            VALUES ($1, $2, $3, $4)
            ON CONFLICT (user_id, stock_symbol)
            DO UPDATE SET 
                quantity = portfolios.quantity + $3,
                avg_buy_price = ((portfolios.avg_buy_price * portfolios.quantity) + ($4 * $3)) / (portfolios.quantity + $3)
        `, trade.BuyerID, trade.Stock, trade.Quantity, trade.Price)

		// Update seller portfolio (remove stock)
		_, err = tx.Exec(`
            UPDATE portfolios 
            SET quantity = quantity - $1
            WHERE user_id = $2 AND stock_symbol = $3
        `, trade.Quantity, trade.SellerID, trade.Stock)

		// Update buyer balance (deduct payment)
		_, err = tx.Exec(`
            UPDATE users 
            SET balance = balance - $1
            WHERE id = $2
        `, trade.Price*float64(trade.Quantity), trade.BuyerID)

		// Update seller balance (add payment)
		_, err = tx.Exec(`
            UPDATE users 
            SET balance = balance + $1
            WHERE id = $2
        `, trade.Price*float64(trade.Quantity), trade.SellerID)

		// Update order status
		_, err = tx.Exec(`
            UPDATE orders 
            SET status = 'FILLED', filled_quantity = filled_quantity + $1
            WHERE id = $2
        `, trade.Quantity, trade.BuyOrderID)

		_, err = tx.Exec(`
            UPDATE orders 
            SET status = 'FILLED', filled_quantity = filled_quantity + $1
            WHERE id = $2
        `, trade.Quantity, trade.SellOrderID)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	// Broadcast trade to WebSocket clients
	for _, trade := range trades {
		s.broadcastTrade(trade)
	}

	// Invalidate order book cache
	s.invalidateOrderBookCache(req.Stock)

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "success",
		"order_id":  orderID,
		"trades":    trades,
		"remaining": order.Quantity,
	})
}

func (s *Server) handleGetOrderBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stock := vars["stock"]

	// Check cache
	s.orderBookCache.mu.RLock()
	lastUpdate, exists := s.orderBookCache.lastUpdate[stock]
	s.orderBookCache.mu.RUnlock()

	// If cache is fresh (< 100ms), use in-memory data
	if exists && time.Since(lastUpdate) < 100*time.Millisecond {
		orderBook := s.matchingEngine.GetOrderBook(stock)

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		json.NewEncoder(w).Encode(orderBook)
		return
	}

	// Cache miss or stale - get from matching engine
	orderBook := s.matchingEngine.GetOrderBook(stock)

	// Update cache
	s.orderBookCache.mu.Lock()
	s.orderBookCache.lastUpdate[stock] = time.Now()
	s.orderBookCache.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	json.NewEncoder(w).Encode(orderBook)
}

func (s *Server) handleGetPortfolio(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var userID int
	fmt.Sscanf(vars["user_id"], "%d", &userID)

	// Get user balance
	var balance float64
	err := s.db.QueryRow("SELECT balance FROM users WHERE id = $1", userID).Scan(&balance)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Get holdings with current prices
	query := `
        SELECT 
            p.stock_symbol,
            p.quantity,
            p.avg_buy_price,
            COALESCE(recent.current_price, p.avg_buy_price) as current_price
        FROM portfolios p
        LEFT JOIN (
            SELECT stock_symbol, 
                   AVG(price) as current_price
            FROM trades
            WHERE executed_at > NOW() - INTERVAL '5 minutes'
            GROUP BY stock_symbol
        ) recent ON p.stock_symbol = recent.stock_symbol
        WHERE p.user_id = $1 AND p.quantity > 0
    `

	rows, err := s.db.Query(query, userID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var holdings []Holding
	var totalValue float64
	var totalPL float64

	for rows.Next() {
		var h Holding
		rows.Scan(&h.Stock, &h.Quantity, &h.AvgBuyPrice, &h.CurrentPrice)

		h.ProfitLoss = (h.CurrentPrice - h.AvgBuyPrice) * float64(h.Quantity)
		totalValue += h.CurrentPrice * float64(h.Quantity)
		totalPL += h.ProfitLoss

		holdings = append(holdings, h)
	}

	stats := PortfolioStats{
		UserID:     userID,
		Balance:    balance,
		Holdings:   holdings,
		TotalValue: totalValue,
		TotalPL:    totalPL,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (s *Server) handleGetTrades(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stock := vars["stock"]

	hours := 24
	if h := r.URL.Query().Get("hours"); h != "" {
		fmt.Sscanf(h, "%d", &hours)
	}

	limit := 1000
	if l := r.URL.Query().Get("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}

	query := `
    SELECT 
        t.id,
        t.stock_symbol,
        t.buyer_id,
        t.seller_id,
        t.price,
        t.quantity,
        t.executed_at
    FROM trades t
    WHERE t.stock_symbol = $1 
      AND t.executed_at > NOW() - ($2 || ' hours')::INTERVAL
    ORDER BY t.executed_at DESC
    LIMIT $3
`

	rows, err := s.db.Query(query, stock, fmt.Sprintf("%d", hours), limit)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var trades []Trade
	for rows.Next() {
		var t Trade
		rows.Scan(&t.ID, &t.Stock, &t.BuyerID, &t.SellerID, &t.Price, &t.Quantity, &t.ExecutedAt)
		trades = append(trades, t)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trades)
}

func (s *Server) handleCancelOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var orderID int
	fmt.Sscanf(vars["id"], "%d", &orderID)

	// Update order status in database
	result, err := s.db.Exec(`
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

	// TODO: Remove from in-memory order book (requires additional tracking)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "cancelled"})
}

// ============ WebSocket Handler ============

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	s.wsMu.Lock()
	s.wsClients[conn] = true
	s.wsMu.Unlock()

	log.Println("🔌 WebSocket client connected")

	// Keep connection alive with ping messages
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				break
			}
		}
	}()

	defer func() {
		s.wsMu.Lock()
		delete(s.wsClients, conn)
		s.wsMu.Unlock()
		conn.Close()
		log.Println(" WebSocket client disconnected")
	}()

	// Keep reading to detect disconnect
	for {
		if _, _, err := conn.NextReader(); err != nil {
			break
		}
	}
}

func (s *Server) broadcastTrade(trade *Trade) {
	s.wsMu.Lock()
	defer s.wsMu.Unlock()

	message, _ := json.Marshal(trade)

	for client := range s.wsClients {
		err := client.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			client.Close()
			delete(s.wsClients, client)
		}
	}
}

func (s *Server) invalidateOrderBookCache(stock string) {
	s.orderBookCache.mu.Lock()
	delete(s.orderBookCache.lastUpdate, stock)
	s.orderBookCache.mu.Unlock()
}

// ============ Main ============

func main() {
	// dbURL := "postgres://user:password@localhost/trading?sslmode=disable"
	dbURL := "postgres://user:asmit2003@localhost/trading?sslmode=disable"

	server, err := NewServer(dbURL)
	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()

	// API routes
	router.HandleFunc("/api/order/place", server.handlePlaceOrder).Methods("POST")
	router.HandleFunc("/api/order/{id}", server.handleCancelOrder).Methods("DELETE")
	router.HandleFunc("/api/orderbook/{stock}", server.handleGetOrderBook).Methods("GET")
	router.HandleFunc("/api/trades/{stock}", server.handleGetTrades).Methods("GET")
	router.HandleFunc("/api/portfolio/{user_id}", server.handleGetPortfolio).Methods("GET")

	// WebSocket
	router.HandleFunc("/ws/trades", server.handleWebSocket)

	// Serve static frontend
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./frontend")))

	//  //Add CORS middleware here
	// handler := cors.New(cors.Options{
	// 	AllowedOrigins:   []string{"http://localhost:3000", "http://127.0.0.1:3000"},
	// 	AllowedMethods:   []string{"GET", "POST", "DELETE", "PUT", "OPTIONS"},
	// 	AllowedHeaders:   []string{"Authorization", "Content-Type"},
	// 	AllowCredentials: true,
	// }).Handler(router)
	handler := cors.AllowAll().Handler(router)

	log.Println("Trading server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
