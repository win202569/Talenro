//go:build integration

package store_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"talenro.local/platform/internal/nodecontrol/authority"
	"talenro.local/platform/internal/store"
)

func TestAuthorityEffectGuards(t *testing.T) {
	databaseURL := os.Getenv("TALENRO_DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("TALENRO_DATABASE_URL is required")
	}
	database, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if closeErr := database.Close(); closeErr != nil {
			t.Errorf("close authority guard database: %v", closeErr)
		}
	})
	task8AssertDownFunctionLockBarrier(t, database)
	now := time.Date(2026, time.August, 29, 16, 30, 0, 123456000, time.UTC)
	operationID := uuid.MustParse("74000000-0000-4000-8000-000000000001")
	nodeID := uuid.MustParse("74000000-0000-4000-8000-000000000002")
	t.Cleanup(func() {
		cleanupCtx := context.Background()
		cleanupTx, cleanupErr := database.BeginTx(cleanupCtx, nil)
		if cleanupErr != nil {
			t.Errorf("begin authority guard fixture cleanup: %v", cleanupErr)
			return
		}
		cleanupOpen := true
		defer func() {
			if cleanupOpen {
				_ = cleanupTx.Rollback()
			}
		}()
		execCleanup := func(label, statement string, args ...any) bool {
			if _, cleanupErr := cleanupTx.ExecContext(cleanupCtx, statement, args...); cleanupErr != nil {
				t.Errorf("%s: %v", label, cleanupErr)
				return false
			}
			return true
		}
		if !execCleanup("enable replica mode for authority guard fixture cleanup", `SET LOCAL session_replication_role=replica`) ||
			!execCleanup("delete authority guard fixture transition", `DELETE FROM nodecontrol.node_state_transitions WHERE authority_operation_id=$1`, operationID) ||
			!execCleanup("delete authority guard fixture fence", `DELETE FROM nodecontrol.control_plane_authority_fences WHERE operation_id=$1`, operationID) ||
			!execCleanup("delete authority guard fixture node", `DELETE FROM nodecontrol.node_inventory WHERE node_id=$1`, nodeID) ||
			!execCleanup("delete authority guard fixture POP", `DELETE FROM nodecontrol.node_pops WHERE pop_code='guard-fixture'`) {
			return
		}
		if cleanupErr := cleanupTx.Commit(); cleanupErr != nil {
			t.Errorf("commit authority guard fixture cleanup: %v", cleanupErr)
			return
		}
		cleanupOpen = false
		var transitionCount, fenceCount, nodeCount, popCount int
		if cleanupErr := database.QueryRowContext(cleanupCtx, `
SELECT
 (SELECT count(*) FROM nodecontrol.node_state_transitions WHERE authority_operation_id=$1),
 (SELECT count(*) FROM nodecontrol.control_plane_authority_fences WHERE operation_id=$1),
 (SELECT count(*) FROM nodecontrol.node_inventory WHERE node_id=$2),
 (SELECT count(*) FROM nodecontrol.node_pops WHERE pop_code='guard-fixture')`, operationID, nodeID).Scan(
			&transitionCount, &fenceCount, &nodeCount, &popCount,
		); cleanupErr != nil {
			t.Errorf("verify authority guard fixture cleanup: %v", cleanupErr)
			return
		}
		if transitionCount != 0 || fenceCount != 0 || nodeCount != 0 || popCount != 0 {
			t.Errorf("authority guard fixture cleanup residue transition/fence/node/pop=%d/%d/%d/%d, want 0/0/0/0",
				transitionCount, fenceCount, nodeCount, popCount)
		}
	})
	if _, err := database.ExecContext(t.Context(), `INSERT INTO nodecontrol.node_pops(pop_code,iso_country,region,operator_state,created_at,updated_at) VALUES('guard-fixture','US','guard-region','enabled',$1,$1)`, now); err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(t.Context(), `INSERT INTO nodecontrol.node_inventory(node_id,pop_code,operator_state,security_state,identity_state,created_at,updated_at) VALUES($1,'guard-fixture','enabled','normal','never_enrolled',$2,$2)`, nodeID, now); err != nil {
		t.Fatal(err)
	}
	seedTx, err := database.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := seedTx.ExecContext(t.Context(), `SET LOCAL session_replication_role=replica`); err != nil {
		_ = seedTx.Rollback()
		t.Fatal(err)
	}
	if _, err := seedTx.ExecContext(t.Context(), `
INSERT INTO nodecontrol.control_plane_authority_fences(
 operation_id,effect_kind,scope_kind,authority_epoch,authority_sequence,scope_digest,provider_reservation_digest,
 provider_status,visibility_state,reserved_at,authority_protocol_profile)
VALUES($1,'operator_transition','node',74,1,$3,decode(repeat('42',32),'hex'),'reserved','fence_pending',$2,'legacy_v6')`, operationID, now, task8NodeScopeDigest(nodeID)); err != nil {
		_ = seedTx.Rollback()
		t.Fatal(err)
	}
	if err := seedTx.Commit(); err != nil {
		t.Fatal(err)
	}
	// The role boundary is an exact current_user/owner capability boundary.
	// session_user membership is explicitly not authorization.
	type functionACL struct {
		name, owner     string
		securityDefiner bool
		definition      string
		source          string
	}
	wantACL := map[string]functionACL{
		"begin_staging_import":              {name: "begin_staging_import", owner: "nodecontrol_staging_importer", securityDefiner: true},
		"v7_acquire_source_freeze_for_seal": {name: "v7_acquire_source_freeze_for_seal", owner: "nodecontrol_upgrade_executor", securityDefiner: true},
		"v7_consume_down_guard":             {name: "v7_consume_down_guard", owner: "nodecontrol_migration_downgrader", securityDefiner: true},
		"v7_insert_downgrade_authorization": {name: "v7_insert_downgrade_authorization", owner: "nodecontrol_migration_downgrader", securityDefiner: true},
		"v7_require_role":                   {name: "v7_require_role", owner: "", securityDefiner: false},
	}
	rows, err := database.QueryContext(t.Context(), `
SELECT p.proname,r.rolname,p.prosecdef,pg_catalog.pg_get_functiondef(p.oid),p.prosrc
FROM pg_catalog.pg_proc p
JOIN pg_catalog.pg_namespace n ON n.oid=p.pronamespace
JOIN pg_catalog.pg_roles r ON r.oid=p.proowner
WHERE n.nspname='nodecontrol' AND p.proname=ANY($1::text[])
ORDER BY p.proname COLLATE "C"`, []string{"begin_staging_import", "v7_acquire_source_freeze_for_seal", "v7_consume_down_guard", "v7_insert_downgrade_authorization", "v7_require_role"})
	if err != nil {
		t.Fatal(err)
	}
	seen := make(map[string]int, len(wantACL))
	for rows.Next() {
		var got functionACL
		if err := rows.Scan(&got.name, &got.owner, &got.securityDefiner, &got.definition, &got.source); err != nil {
			t.Fatal(err)
		}
		want, registered := wantACL[got.name]
		if !registered {
			t.Errorf("unexpected restricted-function row %q", got.name)
			continue
		}
		seen[got.name]++
		if got.name == "v7_require_role" {
			if got.securityDefiner {
				t.Errorf("v7_require_role is SECURITY DEFINER: owner=%s", got.owner)
			}
			task8AssertV7RequireRoleStructuralOracle(t, got.source)
			continue
		}
		if got.owner != want.owner || got.securityDefiner != want.securityDefiner {
			t.Errorf("restricted entry %s metadata = owner:%s security_definer:%t, want %s/%t", got.name, got.owner, got.securityDefiner, want.owner, want.securityDefiner)
		}
		literalRole := want.owner
		if !strings.Contains(strings.ToLower(got.definition), "v7_require_role('"+literalRole+"'") {
			t.Errorf("restricted entry %s lacks literal owner check %s", got.name, literalRole)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	rows.Close()
	for name := range wantACL {
		if seen[name] != 1 {
			t.Errorf("restricted-function row %s count=%d, want 1", name, seen[name])
		}
	}

	var rolesExact int
	if err := database.QueryRowContext(t.Context(), `
SELECT count(*) FROM pg_catalog.pg_roles
WHERE rolname=ANY($1::text[]) AND NOT rolcanlogin AND NOT rolsuper AND NOT rolcreatedb AND NOT rolcreaterole
 AND NOT rolreplication AND NOT rolbypassrls AND rolconfig IS NULL`, []string{
		"nodecontrol_upgrade_executor", "nodecontrol_migration_downgrader", "nodecontrol_staging_importer",
	}).Scan(&rolesExact); err != nil || rolesExact != 3 {
		t.Errorf("closed capability role attributes exact count=%d error=%v, want 3", rolesExact, err)
	}
	var memberships int
	if err := database.QueryRowContext(t.Context(), `
SELECT count(*) FROM pg_catalog.pg_auth_members m
JOIN pg_catalog.pg_roles parent ON parent.oid=m.roleid
JOIN pg_catalog.pg_roles member ON member.oid=m.member
WHERE parent.rolname=ANY($1::text[]) OR member.rolname=ANY($1::text[])`, []string{
		"nodecontrol_upgrade_executor", "nodecontrol_migration_downgrader", "nodecontrol_staging_importer",
	}).Scan(&memberships); err != nil || memberships != 0 {
		t.Errorf("capability role memberships=%d error=%v, want 0", memberships, err)
	}
	task8AssertV7RequireRoleIdentityBoundaries(t, database, databaseURL)
	task8AssertSourceFreezeIsolationBoundary(t, database)
	task8AssertAuthorityProofGroupMatrix(t, database)
	task8AssertSourceGuardTriggerMatrix(t, database)
	task8AssertRealAuthorityProofDMLMatrix(t, database, now.Add(time.Hour))
	task8AssertDeferredProofCommitAndTerminalDelete(t, database)
	task8AssertNewLegacyFenceRejected(t, database, now, nodeID)
	task8AssertSourceFreezeAdvisoryOrdering(t, database, databaseURL, now.Add(2*time.Hour))

	// Every negative uses a fresh transaction and compares the complete table
	// snapshot byte-for-byte, so an exception after a partial write cannot pass.
	guardCases := []struct {
		name         string
		sql          string
		args         []any
		wantSQLState string
	}{
		{
			name: "wrong proof owner effect kind",
			sql: `INSERT INTO nodecontrol.node_state_transitions(
 transition_id,node_id,authority_operation_id,authority_epoch,authority_sequence,dimension,from_state,to_state,reason,
 aggregate_version,occurred_at,retention_until,authority_effect_kind,authority_effect_disposition)
VALUES($1,$2,$3,74,1,'operator','enabled','draining','drain_maintenance',1,$4,$5,'identity_epoch_advance','applied')`,
			args:         []any{uuid.MustParse("74000000-0000-4000-8000-000000000011"), nodeID, operationID, now, now.Add(180 * 24 * time.Hour)},
			wantSQLState: "23514",
		},
		{
			name: "partial proof group",
			sql: `INSERT INTO nodecontrol.node_state_transitions(
 transition_id,node_id,authority_operation_id,authority_epoch,authority_sequence,dimension,from_state,to_state,reason,
 aggregate_version,occurred_at,retention_until,authority_effect_kind,authority_effect_disposition,authority_effect_commitment_jcs)
VALUES($1,$2,$3,74,1,'operator','enabled','draining','drain_maintenance',1,$4,$5,'operator_transition','applied',decode('7b7d','hex'))`,
			args:         []any{uuid.MustParse("74000000-0000-4000-8000-000000000012"), nodeID, operationID, now, now.Add(180 * 24 * time.Hour)},
			wantSQLState: "23514",
		},
		{
			name: "legacy fence profile mismatch",
			sql: `INSERT INTO nodecontrol.node_state_transitions(
 transition_id,node_id,authority_operation_id,authority_epoch,authority_sequence,dimension,from_state,to_state,reason,
 aggregate_version,occurred_at,retention_until,authority_effect_kind,authority_effect_disposition)
VALUES($1,$2,$3,74,1,'operator','enabled','draining','drain_maintenance',1,$4,$5,'operator_transition','not_applied')`,
			args:         []any{uuid.MustParse("74000000-0000-4000-8000-000000000013"), nodeID, operationID, now, now.Add(180 * 24 * time.Hour)},
			wantSQLState: "23514",
		},
	}
	for _, testCase := range guardCases {
		t.Run(testCase.name, func(t *testing.T) {
			before := task8GuardTableSnapshot(t, database)
			tx, err := database.BeginTx(t.Context(), nil)
			if err != nil {
				t.Fatal(err)
			}
			_, executionErr := tx.ExecContext(t.Context(), testCase.sql, testCase.args...)
			if executionErr == nil {
				_ = tx.Rollback()
				t.Fatal("guard accepted forbidden real DML")
			}
			var postgresError *pgconn.PgError
			if !errors.As(executionErr, &postgresError) || postgresError.Code != testCase.wantSQLState || !strings.Contains(strings.ToLower(postgresError.Message), "proof") {
				t.Errorf("guard rejection seam = %#v, want SQLSTATE %s authority guard", postgresError, testCase.wantSQLState)
			}
			_ = tx.Rollback()
			after := task8GuardTableSnapshot(t, database)
			if !bytes.Equal(after, before) {
				t.Fatalf("failed guarded DML changed row bytes\nbefore=%s\nafter=%s", before, after)
			}
		})
	}

	task8AssertAuthorityV7OrdinaryDenial(t, database)
}

func task8AssertV7RequireRoleIdentityBoundaries(t *testing.T, database *sql.DB, databaseURL string) {
	t.Helper()
	const downgrader = "nodecontrol_migration_downgrader"
	const roleCall = `SELECT nodecontrol.v7_require_role('nodecontrol_migration_downgrader'::name)`
	var bootstrapSession, bootstrapCurrent string
	if err := database.QueryRowContext(t.Context(), `SELECT session_user,current_user`).Scan(&bootstrapSession, &bootstrapCurrent); err != nil {
		t.Fatal("read bootstrap identities:", err)
	}
	if bootstrapSession != bootstrapCurrent {
		t.Fatalf("initial bootstrap identities session/current=%q/%q, want equal", bootstrapSession, bootstrapCurrent)
	}

	t.Run("bootstrap session plus literal current role succeeds", func(t *testing.T) {
		tx, err := database.BeginTx(t.Context(), nil)
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback()
		if _, err := tx.ExecContext(t.Context(), `SET LOCAL ROLE nodecontrol_migration_downgrader`); err != nil {
			t.Fatal("set literal downgrader role:", err)
		}
		var sessionIdentity, currentIdentity string
		if err := tx.QueryRowContext(t.Context(), `SELECT session_user,current_user`).Scan(&sessionIdentity, &currentIdentity); err != nil {
			t.Fatal("read set-role identities:", err)
		}
		if sessionIdentity != bootstrapSession || currentIdentity != downgrader {
			t.Fatalf("set-role identities session/current=%q/%q, want %q/%q", sessionIdentity, currentIdentity, bootstrapSession, downgrader)
		}
		if _, err := tx.ExecContext(t.Context(), roleCall); err != nil {
			t.Fatal("literal current_user capability was rejected:", err)
		}
		if err := tx.Rollback(); err != nil {
			t.Fatal("roll back literal-role success case:", err)
		}
	})

	t.Run("bootstrap current user is not the literal capability", func(t *testing.T) {
		tx, err := database.BeginTx(t.Context(), nil)
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback()
		var sessionIdentity, currentIdentity string
		if err := tx.QueryRowContext(t.Context(), `SELECT session_user,current_user`).Scan(&sessionIdentity, &currentIdentity); err != nil {
			t.Fatal("read bootstrap transaction identities:", err)
		}
		if sessionIdentity != bootstrapSession || currentIdentity != bootstrapCurrent {
			t.Fatalf("bootstrap transaction identities session/current=%q/%q, want %q/%q", sessionIdentity, currentIdentity, bootstrapSession, bootstrapCurrent)
		}
		_, callErr := tx.ExecContext(t.Context(), roleCall)
		if err := tx.Rollback(); err != nil {
			t.Fatal("roll back bootstrap-current rejection case:", err)
		}
		var postgresError *pgconn.PgError
		const wantMessage = "nodecontrol v7 role nodecontrol_migration_downgrader is required"
		if !errors.As(callErr, &postgresError) || postgresError.Code != "42501" || postgresError.Message != wantMessage {
			t.Fatalf("bootstrap-current role rejection=%v, want PostgreSQL 42501/%q", callErr, wantMessage)
		}
	})

	t.Run("membership cannot substitute for bootstrap session continuity", func(t *testing.T) {
		const ordinaryRole = "task8_v7_require_role_ordinary"
		var preexisting int
		if err := database.QueryRowContext(t.Context(), `SELECT count(*) FROM pg_catalog.pg_roles WHERE rolname=$1`, ordinaryRole).Scan(&preexisting); err != nil {
			t.Fatal("inspect ordinary identity fixture:", err)
		}
		if preexisting != 0 {
			t.Fatalf("ordinary identity fixture role %q already exists; refusing to alter it", ordinaryRole)
		}

		connection, err := database.Conn(t.Context())
		if err != nil {
			t.Fatal("pin identity-boundary connection:", err)
		}
		defer connection.Close()
		tx, err := connection.BeginTx(t.Context(), nil)
		if err != nil {
			t.Fatal("begin identity-boundary transaction:", err)
		}
		defer func() {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
				t.Errorf("roll back identity-boundary transaction: %v", rollbackErr)
			}
			cleanupContext, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			var restoredSession, restoredCurrent string
			if err := connection.QueryRowContext(cleanupContext, `SELECT session_user,current_user`).Scan(&restoredSession, &restoredCurrent); err != nil {
				t.Errorf("read restored pinned identities: %v", err)
			} else if restoredSession != bootstrapSession || restoredCurrent != bootstrapCurrent {
				t.Errorf("restored pinned identities session/current=%q/%q, want %q/%q", restoredSession, restoredCurrent, bootstrapSession, bootstrapCurrent)
			}

			verificationDatabase, err := sql.Open("pgx", databaseURL)
			if err != nil {
				t.Errorf("open independent identity cleanup verifier: %v", err)
				return
			}
			defer verificationDatabase.Close()
			var roleCount, membershipCount int
			if err := verificationDatabase.QueryRowContext(cleanupContext, `SELECT count(*) FROM pg_catalog.pg_roles WHERE rolname=$1`, ordinaryRole).Scan(&roleCount); err != nil {
				t.Errorf("verify ordinary fixture role absence: %v", err)
				return
			}
			if err := verificationDatabase.QueryRowContext(cleanupContext, `
SELECT count(*)
FROM pg_catalog.pg_auth_members AS membership
JOIN pg_catalog.pg_roles AS granted_role ON granted_role.oid=membership.roleid
JOIN pg_catalog.pg_roles AS member_role ON member_role.oid=membership.member
WHERE granted_role.rolname=$1 OR member_role.rolname=$1`, ordinaryRole).Scan(&membershipCount); err != nil {
				t.Errorf("verify ordinary fixture membership absence: %v", err)
				return
			}
			if roleCount != 0 || membershipCount != 0 {
				t.Errorf("ordinary identity fixture cleanup role/membership=%d/%d, want 0/0", roleCount, membershipCount)
			}
		}()

		if _, err := tx.ExecContext(t.Context(), `
CREATE ROLE task8_v7_require_role_ordinary
NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE INHERIT NOREPLICATION NOBYPASSRLS
CONNECTION LIMIT -1 PASSWORD NULL`); err != nil {
			t.Fatal("create ordinary identity fixture:", err)
		}
		if _, err := tx.ExecContext(t.Context(), `GRANT nodecontrol_migration_downgrader TO task8_v7_require_role_ordinary WITH SET TRUE`); err != nil {
			t.Fatal("grant SET-only capability membership fixture:", err)
		}
		if _, err := tx.ExecContext(t.Context(), `SET LOCAL SESSION AUTHORIZATION task8_v7_require_role_ordinary`); err != nil {
			t.Fatal("set ordinary session authorization:", err)
		}
		if _, err := tx.ExecContext(t.Context(), `SET LOCAL ROLE nodecontrol_migration_downgrader`); err != nil {
			t.Fatal("set membership-granted literal role:", err)
		}
		var sessionIdentity, currentIdentity string
		if err := tx.QueryRowContext(t.Context(), `SELECT session_user,current_user`).Scan(&sessionIdentity, &currentIdentity); err != nil {
			t.Fatal("read membership-boundary identities:", err)
		}
		if sessionIdentity != ordinaryRole || currentIdentity != downgrader {
			t.Fatalf("membership-boundary identities session/current=%q/%q, want %q/%q", sessionIdentity, currentIdentity, ordinaryRole, downgrader)
		}
		_, callErr := tx.ExecContext(t.Context(), roleCall)
		var postgresError *pgconn.PgError
		const wantMessage = "authority v7 Down requires the exact bootstrap owner"
		if !errors.As(callErr, &postgresError) || postgresError.Code != "42501" || postgresError.Message != wantMessage {
			t.Fatalf("membership-boundary rejection=%v, want PostgreSQL 42501/%q", callErr, wantMessage)
		}
	})
}

func task8AssertSourceFreezeIsolationBoundary(t *testing.T, database *sql.DB) {
	t.Helper()
	for _, testCase := range []struct {
		name      string
		isolation sql.IsolationLevel
	}{
		{name: "repeatable read", isolation: sql.LevelRepeatableRead},
		{name: "serializable", isolation: sql.LevelSerializable},
	} {
		t.Run("source sealer rejects "+testCase.name, func(t *testing.T) {
			tx, err := database.BeginTx(t.Context(), &sql.TxOptions{Isolation: testCase.isolation})
			if err != nil {
				t.Fatal("begin non-read-committed source-seal transaction:", err)
			}
			_, callErr := tx.ExecContext(t.Context(), `SELECT nodecontrol.v7_acquire_source_freeze_for_seal()`)
			if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
				t.Fatal("roll back rejected source-seal transaction:", err)
			}
			var postgresError *pgconn.PgError
			const wantMessage = "authority v7 source sealing requires read committed"
			if !errors.As(callErr, &postgresError) || postgresError.Code != "25001" || postgresError.Message != wantMessage {
				t.Fatalf("%s source-seal rejection=%v, want PostgreSQL 25001/%q", testCase.name, callErr, wantMessage)
			}
		})
	}
}

func task8AssertAuthorityV7OrdinaryDenial(t *testing.T, database *sql.DB) {
	t.Helper()
	t.Run("ordinary principal direct DML and truncate denied", func(t *testing.T) {
		if _, err := database.ExecContext(t.Context(), `CREATE ROLE task8_guard_ordinary NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS`); err != nil {
			t.Fatal(err)
		}
		defer func() {
			_, _ = database.ExecContext(t.Context(), `DROP ROLE task8_guard_ordinary`)
		}()
		before := task8GuardTableSnapshot(t, database)
		for _, statement := range []string{
			`SET LOCAL ROLE task8_guard_ordinary; INSERT INTO nodecontrol.node_pops(pop_code,iso_country,region,operator_state,created_at,updated_at) VALUES('forbidden','US','x','enabled',clock_timestamp(),clock_timestamp())`,
			`SET LOCAL ROLE task8_guard_ordinary; TRUNCATE nodecontrol.node_state_transitions`,
			`SET LOCAL ROLE task8_guard_ordinary; SELECT nodecontrol.v7_acquire_source_freeze_for_seal()`,
			`SET LOCAL ROLE task8_guard_ordinary; SELECT * FROM nodecontrol.v7_consume_down_guard(NULL::uuid,NULL::bytea,NULL::bytea,NULL::bytea,NULL::bytea,NULL::bytea)`,
			`SET LOCAL ROLE task8_guard_ordinary; SELECT * FROM nodecontrol.begin_staging_import(NULL::bytea,NULL::bytea,NULL::bytea,NULL::bytea,NULL::bytea,NULL::bytea,NULL::bytea,NULL::jsonb)`,
		} {
			tx, err := database.BeginTx(t.Context(), nil)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := tx.ExecContext(t.Context(), statement); err == nil {
				_ = tx.Rollback()
				t.Fatalf("ordinary principal statement succeeded: %s", statement)
			}
			_ = tx.Rollback()
		}
		if after := task8GuardTableSnapshot(t, database); !bytes.Equal(after, before) {
			t.Fatalf("ordinary-principal reject changed protected bytes\nbefore=%s\nafter=%s", before, after)
		}
	})

	t.Run("non-upgrade capabilities cannot invoke source sealer", func(t *testing.T) {
		for _, role := range []string{"nodecontrol_migration_downgrader", "nodecontrol_staging_importer"} {
			t.Run(role, func(t *testing.T) {
				tx, err := database.BeginTx(t.Context(), nil)
				if err != nil {
					t.Fatal(err)
				}
				defer tx.Rollback()
				if _, err := tx.ExecContext(t.Context(), "SET LOCAL ROLE "+role); err != nil {
					t.Fatal("set non-upgrade capability role:", err)
				}
				_, callErr := tx.ExecContext(t.Context(), `SELECT nodecontrol.v7_acquire_source_freeze_for_seal()`)
				var postgresError *pgconn.PgError
				if !errors.As(callErr, &postgresError) || postgresError.Code != "42501" ||
					postgresError.Message != "permission denied for function v7_acquire_source_freeze_for_seal" {
					t.Fatalf("%s source-sealer denial=%v, want exact PostgreSQL 42501 function denial", role, callErr)
				}
			})
		}
	})
}

type task8AuthorityProofGroup struct {
	table           string
	prefix          string
	operationColumn string
	kinds           []string
}

var task8AuthorityProofGroups = []task8AuthorityProofGroup{
	{"nodecontrol.node_enrollment_grants", "create_", "authority_operation_id", []string{"grant_create"}},
	{"nodecontrol.node_enrollment_grants", "claim_", "claim_authority_operation_id", []string{"grant_claim"}},
	{"nodecontrol.node_certificate_issuances", "activation_", "authority_operation_id", []string{"certificate_activate"}},
	{"nodecontrol.node_certificates", "revoke_", "revoke_authority_operation_id", []string{"certificate_revoke"}},
	{"nodecontrol.node_state_transitions", "", "authority_operation_id", []string{"identity_epoch_advance", "operator_transition"}},
	{"nodecontrol.node_security_incidents", "open_", "authority_operation_id", []string{"security_incident_open"}},
	{"nodecontrol.node_security_incidents", "resolve_", "resolution_authority_operation_id", []string{"security_incident_resolve"}},
	{"nodecontrol.node_resource_envelopes", "activation_", "authority_operation_id", []string{"resource_envelope_activate"}},
	{"nodecontrol.node_state_signing_intents", "activation_", "authority_operation_id", []string{"desired_activate", "recovery_activate"}},
	{"nodecontrol.node_root_metadata_publish_intents", "activation_", "authority_operation_id", []string{"root_publish", "metadata_publish"}},
}

var task8AuthorityProofSuffixes = []string{
	"authority_effect_commitment_jcs",
	"authority_effect_commitment_digest",
	"authority_provider_head_jcs",
	"authority_provider_head_digest",
	"authority_checkpoint_anchor_jcs",
	"authority_checkpoint_anchor_digest",
	"authority_effect_reason",
	"authority_attestation_expires_at",
	"authority_activation_deadline",
	"authority_expected_provider_identity_digest",
	"authority_activation_evidence_jcs",
	"authority_activation_evidence_digest",
	"authority_effect_resolution_jcs",
	"authority_effect_resolution_digest",
}

