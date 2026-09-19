//go:build integration

package testinfra

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const c12MarkerCreateSQL = `CREATE TABLE public.c12_authority_pitr_markers(marker text PRIMARY KEY CHECK(marker ~ '^[0-9a-f]{64}$'))`
const c12RequiredObjectsSQL = `LOCK TABLE public.c12_authority_pitr_markers,nodecontrol.control_plane_authority_fences,nodecontrol.node_certificates,nodecontrol.authority_task7_crash_effects,nodecontrol.node_operator_audit,public.transactional_outbox IN ACCESS SHARE MODE`

func (state *c12AuthorityPITRState) initializeC12SQLCapabilities(ctx context.Context) error {
	if !c12AuthorityPITRRunPattern.MatchString(state.descriptor.RunSuffix) || state.database == nil {
		return C12PITRInvalidHandle
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	tx, err := state.database.BeginTx(ctx, nil)
	if err != nil {
		return C12PITRDependencyFailure
	}
	defer tx.Rollback()
	var systemText, lsn string
	var timeline int64
	if err := tx.QueryRowContext(ctx, c12GetNodeControlDatabaseIdentitySQL).Scan(&systemText, &timeline, &lsn); err != nil {
		return C12PITRDependencyFailure
	}
	systemID, err := strconv.ParseUint(systemText, 10, 64)
	if err != nil || systemID == 0 || timeline <= 0 || !c12LSN(lsn) {
		return C12PITRDependencyFailure
	}
	roleName := "talenro_c12_" + state.descriptor.RunSuffix[:24] + "_candidate"
	role := pgx.Identifier{roleName}.Sanitize()
	database := pgx.Identifier{"talenro_c12_" + state.descriptor.RunSuffix}.Sanitize()
	var secret [32]byte
	if _, err := rand.Read(secret[:]); err != nil {
		return C12PITRDependencyFailure
	}
	password := hex.EncodeToString(secret[:])
	clear(secret[:])
	statements := []string{
		"CREATE ROLE " + role + " LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS PASSWORD '" + password + "'",
		c12MarkerCreateSQL, c12AuxDDLSQL,
		"GRANT CONNECT ON DATABASE " + database + " TO " + role,
		"GRANT USAGE ON SCHEMA nodecontrol,public TO " + role,
		"GRANT SELECT ON nodecontrol.control_plane_authority_fences,nodecontrol.node_certificates,nodecontrol.authority_task7_crash_effects,nodecontrol.node_operator_audit,public.transactional_outbox TO " + role,
		"GRANT UPDATE(operation_id) ON nodecontrol.control_plane_authority_fences,nodecontrol.authority_task7_crash_effects TO " + role,
		"GRANT UPDATE(certificate_id) ON nodecontrol.node_certificates TO " + role,
		"GRANT EXECUTE ON FUNCTION pg_catalog.pg_control_system(),pg_catalog.pg_control_checkpoint(),pg_catalog.pg_current_wal_insert_lsn() TO " + role,
		c12RequiredObjectsSQL,
	}
	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return C12PITRDependencyFailure
		}
	}
	if err := tx.Commit(); err != nil {
		return C12PITRIndeterminate
	}
	state.candidateRole, state.candidatePassword = roleName, password
	state.fixture = &c12FixtureLedger{epoch: 31, systemID: systemID, timeline: timeline, certificates: make(map[uuid.UUID]c12FixtureCertificate)}
	return nil
}

func (state *c12AuthorityPITRState) closeC12Setup(ctx context.Context) error {
	state.fixtureMu.Lock()
	state.setupClosed = true
	state.fixtureMu.Unlock()
	tx, err := state.database.BeginTx(ctx, nil)
	if err != nil {
		return C12PITRDependencyFailure
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, c12RequiredObjectsSQL); err != nil {
		return C12PITRDependencyFailure
	}
	if err := tx.Commit(); err != nil {
		return C12PITRDependencyFailure
	}
	return nil
}

type c12SQLMode uint8

const (
	c12SQLPrimary   c12SQLMode = 1
	c12SQLSetup     c12SQLMode = 2
	c12SQLCandidate c12SQLMode = 4
)

type c12SQLCall uint8

const (
	c12SQLExec c12SQLCall = iota
	c12SQLQuery
	c12SQLQueryRow
)

type c12FixtureOperation struct {
	operationID, nodeID, certificateID             uuid.UUID
	epoch, sequence                                int64
	scopeDigest                                    [32]byte
	reservationDigest, effectDigest, receiptDigest [32]byte
	requiredLSN                                    string
}
type c12FixtureCertificate struct {
	nodeID, issuanceID, historyOperationID uuid.UUID
	historySequence                        int64
	scopeDigest                            [32]byte
	lineageID                              uuid.UUID
	attemptID                              uuid.UUID
}
type c12FixtureLedger struct {
	activationID   uuid.UUID
	epoch          int64
	certificates   map[uuid.UUID]c12FixtureCertificate
	operations     []c12FixtureOperation
	systemID       uint64
	timeline       int64
	installationID uuid.UUID
	setup          c12FixtureSetup
}

// One fixed setup closure and one in-progress certificate group, never an
// arbitrary registration collection. Only complete groups reach the ledger.
type c12FixtureSetup struct {
	replica                                  bool
	closureStep                              uint8
	activationID                             uuid.UUID
	at                                       time.Time
	nodeID, lineageID, historyID, issuanceID uuid.UUID
	attemptID                                uuid.UUID
	sequence                                 int64
	scope                                    [32]byte
	certificateStep                          uint8
	certificateAt                            time.Time
}

