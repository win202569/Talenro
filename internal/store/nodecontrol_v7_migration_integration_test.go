//go:build integration

package store_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gowebpki/jcs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	migrations "talenro.local/platform/db/migrations"
	"talenro.local/platform/internal/nodecontrol/authority"
	"talenro.local/platform/internal/nodecontrol/contracts"
)

func TestNodeControlV7LegacyShapes(t *testing.T) {
	// The authority-v7 runner already owns and migrated this exact database.
	// Reset that one database through the production Down path; creating a
	// per-shape scratch database would bypass cluster-global role semantics and
	// could never prove the unchanged combined five-minute gate.
	databaseURL := os.Getenv("TALENRO_DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("TALENRO_DATABASE_URL is required")
	}
	database, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	repositoryRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	migrationFS := os.DirFS(filepath.Join(repositoryRoot, "db", "migrations"))
	phase := "production Down reset"
	t.Cleanup(func() {
		defer database.Close()
		var version int64
		if queryErr := database.QueryRowContext(context.Background(), `SELECT coalesce(max(version_id) FILTER (WHERE is_applied),0) FROM public.goose_db_version`).Scan(&version); queryErr != nil {
			t.Errorf("legacy phase %q cleanup cannot inspect migration version: %v", phase, queryErr)
			return
		}
		if version == 6 {
			installAuthorityV7Fixture(t, database, migrationFS, "legacy-shapes-cleanup")
		}
	})
	task8ResetFormalAuthorityV7ToBase(t, database, migrationFS)
	phase = "seed five legacy shapes"
	baseTime := time.Date(2026, time.August, 29, 12, 34, 56, 123456000, time.UTC)
	fixtures := []struct {
		name                               string
		operationID                        uuid.UUID
		effectDigest, receiptDigest        []byte
		providerStatus, visibilityState    string
		databaseSystemID, databaseTimeline any
		requiredLSN                        any
		abortReason                        any
		effectBoundAt, terminalAt          any
	}{
		{name: "R0", operationID: uuid.MustParse("71000000-0000-4000-8000-000000000001"), providerStatus: "reserved", visibilityState: "fence_pending"},
		{name: "R1", operationID: uuid.MustParse("71000000-0000-4000-8000-000000000002"), effectDigest: task8LegacyDigest(0x31), providerStatus: "reserved", visibilityState: "fence_pending", databaseSystemID: "2718281828459045235", databaseTimeline: int64(17), requiredLSN: "0/1700000", effectBoundAt: baseTime.Add(time.Second)},
		{name: "C", operationID: uuid.MustParse("71000000-0000-4000-8000-000000000003"), effectDigest: task8LegacyDigest(0x43), receiptDigest: task8LegacyDigest(0x53), providerStatus: "committed", visibilityState: "active", databaseSystemID: "3141592653589793238", databaseTimeline: int64(19), requiredLSN: "0/1900000", effectBoundAt: baseTime.Add(2 * time.Second), terminalAt: baseTime.Add(3 * time.Second)},
		{name: "A0", operationID: uuid.MustParse("71000000-0000-4000-8000-000000000004"), receiptDigest: task8LegacyDigest(0x64), providerStatus: "aborted", visibilityState: "aborted", abortReason: "provider_dependency_failed", terminalAt: baseTime.Add(4 * time.Second)},
		{name: "A1", operationID: uuid.MustParse("71000000-0000-4000-8000-000000000005"), effectDigest: task8LegacyDigest(0x75), receiptDigest: task8LegacyDigest(0x85), providerStatus: "aborted", visibilityState: "aborted", databaseSystemID: "1618033988749894848", databaseTimeline: int64(23), requiredLSN: "0/2300000", abortReason: "activation_deadline_expired", effectBoundAt: baseTime.Add(5 * time.Second), terminalAt: baseTime.Add(6 * time.Second)},
	}
	type legacySnapshot struct {
		fence, transition, inventory, pop []byte
	}
	before := make(map[uuid.UUID]legacySnapshot, len(fixtures))
	for sequence, fixture := range fixtures {
		task8SeedLegacyFenceAndDomainRow(t, database, fixture.operationID, int64(sequence+1), baseTime,
			fixture.effectDigest, fixture.receiptDigest, fixture.providerStatus, fixture.visibilityState,
			fixture.databaseSystemID, fixture.databaseTimeline, fixture.requiredLSN, fixture.abortReason,
			fixture.effectBoundAt, fixture.terminalAt)
		before[fixture.operationID] = task8LegacySnapshot(t, database, fixture.operationID, false)
	}

	// One database and one Up are deliberate: the five shapes are independent
	// legacy rows, not five empty-database migration smoke tests. This also
	// keeps the unchanged five-minute integration deadline meaningful.
	installAuthorityV7Fixture(t, database, migrationFS, "legacy-shapes-r0-r1-c-a0-a1")

	pool, err := pgxpool.New(t.Context(), databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	repository, err := authority.NewPostgresRepository(pool)
	if err != nil {
		t.Fatal(err)
	}

	for _, fixture := range fixtures {
		fixture := fixture
		t.Run(fixture.name, func(t *testing.T) {
			var profile string
			var abortClaimedAt, activationID any
			if err := database.QueryRowContext(t.Context(), `SELECT authority_protocol_profile,abort_claimed_at,protocol_activation_id FROM nodecontrol.control_plane_authority_fences WHERE operation_id=$1`, fixture.operationID).Scan(&profile, &abortClaimedAt, &activationID); err != nil {
				t.Fatal("read v7 legacy fence profile:", err)
			}
			if profile != "legacy_v6" || abortClaimedAt != nil || activationID != nil {
				t.Fatalf("legacy shape profile/claim/activation = %q/%v/%v, want literal legacy_v6/nil/nil", profile, abortClaimedAt, activationID)
			}

			after := task8LegacySnapshot(t, database, fixture.operationID, true)
			want := before[fixture.operationID]
			if !bytes.Equal(after.fence, want.fence) || !bytes.Equal(after.transition, want.transition) ||
				!bytes.Equal(after.inventory, want.inventory) || !bytes.Equal(after.pop, want.pop) {
				t.Fatalf("legacy %s seeded/FK bytes changed across the single Up\nfence before=%s\nfence after =%s\ntransition before=%s\ntransition after =%s\ninventory before=%s\ninventory after =%s\npop before=%s\npop after =%s",
					fixture.name, want.fence, after.fence, want.transition, after.transition,
					want.inventory, after.inventory, want.pop, after.pop)
			}
			var proofFields int
			if err := database.QueryRowContext(t.Context(), `
SELECT num_nonnulls(authority_effect_kind,authority_effect_disposition,
 authority_effect_commitment_jcs,authority_effect_commitment_digest,authority_provider_head_jcs,authority_provider_head_digest,
 authority_checkpoint_anchor_jcs,authority_checkpoint_anchor_digest,authority_effect_reason,authority_attestation_expires_at,
 authority_activation_deadline,authority_expected_provider_identity_digest,authority_activation_evidence_jcs,
 authority_activation_evidence_digest,authority_effect_resolution_jcs,authority_effect_resolution_digest)
FROM nodecontrol.node_state_transitions WHERE authority_operation_id=$1`, fixture.operationID).Scan(&proofFields); err != nil {
				t.Fatal("inspect legacy proof fields:", err)
			}
			if proofFields != 0 {
				t.Fatalf("legacy domain proof non-NULL count = %d, want 0", proofFields)
			}

			stored, err := repository.GetStoredFence(t.Context(), fixture.operationID)
			if err != nil {
				t.Fatalf("legacy %s GetStoredFence inferred history instead of projecting the row: %v", fixture.name, err)
			}
			if stored.AbortClaim != nil || stored.PersistedOutcome != nil {
				t.Fatalf("legacy %s synthesized v7 claim/outcome = %#v/%#v, want nil/nil", fixture.name, stored.AbortClaim, stored.PersistedOutcome)
			}
		})
	}
}

func task8ResetFormalAuthorityV7ToBase(t *testing.T, database *sql.DB, migrationFS fs.FS) {
	t.Helper()
	now := time.Now().UTC().Truncate(time.Microsecond)
	facts := disposableAuthorityV7DownFacts(now, "legacy-shapes-formal-reset")
	var installationKind, downState string
	var migrationVersion int64
	var databaseIdentity, upCatalog, latchDigest []byte
	if err := database.QueryRowContext(t.Context(), `
SELECT installation_id,installation_kind,migration_version,database_identity_digest,up_catalog_digest,
       down_state,installed_at,body_digest
FROM nodecontrol.control_plane_authority_protocol_migration_latches`).Scan(
		&facts.MigrationLatch.InstallationID, &installationKind, &migrationVersion, &databaseIdentity, &upCatalog,
		&downState, &facts.MigrationLatch.InstalledAt, &latchDigest,
	); err != nil {
		t.Fatal("read formal authority-v7 migration latch for production Down reset:", err)
	}
	facts.MigrationLatch.InstallationKind = contracts.AuthorityV7InstallationKindV1(installationKind)
	facts.MigrationLatch.MigrationVersion = uint64(migrationVersion)
	copy(facts.MigrationLatch.DatabaseIdentityDigest[:], databaseIdentity)
	copy(facts.MigrationLatch.UpCatalogDigest[:], upCatalog)
	facts.MigrationLatch.DownState = contracts.AuthorityV7MigrationDownStateV1(downState)
	copy(facts.MigrationLatchDigest[:], latchDigest)
	facts.CurrentCatalogDigest = facts.MigrationLatch.UpCatalogDigest
	facts.ProviderRetirementSet.InstallationID = facts.MigrationLatch.InstallationID
	for index := range facts.ProviderRetirementSet.RetiredMembers {
		facts.ProviderRetirementSet.RetiredMembers[index].DatabaseIdentityDigest = facts.MigrationLatch.DatabaseIdentityDigest
	}
	grant, err := task8NewDisposableAuthorityV7DownGrant(t, facts)
	if err != nil {
		t.Fatal("construct formal database production Down reset grant:", err)
	}
	ctx := migrations.WithAuthorityV7MigrationContext(t.Context(), migrations.AuthorityV7MigrationContext{
		InstallationKind: migrations.InstallationKindDisposableFixture,
		DownGrant:        &grant,
		DownAuthorizer:   &fixtureAuthorityV7DownAuthorizer{},
	})
	result, err := newAuthorityV7Provider(t, database, migrationFS).Down(ctx)
	if err != nil {
		t.Fatal("legacy phase production Down reset semantic reject:", err)
	}
	if result.Source.Version != 7 {
		t.Fatalf("legacy production Down reset version = %d, want 7", result.Source.Version)
	}
}

func task8LegacySnapshot(t *testing.T, database *sql.DB, operationID uuid.UUID, postV7 bool) struct {
	fence, transition, inventory, pop []byte
} {
	t.Helper()
	fenceProjection := "to_jsonb(f)"
	transitionProjection := "to_jsonb(s)"
	if postV7 {
		fenceProjection = "(to_jsonb(f)-'authority_protocol_profile'-'abort_claimed_at'-'protocol_activation_id')"
		transitionProjection = `(to_jsonb(s)
 -'authority_effect_kind'-'authority_effect_disposition'
 -'authority_effect_commitment_jcs'-'authority_effect_commitment_digest'
 -'authority_provider_head_jcs'-'authority_provider_head_digest'
 -'authority_checkpoint_anchor_jcs'-'authority_checkpoint_anchor_digest'
 -'authority_effect_reason'-'authority_attestation_expires_at'-'authority_activation_deadline'
 -'authority_expected_provider_identity_digest'-'authority_activation_evidence_jcs'-'authority_activation_evidence_digest'
 -'authority_effect_resolution_jcs'-'authority_effect_resolution_digest')`
	}
	var snapshot struct {
		fence, transition, inventory, pop []byte
	}
	query := fmt.Sprintf(`
SELECT convert_to(%s::text,'UTF8'), convert_to(%s::text,'UTF8'),
       convert_to(to_jsonb(i)::text,'UTF8'), convert_to(to_jsonb(p)::text,'UTF8')
FROM nodecontrol.control_plane_authority_fences AS f
JOIN nodecontrol.node_state_transitions AS s ON s.authority_operation_id=f.operation_id
JOIN nodecontrol.node_inventory AS i ON i.node_id=s.node_id
JOIN nodecontrol.node_pops AS p ON p.pop_code=i.pop_code
WHERE f.operation_id=$1`, fenceProjection, transitionProjection)
	if err := database.QueryRowContext(t.Context(), query, operationID).Scan(
		&snapshot.fence, &snapshot.transition, &snapshot.inventory, &snapshot.pop,
	); err != nil {
		t.Fatal("snapshot complete legacy fence/domain/FK bytes:", err)
	}
	return snapshot
}

func task8LegacyDigest(value byte) []byte {
	return bytes.Repeat([]byte{value}, 32)
}

func task8SeedLegacyFenceAndDomainRow(
	t *testing.T,
	database *sql.DB,
	operationID uuid.UUID,
	sequence int64,
	reservedAt time.Time,
	effectDigest, receiptDigest []byte,
	providerStatus, visibilityState string,
	databaseSystemID, databaseTimeline, requiredLSN, abortReason, effectBoundAt, terminalAt any,
) {
	t.Helper()
	popCode := fmt.Sprintf("legacy-%d", sequence)
	nodeID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("task8-legacy-node:"+popCode))
	scopeTranscript := append([]byte("TALENRO-NODE-AUTHORITY-SCOPE-V1\x00"), nodeID[:]...)
	scopeDigest := sha256.Sum256(scopeTranscript)
	reservationDigest := task8LegacyAuthorityDigest(t,
		"TALENRO-CONTROL-PLANE-AUTHORITY-RESERVATION-V1\x00",
		map[string]any{
			"authority_epoch": "71", "authority_sequence": strconv.FormatInt(sequence, 10),
			"effect_kind": "operator_transition", "operation_id": operationID.String(),
			"scope_digest": hex.EncodeToString(scopeDigest[:]), "scope_kind": "node",
		})
	if receiptDigest != nil {
		var databasePoint any
		if databaseSystemID != nil {
			databasePoint = map[string]any{
				"db_system_id": fmt.Sprint(databaseSystemID), "db_timeline": fmt.Sprint(databaseTimeline),
				"required_lsn": fmt.Sprint(requiredLSN),
			}
		}
		var effectDigestHex any
		if effectDigest != nil {
			effectDigestHex = hex.EncodeToString(effectDigest)
		}
		receiptDigest = task8LegacyAuthorityDigest(t,
			"TALENRO-CONTROL-PLANE-AUTHORITY-RECEIPT-V1\x00",
			map[string]any{
				"abort_reason": abortReason, "authority_epoch": "71",
				"authority_sequence": strconv.FormatInt(sequence, 10), "database_point": databasePoint,
				"effect_digest": effectDigestHex, "effect_kind": "operator_transition",
				"operation_id": operationID.String(), "reservation_digest": hex.EncodeToString(reservationDigest),
				"scope_digest": hex.EncodeToString(scopeDigest[:]), "scope_kind": "node", "status": providerStatus,
			})
	}
	if _, err := database.ExecContext(t.Context(), `
INSERT INTO nodecontrol.control_plane_authority_fences(
 operation_id,effect_kind,scope_kind,authority_epoch,authority_sequence,scope_digest,provider_reservation_digest,
	 provider_status,visibility_state,reserved_at)
VALUES($1,'operator_transition','node',71,$2,$3,$4,'reserved','fence_pending',$5)`,
		operationID, sequence, scopeDigest[:], reservationDigest, reservedAt); err != nil {
		t.Fatal("seed legacy unbound fence reservation:", err)
	}
	if effectDigest != nil || providerStatus != "reserved" || receiptDigest != nil || databaseSystemID != nil ||
		databaseTimeline != nil || requiredLSN != nil || abortReason != nil || visibilityState != "fence_pending" ||
		effectBoundAt != nil || terminalAt != nil {
		if _, err := database.ExecContext(t.Context(), `
UPDATE nodecontrol.control_plane_authority_fences
SET effect_digest=$2,provider_status=$3,provider_receipt_digest=$4,db_system_id=$5,db_timeline=$6,
    required_lsn=$7,abort_reason=$8,visibility_state=$9,effect_bound_at=$10,terminal_at=$11
WHERE operation_id=$1`, operationID, effectDigest, providerStatus, receiptDigest, databaseSystemID,
			databaseTimeline, requiredLSN, abortReason, visibilityState, effectBoundAt, terminalAt); err != nil {
			t.Fatal("advance legacy fence through its V6 state machine:", err)
		}
	}
	if _, err := database.ExecContext(t.Context(), `INSERT INTO nodecontrol.node_pops(pop_code,iso_country,region,operator_state,created_at,updated_at) VALUES($1,'US','legacy-region','enabled',$2,$2)`, popCode, reservedAt); err != nil {
		t.Fatal("seed legacy POP:", err)
	}
	if _, err := database.ExecContext(t.Context(), `INSERT INTO nodecontrol.node_inventory(node_id,pop_code,operator_state,security_state,identity_state,created_at,updated_at) VALUES($1,$2,'enabled','normal','never_enrolled',$3,$3)`, nodeID, popCode, reservedAt); err != nil {
		t.Fatal("seed legacy inventory:", err)
	}
	if _, err := database.ExecContext(t.Context(), `
INSERT INTO nodecontrol.node_state_transitions(
 transition_id,node_id,authority_operation_id,authority_epoch,authority_sequence,dimension,from_state,to_state,reason,
 aggregate_version,occurred_at,retention_until)
VALUES($1,$2,$3,71,$4,'operator','enabled','draining','drain_maintenance',1,$5,$6)`,
		uuid.NewSHA1(uuid.NameSpaceOID, []byte("task8-legacy-transition:"+popCode)), nodeID, operationID, sequence,
		reservedAt, reservedAt.Add(180*24*time.Hour)); err != nil {
		t.Fatal("seed legacy authority-bearing domain row:", err)
	}
}

