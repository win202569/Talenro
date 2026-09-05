//go:build integration

package testinfra

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nats-io/nats.go"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
	migrations "talenro.local/platform/db/migrations"
	"talenro.local/platform/internal/nodecontrol/authority"
	"talenro.local/platform/internal/nodecontrol/contracts"
)

var c12DatabaseNamePattern = regexp.MustCompile(`^talenro_c12_([0-9a-f]{32})$`)

func TestC12DependenciesAreIsolatedAndBaseMigrated(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	databaseURL, runSuffix := requireC12PostgresEndpoint(t, os.Getenv("TALENRO_DATABASE_URL"))
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal("open C12 PostgreSQL endpoint")
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		t.Fatal("ping C12 PostgreSQL endpoint")
	}
	var serverVersion string
	if err := pool.QueryRow(ctx, `SHOW server_version`).Scan(&serverVersion); err != nil {
		t.Fatal("read C12 PostgreSQL version")
	}
	if serverVersion != "18.4" {
		t.Fatalf("C12 PostgreSQL version = %q, want 18.4", serverVersion)
	}
	var migrationVersion int64
	var migrationApplied bool
	if err := pool.QueryRow(ctx, `SELECT version_id, is_applied FROM public.goose_db_version ORDER BY id DESC LIMIT 1`).Scan(&migrationVersion, &migrationApplied); err != nil {
		t.Fatal("read ordinary Goose migration high-water")
	}
	if migrationVersion != 6 || !migrationApplied {
		t.Fatalf("ordinary Goose migration high-water = (%d,%t), want (6,true)", migrationVersion, migrationApplied)
	}
	var tableCount int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM information_schema.tables WHERE table_schema='nodecontrol' AND table_type='BASE TABLE'`).Scan(&tableCount); err != nil {
		t.Fatal("count base nodecontrol catalog")
	}
	if tableCount != 25 {
		t.Fatalf("base nodecontrol table count = %d, want 25", tableCount)
	}

	redisAddress := requireC12HostPort(t, "TALENRO_REDIS_ADDRESS", os.Getenv("TALENRO_REDIS_ADDRESS"))
	rdb := redis.NewClient(&redis.Options{Addr: redisAddress})
	defer rdb.Close()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Fatal("ping C12 Redis endpoint")
	}
	if size, err := rdb.DBSize(ctx).Result(); err != nil {
		t.Fatal("inspect C12 Redis isolation")
	} else if size != 0 {
		t.Fatalf("C12 Redis database is not isolated: initial key count = %d", size)
	}
	markerKey := "talenro:c12:run:" + runSuffix
	created, err := rdb.SetNX(ctx, markerKey, runSuffix, 5*time.Minute).Result()
	if err != nil || !created {
		t.Fatal("write unique C12 Redis run marker")
	}
	if marker, err := rdb.Get(ctx, markerKey).Result(); err != nil || marker != runSuffix {
		t.Fatal("read exact C12 Redis run marker")
	}

	natsURL := requireC12NATSEndpoint(t, os.Getenv("TALENRO_NATS_URL"))
	nc, err := nats.Connect(natsURL, nats.Timeout(3*time.Second), nats.Name("talenro-c12-dependency-gate"))
	if err != nil {
		t.Fatal("connect C12 NATS endpoint")
	}
	defer nc.Close()
	wantServerName := "talenro-c12-" + runSuffix + "-nats"
	if got := nc.ConnectedServerName(); got != wantServerName {
		t.Fatalf("C12 NATS run marker = %q, want %q", got, wantServerName)
	}
	jetStream, err := nc.JetStream()
	if err != nil {
		t.Fatal("create C12 JetStream context")
	}
	if _, err := jetStream.AccountInfo(nats.Context(ctx)); err != nil {
		t.Fatal("query C12 JetStream account")
	}
}

func TestPrepareC12AuthorityV7Database(t *testing.T) {
	databaseURL, runSuffix := requireC12PostgresEndpoint(t, os.Getenv("TALENRO_DATABASE_URL"))
	if os.Getenv("TALENRO_C12_AUTHORITY_V7_PROFILE") != "authority-v7" && os.Getenv("TALENRO_C12_AUTHORITY_V7_PROFILE") != "authority-v7-pitr" {
		t.Fatal("authority-v7 initializer received an invalid closed profile")
	}
	if os.Getenv("TALENRO_INSTALLATION_KIND") != "disposable_fixture" {
		t.Fatal("authority-v7 initializer requires the disposable fixture marker")
	}
	if got := os.Getenv("TALENRO_C12_AUTHORITY_V7_RUN_SUFFIX"); got != runSuffix {
		t.Fatal("authority-v7 initializer run suffix is not bound to the database endpoint")
	}
	initNonce, err := hex.DecodeString(os.Getenv("TALENRO_C12_AUTHORITY_V7_INIT_NONCE"))
	if err != nil || len(initNonce) != 32 {
		t.Fatal("authority-v7 initializer nonce must be exactly 64 hexadecimal characters")
	}

	database, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatal("open authority-v7 initializer database:", err)
	}
	defer database.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	var highWater int64
	if err := database.QueryRowContext(ctx, `SELECT max(version_id) FROM public.goose_db_version WHERE is_applied`).Scan(&highWater); err != nil || highWater != 6 {
		t.Fatalf("authority-v7 initializer base high-water = %d, %v; want 6", highWater, err)
	}
	var preTableCount, preLatchCount int
	if err := database.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM pg_catalog.pg_class c JOIN pg_catalog.pg_namespace n ON n.oid=c.relnamespace WHERE n.nspname='nodecontrol' AND c.relkind='r'), (SELECT count(*) FROM pg_catalog.pg_class c JOIN pg_catalog.pg_namespace n ON n.oid=c.relnamespace WHERE n.nspname='nodecontrol' AND c.relname='control_plane_authority_protocol_migration_latches')`).Scan(&preTableCount, &preLatchCount); err != nil {
		t.Fatal("inspect authority-v7 initializer base catalog:", err)
	}
	if preTableCount != 25 || preLatchCount != 0 {
		t.Fatalf("authority-v7 initializer base catalog = tables %d latch relations %d, want 25/0", preTableCount, preLatchCount)
	}

	now := time.Now().UTC().Truncate(time.Microsecond)
	digest := func(label string) contracts.Digest {
		return contracts.Digest(sha256.Sum256([]byte("talenro-c12-authority-v7:" + runSuffix + ":" + label)))
	}
	var transactionNonce contracts.Digest
	copy(transactionNonce[:], initNonce)
	latch := contracts.AuthorityV7MigrationLatchFactsV1{
		InstallationID:         uuid.NewSHA1(uuid.NameSpaceOID, []byte("talenro-c12-authority-v7:"+runSuffix)),
		InstallationKind:       contracts.AuthorityV7InstallationKindDisposableFixture,
		MigrationVersion:       7,
		DatabaseIdentityDigest: digest("database-identity"),
		UpCatalogDigest:        digest("catalog-51"),
		DownState:              contracts.AuthorityV7MigrationDownStateLocked,
		InstalledAt:            now,
	}
	facts := contracts.AuthorityV7UpMigrationFactsV1{
		MigrationLatch:              latch,
		MigrationLatchDigest:        digest("migration-latch"),
		AuthorityProtocolProfile:    contracts.AuthorityV7ProtocolProfileLegacyV6,
		LocalRuntimeIsolationDigest: digest("runtime-isolation"),
		TransactionNonce:            transactionNonce,
		ExpiresAt:                   now.Add(90 * time.Second),
	}
	grant, err := authority.NewDisposableAuthorityV7UpGrantForIntegration(facts)
	if err != nil {
		t.Fatal("construct authority-v7 initializer grant:", err)
	}
	repositoryRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	provider, err := goose.NewProvider(
		goose.DialectPostgres,
		database,
		os.DirFS(filepath.Join(repositoryRoot, "db", "migrations")),
		goose.WithDisableGlobalRegistry(true),
		goose.WithGoMigrations(migrations.NodeControlAuthorityV7Migration()),
	)
	if err != nil {
		t.Fatal("construct provider-scoped authority-v7 initializer:", err)
	}
	migrationContext := migrations.WithAuthorityV7MigrationContext(ctx, migrations.AuthorityV7MigrationContext{
		InstallationKind: migrations.InstallationKindDisposableFixture,
		UpGrant:          &grant,
	})
	results, err := provider.UpTo(migrationContext, 7)
	if err != nil {
		t.Fatal("apply provider-scoped authority-v7 initializer:", err)
	}
	if len(results) != 1 || results[0].Source.Version != 7 {
		t.Fatalf("authority-v7 initializer results = %#v, want exactly version 7", results)
	}
}

