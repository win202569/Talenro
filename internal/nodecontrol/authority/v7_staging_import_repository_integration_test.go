//go:build integration

package authority_test

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
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gowebpki/jcs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	migrations "talenro.local/platform/db/migrations"
	authority "talenro.local/platform/internal/nodecontrol/authority"
	"talenro.local/platform/internal/nodecontrol/contracts"
)

const (
	task8FixedCapabilityDigestHex = "631af925575d0c2a71a220b8b2245c2d7dc77c42a26598e5a2867ccf938e8178"
	task8FixedManifestDigestHex   = "e2f9a5cc76ffb23e22f6abe3e783bf675156555f0ef257df347398884ce5b285"
	task8FixedLeaseDigestHex      = "78fcc895ee976ad4854f3ac5a547435758e991823953566e6adb845714ef00cc"
	task8FixedProviderHeadHex     = "aa4e8bd2310938828695fd39dca6db9fb5909d4514cf29a7a82a5d6861b79910"
	task8FixedProofDigestHex      = "1e7c5d3674a2aa1adb81091f05e9b654c0610e7317462541bec074b0a739b9d3"

	// These four literal source-member registries belong only to this Task-1
	// fixture. They freeze flatten/dedupe/reconstruction behavior without
	// claiming to define or implement the future production B01 registries.
	task8FixedCapabilitySourceBundle = `{"evidence":[{"body_digest":"1e7c5d3674a2aa1adb81091f05e9b654c0610e7317462541bec074b0a739b9d3","canonical_body_or_null":null,"canonical_envelope_or_null":{"body":{"database_identity_digest":"5454545454545454545454545454545454545454545454545454545454545454","database_point":"0/16B6C50"},"body_digest":"1e7c5d3674a2aa1adb81091f05e9b654c0610e7317462541bec074b0a739b9d3","schema":"database-incarnation-proof.v1","signature":"bUnQkvSZ58TCn_1OKSlMzO1ZjmLrw0tZrOk2KLmMaLBvRBDcXVkOmmI-Gk5hup8g6chtcILKr21smj7odSixCg","signature_algorithm":"ed25519","signature_policy_version":"1","signer_key_id":"56475aa75463474c0285df5dbf2bcab73da651358839e9b77481b2eab107708c","signer_role":"claim_v1_provider","trust_root_digest":"a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1"},"evidence_kind":"external_signed_envelope","schema":"database-incarnation-proof.v1"},{"body_digest":"e2f9a5cc76ffb23e22f6abe3e783bf675156555f0ef257df347398884ce5b285","canonical_body_or_null":null,"canonical_envelope_or_null":{"body":{"manifest_id":"52000000-0000-4000-8000-000000000001","object_count":"4","single_use_apply_id":"52000000-0000-4000-8000-000000000002"},"body_digest":"e2f9a5cc76ffb23e22f6abe3e783bf675156555f0ef257df347398884ce5b285","schema":"fresh-restore-import-manifest.v1","signature":"yEqWbcsqdNQVOywki-qEZdJ6KrrxU482EzXuMNyNK0QvQJ3WnBFzWysLJQCwP1XNCtx1EajtpRDk8hMgvFCBBg","signature_algorithm":"ed25519","signature_policy_version":"1","signer_key_id":"56475aa75463474c0285df5dbf2bcab73da651358839e9b77481b2eab107708c","signer_role":"fresh_restore_export_operator","trust_root_digest":"a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1"},"evidence_kind":"external_signed_envelope","schema":"fresh-restore-import-manifest.v1"}],"evidence_count":"2","message_body_digest":"631af925575d0c2a71a220b8b2245c2d7dc77c42a26598e5a2867ccf938e8178","message_schema":"staging-import-capability.v1"}`
	task8FixedManifestSourceBundle   = `{"evidence":[{"body_digest":"46251d95f89fecf930e877e0321822f509129d20e0093df2949b93ab23d42f9f","canonical_body_or_null":{"normalized_catalog_digest":"28ba9545623b38b16a33fa2ee72daf367aa440c06bc1144300fa8feea2f6739b","object_count":"2","objects":[{"canonical_key":"54000000-0000-4000-8000-000000000001","normalized_payload":{"active_pointer_set":[],"authority_anchor_set":[],"health_state":"unknown","identity_epoch":"0","identity_state":"unauthorized","inventory_version":"1","next_desired_generation":"1","next_recovery_generation":"1","node_id":"54000000-0000-4000-8000-000000000001","operator_state":"disabled","pending_operator_transition_or_null":null,"pop_code":"fixture-pop","resume_operator_state_or_null":null,"security_state":"quarantined","security_version":"1"},"object_type":"node_reconstruction_seed"},{"canonical_key":"fixture-pop","normalized_payload":{"iso_country":"US","operator_state":"disabled","pop_code":"fixture-pop","region":"fixture-region","version":"1"},"object_type":"pop"}],"projection_version":"1","target_activation_id":"54000000-0000-4000-8000-000000000002","target_database_identity_digest":"787f2747e2d7b0b194f15236f69d53d37a8248ef387ae9f11c8ae6599e5c24bb","target_deployment_id":"54000000-0000-4000-8000-000000000003"},"canonical_envelope_or_null":null,"evidence_kind":"database_immutable_body","schema":"fresh-import-topology-projection.v1"}],"evidence_count":"1","message_body_digest":"e2f9a5cc76ffb23e22f6abe3e783bf675156555f0ef257df347398884ce5b285","message_schema":"fresh-restore-import-manifest.v1"}`
	task8FixedLeaseSourceBundle      = `{"evidence":[{"body_digest":"631af925575d0c2a71a220b8b2245c2d7dc77c42a26598e5a2867ccf938e8178","canonical_body_or_null":null,"canonical_envelope_or_null":{"body":{"capability_id":"51000000-0000-4000-8000-000000000001","issued_at":"2026-08-29T17:00:00.123456Z","transaction_nonce":"0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"},"body_digest":"631af925575d0c2a71a220b8b2245c2d7dc77c42a26598e5a2867ccf938e8178","schema":"staging-import-capability.v1","signature":"iWtuZU-M_v65B4-7OfhBGNpDBvRiVgyUkJaoLXbLoA8JrO6ea9oFkWqP-lC-OxaPS389J27IgJ0oBZxe2Sl0DQ","signature_algorithm":"ed25519","signature_policy_version":"1","signer_key_id":"56475aa75463474c0285df5dbf2bcab73da651358839e9b77481b2eab107708c","signer_role":"fresh_restore_staging_import_authorizer","trust_root_digest":"a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1"},"evidence_kind":"external_signed_envelope","schema":"staging-import-capability.v1"},{"body_digest":"aa4e8bd2310938828695fd39dca6db9fb5909d4514cf29a7a82a5d6861b79910","canonical_body_or_null":null,"canonical_envelope_or_null":{"body":{"provider_control_sequence":"7","provider_phase":"fresh_v7_staging_closed"},"body_digest":"aa4e8bd2310938828695fd39dca6db9fb5909d4514cf29a7a82a5d6861b79910","schema":"claim-v1-provider-head.v1","signature":"k2Se04MvINfYJHP0Ck9lh1mhOnLXjQp5Qc9ZUCCdl_gItIwv1pjPvkQTMkNYmny8P0CPFUFcWuXrb1d1GH8yBw","signature_algorithm":"ed25519","signature_policy_version":"1","signer_key_id":"56475aa75463474c0285df5dbf2bcab73da651358839e9b77481b2eab107708c","signer_role":"claim_v1_provider","trust_root_digest":"a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1"},"evidence_kind":"external_signed_envelope","schema":"claim-v1-provider-head.v1"}],"evidence_count":"2","message_body_digest":"78fcc895ee976ad4854f3ac5a547435758e991823953566e6adb845714ef00cc","message_schema":"fresh-v7-staging-exclusion-lease.v1"}`
	task8FixedProofSourceBundle      = `{"evidence":[{"body_digest":"aa4e8bd2310938828695fd39dca6db9fb5909d4514cf29a7a82a5d6861b79910","canonical_body_or_null":null,"canonical_envelope_or_null":{"body":{"provider_control_sequence":"7","provider_phase":"fresh_v7_staging_closed"},"body_digest":"aa4e8bd2310938828695fd39dca6db9fb5909d4514cf29a7a82a5d6861b79910","schema":"claim-v1-provider-head.v1","signature":"k2Se04MvINfYJHP0Ck9lh1mhOnLXjQp5Qc9ZUCCdl_gItIwv1pjPvkQTMkNYmny8P0CPFUFcWuXrb1d1GH8yBw","signature_algorithm":"ed25519","signature_policy_version":"1","signer_key_id":"56475aa75463474c0285df5dbf2bcab73da651358839e9b77481b2eab107708c","signer_role":"claim_v1_provider","trust_root_digest":"a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1"},"evidence_kind":"external_signed_envelope","schema":"claim-v1-provider-head.v1"}],"evidence_count":"1","message_body_digest":"1e7c5d3674a2aa1adb81091f05e9b654c0610e7317462541bec074b0a739b9d3","message_schema":"database-incarnation-proof.v1"}`
)

var task8FixedStagingPreimages = [14][]byte{
	[]byte(`{"capability_id":"51000000-0000-4000-8000-000000000001","issued_at":"2026-08-29T17:00:00.123456Z","transaction_nonce":"0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"}`),
	[]byte(`{"body":{"capability_id":"51000000-0000-4000-8000-000000000001","issued_at":"2026-08-29T17:00:00.123456Z","transaction_nonce":"0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"},"body_digest":"631af925575d0c2a71a220b8b2245c2d7dc77c42a26598e5a2867ccf938e8178","schema":"staging-import-capability.v1","signature":"iWtuZU-M_v65B4-7OfhBGNpDBvRiVgyUkJaoLXbLoA8JrO6ea9oFkWqP-lC-OxaPS389J27IgJ0oBZxe2Sl0DQ","signature_algorithm":"ed25519","signature_policy_version":"1","signer_key_id":"56475aa75463474c0285df5dbf2bcab73da651358839e9b77481b2eab107708c","signer_role":"fresh_restore_staging_import_authorizer","trust_root_digest":"a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1"}`),
	[]byte(task8FixedCapabilitySourceBundle),
	[]byte(`{"manifest_id":"52000000-0000-4000-8000-000000000001","object_count":"4","single_use_apply_id":"52000000-0000-4000-8000-000000000002"}`),
	[]byte(`{"body":{"manifest_id":"52000000-0000-4000-8000-000000000001","object_count":"4","single_use_apply_id":"52000000-0000-4000-8000-000000000002"},"body_digest":"e2f9a5cc76ffb23e22f6abe3e783bf675156555f0ef257df347398884ce5b285","schema":"fresh-restore-import-manifest.v1","signature":"yEqWbcsqdNQVOywki-qEZdJ6KrrxU482EzXuMNyNK0QvQJ3WnBFzWysLJQCwP1XNCtx1EajtpRDk8hMgvFCBBg","signature_algorithm":"ed25519","signature_policy_version":"1","signer_key_id":"56475aa75463474c0285df5dbf2bcab73da651358839e9b77481b2eab107708c","signer_role":"fresh_restore_export_operator","trust_root_digest":"a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1"}`),
	[]byte(task8FixedManifestSourceBundle),
	[]byte(`{"exclusion_id":"53000000-0000-4000-8000-000000000001","provider_control_sequence":"7","state":"held"}`),
	[]byte(`{"body":{"exclusion_id":"53000000-0000-4000-8000-000000000001","provider_control_sequence":"7","state":"held"},"body_digest":"78fcc895ee976ad4854f3ac5a547435758e991823953566e6adb845714ef00cc","schema":"fresh-v7-staging-exclusion-lease.v1","signature":"WIcy0W5MtCVcYJvTIItcc8jAMT-9vdScnOpUomuXmMrKajEGMy0Og7IXgNwSGtg2i256Iusje1X9ejvJaU_JDw","signature_algorithm":"ed25519","signature_policy_version":"1","signer_key_id":"56475aa75463474c0285df5dbf2bcab73da651358839e9b77481b2eab107708c","signer_role":"claim_v1_provider","trust_root_digest":"a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1"}`),
	[]byte(task8FixedLeaseSourceBundle),
	[]byte(`{"provider_control_sequence":"7","provider_phase":"fresh_v7_staging_closed"}`),
	[]byte(`{"body":{"provider_control_sequence":"7","provider_phase":"fresh_v7_staging_closed"},"body_digest":"aa4e8bd2310938828695fd39dca6db9fb5909d4514cf29a7a82a5d6861b79910","schema":"claim-v1-provider-head.v1","signature":"k2Se04MvINfYJHP0Ck9lh1mhOnLXjQp5Qc9ZUCCdl_gItIwv1pjPvkQTMkNYmny8P0CPFUFcWuXrb1d1GH8yBw","signature_algorithm":"ed25519","signature_policy_version":"1","signer_key_id":"56475aa75463474c0285df5dbf2bcab73da651358839e9b77481b2eab107708c","signer_role":"claim_v1_provider","trust_root_digest":"a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1"}`),
	[]byte(`{"database_identity_digest":"5454545454545454545454545454545454545454545454545454545454545454","database_point":"0/16B6C50"}`),
	[]byte(`{"body":{"database_identity_digest":"5454545454545454545454545454545454545454545454545454545454545454","database_point":"0/16B6C50"},"body_digest":"1e7c5d3674a2aa1adb81091f05e9b654c0610e7317462541bec074b0a739b9d3","schema":"database-incarnation-proof.v1","signature":"bUnQkvSZ58TCn_1OKSlMzO1ZjmLrw0tZrOk2KLmMaLBvRBDcXVkOmmI-Gk5hup8g6chtcILKr21smj7odSixCg","signature_algorithm":"ed25519","signature_policy_version":"1","signer_key_id":"56475aa75463474c0285df5dbf2bcab73da651358839e9b77481b2eab107708c","signer_role":"claim_v1_provider","trust_root_digest":"a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1"}`),
	[]byte(task8FixedProofSourceBundle),
}

