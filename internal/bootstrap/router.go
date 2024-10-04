package bootstrap

import (
	"cosmic-go/internal/handler"
	"net/http"
)

func InitializeRouter(h *handler.Handler) *http.ServeMux {
	router := http.NewServeMux()
	router.HandleFunc("/allocate", h.Allocate)
	router.HandleFunc("/deallocate", h.Deallocate)
	return router
}
