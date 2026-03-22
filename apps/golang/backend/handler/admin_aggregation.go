package handler

import (
	"net/http"

	"github.com/user/micro-dp/internal/openapi"
	"github.com/user/micro-dp/usecase"
)

type AdminAggregationHandler struct {
	backfill *usecase.AggregationBackfillService
}

func NewAdminAggregationHandler(backfill *usecase.AggregationBackfillService) *AdminAggregationHandler {
	return &AdminAggregationHandler{backfill: backfill}
}

func (h *AdminAggregationHandler) TriggerBackfill(w http.ResponseWriter, r *http.Request) {
	var req openapi.BackfillRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	startDate := req.StartDate.Time
	endDate := req.EndDate.Time

	if endDate.Before(startDate) {
		writeError(w, http.StatusBadRequest, "end_date must be on or after start_date")
		return
	}

	tenantID := ""
	if req.TenantId != nil {
		tenantID = *req.TenantId
	}
	force := false
	if req.Force != nil {
		force = *req.Force
	}

	result, err := h.backfill.TriggerBackfill(r.Context(), tenantID, startDate, endDate, force)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "backfill failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, openapi.BackfillResponse{
		Enqueued: result.Enqueued,
		Skipped:  result.Skipped,
	})
}
