package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nats-io/nats.go"
	"talenro.local/platform/internal/buildinfo"
	"talenro.local/platform/internal/config"
	contractevents "talenro.local/platform/internal/contracts/events"
	"talenro.local/platform/internal/deviceauth"
	"talenro.local/platform/internal/errorreport"
	"talenro.local/platform/internal/identity"
	"talenro.local/platform/internal/observability"
	"talenro.local/platform/internal/platform"
	"talenro.local/platform/internal/ratelimit"
	"talenro.local/platform/internal/secret"
	"talenro.local/platform/internal/sensitive"
	"talenro.local/platform/internal/store"
	"talenro.local/platform/internal/trust"
)

func TestRunReportsSanitizedDependencyError(t *testing.T) {
	logs := captureLogs(t)
	private := "credential@private.example:6543"
	cause := errors.New(private)
	lookup := testLookup(map[string]string{
		"TALENRO_DATABASE_URL": "postgres://unused",
	})
	factory := func(context.Context, config.Config, *observability.Registry) (*openedRuntime, error) {
		return nil, cause
	}

	err := runWithFactory(context.Background(), lookup, factory)
	reportStopped(err)
	if got := err.Error(); got != string(platform.CategoryDependencies) {
		t.Fatalf("error = %q, want %q", got, platform.CategoryDependencies)
	}
	if !errors.Is(err, cause) {
		t.Fatal("categorized error did not retain internal cause")
	}
	assertSanitizedLog(t, logs.String(), "control_api_stopped", string(platform.CategoryDependencies), private)
}

func TestRunShutsDownHTTPServerAndClosesRuntime(t *testing.T) {
	address := unusedLocalAddress(t)
	lookup := testLookup(map[string]string{
		"TALENRO_DATABASE_URL":    "postgres://unused",
		"TALENRO_HTTP_ADDRESS":    address,
		"TALENRO_METRICS_ADDRESS": unusedLocalAddress(t),
	})

	var closeCalls atomic.Int32
	factory := func(context.Context, config.Config, *observability.Registry) (*openedRuntime, error) {
		return &openedRuntime{
			handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
			close: func(context.Context) error { closeCalls.Add(1); return nil },
		}, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	errCh := make(chan error, 1)
	go func() { errCh <- runWithFactory(ctx, lookup, factory) }()
	waitForServer(t, "http://"+address)
	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("runWithFactory returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("runWithFactory did not return within 2s of cancellation")
	}
	if got := closeCalls.Load(); got != 1 {
		t.Fatalf("runtime close calls = %d, want 1", got)
	}

	client := &http.Client{Timeout: 100 * time.Millisecond}
	if response, err := client.Do(newGETRequest(t, "http://"+address)); err == nil {
		_ = response.Body.Close()
		t.Fatal("HTTP server still accepted requests after shutdown")
	}
}

func TestRunServesMetricsOnlyOnPrivateListener(t *testing.T) {
	publicAddress := unusedLocalAddress(t)
	metricsAddress := unusedLocalAddress(t)
	lookup := testLookup(map[string]string{
		"TALENRO_DATABASE_URL":    "postgres://unused",
		"TALENRO_HTTP_ADDRESS":    publicAddress,
		"TALENRO_METRICS_ADDRESS": metricsAddress,
	})

	factory := func(context.Context, config.Config, *observability.Registry) (*openedRuntime, error) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /livez", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		return &openedRuntime{handler: mux, close: func(context.Context) error { return nil }}, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	errCh := make(chan error, 1)
	go func() { errCh <- runWithFactory(ctx, lookup, factory) }()
	waitForServer(t, "http://"+publicAddress+"/livez")
	waitForServer(t, "http://"+metricsAddress+"/metrics")

	assertHTTPStatus(t, "http://"+publicAddress+"/metrics", http.StatusNotFound)
	assertHTTPStatus(t, "http://"+metricsAddress+"/livez", http.StatusNotFound)
	response, err := http.DefaultClient.Do(newGETRequest(t, "http://"+metricsAddress+"/metrics"))
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"talenro_control_build_info",
		"talenro_control_http_requests_total",
		`route="/livez"`,
	} {
		if !strings.Contains(string(body), expected) {
			t.Fatalf("metrics response missing %q: %s", expected, body)
		}
	}

	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("runWithFactory returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("runWithFactory did not stop both listeners")
	}
}

func TestRunClosesRuntimeWhenHTTPServerFails(t *testing.T) {
	logs := captureLogs(t)
	listener, err := new(net.ListenConfig).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = listener.Close() }()

	lookup := testLookup(map[string]string{
		"TALENRO_DATABASE_URL":    "postgres://unused",
		"TALENRO_HTTP_ADDRESS":    listener.Addr().String(),
		"TALENRO_METRICS_ADDRESS": unusedLocalAddress(t),
	})
	var closeCalls atomic.Int32
	factory := func(context.Context, config.Config, *observability.Registry) (*openedRuntime, error) {
		return &openedRuntime{
			handler: http.NewServeMux(),
			close:   func(context.Context) error { closeCalls.Add(1); return nil },
		}, nil
	}

	err = runWithFactory(context.Background(), lookup, factory)
	reportStopped(err)
	if err == nil || err.Error() != string(platform.CategoryHTTPListenOrServe) {
		t.Fatalf("runWithFactory error = %v, want %q", err, platform.CategoryHTTPListenOrServe)
	}
	if got := closeCalls.Load(); got != 1 {
		t.Fatalf("runtime close calls = %d, want 1", got)
	}
	assertSanitizedLog(
		t,
		logs.String(),
		"control_api_stopped",
		string(platform.CategoryHTTPListenOrServe),
		listener.Addr().String(),
	)
	assertServerStopped(t, "http://"+lookupValue(t, lookup, "TALENRO_METRICS_ADDRESS")+"/metrics")
}

func TestShutdownForcesClosedBlockingHandler(t *testing.T) {
	logs := captureLogs(t)
	listener, err := new(net.ListenConfig).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })

	requestStarted := make(chan struct{})
	handlerFinished := make(chan struct{})
	handler := http.HandlerFunc(func(_ http.ResponseWriter, request *http.Request) {
		close(requestStarted)
		<-request.Context().Done()
		close(handlerFinished)
	})
	server := platform.NewHTTPServer(listener.Addr().String(), handler)
	serveDone := make(chan struct{})
	go func() {
		_ = server.Serve(listener)
		close(serveDone)
	}()

	clientDone := make(chan struct{})
	clientRequest := newGETRequest(t, "http://"+listener.Addr().String())
	go func() {
		client := &http.Client{Timeout: 2 * time.Second}
		response, requestErr := client.Do(clientRequest)
		if response != nil {
			_ = response.Body.Close()
		}
		_ = requestErr
		close(clientDone)
	}()

	select {
	case <-requestStarted:
	case <-time.After(time.Second):
		t.Fatal("blocking handler did not start within 1s")
	}

	started := time.Now()
	err = shutdownHTTPServer(server, 50*time.Millisecond)
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("forced shutdown took %s, want at most 1s", elapsed)
	}
	if err == nil || err.Error() != string(platform.CategoryHTTPShutdown) {
		t.Fatalf("shutdown error = %v, want %q", err, platform.CategoryHTTPShutdown)
	}

	for name, done := range map[string]<-chan struct{}{
		"handler": handlerFinished,
		"client":  clientDone,
		"server":  serveDone,
	} {
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatalf("%s remained active after forced close", name)
		}
	}
	assertSanitizedLog(
		t,
		logs.String(),
		"control_api_forced_close",
		string(platform.CategoryHTTPForcedClose),
		listener.Addr().String(),
	)
}