func task8LegacyAuthorityDigest(t *testing.T, domain string, payload any) []byte {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatal("marshal legacy authority digest payload:", err)
	}
	canonical, err := jcs.Transform(body)
	if err != nil {
		t.Fatal("canonicalize legacy authority digest payload:", err)
	}
	hash := sha256.New()
	_, _ = hash.Write([]byte(domain))
	_, _ = hash.Write(canonical)
	return hash.Sum(nil)
}

func TestNodeControlV7UpFromR0R1CA0A1(t *testing.T) {
	for _, fixture := range []struct {
		name       string
		baseMarker string
	}{
		{name: "R0", baseMarker: "empty"},
		{name: "R1", baseMarker: "pending-fence"},
		{name: "C", baseMarker: "committed-fence"},
		{name: "A0", baseMarker: "legacy-aborted"},
		{name: "A1", baseMarker: "legacy-aborted-bound"},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			database, migrationFS, _ := openOwnedAuthorityV7Database(t)
			applyAuthorityV7Base(t, database, migrationFS)
			installAuthorityV7Fixture(t, database, migrationFS, fixture.name)

			var version int64
			if err := database.QueryRowContext(t.Context(), `SELECT max(version_id) FROM public.goose_db_version WHERE is_applied`).Scan(&version); err != nil {
				t.Fatal("inspect migration high-water:", err)
			}
			if version != 7 {
				t.Fatalf("migration high-water = %d, want 7", version)
			}
			var tableCount int
			if err := database.QueryRowContext(t.Context(), `SELECT count(*) FROM pg_catalog.pg_class c JOIN pg_catalog.pg_namespace n ON n.oid=c.relnamespace WHERE n.nspname='nodecontrol' AND c.relkind='r'`).Scan(&tableCount); err != nil {
				t.Fatal("inspect v7 table count:", err)
			}
			if tableCount != 51 {
				t.Fatalf("nodecontrol table count = %d, want 51", tableCount)
			}
			var latchCount int
			if err := database.QueryRowContext(t.Context(), `SELECT count(*) FROM nodecontrol.control_plane_authority_protocol_migration_latches`).Scan(&latchCount); err != nil {
				t.Fatal("inspect v7 migration latch:", err)
			}
			if latchCount != 1 {
				t.Fatalf("v7 latch count = %d, want 1", latchCount)
			}
			t.Logf("authority v7 installed from %s base marker %s", fixture.name, fixture.baseMarker)
		})
	}
}

