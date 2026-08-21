//nolint:revive // Generated method names are the OpenAPI operation contract.
package controlapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	controlapiv1 "talenro.local/platform/gen/go/talenro/controlapi/v1"
	"talenro.local/platform/internal/apierrors"
	"talenro.local/platform/internal/deviceauth"
	"talenro.local/platform/internal/identity"
	"talenro.local/platform/internal/secret"
	"talenro.local/platform/internal/securitykit"
	"talenro.local/platform/internal/strictjson"
	"talenro.local/platform/internal/trust"
)

const maximumRequestBodyBytes int64 = 64 << 10

var _ controlapiv1.ServerInterface = (*Handler)(nil)

type emptyRequest struct{}

type canonicalUUID struct{ value uuid.UUID }

func (value *canonicalUUID) UnmarshalJSON(data []byte) error {
	var text string
	if err := json.Unmarshal(data, &text); err != nil {
		return malformedRequestError()
	}
	parsed, err := uuid.Parse(text)
	version := parsed.Version()
	if err != nil || parsed == uuid.Nil || parsed.String() != text || version < 1 || version > 8 || parsed.Variant() != uuid.RFC4122 {
		return malformedRequestError()
	}
	value.value = parsed
	return nil
}

func (value canonicalUUID) IsZero() bool    { return value.value == uuid.Nil }
func (value canonicalUUID) String() string  { return value.value.String() }
func (value canonicalUUID) UUID() uuid.UUID { return value.value }

type reauthenticationRequest struct {
	Method     string                                       `json:"method"`
	Password   string                                       `json:"password,omitempty"`
	Code       string                                       `json:"code,omitempty"`
	CeremonyID canonicalUUID                                `json:"ceremony_id,omitempty"`
	Response   *controlapiv1.WebAuthnAuthenticationResponse `json:"response,omitempty"`
}

type reauthenticationEnvelope struct {
	Reauthentication reauthenticationRequest `json:"reauthentication"`
}

type accountSessionRequest struct {
	Method                 string                                       `json:"method"`
	Email                  string                                       `json:"email,omitempty"`
	Password               string                                       `json:"password,omitempty"`
	CeremonyID             canonicalUUID                                `json:"ceremony_id,omitempty"`
	Response               *controlapiv1.WebAuthnAuthenticationResponse `json:"response,omitempty"`
	ClientSigningPublicKey string                                       `json:"client_signing_public_key"`
}

type deviceChallengeRequest struct {
	EnrollmentGrant  string `json:"enrollment_grant,omitempty"`
	RefreshToken     string `json:"refresh_token,omitempty"`
	RequestNonce     string `json:"request_nonce"`
	SigningPublicKey string `json:"signing_public_key,omitempty"`
	HPKEPublicKey    string `json:"hpke_public_key,omitempty"`
}

type revokeSessionsRequest struct {
	Scope            string                  `json:"scope"`
	SessionID        *canonicalUUID          `json:"session_id,omitempty"`
	Reauthentication reauthenticationRequest `json:"reauthentication"`
}

func (h *Handler) operationRequest(request *http.Request) (*http.Request, context.CancelFunc, error) {
	if h == nil || request == nil || h.deadline <= 0 {
		return nil, func() {}, dependencyUnavailableError()
	}
	ctx, cancel := context.WithTimeout(request.Context(), h.deadline)
	return request.WithContext(ctx), cancel, nil
}

func (h *Handler) decodeBody(w http.ResponseWriter, request *http.Request, target any) error {
	if request == nil || request.Body == nil {
		return malformedRequestError()
	}
	if value := request.Header.Get("Content-Type"); value != "" {
		mediaType, _, err := mime.ParseMediaType(value)
		if err != nil || mediaType != "application/json" {
			return apierrors.New(apierrors.UnsupportedSchema, apierrors.UpgradeClient)
		}
	}
	request.Body = http.MaxBytesReader(w, request.Body, maximumRequestBodyBytes)
	body, err := io.ReadAll(request.Body)
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			return apierrors.New(apierrors.RequestTooLarge, apierrors.ContactSupport)
		}
		return malformedRequestError()
	}
	if err := strictjson.Decode(bytes.NewReader(body), maximumRequestBodyBytes, target); err != nil {
		if errors.Is(err, strictjson.ErrTooLarge) {
			return apierrors.New(apierrors.RequestTooLarge, apierrors.ContactSupport)
		}
		return malformedRequestError()
	}
	return nil
}

func validIdempotencyKey(value string) bool {
	if len(value) < 22 || len(value) > 86 {
		return false
	}
	for index := range len(value) {
		character := value[index]
		if (character >= 'A' && character <= 'Z') || (character >= 'a' && character <= 'z') ||
			(character >= '0' && character <= '9') || character == '_' || character == '-' {
			continue
		}
		return false
	}
	return true
}

func validPassword(value string) bool {
	return utf8.ValidString(value) && len(value) >= 12 && len(value) <= 1024
}

func validEmail(value string) bool {
	_, err := identity.CanonicalizeEmail(value)
	return err == nil
}

func validLocale(value string) bool {
	if len(value) < 2 || len(value) > 35 {
		return false
	}
	parts := strings.Split(value, "-")
	if len(parts[0]) < 2 || len(parts[0]) > 3 || !allASCII(parts[0], true) {
		return false
	}
	for _, part := range parts[1:] {
		if len(part) < 2 || len(part) > 8 || !allASCII(part, false) {
			return false
		}
	}
	return true
}

func allASCII(value string, lettersOnly bool) bool {
	for index := range len(value) {
		character := value[index]
		letter := character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z'
		if !letter && (lettersOnly || character < '0' || character > '9') {
			return false
		}
	}
	return true
}

func validTOTPCode(value string) bool {
	return len(value) == 6 && allASCIIDigits(value)
}

func allASCIIDigits(value string) bool {
	for index := range len(value) {
		if value[index] < '0' || value[index] > '9' {
			return false
		}
	}
	return true
}

func validRecoveryCode(value string) bool {
	if len(value) != 39 {
		return false
	}
	for index := range len(value) {
		if index%5 == 4 {
			if value[index] != '-' {
				return false
			}
			continue
		}
		if (value[index] < 'A' || value[index] > 'Z') && (value[index] < '2' || value[index] > '7') {
			return false
		}
	}
	return true
}

func validWebAuthnEncoded(value string, minimumDecoded, maximumDecoded, maximumEncoded int) bool {
	if len(value) == 0 || len(value) > maximumEncoded {
		return false
	}
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(value)
	defer clear(decoded)
	return err == nil && len(decoded) >= minimumDecoded && len(decoded) <= maximumDecoded && base64.RawURLEncoding.EncodeToString(decoded) == value
}

func validWebAuthnCredentialID(value string) bool {
	return validWebAuthnEncoded(value, 16, 1024, 1366)
}

func validWebAuthnData(value string) bool {
	return validWebAuthnEncoded(value, 1, int(maximumRequestBodyBytes), int(maximumRequestBodyBytes))
}

func validWebAuthnAttachment(value *string) bool {
	return value == nil || *value == "platform" || *value == "cross-platform"
}

func validWebAuthnTransports(values *[]controlapiv1.WebAuthnTransport) bool {
	if values == nil {
		return true
	}
	if len(*values) > 8 {
		return false
	}
	seen := make(map[string]struct{}, len(*values))
	for _, value := range *values {
		transport := string(value)
		switch transport {
		case "usb", "nfc", "ble", "smart-card", "hybrid", "internal", "cable":
		default:
			return false
		}
		if _, duplicate := seen[transport]; duplicate {
			return false
		}
		seen[transport] = struct{}{}
	}
	return true
}

