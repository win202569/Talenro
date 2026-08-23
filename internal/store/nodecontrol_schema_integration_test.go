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
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type nodeControlFunctionCatalogSpec struct {
	language         string
	volatility       string
	securityDefiner  bool
	definitionSHA256 string
}

type nodeControlCatalogFunction struct {
	name, schema, language, volatility string
	securityDefiner                    bool
	definition                         string
}

var expectedNodeControlFunctionCatalog = map[string]nodeControlFunctionCatalogSpec{
	"bytea_array_is_sorted_unique_32":       {language: "sql", volatility: "immutable", definitionSHA256: "3d4197ad58db883fbeb481ae7b852b7a06c6bb3fa9b503af032489d777872141"},
	"enforce_authority_fence_update":        {language: "plpgsql", volatility: "volatile", definitionSHA256: "18986f3677f2572ab276cf45baa7b90143dbcb7928023e79fb65d6576099b062"},
	"enforce_capacity_profile_immutability": {language: "plpgsql", volatility: "volatile", definitionSHA256: "36f60643b0ce6fafa626b51cd0fae56457b609f97c1d29944c8975568e5b543e"},
	"enforce_certificate_issuance_workflow": {language: "plpgsql", volatility: "volatile", definitionSHA256: "fb08394e3a8a1ca54e78f9db3ba8844ec4f11a514b3e2ce339958e5fcb35b651"},
	"enforce_certificate_workflow":          {language: "plpgsql", volatility: "volatile", definitionSHA256: "f0f49c05e4d929826170a5e99b8037ad67b6a96389f54e4966daa4ce006678c4"},
	"enforce_enrollment_grant_workflow":     {language: "plpgsql", volatility: "volatile", definitionSHA256: "6ea9f598f6446c511e08fb9f23835212b4552bdbc0e498e8064b1f4cef1c3e4e"},
	"enforce_inventory_pointers":            {language: "plpgsql", volatility: "volatile", definitionSHA256: "13c741d7128b314847d3b1ae32d31b0e3adb650d5cdd3069549b9b476f8af6de"},
	"enforce_observed_state":                {language: "plpgsql", volatility: "volatile", definitionSHA256: "88283c5dc528bf6e3c31442f12c7395eb186e992870797e242c0e1b54a40060f"},
	"enforce_process_slot_cap":              {language: "plpgsql", volatility: "volatile", definitionSHA256: "dcdd2407b4cf55f7dfb130a99d03db70b08bc146f52f48422f8945924aa000bb"},
	"enforce_recovery_session_workflow":     {language: "plpgsql", volatility: "volatile", definitionSHA256: "6801dcb8c07d6068e80db0de9ab545b3715bce964fa283b27a5cefb533d99481"},
	"enforce_restore_approval_workflow":     {language: "plpgsql", volatility: "volatile", definitionSHA256: "67b092d4385f2c3d52b75f2de317040cb1cc4160fe0708ab128dc85ef4b2feaf"},
	"enforce_root_publish_workflow":         {language: "plpgsql", volatility: "volatile", definitionSHA256: "a00da4a797bf9136e3a79d2a445b35e839ca58421218ef1400f6c5973d927236"},
	"enforce_root_share_binding":            {language: "plpgsql", volatility: "volatile", definitionSHA256: "71fac98750a6db71d9b720f36eec4201ad68c676d34dd0a217b19bd571d91424"},
	"enforce_security_fault_receipt":        {language: "plpgsql", volatility: "volatile", definitionSHA256: "16aff4851ce74b4c584ca54f675ceeec1367d35413a44ab38001938626b73fd7"},
	"enforce_security_incident_workflow":    {language: "plpgsql", volatility: "volatile", definitionSHA256: "af9d4045ac07d521bad979d2d4153046c006c22e5f67f710bb86533b3ceaec5a"},
	"enforce_signing_intent_workflow":       {language: "plpgsql", volatility: "volatile", definitionSHA256: "9f06029ea7ae28b653e530a9680805c40d6b8c787fff966f8e18c79c8ebe0921"},
	"enforce_trust_bundle_high_water":       {language: "plpgsql", volatility: "volatile", definitionSHA256: "0e65dc0a808e05608715962115d8424e8cf1b2fe6aaa23cdc85ce6944d7192bf"},
	"reject_row_mutation":                   {language: "plpgsql", volatility: "volatile", definitionSHA256: "400c44dcb2d5693e961289c31c69ec1bbbf091a1b295abb55e329b1eea12ae42"},
	"text_array_is_sorted_unique":           {language: "sql", volatility: "immutable", definitionSHA256: "d7c18d427a459231fac6ca49bc39cc8dd1d1101f1d09819fcc2a857ca78ff9de"},
}

func validateNodeControlFunctionCatalog(got []nodeControlCatalogFunction, want map[string]nodeControlFunctionCatalogSpec) error {
	if len(got) != len(want) {
		return fmt.Errorf("nodecontrol function count = %d, want %d", len(got), len(want))
	}
	seen := make(map[string]struct{}, len(got))
	for _, function := range got {
		if _, duplicate := seen[function.name]; duplicate {
			return fmt.Errorf("duplicate nodecontrol function %s", function.name)
		}
		seen[function.name] = struct{}{}
		expected, exists := want[function.name]
		if !exists {
			return fmt.Errorf("unexpected nodecontrol function %s", function.name)
		}
		digest := sha256.Sum256([]byte(function.definition))
		definitionSHA256 := hex.EncodeToString(digest[:])
		if function.schema != "nodecontrol" || function.language != expected.language || function.volatility != expected.volatility || function.securityDefiner != expected.securityDefiner || definitionSHA256 != expected.definitionSHA256 {
			return fmt.Errorf("nodecontrol function %s metadata/body mismatch: schema=%s language=%s volatility=%s security_definer=%t sha256=%s", function.name, function.schema, function.language, function.volatility, function.securityDefiner, definitionSHA256)
		}
	}
	return nil
}

func TestNodeControlFunctionCatalogRejectsNoopBody(t *testing.T) {
	want := map[string]nodeControlFunctionCatalogSpec{
		"guard": {
			language:         "plpgsql",
			volatility:       "volatile",
			securityDefiner:  false,
			definitionSHA256: "c105b5b9f479126296850d4a35f8209b11eecfb0d870772864e65425efc5b7e7",
		},
	}
	got := []nodeControlCatalogFunction{{
		name:            "guard",
		schema:          "nodecontrol",
		language:        "plpgsql",
		volatility:      "volatile",
		securityDefiner: false,
		definition:      "CREATE FUNCTION guard() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RETURN NEW; END $$;",
	}}
	if err := validateNodeControlFunctionCatalog(got, want); err == nil {
		t.Fatal("independent function allowlist accepted a weakened no-op trigger body")
	}
}

const pendingRootPublishSQL = `
INSERT INTO nodecontrol.node_root_metadata_publish_intents(
  publish_id,authority_operation_id,authority_epoch,authority_sequence,publish_kind,reason,
  base_root_version,base_metadata_version,reserved_version,canonical_payload,payload_digest,
  key_set_digest,current_key_ids,new_key_ids,current_threshold,new_threshold,activation_deadline,
  status,created_at,updated_at)
VALUES($1,$2,1,$3,$4,$5,$6,$7,$8,decode('01','hex'),$9,decode(repeat('aa',32),'hex'),
       $10,$11,$12,$13,$14,'pending',$15,$15)`

const pendingDesiredSigningSQL = `
INSERT INTO nodecontrol.node_state_signing_intents(
  signing_id,authority_operation_id,authority_epoch,authority_sequence,node_id,signing_kind,
  idempotency_key_digest,base_generation,reserved_generation,canonical_payload,payload_digest,
  root_version,metadata_version,expected_key_id,expected_public_key_digest,captured_inventory_version,
  captured_identity_epoch,captured_security_version,activation_deadline,status,created_at,updated_at)
VALUES($1,$2,1,$3,$4,'desired',decode(repeat('10',32),'hex'),0,1,decode('01','hex'),
       decode(repeat('20',32),'hex'),1,1,$5,$6,1,0,1,$8,'pending',$7,$7)`

const activateSigningSQL = `
UPDATE nodecontrol.node_state_signing_intents
SET signature=$2,signature_verified_at=$3,status='active',terminal_at=$3,updated_at=$3
WHERE signing_id=$1`

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