func TestNodeControlV7CatalogRoundTrip(t *testing.T) {
	database, migrationFS, _ := openOwnedAuthorityV7Database(t)
	applyAuthorityV7Base(t, database, migrationFS)
	installAuthorityV7Fixture(t, database, migrationFS, "roundtrip")

	var names []string
	rows, err := database.QueryContext(t.Context(), `SELECT c.relname FROM pg_catalog.pg_class c JOIN pg_catalog.pg_namespace n ON n.oid=c.relnamespace WHERE n.nspname='nodecontrol' AND c.relkind='r' ORDER BY c.relname COLLATE "C"`)
	if err != nil {
		t.Fatal("query v7 catalog:", err)
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatal("scan v7 catalog:", err)
		}
		names = append(names, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatal("iterate v7 catalog:", err)
	}
	if len(names) != 51 {
		t.Fatalf("v7 catalog size = %d, want 51", len(names))
	}
	if strings.Join(names, "\n") == "" {
		t.Fatal("v7 catalog digest input is empty")
	}
}

func TestNodeControlV7ManifestMatchesNormalizedCatalog(t *testing.T) {
	database, migrationFS, databaseURL := openOwnedAuthorityV7Database(t)
	applyAuthorityV7Base(t, database, migrationFS)
	installAuthorityV7Fixture(t, database, migrationFS, "manifest-roundtrip")
	raw, err := os.ReadFile("../../db/schema/nodecontrol.v1.yaml")
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := decodeNodeControlManifest(raw)
	if err != nil {
		t.Fatal(err)
	}
	manifest = final51Manifest(manifest)
	pool, err := pgxpool.New(t.Context(), databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	rows, err := pool.Query(t.Context(), `
SELECT c.relname FROM pg_catalog.pg_class c
JOIN pg_catalog.pg_namespace n ON n.oid=c.relnamespace
WHERE n.nspname='nodecontrol' AND c.relkind='r' ORDER BY c.relname COLLATE "C"`)
	if err != nil {
		t.Fatal(err)
	}
	actualNames, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		t.Fatal(err)
	}
	assertExactNamedSet(t, "authority-v7 normalized catalog tables", actualNames, tableNames(manifest.Tables))
	for _, table := range manifest.Tables {
		assertCatalogColumns(t.Context(), t, pool, table)
		assertCatalogConstraints(t.Context(), t, pool, table)
		assertCatalogColumnChecks(t.Context(), t, pool, table)
		assertCatalogIndexes(t.Context(), t, pool, table)
		assertCatalogTriggers(t.Context(), t, pool, table)
	}
	task8AssertSourceGuardTriggerMatrix(t, database)
	assertAuthorityV7FunctionsMatchManifest(t.Context(), t, pool, manifest.Functions)
}

