//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/pquerna/otp/totp"
	"talenro.local/platform/internal/buildinfo"
	"talenro.local/platform/internal/errorreport"
	"talenro.local/platform/internal/observability"
)

const (
	fixturePrivacyCaptureLimit    = 64 << 10
	fixturePrivacyLogBaseline     = "c11 privacy log sink exercised"
	fixturePrivacyProviderSurface = "provider-error"
	fixturePrivacyPanicSurface    = "provider-panic"
	fixturePrivacyLogEvidence     = 1 << iota
	fixturePrivacyReportEvidence
	fixturePrivacyProviderEvidence
	fixturePrivacyPanicEvidence
	fixturePrivacyAllEvidence = fixturePrivacyLogEvidence | fixturePrivacyReportEvidence |
		fixturePrivacyProviderEvidence | fixturePrivacyPanicEvidence
)

type privacyCapture struct {
	mu       sync.Mutex
	logs     bytes.Buffer
	reports  bytes.Buffer
	sources  bytes.Buffer
	evidence int
	overflow bool
}

func newPrivacyCapture() *privacyCapture { return &privacyCapture{} }

func (capture *privacyCapture) Write(value []byte) (int, error) {
	if capture == nil {
		return len(value), nil
	}
	capture.mu.Lock()
	defer capture.mu.Unlock()
	if bytes.Contains(value, []byte(fixturePrivacyLogBaseline)) {
		capture.evidence |= fixturePrivacyLogEvidence
	}
	capture.appendBounded(&capture.logs, value)
	return len(value), nil
}

func (capture *privacyCapture) recordReports(reports []errorreport.Report) {
	if capture == nil || len(reports) == 0 {
		return
	}
	encoded, err := json.Marshal(reports)
	capture.mu.Lock()
	defer capture.mu.Unlock()
	if err != nil {
		capture.overflow = true
		return
	}
	capture.evidence |= fixturePrivacyReportEvidence
	capture.appendBounded(&capture.reports, encoded)
}

func (capture *privacyCapture) recordSource(surface, value string) {
	if capture == nil || value == "" {
		return
	}
	capture.mu.Lock()
	defer capture.mu.Unlock()
	switch surface {
	case fixturePrivacyProviderSurface:
		capture.evidence |= fixturePrivacyProviderEvidence
	case fixturePrivacyPanicSurface:
		capture.evidence |= fixturePrivacyPanicEvidence
	default:
		capture.overflow = true
		return
	}
	capture.appendBounded(&capture.sources, []byte(surface+"="+value+"\n"))
}

func (capture *privacyCapture) appendBounded(destination *bytes.Buffer, value []byte) {
	if capture.overflow || destination == nil || len(value) > fixturePrivacyCaptureLimit-destination.Len() {
		capture.overflow = true
		return
	}
	_, _ = destination.Write(value)
}

func (capture *privacyCapture) scan(canaries []string) (bool, int) {
	if capture == nil {
		return false, 0
	}
	capture.mu.Lock()
	defer capture.mu.Unlock()
	if capture.overflow || capture.evidence != fixturePrivacyAllEvidence || capture.sources.Len() == 0 {
		return false, capture.evidence
	}
	sinks := make([]byte, 0, capture.logs.Len()+capture.reports.Len())
	sinks = append(sinks, capture.logs.Bytes()...)
	sinks = append(sinks, capture.reports.Bytes()...)
	for _, canary := range canaries {
		if canary == "" || bytes.Contains(sinks, []byte(canary)) {
			return false, capture.evidence
		}
	}
	return true, capture.evidence
}

func TestC11PrivacyAndCardinalityMatrix(t *testing.T) {
	fixture := requireC11Fixture(t)
	fixture.runPrivacyMatrix(t)
}

func TestVerifyC11PowerShellContract(t *testing.T) {
	requirePowerShellContract(t)
}

func TestVerifyC11BashContract(t *testing.T) {
	requireBashContract(t)
}

func TestScriptCleanupExitStatusContracts(t *testing.T) {
	repoRoot := e2eRepositoryRoot(t)
	harnessDirectory := t.TempDir()
	powerShellHarness := filepath.Join(harnessDirectory, "cleanup-contract.ps1")
	bashHarness := filepath.Join(harnessDirectory, "cleanup-contract.sh")
	writeFakeTool(t, powerShellHarness, cleanupContractPowerShell())
	writeFakeTool(t, bashHarness, cleanupContractBash())

	testCases := []struct {
		name          string
		primaryExit   int
		downExit      int
		preLocalExit  int
		postLocalExit int
		smokeOnly     bool
		want          int
	}{
		{name: "sole_down_failure", downExit: 73, want: 73},
		{name: "sole_down_timeout", downExit: 124, want: 124},
		{name: "primary_precedes_down", primaryExit: 73, downExit: 91, want: 73},
		{name: "local_cleanup_failure", postLocalExit: 1, want: 1},
		{name: "pre_local_then_down", preLocalExit: 1, downExit: 73, smokeOnly: true, want: 73},
		{name: "down_then_post_local", downExit: 124, postLocalExit: 1, want: 124},
	}
	entrypoints := []struct {
		name  string
		mode  string
		path  string
		shell string
	}{
		{name: "PowerShell/verify", mode: "verify", path: filepath.Join(repoRoot, "scripts", "verify-c11.ps1"), shell: "powershell"},
		{name: "PowerShell/smoke", mode: "smoke", path: filepath.Join(repoRoot, "scripts", "smoke.ps1"), shell: "powershell"},
		{name: "Bash/verify", mode: "verify", path: filepath.Join(repoRoot, "scripts", "verify-c11.sh"), shell: "bash"},
		{name: "Bash/smoke", mode: "smoke", path: filepath.Join(repoRoot, "scripts", "smoke.sh"), shell: "bash"},
	}
	for _, entrypoint := range entrypoints {
		for _, testCase := range testCases {
			if testCase.smokeOnly && entrypoint.mode != "smoke" {
				continue
			}
			t.Run(entrypoint.name+"/"+testCase.name, func(t *testing.T) {
				requireScriptCleanupExitStatus(t, entrypoint.shell, entrypoint.mode, entrypoint.path,
					powerShellHarness, bashHarness, harnessDirectory, testCase.primaryExit,
					testCase.downExit, testCase.preLocalExit, testCase.postLocalExit, testCase.want)
			})
		}
	}
}

func TestPrivacyCaptureRequiresExercisedCleanSinks(t *testing.T) {
	t.Parallel()
	capture := newPrivacyCapture()
	_, _ = capture.Write([]byte(fixturePrivacyLogBaseline))
	capture.recordReports([]errorreport.Report{fixtureFiniteReport()})
	capture.recordSource(fixturePrivacyProviderSurface, "C11-PROVIDER-SOURCE-PRIVATE")
	capture.recordSource(fixturePrivacyPanicSurface, "C11-PANIC-SOURCE-PRIVATE")
	clean, evidence := capture.scan([]string{"C11-PROVIDER-SOURCE-PRIVATE", "C11-PANIC-SOURCE-PRIVATE"})
	if !clean || evidence != fixturePrivacyAllEvidence {
		t.Fatal("clean exercised privacy capture was rejected")
	}
	_, _ = capture.Write([]byte("C11-PROVIDER-SOURCE-PRIVATE"))
	if clean, _ := capture.scan([]string{"C11-PROVIDER-SOURCE-PRIVATE"}); clean {
		t.Fatal("privacy capture accepted a log leak")
	}
}

func TestPrivacySurfaceCanariesCoverEachSensitiveClass(t *testing.T) {
	t.Parallel()
	canaries := privacySurfaceCanaries()
	want := []string{
		"email", "password", "token", "nonce", "principal_id", "device_id", "locator", "key", "ciphertext",
		"url_credential", "provider_error", "provider_panic", "http_body",
	}
	if len(canaries) != len(want) {
		t.Fatal("privacy surface canary matrix is incomplete")
	}
	seen := make(map[string]struct{}, len(canaries))
	for _, name := range want {
		value := canaries[name]
		if !validPrivacyCanary(value) {
			t.Fatalf("privacy surface %s has no valid canary", name)
		}
		if _, duplicate := seen[value]; duplicate {
			t.Fatalf("privacy surface %s reused another canary", name)
		}
		seen[value] = struct{}{}
	}
}

func TestVerifyFakeToolsExerciseEveryPrivacyCanary(t *testing.T) {
	t.Parallel()
	for name, fake := range map[string]string{
		"PowerShell": fakeGoBatch(`C:\\task19\\calls.log`),
		"Bash":       fakeGoShell(),
	} {
		for surface, canary := range privacySurfaceCanaries() {
			if strings.Count(fake, canary) < 2 {
				t.Fatalf("verify-c11 %s fake did not inject %s into both command streams", name, surface)
			}
		}
	}
}

