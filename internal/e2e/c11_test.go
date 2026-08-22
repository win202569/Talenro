//go:build e2e

package e2e

import (
	"bufio"
	"bytes"
	"context"
	"crypto/ecdh"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
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
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	controlapiv1 "talenro.local/platform/gen/go/talenro/controlapi/v1"
	"talenro.local/platform/internal/config"
	contractevents "talenro.local/platform/internal/contracts/events"
	"talenro.local/platform/internal/controlapi"
	"talenro.local/platform/internal/deviceauth"
	"talenro.local/platform/internal/errorreport"
	"talenro.local/platform/internal/idempotency"
	"talenro.local/platform/internal/identity"
	"talenro.local/platform/internal/observability"
	"talenro.local/platform/internal/outbox"
	"talenro.local/platform/internal/platform"
	postgresprobe "talenro.local/platform/internal/platform/postgres"
	redisprobe "talenro.local/platform/internal/platform/redis"
	"talenro.local/platform/internal/ratelimit"
	"talenro.local/platform/internal/readiness"
	"talenro.local/platform/internal/secret"
	"talenro.local/platform/internal/securitykit"
	"talenro.local/platform/internal/sensitive"
	"talenro.local/platform/internal/strictjson"
	"talenro.local/platform/internal/trust"
	"talenro.local/platform/internal/trustclient"
)

const (
	fixtureChildEnvironment     = "C11_E2E_FIXTURE_CHILD"
	fixtureChildValue           = "1"
	fixtureExternalRuntime      = "C11_E2E_EXTERNAL_RUNTIME"
	fixturePrimaryURL           = "C11_E2E_PRIMARY_URL"
	fixtureMirrorAURL           = "C11_E2E_MIRROR_A_URL"
	fixtureMirrorBURL           = "C11_E2E_MIRROR_B_URL"
	fixturePrimaryOrigin        = "C11_E2E_PRIMARY_ORIGIN"
	fixtureMirrorAOrigin        = "C11_E2E_MIRROR_A_ORIGIN"
	fixtureMirrorBOrigin        = "C11_E2E_MIRROR_B_ORIGIN"
	fixtureComposeProject       = "C11_E2E_COMPOSE_PROJECT"
	fixtureDatabase             = "C11_E2E_DATABASE_URL"
	fixtureRedis                = "C11_E2E_REDIS_ADDRESS"
	fixtureNATS                 = "C11_E2E_NATS_URL"
	fixtureConformancePath      = "C11_CONFORMANCE_BINARY"
	fixtureRootSeedName         = "TALENRO_LOCAL_ROOT_SIGNING_SEED_B64"
	fixtureConfigSeedName       = "TALENRO_LOCAL_CONFIG_SIGNING_SEED_B64"
	fixtureWebAuthnOrigin       = "TALENRO_WEBAUTHN_ORIGINS"
	fixtureWebAuthnRPID         = "TALENRO_WEBAUTHN_RP_ID"
	fixtureDefaultRootSeed      = "cm9vdC1zaWduaW5nLXNlZWQtZm9yLWxvY2FsLXRlc3Q"
	fixtureDefaultCfgSeed       = "Y29uZmlnLXNpZ25pbmctc2VlZC1sb2NhbC10ZXN0LTE"
	fixtureCommandLimit         = 32 << 10
	fixtureStageTimeout         = 2 * time.Minute
	fixtureRequestTimeout       = 15 * time.Second
	fixtureShutdownTimeout      = 5 * time.Second
	fixtureEmailProviderTimeout = 5 * time.Second
	fixtureNATSAckWait          = 6 * time.Second
	fixtureNATSAckTimeout       = time.Second
	fixtureNATSSettle           = 500 * time.Millisecond
	fixtureDatabaseURL          = "postgres://talenro:talenro_dev@127.0.0.1:5432/talenro?sslmode=disable" //nolint:gosec // Fixed loopback-only Docker fixture credential.
)

var fixtureKeySequence atomic.Uint64

var fixtureProjectPattern = regexp.MustCompile(`^talenro-c11-(?:verify|smoke)-[0-9a-f]{12,32}$`)
var fixtureDockerObjectPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

// TestMain turns the tagged test binary into a dependency-owning fixture only
// for an explicitly spawned child. Controls use inherited anonymous pipes; no
// fixture operation is reachable through the production router.
func TestMain(m *testing.M) {
	if os.Getenv(fixtureChildEnvironment) == fixtureChildValue {
		capture := newPrivacyCapture()
		slog.SetDefault(slog.New(slog.NewTextHandler(capture, nil)))
		os.Exit(runFixtureChild(os.Stdin, os.Stdout, capture))
	}
	os.Exit(m.Run())
}

func TestC11HappyPath(t *testing.T) {
	external, enabled, err := loadExternalRuntimeConfiguration(os.LookupEnv)
	if err != nil {
		t.Fatal("c11 external runtime contract invalid")
	}
	if enabled {
		fixture, cleanup, fixtureErr := newExternalC11Fixture(external, os.LookupEnv)
		if fixtureErr != nil {
			t.Fatal("c11 external runtime trust fixture unavailable")
		}
		t.Cleanup(cleanup)
		fixture.runExternalHappyPath(t)
		return
	}
	fixture := requireC11Fixture(t)
	fixture.runHappyPath(t)
}

type c11Fixture struct {
	repoRoot            string
	compose             *composeFixture
	child               *fixtureChild
	authorityNow        func() (time.Time, error)
	ready               fixtureReady
	client              *http.Client
	webAuthnOrigin      string
	webAuthnRPID        string
	deviceProofAudience string
	expectedSourceURLs  [3]string
	externalRuntime     bool
}

type composeFixture struct {
	repoRoot string
	project  string
	mu       sync.Mutex
}

type fixtureChild struct {
	command         *exec.Cmd
	input           io.WriteCloser
	output          io.ReadCloser
	encoder         *json.Encoder
	decoder         *json.Decoder
	cancel          context.CancelFunc
	mu              sync.Mutex
	cancelOnce      sync.Once
	wait            chan error
	closed          bool
	shutdownTimeout time.Duration
}

type fixtureReady struct {
	Kind        string `json:"kind"`
	RequiredURL string `json:"required_url"`
	GraceURL    string `json:"grace_url"`
	DisabledURL string `json:"disabled_url"`
	MirrorAURL  string `json:"mirror_a_url"`
	MirrorBURL  string `json:"mirror_b_url"`
	MetricsURL  string `json:"metrics_url"`
	RootKeyID   string `json:"root_key_id"`
	RootPublic  string `json:"root_public"`
	Metadata    string `json:"metadata"`
}

type externalRuntimeConfiguration struct {
	PrimaryURL string
	MirrorAURL string
	MirrorBURL string
	Origins    [3]string
}

type externalDependencyConfiguration struct {
	Project      string
	DatabaseURL  string
	RedisAddress string
	NATSURL      string
}

type fixtureRequest struct {
	ID        uint64 `json:"id"`
	Operation string `json:"operation"`
	Value     string `json:"value,omitempty"`
	Recipient string `json:"recipient,omitempty"`
	Template  string `json:"template,omitempty"`
}

type fixtureResponse struct {
	ID    uint64 `json:"id"`
	OK    bool   `json:"ok"`
	Value string `json:"value,omitempty"`
	Count int    `json:"count,omitempty"`
}

func requireC11Fixture(t *testing.T) *c11Fixture {
	t.Helper()
	dependencies, enabled, err := loadExternalDependencyConfiguration(os.LookupEnv)
	if err != nil || !enabled {
		t.Fatal("c11 e2e fixture unavailable: explicit dependency ownership contract is required")
	}
	if _, err := exec.LookPath("docker"); err != nil {
		t.Fatal("c11 e2e fixture unavailable: docker executable is required")
	}
	repoRoot := e2eRepositoryRoot(t)
	compose := &composeFixture{repoRoot: repoRoot, project: dependencies.Project}
	child, ready, err := startFixtureChild(t.Context(), repoRoot)
	if err != nil {
		t.Fatal("c11 e2e fixture unavailable: test runtime failed")
	}
	t.Cleanup(func() {
		if err := child.close(); err != nil {
			t.Error("c11 e2e cleanup failed: fixture child")
		}
	})
	return &c11Fixture{
		repoRoot: repoRoot, compose: compose, child: child, ready: ready,
		authorityNow: func() (time.Time, error) {
			response, callErr := child.call("clock-now", "", "", "")
			if callErr != nil {
				return time.Time{}, callErr
			}
			now, parseErr := time.Parse(time.RFC3339, response.Value)
			if parseErr != nil || now.Location() != time.UTC || now.Format(time.RFC3339) != response.Value {
				return time.Time{}, errors.New("fixture authority clock invalid")
			}
			return now, nil
		},
		webAuthnOrigin: configuredOriginForListener(ready.RequiredURL), webAuthnRPID: "localhost",
		deviceProofAudience: configuredOriginForListener(ready.RequiredURL),
		expectedSourceURLs: [3]string{
			configuredOriginForListener(ready.RequiredURL),
			configuredOriginForListener(ready.MirrorAURL),
			configuredOriginForListener(ready.MirrorBURL),
		},
		client: &http.Client{Timeout: fixtureRequestTimeout, CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		}},
	}
}

type memoryMetadataRepository struct {
	records []trust.SignedRootMetadataV1
}

func (repository *memoryMetadataRepository) List(ctx context.Context) ([]trust.SignedRootMetadataV1, error) {
	if repository == nil || ctx == nil || ctx.Err() != nil {
		return nil, errors.New("metadata repository unavailable")
	}
	return append([]trust.SignedRootMetadataV1(nil), repository.records...), nil
}

func (repository *memoryMetadataRepository) Publish(ctx context.Context, value trust.SignedRootMetadataV1) error {
	if repository == nil || ctx == nil || ctx.Err() != nil || len(repository.records) != 0 {
		return errors.New("metadata publication rejected")
	}
	repository.records = append(repository.records, value)
	return nil
}

func newExternalC11Fixture(
	runtimeConfiguration externalRuntimeConfiguration,
	lookup func(string) (string, bool),
) (*c11Fixture, func(), error) {
	trustConfiguration, err := loadExternalTrustConfiguration(lookup)
	if err != nil {
		return nil, func() {}, errors.New("external runtime configuration unavailable")
	}
	defer trustConfiguration.rootSeed.Clear()
	defer trustConfiguration.configSeed.Clear()
	rootSigner, err := trust.NewLocalRootSigner(trustConfiguration.rootSeed)
	if err != nil {
		return nil, func() {}, errors.New("external root trust unavailable")
	}
	configSigner, err := trust.NewLocalConfigSigner(trustConfiguration.configSeed)
	if err != nil {
		_ = rootSigner.Close()
		return nil, func() {}, errors.New("external config trust unavailable")
	}
	repository := &memoryMetadataRepository{}
	metadataContext, metadataCancel := context.WithTimeout(context.Background(), fixtureRequestTimeout)
	defer metadataCancel()
	_, signedMetadata, err := fixtureMetadata(metadataContext, repository, rootSigner, configSigner, time.Now().UTC())
	if err != nil {
		_ = configSigner.Close()
		_ = rootSigner.Close()
		return nil, func() {}, err
	}
	metadataBytes, err := json.Marshal(signedMetadata)
	if err != nil {
		_ = configSigner.Close()
		_ = rootSigner.Close()
		return nil, func() {}, errors.New("external metadata encoding failed")
	}
	defer clear(metadataBytes)
	rootPublic := rootSigner.PublicKey()
	defer clear(rootPublic)
	ready := fixtureReady{
		Kind: "ready", RequiredURL: runtimeConfiguration.PrimaryURL, GraceURL: runtimeConfiguration.PrimaryURL,
		DisabledURL: runtimeConfiguration.PrimaryURL, MirrorAURL: runtimeConfiguration.MirrorAURL, MirrorBURL: runtimeConfiguration.MirrorBURL,
		RootKeyID: rootSigner.KeyID(), RootPublic: base64.RawURLEncoding.EncodeToString(rootPublic),
		Metadata: base64.RawURLEncoding.EncodeToString(metadataBytes),
	}
	client := &http.Client{Timeout: fixtureRequestTimeout, CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	var closeOnce sync.Once
	cleanup := func() {
		closeOnce.Do(func() {
			client.CloseIdleConnections()
			_ = configSigner.Close()
			_ = rootSigner.Close()
		})
	}
	return &c11Fixture{
		ready: ready, client: client, webAuthnOrigin: trustConfiguration.webAuthnOrigin, webAuthnRPID: trustConfiguration.webAuthnRPID,
		deviceProofAudience: trustConfiguration.webAuthnOrigin,
		expectedSourceURLs:  runtimeConfiguration.Origins,
		externalRuntime:     true,
		authorityNow: func() (time.Time, error) {
			return time.Now().UTC(), nil
		},
	}, cleanup, nil
}

type externalTrustConfiguration struct {
	rootSeed       secret.Bytes
	configSeed     secret.Bytes
	webAuthnOrigin string
	webAuthnRPID   string
}

func loadExternalTrustConfiguration(lookup func(string) (string, bool)) (externalTrustConfiguration, error) {
	if lookup == nil {
		return externalTrustConfiguration{}, errors.New("external trust lookup unavailable")
	}
	rootEncoded := fixtureDefaultRootSeed
	if value, exists := lookup(fixtureRootSeedName); exists {
		rootEncoded = value
	}
	configEncoded := fixtureDefaultCfgSeed
	if value, exists := lookup(fixtureConfigSeedName); exists {
		configEncoded = value
	}
	rootBytes, rootErr := base64.RawURLEncoding.DecodeString(rootEncoded)
	configBytes, configErr := base64.RawURLEncoding.DecodeString(configEncoded)
	if rootErr != nil || configErr != nil || len(rootBytes) != ed25519.SeedSize || len(configBytes) != ed25519.SeedSize {
		clear(rootBytes)
		clear(configBytes)
		return externalTrustConfiguration{}, errors.New("external trust seed invalid")
	}
	configuration := externalTrustConfiguration{rootSeed: secret.NewBytes(rootBytes), configSeed: secret.NewBytes(configBytes)}
	clear(rootBytes)
	clear(configBytes)
	configuration.webAuthnOrigin = "http://localhost:8080"
	if value, exists := lookup(fixtureWebAuthnOrigin); exists {
		origins := strings.Split(value, ",")
		if len(origins) != 1 {
			configuration.rootSeed.Clear()
			configuration.configSeed.Clear()
			return externalTrustConfiguration{}, errors.New("external WebAuthn origin invalid")
		}
		configuration.webAuthnOrigin = origins[0]
	}
	configuration.webAuthnRPID = "localhost"
	if value, exists := lookup(fixtureWebAuthnRPID); exists {
		configuration.webAuthnRPID = value
	}
	origin, err := url.ParseRequestURI(configuration.webAuthnOrigin)
	if err != nil || origin.Scheme != "http" || origin.Hostname() != "localhost" || origin.Port() != "8080" || origin.User != nil ||
		origin.Path != "" || origin.RawQuery != "" || origin.Fragment != "" || configuration.webAuthnRPID != "localhost" {
		configuration.rootSeed.Clear()
		configuration.configSeed.Clear()
		return externalTrustConfiguration{}, errors.New("external WebAuthn contract invalid")
	}
	return configuration, nil
}

func loadExternalRuntimeConfiguration(lookup func(string) (string, bool)) (externalRuntimeConfiguration, bool, error) {
	if lookup == nil {
		return externalRuntimeConfiguration{}, false, errors.New("external runtime lookup unavailable")
	}
	flag, exists := lookup(fixtureExternalRuntime)
	if !exists || flag == "" {
		return externalRuntimeConfiguration{}, false, nil
	}
	if flag != fixtureChildValue {
		return externalRuntimeConfiguration{}, true, errors.New("external runtime flag invalid")
	}
	configuration := externalRuntimeConfiguration{}
	configuration.PrimaryURL, _ = lookup(fixturePrimaryURL)
	configuration.MirrorAURL, _ = lookup(fixtureMirrorAURL)
	configuration.MirrorBURL, _ = lookup(fixtureMirrorBURL)
	configuration.Origins[0], _ = lookup(fixturePrimaryOrigin)
	configuration.Origins[1], _ = lookup(fixtureMirrorAOrigin)
	configuration.Origins[2], _ = lookup(fixtureMirrorBOrigin)
	parsed := make([]*url.URL, 0, 3)
	for _, raw := range []string{configuration.PrimaryURL, configuration.MirrorAURL, configuration.MirrorBURL} {
		endpoint, err := parseOwnedLoopbackURL(raw, "http")
		if err != nil {
			return externalRuntimeConfiguration{}, true, errors.New("external runtime endpoint invalid")
		}
		parsed = append(parsed, endpoint)
	}
	ports := map[string]struct{}{}
	for _, endpoint := range parsed {
		if _, duplicate := ports[endpoint.Port()]; duplicate {
			return externalRuntimeConfiguration{}, true, errors.New("external runtime endpoints overlap")
		}
		ports[endpoint.Port()] = struct{}{}
	}
	for index, raw := range configuration.Origins {
		if _, err := parseConfiguredBundleOrigin(raw, parsed[index].Port()); err != nil {
			return externalRuntimeConfiguration{}, true, errors.New("external runtime origin invalid")
		}
	}
	return configuration, true, nil
}

func parseConfiguredBundleOrigin(raw, listenerPort string) (*url.URL, error) {
	parsed, err := url.ParseRequestURI(raw)
	if err != nil || parsed.Scheme != "http" || parsed.Hostname() != "localhost" || parsed.Port() != listenerPort ||
		parsed.User != nil || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.ForceQuery ||
		raw != "http://localhost:"+listenerPort || !validFixturePort(listenerPort) {
		return nil, errors.New("configured bundle origin invalid")
	}
	return parsed, nil
}

func configuredOriginForListener(listener string) string {
	parsed, err := parseOwnedLoopbackURL(listener, "http")
	if err != nil {
		return ""
	}
	return "http://localhost:" + parsed.Port()
}

func loadExternalDependencyConfiguration(lookup func(string) (string, bool)) (externalDependencyConfiguration, bool, error) {
	if lookup == nil {
		return externalDependencyConfiguration{}, false, errors.New("external dependency lookup unavailable")
	}
	names := [...]string{fixtureComposeProject, fixtureDatabase, fixtureRedis, fixtureNATS}
	present := 0
	for _, name := range names {
		if _, exists := lookup(name); exists {
			present++
		}
	}
	if present == 0 {
		return externalDependencyConfiguration{}, false, nil
	}
	if present != len(names) {
		return externalDependencyConfiguration{}, true, errors.New("external dependency ownership contract incomplete")
	}
	configuration := externalDependencyConfiguration{}
	configuration.Project, _ = lookup(fixtureComposeProject)
	configuration.DatabaseURL, _ = lookup(fixtureDatabase)
	configuration.RedisAddress, _ = lookup(fixtureRedis)
	configuration.NATSURL, _ = lookup(fixtureNATS)
	if !fixtureProjectPattern.MatchString(configuration.Project) || !validFixtureDatabaseURL(configuration.DatabaseURL) ||
		!validLoopbackAddress(configuration.RedisAddress) {
		return externalDependencyConfiguration{}, true, errors.New("external dependency ownership contract invalid")
	}
	if _, err := parseOwnedLoopbackURL(configuration.NATSURL, "nats"); err != nil {
		return externalDependencyConfiguration{}, true, errors.New("external dependency ownership contract invalid")
	}
	return configuration, true, nil
}

func parseOwnedLoopbackURL(raw, scheme string) (*url.URL, error) {
	parsed, err := url.ParseRequestURI(raw)
	if err != nil || parsed.Scheme != scheme || parsed.Hostname() != "127.0.0.1" || parsed.User != nil || parsed.Port() == "" ||
		parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.ForceQuery || !validFixturePort(parsed.Port()) {
		return nil, errors.New("owned loopback URL invalid")
	}
	return parsed, nil
}

func validFixtureDatabaseURL(raw string) bool {
	parsed, err := url.ParseRequestURI(raw)
	if err != nil || parsed.Scheme != "postgres" || parsed.Hostname() != "127.0.0.1" || parsed.Port() == "" ||
		parsed.User == nil || parsed.User.Username() == "" || parsed.Path != "/talenro" || parsed.Fragment != "" || !validFixturePort(parsed.Port()) {
		return false
	}
	query := parsed.Query()
	return len(query) == 1 && query.Get("sslmode") == "disable"
}

func validLoopbackAddress(raw string) bool {
	host, port, err := net.SplitHostPort(raw)
	return err == nil && host == "127.0.0.1" && validFixturePort(port)
}

func validFixturePort(raw string) bool {
	port, err := strconv.Atoi(raw)
	return err == nil && port > 0 && port <= 65535 && strconv.Itoa(port) == raw
}

func e2eRepositoryRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("c11 e2e fixture unavailable: repository resolution failed")
	}
	root, err := filepath.Abs(filepath.Join(filepath.Dir(filename), "..", ".."))
	if err != nil {
		t.Fatal("c11 e2e fixture unavailable: repository resolution failed")
	}
	return root
}

func (fixture *composeFixture) run(ctx context.Context, arguments ...string) error {
	fixture.mu.Lock()
	defer fixture.mu.Unlock()
	base := []string{"compose", "-p", fixture.project, "-f", filepath.Join(fixture.repoRoot, "deploy", "dev", "compose.yaml")}
	return runBoundedCommand(ctx, fixture.repoRoot, "docker", append(base, arguments...)...)
}

type fixtureNetworkFaultTarget struct {
	containerID string
	networkID   string
}