func TestC12DependenciesAreIsolatedAndAuthorityV7Migrated(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	databaseURL, runSuffix := requireC12PostgresEndpoint(t, os.Getenv("TALENRO_DATABASE_URL"))
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal("open authority-v7 C12 PostgreSQL endpoint")
	}
	defer pool.Close()
	var serverVersion string
	var highWater int64
	var tableCount, latchCount int
	var kind string
	var databaseIdentity []byte
	if err := pool.QueryRow(ctx, `SHOW server_version`).Scan(&serverVersion); err != nil {
		t.Fatal("read authority-v7 PostgreSQL version")
	}
	if err := pool.QueryRow(ctx, `SELECT max(version_id) FROM public.goose_db_version WHERE is_applied`).Scan(&highWater); err != nil {
		t.Fatal("read authority-v7 migration high-water")
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM pg_catalog.pg_class c JOIN pg_catalog.pg_namespace n ON n.oid=c.relnamespace WHERE n.nspname='nodecontrol' AND c.relkind='r'`).Scan(&tableCount); err != nil {
		t.Fatal("count authority-v7 catalog")
	}
	if err := pool.QueryRow(ctx, `SELECT count(*), min(installation_kind), min(database_identity_digest) FROM nodecontrol.control_plane_authority_protocol_migration_latches`).Scan(&latchCount, &kind, &databaseIdentity); err != nil {
		t.Fatal("inspect authority-v7 migration latch")
	}
	wantIdentity := sha256.Sum256([]byte("talenro-c12-authority-v7:" + runSuffix + ":database-identity"))
	if serverVersion != "18.4" || highWater != 7 || tableCount != 51 || latchCount != 1 || kind != "disposable_fixture" || !strings.EqualFold(hex.EncodeToString(databaseIdentity), hex.EncodeToString(wantIdentity[:])) {
		t.Fatalf("authority-v7 profile facts = postgres %s migration %d tables %d latch %d/%s identity %x", serverVersion, highWater, tableCount, latchCount, kind, databaseIdentity)
	}

	redisAddress := requireC12HostPort(t, "TALENRO_REDIS_ADDRESS", os.Getenv("TALENRO_REDIS_ADDRESS"))
	rdb := redis.NewClient(&redis.Options{Addr: redisAddress})
	defer rdb.Close()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Fatal("ping authority-v7 Redis endpoint")
	}
	natsURL := requireC12NATSEndpoint(t, os.Getenv("TALENRO_NATS_URL"))
	nc, err := nats.Connect(natsURL, nats.Timeout(3*time.Second), nats.Name("talenro-c12-authority-v7-gate"))
	if err != nil {
		t.Fatal("connect authority-v7 NATS endpoint")
	}
	defer nc.Close()
}

func TestC12AuthorityPITRProfile(t *testing.T) {
	controller, err := OpenC12AuthorityPITR()
	if err != nil {
		t.Fatal("open closed authority PITR controller:", err)
	}
	cleanupComplete := false
	t.Cleanup(func() {
		if cleanupComplete {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		if err := cleanupC12AuthorityPITR(ctx, controller); err != nil {
			t.Error("cleanup authority PITR controller:", err)
		}
	})
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()
	backup, err := controller.CreateBaseBackup(ctx)
	if err != nil {
		t.Fatal("create real authority PITR base backup:", err)
	}
	if backup.BackupID == "" || !validC12AuthorityPITRLSN(backup.StartLSN) || !validC12AuthorityPITRLSN(backup.EndLSN) {
		t.Fatalf("authority PITR base backup is malformed: %+v", backup)
	}
	if err := issueC12AuthorityPITRRollbackProbe(ctx, controller); err != nil {
		t.Fatal("prove logical rollback invisibility:", err)
	}
	immediate, err := issueC12AuthorityPITRCommit(ctx, controller, C12AuthorityPITRCommitImmediate)
	if err != nil {
		t.Fatal("observe immediate logical terminal:", err)
	}
	prepared, err := issueC12AuthorityPITRCommit(ctx, controller, C12AuthorityPITRCommitPrepared)
	if err != nil {
		t.Fatal("observe prepared logical terminal:", err)
	}
	if immediate.SQLXID == prepared.SQLXID || immediate.GID != "" || prepared.GID == "" || !c12AuthorityPITRLSNGreaterOrEqual(prepared.EndLSN, immediate.EndLSN) {
		t.Fatalf("authority PITR logical terminals are inconsistent: immediate=%+v prepared=%+v", immediate, prepared)
	}

	t.Run("malformed", func(t *testing.T) {
		if _, err := parseC12AuthorityPITRTerminal([]c12AuthorityPITRDecodedRow{{LSN: "bad", XID: 7, Data: "COMMIT 7"}}, C12AuthorityPITRCommitImmediate, 7, ""); err == nil {
			t.Fatal("malformed logical terminal was accepted")
		}
	})
	t.Run("duplicate", func(t *testing.T) {
		rows := []c12AuthorityPITRDecodedRow{{LSN: "0/10", XID: 7, Data: "COMMIT 7"}, {LSN: "0/20", XID: 7, Data: "COMMIT 7"}}
		if _, err := parseC12AuthorityPITRTerminal(rows, C12AuthorityPITRCommitImmediate, 7, ""); err == nil {
			t.Fatal("duplicate logical terminal was accepted")
		}
	})
	t.Run("gid-mismatch", func(t *testing.T) {
		rows := []c12AuthorityPITRDecodedRow{{LSN: "0/10", XID: 7, Data: "COMMIT PREPARED 'wrong'"}}
		if _, err := parseC12AuthorityPITRTerminal(rows, C12AuthorityPITRCommitPrepared, 7, "expected"); err == nil {
			t.Fatal("prepared logical terminal with mismatched GID was accepted")
		}
	})

	cut, err := controller.CrashPrimary(ctx, backup, prepared)
	if err != nil {
		t.Fatal("crash primary at authenticated terminal cut:", err)
	}
	candidate, err := controller.RestoreAtCut(ctx, cut)
	if err != nil {
		t.Fatal("restore real authority PITR candidate:", err)
	}
	candidate, err = controller.PromoteCandidate(ctx, candidate)
	if err != nil {
		t.Fatal("promote real authority PITR candidate:", err)
	}
	timeline, err := controller.InspectTimeline(ctx, candidate)
	if err != nil {
		t.Fatal("inspect promoted authority PITR timeline:", err)
	}
	if !timeline.Promoted || timeline.CandidateIndex != candidate.Index || timeline.TimelineID < 2 || !c12AuthorityPITRLSNGreaterOrEqual(timeline.ReplayLSN, prepared.EndLSN) {
		t.Fatalf("promoted authority PITR timeline is inconsistent: %+v", timeline)
	}
	if err := cleanupC12AuthorityPITR(ctx, controller); err != nil {
		t.Fatal("verify ownership WAL and exact candidate cleanup:", err)
	}
	cleanupComplete = true
}

func TestC12AuthorityPITROwnershipWALFailureSeam(t *testing.T) {
	controller, err := OpenC12AuthorityPITR()
	if err != nil {
		t.Fatal("open seam-bound authority PITR controller:", err)
	}
	if controller.state.descriptor.FailureSeam == "" {
		t.Fatal("private authority PITR seam test lacks its signed closed seam")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()
	backup, err := controller.CreateBaseBackup(ctx)
	if err != nil {
		t.Fatal("create authority PITR seam base backup:", err)
	}
	terminal, err := issueC12AuthorityPITRCommit(ctx, controller, C12AuthorityPITRCommitImmediate)
	if err != nil {
		t.Fatal("create authority PITR seam terminal:", err)
	}
	cut, err := controller.CrashPrimary(ctx, backup, terminal)
	if err != nil {
		t.Fatal("crash authority PITR seam primary:", err)
	}
	candidate, restoreErr := controller.RestoreAtCut(ctx, cut)
	creationSeam := strings.HasPrefix(controller.state.descriptor.FailureSeam, "after-intent-") ||
		strings.HasPrefix(controller.state.descriptor.FailureSeam, "after-create-") ||
		strings.HasPrefix(controller.state.descriptor.FailureSeam, "after-actual-")
	if creationSeam {
		if restoreErr == nil || !strings.Contains(restoreErr.Error(), controller.state.descriptor.FailureSeam) {
			t.Fatalf("authority PITR creation seam %s did not interrupt at its exact boundary: %v", controller.state.descriptor.FailureSeam, restoreErr)
		}
	} else if restoreErr != nil {
		t.Fatal("restore authority PITR cleanup-seam candidate:", restoreErr)
	} else if candidate.Name == "" {
		t.Fatal("authority PITR cleanup seam did not create a candidate")
	}
	if err := cleanupC12AuthorityPITR(ctx, controller); err != nil {
		t.Fatal("authority PITR seam recovery did not converge to exact absence:", err)
	}
}

func requireC12PostgresEndpoint(t *testing.T, raw string) (string, string) {
	t.Helper()
	if raw == "" {
		t.Fatal("TALENRO_DATABASE_URL is missing")
	}
	parsed, err := url.Parse(raw)
	if err != nil || (parsed.Scheme != "postgres" && parsed.Scheme != "postgresql") || parsed.Hostname() != "127.0.0.1" {
		t.Fatal("TALENRO_DATABASE_URL is not an isolated loopback PostgreSQL endpoint")
	}
	if parsed.User == nil || parsed.User.Username() != "talenro" {
		t.Fatal("TALENRO_DATABASE_URL has the wrong run-owned user")
	}
	if _, present := parsed.User.Password(); !present {
		t.Fatal("TALENRO_DATABASE_URL lacks its run-owned credential")
	}
	if err := requireC12Port(parsed.Port()); err != nil {
		t.Fatal("TALENRO_DATABASE_URL has an invalid mapped port")
	}
	databaseName := strings.TrimPrefix(parsed.EscapedPath(), "/")
	match := c12DatabaseNamePattern.FindStringSubmatch(databaseName)
	if match == nil {
		t.Fatal("TALENRO_DATABASE_URL lacks the cryptographic C12 run marker")
	}
	return raw, match[1]
}

func requireC12HostPort(t *testing.T, variable, raw string) string {
	t.Helper()
	host, port, err := net.SplitHostPort(raw)
	if err != nil || host != "127.0.0.1" || requireC12Port(port) != nil {
		t.Fatalf("%s is not an isolated loopback endpoint", variable)
	}
	return raw
}

func requireC12NATSEndpoint(t *testing.T, raw string) string {
	t.Helper()
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "nats" || parsed.Hostname() != "127.0.0.1" || requireC12Port(parsed.Port()) != nil || parsed.User != nil || parsed.Path != "" {
		t.Fatal("TALENRO_NATS_URL is not an isolated loopback NATS endpoint")
	}
	return raw
}

func requireC12Port(raw string) error {
	port, err := strconv.Atoi(raw)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("invalid port")
	}
	return nil
}
