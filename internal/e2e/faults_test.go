//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"talenro.local/platform/internal/errorreport"
	"talenro.local/platform/internal/identity"
	"talenro.local/platform/internal/secret"
	"talenro.local/platform/internal/trust"
)

func TestC11DependencyAndProviderFaultMatrix(t *testing.T) {
	fixture := requireC11Fixture(t)
	fixture.runFaultMatrix(t)
}

func TestControlledFaultProvidersCountAttemptedModes(t *testing.T) {
	t.Parallel()
	email := newControlledEmailSender()
	if !email.setMode(fixtureModeReject) {
		t.Fatal("controlled email reject mode unavailable")
	}
	recovered, rejectErr := invokeFaultProvider(func() error { return email.Send(t.Context(), identity.Delivery{}) })
	if rejectErr == nil || rejectErr.Error() != "e2e email rejected" || recovered != nil {
		t.Fatal("controlled email reject mode did not return the exact fault")
	}
	if email.attemptCount(fixtureModeReject) != 1 {
		t.Fatal("controlled email did not count the reject attempt")
	}
	if !email.setMode(fixtureModePanic) {
		t.Fatal("controlled email panic mode unavailable")
	}
	recovered, panicErr := invokeFaultProvider(func() error { return email.Send(t.Context(), identity.Delivery{}) })
	if panicErr != nil || recovered != "e2e email panic" {
		t.Fatal("controlled email panic mode did not recover the exact fault")
	}
	if email.attemptCount(fixtureModePanic) != 1 {
		t.Fatal("controlled email did not count the panic attempt")
	}

	reporter := newControlledReportProvider()
	canaries := privacySurfaceCanaries()
	if !reporter.setPrivacyCanaries(canaries["provider_error"], canaries["provider_panic"]) {
		t.Fatal("controlled reporter privacy canaries unavailable")
	}
	if !reporter.setMode(fixtureModeReject) {
		t.Fatal("controlled reporter reject mode unavailable")
	}
	recovered, rejectErr = invokeFaultProvider(func() error {
		return reporter.Send(t.Context(), []errorreport.Report{fixtureFiniteReport()})
	})
	if rejectErr == nil || rejectErr.Error() != canaries["provider_error"] || recovered != nil {
		t.Fatal("controlled reporter reject mode did not return the injected fault")
	}
	if reporter.attemptCount(fixtureModeReject) != 1 {
		t.Fatal("controlled reporter did not count the reject attempt")
	}
	if !reporter.setMode(fixtureModePanic) {
		t.Fatal("controlled reporter panic mode unavailable")
	}
	recovered, panicErr = invokeFaultProvider(func() error {
		return reporter.Send(t.Context(), []errorreport.Report{fixtureFiniteReport()})
	})
	if panicErr != nil || recovered != canaries["provider_panic"] {
		t.Fatal("controlled reporter panic mode did not recover the injected fault")
	}
	if reporter.attemptCount(fixtureModePanic) != 1 {
		t.Fatal("controlled reporter did not count the panic attempt")
	}
}

func TestControlledEmailCountsRawAttemptsBeforeEffectDeduplication(t *testing.T) {
	t.Parallel()
	sender := newControlledEmailSender()
	delivery := identity.Delivery{
		DeliveryID: uuid.New(), TemplateID: identity.VerifyEmailTemplate, Locale: "en", Recipient: "raw-attempt@example.test",
		OneTimeToken: secret.NewBytes(bytes.Repeat([]byte{0x19}, 32)),
	}
	defer delivery.OneTimeToken.Clear()
	for range 2 {
		if err := sender.Send(t.Context(), delivery); err != nil {
			t.Fatal("controlled email success fixture failed")
		}
	}
	if sender.rawAttemptCount(delivery.Recipient, delivery.TemplateID) != 2 {
		t.Fatal("controlled email did not count raw duplicate attempts")
	}
	if _, effects, ok := sender.wait(delivery.Recipient, delivery.TemplateID, time.Second); !ok || effects != 1 {
		t.Fatal("controlled email side-effect deduplication changed")
	}
}

func TestNATSRecoveryOracleRejectsEarlyAndBrokerRedeliveryFalsePasses(t *testing.T) {
	t.Parallel()
	before := natsRecoverySnapshot{}
	early := natsRecoverySnapshot{brokerDeliveries: 1, providerAttempts: 1, providerEffects: 1}
	legacyPasses := early.providerAttempts-before.providerAttempts == 1 &&
		early.providerEffects-before.providerEffects == 1
	if !legacyPasses {
		t.Fatal("legacy downstream NATS oracle no longer demonstrates the false-pass precondition")
	}
	if validNATSRecoveryObservation(before, early, fixtureNATSAckWait-time.Millisecond) {
		t.Fatal("NATS recovery oracle accepted success before the broker redelivery window")
	}

	redelivered := early
	redelivered.brokerDeliveries++
	if !legacyPasses || validNATSRecoveryObservation(before, redelivered, fixtureNATSAckWait+fixtureNATSSettle) {
		t.Fatal("NATS recovery oracle accepted a broker redelivery hidden by downstream deduplication")
	}
	if !validNATSRecoveryObservation(before, early, fixtureNATSAckWait+fixtureNATSSettle) {
		t.Fatal("NATS recovery oracle rejected one acknowledged delivery and one provider effect")
	}
	ackFailed := early
	ackFailed.ackFailures++
	if validNATSRecoveryObservation(before, ackFailed, fixtureNATSAckWait+fixtureNATSSettle) {
		t.Fatal("NATS recovery oracle accepted a failed broker acknowledgement")
	}
	if fixtureNATSAckWait <= fixtureEmailProviderTimeout || fixtureNATSAckWait > 10*time.Second {
		t.Fatal("NATS AckWait is not a finite bound above the provider timeout")
	}
}

type failingFixtureAcker struct {
	calls int
}

func (acker *failingFixtureAcker) AckSync(...nats.AckOpt) error {
	acker.calls++
	return errors.New("fixture ack rejected")
}