func TestShutdownContextRetainsLifecycleValuesAfterCancellation(t *testing.T) {
	type contextKey struct{}

	parent, cancelParent := context.WithCancel(context.WithValue(t.Context(), contextKey{}, "request-id"))
	cancelParent()

	shutdownCtx, cancelShutdown := deriveShutdownContext(parent, time.Second)
	defer cancelShutdown()

	if err := shutdownCtx.Err(); err != nil {
		t.Fatalf("shutdown context inherited cancellation: %v", err)
	}
	if got := shutdownCtx.Value(contextKey{}); got != "request-id" {
		t.Fatalf("shutdown context value = %v, want request-id", got)
	}
	if _, ok := shutdownCtx.Deadline(); !ok {
		t.Fatal("shutdown context has no deadline")
	}
}

func TestTask18RuntimeValidationRejectsUnsafeProductionAndUnavailableAdapters(t *testing.T) {
	local, err := config.Load(testLookup(map[string]string{"TALENRO_DATABASE_URL": "postgres://unused"}))
	if err != nil {
		t.Fatal(err)
	}
	if err := validateRuntimeComposition(local); err != nil {
		t.Fatalf("validated local configuration rejected: %v", err)
	}

	production := local
	production.Security.Profile = config.ProfileProduction
	production.Security.EmailVerification = config.EmailRequired
	production.Security.PublicBaseURL = "https://api.example"
	production.Security.BundleBaseURLs = [3]string{"https://primary.example", "https://mirror-a.example", "https://mirror-b.example"}
	production.Security.WebAuthnOrigins = []string{"https://api.example"}
	production.Security.SignerProvider = config.ProviderExternal
	production.Security.FieldProtectorProvider = config.ProviderExternal
	production.Security.EmailProvider = config.ProviderExternal
	production.Security.ErrorReporterProvider = config.ProviderExternal

	tests := []struct {
		name   string
		mutate func(*config.Config)
		want   error
	}{
		{name: "local signer", mutate: func(value *config.Config) { value.Security.SignerProvider = config.ProviderLocal }, want: errRuntimeConfiguration},
		{name: "local protector", mutate: func(value *config.Config) { value.Security.FieldProtectorProvider = config.ProviderLocal }, want: errRuntimeConfiguration},
		{name: "local email", mutate: func(value *config.Config) { value.Security.EmailProvider = config.ProviderLocal }, want: errRuntimeConfiguration},
		{name: "discard reporter", mutate: func(value *config.Config) { value.Security.ErrorReporterProvider = config.ProviderDiscard }, want: errRuntimeConfiguration},
		{name: "HTTP public origin", mutate: func(value *config.Config) { value.Security.PublicBaseURL = "http://api.example" }, want: errRuntimeConfiguration},
		{name: "HTTP bundle source", mutate: func(value *config.Config) { value.Security.BundleBaseURLs[1] = "http://localhost:8081" }, want: errRuntimeConfiguration},
		{name: "HTTP WebAuthn origin", mutate: func(value *config.Config) { value.Security.WebAuthnOrigins[0] = "http://localhost:8080" }, want: errRuntimeConfiguration},
		{name: "external adapters not linked", mutate: func(*config.Config) {}, want: errRuntimeAdapterUnavailable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			candidate := production
			candidate.Security.WebAuthnOrigins = append([]string(nil), production.Security.WebAuthnOrigins...)
			test.mutate(&candidate)
			err := validateRuntimeComposition(candidate)
			if !errors.Is(err, test.want) {
				t.Fatalf("validation error = %v, want %v", err, test.want)
			}
			if err != nil && strings.Contains(err.Error(), "api.example") {
				t.Fatalf("validation exposed configured URL: %q", err)
			}
		})
	}
}

func TestTask18LocalProfileConstructsAndClosesProviderFixtures(t *testing.T) {
	cfg, err := config.Load(testLookup(map[string]string{"TALENRO_DATABASE_URL": "postgres://unused"}))
	if err != nil {
		t.Fatal(err)
	}
	providers, err := openRuntimeProviders(t.Context(), cfg, observability.NewRegistry())
	if err != nil {
		t.Fatal(err)
	}
	if providers.protector == nil || providers.email == nil || providers.reporter == nil ||
		providers.rootSigner == nil || providers.configSigner == nil || providers.signer == nil {
		t.Fatalf("local provider set is incomplete: %#v", providers)
	}
	shutdown, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	providers.Close(shutdown)
	providers.Close(shutdown)
}

func TestTask18LocalProvidersDriveCryptoMetricsFromRealOperations(t *testing.T) {
	const canary = "CANARY-private-plaintext-8472"
	cfg, err := config.Load(testLookup(map[string]string{"TALENRO_DATABASE_URL": "postgres://unused"}))
	if err != nil {
		t.Fatal(err)
	}
	metrics := observability.NewRegistry()
	providers, err := openRuntimeProviders(t.Context(), cfg, metrics)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { providers.Close(context.Background()) })
	if digest := providers.protector.LookupDigest("task18.lookup", []byte(canary)); digest == [32]byte{} {
		t.Fatal("real local lookup failed")
	}
	protected, err := providers.protector.Encrypt("task18.field", []byte(canary))
	if err != nil {
		t.Fatal(err)
	}
	opened, err := providers.protector.Decrypt("task18.field", protected)
	clear(protected.Ciphertext)
	if err != nil || string(opened) != canary {
		t.Fatalf("real local decrypt failed: %v", err)
	}
	clear(opened)
	signature, err := providers.signer.Sign(t.Context(), []byte(canary))
	if err != nil || len(signature) != 64 {
		t.Fatalf("real local sign = %d bytes, %v", len(signature), err)
	}
	clear(signature)

	want := map[string]float64{
		task18CryptoMetricKey(observability.CryptoOperationLookup, observability.MetricResultSuccess, observability.CryptoReasonNone):          1,
		task18CryptoMetricKey(observability.CryptoOperationEncrypt, observability.MetricResultSuccess, observability.CryptoReasonNone):         1,
		task18CryptoMetricKey(observability.CryptoOperationDecrypt, observability.MetricResultSuccess, observability.CryptoReasonNone):         1,
		task18CryptoMetricKey(observability.CryptoOperation("bundle_sign"), observability.MetricResultSuccess, observability.CryptoReasonNone): 1,
	}
	got := task18GatheredCryptoSeries(t, metrics, canary)
	if len(got) != len(want) {
		t.Fatalf("production crypto series = %v, want exactly %v", got, want)
	}
	for labels, count := range want {
		if got[labels] != count {
			t.Errorf("production crypto series %q = %v, want %v", labels, got[labels], count)
		}
	}
}

