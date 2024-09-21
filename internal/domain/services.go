package domain

import "errors"

var (
	ErrInvalidSku = errors.New("invalid sku")
)

type Service struct {
	repo Repository
}

type Repository interface {
	Add(*Batch) error
	Get(ref string) (*Batch, error)
	List() ([]*Batch, error)
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Allocate(ol *OrderLine) (string, error) {
	batches, err := s.repo.List()
	if err != nil {
		return "", err
	}

	if !isValidSku(ol.Product.SKU, batches) {
		return "", ErrInvalidSku
	}

	return Allocate(ol, batches)
}

func isValidSku(sku string, batches []*Batch) bool {
	for _, batch := range batches {
		if batch.Product.SKU == sku {
			return true
		}
	}
	return false
}
