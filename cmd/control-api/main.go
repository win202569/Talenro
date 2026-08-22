// Package main runs the Talenro control API process.
package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
	controlapiv1 "talenro.local/platform/gen/go/talenro/controlapi/v1"
	"talenro.local/platform/internal/buildinfo"
	"talenro.local/platform/internal/config"
	contractevents "talenro.local/platform/internal/contracts/events"
	"talenro.local/platform/internal/controlapi"
	"talenro.local/platform/internal/deviceauth"
	"talenro.local/platform/internal/errorreport"
	"talenro.local/platform/internal/identity"
	"talenro.local/platform/internal/observability"
	"talenro.local/platform/internal/outbox"
	"talenro.local/platform/internal/platform"
	postgresprobe "talenro.local/platform/internal/platform/postgres"
	redisprobe "talenro.local/platform/internal/platform/redis"
	"talenro.local/platform/internal/ratelimit"
	"talenro.local/platform/internal/readiness"
	"talenro.local/platform/internal/securitykit"
	"talenro.local/platform/internal/sensitive"
	"talenro.local/platform/internal/trust"
)

var (
	errRuntimeConfiguration      = errors.New("control-api: invalid runtime configuration")
	errRuntimeAdapterUnavailable = errors.New("control-api: configured adapter unavailable")
)

type openedRuntime struct {
	handler http.Handler
	close   func(context.Context) error
}

type runtimeFactory func(context.Context, config.Config, *observability.Registry) (*openedRuntime, error)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, os.LookupEnv); err != nil {
		reportStopped(err)
		os.Exit(1)
	}
}

func reportStopped(err error) {
	category := platform.ErrorCategoryOf(err, platform.CategoryInternal)
	slog.Error("control_api_stopped", "category", category)
}

func run(ctx context.Context, lookup config.Lookup) error {
	return runWithFactory(ctx, lookup, openRuntime)
}

func runWithFactory(ctx context.Context, lookup config.Lookup, factory runtimeFactory) error {
	cfg, err := config.Load(lookup)
	if err != nil {
		return platform.NewCategorizedError(platform.CategoryConfiguration, err)
	}

	metrics := observability.NewRegistry()
	runtime, err := factory(ctx, cfg, metrics)
	if err != nil {
		return platform.NewCategorizedError(platform.CategoryDependencies, err)
	}
	if runtime == nil || runtime.handler == nil || runtime.close == nil {
		return platform.NewCategorizedError(platform.CategoryDependencies, errRuntimeConfiguration)
	}
	defer func() {
		shutdownCtx, cancel := deriveShutdownContext(ctx, cfg.ShutdownTimeout)
		defer cancel()
		_ = runtime.close(shutdownCtx)
	}()

	slog.Info("control_api_starting", "category", platform.CategoryStartup)

	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", metrics.Handler())
	metricsServer := platform.NewHTTPServer(cfg.MetricsAddress, metricsMux)
	publicServer := platform.NewHTTPServer(cfg.HTTPAddress, metrics.Middleware("", runtime.handler))

	errCh := make(chan error, 2)
	go func() {
		errCh <- metricsServer.ListenAndServe()
	}()
	go func() {
		errCh <- publicServer.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		return shutdownHTTPServers(ctx, metricsServer, publicServer, cfg.ShutdownTimeout)
	case err := <-errCh:
		shutdownErr := shutdownHTTPServers(ctx, metricsServer, publicServer, cfg.ShutdownTimeout)
		if errors.Is(err, http.ErrServerClosed) {
			return shutdownErr
		}
		return platform.NewCategorizedError(platform.CategoryHTTPListenOrServe, err)
	}
}

func shutdownHTTPServer(server *http.Server, timeout time.Duration) error {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return shutdownHTTPServerWithContext(shutdownCtx, server)
}

func shutdownHTTPServers(
	lifecycleCtx context.Context,
	metricsServer, publicServer *http.Server,
	timeout time.Duration,
) error {
	shutdownCtx, cancel := deriveShutdownContext(lifecycleCtx, timeout)
	defer cancel()

	metricsErr := shutdownHTTPServerWithContext(shutdownCtx, metricsServer)
	publicErr := shutdownHTTPServerWithContext(shutdownCtx, publicServer)
	if metricsErr != nil {
		return metricsErr
	}
	return publicErr
}

func deriveShutdownContext(lifecycleCtx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(lifecycleCtx), timeout)
}

