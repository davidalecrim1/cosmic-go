package domain

type Event interface {
	GetEventName() string
}

type OutOfStock struct {
	SKU string
}

func (o *OutOfStock) GetEventName() string {
	return "OutOfStockEvent"
}

type BatchQuantityChangedRealocationIsNeeded struct {
	OrderID  string
	SKU      string
	Quantity int
}

func (b *BatchQuantityChangedRealocationIsNeeded) GetEventName() string {
	return "BatchChangedQuantityRealocationIsNeededEvent"
}