func task8AssertAuthorityProofGroupMatrix(t *testing.T, database *sql.DB) {
	t.Helper()
	if len(task8AuthorityProofGroups) != 10 {
		t.Fatalf("authority proof group count=%d, want literal 10", len(task8AuthorityProofGroups))
	}
	kindCount := 0
	owners := make(map[string]struct{})
	for _, group := range task8AuthorityProofGroups {
		kindCount += len(group.kinds)
		owners[group.table] = struct{}{}
		parts := strings.Split(group.table, ".")
		if len(parts) != 2 {
			t.Fatalf("invalid literal proof owner %q", group.table)
		}
		rows, err := database.QueryContext(t.Context(), `
SELECT column_name
FROM information_schema.columns
WHERE table_schema=$1 AND table_name=$2
  AND (column_name=$3 OR column_name=ANY($4::text[]))
ORDER BY column_name COLLATE "C"`, parts[0], parts[1], group.operationColumn, task8PrefixedProofColumns(group.prefix))
		if err != nil {
			t.Fatal("inspect authority proof owner columns:", err)
		}
		got := make([]string, 0, len(task8AuthorityProofSuffixes)+1)
		for rows.Next() {
			var column string
			if err := rows.Scan(&column); err != nil {
				rows.Close()
				t.Fatal(err)
			}
			got = append(got, column)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			t.Fatal(err)
		}
		rows.Close()
		want := append(task8PrefixedProofColumns(group.prefix), group.operationColumn)
		sort.Strings(want)
		if strings.Join(got, "\n") != strings.Join(want, "\n") {
			t.Errorf("proof owner %s/%s columns=%v, want exact operation + 14-field group %v", group.table, group.prefix, got, want)
		}
	}
	if kindCount != 13 || len(owners) != 8 {
		t.Fatalf("authority proof registry kinds/owners=%d/%d, want literal 13/8", kindCount, len(owners))
	}

	ownerNames := make([]string, 0, len(owners))
	for owner := range owners {
		ownerNames = append(ownerNames, strings.TrimPrefix(owner, "nodecontrol."))
	}
	sort.Strings(ownerNames)
	rows, err := database.QueryContext(t.Context(), `
SELECT c.relname,t.tgname,t.tgtype,t.tgenabled
FROM pg_catalog.pg_trigger t
JOIN pg_catalog.pg_class c ON c.oid=t.tgrelid
JOIN pg_catalog.pg_namespace n ON n.oid=c.relnamespace
WHERE n.nspname='nodecontrol' AND c.relname=ANY($1::text[])
  AND t.tgfoid='nodecontrol.v7_guard_authority_proof_transition()'::regprocedure::oid
  AND NOT t.tgisinternal
ORDER BY c.relname COLLATE "C"`, ownerNames)
	if err != nil {
		t.Fatal("inspect proof guard trigger coverage:", err)
	}
	triggerCount := 0
	ownerTriggerCount := make(map[string]int, len(ownerNames))
	for rows.Next() {
		var owner, triggerName, enabled string
		var triggerType int
		if err := rows.Scan(&owner, &triggerName, &triggerType, &enabled); err != nil {
			rows.Close()
			t.Fatal(err)
		}
		triggerCount++
		ownerTriggerCount[owner]++
		if want := "ncv7_" + owner + "_proof_guard"; triggerName != want {
			t.Errorf("proof guard trigger %s name=%q, want %q", owner, triggerName, want)
		}
		// PostgreSQL tgtype 31 is ROW|BEFORE|INSERT|DELETE|UPDATE. INSERT
		// prevents direct terminal installation; DELETE protects the OLD tuple.
		if triggerType != 31 {
			t.Errorf("proof guard trigger %s tgtype=%d, want exact BEFORE INSERT OR UPDATE OR DELETE row trigger (31)", owner, triggerType)
		}
		if enabled != "O" {
			t.Errorf("proof guard trigger %s tgenabled=%q, want O", owner, enabled)
		}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		t.Fatal(err)
	}
	rows.Close()
	if triggerCount != 8 {
		t.Errorf("proof guard trigger owner count=%d, want exact 8", triggerCount)
	}
	for _, owner := range ownerNames {
		if ownerTriggerCount[owner] != 1 {
			t.Errorf("proof guard trigger %s count=%d, want exact 1", owner, ownerTriggerCount[owner])
		}
	}

	digest := bytes.Repeat([]byte{0x41}, 32)
	rollbackTerminal := []any{
		[]byte(`{"schema_version":"1","stage":"final"}`), digest,
		[]byte(`{"provider_head":"literal"}`), bytes.Repeat([]byte{0x42}, 32),
		nil, nil,
		"none", time.Date(2026, time.August, 29, 17, 10, 0, 0, time.UTC),
		time.Date(2026, time.August, 29, 17, 9, 0, 0, time.UTC), bytes.Repeat([]byte{0x44}, 32),
		[]byte(`{"decision_capability":"may_apply"}`), bytes.Repeat([]byte{0x45}, 32),
		[]byte(`{"disposition":"applied"}`), bytes.Repeat([]byte{0x46}, 32),
	}
	prepared := make([]any, 14)
	prepared[0], prepared[1] = []byte(`{"schema_version":"1","stage":"prepared"}`), bytes.Repeat([]byte{0x31}, 32)
	finalNotApplied := append([]any(nil), rollbackTerminal...)
	finalNotApplied[6] = "failed"
	for _, index := range []int{7, 8, 9} {
		finalNotApplied[index] = nil
	}
	higherAuthority := append([]any(nil), finalNotApplied...)
	higherAuthority[4] = []byte(`{"checkpoint":"literal"}`)
	higherAuthority[5] = bytes.Repeat([]byte{0x43}, 32)
	higherAuthority[6] = "superseded"
	deadlineExpired := append([]any(nil), rollbackTerminal...)
	deadlineExpired[6] = "activation_deadline_expired"
	for _, shape := range []struct {
		name string
		args []any
	}{
		{"legacy all-null", make([]any, 14)},
		{"commitment-only prepared", prepared},
		{"final-not-applied terminal", finalNotApplied},
		{"higher-authority terminal", higherAuthority},
		{"rollback-resistant terminal", rollbackTerminal},
		{"deadline-expired rollback-resistant terminal", deadlineExpired},
	} {
		var valid bool
		if err := database.QueryRowContext(t.Context(), task8ProofGroupValiditySQL, shape.args...).Scan(&valid); err != nil {
			t.Fatalf("execute proof-group shape %s: %v", shape.name, err)
		}
		if !valid {
			t.Errorf("production proof-group validator rejected legal %s shape", shape.name)
		}
	}

	invalidFieldIndexes := map[int]struct{}{0: {}, 1: {}, 2: {}, 3: {}, 4: {}, 5: {}, 6: {}, 9: {}, 10: {}, 11: {}, 12: {}, 13: {}}
	for fieldIndex, suffix := range task8AuthorityProofSuffixes {
		if _, invalid := invalidFieldIndexes[fieldIndex]; !invalid {
			continue
		}
		mutated := append([]any(nil), rollbackTerminal...)
		switch fieldIndex {
		case 0, 2, 4, 10, 12:
			mutated[fieldIndex] = []byte{}
		case 1, 3, 5, 9, 11, 13:
			mutated[fieldIndex] = bytes.Repeat([]byte{byte(0x70 + fieldIndex)}, 31)
		case 6:
			mutated[fieldIndex] = ""
		}
		var valid bool
		if err := database.QueryRowContext(t.Context(), task8ProofGroupValiditySQL, mutated...).Scan(&valid); err != nil {
			t.Fatalf("execute proof-group field mutation %s: %v", suffix, err)
		}
		if valid {
			t.Errorf("production proof-group validator accepted structurally invalid field mutation %s", suffix)
		}
	}

	for fieldIndex, replacement := range map[int]time.Time{
		7: time.Date(2026, time.August, 29, 17, 20, 0, 0, time.UTC),
		8: time.Date(2026, time.August, 29, 17, 8, 0, 0, time.UTC),
	} {
		alternate := append([]any(nil), rollbackTerminal...)
		alternate[fieldIndex] = replacement
		var valid bool
		if err := database.QueryRowContext(t.Context(), task8ProofGroupValiditySQL, alternate...).Scan(&valid); err != nil {
			t.Fatalf("execute proof-group legal time alternate %s: %v", task8AuthorityProofSuffixes[fieldIndex], err)
		}
		if !valid {
			t.Errorf("production proof-group validator rejected legal time alternate %s", task8AuthorityProofSuffixes[fieldIndex])
		}
	}

	nullReasonTerminal := append([]any(nil), rollbackTerminal...)
	nullReasonTerminal[6] = nil
	var nullReasonValid sql.NullBool
	if err := database.QueryRowContext(t.Context(), task8ProofGroupValiditySQL, nullReasonTerminal...).Scan(&nullReasonValid); err != nil {
		t.Fatalf("execute proof-group NULL closed-reason mutation: %v", err)
	}
	if !nullReasonValid.Valid || nullReasonValid.Bool {
		t.Errorf("production proof-group validator NULL closed-reason result=%#v, want non-NULL false", nullReasonValid)
	}
}

func task8AssertSourceGuardTriggerMatrix(t *testing.T, database *sql.DB) {
	t.Helper()
	sourceTables := []string{
		"node_pops", "node_failure_domains", "node_capacity_profiles", "node_inventory",
		"node_failure_domain_membership", "node_endpoints", "node_process_slots",
		"node_resource_envelopes", "node_certificate_issuances", "node_enrollment_grants",
		"node_certificates", "node_security_incidents", "node_security_fault_receipts",
		"node_recovery_sessions", "node_restore_reauthorization_approvals",
		"node_state_signing_intents", "node_root_metadata_publish_intents",
		"node_root_metadata_signature_shares", "node_desired_states", "node_recovery_states",
		"node_observed_states", "node_operator_audit", "node_state_transitions",
		"control_plane_trust_bundle_high_waters",
	}
	want := make(map[string]int, len(sourceTables)*2+1)
	for _, table := range sourceTables {
		want[table+"|ncv7_00_source_lock"] = 30
		want[table+"|ncv7_01_source_guard"] = 31
	}
	want["control_plane_authority_fences|ncv7_00_fence_source_lock"] = 30

	rows, err := database.QueryContext(t.Context(), `
SELECT relation.relname,trigger.tgname,trigger.tgtype,trigger.tgenabled
FROM pg_catalog.pg_trigger AS trigger
JOIN pg_catalog.pg_class AS relation ON relation.oid=trigger.tgrelid
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid=relation.relnamespace
WHERE namespace.nspname='nodecontrol'
  AND trigger.tgfoid='nodecontrol.v7_assert_source_writable()'::regprocedure::oid
  AND NOT trigger.tgisinternal
ORDER BY relation.relname COLLATE "C",trigger.tgname COLLATE "C"`)
	if err != nil {
		t.Fatal("inspect source guard trigger topology:", err)
	}
	defer rows.Close()
	seen := make(map[string]int, len(want))
	for rows.Next() {
		var table, triggerName, enabled string
		var triggerType int
		if err := rows.Scan(&table, &triggerName, &triggerType, &enabled); err != nil {
			t.Fatal(err)
		}
		key := table + "|" + triggerName
		wantType, registered := want[key]
		if !registered {
			t.Errorf("unexpected source guard trigger %s tgtype=%d enabled=%q", key, triggerType, enabled)
			continue
		}
		seen[key]++
		if triggerType != wantType || enabled != "O" {
			t.Errorf("source guard trigger %s tgtype/enabled=%d/%q, want %d/O", key, triggerType, enabled, wantType)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	rows.Close()
	if len(seen) != len(want) {
		t.Errorf("source guard trigger topology count=%d, want exact %d", len(seen), len(want))
	}
	for key := range want {
		if seen[key] != 1 {
			t.Errorf("source guard trigger %s count=%d, want exact 1", key, seen[key])
		}
	}

	type fenceTriggerExpectation struct {
		triggerType                      int
		functionSchema, functionName     string
		constraint, deferrable, deferred bool
	}
	wantFence := map[string]fenceTriggerExpectation{
		"ncv7_00_fence_source_lock": {
			triggerType: 30, functionSchema: "nodecontrol", functionName: "v7_assert_source_writable",
		},
		"control_plane_authority_fences_enforce_update": {
			triggerType: 31, functionSchema: "nodecontrol", functionName: "enforce_authority_fence_update",
		},
		"ncv7_fence_owner_closure": {
			triggerType: 17, functionSchema: "nodecontrol", functionName: "v7_assert_activation_barrier",
			constraint: true, deferrable: true, deferred: true,
		},
	}
	fenceRows, err := database.QueryContext(t.Context(), `
SELECT trigger.tgname,trigger.tgtype,trigger.tgenabled,trigger.tgconstraint<>0,
       trigger.tgdeferrable,trigger.tginitdeferred,
       procedure_namespace.nspname,procedure.proname,pg_catalog.pg_get_function_identity_arguments(procedure.oid)
FROM pg_catalog.pg_trigger AS trigger
JOIN pg_catalog.pg_class AS relation ON relation.oid=trigger.tgrelid
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid=relation.relnamespace
JOIN pg_catalog.pg_proc AS procedure ON procedure.oid=trigger.tgfoid
JOIN pg_catalog.pg_namespace AS procedure_namespace ON procedure_namespace.oid=procedure.pronamespace
WHERE namespace.nspname='nodecontrol'
  AND relation.relname='control_plane_authority_fences'
  AND NOT trigger.tgisinternal
ORDER BY trigger.tgname COLLATE "C"`)
	if err != nil {
		t.Fatal("inspect complete fence trigger topology:", err)
	}
	defer fenceRows.Close()
	seenFence := make(map[string]int, len(wantFence))
	for fenceRows.Next() {
		var name, enabled, functionSchema, functionName, functionArguments string
		var triggerType int
		var constraint, deferrable, deferred bool
		if err := fenceRows.Scan(&name, &triggerType, &enabled, &constraint, &deferrable, &deferred,
			&functionSchema, &functionName, &functionArguments); err != nil {
			t.Fatal(err)
		}
		expected, registered := wantFence[name]
		if !registered {
			t.Errorf("unexpected authority fence trigger %s", name)
			continue
		}
		seenFence[name]++
		if triggerType != expected.triggerType || enabled != "O" || functionSchema != expected.functionSchema ||
			functionName != expected.functionName || functionArguments != "" ||
			constraint != expected.constraint || deferrable != expected.deferrable || deferred != expected.deferred {
			t.Errorf("authority fence trigger %s metadata=%d/%q/%s.%s(%s)/%t/%t/%t, want %d/O/%s.%s()/%t/%t/%t",
				name, triggerType, enabled, functionSchema, functionName, functionArguments, constraint, deferrable, deferred,
				expected.triggerType, expected.functionSchema, expected.functionName, expected.constraint, expected.deferrable, expected.deferred)
		}
	}
	if err := fenceRows.Err(); err != nil {
		t.Fatal(err)
	}
	if len(seenFence) != len(wantFence) {
		t.Errorf("authority fence trigger topology count=%d, want exact %d", len(seenFence), len(wantFence))
	}
	for name := range wantFence {
		if seenFence[name] != 1 {
			t.Errorf("authority fence trigger %s count=%d, want exact 1", name, seenFence[name])
		}
	}
}

func task8AssertRealAuthorityProofDMLMatrix(t *testing.T, database *sql.DB, now time.Time) {
	t.Helper()
	tx, err := database.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatal("begin real proof-owner DML matrix:", err)
	}
	defer tx.Rollback()
	caseIndex := 0
	fieldMutationGroupCount := 0
	for _, group := range task8AuthorityProofGroups {
		for _, effectKind := range group.kinds {
			group, effectKind := group, effectKind
			t.Run("real proof lifecycle "+effectKind, func(t *testing.T) {
				savepoint := "task8_proof_" + fmt.Sprintf("%02d", caseIndex)
				caseIndex++
				if _, err := tx.ExecContext(t.Context(), "SAVEPOINT "+savepoint); err != nil {
					t.Fatal(err)
				}
				defer func() {
					_, _ = tx.ExecContext(context.Background(), "ROLLBACK TO SAVEPOINT "+savepoint)
					_, _ = tx.ExecContext(context.Background(), "RELEASE SAVEPOINT "+savepoint)
				}()

				activationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("task8-proof-activation:"+effectKind))
				task8SeedProofActivation(t, tx, activationID, caseIndex, now)
				operationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("task8-proof-operation:"+effectKind))
				task8InsertClaimFence(t, tx, operationID, activationID, int64(800+caseIndex), effectKind, now)
				preparedCTID := task8SeedPreparedProofOwner(t, tx, group, effectKind, operationID, int64(800+caseIndex), now)
				task8ExpectProofGuardReject(t, tx, "reserved prepared owner DELETE "+effectKind, fmt.Sprintf(
					"DELETE FROM %s WHERE ctid=$1::tid", group.table), preparedCTID)
				task8ExpectTerminalProofBeforeFenceReject(t, tx, group, effectKind, preparedCTID)
				task8ExpectProofGuardReject(t, tx, "combined fence Bind and Finalize "+effectKind, `
UPDATE nodecontrol.control_plane_authority_fences
SET effect_digest=decode(repeat('81',32),'hex'),db_system_id=1,db_timeline=1,required_lsn='0/1',effect_bound_at=$2,
 provider_status='committed',provider_receipt_digest=decode(repeat('82',32),'hex'),visibility_state='active',terminal_at=$2
WHERE operation_id=$1`, operationID, now)
				task8ExpectAuthorityGuardReject(t, tx,
					"Bind commitment digest mismatch "+effectKind,
					"does not match owner commitment", `
UPDATE nodecontrol.control_plane_authority_fences
SET effect_digest=decode(repeat('8f',32),'hex'),
    db_system_id=1,db_timeline=1,required_lsn='0/1',effect_bound_at=$2
WHERE operation_id=$1
  AND provider_status='reserved'
  AND effect_digest IS NULL`, operationID, now)
				task8BindProofFence(t, tx, operationID, now)
				task8FinalizeProofFence(t, tx, operationID, now)
				task8AssertPreparedProofOwnerFrozen(t, tx, group, effectKind, preparedCTID, now)
				task8AssertPreparedCommitmentCannotBeReplaced(t, tx, group, effectKind, preparedCTID)
				task8ExpectProofGuardReject(t, tx, "committed prepared owner DELETE "+effectKind, fmt.Sprintf(
					"DELETE FROM %s WHERE ctid=$1::tid", group.table), preparedCTID)

				// A prepared commitment has no Head. Changing a domain disposition
				// must not synthesize Head/evidence from the fence or provider state.
				if group.table == "nodecontrol.node_state_transitions" {
					mutatedDisposition := "not_applied"
					if effectKind == "identity_epoch_advance" {
						mutatedDisposition = "applied"
					}
					task8ExpectProofGuardReject(t, tx, "no-Head synthesis", fmt.Sprintf(
						"UPDATE %s SET authority_effect_disposition=$2 WHERE ctid=$1::tid", group.table), preparedCTID, mutatedDisposition)
				}

				if effectKind == group.kinds[0] {
					fieldMutationGroupCount++
					task8AssertTerminalProofFieldMutations(t, tx, group, effectKind, preparedCTID, now)
				}

				terminalCTID := task8InstallTerminalProof(t, tx, group, effectKind, preparedCTID, task8CanonicalTerminalShape(effectKind), now)
				secondReason := "failed"
				if task8CanonicalTerminalShape(effectKind) == "final_not_applied" {
					secondReason = "validation_rejected"
				}
				task8ExpectProofGuardReject(t, tx, "second terminal", fmt.Sprintf(
					"UPDATE %s SET %sauthority_effect_reason=$2 WHERE ctid=$1::tid", group.table, group.prefix), terminalCTID, secondReason)
			})
		}
	}
	if caseIndex != 13 {
		t.Fatalf("executed real proof lifecycle count=%d, want exact 13", caseIndex)
	}
	if fieldMutationGroupCount != 10 {
		t.Fatalf("executed fourteen-field terminal mutation group count=%d, want exact 10", fieldMutationGroupCount)
	}
	task8AssertActivationBarrierMutations(t, tx, now)
	task8AssertProofOwnershipNegatives(t, tx, now)
	if err := tx.Rollback(); err != nil {
		t.Fatal("rollback real proof-owner DML matrix:", err)
	}
}

func task8AssertDeferredProofCommitAndTerminalDelete(t *testing.T, database *sql.DB) {
	t.Helper()
	const effectKind = "operator_transition"
	group := task8AuthorityProofGroups[4]
	negativeActivationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("task8-deferred-missing-terminal-activation"))
	negativeOperationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("task8-deferred-missing-terminal-operation"))
	positiveActivationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("task8-deferred-terminal-activation"))
	positiveOperationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("task8-deferred-terminal-operation"))
	cleanupDone := false
	defer func() {
		if !cleanupDone {
			task8CleanupCommittedProofFixtures(t, database,
				[]uuid.UUID{negativeOperationID, positiveOperationID},
				[]uuid.UUID{negativeActivationID, positiveActivationID}, effectKind)
		}
	}()

	before := task8GuardTableSnapshot(t, database)
	var fixtureNow time.Time
	if err := database.QueryRowContext(t.Context(), `SELECT clock_timestamp()-interval '182 days'`).Scan(&fixtureNow); err != nil {
		t.Fatal("load expired-retention fixture time:", err)
	}

	negativeTx, err := database.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatal("begin deferred negative proof transaction:", err)
	}
	negativeOpen := true
	defer func() {
		if negativeOpen {
			_ = negativeTx.Rollback()
		}
	}()
	task8SeedProofActivation(t, negativeTx, negativeActivationID, 901, fixtureNow)
	task8InsertClaimFence(t, negativeTx, negativeOperationID, negativeActivationID, 1901, effectKind, fixtureNow)
	task8SeedPreparedProofOwner(t, negativeTx, group, effectKind, negativeOperationID, 1901, fixtureNow)
	task8BindProofFence(t, negativeTx, negativeOperationID, fixtureNow)
	task8FinalizeProofFence(t, negativeTx, negativeOperationID, fixtureNow)
	commitErr := negativeTx.Commit()
	negativeOpen = false
	if commitErr == nil {
		t.Fatal("COMMIT accepted committed fence without terminal proof owner")
	}
	var postgresError *pgconn.PgError
	if !errors.As(commitErr, &postgresError) || postgresError.Code != "23514" ||
		!strings.Contains(strings.ToLower(postgresError.Message), "terminal proof owner") {
		t.Fatalf("deferred COMMIT seam=%#v, want 23514 terminal proof owner", postgresError)
	}
	if after := task8GuardTableSnapshot(t, database); !bytes.Equal(after, before) {
		t.Fatalf("rejected deferred COMMIT changed protected bytes\nbefore=%s\nafter=%s", before, after)
	}

	positiveTx, err := database.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatal("begin deferred positive proof transaction:", err)
	}
	positiveOpen := true
	defer func() {
		if positiveOpen {
			_ = positiveTx.Rollback()
		}
	}()
	task8SeedProofActivation(t, positiveTx, positiveActivationID, 902, fixtureNow)
	task8InsertClaimFence(t, positiveTx, positiveOperationID, positiveActivationID, 1902, effectKind, fixtureNow)
	preparedCTID := task8SeedPreparedProofOwner(t, positiveTx, group, effectKind, positiveOperationID, 1902, fixtureNow)
	task8BindProofFence(t, positiveTx, positiveOperationID, fixtureNow)
	task8FinalizeProofFence(t, positiveTx, positiveOperationID, fixtureNow)
	task8InstallTerminalProof(t, positiveTx, group, effectKind, preparedCTID, "rollback_resistant", fixtureNow)
	if err := positiveTx.Commit(); err != nil {
		positiveOpen = false
		t.Fatal("exact terminal proof COMMIT:", err)
	}
	positiveOpen = false

	var exactTerminal bool
	if err := database.QueryRowContext(t.Context(), `
SELECT EXISTS (
 SELECT 1
 FROM nodecontrol.control_plane_authority_fences AS fence
 JOIN nodecontrol.node_state_transitions AS owner
   ON owner.authority_operation_id=fence.operation_id
 WHERE fence.operation_id=$1
   AND fence.provider_status='committed'
   AND fence.visibility_state='active'
   AND fence.terminal_at IS NOT NULL
   AND owner.authority_provider_head_digest IS NOT NULL
   AND owner.authority_effect_commitment_digest=fence.effect_digest
)`, positiveOperationID).Scan(&exactTerminal); err != nil || !exactTerminal {
		t.Fatalf("committed exact terminal closure=%t error=%v", exactTerminal, err)
	}

	deleteTx, err := database.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatal("begin terminal-retention DELETE transaction:", err)
	}
	deleteOpen := true
	defer func() {
		if deleteOpen {
			_ = deleteTx.Rollback()
		}
	}()
	var deletedID uuid.UUID
	if err := deleteTx.QueryRowContext(t.Context(), `
DELETE FROM nodecontrol.node_state_transitions
WHERE authority_operation_id=$1
  AND retention_until <= clock_timestamp()
RETURNING transition_id`, positiveOperationID).Scan(&deletedID); err != nil {
		t.Fatal("terminal-retention DELETE:", err)
	}
	if deletedID != task8ProofFixtureUUID(effectKind, "owner") {
		t.Fatalf("deleted transition=%s, want exact terminal owner", deletedID)
	}
	if err := deleteTx.Commit(); err != nil {
		deleteOpen = false
		t.Fatal("terminal-retention DELETE COMMIT:", err)
	}
	deleteOpen = false

	var ownerCount int
	var fenceStillTerminal bool
	if err := database.QueryRowContext(t.Context(), `
SELECT
 (SELECT count(*) FROM nodecontrol.node_state_transitions WHERE authority_operation_id=$1),
 EXISTS (
   SELECT 1 FROM nodecontrol.control_plane_authority_fences
   WHERE operation_id=$1
     AND provider_status='committed'
     AND visibility_state='active'
     AND terminal_at IS NOT NULL
 )`, positiveOperationID).Scan(&ownerCount, &fenceStillTerminal); err != nil {
		t.Fatal("verify terminal-retention DELETE:", err)
	}
	if ownerCount != 0 || !fenceStillTerminal {
		t.Fatalf("terminal DELETE owner/fence=%d/%t, want 0/true", ownerCount, fenceStillTerminal)
	}

	task8CleanupCommittedProofFixtures(t, database,
		[]uuid.UUID{negativeOperationID, positiveOperationID},
		[]uuid.UUID{negativeActivationID, positiveActivationID}, effectKind)
	cleanupDone = true
	if after := task8GuardTableSnapshot(t, database); !bytes.Equal(after, before) {
		t.Fatalf("deferred/DELETE fixture cleanup mismatch\nbefore=%s\nafter=%s", before, after)
	}
}

func task8AssertNewLegacyFenceRejected(t *testing.T, database *sql.DB, now time.Time, nodeID uuid.UUID) {
	t.Helper()
	tx, err := database.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatal("begin new legacy fence rejection transaction:", err)
	}
	defer tx.Rollback()
	operationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("task8-new-legacy-fence"))
	task8ExpectAuthorityGuardReject(t, tx, "new exact legacy_v6 fence", "new legacy_v6 authority fences", `
INSERT INTO nodecontrol.control_plane_authority_fences(
 operation_id,effect_kind,scope_kind,authority_epoch,authority_sequence,scope_digest,provider_reservation_digest,
 provider_status,visibility_state,reserved_at,authority_protocol_profile)
VALUES($1,'operator_transition','node',7401,1,$2,decode(repeat('42',32),'hex'),
 'reserved','fence_pending',$3,'legacy_v6')`, operationID, task8NodeScopeDigest(nodeID), now)
	if err := tx.Rollback(); err != nil {
		t.Fatal("rollback new legacy fence rejection transaction:", err)
	}
}

