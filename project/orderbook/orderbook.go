package orderbook

import (
	"container/list"
	"math"
	"sync"
	"time"

	"cda-simulator/order"
)

type OrderBook struct {
	mu          sync.RWMutex
	bids        map[float64]*list.List // price -> list of *order.Order (FIFO)
	asks        map[float64]*list.List
	orders      map[string]*list.Element // ID -> element for quick cancel
	lastPrice   float64
	bestBid     float64
	bestAsk     float64
	processChan chan Event
	OnTrade     func(Trade)
}

type Event interface {
	process(*OrderBook)
}
