package metrics

import (
	"encoding/json"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// MetricsCollector tracks various performance metrics
type MetricsCollector struct {
	mutex               sync.RWMutex
	startTime          time.Time            // time when collection started
	endTime            time.Time            // time when collection stopped
	totalTrades        int64                // counter for trades observed
	totalOrders        int64                // counter for orders observed
	totalCancellations int64                // counter for cancellations
	orderLatencies     []time.Duration      // collected latencies for orders
	tradeLatencies     []time.Duration      // collected latencies for trades
	memoryUsage        []MemorySnapshot     // history of memory snapshots
	goroutineCount     []int                 // history of goroutine counts
	timestamps         []time.Time           // timestamps corresponding to snapshots
	
	// Go-specific metrics
	gcStats            []GCSnapshot
	heapAllocations    []uint64
	numCPU             int
	maxGoroutines      int
}

// MemorySnapshot captures memory usage at a point in time
type MemorySnapshot struct {
	Timestamp    time.Time `json:"timestamp"`
	HeapAlloc    uint64    `json:"heap_alloc"`
	HeapSys      uint64    `json:"heap_sys"`
	HeapInuse    uint64    `json:"heap_inuse"`
	StackInuse   uint64    `json:"stack_inuse"`
	NumGoroutine int       `json:"num_goroutine"`
}

// GCSnapshot captures garbage collection statistics
type GCSnapshot struct {
	Timestamp    time.Time `json:"timestamp"`
	NumGC        uint32    `json:"num_gc"`
	PauseTotalNs uint64    `json:"pause_total_ns"`
	LastPauseNs  uint64    `json:"last_pause_ns"`
}

// PerformanceMetrics contains all collected performance data
type PerformanceMetrics struct {
	Duration              time.Duration       `json:"duration"`
	TotalTrades          int64               `json:"total_trades"`
	TotalOrders          int64               `json:"total_orders"`
	TotalCancellations   int64               `json:"total_cancellations"`
	TradesPerSecond      float64             `json:"trades_per_second"`
	OrdersPerSecond      float64             `json:"orders_per_second"`
	AvgOrderLatency      time.Duration       `json:"avg_order_latency"`
	AvgTradeLatency      time.Duration       `json:"avg_trade_latency"`
	P95OrderLatency      time.Duration       `json:"p95_order_latency"`
	P99OrderLatency      time.Duration       `json:"p99_order_latency"`
	PeakMemoryUsage      uint64              `json:"peak_memory_usage"`
	MaxGoroutines        int                 `json:"max_goroutines"`
	NumCPU               int                 `json:"num_cpu"`
	TotalGCPauses        uint64              `json:"total_gc_pauses"`
	MemorySnapshots      []MemorySnapshot    `json:"memory_snapshots"`
	GCSnapshots          []GCSnapshot        `json:"gc_snapshots"`
	OrderLatencyHist     []int64             `json:"order_latency_histogram"`
	TradeLatencyHist     []int64             `json:"trade_latency_histogram"`
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		startTime:       time.Now(),
		orderLatencies:  make([]time.Duration, 0, 10000),
		tradeLatencies:  make([]time.Duration, 0, 10000),
		memoryUsage:     make([]MemorySnapshot, 0, 1000),
		goroutineCount:  make([]int, 0, 1000),
		timestamps:      make([]time.Time, 0, 1000),
		gcStats:         make([]GCSnapshot, 0, 1000),
		heapAllocations: make([]uint64, 0, 1000),
		numCPU:          runtime.NumCPU(),
	}
}

// Start begins metrics collection
func (mc *MetricsCollector) Start() {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	mc.startTime = time.Now()
}

// Stop ends metrics collection
func (mc *MetricsCollector) Stop() {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	mc.endTime = time.Now()
}

// RecordOrder records an order submission with latency
func (mc *MetricsCollector) RecordOrder(latency time.Duration) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	mc.totalOrders++
	mc.orderLatencies = append(mc.orderLatencies, latency)
}

