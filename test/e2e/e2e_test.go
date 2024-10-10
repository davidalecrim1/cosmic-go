//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"cosmic-go/internal/bootstrap"
	"cosmic-go/internal/domain"
	"cosmic-go/internal/handler"
	"cosmic-go/pkg/env"
	"cosmic-go/test/helpers"

	"github.com/stretchr/testify/assert"
)

func TestE2E_Allocation(t *testing.T) {
	db := bootstrap.InitializeDatabase()
	defer db.Close()

	t.Run("api valid allocation returns 201",
		func(t *testing.T) {
			orderId := "order-001"
			product := domain.Product{SKU: "SMALL-TABLE"}
			quantity := 10

			earlyBatch := domain.NewBatch("batch-001", product, 100, time.Now())
			mediumBatch := domain.NewBatch("batch-002", product, 100, time.Now().Add(time.Hour*24))
			otherBatch := domain.NewBatch("batch-003", product, 100, time.Time{})

			helpers.CreateTestBatchData(
				t,
				db,
				helpers.WithProduct(&product),
				helpers.WithBatch(earlyBatch),
				helpers.WithBatch(mediumBatch),
				helpers.WithBatch(otherBatch),
			)

			requestBody, err := json.Marshal(map[string]any{
				"orderid":  orderId,
				"sku":      product.SKU,
				"quantity": quantity,
			})
			assert.NoError(t, err)

			apiEndpoint := getApiEndpoint(t)
			resp, err := http.Post(
				apiEndpoint+"/allocate",
				"application/json",
				bytes.NewBuffer(requestBody),
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

	t.Run("api invalid allocation returns 400 and error message",
		func(t *testing.T) {
			unknownSku, unknownOrderId, quantity := "UNKNOWN-SKU-001", "unknown-order-001", 10
			apiEndpoint := getApiEndpoint(t)

			requestBody, err := json.Marshal(map[string]any{
				"orderid":  unknownOrderId,
				"sku":      unknownSku,
				"quantity": quantity,
			})
			assert.NoError(t, err)

			resp, err := http.Post(apiEndpoint+"/allocate",
				"application/json",
				bytes.NewBuffer(requestBody))

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
}

func TestE2E_Deallocation(t *testing.T) {
	db := bootstrap.InitializeDatabase()
	defer db.Close()

	t.Run("api returns 200 for deallocate",
		func(t *testing.T) {
			product := domain.Product{SKU: "SMALL-TABLE"}

			orderLine := &domain.OrderLine{
				Product:  product,
				Quantity: 5,
				OrderId:  "order-001",
			}

			batch := domain.NewBatch("batch-001", product, 20, time.Now())

			helpers.CreateTestBatchData(
				t,
				db,
				helpers.WithBatch(batch),
				helpers.WithProduct(&product),
				helpers.WithOrderLine(orderLine),
				helpers.WithAllocation(orderLine.OrderId, batch.Reference),
			)

			apiEndpoint := getApiEndpoint(t)
			requestBody, err := json.Marshal(map[string]any{
				"orderid": orderLine.OrderId,
				"sku":     product.SKU,
			})
			assert.NoError(t, err)

			resp, err := http.Post(apiEndpoint+"/deallocate",
				"application/json",
				bytes.NewBuffer(requestBody),
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
			product := domain.Product{SKU: "SMALL-TABLE"}

			orderLine := &domain.OrderLine{
				Product:  product,
				Quantity: 5,
				OrderId:  "order-001",
			}

			batch := domain.NewBatch("batch-001", product, 20, time.Now())

			helpers.CreateTestBatchData(
				t,
				db,
				helpers.WithProduct(&product),
				helpers.WithBatch(batch),
				helpers.WithOrderLine(orderLine),
				helpers.WithAllocation(orderLine.OrderId, batch.Reference),
			)

			invalidOrderId := "order-002"
			requestBody, err := json.Marshal(map[string]any{
				"orderid": invalidOrderId,
				"sku":     product.SKU,
			})

			assert.NoError(t, err)

			apiEndpoint := getApiEndpoint(t)
			resp, err := http.Post(apiEndpoint+"/deallocate",
				"application/json",
				bytes.NewBuffer(requestBody),
			)
			assert.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

			t.Cleanup(func() {
				helpers.CleanUpRepositoryHelper(db)
			})
		})
}

func TestE2E_AddBatch(t *testing.T) {
	db := bootstrap.InitializeDatabase()
	defer db.Close()

	t.Run("api returns 201 for adding a new batch WITH eta", func(t *testing.T) {
		product := domain.Product{SKU: "SMALL-TABLE"}
		batch := domain.NewBatch("batch-001", product, 100, time.Now())
		validRequestBody := map[string]any{
			"reference":          batch.Reference,
			"product":            product,
			"purchased_quantity": batch.PurchasedQuantity,
		}
		addBatch(t, validRequestBody, http.StatusCreated)

		t.Cleanup(func() {
			helpers.CleanUpRepositoryHelper(db)
		})
	})

	t.Run("api returns 201 for adding a new batch WITHOUT eta", func(t *testing.T) {
		product := domain.Product{SKU: "SMALL-TABLE"}
		batch := domain.NewBatchWithoutETA("batch-001", product, 100)
		validRequestBody := map[string]any{
			"reference":          batch.Reference,
			"product":            product,
			"purchased_quantity": batch.PurchasedQuantity,
		}
		addBatch(t, validRequestBody, http.StatusCreated)

		t.Cleanup(func() {
			helpers.CleanUpRepositoryHelper(db)
		})
	})

	t.Run("api returns 400 for adding a batch with invalid request", func(t *testing.T) {
		invalidRequestBody := map[string]any{
			"reference":          "",                                 // Invalid reference
			"product":            domain.Product{SKU: "INVALID-SKU"}, // Invalid product SKU
			"purchased_quantity": -1,                                 // Invalid quantity
		}
		addBatch(t, invalidRequestBody, http.StatusBadRequest)
	})
}

func getApiEndpoint(t *testing.T) string {
	t.Helper()
	return env.GetEnvOrSetDefault("API_URL", "http://localhost:8080")
}

func addBatch(
	t *testing.T,
	requestBody map[string]any,
	expectedStatus int,
) {
	apiEndpoint := getApiEndpoint(t)
	body, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	resp, err := http.Post(apiEndpoint+"/batches",
		"application/json",
		bytes.NewBuffer(body),
	)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, expectedStatus, resp.StatusCode)
}
