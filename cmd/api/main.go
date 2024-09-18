package main

import (
	"errors"
	"time"
)

var (
	ErrOrderLinesOverBatch                  = errors.New("the order lines are over the available quantity in batch")
	ErrProductSkuMismatch                   = errors.New("product sku mismatch in batch")
	ErrCannotDeallocateUnallocatedOrderLine = errors.New("cannot deallocate unallocated order line")
)

type Reference string

type Product struct {
	SKU string
}

type OrderLine struct {
	Product  Product
	Quantity int
}

type Order struct {
	Reference Reference
	Lines     []OrderLine
}

type Batch struct {
	Reference         Reference
	Product           Product
	PurchasedQuantity int
	Allocations       map[Reference]OrderLine
	ETA               time.Time
}

func NewBatch(ref Reference, product Product, quantity int, eta time.Time) *Batch {
	return &Batch{
		Reference:         ref,
		Product:           product,
		PurchasedQuantity: quantity,
		Allocations:       make(map[Reference]OrderLine),
		ETA:               eta,
	}
}

func (sb *Batch) Allocate(o *Order) error {
	total := 0
	for _, line := range o.Lines {
		if sb.Product.SKU != line.Product.SKU {
			return ErrProductSkuMismatch
		}
		total += line.Quantity
	}

	if sb.PurchasedQuantity < total {
		return ErrOrderLinesOverBatch
	}

	for _, line := range o.Lines {
		sb.Allocations[o.Reference] = line
	}

	return nil
}

func (sb *Batch) Deallocate(o *Order) error {
	total := 0
	for _, line := range o.Lines {
		if sb.Product.SKU != line.Product.SKU {
			return ErrCannotDeallocateUnallocatedOrderLine
		}
		total += line.Quantity
	}
	if sb.PurchasedQuantity < total {
		return ErrOrderLinesOverBatch
	}

	delete(sb.Allocations, o.Reference)
	return nil
}

func (sb *Batch) AvailableQuantity() int {
	allocated := 0

	for _, line := range sb.Allocations {
		allocated += line.Quantity
	}

	return sb.PurchasedQuantity - allocated
}
