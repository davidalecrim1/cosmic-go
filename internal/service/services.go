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
	Add(*domain.Batch) error
	Get(ref string) (*domain.Batch, error)
	List() ([]*domain.Batch, error)
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Allocate(ol *domain.OrderLine) (string, error) {
	batches, err := s.repo.List()
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