var task8FixedStagingEnvelopeBindings = []struct {
	body, envelope []byte
	schema, role   string
}{
	{task8FixedStagingPreimages[0], task8FixedStagingPreimages[1], "staging-import-capability.v1", "fresh_restore_staging_import_authorizer"},
	{task8FixedStagingPreimages[3], task8FixedStagingPreimages[4], "fresh-restore-import-manifest.v1", "fresh_restore_export_operator"},
	{task8FixedStagingPreimages[6], task8FixedStagingPreimages[7], "fresh-v7-staging-exclusion-lease.v1", "claim_v1_provider"},
	{task8FixedStagingPreimages[9], task8FixedStagingPreimages[10], "claim-v1-provider-head.v1", "claim_v1_provider"},
	{task8FixedStagingPreimages[11], task8FixedStagingPreimages[12], "database-incarnation-proof.v1", "claim_v1_provider"},
}

const task8FixedFreshProjectionBody = `{"normalized_catalog_digest":"28ba9545623b38b16a33fa2ee72daf367aa440c06bc1144300fa8feea2f6739b","object_count":"2","objects":[{"canonical_key":"54000000-0000-4000-8000-000000000001","normalized_payload":{"active_pointer_set":[],"authority_anchor_set":[],"health_state":"unknown","identity_epoch":"0","identity_state":"unauthorized","inventory_version":"1","next_desired_generation":"1","next_recovery_generation":"1","node_id":"54000000-0000-4000-8000-000000000001","operator_state":"disabled","pending_operator_transition_or_null":null,"pop_code":"fixture-pop","resume_operator_state_or_null":null,"security_state":"quarantined","security_version":"1"},"object_type":"node_reconstruction_seed"},{"canonical_key":"fixture-pop","normalized_payload":{"iso_country":"US","operator_state":"disabled","pop_code":"fixture-pop","region":"fixture-region","version":"1"},"object_type":"pop"}],"projection_version":"1","target_activation_id":"54000000-0000-4000-8000-000000000002","target_database_identity_digest":"787f2747e2d7b0b194f15236f69d53d37a8248ef387ae9f11c8ae6599e5c24bb","target_deployment_id":"54000000-0000-4000-8000-000000000003"}`
const task8FixedFreshProjectionDigestHex = "46251d95f89fecf930e877e0321822f509129d20e0093df2949b93ab23d42f9f"
const task8FixedEmptyFreshProjectionBody = `{"normalized_catalog_digest":"28ba9545623b38b16a33fa2ee72daf367aa440c06bc1144300fa8feea2f6739b","object_count":"0","objects":[],"projection_version":"1","target_activation_id":"54000000-0000-4000-8000-000000000002","target_database_identity_digest":"787f2747e2d7b0b194f15236f69d53d37a8248ef387ae9f11c8ae6599e5c24bb","target_deployment_id":"54000000-0000-4000-8000-000000000003"}`
const task8FixedEmptyFreshProjectionDigestHex = "71f16c2c7e8caa90ea81d69c92521ce258ea3a543abcf256f7fbfc2238bcc8c8"

func TestVerifiedFreshRestoreImportCreatesOnlyUnauthorizedZeroState(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	database, pool, migrationFS := openOwnedFreshImportDatabase(t)
	applyFreshImportV7(t, ctx, database, migrationFS)

	now := time.Now().UTC().Truncate(time.Microsecond)
	facts, projection, popCode, nodeID := freshImportAdmissionFixture(t, now)
	if _, err := pool.Exec(ctx, `
INSERT INTO nodecontrol.node_inventory(node_id,pop_code,operator_state,security_state,identity_state,created_at,updated_at)
VALUES($1,'blocked','disabled','quarantined','unauthorized',$2,$2)`, uuid.New(), now); err == nil {
		t.Fatal("caller-selectable unauthorized inventory insert succeeded")
	} else {
		var postgresError *pgconn.PgError
		if !errors.As(err, &postgresError) || postgresError.Code != "42501" {
			t.Fatalf("direct unauthorized insert error = %v, want SQLSTATE 42501", err)
		}
	}

	capability := facts.StagingImportCapability
	if _, err := pool.Exec(ctx, `
INSERT INTO nodecontrol.control_plane_authority_staging_import_capabilities(
  capability_id,single_use_apply_id,manifest_digest,target_activation_id,
  target_database_identity_digest,database_timeline_lineage_chain_digest,
  target_database_incarnation_registration_digest,runtime_rebind_chain_digest,
  runtime_instance_binding_digest,planned_staging_exclusion_id,
  expected_pre_acquire_provider_head_digest,pre_acquire_database_incarnation_proof_digest,
  provider_phase,provider_serving_lease_absent_digest,database_route_closed_digest,
  pre_import_inventory_digest,allowed_object_set_digest,expected_post_import_inventory_digest,
  transaction_nonce,issued_at,expires_at,canonical_evidence_bundle_jcs,canonical_body_jcs,body_digest)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24)`,
		capability.CapabilityID, capability.SingleUseApplyID, capability.ManifestDigest[:], capability.TargetActivationID,
		capability.TargetDatabaseIdentityDigest[:], capability.DatabaseTimelineLineageChainDigest[:],
		capability.TargetDatabaseIncarnationRegistrationDigest[:], capability.RuntimeRebindChainDigest[:],
		capability.RuntimeInstanceBindingDigest[:], capability.PlannedStagingExclusionID,
		capability.ExpectedPreAcquireProviderHeadDigest[:], capability.PreAcquireDatabaseIncarnationProofDigest[:],
		string(capability.ProviderPhase), capability.ProviderServingLeaseAbsentDigest[:], capability.DatabaseRouteClosedDigest[:],
		capability.PreImportInventoryDigest[:], capability.AllowedObjectSetDigest[:], capability.ExpectedPostImportInventoryDigest[:],
		capability.TransactionNonce[:], capability.IssuedAt, capability.ExpiresAt, []byte(`{"capability":"evidence"}`),
		[]byte(`{"capability":"body"}`), capability.CapabilityDigest[:]); err != nil {
		t.Fatal("seed verified staging capability:", err)
	}

	repository, err := authority.NewPostgresRepository(pool)
	if err != nil {
		t.Fatal(err)
	}
	admission, err := task8NewFreshRestoreImportAdmission(t, facts)
	if err != nil {
		t.Fatal("construct verified fresh-import admission:", err)
	}
	copyOfAdmission := admission
	application, err := repository.ConsumeVerifiedFreshRestoreImport(ctx, admission, projection)
	if err != nil {
		t.Fatal("consume verified fresh-import admission:", err)
	}
	if application.SingleUseApplyID != facts.ManifestTopology.SingleUseApplyID || application.ImportedObjectCount != 2 || application.CompleteNodeSetDigest != facts.ManifestTopology.CompleteNodeSetDigest {
		t.Fatalf("fresh-import application mismatch: %+v", application)
	}
	if _, err := repository.ConsumeVerifiedFreshRestoreImport(ctx, copyOfAdmission, projection); !errors.Is(err, authority.ErrConflict) {
		t.Fatalf("copied admission second consume = %v, want ErrConflict", err)
	}

	var operatorState, securityState, identityState string
	var identityEpoch, inventoryVersion, securityVersion, nextDesired, nextRecovery int64
	var lineageID, resumeState, pendingTransition, lastAuthority any
	if err := pool.QueryRow(ctx, `
SELECT operator_state,security_state,identity_state,identity_epoch,inventory_version,security_version,
       next_desired_generation,next_recovery_generation,lineage_id,resume_operator_state,
       pending_operator_transition,last_authority_operation_id
FROM nodecontrol.node_inventory WHERE node_id=$1 AND pop_code=$2`, nodeID, popCode).Scan(
		&operatorState, &securityState, &identityState, &identityEpoch, &inventoryVersion, &securityVersion,
		&nextDesired, &nextRecovery, &lineageID, &resumeState, &pendingTransition, &lastAuthority); err != nil {
		t.Fatal("read imported unauthorized inventory:", err)
	}
	if operatorState != "disabled" || securityState != "quarantined" || identityState != "unauthorized" ||
		identityEpoch != 0 || inventoryVersion != 1 || securityVersion != 1 || nextDesired != 1 || nextRecovery != 1 ||
		lineageID != nil || resumeState != nil || pendingTransition != nil || lastAuthority != nil {
		t.Fatalf("imported inventory is not the exact unauthorized zero state")
	}
	var applicationCount, popCount, nodeCount int
	if err := pool.QueryRow(ctx, `SELECT
 (SELECT count(*) FROM nodecontrol.control_plane_authority_fresh_restore_import_applications),
 (SELECT count(*) FROM nodecontrol.node_pops),
 (SELECT count(*) FROM nodecontrol.node_inventory)`).Scan(&applicationCount, &popCount, &nodeCount); err != nil {
		t.Fatal(err)
	}
	if applicationCount != 1 || popCount != 1 || nodeCount != 1 {
		t.Fatalf("fresh import counts = applications:%d pops:%d nodes:%d, want 1/1/1", applicationCount, popCount, nodeCount)
	}
}

