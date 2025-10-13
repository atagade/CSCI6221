package orderbook

import (
	"testing"
	"time"

	"cda-simulator/order"
)

func TestMatchLimit(t *testing.T) {
	ob := New()

	o1 := &order.Order{ID: "1", AgentID: "a", Side: order.Sell, Type: order.Limit, Price: 100, Quantity: 10, Stock: "GOOG", Time: time.Now()}
	ob.Submit(o1)

	o2 := &order.Order{ID: "2", AgentID: "b", Side: order.Buy, Type: order.Limit, Price: 101, Quantity: 5, Stock: "GOOG", Time: time.Now()}
	trades := ob.Submit(o2)

	if len(trades) != 1 {
		// Expect exactly one match for the limit orders
		t.Errorf("expected 1 trade, got %d", len(trades))
	}
	if trades[0].Quantity != 5 || trades[0].Price != 100 {
		// Verify trade price and filled quantity are correct (FIFO behavior)
		t.Error("trade mismatch")
	}
	if ob.GetLastPrice() != 100 {
		t.Error("last price not updated")
	}
	if ob.GetBestAsk() != 100 {
		t.Error("best ask not updated")
	}
}

func TestMatchMarket(t *testing.T) {
	ob := New()

	o1 := &order.Order{ID: "1", AgentID: "a", Side: order.Sell, Type: order.Limit, Price: 100, Quantity: 10, Stock: "GOOG", Time: time.Now()}
	ob.Submit(o1)

	o2 := &order.Order{ID: "2", AgentID: "b", Side: order.Buy, Type: order.Market, Quantity: 15, Stock: "GOOG", Time: time.Now()}
	trades := ob.Submit(o2)

	if len(trades) != 1 {
		// Market buy against single limit sell should generate one trade with partial/full fill
		t.Errorf("expected 1 trade, got %d", len(trades))
	}
	if trades[0].Quantity != 10 {
		// Market buy of 15 should fill 10 at existing ask, remaining unfilled
		t.Error("should fill 10, rest canceled for IOC")
	}
	if ob.GetBestAsk() != 0 {
		t.Error("book should be empty")
	}
}

func TestCancel(t *testing.T) {
	ob := New()

	o1 := &order.Order{ID: "1", AgentID: "a", Side: order.Sell, Type: order.Limit, Price: 100, Quantity: 10, Stock: "GOOG", Time: time.Now()}
	ob.Submit(o1)

	canceled := ob.Cancel("1")
	if !canceled {
		// Cancel must return true for an existing order
		t.Error("cancel failed")
	}

	o2 := &order.Order{ID: "2", AgentID: "b", Side: order.Buy, Type: order.Market, Quantity: 5, Stock: "GOOG", Time: time.Now()}
	trades := ob.Submit(o2)
	if len(trades) != 0 {
		t.Error("should no match after cancel")
	}
}

func TestPartialFill(t *testing.T) {
	ob := New()

	o1 := &order.Order{ID: "1", AgentID: "a", Side: order.Sell, Type: order.Limit, Price: 100, Quantity: 10, Stock: "GOOG", Time: time.Now()}
	ob.Submit(o1)

	o2 := &order.Order{ID: "2", AgentID: "b", Side: order.Buy, Type: order.Limit, Price: 105, Quantity: 7, Stock: "GOOG", Time: time.Now()}
	trades := ob.Submit(o2)

	if len(trades) != 1 {
		// Partial fill case: one trade for partial quantity
		t.Errorf("expected 1 trade, got %d", len(trades))
	}
	if trades[0].Quantity != 7 {
		t.Error("should partial fill 7")
	}
	if ob.GetBestAsk() != 100 {
		t.Error("level still exists")
	}
	// remaining qty in book: 3
	o3 := &order.Order{ID: "3", AgentID: "c", Side: order.Buy, Type: order.Market, Quantity: 5, Stock: "GOOG", Time: time.Now()}
	trades2 := ob.Submit(o3)
	if len(trades2) != 1 || trades2[0].Quantity != 3 {
		t.Error("should fill remaining 3")
	}
}

func TestBestPriceTracking(t *testing.T) {
	ob := New()

	o1 := &order.Order{ID: "1", AgentID: "a", Side: order.Buy, Type: order.Limit, Price: 99, Quantity: 5, Stock: "GOOG", Time: time.Now()}
	ob.Submit(o1)
	if ob.GetBestBid() != 99 {
		t.Error("best bid not set")
	}

	o2 := &order.Order{ID: "2", AgentID: "b", Side: order.Sell, Type: order.Limit, Price: 101, Quantity: 5, Stock: "GOOG", Time: time.Now()}
	ob.Submit(o2)
	if ob.GetBestAsk() != 101 {
		t.Error("best ask not set")
	}

	// Use price 100 (between bid and ask) to avoid immediate matching
	o3 := &order.Order{ID: "3", AgentID: "c", Side: order.Buy, Type: order.Limit, Price: 100, Quantity: 3, Stock: "GOOG", Time: time.Now()}
	ob.Submit(o3)
	if ob.GetBestBid() != 100 {
		t.Error("best bid not updated to higher")
	}

	ob.Cancel("1")
	if ob.GetBestBid() != 100 {
		t.Error("best bid unchanged after cancel non-best")
	}

	ob.Cancel("3")
	if ob.GetBestBid() != 0 {
		t.Error("best bid not reset after cancel best - expected 0 since no more bids")
	}
}