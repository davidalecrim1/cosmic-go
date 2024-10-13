//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"cosmic-go/internal/domain"
	"cosmic-go/internal/handler"
	"cosmic-go/internal/infra/database"
	"cosmic-go/internal/server"
	"cosmic-go/test/helpers"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
)

var (
	db     *pgxpool.Pool
	ts     *httptest.Server
	router *http.ServeMux
)

func TestMain(m *testing.M) {
	db = database.InitializeDatabase()
	defer db.Close()

	router = server.InitializeServer(db)
	ts = httptest.NewServer(router)
	defer ts.Close()

	code := m.Run()
	os.Exit(code)
}

func TestE2E_Allocation(t *testing.T) {
	t.Run("api valid allocation returns 201",
		func(t *testing.T) {
			earlyBatchRequest := handler.AddBatchRequest{
				Reference:         "batch-001",
				Product:           handler.ProductDTO{SKU: "SMALL-TABLE"},
				PurchasedQuantity: 100,
				ETA:               time.Now(),
			}
			addBatchRequestPostWrapper(t, ts, earlyBatchRequest, http.StatusCreated)

			mediumBatchRequest := handler.AddBatchRequest{
				Reference:         "batch-002",
				Product:           handler.ProductDTO{SKU: "SMALL-TABLE"},
				PurchasedQuantity: 100,
				ETA:               time.Now().Add(time.Hour * 24),
			}
			addBatchRequestPostWrapper(t, ts, mediumBatchRequest, http.StatusCreated)

			inStockBatchRequest := handler.AddBatchRequest{
				Reference:         "batch-003",
				Product:           handler.ProductDTO{SKU: "SMALL-TABLE"},
				PurchasedQuantity: 100,
			}
			addBatchRequestPostWrapper(t, ts, inStockBatchRequest, http.StatusCreated)

			orderId := "order-001"
			orderQuantity := 10

			allocationRequest := handler.AllocationRequest{
				OrderID:  orderId,
				SKU:      "SMALL-TABLE",
				Quantity: orderQuantity,
			}

			respBody := allocateRequestPostWrapper(t, ts, allocationRequest, http.StatusCreated)
			respAllocation := &handler.AllocationResponse{}
			err := json.Unmarshal(respBody, respAllocation)
			assert.NoError(t, err)

			expectedBatch := "batch-003"
			assert.Equal(t, expectedBatch, respAllocation.BatchRef)

			t.Cleanup(func() {
				helpers.CleanUpRepositoryHelper(db)
			})
		})

	t.Run("api invalid allocation returns 400 and error message",
		func(t *testing.T) {
			invalidRequestBody := handler.AllocationRequest{
				OrderID:  "unknownOrderId",
				SKU:      "unknownSku",
				Quantity: 10,
			}

			respBody := allocateRequestPostWrapper(t, ts, invalidRequestBody, http.StatusBadRequest)
			allocationResponse := &handler.BadRequestResponse{}

			err := json.Unmarshal(respBody, allocationResponse)
			assert.NoError(t, err)
			assert.Equal(t, "invalid sku", allocationResponse.Message)

			t.Cleanup(func() {
				helpers.CleanUpRepositoryHelper(db)
			})
		})
}

func TestE2E_Deallocation(t *testing.T) {
	t.Run("api returns 200 for deallocate",
		func(t *testing.T) {
			validBatch := handler.AddBatchRequest{
				Reference:         "batch-001",
				Product:           handler.ProductDTO{SKU: "SMALL-TABLE"},
				PurchasedQuantity: 20,
				ETA:               time.Now(),
			}
			addBatchRequestPostWrapper(t, ts, validBatch, http.StatusCreated)

			validAllocation := handler.AllocationRequest{
				OrderID:  "order-001",
				Quantity: 10,
				SKU:      "SMALL-TABLE",
			}
			_ = allocateRequestPostWrapper(t, ts, validAllocation, http.StatusCreated)

			validDeallocateRequest := handler.DeallocateRequest{
				OrderID: "order-001",
				SKU:     "SMALL-TABLE",
			}
			body := DeallocateRequestPostWrapper(t, ts, validDeallocateRequest, http.StatusOK)
			t.Log(string(body))

			t.Cleanup(func() {
				helpers.CleanUpRepositoryHelper(db)
			})
		})

	t.Run("invalid order id for deallocation",
		func(t *testing.T) {
			invalidOrderId := "invalid-order-id"

			validRequestBody := handler.DeallocateRequest{
				OrderID: invalidOrderId,
				SKU:     "SMALL-TABLE",
			}

			_ = DeallocateRequestPostWrapper(t, ts, validRequestBody, http.StatusBadRequest)

			t.Cleanup(func() {
				helpers.CleanUpRepositoryHelper(db)
			})
		})
}