func TestVerifiedFreshRestoreImportRejectMatrix(t *testing.T) {
	task8AssertFutureStagingFactorySignature(t)
	task8AssertFixedStagingEnvelopeConsistency(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	database, pool, migrationFS := openOwnedFreshImportDatabase(t)
	applyFreshImportV7(t, ctx, database, migrationFS)
	now := time.Date(2099, time.August, 29, 17, 0, 0, 123456000, time.UTC)
	facts, projection, _, _ := freshImportAdmissionFixture(t, now)
	capability := facts.StagingImportCapability
	if _, err := pool.Exec(ctx, `
INSERT INTO nodecontrol.control_plane_authority_staging_import_capabilities(
 capability_id,single_use_apply_id,manifest_digest,target_activation_id,target_database_identity_digest,
 database_timeline_lineage_chain_digest,target_database_incarnation_registration_digest,runtime_rebind_chain_digest,
 runtime_instance_binding_digest,planned_staging_exclusion_id,expected_pre_acquire_provider_head_digest,
 pre_acquire_database_incarnation_proof_digest,provider_phase,provider_serving_lease_absent_digest,database_route_closed_digest,
 pre_import_inventory_digest,allowed_object_set_digest,expected_post_import_inventory_digest,transaction_nonce,issued_at,expires_at,
 canonical_evidence_bundle_jcs,canonical_body_jcs,body_digest)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24)`,
		capability.CapabilityID, capability.SingleUseApplyID, capability.ManifestDigest[:], capability.TargetActivationID,
		capability.TargetDatabaseIdentityDigest[:], capability.DatabaseTimelineLineageChainDigest[:],
		capability.TargetDatabaseIncarnationRegistrationDigest[:], capability.RuntimeRebindChainDigest[:],
		capability.RuntimeInstanceBindingDigest[:], capability.PlannedStagingExclusionID,
		capability.ExpectedPreAcquireProviderHeadDigest[:], capability.PreAcquireDatabaseIncarnationProofDigest[:],
		string(capability.ProviderPhase), capability.ProviderServingLeaseAbsentDigest[:], capability.DatabaseRouteClosedDigest[:],
		capability.PreImportInventoryDigest[:], capability.AllowedObjectSetDigest[:], capability.ExpectedPostImportInventoryDigest[:],
		capability.TransactionNonce[:], capability.IssuedAt, capability.ExpiresAt, []byte(`{"capability":"evidence"}`),
		[]byte(`{"capability":"body"}`), capability.CapabilityDigest[:]); err != nil {
		t.Fatal("seed reject-matrix staging capability:", err)
	}
	repository, err := authority.NewPostgresRepository(pool)
	if err != nil {
		t.Fatal(err)
	}
	// A mutation arm is meaningful only after both registered and revised
	// unmutated routes prove that this exact fixture reaches admission.
	task8AssertCurrentStagingRoutePreflight(t, ctx, pool, facts)
	futureClosure, ok := task8AssertFutureEightByteaBaseline(t, ctx, pool, facts)
	if !ok {
		return
	}
	task8AssertFutureEvidenceFixtureMutations(t, futureClosure, task8FixedFutureSourceBundleBodies())

	// These are verifier-side failures. They begin from one valid fixture and
	// mutate exactly one caller-owned fact. No repository transaction may start.
	for _, testCase := range []struct {
		name   string
		mutate func(*contracts.FreshRestoreImportProjectionInputV1)
	}{
		{
			name: "duplicate manifest canonical key",
			mutate: func(value *contracts.FreshRestoreImportProjectionInputV1) {
				value.ManifestTopology.Objects[1].CanonicalKey = value.ManifestTopology.Objects[0].CanonicalKey
			},
		},
		{
			name: "noncanonical manifest payload bytes",
			mutate: func(value *contracts.FreshRestoreImportProjectionInputV1) {
				value.ManifestTopology.Objects[0].Payload = json.RawMessage(`{ "node_id":"` + value.ManifestTopology.Objects[0].CanonicalKey + `"}`)
			},
		},
		{
			name: "manifest closure exceeds one MiB",
			mutate: func(value *contracts.FreshRestoreImportProjectionInputV1) {
				value.ManifestTopology.Objects[0].Payload = json.RawMessage(`"` + strings.Repeat("x", 1048577) + `"`)
			},
		},
		{
			name: "capability runtime binding mismatch",
			mutate: func(value *contracts.FreshRestoreImportProjectionInputV1) {
				value.RuntimeInstanceBindingDigest[0] ^= 0xff
			},
		},
		{
			name: "lease provider Head mismatch",
			mutate: func(value *contracts.FreshRestoreImportProjectionInputV1) {
				value.StagingExclusionLease.AcquisitionLockedProviderHeadDigest[0] ^= 0xff
			},
		},
	} {
		t.Run("opaque "+testCase.name, func(t *testing.T) {
			mutated := facts
			mutated.ManifestTopology.Objects = append([]contracts.FreshRestoreImportManifestObjectV1(nil), facts.ManifestTopology.Objects...)
			testCase.mutate(&mutated)
			before := freshImportResidueSnapshot(t, ctx, pool)
			if _, err := task8NewFreshRestoreImportAdmission(t, mutated); err == nil {
				t.Errorf("opaque constructor accepted %s", testCase.name)
			}
			if after := freshImportResidueSnapshot(t, ctx, pool); !bytes.Equal(after, before) {
				t.Fatalf("opaque reject changed database residue\nbefore=%s\nafter=%s", before, after)
			}
		})
	}
	for _, testCase := range []struct {
		name   string
		mutate func([][]byte)
	}{
		{"capability body/envelope digest mismatch", func(values [][]byte) {
			values[0] = bytes.Replace(values[0], []byte("01020304"), []byte("ff020304"), 1)
		}},
		{"duplicate signed-envelope key", func(values [][]byte) {
			values[1] = bytes.Replace(values[1], []byte(`{"body":`), []byte(`{"schema":"duplicate","body":`), 1)
		}},
		{"noncanonical Manifest body", func(values [][]byte) { values[3] = append([]byte("{ "), values[3][1:]...) }},
		{"wrong signature bytes", func(values [][]byte) {
			values[1] = bytes.Replace(values[1], []byte(`"signature":"iW`), []byte(`"signature":"jW`), 1)
		}},
		{"wrong schema metadata", func(values [][]byte) {
			values[4] = bytes.Replace(values[4], []byte("fresh-restore-import-manifest.v1"), []byte("fresh-restore-import-manifest.v2"), 1)
		}},
		{"wrong signer role", func(values [][]byte) {
			values[7] = bytes.Replace(values[7], []byte("claim_v1_provider"), []byte("fresh_restore_export_operator"), 1)
		}},
		{"wrong signer key", func(values [][]byte) {
			values[10] = bytes.Replace(values[10], []byte("56475aa7"), []byte("66475aa7"), 1)
		}},
		{"wrong signature policy", func(values [][]byte) {
			values[12] = bytes.Replace(values[12], []byte(`"signature_policy_version":"1"`), []byte(`"signature_policy_version":"2"`), 1)
		}},
		{"wrong trust root", func(values [][]byte) {
			values[1] = bytes.Replace(values[1], []byte(strings.Repeat("a1", 32)), []byte(strings.Repeat("b1", 32)), 1)
		}},
		{"capability source bundle missing member", func(values [][]byte) { values[2] = task8MutateFixedSourceBundle(t, values[2], "missing-member") }},
		{"Manifest source bundle extra member", func(values [][]byte) { values[5] = task8MutateFixedSourceBundle(t, values[5], "extra-member") }},
		{"lease source bundle wrong kind", func(values [][]byte) { values[8] = task8MutateFixedSourceBundle(t, values[8], "wrong-kind") }},
		{"proof source bundle wrong schema", func(values [][]byte) { values[13] = task8MutateFixedSourceBundle(t, values[13], "wrong-schema") }},
		{"capability source bundle non-reconstructible", func(values [][]byte) { values[2] = task8MutateFixedSourceBundle(t, values[2], "non-reconstructible") }},
		{"seven-root closure exceeds one MiB", func(values [][]byte) { values[13] = []byte(`"` + strings.Repeat("x", 1048577) + `"`) }},
	} {
		t.Run("opaque preimage "+testCase.name, func(t *testing.T) {
			before := freshImportResidueSnapshot(t, ctx, pool)
			if _, err := task8NewFreshRestoreImportAdmissionWithPreimages(t, facts, testCase.mutate); err == nil {
				t.Errorf("revised opaque fixture accepted %s", testCase.name)
			}
			if after := freshImportResidueSnapshot(t, ctx, pool); !bytes.Equal(after, before) {
				t.Fatalf("opaque preimage reject changed residue\nbefore=%s\nafter=%s", before, after)
			}
		})
	}

	for parameterIndex, label := range []string{
		"capability digest", "Manifest digest", "staging exclusion lease digest", "acquisition-locked Head digest",
		"projection body", "projection digest", "application body", "application evidence seven-root closure",
	} {
		t.Run("restricted eight-bytea "+label, func(t *testing.T) {
			task8RunRestrictedStagingReject(t, ctx, pool, facts, label, func(arguments *[8][]byte) {
				switch parameterIndex {
				case 4:
					arguments[4] = task8MutateCanonicalJSONObjectField(t, arguments[4], "target_activation_id")
				case 6:
					arguments[6] = task8MutateCanonicalJSONObjectField(t, arguments[6], "manifest_digest")
				case 7:
					arguments[7] = bytes.Replace(arguments[7], []byte(`"evidence_count":"7"`), []byte(`"evidence_count":"6"`), 1)
				default:
					arguments[parameterIndex][0] ^= 0x01
				}
			})
		})
	}

	task8AssertDirectStagingBodyMutationMatrix(t, ctx, pool, facts)

	for _, testCase := range []struct {
		name   string
		mutate func(*contracts.FreshImportTopologyProjectionV1)
	}{
		{"target activation digest parameter", func(value *contracts.FreshImportTopologyProjectionV1) { value.TargetActivationID = uuid.New() }},
		{"target deployment normalized field", func(value *contracts.FreshImportTopologyProjectionV1) { value.TargetDeploymentID = uuid.New() }},
		{"database identity digest parameter", func(value *contracts.FreshImportTopologyProjectionV1) { value.TargetDatabaseIdentityDigest[0] ^= 0xff }},
		{"normalized catalog digest parameter", func(value *contracts.FreshImportTopologyProjectionV1) { value.NormalizedCatalogDigest[0] ^= 0xff }},
		{"manifest object count", func(value *contracts.FreshImportTopologyProjectionV1) { value.ObjectCount++ }},
		{"manifest object order", func(value *contracts.FreshImportTopologyProjectionV1) {
			value.Objects[0], value.Objects[1] = value.Objects[1], value.Objects[0]
		}},
		{"manifest object kind", func(value *contracts.FreshImportTopologyProjectionV1) {
			value.Objects[0].ObjectType = contracts.FreshRestoreImportObjectPOP
		}},
		{"manifest object canonical key", func(value *contracts.FreshImportTopologyProjectionV1) { value.Objects[0].CanonicalKey = "wrong-key" }},
		{"manifest normalized payload", func(value *contracts.FreshImportTopologyProjectionV1) {
			value.Objects[0].NormalizedPayload = json.RawMessage(`{"node_id":"wrong"}`)
		}},
	} {
		t.Run("SQL-visible "+testCase.name, func(t *testing.T) {
			mutated := projection
			mutated.Objects = append([]contracts.FreshImportTopologyProjectionObjectV1(nil), projection.Objects...)
			testCase.mutate(&mutated)
			admission, err := task8NewFreshRestoreImportAdmission(t, facts)
			if err != nil {
				t.Fatal(err)
			}
			before := freshImportResidueSnapshot(t, ctx, pool)
			if _, err := repository.ConsumeVerifiedFreshRestoreImport(ctx, admission, mutated); err == nil {
				t.Errorf("repository accepted %s", testCase.name)
			}
			if after := freshImportResidueSnapshot(t, ctx, pool); !bytes.Equal(after, before) {
				t.Fatalf("SQL-visible reject changed application/domain/control rows\nbefore=%s\nafter=%s", before, after)
			}
		})
	}
}

const task8CurrentStagingProjectionObjects = `[{"object_type":"node_reconstruction_seed","canonical_key":"54000000-0000-4000-8000-000000000001","body":{"node_id":"54000000-0000-4000-8000-000000000001","pop_code":"fixture-pop","operator_state":"disabled","security_state":"quarantined","identity_state":"unauthorized","health_state":"unknown","identity_epoch":"0","inventory_version":"1","security_version":"1","next_desired_generation":"1","next_recovery_generation":"1","resume_operator_state_or_null":null,"pending_operator_transition_or_null":null,"active_pointer_set":[],"authority_anchor_set":[]}},{"object_type":"pop","canonical_key":"fixture-pop","body":{"pop_code":"fixture-pop","iso_country":"US","region":"fixture-region","operator_state":"disabled","version":"1"}}]`

func task8BeginExactStagingPrincipal(t *testing.T, ctx context.Context, pool *pgxpool.Pool) pgx.Tx {
	t.Helper()
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `SET LOCAL SESSION AUTHORIZATION nodecontrol_staging_importer`); err != nil {
		_ = tx.Rollback(ctx)
		t.Fatal("assume exact staging session principal:", err)
	}
	var sessionUser string
	var importerExecute, upgraderExecute, downgraderExecute, directPOPInsert bool
	if err := tx.QueryRow(ctx, `SELECT session_user::text,
 has_function_privilege('nodecontrol_staging_importer','nodecontrol.begin_staging_import(bytea,bytea,bytea,bytea,bytea,bytea,bytea,jsonb)','EXECUTE'),
 has_function_privilege('nodecontrol_upgrade_executor','nodecontrol.begin_staging_import(bytea,bytea,bytea,bytea,bytea,bytea,bytea,jsonb)','EXECUTE'),
 has_function_privilege('nodecontrol_migration_downgrader','nodecontrol.begin_staging_import(bytea,bytea,bytea,bytea,bytea,bytea,bytea,jsonb)','EXECUTE'),
 has_table_privilege(session_user,'nodecontrol.node_pops','INSERT')`).Scan(
		&sessionUser, &importerExecute, &upgraderExecute, &downgraderExecute, &directPOPInsert); err != nil {
		_ = tx.Rollback(ctx)
		t.Fatal("inspect exact staging principal privileges:", err)
	}
	if sessionUser != "nodecontrol_staging_importer" || !importerExecute || upgraderExecute || downgraderExecute || directPOPInsert {
		_ = tx.Rollback(ctx)
		t.Fatalf("staging principal/EXECUTE/direct-DML = %q/%t/%t/%t/%t, want importer/true/false/false/false", sessionUser, importerExecute, upgraderExecute, downgraderExecute, directPOPInsert)
	}
	return tx
}

func task8AssertCurrentStagingRoutePreflight(t *testing.T, ctx context.Context, pool *pgxpool.Pool, facts contracts.FreshRestoreImportProjectionInputV1) {
	t.Helper()
	before := freshImportResidueSnapshot(t, ctx, pool)
	tx := task8BeginExactStagingPrincipal(t, ctx, pool)
	var rowCount int
	err := tx.QueryRow(ctx, `SELECT count(*) FROM nodecontrol.begin_staging_import(
 $1::bytea,$2::bytea,$3::bytea,$4::bytea,$5::bytea,$6::bytea,$7::bytea,$8::jsonb)`,
		facts.StagingImportCapability.CapabilityDigest[:], facts.ManifestTopology.ManifestDigest[:],
		facts.StagingExclusionLease.LeaseDigest[:], facts.CurrentProviderHeadDigest[:], facts.DatabaseRouteClosedDigest[:],
		[]byte(task8FixedFreshProjectionBody), facts.StagingImportCapability.ExpectedPostImportInventoryDigest[:],
		task8CurrentStagingProjectionObjects).Scan(&rowCount)
	_ = tx.Rollback(ctx)
	if err != nil || rowCount != 1 {
		t.Fatalf("registered seven-bytea/jsonb positive control rows/error = %d/%v, want exactly 1/nil", rowCount, err)
	}
	if after := freshImportResidueSnapshot(t, ctx, pool); !bytes.Equal(after, before) {
		t.Fatalf("rolled-back current-route positive control changed residue\nbefore=%s\nafter=%s", before, after)
	}
}

func task8AssertFutureEightByteaBaseline(t *testing.T, ctx context.Context, pool *pgxpool.Pool, facts contracts.FreshRestoreImportProjectionInputV1) ([]byte, bool) {
	t.Helper()
	before := freshImportResidueSnapshot(t, ctx, pool)
	tx := task8BeginExactStagingPrincipal(t, ctx, pool)
	arguments := task8BuildFutureEightBytea(t, ctx, tx, facts)
	var rowCount int
	err := tx.QueryRow(ctx, `SELECT count(*) FROM nodecontrol.begin_staging_import(
 $1::bytea,$2::bytea,$3::bytea,$4::bytea,$5::bytea,$6::bytea,$7::bytea,$8::bytea)`,
		arguments[0], arguments[1], arguments[2], arguments[3], arguments[4], arguments[5], arguments[6], arguments[7]).Scan(&rowCount)
	_ = tx.Rollback(ctx)
	if err != nil {
		var postgresError *pgconn.PgError
		if errors.As(err, &postgresError) && postgresError.Code == "42883" {
			t.Errorf("revised eight-bytea staging entry is absent: SQLSTATE 42883 (%s)", postgresError.Message)
		} else {
			t.Errorf("unmutated revised eight-bytea baseline failed before admission: %v", err)
		}
		return nil, false
	}
	if rowCount != 1 {
		t.Errorf("unmutated revised eight-bytea baseline row count=%d, want exactly 1", rowCount)
		return nil, false
	}
	if after := freshImportResidueSnapshot(t, ctx, pool); !bytes.Equal(after, before) {
		t.Fatalf("rolled-back revised baseline changed residue\nbefore=%s\nafter=%s", before, after)
	}
	return append([]byte(nil), arguments[7]...), true
}