func TestVerifyCommandContractRejectsLegacyMissingRound2Gates(t *testing.T) {
	t.Parallel()
	required := []string{
		"go test -tags=integration", "go build -o", "docker compose --project-name owned config --quiet",
		"docker compose -f deploy/dev/compose.yaml config --quiet", "go test -tags=e2e",
	}
	legacy := "go test -tags=integration\ndocker compose --project-name owned config --quiet\ngo test -tags=e2e\n"
	if validateCommandOrder(legacy, required) == nil {
		t.Fatal("verify-c11 command contract accepted the legacy missing gates")
	}
	complete := "go test -tags=integration\ngo build -o owned/trust-conformance ./cmd/trust-conformance\n" +
		"docker compose --project-name owned config --quiet\ndocker compose -f deploy/dev/compose.yaml config --quiet\n" +
		"go test -tags=e2e\n"
	if validateCommandOrder(complete, required) != nil {
		t.Fatal("verify-c11 command contract rejected the complete Round2 gate order")
	}
}

func TestFiniteMetricContractRejectsUnknownLabelsAndUnboundedSeries(t *testing.T) {
	t.Parallel()
	info := buildinfo.Current()
	valid := []byte("# HELP talenro_control_build_info Talenro control API build identity.\n" +
		"# TYPE talenro_control_build_info gauge\n" +
		fmt.Sprintf("talenro_control_build_info{commit=%s,version=%s} 1\n", strconv.Quote(info.Commit), strconv.Quote(info.Version)))
	if err := validateFiniteMetrics(valid); err != nil {
		t.Fatal("finite metric contract rejected an allowlisted family")
	}
	for _, invalid := range [][]byte{
		bytes.ReplaceAll(valid, []byte("commit="+strconv.Quote(info.Commit)), []byte("commit=\"C11-METRIC-CANARY\"")),
		append(append([]byte(nil), valid...), []byte("talenro_unbounded_private_total{identifier=\"C11\"} 1\n")...),
		append(bytes.Repeat([]byte(fmt.Sprintf("talenro_control_build_info{commit=%s,version=%s} 1\n", strconv.Quote(info.Commit), strconv.Quote(info.Version))), 2), valid...),
	} {
		if err := validateFiniteMetrics(invalid); err == nil {
			t.Fatal("finite metric contract accepted an unbounded metric surface")
		}
	}
}

func TestFiniteMetricContractAcceptsRealRegistryExposition(t *testing.T) {
	t.Parallel()
	registry := observability.NewRegistry()
	registry.Requests.WithLabelValues("GET", "/livez", "2xx").Inc()
	registry.Duration.WithLabelValues("GET", "/livez", "2xx").Observe(0.01)
	if !registry.RecordSecurity(observability.SecurityOperationAccountAuth, observability.MetricResultFailure, observability.SecurityReasonInvalidCredential) ||
		!registry.RecordCrypto(observability.CryptoOperationTokenVerify, observability.MetricResultFailure, observability.CryptoReasonInvalid) ||
		!registry.SetOutbox(1, time.Second) {
		t.Fatal("real metric registry fixture failed")
	}
	registry.Record(errorreport.ComponentControlAPI, errorreport.ResultProviderFailed)
	recorder := httptest.NewRecorder()
	registry.Handler().ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/metrics", nil))
	if recorder.Code != http.StatusOK || validateFiniteMetrics(recorder.Body.Bytes()) != nil {
		t.Fatal("finite metric contract rejected the real registry exposition")
	}
}

func privacySurfaceCanaries() map[string]string {
	return map[string]string{ //nolint:gosec // Deliberate non-secret synthetic privacy canaries.
		"email":          "c11-email-private-19@example.test",
		"password":       "C11-PASSWORD-PRIVATE-19!Aa7",
		"token":          privacyBase64Canary("C11-TOKEN-PRIVATE-19"),
		"nonce":          privacyBase64Canary("C11-NONCE-PRIVATE-19"),
		"principal_id":   "01900000-0000-7000-8000-000000000019",
		"device_id":      "01900000-0000-7000-9000-000000000019",
		"locator":        privacyBase64Canary("C11-LOCATOR-PRIVATE-19"),
		"key":            privacyBase64Canary("C11-KEY-PRIVATE-19"),
		"ciphertext":     privacyBase64Canary("C11-CIPHERTEXT-PRIVATE-19"),
		"url_credential": "C11-URL-CREDENTIAL-PRIVATE-19",
		"provider_error": "C11-PROVIDER-ERROR-PRIVATE-19",
		"provider_panic": "C11-PROVIDER-PANIC-PRIVATE-19",
		"http_body":      "C11-RAW-HTTP-BODY-PRIVATE-19",
	}
}

func privacyBase64Canary(label string) string {
	value := make([]byte, 32)
	copy(value, label)
	for index := len(label); index < len(value); index++ {
		value[index] = 'x'
	}
	return base64.RawURLEncoding.EncodeToString(value)
}

func c11PasskeyRegistrationResponse(t *testing.T, challenge, origin, rpID string, credentialID []byte) json.RawMessage {
	t.Helper()
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal("c11 passkey key generation failed")
	}
	clientData, err := json.Marshal(map[string]any{
		"type": "webauthn.create", "challenge": challenge, "origin": origin, "crossOrigin": false,
	})
	if err != nil {
		t.Fatal("c11 passkey client data failed")
	}
	publicKey := c11COSEPublicKey(t, &privateKey.PublicKey)
	rpIDHash := sha256.Sum256([]byte(rpID))
	authenticatorData := make([]byte, 0, 32+1+4+16+2+len(credentialID)+len(publicKey))
	authenticatorData = append(authenticatorData, rpIDHash[:]...)
	authenticatorData = append(authenticatorData, byte(protocol.FlagUserPresent|protocol.FlagUserVerified|protocol.FlagAttestedCredentialData))
	authenticatorData = binary.BigEndian.AppendUint32(authenticatorData, 0)
	authenticatorData = append(authenticatorData, make([]byte, 16)...)
	// #nosec G115 -- the generated credential ID is fixed at 32 bytes.
	authenticatorData = binary.BigEndian.AppendUint16(authenticatorData, uint16(len(credentialID)))
	authenticatorData = append(authenticatorData, credentialID...)
	authenticatorData = append(authenticatorData, publicKey...)
	attestationObject, err := cbor.Marshal(map[string]any{"fmt": "none", "attStmt": map[string]any{}, "authData": authenticatorData})
	if err != nil {
		t.Fatal("c11 passkey attestation failed")
	}
	encodedID := base64.RawURLEncoding.EncodeToString(credentialID)
	response, err := json.Marshal(map[string]any{
		"id": encodedID, "rawId": encodedID, "type": "public-key", "authenticatorAttachment": "platform",
		"clientExtensionResults": map[string]any{}, "response": map[string]any{
			"clientDataJSON":    base64.RawURLEncoding.EncodeToString(clientData),
			"attestationObject": base64.RawURLEncoding.EncodeToString(attestationObject), "transports": []string{"internal"},
		},
	})
	if err != nil {
		t.Fatal("c11 passkey response failed")
	}
	return response
}

func c11COSEPublicKey(t *testing.T, publicKey *ecdsa.PublicKey) []byte {
	t.Helper()
	encoded, err := publicKey.Bytes()
	if err != nil || len(encoded) != 65 || encoded[0] != 4 {
		t.Fatal("c11 passkey public key failed")
	}
	xCoordinate := append([]byte(nil), encoded[1:33]...)
	yCoordinate := append([]byte(nil), encoded[33:65]...)
	defer clear(xCoordinate)
	defer clear(yCoordinate)
	result, err := cbor.Marshal(map[int]any{1: 2, 3: -7, -1: 1, -2: xCoordinate, -3: yCoordinate})
	if err != nil {
		t.Fatal("c11 passkey COSE key failed")
	}
	return result
}

func c11TOTPCode(t *testing.T, secret string, now time.Time) string {
	t.Helper()
	code, err := totp.GenerateCode(secret, now)
	if err != nil || len(code) != 6 {
		t.Fatal("c11 TOTP fixture failed")
	}
	return code
}