func TestTask18RuntimeCryptoObserverRejectsNonFiniteEvents(t *testing.T) {
	const canary = "CANARY_private_crypto_label_8472"
	metrics := observability.NewRegistry()
	observer := runtimeCryptoObserver{metrics: metrics}
	observer.ObserveSensitiveCrypto(sensitive.CryptoEvent{
		Operation: sensitive.CryptoOperation(canary), Result: sensitive.CryptoResult(canary), Reason: sensitive.CryptoReason(canary),
	})
	observer.ObserveDeviceCrypto(deviceauth.CryptoEvent{
		Operation: deviceauth.CryptoOperation(canary), Result: deviceauth.CryptoResult(canary), Reason: deviceauth.CryptoReason(canary),
	})
	observer.ObserveTrustCrypto(trust.CryptoEvent{
		Operation: trust.CryptoOperation(canary), Result: trust.CryptoResult(canary), Reason: trust.CryptoReason(canary),
	})
	if got := task18GatheredCryptoSeries(t, metrics, canary); len(got) != 0 {
		t.Fatalf("non-finite crypto events created series: %v", got)
	}
}

func TestTask18RuntimeShutdownUsesRequiredOrderOnce(t *testing.T) {
	var order []string
	step := func(name string) func(context.Context) error {
		return func(context.Context) error { order = append(order, name); return nil }
	}
	shutdown := &runtimeShutdown{
		stopIntake:        step("intake"),
		stopPublisher:     step("publisher"),
		drainReporter:     func(context.Context) { order = append(order, "reporter") },
		closeProviders:    func() { order = append(order, "providers") },
		closeDependencies: func() { order = append(order, "nats-redis-postgres") },
	}
	if err := shutdown.Close(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := shutdown.Close(context.Background()); err != nil {
		t.Fatal(err)
	}
	want := []string{"intake", "publisher", "reporter", "providers", "nats-redis-postgres"}
	if strings.Join(order, ",") != strings.Join(want, ",") {
		t.Fatalf("shutdown order = %v, want %v", order, want)
	}
}

func TestTask18ComposeApplicationsFillsEveryHTTPApplicationSurface(t *testing.T) {
	cfg, err := config.Load(testLookup(map[string]string{"TALENRO_DATABASE_URL": "postgres://unused"}))
	if err != nil {
		t.Fatal(err)
	}
	metrics := observability.NewRegistry()
	providers, err := openRuntimeProviders(t.Context(), cfg, metrics)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { providers.Close(context.Background()) })
	database := task18PGXDatabase{}
	identityRepository, err := identity.NewPostgresRepository(database, providers.protector)
	if err != nil {
		t.Fatal(err)
	}
	deviceRepository, err := deviceauth.NewPostgresRepository(database, providers.protector)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)
	metadata := trust.RootMetadataV1{
		SchemaVersion: trust.TrustMetadataSchemaV1, Version: "1", RootKeyID: providers.rootSigner.KeyID(),
		RootAlgorithm: trust.SignatureAlgorithm, ValidFrom: now.Add(-time.Hour).Format(time.RFC3339),
		ValidUntil: now.Add(365 * 24 * time.Hour).Format(time.RFC3339),
		SigningKeys: []trust.SigningKeyMetadataV1{{
			KeyID: providers.configSigner.KeyID(), Algorithm: trust.SignatureAlgorithm,
			PublicKey: base64.RawURLEncoding.EncodeToString(providers.configSigner.PublicKey()), State: "active",
			NotBefore: now.Add(-time.Hour).Format(time.RFC3339), NotAfter: now.Add(365 * 24 * time.Hour).Format(time.RFC3339),
		}},
	}
	immutable := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	apps, err := composeRuntimeApplications(cfg, providers, runtimeApplicationDependencies{
		identityRepository: identityRepository, deviceRepository: deviceRepository,
		limiter: task18Limiter{}, identityChallenges: task18IdentityChallenges{}, deviceChallenges: task18DeviceChallenges{},
		trustRepository: task18TrustRepository{}, metadata: metadata, testConfig: task18TestConfig{},
		immutable: immutable, clock: task18Clock{now: now},
	})
	if err != nil {
		t.Fatal(err)
	}
	if apps.Identity == nil || apps.StrongAuth == nil || apps.AccountAuth == nil ||
		apps.Device == nil || apps.DeviceAuth == nil || apps.Trust == nil || apps.ImmutableBundle == nil {
		t.Fatalf("composition left an HTTP application surface empty: %#v", apps)
	}
	accessToken := secret.NewBytes(bytes.Repeat([]byte{0x77}, 32))
	_, authenticationErr := apps.DeviceAuth.Authenticate(t.Context(), accessToken)
	accessToken.Clear()
	if authenticationErr == nil {
		t.Fatal("fixture device authentication unexpectedly succeeded")
	}
	want := task18CryptoMetricKey(
		observability.CryptoOperationTokenVerify,
		observability.MetricResultFailure,
		observability.CryptoReasonKeyUnavailable,
	)
	if got := task18GatheredCryptoSeries(t, metrics, "CANARY"); got[want] != 1 {
		t.Fatalf("composed device crypto series = %v, want %q once", got, want)
	}
}

func TestTask18EmailIntakeAcknowledgesOnlyAfterConsumerSuccess(t *testing.T) {
	now := time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)
	message := &task18RuntimeMessage{data: []byte("canonical-envelope")}
	subscription := newTask18RuntimeSubscription(message)
	consumer := &task18RuntimeEmailConsumer{called: make(chan []byte, 1)}
	intake, err := newRuntimeEmailIntake(subscription, consumer, task18Clock{now: now}, newTask18RuntimeReporter())
	if err != nil {
		t.Fatal(err)
	}
	if !intake.Start(context.Background()) {
		t.Fatal("email intake did not start")
	}
	select {
	case body := <-consumer.called:
		if string(body) != "canonical-envelope" || !consumer.consumedAt.Equal(now) {
			t.Fatalf("consumer input = %q at %s", body, consumer.consumedAt)
		}
	case <-time.After(time.Second):
		t.Fatal("email intake did not consume the message")
	}
	deadline := time.Now().Add(time.Second)
	for message.acks.Load() == 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if message.acks.Load() != 1 || message.naks.Load() != 0 {
		t.Fatalf("message ack/nak = %d/%d, want 1/0", message.acks.Load(), message.naks.Load())
	}
	shutdown, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := intake.Close(shutdown); err != nil {
		t.Fatal(err)
	}
	if subscription.unsubscribes.Load() != 1 {
		t.Fatalf("unsubscribe calls = %d, want 1", subscription.unsubscribes.Load())
	}
}

func TestTask18EmailIntakeNaksFailureAndHonorsShutdownDeadline(t *testing.T) {
	message := &task18RuntimeMessage{data: []byte("provider-private-CANARY")}
	subscription := newTask18RuntimeSubscription(message)
	consumer := &task18RuntimeEmailConsumer{called: make(chan []byte, 1), err: errors.New("CANARY provider detail")}
	intake, err := newRuntimeEmailIntake(subscription, consumer, task18Clock{now: time.Now().UTC()}, newTask18RuntimeReporter())
	if err != nil || !intake.Start(context.Background()) {
		t.Fatalf("start intake: %v", err)
	}
	select {
	case <-consumer.called:
	case <-time.After(time.Second):
		t.Fatal("email intake did not consume failing message")
	}
	deadline := time.Now().Add(time.Second)
	for message.naks.Load() == 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if message.acks.Load() != 0 || message.naks.Load() != 1 {
		t.Fatalf("message ack/nak = %d/%d, want 0/1", message.acks.Load(), message.naks.Load())
	}
	shutdown, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := intake.Close(shutdown); err != nil {
		t.Fatal(err)
	}
	if err := intake.Close(shutdown); err != nil || subscription.unsubscribes.Load() != 1 {
		t.Fatalf("second close = %v, unsubscribe calls = %d", err, subscription.unsubscribes.Load())
	}
}

