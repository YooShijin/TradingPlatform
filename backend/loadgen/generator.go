package loadgen

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type LoadGenerator struct {
	serverURL    string
	numClients   int
	duration     time.Duration
	workloadType string

	totalRequests int64
	successCount  int64
	failureCount  int64
	totalLatency  int64
}

func NewLoadGenerator(url string, clients int, duration time.Duration, workload string) *LoadGenerator {
	return &LoadGenerator{
		serverURL:    url,
		numClients:   clients,
		duration:     duration,
		workloadType: workload,
	}
}

func (lg *LoadGenerator) Run() {
	var wg sync.WaitGroup
	startTime := time.Now()
	stopTime := startTime.Add(lg.duration)

	fmt.Printf("Starting load test with %d clients for %v\n", lg.numClients, lg.duration)
	fmt.Printf("Workload: %s\n", lg.workloadType)
	fmt.Printf("Target: %s\n\n", lg.serverURL)

	for i := 0; i < lg.numClients; i++ {
		wg.Add(1)
		go lg.clientWorker(i, stopTime, &wg)
	}

	wg.Wait()
	lg.printResults()
}

func (lg *LoadGenerator) clientWorker(clientID int, stopTime time.Time, wg *sync.WaitGroup) {
	defer wg.Done()

	client := &http.Client{Timeout: 30 * time.Second}

	for time.Now().Before(stopTime) {
		url := lg.generateRequest(clientID)

		reqStart := time.Now()
		resp, err := client.Get(url)
		latency := time.Since(reqStart).Milliseconds()

		atomic.AddInt64(&lg.totalRequests, 1)
		atomic.AddInt64(&lg.totalLatency, latency)

		if err != nil || resp.StatusCode != 200 {
			atomic.AddInt64(&lg.failureCount, 1)
			if resp != nil {
				resp.Body.Close()
			}
			continue
		}

		resp.Body.Close()
		atomic.AddInt64(&lg.successCount, 1)

		// Zero think time (continuous load)
	}
}

func (lg *LoadGenerator) generateRequest(clientID int) string {
	switch lg.workloadType {
	case "io-bound":
		// Generate unique stock symbols → Always cache miss
		randomStock := fmt.Sprintf("STOCK_%d_%d", clientID, rand.Intn(1000000))
		return fmt.Sprintf("%s/api/trades/%s", lg.serverURL, randomStock)

	case "cpu-bound":
		// Generate from small set of stocks → Always cache hit
		popularStocks := []string{"AAPL", "GOOGL", "MSFT", "AMZN", "TSLA",
			"META", "NVDA", "NFLX", "BABA", "ORCL"}
		stock := popularStocks[rand.Intn(len(popularStocks))]
		return fmt.Sprintf("%s/api/trades/%s", lg.serverURL, stock)

	default:
		return fmt.Sprintf("%s/api/trades/AAPL", lg.serverURL)
	}
}

func (lg *LoadGenerator) printResults() {
	elapsed := lg.duration.Seconds()

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("LOAD TEST RESULTS")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Workload Type:       %s\n", lg.workloadType)
	fmt.Printf("Number of Clients:   %d\n", lg.numClients)
	fmt.Printf("Test Duration:       %.2f seconds\n", elapsed)
	fmt.Printf("Total Requests:      %d\n", lg.totalRequests)
	fmt.Printf("Successful:          %d\n", lg.successCount)
	fmt.Printf("Failed:              %d\n", lg.failureCount)
	fmt.Printf("\nAverage Throughput:  %.2f requests/sec\n",
		float64(lg.successCount)/elapsed)
	fmt.Printf("Average Latency:     %.2f ms\n",
		float64(lg.totalLatency)/float64(lg.totalRequests))
	fmt.Println(strings.Repeat("=", 60))
}

func main() {
	serverURL := flag.String("url", "http://localhost:8080", "Server URL")
	numClients := flag.Int("clients", 10, "Number of concurrent clients")
	duration := flag.Int("duration", 300, "Test duration in seconds")
	workload := flag.String("workload", "io-bound", "Workload type: io-bound or cpu-bound")

	flag.Parse()

	lg := NewLoadGenerator(*serverURL, *numClients, time.Duration(*duration)*time.Second, *workload)
	lg.Run()
}
