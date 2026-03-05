package handler

import (
	"errors"
	"net/http"

	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/internal/openapi"
	"github.com/user/micro-dp/usecase"
)

type ChartHandler struct {
	charts *usecase.ChartService
}

func NewChartHandler(charts *usecase.ChartService) *ChartHandler {
	return &ChartHandler{charts: charts}
}

func (h *ChartHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req openapi.CreateChartRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.DatasetId == "" || req.Measure == "" || req.Dimension == "" {
		writeError(w, http.StatusBadRequest, "name, dataset_id, chart_type, measure, and dimension are required")
		return
	}

	c, err := h.charts.Create(r.Context(), req.Name, string(req.ChartType), req.DatasetId, req.Measure, req.Dimension, req.ConfigJson)
	if err != nil {
		if errors.Is(err, domain.ErrDatasetNotFound) {
			writeError(w, http.StatusNotFound, "dataset not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusCreated, toOpenAPIChart(c))
}

func (h *ChartHandler) List(w http.ResponseWriter, r *http.Request) {
	charts, err := h.charts.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if charts == nil {
		charts = []domain.Chart{}
	}
	items := make([]openapi.Chart, len(charts))
	for i := range charts {
		items[i] = toOpenAPIChart(&charts[i])
	}
	writeJSON(w, http.StatusOK, struct {
		Items []openapi.Chart `json:"items"`
	}{Items: items})
}

func (h *ChartHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}

	c, err := h.charts.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrChartNotFound) {
			writeError(w, http.StatusNotFound, "chart not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, toOpenAPIChart(c))
}

func (h *ChartHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}

	var req openapi.UpdateChartRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.DatasetId == "" || req.Measure == "" || req.Dimension == "" {
		writeError(w, http.StatusBadRequest, "name, dataset_id, chart_type, measure, and dimension are required")
		return
	}

	c, err := h.charts.Update(r.Context(), id, req.Name, string(req.ChartType), req.DatasetId, req.Measure, req.Dimension, req.ConfigJson)
	if err != nil {
		if errors.Is(err, domain.ErrChartNotFound) {
			writeError(w, http.StatusNotFound, "chart not found")
			return
		}
		if errors.Is(err, domain.ErrDatasetNotFound) {
			writeError(w, http.StatusNotFound, "dataset not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, toOpenAPIChart(c))
}

func (h *ChartHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}

	if err := h.charts.Delete(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrChartNotFound) {
			writeError(w, http.StatusNotFound, "chart not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ChartHandler) GetData(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "not yet implemented")
}