const pitrCertificateSnapshotSQL = `SELECT status,revoke_authority_operation_id,revoke_authority_effect_commitment_jcs,revoke_authority_activation_evidence_jcs,revoke_authority_effect_resolution_jcs FROM nodecontrol.node_certificates WHERE certificate_id=$1`
const pitrOperationSnapshotSQL = `SELECT provider_status,visibility_state,authority_protocol_profile FROM nodecontrol.control_plane_authority_fences WHERE operation_id=$1`
const c12MarkerInsertSQL = `INSERT INTO public.c12_authority_pitr_markers(marker) VALUES($1)`
const pitrAuxCountSQL = `SELECT count(*) FROM nodecontrol.authority_task7_crash_effects WHERE operation_id=$1`
const c12GetNodeControlDatabaseIdentitySQL = `-- name: GetNodeControlDatabaseIdentity :one
SELECT (CASE WHEN system_identifier < 0
               THEN system_identifier::numeric + 18446744073709551616::numeric
               ELSE system_identifier::numeric
        END)::numeric(20,0) AS system_id,
       (CASE WHEN timeline_id < 0
               THEN timeline_id::bigint + 4294967296::bigint
               ELSE timeline_id::bigint
        END)::bigint AS timeline,
       pg_current_wal_insert_lsn()::text AS required_lsn
FROM pg_control_system(), pg_control_checkpoint()
`

const c12GetAuthorityFenceHeadSQL = `-- name: GetAuthorityFenceHead :one
WITH latest_epoch AS (
  SELECT COALESCE(MAX(authority_epoch), 0)::bigint AS authority_epoch
  FROM nodecontrol.control_plane_authority_fences
), epoch_rows AS (
  SELECT f.operation_id, f.effect_kind, f.scope_kind, f.authority_epoch, f.authority_sequence, f.scope_digest, f.provider_reservation_digest, f.effect_digest, f.provider_status, f.provider_receipt_digest, f.db_system_id, f.db_timeline, f.required_lsn, f.abort_reason, f.visibility_state, f.reserved_at, f.effect_bound_at, f.terminal_at, f.authority_protocol_profile, f.abort_claimed_at, f.protocol_activation_id FROM nodecontrol.control_plane_authority_fences f
  JOIN latest_epoch e ON f.authority_epoch = e.authority_epoch
), aggregate_head AS (
  SELECT COUNT(*)::bigint AS record_count,
         COALESCE(MAX(authority_sequence), 0)::bigint AS latest_reserved_sequence,
         COALESCE(MAX(authority_sequence) FILTER (
           WHERE provider_status = 'committed' AND visibility_state = 'active'), 0)::bigint AS latest_committed_sequence,
         COUNT(*) FILTER (WHERE provider_status = 'reserved')::bigint AS pending_count
  FROM epoch_rows
), latest_record AS (
  SELECT provider_reservation_digest FROM epoch_rows ORDER BY authority_sequence DESC LIMIT 1
), latest_committed AS (
  SELECT operation_id, provider_receipt_digest, db_system_id, db_timeline, required_lsn
  FROM epoch_rows
  WHERE provider_status = 'committed' AND visibility_state = 'active'
  ORDER BY authority_sequence DESC LIMIT 1
)
SELECT e.authority_epoch, a.record_count, a.latest_reserved_sequence,
       a.latest_committed_sequence, a.pending_count, r.provider_reservation_digest,
       c.operation_id AS latest_committed_operation_id,
       c.provider_receipt_digest, c.db_system_id, c.db_timeline, c.required_lsn,
       (a.record_count <> a.latest_reserved_sequence) AS has_sequence_gap
FROM latest_epoch e CROSS JOIN aggregate_head a
LEFT JOIN latest_record r ON true
LEFT JOIN latest_committed c ON true
`

const c12ListPendingAuthorityFencesSQL = `-- name: ListPendingAuthorityFences :many
SELECT operation_id, effect_kind, scope_kind, scope_digest, authority_epoch,
       authority_sequence, provider_reservation_digest, effect_digest,
       db_system_id, db_timeline, required_lsn, reserved_at, effect_bound_at
FROM nodecontrol.control_plane_authority_fences
WHERE authority_epoch = $1
  AND provider_status = 'reserved' AND visibility_state = 'fence_pending'
ORDER BY authority_sequence
`

const c12LockAuthorityFenceSQL = `-- name: LockAuthorityFence :one
SELECT operation_id, effect_kind, scope_kind, authority_epoch, authority_sequence, scope_digest, provider_reservation_digest, effect_digest, provider_status, provider_receipt_digest, db_system_id, db_timeline, required_lsn, abort_reason, visibility_state, reserved_at, effect_bound_at, terminal_at, authority_protocol_profile, abort_claimed_at, protocol_activation_id FROM nodecontrol.control_plane_authority_fences WHERE operation_id = $1 FOR UPDATE
`

const c12LockCertificateRevocationOutcomeSQL = `-- name: LockCertificateRevocationOutcome :one
SELECT revoke_authority_effect_commitment_jcs AS commitment_jcs,
       revoke_authority_effect_commitment_digest AS commitment_digest,
       revoke_authority_provider_head_jcs AS provider_head_jcs,
       revoke_authority_provider_head_digest AS provider_head_digest,
       revoke_authority_checkpoint_anchor_jcs AS checkpoint_anchor_jcs,
       revoke_authority_checkpoint_anchor_digest AS checkpoint_anchor_digest,
       revoke_authority_effect_reason AS effect_reason,
       revoke_authority_attestation_expires_at AS attestation_expires_at,
       revoke_authority_activation_deadline AS activation_deadline,
       revoke_authority_expected_provider_identity_digest AS expected_provider_identity_digest,
       revoke_authority_activation_evidence_jcs AS activation_evidence_jcs,
       revoke_authority_activation_evidence_digest AS activation_evidence_digest,
       revoke_authority_effect_resolution_jcs AS effect_resolution_jcs,
       revoke_authority_effect_resolution_digest AS effect_resolution_digest
FROM nodecontrol.node_certificates
WHERE revoke_authority_operation_id = $1 FOR UPDATE
`

