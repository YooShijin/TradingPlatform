# Trading System Architecture

## Overview

A high-performance stock trading platform built with Go that provides real-time order matching, portfolio management, and live trade updates via WebSocket.

## System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         Frontend                            │
│                  (Static Files/Next)                        │
└──────────────────────┬──────────────────────────────────────┘
                       │ HTTP/WebSocket
┌──────────────────────┴──────────────────────────────────────┐
│                      Main Server                            │
│                    (main.go + Router)                       │
└────────┬─────────┬─────────┬──────────┬─────────────────────┘
         │         │         │          │
    ┌────▼───┐ ┌───▼───┐ ┌───▼─────┐ ┌──▼─────────┐
    │ Order  │ │Trade  │ │Portfolio│ │ WebSocket  │
    │Handler │ │Handler│ │ Handler │ │  Handler   │
    └────┬───┘ └──┬────┘ └────┬────┘ └──┬─────────┘
         │        │           │         │
         │   ┌────▼───────────▼─────────▼───┐
         │   │      PostgreSQL Database     │
         │   └──────────────────────────────┘
         │
    ┌────▼───────────────────┐
    │  Matching Engine       │
    │  (In-Memory)           │
    │  ┌──────────────────┐  │
    │  │  Order Books     │  │
    │  │  (Per Stock)     │  │
    │  │  ┌────┐   ┌────┐ │  │
    │  │  │Buy │   │Sell│ │  │
    │  │  │Heap│   │Heap│ │  │
    │  │  └────┘   └────┘ │  │
    │  └──────────────────┘  │
    └────────────────────────┘
```

## Module Breakdown

### 1. **main.go** - Entry Point

**Purpose**: Application bootstrap and routing setup

**Key Responsibilities**:

- Initialize database connection
- Create matching engine instance
- Set up HTTP router with all routes
- Apply CORS middleware
- Start HTTP server

**Flow**:

```
main() → Load Config → Init DB → Create Handlers → Setup Routes → Start Server
```

### 2. **config/** - Configuration Management

#### config.go

**Purpose**: Centralized configuration

**Features**:

- Environment variable loading
- Database URL management
- Connection pool configuration
- Default values for development

**Usage**:

```go
cfg := config.Load()
db, err := config.InitDB(cfg.DatabaseURL)
```

### 3. **models/** - Data Structures

#### models.go

**Purpose**: Define all data types used across the system

**Key Models**:

- **Order**: Represents a buy/sell order
- **Trade**: Executed trade between two parties
- **OrderBookResponse**: Aggregated market depth data
- **PortfolioStats**: User's portfolio with P&L
- **Holding**: Individual stock position

**Design Pattern**: DTOs (Data Transfer Objects) for clean API contracts

### 4. **matching/** - Order Matching Engine

#### engine.go

**Purpose**: Core business logic for order matching

**How It Works**:

1. **Order Book per Stock**: Each stock has separate buy/sell heaps
2. **Price-Time Priority**: Orders matched by best price, then earliest time
3. **Immediate or Cancel (IOC)**: Orders match immediately or rest in book

**Matching Algorithm**:

```
For BUY order at price P:
  While order has quantity AND sell orders exist:
    If best sell price ≤ P:
      Execute trade at sell price (resting order price)
      Reduce quantities
      Remove filled orders
    Else:
      Break (no more matches)
  If quantity remains:
    Add to buy order book
```

**Thread Safety**: Uses RWMutex for concurrent access

#### heaps.go

**Purpose**: Priority queue implementation

**MaxHeap (Buy Orders)**:

- Highest price first
- FIFO for same price
- Used for buy orders

**MinHeap (Sell Orders)**:

- Lowest price first
- FIFO for same price
- Used for sell orders

**Example**:

```
Buy Orders (MaxHeap):     Sell Orders (MinHeap):
[100.50 - 10 shares]      [99.50 - 5 shares]
[100.00 - 20 shares]      [99.75 - 15 shares]
[99.75 - 5 shares]        [100.00 - 10 shares]
```

### 5. **handlers/** - HTTP Request Handlers

#### order_handler.go

**Endpoints**:

- `POST /api/order/place` - Place new order
- `GET /api/orderbook/{stock}` - Get market depth
- `DELETE /api/order/{id}` - Cancel order

**Order Placement Flow**:

```
1. Validate request (type, price, quantity)
2. Start database transaction
3. Check user balance/holdings
4. Insert order record
5. Match order in-memory
6. Process all trades:
   - Insert trade records
   - Update buyer portfolio (add stocks)
   - Update seller portfolio (remove stocks)
   - Update buyer balance (deduct payment)
   - Update seller balance (add payment)
   - Mark orders as filled