// RecordTrade records a trade execution with latency
func (mc *MetricsCollector) RecordTrade(latency time.Duration) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	mc.totalTrades++
	mc.tradeLatencies = append(mc.tradeLatencies, latency)
}

// RecordCancellation records an order cancellation
func (mc *MetricsCollector) RecordCancellation() {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	mc.totalCancellations++
}

// TakeSnapshot captures current system state
func (mc *MetricsCollector) TakeSnapshot() {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	now := time.Now()
	numGoroutines := runtime.NumGoroutine()
	
	// Memory snapshot
	memSnapshot := MemorySnapshot{
		Timestamp:    now,
		HeapAlloc:    memStats.HeapAlloc,
		HeapSys:      memStats.HeapSys,
		HeapInuse:    memStats.HeapInuse,
		StackInuse:   memStats.StackInuse,
		NumGoroutine: numGoroutines,
	}
	mc.memoryUsage = append(mc.memoryUsage, memSnapshot)
	
	// GC snapshot
	gcSnapshot := GCSnapshot{
		Timestamp:    now,
		NumGC:        memStats.NumGC,
		PauseTotalNs: memStats.PauseTotalNs,
		// LastPauseNs: get last pause safely using ring buffer index
		LastPauseNs:  memStats.PauseNs[(memStats.NumGC+255)%256],
	}
	mc.gcStats = append(mc.gcStats, gcSnapshot)
	
	// Track maximums
	if numGoroutines > mc.maxGoroutines {
		mc.maxGoroutines = numGoroutines
	}
	
	mc.goroutineCount = append(mc.goroutineCount, numGoroutines)
	mc.timestamps = append(mc.timestamps, now)
	mc.heapAllocations = append(mc.heapAllocations, memStats.HeapAlloc)
}

// GetMetrics returns comprehensive performance metrics
func (mc *MetricsCollector) GetMetrics() PerformanceMetrics {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()
	
	duration := mc.endTime.Sub(mc.startTime)
	if duration == 0 {
		duration = time.Since(mc.startTime)
	}
	
	var avgOrderLatency, avgTradeLatency time.Duration
	var p95OrderLatency, p99OrderLatency time.Duration
	var peakMemory uint64
	var totalGCPauses uint64
	
	// Calculate order latency statistics
	if len(mc.orderLatencies) > 0 {
		total := time.Duration(0)
		for _, lat := range mc.orderLatencies {
			total += lat
		}
		avgOrderLatency = total / time.Duration(len(mc.orderLatencies))
		
		// Calculate percentiles
		sortedLatencies := make([]time.Duration, len(mc.orderLatencies))
		copy(sortedLatencies, mc.orderLatencies)
		for i := 0; i < len(sortedLatencies)-1; i++ {
			for j := i + 1; j < len(sortedLatencies); j++ {
				if sortedLatencies[i] > sortedLatencies[j] {
					sortedLatencies[i], sortedLatencies[j] = sortedLatencies[j], sortedLatencies[i]
				}
			}
		}
		
		p95Index := int(float64(len(sortedLatencies)) * 0.95)
		p99Index := int(float64(len(sortedLatencies)) * 0.99)
		if p95Index < len(sortedLatencies) {
			p95OrderLatency = sortedLatencies[p95Index]
		}
		if p99Index < len(sortedLatencies) {
			p99OrderLatency = sortedLatencies[p99Index]
		}
	}
	
	// Calculate trade latency statistics
	if len(mc.tradeLatencies) > 0 {
		total := time.Duration(0)
		for _, lat := range mc.tradeLatencies {
			total += lat
		}
		avgTradeLatency = total / time.Duration(len(mc.tradeLatencies))
	}
	
	// Find peak memory usage
	for _, snapshot := range mc.memoryUsage {
		if snapshot.HeapAlloc > peakMemory {
			peakMemory = snapshot.HeapAlloc
		}
	}
	
	// Calculate total GC pauses
	if len(mc.gcStats) > 0 {
		totalGCPauses = mc.gcStats[len(mc.gcStats)-1].PauseTotalNs
	}
	
	// Create histograms
	orderLatencyHist := mc.createLatencyHistogram(mc.orderLatencies)
	tradeLatencyHist := mc.createLatencyHistogram(mc.tradeLatencies)
	
	return PerformanceMetrics{
		Duration:             duration,
		TotalTrades:          mc.totalTrades,
		TotalOrders:          mc.totalOrders,
		TotalCancellations:   mc.totalCancellations,
		TradesPerSecond:      float64(mc.totalTrades) / duration.Seconds(),
		OrdersPerSecond:      float64(mc.totalOrders) / duration.Seconds(),
		AvgOrderLatency:      avgOrderLatency,
		AvgTradeLatency:      avgTradeLatency,
		P95OrderLatency:      p95OrderLatency,
		P99OrderLatency:      p99OrderLatency,
		PeakMemoryUsage:      peakMemory,
		MaxGoroutines:        mc.maxGoroutines,
		NumCPU:               mc.numCPU,
		TotalGCPauses:        totalGCPauses,
		MemorySnapshots:      mc.memoryUsage,
		GCSnapshots:          mc.gcStats,
		OrderLatencyHist:     orderLatencyHist,
		TradeLatencyHist:     tradeLatencyHist,
	}
}