func validWebAuthnRegistrationResponse(value controlapiv1.WebAuthnRegistrationResponse) bool {
	attachment := (*string)(value.AuthenticatorAttachment)
	response := value.Response
	if !validWebAuthnCredentialID(value.Id) || !validWebAuthnCredentialID(value.RawId) || string(value.Type) != "public-key" ||
		!validWebAuthnAttachment(attachment) || !validWebAuthnData(response.ClientDataJSON) || !validWebAuthnData(response.AttestationObject) ||
		(response.AuthenticatorData != nil && !validWebAuthnData(*response.AuthenticatorData)) ||
		(response.PublicKey != nil && !validWebAuthnData(*response.PublicKey)) || !validWebAuthnTransports(response.Transports) {
		return false
	}
	return response.PublicKeyAlgorithm == nil || *response.PublicKeyAlgorithm >= -65536 && *response.PublicKeyAlgorithm <= 65536
}

func validWebAuthnAuthenticationResponse(value controlapiv1.WebAuthnAuthenticationResponse) bool {
	attachment := (*string)(value.AuthenticatorAttachment)
	response := value.Response
	if !validWebAuthnCredentialID(value.Id) || !validWebAuthnCredentialID(value.RawId) || string(value.Type) != "public-key" ||
		!validWebAuthnAttachment(attachment) || !validWebAuthnData(response.ClientDataJSON) ||
		!validWebAuthnData(response.AuthenticatorData) || !validWebAuthnData(response.Signature) {
		return false
	}
	return response.UserHandle == nil || validWebAuthnEncoded(*response.UserHandle, 1, 64, 86)
}

func decodeFixedBase64(value string, length int) ([]byte, error) {
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(value)
	if err != nil || len(decoded) != length || base64.RawURLEncoding.EncodeToString(decoded) != value {
		clear(decoded)
		return nil, malformedRequestError()
	}
	return decoded, nil
}

func decodeCredentialID(value string) ([]byte, error) {
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(value)
	if err != nil || len(decoded) < 16 || len(decoded) > 1024 || base64.RawURLEncoding.EncodeToString(decoded) != value {
		clear(decoded)
		return nil, malformedRequestError()
	}
	return decoded, nil
}

func decodeOpaque(value string) (secret.Bytes, error) {
	valueBytes, err := securitykit.DecodeOpaqueToken(value)
	if err != nil {
		return secret.Bytes{}, malformedRequestError()
	}
	return valueBytes, nil
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeNoContent(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusNoContent)
}

func unavailableApplication(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	kind := reflected.Kind()
	return (kind == reflect.Chan || kind == reflect.Func || kind == reflect.Interface || kind == reflect.Map || kind == reflect.Pointer || kind == reflect.Slice) && reflected.IsNil()
}

func (h *Handler) fail(w http.ResponseWriter, err error) { h.writeError(w, err) }

func reauthentication(authority identity.AccountAuthority, request reauthenticationRequest) (identity.Reauthentication, error) {
	result := identity.Reauthentication{SessionID: authority.SessionID}
	switch request.Method {
	case "password":
		if !validPassword(request.Password) || request.Code != "" || !request.CeremonyID.IsZero() || request.Response != nil {
			return identity.Reauthentication{}, malformedRequestError()
		}
		result.Method = identity.ReauthPassword
		result.Proof = secret.NewBytes([]byte(request.Password))
	case "totp":
		if !validTOTPCode(request.Code) || request.Password != "" || !request.CeremonyID.IsZero() || request.Response != nil {
			return identity.Reauthentication{}, malformedRequestError()
		}
		result.Method = identity.ReauthTOTP
		result.Proof = secret.NewBytes([]byte(request.Code))
	case "recovery_code":
		if !validRecoveryCode(request.Code) || request.Password != "" || !request.CeremonyID.IsZero() || request.Response != nil {
			return identity.Reauthentication{}, malformedRequestError()
		}
		result.Method = identity.ReauthRecoveryCode
		result.Proof = secret.NewBytes([]byte(request.Code))
	case "passkey":
		if request.CeremonyID.IsZero() || request.Response == nil || !validWebAuthnAuthenticationResponse(*request.Response) || request.Password != "" || request.Code != "" {
			return identity.Reauthentication{}, malformedRequestError()
		}
		proof, err := json.Marshal(request.Response)
		if err != nil {
			return identity.Reauthentication{}, malformedRequestError()
		}
		result.Method = identity.ReauthPasskey
		result.Proof = secret.NewBytes(proof)
		clear(proof)
	default:
		return identity.Reauthentication{}, malformedRequestError()
	}
	return result, nil
}

func sessionTokensBody(ctx context.Context, authenticator identity.AccountAuthenticator, tokens identity.SessionTokens) (controlapiv1.AccountTokens, error) {
	if unavailableApplication(authenticator) {
		return controlapiv1.AccountTokens{}, dependencyUnavailableError()
	}
	authority, err := authenticator.Authenticate(ctx, tokens.AccessToken)
	if err != nil {
		return controlapiv1.AccountTokens{}, err
	}
	principalID, err := uuid.Parse(string(authority.PrincipalID))
	if err != nil {
		return controlapiv1.AccountTokens{}, dependencyUnavailableError()
	}
	sessionID, err := uuid.Parse(string(authority.SessionID))
	if err != nil {
		return controlapiv1.AccountTokens{}, dependencyUnavailableError()
	}
	access := securitykit.EncodeOpaqueToken(tokens.AccessToken)
	refresh := securitykit.EncodeOpaqueToken(tokens.RefreshToken)
	if access == "" || refresh == "" {
		return controlapiv1.AccountTokens{}, dependencyUnavailableError()
	}
	return controlapiv1.AccountTokens{
		PrincipalId: principalID, SessionId: sessionID, AccessToken: access, RefreshToken: refresh, ExpiresAt: tokens.AccessExpiresAt,
	}, nil
}

func deviceTokensBody(tokens deviceauth.DeviceTokens) (controlapiv1.DeviceTokens, error) {
	access := securitykit.EncodeOpaqueToken(tokens.AccessToken)
	refresh := securitykit.EncodeOpaqueToken(tokens.RefreshToken)
	if access == "" || refresh == "" || tokens.DeviceID == uuid.Nil || tokens.AuthorizationID == uuid.Nil || tokens.FamilyID == uuid.Nil {
		return controlapiv1.DeviceTokens{}, dependencyUnavailableError()
	}
	return controlapiv1.DeviceTokens{
		DeviceId: tokens.DeviceID, AuthorizationId: tokens.AuthorizationID, FamilyId: tokens.FamilyID, AccessToken: access, RefreshToken: refresh, ExpiresAt: tokens.AccessExpiresAt,
	}, nil
}

func challengeBody(challengeID string, challenge [32]byte, expiresAt time.Time) (controlapiv1.AuthChallenge, error) {
	parsedID, err := uuid.Parse(challengeID)
	if err != nil || challenge == [32]byte{} || expiresAt.IsZero() {
		return controlapiv1.AuthChallenge{}, dependencyUnavailableError()
	}
	return controlapiv1.AuthChallenge{ChallengeId: parsedID, Challenge: base64.RawURLEncoding.EncodeToString(challenge[:]), ExpiresAt: expiresAt}, nil
}