// The sealer-first case proves statement DML/marker lock ordering. The
// writer-first case additionally drives the real trusted application wrapper,
// observes its generated acquire call waiting, and proves its scan runs only
// after the conflicting writer commits.
func task8AssertSourceFreezeAdvisoryOrdering(t *testing.T, database *sql.DB, databaseURL string, now time.Time) {
	t.Helper()
	const sourceSealID = "11111111-1111-4111-8111-000000000014"
	sealSeed := task8LiteralDownRelationSeeds[13]
	if sealSeed.table != "nodecontrol.control_plane_authority_legacy_source_seals" ||
		!strings.Contains(sealSeed.insertSQL, "VALUES ('"+sourceSealID+"'::uuid,") {
		t.Fatalf("source-freeze seal seed=%q does not contain exact fixture identity %s", sealSeed.table, sourceSealID)
	}

	t.Run("committed seal wakes and rejects blocked writer", func(t *testing.T) {
		const popCode = "source-race-sealer-first"
		beforeAll := task8GuardTableSnapshot(t, database)
		seedCtx, cancelSeed := context.WithTimeout(t.Context(), 15*time.Second)
		_, err := database.ExecContext(seedCtx, `
INSERT INTO nodecontrol.node_pops(pop_code,iso_country,region,operator_state,created_at,updated_at)
VALUES($1,'US','source-race-base','enabled',$2,$2)`, popCode, now)
		cancelSeed()
		if err != nil {
			t.Fatal("seed sealer-first source row:", err)
		}
		cleaned := false
		defer func() {
			if !cleaned {
				if err := task8CleanupSourceFreezeAdvisoryFixture(database, popCode, sourceSealID); err != nil {
					t.Errorf("fallback sealer-first source-race cleanup: %v", err)
				}
			}
		}()
		beforeRow := task8SourceFreezePOPBytes(t, database, popCode)

		sealerConn, err := database.Conn(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		defer sealerConn.Close()
		sealerTx, err := sealerConn.BeginTx(t.Context(), nil)
		if err != nil {
			t.Fatal(err)
		}
		sealerOpen := true
		defer func() {
			if sealerOpen {
				_ = sealerTx.Rollback()
			}
		}()
		var sealerPID int
		var replicationRole string
		if err := sealerTx.QueryRowContext(t.Context(), `SELECT pg_backend_pid(),current_setting('session_replication_role')`).Scan(&sealerPID, &replicationRole); err != nil {
			t.Fatal("inspect source sealer backend:", err)
		}
		if replicationRole != "origin" {
			t.Fatalf("source sealer replication role=%q, want origin", replicationRole)
		}
		var sealCTID string
		sealInsertCtx, cancelSealInsert := context.WithTimeout(t.Context(), 15*time.Second)
		err = sealerTx.QueryRowContext(sealInsertCtx, sealSeed.insertSQL).Scan(&sealCTID)
		cancelSealInsert()
		if err != nil || sealCTID == "" {
			t.Fatalf("insert origin-mode source seal ctid/error=%q/%v", sealCTID, err)
		}

		writerConn, err := database.Conn(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		defer writerConn.Close()
		writerTx, err := writerConn.BeginTx(t.Context(), nil)
		if err != nil {
			t.Fatal(err)
		}
		writerOpen := true
		defer func() {
			if writerOpen {
				_ = writerTx.Rollback()
			}
		}()
		var writerPID int
		if err := writerTx.QueryRowContext(t.Context(), `SELECT pg_backend_pid()`).Scan(&writerPID); err != nil {
			t.Fatal(err)
		}
		opCtx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
		defer cancel()
		writerResult := make(chan error, 1)
		go func() {
			_, writerErr := writerTx.ExecContext(opCtx, `
UPDATE nodecontrol.node_pops
SET region='source-race-rejected',version=version+1,updated_at=$2
WHERE pop_code=$1`, popCode, now.Add(time.Second))
			writerResult <- writerErr
		}()
		task8AwaitExactAdvisoryWait(t, database, writerResult, sealerPID, writerPID, "source writer behind source seal")

		if err := sealerTx.Commit(); err != nil {
			t.Fatal("commit source seal before blocked writer:", err)
		}
		sealerOpen = false
		var writerErr error
		select {
		case writerErr = <-writerResult:
		case <-time.After(5 * time.Second):
			t.Fatal("source writer did not wake after seal commit")
		}
		var postgresError *pgconn.PgError
		if !errors.As(writerErr, &postgresError) || postgresError.Code != "55000" ||
			postgresError.Message != "nodecontrol legacy source is frozen" {
			t.Fatalf("source writer rejection=%v, want exact PostgreSQL 55000 source freeze", writerErr)
		}
		if err := writerTx.Rollback(); err != nil {
			t.Fatal("rollback source writer after freeze rejection:", err)
		}
		writerOpen = false
		if afterRow := task8SourceFreezePOPBytes(t, database, popCode); !bytes.Equal(afterRow, beforeRow) {
			t.Fatalf("blocked source writer changed row bytes\nbefore=%s\nafter=%s", beforeRow, afterRow)
		}

		if err := task8CleanupSourceFreezeAdvisoryFixture(database, popCode, sourceSealID); err != nil {
			t.Fatal("clean sealer-first source-race fixture:", err)
		}
		cleaned = true
		if afterAll := task8GuardTableSnapshot(t, database); !bytes.Equal(afterAll, beforeAll) {
			t.Fatalf("sealer-first source-race cleanup mismatch\nbefore=%s\nafter=%s", beforeAll, afterAll)
		}
	})

	t.Run("trusted sealer waits then scans committed writer and rejects later writer", func(t *testing.T) {
		const popCode = "source-race-writer-first"
		const sealerApplicationName = "task8-authority-v7-source-sealer"
		beforeAll := task8GuardTableSnapshot(t, database)
		seedCtx, cancelSeed := context.WithTimeout(t.Context(), 15*time.Second)
		_, err := database.ExecContext(seedCtx, `
INSERT INTO nodecontrol.node_pops(pop_code,iso_country,region,operator_state,created_at,updated_at)
VALUES($1,'US','source-race-base','enabled',$2,$2)`, popCode, now)
		cancelSeed()
		if err != nil {
			t.Fatal("seed writer-first source row:", err)
		}
		cleaned := false
		defer func() {
			if !cleaned {
				if err := task8CleanupSourceFreezeAdvisoryFixture(database, popCode, sourceSealID); err != nil {
					t.Errorf("fallback writer-first source-race cleanup: %v", err)
				}
			}
		}()

		writerConn, err := database.Conn(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		defer writerConn.Close()
		writerTx, err := writerConn.BeginTx(t.Context(), nil)
		if err != nil {
			t.Fatal(err)
		}
		writerOpen := true
		defer func() {
			if writerOpen {
				_ = writerTx.Rollback()
			}
		}()
		var writerPID int
		if err := writerTx.QueryRowContext(t.Context(), `SELECT pg_backend_pid()`).Scan(&writerPID); err != nil {
			t.Fatal(err)
		}
		writerMutationCtx, cancelWriterMutation := context.WithTimeout(t.Context(), 15*time.Second)
		result, err := writerTx.ExecContext(writerMutationCtx, `
UPDATE nodecontrol.node_pops
SET region='source-race-committed',version=version+1,updated_at=$2
WHERE pop_code=$1`, popCode, now.Add(time.Second))
		cancelWriterMutation()
		if err != nil {
			t.Fatal("execute writer-first source mutation:", err)
		}
		if rows, err := result.RowsAffected(); err != nil || rows != 1 {
			t.Fatalf("writer-first source mutation rows/error=%d/%v, want 1/nil", rows, err)
		}

		sealerConfig, err := pgxpool.ParseConfig(databaseURL)
		if err != nil {
			t.Fatal("parse trusted source-sealer database URL:", err)
		}
		sealerConfig.MaxConns = 1
		sealerConfig.ConnConfig.RuntimeParams["application_name"] = sealerApplicationName
		sealerPool, err := pgxpool.NewWithConfig(t.Context(), sealerConfig)
		if err != nil {
			t.Fatal("open trusted source-sealer pool:", err)
		}
		defer sealerPool.Close()
		sourceSealer, err := authority.NewAuthorityV7SourceSealer(sealerPool)
		if err != nil {
			t.Fatal("construct trusted source sealer:", err)
		}
		opCtx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
		defer cancel()
		type sourceObservation struct {
			pid             int
			replicationRole string
			region          string
			version         int64
		}
		observed := make(chan sourceObservation, 1)
		releaseSeal := make(chan struct{})
		sealReleased := false
		defer func() {
			if !sealReleased {
				close(releaseSeal)
			}
		}()
		sealResult := make(chan error, 1)
		var sealCTID string
		go func() {
			sealResult <- sourceSealer.Run(opCtx, func(ctx context.Context, dbtx store.DBTX, _ *store.Queries) error {
				var observation sourceObservation
				if err := dbtx.QueryRow(ctx, `
SELECT pg_backend_pid(),current_setting('session_replication_role'),region,version
FROM nodecontrol.node_pops WHERE pop_code=$1`, popCode).Scan(
					&observation.pid, &observation.replicationRole, &observation.region, &observation.version,
				); err != nil {
					return fmt.Errorf("scan writer-first source projection: %w", err)
				}
				observed <- observation
				select {
				case <-releaseSeal:
				case <-ctx.Done():
					return ctx.Err()
				}
				if err := dbtx.QueryRow(ctx, sealSeed.insertSQL).Scan(&sealCTID); err != nil {
					return fmt.Errorf("insert writer-first source seal: %w", err)
				}
				return nil
			})
		}()
		waitingSealerPID := task8AwaitNamedAdvisoryWait(
			t, database, sealResult, writerPID, sealerApplicationName, "trusted source sealer behind source writer",
		)
		select {
		case observation := <-observed:
			t.Fatalf("trusted source-seal callback ran before acquire completed: %+v", observation)
		default:
		}

		if err := writerTx.Commit(); err != nil {
			t.Fatal("commit writer before waiting source seal:", err)
		}
		writerOpen = false
		var observation sourceObservation
		select {
		case observation = <-observed:
		case <-time.After(5 * time.Second):
			t.Fatal("trusted source sealer did not scan after writer commit")
		}
		if observation.pid != waitingSealerPID || observation.replicationRole != "origin" ||
			observation.region != "source-race-committed" || observation.version != 2 {
			t.Fatalf("trusted source-seal observation=%+v waiting_pid=%d, want same pid/origin/source-race-committed/2", observation, waitingSealerPID)
		}

		lateWriterConn, err := database.Conn(t.Context())
		if err != nil {
			t.Fatal("open late source writer connection:", err)
		}
		defer lateWriterConn.Close()
		lateWriterTx, err := lateWriterConn.BeginTx(t.Context(), nil)
		if err != nil {
			t.Fatal("begin late source writer transaction:", err)
		}
		lateWriterOpen := true
		defer func() {
			if lateWriterOpen {
				_ = lateWriterTx.Rollback()
			}
		}()
		var lateWriterPID int
		if err := lateWriterTx.QueryRowContext(t.Context(), `SELECT pg_backend_pid()`).Scan(&lateWriterPID); err != nil {
			t.Fatal("inspect late source writer backend:", err)
		}
		lateWriterResult := make(chan error, 1)
		go func() {
			_, writerErr := lateWriterTx.ExecContext(opCtx, `
UPDATE nodecontrol.node_pops
SET region='source-race-late',version=version+1,updated_at=$2
WHERE pop_code=$1`, popCode, now.Add(2*time.Second))
			lateWriterResult <- writerErr
		}()
		task8AwaitExactAdvisoryWait(t, database, lateWriterResult, observation.pid, lateWriterPID, "late source writer behind trusted sealer")

		close(releaseSeal)
		sealReleased = true
		select {
		case err := <-sealResult:
			if err != nil || sealCTID == "" {
				t.Fatalf("trusted source seal after writer commit ctid/error=%q/%v", sealCTID, err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("trusted source sealer did not commit after release")
		}
		var lateWriterErr error
		select {
		case lateWriterErr = <-lateWriterResult:
		case <-time.After(5 * time.Second):
			t.Fatal("late source writer did not wake after trusted seal commit")
		}
		var postgresError *pgconn.PgError
		if !errors.As(lateWriterErr, &postgresError) || postgresError.Code != "55000" ||
			postgresError.Message != "nodecontrol legacy source is frozen" {
			t.Fatalf("late source writer rejection=%v, want exact PostgreSQL 55000 source freeze", lateWriterErr)
		}
		if err := lateWriterTx.Rollback(); err != nil {
			t.Fatal("rollback late source writer after freeze rejection:", err)
		}
		lateWriterOpen = false

		var frozen bool
		var region string
		var version int64
		if err := database.QueryRowContext(t.Context(), `
SELECT nodecontrol.v7_source_is_frozen(),region,version
FROM nodecontrol.node_pops WHERE pop_code=$1`, popCode).Scan(&frozen, &region, &version); err != nil {
			t.Fatal("verify writer-first source/seal state:", err)
		}
		if !frozen || region != "source-race-committed" || version != 2 {
			t.Fatalf("writer-first frozen/region/version=%t/%q/%d, want true/source-race-committed/2", frozen, region, version)
		}

		if err := task8CleanupSourceFreezeAdvisoryFixture(database, popCode, sourceSealID); err != nil {
			t.Fatal("clean writer-first source-race fixture:", err)
		}
		cleaned = true
		if afterAll := task8GuardTableSnapshot(t, database); !bytes.Equal(afterAll, beforeAll) {
			t.Fatalf("writer-first source-race cleanup mismatch\nbefore=%s\nafter=%s", beforeAll, afterAll)
		}
	})
}

func task8AwaitExactAdvisoryWait(t *testing.T, database *sql.DB, result <-chan error, blockerPID, waiterPID int, label string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case err := <-result:
			t.Fatalf("%s returned before advisory wait: %v", label, err)
		default:
		}
		var advisoryWait, exactBlocker bool
		observeCtx, cancelObserve := context.WithDeadline(t.Context(), deadline)
		err := database.QueryRowContext(observeCtx, `
SELECT coalesce(activity.wait_event_type='Lock' AND activity.wait_event='advisory',false),
       pg_catalog.pg_blocking_pids($2)=ARRAY[$1]::integer[]
FROM pg_catalog.pg_stat_activity AS activity
WHERE activity.pid=$2`, blockerPID, waiterPID).Scan(&advisoryWait, &exactBlocker)
		cancelObserve()
		if err != nil {
			t.Fatal("observe "+label+":", err)
		}
		if advisoryWait && exactBlocker {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("%s did not expose exact advisory blocker within five seconds", label)
}

func task8AwaitNamedAdvisoryWait(
	t *testing.T,
	database *sql.DB,
	result <-chan error,
	blockerPID int,
	applicationName string,
	label string,
) int {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case err := <-result:
			t.Fatalf("%s returned before advisory wait: %v", label, err)
		default:
		}
		var waiterPID int
		var advisoryWait, exactBlocker bool
		observeCtx, cancelObserve := context.WithDeadline(t.Context(), deadline)
		err := database.QueryRowContext(observeCtx, `
SELECT activity.pid,
       coalesce(activity.wait_event_type='Lock' AND activity.wait_event='advisory',false),
       pg_catalog.pg_blocking_pids(activity.pid)=ARRAY[$1]::integer[]
FROM pg_catalog.pg_stat_activity AS activity
WHERE activity.application_name=$2
  AND position('v7_acquire_source_freeze_for_seal' in lower(activity.query))>0
ORDER BY activity.pid
LIMIT 1`, blockerPID, applicationName).Scan(&waiterPID, &advisoryWait, &exactBlocker)
		cancelObserve()
		if err == nil && advisoryWait && exactBlocker {
			return waiterPID
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			t.Fatal("observe "+label+":", err)
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("%s did not expose exact advisory blocker within five seconds", label)
	return 0
}

func task8SourceFreezePOPBytes(t *testing.T, database *sql.DB, popCode string) []byte {
	t.Helper()
	var snapshot []byte
	if err := database.QueryRowContext(t.Context(), `
SELECT convert_to(to_jsonb(pop)::text,'UTF8')
FROM nodecontrol.node_pops AS pop WHERE pop.pop_code=$1`, popCode).Scan(&snapshot); err != nil {
		t.Fatal("snapshot source-race POP:", err)
	}
	return snapshot
}

func task8CleanupSourceFreezeAdvisoryFixture(database *sql.DB, popCode, sourceSealID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin source-race cleanup: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `SET LOCAL session_replication_role=replica`); err != nil {
		return fmt.Errorf("enable replica mode for source-race cleanup: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
DELETE FROM nodecontrol.control_plane_authority_legacy_source_seals
WHERE source_seal_id=$1::uuid`, sourceSealID); err != nil {
		return fmt.Errorf("delete source-race seal: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM nodecontrol.node_pops WHERE pop_code=$1`, popCode); err != nil {
		return fmt.Errorf("delete source-race POP: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit source-race cleanup: %w", err)
	}
	var sealCount, popCount int
	var frozen bool
	if err := database.QueryRowContext(ctx, `
SELECT
 (SELECT count(*) FROM nodecontrol.control_plane_authority_legacy_source_seals WHERE source_seal_id=$1::uuid),
 (SELECT count(*) FROM nodecontrol.node_pops WHERE pop_code=$2),
 nodecontrol.v7_source_is_frozen()`, sourceSealID, popCode).Scan(&sealCount, &popCount, &frozen); err != nil {
		return fmt.Errorf("verify source-race cleanup: %w", err)
	}
	if sealCount != 0 || popCount != 0 || frozen {
		return fmt.Errorf("source-race cleanup seal/pop/frozen=%d/%d/%t, want 0/0/false", sealCount, popCount, frozen)
	}
	return nil
}

func task8AssertActivationBarrierMutations(t *testing.T, tx *sql.Tx, now time.Time) {
	t.Helper()
	const effectKind = "operator_transition"
	mutations := []struct {
		name      string
		statement string
		args      func(uuid.UUID) []any
	}{
		{
			name:      "activation attempt digest",
			statement: `UPDATE nodecontrol.control_plane_authority_protocol_activations SET attempt_digest=decode(repeat('f1',32),'hex') WHERE activation_id=$1`,
			args:      func(activationID uuid.UUID) []any { return []any{activationID} },
		},
		{
			name:      "completion activation digest",
			statement: `UPDATE nodecontrol.control_plane_authority_protocol_activation_completions SET activation_digest=decode(repeat('f2',32),'hex') WHERE activation_id=$1`,
			args:      func(activationID uuid.UUID) []any { return []any{activationID} },
		},
		{
			name:      "release completion digest",
			statement: `UPDATE nodecontrol.control_plane_authority_protocol_activation_releases SET completion_digest=decode(repeat('f3',32),'hex') WHERE activation_id=$1`,
			args:      func(activationID uuid.UUID) []any { return []any{activationID} },
		},
		{
			name:      "completion runtime binding",
			statement: `UPDATE nodecontrol.control_plane_authority_protocol_activation_completions SET current_runtime_instance_binding_digest=decode(repeat('f4',32),'hex') WHERE activation_id=$1`,
			args:      func(activationID uuid.UUID) []any { return []any{activationID} },
		},
		{
			name:      "release before completion",
			statement: `UPDATE nodecontrol.control_plane_authority_protocol_activation_releases SET released_at=$2 WHERE activation_id=$1`,
			args:      func(activationID uuid.UUID) []any { return []any{activationID, now.Add(-time.Second)} },
		},
		{
			name:      "attempt request nonce",
			statement: `UPDATE nodecontrol.control_plane_authority_protocol_upgrade_attempts SET request_nonce=decode(repeat('f5',32),'hex') WHERE activation_id=$1`,
			args:      func(activationID uuid.UUID) []any { return []any{activationID} },
		},
		{
			name:      "registration credential policy",
			statement: `UPDATE nodecontrol.control_plane_authority_runtime_registration_results SET credential_policy_digest=decode(repeat('f6',32),'hex') WHERE activation_id=$1`,
			args:      func(activationID uuid.UUID) []any { return []any{activationID} },
		},
		{
			name:      "registration after attempt",
			statement: `UPDATE nodecontrol.control_plane_authority_runtime_registration_results SET recorded_at=$2 WHERE activation_id=$1`,
			args:      func(activationID uuid.UUID) []any { return []any{activationID, now.Add(time.Second)} },
		},
	}
	if len(mutations) != 8 {
		t.Fatalf("activation barrier mutation count=%d, want exact 8", len(mutations))
	}
	for index, mutation := range mutations {
		savepoint := fmt.Sprintf("task8_barrier_%02d", index)
		if _, err := tx.ExecContext(t.Context(), "SAVEPOINT "+savepoint); err != nil {
			t.Fatal(err)
		}
		activationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("task8-barrier-mutation:"+mutation.name))
		task8SeedProofActivation(t, tx, activationID, 930+index, now)
		if _, err := tx.ExecContext(t.Context(), `SET LOCAL session_replication_role=replica`); err != nil {
			t.Fatal("enable replica mode for exact barrier mutation:", err)
		}
		if _, err := tx.ExecContext(t.Context(), mutation.statement, mutation.args(activationID)...); err != nil {
			t.Fatalf("install CHECK-valid activation barrier mutation %s: %v", mutation.name, err)
		}
		if _, err := tx.ExecContext(t.Context(), `SET LOCAL session_replication_role=origin`); err != nil {
			t.Fatal("restore origin mode after exact barrier mutation:", err)
		}
		operationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("task8-barrier-operation:"+mutation.name))
		task8ExpectAuthorityGuardReject(t, tx, "inexact activation barrier "+mutation.name, "activation", `
INSERT INTO nodecontrol.control_plane_authority_fences(
 operation_id,effect_kind,scope_kind,authority_epoch,authority_sequence,scope_digest,provider_reservation_digest,
 provider_status,visibility_state,reserved_at,authority_protocol_profile,protocol_activation_id)
VALUES($1,'operator_transition','node',$2,1,$3,decode(repeat('72',32),'hex'),
 'reserved','fence_pending',$4,'claim_v1',$5)`, operationID, int64(2930+index), task8ProofScopeDigest(effectKind), now, activationID)
		if _, err := tx.ExecContext(t.Context(), "ROLLBACK TO SAVEPOINT "+savepoint); err != nil {
			t.Fatal("restore exact activation barrier fixture:", err)
		}
		if _, err := tx.ExecContext(t.Context(), "RELEASE SAVEPOINT "+savepoint); err != nil {
			t.Fatal(err)
		}
	}
}

func task8CleanupCommittedProofFixtures(t *testing.T, database *sql.DB, operationIDs, activationIDs []uuid.UUID, effectKind string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Errorf("begin committed proof fixture cleanup: %v", err)
		return
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `SET LOCAL session_replication_role=replica`); err != nil {
		t.Errorf("enable replica mode for committed proof fixture cleanup: %v", err)
		return
	}
	for _, operationID := range operationIDs {
		if _, err := tx.ExecContext(ctx, `DELETE FROM nodecontrol.node_state_transitions WHERE authority_operation_id=$1`, operationID); err != nil {
			t.Errorf("delete committed proof owner %s: %v", operationID, err)
			return
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM nodecontrol.control_plane_authority_fences WHERE operation_id=$1`, operationID); err != nil {
			t.Errorf("delete committed proof fence %s: %v", operationID, err)
			return
		}
	}
	for _, activationID := range activationIDs {
		for _, relation := range []string{
			"control_plane_authority_protocol_activation_releases",
			"control_plane_authority_protocol_activation_completions",
			"control_plane_authority_protocol_activations",
			"control_plane_authority_protocol_upgrade_attempts",
			"control_plane_authority_runtime_registration_results",
			"control_plane_authority_protocol_upgrade_intents",
		} {
			if _, err := tx.ExecContext(ctx, fmt.Sprintf("DELETE FROM nodecontrol.%s WHERE activation_id=$1", relation), activationID); err != nil {
				t.Errorf("delete committed activation fixture %s/%s: %v", relation, activationID, err)
				return
			}
		}
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM nodecontrol.node_inventory WHERE node_id=$1`, task8ProofFixtureUUID(effectKind, "node")); err != nil {
		t.Errorf("delete committed proof node: %v", err)
		return
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM nodecontrol.node_pops WHERE pop_code='proof-fixture'`); err != nil {
		t.Errorf("delete committed proof POP: %v", err)
		return
	}
	if err := tx.Commit(); err != nil {
		t.Errorf("commit committed proof fixture cleanup: %v", err)
	}
}

func task8ExpectTerminalProofBeforeFenceReject(t *testing.T, tx *sql.Tx, group task8AuthorityProofGroup, effectKind, rowCTID string) {
	t.Helper()
	statement, args := task8ProofOnlyTerminalUpdate(group, rowCTID, nil, nil)
	task8ExpectProofGuardReject(t, tx, "terminal proof before committed fence "+effectKind, statement, args...)
}

func task8AssertPreparedCommitmentCannotBeReplaced(t *testing.T, tx *sql.Tx, group task8AuthorityProofGroup, effectKind, rowCTID string) {
	t.Helper()
	changedJCS := []byte(`{"schema":"authority-effect-commitment.v1","stage":"replacement"}`)
	statement, args := task8ProofOnlyTerminalUpdate(group, rowCTID, changedJCS, nil)
	task8ExpectProofGuardReject(t, tx, "prepared terminal commitment JCS replacement "+effectKind, statement, args...)
	changedDigest := bytes.Repeat([]byte{0x8f}, 32)
	statement, args = task8ProofOnlyTerminalUpdate(group, rowCTID, nil, changedDigest)
	task8ExpectProofGuardReject(t, tx, "prepared terminal commitment digest replacement "+effectKind, statement, args...)
}

func task8ProofOnlyTerminalUpdate(group task8AuthorityProofGroup, rowCTID string, commitmentJCS, commitmentDigest []byte) (string, []any) {
	sets := make([]string, 0, 14)
	args := []any{rowCTID}
	add := func(column string, value any) {
		sets = append(sets, fmt.Sprintf("%s=$%d", column, len(args)+1))
		args = append(args, value)
	}
	if commitmentJCS != nil {
		add(group.prefix+"authority_effect_commitment_jcs", commitmentJCS)
	}
	if commitmentDigest != nil {
		add(group.prefix+"authority_effect_commitment_digest", commitmentDigest)
	}
	add(group.prefix+"authority_provider_head_jcs", []byte(`{"schema":"claim-v1-provider-head.v1","sequence":"811"}`))
	add(group.prefix+"authority_provider_head_digest", bytes.Repeat([]byte{0x82}, 32))
	add(group.prefix+"authority_effect_reason", "failed")
	add(group.prefix+"authority_activation_evidence_jcs", []byte(`{"schema":"authority-activation-evidence.v1","result":"not_applied"}`))
	add(group.prefix+"authority_activation_evidence_digest", bytes.Repeat([]byte{0x85}, 32))
	add(group.prefix+"authority_effect_resolution_jcs", []byte(`{"schema":"authority-effect-resolution.v1","disposition":"not_applied"}`))
	add(group.prefix+"authority_effect_resolution_digest", bytes.Repeat([]byte{0x86}, 32))
	return fmt.Sprintf("UPDATE %s SET %s WHERE ctid=$1::tid", group.table, strings.Join(sets, ",")), args
}

func task8AssertPreparedProofOwnerFrozen(t *testing.T, tx *sql.Tx, group task8AuthorityProofGroup, effectKind, rowCTID string, now time.Time) {
	t.Helper()
	terminalAt := now.Add(time.Minute)
	statement := ""
	args := []any{rowCTID, terminalAt}
	switch effectKind {
	case "grant_create", "grant_claim":
		terminalAt = now.Add(6 * time.Minute)
		args[1] = terminalAt
		statement = fmt.Sprintf(`UPDATE %s SET expired_at=$2::timestamptz,terminal_reason='expired',terminal_at=$2::timestamptz,
	 retention_until=$2::timestamptz+interval '31 days' WHERE ctid=$1::tid`, group.table)
	case "certificate_activate":
		statement = fmt.Sprintf(`UPDATE %s SET status='failed',failure_reason='provider_failed',updated_at=$2::timestamptz,terminal_at=$2::timestamptz,
	 retention_until=$2::timestamptz+interval '31 days' WHERE ctid=$1::tid`, group.table)
	case "certificate_revoke":
		statement = fmt.Sprintf(`UPDATE %s SET revoked_at=$2::timestamptz,revoke_reason='scheduled',status='revoked',updated_at=$2::timestamptz,
	 retention_until=$2::timestamptz+interval '31 days' WHERE ctid=$1::tid`, group.table)
	case "identity_epoch_advance", "operator_transition":
		disposition := "not_applied"
		if effectKind == "identity_epoch_advance" {
			disposition = "applied"
		}
		statement = fmt.Sprintf("UPDATE %s SET authority_effect_disposition=$2 WHERE ctid=$1::tid", group.table)
		args[1] = disposition
	case "security_incident_open":
		statement = fmt.Sprintf(`UPDATE %s SET last_evidence_digest=decode(repeat('ab',32),'hex'),
 occurrence_count=occurrence_count+1,last_occurred_at=$2 WHERE ctid=$1::tid`, group.table)
	case "security_incident_resolve":
		statement = fmt.Sprintf(`UPDATE %s SET remediation_digest=decode(repeat('ac',32),'hex'),resolution_at=$2,
 status='resolution_pending_agent_ack' WHERE ctid=$1::tid`, group.table)
	case "resource_envelope_activate":
		args[1] = now.Add(-time.Minute)
		statement = fmt.Sprintf("UPDATE %s SET issued_at=$2 WHERE ctid=$1::tid", group.table)
	case "desired_activate", "recovery_activate", "root_publish", "metadata_publish":
		statement = fmt.Sprintf("UPDATE %s SET status='failed',failure_reason='validation_failed',updated_at=$2,terminal_at=$2 WHERE ctid=$1::tid", group.table)
	default:
		t.Fatalf("unregistered prepared-owner freeze effect %q", effectKind)
	}
	task8ExpectProofGuardReject(t, tx, "prepared owner business advance "+effectKind, statement, args...)
}

func task8AssertTerminalProofFieldMutations(t *testing.T, tx *sql.Tx, group task8AuthorityProofGroup, effectKind, preparedCTID string, now time.Time) {
	t.Helper()
	for fieldIndex, suffix := range task8AuthorityProofSuffixes {
		savepoint := fmt.Sprintf("task8_terminal_field_%02d", fieldIndex)
		if _, err := tx.ExecContext(t.Context(), "SAVEPOINT "+savepoint); err != nil {
			t.Fatal(err)
		}
		shape := task8CanonicalTerminalShape(effectKind)
		if fieldIndex == 4 || fieldIndex == 5 {
			shape = "higher_authority"
		} else if fieldIndex == 6 {
			shape = "final_not_applied"
		} else if fieldIndex == 7 || fieldIndex == 8 || fieldIndex == 9 {
			shape = "rollback_resistant"
		}
		terminalCTID := task8InstallTerminalProof(t, tx, group, effectKind, preparedCTID, shape, now)
		task8PreflightTerminalProofGroup(t, tx, group, terminalCTID)
		column := group.prefix + suffix
		var value any
		switch fieldIndex {
		case 0, 2, 4, 10, 12:
			value = []byte(fmt.Sprintf(`{"field":%q,"mutation":2}`, suffix))
		case 1, 3, 5, 9, 11, 13:
			value = bytes.Repeat([]byte{byte(0x90 + fieldIndex)}, 32)
		case 6:
			value = "validation_rejected"
		case 7:
			value = now.Add(20 * time.Minute)
		case 8:
			value = now.Add(4 * time.Minute)
		}
		task8PreflightTerminalProofMutation(t, tx, group, terminalCTID, fieldIndex, value)
		task8ExpectAuthorityGuardReject(t, tx, "terminal field "+group.table+"/"+group.prefix+suffix, "", fmt.Sprintf(
			"UPDATE %s SET %s=$2 WHERE ctid=$1::tid", group.table, column), terminalCTID, value)
		if _, err := tx.ExecContext(t.Context(), "ROLLBACK TO SAVEPOINT "+savepoint); err != nil {
			t.Fatal("restore prepared proof after field mutation:", err)
		}
		if _, err := tx.ExecContext(t.Context(), "RELEASE SAVEPOINT "+savepoint); err != nil {
			t.Fatal(err)
		}
	}
}

func task8PreflightTerminalProofMutation(t *testing.T, tx *sql.Tx, group task8AuthorityProofGroup, rowCTID string, fieldIndex int, value any) {
	t.Helper()
	columns := task8PrefixedProofColumns(group.prefix)
	arguments := make([]string, len(columns))
	copy(arguments, columns)
	arguments[fieldIndex] = "$2"
	var valid bool
	if err := tx.QueryRowContext(t.Context(), fmt.Sprintf(`
SELECT nodecontrol.v7_authority_proof_group_valid(%s)
FROM %s WHERE ctid=$1::tid`, strings.Join(arguments, ","), group.table), rowCTID, value).Scan(&valid); err != nil {
		t.Fatalf("preflight legal terminal mutation %s/%s: %v", group.table, group.prefix+task8AuthorityProofSuffixes[fieldIndex], err)
	}
	if !valid {
		t.Fatalf("terminal mutation fixture %s/%s is not shape-valid before OLD/NEW guard", group.table, group.prefix+task8AuthorityProofSuffixes[fieldIndex])
	}
}

func task8PreflightTerminalProofGroup(t *testing.T, tx *sql.Tx, group task8AuthorityProofGroup, rowCTID string) {
	t.Helper()
	columns := task8PrefixedProofColumns(group.prefix)
	query := fmt.Sprintf(`SELECT nodecontrol.v7_authority_proof_group_valid(%s)
FROM %s WHERE ctid=$1::tid`, strings.Join(columns, ","), group.table)
	var valid bool
	if err := tx.QueryRowContext(t.Context(), query, rowCTID).Scan(&valid); err != nil {
		t.Fatalf("preflight complete terminal proof group %s/%s: %v", group.table, group.prefix, err)
	}
	if !valid {
		t.Fatalf("complete terminal fixture for %s/%s is not valid before single-field mutations", group.table, group.prefix)
	}
}

func task8SeedProofActivation(t *testing.T, tx *sql.Tx, activationID uuid.UUID, discriminator int, now time.Time) {
	t.Helper()
	var installationID uuid.UUID
	if err := tx.QueryRowContext(t.Context(), `SELECT installation_id FROM nodecontrol.control_plane_authority_protocol_migration_latches WHERE singleton_key ORDER BY installation_id LIMIT 1`).Scan(&installationID); err != nil {
		t.Fatal("load exact installed migration latch for claim-v1 closure:", err)
	}
	if _, err := tx.ExecContext(t.Context(), `SET LOCAL session_replication_role=replica`); err != nil {
		t.Fatal(err)
	}
	deploymentID := uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("task8-proof-deployment:%d", discriminator)))
	intentID := uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("task8-proof-intent:%d", discriminator)))
	registrationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("task8-proof-registration:%d", discriminator)))
	attemptID := uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("task8-proof-attempt:%d", discriminator)))
	preparationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("task8-proof-preparation:%d", discriminator)))
	completionID := uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("task8-proof-completion:%d", discriminator)))
	releasePreparationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("task8-proof-release:%d", discriminator)))
	openID := uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("task8-proof-open:%d", discriminator)))
	if _, err := tx.ExecContext(t.Context(), `
INSERT INTO nodecontrol.control_plane_authority_protocol_upgrade_intents(
 intent_id,installation_id,installation_kind,activation_id,request_nonce,incarnation_registration_id,
 provider_absence_proof_id,observed_deployment_id,database_identity_digest,local_runtime_isolation_digest,
 classification_state,created_at,canonical_evidence_bundle_jcs,canonical_body_jcs,body_digest)
VALUES($1,$2,'production',$3,decode(repeat('91',32),'hex'),$4,$5,$6,decode(repeat('92',32),'hex'),
 decode(repeat('93',32),'hex'),'pending',$7,decode('7b7d','hex'),decode('7b7d','hex'),decode(repeat('a1',32),'hex'))`,
		intentID, installationID, activationID,
		uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("task8-proof-incarnation:%d", discriminator))),
		uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("task8-proof-absence:%d", discriminator))),
		deploymentID, now); err != nil {
		t.Fatal("seed legal claim-v1 upgrade intent:", err)
	}
	if _, err := tx.ExecContext(t.Context(), `
INSERT INTO nodecontrol.control_plane_authority_runtime_registration_results(
 registration_id,upgrade_intent_digest,activation_id,provider_identity_digest,provider_endpoint_identity_digest,
 namespace,credential_policy_digest,database_incarnation_attestation_digest,provider_registration_digest,
 genesis_database_identity_digest,database_timeline_lineage_chain_digest,runtime_instance_binding_digest,
 runtime_instance_id,runtime_instance_generation,attestor_runtime_lease_digest,runtime_rebind_chain_digest,
 provider_head_digest,provider_phase,provider_control_sequence,database_point,recorded_at,
 canonical_evidence_bundle_jcs,canonical_body_jcs,body_digest)
VALUES($1,decode(repeat('a1',32),'hex'),$2,decode(repeat('94',32),'hex'),decode(repeat('95',32),'hex'),
 'task8-proof',decode(repeat('a6',32),'hex'),decode(repeat('ab',32),'hex'),decode(repeat('98',32),'hex'),
 decode(repeat('92',32),'hex'),decode(repeat('9a',32),'hex'),decode(repeat('ae',32),'hex'),$3,1,
 decode(repeat('9c',32),'hex'),decode(repeat('ad',32),'hex'),decode(repeat('9e',32),'hex'),
 'registered_pending_genesis',0,decode('01','hex'),$4,decode('7b7d','hex'),decode('7b7d','hex'),decode(repeat('a2',32),'hex'))`,
		registrationID, activationID,
		uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("task8-proof-runtime:%d", discriminator))), now); err != nil {
		t.Fatal("seed legal claim-v1 runtime registration:", err)
	}
	if _, err := tx.ExecContext(t.Context(), `
INSERT INTO nodecontrol.control_plane_authority_protocol_upgrade_attempts(
 attempt_id,upgrade_intent_digest,activation_id,preparation_id,completion_id,release_preparation_id,open_id,
 mode,deployment_id,request_nonce,environment_inventory_digest,environment_inventory_anchor_set_digest,
 local_runtime_isolation_digest,legacy_runtime_shutdown_digest,legacy_runtime_shutdown_set_digest,
 credential_policy_digest,database_legacy_absence_projection_digest,attempt_database_observation_digest,
 provider_namespace_absence_digest,database_inventory_digest,database_incarnation_attestation_digest,
 database_incarnation_registration_digest,runtime_registration_result_digest,runtime_rebind_chain_digest,
 runtime_instance_binding_digest,epoch_evidence_digest,genesis_epoch_transition_root_digest,selected_genesis_epoch,
 created_at,canonical_evidence_bundle_jcs,canonical_body_jcs,body_digest)
VALUES($1,decode(repeat('a1',32),'hex'),$2,$3,$4,$5,$6,'empty_in_place',$7,decode(repeat('91',32),'hex'),
 decode(repeat('a1',32),'hex'),decode(repeat('a2',32),'hex'),decode(repeat('93',32),'hex'),decode(repeat('a4',32),'hex'),
 decode(repeat('a5',32),'hex'),decode(repeat('a6',32),'hex'),decode(repeat('a7',32),'hex'),decode(repeat('a8',32),'hex'),
 decode(repeat('a9',32),'hex'),decode(repeat('aa',32),'hex'),decode(repeat('ab',32),'hex'),decode(repeat('ac',32),'hex'),
 decode(repeat('a2',32),'hex'),decode(repeat('ad',32),'hex'),decode(repeat('ae',32),'hex'),decode(repeat('af',32),'hex'),
 decode(repeat('b0',32),'hex'),1,$8,decode('7b7d','hex'),decode('7b7d','hex'),decode(repeat('a3',32),'hex'))`,
		attemptID, activationID, preparationID, completionID, releasePreparationID, openID, deploymentID, now); err != nil {
		t.Fatal("seed legal claim-v1 upgrade attempt:", err)
	}
	_, err := tx.ExecContext(t.Context(), `
INSERT INTO nodecontrol.control_plane_authority_protocol_activations(
 activation_id,mode,attempt_digest,deployment_id,database_identity_digest,database_incarnation_attestation_digest,
 genesis_database_incarnation_registration_digest,runtime_registration_result_digest,runtime_rebind_chain_digest,
 activation_runtime_instance_binding_digest,database_legacy_absence_projection_digest,attempt_database_inventory_digest,
 activation_database_observation_digest,provider_namespace_absence_digest,environment_inventory_digest,
 environment_inventory_anchor_set_digest,local_runtime_isolation_digest,legacy_runtime_shutdown_digest,
 legacy_runtime_shutdown_set_digest,credential_policy_digest,epoch_evidence_digest,genesis_epoch_transition_root_digest,
 selected_genesis_epoch,provider_identity_digest,provider_endpoint_identity_digest,namespace,protocol_profile,
 preparation_digest,activated_at,canonical_evidence_bundle_jcs,canonical_body_jcs,body_digest)
VALUES($1,'empty_in_place',decode(repeat('a3',32),'hex'),$2,decode(repeat('92',32),'hex'),decode(repeat('ab',32),'hex'),
	 decode(repeat('ac',32),'hex'),decode(repeat('a2',32),'hex'),decode(repeat('ad',32),'hex'),decode(repeat('ae',32),'hex'),
	 decode(repeat('a7',32),'hex'),decode(repeat('aa',32),'hex'),decode(repeat('5a',32),'hex'),decode(repeat('a9',32),'hex'),
	 decode(repeat('a1',32),'hex'),decode(repeat('a2',32),'hex'),decode(repeat('93',32),'hex'),decode(repeat('a4',32),'hex'),
	 decode(repeat('a5',32),'hex'),decode(repeat('a6',32),'hex'),decode(repeat('af',32),'hex'),decode(repeat('b0',32),'hex'),
	 1,decode(repeat('94',32),'hex'),decode(repeat('95',32),'hex'),'task8-proof','claim_v1',decode(repeat('66',32),'hex'),
	 $3,decode('7b7d','hex'),decode('7b7d','hex'),decode(repeat('a4',32),'hex'))`, activationID, deploymentID, now)
	if err != nil {
		t.Fatal("seed CHECK-valid claim-v1 activation:", err)
	}
	if _, err := tx.ExecContext(t.Context(), `
INSERT INTO nodecontrol.control_plane_authority_protocol_activation_completions(
 completion_id,activation_id,activation_digest,preparation_digest,provider_completion_digest,
 provider_completion_phase,database_activation_attestation_digest,current_database_incarnation_registration_digest,
 latest_runtime_rebind_result_digest_or_null,runtime_rebind_chain_digest,current_runtime_instance_binding_digest,
 credential_policy_digest,epoch_evidence_digest,genesis_epoch_transition_root_digest,selected_genesis_epoch,
 completed_at,canonical_evidence_bundle_jcs,canonical_body_jcs,body_digest)
VALUES($1,$2,decode(repeat('a4',32),'hex'),decode(repeat('66',32),'hex'),decode(repeat('b1',32),'hex'),
	 'genesis_completed_pending_release',decode(repeat('b2',32),'hex'),decode(repeat('ac',32),'hex'),NULL,
	 decode(repeat('ad',32),'hex'),decode(repeat('ae',32),'hex'),decode(repeat('a6',32),'hex'),
	 decode(repeat('af',32),'hex'),decode(repeat('b0',32),'hex'),1,$3,decode('7b7d','hex'),decode('7b7d','hex'),
 decode(repeat('a5',32),'hex'))`, completionID, activationID, now); err != nil {
		t.Fatal("seed CHECK-valid claim-v1 activation completion:", err)
	}
	if _, err := tx.ExecContext(t.Context(), `
INSERT INTO nodecontrol.control_plane_authority_protocol_activation_releases(
 release_preparation_id,open_id,activation_id,activation_digest,completion_digest,
 provider_release_preparation_digest,provider_release_phase,database_completion_attestation_digest,open_nonce,
 current_database_incarnation_registration_digest,latest_runtime_rebind_result_digest_or_null,runtime_rebind_chain_digest,
 current_runtime_instance_binding_digest,credential_policy_digest,epoch_evidence_digest,genesis_epoch_transition_root_digest,
 selected_genesis_epoch,released_at,canonical_evidence_bundle_jcs,canonical_body_jcs,body_digest)
VALUES($1,$2,$3,decode(repeat('a4',32),'hex'),decode(repeat('a5',32),'hex'),decode(repeat('b3',32),'hex'),
	 'genesis_release_prepared',decode(repeat('b4',32),'hex'),decode(repeat('b5',32),'hex'),decode(repeat('ac',32),'hex'),
	 NULL,decode(repeat('ad',32),'hex'),decode(repeat('ae',32),'hex'),decode(repeat('a6',32),'hex'),
	 decode(repeat('af',32),'hex'),decode(repeat('b0',32),'hex'),1,$4,decode('7b7d','hex'),decode('7b7d','hex'),
 decode(repeat('a6',32),'hex'))`, releasePreparationID, openID, activationID, now); err != nil {
		t.Fatal("seed CHECK-valid claim-v1 activation release:", err)
	}
	if _, err := tx.ExecContext(t.Context(), `SET LOCAL session_replication_role=origin`); err != nil {
		t.Fatal(err)
	}
}

func task8InsertClaimFence(t *testing.T, tx *sql.Tx, operationID, activationID uuid.UUID, authorityEpoch int64, effectKind string, now time.Time) {
	t.Helper()
	scopeDigest := task8ProofScopeDigest(effectKind)
	if _, err := tx.ExecContext(t.Context(), `
INSERT INTO nodecontrol.control_plane_authority_fences(
 operation_id,effect_kind,scope_kind,authority_epoch,authority_sequence,scope_digest,provider_reservation_digest,
 provider_status,visibility_state,reserved_at,authority_protocol_profile,protocol_activation_id)
VALUES($1,$2,$3,$4,1,$5,decode(repeat('72',32),'hex'),
 'reserved','fence_pending',$6,'claim_v1',$7)`, operationID, effectKind, task8ProofScope(effectKind), authorityEpoch, scopeDigest, now, activationID); err != nil {
		t.Fatal("insert claim-v1 proof fence:", err)
	}
}

func task8InsertWrongKindClaimFence(t *testing.T, tx *sql.Tx, operationID, referenceOperationID uuid.UUID, authorityEpoch int64, effectKind string, now time.Time) {
	t.Helper()
	result, err := tx.ExecContext(t.Context(), `
INSERT INTO nodecontrol.control_plane_authority_fences(
 operation_id,effect_kind,scope_kind,authority_epoch,authority_sequence,scope_digest,provider_reservation_digest,
 provider_status,visibility_state,reserved_at,authority_protocol_profile,protocol_activation_id)
SELECT $1,$2,reference.scope_kind,$3,2,reference.scope_digest,pg_catalog.sha256(pg_catalog.uuid_send($1)),
       'reserved','fence_pending',$4,'claim_v1',reference.protocol_activation_id
FROM nodecontrol.control_plane_authority_fences AS reference
WHERE reference.operation_id=$5`, operationID, effectKind, authorityEpoch, now, referenceOperationID)
	if err != nil {
		t.Fatal("insert wrong-kind claim-v1 fence:", err)
	}
	if rows, err := result.RowsAffected(); err != nil || rows != 1 {
		t.Fatalf("wrong-kind claim-v1 fence rows/error=%d/%v, want 1/nil", rows, err)
	}
}

func task8ProofScope(effectKind string) string {
	switch effectKind {
	case "root_publish", "metadata_publish":
		return "global_node_trust"
	default:
		return "node"
	}
}

func task8ProofScopeDigest(effectKind string) []byte {
	if task8ProofScope(effectKind) == "global_node_trust" {
		digest, err := hex.DecodeString("f9911591db5e19d9c4eff72a30b4fd0325416e63d8fb00f8fbd81f0860f6b0a1")
		if err != nil {
			panic(err)
		}
		return digest
	}
	return task8NodeScopeDigest(task8ProofFixtureUUID(effectKind, "node"))
}

func task8NodeScopeDigest(nodeID uuid.UUID) []byte {
	preimage := append([]byte("TALENRO-NODE-AUTHORITY-SCOPE-V1\x00"), nodeID[:]...)
	digest := sha256.Sum256(preimage)
	return digest[:]
}

func task8ProofFixtureUUID(effectKind, label string) uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte("task8-proof-fixture:"+effectKind+":"+label))
}

func task8WithReplicaDependencies(t *testing.T, tx *sql.Tx, insert func()) {
	t.Helper()
	if _, err := tx.ExecContext(t.Context(), `SET LOCAL session_replication_role=replica`); err != nil {
		t.Fatal("enable replica mode for inert proof dependencies:", err)
	}
	defer func() {
		if _, err := tx.ExecContext(context.Background(), `SET LOCAL session_replication_role=origin`); err != nil {
			t.Fatal("restore origin mode after inert proof dependencies:", err)
		}
	}()
	insert()
}

func task8SeedProofNode(t *testing.T, tx *sql.Tx, effectKind string, now time.Time) uuid.UUID {
	t.Helper()
	nodeID := task8ProofFixtureUUID(effectKind, "node")
	lineageID := task8ProofFixtureUUID(effectKind, "lineage")
	task8WithReplicaDependencies(t, tx, func() {
		if _, err := tx.ExecContext(t.Context(), `
INSERT INTO nodecontrol.node_pops(pop_code,iso_country,region,operator_state,created_at,updated_at)
VALUES('proof-fixture','US','proof-region','enabled',$1,$1)`, now); err != nil {
			t.Fatal("seed proof dependency POP:", err)
		}
		if _, err := tx.ExecContext(t.Context(), `
INSERT INTO nodecontrol.node_inventory(
 node_id,pop_code,operator_state,security_state,identity_state,identity_epoch,lineage_id,
 inventory_version,security_version,next_desired_generation,next_recovery_generation,created_at,updated_at)
VALUES($1,'proof-fixture','enabled','normal','active',1,$2,1,1,1,1,$3,$3)`, nodeID, lineageID, now); err != nil {
			t.Fatal("seed proof dependency node:", err)
		}
	})
	return nodeID
}

func task8SeedActiveTrustForSigning(t *testing.T, tx *sql.Tx, effectKind string, nodeID uuid.UUID, epoch int64, now time.Time) {
	t.Helper()
	rootOperationID := task8ProofFixtureUUID(effectKind, "active-root-operation")
	metadataOperationID := task8ProofFixtureUUID(effectKind, "active-metadata-operation")
	rootPublishID := task8ProofFixtureUUID(effectKind, "active-root-publish")
	metadataPublishID := task8ProofFixtureUUID(effectKind, "active-metadata-publish")
	task8InsertLegacyCommittedFence(t, tx, rootOperationID, epoch+4000, "root_publish", nodeID, now)
	task8InsertLegacyCommittedFence(t, tx, metadataOperationID, epoch+5000, "metadata_publish", nodeID, now)
	task8WithReplicaDependencies(t, tx, func() {
		if _, err := tx.ExecContext(t.Context(), `
INSERT INTO nodecontrol.node_root_metadata_publish_intents(
 publish_id,authority_operation_id,authority_epoch,authority_sequence,publish_kind,reason,base_root_version,
 base_metadata_version,reserved_version,canonical_payload,payload_digest,key_set_digest,current_key_ids,current_threshold,
 activation_deadline,published_envelope,published_envelope_digest,status,created_at,updated_at,terminal_at)
VALUES
 ($1,$2,$3,1,'root','normal',0,0,1,decode('7b7d','hex'),decode(repeat('91',32),'hex'),decode(repeat('92',32),'hex'),
  ARRAY[decode(repeat('93',32),'hex')]::bytea[],1,$7,decode('01','hex'),decode(repeat('94',32),'hex'),'active',$6,$6,$6),
 ($4,$5,$8,1,'metadata','normal',1,0,1,decode('7b7d','hex'),decode(repeat('95',32),'hex'),decode(repeat('96',32),'hex'),
  ARRAY[decode(repeat('97',32),'hex')]::bytea[],1,$7,decode('01','hex'),decode(repeat('98',32),'hex'),'active',$6,$6,$6)`,
			rootPublishID, rootOperationID, epoch+4000, metadataPublishID, metadataOperationID, now,
			time.Now().UTC().Add(24*time.Hour), epoch+5000); err != nil {
			t.Fatal("seed active root/metadata trust for signing proof:", err)
		}
		if _, err := tx.ExecContext(t.Context(), `
UPDATE nodecontrol.node_inventory
SET active_root_publish_id=$2,active_root_version=1,
    active_metadata_publish_id=$3,active_metadata_version=1,updated_at=$4
WHERE node_id=$1`, nodeID, rootPublishID, metadataPublishID, now); err != nil {
			t.Fatal("bind active root/metadata trust to signing proof node:", err)
		}
	})
}

func task8InsertLegacyCommittedFence(t *testing.T, tx *sql.Tx, operationID uuid.UUID, epoch int64, effectKind string, ownerNodeID uuid.UUID, now time.Time) {
	t.Helper()
	scopeDigest := task8ProofScopeDigest(effectKind)
	if task8ProofScope(effectKind) == "node" {
		scopeDigest = task8NodeScopeDigest(ownerNodeID)
	}
	task8WithReplicaDependencies(t, tx, func() {
		if _, err := tx.ExecContext(t.Context(), `
INSERT INTO nodecontrol.control_plane_authority_fences(
 operation_id,effect_kind,scope_kind,authority_epoch,authority_sequence,scope_digest,provider_reservation_digest,
 effect_digest,provider_status,provider_receipt_digest,db_system_id,db_timeline,required_lsn,
 visibility_state,reserved_at,effect_bound_at,terminal_at,authority_protocol_profile)
VALUES($1,$2,$3,$4,1,$5,decode(repeat('62',32),'hex'),decode(repeat('63',32),'hex'),
 'committed',decode(repeat('64',32),'hex'),1,1,'0/1','active',$6,$6,$6,'legacy_v6')`,
			operationID, effectKind, task8ProofScope(effectKind), epoch, scopeDigest, now); err != nil {
			t.Fatal("insert historical legacy proof dependency fence:", err)
		}
	})
}

func task8BindProofFence(t *testing.T, tx *sql.Tx, operationID uuid.UUID, now time.Time) {
	t.Helper()
	result, err := tx.ExecContext(t.Context(), `
UPDATE nodecontrol.control_plane_authority_fences
	SET effect_digest=decode(repeat('81',32),'hex'),db_system_id=1,db_timeline=1,required_lsn='0/1',effect_bound_at=$2
WHERE operation_id=$1 AND provider_status='reserved' AND effect_digest IS NULL`, operationID, now)
	if err != nil {
		t.Fatal("bind proof fence:", err)
	}
	if rows, err := result.RowsAffected(); err != nil || rows != 1 {
		t.Fatalf("bind proof fence rows=%d error=%v, want exactly one", rows, err)
	}
	var bound bool
	if err := tx.QueryRowContext(t.Context(), `
SELECT provider_status='reserved' AND visibility_state='fence_pending' AND effect_digest=decode(repeat('81',32),'hex')
FROM nodecontrol.control_plane_authority_fences WHERE operation_id=$1`, operationID).Scan(&bound); err != nil || !bound {
		t.Fatalf("proof fence is not reserved-bound after Bind: bound=%t error=%v", bound, err)
	}
}

func task8FinalizeProofFence(t *testing.T, tx *sql.Tx, operationID uuid.UUID, now time.Time) {
	t.Helper()
	result, err := tx.ExecContext(t.Context(), `
UPDATE nodecontrol.control_plane_authority_fences
SET provider_status='committed',provider_receipt_digest=decode(repeat('82',32),'hex'),
	visibility_state='active',terminal_at=$2
WHERE operation_id=$1 AND provider_status='reserved' AND effect_digest=decode(repeat('81',32),'hex')`, operationID, now)
	if err != nil {
		t.Fatal("finalize proof fence:", err)
	}
	if rows, err := result.RowsAffected(); err != nil || rows != 1 {
		t.Fatalf("finalize proof fence rows=%d error=%v, want exactly one", rows, err)
	}
	var committed bool
	if err := tx.QueryRowContext(t.Context(), `
SELECT provider_status='committed' AND visibility_state='active' AND effect_digest IS NOT NULL
FROM nodecontrol.control_plane_authority_fences WHERE operation_id=$1`, operationID).Scan(&committed); err != nil || !committed {
		t.Fatalf("proof fence is not committed-active after finalization: committed=%t error=%v", committed, err)
	}
}

func task8SeedPreparedProofOwner(t *testing.T, tx *sql.Tx, group task8AuthorityProofGroup, effectKind string, operationID uuid.UUID, epoch int64, now time.Time) string {
	t.Helper()
	nodeID := task8SeedProofNode(t, tx, effectKind, now)
	ownerID := task8ProofFixtureUUID(effectKind, "owner")
	commitmentJCS := []byte(`{"schema":"authority-effect-commitment.v1","stage":"prepared"}`)
	commitmentDigest := bytes.Repeat([]byte{0x81}, 32)
	var rowCTID string

	switch effectKind {
	case "grant_create":
		err := tx.QueryRowContext(t.Context(), `
INSERT INTO nodecontrol.node_enrollment_grants(
 grant_id,authority_operation_id,authority_epoch,authority_sequence,node_id,identity_epoch,
 token_digest,csr_digest,idempotency_digest,created_at,expires_at,
 create_authority_effect_commitment_jcs,create_authority_effect_commitment_digest)
VALUES($1,$2,$3,1,$4,1,decode(repeat('11',32),'hex'),decode(repeat('12',32),'hex'),decode(repeat('13',32),'hex'),$5,$6,$7,$8)
RETURNING ctid::text`, ownerID, operationID, epoch, nodeID, now, now.Add(5*time.Minute), commitmentJCS, commitmentDigest).Scan(&rowCTID)
		if err != nil {
			t.Fatal("insert prepared grant-create owner:", err)
		}

	case "grant_claim":
		createOperationID := task8ProofFixtureUUID(effectKind, "create-operation")
		task8InsertLegacyCommittedFence(t, tx, createOperationID, epoch+1000, "grant_create", nodeID, now)
		activateOperationID := task8ProofFixtureUUID(effectKind, "activate-operation")
		task8InsertLegacyCommittedFence(t, tx, activateOperationID, epoch+2000, "certificate_activate", nodeID, now)
		issuanceID := task8ProofFixtureUUID(effectKind, "result-issuance")
		task8WithReplicaDependencies(t, tx, func() {
			if _, err := tx.ExecContext(t.Context(), `
INSERT INTO nodecontrol.node_certificate_issuances(
 issuance_id,authority_operation_id,authority_epoch,authority_sequence,node_id,attempt_id,issuance_kind,identity_epoch,
 lineage_id,issuer_id,csr_sha256,public_key_sha256,template_sha256,request_digest,status,created_at,updated_at)
VALUES($1,$2,$3,1,$4,$5,'initial',1,$6,'fixture-ca',decode(repeat('21',32),'hex'),decode(repeat('22',32),'hex'),
 decode(repeat('23',32),'hex'),decode(repeat('24',32),'hex'),'pending',$7,$7)`, issuanceID, activateOperationID, epoch+2000, nodeID,
				task8ProofFixtureUUID(effectKind, "attempt"), task8ProofFixtureUUID(effectKind, "lineage"), now); err != nil {
				t.Fatal("seed grant-claim result issuance dependency:", err)
			}
		})
		task8WithReplicaDependencies(t, tx, func() {
			if err := tx.QueryRowContext(t.Context(), `
INSERT INTO nodecontrol.node_enrollment_grants(
 grant_id,authority_operation_id,authority_epoch,authority_sequence,node_id,identity_epoch,
 token_digest,csr_digest,idempotency_digest,created_at,expires_at)
VALUES($1,$2,$3,1,$4,1,decode(repeat('25',32),'hex'),decode(repeat('26',32),'hex'),decode(repeat('27',32),'hex'),$5,$6)
RETURNING ctid::text`, ownerID, createOperationID, epoch+1000, nodeID, now, now.Add(5*time.Minute)).Scan(&rowCTID); err != nil {
				t.Fatal("seed historical live grant-claim owner:", err)
			}
		})
		wrongKindOperationID := task8ProofFixtureUUID(effectKind, "wrong-kind-operation")
		task8InsertWrongKindClaimFence(t, tx, wrongKindOperationID, operationID, epoch, "grant_create", now)
		task8ExpectAuthorityGuardReject(t, tx, "grant claim prefix against grant-create fence",
			"authority proof/fence effect kind or authority tuple mismatch", `
UPDATE nodecontrol.node_enrollment_grants
SET claim_authority_operation_id=$2,claim_authority_epoch=$3,claim_authority_sequence=2,
 claim_authority_effect_commitment_jcs=$4,claim_authority_effect_commitment_digest=$5
WHERE ctid=$1::tid`, rowCTID, wrongKindOperationID, epoch, commitmentJCS, commitmentDigest)
		smuggledAt := now.Add(6 * time.Minute)
		task8ExpectProofGuardReject(t, tx, "grant claim preparation cannot smuggle expiration", `
UPDATE nodecontrol.node_enrollment_grants
SET claim_authority_operation_id=$2,claim_authority_epoch=$3,claim_authority_sequence=1,
 claim_authority_effect_commitment_jcs=$4,claim_authority_effect_commitment_digest=$5,
	 expired_at=$6::timestamptz,terminal_reason='expired',terminal_at=$6::timestamptz,retention_until=$6::timestamptz+interval '31 days'
WHERE ctid=$1::tid`, rowCTID, operationID, epoch, commitmentJCS, commitmentDigest, smuggledAt)
		if err := tx.QueryRowContext(t.Context(), `
UPDATE nodecontrol.node_enrollment_grants
SET claim_authority_operation_id=$2,claim_authority_epoch=$3,claim_authority_sequence=1,
 claim_authority_effect_commitment_jcs=$4,claim_authority_effect_commitment_digest=$5
WHERE ctid=$1::tid RETURNING ctid::text`, rowCTID, operationID, epoch, commitmentJCS, commitmentDigest).Scan(&rowCTID); err != nil {
			t.Fatal("prepare grant-claim proof through real UPDATE:", err)
		}

	case "certificate_activate":
		if err := tx.QueryRowContext(t.Context(), `
INSERT INTO nodecontrol.node_certificate_issuances(
 issuance_id,authority_operation_id,authority_epoch,authority_sequence,node_id,attempt_id,issuance_kind,identity_epoch,
 lineage_id,issuer_id,csr_sha256,public_key_sha256,template_sha256,request_digest,status,created_at,updated_at)
VALUES($1,$2,$3,1,$4,$5,'initial',1,$6,'fixture-ca',decode(repeat('31',32),'hex'),decode(repeat('32',32),'hex'),
 decode(repeat('33',32),'hex'),decode(repeat('34',32),'hex'),'pending',$7,$7)
RETURNING ctid::text`, ownerID, operationID, epoch, nodeID, task8ProofFixtureUUID(effectKind, "attempt"),
			task8ProofFixtureUUID(effectKind, "lineage"), now).Scan(&rowCTID); err != nil {
			t.Fatal("insert domain-prepared certificate-activate owner:", err)
		}
		resultAt := now.Add(time.Second)
		if err := tx.QueryRowContext(t.Context(), `
UPDATE nodecontrol.node_certificate_issuances
SET serial_bytes=decode('01','hex'),leaf_der=decode('02','hex'),leaf_der_sha256=decode(repeat('35',32),'hex'),
 chain_der=decode('03','hex'),chain_der_sha256=decode(repeat('36',32),'hex'),not_before=$2,not_after=$3,updated_at=$2
WHERE ctid=$1::tid RETURNING ctid::text`, rowCTID, resultAt, resultAt.Add(time.Hour)).Scan(&rowCTID); err != nil {
			t.Fatal("record one-time certificate issuance result before commitment:", err)
		}
		if err := tx.QueryRowContext(t.Context(), `
UPDATE nodecontrol.node_certificate_issuances
SET activation_authority_effect_commitment_jcs=$2,activation_authority_effect_commitment_digest=$3
WHERE ctid=$1::tid RETURNING ctid::text`, rowCTID, commitmentJCS, commitmentDigest).Scan(&rowCTID); err != nil {
			t.Fatal("commit certificate-activate proof after exact issuance result:", err)
		}

	case "certificate_revoke":
		activateOperationID := task8ProofFixtureUUID(effectKind, "activate-operation")
		issuanceID := task8ProofFixtureUUID(effectKind, "issuance")
		lineageID := task8ProofFixtureUUID(effectKind, "lineage")
		task8InsertLegacyCommittedFence(t, tx, activateOperationID, epoch+1000, "certificate_activate", nodeID, now)
		task8WithReplicaDependencies(t, tx, func() {
			if _, err := tx.ExecContext(t.Context(), `
INSERT INTO nodecontrol.node_certificate_issuances(
 issuance_id,authority_operation_id,authority_epoch,authority_sequence,node_id,attempt_id,issuance_kind,identity_epoch,lineage_id,
 issuer_id,csr_sha256,public_key_sha256,template_sha256,request_digest,status,serial_bytes,leaf_der,leaf_der_sha256,
 chain_der,chain_der_sha256,not_before,not_after,created_at,updated_at,terminal_at,retention_until)
VALUES($1,$2,$3,1,$4,$5,'initial',1,$6,'fixture-ca',decode(repeat('41',32),'hex'),decode(repeat('42',32),'hex'),
 decode(repeat('43',32),'hex'),decode(repeat('44',32),'hex'),'active',decode('01','hex'),decode('02','hex'),
 decode(repeat('45',32),'hex'),decode('03','hex'),decode(repeat('46',32),'hex'),$7,$8,$7,$7,$7,$9)`, issuanceID,
				activateOperationID, epoch+1000, nodeID, task8ProofFixtureUUID(effectKind, "attempt"), lineageID,
				now.Add(-time.Minute), now.Add(time.Hour), now.Add(31*24*time.Hour)); err != nil {
				t.Fatal("seed active issuance for certificate revoke:", err)
			}
		})
		if err := tx.QueryRowContext(t.Context(), `
INSERT INTO nodecontrol.node_certificates(
 certificate_id,issuance_id,authority_operation_id,authority_epoch,authority_sequence,node_id,identity_epoch,lineage_id,
 issuer_id,serial_bytes,leaf_der,leaf_der_sha256,public_key_sha256,chain_der_sha256,valid_from,valid_until,status,
 created_at,updated_at,retention_until)
VALUES($1,$2,$3,$4,1,$5,1,$6,'fixture-ca',decode('01','hex'),decode('02','hex'),decode(repeat('45',32),'hex'),
 decode(repeat('42',32),'hex'),decode(repeat('46',32),'hex'),$7,$8,'active',$7,$7,$9)
RETURNING ctid::text`, ownerID, issuanceID, activateOperationID, epoch+1000, nodeID, lineageID,
			now.Add(-time.Minute), now.Add(time.Hour), now.Add(31*24*time.Hour)).Scan(&rowCTID); err != nil {
			t.Fatal("insert active certificate-revoke owner:", err)
		}
		smuggledAt := now.Add(time.Minute)
		task8ExpectProofGuardReject(t, tx, "certificate revoke preparation cannot smuggle revocation", `
UPDATE nodecontrol.node_certificates
SET revoke_authority_operation_id=$2,revoke_authority_epoch=$3,revoke_authority_sequence=1,
 revoke_authority_effect_commitment_jcs=$4,revoke_authority_effect_commitment_digest=$5,
	 revoked_at=$6::timestamptz,revoke_reason='scheduled',status='revoked',updated_at=$6::timestamptz,retention_until=$6::timestamptz+interval '31 days'
WHERE ctid=$1::tid`, rowCTID, operationID, epoch, commitmentJCS, commitmentDigest, smuggledAt)
		if err := tx.QueryRowContext(t.Context(), `
UPDATE nodecontrol.node_certificates
SET revoke_authority_operation_id=$2,revoke_authority_epoch=$3,revoke_authority_sequence=1,
 revoke_authority_effect_commitment_jcs=$4,revoke_authority_effect_commitment_digest=$5
WHERE ctid=$1::tid RETURNING ctid::text`, rowCTID, operationID, epoch, commitmentJCS, commitmentDigest).Scan(&rowCTID); err != nil {
			t.Fatal("prepare certificate-revoke proof through real UPDATE:", err)
		}

	case "identity_epoch_advance", "operator_transition":
		dimension, fromState, toState, reason := "operator", "enabled", "draining", "drain_maintenance"
		if effectKind == "identity_epoch_advance" {
			dimension, fromState, toState, reason = "identity", "active", "recovery_pending", "identity_compromise"
		}
		if err := tx.QueryRowContext(t.Context(), `
INSERT INTO nodecontrol.node_state_transitions(
 transition_id,node_id,authority_operation_id,authority_epoch,authority_sequence,dimension,from_state,to_state,reason,
 aggregate_version,occurred_at,retention_until,authority_effect_kind,authority_effect_disposition,
 authority_effect_commitment_jcs,authority_effect_commitment_digest)
 VALUES($1,$2,$3,$4,1,$5,$6,$7,$8,1,$9,$10,$11,NULL,$12,$13)
RETURNING ctid::text`, ownerID, nodeID, operationID, epoch, dimension, fromState, toState, reason, now,
			now.Add(181*24*time.Hour), effectKind, commitmentJCS, commitmentDigest).Scan(&rowCTID); err != nil {
			t.Fatal("insert prepared state-transition owner:", err)
		}

	case "security_incident_open":
		if err := tx.QueryRowContext(t.Context(), `
INSERT INTO nodecontrol.node_security_incidents(
 incident_id,authority_operation_id,authority_epoch,authority_sequence,node_id,identity_epoch,fault_subtype,subtype_slot,status,
 first_evidence_digest,last_evidence_digest,occurrence_count,trust_context_digest,first_occurred_at,last_occurred_at,
 open_authority_effect_commitment_jcs,open_authority_effect_commitment_digest)
VALUES($1,$2,$3,1,$4,1,'identity_compromise',1,'open',decode(repeat('51',32),'hex'),decode(repeat('52',32),'hex'),
 1,decode(repeat('53',32),'hex'),$5,$5,$6,$7)
RETURNING ctid::text`, ownerID, operationID, epoch, nodeID, now, commitmentJCS, commitmentDigest).Scan(&rowCTID); err != nil {
			t.Fatal("insert prepared security-incident-open owner:", err)
		}

	case "security_incident_resolve":
		openOperationID := task8ProofFixtureUUID(effectKind, "open-operation")
		task8InsertLegacyCommittedFence(t, tx, openOperationID, epoch+1000, "security_incident_open", nodeID, now)
		task8WithReplicaDependencies(t, tx, func() {
			if err := tx.QueryRowContext(t.Context(), `
INSERT INTO nodecontrol.node_security_incidents(
 incident_id,authority_operation_id,authority_epoch,authority_sequence,node_id,identity_epoch,fault_subtype,subtype_slot,status,
 first_evidence_digest,last_evidence_digest,occurrence_count,trust_context_digest,first_occurred_at,last_occurred_at)
VALUES($1,$2,$3,1,$4,1,'identity_compromise',1,'open',decode(repeat('54',32),'hex'),decode(repeat('55',32),'hex'),
 1,decode(repeat('56',32),'hex'),$5,$5)
RETURNING ctid::text`, ownerID, openOperationID, epoch+1000, nodeID, now).Scan(&rowCTID); err != nil {
				t.Fatal("seed historical open security incident for resolve:", err)
			}
		})
		wrongKindOperationID := task8ProofFixtureUUID(effectKind, "wrong-kind-operation")
		task8InsertWrongKindClaimFence(t, tx, wrongKindOperationID, operationID, epoch, "security_incident_open", now)
		task8ExpectAuthorityGuardReject(t, tx, "incident resolve prefix against incident-open fence",
			"authority proof/fence effect kind or authority tuple mismatch", `
UPDATE nodecontrol.node_security_incidents
SET resolution_authority_operation_id=$2,resolution_authority_epoch=$3,resolution_authority_sequence=2,
 resolve_authority_effect_commitment_jcs=$4,resolve_authority_effect_commitment_digest=$5
WHERE ctid=$1::tid`, rowCTID, wrongKindOperationID, epoch, commitmentJCS, commitmentDigest)
		if err := tx.QueryRowContext(t.Context(), `
UPDATE nodecontrol.node_security_incidents
SET resolution_authority_operation_id=$2,resolution_authority_epoch=$3,resolution_authority_sequence=1,
 resolve_authority_effect_commitment_jcs=$4,resolve_authority_effect_commitment_digest=$5
WHERE ctid=$1::tid RETURNING ctid::text`, rowCTID, operationID, epoch, commitmentJCS, commitmentDigest).Scan(&rowCTID); err != nil {
			t.Fatal("prepare security-incident-resolve proof through real UPDATE:", err)
		}

	case "resource_envelope_activate":
		if err := tx.QueryRowContext(t.Context(), `
INSERT INTO nodecontrol.node_resource_envelopes(
 node_id,envelope_version,authority_operation_id,authority_epoch,authority_sequence,envelope_digest,canonical_package,
 deployment_key_id,signature,agent_cpu_millicores,agent_memory_bytes,agent_task_limit,agent_file_descriptor_limit,
 supervisor_cpu_millicores,supervisor_memory_bytes,supervisor_task_limit,supervisor_file_descriptor_limit,
 core_parent_cpu_millicores,core_parent_memory_bytes,core_parent_task_limit,core_parent_file_descriptor_limit,
 aggregate_slot_file_descriptor_limit,aggregate_slot_tmpfs_bytes,aggregate_slot_tmpfs_inodes,detected_host_capacity_digest,
 issued_at,created_at,activation_authority_effect_commitment_jcs,activation_authority_effect_commitment_digest)
VALUES($1,1,$2,$3,1,decode(repeat('61',32),'hex'),decode('7b7d','hex'),decode(repeat('62',32),'hex'),
 decode(repeat('63',64),'hex'),100,67108864,32,64,100,67108864,32,64,100,67108864,32,64,64,1048576,1,
 decode(repeat('64',32),'hex'),$4,$4,$5,$6)
RETURNING ctid::text`, nodeID, operationID, epoch, now, commitmentJCS, commitmentDigest).Scan(&rowCTID); err != nil {
			t.Fatal("insert prepared resource-envelope owner:", err)
		}

	case "desired_activate", "recovery_activate":
		signingKind := "desired"
		var recoveryValues []any
		var recoveryColumns, recoveryPlaceholders string
		if effectKind == "recovery_activate" {
			signingKind = "recovery"
			recoveryID := task8ProofFixtureUUID(effectKind, "recovery")
			recoveryOperationID := task8ProofFixtureUUID(effectKind, "recovery-operation")
			task8InsertLegacyCommittedFence(t, tx, recoveryOperationID, epoch+3000, "identity_epoch_advance", nodeID, now)
			task8WithReplicaDependencies(t, tx, func() {
				if _, err := tx.ExecContext(t.Context(), `
INSERT INTO nodecontrol.node_recovery_sessions(
 recovery_id,authority_operation_id,authority_epoch,authority_sequence,node_id,identity_epoch,reason,version,status,
 incident_set_digest,resume_operator_state,created_at,updated_at)
VALUES($1,$2,$3,1,$4,1,'authority_restore',1,'pending',decode(repeat('71',32),'hex'),'disabled',$5,$5)`,
					recoveryID, recoveryOperationID, epoch+3000, nodeID, now); err != nil {
					t.Fatal("seed recovery signing dependency:", err)
				}
			})
			recoveryColumns = ",recovery_id,recovery_reason,recovery_session_version,recovery_session_status,recovery_incident_set_digest,recovery_local_bindings_digest,recovery_supervisor_bindings_digest,recovery_required_action"
			recoveryPlaceholders = ",$8,'authority_restore',1,'pending',decode(repeat('71',32),'hex'),decode(repeat('72',32),'hex'),decode(repeat('73',32),'hex'),'hold_stopped'"
			recoveryValues = []any{recoveryID}
		}
		if effectKind == "desired_activate" {
			task8SeedActiveTrustForSigning(t, tx, effectKind, nodeID, epoch, now)
		}
		statement := fmt.Sprintf(`
INSERT INTO nodecontrol.node_state_signing_intents(
 signing_id,authority_operation_id,authority_epoch,authority_sequence,node_id,signing_kind,idempotency_key_digest,
 base_generation,reserved_generation,canonical_payload,payload_digest,root_version,metadata_version,expected_key_id,
 expected_public_key_digest,captured_inventory_version,captured_identity_epoch,captured_security_version,activation_deadline,
 status,created_at,updated_at%s)
VALUES($1,$2,$3,1,$4,$5,decode(repeat('74',32),'hex'),0,1,decode('7b7d','hex'),decode(repeat('75',32),'hex'),
 1,1,decode(repeat('76',32),'hex'),decode(repeat('77',32),'hex'),1,1,1,$6,'pending',$7,$7%s)
RETURNING ctid::text`, recoveryColumns, recoveryPlaceholders)
		args := []any{ownerID, operationID, epoch, nodeID, signingKind, time.Now().UTC().Add(24 * time.Hour), now}
		if effectKind == "recovery_activate" {
			args = append(args, recoveryValues...)
		}
		if err := tx.QueryRowContext(t.Context(), statement, args...).Scan(&rowCTID); err != nil {
			t.Fatal("insert domain-prepared state-signing owner:", err)
		}
		signedAt := now.Add(time.Second)
		if err := tx.QueryRowContext(t.Context(), `
UPDATE nodecontrol.node_state_signing_intents
SET signature=decode(repeat('78',64),'hex'),signature_verified_at=$2,updated_at=$2
WHERE ctid=$1::tid RETURNING ctid::text`, rowCTID, signedAt).Scan(&rowCTID); err != nil {
			t.Fatal("record one-time state signature before commitment:", err)
		}
		if err := tx.QueryRowContext(t.Context(), `
UPDATE nodecontrol.node_state_signing_intents
SET activation_authority_effect_commitment_jcs=$2,activation_authority_effect_commitment_digest=$3
WHERE ctid=$1::tid RETURNING ctid::text`, rowCTID, commitmentJCS, commitmentDigest).Scan(&rowCTID); err != nil {
			t.Fatal("commit state-signing proof after exact signature result:", err)
		}

	case "root_publish", "metadata_publish":
		publishKind := "root"
		if effectKind == "metadata_publish" {
			publishKind = "metadata"
		}
		if err := tx.QueryRowContext(t.Context(), `
INSERT INTO nodecontrol.node_root_metadata_publish_intents(
 publish_id,authority_operation_id,authority_epoch,authority_sequence,publish_kind,reason,base_root_version,
 base_metadata_version,reserved_version,canonical_payload,payload_digest,key_set_digest,current_key_ids,current_threshold,
 activation_deadline,status,created_at,updated_at)
VALUES($1,$2,$3,1,$4,'normal',0,0,1,decode('7b7d','hex'),decode(repeat('81',32),'hex'),decode(repeat('82',32),'hex'),
	 ARRAY[decode(repeat('83',32),'hex')]::bytea[],1,$5,'pending',$6,$6)
RETURNING ctid::text`, ownerID, operationID, epoch, publishKind, time.Now().UTC().Add(24*time.Hour), now).Scan(&rowCTID); err != nil {
			t.Fatal("insert domain-prepared root/metadata publish owner:", err)
		}
		signatureRole := "current_root"
		if publishKind == "metadata" {
			signatureRole = "metadata"
		}
		if _, err := tx.ExecContext(t.Context(), `
INSERT INTO nodecontrol.node_root_metadata_signature_shares(
 publish_id,key_id,physical_key_id,payload_digest,signature_role,signature,verified_at)
VALUES($1,decode(repeat('83',32),'hex'),decode(repeat('85',32),'hex'),decode(repeat('81',32),'hex'),$2,
 decode(repeat('86',64),'hex'),$3)`, ownerID, signatureRole, now); err != nil {
			t.Fatal("insert captured publish threshold share:", err)
		}
		publishedAt := now.Add(time.Second)
		if err := tx.QueryRowContext(t.Context(), `
UPDATE nodecontrol.node_root_metadata_publish_intents
SET published_envelope=decode('7b7d','hex'),published_envelope_digest=decode(repeat('84',32),'hex'),updated_at=$2
WHERE ctid=$1::tid RETURNING ctid::text`, rowCTID, publishedAt).Scan(&rowCTID); err != nil {
			t.Fatal("record one-time publish envelope before commitment:", err)
		}
		if err := tx.QueryRowContext(t.Context(), `
UPDATE nodecontrol.node_root_metadata_publish_intents
SET activation_authority_effect_commitment_jcs=$2,activation_authority_effect_commitment_digest=$3
WHERE ctid=$1::tid RETURNING ctid::text`, rowCTID, commitmentJCS, commitmentDigest).Scan(&rowCTID); err != nil {
			t.Fatal("commit publish proof after exact envelope result:", err)
		}

	default:
		t.Fatalf("unregistered real proof lifecycle effect %q for %s/%s", effectKind, group.table, group.prefix)
	}

	proofColumns := task8PrefixedProofColumns(group.prefix)
	var prepared bool
	if err := tx.QueryRowContext(t.Context(), fmt.Sprintf(`
SELECT num_nonnulls(%s)=2 AND %s=$2
FROM %s WHERE ctid=$1::tid`, strings.Join(proofColumns, ","), group.operationColumn, group.table), rowCTID, operationID).Scan(&prepared); err != nil || !prepared {
		t.Fatalf("real prepared proof preflight %s/%s/%s = %t, error=%v", effectKind, group.table, group.prefix, prepared, err)
	}
	return rowCTID
}

func task8CanonicalTerminalShape(effectKind string) string {
	switch effectKind {
	case "certificate_activate", "identity_epoch_advance", "desired_activate", "recovery_activate", "root_publish", "metadata_publish":
		return "final_not_applied"
	default:
		return "rollback_resistant"
	}
}

func task8InstallTerminalProof(t *testing.T, tx *sql.Tx, group task8AuthorityProofGroup, effectKind, rowCTID, shape string, now time.Time) string {
	t.Helper()
	columns := task8PrefixedProofColumns(group.prefix)
	sets := make([]string, 0, 20)
	args := []any{rowCTID}
	terminalValues := []any{
		[]byte(`{"schema":"claim-v1-provider-head.v1","sequence":"811"}`), bytes.Repeat([]byte{0x82}, 32),
		nil, nil,
		"none", now.Add(10 * time.Minute), now.Add(5 * time.Minute), bytes.Repeat([]byte{0x84}, 32),
		[]byte(`{"schema":"authority-activation-evidence.v1","result":"recorded"}`), bytes.Repeat([]byte{0x85}, 32),
		[]byte(`{"schema":"authority-effect-resolution.v1","disposition":"recorded"}`), bytes.Repeat([]byte{0x86}, 32),
	}
	switch shape {
	case "rollback_resistant":
	case "higher_authority":
		terminalValues[2] = []byte(`{"schema":"authority-checkpoint-anchor.v1","lsn":"0/811"}`)
		terminalValues[3] = bytes.Repeat([]byte{0x83}, 32)
		terminalValues[4] = "superseded"
		terminalValues[5], terminalValues[6], terminalValues[7] = nil, nil, nil
	case "final_not_applied":
		terminalValues[4] = "failed"
		terminalValues[5], terminalValues[6], terminalValues[7] = nil, nil, nil
	default:
		t.Fatalf("unregistered terminal proof shape %q", shape)
	}
	applied := shape == "rollback_resistant"
	for index, column := range columns[2:] {
		sets = append(sets, fmt.Sprintf("%s=$%d", column, len(args)+1))
		args = append(args, terminalValues[index])
	}
	add := func(column string, value any) {
		sets = append(sets, fmt.Sprintf("%s=$%d", column, len(args)+1))
		args = append(args, value)
	}
	terminalAt := now.Add(time.Minute)
	switch effectKind {
	case "grant_claim":
		if applied {
			add("consumed_at", terminalAt)
			add("consumption_attempt_id", task8ProofFixtureUUID(effectKind, "consumption-attempt"))
			add("consumption_request_digest", bytes.Repeat([]byte{0x91}, 32))
			add("result_issuance_id", task8ProofFixtureUUID(effectKind, "result-issuance"))
			sets = append(sets, "terminal_reason='consumed'")
			add("terminal_at", terminalAt)
			add("retention_until", terminalAt.Add(31*24*time.Hour))
		}
	case "certificate_activate":
		if applied {
			sets = append(sets, "status='active'", "failure_reason=NULL")
		} else {
			sets = append(sets, "status='failed'", "failure_reason='provider_failed'")
		}
		add("updated_at", terminalAt)
		add("terminal_at", terminalAt)
		add("retention_until", terminalAt.Add(31*24*time.Hour))
	case "certificate_revoke":
		if applied {
			add("revoked_at", terminalAt)
			sets = append(sets, "revoke_reason='scheduled'", "status='revoked'")
			add("updated_at", terminalAt)
		}
	case "identity_epoch_advance", "operator_transition":
		disposition := "applied"
		if shape != "rollback_resistant" {
			disposition = "not_applied"
		}
		add("authority_effect_disposition", disposition)
	case "security_incident_resolve":
		if applied {
			add("remediation_digest", bytes.Repeat([]byte{0x92}, 32))
			add("resolution_at", terminalAt)
			sets = append(sets, "status='resolution_pending_agent_ack'")
		}
	case "desired_activate", "recovery_activate":
		if applied {
			sets = append(sets, "status='active'", "failure_reason=NULL")
		} else {
			sets = append(sets, "status='failed'", "failure_reason='validation_failed'")
		}
		add("updated_at", terminalAt)
		add("terminal_at", terminalAt)
	case "root_publish", "metadata_publish":
		if applied {
			sets = append(sets, "status='active'", "failure_reason=NULL")
		} else {
			sets = append(sets, "status='failed'", "failure_reason='validation_failed'")
		}
		add("updated_at", terminalAt)
		add("terminal_at", terminalAt)
	}
	statement := fmt.Sprintf("UPDATE %s SET %s WHERE ctid=$1::tid RETURNING ctid::text", group.table, strings.Join(sets, ","))
	var updatedCTID string
	if err := tx.QueryRowContext(t.Context(), statement, args...).Scan(&updatedCTID); err != nil {
		t.Fatalf("legal terminal proof DML for %s/%s/%s did not reach proof guard: %v", effectKind, group.table, group.prefix, err)
	}
	return updatedCTID
}

func task8ExpectProofGuardReject(t *testing.T, tx *sql.Tx, label, statement string, args ...any) {
	t.Helper()
	task8ExpectAuthorityGuardReject(t, tx, label, "proof", statement, args...)
}

func task8ExpectAuthorityGuardReject(t *testing.T, tx *sql.Tx, label, wantMessageFragment, statement string, args ...any) {
	t.Helper()
	task8ExpectAuthorityGuardRejectCode(t, tx, label, "23514", wantMessageFragment, statement, args...)
}

func task8ExpectAuthorityGuardRejectCode(t *testing.T, tx *sql.Tx, label, wantSQLState, wantMessageFragment, statement string, args ...any) {
	t.Helper()
	const savepoint = "task8_guard_reject"
	if _, err := tx.ExecContext(t.Context(), "SAVEPOINT "+savepoint); err != nil {
		t.Fatal(err)
	}
	before := task8GuardTableSnapshot(t, tx)
	_, executionErr := tx.ExecContext(t.Context(), statement, args...)
	if _, err := tx.ExecContext(t.Context(), "ROLLBACK TO SAVEPOINT "+savepoint); err != nil {
		t.Fatal("restore after expected authority rejection:", err)
	}
	if _, err := tx.ExecContext(t.Context(), "RELEASE SAVEPOINT "+savepoint); err != nil {
		t.Fatal(err)
	}
	if executionErr == nil {
		t.Errorf("authority guard accepted forbidden %s", label)
	} else {
		var postgresError *pgconn.PgError
		if !errors.As(executionErr, &postgresError) || postgresError.Code != wantSQLState ||
			!strings.Contains(strings.ToLower(postgresError.Message), strings.ToLower(wantMessageFragment)) {
			t.Errorf("authority guard %s seam=%#v, want SQLSTATE %s containing %q", label, postgresError, wantSQLState, wantMessageFragment)
		}
	}
	if after := task8GuardTableSnapshot(t, tx); !bytes.Equal(after, before) {
		t.Fatalf("failed authority mutation %s changed complete protected bytes\nbefore=%s\nafter=%s", label, before, after)
	}
}

func task8AssertProofOwnershipNegatives(t *testing.T, tx *sql.Tx, now time.Time) {
	t.Helper()
	if _, err := tx.ExecContext(t.Context(), `SAVEPOINT task8_ownership`); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = tx.ExecContext(context.Background(), `ROLLBACK TO SAVEPOINT task8_ownership`)
		_, _ = tx.ExecContext(context.Background(), `RELEASE SAVEPOINT task8_ownership`)
	}()
	activationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("task8-proof-ownership-activation"))
	task8SeedProofActivation(t, tx, activationID, 77, now)
	operationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("task8-proof-ownership-operation"))
	task8InsertClaimFence(t, tx, operationID, activationID, 977, "operator_transition", now)
	group := task8AuthorityProofGroups[4]
	first := task8SeedPreparedProofOwner(t, tx, group, "operator_transition", operationID, 977, now)
	nodeID := task8ProofFixtureUUID("operator_transition", "node")
	task8ExpectProofGuardReject(t, tx, "operation reuse by second owner row", fmt.Sprintf(`
INSERT INTO %s(
 transition_id,node_id,authority_operation_id,authority_epoch,authority_sequence,dimension,from_state,to_state,reason,
 aggregate_version,occurred_at,retention_until,authority_effect_kind,authority_effect_disposition,
 authority_effect_commitment_jcs,authority_effect_commitment_digest)
VALUES($1,$2,$3,977,1,'operator','enabled','draining','drain_maintenance',2,$4,$5,
 'operator_transition','applied',decode('7b7d','hex'),decode(repeat('81',32),'hex'))`, group.table),
		task8ProofFixtureUUID("operator_transition", "second-owner"), nodeID, operationID, now, now.Add(181*24*time.Hour))

	wrongKindID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("task8-proof-wrong-kind"))
	task8InsertClaimFence(t, tx, wrongKindID, activationID, 978, "grant_create", now)
	task8ExpectProofGuardReject(t, tx, "wrong owner/effect kind", fmt.Sprintf(`
INSERT INTO %s(
 transition_id,node_id,authority_operation_id,authority_epoch,authority_sequence,dimension,from_state,to_state,reason,
 aggregate_version,occurred_at,retention_until,authority_effect_kind,authority_effect_disposition,
 authority_effect_commitment_jcs,authority_effect_commitment_digest)
VALUES($1,$2,$3,978,1,'operator','enabled','draining','drain_maintenance',3,$4,$5,
 'operator_transition','applied',decode('7b7d','hex'),decode(repeat('81',32),'hex'))`, group.table),
		task8ProofFixtureUUID("operator_transition", "wrong-kind-owner"), nodeID, wrongKindID, now, now.Add(181*24*time.Hour))

	directTerminalID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("task8-proof-direct-terminal"))
	task8InsertClaimFence(t, tx, directTerminalID, activationID, 979, "operator_transition", now)
	task8ExpectProofGuardReject(t, tx, "direct terminal INSERT", fmt.Sprintf(`
INSERT INTO %s(
 transition_id,node_id,authority_operation_id,authority_epoch,authority_sequence,dimension,from_state,to_state,reason,
 aggregate_version,occurred_at,retention_until,authority_effect_kind,authority_effect_disposition,
 authority_effect_commitment_jcs,authority_effect_commitment_digest,authority_provider_head_jcs,
 authority_provider_head_digest,authority_effect_reason,authority_activation_evidence_jcs,
 authority_activation_evidence_digest,authority_effect_resolution_jcs,authority_effect_resolution_digest)
VALUES($1,$2,$3,979,1,'operator','enabled','draining','drain_maintenance',4,$4,$5,
 'operator_transition','applied',decode('7b7d','hex'),decode(repeat('81',32),'hex'),decode('7b7d','hex'),
 decode(repeat('82',32),'hex'),'failed',decode('7b7d','hex'),decode(repeat('83',32),'hex'),decode('7b7d','hex'),
 decode(repeat('84',32),'hex'))`, group.table), task8ProofFixtureUUID("operator_transition", "direct-terminal-owner"),
		nodeID, directTerminalID, now, now.Add(181*24*time.Hour))

	// Once a source seal exists, no authority-bearing owner may advance. The
	// guard must reject at the source-freeze seam rather than using provider
	// Head or a staging artifact to synthesize authority.
	task8InsertLiteralTargetOnlyRelation(t, tx, task8LiteralDownRelationSeeds[13])
	var frozen bool
	if err := tx.QueryRowContext(t.Context(), `SELECT nodecontrol.v7_source_is_frozen()`).Scan(&frozen); err != nil || !frozen {
		t.Fatalf("real source-seal fixture did not freeze source: frozen=%t error=%v", frozen, err)
	}
	task8ExpectAuthorityGuardRejectCode(t, tx, "authority mutation after source seal", "55000", "source is frozen", fmt.Sprintf(
		"UPDATE %s SET authority_effect_commitment_jcs=decode('7b7d','hex') WHERE ctid=$1::tid", group.table), first)
	task8ExpectAuthorityGuardRejectCode(t, tx, "authority DELETE after source seal", "55000", "source is frozen", fmt.Sprintf(
		"DELETE FROM %s WHERE ctid=$1::tid", group.table), first)
}

func task8PrefixedProofColumns(prefix string) []string {
	columns := make([]string, len(task8AuthorityProofSuffixes))
	for index, suffix := range task8AuthorityProofSuffixes {
		columns[index] = prefix + suffix
	}
	return columns
}

const task8ProofGroupValiditySQL = `SELECT nodecontrol.v7_authority_proof_group_valid(
 $1::bytea,$2::bytea,$3::bytea,$4::bytea,$5::bytea,$6::bytea,$7::text,$8::timestamptz,
 $9::timestamptz,$10::bytea,$11::bytea,$12::bytea,$13::bytea,$14::bytea)`

func task8AssertDownFunctionLockBarrier(t *testing.T, database *sql.DB) {
	t.Helper()
	// This test observes lock transitions through PostgreSQL catalog queries.
	// Leave enough time for a heavily loaded Docker Desktop VM to validate the
	// complete 52-relation inventory after the middle sentinel is released.
	const observationTimeout = 30 * time.Second
	lockOrder := []string{
		"nodecontrol.control_plane_authority_protocol_migration_latches",
		"nodecontrol.control_plane_authority_protocol_downgrade_authorizations",
	}
	for _, pair := range task8PristineDownRegistry {
		if pair.classification == "authority_v7_non_control" {
			lockOrder = append(lockOrder, pair.table)
		}
	}
	for _, pair := range task8PristineDownRegistry {
		if pair.classification == "base_v6" {
			lockOrder = append(lockOrder, pair.table)
		}
	}
	lockOrder = append(lockOrder, "public.goose_db_version")
	if len(lockOrder) != 52 {
		t.Fatalf("live Down lock-order fixture count=%d, want literal 52", len(lockOrder))
	}
	oids := make([]int64, len(lockOrder))
	for index, relation := range lockOrder {
		if err := database.QueryRowContext(t.Context(), `SELECT $1::regclass::oid::bigint`, relation).Scan(&oids[index]); err != nil {
			t.Fatalf("resolve live Down lock relation %d %s: %v", index, relation, err)
		}
	}

	const insertCall = `SELECT * FROM nodecontrol.v7_insert_downgrade_authorization(
 $1::uuid,$2::uuid,$3::bytea,$4::bytea,$5::bigint,$6::bytea,$7::bytea,$8::bytea,$9::bytea,$10::numeric,
 $11::bytea,$12::text,$13::timestamptz,$14::timestamptz,$15::bytea,$16::bytea,$17::bytea,$18::bytea,$19::bytea,$20::bytea)`
	const consumeCall = `SELECT * FROM nodecontrol.v7_consume_down_guard(
 $1::uuid,$2::bytea,$3::bytea,$4::bytea,$5::bytea,$6::bytea)`
	const seedAuthorization = `INSERT INTO nodecontrol.control_plane_authority_protocol_downgrade_authorizations(
 authorization_id,installation_id,migration_latch_digest,database_identity_digest,migration_version,current_catalog_digest,
 pristine_downgrade_inventory_digest,provider_protocol_downgrade_retirement_set_digest,environment_inventory_anchor_set_digest,
 database_transaction_id,transaction_nonce,authorization_scope,issued_at,expires_at,authorization_envelope_jcs,
 pristine_inventory_body_jcs,provider_retirement_set_body_jcs,provider_retirement_evidence_jcs,canonical_body_jcs,body_digest)
 VALUES($1::uuid,$2::uuid,$3::bytea,$4::bytea,$5::bigint,$6::bytea,$7::bytea,$8::bytea,$9::bytea,$10::numeric,
 $11::bytea,$12::text,$13::timestamptz,$14::timestamptz,$16::bytea,$18::bytea,$19::bytea,$20::bytea,$15::bytea,$17::bytea)`
	for callIndex, call := range []struct {
		name              string
		sql               string
		seedAuthorization bool
	}{{"v7_insert_downgrade_authorization", insertCall, false}, {"v7_consume_down_guard", consumeCall, true}} {
		call := call
		t.Run(call.name+" exact ordered ACCESS EXCLUSIVE barrier", func(t *testing.T) {
			blocker, err := database.Conn(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			defer blocker.Close()
			blockerTx, err := blocker.BeginTx(t.Context(), nil)
			if err != nil {
				t.Fatal(err)
			}
			blockerOpen := true
			defer func() {
				if blockerOpen {
					_ = blockerTx.Rollback()
				}
			}()
			const sentinelIndex = 25
			if _, err := blockerTx.ExecContext(t.Context(), `LOCK TABLE nodecontrol.control_plane_authority_staging_import_capability_revocation_applications IN ACCESS EXCLUSIVE MODE`); err != nil {
				t.Fatal("lock middle Down-order sentinel:", err)
			}

			caller, err := database.Conn(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			defer caller.Close()
			callerTx, err := caller.BeginTx(t.Context(), nil)
			if err != nil {
				t.Fatal(err)
			}
			callerOpen := true
			defer func() {
				if callerOpen {
					_ = callerTx.Rollback()
				}
			}()
			fixture := task8BuildRound5RawInsertFixture(t, callerTx, task8Round5RawFixtureConfig{
				signatureAlgorithm:     "ed25519",
				signaturePolicyVersion: "1",
			})
			callArguments := fixture.arguments
			if call.seedAuthorization {
				if _, err := callerTx.ExecContext(t.Context(), seedAuthorization, fixture.arguments...); err != nil {
					t.Fatal("seed transaction-bound authorization for consume lock barrier:", err)
				}
				callArguments = []any{
					fixture.arguments[1], fixture.arguments[2], fixture.arguments[16], fixture.arguments[10],
					fixture.authorizedStateBody, fixture.authorizedStateDigest,
				}
			}
			var callerPID int
			if err := callerTx.QueryRowContext(t.Context(), `SELECT pg_backend_pid()`).Scan(&callerPID); err != nil {
				t.Fatal(err)
			}
			callResult := make(chan error, 1)
			go func() {
				_, callErr := callerTx.ExecContext(context.Background(), call.sql, callArguments...)
				callResult <- callErr
			}()

			waitDeadline := time.Now().Add(observationTimeout)
			waitObserved := false
			earlyReturned := false
			var earlyResult error
			for time.Now().Before(waitDeadline) {
				select {
				case earlyResult = <-callResult:
					earlyReturned = true
					goto barrierObservationComplete
				default:
				}
				var waiting bool
				if err := blockerTx.QueryRowContext(t.Context(), `
SELECT EXISTS(SELECT 1 FROM pg_catalog.pg_locks WHERE pid=$1 AND relation::oid::bigint=$2 AND mode='AccessExclusiveLock' AND NOT granted)`, callerPID, oids[sentinelIndex]).Scan(&waiting); err != nil {
					t.Fatal("observe middle-sentinel waiter:", err)
				}
				if waiting {
					waitObserved = true
					break
				}
				time.Sleep(20 * time.Millisecond)
			}

		barrierObservationComplete:
			if !waitObserved {
				t.Errorf("%s returned before waiting on lock-order sentinel: error=%v; want exact 20/6 function to lock all 52 relations before validation", call.name, earlyResult)
				_ = blockerTx.Rollback()
				blockerOpen = false
				if !earlyReturned {
					select {
					case <-callResult:
					case <-time.After(observationTimeout):
						t.Error("restricted Down function did not return after sentinel release")
					}
				}
				_ = callerTx.Rollback()
				callerOpen = false
				return
			}

			locked := task8DownAccessExclusiveLocks(t, blockerTx, callerPID)
			if got, want := len(locked), sentinelIndex+1; got != want {
				t.Errorf("%s pre-sentinel ACCESS EXCLUSIVE relation count = %d, want exact %d", call.name, got, want)
			}
			for index, oid := range oids {
				granted, present := locked[oid]
				switch {
				case index < sentinelIndex && (!present || !granted):
					t.Errorf("%s prefix relation %d %s is not granted before sentinel", call.name, index, lockOrder[index])
				case index == sentinelIndex && (!present || granted):
					t.Errorf("%s sentinel relation lock = present:%t granted:%t, want present/false", call.name, present, granted)
				case index > sentinelIndex && present:
					t.Errorf("%s acquired suffix relation %d %s before sentinel", call.name, index, lockOrder[index])
				}
			}
			if err := blockerTx.Commit(); err != nil {
				t.Fatal("release middle-sentinel barrier:", err)
			}
			blockerOpen = false
			var callErr error
			select {
			case callErr = <-callResult:
			case <-time.After(observationTimeout):
				t.Fatal("restricted Down function did not finish validation after sentinel release")
			}
			if callErr != nil {
				t.Fatalf("%s spec-valid barrier call failed after acquiring all locks: %v", call.name, callErr)
			}

			locked = task8DownAccessExclusiveLocks(t, database, callerPID)
			if got, want := len(locked), len(oids); got != want {
				t.Errorf("%s full ACCESS EXCLUSIVE relation count = %d, want exact %d", call.name, got, want)
			}
			for index, oid := range oids {
				if granted, present := locked[oid]; !present || !granted {
					t.Errorf("%s full ordered lock set omits relation %d %s: present=%t granted=%t", call.name, index, lockOrder[index], present, granted)
				}
			}

			writer, err := database.Conn(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			defer writer.Close()
			var writerPID int
			if err := writer.QueryRowContext(t.Context(), `SELECT pg_backend_pid()`).Scan(&writerPID); err != nil {
				t.Fatal(err)
			}
			writerResult := make(chan error, 1)
			popCode := "down-lock-writer-" + string(rune('a'+callIndex))
			go func() {
				_, writerErr := writer.ExecContext(context.Background(), `INSERT INTO nodecontrol.node_pops(pop_code,iso_country,region,operator_state,created_at,updated_at) VALUES($1,'US','down-lock-region','enabled',clock_timestamp(),clock_timestamp())`, popCode)
				writerResult <- writerErr
			}()
			writerDeadline := time.Now().Add(observationTimeout)
			writerBlocked := false
			for time.Now().Before(writerDeadline) {
				select {
				case writerErr := <-writerResult:
					t.Errorf("writer completed before %s transaction released all 52 locks: %v", call.name, writerErr)
					goto writerObservationComplete
				default:
				}
				var waitsOnLock, blockedByCaller bool
				if err := database.QueryRowContext(t.Context(), `
SELECT coalesce(wait_event_type='Lock',false), $1=ANY(pg_catalog.pg_blocking_pids($2))
FROM pg_catalog.pg_stat_activity WHERE pid=$2`, callerPID, writerPID).Scan(&waitsOnLock, &blockedByCaller); err != nil {
					t.Fatal("observe concurrent writer blocker:", err)
				}
				if waitsOnLock && blockedByCaller {
					writerBlocked = true
					break
				}
				time.Sleep(20 * time.Millisecond)
			}

		writerObservationComplete:
			if !writerBlocked {
				t.Errorf("writer did not expose pg_stat_activity Lock wait with %s backend as exact blocker", call.name)
			}
			_ = callerTx.Rollback()
			callerOpen = false
			if writerBlocked {
				select {
				case writerErr := <-writerResult:
					if writerErr != nil {
						t.Errorf("writer after full-lock release: %v", writerErr)
					}
				case <-time.After(observationTimeout):
					t.Error("writer remained blocked after full-lock release")
				}
			}
			_, _ = database.ExecContext(t.Context(), `DELETE FROM nodecontrol.node_pops WHERE pop_code=$1`, popCode)
		})
	}
}

type task8DownLockQuerier interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func task8DownAccessExclusiveLocks(t *testing.T, querier task8DownLockQuerier, pid int) map[int64]bool {
	t.Helper()
	rows, err := querier.QueryContext(t.Context(), `
SELECT relation::oid::bigint,granted FROM pg_catalog.pg_locks
WHERE pid=$1 AND locktype='relation' AND mode='AccessExclusiveLock' AND relation IS NOT NULL`, pid)
	if err != nil {
		t.Fatal("observe Down ACCESS EXCLUSIVE locks:", err)
	}
	defer rows.Close()
	result := make(map[int64]bool)
	for rows.Next() {
		var oid int64
		var granted bool
		if err := rows.Scan(&oid, &granted); err != nil {
			t.Fatal(err)
		}
		if previous, duplicate := result[oid]; duplicate && previous != granted {
			t.Fatalf("relation OID %d has conflicting granted states", oid)
		}
		result[oid] = granted
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return result
}

func task8GuardTableSnapshot(t *testing.T, database interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}) []byte {
	t.Helper()
	var snapshot []byte
	if err := database.QueryRowContext(t.Context(), `
SELECT convert_to(jsonb_build_object(
  'transitions',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY v.transition_id),'[]'::jsonb) FROM nodecontrol.node_state_transitions v),
  'enrollment_grants',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY v.grant_id),'[]'::jsonb) FROM nodecontrol.node_enrollment_grants v),
  'certificate_issuances',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY v.issuance_id),'[]'::jsonb) FROM nodecontrol.node_certificate_issuances v),
  'certificates',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY v.certificate_id),'[]'::jsonb) FROM nodecontrol.node_certificates v),
  'security_incidents',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY v.incident_id),'[]'::jsonb) FROM nodecontrol.node_security_incidents v),
  'resource_envelopes',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY v.node_id,v.envelope_version),'[]'::jsonb) FROM nodecontrol.node_resource_envelopes v),
  'state_signing_intents',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY v.signing_id),'[]'::jsonb) FROM nodecontrol.node_state_signing_intents v),
  'root_metadata_publish_intents',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY v.publish_id),'[]'::jsonb) FROM nodecontrol.node_root_metadata_publish_intents v),
  'fences',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY v.operation_id),'[]'::jsonb) FROM nodecontrol.control_plane_authority_fences v),
  'activations',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY v.activation_id),'[]'::jsonb) FROM nodecontrol.control_plane_authority_protocol_activations v),
  'activation_completions',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY v.completion_id),'[]'::jsonb) FROM nodecontrol.control_plane_authority_protocol_activation_completions v),
  'activation_releases',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY v.release_preparation_id),'[]'::jsonb) FROM nodecontrol.control_plane_authority_protocol_activation_releases v),
  'upgrade_intents',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY v.intent_id),'[]'::jsonb) FROM nodecontrol.control_plane_authority_protocol_upgrade_intents v),
  'runtime_registrations',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY v.registration_id),'[]'::jsonb) FROM nodecontrol.control_plane_authority_runtime_registration_results v),
  'upgrade_attempts',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY v.attempt_id),'[]'::jsonb) FROM nodecontrol.control_plane_authority_protocol_upgrade_attempts v),
  'legacy_source_seals',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY v.source_seal_id),'[]'::jsonb) FROM nodecontrol.control_plane_authority_legacy_source_seals v),
  'indeterminate_source_seals',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY v.source_seal_id),'[]'::jsonb) FROM nodecontrol.control_plane_authority_indeterminate_source_seals v),
  'staging_capabilities',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY v.capability_id),'[]'::jsonb) FROM nodecontrol.control_plane_authority_staging_import_capabilities v),
  'staging_applications',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY v.single_use_apply_id),'[]'::jsonb) FROM nodecontrol.control_plane_authority_fresh_restore_import_applications v),
  'pops',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY v.pop_code),'[]'::jsonb) FROM nodecontrol.node_pops v),
  'inventory',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY v.node_id),'[]'::jsonb) FROM nodecontrol.node_inventory v)
)::text,'UTF8')`).Scan(&snapshot); err != nil {
		t.Fatal(err)
	}
	return snapshot
}

func task8AssertAuthorityV7RejectsPreexistingRoles(t *testing.T, database *sql.DB) {
	t.Helper()
	upAsset := string(task8ReadFile(t, "../../db/migrations/assets/nodecontrol_authority_v7_up.sql"))
	roleStatement := task8AuthorityV7RolePreflightStatement(t, upAsset)
	roleNames := []string{
		"nodecontrol_upgrade_executor",
		"nodecontrol_migration_downgrader",
		"nodecontrol_staging_importer",
	}
	type roleSnapshotQueryer interface {
		QueryRowContext(context.Context, string, ...any) *sql.Row
	}
	type roleSnapshot struct {
		encoded         []byte
		roleCount       int64
		exactRoleCount  int64
		membershipCount int64
		gooseHighwater  int64
		latchCount      int64
	}
	snapshotRoles := func(queryer roleSnapshotQueryer) roleSnapshot {
		var snapshot roleSnapshot
		if err := queryer.QueryRowContext(t.Context(), `
