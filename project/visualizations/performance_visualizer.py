#!/usr/bin/env python3
"""
CDA Exchange Simulator Performance Visualization

This module provides comprehensive visualization capabilities for analyzing
the performance of the Go-based Continuous Double Auction Exchange Simulator.
It generates various charts and graphs to showcase Go's concurrency advantages.
"""

import json
import csv
import argparse
import matplotlib.pyplot as plt
import numpy as np
import pandas as pd
from datetime import datetime
import seaborn as sns
from pathlib import Path
import sys

# Set up matplotlib for better plots
plt.style.use('seaborn-v0_8')
sns.set_palette("husl")

class PerformanceVisualizer:
    """Main class for generating performance visualizations"""
    
    def __init__(self, output_dir="visualizations"):
        """Initialize the visualizer with output directory"""
        self.output_dir = Path(output_dir)
        self.output_dir.mkdir(exist_ok=True)
        
    def load_metrics_json(self, json_file):
        """Load performance metrics from JSON file"""
        try:
            with open(json_file, 'r') as f:
                return json.load(f)
        except Exception as e:
            print(f"Error loading JSON file: {e}")
            return None
    
    def load_comparison_json(self, json_file):
        """Load concurrency comparison data from JSON file"""
        try:
            with open(json_file, 'r') as f:
                return json.load(f)
        except Exception as e:
            print(f"Error loading comparison JSON file: {e}")
            return None
    
    def create_throughput_chart(self, metrics, title="Trading Throughput Analysis"):
        """Create a throughput analysis chart"""
        fig, ((ax1, ax2), (ax3, ax4)) = plt.subplots(2, 2, figsize=(15, 12))
        fig.suptitle(title, fontsize=16, fontweight='bold')
        
        # 1. Trades and Orders per Second
        categories = ['Trades/Sec', 'Orders/Sec']
        values = [metrics['trades_per_second'], metrics['orders_per_second']]
        colors = ['#FF6B6B', '#4ECDC4']
        
        bars1 = ax1.bar(categories, values, color=colors, alpha=0.8)
        ax1.set_title('Throughput Metrics', fontweight='bold')
        ax1.set_ylabel('Operations per Second')
        
        # Add value labels on bars
        for bar, value in zip(bars1, values):
            ax1.text(bar.get_x() + bar.get_width()/2, bar.get_height() + 0.5,
                    f'{value:.1f}', ha='center', va='bottom', fontweight='bold')
        
        # 2. Total Operations Breakdown
        operations = ['Total Trades', 'Total Orders', 'Cancellations']
        op_values = [metrics['total_trades'], metrics['total_orders'], metrics['total_cancellations']]
        colors2 = ['#FF9F43', '#10AC84', '#EE5A24']
        
        bars2 = ax2.bar(operations, op_values, color=colors2, alpha=0.8)
        ax2.set_title('Total Operations', fontweight='bold')
        ax2.set_ylabel('Count')
        ax2.tick_params(axis='x', rotation=45)
        
        # Add value labels
        for bar, value in zip(bars2, op_values):
            ax2.text(bar.get_x() + bar.get_width()/2, bar.get_height() + max(op_values)*0.01,
                    f'{value}', ha='center', va='bottom', fontweight='bold')
        
        # 3. Latency Analysis
        latencies = ['Avg Order Latency', 'Avg Trade Latency']
        # Convert nanoseconds to microseconds for readability
        lat_values = [
            metrics['avg_order_latency'] / 1000,  # ns to μs
            metrics['avg_trade_latency'] / 1000   # ns to μs
        ]
        colors3 = ['#A55EEA', '#26C6DA']
        
        bars3 = ax3.bar(latencies, lat_values, color=colors3, alpha=0.8)
        ax3.set_title('Latency Analysis', fontweight='bold')
        ax3.set_ylabel('Latency (μs)')
        ax3.tick_params(axis='x', rotation=45)
        
        # Add value labels
        for bar, value in zip(bars3, lat_values):
            ax3.text(bar.get_x() + bar.get_width()/2, bar.get_height() + max(lat_values)*0.01,
                    f'{value:.1f}μs', ha='center', va='bottom', fontweight='bold')
        
        # 4. Resource Utilization
        resources = ['Peak Memory\n(MB)', 'Max Goroutines', 'CPUs']
        res_values = [
            metrics['peak_memory_usage'] / (1024 * 1024),  # bytes to MB
            metrics['max_goroutines'],
            metrics['num_cpu']
        ]
        colors4 = ['#FD79A8', '#FDCB6E', '#6C5CE7']
        
        bars4 = ax4.bar(resources, res_values, color=colors4, alpha=0.8)
        ax4.set_title('Resource Utilization', fontweight='bold')
        ax4.set_ylabel('Count/Size')
        
        # Add value labels
        for bar, value in zip(bars4, res_values):
            ax4.text(bar.get_x() + bar.get_width()/2, bar.get_height() + max(res_values)*0.01,
                    f'{value:.1f}', ha='center', va='bottom', fontweight='bold')
        
        plt.tight_layout()
        output_file = self.output_dir / f"throughput_analysis.png"
        plt.savefig(output_file, dpi=300, bbox_inches='tight')
        print(f"Throughput chart saved to: {output_file}")
        plt.show()
        
        return fig
    
    def create_latency_histogram(self, metrics, title="Latency Distribution Analysis"):
        """Create latency histogram visualization"""
        fig, (ax1, ax2) = plt.subplots(1, 2, figsize=(15, 6))
        fig.suptitle(title, fontsize=16, fontweight='bold')
        
        # Order Latency Histogram
        if 'order_latency_histogram' in metrics and metrics['order_latency_histogram']:
            hist_data = metrics['order_latency_histogram']
            bins = range(len(hist_data))
            
            ax1.bar(bins, hist_data, color='#FF6B6B', alpha=0.7, edgecolor='black')
            ax1.set_title('Order Latency Distribution', fontweight='bold')
            ax1.set_xlabel('Latency Buckets')
            ax1.set_ylabel('Frequency')
            ax1.grid(True, alpha=0.3)
        
        # Trade Latency Histogram
        if 'trade_latency_histogram' in metrics and metrics['trade_latency_histogram']:
            hist_data = metrics['trade_latency_histogram']
            bins = range(len(hist_data))
            
            ax2.bar(bins, hist_data, color='#4ECDC4', alpha=0.7, edgecolor='black')
            ax2.set_title('Trade Latency Distribution', fontweight='bold')
            ax2.set_xlabel('Latency Buckets')
            ax2.set_ylabel('Frequency')
            ax2.grid(True, alpha=0.3)
        
        plt.tight_layout()
        output_file = self.output_dir / f"latency_histogram.png"
        plt.savefig(output_file, dpi=300, bbox_inches='tight')
        print(f"Latency histogram saved to: {output_file}")
        plt.show()
        
        return fig
    
    def create_memory_timeline(self, metrics, title="Memory Usage Timeline"):
        """Create memory usage timeline visualization"""
        if 'memory_snapshots' not in metrics or not metrics['memory_snapshots']:
            print("No memory snapshots available for timeline")
            return None
        
        # Extract memory data
        snapshots = metrics['memory_snapshots']
        timestamps = [datetime.fromisoformat(s['timestamp'].replace('Z', '+00:00')) for s in snapshots]
        heap_alloc = [s['heap_alloc'] / (1024 * 1024) for s in snapshots]  # Convert to MB
        heap_sys = [s['heap_sys'] / (1024 * 1024) for s in snapshots]
        goroutines = [s['num_goroutine'] for s in snapshots]
        
        fig, (ax1, ax2) = plt.subplots(2, 1, figsize=(15, 10))
        fig.suptitle(title, fontsize=16, fontweight='bold')
        
        # Memory usage over time
        ax1.plot(timestamps, heap_alloc, label='Heap Allocated', color='#FF6B6B', linewidth=2)
        ax1.plot(timestamps, heap_sys, label='Heap System', color='#4ECDC4', linewidth=2)
        ax1.set_title('Memory Usage Over Time', fontweight='bold')
        ax1.set_ylabel('Memory (MB)')
        ax1.legend()
        ax1.grid(True, alpha=0.3)
        
        # Goroutine count over time
        ax2.plot(timestamps, goroutines, label='Goroutines', color='#FFA726', linewidth=2)
        ax2.set_title('Goroutine Count Over Time', fontweight='bold')
        ax2.set_xlabel('Time')
        ax2.set_ylabel('Number of Goroutines')
        ax2.legend()
        ax2.grid(True, alpha=0.3)
        
        plt.tight_layout()
        output_file = self.output_dir / f"memory_timeline.png"
        plt.savefig(output_file, dpi=300, bbox_inches='tight')
        print(f"Memory timeline saved to: {output_file}")
        plt.show()
        
        return fig
    
    def create_concurrency_comparison(self, comparison_data, title="Go Concurrency vs Sequential Performance"):
        """Create concurrency comparison visualization"""
        if not comparison_data:
            print("No comparison data available")
            return None
        
        fig, ((ax1, ax2), (ax3, ax4)) = plt.subplots(2, 2, figsize=(16, 12))
        fig.suptitle(title, fontsize=16, fontweight='bold')
        
        concurrent = comparison_data['concurrent_result']
        sequential = comparison_data['sequential_result']
        
        # 1. Throughput Comparison
        categories = ['Trades/Second', 'Peak Memory (MB)', 'GC Pause (ms)']
        concurrent_values = [
            concurrent.get('trades_per_second', 0),
            concurrent.get('peak_memory_mb', 0),
            concurrent.get('gc_pause_time_ms', 0)
        ]
        sequential_values = [
            sequential.get('trades_per_second', 0),
            sequential.get('peak_memory_mb', 0),
            sequential.get('gc_pause_time_ms', 0)
        ]
        
        x = np.arange(len(categories))
        width = 0.35
        
        bars1 = ax1.bar(x - width/2, concurrent_values, width, label='Concurrent', 
                       color='#4ECDC4', alpha=0.8)
        bars2 = ax1.bar(x + width/2, sequential_values, width, label='Sequential', 
                       color='#FF6B6B', alpha=0.8)
        
        ax1.set_title('Performance Comparison', fontweight='bold')
        ax1.set_xticks(x)
        ax1.set_xticklabels(categories)
        ax1.legend()
        ax1.set_ylabel('Value')
        
        # Add value labels
        for bars in [bars1, bars2]:
            for bar in bars:
                height = bar.get_height()
                ax1.text(bar.get_x() + bar.get_width()/2., height + height*0.01,
                        f'{height:.1f}', ha='center', va='bottom', fontsize=9)
        
        # 2. Speedup and Efficiency
        metrics = ['Speedup Ratio', 'Efficiency Gain (%)', 'Memory Overhead (%)']
        values = [
            comparison_data.get('speedup_ratio', 1.0),
            comparison_data.get('efficiency_gain', 0.0),
            comparison_data.get('memory_overhead', 0.0)
        ]
        colors = ['#10AC84', '#FFA726', '#EE5A24']
        
        bars3 = ax2.bar(metrics, values, color=colors, alpha=0.8)
        ax2.set_title('Concurrency Benefits', fontweight='bold')
        ax2.set_ylabel('Ratio/Percentage')
        ax2.tick_params(axis='x', rotation=45)
        
        # Add value labels
        for bar, value in zip(bars3, values):
            height = bar.get_height()
            ax2.text(bar.get_x() + bar.get_width()/2, height + abs(height)*0.05,
                    f'{value:.2f}', ha='center', va='bottom', fontweight='bold')
        
        # 3. Latency Comparison
        latency_types = ['Avg Latency (ns)']
        conc_latency = [concurrent.get('avg_latency', 0)]
        seq_latency = [sequential.get('avg_latency', 0)]
        
        x2 = np.arange(len(latency_types))
        bars4 = ax3.bar(x2 - width/2, conc_latency, width, label='Concurrent', 
                       color='#4ECDC4', alpha=0.8)
        bars5 = ax3.bar(x2 + width/2, seq_latency, width, label='Sequential', 
                       color='#FF6B6B', alpha=0.8)
        
        ax3.set_title('Latency Comparison', fontweight='bold')
        ax3.set_xticks(x2)
        ax3.set_xticklabels(latency_types)
        ax3.legend()
        ax3.set_ylabel('Latency (ns)')
        
        # Add value labels
        for bars in [bars4, bars5]:
            for bar in bars:
                height = bar.get_height()
                ax3.text(bar.get_x() + bar.get_width()/2., height + height*0.01,
                        f'{height:.0f}', ha='center', va='bottom', fontsize=10)
        
        # 4. Go Advantages Summary
        advantages = ['Goroutines\nEfficiency', 'Memory\nManagement', 'GC\nPerformance']
        speedup_ratio = comparison_data.get('speedup_ratio', 1.0)
        memory_overhead = comparison_data.get('memory_overhead', 0.0)
        gc_pause = concurrent.get('gc_pause_time_ms', 0)
        
        scores = [
            min(100, max(0, speedup_ratio * 30)),  # Goroutine efficiency score
            max(0, min(100, 100 - abs(memory_overhead))),  # Memory efficiency score
            max(0, min(100, 100 - gc_pause * 10))  # GC performance score
        ]
        
        colors_adv = ['#6C5CE7', '#A55EEA', '#FD79A8']
        bars6 = ax4.bar(advantages, scores, color=colors_adv, alpha=0.8)
        ax4.set_title('Go Concurrency Advantages', fontweight='bold')
        ax4.set_ylabel('Performance Score (0-100)')
        ax4.set_ylim(0, 100)
        
        # Add value labels
        for bar, score in zip(bars6, scores):
            ax4.text(bar.get_x() + bar.get_width()/2, bar.get_height() + 2,
                    f'{score:.0f}', ha='center', va='bottom', fontweight='bold')
        
        plt.tight_layout()
        output_file = self.output_dir / f"concurrency_comparison.png"
        plt.savefig(output_file, dpi=300, bbox_inches='tight')
        print(f"Concurrency comparison saved to: {output_file}")
        plt.show()
        
        return fig
    
    def create_scalability_analysis(self, scalability_data, title="Scalability Analysis"):
        """Create scalability analysis visualization"""
        if not scalability_data or len(scalability_data) < 2:
            print("Insufficient scalability data")
            return None
        
        # Extract data
        agent_counts = []
        throughput = []
        memory_usage = []
        max_goroutines = []
        
        for result in scalability_data:
            # Extract agent count from name (e.g., "Concurrent_5_agents" -> 5)
            name = result.get('name', '')
            try:
                if 'agents' in name:
                    count = int(name.split('_')[1])
                    agent_counts.append(count)
                else:
                    # Fallback: use index
                    agent_counts.append(len(agent_counts) + 1)
            except (IndexError, ValueError):
                # Fallback: use index
                agent_counts.append(len(agent_counts) + 1)
            
            throughput.append(result.get('trades_per_second', 0))
            memory_usage.append(result.get('peak_memory_mb', 0))
            max_goroutines.append(result.get('max_goroutines', 0))
        
        fig, ((ax1, ax2), (ax3, ax4)) = plt.subplots(2, 2, figsize=(15, 12))
        fig.suptitle(title, fontsize=16, fontweight='bold')
        
        # 1. Throughput vs Agents
        ax1.plot(agent_counts, throughput, marker='o', linewidth=3, markersize=8, color='#4ECDC4')
        ax1.set_title('Throughput Scalability', fontweight='bold')
        ax1.set_xlabel('Number of Agents')
        ax1.set_ylabel('Trades per Second')
        ax1.grid(True, alpha=0.3)
        
        # Add value labels
        for x, y in zip(agent_counts, throughput):
            ax1.text(x, y + max(throughput)*0.02, f'{y:.1f}', ha='center', va='bottom')
        
        # 2. Memory Usage vs Agents
        ax2.plot(agent_counts, memory_usage, marker='s', linewidth=3, markersize=8, color='#FF6B6B')
        ax2.set_title('Memory Usage Scalability', fontweight='bold')
        ax2.set_xlabel('Number of Agents')
        ax2.set_ylabel('Peak Memory (MB)')
        ax2.grid(True, alpha=0.3)
        
        # Add value labels
        for x, y in zip(agent_counts, memory_usage):
            ax2.text(x, y + max(memory_usage)*0.02, f'{y:.1f}', ha='center', va='bottom')
        
        # 3. Goroutines vs Agents
        ax3.plot(agent_counts, max_goroutines, marker='^', linewidth=3, markersize=8, color='#FFA726')
        ax3.set_title('Goroutine Scalability', fontweight='bold')
        ax3.set_xlabel('Number of Agents')
        ax3.set_ylabel('Max Goroutines')
        ax3.grid(True, alpha=0.3)
        
        # Add value labels
        for x, y in zip(agent_counts, max_goroutines):
            ax3.text(x, y + max(max_goroutines)*0.02, f'{y}', ha='center', va='bottom')
        
        # 4. Efficiency Ratio (Throughput per Agent)
        efficiency = [t/a if a > 0 else 0 for t, a in zip(throughput, agent_counts)]
        ax4.plot(agent_counts, efficiency, marker='D', linewidth=3, markersize=8, color='#10AC84')
        ax4.set_title('Efficiency per Agent', fontweight='bold')
        ax4.set_xlabel('Number of Agents')
        ax4.set_ylabel('Trades per Second per Agent')
        ax4.grid(True, alpha=0.3)
        
        # Add value labels
        for x, y in zip(agent_counts, efficiency):
            ax4.text(x, y + max(efficiency)*0.02, f'{y:.2f}', ha='center', va='bottom')
        
        plt.tight_layout()
        output_file = self.output_dir / f"scalability_analysis.png"
        plt.savefig(output_file, dpi=300, bbox_inches='tight')
        print(f"Scalability analysis saved to: {output_file}")
        plt.show()
        
        return fig
    
    def generate_comprehensive_report(self, metrics_file, comparison_file=None, scalability_file=None):
        """Generate a comprehensive visualization report"""
        print("Generating comprehensive performance report...")
        
        # Load main metrics
        metrics = self.load_metrics_json(metrics_file)
        if not metrics:
            print(f"Failed to load metrics from {metrics_file}")
            return
        
        # Determine the type of data we're dealing with
        if isinstance(metrics, dict):
            # Check if it's a comparison file
            if 'concurrent_result' in metrics and 'sequential_result' in metrics:
                print("Detected concurrency comparison data")
                self.create_concurrency_comparison(metrics)
                return
            # Check if it's regular metrics data
            elif 'trades_per_second' in metrics:
                print("Detected regular metrics data")
                self.create_throughput_chart(metrics)
                self.create_latency_histogram(metrics)
                self.create_memory_timeline(metrics)
            else:
                print("Detected stress test or other result data")
                # Handle stress test results (dict of results)
                for name, result in metrics.items():
                    if isinstance(result, dict) and 'trades_per_second' in result:
                        print(f"Creating chart for {name}")
                        self.create_single_result_chart(result, f"Stress Test: {name}")
        elif isinstance(metrics, list):
            # Scalability data (list of benchmark results)
            print("Detected scalability data")
            self.create_scalability_analysis(metrics)
        else:
            print(f"Unknown data format in {metrics_file}")
            return
        
        # Create comparison visualization if data available
        if comparison_file:
            comparison_data = self.load_comparison_json(comparison_file)
            if comparison_data:
                self.create_concurrency_comparison(comparison_data)
        
        # Create scalability visualization if data available
        if scalability_file:
            scalability_data = self.load_metrics_json(scalability_file)
            if scalability_data and isinstance(scalability_data, list):
                self.create_scalability_analysis(scalability_data)
        
        print(f"\nAll visualizations saved to: {self.output_dir}")
    
    def create_single_result_chart(self, result, title):
        """Create a chart for a single benchmark result"""
        fig, ((ax1, ax2), (ax3, ax4)) = plt.subplots(2, 2, figsize=(15, 10))
        fig.suptitle(title, fontsize=16, fontweight='bold')
        
        # Basic metrics
        metrics_names = ['Trades/Sec', 'Peak Memory (MB)', 'Max Goroutines']
        metrics_values = [
            result.get('trades_per_second', 0),
            result.get('peak_memory_mb', 0),
            result.get('max_goroutines', 0)
        ]
        
        bars1 = ax1.bar(metrics_names, metrics_values, color=['#FF6B6B', '#4ECDC4', '#FFA726'], alpha=0.8)
        ax1.set_title('Performance Metrics', fontweight='bold')
        ax1.set_ylabel('Value')
        ax1.tick_params(axis='x', rotation=45)
        
        # Add value labels
        for bar, value in zip(bars1, metrics_values):
            ax1.text(bar.get_x() + bar.get_width()/2, bar.get_height() + max(metrics_values)*0.01,
                    f'{value:.2f}', ha='center', va='bottom', fontweight='bold')
        
        # Additional metrics
        ax2.text(0.5, 0.7, f"Name: {result.get('name', 'Unknown')}", 
                ha='center', va='center', transform=ax2.transAxes, fontsize=12, fontweight='bold')
        ax2.text(0.5, 0.5, f"Duration: {result.get('duration', 0)/1e9:.2f}s", 
                ha='center', va='center', transform=ax2.transAxes, fontsize=12)
        ax2.text(0.5, 0.3, f"Total Trades: {result.get('total_trades', 0)}", 
                ha='center', va='center', transform=ax2.transAxes, fontsize=12)
        ax2.set_title('Test Information', fontweight='bold')
        ax2.axis('off')
        
        # Performance indicators
        ax3.text(0.5, 0.7, f"CPU Utilization: {result.get('cpu_utilization', 0):.1f}%", 
                ha='center', va='center', transform=ax3.transAxes, fontsize=12)
        ax3.text(0.5, 0.5, f"GC Pause: {result.get('gc_pause_time_ms', 0):.2f}ms", 
                ha='center', va='center', transform=ax3.transAxes, fontsize=12)
        ax3.text(0.5, 0.3, f"Avg Latency: {result.get('avg_latency', 0)}ns", 
                ha='center', va='center', transform=ax3.transAxes, fontsize=12)
        ax3.set_title('Performance Details', fontweight='bold')
        ax3.axis('off')
        
        # Summary
        ax4.text(0.5, 0.6, "Go Concurrency Advantages:", 
                ha='center', va='center', transform=ax4.transAxes, fontsize=14, fontweight='bold')
        ax4.text(0.5, 0.4, "✓ Lightweight goroutines", 
                ha='center', va='center', transform=ax4.transAxes, fontsize=12)
        ax4.text(0.5, 0.2, "✓ Efficient memory usage", 
                ha='center', va='center', transform=ax4.transAxes, fontsize=12)
        ax4.set_title('Summary', fontweight='bold')
        ax4.axis('off')
        
        plt.tight_layout()
        safe_title = title.replace(' ', '_').replace(':', '_')
        output_file = self.output_dir / f"{safe_title.lower()}_chart.png"
        plt.savefig(output_file, dpi=300, bbox_inches='tight')
        print(f"Chart saved to: {output_file}")
        plt.show()
        
        return fig

def main():
    """Main function for command-line usage"""
    parser = argparse.ArgumentParser(description='CDA Exchange Simulator Performance Visualizer')
    parser.add_argument('metrics_file', help='JSON file containing performance metrics')
    parser.add_argument('--comparison', help='JSON file containing concurrency comparison data')
    parser.add_argument('--scalability', help='JSON file containing scalability test data')
    parser.add_argument('--output-dir', default='visualizations', help='Output directory for charts')
    
    args = parser.parse_args()
    
    # Create visualizer
    visualizer = PerformanceVisualizer(args.output_dir)
    
    # Generate comprehensive report
    visualizer.generate_comprehensive_report(
        args.metrics_file,
        args.comparison,
        args.scalability
    )

if __name__ == "__main__":
    main()