package application

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"cosmic-go/internal/domain"
	utils "cosmic-go/pkg/utils"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}

func TestService(t *testing.T) {
	t.Run("allocate batch",
		func(t *testing.T) {
			imp := NewFakeInternalMessagePublisher()

			repo := NewFakeRepository()
			uow := NewFakeUnitOfWork(repo, imp)
			svc := NewAllocationService(uow)

			imp.RegisterCommandHandler(&domain.CreateProduct{}, svc.AddProduct)

			eta := time.Now()
			command := &domain.CreateProduct{
				SKU: "SMALL-TABLE",
				Batches: []*domain.Batch{
					domain.NewBatch("batch-001", "SMALL-TABLE", 100, &eta),
				},
			}

			errChan := imp.PublishCommand(command)
			assert.True(t, utils.ErrChanIsEmpty(errChan))

			ctx := context.Background()
			product, err := repo.GetProduct(ctx, "SMALL-TABLE")
			assert.NoError(t, err)

			assert.Equal(t, product.Batches[0].Reference, "batch-001")
		})

	t.Run("error for invalid sku on allocate",
		func(t *testing.T) {
			imp := NewFakeInternalMessagePublisher()

			repo := NewFakeRepository()
			uow := NewFakeUnitOfWork(repo, imp)
			svc := NewAllocationService(uow)

			imp.RegisterCommandHandler(&domain.CreateProduct{}, svc.AddProduct)
			imp.RegisterCommandHandler(&domain.Allocate{}, svc.Allocate)

			createProductEvent := &domain.CreateProduct{
				SKU: "SMALL-TABLE",
				Batches: []*domain.Batch{
					domain.NewBatch("batch-001", "SMALL-TABLE", 100, nil),
				},
			}

			errChan := imp.PublishCommand(createProductEvent)
			assert.True(t, utils.ErrChanIsEmpty(errChan))

			allocationEvent := &domain.Allocate{
				OrderID:  "order-001",
				SKU:      "LARGE-TABLE",
				Quantity: 10,
			}

			errChan = imp.PublishCommand(allocationEvent)
			assert.NotNil(t, utils.ErrChanWithAny(errChan, ErrProductNotFound))
		})

	t.Run("add product",
		func(t *testing.T) {
			imp := NewFakeInternalMessagePublisher()

			repo := NewFakeRepository()
			uow := NewFakeUnitOfWork(repo, imp)
			svc := NewAllocationService(uow)

			imp.RegisterCommandHandler(&domain.CreateProduct{}, svc.AddProduct)

			sku := "SMALL-TABLE"
			createProductEvent := &domain.CreateProduct{
				SKU: sku,
				Batches: []*domain.Batch{
					domain.NewBatch("batch-001", sku, 100, nil),
				},
			}

			errChan := imp.PublishCommand(createProductEvent)
			assert.True(t, utils.ErrChanIsEmpty(errChan))

			ctx := context.Background()
			persistedProduct, err := repo.GetProduct(ctx, sku)
			assert.NoError(t, err)

			assert.Equal(t, persistedProduct.SKU, sku)
			assert.Equal(t, persistedProduct.Batches, createProductEvent.Batches)
		})

	t.Run("deallocate invalid orderline from batch with wrong sku",
		func(t *testing.T) {
			ctx := context.Background()

			imp := NewFakeInternalMessagePublisher()

			repo := NewFakeRepository()
			uow := NewFakeUnitOfWork(repo, imp)
			svc := NewAllocationService(uow)

			imp.RegisterCommandHandler(&domain.CreateProduct{}, svc.AddProduct)
			imp.RegisterCommandHandler(&domain.Allocate{}, svc.Allocate)
			imp.RegisterCommandHandler(&domain.Deallocate{}, svc.Deallocate)

			sku := "SMALL-TABLE"
			orderID := "order-001"
			createProductEvent := &domain.CreateProduct{
				SKU: sku,
				Batches: []*domain.Batch{
					domain.NewBatch("batch-001", sku, 100, nil),
				},
			}

			errChan := imp.PublishCommand(createProductEvent)
			assert.True(t, utils.ErrChanIsEmpty(errChan))

			allocateEvent := &domain.Allocate{
				OrderID:  orderID,
				SKU:      sku,
				Quantity: 10,
			}

			errChan = imp.PublishCommand(allocateEvent)
			assert.True(t, utils.ErrChanIsEmpty(errChan))

			invalidProductSKU := "INVALID_SKU"
			deallocateEvent := &domain.Deallocate{
				OrderID: orderID,
				SKU:     invalidProductSKU,
			}

			errChan = imp.PublishCommand(deallocateEvent)
			assert.NotNil(t, utils.ErrChanWithAny(errChan, ErrProductNotFound))

			product, err := repo.GetProduct(ctx, sku)
			assert.NoError(t, err)

			_, ok := product.Batches[0].Allocations[domain.OrderID(orderID)]
			assert.True(t, ok)
		})

	t.Run("deallocate invalid orderline from batch with wrong orderid",
		func(t *testing.T) {
			ctx := context.Background()

			imp := NewFakeInternalMessagePublisher()

			repo := NewFakeRepository()
			uow := NewFakeUnitOfWork(repo, imp)
			svc := NewAllocationService(uow)

			imp.RegisterCommandHandler(&domain.CreateProduct{}, svc.AddProduct)
			imp.RegisterCommandHandler(&domain.Allocate{}, svc.Allocate)
			imp.RegisterCommandHandler(&domain.Deallocate{}, svc.Deallocate)

			sku := "SMALL-TABLE"
			orderID := "order-001"
			createProductEvent := &domain.CreateProduct{
				SKU: sku,
				Batches: []*domain.Batch{
					domain.NewBatch("batch-001", sku, 100, nil),
				},
			}

			errChan := imp.PublishCommand(createProductEvent)
			assert.True(t, utils.ErrChanIsEmpty(errChan))

			allocateEvent := &domain.Allocate{
				OrderID:  orderID,
				SKU:      sku,
				Quantity: 10,
			}

			errChan = imp.PublishCommand(allocateEvent)
			assert.True(t, utils.ErrChanIsEmpty(errChan))

			var invalidOrderId domain.OrderID = "INVALID_ORDERID"
			deallocateEvent := &domain.Deallocate{
				OrderID: string(invalidOrderId),
				SKU:     sku,
			}

			errChan = imp.PublishCommand(deallocateEvent)
			assert.NotNil(t, utils.ErrChanWithAny(errChan, ErrInvalidOrderID))

			product, err := repo.GetProduct(ctx, sku)
			assert.NoError(t, err)

			_, ok := product.Batches[0].Allocations[domain.OrderID(orderID)]
			assert.True(t, ok)
		})

	t.Run("deallocate valid orderline from batch",
		func(t *testing.T) {
			ctx := context.Background()

			imp := NewFakeInternalMessagePublisher()

			repo := NewFakeRepository()
			uow := NewFakeUnitOfWork(repo, imp)
			svc := NewAllocationService(uow)

			imp.RegisterCommandHandler(&domain.CreateProduct{}, svc.AddProduct)
			imp.RegisterCommandHandler(&domain.Allocate{}, svc.Allocate)
			imp.RegisterCommandHandler(&domain.Deallocate{}, svc.Deallocate)

			sku := "SMALL-TABLE"
			orderID := "order-001"
			createProductEvent := &domain.CreateProduct{
				SKU: sku,
				Batches: []*domain.Batch{
					domain.NewBatch("batch-001", sku, 100, nil),
				},
			}

			errChan := imp.PublishCommand(createProductEvent)
			assert.True(t, utils.ErrChanIsEmpty(errChan))

			allocateEvent := &domain.Allocate{
				OrderID:  orderID,
				SKU:      sku,
				Quantity: 10,
			}

			errChan = imp.PublishCommand(allocateEvent)
			assert.True(t, utils.ErrChanIsEmpty(errChan))

			product, err := repo.GetProduct(ctx, sku)
			assert.NoError(t, err)
			_, ok := product.Batches[0].Allocations[domain.OrderID(orderID)]
			assert.True(t, ok)

			deallocateEvent := &domain.Deallocate{
				OrderID: allocateEvent.OrderID,
				SKU:     sku,
			}

			errChan = imp.PublishCommand(deallocateEvent)
			assert.True(t, utils.ErrChanIsEmpty(errChan))

			product, err = repo.GetProduct(ctx, sku)
			assert.NoError(t, err)
			_, ok = product.Batches[0].Allocations[domain.OrderID(orderID)]
			assert.False(t, ok)
		})

	t.Run("reallocate allocated orderline",
		func(t *testing.T) {
			imp := NewFakeInternalMessagePublisher()

			repo := NewFakeRepository()
			uow := NewFakeUnitOfWork(repo, imp)
			svc := NewAllocationService(uow)

			imp.RegisterCommandHandler(&domain.CreateProduct{}, svc.AddProduct)
			imp.RegisterCommandHandler(&domain.Reallocate{}, svc.Reallocate)

			sku := "SMALL-TABLE"
			otherBatch := domain.NewBatch("batch-999", sku, 100, nil)

			eta := time.Now().Add(time.Hour * 48)
			existingBatch := domain.NewBatch("batch-001", sku, 100, &eta)
			order := &domain.OrderLine{
				SKU:      sku,
				Quantity: 25,
				OrderId:  domain.OrderID("order-001"),
			}
			err := existingBatch.Allocate(order)
			assert.NoError(t, err)

			createProduct := &domain.CreateProduct{
				SKU: sku,
				Batches: []*domain.Batch{
					existingBatch,
					otherBatch,
				},
			}

			errChan := imp.PublishCommand(createProduct)
			assert.True(t, utils.ErrChanIsEmpty(errChan))

			reallocateEvent := &domain.Reallocate{
				OrderID:  string(order.OrderId),
				SKU:      order.SKU,
				Quantity: order.Quantity,
			}

			errChan = imp.PublishCommand(reallocateEvent)
			assert.True(t, utils.ErrChanIsEmpty(errChan))

			assert.Equal(
				t,
				existingBatch.AvailableQuantity(),
				existingBatch.PurchasedQuantity,
				"should ensure the orderline was deallocated from this batch",
			)
		})

	t.Run("out of stock creates an command for external services",
		func(t *testing.T) {
			eventHandler := &MockEventHandler{}
			imp := NewFakeInternalMessagePublisher()

			repo := NewFakeRepository()
			uow := NewFakeUnitOfWork(repo, imp)
			svc := NewAllocationService(uow)

			imp.RegisterCommandHandler(&domain.CreateProduct{}, svc.AddProduct)
			imp.RegisterCommandHandler(&domain.Allocate{}, svc.Allocate)
			imp.RegisterEventHandler(&domain.OutOfStock{}, eventHandler.Handle)

			sku := "SMALL-TABLE"
			batch := domain.NewBatch("batch-001", sku, 10, nil)

			errChan := imp.PublishCommand(&domain.CreateProduct{
				SKU:     sku,
				Batches: []*domain.Batch{batch},
			})
			assert.True(t, utils.ErrChanIsEmpty(errChan))

			errChan = imp.PublishCommand(&domain.Allocate{
				OrderID:  "order-001",
				SKU:      sku,
				Quantity: 15,
			})

			assert.NotNil(t, utils.ErrChanWithAny(errChan, domain.ErrOutOfStock))
			assert.Equal(t, 1, len(eventHandler.ReceivedEvents))
		})

	t.Run("command ChangeBatchQuantity changes available quantity of a batch",
		func(t *testing.T) {
			repo := NewFakeRepository()
			imp := NewFakeInternalMessagePublisher()
			uow := NewFakeUnitOfWork(repo, imp)
			svc := NewAllocationService(uow)

			imp.RegisterCommandHandler(&domain.CreateProduct{}, svc.AddProduct)
			imp.RegisterCommandHandler(&domain.ChangeBatchQuantity{}, svc.ChangeBatchQuantity)

			eta := time.Now()

			command := &domain.CreateProduct{
				SKU: "SMALL-TABLE",
				Batches: []*domain.Batch{
					domain.NewBatch("batch-001", "SMALL-TABLE", 100, &eta),
				},
			}

			errChan := imp.PublishCommand(command)
			assert.True(t, utils.ErrChanIsEmpty(errChan))

			batchEvent := &domain.ChangeBatchQuantity{
				BatchReference:    "batch-001",
				ChangedToQuantity: 10,
			}

			errChan = imp.PublishCommand(batchEvent)
			assert.True(t, utils.ErrChanIsEmpty(errChan))

			ctx := context.Background()
			product, err := repo.GetProduct(ctx, "SMALL-TABLE")
			assert.NoError(t, err)

			foundBatch := false
			for _, batch := range product.Batches {
				if batch.Reference == "batch-001" {
					assert.Equal(t, 10, batch.AvailableQuantity())
					foundBatch = true
				}
			}
			assert.Equal(t, true, foundBatch)
		})

	t.Run("command changes available quantity of a batch with reallocation when needed",
		func(t *testing.T) {
			ctx := context.Background()

			repo := NewFakeRepository()
			imp := NewFakeInternalMessagePublisher()
			uow := NewFakeUnitOfWork(repo, imp)
			svc := NewAllocationService(uow)

			imp.RegisterCommandHandler(&domain.CreateProduct{}, svc.AddProduct)
			imp.RegisterCommandHandler(&domain.Allocate{}, svc.Allocate)
			imp.RegisterCommandHandler(&domain.ChangeBatchQuantity{}, svc.ChangeBatchQuantity)
			imp.RegisterEventHandler(&domain.BatchQuantityChangedRealocationIsNeeded{}, svc.AllocationIsNeeded)

			eta := time.Now()

			productCommand := &domain.CreateProduct{
				SKU: "SMALL-TABLE",
				Batches: []*domain.Batch{
					domain.NewBatch("batch-001", "SMALL-TABLE", 100, nil),
					domain.NewBatch("batch-002", "SMALL-TABLE", 50, &eta),
				},
			}
			errChan := imp.PublishCommand(productCommand)
			assert.True(t, utils.ErrChanIsEmpty(errChan))

			allocationEvent := &domain.Allocate{
				OrderID:  "order-001",
				SKU:      "SMALL-TABLE",
				Quantity: 20,
			}
			errChan = imp.PublishCommand(allocationEvent)
			assert.True(t, utils.ErrChanIsEmpty(errChan))

			allocationEvent = &domain.Allocate{
				OrderID:  "order-002",
				SKU:      "SMALL-TABLE",
				Quantity: 20,
			}
			errChan = imp.PublishCommand(allocationEvent)
			assert.True(t, utils.ErrChanIsEmpty(errChan))

			product, err := repo.GetProduct(ctx, "SMALL-TABLE")
			assert.NoError(t, err)

			foundBatchOne := false
			foundBatchTwo := false
			for _, batch := range product.Batches {
				if batch.Reference == "batch-001" {
					assert.Equal(t, 60, batch.AvailableQuantity())
					foundBatchOne = true
				}
				if batch.Reference == "batch-002" {
					assert.Equal(t, 50, batch.AvailableQuantity())
					foundBatchTwo = true

				}
			}
			assert.True(t, foundBatchOne)
			assert.True(t, foundBatchTwo)

			command := &domain.ChangeBatchQuantity{
				BatchReference:    "batch-001",
				ChangedToQuantity: 30,
			}

			errChan = imp.PublishCommand(command)
			assert.True(t, utils.ErrChanIsEmpty(errChan))

			product, err = repo.GetProduct(ctx, "SMALL-TABLE")
			assert.NoError(t, err)

			foundBatchOne = false
			foundBatchTwo = false
			for _, batch := range product.Batches {
				if batch.Reference == "batch-001" {
					assert.Equal(t, 10, batch.AvailableQuantity(), "order-001 or order-002 should be reallocated")
					foundBatchOne = true
				}
				if batch.Reference == "batch-002" {
					assert.Equal(t, 30, batch.AvailableQuantity(), "order-002 should be reallocated in another batch")
					foundBatchTwo = true
				}
			}
			assert.True(t, foundBatchOne)
			assert.True(t, foundBatchTwo)
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

func (r *FakeRepository) GetProductByBatchReference(_ context.Context, batchReference string) (*domain.Product, error) {
	for _, p := range r.products {
		for _, b := range p.Batches {
			if b.Reference == batchReference {
				return p, nil
			}
		}
	}

	return nil, ErrProductNotFound
}

type FakeUoW struct {
	adapters Adapters
	events   []domain.Event
	imp      MessagePublisher
}

func NewFakeUnitOfWork(repo Repository, imp MessagePublisher) *FakeUoW {
	return &FakeUoW{
		adapters: Adapters{Repository: repo},
		imp:      imp,
	}
}

func (u *FakeUoW) Transact(ctx context.Context, txFunc func(_ Adapters) error) error {
	err := txFunc(u.adapters)
	u.dispatchEvents()
	return err
}

func (u *FakeUoW) AddEvents(events []domain.Event) {
	u.events = append(u.events, events...)
}

func (u *FakeUoW) dispatchEvents() {
	events := u.events
	u.events = nil

	for _, event := range events {
		errChan := u.imp.PublishEvent(event)
		utils.LogErrChan(errChan)
	}
}

type MockEventHandler struct {
	ReceivedEvents []domain.Event
}

func (m *MockEventHandler) Handle(event domain.Event) error {
	m.ReceivedEvents = append(m.ReceivedEvents, event)
	return nil
}

type FakeInternalMessagePublisher struct {
	eventHandlers  map[string][]EventHandler
	commandHandler map[string]CommandHandler
}

func NewFakeInternalMessagePublisher() *FakeInternalMessagePublisher {
	return &FakeInternalMessagePublisher{}
}

func (imp *FakeInternalMessagePublisher) RegisterEventHandler(
	event domain.Event,
	handler EventHandler,
) {
	if imp.eventHandlers == nil {
		imp.eventHandlers = make(map[string][]EventHandler)
	}

	imp.eventHandlers[event.GetEventName()] = append(imp.eventHandlers[event.GetEventName()], handler)
}

func (imp *FakeInternalMessagePublisher) PublishEvent(event domain.Event) <-chan error {
	if eventHandlers, ok := imp.eventHandlers[event.GetEventName()]; ok {
		errChan := make(chan error, len(eventHandlers))

		var wg sync.WaitGroup
		for _, handler := range eventHandlers {
			wg.Add(1)
			go func(e domain.Event) {
				defer wg.Done()
				if err := handler(event); err != nil {
					errChan <- err
				}
			}(event)
		}

		wg.Wait()
		close(errChan)
		return errChan
	}
	return nil
}

func (imp *FakeInternalMessagePublisher) RegisterCommandHandler(
	command domain.Command,
	handler CommandHandler,
) {
	if imp.commandHandler == nil {
		imp.commandHandler = make(map[string]CommandHandler)
	}

	imp.commandHandler[command.GetCommandName()] = handler
}

func (imp *FakeInternalMessagePublisher) PublishCommand(command domain.Command) <-chan error {
	errChan := make(chan error, 1)

	if handler, ok := imp.commandHandler[command.GetCommandName()]; ok {
		if err := handler(command); err != nil {
			errChan <- err
		}
	}

	close(errChan)
	return errChan
}
