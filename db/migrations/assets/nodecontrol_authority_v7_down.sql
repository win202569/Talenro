-- The registered callback has already taken the fixed ACCESS EXCLUSIVE lock set
-- and completed nodecontrol.v7_consume_down_guard before this asset is entered.

-- talenro:statement
REVOKE EXECUTE ON FUNCTION nodecontrol.v7_acquire_source_freeze_for_seal() FROM CURRENT_USER;
REVOKE EXECUTE ON FUNCTION nodecontrol.v7_insert_downgrade_authorization(uuid,uuid,bytea,bytea,bigint,bytea,bytea,bytea,bytea,numeric,bytea,text,timestamptz,timestamptz,bytea,bytea,bytea,bytea,bytea,bytea) FROM CURRENT_USER;
REVOKE EXECUTE ON FUNCTION nodecontrol.v7_consume_down_guard(uuid,bytea,bytea,bytea,bytea,bytea) FROM CURRENT_USER;
REVOKE EXECUTE ON FUNCTION nodecontrol.begin_staging_import(bytea,bytea,bytea,bytea,bytea,bytea,bytea,jsonb) FROM CURRENT_USER;

-- talenro:statement
REVOKE EXECUTE ON FUNCTION nodecontrol.v7_require_role(name) FROM nodecontrol_upgrade_executor;
REVOKE USAGE ON SCHEMA nodecontrol FROM nodecontrol_upgrade_executor;

-- talenro:statement
REVOKE EXECUTE ON FUNCTION nodecontrol.v7_require_role(name) FROM nodecontrol_staging_importer;
REVOKE INSERT ON
  nodecontrol.node_pops,
  nodecontrol.node_failure_domains,
  nodecontrol.node_capacity_profiles,
  nodecontrol.node_inventory,
  nodecontrol.control_plane_authority_fresh_restore_import_applications
FROM nodecontrol_staging_importer;
REVOKE UPDATE (body_digest) ON nodecontrol.control_plane_authority_staging_import_capabilities FROM nodecontrol_staging_importer;
REVOKE SELECT ON
  nodecontrol.control_plane_authority_staging_import_capabilities,
  nodecontrol.control_plane_authority_staging_import_capability_recovery_intents,
  nodecontrol.control_plane_authority_staging_import_capability_recovery_applications,
  nodecontrol.control_plane_authority_staging_import_capability_revocation_applications,
  nodecontrol.control_plane_authority_fresh_restore_import_applications
FROM nodecontrol_staging_importer;
REVOKE USAGE ON SCHEMA nodecontrol FROM nodecontrol_staging_importer;

-- talenro:statement
REVOKE EXECUTE ON FUNCTION nodecontrol.v7_require_role(name) FROM nodecontrol_migration_downgrader;
REVOKE DELETE ON nodecontrol.control_plane_authority_protocol_migration_latches FROM nodecontrol_migration_downgrader;
REVOKE INSERT, DELETE ON nodecontrol.control_plane_authority_protocol_downgrade_authorizations FROM nodecontrol_migration_downgrader;
REVOKE MAINTAIN ON
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
  public.goose_db_version
FROM nodecontrol_migration_downgrader;
REVOKE SELECT ON
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
  nodecontrol.node_state_transitions
FROM nodecontrol_migration_downgrader;
REVOKE USAGE ON SCHEMA nodecontrol, public FROM nodecontrol_migration_downgrader;

-- talenro:statement
DROP FUNCTION nodecontrol.begin_staging_import(bytea,bytea,bytea,bytea,bytea,bytea,bytea,jsonb);

-- talenro:statement
DROP FUNCTION nodecontrol.v7_consume_down_guard(uuid,bytea,bytea,bytea,bytea,bytea);

-- talenro:statement
DROP FUNCTION nodecontrol.v7_insert_downgrade_authorization(uuid,uuid,bytea,bytea,bigint,bytea,bytea,bytea,bytea,numeric,bytea,text,timestamptz,timestamptz,bytea,bytea,bytea,bytea,bytea,bytea);

-- talenro:statement
DROP FUNCTION nodecontrol.v7_acquire_source_freeze_for_seal();

-- talenro:statement
DROP TRIGGER ncv7_fence_owner_closure ON nodecontrol.control_plane_authority_fences;

-- talenro:statement
DROP TRIGGER ncv7_00_fence_source_lock ON nodecontrol.control_plane_authority_fences;

