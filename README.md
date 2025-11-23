# Trading System Backend (Go)

A high-performance backend for a stock trading platform.  
Supports real-time order matching, portfolio tracking, WebSocket live trade updates, and LRU caching for fast reads.

---

## Features

- Buy and sell order placement
- In-memory matching engine (price-time priority)
- Portfolio management with P/L calculation
- Real-time WebSocket trade updates
- LRU cache for frequently requested trade history
- PostgreSQL persistence
- Metrics API for system load benchmarking

---

## Architecture

```mermaid
flowchart TD

    subgraph Clients
        UI["Trading App"]
        Admin["Admin Dashboard"]
        LG["Load Generator"]
    end

    subgraph TradingServer["Go Trading Server"]
        OrdersAPI["POST /api/order"]
        TradesAPI["GET /api/trades"]
        PortfolioAPI["GET /api/portfolio"]
        DiskAPI["POST /api/disk"]
        MetricsAPI["GET /api/metrics"]
        WebSocketAPI["WS /ws/trades"]

        OrderHandler
        TradeHandler
        PortfolioHandler
        DiskWriteHandler
        MetricsHandler
        WebSocketHandler

        MatchingEngine["Matching Engine (In-Memory)"]
        Cache["LRU Cache (Trade Results)"]
    end

    DB[("PostgreSQL Database")]
    OS["CPU / Memory / Disk / Network"]

    UI --> OrdersAPI --> OrderHandler --> MatchingEngine
    UI --> TradesAPI --> TradeHandler
    UI --> PortfolioAPI --> PortfolioHandler
    UI --> WebSocketAPI --> WebSocketHandler

    LG --> TradesAPI
    LG --> DiskAPI
    LG --> MetricsAPI

    OrderHandler --> DB
    TradeHandler --> Cache
    TradeHandler --> DB
    PortfolioHandler --> DB
    DiskWriteHandler --> DB
    MetricsHandler --> OS
    MatchingEngine --> WebSocketHandler
```

````

---

## API Endpoints

| Method | Endpoint                   | Description                                |
| ------ | -------------------------- | ------------------------------------------ |
| POST   | `/api/order/place`         | Place a buy/sell order                     |
| GET    | `/api/orderbook/{stock}`   | Get order book                             |
| DELETE | `/api/order/{id}`          | Cancel order                               |
| GET    | `/api/trades/{stock}`      | Get trade history (cached)                 |
| GET    | `/api/portfolio/{user_id}` | Get portfolio + P/L                        |
| WS     | `/ws/trades`               | Live trade WebSocket updates               |
| GET    | `/api/metrics`             | System CPU / memory / disk / network stats |

---

## Tech Stack

- Go
- PostgreSQL
- Gorilla Mux
- WebSockets
- Custom LRU Cache

---

## Running the Server

```bash
go run main.go
```

**Environment variables**

```bash
DATABASE_URL=postgres://username:password@localhost:5432/trading
CACHE_SIZE=100
CACHE_TTL_SECONDS=1
```

---

## Load Testing

A load generator is included to test three types of performance workloads:

| Workload Mode | Description                          |
| ------------- | ------------------------------------ |
| cpu-bound     | Cached reads (no DB load)            |
| io-bound      | Cold read queries that always hit DB |
| disk-write    | Heavy high-volume insert/write load  |

Example:

```bash
go run loadgen.go -clients=50 -duration=60 -workload=cpu-bound
```

---

## Folder Structure

```
/handlers      HTTP request handlers
/matching      Matching engine (in-memory)
/cache         LRU cache implementation
/models        Shared data structs
/config        Database + environment setup
/middleware    Logging + panic recovery
```

---

## Project Status

This backend focuses on performance and core trading logic.
Authentication, UI and risk control layers can be added on top.

---


````
