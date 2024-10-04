package service

import (
	"context"
	"cosmic-go/internal/domain"
	"errors"
)

var (
	ErrInvalidSku     = errors.New("invalid sku")
	ErrInvalidOrderID = errors.New("invalid order id")
)

type Service struct {
	repo Repository
}

type Repository interface {
	AddBatch(ctx context.Context, b *domain.Batch) error
	GetBatchByReference(ctx context.Context, ref string) (*domain.Batch, error)
	ListBatches(ctx context.Context) ([]*domain.Batch, error)
	GetBatchBySku(ctx context.Context, sku string) (*domain.Batch, error)
	UpdateBatch(ctx context.Context, existingB *domain.Batch, updatedB *domain.Batch) error
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Allocate(ctx context.Context, ol *domain.OrderLine) (string, error) {
	batches, err := s.repo.ListBatches(ctx)
	if err != nil {
		return "", err
	}

	if !isValidSku(ol.Product.SKU, batches) {
		return "", ErrInvalidSku
	}

	return domain.Allocate(ol, batches)
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
	return s.repo.AddBatch(ctx, b)
}

func (s *Service) Deallocate(ctx context.Context, orderid string, sku string) error {
	existingBatch, err := s.repo.GetBatchBySku(ctx, sku)
	if err != nil {
		return ErrInvalidSku
	}

	updatedBatch := *existingBatch // coping it
	orderline, err := s.getOrderLineAllocatedFromBatch(orderid, existingBatch)
	if err != nil {
		return err
	}

	err = updatedBatch.Deallocate(orderline)
	if err != nil {
		return err
	}

	err = s.repo.UpdateBatch(ctx, existingBatch, &updatedBatch)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) getOrderLineAllocatedFromBatch(
	orderid string,
	b *domain.Batch,
) (*domain.OrderLine, error) {
	for _, al := range b.Allocations {
		if al.OrderId == orderid {
			return &al, nil
		}
	}

	return nil, ErrInvalidOrderID
}
