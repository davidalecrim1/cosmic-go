package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"cosmic-go/internal/application"
	"cosmic-go/internal/domain"
	eventpublisher "cosmic-go/internal/infra/event_publisher"
	utils "cosmic-go/pkg/utils"

	"github.com/go-playground/validator/v10"
)

var defaultRequestTimeout = time.Second * 30

type AllocationHandler struct {
	eventPublisher *eventpublisher.EventPublisher
}

func NewAllocationHandler(e *eventpublisher.EventPublisher) *AllocationHandler {
	return &AllocationHandler{
		eventPublisher: e,
	}
}

func (h *AllocationHandler) Allocate(w http.ResponseWriter, r *http.Request) {
	_, cancel := context.WithTimeout(context.Background(), defaultRequestTimeout)
	defer cancel()

	reqBody := &AllocationRequest{}

	if err := json.NewDecoder(r.Body).Decode(reqBody); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("failed to decode allocation request: ", err)
		return
	}

	validator := validator.New()
	if err := validator.Struct(reqBody); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("failed to validate add batch request: ", err)
		return
	}

	event := &domain.AllocationRequired{
		OrderID:  reqBody.OrderID,
		SKU:      reqBody.SKU,
		Quantity: reqBody.Quantity,
	}

	errChan := h.eventPublisher.Publish(event)
	batchref := ""

	if err := utils.ErrChanWithAny(
		errChan,
		application.ErrProductNotFound,
	); err != nil {
		w.WriteHeader(http.StatusBadRequest)

		response := &BadRequestResponse{
			Message: err.Error(),
		}
		if err = json.NewEncoder(w).Encode(response); err != nil {
			log.Println("failed to structure allocation bad request: ", err)
		}

		return
	}

	if utils.ErrChanIsNotEmpty(errChan) {
		w.WriteHeader(http.StatusInternalServerError)
		utils.LogErrChan(errChan)
		return
	}

	response := &AllocationResponse{
		BatchRef: batchref,
	}

	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *AllocationHandler) Deallocate(w http.ResponseWriter, r *http.Request) {
	_, cancel := context.WithTimeout(context.Background(), defaultRequestTimeout)
	defer cancel()

	reqBody := &DeallocateRequest{}

	if err := json.NewDecoder(r.Body).Decode(reqBody); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("failed to decode deallocation request: ", err)
		return
	}

	validator := validator.New()
	if err := validator.Struct(reqBody); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("failed to validate deallocation request: ", err)
		return
	}

	event := &domain.DeallocationRequired{
		OrderID: reqBody.OrderID,
		SKU:     reqBody.SKU,
	}

	errChan := h.eventPublisher.Publish(event)
	if err := utils.ErrChanWithAny(
		errChan,
		application.ErrProductNotFound,
		application.ErrInvalidOrderID,
	); err != nil {
		w.WriteHeader(http.StatusBadRequest)

		response := &BadRequestResponse{
			Message: err.Error(),
		}
		if err = json.NewEncoder(w).Encode(response); err != nil {
			log.Println("failed to structure allocation bad request: ", err)
		}

		return
	}

	if utils.ErrChanIsNotEmpty(errChan) {
		w.WriteHeader(http.StatusInternalServerError)

		utils.LogErrChan(errChan)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *AllocationHandler) AddProduct(w http.ResponseWriter, r *http.Request) {
	_, cancel := context.WithTimeout(context.Background(), defaultRequestTimeout)
	defer cancel()

	reqBody := &AddProductRequest{}

	if err := json.NewDecoder(r.Body).Decode(reqBody); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("failed to decode add batch request: ", err)
		return
	}

	validator := validator.New()
	if err := validator.Struct(reqBody); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("failed to validate add batch request: ", err)
		return
	}

	var batches []*domain.Batch = make([]*domain.Batch, 0, len(reqBody.Batches))
	for _, dto := range reqBody.Batches {
		batch := domain.NewBatch(
			dto.Reference,
			reqBody.SKU,
			dto.PurchasedQuantity,
			dto.ETA,
		)
		batches = append(batches, batch)
	}

	event := &domain.CreateProduct{
		SKU:     reqBody.SKU,
		Batches: batches,
	}

	errChan := h.eventPublisher.Publish(event)
	if utils.ErrChanIsNotEmpty(errChan) {
		w.WriteHeader(http.StatusInternalServerError)

		utils.LogErrChan(errChan)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
