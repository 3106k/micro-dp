package handler

import (
	"errors"
	"net/http"

	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/internal/openapi"
	"github.com/user/micro-dp/usecase"
)

type AuthHandler struct {
	auth *usecase.AuthService
}

func NewAuthHandler(auth *usecase.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req openapi.RegisterRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if string(req.Email) == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	displayName := ""
	if req.DisplayName != nil {
		displayName = *req.DisplayName
	}

	userID, tenantID, err := h.auth.Register(r.Context(), string(req.Email), req.Password, displayName)
	if err != nil {
		if errors.Is(err, domain.ErrEmailAlreadyExists) {
			writeError(w, http.StatusConflict, "email already registered")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, openapi.RegisterResponse{
		UserId:   userID,
		TenantId: tenantID,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req openapi.LoginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if string(req.Email) == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	token, err := h.auth.Login(r.Context(), string(req.Email), req.Password)
	if err != nil {
		if err.Error() == "invalid credentials" {
			writeError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, openapi.LoginResponse{Token: token})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := domain.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, tenants, err := h.auth.Me(r.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	oaTenants := make([]tenantResponse, len(tenants))
	for i, t := range tenants {
		oaTenants[i] = toOpenAPITenant(t)
	}

	writeJSON(w, http.StatusOK, struct {
		UserID       string           `json:"user_id"`
		Email        string           `json:"email"`
		DisplayName  string           `json:"display_name"`
		IsSuperadmin bool             `json:"is_superadmin"`
		Tenants      []tenantResponse `json:"tenants"`
	}{
		UserID:       user.ID,
		Email:        user.Email,
		DisplayName:  user.DisplayName,
		IsSuperadmin: user.IsSuperadmin,
		Tenants:      oaTenants,
	})
}
