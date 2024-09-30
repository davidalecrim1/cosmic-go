package service

import (
	"cosmic-go/internal/domain"
	"errors"
)

var (
	ErrInvalidSku = errors.New("invalid sku")
)

type Service struct {
	repo Repository
}

type Repository interface {
	AddBatch(*domain.Batch) error
	GetBatch(ref string) (*domain.Batch, error)
	ListBatches() ([]*domain.Batch, error)
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Allocate(ol *domain.OrderLine) (string, error) {
	batches, err := s.repo.ListBatches()
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

func (s *Service) AddBatch(b *domain.Batch) error {
	return s.repo.AddBatch(b)
}