func (fixture *composeFixture) disconnectNetwork(ctx context.Context, service string) (fixtureNetworkFaultTarget, error) {
	if fixture == nil {
		return fixtureNetworkFaultTarget{}, errors.New("c11 e2e external stage failed")
	}
	fixture.mu.Lock()
	defer fixture.mu.Unlock()
	if !fixtureProjectPattern.MatchString(fixture.project) || !validFixtureFaultService(service) {
		return fixtureNetworkFaultTarget{}, errors.New("c11 e2e external stage failed")
	}

	base := []string{"compose", "-p", fixture.project, "-f", filepath.Join(fixture.repoRoot, "deploy", "dev", "compose.yaml")}
	containerID, err := captureBoundedCommand(ctx, fixture.repoRoot, "docker", append(base, "ps", "--quiet", "--no-trunc", service)...)
	if err != nil || !fixtureDockerObjectPattern.MatchString(containerID) {
		return fixtureNetworkFaultTarget{}, errors.New("c11 e2e external stage failed")
	}
	projectLabel, err := captureBoundedCommand(ctx, fixture.repoRoot, "docker", "inspect", "--type", "container", "--format",
		`{{ index .Config.Labels "com.docker.compose.project" }}`, containerID)
	if err != nil || projectLabel != fixture.project {
		return fixtureNetworkFaultTarget{}, errors.New("c11 e2e external stage failed")
	}
	serviceLabel, err := captureBoundedCommand(ctx, fixture.repoRoot, "docker", "inspect", "--type", "container", "--format",
		`{{ index .Config.Labels "com.docker.compose.service" }}`, containerID)
	if err != nil || serviceLabel != service {
		return fixtureNetworkFaultTarget{}, errors.New("c11 e2e external stage failed")
	}

	networkID, err := captureBoundedCommand(ctx, fixture.repoRoot, "docker", "network", "ls", "--quiet", "--no-trunc",
		"--filter", "label=com.docker.compose.project="+fixture.project)
	if err != nil || !fixtureDockerObjectPattern.MatchString(networkID) {
		return fixtureNetworkFaultTarget{}, errors.New("c11 e2e external stage failed")
	}
	networkProjectLabel, err := captureBoundedCommand(ctx, fixture.repoRoot, "docker", "network", "inspect", "--format",
		`{{ index .Labels "com.docker.compose.project" }}`, networkID)
	if err != nil || networkProjectLabel != fixture.project {
		return fixtureNetworkFaultTarget{}, errors.New("c11 e2e external stage failed")
	}
	networkNameLabel, err := captureBoundedCommand(ctx, fixture.repoRoot, "docker", "network", "inspect", "--format",
		`{{ index .Labels "com.docker.compose.network" }}`, networkID)
	if err != nil || networkNameLabel != "default" {
		return fixtureNetworkFaultTarget{}, errors.New("c11 e2e external stage failed")
	}

	target := fixtureNetworkFaultTarget{containerID: containerID, networkID: networkID}
	if err := runBoundedCommand(ctx, fixture.repoRoot, "docker", "network", "disconnect", target.networkID, target.containerID); err != nil {
		return fixtureNetworkFaultTarget{}, err
	}
	return target, nil
}

func (fixture *composeFixture) reconnectNetwork(ctx context.Context, target fixtureNetworkFaultTarget) error {
	if fixture == nil || !fixtureDockerObjectPattern.MatchString(target.containerID) || !fixtureDockerObjectPattern.MatchString(target.networkID) {
		return errors.New("c11 e2e external stage failed")
	}
	fixture.mu.Lock()
	defer fixture.mu.Unlock()
	return runBoundedCommand(ctx, fixture.repoRoot, "docker", "network", "connect", target.networkID, target.containerID)
}

func validFixtureFaultService(service string) bool {
	switch service {
	case "postgres", "redis", "nats":
		return true
	default:
		return false
	}
}

type boundedFixtureCommandOutput struct {
	buffer   bytes.Buffer
	overflow bool
}

func (output *boundedFixtureCommandOutput) Write(value []byte) (int, error) {
	if output.overflow || len(value) > fixtureCommandLimit-output.buffer.Len() {
		output.overflow = true
		return len(value), nil
	}
	_, _ = output.buffer.Write(value)
	return len(value), nil
}

func captureBoundedCommand(parent context.Context, directory, name string, arguments ...string) (string, error) {
	ctx, cancel := context.WithTimeout(parent, fixtureStageTimeout)
	defer cancel()
	output := new(boundedFixtureCommandOutput)
	command := exec.CommandContext(ctx, name, arguments...) //nolint:gosec // Fixed local tools and validated fixture arguments only.
	command.Dir = directory
	command.Stdout = output
	command.Stderr = io.Discard
	if err := command.Run(); err != nil || output.overflow {
		return "", errors.New("c11 e2e external stage failed")
	}
	return strings.TrimSpace(output.buffer.String()), nil
}

func runBoundedCommand(parent context.Context, directory, name string, arguments ...string) error {
	ctx, cancel := context.WithTimeout(parent, fixtureStageTimeout)
	defer cancel()
	command := exec.CommandContext(ctx, name, arguments...) //nolint:gosec // Fixed local tools and validated fixture arguments only.
	command.Dir = directory
	command.Stdout = io.Discard
	command.Stderr = io.Discard
	if err := command.Run(); err != nil {
		return errors.New("c11 e2e external stage failed")
	}
	return nil
}

func startFixtureChild(startupContext context.Context, repoRoot string) (*fixtureChild, fixtureReady, error) {
	executable, err := os.Executable()
	if err != nil {
		return nil, fixtureReady{}, errors.New("fixture executable unavailable")
	}
	lifecycleContext, lifecycleCancel := context.WithCancel(context.WithoutCancel(startupContext))
	// The child must outlive the startup/test context so cleanup can first request
	// graceful pipe shutdown, then cancel only this dedicated lifecycle.
	command := exec.CommandContext(lifecycleContext, executable) //nolint:gosec // Exact current tagged test binary.
	command.WaitDelay = fixtureRequestTimeout
	command.Dir = repoRoot
	command.Env = fixtureChildEnvironmentValues(os.Environ())
	input, err := command.StdinPipe()
	if err != nil {
		lifecycleCancel()
		return nil, fixtureReady{}, errors.New("fixture input unavailable")
	}
	output, err := command.StdoutPipe()
	if err != nil {
		_ = input.Close()
		lifecycleCancel()
		return nil, fixtureReady{}, errors.New("fixture output unavailable")
	}
	command.Stderr = io.Discard
	if err := command.Start(); err != nil {
		_ = input.Close()
		_ = output.Close()
		lifecycleCancel()
		return nil, fixtureReady{}, errors.New("fixture start failed")
	}
	child := &fixtureChild{
		command: command, input: input, output: output, encoder: json.NewEncoder(input),
		decoder: json.NewDecoder(bufio.NewReaderSize(output, fixtureCommandLimit)), cancel: lifecycleCancel,
		wait: make(chan error, 1), shutdownTimeout: fixtureShutdownTimeout,
	}
	go func() { child.wait <- command.Wait() }()
	type readinessResult struct {
		ready fixtureReady
		err   error
	}
	readiness := make(chan readinessResult, 1)
	go func() {
		var ready fixtureReady
		err := child.decoder.Decode(&ready)
		readiness <- readinessResult{ready: ready, err: err}
	}()
	timer := time.NewTimer(fixtureRequestTimeout)
	defer timer.Stop()
	var result readinessResult
	select {
	case result = <-readiness:
	case <-startupContext.Done():
		_ = child.forceClose()
		return nil, fixtureReady{}, errors.New("fixture readiness canceled")
	case <-timer.C:
		_ = child.forceClose()
		return nil, fixtureReady{}, errors.New("fixture readiness deadline exceeded")
	}
	if result.err != nil || result.ready.Kind != "ready" || !validFixtureReady(result.ready) {
		_ = child.forceClose()
		return nil, fixtureReady{}, errors.New("fixture readiness failed")
	}
	return child, result.ready, nil
}

func fixtureChildEnvironmentValues(values []string) []string {
	result := make([]string, 0, len(values)+1)
	explicit := make(map[string]string, 4)
	for _, value := range values {
		name, field, found := strings.Cut(value, "=")
		upper := strings.ToUpper(name)
		if !found {
			continue
		}
		if strings.HasPrefix(upper, "TALENRO_") || strings.HasPrefix(upper, "C11_") {
			switch upper {
			case fixtureComposeProject, fixtureDatabase, fixtureRedis, fixtureNATS:
				explicit[upper] = field
			}
			continue
		}
		result = append(result, value)
	}
	for _, name := range []string{fixtureComposeProject, fixtureDatabase, fixtureRedis, fixtureNATS} {
		if value, exists := explicit[name]; exists {
			result = append(result, name+"="+value)
		}
	}
	return append(result, fixtureChildEnvironment+"="+fixtureChildValue)
}

func validFixtureReady(ready fixtureReady) bool {
	for _, value := range []string{ready.RequiredURL, ready.GraceURL, ready.DisabledURL, ready.MirrorAURL, ready.MirrorBURL, ready.MetricsURL} {
		if !strings.HasPrefix(value, "http://127.0.0.1:") {
			return false
		}
	}
	root, rootErr := base64.RawURLEncoding.DecodeString(ready.RootPublic)
	metadata, metadataErr := base64.RawURLEncoding.DecodeString(ready.Metadata)
	return ready.RootKeyID != "" && rootErr == nil && len(root) == ed25519.PublicKeySize && metadataErr == nil && len(metadata) >= 2
}

func (child *fixtureChild) call(operation, value, recipient, template string) (fixtureResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), fixtureRequestTimeout)
	defer cancel()
	return child.callContext(ctx, operation, value, recipient, template)
}

func (child *fixtureChild) callContext(ctx context.Context, operation, value, recipient, template string) (fixtureResponse, error) {
	boundedContext, cancel := context.WithTimeout(ctx, fixtureRequestTimeout)
	defer cancel()
	child.mu.Lock()
	defer child.mu.Unlock()
	if child.closed {
		return fixtureResponse{}, errors.New("fixture closed")
	}
	id := fixtureKeySequence.Add(1)
	type callResult struct {
		response fixtureResponse
		err      error
	}
	result := make(chan callResult, 1)
	go func() {
		if err := child.encoder.Encode(fixtureRequest{ID: id, Operation: operation, Value: value, Recipient: recipient, Template: template}); err != nil {
			result <- callResult{err: err}
			return
		}
		var response fixtureResponse
		err := child.decoder.Decode(&response)
		result <- callResult{response: response, err: err}
	}()
	select {
	case completed := <-result:
		if completed.err != nil || completed.response.ID != id || !completed.response.OK {
			return fixtureResponse{}, errors.New("fixture response failed")
		}
		return completed.response, nil
	case <-boundedContext.Done():
		child.closed = true
		child.abort()
		return fixtureResponse{}, errors.New("fixture response deadline exceeded")
	}
}

func (child *fixtureChild) close() error {
	child.mu.Lock()
	if child.closed {
		child.mu.Unlock()
		_ = child.forceClose()
		return nil
	}
	child.closed = true
	id := fixtureKeySequence.Add(1)
	encoded := make(chan error, 1)
	go func() { encoded <- child.encoder.Encode(fixtureRequest{ID: id, Operation: "shutdown"}) }()
	timeout := child.shutdownTimeout
	if timeout <= 0 {
		timeout = fixtureRequestTimeout
	}
	timer := time.NewTimer(timeout)
	select {
	case encodeErr := <-encoded:
		_ = child.input.Close()
		child.mu.Unlock()
		if encodeErr != nil {
			_ = child.forceClose()
			return errors.New("fixture shutdown request failed")
		}
	case <-timer.C:
		child.mu.Unlock()
		_ = child.forceClose()
		return errors.New("fixture shutdown request deadline exceeded")
	}
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
	timer.Reset(timeout)
	select {
	case err := <-child.wait:
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		_ = child.output.Close()
		child.cancelLifecycle()
		return err
	case <-timer.C:
		_ = child.forceClose()
		return errors.New("fixture shutdown deadline exceeded")
	}
}

func (child *fixtureChild) forceClose() error {
	if child == nil {
		return nil
	}
	child.abort()
	if child.wait == nil {
		return nil
	}
	timeout := child.shutdownTimeout
	if timeout <= 0 {
		timeout = fixtureRequestTimeout
	}
	select {
	case <-child.wait:
		return nil
	case <-time.After(timeout):
		return errors.New("fixture process termination deadline exceeded")
	}
}

func (child *fixtureChild) abort() {
	if child == nil {
		return
	}
	_ = child.input.Close()
	_ = child.output.Close()
	child.cancelLifecycle()
	if child.command != nil && child.command.Process != nil {
		_ = child.command.Process.Kill()
	}
}

func (child *fixtureChild) cancelLifecycle() {
	if child == nil || child.cancel == nil {
		return
	}
	child.cancelOnce.Do(child.cancel)
}

type fixtureRuntime struct {
	cancel              context.CancelFunc
	servers             []*http.Server
	listeners           []net.Listener
	dependencies        *platform.Dependencies
	publisher           *outbox.Publisher
	subscription        *nats.Subscription
	reporter            *errorreport.AsyncReporter
	protector           *sensitive.Local
	rootSigner          *trust.LocalRootSigner
	configSigner        *trust.LocalConfigSigner
	timeoutSigner       *trust.TimeoutConfigSigner
	email               *controlledEmailSender
	reports             *controlledReportProvider
	reportFull          atomic.Int64
	deliveryFetches     atomic.Int32
	deliveryAckFailures atomic.Int32
	signer              *controlledConfigSigner
	testConfig          *controlledTestConfig
	sources             map[string]*restartableSource
	ready               fixtureReady
	privacy             *privacyCapture
	clock               *fixtureClock
}

type fixtureClock struct {
	mu      sync.RWMutex
	now     time.Time
	maximum time.Time
}

func newFixtureClock() *fixtureClock {
	now := time.Now().UTC().Truncate(time.Second)
	return &fixtureClock{now: now, maximum: now.Add(48 * time.Hour)}
}

func (clock *fixtureClock) Now() time.Time {
	if clock == nil {
		return time.Time{}
	}
	clock.mu.RLock()
	defer clock.mu.RUnlock()
	return clock.now
}

func (clock *fixtureClock) set(raw string) bool {
	if clock == nil {
		return false
	}
	next, err := time.Parse(time.RFC3339, raw)
	if err != nil || next.Location() != time.UTC || next.Format(time.RFC3339) != raw {
		return false
	}
	clock.mu.Lock()
	defer clock.mu.Unlock()
	if next.Before(clock.now) || next.After(clock.maximum) {
		return false
	}
	clock.now = next
	return true
}

func (fixtureRuntime *fixtureRuntime) outboxClock() securitykit.Clock {
	if fixtureRuntime == nil || fixtureRuntime.clock == nil {
		return nil
	}
	return fixtureRuntime.clock
}

type controlledTestConfig struct {
	mu       sync.RWMutex
	sequence string
}

func (source *controlledTestConfig) Current(context.Context) (trust.TestConfigV1, error) {
	if source == nil {
		return trust.TestConfigV1{}, errors.New("e2e test config rejected")
	}
	source.mu.RLock()
	defer source.mu.RUnlock()
	return trust.TestConfigV1{Message: "C1.1 E2E fixture", Sequence: source.sequence}, nil
}

func (source *controlledTestConfig) setSequence(sequence string) bool {
	if source == nil || sequence != "1" && sequence != "2" && sequence != "3" {
		return false
	}
	source.mu.Lock()
	source.sequence = sequence
	source.mu.Unlock()
	return true
}

type restartableSource struct {
	address    string
	handler    http.Handler
	mu         sync.Mutex
	server     *http.Server
	listener   net.Listener
	generation int
}

func startRestartableSource(listener net.Listener, handler http.Handler) *restartableSource {
	if listener == nil || handler == nil {
		return nil
	}
	source := &restartableSource{address: listener.Addr().String(), handler: handler, listener: listener, generation: 1}
	source.server = platform.NewHTTPServer(source.address, http.HandlerFunc(source.serveHTTP))
	server := source.server
	go func() { _ = server.Serve(listener) }()
	return source
}

