package application

import (
	"context"
	"errors"

	"cosmic-go/internal/domain"
)

var (
	ErrInvalidSku     = errors.New("invalid sku")
	ErrInvalidOrderID = errors.New("invalid order id")
	ErrBatchNotFound  = errors.New("batch not found")
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

func (s *Service) Allocate(ctx context.Context, ol *domain.OrderLine) (string, error) {
	var updatedBatchRef string

	err := s.uow.Transact(ctx, func(adapters Adapters) error {
		batchRef, err := s.processAllocation(ctx, ol, adapters)
		if err != nil {
			return err
		}

		updatedBatchRef = batchRef
		return err
	})
	if err != nil {
		return "", err
	}

	return updatedBatchRef, nil
}

func (s *Service) processAllocation(ctx context.Context, ol *domain.OrderLine, adapters Adapters) (string, error) {
	var updatedBatchRef string

	batches, err := adapters.Repository.ListBatches(ctx)
	if err != nil {
		return "", err
	}

	if !isValidSku(ol.Product.SKU, batches) {
		return "", ErrInvalidSku
	}

	updatedBatchRef, err = domain.Allocate(ol, batches)
	if err != nil {
		return "", err
	}

	existingBatch, err := adapters.Repository.GetBatchByReference(ctx, updatedBatchRef)
	if err != nil {
		return "", err
	}

	for _, batch := range batches {
		if batch.Reference == updatedBatchRef {
			err = adapters.Repository.UpdateBatch(ctx, existingBatch, batch)
			if err != nil {
				return "", err
			}
		}
	}
	return updatedBatchRef, nil
}

func isValidSku(sku string, batches []*domain.Batch) bool {
	for _, batch := range batches {
		if batch.Product.SKU == sku {
			return true
		}
	}
	return false
}

func (s *Service) AddBatch(ctx context.Context, b *domain.Batch) error {
	return s.uow.Transact(ctx, func(adapters Adapters) error {
		return adapters.Repository.AddBatch(ctx, b)
	})
}

func (s *Service) Deallocate(ctx context.Context, orderid string, sku string) error {
	return s.uow.Transact(ctx, func(adapters Adapters) error {
		return s.processDeallocation(ctx, orderid, sku, adapters)
	})
}

func (s *Service) processDeallocation(ctx context.Context, orderid string, sku string, adapters Adapters) error {
	existingBatch, err := adapters.Repository.GetBatchBySku(ctx, sku)
	if err != nil {
		return ErrInvalidSku
	}

	// coping it, here we don't need a hard copy
	updatedBatch := *existingBatch
	orderline, err := s.findOrderLineInBatch(orderid, existingBatch)
	if err != nil {
		return err
	}

	err = updatedBatch.Deallocate(orderline)
	if err != nil {
		return err
	}

	return adapters.Repository.UpdateBatch(ctx, existingBatch, &updatedBatch)
}

func (s *Service) findOrderLineInBatch(
	orderid string,
	b *domain.Batch,
) (*domain.OrderLine, error) {
	for _, al := range b.Allocations {
		if string(al.OrderId) == orderid {
			return &al, nil
		}
	}

	return nil, ErrInvalidOrderID
}

func (s *Service) Reallocate(ctx context.Context, ol *domain.OrderLine) (string, error) {
	var updatedBatchRef string

	err := s.uow.Transact(ctx, func(adapters Adapters) error {
		batch, err := adapters.Repository.GetBatchBySku(ctx, ol.Product.SKU)
		if err != nil {
			return err
		}

		err = batch.Deallocate(ol)
		if err != nil {
			return err
		}

		updatedBatchRef, err = s.processAllocation(ctx, ol, adapters)
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
