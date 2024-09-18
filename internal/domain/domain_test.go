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
			batch.Allocate(order)
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
			batch := NewBatch("batch-001", Product{"ELEC-TRUMPET"}, 10, time.Now())
			line := &OrderLine{Product{"SMALL-TABLE"}, 10}
			err := batch.Allocate(line)
			assert.Error(t, err)
		})

	t.Run("can only deallocate allocated order lines",
		func(t *testing.T) {
			batch := NewBatch("batch-001", Product{"ELEC-TRUMPET"}, 10, time.Now())
			unallocatedLine := &OrderLine{Product{"SMALL-TABLE"}, 1}
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

	t.Run("test prefers current stock batches to shipments",
		func(t *testing.T) {
			inStockBatch := NewBatchWithoutETA("in_stock_batch", Product{"RETRO-CLOCK"}, 100)
			shipmentBatch := NewBatch("shipment_batch", Product{"RETRO-CLOCK"}, 100, time.Now().Add(time.Hour*24))
			line := &OrderLine{Product{"RETRO-CLOCK"}, 10}

			Allocate(line, []Batch{*inStockBatch, *shipmentBatch})
			assert.Equal(t, 90, inStockBatch.AvailableQuantity())
			assert.Equal(t, 100, shipmentBatch.AvailableQuantity())
		})

	t.Run("prefers earliest batch to later batches",
		func(t *testing.T) {
			earliest := NewBatch("earliest_batch", Product{"RETRO-CLOCK"}, 100, time.Now())
			medium := NewBatch("medium_batch", Product{"RETRO-CLOCK"}, 100, time.Now().Add(time.Hour*24))
			later := NewBatch("later_batch", Product{"RETRO-CLOCK"}, 100, time.Now().Add(time.Hour*48))

			line := &OrderLine{Product{"RETRO-CLOCK"}, 10}

			Allocate(line, []Batch{*later, *earliest, *medium})
			assert.Equal(t, 90, earliest.AvailableQuantity())
			assert.Equal(t, 100, medium.AvailableQuantity())
			assert.Equal(t, 100, later.AvailableQuantity())
		})

	t.Run("out of stock error if cannot allocate",
		func(t *testing.T) {
			batch, line := createBatchAndOrderLine(t, "ELEC-TRUMPET", 10, 10)
			anotherLine := &OrderLine{Product{"ELEC-TRUMPET"}, 2}

			Allocate(line, []Batch{*batch})
			_, err := Allocate(anotherLine, []Batch{*batch})
			assert.ErrorIs(t, err, ErrOrderLinesOverBatch)
		})
}

func createBatchAndOrderLine(t *testing.T, sku Reference, batchQty, lineQty int) (*Batch, *OrderLine) {
	t.Helper()
	batch := NewBatch("batch-001", Product{sku}, batchQty, time.Now())
	line := OrderLine{Product{sku}, lineQty}
	return batch, &line
}