WITH role_state AS (
  SELECT
    coalesce(jsonb_agg(jsonb_build_array(
      oid::text,rolname,rolcanlogin,rolsuper,rolcreatedb,rolcreaterole,rolinherit,
      rolreplication,rolbypassrls,rolconnlimit,rolvaliduntil,coalesce(rolconfig,ARRAY[]::text[])
    ) ORDER BY rolname COLLATE "C"),'[]'::jsonb) AS roles,
    count(*) AS role_count,
    count(*) FILTER (WHERE
      NOT rolcanlogin AND NOT rolsuper AND NOT rolcreatedb AND NOT rolcreaterole AND rolinherit AND
      NOT rolreplication AND NOT rolbypassrls AND rolconnlimit=-1 AND rolvaliduntil IS NULL AND rolconfig IS NULL
    ) AS exact_role_count
  FROM pg_catalog.pg_roles
  WHERE rolname=ANY($1::text[])
), membership_state AS (
  SELECT count(*) AS membership_count
  FROM pg_catalog.pg_auth_members AS membership
  JOIN pg_catalog.pg_roles AS member_role ON member_role.oid=membership.member
  JOIN pg_catalog.pg_roles AS granted_role ON granted_role.oid=membership.roleid
  WHERE member_role.rolname=ANY($1::text[]) OR granted_role.rolname=ANY($1::text[])
), database_state AS (
  SELECT
    (SELECT coalesce(max(version_id) FILTER (WHERE is_applied),0) FROM public.goose_db_version) AS goose_highwater,
    (SELECT count(*) FROM nodecontrol.control_plane_authority_protocol_migration_latches) AS latch_count
)
SELECT convert_to(jsonb_build_object(
         'roles',role_state.roles,
         'role_count',role_state.role_count,
         'exact_role_count',role_state.exact_role_count,
         'membership_count',membership_state.membership_count,
         'goose_highwater',database_state.goose_highwater,
         'latch_count',database_state.latch_count
       )::text,'UTF8'),
       role_state.role_count,role_state.exact_role_count,membership_state.membership_count,
       database_state.goose_highwater,database_state.latch_count
