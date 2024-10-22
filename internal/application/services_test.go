package application

import (
	"context"
	"os"
	"testing"
	"time"

	"cosmic-go/internal/domain"
	"cosmic-go/internal/infra/events/publisher"
	unitofwork "cosmic-go/internal/uow"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}

func TestService(t *testing.T) {
	t.Run("allocate batch",
		func(t *testing.T) {
			ctx := context.Background()

			sku := "SMALL-TABLE"
			batch := domain.NewBatch("batch-001", sku, 100, nil)
			product := domain.NewProduct(sku, []*domain.Batch{batch}, 0)

			ep := publisher.NewEventPublisher()

			repo := NewFakeRepository()
			uow := NewFakeUnitOfWork(repo, ep)
			svc := NewService(uow)

			err := svc.AddProduct(ctx, product)
			assert.NoError(t, err)

			line := &domain.OrderLine{
				SKU:      sku,
				Quantity: 10,
			}

			batchRef, err := svc.Allocate(ctx, line)
			assert.NoError(t, err)
			assert.Equal(t, "batch-001", batchRef)
		})

	t.Run("error for invalid sku on allocate",
		func(t *testing.T) {
			ctx := context.Background()

			ep := publisher.NewEventPublisher()

			repo := NewFakeRepository()
			uow := NewFakeUnitOfWork(repo, ep)
			svc := NewService(uow)

			sku := "SMALL-TABLE"
			batch := domain.NewBatch("batch-001", sku, 100, nil)
			product := domain.NewProduct(sku, []*domain.Batch{batch}, 0)

			err := svc.AddProduct(ctx, product)
			assert.NoError(t, err)

			line := &domain.OrderLine{
				SKU:      "ANOTHER-SKU",
				Quantity: 10,
			}

			_, err = svc.Allocate(ctx, line)
			assert.Error(t, err, ErrProductNotFound)
		})

	t.Run("add product",
		func(t *testing.T) {
			ctx := context.Background()

			ep := publisher.NewEventPublisher()

			repo := NewFakeRepository()
			uow := NewFakeUnitOfWork(repo, ep)
			svc := NewService(uow)

			sku := "SMALL-TABLE"
			batch := domain.NewBatch("batch-001", sku, 100, nil)
			product := domain.NewProduct(sku, []*domain.Batch{batch}, 0)

			err := svc.AddProduct(ctx, product)
			assert.NoError(t, err)

			persistedProduct, err := repo.GetProduct(ctx, sku)
			assert.NoError(t, err)

			assert.Equal(t, persistedProduct, product)
		})

	t.Run("deallocate invalid orderline from batch with wrong sku",
		func(t *testing.T) {
			ctx := context.Background()

			ep := publisher.NewEventPublisher()

			repo := NewFakeRepository()
			uow := NewFakeUnitOfWork(repo, ep)
			svc := NewService(uow)

			sku := "SMALL-TABLE"
			batch := domain.NewBatch("batch-001", sku, 100, nil)

			var orderID domain.OrderID = "order-001"

			batch.Allocations = map[domain.OrderID]domain.OrderLine{
				orderID: {
					SKU:      sku,
					Quantity: 10,
					OrderId:  orderID,
				},
			}

			product := domain.NewProduct(sku, []*domain.Batch{batch}, 0)
			err := svc.AddProduct(ctx, product)
			assert.NoError(t, err)

			invalidProductSKU := "INVALID_SKU"
			err = svc.Deallocate(ctx, orderID, invalidProductSKU)
			assert.ErrorIs(t, err, ErrProductNotFound)

			_, ok := batch.Allocations[orderID]
			assert.True(t, ok)
		})

	t.Run("deallocate invalid orderline from batch with wrong orderid",
		func(t *testing.T) {
			ctx := context.Background()

			ep := publisher.NewEventPublisher()

			repo := NewFakeRepository()
			uow := NewFakeUnitOfWork(repo, ep)
			svc := NewService(uow)

			var orderID domain.OrderID = "order-001"

			sku := "SMALL-TABLE"
			batch := domain.NewBatch("batch-001", sku, 100, nil)
			product := domain.NewProduct(sku, []*domain.Batch{batch}, 0)

			batch.Allocations = map[domain.OrderID]domain.OrderLine{
				orderID: {
					SKU:      sku,
					Quantity: 10,
					OrderId:  orderID,
				},
			}

			err := svc.AddProduct(ctx, product)
			assert.NoError(t, err)

			var invalidOrderId domain.OrderID = "INVALID_ORDERID"
			err = svc.Deallocate(ctx, invalidOrderId, product.SKU)
			assert.ErrorIs(t, err, ErrInvalidOrderID)

			_, ok := batch.Allocations[orderID]
			assert.True(t, ok)
		})

	t.Run("deallocate valid orderline from batch",
		func(t *testing.T) {
			ctx := context.Background()

			var orderID domain.OrderID = "order-001"

			sku := "SMALL-TABLE"
			batch := domain.NewBatch("batch-001", sku, 100, nil)
			product := domain.NewProduct(sku, []*domain.Batch{batch}, 0)

			batch.Allocations = map[domain.OrderID]domain.OrderLine{
				orderID: {
					SKU:      sku,
					Quantity: 10,
					OrderId:  orderID,
				},
			}

			ep := publisher.NewEventPublisher()

			repo := NewFakeRepository()
			uow := NewFakeUnitOfWork(repo, ep)
			svc := NewService(uow)

			err := svc.AddProduct(ctx, product)
			assert.NoError(t, err)

			err = svc.Deallocate(ctx, orderID, sku)
			assert.NoError(t, err)

			_, ok := batch.Allocations[orderID]
			assert.False(t, ok)
		})

	t.Run("reallocate allocated orderline",
		func(t *testing.T) {
			ep := publisher.NewEventPublisher()

			repo := NewFakeRepository()
			uow := NewFakeUnitOfWork(repo, ep)
			svc := NewService(uow)

			ctx := context.Background()

			sku := "SMALL-TABLE"
			otherBatch := domain.NewBatch("batch-999", sku, 100, nil)

			orderLine := &domain.OrderLine{
				SKU:      sku,
				Quantity: 25,
				OrderId:  "order-001",
			}
			eta := time.Now().Add(time.Hour * 48)
			existingBatch := domain.NewBatch("batch-001", sku, 100, &eta)
			err := existingBatch.Allocate(orderLine)
			assert.NoError(t, err)

			product := domain.NewProduct(
				sku,
				[]*domain.Batch{
					existingBatch,
					otherBatch,
				},
				0)

			err = svc.AddProduct(ctx, product)
			assert.NoError(t, err)

			updatedBatchRef, err := svc.Reallocate(ctx, orderLine)
			assert.NoError(t, err)

			assert.Equal(t, otherBatch.Reference, updatedBatchRef)
			assert.Equal(t, existingBatch.AvailableQuantity(), existingBatch.PurchasedQuantity)
		})

	t.Run("out of stock creates an event for external services",
		func(t *testing.T) {
			eventHandler := &MockEventHandler{}
			ep := publisher.NewEventPublisher()

			repo := NewFakeRepository()
			uow := NewFakeUnitOfWork(repo, ep)
			svc := NewService(uow)

			ctx := context.Background()

			sku := "SMALL-TABLE"
			batch := domain.NewBatch("batch-001", sku, 10, nil)
			product := domain.NewProduct(sku, []*domain.Batch{batch}, 0)

			event := &domain.OutOfStockEvent{
				SKU: "SMALL-TABLE",
			}
			ep.RegisterHandler(event, eventHandler)

			err := svc.AddProduct(ctx, product)
			assert.NoError(t, err)

			_, err = svc.Allocate(ctx, &domain.OrderLine{
				SKU:      sku,
				Quantity: 15,
			})
			assert.ErrorIs(t, err, domain.ErrOutOfStock)
			assert.Equal(t, 1, len(eventHandler.ReceivedEvents))
			assert.Equal(t, eventHandler.ReceivedEvents[0], event)
		})
}