const c12GetStoredAuthorityFenceSQL = `-- name: GetStoredAuthorityFence :one
SELECT operation_id, effect_kind, scope_kind, authority_epoch, authority_sequence, scope_digest, provider_reservation_digest, effect_digest, provider_status, provider_receipt_digest, db_system_id, db_timeline, required_lsn, abort_reason, visibility_state, reserved_at, effect_bound_at, terminal_at, authority_protocol_profile, abort_claimed_at, protocol_activation_id FROM nodecontrol.control_plane_authority_fences WHERE operation_id = $1
`

const c12GetAuthorityFenceForUpdateSQL = `-- name: GetAuthorityFenceForUpdate :one
SELECT operation_id, effect_kind, scope_kind, authority_epoch, authority_sequence,
       scope_digest, provider_reservation_digest, effect_digest, provider_status,
       provider_receipt_digest, db_system_id, db_timeline, required_lsn, abort_reason,
       visibility_state, reserved_at, effect_bound_at, terminal_at
FROM nodecontrol.control_plane_authority_fences
WHERE operation_id = $1
FOR UPDATE
`

const c12InsertClaimV1AuthorityFencePendingSQL = `-- name: InsertClaimV1AuthorityFencePending :execrows
INSERT INTO nodecontrol.control_plane_authority_fences
  (operation_id, effect_kind, scope_kind, authority_epoch, authority_sequence, scope_digest,
   provider_reservation_digest, provider_status, visibility_state, reserved_at,
   authority_protocol_profile, protocol_activation_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, 'reserved', 'fence_pending', $8, 'claim_v1', $9)
ON CONFLICT (operation_id) DO NOTHING
`

const c12BindAuthorityFenceEffectSQL = `-- name: BindAuthorityFenceEffect :execrows
UPDATE nodecontrol.control_plane_authority_fences
SET effect_digest = $2, db_system_id = $3, db_timeline = $4,
    required_lsn = $5, effect_bound_at = $6
WHERE operation_id = $1 AND provider_status = 'reserved' AND visibility_state = 'fence_pending'
  AND effect_digest IS NULL
`

const c12ActivateCommittedAuthorityFenceSQL = `-- name: ActivateCommittedAuthorityFence :execrows
UPDATE nodecontrol.control_plane_authority_fences
SET provider_status = 'committed', provider_receipt_digest = $2, visibility_state = 'active', terminal_at = $3
WHERE operation_id = $1 AND provider_status = 'reserved' AND visibility_state = 'fence_pending'
  AND effect_digest = $4 AND db_system_id = $5 AND db_timeline = $6 AND required_lsn = $7
`

const c12AuxResolveSQL = `SELECT effect_kind,scope_kind,scope_digest,effect_digest,effect_state FROM nodecontrol.authority_task7_crash_effects WHERE operation_id=$1 FOR UPDATE`

const c12CertificateResolutionSQL = `SELECT revoke_authority_effect_resolution_jcs IS NOT NULL FROM nodecontrol.node_certificates WHERE revoke_authority_operation_id=$1 FOR UPDATE`

const c12CertificateInputSQL = `SELECT node_id,certificate_id,leaf_der_sha256,revoke_authority_effect_commitment_jcs FROM nodecontrol.node_certificates WHERE revoke_authority_operation_id=$1 FOR UPDATE`

const c12PersistedEffectSQL = `SELECT status,revoke_authority_activation_evidence_jcs,revoke_authority_effect_resolution_jcs FROM nodecontrol.node_certificates WHERE revoke_authority_operation_id=$1 FOR UPDATE`

const c12AuditOutboxCountsSQL = `SELECT (SELECT count(*) FROM nodecontrol.node_operator_audit WHERE authority_operation_id=$1),(SELECT count(*) FROM public.transactional_outbox WHERE event_id=$1)`

const c12OutboxPayloadSQL = `SELECT payload FROM public.transactional_outbox WHERE event_id=$1 AND aggregate_id=$2 AND aggregate_version=$3`

const c12CommitmentReadSQL = `SELECT revoke_authority_effect_commitment_jcs FROM nodecontrol.node_certificates WHERE revoke_authority_operation_id=$1`

const c12CommitmentWriteSQL = `UPDATE nodecontrol.node_certificates SET revoke_authority_operation_id=$2,revoke_authority_epoch=$3,revoke_authority_sequence=$4,revoke_authority_effect_commitment_jcs=$5,revoke_authority_effect_commitment_digest=$6 WHERE certificate_id=$1`

const c12AuxInsertSQL = `INSERT INTO nodecontrol.authority_task7_crash_effects(operation_id,effect_kind,scope_kind,scope_digest,effect_digest,effect_state) VALUES($1,$2,$3,$4,$5,$6)`

const c12CertificateActivateSQL = `UPDATE nodecontrol.node_certificates SET
 status='revoked',revoked_at=$2,revoke_reason='scheduled',updated_at=$2,
 revoke_authority_provider_head_jcs=$3,revoke_authority_provider_head_digest=$4,
 revoke_authority_effect_reason=$5,revoke_authority_attestation_expires_at=$6,revoke_authority_activation_deadline=$7,
 revoke_authority_expected_provider_identity_digest=$8,revoke_authority_activation_evidence_jcs=$9,revoke_authority_activation_evidence_digest=$10,
 revoke_authority_effect_resolution_jcs=$11,revoke_authority_effect_resolution_digest=$12
 WHERE revoke_authority_operation_id=$1`

const c12AuditInsertSQL = `INSERT INTO nodecontrol.node_operator_audit(audit_id,command_id,authority_operation_id,authority_epoch,authority_sequence,operator_id,credential_digest,role,action,target_kind,target_id,reason,result,occurred_at,retention_until)
 VALUES($1,$1,$1,$2,$3,'task9-fixture',$4,'node_security_admin','disable_node','node',$5,'administrative_disable','accepted',$6,$7)`

const c12OutboxInsertSQL = `INSERT INTO public.transactional_outbox(event_id,event_type,aggregate_type,aggregate_id,aggregate_version,idempotency_key,payload,occurred_at,available_at)
 VALUES($1,'talenro.node.certificate_revoked','node',$2,$3,$4,$5,$6,$6)`