-- talenro:statement
DO $source_guards$
DECLARE
  relation_name text;
BEGIN
  FOREACH relation_name IN ARRAY ARRAY[
    'control_plane_trust_bundle_high_waters','node_state_transitions','node_operator_audit',
    'node_observed_states','node_recovery_states','node_desired_states',
    'node_root_metadata_signature_shares','node_root_metadata_publish_intents',
    'node_state_signing_intents','node_restore_reauthorization_approvals',
    'node_recovery_sessions','node_security_fault_receipts','node_security_incidents',
    'node_certificates','node_enrollment_grants','node_certificate_issuances',
    'node_resource_envelopes','node_process_slots','node_endpoints',
    'node_failure_domain_membership','node_inventory','node_capacity_profiles',
    'node_failure_domains','node_pops'
  ] LOOP
    EXECUTE format('DROP TRIGGER ncv7_01_source_guard ON nodecontrol.%I', relation_name);
    EXECUTE format('DROP TRIGGER ncv7_00_source_lock ON nodecontrol.%I', relation_name);
  END LOOP;
END
$source_guards$;

-- talenro:statement
DO $proof_triggers$
DECLARE
  relation_name text;
BEGIN
  FOREACH relation_name IN ARRAY ARRAY['node_root_metadata_publish_intents','node_state_signing_intents','node_resource_envelopes','node_security_incidents','node_state_transitions','node_certificates','node_certificate_issuances','node_enrollment_grants'] LOOP
    EXECUTE format('DROP TRIGGER %I ON nodecontrol.%I', 'ncv7_' || relation_name || '_proof_guard', relation_name);
  END LOOP;
END
$proof_triggers$;

-- talenro:statement
DROP TRIGGER node_resource_envelopes_immutable ON nodecontrol.node_resource_envelopes;
CREATE TRIGGER node_resource_envelopes_immutable
BEFORE UPDATE OR DELETE ON nodecontrol.node_resource_envelopes
FOR EACH ROW EXECUTE FUNCTION nodecontrol.reject_row_mutation();

-- talenro:statement
CREATE TRIGGER node_state_transitions_immutable
BEFORE UPDATE OR DELETE ON nodecontrol.node_state_transitions
FOR EACH ROW EXECUTE FUNCTION nodecontrol.reject_row_mutation();

-- talenro:statement
ALTER TABLE nodecontrol.node_enrollment_grants
  DROP CONSTRAINT ncv7_grant_claim_lifecycle_ck,
  DROP CONSTRAINT ncv7_grant_consumption_result_ck,
  DROP CONSTRAINT ncv7_grant_claim_authority_tuple_ck,
  ADD CONSTRAINT node_enrollment_grants_consumption_all_or_none CHECK (num_nonnulls(consumed_at,consumption_attempt_id,consumption_request_digest,claim_authority_operation_id,claim_authority_epoch,claim_authority_sequence,result_issuance_id) IN (0,7));

-- talenro:statement
ALTER TABLE nodecontrol.node_certificates
  DROP CONSTRAINT ncv7_certificate_revoke_lifecycle_ck,
  DROP CONSTRAINT ncv7_certificate_revoke_result_ck,
  DROP CONSTRAINT ncv7_certificate_revoke_authority_tuple_ck,
  ADD CONSTRAINT node_certificates_revoke_all_or_none CHECK (num_nonnulls(revoke_authority_operation_id,revoke_authority_epoch,revoke_authority_sequence,revoked_at,revoke_reason) IN (0,5));

-- talenro:statement
ALTER TABLE nodecontrol.node_security_incidents
  DROP CONSTRAINT ncv7_incident_resolution_lifecycle_ck,
  DROP CONSTRAINT ncv7_incident_resolution_result_ck,
  DROP CONSTRAINT ncv7_incident_resolution_authority_tuple_ck,
  ADD CONSTRAINT node_security_incidents_resolution_all_or_none CHECK (num_nonnulls(resolution_authority_operation_id,resolution_authority_epoch,resolution_authority_sequence,remediation_digest,resolution_at) IN (0,5));

