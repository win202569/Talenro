//go:build integration

package store_test

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gowebpki/jcs"
	"github.com/jackc/pgx/v5/pgconn"
	migrations "talenro.local/platform/db/migrations"
	"talenro.local/platform/internal/nodecontrol/authority"
	"talenro.local/platform/internal/nodecontrol/contracts"
)

func TestNodeControlV7DownRejectMatrix(t *testing.T) {
	// This literal 49-relation registry is shared only with the independent
	// catalog test; every caller inventory that is not its exact projection is
	// rejected before destructive DDL.
	if len(task8PristineDownRegistry) != 49 {
		t.Fatalf("pristine Down relation registry count = %d, want 49", len(task8PristineDownRegistry))
	}
	database, migrationFS, _ := openOwnedAuthorityV7Database(t)
	applyAuthorityV7Base(t, database, migrationFS)
	const matrixLabel = "down-real-49-snapshot-matrix"
	installAuthorityV7Fixture(t, database, migrationFS, matrixLabel)
	t.Run("unfrozen source guard accepts non inventory row shape", func(t *testing.T) {
		tx, err := database.BeginTx(t.Context(), nil)
		if err != nil {
			t.Fatal("begin source guard regression transaction:", err)
		}
		defer tx.Rollback()
		var replicationRole string
		var frozen bool
		if err := tx.QueryRowContext(t.Context(), `SELECT current_setting('session_replication_role'), nodecontrol.v7_source_is_frozen()`).Scan(&replicationRole, &frozen); err != nil {
			t.Fatal("inspect source guard regression preconditions:", err)
		}
		if replicationRole != "origin" || frozen {
			t.Fatalf("source guard regression preconditions = role %q, frozen %t; want origin/false", replicationRole, frozen)
		}
		now := time.Date(2026, time.August, 29, 17, 59, 59, 0, time.UTC)
		if _, err := tx.ExecContext(t.Context(), `INSERT INTO nodecontrol.node_pops(pop_code,iso_country,region,operator_state,created_at,updated_at) VALUES('source-shape-regression','US','source-shape-region','enabled',$1,$1)`, now); err != nil {
			t.Fatal("insert unfrozen non-inventory source row:", err)
		}
		result, err := tx.ExecContext(t.Context(), `UPDATE nodecontrol.node_pops SET operator_state='disabled',updated_at=$1 WHERE pop_code='source-shape-regression'`, now.Add(time.Second))
		if err != nil {
			t.Fatal("update unfrozen non-inventory source row:", err)
		}
		if rows, err := result.RowsAffected(); err != nil || rows != 1 {
			t.Fatalf("updated unfrozen non-inventory source rows=%d, err=%v; want 1/nil", rows, err)
		}
		result, err = tx.ExecContext(t.Context(), `DELETE FROM nodecontrol.node_pops WHERE pop_code='source-shape-regression'`)
		if err != nil {
			t.Fatal("delete unfrozen non-inventory source row:", err)
		}
		if rows, err := result.RowsAffected(); err != nil || rows != 1 {
			t.Fatalf("deleted unfrozen non-inventory source rows=%d, err=%v; want 1/nil", rows, err)
		}
	})
	matrixNow := time.Date(2026, time.August, 29, 18, 0, 0, 0, time.UTC)
	matrixTx, err := database.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatal("begin one-connection 49-relation Down matrix:", err)
	}
	defer matrixTx.Rollback()
	// Fixture validity is proved independently before any production Down call;
	// a bad INSERT therefore cannot impersonate a relation-specific RED.
	task8PreflightLiteralDownRelationSeeds(t, matrixTx)
	for relationIndex, pair := range task8PristineDownRegistry {
		pair := pair
		t.Run("actual snapshot mismatch "+pair.table, func(t *testing.T) {
			relationSavepoint := fmt.Sprintf("task8_relation_%02d", relationIndex)
			if _, err := matrixTx.ExecContext(t.Context(), "SAVEPOINT "+relationSavepoint); err != nil {
				t.Fatal("create relation seed savepoint:", err)
			}
			defer func() {
				_, _ = matrixTx.ExecContext(context.Background(), "ROLLBACK TO SAVEPOINT "+relationSavepoint)
				_, _ = matrixTx.ExecContext(context.Background(), "RELEASE SAVEPOINT "+relationSavepoint)
			}()
			// This is deliberately a target-only dangling-FK physical fixture, not
			// a purported legal domain graph. Every statement is literal and every
			// NOT NULL/CHECK remains live; only FK/ordinary triggers are isolated.
			task8InsertLiteralTargetOnlyRelation(t, matrixTx, task8LiteralDownRelationSeeds[relationIndex])
			var targetRows int64
			if err := matrixTx.QueryRowContext(t.Context(), "SELECT count(*) FROM "+pair.table).Scan(&targetRows); err != nil {
				t.Fatal("count isolated target relation:", err)
			}
			if targetRows != 1 {
				t.Fatalf("isolated target %s row count=%d, want exactly 1", pair.table, targetRows)
			}
			facts := disposableAuthorityV7DownFacts(matrixNow, matrixLabel)
			body, err := json.Marshal(struct {
				Schema    string                                            `json:"schema"`
				Inventory []contracts.AuthorityV7StableTableInventoryItemV1 `json:"inventory"`
			}{Schema: "PristineDowngradeInventoryV1", Inventory: facts.StableTableInventory})
			if err != nil {
				t.Fatal(err)
			}
			inventoryDigest := sha256.Sum256(body)
			before := task8DownProtectedSnapshot(t, matrixTx)
			callSavepoint := fmt.Sprintf("task8_call_%02d", relationIndex)
			if _, err := matrixTx.ExecContext(t.Context(), "SAVEPOINT "+callSavepoint); err != nil {
				t.Fatal("create restricted-call savepoint:", err)
			}
			_, executionErr := matrixTx.ExecContext(t.Context(), `
SELECT * FROM nodecontrol.v7_insert_downgrade_authorization(
 $1,$2,$3,$4,7,$5,$6,$7,$8,txid_current(),$9,'down_00007_only',$10,$11,$12,$13,$14,$15,$16,$17)`,
				facts.AuthorizationID, facts.MigrationLatch.InstallationID, facts.MigrationLatchDigest[:], facts.MigrationLatch.DatabaseIdentityDigest[:],
				facts.CurrentCatalogDigest[:], inventoryDigest[:], facts.ProviderRetirementSetDigest[:], facts.EnvironmentAnchorSetDigest[:],
				facts.TransactionNonce[:], matrixNow, matrixNow.Add(time.Minute), []byte(`{"schema":"AuthorityV7DownAuthorizationV1"}`),
				[]byte(`{"schema":"SignedEnvelopeV1"}`), task8LegacyDigest(0x91), body,
				[]byte(`{"schema":"AuthorityV7ProviderRetirementSetFactsV1"}`), []byte(`{"schema":"CanonicalEvidenceBundleV1"}`))
			if _, err := matrixTx.ExecContext(t.Context(), "ROLLBACK TO SAVEPOINT "+callSavepoint); err != nil {
				t.Fatal("restore transaction after expected restricted rejection:", err)
			}
			if _, err := matrixTx.ExecContext(t.Context(), "RELEASE SAVEPOINT "+callSavepoint); err != nil {
				t.Fatal("release restricted-call savepoint:", err)
			}
			if executionErr == nil {
				t.Fatalf("restricted Down insert accepted pristine row_count=0 over real %s row", pair.table)
			}
			var postgresError *pgconn.PgError
			if !errors.As(executionErr, &postgresError) || postgresError.Code == "42883" ||
				(!strings.Contains(strings.ToLower(postgresError.Message), "inventory") && !strings.Contains(strings.ToLower(postgresError.Message), "pristine")) ||
				!strings.Contains(strings.ToLower(postgresError.Message), strings.TrimPrefix(pair.table, "nodecontrol.")) {
				t.Errorf("restricted Down relation %s rejection seam = %#v, want registered 20-input target-specific actual-inventory rejection", pair.table, postgresError)
			}
			if after := task8DownProtectedSnapshot(t, matrixTx); !bytes.Equal(after, before) {
				t.Fatalf("relation-specific rejected Down changed protected state\nbefore=%s\nafter=%s", before, after)
			}
		})
	}
	if err := matrixTx.Rollback(); err != nil {
		t.Fatal("rollback 49-relation Down matrix:", err)
	}

	// Deterministic current production break: a real base row is present while
	// the caller reports node_pops row_count=0. The current 19/4 SQL path does not
	// bind the caller inventory and proceeds to destructive Down.
	const label = "down-actual-snapshot-mismatch"
	// The 49-way matrix above intentionally leaves the one v7 fixture intact;
	// the explicit real-row case uses that same database and matching latch.
	now := time.Date(2026, time.August, 29, 18, 30, 0, 0, time.UTC)
	if _, err := database.ExecContext(t.Context(), `INSERT INTO nodecontrol.node_pops(pop_code,iso_country,region,operator_state,created_at,updated_at) VALUES('down-mismatch','US','down-region','enabled',$1,$1)`, now); err != nil {
		t.Fatal(err)
	}
	before := task8DownProtectedSnapshot(t, database)
	facts := disposableAuthorityV7DownFacts(now, matrixLabel)
	var nodePOPs *contracts.AuthorityV7StableTableInventoryItemV1
	for index := range facts.StableTableInventory {
		if facts.StableTableInventory[index].TableName == "nodecontrol.node_pops" {
			nodePOPs = &facts.StableTableInventory[index]
		}
	}
	if nodePOPs == nil || nodePOPs.RowCount != 0 {
		t.Fatalf("negative fixture is not literal caller node_pops row_count=0: %#v", facts.StableTableInventory)
	}
	grant, err := task8NewDisposableAuthorityV7DownGrant(t, facts)
	if err != nil {
		t.Fatal(err)
	}
	ctx := migrations.WithAuthorityV7MigrationContext(t.Context(), migrations.AuthorityV7MigrationContext{
		InstallationKind: migrations.InstallationKindDisposableFixture,
		DownGrant:        &grant,
		DownAuthorizer:   &fixtureAuthorityV7DownAuthorizer{},
	})
	if result, err := newAuthorityV7Provider(t, database, migrationFS).Down(ctx); err == nil {
		t.Fatalf("Down accepted caller row_count=0 over a real node_pops row and ran destructive version %d", result.Source.Version)
	}
	after := task8DownProtectedSnapshot(t, database)
	if !bytes.Equal(after, before) {
		t.Fatalf("rejected Down changed protected catalog/rows\nbefore=%s\nafter=%s", before, after)
	}
}

type task8LiteralDownRelationSeed struct {
	table     string
	insertSQL string
}

func task8InsertLiteralTargetOnlyRelation(t *testing.T, tx *sql.Tx, seed task8LiteralDownRelationSeed) string {
	t.Helper()
	if _, err := tx.ExecContext(t.Context(), `SET LOCAL session_replication_role=replica`); err != nil {
		t.Fatal("enter FK-isolated target-relation seed mode:", err)
	}
	var rowCTID string
	if err := tx.QueryRowContext(t.Context(), seed.insertSQL).Scan(&rowCTID); err != nil {
		t.Fatalf("literal CHECK-valid INSERT preflight for %s: %v\nSQL=%s", seed.table, err, seed.insertSQL)
	}
	if _, err := tx.ExecContext(t.Context(), `SET LOCAL session_replication_role=origin`); err != nil {
		t.Fatal("restore production trigger mode before restricted Down:", err)
	}
	var replicationRole string
	if err := tx.QueryRowContext(t.Context(), `SELECT current_setting('session_replication_role')`).Scan(&replicationRole); err != nil || replicationRole != "origin" {
		t.Fatalf("literal seed left session_replication_role=%q error=%v, want origin", replicationRole, err)
	}
	var targetCount int
	if err := tx.QueryRowContext(t.Context(), "SELECT count(*) FROM "+seed.table).Scan(&targetCount); err != nil {
		t.Fatalf("count literal target %s: %v", seed.table, err)
	}
	if targetCount != 1 || rowCTID == "" {
		t.Fatalf("literal target %s preflight count/ctid = %d/%q, want 1/nonempty", seed.table, targetCount, rowCTID)
	}
	return rowCTID
}

func task8PreflightLiteralDownRelationSeeds(t *testing.T, tx *sql.Tx) {
	t.Helper()
	if len(task8LiteralDownRelationSeeds) != len(task8PristineDownRegistry) {
		t.Fatalf("literal Down seed count = %d, want exact registry count %d", len(task8LiteralDownRelationSeeds), len(task8PristineDownRegistry))
	}
	seen := make(map[string]bool, len(task8LiteralDownRelationSeeds))
	for index, seed := range task8LiteralDownRelationSeeds {
		if seed.table != task8PristineDownRegistry[index].table {
			t.Fatalf("literal Down seed[%d] table = %q, want %q", index, seed.table, task8PristineDownRegistry[index].table)
		}
		if seen[seed.table] {
			t.Fatalf("duplicate literal Down seed %q", seed.table)
		}
		seen[seed.table] = true
		savepoint := fmt.Sprintf("literal_seed_preflight_%02d", index)
		if _, err := tx.ExecContext(t.Context(), "SAVEPOINT "+savepoint); err != nil {
			t.Fatal(err)
		}
		task8InsertLiteralTargetOnlyRelation(t, tx, seed)
		if _, err := tx.ExecContext(t.Context(), "ROLLBACK TO SAVEPOINT "+savepoint); err != nil {
			t.Fatal(err)
		}
		if _, err := tx.ExecContext(t.Context(), "RELEASE SAVEPOINT "+savepoint); err != nil {
			t.Fatal(err)
		}
	}
}

