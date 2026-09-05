package migrations

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gowebpki/jcs"
	"github.com/pressly/goose/v3"
	"talenro.local/platform/internal/nodecontrol/authority"
	"talenro.local/platform/internal/nodecontrol/contracts"
)

type InstallationKind = contracts.AuthorityV7InstallationKindV1

const (
	InstallationKindProduction        = contracts.AuthorityV7InstallationKindProduction
	InstallationKindDisposableFixture = contracts.AuthorityV7InstallationKindDisposableFixture
)

type AuthorityV7MigrationContext struct {
	InstallationKind InstallationKind
	UpGrant          *authority.VerifiedAuthorityV7UpGrant
	DownGrant        *authority.VerifiedAuthorityV7DownGrant
	DownAuthorizer   authority.AuthorityProtocolDowngradeAuthorizer
}

type authorityV7MigrationContextKey struct{}

//go:embed assets/nodecontrol_authority_v7_up.sql
var nodeControlAuthorityV7UpSQL string

//go:embed assets/nodecontrol_authority_v7_down.sql
var nodeControlAuthorityV7DownSQL string

func NodeControlAuthorityV7Migration() *goose.Migration {
	return goose.NewGoMigration(7,
		&goose.GoFunc{RunTx: upNodeControlAuthorityV7},
		&goose.GoFunc{RunTx: downNodeControlAuthorityV7},
	)
}

func WithAuthorityV7MigrationContext(ctx context.Context, input AuthorityV7MigrationContext) context.Context {
	if ctx == nil {
		return nil
	}
	cloned := input
	if input.UpGrant != nil {
		grant := *input.UpGrant
		cloned.UpGrant = &grant
	}
	if input.DownGrant != nil {
		grant := *input.DownGrant
		cloned.DownGrant = &grant
	}
	return context.WithValue(ctx, authorityV7MigrationContextKey{}, cloned)
}

