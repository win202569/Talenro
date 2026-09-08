//go:build integration

package authority

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/store"
)

func TestPITRBeforeRevocationFailsClosed(t *testing.T) {
	harness := newPITRDockerHarness(t)
	t.Cleanup(harness.cleanup)

	primaryVolume := harness.createVolume("primary")
	backupVolume := harness.createVolume("backup")
	primaryPort := reservePITRLoopbackPort(t)
	primaryName := harness.startPostgres("primary", primaryPort, primaryVolume, "", backupVolume)
	primaryPool := openPITRPool(t, primaryPort)
	applyPITRAuthoritySchema(t, primaryPool)
	createPITREffectTable(t, primaryPool)

	provider, err := NewDeterministicProvider(29)
	if err != nil {
		t.Fatal(err)
	}
	effects := newCoordinatorEffectResolver()
	clock := coordinatorClock{now: time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC)}
	primaryRepository, err := NewPostgresRepository(primaryPool)
	if err != nil {
		t.Fatal(err)
	}
	tracingRepository := &pitrTracingRepository{PostgresRepository: primaryRepository}
	primaryCoordinator := mustNewCoordinatorForTest(t, provider, tracingRepository, effects, clock)

	for sequence := 1; sequence <= 11; sequence++ {
		commitPITRAuthorityEffect(t, primaryPool, primaryCoordinator, tracingRepository, effects, sequence, tracingRepository)
	}
	providerAtBackup, err := provider.Head(t.Context())
	if err != nil || providerAtBackup.LatestReservedSequence != 11 || providerAtBackup.LatestCommittedSequence != 11 {
		t.Fatalf("provider head at backup = %#v, %v; want committed sequence 11", providerAtBackup, err)
	}
	identityAtBackup, err := primaryRepository.CaptureDatabasePoint(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	harness.physicalBaseBackup(primaryName)

	revocationReceipt := commitPITRAuthorityEffect(t, primaryPool, primaryCoordinator, tracingRepository, effects, 12, tracingRepository)
	providerAfterRevocation, err := provider.Head(t.Context())
	if err != nil || providerAfterRevocation.LatestReservedSequence != 12 || providerAfterRevocation.LatestCommittedSequence != 12 ||
		providerAfterRevocation.LatestCommittedReceiptDigest != revocationReceipt.ReceiptDigest {
		t.Fatalf("provider head after revoke = %#v, %v; want unchanged external sequence 12", providerAfterRevocation, err)
	}

	primaryPool.Close()
	harness.stopContainer(primaryName)
	restorePort := reservePITRLoopbackPort(t)
	harness.startPostgres("restore", restorePort, backupVolume, "base", "")
	restoredPool := openPITRPool(t, restorePort)
	restoredRepository, err := NewPostgresRepository(restoredPool)
	if err != nil {
		t.Fatal(err)
	}
	restoredCoordinator := mustNewCoordinatorForTest(t, provider, restoredRepository, effects, clock)
	restoredIdentity, err := restoredRepository.CaptureDatabasePoint(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if restoredIdentity.SystemID != identityAtBackup.SystemID || restoredIdentity.Timeline != identityAtBackup.Timeline {
		t.Fatalf("historical restore identity = %#v, want preserved system/timeline from %#v", restoredIdentity, identityAtBackup)
	}
	var staleActive int
	if err := restoredPool.QueryRow(t.Context(), `SELECT count(*) FROM nodecontrol.control_plane_authority_fences WHERE provider_status='committed' AND visibility_state='active'`).Scan(&staleActive); err != nil {
		t.Fatal(err)
	}
	if staleActive != 11 {
		t.Fatalf("historical restore active fixture rows = %d, want 11", staleActive)
	}

	readiness, err := restoredCoordinator.CheckReady(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if readiness.Ready || string(readiness.Reason) != "database_behind_provider" {
		t.Fatalf("historical restore readiness = %#v, want database_behind_provider", readiness)
	}
	if err := queryPITRActiveAuthorityFixture(t.Context(), restoredCoordinator, restoredPool); err != ErrAuthorityUnavailable {
		t.Fatalf("guarded active fixture query error = %v, want ErrAuthorityUnavailable", err)
	}

	freshVolume := harness.createVolume("different")
	freshPort := reservePITRLoopbackPort(t)
	harness.startPostgres("different", freshPort, freshVolume, "", "")
	freshPool := openPITRPool(t, freshPort)
	applyPITRAuthoritySchema(t, freshPool)
	freshRepository, err := NewPostgresRepository(freshPool)
	if err != nil {
		t.Fatal(err)
	}
	freshCoordinator := mustNewCoordinatorForTest(t, provider, freshRepository, effects, clock)
	freshReadiness, err := freshCoordinator.CheckReady(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if freshReadiness.Ready || string(freshReadiness.Reason) != "database_identity_mismatch" {
		t.Fatalf("different-cluster readiness = %#v, want database_identity_mismatch precedence", freshReadiness)
	}

	providerStillExternal, err := provider.Head(t.Context())
	if err != nil || providerStillExternal.LatestCommittedSequence != 12 ||
		providerStillExternal.LatestCommittedReceiptDigest != revocationReceipt.ReceiptDigest {
		t.Fatalf("provider changed during PITR = %#v, %v", providerStillExternal, err)
	}
}

func commitPITRAuthorityEffect(
	t *testing.T,
	pool *pgxpool.Pool,
	coordinator *Coordinator,
	repository Repository,
	effects *coordinatorEffectResolver,
	sequence int,
	trace *pitrTracingRepository,
) Receipt {
	t.Helper()
	operationID := uuid.NewSHA1(uuid.MustParse("87bd4647-c131-4d50-96d4-f690ab276760"), []byte(fmt.Sprintf("authority-%02d", sequence)))
	request := ReserveRequest{
		OperationID: operationID,
		Kind:        EffectCertificateRevoke,
		ScopeKind:   ScopeNode,
		ScopeDigest: sha256.Sum256([]byte(fmt.Sprintf("pitr-node-%02d", sequence))),
	}
	reservation, err := coordinator.Reserve(t.Context(), request)
	if err != nil {
		t.Fatalf("sequence %d reserve: %v", sequence, err)
	}
	effectDigest := sha256.Sum256([]byte(fmt.Sprintf("pitr-effect-%02d", sequence)))
	if err := inAuthorityTransaction(t.Context(), pool, func(transaction pgx.Tx) error {
		if err := repository.RecordPending(t.Context(), transaction, reservation, coordinator.clock.Now().Add(-time.Second)); err != nil {
			return err
		}
		_, err := transaction.Exec(t.Context(), `INSERT INTO nodecontrol.authority_task7_effects(operation_id,effect_digest) VALUES ($1,$2)`, operationID, effectDigest[:])
		return err
	}); err != nil {
		t.Fatalf("sequence %d domain transaction: %v", sequence, err)
	}
	effectDigest = effects.commit(t, request, effectDigest, WALPosition(fmt.Sprintf("0/%X", 0x100+sequence)))
	trace.resetTrace()
	receipt, err := coordinator.Finalize(t.Context(), CoordinatorFinalizeRequest{OperationID: operationID, EffectDigest: effectDigest})
	if err != nil {
		var reservedAt time.Time
		if queryErr := pool.QueryRow(t.Context(), `SELECT reserved_at FROM nodecontrol.control_plane_authority_fences WHERE operation_id=$1`, operationID).Scan(&reservedAt); queryErr != nil {
			t.Fatalf("sequence %d finalize: %v; trace=%v; reserved-at probe: %v", sequence, err, trace.snapshotTrace(), queryErr)
		}
		coordinatorNow := coordinator.clock.Now()
		t.Fatalf("sequence %d finalize: %v; trace=%v; reserved_at=%s coordinator_now=%s bound_before_reserved=%t",
			sequence, err, trace.snapshotTrace(), reservedAt.UTC().Format(time.RFC3339Nano), coordinatorNow.Format(time.RFC3339Nano), coordinatorNow.Before(reservedAt))
	}
	return receipt
}

type pitrTracingRepository struct {
	*PostgresRepository
	mu    sync.Mutex
	trace []string
}

func (repository *pitrTracingRepository) Get(ctx context.Context, operationID uuid.UUID) (Record, error) {
	record, err := repository.PostgresRepository.Get(ctx, operationID)
	repository.addTrace(fmt.Sprintf("get:%s:bound=%t:terminal=%t", pitrTraceError(err), record.BoundEffectDigest != nil, record.TerminalReceipt != nil))
	return record, err
}

func (repository *pitrTracingRepository) CaptureDatabasePoint(ctx context.Context) (DatabasePoint, error) {
	point, err := repository.PostgresRepository.CaptureDatabasePoint(ctx)
	repository.addTrace(fmt.Sprintf("capture:%s:valid=%t", pitrTraceError(err), point.Validate() == nil))
	return point, err
}

func (repository *pitrTracingRepository) BindEffect(ctx context.Context, dbtx store.DBTX, operationID uuid.UUID, digest contracts.Digest, point DatabasePoint, at time.Time) error {
	err := repository.PostgresRepository.BindEffect(ctx, dbtx, operationID, digest, point, at)
	repository.addTrace("bind:" + pitrTraceError(err))
	return err
}

func (repository *pitrTracingRepository) ActivateCommitted(ctx context.Context, dbtx store.DBTX, receipt Receipt, at time.Time) error {
	err := repository.PostgresRepository.ActivateCommitted(ctx, dbtx, receipt, at)
	repository.addTrace("activate:" + pitrTraceError(err))
	return err
}

func (repository *pitrTracingRepository) withAuthorityTransaction(ctx context.Context, operation func(store.DBTX) error) error {
	repository.addTrace("transaction:begin")
	err := repository.PostgresRepository.withAuthorityTransaction(ctx, func(dbtx store.DBTX) error {
		repository.addTrace("transaction:callback")
		callbackErr := operation(dbtx)
		repository.addTrace("transaction:callback:" + pitrTraceError(callbackErr))
		return callbackErr
	})
	repository.addTrace("transaction:end:" + pitrTraceError(err))
	return err
}

func (repository *pitrTracingRepository) addTrace(value string) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	repository.trace = append(repository.trace, value)
}

func (repository *pitrTracingRepository) resetTrace() {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	repository.trace = nil
}

func (repository *pitrTracingRepository) snapshotTrace() []string {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	return append([]string(nil), repository.trace...)
}

func pitrTraceError(err error) string {
	switch err {
	case nil:
		return "ok"
	case ErrInvalidArgument:
		return "invalid_argument"
	case ErrConflict:
		return "conflict"
	case ErrTerminalConflict:
		return "terminal_conflict"
	case ErrNotFound:
		return "not_found"
	case ErrCanceled:
		return "canceled"
	case ErrInjectedFailure:
		return "injected_failure"
	case ErrResponseLost:
		return "response_lost"
	default:
		return "other"
	}
}

func createPITREffectTable(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	if _, err := pool.Exec(t.Context(), `CREATE TABLE nodecontrol.authority_task7_effects(operation_id uuid PRIMARY KEY,effect_digest bytea NOT NULL CHECK(octet_length(effect_digest)=32))`); err != nil {
		t.Fatal(err)
	}
}

func queryPITRActiveAuthorityFixture(ctx context.Context, coordinator *Coordinator, pool *pgxpool.Pool) error {
	readiness, err := coordinator.CheckReady(ctx)
	if err != nil {
		return err
	}
	if !readiness.Ready {
		return ErrAuthorityUnavailable
	}
	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM nodecontrol.control_plane_authority_fences WHERE provider_status='committed' AND visibility_state='active'`).Scan(&count); err != nil {
		return ErrInjectedFailure
	}
	return nil
}

func applyPITRAuthoritySchema(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	raw, err := os.ReadFile("../../../db/migrations/00006_nodecontrol.sql")
	if err != nil {
		t.Fatal(err)
	}
	body := string(raw)
	upMarker := strings.Index(body, "-- +goose Up")
	downMarker := strings.Index(body, "-- +goose Down")
	if upMarker < 0 || downMarker <= upMarker {
		t.Fatal("nodecontrol migration lacks ordered goose sections")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	if _, err := pool.Exec(ctx, body[upMarker+len("-- +goose Up"):downMarker]); err != nil {
		t.Fatal(err)
	}
}

const pitrPostgresPassword = "task7-loopback-only-password"

type pitrDockerHarness struct {
	t          *testing.T
	runID      string
	image      string
	containers []string
	volumes    []string
}

func newPITRDockerHarness(t *testing.T) *pitrDockerHarness {
	t.Helper()
	if _, err := exec.LookPath("docker"); err != nil {
		t.Fatal("docker CLI is required for physical PITR integration:", err)
	}
	harness := &pitrDockerHarness{
		t:     t,
		runID: strings.ReplaceAll(uuid.NewString(), "-", ""),
		image: "postgres:18.4-alpine3.23",
	}
	if output, err := harness.docker("image", "inspect", harness.image, "--format", "{{.Id}}"); err != nil {
		t.Fatalf("pinned PostgreSQL image %s is not preloaded: %v: %s", harness.image, err, output)
	}
	t.Logf("Task 7 PITR run_id=%s image=%s", harness.runID, harness.image)
	return harness
}

func (harness *pitrDockerHarness) createVolume(role string) string {
	harness.t.Helper()
	name := "talenro-task7-" + harness.runID + "-" + role
	if _, err := harness.docker("volume", "inspect", name); err == nil {
		harness.t.Fatalf("refusing to adopt pre-existing Docker volume %s", name)
	}
	output, err := harness.docker("volume", "create", "--label", "talenro.task7.run="+harness.runID, name)
	if err != nil || strings.TrimSpace(output) != name {
		harness.t.Fatalf("create volume %s: %v: %s", name, err, output)
	}
	harness.volumes = append(harness.volumes, name)
	harness.t.Logf("Task 7 PITR volume role=%s name=%s", role, name)
	return name
}

func (harness *pitrDockerHarness) startPostgres(role string, port int, dataVolume, dataSubdirectory, backupVolume string) string {
	harness.t.Helper()
	name := "talenro-task7-" + harness.runID + "-" + role
	if _, err := harness.docker("container", "inspect", name); err == nil {
		harness.t.Fatalf("refusing to adopt pre-existing Docker container %s", name)
	}
	dataTarget := "/var/lib/postgresql"
	args := []string{
		"run", "-d", "--pull=never", "--name", name,
		"--label", "talenro.task7.run=" + harness.runID,
		"-e", "POSTGRES_PASSWORD=" + pitrPostgresPassword,
		"-p", fmt.Sprintf("127.0.0.1:%d:5432", port),
		"--mount", "type=volume,src=" + dataVolume + ",dst=" + dataTarget,
	}
	if dataSubdirectory != "" {
		args = append(args, "-e", "PGDATA="+dataTarget+"/"+dataSubdirectory)
	}
	if backupVolume != "" {
		args = append(args, "--mount", "type=volume,src="+backupVolume+",dst=/backup")
	}
	args = append(args, harness.image, "-c", "wal_level=replica", "-c", "max_wal_senders=4")
	containerID, err := harness.docker(args...)
	if err != nil {
		harness.t.Fatalf("start %s: %v: %s", name, err, containerID)
	}
	harness.containers = append(harness.containers, name)
	harness.t.Logf("Task 7 PITR container role=%s name=%s id=%s loopback_port=%d", role, name, strings.TrimSpace(containerID), port)
	return name
}

func (harness *pitrDockerHarness) physicalBaseBackup(primaryName string) {
	harness.t.Helper()
	if output, err := harness.docker("exec", "--user", "root", primaryName, "mkdir", "-p", "/backup/base"); err != nil {
		harness.t.Fatalf("prepare backup directory: %v: %s", err, output)
	}
	if output, err := harness.docker("exec", "--user", "root", primaryName, "chown", "postgres:postgres", "/backup/base"); err != nil {
		harness.t.Fatalf("chown backup directory: %v: %s", err, output)
	}
	started := time.Now()
	if output, err := harness.docker("exec", "--user", "postgres", primaryName, "pg_basebackup", "-D", "/backup/base", "-Fp", "-X", "stream", "-c", "fast", "-U", "postgres"); err != nil {
		harness.t.Fatalf("physical pg_basebackup: %v: %s", err, output)
	}
	harness.t.Logf("Task 7 physical pg_basebackup completed in %s", time.Since(started))
}

func (harness *pitrDockerHarness) stopContainer(name string) {
	harness.t.Helper()
	if output, err := harness.docker("stop", "--time", "20", name); err != nil {
		harness.t.Fatalf("stop container %s: %v: %s", name, err, output)
	}
}

func (harness *pitrDockerHarness) cleanup() {
	for index := len(harness.containers) - 1; index >= 0; index-- {
		name := harness.containers[index]
		if output, err := harness.docker("rm", "--force", "--volumes", name); err != nil && !strings.Contains(output, "No such container") {
			harness.t.Errorf("remove owned container %s: %v: %s", name, err, output)
		}
		for attempt := 1; attempt <= 2; attempt++ {
			if output, err := harness.docker("container", "inspect", name); err == nil {
				harness.t.Errorf("owned container %s remains after cleanup check %d: %s", name, attempt, output)
			}
		}
	}
	for index := len(harness.volumes) - 1; index >= 0; index-- {
		name := harness.volumes[index]
		if output, err := harness.docker("volume", "rm", name); err != nil && !strings.Contains(output, "No such volume") {
			harness.t.Errorf("remove owned volume %s: %v: %s", name, err, output)
		}
		for attempt := 1; attempt <= 2; attempt++ {
			if output, err := harness.docker("volume", "inspect", name); err == nil {
				harness.t.Errorf("owned volume %s remains after cleanup check %d: %s", name, attempt, output)
			}
		}
	}
}

func (harness *pitrDockerHarness) docker(arguments ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, "docker", arguments...)
	output, err := command.CombinedOutput()
	return string(output), err
}

func reservePITRLoopbackPort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	return port
}

func openPITRPool(t *testing.T, port int) *pgxpool.Pool {
	t.Helper()
	url := fmt.Sprintf("postgres://postgres:%s@127.0.0.1:%d/postgres?sslmode=disable", pitrPostgresPassword, port)
	deadline := time.Now().Add(90 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		pool, err := pgxpool.New(context.Background(), url)
		if err == nil {
			pingCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			lastErr = pool.Ping(pingCtx)
			cancel()
			if lastErr == nil {
				t.Cleanup(pool.Close)
				return pool
			}
			pool.Close()
		} else {
			lastErr = err
		}
		time.Sleep(250 * time.Millisecond)
	}
	t.Fatalf("PostgreSQL on loopback port %d did not become ready: %v", port, lastErr)
	return nil
}