func (fixture *c11Fixture) runPrivacyMatrix(t *testing.T) {
	t.Helper()
	canaries := privacySurfaceCanaries()
	canaryValues := privacyCanaryValues(canaries)
	client := &c11HTTPClient{t: t, client: fixture.client}
	account := fixture.runAccountSecurity(t, client)
	responses := bytes.Buffer{}
	exercised := make(map[string]int, len(canaries))
	request := func(name, method, target, bearer, idempotency string, body []byte, expected ...int) {
		t.Helper()
		output := executePrivacyHTTPRequest(t, fixture.client, method, target, bearer, idempotency, body, expected...)
		appendPrivacySink(t, &responses, output)
		exercised[name]++
	}
	request("email", http.MethodPost, fixture.ready.RequiredURL+"/v1/accounts", "", nextFixtureKey("privacy-register"), privacyJSON(t, map[string]any{
		"email": canaries["email"], "password": canaries["password"], "locale": "en",
	}), http.StatusAccepted)
	exercised["password"]++
	request("token", http.MethodPost, fixture.ready.RequiredURL+"/v1/email-verifications", "", nextFixtureKey("privacy-token"), privacyJSON(t, map[string]any{
		"token": canaries["token"],
	}), http.StatusUnauthorized)
	request("nonce", http.MethodPost, fixture.ready.RequiredURL+"/v1/device-auth-challenges", "", nextFixtureKey("privacy-nonce"), privacyJSON(t, map[string]any{
		"refresh_token": privacyBase64Canary("C11-SAFE-REFRESH-19"), "request_nonce": canaries["nonce"],
	}), http.StatusUnauthorized)
	request("principal_id", http.MethodPost, fixture.ready.RequiredURL+"/v1/account-session-revocations", account.accessToken, nextFixtureKey("privacy-principal"), privacyJSON(t, map[string]any{
		"scope": "all", "principal_id": canaries["principal_id"],
		"reauthentication": map[string]any{"method": "password", "password": "Task19-Privacy-Wrong!71"}, //nolint:gosec // Synthetic negative fixture credential.
	}), http.StatusBadRequest)
	request("device_id", http.MethodPost, fixture.ready.RequiredURL+"/v1/device-revocations", account.accessToken, nextFixtureKey("privacy-device"), privacyJSON(t, map[string]any{
		"device_id":        canaries["device_id"],
		"reauthentication": map[string]any{"method": "password", "password": "Task19-Privacy-Wrong!72"}, //nolint:gosec // Synthetic negative fixture credential.
	}), http.StatusUnauthorized)
	request("locator", http.MethodGet, fixture.ready.MirrorAURL+"/b/"+url.PathEscape(canaries["locator"]), "", "", nil, http.StatusNotFound)
	request("key", http.MethodPost, fixture.ready.RequiredURL+"/v1/account-sessions", "", nextFixtureKey("privacy-key"), privacyJSON(t, map[string]any{ //nolint:gosec // Synthetic negative fixture credential.
		"method": "password", "email": "privacy-key-missing@example.test", "password": "Task19-Privacy-Safe!73",
		"client_signing_public_key": canaries["key"],
	}), http.StatusUnauthorized)
	request("ciphertext", http.MethodPost, fixture.ready.RequiredURL+"/v1/accounts", "", nextFixtureKey("privacy-ciphertext"), privacyJSON(t, map[string]any{ //nolint:gosec // Synthetic negative fixture credential.
		"email": "privacy-ciphertext@example.test", "password": "Task19-Privacy-Safe!74", "locale": "en",
		"ciphertext": canaries["ciphertext"],
	}), http.StatusBadRequest)
	request("url_credential", http.MethodGet, fixture.ready.RequiredURL+"/livez?access_token="+url.QueryEscape(canaries["url_credential"]), "", "", nil, http.StatusOK)
	request("http_body", http.MethodPost, fixture.ready.RequiredURL+"/v1/accounts", "", nextFixtureKey("privacy-raw"), []byte(`{"raw":"`+canaries["http_body"]+`"`), http.StatusBadRequest)
	for _, name := range []string{"email", "password", "token", "nonce", "principal_id", "device_id", "locator", "key", "ciphertext", "url_credential", "http_body"} {
		if exercised[name] != 1 {
			t.Fatal("c11 privacy HTTP surface was not exercised exactly once")
		}
	}
	assertNoPrivacyCanaries(t, responses.Bytes(), canaryValues)

	successBefore := fixtureOperationCount(t, fixture.child, "reporter-attempts", fixtureModeSuccess)
	if _, err := fixture.child.call("privacy-baseline", "", "", ""); err != nil {
		t.Fatal("c11 privacy baseline capture failed")
	}
	waitFixtureOperationIncrease(t, fixture.child, "reporter-attempts", fixtureModeSuccess, successBefore, 5*time.Second)
	if _, err := fixture.child.call("privacy-canaries", canaries["provider_error"], "", canaries["provider_panic"]); err != nil {
		t.Fatal("c11 privacy provider canary configuration failed")
	}
	for _, mode := range []string{fixtureModeReject, fixtureModePanic} {
		before := fixtureOperationCount(t, fixture.child, "reporter-attempts", mode)
		if _, err := fixture.child.call("reporter-mode", mode, "", ""); err != nil {
			t.Fatal("c11 privacy reporter injection failed")
		}
		if _, err := fixture.child.call("report", "", "", ""); err != nil {
			t.Fatal("c11 privacy reporter exercise failed")
		}
		waitFixtureOperationIncrease(t, fixture.child, "reporter-attempts", mode, before, 5*time.Second)
	}
	if _, err := fixture.child.call("reporter-mode", fixtureModeSuccess, "", ""); err != nil {
		t.Fatal("c11 privacy reporter recovery failed")
	}
	recoveryBefore := fixtureOperationCount(t, fixture.child, "reporter-attempts", fixtureModeSuccess)
	if _, err := fixture.child.call("report", "", "", ""); err != nil {
		t.Fatal("c11 privacy reporter recovery exercise failed")
	}
	waitFixtureOperationIncrease(t, fixture.child, "reporter-attempts", fixtureModeSuccess, recoveryBefore, 5*time.Second)
	assertPrivacyChildCapture(t, fixture.child, canaryValues)

	metrics, _ := client.get(fixture.ready.MetricsURL+"/metrics", http.StatusOK)
	assertNoPrivacyCanaries(t, metrics, canaryValues)
	assertFiniteMetricLabels(t, metrics)
	requirePrivacyMetricFamilies(t, metrics)
}

func privacyCanaryValues(canaries map[string]string) []string {
	names := []string{
		"email", "password", "token", "nonce", "principal_id", "device_id", "locator", "key", "ciphertext",
		"url_credential", "provider_error", "provider_panic", "http_body",
	}
	values := make([]string, 0, len(names))
	for _, name := range names {
		values = append(values, canaries[name])
	}
	return values
}

func privacyJSON(t *testing.T, value any) []byte {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil || len(encoded) == 0 || len(encoded) > fixturePrivacyCaptureLimit {
		t.Fatal("c11 privacy request encoding failed")
	}
	return encoded
}

func executePrivacyHTTPRequest(
	t *testing.T,
	client *http.Client,
	method, target, bearer, idempotency string,
	body []byte,
	expected ...int,
) []byte {
	t.Helper()
	if client == nil || len(body) > fixturePrivacyCaptureLimit || len(expected) == 0 {
		t.Fatal("c11 privacy HTTP contract invalid")
	}
	requestContext, cancel := context.WithTimeout(t.Context(), fixtureRequestTimeout)
	defer cancel()
	request, err := http.NewRequestWithContext(requestContext, method, target, bytes.NewReader(body))
	if err != nil {
		t.Fatal("c11 privacy HTTP request construction failed")
	}
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if bearer != "" {
		request.Header.Set("Authorization", "Bearer "+bearer)
	}
	if idempotency != "" {
		request.Header.Set("Idempotency-Key", idempotency)
	}
	response, err := client.Do(request)
	clear(body)
	if err != nil {
		t.Fatal("c11 privacy HTTP request failed")
	}
	defer func() { _ = response.Body.Close() }()
	accepted := false
	for _, status := range expected {
		accepted = accepted || response.StatusCode == status
	}
	if !accepted {
		t.Fatal("c11 privacy HTTP status contract failed")
	}
	responseBody, err := io.ReadAll(io.LimitReader(response.Body, fixturePrivacyCaptureLimit+1))
	if err != nil || len(responseBody) > fixturePrivacyCaptureLimit {
		clear(responseBody)
		t.Fatal("c11 privacy HTTP response bound failed")
	}
	output := bytes.Buffer{}
	if err := response.Header.Write(&output); err != nil || output.Len()+len(responseBody) > 2*fixturePrivacyCaptureLimit {
		clear(responseBody)
		t.Fatal("c11 privacy HTTP header capture failed")
	}
	_, _ = output.Write(responseBody)
	clear(responseBody)
	return output.Bytes()
}

func appendPrivacySink(t *testing.T, destination *bytes.Buffer, value []byte) {
	t.Helper()
	if destination == nil || destination.Len()+len(value) > 1<<20 {
		t.Fatal("c11 privacy aggregate capture bound failed")
	}
	_, _ = destination.Write(value)
}

func assertNoPrivacyCanaries(t *testing.T, sink []byte, canaries []string) {
	t.Helper()
	for _, canary := range canaries {
		if canary == "" || bytes.Contains(sink, []byte(canary)) {
			t.Fatal("c11 privacy canary appeared in an observable sink")
		}
	}
}

func assertPrivacyChildCapture(t *testing.T, child *fixtureChild, canaries []string) {
	t.Helper()
	encoded, err := json.Marshal(canaries)
	if err != nil || len(encoded) > fixtureCommandLimit {
		t.Fatal("c11 privacy child scan request failed")
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		response, callErr := child.call("privacy-scan", string(encoded), "", "")
		if callErr != nil {
			t.Fatal("c11 privacy child scan failed")
		}
		if response.Count == fixturePrivacyAllEvidence {
			if response.Value != "true" {
				t.Fatal("c11 privacy child sink disclosed a canary")
			}
			return
		}
		if !time.Now().Before(deadline) {
			t.Fatal("c11 privacy child sink evidence deadline exceeded")
		}
		time.Sleep(25 * time.Millisecond)
	}
}

func requirePrivacyMetricFamilies(t *testing.T, metrics []byte) {
	t.Helper()
	for _, family := range []string{
		"talenro_control_http_requests_total", "talenro_control_http_request_duration_seconds",
		"talenro_control_build_info", "talenro_security_events_total", "talenro_error_reports_total",
	} {
		if !bytes.Contains(metrics, []byte("# HELP "+family+" ")) || !bytes.Contains(metrics, []byte("# TYPE "+family+" ")) {
			t.Fatal("c11 privacy metric family was not exercised")
		}
	}
}

