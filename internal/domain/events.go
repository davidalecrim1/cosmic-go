package domain

type Event interface {
	EventName() string
}

type OutOfStockEvent struct {
	SKU string
}

func (o *OutOfStockEvent) EventName() string {
	return "OutOfStock"
}