func TestFixtureDeliveryAckFailureCannotPassSilently(t *testing.T) {
	t.Parallel()
	acker := new(failingFixtureAcker)
	var failures atomic.Int32
	if acknowledgeFixtureDelivery(t.Context(), acker, &failures) || acker.calls != 1 || failures.Load() != 1 {
		t.Fatal("fixture delivery acknowledgement failure passed silently")
	}
}

func invokeFaultProvider(call func() error) (recovered any, result error) {
	if call == nil {
		return nil, errors.New("e2e provider probe rejected")
	}
	defer func() { recovered = recover() }()
	result = call()
	return nil, result
}

func TestRaceOutcomeRequiresExactlyOneWinnerAndOneReplay(t *testing.T) {
	t.Parallel()
	if !validRaceOutcome([]int{http.StatusCreated, http.StatusUnauthorized}, http.StatusCreated) ||
		!validRaceOutcome([]int{http.StatusUnauthorized, http.StatusOK}, http.StatusOK) {
		t.Fatal("valid concurrent winner/replay outcome was rejected")
	}
	for _, invalid := range [][]int{
		{http.StatusCreated, http.StatusCreated},
		{http.StatusUnauthorized, http.StatusUnauthorized},
		{http.StatusCreated},
		{http.StatusCreated, http.StatusServiceUnavailable},
	} {
		if validRaceOutcome(invalid, http.StatusCreated) {
			t.Fatal("invalid concurrent winner/replay outcome was accepted")
		}
	}
}

const (
	fixtureModeSuccess = "success"
	fixtureModeReject  = "reject"
	fixtureModeTimeout = "timeout"
	fixtureModePanic   = "panic"
	fixtureModeBlock   = "block"
)

type controlledDelivery struct {
	id        uuid.UUID
	recipient string
	template  identity.TemplateID
	token     string
}

type controlledEmailSender struct {
	mu          sync.Mutex
	mode        string
	attempts    map[string]int
	rawAttempts map[controlledDeliveryAttemptKey]int
	deliveries  []controlledDelivery
	accepted    map[uuid.UUID]struct{}
}

type controlledDeliveryAttemptKey struct {
	id        uuid.UUID
	recipient string
	template  identity.TemplateID
}

func newControlledEmailSender() *controlledEmailSender {
	return &controlledEmailSender{
		mode: fixtureModeSuccess, attempts: make(map[string]int), rawAttempts: make(map[controlledDeliveryAttemptKey]int),
		accepted: make(map[uuid.UUID]struct{}),
	}
}

func (sender *controlledEmailSender) Send(ctx context.Context, delivery identity.Delivery) error {
	if sender == nil || ctx == nil || ctx.Err() != nil {
		return errors.New("e2e email rejected")
	}
	sender.mu.Lock()
	mode := sender.mode
	sender.attempts[mode]++
	sender.rawAttempts[controlledDeliveryAttemptKey{
		id: delivery.DeliveryID, recipient: delivery.Recipient, template: delivery.TemplateID,
	}]++
	sender.mu.Unlock()
	switch mode {
	case fixtureModeReject:
		return errors.New("e2e email rejected")
	case fixtureModePanic:
		panic("e2e email panic")
	case fixtureModeTimeout, fixtureModeBlock:
		<-ctx.Done()
		return errors.New("e2e email timeout")
	case fixtureModeSuccess:
	default:
		return errors.New("e2e email rejected")
	}
	token := delivery.OneTimeToken.Copy()
	defer clear(token)
	if delivery.DeliveryID == uuid.Nil || len(token) != 32 {
		return errors.New("e2e email rejected")
	}
	sender.mu.Lock()
	defer sender.mu.Unlock()
	if _, exists := sender.accepted[delivery.DeliveryID]; exists {
		return nil
	}
	sender.accepted[delivery.DeliveryID] = struct{}{}
	sender.deliveries = append(sender.deliveries, controlledDelivery{
		id: delivery.DeliveryID, recipient: delivery.Recipient, template: delivery.TemplateID,
		token: base64.RawURLEncoding.EncodeToString(token),
	})
	return nil
}

func (sender *controlledEmailSender) attemptCount(mode string) int {
	if sender == nil {
		return 0
	}
	sender.mu.Lock()
	defer sender.mu.Unlock()
	return sender.attempts[mode]
}

func (sender *controlledEmailSender) rawAttemptCount(recipient string, template identity.TemplateID) int {
	if sender == nil {
		return 0
	}
	sender.mu.Lock()
	defer sender.mu.Unlock()
	total := 0
	for key, count := range sender.rawAttempts {
		if key.recipient == recipient && key.template == template {
			total += count
		}
	}
	return total
}

func (sender *controlledEmailSender) effectCount(recipient string, template identity.TemplateID) int {
	if sender == nil {
		return 0
	}
	sender.mu.Lock()
	defer sender.mu.Unlock()
	count := 0
	for _, delivery := range sender.deliveries {
		if delivery.recipient == recipient && delivery.template == template {
			count++
		}
	}
	return count
}

func (sender *controlledEmailSender) setMode(mode string) bool {
	switch mode {
	case fixtureModeSuccess, fixtureModeReject, fixtureModeTimeout, fixtureModePanic:
	default:
		return false
	}
	sender.mu.Lock()
	sender.mode = mode
	sender.mu.Unlock()
	return true
}

func (sender *controlledEmailSender) wait(recipient string, template identity.TemplateID, timeout time.Duration) (string, int, bool) {
	deadline := time.Now().Add(timeout)
	for {
		sender.mu.Lock()
		var value string
		count := 0
		for _, delivery := range sender.deliveries {
			if delivery.recipient == recipient && delivery.template == template {
				value = delivery.token
				count++
			}
		}
		sender.mu.Unlock()
		if count > 0 {
			return value, count, true
		}
		if !time.Now().Before(deadline) {
			return "", 0, false
		}
		time.Sleep(25 * time.Millisecond)
	}
}

type controlledConfigSigner struct {
	base     trust.ConfigSigner
	mu       sync.RWMutex
	mode     string
	attempts map[string]int
}

