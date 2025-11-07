#!/bin/bash

echo "Testing cache functionality..."
echo ""

# Clear cache
echo "1. Clearing cache..."
CLEAR_RESULT=$(curl -s -X POST http://localhost:8080/api/cache/clear)
echo "   Response: $CLEAR_RESULT"
sleep 1
echo ""

# First request
echo "2. First request (expect MISS)..."
CACHE1=$(curl -s -v http://localhost:8080/api/trades/AAPL 2>&1 | grep "X-Cache:" | awk '{print $3}' | tr -d '\r\n ')
echo "   Cache status: [$CACHE1]"
echo ""

sleep 0.5

# Second request
echo "3. Second request (expect HIT)..."
CACHE2=$(curl -s -v http://localhost:8080/api/trades/AAPL 2>&1 | grep "X-Cache:" | awk '{print $3}' | tr -d '\r\n ')
echo "   Cache status: [$CACHE2]"
echo ""

# Stats (without JSON parsing)
echo "4. Cache statistics (raw)..."
STATS=$(curl -s http://localhost:8080/api/cache/stats)
echo "$STATS"
echo ""

# Simple verification
if [[ "$CACHE1" == *"MISS"* ]] && [[ "$CACHE2" == *"HIT"* ]]; then
    echo "✅ CACHE IS WORKING!"
elif [[ "$CACHE1" == *"HIT"* ]] && [[ "$CACHE2" == *"HIT"* ]]; then
    echo "⚠️  Both requests hit cache - cache not clearing properly"
else
    echo "❌ Unexpected results: $CACHE1 then $CACHE2"
fi