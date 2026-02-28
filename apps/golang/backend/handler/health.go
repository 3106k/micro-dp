package handler

import (
	"database/sql"
	"net/http"
)

type HealthHandler struct {
	db *sql.DB
}

func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) Healthz(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{"status": "ok"}

	if h.db != nil {
		if err := h.db.PingContext(r.Context()); err != nil {
			resp["status"] = "degraded"
			resp["db"] = err.Error()
		} else {
			resp["db"] = "ok"
		}
	}

	writeJSON(w, http.StatusOK, resp)
}