func TestTask18EmailIntakeBacksOffEmptyOrFailedPulls(t *testing.T) {
	for _, test := range []struct {
		name string
		err  error
	}{
		{name: "failed pull", err: errors.New("CANARY nats private detail")},
		{name: "empty pull"},
	} {
		t.Run(test.name, func(t *testing.T) {
			subscription := &task18FailingRuntimeSubscription{first: make(chan struct{}), err: test.err}
			consumer := &task18RuntimeEmailConsumer{called: make(chan []byte, 1)}
			intake, err := newRuntimeEmailIntake(subscription, consumer, task18Clock{now: time.Now().UTC()}, newTask18RuntimeReporter())
			if err != nil || !intake.Start(context.Background()) {
				t.Fatalf("start intake: %v", err)
			}
			select {
			case <-subscription.first:
			case <-time.After(time.Second):
				t.Fatal("email intake did not attempt the first pull")
			}
			time.Sleep(25 * time.Millisecond)
			if calls := subscription.calls.Load(); calls != 1 {
				t.Fatalf("pull calls before retry budget = %d, want 1", calls)
			}
			shutdown, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			if err := intake.Close(shutdown); err != nil {
				t.Fatal(err)
			}
			if subscription.unsubscribes.Load() != 1 {
				t.Fatalf("unsubscribe calls = %d, want 1", subscription.unsubscribes.Load())
			}
		})
	}
}

func TestTask18EmailIntakeCancelsBeforeUnsubscribe(t *testing.T) {
	subscription := &task18CancellationOrderSubscription{started: make(chan struct{})}
	reporter := newTask18RuntimeReporter()
	intake, err := newRuntimeEmailIntake(
		subscription,
		&task18RuntimeEmailConsumer{called: make(chan []byte, 1)},
		task18Clock{now: time.Now().UTC()},
		reporter,
	)
	if err != nil || !intake.Start(context.Background()) {
		t.Fatalf("start intake: %v", err)
	}
	select {
	case <-subscription.started:
	case <-time.After(time.Second):
		t.Fatal("email intake did not enter its blocking pull")
	}

	shutdown, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := intake.Close(shutdown); err != nil {
		t.Fatal(err)
	}
	if !subscription.cancelledBeforeUnsubscribe.Load() {
		t.Fatal("email intake unsubscribed before cancelling its worker context")
	}
	select {
	case report := <-reporter.reports:
		t.Fatalf("cancelled pull submitted report %#v", report)
	default:
	}
}

func TestTask18EmailIntakeDoesNotReportOrAcknowledgeAfterCancellation(t *testing.T) {
	for _, test := range []struct {
		name       string
		consumeErr bool
	}{
		{name: "consumer returns cancellation", consumeErr: true},
		{name: "consumer reports success after cancellation"},
	} {
		t.Run(test.name, func(t *testing.T) {
			message := &task18RuntimeMessage{data: []byte("cancellation-boundary")}
			subscription := newTask18RuntimeSubscription(message)
			consumer := &task18CancellationBarrierConsumer{
				started:    make(chan struct{}),
				consumeErr: test.consumeErr,
			}
			reporter := newTask18RuntimeReporter()
			intake, err := newRuntimeEmailIntake(
				subscription,
				consumer,
				task18Clock{now: time.Now().UTC()},
				reporter,
			)
			if err != nil || !intake.Start(context.Background()) {
				t.Fatalf("start intake: %v", err)
			}
			select {
			case <-consumer.started:
			case <-time.After(time.Second):
				t.Fatal("email intake did not enter consumer")
			}

			shutdown, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			if err := intake.Close(shutdown); err != nil {
				t.Fatal(err)
			}
			if message.acks.Load() != 0 || message.naks.Load() != 0 {
				t.Fatalf("post-cancellation ack/nak = %d/%d, want 0/0", message.acks.Load(), message.naks.Load())
			}
			select {
			case report := <-reporter.reports:
				t.Fatalf("post-cancellation report = %#v", report)
			default:
			}
		})
	}
}

func TestTask18EmailIntakeCloseIsBoundedDuringBlockedExternalCallbacks(t *testing.T) {
	for _, callback := range []string{"ack", "nak", "reporter"} {
		t.Run(callback, func(t *testing.T) {
			blocked := newTask18BlockingCall()
			message := &task18BlockingRuntimeMessage{data: []byte("bounded-close")}
			consumer := &task18RuntimeEmailConsumer{called: make(chan []byte, 1)}
			var reporter errorreport.Reporter = newTask18RuntimeReporter()
			switch callback {
			case "ack":
				message.ack = blocked
			case "nak":
				message.nak = blocked
				consumer.err = errors.New("finite consumer failure")
			case "reporter":
				consumer.err = errors.New("finite consumer failure")
				reporter = task18BlockingReporter{call: blocked}
			}

			subscription := newTask18TrackedRuntimeSubscription(message)
			intake, err := newRuntimeEmailIntake(
				subscription,
				consumer,
				task18Clock{now: time.Now().UTC()},
				reporter,
			)
			if err != nil || !intake.Start(context.Background()) {
				t.Fatalf("start intake: %v", err)
			}
			select {
			case <-blocked.entered:
			case <-time.After(time.Second):
				blocked.Release()
				t.Fatal("email intake did not enter blocked callback")
			}

			shutdown, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()
			closeResult := make(chan error, 1)
			go func() { closeResult <- intake.Close(shutdown) }()

			var closeErr error
			select {
			case closeErr = <-closeResult:
			case <-time.After(500 * time.Millisecond):
				blocked.Release()
				closeErr = <-closeResult
				t.Fatalf("Close remained blocked beyond its deadline: %v", closeErr)
			}
			if !errors.Is(shutdown.Err(), context.DeadlineExceeded) || closeErr == nil {
				t.Fatalf("Close result/deadline = %v/%v, want bounded deadline failure", closeErr, shutdown.Err())
			}
			if !subscription.cancelledBeforeUnsubscribe.Load() {
				t.Fatal("Close did not cancel before unsubscribe while callback was blocked")
			}

			blocked.Release()
			select {
			case <-intake.done:
			case <-time.After(time.Second):
				t.Fatal("email worker did not exit after blocked callback was released")
			}
		})
	}
}