func newControlledConfigSigner(base trust.ConfigSigner) *controlledConfigSigner {
	return &controlledConfigSigner{base: base, mode: fixtureModeSuccess, attempts: make(map[string]int)}
}

func (signer *controlledConfigSigner) KeyID() string {
	if signer == nil || signer.base == nil {
		return ""
	}
	return signer.base.KeyID()
}

func (signer *controlledConfigSigner) Sign(ctx context.Context, message []byte) ([]byte, error) {
	if signer == nil || signer.base == nil || ctx == nil || ctx.Err() != nil {
		return nil, errors.New("e2e signer rejected")
	}
	signer.mu.Lock()
	mode := signer.mode
	signer.attempts[mode]++
	signer.mu.Unlock()
	switch mode {
	case fixtureModeSuccess:
		return signer.base.Sign(ctx, message)
	case fixtureModeReject:
		return nil, errors.New("e2e signer rejected")
	case fixtureModeTimeout, fixtureModeBlock:
		<-ctx.Done()
		return nil, errors.New("e2e signer timeout")
	case fixtureModePanic:
		panic("e2e signer panic")
	default:
		return nil, errors.New("e2e signer rejected")
	}
}

func (signer *controlledConfigSigner) attemptCount(mode string) int {
	if signer == nil {
		return 0
	}
	signer.mu.RLock()
	defer signer.mu.RUnlock()
	return signer.attempts[mode]
}

func (signer *controlledConfigSigner) setMode(mode string) bool {
	switch mode {
	case fixtureModeSuccess, fixtureModeReject, fixtureModeTimeout:
	default:
		return false
	}
	signer.mu.Lock()
	signer.mode = mode
	signer.mu.Unlock()
	return true
}

type controlledReportProvider struct {
	mu             sync.RWMutex
	mode           string
	attempts       map[string]int
	delivered      atomic.Int64
	capture        *privacyCapture
	providerCanary string
	panicCanary    string
}

func newControlledReportProvider() *controlledReportProvider {
	return &controlledReportProvider{mode: fixtureModeSuccess, attempts: make(map[string]int)}
}

func (provider *controlledReportProvider) Send(ctx context.Context, reports []errorreport.Report) error {
	if provider == nil || ctx == nil || ctx.Err() != nil || len(reports) == 0 {
		return errors.New("e2e reporter rejected")
	}
	provider.mu.Lock()
	mode := provider.mode
	provider.attempts[mode]++
	capture := provider.capture
	providerCanary := provider.providerCanary
	panicCanary := provider.panicCanary
	provider.mu.Unlock()
	if capture != nil {
		capture.recordReports(reports)
	}
	switch mode {
	case fixtureModeSuccess:
		provider.delivered.Add(int64(len(reports)))
		return nil
	case fixtureModeReject:
		if providerCanary != "" {
			if capture != nil {
				capture.recordSource(fixturePrivacyProviderSurface, providerCanary)
			}
			return errors.New(providerCanary)
		}
		return errors.New("e2e reporter rejected")
	case fixtureModePanic:
		if panicCanary != "" {
			if capture != nil {
				capture.recordSource(fixturePrivacyPanicSurface, panicCanary)
			}
			panic(panicCanary)
		}
		panic("e2e reporter panic")
	case fixtureModeBlock, fixtureModeTimeout:
		<-ctx.Done()
		return errors.New("e2e reporter timeout")
	default:
		return errors.New("e2e reporter rejected")
	}
}

func (provider *controlledReportProvider) setPrivacyCanaries(providerCanary, panicCanary string) bool {
	if provider == nil || !validPrivacyCanary(providerCanary) || !validPrivacyCanary(panicCanary) || providerCanary == panicCanary {
		return false
	}
	provider.mu.Lock()
	provider.providerCanary = providerCanary
	provider.panicCanary = panicCanary
	provider.mu.Unlock()
	return true
}

func validPrivacyCanary(value string) bool {
	return len(value) >= 8 && len(value) <= 128 && strings.TrimSpace(value) == value && !strings.ContainsAny(value, "\r\n\x00")
}

func (provider *controlledReportProvider) setMode(mode string) bool {
	switch mode {
	case fixtureModeSuccess, fixtureModeReject, fixtureModePanic, fixtureModeBlock:
	default:
		return false
	}
	provider.mu.Lock()
	provider.mode = mode
	provider.mu.Unlock()
	return true
}

func (provider *controlledReportProvider) count() int { return int(provider.delivered.Load()) }

func (provider *controlledReportProvider) attemptCount(mode string) int {
	if provider == nil {
		return 0
	}
	provider.mu.RLock()
	defer provider.mu.RUnlock()
	return provider.attempts[mode]
}

type faultFixture struct {
	account c11Account
	device  c11Device
	client  *c11HTTPClient
	context context.Context
}

type natsRecoverySnapshot struct {
	brokerDeliveries int
	providerAttempts int
	providerEffects  int
	ackFailures      int
}

func validNATSRecoveryObservation(before, after natsRecoverySnapshot, observed time.Duration) bool {
	return observed >= fixtureNATSAckWait+fixtureNATSSettle &&
		after.brokerDeliveries-before.brokerDeliveries == 1 &&
		after.providerAttempts-before.providerAttempts == 1 &&
		after.providerEffects-before.providerEffects == 1 &&
		after.ackFailures-before.ackFailures == 0
}

func snapshotNATSRecovery(
	ctx context.Context,
	t *testing.T,
	child *fixtureChild,
	recipient string,
) natsRecoverySnapshot {
	t.Helper()
	request := func(operation, template string) int {
		response, err := child.callContext(ctx, operation, "", recipient, template)
		if err != nil {
			t.Fatal("c11 NATS recovery counter unavailable")
		}
		return response.Count
	}
	return natsRecoverySnapshot{
		brokerDeliveries: request("broker-deliveries", ""),
		providerAttempts: request("delivery-attempts", string(identity.VerifyEmailTemplate)),
		providerEffects:  request("delivery-effects", string(identity.VerifyEmailTemplate)),
		ackFailures:      request("broker-ack-failures", ""),
	}
}

