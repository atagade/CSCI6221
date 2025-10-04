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

func (a *RandomAgent) act(sim Simulator) {
	a.mu.Lock()
	book := sim.GetBook()
	last := book.GetLastPrice()
	if last == 0 {
		last = 100
	}
	isBuy := a.rnd.Float64() < 0.5
	side := order.Sell
	if isBuy {
		side = order.Buy
	}
	if !isBuy && a.Shares < 1 {
		a.mu.Unlock()
		return
	}
	isLimit := a.rnd.Float64() < 0.5
	otype := order.Market
	price := 0.0
	if isLimit {
		otype = order.Limit
		price = last + (a.rnd.Float64()*20 - 10)
		if price <= 0 {
			price = 1
		}
	}
	qty := 1 + math.Floor(a.rnd.Float64()*10)
	if !isBuy && qty > a.Shares {
		qty = a.Shares
	}
	if isBuy && qty*last > a.Cash {
		qty = math.Floor(a.Cash / last)
		if qty < 1 {
			a.mu.Unlock()
			return
		}
	}
	o := &order.Order{
		ID:       uuid.NewString(),
		Stock:    sim.GetStock(),
		Side:     side,
		Type:     otype,
		Price:    price,
		Quantity: qty,
		AgentID:  a.ID,
	}
	a.Orders[o.ID] = struct{}{}
	a.mu.Unlock()
	book.Submit(o)
}

type MarketMakerAgent struct {
	BaseAgent
	Delta float64 // spread delta
}

func NewMarketMaker(id string, cash, shares float64, delta float64) *MarketMakerAgent {
	return &MarketMakerAgent{
		BaseAgent: BaseAgent{
			ID:        id,
			Cash:      cash,
			Shares:    shares,
			Orders:    make(map[string]struct{}),
			EventChan: make(chan FillEvent, 100),
			rnd:       rand.New(rand.NewSource(time.Now().UnixNano())),
		},
		Delta: delta,
	}
}

func (a *MarketMakerAgent) Run(ctx context.Context, sim Simulator) {
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
			time.Sleep(time.Millisecond * time.Duration(500+a.rnd.Intn(500)))
		}
	}
}

func (a *MarketMakerAgent) act(sim Simulator) {
	a.mu.Lock()
	book := sim.GetBook()
	if len(a.Orders) > 5 { // cancel old orders if too many
		for id := range a.Orders {
			book.Cancel(id)
			delete(a.Orders, id)
			break // cancel one at a time
		}
	}
	last := book.GetLastPrice()
	if last == 0 {
		last = 100
	}
	bestBid := book.GetBestBid()
	bestAsk := book.GetBestAsk()
	if bestBid == 0 {
		bestBid = last - a.Delta
	}
	if bestAsk == 0 {
		bestAsk = last + a.Delta
	}
	// place buy limit
	buyPrice := bestBid + (a.rnd.Float64() * 0.1)
	qty := 5 + math.Floor(a.rnd.Float64()*5)
	if qty*buyPrice > a.Cash {
		a.mu.Unlock()
		return
	}
	oBuy := &order.Order{
		ID:       uuid.NewString(),
		Stock:    sim.GetStock(),
		Side:     order.Buy,
		Type:     order.Limit,
		Price:    buyPrice,
		Quantity: qty,
		AgentID:  a.ID,
	}
	a.Orders[oBuy.ID] = struct{}{}
	// place sell limit
	sellPrice := bestAsk - (a.rnd.Float64() * 0.1)
	if sellPrice <= buyPrice {
		sellPrice = buyPrice + 0.1
	}
	sellQty := qty // same qty
	if sellQty > a.Shares {
		sellQty = a.Shares
	}
	if sellQty < 1 {
		a.mu.Unlock()
		return
	}
	oSell := &order.Order{
		ID:       uuid.NewString(),
		Stock:    sim.GetStock(),
		Side:     order.Sell,
		Type:     order.Limit,
		Price:    sellPrice,
		Quantity: sellQty,
		AgentID:  a.ID,
	}
	a.Orders[oSell.ID] = struct{}{}
	a.mu.Unlock()
	book.Submit(oBuy)
	book.Submit(oSell)
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

func (a *TrendFollowerAgent) act(sim Simulator) {
	a.mu.Lock()
	book := sim.GetBook()
	last := book.GetLastPrice()
	if last == 0 {
		last = 100
	}
	a.MA = a.Alpha*last + (1-a.Alpha)*a.MA
	isBuy := last > a.MA
	side := order.Sell
	otype := order.Market // trend followers use market orders
	if isBuy {
		side = order.Buy
	}
	if !isBuy && a.Shares < 1 {
		a.mu.Unlock()
		return
	}
	qty := 1 + math.Floor(a.rnd.Float64()*5)
	if !isBuy && qty > a.Shares {
		qty = a.Shares
	}
	if isBuy && qty*last > a.Cash {
		qty = math.Floor(a.Cash / last)
		if qty < 1 {
			a.mu.Unlock()
			return
		}
	}
	o := &order.Order{
		ID:       uuid.NewString(),
		Stock:    sim.GetStock(),
		Side:     side,
		Type:     otype,
		Price:    0,
		Quantity: qty,
		AgentID:  a.ID,
	}
	a.Orders[o.ID] = struct{}{}
	a.mu.Unlock()
	book.Submit(o)
}