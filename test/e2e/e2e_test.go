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

	"cosmic-go/internal/handler"
	"cosmic-go/internal/infra/database"
	"cosmic-go/internal/server"
	"cosmic-go/test/helpers"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var (
	db     *gorm.DB
	ts     *httptest.Server
	router *http.ServeMux
)

func TestMain(m *testing.M) {
	db = database.InitializeDatabase()
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	router = server.InitializeServer(db)
	ts = httptest.NewServer(router)
	defer ts.Close()

	code := m.Run()
	os.Exit(code)
}

func TestE2E_Allocation(t *testing.T) {
	t.Run("api valid allocation returns 201",
		func(t *testing.T) {
			earlyEta := time.Now()
			earlyBatch := &handler.BatchDTO{
				Reference:         "batch-001",
				PurchasedQuantity: 100,
				ETA:               &earlyEta,
			}

			mediumEta := time.Now().Add(time.Hour * 24)
			mediumBatch := &handler.BatchDTO{
				Reference:         "batch-002",
				PurchasedQuantity: 100,
				ETA:               &mediumEta,
			}

			inStockBatch := &handler.BatchDTO{
				Reference:         "batch-003",
				PurchasedQuantity: 100,
			}

			productRequest := handler.AddProductRequest{
				SKU:     "SMALL-TABLE",
				Batches: []*handler.BatchDTO{earlyBatch, mediumBatch, inStockBatch},
			}

			addProductRequestPostWrapper(t, ts, productRequest, http.StatusCreated)

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
			assert.Equal(t, "product not found", allocationResponse.Message)

			t.Cleanup(func() {
				helpers.CleanUpRepositoryHelper(db)
			})
		})
}

func TestE2E_Deallocation(t *testing.T) {
	t.Run("api returns 200 for deallocate",
		func(t *testing.T) {
			validEta := time.Now().Add(time.Hour * 24)
			validBatch := &handler.BatchDTO{
				Reference:         "batch-001",
				PurchasedQuantity: 100,
				ETA:               &validEta,
			}

			validProduct := handler.AddProductRequest{
				SKU:     "SMALL-TABLE",
				Batches: []*handler.BatchDTO{validBatch},
			}

			addProductRequestPostWrapper(t, ts, validProduct, http.StatusCreated)

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
			_ = DeallocateRequestPostWrapper(t, ts, validDeallocateRequest, http.StatusOK)

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

func TestE2E_AddProduct(t *testing.T) {
	t.Run("api returns 201 for adding a new batch WITH eta", func(t *testing.T) {
		validEta := time.Now().Add(time.Hour * 24)
		validBatch := &handler.BatchDTO{
			Reference:         "batch-001",
			PurchasedQuantity: 100,
			ETA:               &validEta,
		}

		validProduct := handler.AddProductRequest{
			SKU:     "SMALL-TABLE",
			Batches: []*handler.BatchDTO{validBatch},
		}

		addProductRequestPostWrapper(t, ts, validProduct, http.StatusCreated)

		t.Cleanup(func() {
			helpers.CleanUpRepositoryHelper(db)
		})
	})

	t.Run("api returns 201 for adding a new batch WITHOUT eta", func(t *testing.T) {
		validBatch := &handler.BatchDTO{
			Reference:         "batch-001",
			PurchasedQuantity: 100,
			ETA:               nil,
		}

		validProduct := handler.AddProductRequest{
			SKU:     "SMALL-TABLE",
			Batches: []*handler.BatchDTO{validBatch},
		}

		addProductRequestPostWrapper(t, ts, validProduct, http.StatusCreated)

		t.Cleanup(func() {
			helpers.CleanUpRepositoryHelper(db)
		})
	})

	t.Run("api returns 400 for adding a batch with invalid request", func(t *testing.T) {
		invalidRequestBody := handler.AddProductRequest{
			SKU:     "",
			Batches: []*handler.BatchDTO{},
		}

		addProductRequestPostWrapper(t, ts, invalidRequestBody, http.StatusBadRequest)

		t.Cleanup(func() {
			helpers.CleanUpRepositoryHelper(db)
		})
	})
}

func addProductRequestPostWrapper(
	t *testing.T,
	ts *httptest.Server,
	requestBody handler.AddProductRequest,
	expectedStatus int,
) {
	body, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	resp, err := http.Post(ts.URL+"/products",
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