func (source *restartableSource) serveHTTP(writer http.ResponseWriter, request *http.Request) {
	if source == nil || source.handler == nil {
		writer.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	source.handler.ServeHTTP(writer, request)
}

func (source *restartableSource) stop(ctx context.Context) error {
	if source == nil || ctx == nil || ctx.Err() != nil {
		return errors.New("source stop rejected")
	}
	source.mu.Lock()
	server := source.server
	listener := source.listener
	source.server = nil
	source.listener = nil
	source.mu.Unlock()
	if server == nil || listener == nil {
		return nil
	}
	shutdownErr := server.Shutdown(ctx)
	closeErr := listener.Close()
	if shutdownErr != nil && !errors.Is(shutdownErr, http.ErrServerClosed) || closeErr != nil && !errors.Is(closeErr, net.ErrClosed) {
		return errors.New("source stop failed")
	}
	return nil
}

func (source *restartableSource) start(ctx context.Context) error {
	if source == nil || source.handler == nil || ctx == nil || ctx.Err() != nil || source.address == "" {
		return errors.New("source start rejected")
	}
	source.mu.Lock()
	defer source.mu.Unlock()
	if source.server != nil || source.listener != nil {
		return nil
	}
	listenConfig := net.ListenConfig{}
	listener, err := listenConfig.Listen(ctx, "tcp", source.address)
	if err != nil {
		return errors.New("source listener restart failed")
	}
	server := platform.NewHTTPServer(source.address, http.HandlerFunc(source.serveHTTP))
	source.listener = listener
	source.server = server
	source.generation++
	go func() { _ = server.Serve(listener) }()
	return nil
}

func (source *restartableSource) currentGeneration() int {
	if source == nil {
		return 0
	}
	source.mu.Lock()
	defer source.mu.Unlock()
	return source.generation
}

func runFixtureChild(input io.Reader, output io.Writer, capture *privacyCapture) int {
	ctx, cancel := context.WithCancel(context.Background())
	fixtureRuntime, err := openFixtureRuntime(ctx, cancel, capture)
	if err != nil {
		cancel()
		return 2
	}
	defer fixtureRuntime.close()
	encoder := json.NewEncoder(output)
	decoder := json.NewDecoder(io.LimitReader(input, 8<<20))
	if err := encoder.Encode(fixtureRuntime.ready); err != nil {
		return 2
	}
	for {
		var request fixtureRequest
		if err := decoder.Decode(&request); err != nil || request.ID == 0 {
			return 2
		}
		response := fixtureRuntime.handle(request)
		if err := encoder.Encode(response); err != nil {
			return 2
		}
		if request.Operation == "shutdown" {
			return 0
		}
	}
}

func (fixtureRuntime *fixtureRuntime) handle(request fixtureRequest) fixtureResponse {
	response := fixtureResponse{ID: request.ID, OK: true}
	switch request.Operation {
	case "shutdown":
	case "delivery":
		value, count, ok := fixtureRuntime.email.wait(request.Recipient, identity.TemplateID(request.Template), 10*time.Second)
		response.OK, response.Value, response.Count = ok, value, count
	case "email-mode":
		response.OK = fixtureRuntime.email.setMode(request.Value)
	case "email-attempts":
		response.Count = fixtureRuntime.email.attemptCount(request.Value)
	case "delivery-attempts":
		response.Count = fixtureRuntime.email.rawAttemptCount(request.Recipient, identity.TemplateID(request.Template))
	case "delivery-effects":
		response.Count = fixtureRuntime.email.effectCount(request.Recipient, identity.TemplateID(request.Template))
	case "broker-deliveries":
		response.Count = int(fixtureRuntime.deliveryFetches.Load())
	case "broker-ack-failures":
		response.Count = int(fixtureRuntime.deliveryAckFailures.Load())
	case "signer-mode":
		response.OK = fixtureRuntime.signer.setMode(request.Value)
	case "signer-attempts":
		response.Count = fixtureRuntime.signer.attemptCount(request.Value)
	case "reporter-mode":
		response.OK = fixtureRuntime.reports.setMode(request.Value)
	case "reporter-attempts":
		response.Count = fixtureRuntime.reports.attemptCount(request.Value)
	case "report":
		if !fixtureRuntime.reporter.TryReport(fixtureFiniteReport()) {
			fixtureRuntime.reportFull.Add(1)
		}
	case "report-full-count":
		response.Count = int(fixtureRuntime.reportFull.Load())
	case "report-count":
		response.Count = fixtureRuntime.reports.count()
	case "privacy-baseline":
		slog.Info(fixturePrivacyLogBaseline, "component", "e2e")
		response.OK = fixtureRuntime.reporter.TryReport(fixtureFiniteReport())
	case "privacy-canaries":
		response.OK = fixtureRuntime.reports.setPrivacyCanaries(request.Value, request.Template)
	case "privacy-scan":
		var canaries []string
		if err := json.Unmarshal([]byte(request.Value), &canaries); err != nil || len(canaries) == 0 || len(canaries) > 32 {
			response.OK = false
			break
		}
		clean, evidence := fixtureRuntime.privacy.scan(canaries)
		response.Value = strconv.FormatBool(clean)
		response.Count = evidence
	case "config-sequence":
		response.OK = fixtureRuntime.testConfig.setSequence(request.Value)
	case "clock-now":
		response.Value = fixtureRuntime.clock.Now().Format(time.RFC3339)
	case "clock-set":
		response.OK = fixtureRuntime.clock.set(request.Value)
		response.Value = fixtureRuntime.clock.Now().Format(time.RFC3339)
	case "nats-status":
		response.OK = fixtureRuntime.dependencies != nil && fixtureRuntime.dependencies.NATS != nil && fixtureRuntime.dependencies.NATS.IsConnected()
	case "source-mode":
		source, exists := fixtureRuntime.sources[request.Value]
		response.OK = exists && (request.Template == "up" || request.Template == "down")
		if response.OK {
			ctx, cancel := context.WithTimeout(context.Background(), fixtureRequestTimeout)
			if request.Template == "up" {
				response.OK = source.start(ctx) == nil
			} else {
				response.OK = source.stop(ctx) == nil
			}
			cancel()
			response.Count = source.currentGeneration()
		}
	default:
		response.OK = false
	}
	return response
}

func openFixtureRuntime(parent context.Context, cancel context.CancelFunc, capture *privacyCapture) (_ *fixtureRuntime, resultErr error) {
	dependencyConfiguration, enabled, err := loadExternalDependencyConfiguration(os.LookupEnv)
	if err != nil || !enabled {
		return nil, errors.New("fixture dependency contract failed")
	}
	listeners := make([]net.Listener, 0, 6)
	listenConfig := net.ListenConfig{}
	for range 6 {
		listener, err := listenConfig.Listen(parent, "tcp", "127.0.0.1:0")
		if err != nil {
			for _, opened := range listeners {
				_ = opened.Close()
			}
			return nil, errors.New("fixture listener failed")
		}
		listeners = append(listeners, listener)
	}
	urls := make([]string, len(listeners))
	for index, listener := range listeners {
		urls[index] = "http://" + listener.Addr().String()
	}
	primaryOrigin := configuredOriginForListener(urls[0])
	mirrorAOrigin := configuredOriginForListener(urls[3])
	mirrorBOrigin := configuredOriginForListener(urls[4])
	if primaryOrigin == "" || mirrorAOrigin == "" || mirrorBOrigin == "" {
		return nil, errors.New("fixture configured origin failed")
	}
	cfg, err := fixtureConfig(primaryOrigin, mirrorAOrigin, mirrorBOrigin, dependencyConfiguration)
	if err != nil {
		return nil, err
	}
	dependencies, err := platform.Open(parent, cfg)
	if err != nil {
		return nil, errors.New("fixture dependencies failed")
	}
	authorityClock := newFixtureClock()
	fixtureRuntime := &fixtureRuntime{
		cancel: cancel, listeners: listeners, dependencies: dependencies,
		testConfig: &controlledTestConfig{sequence: "1"}, sources: make(map[string]*restartableSource, 3),
		privacy: capture, clock: authorityClock,
	}
	defer func() { //nolint:contextcheck // Failure cleanup deliberately derives its own bounded shutdown context.
		if resultErr != nil {
			fixtureRuntime.close()
		}
	}()
	metrics := observability.NewRegistry()
	fixtureRuntime.protector, err = sensitive.NewLocal(cfg.Security.SensitiveLookupKey, cfg.Security.SensitiveEncryptionKey, 1)
	if err != nil {
		return nil, errors.New("fixture protector failed")
	}
	fixtureRuntime.rootSigner, err = trust.NewLocalRootSigner(cfg.Security.LocalRootSigningSeed)
	if err != nil {
		return nil, errors.New("fixture root signer failed")
	}
	fixtureRuntime.configSigner, err = trust.NewLocalConfigSigner(cfg.Security.LocalConfigSigningSeed)
	if err != nil {
		return nil, errors.New("fixture config signer failed")
	}
	fixtureRuntime.signer = newControlledConfigSigner(fixtureRuntime.configSigner)
	fixtureRuntime.timeoutSigner, err = trust.NewTimeoutConfigSigner(fixtureRuntime.signer, cfg.Security.SignerTimeout)
	if err != nil {
		return nil, errors.New("fixture timeout signer failed")
	}
	fixtureRuntime.email = newControlledEmailSender()
	fixtureRuntime.reports = newControlledReportProvider()
	fixtureRuntime.reports.capture = capture
	fixtureRuntime.reporter, err = errorreport.New(parent, fixtureRuntime.reports, cfg.ErrorReportQueue, cfg.ErrorReportBatch, cfg.Security.ErrorReportTimeout, metrics)
	if err != nil {
		return nil, errors.New("fixture reporter failed")
	}
	identityRepository, err := identity.NewPostgresRepository(dependencies.Postgres, fixtureRuntime.protector)
	if err != nil {
		return nil, errors.New("fixture identity repository failed")
	}
	deviceRepository, err := deviceauth.NewPostgresRepository(dependencies.Postgres, fixtureRuntime.protector)
	if err != nil {
		return nil, errors.New("fixture device repository failed")
	}
	trustRepository, err := trust.NewPostgresRepository(dependencies.Postgres, fixtureRuntime.protector)
	if err != nil {
		return nil, errors.New("fixture trust repository failed")
	}
	byteStore, err := trust.NewPostgresByteStore(dependencies.Postgres)
	if err != nil {
		return nil, errors.New("fixture byte store failed")
	}
	distribution, err := trust.NewDistribution(byteStore, fixtureRuntime.clock)
	if err != nil {
		return nil, errors.New("fixture distribution failed")
	}
	rootPublic := fixtureRuntime.rootSigner.PublicKey()
	metadataRepository, err := trust.NewPostgresMetadataRepository(dependencies.Postgres, map[string]ed25519.PublicKey{fixtureRuntime.rootSigner.KeyID(): rootPublic})
	if err != nil {
		clear(rootPublic)
		return nil, errors.New("fixture metadata repository failed")
	}
	metadata, signedMetadata, err := fixtureMetadata(parent, metadataRepository, fixtureRuntime.rootSigner, fixtureRuntime.configSigner, fixtureRuntime.clock.Now())
	if err != nil {
		clear(rootPublic)
		return nil, err
	}
	limiter, err := ratelimit.NewRedis(dependencies.Redis, cfg.Security.RedisTimeout)
	if err != nil {
		clear(rootPublic)
		return nil, errors.New("fixture limiter failed")
	}
	identityChallenges, err := identity.NewRedisChallengeStore(dependencies.Redis, cfg.Security.RedisTimeout)
	if err != nil {
		clear(rootPublic)
		return nil, errors.New("fixture identity challenge store failed")
	}
	deviceChallenges, err := deviceauth.NewRedisChallengeStore(dependencies.Redis, cfg.Security.RedisTimeout, fixtureRuntime.clock)
	if err != nil {
		clear(rootPublic)
		return nil, errors.New("fixture device challenge store failed")
	}
	outboxHealth, err := readiness.NewPostgresOutboxHealthProbe(dependencies.Postgres, fixtureRuntime.outboxClock())
	if err != nil {
		clear(rootPublic)
		return nil, errors.New("fixture outbox health failed")
	}
	checker, err := readiness.NewWithPolicy(cfg.DependencyTimeout, readiness.Policy{
		RedisDownAfterFailures: cfg.RedisDownAfterFailures, RedisRecoverAfterSuccesses: cfg.RedisRecoverAfterSuccesses,
		OutboxDegradedBacklog: cfg.OutboxDegradedBacklog, OutboxDownBacklog: cfg.OutboxDownBacklog,
		OutboxDegradedAge: cfg.OutboxDegradedAge, OutboxDownAge: cfg.OutboxDownAge,
	}, postgresprobe.Probe{Pool: dependencies.Postgres}, redisprobe.Probe{Client: dependencies.Redis}, outboxHealth)
	if err != nil {
		clear(rootPublic)
		return nil, errors.New("fixture readiness failed")
	}
	shared := fixtureApplicationDependencies{
		identityRepository: identityRepository, deviceRepository: deviceRepository, trustRepository: trustRepository,
		limiter: limiter, identityChallenges: identityChallenges, deviceChallenges: deviceChallenges,
		metadata: metadata, immutable: distribution,
	}
	profiles := []config.EmailVerificationMode{config.EmailRequired, config.EmailGrace, config.EmailDisabled}
	servers := make([]*http.Server, 0, 3)
	for index, mode := range profiles {
		profileConfig := cfg
		profileConfig.Security.EmailVerification = mode
		apps, composeErr := composeFixtureApplications(profileConfig, fixtureRuntime, shared)
		if composeErr != nil {
			clear(rootPublic)
			return nil, composeErr
		}
		api := controlapi.NewHandler(checker, apps, cfg.Security.RequestDeadline, rand.Reader)
		mux := http.NewServeMux()
		handler := controlapiv1.HandlerWithOptions(api, controlapiv1.StdHTTPServerOptions{BaseRouter: mux, ErrorHandlerFunc: api.GeneratedParameterError})
		observedHandler := metrics.Middleware("", handler)
		if index == 0 {
			fixtureRuntime.sources["primary"] = startRestartableSource(listeners[index], observedHandler)
			continue
		}
		server := platform.NewHTTPServer(listeners[index].Addr().String(), observedHandler)
		servers = append(servers, server)
		listener := listeners[index]
		go func() { _ = server.Serve(listener) }()
	}
	for index := 3; index <= 4; index++ {
		mirror, mirrorErr := trust.NewMirror(byteStore, fixtureRuntime.clock)
		if mirrorErr != nil {
			clear(rootPublic)
			return nil, errors.New("fixture mirror failed")
		}
		name := "mirror-a"
		if index == 4 {
			name = "mirror-b"
		}
		fixtureRuntime.sources[name] = startRestartableSource(listeners[index], mirror)
	}
	metricsServer := platform.NewHTTPServer(listeners[5].Addr().String(), metrics.Handler())
	servers = append(servers, metricsServer)
	go func() { _ = metricsServer.Serve(listeners[5]) }()
	fixtureRuntime.servers = servers
	if err := fixtureRuntime.startDelivery(parent, cfg); err != nil {
		clear(rootPublic)
		return nil, err
	}
	metadataBytes, err := json.Marshal(signedMetadata)
	if err != nil {
		clear(rootPublic)
		return nil, errors.New("fixture metadata encoding failed")
	}
	fixtureRuntime.ready = fixtureReady{
		Kind: "ready", RequiredURL: urls[0], GraceURL: urls[1], DisabledURL: urls[2], MirrorAURL: urls[3], MirrorBURL: urls[4], MetricsURL: urls[5],
		RootKeyID: fixtureRuntime.rootSigner.KeyID(), RootPublic: base64.RawURLEncoding.EncodeToString(rootPublic), Metadata: base64.RawURLEncoding.EncodeToString(metadataBytes),
	}
	clear(rootPublic)
	clear(metadataBytes)
	return fixtureRuntime, nil
}

type fixtureApplicationDependencies struct {
	identityRepository interface {
		identity.Repository
		identity.DeviceTransactionParticipant
	}
	deviceRepository interface {
		deviceauth.Repository
		identity.DeviceAuthorizationParticipant
	}
	trustRepository    trust.Repository
	limiter            ratelimit.Limiter
	identityChallenges identity.ChallengeStore
	deviceChallenges   deviceauth.ChallengeStore
	metadata           trust.RootMetadataV1
	immutable          http.Handler
}

func composeFixtureApplications(cfg config.Config, fixtureRuntime *fixtureRuntime, dependencies fixtureApplicationDependencies) (controlapi.Applications, error) {
	identityApplication, err := identity.NewApplication(identity.ApplicationDependencies{
		Repository: dependencies.identityRepository, Protector: fixtureRuntime.protector, Random: rand.Reader, Clock: fixtureRuntime.clock,
		Limiter: dependencies.limiter, ChallengeStore: dependencies.identityChallenges, RateLimitKey: cfg.Security.SensitiveLookupKey,
		Security: cfg.Security, DeviceAuthorizationParticipant: dependencies.deviceRepository,
	})
	if err != nil {
		return controlapi.Applications{}, errors.New("fixture identity application failed")
	}
	deviceApplication, err := deviceauth.NewApplication(deviceauth.ApplicationDependencies{
		Repository: dependencies.deviceRepository, IdentityParticipant: dependencies.identityRepository,
		Protector: fixtureRuntime.protector, Random: rand.Reader, Clock: fixtureRuntime.clock, Limiter: dependencies.limiter,
		ChallengeStore: dependencies.deviceChallenges, RateLimitKey: cfg.Security.SensitiveLookupKey, Security: cfg.Security,
	})
	if err != nil {
		return controlapi.Applications{}, errors.New("fixture device application failed")
	}
	trustApplication, err := trust.NewApplication(trust.ApplicationDependencies{
		Repository: dependencies.trustRepository, Device: deviceApplication, Signer: fixtureRuntime.timeoutSigner,
		Metadata: dependencies.metadata, Random: rand.Reader, Clock: fixtureRuntime.clock, TestConfig: fixtureRuntime.testConfig,
		BundleBaseURLs: cfg.Security.BundleBaseURLs, RequestDeadline: cfg.Security.RequestDeadline,
	})
	if err != nil {
		return controlapi.Applications{}, errors.New("fixture trust application failed")
	}
	return controlapi.Applications{
		Identity: identityApplication, StrongAuth: identityApplication, AccountAuth: identityApplication,
		Device: deviceApplication, DeviceAuth: deviceApplication, Trust: trustApplication, ImmutableBundle: dependencies.immutable,
	}, nil
}

func fixtureMetadata(ctx context.Context, repository trust.MetadataRepository, root *trust.LocalRootSigner, signer *trust.LocalConfigSigner, now time.Time) (trust.RootMetadataV1, trust.SignedRootMetadataV1, error) {
	validFrom := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	validUntil := time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)
	if err := trust.EnsureLocalMetadata(ctx, repository, root, signer, validFrom, validUntil); err != nil {
		return trust.RootMetadataV1{}, trust.SignedRootMetadataV1{}, errors.New("fixture metadata bootstrap failed")
	}
	records, err := repository.List(ctx)
	if err != nil || len(records) != 1 {
		return trust.RootMetadataV1{}, trust.SignedRootMetadataV1{}, errors.New("fixture metadata read failed")
	}
	rootPublic := root.PublicKey()
	defer clear(rootPublic)
	payload, err := trust.VerifyRootMetadataV1(records[0], map[string]ed25519.PublicKey{root.KeyID(): rootPublic}, 0, now)
	if err != nil {
		return trust.RootMetadataV1{}, trust.SignedRootMetadataV1{}, errors.New("fixture metadata verify failed")
	}
	return payload, records[0], nil
}

func fixtureConfig(primary, mirrorA, mirrorB string, dependencies externalDependencyConfiguration) (config.Config, error) {
	values := map[string]string{
		"TALENRO_HTTP_ADDRESS": "127.0.0.1:8080", "TALENRO_METRICS_ADDRESS": "127.0.0.1:9090",
		"TALENRO_ALLOW_PUBLIC_METRICS": "false", "TALENRO_DATABASE_URL": dependencies.DatabaseURL,
		"TALENRO_REDIS_ADDRESS": dependencies.RedisAddress, "TALENRO_NATS_URL": dependencies.NATSURL,
		"TALENRO_ALLOW_PUBLIC_HTTP": "false", "TALENRO_PROFILE": "test", "TALENRO_PUBLIC_BASE_URL": primary,
		"TALENRO_PRIMARY_BUNDLE_BASE_URL": primary, "TALENRO_MIRROR_A_BASE_URL": mirrorA, "TALENRO_MIRROR_B_BASE_URL": mirrorB,
		"TALENRO_WEBAUTHN_RP_ID": "localhost", "TALENRO_WEBAUTHN_ORIGINS": primary,
		"TALENRO_EMAIL_VERIFICATION_MODE": "required", "TALENRO_SIGNER_PROVIDER": "local",
		"TALENRO_FIELD_PROTECTOR_PROVIDER": "local", "TALENRO_EMAIL_PROVIDER": "local", "TALENRO_ERROR_REPORTER_PROVIDER": "local",
		"TALENRO_REQUEST_DEADLINE": fixtureEmailProviderTimeout.String(), "TALENRO_REDIS_TIMEOUT": "250ms", "TALENRO_SIGNER_TIMEOUT": "500ms", "TALENRO_ERROR_REPORT_TIMEOUT": "250ms",
		"TALENRO_REDIS_DOWN_AFTER_FAILURES": "3", "TALENRO_REDIS_RECOVER_AFTER_SUCCESSES": "2",
		"TALENRO_OUTBOX_DEGRADED_BACKLOG": "100", "TALENRO_OUTBOX_DOWN_BACKLOG": "101", "TALENRO_OUTBOX_DEGRADED_AGE": "10s", "TALENRO_OUTBOX_DOWN_AGE": "11s",
		"TALENRO_ERROR_REPORT_QUEUE": "10", "TALENRO_ERROR_REPORT_BATCH": "1", "TALENRO_CLOCK_SKEW": "120s",
		"TALENRO_LOGIN_RATE_LIMIT": "50", "TALENRO_LOGIN_RATE_WINDOW": "15m", "TALENRO_DELIVERY_RATE_LIMIT": "20", "TALENRO_DELIVERY_RATE_WINDOW": "1h",
		"TALENRO_CHALLENGE_RATE_LIMIT": "100", "TALENRO_CHALLENGE_RATE_WINDOW": "5m",
		"TALENRO_SENSITIVE_LOOKUP_KEY_B64":      "YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXowMTIzNDU",
		"TALENRO_SENSITIVE_ENCRYPTION_KEY_B64":  "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY",
		"TALENRO_LOCAL_ROOT_SIGNING_SEED_B64":   "cm9vdC1zaWduaW5nLXNlZWQtZm9yLWxvY2FsLXRlc3Q",
		"TALENRO_LOCAL_CONFIG_SIGNING_SEED_B64": "Y29uZmlnLXNpZ25pbmctc2VlZC1sb2NhbC10ZXN0LTE",
	}
	return config.Load(func(key string) (string, bool) { value, ok := values[key]; return value, ok })
}

func TestC11FixtureConfigUsesValidatedRateLimits(t *testing.T) {
	t.Parallel()

	cfg, err := fixtureConfig(
		"http://localhost:8080",
		"http://localhost:8081",
		"http://localhost:8082",
		externalDependencyConfiguration{
			DatabaseURL:  "postgres://talenro:talenro_dev@127.0.0.1:5432/talenro?sslmode=disable",
			RedisAddress: "127.0.0.1:6379",
			NATSURL:      "nats://127.0.0.1:4222",
		},
	)
	if err != nil {
		t.Fatal("C1.1 fixture rate-limit configuration rejected")
	}
	if cfg.Security.LoginRateLimit != (config.RateLimitPolicy{Limit: 50, Window: 15 * time.Minute}) {
		t.Fatal("C1.1 fixture login rate-limit policy changed")
	}
	if cfg.Security.DeliveryRateLimit != (config.RateLimitPolicy{Limit: 20, Window: time.Hour}) {
		t.Fatal("C1.1 fixture delivery rate-limit policy changed")
	}
	if cfg.Security.ChallengeRateLimit != (config.RateLimitPolicy{Limit: 100, Window: 5 * time.Minute}) {
		t.Fatal("C1.1 fixture challenge rate-limit policy changed")
	}
}

func TestC11FixtureIdempotencyKeysMeetProductionContract(t *testing.T) {
	t.Parallel()

	for _, prefix := range []string{"pg-ack", "grace-expired-rotation"} {
		if _, err := idempotency.KeyDigest(nextFixtureKey(prefix)); err != nil {
			t.Fatalf("C1.1 fixture idempotency key for prefix %q rejected by production contract", prefix)
		}
	}
}

func (fixtureRuntime *fixtureRuntime) startDelivery(ctx context.Context, cfg config.Config) error {
	jetStream, err := fixtureRuntime.dependencies.NATS.JetStream()
	if err != nil {
		return errors.New("fixture JetStream failed")
	}
	if _, err = jetStream.StreamInfo(outbox.EventStreamName); errors.Is(err, nats.ErrStreamNotFound) {
		_, err = jetStream.AddStream(&nats.StreamConfig{
			Name: outbox.EventStreamName, Subjects: []string{"talenro.>"}, Retention: nats.LimitsPolicy,
			Storage: nats.FileStorage, Discard: nats.DiscardOld, MaxMsgSize: 264 * 1024, Duplicates: 2 * time.Minute,
		})
	}
	if err != nil {
		return errors.New("fixture JetStream configuration failed")
	}
	subscription, err := jetStream.PullSubscribe(
		contractevents.EmailDeliveryRequestedType, fmt.Sprintf("c11-e2e-%d", os.Getpid()),
		nats.BindStream(outbox.EventStreamName), nats.ManualAck(), nats.AckExplicit(), nats.AckWait(fixtureNATSAckWait), nats.MaxAckPending(100),
	)
	if err != nil {
		return errors.New("fixture email subscription failed")
	}
	fixtureRuntime.subscription = subscription
	emailRepository, err := identity.NewPostgresEmailDeliveryRepository(fixtureRuntime.dependencies.Postgres)
	if err != nil {
		return errors.New("fixture email repository failed")
	}
	consumer, err := identity.NewEmailConsumer(emailRepository, fixtureRuntime.protector, fixtureRuntime.email, cfg.Security.RequestDeadline)
	if err != nil {
		return errors.New("fixture email consumer failed")
	}
	go fixtureEmailLoop(ctx, subscription, consumer, fixtureRuntime.clock, &fixtureRuntime.deliveryFetches, &fixtureRuntime.deliveryAckFailures)
	publisherStore, err := outbox.NewPostgresPublisherStore(fixtureRuntime.dependencies.Postgres)
	if err != nil {
		return errors.New("fixture publisher store failed")
	}
	broker, err := outbox.NewNATSBroker(jetStream, cfg.DependencyTimeout)
	if err != nil {
		return errors.New("fixture broker failed")
	}
	fixtureRuntime.publisher, err = outbox.NewPublisher(publisherStore, broker, fixtureRuntime.outboxClock(), rand.Reader, 100*time.Millisecond, fixtureRuntime.reporter)
	if err != nil || !fixtureRuntime.publisher.Start(ctx) {
		return errors.New("fixture publisher failed")
	}
	return nil
}

