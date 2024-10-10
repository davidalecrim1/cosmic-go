package helpers

import (
	"context"
	"testing"

	"cosmic-go/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TestBatchData struct {
	Products    []*domain.Product
	Batches     []*domain.Batch
	OrderLines  []*domain.OrderLine
	Allocations []AllocatonData
}

type AllocatonData struct {
	OrderID  string
	BatchRef string
}

type TestBatchDataOption func(*TestBatchData)

func CreateTestBatchData(t *testing.T, db *pgxpool.Pool, opts ...TestBatchDataOption) *TestBatchData {
	t.Helper()

	b := &TestBatchData{}
	for _, opt := range opts {
		opt(b)
	}

	ctx := context.Background()
	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatal("failed to create transaction: ", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			err := tx.Commit(ctx)
			if err != nil {
				t.Fatal("failed to commit transaction: ", err)
			}
		}
	}()

	b.createProduct(t, ctx, tx)
	b.createBatches(t, ctx, tx)
	b.createOrderLines(t, ctx, tx)
	b.createAllocations(t, ctx, tx)

	return b
}

func (b *TestBatchData) createProduct(t *testing.T, ctx context.Context, tx pgx.Tx) {
	if len(b.Products) == 0 {
		return
	}

	query := `
	INSERT INTO products (sku)
	VALUES ($1);
	`

	for _, p := range b.Products {
		_, err := tx.Exec(ctx, query, p.SKU)
		if err != nil {
			t.Fatal("failed to insert product: ", err)
		}
	}
}

func (b *TestBatchData) createBatches(t *testing.T, ctx context.Context, tx pgx.Tx) {
	if len(b.Batches) == 0 {
		return
	}

	query := `
	INSERT INTO batches (reference, product_sku, purchased_quantity, eta)
	VALUES ($1, $2, $3, $4);
	`

	for _, batch := range b.Batches {
		_, err := tx.Exec(ctx, query, batch.Reference, batch.Product.SKU, batch.PurchasedQuantity, batch.GetETA())
		if err != nil {
			t.Fatal("failed to insert batch: ", err)
		}
	}
}

func (b *TestBatchData) createOrderLines(t *testing.T, ctx context.Context, tx pgx.Tx) {
	if len(b.OrderLines) == 0 {
		return
	}

	query := `
	INSERT INTO order_lines (product_sku, quantity, orderid)
	VALUES ($1, $2, $3);
	`

	for _, ol := range b.OrderLines {
		_, err := tx.Exec(ctx, query, ol.Product.SKU, ol.Quantity, ol.OrderId)
		if err != nil {
			t.Fatal("failed to insert order line: ", err)
		}
	}
}

func (b *TestBatchData) createAllocations(t *testing.T, ctx context.Context, tx pgx.Tx) {
	if len(b.Allocations) == 0 {
		return
	}

	query := `
	INSERT INTO allocations (orderline_id, batch_reference)
	VALUES (
		(SELECT id FROM order_lines WHERE orderid = $1), $2);
	`

	for _, alloc := range b.Allocations {
		_, err := tx.Exec(ctx, query, alloc.OrderID, alloc.BatchRef)
		if err != nil {
			t.Fatal("failed to insert allocation: ", err)
		}
	}
}

func WithProduct(p *domain.Product) TestBatchDataOption {
	return func(data *TestBatchData) {
		data.Products = append(data.Products, p)
	}
}

func WithBatch(b *domain.Batch) TestBatchDataOption {
	return func(data *TestBatchData) {
		data.Batches = append(data.Batches, b)
	}
}

func WithOrderLine(ol *domain.OrderLine) TestBatchDataOption {
	return func(data *TestBatchData) {
		data.OrderLines = append(data.OrderLines, ol)
	}
}

func WithAllocation(orderid string, batchRef string) TestBatchDataOption {
	return func(data *TestBatchData) {
		data.Allocations = append(data.Allocations, AllocatonData{
			OrderID:  orderid,
			BatchRef: batchRef,
		})
	}
}