-- talenro:statement
CREATE OR REPLACE FUNCTION nodecontrol.enforce_enrollment_grant_workflow()
RETURNS trigger
LANGUAGE plpgsql
AS $fn$
BEGIN
  IF TG_OP = 'INSERT' THEN
    IF num_nonnulls(NEW.consumed_at,NEW.consumption_attempt_id,NEW.consumption_request_digest,
                    NEW.claim_authority_operation_id,NEW.claim_authority_epoch,NEW.claim_authority_sequence,
                    NEW.result_issuance_id,NEW.expired_at,NEW.invalidated_at,NEW.terminal_reason,
                    NEW.terminal_at,NEW.retention_until) <> 0 THEN
      RAISE EXCEPTION 'enrollment grants must start live and unterminated' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
  END IF;
  IF TG_OP = 'DELETE' THEN
    IF OLD.retention_until IS NULL OR OLD.retention_until > transaction_timestamp() THEN
      RAISE EXCEPTION 'enrollment grant retention has not elapsed' USING ERRCODE = '23514';
    END IF;
    RETURN OLD;
  END IF;
  IF num_nonnulls(OLD.consumed_at,OLD.consumption_attempt_id,OLD.consumption_request_digest,
                  OLD.claim_authority_operation_id,OLD.claim_authority_epoch,OLD.claim_authority_sequence,
                  OLD.result_issuance_id,OLD.expired_at,OLD.invalidated_at,OLD.terminal_reason,
                  OLD.terminal_at,OLD.retention_until) > 0 THEN
    RAISE EXCEPTION 'terminal enrollment grant is immutable' USING ERRCODE = '23514';
  END IF;
  IF (NEW.grant_id,NEW.authority_operation_id,NEW.authority_epoch,NEW.authority_sequence,NEW.node_id,
      NEW.identity_epoch,NEW.token_digest,NEW.csr_digest,NEW.idempotency_digest,NEW.created_at,NEW.expires_at)
     IS DISTINCT FROM
     (OLD.grant_id,OLD.authority_operation_id,OLD.authority_epoch,OLD.authority_sequence,OLD.node_id,
      OLD.identity_epoch,OLD.token_digest,OLD.csr_digest,OLD.idempotency_digest,OLD.created_at,OLD.expires_at) THEN
    RAISE EXCEPTION 'enrollment grant inputs are immutable' USING ERRCODE = '23514';
  END IF;
  RETURN NEW;
END
$fn$;

-- talenro:statement
CREATE OR REPLACE FUNCTION nodecontrol.enforce_security_incident_workflow()
RETURNS trigger
LANGUAGE plpgsql
SECURITY INVOKER
AS $fn$
DECLARE
  open_count integer;
