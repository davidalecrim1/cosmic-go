package handler

import "time"

type AllocationRequest struct {
	OrderID  string `json:"orderid" validate:"required"`
	SKU      string `json:"sku" validate:"required"`
	Quantity int    `json:"quantity" validate:"required"`
}

type AllocationResponse struct {
	BatchRef string `json:"batchref"`
}

type DeallocateRequest struct {
	OrderID string `json:"orderid" validate:"required"`
	SKU     string `json:"sku" validate:"required"`
}

type AddBatchRequest struct {
	Reference         string     `json:"reference" validate:"required"`
	Product           ProductDTO `json:"product"`
	PurchasedQuantity int        `json:"purchased_quantity" validate:"required,gt=0"`
	ETA               time.Time  `json:"eta"`
}

type ProductDTO struct {
	SKU string `json:"sku" validate:"required"`
}

type BadRequestResponse struct {
	Message string `json:"message"`
}
