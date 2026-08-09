//nolint:revive // Generated-interface method names document the temporary fail-closed operations.
package controlapi

import (
	"encoding/json"
	"net/http"

	controlapiv1 "talenro.local/platform/gen/go/talenro/controlapi/v1"
)

// C1Unavailable fails closed until each C1.1 operation has a production adapter.
type C1Unavailable struct{}

func writeC1Unavailable(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusServiceUnavailable)
	_ = json.NewEncoder(w).Encode(controlapiv1.PublicError{
		Code:    controlapiv1.PublicErrorCode("dependency_unavailable"),
		Action:  controlapiv1.PublicErrorAction("retry"),
		TraceId: "c1-not-wired-0001",
	})
}

func (C1Unavailable) CreateAccount(w http.ResponseWriter, _ *http.Request, _ controlapiv1.CreateAccountParams) {
	writeC1Unavailable(w)
}

func (C1Unavailable) CreateEmailVerificationDelivery(w http.ResponseWriter, _ *http.Request, _ controlapiv1.CreateEmailVerificationDeliveryParams) {
	writeC1Unavailable(w)
}

func (C1Unavailable) VerifyEmail(w http.ResponseWriter, _ *http.Request, _ controlapiv1.VerifyEmailParams) {
	writeC1Unavailable(w)
}

func (C1Unavailable) CreatePasswordResetDelivery(w http.ResponseWriter, _ *http.Request, _ controlapiv1.CreatePasswordResetDeliveryParams) {
	writeC1Unavailable(w)
}

func (C1Unavailable) ResetPassword(w http.ResponseWriter, _ *http.Request, _ controlapiv1.ResetPasswordParams) {
	writeC1Unavailable(w)
}

func (C1Unavailable) ChangePassword(w http.ResponseWriter, _ *http.Request, _ controlapiv1.ChangePasswordParams) {
	writeC1Unavailable(w)
}

func (C1Unavailable) CreateAccountSession(w http.ResponseWriter, _ *http.Request, _ controlapiv1.CreateAccountSessionParams) {
	writeC1Unavailable(w)
}

func (C1Unavailable) CreateAccountAuthChallenge(w http.ResponseWriter, _ *http.Request, _ controlapiv1.CreateAccountAuthChallengeParams) {
	writeC1Unavailable(w)
}

func (C1Unavailable) RotateAccountToken(w http.ResponseWriter, _ *http.Request, _ controlapiv1.RotateAccountTokenParams) {
	writeC1Unavailable(w)
}

func (C1Unavailable) RevokeAccountSessions(w http.ResponseWriter, _ *http.Request, _ controlapiv1.RevokeAccountSessionsParams) {
	writeC1Unavailable(w)
}

func (C1Unavailable) CreatePasskeyRegistrationOptions(w http.ResponseWriter, _ *http.Request, _ controlapiv1.CreatePasskeyRegistrationOptionsParams) {
	writeC1Unavailable(w)
}

func (C1Unavailable) CreatePasskeyCredential(w http.ResponseWriter, _ *http.Request, _ controlapiv1.CreatePasskeyCredentialParams) {
	writeC1Unavailable(w)
}

func (C1Unavailable) CreatePasskeyAuthenticationOptions(w http.ResponseWriter, _ *http.Request, _ controlapiv1.CreatePasskeyAuthenticationOptionsParams) {
	writeC1Unavailable(w)
}

func (C1Unavailable) RevokePasskey(w http.ResponseWriter, _ *http.Request, _ controlapiv1.RevokePasskeyParams) {
	writeC1Unavailable(w)
}

func (C1Unavailable) CreateTOTPEnrollment(w http.ResponseWriter, _ *http.Request, _ controlapiv1.CreateTOTPEnrollmentParams) {
	writeC1Unavailable(w)
}

func (C1Unavailable) VerifyTOTPEnrollment(w http.ResponseWriter, _ *http.Request, _ controlapiv1.VerifyTOTPEnrollmentParams) {
	writeC1Unavailable(w)
}

func (C1Unavailable) RevokeTOTP(w http.ResponseWriter, _ *http.Request, _ controlapiv1.RevokeTOTPParams) {
	writeC1Unavailable(w)
}

func (C1Unavailable) RotateRecoveryCodes(w http.ResponseWriter, _ *http.Request, _ controlapiv1.RotateRecoveryCodesParams) {
	writeC1Unavailable(w)
}

func (C1Unavailable) ConsumeRecoveryCode(w http.ResponseWriter, _ *http.Request, _ controlapiv1.ConsumeRecoveryCodeParams) {
	writeC1Unavailable(w)
}

func (C1Unavailable) CreateDeviceEnrollmentGrant(w http.ResponseWriter, _ *http.Request, _ controlapiv1.CreateDeviceEnrollmentGrantParams) {
	writeC1Unavailable(w)
}

func (C1Unavailable) CreateDeviceAuthChallenge(w http.ResponseWriter, _ *http.Request, _ controlapiv1.CreateDeviceAuthChallengeParams) {
	writeC1Unavailable(w)
}

func (C1Unavailable) RegisterDevice(w http.ResponseWriter, _ *http.Request, _ controlapiv1.RegisterDeviceParams) {
	writeC1Unavailable(w)
}

func (C1Unavailable) RotateDeviceToken(w http.ResponseWriter, _ *http.Request, _ controlapiv1.RotateDeviceTokenParams) {
	writeC1Unavailable(w)
}

func (C1Unavailable) RevokeDevice(w http.ResponseWriter, _ *http.Request, _ controlapiv1.RevokeDeviceParams) {
	writeC1Unavailable(w)
}

func (C1Unavailable) ResolveConfigBundle(w http.ResponseWriter, _ *http.Request, _ controlapiv1.ResolveConfigBundleParams) {
	writeC1Unavailable(w)
}

func (C1Unavailable) AcknowledgeConfigBundle(w http.ResponseWriter, _ *http.Request, _ controlapiv1.AcknowledgeConfigBundleParams) {
	writeC1Unavailable(w)
}

func (C1Unavailable) GetImmutableBundle(w http.ResponseWriter, _ *http.Request, _ controlapiv1.BundleLocator) {
	writeC1Unavailable(w)
}
