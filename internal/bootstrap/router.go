package bootstrap

import (
	"cosmic-go/cmd/api/handler"
	"net/http"
)

func InitializeRouter(h *handler.Handler) *http.ServeMux {
	router := http.NewServeMux()
	router.HandleFunc("/allocate", h.Allocate)
	return router
}