BEGIN
  IF TG_OP <> 'DELETE' AND pg_catalog.current_setting('transaction_isolation') <> 'read committed' THEN
    RAISE EXCEPTION 'nodecontrol security incident mutation requires read committed isolation' USING ERRCODE = '25001';
  END IF;
  IF TG_OP = 'INSERT' THEN
    PERFORM 1 FROM nodecontrol.node_inventory WHERE node_id = NEW.node_id FOR UPDATE;
    SELECT count(*) INTO open_count
    FROM nodecontrol.node_security_incidents
    WHERE node_id = NEW.node_id
      AND status IN ('open','resolution_pending_agent_ack','overflow');
    IF open_count >= 16 THEN
      RAISE EXCEPTION 'node security incident cap reached' USING ERRCODE = '23514';
    END IF;
    IF (NEW.fault_subtype = 'incident_overflow' AND NEW.status <> 'overflow')
       OR (NEW.fault_subtype <> 'incident_overflow' AND NEW.status <> 'open') THEN
      RAISE EXCEPTION 'security incident must start in its nonterminal opening status' USING ERRCODE = '23514';
    END IF;
    IF NOT EXISTS (
      SELECT 1 FROM nodecontrol.control_plane_authority_fences fence
      WHERE (fence.operation_id,fence.authority_epoch,fence.authority_sequence) =
            (NEW.authority_operation_id,NEW.authority_epoch,NEW.authority_sequence)
        AND fence.provider_status = 'committed'
        AND fence.visibility_state = 'active'
        AND fence.effect_kind = 'security_incident_open'
        AND fence.scope_kind = 'node'
    ) THEN
      RAISE EXCEPTION 'security incident opening requires its committed node fence' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
  END IF;
  IF TG_OP = 'DELETE' THEN
    IF OLD.status <> 'resolved' OR OLD.retention_until > transaction_timestamp() THEN
      RAISE EXCEPTION 'security incident retention has not elapsed' USING ERRCODE = '23514';
    END IF;
    RETURN OLD;
  END IF;
  PERFORM 1 FROM nodecontrol.node_inventory WHERE node_id = NEW.node_id FOR UPDATE;
  IF NOT FOUND THEN
    RAISE EXCEPTION 'unknown node for security incident' USING ERRCODE = '23503';
  END IF;
  IF OLD.status = 'resolved' THEN
    RAISE EXCEPTION 'resolved security incident is immutable' USING ERRCODE = '23514';
  END IF;
  IF (NEW.incident_id,NEW.authority_operation_id,NEW.authority_epoch,NEW.authority_sequence,
      NEW.node_id,NEW.identity_epoch,NEW.fault_subtype,NEW.subtype_slot,NEW.first_evidence_digest,
      NEW.trust_context_digest,NEW.first_occurred_at)
     IS DISTINCT FROM
     (OLD.incident_id,OLD.authority_operation_id,OLD.authority_epoch,OLD.authority_sequence,
      OLD.node_id,OLD.identity_epoch,OLD.fault_subtype,OLD.subtype_slot,OLD.first_evidence_digest,
      OLD.trust_context_digest,OLD.first_occurred_at) THEN
    RAISE EXCEPTION 'security incident identity is immutable' USING ERRCODE = '23514';
  END IF;
  IF NEW.occurrence_count < OLD.occurrence_count OR NEW.last_occurred_at < OLD.last_occurred_at THEN
    RAISE EXCEPTION 'security incident evidence cannot move backward' USING ERRCODE = '23514';
  END IF;
  IF OLD.resolution_at IS NOT NULL
     AND (NEW.resolution_authority_operation_id,NEW.resolution_authority_epoch,
          NEW.resolution_authority_sequence,NEW.remediation_digest,NEW.resolution_at)
         IS DISTINCT FROM
         (OLD.resolution_authority_operation_id,OLD.resolution_authority_epoch,
          OLD.resolution_authority_sequence,OLD.remediation_digest,OLD.resolution_at) THEN
    RAISE EXCEPTION 'security incident resolution binding is immutable' USING ERRCODE = '23514';
  END IF;
  IF NEW.resolution_at IS NOT NULL AND NOT EXISTS (
    SELECT 1 FROM nodecontrol.control_plane_authority_fences fence
    WHERE (fence.operation_id,fence.authority_epoch,fence.authority_sequence) =
          (NEW.resolution_authority_operation_id,NEW.resolution_authority_epoch,NEW.resolution_authority_sequence)
      AND fence.provider_status = 'committed'
      AND fence.visibility_state = 'active'
      AND fence.effect_kind = 'security_incident_resolve'
      AND fence.scope_kind = 'node'
  ) THEN
    RAISE EXCEPTION 'security incident resolution requires its committed node fence' USING ERRCODE = '23514';
  END IF;
  IF NEW.status = 'resolved'
     AND EXISTS (
       SELECT 1 FROM nodecontrol.node_security_fault_receipts
       WHERE incident_id = OLD.incident_id AND binding_status = 'active'
     ) THEN
    RAISE EXCEPTION 'active security-fault bindings prevent direct resolution' USING ERRCODE = '23514';
  END IF;
  IF NEW.status IS DISTINCT FROM OLD.status AND NOT (
    (OLD.status IN ('open','overflow') AND NEW.status IN ('resolution_pending_agent_ack','resolved')) OR
    (OLD.status = 'resolution_pending_agent_ack' AND NEW.status = 'resolved')
  ) THEN
    RAISE EXCEPTION 'illegal security incident status transition' USING ERRCODE = '23514';
  END IF;
  RETURN NEW;
END
$fn$;
ALTER FUNCTION nodecontrol.enforce_security_incident_workflow() RESET ALL;