FROM role_state CROSS JOIN membership_state CROSS JOIN database_state`, roleNames).Scan(
			&snapshot.encoded, &snapshot.roleCount, &snapshot.exactRoleCount, &snapshot.membershipCount,
			&snapshot.gooseHighwater, &snapshot.latchCount,
		); err != nil {
			t.Fatal("snapshot installed authority-v7 capability roles:", err)
		}
		return snapshot
	}
	assertInstalledSnapshot := func(label string, snapshot roleSnapshot) {
		if snapshot.roleCount != int64(len(roleNames)) || snapshot.exactRoleCount != int64(len(roleNames)) ||
			snapshot.membershipCount != 0 || snapshot.gooseHighwater != 7 || snapshot.latchCount != 1 {
			t.Fatalf("%s role/exact/membership/highwater/latch = %d/%d/%d/%d/%d, want 3/3/0/7/1; snapshot=%s",
				label, snapshot.roleCount, snapshot.exactRoleCount, snapshot.membershipCount,
				snapshot.gooseHighwater, snapshot.latchCount, snapshot.encoded)
		}
	}
	tx, err := database.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatal("begin installed-role preflight probe:", err)
	}
	defer tx.Rollback()
	const fixtureLock int64 = 0x54414c454e524f37
	if _, err := tx.ExecContext(t.Context(), `SELECT pg_advisory_xact_lock($1)`, fixtureLock); err != nil {
		t.Fatal("acquire installed-role preflight lock:", err)
	}
	before := snapshotRoles(tx)
	assertInstalledSnapshot("before rejected role preflight", before)
	if _, err := tx.ExecContext(t.Context(), `SAVEPOINT task8_installed_role_preflight`); err != nil {
		t.Fatal("create installed-role preflight savepoint:", err)
	}
	_, applyErr := tx.ExecContext(t.Context(), roleStatement)
	var postgresError *pgconn.PgError
	const wantMessage = "authority v7 capability role names must all be absent before creation"
	if !errors.As(applyErr, &postgresError) || postgresError.Code != "42710" || postgresError.Message != wantMessage {
		t.Fatalf("installed-role preflight error = %v, want PostgreSQL 42710/%q before any CREATE ROLE", applyErr, wantMessage)
	}
	if _, err := tx.ExecContext(t.Context(), `ROLLBACK TO SAVEPOINT task8_installed_role_preflight`); err != nil {
		t.Fatal("roll back rejected installed-role preflight:", err)
	}
	after := snapshotRoles(tx)
	assertInstalledSnapshot("after rejected role preflight", after)
	if !bytes.Equal(after.encoded, before.encoded) {
		t.Fatalf("installed authority-v7 state changed across rejected role preflight: before=%s after=%s", before.encoded, after.encoded)
	}
	if _, err := tx.ExecContext(t.Context(), `RELEASE SAVEPOINT task8_installed_role_preflight`); err != nil {
		t.Fatal("release installed-role preflight savepoint:", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal("roll back installed-role preflight probe:", err)
	}
}

func task8AssertAuthorityV7ExactACLs(t *testing.T, database *sql.DB) {
	t.Helper()
	const downgrader = "nodecontrol_migration_downgrader"
	const staging = "nodecontrol_staging_importer"
	const upgrade = "nodecontrol_upgrade_executor"
	capabilityRoles := []string{downgrader, staging, upgrade}
	selectRelations := []string{
		"nodecontrol.control_plane_authority_protocol_migration_latches",
		"nodecontrol.control_plane_authority_protocol_downgrade_authorizations",
	}
	for _, relation := range task8PristineDownRegistry {
		selectRelations = append(selectRelations, relation.table)
	}
	lockRelations := append(append([]string(nil), selectRelations...), "public.goose_db_version")
	relationOIDByName := make(map[string]int64, len(lockRelations))
	relationNameByOID := make(map[int64]string, len(lockRelations))
	for _, relation := range lockRelations {
		var oid int64
		if err := database.QueryRowContext(t.Context(), `SELECT $1::regclass::oid::bigint`, relation).Scan(&oid); err != nil {
			t.Fatalf("resolve ACL relation %s: %v", relation, err)
		}
		if previous, duplicate := relationNameByOID[oid]; duplicate && previous != relation {
			t.Fatalf("ACL relation OID collision %d = %s/%s", oid, previous, relation)
		}
		relationOIDByName[relation] = oid
		relationNameByOID[oid] = relation
	}
	aclKey := func(role, relation, column, privilege string) string {
		return strings.Join([]string{role, relation, column, privilege}, "|")
	}
	wantRelationACLs := make([]string, 0, 117)
	for _, relation := range selectRelations {
		wantRelationACLs = append(wantRelationACLs, aclKey(downgrader, relation, "", "SELECT"))
	}
	for _, relation := range lockRelations {
		wantRelationACLs = append(wantRelationACLs, aclKey(downgrader, relation, "", "MAINTAIN"))
	}
	wantRelationACLs = append(wantRelationACLs,
		aclKey(downgrader, "nodecontrol.control_plane_authority_protocol_downgrade_authorizations", "", "INSERT"),
		aclKey(downgrader, "nodecontrol.control_plane_authority_protocol_downgrade_authorizations", "", "DELETE"),
		aclKey(downgrader, "nodecontrol.control_plane_authority_protocol_migration_latches", "", "DELETE"),
	)
	for _, relation := range []string{
		"nodecontrol.control_plane_authority_staging_import_capabilities",
		"nodecontrol.control_plane_authority_staging_import_capability_recovery_intents",
		"nodecontrol.control_plane_authority_staging_import_capability_recovery_applications",
		"nodecontrol.control_plane_authority_staging_import_capability_revocation_applications",
		"nodecontrol.control_plane_authority_fresh_restore_import_applications",
	} {
		wantRelationACLs = append(wantRelationACLs, aclKey(staging, relation, "", "SELECT"))
	}
	for _, relation := range []string{
		"nodecontrol.node_pops",
		"nodecontrol.node_failure_domains",
		"nodecontrol.node_capacity_profiles",
		"nodecontrol.node_inventory",
		"nodecontrol.control_plane_authority_fresh_restore_import_applications",
	} {
		wantRelationACLs = append(wantRelationACLs, aclKey(staging, relation, "", "INSERT"))
	}
	wantRelationACLs = append(wantRelationACLs,
		aclKey(staging, "nodecontrol.control_plane_authority_staging_import_capabilities", "body_digest", "UPDATE"),
	)
	sort.Strings(wantRelationACLs)
	rows, err := database.QueryContext(t.Context(), `
