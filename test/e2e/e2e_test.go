//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"cosmic-go/internal/bootstrap"
	"cosmic-go/internal/domain"
	"cosmic-go/internal/handler"
	"cosmic-go/pkg/env"
	"cosmic-go/test/helpers"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
)

func TestE2E(t *testing.T) {
	db := bootstrap.InitializeDatabase()
	defer db.Close()

	t.Run("api returns allocation with 201",
		func(t *testing.T) {
			orderId := "order-001"
			sku := "SMALL-TABLE"
			quantity := 10

			earlyBatch := domain.NewBatch("batch-001", domain.Product{SKU: sku}, 100, time.Now())
			mediumBatch := domain.NewBatch("batch-002", domain.Product{SKU: sku}, 100, time.Now().Add(time.Hour*24))
			otherBatch := domain.NewBatch("batch-003", domain.Product{SKU: sku}, 100, time.Time{})

			insertProductHelper(t, db, domain.Product{SKU: sku})
			apiEndpoint := getApiEndpointHelper(t)

			insertBatchHelper(t, db,
				[]*domain.Batch{
					earlyBatch,
					mediumBatch,
					otherBatch,
				})

			requestBody := fmt.Sprintf(`
			{
				"orderid": "%s",
				"sku": "%s",
				"quantity": %d
			}
			`, orderId, sku, quantity)

			resp, err := http.Post(
				apiEndpoint+"/allocate",
				"application/json",
				bytes.NewBuffer([]byte(requestBody)),
			)

			assert.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusCreated, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)

			aResp := &handler.AllocationResponse{}
			err = json.Unmarshal(body, aResp)
			assert.NoError(t, err)

			expectedBatch := "batch-003"
			assert.Equal(t, expectedBatch, aResp.BatchRef)

			t.Cleanup(func() {
				helpers.CleanUpRepositoryHelper(db)
			})
		})

	t.Run("api returns 400 and error message",
		func(t *testing.T) {
			unknownSku, unknownOrderId, quantity := "UNKNOWN-SKU-001", "unknown-order-001", 10
			apiEndpoint := getApiEndpointHelper(t)

			requestBody := fmt.Sprintf(`
			{
				"orderid": "%s",
				"sku": "%s",
				"quantity": %d
			}
			`, unknownOrderId, unknownSku, quantity)

			resp, err := http.Post(apiEndpoint+"/allocate",
				"application/json",
				bytes.NewBuffer([]byte(requestBody)))

			assert.NoError(t, err)
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)

			aBody := &handler.BadRequestResponse{}

			err = json.Unmarshal(body, aBody)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			assert.Equal(t, "invalid sku", aBody.Message)

			t.Cleanup(func() {
				helpers.CleanUpRepositoryHelper(db)
			})
		})

	t.Run("api returns 200 for deallocate",
		func(t *testing.T) {
			sku := "SMALL-TABLE"

			orderLine := &domain.OrderLine{
				Product:  domain.Product{SKU: sku},
				Quantity: 5,
				OrderId:  "order-001",
			}

			batch := domain.NewBatch("batch-001", domain.Product{SKU: sku}, 20, time.Now())

			insertProductHelper(t, db, domain.Product{SKU: sku})
			insertBatchHelper(t, db, []*domain.Batch{batch})
			insertOrderLineAndAllocationsToBatchHelper(t, db, orderLine, batch.Reference)

			apiEndpoint := getApiEndpointHelper(t)
			requestBody := fmt.Sprintf(`
				{
					"orderid": "%s",
					"sku": "%s"
				}
				`,
				orderLine.OrderId,
				sku,
			)

			resp, err := http.Post(apiEndpoint+"/deallocate",
				"application/json",
				bytes.NewBuffer([]byte(requestBody)),
			)
			assert.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, http.StatusOK, resp.StatusCode)

			t.Cleanup(func() {
				helpers.CleanUpRepositoryHelper(db)
			})
		})

	t.Run("invalid order id for deallocation",
		func(t *testing.T) {
			sku := "SMALL-TABLE"

			orderLine := &domain.OrderLine{
				Product:  domain.Product{SKU: sku},
				Quantity: 5,
				OrderId:  "order-001",
			}

			batch := domain.NewBatch("batch-001", domain.Product{SKU: sku}, 20, time.Now())

			insertProductHelper(t, db, domain.Product{SKU: sku})
			insertBatchHelper(t, db, []*domain.Batch{batch})
			insertOrderLineAndAllocationsToBatchHelper(t, db, orderLine, batch.Reference)

			apiEndpoint := getApiEndpointHelper(t)
			invalidOrderId := "order-002"

			requestBody := fmt.Sprintf(`
				{
					"orderid": "%s",
					"sku": "%s"
				}
				`,
				invalidOrderId,
				sku,
			)

			resp, err := http.Post(apiEndpoint+"/deallocate",
				"application/json",
				bytes.NewBuffer([]byte(requestBody)),
			)
			assert.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

			t.Cleanup(func() {
				helpers.CleanUpRepositoryHelper(db)
			})
		})
}

func getApiEndpointHelper(t *testing.T) string {
	t.Helper()
	return env.GetEnvOrSetDefault("API_URL", "http://localhost:8080")
}

func insertProductHelper(t *testing.T, db *pgxpool.Pool, product domain.Product) {
	t.Helper()
	ctx := context.Background()

	query := `
	INSERT INTO products (sku)
	VALUES ($1);
	`
	_, err := db.Exec(ctx, query, product.SKU)
	if err != nil {
		t.Fatal("failed to insert product: ", err)
	}
}

func insertBatchHelper(t *testing.T, db *pgxpool.Pool, batches []*domain.Batch) {
	t.Helper()
	ctx := context.Background()

	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatal("failed to insert batch helper: ", err)
	}
	defer tx.Rollback(ctx)

	query := `
	INSERT INTO batches (reference, product_sku, purchased_quantity, eta)
	VALUES ($1, $2, $3, $4);
	`

	for _, batch := range batches {
		_, err = tx.Exec(ctx, query, batch.Reference, batch.Product.SKU, batch.PurchasedQuantity, batch.ETA)
		if err != nil {
			t.Fatal("failed to insert batch: ", err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		t.Fatal("failed to commit tx: ", err)
	}
}

func insertOrderLineAndAllocationsToBatchHelper(
	t *testing.T,
	db *pgxpool.Pool,
	ol *domain.OrderLine,
	batchRef string,
) {
	t.Helper()
	ctx := context.Background()

	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatal("failed to insert batch helper: ", err)
	}
	defer tx.Rollback(ctx)

	var orderlineId int
	query := `INSERT INTO order_lines (product_sku, quantity, orderid)
	VALUES ($1, $2, $3) RETURNING id;`
	err = tx.QueryRow(ctx, query, ol.Product.SKU, ol.Quantity, ol.OrderId).Scan(&orderlineId)
	assert.NoError(t, err)

	query = `INSERT INTO allocations (orderline_id, batch_reference)
	VALUES ($1, $2);`
	_, err = tx.Exec(ctx, query, orderlineId, batchRef)
	assert.NoError(t, err)

	err = tx.Commit(ctx)
	if err != nil {
		t.Fatal("failed to commit tx: ", err)
	}
}