var task8LiteralDownRelationSeeds = [...]task8LiteralDownRelationSeed{
	{table: "nodecontrol.control_plane_authority_epoch_transition_applications", insertSQL: `INSERT INTO nodecontrol.control_plane_authority_epoch_transition_applications (transition_id,resolution_id,terminal_application_id,transition_intent_digest,provider_transition_request_digest,provider_transition_digest,previous_epoch,next_epoch,previous_epoch_transition_chain_digest,next_epoch_transition_chain_digest,previous_epoch_transition_terminal_chain_digest,expected_provider_head_digest,pending_provider_head_digest,provider_control_sequence,current_database_identity_digest,database_timeline_lineage_chain_digest,current_database_incarnation_registration_digest,runtime_rebind_chain_digest,serving_lease_digest,provider_prepared_at,recorded_database_point,recorded_at,canonical_evidence_bundle_jcs,canonical_body_jcs,body_digest) VALUES ('11111111-1111-4111-8111-000000000001'::uuid,'11111111-1111-4111-8111-000000000001'::uuid,'11111111-1111-4111-8111-000000000001'::uuid,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),1,2,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),0,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'2026-08-23T12:00:00Z'::timestamptz,'{}'::bytea,'2026-08-23T12:00:00Z'::timestamptz,'{}'::bytea,'{}'::bytea,decode('1111111111111111111111111111111111111111111111111111111111111111','hex')) RETURNING ctid::text`},
	{table: "nodecontrol.control_plane_authority_epoch_transition_cancellations", insertSQL: `INSERT INTO nodecontrol.control_plane_authority_epoch_transition_cancellations (cancellation_id,terminal_application_id,transition_id,resolution_id,transition_intent_digest,provider_rebind_digest,runtime_rebind_result_digest,cancellation_kind,preserved_epoch,preserved_epoch_transition_chain_digest,previous_epoch_transition_terminal_chain_digest,resulting_epoch_transition_terminal_chain_digest,current_database_identity_digest,database_timeline_lineage_chain_digest,status,database_transaction_id,transaction_snapshot_digest,database_point,cancelled_at,canonical_evidence_bundle_jcs,canonical_body_jcs,body_digest) VALUES ('11111111-1111-4111-8111-000000000002'::uuid,'11111111-1111-4111-8111-000000000002'::uuid,'11111111-1111-4111-8111-000000000002'::uuid,'11111111-1111-4111-8111-000000000002'::uuid,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'preparation_absent_rebind',1,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'cancelled_without_epoch_advance',1,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'{}'::bytea,'2026-08-23T12:00:00Z'::timestamptz,'{}'::bytea,'{}'::bytea,decode('1111111111111111111111111111111111111111111111111111111111111111','hex')) RETURNING ctid::text`},
	{table: "nodecontrol.control_plane_authority_epoch_transition_intents", insertSQL: `INSERT INTO nodecontrol.control_plane_authority_epoch_transition_intents (transition_id,resolution_id,cancellation_id,terminal_application_id,activation_id,previous_epoch,next_epoch,previous_transition_digest,previous_epoch_transition_terminal_chain_digest,reason,expected_provider_head_digest,pre_transition_database_authority_head_digest,current_database_identity_digest,database_timeline_lineage_chain_digest,database_transaction_id,transaction_snapshot_digest,database_point,created_at,canonical_evidence_bundle_jcs,canonical_body_jcs,body_digest) VALUES ('11111111-1111-4111-8111-000000000003'::uuid,'11111111-1111-4111-8111-000000000003'::uuid,'11111111-1111-4111-8111-000000000003'::uuid,'11111111-1111-4111-8111-000000000003'::uuid,'11111111-1111-4111-8111-000000000003'::uuid,1,2,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'scheduled_authority_rotation',decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),1,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'{}'::bytea,'2026-08-23T12:00:00Z'::timestamptz,'{}'::bytea,'{}'::bytea,decode('1111111111111111111111111111111111111111111111111111111111111111','hex')) RETURNING ctid::text`},
	{table: "nodecontrol.control_plane_authority_epoch_transition_recovery_applications", insertSQL: `INSERT INTO nodecontrol.control_plane_authority_epoch_transition_recovery_applications (recovery_application_id,recovery_intent_digest,recovery_transcript_id,provider_epoch_transition_recovery_digest,recovery_prefix_key_digest,epoch_transition_recovery_prefix_decision_digest,transition_id,resolution_id,cancellation_id,terminal_application_id,terminal_outcome,deferred_runtime_rebind_suffix_id,deferred_runtime_rebind_count,provider_tail_runtime_rebind_chain_digest,database_tail_runtime_rebind_chain_digest,post_rehydration_database_identity_digest,post_rehydration_database_timeline_lineage_chain_digest,source_database_commit_anchor_schema,source_database_commit_anchor_digest,pre_rehydration_database_authority_head_digest,rehydrated_row_count,rehydrated_rows,post_rehydration_database_authority_head_digest,database_transaction_id,transaction_snapshot_digest,database_point,rehydrated_at,canonical_evidence_bundle_jcs,canonical_body_jcs,body_digest) VALUES ('11111111-1111-4111-8111-000000000004'::uuid,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'11111111-1111-4111-8111-000000000004'::uuid,'11111111-1111-4111-8111-000000000004'::uuid,'11111111-1111-4111-8111-000000000004'::uuid,'11111111-1111-4111-8111-000000000004'::uuid,'resolved','11111111-1111-4111-8111-000000000004'::uuid,0,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'task8',decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),0,'[]'::jsonb,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),1,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'{}'::bytea,'2026-08-23T12:00:00Z'::timestamptz,'{}'::bytea,'{}'::bytea,decode('1111111111111111111111111111111111111111111111111111111111111111','hex')) RETURNING ctid::text`},
	{table: "nodecontrol.control_plane_authority_epoch_transition_recovery_intents", insertSQL: `INSERT INTO nodecontrol.control_plane_authority_epoch_transition_recovery_intents (recovery_intent_id,recovery_id,recovery_application_id,transition_id,resolution_id,cancellation_id,terminal_application_id,activation_id,expected_terminal_outcome,provider_terminal_inspect_digest,provider_terminal_runtime_rebind_chain_digest,deferred_runtime_rebind_suffix_id,expected_provider_head_digest,expected_provider_current_epoch,expected_provider_epoch_transition_chain_digest,expected_provider_epoch_transition_terminal_chain_digest,observed_transition_intent_row_count,observed_transition_application_row_count,observed_transition_resolution_row_count,observed_transition_cancellation_row_count,observed_terminal_application_row_count,pre_intent_database_authority_head_digest,database_rebind_gap_anchor_runtime_chain_digest,current_database_identity_digest,database_timeline_lineage_chain_digest,current_database_incarnation_registration_digest,runtime_rebind_chain_digest,runtime_instance_binding_digest,database_transaction_id,transaction_snapshot_digest,database_point,transaction_nonce,recovery_reason,created_at,canonical_evidence_bundle_jcs,canonical_body_jcs,body_digest) VALUES ('11111111-1111-4111-8111-000000000005'::uuid,'11111111-1111-4111-8111-000000000005'::uuid,'11111111-1111-4111-8111-000000000005'::uuid,'11111111-1111-4111-8111-000000000005'::uuid,'11111111-1111-4111-8111-000000000005'::uuid,'11111111-1111-4111-8111-000000000005'::uuid,'11111111-1111-4111-8111-000000000005'::uuid,'11111111-1111-4111-8111-000000000005'::uuid,'resolved',decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'11111111-1111-4111-8111-000000000005'::uuid,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),1,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),0,0,0,0,0,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),1,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'{}'::bytea,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'pitr_epoch_db_preimage_missing','2026-08-23T12:00:00Z'::timestamptz,'{}'::bytea,'{}'::bytea,decode('1111111111111111111111111111111111111111111111111111111111111111','hex')) RETURNING ctid::text`},
	{table: "nodecontrol.control_plane_authority_epoch_transition_recovery_prefix_decisions", insertSQL: `INSERT INTO nodecontrol.control_plane_authority_epoch_transition_recovery_prefix_decisions (decision_id,recovery_prefix_key_digest,activation_id,recovery_id,recovery_intent_digest,recovery_transcript_id,provider_epoch_transition_recovery_digest,deferred_runtime_rebind_suffix_id,deferred_runtime_rebind_count,expected_provider_head_digest,expected_provider_recovery_request_digest,decision_kind,pre_decision_database_authority_head_digest,current_database_identity_digest,database_timeline_lineage_chain_digest,current_database_incarnation_registration_digest,runtime_rebind_chain_digest,runtime_instance_binding_digest,database_transaction_id,transaction_snapshot_digest,database_point,transaction_nonce,decided_at,canonical_evidence_bundle_jcs,canonical_body_jcs,body_digest) VALUES ('11111111-1111-4111-8111-000000000006'::uuid,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'11111111-1111-4111-8111-000000000006'::uuid,'11111111-1111-4111-8111-000000000006'::uuid,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'11111111-1111-4111-8111-000000000006'::uuid,0,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'apply',decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),1,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'{}'::bytea,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'2026-08-23T13:00:00Z'::timestamptz,'{}'::bytea,'{}'::bytea,decode('1111111111111111111111111111111111111111111111111111111111111111','hex')) RETURNING ctid::text`},
	{table: "nodecontrol.control_plane_authority_epoch_transition_resolutions", insertSQL: `INSERT INTO nodecontrol.control_plane_authority_epoch_transition_resolutions (resolution_id,transition_id,terminal_application_id,transition_intent_digest,transition_application_digest,previous_epoch_transition_terminal_chain_digest,resolved_epoch,resolved_epoch_transition_chain_digest,current_database_identity_digest,database_timeline_lineage_chain_digest,status,database_transaction_id,transaction_snapshot_digest,resolved_database_point,resolved_at,canonical_evidence_bundle_jcs,canonical_body_jcs,body_digest) VALUES ('11111111-1111-4111-8111-000000000007'::uuid,'11111111-1111-4111-8111-000000000007'::uuid,'11111111-1111-4111-8111-000000000007'::uuid,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),2,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'approved_for_provider_resolution',1,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'{}'::bytea,'2026-08-23T13:00:00Z'::timestamptz,'{}'::bytea,'{}'::bytea,decode('1111111111111111111111111111111111111111111111111111111111111111','hex')) RETURNING ctid::text`},
	{table: "nodecontrol.control_plane_authority_epoch_transition_terminal_applications", insertSQL: `INSERT INTO nodecontrol.control_plane_authority_epoch_transition_terminal_applications (terminal_application_id,transition_id,resolution_id,transition_intent_digest,transition_application_digest_or_null,transition_resolution_digest_or_null,outcome,provider_resolution_receipt_digest_or_null,previous_epoch_transition_terminal_chain_digest,terminal_epoch_transition_terminal_chain_digest,effective_epoch,effective_epoch_transition_chain_digest,current_database_identity_digest,database_timeline_lineage_chain_digest,database_transaction_id,transaction_snapshot_digest,database_point,applied_at,canonical_evidence_bundle_jcs,canonical_body_jcs,body_digest) VALUES ('11111111-1111-4111-8111-000000000008'::uuid,'11111111-1111-4111-8111-000000000008'::uuid,'11111111-1111-4111-8111-000000000008'::uuid,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'resolved',decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),1,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),1,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'{}'::bytea,'2026-08-23T13:00:00Z'::timestamptz,'{}'::bytea,'{}'::bytea,decode('1111111111111111111111111111111111111111111111111111111111111111','hex')) RETURNING ctid::text`},
	{table: "nodecontrol.control_plane_authority_fences", insertSQL: `INSERT INTO nodecontrol.control_plane_authority_fences (operation_id,effect_kind,scope_kind,authority_epoch,authority_sequence,scope_digest,provider_reservation_digest,provider_status,visibility_state,reserved_at,authority_protocol_profile) VALUES ('11111111-1111-4111-8111-000000000009'::uuid,'grant_create','node',1,1,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'reserved','fence_pending','2026-08-23T12:00:00Z'::timestamptz,'legacy_v6') RETURNING ctid::text`},
	{table: "nodecontrol.control_plane_authority_fresh_restore_import_applications", insertSQL: `INSERT INTO nodecontrol.control_plane_authority_fresh_restore_import_applications (single_use_apply_id,staging_import_capability_digest,manifest_digest,target_activation_id,current_database_identity_digest,database_timeline_lineage_chain_digest,target_database_incarnation_registration_digest,runtime_rebind_chain_digest,runtime_instance_binding_digest,staging_exclusion_lease_digest,acquisition_locked_provider_head_digest,database_route_closed_digest,pre_import_inventory_digest,post_import_inventory_digest,imported_object_count,complete_node_set_digest,forbidden_state_zero_digest,database_transaction_id,transaction_snapshot_digest,applied_at,canonical_evidence_bundle_jcs,canonical_body_jcs,body_digest) VALUES ('11111111-1111-4111-8111-000000000010'::uuid,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'11111111-1111-4111-8111-000000000010'::uuid,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),0,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),1,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'2026-08-23T13:00:00Z'::timestamptz,'{}'::bytea,'{}'::bytea,decode('1111111111111111111111111111111111111111111111111111111111111111','hex')) RETURNING ctid::text`},
	{table: "nodecontrol.control_plane_authority_fresh_restore_requirements", insertSQL: `INSERT INTO nodecontrol.control_plane_authority_fresh_restore_requirements (requirement_id,upgrade_intent_digest,source_seal_kind,source_seal_digest,classification,reason,failed_checks,missing_external_evidence,source_database_identity_digest,local_runtime_isolation_digest,required_new_database_identity,required_new_deployment_id,required_new_provider_namespace,required_non_exportable_incarnation,blocked_capabilities,created_at,canonical_evidence_bundle_jcs,canonical_body_jcs,body_digest) VALUES ('11111111-1111-4111-8111-000000000011'::uuid,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'indeterminate',decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'indeterminate','catalog_unknown',ARRAY['catalog']::text[],ARRAY[]::text[],decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),true,true,true,true,ARRAY['c12_root_metadata_signer_rotation','node_authority_restore_reenrollment','node_operator_server_ca_rotation','operator_authorizer_change','trust_bundle_publish']::text[],'2026-08-23T12:00:00Z'::timestamptz,'{}'::bytea,'{}'::bytea,decode('1111111111111111111111111111111111111111111111111111111111111111','hex')) RETURNING ctid::text`},
	{table: "nodecontrol.control_plane_authority_indeterminate_source_seals", insertSQL: `INSERT INTO nodecontrol.control_plane_authority_indeterminate_source_seals (source_seal_id,upgrade_intent_digest,activation_id,request_nonce,classification,reason,failed_checks,local_runtime_isolation_digest,observed_deployment_id,observed_database_identity_digest,migration_version,missing_external_evidence,sealed_at,canonical_evidence_bundle_jcs,canonical_body_jcs,body_digest) VALUES ('11111111-1111-4111-8111-000000000012'::uuid,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'11111111-1111-4111-8111-000000000012'::uuid,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'indeterminate','catalog_unknown',ARRAY['catalog']::text[],decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'11111111-1111-4111-8111-000000000012'::uuid,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),7,ARRAY[]::text[],'2026-08-23T12:00:00Z'::timestamptz,'{}'::bytea,'{}'::bytea,decode('1111111111111111111111111111111111111111111111111111111111111111','hex')) RETURNING ctid::text`},
	{table: "nodecontrol.control_plane_authority_legacy_database_source_retirements", insertSQL: `INSERT INTO nodecontrol.control_plane_authority_legacy_database_source_retirements (database_retirement_id,retirement_plan_id,source_retirement_authorization_digest,source_membership_retirement_digest,upgrade_intent_digest,activation_id,environment_inventory_digest,environment_inventory_anchor_set_digest,environment_record_digest,deployment_id,database_identity_digest,local_runtime_isolation_digest,legacy_runtime_shutdown_digest,legacy_source_inventory_digest,normalized_catalog_digest,transaction_snapshot_digest,database_point,phase,retired_at,canonical_evidence_bundle_jcs,canonical_body_jcs,body_digest) VALUES ('11111111-1111-4111-8111-000000000013'::uuid,'11111111-1111-4111-8111-000000000013'::uuid,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'11111111-1111-4111-8111-000000000013'::uuid,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'11111111-1111-4111-8111-000000000013'::uuid,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'{}'::bytea,'source_database_retired','2026-08-23T12:00:00Z'::timestamptz,'{}'::bytea,'{}'::bytea,decode('1111111111111111111111111111111111111111111111111111111111111111','hex')) RETURNING ctid::text`},
	{table: "nodecontrol.control_plane_authority_legacy_source_seals", insertSQL: `INSERT INTO nodecontrol.control_plane_authority_legacy_source_seals (source_seal_id,upgrade_intent_digest,activation_id,classification,environment_inventory_digest,environment_inventory_anchor_set_digest,local_runtime_isolation_digest,legacy_runtime_shutdown_digest,legacy_runtime_shutdown_set_digest,source_membership_retirement_digest,legacy_source_retirement_set_digest,deployment_id,database_identity_digest,normalized_catalog_digest,legacy_source_inventory_digest,transaction_snapshot_digest,provider_identity_digest,provider_endpoint_identity_digest,provider_namespace,provider_profile,provider_head_digest,provider_history_high_water,credential_policy_digest,database_route_closed_digest,deployment_capability_revocation_digest,legacy_epoch_source_set_digest,legacy_epoch_maximum,legacy_epoch_maximum_evidence_digest,sealed_at,canonical_evidence_bundle_jcs,canonical_body_jcs,body_digest) VALUES ('11111111-1111-4111-8111-000000000014'::uuid,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'11111111-1111-4111-8111-000000000014'::uuid,'legacy_or_nonempty',decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'11111111-1111-4111-8111-000000000014'::uuid,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'task8','legacy_v6',decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),1,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),1,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'2026-08-23T12:00:00Z'::timestamptz,'{}'::bytea,'{}'::bytea,decode('1111111111111111111111111111111111111111111111111111111111111111','hex')) RETURNING ctid::text`},
	{table: "nodecontrol.control_plane_authority_protocol_activation_completions", insertSQL: `INSERT INTO nodecontrol.control_plane_authority_protocol_activation_completions (completion_id,activation_id,activation_digest,preparation_digest,provider_completion_digest,provider_completion_phase,database_activation_attestation_digest,current_database_incarnation_registration_digest,runtime_rebind_chain_digest,current_runtime_instance_binding_digest,credential_policy_digest,epoch_evidence_digest,genesis_epoch_transition_root_digest,selected_genesis_epoch,completed_at,canonical_evidence_bundle_jcs,canonical_body_jcs,body_digest) VALUES ('11111111-1111-4111-8111-000000000015'::uuid,'11111111-1111-4111-8111-000000000015'::uuid,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'genesis_completed_pending_release',decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),1,'2026-08-23T13:00:00Z'::timestamptz,'{}'::bytea,'{}'::bytea,decode('1111111111111111111111111111111111111111111111111111111111111111','hex')) RETURNING ctid::text`},
	{table: "nodecontrol.control_plane_authority_protocol_activation_releases", insertSQL: `INSERT INTO nodecontrol.control_plane_authority_protocol_activation_releases (release_preparation_id,open_id,activation_id,activation_digest,completion_digest,provider_release_preparation_digest,provider_release_phase,database_completion_attestation_digest,open_nonce,current_database_incarnation_registration_digest,runtime_rebind_chain_digest,current_runtime_instance_binding_digest,credential_policy_digest,epoch_evidence_digest,genesis_epoch_transition_root_digest,selected_genesis_epoch,released_at,canonical_evidence_bundle_jcs,canonical_body_jcs,body_digest) VALUES ('11111111-1111-4111-8111-000000000016'::uuid,'11111111-1111-4111-8111-000000000016'::uuid,'11111111-1111-4111-8111-000000000016'::uuid,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'genesis_release_prepared',decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),1,'2026-08-23T12:00:00Z'::timestamptz,'{}'::bytea,'{}'::bytea,decode('1111111111111111111111111111111111111111111111111111111111111111','hex')) RETURNING ctid::text`},
	{table: "nodecontrol.control_plane_authority_protocol_activations", insertSQL: `INSERT INTO nodecontrol.control_plane_authority_protocol_activations (activation_id,mode,attempt_digest,deployment_id,database_identity_digest,database_incarnation_attestation_digest,genesis_database_incarnation_registration_digest,runtime_registration_result_digest,runtime_rebind_chain_digest,activation_runtime_instance_binding_digest,database_legacy_absence_projection_digest,attempt_database_inventory_digest,activation_database_observation_digest,provider_namespace_absence_digest,environment_inventory_digest,environment_inventory_anchor_set_digest,local_runtime_isolation_digest,legacy_runtime_shutdown_digest,legacy_runtime_shutdown_set_digest,credential_policy_digest,epoch_evidence_digest,genesis_epoch_transition_root_digest,selected_genesis_epoch,provider_identity_digest,provider_endpoint_identity_digest,namespace,protocol_profile,preparation_digest,activated_at,canonical_evidence_bundle_jcs,canonical_body_jcs,body_digest) VALUES ('11111111-1111-4111-8111-000000000017'::uuid,'empty_in_place',decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'11111111-1111-4111-8111-000000000017'::uuid,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),1,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'task8','claim_v1',decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'2026-08-23T13:00:00Z'::timestamptz,'{}'::bytea,'{}'::bytea,decode('1111111111111111111111111111111111111111111111111111111111111111','hex')) RETURNING ctid::text`},
	{table: "nodecontrol.control_plane_authority_protocol_upgrade_attempts", insertSQL: `INSERT INTO nodecontrol.control_plane_authority_protocol_upgrade_attempts (attempt_id,upgrade_intent_digest,activation_id,preparation_id,completion_id,release_preparation_id,open_id,mode,deployment_id,request_nonce,environment_inventory_digest,environment_inventory_anchor_set_digest,local_runtime_isolation_digest,legacy_runtime_shutdown_digest,legacy_runtime_shutdown_set_digest,credential_policy_digest,database_legacy_absence_projection_digest,attempt_database_observation_digest,provider_namespace_absence_digest,database_inventory_digest,database_incarnation_attestation_digest,database_incarnation_registration_digest,runtime_registration_result_digest,runtime_rebind_chain_digest,runtime_instance_binding_digest,epoch_evidence_digest,genesis_epoch_transition_root_digest,selected_genesis_epoch,created_at,canonical_evidence_bundle_jcs,canonical_body_jcs,body_digest) VALUES ('11111111-1111-4111-8111-000000000018'::uuid,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'11111111-1111-4111-8111-000000000018'::uuid,'11111111-1111-4111-8111-000000000018'::uuid,'11111111-1111-4111-8111-000000000018'::uuid,'11111111-1111-4111-8111-000000000018'::uuid,'11111111-1111-4111-8111-000000000018'::uuid,'empty_in_place','11111111-1111-4111-8111-000000000018'::uuid,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),1,'2026-08-23T12:00:00Z'::timestamptz,'{}'::bytea,'{}'::bytea,decode('1111111111111111111111111111111111111111111111111111111111111111','hex')) RETURNING ctid::text`},
	{table: "nodecontrol.control_plane_authority_protocol_upgrade_intents", insertSQL: `INSERT INTO nodecontrol.control_plane_authority_protocol_upgrade_intents (intent_id,installation_id,installation_kind,activation_id,request_nonce,incarnation_registration_id,provider_absence_proof_id,observed_deployment_id,database_identity_digest,local_runtime_isolation_digest,classification_state,created_at,canonical_evidence_bundle_jcs,canonical_body_jcs,body_digest) VALUES ('11111111-1111-4111-8111-000000000019'::uuid,'11111111-1111-4111-8111-000000000019'::uuid,'production','11111111-1111-4111-8111-000000000019'::uuid,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'11111111-1111-4111-8111-000000000019'::uuid,'11111111-1111-4111-8111-000000000019'::uuid,'11111111-1111-4111-8111-000000000019'::uuid,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'pending','2026-08-23T12:00:00Z'::timestamptz,'{}'::bytea,'{}'::bytea,decode('1111111111111111111111111111111111111111111111111111111111111111','hex')) RETURNING ctid::text`},
	{table: "nodecontrol.control_plane_authority_runtime_rebind_results", insertSQL: `INSERT INTO nodecontrol.control_plane_authority_runtime_rebind_results (rebind_id,activation_id,runtime_registration_result_digest,provider_rebind_request_digest,provider_rebind_digest,authorized_database_authority_head_digest,authorized_database_point,previous_database_identity_digest,current_database_identity_digest,previous_database_timeline_lineage_chain_digest,current_database_timeline_lineage_chain_digest,previous_database_incarnation_registration_digest,current_database_incarnation_registration_digest,previous_runtime_rebind_chain_digest,current_runtime_rebind_chain_digest,previous_runtime_instance_binding_digest,current_runtime_instance_binding_digest,current_runtime_instance_id,current_runtime_instance_generation,current_attestor_runtime_lease_digest,database_result_application,provider_phase,provider_head_digest,provider_control_sequence,recorded_database_point,recorded_at,canonical_evidence_bundle_jcs,canonical_body_jcs,body_digest) VALUES ('11111111-1111-4111-8111-000000000020'::uuid,'11111111-1111-4111-8111-000000000020'::uuid,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'{}'::bytea,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'11111111-1111-4111-8111-000000000020'::uuid,1,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'{}'::jsonb,'registered_pending_genesis',decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),0,'{}'::bytea,'2026-08-23T12:00:00Z'::timestamptz,'{}'::bytea,'{}'::bytea,decode('1111111111111111111111111111111111111111111111111111111111111111','hex')) RETURNING ctid::text`},
	{table: "nodecontrol.control_plane_authority_runtime_registration_results", insertSQL: `INSERT INTO nodecontrol.control_plane_authority_runtime_registration_results (registration_id,upgrade_intent_digest,activation_id,provider_identity_digest,provider_endpoint_identity_digest,namespace,credential_policy_digest,database_incarnation_attestation_digest,provider_registration_digest,genesis_database_identity_digest,database_timeline_lineage_chain_digest,runtime_instance_binding_digest,runtime_instance_id,runtime_instance_generation,attestor_runtime_lease_digest,runtime_rebind_chain_digest,provider_head_digest,provider_phase,provider_control_sequence,database_point,recorded_at,canonical_evidence_bundle_jcs,canonical_body_jcs,body_digest) VALUES ('11111111-1111-4111-8111-000000000021'::uuid,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'11111111-1111-4111-8111-000000000021'::uuid,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'task8',decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'11111111-1111-4111-8111-000000000021'::uuid,1,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'registered_pending_genesis',0,'{}'::bytea,'2026-08-23T12:00:00Z'::timestamptz,'{}'::bytea,'{}'::bytea,decode('1111111111111111111111111111111111111111111111111111111111111111','hex')) RETURNING ctid::text`},
	{table: "nodecontrol.control_plane_authority_staging_import_capabilities", insertSQL: `INSERT INTO nodecontrol.control_plane_authority_staging_import_capabilities (capability_id,single_use_apply_id,manifest_digest,target_activation_id,target_database_identity_digest,database_timeline_lineage_chain_digest,target_database_incarnation_registration_digest,runtime_rebind_chain_digest,runtime_instance_binding_digest,planned_staging_exclusion_id,expected_pre_acquire_provider_head_digest,pre_acquire_database_incarnation_proof_digest,provider_phase,provider_serving_lease_absent_digest,database_route_closed_digest,pre_import_inventory_digest,allowed_object_set_digest,expected_post_import_inventory_digest,transaction_nonce,issued_at,expires_at,canonical_evidence_bundle_jcs,canonical_body_jcs,body_digest) VALUES ('11111111-1111-4111-8111-000000000022'::uuid,'11111111-1111-4111-8111-000000000022'::uuid,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'11111111-1111-4111-8111-000000000022'::uuid,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'11111111-1111-4111-8111-000000000022'::uuid,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'fresh_v7_staging_closed',decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'2026-08-23T12:00:00Z'::timestamptz,'2026-08-23T12:04:00Z','{}'::bytea,'{}'::bytea,decode('1111111111111111111111111111111111111111111111111111111111111111','hex')) RETURNING ctid::text`},
	{table: "nodecontrol.control_plane_authority_staging_import_capability_recovery_applications", insertSQL: `INSERT INTO nodecontrol.control_plane_authority_staging_import_capability_recovery_applications (recovery_application_id,provider_staging_import_capability_recovery_digest,staging_import_capability_recovery_intent_digest,recovery_id,capability_id,single_use_apply_id,staging_import_capability_digest,staging_exclusion_id,staging_exclusion_acquire_request_digest,staging_exclusion_digest,current_held_provider_head_digest,observed_capability_row_count,materialized_capability_row_count,capability_materialization_kind,current_database_identity_digest,database_timeline_lineage_chain_digest,current_database_incarnation_registration_digest,runtime_rebind_chain_digest,runtime_instance_binding_digest,recovery_scope,database_transaction_id,transaction_snapshot_digest,database_point,recovered_at,canonical_evidence_bundle_jcs,canonical_body_jcs,body_digest) VALUES ('11111111-1111-4111-8111-000000000023'::uuid,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'11111111-1111-4111-8111-000000000023'::uuid,'11111111-1111-4111-8111-000000000023'::uuid,'11111111-1111-4111-8111-000000000023'::uuid,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'11111111-1111-4111-8111-000000000023'::uuid,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),0,1,'inserted_from_provider_bundle',decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'rehydrate_for_revocation_only',1,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'{}'::bytea,'2026-08-23T12:00:00Z'::timestamptz,'{}'::bytea,'{}'::bytea,decode('1111111111111111111111111111111111111111111111111111111111111111','hex')) RETURNING ctid::text`},
	{table: "nodecontrol.control_plane_authority_staging_import_capability_recovery_intents", insertSQL: `INSERT INTO nodecontrol.control_plane_authority_staging_import_capability_recovery_intents (recovery_intent_id,recovery_id,recovery_application_id,revocation_id,revocation_application_id,terminal_transition_id,target_activation_id,capability_id,single_use_apply_id,staging_import_capability_digest,capability_registration_commit_attestation_digest,staging_exclusion_id,staging_exclusion_acquire_request_digest,staging_exclusion_digest,expected_current_held_provider_head_digest,expected_staging_exclusion_state,observed_capability_row_count,observed_import_application_row_count,observed_revocation_application_row_count,observed_recovery_application_row_count,current_database_identity_digest,current_database_incarnation_registration_digest,database_timeline_lineage_chain_digest,runtime_rebind_chain_digest,runtime_instance_binding_digest,pre_intent_database_authority_head_digest,recovery_reason,recovery_scope,database_transaction_id,transaction_snapshot_digest,database_point,transaction_nonce,created_at,canonical_evidence_bundle_jcs,canonical_body_jcs,body_digest) VALUES ('11111111-1111-4111-8111-000000000024'::uuid,'11111111-1111-4111-8111-000000000024'::uuid,'11111111-1111-4111-8111-000000000024'::uuid,'11111111-1111-4111-8111-000000000024'::uuid,'11111111-1111-4111-8111-000000000024'::uuid,'11111111-1111-4111-8111-000000000024'::uuid,'11111111-1111-4111-8111-000000000024'::uuid,'11111111-1111-4111-8111-000000000024'::uuid,'11111111-1111-4111-8111-000000000024'::uuid,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'11111111-1111-4111-8111-000000000024'::uuid,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'held',0,0,0,0,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'pitr_capability_row_missing','rehydrate_for_revocation_only',1,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'{}'::bytea,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'2026-08-23T12:00:00Z'::timestamptz,'{}'::bytea,'{}'::bytea,decode('1111111111111111111111111111111111111111111111111111111111111111','hex')) RETURNING ctid::text`},
	{table: "nodecontrol.control_plane_authority_staging_import_capability_revocation_applications", insertSQL: `INSERT INTO nodecontrol.control_plane_authority_staging_import_capability_revocation_applications (revocation_application_id,staging_import_capability_revocation_digest,staging_import_capability_digest,capability_id,single_use_apply_id,manifest_digest,target_activation_id,target_database_incarnation_registration_digest,runtime_rebind_chain_digest,runtime_instance_binding_digest,current_database_identity_digest,current_database_timeline_lineage_chain_digest,current_database_incarnation_registration_digest,current_runtime_rebind_chain_digest,current_runtime_instance_binding_digest,import_application_absence_digest,database_transaction_id,transaction_snapshot_digest,applied_at,canonical_evidence_bundle_jcs,canonical_body_jcs,body_digest) VALUES ('11111111-1111-4111-8111-000000000025'::uuid,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'11111111-1111-4111-8111-000000000025'::uuid,'11111111-1111-4111-8111-000000000025'::uuid,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'11111111-1111-4111-8111-000000000025'::uuid,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),1,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'2026-08-23T13:00:00Z'::timestamptz,'{}'::bytea,'{}'::bytea,decode('1111111111111111111111111111111111111111111111111111111111111111','hex')) RETURNING ctid::text`},
	{table: "nodecontrol.control_plane_trust_bundle_high_waters", insertSQL: `INSERT INTO nodecontrol.control_plane_trust_bundle_high_waters (purpose,listener_kind,trust_domain,authority_operation_id,authority_epoch,authority_sequence,bundle_version,bundle_digest,cumulative_set_digest,cumulative_set_count,updated_at) VALUES ('bootstrap_server','bootstrap','task8.example','11111111-1111-4111-8111-000000000026'::uuid,1,1,1,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),0,'2026-08-23T13:00:00Z'::timestamptz) RETURNING ctid::text`},
	{table: "nodecontrol.node_capacity_profiles", insertSQL: `INSERT INTO nodecontrol.node_capacity_profiles (profile_id,version,adapter,egress_limit_bps,connection_limit,handshake_limit_per_second,cpu_quota_millicores,cpu_limit_basis_points,memory_limit_bytes,task_limit,file_descriptor_limit,queue_limit,packet_loss_limit_basis_points,required_metrics,created_at) VALUES ('task8-profile',1,'fixture',1000000,1,1,100,1,67108864,32,64,1,1,ARRAY['cpu_basis_points']::text[],'2026-08-23T12:00:00Z'::timestamptz) RETURNING ctid::text`},
	{table: "nodecontrol.node_certificate_issuances", insertSQL: `INSERT INTO nodecontrol.node_certificate_issuances (issuance_id,authority_operation_id,authority_epoch,authority_sequence,node_id,attempt_id,issuance_kind,identity_epoch,lineage_id,issuer_id,csr_sha256,public_key_sha256,template_sha256,request_digest,status,created_at,updated_at) VALUES ('11111111-1111-4111-8111-000000000028'::uuid,'11111111-1111-4111-8111-000000000028'::uuid,1,1,'11111111-1111-4111-8111-000000000028'::uuid,'11111111-1111-4111-8111-000000000028'::uuid,'initial',1,'11111111-1111-4111-8111-000000000028'::uuid,'fixture-ca',decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'pending','2026-08-23T12:00:00Z'::timestamptz,'2026-08-23T13:00:00Z'::timestamptz) RETURNING ctid::text`},
	{table: "nodecontrol.node_certificates", insertSQL: `INSERT INTO nodecontrol.node_certificates (certificate_id,issuance_id,authority_operation_id,authority_epoch,authority_sequence,node_id,identity_epoch,lineage_id,issuer_id,serial_bytes,leaf_der,leaf_der_sha256,public_key_sha256,chain_der_sha256,valid_from,valid_until,status,created_at,updated_at,retention_until) VALUES ('11111111-1111-4111-8111-000000000029'::uuid,'11111111-1111-4111-8111-000000000029'::uuid,'11111111-1111-4111-8111-000000000029'::uuid,1,1,'11111111-1111-4111-8111-000000000029'::uuid,1,'11111111-1111-4111-8111-000000000029'::uuid,'fixture-ca','{}'::bytea,'{}'::bytea,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'2026-08-23T12:00:00Z'::timestamptz,'2026-08-23T14:00:00Z'::timestamptz,'active','2026-08-23T12:00:00Z'::timestamptz,'2026-08-23T13:00:00Z'::timestamptz,'2027-10-01T00:00:00Z'::timestamptz) RETURNING ctid::text`},
	{table: "nodecontrol.node_desired_states", insertSQL: `INSERT INTO nodecontrol.node_desired_states (node_id,generation,signing_id,authority_operation_id,authority_epoch,authority_sequence,inventory_version,resource_envelope_version,resource_envelope_digest,root_version,root_publish_id,metadata_version,metadata_publish_id,signing_key_id,canonical_payload,payload_digest,signature,issued_at,effective_deadline,valid_until,reason,created_at,retention_until) VALUES ('11111111-1111-4111-8111-000000000030'::uuid,1,'11111111-1111-4111-8111-000000000030'::uuid,'11111111-1111-4111-8111-000000000030'::uuid,1,1,1,1,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),1,'11111111-1111-4111-8111-000000000030'::uuid,1,'11111111-1111-4111-8111-000000000030'::uuid,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'{}'::bytea,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('22222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222','hex'),'2026-08-23T12:00:00Z'::timestamptz,'2026-08-23T14:00:00Z'::timestamptz,'2026-08-23T14:00:00Z'::timestamptz,'inventory_update','2026-08-23T12:00:00Z'::timestamptz,'2027-10-01T00:00:00Z'::timestamptz) RETURNING ctid::text`},
	{table: "nodecontrol.node_endpoints", insertSQL: `INSERT INTO nodecontrol.node_endpoints (endpoint_id,node_id,address,port,transport,protocol_capability,operator_state,inventory_version,created_at,updated_at) VALUES ('11111111-1111-4111-8111-000000000031'::uuid,'11111111-1111-4111-8111-000000000031'::uuid,'127.0.0.1',1,'tcp','bootstrap_v1','enabled',1,'2026-08-23T12:00:00Z'::timestamptz,'2026-08-23T13:00:00Z'::timestamptz) RETURNING ctid::text`},
	{table: "nodecontrol.node_enrollment_grants", insertSQL: `INSERT INTO nodecontrol.node_enrollment_grants (grant_id,authority_operation_id,authority_epoch,authority_sequence,node_id,identity_epoch,token_digest,csr_digest,idempotency_digest,created_at,expires_at) VALUES ('11111111-1111-4111-8111-000000000032'::uuid,'11111111-1111-4111-8111-000000000032'::uuid,1,1,'11111111-1111-4111-8111-000000000032'::uuid,1,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'2026-08-23T12:00:00Z'::timestamptz,'2026-08-23T12:10:00Z'::timestamptz) RETURNING ctid::text`},
	{table: "nodecontrol.node_failure_domain_membership", insertSQL: `INSERT INTO nodecontrol.node_failure_domain_membership (node_id,failure_domain_id,domain_type,inventory_version,created_at) VALUES ('11111111-1111-4111-8111-000000000033'::uuid,'11111111-1111-4111-8111-000000000033'::uuid,'facility',1,'2026-08-23T12:00:00Z'::timestamptz) RETURNING ctid::text`},
	{table: "nodecontrol.node_failure_domains", insertSQL: `INSERT INTO nodecontrol.node_failure_domains (failure_domain_id,domain_type,stable_id,created_at,updated_at) VALUES ('11111111-1111-4111-8111-000000000034'::uuid,'facility','task8-stable','2026-08-23T12:00:00Z'::timestamptz,'2026-08-23T13:00:00Z'::timestamptz) RETURNING ctid::text`},
	{table: "nodecontrol.node_inventory", insertSQL: `INSERT INTO nodecontrol.node_inventory (node_id,pop_code,operator_state,security_state,identity_state,created_at,updated_at) VALUES ('11111111-1111-4111-8111-000000000035'::uuid,'task8-down-seed','enabled','normal','never_enrolled','2026-08-23T12:00:00Z'::timestamptz,'2026-08-23T13:00:00Z'::timestamptz) RETURNING ctid::text`},
	{table: "nodecontrol.node_observed_states", insertSQL: `INSERT INTO nodecontrol.node_observed_states (node_id,boot_id,sequence,request_digest,observation_digest,canonical_observation,sample_ended_at,arrived_at,inventory_version,reducer_version,reducer_input_digest,reducer_state_digest,health_state,health_reason,capacity_accepting,agent_accepting,final_accepting,created_at,updated_at) VALUES ('11111111-1111-4111-8111-000000000036'::uuid,'11111111-1111-4111-8111-000000000036'::uuid,1,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'{}'::bytea,'2026-08-23T12:00:00Z'::timestamptz,'2026-08-23T12:00:00Z'::timestamptz,1,1,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'healthy','none',true,true,true,'2026-08-23T12:00:00Z'::timestamptz,'2026-08-23T13:00:00Z'::timestamptz) RETURNING ctid::text`},
	{table: "nodecontrol.node_operator_audit", insertSQL: `INSERT INTO nodecontrol.node_operator_audit (audit_id,command_id,operator_id,credential_digest,role,action,target_kind,target_id,reason,result,occurred_at,retention_until) VALUES ('11111111-1111-4111-8111-000000000037'::uuid,'11111111-1111-4111-8111-000000000037'::uuid,'task8-operator',decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'inventory_writer','update_inventory','node','task8-target','inventory_update','accepted','2026-08-23T12:00:00Z'::timestamptz,'2027-10-01T00:00:00Z'::timestamptz) RETURNING ctid::text`},
	{table: "nodecontrol.node_pops", insertSQL: `INSERT INTO nodecontrol.node_pops (pop_code,iso_country,region,operator_state,created_at,updated_at) VALUES ('task8-down-seed','US','task8-region','enabled','2026-08-23T12:00:00Z'::timestamptz,'2026-08-23T13:00:00Z'::timestamptz) RETURNING ctid::text`},
	{table: "nodecontrol.node_process_slots", insertSQL: `INSERT INTO nodecontrol.node_process_slots (node_id,slot_id,adapter,capacity_profile_id,capacity_profile_version,required,operator_state,inventory_version,created_at,updated_at) VALUES ('11111111-1111-4111-8111-000000000039'::uuid,'task8-slot','fixture','task8',1,true,'enabled',1,'2026-08-23T12:00:00Z'::timestamptz,'2026-08-23T13:00:00Z'::timestamptz) RETURNING ctid::text`},
	{table: "nodecontrol.node_recovery_sessions", insertSQL: `INSERT INTO nodecontrol.node_recovery_sessions (recovery_id,authority_operation_id,authority_epoch,authority_sequence,node_id,identity_epoch,reason,version,status,incident_set_digest,resume_operator_state,created_at,updated_at) VALUES ('11111111-1111-4111-8111-000000000040'::uuid,'11111111-1111-4111-8111-000000000040'::uuid,1,1,'11111111-1111-4111-8111-000000000040'::uuid,1,'authority_restore',1,'pending',decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'disabled','2026-08-23T12:00:00Z'::timestamptz,'2026-08-23T13:00:00Z'::timestamptz) RETURNING ctid::text`},
	{table: "nodecontrol.node_recovery_states", insertSQL: `INSERT INTO nodecontrol.node_recovery_states (node_id,recovery_generation,signing_id,authority_operation_id,authority_epoch,authority_sequence,identity_epoch,recovery_id,recovery_reason,recovery_session_version,incident_set_digest,incident_count,local_fault_bindings_digest,local_fault_binding_count,supervisor_fault_bindings_digest,supervisor_fault_binding_count,root_version,metadata_version,recovery_action,canonical_payload,payload_digest,signing_key_id,signature,issued_at,valid_until,created_at,retention_until) VALUES ('11111111-1111-4111-8111-000000000041'::uuid,1,'11111111-1111-4111-8111-000000000041'::uuid,'11111111-1111-4111-8111-000000000041'::uuid,1,1,1,'11111111-1111-4111-8111-000000000041'::uuid,'authority_restore',1,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),1,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),1,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),1,1,1,'hold_stopped','{}'::bytea,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('22222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222','hex'),'2026-08-23T12:00:00Z'::timestamptz,'2026-08-23T12:10:00Z'::timestamptz,'2026-08-23T12:00:00Z'::timestamptz,'2027-10-01T00:00:00Z'::timestamptz) RETURNING ctid::text`},
	{table: "nodecontrol.node_resource_envelopes", insertSQL: `INSERT INTO nodecontrol.node_resource_envelopes (node_id,envelope_version,authority_operation_id,authority_epoch,authority_sequence,envelope_digest,canonical_package,deployment_key_id,signature,agent_cpu_millicores,agent_memory_bytes,agent_task_limit,agent_file_descriptor_limit,supervisor_cpu_millicores,supervisor_memory_bytes,supervisor_task_limit,supervisor_file_descriptor_limit,core_parent_cpu_millicores,core_parent_memory_bytes,core_parent_task_limit,core_parent_file_descriptor_limit,aggregate_slot_file_descriptor_limit,aggregate_slot_tmpfs_bytes,aggregate_slot_tmpfs_inodes,detected_host_capacity_digest,issued_at,created_at) VALUES ('11111111-1111-4111-8111-000000000042'::uuid,1,'11111111-1111-4111-8111-000000000042'::uuid,1,1,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'{}'::bytea,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('22222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222','hex'),100,67108864,32,64,100,67108864,32,64,100,67108864,32,64,64,1048576,1,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'2026-08-23T12:00:00Z'::timestamptz,'2026-08-23T12:00:00Z'::timestamptz) RETURNING ctid::text`},
	{table: "nodecontrol.node_restore_reauthorization_approvals", insertSQL: `INSERT INTO nodecontrol.node_restore_reauthorization_approvals (approval_id,authority_operation_id,authority_epoch,authority_sequence,node_id,recovery_id,effect_digest,scope_digest,role,operator_id,credential_digest,leaf_der_sha256,operator_authority_epoch,operator_authority_sequence,authorizer_version,security_admin_binding_digest,pop_scope,evidence_completed_at,credential_expires_at,created_at,expires_at,status) VALUES ('11111111-1111-4111-8111-000000000043'::uuid,'11111111-1111-4111-8111-000000000043'::uuid,1,1,'11111111-1111-4111-8111-000000000043'::uuid,'11111111-1111-4111-8111-000000000043'::uuid,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'proposal','task8-operator',decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),1,1,1,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'task8','2026-08-23T13:00:00Z'::timestamptz,'2026-08-23T14:00:00Z'::timestamptz,'2026-08-23T12:00:00Z'::timestamptz,'2026-08-23T12:10:00Z'::timestamptz,'pending') RETURNING ctid::text`},
	{table: "nodecontrol.node_root_metadata_publish_intents", insertSQL: `INSERT INTO nodecontrol.node_root_metadata_publish_intents (publish_id,authority_operation_id,authority_epoch,authority_sequence,publish_kind,reason,base_root_version,base_metadata_version,reserved_version,canonical_payload,payload_digest,key_set_digest,current_key_ids,current_threshold,activation_deadline,status,created_at,updated_at) VALUES ('11111111-1111-4111-8111-000000000044'::uuid,'11111111-1111-4111-8111-000000000044'::uuid,1,1,'root','normal',0,0,1,'{}'::bytea,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),ARRAY[decode('1111111111111111111111111111111111111111111111111111111111111111','hex')]::bytea[],1,'2026-08-23T14:00:00Z'::timestamptz,'pending','2026-08-23T12:00:00Z'::timestamptz,'2026-08-23T13:00:00Z'::timestamptz) RETURNING ctid::text`},
	{table: "nodecontrol.node_root_metadata_signature_shares", insertSQL: `INSERT INTO nodecontrol.node_root_metadata_signature_shares (publish_id,key_id,physical_key_id,payload_digest,signature_role,signature,verified_at) VALUES ('11111111-1111-4111-8111-000000000045'::uuid,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'current_root',decode('22222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222','hex'),'2026-08-23T12:00:00Z'::timestamptz) RETURNING ctid::text`},
	{table: "nodecontrol.node_security_fault_receipts", insertSQL: `INSERT INTO nodecontrol.node_security_fault_receipts (receipt_id,authority_operation_id,authority_epoch,authority_sequence,node_id,identity_epoch,local_fault_id,request_digest,fault_subtype,evidence_digest,agent_boot_id,incident_id,local_binding_slot,result,delivery_status,binding_status,created_at,updated_at) VALUES ('11111111-1111-4111-8111-000000000046'::uuid,'11111111-1111-4111-8111-000000000046'::uuid,1,1,'11111111-1111-4111-8111-000000000046'::uuid,1,'11111111-1111-4111-8111-000000000046'::uuid,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'identity_compromise',decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'11111111-1111-4111-8111-000000000046'::uuid,'11111111-1111-4111-8111-000000000046'::uuid,1,'accepted','pending','active','2026-08-23T12:00:00Z'::timestamptz,'2026-08-23T13:00:00Z'::timestamptz) RETURNING ctid::text`},
	{table: "nodecontrol.node_security_incidents", insertSQL: `INSERT INTO nodecontrol.node_security_incidents (incident_id,authority_operation_id,authority_epoch,authority_sequence,node_id,identity_epoch,fault_subtype,subtype_slot,status,first_evidence_digest,last_evidence_digest,occurrence_count,trust_context_digest,first_occurred_at,last_occurred_at) VALUES ('11111111-1111-4111-8111-000000000047'::uuid,'11111111-1111-4111-8111-000000000047'::uuid,1,1,'11111111-1111-4111-8111-000000000047'::uuid,1,'identity_compromise',1,'open',decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),1,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),'2026-08-23T12:00:00Z'::timestamptz,'2026-08-23T12:00:00Z'::timestamptz) RETURNING ctid::text`},
	{table: "nodecontrol.node_state_signing_intents", insertSQL: `INSERT INTO nodecontrol.node_state_signing_intents (signing_id,authority_operation_id,authority_epoch,authority_sequence,node_id,signing_kind,idempotency_key_digest,base_generation,reserved_generation,canonical_payload,payload_digest,root_version,metadata_version,expected_key_id,expected_public_key_digest,captured_inventory_version,captured_identity_epoch,captured_security_version,activation_deadline,status,created_at,updated_at) VALUES ('11111111-1111-4111-8111-000000000048'::uuid,'11111111-1111-4111-8111-000000000048'::uuid,1,1,'11111111-1111-4111-8111-000000000048'::uuid,'desired',decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),0,2,'{}'::bytea,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),1,1,decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),decode('1111111111111111111111111111111111111111111111111111111111111111','hex'),1,1,1,'2026-08-23T14:00:00Z'::timestamptz,'pending','2026-08-23T12:00:00Z'::timestamptz,'2026-08-23T13:00:00Z'::timestamptz) RETURNING ctid::text`},
	{table: "nodecontrol.node_state_transitions", insertSQL: `INSERT INTO nodecontrol.node_state_transitions (transition_id,node_id,dimension,from_state,to_state,reason,aggregate_version,occurred_at,retention_until) VALUES ('11111111-1111-4111-8111-000000000049'::uuid,'11111111-1111-4111-8111-000000000049'::uuid,'operator','enabled','draining','drain_maintenance',1,'2026-08-23T12:00:00Z'::timestamptz,'2027-10-01T00:00:00Z'::timestamptz) RETURNING ctid::text`},
}