func (fixture *c11Fixture) runFaultMatrix(t *testing.T) {
	t.Helper()
	client := &c11HTTPClient{t: t, client: fixture.client}
	account := fixture.runAccountSecurity(t, client)
	device := fixture.runDeviceAndTrust(t, client, account)
	state := faultFixture{account: account, device: device, client: client}
	rows := []struct {
		name  string
		bound time.Duration
		run   func(*testing.T, faultFixture)
	}{
		{"postgres", 45 * time.Second, fixture.postgresFault},
		{"redis", 45 * time.Second, fixture.redisFault},
		{"nats", 45 * time.Second, fixture.natsFault},
		{"email", 30 * time.Second, fixture.emailFault},
		{"signer", 30 * time.Second, fixture.signerFault},
		{"reporter", 30 * time.Second, fixture.reporterFault},
		{"sources", 30 * time.Second, fixture.sourceFaults},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			rowContext, cancel := context.WithTimeout(t.Context(), row.bound)
			defer cancel()
			rowState := state
			rowState.context = rowContext
			rowClient := *state.client
			rowClient.context = rowContext
			rowState.client = &rowClient
			row.run(t, rowState)
			if errors.Is(rowContext.Err(), context.DeadlineExceeded) {
				t.Fatal("c11 fault row exceeded its active deadline")
			}
		})
	}
	fixture.runRaceMatrix(t, state)
}

func (fixture *c11Fixture) postgresFault(t *testing.T, state faultFixture) {
	rotation := fixture.prepareDeviceRotation(t, state.client, state.device)
	if err := fixture.compose.run(state.context, "stop", "postgres"); err != nil {
		t.Fatal("c11 postgres fault injection failed")
	}
	restored := false
	defer func() {
		if !restored {
			cleanupContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			_ = fixture.compose.run(cleanupContext, "start", "postgres")
			cancel()
		}
	}()
	waitReadinessCheck(state.context, t, fixture.client, fixture.ready.RequiredURL+"/readyz", http.StatusServiceUnavailable, "postgres", "down", 15*time.Second)
	waitHTTPStatus(state.context, t, fixture.client, fixture.ready.RequiredURL+"/livez", http.StatusOK, 5*time.Second)
	_, _ = state.client.post(state.account.baseURL, "/v1/account-sessions", "", nextFixtureKey("pg-login"), map[string]any{
		"method": "password", "email": state.account.email, "password": state.account.password,
		"client_signing_public_key": base64.RawURLEncoding.EncodeToString(state.account.signingPublic),
	}, http.StatusServiceUnavailable)
	_, _ = state.client.post(state.account.baseURL, "/v1/device-enrollment-grants", state.account.accessToken, nextFixtureKey("pg-grant"), map[string]any{
		"reauthentication": map[string]any{"method": "password", "password": state.account.password},
	}, http.StatusServiceUnavailable)
	_, _ = state.client.post(state.device.baseURL, "/v1/device-token-rotations", "", nextFixtureKey("pg-rotation"), rotation, http.StatusServiceUnavailable)
	_, _ = state.client.post(state.device.baseURL, "/v1/config-bundle-resolutions", state.device.accessToken, nextFixtureKey("pg-resolve"), map[string]any{}, http.StatusServiceUnavailable)
	_, _ = state.client.post(state.device.baseURL, "/v1/config-bundle-acknowledgements", state.device.accessToken, nextFixtureKey("pg-ack"), map[string]any{
		"bundle_id": uuid.NewString(), "bundle_version": "1",
	}, http.StatusServiceUnavailable)
	if err := fixture.compose.run(state.context, "start", "postgres"); err != nil {
		t.Fatal("c11 postgres recovery failed")
	}
	restored = true
	waitReadinessCheck(state.context, t, fixture.client, fixture.ready.RequiredURL+"/readyz", http.StatusOK, "postgres", "up", 15*time.Second)
}

func (fixture *c11Fixture) redisFault(t *testing.T, state faultFixture) {
	rotation := fixture.prepareDeviceRotation(t, state.client, state.device)
	if err := fixture.compose.run(state.context, "stop", "redis"); err != nil {
		t.Fatal("c11 redis fault injection failed")
	}
	restored := false
	defer func() {
		if !restored {
			cleanupContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			_ = fixture.compose.run(cleanupContext, "start", "redis")
			cancel()
		}
	}()
	waitReadinessCheck(state.context, t, fixture.client, fixture.ready.RequiredURL+"/readyz", http.StatusOK, "redis", "degraded", 5*time.Second)
	waitReadinessCheck(state.context, t, fixture.client, fixture.ready.RequiredURL+"/readyz", http.StatusServiceUnavailable, "redis", "down", 15*time.Second)
	_, _ = state.client.post(state.device.baseURL, "/v1/config-bundle-resolutions", state.device.accessToken, nextFixtureKey("redis-existing"), map[string]any{}, http.StatusOK)
	_, _ = state.client.post(state.account.baseURL, "/v1/account-sessions", "", nextFixtureKey("redis-login"), map[string]any{
		"method": "password", "email": state.account.email, "password": state.account.password,
		"client_signing_public_key": base64.RawURLEncoding.EncodeToString(state.account.signingPublic),
	}, http.StatusServiceUnavailable)
	_, _ = state.client.post(state.device.baseURL, "/v1/device-auth-challenges", "", nextFixtureKey("redis-challenge"), map[string]any{
		"refresh_token": state.device.refreshToken, "request_nonce": encode32(random32(t)),
	}, http.StatusServiceUnavailable)
	_, _ = state.client.post(state.device.baseURL, "/v1/device-token-rotations", "", nextFixtureKey("redis-rotation"), rotation, http.StatusServiceUnavailable)
	if err := fixture.compose.run(state.context, "start", "redis"); err != nil {
		t.Fatal("c11 redis recovery failed")
	}
	restored = true
	waitReadinessCheck(state.context, t, fixture.client, fixture.ready.RequiredURL+"/readyz", http.StatusOK, "redis", "up", 15*time.Second)
}