type task8FutureApplicationWire struct {
	AcquisitionLockedProviderHeadDigest                  string  `json:"acquisition_locked_provider_head_digest"`
	AppliedAt                                            string  `json:"applied_at"`
	CompleteNodeSetDigest                                string  `json:"complete_node_set_digest"`
	CurrentDatabaseIdentityDigest                        string  `json:"current_database_identity_digest"`
	DatabaseRouteClosedDigest                            string  `json:"database_route_closed_digest"`
	DatabaseTimelineLineageChainDigest                   string  `json:"database_timeline_lineage_chain_digest"`
	DatabaseTransactionID                                string  `json:"database_transaction_id"`
	ForbiddenStateZeroDigest                             string  `json:"forbidden_state_zero_digest"`
	ImportedObjectCount                                  string  `json:"imported_object_count"`
	ManifestDigest                                       string  `json:"manifest_digest"`
	PostImportInventoryDigest                            string  `json:"post_import_inventory_digest"`
	PreImportInventoryDigest                             string  `json:"pre_import_inventory_digest"`
	RuntimeInstanceBindingDigest                         string  `json:"runtime_instance_binding_digest"`
	RuntimeRebindChainDigest                             string  `json:"runtime_rebind_chain_digest"`
	SingleUseApplyID                                     string  `json:"single_use_apply_id"`
	StagingExclusionLeaseDigest                          string  `json:"staging_exclusion_lease_digest"`
	StagingImportCapabilityDigest                        string  `json:"staging_import_capability_digest"`
	StagingImportCapabilityRecoveryApplicationDigestNull *string `json:"staging_import_capability_recovery_application_digest_or_null"`
	StagingImportCapabilityRecoveryIntentDigestNull      *string `json:"staging_import_capability_recovery_intent_digest_or_null"`
	TargetActivationID                                   string  `json:"target_activation_id"`
	TargetDatabaseIncarnationRegistrationDigest          string  `json:"target_database_incarnation_registration_digest"`
	TransactionSnapshotDigest                            string  `json:"transaction_snapshot_digest"`
}

type task8FutureEvidenceItem struct {
	BodyDigest              string          `json:"body_digest"`
	CanonicalBodyOrNull     json.RawMessage `json:"canonical_body_or_null"`
	CanonicalEnvelopeOrNull json.RawMessage `json:"canonical_envelope_or_null"`
	EvidenceKind            string          `json:"evidence_kind"`
	Schema                  string          `json:"schema"`
}

type task8FutureEvidenceBundle struct {
	Evidence          []task8FutureEvidenceItem `json:"evidence"`
	EvidenceCount     string                    `json:"evidence_count"`
	MessageBodyDigest string                    `json:"message_body_digest"`
	MessageSchema     string                    `json:"message_schema"`
}

func task8BuildFutureEightBytea(t *testing.T, ctx context.Context, tx pgx.Tx, facts contracts.FreshRestoreImportProjectionInputV1) [8][]byte {
	t.Helper()
	var transactionID string
	var transactionSnapshot string
	var appliedAt time.Time
	if err := tx.QueryRow(ctx, `SELECT txid_current()::text,txid_current_snapshot()::text,transaction_timestamp()`).Scan(
		&transactionID, &transactionSnapshot, &appliedAt); err != nil {
		t.Fatal("read future staging transaction facts:", err)
	}
	parsedTransactionID, err := strconv.ParseUint(transactionID, 10, 64)
	if err != nil || parsedTransactionID == 0 || strconv.FormatUint(parsedTransactionID, 10) != transactionID || transactionSnapshot == "" || appliedAt.IsZero() {
		t.Fatalf("future staging transaction facts are not closed: xid=%q snapshot=%q applied=%s error=%v", transactionID, transactionSnapshot, appliedAt, err)
	}
	snapshotDigest := sha256.Sum256([]byte(transactionSnapshot))
	forbiddenDigest := sha256.Sum256(append(append([]byte("talenro.c12.fresh-import-forbidden-state-zero.v1"), 0), []byte(`{"forbidden_row_count":"0"}`)...))
	application := task8FutureApplicationWire{
		AcquisitionLockedProviderHeadDigest:         hex.EncodeToString(facts.CurrentProviderHeadDigest[:]),
		AppliedAt:                                   appliedAt.UTC().Format(time.RFC3339Nano),
		CompleteNodeSetDigest:                       hex.EncodeToString(facts.ManifestTopology.CompleteNodeSetDigest[:]),
		CurrentDatabaseIdentityDigest:               hex.EncodeToString(facts.CurrentDatabaseIdentityDigest[:]),
		DatabaseRouteClosedDigest:                   hex.EncodeToString(facts.DatabaseRouteClosedDigest[:]),
		DatabaseTimelineLineageChainDigest:          hex.EncodeToString(facts.DatabaseTimelineLineageChainDigest[:]),
		DatabaseTransactionID:                       transactionID,
		ForbiddenStateZeroDigest:                    hex.EncodeToString(forbiddenDigest[:]),
		ImportedObjectCount:                         strconv.FormatUint(facts.ManifestTopology.ObjectCount, 10),
		ManifestDigest:                              hex.EncodeToString(facts.ManifestTopology.ManifestDigest[:]),
		PostImportInventoryDigest:                   hex.EncodeToString(facts.StagingImportCapability.ExpectedPostImportInventoryDigest[:]),
		PreImportInventoryDigest:                    hex.EncodeToString(facts.PreImportInventoryDigest[:]),
		RuntimeInstanceBindingDigest:                hex.EncodeToString(facts.RuntimeInstanceBindingDigest[:]),
		RuntimeRebindChainDigest:                    hex.EncodeToString(facts.RuntimeRebindChainDigest[:]),
		SingleUseApplyID:                            facts.ManifestTopology.SingleUseApplyID.String(),
		StagingExclusionLeaseDigest:                 hex.EncodeToString(facts.StagingExclusionLease.LeaseDigest[:]),
		StagingImportCapabilityDigest:               hex.EncodeToString(facts.StagingImportCapability.CapabilityDigest[:]),
		TargetActivationID:                          facts.ManifestTopology.TargetActivationID.String(),
		TargetDatabaseIncarnationRegistrationDigest: hex.EncodeToString(facts.CurrentDatabaseIncarnationRegistrationDigest[:]),
		TransactionSnapshotDigest:                   hex.EncodeToString(snapshotDigest[:]),
	}
	applicationBody, err := json.Marshal(application)
	if err != nil {
		t.Fatal(err)
	}
	applicationBody, err = jcs.Transform(applicationBody)
	if err != nil {
		t.Fatal("canonicalize future application body:", err)
	}
	task8AssertFutureApplicationWire(t, applicationBody, application, facts)
	applicationDigest := sha256.Sum256(append(append([]byte("talenro.c12.fresh-restore-import-application.v1"), 0), applicationBody...))
	roots := []task8FutureEvidenceItem{
		{task8FixedCapabilityDigestHex, json.RawMessage("null"), task8FixedStagingPreimages[1], "external_signed_envelope", "staging-import-capability.v1"},
		{task8FixedManifestDigestHex, json.RawMessage("null"), task8FixedStagingPreimages[4], "external_signed_envelope", "fresh-restore-import-manifest.v1"},
		{task8FixedLeaseDigestHex, json.RawMessage("null"), task8FixedStagingPreimages[7], "external_signed_envelope", "fresh-v7-staging-exclusion-lease.v1"},
		{task8FixedProviderHeadHex, json.RawMessage("null"), task8FixedStagingPreimages[10], "external_signed_envelope", "claim-v1-provider-head.v1"},
		{task8FixedProofDigestHex, json.RawMessage("null"), task8FixedStagingPreimages[12], "external_signed_envelope", "database-incarnation-proof.v1"},
		{task8FixedEmptyFreshProjectionDigestHex, json.RawMessage(task8FixedEmptyFreshProjectionBody), json.RawMessage("null"), "database_immutable_body", "fresh-import-topology-projection.v1"},
		{task8FixedFreshProjectionDigestHex, json.RawMessage(task8FixedFreshProjectionBody), json.RawMessage("null"), "database_immutable_body", "fresh-import-topology-projection.v1"},
	}
	sourceBundles := task8FixedFutureSourceBundleBodies()
	evidenceBody, err := task8BuildFutureEvidenceClosure(hex.EncodeToString(applicationDigest[:]), roots, sourceBundles)
	if err != nil {
		t.Fatal("construct exact future application evidence closure:", err)
	}
	projectionDigest := mustFreshImportDigestHex(t, task8FixedFreshProjectionDigestHex)
	if !bytes.Equal(projectionDigest, facts.StagingImportCapability.ExpectedPostImportInventoryDigest[:]) {
		t.Fatal("reviewed projection digest is not the typed fixture post-import digest")
	}
	preProjectionDigest := mustFreshImportDigestHex(t, task8FixedEmptyFreshProjectionDigestHex)
	if !bytes.Equal(preProjectionDigest, facts.PreImportInventoryDigest[:]) {
		t.Fatal("reviewed empty projection digest is not the typed fixture pre-import digest")
	}
	for _, binding := range []struct {
		name string
		got  []byte
		want string
	}{
		{"capability", facts.StagingImportCapability.CapabilityDigest[:], task8FixedCapabilityDigestHex},
		{"Manifest", facts.ManifestTopology.ManifestDigest[:], task8FixedManifestDigestHex},
		{"lease", facts.StagingExclusionLease.LeaseDigest[:], task8FixedLeaseDigestHex},
		{"provider Head", facts.CurrentProviderHeadDigest[:], task8FixedProviderHeadHex},
		{"incarnation proof", facts.StagingImportCapability.PreAcquireDatabaseIncarnationProofDigest[:], task8FixedProofDigestHex},
	} {
		if hex.EncodeToString(binding.got) != binding.want {
			t.Fatalf("typed %s digest=%x, want exact envelope body digest %s", binding.name, binding.got, binding.want)
		}
	}
	return [8][]byte{
		append([]byte(nil), facts.StagingImportCapability.CapabilityDigest[:]...),
		append([]byte(nil), facts.ManifestTopology.ManifestDigest[:]...),
		append([]byte(nil), facts.StagingExclusionLease.LeaseDigest[:]...),
		append([]byte(nil), facts.CurrentProviderHeadDigest[:]...),
		[]byte(task8FixedFreshProjectionBody), projectionDigest, applicationBody, evidenceBody,
	}
}

func task8FixedFutureSourceBundleBodies() [][]byte {
	return [][]byte{
		[]byte(task8FixedCapabilitySourceBundle), []byte(task8FixedManifestSourceBundle),
		[]byte(task8FixedLeaseSourceBundle), []byte(task8FixedProofSourceBundle),
	}
}

func task8AssertFutureApplicationWire(
	t *testing.T,
	body []byte,
	wire task8FutureApplicationWire,
	facts contracts.FreshRestoreImportProjectionInputV1,
) {
	t.Helper()
	var object map[string]json.RawMessage
	if err := json.Unmarshal(body, &object); err != nil {
		t.Fatal("decode exact future application wire:", err)
	}
	want := map[string]string{
		"acquisition_locked_provider_head_digest": hex.EncodeToString(facts.CurrentProviderHeadDigest[:]),
		"applied_at":                                      wire.AppliedAt,
		"complete_node_set_digest":                        hex.EncodeToString(facts.ManifestTopology.CompleteNodeSetDigest[:]),
		"current_database_identity_digest":                hex.EncodeToString(facts.CurrentDatabaseIdentityDigest[:]),
		"database_route_closed_digest":                    hex.EncodeToString(facts.DatabaseRouteClosedDigest[:]),
		"database_timeline_lineage_chain_digest":          hex.EncodeToString(facts.DatabaseTimelineLineageChainDigest[:]),
		"database_transaction_id":                         wire.DatabaseTransactionID,
		"forbidden_state_zero_digest":                     wire.ForbiddenStateZeroDigest,
		"imported_object_count":                           strconv.FormatUint(facts.ManifestTopology.ObjectCount, 10),
		"manifest_digest":                                 hex.EncodeToString(facts.ManifestTopology.ManifestDigest[:]),
		"post_import_inventory_digest":                    hex.EncodeToString(facts.StagingImportCapability.ExpectedPostImportInventoryDigest[:]),
		"pre_import_inventory_digest":                     hex.EncodeToString(facts.PreImportInventoryDigest[:]),
		"runtime_instance_binding_digest":                 hex.EncodeToString(facts.RuntimeInstanceBindingDigest[:]),
		"runtime_rebind_chain_digest":                     hex.EncodeToString(facts.RuntimeRebindChainDigest[:]),
		"single_use_apply_id":                             facts.ManifestTopology.SingleUseApplyID.String(),
		"staging_exclusion_lease_digest":                  hex.EncodeToString(facts.StagingExclusionLease.LeaseDigest[:]),
		"staging_import_capability_digest":                hex.EncodeToString(facts.StagingImportCapability.CapabilityDigest[:]),
		"target_activation_id":                            facts.ManifestTopology.TargetActivationID.String(),
		"target_database_incarnation_registration_digest": hex.EncodeToString(facts.CurrentDatabaseIncarnationRegistrationDigest[:]),
		"transaction_snapshot_digest":                     wire.TransactionSnapshotDigest,
	}
	if len(object) != 22 || len(want) != 20 {
		t.Fatalf("future application wire field count=%d/want-bindings=%d, want exact 22/20", len(object), len(want))
	}
	for field, value := range want {
		if got := string(object[field]); got != strconv.Quote(value) {
			t.Fatalf("future application wire %s=%s, want exact JSON string %q", field, got, value)
		}
		if strings.HasSuffix(field, "_digest") && (len(value) != 64 || value != strings.ToLower(value)) {
			t.Fatalf("future application wire %s=%q, want lowercase 64-hex digest", field, value)
		}
	}
	for _, field := range []string{
		"staging_import_capability_recovery_application_digest_or_null",
		"staging_import_capability_recovery_intent_digest_or_null",
	} {
		if got := string(object[field]); got != "null" {
			t.Fatalf("future application wire %s=%s, want literal null", field, got)
		}
	}
	if got := string(object["imported_object_count"]); got != `"`+strconv.FormatUint(facts.ManifestTopology.ObjectCount, 10)+`"` {
		t.Fatalf("future application imported_object_count=%s, want decimal JSON string", got)
	}
	if got := string(object["database_transaction_id"]); got != `"`+wire.DatabaseTransactionID+`"` {
		t.Fatalf("future application database_transaction_id=%s, want same-transaction decimal JSON string", got)
	}
}