func upNodeControlAuthorityV7(ctx context.Context, tx *sql.Tx) error {
	input, err := requireAuthorityV7MigrationContext(ctx, tx)
	if err != nil || input.UpGrant == nil || input.DownGrant != nil || input.DownAuthorizer != nil {
		return fmt.Errorf("nodecontrol authority v7 Up context: %w", errors.Join(err, errInvalidAuthorityV7MigrationContext))
	}
	facts, err := authority.ConsumeVerifiedAuthorityV7UpGrant(*input.UpGrant)
	if err != nil {
		return fmt.Errorf("consume authority v7 Up grant: %w", err)
	}
	if facts.MigrationLatch.InstallationKind != input.InstallationKind || facts.Validate() != nil {
		return errInvalidAuthorityV7MigrationContext
	}
	if err := executeAuthorityV7Asset(ctx, tx, nodeControlAuthorityV7UpSQL); err != nil {
		return fmt.Errorf("execute authority v7 Up asset: %w", err)
	}
	latchJCS, err := authorityV7MigrationLatchJCS(facts.MigrationLatch)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO nodecontrol.control_plane_authority_protocol_migration_latches (
    installation_id, installation_kind, migration_version, database_identity_digest,
    up_catalog_digest, down_state, installed_at, canonical_body_jcs, body_digest
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		facts.MigrationLatch.InstallationID, string(facts.MigrationLatch.InstallationKind), int64(facts.MigrationLatch.MigrationVersion),
		facts.MigrationLatch.DatabaseIdentityDigest[:], facts.MigrationLatch.UpCatalogDigest[:], string(facts.MigrationLatch.DownState),
		facts.MigrationLatch.InstalledAt.UTC(), latchJCS, facts.MigrationLatchDigest[:]); err != nil {
		return fmt.Errorf("insert authority v7 migration latch: %w", err)
	}
	if facts.UpgradeIntentOrNull == nil {
		return nil
	}
	intent := *facts.UpgradeIntentOrNull
	intentJCS, err := authorityV7UpgradeIntentJCS(intent)
	if err != nil {
		return err
	}
	var credentialPolicyUpdateID any
	if intent.CredentialPolicyUpdateID != nil {
		credentialPolicyUpdateID = *intent.CredentialPolicyUpdateID
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO nodecontrol.control_plane_authority_protocol_upgrade_intents (
    intent_id, installation_id, installation_kind, activation_id, request_nonce,
    credential_policy_update_id_or_null, incarnation_registration_id, provider_absence_proof_id,
    observed_deployment_id, database_identity_digest, local_runtime_isolation_digest,
    classification_state, created_at, canonical_evidence_bundle_jcs, canonical_body_jcs, body_digest
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		intent.IntentID, intent.InstallationID, string(intent.InstallationKind), intent.ActivationID, intent.RequestNonce[:], credentialPolicyUpdateID,
		intent.IncarnationRegistrationID, intent.ProviderAbsenceProofID, intent.ObservedDeploymentID, intent.DatabaseIdentityDigest[:],
		intent.LocalRuntimeIsolationDigest[:], string(intent.ClassificationState), intent.CreatedAt.UTC(), []byte("{}"), intentJCS, (*facts.UpgradeIntentDigestOrNull)[:]); err != nil {
		return fmt.Errorf("insert authority v7 upgrade intent: %w", err)
	}
	return nil
}

func downNodeControlAuthorityV7(ctx context.Context, tx *sql.Tx) error {
	input, err := requireAuthorityV7MigrationContext(ctx, tx)
	if err != nil || input.DownGrant == nil || input.UpGrant != nil || input.DownAuthorizer == nil || input.InstallationKind != InstallationKindDisposableFixture {
		return fmt.Errorf("nodecontrol authority v7 Down context: %w", errors.Join(err, errInvalidAuthorityV7MigrationContext))
	}
	view, err := authority.ConsumeVerifiedAuthorityV7DownGrant(*input.DownGrant)
	if err != nil {
		return fmt.Errorf("consume authority v7 Down grant: %w", err)
	}
	facts := view.Facts
	if facts.Validate() != nil || facts.MigrationLatch.InstallationKind != input.InstallationKind {
		return errInvalidAuthorityV7MigrationContext
	}
	if err := lockAuthorityV7DownCatalog(ctx, tx); err != nil {
		return fmt.Errorf("lock authority v7 Down catalog: %w", err)
	}
	var transactionIDText string
	var transactionStartedAt time.Time
	if err := tx.QueryRowContext(ctx, `SELECT txid_current()::text, transaction_timestamp()`).Scan(&transactionIDText, &transactionStartedAt); err != nil {
		return fmt.Errorf("read authority v7 Down transaction ID: %w", err)
	}
	transactionID, err := strconv.ParseUint(transactionIDText, 10, 64)
	if err != nil || transactionID == 0 {
		return fmt.Errorf("parse authority v7 Down transaction ID: %w", errInvalidAuthorityV7MigrationContext)
	}
	pristine := contracts.AuthorityV7PristineDowngradeInventoryFactsV1{
		InstallationID:              facts.MigrationLatch.InstallationID,
		InstallationKind:            facts.MigrationLatch.InstallationKind,
		DatabaseIdentityDigest:      facts.MigrationLatch.DatabaseIdentityDigest,
		MigrationLatchDigest:        facts.MigrationLatchDigest,
		CurrentCatalogDigest:        facts.CurrentCatalogDigest,
		ManifestID:                  facts.ManifestID,
		ManifestDigest:              facts.ManifestDigest,
		StableTableCount:            facts.StableTableCount,
		StableTableInventory:        append([]contracts.AuthorityV7StableTableInventoryItemV1(nil), facts.StableTableInventory...),
		ControlTableCount:           2,
		NonControlProtocolRowCount:  0,
		MigrationLatchCount:         1,
		DowngradeAuthorizationCount: 0,
		DatabaseTransactionID:       transactionID,
		TransactionNonce:            facts.TransactionNonce,
		Stage:                       contracts.AuthorityV7DowngradeStagePreAuthorization,
		ObservedAt:                  transactionStartedAt.UTC(),
	}
	pristineJCS, pristineDigest, err := authorityV7PristineDowngradeInventoryJCS(pristine)
	if err != nil || pristine.Validate() != nil {
		return fmt.Errorf("construct authority v7 pristine inventory: %w", errInvalidAuthorityV7MigrationContext)
	}
	request := contracts.AuthorityV7DownAuthorizationRequestV1{
		AuthorizationID:             facts.AuthorizationID,
		MigrationLatch:              facts.MigrationLatch,
		MigrationLatchDigest:        facts.MigrationLatchDigest,
		CurrentCatalogDigest:        facts.CurrentCatalogDigest,
		PristineInventory:           pristine,
		PristineInventoryDigest:     pristineDigest,
		ProviderRetirementSet:       facts.ProviderRetirementSet,
		ProviderRetirementSetDigest: facts.ProviderRetirementSetDigest,
		EnvironmentAnchorSetDigest:  facts.EnvironmentAnchorSetDigest,
		ActualDatabaseTransactionID: transactionID,
		TransactionNonce:            facts.TransactionNonce,
		AuthorizationScope:          facts.AuthorizationScope,
	}
	if request.Validate() != nil {
		return errInvalidAuthorityV7MigrationContext
	}
	verified, err := input.DownAuthorizer.AuthorizeAuthorityProtocolDowngrade(ctx, request)
	if err != nil {
		return fmt.Errorf("authorize authority v7 Down: %w", err)
	}
	authorizationView, err := authority.ConsumeVerifiedAuthorityV7DownAuthorization(verified)
	if err != nil {
		return fmt.Errorf("consume authority v7 Down authorization: %w", err)
	}
	authorization := authorizationView.Facts
	if authorization.Validate() != nil || authorization.Request.ActualDatabaseTransactionID != transactionID || authorization.Request.TransactionNonce != facts.TransactionNonce {
		return errInvalidAuthorityV7MigrationContext
	}
	var insertedPristineDigest []byte
	var insertedLatchDigest []byte
	var insertedAuthorizationDigest []byte
	var insertedProviderRetirementDigest []byte
	var insertedTransactionID string
	var insertedNonce []byte
	var insertedLatchCount int64
	var insertedAuthorizationCount int64
	var insertedNonControlCount int64
	var insertedStage string
	var insertedAt time.Time
	if err := tx.QueryRowContext(ctx, `
SELECT * FROM nodecontrol.v7_insert_downgrade_authorization(
    $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20
)`, authorization.Authorization.AuthorizationID, authorization.Authorization.InstallationID,
		authorization.Authorization.MigrationLatchDigest[:], authorization.Authorization.DatabaseIdentityDigest[:], int64(authorization.Authorization.MigrationVersion),
		authorization.Authorization.CurrentCatalogDigest[:], authorization.Authorization.PristineDowngradeInventoryDigest[:],
		authorization.Authorization.ProviderProtocolDowngradeRetirementSetDigest[:], authorization.Authorization.EnvironmentInventoryAnchorSetDigest[:],
		transactionIDText, authorization.Authorization.TransactionNonce[:], string(authorization.Authorization.AuthorizationScope),
		authorization.Authorization.IssuedAt.UTC(), authorization.Authorization.ExpiresAt.UTC(), authorizationView.AuthorizationBodyJCS,
		authorizationView.AuthorizationEnvelopeJCS, authorization.AuthorizationDigest[:], pristineJCS,
		view.ProviderRetirementSetBodyJCS, view.ProviderRetirementEvidenceBundleJCS).Scan(
		&insertedPristineDigest, &insertedLatchDigest, &insertedAuthorizationDigest, &insertedProviderRetirementDigest,
		&insertedTransactionID, &insertedNonce, &insertedLatchCount, &insertedAuthorizationCount,
		&insertedNonControlCount, &insertedStage, &insertedAt,
	); err != nil {
		return fmt.Errorf("insert authority v7 Down authorization: %w", err)
	}
	if !bytes.Equal(insertedPristineDigest, pristineDigest[:]) ||
		!bytes.Equal(insertedLatchDigest, facts.MigrationLatchDigest[:]) ||
		!bytes.Equal(insertedAuthorizationDigest, authorization.AuthorizationDigest[:]) ||
		!bytes.Equal(insertedProviderRetirementDigest, facts.ProviderRetirementSetDigest[:]) ||
		insertedTransactionID != transactionIDText || !bytes.Equal(insertedNonce, facts.TransactionNonce[:]) ||
		insertedLatchCount != 1 || insertedAuthorizationCount != 1 || insertedNonControlCount != 0 ||
		insertedStage != "authorization_inserted" || !insertedAt.Equal(transactionStartedAt) {
		return fmt.Errorf("validate authority v7 Down authorization insert receipt: %w", errInvalidAuthorityV7MigrationContext)
	}
	authorizedState := map[string]any{
		"pristine_downgrade_inventory_digest":               hex.EncodeToString(insertedPristineDigest),
		"migration_latch_digest":                            hex.EncodeToString(insertedLatchDigest),
		"downgrade_authorization_digest":                    hex.EncodeToString(insertedAuthorizationDigest),
		"provider_protocol_downgrade_retirement_set_digest": hex.EncodeToString(insertedProviderRetirementDigest),
		"database_transaction_id":                           insertedTransactionID,
		"transaction_nonce":                                 hex.EncodeToString(insertedNonce),
		"migration_latch_count":                             strconv.FormatInt(insertedLatchCount, 10),
		"downgrade_authorization_count":                     strconv.FormatInt(insertedAuthorizationCount, 10),
		"non_control_protocol_row_count":                    strconv.FormatInt(insertedNonControlCount, 10),
		"stage":                                             insertedStage,
		"observed_at":                                       insertedAt.UTC().Format(time.RFC3339Nano),
	}
	authorizedStateJCS, authorizedStateDigest, err := authorityV7CanonicalBody("pristine-downgrade-authorized-state.v1", authorizedState)
	if err != nil {
		return fmt.Errorf("construct authority v7 authorized state: %w", err)
	}
	var consumedPristineDigest []byte
	var consumedAuthorizedStateDigest []byte
	var consumedLatchDigest []byte
	var consumedAuthorizationDigest []byte
	var consumedTransactionID string
	var consumedNonce []byte
	var postConsumeLatchCount int64
	var postConsumeAuthorizationCount int64
	var postConsumeNonControlCount int64
	var consumedStage string
	var consumedAt time.Time
	if err := tx.QueryRowContext(ctx, `
SELECT * FROM nodecontrol.v7_consume_down_guard($1,$2,$3,$4,$5,$6)`,
		facts.MigrationLatch.InstallationID, facts.MigrationLatchDigest[:], authorization.AuthorizationDigest[:], facts.TransactionNonce[:],
		authorizedStateJCS, authorizedStateDigest[:]).Scan(
		&consumedPristineDigest, &consumedAuthorizedStateDigest, &consumedLatchDigest, &consumedAuthorizationDigest,
		&consumedTransactionID, &consumedNonce, &postConsumeLatchCount, &postConsumeAuthorizationCount,
		&postConsumeNonControlCount, &consumedStage, &consumedAt,
	); err != nil {
		return fmt.Errorf("consume authority v7 Down guard: %w", errors.Join(err, errInvalidAuthorityV7MigrationContext))
	}
	if !bytes.Equal(consumedPristineDigest, pristineDigest[:]) ||
		!bytes.Equal(consumedAuthorizedStateDigest, authorizedStateDigest[:]) ||
		!bytes.Equal(consumedLatchDigest, facts.MigrationLatchDigest[:]) ||
		!bytes.Equal(consumedAuthorizationDigest, authorization.AuthorizationDigest[:]) ||
		consumedTransactionID != transactionIDText || !bytes.Equal(consumedNonce, facts.TransactionNonce[:]) ||
		postConsumeLatchCount != 0 || postConsumeAuthorizationCount != 0 || postConsumeNonControlCount != 0 ||
		consumedStage != "guard_consumed" || !authorityV7DownConsumeTimeValid(transactionStartedAt, consumedAt) {
		return fmt.Errorf("validate authority v7 Down consume receipt: %w", errInvalidAuthorityV7MigrationContext)
	}
	if err := executeAuthorityV7Asset(ctx, tx, nodeControlAuthorityV7DownSQL); err != nil {
		return fmt.Errorf("execute authority v7 Down asset: %w", err)
	}
	return nil
}

var errInvalidAuthorityV7MigrationContext = errors.New("invalid authority v7 migration context")

func requireAuthorityV7MigrationContext(ctx context.Context, tx *sql.Tx) (AuthorityV7MigrationContext, error) {
	if ctx == nil || tx == nil {
		return AuthorityV7MigrationContext{}, errInvalidAuthorityV7MigrationContext
	}
	input, ok := ctx.Value(authorityV7MigrationContextKey{}).(AuthorityV7MigrationContext)
	if !ok || (input.InstallationKind != InstallationKindProduction && input.InstallationKind != InstallationKindDisposableFixture) {
		return AuthorityV7MigrationContext{}, errInvalidAuthorityV7MigrationContext
	}
	return input, nil
}

func executeAuthorityV7Asset(ctx context.Context, tx *sql.Tx, asset string) error {
	for index, statement := range strings.Split(asset, "-- talenro:statement") {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("asset statement %d: %w", index, err)
		}
	}
	return nil
}