func TestTask18EmailIntakeContainsExternalPanicsAndContinues(t *testing.T) {
	for _, callback := range []string{"ack", "nak", "reporter"} {
		t.Run(callback, func(t *testing.T) {
			continued := make(chan struct{})
			first := &task18PanickingRuntimeMessage{data: []byte("panic-boundary")}
			second := &task18PanickingRuntimeMessage{data: []byte("continued"), acknowledged: continued}
			consumer := &task18ScriptedRuntimeEmailConsumer{}
			var reporter errorreport.Reporter = newTask18RuntimeReporter()
			switch callback {
			case "ack":
				first.panicAck = true
			case "nak":
				first.panicNak = true
				consumer.firstErr = errors.New("finite consumer failure")
			case "reporter":
				consumer.firstErr = errors.New("finite consumer failure")
				reporter = task18PanickingReporter{}
			}

			intake, err := newRuntimeEmailIntake(
				newTask18RuntimeSubscription(first, second),
				consumer,
				task18Clock{now: time.Now().UTC()},
				reporter,
			)
			if err != nil {
				t.Fatal(err)
			}
			workerCtx, cancelWorker := context.WithCancel(context.Background())
			intake.started = true
			intake.cancel = cancelWorker
			intake.done = make(chan struct{})
			recovered := make(chan any, 1)
			go func() {
				var value any
				func() {
					defer func() { value = recover() }()
					intake.run(workerCtx)
				}()
				recovered <- value
			}()

			select {
			case <-continued:
			case value := <-recovered:
				t.Fatalf("%s panic escaped the worker: %v", callback, value)
			case <-time.After(time.Second):
				t.Fatalf("worker did not continue after %s panic", callback)
			}

			shutdown, cancelShutdown := context.WithTimeout(context.Background(), time.Second)
			if err := intake.Close(shutdown); err != nil {
				cancelShutdown()
				t.Fatalf("close after contained %s panic: %v", callback, err)
			}
			cancelShutdown()
			select {
			case value := <-recovered:
				if value != nil {
					t.Fatalf("%s panic escaped during shutdown: %v", callback, value)
				}
			case <-time.After(time.Second):
				t.Fatalf("worker did not stop after contained %s panic", callback)
			}
		})
	}
}

func TestTask18EmailIntakeReportsOnlyFixedPrivacySafeFaults(t *testing.T) {
	const canary = "nats://private-user:private-password@CANARY.internal/message-8472" // #nosec G101 -- an intentional privacy canary, not a credential.
	tests := []struct {
		name         string
		subscription runtimeEmailSubscription
		consumer     *task18RuntimeEmailConsumer
		wantReports  int
	}{
		{
			name:         "pull failure",
			subscription: &task18FailingRuntimeSubscription{first: make(chan struct{}), err: errors.New(canary)},
			consumer:     &task18RuntimeEmailConsumer{called: make(chan []byte, 1)},
			wantReports:  1,
		},
		{
			name:         "consumer failure",
			subscription: newTask18RuntimeSubscription(&task18RuntimeMessage{data: []byte(canary)}),
			consumer:     &task18RuntimeEmailConsumer{called: make(chan []byte, 1), err: errors.New(canary)},
			wantReports:  1,
		},
		{
			name:         "ack failure",
			subscription: newTask18RuntimeSubscription(&task18RuntimeMessage{data: []byte(canary), ackErr: errors.New(canary)}),
			consumer:     &task18RuntimeEmailConsumer{called: make(chan []byte, 1)},
			wantReports:  1,
		},
		{
			name:         "consumer and nak failure",
			subscription: newTask18RuntimeSubscription(&task18RuntimeMessage{data: []byte(canary), nakErr: errors.New(canary)}),
			consumer:     &task18RuntimeEmailConsumer{called: make(chan []byte, 1), err: errors.New(canary)},
			wantReports:  2,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reporter := newTask18RuntimeReporter()
			intake, err := newRuntimeEmailIntake(
				test.subscription,
				test.consumer,
				task18Clock{now: time.Now().UTC()},
				reporter,
			)
			if err != nil || !intake.Start(context.Background()) {
				t.Fatalf("start intake: %v", err)
			}
			t.Cleanup(func() { _ = intake.Close(context.Background()) })

			for index := 0; index < test.wantReports; index++ {
				select {
				case report := <-reporter.reports:
					want := errorreport.Report{
						Event:        errorreport.EventWorkerFailure,
						Category:     errorreport.CategoryDependency,
						Component:    errorreport.ComponentEmail,
						Outcome:      errorreport.OutcomeFailure,
						Fingerprint:  errorreport.FingerprintDependency,
						BuildVersion: buildinfo.Current().Version,
						TraceID:      report.TraceID,
					}
					if report != want {
						t.Fatalf("email worker report = %#v, want %#v", report, want)
					}
					if len(report.TraceID) != 32 || strings.Trim(report.TraceID, "0123456789abcdef") != "" {
						t.Fatalf("email worker trace ID = %q, want fixed lowercase hex", report.TraceID)
					}
					encoded := strings.Join([]string{
						string(report.Event), string(report.Category), string(report.Component), string(report.Outcome),
						string(report.Fingerprint), report.BuildVersion, report.TraceID,
					}, "|")
					if strings.Contains(encoded, canary) {
						t.Fatalf("email worker report retained provider or message data: %s", encoded)
					}
				case <-time.After(time.Second):
					t.Fatalf("email worker submitted %d reports, want %d", index, test.wantReports)
				}
			}
		})
	}
}

