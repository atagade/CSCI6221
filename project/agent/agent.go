package agent

import (
	"context"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"

	"cda-simulator/order"
	"cda-simulator/orderbook"
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

func (a *BaseAgent) handle(ev FillEvent) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if ev.IsBuy {
		a.Cash -= ev.Price * ev.Quantity
		a.Shares += ev.Quantity
	} else {
		a.Cash += ev.Price * ev.Quantity
		a.Shares -= ev.Quantity
	}
	// Note: For partial fills, this deletes on first fill; in production, track remaining qty
	delete(a.Orders, ev.OrderID)
}

type RandomAgent struct {
	BaseAgent
}

func NewRandom(id string, cash, shares float64) *RandomAgent {
	return &RandomAgent{
		BaseAgent: BaseAgent{
			ID:        id,
			Cash:      cash,
			Shares:    shares,
			Orders:    make(map[string]struct{}),
			EventChan: make(chan FillEvent, 100),
			rnd:       rand.New(rand.NewSource(time.Now().UnixNano())),
		},
	}
}

func (a *RandomAgent) Run(ctx context.Context, sim Simulator) {
	go func() {
		for ev := range a.EventChan {
			a.handle(ev)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			close(a.EventChan)
			return
		default:
			a.act(sim)
			time.Sleep(time.Millisecond * time.Duration(100+a.rnd.Intn(900)))
		}
	}
}

type TrendFollowerAgent struct {
	BaseAgent
	MA    float64 // moving average
	Alpha float64 // EMA alpha
}

func NewTrendFollower(id string, cash, shares float64, alpha float64) *TrendFollowerAgent {
	return &TrendFollowerAgent{
		BaseAgent: BaseAgent{
			ID:        id,
			Cash:      cash,
			Shares:    shares,
			Orders:    make(map[string]struct{}),
			EventChan: make(chan FillEvent, 100),
			rnd:       rand.New(rand.NewSource(time.Now().UnixNano())),
		},
		MA:    100, // initial
		Alpha: alpha,
	}
}

func (a *TrendFollowerAgent) Run(ctx context.Context, sim Simulator) {
	go func() {
		for ev := range a.EventChan {
			a.handle(ev)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			close(a.EventChan)
			return
		default:
			a.act(sim)
			time.Sleep(time.Millisecond * time.Duration(200+a.rnd.Intn(800)))
		}
	}
}