// CreateAccount maps the public account-registration route.
func (h *Handler) CreateAccount(w http.ResponseWriter, request *http.Request, params controlapiv1.CreateAccountParams) {
	request, cancel, err := h.operationRequest(request)
	defer cancel()
	if err != nil {
		h.fail(w, dependencyUnavailableError())
		return
	}
	if !validIdempotencyKey(params.IdempotencyKey) {
		h.fail(w, malformedRequestError())
		return
	}
	var body controlapiv1.CreateAccountRequest
	if err = h.decodeBody(w, request, &body); err != nil || !validEmail(string(body.Email)) || !validLocale(body.Locale) || !validPassword(body.Password) {
		if err == nil {
			err = malformedRequestError()
		}
		h.fail(w, err)
		return
	}
	if unavailableApplication(h.apps.Identity) {
		h.fail(w, dependencyUnavailableError())
		return
	}
	password := secret.NewBytes([]byte(body.Password))
	defer password.Clear()
	_, err = h.apps.Identity.RegisterAccount(request.Context(), identity.RegisterAccountCommand{
		Email: string(body.Email), Password: password, Locale: body.Locale, IdempotencyKey: params.IdempotencyKey,
	})
	if err != nil {
		h.fail(w, err)
		return
	}
	h.writeJSON(w, http.StatusAccepted, controlapiv1.GenericAccepted{Status: controlapiv1.Accepted})
}

func (h *Handler) CreateEmailVerificationDelivery(w http.ResponseWriter, request *http.Request, params controlapiv1.CreateEmailVerificationDeliveryParams) {
	request, cancel, err := h.operationRequest(request)
	defer cancel()
	if err != nil {
		h.fail(w, dependencyUnavailableError())
		return
	}
	if !validIdempotencyKey(params.IdempotencyKey) {
		h.fail(w, malformedRequestError())
		return
	}
	var body controlapiv1.CreateEmailVerificationDeliveryRequest
	if err = h.decodeBody(w, request, &body); err != nil || !validEmail(string(body.Email)) || !validLocale(body.Locale) {
		if err == nil {
			err = malformedRequestError()
		}
		h.fail(w, err)
		return
	}
	if unavailableApplication(h.apps.Identity) {
		h.fail(w, dependencyUnavailableError())
		return
	}
	_, err = h.apps.Identity.CreateEmailVerificationDelivery(request.Context(), identity.CreateEmailVerificationDeliveryCommand{
		Email: string(body.Email), Locale: body.Locale, IdempotencyKey: params.IdempotencyKey,
	})
	if err != nil {
		h.fail(w, err)
		return
	}
	h.writeJSON(w, http.StatusAccepted, controlapiv1.GenericAccepted{Status: controlapiv1.Accepted})
}

func (h *Handler) VerifyEmail(w http.ResponseWriter, request *http.Request, params controlapiv1.VerifyEmailParams) {
	request, cancel, err := h.operationRequest(request)
	defer cancel()
	if err != nil {
		h.fail(w, dependencyUnavailableError())
		return
	}
	if !validIdempotencyKey(params.IdempotencyKey) {
		h.fail(w, malformedRequestError())
		return
	}
	var body controlapiv1.VerifyEmailRequest
	if err = h.decodeBody(w, request, &body); err != nil {
		h.fail(w, err)
		return
	}
	token, err := decodeOpaque(body.Token)
	if err != nil {
		h.fail(w, err)
		return
	}
	defer token.Clear()
	if unavailableApplication(h.apps.Identity) {
		h.fail(w, dependencyUnavailableError())
		return
	}
	if err = h.apps.Identity.VerifyEmail(request.Context(), identity.VerifyEmailCommand{Token: token, IdempotencyKey: params.IdempotencyKey}); err != nil {
		h.fail(w, err)
		return
	}
	writeNoContent(w)
}

func (h *Handler) CreatePasswordResetDelivery(w http.ResponseWriter, request *http.Request, params controlapiv1.CreatePasswordResetDeliveryParams) {
	request, cancel, err := h.operationRequest(request)
	defer cancel()
	if err != nil {
		h.fail(w, dependencyUnavailableError())
		return
	}
	if !validIdempotencyKey(params.IdempotencyKey) {
		h.fail(w, malformedRequestError())
		return
	}
	var body controlapiv1.CreatePasswordResetDeliveryRequest
	if err = h.decodeBody(w, request, &body); err != nil || !validEmail(string(body.Email)) || !validLocale(body.Locale) {
		if err == nil {
			err = malformedRequestError()
		}
		h.fail(w, err)
		return
	}
	if unavailableApplication(h.apps.Identity) {
		h.fail(w, dependencyUnavailableError())
		return
	}
	_, err = h.apps.Identity.CreatePasswordResetDelivery(request.Context(), identity.CreatePasswordResetDeliveryCommand{
		Email: string(body.Email), Locale: body.Locale, IdempotencyKey: params.IdempotencyKey,
	})
	if err != nil {
		h.fail(w, err)
		return
	}
	h.writeJSON(w, http.StatusAccepted, controlapiv1.GenericAccepted{Status: controlapiv1.Accepted})
}

func (h *Handler) ResetPassword(w http.ResponseWriter, request *http.Request, params controlapiv1.ResetPasswordParams) {
	request, cancel, err := h.operationRequest(request)
	defer cancel()
	if err != nil {
		h.fail(w, dependencyUnavailableError())
		return
	}
	if !validIdempotencyKey(params.IdempotencyKey) {
		h.fail(w, malformedRequestError())
		return
	}
	var body controlapiv1.ResetPasswordRequest
	if err = h.decodeBody(w, request, &body); err != nil || !validEmail(string(body.Email)) || !validPassword(body.NewPassword) {
		if err == nil {
			err = malformedRequestError()
		}
		h.fail(w, err)
		return
	}
	resetToken, err := decodeOpaque(body.Token)
	if err != nil {
		h.fail(w, err)
		return
	}
	defer resetToken.Clear()
	key, err := decodeFixedBase64(body.ClientSigningPublicKey, 32)
	if err != nil {
		h.fail(w, err)
		return
	}
	defer clear(key)
	password := secret.NewBytes([]byte(body.NewPassword))
	defer password.Clear()
	var signingKey [32]byte
	copy(signingKey[:], key)
	if unavailableApplication(h.apps.Identity) {
		h.fail(w, dependencyUnavailableError())
		return
	}
	tokens, err := h.apps.Identity.ResetPassword(request.Context(), identity.ResetPasswordCommand{
		Email: string(body.Email), Token: resetToken, NewPassword: password, ClientSigningPublicKey: signingKey, IdempotencyKey: params.IdempotencyKey,
	})
	if err != nil {
		h.fail(w, err)
		return
	}
	defer tokens.AccessToken.Clear()
	defer tokens.RefreshToken.Clear()
	response, err := sessionTokensBody(request.Context(), h.apps.AccountAuth, tokens)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, response)
}