const c12LatchSQL = `SELECT installation_id FROM nodecontrol.control_plane_authority_protocol_migration_latches WHERE singleton_key ORDER BY installation_id LIMIT 1`

const c12ReplicaSQL = `SET LOCAL session_replication_role=replica`

const c12UpgradeIntentSQL = `
INSERT INTO nodecontrol.control_plane_authority_protocol_upgrade_intents(
 intent_id,installation_id,installation_kind,activation_id,request_nonce,incarnation_registration_id,
 provider_absence_proof_id,observed_deployment_id,database_identity_digest,local_runtime_isolation_digest,
 classification_state,created_at,canonical_evidence_bundle_jcs,canonical_body_jcs,body_digest)
VALUES($1,$2,'production',$3,decode(repeat('91',32),'hex'),$4,$5,$6,decode(repeat('92',32),'hex'),
 decode(repeat('93',32),'hex'),'pending',$7,decode('7b7d','hex'),decode('7b7d','hex'),decode(repeat('a1',32),'hex'))`

const c12RuntimeRegistrationSQL = `
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
 'registered_pending_genesis',0,decode('01','hex'),$4,decode('7b7d','hex'),decode('7b7d','hex'),decode(repeat('a2',32),'hex'))`

const c12UpgradeAttemptSQL = `
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
 decode(repeat('b0',32),'hex'),1,$8,decode('7b7d','hex'),decode('7b7d','hex'),decode(repeat('a3',32),'hex'))`

const c12ProtocolActivationSQL = `
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
	 $3,decode('7b7d','hex'),decode('7b7d','hex'),decode(repeat('a4',32),'hex'))`

const c12ActivationCompletionSQL = `
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
 decode(repeat('a5',32),'hex'))`

const c12ActivationReleaseSQL = `
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
 decode(repeat('a6',32),'hex'))`

const c12OriginSQL = `SET LOCAL session_replication_role=origin`

const c12NodePopSQL = `INSERT INTO nodecontrol.node_pops(pop_code,iso_country,region,operator_state,created_at,updated_at)
 VALUES('coordinator-crash','US','fixture','enabled',$1,$1) ON CONFLICT DO NOTHING`

const c12NodeInventorySQL = `INSERT INTO nodecontrol.node_inventory(node_id,pop_code,operator_state,security_state,identity_state,identity_epoch,lineage_id,created_at,updated_at)
 VALUES($1,'coordinator-crash','enabled','normal','active',1,$2,$3,$3)`

const c12LegacyFenceSQL = `INSERT INTO nodecontrol.control_plane_authority_fences(operation_id,effect_kind,scope_kind,authority_epoch,authority_sequence,scope_digest,provider_reservation_digest,effect_digest,provider_status,provider_receipt_digest,db_system_id,db_timeline,required_lsn,visibility_state,reserved_at,effect_bound_at,terminal_at,authority_protocol_profile)
 VALUES($1,'certificate_activate','node',1,$2,$3,$3,$3,'committed',$3,$4,$5,$6::pg_lsn,'active',$7,$7,$7,'legacy_v6')`

const c12IssuanceSQL = `INSERT INTO nodecontrol.node_certificate_issuances(
 issuance_id,authority_operation_id,authority_epoch,authority_sequence,node_id,attempt_id,issuance_kind,identity_epoch,lineage_id,
 issuer_id,csr_sha256,public_key_sha256,template_sha256,request_digest,status,serial_bytes,leaf_der,leaf_der_sha256,
 chain_der,chain_der_sha256,not_before,not_after,created_at,updated_at,terminal_at,retention_until)
 VALUES($1,$2,1,$3,$4,$5,'initial',1,$6,'fixture-ca',$7,$7,$7,$7,'active',decode('01','hex'),decode('02','hex'),$7,
 decode('03','hex'),$7,$8,$9,$8,$8,$8,$10)`

const c12CertificateSeedSQL = `INSERT INTO nodecontrol.node_certificates(
 certificate_id,issuance_id,authority_operation_id,authority_epoch,authority_sequence,node_id,identity_epoch,lineage_id,
 issuer_id,serial_bytes,leaf_der,leaf_der_sha256,public_key_sha256,chain_der_sha256,valid_from,valid_until,status,created_at,updated_at,retention_until)
 VALUES($1,$2,$3,1,$4,$5,1,$6,'fixture-ca',decode('01','hex'),decode('02','hex'),$7,$7,$7,$8,$9,'active',$8,$8,$10)`

const c12AuxDDLSQL = `CREATE TABLE nodecontrol.authority_task7_crash_effects(
 operation_id uuid NOT NULL,effect_kind text NOT NULL,scope_kind text NOT NULL,
 scope_digest bytea NOT NULL CHECK(octet_length(scope_digest)=32),
 effect_digest bytea NOT NULL CHECK(octet_length(effect_digest)=32),effect_state text NOT NULL)`

func c12AuthorizeSQL(mode c12SQLMode, call c12SQLCall, query string, args []any, f *c12FixtureLedger) error {
	rule, ok := c12AuthoritySQLRules[query]
	if !ok || rule.call != call || rule.modes&uint8(mode) == 0 || f == nil ||
		(mode != c12SQLPrimary && mode != c12SQLSetup && mode != c12SQLPrimary|c12SQLSetup && mode != c12SQLCandidate) || len(args) != rule.arity {
		return C12PITRInvalidHandle
	}
	if call == c12SQLExec && rule.modes&uint8(c12SQLSetup) == 0 && f.setup.replica {
		return C12PITRInvalidHandle
	}
	if !rule.validate(args, f) {
		return C12PITRInvalidHandle
	}
	return nil
}

type c12SQLRule struct {
	call     c12SQLCall
	modes    uint8
	arity    int
	validate func([]any, *c12FixtureLedger) bool
}

