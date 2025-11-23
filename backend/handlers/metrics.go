package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
)

type MetricsHandler struct {
	pid          int
	lastDiskIo   uint64
	lastDiskTime time.Time
}

func NewMetricsHandler() *MetricsHandler {
	return &MetricsHandler{
		pid:          os.Getpid(),
		lastDiskTime: time.Now(),
	}
}

type MetricsResponse struct {
	CPU     float64 `json:"cpu"`     // Total % CPU across all allowed cores (0-400% for 4 cores)
	Memory  float64 `json:"memory"`  // % memory usage
	Disk    float64 `json:"disk"`    // %util like iostat -xd
	Network float64 `json:"network"` // MB transferred
	Time    string  `json:"time"`
}

func (h *MetricsHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	cpu := readProcessCPU(h.pid)
	memStats, _ := mem.VirtualMemory()
	diskUtil := h.getDiskUtil()
	netStats, _ := net.IOCounters(false)

	network := float64(0)
	if len(netStats) > 0 {
		network = float64(netStats[0].BytesSent+netStats[0].BytesRecv) / (1024 * 1024)
	}

	resp := MetricsResponse{
		CPU:     cpu,
		Memory:  memStats.UsedPercent,
		Disk:    diskUtil,
		Network: network,
		Time:    time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// █ CPU — Total % across all allowed cores (0-400% for 4 cores pinned)
func readProcessCPU(pid int) float64 {
	stat1 := readProcStat(pid)
	time.Sleep(300 * time.Millisecond)
	stat2 := readProcStat(pid)

	if len(stat1) < 17 || len(stat2) < 17 {
		return 0
	}

	ut1, _ := strconv.ParseFloat(stat1[13], 64)
	st1, _ := strconv.ParseFloat(stat1[14], 64)
	ut2, _ := strconv.ParseFloat(stat2[13], 64)
	st2, _ := strconv.ParseFloat(stat2[14], 64)

	delta := (ut2 + st2) - (ut1 + st1)

	// delta is in ticks (USER_HZ, typically 100 ticks/sec)
	// For 0.3 seconds, max ticks = 0.3 * 100 = 30 ticks per core
	// CPU% = (actual_ticks / max_possible_ticks) * 100
	// For multiple cores: actual_ticks can be up to (30 * num_cores)
	clkTck := 100.0
	elapsedSec := 0.3
	maxTicks := clkTck * elapsedSec

	cpuPercent := (delta / maxTicks) * 100

	if cpuPercent < 0 {
		cpuPercent = 0
	}

	return cpuPercent
}

func readProcStat(pid int) []string {
	data, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/stat")
	if err != nil {
		return []string{}
	}
	return strings.Fields(string(data))
}

// █ DISK UTIL — %util like iostat -xd
func (h *MetricsHandler) getDiskUtil() float64 {
	stats, err := disk.IOCounters()
	if err != nil || len(stats) == 0 {
		return 0
	}

	var totalIoTime uint64
	for _, s := range stats {
		totalIoTime += s.IoTime
	}

	now := time.Now()
	if h.lastDiskIo == 0 {
		h.lastDiskIo = totalIoTime
		h.lastDiskTime = now
		return 0
	}

	deltaIo := float64(totalIoTime - h.lastDiskIo)
	deltaT := float64(now.Sub(h.lastDiskTime).Milliseconds())

	h.lastDiskIo = totalIoTime
	h.lastDiskTime = now

	if deltaT <= 0 {
		return 0
	}

	util := (deltaIo / deltaT) * 100
	if util < 0 {
		util = 0
	}
	if util > 100 {
		util = 100
	}

	return util
}
