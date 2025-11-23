#!/bin/bash
# ============================================================
# DISK I/O BOTTLENECK SETUP AND TEST SCRIPT
# ============================================================

set -e

SERVER_URL="http://localhost:8080"
POSTGRES_USER="postgres"
DB_NAME="trading"

echo "============================================"
echo "STEP 1: Configure PostgreSQL for Disk I/O Testing"
echo "============================================"

# CRITICAL: Reduce PostgreSQL shared_buffers to force disk reads
# Add to postgresql.conf or run:
sudo -u $POSTGRES_USER psql -c "
ALTER SYSTEM SET shared_buffers = '128MB';        -- Small buffer = more disk reads
ALTER SYSTEM SET effective_cache_size = '256MB';  -- Tell planner memory is limited
ALTER SYSTEM SET work_mem = '4MB';                -- Small work memory
ALTER SYSTEM SET random_page_cost = 1.1;          -- Encourage index use
"

echo "Restarting PostgreSQL to apply settings..."
sudo systemctl restart postgresql

sleep 5

echo "============================================"
echo "STEP 2: Generate Large Dataset (This takes time!)"
echo "============================================"

# Generate 50 million trades - adjust based on your RAM
# Goal: Data should be 2-3x your RAM size
sudo -u $POSTGRES_USER psql -d $DB_NAME << 'EOF'

-- Check current size
SELECT pg_size_pretty(pg_total_relation_size('trades')) as current_size;

-- Drop old indexes for faster insert
DROP INDEX IF EXISTS idx_trades_stock_symbol;
DROP INDEX IF EXISTS idx_trades_id;

-- Generate data in batches to avoid OOM
DO $$
DECLARE
    batch_size INTEGER := 1000000;
    total_rows INTEGER := 50000000;
    i INTEGER := 0;
BEGIN
    WHILE i < total_rows LOOP
        INSERT INTO trades (stock_symbol, buyer_id, seller_id, buy_order_id, sell_order_id, price, quantity, executed_at)
        SELECT 
            'STOCK_' || (random() * 10000)::int,
            (random() * 1000)::int + 1,
            (random() * 1000)::int + 1,
            (random() * 10000000)::int,
            (random() * 10000000)::int,
            (random() * 999 + 1)::numeric(10,2),
            (random() * 999 + 1)::int,
            NOW() - (random() * 365)::int * INTERVAL '1 day'
        FROM generate_series(1, batch_size);
        
        i := i + batch_size;
        RAISE NOTICE 'Inserted % rows', i;
        COMMIT;
    END LOOP;
END $$;

-- Recreate indexes
CREATE INDEX CONCURRENTLY idx_trades_id ON trades(id);
CREATE INDEX CONCURRENTLY idx_trades_stock_symbol ON trades(stock_symbol);

-- Get final size
VACUUM ANALYZE trades;
SELECT 
    pg_size_pretty(pg_total_relation_size('trades')) as total_size,
    (SELECT count(*) FROM trades) as row_count;

EOF

echo "============================================"
echo "STEP 3: Clear all caches before testing"
echo "============================================"

# Clear PostgreSQL cache by restarting
sudo systemctl restart postgresql

# Clear Linux page cache (requires root)
sync
echo 3 | sudo tee /proc/sys/vm/drop_caches

echo "============================================"
echo "STEP 4: Verify setup"
echo "============================================"

# Check that data exists and is large
sudo -u $POSTGRES_USER psql -d $DB_NAME -c "
SELECT 
    pg_size_pretty(pg_total_relation_size('trades')) as data_size,
    (SELECT count(*) FROM trades) as total_trades,
    pg_size_pretty(pg_relation_size('idx_trades_id')) as index_size;
"

echo "============================================"
echo "Setup complete! Now run the load test."
echo "============================================"