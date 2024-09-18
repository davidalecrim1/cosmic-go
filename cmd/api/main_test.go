package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDomain(t *testing.T) {
	t.Run("allocating to a batch reduces the available quantity",
		func(t *testing.T) {
			batch, order := createBatchAndOrder(t, "SMALL-TABLE", 20, 2)
			batch.Allocate(order)
			assert.Equal(t, 18, batch.AvailableQuantity)
		})

	t.Run("can allocate if available greater than required",
		func(t *testing.T) {
			smallBatch, largeOrder := createBatchAndOrder(t, "ELEC-TRUMPET", 10, 20)
			err := smallBatch.Allocate(largeOrder)
			assert.Error(t, err)
		})

	t.Run("can allocate if available equal to required",
		func(t *testing.T) {
			smallBatch, largeOrder := createBatchAndOrder(t, "ELEC-TRUMPET", 10, 10)
			err := smallBatch.Allocate(largeOrder)
			assert.NoError(t, err)
			assert.Equal(t, 0, smallBatch.AvailableQuantity)
		})

	t.Run("cannot allocate if skus do not match",
		func(t *testing.T) {
			batch := &Batch{"batch-001", Product{"ELEC-TRUMPET"}, 10, time.Now()}
			order := &Order{"order-ref", []OrderLine{{Product{"SMALL-TABLE"}, 10}}}
			err := batch.Allocate(order)
			assert.Error(t, err)
		})
}

func createBatchAndOrder(t *testing.T, sku string, batchQty, orderQty int) (*Batch, *Order) {
	t.Helper()
	batch := Batch{"batch-001", Product{sku}, batchQty, time.Now()}
	line := OrderLine{Product{sku}, orderQty}
	order := Order{"order-ref", []OrderLine{line}}
	return &batch, &order
}
