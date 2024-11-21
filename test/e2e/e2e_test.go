//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"cosmic-go/internal/domain"
	"cosmic-go/internal/handler"
	"cosmic-go/internal/infra/database"
	"cosmic-go/internal/infra/messagepublisher"
	"cosmic-go/internal/server"

	"cosmic-go/test/helpers"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	testcontainerRedis "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"

	"gorm.io/gorm"
)

var (
	db          *gorm.DB
	ts          *httptest.Server
	redisClient *redis.Client
)

func TestMain(m *testing.M) {
	db = database.NewDatabase()
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rdContainer := createTestRedis(ctx)
	defer func() {
		err := rdContainer.Terminate(ctx)
		if err != nil {
			log.Fatalf("failed to terminate container: %v", err)
		}
	}()

	host, err := rdContainer.Host(ctx)
	if err != nil {
		log.Fatalf("failed to retrieve the host string: %v", err)
	}

	port, err := rdContainer.MappedPort(ctx, "6379/tcp")
	if err != nil {
		log.Fatalf("failed to retrieve the port string: %v", err)
	}

	conn := strings.Join([]string{host, port.Port()}, ":")
	redisClient = messagepublisher.InitializeRedis(conn)

	s := server.NewServer()
	s.InitializeDependencies(ctx, db, redisClient)
	ts = httptest.NewServer(s.Router)
	defer ts.Close()

	code := m.Run()
	os.Exit(code)
}

func createTestRedis(ctx context.Context) (container *testcontainerRedis.RedisContainer) {
	rdContainer, err := testcontainerRedis.Run(
		ctx,
		"docker.io/redis:7.4",
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections tcp").
				WithOccurrence(1).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		log.Fatalf("failed to create redis for tests: %v", err)
	}

	return rdContainer
}

func TestE2E_Allocation(t *testing.T) {
	t.Run("api valid allocation returns 202",
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

			_ = allocateRequestPostWrapper(t, ts, allocationRequest, http.StatusCreated)

			respBody := allocationsRequestGetWrapper(t, ts, "order-001")
			var response handler.AllocationsResponse

			err := json.Unmarshal(respBody, &response)
			assert.NoError(t, err)

			assert.Equal(t, response.Allocations[0].SKU, "SMALL-TABLE")

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
			_ = deallocateRequestPostWrapper(t, ts, validDeallocateRequest, http.StatusOK)

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

			_ = deallocateRequestPostWrapper(t, ts, validRequestBody, http.StatusBadRequest)

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

func TestE2E_ChangeBatchQuantityEvent(t *testing.T) {
	t.Run("trigger ChangeBatchQuantity command to reallocate",
		func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
			defer cancel()

			t.Cleanup(func() {
				helpers.CleanUpRepositoryHelper(db)
			})

			tomorrowEta := time.Now().Add(time.Hour * 24)
			product := handler.AddProductRequest{
				SKU: "SMALL-TABLE",
				Batches: []*handler.BatchDTO{
					{
						Reference:         "batch-001",
						PurchasedQuantity: 25,
						ETA:               nil,
					},
					{
						Reference:         "batch-002",
						PurchasedQuantity: 50,
						ETA:               &tomorrowEta,
					},
				},
			}
			addProductRequestPostWrapper(t, ts, product, http.StatusCreated)

			allocation := handler.AllocationRequest{
				OrderID:  "order-001",
				SKU:      "SMALL-TABLE",
				Quantity: 20,
			}

			_ = allocateRequestPostWrapper(t, ts, allocation, http.StatusCreated)

			expectedExternalEvent := &domain.Allocated{}
			expectedEventChannel := redisClient.
				Subscribe(ctx, expectedExternalEvent.
					GetEventName()).
				Channel()

			command := &domain.ChangeBatchQuantity{
				BatchReference:    "batch-001",
				ChangedToQuantity: 10,
			}
			commandAsJson, err := command.ToJson()
			assert.NoError(t, err)

			err = redisClient.Publish(ctx, command.GetCommandName(), commandAsJson).Err()
			assert.NoError(t, err)

			for {
				select {
				case <-ctx.Done():
					t.Error("timeout waiting for external event")
					return
				case event := <-expectedEventChannel:
					resultedEvent, err := domain.NewAllocatedEventFromJson(event.Payload)
					assert.NoError(t, err)
					assert.Equal(t, "batch-002", resultedEvent.BatchRef)
					return
				}
			}
		})
}

func TestE2E_Allocations(t *testing.T) {
	t.Run("api returns 200 for get allocations", func(t *testing.T) {
		t.Cleanup(func() {
			helpers.CleanUpRepositoryHelper(db)
		})

		validEta := time.Now().Add(time.Hour * 24)
		firstValidBatch := &handler.BatchDTO{
			Reference:         "batch-001",
			PurchasedQuantity: 100,
			ETA:               &validEta,
		}

		firstValidProduct := handler.AddProductRequest{
			SKU:     "SMALL-TABLE",
			Batches: []*handler.BatchDTO{firstValidBatch},
		}

		secondValidBatch := &handler.BatchDTO{
			Reference:         "batch-002",
			PurchasedQuantity: 100,
			ETA:               nil,
		}

		secondValidProduct := handler.AddProductRequest{
			SKU:     "LARGE-TABLE",
			Batches: []*handler.BatchDTO{secondValidBatch},
		}

		addProductRequestPostWrapper(t, ts, firstValidProduct, http.StatusCreated)
		addProductRequestPostWrapper(t, ts, secondValidProduct, http.StatusCreated)

		firstValidAllocation := handler.AllocationRequest{
			OrderID:  "order-001",
			Quantity: 20,
			SKU:      "SMALL-TABLE",
		}

		secondValidAllocation := handler.AllocationRequest{
			OrderID:  "order-001",
			Quantity: 40,
			SKU:      "LARGE-TABLE",
		}
		_ = allocateRequestPostWrapper(t, ts, firstValidAllocation, http.StatusCreated)
		_ = allocateRequestPostWrapper(t, ts, secondValidAllocation, http.StatusCreated)

		var response handler.AllocationsResponse
		respBody := allocationsRequestGetWrapper(t, ts, "order-001")
		err := json.Unmarshal(respBody, &response)
		assert.NoError(t, err)

		assert.Equal(t, response.Allocations[0].SKU, firstValidAllocation.SKU)
		assert.Equal(t, response.Allocations[0].BatchReference, firstValidBatch.Reference)

		assert.Equal(t, response.Allocations[1].SKU, secondValidAllocation.SKU)
		assert.Equal(t, response.Allocations[1].BatchReference, secondValidBatch.Reference)
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
		ts.URL+"/products/allocations/allocate",
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

func deallocateRequestPostWrapper(
	t *testing.T,
	ts *httptest.Server,
	requestBody handler.DeallocateRequest,
	expectedStatus int,
) (respBody []byte) {
	body, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	resp, err := http.Post(ts.URL+"/products/allocations/deallocate",
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

func allocationsRequestGetWrapper(
	t *testing.T,
	ts *httptest.Server,
	orderID string,
) (respBody []byte) {
	resp, err := http.Get(ts.URL + "/products/allocations/" + orderID)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	respBody, err = io.ReadAll(resp.Body)
	assert.NoError(t, err)

	return respBody
}