func task8DownProtectedSnapshot(t *testing.T, database interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}) []byte {
	t.Helper()
	registryQueries := make([]string, len(task8PristineDownRegistry))
	for index, relation := range task8PristineDownRegistry {
		registryQueries[index] = fmt.Sprintf(`SELECT %s::text AS relation_name,
  (SELECT coalesce(jsonb_agg(to_jsonb(row_value) ORDER BY to_jsonb(row_value)::text COLLATE "C"),'[]'::jsonb) FROM %s AS row_value) AS relation_rows`,
			task8SQLLiteral(relation.table), relation.table)
	}
	query := `WITH relation_snapshots AS (` + strings.Join(registryQueries, " UNION ALL ") + `)
SELECT convert_to(jsonb_build_object(
 'version',(SELECT max(version_id) FROM public.goose_db_version WHERE is_applied),
 'tables',(SELECT count(*) FROM pg_catalog.pg_class c JOIN pg_catalog.pg_namespace n ON n.oid=c.relnamespace WHERE n.nspname='nodecontrol' AND c.relkind='r'),
 'latches',(SELECT coalesce(jsonb_agg(to_jsonb(l) ORDER BY l.installation_id),'[]'::jsonb) FROM nodecontrol.control_plane_authority_protocol_migration_latches l),
 'authorizations',(SELECT coalesce(jsonb_agg(to_jsonb(a) ORDER BY a.authorization_id),'[]'::jsonb) FROM nodecontrol.control_plane_authority_protocol_downgrade_authorizations a),
	'roles',(SELECT coalesce(jsonb_agg(jsonb_build_object('name',r.rolname,'login',r.rolcanlogin,'super',r.rolsuper,'createdb',r.rolcreatedb,'createrole',r.rolcreaterole,'replication',r.rolreplication,'bypassrls',r.rolbypassrls,'config',r.rolconfig) ORDER BY r.rolname COLLATE "C"),'[]'::jsonb) FROM pg_catalog.pg_roles r WHERE r.rolname=ANY(ARRAY['nodecontrol_upgrade_executor','nodecontrol_migration_downgrader','nodecontrol_staging_importer'])),
 'relations',(SELECT jsonb_object_agg(relation_name,relation_rows ORDER BY relation_name COLLATE "C") FROM relation_snapshots)
)::text,'UTF8')`
	var snapshot []byte
	if err := database.QueryRowContext(t.Context(), query).Scan(&snapshot); err != nil {
		t.Fatal("take one-roundtrip complete 49-relation Down snapshot:", err)
	}
	return snapshot
}