SELECT grantee.rolname,relation.oid::bigint,''::text,acl.privilege_type,acl.is_grantable
FROM pg_catalog.pg_class relation
CROSS JOIN LATERAL pg_catalog.aclexplode(coalesce(relation.relacl,pg_catalog.acldefault('r',relation.relowner))) acl
JOIN pg_catalog.pg_roles grantee ON grantee.oid=acl.grantee
WHERE grantee.rolname=ANY($1::text[])
UNION ALL
SELECT grantee.rolname,relation.oid::bigint,attribute.attname,acl.privilege_type,acl.is_grantable
FROM pg_catalog.pg_attribute attribute
JOIN pg_catalog.pg_class relation ON relation.oid=attribute.attrelid
CROSS JOIN LATERAL pg_catalog.aclexplode(attribute.attacl) acl
JOIN pg_catalog.pg_roles grantee ON grantee.oid=acl.grantee
WHERE attribute.attacl IS NOT NULL AND grantee.rolname=ANY($1::text[])`, capabilityRoles)
	if err != nil {
		t.Fatal("query exact authority-v7 relation ACLs:", err)
	}
	var gotRelationACLs []string
	for rows.Next() {
		var role, column, privilege string
		var oid int64
		var grantable bool
		if err := rows.Scan(&role, &oid, &column, &privilege, &grantable); err != nil {
			rows.Close()
			t.Fatal(err)
		}
		relation := relationNameByOID[oid]
		if relation == "" {
			relation = fmt.Sprintf("oid:%d", oid)
		}
		if grantable {
			privilege += ":GRANTABLE"
		}
		gotRelationACLs = append(gotRelationACLs, aclKey(role, relation, column, privilege))
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		t.Fatal(err)
	}
	rows.Close()
	sort.Strings(gotRelationACLs)
	if strings.Join(gotRelationACLs, "\n") != strings.Join(wantRelationACLs, "\n") {
		t.Fatalf("authority-v7 relation/column ACL registry mismatch\n got:\n%s\nwant:\n%s", strings.Join(gotRelationACLs, "\n"), strings.Join(wantRelationACLs, "\n"))
	}

	wantSchemaACLs := []string{
		downgrader + "|nodecontrol|USAGE",
		downgrader + "|public|USAGE",
		staging + "|nodecontrol|USAGE",
		upgrade + "|nodecontrol|USAGE",
	}
	rows, err = database.QueryContext(t.Context(), `