func lockAuthorityV7DownCatalog(ctx context.Context, tx *sql.Tx) error {
	const statement = `LOCK TABLE
nodecontrol.control_plane_authority_protocol_migration_latches,
nodecontrol.control_plane_authority_protocol_downgrade_authorizations,
nodecontrol.control_plane_authority_epoch_transition_applications,
nodecontrol.control_plane_authority_epoch_transition_cancellations,
nodecontrol.control_plane_authority_epoch_transition_intents,
nodecontrol.control_plane_authority_epoch_transition_recovery_applications,
nodecontrol.control_plane_authority_epoch_transition_recovery_intents,
nodecontrol.control_plane_authority_epoch_transition_recovery_prefix_decisions,
nodecontrol.control_plane_authority_epoch_transition_resolutions,
nodecontrol.control_plane_authority_epoch_transition_terminal_applications,
nodecontrol.control_plane_authority_fresh_restore_import_applications,
nodecontrol.control_plane_authority_fresh_restore_requirements,
nodecontrol.control_plane_authority_indeterminate_source_seals,
nodecontrol.control_plane_authority_legacy_database_source_retirements,
nodecontrol.control_plane_authority_legacy_source_seals,
nodecontrol.control_plane_authority_protocol_activation_completions,
nodecontrol.control_plane_authority_protocol_activation_releases,
nodecontrol.control_plane_authority_protocol_activations,
nodecontrol.control_plane_authority_protocol_upgrade_attempts,
nodecontrol.control_plane_authority_protocol_upgrade_intents,
nodecontrol.control_plane_authority_runtime_rebind_results,
nodecontrol.control_plane_authority_runtime_registration_results,
nodecontrol.control_plane_authority_staging_import_capabilities,
nodecontrol.control_plane_authority_staging_import_capability_recovery_applications,
nodecontrol.control_plane_authority_staging_import_capability_recovery_intents,
nodecontrol.control_plane_authority_staging_import_capability_revocation_applications,
nodecontrol.control_plane_authority_fences,
nodecontrol.control_plane_trust_bundle_high_waters,
nodecontrol.node_capacity_profiles,
nodecontrol.node_certificate_issuances,
nodecontrol.node_certificates,
nodecontrol.node_desired_states,
nodecontrol.node_endpoints,
nodecontrol.node_enrollment_grants,
nodecontrol.node_failure_domain_membership,
nodecontrol.node_failure_domains,
nodecontrol.node_inventory,
nodecontrol.node_observed_states,
nodecontrol.node_operator_audit,
nodecontrol.node_pops,
nodecontrol.node_process_slots,
nodecontrol.node_recovery_sessions,
nodecontrol.node_recovery_states,
nodecontrol.node_resource_envelopes,
nodecontrol.node_restore_reauthorization_approvals,
nodecontrol.node_root_metadata_publish_intents,
nodecontrol.node_root_metadata_signature_shares,
nodecontrol.node_security_fault_receipts,
nodecontrol.node_security_incidents,
nodecontrol.node_state_signing_intents,
nodecontrol.node_state_transitions,
public.goose_db_version IN ACCESS EXCLUSIVE MODE`
	_, err := tx.ExecContext(ctx, statement)
	return err
}

