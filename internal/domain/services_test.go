package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestService(t *testing.T) {
	t.Run("returns allocation",
		func(t *testing.T) {
			product := Product{SKU: "SMALL-TABLE"}
			batch := NewBatchWithoutETA("batch-001", product, 100)

			repo := NewFakeRepositoryWithBatch([]Batch{*batch})
			svc := NewService(repo)

			line := &OrderLine{product, 10}

			batchRef, err := svc.Allocate(line)
			assert.NoError(t, err)
			assert.Equal(t, "batch-001", batchRef)
		})

	t.Run("error for invalid sku",
		func(t *testing.T) {
			product := Product{SKU: "SMALL-TABLE"}
			batch := NewBatchWithoutETA("batch-001", product, 100)

			repo := NewFakeRepositoryWithBatch([]Batch{*batch})
			svc := NewService(repo)

			line := &OrderLine{Product{SKU: "ANOTHER-SKU"}, 10}

			_, err := svc.Allocate(line)
			assert.Error(t, err, ErrInvalidSku)
		})
}
