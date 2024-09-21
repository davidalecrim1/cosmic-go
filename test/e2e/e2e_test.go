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

	t.Run("api returns allocation with 201",
		func(t *testing.T) {
			orderId := "order-001"
			sku := "SMALL-TABLE"
			quantity := 10

			earlyBatch := domain.NewBatch("batch-001", domain.Product{SKU: sku}, 100, time.Now())
			mediumBatch := domain.NewBatch("batch-002", domain.Product{SKU: sku}, 100, time.Now().Add(time.Hour*24))
			otherBatch := domain.NewBatch("batch-003", domain.Product{SKU: sku}, 100, time.Time{})

			insertProductHelper(t, db, sku)
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

			helpers.CleanUpRepositoryHelper(db)
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

			aBody := &handler.AllocationBadRequestResponse{}

			err = json.Unmarshal(body, aBody)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			assert.Equal(t, "invalid sku", aBody.Message)

			helpers.CleanUpRepositoryHelper(db)
		})
}

func getApiEndpointHelper(t *testing.T) string {
	t.Helper()
	return env.GetEnvOrSetDefault("API_URL", "http://localhost:8080")
}

func insertProductHelper(t *testing.T, db *pgxpool.Pool, sku string) {
	t.Helper()
	ctx := context.Background()

	query := `
	INSERT INTO products (sku)
	VALUES ($1);
	`
	_, err := db.Exec(ctx, query, sku)
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
