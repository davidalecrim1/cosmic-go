package domain

import "encoding/json"

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

// external events

type Allocated struct {
	OrderID  string `json:"order_id"`
	SKU      string `json:"sku"`
	Quantity int    `json:"quantity"`
	BatchRef string `json:"batch_ref"`
}

func (a *Allocated) GetEventName() string {
	return "AllocatedEvent"
}

func (a *Allocated) ToJson() (string, error) {
	data, err := json.Marshal(a)
	if err != nil {
		return "", err
	}

	return string(data), err
}