func assertAuthorityV7FunctionsMatchManifest(ctx context.Context, t *testing.T, pool *pgxpool.Pool, want []nodeControlFunctionSpec) {
	t.Helper()
	rows, err := pool.Query(ctx, `
SELECT p.proname, pg_catalog.pg_get_function_arguments(p.oid), l.lanname,
       CASE p.provolatile WHEN 'i' THEN 'immutable' WHEN 's' THEN 'stable' ELSE 'volatile' END,
       p.prosecdef, pg_catalog.pg_get_functiondef(p.oid)
FROM pg_catalog.pg_proc p
JOIN pg_catalog.pg_namespace n ON n.oid=p.pronamespace
JOIN pg_catalog.pg_language l ON l.oid=p.prolang
WHERE n.nspname='nodecontrol'
ORDER BY p.proname COLLATE "C", pg_catalog.pg_get_function_identity_arguments(p.oid) COLLATE "C"`)
	if err != nil {
		t.Fatal("query authority-v7 functions:", err)
	}
	type catalogFunction struct {
		name, arguments, language, volatility, definition string
		securityDefiner                                   bool
	}
	actual := make(map[string]catalogFunction, len(want))
	for rows.Next() {
		var function catalogFunction
		if err := rows.Scan(&function.name, &function.arguments, &function.language, &function.volatility, &function.securityDefiner, &function.definition); err != nil {
			rows.Close()
			t.Fatal("scan authority-v7 function:", err)
		}
		key := function.name + "(" + function.arguments + ")"
		if _, duplicate := actual[key]; duplicate {
			rows.Close()
			t.Fatal("duplicate authority-v7 function identity:", key)
		}
		actual[key] = function
	}
	if err := rows.Err(); err != nil {
		t.Fatal("iterate authority-v7 functions:", err)
	}
	rows.Close()
	mismatches := make([]string, 0)
	if len(actual) != len(want) {
		mismatches = append(mismatches, fmt.Sprintf("authority-v7 function count = %d, manifest = %d", len(actual), len(want)))
	}
	for _, expected := range want {
		key := expected.Name + "(" + expected.Arguments + ")"
		got, exists := actual[key]
		if !exists {
			mismatches = append(mismatches, fmt.Sprintf("authority-v7 catalog omits manifest function %s", key))
			continue
		}
		digest := sha256.Sum256([]byte(got.definition))
		if got.language != expected.Language || got.volatility != expected.Volatility || got.securityDefiner != expected.SecurityDefiner || fmt.Sprintf("%x", digest[:]) != expected.DefinitionSHA256 {
			mismatches = append(mismatches, fmt.Sprintf("authority-v7 function %s mismatch: language=%s volatility=%s security_definer=%t sha256=%x; manifest=%+v", key, got.language, got.volatility, got.securityDefiner, digest, expected))
		}
	}
	if len(mismatches) != 0 {
		sort.Strings(mismatches)
		t.Fatalf("authority-v7 function catalog mismatches:\n%s", strings.Join(mismatches, "\n"))
	}
}

