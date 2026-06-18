package main

import (
	"encoding/json"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"
)

type StatusResponse struct {
	Status    string `json:"status"`
	ServedBy  string `json:"servedBy"`
	Timestamp string `json:"timestamp"`
}

type MetricsResponse struct {
	ServedBy      string  `json:"servedBy"`
	RequestCount  int64   `json:"requestCount"`
	UptimeSeconds int64   `json:"uptimeSeconds"`
	MemoryUsageMB float64 `json:"memoryUsageMB"`
}

var (
	requestCount int64
	startTime    = time.Now()
)

func hostname() string {
	h, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return h
}

func countRequest(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&requestCount, 1)
		next(w, r)
	}
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	resp := StatusResponse{
		Status:    "ok",
		ServedBy:  hostname(),
		Timestamp: time.Now().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(resp)
}

func heavyHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	loadStr := r.URL.Query().Get("load")
	load, err := strconv.Atoi(loadStr)
	if err != nil || load <= 0 {
		load = 10000000
	}

	var result uint64

	for i := 0; i < load; i++ {
		result += uint64(i * i)
	}

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	resp := map[string]interface{}{
		"servedBy":      hostname(),
		"durationMs":    time.Since(start).Milliseconds(),
		"memoryUsageMB": float64(mem.HeapAlloc) / 1024 / 1024,
		"result":        result,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	resp := MetricsResponse{
		ServedBy:      hostname(),
		RequestCount:  atomic.LoadInt64(&requestCount),
		UptimeSeconds: int64(time.Since(startTime).Seconds()),
		MemoryUsageMB: float64(mem.HeapAlloc) / 1024 / 1024,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	http.HandleFunc("/api/status", countRequest(statusHandler))
	http.HandleFunc("/api/metrics", countRequest(metricsHandler))
	http.HandleFunc("/api/heavy", countRequest(heavyHandler))
	http.ListenAndServe(":3000", nil)
}