type fixtureDeliveryAcker interface {
	AckSync(...nats.AckOpt) error
}

type fixtureEmailSubscription interface {
	Fetch(int, ...nats.PullOpt) ([]*nats.Msg, error)
}

type fixtureEmailConsumer interface {
	Consume(context.Context, []byte, time.Time) error
}

func acknowledgeFixtureDelivery(ctx context.Context, message fixtureDeliveryAcker, failures *atomic.Int32) bool {
	if ctx == nil || ctx.Err() != nil || message == nil || failures == nil {
		return false
	}
	ackContext, cancel := context.WithTimeout(ctx, fixtureNATSAckTimeout)
	defer cancel()
	if err := message.AckSync(nats.Context(ackContext)); err != nil {
		failures.Add(1)
		return false
	}
	return true
}

func fixtureEmailLoop(
	ctx context.Context,
	subscription fixtureEmailSubscription,
	consumer fixtureEmailConsumer,
	clock securitykit.Clock,
	deliveries *atomic.Int32,
	ackFailures *atomic.Int32,
) {
	for ctx.Err() == nil {
		messages, err := subscription.Fetch(1, nats.MaxWait(250*time.Millisecond))
		if errors.Is(err, nats.ErrTimeout) {
			continue
		}
		if err != nil || len(messages) != 1 {
			continue
		}
		message := messages[0]
		deliveries.Add(1)
		if err := consumer.Consume(ctx, message.Data, clock.Now()); err != nil {
			if ctx.Err() != nil {
				return
			}
			nakContext, cancel := context.WithTimeout(ctx, fixtureNATSAckTimeout)
			if err := message.Nak(nats.Context(nakContext)); err != nil {
				ackFailures.Add(1)
			}
			cancel()
			continue
		}
		if ctx.Err() != nil {
			return
		}
		_ = acknowledgeFixtureDelivery(ctx, message, ackFailures)
	}
}

func (fixtureRuntime *fixtureRuntime) close() {
	if fixtureRuntime == nil {
		return
	}
	if fixtureRuntime.cancel != nil {
		fixtureRuntime.cancel()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for _, source := range fixtureRuntime.sources {
		_ = source.stop(ctx)
	}
	for _, server := range fixtureRuntime.servers {
		_ = server.Shutdown(ctx)
	}
	if fixtureRuntime.subscription != nil {
		_ = fixtureRuntime.subscription.Unsubscribe()
	}
	if fixtureRuntime.publisher != nil {
		_ = fixtureRuntime.publisher.Close(ctx)
	}
	if fixtureRuntime.reporter != nil {
		fixtureRuntime.reporter.Close(ctx)
	}
	if fixtureRuntime.timeoutSigner != nil {
		_ = fixtureRuntime.timeoutSigner.Close()
	}
	if fixtureRuntime.configSigner != nil {
		_ = fixtureRuntime.configSigner.Close()
	}
	if fixtureRuntime.rootSigner != nil {
		_ = fixtureRuntime.rootSigner.Close()
	}
	if fixtureRuntime.protector != nil {
		_ = fixtureRuntime.protector.Close()
	}
	if fixtureRuntime.dependencies != nil {
		fixtureRuntime.dependencies.Close()
	}
}

func fixtureFiniteReport() errorreport.Report {
	return errorreport.Report{
		Event: errorreport.EventRequestFailure, Category: errorreport.CategoryInternal, Component: errorreport.ComponentControlAPI,
		Outcome: errorreport.OutcomeFailure, Fingerprint: errorreport.FingerprintInternal, BuildVersion: "e2e", TraceID: "00000000000000000000000000000019",
	}
}

func (fixture *c11Fixture) runHappyPath(t *testing.T) {
	t.Helper()
	client := &c11HTTPClient{t: t, client: fixture.client}
	fixture.runAccountProfiles(t, client)
	account := fixture.runAccountSecurity(t, client)
	device := fixture.runDeviceAndTrust(t, client, account)
	fixture.verifyBundle(t, client, device)
	fixture.runEmailGraceLifecycle(t, client)
}

func (fixture *c11Fixture) runExternalHappyPath(t *testing.T) {
	t.Helper()
	if fixture == nil || !fixture.externalRuntime || fixture.child != nil || fixture.compose != nil {
		t.Fatal("c11 external happy path attempted to use a replacement runtime")
	}
	client := &c11HTTPClient{t: t, client: fixture.client}
	_, _ = client.get(fixture.ready.RequiredURL+"/livez", http.StatusOK)
	_, _ = client.get(fixture.ready.MirrorAURL+"/", http.StatusNotFound)
	_, _ = client.get(fixture.ready.MirrorBURL+"/", http.StatusNotFound)
	account := fixture.runExternalAccountSecurity(t, client)
	device := fixture.runDeviceAndTrust(t, client, account)
	fixture.verifyBundle(t, client, device)
}

func (fixture *c11Fixture) runExternalAccountSecurity(t *testing.T, client *c11HTTPClient) c11Account {
	t.Helper()
	suffix := fixtureKeySequence.Add(1)
	email := fmt.Sprintf("external-account-%d@example.test", suffix)
	genericProbeEmail, neverRegisteredEmail := externalEnumerationProbeEmails(suffix)
	password := "Task19-External-Password!61" //nolint:gosec // Synthetic E2E account credential.
	registered, _ := client.post(fixture.ready.RequiredURL, "/v1/accounts", "", nextFixtureKey("external-register"), map[string]any{
		"email": email, "password": password, "locale": "en",
	}, http.StatusAccepted)
	duplicate, _ := client.post(fixture.ready.RequiredURL, "/v1/accounts", "", nextFixtureKey("external-duplicate"), map[string]any{
		"email": email, "password": password, "locale": "en",
	}, http.StatusAccepted)
	unknown, _ := client.post(fixture.ready.RequiredURL, "/v1/accounts", "", nextFixtureKey("external-generic"), map[string]any{
		"email": genericProbeEmail, "password": password, "locale": "en",
	}, http.StatusAccepted)
	if !bytes.Equal(registered, duplicate) || !bytes.Equal(registered, unknown) {
		t.Fatal("c11 external registration disclosed account existence")
	}
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal("c11 external account signing key generation failed")
	}
	t.Cleanup(func() { clear(privateKey) })
	loginRequest := func(candidateEmail, candidatePassword, key string, expected int) []byte {
		body, _ := client.post(fixture.ready.RequiredURL, "/v1/account-sessions", "", key, map[string]any{
			"method": "password", "email": candidateEmail, "password": candidatePassword,
			"client_signing_public_key": base64.RawURLEncoding.EncodeToString(publicKey),
		}, expected)
		return body
	}
	wrong := loginRequest(email, "Task19-Wrong-Password!62", nextFixtureKey("external-wrong"), http.StatusUnauthorized)
	missing := loginRequest(neverRegisteredEmail, password, nextFixtureKey("external-missing"), http.StatusUnauthorized)
	if err := compareExternalAuthenticationErrors(wrong, missing); err != nil {
		t.Fatal("c11 external login disclosed account existence")
	}
	loginBody := loginRequest(email, password, nextFixtureKey("external-login"), http.StatusOK)
	account := decodeAccountTokens(t, fixture.ready.RequiredURL, email, password, publicKey, privateKey, loginBody)
	changedPassword := "Task19-External-Changed!63" //nolint:gosec // Synthetic E2E account credential.
	changedBody, _ := client.post(fixture.ready.RequiredURL, "/v1/password-changes", account.accessToken, nextFixtureKey("external-change"), map[string]any{
		"current_password": password, "new_password": changedPassword,
		"reauthentication":          map[string]any{"method": "password", "password": password},
		"client_signing_public_key": base64.RawURLEncoding.EncodeToString(publicKey),
	}, http.StatusOK)
	account = decodeAccountTokens(t, fixture.ready.RequiredURL, email, changedPassword, publicKey, privateKey, changedBody)
	fixture.exerciseStrongAuthentication(t, client, &account)
	fixture.exerciseSessionRevocation(t, client, &account)
	return account
}

func externalEnumerationProbeEmails(suffix uint64) (genericProbe, neverRegistered string) {
	return fmt.Sprintf("external-generic-%d@example.test", suffix),
		fmt.Sprintf("external-never-registered-%d@example.test", suffix)
}

func compareExternalAuthenticationErrors(leftBody, rightBody []byte) error {
	left, err := decodeExternalAuthenticationError(leftBody)
	if err != nil {
		return errors.New("external public error invalid")
	}
	right, err := decodeExternalAuthenticationError(rightBody)
	if err != nil {
		return errors.New("external public error invalid")
	}
	if left.Code != right.Code || left.Action != right.Action || left.RetryAfterMs != nil || right.RetryAfterMs != nil {
		return errors.New("external public error mismatch")
	}
	return nil
}

func decodeExternalAuthenticationError(body []byte) (controlapiv1.PublicError, error) {
	var result controlapiv1.PublicError
	if err := strictjson.Decode(bytes.NewReader(body), 4<<10, &result); err != nil ||
		result.Code != controlapiv1.PublicErrorCodeAuthenticationFailed || result.Action != controlapiv1.Reauthenticate ||
		result.RetryAfterMs != nil || !validExternalTraceID(result.TraceId) {
		return controlapiv1.PublicError{}, errors.New("external authentication error invalid")
	}
	return result, nil
}

func validExternalTraceID(value string) bool {
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	defer clear(decoded)
	return err == nil && len(decoded) == 16 && base64.RawURLEncoding.EncodeToString(decoded) == value
}

func (fixture *c11Fixture) deviceProofAudienceFor(networkEndpoint string) (string, error) {
	if fixture == nil || fixture.deviceProofAudience == "" {
		return "", errors.New("device proof audience unavailable")
	}
	if _, err := parseOwnedLoopbackURL(networkEndpoint, "http"); err != nil {
		return "", errors.New("device network endpoint invalid")
	}
	if networkEndpoint != fixture.ready.RequiredURL && networkEndpoint != fixture.ready.GraceURL && networkEndpoint != fixture.ready.DisabledURL {
		return "", errors.New("device network endpoint ownership invalid")
	}
	required, err := parseOwnedLoopbackURL(fixture.ready.RequiredURL, "http")
	if err != nil {
		return "", errors.New("device proof audience unavailable")
	}
	if _, err = parseConfiguredBundleOrigin(fixture.deviceProofAudience, required.Port()); err != nil {
		return "", errors.New("device proof audience invalid")
	}
	return fixture.deviceProofAudience, nil
}

type c11HTTPClient struct {
	t       *testing.T
	client  *http.Client
	context context.Context
}

type clearingRequestBody struct {
	mu     sync.Mutex
	reader *bytes.Reader
	value  []byte
	closed bool
}

func newClearingRequestBody(value []byte) *clearingRequestBody {
	return &clearingRequestBody{reader: bytes.NewReader(value), value: value}
}

func (body *clearingRequestBody) Read(destination []byte) (int, error) {
	if body == nil {
		return 0, io.ErrClosedPipe
	}
	body.mu.Lock()
	defer body.mu.Unlock()
	if body.closed {
		return 0, io.ErrClosedPipe
	}
	return body.reader.Read(destination)
}

func (body *clearingRequestBody) Close() error {
	if body == nil {
		return nil
	}
	body.mu.Lock()
	defer body.mu.Unlock()
	if body.closed {
		return nil
	}
	body.closed = true
	clear(body.value)
	body.value = nil
	body.reader.Reset(nil)
	return nil
}

func (*clearingRequestBody) String() string   { return "[REDACTED]" }
func (*clearingRequestBody) GoString() string { return "[REDACTED]" }

func newC11POSTRequest(ctx context.Context, rawURL string, encoded []byte) (*http.Request, error) {
	body := newClearingRequestBody(encoded)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, rawURL, body)
	if err != nil {
		_ = body.Close()
		return nil, err
	}
	request.ContentLength = int64(len(encoded))
	return request, nil
}

func (client *c11HTTPClient) post(baseURL, path, bearer, idempotency string, body any, expected ...int) ([]byte, int) {
	client.t.Helper()
	encoded, err := json.Marshal(body)
	if err != nil {
		client.t.Fatal("c11 HTTP fixture encoding failed")
	}
	requestContext, cancel := context.WithTimeout(client.parentContext(), fixtureRequestTimeout)
	defer cancel()
	request, err := newC11POSTRequest(requestContext, baseURL+path, encoded)
	if err != nil {
		client.t.Fatal("c11 HTTP request construction failed")
	}
	request.Header.Set("Content-Type", "application/json")
	if bearer != "" {
		request.Header.Set("Authorization", "Bearer "+bearer)
	}
	if idempotency != "" {
		request.Header.Set("Idempotency-Key", idempotency)
	}
	response, err := client.client.Do(request)
	if err != nil {
		client.t.Fatalf("c11 HTTP stage %s failed", path)
	}
	defer func() { _ = response.Body.Close() }()
	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		client.t.Fatalf("c11 HTTP stage %s response failed", path)
	}
	for _, status := range expected {
		if response.StatusCode == status {
			return responseBody, response.StatusCode
		}
	}
	clear(responseBody)
	client.t.Fatalf("c11 HTTP stage %s returned status %d", path, response.StatusCode)
	return nil, response.StatusCode
}

func (client *c11HTTPClient) get(url string, expected ...int) ([]byte, int) {
	client.t.Helper()
	requestContext, cancel := context.WithTimeout(client.parentContext(), fixtureRequestTimeout)
	defer cancel()
	request, err := http.NewRequestWithContext(requestContext, http.MethodGet, url, nil)
	if err != nil {
		client.t.Fatal("c11 HTTP GET construction failed")
	}
	response, err := client.client.Do(request)
	if err != nil {
		client.t.Fatal("c11 HTTP GET failed")
	}
	defer func() { _ = response.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(response.Body, 2<<20))
	if err != nil {
		client.t.Fatal("c11 HTTP GET response failed")
	}
	for _, status := range expected {
		if response.StatusCode == status {
			return body, response.StatusCode
		}
	}
	clear(body)
	client.t.Fatalf("c11 HTTP GET returned status %d", response.StatusCode)
	return nil, response.StatusCode
}

func (client *c11HTTPClient) parentContext() context.Context {
	if client != nil && client.context != nil {
		return client.context
	}
	if client != nil && client.t != nil {
		return client.t.Context()
	}
	return context.Background()
}

func nextFixtureKey(prefix string) string {
	padding := 20 - len([]rune(prefix))
	if padding < 0 {
		padding = 0
	}
	return fmt.Sprintf("%s%s%012d", prefix, strings.Repeat("-", padding), fixtureKeySequence.Add(1))
}

func decodeObject(t *testing.T, body []byte) map[string]any {
	t.Helper()
	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatal("c11 HTTP response schema failed")
	}
	return result
}

func stringField(object map[string]any, name string) string {
	value, _ := object[name].(string)
	return value
}

type c11Account struct {
	baseURL       string
	email         string
	password      string
	accessToken   string
	refreshToken  string
	principalID   uuid.UUID
	sessionID     uuid.UUID
	signingPublic ed25519.PublicKey
	signingKey    ed25519.PrivateKey
}

type c11Device struct {
	baseURL         string
	accessToken     string
	refreshToken    string
	deviceID        uuid.UUID
	authorizationID uuid.UUID
	familyID        uuid.UUID
	signingPublic   ed25519.PublicKey
	signingKey      ed25519.PrivateKey
	hpkePrivate     *ecdh.PrivateKey
	resolution      map[string]any
	expiresAt       string
}

func (fixture *c11Fixture) runAccountProfiles(t *testing.T, client *c11HTTPClient) {
	t.Helper()
	password := "Task19-Required-Password!47" //nolint:gosec // Synthetic E2E account credential.
	required := "required-c11@example.test"
	first, _ := client.post(fixture.ready.RequiredURL, "/v1/accounts", "", nextFixtureKey("register-required"), map[string]any{"email": required, "password": password, "locale": "en"}, http.StatusAccepted)
	duplicate, _ := client.post(fixture.ready.RequiredURL, "/v1/accounts", "", nextFixtureKey("register-duplicate"), map[string]any{"email": required, "password": password, "locale": "en"}, http.StatusAccepted)
	unknown, _ := client.post(fixture.ready.RequiredURL, "/v1/accounts", "", nextFixtureKey("register-generic"), map[string]any{"email": "generic-c11@example.test", "password": password, "locale": "en"}, http.StatusAccepted)
	if !bytes.Equal(first, duplicate) || !bytes.Equal(first, unknown) {
		t.Fatal("c11 registration disclosed account existence")
	}
	public, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal("c11 account signing key generation failed")
	}
	requiredDenied, _ := client.post(fixture.ready.RequiredURL, "/v1/account-sessions", "", nextFixtureKey("required-preverify"), map[string]any{
		"method": "password", "email": required, "password": password, "client_signing_public_key": base64.RawURLEncoding.EncodeToString(public),
	}, http.StatusUnauthorized)
	wrongDenied, _ := client.post(fixture.ready.RequiredURL, "/v1/account-sessions", "", nextFixtureKey("required-wrong"), map[string]any{
		"method": "password", "email": required, "password": "Task19-Wrong-Password!82", "client_signing_public_key": base64.RawURLEncoding.EncodeToString(public),
	}, http.StatusUnauthorized)
	unknownDenied, _ := client.post(fixture.ready.RequiredURL, "/v1/account-sessions", "", nextFixtureKey("required-unknown"), map[string]any{
		"method": "password", "email": "required-missing-c11@example.test", "password": password, "client_signing_public_key": base64.RawURLEncoding.EncodeToString(public),
	}, http.StatusUnauthorized)
	if compareExternalAuthenticationErrors(requiredDenied, wrongDenied) != nil || compareExternalAuthenticationErrors(requiredDenied, unknownDenied) != nil {
		t.Fatal("c11 required pre-verification denial disclosed account state")
	}
	fixture.verifyEmail(t, client, fixture.ready.RequiredURL, required)
	_, _ = client.post(fixture.ready.RequiredURL, "/v1/account-sessions", "", nextFixtureKey("required-verified"), map[string]any{
		"method": "password", "email": required, "password": password, "client_signing_public_key": base64.RawURLEncoding.EncodeToString(public),
	}, http.StatusOK)
	for _, profile := range []struct{ url, email string }{
		{fixture.ready.GraceURL, "grace-c11@example.test"},
		{fixture.ready.DisabledURL, "disabled-c11@example.test"},
	} {
		_, _ = client.post(profile.url, "/v1/accounts", "", nextFixtureKey("register-profile"), map[string]any{"email": profile.email, "password": password, "locale": "en"}, http.StatusAccepted)
		_, _ = client.post(profile.url, "/v1/account-sessions", "", nextFixtureKey("login-profile"), map[string]any{
			"method": "password", "email": profile.email, "password": password, "client_signing_public_key": base64.RawURLEncoding.EncodeToString(public),
		}, http.StatusOK)
	}
}

func (fixture *c11Fixture) runEmailGraceLifecycle(t *testing.T, client *c11HTTPClient) {
	t.Helper()
	clockResponse, err := fixture.child.call("clock-now", "", "", "")
	start, parseErr := time.Parse(time.RFC3339, clockResponse.Value)
	if err != nil || parseErr != nil || start.Location() != time.UTC {
		t.Fatal("c11 grace clock unavailable")
	}
	deadline := start.Add(24 * time.Hour)
	password := "Task19-Grace-Password!81" //nolint:gosec // Synthetic E2E account credential.
	email := fmt.Sprintf("grace-lifecycle-%d@example.test", fixtureKeySequence.Add(1))
	_, _ = client.post(fixture.ready.GraceURL, "/v1/accounts", "", nextFixtureKey("grace-register"), map[string]any{
		"email": email, "password": password, "locale": "en",
	}, http.StatusAccepted)
	first := fixture.createGraceSession(t, client, email, password, start.Add(10*time.Minute))
	firstGrant, firstGrantExpiry := fixture.createGraceGrant(t, client, first)
	if firstGrantExpiry != start.Add(10*time.Minute).Format(time.RFC3339) {
		t.Fatal("c11 first grace grant did not use the literal ten-minute expiry")
	}

	next := start.Add(time.Second)
	if response, setErr := fixture.child.call("clock-set", next.Format(time.RFC3339), "", ""); setErr != nil || response.Value != next.Format(time.RFC3339) {
		t.Fatal("c11 grace clock advance failed")
	}
	second := fixture.createGraceSession(t, client, email, password, next.Add(10*time.Minute))
	secondGrant, secondGrantExpiry := fixture.createGraceGrant(t, client, second)
	if first.sessionID == second.sessionID || firstGrant == secondGrant || secondGrantExpiry != next.Add(10*time.Minute).Format(time.RFC3339) {
		t.Fatal("c11 independent grace sessions or grants were not distinct and fixed")
	}
	device := fixture.registerDevice(t, client, fixture.ready.GraceURL, firstGrant)
	if device.expiresAt != next.Add(10*time.Minute).Format(time.RFC3339) {
		t.Fatal("c11 provisional registration did not use the literal ten-minute access expiry")
	}
	fixture.requireGraceGrantDenied(t, client, secondGrant)
	wantTrialPolicy := `{"expires_at":"` + deadline.Format(time.RFC3339) + `","max_devices":"1","mode":"trial_restricted"}`
	if got := fixture.resolveBundlePolicy(t, client, device, next); got != wantTrialPolicy {
		t.Fatalf("c11 provisional policy mismatch: got %s", got)
	}
	device = fixture.rotateDeviceOnce(t, client, device)
	if device.expiresAt != next.Add(10*time.Minute).Format(time.RFC3339) {
		t.Fatal("c11 provisional rotation did not use the literal ten-minute access expiry")
	}

	probeEmail := fmt.Sprintf("grace-expiry-%d@example.test", fixtureKeySequence.Add(1))
	_, _ = client.post(fixture.ready.GraceURL, "/v1/accounts", "", nextFixtureKey("grace-expiry-register"), map[string]any{
		"email": probeEmail, "password": password, "locale": "en",
	}, http.StatusAccepted)
	probeDeadline := next.Add(24 * time.Hour)
	probeAccount := fixture.createGraceSession(t, client, probeEmail, password, next.Add(10*time.Minute))
	probeGrant, _ := fixture.createGraceGrant(t, client, probeAccount)
	probeDevice := fixture.registerDevice(t, client, fixture.ready.GraceURL, probeGrant)

	fixture.verifyEmail(t, client, fixture.ready.GraceURL, email)
	if _, sequenceErr := fixture.child.call("config-sequence", "2", "", ""); sequenceErr != nil {
		t.Fatal("c11 verified policy sequence advance failed")
	}
	if got := fixture.resolveBundlePolicy(t, client, device, next); got != `{"mode":"standard"}` {
		t.Fatalf("c11 verified effective policy mismatch: got %s", got)
	}
	if response, setErr := fixture.child.call("clock-set", probeDeadline.Format(time.RFC3339), "", ""); setErr != nil || response.Value != probeDeadline.Format(time.RFC3339) {
		t.Fatal("c11 grace equality clock advance failed")
	}
	_ = fixture.rotateDeviceOnce(t, client, device)
	_, _ = client.post(probeDevice.baseURL, "/v1/config-bundle-resolutions", probeDevice.accessToken, nextFixtureKey("grace-expired-bundle"), map[string]any{}, http.StatusUnauthorized)
	probeNonce := random32(t)
	_, _ = client.post(probeDevice.baseURL, "/v1/device-auth-challenges", "", nextFixtureKey("grace-expired-rotation"), map[string]any{
		"refresh_token": probeDevice.refreshToken, "request_nonce": encode32(probeNonce),
	}, http.StatusUnauthorized)
	_, _ = client.post(probeAccount.baseURL, "/v1/device-enrollment-grants", probeAccount.accessToken, nextFixtureKey("grace-expired-account"), map[string]any{
		"reauthentication": map[string]any{"method": "password", "password": password},
	}, http.StatusUnauthorized)
	_, _ = client.post(probeAccount.baseURL, "/v1/account-auth-challenges", "", nextFixtureKey("grace-expired-refresh"), map[string]any{
		"refresh_token": probeAccount.refreshToken, "request_nonce": encode32(random32(t)),
	}, http.StatusUnauthorized)
}