func shutdownHTTPServerWithContext(shutdownCtx context.Context, server *http.Server) error {
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Warn("control_api_shutdown_failed", "category", platform.CategoryHTTPShutdown)
		slog.Warn("control_api_forced_close", "category", platform.CategoryHTTPForcedClose)
		if closeErr := server.Close(); closeErr != nil {
			return platform.NewCategorizedError(platform.CategoryHTTPForcedClose, closeErr)
		}
		return platform.NewCategorizedError(platform.CategoryHTTPShutdown, err)
	}
	return nil
}

func openRuntime(ctx context.Context, cfg config.Config, metrics *observability.Registry) (*openedRuntime, error) {
	if err := validateRuntimeComposition(cfg); err != nil {
		return nil, err
	}
	if metrics == nil {
		return nil, errRuntimeConfiguration
	}
	deps, err := platform.Open(ctx, cfg)
	if err != nil {
		return nil, err
	}
	providers, err := openRuntimeProviders(ctx, cfg, metrics)
	if err != nil {
		deps.Close()
		return nil, err
	}
	shutdown := &runtimeShutdown{
		drainReporter: providers.reporter.Close, closeProviders: providers.closeResources,
		closeDependencies: deps.Close,
	}
	fail := func(failure error) (*openedRuntime, error) {
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), cfg.ShutdownTimeout)
		defer cancel()
		_ = shutdown.Close(shutdownCtx)
		return nil, failure
	}

	clock := systemClock{}
	limiter, err := ratelimit.NewRedis(deps.Redis, cfg.Security.RedisTimeout)
	if err != nil {
		return fail(errRuntimeConfiguration)
	}
	identityChallenges, err := identity.NewRedisChallengeStore(deps.Redis, cfg.Security.RedisTimeout)
	if err != nil {
		return fail(errRuntimeConfiguration)
	}
	deviceChallenges, err := deviceauth.NewRedisChallengeStore(deps.Redis, cfg.Security.RedisTimeout, clock)
	if err != nil {
		return fail(errRuntimeConfiguration)
	}
	identityRepository, err := identity.NewPostgresRepository(deps.Postgres, providers.protector)
	if err != nil {
		return fail(errRuntimeConfiguration)
	}
	deviceRepository, err := deviceauth.NewPostgresRepository(deps.Postgres, providers.protector)
	if err != nil {
		return fail(errRuntimeConfiguration)
	}
	trustRepository, err := trust.NewPostgresRepository(deps.Postgres, providers.protector)
	if err != nil {
		return fail(errRuntimeConfiguration)
	}
	byteStore, err := trust.NewPostgresByteStore(deps.Postgres)
	if err != nil {
		return fail(errRuntimeConfiguration)
	}
	distribution, err := trust.NewDistribution(byteStore, clock)
	if err != nil {
		return fail(errRuntimeConfiguration)
	}
	rootPublic := providers.rootSigner.PublicKey()
	metadataRepository, err := trust.NewPostgresMetadataRepository(
		deps.Postgres, map[string]ed25519.PublicKey{providers.rootSigner.KeyID(): rootPublic},
	)
	clear(rootPublic)
	if err != nil {
		return fail(errRuntimeConfiguration)
	}
	metadata, err := bootstrapLocalMetadata(ctx, metadataRepository, providers.rootSigner, providers.configSigner, clock.Now())
	if err != nil {
		return fail(errRuntimeAdapterUnavailable)
	}
	apps, err := composeRuntimeApplications(cfg, providers, runtimeApplicationDependencies{
		identityRepository: identityRepository, deviceRepository: deviceRepository,
		limiter: limiter, identityChallenges: identityChallenges, deviceChallenges: deviceChallenges,
		trustRepository: trustRepository, metadata: metadata, testConfig: localTestConfigSource{},
		immutable: distribution, clock: clock,
	})
	if err != nil {
		return fail(err)
	}

	emailRepository, err := identity.NewPostgresEmailDeliveryRepository(deps.Postgres)
	if err != nil {
		return fail(errRuntimeConfiguration)
	}
	emailConsumer, err := identity.NewEmailConsumer(emailRepository, providers.protector, providers.email, cfg.Security.RequestDeadline)
	if err != nil {
		return fail(errRuntimeConfiguration)
	}
	eventStream, err := openRuntimeEmailSubscription(deps.NATS, defaultRuntimeNATSOperations())
	if err != nil {
		return fail(err)
	}
	emailIntake, err := newRuntimeEmailIntake(eventStream.subscription, emailConsumer, clock, providers.reporter)
	if err != nil || !emailIntake.Start(ctx) {
		_ = eventStream.subscription.Unsubscribe()
		return fail(errRuntimeAdapterUnavailable)
	}
	shutdown.stopIntake = emailIntake.Close

	publisherStore, err := outbox.NewPostgresPublisherStore(deps.Postgres)
	if err != nil {
		return fail(errRuntimeConfiguration)
	}
	broker, err := outbox.NewNATSBroker(eventStream.publisher, cfg.DependencyTimeout)
	if err != nil {
		return fail(errRuntimeConfiguration)
	}
	publisher, err := outbox.NewPublisher(publisherStore, broker, clock, rand.Reader, 250*time.Millisecond, providers.reporter)
	if err != nil || !publisher.Start(ctx) {
		return fail(errRuntimeConfiguration)
	}
	shutdown.stopPublisher = publisher.Close

	outboxHealth, err := readiness.NewPostgresOutboxHealthProbe(deps.Postgres, clock)
	if err != nil {
		return fail(errRuntimeConfiguration)
	}
	checker, err := readiness.NewWithPolicy(cfg.DependencyTimeout, readiness.Policy{
		RedisDownAfterFailures: cfg.RedisDownAfterFailures, RedisRecoverAfterSuccesses: cfg.RedisRecoverAfterSuccesses,
		OutboxDegradedBacklog: cfg.OutboxDegradedBacklog, OutboxDownBacklog: cfg.OutboxDownBacklog,
		OutboxDegradedAge: cfg.OutboxDegradedAge, OutboxDownAge: cfg.OutboxDownAge,
	}, postgresprobe.Probe{Pool: deps.Postgres}, redisprobe.Probe{Client: deps.Redis}, runtimeOutboxHealthProbe{probe: outboxHealth, metrics: metrics})
	if err != nil {
		return fail(errRuntimeConfiguration)
	}
	mux := http.NewServeMux()
	apiHandler := controlapi.NewHandler(checker, apps, cfg.Security.RequestDeadline, rand.Reader)
	handler := controlapiv1.HandlerWithOptions(apiHandler, controlapiv1.StdHTTPServerOptions{
		BaseRouter: mux, ErrorHandlerFunc: apiHandler.GeneratedParameterError,
	})

	return &openedRuntime{handler: handler, close: shutdown.Close}, nil
}

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now().UTC() }

