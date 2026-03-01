package handler

import (
	"errors"
	"net/http"

	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/internal/observability"
	"github.com/user/micro-dp/internal/openapi"
	"github.com/user/micro-dp/usecase"
)

type EventHandler struct {
	events  *usecase.EventService
	metrics *observability.EventMetrics
}

func NewEventHandler(events *usecase.EventService, metrics *observability.EventMetrics) *EventHandler {
	return &EventHandler{events: events, metrics: metrics}
}

func (h *EventHandler) Summary(w http.ResponseWriter, r *http.Request) {
	counts, err := h.events.Summary(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	var items []openapi.EventCount
	var total int64
	for name, count := range counts {
		items = append(items, openapi.EventCount{
			EventName: name,
			Count:     count,
		})
		total += count
	}
	if items == nil {
		items = []openapi.EventCount{}
	}

	writeJSON(w, http.StatusOK, openapi.EventsSummaryResponse{
		Counts: items,
		Total:  total,
	})
}

func (h *EventHandler) Ingest(w http.ResponseWriter, r *http.Request) {
	var req openapi.IngestEventRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.EventId == "" {
		writeError(w, http.StatusBadRequest, "event_id is required")
		return
	}
	if req.EventName == "" {
		writeError(w, http.StatusBadRequest, "event_name is required")
		return
	}
	if req.EventTime.IsZero() {
		writeError(w, http.StatusBadRequest, "event_time is required")
		return
	}

	h.metrics.ReceivedTotal.Add(r.Context(), 1)

	var props map[string]any
	if req.Properties != nil {
		props = *req.Properties
	}

	err := h.events.Ingest(r.Context(), req.EventId, req.EventName, props, req.EventTime)
	if err != nil {
		if errors.Is(err, domain.ErrEventDuplicate) {
			h.metrics.DuplicateTotal.Add(r.Context(), 1)
			writeError(w, http.StatusConflict, "event already processed")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.metrics.EnqueuedTotal.Add(r.Context(), 1)

	writeJSON(w, http.StatusAccepted, openapi.IngestEventResponse{
		EventId: req.EventId,
		Status:  openapi.IngestEventResponseStatusAccepted,
	})
}