func (fixture *c11Fixture) createGraceSession(
	t *testing.T,
	client *c11HTTPClient,
	email, password string,
	expectedAccessExpiry time.Time,
) c11Account {
	t.Helper()
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal("c11 grace account key generation failed")
	}
	t.Cleanup(func() { clear(privateKey) })
	body, _ := client.post(fixture.ready.GraceURL, "/v1/account-sessions", "", nextFixtureKey("grace-login"), map[string]any{
		"method": "password", "email": email, "password": password,
		"client_signing_public_key": base64.RawURLEncoding.EncodeToString(publicKey),
	}, http.StatusOK)
	object := decodeObject(t, body)
	if stringField(object, "expires_at") != expectedAccessExpiry.Format(time.RFC3339) {
		t.Fatal("c11 grace session did not use the expected access expiry")
	}
	return decodeAccountTokens(t, fixture.ready.GraceURL, email, password, publicKey, privateKey, body)
}

func (fixture *c11Fixture) createGraceGrant(t *testing.T, client *c11HTTPClient, account c11Account) (string, string) {
	t.Helper()
	body, _ := client.post(account.baseURL, "/v1/device-enrollment-grants", account.accessToken, nextFixtureKey("grace-grant"), map[string]any{
		"reauthentication": map[string]any{"method": "password", "password": account.password},
	}, http.StatusCreated)
	object := decodeObject(t, body)
	grant := stringField(object, "enrollment_grant")
	if grant == "" {
		t.Fatal("c11 grace enrollment grant missing")
	}
	return grant, stringField(object, "expires_at")
}

func (fixture *c11Fixture) requireGraceGrantDenied(t *testing.T, client *c11HTTPClient, grant string) {
	t.Helper()
	signingPublic, signingPrivate, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal("c11 denied grace device key generation failed")
	}
	defer clear(signingPrivate)
	hpkePrivate, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal("c11 denied grace HPKE key generation failed")
	}
	_, _ = client.post(fixture.ready.GraceURL, "/v1/device-auth-challenges", "", nextFixtureKey("grace-second-denied"), map[string]any{
		"enrollment_grant": grant, "request_nonce": encode32(random32(t)),
		"signing_public_key": base64.RawURLEncoding.EncodeToString(signingPublic),
		"hpke_public_key":    base64.RawURLEncoding.EncodeToString(hpkePrivate.PublicKey().Bytes()),
	}, http.StatusUnauthorized)
}

func (fixture *c11Fixture) verifyEmail(t *testing.T, client *c11HTTPClient, baseURL, email string) {
	t.Helper()
	response, err := fixture.child.call("delivery", "", email, string(identity.VerifyEmailTemplate))
	if err != nil || response.Count != 1 {
		t.Fatal("c11 verification delivery unavailable")
	}
	_, _ = client.post(baseURL, "/v1/email-verifications", "", nextFixtureKey("verify-email"), map[string]any{"token": response.Value}, http.StatusNoContent)
}

func (fixture *c11Fixture) runAccountSecurity(t *testing.T, client *c11HTTPClient) c11Account {
	t.Helper()
	email := fmt.Sprintf("account-%d@example.test", time.Now().UnixNano())
	password := "Task19-Account-Password!48" //nolint:gosec // Synthetic E2E account credential.
	_, _ = client.post(fixture.ready.RequiredURL, "/v1/accounts", "", nextFixtureKey("register-account"), map[string]any{"email": email, "password": password, "locale": "en"}, http.StatusAccepted)
	fixture.verifyEmail(t, client, fixture.ready.RequiredURL, email)
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal("c11 account signing key generation failed")
	}
	loginBody, _ := client.post(fixture.ready.RequiredURL, "/v1/account-sessions", "", nextFixtureKey("password-login"), map[string]any{
		"method": "password", "email": email, "password": password, "client_signing_public_key": base64.RawURLEncoding.EncodeToString(publicKey),
	}, http.StatusOK)
	account := decodeAccountTokens(t, fixture.ready.RequiredURL, email, password, publicKey, privateKey, loginBody)
	preResetSession := account
	acceptedKnown, _ := client.post(fixture.ready.RequiredURL, "/v1/password-reset-deliveries", "", nextFixtureKey("reset-known"), map[string]any{"email": email, "locale": "en"}, http.StatusAccepted)
	acceptedUnknown, _ := client.post(fixture.ready.RequiredURL, "/v1/password-reset-deliveries", "", nextFixtureKey("reset-unknown"), map[string]any{"email": "missing-c11@example.test", "locale": "en"}, http.StatusAccepted)
	if !bytes.Equal(acceptedKnown, acceptedUnknown) {
		t.Fatal("c11 reset delivery disclosed account existence")
	}
	delivery, err := fixture.child.call("delivery", "", email, string(identity.ResetPasswordTemplate))
	if err != nil || delivery.Count != 1 {
		t.Fatal("c11 reset delivery unavailable")
	}
	resetPassword := "Task19-Reset-Password!49" //nolint:gosec // Synthetic E2E account credential.
	resetBody, _ := client.post(fixture.ready.RequiredURL, "/v1/password-resets", "", nextFixtureKey("reset-password"), map[string]any{
		"email": email, "token": delivery.Value, "new_password": resetPassword, "client_signing_public_key": base64.RawURLEncoding.EncodeToString(publicKey),
	}, http.StatusOK)
	resetAccount := decodeAccountTokens(t, fixture.ready.RequiredURL, email, resetPassword, publicKey, privateKey, resetBody)
	_, _ = client.post(fixture.ready.RequiredURL, "/v1/device-enrollment-grants", preResetSession.accessToken, nextFixtureKey("pre-reset-review"), map[string]any{
		"reauthentication": map[string]any{"method": "password", "password": resetPassword},
	}, http.StatusUnauthorized)
	account = resetAccount
	changedPassword := "Task19-Changed-Password!50" //nolint:gosec // Synthetic E2E account credential.
	changedBody, _ := client.post(fixture.ready.RequiredURL, "/v1/password-changes", account.accessToken, nextFixtureKey("change-password"), map[string]any{
		"current_password": resetPassword, "new_password": changedPassword,
		"reauthentication":          map[string]any{"method": "password", "password": resetPassword},
		"client_signing_public_key": base64.RawURLEncoding.EncodeToString(publicKey),
	}, http.StatusOK)
	account = decodeAccountTokens(t, fixture.ready.RequiredURL, email, changedPassword, publicKey, privateKey, changedBody)
	fixture.exerciseStrongAuthentication(t, client, &account)
	fixture.exerciseSessionRevocation(t, client, &account)
	return account
}

func decodeAccountTokens(t *testing.T, baseURL, email, password string, public ed25519.PublicKey, private ed25519.PrivateKey, body []byte) c11Account {
	t.Helper()
	object := decodeObject(t, body)
	principal, principalErr := uuid.Parse(stringField(object, "principal_id"))
	session, sessionErr := uuid.Parse(stringField(object, "session_id"))
	if principalErr != nil || sessionErr != nil || stringField(object, "access_token") == "" || stringField(object, "refresh_token") == "" {
		t.Fatal("c11 account token response invalid")
	}
	return c11Account{
		baseURL: baseURL, email: email, password: password, accessToken: stringField(object, "access_token"), refreshToken: stringField(object, "refresh_token"),
		principalID: principal, sessionID: session, signingPublic: public, signingKey: private,
	}
}

func (fixture *c11Fixture) exerciseStrongAuthentication(t *testing.T, client *c11HTTPClient, account *c11Account) {
	t.Helper()
	fixture.exercisePasskey(t, client, account)
	fixture.exerciseTOTPAndRecovery(t, client, account)
}

func (fixture *c11Fixture) exercisePasskey(t *testing.T, client *c11HTTPClient, account *c11Account) {
	t.Helper()
	optionsBody, _ := client.post(account.baseURL, "/v1/passkey-registration-options", account.accessToken, nextFixtureKey("passkey-options"), map[string]any{
		"reauthentication": map[string]any{"method": "password", "password": account.password},
	}, http.StatusOK)
	options := decodeObject(t, optionsBody)
	publicKey, _ := options["publicKey"].(map[string]any)
	credentialID := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, credentialID); err != nil {
		t.Fatal("c11 passkey credential generation failed")
	}
	origin := fixture.webAuthnOrigin
	rpID := fixture.webAuthnRPID
	if origin == "" {
		origin = account.baseURL
	}
	if rpID == "" {
		rpID = "127.0.0.1"
	}
	response := c11PasskeyRegistrationResponse(t, stringField(publicKey, "challenge"), origin, rpID, credentialID)
	_, _ = client.post(account.baseURL, "/v1/passkey-credentials", account.accessToken, nextFixtureKey("passkey-create"), map[string]any{
		"ceremony_id": stringField(options, "ceremony_id"), "response": json.RawMessage(response),
		"reauthentication": map[string]any{"method": "password", "password": account.password},
	}, http.StatusNoContent)
	_, _ = client.post(account.baseURL, "/v1/passkey-revocations", account.accessToken, nextFixtureKey("passkey-revoke"), map[string]any{
		"credential_id": base64.RawURLEncoding.EncodeToString(credentialID), "reauthentication": map[string]any{"method": "password", "password": account.password},
	}, http.StatusNoContent)
}

func (fixture *c11Fixture) exerciseTOTPAndRecovery(t *testing.T, client *c11HTTPClient, account *c11Account) {
	t.Helper()
	enrollmentBody, _ := client.post(account.baseURL, "/v1/totp-enrollments", account.accessToken, nextFixtureKey("totp-enroll"), map[string]any{
		"reauthentication": map[string]any{"method": "password", "password": account.password},
	}, http.StatusCreated)
	enrollment := decodeObject(t, enrollmentBody)
	code := c11TOTPCode(t, stringField(enrollment, "secret"), fixture.authorityTime(t))
	_, _ = client.post(account.baseURL, "/v1/totp-verifications", account.accessToken, nextFixtureKey("totp-verify"), map[string]any{
		"code": code, "reauthentication": map[string]any{"method": "password", "password": account.password},
	}, http.StatusNoContent)
	recoveryBody, _ := client.post(account.baseURL, "/v1/recovery-code-rotations", account.accessToken, nextFixtureKey("recovery-rotate"), map[string]any{
		"reauthentication": map[string]any{"method": "password", "password": account.password},
	}, http.StatusCreated)
	recovery := decodeObject(t, recoveryBody)
	codes, _ := recovery["codes"].([]any)
	if len(codes) != 10 {
		t.Fatal("c11 recovery code response invalid")
	}
	_, _ = client.post(account.baseURL, "/v1/totp-revocations", account.accessToken, nextFixtureKey("totp-revoke"), map[string]any{
		"reauthentication": map[string]any{"method": "password", "password": account.password},
	}, http.StatusNoContent)
	recoverySession, _ := client.post(account.baseURL, "/v1/recovery-code-consumptions", "", nextFixtureKey("recovery-consume"), map[string]any{
		"email": account.email, "code": fmt.Sprint(codes[0]), "client_signing_public_key": base64.RawURLEncoding.EncodeToString(account.signingPublic),
	}, http.StatusOK)
	updated := decodeAccountTokens(t, account.baseURL, account.email, account.password, account.signingPublic, account.signingKey, recoverySession)
	account.accessToken, account.refreshToken, account.sessionID = updated.accessToken, updated.refreshToken, updated.sessionID
}

func (fixture *c11Fixture) authorityTime(t *testing.T) time.Time {
	t.Helper()
	if fixture == nil || fixture.authorityNow == nil {
		t.Fatal("c11 authority clock unavailable")
	}
	now, err := fixture.authorityNow()
	if err != nil || now.IsZero() || now.Location() != time.UTC {
		t.Fatal("c11 authority clock invalid")
	}
	return now
}

func (fixture *c11Fixture) exerciseSessionRevocation(t *testing.T, client *c11HTTPClient, account *c11Account) {
	t.Helper()
	secondBody, _ := client.post(account.baseURL, "/v1/account-sessions", "", nextFixtureKey("session-second"), map[string]any{
		"method": "password", "email": account.email, "password": account.password, "client_signing_public_key": base64.RawURLEncoding.EncodeToString(account.signingPublic),
	}, http.StatusOK)
	second := decodeAccountTokens(t, account.baseURL, account.email, account.password, account.signingPublic, account.signingKey, secondBody)
	_, _ = client.post(account.baseURL, "/v1/account-session-revocations", second.accessToken, nextFixtureKey("session-revoke"), map[string]any{
		"scope": "one", "session_id": account.sessionID.String(), "reauthentication": map[string]any{"method": "password", "password": account.password},
	}, http.StatusNoContent)
	_, _ = client.post(account.baseURL, "/v1/device-enrollment-grants", account.accessToken, nextFixtureKey("revoked-session"), map[string]any{
		"reauthentication": map[string]any{"method": "password", "password": account.password},
	}, http.StatusUnauthorized)
	account.accessToken, account.refreshToken, account.sessionID = second.accessToken, second.refreshToken, second.sessionID
}

func (fixture *c11Fixture) runDeviceAndTrust(t *testing.T, client *c11HTTPClient, account c11Account) c11Device {
	t.Helper()
	grantBody, _ := client.post(account.baseURL, "/v1/device-enrollment-grants", account.accessToken, nextFixtureKey("enrollment-grant"), map[string]any{
		"reauthentication": map[string]any{"method": "password", "password": account.password},
	}, http.StatusCreated)
	grant := stringField(decodeObject(t, grantBody), "enrollment_grant")
	device := fixture.registerDevice(t, client, account.baseURL, grant)
	_, _ = client.post(account.baseURL, "/v1/config-bundle-resolutions", account.accessToken, nextFixtureKey("wrong-account-domain"), map[string]any{}, http.StatusUnauthorized)
	_, _ = client.post(account.baseURL, "/v1/device-enrollment-grants", device.accessToken, nextFixtureKey("wrong-device-domain"), map[string]any{
		"reauthentication": map[string]any{"method": "password", "password": account.password},
	}, http.StatusUnauthorized)
	device = fixture.rotateCompromiseAndReenroll(t, client, account, device)
	resolutionBody, _ := client.post(account.baseURL, "/v1/config-bundle-resolutions", device.accessToken, nextFixtureKey("bundle-resolve"), map[string]any{}, http.StatusOK)
	device.resolution = decodeObject(t, resolutionBody)
	return device
}

func TestClearFixtureSigningKeyUnlessTransferredIsFailClosed(t *testing.T) {
	t.Parallel()
	t.Run("not-transferred", func(t *testing.T) {
		key := ed25519.PrivateKey{0x11, 0x12}
		func() {
			transferred := false
			defer clearFixtureSigningKeyUnlessTransferred(key, &transferred)
		}()
		if !bytes.Equal(key, []byte{0, 0}) {
			t.Fatal("fixture signing key guard retained untransferred ownership")
		}
	})
	t.Run("missing-flag", func(t *testing.T) {
		key := ed25519.PrivateKey{0x21, 0x22}
		func() {
			defer clearFixtureSigningKeyUnlessTransferred(key, nil)
		}()
		if !bytes.Equal(key, []byte{0, 0}) {
			t.Fatal("fixture signing key guard accepted missing ownership state")
		}
	})
	t.Run("transferred", func(t *testing.T) {
		key := ed25519.PrivateKey{0x31, 0x32}
		func() {
			transferred := false
			defer clearFixtureSigningKeyUnlessTransferred(key, &transferred)
			transferred = true
		}()
		if !bytes.Equal(key, []byte{0x31, 0x32}) {
			t.Fatal("fixture signing key guard cleared transferred ownership")
		}
	})
}

type preparedDeviceRegistration struct {
	request       map[string]any
	signingPublic ed25519.PublicKey
	signingKey    ed25519.PrivateKey
	hpkePrivate   *ecdh.PrivateKey
}

func clearFixtureSigningKeyUnlessTransferred(key ed25519.PrivateKey, transferred *bool) {
	if transferred == nil || !*transferred {
		clear(key)
	}
}

func (fixture *c11Fixture) prepareDeviceRegistration(t *testing.T, client *c11HTTPClient, baseURL, grant string) preparedDeviceRegistration {
	t.Helper()
	audience, err := fixture.deviceProofAudienceFor(baseURL)
	if err != nil {
		t.Fatal("c11 device proof audience invalid")
	}
	signingPublic, signingPrivate, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal("c11 device signing key generation failed")
	}
	transferred := false
	defer clearFixtureSigningKeyUnlessTransferred(signingPrivate, &transferred)
	hpkePrivate, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		clear(signingPrivate)
		t.Fatal("c11 device HPKE key generation failed")
	}
	requestNonce := random32(t)
	challengeBody := map[string]any{
		"enrollment_grant": grant, "request_nonce": encode32(requestNonce), "signing_public_key": base64.RawURLEncoding.EncodeToString(signingPublic),
		"hpke_public_key": base64.RawURLEncoding.EncodeToString(hpkePrivate.PublicKey().Bytes()),
	}
	challengeEncoded, _ := client.post(baseURL, "/v1/device-auth-challenges", "", nextFixtureKey("device-challenge"), challengeBody, http.StatusCreated)
	challenge := decodeObject(t, challengeEncoded)
	challengeBytes := decode32(t, stringField(challenge, "challenge"))
	grantToken, err := securitykit.DecodeOpaqueToken(grant)
	if err != nil {
		t.Fatal("c11 enrollment grant invalid")
	}
	grantDigest := securitykit.DigestToken(securitykit.EnrollmentGrantToken, grantToken)
	grantToken.Clear()
	var signingArray, hpkeArray [32]byte
	copy(signingArray[:], signingPublic)
	copy(hpkeArray[:], hpkePrivate.PublicKey().Bytes())
	proof := deviceauth.ProofBytes(deviceauth.ProofInput{
		ProtocolVersion: "device-pop-v1", Challenge: challengeBytes, GrantDigest: grantDigest, SigningPublicKey: signingArray,
		HPKEPublicKey: hpkeArray, Operation: "register_device", Audience: audience, RequestNonce: requestNonce,
	})
	signature := ed25519.Sign(signingPrivate, proof)
	clear(proof)
	request := map[string]any{
		"enrollment_grant": grant, "challenge_id": stringField(challenge, "challenge_id"), "request_nonce": encode32(requestNonce),
		"signing_public_key": base64.RawURLEncoding.EncodeToString(signingPublic), "hpke_public_key": base64.RawURLEncoding.EncodeToString(hpkePrivate.PublicKey().Bytes()),
		"display_name": "C1.1 E2E device", "signature": base64.RawURLEncoding.EncodeToString(signature),
	}
	clear(signature)
	prepared := preparedDeviceRegistration{request: request, signingPublic: signingPublic, signingKey: signingPrivate, hpkePrivate: hpkePrivate}
	transferred = true
	return prepared
}

