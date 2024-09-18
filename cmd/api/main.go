package main

import (
	"errors"
	"time"
)

var (
	ErrOrderLinesOverBatch = errors.New("the order lines are over the available quantity in batch")
	ErrProductSkuMismatch  = errors.New("product sku mismatch in batch")
)

type Product struct {
	SKU string
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

func (sb *Batch) Allocate(o *Order) error {
	total := 0
	for _, ol := range o.Lines {
		if sb.Product.SKU != ol.Product.SKU {
			return ErrProductSkuMismatch
		}
		total += ol.Quantity
	}

	if sb.AvailableQuantity < total {
		return ErrOrderLinesOverBatch
	}

	sb.AvailableQuantity -= total
	return nil
	// TODO: Validate multiple allocations from the same orderline
}
