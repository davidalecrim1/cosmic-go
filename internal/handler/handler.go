package handler

import (
	"cosmic-go/internal/domain"
	"cosmic-go/internal/service"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type Handler struct {
	svc *service.Service
}

func NewHandler(svc *service.Service) *Handler {
	return &Handler{
		svc: svc,
	}
}

func (h *Handler) Allocate(w http.ResponseWriter, r *http.Request) {
	reqBody := &AllocationRequest{}

	if err := json.NewDecoder(r.Body).Decode(reqBody); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("failed to decode allocation request: ", err)
		return
	}

	product := domain.Product{SKU: reqBody.SKU}
	line := &domain.OrderLine{
		Product:  product,
		Quantity: reqBody.Quantity}

	batchref, err := h.svc.Allocate(line)

	if errors.Is(err, service.ErrInvalidSku) {
		w.WriteHeader(http.StatusBadRequest)

		response := &AllocationBadRequestResponse{
			Message: err.Error(),
		}
		if err = json.NewEncoder(w).Encode(response); err != nil {
			log.Println("failed to structure allocation bad request: ", err)
		}

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
	OrderID  string `json:"order_id"`
	SKU      string `json:"sku"`
	Quantity int    `json:"quantity"`
}

type AllocationResponse struct {
	BatchRef string `json:"batchref"`
}

type AllocationBadRequestResponse struct {
	Message string `json:"message"`
}
