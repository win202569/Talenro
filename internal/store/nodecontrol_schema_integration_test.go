//go:build integration

package store_test

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var expectedNodeControlFunctions = []string{
	"bytea_array_is_sorted_unique_32",
	"enforce_authority_fence_update",
	"enforce_capacity_profile_immutability",
	"enforce_certificate_issuance_workflow",
	"enforce_certificate_workflow",
	"enforce_enrollment_grant_workflow",
	"enforce_inventory_pointers",
	"enforce_observed_state",
	"enforce_process_slot_cap",
	"enforce_recovery_session_workflow",
	"enforce_restore_approval_workflow",
	"enforce_root_publish_workflow",
	"enforce_root_share_binding",
	"enforce_security_fault_receipt",
	"enforce_security_incident_workflow",
	"enforce_signing_intent_workflow",
	"enforce_trust_bundle_high_water",
	"reject_row_mutation",
	"text_array_is_sorted_unique",
}

func TestNodeControlMigrationUpDownUp(t *testing.T) {
	manifest, _ := loadNodeControlManifest(t)
	upSQL, downSQL := loadNodeControlMigrationSections(t)
	pool := openOwnedNodeControlDatabase(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	var version string
	if err := pool.QueryRow(ctx, `SHOW server_version`).Scan(&version); err != nil {
		t.Fatal("read PostgreSQL version:", err)
	}
	if version != "18.4" {
		t.Fatalf("PostgreSQL version = %q, want exactly 18.4", version)
	}
	if _, err := pool.Exec(ctx, upSQL); err != nil {
		t.Fatal("first nodecontrol Up failed:", err)
	}
	assertNodeControlCatalogMatchesManifest(ctx, t, pool, manifest)
	firstDigest := nodeControlCatalogDigest(ctx, t, pool)
	assertAuthorityFenceRejectsForbiddenMutations(ctx, t, pool)

	if _, err := pool.Exec(ctx, downSQL); err != nil {
		t.Fatal("nodecontrol Down failed:", err)
	}
	var schemaExists bool
	if err := pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname='nodecontrol')`).Scan(&schemaExists); err != nil {
		t.Fatal("inspect schema after Down:", err)
	}
	if schemaExists {
		t.Fatal("nodecontrol schema remains after Down")
	}
	if _, err := pool.Exec(ctx, upSQL); err != nil {
		t.Fatal("second nodecontrol Up failed:", err)
	}
	assertNodeControlCatalogMatchesManifest(ctx, t, pool, manifest)
	secondDigest := nodeControlCatalogDigest(ctx, t, pool)
	if firstDigest != secondDigest {
		t.Fatalf("Up/Down/Up catalog digest changed: first=%s second=%s", firstDigest, secondDigest)
	}
	t.Logf("PostgreSQL %s first/second Up catalog digest: %s", version, firstDigest)
}

func TestNodeControlMigrationCatalogEnforcesCapsAndImmutableRows(t *testing.T) {
	manifest, _ := loadNodeControlManifest(t)
	upSQL, _ := loadNodeControlMigrationSections(t)
	pool := openOwnedNodeControlDatabase(t)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	if _, err := pool.Exec(ctx, upSQL); err != nil {
		t.Fatal("nodecontrol Up failed:", err)
	}
	assertNodeControlCatalogMatchesManifest(ctx, t, pool, manifest)
	assertSlotCapAndProfileImmutability(ctx, t, pool)
}

func TestNodeControlMigrationEnforcesRootPublishThresholds(t *testing.T) {
	manifest, _ := loadNodeControlManifest(t)
	upSQL, _ := loadNodeControlMigrationSections(t)
	pool := openOwnedNodeControlDatabase(t)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	if _, err := pool.Exec(ctx, upSQL); err != nil {
		t.Fatal("nodecontrol Up failed:", err)
	}
	assertNodeControlCatalogMatchesManifest(ctx, t, pool, manifest)

	publishID := uuid.New()
	operationID := uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)
	insertCommittedFence(ctx, t, pool, operationID, 1, "root_publish", "global_node_trust", now)
	keyID := bytesOf(0x11, 32)
	payloadDigest := bytesOf(0x22, 32)
	if _, err := pool.Exec(ctx, `
INSERT INTO nodecontrol.node_root_metadata_publish_intents(
  publish_id,authority_operation_id,authority_epoch,authority_sequence,publish_kind,reason,
  base_root_version,base_metadata_version,reserved_version,canonical_payload,payload_digest,
  key_set_digest,current_key_ids,current_threshold,activation_deadline,published_envelope,
  published_envelope_digest,status,created_at,updated_at,terminal_at)
VALUES($1,$2,1,1,'root','normal',0,0,1,$3,$4,$5,ARRAY[$6::bytea],1,$7,$8,$9,'active',$10,$10,$10)`,
		uuid.New(), operationID, []byte{0x01}, payloadDigest, bytesOf(0x33, 32), keyID,
		now.Add(5*time.Minute), []byte{0x02}, bytesOf(0x44, 32), now); err == nil {
		t.Fatal("root publish workflow accepted a directly inserted active intent")
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO nodecontrol.node_root_metadata_publish_intents(
  publish_id,authority_operation_id,authority_epoch,authority_sequence,publish_kind,reason,
  base_root_version,base_metadata_version,reserved_version,canonical_payload,payload_digest,
  key_set_digest,current_key_ids,current_threshold,activation_deadline,status,created_at,updated_at)
VALUES($1,$2,1,1,'root','normal',0,0,1,$3,$4,$5,ARRAY[$6::bytea],1,$7,'pending',$8,$8)`,
		publishID, operationID, []byte{0x01}, payloadDigest, bytesOf(0x33, 32), keyID, now.Add(5*time.Minute), now); err != nil {
		t.Fatal("insert pending root publish:", err)
	}
	activate := `UPDATE nodecontrol.node_root_metadata_publish_intents
SET published_envelope=$2,published_envelope_digest=$3,status='active',terminal_at=$4,updated_at=$4
WHERE publish_id=$1`
	if _, err := pool.Exec(ctx, activate, publishID, []byte{0x02}, bytesOf(0x44, 32), now.Add(time.Second)); err == nil {
		t.Fatal("root publish activated without the captured current-root threshold")
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO nodecontrol.node_root_metadata_signature_shares(
  publish_id,key_id,physical_key_id,payload_digest,signature_role,signature,verified_at)
VALUES($1,$2,$3,$4,'current_root',$5,$6)`,
		publishID, keyID, bytesOf(0x55, 32), payloadDigest, bytesOf(0x66, 64), now); err != nil {
		t.Fatal("insert threshold share:", err)
	}
	if _, err := pool.Exec(ctx, activate, publishID, []byte{0x02}, bytesOf(0x44, 32), now.Add(time.Second)); err != nil {
		t.Fatal("root publish with its captured threshold failed:", err)
	}
}

func TestNodeControlMigrationRejectsTerminalSigningIntentInsert(t *testing.T) {
	manifest, _ := loadNodeControlManifest(t)
	upSQL, _ := loadNodeControlMigrationSections(t)
	pool := openOwnedNodeControlDatabase(t)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	if _, err := pool.Exec(ctx, upSQL); err != nil {
		t.Fatal("nodecontrol Up failed:", err)
	}
	assertNodeControlCatalogMatchesManifest(ctx, t, pool, manifest)

	now := time.Now().UTC().Truncate(time.Microsecond)
	nodeID := uuid.New()
	operationID := uuid.New()
	insertCommittedFence(ctx, t, pool, operationID, 1, "desired_activate", "node", now)
	if _, err := pool.Exec(ctx, `INSERT INTO nodecontrol.node_pops(pop_code,iso_country,region,operator_state,created_at,updated_at) VALUES('intent-test','US','test-region','enabled',$1,$1)`, now); err != nil {
		t.Fatal("insert signing-intent POP:", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO nodecontrol.node_inventory(node_id,pop_code,operator_state,security_state,identity_state,created_at,updated_at) VALUES($1,'intent-test','enabled','normal','never_enrolled',$2,$2)`, nodeID, now); err != nil {
		t.Fatal("insert signing-intent node:", err)
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO nodecontrol.node_state_signing_intents(
  signing_id,authority_operation_id,authority_epoch,authority_sequence,node_id,signing_kind,
  idempotency_key_digest,base_generation,reserved_generation,canonical_payload,payload_digest,
  root_version,metadata_version,expected_key_id,expected_public_key_digest,captured_inventory_version,
  captured_identity_epoch,captured_security_version,signature,signature_verified_at,activation_deadline,
  status,created_at,updated_at,terminal_at)
VALUES($1,$2,1,1,$3,'desired',$4,0,1,$5,$6,1,1,$7,$8,1,0,1,$9,$10,$11,'active',$10,$10,$10)`,
		uuid.New(), operationID, nodeID, bytesOf(0x10, 32), []byte{0x01}, bytesOf(0x20, 32),
		bytesOf(0x30, 32), bytesOf(0x40, 32), bytesOf(0x50, 64), now, now.Add(5*time.Minute)); err == nil {
		t.Fatal("state signing workflow accepted a directly inserted active intent")
	}
}

func TestNodeControlMigrationRejectsWrongFenceScopeAtInventoryPointer(t *testing.T) {
	manifest, _ := loadNodeControlManifest(t)
	upSQL, _ := loadNodeControlMigrationSections(t)
	pool := openOwnedNodeControlDatabase(t)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	if _, err := pool.Exec(ctx, upSQL); err != nil {
		t.Fatal("nodecontrol Up failed:", err)
	}
	assertNodeControlCatalogMatchesManifest(ctx, t, pool, manifest)

	now := time.Now().UTC().Truncate(time.Microsecond)
	nodeID := uuid.New()
	operationID := uuid.New()
	insertCommittedFence(ctx, t, pool, operationID, 1, "root_publish", "global_node_trust", now)
	if _, err := pool.Exec(ctx, `INSERT INTO nodecontrol.node_pops(pop_code,iso_country,region,operator_state,created_at,updated_at) VALUES('pointer-test','US','test-region','enabled',$1,$1)`, now); err != nil {
		t.Fatal("insert pointer-test POP:", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO nodecontrol.node_inventory(node_id,pop_code,operator_state,security_state,identity_state,created_at,updated_at) VALUES($1,'pointer-test','enabled','normal','never_enrolled',$2,$2)`, nodeID, now); err != nil {
		t.Fatal("insert pointer-test node:", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE nodecontrol.node_inventory SET last_authority_operation_id=$2,last_authority_epoch=1,last_authority_sequence=1,updated_at=$3 WHERE node_id=$1`, nodeID, operationID, now.Add(time.Second)); err == nil {
		t.Fatal("inventory accepted a global-scope fence as its per-node authority pointer")
	}
}

func loadNodeControlMigrationSections(t *testing.T) (string, string) {
	t.Helper()
	raw, err := os.ReadFile("../../db/migrations/00006_nodecontrol.sql")
	if err != nil {
		t.Fatal(err)
	}
	body := string(raw)
	upMarker := strings.Index(body, "-- +goose Up")
	downMarker := strings.Index(body, "-- +goose Down")
	if upMarker < 0 || downMarker <= upMarker {
		t.Fatal("migration lacks ordered goose sections")
	}
	return body[upMarker+len("-- +goose Up") : downMarker], body[downMarker+len("-- +goose Down"):]
}

func openOwnedNodeControlDatabase(t *testing.T) *pgxpool.Pool {
	t.Helper()
	databaseURL := os.Getenv("TALENRO_DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("TALENRO_DATABASE_URL is required for integration tests")
	}
	parsed, err := url.Parse(databaseURL)
	if err != nil {
		t.Fatal("parse TALENRO_DATABASE_URL:", err)
	}
	adminConfig, err := pgx.ParseConfig(databaseURL)
	if err != nil {
		t.Fatal("parse PostgreSQL config:", err)
	}
	admin, err := pgx.ConnectConfig(context.Background(), adminConfig)
	if err != nil {
		t.Fatal("connect PostgreSQL admin database:", err)
	}
	databaseName := "nodecontrol_task4_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	databaseIdentifier := pgx.Identifier{databaseName}.Sanitize()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if _, err := admin.Exec(ctx, "CREATE DATABASE "+databaseIdentifier); err != nil {
		admin.Close(context.Background())
		t.Fatal("create owned integration database:", err)
	}
	admin.Close(context.Background())

	parsed.Path = "/" + databaseName
	ownedConfig, err := pgxpool.ParseConfig(parsed.String())
	if err != nil {
		t.Fatal("parse owned database config:", err)
	}
	pool, err := pgxpool.NewWithConfig(context.Background(), ownedConfig)
	if err != nil {
		t.Fatal("open owned integration database:", err)
	}
	t.Cleanup(func() {
		pool.Close()
		cleanup, cleanupErr := pgx.ConnectConfig(context.Background(), adminConfig)
		if cleanupErr != nil {
			t.Error("reconnect for owned database cleanup:", cleanupErr)
			return
		}
		defer cleanup.Close(context.Background())
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cleanupCancel()
		if _, cleanupErr = cleanup.Exec(cleanupCtx, "DROP DATABASE "+databaseIdentifier); cleanupErr != nil {
			t.Error("drop owned integration database:", cleanupErr)
		}
	})
	return pool
}

func assertNodeControlCatalogMatchesManifest(ctx context.Context, t *testing.T, pool *pgxpool.Pool, manifest nodeControlManifest) {
	t.Helper()
	rows, err := pool.Query(ctx, `
SELECT c.relname
FROM pg_catalog.pg_class c
JOIN pg_catalog.pg_namespace n ON n.oid=c.relnamespace
WHERE n.nspname='nodecontrol' AND c.relkind='r'
ORDER BY c.relname`)
	if err != nil {
		t.Fatal("query nodecontrol tables:", err)
	}
	actualTables, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		t.Fatal("collect nodecontrol tables:", err)
	}
	assertExactNamedSet(t, "catalog tables", actualTables, nodeControlAuthorityTables)
	for _, table := range manifest.Tables {
		assertCatalogColumns(ctx, t, pool, table)
		assertCatalogConstraints(ctx, t, pool, table)
		assertCatalogIndexes(ctx, t, pool, table)
		assertCatalogTriggers(ctx, t, pool, table)
	}
	rows, err = pool.Query(ctx, `
SELECT p.proname
FROM pg_catalog.pg_proc p
JOIN pg_catalog.pg_namespace n ON n.oid=p.pronamespace
WHERE n.nspname='nodecontrol'
ORDER BY p.proname`)
	if err != nil {
		t.Fatal("query nodecontrol functions:", err)
	}
	functions, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		t.Fatal("collect nodecontrol functions:", err)
	}
	assertExactNamedSet(t, "catalog functions", functions, expectedNodeControlFunctions)
}

func assertCatalogColumns(ctx context.Context, t *testing.T, pool *pgxpool.Pool, table nodeControlTableSpec) {
	t.Helper()
	rows, err := pool.Query(ctx, `
SELECT a.attname,
       pg_catalog.format_type(a.atttypid,a.atttypmod),
       NOT a.attnotnull,
       pg_catalog.pg_get_expr(d.adbin,d.adrelid),
       CASE WHEN a.attcollation=0 THEN NULL ELSE coll.collname END
FROM pg_catalog.pg_attribute a
JOIN pg_catalog.pg_class c ON c.oid=a.attrelid
JOIN pg_catalog.pg_namespace n ON n.oid=c.relnamespace
LEFT JOIN pg_catalog.pg_attrdef d ON d.adrelid=a.attrelid AND d.adnum=a.attnum
LEFT JOIN pg_catalog.pg_collation coll ON coll.oid=a.attcollation
WHERE n.nspname='nodecontrol' AND c.relname=$1 AND a.attnum>0 AND NOT a.attisdropped
ORDER BY a.attnum`, table.Name)
	if err != nil {
		t.Fatal("query columns for "+table.Name+":", err)
	}
	type catalogColumn struct {
		name       string
		sqlType    string
		nullable   bool
		defaultSQL sql.NullString
		collation  sql.NullString
	}
	actual := make([]catalogColumn, 0, len(table.Columns))
	for rows.Next() {
		var column catalogColumn
		if err := rows.Scan(&column.name, &column.sqlType, &column.nullable, &column.defaultSQL, &column.collation); err != nil {
			rows.Close()
			t.Fatal("scan column for "+table.Name+":", err)
		}
		actual = append(actual, column)
	}
	if err := rows.Err(); err != nil {
		t.Fatal("iterate columns for "+table.Name+":", err)
	}
	rows.Close()
	if len(actual) != len(table.Columns) {
		t.Fatalf("catalog %s column count = %d, manifest = %d", table.Name, len(actual), len(table.Columns))
	}
	for index, want := range table.Columns {
		got := actual[index]
		if got.name != want.Name || got.sqlType != want.SQLType || got.nullable != want.Nullable || !equalOptionalString(got.defaultSQL, want.DefaultSQL) || !equalOptionalString(got.collation, want.Collation) {
			t.Fatalf("catalog column %s[%d] = {name:%q type:%q nullable:%t default:%v collation:%v}, manifest = %+v", table.Name, index, got.name, got.sqlType, got.nullable, got.defaultSQL, got.collation, want)
		}
	}
}

func assertCatalogConstraints(ctx context.Context, t *testing.T, pool *pgxpool.Pool, table nodeControlTableSpec) {
	t.Helper()
	rows, err := pool.Query(ctx, `
SELECT con.conname, con.contype::text, pg_catalog.pg_get_constraintdef(con.oid,true),
       con.condeferrable, con.condeferred,
       COALESCE(ARRAY(SELECT att.attname FROM unnest(con.conkey) WITH ORDINALITY key(attnum,ord)
                      JOIN pg_catalog.pg_attribute att ON att.attrelid=con.conrelid AND att.attnum=key.attnum ORDER BY key.ord), ARRAY[]::name[])::text[],
       ref.relname,
       COALESCE(ARRAY(SELECT att.attname FROM unnest(con.confkey) WITH ORDINALITY key(attnum,ord)
                      JOIN pg_catalog.pg_attribute att ON att.attrelid=con.confrelid AND att.attnum=key.attnum ORDER BY key.ord), ARRAY[]::name[])::text[],
       con.confupdtype::text, con.confdeltype::text
FROM pg_catalog.pg_constraint con
JOIN pg_catalog.pg_class rel ON rel.oid=con.conrelid
JOIN pg_catalog.pg_namespace n ON n.oid=rel.relnamespace
LEFT JOIN pg_catalog.pg_class ref ON ref.oid=con.confrelid
WHERE n.nspname='nodecontrol' AND rel.relname=$1 AND con.contype <> 't'
ORDER BY con.conname`, table.Name)
	if err != nil {
		t.Fatal("query constraints for "+table.Name+":", err)
	}
	type catalogConstraint struct {
		name, kind, definition        string
		deferrable, initiallyDeferred bool
		columns                       []string
		referencedTable               sql.NullString
		referencedColumns             []string
		onUpdateCode, onDeleteCode    string
	}
	actual := make(map[string]catalogConstraint)
	for rows.Next() {
		var constraint catalogConstraint
		if err := rows.Scan(&constraint.name, &constraint.kind, &constraint.definition, &constraint.deferrable, &constraint.initiallyDeferred, &constraint.columns, &constraint.referencedTable, &constraint.referencedColumns, &constraint.onUpdateCode, &constraint.onDeleteCode); err != nil {
			rows.Close()
			t.Fatal("scan constraint for "+table.Name+":", err)
		}
		actual[constraint.name] = constraint
	}
	if err := rows.Err(); err != nil {
		t.Fatal("iterate constraints for "+table.Name+":", err)
	}
	rows.Close()
	wantNames := append([]string{table.PrimaryKey.Name}, constraintNames(table.Constraints)...)
	actualNames := make([]string, 0, len(actual))
	for name := range actual {
		actualNames = append(actualNames, name)
	}
	assertExactNamedSet(t, table.Name+" catalog constraints", actualNames, wantNames)
	primary := actual[table.PrimaryKey.Name]
	if primary.kind != "p" || !equalStrings(primary.columns, table.PrimaryKey.Columns) || primary.deferrable != table.PrimaryKey.Deferrable || primary.initiallyDeferred != table.PrimaryKey.InitiallyDeferred {
		t.Fatalf("catalog primary key %s does not match manifest", table.PrimaryKey.Name)
	}
	for _, want := range table.Constraints {
		got := actual[want.Name]
		if catalogConstraintKind(got.kind) != want.Kind || normalizeCatalogSQL(got.definition) != normalizeCatalogSQL(want.DefinitionSQL) || !equalStrings(got.columns, want.Columns) || got.deferrable != want.Deferrable || got.initiallyDeferred != want.InitiallyDeferred {
			t.Fatalf("catalog constraint %s mismatch: got kind=%s sql=%q columns=%v deferrable=%t deferred=%t; manifest=%+v", want.Name, catalogConstraintKind(got.kind), got.definition, got.columns, got.deferrable, got.initiallyDeferred, want)
		}
		if want.Kind == "foreign_key" {
			if !got.referencedTable.Valid || want.ReferencedTable == nil || got.referencedTable.String != *want.ReferencedTable || !equalStrings(got.referencedColumns, want.ReferencedColumns) || decodeFKAction(got.onUpdateCode) != *want.OnUpdate || decodeFKAction(got.onDeleteCode) != *want.OnDelete {
				t.Fatalf("catalog foreign key %s reference/action mismatch", want.Name)
			}
		}
	}
}

func assertCatalogIndexes(ctx context.Context, t *testing.T, pool *pgxpool.Pool, table nodeControlTableSpec) {
	t.Helper()
	rows, err := pool.Query(ctx, `
SELECT idx.relname, i.indisunique, am.amname, i.indnatts, i.indnkeyatts,
       pg_catalog.pg_get_expr(i.indpred,i.indrelid,true)
FROM pg_catalog.pg_index i
JOIN pg_catalog.pg_class rel ON rel.oid=i.indrelid
JOIN pg_catalog.pg_namespace n ON n.oid=rel.relnamespace
JOIN pg_catalog.pg_class idx ON idx.oid=i.indexrelid
JOIN pg_catalog.pg_am am ON am.oid=idx.relam
WHERE n.nspname='nodecontrol' AND rel.relname=$1 AND NOT i.indisprimary
  AND NOT EXISTS (SELECT 1 FROM pg_catalog.pg_constraint con WHERE con.conindid=i.indexrelid)
ORDER BY idx.relname`, table.Name)
	if err != nil {
		t.Fatal("query indexes for "+table.Name+":", err)
	}
	type catalogIndex struct {
		name            string
		unique          bool
		method          string
		total, keyCount int
		predicate       sql.NullString
	}
	actual := make(map[string]catalogIndex)
	for rows.Next() {
		var index catalogIndex
		if err := rows.Scan(&index.name, &index.unique, &index.method, &index.total, &index.keyCount, &index.predicate); err != nil {
			rows.Close()
			t.Fatal("scan index for "+table.Name+":", err)
		}
		actual[index.name] = index
	}
	if err := rows.Err(); err != nil {
		t.Fatal("iterate indexes for "+table.Name+":", err)
	}
	rows.Close()
	actualNames := make([]string, 0, len(actual))
	for name := range actual {
		actualNames = append(actualNames, name)
	}
	assertExactNamedSet(t, table.Name+" catalog indexes", actualNames, indexNames(table.Indexes))
	for _, want := range table.Indexes {
		got := actual[want.Name]
		keys := make([]string, 0, got.keyCount)
		include := make([]string, 0, got.total-got.keyCount)
		for position := 1; position <= got.total; position++ {
			var definition string
			var descending bool
			if err := pool.QueryRow(ctx, `SELECT pg_catalog.pg_get_indexdef(i.indexrelid,$2,true), CASE WHEN $2 <= i.indnkeyatts THEN (i.indoption[$2-1] & 1) = 1 ELSE false END FROM pg_catalog.pg_index i JOIN pg_catalog.pg_class idx ON idx.oid=i.indexrelid JOIN pg_catalog.pg_namespace n ON n.oid=idx.relnamespace WHERE n.nspname='nodecontrol' AND idx.relname=$1`, want.Name, position).Scan(&definition, &descending); err != nil {
				t.Fatal("read index key "+want.Name+":", err)
			}
			if descending {
				definition += " DESC"
			}
			if position <= got.keyCount {
				keys = append(keys, normalizeCatalogSQL(definition))
			} else {
				include = append(include, normalizeCatalogSQL(definition))
			}
		}
		if got.unique != want.Unique || got.method != want.Method || !equalStrings(keys, normalizedStrings(want.Keys)) || !equalStrings(include, normalizedStrings(want.Include)) || !equalOptionalNormalizedSQL(got.predicate, want.Predicate) {
			t.Fatalf("catalog index %s mismatch: got unique=%t method=%s keys=%v include=%v predicate=%v, manifest=%+v", want.Name, got.unique, got.method, keys, include, got.predicate, want)
		}
	}
}

func assertCatalogTriggers(ctx context.Context, t *testing.T, pool *pgxpool.Pool, table nodeControlTableSpec) {
	t.Helper()
	rows, err := pool.Query(ctx, `
SELECT trg.tgname, trg.tgtype::integer, proc.proname || '()',
       pg_catalog.pg_get_expr(trg.tgqual,trg.tgrelid,true)
FROM pg_catalog.pg_trigger trg
JOIN pg_catalog.pg_class rel ON rel.oid=trg.tgrelid
JOIN pg_catalog.pg_namespace n ON n.oid=rel.relnamespace
JOIN pg_catalog.pg_proc proc ON proc.oid=trg.tgfoid
WHERE n.nspname='nodecontrol' AND rel.relname=$1 AND NOT trg.tgisinternal
ORDER BY trg.tgname`, table.Name)
	if err != nil {
		t.Fatal("query triggers for "+table.Name+":", err)
	}
	type catalogTrigger struct {
		name, function string
		typeBits       int
		whenSQL        sql.NullString
	}
	actual := make(map[string]catalogTrigger)
	for rows.Next() {
		var trigger catalogTrigger
		if err := rows.Scan(&trigger.name, &trigger.typeBits, &trigger.function, &trigger.whenSQL); err != nil {
			rows.Close()
			t.Fatal("scan trigger for "+table.Name+":", err)
		}
		actual[trigger.name] = trigger
	}
	if err := rows.Err(); err != nil {
		t.Fatal("iterate triggers for "+table.Name+":", err)
	}
	rows.Close()
	actualNames := make([]string, 0, len(actual))
	for name := range actual {
		actualNames = append(actualNames, name)
	}
	assertExactNamedSet(t, table.Name+" catalog triggers", actualNames, triggerNames(table.Triggers))
	for _, want := range table.Triggers {
		got := actual[want.Name]
		if triggerTiming(got.typeBits) != want.Timing || !equalStrings(triggerEvents(got.typeBits), want.Events) || got.function != want.Function || !equalOptionalNormalizedSQL(got.whenSQL, want.WhenSQL) {
			t.Fatalf("catalog trigger %s mismatch: timing=%s events=%v function=%s when=%v, manifest=%+v", want.Name, triggerTiming(got.typeBits), triggerEvents(got.typeBits), got.function, got.whenSQL, want)
		}
	}
}

func nodeControlCatalogDigest(ctx context.Context, t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()
	queries := []string{
		`SELECT 'column|'||c.relname||'|'||a.attnum||'|'||a.attname||'|'||pg_catalog.format_type(a.atttypid,a.atttypmod)||'|'||a.attnotnull||'|'||COALESCE(pg_catalog.pg_get_expr(d.adbin,d.adrelid),'')||'|'||COALESCE(coll.collname,'') FROM pg_catalog.pg_attribute a JOIN pg_catalog.pg_class c ON c.oid=a.attrelid JOIN pg_catalog.pg_namespace n ON n.oid=c.relnamespace LEFT JOIN pg_catalog.pg_attrdef d ON d.adrelid=a.attrelid AND d.adnum=a.attnum LEFT JOIN pg_catalog.pg_collation coll ON coll.oid=a.attcollation WHERE n.nspname='nodecontrol' AND a.attnum>0 AND NOT a.attisdropped ORDER BY c.relname,a.attnum`,
		`SELECT 'constraint|'||rel.relname||'|'||con.conname||'|'||pg_catalog.pg_get_constraintdef(con.oid,true)||'|'||con.condeferrable||'|'||con.condeferred FROM pg_catalog.pg_constraint con JOIN pg_catalog.pg_class rel ON rel.oid=con.conrelid JOIN pg_catalog.pg_namespace n ON n.oid=rel.relnamespace WHERE n.nspname='nodecontrol' ORDER BY rel.relname,con.conname`,
		`SELECT 'index|'||rel.relname||'|'||idx.relname||'|'||pg_catalog.pg_get_indexdef(idx.oid) FROM pg_catalog.pg_index i JOIN pg_catalog.pg_class rel ON rel.oid=i.indrelid JOIN pg_catalog.pg_class idx ON idx.oid=i.indexrelid JOIN pg_catalog.pg_namespace n ON n.oid=rel.relnamespace WHERE n.nspname='nodecontrol' ORDER BY rel.relname,idx.relname`,
		`SELECT 'trigger|'||rel.relname||'|'||trg.tgname||'|'||pg_catalog.pg_get_triggerdef(trg.oid,true) FROM pg_catalog.pg_trigger trg JOIN pg_catalog.pg_class rel ON rel.oid=trg.tgrelid JOIN pg_catalog.pg_namespace n ON n.oid=rel.relnamespace WHERE n.nspname='nodecontrol' AND NOT trg.tgisinternal ORDER BY rel.relname,trg.tgname`,
		`SELECT 'function|'||p.proname||'|'||pg_catalog.pg_get_functiondef(p.oid) FROM pg_catalog.pg_proc p JOIN pg_catalog.pg_namespace n ON n.oid=p.pronamespace WHERE n.nspname='nodecontrol' ORDER BY p.proname`,
	}
	hash := sha256.New()
	for _, query := range queries {
		rows, err := pool.Query(ctx, query)
		if err != nil {
			t.Fatal("query catalog digest:", err)
		}
		for rows.Next() {
			var line string
			if err := rows.Scan(&line); err != nil {
				rows.Close()
				t.Fatal("scan catalog digest:", err)
			}
			hash.Write([]byte(line))
			hash.Write([]byte{0})
		}
		if err := rows.Err(); err != nil {
			t.Fatal("iterate catalog digest:", err)
		}
		rows.Close()
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func assertAuthorityFenceRejectsForbiddenMutations(ctx context.Context, t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	immutableAssignments := []string{
		"operation_id='00000000-0000-0000-0000-000000000001'::uuid",
		"effect_kind='grant_claim'",
		"scope_kind='global_node_trust'",
		"scope_digest=decode(repeat('cc',32),'hex')",
		"provider_reservation_digest=decode(repeat('dd',32),'hex')",
		"authority_epoch=2",
		"authority_sequence=2",
		"reserved_at=reserved_at + interval '1 second'",
	}
	for index, assignment := range immutableAssignments {
		operationID := uuid.New()
		insertReservedFence(ctx, t, pool, operationID, int64(index+1))
		if _, err := pool.Exec(ctx, "UPDATE nodecontrol.control_plane_authority_fences SET "+assignment+" WHERE operation_id=$1", operationID); err == nil {
			t.Fatalf("authority fence accepted forbidden mutation %s", assignment)
		}
	}

	boundID := uuid.New()
	insertReservedFence(ctx, t, pool, boundID, 100)
	if _, err := pool.Exec(ctx, `UPDATE nodecontrol.control_plane_authority_fences SET effect_digest=decode(repeat('11',32),'hex'),db_system_id=1,db_timeline=1,required_lsn='0/1',effect_bound_at=reserved_at WHERE operation_id=$1`, boundID); err != nil {
		t.Fatal("legal first database binding failed:", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE nodecontrol.control_plane_authority_fences SET effect_digest=decode(repeat('22',32),'hex') WHERE operation_id=$1`, boundID); err == nil {
		t.Fatal("authority fence accepted a second different effect binding")
	}

	terminalID := uuid.New()
	insertReservedFence(ctx, t, pool, terminalID, 101)
	if _, err := pool.Exec(ctx, `UPDATE nodecontrol.control_plane_authority_fences SET provider_status='aborted',provider_receipt_digest=decode(repeat('33',32),'hex'),abort_reason='validation_failed',visibility_state='aborted',terminal_at=reserved_at WHERE operation_id=$1`, terminalID); err != nil {
		t.Fatal("legal reserved-to-aborted transition failed:", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE nodecontrol.control_plane_authority_fences SET abort_reason='superseded' WHERE operation_id=$1`, terminalID); err == nil {
		t.Fatal("authority fence accepted a terminal-row update")
	}

	if _, err := pool.Exec(ctx, `INSERT INTO nodecontrol.control_plane_authority_fences(operation_id,effect_kind,scope_kind,authority_epoch,authority_sequence,scope_digest,provider_reservation_digest,provider_status,visibility_state,reserved_at) VALUES($1,'root_publish','node',1,1000,decode(repeat('44',32),'hex'),decode(repeat('55',32),'hex'),'reserved','fence_pending',transaction_timestamp())`, uuid.New()); err == nil {
		t.Fatal("authority fence accepted an illegal effect/scope pair")
	}
}

func insertReservedFence(ctx context.Context, t *testing.T, pool *pgxpool.Pool, operationID uuid.UUID, sequence int64) {
	t.Helper()
	_, err := pool.Exec(ctx, `INSERT INTO nodecontrol.control_plane_authority_fences(operation_id,effect_kind,scope_kind,authority_epoch,authority_sequence,scope_digest,provider_reservation_digest,provider_status,visibility_state,reserved_at) VALUES($1,'grant_create','node',1,$2,decode(repeat('aa',32),'hex'),decode(repeat('bb',32),'hex'),'reserved','fence_pending',transaction_timestamp())`, operationID, sequence)
	if err != nil {
		t.Fatal("insert reserved authority fence:", err)
	}
}

func insertCommittedFence(ctx context.Context, t *testing.T, pool *pgxpool.Pool, operationID uuid.UUID, sequence int64, effectKind, scopeKind string, now time.Time) {
	t.Helper()
	_, err := pool.Exec(ctx, `
INSERT INTO nodecontrol.control_plane_authority_fences(
  operation_id,effect_kind,scope_kind,authority_epoch,authority_sequence,scope_digest,
  provider_reservation_digest,effect_digest,provider_status,provider_receipt_digest,
  db_system_id,db_timeline,required_lsn,visibility_state,reserved_at,effect_bound_at,terminal_at)
VALUES($1,$2,$3,1,$4,$5,$6,$7,'committed',$8,1,1,'0/1','active',$9,$9,$9)`,
		operationID, effectKind, scopeKind, sequence, bytesOf(0x71, 32), bytesOf(0x72, 32), bytesOf(0x73, 32), bytesOf(0x74, 32), now)
	if err != nil {
		t.Fatal("insert committed authority fence:", err)
	}
}

func bytesOf(value byte, count int) []byte {
	result := make([]byte, count)
	for index := range result {
		result[index] = value
	}
	return result
}

func assertSlotCapAndProfileImmutability(ctx context.Context, t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	nodeID := uuid.New()
	profileID := uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)
	if _, err := pool.Exec(ctx, `INSERT INTO nodecontrol.node_pops(pop_code,iso_country,region,operator_state,created_at,updated_at) VALUES('cap-test','US','test-region','enabled',$1,$1)`, now); err != nil {
		t.Fatal("insert slot-cap POP:", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO nodecontrol.node_inventory(node_id,pop_code,operator_state,security_state,identity_state,created_at,updated_at) VALUES($1,'cap-test','enabled','normal','never_enrolled',$2,$2)`, nodeID, now); err != nil {
		t.Fatal("insert slot-cap node:", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO nodecontrol.node_capacity_profiles(profile_id,version,adapter,egress_limit_bps,connection_limit,handshake_limit_per_second,cpu_quota_millicores,cpu_limit_basis_points,memory_limit_bytes,task_limit,file_descriptor_limit,queue_limit,packet_loss_limit_basis_points,required_metrics,created_at) VALUES($1,1,'fixture',1000000,1,1,100,1,67108864,32,64,1,1,ARRAY['cpu_usage_basis_points']::text[],$2)`, profileID, now); err != nil {
		t.Fatal("insert capacity profile:", err)
	}
	for slot := 1; slot <= 8; slot++ {
		if _, err := pool.Exec(ctx, `INSERT INTO nodecontrol.node_process_slots(node_id,slot_id,adapter,capacity_profile_id,capacity_profile_version,required,operator_state,inventory_version,created_at,updated_at) VALUES($1,$2,'fixture',$3,1,true,'enabled',1,$4,$4)`, nodeID, slot, profileID, now); err != nil {
			t.Fatalf("insert slot %d: %v", slot, err)
		}
	}
	if _, err := pool.Exec(ctx, `INSERT INTO nodecontrol.node_process_slots(node_id,slot_id,adapter,capacity_profile_id,capacity_profile_version,required,operator_state,inventory_version,created_at,updated_at) VALUES($1,9,'fixture',$2,1,true,'enabled',1,$3,$3)`, nodeID, profileID, now); err == nil {
		t.Fatal("node process-slot trigger accepted a ninth slot")
	}
	if _, err := pool.Exec(ctx, `UPDATE nodecontrol.node_capacity_profiles SET queue_limit=2 WHERE profile_id=$1 AND version=1`, profileID); err == nil {
		t.Fatal("referenced capacity profile accepted an update")
	}
	if _, err := pool.Exec(ctx, `DELETE FROM nodecontrol.node_capacity_profiles WHERE profile_id=$1 AND version=1`, profileID); err == nil {
		t.Fatal("referenced capacity profile accepted a delete")
	}
}

func equalOptionalString(actual sql.NullString, expected *string) bool {
	if expected == nil {
		return !actual.Valid
	}
	return actual.Valid && actual.String == *expected
}

func equalOptionalNormalizedSQL(actual sql.NullString, expected *string) bool {
	if expected == nil {
		return !actual.Valid
	}
	return actual.Valid && normalizeCatalogSQL(actual.String) == normalizeCatalogSQL(*expected)
}

func normalizeCatalogSQL(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

func normalizedStrings(values []string) []string {
	result := make([]string, len(values))
	for index, value := range values {
		result[index] = normalizeCatalogSQL(value)
	}
	return result
}

func catalogConstraintKind(code string) string {
	switch code {
	case "c":
		return "check"
	case "f":
		return "foreign_key"
	case "u":
		return "unique"
	case "n":
		return "not_null"
	default:
		return code
	}
}

func decodeFKAction(code string) string {
	switch code {
	case "a":
		return "NO ACTION"
	case "r":
		return "RESTRICT"
	case "c":
		return "CASCADE"
	case "n":
		return "SET NULL"
	case "d":
		return "SET DEFAULT"
	default:
		return code
	}
}

func triggerTiming(typeBits int) string {
	if typeBits&64 != 0 {
		return "INSTEAD OF"
	}
	if typeBits&2 != 0 {
		return "BEFORE"
	}
	return "AFTER"
}

func triggerEvents(typeBits int) []string {
	events := make([]string, 0, 4)
	for _, event := range []struct {
		bit  int
		name string
	}{{4, "INSERT"}, {16, "UPDATE"}, {8, "DELETE"}, {32, "TRUNCATE"}} {
		if typeBits&event.bit != 0 {
			events = append(events, event.name)
		}
	}
	return events
}

func sortedStrings(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	return result
}

func parseInt(value string) int {
	parsed, _ := strconv.Atoi(value)
	return parsed
}

func formatCatalogObject(parts ...any) string {
	formatted := make([]string, len(parts))
	for index, part := range parts {
		formatted[index] = fmt.Sprint(part)
	}
	return strings.Join(formatted, "|")
}
