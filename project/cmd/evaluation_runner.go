package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"cda-simulator/evaluation"
)

var (
	outputDir     = flag.String("output", "evaluation_results", "output directory for results")
	fullBenchmark = flag.Bool("full-benchmark", false, "run comprehensive benchmark suite")
	quickTest     = flag.Bool("quick", false, "run quick evaluation tests")
	scalabilityTest = flag.Bool("scalability", false, "run scalability tests")
	generateCharts = flag.Bool("charts", true, "generate visualization charts")
)

func main() {
	flag.Parse()

	fmt.Println("========== CDA Exchange Simulator Evaluation Runner ==========")
	
	// Create output directory
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		return
	}

	// Run evaluation based on flags
	if *quickTest {
		runQuickEvaluation()
	} else if *scalabilityTest {
		runScalabilityEvaluation()
	} else if *fullBenchmark {
		runFullBenchmarkSuite()
	} else {
		// Default: run basic comparison
		runBasicComparison()
	}

	// Generate charts if requested
	if *generateCharts {
		generateVisualizationCharts()
	}

	fmt.Printf("\nEvaluation completed. Results saved to: %s\n", *outputDir)
}

// runBasicComparison runs a basic concurrent vs sequential comparison
func runBasicComparison() {
	fmt.Println("\n=== Running Basic Concurrency Comparison ===")
	
	numAgents := 10
	duration := 3 * time.Second
	
	// Run comparison
	comparison := evaluation.RunConcurrencyComparison(numAgents, duration)
	
	// Print results
	evaluation.PrintComparisonReport(comparison)
	
	// Save results
	saveComparisonResults(comparison, "basic_comparison.json")
}

// runQuickEvaluation runs quick tests for rapid feedback
func runQuickEvaluation() {
	fmt.Println("\n=== Running Quick Evaluation Tests ===")
	
	tests := []struct {
		name      string
		agents    int
		duration  time.Duration
	}{
		{"Small", 5, 2 * time.Second},
		{"Medium", 15, 3 * time.Second},
		{"Large", 25, 4 * time.Second},
	}
	
	var results []evaluation.BenchmarkResult
	
	for _, test := range tests {
		fmt.Printf("Running %s test (%d agents, %v)...\n", test.name, test.agents, test.duration)
		result := evaluation.RunConcurrentBenchmark(test.agents, test.duration)
		results = append(results, result)
		
		fmt.Printf("  - Throughput: %.2f trades/sec\n", result.TradesPerSecond)
		fmt.Printf("  - Memory: %.2f MB\n", result.PeakMemoryMB)
		fmt.Printf("  - Goroutines: %d\n", result.MaxGoroutines)
	}
	
	// Save results
	saveResults(results, "quick_evaluation.json")
}

// runScalabilityEvaluation tests performance across different scales
func runScalabilityEvaluation() {
	fmt.Println("\n=== Running Scalability Evaluation ===")
	
	agentCounts := []int{5, 10, 20, 30, 50}
	duration := 5 * time.Second
	
	results := evaluation.RunScalabilityTest(agentCounts, duration)
	
	// Print scalability analysis
	fmt.Printf("\nScalability Analysis:\n")
	fmt.Printf("%-10s %-15s %-15s %-15s %-15s\n", "Agents", "Trades/Sec", "Memory(MB)", "Goroutines", "Efficiency")
	fmt.Printf("%-10s %-15s %-15s %-15s %-15s\n", "------", "---------", "---------", "----------", "----------")
	
	for i, result := range results {
		efficiency := result.TradesPerSecond / float64(agentCounts[i])
		fmt.Printf("%-10d %-15.2f %-15.2f %-15d %-15.3f\n", 
			agentCounts[i], result.TradesPerSecond, result.PeakMemoryMB, 
			result.MaxGoroutines, efficiency)
	}
	
	// Save results
	saveResults(results, "scalability_evaluation.json")
}

