#!/bin/bash

SERVER_URL="http://localhost:8080"
TEST_DURATION=60
LOADGEN="./loadgen/loadgen"
OUTPUT_DIR="./test_results"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
RESULT_DIR="${OUTPUT_DIR}/IO_${TIMESTAMP}"

mkdir -p "$RESULT_DIR"
echo "Results stored in: $RESULT_DIR"

# Health check
if ! curl -s --max-time 5 "$SERVER_URL/api/health" >/dev/null; then
    echo " Server not running"
    exit 1
fi

run_test() {
    clients=$1
    outfile="${RESULT_DIR}/disk_c${clients}.txt"
    echo "▶ Running IO BOUND with $clients clients..."
    taskset -c 0-6 $LOADGEN \
        -url "$SERVER_URL" \
        -clients "$clients" \
        -duration "$TEST_DURATION" \
        -workload "disk-write" \
        2>&1 | tee "$outfile"
    sleep 5
}

# Test configurations
for clients in 1 5 10 15 20 25 30 50 75; do
    run_test $clients
done

# Create metrics.csv
SUMMARY="${RESULT_DIR}/metrics.csv"
echo "timestamp,workload,clients,throughput,avg_latency,median_latency,cpu,memory,disk,network,success,failure" > "$SUMMARY"

# Parse only the files generated now
for f in "$RESULT_DIR"/disk_c*.txt; do
    csv=$(grep "^CSV" "$f" | tail -1)
    clients=$(basename "$f" | grep -oE '[0-9]+')

    IFS=',' read _ workload clients_ throughput avg med cpu mem disk net success fail <<< "$csv"
    echo "$TIMESTAMP,io-bound,$clients,$throughput,$avg,$med,$cpu,$mem,$disk,$net,$success,$fail" >> "$SUMMARY"
done

echo "✔ IO Bound results saved: $SUMMARY"
