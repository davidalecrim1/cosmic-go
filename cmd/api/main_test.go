package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDomain(t *testing.T) {
	t.Run("allocating to a batch reduces the available quantity",
		func(t *testing.T) {
			batch := Batch{"batch-001", Product{"SMALL-TABLE"}, 20, time.Now()}
			line := OrderLine{Product{"SMALL-TABLE"}, 2}
			order := Order{"order-ref", []OrderLine{line}}

			batch.Allocate(order)
			assert.Equal(t, 18, batch.AvailableQuantity)
		})
}
