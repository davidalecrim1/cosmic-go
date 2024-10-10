package server

import (
	"net/http"

	"cosmic-go/internal/handler"
)

func InitializeRouter(h *handler.Handler) *http.ServeMux {
	router := http.NewServeMux()
	router.HandleFunc("/allocate", h.Allocate)
	router.HandleFunc("/deallocate", h.Deallocate)
	router.HandleFunc("POST /batches", h.AddBatch)
	return router
}