type FakeRepository struct {
	products map[string]*domain.Product
}

func NewFakeRepository() *FakeRepository {
	return &FakeRepository{
		products: make(map[string]*domain.Product),
	}
}

func (r *FakeRepository) AddProduct(_ context.Context, p *domain.Product) error {
	r.products[p.SKU] = p
	return nil
}

func (r *FakeRepository) GetProduct(_ context.Context, sku string) (*domain.Product, error) {
	for _, p := range r.products {
		if p.SKU == sku {
			return r.products[p.SKU], nil
		}
	}
	return nil, domain.ErrProductNotFound
}

func (r *FakeRepository) UpdateProduct(
	_ context.Context,
	p *domain.Product,
) error {
	r.products[p.SKU] = p
	return nil
}

type FakeUoW struct {
	adapters unitofwork.Adapters
	events   []domain.Event
	ep       unitofwork.EventPublisher
}

func NewFakeUnitOfWork(repo unitofwork.Repository, ep unitofwork.EventPublisher) *FakeUoW {
	return &FakeUoW{
		adapters: unitofwork.Adapters{Repository: repo},
		ep:       ep,
	}
}

func (u *FakeUoW) Transact(ctx context.Context, txFunc func(_ unitofwork.Adapters) error) error {
	err := txFunc(u.adapters)
	u.dispatchEvents()
	return err
}

func (u *FakeUoW) AddEvent(event domain.Event) {
	u.events = append(u.events, event)
}

func (u *FakeUoW) dispatchEvents() {
	for _, event := range u.events {
		u.ep.Publish(event)
	}
	u.events = nil
}

type MockEventHandler struct {
	ReceivedEvents []domain.Event
}

func (m *MockEventHandler) Handle(event domain.Event) error {
	m.ReceivedEvents = append(m.ReceivedEvents, event)
	return nil
}
