package handler

import (
	"errors"
	"net/http"

	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/usecase"
)

type AdminTenantHandler struct {
	adminTenants *usecase.AdminTenantService
}

func NewAdminTenantHandler(adminTenants *usecase.AdminTenantService) *AdminTenantHandler {
	return &AdminTenantHandler{adminTenants: adminTenants}
}

type createAdminTenantRequest struct {
	Name string `json:"name"`
}

type patchAdminTenantRequest struct {
	Name     *string `json:"name,omitempty"`
	IsActive *bool   `json:"is_active,omitempty"`
}

func (h *AdminTenantHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createAdminTenantRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	actorUserID, ok := domain.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	tenant, err := h.adminTenants.CreateTenant(r.Context(), actorUserID, req.Name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, toOpenAPITenant(*tenant))
}

func (h *AdminTenantHandler) List(w http.ResponseWriter, r *http.Request) {
	tenants, err := h.adminTenants.ListTenants(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if tenants == nil {
		tenants = []domain.Tenant{}
	}

	items := make([]tenantResponse, len(tenants))
	for i := range tenants {
		items[i] = toOpenAPITenant(tenants[i])
	}
	writeJSON(w, http.StatusOK, struct {
		Items []tenantResponse `json:"items"`
	}{Items: items})
}

func (h *AdminTenantHandler) Patch(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}

	var req patchAdminTenantRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == nil && req.IsActive == nil {
		writeError(w, http.StatusBadRequest, "at least one field is required")
		return
	}

	actorUserID, ok := domain.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	tenant, err := h.adminTenants.UpdateTenant(r.Context(), actorUserID, id, req.Name, req.IsActive)
	if err != nil {
		if errors.Is(err, domain.ErrTenantNotFound) {
			writeError(w, http.StatusNotFound, "tenant not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, toOpenAPITenant(*tenant))
}
