-- ============================================================
-- DISK I/O TEST DATA SETUP (SAFE + REPEATABLE)
-- Run: psql -d trading -f setup_disk_test.sql
-- WARNING: This TRUNCATES portfolios, orders, trades.
-- ============================================================

-- Step 0: Clean heavy tables
TRUNCATE trades, orders, portfolios RESTART IDENTITY CASCADE;

-- Step 0.1: Drop foreign keys on trades to avoid FK errors during random seeding
ALTER TABLE trades DROP CONSTRAINT IF EXISTS trades_buyer_id_fkey;
ALTER TABLE trades DROP CONSTRAINT IF EXISTS trades_seller_id_fkey;
ALTER TABLE trades DROP CONSTRAINT IF EXISTS trades_buy_order_id_fkey;
ALTER TABLE trades DROP CONSTRAINT IF EXISTS trades_sell_order_id_fkey;

-- ============================================================
-- Step 1: Ensure ~100,000 additional users
-- These usernames are "user_1001", "user_1002", ... (note the underscore)
-- They will NOT conflict with the original 'user1', 'user2', ... if present.
-- ============================================================

INSERT INTO users (username, balance)
SELECT 
    'user_' || gs.id,
    (random() * 100000 + 1000)::numeric(12,2)
FROM generate_series(1001, 100000) AS gs(id)
ON CONFLICT (username) DO NOTHING;

SELECT 'Users total: ' || count(*) AS info FROM users;

-- ============================================================
-- Step 2: Create portfolios (~1.5M rows)
-- For each user (id > 100), assign 15 distinct stock symbols.
-- No ON CONFLICT: we guarantee unique (user_id, stock_symbol) in this query.
-- ============================================================

-- We'll use stock symbols STK_000 to STK_999 (all <= 10 chars)
INSERT INTO portfolios (user_id, stock_symbol, quantity, avg_buy_price)
SELECT 
    u.id AS user_id,
    s.symbol,
    (10 + (random() * 500)::int) AS quantity,              -- 10–510 shares
    (50 + random() * 400)::numeric(10,2) AS avg_buy_price  -- $50–$450
FROM users u
JOIN LATERAL (
    SELECT 
        'STK_' || LPAD(i::text, 3, '0') AS symbol
    FROM generate_series(0, 999) AS i
    ORDER BY i
    LIMIT 15                                             -- 15 holdings per user
) AS s ON TRUE
WHERE u.id > 100;  -- leave first 100 "seed" users untouched if you had them

SELECT 'Portfolios total: ' || count(*) AS info FROM portfolios;

-- ============================================================
-- Step 3: Create ~100k orders
-- We create orders for first ~20k users, each with 5 stock entries.
-- ============================================================

INSERT INTO orders (user_id, stock_symbol, order_type, price, quantity, status)
SELECT 
    u.id,
    'STK_' || LPAD((i % 1000)::text, 3, '0') AS stock_symbol,
    CASE WHEN random() < 0.5 THEN 'BUY' ELSE 'SELL' END AS order_type,
    (50 + random() * 400)::numeric(10,2) AS price,
    (1 + random() * 100)::int AS quantity,
    'FILLED'
FROM users u
JOIN generate_series(1, 5) AS i ON TRUE
WHERE u.id <= 20000;   -- 20k users * 5 ≈ 100k orders

SELECT 'Orders total: ' || count(*) AS info FROM orders;

-- ============================================================
-- Step 4: Create ~5 million trades (disk I/O main table)
-- We don't enforce FKs on trades to avoid random FK failures.
-- We still use valid-looking IDs and symbols.
-- ============================================================

DO $$
DECLARE
    batch_size INT := 250000;     -- 0.25M per batch
    total_rows INT := 5000000;    -- target 5M
    inserted   INT := 0;
    max_user   INT;
    min_order  INT;
    max_order  INT;
BEGIN
    SELECT MAX(id) INTO max_user FROM users;
    SELECT MIN(id), MAX(id) INTO min_order, max_order FROM orders;

    IF max_user IS NULL OR max_user = 0 THEN
        RAISE EXCEPTION 'No users exist, cannot seed trades.';
    END IF;

    IF min_order IS NULL OR max_order IS NULL THEN
        RAISE EXCEPTION 'No orders exist, cannot seed trades.';
    END IF;

    WHILE inserted < total_rows LOOP
        INSERT INTO trades (stock_symbol, buyer_id, seller_id, buy_order_id, sell_order_id, price, quantity, executed_at)
        SELECT 
            'STK_' || LPAD((floor(random() * 1000))::int::text, 3, '0') AS stock_symbol,
            (1 + floor(random() * max_user))::int AS buyer_id,          -- may not all exist, but FK is dropped
            (1 + floor(random() * max_user))::int AS seller_id,
            (min_order + floor(random() * (max_order - min_order + 1)))::int AS buy_order_id,
            (min_order + floor(random() * (max_order - min_order + 1)))::int AS sell_order_id,
            (50 + random() * 400)::numeric(10,2) AS price,
            (1 + random() * 100)::int AS quantity,
            NOW() - (floor(random() * 60))::int * INTERVAL '1 minute'
        FROM generate_series(1, batch_size);

        inserted := inserted + batch_size;
        RAISE NOTICE 'Inserted % trades so far...', inserted;
    END LOOP;
END $$;

-- ============================================================
-- Step 5: Summary of rows and sizes
-- ============================================================

SELECT 
    'users' AS table_name,
    count(*) AS row_count,
    pg_size_pretty(pg_total_relation_size('users')) AS total_size
FROM users
UNION ALL
SELECT 
    'portfolios',
    count(*),
    pg_size_pretty(pg_total_relation_size('portfolios'))
FROM portfolios
UNION ALL
SELECT 
    'orders',
    count(*),
    pg_size_pretty(pg_total_relation_size('orders'))
FROM orders
UNION ALL
SELECT 
    'trades',
    count(*),
    pg_size_pretty(pg_total_relation_size('trades'))
FROM trades;

-- ============================================================
-- Step 6: ANALYZE tables for planner statistics
-- ============================================================

VACUUM ANALYZE users;
VACUUM ANALYZE portfolios;
VACUUM ANALYZE orders;
VACUUM ANALYZE trades;

-- ============================================================
-- Step 7: Total DB size
-- ============================================================

SELECT pg_size_pretty(pg_database_size(current_database())) AS total_db_size;

-- ============================================================
-- NOTES:
--  • FKs on trades were dropped to ensure this script never
--    fails on random IDs.
--  • Data volume: expect trades table to dominate size (GBs).
--  • You can re-add FKs later if you want strict integrity.
-- ============================================================
