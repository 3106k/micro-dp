package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/internal/observability"
	"github.com/user/micro-dp/internal/openapi"
	"github.com/user/micro-dp/usecase"
)

type CollectHandler struct {
	events  *usecase.EventService
	plans   *usecase.PlanService
	metrics *observability.EventMetrics
}

func NewCollectHandler(events *usecase.EventService, plans *usecase.PlanService, metrics *observability.EventMetrics) *CollectHandler {
	return &CollectHandler{events: events, plans: plans, metrics: metrics}
}

func (h *CollectHandler) Collect(w http.ResponseWriter, r *http.Request) {
	if err := h.plans.CheckEventsQuota(r.Context()); err != nil {
		if errors.Is(err, domain.ErrQuotaExceeded) {
			writeError(w, http.StatusPaymentRequired, "events quota exceeded")
			return
		}
	}

	var req openapi.CollectEventsRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.Events) == 0 {
		writeError(w, http.StatusBadRequest, "events array is required")
		return
	}

	accepted := 0
	errCount := 0

	for _, ev := range req.Events {
		if ev.EventId == "" || ev.EventName == "" || ev.EventTime.IsZero() {
			errCount++
			continue
		}

		h.metrics.ReceivedTotal.Add(r.Context(), 1)

		var props map[string]any
		if ev.Properties != nil {
			props = *ev.Properties
		}

		var sessionID string
		if ev.SessionId != nil {
			sessionID = *ev.SessionId
		}

		var ctxJSON string
		if ev.Context != nil {
			b, _ := json.Marshal(ev.Context)
			ctxJSON = string(b)
		}

		err := h.events.Ingest(r.Context(), ev.EventId, ev.EventName, sessionID, props, ev.EventTime, ctxJSON)
		if err != nil {
			if errors.Is(err, domain.ErrEventDuplicate) {
				h.metrics.DuplicateTotal.Add(r.Context(), 1)
			}
			errCount++
			continue
		}

		h.metrics.EnqueuedTotal.Add(r.Context(), 1)
		accepted++
	}

	writeJSON(w, http.StatusAccepted, openapi.CollectEventsResponse{
		Accepted: accepted,
		Errors:   errCount,
	})
}