func task8SQLLiteral(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func TestNodeControlV7DownThreeStageMatrix(t *testing.T) {
	database, migrationFS, _ := openOwnedAuthorityV7Database(t)
	applyAuthorityV7Base(t, database, migrationFS)
	const label = "down-three-stage"
	installAuthorityV7Fixture(t, database, migrationFS, label)
	assertAuthorityV7ControlCardinality(t, database, 1, 0)

	facts := disposableAuthorityV7DownFacts(time.Now().UTC().Truncate(time.Microsecond), label)
	grant, err := task8NewDisposableAuthorityV7DownGrant(t, facts)
	if err != nil {
		t.Fatal("construct disposable authority-v7 Down grant:", err)
	}
	authorizer := &fixtureAuthorityV7DownAuthorizer{}
	ctx := migrations.WithAuthorityV7MigrationContext(t.Context(), migrations.AuthorityV7MigrationContext{
		InstallationKind: migrations.InstallationKindDisposableFixture,
		DownGrant:        &grant,
		DownAuthorizer:   authorizer,
	})
	result, err := newAuthorityV7Provider(t, database, migrationFS).Down(ctx)
	if err != nil {
		t.Fatal("apply provider-scoped authority-v7 Down:", err)
	}
	if result.Source.Version != 7 || authorizer.CallCount() != 1 {
		t.Fatalf("Down result/calls = version %d, calls %d; want 7/1", result.Source.Version, authorizer.CallCount())
	}

	var v7TableCount int
	if err := database.QueryRowContext(t.Context(), `SELECT count(*) FROM pg_catalog.pg_class c JOIN pg_catalog.pg_namespace n ON n.oid=c.relnamespace WHERE n.nspname='nodecontrol' AND c.relkind='r' AND c.relname LIKE 'control_plane_authority_%' AND c.relname <> 'control_plane_authority_fences'`).Scan(&v7TableCount); err != nil {
		t.Fatal("inspect catalog after authority-v7 Down:", err)
	}
	if v7TableCount != 0 {
		t.Fatalf("additive authority-v7 table count after Down = %d, want 0", v7TableCount)
	}
	var baseTableCount int
	if err := database.QueryRowContext(t.Context(), `SELECT count(*) FROM pg_catalog.pg_class c JOIN pg_catalog.pg_namespace n ON n.oid=c.relnamespace WHERE n.nspname='nodecontrol' AND c.relkind='r'`).Scan(&baseTableCount); err != nil {
		t.Fatal("inspect base catalog after authority-v7 Down:", err)
	}
	if baseTableCount != 25 {
		t.Fatalf("base table count after authority-v7 Down = %d, want 25", baseTableCount)
	}

	installAuthorityV7Fixture(t, database, migrationFS, label+"-second")
	var finalTableCount int
	if err := database.QueryRowContext(t.Context(), `SELECT count(*) FROM pg_catalog.pg_class c JOIN pg_catalog.pg_namespace n ON n.oid=c.relnamespace WHERE n.nspname='nodecontrol' AND c.relkind='r'`).Scan(&finalTableCount); err != nil {
		t.Fatal("inspect second authority-v7 Up:", err)
	}
	if finalTableCount != 51 {
		t.Fatalf("second authority-v7 Up table count = %d, want 51", finalTableCount)
	}
}

func TestNodeControlV7DownCrashSeams(t *testing.T) {
	database, migrationFS, _ := openOwnedAuthorityV7Database(t)
	applyAuthorityV7Base(t, database, migrationFS)
	const label = "down-rollback"
	installAuthorityV7Fixture(t, database, migrationFS, label)

	facts := disposableAuthorityV7DownFacts(time.Now().UTC().Truncate(time.Microsecond), label)
	grant, err := task8NewDisposableAuthorityV7DownGrant(t, facts)
	if err != nil {
		t.Fatal(err)
	}
	wantFailure := errors.New("fixture authorizer response loss")
	authorizer := &fixtureAuthorityV7DownAuthorizer{failure: wantFailure}
	ctx := migrations.WithAuthorityV7MigrationContext(t.Context(), migrations.AuthorityV7MigrationContext{
		InstallationKind: migrations.InstallationKindDisposableFixture,
		DownGrant:        &grant,
		DownAuthorizer:   authorizer,
	})
	if _, err := newAuthorityV7Provider(t, database, migrationFS).Down(ctx); !errors.Is(err, wantFailure) {
		t.Fatalf("Down authorizer failure = %v, want %v", err, wantFailure)
	}
	assertAuthorityV7ControlCardinality(t, database, 1, 0)
	if _, err := newAuthorityV7Provider(t, database, migrationFS).Down(ctx); err == nil {
		t.Fatal("rolled-back Down resurrected its consumed grant")
	}
	assertAuthorityV7ControlCardinality(t, database, 1, 0)
}

func TestNodeControlV7ProductionDownRejected(t *testing.T) {
	database, migrationFS, _ := openOwnedAuthorityV7Database(t)
	applyAuthorityV7Base(t, database, migrationFS)
	installAuthorityV7Fixture(t, database, migrationFS, "production-down-rejected")
	ctx := migrations.WithAuthorityV7MigrationContext(t.Context(), migrations.AuthorityV7MigrationContext{
		InstallationKind: migrations.InstallationKindProduction,
		DownAuthorizer:   &fixtureAuthorityV7DownAuthorizer{},
	})
	if _, err := newAuthorityV7Provider(t, database, migrationFS).Down(ctx); err == nil {
		t.Fatal("production authority-v7 Down proceeded without a disposable-only capability")
	}
}

type fixtureAuthorityV7DownAuthorizer struct {
	mu      sync.Mutex
	calls   int
	failure error
}

func (authorizer *fixtureAuthorityV7DownAuthorizer) AuthorizeAuthorityProtocolDowngrade(_ context.Context, request contracts.AuthorityV7DownAuthorizationRequestV1) (authority.VerifiedAuthorityV7DownAuthorization, error) {
	authorizer.mu.Lock()
	authorizer.calls++
	failure := authorizer.failure
	authorizer.mu.Unlock()
	if failure != nil {
		return authority.VerifiedAuthorityV7DownAuthorization{}, failure
	}
	now := time.Now().UTC().Truncate(time.Microsecond)
	digest := func(value string) contracts.Digest {
		return contracts.Digest(sha256.Sum256([]byte("task8-down-authorizer:" + value)))
	}
	facts := contracts.AuthorityV7DownAuthorizationFactsV1{
		Request: request,
		Authorization: contracts.AuthorityV7ProtocolDowngradeAuthorizationFactsV1{
			AuthorizationID:                              request.AuthorizationID,
			InstallationID:                               request.MigrationLatch.InstallationID,
			MigrationLatchDigest:                         request.MigrationLatchDigest,
			DatabaseIdentityDigest:                       request.MigrationLatch.DatabaseIdentityDigest,
			MigrationVersion:                             7,
			CurrentCatalogDigest:                         request.CurrentCatalogDigest,
			PristineDowngradeInventoryDigest:             request.PristineInventoryDigest,
			ProviderProtocolDowngradeRetirementSetDigest: request.ProviderRetirementSetDigest,
			EnvironmentInventoryAnchorSetDigest:          request.EnvironmentAnchorSetDigest,
			DatabaseTransactionID:                        request.ActualDatabaseTransactionID,
			TransactionNonce:                             request.TransactionNonce,
			AuthorizationScope:                           request.AuthorizationScope,
			IssuedAt:                                     now,
			ExpiresAt:                                    now.Add(time.Minute),
		},
		AuthorizationDigest:      digest(request.AuthorizationID.String()),
		AuthorizationEnvelopeJCS: []byte(`{"schema":"authority-v7-down-fixture/v1"}`),
		SignerRole:               "authority_protocol_downgrade_authorizer",
		SignerKeyID:              "task8-integration-key",
		SignaturePolicyVersion:   1,
		TrustRootDigest:          digest("trust-root"),
		SignatureAlgorithm:       "ed25519",
	}
	return task8NewAuthorityV7DownAuthorization(t8NoopTestingT{}, facts)
}

type task8TestReporter interface {
	Helper()
}

func task8NewDisposableAuthorityV7DownGrant(t task8TestReporter, facts contracts.AuthorityV7DownMigrationFactsV1) (authority.VerifiedAuthorityV7DownGrant, error) {
	t.Helper()
	boundFacts, body, bundle, err := task8BuildOpaqueV7DownInputs(facts)
	if err != nil {
		return authority.VerifiedAuthorityV7DownGrant{}, err
	}
	grant, err := authority.NewDisposableAuthorityV7DownGrantForIntegration(boundFacts, body, bundle)
	// The factory contract is snapshot-based: subsequent caller mutation must
	// not alter the verified capability that it returned.
	body[0] ^= 0xff
	bundle[0] ^= 0xff
	return grant, err
}

func task8NewAuthorityV7DownAuthorization(t task8TestReporter, facts contracts.AuthorityV7DownAuthorizationFactsV1) (authority.VerifiedAuthorityV7DownAuthorization, error) {
	t.Helper()
	body, err := task8OpaqueV7AuthorizationBody(facts.Authorization)
	if err != nil {
		return authority.VerifiedAuthorityV7DownAuthorization{}, err
	}
	signed, err := task8OpaqueV7Sign("authority-protocol-downgrade-authorization.v1", "authority_protocol_downgrade_authorizer", body)
	if err != nil {
		return authority.VerifiedAuthorityV7DownAuthorization{}, err
	}
	facts.AuthorizationDigest = signed.digest
	facts.AuthorizationEnvelopeJCS = append([]byte(nil), signed.envelope...)
	facts.SignerRole = "authority_protocol_downgrade_authorizer"
	facts.SignerKeyID = task8OpaqueV7SignerKeyID
	facts.SignaturePolicyVersion = 1
	trustRoot, err := hex.DecodeString(strings.Repeat("a1", 32))
	if err != nil {
		return authority.VerifiedAuthorityV7DownAuthorization{}, err
	}
	copy(facts.TrustRootDigest[:], trustRoot)
	facts.SignatureAlgorithm = "ed25519"
	authorization, err := authority.NewAuthorityV7DownAuthorizationForIntegration(facts, body)
	body[0] ^= 0xff
	facts.AuthorizationEnvelopeJCS[0] ^= 0xff
	return authorization, err
}

type t8NoopTestingT struct{}

func (t8NoopTestingT) Helper() {}

const task8OpaqueV7SignerKeyID = "56475aa75463474c0285df5dbf2bcab73da651358839e9b77481b2eab107708c"

type task8OpaqueV7SignedArtifact struct {
	schema   string
	digest   contracts.Digest
	body     []byte
	envelope []byte
}

type task8OpaqueV7EvidenceMember struct {
	kind     string
	schema   string
	digest   contracts.Digest
	body     []byte
	envelope []byte
}

func task8BuildOpaqueV7DownInputs(facts contracts.AuthorityV7DownMigrationFactsV1) (contracts.AuthorityV7DownMigrationFactsV1, []byte, []byte, error) {
	if facts.ProviderRetirementSet.ProviderCount != 1 || len(facts.ProviderRetirementSet.Retirements) != 1 ||
		facts.ProviderRetirementSet.RetiredMemberCount != 1 || len(facts.ProviderRetirementSet.RetiredMembers) != 1 {
		return contracts.AuthorityV7DownMigrationFactsV1{}, nil, nil, fmt.Errorf("opaque Down fixture requires one provider and one retired member")
	}
	facts.StableTableInventory = append([]contracts.AuthorityV7StableTableInventoryItemV1(nil), facts.StableTableInventory...)
	facts.ProviderRetirementSet.Retirements = append([]contracts.AuthorityV7ProviderRetirementFactsV1(nil), facts.ProviderRetirementSet.Retirements...)
	facts.ProviderRetirementSet.RetiredMembers = append([]contracts.AuthorityV7RetiredEnvironmentMemberFactsV1(nil), facts.ProviderRetirementSet.RetiredMembers...)

	now := time.Now().UTC().Truncate(time.Microsecond)
	facts.ExpiresAt = now.Add(time.Minute)
	set := &facts.ProviderRetirementSet
	set.CreatedAt = now.Add(-15 * time.Second)
	member := set.RetiredMembers[0]
	retirement := &set.Retirements[0]
	fixtureLabel := facts.AuthorizationID.String()
	fixtureUUID := func(suffix string) uuid.UUID {
		return uuid.NewSHA1(uuid.NameSpaceOID, []byte("task8-opaque-v7:"+fixtureLabel+":"+suffix))
	}
	fixtureDigest := func(suffix string) contracts.Digest {
		return contracts.Digest(sha256.Sum256([]byte("task8-opaque-v7:" + fixtureLabel + ":" + suffix)))
	}

	recordDigestText := task8OpaqueV7DigestText(member.EnvironmentRecordDigest)
	memberSetBody, err := task8OpaqueV7Canonical([]string{recordDigestText})
	if err != nil {
		return contracts.AuthorityV7DownMigrationFactsV1{}, nil, nil, err
	}
	memberSetDigest := task8OpaqueV7DomainDigest("environment-inventory-member-set.v1", memberSetBody)
	membershipRequest := map[string]any{
		"membership_retirement_id":                      fixtureUUID("membership-retirement").String(),
		"installation_id":                               set.InstallationID.String(),
		"installation_kind":                             string(facts.MigrationLatch.InstallationKind),
		"migration_latch_digest":                        task8OpaqueV7DigestText(facts.MigrationLatchDigest),
		"database_identity_digest":                      task8OpaqueV7DigestText(facts.MigrationLatch.DatabaseIdentityDigest),
		"release_scope":                                 set.ReleaseScope,
		"final_environment_inventory_digest":            task8OpaqueV7DigestText(set.EnvironmentInventoryDigest),
		"final_inventory_sequence":                      "1",
		"final_environment_inventory_anchor_set_digest": task8OpaqueV7DigestText(set.EnvironmentInventoryAnchorSetDigest),
		"environment_member_set_digest":                 task8OpaqueV7DigestText(memberSetDigest),
		"environment_count":                             strconv.FormatUint(set.RetiredMemberCount, 10),
		"authorization_nonce":                           task8OpaqueV7DigestText(fixtureDigest("membership-authorization-nonce")),
		"request_nonce":                                 task8OpaqueV7DigestText(fixtureDigest("membership-request-nonce")),
		"issued_at":                                     task8OpaqueV7Time(now.Add(-time.Minute)),
		"expires_at":                                    task8OpaqueV7Time(now.Add(time.Minute)),
	}
	membershipRequestBody, err := task8OpaqueV7Canonical(membershipRequest)
	if err != nil {
		return contracts.AuthorityV7DownMigrationFactsV1{}, nil, nil, err
	}
	membershipResponse := make(map[string]any, len(membershipRequest)+4)
	for key, value := range membershipRequest {
		membershipResponse[key] = value
	}
	membershipResponse["request_digest"] = task8OpaqueV7DigestText(task8OpaqueV7DomainDigest("retire-environment-inventory-membership-request.v1", membershipRequestBody))
	membershipResponse["membership_retirement_tombstone_id"] = fixtureUUID("membership-tombstone").String()
	membershipResponse["phase"] = "inventory_membership_retired"
	membershipResponse["retired_at"] = task8OpaqueV7Time(now.Add(-30 * time.Second))
	membershipBody, err := task8OpaqueV7Canonical(membershipResponse)
	if err != nil {
		return contracts.AuthorityV7DownMigrationFactsV1{}, nil, nil, err
	}
	membership, err := task8OpaqueV7Sign("environment-inventory-membership-retirement.v1", "release_deployment_operator", membershipBody)
	if err != nil {
		return contracts.AuthorityV7DownMigrationFactsV1{}, nil, nil, err
	}
	set.EnvironmentInventoryMembershipRetirementDigest = membership.digest

	membersBody, err := task8OpaqueV7Canonical([]map[string]any{task8OpaqueV7RetiredMemberWire(member)})
	if err != nil {
		return contracts.AuthorityV7DownMigrationFactsV1{}, nil, nil, err
	}
	retirementMemberSetDigest := task8OpaqueV7DomainDigest("provider-protocol-retirement-member-set.v1", membersBody)
	retirementID := fixtureUUID("provider-retirement")
	adminBody, err := task8OpaqueV7Canonical(map[string]any{
		"authorization_id":                                   fixtureUUID("provider-retirement-authorization").String(),
		"retirement_id":                                      retirementID.String(),
		"installation_id":                                    set.InstallationID.String(),
		"installation_kind":                                  string(facts.MigrationLatch.InstallationKind),
		"migration_latch_digest":                             task8OpaqueV7DigestText(facts.MigrationLatchDigest),
		"database_identity_digest":                           task8OpaqueV7DigestText(facts.MigrationLatch.DatabaseIdentityDigest),
		"release_scope":                                      set.ReleaseScope,
		"environment_inventory_digest":                       task8OpaqueV7DigestText(set.EnvironmentInventoryDigest),
		"environment_inventory_anchor_set_digest":            task8OpaqueV7DigestText(set.EnvironmentInventoryAnchorSetDigest),
		"environment_inventory_membership_retirement_digest": task8OpaqueV7DigestText(set.EnvironmentInventoryMembershipRetirementDigest),
		"provider_identity_digest":                           task8OpaqueV7DigestText(retirement.ProviderIdentityDigest),
		"provider_endpoint_identity_digest":                  task8OpaqueV7DigestText(retirement.ProviderEndpointIdentityDigest),
		"retirement_member_set_digest":                       task8OpaqueV7DigestText(retirementMemberSetDigest),
		"member_count":                                       "1",
		"authorization_scope":                                "permanent_disposable_namespace_retirement",
		"authorization_nonce":                                task8OpaqueV7DigestText(fixtureDigest("provider-retirement-authorization-nonce")),
		"issued_at":                                          task8OpaqueV7Time(now.Add(-time.Minute)),
		"expires_at":                                         task8OpaqueV7Time(now.Add(time.Minute)),
	})
	if err != nil {
		return contracts.AuthorityV7DownMigrationFactsV1{}, nil, nil, err
	}
	admin, err := task8OpaqueV7Sign("provider-protocol-downgrade-retirement-authorization.v1", "provider_downgrade_retirement_admin", adminBody)
	if err != nil {
		return contracts.AuthorityV7DownMigrationFactsV1{}, nil, nil, err
	}

	historyFields := map[string]any{
		"provider_identity_digest":          task8OpaqueV7DigestText(retirement.ProviderIdentityDigest),
		"provider_endpoint_identity_digest": task8OpaqueV7DigestText(retirement.ProviderEndpointIdentityDigest),
		"namespace":                         member.ProviderNamespace,
		"observed_profiles":                 []string{"claim_v1", "legacy_v6"},
		"observed_at":                       task8OpaqueV7Time(now.Add(-time.Minute)),
		"expires_at":                        task8OpaqueV7Time(now.Add(time.Minute)),
	}
	for _, field := range []string{
		"unknown_profile_count", "unknown_record_count", "legacy_reservation_count", "legacy_terminal_count", "legacy_epoch_transition_count", "legacy_cutover_count",
		"claim_v1_credential_policy_count", "claim_v1_accepted_key_mutation_count", "incarnation_registration_count", "genesis_preparation_count",
		"genesis_completion_count", "genesis_release_preparation_count", "genesis_open_count", "claim_v1_reservation_count", "claim_v1_terminal_count",
		"claim_v1_epoch_transition_count", "claim_v1_epoch_recovery_count", "serving_lease_event_count", "runtime_rebind_count", "timeline_lineage_event_count",
		"staging_exclusion_count", "staging_recovery_count", "source_retirement_count", "genesis_authorizing_control_count", "other_mutation_count",
		"provider_history_high_water",
	} {
		historyFields[field] = "0"
	}
	historyBody, err := task8OpaqueV7Canonical(historyFields)
	if err != nil {
		return contracts.AuthorityV7DownMigrationFactsV1{}, nil, nil, err
	}
	history, err := task8OpaqueV7Sign("provider-protocol-history-zero-projection.v1", "claim_v1_provider_history_auditor", historyBody)
	if err != nil {
		return contracts.AuthorityV7DownMigrationFactsV1{}, nil, nil, err
	}

	providerRequest := map[string]any{
		"retirement_id":    retirementID.String(),
		"retirement_nonce": task8OpaqueV7DigestText(fixtureDigest("provider-retirement-nonce")),
		"provider_downgrade_retirement_authorization_digest": task8OpaqueV7DigestText(admin.digest),
		"environment_inventory_membership_retirement_digest": task8OpaqueV7DigestText(set.EnvironmentInventoryMembershipRetirementDigest),
		"installation_id":                         set.InstallationID.String(),
		"release_scope":                           set.ReleaseScope,
		"environment_inventory_digest":            task8OpaqueV7DigestText(set.EnvironmentInventoryDigest),
		"environment_inventory_anchor_set_digest": task8OpaqueV7DigestText(set.EnvironmentInventoryAnchorSetDigest),
		"provider_identity_digest":                task8OpaqueV7DigestText(retirement.ProviderIdentityDigest),
		"provider_endpoint_identity_digest":       task8OpaqueV7DigestText(retirement.ProviderEndpointIdentityDigest),
		"member_count":                            "1",
		"members":                                 json.RawMessage(membersBody),
		"retirement_member_set_digest":            task8OpaqueV7DigestText(retirementMemberSetDigest),
		"request_nonce":                           task8OpaqueV7DigestText(fixtureDigest("provider-retirement-request-nonce")),
	}
	providerRequestBody, err := task8OpaqueV7Canonical(providerRequest)
	if err != nil {
		return contracts.AuthorityV7DownMigrationFactsV1{}, nil, nil, err
	}
	providerResponse := make(map[string]any, len(providerRequest)+5)
	for key, value := range providerRequest {
		providerResponse[key] = value
	}
	providerResponse["request_digest"] = task8OpaqueV7DigestText(task8OpaqueV7DomainDigest("retire-provider-protocol-for-downgrade-request.v1", providerRequestBody))
	providerResponse["provider_control_sequence"] = "1"
	providerResponse["namespace_states"] = []map[string]any{{
		"namespace":                       member.ProviderNamespace,
		"environment_record_digests":      []string{recordDigestText},
		"database_identity_digests":       []string{task8OpaqueV7DigestText(member.DatabaseIdentityDigest)},
		"history_zero_projection_digest":  task8OpaqueV7DigestText(history.digest),
		"pre_retirement_control_sequence": "0",
		"provider_history_high_water":     "0",
		"retirement_tombstone_id":         fixtureUUID("retirement-tombstone").String(),
	}}
	providerResponse["phase"] = "down_retired"
	providerResponse["retired_at"] = task8OpaqueV7Time(now.Add(-30 * time.Second))
	retirementBody, err := task8OpaqueV7Canonical(providerResponse)
	if err != nil {
		return contracts.AuthorityV7DownMigrationFactsV1{}, nil, nil, err
	}
	providerRetirement, err := task8OpaqueV7Sign("provider-protocol-downgrade-retirement.v1", "claim_v1_provider", retirementBody)
	if err != nil {
		return contracts.AuthorityV7DownMigrationFactsV1{}, nil, nil, err
	}
	retirement.ProviderProtocolDowngradeRetirementDigest = providerRetirement.digest

	setBody, err := task8OpaqueV7RetirementSetBody(*set)
	if err != nil {
		return contracts.AuthorityV7DownMigrationFactsV1{}, nil, nil, err
	}
	facts.ProviderRetirementSetDigest = task8OpaqueV7DomainDigest("provider-protocol-downgrade-retirement-set.v1", setBody)
	bundle, err := task8OpaqueV7EvidenceBundle(
		"provider-protocol-downgrade-retirement-set.v1",
		facts.ProviderRetirementSetDigest,
		[]task8OpaqueV7SignedArtifact{membership, admin, history, providerRetirement},
	)
	if err != nil {
		return contracts.AuthorityV7DownMigrationFactsV1{}, nil, nil, err
	}
	return facts, setBody, bundle, nil
}

func task8OpaqueV7Canonical(value any) ([]byte, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return jcs.Transform(raw)
}

func task8OpaqueV7DomainDigest(schema string, body []byte) contracts.Digest {
	hash := sha256.New()
	_, _ = hash.Write([]byte("talenro.c12." + schema))
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write(body)
	var digest contracts.Digest
	copy(digest[:], hash.Sum(nil))
	return digest
}

func task8OpaqueV7Sign(schema, role string, body []byte) (task8OpaqueV7SignedArtifact, error) {
	digest := task8OpaqueV7DomainDigest(schema, body)
	metadata, err := task8OpaqueV7Canonical(map[string]any{
		"schema":                   schema,
		"body_digest":              task8OpaqueV7DigestText(digest),
		"signer_role":              role,
		"signer_key_id":            task8OpaqueV7SignerKeyID,
		"signature_policy_version": "1",
		"trust_root_digest":        strings.Repeat("a1", 32),
		"signature_algorithm":      "ed25519",
	})
	if err != nil {
		return task8OpaqueV7SignedArtifact{}, err
	}
	signatureInput := append([]byte("talenro.c12.signature-envelope.v1"), 0)
	signatureInput = append(signatureInput, metadata...)
	seed := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31}
	signature := ed25519.Sign(ed25519.NewKeyFromSeed(seed), signatureInput)
	envelope, err := task8OpaqueV7Canonical(map[string]any{
		"schema":                   schema,
		"body":                     json.RawMessage(body),
		"body_digest":              task8OpaqueV7DigestText(digest),
		"signer_role":              role,
		"signer_key_id":            task8OpaqueV7SignerKeyID,
		"signature_policy_version": "1",
		"trust_root_digest":        strings.Repeat("a1", 32),
		"signature_algorithm":      "ed25519",
		"signature":                base64.RawURLEncoding.EncodeToString(signature),
	})
	if err != nil {
		return task8OpaqueV7SignedArtifact{}, err
	}
	return task8OpaqueV7SignedArtifact{schema: schema, digest: digest, body: append([]byte(nil), body...), envelope: envelope}, nil
}