SELECT grantee.rolname,namespace.nspname,acl.privilege_type,acl.is_grantable
FROM pg_catalog.pg_namespace namespace
CROSS JOIN LATERAL pg_catalog.aclexplode(coalesce(namespace.nspacl,pg_catalog.acldefault('n',namespace.nspowner))) acl
JOIN pg_catalog.pg_roles grantee ON grantee.oid=acl.grantee
WHERE grantee.rolname=ANY($1::text[])`, capabilityRoles)
	if err != nil {
		t.Fatal("query exact authority-v7 schema ACLs:", err)
	}
	var gotSchemaACLs []string
	for rows.Next() {
		var role, schema, privilege string
		var grantable bool
		if err := rows.Scan(&role, &schema, &privilege, &grantable); err != nil {
			rows.Close()
			t.Fatal(err)
		}
		if grantable {
			privilege += ":GRANTABLE"
		}
		gotSchemaACLs = append(gotSchemaACLs, role+"|"+schema+"|"+privilege)
	}
	rows.Close()
	sort.Strings(gotSchemaACLs)
	sort.Strings(wantSchemaACLs)
	if strings.Join(gotSchemaACLs, "\n") != strings.Join(wantSchemaACLs, "\n") {
		t.Fatalf("authority-v7 schema ACL registry = %v, want %v", gotSchemaACLs, wantSchemaACLs)
	}

	var bootstrap string
	if err := database.QueryRowContext(t.Context(), `SELECT session_user`).Scan(&bootstrap); err != nil {
		t.Fatal(err)
	}
	wantNonOwnerFunctionACLs := []string{
		bootstrap + "|begin_staging_import|EXECUTE",
		bootstrap + "|v7_acquire_source_freeze_for_seal|EXECUTE",
		bootstrap + "|v7_consume_down_guard|EXECUTE",
		bootstrap + "|v7_insert_downgrade_authorization|EXECUTE",
		downgrader + "|v7_require_role|EXECUTE",
		staging + "|v7_require_role|EXECUTE",
		upgrade + "|v7_require_role|EXECUTE",
	}
	rows, err = database.QueryContext(t.Context(), `
SELECT coalesce(grantee.rolname,'PUBLIC'),function.proname,acl.privilege_type,acl.is_grantable
FROM pg_catalog.pg_proc function
JOIN pg_catalog.pg_namespace namespace ON namespace.oid=function.pronamespace
CROSS JOIN LATERAL pg_catalog.aclexplode(coalesce(function.proacl,pg_catalog.acldefault('f',function.proowner))) acl
LEFT JOIN pg_catalog.pg_roles grantee ON grantee.oid=acl.grantee
WHERE namespace.nspname='nodecontrol'
  AND function.proname IN ('begin_staging_import','v7_acquire_source_freeze_for_seal','v7_consume_down_guard','v7_insert_downgrade_authorization','v7_require_role')
  AND acl.grantee<>function.proowner`)
	if err != nil {
		t.Fatal("query exact authority-v7 non-owner function ACLs:", err)
	}
	var gotNonOwnerFunctionACLs []string
	for rows.Next() {
		var role, functionName, privilege string
		var grantable bool
		if err := rows.Scan(&role, &functionName, &privilege, &grantable); err != nil {
			rows.Close()
			t.Fatal(err)
		}
		if grantable {
			privilege += ":GRANTABLE"
		}
		gotNonOwnerFunctionACLs = append(gotNonOwnerFunctionACLs, role+"|"+functionName+"|"+privilege)
	}
	rows.Close()
	sort.Strings(gotNonOwnerFunctionACLs)
	sort.Strings(wantNonOwnerFunctionACLs)
	if strings.Join(gotNonOwnerFunctionACLs, "\n") != strings.Join(wantNonOwnerFunctionACLs, "\n") {
		t.Fatalf("authority-v7 non-owner function ACL registry = %v, want %v", gotNonOwnerFunctionACLs, wantNonOwnerFunctionACLs)
	}
}

func task8AssertAuthorityV7Round5ProductionDefinitions(t *testing.T, database *sql.DB) {
	t.Helper()
	definitions := make(map[string]string, 3)
	rows, err := database.QueryContext(t.Context(), `
SELECT procedure.proname,lower(pg_catalog.pg_get_functiondef(procedure.oid))
FROM pg_catalog.pg_proc AS procedure
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid=procedure.pronamespace
WHERE namespace.nspname='nodecontrol'
  AND procedure.proname=ANY($1::text[])
