package controlapi

import (
	"encoding/json"
	"net/http"

	controlapiv1 "talenro.local/platform/gen/go/talenro/controlapi/v1"
	"talenro.local/platform/internal/readiness"
)

// Handler implements the generated liveness and readiness endpoints.
type Handler struct {
	checker *readiness.Checker
}

// NewHandler constructs health handlers backed by checker.
func NewHandler(checker *readiness.Checker) *Handler {
	return &Handler{checker: checker}
}

// GetLiveness reports whether the control API process is running.
func (h *Handler) GetLiveness(w http.ResponseWriter, _ *http.Request) {
	writeHealth(w, http.StatusOK, controlapiv1.HealthResponse{
		Status: controlapiv1.Ok,
		Checks: map[string]string{},
	})
}

// GetReadiness reports bounded dependency health without exposing error details.
func (h *Handler) GetReadiness(w http.ResponseWriter, r *http.Request) {
	ready, checks := h.checker.Check(r.Context())
	if ready {
		writeHealth(w, http.StatusOK, controlapiv1.HealthResponse{
			Status: controlapiv1.Ok,
			Checks: checks,
		})
		return
	}

	writeHealth(w, http.StatusServiceUnavailable, controlapiv1.HealthResponse{
		Status: controlapiv1.Unavailable,
		Checks: checks,
	})
}

func writeHealth(w http.ResponseWriter, status int, body controlapiv1.HealthResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