func openOwnedAuthorityV7Database(t *testing.T) (*sql.DB, fs.FS, string) {
	t.Helper()
	rawURL := os.Getenv("TALENRO_DATABASE_URL")
	if rawURL == "" {
		t.Fatal("TALENRO_DATABASE_URL is required")
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatal("parse integration database URL:", err)
	}
	admin, err := sql.Open("pgx", rawURL)
	if err != nil {
		t.Fatal("open integration database admin connection:", err)
	}
	lockConnection, err := admin.Conn(t.Context())
	if err != nil {
		admin.Close()
		t.Fatal("reserve cluster-wide authority-v7 fixture lock connection:", err)
	}
	const fixtureLock int64 = 0x54414c454e524f37
	if _, err := lockConnection.ExecContext(t.Context(), `SELECT pg_advisory_lock($1)`, fixtureLock); err != nil {
		lockConnection.Close()
		admin.Close()
		t.Fatal("acquire cluster-wide authority-v7 fixture lock:", err)
	}
	var preexistingRoles int
	if err := lockConnection.QueryRowContext(t.Context(), `SELECT count(*) FROM pg_catalog.pg_roles WHERE rolname=ANY($1::text[])`, []string{
		"nodecontrol_upgrade_executor", "nodecontrol_migration_downgrader", "nodecontrol_staging_importer",
	}).Scan(&preexistingRoles); err != nil || preexistingRoles != 0 {
		_, _ = lockConnection.ExecContext(context.Background(), `SELECT pg_advisory_unlock($1)`, fixtureLock)
		lockConnection.Close()
		admin.Close()
		t.Fatalf("authority-v7 fixture requires all capability role names absent, count=%d error=%v", preexistingRoles, err)
	}
	fixtureOwnsRoles := true
	databaseName := "nodecontrol_t8_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	databaseIdentifier := pgx.Identifier{databaseName}.Sanitize()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if _, err := admin.ExecContext(ctx, "CREATE DATABASE "+databaseIdentifier); err != nil {
		admin.Close()
		t.Fatal("create owned authority-v7 database:", err)
	}
	parsed.Path = "/" + databaseName
	database, err := sql.Open("pgx", parsed.String())
	if err != nil {
		admin.Close()
		t.Fatal("open owned authority-v7 database:", err)
	}
	t.Cleanup(func() {
		_ = database.Close()
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cleanupCancel()
		if _, cleanupErr := admin.ExecContext(cleanupCtx, "DROP DATABASE "+databaseIdentifier+" WITH (FORCE)"); cleanupErr != nil {
			t.Error("drop owned authority-v7 database:", cleanupErr)
		}
		if fixtureOwnsRoles {
			if _, cleanupErr := lockConnection.ExecContext(cleanupCtx, `DROP ROLE IF EXISTS nodecontrol_staging_importer,nodecontrol_migration_downgrader,nodecontrol_upgrade_executor`); cleanupErr != nil {
				t.Error("drop exact authority-v7 cluster capability roles:", cleanupErr)
			}
			var remainingRoles int
			if cleanupErr := lockConnection.QueryRowContext(cleanupCtx, `SELECT count(*) FROM pg_catalog.pg_roles WHERE rolname=ANY($1::text[])`, []string{
				"nodecontrol_upgrade_executor", "nodecontrol_migration_downgrader", "nodecontrol_staging_importer",
			}).Scan(&remainingRoles); cleanupErr != nil || remainingRoles != 0 {
				t.Errorf("authority-v7 fixture capability-role residue count=%d error=%v", remainingRoles, cleanupErr)
			}
		}
		if _, cleanupErr := lockConnection.ExecContext(cleanupCtx, `SELECT pg_advisory_unlock($1)`, fixtureLock); cleanupErr != nil {
			t.Error("release cluster-wide authority-v7 fixture lock:", cleanupErr)
		}
		_ = lockConnection.Close()
		_ = admin.Close()
	})
	repositoryRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return database, os.DirFS(filepath.Join(repositoryRoot, "db", "migrations")), parsed.String()
}