func TestNodeControlMigrationRejectsIncompleteEnrollmentGrantTerminalGroups(t *testing.T) {
	upSQL, _ := loadNodeControlMigrationSections(t)
	pool := openOwnedNodeControlDatabase(t)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	if _, err := pool.Exec(ctx, upSQL); err != nil {
		t.Fatal("nodecontrol Up failed:", err)
	}

	now := time.Now().UTC().Truncate(time.Microsecond)
	nodeID := insertNodeControlFixtureNode(ctx, t, pool, "grant-terminal-a", now)
	createOperationID := uuid.New()
	insertCommittedFence(ctx, t, pool, createOperationID, 1, "grant_create", "node", now)
	directTerminalNodeID := insertNodeControlFixtureNode(ctx, t, pool, "grant-terminal-insert", now)
	if _, err := pool.Exec(ctx, `
INSERT INTO nodecontrol.node_enrollment_grants(
  grant_id,authority_operation_id,authority_epoch,authority_sequence,node_id,identity_epoch,
  token_digest,csr_digest,idempotency_digest,created_at,expires_at,
  expired_at,terminal_reason,terminal_at,retention_until)
VALUES($1,$2,1,1,$3,1,$4,$5,$6,$7,$8,$9,'expired',$9,$10)`, uuid.New(), createOperationID,
		directTerminalNodeID, bytesOf(0x01, 32), bytesOf(0x02, 32), bytesOf(0x03, 32), now,
		now.Add(10*time.Minute), now.Add(time.Minute), now.Add(31*24*time.Hour)); err == nil {
		t.Fatal("enrollment grant accepted a directly inserted terminal lifecycle row")
	}
	insertLiveGrant := func(grantID, grantNodeID uuid.UUID, tokenByte byte) {
		t.Helper()
		if _, err := pool.Exec(ctx, `
INSERT INTO nodecontrol.node_enrollment_grants(
  grant_id,authority_operation_id,authority_epoch,authority_sequence,node_id,identity_epoch,
  token_digest,csr_digest,idempotency_digest,created_at,expires_at)
VALUES($1,$2,1,1,$3,1,$4,$5,$6,$7,$8)`, grantID, createOperationID, grantNodeID,
			bytesOf(tokenByte, 32), bytesOf(tokenByte+1, 32), bytesOf(tokenByte+2, 32), now, now.Add(10*time.Minute)); err != nil {
			t.Fatal("insert live enrollment grant:", err)
		}
	}

	expiredGrantID := uuid.New()
	insertLiveGrant(expiredGrantID, nodeID, 0x11)
	if _, err := pool.Exec(ctx, `UPDATE nodecontrol.node_enrollment_grants SET expired_at=$2 WHERE grant_id=$1`, expiredGrantID, now.Add(time.Minute)); err == nil {
		t.Fatal("enrollment grant accepted expired_at without terminal cause, timestamp, and retention")
	}

	invalidatedGrantID := uuid.New()
	invalidatedNodeID := insertNodeControlFixtureNode(ctx, t, pool, "grant-terminal-b", now)
	insertLiveGrant(invalidatedGrantID, invalidatedNodeID, 0x21)
	if _, err := pool.Exec(ctx, `UPDATE nodecontrol.node_enrollment_grants SET invalidated_at=$2 WHERE grant_id=$1`, invalidatedGrantID, now.Add(time.Minute)); err == nil {
		t.Fatal("enrollment grant accepted invalidated_at without terminal cause, timestamp, and retention")
	}

	claimOperationID := uuid.New()
	insertCommittedFence(ctx, t, pool, claimOperationID, 2, "grant_claim", "node", now)
	issuanceOperationID := uuid.New()
	insertCommittedFence(ctx, t, pool, issuanceOperationID, 3, "certificate_activate", "node", now)
	consumedNodeID := insertNodeControlFixtureNode(ctx, t, pool, "grant-terminal-c", now)
	issuanceID := uuid.New()
	if _, err := pool.Exec(ctx, `
INSERT INTO nodecontrol.node_certificate_issuances(
  issuance_id,authority_operation_id,authority_epoch,authority_sequence,node_id,attempt_id,
  issuance_kind,identity_epoch,lineage_id,issuer_id,csr_sha256,public_key_sha256,
  template_sha256,request_digest,status,created_at,updated_at)
VALUES($1,$2,1,3,$3,$4,'initial',1,$5,'fixture-ca',$6,$7,$8,$9,'pending',$10,$10)`,
		issuanceID, issuanceOperationID, consumedNodeID, uuid.New(), uuid.New(), bytesOf(0x31, 32),
		bytesOf(0x32, 32), bytesOf(0x33, 32), bytesOf(0x34, 32), now); err != nil {
		t.Fatal("insert grant-result issuance:", err)
	}
	consumedGrantID := uuid.New()
	insertLiveGrant(consumedGrantID, consumedNodeID, 0x41)
	if _, err := pool.Exec(ctx, `
UPDATE nodecontrol.node_enrollment_grants
SET consumed_at=$2,consumption_attempt_id=$3,consumption_request_digest=$4,
    claim_authority_operation_id=$5,claim_authority_epoch=1,claim_authority_sequence=2,
    result_issuance_id=$6
WHERE grant_id=$1`, consumedGrantID, now.Add(time.Minute), uuid.New(), bytesOf(0x51, 32), claimOperationID, issuanceID); err == nil {
		t.Fatal("enrollment grant accepted complete consumption without terminal cause, timestamp, and retention")
	}

	mixedNodeID := insertNodeControlFixtureNode(ctx, t, pool, "grant-terminal-d", now)
	mixedGrantID := uuid.New()
	insertLiveGrant(mixedGrantID, mixedNodeID, 0x61)
	if _, err := pool.Exec(ctx, `
UPDATE nodecontrol.node_enrollment_grants
SET expired_at=$2,invalidated_at=$2,terminal_reason='expired',terminal_at=$2,retention_until=$3
WHERE grant_id=$1`, mixedGrantID, now.Add(time.Minute), now.Add(31*24*time.Hour)); err == nil {
		t.Fatal("enrollment grant accepted two mutually exclusive terminal kinds")
	}

	terminalNodeID := insertNodeControlFixtureNode(ctx, t, pool, "grant-terminal-e", now)
	terminalGrantID := uuid.New()
	insertLiveGrant(terminalGrantID, terminalNodeID, 0x71)
	if _, err := pool.Exec(ctx, `
UPDATE nodecontrol.node_enrollment_grants
SET expired_at=$2,terminal_reason='expired',terminal_at=$2,retention_until=$3
WHERE grant_id=$1`, terminalGrantID, now.Add(time.Minute), now.Add(31*24*time.Hour)); err != nil {
		t.Fatal("expire a complete grant terminal group:", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE nodecontrol.node_enrollment_grants SET retention_until=$2 WHERE grant_id=$1`, terminalGrantID, now.Add(32*24*time.Hour)); err == nil {
		t.Fatal("terminal enrollment grant accepted a lifecycle mutation")
	}

	liveNodeID := insertNodeControlFixtureNode(ctx, t, pool, "grant-terminal-f", now)
	insertLiveGrant(uuid.New(), liveNodeID, 0x81)
	if _, err := pool.Exec(ctx, `
INSERT INTO nodecontrol.node_enrollment_grants(
  grant_id,authority_operation_id,authority_epoch,authority_sequence,node_id,identity_epoch,
  token_digest,csr_digest,idempotency_digest,created_at,expires_at)
VALUES($1,$2,1,1,$3,1,$4,$5,$6,$7,$8)`, uuid.New(), createOperationID, liveNodeID,
		bytesOf(0x91, 32), bytesOf(0x92, 32), bytesOf(0x93, 32), now, now.Add(10*time.Minute)); err == nil {
		t.Fatal("enrollment grants accepted two live rows for one node identity epoch")
	}
}

func TestNodeControlMigrationRejectsAuthorityFenceTerminalInsertAndDelete(t *testing.T) {
	upSQL, _ := loadNodeControlMigrationSections(t)
	pool := openOwnedNodeControlDatabase(t)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	if _, err := pool.Exec(ctx, upSQL); err != nil {
		t.Fatal("nodecontrol Up failed:", err)
	}
	now := time.Now().UTC().Truncate(time.Microsecond)
	if _, err := pool.Exec(ctx, `
INSERT INTO nodecontrol.control_plane_authority_fences(
  operation_id,effect_kind,scope_kind,authority_epoch,authority_sequence,scope_digest,
  provider_reservation_digest,effect_digest,provider_status,provider_receipt_digest,
  db_system_id,db_timeline,required_lsn,visibility_state,reserved_at,effect_bound_at,terminal_at)
VALUES($1,'desired_activate','node',1,1,$2,$3,$4,'committed',$5,1,1,'0/1','active',$6,$6,$6)`,
		uuid.New(), bytesOf(0x61, 32), bytesOf(0x62, 32), bytesOf(0x63, 32), bytesOf(0x64, 32), now); err == nil {
		t.Fatal("authority fence accepted a directly inserted committed row")
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO nodecontrol.control_plane_authority_fences(
  operation_id,effect_kind,scope_kind,authority_epoch,authority_sequence,scope_digest,
  provider_reservation_digest,provider_status,abort_reason,visibility_state,reserved_at,terminal_at)
VALUES($1,'desired_activate','node',1,3,$2,$3,'aborted','validation_failed','aborted',$4,$4)`,
		uuid.New(), bytesOf(0x65, 32), bytesOf(0x66, 32), now); err == nil {
		t.Fatal("authority fence accepted a directly inserted aborted row")
	}
	operationID := uuid.New()
	insertReservedFence(ctx, t, pool, operationID, 2)
	if _, err := pool.Exec(ctx, `DELETE FROM nodecontrol.control_plane_authority_fences WHERE operation_id=$1`, operationID); err == nil {
		t.Fatal("authority fence accepted DELETE")
	}
}

func TestNodeControlMigrationRejectsFabricatedCertificate(t *testing.T) {
	upSQL, _ := loadNodeControlMigrationSections(t)
	pool := openOwnedNodeControlDatabase(t)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	if _, err := pool.Exec(ctx, upSQL); err != nil {
		t.Fatal("nodecontrol Up failed:", err)
	}
	now := time.Now().UTC().Truncate(time.Microsecond)
	nodeID := insertNodeControlFixtureNode(ctx, t, pool, "certificate-binding", now)
	operationID := uuid.New()
	insertCommittedFence(ctx, t, pool, operationID, 1, "certificate_activate", "node", now)
	issuanceID := uuid.New()
	lineageID := uuid.New()
	if _, err := pool.Exec(ctx, `
INSERT INTO nodecontrol.node_certificate_issuances(
  issuance_id,authority_operation_id,authority_epoch,authority_sequence,node_id,attempt_id,
  issuance_kind,identity_epoch,lineage_id,issuer_id,csr_sha256,public_key_sha256,
  template_sha256,request_digest,status,created_at,updated_at)
VALUES($1,$2,1,1,$3,$4,'initial',1,$5,'fixture-ca',$6,$7,$8,$9,'pending',$10,$10)`,
		issuanceID, operationID, nodeID, uuid.New(), lineageID, bytesOf(0x11, 32), bytesOf(0x12, 32),
		bytesOf(0x13, 32), bytesOf(0x14, 32), now); err != nil {
		t.Fatal("insert pending issuance:", err)
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO nodecontrol.node_certificates(
  certificate_id,issuance_id,authority_operation_id,authority_epoch,authority_sequence,node_id,
  identity_epoch,lineage_id,issuer_id,serial_bytes,leaf_der,leaf_der_sha256,public_key_sha256,
  chain_der_sha256,valid_from,valid_until,status,created_at,updated_at,retention_until)
VALUES($1,$2,$3,1,1,$4,1,$5,'fixture-ca',$6,$7,$8,$9,$10,$11,$12,'active',$11,$11,$13)`,
		uuid.New(), issuanceID, operationID, nodeID, lineageID, []byte{0x01}, []byte{0x7f},
		bytesOf(0x21, 32), bytesOf(0x22, 32), bytesOf(0x23, 32), now, now.Add(24*time.Hour), now.Add(60*24*time.Hour)); err == nil {
		t.Fatal("certificate INSERT accepted fabricated fields against a pending unrelated issuance result")
	}
}

func TestNodeControlMigrationClosesIncidentAndReceiptLifecycleBypasses(t *testing.T) {
	upSQL, _ := loadNodeControlMigrationSections(t)
	pool := openOwnedNodeControlDatabase(t)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	if _, err := pool.Exec(ctx, upSQL); err != nil {
		t.Fatal("nodecontrol Up failed:", err)
	}
	now := time.Now().UTC().Truncate(time.Microsecond)
	nodeA := insertNodeControlFixtureNode(ctx, t, pool, "incident-a", now)
	nodeB := insertNodeControlFixtureNode(ctx, t, pool, "incident-b", now)
	openOperationID := uuid.New()
	resolveOperationID := uuid.New()
	insertCommittedFence(ctx, t, pool, openOperationID, 1, "security_incident_open", "node", now)
	insertCommittedFence(ctx, t, pool, resolveOperationID, 2, "security_incident_resolve", "node", now)

	if _, err := pool.Exec(ctx, `
INSERT INTO nodecontrol.node_security_incidents(
  incident_id,authority_operation_id,authority_epoch,authority_sequence,node_id,identity_epoch,
  fault_subtype,subtype_slot,status,first_evidence_digest,last_evidence_digest,occurrence_count,
  trust_context_digest,first_occurred_at,last_occurred_at,resolution_authority_operation_id,
  resolution_authority_epoch,resolution_authority_sequence,remediation_digest,resolution_at,retention_until)
VALUES($1,$2,1,1,$3,1,'identity_compromise',1,'resolved',$4,$4,1,$5,$6,$6,$7,1,2,$8,$9,$10)`,
		uuid.New(), openOperationID, nodeA, bytesOf(0x31, 32), bytesOf(0x32, 32), now,
		resolveOperationID, bytesOf(0x33, 32), now.Add(time.Minute), now.Add(181*24*time.Hour)); err == nil {
		t.Fatal("non-overflow incident accepted a direct resolved INSERT")
	}

	incidentID := uuid.New()
	insertOpenIncident(ctx, t, pool, incidentID, openOperationID, 1, nodeA, "online_signer_equivocation", 2, now)
	if _, err := pool.Exec(ctx, `
INSERT INTO nodecontrol.node_security_fault_receipts(
  receipt_id,authority_operation_id,authority_epoch,authority_sequence,node_id,identity_epoch,
  local_fault_id,request_digest,fault_subtype,evidence_digest,agent_boot_id,incident_id,
  local_binding_slot,result,delivery_status,binding_status,cleared_at,clear_attestation_digest,created_at,updated_at)
VALUES($1,$2,1,1,$3,1,$4,$5,'online_signer_equivocation',$6,$7,$8,1,'accepted','pending','cleared',$9,$10,$11,$11)`,
		uuid.New(), openOperationID, nodeA, uuid.New(), bytesOf(0x41, 32), bytesOf(0x42, 32), uuid.New(),
		incidentID, now.Add(time.Minute), bytesOf(0x43, 32), now); err == nil {
		t.Fatal("security fault receipt accepted a directly cleared INSERT")
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO nodecontrol.node_security_fault_receipts(
  receipt_id,authority_operation_id,authority_epoch,authority_sequence,node_id,identity_epoch,
  local_fault_id,request_digest,fault_subtype,evidence_digest,agent_boot_id,incident_id,
  local_binding_slot,result,delivery_status,binding_status,created_at,updated_at)
VALUES($1,$2,1,1,$3,1,$4,$5,'online_signer_equivocation',$6,$7,$8,1,'accepted','pending','active',$9,$9)`,
		uuid.New(), openOperationID, nodeB, uuid.New(), bytesOf(0x51, 32), bytesOf(0x52, 32), uuid.New(), incidentID, now); err == nil {
		t.Fatal("security fault receipt cross-bound an incident from another node")
	}

	boundIncidentID := uuid.New()
	insertOpenIncident(ctx, t, pool, boundIncidentID, openOperationID, 1, nodeA, "metadata_rollback", 3, now)
	boundReceiptID := uuid.New()
	insertActiveFaultReceipt(ctx, t, pool, boundReceiptID, openOperationID, 1, nodeA, boundIncidentID, "metadata_rollback", 3, now)
	if _, err := pool.Exec(ctx, `
UPDATE nodecontrol.node_security_incidents
SET status='resolution_pending_agent_ack',resolution_authority_operation_id=$2,
    resolution_authority_epoch=1,resolution_authority_sequence=2,remediation_digest=$3,resolution_at=$4
WHERE incident_id=$1`, boundIncidentID, resolveOperationID, bytesOf(0x61, 32), now.Add(time.Minute)); err != nil {
		t.Fatal("move incident to pending acknowledgement:", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE nodecontrol.node_security_incidents SET status='resolved',retention_until=$2 WHERE incident_id=$1`, boundIncidentID, now.Add(181*24*time.Hour)); err == nil {
		t.Fatal("pending-ack incident resolved while an active receipt remained bound")
	}
	if _, err := pool.Exec(ctx, `UPDATE nodecontrol.node_security_fault_receipts SET delivery_status='deliverable',delivered_at=$2,updated_at=$2 WHERE receipt_id=$1`, boundReceiptID, now.Add(2*time.Minute)); err != nil {
		t.Fatal("mark exact receipt deliverable:", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE nodecontrol.node_security_fault_receipts SET delivered_at=$2,updated_at=$2 WHERE receipt_id=$1`, boundReceiptID, now.Add(3*time.Minute)); err == nil {
		t.Fatal("delivered receipt accepted mutable delivery metadata")
	}
	if _, err := pool.Exec(ctx, `UPDATE nodecontrol.node_security_fault_receipts SET binding_status='cleared',cleared_at=$2,clear_attestation_digest=$3,updated_at=$2 WHERE receipt_id=$1`, boundReceiptID, now.Add(4*time.Minute), bytesOf(0x62, 32)); err != nil {
		t.Fatal("clear exact receipt binding:", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE nodecontrol.node_security_fault_receipts SET clear_attestation_digest=$2,updated_at=$3 WHERE receipt_id=$1`, boundReceiptID, bytesOf(0x63, 32), now.Add(5*time.Minute)); err == nil {
		t.Fatal("cleared receipt accepted mutable clear metadata")
	}
}

func TestNodeControlMigrationEnforcesTrustHighWaterAppendOnly(t *testing.T) {
	t.Run("delete", func(t *testing.T) {
		ctx, pool := openMigratedNodeControlDatabase(t)
		now := time.Now().UTC().Truncate(time.Microsecond)
		operationID := uuid.New()
		insertCommittedFence(ctx, t, pool, operationID, 1, "trust_bundle_publish", "global_node_trust", now)
		insertTrustHighWater(ctx, t, pool, operationID, 1, 2, 2, 0x11, 0x21, now)
		if _, err := pool.Exec(ctx, `DELETE FROM nodecontrol.control_plane_trust_bundle_high_waters WHERE purpose='bootstrap_server' AND listener_kind='bootstrap' AND trust_domain='example.com'`); err == nil {
			t.Fatal("trust bundle high-water accepted DELETE")
		}
	})
	t.Run("bundle version independent of authority", func(t *testing.T) {
		ctx, pool := openMigratedNodeControlDatabase(t)
		now := time.Now().UTC().Truncate(time.Microsecond)
		firstOperationID := uuid.New()
		secondOperationID := uuid.New()
		insertCommittedFence(ctx, t, pool, firstOperationID, 1, "trust_bundle_publish", "global_node_trust", now)
		insertCommittedFence(ctx, t, pool, secondOperationID, 2, "trust_bundle_publish", "global_node_trust", now)
		insertTrustHighWater(ctx, t, pool, firstOperationID, 1, 2, 2, 0x12, 0x22, now)
		if _, err := pool.Exec(ctx, `
UPDATE nodecontrol.control_plane_trust_bundle_high_waters
SET authority_operation_id=$1,authority_sequence=2,bundle_version=1,bundle_digest=$2,
    cumulative_set_digest=$3,cumulative_set_count=3,updated_at=$4
WHERE purpose='bootstrap_server' AND listener_kind='bootstrap' AND trust_domain='example.com'`,
			secondOperationID, bytesOf(0x13, 32), bytesOf(0x23, 32), now.Add(time.Second)); err == nil {
			t.Fatal("trust high-water accepted a lower bundle version under a higher authority sequence")
		}
	})
	t.Run("same count different cumulative digest", func(t *testing.T) {
		ctx, pool := openMigratedNodeControlDatabase(t)
		now := time.Now().UTC().Truncate(time.Microsecond)
		firstOperationID := uuid.New()
		secondOperationID := uuid.New()
		insertCommittedFence(ctx, t, pool, firstOperationID, 1, "trust_bundle_publish", "global_node_trust", now)
		insertCommittedFence(ctx, t, pool, secondOperationID, 2, "trust_bundle_publish", "global_node_trust", now)
		insertTrustHighWater(ctx, t, pool, firstOperationID, 1, 1, 2, 0x14, 0x24, now)
		if _, err := pool.Exec(ctx, `
UPDATE nodecontrol.control_plane_trust_bundle_high_waters
SET authority_operation_id=$1,authority_sequence=2,bundle_version=2,bundle_digest=$2,
    cumulative_set_digest=$3,updated_at=$4
WHERE purpose='bootstrap_server' AND listener_kind='bootstrap' AND trust_domain='example.com'`,
			secondOperationID, bytesOf(0x15, 32), bytesOf(0x25, 32), now.Add(time.Second)); err == nil {
			t.Fatal("trust high-water accepted a same-count cumulative digest fork")
		}
	})
	t.Run("invalid cumulative count and digest transitions", func(t *testing.T) {
		ctx, pool := openMigratedNodeControlDatabase(t)
		now := time.Now().UTC().Truncate(time.Microsecond)
		firstOperationID := uuid.New()
		secondOperationID := uuid.New()
		insertCommittedFence(ctx, t, pool, firstOperationID, 1, "trust_bundle_publish", "global_node_trust", now)
		insertCommittedFence(ctx, t, pool, secondOperationID, 2, "trust_bundle_publish", "global_node_trust", now)
		insertTrustHighWater(ctx, t, pool, firstOperationID, 1, 1, 2, 0x16, 0x26, now)
		if _, err := pool.Exec(ctx, `
UPDATE nodecontrol.control_plane_trust_bundle_high_waters
SET authority_operation_id=$1,authority_sequence=2,bundle_version=2,bundle_digest=$2,
    cumulative_set_digest=$3,cumulative_set_count=1,updated_at=$4
WHERE purpose='bootstrap_server' AND listener_kind='bootstrap' AND trust_domain='example.com'`,
			secondOperationID, bytesOf(0x17, 32), bytesOf(0x27, 32), now.Add(time.Second)); err == nil {
			t.Fatal("trust high-water accepted a lower cumulative-set count")
		}
		if _, err := pool.Exec(ctx, `
UPDATE nodecontrol.control_plane_trust_bundle_high_waters
SET authority_operation_id=$1,authority_sequence=2,bundle_version=2,bundle_digest=$2,
    cumulative_set_digest=$3,cumulative_set_count=3,updated_at=$4
WHERE purpose='bootstrap_server' AND listener_kind='bootstrap' AND trust_domain='example.com'`,
			secondOperationID, bytesOf(0x18, 32), bytesOf(0x26, 32), now.Add(time.Second)); err == nil {
			t.Fatal("trust high-water accepted a larger count with an unchanged cumulative digest")
		}
		if _, err := pool.Exec(ctx, `
UPDATE nodecontrol.control_plane_trust_bundle_high_waters
SET authority_operation_id=$1,authority_sequence=2,bundle_version=2,bundle_digest=$2,
    cumulative_set_digest=$3,cumulative_set_count=3,updated_at=$4
WHERE purpose='bootstrap_server' AND listener_kind='bootstrap' AND trust_domain='example.com'`,
			secondOperationID, bytesOf(0x19, 32), bytesOf(0x29, 32), now.Add(time.Second)); err != nil {
			t.Fatalf("trust high-water rejected an observable strict append: %v", err)
		}
		if _, err := pool.Exec(ctx, `UPDATE nodecontrol.control_plane_trust_bundle_high_waters SET updated_at=updated_at WHERE purpose='bootstrap_server' AND listener_kind='bootstrap' AND trust_domain='example.com'`); err != nil {
			t.Fatalf("trust high-water rejected an exact retry: %v", err)
		}
	})
}

func TestNodeControlMigrationRejectsObservedStateDelete(t *testing.T) {
	ctx, pool := openMigratedNodeControlDatabase(t)
	now := time.Now().UTC().Truncate(time.Microsecond)
	nodeID := insertNodeControlFixtureNode(ctx, t, pool, "observed-delete", now)
	if _, err := pool.Exec(ctx, `
INSERT INTO nodecontrol.node_observed_states(
  node_id,boot_id,sequence,request_digest,observation_digest,canonical_observation,
  sample_ended_at,arrived_at,inventory_version,reducer_version,reducer_input_digest,
  reducer_state_digest,health_state,health_reason,capacity_accepting,agent_accepting,
  final_accepting,created_at,updated_at)
VALUES($1,$2,1,$3,$4,$5,$6,$6,1,1,$7,$8,'healthy','none',true,true,true,$6,$6)`,
		nodeID, uuid.New(), bytesOf(0x31, 32), bytesOf(0x32, 32), []byte{0x01}, now,
		bytesOf(0x33, 32), bytesOf(0x34, 32)); err != nil {
		t.Fatal("insert observed state:", err)
	}
	if _, err := pool.Exec(ctx, `DELETE FROM nodecontrol.node_observed_states WHERE node_id=$1`, nodeID); err == nil {
		t.Fatal("observed-state high-water accepted DELETE")
	}
}

func TestNodeControlMigrationEnforcesRootLimitsAndSignerExclusion(t *testing.T) {
	t.Run("at most five captured root keys", func(t *testing.T) {
		ctx, pool := openMigratedNodeControlDatabase(t)
		now := time.Now().UTC().Truncate(time.Microsecond)
		operationID := uuid.New()
		insertCommittedFence(ctx, t, pool, operationID, 1, "root_publish", "global_node_trust", now)
		keys := [][]byte{bytesOf(1, 32), bytesOf(2, 32), bytesOf(3, 32), bytesOf(4, 32), bytesOf(5, 32), bytesOf(6, 32)}
		if _, err := pool.Exec(ctx, pendingRootPublishSQL, uuid.New(), operationID, int64(1), "root", "normal", int64(0), int64(0), int64(1), bytesOf(0x41, 32), keys, nil, 1, nil, now.Add(5*time.Minute), now); err == nil {
			t.Fatal("root publish accepted six captured current root keys")
		}
	})
	t.Run("root rotation only applies to root publish", func(t *testing.T) {
		ctx, pool := openMigratedNodeControlDatabase(t)
		now := time.Now().UTC().Truncate(time.Microsecond)
		operationID := uuid.New()
		insertCommittedFence(ctx, t, pool, operationID, 1, "metadata_publish", "global_node_trust", now)
		keys := [][]byte{bytesOf(1, 32)}
		newKeys := [][]byte{bytesOf(2, 32)}
		newThreshold := 1
		if _, err := pool.Exec(ctx, pendingRootPublishSQL, uuid.New(), operationID, int64(1), "metadata", "root_rotation", int64(0), int64(0), int64(1), bytesOf(0x42, 32), keys, newKeys, 1, newThreshold, now.Add(5*time.Minute), now); err == nil {
			t.Fatal("metadata publish accepted root_rotation reason")
		}
	})
	t.Run("share first then online signer", func(t *testing.T) {
		ctx, pool := openMigratedNodeControlDatabase(t)
		now := time.Now().UTC().Truncate(time.Microsecond)
		rootOperationID := uuid.New()
		insertCommittedFence(ctx, t, pool, rootOperationID, 1, "root_publish", "global_node_trust", now)
		publishID := uuid.New()
		keyID := bytesOf(0x51, 32)
		physicalKeyID := bytesOf(0x52, 32)
		payloadDigest := bytesOf(0x53, 32)
		if _, err := pool.Exec(ctx, pendingRootPublishSQL, publishID, rootOperationID, int64(1), "root", "normal", int64(0), int64(0), int64(1), payloadDigest, [][]byte{keyID}, nil, 1, nil, now.Add(5*time.Minute), now); err != nil {
			t.Fatal("insert pending root publish:", err)
		}
		if _, err := pool.Exec(ctx, `INSERT INTO nodecontrol.node_root_metadata_signature_shares(publish_id,key_id,physical_key_id,payload_digest,signature_role,signature,verified_at) VALUES($1,$2,$3,$4,'current_root',$5,$6)`, publishID, keyID, physicalKeyID, payloadDigest, bytesOf(0x54, 64), now); err != nil {
			t.Fatal("insert root share before online signer:", err)
		}
		nodeID := insertNodeControlFixtureNode(ctx, t, pool, "share-first", now)
		signingOperationID := uuid.New()
		insertCommittedFence(ctx, t, pool, signingOperationID, 2, "desired_activate", "node", now)
		if _, err := pool.Exec(ctx, pendingDesiredSigningSQL, uuid.New(), signingOperationID, int64(2), nodeID, keyID, physicalKeyID, now, now.Add(5*time.Minute)); err == nil {
			t.Fatal("online signing intent reused an identity already captured by a root share")
		}
	})
}

func TestNodeControlMigrationSigningActivationRechecksFenceDeadlineAndCapture(t *testing.T) {
	t.Run("deadline", func(t *testing.T) {
		ctx, pool := openMigratedNodeControlDatabase(t)
		now := time.Now().UTC().Truncate(time.Microsecond)
		nodeID := insertNodeControlFixtureNode(ctx, t, pool, "sign-deadline", now)
		operationID := uuid.New()
		insertCommittedFence(ctx, t, pool, operationID, 1, "desired_activate", "node", now)
		signingID := uuid.New()
		if _, err := pool.Exec(ctx, pendingDesiredSigningSQL, signingID, operationID, int64(1), nodeID, bytesOf(0x61, 32), bytesOf(0x62, 32), now.Add(-time.Hour), now.Add(-time.Minute)); err != nil {
			t.Fatal("insert expired pending signing intent:", err)
		}
		assertSQLFailureContains(ctx, t, pool, activateSigningSQL, "deadline", signingID, bytesOf(0x63, 64), now)
	})
	t.Run("committed fence", func(t *testing.T) {
		ctx, pool := openMigratedNodeControlDatabase(t)
		now := time.Now().UTC().Truncate(time.Microsecond)
		nodeID := insertNodeControlFixtureNode(ctx, t, pool, "sign-fence", now)
		operationID := uuid.New()
		insertPendingFence(ctx, t, pool, operationID, 1, "desired_activate", "node", now)
		signingID := uuid.New()
		if _, err := pool.Exec(ctx, pendingDesiredSigningSQL, signingID, operationID, int64(1), nodeID, bytesOf(0x64, 32), bytesOf(0x65, 32), now, now.Add(5*time.Minute)); err != nil {
			t.Fatal("insert pending signing intent:", err)
		}
		assertSQLFailureContains(ctx, t, pool, activateSigningSQL, "committed", signingID, bytesOf(0x66, 64), now.Add(time.Second))
	})
	t.Run("captured inventory", func(t *testing.T) {
		ctx, pool := openMigratedNodeControlDatabase(t)
		now := time.Now().UTC().Truncate(time.Microsecond)
		nodeID := insertNodeControlFixtureNode(ctx, t, pool, "sign-capture", now)
		rootPublishID := insertActiveRootOrMetadataPublish(ctx, t, pool, "root", 1, nodeID, now)
		metadataPublishID := insertActiveRootOrMetadataPublish(ctx, t, pool, "metadata", 2, nodeID, now)
		if _, err := pool.Exec(ctx, `UPDATE nodecontrol.node_inventory SET active_root_publish_id=$2,active_root_version=1,active_metadata_publish_id=$3,active_metadata_version=1,updated_at=$4 WHERE node_id=$1`, nodeID, rootPublishID, metadataPublishID, now.Add(time.Second)); err != nil {
			t.Fatal("set captured-inventory trust pointers:", err)
		}
		operationID := uuid.New()
		insertCommittedFence(ctx, t, pool, operationID, 3, "desired_activate", "node", now)
		signingID := uuid.New()
		if _, err := pool.Exec(ctx, pendingDesiredSigningSQL, signingID, operationID, int64(3), nodeID, bytesOf(0x67, 32), bytesOf(0x68, 32), now.Add(time.Second), now.Add(5*time.Minute)); err != nil {
			t.Fatal("insert pending signing intent:", err)
		}
		if _, err := pool.Exec(ctx, `UPDATE nodecontrol.node_inventory SET inventory_version=2,updated_at=$2 WHERE node_id=$1`, nodeID, now.Add(2*time.Second)); err != nil {
			t.Fatal("advance captured inventory value:", err)
		}
		assertSQLFailureContains(ctx, t, pool, activateSigningSQL, "captured", signingID, bytesOf(0x69, 64), now.Add(3*time.Second))
	})
	t.Run("current root and metadata publishes", func(t *testing.T) {
		ctx, pool := openMigratedNodeControlDatabase(t)
		now := time.Now().UTC().Truncate(time.Microsecond)
		nodeID := insertNodeControlFixtureNode(ctx, t, pool, "sign-current-trust", now)
		rootPublishID := insertActiveRootOrMetadataPublish(ctx, t, pool, "root", 1, nodeID, now)
		metadataPublishID := insertActiveRootOrMetadataPublish(ctx, t, pool, "metadata", 2, nodeID, now)
		if _, err := pool.Exec(ctx, `
UPDATE nodecontrol.node_inventory
SET active_root_publish_id=$2,active_root_version=1,
    active_metadata_publish_id=$3,active_metadata_version=1,updated_at=$4
WHERE node_id=$1`, nodeID, rootPublishID, metadataPublishID, now.Add(2*time.Second)); err != nil {
			t.Fatal("set signing trust pointers:", err)
		}
		desiredOperationID := uuid.New()
		insertCommittedFence(ctx, t, pool, desiredOperationID, 3, "desired_activate", "node", now)
		signingID := uuid.New()
		if _, err := pool.Exec(ctx, pendingDesiredSigningSQL, signingID, desiredOperationID, int64(3), nodeID, bytesOf(0x6a, 32), bytesOf(0x6b, 32), now.Add(2*time.Second), now.Add(10*time.Minute)); err != nil {
			t.Fatal("insert signing intent before metadata supersession:", err)
		}

		nextMetadataOperationID := uuid.New()
		insertCommittedFence(ctx, t, pool, nextMetadataOperationID, 4, "metadata_publish", "global_node_trust", now)
		tx, err := pool.Begin(ctx)
		if err != nil {
			t.Fatal("begin metadata supersession:", err)
		}
		nextMetadataPublishID := uuid.New()
		payloadDigest := bytesOf(0x6c, 32)
		metadataKeyID := bytesOf(0x6d, 32)
		if _, err = tx.Exec(ctx, pendingRootPublishSQL, nextMetadataPublishID, nextMetadataOperationID, int64(4), "metadata", "normal", int64(1), int64(1), int64(2), payloadDigest, [][]byte{metadataKeyID}, nil, 1, nil, now.Add(10*time.Minute), now.Add(3*time.Second)); err != nil {
			tx.Rollback(ctx)
			t.Fatal("insert next metadata intent:", err)
		}
		if _, err = tx.Exec(ctx, `INSERT INTO nodecontrol.node_root_metadata_signature_shares(publish_id,key_id,physical_key_id,payload_digest,signature_role,signature,verified_at) VALUES($1,$2,$3,$4,'metadata',$5,$6)`, nextMetadataPublishID, metadataKeyID, bytesOf(0x6e, 32), payloadDigest, bytesOf(0x6f, 64), now.Add(3*time.Second)); err != nil {
			tx.Rollback(ctx)
			t.Fatal("insert next metadata share:", err)
		}
		if _, err = tx.Exec(ctx, `UPDATE nodecontrol.node_root_metadata_publish_intents SET published_envelope=decode('01','hex'),published_envelope_digest=$2,status='active',terminal_at=$3,updated_at=$3 WHERE publish_id=$1`, nextMetadataPublishID, bytesOf(0x70, 32), now.Add(4*time.Second)); err != nil {
			tx.Rollback(ctx)
			t.Fatal("activate next metadata intent:", err)
		}
		if _, err = tx.Exec(ctx, `UPDATE nodecontrol.node_root_metadata_publish_intents SET status='superseded',failure_reason='superseded',terminal_at=$2,updated_at=$2 WHERE publish_id=$1`, metadataPublishID, now.Add(4*time.Second)); err != nil {
			tx.Rollback(ctx)
			t.Fatal("supersede prior metadata intent:", err)
		}
		if err = tx.Commit(ctx); err != nil {
			t.Fatal("commit metadata supersession:", err)
		}
		assertSQLFailureContains(ctx, t, pool, activateSigningSQL, "captured", signingID, bytesOf(0x71, 64), now.Add(5*time.Second))
	})
	t.Run("captured recovery session", func(t *testing.T) {
		ctx, pool := openMigratedNodeControlDatabase(t)
		now := time.Now().UTC().Truncate(time.Microsecond)
		nodeID := insertNodeControlFixtureNode(ctx, t, pool, "sign-recovery-capture", now)
		rootPublishID := insertActiveRootOrMetadataPublish(ctx, t, pool, "root", 1, nodeID, now)
		metadataPublishID := insertActiveRootOrMetadataPublish(ctx, t, pool, "metadata", 2, nodeID, now)
		if _, err := pool.Exec(ctx, `
UPDATE nodecontrol.node_inventory
SET operator_state='disabled',security_state='quarantined',identity_state='recovery_pending',
    resume_operator_state='enabled',identity_epoch=1,lineage_id=$2,
    active_root_publish_id=$3,active_root_version=1,
    active_metadata_publish_id=$4,active_metadata_version=1,updated_at=$5
WHERE node_id=$1`, nodeID, uuid.New(), rootPublishID, metadataPublishID, now.Add(time.Second)); err != nil {
			t.Fatal("set recovery signing inventory capture:", err)
		}
		recoveryOperationID := uuid.New()
		insertCommittedFence(ctx, t, pool, recoveryOperationID, 3, "recovery_activate", "node", now)
		recoveryID := uuid.New()
		incidentSetDigest := bytesOf(0x72, 32)
		if _, err := pool.Exec(ctx, `
INSERT INTO nodecontrol.node_recovery_sessions(
  recovery_id,authority_operation_id,authority_epoch,authority_sequence,node_id,identity_epoch,
  reason,version,status,incident_set_digest,resume_operator_state,created_at,updated_at)
VALUES($1,$2,1,3,$3,1,'authority_restore',1,'pending',$4,'disabled',$5,$5)`, recoveryID,
			recoveryOperationID, nodeID, incidentSetDigest, now.Add(time.Second)); err != nil {
			t.Fatal("insert captured recovery session:", err)
		}
		signingID := uuid.New()
		if _, err := pool.Exec(ctx, `
INSERT INTO nodecontrol.node_state_signing_intents(
  signing_id,authority_operation_id,authority_epoch,authority_sequence,node_id,signing_kind,
  idempotency_key_digest,base_generation,reserved_generation,canonical_payload,payload_digest,
  root_version,metadata_version,expected_key_id,expected_public_key_digest,captured_inventory_version,
  captured_identity_epoch,captured_security_version,recovery_id,recovery_reason,recovery_session_version,
  recovery_session_status,recovery_incident_set_digest,recovery_local_bindings_digest,
  recovery_supervisor_bindings_digest,recovery_required_action,activation_deadline,status,created_at,updated_at)
VALUES($1,$2,1,3,$3,'recovery',$4,0,1,decode('01','hex'),$5,1,1,$6,$7,1,1,1,
       $8,'authority_restore',1,'pending',$9,$10,$11,'hold_stopped',$12,'pending',$13,$13)`,
			signingID, recoveryOperationID, nodeID, bytesOf(0x73, 32), bytesOf(0x74, 32),
			bytesOf(0x75, 32), bytesOf(0x76, 32), recoveryID, incidentSetDigest,
			bytesOf(0x77, 32), bytesOf(0x78, 32), now.Add(10*time.Minute), now.Add(time.Second)); err != nil {
			t.Fatal("insert captured recovery signing intent:", err)
		}
		if _, err := pool.Exec(ctx, `UPDATE nodecontrol.node_recovery_sessions SET status='completed',terminal_at=$2,retention_until=$3,updated_at=$2 WHERE recovery_id=$1`, recoveryID, now.Add(2*time.Second), now.Add(181*24*time.Hour)); err != nil {
			t.Fatal("advance captured recovery session:", err)
		}
		assertSQLFailureContains(ctx, t, pool, activateSigningSQL, "recovery session", signingID, bytesOf(0x79, 64), now.Add(3*time.Second))
	})
}

func TestNodeControlMigrationEnforcesRestoreDualControl(t *testing.T) {
	t.Run("recovery belongs to the same node", func(t *testing.T) {
		ctx, pool := openMigratedNodeControlDatabase(t)
		now := time.Now().UTC().Truncate(time.Microsecond)
		nodeA := insertNodeControlFixtureNode(ctx, t, pool, "restore-node-a", now)
		nodeB := insertNodeControlFixtureNode(ctx, t, pool, "restore-node-b", now)
		recoveryID := insertCompletedRecoverySession(ctx, t, pool, nodeA, now)
		operationID := uuid.New()
		insertCommittedFence(ctx, t, pool, operationID, 2, "operator_transition", "node", now)
		if err := insertRestoreApproval(ctx, pool, uuid.New(), operationID, nodeB, recoveryID, "proposal", "operator-a", bytesOf(0x81, 32), now, now.Add(10*time.Minute)); err == nil {
			t.Fatal("restore proposal cross-bound a recovery session owned by another node")
		}
	})
	t.Run("proposal and approval exact evidence", func(t *testing.T) {
		ctx, pool := openMigratedNodeControlDatabase(t)
		now := time.Now().UTC().Truncate(time.Microsecond)
		nodeID := insertNodeControlFixtureNode(ctx, t, pool, "restore-evidence", now)
		recoveryID := insertCompletedRecoverySession(ctx, t, pool, nodeID, now)
		operationID := uuid.New()
		insertCommittedFence(ctx, t, pool, operationID, 2, "operator_transition", "node", now)
		if err := insertRestoreApproval(ctx, pool, uuid.New(), operationID, nodeID, recoveryID, "proposal", "operator-a", bytesOf(0x82, 32), now, now.Add(10*time.Minute)); err != nil {
			t.Fatal("insert restore proposal:", err)
		}
		if err := insertRestoreApproval(ctx, pool, uuid.New(), operationID, nodeID, recoveryID, "approval", "operator-b", bytesOf(0x83, 32), now, now.Add(10*time.Minute)); err == nil {
			t.Fatal("restore approval disagreed with proposal security-admin evidence")
		}
	})
	t.Run("expired approval", func(t *testing.T) {
		ctx, pool := openMigratedNodeControlDatabase(t)
		now := time.Now().UTC().Truncate(time.Microsecond)
		nodeID := insertNodeControlFixtureNode(ctx, t, pool, "restore-expired", now)
		recoveryID := insertCompletedRecoverySession(ctx, t, pool, nodeID, now)
		operationID := uuid.New()
		insertCommittedFence(ctx, t, pool, operationID, 2, "operator_transition", "node", now)
		createdAt := now.Add(-10 * time.Minute)
		if err := insertRestoreApproval(ctx, pool, uuid.New(), operationID, nodeID, recoveryID, "proposal", "operator-a", bytesOf(0x84, 32), createdAt, now.Add(-time.Minute)); err == nil {
			t.Fatal("restore workflow accepted an already expired approval row")
		}
	})
	t.Run("single consumption rejected and pair consumption atomic", func(t *testing.T) {
		ctx, pool := openMigratedNodeControlDatabase(t)
		now := time.Now().UTC().Truncate(time.Microsecond)
		nodeID := insertNodeControlFixtureNode(ctx, t, pool, "restore-consume", now)
		recoveryID := insertCompletedRecoverySession(ctx, t, pool, nodeID, now)
		operationID := uuid.New()
		insertCommittedFence(ctx, t, pool, operationID, 2, "operator_transition", "node", now)
		binding := bytesOf(0x85, 32)
		proposalID := uuid.New()
		approvalID := uuid.New()
		if err := insertRestoreApproval(ctx, pool, proposalID, operationID, nodeID, recoveryID, "proposal", "operator-a", binding, now, now.Add(10*time.Minute)); err != nil {
			t.Fatal("insert restore proposal:", err)
		}
		if err := insertRestoreApproval(ctx, pool, approvalID, operationID, nodeID, recoveryID, "approval", "operator-b", binding, now, now.Add(10*time.Minute)); err != nil {
			t.Fatal("insert restore approval:", err)
		}
		tx, err := pool.Begin(ctx)
		if err != nil {
			t.Fatal("begin single-consumption transaction:", err)
		}
		if _, err = tx.Exec(ctx, `UPDATE nodecontrol.node_restore_reauthorization_approvals SET status='consumed',terminal_at=$2,retention_until=$3 WHERE approval_id=$1`, proposalID, now.Add(time.Second), now.Add(181*24*time.Hour)); err != nil {
			t.Fatal("stage single consumption:", err)
		}
		if err = tx.Commit(ctx); err == nil {
			t.Fatal("restore dual control allowed only one approval row to be consumed")
		}

		tx, err = pool.Begin(ctx)
		if err != nil {
			t.Fatal("begin pair-consumption transaction:", err)
		}
		if _, err = tx.Exec(ctx, `UPDATE nodecontrol.node_restore_reauthorization_approvals SET status='consumed',terminal_at=$2,retention_until=$3 WHERE approval_id IN ($1,$4)`, proposalID, now.Add(2*time.Second), now.Add(181*24*time.Hour), approvalID); err != nil {
			t.Fatal("stage pair consumption:", err)
		}
		if err = tx.Commit(ctx); err != nil {
			t.Fatal("atomic exact pair consumption failed:", err)
		}
	})
}

func TestNodeControlMigrationSerializesCapacityReferences(t *testing.T) {
	t.Run("profile update versus first reference", func(t *testing.T) {
		ctx, pool := openMigratedNodeControlDatabase(t)
		now := time.Now().UTC().Truncate(time.Microsecond)
		nodeID, profileID := insertCapacityFixture(ctx, t, pool, "capacity-race", now)
		tx, err := pool.Begin(ctx)
		if err != nil {
			t.Fatal("begin slot transaction:", err)
		}
		if _, err = tx.Exec(ctx, `INSERT INTO nodecontrol.node_process_slots(node_id,slot_id,adapter,capacity_profile_id,capacity_profile_version,required,operator_state,inventory_version,created_at,updated_at) VALUES($1,1,'fixture',$2,1,true,'enabled',1,$3,$3)`, nodeID, profileID, now); err != nil {
			tx.Rollback(ctx)
			t.Fatal("insert first uncommitted slot:", err)
		}
		updateResult := make(chan error, 1)
		go func() {
			_, updateErr := pool.Exec(ctx, `UPDATE nodecontrol.node_capacity_profiles SET adapter='xray' WHERE profile_id=$1 AND version=1`, profileID)
			updateResult <- updateErr
		}()
		select {
		case updateErr := <-updateResult:
			tx.Rollback(ctx)
			if updateErr == nil {
				t.Fatal("capacity profile update raced past its first uncommitted slot reference")
			}
			t.Fatalf("capacity profile update returned early instead of serializing: %v", updateErr)
		case <-time.After(250 * time.Millisecond):
		}
		if err = tx.Commit(ctx); err != nil {
			t.Fatal("commit first slot reference:", err)
		}
		select {
		case updateErr := <-updateResult:
			if updateErr == nil {
				t.Fatal("capacity profile changed after waiting for its first committed reference")
			}
		case <-time.After(5 * time.Second):
			t.Fatal("capacity profile update remained blocked after slot commit")
		}
	})
	t.Run("concurrent eighth and ninth slot", func(t *testing.T) {
		ctx, pool := openMigratedNodeControlDatabase(t)
		now := time.Now().UTC().Truncate(time.Microsecond)
		nodeID, profileID := insertCapacityFixture(ctx, t, pool, "slot-cap-race", now)
		for slot := 1; slot <= 7; slot++ {
			if _, err := pool.Exec(ctx, `INSERT INTO nodecontrol.node_process_slots(node_id,slot_id,adapter,capacity_profile_id,capacity_profile_version,required,operator_state,inventory_version,created_at,updated_at) VALUES($1,$2,'fixture',$3,1,true,'enabled',1,$4,$4)`, nodeID, slot, profileID, now); err != nil {
				t.Fatalf("seed slot %d: %v", slot, err)
			}
		}
		first, err := pool.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = first.Exec(ctx, `INSERT INTO nodecontrol.node_process_slots(node_id,slot_id,adapter,capacity_profile_id,capacity_profile_version,required,operator_state,inventory_version,created_at,updated_at) VALUES($1,8,'fixture',$2,1,true,'enabled',1,$3,$3)`, nodeID, profileID, now); err != nil {
			first.Rollback(ctx)
			t.Fatal("stage eighth slot:", err)
		}
		second, err := pool.Begin(ctx)
		if err != nil {
			first.Rollback(ctx)
			t.Fatal(err)
		}
		secondResult := make(chan error, 1)
		go func() {
			_, insertErr := second.Exec(ctx, `INSERT INTO nodecontrol.node_process_slots(node_id,slot_id,adapter,capacity_profile_id,capacity_profile_version,required,operator_state,inventory_version,created_at,updated_at) VALUES($1,9,'fixture',$2,1,true,'enabled',1,$3,$3)`, nodeID, profileID, now)
			secondResult <- insertErr
		}()
		select {
		case insertErr := <-secondResult:
			first.Rollback(ctx)
			second.Rollback(ctx)
			t.Fatalf("ninth-slot contender did not serialize: %v", insertErr)
		case <-time.After(250 * time.Millisecond):
		}
		if err = first.Commit(ctx); err != nil {
			second.Rollback(ctx)
			t.Fatal("commit eighth slot:", err)
		}
		if insertErr := <-secondResult; insertErr == nil {
			second.Rollback(ctx)
			t.Fatal("concurrent slot contenders committed both slots eight and nine")
		}
		if err = second.Rollback(ctx); err != nil && err != pgx.ErrTxClosed {
			t.Fatal("rollback rejected ninth slot transaction:", err)
		}
	})
	t.Run("concurrent eighth and ninth root publish", func(t *testing.T) {
		ctx, pool := openMigratedNodeControlDatabase(t)
		now := time.Now().UTC().Truncate(time.Microsecond)
		operationIDs := make([]uuid.UUID, 9)
		for index := range operationIDs {
			operationIDs[index] = uuid.New()
			insertCommittedFence(ctx, t, pool, operationIDs[index], int64(index+1), "root_publish", "global_node_trust", now)
		}
		insertPendingRoot := func(executor interface {
			Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
		}, index int) error {
			_, err := executor.Exec(ctx, pendingRootPublishSQL, uuid.New(), operationIDs[index], int64(index+1), "root", "normal", int64(index), int64(0), int64(index+1), bytesOf(byte(0x80+index), 32), [][]byte{bytesOf(byte(0x90+index), 32)}, nil, 1, nil, now.Add(10*time.Minute), now)
			return err
		}
		for index := 0; index < 7; index++ {
			if err := insertPendingRoot(pool, index); err != nil {
				t.Fatalf("seed pending root publish %d: %v", index+1, err)
			}
		}
		first, err := pool.Begin(ctx)
		if err != nil {
			t.Fatal("begin eighth-root transaction:", err)
		}
		if err = insertPendingRoot(first, 7); err != nil {
			first.Rollback(ctx)
			t.Fatal("stage eighth pending root publish:", err)
		}
		second, err := pool.Begin(ctx)
		if err != nil {
			first.Rollback(ctx)
			t.Fatal("begin ninth-root transaction:", err)
		}
		secondResult := make(chan error, 1)
		go func() { secondResult <- insertPendingRoot(second, 8) }()
		select {
		case insertErr := <-secondResult:
			first.Rollback(ctx)
			second.Rollback(ctx)
			t.Fatalf("ninth-root contender did not serialize: %v", insertErr)
		case <-time.After(250 * time.Millisecond):
		}
		if err = first.Commit(ctx); err != nil {
			second.Rollback(ctx)
			t.Fatal("commit eighth pending root publish:", err)
		}
		if insertErr := <-secondResult; insertErr == nil {
			second.Rollback(ctx)
			t.Fatal("concurrent root-publish contenders committed both pending rows eight and nine")
		}
		if err = second.Rollback(ctx); err != nil && err != pgx.ErrTxClosed {
			t.Fatal("rollback rejected ninth-root transaction:", err)
		}
	})
	t.Run("concurrent final local fault binding slot", func(t *testing.T) {
		ctx, pool := openMigratedNodeControlDatabase(t)
		now := time.Now().UTC().Truncate(time.Microsecond)
		nodeID := insertNodeControlFixtureNode(ctx, t, pool, "binding-cap-race", now)
		operationID := uuid.New()
		insertCommittedFence(ctx, t, pool, operationID, 1, "security_incident_open", "node", now)
		incidentID := uuid.New()
		insertOpenIncident(ctx, t, pool, incidentID, operationID, 1, nodeID, "identity_compromise", 1, now)
		insertBinding := func(executor interface {
			Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
		}, slot int, discriminator byte) error {
			_, err := executor.Exec(ctx, `
INSERT INTO nodecontrol.node_security_fault_receipts(
  receipt_id,authority_operation_id,authority_epoch,authority_sequence,node_id,identity_epoch,
  local_fault_id,request_digest,fault_subtype,evidence_digest,agent_boot_id,incident_id,
  local_binding_slot,result,delivery_status,binding_status,created_at,updated_at)
VALUES($1,$2,1,1,$3,1,$4,$5,'identity_compromise',$6,$7,$8,$9,'accepted','pending','active',$10,$10)`,
				uuid.New(), operationID, nodeID, uuid.New(), bytesOf(discriminator, 32), bytesOf(discriminator+1, 32), uuid.New(), incidentID, slot, now)
			return err
		}
		for slot := 1; slot <= 63; slot++ {
			if err := insertBinding(pool, slot, byte(slot)); err != nil {
				t.Fatalf("seed active local binding %d: %v", slot, err)
			}
		}
		first, err := pool.Begin(ctx)
		if err != nil {
			t.Fatal("begin first final-slot transaction:", err)
		}
		if err = insertBinding(first, 64, 0xc1); err != nil {
			first.Rollback(ctx)
			t.Fatal("stage first final local binding slot:", err)
		}
		second, err := pool.Begin(ctx)
		if err != nil {
			first.Rollback(ctx)
			t.Fatal("begin second final-slot transaction:", err)
		}
		secondResult := make(chan error, 1)
		go func() { secondResult <- insertBinding(second, 64, 0xd1) }()
		select {
		case insertErr := <-secondResult:
			first.Rollback(ctx)
			second.Rollback(ctx)
			t.Fatalf("final-binding contender did not serialize: %v", insertErr)
		case <-time.After(250 * time.Millisecond):
		}
		if err = first.Commit(ctx); err != nil {
			second.Rollback(ctx)
			t.Fatal("commit first final local binding:", err)
		}
		if insertErr := <-secondResult; insertErr == nil {
			second.Rollback(ctx)
			t.Fatal("concurrent final local-binding contenders both committed")
		}
		if err = second.Rollback(ctx); err != nil && err != pgx.ErrTxClosed {
			t.Fatal("rollback second final-binding transaction:", err)
		}
	})
}

func TestNodeControlMigrationAllowsOnlySafeRetentionDeletes(t *testing.T) {
	t.Run("desired state before retention", func(t *testing.T) {
		ctx, pool := openMigratedNodeControlDatabase(t)
		now := time.Now().UTC().Truncate(time.Microsecond)
		nodeID, _ := insertDesiredStateFixture(ctx, t, pool, "desired-retained", now, now.Add(181*24*time.Hour))
		if _, err := pool.Exec(ctx, `DELETE FROM nodecontrol.node_desired_states WHERE node_id=$1 AND generation=1`, nodeID); err == nil {
			t.Fatal("desired state deleted before retention elapsed")
		}
	})
	t.Run("desired state after retention", func(t *testing.T) {
		ctx, pool := openMigratedNodeControlDatabase(t)
		createdAt := time.Now().UTC().Add(-181 * 24 * time.Hour).Truncate(time.Microsecond)
		nodeID, _ := insertDesiredStateFixture(ctx, t, pool, "desired-expired", createdAt, createdAt.Add(180*24*time.Hour))
		if _, err := pool.Exec(ctx, `DELETE FROM nodecontrol.node_desired_states WHERE node_id=$1 AND generation=1`, nodeID); err != nil {
			t.Fatalf("unreferenced desired state remained permanently undeletable after retention: %v", err)
		}
	})
	t.Run("desired state with outbox reference", func(t *testing.T) {
		ctx, pool := openMigratedNodeControlDatabase(t)
		createdAt := time.Now().UTC().Add(-181 * 24 * time.Hour).Truncate(time.Microsecond)
		nodeID, signingID := insertDesiredStateFixture(ctx, t, pool, "desired-outbox", createdAt, createdAt.Add(180*24*time.Hour))
		createOutboxReference(ctx, t, pool, "node_desired_state", signingID, 1)
		if _, err := pool.Exec(ctx, `DELETE FROM nodecontrol.node_desired_states WHERE node_id=$1 AND generation=1`, nodeID); err == nil {
			t.Fatal("desired state deleted while an outbox row still referenced it")
		}
	})
	t.Run("operator audit before and after retention", func(t *testing.T) {
		ctx, pool := openMigratedNodeControlDatabase(t)
		now := time.Now().UTC().Truncate(time.Microsecond)
		futureAuditID := insertOperatorAuditFixture(ctx, t, pool, now, now.Add(181*24*time.Hour))
		if _, err := pool.Exec(ctx, `DELETE FROM nodecontrol.node_operator_audit WHERE audit_id=$1`, futureAuditID); err == nil {
			t.Fatal("operator audit deleted before retention elapsed")
		}
		oldAt := now.Add(-181 * 24 * time.Hour)
		oldAuditID := insertOperatorAuditFixture(ctx, t, pool, oldAt, oldAt.Add(180*24*time.Hour))
		if _, err := pool.Exec(ctx, `DELETE FROM nodecontrol.node_operator_audit WHERE audit_id=$1`, oldAuditID); err != nil {
			t.Fatalf("unreferenced operator audit remained permanently undeletable after retention: %v", err)
		}
	})
}

func TestNodeControlMigrationDesiredPointerBindsExactResourceEnvelope(t *testing.T) {
	ctx, pool := openMigratedNodeControlDatabase(t)
	now := time.Now().UTC().Truncate(time.Microsecond)
	nodeID := insertNodeControlFixtureNode(ctx, t, pool, "desired-pointer", now)
	rootPublishID := insertActiveRootOrMetadataPublish(ctx, t, pool, "root", 1, nodeID, now)
	metadataPublishID := insertActiveRootOrMetadataPublish(ctx, t, pool, "metadata", 2, nodeID, now)
	firstEnvelopeDigest := bytesOf(0xc1, 32)
	insertResourceEnvelope(ctx, t, pool, nodeID, 1, 3, firstEnvelopeDigest, now)
	if _, err := pool.Exec(ctx, `
UPDATE nodecontrol.node_inventory
SET resource_envelope_version=1,resource_envelope_digest=$2,
    active_root_publish_id=$3,active_root_version=1,
    active_metadata_publish_id=$4,active_metadata_version=1,updated_at=$5
WHERE node_id=$1`, nodeID, firstEnvelopeDigest, rootPublishID, metadataPublishID, now.Add(time.Second)); err != nil {
		t.Fatal("set initial resource/root/metadata pointers:", err)
	}
	desiredOperationID := uuid.New()
	insertCommittedFence(ctx, t, pool, desiredOperationID, 4, "desired_activate", "node", now)
	signingID := uuid.New()
	signingKeyID := bytesOf(0xc2, 32)
	signature := bytesOf(0xc3, 64)
	if _, err := pool.Exec(ctx, pendingDesiredSigningSQL, signingID, desiredOperationID, int64(4), nodeID, signingKeyID, bytesOf(0xc4, 32), now, now.Add(5*time.Minute)); err != nil {
		t.Fatal("insert exact desired signing intent:", err)
	}
	if _, err := pool.Exec(ctx, activateSigningSQL, signingID, signature, now.Add(2*time.Second)); err != nil {
		t.Fatal("activate exact desired signing intent:", err)
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO nodecontrol.node_desired_states(
  node_id,generation,signing_id,authority_operation_id,authority_epoch,authority_sequence,
  inventory_version,resource_envelope_version,resource_envelope_digest,root_version,root_publish_id,
  metadata_version,metadata_publish_id,signing_key_id,canonical_payload,payload_digest,signature,
  issued_at,effective_deadline,valid_until,reason,created_at,retention_until)
VALUES($1,1,$2,$3,1,4,1,1,$4,1,$5,1,$6,$7,decode('01','hex'),decode(repeat('20',32),'hex'),
       $8,$9,$10,$11,'provision',$9,$12)`, nodeID, signingID, desiredOperationID, firstEnvelopeDigest,
		rootPublishID, metadataPublishID, signingKeyID, signature, now, now.Add(time.Hour), now.Add(2*time.Hour), now.Add(181*24*time.Hour)); err != nil {
		t.Fatal("insert exact desired state:", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE nodecontrol.node_inventory SET active_desired_generation=1,next_desired_generation=2,updated_at=$2 WHERE node_id=$1`, nodeID, now.Add(3*time.Second)); err != nil {
		t.Fatal("activate exact desired pointer:", err)
	}
	otherRootPublishID := insertActiveRootOrMetadataPublish(ctx, t, pool, "root", 5, nodeID, now.Add(4*time.Second))
	if _, err := pool.Exec(ctx, `UPDATE nodecontrol.node_inventory SET active_root_publish_id=$2,updated_at=$3 WHERE node_id=$1`, nodeID, otherRootPublishID, now.Add(5*time.Second)); err == nil {
		t.Fatal("active desired pointer accepted a different active root publish at the same version")
	}
	otherMetadataPublishID := insertActiveRootOrMetadataPublish(ctx, t, pool, "metadata", 6, nodeID, now.Add(5*time.Second))
	if _, err := pool.Exec(ctx, `UPDATE nodecontrol.node_inventory SET active_metadata_publish_id=$2,updated_at=$3 WHERE node_id=$1`, nodeID, otherMetadataPublishID, now.Add(6*time.Second)); err == nil {
		t.Fatal("active desired pointer accepted a different active metadata publish at the same version")
	}
	secondEnvelopeDigest := bytesOf(0xc5, 32)
	insertResourceEnvelope(ctx, t, pool, nodeID, 2, 7, secondEnvelopeDigest, now.Add(7*time.Second))
	if _, err := pool.Exec(ctx, `UPDATE nodecontrol.node_inventory SET resource_envelope_version=2,resource_envelope_digest=$2,updated_at=$3 WHERE node_id=$1`, nodeID, secondEnvelopeDigest, now.Add(8*time.Second)); err == nil {
		t.Fatal("active desired pointer remained valid after inventory moved to a different resource envelope")
	}
}

func TestNodeControlMigrationRejectsLateRootSupersession(t *testing.T) {
	t.Run("later transaction rejected", func(t *testing.T) {
		ctx, pool := openMigratedNodeControlDatabase(t)
		now := time.Now().UTC().Truncate(time.Microsecond)
		firstPublishID := insertActiveRootPublishVersion(ctx, t, pool, 1, 0, 1, now)
		_ = insertActiveRootPublishVersion(ctx, t, pool, 2, 1, 2, now.Add(2*time.Second))
		if _, err := pool.Exec(ctx, `
UPDATE nodecontrol.node_root_metadata_publish_intents
SET status='superseded',failure_reason='superseded',terminal_at=$2,updated_at=$2
WHERE publish_id=$1`, firstPublishID, now.Add(4*time.Second)); err == nil {
			t.Fatal("old active root was superseded in a transaction later than the next activation")
		}
	})
	t.Run("same next activation transaction allowed", func(t *testing.T) {
		ctx, pool := openMigratedNodeControlDatabase(t)
		now := time.Now().UTC().Truncate(time.Microsecond)
		firstPublishID := insertActiveRootPublishVersion(ctx, t, pool, 1, 0, 1, now)
		nextOperationID := uuid.New()
		insertCommittedFence(ctx, t, pool, nextOperationID, 2, "root_publish", "global_node_trust", now)
		tx, err := pool.Begin(ctx)
		if err != nil {
			t.Fatal("begin atomic root activation:", err)
		}
		nextPublishID := uuid.New()
		payloadDigest := bytesOf(0xe1, 32)
		keyID := bytesOf(0xe2, 32)
		if _, err = tx.Exec(ctx, pendingRootPublishSQL, nextPublishID, nextOperationID, int64(2), "root", "normal", int64(1), int64(0), int64(2), payloadDigest, [][]byte{keyID}, nil, 1, nil, now.Add(5*time.Minute), now.Add(2*time.Second)); err != nil {
			tx.Rollback(ctx)
			t.Fatal("insert atomic next root intent:", err)
		}
		if _, err = tx.Exec(ctx, `INSERT INTO nodecontrol.node_root_metadata_signature_shares(publish_id,key_id,physical_key_id,payload_digest,signature_role,signature,verified_at) VALUES($1,$2,$3,$4,'current_root',$5,$6)`, nextPublishID, keyID, bytesOf(0xe3, 32), payloadDigest, bytesOf(0xe4, 64), now.Add(2*time.Second)); err != nil {
			tx.Rollback(ctx)
			t.Fatal("insert atomic next root share:", err)
		}
		if _, err = tx.Exec(ctx, `UPDATE nodecontrol.node_root_metadata_publish_intents SET published_envelope=decode('01','hex'),published_envelope_digest=$2,status='active',terminal_at=$3,updated_at=$3 WHERE publish_id=$1`, nextPublishID, bytesOf(0xe5, 32), now.Add(3*time.Second)); err != nil {
			tx.Rollback(ctx)
			t.Fatal("activate atomic next root:", err)
		}
		if _, err = tx.Exec(ctx, `UPDATE nodecontrol.node_root_metadata_publish_intents SET status='superseded',failure_reason='superseded',terminal_at=$2,updated_at=$2 WHERE publish_id=$1`, firstPublishID, now.Add(3*time.Second)); err != nil {
			tx.Rollback(ctx)
			t.Fatal("supersede prior root in next activation transaction:", err)
		}
		if err = tx.Commit(ctx); err != nil {
			t.Fatal("commit atomic root activation and supersession:", err)
		}
	})
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

func openMigratedNodeControlDatabase(t *testing.T) (context.Context, *pgxpool.Pool) {
	t.Helper()
	upSQL, _ := loadNodeControlMigrationSections(t)
	pool := openOwnedNodeControlDatabase(t)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	t.Cleanup(cancel)
	if _, err := pool.Exec(ctx, upSQL); err != nil {
		t.Fatal("nodecontrol Up failed:", err)
	}
	return ctx, pool
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
		assertCatalogColumnChecks(ctx, t, pool, table)
		assertCatalogIndexes(ctx, t, pool, table)
		assertCatalogTriggers(ctx, t, pool, table)
	}
	rows, err = pool.Query(ctx, `
SELECT p.proname, n.nspname, lang.lanname,
       CASE p.provolatile WHEN 'i' THEN 'immutable' WHEN 's' THEN 'stable' ELSE 'volatile' END,
       p.prosecdef, pg_catalog.pg_get_functiondef(p.oid)
FROM pg_catalog.pg_proc p
JOIN pg_catalog.pg_namespace n ON n.oid=p.pronamespace
JOIN pg_catalog.pg_language lang ON lang.oid=p.prolang
WHERE n.nspname='nodecontrol'
ORDER BY p.proname`)
	if err != nil {
		t.Fatal("query nodecontrol functions:", err)
	}
	functions := make([]nodeControlCatalogFunction, 0, len(expectedNodeControlFunctionCatalog))
	for rows.Next() {
		var function nodeControlCatalogFunction
		if err := rows.Scan(&function.name, &function.schema, &function.language, &function.volatility, &function.securityDefiner, &function.definition); err != nil {
			rows.Close()
			t.Fatal("scan nodecontrol function:", err)
		}
		functions = append(functions, function)
	}
	if err := rows.Err(); err != nil {
		t.Fatal("iterate nodecontrol functions:", err)
	}
	rows.Close()
	if err := validateNodeControlFunctionCatalog(functions, expectedNodeControlFunctionCatalog); err != nil {
		t.Fatal(err)
	}
}

func assertCatalogColumnChecks(ctx context.Context, t *testing.T, pool *pgxpool.Pool, table nodeControlTableSpec) {
	t.Helper()
	compiled := make(map[string]string)
	tableIdentifier := pgx.Identifier{"nodecontrol", table.Name}.Sanitize()
	probeIdentifier := pgx.Identifier{"nodecontrol_catalog_check_probe"}.Sanitize()
	for _, column := range table.Columns {
		want := make([]string, 0, len(column.CheckSQL))
		for _, checkSQL := range column.CheckSQL {
			definition, exists := compiled[checkSQL]
			if !exists {
				if _, err := pool.Exec(ctx, "ALTER TABLE "+tableIdentifier+" ADD CONSTRAINT "+probeIdentifier+" CHECK ("+checkSQL+") NOT VALID"); err != nil {
					t.Fatalf("compile %s.%s check_sql %q: %v", table.Name, column.Name, checkSQL, err)
				}
				if err := pool.QueryRow(ctx, `
SELECT pg_catalog.pg_get_constraintdef(con.oid,true)
FROM pg_catalog.pg_constraint con
JOIN pg_catalog.pg_class rel ON rel.oid=con.conrelid
JOIN pg_catalog.pg_namespace n ON n.oid=rel.relnamespace
WHERE n.nspname='nodecontrol' AND rel.relname=$1 AND con.conname='nodecontrol_catalog_check_probe'`, table.Name).Scan(&definition); err != nil {
					t.Fatalf("read compiled %s.%s check_sql: %v", table.Name, column.Name, err)
				}
				if _, err := pool.Exec(ctx, "ALTER TABLE "+tableIdentifier+" DROP CONSTRAINT "+probeIdentifier); err != nil {
					t.Fatalf("drop %s.%s check_sql probe: %v", table.Name, column.Name, err)
				}
				definition = normalizeCatalogSQL(strings.TrimSuffix(definition, " NOT VALID"))
				compiled[checkSQL] = definition
			}
			want = append(want, definition)
		}
		rows, err := pool.Query(ctx, `
SELECT pg_catalog.pg_get_constraintdef(con.oid,true)
FROM pg_catalog.pg_constraint con
JOIN pg_catalog.pg_class rel ON rel.oid=con.conrelid
JOIN pg_catalog.pg_namespace n ON n.oid=rel.relnamespace
JOIN pg_catalog.pg_attribute att ON att.attrelid=rel.oid AND att.attname=$2
WHERE n.nspname='nodecontrol' AND rel.relname=$1 AND con.contype='c' AND att.attnum=ANY(con.conkey)
ORDER BY con.conname`, table.Name, column.Name)
		if err != nil {
			t.Fatalf("query catalog checks for %s.%s: %v", table.Name, column.Name, err)
		}
		got, err := pgx.CollectRows(rows, pgx.RowTo[string])
		if err != nil {
			t.Fatalf("collect catalog checks for %s.%s: %v", table.Name, column.Name, err)
		}
		for index := range got {
			got[index] = normalizeCatalogSQL(got[index])
		}
		got = sortedStrings(got)
		want = sortedStrings(want)
		if !equalStrings(got, want) {
			t.Errorf("%s.%s catalog-linked check_sql = %v, want exactly %v", table.Name, column.Name, got, want)
		}
	}
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
SELECT trg.tgname, trg.tgtype::integer, procns.nspname, proc.proname || '()',
       pg_catalog.pg_get_expr(trg.tgqual,trg.tgrelid,true)
FROM pg_catalog.pg_trigger trg
JOIN pg_catalog.pg_class rel ON rel.oid=trg.tgrelid
JOIN pg_catalog.pg_namespace n ON n.oid=rel.relnamespace
JOIN pg_catalog.pg_proc proc ON proc.oid=trg.tgfoid
JOIN pg_catalog.pg_namespace procns ON procns.oid=proc.pronamespace
WHERE n.nspname='nodecontrol' AND rel.relname=$1 AND NOT trg.tgisinternal
ORDER BY trg.tgname`, table.Name)
	if err != nil {
		t.Fatal("query triggers for "+table.Name+":", err)
	}
	type catalogTrigger struct {
		name, functionSchema, function string
		typeBits                       int
		whenSQL                        sql.NullString
	}
	actual := make(map[string]catalogTrigger)
	for rows.Next() {
		var trigger catalogTrigger
		if err := rows.Scan(&trigger.name, &trigger.typeBits, &trigger.functionSchema, &trigger.function, &trigger.whenSQL); err != nil {
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
		if got.functionSchema != "nodecontrol" || triggerTiming(got.typeBits) != want.Timing || !equalStrings(triggerEvents(got.typeBits), want.Events) || got.function != want.Function || !equalOptionalNormalizedSQL(got.whenSQL, want.WhenSQL) {
			t.Fatalf("catalog trigger %s mismatch: timing=%s events=%v function=%s.%s when=%v, manifest=%+v", want.Name, triggerTiming(got.typeBits), triggerEvents(got.typeBits), got.functionSchema, got.function, got.whenSQL, want)
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

func insertPendingFence(ctx context.Context, t *testing.T, pool *pgxpool.Pool, operationID uuid.UUID, sequence int64, effectKind, scopeKind string, now time.Time) {
	t.Helper()
	_, err := pool.Exec(ctx, `
INSERT INTO nodecontrol.control_plane_authority_fences(
  operation_id,effect_kind,scope_kind,authority_epoch,authority_sequence,scope_digest,
  provider_reservation_digest,provider_status,visibility_state,reserved_at)
VALUES($1,$2,$3,1,$4,$5,$6,'reserved','fence_pending',$7)`, operationID, effectKind, scopeKind,
		sequence, bytesOf(0x75, 32), bytesOf(0x76, 32), now)
	if err != nil {
		t.Fatal("insert pending authority fence:", err)
	}
}

func insertCommittedFence(ctx context.Context, t *testing.T, pool *pgxpool.Pool, operationID uuid.UUID, sequence int64, effectKind, scopeKind string, now time.Time) {
	t.Helper()
	_, err := pool.Exec(ctx, `
INSERT INTO nodecontrol.control_plane_authority_fences(
  operation_id,effect_kind,scope_kind,authority_epoch,authority_sequence,scope_digest,
  provider_reservation_digest,provider_status,visibility_state,reserved_at)
VALUES($1,$2,$3,1,$4,$5,$6,'reserved','fence_pending',$7)`,
		operationID, effectKind, scopeKind, sequence, bytesOf(0x71, 32), bytesOf(0x72, 32), now)
	if err == nil {
		_, err = pool.Exec(ctx, `
UPDATE nodecontrol.control_plane_authority_fences
SET effect_digest=$2,provider_status='committed',provider_receipt_digest=$3,
    db_system_id=1,db_timeline=1,required_lsn='0/1',visibility_state='active',
    effect_bound_at=$4,terminal_at=$4
WHERE operation_id=$1`, operationID, bytesOf(0x73, 32), bytesOf(0x74, 32), now)
	}
	if err != nil {
		t.Fatal("insert committed authority fence:", err)
	}
}

func insertTrustHighWater(ctx context.Context, t *testing.T, pool *pgxpool.Pool, operationID uuid.UUID, sequence, bundleVersion int64, cumulativeCount int, bundleByte, cumulativeByte byte, now time.Time) {
	t.Helper()
	if _, err := pool.Exec(ctx, `
INSERT INTO nodecontrol.control_plane_trust_bundle_high_waters(
  purpose,listener_kind,trust_domain,authority_operation_id,authority_epoch,authority_sequence,
  bundle_version,bundle_digest,cumulative_set_digest,cumulative_set_count,updated_at)
VALUES('bootstrap_server','bootstrap','example.com',$1,1,$2,$3,$4,$5,$6,$7)`,
		operationID, sequence, bundleVersion, bytesOf(bundleByte, 32), bytesOf(cumulativeByte, 32), cumulativeCount, now); err != nil {
		t.Fatal("insert trust bundle high-water:", err)
	}
}

func assertSQLFailureContains(ctx context.Context, t *testing.T, pool *pgxpool.Pool, statement, want string, args ...any) {
	t.Helper()
	if _, err := pool.Exec(ctx, statement, args...); err == nil {
		t.Fatalf("SQL unexpectedly succeeded; wanted error containing %q", want)
	} else if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(want)) {
		t.Fatalf("SQL error %q does not contain %q", err, want)
	}
}

func bytesOf(value byte, count int) []byte {
	result := make([]byte, count)
	for index := range result {
		result[index] = value
	}
	return result
}

func insertNodeControlFixtureNode(ctx context.Context, t *testing.T, pool *pgxpool.Pool, popCode string, now time.Time) uuid.UUID {
	t.Helper()
	nodeID := uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO nodecontrol.node_pops(pop_code,iso_country,region,operator_state,created_at,updated_at) VALUES($1,'US','test-region','enabled',$2,$2)`, popCode, now); err != nil {
		t.Fatal("insert fixture POP:", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO nodecontrol.node_inventory(node_id,pop_code,operator_state,security_state,identity_state,created_at,updated_at) VALUES($1,$2,'enabled','normal','never_enrolled',$3,$3)`, nodeID, popCode, now); err != nil {
		t.Fatal("insert fixture node:", err)
	}
	return nodeID
}

func insertCapacityFixture(ctx context.Context, t *testing.T, pool *pgxpool.Pool, popCode string, now time.Time) (uuid.UUID, uuid.UUID) {
	t.Helper()
	nodeID := insertNodeControlFixtureNode(ctx, t, pool, popCode, now)
	profileID := uuid.New()
	if _, err := pool.Exec(ctx, `
INSERT INTO nodecontrol.node_capacity_profiles(
  profile_id,version,adapter,egress_limit_bps,connection_limit,handshake_limit_per_second,
  cpu_quota_millicores,cpu_limit_basis_points,memory_limit_bytes,task_limit,file_descriptor_limit,
  queue_limit,packet_loss_limit_basis_points,required_metrics,created_at)
VALUES($1,1,'fixture',1000000,1,1,100,1,67108864,32,64,1,1,ARRAY['cpu_usage_basis_points']::text[],$2)`, profileID, now); err != nil {
		t.Fatal("insert capacity fixture profile:", err)
	}
	return nodeID, profileID
}

func insertOpenIncident(ctx context.Context, t *testing.T, pool *pgxpool.Pool, incidentID, operationID uuid.UUID, sequence int64, nodeID uuid.UUID, subtype string, slot int, now time.Time) {
	t.Helper()
	if _, err := pool.Exec(ctx, `
INSERT INTO nodecontrol.node_security_incidents(
  incident_id,authority_operation_id,authority_epoch,authority_sequence,node_id,identity_epoch,
  fault_subtype,subtype_slot,status,first_evidence_digest,last_evidence_digest,occurrence_count,
  trust_context_digest,first_occurred_at,last_occurred_at)
VALUES($1,$2,1,$3,$4,1,$5,$6,'open',$7,$7,1,$8,$9,$9)`,
		incidentID, operationID, sequence, nodeID, subtype, slot, bytesOf(byte(0x70+slot), 32), bytesOf(byte(0x50+slot), 32), now); err != nil {
		t.Fatal("insert open security incident:", err)
	}
}

func insertActiveFaultReceipt(ctx context.Context, t *testing.T, pool *pgxpool.Pool, receiptID, operationID uuid.UUID, sequence int64, nodeID, incidentID uuid.UUID, subtype string, slot int, now time.Time) {
	t.Helper()
	if _, err := pool.Exec(ctx, `
INSERT INTO nodecontrol.node_security_fault_receipts(
  receipt_id,authority_operation_id,authority_epoch,authority_sequence,node_id,identity_epoch,
  local_fault_id,request_digest,fault_subtype,evidence_digest,agent_boot_id,incident_id,
  local_binding_slot,result,delivery_status,binding_status,created_at,updated_at)
VALUES($1,$2,1,$3,$4,1,$5,$6,$7,$8,$9,$10,$11,'accepted','pending','active',$12,$12)`,
		receiptID, operationID, sequence, nodeID, uuid.New(), bytesOf(byte(0x20+slot), 32), subtype,
		bytesOf(byte(0x30+slot), 32), uuid.New(), incidentID, slot, now); err != nil {
		t.Fatal("insert active security-fault receipt:", err)
	}
}

func insertCompletedRecoverySession(ctx context.Context, t *testing.T, pool *pgxpool.Pool, nodeID uuid.UUID, now time.Time) uuid.UUID {
	t.Helper()
	operationID := uuid.New()
	insertCommittedFence(ctx, t, pool, operationID, 1, "recovery_activate", "node", now)
	recoveryID := uuid.New()
	if _, err := pool.Exec(ctx, `
INSERT INTO nodecontrol.node_recovery_sessions(
  recovery_id,authority_operation_id,authority_epoch,authority_sequence,node_id,identity_epoch,
  reason,version,status,incident_set_digest,resume_operator_state,created_at,updated_at)
VALUES($1,$2,1,1,$3,1,'authority_restore',1,'pending',$4,'disabled',$5,$5)`,
		recoveryID, operationID, nodeID, bytesOf(0x91, 32), now); err != nil {
		t.Fatal("insert pending recovery session:", err)
	}
	if _, err := pool.Exec(ctx, `
UPDATE nodecontrol.node_recovery_sessions
SET status='completed',terminal_at=$2,retention_until=$3,updated_at=$2
WHERE recovery_id=$1`, recoveryID, now.Add(time.Second), now.Add(181*24*time.Hour)); err != nil {
		t.Fatal("complete recovery session:", err)
	}
	return recoveryID
}

func insertRestoreApproval(ctx context.Context, pool *pgxpool.Pool, approvalID, operationID, nodeID, recoveryID uuid.UUID, role, operatorID string, securityBinding []byte, createdAt, expiresAt time.Time) error {
	_, err := pool.Exec(ctx, `
INSERT INTO nodecontrol.node_restore_reauthorization_approvals(
  approval_id,authority_operation_id,authority_epoch,authority_sequence,node_id,recovery_id,
  effect_digest,scope_digest,role,operator_id,credential_digest,leaf_der_sha256,
  operator_authority_epoch,operator_authority_sequence,authorizer_version,
  security_admin_binding_digest,pop_scope,evidence_completed_at,credential_expires_at,
  created_at,expires_at,status)
VALUES($1,$2,1,2,$3,$4,$5,$6,$7,$8,$9,$10,1,1,1,$11,'restore-scope',$12,$13,$12,$14,'pending')`,
		approvalID, operationID, nodeID, recoveryID, bytesOf(0x73, 32), bytesOf(0x71, 32), role, operatorID,
		bytesOf(0x92, 32), bytesOf(0x93, 32), securityBinding, createdAt, createdAt.Add(15*time.Minute), expiresAt)
	return err
}

func insertDesiredStateFixture(ctx context.Context, t *testing.T, pool *pgxpool.Pool, popCode string, createdAt, retentionUntil time.Time) (uuid.UUID, uuid.UUID) {
	t.Helper()
	nodeID := insertNodeControlFixtureNode(ctx, t, pool, popCode, createdAt)
	rootOperationID := uuid.New()
	insertCommittedFence(ctx, t, pool, rootOperationID, 1, "root_publish", "global_node_trust", createdAt)
	rootPublishID := uuid.New()
	if _, err := pool.Exec(ctx, pendingRootPublishSQL, rootPublishID, rootOperationID, int64(1), "root", "normal", int64(0), int64(0), int64(1), bytesOf(0xa1, 32), [][]byte{bytesOf(0xa2, 32)}, nil, 1, nil, createdAt.Add(5*time.Minute), createdAt); err != nil {
		t.Fatal("insert desired fixture root intent:", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE nodecontrol.node_root_metadata_publish_intents SET status='failed',failure_reason='validation_failed',terminal_at=$2,updated_at=$2 WHERE publish_id=$1`, rootPublishID, createdAt.Add(time.Second)); err != nil {
		t.Fatal("terminate desired fixture root intent:", err)
	}
	metadataOperationID := uuid.New()
	insertCommittedFence(ctx, t, pool, metadataOperationID, 2, "metadata_publish", "global_node_trust", createdAt)
	metadataPublishID := uuid.New()
	if _, err := pool.Exec(ctx, pendingRootPublishSQL, metadataPublishID, metadataOperationID, int64(2), "metadata", "normal", int64(1), int64(0), int64(1), bytesOf(0xa3, 32), [][]byte{bytesOf(0xa4, 32)}, nil, 1, nil, createdAt.Add(5*time.Minute), createdAt); err != nil {
		t.Fatal("insert desired fixture metadata intent:", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE nodecontrol.node_root_metadata_publish_intents SET status='failed',failure_reason='validation_failed',terminal_at=$2,updated_at=$2 WHERE publish_id=$1`, metadataPublishID, createdAt.Add(time.Second)); err != nil {
		t.Fatal("terminate desired fixture metadata intent:", err)
	}
	desiredOperationID := uuid.New()
	insertCommittedFence(ctx, t, pool, desiredOperationID, 3, "desired_activate", "node", createdAt)
	signingID := uuid.New()
	signingKeyID := bytesOf(0xa5, 32)
	if _, err := pool.Exec(ctx, pendingDesiredSigningSQL, signingID, desiredOperationID, int64(3), nodeID, signingKeyID, bytesOf(0xa6, 32), createdAt, createdAt.Add(5*time.Minute)); err != nil {
		t.Fatal("insert desired fixture signing intent:", err)
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO nodecontrol.node_desired_states(
  node_id,generation,signing_id,authority_operation_id,authority_epoch,authority_sequence,
  inventory_version,resource_envelope_version,resource_envelope_digest,root_version,root_publish_id,
  metadata_version,metadata_publish_id,signing_key_id,canonical_payload,payload_digest,signature,
  issued_at,effective_deadline,valid_until,reason,created_at,retention_until)
VALUES($1,1,$2,$3,1,3,1,1,$4,1,$5,1,$6,$7,decode('01','hex'),$8,$9,$10,$11,$12,
       'provision',$10,$13)`, nodeID, signingID, desiredOperationID, bytesOf(0xa7, 32), rootPublishID,
		metadataPublishID, signingKeyID, bytesOf(0x20, 32), bytesOf(0xa8, 64), createdAt,
		createdAt.Add(time.Hour), createdAt.Add(2*time.Hour), retentionUntil); err != nil {
		t.Fatal("insert desired state fixture:", err)
	}
	return nodeID, signingID
}

func createOutboxReference(ctx context.Context, t *testing.T, pool *pgxpool.Pool, aggregateType string, aggregateID uuid.UUID, aggregateVersion int64) {
	t.Helper()
	if _, err := pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS public.transactional_outbox(aggregate_type text NOT NULL,aggregate_id uuid NOT NULL,aggregate_version bigint NOT NULL)`); err != nil {
		t.Fatal("create outbox fixture:", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO public.transactional_outbox(aggregate_type,aggregate_id,aggregate_version) VALUES($1,$2,$3)`, aggregateType, aggregateID, aggregateVersion); err != nil {
		t.Fatal("insert outbox fixture:", err)
	}
}

func insertOperatorAuditFixture(ctx context.Context, t *testing.T, pool *pgxpool.Pool, occurredAt, retentionUntil time.Time) uuid.UUID {
	t.Helper()
	auditID := uuid.New()
	if _, err := pool.Exec(ctx, `
INSERT INTO nodecontrol.node_operator_audit(
  audit_id,command_id,operator_id,credential_digest,role,action,target_kind,target_id,reason,
  result,occurred_at,retention_until)
VALUES($1,$2,'operator-a',$3,'inventory_writer','update_inventory','node',$4,
       'inventory_update','accepted',$5,$6)`, auditID, uuid.New(), bytesOf(0xb1, 32), uuid.NewString(), occurredAt, retentionUntil); err != nil {
		t.Fatal("insert operator audit fixture:", err)
	}
	return auditID
}

func insertActiveRootOrMetadataPublish(ctx context.Context, t *testing.T, pool *pgxpool.Pool, kind string, sequence int64, _ uuid.UUID, now time.Time) uuid.UUID {
	t.Helper()
	operationID := uuid.New()
	effectKind := kind + "_publish"
	insertCommittedFence(ctx, t, pool, operationID, sequence, effectKind, "global_node_trust", now)
	publishID := uuid.New()
	keyID := bytesOf(byte(0xd0+sequence), 32)
	physicalKeyID := bytesOf(byte(0xe0+sequence), 32)
	payloadDigest := bytesOf(byte(0xb0+sequence), 32)
	baseRootVersion := int64(0)
	baseMetadataVersion := int64(0)
	if kind == "metadata" {
		baseRootVersion = 1
	}
	if _, err := pool.Exec(ctx, pendingRootPublishSQL, publishID, operationID, sequence, kind, "normal", baseRootVersion, baseMetadataVersion, int64(1), payloadDigest, [][]byte{keyID}, nil, 1, nil, now.Add(5*time.Minute), now); err != nil {
		t.Fatal("insert active-publish fixture intent:", err)
	}
	role := "current_root"
	if kind == "metadata" {
		role = "metadata"
	}
	if _, err := pool.Exec(ctx, `INSERT INTO nodecontrol.node_root_metadata_signature_shares(publish_id,key_id,physical_key_id,payload_digest,signature_role,signature,verified_at) VALUES($1,$2,$3,$4,$5,$6,$7)`, publishID, keyID, physicalKeyID, payloadDigest, role, bytesOf(0xd9, 64), now); err != nil {
		t.Fatal("insert active-publish fixture share:", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE nodecontrol.node_root_metadata_publish_intents SET published_envelope=decode('01','hex'),published_envelope_digest=$2,status='active',terminal_at=$3,updated_at=$3 WHERE publish_id=$1`, publishID, bytesOf(byte(0xa0+sequence), 32), now.Add(time.Second)); err != nil {
		t.Fatal("activate root/metadata publish fixture:", err)
	}
	return publishID
}

func insertActiveRootPublishVersion(ctx context.Context, t *testing.T, pool *pgxpool.Pool, sequence, baseVersion, reservedVersion int64, now time.Time) uuid.UUID {
	t.Helper()
	operationID := uuid.New()
	insertCommittedFence(ctx, t, pool, operationID, sequence, "root_publish", "global_node_trust", now)
	publishID := uuid.New()
	keyID := bytesOf(byte(0x30+sequence), 32)
	payloadDigest := bytesOf(byte(0x40+sequence), 32)
	if _, err := pool.Exec(ctx, pendingRootPublishSQL, publishID, operationID, sequence, "root", "normal", baseVersion, int64(0), reservedVersion, payloadDigest, [][]byte{keyID}, nil, 1, nil, now.Add(5*time.Minute), now); err != nil {
		t.Fatal("insert versioned root intent:", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO nodecontrol.node_root_metadata_signature_shares(publish_id,key_id,physical_key_id,payload_digest,signature_role,signature,verified_at) VALUES($1,$2,$3,$4,'current_root',$5,$6)`, publishID, keyID, bytesOf(byte(0x50+sequence), 32), payloadDigest, bytesOf(0x60, 64), now); err != nil {
		t.Fatal("insert versioned root share:", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE nodecontrol.node_root_metadata_publish_intents SET published_envelope=decode('01','hex'),published_envelope_digest=$2,status='active',terminal_at=$3,updated_at=$3 WHERE publish_id=$1`, publishID, bytesOf(byte(0x70+sequence), 32), now.Add(time.Second)); err != nil {
		t.Fatal("activate versioned root intent:", err)
	}
	return publishID
}

func insertResourceEnvelope(ctx context.Context, t *testing.T, pool *pgxpool.Pool, nodeID uuid.UUID, version, sequence int64, digest []byte, now time.Time) {
	t.Helper()
	operationID := uuid.New()
	insertCommittedFence(ctx, t, pool, operationID, sequence, "resource_envelope_activate", "node", now)
	if _, err := pool.Exec(ctx, `
INSERT INTO nodecontrol.node_resource_envelopes(
  node_id,envelope_version,authority_operation_id,authority_epoch,authority_sequence,envelope_digest,
  canonical_package,deployment_key_id,signature,max_slots,agent_cpu_millicores,agent_memory_bytes,
  agent_task_limit,agent_file_descriptor_limit,supervisor_cpu_millicores,supervisor_memory_bytes,
  supervisor_task_limit,supervisor_file_descriptor_limit,core_parent_cpu_millicores,
  core_parent_memory_bytes,core_parent_task_limit,core_parent_file_descriptor_limit,
  aggregate_slot_file_descriptor_limit,aggregate_slot_tmpfs_bytes,aggregate_slot_tmpfs_inodes,
  detected_host_capacity_digest,issued_at,created_at)
VALUES($1,$2,$3,1,$4,$5,decode('01','hex'),$6,$7,8,100,67108864,32,64,100,67108864,
       32,64,100,67108864,32,64,64,1048576,1,$8,$9,$9)`, nodeID, version, operationID, sequence,
		digest, bytesOf(0xf1, 32), bytesOf(0xf2, 64), bytesOf(0xf3, 32), now); err != nil {
		t.Fatal("insert resource envelope fixture:", err)
	}
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