func task8BuildFutureEvidenceClosure(
	applicationDigest string,
	roots []task8FutureEvidenceItem,
	sourceBodies [][]byte,
) ([]byte, error) {
	if len(roots) != 7 || len(sourceBodies) != 4 {
		return nil, fmt.Errorf("future evidence roots/source bundles=%d/%d, want exact 7/4", len(roots), len(sourceBodies))
	}
	allItems := append([]task8FutureEvidenceItem(nil), roots...)
	for index, sourceBody := range sourceBodies {
		source, err := task8ParseFutureEvidenceBundle(sourceBody)
		if err != nil {
			return nil, fmt.Errorf("source bundle %d: %w", index, err)
		}
		allItems = append(allItems, source.Evidence...)
	}
	deduplicated := make(map[string]task8FutureEvidenceItem, len(allItems))
	encoded := make(map[string][]byte, len(allItems))
	for _, item := range allItems {
		itemBytes, err := task8CanonicalFutureEvidenceItem(item)
		if err != nil {
			return nil, err
		}
		validated, err := task8ParseFutureEvidenceItem(itemBytes)
		if err != nil {
			return nil, err
		}
		identity := validated.BodyDigest + "\x00" + validated.Schema
		if previous, exists := encoded[identity]; exists {
			if !bytes.Equal(previous, itemBytes) {
				return nil, fmt.Errorf("conflicting evidence for digest/schema %s/%s", validated.BodyDigest, validated.Schema)
			}
			continue
		}
		deduplicated[identity] = validated
		encoded[identity] = itemBytes
	}
	items := make([]task8FutureEvidenceItem, 0, len(deduplicated))
	for _, item := range deduplicated {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool { return task8CompareFutureEvidenceItems(items[i], items[j]) < 0 })
	bundle := task8FutureEvidenceBundle{
		Evidence:          items,
		EvidenceCount:     strconv.Itoa(len(items)),
		MessageBodyDigest: applicationDigest,
		MessageSchema:     "fresh-restore-import-application.v1",
	}
	body, err := task8CanonicalFutureEvidenceBundle(bundle)
	if err != nil {
		return nil, err
	}
	if _, err := task8ParseFutureEvidenceBundle(body); err != nil {
		return nil, fmt.Errorf("constructed application bundle: %w", err)
	}
	if err := task8ReconstructFutureSourceBundles(body, sourceBodies); err != nil {
		return nil, err
	}
	return body, nil
}

