package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"cosmic-go/internal/application"
	"cosmic-go/internal/domain"

	"github.com/go-playground/validator/v10"
)

var defaultRequestTimeout = time.Second * 30

type Handler struct {
	svc *application.Service
}

func NewHandler(svc *application.Service) *Handler {
	return &Handler{
		svc: svc,
	}
}

func (h *Handler) Allocate(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultRequestTimeout)
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

	line := &domain.OrderLine{
		SKU:      reqBody.SKU,
		Quantity: reqBody.Quantity,
		OrderId:  domain.OrderID(reqBody.OrderID),
	}

	batchref, err := h.svc.Allocate(ctx, line)

	if errors.Is(err, application.ErrProductNotFound) {
		w.WriteHeader(http.StatusBadRequest)

		response := &BadRequestResponse{
			Message: err.Error(),
		}
		if err = json.NewEncoder(w).Encode(response); err != nil {
			log.Println("failed to structure allocation bad request: ", err)
		}

		return
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("unexpected error: ", err)
		return
	}

	response := &AllocationResponse{
		BatchRef: batchref,
	}

	w.WriteHeader(http.StatusCreated)

	if err = json.NewEncoder(w).Encode(response); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Deallocate(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultRequestTimeout)
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

	err := h.svc.Deallocate(ctx, domain.OrderID(reqBody.OrderID), reqBody.SKU)
	if errors.Is(err, application.ErrProductNotFound) || errors.Is(err, application.ErrInvalidOrderID) {
		w.WriteHeader(http.StatusBadRequest)

		response := &BadRequestResponse{
			Message: err.Error(),
		}
		if err = json.NewEncoder(w).Encode(response); err != nil {
			log.Println("failed to structure allocation bad request: ", err)
		}

		return
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("unexpected error: ", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) AddProduct(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultRequestTimeout)
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

	product := domain.NewProduct(
		reqBody.SKU,
		batches,
		0,
	)

	err := h.svc.AddProduct(ctx, product)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("unexpected error: ", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
