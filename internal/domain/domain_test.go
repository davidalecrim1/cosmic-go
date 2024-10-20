package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDomainModel(t *testing.T) {
	t.Run("allocating to a batch reduces the available quantity",
		func(t *testing.T) {
			batch, order := createBatchAndOrderLine(t, "SMALL-TABLE", 20, 2)
			err := batch.Allocate(order)
			assert.NoError(t, err)
			assert.Equal(t, 18, batch.AvailableQuantity())
		})

	t.Run("can allocate if available greater than required",
		func(t *testing.T) {
			smallBatch, largeOrder := createBatchAndOrderLine(t, "ELEC-TRUMPET", 10, 20)
			err := smallBatch.Allocate(largeOrder)
			assert.Error(t, err)
		})

	t.Run("can allocate if available equal to required",
		func(t *testing.T) {
			smallBatch, largeOrder := createBatchAndOrderLine(t, "ELEC-TRUMPET", 10, 10)
			err := smallBatch.Allocate(largeOrder)
			assert.NoError(t, err)
			assert.Equal(t, 0, smallBatch.AvailableQuantity())
		})

	t.Run("cannot allocate if skus do not match",
		func(t *testing.T) {
			eta := time.Now()
			batch := NewBatch("batch-001", "ELEC-TRUMPET", 10, &eta)
			line := &OrderLine{"order-001", "SMALL-TABLE", 10}
			err := batch.Allocate(line)
			assert.Error(t, err)
		})

	t.Run("can only deallocate allocated order lines",
		func(t *testing.T) {
			eta := time.Now()
			batch := NewBatch("batch-001", "ELEC-TRUMPET", 10, &eta)
			unallocatedLine := &OrderLine{"order-001", "SMALL-TABLE", 1}
			err := batch.Deallocate(unallocatedLine)
			assert.Error(t, err)
			assert.Equal(t, 10, batch.AvailableQuantity())
		})

	t.Run("allocation is idempotent",
		func(t *testing.T) {
			batch, order := createBatchAndOrderLine(t, "ELEC-TRUMPET", 10, 2)

			err := batch.Allocate(order)
			assert.NoError(t, err)
			err = batch.Allocate(order)
			assert.NoError(t, err)

			assert.Equal(t, 8, batch.AvailableQuantity())
		})

	t.Run("prefers current stock batches to shipments",
		func(t *testing.T) {
			inStockBatch := NewBatch("in_stock_batch", "RETRO-CLOCK", 100, nil)
			eta := time.Now().Add(time.Hour * 24)
			shipmentBatch := NewBatch("shipment_batch", "RETRO-CLOCK", 100, &eta)
			line := &OrderLine{"order-001", "RETRO-CLOCK", 10}

			product := NewProduct("RETRO-CLOCK", []*Batch{inStockBatch, shipmentBatch}, 0)
			_, err := product.Allocate(line)
			assert.NoError(t, err)

			assert.Equal(t, 90, inStockBatch.AvailableQuantity())
			assert.Equal(t, 100, shipmentBatch.AvailableQuantity())
		})

	t.Run("prefers earliest batch to later batches",
		func(t *testing.T) {
			earlistEta := time.Now()
			earliest := NewBatch("earliest_batch", "RETRO-CLOCK", 100, &earlistEta)

			mediumEta := time.Now().Add(time.Hour * 24)
			medium := NewBatch("medium_batch", "RETRO-CLOCK", 100, &mediumEta)

			laterEta := time.Now().Add(time.Hour * 48)
			later := NewBatch("later_batch", "RETRO-CLOCK", 100, &laterEta)

			line := &OrderLine{"order-001", "RETRO-CLOCK", 10}

			product := NewProduct("RETRO-CLOCK", []*Batch{earliest, medium, later}, 0)
			_, err := product.Allocate(line)
			assert.NoError(t, err)

			assert.Equal(t, 90, earliest.AvailableQuantity())
			assert.Equal(t, 100, medium.AvailableQuantity())
			assert.Equal(t, 100, later.AvailableQuantity())
		})

	t.Run("out of stock error if cannot allocate",
		func(t *testing.T) {
			batch, line := createBatchAndOrderLine(t, "ELEC-TRUMPET", 10, 10)
			anotherLine := &OrderLine{"order-001", "ELEC-TRUMPET", 10}

			product := NewProduct("ELEC-TRUMPET", []*Batch{batch}, 0)
			_, err := product.Allocate(line)
			assert.NoError(t, err, "should allocate as expected")

			_, err = product.Allocate(anotherLine)
			assert.ErrorIs(t, err, ErrOutOfStock, "make sure we are out of stock")
		})
}

func createBatchAndOrderLine(t *testing.T, sku string, batchQty, lineQty int) (*Batch, *OrderLine) {
	t.Helper()
	eta := time.Now()
	batch := NewBatch("batch-001", sku, batchQty, &eta)
	line := OrderLine{"order-001", sku, lineQty}
	return batch, &line
}