func task8ParseFutureEvidenceBundle(body []byte) (task8FutureEvidenceBundle, error) {
	canonical, err := jcs.Transform(body)
	if err != nil || !bytes.Equal(canonical, body) {
		return task8FutureEvidenceBundle{}, fmt.Errorf("evidence bundle is not canonical JCS: %w", err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return task8FutureEvidenceBundle{}, err
	}
	if !task8HasExactJSONKeys(raw, []string{"evidence", "evidence_count", "message_body_digest", "message_schema"}) {
		return task8FutureEvidenceBundle{}, fmt.Errorf("evidence bundle top-level keys are not the exact four-field registry")
	}
	var rawItems []json.RawMessage
	if err := json.Unmarshal(raw["evidence"], &rawItems); err != nil {
		return task8FutureEvidenceBundle{}, err
	}
	var bundle task8FutureEvidenceBundle
	if err := json.Unmarshal(body, &bundle); err != nil {
		return task8FutureEvidenceBundle{}, err
	}
	count, err := strconv.ParseUint(bundle.EvidenceCount, 10, 64)
	if err != nil || strconv.FormatUint(count, 10) != bundle.EvidenceCount || count != uint64(len(rawItems)) {
		return task8FutureEvidenceBundle{}, fmt.Errorf("evidence_count=%q/items=%d is not one exact decimal binding", bundle.EvidenceCount, len(rawItems))
	}
	if err := task8RequireLowerDigest(bundle.MessageBodyDigest); err != nil || bundle.MessageSchema == "" {
		return task8FutureEvidenceBundle{}, fmt.Errorf("invalid message binding: %w", err)
	}
	bundle.Evidence = make([]task8FutureEvidenceItem, len(rawItems))
	for index, rawItem := range rawItems {
		item, err := task8ParseFutureEvidenceItem(rawItem)
		if err != nil {
			return task8FutureEvidenceBundle{}, fmt.Errorf("evidence item %d: %w", index, err)
		}
		if index > 0 && task8CompareFutureEvidenceItems(bundle.Evidence[index-1], item) >= 0 {
			return task8FutureEvidenceBundle{}, fmt.Errorf("evidence is not strictly sorted and duplicate-free at item %d", index)
		}
		bundle.Evidence[index] = item
	}
	return bundle, nil
}

func task8ParseFutureEvidenceItem(raw []byte) (task8FutureEvidenceItem, error) {
	canonical, err := jcs.Transform(raw)
	if err != nil || !bytes.Equal(canonical, raw) {
		return task8FutureEvidenceItem{}, fmt.Errorf("item is not canonical JCS: %w", err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return task8FutureEvidenceItem{}, err
	}
	if !task8HasExactJSONKeys(fields, []string{"body_digest", "canonical_body_or_null", "canonical_envelope_or_null", "evidence_kind", "schema"}) {
		return task8FutureEvidenceItem{}, fmt.Errorf("item does not have the exact five-field registry")
	}
	var item task8FutureEvidenceItem
	if err := json.Unmarshal(raw, &item); err != nil {
		return task8FutureEvidenceItem{}, err
	}
	if err := task8RequireLowerDigest(item.BodyDigest); err != nil || item.Schema == "" {
		return task8FutureEvidenceItem{}, fmt.Errorf("invalid item digest/schema: %w", err)
	}
	bodyNull := bytes.Equal(item.CanonicalBodyOrNull, []byte("null"))
	envelopeNull := bytes.Equal(item.CanonicalEnvelopeOrNull, []byte("null"))
	if bodyNull == envelopeNull {
		return task8FutureEvidenceItem{}, fmt.Errorf("item must have exactly one non-null preimage arm")
	}
	switch item.EvidenceKind {
	case "database_immutable_body":
		if bodyNull || !envelopeNull {
			return task8FutureEvidenceItem{}, fmt.Errorf("database body item has wrong preimage arm")
		}
		if canonical, err := jcs.Transform(item.CanonicalBodyOrNull); err != nil || !bytes.Equal(canonical, item.CanonicalBodyOrNull) {
			return task8FutureEvidenceItem{}, fmt.Errorf("database body preimage is noncanonical: %w", err)
		}
		digest := sha256.Sum256(append(append([]byte("talenro.c12."+item.Schema), 0), item.CanonicalBodyOrNull...))
		if item.BodyDigest != hex.EncodeToString(digest[:]) {
			return task8FutureEvidenceItem{}, fmt.Errorf("database body digest/preimage mismatch")
		}
	case "external_signed_envelope":
		if !bodyNull || envelopeNull {
			return task8FutureEvidenceItem{}, fmt.Errorf("signed envelope item has wrong preimage arm")
		}
		if canonical, err := jcs.Transform(item.CanonicalEnvelopeOrNull); err != nil || !bytes.Equal(canonical, item.CanonicalEnvelopeOrNull) {
			return task8FutureEvidenceItem{}, fmt.Errorf("envelope preimage is noncanonical: %w", err)
		}
		var envelope struct {
			Body       json.RawMessage `json:"body"`
			BodyDigest string          `json:"body_digest"`
			Schema     string          `json:"schema"`
		}
		if err := json.Unmarshal(item.CanonicalEnvelopeOrNull, &envelope); err != nil {
			return task8FutureEvidenceItem{}, err
		}
		if envelope.Schema != item.Schema || envelope.BodyDigest != item.BodyDigest {
			return task8FutureEvidenceItem{}, fmt.Errorf("envelope metadata does not match item schema/digest")
		}
		digest := sha256.Sum256(append(append([]byte("talenro.c12."+envelope.Schema), 0), envelope.Body...))
		if envelope.BodyDigest != hex.EncodeToString(digest[:]) {
			return task8FutureEvidenceItem{}, fmt.Errorf("envelope body digest/preimage mismatch")
		}
	default:
		return task8FutureEvidenceItem{}, fmt.Errorf("unknown evidence kind %q", item.EvidenceKind)
	}
	return item, nil
}

func task8CanonicalFutureEvidenceItem(item task8FutureEvidenceItem) ([]byte, error) {
	body, err := json.Marshal(item)
	if err != nil {
		return nil, err
	}
	return jcs.Transform(body)
}

func task8CanonicalFutureEvidenceBundle(bundle task8FutureEvidenceBundle) ([]byte, error) {
	body, err := json.Marshal(bundle)
	if err != nil {
		return nil, err
	}
	return jcs.Transform(body)
}

func task8CompareFutureEvidenceItems(left, right task8FutureEvidenceItem) int {
	leftDigest, _ := hex.DecodeString(left.BodyDigest)
	rightDigest, _ := hex.DecodeString(right.BodyDigest)
	if compared := bytes.Compare(leftDigest, rightDigest); compared != 0 {
		return compared
	}
	return strings.Compare(left.Schema, right.Schema)
}

func task8ReconstructFutureSourceBundles(closureBody []byte, sourceBodies [][]byte) error {
	closure, err := task8ParseFutureEvidenceBundle(closureBody)
	if err != nil {
		return err
	}
	closureItems := make(map[string]task8FutureEvidenceItem, len(closure.Evidence))
	for _, item := range closure.Evidence {
		closureItems[item.BodyDigest+"\x00"+item.Schema] = item
	}
	for index, sourceBody := range sourceBodies {
		source, err := task8ParseFutureEvidenceBundle(sourceBody)
		if err != nil {
			return err
		}
		reconstructed := source
		reconstructed.Evidence = make([]task8FutureEvidenceItem, len(source.Evidence))
		for itemIndex, sourceItem := range source.Evidence {
			closureItem, ok := closureItems[sourceItem.BodyDigest+"\x00"+sourceItem.Schema]
			if !ok {
				return fmt.Errorf("source bundle %d cannot reconstruct member %s/%s", index, sourceItem.BodyDigest, sourceItem.Schema)
			}
			sourceBytes, _ := task8CanonicalFutureEvidenceItem(sourceItem)
			closureBytes, _ := task8CanonicalFutureEvidenceItem(closureItem)
			if !bytes.Equal(sourceBytes, closureBytes) {
				return fmt.Errorf("source bundle %d member bytes conflict", index)
			}
			reconstructed.Evidence[itemIndex] = closureItem
		}
		reconstructedBody, err := task8CanonicalFutureEvidenceBundle(reconstructed)
		if err != nil || !bytes.Equal(reconstructedBody, sourceBody) {
			return fmt.Errorf("source bundle %d did not reconstruct byte-for-byte: %w", index, err)
		}
	}
	return nil
}

func task8HasExactJSONKeys(object map[string]json.RawMessage, want []string) bool {
	if len(object) != len(want) {
		return false
	}
	for _, key := range want {
		if _, ok := object[key]; !ok {
			return false
		}
	}
	return true
}

func task8RequireLowerDigest(value string) error {
	decoded, err := hex.DecodeString(value)
	if err != nil || len(decoded) != sha256.Size || value != strings.ToLower(value) {
		return fmt.Errorf("%q is not lowercase 64-hex", value)
	}
	return nil
}

func task8AssertFutureEvidenceFixtureMutations(t *testing.T, closureBody []byte, sourceBodies [][]byte) {
	t.Helper()
	closure, err := task8ParseFutureEvidenceBundle(closureBody)
	if err != nil || len(closure.Evidence) != 7 || closure.EvidenceCount != "7" {
		t.Fatalf("exact future closure validity/count = %d/%q/%v, want 7/7/nil", len(closure.Evidence), closure.EvidenceCount, err)
	}
	if err := task8ReconstructFutureSourceBundles(closureBody, sourceBodies); err != nil {
		t.Fatal("valid future closure did not reconstruct four source bundles:", err)
	}
	for _, mode := range []string{
		"missing-root", "extra-root", "reordered-roots", "wrong-kind",
		"wrong-schema", "evidence-count", "non-reconstructible",
	} {
		mutated := task8MutateFutureEvidenceClosure(t, closureBody, mode)
		mutatedBundle, parseErr := task8ParseFutureEvidenceBundle(mutated)
		exactCoverErr := parseErr
		reconstructionErr := parseErr
		if parseErr == nil {
			exactCoverErr = task8RequireFutureEvidenceExactCover(closure, mutatedBundle)
			reconstructionErr = task8ReconstructFutureSourceBundles(mutated, sourceBodies)
		}
		if exactCoverErr == nil && reconstructionErr == nil {
			t.Fatalf("fixture structural mutation %q remained parseable, exact-cover, and source-bundle reconstructible", mode)
		}
	}
}

func task8RequireFutureEvidenceExactCover(want, got task8FutureEvidenceBundle) error {
	if len(want.Evidence) != len(got.Evidence) {
		return fmt.Errorf("evidence exact-cover count=%d, want %d", len(got.Evidence), len(want.Evidence))
	}
	for index := range want.Evidence {
		wantBytes, _ := task8CanonicalFutureEvidenceItem(want.Evidence[index])
		gotBytes, _ := task8CanonicalFutureEvidenceItem(got.Evidence[index])
		if !bytes.Equal(wantBytes, gotBytes) {
			return fmt.Errorf("evidence exact-cover item %d differs", index)
		}
	}
	return nil
}

func task8AssertDirectStagingBodyMutationMatrix(t *testing.T, ctx context.Context, pool *pgxpool.Pool, facts contracts.FreshRestoreImportProjectionInputV1) {
	t.Helper()
	wantApplicationFields := []string{
		"acquisition_locked_provider_head_digest", "applied_at", "complete_node_set_digest", "current_database_identity_digest",
		"database_route_closed_digest", "database_timeline_lineage_chain_digest", "database_transaction_id",
		"forbidden_state_zero_digest", "imported_object_count", "manifest_digest", "post_import_inventory_digest",
		"pre_import_inventory_digest", "runtime_instance_binding_digest", "runtime_rebind_chain_digest", "single_use_apply_id",
		"staging_exclusion_lease_digest", "staging_import_capability_digest",
		"staging_import_capability_recovery_application_digest_or_null", "staging_import_capability_recovery_intent_digest_or_null",
		"target_activation_id", "target_database_incarnation_registration_digest", "transaction_snapshot_digest",
	}
	applicationType := reflect.TypeOf(contracts.FreshRestoreImportApplicationV1{})
	gotApplicationFields := make([]string, 0, applicationType.NumField())
	for index := 0; index < applicationType.NumField(); index++ {
		gotApplicationFields = append(gotApplicationFields, applicationType.Field(index).Tag.Get("json"))
	}
	sort.Strings(gotApplicationFields)
	if !reflect.DeepEqual(gotApplicationFields, wantApplicationFields) {
		t.Fatalf("fixed application body fields=%v, want exact 22-field body %v", gotApplicationFields, wantApplicationFields)
	}
	for _, field := range wantApplicationFields {
		task8RunRestrictedStagingReject(t, ctx, pool, facts, "application field "+field, func(arguments *[8][]byte) {
			arguments[6] = task8MutateCanonicalJSONObjectField(t, arguments[6], field)
		})
	}

	wantProjectionFields := []string{"normalized_catalog_digest", "object_count", "objects", "projection_version", "target_activation_id", "target_database_identity_digest", "target_deployment_id"}
	for _, field := range wantProjectionFields {
		task8RunRestrictedStagingReject(t, ctx, pool, facts, "projection field "+field, func(arguments *[8][]byte) {
			arguments[4] = task8MutateCanonicalJSONObjectField(t, arguments[4], field)
		})
	}
	for _, mutation := range []struct {
		name, old, replacement string
	}{
		{"Manifest object order", `{"canonical_key":"54000000-0000-4000-8000-000000000001"`, `{"canonical_key":"fixture-pop"`},
		{"Manifest object kind", `"object_type":"node_reconstruction_seed"`, `"object_type":"pop"`},
		{"Manifest object canonical key", `"canonical_key":"fixture-pop"`, `"canonical_key":"wrong-pop"`},
		{"node identity state", `"identity_state":"unauthorized"`, `"identity_state":"active"`},
		{"node operator state", `"operator_state":"disabled"`, `"operator_state":"enabled"`},
		{"node health state", `"health_state":"unknown"`, `"health_state":"healthy"`},
		{"node lineage generation", `"identity_epoch":"0"`, `"identity_epoch":"1"`},
		{"active pointer closure", `"active_pointer_set":[]`, `"active_pointer_set":[{}]`},
		{"authority anchor closure", `"authority_anchor_set":[]`, `"authority_anchor_set":[{}]`},
	} {
		task8RunRestrictedStagingReject(t, ctx, pool, facts, mutation.name, func(arguments *[8][]byte) {
			if bytes.Count(arguments[4], []byte(mutation.old)) != 1 {
				t.Fatalf("fixed projection mutation %s old literal count !=1", mutation.name)
			}
			arguments[4] = bytes.Replace(arguments[4], []byte(mutation.old), []byte(mutation.replacement), 1)
		})
	}

	for _, mutation := range []struct {
		name, mode string
	}{
		{"seven-root missing root", "missing-root"},
		{"seven-root extra root", "extra-root"},
		{"seven-root reordered roots", "reordered-roots"},
		{"seven-root wrong kind", "wrong-kind"},
		{"seven-root wrong schema", "wrong-schema"},
		{"seven-root count mismatch", "evidence-count"},
		{"source bundle non-reconstructible preimage", "non-reconstructible"},
	} {
		task8RunRestrictedStagingReject(t, ctx, pool, facts, "application evidence "+mutation.name, func(arguments *[8][]byte) {
			arguments[7] = task8MutateFutureEvidenceClosure(t, arguments[7], mutation.mode)
		})
	}
}

func task8MutateCanonicalJSONObjectField(t *testing.T, body []byte, field string) []byte {
	t.Helper()
	var object map[string]json.RawMessage
	if err := json.Unmarshal(body, &object); err != nil {
		t.Fatal(err)
	}
	original, ok := object[field]
	if !ok {
		t.Fatalf("fixed body lacks mutation field %s", field)
	}
	if bytes.Equal(original, []byte("null")) {
		object[field] = json.RawMessage(`"dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"`)
	} else if len(original) > 0 && original[0] == '"' {
		var text string
		if err := json.Unmarshal(original, &text); err != nil {
			t.Fatal(err)
		}
		switch {
		case strings.HasSuffix(field, "_count") || strings.HasSuffix(field, "_version") || strings.HasSuffix(field, "_id") && field == "database_transaction_id":
			text += "1"
		case strings.Contains(field, "_at"):
			text = "2026-08-29T17:00:02.123456Z"
		default:
			if len(text) > 0 {
				last := text[len(text)-1]
				replacement := byte('0')
				if last == '0' {
					replacement = '1'
				}
				text = text[:len(text)-1] + string(replacement)
			}
		}
		encoded, _ := json.Marshal(text)
		object[field] = encoded
	} else if field == "objects" {
		var values []json.RawMessage
		if err := json.Unmarshal(original, &values); err != nil || len(values) != 2 {
			t.Fatalf("decode fixed projection objects: count=%d error=%v", len(values), err)
		}
		values[0], values[1] = values[1], values[0]
		object[field], _ = json.Marshal(values)
	} else {
		object[field] = json.RawMessage(`"mutated"`)
	}
	mutated, err := json.Marshal(object)
	if err != nil {
		t.Fatal(err)
	}
	mutated, err = jcs.Transform(mutated)
	if err != nil {
		t.Fatal(err)
	}
	return mutated
}

func task8MutateFutureEvidenceClosure(t *testing.T, body []byte, mode string) []byte {
	t.Helper()
	var object map[string]json.RawMessage
	if err := json.Unmarshal(body, &object); err != nil {
		t.Fatal(err)
	}
	var items []map[string]json.RawMessage
	if err := json.Unmarshal(object["evidence"], &items); err != nil || len(items) != 7 {
		t.Fatalf("decode seven-root future closure: count=%d error=%v", len(items), err)
	}
	switch mode {
	case "missing-root":
		items = items[1:]
		object["evidence_count"] = json.RawMessage(strconv.Quote(strconv.Itoa(len(items))))
	case "extra-root":
		extra := map[string]json.RawMessage{
			"body_digest":                json.RawMessage(`"47a39ce89c8efea3bfe58f47d1ce1b4ad9c783fee0d78f0b5922f5626261bbce"`),
			"canonical_body_or_null":     json.RawMessage(`{"extra":"not-in-source-registry"}`),
			"canonical_envelope_or_null": json.RawMessage(`null`),
			"evidence_kind":              json.RawMessage(`"database_immutable_body"`),
			"schema":                     json.RawMessage(`"fresh-import-topology-projection.v1"`),
		}
		items = append(items, nil)
		copy(items[3:], items[2:])
		items[2] = extra
		object["evidence_count"] = json.RawMessage(strconv.Quote(strconv.Itoa(len(items))))
	case "reordered-roots":
		items[0], items[1] = items[1], items[0]
	case "wrong-kind":
		items[0]["evidence_kind"] = json.RawMessage(`"database_immutable_body"`)
	case "wrong-schema":
		items[0]["schema"] = json.RawMessage(`"fresh-import-topology-projection.v1"`)
	case "evidence-count":
		object["evidence_count"] = json.RawMessage(`"6"`)
	case "non-reconstructible":
		var envelope map[string]json.RawMessage
		if err := json.Unmarshal(items[0]["canonical_envelope_or_null"], &envelope); err != nil {
			t.Fatal("decode source-required envelope preimage:", err)
		}
		var signature string
		if err := json.Unmarshal(envelope["signature"], &signature); err != nil || signature == "" {
			t.Fatalf("decode source-required signature: %q/%v", signature, err)
		}
		if signature[0] == 'A' {
			signature = "B" + signature[1:]
		} else {
			signature = "A" + signature[1:]
		}
		envelope["signature"], _ = json.Marshal(signature)
		envelopeBytes, err := json.Marshal(envelope)
		if err != nil {
			t.Fatal(err)
		}
		envelopeBytes, err = jcs.Transform(envelopeBytes)
		if err != nil {
			t.Fatal(err)
		}
		items[0]["canonical_envelope_or_null"] = envelopeBytes
	default:
		t.Fatalf("unknown future evidence mutation %q", mode)
	}
	object["evidence"], _ = json.Marshal(items)
	mutated, err := json.Marshal(object)
	if err != nil {
		t.Fatal(err)
	}
	mutated, err = jcs.Transform(mutated)
	if err != nil {
		t.Fatal(err)
	}
	return mutated
}

func task8MutateFixedSourceBundle(t *testing.T, body []byte, mode string) []byte {
	t.Helper()
	var object map[string]json.RawMessage
	if err := json.Unmarshal(body, &object); err != nil {
		t.Fatal(err)
	}
	var items []map[string]json.RawMessage
	if err := json.Unmarshal(object["evidence"], &items); err != nil || len(items) == 0 {
		t.Fatalf("decode fixed source bundle items: count=%d error=%v", len(items), err)
	}
	switch mode {
	case "missing-member":
		items = items[:0]
	case "extra-member":
		items = append(items, items[0])
	case "wrong-kind":
		items[0]["evidence_kind"] = json.RawMessage(`"database_immutable_body"`)
	case "wrong-schema":
		items[0]["schema"] = json.RawMessage(`"fresh-import-topology-projection.v1"`)
	case "non-reconstructible":
		var envelope map[string]json.RawMessage
		if err := json.Unmarshal(items[0]["canonical_envelope_or_null"], &envelope); err != nil {
			t.Fatal(err)
		}
		var signature string
		if err := json.Unmarshal(envelope["signature"], &signature); err != nil || signature == "" {
			t.Fatalf("decode fixed source signature: %q/%v", signature, err)
		}
		signature = "A" + signature[1:]
		envelope["signature"], _ = json.Marshal(signature)
		envelopeBytes, err := json.Marshal(envelope)
		if err != nil {
			t.Fatal(err)
		}
		envelopeBytes, err = jcs.Transform(envelopeBytes)
		if err != nil {
			t.Fatal(err)
		}
		items[0]["canonical_envelope_or_null"] = envelopeBytes
	default:
		t.Fatalf("unknown fixed source mutation %q", mode)
	}
	object["evidence"], _ = json.Marshal(items)
	object["evidence_count"] = json.RawMessage(strconv.Quote(strconv.Itoa(len(items))))
	mutated, err := json.Marshal(object)
	if err != nil {
		t.Fatal(err)
	}
	mutated, err = jcs.Transform(mutated)
	if err != nil {
		t.Fatal(err)
	}
	return mutated
}

func task8RunRestrictedStagingReject(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	facts contracts.FreshRestoreImportProjectionInputV1,
	label string,
	mutate func(*[8][]byte),
) {
	t.Helper()
	before := freshImportResidueSnapshot(t, ctx, pool)
	tx := task8BeginExactStagingPrincipal(t, ctx, pool)
	arguments := task8BuildFutureEightBytea(t, ctx, tx, facts)
	mutate(&arguments)
	var rowCount int
	executionErr := tx.QueryRow(ctx, `SELECT count(*) FROM nodecontrol.begin_staging_import($1::bytea,$2::bytea,$3::bytea,$4::bytea,$5::bytea,$6::bytea,$7::bytea,$8::bytea)`,
		arguments[0], arguments[1], arguments[2], arguments[3], arguments[4], arguments[5], arguments[6], arguments[7]).Scan(&rowCount)
	_ = tx.Rollback(ctx)
	if executionErr != nil || rowCount != 0 {
		t.Errorf("restricted staging %s rows/error = %d/%v, want exact bare-RETURN rejection 0/nil", label, rowCount, executionErr)
	}
	if after := freshImportResidueSnapshot(t, ctx, pool); !bytes.Equal(after, before) {
		t.Fatalf("restricted staging %s changed full residue\nbefore=%s\nafter=%s", label, before, after)
	}
}

func task8NewFreshRestoreImportAdmission(t *testing.T, facts contracts.FreshRestoreImportProjectionInputV1) (authority.VerifiedFreshRestoreImportAdmission, error) {
	return task8NewFreshRestoreImportAdmissionWithPreimages(t, facts, nil)
}

func task8NewFreshRestoreImportAdmissionWithPreimages(t *testing.T, facts contracts.FreshRestoreImportProjectionInputV1, mutate func([][]byte)) (authority.VerifiedFreshRestoreImportAdmission, error) {
	t.Helper()
	factory := reflect.ValueOf(authority.NewFreshRestoreImportAdmissionForIntegration)
	if got := factory.Type().NumIn(); got != 15 {
		t.Errorf("NewFreshRestoreImportAdmissionForIntegration input count = %d, want facts + exact 14 preimages", got)
	}
	preimages := make([][]byte, len(task8FixedStagingPreimages))
	for index := range task8FixedStagingPreimages {
		preimages[index] = append([]byte(nil), task8FixedStagingPreimages[index]...)
	}
	if mutate != nil {
		mutate(preimages)
	}
	arguments := []reflect.Value{reflect.ValueOf(facts)}
	if factory.Type().NumIn() == 15 {
		for _, preimage := range preimages {
			arguments = append(arguments, reflect.ValueOf(preimage))
		}
	} else if factory.Type().NumIn() != 1 {
		return authority.VerifiedFreshRestoreImportAdmission{}, fmt.Errorf("unsupported fresh-import integration factory arity %d", factory.Type().NumIn())
	}
	results := factory.Call(arguments)
	for _, preimage := range preimages {
		preimage[0] ^= 0xff
	}
	admission, ok := results[0].Interface().(authority.VerifiedFreshRestoreImportAdmission)
	if !ok {
		return authority.VerifiedFreshRestoreImportAdmission{}, fmt.Errorf("fresh-import integration factory returned %T", results[0].Interface())
	}
	if !results[1].IsNil() {
		return admission, results[1].Interface().(error)
	}
	return admission, nil
}

func task8AssertFutureStagingFactorySignature(t *testing.T) {
	t.Helper()
	source, err := os.ReadFile("v7_opaque_integration_fixture.go")
	if err != nil {
		t.Fatal("read production integration boundary:", err)
	}
	file, err := parser.ParseFile(token.NewFileSet(), "v7_opaque_integration_fixture.go", source, 0)
	if err != nil {
		t.Fatal("parse production integration boundary:", err)
	}
	wantNames := []string{
		"facts", "capabilityBodyJCS", "capabilityEnvelopeJCS", "capabilityEvidenceBundleJCS",
		"manifestBodyJCS", "manifestEnvelopeJCS", "manifestEvidenceBundleJCS",
		"stagingExclusionLeaseBodyJCS", "stagingExclusionLeaseEnvelopeJCS", "stagingExclusionLeaseEvidenceBundleJCS",
		"providerHeadBodyJCS", "providerHeadEnvelopeJCS", "databaseIncarnationProofBodyJCS",
		"databaseIncarnationProofEnvelopeJCS", "databaseIncarnationProofEvidenceBundleJCS",
	}
	var declaration *ast.FuncDecl
	for _, item := range file.Decls {
		if function, ok := item.(*ast.FuncDecl); ok && function.Name.Name == "NewFreshRestoreImportAdmissionForIntegration" {
			if declaration != nil {
				t.Fatal("production integration boundary declares staging factory more than once")
			}
			declaration = function
		}
	}
	if declaration == nil {
		t.Fatal("production integration boundary lacks NewFreshRestoreImportAdmissionForIntegration")
	}
	gotNames := make([]string, 0, 15)
	gotTypes := make([]string, 0, 15)
	for _, field := range declaration.Type.Params.List {
		for _, name := range field.Names {
			gotNames = append(gotNames, name.Name)
			switch expression := field.Type.(type) {
			case *ast.SelectorExpr:
				packageName, _ := expression.X.(*ast.Ident)
				gotTypes = append(gotTypes, packageName.Name+"."+expression.Sel.Name)
			case *ast.ArrayType:
				if expression.Len != nil {
					gotTypes = append(gotTypes, "fixed-array")
				} else if element, ok := expression.Elt.(*ast.Ident); ok {
					gotTypes = append(gotTypes, "[]"+element.Name)
				}
			case *ast.Ellipsis:
				gotTypes = append(gotTypes, "variadic")
			default:
				gotTypes = append(gotTypes, fmt.Sprintf("%T", field.Type))
			}
		}
	}
	wantTypes := append([]string{"contracts.FreshRestoreImportProjectionInputV1"}, make([]string, 14)...)
	for index := 1; index < len(wantTypes); index++ {
		wantTypes[index] = "[]byte"
	}
	if !reflect.DeepEqual(gotNames, wantNames) || !reflect.DeepEqual(gotTypes, wantTypes) {
		t.Errorf("staging integration factory params names/types=%v/%v, want exact nonvariadic 15-param %v/%v", gotNames, gotTypes, wantNames, wantTypes)
	}
	if declaration.Type.Results == nil || len(declaration.Type.Results.List) != 2 {
		t.Errorf("staging integration factory result arity=%v, want (VerifiedFreshRestoreImportAdmission,error)", declaration.Type.Results)
	}
}

func task8AssertFixedStagingEnvelopeConsistency(t *testing.T) {
	t.Helper()
	if len(task8FixedStagingPreimages) != 14 {
		t.Fatalf("fixed staging preimage count=%d, want exact 14", len(task8FixedStagingPreimages))
	}
	for left := range task8FixedStagingPreimages {
		for right := left + 1; right < len(task8FixedStagingPreimages); right++ {
			if bytes.Equal(task8FixedStagingPreimages[left], task8FixedStagingPreimages[right]) {
				t.Fatalf("fixed staging preimages %d/%d alias identical bytes", left, right)
			}
		}
	}
	seed := [32]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31}
	publicKey := ed25519.NewKeyFromSeed(seed[:]).Public().(ed25519.PublicKey)
	for _, binding := range task8FixedStagingEnvelopeBindings {
		canonical, err := jcs.Transform(binding.envelope)
		if err != nil || !bytes.Equal(canonical, binding.envelope) {
			t.Fatalf("fixed %s envelope is not literal canonical JCS: %v", binding.schema, err)
		}
		var envelope struct {
			Body                   json.RawMessage `json:"body"`
			BodyDigest             string          `json:"body_digest"`
			Schema                 string          `json:"schema"`
			Signature              string          `json:"signature"`
			SignatureAlgorithm     string          `json:"signature_algorithm"`
			SignaturePolicyVersion string          `json:"signature_policy_version"`
			SignerKeyID            string          `json:"signer_key_id"`
			SignerRole             string          `json:"signer_role"`
			TrustRootDigest        string          `json:"trust_root_digest"`
		}
		if err := json.Unmarshal(binding.envelope, &envelope); err != nil {
			t.Fatal(err)
		}
		if envelope.Schema != binding.schema || envelope.SignerRole != binding.role || envelope.SignatureAlgorithm != "ed25519" ||
			envelope.SignaturePolicyVersion != "1" || !bytes.Equal(envelope.Body, binding.body) {
			t.Fatalf("fixed %s envelope metadata/body binding changed", binding.schema)
		}
		bodyDigest := sha256.Sum256(append(append([]byte("talenro.c12."+binding.schema), 0), binding.body...))
		if envelope.BodyDigest != hex.EncodeToString(bodyDigest[:]) {
			t.Fatalf("fixed %s body digest=%s, want literal domain digest %x", binding.schema, envelope.BodyDigest, bodyDigest)
		}
		metadataRaw, err := json.Marshal(map[string]any{
			"body": json.RawMessage(binding.body), "body_digest": envelope.BodyDigest, "schema": envelope.Schema,
			"signature_algorithm": envelope.SignatureAlgorithm, "signature_policy_version": envelope.SignaturePolicyVersion,
			"signer_key_id": envelope.SignerKeyID, "signer_role": envelope.SignerRole, "trust_root_digest": envelope.TrustRootDigest,
		})
		if err != nil {
			t.Fatal(err)
		}
		metadataRaw, err = jcs.Transform(metadataRaw)
		if err != nil {
			t.Fatal(err)
		}
		signature, err := base64.RawURLEncoding.DecodeString(envelope.Signature)
		if err != nil || !ed25519.Verify(publicKey, append(append([]byte("talenro.c12.signature-envelope.v1"), 0), metadataRaw...), signature) {
			t.Fatalf("fixed %s Ed25519 envelope is internally inconsistent: %v", binding.schema, err)
		}
	}
	type expectedMember struct {
		kind, schema, digest string
		body, envelope       []byte
	}
	type expectedSource struct {
		body, schema, digest string
		members              []expectedMember
	}
	null := []byte("null")
	for index, source := range []expectedSource{
		{
			task8FixedCapabilitySourceBundle, "staging-import-capability.v1", task8FixedCapabilityDigestHex,
			[]expectedMember{
				{"external_signed_envelope", "database-incarnation-proof.v1", task8FixedProofDigestHex, null, task8FixedStagingPreimages[12]},
				{"external_signed_envelope", "fresh-restore-import-manifest.v1", task8FixedManifestDigestHex, null, task8FixedStagingPreimages[4]},
			},
		},
		{
			task8FixedManifestSourceBundle, "fresh-restore-import-manifest.v1", task8FixedManifestDigestHex,
			[]expectedMember{
				{"database_immutable_body", "fresh-import-topology-projection.v1", task8FixedFreshProjectionDigestHex, []byte(task8FixedFreshProjectionBody), null},
			},
		},
		{
			task8FixedLeaseSourceBundle, "fresh-v7-staging-exclusion-lease.v1", task8FixedLeaseDigestHex,
			[]expectedMember{
				{"external_signed_envelope", "staging-import-capability.v1", task8FixedCapabilityDigestHex, null, task8FixedStagingPreimages[1]},
				{"external_signed_envelope", "claim-v1-provider-head.v1", task8FixedProviderHeadHex, null, task8FixedStagingPreimages[10]},
			},
		},
		{
			task8FixedProofSourceBundle, "database-incarnation-proof.v1", task8FixedProofDigestHex,
			[]expectedMember{
				{"external_signed_envelope", "claim-v1-provider-head.v1", task8FixedProviderHeadHex, null, task8FixedStagingPreimages[10]},
			},
		},
	} {
		bundle, err := task8ParseFutureEvidenceBundle([]byte(source.body))
		if err != nil || bundle.MessageSchema != source.schema || bundle.MessageBodyDigest != source.digest ||
			bundle.EvidenceCount != strconv.Itoa(len(source.members)) || len(bundle.Evidence) != len(source.members) {
			t.Fatalf("literal source bundle %d binding/count/items=%q/%q/%q/%d error=%v", index, bundle.MessageSchema, bundle.MessageBodyDigest, bundle.EvidenceCount, len(bundle.Evidence), err)
		}
		for memberIndex, expected := range source.members {
			got := bundle.Evidence[memberIndex]
			if got.EvidenceKind != expected.kind || got.Schema != expected.schema || got.BodyDigest != expected.digest ||
				!bytes.Equal(got.CanonicalBodyOrNull, expected.body) || !bytes.Equal(got.CanonicalEnvelopeOrNull, expected.envelope) {
				t.Fatalf("literal source bundle %d member %d kind/schema/digest/body/envelope mismatch", index, memberIndex)
			}
		}
	}
	// This freezes reviewed JCS/domain/signature bytes and Task-1 fixture-owned
	// member lists only. It deliberately does not claim B01 verification or a
	// production registry, pinned-root, role, policy, or key implementation.
}

func freshImportResidueSnapshot(t *testing.T, ctx context.Context, pool *pgxpool.Pool) []byte {
	t.Helper()
	var snapshot []byte
	if err := pool.QueryRow(ctx, `
SELECT convert_to(jsonb_build_object(
	'applications',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY v.single_use_apply_id),'[]'::jsonb) FROM nodecontrol.control_plane_authority_fresh_restore_import_applications v),
	'capability',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY v.capability_id),'[]'::jsonb) FROM nodecontrol.control_plane_authority_staging_import_capabilities v),
	'capability_revocations',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY v.revocation_application_id),'[]'::jsonb) FROM nodecontrol.control_plane_authority_staging_import_capability_revocation_applications v),
	'capability_recovery_intents',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY v.recovery_intent_id),'[]'::jsonb) FROM nodecontrol.control_plane_authority_staging_import_capability_recovery_intents v),
	'capability_recovery_applications',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY v.recovery_application_id),'[]'::jsonb) FROM nodecontrol.control_plane_authority_staging_import_capability_recovery_applications v),
	'pops',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY v.pop_code),'[]'::jsonb) FROM nodecontrol.node_pops v),
	'failure_domains',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY v.failure_domain_id),'[]'::jsonb) FROM nodecontrol.node_failure_domains v),
	'capacity',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY v.profile_id),'[]'::jsonb) FROM nodecontrol.node_capacity_profiles v),
	'inventory',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY v.node_id),'[]'::jsonb) FROM nodecontrol.node_inventory v),
	'latch',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY v.installation_id),'[]'::jsonb) FROM nodecontrol.control_plane_authority_protocol_migration_latches v),
	'upgrade_intents',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY v.intent_id),'[]'::jsonb) FROM nodecontrol.control_plane_authority_protocol_upgrade_intents v),
	'upgrade_attempts',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY v.attempt_id),'[]'::jsonb) FROM nodecontrol.control_plane_authority_protocol_upgrade_attempts v),
	'activations',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY v.activation_id),'[]'::jsonb) FROM nodecontrol.control_plane_authority_protocol_activations v),
	'runtime_registrations',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY to_jsonb(v)::text COLLATE "C"),'[]'::jsonb) FROM nodecontrol.control_plane_authority_runtime_registration_results v),
	'runtime_rebinds',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY to_jsonb(v)::text COLLATE "C"),'[]'::jsonb) FROM nodecontrol.control_plane_authority_runtime_rebind_results v),
	'manifest_sources',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY to_jsonb(v)::text COLLATE "C"),'[]'::jsonb) FROM nodecontrol.control_plane_authority_fresh_restore_requirements v)
)::text,'UTF8')`).Scan(&snapshot); err != nil {
		t.Fatal(err)
	}
	return snapshot
}

func openOwnedFreshImportDatabase(t *testing.T) (*sql.DB, *pgxpool.Pool, fs.FS) {
	t.Helper()
	rawURL := os.Getenv("TALENRO_DATABASE_URL")
	parsed, err := url.Parse(rawURL)
	if err != nil || rawURL == "" {
		t.Fatal("TALENRO_DATABASE_URL is required")
	}
	admin, err := sql.Open("pgx", rawURL)
	if err != nil {
		t.Fatal(err)
	}
	lockConnection, err := admin.Conn(t.Context())
	if err != nil {
		admin.Close()
		t.Fatal(err)
	}
	const fixtureLock int64 = 0x54414c454e524f37
	if _, err := lockConnection.ExecContext(t.Context(), `SELECT pg_advisory_lock($1)`, fixtureLock); err != nil {
		lockConnection.Close()
		admin.Close()
		t.Fatal(err)
	}
	var preexistingRoles int
	if err := lockConnection.QueryRowContext(t.Context(), `SELECT count(*) FROM pg_catalog.pg_roles WHERE rolname=ANY($1::text[])`, []string{
		"nodecontrol_upgrade_executor", "nodecontrol_migration_downgrader", "nodecontrol_staging_importer",
	}).Scan(&preexistingRoles); err != nil || (preexistingRoles != 0 && preexistingRoles != 3) {
		_, _ = lockConnection.ExecContext(context.Background(), `SELECT pg_advisory_unlock($1)`, fixtureLock)
		lockConnection.Close()
		admin.Close()
		t.Fatalf("fresh-import fixture partial capability role count=%d error=%v", preexistingRoles, err)
	}
	fixtureOwnsRoles := preexistingRoles == 0
	databaseName := "nodecontrol_t8_import_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	databaseIdentifier := pgx.Identifier{databaseName}.Sanitize()
	if _, err := admin.ExecContext(t.Context(), "CREATE DATABASE "+databaseIdentifier); err != nil {
		admin.Close()
		t.Fatal(err)
	}
	parsed.Path = "/" + databaseName
	database, err := sql.Open("pgx", parsed.String())
	if err != nil {
		admin.Close()
		t.Fatal(err)
	}
	pool, err := pgxpool.New(t.Context(), parsed.String())
	if err != nil {
		database.Close()
		admin.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		pool.Close()
		_ = database.Close()
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cleanupCancel()
		if _, cleanupErr := admin.ExecContext(cleanupCtx, "DROP DATABASE "+databaseIdentifier+" WITH (FORCE)"); cleanupErr != nil {
			t.Error(cleanupErr)
		}
		if fixtureOwnsRoles {
			if _, cleanupErr := lockConnection.ExecContext(cleanupCtx, `DROP ROLE IF EXISTS nodecontrol_staging_importer,nodecontrol_migration_downgrader,nodecontrol_upgrade_executor`); cleanupErr != nil {
				t.Error(cleanupErr)
			}
			var remainingRoles int
			if cleanupErr := lockConnection.QueryRowContext(cleanupCtx, `SELECT count(*) FROM pg_catalog.pg_roles WHERE rolname=ANY($1::text[])`, []string{
				"nodecontrol_upgrade_executor", "nodecontrol_migration_downgrader", "nodecontrol_staging_importer",
			}).Scan(&remainingRoles); cleanupErr != nil || remainingRoles != 0 {
				t.Errorf("fresh-import fixture capability role residue count=%d error=%v", remainingRoles, cleanupErr)
			}
		}
		if _, cleanupErr := lockConnection.ExecContext(cleanupCtx, `SELECT pg_advisory_unlock($1)`, fixtureLock); cleanupErr != nil {
			t.Error(cleanupErr)
		}
		_ = lockConnection.Close()
		_ = admin.Close()
	})
	repositoryRoot, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return database, pool, os.DirFS(filepath.Join(repositoryRoot, "db", "migrations"))
}

