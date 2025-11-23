package main

import (
	"log"
	"net/http"

	"trading/config"
	"trading/handlers"
	"trading/matching"
	"trading/middleware"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := config.InitDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize matching engine
	engine := matching.NewMatchingEngine()

	// Initialize handlers
	orderHandler := handlers.NewOrderHandler(db, engine)
	portfolioHandler := handlers.NewPortfolioHandler(db)
	tradeHandler := handlers.NewTradeHandler(db, cfg.CacheSize, cfg.CacheTTLSeconds)
	wsHandler := handlers.NewWebSocketHandler()
	diskWriteHandler := handlers.NewDiskWriteHandler(db) // NEW

	// Setup router
	router := setupRouter(orderHandler, portfolioHandler, tradeHandler, wsHandler, diskWriteHandler)

	// Apply CORS middleware
	handler := cors.AllowAll().Handler(router)

	// Start server
	log.Printf("Trading server starting on %s", cfg.ServerAddress)
	log.Fatal(http.ListenAndServe(cfg.ServerAddress, handler))
}

func setupRouter(
	orderHandler *handlers.OrderHandler,
	portfolioHandler *handlers.PortfolioHandler,
	tradeHandler *handlers.TradeHandler,
	wsHandler *handlers.WebSocketHandler,
	diskWriteHandler *handlers.DiskWriteHandler, // NEW
) *mux.Router {
	router := mux.NewRouter()

	// API routes with middleware
	api := router.PathPrefix("/api").Subrouter()
	api.Use(middleware.Logger)
	api.Use(middleware.Recovery)

	// Health check
	api.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}).Methods("GET")

	metricsHandler := handlers.NewMetricsHandler()
	api.HandleFunc("/metrics", metricsHandler.GetMetrics).Methods("GET")

	// Order routes
	api.HandleFunc("/order/place", orderHandler.PlaceOrder).Methods("POST")
	api.HandleFunc("/order/{id}", orderHandler.CancelOrder).Methods("DELETE")
	api.HandleFunc("/orderbook/{stock}", orderHandler.GetOrderBook).Methods("GET")

	// Trade routes
	api.HandleFunc("/trades/{stock}", tradeHandler.GetTrades).Methods("GET")

	// Cache management endpoints
	api.HandleFunc("/cache/stats", tradeHandler.GetCacheStats).Methods("GET")
	api.HandleFunc("/cache/reset-stats", tradeHandler.ResetCacheStats).Methods("POST")
	api.HandleFunc("/cache/clear", tradeHandler.ClearCache).Methods("POST")

	// Portfolio routes
	api.HandleFunc("/portfolio/{user_id}", portfolioHandler.GetPortfolio).Methods("GET")

	// Disk bound routes
	api.HandleFunc("/disk/trade", diskWriteHandler.InsertTrade).Methods("POST")
	api.HandleFunc("/disk/trades/bulk", diskWriteHandler.BulkInsertTrades).Methods("POST")
	api.HandleFunc("/disk/balance", diskWriteHandler.UpdateBalance).Methods("POST")

	// WebSocket endpoint
	router.HandleFunc("/ws/trades", wsHandler.HandleConnection)

	// Serve static frontend
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./frontend")))

	return router
}
