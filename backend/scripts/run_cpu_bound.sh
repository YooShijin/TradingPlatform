#!/bin/bash

SERVER_URL="http://localhost:8080"
TEST_DURATION=60
LOADGEN="./loadgen/loadgen"
OUTPUT_DIR="./test_results"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
RESULT_DIR="${OUTPUT_DIR}/CPU_${TIMESTAMP}"

mkdir -p "$RESULT_DIR"
echo "Results stored in: $RESULT_DIR"

# Health check
if ! curl -s --max-time 5 "$SERVER_URL/api/health" >/dev/null; then
    echo " Server not running"
    exit 1
fi

run_test() {
    clients=$1
    outfile="${RESULT_DIR}/cpu_c${clients}.txt"
    echo "▶ Running CPU load test with $clients clients..."
    taskset -c 0-7 $LOADGEN -url "$SERVER_URL" -clients "$clients" -duration "$TEST_DURATION" -workload "cpu-bound" \
        2>&1 | tee "$outfile"
    sleep 5
}

# Test matrix
for clients in 1 3 5 10 25 50 75 100 150; do
    run_test $clients
done

# Create metrics.csv
SUMMARY="${RESULT_DIR}/metrics.csv"
echo "timestamp,workload,clients,throughput,avg_latency,median_latency,cpu,memory,disk,network,success,failure" > "$SUMMARY"

# Parse results
for f in "$RESULT_DIR"/cpu_c*.txt; do
    csv=$(grep "^CSV" "$f" | tail -1)
    clients=$(basename "$f" | grep -oE '[0-9]+')

    IFS=',' read _ workload clients_ throughput avg med cpu mem disk net success fail <<< "$csv"

    echo "$TIMESTAMP,cpu-bound,$clients,$throughput,$avg,$med,$cpu,$mem,$disk,$net,$success,$fail" >> "$SUMMARY"
done

echo "✔ CPU results saved: $SUMMARY"
