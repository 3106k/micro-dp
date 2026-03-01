package handler

import (
	"errors"
	"net/http"

	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/internal/openapi"
	"github.com/user/micro-dp/usecase"
)

type MemberHandler struct {
	members *usecase.MemberService
}

func NewMemberHandler(members *usecase.MemberService) *MemberHandler {
	return &MemberHandler{members: members}
}

func (h *MemberHandler) List(w http.ResponseWriter, r *http.Request) {
	members, err := h.members.ListMembers(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if members == nil {
		members = []domain.TenantMember{}
	}

	items := make([]openapi.TenantMember, len(members))
	for i := range members {
		items[i] = toOpenAPITenantMember(&members[i])
	}
	writeJSON(w, http.StatusOK, struct {
		Items []openapi.TenantMember `json:"items"`
	}{Items: items})
}

func (h *MemberHandler) CreateInvitation(w http.ResponseWriter, r *http.Request) {
	var req openapi.CreateInvitationRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Email == "" {
		writeError(w, http.StatusBadRequest, "email is required")
		return
	}
	if req.Role == "" {
		writeError(w, http.StatusBadRequest, "role is required")
		return
	}

	inv, err := h.members.CreateInvitation(r.Context(), string(req.Email), string(req.Role))
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInsufficientRole):
			writeError(w, http.StatusForbidden, "insufficient role")
		case errors.Is(err, domain.ErrAlreadyMember):
			writeError(w, http.StatusConflict, "user is already a member")
		case errors.Is(err, domain.ErrDuplicateInvitation):
			writeError(w, http.StatusConflict, "pending invitation already exists")
		default:
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusCreated, toOpenAPITenantInvitation(inv, true))
}

func (h *MemberHandler) AcceptInvitation(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	if token == "" {
		writeError(w, http.StatusBadRequest, "missing token")
		return
	}

	inv, err := h.members.AcceptInvitation(r.Context(), token)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvitationNotFound):
			writeError(w, http.StatusNotFound, "invitation not found")
		case errors.Is(err, domain.ErrInvitationExpired):
			writeError(w, http.StatusGone, "invitation expired")
		case errors.Is(err, domain.ErrInvitationAlreadyUsed):
			writeError(w, http.StatusConflict, "invitation already used")
		case errors.Is(err, domain.ErrAlreadyMember):
			writeError(w, http.StatusConflict, "already a member")
		default:
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusOK, toOpenAPITenantInvitation(inv, false))
}

func (h *MemberHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("user_id")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "missing user_id")
		return
	}

	var req openapi.UpdateMemberRoleRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Role == "" {
		writeError(w, http.StatusBadRequest, "role is required")
		return
	}

	member, err := h.members.UpdateMemberRole(r.Context(), userID, string(req.Role))
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInsufficientRole):
			writeError(w, http.StatusForbidden, "insufficient role")
		case errors.Is(err, domain.ErrCannotChangeOwnRole):
			writeError(w, http.StatusForbidden, "cannot change own role")
		case errors.Is(err, domain.ErrTenantNotFound):
			writeError(w, http.StatusNotFound, "member not found")
		default:
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusOK, toOpenAPITenantMember(member))
}

func (h *MemberHandler) Remove(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("user_id")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "missing user_id")
		return
	}

	err := h.members.RemoveMember(r.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInsufficientRole):
			writeError(w, http.StatusForbidden, "insufficient role")
		case errors.Is(err, domain.ErrCannotRemoveLastOwner):
			writeError(w, http.StatusConflict, "cannot remove the last owner")
		case errors.Is(err, domain.ErrTenantNotFound):
			writeError(w, http.StatusNotFound, "member not found")
		default:
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
