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

func (ob *OrderBook) processor() {
	for ev := range ob.processChan {
		ob.mu.Lock()
		ev.process(ob)
		ob.mu.Unlock()
	}
}

func (ob *OrderBook) Submit(o *order.Order) []Trade {
	done := make(chan []Trade, 1)
	ob.processChan <- &Submit{Order: o, Done: done}
	return <-done
}

func (ob *OrderBook) Cancel(id string) bool {
	done := make(chan bool, 1)
	ob.processChan <- &Cancel{ID: id, Done: done}
	return <-done
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

	if len(trades) > 0 {
		ob.lastPrice = trades[len(trades)-1].Price
		for _, t := range trades {
			if ob.OnTrade != nil {
				ob.OnTrade(t)
			}
		}
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

func (ob *OrderBook) findBestBidNoLock() float64 {
	if len(ob.bids) == 0 {
		return 0
	}
	max := 0.0
	for p := range ob.bids {
		if p > max {
			max = p
		}
	}
	return max
}

func (ob *OrderBook) findBestAskNoLock() float64 {
	if len(ob.asks) == 0 {
		return 0
	}
	min := math.MaxFloat64
	for p := range ob.asks {
		if p < min {
			min = p
		}
	}
	return min
}

func (ob *OrderBook) addToBook(o *order.Order) {
	m := ob.bids
	if o.Side == order.Sell {
		m = ob.asks
	}
	l, ok := m[o.Price]
	if !ok {
		l = list.New()
		m[o.Price] = l
	}
	elem := l.PushBack(o)
	ob.orders[o.ID] = elem

	if o.Side == order.Buy {
		if ob.bestBid < o.Price || ob.bestBid == 0 {
			ob.bestBid = o.Price
		}
	} else {
		if ob.bestAsk == 0 || ob.bestAsk > o.Price {
			ob.bestAsk = o.Price
		}
	}
}

func (ob *OrderBook) matchMarket(o *order.Order) []Trade {
	trades := []Trade{}
	for o.Quantity > 0 {
		oppPrice := ob.getBestOppositeNoLock(o.Side)
		if oppPrice == 0 {
			break
		}
		trades = append(trades, ob.matchAtPrice(o, oppPrice)...)
	}
	return trades
}

func (ob *OrderBook) matchLimit(o *order.Order) []Trade {
	trades := []Trade{}
	for o.Quantity > 0 {
		oppPrice := ob.getBestOppositeNoLock(o.Side)
		if oppPrice == 0 {
			break
		}
		if (o.Side == order.Buy && oppPrice > o.Price) || (o.Side == order.Sell && oppPrice < o.Price) {
			break
		}
		trades = append(trades, ob.matchAtPrice(o, oppPrice)...)
	}
	return trades
}

func (ob *OrderBook) getBestOppositeNoLock(side order.Side) float64 {
	if side == order.Buy {
		return ob.bestAsk
	}
	return ob.bestBid
}

func (ob *OrderBook) getBestBid() float64 {
	ob.mu.RLock()
	defer ob.mu.RUnlock()
	return ob.bestBid
}

func (ob *OrderBook) getBestAsk() float64 {
	ob.mu.RLock()
	defer ob.mu.RUnlock()
	return ob.bestAsk
}

func (ob *OrderBook) matchAtPrice(o *order.Order, price float64) []Trade {
	opposite := ob.asks
	if o.Side == order.Buy {
		opposite = ob.asks
	} else {
		opposite = ob.bids
	}
	l, ok := opposite[price]
	if !ok || l.Len() == 0 {
		return nil
	}
	trades := []Trade{}
	for o.Quantity > 0 && l.Len() > 0 {
		frontElem := l.Front()
		front := frontElem.Value.(*order.Order)
		fillQty := math.Min(o.Quantity, front.Quantity)
		o.Quantity -= fillQty
		front.Quantity -= fillQty
		trade := Trade{
			Price:    price,
			Quantity: fillQty,
		}
		if o.Side == order.Buy {
			trade.BuyerID = o.AgentID
			trade.SellerID = front.AgentID
			trade.BuyerOrderID = o.ID
			trade.SellerOrderID = front.ID
		} else {
			trade.BuyerID = front.AgentID
			trade.SellerID = o.AgentID
			trade.BuyerOrderID = front.ID
			trade.SellerOrderID = o.ID
		}
		trades = append(trades, trade)
		if front.Quantity <= 0 {
			l.Remove(frontElem)
			delete(ob.orders, front.ID)
		}
	}
	if l.Len() == 0 {
		delete(opposite, price)
		if o.Side == order.Buy {
			if price == ob.bestAsk {
				ob.bestAsk = ob.findBestAskNoLock()
			}
		} else {
			if price == ob.bestBid {
				ob.bestBid = ob.findBestBidNoLock()
			}
		}
	}
	return trades
}

func (ob *OrderBook) GetLastPrice() float64 {
	ob.mu.RLock()
	defer ob.mu.RUnlock()
	return ob.lastPrice
}

func (ob *OrderBook) GetBestBid() float64 {
	ob.mu.RLock()
	defer ob.mu.RUnlock()
	return ob.bestBid
}

func (ob *OrderBook) GetBestAsk() float64 {
	ob.mu.RLock()
	defer ob.mu.RUnlock()
	return ob.bestAsk
}