-- talenro:statement
CREATE OR REPLACE FUNCTION nodecontrol.enforce_authority_fence_update()
RETURNS trigger
LANGUAGE plpgsql
SECURITY INVOKER
AS $fn$
BEGIN
  IF TG_OP = 'INSERT' THEN
    IF NEW.provider_status <> 'reserved'
       OR NEW.visibility_state <> 'fence_pending'
       OR NEW.effect_digest IS NOT NULL
       OR NEW.provider_receipt_digest IS NOT NULL
       OR NEW.db_system_id IS NOT NULL
       OR NEW.db_timeline IS NOT NULL
       OR NEW.required_lsn IS NOT NULL
       OR NEW.abort_reason IS NOT NULL
       OR NEW.effect_bound_at IS NOT NULL
       OR NEW.terminal_at IS NOT NULL THEN
      RAISE EXCEPTION 'authority fences must begin as unbound reservations' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
  END IF;

  IF TG_OP = 'DELETE' THEN
    RAISE EXCEPTION 'authority fence rows are append-only' USING ERRCODE = '23514';
  END IF;

  IF NEW.operation_id IS DISTINCT FROM OLD.operation_id
     OR NEW.effect_kind IS DISTINCT FROM OLD.effect_kind
     OR NEW.scope_kind IS DISTINCT FROM OLD.scope_kind
     OR NEW.scope_digest IS DISTINCT FROM OLD.scope_digest
     OR NEW.provider_reservation_digest IS DISTINCT FROM OLD.provider_reservation_digest
     OR NEW.authority_epoch IS DISTINCT FROM OLD.authority_epoch
     OR NEW.authority_sequence IS DISTINCT FROM OLD.authority_sequence
     OR NEW.reserved_at IS DISTINCT FROM OLD.reserved_at THEN
    RAISE EXCEPTION 'authority fence reservation identity is immutable' USING ERRCODE = '23514';
  END IF;

  IF OLD.provider_status <> 'reserved' OR OLD.terminal_at IS NOT NULL THEN
    RAISE EXCEPTION 'terminal authority fence rows are immutable' USING ERRCODE = '23514';
  END IF;

  IF OLD.effect_digest IS NOT NULL
     AND (NEW.effect_digest, NEW.db_system_id, NEW.db_timeline, NEW.required_lsn, NEW.effect_bound_at)
         IS DISTINCT FROM
         (OLD.effect_digest, OLD.db_system_id, OLD.db_timeline, OLD.required_lsn, OLD.effect_bound_at) THEN
    RAISE EXCEPTION 'authority fence database binding cannot change' USING ERRCODE = '23514';
  END IF;

  IF NEW.provider_status IS DISTINCT FROM OLD.provider_status
     AND NOT (OLD.provider_status = 'reserved' AND NEW.provider_status IN ('committed','aborted')) THEN
    RAISE EXCEPTION 'illegal authority fence status transition' USING ERRCODE = '23514';
  END IF;

  RETURN NEW;
END
$fn$;
ALTER FUNCTION nodecontrol.enforce_authority_fence_update() RESET ALL;

-- talenro:statement
ALTER TABLE nodecontrol.control_plane_authority_fresh_restore_import_applications DROP CONSTRAINT ncv7_fi_fk01;
ALTER TABLE nodecontrol.control_plane_authority_staging_import_capability_recovery_applications DROP CONSTRAINT ncv7_sa_fk02, DROP CONSTRAINT ncv7_sa_fk01;
ALTER TABLE nodecontrol.control_plane_authority_epoch_transition_recovery_applications DROP CONSTRAINT ncv7_ra_fk02, DROP CONSTRAINT ncv7_ra_fk01;
ALTER TABLE nodecontrol.control_plane_authority_epoch_transition_recovery_prefix_decisions DROP CONSTRAINT ncv7_pd_fk01;
ALTER TABLE nodecontrol.control_plane_authority_epoch_transition_terminal_applications DROP CONSTRAINT ncv7_et_fk03, DROP CONSTRAINT ncv7_et_fk02, DROP CONSTRAINT ncv7_et_fk01;
ALTER TABLE nodecontrol.control_plane_authority_epoch_transition_cancellations DROP CONSTRAINT ncv7_ec_fk02, DROP CONSTRAINT ncv7_ec_fk01;
ALTER TABLE nodecontrol.control_plane_authority_epoch_transition_resolutions DROP CONSTRAINT ncv7_er_fk02, DROP CONSTRAINT ncv7_er_fk01;
ALTER TABLE nodecontrol.control_plane_authority_epoch_transition_applications DROP CONSTRAINT ncv7_ea_fk01;
ALTER TABLE nodecontrol.control_plane_authority_epoch_transition_intents DROP CONSTRAINT ncv7_ei_fk01;
ALTER TABLE nodecontrol.control_plane_authority_protocol_activation_releases DROP CONSTRAINT ncv7_pr_fk02, DROP CONSTRAINT ncv7_pr_fk01;
ALTER TABLE nodecontrol.control_plane_authority_protocol_activation_completions DROP CONSTRAINT ncv7_pc_fk01;
ALTER TABLE nodecontrol.control_plane_authority_protocol_activations DROP CONSTRAINT ncv7_pa_fk03, DROP CONSTRAINT ncv7_pa_fk02, DROP CONSTRAINT ncv7_pa_fk01;
ALTER TABLE nodecontrol.control_plane_authority_protocol_upgrade_attempts DROP CONSTRAINT ncv7_ua_fk03, DROP CONSTRAINT ncv7_ua_fk02, DROP CONSTRAINT ncv7_ua_fk01;
ALTER TABLE nodecontrol.control_plane_authority_runtime_rebind_results DROP CONSTRAINT ncv7_rb_fk02, DROP CONSTRAINT ncv7_rb_fk01;
ALTER TABLE nodecontrol.control_plane_authority_runtime_registration_results DROP CONSTRAINT ncv7_rr_fk01;

