# Continuous Double-Auction (CDA) Exchange Simulator

A high-performance, concurrent Continuous Double-Auction exchange simulator implemented in Go, demonstrating advanced market microstructure simulation with autonomous trading agents and real-time order matching.

## 🎯 Project Overview

This project implements a robust MVP of a CDA exchange simulator that showcases Go's concurrency strengths through goroutines and channels. The simulator features a central order book with FIFO price-time priority matching and autonomous agents (random, market makers, trend followers) operating in parallel. The design achieves high throughput (easily >10k trades/sec on modest hardware, scalable to 100k+ with tuning) while maintaining strict order priority and thread safety.

### 🔬 Academic & Research Purpose

This simulator is designed for:
- **Market Microstructure Research**: Study how different agent behaviors affect market dynamics
- **Algorithm Testing**: Test trading strategies in controlled environments  
- **Performance Benchmarking**: Compare Go's concurrency model with other languages
- **Financial Education**: Demonstrate order book mechanics and market behavior
- **Flash Crash Studies**: Simulate market instability scenarios (e.g., trend-follower cascades)

## 🏗️ Architecture Overview

### Core Components

The simulator consists of four main packages, each with specific responsibilities:

```
cda-simulator/
├── main.go                 # Application entry point and orchestration
├── order/                  # Order structures and types
├── orderbook/              # Core matching engine with FIFO priority
├── agent/                  # Autonomous trading agent implementations  
└── simulation/             # Coordination and trade event handling
```

### 🧠 Design Principles

1. **Single-Writer Pattern**: OrderBook uses one goroutine for mutations (strict FIFO integrity)
2. **Multiple-Reader Access**: Concurrent read access via RWMutex for price queries
3. **Channel-Based Communication**: Event-driven architecture for order submission/cancellation
4. **Lock-Free Counters**: Atomic operations for high-frequency trade counting
5. **Memory Efficiency**: Best-price tracking with O(1) queries and rare O(n) rescans

## 📊 Key Features

### 🔄 Order Book Engine
- **Order Types**: 
  - **Limit Orders (GTC)**: Good-Till-Cancelled with specific price and quantity
  - **Market Orders (IOC)**: Immediate-Or-Cancel, executed at best available prices
- **Matching Algorithm**: 
  - **Price-Time Priority**: Best price first, then FIFO within price levels
  - **Level-by-Level Crossing**: Aggressive orders walk through price levels until filled
  - **Partial Fill Support**: Orders can be partially filled across multiple price levels
- **Data Structures**:
  - **FIFO Queues**: `container/list` for each price level ensuring time priority
  - **Price Maps**: Separate maps for bids and asks with fast price-level lookup
  - **Order Tracking**: Quick order cancellation via ID-to-element mapping
- **Concurrency Model**:
  - **Single Processor Goroutine**: Serializes all order book modifications
  - **RWMutex Protection**: Allows concurrent read access for price queries
  - **Channel Communication**: Thread-safe order submission and cancellation



---

**For questions, contributions, or extended research collaborations, please refer to the project documentation or contact the development team.**