package handler

import (
	"context"
	"cosmic-go/internal/domain"
	"cosmic-go/internal/service"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
)

type Handler struct {
	svc *service.Service
}

func NewHandler(svc *service.Service) *Handler {
	return &Handler{
		svc: svc,
	}
}

var (
	defaultRequestTimeout = time.Second * 30
)

func (h *Handler) Allocate(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultRequestTimeout)
	defer cancel()

	reqBody := &AllocationRequest{}

	if err := json.NewDecoder(r.Body).Decode(reqBody); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("failed to decode allocation request: ", err)
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

type AllocationRequest struct {
	OrderID  string `json:"orderid"`
	SKU      string `json:"sku"`
	Quantity int    `json:"quantity"`
}

type AllocationResponse struct {
	BatchRef string `json:"batchref"`
}

type BadRequestResponse struct {
	Message string `json:"message"`
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

type DeallocateRequest struct {
	OrderID string `json:"orderid"`
	SKU     string `json:"sku"`
}
