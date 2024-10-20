package database

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	SKU       string   `gorm:"primaryKey"`
	Batches   []*Batch `gorm:"foreignKey:SKU;references:SKU"`
	VersionId int
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Batch struct {
	Reference         string `gorm:"primaryKey"`
	SKU               string // Foreign key to Product
	PurchasedQuantity int
	ETA               *time.Time
	Allocations       []*Allocation `gorm:"foreignKey:BatchReference;references:Reference"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         gorm.DeletedAt `gorm:"index"`
}

func (b *Batch) BeforeSave(tx *gorm.DB) (err error) {
	deleteDeallocatedOrderLinesFromBatch(b, tx)
	return nil
}

func deleteDeallocatedOrderLinesFromBatch(b *Batch, tx *gorm.DB) error {
	existingAllocations := []*Allocation{}
	err := tx.Preload("OrderLine").Where("batch_reference = ?", b.Reference).Find(&existingAllocations).Error
	if err != nil {
		return err
	}

	if len(existingAllocations) == 0 {
		return nil
	}

	if len(b.Allocations) == 0 {
		tx.Where("batch_reference = ?", b.Reference).Delete(&existingAllocations)
		return nil
	}

	updatedAllocations := make(map[string]struct{})
	for _, alloc := range b.Allocations {
		updatedAllocations[alloc.OrderLine.OrderId] = struct{}{}
	}

	for _, alloc := range existingAllocations {
		if _, ok := updatedAllocations[alloc.OrderLine.OrderId]; !ok {
			tx.Where("batch_reference = ?", b.Reference).Where("id = ?", alloc.ID).Delete(&alloc)
		}
	}

	return nil
}

type Allocation struct {
	ID             uint
	OrderLineID    uint       // Foreign key to OrderLine
	OrderLine      *OrderLine `gorm:"foreignKey:OrderLineID;references:ID"`
	BatchReference string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

type OrderLine struct {
	gorm.Model
	OrderId  string
	SKU      string
	Quantity int
}