func assertFiniteMetricLabels(t *testing.T, metrics []byte) {
	t.Helper()
	if err := validateFiniteMetrics(metrics); err != nil {
		t.Fatal("c11 metrics exceeded the finite privacy contract")
	}
}

type finiteMetricContract struct {
	help      string
	typeName  string
	maxSeries int
}

var finiteMetricContracts = map[string]finiteMetricContract{
	"talenro_control_http_requests_total": {
		help: "Completed control API HTTP requests.", typeName: "counter", maxSeries: 30,
	},
	"talenro_control_http_request_duration_seconds": {
		help: "Control API HTTP request duration.", typeName: "histogram", maxSeries: 420,
	},
	"talenro_control_build_info": {
		help: "Talenro control API build identity.", typeName: "gauge", maxSeries: 1,
	},
	"talenro_security_events_total": {
		help: "Security outcomes by finite operation, result, and reason.", typeName: "counter", maxSeries: 128,
	},
	"talenro_outbox_backlog": {
		help: "Current number of unpublished transactional outbox events.", typeName: "gauge", maxSeries: 1,
	},
	"talenro_outbox_oldest_seconds": {
		help: "Current age in seconds of the oldest unpublished outbox event.", typeName: "gauge", maxSeries: 1,
	},
	"talenro_error_reports_total": {
		help: "Privacy-safe error report delivery outcomes by finite component and result.", typeName: "counter", maxSeries: 21,
	},
	"talenro_crypto_validations_total": {
		help: "Cryptographic validation outcomes by finite operation, result, and reason.", typeName: "counter", maxSeries: 144,
	},
}

var finiteMetricLabelPattern = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*)=("(?:\\.|[^"\\])*")`)

func validateFiniteMetrics(metrics []byte) error {
	if len(metrics) == 0 || len(metrics) > 2<<20 {
		return errors.New("metric exposition bound failed")
	}
	helps := make(map[string]string, len(finiteMetricContracts))
	types := make(map[string]string, len(finiteMetricContracts))
	series := make(map[string]int, len(finiteMetricContracts))
	for _, rawLine := range strings.Split(string(metrics), "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "# HELP ") {
			family, help, ok := cutMetricDescriptor(strings.TrimPrefix(line, "# HELP "))
			contract, known := finiteMetricContracts[family]
			if !ok || !known || help != contract.help || helps[family] != "" {
				return errors.New("metric help contract failed")
			}
			helps[family] = help
			continue
		}
		if strings.HasPrefix(line, "# TYPE ") {
			family, typeName, ok := cutMetricDescriptor(strings.TrimPrefix(line, "# TYPE "))
			contract, known := finiteMetricContracts[family]
			if !ok || !known || typeName != contract.typeName || types[family] != "" {
				return errors.New("metric type contract failed")
			}
			types[family] = typeName
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		sampleName, labels, value, err := parseMetricSample(line)
		if err != nil {
			return err
		}
		family := metricFamily(sampleName)
		contract, known := finiteMetricContracts[family]
		if !known || !strings.HasPrefix(sampleName, "talenro_") || math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
			return errors.New("metric sample contract failed")
		}
		if err := validateMetricLabels(family, sampleName, labels); err != nil {
			return err
		}
		series[family]++
		if series[family] > contract.maxSeries {
			return errors.New("metric series bound failed")
		}
	}
	for family, count := range series {
		if count == 0 || helps[family] == "" || types[family] == "" {
			return errors.New("metric descriptor missing")
		}
	}
	for family := range helps {
		if types[family] == "" {
			return errors.New("metric descriptor pair missing")
		}
	}
	for family := range types {
		if helps[family] == "" {
			return errors.New("metric descriptor pair missing")
		}
	}
	return nil
}

func cutMetricDescriptor(value string) (string, string, bool) {
	family, description, found := strings.Cut(value, " ")
	return family, description, found && family != "" && description != ""
}

func parseMetricSample(line string) (string, map[string]string, float64, error) {
	nameEnd := strings.IndexAny(line, "{ ")
	if nameEnd <= 0 {
		return "", nil, 0, errors.New("metric sample name failed")
	}
	name := line[:nameEnd]
	remainder := line[nameEnd:]
	labels := map[string]string{}
	if remainder[0] == '{' {
		closeIndex := strings.IndexByte(remainder, '}')
		if closeIndex <= 1 {
			return "", nil, 0, errors.New("metric labels failed")
		}
		var err error
		labels, err = parseMetricLabels(remainder[1:closeIndex])
		if err != nil {
			return "", nil, 0, err
		}
		remainder = remainder[closeIndex+1:]
	}
	fields := strings.Fields(remainder)
	if len(fields) != 1 {
		return "", nil, 0, errors.New("metric value shape failed")
	}
	value, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return "", nil, 0, errors.New("metric value failed")
	}
	return name, labels, value, nil
}

func parseMetricLabels(raw string) (map[string]string, error) {
	result := make(map[string]string, 4)
	for raw != "" {
		match := finiteMetricLabelPattern.FindStringSubmatchIndex(raw)
		if len(match) == 0 || match[0] != 0 {
			return nil, errors.New("metric label syntax failed")
		}
		name := raw[match[2]:match[3]]
		if _, duplicate := result[name]; duplicate {
			return nil, errors.New("metric label duplicate")
		}
		value, err := strconv.Unquote(raw[match[4]:match[5]])
		if err != nil {
			return nil, errors.New("metric label encoding failed")
		}
		result[name] = value
		raw = raw[match[1]:]
		if raw == "" {
			break
		}
		if raw[0] != ',' {
			return nil, errors.New("metric label separator failed")
		}
		raw = raw[1:]
	}
	return result, nil
}

func metricFamily(sampleName string) string {
	const duration = "talenro_control_http_request_duration_seconds"
	for _, suffix := range []string{"_bucket", "_sum", "_count"} {
		if sampleName == duration+suffix {
			return duration
		}
	}
	return sampleName
}

func validateMetricLabels(family, sampleName string, labels map[string]string) error {
	var allowed map[string]map[string]struct{}
	switch family {
	case "talenro_control_http_requests_total", "talenro_control_http_request_duration_seconds":
		allowed = map[string]map[string]struct{}{
			"method":       stringSet("GET", "other"),
			"route":        stringSet("/livez", "/readyz", "unmatched"),
			"status_class": stringSet("1xx", "2xx", "3xx", "4xx", "5xx"),
		}
		if strings.HasSuffix(sampleName, "_bucket") {
			allowed["le"] = stringSet("0.005", "0.01", "0.025", "0.05", "0.1", "0.25", "0.5", "1", "2.5", "5", "10", "+Inf")
		}
	case "talenro_control_build_info":
		info := buildinfo.Current()
		allowed = map[string]map[string]struct{}{"version": stringSet(info.Version), "commit": stringSet(info.Commit)}
	case "talenro_security_events_total":
		allowed = map[string]map[string]struct{}{
			"operation": stringSet("account_registration", "account_authentication", "account_recovery", "device_enrollment", "device_authentication", "device_revocation", "bundle_resolution", "bundle_acknowledgement"),
			"result":    stringSet("success", "failure"),
			"reason":    stringSet("none", "invalid_credential", "expired", "revoked", "rate_limited", "dependency", "internal", "replay"),
		}
	case "talenro_outbox_backlog", "talenro_outbox_oldest_seconds":
		allowed = map[string]map[string]struct{}{}
	case "talenro_error_reports_total":
		allowed = map[string]map[string]struct{}{
			"component": stringSet("control_api", "identity", "deviceauth", "trust", "outbox", "email", "reporter"),
			"result":    stringSet("sent", "dropped", "provider_failed"),
		}
	case "talenro_crypto_validations_total":
		allowed = map[string]map[string]struct{}{
			"operation": stringSet("lookup", "encrypt", "decrypt", "token_verify", "proof_verify", "metadata_verify", "bundle_verify", "bundle_decrypt", "bundle_sign"),
			"result":    stringSet("success", "failure"),
			"reason":    stringSet("none", "malformed", "invalid", "expired", "revoked", "rollback", "key_unavailable", "wrong_domain"),
		}
	default:
		return errors.New("metric family failed")
	}
	if len(labels) != len(allowed) {
		return errors.New("metric label name set failed")
	}
	for name, value := range labels {
		values, exists := allowed[name]
		if _, finite := values[value]; !exists || !finite {
			return errors.New("metric label value failed")
		}
	}
	return nil
}

func stringSet(values ...string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		result[value] = struct{}{}
	}
	return result
}