func authorityV7MigrationLatchJCS(value contracts.AuthorityV7MigrationLatchFactsV1) ([]byte, error) {
	body := map[string]any{
		"installation_id": value.InstallationID.String(), "installation_kind": value.InstallationKind,
		"migration_version": strconv.FormatUint(value.MigrationVersion, 10), "database_identity_digest": hex.EncodeToString(value.DatabaseIdentityDigest[:]),
		"up_catalog_digest": hex.EncodeToString(value.UpCatalogDigest[:]), "down_state": value.DownState, "installed_at": value.InstalledAt.UTC().Format(time.RFC3339Nano),
	}
	canonical, _, err := authorityV7CanonicalBody("authority-protocol-migration-latch.v1", body)
	return canonical, err
}

func authorityV7UpgradeIntentJCS(value contracts.AuthorityV7UpgradeIntentFactsV1) ([]byte, error) {
	var policyID any
	if value.CredentialPolicyUpdateID != nil {
		policyID = value.CredentialPolicyUpdateID.String()
	}
	body := map[string]any{
		"intent_id": value.IntentID.String(), "installation_id": value.InstallationID.String(), "installation_kind": value.InstallationKind,
		"activation_id": value.ActivationID.String(), "request_nonce": hex.EncodeToString(value.RequestNonce[:]), "credential_policy_update_id_or_null": policyID,
		"incarnation_registration_id": value.IncarnationRegistrationID.String(), "provider_absence_proof_id": value.ProviderAbsenceProofID.String(),
		"observed_deployment_id": value.ObservedDeploymentID.String(), "database_identity_digest": hex.EncodeToString(value.DatabaseIdentityDigest[:]),
		"local_runtime_isolation_digest": hex.EncodeToString(value.LocalRuntimeIsolationDigest[:]), "classification_state": value.ClassificationState,
		"created_at": value.CreatedAt.UTC().Format(time.RFC3339Nano),
	}
	canonical, _, err := authorityV7CanonicalBody("authority-protocol-upgrade-intent.v1", body)
	return canonical, err
}

