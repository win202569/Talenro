package controlapi

import (
	"context"
	"net/http"
	"strings"

	"talenro.local/platform/internal/deviceauth"
	"talenro.local/platform/internal/identity"
	"talenro.local/platform/internal/securitykit"
)

type accountAuthorityKey struct{}
type deviceAuthorityKey struct{}

func (h *Handler) authenticateAccount(request *http.Request) (*http.Request, error) {
	if request == nil {
		return nil, dependencyUnavailableError()
	}
	token, err := exactBearerToken(request)
	if err != nil {
		return nil, err
	}
	raw, err := securitykit.DecodeOpaqueToken(token)
	if err != nil {
		return nil, authenticationFailedError()
	}
	defer raw.Clear()
	if h == nil || unavailableApplication(h.apps.AccountAuth) {
		return nil, dependencyUnavailableError()
	}
	authority, err := h.apps.AccountAuth.Authenticate(request.Context(), raw)
	if err != nil {
		return nil, err
	}
	ctx := context.WithValue(request.Context(), accountAuthorityKey{}, authority)
	return request.WithContext(ctx), nil
}

func (h *Handler) authenticateDevice(request *http.Request) (*http.Request, error) {
	if request == nil {
		return nil, dependencyUnavailableError()
	}
	token, err := exactBearerToken(request)
	if err != nil {
		return nil, err
	}
	raw, err := securitykit.DecodeOpaqueToken(token)
	if err != nil {
		return nil, authenticationFailedError()
	}
	defer raw.Clear()
	if h == nil || unavailableApplication(h.apps.DeviceAuth) {
		return nil, dependencyUnavailableError()
	}
	authority, err := h.apps.DeviceAuth.Authenticate(request.Context(), raw)
	if err != nil {
		return nil, err
	}
	ctx := context.WithValue(request.Context(), deviceAuthorityKey{}, authority)
	return request.WithContext(ctx), nil
}

func exactBearerToken(request *http.Request) (string, error) {
	values := request.Header.Values("Authorization")
	if len(values) != 1 || !strings.HasPrefix(values[0], "Bearer ") || strings.Count(values[0], " ") != 1 {
		return "", authenticationFailedError()
	}
	token := strings.TrimPrefix(values[0], "Bearer ")
	if token == "" || strings.ContainsAny(token, " \t\r\n") {
		return "", authenticationFailedError()
	}
	return token, nil
}

func accountAuthority(ctx context.Context) identity.AccountAuthority {
	authority, _ := ctx.Value(accountAuthorityKey{}).(identity.AccountAuthority)
	return authority
}

func deviceAuthority(ctx context.Context) deviceauth.BundleAuthority {
	authority, _ := ctx.Value(deviceAuthorityKey{}).(deviceauth.BundleAuthority)
	return authority
}