type localTestConfigSource struct{}

func (localTestConfigSource) Current(context.Context) (trust.TestConfigV1, error) {
	return trust.TestConfigV1{Message: "Talenro local fixture", Sequence: "1"}, nil
}

type runtimeOutboxHealthProbe struct {
	probe   readiness.OutboxHealthProbe
	metrics *observability.Registry
}

func (probe runtimeOutboxHealthProbe) Snapshot(ctx context.Context) (int64, time.Duration, error) {
	backlog, oldest, err := probe.probe.Snapshot(ctx)
	if err == nil {
		probe.metrics.SetOutbox(backlog, oldest)
	}
	return backlog, oldest, err
}

func validateRuntimeComposition(cfg config.Config) error {
	security := cfg.Security
	if security.Profile != config.ProfileLocal && security.Profile != config.ProfileTest && security.Profile != config.ProfileProduction {
		return errRuntimeConfiguration
	}
	if security.Profile == config.ProfileProduction {
		if security.EmailVerification == config.EmailDisabled ||
			security.SignerProvider != config.ProviderExternal ||
			security.FieldProtectorProvider != config.ProviderExternal ||
			security.EmailProvider != config.ProviderExternal ||
			security.ErrorReporterProvider != config.ProviderExternal ||
			!isRuntimeHTTPSURL(security.PublicBaseURL) {
			return errRuntimeConfiguration
		}
		for _, value := range security.BundleBaseURLs {
			if !isRuntimeHTTPSURL(value) {
				return errRuntimeConfiguration
			}
		}
		if len(security.WebAuthnOrigins) == 0 {
			return errRuntimeConfiguration
		}
		for _, value := range security.WebAuthnOrigins {
			if !isRuntimeHTTPSURL(value) {
				return errRuntimeConfiguration
			}
		}
		return errRuntimeAdapterUnavailable
	}

	if security.SignerProvider == config.ProviderExternal || security.FieldProtectorProvider == config.ProviderExternal ||
		security.EmailProvider == config.ProviderExternal || security.ErrorReporterProvider == config.ProviderExternal {
		return errRuntimeAdapterUnavailable
	}
	if security.SignerProvider != config.ProviderLocal || security.FieldProtectorProvider != config.ProviderLocal ||
		security.EmailProvider != config.ProviderLocal ||
		security.ErrorReporterProvider != config.ProviderLocal && security.ErrorReporterProvider != config.ProviderDiscard {
		return errRuntimeConfiguration
	}
	return nil
}