func TestTask18RuntimeEventStreamRequiresDurableFiniteConfiguration(t *testing.T) {
	t.Run("accepts exact existing file stream", func(t *testing.T) {
		jetStream := &task18FiniteStreamJetStream{info: task18ExactEventStreamInfo()}
		if err := ensureRuntimeEventStream(jetStream); err != nil {
			t.Fatalf("ensure exact file stream: %v", err)
		}
		if jetStream.queriedName != "TALENRO_EVENTS" || jetStream.addConfig != nil {
			t.Fatalf("stream lookup/create = %q / %#v", jetStream.queriedName, jetStream.addConfig)
		}
	})

	invalidConfigurations := []struct {
		name   string
		mutate func(*nats.StreamConfig)
	}{
		{name: "name", mutate: func(config *nats.StreamConfig) { config.Name = "OTHER" }},
		{name: "subjects", mutate: func(config *nats.StreamConfig) { config.Subjects = []string{"talenro.>", "private.>"} }},
		{name: "retention", mutate: func(config *nats.StreamConfig) { config.Retention = nats.InterestPolicy }},
		{name: "storage", mutate: func(config *nats.StreamConfig) { config.Storage = nats.MemoryStorage }},
		{name: "discard", mutate: func(config *nats.StreamConfig) { config.Discard = nats.DiscardNew }},
		{name: "maximum message size", mutate: func(config *nats.StreamConfig) { config.MaxMsgSize = 264*1024 - 1 }},
		{name: "duplicate window", mutate: func(config *nats.StreamConfig) { config.Duplicates = time.Minute }},
	}
	for _, test := range invalidConfigurations {
		t.Run("rejects existing "+test.name, func(t *testing.T) {
			info := task18ExactEventStreamInfo()
			test.mutate(&info.Config)
			err := ensureRuntimeEventStream(&task18FiniteStreamJetStream{info: info})
			if !errors.Is(err, errRuntimeAdapterUnavailable) {
				t.Fatalf("ensure mismatched existing stream error = %v", err)
			}
		})
	}

	t.Run("creates and validates exact file stream", func(t *testing.T) {
		jetStream := &task18FiniteStreamJetStream{
			infoErr: nats.ErrStreamNotFound,
			addInfo: task18ExactEventStreamInfo(),
		}
		if err := ensureRuntimeEventStream(jetStream); err != nil {
			t.Fatalf("ensure missing stream: %v", err)
		}
		config := jetStream.addConfig
		if config == nil || config.Name != "TALENRO_EVENTS" || len(config.Subjects) != 1 ||
			config.Subjects[0] != "talenro.>" || config.Retention != nats.LimitsPolicy ||
			config.Storage != nats.FileStorage || config.Discard != nats.DiscardOld ||
			config.MaxMsgSize != 264*1024 || config.Duplicates != 2*time.Minute {
			t.Fatalf("created stream configuration = %#v", config)
		}
	})

	for _, test := range invalidConfigurations {
		t.Run("rejects created "+test.name, func(t *testing.T) {
			info := task18ExactEventStreamInfo()
			test.mutate(&info.Config)
			err := ensureRuntimeEventStream(&task18FiniteStreamJetStream{
				infoErr: nats.ErrStreamNotFound,
				addInfo: info,
			})
			if !errors.Is(err, errRuntimeAdapterUnavailable) {
				t.Fatalf("ensure mismatched created stream error = %v", err)
			}
		})
	}

	t.Run("rejects nil and failed create response", func(t *testing.T) {
		canary := errors.New("CANARY stream create detail")
		for _, jetStream := range []*task18FiniteStreamJetStream{
			{infoErr: nats.ErrStreamNotFound},
			{infoErr: nats.ErrStreamNotFound, addErr: canary},
		} {
			err := ensureRuntimeEventStream(jetStream)
			if !errors.Is(err, errRuntimeAdapterUnavailable) || strings.Contains(err.Error(), "CANARY") {
				t.Fatalf("stream create error = %v, want fixed adapter category", err)
			}
		}
	})

	t.Run("failed validation cannot expose publisher or subscribe", func(t *testing.T) {
		info := task18ExactEventStreamInfo()
		info.Config.Storage = nats.MemoryStorage
		jetStream := &task18FiniteStreamJetStream{info: info}
		subscribeCalls := 0
		opened, err := openRuntimeEmailSubscription(nil, runtimeNATSOperations{
			jetStream:    func(*nats.Conn) (runtimeJetStream, error) { return jetStream, nil },
			ensureStream: ensureRuntimeEventStream,
			subscribeEmail: func(runtimeJetStream) (runtimeEmailSubscription, error) {
				subscribeCalls++
				return newTask18RuntimeSubscription(), nil
			},
		})
		if !errors.Is(err, errRuntimeAdapterUnavailable) || opened.subscription != nil ||
			opened.publisher != nil || subscribeCalls != 0 || jetStream.publishCalls != 0 {
			t.Fatalf("invalid stream opened = %#v, error = %v, subscribe/publish = %d/%d",
				opened, err, subscribeCalls, jetStream.publishCalls)
		}
	})
}

func TestTask18NATSEmailSubscriptionEnsuresFiniteStreamBeforeDurableConsumer(t *testing.T) {
	var order []string
	subscription := newTask18RuntimeSubscription()
	jetStream := &task18RuntimeJetStream{}
	operations := runtimeNATSOperations{
		jetStream: func(*nats.Conn) (runtimeJetStream, error) {
			order = append(order, "jetstream")
			return jetStream, nil
		},
		ensureStream: func(got runtimeJetStream) error {
			if got != jetStream {
				t.Fatal("finite stream validation received a different JetStream context")
			}
			order = append(order, "stream")
			return nil
		},
		subscribeEmail: func(got runtimeJetStream) (runtimeEmailSubscription, error) {
			if got != jetStream {
				t.Fatal("durable consumer received a different JetStream context")
			}
			order = append(order, "consumer:"+contractevents.EmailDeliveryRequestedType)
			return subscription, nil
		},
	}
	got, err := openRuntimeEmailSubscription(nil, operations)
	if err != nil {
		t.Fatal(err)
	}
	if got.subscription != subscription || got.publisher != jetStream ||
		strings.Join(order, ",") != "jetstream,stream,consumer:"+contractevents.EmailDeliveryRequestedType {
		t.Fatalf("event stream/order = %#v / %v", got, order)
	}

	for index, fail := range []string{"jetstream", "stream", "consumer"} {
		t.Run(fail, func(t *testing.T) {
			private := errors.New("CANARY nats private detail")
			failing := operations
			switch index {
			case 0:
				failing.jetStream = func(*nats.Conn) (runtimeJetStream, error) { return nil, private }
			case 1:
				failing.ensureStream = func(runtimeJetStream) error { return private }
			case 2:
				failing.subscribeEmail = func(runtimeJetStream) (runtimeEmailSubscription, error) { return nil, private }
			}
			_, err := openRuntimeEmailSubscription(nil, failing)
			if !errors.Is(err, errRuntimeAdapterUnavailable) || strings.Contains(err.Error(), "CANARY") {
				t.Fatalf("subscription error = %v, want fixed adapter category", err)
			}
		})
	}
}

func TestTask18LocalMetadataBootstrapReturnsAuthenticatedDeterministicV1(t *testing.T) {
	cfg, err := config.Load(testLookup(map[string]string{"TALENRO_DATABASE_URL": "postgres://unused"}))
	if err != nil {
		t.Fatal(err)
	}
	providers, err := openRuntimeProviders(t.Context(), cfg, observability.NewRegistry())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { providers.Close(context.Background()) })
	now := time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)
	repository := newTask18MetadataRepository()
	payload, err := bootstrapLocalMetadata(context.Background(), repository, providers.rootSigner, providers.configSigner, now)
	if err != nil {
		t.Fatal(err)
	}
	if payload.Version != "1" || payload.RootKeyID != providers.rootSigner.KeyID() ||
		len(payload.SigningKeys) != 1 || payload.SigningKeys[0].KeyID != providers.configSigner.KeyID() {
		t.Fatalf("bootstrapped metadata = %#v", payload)
	}
	if repository.publishes != 1 {
		t.Fatalf("metadata publishes = %d, want 1", repository.publishes)
	}
	replayed, err := bootstrapLocalMetadata(context.Background(), repository, providers.rootSigner, providers.configSigner, now)
	if err != nil || replayed.Version != payload.Version || repository.publishes != 1 {
		t.Fatalf("metadata replay = %#v / %v; publishes=%d", replayed, err, repository.publishes)
	}
}

type task18MetadataRepository struct {
	records   []trust.SignedRootMetadataV1
	publishes int
}

func newTask18MetadataRepository() *task18MetadataRepository { return new(task18MetadataRepository) }

func (repository *task18MetadataRepository) List(context.Context) ([]trust.SignedRootMetadataV1, error) {
	return append([]trust.SignedRootMetadataV1(nil), repository.records...), nil
}

func (repository *task18MetadataRepository) Publish(_ context.Context, value trust.SignedRootMetadataV1) error {
	repository.publishes++
	repository.records = append(repository.records, value)
	return nil
}

type task18PGXDatabase struct{}

func (task18PGXDatabase) Begin(context.Context) (pgx.Tx, error) { return nil, errors.New("unused") }
func (task18PGXDatabase) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, errors.New("unused")
}
func (task18PGXDatabase) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, errors.New("unused")
}
func (task18PGXDatabase) QueryRow(context.Context, string, ...any) pgx.Row { return task18PGXRow{} }

type task18PGXRow struct{}

func (task18PGXRow) Scan(...any) error { return errors.New("unused") }