var c12AuthoritySQLRules = func() map[string]c12SQLRule {
	rules := make(map[string]c12SQLRule)
	add := func(sql string, call c12SQLCall, modes c12SQLMode, arity int) {
		rules[sql] = c12SQLRule{call, uint8(modes), arity, func(a []any, f *c12FixtureLedger) bool { return c12ValidateStatement(sql, a, f) }}
	}
	both := c12SQLPrimary | c12SQLCandidate
	add(c12GetNodeControlDatabaseIdentitySQL, c12SQLQueryRow, both, 0)
	add(c12GetAuthorityFenceHeadSQL, c12SQLQueryRow, both, 0)
	add(c12ListPendingAuthorityFencesSQL, c12SQLQuery, both, 1)
	add(c12LockAuthorityFenceSQL, c12SQLQueryRow, both, 1)
	add(c12LockCertificateRevocationOutcomeSQL, c12SQLQueryRow, both, 1)
	add(c12GetStoredAuthorityFenceSQL, c12SQLQueryRow, c12SQLPrimary, 1)
	add(c12GetAuthorityFenceForUpdateSQL, c12SQLQueryRow, c12SQLPrimary, 1)
	add(c12InsertClaimV1AuthorityFencePendingSQL, c12SQLExec, c12SQLPrimary, 9)
	add(c12BindAuthorityFenceEffectSQL, c12SQLExec, c12SQLPrimary, 6)
	add(c12ActivateCommittedAuthorityFenceSQL, c12SQLExec, c12SQLPrimary, 7)
	add(c12AuxResolveSQL, c12SQLQuery, both, 1)
	for _, sql := range []string{c12CertificateResolutionSQL, c12CertificateInputSQL, c12PersistedEffectSQL, c12AuditOutboxCountsSQL, pitrOperationSnapshotSQL, pitrAuxCountSQL, pitrCertificateSnapshotSQL} {
		add(sql, c12SQLQueryRow, both, 1)
	}
	add(c12OutboxPayloadSQL, c12SQLQueryRow, both, 3)
	add(c12CommitmentReadSQL, c12SQLQueryRow, c12SQLPrimary, 1)
	for sql, n := range map[string]int{c12CommitmentWriteSQL: 6, c12AuxInsertSQL: 6, c12CertificateActivateSQL: 12, c12AuditInsertSQL: 7, c12OutboxInsertSQL: 6} {
		add(sql, c12SQLExec, c12SQLPrimary, n)
	}
	add(c12LatchSQL, c12SQLQueryRow, c12SQLSetup, 0)
	for sql, n := range map[string]int{c12ReplicaSQL: 0, c12OriginSQL: 0, c12UpgradeIntentSQL: 7, c12RuntimeRegistrationSQL: 4, c12UpgradeAttemptSQL: 8, c12ProtocolActivationSQL: 3, c12ActivationCompletionSQL: 3, c12ActivationReleaseSQL: 4, c12NodePopSQL: 1, c12NodeInventorySQL: 3, c12LegacyFenceSQL: 7, c12IssuanceSQL: 10, c12CertificateSeedSQL: 10} {
		add(sql, c12SQLExec, c12SQLSetup, n)
	}
	return rules
}()

func (f *c12FixtureLedger) clone() *c12FixtureLedger {
	if f == nil {
		return nil
	}
	out := *f
	out.certificates = make(map[uuid.UUID]c12FixtureCertificate, len(f.certificates))
	for k, v := range f.certificates {
		out.certificates[k] = v
	}
	out.operations = append([]c12FixtureOperation(nil), f.operations...)
	return &out
}
func (f *c12FixtureLedger) operation(value any) *c12FixtureOperation {
	id, ok := value.(uuid.UUID)
	if !ok || id == uuid.Nil {
		return nil
	}
	for i := range f.operations {
		if f.operations[i].operationID == id {
			return &f.operations[i]
		}
	}
	return nil
}
func c12UUID(value any) (uuid.UUID, bool) { v, ok := value.(uuid.UUID); return v, ok && v != uuid.Nil }
func c12Int(value any, want int64) bool   { v, ok := value.(int64); return ok && v == want }
func c12Bytes(value any, n int) bool {
	v, ok := value.([]byte)
	return ok && len(v) == n && !bytes.Equal(v, make([]byte, n))
}
func c12Digest(value any, want [32]byte) bool {
	v, ok := value.([]byte)
	return ok && len(v) == 32 && bytes.Equal(v, want[:])
}
func c12Body(value any, limit int) bool {
	v, ok := value.([]byte)
	return ok && len(v) > 0 && len(v) <= limit
}
func c12Time(value any) bool {
	v, ok := value.(time.Time)
	return ok && !v.IsZero() && v.Year() >= 1 && v.Year() <= 9999 && v.Location() == time.UTC
}
func c12Timestamp(value any) bool {
	v, ok := value.(pgtype.Timestamptz)
	return ok && v.Valid && v.InfinityModifier == pgtype.Finite && c12Time(v.Time) && v.Time.Nanosecond()%1000 == 0
}
func c12String(value any, want, name string) bool {
	if v, ok := value.(string); ok {
		return v == want
	}
	typ := reflect.TypeOf(value)
	return name != "" && typ != nil && typ.Kind() == reflect.String &&
		typ.PkgPath() == "talenro.local/platform/internal/nodecontrol/authority" && typ.Name() == name &&
		reflect.ValueOf(value).String() == want
}
func c12LSN(value any) bool {
	v, ok := value.(string)
	if !ok || len(v) > 17 || !c12AuthorityPITRLSNPattern.MatchString(v) {
		return false
	}
	n, ok := c12AuthorityPITRLSNValue(v)
	return ok && n > 0
}
func c12Physical(a, b, c any, f *c12FixtureLedger) bool {
	n, ok := a.(pgtype.Numeric)
	t, tok := b.(pgtype.Int8)
	return ok && n.Valid && !n.NaN && n.InfinityModifier == pgtype.Finite && n.Exp == 0 && n.Int != nil && n.Int.Sign() > 0 &&
		n.Int.IsUint64() && f.systemID > 0 && n.Int.Uint64() == f.systemID && tok && t.Valid && t.Int64 == f.timeline &&
		f.timeline > 0 && c12LSN(c)
}
func c12Scope(node uuid.UUID) [32]byte {
	return sha256.Sum256(append([]byte("TALENRO-NODE-AUTHORITY-SCOPE-V1\x00"), node[:]...))
}
func c12SetupID(label string) uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte("task8-proof-"+label+":79"))
}