func isRuntimeHTTPSURL(value string) bool {
	parsed, err := url.ParseRequestURI(value)
	return err == nil && parsed.Scheme == "https" && parsed.Host != "" && parsed.User == nil && parsed.Opaque == "" &&
		(parsed.Path == "" || parsed.Path == "/") && parsed.RawQuery == "" && !parsed.ForceQuery && parsed.Fragment == ""
}

type runtimeProviders struct {
	protector    *sensitive.ObservedProtector
	email        *identity.LocalEmailSender
	reporter     *errorreport.AsyncReporter
	rootSigner   *trust.LocalRootSigner
	configSigner *trust.LocalConfigSigner
	signer       *trust.TimeoutConfigSigner
	crypto       *runtimeCryptoObserver

	closeOnce sync.Once
}

type runtimeCryptoObserver struct {
	metrics *observability.Registry
}

func (observer *runtimeCryptoObserver) ObserveSensitiveCrypto(event sensitive.CryptoEvent) {
	if observer == nil || observer.metrics == nil {
		return
	}
	switch event.Operation {
	case sensitive.CryptoOperationLookup, sensitive.CryptoOperationEncrypt, sensitive.CryptoOperationDecrypt:
	default:
		return
	}
	switch event.Result {
	case sensitive.CryptoResultSuccess, sensitive.CryptoResultFailure:
	default:
		return
	}
	switch event.Reason {
	case sensitive.CryptoReasonNone, sensitive.CryptoReasonInvalid, sensitive.CryptoReasonKeyUnavailable:
	default:
		return
	}
	if !validRuntimeCryptoOutcome(string(event.Result), string(event.Reason)) {
		return
	}
	observer.metrics.RecordCrypto(
		observability.CryptoOperation(event.Operation),
		observability.MetricResult(event.Result),
		observability.CryptoReason(event.Reason),
	)
}

func (observer *runtimeCryptoObserver) ObserveDeviceCrypto(event deviceauth.CryptoEvent) {
	if observer == nil || observer.metrics == nil {
		return
	}
	switch event.Operation {
	case deviceauth.CryptoOperationTokenVerify, deviceauth.CryptoOperationProofVerify:
	default:
		return
	}
	switch event.Result {
	case deviceauth.CryptoResultSuccess, deviceauth.CryptoResultFailure:
	default:
		return
	}
	switch event.Reason {
	case deviceauth.CryptoReasonNone, deviceauth.CryptoReasonInvalid, deviceauth.CryptoReasonKeyUnavailable:
	default:
		return
	}
	if !validRuntimeCryptoOutcome(string(event.Result), string(event.Reason)) {
		return
	}
	observer.metrics.RecordCrypto(
		observability.CryptoOperation(event.Operation),
		observability.MetricResult(event.Result),
		observability.CryptoReason(event.Reason),
	)
}

func (observer *runtimeCryptoObserver) ObserveTrustCrypto(event trust.CryptoEvent) {
	if observer == nil || observer.metrics == nil || event.Operation != trust.CryptoOperationBundleSign {
		return
	}
	switch event.Result {
	case trust.CryptoResultSuccess, trust.CryptoResultFailure:
	default:
		return
	}
	switch event.Reason {
	case trust.CryptoReasonNone, trust.CryptoReasonKeyUnavailable:
	default:
		return
	}
	if !validRuntimeCryptoOutcome(string(event.Result), string(event.Reason)) {
		return
	}
	observer.metrics.RecordCrypto(
		observability.CryptoOperationBundleSign,
		observability.MetricResult(event.Result),
		observability.CryptoReason(event.Reason),
	)
}

func validRuntimeCryptoOutcome(result, reason string) bool {
	return result == string(observability.MetricResultSuccess) && reason == string(observability.CryptoReasonNone) ||
		result == string(observability.MetricResultFailure) && reason != string(observability.CryptoReasonNone)
}

type runtimeIdentityRepository interface {
	identity.Repository
	identity.DeviceTransactionParticipant
}

type runtimeDeviceRepository interface {
	deviceauth.Repository
	identity.DeviceAuthorizationParticipant
}

type runtimeApplicationDependencies struct {
	identityRepository runtimeIdentityRepository
	deviceRepository   runtimeDeviceRepository
	limiter            ratelimit.Limiter
	identityChallenges identity.ChallengeStore
	deviceChallenges   deviceauth.ChallengeStore
	trustRepository    trust.Repository
	metadata           trust.RootMetadataV1
	testConfig         trust.TestConfigSource
	immutable          http.Handler
	clock              securitykit.Clock
}

