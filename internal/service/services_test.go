package service

import (
	"context"
	"cosmic-go/internal/domain"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestService(t *testing.T) {

	t.Run("allocate batch",
		func(t *testing.T) {
			ctx := context.Background()

			product := domain.Product{SKU: "SMALL-TABLE"}
			batch := domain.NewBatchWithoutETA("batch-001", product, 100)

			repo := NewFakeRepositoryWithBatch([]domain.Batch{*batch})
			svc := NewService(repo)

			line := &domain.OrderLine{
				Product:  product,
				Quantity: 10,
			}

			batchRef, err := svc.Allocate(ctx, line)
			assert.NoError(t, err)
			assert.Equal(t, "batch-001", batchRef)
		})

	t.Run("error for invalid sku on allocate",
		func(t *testing.T) {
			ctx := context.Background()

			product := domain.Product{SKU: "SMALL-TABLE"}
			batch := domain.NewBatchWithoutETA("batch-001", product, 100)

			repo := NewFakeRepositoryWithBatch([]domain.Batch{*batch})
			svc := NewService(repo)

			line := &domain.OrderLine{
				Product:  domain.Product{SKU: "ANOTHER-SKU"},
				Quantity: 10,
			}

			_, err := svc.Allocate(ctx, line)
			assert.Error(t, err, ErrInvalidSku)
		})

	t.Run("add batch",
		func(t *testing.T) {
			ctx := context.Background()

			batch := domain.NewBatchWithoutETA(
				"batch-001",
				domain.Product{SKU: "SMALL-TABLE"},
				100)

			repo := NewFakeRepository()
			svc := NewService(repo)

			err := svc.AddBatch(ctx, batch)
			assert.NoError(t, err)

			persistedBatch, err := repo.GetBatchByReference(ctx, "batch-001")
			assert.NoError(t, err)

			assert.Equal(t, persistedBatch, batch)
		})

	t.Run("deallocate invalid orderline from batch with wrong sku",
		func(t *testing.T) {
			ctx := context.Background()

			product := domain.Product{SKU: "SMALL-TABLE"}
			orderID := "order-001"

			batch := domain.NewBatchWithoutETA("batch-001", product, 100)
			batch.Allocations = map[string]domain.OrderLine{
				product.SKU: {
					Product:  product,
					Quantity: 10,
					OrderId:  orderID,
				},
			}

			repo := NewFakeRepositoryWithBatch([]domain.Batch{*batch})
			svc := NewService(repo)

			invalidProductSKU := "INVALID_SKU"
			err := svc.Deallocate(ctx, orderID, invalidProductSKU)
			assert.ErrorIs(t, err, ErrInvalidSku)

			_, ok := batch.Allocations[product.SKU]
			assert.True(t, ok)
		})

	t.Run("deallocate invalid orderline from batch with wrong orderid",
		func(t *testing.T) {
			ctx := context.Background()

			product := domain.Product{SKU: "SMALL-TABLE"}
			orderID := "order-001"

			batch := domain.NewBatchWithoutETA("batch-001", product, 100)
			batch.Allocations = map[string]domain.OrderLine{
				product.SKU: {
					Product:  product,
					Quantity: 10,
					OrderId:  orderID,
				},
			}

			repo := NewFakeRepositoryWithBatch([]domain.Batch{*batch})
			svc := NewService(repo)

			invalidOrderId := "INVALID_ORDERID"
			err := svc.Deallocate(ctx, invalidOrderId, product.SKU)
			assert.ErrorIs(t, err, ErrInvalidOrderID)

			_, ok := batch.Allocations[product.SKU]
			assert.True(t, ok)
		})

	t.Run("deallocate valid orderline from batch",
		func(t *testing.T) {
			ctx := context.Background()

			product := domain.Product{SKU: "SMALL-TABLE"}
			orderID := "order-001"

			batch := domain.NewBatchWithoutETA("batch-001", product, 100)
			batch.Allocations = map[string]domain.OrderLine{
				product.SKU: {
					Product:  product,
					Quantity: 10,
					OrderId:  orderID,
				},
			}

			repo := NewFakeRepositoryWithBatch([]domain.Batch{*batch})
			svc := NewService(repo)

			err := svc.Deallocate(ctx, orderID, product.SKU)
			assert.NoError(t, err)

			_, ok := batch.Allocations[product.SKU]
			assert.False(t, ok)
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

func (r *FakeRepository) AddBatch(_ context.Context, batch *domain.Batch) error {
	r.batches[batch.Reference] = batch
	return nil
}

func (r *FakeRepository) GetBatchByReference(_ context.Context, ref string) (*domain.Batch, error) {
	return r.batches[ref], nil
}

func (r *FakeRepository) ListBatches(_ context.Context) ([]*domain.Batch, error) {
	batches := make([]*domain.Batch, 0, len(r.batches))
	for _, batch := range r.batches {
		batches = append(batches, batch)
	}
	return batches, nil
}

func (r *FakeRepository) GetBatchBySku(_ context.Context, sku string) (*domain.Batch, error) {
	for _, batch := range r.batches {
		if batch.Product.SKU == sku {
			return batch, nil
		}
	}
	return nil, errors.New("sku not found in the fake in memory database")
}

func (r *FakeRepository) UpdateBatch(
	_ context.Context,
	existingB *domain.Batch,
	updatedB *domain.Batch,
) error {
	r.batches[existingB.Reference] = updatedB
	return nil
}