func task8OpaqueV7EvidenceBundle(messageSchema string, messageDigest contracts.Digest, signed []task8OpaqueV7SignedArtifact) ([]byte, error) {
	members := make([]task8OpaqueV7EvidenceMember, len(signed))
	for index, artifact := range signed {
		members[index] = task8OpaqueV7EvidenceMember{
			kind: "external_signed_envelope", schema: artifact.schema, digest: artifact.digest,
			body: artifact.body, envelope: artifact.envelope,
		}
	}
	sort.Slice(members, func(left, right int) bool {
		if compared := bytes.Compare(members[left].digest[:], members[right].digest[:]); compared != 0 {
			return compared < 0
		}
		return members[left].schema < members[right].schema
	})
	items := make([]map[string]any, len(members))
	for index, member := range members {
		items[index] = map[string]any{
			"evidence_kind":              member.kind,
			"schema":                     member.schema,
			"body_digest":                task8OpaqueV7DigestText(member.digest),
			"canonical_body_or_null":     nil,
			"canonical_envelope_or_null": json.RawMessage(member.envelope),
		}
	}
	return task8OpaqueV7Canonical(map[string]any{
		"message_schema":      messageSchema,
		"message_body_digest": task8OpaqueV7DigestText(messageDigest),
		"evidence_count":      strconv.Itoa(len(items)),
		"evidence":            items,
	})
}

