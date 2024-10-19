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