func TestE2E_AddBatch(t *testing.T) {
	t.Run("api returns 201 for adding a new batch WITH eta", func(t *testing.T) {
		product := domain.Product{SKU: "SMALL-TABLE"}
		eta := time.Now()
		batch := domain.NewBatch("batch-001", product, 100, &eta)

		validRequestBody := handler.AddBatchRequest{
			Reference:         batch.Reference,
			Product:           handler.ProductDTO{SKU: product.SKU},
			PurchasedQuantity: batch.PurchasedQuantity,
		}
		if batch.GetETA() != nil {
			validRequestBody.ETA = *batch.GetETA()
		}

		addBatchRequestPostWrapper(t, ts, validRequestBody, http.StatusCreated)

		t.Cleanup(func() {
			helpers.CleanUpRepositoryHelper(db)
		})
	})

	t.Run("api returns 201 for adding a new batch WITHOUT eta", func(t *testing.T) {
		product := domain.Product{SKU: "SMALL-TABLE"}
		batch := domain.NewBatch("batch-001", product, 100, nil)

		validRequestBody := handler.AddBatchRequest{
			Reference:         batch.Reference,
			Product:           handler.ProductDTO{SKU: product.SKU},
			PurchasedQuantity: batch.PurchasedQuantity,
		}
		if batch.GetETA() != nil {
			validRequestBody.ETA = *batch.GetETA()
		}

		addBatchRequestPostWrapper(t, ts, validRequestBody, http.StatusCreated)

		t.Cleanup(func() {
			helpers.CleanUpRepositoryHelper(db)
		})
	})

	t.Run("api returns 400 for adding a batch with invalid request", func(t *testing.T) {
		invalidRequestBody := handler.AddBatchRequest{
			Reference:         "",
			Product:           handler.ProductDTO{SKU: "INVALID_SKU"},
			PurchasedQuantity: -1,
		}

		addBatchRequestPostWrapper(t, ts, invalidRequestBody, http.StatusBadRequest)

		t.Cleanup(func() {
			helpers.CleanUpRepositoryHelper(db)
		})
	})
}

func addBatchRequestPostWrapper(
	t *testing.T,
	ts *httptest.Server,
	requestBody handler.AddBatchRequest,
	expectedStatus int,
) {
	body, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	resp, err := http.Post(ts.URL+"/batches",
		"application/json",
		bytes.NewBuffer(body),
	)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, expectedStatus, resp.StatusCode)
}

func allocateRequestPostWrapper(
	t *testing.T,
	ts *httptest.Server,
	requestBody handler.AllocationRequest,
	expectedStatus int,
) (respBody []byte) {
	body, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	resp, err := http.Post(
		ts.URL+"/allocate",
		"application/json",
		bytes.NewBuffer(body),
	)

	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, expectedStatus, resp.StatusCode)

	respBody, err = io.ReadAll(resp.Body)
	assert.NoError(t, err)
	return respBody
}

func DeallocateRequestPostWrapper(
	t *testing.T,
	ts *httptest.Server,
	requestBody handler.DeallocateRequest,
	expectedStatus int,
) (respBody []byte) {
	body, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	resp, err := http.Post(ts.URL+"/deallocate",
		"application/json",
		bytes.NewBuffer(body),
	)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, expectedStatus, resp.StatusCode)

	respBody, err = io.ReadAll(resp.Body)
	assert.NoError(t, err)

	return respBody
}