func newAuthorityV7Provider(t *testing.T, database *sql.DB, migrationFS fs.FS) *goose.Provider {
	t.Helper()
	provider, err := goose.NewProvider(
		goose.DialectPostgres,
		database,
		migrationFS,
		goose.WithDisableGlobalRegistry(true),
		goose.WithGoMigrations(migrations.NodeControlAuthorityV7Migration()),
	)
	if err != nil {
		t.Fatal("construct provider-scoped authority-v7 migration provider:", err)
	}
	return provider
}

func applyAuthorityV7Base(t *testing.T, database *sql.DB, migrationFS fs.FS) {
	t.Helper()
	results, err := newAuthorityV7Provider(t, database, migrationFS).UpTo(t.Context(), 6)
	if err != nil {
		t.Fatal("apply authority-v7 base migrations:", err)
	}
	if len(results) != 6 {
		t.Fatalf("base migration result count = %d, want 6", len(results))
	}
}

func installAuthorityV7Fixture(t *testing.T, database *sql.DB, migrationFS fs.FS, label string) {
	t.Helper()
	now := time.Now().UTC().Truncate(time.Microsecond)
	facts := disposableAuthorityV7UpFacts(now, label)
	grant, err := authority.NewDisposableAuthorityV7UpGrantForIntegration(facts)
	if err != nil {
		t.Fatal("construct disposable authority-v7 Up grant:", err)
	}
	ctx := migrations.WithAuthorityV7MigrationContext(t.Context(), migrations.AuthorityV7MigrationContext{
		InstallationKind: migrations.InstallationKindDisposableFixture,
		UpGrant:          &grant,
	})
	results, err := newAuthorityV7Provider(t, database, migrationFS).UpTo(ctx, 7)
	if err != nil {
		t.Fatal("apply provider-scoped authority-v7 migration:", err)
	}
	if len(results) != 1 || results[0].Source.Version != 7 {
		t.Fatalf("authority-v7 migration results = %#v, want exactly version 7", results)
	}
}

