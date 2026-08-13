package controlapi

import (
	"encoding/json"
	"net/http"
	"time"

	controlapiv1 "talenro.local/platform/gen/go/talenro/controlapi/v1"
	"talenro.local/platform/internal/deviceauth"
	"talenro.local/platform/internal/identity"
	"talenro.local/platform/internal/readiness"
	"talenro.local/platform/internal/securitykit"
	"talenro.local/platform/internal/trust"
)

// Applications contains the bounded domain surfaces consumed by the HTTP adapter.
type Applications struct {
	Identity        identity.Application
	StrongAuth      identity.StrongAuthApplication
	AccountAuth     identity.AccountAuthenticator
	Device          deviceauth.Application
	DeviceAuth      deviceauth.DeviceAuthenticator
	Trust           trust.Application
	ImmutableBundle http.Handler
}

// Handler implements the generated control API without owning domain policy.
type Handler struct {
	checker  *readiness.Checker
	apps     Applications
	deadline time.Duration
	random   securitykit.RandomSource
}

// NewHandler constructs bounded handlers backed by explicit applications.
func NewHandler(checker *readiness.Checker, apps Applications, deadline time.Duration, random securitykit.RandomSource) *Handler {
	return &Handler{checker: checker, apps: apps, deadline: deadline, random: random}
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
	if h == nil || h.checker == nil || r == nil {
		writeHealth(w, http.StatusServiceUnavailable, controlapiv1.HealthResponse{
			Status: controlapiv1.Unavailable,
			Checks: map[string]string{},
		})
		return
	}
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