func applyFreshImportV7(t *testing.T, ctx context.Context, database *sql.DB, migrationFS fs.FS) {
	t.Helper()
	provider, err := goose.NewProvider(goose.DialectPostgres, database, migrationFS,
		goose.WithDisableGlobalRegistry(true), goose.WithGoMigrations(migrations.NodeControlAuthorityV7Migration()))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.UpTo(ctx, 6); err != nil {
		t.Fatal("apply v6 base:", err)
	}
	now := time.Now().UTC().Truncate(time.Microsecond)
	digest := func(label string) contracts.Digest { return sha256.Sum256([]byte("fresh-import-up:" + label)) }
	facts := contracts.AuthorityV7UpMigrationFactsV1{
		MigrationLatch: contracts.AuthorityV7MigrationLatchFactsV1{
			InstallationID: uuid.New(), InstallationKind: contracts.AuthorityV7InstallationKindDisposableFixture,
			MigrationVersion: 7, DatabaseIdentityDigest: digest("database"), UpCatalogDigest: digest("catalog"),
			DownState: contracts.AuthorityV7MigrationDownStateLocked, InstalledAt: now,
		},
		MigrationLatchDigest: digest("latch"), AuthorityProtocolProfile: contracts.AuthorityV7ProtocolProfileLegacyV6,
		LocalRuntimeIsolationDigest: digest("runtime"), TransactionNonce: digest("nonce"), ExpiresAt: now.Add(4 * time.Minute),
	}
	grant, err := authority.NewDisposableAuthorityV7UpGrantForIntegration(facts)
	if err != nil {
		t.Fatal(err)
	}
	migrationContext := migrations.WithAuthorityV7MigrationContext(ctx, migrations.AuthorityV7MigrationContext{
		InstallationKind: migrations.InstallationKindDisposableFixture, UpGrant: &grant,
	})
	if _, err := provider.UpTo(migrationContext, 7); err != nil {
		t.Fatal("apply provider-scoped v7:", err)
	}
}

