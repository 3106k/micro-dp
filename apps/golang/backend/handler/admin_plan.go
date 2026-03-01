package handler

import (
	"errors"
	"net/http"

	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/internal/openapi"
	"github.com/user/micro-dp/usecase"
)

type AdminPlanHandler struct {
	plans *usecase.PlanService
}

func NewAdminPlanHandler(plans *usecase.PlanService) *AdminPlanHandler {
	return &AdminPlanHandler{plans: plans}
}

func (h *AdminPlanHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req openapi.CreatePlanRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.DisplayName == "" {
		writeError(w, http.StatusBadRequest, "display_name is required")
		return
	}

	maxEvents := -1
	if req.MaxEventsPerDay != nil {
		maxEvents = *req.MaxEventsPerDay
	}
	maxRows := -1
	if req.MaxRowsPerDay != nil {
		maxRows = *req.MaxRowsPerDay
	}
	maxUploads := -1
	if req.MaxUploadsPerDay != nil {
		maxUploads = *req.MaxUploadsPerDay
	}
	var maxStorage int64 = -1
	if req.MaxStorageBytes != nil {
		maxStorage = *req.MaxStorageBytes
	}

	plan, err := h.plans.CreatePlan(r.Context(), req.Name, req.DisplayName, maxEvents, maxRows, maxUploads, maxStorage)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusCreated, toOpenAPIPlan(plan))
}

func (h *AdminPlanHandler) List(w http.ResponseWriter, r *http.Request) {
	plans, err := h.plans.ListPlans(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if plans == nil {
		plans = []domain.Plan{}
	}
	items := make([]openapi.Plan, len(plans))
	for i := range plans {
		items[i] = toOpenAPIPlan(&plans[i])
	}
	writeJSON(w, http.StatusOK, struct {
		Items []openapi.Plan `json:"items"`
	}{Items: items})
}

func (h *AdminPlanHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}

	var req openapi.UpdatePlanRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	plan, err := h.plans.UpdatePlan(r.Context(), id, req.DisplayName, req.MaxEventsPerDay, req.MaxRowsPerDay, req.MaxUploadsPerDay, req.MaxStorageBytes)
	if err != nil {
		if errors.Is(err, domain.ErrPlanNotFound) {
			writeError(w, http.StatusNotFound, "plan not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, toOpenAPIPlan(plan))
}

func (h *AdminPlanHandler) AssignPlan(w http.ResponseWriter, r *http.Request) {
	tenantID := r.PathValue("tenant_id")
	if tenantID == "" {
		writeError(w, http.StatusBadRequest, "missing tenant_id")
		return
	}

	var req openapi.AssignPlanRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.PlanId == "" {
		writeError(w, http.StatusBadRequest, "plan_id is required")
		return
	}

	plan, tp, err := h.plans.AssignPlan(r.Context(), tenantID, req.PlanId)
	if err != nil {
		if errors.Is(err, domain.ErrPlanNotFound) {
			writeError(w, http.StatusNotFound, "plan not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, toOpenAPITenantPlanResponse(plan, tp))
}
