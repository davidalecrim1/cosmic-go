package service

import (
	"cosmic-go/internal/domain"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestService(t *testing.T) {
	t.Run("returns allocation",
		func(t *testing.T) {
			product := domain.Product{SKU: "SMALL-TABLE"}
			batch := domain.NewBatchWithoutETA("batch-001", product, 100)

			repo := NewFakeRepositoryWithBatch([]domain.Batch{*batch})
			svc := NewService(repo)

			line := &domain.OrderLine{
				Product:  product,
				Quantity: 10,
			}

			batchRef, err := svc.Allocate(line)
			assert.NoError(t, err)
			assert.Equal(t, "batch-001", batchRef)
		})

	t.Run("error for invalid sku",
		func(t *testing.T) {
			product := domain.Product{SKU: "SMALL-TABLE"}
			batch := domain.NewBatchWithoutETA("batch-001", product, 100)

			repo := NewFakeRepositoryWithBatch([]domain.Batch{*batch})
			svc := NewService(repo)

			line := &domain.OrderLine{
				Product:  domain.Product{SKU: "ANOTHER-SKU"},
				Quantity: 10,
			}

			_, err := svc.Allocate(line)
			assert.Error(t, err, ErrInvalidSku)
		})
}

type FakeRepository struct {
	batches map[string]*domain.Batch
}

func NewFakeRepository() *FakeRepository {
	return &FakeRepository{
		batches: make(map[string]*domain.Batch),
	}
}

func NewFakeRepositoryWithBatch(batches []domain.Batch) *FakeRepository {
	repo := &FakeRepository{
		batches: make(map[string]*domain.Batch),
	}

	for _, batch := range batches {
		repo.batches[batch.Reference] = &batch
	}

	return repo
}

func (r *FakeRepository) Add(batch *domain.Batch) error {
	r.batches[string(batch.Reference)] = batch
	return nil
}

func (r *FakeRepository) Get(reference string) (*domain.Batch, error) {
	return r.batches[reference], nil
}

func (r *FakeRepository) List() ([]*domain.Batch, error) {
	batches := make([]*domain.Batch, 0, len(r.batches))
	for _, batch := range r.batches {
		batches = append(batches, batch)
	}
	return batches, nil
}
