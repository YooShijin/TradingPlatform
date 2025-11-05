package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"trading/models"

	"github.com/gorilla/mux"
)

// PortfolioHandler handles portfolio-related HTTP requests
type PortfolioHandler struct {
	db *sql.DB
}

// NewPortfolioHandler creates a new portfolio handler
func NewPortfolioHandler(db *sql.DB) *PortfolioHandler {
	return &PortfolioHandler{db: db}
}

// GetPortfolio handles GET /api/portfolio/{user_id}
func (h *PortfolioHandler) GetPortfolio(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var userID int
	fmt.Sscanf(vars["user_id"], "%d", &userID)

	// Get user balance
	var balance float64
	err := h.db.QueryRow("SELECT balance FROM users WHERE id = $1", userID).Scan(&balance)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Get holdings with current market prices
	holdings, err := h.getHoldings(userID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Calculate totals
	var totalValue float64
	var totalPL float64
	for _, h := range holdings {
		totalValue += h.CurrentPrice * float64(h.Quantity)
		totalPL += h.ProfitLoss
	}

	stats := models.PortfolioStats{
		UserID:     userID,
		Balance:    balance,
		Holdings:   holdings,
		TotalValue: totalValue,
		TotalPL:    totalPL,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// getHoldings retrieves user's stock holdings with current prices
func (h *PortfolioHandler) getHoldings(userID int) ([]models.Holding, error) {
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

	rows, err := h.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var holdings []models.Holding
	for rows.Next() {
		var holding models.Holding
		err := rows.Scan(
			&holding.Stock,
			&holding.Quantity,
			&holding.AvgBuyPrice,
			&holding.CurrentPrice,
		)
		if err != nil {
			return nil, err
		}

		// Calculate profit/loss
		holding.ProfitLoss = (holding.CurrentPrice - holding.AvgBuyPrice) * float64(holding.Quantity)
		holdings = append(holdings, holding)
	}

	return holdings, nil
}