type task18Limiter struct{}

func (task18Limiter) Allow(context.Context, ratelimit.Operation, [sha256.Size]byte, config.RateLimitPolicy) (bool, error) {
	return true, nil
}

type task18IdentityChallenges struct{}

func (task18IdentityChallenges) Create(context.Context, identity.ChallengeRecord, time.Duration) error {
	return nil
}
func (task18IdentityChallenges) Consume(context.Context, string, [sha256.Size]byte) (identity.ChallengeRecord, error) {
	return identity.ChallengeRecord{}, nil
}

type task18DeviceChallenges struct{}

func (task18DeviceChallenges) Create(context.Context, deviceauth.ChallengeRecord, time.Duration) error {
	return nil
}
func (task18DeviceChallenges) Consume(context.Context, string, [sha256.Size]byte, [sha256.Size]byte) (deviceauth.ChallengeRecord, error) {
	return deviceauth.ChallengeRecord{}, nil
}

type task18TrustRepository struct{}

func (task18TrustRepository) WithinTransaction(context.Context, func(context.Context, trust.Transaction) error) error {
	return errors.New("unused")
}

type task18TestConfig struct{}

func (task18TestConfig) Current(context.Context) (trust.TestConfigV1, error) {
	return trust.TestConfigV1{Message: "local", Sequence: "1"}, nil
}

type task18Clock struct{ now time.Time }

func (clock task18Clock) Now() time.Time { return clock.now }

type task18RuntimeMessage struct {
	data   []byte
	ackErr error
	nakErr error
	acks   atomic.Int32
	naks   atomic.Int32
}

type task18BlockingCall struct {
	entered     chan struct{}
	release     chan struct{}
	enteredOnce sync.Once
	releaseOnce sync.Once
}

func newTask18BlockingCall() *task18BlockingCall {
	return &task18BlockingCall{entered: make(chan struct{}), release: make(chan struct{})}
}

func (call *task18BlockingCall) Invoke() error {
	call.enteredOnce.Do(func() { close(call.entered) })
	<-call.release
	return nil
}

func (call *task18BlockingCall) Release() {
	call.releaseOnce.Do(func() { close(call.release) })
}

type task18BlockingRuntimeMessage struct {
	data []byte
	ack  *task18BlockingCall
	nak  *task18BlockingCall
}

func (message *task18BlockingRuntimeMessage) Data() []byte {
	return append([]byte(nil), message.data...)
}

func (message *task18BlockingRuntimeMessage) Ack() error {
	if message.ack == nil {
		return nil
	}
	return message.ack.Invoke()
}

func (message *task18BlockingRuntimeMessage) Nak() error {
	if message.nak == nil {
		return nil
	}
	return message.nak.Invoke()
}

type task18PanickingRuntimeMessage struct {
	data         []byte
	panicAck     bool
	panicNak     bool
	acknowledged chan struct{}
	ackOnce      sync.Once
}

func (message *task18PanickingRuntimeMessage) Data() []byte {
	return append([]byte(nil), message.data...)
}

func (message *task18PanickingRuntimeMessage) Ack() error {
	if message.panicAck {
		panic("private ack panic")
	}
	if message.acknowledged != nil {
		message.ackOnce.Do(func() { close(message.acknowledged) })
	}
	return nil
}

func (message *task18PanickingRuntimeMessage) Nak() error {
	if message.panicNak {
		panic("private nak panic")
	}
	return nil
}

type task18BlockingReporter struct{ call *task18BlockingCall }

func (reporter task18BlockingReporter) TryReport(errorreport.Report) bool {
	_ = reporter.call.Invoke()
	return true
}

type task18PanickingReporter struct{}

func (task18PanickingReporter) TryReport(errorreport.Report) bool {
	panic("private reporter panic")
}

type task18ScriptedRuntimeEmailConsumer struct {
	firstErr error
	calls    atomic.Int32
}

func (consumer *task18ScriptedRuntimeEmailConsumer) Consume(context.Context, []byte, time.Time) error {
	if consumer.calls.Add(1) == 1 {
		return consumer.firstErr
	}
	return nil
}

type task18RuntimeJetStream struct{}

func (*task18RuntimeJetStream) PublishMsg(*nats.Msg, ...nats.PubOpt) (*nats.PubAck, error) {
	return &nats.PubAck{Stream: runtimeEventStreamName, Sequence: 1}, nil
}

func (*task18RuntimeJetStream) StreamInfo(string, ...nats.JSOpt) (*nats.StreamInfo, error) {
	return nil, nil
}

func (*task18RuntimeJetStream) AddStream(*nats.StreamConfig, ...nats.JSOpt) (*nats.StreamInfo, error) {
	return nil, nil
}

func (*task18RuntimeJetStream) PullSubscribe(string, string, ...nats.SubOpt) (*nats.Subscription, error) {
	return nil, nil
}

type task18FiniteStreamJetStream struct {
	info         *nats.StreamInfo
	infoErr      error
	addInfo      *nats.StreamInfo
	addErr       error
	queriedName  string
	addConfig    *nats.StreamConfig
	publishCalls int
}

func (jetStream *task18FiniteStreamJetStream) PublishMsg(*nats.Msg, ...nats.PubOpt) (*nats.PubAck, error) {
	jetStream.publishCalls++
	return &nats.PubAck{Stream: "TALENRO_EVENTS", Sequence: 1}, nil
}

func (jetStream *task18FiniteStreamJetStream) StreamInfo(name string, _ ...nats.JSOpt) (*nats.StreamInfo, error) {
	jetStream.queriedName = name
	return jetStream.info, jetStream.infoErr
}

func (jetStream *task18FiniteStreamJetStream) AddStream(config *nats.StreamConfig, _ ...nats.JSOpt) (*nats.StreamInfo, error) {
	if config != nil {
		copied := *config
		copied.Subjects = append([]string(nil), config.Subjects...)
		jetStream.addConfig = &copied
	}
	return jetStream.addInfo, jetStream.addErr
}

func (*task18FiniteStreamJetStream) PullSubscribe(string, string, ...nats.SubOpt) (*nats.Subscription, error) {
	return nil, nil
}

func task18ExactEventStreamInfo() *nats.StreamInfo {
	return &nats.StreamInfo{Config: nats.StreamConfig{
		Name:       "TALENRO_EVENTS",
		Subjects:   []string{"talenro.>"},
		Retention:  nats.LimitsPolicy,
		Storage:    nats.FileStorage,
		Discard:    nats.DiscardOld,
		MaxMsgSize: 264 * 1024,
		Duplicates: 2 * time.Minute,
	}}
}

func (message *task18RuntimeMessage) Data() []byte { return append([]byte(nil), message.data...) }
func (message *task18RuntimeMessage) Ack() error   { message.acks.Add(1); return message.ackErr }
func (message *task18RuntimeMessage) Nak() error   { message.naks.Add(1); return message.nakErr }

type task18RuntimeReporter struct {
	reports chan errorreport.Report
}

func newTask18RuntimeReporter() *task18RuntimeReporter {
	return &task18RuntimeReporter{reports: make(chan errorreport.Report, 32)}
}

func (reporter *task18RuntimeReporter) TryReport(report errorreport.Report) bool {
	select {
	case reporter.reports <- report:
		return true
	default:
		return false
	}
}

