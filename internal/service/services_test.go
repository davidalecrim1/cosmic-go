package service

import (
	"cosmic-go/internal/domain"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestService(t *testing.T) {
	t.Run("allocate batch",
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

	t.Run("error for invalid sku on allocate",
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

	t.Run("add batch",
		func(t *testing.T) {
			batch := domain.NewBatchWithoutETA(
				"batch-001",
				domain.Product{SKU: "SMALL-TABLE"},
				100)

			repo := NewFakeRepository()
			svc := NewService(repo)

			err := svc.AddBatch(batch)
			assert.NoError(t, err)

			persistedBatch, err := repo.GetBatch("batch-001")
			assert.NoError(t, err)

			assert.Equal(t, persistedBatch, batch)
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

func (r *FakeRepository) AddBatch(batch *domain.Batch) error {
	r.batches[batch.Reference] = batch
	return nil
}

func (r *FakeRepository) GetBatch(reference string) (*domain.Batch, error) {
	return r.batches[reference], nil
}

func (r *FakeRepository) ListBatches() ([]*domain.Batch, error) {
	batches := make([]*domain.Batch, 0, len(r.batches))
	for _, batch := range r.batches {
		batches = append(batches, batch)
	}
	return batches, nil
}
