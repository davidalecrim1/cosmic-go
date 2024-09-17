package main

import "time"

type Product struct {
	sku string
}

type OrderLine struct {
	Product  Product
	Quantity int
}

type Order struct {
	Reference string
	Lines     []OrderLine
}

type Batch struct {
	Reference         string
	Product           Product
	AvailableQuantity int
	ETA               time.Time
}

func (sb *Batch) Allocate(o Order) {
	sb.AvailableQuantity -= o.Lines[0].Quantity
	// TODO: Validate multiple allocations from the same orderline
}
