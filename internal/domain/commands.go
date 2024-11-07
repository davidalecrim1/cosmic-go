package domain

import "encoding/json"

type Command interface {
	GetCommandName() string
}

type CreateProduct struct {
	SKU     string
	Batches []*Batch
}

func (b *CreateProduct) GetCommandName() string {
	return "CreateProduct"
}

type Allocate struct {
	OrderID  string
	SKU      string
	Quantity int
}

func (a *Allocate) GetCommandName() string {
	return "AllocateCommand"
}

type Deallocate struct {
	OrderID string
	SKU     string
}

func (d *Deallocate) GetCommandName() string {
	return "DeallocateCommand"
}

type Reallocate struct {
	OrderID  string
	SKU      string
	Quantity int
}

func (r *Reallocate) GetCommandName() string {
	return "ReallocateCommand"
}

type ChangeBatchQuantity struct {
	BatchReference    string `json:"batch_reference"`
	ChangedToQuantity int    `json:"changed_to_quantity"`
}

func (b *ChangeBatchQuantity) GetCommandName() string {
	return "ChangeBatchQuantityCommand"
}

func (b *ChangeBatchQuantity) ToJson() (string, error) {
	data, err := json.Marshal(b)
	if err != nil {
		return "", err
	}

	return string(data), err
}

func NewChangeBatchQuantityFromJson(data string) (*ChangeBatchQuantity, error) {
	command := &ChangeBatchQuantity{}
	err := json.Unmarshal([]byte(data), command)
	if err != nil {
		return nil, err
	}

	return command, nil
}