ORDER BY procedure.proname COLLATE "C"`, []string{
		"v7_consume_down_guard", "v7_insert_downgrade_authorization", "v7_require_role",
	})
	if err != nil {
		t.Fatal("query round5 production definitions:", err)
	}
	defer rows.Close()
	for rows.Next() {
		var name, definition string
		if err := rows.Scan(&name, &definition); err != nil {
			t.Fatal("scan round5 production definition:", err)
		}
		definitions[name] = definition
	}
	if err := rows.Err(); err != nil {
		t.Fatal("iterate round5 production definitions:", err)
	}
	if len(definitions) != 3 {
		t.Fatalf("round5 production definition count=%d, want 3", len(definitions))
	}
	insert := definitions["v7_insert_downgrade_authorization"]
	consume := definitions["v7_consume_down_guard"]
	requireRole := definitions["v7_require_role"]

	t.Run("runtime exact role and ACL closure", func(t *testing.T) {
		for _, stale := range []string{
			"nodecontrol_authority_v7_upgrade_executor",
			"nodecontrol_authority_v7_protocol_downgrader",
			"nodecontrol_authority_v7_inventory_staging_writer",
		} {
			if strings.Contains(consume, stale) {
				t.Errorf("installed consume definition retains stale capability-role identifier %q", stale)
			}
		}
		roleGate := regexp.MustCompile(`perform\s+(?:nodecontrol\.)?v7_require_role\('nodecontrol_migration_downgrader'(?:::name)?\)`)
		for name, body := range map[string]string{"insert": insert, "consume": consume} {
			if got := len(roleGate.FindAllStringIndex(body, -1)); got != 2 {
				t.Errorf("installed %s capability closure count=%d, want pre/post-lock 2", name, got)
			}
		}
		for _, required := range []string{
			"rolinherit", "rolconnlimit", "rolvaliduntil", "pg_db_role_setting", "pg_auth_members",
			"pg_shdepend", "pg_default_acl", "pg_init_privs", "aclexplode", "except all",
		} {
			if !strings.Contains(requireRole, required) {
				t.Errorf("installed v7_require_role exact closure lacks %q", required)
			}
		}
	})

	t.Run("runtime multi namespace evidence exact cover", func(t *testing.T) {
		for _, required := range []string{"v_provider_namespace_count", "v_namespace_evidence", "v_namespace_states_jcs"} {
			if !strings.Contains(insert, required) {
				t.Errorf("installed multi-namespace evidence closure lacks %q", required)
			}
		}
		formula := regexp.MustCompile(`1\s*\+\s*2\s*\*\s*jsonb_array_length\(\s*v_retirement_set\s*->\s*'retirements'(?:::text)?\s*\)\s*\+\s*v_provider_namespace_count`)
		if !formula.MatchString(insert) {
			t.Error("installed multi-namespace evidence closure lacks the 1+2P+N formula")
		}
		staleFormula := regexp.MustCompile(`1\s*\+\s*3\s*\*\s*jsonb_array_length\(\s*v_retirement_set\s*->\s*'retirements'(?:::text)?\s*\)`)
		if staleFormula.MatchString(insert) {
			t.Error("installed retirement evidence still uses the single-namespace 1+3P formula")
		}
	})

	t.Run("runtime signature metadata closure", func(t *testing.T) {
		if got := strings.Count(insert, "ecdsa-p256-sha256"); got < 2 {
			t.Errorf("installed P-256 signature algorithm registry count=%d, want top-level and nested", got)
		}
		if got := strings.Count(insert, "^[1-9][0-9]*$"); got < 2 {
			t.Errorf("installed canonical signature-policy registry count=%d, want top-level and nested", got)
		}
		if got := strings.Count(insert, "^[a-za-z0-9_-]{85}[aqgw]$"); got < 2 {
			t.Errorf("installed canonical raw-64 signature registry count=%d, want top-level and nested", got)
		}
	})

	t.Run("runtime response request digest binding", func(t *testing.T) {
		for _, required := range []string{
			"talenro.c12.retire-environment-inventory-membership-request.v1",
			"talenro.c12.retire-provider-protocol-for-downgrade-request.v1",
			"v_membership_request_body_jcs", "v_provider_request_body_jcs",
		} {
			if !strings.Contains(insert, required) {
				t.Errorf("installed response request-digest binding lacks %q", required)
			}
		}
	})

	t.Run("runtime historical evidence validity", func(t *testing.T) {
		for _, required := range []string{
			"v_validation_now", "v_membership_retired_at", "v_admin_issued_at", "v_admin_expires_at",
			"v_history_observed_at", "v_history_expires_at", "v_final_retired_at",
		} {
			if !strings.Contains(insert, required) {
				t.Errorf("installed historical evidence validation lacks %q", required)
			}
		}
		if strings.Contains(insert, "::timestamp with time zone <= clock_timestamp()") {
			t.Error("installed historical evidence still requires current freshness")
		}
	})

	t.Run("runtime retired member six-key ordering", func(t *testing.T) {
		stale := regexp.MustCompile(`(?:previous|value)->>'postgres_system_id'\)::numeric,\s*\((?:previous|value)->>'timeline'`)
		if got := len(stale.FindAllStringIndex(insert, -1)); got != 0 {
			t.Errorf("installed retired_members ordering includes timeline %d times", got)
		}
		exact := regexp.MustCompile(`(?:previous|value)->>'postgres_system_id'\)::numeric,\s*\((?:previous|value)->>'database_oid'`)
		if got := len(exact.FindAllStringIndex(insert, -1)); got != 2 {
			t.Errorf("installed retired_members six-key adjacency count=%d, want exact previous/current 2", got)
		}
	})
}

func task8AssertAuthorityV7RoleAttributeDrift(t *testing.T, database *sql.DB) {
	t.Helper()
	const insertCall = `SELECT * FROM nodecontrol.v7_insert_downgrade_authorization(
 NULL::uuid,NULL::uuid,NULL::bytea,NULL::bytea,NULL::bigint,NULL::bytea,NULL::bytea,NULL::bytea,NULL::bytea,NULL::numeric,
 NULL::bytea,NULL::text,NULL::timestamptz,NULL::timestamptz,NULL::bytea,NULL::bytea,NULL::bytea,NULL::bytea,NULL::bytea,NULL::bytea)`
	for _, testCase := range []struct {
		name     string
		mutation string
	}{
		{"NOINHERIT", `ALTER ROLE nodecontrol_migration_downgrader NOINHERIT`},
		{"connection limit", `ALTER ROLE nodecontrol_migration_downgrader CONNECTION LIMIT 1`},
		{"valid until", `ALTER ROLE nodecontrol_migration_downgrader VALID UNTIL '2037-01-01 00:00:00+00'`},
	} {
		t.Run("runtime role drift "+testCase.name, func(t *testing.T) {
			tx, err := database.BeginTx(t.Context(), nil)
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback()
			if _, err := tx.ExecContext(t.Context(), testCase.mutation); err != nil {
				t.Fatal("apply role-attribute drift:", err)
			}
			_, callErr := tx.ExecContext(t.Context(), insertCall)
			var postgresError *pgconn.PgError
			if !errors.As(callErr, &postgresError) || postgresError.Code != "55000" || postgresError.Message != "authority v7 capability role drift" {
				t.Errorf("role-attribute drift error=%v, want PostgreSQL 55000 exact capability-role drift", callErr)
			}
			if err := tx.Rollback(); err != nil {
				t.Fatal("roll back role-attribute drift:", err)
			}
		})
	}
}

type task8Round5SignedEvidence struct {
	schema   string
	digest   []byte
	body     []byte
	envelope []byte
}

type task8Round5RetiredMember struct {
	environmentRecordDigest       []byte
	environmentAttestationDigest  []byte
	deploymentID                  uuid.UUID
	postgresSystemID              uint64
	timeline                      uint64
	databaseOID                   uint64
	databaseName                  string
	databaseIdentityDigest        []byte
	environmentInstanceGeneration uint64
	providerIdentityDigest        []byte
	providerEndpointDigest        []byte
	providerNamespace             string
	providerProfile               string
}

func (member task8Round5RetiredMember) wire() map[string]any {
	return map[string]any{
		"environment_record_digest":         hex.EncodeToString(member.environmentRecordDigest),
		"environment_attestation_digest":    hex.EncodeToString(member.environmentAttestationDigest),
		"deployment_id":                     member.deploymentID.String(),
		"postgres_system_id":                strconv.FormatUint(member.postgresSystemID, 10),
		"timeline":                          strconv.FormatUint(member.timeline, 10),
		"database_oid":                      strconv.FormatUint(member.databaseOID, 10),
		"database_name":                     member.databaseName,
		"database_identity_digest":          hex.EncodeToString(member.databaseIdentityDigest),
		"environment_instance_generation":   strconv.FormatUint(member.environmentInstanceGeneration, 10),
		"provider_identity_digest":          hex.EncodeToString(member.providerIdentityDigest),
		"provider_endpoint_identity_digest": hex.EncodeToString(member.providerEndpointDigest),
		"provider_namespace":                member.providerNamespace,
		"provider_profile":                  member.providerProfile,
	}
}

type task8Round5ProviderPair struct {
	identityDigest []byte
	endpointDigest []byte
	members        []task8Round5RetiredMember
	retirementID   uuid.UUID
	admin          task8Round5SignedEvidence
	historyByNS    map[string]task8Round5SignedEvidence
	final          task8Round5SignedEvidence
}

type task8Round5RawFixtureConfig struct {
	multiNamespace          bool
	inverseTimelineOrder    bool
	historicalEvidence      bool
	signatureAlgorithm      string
	signaturePolicyVersion  string
	corruptMembershipDigest bool
	corruptProviderDigest   bool
}

type task8Round5RawInsertFixture struct {
	arguments             []any
	authorizedStateBody   []byte
	authorizedStateDigest []byte
}

func task8Round5JCS(t *testing.T, value any) []byte {
	t.Helper()
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		t.Fatal("encode round5 canonical JSON:", err)
	}
	return bytes.TrimSuffix(buffer.Bytes(), []byte{'\n'})
}

func task8Round5Digest(label string) []byte {
	digest := sha256.Sum256([]byte("task8-round5:" + label))
	return append([]byte(nil), digest[:]...)
}

func task8Round5DomainDigest(schema string, body []byte) []byte {
	hash := sha256.New()
	_, _ = hash.Write([]byte("talenro.c12." + schema))
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write(body)
	return hash.Sum(nil)
}

func task8Round5Sign(t *testing.T, schema, role, algorithm, policyVersion string, body []byte) task8Round5SignedEvidence {
	t.Helper()
	digest := task8Round5DomainDigest(schema, body)
	envelope := task8Round5JCS(t, map[string]any{
		"body":                     json.RawMessage(body),
		"body_digest":              hex.EncodeToString(digest),
		"schema":                   schema,
		"signature":                strings.Repeat("A", 86),
		"signature_algorithm":      algorithm,
		"signature_policy_version": policyVersion,
		"signer_key_id":            "task8-round5-key",
		"signer_role":              role,
		"trust_root_digest":        hex.EncodeToString(task8Round5Digest("trust-root")),
	})
	return task8Round5SignedEvidence{schema: schema, digest: digest, body: body, envelope: envelope}
}

func task8Round5EvidenceBundle(t *testing.T, messageDigest []byte, evidence []task8Round5SignedEvidence) []byte {
	t.Helper()
	sorted := append([]task8Round5SignedEvidence(nil), evidence...)
	sort.Slice(sorted, func(left, right int) bool {
		if compared := bytes.Compare(sorted[left].digest, sorted[right].digest); compared != 0 {
			return compared < 0
		}
		return sorted[left].schema < sorted[right].schema
	})
	items := make([]map[string]any, len(sorted))
	for index, item := range sorted {
		items[index] = map[string]any{
			"body_digest":                hex.EncodeToString(item.digest),
			"canonical_body_or_null":     nil,
			"canonical_envelope_or_null": json.RawMessage(item.envelope),
			"evidence_kind":              "external_signed_envelope",
			"schema":                     item.schema,
		}
	}
	return task8Round5JCS(t, map[string]any{
		"evidence":            items,
		"evidence_count":      strconv.Itoa(len(items)),
		"message_body_digest": hex.EncodeToString(messageDigest),
		"message_schema":      "provider-protocol-downgrade-retirement-set.v1",
	})
}

func task8BuildRound5RawInsertFixture(t *testing.T, tx *sql.Tx, config task8Round5RawFixtureConfig) task8Round5RawInsertFixture {
	t.Helper()
	var installationID uuid.UUID
	var installationKind, databaseTransactionID string
	var migrationLatchDigest, databaseIdentityDigest, currentCatalogDigest []byte
	var transactionTime time.Time
	if err := tx.QueryRowContext(t.Context(), `
SELECT installation_id,installation_kind,body_digest,database_identity_digest,up_catalog_digest,
       txid_current()::text,transaction_timestamp()
FROM nodecontrol.control_plane_authority_protocol_migration_latches`).Scan(
		&installationID, &installationKind, &migrationLatchDigest, &databaseIdentityDigest, &currentCatalogDigest,
		&databaseTransactionID, &transactionTime,
	); err != nil {
		t.Fatal("read authority-v7 latch for raw retirement fixture:", err)
	}
	if installationKind != "disposable_fixture" {
		t.Fatalf("raw retirement fixture installation kind=%q, want disposable_fixture", installationKind)
	}
	transactionTime = transactionTime.UTC()
	if config.signatureAlgorithm == "" {
		config.signatureAlgorithm = "ed25519"
	}
	if config.signaturePolicyVersion == "" {
		config.signaturePolicyVersion = "1"
	}

	providerA := task8Round5Digest("provider-a")
	endpointA := task8Round5Digest("endpoint-a")
	providerB := task8Round5Digest("provider-b")
	endpointB := task8Round5Digest("endpoint-b")
	member := func(index int, provider, endpoint []byte, namespace string, timeline, databaseOID uint64, deployment uuid.UUID) task8Round5RetiredMember {
		return task8Round5RetiredMember{
			environmentRecordDigest:       task8Round5Digest(fmt.Sprintf("environment-record-%d", index)),
			environmentAttestationDigest:  task8Round5Digest(fmt.Sprintf("environment-attestation-%d", index)),
			deploymentID:                  deployment,
			postgresSystemID:              42,
			timeline:                      timeline,
			databaseOID:                   databaseOID,
			databaseName:                  fmt.Sprintf("round5_%02d", index),
			databaseIdentityDigest:        task8Round5Digest(fmt.Sprintf("database-identity-%d", index)),
			environmentInstanceGeneration: uint64(index),
			providerIdentityDigest:        append([]byte(nil), provider...),
			providerEndpointDigest:        append([]byte(nil), endpoint...),
			providerNamespace:             namespace,
			providerProfile:               "legacy_v6",
		}
	}
	deploymentA := uuid.MustParse("81000000-0000-4000-8000-000000000001")
	deploymentB := uuid.MustParse("82000000-0000-4000-8000-000000000001")
	members := []task8Round5RetiredMember{
		member(1, providerA, endpointA, "alpha", 1, 1, deploymentA),
	}
	if config.inverseTimelineOrder {
		members[0].timeline = 2
		members = append(members, member(2, providerA, endpointA, "alpha", 1, 2, deploymentA))
	} else if config.multiNamespace {
		members = append(members,
			member(2, providerA, endpointA, "beta", 2, 2, deploymentA),
			member(3, providerB, endpointB, "gamma", 1, 1, deploymentB),
		)
	}
	pairs := []task8Round5ProviderPair{{
		identityDigest: providerA,
		endpointDigest: endpointA,
		retirementID:   uuid.NewSHA1(uuid.NameSpaceOID, []byte("task8-round5-retirement-a")),
		historyByNS:    make(map[string]task8Round5SignedEvidence),
	}}
	if config.multiNamespace {
		pairs = append(pairs, task8Round5ProviderPair{
			identityDigest: providerB,
			endpointDigest: endpointB,
			retirementID:   uuid.NewSHA1(uuid.NameSpaceOID, []byte("task8-round5-retirement-b")),
			historyByNS:    make(map[string]task8Round5SignedEvidence),
		})
	}
	sort.Slice(pairs, func(left, right int) bool {
		if compared := bytes.Compare(pairs[left].identityDigest, pairs[right].identityDigest); compared != 0 {
			return compared < 0
		}
		return bytes.Compare(pairs[left].endpointDigest, pairs[right].endpointDigest) < 0
	})
	for index := range pairs {
		for _, candidate := range members {
			if bytes.Equal(candidate.providerIdentityDigest, pairs[index].identityDigest) && bytes.Equal(candidate.providerEndpointDigest, pairs[index].endpointDigest) {
				pairs[index].members = append(pairs[index].members, candidate)
			}
		}
	}

	releaseScope := "release/task8-round5"
	environmentInventoryDigest := task8Round5Digest("environment-inventory")
	environmentAnchorDigest := task8Round5Digest("environment-anchor")
	membershipIssuedAt := transactionTime.Add(-time.Minute)
	membershipExpiresAt := transactionTime.Add(time.Minute)
	adminIssuedAt := transactionTime.Add(-time.Minute)
	adminExpiresAt := transactionTime.Add(time.Minute)
	historyObservedAt := transactionTime.Add(-time.Minute)
	historyExpiresAt := transactionTime.Add(time.Minute)
	retiredAt := transactionTime.Add(-30 * time.Second)
	if config.historicalEvidence {
		membershipIssuedAt = transactionTime.Add(-10 * time.Minute)
		membershipExpiresAt = transactionTime.Add(-8 * time.Minute)
		adminIssuedAt = transactionTime.Add(-10 * time.Minute)
		adminExpiresAt = transactionTime.Add(-8 * time.Minute)
		historyObservedAt = transactionTime.Add(-10 * time.Minute)
		historyExpiresAt = transactionTime.Add(-17 * time.Minute / 2)
		retiredAt = transactionTime.Add(-9 * time.Minute)
	}
	timeText := func(value time.Time) string { return value.UTC().Format(time.RFC3339Nano) }
	recordDigests := make([]string, len(members))
	memberWires := make([]map[string]any, len(members))
	for index, value := range members {
		recordDigests[index] = hex.EncodeToString(value.environmentRecordDigest)
		memberWires[index] = value.wire()
	}
	environmentMemberSetDigest := task8Round5DomainDigest("environment-inventory-member-set.v1", task8Round5JCS(t, recordDigests))
	membershipRequestNonce := task8Round5Digest("membership-request-nonce")
	membershipRequest := map[string]any{
		"authorization_nonce":                           hex.EncodeToString(task8Round5Digest("membership-authorization-nonce")),
		"database_identity_digest":                      hex.EncodeToString(databaseIdentityDigest),
		"environment_count":                             strconv.Itoa(len(members)),
		"environment_member_set_digest":                 hex.EncodeToString(environmentMemberSetDigest),
		"expires_at":                                    timeText(membershipExpiresAt),
		"final_environment_inventory_anchor_set_digest": hex.EncodeToString(environmentAnchorDigest),
		"final_environment_inventory_digest":            hex.EncodeToString(environmentInventoryDigest),
		"final_inventory_sequence":                      "1",
		"installation_id":                               installationID.String(),
		"installation_kind":                             installationKind,
		"issued_at":                                     timeText(membershipIssuedAt),
		"membership_retirement_id":                      uuid.NewSHA1(uuid.NameSpaceOID, []byte("task8-round5-membership")).String(),
		"migration_latch_digest":                        hex.EncodeToString(migrationLatchDigest),
		"release_scope":                                 releaseScope,
		"request_nonce":                                 hex.EncodeToString(membershipRequestNonce),
	}
	membershipRequestDigest := task8Round5DomainDigest(
		"retire-environment-inventory-membership-request.v1",
		task8Round5JCS(t, membershipRequest),
	)
	if config.corruptMembershipDigest {
		membershipRequestDigest = task8Round5Digest("corrupt-membership-request")
	}
	membershipResponse := make(map[string]any, len(membershipRequest)+4)
	for key, value := range membershipRequest {
		membershipResponse[key] = value
	}
	membershipResponse["request_digest"] = hex.EncodeToString(membershipRequestDigest)
	membershipResponse["membership_retirement_tombstone_id"] = uuid.NewSHA1(uuid.NameSpaceOID, []byte("task8-round5-membership-tombstone")).String()
	membershipResponse["phase"] = "inventory_membership_retired"
	membershipResponse["retired_at"] = timeText(retiredAt)
	membership := task8Round5Sign(t, "environment-inventory-membership-retirement.v1", "release_deployment_operator", config.signatureAlgorithm, config.signaturePolicyVersion, task8Round5JCS(t, membershipResponse))

	evidence := []task8Round5SignedEvidence{membership}
	for pairIndex := range pairs {
		pair := &pairs[pairIndex]
		pairMemberWires := make([]map[string]any, len(pair.members))
		for index, value := range pair.members {
			pairMemberWires[index] = value.wire()
		}
		pairMembersJCS := task8Round5JCS(t, pairMemberWires)
		memberSetDigest := task8Round5DomainDigest("provider-protocol-retirement-member-set.v1", pairMembersJCS)
		adminBody := task8Round5JCS(t, map[string]any{
			"authorization_id":                                   uuid.NewSHA1(uuid.NameSpaceOID, append([]byte("task8-round5-admin:"), pair.identityDigest...)).String(),
			"retirement_id":                                      pair.retirementID.String(),
			"installation_id":                                    installationID.String(),
			"installation_kind":                                  installationKind,
			"migration_latch_digest":                             hex.EncodeToString(migrationLatchDigest),
			"database_identity_digest":                           hex.EncodeToString(databaseIdentityDigest),
			"release_scope":                                      releaseScope,
			"environment_inventory_digest":                       hex.EncodeToString(environmentInventoryDigest),
			"environment_inventory_anchor_set_digest":            hex.EncodeToString(environmentAnchorDigest),
			"environment_inventory_membership_retirement_digest": hex.EncodeToString(membership.digest),
			"provider_identity_digest":                           hex.EncodeToString(pair.identityDigest),
			"provider_endpoint_identity_digest":                  hex.EncodeToString(pair.endpointDigest),
			"retirement_member_set_digest":                       hex.EncodeToString(memberSetDigest),
			"member_count":                                       strconv.Itoa(len(pair.members)),
			"authorization_scope":                                "permanent_disposable_namespace_retirement",
			"authorization_nonce":                                hex.EncodeToString(task8Round5Digest(fmt.Sprintf("admin-nonce-%d", pairIndex))),
			"issued_at":                                          timeText(adminIssuedAt),
			"expires_at":                                         timeText(adminExpiresAt),
		})
		pair.admin = task8Round5Sign(t, "provider-protocol-downgrade-retirement-authorization.v1", "provider_downgrade_retirement_admin", config.signatureAlgorithm, config.signaturePolicyVersion, adminBody)
		evidence = append(evidence, pair.admin)

		namespaces := make([]string, 0, len(pair.members))
		for _, value := range pair.members {
			if len(namespaces) == 0 || namespaces[len(namespaces)-1] != value.providerNamespace {
				namespaces = append(namespaces, value.providerNamespace)
			}
		}
		sort.Strings(namespaces)
		for _, namespace := range namespaces {
			historyBody := map[string]any{
				"provider_identity_digest":          hex.EncodeToString(pair.identityDigest),
				"provider_endpoint_identity_digest": hex.EncodeToString(pair.endpointDigest),
				"namespace":                         namespace,
				"observed_profiles":                 []string{"claim_v1", "legacy_v6"},
				"observed_at":                       timeText(historyObservedAt),
				"expires_at":                        timeText(historyExpiresAt),
			}
			for _, field := range []string{
				"unknown_profile_count", "unknown_record_count", "legacy_reservation_count", "legacy_terminal_count", "legacy_epoch_transition_count", "legacy_cutover_count",
				"claim_v1_credential_policy_count", "claim_v1_accepted_key_mutation_count", "incarnation_registration_count", "genesis_preparation_count",
				"genesis_completion_count", "genesis_release_preparation_count", "genesis_open_count", "claim_v1_reservation_count", "claim_v1_terminal_count",
				"claim_v1_epoch_transition_count", "claim_v1_epoch_recovery_count", "serving_lease_event_count", "runtime_rebind_count", "timeline_lineage_event_count",
				"staging_exclusion_count", "staging_recovery_count", "source_retirement_count", "genesis_authorizing_control_count", "other_mutation_count",
				"provider_history_high_water",
			} {
				historyBody[field] = "0"
			}
			history := task8Round5Sign(t, "provider-protocol-history-zero-projection.v1", "claim_v1_provider_history_auditor", config.signatureAlgorithm, config.signaturePolicyVersion, task8Round5JCS(t, historyBody))
			pair.historyByNS[namespace] = history
			evidence = append(evidence, history)
		}

		requestNonce := task8Round5Digest(fmt.Sprintf("provider-request-nonce-%d", pairIndex))
		retirementNonce := task8Round5Digest(fmt.Sprintf("provider-retirement-nonce-%d", pairIndex))
		providerRequest := map[string]any{
			"environment_inventory_anchor_set_digest":            hex.EncodeToString(environmentAnchorDigest),
			"environment_inventory_digest":                       hex.EncodeToString(environmentInventoryDigest),
			"environment_inventory_membership_retirement_digest": hex.EncodeToString(membership.digest),
			"installation_id":                                    installationID.String(),
			"member_count":                                       strconv.Itoa(len(pair.members)),
			"members":                                            json.RawMessage(pairMembersJCS),
			"provider_downgrade_retirement_authorization_digest": hex.EncodeToString(pair.admin.digest),
			"provider_endpoint_identity_digest":                  hex.EncodeToString(pair.endpointDigest),
			"provider_identity_digest":                           hex.EncodeToString(pair.identityDigest),
			"release_scope":                                      releaseScope,
			"request_nonce":                                      hex.EncodeToString(requestNonce),
			"retirement_id":                                      pair.retirementID.String(),
			"retirement_member_set_digest":                       hex.EncodeToString(memberSetDigest),
			"retirement_nonce":                                   hex.EncodeToString(retirementNonce),
		}
		providerRequestDigest := task8Round5DomainDigest("retire-provider-protocol-for-downgrade-request.v1", task8Round5JCS(t, providerRequest))
		if config.corruptProviderDigest && pairIndex == 0 {
			providerRequestDigest = task8Round5Digest("corrupt-provider-request")
		}
		namespaceStates := make([]map[string]any, len(namespaces))
		for namespaceIndex, namespace := range namespaces {
			var environmentDigests, databaseDigests []string
			for _, value := range pair.members {
				if value.providerNamespace == namespace {
					environmentDigests = append(environmentDigests, hex.EncodeToString(value.environmentRecordDigest))
					databaseDigests = append(databaseDigests, hex.EncodeToString(value.databaseIdentityDigest))
				}
			}
			namespaceStates[namespaceIndex] = map[string]any{
				"namespace":                       namespace,
				"environment_record_digests":      environmentDigests,
				"database_identity_digests":       databaseDigests,
				"history_zero_projection_digest":  hex.EncodeToString(pair.historyByNS[namespace].digest),
				"pre_retirement_control_sequence": "0",
				"provider_history_high_water":     "0",
				"retirement_tombstone_id":         uuid.NewSHA1(uuid.NameSpaceOID, []byte("task8-round5-tombstone:"+namespace+":"+hex.EncodeToString(pair.identityDigest))).String(),
			}
		}
		finalBody := make(map[string]any, len(providerRequest)+5)
		for key, value := range providerRequest {
			finalBody[key] = value
		}
		finalBody["request_digest"] = hex.EncodeToString(providerRequestDigest)
		finalBody["provider_control_sequence"] = "1"
		finalBody["namespace_states"] = namespaceStates
		finalBody["phase"] = "down_retired"
		finalBody["retired_at"] = timeText(retiredAt)
		pair.final = task8Round5Sign(t, "provider-protocol-downgrade-retirement.v1", "claim_v1_provider", config.signatureAlgorithm, config.signaturePolicyVersion, task8Round5JCS(t, finalBody))
		evidence = append(evidence, pair.final)
	}

	retirements := make([]map[string]any, len(pairs))
	for index, pair := range pairs {
		retirements[index] = map[string]any{
			"provider_endpoint_identity_digest":             hex.EncodeToString(pair.endpointDigest),
			"provider_identity_digest":                      hex.EncodeToString(pair.identityDigest),
			"provider_protocol_downgrade_retirement_digest": hex.EncodeToString(pair.final.digest),
		}
	}
	retirementSetBody := task8Round5JCS(t, map[string]any{
		"created_at": timeText(transactionTime.Add(-11 * time.Minute)),
		"environment_inventory_anchor_set_digest":            hex.EncodeToString(environmentAnchorDigest),
		"environment_inventory_digest":                       hex.EncodeToString(environmentInventoryDigest),
		"environment_inventory_membership_retirement_digest": hex.EncodeToString(membership.digest),
		"installation_id":                                    installationID.String(),
		"provider_count":                                     strconv.Itoa(len(pairs)),
		"release_scope":                                      releaseScope,
		"retired_member_count":                               strconv.Itoa(len(members)),
		"retired_members":                                    memberWires,
		"retirements":                                        retirements,
	})
	retirementSetDigest := task8Round5DomainDigest("provider-protocol-downgrade-retirement-set.v1", retirementSetBody)
	evidenceBundle := task8Round5EvidenceBundle(t, retirementSetDigest, evidence)

	inventory := make([]map[string]any, len(task8PristineDownRegistry))
	for index, relation := range task8PristineDownRegistry {
		inventory[index] = map[string]any{
			"classification": relation.classification,
			"content_digest": task8PristineEmptyTableDigests[index],
			"row_count":      "0",
			"table_name":     relation.table,
		}
	}
	transactionNonce := task8Round5Digest("transaction-nonce")
	manifestID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("task8-round5-manifest"))
	pristineBody := task8Round5JCS(t, map[string]any{
		"control_table_count":            "2",
		"current_catalog_digest":         hex.EncodeToString(currentCatalogDigest),
		"database_identity_digest":       hex.EncodeToString(databaseIdentityDigest),
		"database_transaction_id":        databaseTransactionID,
		"downgrade_authorization_count":  "0",
		"installation_id":                installationID.String(),
		"installation_kind":              installationKind,
		"manifest_digest":                hex.EncodeToString(task8Round5Digest("manifest")),
		"manifest_id":                    manifestID.String(),
		"migration_latch_count":          "1",
		"migration_latch_digest":         hex.EncodeToString(migrationLatchDigest),
		"non_control_protocol_row_count": "0",
		"observed_at":                    timeText(transactionTime),
		"stable_table_count":             strconv.Itoa(len(inventory)),
		"stable_table_inventory":         inventory,
		"stage":                          "pre_authorization",
		"transaction_nonce":              hex.EncodeToString(transactionNonce),
	})
	pristineDigest := task8Round5DomainDigest("pristine-downgrade-inventory.v1", pristineBody)
	authorizationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("task8-round5-authorization"))
	authorizationIssuedAt := transactionTime.Add(-30 * time.Second)
	authorizationExpiresAt := transactionTime.Add(time.Minute)
	authorizationBody := task8Round5JCS(t, map[string]any{
		"authorization_id":                                  authorizationID.String(),
		"authorization_scope":                               "down_00007_only",
		"current_catalog_digest":                            hex.EncodeToString(currentCatalogDigest),
		"database_identity_digest":                          hex.EncodeToString(databaseIdentityDigest),
		"database_transaction_id":                           databaseTransactionID,
		"environment_inventory_anchor_set_digest":           hex.EncodeToString(environmentAnchorDigest),
		"expires_at":                                        timeText(authorizationExpiresAt),
		"installation_id":                                   installationID.String(),
		"issued_at":                                         timeText(authorizationIssuedAt),
		"migration_latch_digest":                            hex.EncodeToString(migrationLatchDigest),
		"migration_version":                                 "7",
		"pristine_downgrade_inventory_digest":               hex.EncodeToString(pristineDigest),
		"provider_protocol_downgrade_retirement_set_digest": hex.EncodeToString(retirementSetDigest),
		"transaction_nonce":                                 hex.EncodeToString(transactionNonce),
	})
	authorization := task8Round5Sign(t, "authority-protocol-downgrade-authorization.v1", "authority_protocol_downgrade_authorizer", config.signatureAlgorithm, config.signaturePolicyVersion, authorizationBody)
	authorizedStateBody := task8Round5JCS(t, map[string]any{
		"database_transaction_id":                           databaseTransactionID,
		"downgrade_authorization_count":                     "1",
		"downgrade_authorization_digest":                    hex.EncodeToString(authorization.digest),
		"migration_latch_count":                             "1",
		"migration_latch_digest":                            hex.EncodeToString(migrationLatchDigest),
		"non_control_protocol_row_count":                    "0",
		"observed_at":                                       timeText(transactionTime),
		"pristine_downgrade_inventory_digest":               hex.EncodeToString(pristineDigest),
		"provider_protocol_downgrade_retirement_set_digest": hex.EncodeToString(retirementSetDigest),
		"stage":             "authorization_inserted",
		"transaction_nonce": hex.EncodeToString(transactionNonce),
	})

	return task8Round5RawInsertFixture{
		arguments: []any{
			authorizationID, installationID, migrationLatchDigest, databaseIdentityDigest, int64(7), currentCatalogDigest,
			pristineDigest, retirementSetDigest, environmentAnchorDigest, databaseTransactionID, transactionNonce,
			"down_00007_only", authorizationIssuedAt, authorizationExpiresAt, authorizationBody, authorization.envelope,
			authorization.digest, pristineBody, retirementSetBody, evidenceBundle,
		},
		authorizedStateBody:   authorizedStateBody,
		authorizedStateDigest: task8Round5DomainDigest("pristine-downgrade-authorized-state.v1", authorizedStateBody),
	}
}

func task8AssertAuthorityV7Round5RawRetirementMatrix(t *testing.T, database *sql.DB) {
	t.Helper()
	const insertCall = `
SELECT * FROM nodecontrol.v7_insert_downgrade_authorization(
 $1::uuid,$2::uuid,$3::bytea,$4::bytea,$5::bigint,$6::bytea,$7::bytea,$8::bytea,$9::bytea,$10::numeric,
 $11::bytea,$12::text,$13::timestamptz,$14::timestamptz,$15::bytea,$16::bytea,$17::bytea,$18::bytea,$19::bytea,$20::bytea)`
	for _, testCase := range []struct {
		name             string
		config           task8Round5RawFixtureConfig
		wantSuccess      bool
		wantErrorMessage string
	}{
		{
			name:        "P2 N3 multi namespace exact cover",
			config:      task8Round5RawFixtureConfig{multiNamespace: true, signatureAlgorithm: "ed25519", signaturePolicyVersion: "1"},
			wantSuccess: true,
		},
		{
			name:        "P-256 rotated policy metadata",
			config:      task8Round5RawFixtureConfig{signatureAlgorithm: "ecdsa-p256-sha256", signaturePolicyVersion: "2"},
			wantSuccess: true,
		},
		{
			name:        "historical expired evidence valid at retirement",
			config:      task8Round5RawFixtureConfig{historicalEvidence: true, signatureAlgorithm: "ed25519", signaturePolicyVersion: "1"},
			wantSuccess: true,
		},
		{
			name:        "six-key member order ignores timeline",
			config:      task8Round5RawFixtureConfig{inverseTimelineOrder: true, signatureAlgorithm: "ed25519", signaturePolicyVersion: "1"},
			wantSuccess: true,
		},
		{
			name:             "membership response request digest mismatch",
			config:           task8Round5RawFixtureConfig{corruptMembershipDigest: true, signatureAlgorithm: "ed25519", signaturePolicyVersion: "1"},
			wantErrorMessage: "authority v7 membership-retirement request digest binding mismatch",
		},
		{
			name:             "provider response request digest mismatch",
			config:           task8Round5RawFixtureConfig{corruptProviderDigest: true, signatureAlgorithm: "ed25519", signaturePolicyVersion: "1"},
			wantErrorMessage: "authority v7 provider-retirement request digest binding mismatch",
		},
	} {
		t.Run("raw SQL retirement "+testCase.name, func(t *testing.T) {
			tx, err := database.BeginTx(t.Context(), nil)
			if err != nil {
				t.Fatal("begin raw retirement transaction:", err)
			}
			defer tx.Rollback()
			fixture := task8BuildRound5RawInsertFixture(t, tx, testCase.config)
			_, callErr := tx.ExecContext(t.Context(), insertCall, fixture.arguments...)
			if testCase.wantSuccess {
				if callErr != nil {
					var postgresError *pgconn.PgError
					if errors.As(callErr, &postgresError) {
						t.Errorf("spec-valid raw retirement fixture rejected: code=%s message=%s detail=%s where=%s internal=%s", postgresError.Code, postgresError.Message, postgresError.Detail, postgresError.Where, postgresError.InternalQuery)
					} else {
						t.Errorf("spec-valid raw retirement fixture rejected: %v", callErr)
					}
				} else {
					var latchCount, authorizationCount int
					if err := tx.QueryRowContext(t.Context(), `
SELECT (SELECT count(*) FROM nodecontrol.control_plane_authority_protocol_migration_latches),
       (SELECT count(*) FROM nodecontrol.control_plane_authority_protocol_downgrade_authorizations)`).Scan(&latchCount, &authorizationCount); err != nil {
						t.Fatal("inspect raw retirement insert cardinality:", err)
					}
					if latchCount != 1 || authorizationCount != 1 {
						t.Errorf("raw retirement insert cardinality=%d/%d, want 1/1", latchCount, authorizationCount)
					}
				}
			} else {
				var postgresError *pgconn.PgError
				if !errors.As(callErr, &postgresError) || postgresError.Code != "22023" || postgresError.Message != testCase.wantErrorMessage {
					if postgresError != nil {
						t.Errorf("raw retirement rejection code=%s message=%s detail=%s where=%s internal=%s, want PostgreSQL 22023/%q", postgresError.Code, postgresError.Message, postgresError.Detail, postgresError.Where, postgresError.InternalQuery, testCase.wantErrorMessage)
					} else {
						t.Errorf("raw retirement rejection=%v, want PostgreSQL 22023/%q", callErr, testCase.wantErrorMessage)
					}
				}
			}
			if err := tx.Rollback(); err != nil {
				t.Fatal("roll back raw retirement transaction:", err)
			}
		})
	}
}

func TestNodeControlV7ACLAndGuardCatalog(t *testing.T) {
	databaseURL := os.Getenv("TALENRO_DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("TALENRO_DATABASE_URL is required")
	}
	database, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	task8AssertAuthorityV7RejectsPreexistingRoles(t, database)
	manifest, err := decodeNodeControlManifest(task8ReadFile(t, "../../db/schema/nodecontrol.v1.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	manifest = final51Manifest(manifest)
	oraclePool, err := pgxpool.New(t.Context(), databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer oraclePool.Close()
	assertAuthorityV7FunctionsMatchManifest(t.Context(), t, oraclePool, manifest.Functions)
	task8AssertAuthorityV7Round5ProductionDefinitions(t, database)
	task8AssertAuthorityV7RoleAttributeDrift(t, database)
	task8AssertAuthorityV7Round5RawRetirementMatrix(t, database)
	task8AssertAuthorityV7ExactACLs(t, database)
	task8AssertAuthorityV7OrdinaryDenial(t, database)
	task8AssertDownFunctionLockBarrier(t, database)

	wantRoles := []string{
		"nodecontrol_migration_downgrader",
		"nodecontrol_staging_importer",
		"nodecontrol_upgrade_executor",
	}
	rows, err := database.QueryContext(t.Context(), `
SELECT rolname
FROM pg_catalog.pg_roles
WHERE rolname = ANY($1::text[]) AND NOT rolcanlogin
ORDER BY rolname COLLATE "C"`, wantRoles)
	if err != nil {
		t.Fatal("query authority-v7 roles:", err)
	}
	var gotRoles []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatal(err)
		}
		gotRoles = append(gotRoles, name)
	}
	rows.Close()
	if strings.Join(gotRoles, "\n") != strings.Join(wantRoles, "\n") {
		t.Fatalf("NOLOGIN authority-v7 roles = %v, want %v", gotRoles, wantRoles)
	}
	var memberships int
	if err := database.QueryRowContext(t.Context(), `
SELECT count(*)
FROM pg_catalog.pg_auth_members m
JOIN pg_catalog.pg_roles member ON member.oid=m.member
JOIN pg_catalog.pg_roles role ON role.oid=m.roleid
WHERE member.rolname = ANY($1::text[]) OR role.rolname = ANY($1::text[])`, wantRoles).Scan(&memberships); err != nil {
		t.Fatal("query authority-v7 role memberships:", err)
	}
	if memberships != 0 {
		t.Fatalf("authority-v7 role membership count = %d, want 0", memberships)
	}

	wantHelpers := []string{
		"begin_staging_import",
		"v7_acquire_source_freeze_for_seal",
		"v7_assert_activation_barrier",
		"v7_assert_source_writable",
		"v7_authority_proof_group_valid",
		"v7_consume_down_guard",
		"v7_guard_authority_proof_transition",
		"v7_insert_downgrade_authorization",
		"v7_reject_immutable_mutation",
		"v7_require_role",
		"v7_source_is_frozen",
		"v7_text_array_is_sorted_unique",
	}
	sort.Strings(wantHelpers)
	rows, err = database.QueryContext(t.Context(), `
SELECT p.proname
FROM pg_catalog.pg_proc p
JOIN pg_catalog.pg_namespace n ON n.oid=p.pronamespace
WHERE n.nspname='nodecontrol' AND (left(p.proname,3)='v7_' OR p.proname='begin_staging_import')
ORDER BY p.proname COLLATE "C"`)
	if err != nil {
		t.Fatal("query authority-v7 helpers:", err)
	}
	var gotHelpers []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatal(err)
		}
		gotHelpers = append(gotHelpers, name)
	}
	rows.Close()
	if strings.Join(gotHelpers, "\n") != strings.Join(wantHelpers, "\n") {
		t.Fatalf("authority-v7 helper registry = %v, want %v", gotHelpers, wantHelpers)
	}

	rows, err = database.QueryContext(t.Context(), `
SELECT p.proname, p.prosecdef, coalesce(array_to_string(p.proconfig, ','), ''), has_function_privilege('public', p.oid, 'EXECUTE')
FROM pg_catalog.pg_proc p
JOIN pg_catalog.pg_namespace n ON n.oid=p.pronamespace
WHERE n.nspname='nodecontrol' AND p.proname IN ('v7_require_role','v7_acquire_source_freeze_for_seal','v7_insert_downgrade_authorization','v7_consume_down_guard','begin_staging_import')
ORDER BY p.proname COLLATE "C"`)
	if err != nil {
		t.Fatal("query authority-v7 restricted helper ACLs:", err)
	}
	checked := 0
	for rows.Next() {
		var name, config string
		var securityDefiner, publicExecute bool
		if err := rows.Scan(&name, &securityDefiner, &config, &publicExecute); err != nil {
			t.Fatal(err)
		}
		checked++
		wantSecurityDefiner := name != "v7_require_role"
		if securityDefiner != wantSecurityDefiner || config != "search_path=pg_catalog, nodecontrol" || publicExecute {
			t.Errorf("restricted helper %s metadata = security_definer:%t config:%q public_execute:%t", name, securityDefiner, config, publicExecute)
		}
	}
	rows.Close()
	if checked != 5 {
		t.Fatalf("restricted helper ACL row count = %d, want 5", checked)
	}

	for _, role := range wantRoles {
		var mayInsert bool
		if err := database.QueryRowContext(t.Context(), `SELECT has_table_privilege($1, 'nodecontrol.control_plane_authority_protocol_migration_latches', 'INSERT')`, role).Scan(&mayInsert); err != nil {
			t.Fatal("query direct immutable-table privilege:", err)
		}
		if mayInsert {
			t.Errorf("role %s has direct mutation privilege on the migration latch", role)
		}
	}
}