func (h *Handler) CreateAccountSession(w http.ResponseWriter, request *http.Request, params controlapiv1.CreateAccountSessionParams) {
	request, cancel, err := h.operationRequest(request)
	defer cancel()
	if err != nil || !validIdempotencyKey(params.IdempotencyKey) {
		if err == nil {
			err = malformedRequestError()
		}
		h.fail(w, err)
		return
	}
	var body accountSessionRequest
	if err = h.decodeBody(w, request, &body); err != nil {
		h.fail(w, err)
		return
	}
	key, err := decodeFixedBase64(body.ClientSigningPublicKey, 32)
	if err != nil {
		h.fail(w, err)
		return
	}
	defer clear(key)
	var signingKey [32]byte
	copy(signingKey[:], key)
	var tokens identity.SessionTokens
	switch body.Method {
	case "password":
		if !validEmail(body.Email) || !validPassword(body.Password) || !body.CeremonyID.IsZero() || body.Response != nil {
			h.fail(w, malformedRequestError())
			return
		}
		if unavailableApplication(h.apps.Identity) {
			h.fail(w, dependencyUnavailableError())
			return
		}
		password := secret.NewBytes([]byte(body.Password))
		defer password.Clear()
		tokens, err = h.apps.Identity.CreateSession(request.Context(), identity.CreateSessionCommand{
			Method: identity.SessionPassword, Email: body.Email, Password: password, ClientSigningPublicKey: signingKey, IdempotencyKey: params.IdempotencyKey,
		})
	case "passkey":
		if body.CeremonyID.IsZero() || body.Response == nil || !validWebAuthnAuthenticationResponse(*body.Response) || body.Email != "" || body.Password != "" {
			h.fail(w, malformedRequestError())
			return
		}
		if unavailableApplication(h.apps.StrongAuth) {
			h.fail(w, dependencyUnavailableError())
			return
		}
		response, marshalErr := json.Marshal(body.Response)
		if marshalErr != nil {
			h.fail(w, malformedRequestError())
			return
		}
		defer clear(response)
		tokens, err = h.apps.StrongAuth.FinishPasskeyAuthentication(request.Context(), identity.FinishPasskeyAuthenticationCommand{
			CeremonyID: body.CeremonyID.String(), Response: response, ClientSigningPublicKey: signingKey, IdempotencyKey: params.IdempotencyKey,
		})
	default:
		h.fail(w, malformedRequestError())
		return
	}
	if err != nil {
		h.fail(w, err)
		return
	}
	defer tokens.AccessToken.Clear()
	defer tokens.RefreshToken.Clear()
	response, err := sessionTokensBody(request.Context(), h.apps.AccountAuth, tokens)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, response)
}

func (h *Handler) CreateAccountAuthChallenge(w http.ResponseWriter, request *http.Request, params controlapiv1.CreateAccountAuthChallengeParams) {
	request, cancel, err := h.operationRequest(request)
	defer cancel()
	if err != nil {
		h.fail(w, dependencyUnavailableError())
		return
	}
	if !validIdempotencyKey(params.IdempotencyKey) {
		h.fail(w, malformedRequestError())
		return
	}
	var body controlapiv1.CreateAccountAuthChallengeRequest
	if err = h.decodeBody(w, request, &body); err != nil {
		h.fail(w, err)
		return
	}
	refresh, err := decodeOpaque(body.RefreshToken)
	if err != nil {
		h.fail(w, err)
		return
	}
	defer refresh.Clear()
	nonce, err := decodeFixedBase64(body.RequestNonce, 32)
	if err != nil {
		h.fail(w, err)
		return
	}
	defer clear(nonce)
	var fixedNonce [32]byte
	copy(fixedNonce[:], nonce)
	if unavailableApplication(h.apps.Identity) {
		h.fail(w, dependencyUnavailableError())
		return
	}
	challenge, err := h.apps.Identity.CreateSessionChallenge(request.Context(), identity.CreateSessionChallengeCommand{
		RefreshToken: refresh, RequestNonce: fixedNonce, IdempotencyKey: params.IdempotencyKey,
	})
	if err != nil {
		h.fail(w, err)
		return
	}
	response, err := challengeBody(challenge.ChallengeID, challenge.Challenge, challenge.ExpiresAt)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.writeJSON(w, http.StatusCreated, response)
}

func (h *Handler) RotateAccountToken(w http.ResponseWriter, request *http.Request, params controlapiv1.RotateAccountTokenParams) {
	request, cancel, err := h.operationRequest(request)
	defer cancel()
	if err != nil {
		h.fail(w, dependencyUnavailableError())
		return
	}
	if !validIdempotencyKey(params.IdempotencyKey) {
		h.fail(w, malformedRequestError())
		return
	}
	var body struct {
		ChallengeID  canonicalUUID `json:"challenge_id"`
		RefreshToken string        `json:"refresh_token"`
		RequestNonce string        `json:"request_nonce"`
		Signature    string        `json:"signature"`
	}
	if err = h.decodeBody(w, request, &body); err != nil || body.ChallengeID.IsZero() {
		if err == nil {
			err = malformedRequestError()
		}
		h.fail(w, err)
		return
	}
	refresh, err := decodeOpaque(body.RefreshToken)
	if err != nil {
		h.fail(w, err)
		return
	}
	defer refresh.Clear()
	nonce, err := decodeFixedBase64(body.RequestNonce, 32)
	if err != nil {
		h.fail(w, err)
		return
	}
	defer clear(nonce)
	signature, err := decodeFixedBase64(body.Signature, 64)
	if err != nil {
		h.fail(w, err)
		return
	}
	defer clear(signature)
	var fixedNonce [32]byte
	var fixedSignature [64]byte
	copy(fixedNonce[:], nonce)
	copy(fixedSignature[:], signature)
	if unavailableApplication(h.apps.Identity) {
		h.fail(w, dependencyUnavailableError())
		return
	}
	tokens, err := h.apps.Identity.RotateSession(request.Context(), identity.RotateSessionCommand{
		RefreshToken: refresh, ChallengeID: body.ChallengeID.String(), RequestNonce: fixedNonce, Signature: fixedSignature, IdempotencyKey: params.IdempotencyKey,
	})
	if err != nil {
		h.fail(w, err)
		return
	}
	defer tokens.AccessToken.Clear()
	defer tokens.RefreshToken.Clear()
	response, err := sessionTokensBody(request.Context(), h.apps.AccountAuth, tokens)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, response)
}

func (h *Handler) authenticatedAccountRequest(request *http.Request) (*http.Request, context.CancelFunc, identity.AccountAuthority, error) {
	request, cancel, err := h.operationRequest(request)
	if err != nil {
		return nil, cancel, identity.AccountAuthority{}, err
	}
	request, err = h.authenticateAccount(request)
	if err != nil {
		cancel()
		return nil, func() {}, identity.AccountAuthority{}, err
	}
	return request, cancel, accountAuthority(request.Context()), nil
}

func (h *Handler) ChangePassword(w http.ResponseWriter, request *http.Request, params controlapiv1.ChangePasswordParams) {
	request, cancel, authority, err := h.authenticatedAccountRequest(request)
	defer cancel()
	if err != nil {
		h.fail(w, err)
		return
	}
	if !validIdempotencyKey(params.IdempotencyKey) {
		h.fail(w, malformedRequestError())
		return
	}
	type changePasswordRequest struct {
		CurrentPassword        string                  `json:"current_password"`
		NewPassword            string                  `json:"new_password"`
		Reauthentication       reauthenticationRequest `json:"reauthentication"`
		ClientSigningPublicKey string                  `json:"client_signing_public_key"`
	}
	var body changePasswordRequest
	if err = h.decodeBody(w, request, &body); err != nil || !validPassword(body.CurrentPassword) || !validPassword(body.NewPassword) {
		if err == nil {
			err = malformedRequestError()
		}
		h.fail(w, err)
		return
	}
	reauth, err := reauthentication(authority, body.Reauthentication)
	if err != nil {
		h.fail(w, err)
		return
	}
	defer reauth.Proof.Clear()
	key, err := decodeFixedBase64(body.ClientSigningPublicKey, 32)
	if err != nil {
		h.fail(w, err)
		return
	}
	defer clear(key)
	currentPassword := secret.NewBytes([]byte(body.CurrentPassword))
	newPassword := secret.NewBytes([]byte(body.NewPassword))
	defer currentPassword.Clear()
	defer newPassword.Clear()
	var signingKey [32]byte
	copy(signingKey[:], key)
	if unavailableApplication(h.apps.Identity) {
		h.fail(w, dependencyUnavailableError())
		return
	}
	tokens, err := h.apps.Identity.ChangePassword(request.Context(), identity.ChangePasswordCommand{
		PrincipalID: authority.PrincipalID, CurrentPassword: currentPassword, NewPassword: newPassword,
		Reauthentication: reauth, ClientSigningPublicKey: signingKey, IdempotencyKey: params.IdempotencyKey,
	})
	if err != nil {
		h.fail(w, err)
		return
	}
	defer tokens.AccessToken.Clear()
	defer tokens.RefreshToken.Clear()
	response, err := sessionTokensBody(request.Context(), h.apps.AccountAuth, tokens)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, response)
}