// Copy before authorization, and never invoke a caller-controlled codec,
// Stringer, Valuer, or QueryRewriter. Mutable values have bounded private copies.
func c12CopySQLArguments(args []any) ([]any, error) {
	if len(args) > 12 {
		return nil, C12PITRInvalidHandle
	}
	out := make([]any, len(args))
	for i, arg := range args {
		switch v := arg.(type) {
		case uuid.UUID, uuid.NullUUID, int64, time.Time, pgtype.Timestamptz, pgtype.Int8:
			out[i] = v
		case string:
			if len(v) > 1<<20 {
				return nil, C12PITRInvalidHandle
			}
			out[i] = strings.Clone(v)
		case []byte:
			if len(v) > 1<<20 {
				return nil, C12PITRInvalidHandle
			}
			out[i] = bytes.Clone(v)
		case pgtype.Numeric:
			if v.Int != nil {
				if v.Int.BitLen() > 64 {
					return nil, C12PITRInvalidHandle
				}
				v.Int = new(big.Int).Set(v.Int)
			}
			out[i] = v
		default:
			typ := reflect.TypeOf(arg)
			if typ == nil || typ.Kind() != reflect.String || typ.PkgPath() != "talenro.local/platform/internal/nodecontrol/authority" {
				return nil, C12PITRInvalidHandle
			}
			switch typ.Name() {
			case "EffectKind", "ScopeKind", "EffectState", "AuthorityEffectReason":
			default:
				return nil, C12PITRInvalidHandle
			}
			out[i] = arg // Named string values are immutable; retain identity for validation.
		}
	}
	return out, nil
}

func c12ValidateStatement(sql string, a []any, f *c12FixtureLedger) bool {
	switch sql {
	case c12GetNodeControlDatabaseIdentitySQL, c12GetAuthorityFenceHeadSQL:
		return true
	case c12ListPendingAuthorityFencesSQL:
		return c12Int(a[0], 31)
	case c12LockCertificateRevocationOutcomeSQL:
		v, ok := a[0].(uuid.NullUUID)
		return ok && v.Valid && f.operation(v.UUID) != nil
	case pitrCertificateSnapshotSQL:
		id, ok := c12UUID(a[0])
		_, found := f.certificates[id]
		return ok && found
	case c12LatchSQL:
		return f.activationID == uuid.Nil && f.setup.closureStep == 0
	case c12ReplicaSQL:
		return !f.setup.replica
	case c12OriginSQL:
		return f.setup.replica
	case c12UpgradeIntentSQL, c12RuntimeRegistrationSQL, c12UpgradeAttemptSQL, c12ProtocolActivationSQL, c12ActivationCompletionSQL, c12ActivationReleaseSQL, c12NodePopSQL, c12NodeInventorySQL, c12LegacyFenceSQL, c12IssuanceSQL, c12CertificateSeedSQL:
		return c12ValidateSetup(sql, a, f)
	case c12InsertClaimV1AuthorityFencePendingSQL:
		id, ok := c12UUID(a[0])
		activation, aok := a[8].(uuid.NullUUID)
		cert, cok := f.certificates[uuid.NewSHA1(id, []byte("certificate"))]
		seq, sok := a[4].(int64)
		if !ok || !cok || !sok || seq < 1 || seq > 3 || f.epoch != 31 || !c12Int(a[3], 31) || !c12String(a[1], "certificate_revoke", "") || !c12String(a[2], "node", "") || !c12Digest(a[5], cert.scopeDigest) || !c12Bytes(a[6], 32) || !c12Timestamp(a[7]) || !aok || !activation.Valid || activation.UUID == uuid.Nil || activation.UUID != f.activationID {
			return false
		}
		if cert.historySequence != seq || cert.issuanceID != uuid.NewSHA1(id, []byte("issuance")) || cert.historyOperationID != uuid.NewSHA1(id, []byte("issuance-operation")) || cert.lineageID != uuid.NewSHA1(id, []byte("lineage")) || cert.attemptID != uuid.NewSHA1(id, []byte("attempt")) {
			return false
		}
		if op := f.operation(id); op != nil {
			return seq == op.sequence && c12Digest(a[6], op.reservationDigest)
		}
		return len(f.operations) < 3 && seq == int64(len(f.operations)+1)
	}
	op := f.operation(a[0])
	if sql == c12CommitmentWriteSQL {
		op = f.operation(a[1])
	}
	if op == nil {
		return false
	}
	switch sql {
	case c12LockAuthorityFenceSQL, c12GetStoredAuthorityFenceSQL, c12GetAuthorityFenceForUpdateSQL, c12AuxResolveSQL, c12CertificateResolutionSQL, c12CertificateInputSQL, c12PersistedEffectSQL, c12AuditOutboxCountsSQL, c12CommitmentReadSQL, pitrOperationSnapshotSQL, pitrAuxCountSQL:
		return true
	case c12OutboxPayloadSQL:
		return a[1] == op.nodeID && c12Int(a[2], op.sequence)
	case c12CommitmentWriteSQL:
		return a[0] == op.certificateID && op.certificateID == uuid.NewSHA1(op.operationID, []byte("certificate")) &&
			c12Int(a[2], 31) && c12Int(a[3], op.sequence) && c12Body(a[4], 4096) && c12Bytes(a[5], 32)
	case c12AuxInsertSQL:
		return c12String(a[1], "certificate_revoke", "EffectKind") && c12String(a[2], "node", "ScopeKind") &&
			c12Digest(a[3], op.scopeDigest) && c12Digest(a[4], op.effectDigest) && op.effectDigest != [32]byte{} &&
			c12String(a[5], "committed", "EffectState")
	case c12BindAuthorityFenceEffectSQL:
		return op.effectDigest != [32]byte{} && c12Digest(a[1], op.effectDigest) && c12Physical(a[2], a[3], a[4], f) && c12Timestamp(a[5]) && (op.requiredLSN == "" || a[4] == op.requiredLSN)
	case c12ActivateCommittedAuthorityFenceSQL:
		return c12Bytes(a[1], 32) && c12Timestamp(a[2]) && op.effectDigest != [32]byte{} && c12Digest(a[3], op.effectDigest) &&
			c12Physical(a[4], a[5], a[6], f) && op.requiredLSN != "" && a[6] == op.requiredLSN
	case c12CertificateActivateSQL:
		return c12Time(a[1]) && c12Body(a[2], 4096) && c12Bytes(a[3], 32) &&
			c12String(a[4], "none", "AuthorityEffectReason") && c12Time(a[5]) && c12Time(a[6]) &&
			a[5].(time.Time).After(a[1].(time.Time)) && a[6].(time.Time).After(a[1].(time.Time)) && c12Bytes(a[7], 32) &&
			c12Body(a[8], 4096) && c12Bytes(a[9], 32) && c12Body(a[10], 4096) &&
			c12Bytes(a[11], 32)
	case c12AuditInsertSQL:
		return c12Int(a[1], 31) && c12Int(a[2], op.sequence) && c12Digest(a[3], op.receiptDigest) &&
			op.receiptDigest != [32]byte{} && c12String(a[4], op.nodeID.String(), "") && c12Time(a[5]) && c12Time(a[6]) &&
			a[6].(time.Time).Equal(a[5].(time.Time).Add(181*24*time.Hour))
	case c12OutboxInsertSQL:
		return a[1] == op.nodeID && c12Int(a[2], op.sequence) && c12String(a[3], op.operationID.String(), "") && c12Body(a[4], 262144) && c12Time(a[5])
	}
	return false
}