func requirePowerShellContract(t *testing.T) {
	t.Helper()
	if runtime.GOOS != "windows" {
		t.Skip("verify-c11 PowerShell contract is isolated to Windows")
	}
	repoRoot := e2eRepositoryRoot(t)
	script := filepath.Join(repoRoot, "scripts", "verify-c11.ps1")
	if info, err := os.Stat(script); err != nil || info.IsDir() {
		t.Fatal("verify-c11 PowerShell entrypoint is missing")
	}
	powerShell, err := exec.LookPath("powershell")
	if err != nil {
		t.Fatal("verify-c11 PowerShell executable is missing")
	}
	fakeDirectory := t.TempDir()
	logPath := filepath.Join(fakeDirectory, "calls.log")
	projectPath := filepath.Join(fakeDirectory, "project.txt")
	writeFakeTool(t, filepath.Join(fakeDirectory, "go.cmd"), fakeGoBatch(logPath))
	writeFakeTool(t, filepath.Join(fakeDirectory, "git.cmd"), fakeGitBatch(logPath))
	writeFakeTool(t, filepath.Join(fakeDirectory, "docker.cmd"), fakeDockerBatch(logPath, projectPath))
	writeFakeTool(t, filepath.Join(fakeDirectory, "gofmt.cmd"), fakeGofmtBatch(logPath))
	commandContext, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	command := exec.CommandContext(commandContext, powerShell, "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", script) //nolint:gosec // Fixed reviewed script path.
	command.WaitDelay = 5 * time.Second
	command.Dir = filepath.Dir(repoRoot)
	command.Env = contractEnvironment(os.Environ(), fakeDirectory, logPath, projectPath)
	outputCapture := newBoundedCommandCapture()
	command.Stdout = outputCapture
	command.Stderr = outputCapture
	runErr := command.Run()
	if commandContext.Err() != nil || outputCapture.overflowed() {
		t.Fatal("verify-c11 PowerShell contract exceeded its hard bound")
	}
	output := outputCapture.bytes()
	var exitError *exec.ExitError
	if !errors.As(runErr, &exitError) || exitError.ExitCode() != 73 {
		actual := 0
		if runErr != nil {
			actual = -1
		}
		if exitError != nil {
			actual = exitError.ExitCode()
		}
		transcript, _ := os.ReadFile(logPath) //nolint:gosec // Exact test-owned diagnostic path.
		t.Fatalf("verify-c11 PowerShell did not preserve the injected e2e exit code: got %d, stages=%q, commands=%q", actual,
			observedVerifyStages(output), observedCommandClasses(string(transcript)))
	}
	assertVerifyScriptResult(t, "PowerShell", output, logPath)
}

func requireBashContract(t *testing.T) {
	t.Helper()
	repoRoot := e2eRepositoryRoot(t)
	script := filepath.Join(repoRoot, "scripts", "verify-c11.sh")
	if info, err := os.Stat(script); err != nil || info.IsDir() {
		t.Fatal("verify-c11 Bash entrypoint is missing")
	}
	bash := findContractBash()
	if bash == "" {
		t.Fatal("verify-c11 Bash executable is missing")
	}
	fakeDirectory := t.TempDir()
	logPath := filepath.Join(fakeDirectory, "calls.log")
	projectPath := filepath.Join(fakeDirectory, "project.txt")
	writeFakeTool(t, filepath.Join(fakeDirectory, "go"), fakeGoShell())
	writeFakeTool(t, filepath.Join(fakeDirectory, "git"), fakeGitShell())
	writeFakeTool(t, filepath.Join(fakeDirectory, "docker"), fakeDockerShell())
	writeFakeTool(t, filepath.Join(fakeDirectory, "gofmt"), fakeGofmtShell())
	commandContext, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	scriptArgument := script
	bashLogPath := logPath
	bashProjectPath := projectPath
	if runtime.GOOS == "windows" {
		scriptArgument = gitBashPath(script)
		bashLogPath = gitBashPath(logPath)
		bashProjectPath = gitBashPath(projectPath)
	}
	command := exec.CommandContext(commandContext, bash, scriptArgument) //nolint:gosec // Fixed reviewed Bash and script paths.
	command.WaitDelay = 5 * time.Second
	command.Dir = repoRoot
	command.Env = contractEnvironment(os.Environ(), fakeDirectory, bashLogPath, bashProjectPath)
	outputCapture := newBoundedCommandCapture()
	command.Stdout = outputCapture
	command.Stderr = outputCapture
	runErr := command.Run()
	if commandContext.Err() != nil || outputCapture.overflowed() {
		transcript, _ := os.ReadFile(logPath) //nolint:gosec // Exact test-owned diagnostic path.
		t.Fatalf("verify-c11 Bash contract exceeded its hard bound; stages=%q, summary=%+v",
			observedVerifyStages(outputCapture.bytes()), summarizeCommandTranscript(string(transcript)))
	}
	var exitError *exec.ExitError
	if !errors.As(runErr, &exitError) || exitError.ExitCode() != 73 {
		t.Fatal("verify-c11 Bash did not preserve the injected e2e exit code")
	}
	assertVerifyScriptResult(t, "Bash", outputCapture.bytes(), logPath)
}

func requireScriptCleanupExitStatus(t *testing.T, shell, mode, scriptPath, powerShellHarness, bashHarness,
	ownedDirectory string, primaryExit, downExit, preLocalExit, postLocalExit, want int,
) {
	t.Helper()
	commandContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	var command *exec.Cmd
	switch shell {
	case "powershell":
		if runtime.GOOS != "windows" {
			t.Skip("cleanup PowerShell contract is isolated to Windows")
		}
		powerShell, err := exec.LookPath("powershell")
		if err != nil {
			t.Fatal("cleanup PowerShell contract executable is missing")
		}
		//nolint:gosec // Fixed reviewed harness and test-owned arguments.
		command = exec.CommandContext(commandContext, powerShell, "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", powerShellHarness,
			"-ScriptPath", scriptPath, "-Mode", mode, "-PrimaryExit", strconv.Itoa(primaryExit),
			"-DownExit", strconv.Itoa(downExit), "-PreLocalExit", strconv.Itoa(preLocalExit),
			"-PostLocalExit", strconv.Itoa(postLocalExit), "-OwnedDirectory", ownedDirectory)
	case "bash":
		bash := findContractBash()
		if bash == "" {
			t.Fatal("cleanup Bash contract executable is missing")
		}
		harnessArgument := bashHarness
		scriptArgument := scriptPath
		ownedArgument := ownedDirectory
		if runtime.GOOS == "windows" {
			harnessArgument = gitBashPath(harnessArgument)
			scriptArgument = gitBashPath(scriptArgument)
			ownedArgument = gitBashPath(ownedArgument)
		}
		//nolint:gosec // Fixed reviewed harness and test-owned arguments.
		command = exec.CommandContext(commandContext, bash, harnessArgument, scriptArgument, mode,
			strconv.Itoa(primaryExit), strconv.Itoa(downExit), strconv.Itoa(preLocalExit), strconv.Itoa(postLocalExit), ownedArgument)
	default:
		t.Fatal("unknown cleanup contract shell")
	}
	command.Dir = filepath.Dir(scriptPath)
	command.WaitDelay = 2 * time.Second
	outputCapture := newBoundedCommandCapture()
	command.Stdout = outputCapture
	command.Stderr = outputCapture
	runErr := command.Run()
	if commandContext.Err() != nil || outputCapture.overflowed() {
		t.Fatal("cleanup status contract exceeded its hard bound")
	}
	actual := commandExitStatus(runErr)
	if actual != want {
		t.Fatalf("%s %s cleanup status mismatch: got %d, want %d", shell, mode, actual, want)
	}
	if bytes.Contains(outputCapture.bytes(), []byte("TASK19-CLEANUP-CANARY")) {
		t.Fatal("cleanup status contract disclosed an exception canary")
	}
}

func commandExitStatus(runErr error) int {
	if runErr == nil {
		return 0
	}
	var exitError *exec.ExitError
	if errors.As(runErr, &exitError) {
		return exitError.ExitCode()
	}
	return -1
}

