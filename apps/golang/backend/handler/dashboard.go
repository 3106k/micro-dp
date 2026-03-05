package handler

import (
	"errors"
	"net/http"

	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/internal/openapi"
	"github.com/user/micro-dp/usecase"
)

type DashboardHandler struct {
	dashboards *usecase.DashboardService
}

func NewDashboardHandler(dashboards *usecase.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashboards: dashboards}
}

func (h *DashboardHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req openapi.CreateDashboardRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	d, err := h.dashboards.Create(r.Context(), req.Name, req.Description)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusCreated, toOpenAPIDashboard(d))
}

func (h *DashboardHandler) List(w http.ResponseWriter, r *http.Request) {
	dashboards, err := h.dashboards.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if dashboards == nil {
		dashboards = []domain.Dashboard{}
	}
	items := make([]openapi.Dashboard, len(dashboards))
	for i := range dashboards {
		items[i] = toOpenAPIDashboard(&dashboards[i])
	}
	writeJSON(w, http.StatusOK, struct {
		Items []openapi.Dashboard `json:"items"`
	}{Items: items})
}

func (h *DashboardHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}

	d, err := h.dashboards.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrDashboardNotFound) {
			writeError(w, http.StatusNotFound, "dashboard not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, toOpenAPIDashboard(d))
}

func (h *DashboardHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}

	var req openapi.UpdateDashboardRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	d, err := h.dashboards.Update(r.Context(), id, req.Name, req.Description)
	if err != nil {
		if errors.Is(err, domain.ErrDashboardNotFound) {
			writeError(w, http.StatusNotFound, "dashboard not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, toOpenAPIDashboard(d))
}

func (h *DashboardHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}

	if err := h.dashboards.Delete(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrDashboardNotFound) {
			writeError(w, http.StatusNotFound, "dashboard not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *DashboardHandler) CreateWidget(w http.ResponseWriter, r *http.Request) {
	dashboardID := r.PathValue("dashboard_id")
	if dashboardID == "" {
		writeError(w, http.StatusBadRequest, "missing dashboard_id")
		return
	}

	var req openapi.CreateDashboardWidgetRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.ChartId == "" {
		writeError(w, http.StatusBadRequest, "chart_id is required")
		return
	}

	widget, err := h.dashboards.CreateWidget(r.Context(), dashboardID, req.ChartId, req.Position)
	if err != nil {
		if errors.Is(err, domain.ErrDashboardNotFound) {
			writeError(w, http.StatusNotFound, "dashboard not found")
			return
		}
		if errors.Is(err, domain.ErrChartNotFound) {
			writeError(w, http.StatusNotFound, "chart not found")
			return
		}
		if errors.Is(err, domain.ErrWidgetDuplicate) {
			writeError(w, http.StatusConflict, "widget already exists for this chart on this dashboard")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusCreated, toOpenAPIWidget(widget))
}

func (h *DashboardHandler) ListWidgets(w http.ResponseWriter, r *http.Request) {
	dashboardID := r.PathValue("dashboard_id")
	if dashboardID == "" {
		writeError(w, http.StatusBadRequest, "missing dashboard_id")
		return
	}

	widgets, err := h.dashboards.ListWidgets(r.Context(), dashboardID)
	if err != nil {
		if errors.Is(err, domain.ErrDashboardNotFound) {
			writeError(w, http.StatusNotFound, "dashboard not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if widgets == nil {
		widgets = []domain.DashboardWidget{}
	}
	items := make([]openapi.DashboardWidget, len(widgets))
	for i := range widgets {
		items[i] = toOpenAPIWidget(&widgets[i])
	}
	writeJSON(w, http.StatusOK, struct {
		Items []openapi.DashboardWidget `json:"items"`
	}{Items: items})
}

func (h *DashboardHandler) DeleteWidget(w http.ResponseWriter, r *http.Request) {
	dashboardID := r.PathValue("dashboard_id")
	widgetID := r.PathValue("widget_id")
	if dashboardID == "" || widgetID == "" {
		writeError(w, http.StatusBadRequest, "missing dashboard_id or widget_id")
		return
	}

	if err := h.dashboards.DeleteWidget(r.Context(), dashboardID, widgetID); err != nil {
		if errors.Is(err, domain.ErrDashboardNotFound) {
			writeError(w, http.StatusNotFound, "dashboard not found")
			return
		}
		if errors.Is(err, domain.ErrWidgetNotFound) {
			writeError(w, http.StatusNotFound, "widget not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
