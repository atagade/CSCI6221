package evaluation

import (
	"testing"
	"time"
)

func TestConcurrentBenchmark(t *testing.T) {
	result := RunConcurrentBenchmark(5, 2*time.Second)
	
	if result.TotalTrades < 0 {
		t.Errorf("Expected non-negative trades, got %d", result.TotalTrades)
	}
	
	if result.TradesPerSecond < 0 {
		t.Errorf("Expected non-negative trades per second, got %.2f", result.TradesPerSecond)
	}
	
	if result.PeakMemoryMB <= 0 {
		t.Errorf("Expected positive memory usage, got %.2f MB", result.PeakMemoryMB)
	}
	
	t.Logf("Concurrent benchmark completed: %.2f trades/sec, %.2f MB peak memory", 
		result.TradesPerSecond, result.PeakMemoryMB)
}

func TestSequentialBenchmark(t *testing.T) {
	result := RunSequentialBenchmark(5, 2*time.Second)
	
	if result.TotalTrades < 0 {
		t.Errorf("Expected non-negative trades, got %d", result.TotalTrades)
	}
	
	if result.TradesPerSecond < 0 {
		t.Errorf("Expected non-negative trades per second, got %.2f", result.TradesPerSecond)
	}
	
	if result.PeakMemoryMB <= 0 {
		t.Errorf("Expected positive memory usage, got %.2f MB", result.PeakMemoryMB)
	}
	
	t.Logf("Sequential benchmark completed: %.2f trades/sec, %.2f MB peak memory", 
		result.TradesPerSecond, result.PeakMemoryMB)
}

func TestConcurrencyComparison(t *testing.T) {
	comparison := RunConcurrencyComparison(3, 1*time.Second)
	
	if comparison.SpeedupRatio <= 0 {
		t.Errorf("Expected positive speedup ratio, got %.2f", comparison.SpeedupRatio)
	}
	
	if comparison.ConcurrentResult.TotalTrades < 0 {
		t.Errorf("Expected non-negative concurrent trades, got %d", comparison.ConcurrentResult.TotalTrades)
	}
	
	if comparison.SequentialResult.TotalTrades < 0 {
		t.Errorf("Expected non-negative sequential trades, got %d", comparison.SequentialResult.TotalTrades)
	}
	
	PrintComparisonReport(comparison)
	
	t.Logf("Concurrency comparison completed: %.2fx speedup, %.2f%% efficiency gain", 
		comparison.SpeedupRatio, comparison.EfficiencyGain)
}

func TestScalabilityTest(t *testing.T) {
	agentCounts := []int{2, 4, 6}
	results := RunScalabilityTest(agentCounts, 1*time.Second)
	
	if len(results) != len(agentCounts) {
		t.Errorf("Expected %d results, got %d", len(agentCounts), len(results))
	}
	
	for i, result := range results {
		// Ensure sensible values (non-negative); detailed correctness is validated elsewhere
		if result.TotalTrades < 0 {
			t.Errorf("Test %d: Expected non-negative trades, got %d", i, result.TotalTrades)
		}
        
		// Log a concise summary for maintainers to inspect test outputs
		t.Logf("Scalability test %d agents: %.2f trades/sec, %.2f MB peak memory", 
			agentCounts[i], result.TradesPerSecond, result.PeakMemoryMB)
	}
}

// Benchmark functions for Go's testing framework

func BenchmarkConcurrent10Agents(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RunConcurrentBenchmark(10, 500*time.Millisecond)
	}
}

func BenchmarkConcurrent20Agents(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RunConcurrentBenchmark(20, 500*time.Millisecond)
	}
}

func BenchmarkSequential10Agents(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RunSequentialBenchmark(10, 500*time.Millisecond)
	}
}

func BenchmarkSequential20Agents(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RunSequentialBenchmark(20, 500*time.Millisecond)
	}
}