func c12ValidateSetup(sql string, a []any, f *c12FixtureLedger) bool {
	s := &f.setup
	switch sql {
	case c12UpgradeIntentSQL:
		return s.replica && s.closureStep == 0 && f.activationID == uuid.Nil && f.installationID != uuid.Nil &&
			a[0] == c12SetupID("intent") && a[1] == f.installationID &&
			a[2] == uuid.MustParse("79000000-0000-4000-8000-000000000001") && a[3] == c12SetupID("incarnation") &&
			a[4] == c12SetupID("absence") && a[5] == c12SetupID("deployment") && c12Time(a[6])
	case c12RuntimeRegistrationSQL:
		return s.replica && s.closureStep == 1 && a[0] == c12SetupID("registration") && a[1] == s.activationID && a[2] == c12SetupID("runtime") && a[3] == s.at
	case c12UpgradeAttemptSQL:
		return s.replica && s.closureStep == 2 && a[0] == c12SetupID("attempt") && a[1] == s.activationID &&
			a[2] == c12SetupID("preparation") && a[3] == c12SetupID("completion") && a[4] == c12SetupID("release") &&
			a[5] == c12SetupID("open") && a[6] == c12SetupID("deployment") && a[7] == s.at
	case c12ProtocolActivationSQL:
		return s.replica && s.closureStep == 3 && a[0] == s.activationID && a[1] == c12SetupID("deployment") && a[2] == s.at
	case c12ActivationCompletionSQL:
		return s.replica && s.closureStep == 4 && a[0] == c12SetupID("completion") && a[1] == s.activationID && a[2] == s.at
	case c12ActivationReleaseSQL:
		return s.replica && s.closureStep == 5 && a[0] == c12SetupID("release") && a[1] == c12SetupID("open") && a[2] == s.activationID && a[3] == s.at
	case c12NodePopSQL:
		return s.replica && s.certificateStep == 0 && len(f.certificates) < 3 && c12Time(a[0])
	case c12NodeInventorySQL:
		node, ok := c12UUID(a[0])
		lineage, lok := c12UUID(a[1])
		return s.replica && s.certificateStep == 1 && ok && lok && node != lineage && !f.usedSetupIdentity(node) && !f.usedSetupIdentity(lineage) && a[2] == s.certificateAt
	case c12LegacyFenceSQL:
		id, ok := c12UUID(a[0])
		return s.replica && s.certificateStep == 2 && ok && !f.usedSetupIdentity(id) && id != s.nodeID && id != s.lineageID &&
			c12Int(a[1], int64(len(f.certificates)+1)) && c12Digest(a[2], c12Scope(s.nodeID)) && f.systemID > 0 &&
			c12String(a[3], strconv.FormatUint(f.systemID, 10), "") && f.timeline > 0 && c12Int(a[4], f.timeline) && c12LSN(a[5]) &&
			a[6] == s.certificateAt
	case c12IssuanceSQL:
		id, ok := c12UUID(a[0])
		attemptID, attempt := c12UUID(a[4])
		return s.replica && s.certificateStep == 3 && ok && attempt && id != attemptID && !f.usedSetupIdentity(id) &&
			!f.usedSetupIdentity(attemptID) && id != s.nodeID && id != s.lineageID && id != s.historyID && attemptID != s.nodeID &&
			attemptID != s.lineageID && attemptID != s.historyID && a[1] == s.historyID && c12Int(a[2], s.sequence) &&
			a[3] == s.nodeID && a[5] == s.lineageID && c12Digest(a[6], s.scope) && a[7] == s.certificateAt &&
			a[8] == s.certificateAt.Add(time.Hour) && c12Time(a[9]) && a[9].(time.Time).After(a[8].(time.Time))
	case c12CertificateSeedSQL:
		id, ok := c12UUID(a[0])
		return !s.replica && s.certificateStep == 4 && ok && !f.usedSetupIdentity(id) && id != s.nodeID && id != s.lineageID &&
			id != s.historyID && id != s.issuanceID && id != s.attemptID && a[1] == s.issuanceID && a[2] == s.historyID &&
			c12Int(a[3], s.sequence) && a[4] == s.nodeID && a[5] == s.lineageID && c12Digest(a[6], s.scope) &&
			a[7] == s.certificateAt && a[8] == s.certificateAt.Add(time.Hour) && c12Time(a[9]) &&
			a[9].(time.Time).After(a[8].(time.Time))
	}
	return false
}

