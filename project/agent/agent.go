package agent

import (
	"context"
	"math"
	"math/rand"
	"sync"
	"time"

)

type Agent interface {
	Run(context.Context, Simulator)
}

type Simulator interface {
	GetBook() *orderbook.OrderBook
	GetStock() string
}

type FillEvent struct {
	OrderID  string
	Price    float64
	Quantity float64
	IsBuy    bool
}

type BaseAgent struct {
	mu        sync.Mutex
	ID        string
	Cash      float64
	Shares    float64
	Orders    map[string]struct{}
	EventChan chan FillEvent
	rnd       *rand.Rand
}