func composeRuntimeApplications(
	cfg config.Config,
	providers *runtimeProviders,
	dependencies runtimeApplicationDependencies,
) (controlapi.Applications, error) {
	if providers == nil || providers.protector == nil || providers.signer == nil || providers.crypto == nil || dependencies.immutable == nil {
		return controlapi.Applications{}, errRuntimeConfiguration
	}
	identityApplication, err := identity.NewApplication(identity.ApplicationDependencies{
		Repository: dependencies.identityRepository, Protector: providers.protector,
		Random: rand.Reader, Clock: dependencies.clock, Limiter: dependencies.limiter,
		ChallengeStore: dependencies.identityChallenges, RateLimitKey: cfg.Security.SensitiveLookupKey,
		Security: cfg.Security, DeviceAuthorizationParticipant: dependencies.deviceRepository,
	})
	if err != nil {
		return controlapi.Applications{}, errRuntimeConfiguration
	}
	deviceApplication, err := deviceauth.NewApplication(deviceauth.ApplicationDependencies{
		Repository: dependencies.deviceRepository, IdentityParticipant: dependencies.identityRepository,
		Protector: providers.protector, Random: rand.Reader, Clock: dependencies.clock,
		Limiter: dependencies.limiter, ChallengeStore: dependencies.deviceChallenges,
		RateLimitKey: cfg.Security.SensitiveLookupKey, Security: cfg.Security,
	})
	if err != nil {
		return controlapi.Applications{}, errRuntimeConfiguration
	}
	observedDevice, err := deviceauth.NewObservedApplication(deviceApplication, providers.crypto)
	if err != nil {
		return controlapi.Applications{}, errRuntimeConfiguration
	}
	trustApplication, err := trust.NewApplication(trust.ApplicationDependencies{
		Repository: dependencies.trustRepository, Device: observedDevice, Signer: providers.signer,
		Metadata: dependencies.metadata, Random: rand.Reader, Clock: dependencies.clock,
		TestConfig: dependencies.testConfig, BundleBaseURLs: cfg.Security.BundleBaseURLs,
		RequestDeadline: cfg.Security.RequestDeadline,
	})
	if err != nil {
		return controlapi.Applications{}, errRuntimeConfiguration
	}
	return controlapi.Applications{
		Identity: identityApplication, StrongAuth: identityApplication, AccountAuth: identityApplication,
		Device: observedDevice, DeviceAuth: observedDevice, Trust: trustApplication,
		ImmutableBundle: dependencies.immutable,
	}, nil
}

func bootstrapLocalMetadata(
	ctx context.Context,
	repository trust.MetadataRepository,
	rootSigner *trust.LocalRootSigner,
	configSigner *trust.LocalConfigSigner,
	now time.Time,
) (trust.RootMetadataV1, error) {
	if ctx == nil || repository == nil || rootSigner == nil || configSigner == nil || now.IsZero() {
		return trust.RootMetadataV1{}, errRuntimeConfiguration
	}
	validFrom := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	validUntil := time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)
	if err := trust.EnsureLocalMetadata(ctx, repository, rootSigner, configSigner, validFrom, validUntil); err != nil {
		return trust.RootMetadataV1{}, errRuntimeConfiguration
	}
	records, err := repository.List(ctx)
	if err != nil || len(records) != 1 {
		return trust.RootMetadataV1{}, errRuntimeConfiguration
	}
	rootPublic := rootSigner.PublicKey()
	defer clear(rootPublic)
	payload, err := trust.VerifyRootMetadataV1(
		records[0], map[string]ed25519.PublicKey{rootSigner.KeyID(): rootPublic}, 0, now.UTC(),
	)
	if err != nil || payload.Version != "1" || len(payload.SigningKeys) != 1 ||
		payload.SigningKeys[0].KeyID != configSigner.KeyID() {
		return trust.RootMetadataV1{}, errRuntimeConfiguration
	}
	return payload, nil
}