func task8OpaqueV7RetiredMemberWire(member contracts.AuthorityV7RetiredEnvironmentMemberFactsV1) map[string]any {
	return map[string]any{
		"environment_record_digest":         task8OpaqueV7DigestText(member.EnvironmentRecordDigest),
		"environment_attestation_digest":    task8OpaqueV7DigestText(member.EnvironmentAttestationDigest),
		"deployment_id":                     member.DeploymentID.String(),
		"postgres_system_id":                strconv.FormatUint(member.PostgresSystemID, 10),
		"timeline":                          strconv.FormatUint(member.Timeline, 10),
		"database_oid":                      strconv.FormatUint(member.DatabaseOID, 10),
		"database_name":                     member.DatabaseName,
		"database_identity_digest":          task8OpaqueV7DigestText(member.DatabaseIdentityDigest),
		"environment_instance_generation":   strconv.FormatUint(member.EnvironmentInstanceGeneration, 10),
		"provider_identity_digest":          task8OpaqueV7DigestText(member.ProviderIdentityDigest),
		"provider_endpoint_identity_digest": task8OpaqueV7DigestText(member.ProviderEndpointIdentityDigest),
		"provider_namespace":                member.ProviderNamespace,
		"provider_profile":                  member.ProviderProfile,
	}
}

