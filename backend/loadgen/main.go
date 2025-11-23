package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type LoadGenerator struct {
	serverURL    string
	numClients   int
	duration     time.Duration
	workloadType string

	successCount int64
	failureCount int64

	latencies []int64
	metrics   []MetricsSample

	mu sync.Mutex
}

type MetricsSample struct {
	CPU     float64 `json:"cpu"`
	Memory  float64 `json:"memory"`
	Disk    float64 `json:"disk"`
	Network float64 `json:"network"`
}

var hotStocks = []string{"AAPL", "GOOGL", "MSFT", "AMZN", "TSLA"}

func main() {
	url := flag.String("url", "http://localhost:8080", "Server URL")
	clients := flag.Int("clients", 10, "Concurrent clients")
	duration := flag.Int("duration", 60, "Duration seconds")
	workload := flag.String("workload", "cpu-bound", "cpu-bound | io-bound | disk-write")
	flag.Parse()

	lg := &LoadGenerator{
		serverURL:    *url,
		numClients:   *clients,
		duration:     time.Duration(*duration) * time.Second,
		workloadType: *workload,
		latencies:    make([]int64, 0, 500000),
		metrics:      make([]MetricsSample, 0, 5000),
	}
	lg.Run()
}

func (lg *LoadGenerator) Run() {
	stop := time.Now().Add(lg.duration)
	var wg sync.WaitGroup

	go lg.collectMetrics(stop)

	for i := 0; i < lg.numClients; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			c := &http.Client{Timeout: 20 * time.Second}
			r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(id)))
			for time.Now().Before(stop) {
				var latency int64
				var err error

				switch lg.workloadType {
				case "cpu-bound":
					latency, err = lg.doGET(c, fmt.Sprintf("%s/api/trades/%s", lg.serverURL, hotStocks[r.Intn(len(hotStocks))]))
				case "io-bound":
					latency, err = lg.doGET(c, fmt.Sprintf("%s/api/trades/COLD_%d", lg.serverURL, r.Intn(9999999)))
				case "disk-write":
					body := map[string]any{
						"stock":    fmt.Sprintf("S_%d", r.Intn(2000)),
						"buyer_id": r.Intn(9999),
						"price":    float64(r.Intn(150) + 50),
					}
					latency, err = lg.doPOST(c, lg.serverURL+"/api/disk/trade", body)
				}

				lg.mu.Lock()
				lg.latencies = append(lg.latencies, latency)
				lg.mu.Unlock()

				if err != nil {
					atomic.AddInt64(&lg.failureCount, 1)
				} else {
					atomic.AddInt64(&lg.successCount, 1)
				}
			}
		}(i)
	}

	wg.Wait()
	lg.printCSV()
}

func (lg *LoadGenerator) doGET(c *http.Client, url string) (int64, error) {
	start := time.Now()
	resp, err := c.Get(url)
	if err != nil {
		return time.Since(start).Milliseconds(), err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	return time.Since(start).Milliseconds(), nil
}

func (lg *LoadGenerator) doPOST(c *http.Client, url string, body interface{}) (int64, error) {
	start := time.Now()
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.Do(req)
	if err != nil {
		return time.Since(start).Milliseconds(), err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	return time.Since(start).Milliseconds(), nil
}

func (lg *LoadGenerator) collectMetrics(stop time.Time) {
	c := &http.Client{Timeout: 2 * time.Second}
	for time.Now().Before(stop) {
		resp, err := c.Get(lg.serverURL + "/api/metrics")
		if err == nil {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			var m MetricsSample
			json.Unmarshal(body, &m)
			lg.mu.Lock()
			lg.metrics = append(lg.metrics, m)
			lg.mu.Unlock()
		}
		time.Sleep(1 * time.Second)
	}
}

func (lg *LoadGenerator) printCSV() {
	if len(lg.latencies) == 0 {
		fmt.Println("NO_DATA")
		return
	}

	sort.Slice(lg.latencies, func(i, j int) bool { return lg.latencies[i] < lg.latencies[j] })
	median := float64(lg.latencies[len(lg.latencies)/2])

	var sum int64
	for _, x := range lg.latencies {
		sum += x
	}
	avg := float64(sum) / float64(len(lg.latencies))
	throughput := float64(lg.successCount) / lg.duration.Seconds()

	var cpu, mem, disk, net float64
	for _, m := range lg.metrics {
		cpu += m.CPU
		mem += m.Memory
		disk += m.Disk
		net += m.Network
	}
	n := float64(len(lg.metrics))
	cpu /= n
	mem /= n
	disk /= n
	net /= n

	fmt.Printf(
		"CSV,%s,%d,%.4f,%.4f,%.4f,%.2f,%.2f,%.2f,%.2f,%d,%d\n",
		lg.workloadType,
		lg.numClients,
		throughput,
		avg,
		median,
		cpu,
		mem,
		disk,
		net,
		lg.successCount,
		lg.failureCount,
	)
}
