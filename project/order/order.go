package order

import "time"

type Side string

const (
	Buy  Side = "buy"
	Sell Side = "sell"
)

type OrderType string

const (
	Limit OrderType = "limit"
	Market OrderType = "market"
)

type Order struct {
	ID       string
	Stock    string
	Side     Side
	Type     OrderType
	Price    float64
	Quantity float64
	Time     time.Time
	AgentID  string
}