func cleanupContractPowerShell() string {
	return `param(
  [Parameter(Mandatory)][string]$ScriptPath,
  [Parameter(Mandatory)][ValidateSet('verify', 'smoke')][string]$Mode,
  [Parameter(Mandatory)][int]$PrimaryExit,
  [Parameter(Mandatory)][int]$DownExit,
  [Parameter(Mandatory)][int]$PreLocalExit,
  [Parameter(Mandatory)][int]$PostLocalExit,
  [Parameter(Mandatory)][string]$OwnedDirectory
)
$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

function Get-Definitions {
  param([string[]]$Names)
  $tokens = $null
  $errors = $null
  $ast = [System.Management.Automation.Language.Parser]::ParseFile($ScriptPath, [ref]$tokens, [ref]$errors)
  if ($errors.Count -ne 0) { throw 'source parse failed' }
  $definitions = @($ast.FindAll({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] }, $true))
  $result = [System.Text.StringBuilder]::new()
  foreach ($name in $Names) {
    $match = @($definitions | Where-Object Name -eq $name)
    if ($match.Count -ne 1) { throw 'function resolution failed' }
    $null = $result.AppendLine($match[0].Extent.Text)
  }
  return $result.ToString()
}

function New-Failure {
  param([int]$ExitCode, [string]$Stage)
  $failure = [System.Exception]::new('TASK19-CLEANUP-CANARY')
  $failure.Data['ExitCode'] = $ExitCode
  $failure.Data['Stage'] = $Stage
  return $failure
}

function Get-FinalExit {
  param($PrimaryFailure, $CleanupFailure)
  $failure = $PrimaryFailure
  if ($null -eq $failure) { $failure = $CleanupFailure }
  if ($null -eq $failure) { return 0 }
  $candidate = [int]$failure.Data['ExitCode']
  if ($candidate -eq 0) { return 1 }
  return $candidate
}

$script:DownExit = $DownExit
if ($Mode -eq 'verify') {
  Invoke-Expression (Get-Definitions -Names @('Stop-C11Verification', 'Close-C11Dependencies', 'Remove-C11RunDirectory', 'Invoke-C11Main'))
  function Invoke-C11External {
    if ($script:DownExit -ne 0) { throw (New-Failure -ExitCode $script:DownExit -Stage 'verify-c11: dependency cleanup') }
  }
  $tokens = $null
  $errors = $null
  $ast = [System.Management.Automation.Language.Parser]::ParseFile($ScriptPath, [ref]$tokens, [ref]$errors)
  $main = @($ast.FindAll({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Invoke-C11Main' }, $true))[0]
  $finally = @($main.Body.FindAll({ param($node) $node -is [System.Management.Automation.Language.TryStatementAst] -and $null -ne $node.Finally -and $node.Finally.Extent.Text -match 'Close-C11Dependencies' }, $true))[0].Finally
  $body = [scriptblock]::Create($finally.Extent.Text.Substring(1, $finally.Extent.Text.Length - 2))
  $primaryFailure = if ($PrimaryExit -ne 0) { New-Failure -ExitCode $PrimaryExit -Stage 'verify-c11: primary' } else { $null }
  $cleanupFailure = if ($PreLocalExit -ne 0) { New-Failure -ExitCode 1 -Stage 'verify-c11: pre-local cleanup' } else { $null }
  $script:C11ComposeTouched = $true
  $script:C11ComposeArguments = @()
  $script:C11RunDirectory = $null
  if ($PostLocalExit -ne 0) {
    $script:C11RunDirectory = $OwnedDirectory
    $null = [System.IO.Directory]::CreateDirectory($OwnedDirectory)
  }
  $locationPushed = $false
  . $body
  exit (Get-FinalExit -PrimaryFailure $primaryFailure -CleanupFailure $cleanupFailure)
}

Invoke-Expression (Get-Definitions -Names @('Stop-Smoke', 'Get-SmokeFailure'))
function Invoke-QuietExternal {
  if ($script:DownExit -ne 0) { throw (New-Failure -ExitCode $script:DownExit -Stage 'smoke: compose down') }
}
$tokens = $null
$errors = $null
$ast = [System.Management.Automation.Language.Parser]::ParseFile($ScriptPath, [ref]$tokens, [ref]$errors)
$finally = @($ast.FindAll({ param($node) $node -is [System.Management.Automation.Language.TryStatementAst] -and $null -ne $node.Finally -and $node.Finally.Extent.Text -match '\$composeTouched' }, $true))[0].Finally
$body = [scriptblock]::Create($finally.Extent.Text.Substring(1, $finally.Extent.Text.Length - 2))
$primaryFailure = if ($PrimaryExit -ne 0) { @{ ExitCode = $PrimaryExit; Stage = 'smoke: primary' } } else { $null }
$cleanupFailure = $null
$mirrorB = $null
$mirrorA = $null
$controlAPI = $null
$artifactPaths = @()
$client = $null
if ($PreLocalExit -ne 0) {
  $client = [pscustomobject]@{}
  $client | Add-Member -MemberType ScriptMethod -Name Dispose -Value { throw [System.Exception]::new('TASK19-CLEANUP-CANARY') }
}
$composeTouched = $true
$composeArguments = @()
$buildDirectory = if ($PostLocalExit -ne 0) { $OwnedDirectory } else { $null }
$controlAPIBinary = $null
$mirrorBinary = $null
$conformanceBinary = $null
$composeOverride = $null
$natsConfig = $null
function Remove-SmokeBuildDirectory {
  if ($PostLocalExit -ne 0) { throw [System.Exception]::new('TASK19-CLEANUP-CANARY') }
}
. $body
$failure = $primaryFailure
if ($null -eq $failure) { $failure = $cleanupFailure }
if ($null -eq $failure) { exit 0 }
$candidate = [int]$failure.ExitCode
if ($candidate -eq 0) { $candidate = 1 }
exit $candidate
`
}

func cleanupContractBash() string {
	return `#!/usr/bin/env bash
set -u
PATH=/usr/bin:/bin:${PATH:-}
export PATH
script_path=$1
mode=$2
primary_exit=$3
down_exit=$4
pre_local_exit=$5
post_local_exit=$6
owned_directory=$7

extract_function() {
  local name=$1
  awk -v signature="${name}() {" '
    $0 == signature { printing = 1 }
    printing { print }
    printing && $0 == "}" { exit }
  ' "${script_path}"
}

if [[ "${mode}" == verify ]]; then
  eval "$(extract_function stop_dependencies)"
  eval "$(extract_function cleanup)"
  compose_touched=1
  compose=(cleanup-command)
  run_dir=${owned_directory}
  invoke_quiet() { return "${down_exit}"; }
  remove_run_directory() { return "${post_local_exit}"; }
else
  eval "$(extract_function cleanup)"
  compose_touched=1
  compose=(cleanup-command)
  quiet_exit=0
  quiet_output=''
  build_dir=''
  mirror_b_pid=''
  mirror_a_pid=''
  api_pid=''
  terminate_pid() { return "${pre_local_exit}"; }
  scan_artifacts() { return 0; }
  invoke_quiet() { quiet_exit=${down_exit}; quiet_output=''; }
  remove_build_directory() { return "${post_local_exit}"; }
fi

set +e
(exit "${primary_exit}")
cleanup
`
}

func writeFakeTool(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o700); err != nil { //nolint:gosec // Executable fake lives only below t.TempDir.
		t.Fatal("verify-c11 fake tool creation failed")
	}
}

func fakeGoBatch(logPath string) string {
	body := `@echo off
setlocal EnableDelayedExpansion
set "TASK19_FAKE_LOG=__LOG_PATH__"
echo go %* CGO_ENABLED=%CGO_ENABLED% >>"%TASK19_FAKE_LOG%"
if defined C11_UNOWNED_PRIVATE echo ambient-leak >>"%TASK19_FAKE_LOG%"
if defined TALENRO_UNOWNED_PRIVATE echo ambient-leak >>"%TASK19_FAKE_LOG%"
if defined COMPOSE_PROJECT_NAME echo ambient-leak >>"%TASK19_FAKE_LOG%"
if "%1 %2 %3"=="tool buf --version" (echo 1.72.0& exit /b 0)
if "%1 %2 %3"=="tool protoc-gen-go --version" (echo protoc-gen-go v1.36.11& exit /b 0)
if "%1 %2 %3"=="tool oapi-codegen --version" (echo v2.8.0& exit /b 0)
if "%1 %2 %3"=="tool sqlc version" (echo v1.31.1& exit /b 0)
if "%1 %2 %3"=="tool goose -version" (echo goose version: v3.27.1& exit /b 0)
if "%1 %2 %3"=="tool golangci-lint version" (echo golangci-lint has version 2.12.2 built with go1.26.5& exit /b 0)
if defined C11_CONFORMANCE_BINARY (
    echo %*| findstr /C:"-tags=e2e" >nul
    if errorlevel 1 echo conformance-env-leak >>"%TASK19_FAKE_LOG%"
)
if "%1"=="build" (
    if not "%2"=="-o" (echo conformance-build-invalid >>"%TASK19_FAKE_LOG%"& exit /b 72)
    if not "%~4"=="./cmd/trust-conformance" (echo conformance-build-invalid >>"%TASK19_FAKE_LOG%"& exit /b 72)
    >"%~3" echo task19 fake conformance
    echo conformance-build-ok >>"%TASK19_FAKE_LOG%"
    exit /b 0
)
echo %*| findstr /C:"-tags=e2e" >nul
if not errorlevel 1 (
    set "exact=exact4-ok"
    if not defined C11_E2E_COMPOSE_PROJECT set "exact=exact4-missing"
    if not defined C11_E2E_DATABASE_URL set "exact=exact4-missing"
    if not defined C11_E2E_REDIS_ADDRESS set "exact=exact4-missing"
    if not defined C11_E2E_NATS_URL set "exact=exact4-missing"
    echo !exact! >>"%TASK19_FAKE_LOG%"
    if defined C11_E2E_EXTERNAL_DEPENDENCIES echo legacy-marker >>"%TASK19_FAKE_LOG%"
    set "conformance=conformance-ok"
    if not defined C11_CONFORMANCE_BINARY set "conformance=conformance-missing"
    if defined C11_CONFORMANCE_BINARY if not exist "!C11_CONFORMANCE_BINARY!" set "conformance=conformance-missing"
    if defined C11_CONFORMANCE_BINARY echo !C11_CONFORMANCE_BINARY!| findstr /C:"talenro-verify-c11-" >nul
    if defined C11_CONFORMANCE_BINARY if errorlevel 1 set "conformance=conformance-missing"
    echo !conformance! >>"%TASK19_FAKE_LOG%"
    setlocal DisableDelayedExpansion
__PRIVACY_OUTPUT__
    endlocal
    exit /b 73
)
exit /b 0
`
	body = strings.ReplaceAll(body, "__LOG_PATH__", batchLiteralPath(logPath))
	return strings.ReplaceAll(body, "__PRIVACY_OUTPUT__", privacyFakeOutputBatch())
}

