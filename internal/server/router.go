package server

import (
	"net/http"

	"cosmic-go/internal/handler"
)

func InitializeRouter(
	h *handler.AllocationHandler,
	hv *handler.AllocationViewHandler,
) *http.ServeMux {
	router := http.NewServeMux()
	// TODO: make this more restful later
	router.HandleFunc("/allocate", h.Allocate)
	router.HandleFunc("/deallocate", h.Deallocate)
	router.HandleFunc("POST /products", h.AddProduct)
	router.HandleFunc("GET /allocations/{id}", hv.GetAllocation)
	return router
}