func (h *Handler) RevokeAccountSessions(w http.ResponseWriter, request *http.Request, params controlapiv1.RevokeAccountSessionsParams) {
	request, cancel, authority, err := h.authenticatedAccountRequest(request)
	defer cancel()
	if err != nil {
		h.fail(w, err)
		return
	}
	if !validIdempotencyKey(params.IdempotencyKey) {
		h.fail(w, malformedRequestError())
		return
	}
	var body revokeSessionsRequest
	if err = h.decodeBody(w, request, &body); err != nil {
		h.fail(w, err)
		return
	}
	reauth, err := reauthentication(authority, body.Reauthentication)
	if err != nil {
		h.fail(w, err)
		return
	}
	defer reauth.Proof.Clear()
	command := identity.RevokeSessionsCommand{
		PrincipalID: authority.PrincipalID, Reauthentication: reauth, IdempotencyKey: params.IdempotencyKey,
	}
	switch body.Scope {
	case "all":
		if body.SessionID != nil {
			h.fail(w, malformedRequestError())
			return
		}
		command.Scope = identity.RevokeAllSessions
	case "others":
		if body.SessionID != nil {
			h.fail(w, malformedRequestError())
			return
		}
		command.Scope = identity.RevokeOtherSessions
		command.SessionID = authority.SessionID
	case "one":
		if body.SessionID == nil || body.SessionID.IsZero() {
			h.fail(w, malformedRequestError())
			return
		}
		command.SessionID = identity.SessionID(body.SessionID.String())
		command.Scope = identity.RevokeCurrentSession
	default:
		h.fail(w, malformedRequestError())
		return
	}
	if unavailableApplication(h.apps.Identity) {
		h.fail(w, dependencyUnavailableError())
		return
	}
	if err = h.apps.Identity.RevokeSessions(request.Context(), command); err != nil {
		h.fail(w, err)
		return
	}
	writeNoContent(w)
}

func (h *Handler) CreatePasskeyRegistrationOptions(w http.ResponseWriter, request *http.Request, params controlapiv1.CreatePasskeyRegistrationOptionsParams) {
	request, cancel, authority, err := h.authenticatedAccountRequest(request)
	defer cancel()
	if err != nil {
		h.fail(w, err)
		return
	}
	if !validIdempotencyKey(params.IdempotencyKey) {
		h.fail(w, malformedRequestError())
		return
	}
	var body reauthenticationEnvelope
	if err = h.decodeBody(w, request, &body); err != nil {
		h.fail(w, err)
		return
	}
	reauth, err := reauthentication(authority, body.Reauthentication)
	if err != nil {
		h.fail(w, err)
		return
	}
	defer reauth.Proof.Clear()
	if unavailableApplication(h.apps.StrongAuth) {
		h.fail(w, dependencyUnavailableError())
		return
	}
	options, err := h.apps.StrongAuth.BeginPasskeyRegistration(request.Context(), identity.BeginPasskeyRegistrationCommand{
		PrincipalID: authority.PrincipalID, DisplayName: "account", Reauthentication: reauth, IdempotencyKey: params.IdempotencyKey,
	})
	if err != nil {
		h.fail(w, err)
		return
	}
	h.writeRawJSON(w, http.StatusOK, options)
}

