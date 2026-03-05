package handler

import (
	"errors"
	"net/http"

	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/internal/openapi"
	"github.com/user/micro-dp/usecase"
)

type WriteKeyHandler struct {
	writeKeys *usecase.WriteKeyService
}

func NewWriteKeyHandler(writeKeys *usecase.WriteKeyService) *WriteKeyHandler {
	return &WriteKeyHandler{writeKeys: writeKeys}
}

func (h *WriteKeyHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req openapi.CreateWriteKeyRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	wk, rawKey, err := h.writeKeys.Generate(r.Context(), req.Name)
	if err != nil {
		if errors.Is(err, domain.ErrInsufficientRole) {
			writeError(w, http.StatusForbidden, "insufficient role")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, openapi.CreateWriteKeyResponse{
		WriteKey: toOpenAPIWriteKey(wk),
		RawKey:   rawKey,
	})
}

func (h *WriteKeyHandler) List(w http.ResponseWriter, r *http.Request) {
	keys, err := h.writeKeys.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	items := make([]openapi.WriteKey, len(keys))
	for i, k := range keys {
		items[i] = toOpenAPIWriteKey(&k)
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *WriteKeyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}

	err := h.writeKeys.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrWriteKeyNotFound) {
			writeError(w, http.StatusNotFound, "write key not found")
			return
		}
		if errors.Is(err, domain.ErrInsufficientRole) {
			writeError(w, http.StatusForbidden, "insufficient role")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *WriteKeyHandler) Regenerate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}

	wk, rawKey, err := h.writeKeys.Regenerate(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrWriteKeyNotFound) {
			writeError(w, http.StatusNotFound, "write key not found")
			return
		}
		if errors.Is(err, domain.ErrInsufficientRole) {
			writeError(w, http.StatusForbidden, "insufficient role")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, openapi.CreateWriteKeyResponse{
		WriteKey: toOpenAPIWriteKey(wk),
		RawKey:   rawKey,
	})
}