func disposableAuthorityV7UpFacts(now time.Time, label string) contracts.AuthorityV7UpMigrationFactsV1 {
	digest := func(value string) contracts.Digest {
		return contracts.Digest(sha256.Sum256([]byte("task8-integration:" + label + ":" + value)))
	}
	latch := contracts.AuthorityV7MigrationLatchFactsV1{
		InstallationID:         uuid.NewSHA1(uuid.NameSpaceOID, []byte("task8-integration:"+label)),
		InstallationKind:       contracts.AuthorityV7InstallationKindDisposableFixture,
		MigrationVersion:       7,
		DatabaseIdentityDigest: digest("database-identity"),
		UpCatalogDigest:        digest("up-catalog"),
		DownState:              contracts.AuthorityV7MigrationDownStateLocked,
		InstalledAt:            now,
	}
	return contracts.AuthorityV7UpMigrationFactsV1{
		MigrationLatch:              latch,
		MigrationLatchDigest:        digest("migration-latch"),
		AuthorityProtocolProfile:    contracts.AuthorityV7ProtocolProfileLegacyV6,
		LocalRuntimeIsolationDigest: digest("runtime-isolation"),
		TransactionNonce:            digest("transaction-nonce"),
		ExpiresAt:                   now.Add(5 * time.Minute),
	}
}

func authorityV7CatalogDigest(names []string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(names, "\n"))))
}
