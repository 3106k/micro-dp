package handler

import (
	"errors"
	"net/http"

	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/internal/openapi"
	"github.com/user/micro-dp/usecase"
)

type UploadHandler struct {
	uploads *usecase.UploadService
	plans   *usecase.PlanService
}

func NewUploadHandler(uploads *usecase.UploadService, plans *usecase.PlanService) *UploadHandler {
	return &UploadHandler{uploads: uploads, plans: plans}
}

func (h *UploadHandler) Presign(w http.ResponseWriter, r *http.Request) {
	if err := h.plans.CheckUploadQuota(r.Context()); err != nil {
		if errors.Is(err, domain.ErrQuotaExceeded) {
			writeError(w, http.StatusPaymentRequired, "upload quota exceeded")
			return
		}
	}

	var req openapi.CreateUploadPresignRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.Files) == 0 {
		writeError(w, http.StatusBadRequest, "at least one file is required")
		return
	}

	files := make([]usecase.UploadFileInput, len(req.Files))
	for i, f := range req.Files {
		files[i] = usecase.UploadFileInput{
			Filename:    f.Filename,
			ContentType: f.ContentType,
			SizeBytes:   f.SizeBytes,
		}
	}

	result, err := h.uploads.CreatePresign(r.Context(), files)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	respFiles := make([]openapi.UploadFilePresigned, len(result.Files))
	for i, f := range result.Files {
		respFiles[i] = openapi.UploadFilePresigned{
			FileId:       f.FileID,
			Filename:     f.Filename,
			PresignedUrl: f.PresignedURL,
			ObjectKey:    f.ObjectKey,
			ExpiresAt:    &f.ExpiresAt,
		}
	}

	writeJSON(w, http.StatusCreated, openapi.CreateUploadPresignResponse{
		UploadId: result.UploadID,
		Files:    respFiles,
	})
}

func (h *UploadHandler) Complete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing upload id")
		return
	}

	upload, files, err := h.uploads.Complete(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrUploadNotFound) {
			writeError(w, http.StatusNotFound, "upload not found")
			return
		}
		if errors.Is(err, domain.ErrUploadAlreadyComplete) {
			writeError(w, http.StatusConflict, "upload already complete")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, toOpenAPIUpload(upload, files))
}