func task8OpaqueV7RetirementSetBody(set contracts.AuthorityV7ProviderRetirementSetFactsV1) ([]byte, error) {
	retirements := make([]map[string]any, len(set.Retirements))
	for index, retirement := range set.Retirements {
		retirements[index] = map[string]any{
			"provider_identity_digest":                      task8OpaqueV7DigestText(retirement.ProviderIdentityDigest),
			"provider_endpoint_identity_digest":             task8OpaqueV7DigestText(retirement.ProviderEndpointIdentityDigest),
			"provider_protocol_downgrade_retirement_digest": task8OpaqueV7DigestText(retirement.ProviderProtocolDowngradeRetirementDigest),
		}
	}
	members := make([]map[string]any, len(set.RetiredMembers))
	for index, member := range set.RetiredMembers {
		members[index] = task8OpaqueV7RetiredMemberWire(member)
	}
	return task8OpaqueV7Canonical(map[string]any{
		"installation_id":                                    set.InstallationID.String(),
		"release_scope":                                      set.ReleaseScope,
		"environment_inventory_digest":                       task8OpaqueV7DigestText(set.EnvironmentInventoryDigest),
		"environment_inventory_anchor_set_digest":            task8OpaqueV7DigestText(set.EnvironmentInventoryAnchorSetDigest),
		"environment_inventory_membership_retirement_digest": task8OpaqueV7DigestText(set.EnvironmentInventoryMembershipRetirementDigest),
		"provider_count":                                     strconv.FormatUint(set.ProviderCount, 10),
		"retirements":                                        retirements,
		"retired_member_count":                               strconv.FormatUint(set.RetiredMemberCount, 10),
		"retired_members":                                    members,
		"created_at":                                         task8OpaqueV7Time(set.CreatedAt),
	})
}

