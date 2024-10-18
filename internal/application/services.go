package application

import (
	"context"
	"errors"

	"cosmic-go/internal/domain"
)

var (
	ErrInvalidOrderID  = errors.New("invalid order id")
	ErrProductNotFound = errors.New("product not found")
)

type Service struct {
	uow UoW
}

type UoW interface {
	Transact(ctx context.Context, txFunc func(adapters Adapters) error) error
}

func NewService(uow UoW) *Service {
	return &Service{uow: uow}
}

func (s *Service) Allocate(ctx context.Context, line *domain.OrderLine) (string, error) {
	var updatedBatchRef string

	err := s.uow.Transact(ctx, func(adapters Adapters) error {
		batchRef, err := s.processAllocation(ctx, line, adapters)
		if err != nil {
			return err
		}

		updatedBatchRef = batchRef
		return err
	})
	if err != nil {
		if err.Error() == "product not found" {
			return "", ErrProductNotFound
		}
		return "", err
	}

	return updatedBatchRef, nil
}

func (s *Service) processAllocation(ctx context.Context, line *domain.OrderLine, adapters Adapters) (string, error) {
	var updatedBatchRef string

	p, err := adapters.Repository.GetProduct(ctx, line.SKU)
	if err != nil {
		return "", err
	}

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

func (s *Service) AddProduct(ctx context.Context, p *domain.Product) error {
	return s.uow.Transact(ctx, func(adapters Adapters) error {
		return adapters.Repository.AddProduct(ctx, p)
	})
}

func (s *Service) Deallocate(ctx context.Context, orderid domain.OrderID, sku string) error {
	return s.uow.Transact(ctx, func(adapters Adapters) error {
		return s.processDeallocation(ctx, orderid, sku, adapters)
	})
}

func (s *Service) processDeallocation(ctx context.Context, orderid domain.OrderID, sku string, adapters Adapters) error {
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

func (s *Service) Reallocate(ctx context.Context, line *domain.OrderLine) (string, error) {
	var updatedBatchRef string

	err := s.uow.Transact(ctx, func(adapters Adapters) error {
		p, err := adapters.Repository.GetProduct(ctx, line.SKU)
		if err != nil {
			return err
		}

		err = p.Deallocate(line.OrderId)
		if err != nil {
			return err
		}

		updatedBatchRef, err = s.processAllocation(ctx, line, adapters)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return updatedBatchRef, nil
}