func (f *c12FixtureLedger) usedSetupIdentity(id uuid.UUID) bool {
	for certificateID, c := range f.certificates {
		if id == certificateID || id == c.nodeID || id == c.lineageID || id == c.historyOperationID || id == c.issuanceID || id == c.attemptID {
			return true
		}
	}
	return false
}

// Called only after a successful execution; INSERT ... DO NOTHING with zero
// affected rows cannot mint a new identity. The transaction owns this copy.
func (f *c12FixtureLedger) applied(sql string, a []any, affected int64) {
	s := &f.setup
	if sql == c12ReplicaSQL {
		s.replica = true
		return
	}
	if sql == c12OriginSQL {
		s.replica = false
		return
	}
	if sql == c12NodePopSQL && affected == 0 {
		s.certificateAt = a[0].(time.Time)
		s.certificateStep = 1
		return
	}
	if affected != 1 {
		return
	}
	switch sql {
	case c12UpgradeIntentSQL:
		s.activationID = a[2].(uuid.UUID)
		s.at = a[6].(time.Time)
		s.closureStep = 1
	case c12RuntimeRegistrationSQL, c12UpgradeAttemptSQL, c12ProtocolActivationSQL, c12ActivationCompletionSQL:
		s.closureStep++
	case c12ActivationReleaseSQL:
		s.closureStep = 6
		f.activationID = s.activationID
	case c12NodePopSQL:
		s.certificateAt = a[0].(time.Time)
		s.certificateStep = 1
	case c12NodeInventorySQL:
		s.nodeID = a[0].(uuid.UUID)
		s.lineageID = a[1].(uuid.UUID)
		s.certificateStep = 2
	case c12LegacyFenceSQL:
		s.historyID = a[0].(uuid.UUID)
		s.sequence = a[1].(int64)
		copy(s.scope[:], a[2].([]byte))
		s.certificateStep = 3
	case c12IssuanceSQL:
		s.issuanceID = a[0].(uuid.UUID)
		s.attemptID = a[4].(uuid.UUID)
		s.certificateStep = 4
	case c12CertificateSeedSQL:
		f.certificates[a[0].(uuid.UUID)] = c12FixtureCertificate{nodeID: s.nodeID, issuanceID: s.issuanceID, historyOperationID: s.historyID, historySequence: s.sequence, scopeDigest: s.scope, lineageID: s.lineageID, attemptID: s.attemptID}
		s.certificateStep = 0
	case c12InsertClaimV1AuthorityFencePendingSQL:
		id := a[0].(uuid.UUID)
		if f.operation(id) != nil {
			return
		}
		cid := uuid.NewSHA1(id, []byte("certificate"))
		c := f.certificates[cid]
		op := c12FixtureOperation{operationID: id, nodeID: c.nodeID, certificateID: cid, epoch: 31, sequence: a[4].(int64), scopeDigest: c.scopeDigest}
		copy(op.reservationDigest[:], a[6].([]byte))
		f.operations = append(f.operations, op)
	case c12CommitmentWriteSQL:
		copy(f.operation(a[1]).effectDigest[:], a[5].([]byte))
	case c12BindAuthorityFenceEffectSQL:
		f.operation(a[0]).requiredLSN = a[4].(string)
	case c12ActivateCommittedAuthorityFenceSQL:
		copy(f.operation(a[0]).receiptDigest[:], a[1].([]byte))
	}
}

func (lease *c12AccessLease) prepareSQL(tx *c12AuthorityTx, call c12SQLCall, sql string, args []any) ([]any, error) {
	state := lease.owner
	state.fixtureMu.Lock()
	defer state.fixtureMu.Unlock()
	if state.fixture == nil {
		return args, nil
	} // Private Task1 lifecycle test seam only.
	copied, err := c12CopySQLArguments(args)
	if err != nil {
		return nil, err
	}
	mode := c12SQLPrimary
	if lease.binding != nil {
		mode = c12SQLCandidate
	} else if !state.setupClosed && state.phase.Load() == 1 {
		mode |= c12SQLSetup
	}
	f := state.fixture
	if tx != nil {
		f = tx.fixture
	}
	if call == c12SQLExec && (tx == nil || state.writesUncertain) {
		return nil, C12PITRInvalidHandle
	}
	if sql == c12LatchSQL && tx == nil {
		return nil, C12PITRInvalidHandle
	}
	if err := c12AuthorizeSQL(mode, call, sql, copied, f); err != nil {
		return nil, err
	}
	for i, arg := range copied {
		typ := reflect.TypeOf(arg)
		if typ != nil && typ.Kind() == reflect.String && typ.PkgPath() != "" {
			copied[i] = reflect.ValueOf(arg).String()
		}
	}
	return copied, nil
}

func (transaction *c12AuthorityTx) mergeFixtureAfterCommit(err error) {
	if transaction.fixture == nil {
		return
	}
	state := transaction.lease.owner
	state.fixtureMu.Lock()
	defer state.fixtureMu.Unlock()
	if err != nil {
		state.writesUncertain = true
		return
	}
	if !transaction.fixtureDirty {
		return
	}
	// Incomplete setup may exist physically, but never grants fixture authority.
	// A successful partial setup commit irreversibly closes further write access.
	s := transaction.fixture.setup
	if s.replica || (s.closureStep > 0 && s.closureStep < 6) || s.certificateStep != 0 {
		state.writesUncertain = true
		return
	}
	state.fixture = transaction.fixture.clone()
}
