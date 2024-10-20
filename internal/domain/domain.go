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
	ErrOutOfStock                           = errors.New("out of stock")
)

type OrderID string

type OrderLine struct {
	OrderId  OrderID
	SKU      string
	Quantity int
}

type Batch struct {
	Reference         string
	SKU               string
	PurchasedQuantity int
	eta               *time.Time
	Allocations       map[OrderID]OrderLine
}

func NewBatch(ref string, sku string, quantity int, eta *time.Time) *Batch {
	return &Batch{
		Reference:         ref,
		SKU:               sku,
		PurchasedQuantity: quantity,
		eta:               eta,
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

	if b.SKU != line.SKU {
		return ErrProductSkuMismatch
	}

	if b.AvailableQuantity() < line.Quantity {
		return ErrOrderLinesOverBatch
	}

	b.Allocations[line.OrderId] = *line
	return nil
}

func (b *Batch) Deallocate(line *OrderLine) error {
	if b.SKU != line.SKU {
		return ErrCannotDeallocateUnallocatedOrderLine
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

func (b *Batch) isInStock() bool {
	return b.eta == nil
}

type Product struct {
	SKU       string
	Batches   []*Batch
	VersionId int
}

func NewProduct(sku string, batches []*Batch, versionId int) *Product {
	return &Product{
		SKU:       sku,
		Batches:   batches,
		VersionId: versionId,
	}
}

func (p *Product) Allocate(ol *OrderLine) (reference string, err error) {
	sortBasedOnEarliestETA := func(i, j int) bool {
		// no ETA (nil) is the earliest because the batch is in stock
		if p.Batches[i].isInStock() {
			return true
		}

		if p.Batches[j].isInStock() {
			return false
		}

		return p.Batches[i].eta.Before(*p.Batches[j].eta)
	}

	sort.Slice(p.Batches, sortBasedOnEarliestETA)

	for _, batch := range p.Batches {
		err = batch.Allocate(ol)
		if errors.Is(err, ErrOrderLinesOverBatch) {
			continue
		}

		if err != nil {
			return "", err
		}

		p.VersionId++
		return batch.Reference, nil
	}

	return "", ErrOutOfStock
}

func (p *Product) Deallocate(orderid OrderID) error {
	for _, b := range p.Batches {
		for _, allocOrderLine := range b.Allocations {
			if allocOrderLine.OrderId == OrderID(orderid) {
				return b.Deallocate(&allocOrderLine)
			}
		}
	}

	return ErrCannotDeallocateUnallocatedOrderLine
}