func openRuntimeProviders(ctx context.Context, cfg config.Config, metrics *observability.Registry) (_ *runtimeProviders, resultErr error) {
	if validateRuntimeComposition(cfg) != nil || metrics == nil {
		return nil, errRuntimeConfiguration
	}
	providers := &runtimeProviders{crypto: &runtimeCryptoObserver{metrics: metrics}}
	defer func() {
		if resultErr != nil {
			providers.closeResources()
		}
	}()

	var err error
	localProtector, err := sensitive.NewLocal(cfg.Security.SensitiveLookupKey, cfg.Security.SensitiveEncryptionKey, 1)
	if err != nil {
		return nil, errRuntimeConfiguration
	}
	providers.protector, err = sensitive.NewObservedProtector(localProtector, providers.crypto)
	if err != nil {
		_ = localProtector.Close()
		return nil, errRuntimeConfiguration
	}
	providers.rootSigner, err = trust.NewLocalRootSigner(cfg.Security.LocalRootSigningSeed)
	if err != nil {
		return nil, errRuntimeConfiguration
	}
	providers.configSigner, err = trust.NewLocalConfigSigner(cfg.Security.LocalConfigSigningSeed)
	if err != nil {
		return nil, errRuntimeConfiguration
	}
	observedSigner, err := trust.NewObservedConfigSigner(providers.configSigner, providers.crypto)
	if err != nil {
		return nil, errRuntimeConfiguration
	}
	providers.signer, err = trust.NewTimeoutConfigSigner(observedSigner, cfg.Security.SignerTimeout)
	if err != nil {
		return nil, errRuntimeConfiguration
	}
	providers.email = identity.NewLocalEmailSender()
	providers.reporter, err = errorreport.New(
		ctx, errorreport.NewLocalProvider(), cfg.ErrorReportQueue, cfg.ErrorReportBatch,
		cfg.Security.ErrorReportTimeout, metrics,
	)
	if err != nil {
		return nil, errRuntimeConfiguration
	}
	return providers, nil
}

func (providers *runtimeProviders) Close(ctx context.Context) {
	if providers == nil {
		return
	}
	if providers.reporter != nil {
		providers.reporter.Close(ctx)
	}
	providers.closeResources()
}

func (providers *runtimeProviders) closeResources() {
	if providers == nil {
		return
	}
	providers.closeOnce.Do(func() {
		if providers.signer != nil {
			_ = providers.signer.Close()
		}
		if providers.configSigner != nil {
			_ = providers.configSigner.Close()
		}
		if providers.rootSigner != nil {
			_ = providers.rootSigner.Close()
		}
		if providers.protector != nil {
			_ = providers.protector.Close()
		}
	})
}

type runtimeShutdown struct {
	stopIntake        func(context.Context) error
	stopPublisher     func(context.Context) error
	drainReporter     func(context.Context)
	closeProviders    func()
	closeDependencies func()

	closeOnce sync.Once
	err       error
}

type runtimeEmailMessage interface {
	Data() []byte
	Ack() error
	Nak() error
}

type runtimeEmailSubscription interface {
	Next(context.Context) (runtimeEmailMessage, error)
	Unsubscribe() error
}

const (
	runtimeEventStreamName   = outbox.EventStreamName
	runtimeEmailConsumerName = "identity-email-delivery-v1"
	runtimeEmailRetryDelay   = 100 * time.Millisecond
	runtimeEmailTraceID      = "00000000000000000000000000000002"
)

type runtimeJetStream interface {
	outbox.JetStreamPublisher
	StreamInfo(string, ...nats.JSOpt) (*nats.StreamInfo, error)
	AddStream(*nats.StreamConfig, ...nats.JSOpt) (*nats.StreamInfo, error)
	PullSubscribe(string, string, ...nats.SubOpt) (*nats.Subscription, error)
}

type runtimeNATSOperations struct {
	jetStream      func(*nats.Conn) (runtimeJetStream, error)
	ensureStream   func(runtimeJetStream) error
	subscribeEmail func(runtimeJetStream) (runtimeEmailSubscription, error)
}

func defaultRuntimeNATSOperations() runtimeNATSOperations {
	return runtimeNATSOperations{
		jetStream: func(connection *nats.Conn) (runtimeJetStream, error) {
			if connection == nil {
				return nil, errRuntimeAdapterUnavailable
			}
			return connection.JetStream()
		},
		ensureStream: ensureRuntimeEventStream,
		subscribeEmail: func(jetStream runtimeJetStream) (runtimeEmailSubscription, error) {
			subscription, err := jetStream.PullSubscribe(
				contractevents.EmailDeliveryRequestedType,
				runtimeEmailConsumerName,
				nats.BindStream(runtimeEventStreamName),
				nats.ManualAck(),
				nats.AckExplicit(),
				nats.AckWait(30*time.Second),
				nats.MaxAckPending(100),
			)
			if err != nil {
				return nil, err
			}
			return &natsRuntimeSubscription{subscription: subscription}, nil
		},
	}
}

func ensureRuntimeEventStream(jetStream runtimeJetStream) error {
	if jetStream == nil {
		return errRuntimeAdapterUnavailable
	}
	info, err := jetStream.StreamInfo(runtimeEventStreamName)
	if errors.Is(err, nats.ErrStreamNotFound) {
		config := runtimeEventStreamConfig()
		info, err = jetStream.AddStream(&config)
	}
	if err != nil || !validRuntimeEventStream(info) {
		return errRuntimeAdapterUnavailable
	}
	return nil
}

