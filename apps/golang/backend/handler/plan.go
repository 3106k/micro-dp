package handler

import (
	"net/http"

	"github.com/user/micro-dp/usecase"
)

type PlanHandler struct {
	plans *usecase.PlanService
}

func NewPlanHandler(plans *usecase.PlanService) *PlanHandler {
	return &PlanHandler{plans: plans}
}

func (h *PlanHandler) GetPlan(w http.ResponseWriter, r *http.Request) {
	plan, tp, err := h.plans.GetTenantPlan(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, toOpenAPITenantPlanResponse(plan, tp))
}

func (h *PlanHandler) GetUsageSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.plans.GetUsageSummary(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, toOpenAPIUsageSummary(summary))
}