func cloneFixtureObject(source map[string]any) map[string]any {
	result := make(map[string]any, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}

func (fixture *c11Fixture) registerDevice(t *testing.T, client *c11HTTPClient, baseURL, grant string) c11Device {
	t.Helper()
	challengeReplay := fixture.prepareDeviceRegistration(t, client, baseURL, grant)
	invalidProof := cloneFixtureObject(challengeReplay.request)
	invalidProof["signature"] = base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x19}, ed25519.SignatureSize))
	_, _ = client.post(baseURL, "/v1/devices", "", nextFixtureKey("challenge-consume"), invalidProof, http.StatusUnauthorized)
	_, _ = client.post(baseURL, "/v1/devices", "", nextFixtureKey("challenge-replay"), challengeReplay.request, http.StatusUnauthorized)
	clear(challengeReplay.signingKey)

	registration := fixture.prepareDeviceRegistration(t, client, baseURL, grant)
	anonymous := cloneFixtureObject(registration.request)
	anonymous["enrollment_grant"] = base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x29}, 32))
	_, _ = client.post(baseURL, "/v1/devices", "", nextFixtureKey("anonymous-fresh"), anonymous, http.StatusUnauthorized)
	registered, _ := client.post(baseURL, "/v1/devices", "", nextFixtureKey("device-register"), registration.request, http.StatusCreated)
	object := decodeObject(t, registered)
	deviceID, deviceErr := uuid.Parse(stringField(object, "device_id"))
	authorizationID, authorizationErr := uuid.Parse(stringField(object, "authorization_id"))
	familyID, familyErr := uuid.Parse(stringField(object, "family_id"))
	if deviceErr != nil || authorizationErr != nil || familyErr != nil {
		clear(registration.signingKey)
		t.Fatal("c11 device token response invalid")
	}
	grantReplayNonce := random32(t)
	_, _ = client.post(baseURL, "/v1/device-auth-challenges", "", nextFixtureKey("grant-replay"), map[string]any{
		"enrollment_grant": grant, "request_nonce": encode32(grantReplayNonce),
		"signing_public_key": base64.RawURLEncoding.EncodeToString(registration.signingPublic),
		"hpke_public_key":    base64.RawURLEncoding.EncodeToString(registration.hpkePrivate.PublicKey().Bytes()),
	}, http.StatusUnauthorized)
	t.Cleanup(func() { clear(registration.signingKey) })
	return c11Device{
		baseURL: baseURL, accessToken: stringField(object, "access_token"), refreshToken: stringField(object, "refresh_token"),
		deviceID: deviceID, authorizationID: authorizationID, familyID: familyID, hpkePrivate: registration.hpkePrivate,
		signingPublic: registration.signingPublic, signingKey: registration.signingKey, expiresAt: stringField(object, "expires_at"),
	}
}

func (fixture *c11Fixture) prepareDeviceRotation(t *testing.T, client *c11HTTPClient, device c11Device) map[string]any {
	t.Helper()
	audience, err := fixture.deviceProofAudienceFor(device.baseURL)
	if err != nil {
		t.Fatal("c11 device rotation proof audience invalid")
	}
	nonce := random32(t)
	challengeBody, _ := client.post(device.baseURL, "/v1/device-auth-challenges", "", nextFixtureKey("rotation-challenge"), map[string]any{
		"refresh_token": device.refreshToken, "request_nonce": encode32(nonce),
	}, http.StatusCreated)
	challenge := decodeObject(t, challengeBody)
	proof := deviceauth.DeviceRotationProofBytes(deviceauth.DeviceRotationProofInput{
		ProtocolVersion: "device-token-rotation-v1", Challenge: decode32(t, stringField(challenge, "challenge")),
		FamilyID: device.familyID, Operation: "rotate_device_token", Audience: audience, RequestNonce: nonce,
	})
	signature := ed25519.Sign(device.signingKey, proof)
	clear(proof)
	defer clear(signature)
	return map[string]any{
		"refresh_token": device.refreshToken, "challenge_id": stringField(challenge, "challenge_id"),
		"request_nonce": encode32(nonce), "signature": base64.RawURLEncoding.EncodeToString(signature),
	}
}

func (fixture *c11Fixture) rotateDeviceOnce(t *testing.T, client *c11HTTPClient, device c11Device) c11Device {
	t.Helper()
	command := fixture.prepareDeviceRotation(t, client, device)
	body, _ := client.post(device.baseURL, "/v1/device-token-rotations", "", nextFixtureKey("grace-device-rotation"), command, http.StatusOK)
	object := decodeObject(t, body)
	deviceID, deviceErr := uuid.Parse(stringField(object, "device_id"))
	authorizationID, authorizationErr := uuid.Parse(stringField(object, "authorization_id"))
	familyID, familyErr := uuid.Parse(stringField(object, "family_id"))
	if deviceErr != nil || authorizationErr != nil || familyErr != nil || deviceID != device.deviceID ||
		authorizationID != device.authorizationID || familyID != device.familyID {
		t.Fatal("c11 grace device rotation changed fixed authority identifiers")
	}
	device.accessToken = stringField(object, "access_token")
	device.refreshToken = stringField(object, "refresh_token")
	device.expiresAt = stringField(object, "expires_at")
	if device.accessToken == "" || device.refreshToken == "" || device.expiresAt == "" {
		t.Fatal("c11 grace device rotation response invalid")
	}
	return device
}

func (fixture *c11Fixture) resolveBundlePolicy(
	t *testing.T,
	client *c11HTTPClient,
	device c11Device,
	now time.Time,
) string {
	t.Helper()
	resolutionBody, _ := client.post(device.baseURL, "/v1/config-bundle-resolutions", device.accessToken, nextFixtureKey("grace-bundle"), map[string]any{}, http.StatusOK)
	resolution := decodeObject(t, resolutionBody)
	locatorValue := stringField(resolution, "bundle_locator")
	locations, _ := resolution["locations"].([]any)
	if !validBundleSourceLocations(locations, fixture.expectedSourceURLs, locatorValue, false) {
		t.Fatal("c11 grace bundle locations invalid")
	}
	envelope, _ := client.get(fmt.Sprint(locations[0]), http.StatusOK)
	metadataBytes, metadataErr := base64.RawURLEncoding.DecodeString(fixture.ready.Metadata)
	rootPublic, rootErr := base64.RawURLEncoding.DecodeString(fixture.ready.RootPublic)
	metadata, trustErr := trustclient.NewTrustMetadata(metadataBytes, map[string]ed25519.PublicKey{fixture.ready.RootKeyID: rootPublic}, 0)
	locator := decode32(t, locatorValue)
	expected, expectedErr := trustclient.NewExpected(device.authorizationID.String(), locator)
	if metadataErr != nil || rootErr != nil || trustErr != nil || expectedErr != nil {
		t.Fatal("c11 grace trust expectation invalid")
	}
	verified, err := verifyReferenceBundle(t.Context(), envelope, device.hpkePrivate, metadata, expected, trustclient.NewMemoryStore(), now)
	if err != nil {
		t.Fatal("c11 grace reference bundle verification failed")
	}
	payload, err := trust.DecodePayloadV1(verified.PayloadJCS())
	if err != nil {
		t.Fatal("c11 grace verified bundle payload invalid")
	}
	return string(payload.PolicySnapshot)
}

type c11HTTPResult struct {
	body   []byte
	status int
	err    error
}

func doC11JSONPost(
	ctx context.Context,
	client *http.Client,
	baseURL, path, bearer, idempotency string,
	body any,
) c11HTTPResult {
	encoded, err := json.Marshal(body)
	if err != nil {
		return c11HTTPResult{err: errors.New("c11 HTTP encoding failed")}
	}
	request, err := newC11POSTRequest(ctx, baseURL+path, encoded)
	if err != nil {
		return c11HTTPResult{err: errors.New("c11 HTTP request failed")}
	}
	request.Header.Set("Content-Type", "application/json")
	if bearer != "" {
		request.Header.Set("Authorization", "Bearer "+bearer)
	}
	if idempotency != "" {
		request.Header.Set("Idempotency-Key", idempotency)
	}
	response, err := client.Do(request)
	if err != nil {
		return c11HTTPResult{err: errors.New("c11 HTTP transport failed")}
	}
	defer func() { _ = response.Body.Close() }()
	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return c11HTTPResult{err: errors.New("c11 HTTP response failed")}
	}
	return c11HTTPResult{body: responseBody, status: response.StatusCode}
}

func (fixture *c11Fixture) rotateCompromiseAndReenroll(
	t *testing.T,
	client *c11HTTPClient,
	account c11Account,
	device c11Device,
) c11Device {
	t.Helper()
	commands := []map[string]any{
		fixture.prepareDeviceRotation(t, client, device),
		fixture.prepareDeviceRotation(t, client, device),
	}
	keys := []string{nextFixtureKey("rotation-winner-a"), nextFixtureKey("rotation-winner-b")}
	start := make(chan struct{})
	results := make(chan c11HTTPResult, 2)
	for index := range commands {
		index := index
		go func() {
			<-start
			results <- doC11JSONPost(t.Context(), fixture.client, device.baseURL, "/v1/device-token-rotations", "", keys[index], commands[index])
		}()
	}
	close(start)
	first, second := <-results, <-results
	responses := []c11HTTPResult{first, second}
	successes, failures := 0, 0
	var rotated map[string]any
	for _, response := range responses {
		if response.err != nil {
			t.Fatal("c11 concurrent device rotation transport failed")
		}
		switch response.status {
		case http.StatusOK:
			successes++
			rotated = decodeObject(t, response.body)
		case http.StatusUnauthorized:
			failures++
		default:
			t.Fatalf("c11 concurrent device rotation returned status %d", response.status)
		}
		clear(response.body)
	}
	if successes != 1 || failures != 1 || stringField(rotated, "family_id") != device.familyID.String() {
		t.Fatal("c11 concurrent device rotation did not produce one winner and one family-compromise replay")
	}
	_, _ = client.post(device.baseURL, "/v1/config-bundle-resolutions", stringField(rotated, "access_token"), nextFixtureKey("compromised-successor"), map[string]any{}, http.StatusUnauthorized)
	_, _ = client.post(device.baseURL, "/v1/config-bundle-resolutions", device.accessToken, nextFixtureKey("compromised-original"), map[string]any{}, http.StatusUnauthorized)
	_, _ = client.post(account.baseURL, "/v1/device-revocations", account.accessToken, nextFixtureKey("compromise-revoke"), map[string]any{
		"device_id":        device.deviceID.String(),
		"reauthentication": map[string]any{"method": "password", "password": account.password},
	}, http.StatusNoContent)
	_, _ = client.post(account.baseURL, "/v1/device-enrollment-grants", account.accessToken, nextFixtureKey("revoked-bound-grant"), map[string]any{
		"reauthentication": map[string]any{"method": "password", "password": account.password},
	}, http.StatusUnauthorized)
	loginBody, _ := client.post(account.baseURL, "/v1/account-sessions", "", nextFixtureKey("compromise-login"), map[string]any{
		"method": "password", "email": account.email, "password": account.password,
		"client_signing_public_key": base64.RawURLEncoding.EncodeToString(account.signingPublic),
	}, http.StatusOK)
	freshAccount := decodeAccountTokens(t, account.baseURL, account.email, account.password, account.signingPublic, account.signingKey, loginBody)
	if freshAccount.principalID != account.principalID || freshAccount.sessionID == uuid.Nil || freshAccount.sessionID == account.sessionID {
		t.Fatal("c11 device recovery account authority invalid")
	}
	grantBody, _ := client.post(freshAccount.baseURL, "/v1/device-enrollment-grants", freshAccount.accessToken, nextFixtureKey("compromise-reenroll"), map[string]any{
		"reauthentication": map[string]any{"method": "password", "password": freshAccount.password},
	}, http.StatusCreated)
	reenrolled := fixture.registerDevice(t, client, freshAccount.baseURL, stringField(decodeObject(t, grantBody), "enrollment_grant"))
	if reenrolled.deviceID == device.deviceID || reenrolled.authorizationID == device.authorizationID || reenrolled.familyID == device.familyID {
		t.Fatal("c11 device reenrollment reused compromised authority")
	}
	return reenrolled
}

func (fixture *c11Fixture) verifyBundle(t *testing.T, client *c11HTTPClient, device c11Device) {
	t.Helper()
	locatorValue := stringField(device.resolution, "bundle_locator")
	locator := decode32(t, locatorValue)
	locations, _ := device.resolution["locations"].([]any)
	if !validBundleSourceLocations(locations, fixture.expectedSourceURLs, locatorValue, fixture.externalRuntime) {
		t.Fatal("c11 bundle locations invalid")
	}
	bodies := make([][]byte, 0, 3)
	for _, location := range locations {
		body, _ := client.get(fmt.Sprint(location), http.StatusOK)
		bodies = append(bodies, body)
	}
	digest := sha256.Sum256(bodies[0])
	for _, body := range bodies[1:] {
		if sha256.Sum256(body) != digest || !bytes.Equal(body, bodies[0]) {
			t.Fatal("c11 bundle sources were not byte-identical")
		}
	}
	digestClaim := stringField(device.resolution, "envelope_sha256")
	if !validResolutionEnvelopeDigest(digestClaim, bodies[0]) {
		t.Fatal("c11 bundle digest response mismatch")
	}
	tamperedDigestClaim := []byte(digestClaim)
	if tamperedDigestClaim[0] == '0' {
		tamperedDigestClaim[0] = '1'
	} else {
		tamperedDigestClaim[0] = '0'
	}
	if validResolutionEnvelopeDigest(string(tamperedDigestClaim), bodies[0]) {
		t.Fatal("c11 bundle digest tamper was accepted")
	}
	metadataBytes, _ := base64.RawURLEncoding.DecodeString(fixture.ready.Metadata)
	rootPublic, _ := base64.RawURLEncoding.DecodeString(fixture.ready.RootPublic)
	metadata, err := trustclient.NewTrustMetadata(metadataBytes, map[string]ed25519.PublicKey{fixture.ready.RootKeyID: rootPublic}, 0)
	if err != nil {
		t.Fatal("c11 trust metadata provisioning failed")
	}
	expected, err := trustclient.NewExpected(device.authorizationID.String(), locator)
	if err != nil {
		t.Fatal("c11 trust expectation failed")
	}
	now := fixture.authorityTime(t)
	store := trustclient.NewMemoryStore()
	verified, err := verifyReferenceBundle(t.Context(), bodies[0], device.hpkePrivate, metadata, expected, store, now)
	if err != nil {
		t.Fatal("c11 reference client verification failed")
	}
	var payload map[string]any
	if err := json.Unmarshal(verified.PayloadJCS(), &payload); err != nil {
		t.Fatal("c11 verified bundle payload invalid")
	}
	_, _ = client.post(device.baseURL, "/v1/config-bundle-acknowledgements", device.accessToken, nextFixtureKey("bundle-ack"), map[string]any{
		"bundle_id": stringField(payload, "bundle_id"), "bundle_version": stringField(payload, "bundle_version"),
	}, http.StatusNoContent)
	var envelopeFields map[string]string
	if err := json.Unmarshal(bodies[0], &envelopeFields); err != nil {
		t.Fatal("c11 envelope tamper fixture invalid")
	}
	ciphertext, ciphertextErr := base64.RawURLEncoding.DecodeString(envelopeFields["ciphertext"])
	if ciphertextErr != nil || len(ciphertext) == 0 {
		clear(ciphertext)
		t.Fatal("c11 envelope ciphertext fixture invalid")
	}
	ciphertext[len(ciphertext)/2] ^= 1
	tamperRows := []struct {
		name  string
		field string
		value string
	}{
		{name: "envelope version", field: "envelope_version", value: "future"},
		{name: "KEM", field: "kem", value: "DHKEM-P256-HKDF-SHA256"},
		{name: "KDF", field: "kdf", value: "HKDF-SHA512"},
		{name: "AEAD", field: "aead", value: "AES-128-GCM"},
		{name: "recipient selector", field: "recipient_key_id", value: base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x41}, 16))},
		{name: "bundle locator", field: "bundle_locator", value: base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x42}, 32))},
		{name: "encapsulated key", field: "enc", value: base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x43}, 32))},
		{name: "ciphertext", field: "ciphertext", value: base64.RawURLEncoding.EncodeToString(ciphertext)},
	}
	clear(ciphertext)
	for _, row := range tamperRows {
		changed := mutateFixtureEnvelope(t, bodies[0], row.field, row.value)
		if _, err := verifyReferenceBundle(t.Context(), changed, device.hpkePrivate, metadata, expected, trustclient.NewMemoryStore(), now); err == nil {
			t.Fatalf("c11 reference client accepted %s tamper", row.name)
		}
	}
	wrongRecipient, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal("c11 wrong-recipient fixture failed")
	}
	if _, err := verifyReferenceBundle(t.Context(), bodies[0], wrongRecipient, metadata, expected, trustclient.NewMemoryStore(), now); err == nil {
		t.Fatal("c11 reference client accepted wrong recipient key")
	}
	wrongAudience, err := trustclient.NewExpected(uuid.NewString(), locator)
	if err != nil {
		t.Fatal("c11 wrong-audience fixture failed")
	}
	if _, err := verifyReferenceBundle(t.Context(), bodies[0], device.hpkePrivate, metadata, wrongAudience, trustclient.NewMemoryStore(), now); err == nil {
		t.Fatal("c11 reference client accepted wrong audience")
	}
	wrongLocator := random32(t)
	wrongLocatorExpectation, err := trustclient.NewExpected(device.authorizationID.String(), wrongLocator)
	if err != nil {
		t.Fatal("c11 wrong-locator fixture failed")
	}
	if _, err := verifyReferenceBundle(t.Context(), bodies[0], device.hpkePrivate, metadata, wrongLocatorExpectation, trustclient.NewMemoryStore(), now); err == nil {
		t.Fatal("c11 reference client accepted wrong locator")
	}
	tamperedMetadataBytes := mutateFixtureSignedMetadata(t, metadataBytes)
	tamperedMetadata, err := trustclient.NewTrustMetadata(tamperedMetadataBytes, map[string]ed25519.PublicKey{fixture.ready.RootKeyID: rootPublic}, 0)
	clear(tamperedMetadataBytes)
	if err != nil {
		t.Fatal("c11 metadata tamper fixture failed")
	}
	if _, err := verifyReferenceBundle(t.Context(), bodies[0], device.hpkePrivate, tamperedMetadata, expected, trustclient.NewMemoryStore(), now); err == nil {
		t.Fatal("c11 reference client accepted tampered root metadata")
	}
	if _, err := verifyReferenceBundle(t.Context(), bodies[0], device.hpkePrivate, metadata, expected, trustclient.NewMemoryStore(), now.Add(48*time.Hour)); err == nil {
		t.Fatal("c11 reference client accepted expired bundle")
	}
	if _, err := verifyReferenceBundle(t.Context(), bodies[0], device.hpkePrivate, metadata, expected, store, now); !errors.Is(err, trustclient.ErrRollback) {
		t.Fatal("c11 reference client accepted rollback")
	}
	fixture.verifyConformanceBinary(t, bodies[0], metadataBytes, rootPublic, locator, device)
}

func validResolutionEnvelopeDigest(claim string, envelope []byte) bool {
	if len(claim) != sha256.Size*2 || len(envelope) == 0 {
		return false
	}
	digest := sha256.Sum256(envelope)
	return claim == fmt.Sprintf("%x", digest)
}

func verifyReferenceBundle(
	parent context.Context,
	envelope []byte,
	privateKey ecdh.KeyExchanger,
	metadata trustclient.TrustMetadata,
	expected trustclient.Expected,
	store trustclient.Store,
	now time.Time,
) (trustclient.VerifiedBundle, error) {
	ctx, cancel := context.WithTimeout(parent, fixtureRequestTimeout)
	defer cancel()
	return trustclient.VerifyAndStage(ctx, envelope, privateKey, metadata, expected, store, now, 120*time.Second)
}

func mutateFixtureEnvelope(t *testing.T, body []byte, field, value string) []byte {
	t.Helper()
	var envelope map[string]string
	if err := json.Unmarshal(body, &envelope); err != nil || len(envelope) != 8 {
		t.Fatal("c11 envelope mutation decode failed")
	}
	if _, exists := envelope[field]; !exists {
		t.Fatal("c11 envelope mutation field missing")
	}
	envelope[field] = value
	encoded, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal("c11 envelope mutation encode failed")
	}
	return encoded
}

func mutateFixtureSignedMetadata(t *testing.T, body []byte) []byte {
	t.Helper()
	var signed map[string]string
	if err := json.Unmarshal(body, &signed); err != nil || len(signed) != 2 || signed["signature"] == "" {
		t.Fatal("c11 metadata mutation decode failed")
	}
	signature := signed["signature"]
	if signature[0] == 'A' {
		signed["signature"] = "B" + signature[1:]
	} else {
		signed["signature"] = "A" + signature[1:]
	}
	encoded, err := json.Marshal(signed)
	if err != nil {
		t.Fatal("c11 metadata mutation encode failed")
	}
	return encoded
}

func validBundleSourceLocations(locations []any, expected [3]string, locator string, _ bool) bool {
	if len(locations) != len(expected) || locator == "" {
		return false
	}
	expectedPath := "/b/" + url.PathEscape(locator)
	seen := make(map[string]struct{}, len(locations))
	for index, raw := range locations {
		location, ok := raw.(string)
		if !ok || location == "" {
			return false
		}
		actual, err := url.ParseRequestURI(location)
		if err != nil || actual.Scheme != "http" || actual.User != nil || actual.RawQuery != "" || actual.Fragment != "" ||
			actual.ForceQuery || actual.EscapedPath() != expectedPath || actual.Port() == "" || !validFixturePort(actual.Port()) {
			return false
		}
		ownedURL, err := url.ParseRequestURI(expected[index])
		if err != nil {
			return false
		}
		owned, err := parseConfiguredBundleOrigin(expected[index], ownedURL.Port())
		if err != nil || actual.Scheme != owned.Scheme || actual.Host != owned.Host || actual.Hostname() != owned.Hostname() ||
			actual.Port() != owned.Port() {
			return false
		}
		identity := actual.Scheme + "://" + actual.Host
		if _, duplicate := seen[identity]; duplicate {
			return false
		}
		seen[identity] = struct{}{}
	}
	return true
}