// createLatencyHistogram creates a histogram of latencies
func (mc *MetricsCollector) createLatencyHistogram(latencies []time.Duration) []int64 {
	if len(latencies) == 0 {
		return []int64{}
	}
	
	// Create 20 buckets
	buckets := make([]int64, 20)
	
	// Find max latency
	maxLatency := time.Duration(0)
	for _, lat := range latencies {
		if lat > maxLatency {
			maxLatency = lat
		}
	}
	
	if maxLatency == 0 {
		return buckets
	}
	
	bucketSize := maxLatency / time.Duration(len(buckets))
	
	for _, lat := range latencies {
		bucketIndex := int(lat / bucketSize)
		if bucketIndex >= len(buckets) {
			bucketIndex = len(buckets) - 1
		}
		buckets[bucketIndex]++
	}
	
	return buckets
}

// ExportToJSON exports metrics to JSON format
func (mc *MetricsCollector) ExportToJSON() ([]byte, error) {
	metrics := mc.GetMetrics()
	return json.MarshalIndent(metrics, "", "  ")
}

// PrintSummary prints a summary of the metrics
func (mc *MetricsCollector) PrintSummary() {
	metrics := mc.GetMetrics()
	
	fmt.Println("\n========== PERFORMANCE METRICS SUMMARY ==========")
	fmt.Printf("Duration: %v\n", metrics.Duration)
	fmt.Printf("Total Trades: %d\n", metrics.TotalTrades)
	fmt.Printf("Total Orders: %d\n", metrics.TotalOrders)
	fmt.Printf("Total Cancellations: %d\n", metrics.TotalCancellations)
	fmt.Printf("Trades/Second: %.2f\n", metrics.TradesPerSecond)
	fmt.Printf("Orders/Second: %.2f\n", metrics.OrdersPerSecond)
	fmt.Printf("Average Order Latency: %v\n", metrics.AvgOrderLatency)
	fmt.Printf("Average Trade Latency: %v\n", metrics.AvgTradeLatency)
	fmt.Printf("P95 Order Latency: %v\n", metrics.P95OrderLatency)
	fmt.Printf("P99 Order Latency: %v\n", metrics.P99OrderLatency)
	fmt.Printf("Peak Memory Usage: %.2f MB\n", float64(metrics.PeakMemoryUsage)/1024/1024)
	fmt.Printf("Max Goroutines: %d\n", metrics.MaxGoroutines)
	fmt.Printf("Number of CPUs: %d\n", metrics.NumCPU)
	fmt.Printf("Total GC Pauses: %.2f ms\n", float64(metrics.TotalGCPauses)/1e6)
	fmt.Println("=================================================")
}