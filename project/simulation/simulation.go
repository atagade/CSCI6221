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

func New() *Sim {
	s := &Sim{
		Book:   orderbook.New(),
		Agents: make(map[string]*agent.BaseAgent),
		Stock:  "GOOG",
	}
	s.Book.OnTrade = s.handleTrade
	return s
}

func (s *Sim) GetBook() *orderbook.OrderBook {
	return s.Book
}

func (s *Sim) GetStock() string {
	return s.Stock
}

func (s *Sim) handleTrade(t orderbook.Trade) {
	s.agentsMu.RLock()
	defer s.agentsMu.RUnlock()
	
	if buyer, ok := s.Agents[t.BuyerID]; ok {
		buyer.EventChan <- agent.FillEvent{OrderID: t.BuyerOrderID, Price: t.Price, Quantity: t.Quantity, IsBuy: true}
	}
	if seller, ok := s.Agents[t.SellerID]; ok {
		seller.EventChan <- agent.FillEvent{OrderID: t.SellerOrderID, Price: t.Price, Quantity: t.Quantity, IsBuy: false}
	}
	atomic.AddInt64(&s.tradeCount, 1)
}

func (s *Sim) GetTradeCount() int64 {
	return atomic.LoadInt64(&s.tradeCount)
}

func (s *Sim) AddAgent(id string, agent *agent.BaseAgent) {
	s.agentsMu.Lock()
	defer s.agentsMu.Unlock()
	s.Agents[id] = agent
}