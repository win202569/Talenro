package controlapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"talenro.local/platform/internal/apierrors"
)

func (h *Handler) writeError(w http.ResponseWriter, err error) {
	classified := apierrors.New(apierrors.DependencyUnavailable, apierrors.ContactSupport)
	var domainError apierrors.Error
	if errors.As(err, &domainError) {
		classified = domainError
	}
	public := classified.Public(h.traceID())
	status := http.StatusServiceUnavailable
	switch apierrors.Code(public.Code) {
	case apierrors.MalformedRequest:
		status = http.StatusBadRequest
	case apierrors.AuthenticationFailed:
		status = http.StatusUnauthorized
	case apierrors.ActionNotAllowed:
		status = http.StatusForbidden
	case apierrors.IdempotencyConflict, apierrors.StateConflict, apierrors.VersionRollback:
		status = http.StatusConflict
	case apierrors.RequestTooLarge:
		status = http.StatusRequestEntityTooLarge
	case apierrors.UnsupportedSchema:
		status = http.StatusUnsupportedMediaType
	case apierrors.RateLimited:
		status = http.StatusTooManyRequests
	case apierrors.DependencyUnavailable, apierrors.SigningUnavailable:
		status = http.StatusServiceUnavailable
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	if apierrors.Code(public.Code) == apierrors.AuthenticationFailed {
		w.Header().Set("WWW-Authenticate", "Bearer")
	}
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(public)
}

// GeneratedParameterError sanitizes all generated-wrapper parsing errors.
func (h *Handler) GeneratedParameterError(w http.ResponseWriter, _ *http.Request, _ error) {
	h.writeError(w, apierrors.New(apierrors.MalformedRequest, apierrors.ContactSupport))
}

func malformedRequestError() error {
	return apierrors.New(apierrors.MalformedRequest, apierrors.ContactSupport)
}

func authenticationFailedError() error {
	return apierrors.New(apierrors.AuthenticationFailed, apierrors.Reauthenticate)
}

func dependencyUnavailableError() error {
	return apierrors.New(apierrors.DependencyUnavailable, apierrors.Retry)
}