func (h *Handler) writeRawJSON(w http.ResponseWriter, status int, body json.RawMessage) {
	if !json.Valid(body) || len(body) == 0 || len(body) > int(maximumRequestBodyBytes) {
		h.fail(w, dependencyUnavailableError())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

func (h *Handler) CreatePasskeyCredential(w http.ResponseWriter, request *http.Request, params controlapiv1.CreatePasskeyCredentialParams) {
	request, cancel, authority, err := h.authenticatedAccountRequest(request)
	defer cancel()
	if err != nil {
		h.fail(w, err)
		return
	}
	if !validIdempotencyKey(params.IdempotencyKey) {
		h.fail(w, malformedRequestError())
		return
	}
	type credentialRequest struct {
		CeremonyID       canonicalUUID                             `json:"ceremony_id"`
		Response         controlapiv1.WebAuthnRegistrationResponse `json:"response"`
		Reauthentication reauthenticationRequest                   `json:"reauthentication"`
	}
	var body credentialRequest
	if err = h.decodeBody(w, request, &body); err != nil || body.CeremonyID.IsZero() || !validWebAuthnRegistrationResponse(body.Response) {
		if err == nil {
			err = malformedRequestError()
		}
		h.fail(w, err)
		return
	}
	reauth, err := reauthentication(authority, body.Reauthentication)
	if err != nil {
		h.fail(w, err)
		return
	}
	defer reauth.Proof.Clear()
	response, err := json.Marshal(body.Response)
	if err != nil {
		h.fail(w, malformedRequestError())
		return
	}
	defer clear(response)
	if unavailableApplication(h.apps.StrongAuth) {
		h.fail(w, dependencyUnavailableError())
		return
	}
	err = h.apps.StrongAuth.FinishPasskeyRegistration(request.Context(), identity.FinishPasskeyRegistrationCommand{
		PrincipalID: authority.PrincipalID, CeremonyID: body.CeremonyID.String(), Response: response,
		Reauthentication: reauth, IdempotencyKey: params.IdempotencyKey,
	})
	if err != nil {
		h.fail(w, err)
		return
	}
	writeNoContent(w)
}

func (h *Handler) CreatePasskeyAuthenticationOptions(w http.ResponseWriter, request *http.Request, params controlapiv1.CreatePasskeyAuthenticationOptionsParams) {
	request, cancel, err := h.operationRequest(request)
	defer cancel()
	if err != nil {
		h.fail(w, dependencyUnavailableError())
		return
	}
	if !validIdempotencyKey(params.IdempotencyKey) {
		h.fail(w, malformedRequestError())
		return
	}
	var body emptyRequest
	if err = h.decodeBody(w, request, &body); err != nil {
		h.fail(w, err)
		return
	}
	if unavailableApplication(h.apps.StrongAuth) {
		h.fail(w, dependencyUnavailableError())
		return
	}
	options, err := h.apps.StrongAuth.BeginPasskeyAuthentication(request.Context(), identity.BeginPasskeyAuthenticationCommand{IdempotencyKey: params.IdempotencyKey})
	if err != nil {
		h.fail(w, err)
		return
	}
	h.writeRawJSON(w, http.StatusOK, options)
}

func (h *Handler) RevokePasskey(w http.ResponseWriter, request *http.Request, params controlapiv1.RevokePasskeyParams) {
	request, cancel, authority, err := h.authenticatedAccountRequest(request)
	defer cancel()
	if err != nil {
		h.fail(w, err)
		return
	}
	if !validIdempotencyKey(params.IdempotencyKey) {
		h.fail(w, malformedRequestError())
		return
	}
	type revokeRequest struct {
		CredentialID     string                  `json:"credential_id"`
		Reauthentication reauthenticationRequest `json:"reauthentication"`
	}
	var body revokeRequest
	if err = h.decodeBody(w, request, &body); err != nil {
		h.fail(w, err)
		return
	}
	credentialID, err := decodeCredentialID(body.CredentialID)
	if err != nil {
		h.fail(w, err)
		return
	}
	defer clear(credentialID)
	reauth, err := reauthentication(authority, body.Reauthentication)
	if err != nil {
		h.fail(w, err)
		return
	}
	defer reauth.Proof.Clear()
	if unavailableApplication(h.apps.StrongAuth) {
		h.fail(w, dependencyUnavailableError())
		return
	}
	err = h.apps.StrongAuth.RevokePasskey(request.Context(), identity.RevokePasskeyCommand{
		PrincipalID: authority.PrincipalID, CredentialID: credentialID, Reauthentication: reauth, IdempotencyKey: params.IdempotencyKey,
	})
	if err != nil {
		h.fail(w, err)
		return
	}
	writeNoContent(w)
}

func (h *Handler) CreateTOTPEnrollment(w http.ResponseWriter, request *http.Request, params controlapiv1.CreateTOTPEnrollmentParams) {
	request, cancel, authority, err := h.authenticatedAccountRequest(request)
	defer cancel()
	if err != nil {
		h.fail(w, err)
		return
	}
	if !validIdempotencyKey(params.IdempotencyKey) {
		h.fail(w, malformedRequestError())
		return
	}
	var body reauthenticationEnvelope
	if err = h.decodeBody(w, request, &body); err != nil {
		h.fail(w, err)
		return
	}
	reauth, err := reauthentication(authority, body.Reauthentication)
	if err != nil {
		h.fail(w, err)
		return
	}
	defer reauth.Proof.Clear()
	if unavailableApplication(h.apps.StrongAuth) {
		h.fail(w, dependencyUnavailableError())
		return
	}
	enrollment, err := h.apps.StrongAuth.BeginTOTPEnrollment(request.Context(), identity.BeginTOTPEnrollmentCommand{
		PrincipalID: authority.PrincipalID, Reauthentication: reauth, IdempotencyKey: params.IdempotencyKey,
	})
	if err != nil {
		h.fail(w, err)
		return
	}
	h.writeJSON(w, http.StatusCreated, controlapiv1.TOTPEnrollment{Secret: enrollment.Secret, ProvisioningUri: enrollment.URI})
}

func (h *Handler) VerifyTOTPEnrollment(w http.ResponseWriter, request *http.Request, params controlapiv1.VerifyTOTPEnrollmentParams) {
	request, cancel, authority, err := h.authenticatedAccountRequest(request)
	defer cancel()
	if err != nil {
		h.fail(w, err)
		return
	}
	if !validIdempotencyKey(params.IdempotencyKey) {
		h.fail(w, malformedRequestError())
		return
	}
	type verifyRequest struct {
		Code             string                  `json:"code"`
		Reauthentication reauthenticationRequest `json:"reauthentication"`
	}
	var body verifyRequest
	if err = h.decodeBody(w, request, &body); err != nil || !validTOTPCode(body.Code) {
		if err == nil {
			err = malformedRequestError()
		}
		h.fail(w, err)
		return
	}
	reauth, err := reauthentication(authority, body.Reauthentication)
	if err != nil {
		h.fail(w, err)
		return
	}
	defer reauth.Proof.Clear()
	code := secret.NewBytes([]byte(body.Code))
	defer code.Clear()
	if unavailableApplication(h.apps.StrongAuth) {
		h.fail(w, dependencyUnavailableError())
		return
	}
	err = h.apps.StrongAuth.VerifyTOTPEnrollment(request.Context(), identity.VerifyTOTPEnrollmentCommand{
		PrincipalID: authority.PrincipalID, Code: code, Reauthentication: reauth, IdempotencyKey: params.IdempotencyKey,
	})
	if err != nil {
		h.fail(w, err)
		return
	}
	writeNoContent(w)
}

func (h *Handler) RevokeTOTP(w http.ResponseWriter, request *http.Request, params controlapiv1.RevokeTOTPParams) {
	request, cancel, authority, err := h.authenticatedAccountRequest(request)
	defer cancel()
	if err != nil {
		h.fail(w, err)
		return
	}
	if !validIdempotencyKey(params.IdempotencyKey) {
		h.fail(w, malformedRequestError())
		return
	}
	var body reauthenticationEnvelope
	if err = h.decodeBody(w, request, &body); err != nil {
		h.fail(w, err)
		return
	}
	reauth, err := reauthentication(authority, body.Reauthentication)
	if err != nil {
		h.fail(w, err)
		return
	}
	defer reauth.Proof.Clear()
	if unavailableApplication(h.apps.StrongAuth) {
		h.fail(w, dependencyUnavailableError())
		return
	}
	err = h.apps.StrongAuth.RevokeTOTP(request.Context(), identity.RevokeTOTPCommand{
		PrincipalID: authority.PrincipalID, Reauthentication: reauth, IdempotencyKey: params.IdempotencyKey,
	})
	if err != nil {
		h.fail(w, err)
		return
	}
	writeNoContent(w)
}

func (h *Handler) RotateRecoveryCodes(w http.ResponseWriter, request *http.Request, params controlapiv1.RotateRecoveryCodesParams) {
	request, cancel, authority, err := h.authenticatedAccountRequest(request)
	defer cancel()
	if err != nil {
		h.fail(w, err)
		return
	}
	if !validIdempotencyKey(params.IdempotencyKey) {
		h.fail(w, malformedRequestError())
		return
	}
	var body reauthenticationEnvelope
	if err = h.decodeBody(w, request, &body); err != nil {
		h.fail(w, err)
		return
	}
	reauth, err := reauthentication(authority, body.Reauthentication)
	if err != nil {
		h.fail(w, err)
		return
	}
	defer reauth.Proof.Clear()
	if unavailableApplication(h.apps.StrongAuth) {
		h.fail(w, dependencyUnavailableError())
		return
	}
	codes, err := h.apps.StrongAuth.RotateRecoveryCodes(request.Context(), identity.RotateRecoveryCodesCommand{
		PrincipalID: authority.PrincipalID, Reauthentication: reauth, IdempotencyKey: params.IdempotencyKey,
	})
	if err != nil {
		h.fail(w, err)
		return
	}
	h.writeJSON(w, http.StatusCreated, controlapiv1.RecoveryCodes{Codes: codes.Codes})
}

func (h *Handler) ConsumeRecoveryCode(w http.ResponseWriter, request *http.Request, params controlapiv1.ConsumeRecoveryCodeParams) {
	request, cancel, err := h.operationRequest(request)
	defer cancel()
	if err != nil {
		h.fail(w, dependencyUnavailableError())
		return
	}
	if !validIdempotencyKey(params.IdempotencyKey) {
		h.fail(w, malformedRequestError())
		return
	}
	var body controlapiv1.ConsumeRecoveryCodeRequest
	if err = h.decodeBody(w, request, &body); err != nil || !validEmail(string(body.Email)) || !validRecoveryCode(body.Code) {
		if err == nil {
			err = malformedRequestError()
		}
		h.fail(w, err)
		return
	}
	key, err := decodeFixedBase64(body.ClientSigningPublicKey, 32)
	if err != nil {
		h.fail(w, err)
		return
	}
	defer clear(key)
	code := secret.NewBytes([]byte(body.Code))
	defer code.Clear()
	var signingKey [32]byte
	copy(signingKey[:], key)
	if unavailableApplication(h.apps.StrongAuth) {
		h.fail(w, dependencyUnavailableError())
		return
	}
	tokens, err := h.apps.StrongAuth.ConsumeRecoveryCode(request.Context(), identity.ConsumeRecoveryCodeCommand{
		Email: string(body.Email), Code: code, ClientSigningPublicKey: signingKey, IdempotencyKey: params.IdempotencyKey,
	})
	if err != nil {
		h.fail(w, err)
		return
	}
	defer tokens.AccessToken.Clear()
	defer tokens.RefreshToken.Clear()
	response, err := sessionTokensBody(request.Context(), h.apps.AccountAuth, tokens)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, response)
}

func (h *Handler) CreateDeviceEnrollmentGrant(w http.ResponseWriter, request *http.Request, params controlapiv1.CreateDeviceEnrollmentGrantParams) {
	request, cancel, authority, err := h.authenticatedAccountRequest(request)
	defer cancel()
	if err != nil {
		h.fail(w, err)
		return
	}
	if !validIdempotencyKey(params.IdempotencyKey) {
		h.fail(w, malformedRequestError())
		return
	}
	var body reauthenticationEnvelope
	if err = h.decodeBody(w, request, &body); err != nil {
		h.fail(w, err)
		return
	}
	reauth, err := reauthentication(authority, body.Reauthentication)
	if err != nil {
		h.fail(w, err)
		return
	}
	defer reauth.Proof.Clear()
	if unavailableApplication(h.apps.Identity) {
		h.fail(w, dependencyUnavailableError())
		return
	}
	grant, err := h.apps.Identity.CreateEnrollmentGrant(request.Context(), identity.CreateEnrollmentGrantCommand{
		PrincipalID: authority.PrincipalID, SessionID: authority.SessionID, Reauthentication: reauth, IdempotencyKey: params.IdempotencyKey,
	})
	if err != nil {
		h.fail(w, err)
		return
	}
	defer grant.Token.Clear()
	encoded := securitykit.EncodeOpaqueToken(grant.Token)
	if encoded == "" {
		h.fail(w, dependencyUnavailableError())
		return
	}
	h.writeJSON(w, http.StatusCreated, controlapiv1.DeviceEnrollmentGrant{EnrollmentGrant: encoded, ExpiresAt: grant.ExpiresAt})
}

func (h *Handler) CreateDeviceAuthChallenge(w http.ResponseWriter, request *http.Request, params controlapiv1.CreateDeviceAuthChallengeParams) {
	request, cancel, err := h.operationRequest(request)
	defer cancel()
	if err != nil {
		h.fail(w, dependencyUnavailableError())
		return
	}
	if !validIdempotencyKey(params.IdempotencyKey) {
		h.fail(w, malformedRequestError())
		return
	}
	var body deviceChallengeRequest
	if err = h.decodeBody(w, request, &body); err != nil {
		h.fail(w, err)
		return
	}
	nonce, err := decodeFixedBase64(body.RequestNonce, 32)
	if err != nil {
		h.fail(w, err)
		return
	}
	defer clear(nonce)
	command := deviceauth.CreateChallengeCommand{IdempotencyKey: params.IdempotencyKey}
	copy(command.RequestNonce[:], nonce)
	switch {
	case body.EnrollmentGrant != "" && body.RefreshToken == "":
		command.Kind = deviceauth.ChallengeRegistration
		command.EnrollmentGrant, err = decodeOpaque(body.EnrollmentGrant)
		if err != nil {
			h.fail(w, err)
			return
		}
		defer command.EnrollmentGrant.Clear()
		signingKey, keyErr := decodeFixedBase64(body.SigningPublicKey, 32)
		if keyErr != nil {
			h.fail(w, keyErr)
			return
		}
		defer clear(signingKey)
		hpkeKey, hpkeErr := decodeFixedBase64(body.HPKEPublicKey, 32)
		if hpkeErr != nil {
			h.fail(w, hpkeErr)
			return
		}
		defer clear(hpkeKey)
		copy(command.SigningPublicKey[:], signingKey)
		copy(command.HPKEPublicKey[:], hpkeKey)
	case body.RefreshToken != "" && body.EnrollmentGrant == "" && body.SigningPublicKey == "" && body.HPKEPublicKey == "":
		command.Kind = deviceauth.ChallengeRotation
		command.RefreshToken, err = decodeOpaque(body.RefreshToken)
		if err != nil {
			h.fail(w, err)
			return
		}
		defer command.RefreshToken.Clear()
	default:
		h.fail(w, malformedRequestError())
		return
	}
	if unavailableApplication(h.apps.Device) {
		h.fail(w, dependencyUnavailableError())
		return
	}
	challenge, err := h.apps.Device.CreateChallenge(request.Context(), command)
	if err != nil {
		h.fail(w, err)
		return
	}
	response, err := challengeBody(challenge.ChallengeID, challenge.Challenge, challenge.ExpiresAt)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.writeJSON(w, http.StatusCreated, response)
}

func (h *Handler) RegisterDevice(w http.ResponseWriter, request *http.Request, params controlapiv1.RegisterDeviceParams) {
	request, cancel, err := h.operationRequest(request)
	defer cancel()
	if err != nil {
		h.fail(w, dependencyUnavailableError())
		return
	}
	if !validIdempotencyKey(params.IdempotencyKey) {
		h.fail(w, malformedRequestError())
		return
	}
	var body struct {
		ChallengeID      canonicalUUID `json:"challenge_id"`
		DisplayName      string        `json:"display_name"`
		EnrollmentGrant  string        `json:"enrollment_grant"`
		HPKEPublicKey    string        `json:"hpke_public_key"`
		RequestNonce     string        `json:"request_nonce"`
		Signature        string        `json:"signature"`
		SigningPublicKey string        `json:"signing_public_key"`
	}
	if err = h.decodeBody(w, request, &body); err != nil || body.ChallengeID.IsZero() || !utf8.ValidString(body.DisplayName) || len(body.DisplayName) == 0 || len(body.DisplayName) > 256 || utf8.RuneCountInString(body.DisplayName) > 64 {
		if err == nil {
			err = malformedRequestError()
		}
		h.fail(w, err)
		return
	}
	grant, err := decodeOpaque(body.EnrollmentGrant)
	if err != nil {
		h.fail(w, err)
		return
	}
	defer grant.Clear()
	nonce, err := decodeFixedBase64(body.RequestNonce, 32)
	if err != nil {
		h.fail(w, err)
		return
	}
	defer clear(nonce)
	signingKey, err := decodeFixedBase64(body.SigningPublicKey, 32)
	if err != nil {
		h.fail(w, err)
		return
	}
	defer clear(signingKey)
	hpkeKey, err := decodeFixedBase64(body.HPKEPublicKey, 32)
	if err != nil {
		h.fail(w, err)
		return
	}
	defer clear(hpkeKey)
	signature, err := decodeFixedBase64(body.Signature, 64)
	if err != nil {
		h.fail(w, err)
		return
	}
	defer clear(signature)
	command := deviceauth.RegisterDeviceCommand{
		EnrollmentGrant: grant, ChallengeID: body.ChallengeID.String(), DisplayName: body.DisplayName, IdempotencyKey: params.IdempotencyKey,
	}
	copy(command.RequestNonce[:], nonce)
	copy(command.SigningPublicKey[:], signingKey)
	copy(command.HPKEPublicKey[:], hpkeKey)
	copy(command.Signature[:], signature)
	if unavailableApplication(h.apps.Device) {
		h.fail(w, dependencyUnavailableError())
		return
	}
	tokens, err := h.apps.Device.RegisterDevice(request.Context(), command)
	if err != nil {
		h.fail(w, err)
		return
	}
	defer tokens.AccessToken.Clear()
	defer tokens.RefreshToken.Clear()
	response, err := deviceTokensBody(tokens)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.writeJSON(w, http.StatusCreated, response)
}

func (h *Handler) RotateDeviceToken(w http.ResponseWriter, request *http.Request, params controlapiv1.RotateDeviceTokenParams) {
	request, cancel, err := h.operationRequest(request)
	defer cancel()
	if err != nil {
		h.fail(w, dependencyUnavailableError())
		return
	}
	if !validIdempotencyKey(params.IdempotencyKey) {
		h.fail(w, malformedRequestError())
		return
	}
	var body struct {
		ChallengeID  canonicalUUID `json:"challenge_id"`
		RefreshToken string        `json:"refresh_token"`
		RequestNonce string        `json:"request_nonce"`
		Signature    string        `json:"signature"`
	}
	if err = h.decodeBody(w, request, &body); err != nil || body.ChallengeID.IsZero() {
		if err == nil {
			err = malformedRequestError()
		}
		h.fail(w, err)
		return
	}
	refresh, err := decodeOpaque(body.RefreshToken)
	if err != nil {
		h.fail(w, err)
		return
	}
	defer refresh.Clear()
	nonce, err := decodeFixedBase64(body.RequestNonce, 32)
	if err != nil {
		h.fail(w, err)
		return
	}
	defer clear(nonce)
	signature, err := decodeFixedBase64(body.Signature, 64)
	if err != nil {
		h.fail(w, err)
		return
	}
	defer clear(signature)
	command := deviceauth.RotateDeviceTokenCommand{RefreshToken: refresh, ChallengeID: body.ChallengeID.String(), IdempotencyKey: params.IdempotencyKey}
	copy(command.RequestNonce[:], nonce)
	copy(command.Signature[:], signature)
	if unavailableApplication(h.apps.Device) {
		h.fail(w, dependencyUnavailableError())
		return
	}
	tokens, err := h.apps.Device.RotateDeviceToken(request.Context(), command)
	if err != nil {
		h.fail(w, err)
		return
	}
	defer tokens.AccessToken.Clear()
	defer tokens.RefreshToken.Clear()
	response, err := deviceTokensBody(tokens)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, response)
}