-- talenro:statement
DO $proof_columns$
DECLARE
  relation_name text;
  prefix_name text;
  suffix_name text;
  relation_names text[] := ARRAY['node_root_metadata_publish_intents','node_state_signing_intents','node_resource_envelopes','node_security_incidents','node_security_incidents','node_state_transitions','node_certificates','node_certificate_issuances','node_enrollment_grants','node_enrollment_grants'];
  prefix_names text[] := ARRAY['activation_','activation_','activation_','resolve_','open_','','revoke_','activation_','claim_','create_'];
  suffix_names text[] := ARRAY['authority_effect_resolution_digest','authority_effect_resolution_jcs','authority_activation_evidence_digest','authority_activation_evidence_jcs','authority_expected_provider_identity_digest','authority_activation_deadline','authority_attestation_expires_at','authority_effect_reason','authority_checkpoint_anchor_digest','authority_checkpoint_anchor_jcs','authority_provider_head_digest','authority_provider_head_jcs','authority_effect_commitment_digest','authority_effect_commitment_jcs'];
  i integer;
BEGIN
  ALTER TABLE nodecontrol.node_state_transitions
    DROP CONSTRAINT ncv7_transition_operation_uq;
  FOR i IN 1..array_length(relation_names, 1) LOOP
    relation_name := relation_names[i];
    prefix_name := prefix_names[i];
    FOREACH suffix_name IN ARRAY suffix_names LOOP
      EXECUTE format('ALTER TABLE nodecontrol.%I DROP COLUMN %I', relation_name, prefix_name || suffix_name);
    END LOOP;
  END LOOP;
  ALTER TABLE nodecontrol.node_state_transitions
    DROP COLUMN authority_effect_disposition,
    DROP COLUMN authority_effect_kind;
END
$proof_columns$;

-- talenro:statement
ALTER TABLE nodecontrol.node_inventory
  DROP CONSTRAINT ncv7_inventory_identity_state_ck,
  DROP CONSTRAINT ncv7_inventory_identity_lineage_ck,
  DROP CONSTRAINT ncv7_inventory_quarantine_ck,
  ADD CONSTRAINT node_inventory_identity_state_enum CHECK (identity_state IN ('never_enrolled','active','recovery_pending','recovery_limited','revoked')),
  ADD CONSTRAINT node_inventory_identity_lineage_state CHECK ((identity_state = 'never_enrolled' AND identity_epoch = 0 AND lineage_id IS NULL) OR (identity_state <> 'never_enrolled' AND identity_epoch > 0 AND lineage_id IS NOT NULL)),
  ADD CONSTRAINT node_inventory_quarantine_state CHECK ((security_state = 'normal' AND resume_operator_state IS NULL) OR (security_state = 'quarantined' AND operator_state = 'disabled' AND resume_operator_state IS NOT NULL));

-- talenro:statement
DROP INDEX nodecontrol.ncv7_fence_activation_fk_ix;
ALTER TABLE nodecontrol.control_plane_authority_fences
  DROP CONSTRAINT ncv7_fence_activation_fk,
  DROP CONSTRAINT ncv7_fence_claim_ck,
  DROP CONSTRAINT ncv7_fence_profile_ck,
  DROP COLUMN protocol_activation_id,
  DROP COLUMN abort_claimed_at,
  DROP COLUMN authority_protocol_profile,
  ADD CONSTRAINT control_plane_authority_fences_abort_reason_pair CHECK ((provider_status = 'aborted') = (abort_reason IS NOT NULL));