func (fixture *c11Fixture) verifyConformanceBinary(
	t *testing.T,
	envelope []byte,
	signedMetadata []byte,
	rootPublic ed25519.PublicKey,
	locator [32]byte,
	device c11Device,
) {
	t.Helper()
	absolute, err := validatedConformanceBinaryPath(os.Getenv(fixtureConformancePath))
	if err != nil {
		t.Fatal("c11 conformance binary unavailable")
	}
	metadataFile := struct {
		SignedMetadata json.RawMessage `json:"signed_metadata"`
		TrustedRoots   []struct {
			KeyID     string `json:"key_id"`
			PublicKey string `json:"public_key"`
		} `json:"trusted_roots"`
		HighestTrusted        string `json:"highest_trusted_version"`
		ExpectedAudience      string `json:"expected_audience"`
		ExpectedBundleLocator string `json:"expected_bundle_locator"`
	}{
		SignedMetadata: json.RawMessage(append([]byte(nil), signedMetadata...)),
		TrustedRoots: []struct {
			KeyID     string `json:"key_id"`
			PublicKey string `json:"public_key"`
		}{{KeyID: fixture.ready.RootKeyID, PublicKey: base64.RawURLEncoding.EncodeToString(rootPublic)}},
		HighestTrusted: "0", ExpectedAudience: device.authorizationID.String(),
		ExpectedBundleLocator: base64.RawURLEncoding.EncodeToString(locator[:]),
	}
	metadataBody, err := json.Marshal(metadataFile)
	if err != nil {
		t.Fatal("c11 conformance metadata encoding failed")
	}
	defer clear(metadataBody)
	directory := t.TempDir()
	envelopePath := filepath.Join(directory, "envelope.json")
	metadataPath := filepath.Join(directory, "metadata.json")
	privateKeyPath := filepath.Join(directory, "recipient.key")
	if err := os.WriteFile(envelopePath, envelope, 0o600); err != nil {
		t.Fatal("c11 conformance envelope fixture failed")
	}
	if err := os.WriteFile(metadataPath, metadataBody, 0o600); err != nil {
		t.Fatal("c11 conformance metadata fixture failed")
	}
	privateBody := []byte(base64.RawURLEncoding.EncodeToString(device.hpkePrivate.Bytes()) + "\n")
	if err := os.WriteFile(privateKeyPath, privateBody, 0o600); err != nil {
		clear(privateBody)
		t.Fatal("c11 conformance private-key fixture failed")
	}
	clear(privateBody)
	ctx, cancel := context.WithTimeout(t.Context(), fixtureRequestTimeout)
	defer cancel()
	command := exec.CommandContext(ctx, absolute, //nolint:gosec // Exact absolute smoke-built binary selected by the local acceptance script.
		"-envelope", envelopePath, "-metadata", metadataPath, "-private-key", privateKeyPath, "-state-dir", filepath.Join(directory, "state"),
	)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil || stdout.String() != "verified\n" || stderr.Len() != 0 {
		t.Fatal("c11 conformance binary rejected the verified fixture")
	}
}

func validatedConformanceBinaryPath(binary string) (string, error) {
	if binary == "" || !filepath.IsAbs(binary) || filepath.Clean(binary) != binary {
		return "", errors.New("conformance binary path invalid")
	}
	information, err := os.Stat(binary) //nolint:gosec // Explicit absolute local path; no path is derived from HTTP or untrusted fixture data.
	if err != nil || !information.Mode().IsRegular() {
		return "", errors.New("conformance binary unavailable")
	}
	return binary, nil
}

func random32(t *testing.T) [32]byte {
	t.Helper()
	var value [32]byte
	if _, err := io.ReadFull(rand.Reader, value[:]); err != nil {
		t.Fatal("c11 random fixture failed")
	}
	return value
}

func encode32(value [32]byte) string { return base64.RawURLEncoding.EncodeToString(value[:]) }

func decode32(t *testing.T, encoded string) [32]byte {
	t.Helper()
	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil || len(decoded) != 32 {
		t.Fatal("c11 base64url32 response invalid")
	}
	var result [32]byte
	copy(result[:], decoded)
	clear(decoded)
	return result
}

func TestFixtureChildCloseShutsDownBeforeLifecycleCancel(t *testing.T) {
	t.Parallel()
	requestReader, requestWriter := io.Pipe()
	responseReader, responseWriter := io.Pipe()
	events := make(chan string, 2)
	child := &fixtureChild{
		input: requestWriter, output: responseReader, encoder: json.NewEncoder(requestWriter),
		decoder: json.NewDecoder(responseReader), wait: make(chan error, 1), shutdownTimeout: time.Second,
		cancel: func() { events <- "cancel" },
	}
	go func() {
		defer func() { _ = requestReader.Close() }()
		defer func() { _ = responseWriter.Close() }()
		var request fixtureRequest
		if json.NewDecoder(requestReader).Decode(&request) == nil && request.Operation == "shutdown" {
			events <- "shutdown"
			child.wait <- nil
			return
		}
		child.wait <- errors.New("invalid shutdown request")
	}()

	if err := child.close(); err != nil {
		t.Fatalf("fixture child close failed: %v", err)
	}
	if first, second := <-events, <-events; first != "shutdown" || second != "cancel" {
		t.Fatalf("fixture lifecycle events = %q/%q, want shutdown/cancel", first, second)
	}
}

func TestFixtureChildCallHonorsContextDeadline(t *testing.T) {
	t.Parallel()
	requestReader, requestWriter := io.Pipe()
	responseReader, responseWriter := io.Pipe()
	defer func() { _ = requestReader.Close() }()
	defer func() { _ = responseWriter.Close() }()
	canceled := make(chan struct{})
	child := &fixtureChild{
		input: requestWriter, output: responseReader, encoder: json.NewEncoder(requestWriter),
		decoder: json.NewDecoder(responseReader), wait: make(chan error, 1), shutdownTimeout: time.Second,
		cancel: func() { close(canceled) },
	}
	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()
	started := time.Now()
	if _, err := child.callContext(ctx, "never-responds", "", "", ""); err == nil {
		t.Fatal("fixture pipe call unexpectedly succeeded")
	}
	if elapsed := time.Since(started); elapsed > 500*time.Millisecond {
		t.Fatalf("fixture pipe deadline took %s", elapsed)
	}
	select {
	case <-canceled:
	case <-time.After(time.Second):
		t.Fatal("fixture pipe deadline did not cancel the child lifecycle")
	}
}

func TestValidatedConformanceBinaryPathRequiresExplicitFile(t *testing.T) {
	t.Parallel()
	if _, err := validatedConformanceBinaryPath(""); err == nil {
		t.Fatal("missing conformance binary path unexpectedly accepted")
	}
}

func TestLoadExternalRuntimeConfigurationRequiresThreeOwnedListeners(t *testing.T) {
	t.Parallel()
	valid := map[string]string{
		"C11_E2E_EXTERNAL_RUNTIME": "1",
		"C11_E2E_PRIMARY_URL":      "http://127.0.0.1:8080",
		"C11_E2E_MIRROR_A_URL":     "http://127.0.0.1:8081",
		"C11_E2E_MIRROR_B_URL":     "http://127.0.0.1:8082",
		"C11_E2E_PRIMARY_ORIGIN":   "http://localhost:8080",
		"C11_E2E_MIRROR_A_ORIGIN":  "http://localhost:8081",
		"C11_E2E_MIRROR_B_ORIGIN":  "http://localhost:8082",
	}
	configuration, enabled, err := loadExternalRuntimeConfiguration(mapLookup(valid))
	if err != nil || !enabled || configuration.PrimaryURL != valid["C11_E2E_PRIMARY_URL"] ||
		configuration.MirrorAURL != valid["C11_E2E_MIRROR_A_URL"] || configuration.MirrorBURL != valid["C11_E2E_MIRROR_B_URL"] {
		t.Fatal("valid external runtime contract was rejected")
	}
	fixture, cleanup, fixtureErr := newExternalC11Fixture(configuration, mapLookup(nil))
	if fixtureErr != nil {
		t.Fatal("valid external runtime fixture was rejected")
	}
	defer cleanup()
	wantOrigins := [3]string{valid["C11_E2E_PRIMARY_ORIGIN"], valid["C11_E2E_MIRROR_A_ORIGIN"], valid["C11_E2E_MIRROR_B_ORIGIN"]}
	if fixture.expectedSourceURLs != wantOrigins {
		t.Fatal("external runtime conflated listener endpoints with configured bundle origins")
	}

	for name, value := range map[string]string{ //nolint:gosec // Synthetic invalid URL credentials verify rejection only.
		"missing mirror": "",
		"remote primary": "http://192.0.2.19:8080",
		"credential URL": "http://user:secret@127.0.0.1:8080",
		"duplicate port": "http://127.0.0.1:8081",
	} {
		name, value := name, value
		t.Run(name, func(t *testing.T) {
			values := make(map[string]string, len(valid))
			for key, original := range valid {
				values[key] = original
			}
			switch name {
			case "missing mirror":
				values["C11_E2E_MIRROR_B_URL"] = value
			default:
				values["C11_E2E_PRIMARY_URL"] = value
			}
			if _, enabled, err := loadExternalRuntimeConfiguration(mapLookup(values)); err == nil || !enabled {
				t.Fatal("invalid external runtime contract was accepted")
			}
		})
	}
	for _, name := range []string{"C11_E2E_PRIMARY_ORIGIN", "C11_E2E_MIRROR_A_ORIGIN", "C11_E2E_MIRROR_B_ORIGIN"} {
		values := make(map[string]string, len(valid)-1)
		for key, original := range valid {
			if key != name {
				values[key] = original
			}
		}
		if _, enabled, err := loadExternalRuntimeConfiguration(mapLookup(values)); err == nil || !enabled {
			t.Fatalf("external runtime without %s was accepted", name)
		}
	}
	for name, value := range map[string]string{
		"listener alias":    "http://127.0.0.1:8080",
		"wrong port":        "http://localhost:8081",
		"path":              "http://localhost:8080/b/private",
		"query":             "http://localhost:8080?private=1",
		"userinfo":          "http://user@localhost:8080",
		"fragment":          "http://localhost:8080#private",
		"noncanonical port": "http://localhost:08080",
	} {
		name, value := name, value
		t.Run("origin/"+name, func(t *testing.T) {
			values := make(map[string]string, len(valid))
			for key, original := range valid {
				values[key] = original
			}
			values["C11_E2E_PRIMARY_ORIGIN"] = value
			if _, enabled, err := loadExternalRuntimeConfiguration(mapLookup(values)); err == nil || !enabled {
				t.Fatal("invalid configured bundle origin was accepted")
			}
		})
	}
	if _, enabled, err := loadExternalRuntimeConfiguration(mapLookup(nil)); err != nil || enabled {
		t.Fatal("absent external runtime contract did not remain disabled")
	}
}

func TestExternalEnumerationUsesDistinctGenericAndNeverRegisteredAddresses(t *testing.T) {
	t.Parallel()
	genericProbe, neverRegistered := externalEnumerationProbeEmails(19)
	if genericProbe == "" || neverRegistered == "" || genericProbe == neverRegistered {
		t.Fatal("external enumeration fixture reused its registration probe for the missing-login probe")
	}
}

func TestExternalEnumerationComparesPublicErrorsIgnoringOnlyTraceID(t *testing.T) {
	t.Parallel()
	left := controlapiv1.PublicError{
		Code: controlapiv1.PublicErrorCodeAuthenticationFailed, Action: controlapiv1.Reauthenticate,
		TraceId: base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x19}, 16)),
	}
	right := left
	right.TraceId = base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x29}, 16))
	leftBody, leftErr := json.Marshal(left)
	rightBody, rightErr := json.Marshal(right)
	if leftErr != nil || rightErr != nil || compareExternalAuthenticationErrors(leftBody, rightBody) != nil {
		t.Fatal("external enumeration rejected semantically identical public errors")
	}
	retryAfter := int64(1)
	right.RetryAfterMs = &retryAfter
	differentBody, marshalErr := json.Marshal(right)
	if marshalErr != nil || compareExternalAuthenticationErrors(leftBody, differentBody) == nil {
		t.Fatal("external enumeration ignored a non-trace public error difference")
	}
	right = left
	right.TraceId = "short"
	invalidTraceBody, marshalErr := json.Marshal(right)
	if marshalErr != nil || compareExternalAuthenticationErrors(leftBody, invalidTraceBody) == nil {
		t.Fatal("external enumeration accepted an invalid trace identifier")
	}
}

func TestLoadExternalDependencyConfigurationRequiresOwnedProjectAndLoopbackEndpoints(t *testing.T) {
	t.Parallel()
	valid := map[string]string{ //nolint:gosec // Synthetic loopback-only test dependency credentials.
		"C11_E2E_COMPOSE_PROJECT": "talenro-c11-verify-012345abcdef",
		"C11_E2E_DATABASE_URL":    "postgres://talenro:talenro_dev@127.0.0.1:15432/talenro?sslmode=disable",
		"C11_E2E_REDIS_ADDRESS":   "127.0.0.1:16379",
		"C11_E2E_NATS_URL":        "nats://127.0.0.1:14222",
	}
	configuration, enabled, err := loadExternalDependencyConfiguration(mapLookup(valid))
	if err != nil || !enabled || configuration.Project != valid["C11_E2E_COMPOSE_PROJECT"] ||
		configuration.DatabaseURL != valid["C11_E2E_DATABASE_URL"] || configuration.RedisAddress != valid["C11_E2E_REDIS_ADDRESS"] ||
		configuration.NATSURL != valid["C11_E2E_NATS_URL"] {
		t.Fatal("valid external dependency contract was rejected")
	}

	for name, mutate := range map[string]func(map[string]string){
		"foreign project": func(values map[string]string) { values["C11_E2E_COMPOSE_PROJECT"] = "talenro-dev" },
		"remote postgres": func(values map[string]string) {
			values["C11_E2E_DATABASE_URL"] = "postgres://talenro:talenro_dev@192.0.2.19:5432/talenro?sslmode=disable"
		},
		"remote redis": func(values map[string]string) { values["C11_E2E_REDIS_ADDRESS"] = "192.0.2.19:6379" },
		"remote nats":  func(values map[string]string) { values["C11_E2E_NATS_URL"] = "nats://192.0.2.19:4222" },
	} {
		name, mutate := name, mutate
		t.Run(name, func(t *testing.T) {
			values := make(map[string]string, len(valid))
			for key, original := range valid {
				values[key] = original
			}
			mutate(values)
			if _, enabled, err := loadExternalDependencyConfiguration(mapLookup(values)); err == nil || !enabled {
				t.Fatal("invalid external dependency contract was accepted")
			}
		})
	}
	for missing := range valid {
		values := make(map[string]string, len(valid)-1)
		for key, original := range valid {
			if key != missing {
				values[key] = original
			}
		}
		if _, enabled, err := loadExternalDependencyConfiguration(mapLookup(values)); err == nil || !enabled {
			t.Fatalf("partial external dependency contract without %s was accepted", missing)
		}
	}
	if _, enabled, err := loadExternalDependencyConfiguration(mapLookup(nil)); err != nil || enabled {
		t.Fatal("absent external dependency contract did not remain disabled")
	}
	if _, enabled, err := loadExternalDependencyConfiguration(mapLookup(map[string]string{fixtureComposeProject: ""})); err == nil || !enabled {
		t.Fatal("present empty external dependency variable was accepted as an absent contract")
	}
}

func TestResolutionEnvelopeDigestRejectsClaimTamper(t *testing.T) {
	t.Parallel()
	body := []byte(`{"envelope_version":"1"}`)
	digest := sha256.Sum256(body)
	claim := fmt.Sprintf("%x", digest)
	if !validResolutionEnvelopeDigest(claim, body) {
		t.Fatal("valid envelope digest claim was rejected")
	}
	tampered := []byte(claim)
	if tampered[0] == '0' {
		tampered[0] = '1'
	} else {
		tampered[0] = '0'
	}
	if validResolutionEnvelopeDigest(string(tampered), body) {
		t.Fatal("tampered envelope digest claim was accepted")
	}
}

func TestFixtureChildEnvironmentUsesOnlyExplicitDependencyContract(t *testing.T) {
	t.Parallel()
	values := []string{
		"PATH=C:\\fixture-tools", "C11_PRIVATE_CANARY=must-not-survive", "TALENRO_DATABASE_URL=must-not-survive",
		"C11_E2E_EXTERNAL_DEPENDENCIES=must-not-survive", "C11_E2E_COMPOSE_PROJECT=talenro-c11-verify-012345abcdef",
		"C11_E2E_DATABASE_URL=postgres://talenro:talenro_dev@127.0.0.1:15432/talenro?sslmode=disable",
		"C11_E2E_REDIS_ADDRESS=127.0.0.1:16379", "C11_E2E_NATS_URL=nats://127.0.0.1:14222",
	}
	filtered := fixtureChildEnvironmentValues(values)
	joined := strings.Join(filtered, "\n")
	for _, required := range []string{
		"PATH=C:\\fixture-tools", "C11_E2E_FIXTURE_CHILD=1",
		"C11_E2E_COMPOSE_PROJECT=talenro-c11-verify-012345abcdef", "C11_E2E_REDIS_ADDRESS=127.0.0.1:16379",
	} {
		if !strings.Contains(joined, required) {
			t.Fatalf("fixture child environment omitted %q", required)
		}
	}
	if strings.Contains(joined, "PRIVATE_CANARY") || strings.Contains(joined, "TALENRO_DATABASE_URL") ||
		strings.Contains(joined, "EXTERNAL_DEPENDENCIES") {
		t.Fatal("fixture child inherited an ambient private environment value")
	}
}

func TestNewExternalC11FixtureUsesProductionListenersWithoutChild(t *testing.T) {
	t.Parallel()
	configuration := externalRuntimeConfiguration{
		PrimaryURL: "http://127.0.0.1:8080",
		MirrorAURL: "http://127.0.0.1:8081",
		MirrorBURL: "http://127.0.0.1:8082",
		Origins:    [3]string{"http://localhost:8080", "http://localhost:8081", "http://localhost:8082"},
	}
	fixture, cleanup, err := newExternalC11Fixture(configuration, mapLookup(nil))
	if err != nil {
		t.Fatalf("external C1.1 fixture construction failed: %v", err)
	}
	defer cleanup()
	if fixture.child != nil || fixture.compose != nil {
		t.Fatal("external C1.1 fixture launched a test child or compose replacement")
	}
	if fixture.ready.RequiredURL != configuration.PrimaryURL || fixture.ready.MirrorAURL != configuration.MirrorAURL ||
		fixture.ready.MirrorBURL != configuration.MirrorBURL {
		t.Fatal("external C1.1 fixture did not retain the production listener contract")
	}
	metadataBytes, metadataErr := base64.RawURLEncoding.DecodeString(fixture.ready.Metadata)
	rootPublic, rootErr := base64.RawURLEncoding.DecodeString(fixture.ready.RootPublic)
	if metadataErr != nil || rootErr != nil {
		t.Fatal("external C1.1 trust anchor encoding failed")
	}
	if _, err := trustclient.NewTrustMetadata(metadataBytes, map[string]ed25519.PublicKey{fixture.ready.RootKeyID: rootPublic}, 0); err != nil {
		t.Fatal("external C1.1 trust anchor was not independently verifiable")
	}
}

func TestExternalDeviceProofUsesCanonicalPublicOrigin(t *testing.T) {
	t.Parallel()
	configuration := externalRuntimeConfiguration{
		PrimaryURL: "http://127.0.0.1:8080",
		MirrorAURL: "http://127.0.0.1:8081",
		MirrorBURL: "http://127.0.0.1:8082",
		Origins:    [3]string{"http://localhost:8080", "http://localhost:8081", "http://localhost:8082"},
	}
	fixture, cleanup, err := newExternalC11Fixture(configuration, mapLookup(nil))
	if err != nil {
		t.Fatal("external C1.1 fixture construction failed")
	}
	defer cleanup()
	audience, err := fixture.deviceProofAudienceFor(configuration.PrimaryURL)
	if err != nil || audience != "http://localhost:8080" || audience == configuration.PrimaryURL {
		t.Fatal("external device proof did not use the canonical public origin")
	}
}

