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

type FakeRepository struct {
	batches map[string]*Batch
}

func NewFakeRepository() *FakeRepository {
	return &FakeRepository{
		batches: make(map[string]*Batch),
	}
}

func NewFakeRepositoryWithBatch(batches []Batch) *FakeRepository {
	repo := &FakeRepository{
		batches: make(map[string]*Batch),
	}

	for _, batch := range batches {
		repo.batches[batch.Reference] = &batch
	}

	return repo
}

func (r *FakeRepository) Add(batch *Batch) error {
	r.batches[string(batch.Reference)] = batch
	return nil
}

func (r *FakeRepository) Get(reference string) (*Batch, error) {
	return r.batches[reference], nil
}

func (r *FakeRepository) List() ([]*Batch, error) {
	batches := make([]*Batch, 0, len(r.batches))
	for _, batch := range r.batches {
		batches = append(batches, batch)
	}
	return batches, nil
}