func (fixture *c11Fixture) natsFault(t *testing.T, state faultFixture) {
	if err := fixture.compose.run(state.context, "stop", "nats"); err != nil {
		t.Fatal("c11 nats fault injection failed")
	}
	restored := false
	defer func() {
		if !restored {
			cleanupContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			_ = fixture.compose.run(cleanupContext, "start", "nats")
			cancel()
		}
	}()
	email := fmt.Sprintf("nats-%d@example.test", time.Now().UnixNano())
	//nolint:gosec // Synthetic E2E account credential.
	_, _ = state.client.post(fixture.ready.RequiredURL, "/v1/accounts", "", nextFixtureKey("nats-register"), map[string]any{
		"email": email, "password": "Task19-NATS-Password!51", "locale": "en",
	}, http.StatusAccepted)
	waitReadinessCheck(state.context, t, fixture.client, fixture.ready.RequiredURL+"/readyz", http.StatusOK, "outbox", "degraded", 15*time.Second)
	waitReadinessCheck(state.context, t, fixture.client, fixture.ready.RequiredURL+"/readyz", http.StatusServiceUnavailable, "outbox", "down", 5*time.Second)
	before := snapshotNATSRecovery(state.context, t, fixture.child, email)
	if before.providerAttempts != 0 || before.providerEffects != 0 {
		t.Fatal("c11 nats outage delivered before broker recovery")
	}
	if err := fixture.compose.run(state.context, "start", "nats"); err != nil {
		t.Fatal("c11 nats recovery failed")
	}
	restored = true
	waitHTTPStatus(state.context, t, fixture.client, fixture.ready.RequiredURL+"/readyz", http.StatusOK, 20*time.Second)
	delivery, err := fixture.child.callContext(state.context, "delivery", "", email, string(identity.VerifyEmailTemplate))
	if err != nil || delivery.Count != 1 {
		t.Fatal("c11 nats recovery did not produce exactly one delivery effect")
	}
	observationStarted := time.Now()
	settle := time.NewTimer(fixtureNATSAckWait + fixtureNATSSettle)
	defer settle.Stop()
	select {
	case <-state.context.Done():
		t.Fatal("c11 nats recovery attempt deadline exceeded")
	case <-settle.C:
	}
	after := snapshotNATSRecovery(state.context, t, fixture.child, email)
	if !validNATSRecoveryObservation(before, after, time.Since(observationStarted)) {
		t.Fatal("c11 nats recovery did not acknowledge exactly one broker delivery and one provider effect")
	}
}

func (fixture *c11Fixture) emailFault(t *testing.T, state faultFixture) {
	for _, mode := range []string{fixtureModeReject, fixtureModePanic} {
		before := fixtureOperationCount(t, fixture.child, "email-attempts", mode, state.context)
		if _, err := fixture.child.callContext(state.context, "email-mode", mode, "", ""); err != nil {
			t.Fatal("c11 email fault injection failed")
		}
		email := fmt.Sprintf("email-%s-%d@example.test", mode, time.Now().UnixNano())
		//nolint:gosec // Synthetic E2E account credential.
		accepted, _ := state.client.post(fixture.ready.RequiredURL, "/v1/accounts", "", nextFixtureKey("email-register"), map[string]any{
			"email": email, "password": "Task19-Email-Password!52", "locale": "en",
		}, http.StatusAccepted)
		if stringField(decodeObject(t, accepted), "status") != "accepted" {
			t.Fatal("c11 email failure changed generic registration")
		}
		waitFixtureOperationIncrease(t, fixture.child, "email-attempts", mode, before, 5*time.Second, state.context)
		if _, err := fixture.child.callContext(state.context, "email-mode", fixtureModeSuccess, "", ""); err != nil {
			t.Fatal("c11 email recovery failed")
		}
		delivery, err := fixture.child.callContext(state.context, "delivery", "", email, string(identity.VerifyEmailTemplate))
		if err != nil || delivery.Count != 1 {
			t.Fatal("c11 email retry was not provider-idempotent")
		}
	}
	waitHTTPStatus(state.context, t, fixture.client, fixture.ready.RequiredURL+"/readyz", http.StatusOK, 5*time.Second)
}

func (fixture *c11Fixture) signerFault(t *testing.T, state faultFixture) {
	locations, _ := state.device.resolution["locations"].([]any)
	if len(locations) != 3 {
		t.Fatal("c11 signer precondition missing immutable bundle")
	}
	before := fixtureOperationCount(t, fixture.child, "signer-attempts", fixtureModeTimeout, state.context)
	if _, err := fixture.child.callContext(state.context, "signer-mode", fixtureModeTimeout, "", ""); err != nil {
		t.Fatal("c11 signer fault injection failed")
	}
	if _, err := fixture.child.callContext(state.context, "config-sequence", "2", "", ""); err != nil {
		t.Fatal("c11 signer sequence injection failed")
	}
	_, _ = state.client.post(state.device.baseURL, "/v1/config-bundle-resolutions", state.device.accessToken, nextFixtureKey("signer-new"), map[string]any{}, http.StatusServiceUnavailable)
	waitFixtureOperationIncrease(t, fixture.child, "signer-attempts", fixtureModeTimeout, before, 5*time.Second, state.context)
	for _, location := range locations {
		_, _ = state.client.get(fmt.Sprint(location), http.StatusOK)
	}
	if _, err := fixture.child.callContext(state.context, "signer-mode", fixtureModeSuccess, "", ""); err != nil {
		t.Fatal("c11 signer recovery failed")
	}
	_, _ = state.client.post(state.device.baseURL, "/v1/config-bundle-resolutions", state.device.accessToken, nextFixtureKey("signer-recovery"), map[string]any{}, http.StatusOK)
}

