package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"cosmic-go/internal/domain"
	"cosmic-go/internal/service"

	"github.com/go-playground/validator/v10"
)

var defaultRequestTimeout = time.Second * 30

type Handler struct {
	svc *service.Service
}

func NewHandler(svc *service.Service) *Handler {
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

	product := domain.Product{SKU: reqBody.SKU}
	line := &domain.OrderLine{
		Product:  product,
		Quantity: reqBody.Quantity,
		OrderId:  reqBody.OrderID,
	}

	batchref, err := h.svc.Allocate(ctx, line)

	if errors.Is(err, service.ErrInvalidSku) {
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
		log.Println("failed to validate add batch request: ", err)
		return
	}

	err := h.svc.Deallocate(ctx, reqBody.OrderID, reqBody.SKU)
	if errors.Is(err, service.ErrInvalidSku) || errors.Is(err, service.ErrInvalidOrderID) {
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

func (h *Handler) AddBatch(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultRequestTimeout)
	defer cancel()

	reqBody := &AddBatchRequest{}

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

	batch := domain.NewBatch(
		reqBody.Reference,
		domain.Product(reqBody.Product),
		reqBody.PurchasedQuantity,
		reqBody.ETA,
	)

	err := h.svc.AddBatch(ctx, batch)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("unexpected error: ", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