func authorityV7PristineDowngradeInventoryJCS(value contracts.AuthorityV7PristineDowngradeInventoryFactsV1) ([]byte, contracts.Digest, error) {
	inventory := make([]map[string]string, len(value.StableTableInventory))
	for index, item := range value.StableTableInventory {
		inventory[index] = map[string]string{
			"table_name":     item.TableName,
			"classification": item.Classification,
			"row_count":      strconv.FormatUint(item.RowCount, 10),
			"content_digest": hex.EncodeToString(item.ContentDigest[:]),
		}
	}
	body := map[string]any{
		"installation_id":                value.InstallationID.String(),
		"installation_kind":              string(value.InstallationKind),
		"database_identity_digest":       hex.EncodeToString(value.DatabaseIdentityDigest[:]),
		"migration_latch_digest":         hex.EncodeToString(value.MigrationLatchDigest[:]),
		"current_catalog_digest":         hex.EncodeToString(value.CurrentCatalogDigest[:]),
		"manifest_id":                    value.ManifestID.String(),
		"manifest_digest":                hex.EncodeToString(value.ManifestDigest[:]),
		"stable_table_count":             strconv.FormatUint(value.StableTableCount, 10),
		"stable_table_inventory":         inventory,
		"control_table_count":            strconv.FormatUint(value.ControlTableCount, 10),
		"non_control_protocol_row_count": strconv.FormatUint(value.NonControlProtocolRowCount, 10),
		"migration_latch_count":          strconv.FormatUint(value.MigrationLatchCount, 10),
		"downgrade_authorization_count":  strconv.FormatUint(value.DowngradeAuthorizationCount, 10),
		"database_transaction_id":        strconv.FormatUint(value.DatabaseTransactionID, 10),
		"transaction_nonce":              hex.EncodeToString(value.TransactionNonce[:]),
		"stage":                          string(value.Stage),
		"observed_at":                    value.ObservedAt.UTC().Format(time.RFC3339Nano),
	}
	return authorityV7CanonicalBody("pristine-downgrade-inventory.v1", body)
}

func authorityV7CanonicalBody(schema string, value any) ([]byte, contracts.Digest, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, contracts.Digest{}, err
	}
	canonical, err := jcs.Transform(raw)
	if err != nil {
		return nil, contracts.Digest{}, err
	}
	hash := sha256.New()
	_, _ = hash.Write([]byte("talenro.c12." + schema))
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write(canonical)
	var digest contracts.Digest
	copy(digest[:], hash.Sum(nil))
	return canonical, digest, nil
}