func task8OpaqueV7AuthorizationBody(value contracts.AuthorityV7ProtocolDowngradeAuthorizationFactsV1) ([]byte, error) {
	return task8OpaqueV7Canonical(map[string]string{
		"authorization_id":                                  value.AuthorizationID.String(),
		"installation_id":                                   value.InstallationID.String(),
		"migration_latch_digest":                            task8OpaqueV7DigestText(value.MigrationLatchDigest),
		"database_identity_digest":                          task8OpaqueV7DigestText(value.DatabaseIdentityDigest),
		"migration_version":                                 strconv.FormatUint(value.MigrationVersion, 10),
		"current_catalog_digest":                            task8OpaqueV7DigestText(value.CurrentCatalogDigest),
		"pristine_downgrade_inventory_digest":               task8OpaqueV7DigestText(value.PristineDowngradeInventoryDigest),
		"provider_protocol_downgrade_retirement_set_digest": task8OpaqueV7DigestText(value.ProviderProtocolDowngradeRetirementSetDigest),
		"environment_inventory_anchor_set_digest":           task8OpaqueV7DigestText(value.EnvironmentInventoryAnchorSetDigest),
		"database_transaction_id":                           strconv.FormatUint(value.DatabaseTransactionID, 10),
		"transaction_nonce":                                 task8OpaqueV7DigestText(value.TransactionNonce),
		"authorization_scope":                               string(value.AuthorizationScope),
		"issued_at":                                         task8OpaqueV7Time(value.IssuedAt),
		"expires_at":                                        task8OpaqueV7Time(value.ExpiresAt),
	})
}

func task8OpaqueV7DigestText(value contracts.Digest) string { return hex.EncodeToString(value[:]) }

func task8OpaqueV7Time(value time.Time) string { return value.UTC().Format(time.RFC3339Nano) }

func (authorizer *fixtureAuthorityV7DownAuthorizer) CallCount() int {
	authorizer.mu.Lock()
	defer authorizer.mu.Unlock()
	return authorizer.calls
}

func disposableAuthorityV7DownFacts(now time.Time, label string) contracts.AuthorityV7DownMigrationFactsV1 {
	up := disposableAuthorityV7UpFacts(now, label)
	digest := func(value string) contracts.Digest {
		return contracts.Digest(sha256.Sum256([]byte("task8-down:" + label + ":" + value)))
	}
	retirement := contracts.AuthorityV7ProviderRetirementSetFactsV1{
		InstallationID:                                 up.MigrationLatch.InstallationID,
		ReleaseScope:                                   "release/task8-integration",
		EnvironmentInventoryDigest:                     digest("environment-inventory"),
		EnvironmentInventoryAnchorSetDigest:            digest("environment-anchor-set"),
		EnvironmentInventoryMembershipRetirementDigest: digest("membership-retirement"),
		ProviderCount:                                  1,
		Retirements: []contracts.AuthorityV7ProviderRetirementFactsV1{{
			ProviderIdentityDigest:                    digest("provider-identity"),
			ProviderEndpointIdentityDigest:            digest("provider-endpoint"),
			ProviderProtocolDowngradeRetirementDigest: digest("provider-retirement"),
		}},
		RetiredMemberCount: 1,
		RetiredMembers: []contracts.AuthorityV7RetiredEnvironmentMemberFactsV1{{
			EnvironmentRecordDigest:        digest("environment-record"),
			EnvironmentAttestationDigest:   digest("environment-attestation"),
			DeploymentID:                   uuid.NewSHA1(uuid.NameSpaceOID, []byte(label+":deployment")),
			PostgresSystemID:               1,
			Timeline:                       1,
			DatabaseOID:                    1,
			DatabaseName:                   "task8_fixture",
			DatabaseIdentityDigest:         up.MigrationLatch.DatabaseIdentityDigest,
			EnvironmentInstanceGeneration:  1,
			ProviderIdentityDigest:         digest("provider-identity"),
			ProviderEndpointIdentityDigest: digest("provider-endpoint"),
			ProviderNamespace:              "task8/fixture",
			ProviderProfile:                "legacy_v6",
		}},
		CreatedAt: now,
	}
	inventory := make([]contracts.AuthorityV7StableTableInventoryItemV1, len(task8PristineDownRegistry))
	for index, relation := range task8PristineDownRegistry {
		contentDigestBytes, err := hex.DecodeString(task8PristineEmptyTableDigests[index])
		if err != nil || len(contentDigestBytes) != 32 {
			panic("invalid independently reviewed pristine empty-table digest literal")
		}
		var contentDigest contracts.Digest
		copy(contentDigest[:], contentDigestBytes)
		inventory[index] = contracts.AuthorityV7StableTableInventoryItemV1{
			TableName: relation.table, Classification: relation.classification, RowCount: 0,
			ContentDigest: contentDigest,
		}
	}
	return contracts.AuthorityV7DownMigrationFactsV1{
		MigrationLatch:              up.MigrationLatch,
		MigrationLatchDigest:        up.MigrationLatchDigest,
		AuthorityProtocolProfile:    contracts.AuthorityV7ProtocolProfileLegacyV6,
		CurrentCatalogDigest:        up.MigrationLatch.UpCatalogDigest,
		ManifestID:                  uuid.NewSHA1(uuid.NameSpaceOID, []byte(label+":manifest")),
		ManifestDigest:              digest("manifest"),
		StableTableCount:            uint64(len(inventory)),
		StableTableInventory:        inventory,
		ProviderRetirementSet:       retirement,
		ProviderRetirementSetDigest: digest("provider-retirement-set"),
		EnvironmentAnchorSetDigest:  retirement.EnvironmentInventoryAnchorSetDigest,
		AuthorizationID:             uuid.NewSHA1(uuid.NameSpaceOID, []byte(label+":authorization")),
		AuthorizationScope:          contracts.AuthorityV7DowngradeAuthorizationScopeDown00007Only,
		TransactionNonce:            digest("transaction-nonce"),
		ExpiresAt:                   now.Add(5 * time.Minute),
	}
}

// These are independently hand-derived SHA-256 values of
// ASCII("talenro.c12.pristine-downgrade-empty-table.v1") || 0x00 ||
// the exact literal JCS
// {"classification":...,"row_count":"0","table_name":...}.  Their order is
// exactly task8PristineDownRegistry; production code is not used as an oracle.
var task8PristineEmptyTableDigests = [...]string{
	"c470e6bbd2dc4691fddf1c1027834e4b4b0a1447c99e4891a04b65f10be6baec",
	"a026121f25226f6137eedf94a99cba51b6dee159bd0a162477b75b62696a35df",
	"db112af69f80c0937d449eaae61371657fb3224f6c24e736b998cb7fddfd3077",
	"9a23082d1f403fef5f3cf85db3f3db38192fcd90ccd0de31dd1651cbcbe26ee8",
	"a88d359fe5761bc8aff7eda66ac60a0fb0f42188cd14a5302816cf8e422b7c67",
	"07283dadb0b0285db5611cf276f53a4ef30318a31136dddb0f484baa0b1e2ec8",
	"1db6628334bdacee4bd733b03b27fbade9125c319c4c015494a922b93f560cbb",
	"d14e9687d2d66d2276e2365862a34d0d183105ab2cf5e0feab9709e3c547ca66",
	"183d81dfd4db0258aa2f84228d9942742ba58c6076d12978213b1a748033fce3",
	"825ecca5cb0349082c388089dab8f3d1e24081ddc064d63a0694ac1ad4150b04",
	"d6ce42c8192d1595fe1c3ee435fe49b320f2671d933c7a45e1d4f4cb1ea92cad",
	"522790b760dc4990af7225156479a6e1f5cc2345a80d94da2687ea1752b9b8ad",
	"172343f72e66c2fc328c9d44ef888c4afb8dc65c2abaece0243cd412b2abe560",
	"138e333e806e7d26765905a1a1c606fbb468814063ef30697a313d83c2ca18ac",
	"b93b23529c81d20d2c04f338e55e65ec92503676182bf2585a2deb4c9a7ae13d",
	"36bb32fba9e5300a5561fc925e9663a8dad3e2bbcee622ef63fdf194e0d3b072",
	"fdb168e5f9289b8d740c5d7f8b4b082b021b86f9dbf04a7c0d1aca57ef7e08c5",
	"ee00235fce353236a8df1eead343a9a01c3a9180c94276e964b18fa656d5f94d",
	"a3967798dc3d923cfb8b23e216fa19209ffe0cbd7f7441fb71ffb7f63cdb3850",
	"0b6b570464e3292f7feeab974f3682f7ea5efc3c646bff4ccb9f73077f08042b",
	"e3908de845af23894e0ef5f015c1f4ab4806b066f13bfffc69d8ab582c94a335",
	"50db77a90b480556c690d41973c6ed166f0d8cf598fba626a08326d7393a267c",
	"a8258fa01e3fb00cd9db20dce0d5623917bfef47cb75c000fa101b0132767982",
	"eb613ed486b8a64907828f97708614ae0e2266a36de750ca8d69e9214a89e98f",
	"b6389762cbc7ddcac89ed193b43c655d8ea740d8a13f3faa4f73fd0eba418a29",
	"66355cf2a8ddff20be277ce8c582d56f8b0c25a194edd044d3a95cda62f97697",
	"5392f9545a8fcb997055620bfd9891440b74d795e6c35b909d50a1eba5e1b027",
	"2973cf13c70186a1fc3b771259770bf3eae1f5ca0413560d0abb2f3155b3e46f",
	"e334d1637a7384553bce6b77345dfc5e3e21b1ac7f0806fd5acd99bc6468a4a2",
	"b7e76507a860d8467c1e1cf536b4cd0f24e45eb483be6586adf7c563acaf1bd5",
	"c56f5f755671bf33d71e18d6c3104b9dbf7c9e2aa2b4d791839eb8a4d0b8a601",
	"0b7633e103c2345e7f9d6519836e954c2968aedff45e4f9ec057ed24f90034bf",
	"18d208ba4311f0576c3a7d2e255508c97e287509b901bb1095bcb9b85db84a3d",
	"134ad9d8580d113d7d343cd2415d110f0ac1f473cc9fd167d1a648f144bc6fdc",
	"fadc46ea39115458e5fe803a414a6ab7dd3559437b638dca4b852ee9954f80e3",
	"2188ea8b98b418744c754c5734a61f03f4c9d9e7d6f533fd4999e62bfc86b649",
	"9e8752214f3cad944051542c1f3d5a311c221faaa4235ec54149fc37c16ec223",
	"31a068ec186dc7e4bf8cc403d472e2370d858abe25f5e78d4e12102abddcd953",
	"93679e90cb9bf666b8e858de76eb4077cec3b07687e52efc53e6b70d85bde3f9",
	"b37ac2eab5314fd2b043b91b1967dae6992ca8dd1290f6005733f4baf91dd4fc",
	"22c4a4420bc48c579d13c367c170aa49827b53f94c4351108959d2adfbb052d2",
	"bef3fed3db433c096f18c2a090b8b75151e335bea55d308459b1f2c971c26f70",
	"6545a4104c60668f498ee447e55b7af006cc3e5942f4e113266f0e7a038ebcb7",
	"6c269c596b0c81ab0b682886ba6d35a3a451a5302ea2c2f4ceb22a16fd9dbb8e",
	"8ccd9e536e4c25ca5053d7e53b22865428192e036458d5a030868d29287e81cc",
	"7cd1fd5abae9dbe6d5acf5325f4dfcde63bf2efc6742174b27283659a436cbcc",
	"8289e4b3575dde13d48ad4f7ab7aa43f42cd59c5c00de19ddb5cb6226506bd90",
	"dafddde4dd9704451f9a000f1be4f1fb73d6404c5e6238b8ad32aee7b0ba409c",
	"6615d3c0cc4a994c3793eb24d72ef25d9713eb221ec54ff492f88512c3d2b296",
}

func assertAuthorityV7ControlCardinality(t *testing.T, database interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, wantLatch, wantAuthorization int) {
	t.Helper()
	var latchCount, authorizationCount int
	if err := database.QueryRowContext(t.Context(), `SELECT (SELECT count(*) FROM nodecontrol.control_plane_authority_protocol_migration_latches), (SELECT count(*) FROM nodecontrol.control_plane_authority_protocol_downgrade_authorizations)`).Scan(&latchCount, &authorizationCount); err != nil {
		t.Fatal("inspect authority-v7 control cardinality:", err)
	}
	if latchCount != wantLatch || authorizationCount != wantAuthorization {
		t.Fatalf("authority-v7 control cardinality = %d/%d, want %d/%d", latchCount, authorizationCount, wantLatch, wantAuthorization)
	}
}
