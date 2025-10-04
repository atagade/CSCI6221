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

type Submit struct {
	Order *order.Order
	Done  chan<- []Trade
}

type Cancel struct {
	ID   string
	Done chan<- bool
}

type Trade struct {
	Price          float64
	Quantity       float64
	BuyerID        string
	SellerID       string
	BuyerOrderID   string
	SellerOrderID  string
}

func New() *OrderBook {
	ob := &OrderBook{
		bids:        make(map[float64]*list.List),
		asks:        make(map[float64]*list.List),
		orders:      make(map[string]*list.Element),
		processChan: make(chan Event),
	}
	go ob.processor()
	return ob
}


func (s *Submit) process(ob *OrderBook) {
	if s.Order.Quantity <= 0 || (s.Order.Type == order.Limit && s.Order.Price <= 0) {
		s.Done <- nil
		return
	}

	o := s.Order
	o.Time = time.Now()

	var trades []Trade
	if o.Type == order.Market {
		trades = ob.matchMarket(o)
	} else {
		trades = ob.matchLimit(o)
	}

	if o.Quantity > 0 && o.Type == order.Limit {
		ob.addToBook(o)
	}

	s.Done <- trades
}

func (c *Cancel) process(ob *OrderBook) {
	elem, ok := ob.orders[c.ID]
	if !ok {
		c.Done <- false
		return
	}

	o := elem.Value.(*order.Order)
	price := o.Price
	side := o.Side

	// Find the correct list based on price and side
	var l *list.List
	if side == order.Buy {
		l = ob.bids[price]
	} else {
		l = ob.asks[price]
	}
	
	if l != nil {
		l.Remove(elem)
	}
	delete(ob.orders, c.ID)

	if l != nil && l.Len() == 0 {
		m := ob.bids
		isBid := side == order.Buy
		if !isBid {
			m = ob.asks
		}
		delete(m, price)
		if isBid {
			if price == ob.bestBid {
				ob.bestBid = ob.findBestBidNoLock()
			}
		} else {
			if price == ob.bestAsk {
				ob.bestAsk = ob.findBestAskNoLock()
			}
		}
	}
	c.Done <- true
}

