package handlers

import "net/http"

type HealthHandler struct {
	check func(r *http.Request) error
}

func NewHealthHandler(check func(r *http.Request) error) *HealthHandler {
	return &HealthHandler{check: check}
}

func (h *HealthHandler) Get(w http.ResponseWriter, r *http.Request) {
	if err := h.check(r); err != nil {
		writeError(w, http.StatusServiceUnavailable, err)
		return
	}

	writeJSON(w, http.StatusOK, healthDTO{Status: "ok"})
}
