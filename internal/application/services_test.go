package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"cosmic-go/internal/domain"

	"github.com/stretchr/testify/assert"
)

// mode the uow and fake repo here globally

func TestService(t *testing.T) {
	t.Run("allocate batch",
		func(t *testing.T) {
			ctx := context.Background()

			product := domain.Product{SKU: "SMALL-TABLE"}
			batch := domain.NewBatch("batch-001", product, 100, nil)

			repo := NewFakeRepository()
			uow := NewFakeUnitOfWorkFromRepository(repo)
			svc := NewService(uow)
			err := svc.AddBatch(ctx, batch)
			assert.NoError(t, err)

			line := &domain.OrderLine{
				Product:  product,
				Quantity: 10,
			}

			batchRef, err := svc.Allocate(ctx, line)
			assert.NoError(t, err)
			assert.Equal(t, "batch-001", batchRef)

			updatedBatch, err := repo.GetBatchByReference(ctx, "batch-001")
			assert.Equal(t, 90, updatedBatch.AvailableQuantity())
			assert.NoError(t, err)
		})

	t.Run("error for invalid sku on allocate",
		func(t *testing.T) {
			ctx := context.Background()

			product := domain.Product{SKU: "SMALL-TABLE"}
			batch := domain.NewBatch("batch-001", product, 100, nil)

			repo := NewFakeRepository()
			uow := NewFakeUnitOfWorkFromRepository(repo)
			svc := NewService(uow)
			err := svc.AddBatch(ctx, batch)
			assert.NoError(t, err)

			line := &domain.OrderLine{
				Product:  domain.Product{SKU: "ANOTHER-SKU"},
				Quantity: 10,
			}

			_, err = svc.Allocate(ctx, line)
			assert.Error(t, err, ErrInvalidSku)
		})

	t.Run("add batch",
		func(t *testing.T) {
			ctx := context.Background()

			batch := domain.NewBatch(
				"batch-001",
				domain.Product{SKU: "SMALL-TABLE"},
				100,
				nil,
			)

			repo := NewFakeRepository()
			uow := NewFakeUnitOfWorkFromRepository(repo)
			svc := NewService(uow)

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
			var orderID domain.OrderID = "order-001"

			batch := domain.NewBatch("batch-001", product, 100, nil)
			batch.Allocations = map[domain.OrderID]domain.OrderLine{
				orderID: {
					Product:  product,
					Quantity: 10,
					OrderId:  orderID,
				},
			}

			repo := NewFakeRepository()
			uow := NewFakeUnitOfWorkFromRepository(repo)
			svc := NewService(uow)
			err := svc.AddBatch(ctx, batch)
			assert.NoError(t, err)

			invalidProductSKU := "INVALID_SKU"
			err = svc.Deallocate(ctx, string(orderID), invalidProductSKU)
			assert.ErrorIs(t, err, ErrInvalidSku)

			_, ok := batch.Allocations[orderID]
			assert.True(t, ok)
		})

	t.Run("deallocate invalid orderline from batch with wrong orderid",
		func(t *testing.T) {
			ctx := context.Background()

			product := domain.Product{SKU: "SMALL-TABLE"}
			var orderID domain.OrderID = "order-001"

			batch := domain.NewBatch("batch-001", product, 100, nil)
			batch.Allocations = map[domain.OrderID]domain.OrderLine{
				orderID: {
					Product:  product,
					Quantity: 10,
					OrderId:  orderID,
				},
			}

			repo := NewFakeRepository()
			uow := NewFakeUnitOfWorkFromRepository(repo)
			svc := NewService(uow)
			err := svc.AddBatch(ctx, batch)
			assert.NoError(t, err)

			invalidOrderId := "INVALID_ORDERID"
			err = svc.Deallocate(ctx, invalidOrderId, product.SKU)
			assert.ErrorIs(t, err, ErrInvalidOrderID)

			_, ok := batch.Allocations[orderID]
			assert.True(t, ok)
		})

	t.Run("deallocate valid orderline from batch",
		func(t *testing.T) {
			ctx := context.Background()

			product := domain.Product{SKU: "SMALL-TABLE"}
			var orderID domain.OrderID = "order-001"

			batch := domain.NewBatch("batch-001", product, 100, nil)
			batch.Allocations = map[domain.OrderID]domain.OrderLine{
				orderID: {
					Product:  product,
					Quantity: 10,
					OrderId:  orderID,
				},
			}

			repo := NewFakeRepository()
			uow := NewFakeUnitOfWorkFromRepository(repo)
			svc := NewService(uow)

			err := svc.AddBatch(ctx, batch)
			assert.NoError(t, err)

			err = svc.Deallocate(ctx, string(orderID), product.SKU)
			assert.NoError(t, err)

			_, ok := batch.Allocations[orderID]
			assert.False(t, ok)
		})

	t.Run("reallocate allocated orderline",
		func(t *testing.T) {
			repo := NewFakeRepository()
			uow := NewFakeUnitOfWorkFromRepository(repo)
			svc := NewService(uow)

			ctx := context.Background()

			product := domain.Product{SKU: "SMALL-TABLE"}
			orderLine := &domain.OrderLine{
				Product:  product,
				Quantity: 25,
				OrderId:  "order-001",
			}

			eta := time.Now().Add(time.Hour * 48)
			existingBatch := domain.NewBatch("batch-001", product, 100, &eta)
			err := existingBatch.Allocate(orderLine)
			assert.NoError(t, err)

			err = svc.AddBatch(ctx, existingBatch)
			assert.NoError(t, err)

			otherBatch := domain.NewBatch("batch-002", product, 50, nil)

			err = svc.AddBatch(ctx, otherBatch)
			assert.NoError(t, err)

			updatedBatchRef, err := svc.Reallocate(ctx, orderLine)
			assert.NoError(t, err)

			assert.Equal(t, otherBatch.Reference, updatedBatchRef)
			assert.Equal(t, existingBatch.AvailableQuantity(), existingBatch.PurchasedQuantity)
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