func task18GatheredCryptoSeries(t *testing.T, registry *observability.Registry, canary string) map[string]float64 {
	t.Helper()
	families, err := registry.Gatherer.Gather()
	if err != nil {
		t.Fatal(err)
	}
	series := make(map[string]float64)
	for _, family := range families {
		if canary != "" && strings.Contains(family.String(), canary) {
			t.Fatalf("private crypto input reached a metric: %s", family.String())
		}
		if family.GetName() != "talenro_crypto_validations_total" {
			continue
		}
		for _, metric := range family.GetMetric() {
			labels := make(map[string]string)
			for _, label := range metric.GetLabel() {
				labels[label.GetName()] = label.GetValue()
			}
			key := strings.Join([]string{labels["operation"], labels["result"], labels["reason"]}, "\x00")
			series[key] = metric.GetCounter().GetValue()
		}
	}
	return series
}

func task18CryptoMetricKey(
	operation observability.CryptoOperation,
	result observability.MetricResult,
	reason observability.CryptoReason,
) string {
	return strings.Join([]string{string(operation), string(result), string(reason)}, "\x00")
}

type task18RuntimeSubscription struct {
	messages     chan runtimeEmailMessage
	unsubscribes atomic.Int32
}

type task18TrackedRuntimeSubscription struct {
	messages                   chan runtimeEmailMessage
	mu                         sync.Mutex
	workerContext              context.Context
	cancelledBeforeUnsubscribe atomic.Bool
}

func newTask18TrackedRuntimeSubscription(messages ...runtimeEmailMessage) *task18TrackedRuntimeSubscription {
	channel := make(chan runtimeEmailMessage, len(messages))
	for _, message := range messages {
		channel <- message
	}
	return &task18TrackedRuntimeSubscription{messages: channel}
}

func (subscription *task18TrackedRuntimeSubscription) Next(ctx context.Context) (runtimeEmailMessage, error) {
	subscription.mu.Lock()
	subscription.workerContext = ctx
	subscription.mu.Unlock()
	select {
	case message := <-subscription.messages:
		return message, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (subscription *task18TrackedRuntimeSubscription) Unsubscribe() error {
	subscription.mu.Lock()
	workerContext := subscription.workerContext
	subscription.mu.Unlock()
	if workerContext != nil && workerContext.Err() != nil {
		subscription.cancelledBeforeUnsubscribe.Store(true)
	}
	return nil
}

type task18CancellationOrderSubscription struct {
	started                    chan struct{}
	startedOnce                sync.Once
	mu                         sync.Mutex
	workerContext              context.Context
	cancelledBeforeUnsubscribe atomic.Bool
}

func (subscription *task18CancellationOrderSubscription) Next(ctx context.Context) (runtimeEmailMessage, error) {
	subscription.mu.Lock()
	subscription.workerContext = ctx
	subscription.mu.Unlock()
	subscription.startedOnce.Do(func() { close(subscription.started) })
	<-ctx.Done()
	return nil, ctx.Err()
}

func (subscription *task18CancellationOrderSubscription) Unsubscribe() error {
	subscription.mu.Lock()
	workerContext := subscription.workerContext
	subscription.mu.Unlock()
	if workerContext != nil && workerContext.Err() != nil {
		subscription.cancelledBeforeUnsubscribe.Store(true)
	}
	return nil
}

type task18FailingRuntimeSubscription struct {
	err          error
	first        chan struct{}
	firstOnce    sync.Once
	calls        atomic.Int32
	unsubscribes atomic.Int32
}

func (subscription *task18FailingRuntimeSubscription) Next(context.Context) (runtimeEmailMessage, error) {
	subscription.calls.Add(1)
	subscription.firstOnce.Do(func() { close(subscription.first) })
	return nil, subscription.err
}

func (subscription *task18FailingRuntimeSubscription) Unsubscribe() error {
	subscription.unsubscribes.Add(1)
	return nil
}

func newTask18RuntimeSubscription(messages ...runtimeEmailMessage) *task18RuntimeSubscription {
	channel := make(chan runtimeEmailMessage, len(messages))
	for _, message := range messages {
		channel <- message
	}
	return &task18RuntimeSubscription{messages: channel}
}

func (subscription *task18RuntimeSubscription) Next(ctx context.Context) (runtimeEmailMessage, error) {
	select {
	case message := <-subscription.messages:
		return message, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (subscription *task18RuntimeSubscription) Unsubscribe() error {
	subscription.unsubscribes.Add(1)
	return nil
}

type task18RuntimeEmailConsumer struct {
	called     chan []byte
	err        error
	consumedAt time.Time
}

type task18CancellationBarrierConsumer struct {
	started     chan struct{}
	startedOnce sync.Once
	consumeErr  bool
}

func (consumer *task18CancellationBarrierConsumer) Consume(ctx context.Context, _ []byte, _ time.Time) error {
	consumer.startedOnce.Do(func() { close(consumer.started) })
	<-ctx.Done()
	if consumer.consumeErr {
		return ctx.Err()
	}
	return nil
}

func (consumer *task18RuntimeEmailConsumer) Consume(_ context.Context, encoded []byte, consumedAt time.Time) error {
	consumer.consumedAt = consumedAt
	consumer.called <- append([]byte(nil), encoded...)
	return consumer.err
}

var _ store.DBTX = task18PGXDatabase{}

func captureLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })
	return &logs
}

func assertSanitizedLog(t *testing.T, output, event, category, private string) {
	t.Helper()
	if !strings.Contains(output, "msg="+event) || !strings.Contains(output, "category="+category) {
		t.Fatalf("event/category missing from log: %q", output)
	}
	if strings.Contains(output, private) {
		t.Fatalf("private detail %q leaked in log: %q", private, output)
	}
}

func unusedLocalAddress(t *testing.T) string {
	t.Helper()
	listener, err := new(net.ListenConfig).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	return address
}

func testLookup(values map[string]string) config.Lookup {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}

func waitForServer(t *testing.T, url string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	client := &http.Client{Timeout: 100 * time.Millisecond}
	for time.Now().Before(deadline) {
		response, err := client.Do(newGETRequest(t, url))
		if err == nil {
			_, _ = io.Copy(io.Discard, response.Body)
			_ = response.Body.Close()
			if response.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("HTTP server did not become ready within 2s")
}

func assertHTTPStatus(t *testing.T, url string, want int) {
	t.Helper()
	client := &http.Client{Timeout: time.Second}
	response, err := client.Do(newGETRequest(t, url))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != want {
		t.Fatalf("GET %s status = %d, want %d", url, response.StatusCode, want)
	}
}

func assertServerStopped(t *testing.T, url string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	client := &http.Client{Timeout: 100 * time.Millisecond}
	for time.Now().Before(deadline) {
		response, err := client.Do(newGETRequest(t, url))
		if err == nil {
			_ = response.Body.Close()
			t.Fatalf("server still accepted a request at %s", url)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func newGETRequest(t *testing.T, target string) *http.Request {
	t.Helper()
	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, target, nil)
	if err != nil {
		t.Fatal(err)
	}
	return request
}

func lookupValue(t *testing.T, lookup config.Lookup, key string) string {
	t.Helper()
	value, ok := lookup(key)
	if !ok {
		t.Fatalf("test lookup missing %s", key)
	}
	return value
}
