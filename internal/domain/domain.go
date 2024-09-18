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

type Reference string

type Product struct {
	SKU Reference
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
	ETA               time.Time
	allocations       map[Reference]OrderLine
}

func NewBatch(ref Reference, product Product, quantity int, eta time.Time) *Batch {
	return &Batch{
		Reference:         ref,
		Product:           product,
		PurchasedQuantity: quantity,
		ETA:               eta,
		allocations:       make(map[Reference]OrderLine),
	}
}

func NewBatchWithoutETA(ref Reference, product Product, quantity int) *Batch {
	return &Batch{
		Reference:         ref,
		Product:           product,
		PurchasedQuantity: quantity,
		allocations:       make(map[Reference]OrderLine),
	}
}

func (sb *Batch) Allocate(line *OrderLine) error {
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

func Allocate(ol *OrderLine, bt []Batch) (Reference, error) {
	sortBasedOnEarliestETA := func(i, j int) bool {
		return bt[i].ETA.Before(bt[j].ETA)
	}

	sort.Slice(bt, sortBasedOnEarliestETA)
	return bt[0].Reference, bt[0].Allocate(ol)
}
