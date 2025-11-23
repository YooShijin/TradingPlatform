#!/bin/bash

SERVER_URL="http://localhost:8080"
LOADGEN="./loadgen/loadgen"
OUTPUT_DIR="./test_results"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

CPU_DURATION=60
IO_DURATION=60

CPU_CLIENTS=(1 3 5 10 15  25  50 75 100 150 200 250 300  400)
IO_CLIENTS=(1 3 5 10 15  25  35 50 75 100 150 200 250 300  400)

set -e

echo "=============================="
echo "⚙ Running Benchmarks"
echo "=============================="

# Health check
if ! curl -s --max-time 5 "$SERVER_URL/api/health" >/dev/null; then
    echo " Server unreachable at $SERVER_URL"
    exit 1
fi
echo "✓ Server is running"

########################################
# CPU-BOUND TEST
########################################
CPU_RESULT="${OUTPUT_DIR}/CPU_${TIMESTAMP}"
mkdir -p "$CPU_RESULT"
echo " CPU Results → $CPU_RESULT"

run_cpu_test() {
    local clients=$1
    local outfile="${CPU_RESULT}/cpu_c${clients}.txt"
    echo ""
    echo "▶ CPU workload | $clients clients | ${CPU_DURATION}s"
    taskset -c 0-6 $LOADGEN \
        -url "$SERVER_URL" \
        -clients "$clients" \
        -duration "$CPU_DURATION" \
        -workload "cpu-bound" 2>&1 | tee "$outfile"
    sleep 10
}

for c in "${CPU_CLIENTS[@]}"; do
    run_cpu_test "$c"
done

CPU_SUMMARY="${CPU_RESULT}/metrics.csv"
echo "timestamp,workload,clients,throughput,mean_latency_ms,median_latency_ms,success,failure" > "$CPU_SUMMARY"

for f in "$CPU_RESULT"/*.txt; do
    csv_line=$(grep "^CSV" "$f" | tail -1)
    clients=$(basename "$f" | sed -E 's/.*_c([0-9]+)\.txt/\1/')
    echo "$TIMESTAMP,cpu-bound,$clients,$(echo "$csv_line" | cut -d',' -f4-8)" >> "$CPU_SUMMARY"
done

echo "✔ CPU metrics saved → $CPU_SUMMARY"
echo ""
echo "--------------------------------"
echo " Starting DISK I/O Tests"
echo "--------------------------------"


# DISK I/O TEST


IO_RESULT="${OUTPUT_DIR}/IO_${TIMESTAMP}"
mkdir -p "$IO_RESULT"
echo "I/O Results → $IO_RESULT"

run_io_test() {
    local clients=$1
    local outfile="${IO_RESULT}/io_c${clients}.txt"
    echo ""
    echo "▶ DISK I/O workload | $clients clients | ${IO_DURATION}s"
    taskset -c 0-6 $LOADGEN \
        -url "$SERVER_URL" \
        -clients "$clients" \
        -duration "$IO_DURATION" \
        -workload "disk-write" 2>&1 | tee "$outfile"
    sleep 10
}

for c in "${IO_CLIENTS[@]}"; do
    run_io_test "$c"
done

IO_SUMMARY="${IO_RESULT}/metrics.csv"
echo "timestamp,workload,clients,throughput,mean_latency_ms,median_latency_ms,success,failure" > "$IO_SUMMARY"

for f in "$IO_RESULT"/*.txt; do
    csv_line=$(grep "^CSV" "$f" | tail -1)
    clients=$(basename "$f" | sed -E 's/.*_c([0-9]+)\.txt/\1/')
    echo "$TIMESTAMP,io-bound,$clients,$(echo "$csv_line" | cut -d',' -f4-8)" >> "$IO_SUMMARY"
done

echo "✔ DISK metrics saved → $IO_SUMMARY"
echo ""
echo " Benchmark Complete"
echo "CPU: $CPU_RESULT"
echo "IO : $IO_RESULT"
echo "=============================="