func freshImportAdmissionFixture(t *testing.T, now time.Time) (contracts.FreshRestoreImportProjectionInputV1, contracts.FreshImportTopologyProjectionV1, string, uuid.UUID) {
	t.Helper()
	digest := func(label string) contracts.Digest { return sha256.Sum256([]byte("fresh-import:" + label)) }
	fixedDigest := func(value string) contracts.Digest {
		decoded := mustFreshImportDigestHex(t, value)
		var result contracts.Digest
		copy(result[:], decoded)
		return result
	}
	popCode := "fixture-pop"
	nodeID := uuid.MustParse("54000000-0000-4000-8000-000000000001")
	activationID := uuid.MustParse("54000000-0000-4000-8000-000000000002")
	deploymentID := uuid.MustParse("54000000-0000-4000-8000-000000000003")
	applyID := uuid.MustParse("54000000-0000-4000-8000-000000000004")
	normalizedNode := mustFreshImportJSON(t, contracts.FreshRestoreNormalizedNodePayloadV1{
		NodeID: nodeID, POPCode: popCode, OperatorState: "disabled", SecurityState: "quarantined",
		IdentityState: "unauthorized", HealthState: "unknown", IdentityEpoch: "0", InventoryVersion: "1",
		SecurityVersion: "1", NextDesiredGeneration: "1", NextRecoveryGeneration: "1",
		ActivePointerSet: []json.RawMessage{}, AuthorityAnchorSet: []json.RawMessage{},
	})
	normalizedPOP := mustFreshImportJSON(t, contracts.FreshRestoreNormalizedPOPPayloadV1{
		POPCode: popCode, ISOCountry: "US", Region: "fixture-region", OperatorState: "disabled", Version: "1",
	})
	objects := []contracts.FreshImportTopologyProjectionObjectV1{
		{ObjectType: contracts.FreshRestoreImportObjectNodeReconstructionSeed, CanonicalKey: nodeID.String(), NormalizedPayload: normalizedNode},
		{ObjectType: contracts.FreshRestoreImportObjectPOP, CanonicalKey: popCode, NormalizedPayload: normalizedPOP},
	}
	projection := contracts.FreshImportTopologyProjectionV1{
		ProjectionVersion: 1, TargetActivationID: activationID, TargetDeploymentID: deploymentID,
		TargetDatabaseIdentityDigest: digest("database"), NormalizedCatalogDigest: digest("normalized-catalog"),
		ObjectCount: uint64(len(objects)), Objects: objects,
	}
	projectionDigestBytes, err := hex.DecodeString(task8FixedFreshProjectionDigestHex)
	if err != nil || len(projectionDigestBytes) != sha256.Size {
		t.Fatalf("decode reviewed projection digest: length=%d error=%v", len(projectionDigestBytes), err)
	}
	var projectionDigest contracts.Digest
	copy(projectionDigest[:], projectionDigestBytes)
	completeNodeSet := sha256.Sum256(append(append([]byte("talenro.c12.complete-node-set.v1"), 0), []byte(`["`+nodeID.String()+`"]`)...))
	manifestObjects := []contracts.FreshRestoreImportManifestObjectV1{
		{ObjectType: contracts.FreshRestoreImportObjectNodeReconstructionSeed, CanonicalKey: nodeID.String(), Payload: mustFreshImportJSON(t, contracts.FreshRestoreNodeReconstructionSeedPayloadV1{NodeID: nodeID, POPCode: popCode}), PayloadDigest: digest("node-payload")},
		{ObjectType: contracts.FreshRestoreImportObjectPOP, CanonicalKey: popCode, Payload: mustFreshImportJSON(t, contracts.FreshRestorePOPImportPayloadV1{POPCode: popCode, ISOCountry: "US", Region: "fixture-region"}), PayloadDigest: digest("pop-payload")},
	}
	manifestDigest, incarnation, lineage := fixedDigest(task8FixedManifestDigestHex), digest("incarnation"), digest("lineage")
	runtimeChain, runtimeBinding := digest("runtime-chain"), digest("runtime-binding")
	capabilityDigest := fixedDigest(task8FixedCapabilityDigestHex)
	providerHeadDigest := fixedDigest(task8FixedProviderHeadHex)
	proofDigest := fixedDigest(task8FixedProofDigestHex)
	preProjectionDigest := fixedDigest(task8FixedEmptyFreshProjectionDigestHex)
	exclusionID := uuid.MustParse("54000000-0000-4000-8000-000000000005")
	expires := now.Add(4 * time.Minute)
	capability := contracts.FreshRestoreStagingImportCapabilityFactsV1{
		CapabilityDigest: capabilityDigest, CapabilityID: uuid.MustParse("54000000-0000-4000-8000-000000000006"), SingleUseApplyID: applyID,
		ManifestDigest: manifestDigest, TargetActivationID: activationID, TargetDatabaseIdentityDigest: projection.TargetDatabaseIdentityDigest,
		DatabaseTimelineLineageChainDigest: lineage, TargetDatabaseIncarnationRegistrationDigest: incarnation,
		RuntimeRebindChainDigest: runtimeChain, RuntimeInstanceBindingDigest: runtimeBinding,
		PlannedStagingExclusionID: exclusionID, ExpectedPreAcquireProviderHeadDigest: providerHeadDigest,
		PreAcquireDatabaseIncarnationProofDigest: proofDigest, ProviderPhase: contracts.FreshRestoreProviderPhaseStagingClosed,
		ProviderServingLeaseAbsentDigest: digest("lease-absent"), DatabaseRouteClosedDigest: digest("route-closed"),
		PreImportInventoryDigest: preProjectionDigest, AllowedObjectSetDigest: digest("allowed"),
		ExpectedPostImportInventoryDigest: projectionDigest, TransactionNonce: digest("capability-nonce"), IssuedAt: now, ExpiresAt: expires,
	}
	lease := contracts.FreshV7StagingExclusionLeaseFactsV1{
		LeaseDigest: fixedDigest(task8FixedLeaseDigestHex), ExclusionID: exclusionID, StagingImportCapabilityDigest: capabilityDigest,
		CapabilityRegistrationCommitChallengeDigest: digest("challenge"), CapabilityRegistrationCommitAttestationDigest: digest("attestation"),
		TargetActivationID: activationID, TargetDatabaseIdentityDigest: projection.TargetDatabaseIdentityDigest,
		DatabaseTimelineLineageChainDigest: lineage, ExpectedProviderHeadDigest: providerHeadDigest,
		PreAcquireDatabaseIncarnationProofDigest:    capability.PreAcquireDatabaseIncarnationProofDigest,
		TargetDatabaseIncarnationRegistrationDigest: incarnation, RuntimeRebindChainDigest: runtimeChain,
		RuntimeInstanceBindingDigest: runtimeBinding, RequestedAdmissionExpiresAt: expires, RequestNonce: digest("request-nonce"),
		RequestDigest: digest("request"), ExclusionLeaseID: exclusionID, AcquisitionLockedProviderHeadDigest: providerHeadDigest,
		ProviderServingLeaseAbsentDigest: capability.ProviderServingLeaseAbsentDigest, ProviderControlSequence: 1,
		ExclusionState: contracts.FreshRestoreStagingExclusionHeld, IssuedAt: now, AdmissionExpiresAt: expires,
	}
	facts := contracts.FreshRestoreImportProjectionInputV1{
		ManifestTopology: contracts.FreshRestoreImportManifestTopologyFactsV1{
			ManifestID: uuid.MustParse("54000000-0000-4000-8000-000000000007"), SingleUseApplyID: applyID, ManifestDigest: manifestDigest, TargetActivationID: activationID,
			TargetDeploymentID: deploymentID, TargetDatabaseIncarnationRegistrationDigest: incarnation,
			TargetEpochEvidenceDigest: digest("epoch-evidence"), IssuedAt: now, ExpiresAt: now.Add(10 * time.Minute),
			ObjectCount: uint64(len(manifestObjects)), Objects: manifestObjects, CompleteNodeSetDigest: completeNodeSet,
			ForbiddenObjectClassSetDigest: digest("forbidden-class"), ExpectedPostImportInventoryDigest: projectionDigest,
		},
		StagingImportCapability: capability, StagingExclusionLease: lease, CurrentProviderHeadDigest: lease.AcquisitionLockedProviderHeadDigest,
		ProviderPhase: contracts.FreshRestoreProviderPhaseStagingClosed, ServingLeaseAbsentDigest: capability.ProviderServingLeaseAbsentDigest,
		StagingExclusionState: contracts.FreshRestoreStagingExclusionHeld, DatabaseRouteClosedDigest: capability.DatabaseRouteClosedDigest,
		CurrentDatabaseIdentityDigest: capability.TargetDatabaseIdentityDigest, DatabaseTimelineLineageChainDigest: lineage,
		CurrentDatabaseIncarnationRegistrationDigest: incarnation, RuntimeRebindChainDigest: runtimeChain,
		RuntimeInstanceBindingDigest: runtimeBinding, PreImportInventoryDigest: capability.PreImportInventoryDigest,
		NormalizedCatalogDigest: projection.NormalizedCatalogDigest, AdmissionCheckedAt: now.Add(time.Second),
	}
	if err := facts.Validate(); err != nil || projection.Validate() != nil {
		t.Fatalf("fresh-import fixture invalid: facts=%v projection=%v", err, projection.Validate())
	}
	return facts, projection, popCode, nodeID
}

func mustFreshImportJSON(t *testing.T, value any) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func mustFreshImportDigestHex(t *testing.T, value string) []byte {
	t.Helper()
	digest, err := hex.DecodeString(value)
	if err != nil || len(digest) != sha256.Size {
		t.Fatalf("decode fixed fresh-import digest %q: length=%d error=%v", value, len(digest), err)
	}
	return digest
}
