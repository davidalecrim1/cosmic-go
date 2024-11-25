package application

import (
	"context"
	"errors"

	"cosmic-go/internal/domain"
)

var (
	ErrInvalidOrderID   = errors.New("invalid order id")
	ErrProductNotFound  = errors.New("product not found")
	ErrInvalidEventType = errors.New("invalid event type")
)

type UoW interface {
	Transact(ctx context.Context, txFunc func(adapters Adapters) error) error
	AddEvents(events []domain.Event)
}

type AllocationService struct {
	uow UoW
}

func NewAllocationService(uow UoW) *AllocationService {
	return &AllocationService{uow: uow}
}

func (s *AllocationService) Allocate(c domain.Command) error {
	line, err := s.mapCommandToOrderLine(c)
	if err != nil {
		return err
	}

	ctx := context.Background()
	return s.uow.Transact(ctx, func(adapters Adapters) error {
		_, err := s.processAllocation(ctx, line, adapters)
		return err
	})
}

func (s *AllocationService) mapCommandToOrderLine(c domain.Command) (*domain.OrderLine, error) {
	switch event := c.(type) {
	case *domain.Allocate:
		return &domain.OrderLine{
			OrderId:  domain.OrderID(event.OrderID),
			SKU:      event.SKU,
			Quantity: event.Quantity,
		}, nil
	case *domain.Reallocate:
		return &domain.OrderLine{
			OrderId:  domain.OrderID(event.OrderID),
			SKU:      event.SKU,
			Quantity: event.Quantity,
		}, nil
	case *domain.Deallocate:
		return &domain.OrderLine{
			OrderId: domain.OrderID(event.OrderID),
			SKU:     event.SKU,
		}, nil
	default:
		return nil, ErrInvalidEventType
	}
}

func (s *AllocationService) processAllocation(ctx context.Context, line *domain.OrderLine, adapters Adapters) (string, error) {
	var updatedBatchRef string

	p, err := adapters.Repository.GetProduct(ctx, line.SKU)
	if errors.Is(err, domain.ErrProductNotFound) {
		return "", ErrProductNotFound
	}
	if err != nil {
		return "", err
	}
	defer func() {
		s.uow.AddEvents(p.PopEvents())
	}()

	updatedBatchRef, err = p.Allocate(line)
	if err != nil {
		return "", err
	}

	err = adapters.Repository.UpdateProduct(ctx, p)
	if err != nil {
		return "", err
	}

	return updatedBatchRef, nil
}

func (s *AllocationService) AddProduct(c domain.Command) error {
	ctx := context.Background()
	p, err := s.mapEventToProduct(c)
	if err != nil {
		return err
	}

	return s.uow.Transact(ctx, func(adapters Adapters) error {
		return adapters.Repository.AddProduct(ctx, p)
	})
}

func (s *AllocationService) mapEventToProduct(c domain.Command) (*domain.Product, error) {
	switch event := c.(type) {
	case *domain.CreateProduct:
		return domain.NewProduct(event.SKU, event.Batches, 0), nil
	default:
		return nil, ErrInvalidEventType
	}
}

func (s *AllocationService) Deallocate(c domain.Command) error {
	line, err := s.mapCommandToOrderLine(c)
	if err != nil {
		return err
	}

	ctx := context.Background()
	return s.uow.Transact(ctx, func(adapters Adapters) error {
		return s.processDeallocation(ctx, domain.OrderID(line.OrderId), line.SKU, adapters)
	})
}

func (s *AllocationService) processDeallocation(ctx context.Context, orderid domain.OrderID, sku string, adapters Adapters) error {
	p, err := adapters.Repository.GetProduct(ctx, sku)
	if err != nil {
		return ErrProductNotFound
	}

	err = p.Deallocate(orderid)
	if err != nil {
		if errors.Is(err, domain.ErrCannotDeallocateUnallocatedOrderLine) {
			return ErrInvalidOrderID
		}
		return err
	}

	return adapters.Repository.UpdateProduct(ctx, p)
}

func (s *AllocationService) Reallocate(c domain.Command) error {
	line, err := s.mapCommandToOrderLine(c)
	if err != nil {
		return err
	}

	return s.processRealocation(line)
}

func (s *AllocationService) processRealocation(line *domain.OrderLine) error {
	ctx := context.Background()
	return s.uow.Transact(ctx, func(adapters Adapters) error {
		p, err := adapters.Repository.GetProduct(ctx, line.SKU)
		if err != nil {
			return err
		}

		err = p.Deallocate(line.OrderId)
		if err != nil {
			return err
		}

		_, err = s.processAllocation(ctx, line, adapters)
		return err
	})
}

func (s *AllocationService) ChangeBatchQuantity(c domain.Command) error {
	event, err := s.mapEventToBatchReference(c)
	if err != nil {
		return err
	}

	ctx := context.Background()
	err = s.uow.Transact(ctx, func(adapters Adapters) error {
		product, err := adapters.Repository.GetProductByBatchReference(ctx, event.BatchReference)
		if err != nil {
			return err
		}

		err = product.ChangeBatchQuantity(event.BatchReference, event.ChangedToQuantity)
		if err != nil {
			return err
		}

		err = adapters.Repository.UpdateProduct(ctx, product)
		if err != nil {
			return err
		}
		s.uow.AddEvents(product.PopEvents())
		return nil
	})

	return err
}

func (s *AllocationService) mapEventToBatchReference(c domain.Command) (*domain.ChangeBatchQuantity, error) {
	switch event := c.(type) {
	case *domain.ChangeBatchQuantity:
		return event, nil
	default:
		return nil, ErrInvalidEventType
	}
}

func (s *AllocationService) AllocationIsNeeded(event domain.Event) error {
	line, err := s.mapEventToOrderLine(event)
	if err != nil {
		return err
	}

	ctx := context.Background()
	return s.uow.Transact(ctx, func(adapters Adapters) error {
		_, err := s.processAllocation(ctx, line, adapters)
		return err
	})
}

func (s *AllocationService) mapEventToOrderLine(event domain.Event) (*domain.OrderLine, error) {
	switch event := event.(type) {
	case *domain.BatchQuantityChangedRealocationIsNeeded:
		return &domain.OrderLine{
			OrderId:  domain.OrderID(event.OrderID),
			SKU:      event.SKU,
			Quantity: event.Quantity,
		}, nil
	default:
		return nil, ErrInvalidEventType
	}
}