func fakeGitBatch(logPath string) string {
	body := `@echo off
set "TASK19_FAKE_LOG=__LOG_PATH__"
echo git %* >>"%TASK19_FAKE_LOG%"
exit /b 0
`
	return strings.ReplaceAll(body, "__LOG_PATH__", batchLiteralPath(logPath))
}

func fakeDockerBatch(logPath, projectPath string) string {
	body := `@echo off
setlocal EnableDelayedExpansion
set "TASK19_FAKE_LOG=__LOG_PATH__"
set "TASK19_FAKE_PROJECT=__PROJECT_PATH__"
echo docker %* >>"%TASK19_FAKE_LOG%"
if "%1"=="compose" (
  if "%2"=="--project-name" (
    >"%TASK19_FAKE_PROJECT%" echo %3
    if "%6"=="ps" (echo 0123456789abcdef0123456789abcdef& exit /b 0)
    if "%6"=="port" (
      if "%7"=="postgres" (echo 127.0.0.1:15432& exit /b 0)
      if "%7"=="redis" (echo 127.0.0.1:16379& exit /b 0)
      if "%7"=="nats" (echo 127.0.0.1:14222& exit /b 0)
    )
  )
  exit /b 0
)
if "%1"=="inspect" (
  set /p project=<"%TASK19_FAKE_PROJECT%"
  echo !project!
  exit /b 0
)
exit /b 0
`
	body = strings.ReplaceAll(body, "__LOG_PATH__", batchLiteralPath(logPath))
	return strings.ReplaceAll(body, "__PROJECT_PATH__", batchLiteralPath(projectPath))
}

func fakeGofmtBatch(logPath string) string {
	return "@echo off\r\nset \"TASK19_FAKE_LOG=" + batchLiteralPath(logPath) + "\"\r\necho gofmt %* >>\"%TASK19_FAKE_LOG%\"\r\nexit /b 0\r\n"
}

func batchLiteralPath(path string) string {
	return strings.ReplaceAll(path, "%", "%%")
}

func fakeGoShell() string {
	body := `#!/usr/bin/env bash
set -eu
printf 'go %s CGO_ENABLED=%s\n' "$*" "${CGO_ENABLED:-missing}" >>"${TASK19_FAKE_LOG}"
[[ -z "${C11_UNOWNED_PRIVATE:-}" && -z "${TALENRO_UNOWNED_PRIVATE:-}" && -z "${COMPOSE_PROJECT_NAME:-}" ]] || printf '%s\n' ambient-leak >>"${TASK19_FAKE_LOG}"
case "$*" in
  'tool buf --version') printf '%s\n' '1.72.0'; exit 0 ;;
  'tool protoc-gen-go --version') printf '%s\n' 'protoc-gen-go v1.36.11'; exit 0 ;;
  'tool oapi-codegen --version') printf '%s\n' 'v2.8.0'; exit 0 ;;
  'tool sqlc version') printf '%s\n' 'v1.31.1'; exit 0 ;;
  'tool goose -version') printf '%s\n' 'goose version: v3.27.1'; exit 0 ;;
  'tool golangci-lint version') printf '%s\n' 'golangci-lint has version 2.12.2 built with go1.26.5'; exit 0 ;;
esac
if [[ -n "${C11_CONFORMANCE_BINARY:-}" ]]; then
  if [[ "${1:-}" != test || " $* " != *' -tags=e2e '* ]]; then
    printf '%s\n' conformance-env-leak >>"${TASK19_FAKE_LOG}"
  fi
fi
if [[ "${1:-}" == build ]]; then
  if [[ "${2:-}" != -o || "${4:-}" != ./cmd/trust-conformance ]]; then
    printf '%s\n' conformance-build-invalid >>"${TASK19_FAKE_LOG}"
    exit 72
  fi
  printf '%s\n' 'task19 fake conformance' >"${3}"
  chmod 700 "${3}"
  printf '%s\n' conformance-build-ok >>"${TASK19_FAKE_LOG}"
  exit 0
fi
if [[ "${1:-}" == test && " $* " == *' -tags=e2e '* ]]; then
  if [[ -n "${C11_E2E_COMPOSE_PROJECT:-}" && -n "${C11_E2E_DATABASE_URL:-}" &&
        -n "${C11_E2E_REDIS_ADDRESS:-}" && -n "${C11_E2E_NATS_URL:-}" ]]; then
    printf '%s\n' exact4-ok >>"${TASK19_FAKE_LOG}"
  else
    printf '%s\n' exact4-missing >>"${TASK19_FAKE_LOG}"
  fi
  [[ -z "${C11_E2E_EXTERNAL_DEPENDENCIES:-}" ]] || printf '%s\n' legacy-marker >>"${TASK19_FAKE_LOG}"
  if [[ -n "${C11_CONFORMANCE_BINARY:-}" && -x "${C11_CONFORMANCE_BINARY}" ]]; then
    case "${C11_CONFORMANCE_BINARY}" in
      *talenro-verify-c11.*|*talenro-verify-c11-*) printf '%s\n' conformance-ok >>"${TASK19_FAKE_LOG}" ;;
      *) printf '%s\n' conformance-missing >>"${TASK19_FAKE_LOG}" ;;
    esac
  else
    printf '%s\n' conformance-missing >>"${TASK19_FAKE_LOG}"
  fi
__PRIVACY_OUTPUT__
  exit 73
fi
exit 0
`
	return strings.ReplaceAll(body, "__PRIVACY_OUTPUT__", privacyFakeOutputShell())
}

func privacyFakeOutputBatch() string {
	var output strings.Builder
	for _, canary := range privacyCanaryValues(privacySurfaceCanaries()) {
		_, _ = fmt.Fprintf(&output, "    echo(%s\r\n    echo(%s 1>&2\r\n", canary, canary)
	}
	return strings.TrimSuffix(output.String(), "\r\n")
}

func privacyFakeOutputShell() string {
	var output strings.Builder
	for _, canary := range privacyCanaryValues(privacySurfaceCanaries()) {
		_, _ = fmt.Fprintf(&output, "  printf '%%s\\n' '%s'\n  printf '%%s\\n' '%s' >&2\n", canary, canary)
	}
	return strings.TrimSuffix(output.String(), "\n")
}

func fakeGitShell() string {
	return `#!/usr/bin/env bash
set -eu
printf 'git %s\n' "$*" >>"${TASK19_FAKE_LOG}"
for argument in "$@"; do
  if [[ "${argument}" == hash-object ]]; then
    printf '%040d\n' 0
    break
  fi
done
exit 0
`
}

func fakeDockerShell() string {
	return `#!/usr/bin/env bash
set -eu
printf 'docker %s\n' "$*" >>"${TASK19_FAKE_LOG}"
if [[ "${1:-}" == compose && "${2:-}" == --project-name ]]; then
  printf '%s\n' "${3}" >"${TASK19_FAKE_PROJECT}"
  case "${6:-}" in
    ps) printf '%s\n' 0123456789abcdef0123456789abcdef ;;
    port)
      case "${7:-}" in
        postgres) printf '%s\n' 127.0.0.1:15432 ;;
        redis) printf '%s\n' 127.0.0.1:16379 ;;
        nats) printf '%s\n' 127.0.0.1:14222 ;;
      esac
      ;;
  esac
  exit 0
fi
if [[ "${1:-}" == inspect ]]; then
  cat -- "${TASK19_FAKE_PROJECT}"
fi
exit 0
`
}

func fakeGofmtShell() string {
	return `#!/usr/bin/env bash
set -eu
printf 'gofmt %s\n' "$*" >>"${TASK19_FAKE_LOG}"
exit 0
`
}

func contractEnvironment(values []string, fakeDirectory, logPath, projectPath string) []string {
	result := make([]string, 0, len(values)+8)
	for _, value := range values {
		name, _, _ := strings.Cut(value, "=")
		upper := strings.ToUpper(name)
		if upper == "PATH" || upper == "CGO_ENABLED" || strings.HasPrefix(upper, "TALENRO_") || strings.HasPrefix(upper, "C11_") ||
			strings.HasPrefix(upper, "COMPOSE_") || strings.HasPrefix(upper, "TASK19_FAKE_") {
			continue
		}
		result = append(result, value)
	}
	return append(result,
		"PATH="+fakeDirectory+string(os.PathListSeparator)+os.Getenv("PATH"),
		"TASK19_FAKE_LOG="+logPath,
		"TASK19_FAKE_PROJECT="+projectPath,
		"C11_UNOWNED_PRIVATE=C11-AMBIENT-PRIVATE-19",
		"TALENRO_UNOWNED_PRIVATE=TALENRO-AMBIENT-PRIVATE-19",
		"COMPOSE_PROJECT_NAME=COMPOSE-AMBIENT-PRIVATE-19",
		"CGO_ENABLED=invalid-ambient-19",
	)
}

const fixtureCommandCaptureLimit = 128 << 10

type boundedCommandCapture struct {
	mu       sync.Mutex
	buffer   bytes.Buffer
	overflow bool
}