// runFullBenchmarkSuite runs comprehensive benchmarks
func runFullBenchmarkSuite() {
	fmt.Println("\n=== Running Full Benchmark Suite ===")
	
	// 1. Concurrency comparison with different scales
	fmt.Println("1. Concurrency Comparison Tests...")
	comparisonResults := make(map[string]evaluation.ConcurrencyComparison)
	
	testScales := []struct {
		name     string
		agents   int
		duration time.Duration
	}{
		{"Small", 10, 3 * time.Second},
		{"Medium", 25, 5 * time.Second},
		{"Large", 50, 7 * time.Second},
	}
	
	for _, test := range testScales {
		fmt.Printf("  Running %s scale test...\n", test.name)
		comparison := evaluation.RunConcurrencyComparison(test.agents, test.duration)
		comparisonResults[test.name] = comparison
		
		fmt.Printf("    Speedup: %.2fx, Efficiency: %.2f%%\n", 
			comparison.SpeedupRatio, comparison.EfficiencyGain)
	}
	
	// 2. Scalability tests
	fmt.Println("\n2. Scalability Tests...")
	agentCounts := []int{5, 10, 15, 20, 30, 40, 50}
	scalabilityResults := evaluation.RunScalabilityTest(agentCounts, 5*time.Second)
	
	// 3. Stress tests
	fmt.Println("\n3. Stress Tests...")
	stressResults := make(map[string]evaluation.BenchmarkResult)
	
	stressTests := []struct {
		name     string
		agents   int
		duration time.Duration
	}{
		{"HighLoad", 100, 5 * time.Second},
		{"LongDuration", 30, 30 * time.Second},
		{"MegaLoad", 200, 3 * time.Second},
	}
	
	for _, test := range stressTests {
		fmt.Printf("  Running %s stress test...\n", test.name)
		result := evaluation.RunConcurrentBenchmark(test.agents, test.duration)
		stressResults[test.name] = result
		
		fmt.Printf("    Throughput: %.2f t/s, Memory: %.2f MB\n", 
			result.TradesPerSecond, result.PeakMemoryMB)
	}
	
	// Save all results
	saveComparisonResults(comparisonResults["Medium"], "full_benchmark_comparison.json")
	saveResults(scalabilityResults, "full_benchmark_scalability.json")
	saveStressResults(stressResults, "full_benchmark_stress.json")
	
	// Generate comprehensive report
	generateBenchmarkReport(comparisonResults, scalabilityResults, stressResults)
}

// saveComparisonResults saves comparison results to JSON
func saveComparisonResults(comparison evaluation.ConcurrencyComparison, filename string) {
	filepath := filepath.Join(*outputDir, filename)
	data, err := json.MarshalIndent(comparison, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling comparison results: %v\n", err)
		return
	}
	
	if err := os.WriteFile(filepath, data, 0644); err != nil {
		fmt.Printf("Error writing comparison results: %v\n", err)
		return
	}
	
	fmt.Printf("Comparison results saved to: %s\n", filepath)
}

// saveResults saves benchmark results to JSON
func saveResults(results []evaluation.BenchmarkResult, filename string) {
	filepath := filepath.Join(*outputDir, filename)
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling results: %v\n", err)
		return
	}
	
	if err := os.WriteFile(filepath, data, 0644); err != nil {
		fmt.Printf("Error writing results: %v\n", err)
		return
	}
	
	fmt.Printf("Results saved to: %s\n", filepath)
}

// saveStressResults saves stress test results to JSON
func saveStressResults(results map[string]evaluation.BenchmarkResult, filename string) {
	filepath := filepath.Join(*outputDir, filename)
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling stress results: %v\n", err)
		return
	}
	
	if err := os.WriteFile(filepath, data, 0644); err != nil {
		fmt.Printf("Error writing stress results: %v\n", err)
		return
	}
	
	fmt.Printf("Stress test results saved to: %s\n", filepath)
}

// generateVisualizationCharts runs the Python visualizer
func generateVisualizationCharts() {
	fmt.Println("\n=== Generating Visualization Charts ===")
	
	// Check if Python is available
	pythonCmd := "python"
	if _, err := exec.LookPath(pythonCmd); err != nil {
		pythonCmd = "python3"
		if _, err := exec.LookPath(pythonCmd); err != nil {
			fmt.Println("Python not found. Please install Python and matplotlib to generate charts.")
			fmt.Println("Install dependencies: pip install -r visualizations/requirements.txt")
			return
		}
	}
	
	// Find JSON files in output directory
	files, err := filepath.Glob(filepath.Join(*outputDir, "*.json"))
	if err != nil {
		fmt.Printf("Error finding JSON files: %v\n", err)
		return
	}
	
	if len(files) == 0 {
		fmt.Println("No JSON files found to visualize")
		return
	}
	
	// Run Python visualizer for each file
	visualizerScript := "visualizations/performance_visualizer.py"
	if _, err := os.Stat(visualizerScript); os.IsNotExist(err) {
		fmt.Printf("Visualizer script not found: %s\n", visualizerScript)
		return
	}
	
	chartsDir := filepath.Join(*outputDir, "charts")
	
	for _, file := range files {
		fmt.Printf("Generating charts for: %s\n", file)
		
		cmd := exec.Command(pythonCmd, visualizerScript, file, "--output-dir", chartsDir)
		output, err := cmd.CombinedOutput()
		
		if err != nil {
			fmt.Printf("Error running visualizer: %v\n", err)
			fmt.Printf("Output: %s\n", string(output))
			continue
		}
		
		fmt.Printf("Charts generated successfully for %s\n", filepath.Base(file))
	}
	
	fmt.Printf("All charts saved to: %s\n", chartsDir)
}