func (fixture *c11Fixture) reporterFault(t *testing.T, state faultFixture) {
	before := droppedReportMetric(state.context, t, fixture, 5*time.Second)
	fullBefore := fixtureOperationCount(t, fixture.child, "report-full-count", "", state.context)
	for _, mode := range []string{fixtureModeReject, fixtureModePanic, fixtureModeBlock} {
		attemptsBefore := fixtureOperationCount(t, fixture.child, "reporter-attempts", mode, state.context)
		if _, err := fixture.child.callContext(state.context, "reporter-mode", mode, "", ""); err != nil {
			t.Fatal("c11 reporter fault injection failed")
		}
		attempts := 4
		if mode == fixtureModeBlock {
			attempts = 32
		}
		for range attempts {
			if _, err := fixture.child.callContext(state.context, "report", "", "", ""); err != nil {
				t.Fatal("c11 reporter enqueue failed")
			}
		}
		waitFixtureOperationIncrease(t, fixture.child, "reporter-attempts", mode, attemptsBefore, 5*time.Second, state.context)
		waitHTTPStatus(state.context, t, fixture.client, fixture.ready.RequiredURL+"/livez", http.StatusOK, 5*time.Second)
		waitHTTPStatus(state.context, t, fixture.client, fixture.ready.RequiredURL+"/readyz", http.StatusOK, 5*time.Second)
		_, _ = state.client.post(state.device.baseURL, "/v1/config-bundle-resolutions", state.device.accessToken, nextFixtureKey("reporter-business"), map[string]any{}, http.StatusOK)
	}
	after := waitDroppedReportIncrease(state.context, t, fixture, before, 5*time.Second)
	if after <= before {
		t.Fatal("c11 reporter full queue did not increment finite drop metric")
	}
	if _, err := fixture.child.callContext(state.context, "reporter-mode", fixtureModeSuccess, "", ""); err != nil {
		t.Fatal("c11 reporter recovery failed")
	}
	if fullAfter := fixtureOperationCount(t, fixture.child, "report-full-count", "", state.context); fullAfter <= fullBefore {
		t.Fatal("c11 reporter full queue was not actually exercised")
	}
}

func fixtureOperationCount(t *testing.T, child *fixtureChild, operation, mode string, parents ...context.Context) int {
	t.Helper()
	parent := context.Background()
	if len(parents) != 0 && parents[0] != nil {
		parent = parents[0]
	}
	response, err := child.callContext(parent, operation, mode, "", "")
	if err != nil {
		t.Fatal("c11 provider counter unavailable")
	}
	return response.Count
}

func waitFixtureOperationIncrease(
	t *testing.T,
	child *fixtureChild,
	operation, mode string,
	before int,
	bound time.Duration,
	parents ...context.Context,
) {
	t.Helper()
	parent := context.Background()
	if len(parents) != 0 && parents[0] != nil {
		parent = parents[0]
	}
	deadline := time.Now().Add(bound)
	for {
		if fixtureOperationCount(t, child, operation, mode, parent) > before {
			return
		}
		if parent.Err() != nil {
			t.Fatal("c11 provider row deadline exceeded")
		}
		if !time.Now().Before(deadline) {
			t.Fatal("c11 provider fault was not invoked")
		}
		select {
		case <-time.After(25 * time.Millisecond):
		case <-parent.Done():
			t.Fatal("c11 provider row deadline exceeded")
		}
	}
}