func newBoundedCommandCapture() *boundedCommandCapture { return &boundedCommandCapture{} }

func (capture *boundedCommandCapture) Write(value []byte) (int, error) {
	if capture == nil {
		return len(value), nil
	}
	capture.mu.Lock()
	defer capture.mu.Unlock()
	if capture.overflow || len(value) > fixtureCommandCaptureLimit-capture.buffer.Len() {
		capture.overflow = true
		return len(value), nil
	}
	_, _ = capture.buffer.Write(value)
	return len(value), nil
}

func (capture *boundedCommandCapture) bytes() []byte {
	if capture == nil {
		return nil
	}
	capture.mu.Lock()
	defer capture.mu.Unlock()
	return append([]byte(nil), capture.buffer.Bytes()...)
}

func (capture *boundedCommandCapture) overflowed() bool {
	if capture == nil {
		return true
	}
	capture.mu.Lock()
	defer capture.mu.Unlock()
	return capture.overflow
}

func assertVerifyScriptResult(t *testing.T, shell string, output []byte, logPath string) {
	t.Helper()
	forbidden := append(privacyCanaryValues(privacySurfaceCanaries()),
		"C11-FAKE-TOOL-STDOUT-PRIVATE", "C11-FAKE-TOOL-STDERR-PRIVATE",
		"C11-AMBIENT-PRIVATE-19", "TALENRO-AMBIENT-PRIVATE-19", "COMPOSE-AMBIENT-PRIVATE-19",
		"postgres://talenro:talenro_dev@127.0.0.1:",
	)
	for _, forbidden := range forbidden {
		if bytes.Contains(output, []byte(forbidden)) {
			t.Fatalf("verify-c11 %s disclosed command output", shell)
		}
	}
	if bytes.Count(output, []byte("verify-c11: e2e tests failed with exit code 73.")) != 1 {
		t.Fatalf("verify-c11 %s did not emit exactly one fixed failure stage", shell)
	}
	info, err := os.Stat(logPath)
	if err != nil || info.IsDir() || info.Size() <= 0 || info.Size() > 1<<20 {
		t.Fatalf("verify-c11 %s fake-tool transcript bound failed", shell)
	}
	logBytes, err := os.ReadFile(logPath) //nolint:gosec // Exact test-owned bounded file below t.TempDir.
	if err != nil {
		t.Fatalf("verify-c11 %s fake-tool transcript missing", shell)
	}
	transcript := string(logBytes)
	assertGoCGOModes(t, shell, transcript)
	if strings.Contains(transcript, "ambient-leak") || strings.Contains(transcript, "legacy-marker") ||
		strings.Contains(transcript, "exact4-missing") || strings.Contains(transcript, "conformance-env-leak") ||
		strings.Contains(transcript, "conformance-build-invalid") || strings.Contains(transcript, "conformance-missing") ||
		!strings.Contains(transcript, "exact4-ok") || !strings.Contains(transcript, "conformance-build-ok") ||
		!strings.Contains(transcript, "conformance-ok") {
		t.Fatalf("verify-c11 %s environment isolation contract failed; markers=%q, commands=%q", shell,
			observedContractMarkers(transcript), observedCommandClasses(transcript))
	}
	if !regexp.MustCompile(`talenro-c11-verify-[0-9a-f]{12}`).MatchString(transcript) {
		t.Fatalf("verify-c11 %s did not use an owned compose project", shell)
	}
	normalizedTranscript := strings.ReplaceAll(transcript, `"`, "")
	assertCommandOrder(t, normalizedTranscript, []string{
		"go tool buf --version", "go tool buf lint", "go tool oapi-codegen --config", "gofmt -w",
		"go test ./... -count=1", "go test -run ^$ -fuzz ^FuzzDecode$", "go test -race ./... -count=1",
		"go vet ./...", "go tool golangci-lint run ./...", "docker compose --project-name",
		"docker inspect", "go tool goose -dir", "go test -tags=integration ./... -count=1",
		"go build -o", "docker compose --project-name", "docker compose -f deploy/dev/compose.yaml config --quiet",
		"go test -tags=e2e ./internal/e2e -count=1",
		"docker compose --project-name",
	})
}

func assertGoCGOModes(t *testing.T, shell, transcript string) {
	t.Helper()
	goInvocations := 0
	raceInvocations := 0
	for _, line := range strings.Split(transcript, "\n") {
		line = strings.ReplaceAll(strings.TrimSpace(line), `"`, "")
		if !strings.HasPrefix(line, "go ") {
			continue
		}
		goInvocations++
		command, marker, found := strings.Cut(line, " CGO_ENABLED=")
		if !found || (marker != "0" && marker != "1") {
			t.Fatalf("verify-c11 %s fake Go invocation had a missing or invalid CGO marker: %q", shell, line)
		}
		if strings.Contains(command, "go test -race ./... -count=1") {
			raceInvocations++
			if marker != "1" {
				t.Fatalf("verify-c11 %s repository-wide race invocation used CGO_ENABLED=%s", shell, marker)
			}
			continue
		}
		if marker != "0" {
			t.Fatalf("verify-c11 %s ordinary Go invocation used CGO_ENABLED=%s: %q", shell, marker, command)
		}
	}
	if goInvocations == 0 {
		t.Fatalf("verify-c11 %s transcript contained no fake Go invocation", shell)
	}
	if raceInvocations != 1 {
		t.Fatalf("verify-c11 %s transcript contained %d repository-wide race invocations, want 1", shell, raceInvocations)
	}
}

func observedContractMarkers(transcript string) []string {
	markers := make([]string, 0, 9)
	for _, marker := range []string{
		"ambient-leak", "legacy-marker", "exact4-missing", "exact4-ok", "conformance-env-leak",
		"conformance-build-invalid", "conformance-build-ok", "conformance-missing", "conformance-ok",
	} {
		if strings.Contains(transcript, marker) {
			markers = append(markers, marker)
		}
	}
	return markers
}

func findContractBash() string {
	if runtime.GOOS == "windows" {
		for _, candidate := range []string{`C:\Program Files\Git\bin\bash.exe`, `C:\Program Files\Git\usr\bin\bash.exe`} {
			if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
				return candidate
			}
		}
	}
	path, _ := exec.LookPath("bash")
	return path
}

func gitBashPath(path string) string {
	cleaned := filepath.Clean(path)
	volume := filepath.VolumeName(cleaned)
	if len(volume) == 2 && volume[1] == ':' {
		return "/" + strings.ToLower(volume[:1]) + strings.ReplaceAll(cleaned[len(volume):], `\`, "/")
	}
	return filepath.ToSlash(cleaned)
}

func assertCommandOrder(t *testing.T, transcript string, commands []string) {
	t.Helper()
	if err := validateCommandOrder(transcript, commands); err != nil {
		t.Fatalf("verify-c11 omitted a required command class; observed %q", observedCommandClasses(transcript))
	}
}

func validateCommandOrder(transcript string, commands []string) error {
	offset := 0
	for _, command := range commands {
		index := strings.Index(transcript[offset:], command)
		if index < 0 {
			return errors.New("verify-c11 command order incomplete")
		}
		offset += index + len(command)
	}
	return nil
}

func observedCommandClasses(transcript string) []string {
	classes := make([]string, 0, 16)
	for _, line := range strings.Split(transcript, "\n") {
		line = strings.ReplaceAll(strings.TrimSpace(line), `"`, "")
		fields := strings.Fields(line)
		switch {
		case len(fields) >= 7 && fields[0] == "docker" && fields[1] == "compose":
			operation := fields[6]
			if operation == "ps" || operation == "port" {
				operation += " " + fields[len(fields)-1]
			}
			classes = append(classes, "docker compose "+operation)
		case len(fields) >= 2 && fields[0] == "docker" && fields[1] == "inspect":
			classes = append(classes, "docker inspect")
		case strings.HasPrefix(line, "go tool goose"):
			classes = append(classes, "go tool goose [REDACTED]")
		case strings.HasPrefix(line, "go "), strings.HasPrefix(line, "git "):
			if len(fields) > 6 {
				fields = fields[:6]
			}
			classes = append(classes, strings.Join(fields, " "))
		}
	}
	return classes
}

type commandTranscriptSummary struct {
	Total         int
	GitHashObject int
	GitLSFiles    int
	Last          string
}

func summarizeCommandTranscript(transcript string) commandTranscriptSummary {
	summary := commandTranscriptSummary{}
	for _, line := range strings.Split(transcript, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		summary.Total++
		switch {
		case strings.HasPrefix(line, "git hash-object"):
			summary.GitHashObject++
			summary.Last = "git hash-object"
		case strings.HasPrefix(line, "git ls-files"):
			summary.GitLSFiles++
			summary.Last = "git ls-files"
		case strings.HasPrefix(line, "go "):
			summary.Last = "go"
		case strings.HasPrefix(line, "docker "):
			summary.Last = "docker"
		case strings.HasPrefix(line, "gofmt "):
			summary.Last = "gofmt"
		default:
			summary.Last = "marker"
		}
	}
	return summary
}

func observedVerifyStages(output []byte) []string {
	stages := make([]string, 0, 32)
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "verify-c11:") && !strings.Contains(line, "PRIVATE") {
			stages = append(stages, line)
		}
	}
	return stages
}