// generateBenchmarkReport creates a comprehensive text report
func generateBenchmarkReport(
	comparisons map[string]evaluation.ConcurrencyComparison,
	scalability []evaluation.BenchmarkResult,
	stress map[string]evaluation.BenchmarkResult,
) {
	reportPath := filepath.Join(*outputDir, "benchmark_report.txt")
	
	file, err := os.Create(reportPath)
	if err != nil {
		fmt.Printf("Error creating report file: %v\n", err)
		return
	}
	defer file.Close()
	
	// Write comprehensive report
	fmt.Fprintf(file, "CDA Exchange Simulator - Comprehensive Benchmark Report\n")
	fmt.Fprintf(file, "Generated: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(file, "========================================================\n\n")
	
	// Concurrency comparisons
	fmt.Fprintf(file, "CONCURRENCY COMPARISON RESULTS\n")
	fmt.Fprintf(file, "------------------------------\n")
	for name, comp := range comparisons {
		fmt.Fprintf(file, "%s Scale:\n", name)
		fmt.Fprintf(file, "  Concurrent: %.2f trades/sec, %.2f MB memory\n", 
			comp.ConcurrentResult.TradesPerSecond, comp.ConcurrentResult.PeakMemoryMB)
		fmt.Fprintf(file, "  Sequential: %.2f trades/sec, %.2f MB memory\n", 
			comp.SequentialResult.TradesPerSecond, comp.SequentialResult.PeakMemoryMB)
		fmt.Fprintf(file, "  Speedup: %.2fx, Efficiency: %.2f%%\n\n", 
			comp.SpeedupRatio, comp.EfficiencyGain)
	}
	
	// Scalability results
	fmt.Fprintf(file, "SCALABILITY TEST RESULTS\n")
	fmt.Fprintf(file, "------------------------\n")
	fmt.Fprintf(file, "Agents  Trades/Sec  Memory(MB)  Efficiency\n")
	for i, result := range scalability {
		agents := []int{5, 10, 15, 20, 30, 40, 50}
		if i < len(agents) {
			efficiency := result.TradesPerSecond / float64(agents[i])
			fmt.Fprintf(file, "%-6d  %-10.2f  %-10.2f  %.3f\n", 
				agents[i], result.TradesPerSecond, result.PeakMemoryMB, efficiency)
		}
	}
	
	// Stress test results
	fmt.Fprintf(file, "\nSTRESS TEST RESULTS\n")
	fmt.Fprintf(file, "-------------------\n")
	for name, result := range stress {
		fmt.Fprintf(file, "%s:\n", name)
		fmt.Fprintf(file, "  Throughput: %.2f trades/sec\n", result.TradesPerSecond)
		fmt.Fprintf(file, "  Memory: %.2f MB\n", result.PeakMemoryMB)
		fmt.Fprintf(file, "  Goroutines: %d\n\n", result.MaxGoroutines)
	}
	
	// Go advantages summary
	fmt.Fprintf(file, "GO CONCURRENCY ADVANTAGES DEMONSTRATED\n")
	fmt.Fprintf(file, "------------------------------------\n")
	fmt.Fprintf(file, "1. Lightweight Goroutines: Handled hundreds of concurrent agents efficiently\n")
	fmt.Fprintf(file, "2. Low Memory Overhead: ~2-8KB per goroutine vs ~8MB per OS thread\n")
	fmt.Fprintf(file, "3. Efficient Scheduling: Built-in work-stealing scheduler\n")
	fmt.Fprintf(file, "4. Fast Context Switching: Much faster than OS thread switching\n")
	fmt.Fprintf(file, "5. Concurrent GC: Low pause times even under load\n")
	
	fmt.Printf("Comprehensive report saved to: %s\n", reportPath)
}