func runtimeEventStreamConfig() nats.StreamConfig {
	return nats.StreamConfig{
		Name: runtimeEventStreamName, Subjects: []string{"talenro.>"}, Retention: nats.LimitsPolicy,
		Storage: nats.FileStorage, Discard: nats.DiscardOld, MaxMsgSize: 264 * 1024,
		Duplicates: 2 * time.Minute,
	}
}

func validRuntimeEventStream(info *nats.StreamInfo) bool {
	return info != nil && info.Config.Name == runtimeEventStreamName &&
		len(info.Config.Subjects) == 1 && info.Config.Subjects[0] == "talenro.>" &&
		info.Config.Retention == nats.LimitsPolicy && info.Config.Storage == nats.FileStorage &&
		info.Config.Discard == nats.DiscardOld && info.Config.MaxMsgSize == 264*1024 &&
		info.Config.Duplicates == 2*time.Minute
}

type openedRuntimeEventStream struct {
	subscription runtimeEmailSubscription
	publisher    outbox.JetStreamPublisher
}

func openRuntimeEmailSubscription(connection *nats.Conn, operations runtimeNATSOperations) (openedRuntimeEventStream, error) {
	if operations.jetStream == nil || operations.ensureStream == nil || operations.subscribeEmail == nil {
		return openedRuntimeEventStream{}, errRuntimeConfiguration
	}
	jetStream, err := operations.jetStream(connection)
	if err != nil {
		return openedRuntimeEventStream{}, errRuntimeAdapterUnavailable
	}
	if err := operations.ensureStream(jetStream); err != nil {
		return openedRuntimeEventStream{}, errRuntimeAdapterUnavailable
	}
	subscription, err := operations.subscribeEmail(jetStream)
	if err != nil || subscription == nil {
		return openedRuntimeEventStream{}, errRuntimeAdapterUnavailable
	}
	return openedRuntimeEventStream{subscription: subscription, publisher: jetStream}, nil
}

type natsRuntimeSubscription struct{ subscription *nats.Subscription }

func (subscription *natsRuntimeSubscription) Next(ctx context.Context) (runtimeEmailMessage, error) {
	if subscription == nil || subscription.subscription == nil || ctx == nil {
		return nil, errRuntimeAdapterUnavailable
	}
	messages, err := subscription.subscription.Fetch(1, nats.Context(ctx))
	if err != nil || len(messages) != 1 || messages[0] == nil {
		return nil, err
	}
	return natsRuntimeMessage{message: messages[0]}, nil
}

func (subscription *natsRuntimeSubscription) Unsubscribe() error {
	if subscription == nil || subscription.subscription == nil {
		return nil
	}
	return subscription.subscription.Unsubscribe()
}

type natsRuntimeMessage struct{ message *nats.Msg }

func (message natsRuntimeMessage) Data() []byte {
	if message.message == nil {
		return nil
	}
	return append([]byte(nil), message.message.Data...)
}

func (message natsRuntimeMessage) Ack() error {
	if message.message == nil {
		return errRuntimeAdapterUnavailable
	}
	return message.message.Ack()
}

func (message natsRuntimeMessage) Nak() error {
	if message.message == nil {
		return errRuntimeAdapterUnavailable
	}
	return message.message.Nak()
}

type runtimeEmailConsumer interface {
	Consume(context.Context, []byte, time.Time) error
}

type runtimeEmailIntake struct {
	subscription runtimeEmailSubscription
	consumer     runtimeEmailConsumer
	clock        securitykit.Clock
	reporter     errorreport.Reporter

	mu        sync.Mutex
	started   bool
	closed    bool
	cancel    context.CancelFunc
	done      chan struct{}
	closeOnce sync.Once
}

func newRuntimeEmailIntake(
	subscription runtimeEmailSubscription,
	consumer runtimeEmailConsumer,
	clock securitykit.Clock,
	reporter errorreport.Reporter,
) (*runtimeEmailIntake, error) {
	if subscription == nil || consumer == nil || clock == nil || reporter == nil {
		return nil, errRuntimeConfiguration
	}
	return &runtimeEmailIntake{subscription: subscription, consumer: consumer, clock: clock, reporter: reporter}, nil
}

func (intake *runtimeEmailIntake) Start(owner context.Context) bool {
	if intake == nil || owner == nil {
		return false
	}
	intake.mu.Lock()
	defer intake.mu.Unlock()
	if intake.started || intake.closed {
		return false
	}
	workerCtx, cancel := context.WithCancel(context.WithoutCancel(owner))
	intake.started = true
	intake.cancel = cancel
	intake.done = make(chan struct{})
	go intake.run(workerCtx)
	return true
}

