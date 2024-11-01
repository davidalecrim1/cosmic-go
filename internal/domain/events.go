package domain

type Event interface {
	EventName() string
}

type OutOfStock struct {
	SKU string
}

func (o *OutOfStock) EventName() string {
	return "OutOfStock"
}

type CreateProduct struct {
	SKU     string
	Batches []*Batch
}

func (b *CreateProduct) EventName() string {
	return "CreateProduct"
}

type AllocationRequired struct {
	OrderID  string
	SKU      string
	Quantity int
}

func (a *AllocationRequired) EventName() string {
	return "AlocationRequired"
}

type DeallocationRequired struct {
	OrderID string
	SKU     string
}

func (d *DeallocationRequired) EventName() string {
	return "DeallocationRequired"
}

type ReallocationRequired struct {
	OrderID  string
	SKU      string
	Quantity int
}

func (r *ReallocationRequired) EventName() string {
	return "ReallocationRequired"
}

type BatchQuantityChanged struct {
	BatchReference    string
	ChangedToQuantity int
}

func (b *BatchQuantityChanged) EventName() string {
	return "BatchQuantityChanged"
}