func TestDeviceProofAudienceUsesConfiguredPublicOriginForOwnedProfiles(t *testing.T) {
	t.Parallel()
	const publicAudience = "http://localhost:18080"
	fixture := &c11Fixture{
		ready: fixtureReady{
			RequiredURL: "http://127.0.0.1:18080",
			GraceURL:    "http://127.0.0.1:18081",
			DisabledURL: "http://127.0.0.1:18082",
			MirrorAURL:  "http://127.0.0.1:18083",
			MirrorBURL:  "http://127.0.0.1:18084",
			MetricsURL:  "http://127.0.0.1:19090",
		},
		deviceProofAudience: publicAudience,
	}
	for _, profile := range []struct {
		name     string
		endpoint string
	}{
		{name: "required", endpoint: fixture.ready.RequiredURL},
		{name: "grace", endpoint: fixture.ready.GraceURL},
		{name: "disabled", endpoint: fixture.ready.DisabledURL},
	} {
		t.Run(profile.name, func(t *testing.T) {
			audience, err := fixture.deviceProofAudienceFor(profile.endpoint)
			if err != nil || audience != publicAudience {
				t.Fatalf("owned profile proof audience = %q, %v; want %q", audience, err, publicAudience)
			}
		})
	}
	for _, unowned := range []struct {
		name     string
		endpoint string
	}{
		{name: "mirror_a", endpoint: fixture.ready.MirrorAURL},
		{name: "mirror_b", endpoint: fixture.ready.MirrorBURL},
		{name: "metrics", endpoint: fixture.ready.MetricsURL},
		{name: "arbitrary_loopback", endpoint: "http://127.0.0.1:18085"},
	} {
		t.Run("rejects_"+unowned.name, func(t *testing.T) {
			if audience, err := fixture.deviceProofAudienceFor(unowned.endpoint); err == nil || audience != "" {
				t.Fatalf("unowned endpoint proof audience = %q, %v; want rejection", audience, err)
			}
		})
	}
	for _, malformed := range []struct {
		name string
		raw  string
	}{
		{name: "bare_query_marker", raw: "http://127.0.0.1:18081?"},
		{name: "localhost_transport_alias", raw: "http://localhost:18081"},
		{name: "path", raw: "http://127.0.0.1:18081/path"},
		{name: "non_empty_query", raw: "http://127.0.0.1:18081?mode=grace"},
		{name: "userinfo", raw: "http://user@127.0.0.1:18081"},
		{name: "fragment", raw: "http://127.0.0.1:18081#fragment"},
	} {
		t.Run("rejects_owned_"+malformed.name, func(t *testing.T) {
			malformedFixture := *fixture
			malformedFixture.ready.GraceURL = malformed.raw
			if audience, err := malformedFixture.deviceProofAudienceFor(malformed.raw); err == nil || audience != "" {
				t.Fatalf("malformed owned endpoint proof audience = %q, %v; want rejection", audience, err)
			}
		})
	}
	gracePortAudience := *fixture
	gracePortAudience.deviceProofAudience = "http://localhost:18081"
	if audience, err := gracePortAudience.deviceProofAudienceFor(gracePortAudience.ready.GraceURL); err == nil || audience != "" {
		t.Fatalf("grace-port proof audience = %q, %v; want rejection", audience, err)
	}
}

func TestBundleLocationsMatchOnlyExpectedOwnedListenersAndLocator(t *testing.T) {
	t.Parallel()
	expected := [3]string{
		"http://localhost:8080",
		"http://localhost:8081",
		"http://localhost:8082",
	}
	locator := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x42}, 32))
	escapedPath := "/b/" + url.PathEscape(locator)
	valid := []any{
		expected[0] + escapedPath,
		expected[1] + escapedPath,
		expected[2] + escapedPath,
	}
	if !validBundleSourceLocations(valid, expected, locator, false) {
		t.Fatal("bundle locations did not match the exact owned origins and response locator")
	}
	alias := []any{
		"http://127.0.0.1:8080" + escapedPath,
		"http://127.0.0.1:8081" + escapedPath,
		"http://127.0.0.1:8082" + escapedPath,
	}
	if validBundleSourceLocations(alias, expected, locator, false) || validBundleSourceLocations(alias, expected, locator, true) {
		t.Fatal("bundle listener aliases matched configured published origins")
	}
	otherLocator := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x43}, 32))
	wrongLocator := []any{
		expected[0] + "/b/" + otherLocator,
		expected[1] + "/b/" + otherLocator,
		expected[2] + "/b/" + otherLocator,
	}
	if validBundleSourceLocations(wrongLocator, expected, locator, false) {
		t.Fatal("three matching locations with the wrong response locator were accepted")
	}
	tests := []struct {
		name  string
		index int
		value any
	}{
		{name: "legacy path", index: 0, value: expected[0] + "/v1/immutable/config-bundles/" + locator},
		{name: "mismatched locator", index: 1, value: expected[1] + "/b/" + otherLocator},
		{name: "extra path", index: 2, value: expected[2] + escapedPath + "/extra"},
		{name: "query", index: 0, value: expected[0] + escapedPath + "?token=private"},
		{name: "userinfo", index: 1, value: "http://user@127.0.0.1:8081" + escapedPath},
		{name: "fragment", index: 1, value: expected[1] + escapedPath + "#private"},
		{name: "remote origin", index: 2, value: "http://192.0.2.19:8082" + escapedPath},
		{name: "wrong mirror port", index: 2, value: expected[1] + escapedPath},
		{name: "swapped origin", index: 0, value: expected[2] + escapedPath},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			invalid := append([]any(nil), valid...)
			invalid[test.index] = test.value
			if validBundleSourceLocations(invalid, expected, locator, false) {
				t.Fatal("invalid bundle location matched the response contract")
			}
		})
	}
}

func TestRecoveryCodeRotationUsesPasswordReauthentication(t *testing.T) {
	t.Parallel()
	const password = "Task19-Recovery-Password!76" //nolint:gosec // Synthetic E2E request-contract credential.
	const totpSecret = "JBSWY3DPEHPK3PXP"          //nolint:gosec // Synthetic RFC 6238 request-contract secret.
	principalID := uuid.MustParse("0f548c4e-e0cd-41c8-97a0-2f397135d89b")
	clock := &fixtureClock{now: time.Date(2035, time.January, 2, 3, 4, 5, 0, time.UTC), maximum: time.Date(2035, time.January, 4, 3, 4, 5, 0, time.UTC)}
	advanced := time.Date(2035, time.January, 2, 3, 4, 36, 0, time.UTC)
	if !clock.set(advanced.Format(time.RFC3339)) {
		t.Fatal("fixture clock controlled advance failed")
	}
	wantTOTP := c11TOTPCode(t, totpSecret, advanced)
	var totpVerified atomic.Bool
	var recoveryRotated atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(io.LimitReader(request.Body, 1<<20)).Decode(&body); err != nil {
			t.Error("strong-auth request capture failed")
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		writeJSON := func(status int, value any) {
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(status)
			if value != nil {
				_ = json.NewEncoder(writer).Encode(value)
			}
		}
		reauthentication, _ := body["reauthentication"].(map[string]any)
		switch request.URL.Path {
		case "/v1/totp-enrollments":
			writeJSON(http.StatusCreated, map[string]any{"secret": totpSecret})
		case "/v1/totp-verifications":
			if fmt.Sprint(body["code"]) != wantTOTP || reauthentication["method"] != "password" || reauthentication["password"] != password {
				t.Error("TOTP verification request contract changed")
				writer.WriteHeader(http.StatusBadRequest)
				return
			}
			totpVerified.Store(true)
			writer.WriteHeader(http.StatusNoContent)
		case "/v1/recovery-code-rotations":
			if reauthentication["method"] != "password" || reauthentication["password"] != password {
				t.Error("recovery-code rotation did not use the account password reauthentication contract")
				writer.WriteHeader(http.StatusBadRequest)
				return
			}
			recoveryRotated.Store(true)
			codes := make([]string, 10)
			for index := range codes {
				codes[index] = fmt.Sprintf("recovery-%02d", index)
			}
			writeJSON(http.StatusCreated, map[string]any{"codes": codes})
		case "/v1/totp-revocations":
			if reauthentication["method"] != "password" || reauthentication["password"] != password {
				t.Error("TOTP revocation request contract changed")
				writer.WriteHeader(http.StatusBadRequest)
				return
			}
			writer.WriteHeader(http.StatusNoContent)
		case "/v1/recovery-code-consumptions":
			writeJSON(http.StatusOK, map[string]any{
				"principal_id":  principalID.String(),
				"session_id":    "599865a4-a79c-4c87-9016-b574a95ad786",
				"access_token":  "fixture-access-token",
				"refresh_token": "fixture-refresh-token",
			})
		default:
			t.Error("unexpected strong-auth request path")
			writer.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	account := c11Account{
		baseURL: server.URL, email: "recovery-contract@example.test", password: password,
		accessToken: "fixture-access-token", refreshToken: "fixture-refresh-token",
		principalID: principalID, sessionID: uuid.MustParse("157e7584-a3bd-488a-a37e-3618799f5020"),
	}
	fixture := &c11Fixture{authorityNow: func() (time.Time, error) { return clock.Now(), nil }}
	client := &c11HTTPClient{t: t, client: server.Client(), context: t.Context()}
	fixture.exerciseTOTPAndRecovery(t, client, &account)
	if !totpVerified.Load() || !recoveryRotated.Load() {
		t.Fatal("strong-auth request contract did not exercise TOTP verification and password recovery rotation")
	}
}

type fixtureEmailSubscriptionFunc func(int, ...nats.PullOpt) ([]*nats.Msg, error)

func (fetch fixtureEmailSubscriptionFunc) Fetch(batch int, options ...nats.PullOpt) ([]*nats.Msg, error) {
	return fetch(batch, options...)
}

type fixtureEmailConsumerFunc func(context.Context, []byte, time.Time) error

func (consume fixtureEmailConsumerFunc) Consume(ctx context.Context, body []byte, now time.Time) error {
	return consume(ctx, body, now)
}

func TestFixtureEmailLoopUsesSharedAdvancedClock(t *testing.T) {
	t.Parallel()
	clock := &fixtureClock{now: time.Date(2036, time.February, 3, 4, 5, 6, 0, time.UTC), maximum: time.Date(2036, time.February, 5, 4, 5, 6, 0, time.UTC)}
	want := time.Date(2036, time.February, 3, 4, 5, 37, 0, time.UTC)
	if !clock.set(want.Format(time.RFC3339)) {
		t.Fatal("fixture email clock controlled advance failed")
	}
	ctx, cancel := context.WithCancel(t.Context())
	var got time.Time
	var deliveries atomic.Int32
	var ackFailures atomic.Int32
	fetches := 0
	subscription := fixtureEmailSubscriptionFunc(func(batch int, _ ...nats.PullOpt) ([]*nats.Msg, error) {
		fetches++
		if batch != 1 || fetches != 1 {
			t.Fatal("fixture email loop fetch contract changed")
		}
		return []*nats.Msg{{Data: []byte(`{"fixture":"email"}`)}}, nil
	})
	consumer := fixtureEmailConsumerFunc(func(_ context.Context, _ []byte, now time.Time) error {
		got = now
		cancel()
		return nil
	})

	fixtureEmailLoop(ctx, subscription, consumer, clock, &deliveries, &ackFailures)
	if got != want || deliveries.Load() != 1 || ackFailures.Load() != 0 {
		t.Fatalf("fixture email consumer time = %s, deliveries = %d, ack failures = %d; want %s, 1, 0", got.Format(time.RFC3339Nano), deliveries.Load(), ackFailures.Load(), want.Format(time.RFC3339Nano))
	}
}

func TestFixtureTrustVerificationUsesSharedAdvancedClock(t *testing.T) {
	t.Parallel()
	advanced := time.Date(2045, time.January, 2, 3, 4, 5, 0, time.UTC)
	validFrom := advanced.Add(-time.Hour)
	validUntil := advanced.Add(time.Hour)
	rootSigner, err := trust.NewLocalRootSigner(secret.NewBytes(bytes.Repeat([]byte{0x31}, ed25519.SeedSize)))
	if err != nil {
		t.Fatal("new trust root signer failed")
	}
	defer func() { _ = rootSigner.Close() }()
	configSigner, err := trust.NewLocalConfigSigner(secret.NewBytes(bytes.Repeat([]byte{0x21}, ed25519.SeedSize)))
	if err != nil {
		t.Fatal("new trust config signer failed")
	}
	defer func() { _ = configSigner.Close() }()
	boundedRoot, err := trust.NewTimeoutConfigSigner(rootSigner, 2*time.Second)
	if err != nil {
		t.Fatal("bound trust root signer failed")
	}
	defer func() { _ = boundedRoot.Close() }()
	record, err := trust.SignRootMetadataV1(t.Context(), trust.RootMetadataV1{
		SchemaVersion: trust.TrustMetadataSchemaV1, Version: "1", RootKeyID: rootSigner.KeyID(), RootAlgorithm: trust.SignatureAlgorithm,
		ValidFrom: validFrom.Format(time.RFC3339), ValidUntil: validUntil.Format(time.RFC3339),
		SigningKeys: []trust.SigningKeyMetadataV1{{
			KeyID: configSigner.KeyID(), Algorithm: trust.SignatureAlgorithm,
			PublicKey: base64.RawURLEncoding.EncodeToString(configSigner.PublicKey()), State: "active",
			NotBefore: validFrom.Format(time.RFC3339), NotAfter: validUntil.Format(time.RFC3339),
		}},
	}, boundedRoot)
	if err != nil {
		t.Fatal("sign advanced trust metadata failed")
	}
	repository := &memoryMetadataRepository{records: []trust.SignedRootMetadataV1{record}}
	metadata, _, err := fixtureMetadata(t.Context(), repository, rootSigner, configSigner, advanced)
	if err != nil {
		t.Fatalf("fixture metadata rejected shared advanced authority: %v", err)
	}
	if metadata.ValidFrom != validFrom.Format(time.RFC3339) || metadata.ValidUntil != validUntil.Format(time.RFC3339) {
		t.Fatal("fixture metadata did not retain the advanced authority validity window")
	}
}

type fixtureRoundTripFunc func(*http.Request) (*http.Response, error)

func (roundTrip fixtureRoundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return roundTrip(request)
}

type fixtureErrorReader struct{}

func (fixtureErrorReader) Read([]byte) (int, error) { return 0, errors.New("fixture read failure") }

func TestC11POSTRequestBodySurvivesAsyncTransportAndClearsOnClose(t *testing.T) {
	t.Parallel()
	payload := []byte(`{"password":"C11-BODY-PRIVATE-19"}`) //nolint:gosec // Synthetic lifecycle canary.
	want := append([]byte(nil), payload...)
	request, err := newC11POSTRequest(t.Context(), "http://127.0.0.1:8080/v1/accounts", payload)
	if err != nil {
		t.Fatal("clearing request construction failed")
	}
	body, ok := request.Body.(*clearingRequestBody)
	if !ok || fmt.Sprint(body) != "[REDACTED]" || fmt.Sprintf("%#v", body) != "[REDACTED]" {
		t.Fatal("clearing request body formatting was not fixed and redacted")
	}
	release := make(chan struct{})
	consumed := make(chan []byte, 1)
	client := &http.Client{Transport: fixtureRoundTripFunc(func(request *http.Request) (*http.Response, error) {
		go func() {
			<-release
			captured, _ := io.ReadAll(request.Body)
			_ = request.Body.Close()
			consumed <- captured
		}()
		return &http.Response{
			StatusCode: http.StatusNoContent,
			Header:     make(http.Header),
			Body:       io.NopCloser(bytes.NewReader(nil)),
			Request:    request,
		}, nil
	})}
	response, err := client.Do(request)
	if err != nil {
		t.Fatal("delayed transport fixture failed")
	}
	_ = response.Body.Close()
	close(release)
	select {
	case captured := <-consumed:
		if !bytes.Equal(captured, want) {
			t.Fatal("request body was cleared before the asynchronous transport consumed it")
		}
	case <-time.After(time.Second):
		t.Fatal("asynchronous request body consumption deadline exceeded")
	}
	for _, value := range payload {
		if value != 0 {
			t.Fatal("request body backing bytes were not cleared on close")
		}
	}
}

func TestC11POSTRequestBodyClearsOnConstructionAndTransportErrors(t *testing.T) {
	t.Parallel()
	constructionPayload := []byte(`{"password":"C11-CONSTRUCTION-PRIVATE-19"}`) //nolint:gosec // Synthetic lifecycle canary.
	if _, err := newC11POSTRequest(t.Context(), "http://[::1", constructionPayload); err == nil {
		t.Fatal("invalid request URL was accepted")
	}
	for _, value := range constructionPayload {
		if value != 0 {
			t.Fatal("request body was not cleared after construction failure")
		}
	}
	transportPayload := []byte(`{"password":"C11-TRANSPORT-PRIVATE-19"}`) //nolint:gosec // Synthetic lifecycle canary.
	request, err := newC11POSTRequest(t.Context(), "http://127.0.0.1:8080/v1/accounts", transportPayload)
	if err != nil {
		t.Fatal("transport-error request construction failed")
	}
	client := &http.Client{Transport: fixtureRoundTripFunc(func(request *http.Request) (*http.Response, error) {
		_ = request.Body.Close()
		return nil, errors.New("fixture transport failure")
	})}
	response, err := client.Do(request)
	if response != nil {
		_ = response.Body.Close()
	}
	if err == nil {
		t.Fatal("transport error fixture unexpectedly succeeded")
	}
	for _, value := range transportPayload {
		if value != 0 {
			t.Fatal("request body was not cleared after transport failure")
		}
	}
}

func TestDoC11JSONPostFourByHundredDelayedReadsClearOnlyOnTransportClose(t *testing.T) {
	t.Parallel()
	const (
		workers  = 4
		attempts = 100
	)
	payload := map[string]any{"password": "C11-4X100-PRIVATE-19"} //nolint:gosec // Synthetic lifecycle canary.
	want, err := json.Marshal(payload)
	if err != nil {
		t.Fatal("4x100 request body fixture encoding failed")
	}
	release := make(chan struct{})
	type capturedRequest struct {
		body     *clearingRequestBody
		value    []byte
		bodyType bool
	}
	captured := make(chan capturedRequest, workers*attempts)
	client := &http.Client{Transport: fixtureRoundTripFunc(func(request *http.Request) (*http.Response, error) {
		body, ok := request.Body.(*clearingRequestBody)
		go func() {
			<-release
			value, _ := io.ReadAll(request.Body)
			_ = request.Body.Close()
			captured <- capturedRequest{body: body, value: value, bodyType: ok}
		}()
		return &http.Response{
			StatusCode: http.StatusNoContent,
			Header:     make(http.Header),
			Body:       io.NopCloser(bytes.NewReader(nil)),
			Request:    request,
		}, nil
	})}
	results := make(chan c11HTTPResult, workers*attempts)
	var callers sync.WaitGroup
	callers.Add(workers)
	for range workers {
		go func() {
			defer callers.Done()
			for range attempts {
				results <- doC11JSONPost(t.Context(), client, "http://127.0.0.1:8080", "/v1/accounts", "", "", payload)
			}
		}()
	}
	callers.Wait()
	close(release)
	for range workers * attempts {
		result := <-results
		if result.err != nil || result.status != http.StatusNoContent {
			t.Fatal("4x100 request transport failed")
		}
		observed := <-captured
		if !observed.bodyType || observed.body == nil || !bytes.Equal(observed.value, want) {
			t.Fatal("doC11JSONPost did not retain request bytes until transport close")
		}
		observed.body.mu.Lock()
		cleared := observed.body.closed && observed.body.value == nil
		observed.body.mu.Unlock()
		if !cleared {
			t.Fatal("doC11JSONPost request bytes survived transport close")
		}
	}
}

func TestDoC11JSONPostClearsRequestBodyOnEveryTransportTerminalPath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		response  *http.Response
		transport error
		wantError bool
	}{
		{name: "success", response: &http.Response{StatusCode: http.StatusNoContent, Header: make(http.Header), Body: io.NopCloser(bytes.NewReader(nil))}},
		{name: "response read error", response: &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(fixtureErrorReader{})}, wantError: true},
		{name: "transport error", transport: errors.New("fixture transport failure"), wantError: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var observed *clearingRequestBody
			client := &http.Client{Transport: fixtureRoundTripFunc(func(request *http.Request) (*http.Response, error) {
				observed, _ = request.Body.(*clearingRequestBody)
				_ = request.Body.Close()
				if test.response != nil {
					test.response.Request = request
				}
				return test.response, test.transport
			})}
			result := doC11JSONPost(t.Context(), client, "http://127.0.0.1:8080", "/v1/accounts", "", "", map[string]any{
				"password": "C11-TERMINAL-PRIVATE-19", //nolint:gosec // Synthetic lifecycle canary.
			})
			if (result.err != nil) != test.wantError {
				t.Fatal("doC11JSONPost terminal result mismatch")
			}
			if observed == nil {
				t.Fatal("doC11JSONPost did not use the clearing request body")
			}
			observed.mu.Lock()
			cleared := observed.closed && observed.value == nil
			observed.mu.Unlock()
			if !cleared {
				t.Fatal("doC11JSONPost retained request bytes after a terminal transport path")
			}
		})
	}
}

func TestRestartableSourceStopsAndReopensTheSameListener(t *testing.T) {
	t.Parallel()
	listenConfig := net.ListenConfig{}
	listener, err := listenConfig.Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal("restartable source listener fixture failed")
	}
	source := startRestartableSource(listener, http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = source.stop(ctx)
	})
	client := &http.Client{Timeout: time.Second}
	url := "http://" + listener.Addr().String()
	requestStatus := func() (int, error) {
		request, requestErr := http.NewRequestWithContext(t.Context(), http.MethodGet, url, nil)
		if requestErr != nil {
			return 0, requestErr
		}
		response, requestErr := client.Do(request)
		if requestErr != nil {
			return 0, requestErr
		}
		defer func() { _ = response.Body.Close() }()
		return response.StatusCode, nil
	}
	if status, err := requestStatus(); err != nil || status != http.StatusNoContent {
		t.Fatal("restartable source did not serve its initial listener")
	}
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	if err := source.stop(ctx); err != nil {
		cancel()
		t.Fatal("restartable source stop failed")
	}
	cancel()
	if _, err := requestStatus(); err == nil {
		t.Fatal("restartable source listener remained reachable after stop")
	}
	ctx, cancel = context.WithTimeout(t.Context(), time.Second)
	if err := source.start(ctx); err != nil {
		cancel()
		t.Fatal("restartable source restart failed")
	}
	cancel()
	if status, err := requestStatus(); err != nil || status != http.StatusNoContent {
		t.Fatal("restartable source did not reopen its original listener")
	}
}

func mapLookup(values map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}