func droppedReportMetric(parent context.Context, t *testing.T, fixture *c11Fixture, bound time.Duration) float64 {
	t.Helper()
	deadline := time.Now().Add(bound)
	for {
		// The c11 client stores and applies parent to every request it creates.
		body, _ := (&c11HTTPClient{t: t, client: fixture.client, context: parent}).get(fixture.ready.MetricsURL+"/metrics", http.StatusOK) //nolint:contextcheck
		for _, line := range strings.Split(string(body), "\n") {
			if !strings.HasPrefix(line, "talenro_error_reports_total{") || !strings.Contains(line, `result="dropped"`) {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) == 2 {
				value, err := strconv.ParseFloat(fields[1], 64)
				if err == nil {
					return value
				}
			}
		}
		if !time.Now().Before(deadline) {
			return 0
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func waitDroppedReportIncrease(parent context.Context, t *testing.T, fixture *c11Fixture, before float64, bound time.Duration) float64 {
	t.Helper()
	deadline := time.Now().Add(bound)
	for {
		value := droppedReportMetric(parent, t, fixture, 100*time.Millisecond)
		if value > before || !time.Now().Before(deadline) {
			return value
		}
	}
}

func (fixture *c11Fixture) sourceFaults(t *testing.T, state faultFixture) {
	locations, _ := state.device.resolution["locations"].([]any)
	if len(locations) != 3 {
		t.Fatal("c11 source fault precondition failed")
	}
	bodies := make([][]byte, 0, 3)
	for _, location := range locations {
		body, _ := state.client.get(fmt.Sprint(location), http.StatusOK)
		bodies = append(bodies, body)
	}
	if string(bodies[0]) != string(bodies[1]) || string(bodies[0]) != string(bodies[2]) {
		t.Fatal("c11 source bytes diverged before injection")
	}
	names := []string{"primary", "mirror-a", "mirror-b"}
	for failedIndex, name := range names {
		down, err := fixture.child.callContext(state.context, "source-mode", name, "", "down")
		if err != nil || down.Count == 0 {
			t.Fatal("c11 source fault injection failed")
		}
		requireHTTPTransportFailure(t, fixture.client, fmt.Sprint(locations[failedIndex]), 5*time.Second)
		for index, location := range locations {
			if index == failedIndex {
				continue
			}
			body, _ := state.client.get(fmt.Sprint(location), http.StatusOK)
			if string(body) != string(bodies[0]) {
				t.Fatal("c11 surviving source regenerated immutable bytes")
			}
		}
		up, err := fixture.child.callContext(state.context, "source-mode", name, "", "up")
		if err != nil || up.Count <= down.Count {
			t.Fatal("c11 source recovery failed")
		}
		waitHTTPStatus(state.context, t, fixture.client, fmt.Sprint(locations[failedIndex]), http.StatusOK, 5*time.Second)
		recovered, _ := state.client.get(fmt.Sprint(locations[failedIndex]), http.StatusOK)
		if string(recovered) != string(bodies[0]) {
			t.Fatal("c11 recovered source regenerated immutable bytes")
		}
	}
}

func requireHTTPTransportFailure(t *testing.T, client *http.Client, rawURL string, bound time.Duration) {
	t.Helper()
	deadline := time.Now().Add(bound)
	for {
		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
		if err != nil {
			cancel()
			t.Fatal("c11 source failure request construction failed")
		}
		response, requestErr := client.Do(request)
		if response != nil {
			_ = response.Body.Close()
		}
		cancel()
		if requestErr != nil {
			return
		}
		if !time.Now().Before(deadline) {
			t.Fatal("c11 source listener remained reachable after stop")
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func (fixture *c11Fixture) runRaceMatrix(t *testing.T, state faultFixture) {
	t.Helper()
	const iterations = 100
	t.Run("generic-idempotency", func(t *testing.T) {
		for iteration := range iterations {
			key := nextFixtureKey("race-idempotency")
			body := map[string]any{"reauthentication": map[string]any{"method": "password", "password": state.account.password}}
			attempt := c11RacePOST{baseURL: state.account.baseURL, path: "/v1/device-enrollment-grants", bearer: state.account.accessToken, key: key, body: body}
			results := concurrentC11Posts(t, fixture.client, [2]c11RacePOST{attempt, attempt})
			if results[0].status != http.StatusCreated || results[1].status != http.StatusCreated || !bytes.Equal(results[0].body, results[1].body) {
				t.Fatalf("c11 generic idempotency race %d failed", iteration)
			}
			clear(results[0].body)
			clear(results[1].body)
		}
	})
	t.Run("grant-consume", func(t *testing.T) {
		for iteration := range iterations {
			grant := createRaceEnrollmentGrant(t, state.client, state.account)
			left := fixture.prepareDeviceRegistration(t, state.client, state.account.baseURL, grant)
			right := fixture.prepareDeviceRegistration(t, state.client, state.account.baseURL, grant)
			results := concurrentC11Posts(t, fixture.client, [2]c11RacePOST{
				{baseURL: state.account.baseURL, path: "/v1/devices", key: nextFixtureKey("grant-race-left"), body: left.request},
				{baseURL: state.account.baseURL, path: "/v1/devices", key: nextFixtureKey("grant-race-right"), body: right.request},
			})
			clear(left.signingKey)
			clear(right.signingKey)
			statuses := []int{results[0].status, results[1].status}
			clear(results[0].body)
			clear(results[1].body)
			if !validRaceOutcome(statuses, http.StatusCreated) {
				t.Fatalf("c11 grant-consume race %d failed", iteration)
			}
		}
	})
	t.Run("account-refresh", func(t *testing.T) {
		for iteration := range iterations {
			account := freshRaceAccountSession(t, state.client, state.account)
			left := prepareAccountRotation(t, state.client, account)
			right := prepareAccountRotation(t, state.client, account)
			results := concurrentC11Posts(t, fixture.client, [2]c11RacePOST{
				{baseURL: account.baseURL, path: "/v1/account-token-rotations", key: nextFixtureKey("account-race-left"), body: left},
				{baseURL: account.baseURL, path: "/v1/account-token-rotations", key: nextFixtureKey("account-race-right"), body: right},
			})
			statuses := []int{results[0].status, results[1].status}
			clear(results[0].body)
			clear(results[1].body)
			if !validRaceOutcome(statuses, http.StatusOK) {
				t.Fatalf("c11 account refresh race %d failed", iteration)
			}
		}
	})
	t.Run("device-refresh", func(t *testing.T) {
		for iteration := range iterations {
			grant := createRaceEnrollmentGrant(t, state.client, state.account)
			prepared := fixture.prepareDeviceRegistration(t, state.client, state.account.baseURL, grant)
			device := completeRaceDeviceRegistration(t, state.client, state.account.baseURL, prepared)
			left := fixture.prepareDeviceRotation(t, state.client, device)
			right := fixture.prepareDeviceRotation(t, state.client, device)
			results := concurrentC11Posts(t, fixture.client, [2]c11RacePOST{
				{baseURL: device.baseURL, path: "/v1/device-token-rotations", key: nextFixtureKey("device-race-left"), body: left},
				{baseURL: device.baseURL, path: "/v1/device-token-rotations", key: nextFixtureKey("device-race-right"), body: right},
			})
			clear(device.signingKey)
			statuses := []int{results[0].status, results[1].status}
			clear(results[0].body)
			clear(results[1].body)
			if !validRaceOutcome(statuses, http.StatusOK) {
				t.Fatalf("c11 device refresh race %d failed", iteration)
			}
		}
	})
}

type c11RacePOST struct {
	baseURL string
	path    string
	bearer  string
	key     string
	body    any
}

func concurrentC11Posts(t *testing.T, client *http.Client, attempts [2]c11RacePOST) [2]c11HTTPResult {
	t.Helper()
	parent := t.Context()
	started := make(chan struct{})
	results := make(chan c11HTTPResult, len(attempts))
	for _, attempt := range attempts {
		attempt := attempt
		go func() {
			<-started
			ctx, cancel := context.WithTimeout(parent, fixtureRequestTimeout)
			defer cancel()
			results <- doC11JSONPost(ctx, client, attempt.baseURL, attempt.path, attempt.bearer, attempt.key, attempt.body)
		}()
	}
	close(started)
	responses := [2]c11HTTPResult{<-results, <-results}
	for index := range responses {
		if responses[index].err != nil {
			t.Fatal("c11 concurrent HTTP transport failed")
		}
	}
	return responses
}

func validRaceOutcome(statuses []int, winner int) bool {
	if len(statuses) != 2 || winner == http.StatusUnauthorized {
		return false
	}
	winners := 0
	replays := 0
	for _, status := range statuses {
		switch status {
		case winner:
			winners++
		case http.StatusUnauthorized:
			replays++
		default:
			return false
		}
	}
	return winners == 1 && replays == 1
}

func createRaceEnrollmentGrant(t *testing.T, client *c11HTTPClient, account c11Account) string {
	t.Helper()
	body, _ := client.post(account.baseURL, "/v1/device-enrollment-grants", account.accessToken, nextFixtureKey("race-grant"), map[string]any{
		"reauthentication": map[string]any{"method": "password", "password": account.password},
	}, http.StatusCreated)
	grant := stringField(decodeObject(t, body), "enrollment_grant")
	if grant == "" {
		t.Fatal("c11 race enrollment grant missing")
	}
	return grant
}

func freshRaceAccountSession(t *testing.T, client *c11HTTPClient, account c11Account) c11Account {
	t.Helper()
	body, _ := client.post(account.baseURL, "/v1/account-sessions", "", nextFixtureKey("race-login"), map[string]any{
		"method": "password", "email": account.email, "password": account.password,
		"client_signing_public_key": base64.RawURLEncoding.EncodeToString(account.signingPublic),
	}, http.StatusOK)
	return decodeAccountTokens(t, account.baseURL, account.email, account.password, account.signingPublic, account.signingKey, body)
}

func prepareAccountRotation(t *testing.T, client *c11HTTPClient, account c11Account) map[string]any {
	t.Helper()
	nonce := random32(t)
	challengeBody, _ := client.post(account.baseURL, "/v1/account-auth-challenges", "", nextFixtureKey("account-race-challenge"), map[string]any{
		"refresh_token": account.refreshToken, "request_nonce": encode32(nonce),
	}, http.StatusCreated)
	challenge := decodeObject(t, challengeBody)
	proof := identity.AccountRotationProofBytes(identity.AccountRotationProofInput{
		ProtocolVersion: "account-rotation-v1", Challenge: decode32(t, stringField(challenge, "challenge")), SessionID: account.sessionID,
		Operation: "rotate_account_token", Audience: account.baseURL, RequestNonce: nonce,
	})
	if len(proof) == 0 {
		t.Fatal("c11 account race proof failed")
	}
	signature := ed25519.Sign(account.signingKey, proof)
	clear(proof)
	defer clear(signature)
	return map[string]any{
		"refresh_token": account.refreshToken, "challenge_id": stringField(challenge, "challenge_id"),
		"request_nonce": encode32(nonce), "signature": base64.RawURLEncoding.EncodeToString(signature),
	}
}

func completeRaceDeviceRegistration(
	t *testing.T,
	client *c11HTTPClient,
	baseURL string,
	prepared preparedDeviceRegistration,
) c11Device {
	t.Helper()
	body, _ := client.post(baseURL, "/v1/devices", "", nextFixtureKey("race-device-register"), prepared.request, http.StatusCreated)
	object := decodeObject(t, body)
	deviceID, deviceErr := uuid.Parse(stringField(object, "device_id"))
	authorizationID, authorizationErr := uuid.Parse(stringField(object, "authorization_id"))
	familyID, familyErr := uuid.Parse(stringField(object, "family_id"))
	if deviceErr != nil || authorizationErr != nil || familyErr != nil || stringField(object, "access_token") == "" || stringField(object, "refresh_token") == "" {
		clear(prepared.signingKey)
		t.Fatal("c11 race device token response invalid")
	}
	return c11Device{
		baseURL: baseURL, accessToken: stringField(object, "access_token"), refreshToken: stringField(object, "refresh_token"),
		deviceID: deviceID, authorizationID: authorizationID, familyID: familyID, signingPublic: prepared.signingPublic,
		signingKey: prepared.signingKey, hpkePrivate: prepared.hpkePrivate,
	}
}

func waitReadinessCheck(
	parent context.Context,
	t *testing.T,
	client *http.Client,
	url string,
	expectedStatus int,
	check, expectedState string,
	bound time.Duration,
) {
	t.Helper()
	if parent == nil {
		t.Fatal("c11 readiness parent context missing")
	}
	deadline := time.Now().Add(bound)
	for {
		ctx, cancel := context.WithTimeout(parent, 2*time.Second)
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			cancel()
			t.Fatal("c11 readiness request construction failed")
		}
		response, requestErr := client.Do(request)
		var body struct {
			Checks map[string]string `json:"checks"`
		}
		decoded := requestErr == nil && json.NewDecoder(response.Body).Decode(&body) == nil
		if response != nil {
			_ = response.Body.Close()
		}
		cancel()
		if decoded && response.StatusCode == expectedStatus && body.Checks[check] == expectedState {
			return
		}
		if parent.Err() != nil {
			t.Fatal("c11 readiness row deadline exceeded")
		}
		if !time.Now().Before(deadline) {
			t.Fatalf("c11 readiness check %q did not reach %d/%q", check, expectedStatus, expectedState)
		}
		select {
		case <-time.After(100 * time.Millisecond):
		case <-parent.Done():
			t.Fatal("c11 readiness row deadline exceeded")
		}
	}
}

func waitHTTPStatus(parent context.Context, t *testing.T, client *http.Client, url string, expected int, bound time.Duration) {
	t.Helper()
	if parent == nil {
		t.Fatal("c11 health parent context missing")
	}
	deadline := time.Now().Add(bound)
	for {
		ctx, cancel := context.WithTimeout(parent, 2*time.Second)
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			cancel()
			t.Fatal("c11 health request construction failed")
		}
		response, requestErr := client.Do(request)
		if requestErr == nil {
			_ = response.Body.Close()
			if response.StatusCode == expected {
				cancel()
				return
			}
		}
		cancel()
		if parent.Err() != nil {
			t.Fatal("c11 health row deadline exceeded")
		}
		if !time.Now().Before(deadline) {
			t.Fatalf("c11 health status did not reach %d", expected)
		}
		select {
		case <-time.After(100 * time.Millisecond):
		case <-parent.Done():
			t.Fatal("c11 health row deadline exceeded")
		}
	}
}

var _ identity.EmailSender = (*controlledEmailSender)(nil)
var _ trust.ConfigSigner = (*controlledConfigSigner)(nil)
var _ errorreport.Provider = (*controlledReportProvider)(nil)
