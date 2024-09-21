package domain

import (
	"errors"
	"sort"
	"time"
)

var (
	ErrOrderLinesOverBatch                  = errors.New("the order lines are over the available quantity in batch")
	ErrProductSkuMismatch                   = errors.New("product sku mismatch in batch")
	ErrCannotDeallocateUnallocatedOrderLine = errors.New("cannot deallocate unallocated order line")
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
	PurchasedQuantity int
	ETA               time.Time
	allocations       map[string]OrderLine
}

func NewBatch(ref string, product Product, quantity int, eta time.Time) *Batch {
	return &Batch{
		Reference:         ref,
		Product:           product,
		PurchasedQuantity: quantity,
		ETA:               eta,
		allocations:       make(map[string]OrderLine),
	}
}

func NewBatchWithoutETA(ref string, product Product, quantity int) *Batch {
	return &Batch{
		Reference:         ref,
		Product:           product,
		PurchasedQuantity: quantity,
		allocations:       make(map[string]OrderLine),
	}
}

func (sb *Batch) Allocate(line *OrderLine) error {
	if sb.allocations == nil {
		sb.allocations = make(map[string]OrderLine)
	}

	if sb.Product.SKU != line.Product.SKU {
		return ErrProductSkuMismatch
	}

	if sb.AvailableQuantity() < line.Quantity {
		return ErrOrderLinesOverBatch
	}

	sb.allocations[line.Product.SKU] = *line
	return nil
}

func (sb *Batch) Deallocate(line *OrderLine) error {
	if sb.Product.SKU != line.Product.SKU {
		return ErrCannotDeallocateUnallocatedOrderLine
	}

	if sb.AvailableQuantity() < line.Quantity {
		return ErrOrderLinesOverBatch
	}

	delete(sb.allocations, line.Product.SKU)
	return nil
}

func (sb *Batch) AvailableQuantity() int {
	allocated := 0
	for _, line := range sb.allocations {
		allocated += line.Quantity
	}

	return sb.PurchasedQuantity - allocated
}

func Allocate(ol *OrderLine, bt []*Batch) (string, error) {
	sortBasedOnEarliestETA := func(i, j int) bool {
		return bt[i].ETA.Before(bt[j].ETA)
	}

	sort.Slice(bt, sortBasedOnEarliestETA)
	return bt[0].Reference, bt[0].Allocate(ol)
}
