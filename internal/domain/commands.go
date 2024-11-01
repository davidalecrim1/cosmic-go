package domain

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
	BatchReference    string
	ChangedToQuantity int
}

func (b *ChangeBatchQuantity) GetCommandName() string {
	return "ChangeBatchQuantityCommand"
}
