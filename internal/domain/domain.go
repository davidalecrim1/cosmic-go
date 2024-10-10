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

type OrderID string

type OrderLine struct {
	Product  Product
	Quantity int
	OrderId  OrderID
}

type Batch struct {
	Reference         string
	Product           Product
	PurchasedQuantity int
	eta               *time.Time
	Allocations       map[OrderID]OrderLine
}

func NewBatch(ref string, product Product, quantity int, eta *time.Time) *Batch {
	return &Batch{
		Reference:         ref,
		Product:           product,
		PurchasedQuantity: quantity,
		eta:               eta,
		Allocations:       make(map[OrderID]OrderLine),
	}
}

func NewBatchWithoutETA(ref string, product Product, quantity int) *Batch {
	return &Batch{
		Reference:         ref,
		Product:           product,
		PurchasedQuantity: quantity,
		Allocations:       make(map[OrderID]OrderLine),
	}
}

func (b *Batch) GetETA() *time.Time {
	return b.eta
}

func (b *Batch) SetETA(value *time.Time) {
	if value != nil {
		b.eta = value
	}

	b.eta = nil
}

func (b *Batch) Allocate(line *OrderLine) error {
	if b.Allocations == nil {
		b.Allocations = make(map[OrderID]OrderLine)
	}

	if b.Product.SKU != line.Product.SKU {
		return ErrProductSkuMismatch
	}

	if b.AvailableQuantity() < line.Quantity {
		return ErrOrderLinesOverBatch
	}

	b.Allocations[line.OrderId] = *line
	return nil
}

func (b *Batch) Deallocate(line *OrderLine) error {
	if b.Product.SKU != line.Product.SKU {
		return ErrCannotDeallocateUnallocatedOrderLine
	}

	if b.AvailableQuantity() < line.Quantity {
		return ErrOrderLinesOverBatch
	}

	delete(b.Allocations, line.OrderId)
	return nil
}

func (b *Batch) AvailableQuantity() int {
	allocated := 0
	for _, line := range b.Allocations {
		allocated += line.Quantity
	}

	return b.PurchasedQuantity - allocated
}

func Allocate(ol *OrderLine, bt []*Batch) (reference string, err error) {
	sortBasedOnEarliestETA := func(i, j int) bool {
		// no ETA (nil) is the earliest
		if bt[i].eta == nil {
			return true
		}

		if bt[j].eta == nil {
			return false
		}

		return bt[i].eta.Before(*bt[j].eta)
	}

	sort.Slice(bt, sortBasedOnEarliestETA)
	return bt[0].Reference, bt[0].Allocate(ol)
}