func (intake *runtimeEmailIntake) run(ctx context.Context) {
	defer close(intake.done)
	for {
		if intake.stopping(ctx) {
			return
		}
		message, err := intake.subscription.Next(ctx)
		if err != nil {
			if intake.stopping(ctx) {
				return
			}
			intake.reportFailure(ctx)
			if !waitRuntimeEmailRetry(ctx) {
				return
			}
			continue
		}
		if intake.stopping(ctx) {
			return
		}
		if message == nil {
			intake.reportFailure(ctx)
			if !waitRuntimeEmailRetry(ctx) {
				return
			}
			continue
		}
		body := message.Data()
		if intake.stopping(ctx) {
			clear(body)
			return
		}
		now := intake.clock.Now().UTC()
		consumeErr := intake.consumer.Consume(ctx, body, now)
		clear(body)
		if intake.stopping(ctx) {
			return
		}
		if consumeErr != nil {
			intake.reportFailure(ctx)
			called, nakErr := intake.callWhileRunning(ctx, message.Nak)
			if !called {
				return
			}
			if nakErr != nil {
				intake.reportFailure(ctx)
			}
			continue
		}
		called, ackErr := intake.callWhileRunning(ctx, message.Ack)
		if !called {
			return
		}
		if ackErr != nil {
			intake.reportFailure(ctx)
		}
	}
}

func (intake *runtimeEmailIntake) stopping(ctx context.Context) bool {
	if intake == nil || ctx == nil || ctx.Err() != nil {
		return true
	}
	intake.mu.Lock()
	closed := intake.closed
	intake.mu.Unlock()
	return closed || ctx.Err() != nil
}

func (intake *runtimeEmailIntake) callWhileRunning(
	ctx context.Context,
	operation func() error,
) (called bool, err error) {
	if intake == nil || ctx == nil || operation == nil || ctx.Err() != nil {
		return false, nil
	}
	intake.mu.Lock()
	if intake.closed || ctx.Err() != nil {
		intake.mu.Unlock()
		return false, nil
	}
	called = true
	intake.mu.Unlock()
	defer func() {
		if recover() != nil {
			err = errRuntimeAdapterUnavailable
		}
	}()
	return called, operation()
}

func (intake *runtimeEmailIntake) reportFailure(ctx context.Context) {
	if intake == nil || intake.reporter == nil || ctx == nil {
		return
	}
	defer func() { _ = recover() }()
	_, _ = intake.callWhileRunning(ctx, func() error {
		intake.reporter.TryReport(errorreport.Report{
			Event: errorreport.EventWorkerFailure, Category: errorreport.CategoryDependency,
			Component: errorreport.ComponentEmail, Outcome: errorreport.OutcomeFailure,
			Fingerprint: errorreport.FingerprintDependency, BuildVersion: buildinfo.Current().Version,
			TraceID: runtimeEmailTraceID,
		})
		return nil
	})
}

func waitRuntimeEmailRetry(ctx context.Context) bool {
	timer := time.NewTimer(runtimeEmailRetryDelay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func (intake *runtimeEmailIntake) Close(ctx context.Context) error {
	if intake == nil {
		return nil
	}
	intake.closeOnce.Do(func() {
		intake.mu.Lock()
		intake.closed = true
		cancel, done := intake.cancel, intake.done
		intake.mu.Unlock()
		if cancel != nil {
			cancel()
		}
		_ = intake.subscription.Unsubscribe()
		if done != nil {
			if ctx == nil {
				<-done
			} else {
				select {
				case <-done:
				case <-ctx.Done():
				}
			}
		}
	})
	if ctx != nil && ctx.Err() != nil {
		return platform.NewCategorizedError(platform.CategoryInternal, errRuntimeConfiguration)
	}
	return nil
}

func (shutdown *runtimeShutdown) Close(ctx context.Context) error {
	if shutdown == nil {
		return nil
	}
	shutdown.closeOnce.Do(func() {
		if shutdown.stopIntake != nil {
			shutdown.err = shutdown.stopIntake(ctx)
		}
		if shutdown.stopPublisher != nil {
			if err := shutdown.stopPublisher(ctx); shutdown.err == nil {
				shutdown.err = err
			}
		}
		if shutdown.drainReporter != nil {
			shutdown.drainReporter(ctx)
		}
		if shutdown.closeProviders != nil {
			shutdown.closeProviders()
		}
		if shutdown.closeDependencies != nil {
			shutdown.closeDependencies()
		}
		if shutdown.err != nil {
			shutdown.err = platform.NewCategorizedError(platform.CategoryInternal, shutdown.err)
		}
	})
	return shutdown.err
}
