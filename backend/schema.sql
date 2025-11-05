-- schema.sql

-- Users table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) UNIQUE NOT NULL,
    balance DECIMAL(15, 2) DEFAULT 10000.00 CHECK (balance >= 0),
    created_at TIMESTAMP DEFAULT NOW()
);

-- User portfolios (stock holdings)
CREATE TABLE portfolios (
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    stock_symbol VARCHAR(10) NOT NULL,
    quantity INT NOT NULL CHECK (quantity >= 0),
    avg_buy_price DECIMAL(10, 2) NOT NULL,
    updated_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (user_id, stock_symbol)
);

CREATE INDEX idx_portfolios_user ON portfolios(user_id);
CREATE INDEX idx_portfolios_stock ON portfolios(stock_symbol);

-- Orders table
CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    stock_symbol VARCHAR(10) NOT NULL,
    order_type VARCHAR(10) NOT NULL CHECK (order_type IN ('BUY', 'SELL')),
    price DECIMAL(10, 2) NOT NULL CHECK (price > 0),
    quantity INT NOT NULL CHECK (quantity > 0),
    filled_quantity INT DEFAULT 0 CHECK (filled_quantity >= 0),
    status VARCHAR(20) DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'FILLED', 'PARTIAL', 'CANCELLED')),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_orders_user_status ON orders(user_id, status);
CREATE INDEX idx_orders_stock_status ON orders(stock_symbol, status, created_at);
CREATE INDEX idx_orders_status ON orders(status) WHERE status = 'PENDING';

-- Trades table (executed matches)
CREATE TABLE trades (
    id SERIAL PRIMARY KEY,
    stock_symbol VARCHAR(10) NOT NULL,
    buyer_id INT REFERENCES users(id),
    seller_id INT REFERENCES users(id),
    buy_order_id INT REFERENCES orders(id),
    sell_order_id INT REFERENCES orders(id),
    price DECIMAL(10, 2) NOT NULL,
    quantity INT NOT NULL,
    executed_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_trades_stock_time ON trades(stock_symbol, executed_at DESC);
CREATE INDEX idx_trades_buyer ON trades(buyer_id, executed_at DESC);
CREATE INDEX idx_trades_seller ON trades(seller_id, executed_at DESC);
CREATE INDEX idx_trades_time ON trades(executed_at DESC);

-- Order book cache (for fast reads)
CREATE TABLE order_book_cache (
    stock_symbol VARCHAR(10) NOT NULL,
    side VARCHAR(10) NOT NULL CHECK (side IN ('BUY', 'SELL')),
    price_level DECIMAL(10, 2) NOT NULL,
    total_quantity INT NOT NULL,
    order_count INT NOT NULL,
    updated_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (stock_symbol, side, price_level)
);

CREATE INDEX idx_orderbook_stock_side ON order_book_cache(stock_symbol, side, price_level DESC);

-- Seed initial data for testing
INSERT INTO users (username, balance) 
SELECT 'user' || i, 10000.00 
FROM generate_series(1, 1000) i;

-- Create some initial stock holdings for testing
INSERT INTO portfolios (user_id, stock_symbol, quantity, avg_buy_price)
VALUES 
    (1, 'AAPL', 100, 150.00),
    (1, 'GOOGL', 50, 140.00),
    (2, 'TSLA', 75, 200.00),
    (2, 'MSFT', 200, 300.00),
    (3, 'AMZN', 30, 170.00);