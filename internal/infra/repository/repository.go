package repository

import (
	"context"
	"errors"

	"cosmic-go/internal/domain"
	"cosmic-go/internal/infra/database"

	"gorm.io/gorm"
)

type PostgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) GetProduct(
	ctx context.Context,
	sku string,
) (*domain.Product, error) {
	var productDTO database.Product

	if err := r.db.Preload("Batches.Allocations").Preload("Batches.Allocations.OrderLine").First(&productDTO, "sku = ?", sku).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrProductNotFound
		}
		return nil, err
	}

	return r.mapDTOToProduct(productDTO)
}

func (r *PostgresRepository) AddProduct(
	ctx context.Context,
	p *domain.Product,
) error {
	dto := r.mapProductToDTO(p)
	return r.db.WithContext(ctx).Create(dto).Error
}

func (r *PostgresRepository) UpdateProduct(
	ctx context.Context,
	p *domain.Product,
) error {
	dto := r.mapProductToDTO(p)
	CurrentVersionId := dto.VersionId - 1 // this is increased in the domain previously.

	return r.db.
		WithContext(ctx).
		Session(&gorm.Session{FullSaveAssociations: true}).
		Preload("Batch.Allocation.OrderLine").
		Where("version_id = ?", CurrentVersionId).
		Save(dto).
		Error
}

func (r *PostgresRepository) GetProductByBatchReference(
	ctx context.Context,
	batchReference string,
) (*domain.Product, error) {
	var sku string
	err := r.db.Model(&database.Batch{}).
		Select("sku").
		Where("reference = ?", batchReference).
		First(&sku).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrProductNotFound
		}
		return nil, err
	}
	return r.GetProduct(ctx, sku)
}

func (r *PostgresRepository) mapProductToDTO(p *domain.Product) *database.Product {
	productDTO := database.Product{
		SKU:       p.SKU,
		VersionId: p.VersionId,
	}

	batchesDTO := r.mapBatchesToDTO(p.Batches)
	productDTO.Batches = batchesDTO
	return &productDTO
}

func (r *PostgresRepository) mapBatchesToDTO(batches []*domain.Batch) []*database.Batch {
	batchesDTO := make([]*database.Batch, 0, len(batches))

	for _, batch := range batches {
		batchDTO := &database.Batch{
			Reference:         batch.Reference,
			SKU:               batch.SKU,
			PurchasedQuantity: batch.PurchasedQuantity,
			ETA:               batch.GetETA(),
			Allocations:       r.mapAllocationsToDTOs(batch),
		}

		batchesDTO = append(batchesDTO, batchDTO)
	}

	return batchesDTO
}

func (r *PostgresRepository) mapAllocationsToDTOs(batch *domain.Batch) []*database.Allocation {
	allocationsDTO := make([]*database.Allocation, 0, len(batch.Allocations))

	for _, alloc := range batch.Allocations {
		allocationDTO := &database.Allocation{
			BatchReference: batch.Reference,
			OrderLine: &database.OrderLine{
				SKU:      alloc.SKU,
				Quantity: alloc.Quantity,
				OrderId:  string(alloc.OrderId),
			},
		}

		allocationsDTO = append(allocationsDTO, allocationDTO)

	}
	return allocationsDTO
}

func (r *PostgresRepository) mapDTOToProduct(productDTO database.Product) (*domain.Product, error) {
	batches, err := r.mapDTOToBatches(productDTO.Batches)
	if err != nil {
		return nil, err
	}

	return domain.NewProduct(
		productDTO.SKU,
		batches,
		productDTO.VersionId,
	), nil
}

func (r *PostgresRepository) mapDTOToBatches(batchesDTO []*database.Batch) ([]*domain.Batch, error) {
	batches := make([]*domain.Batch, 0, len(batchesDTO))

	for _, batchDTO := range batchesDTO {
		batch := domain.NewBatch(
			batchDTO.Reference,
			batchDTO.SKU,
			batchDTO.PurchasedQuantity,
			batchDTO.ETA,
		)
		batches = append(batches, batch)

		for _, alloc := range batchDTO.Allocations {
			line := &domain.OrderLine{
				SKU:      alloc.OrderLine.SKU,
				Quantity: alloc.OrderLine.Quantity,
				OrderId:  domain.OrderID(alloc.OrderLine.OrderId),
			}
			err := batch.Allocate(line)
			if err != nil {
				return nil, err
			}
		}
	}

	return batches, nil
}
