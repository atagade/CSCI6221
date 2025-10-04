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
