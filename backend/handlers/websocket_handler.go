package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"trading/models"

	"github.com/gorilla/websocket"
)

// WebSocketHandler manages WebSocket connections for real-time trade updates
type WebSocketHandler struct {
	clients  map[*websocket.Conn]bool
	mu       sync.Mutex
	upgrader websocket.Upgrader
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler() *WebSocketHandler {
	return &WebSocketHandler{
		clients: make(map[*websocket.Conn]bool),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins (configure for production)
			},
		},
	}
}

// HandleConnection upgrades HTTP connection to WebSocket
func (h *WebSocketHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Register new client
	h.mu.Lock()
	h.clients[conn] = true
	h.mu.Unlock()

	log.Printf("🔌 WebSocket client connected (total: %d)", len(h.clients))

	// Start ping routine to keep connection alive
	go h.keepAlive(conn)

	defer func() {
		h.mu.Lock()
		delete(h.clients, conn)
		h.mu.Unlock()
		conn.Close()
		log.Printf("❌ WebSocket client disconnected (remaining: %d)", len(h.clients))
	}()

	// Read messages to detect disconnect
	for {
		if _, _, err := conn.NextReader(); err != nil {
			break
		}
	}
}

// BroadcastTrade sends trade update to all connected clients
func (h *WebSocketHandler) BroadcastTrade(trade *models.Trade) {
	h.mu.Lock()
	defer h.mu.Unlock()

	message, err := json.Marshal(trade)
	if err != nil {
		log.Printf("Failed to marshal trade: %v", err)
		return
	}

	// Send to all connected clients
	for client := range h.clients {
		err := client.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("Failed to send to client: %v", err)
			client.Close()
			delete(h.clients, client)
		}
	}
}

// keepAlive sends periodic ping messages to keep connection alive
func (h *WebSocketHandler) keepAlive(conn *websocket.Conn) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
			return
		}
	}
}
