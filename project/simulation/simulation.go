package simulation

import (
	"sync"
	"sync/atomic"

	"cda-simulator/agent"
	"cda-simulator/orderbook"
)

type Sim struct {
	Book       *orderbook.OrderBook
	Agents     map[string]*agent.BaseAgent
	agentsMu   sync.RWMutex
	Stock      string
	tradeCount int64
}

