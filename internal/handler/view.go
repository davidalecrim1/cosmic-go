// the goal of this package is only to serve as entrypoint for
// reads based on the principle CQS for the application
package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"gorm.io/gorm"
)

type AllocationViewHandler struct {
	db *gorm.DB
}

func NewAllocationViewHandler(db *gorm.DB) *AllocationViewHandler {
	return &AllocationViewHandler{db: db}
}

// GET /allocations/{id}
func (hv *AllocationViewHandler) GetAllocation(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultRequestTimeout)
	defer cancel()

	orderId := r.PathValue("id")

	if orderId == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	query := `
	SELECT ol.sku as sku, b.reference as batch_reference
	FROM allocations AS a
	JOIN batches AS b ON a.batch_reference = b.reference
	JOIN order_lines AS ol ON a.order_line_id = ol.id
	WHERE ol.order_id = ?
	ORDER BY ol.order_id ASC;
	`

	var results []AllocationResponse
	err := hv.db.WithContext(ctx).Raw(query, orderId).Scan(&results).Error
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := AllocationsResponse{
		Allocations: make([]*AllocationResponse, len(results)),
	}

	for i, res := range results {
		response.Allocations[i] = &res
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