7. Commit transaction
8. Broadcast trades via WebSocket
9. Invalidate cache
10. Return response
```

**Caching Strategy**:

- Order book cached for 100ms
- Invalidated on new orders
- Reduces database load
- Provides sub-second updates

#### portfolio_handler.go

**Endpoints**:

- `GET /api/portfolio/{user_id}` - Get user portfolio

**Features**:

- Current holdings with quantities
- Average buy price calculation
- Current market price (from recent trades)
- Real-time P&L calculation
- Total portfolio value

**Price Discovery**:

- Uses average price of trades in last 5 minutes
- Falls back to average buy price if no recent trades

#### trade_handler.go

**Endpoints**:

- `GET /api/trades/{stock}` - Get trade history

**Query Parameters**:

- `hours`: Time window (default 24)
- `limit`: Max results (default 1000)

**Use Cases**:

- Market analysis
- Price charts
- Volume tracking
- Order book reconstruction

#### websocket_handler.go

**Endpoint**:

- `GET /ws/trades` - WebSocket connection

**Features**:

- Real-time trade broadcasting
- Connection keep-alive (30s ping)
- Automatic cleanup on disconnect
- Concurrent client management

**Message Format**:

```json
{
  "id": 123,
  "stock": "AAPL",
  "buyer_id": 1,
  "seller_id": 2,
  "price": 150.25,
  "quantity": 10,
  "executed_at": "2025-01-15T10:30:00Z"
}
```

### 6. **middleware/** - HTTP Middleware

#### middleware.go

**Components**:

- **Logger**: Logs all requests with method, path, status, duration
- **Recovery**: Catches panics and returns 500 error

**Usage**:

```go
api.Use(middleware.Logger)
api.Use(middleware.Recovery)
```

## Database Schema

### Tables

**users**

```sql
id SERIAL PRIMARY KEY
balance DECIMAL(15,2)
created_at TIMESTAMP
```

**orders**

```sql
id SERIAL PRIMARY KEY
user_id INT REFERENCES users(id)
stock_symbol VARCHAR(10)
order_type VARCHAR(4) -- 'BUY' or 'SELL'
price DECIMAL(10,2)
quantity INT
filled_quantity INT DEFAULT 0
status VARCHAR(20) -- 'PENDING', 'FILLED', 'CANCELLED'
created_at TIMESTAMP
```

**trades**

```sql
id SERIAL PRIMARY KEY
stock_symbol VARCHAR(10)
buyer_id INT REFERENCES users(id)
seller_id INT REFERENCES users(id)
buy_order_id INT REFERENCES orders(id)
sell_order_id INT REFERENCES orders(id)
price DECIMAL(10,2)
quantity INT
executed_at TIMESTAMP
```

**portfolios**

```sql
user_id INT REFERENCES users(id)
stock_symbol VARCHAR(10)
quantity INT
avg_buy_price DECIMAL(10,2)
PRIMARY KEY (user_id, stock_symbol)
```

## Key Algorithms

### 1. Order Matching (Price-Time Priority)

```
Priority = (Price Level, Timestamp)

For Buy Orders: Higher price wins, then earlier time
For Sell Orders: Lower price wins, then earlier time
```

### 2. Average Price Calculation

```
New Avg = (Old Avg * Old Qty + Trade Price * Trade Qty) / (Old Qty + Trade Qty)
```

### 3. Portfolio P&L

```
Profit/Loss = (Current Price - Avg Buy Price) × Quantity
```

## Performance Optimizations

### 1. In-Memory Matching

- Orders matched in RAM using heaps
- O(log n) insertion and removal
- No database queries during matching

### 2. Caching

- Order book cached for 100ms
- Reduces redundant heap scans
- Cache invalidation on updates

### 3. Database Transactions

- ACID guarantees for trades
- Rollback on failures
- Consistent portfolio state

### 4. Connection Pooling

- Reuses database connections
- MaxOpenConns: 25
- MaxIdleConns: 5

### 5. Concurrent Request Handling

- Goroutines per request
- Mutex protection for shared state
- Non-blocking WebSocket broadcasts

## Error Handling

### Validation Errors (400)

- Invalid order type
- Negative price/quantity
- Insufficient balance/holdings

### Not Found Errors (404)

- User doesn't exist
- Order not found

### Server Errors (500)

- Database connection failures
- Transaction commit failures
- Unexpected panics (caught by recovery middleware)

## Security Considerations

### Current Implementation

- CORS enabled for all origins (development)
- No authentication/authorization
- SQL injection prevented (parameterized queries)

### Production Requirements

- Add JWT authentication
- Rate limiting per user
- Strict CORS policy
- Input sanitization
- API key for WebSocket
- HTTPS/TLS encryption
- Audit logging

## Testing Strategy

### Unit Tests

- Heap operations
- Order matching logic
- Price calculations

### Integration Tests

- API endpoints
- Database transactions
- WebSocket connections

### Load Tests

- Concurrent order placement
- High-frequency trading simulation
- WebSocket scalability

## Deployment

### Development

```bash
go run main.go
# Server runs on :8080
```

### Production

```bash
go build -o trading-server
./trading-server
```

### Environment Variables

```bash
export DATABASE_URL="postgres://user:pass@host/db"
export SERVER_ADDRESS=":8080"
```

## Future Enhancements

1. **Order Types**: Limit, Market, Stop-Loss
2. **Order Modification**: Change price/quantity
3. **Partial Fills**: Better tracking
4. **Market Data**: OHLCV candlesticks
5. **Analytics**: Volume, volatility metrics
6. **Admin Panel**: User management
7. **Margin Trading**: Leverage support
8. **Multi-Asset**: Options, futures
9. **Kubernetes**: Container orchestration
10. **Monitoring**: Prometheus/Grafana

## Conclusion

This trading system provides a solid foundation for a stock exchange with:

- ✅ Real-time order matching
- ✅ ACID-compliant trades
- ✅ Live market data
- ✅ Portfolio tracking
- ✅ WebSocket updates
- ✅ Modular architecture
- ✅ Performance optimizations

The modular design makes it easy to extend and maintain while handling thousands of concurrent trades per second.