func (h *Handler) RevokeDevice(w http.ResponseWriter, request *http.Request, params controlapiv1.RevokeDeviceParams) {
	request, cancel, authority, err := h.authenticatedAccountRequest(request)
	defer cancel()
	if err != nil {
		h.fail(w, err)
		return
	}
	if !validIdempotencyKey(params.IdempotencyKey) {
		h.fail(w, malformedRequestError())
		return
	}
	type revokeRequest struct {
		DeviceID         canonicalUUID           `json:"device_id"`
		Reauthentication reauthenticationRequest `json:"reauthentication"`
	}
	var body revokeRequest
	if err = h.decodeBody(w, request, &body); err != nil || body.DeviceID.IsZero() {
		if err == nil {
			err = malformedRequestError()
		}
		h.fail(w, err)
		return
	}
	reauth, err := reauthentication(authority, body.Reauthentication)
	if err != nil {
		h.fail(w, err)
		return
	}
	defer reauth.Proof.Clear()
	if unavailableApplication(h.apps.Device) {
		h.fail(w, dependencyUnavailableError())
		return
	}
	err = h.apps.Device.RevokeDevice(request.Context(), deviceauth.RevokeDeviceCommand{
		AccountPrincipal: authority.PrincipalID, DeviceID: body.DeviceID.UUID(), Reauthentication: reauth, IdempotencyKey: params.IdempotencyKey,
	})
	if err != nil {
		h.fail(w, err)
		return
	}
	writeNoContent(w)
}

