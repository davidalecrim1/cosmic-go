package handler

import "time"

type AllocationRequest struct {
	OrderID  string `json:"orderid" validate:"required"`
	SKU      string `json:"sku" validate:"required"`
	Quantity int    `json:"quantity" validate:"required"`
}

type DeallocateRequest struct {
	OrderID string `json:"orderid" validate:"required"`
	SKU     string `json:"sku" validate:"required"`
}

type AddProductRequest struct {
	SKU     string      `json:"sku" validate:"required"`
	Batches []*BatchDTO `json:"batches"`
}

type BatchDTO struct {
	Reference         string     `json:"reference" validate:"required"`
	PurchasedQuantity int        `json:"purchased_quantity" validate:"required,gt=0"`
	ETA               *time.Time `json:"eta"`
}

type BadRequestResponse struct {
	Message string `json:"message"`
}

type AllocationResponse struct {
	SKU            string `json:"sku"`
	BatchReference string `json:"batch_reference"`
}

type AllocationsResponse struct {
	Allocations []*AllocationResponse
}