-- talenro:statement
DROP TABLE nodecontrol.control_plane_authority_fresh_restore_import_applications;
-- talenro:statement
DROP TABLE nodecontrol.control_plane_authority_staging_import_capability_revocation_applications;
-- talenro:statement
DROP TABLE nodecontrol.control_plane_authority_staging_import_capability_recovery_applications;
-- talenro:statement
DROP TABLE nodecontrol.control_plane_authority_staging_import_capability_recovery_intents;
-- talenro:statement
DROP TABLE nodecontrol.control_plane_authority_staging_import_capabilities;
-- talenro:statement
DROP TABLE nodecontrol.control_plane_authority_protocol_activation_releases;
-- talenro:statement
DROP TABLE nodecontrol.control_plane_authority_protocol_activation_completions;
-- talenro:statement
DROP TABLE nodecontrol.control_plane_authority_epoch_transition_recovery_applications;
-- talenro:statement
DROP TABLE nodecontrol.control_plane_authority_epoch_transition_recovery_prefix_decisions;
-- talenro:statement
DROP TABLE nodecontrol.control_plane_authority_epoch_transition_recovery_intents;
-- talenro:statement
DROP TABLE nodecontrol.control_plane_authority_epoch_transition_terminal_applications;
-- talenro:statement
DROP TABLE nodecontrol.control_plane_authority_epoch_transition_cancellations;
-- talenro:statement
DROP TABLE nodecontrol.control_plane_authority_epoch_transition_resolutions;
-- talenro:statement
DROP TABLE nodecontrol.control_plane_authority_epoch_transition_applications;
-- talenro:statement
DROP TABLE nodecontrol.control_plane_authority_epoch_transition_intents;
-- talenro:statement
DROP TABLE nodecontrol.control_plane_authority_runtime_rebind_results;
-- talenro:statement
DROP TABLE nodecontrol.control_plane_authority_protocol_activations;
-- talenro:statement
DROP TABLE nodecontrol.control_plane_authority_protocol_upgrade_attempts;
-- talenro:statement
DROP TABLE nodecontrol.control_plane_authority_runtime_registration_results;
-- talenro:statement
DROP TABLE nodecontrol.control_plane_authority_fresh_restore_requirements;
-- talenro:statement
DROP TABLE nodecontrol.control_plane_authority_legacy_source_seals;
-- talenro:statement
DROP TABLE nodecontrol.control_plane_authority_indeterminate_source_seals;
-- talenro:statement
DROP TABLE nodecontrol.control_plane_authority_legacy_database_source_retirements;
-- talenro:statement
DROP TABLE nodecontrol.control_plane_authority_protocol_upgrade_intents;
-- talenro:statement
DROP TABLE nodecontrol.control_plane_authority_protocol_downgrade_authorizations;
-- talenro:statement
DROP TABLE nodecontrol.control_plane_authority_protocol_migration_latches;

-- talenro:statement
DROP FUNCTION nodecontrol.v7_guard_authority_proof_transition();
-- talenro:statement
DROP FUNCTION nodecontrol.v7_authority_proof_group_valid(bytea,bytea,bytea,bytea,bytea,bytea,text,timestamptz,timestamptz,bytea,bytea,bytea,bytea,bytea);
-- talenro:statement
DROP FUNCTION nodecontrol.v7_assert_activation_barrier();
-- talenro:statement
DROP FUNCTION nodecontrol.v7_assert_source_writable();
-- talenro:statement
DROP FUNCTION nodecontrol.v7_source_is_frozen();
-- talenro:statement
DROP FUNCTION nodecontrol.v7_reject_immutable_mutation();
-- talenro:statement
DROP FUNCTION nodecontrol.v7_require_role(name);
-- talenro:statement
DROP FUNCTION nodecontrol.v7_text_array_is_sorted_unique(text[]);

-- talenro:statement
DROP ROLE nodecontrol_staging_importer;
DROP ROLE nodecontrol_migration_downgrader;
DROP ROLE nodecontrol_upgrade_executor;