func (h *Handler) authenticatedDeviceRequest(request *http.Request) (*http.Request, context.CancelFunc, deviceauth.BundleAuthority, secret.Bytes, error) {
	request, cancel, err := h.operationRequest(request)
	if err != nil {
		return nil, cancel, deviceauth.BundleAuthority{}, secret.Bytes{}, err
	}
	tokenValue, err := exactBearerToken(request)
	if err != nil {
		cancel()
		return nil, func() {}, deviceauth.BundleAuthority{}, secret.Bytes{}, err
	}
	token, err := securitykit.DecodeOpaqueToken(tokenValue)
	if err != nil {
		cancel()
		return nil, func() {}, deviceauth.BundleAuthority{}, secret.Bytes{}, authenticationFailedError()
	}
	if unavailableApplication(h.apps.DeviceAuth) {
		token.Clear()
		cancel()
		return nil, func() {}, deviceauth.BundleAuthority{}, secret.Bytes{}, dependencyUnavailableError()
	}
	authority, err := h.apps.DeviceAuth.Authenticate(request.Context(), token)
	if err != nil {
		token.Clear()
		cancel()
		return nil, func() {}, deviceauth.BundleAuthority{}, secret.Bytes{}, err
	}
	request = request.WithContext(context.WithValue(request.Context(), deviceAuthorityKey{}, authority))
	return request, cancel, authority, token, nil
}

func (h *Handler) ResolveConfigBundle(w http.ResponseWriter, request *http.Request, params controlapiv1.ResolveConfigBundleParams) {
	request, cancel, authority, token, err := h.authenticatedDeviceRequest(request)
	defer cancel()
	defer token.Clear()
	if err != nil {
		h.fail(w, err)
		return
	}
	if !validIdempotencyKey(params.IdempotencyKey) {
		h.fail(w, malformedRequestError())
		return
	}
	var body emptyRequest
	if err = h.decodeBody(w, request, &body); err != nil {
		h.fail(w, err)
		return
	}
	if unavailableApplication(h.apps.Trust) {
		h.fail(w, dependencyUnavailableError())
		return
	}
	resolution, err := h.apps.Trust.Resolve(request.Context(), trust.ResolveQuery{
		AuthorizationID: authority.AuthorizationID, AccessToken: token, IdempotencyKey: params.IdempotencyKey,
	})
	if err != nil {
		h.fail(w, err)
		return
	}
	available := controlapiv1.ConfigBundleResolution0{
		Status: controlapiv1.Available, BundleLocator: resolution.Locator, EnvelopeSha256: resolution.EnvelopeSHA256,
		Locations: append([]string(nil), resolution.Locations[:]...), CacheControl: resolution.CacheControl,
	}
	var response controlapiv1.ConfigBundleResolution
	if err = response.FromConfigBundleResolution0(available); err != nil {
		h.fail(w, dependencyUnavailableError())
		return
	}
	h.writeJSON(w, http.StatusOK, response)
}

func (h *Handler) AcknowledgeConfigBundle(w http.ResponseWriter, request *http.Request, params controlapiv1.AcknowledgeConfigBundleParams) {
	request, cancel, authority, token, err := h.authenticatedDeviceRequest(request)
	defer cancel()
	defer token.Clear()
	if err != nil {
		h.fail(w, err)
		return
	}
	if !validIdempotencyKey(params.IdempotencyKey) {
		h.fail(w, malformedRequestError())
		return
	}
	var body struct {
		BundleID      canonicalUUID `json:"bundle_id"`
		BundleVersion string        `json:"bundle_version"`
	}
	if err = h.decodeBody(w, request, &body); err != nil || body.BundleID.IsZero() {
		if err == nil {
			err = malformedRequestError()
		}
		h.fail(w, err)
		return
	}
	version, err := strconv.ParseUint(body.BundleVersion, 10, 64)
	if err != nil || version == 0 || strconv.FormatUint(version, 10) != body.BundleVersion {
		h.fail(w, malformedRequestError())
		return
	}
	if unavailableApplication(h.apps.Trust) {
		h.fail(w, dependencyUnavailableError())
		return
	}
	err = h.apps.Trust.Acknowledge(request.Context(), trust.AcknowledgeCommand{
		AuthorizationID: authority.AuthorizationID, AccessToken: token, BundleID: body.BundleID.UUID(),
		BundleVersion: version, IdempotencyKey: params.IdempotencyKey,
	})
	if err != nil {
		h.fail(w, err)
		return
	}
	writeNoContent(w)
}

func (h *Handler) GetImmutableBundle(w http.ResponseWriter, request *http.Request, _ controlapiv1.BundleLocator) {
	request, cancel, err := h.operationRequest(request)
	defer cancel()
	if err != nil || unavailableApplication(h.apps.ImmutableBundle) {
		h.fail(w, dependencyUnavailableError())
		return
	}
	h.apps.ImmutableBundle.ServeHTTP(w, request)
}
