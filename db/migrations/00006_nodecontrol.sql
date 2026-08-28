-- +goose Up
CREATE SCHEMA nodecontrol;

CREATE FUNCTION nodecontrol.text_array_is_sorted_unique(items text[])
RETURNS boolean
LANGUAGE sql
IMMUTABLE
STRICT
AS $$
  SELECT items = ARRAY(
    SELECT DISTINCT item COLLATE "C" AS sorted_item
    FROM unnest(items) AS item
    ORDER BY sorted_item
  )
$$;

CREATE FUNCTION nodecontrol.bytea_array_is_sorted_unique_32(items bytea[])
RETURNS boolean
LANGUAGE sql
IMMUTABLE
STRICT
AS $$
  SELECT cardinality(items) BETWEEN 1 AND 5
     AND array_position(items, NULL) IS NULL
     AND NOT EXISTS (
       SELECT 1 FROM unnest(items) AS item WHERE octet_length(item) <> 32
     )
     AND items = ARRAY(
       SELECT DISTINCT item AS sorted_item
       FROM unnest(items) AS item
       ORDER BY sorted_item
     )
$$;

CREATE TABLE nodecontrol.control_plane_authority_fences (
  operation_id uuid CONSTRAINT control_plane_authority_fences_operation_id_not_null NOT NULL,
  effect_kind text COLLATE "C" CONSTRAINT control_plane_authority_fences_effect_kind_not_null NOT NULL,
  scope_kind text COLLATE "C" CONSTRAINT control_plane_authority_fences_scope_kind_not_null NOT NULL,
  authority_epoch bigint CONSTRAINT control_plane_authority_fences_authority_epoch_not_null NOT NULL,
  authority_sequence bigint CONSTRAINT control_plane_authority_fences_authority_sequence_not_null NOT NULL,
  scope_digest bytea CONSTRAINT control_plane_authority_fences_scope_digest_not_null NOT NULL,
  provider_reservation_digest bytea CONSTRAINT control_plane_authority_fences_provider_reservation_di_a9960db8 NOT NULL,
  effect_digest bytea,
  provider_status text COLLATE "C" CONSTRAINT control_plane_authority_fences_provider_status_not_null NOT NULL,
  provider_receipt_digest bytea,
  db_system_id numeric(20,0),
  db_timeline bigint,
  required_lsn pg_lsn,
  abort_reason text COLLATE "C",
  visibility_state text COLLATE "C" CONSTRAINT control_plane_authority_fences_visibility_state_not_null NOT NULL,
  reserved_at timestamp with time zone CONSTRAINT control_plane_authority_fences_reserved_at_not_null NOT NULL,
  effect_bound_at timestamp with time zone,
  terminal_at timestamp with time zone,
  CONSTRAINT control_plane_authority_fences_pkey PRIMARY KEY (operation_id),
  CONSTRAINT control_plane_authority_fences_effect_kind_enum CHECK (effect_kind IN ('trust_bundle_publish','root_publish','metadata_publish','grant_create','grant_claim','certificate_activate','certificate_revoke','identity_epoch_advance','security_incident_open','security_incident_resolve','resource_envelope_activate','desired_activate','recovery_activate','operator_transition','operator_authorizer_change')),
  CONSTRAINT control_plane_authority_fences_scope_kind_enum CHECK (scope_kind IN ('node','global_node_trust','global_operator_trust')),
  CONSTRAINT control_plane_authority_fences_authority_epoch_positive CHECK (authority_epoch > 0),
  CONSTRAINT control_plane_authority_fences_authority_sequence_positive CHECK (authority_sequence > 0),
  CONSTRAINT control_plane_authority_fences_scope_digest_length CHECK (octet_length(scope_digest) = 32),
  CONSTRAINT control_plane_authority_fences_provider_reservation_di_1c575b33 CHECK (octet_length(provider_reservation_digest) = 32),
  CONSTRAINT control_plane_authority_fences_effect_digest_length CHECK (effect_digest IS NULL OR octet_length(effect_digest) = 32),
  CONSTRAINT control_plane_authority_fences_provider_status_enum CHECK (provider_status IN ('reserved','committed','aborted')),
  CONSTRAINT control_plane_authority_fences_provider_receipt_digest_length CHECK (provider_receipt_digest IS NULL OR octet_length(provider_receipt_digest) = 32),
  CONSTRAINT control_plane_authority_fences_db_system_id_range CHECK (db_system_id IS NULL OR db_system_id BETWEEN 1 AND 18446744073709551615),
  CONSTRAINT control_plane_authority_fences_db_timeline_range CHECK (db_timeline IS NULL OR db_timeline BETWEEN 1 AND 4294967295),
  CONSTRAINT control_plane_authority_fences_abort_reason_enum CHECK (abort_reason IS NULL OR abort_reason IN ('validation_failed','superseded','provider_dependency_failed','activation_deadline_expired')),
  CONSTRAINT control_plane_authority_fences_visibility_state_enum CHECK (visibility_state IN ('fence_pending','active','aborted')),
  CONSTRAINT control_plane_authority_fences_operation_epoch_sequence_key UNIQUE (operation_id, authority_epoch, authority_sequence),
  CONSTRAINT control_plane_authority_fences_epoch_sequence_key UNIQUE (authority_epoch, authority_sequence),
  CONSTRAINT control_plane_authority_fences_database_binding_all_or_none CHECK ((effect_digest IS NULL AND db_system_id IS NULL AND db_timeline IS NULL AND required_lsn IS NULL AND effect_bound_at IS NULL) OR (effect_digest IS NOT NULL AND db_system_id IS NOT NULL AND db_timeline IS NOT NULL AND required_lsn IS NOT NULL AND effect_bound_at IS NOT NULL)),
  CONSTRAINT control_plane_authority_fences_committed_effect_bound CHECK (provider_status <> 'committed' OR effect_digest IS NOT NULL),
  CONSTRAINT control_plane_authority_fences_receipt_status_pair CHECK ((provider_status = 'reserved') = (provider_receipt_digest IS NULL)),
  CONSTRAINT control_plane_authority_fences_terminal_status_pair CHECK ((provider_status = 'reserved') = (terminal_at IS NULL)),
  CONSTRAINT control_plane_authority_fences_active_status_pair CHECK ((visibility_state = 'active') = (provider_status = 'committed')),
  CONSTRAINT control_plane_authority_fences_aborted_visibility_pair CHECK ((visibility_state = 'aborted') = (provider_status = 'aborted')),
  CONSTRAINT control_plane_authority_fences_abort_reason_pair CHECK ((provider_status = 'aborted') = (abort_reason IS NOT NULL)),
  CONSTRAINT control_plane_authority_fences_terminal_order CHECK (terminal_at IS NULL OR terminal_at >= reserved_at),
  CONSTRAINT control_plane_authority_fences_effect_bound_order CHECK (effect_bound_at IS NULL OR effect_bound_at >= reserved_at),
  CONSTRAINT control_plane_authority_fences_effect_scope_matrix CHECK ((scope_kind = 'global_node_trust' AND effect_kind IN ('trust_bundle_publish','root_publish','metadata_publish')) OR (scope_kind = 'global_operator_trust' AND effect_kind IN ('trust_bundle_publish','operator_authorizer_change')) OR (scope_kind = 'node' AND effect_kind IN ('grant_create','grant_claim','certificate_activate','certificate_revoke','identity_epoch_advance','security_incident_open','security_incident_resolve','resource_envelope_activate','desired_activate','recovery_activate','operator_transition')))
);

CREATE TABLE nodecontrol.node_pops (
  pop_code text COLLATE "C" CONSTRAINT node_pops_pop_code_not_null NOT NULL,
  iso_country character(2) COLLATE "C" CONSTRAINT node_pops_iso_country_not_null NOT NULL,
  region text COLLATE "C" CONSTRAINT node_pops_region_not_null NOT NULL,
  operator_state text COLLATE "C" CONSTRAINT node_pops_operator_state_not_null NOT NULL,
  version bigint CONSTRAINT node_pops_version_not_null NOT NULL DEFAULT 1,
  created_at timestamp with time zone CONSTRAINT node_pops_created_at_not_null NOT NULL,
  updated_at timestamp with time zone CONSTRAINT node_pops_updated_at_not_null NOT NULL,
  CONSTRAINT node_pops_pkey PRIMARY KEY (pop_code),
  CONSTRAINT node_pops_pop_code_format CHECK (octet_length(pop_code) BETWEEN 1 AND 32 AND pop_code ~ '^[a-z0-9](?:[a-z0-9-]{0,30}[a-z0-9])?$'),
  CONSTRAINT node_pops_iso_country_format CHECK (octet_length(iso_country) = 2 AND iso_country ~ '^[A-Z]{2}$'),
  CONSTRAINT node_pops_region_format CHECK (octet_length(region) BETWEEN 1 AND 64 AND region ~ '^[a-z0-9](?:[a-z0-9-]{0,62}[a-z0-9])?$'),
  CONSTRAINT node_pops_operator_state_enum CHECK (operator_state IN ('provisioning','enabled','draining','disabled')),
  CONSTRAINT node_pops_version_range CHECK (version BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_pops_timestamp_order CHECK (updated_at >= created_at),
  CONSTRAINT node_pops_country_region_code_key UNIQUE (iso_country, region, pop_code)
);

CREATE TABLE nodecontrol.node_failure_domains (
  failure_domain_id uuid CONSTRAINT node_failure_domains_failure_domain_id_not_null NOT NULL,
  domain_type text COLLATE "C" CONSTRAINT node_failure_domains_domain_type_not_null NOT NULL,
  stable_id text COLLATE "C" CONSTRAINT node_failure_domains_stable_id_not_null NOT NULL,
  version bigint CONSTRAINT node_failure_domains_version_not_null NOT NULL DEFAULT 1,
  created_at timestamp with time zone CONSTRAINT node_failure_domains_created_at_not_null NOT NULL,
  updated_at timestamp with time zone CONSTRAINT node_failure_domains_updated_at_not_null NOT NULL,
  CONSTRAINT node_failure_domains_pkey PRIMARY KEY (failure_domain_id),
  CONSTRAINT node_failure_domains_domain_type_enum CHECK (domain_type IN ('facility','compute','upstream')),
  CONSTRAINT node_failure_domains_stable_id_format CHECK (octet_length(stable_id) BETWEEN 1 AND 128 AND stable_id ~ '^[!-~]{1,128}$'),
  CONSTRAINT node_failure_domains_version_range CHECK (version BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_failure_domains_timestamp_order CHECK (updated_at >= created_at),
  CONSTRAINT node_failure_domains_type_stable_id_key UNIQUE (domain_type, stable_id),
  CONSTRAINT node_failure_domains_id_type_key UNIQUE (failure_domain_id, domain_type)
);

CREATE TABLE nodecontrol.node_capacity_profiles (
  profile_id text COLLATE "C" CONSTRAINT node_capacity_profiles_profile_id_not_null NOT NULL,
  version bigint CONSTRAINT node_capacity_profiles_version_not_null NOT NULL,
  adapter text COLLATE "C" CONSTRAINT node_capacity_profiles_adapter_not_null NOT NULL,
  egress_limit_bps bigint CONSTRAINT node_capacity_profiles_egress_limit_bps_not_null NOT NULL,
  connection_limit integer CONSTRAINT node_capacity_profiles_connection_limit_not_null NOT NULL,
  handshake_limit_per_second integer CONSTRAINT node_capacity_profiles_handshake_limit_per_second_not_null NOT NULL,
  cpu_quota_millicores integer CONSTRAINT node_capacity_profiles_cpu_quota_millicores_not_null NOT NULL,
  cpu_limit_basis_points integer CONSTRAINT node_capacity_profiles_cpu_limit_basis_points_not_null NOT NULL,
  memory_limit_bytes bigint CONSTRAINT node_capacity_profiles_memory_limit_bytes_not_null NOT NULL,
  task_limit integer CONSTRAINT node_capacity_profiles_task_limit_not_null NOT NULL,
  file_descriptor_limit integer CONSTRAINT node_capacity_profiles_file_descriptor_limit_not_null NOT NULL,
  queue_limit integer CONSTRAINT node_capacity_profiles_queue_limit_not_null NOT NULL,
  packet_loss_limit_basis_points integer CONSTRAINT node_capacity_profiles_packet_loss_limit_basis_points_not_null NOT NULL,
  required_metrics text[] COLLATE "C" CONSTRAINT node_capacity_profiles_required_metrics_not_null NOT NULL,
  created_at timestamp with time zone CONSTRAINT node_capacity_profiles_created_at_not_null NOT NULL,
  CONSTRAINT node_capacity_profiles_pkey PRIMARY KEY (profile_id, version),
  CONSTRAINT node_capacity_profiles_profile_id_format CHECK (octet_length(profile_id) BETWEEN 1 AND 128 AND profile_id ~ '^[A-Za-z0-9][A-Za-z0-9._-]*$'),
  CONSTRAINT node_capacity_profiles_version_range CHECK (version BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_capacity_profiles_adapter_enum CHECK (adapter IN ('fixture','xray','sing_box')),
  CONSTRAINT node_capacity_profiles_egress_limit_range CHECK (egress_limit_bps BETWEEN 1000000 AND 1000000000000),
  CONSTRAINT node_capacity_profiles_connection_limit_range CHECK (connection_limit BETWEEN 1 AND 10000000),
  CONSTRAINT node_capacity_profiles_handshake_limit_range CHECK (handshake_limit_per_second BETWEEN 1 AND 1000000),
  CONSTRAINT node_capacity_profiles_cpu_quota_range CHECK (cpu_quota_millicores BETWEEN 100 AND 64000),
  CONSTRAINT node_capacity_profiles_cpu_basis_points_range CHECK (cpu_limit_basis_points BETWEEN 1 AND 10000),
  CONSTRAINT node_capacity_profiles_memory_limit_range CHECK (memory_limit_bytes BETWEEN 67108864 AND 1099511627776),
  CONSTRAINT node_capacity_profiles_task_limit_range CHECK (task_limit BETWEEN 32 AND 4096),
  CONSTRAINT node_capacity_profiles_file_descriptor_limit_range CHECK (file_descriptor_limit BETWEEN 64 AND 1000000),
  CONSTRAINT node_capacity_profiles_queue_limit_range CHECK (queue_limit BETWEEN 1 AND 1000000),
  CONSTRAINT node_capacity_profiles_packet_loss_range CHECK (packet_loss_limit_basis_points BETWEEN 1 AND 10000),
  CONSTRAINT node_capacity_profiles_required_metrics_exact CHECK (cardinality(required_metrics) BETWEEN 1 AND 5 AND array_position(required_metrics,NULL) IS NULL AND required_metrics <@ ARRAY['cpu_basis_points','egress_bps','memory_bytes','open_file_descriptors','task_count']::text[] AND nodecontrol.text_array_is_sorted_unique(required_metrics))
);

CREATE TABLE nodecontrol.node_inventory (
  node_id uuid CONSTRAINT node_inventory_node_id_not_null NOT NULL,
  pop_code text COLLATE "C" CONSTRAINT node_inventory_pop_code_not_null NOT NULL,
  operator_state text COLLATE "C" CONSTRAINT node_inventory_operator_state_not_null NOT NULL,
  security_state text COLLATE "C" CONSTRAINT node_inventory_security_state_not_null NOT NULL,
  identity_state text COLLATE "C" CONSTRAINT node_inventory_identity_state_not_null NOT NULL,
  resume_operator_state text COLLATE "C",
  pending_operator_transition text COLLATE "C",
  pending_transition_signing_id uuid,
  identity_epoch bigint CONSTRAINT node_inventory_identity_epoch_not_null NOT NULL DEFAULT 0,
  lineage_id uuid,
  inventory_version bigint CONSTRAINT node_inventory_inventory_version_not_null NOT NULL DEFAULT 1,
  security_version bigint CONSTRAINT node_inventory_security_version_not_null NOT NULL DEFAULT 1,
  resource_envelope_version bigint,
  resource_envelope_digest bytea,
  active_desired_generation bigint,
  next_desired_generation bigint CONSTRAINT node_inventory_next_desired_generation_not_null NOT NULL DEFAULT 1,
  active_recovery_generation bigint,
  next_recovery_generation bigint CONSTRAINT node_inventory_next_recovery_generation_not_null NOT NULL DEFAULT 1,
  active_root_publish_id uuid,
  active_root_version bigint,
  active_metadata_publish_id uuid,
  active_metadata_version bigint,
  last_authority_operation_id uuid,
  last_authority_epoch bigint,
  last_authority_sequence bigint,
  created_at timestamp with time zone CONSTRAINT node_inventory_created_at_not_null NOT NULL,
  updated_at timestamp with time zone CONSTRAINT node_inventory_updated_at_not_null NOT NULL,
  CONSTRAINT node_inventory_pkey PRIMARY KEY (node_id),
  CONSTRAINT node_inventory_pop_fk FOREIGN KEY (pop_code) REFERENCES nodecontrol.node_pops(pop_code) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_inventory_pop_code_format CHECK (octet_length(pop_code) BETWEEN 1 AND 32 AND pop_code ~ '^[a-z0-9](?:[a-z0-9-]{0,30}[a-z0-9])?$'),
  CONSTRAINT node_inventory_operator_state_enum CHECK (operator_state IN ('provisioning','enabled','draining','disabled')),
  CONSTRAINT node_inventory_security_state_enum CHECK (security_state IN ('normal','quarantined')),
  CONSTRAINT node_inventory_identity_state_enum CHECK (identity_state IN ('never_enrolled','active','recovery_pending','recovery_limited','revoked')),
  CONSTRAINT node_inventory_resume_operator_state_enum CHECK (resume_operator_state IS NULL OR resume_operator_state IN ('enabled','draining','disabled')),
  CONSTRAINT node_inventory_pending_operator_transition_enum CHECK (pending_operator_transition IS NULL OR pending_operator_transition IN ('draining','resume','restore_reauthorize')),
  CONSTRAINT node_inventory_pending_transition_pair CHECK ((pending_operator_transition IS NULL) = (pending_transition_signing_id IS NULL)),
  CONSTRAINT node_inventory_identity_epoch_range CHECK (identity_epoch BETWEEN 0 AND 9223372036854775807),
  CONSTRAINT node_inventory_identity_lineage_state CHECK ((identity_state = 'never_enrolled' AND identity_epoch = 0 AND lineage_id IS NULL) OR (identity_state <> 'never_enrolled' AND identity_epoch > 0 AND lineage_id IS NOT NULL)),
  CONSTRAINT node_inventory_inventory_version_range CHECK (inventory_version BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_inventory_security_version_range CHECK (security_version BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_inventory_next_desired_generation_range CHECK (next_desired_generation BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_inventory_next_recovery_generation_range CHECK (next_recovery_generation BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_inventory_resource_envelope_pair CHECK ((resource_envelope_version IS NULL) = (resource_envelope_digest IS NULL)),
  CONSTRAINT node_inventory_resource_envelope_version_range CHECK (resource_envelope_version IS NULL OR resource_envelope_version BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_inventory_resource_envelope_digest_length CHECK (resource_envelope_digest IS NULL OR octet_length(resource_envelope_digest) = 32),
  CONSTRAINT node_inventory_desired_generation_order CHECK (active_desired_generation IS NULL OR (active_desired_generation BETWEEN 1 AND 9223372036854775807 AND active_desired_generation < next_desired_generation)),
  CONSTRAINT node_inventory_recovery_generation_order CHECK (active_recovery_generation IS NULL OR (active_recovery_generation BETWEEN 1 AND 9223372036854775807 AND active_recovery_generation < next_recovery_generation)),
  CONSTRAINT node_inventory_root_pointer_pair CHECK ((active_root_publish_id IS NULL) = (active_root_version IS NULL)),
  CONSTRAINT node_inventory_metadata_pointer_pair CHECK ((active_metadata_publish_id IS NULL) = (active_metadata_version IS NULL)),
  CONSTRAINT node_inventory_root_version_range CHECK (active_root_version IS NULL OR active_root_version BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_inventory_metadata_version_range CHECK (active_metadata_version IS NULL OR active_metadata_version BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_inventory_last_authority_all_or_none CHECK (num_nonnulls(last_authority_operation_id,last_authority_epoch,last_authority_sequence) IN (0,3)),
  CONSTRAINT node_inventory_last_authority_epoch_range CHECK (last_authority_epoch IS NULL OR last_authority_epoch BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_inventory_last_authority_sequence_range CHECK (last_authority_sequence IS NULL OR last_authority_sequence BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_inventory_last_authority_fence_fk FOREIGN KEY (last_authority_operation_id, last_authority_epoch, last_authority_sequence) REFERENCES nodecontrol.control_plane_authority_fences(operation_id, authority_epoch, authority_sequence) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_inventory_quarantine_state CHECK ((security_state = 'normal' AND resume_operator_state IS NULL) OR (security_state = 'quarantined' AND operator_state = 'disabled' AND resume_operator_state IS NOT NULL)),
  CONSTRAINT node_inventory_timestamp_order CHECK (updated_at >= created_at)
);

CREATE TABLE nodecontrol.node_failure_domain_membership (
  node_id uuid CONSTRAINT node_failure_domain_membership_node_id_not_null NOT NULL,
  failure_domain_id uuid CONSTRAINT node_failure_domain_membership_failure_domain_id_not_null NOT NULL,
  domain_type text COLLATE "C" CONSTRAINT node_failure_domain_membership_domain_type_not_null NOT NULL,
  inventory_version bigint CONSTRAINT node_failure_domain_membership_inventory_version_not_null NOT NULL,
  created_at timestamp with time zone CONSTRAINT node_failure_domain_membership_created_at_not_null NOT NULL,
  CONSTRAINT node_failure_domain_membership_pkey PRIMARY KEY (node_id, failure_domain_id),
  CONSTRAINT node_failure_domain_membership_node_fk FOREIGN KEY (node_id) REFERENCES nodecontrol.node_inventory(node_id) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_failure_domain_membership_domain_fk FOREIGN KEY (failure_domain_id, domain_type) REFERENCES nodecontrol.node_failure_domains(failure_domain_id, domain_type) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_failure_domain_membership_domain_type_enum CHECK (domain_type IN ('facility','compute','upstream')),
  CONSTRAINT node_failure_domain_membership_inventory_version_range CHECK (inventory_version BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_failure_domain_membership_node_domain_type_key UNIQUE (node_id, domain_type)
);

CREATE TABLE nodecontrol.node_endpoints (
  endpoint_id uuid CONSTRAINT node_endpoints_endpoint_id_not_null NOT NULL,
  node_id uuid CONSTRAINT node_endpoints_node_id_not_null NOT NULL,
  address text COLLATE "C" CONSTRAINT node_endpoints_address_not_null NOT NULL,
  port integer CONSTRAINT node_endpoints_port_not_null NOT NULL,
  transport text COLLATE "C" CONSTRAINT node_endpoints_transport_not_null NOT NULL,
  protocol_capability text COLLATE "C" CONSTRAINT node_endpoints_protocol_capability_not_null NOT NULL,
  operator_state text COLLATE "C" CONSTRAINT node_endpoints_operator_state_not_null NOT NULL,
  inventory_version bigint CONSTRAINT node_endpoints_inventory_version_not_null NOT NULL,
  created_at timestamp with time zone CONSTRAINT node_endpoints_created_at_not_null NOT NULL,
  updated_at timestamp with time zone CONSTRAINT node_endpoints_updated_at_not_null NOT NULL,
  CONSTRAINT node_endpoints_pkey PRIMARY KEY (endpoint_id),
  CONSTRAINT node_endpoints_node_fk FOREIGN KEY (node_id) REFERENCES nodecontrol.node_inventory(node_id) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_endpoints_address_format CHECK (octet_length(address) BETWEEN 1 AND 253 AND address ~ '^[A-Za-z0-9][A-Za-z0-9.:-]*$'),
  CONSTRAINT node_endpoints_port_range CHECK (port BETWEEN 1 AND 65535),
  CONSTRAINT node_endpoints_transport_enum CHECK (transport IN ('tcp','udp')),
  CONSTRAINT node_endpoints_protocol_capability_enum CHECK (protocol_capability IN ('bootstrap_v1','agent_control_v1','agent_observation_v1')),
  CONSTRAINT node_endpoints_operator_state_enum CHECK (operator_state IN ('provisioning','enabled','draining','disabled')),
  CONSTRAINT node_endpoints_inventory_version_range CHECK (inventory_version BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_endpoints_timestamp_order CHECK (updated_at >= created_at),
  CONSTRAINT node_endpoints_exact_endpoint_key UNIQUE (node_id, address, port, transport, protocol_capability)
);

CREATE TABLE nodecontrol.node_process_slots (
  node_id uuid CONSTRAINT node_process_slots_node_id_not_null NOT NULL,
  slot_id text COLLATE "C" CONSTRAINT node_process_slots_slot_id_not_null NOT NULL,
  adapter text COLLATE "C" CONSTRAINT node_process_slots_adapter_not_null NOT NULL,
  capacity_profile_id text COLLATE "C" CONSTRAINT node_process_slots_capacity_profile_id_not_null NOT NULL,
  capacity_profile_version bigint CONSTRAINT node_process_slots_capacity_profile_version_not_null NOT NULL,
  required boolean CONSTRAINT node_process_slots_required_not_null NOT NULL,
  operator_state text COLLATE "C" CONSTRAINT node_process_slots_operator_state_not_null NOT NULL,
  inventory_version bigint CONSTRAINT node_process_slots_inventory_version_not_null NOT NULL,
  created_at timestamp with time zone CONSTRAINT node_process_slots_created_at_not_null NOT NULL,
  updated_at timestamp with time zone CONSTRAINT node_process_slots_updated_at_not_null NOT NULL,
  CONSTRAINT node_process_slots_pkey PRIMARY KEY (node_id, slot_id),
  CONSTRAINT node_process_slots_node_fk FOREIGN KEY (node_id) REFERENCES nodecontrol.node_inventory(node_id) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_process_slots_capacity_profile_fk FOREIGN KEY (capacity_profile_id, capacity_profile_version) REFERENCES nodecontrol.node_capacity_profiles(profile_id, version) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_process_slots_slot_id_format CHECK (octet_length(slot_id) BETWEEN 1 AND 64 AND slot_id ~ '^[A-Za-z0-9][A-Za-z0-9._-]*$'),
  CONSTRAINT node_process_slots_adapter_enum CHECK (adapter IN ('fixture','xray','sing_box')),
  CONSTRAINT node_process_slots_capacity_profile_id_format CHECK (octet_length(capacity_profile_id) BETWEEN 1 AND 128 AND capacity_profile_id ~ '^[A-Za-z0-9][A-Za-z0-9._-]*$'),
  CONSTRAINT node_process_slots_capacity_profile_version_range CHECK (capacity_profile_version BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_process_slots_operator_state_enum CHECK (operator_state IN ('provisioning','enabled','draining','disabled')),
  CONSTRAINT node_process_slots_inventory_version_range CHECK (inventory_version BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_process_slots_timestamp_order CHECK (updated_at >= created_at)
);

CREATE TABLE nodecontrol.node_resource_envelopes (
  node_id uuid CONSTRAINT node_resource_envelopes_node_id_not_null NOT NULL,
  envelope_version bigint CONSTRAINT node_resource_envelopes_envelope_version_not_null NOT NULL,
  authority_operation_id uuid CONSTRAINT node_resource_envelopes_authority_operation_id_not_null NOT NULL,
  authority_epoch bigint CONSTRAINT node_resource_envelopes_authority_epoch_not_null NOT NULL,
  authority_sequence bigint CONSTRAINT node_resource_envelopes_authority_sequence_not_null NOT NULL,
  envelope_digest bytea CONSTRAINT node_resource_envelopes_envelope_digest_not_null NOT NULL,
  canonical_package bytea CONSTRAINT node_resource_envelopes_canonical_package_not_null NOT NULL,
  deployment_key_id bytea CONSTRAINT node_resource_envelopes_deployment_key_id_not_null NOT NULL,
  signature bytea CONSTRAINT node_resource_envelopes_signature_not_null NOT NULL,
  max_slots integer CONSTRAINT node_resource_envelopes_max_slots_not_null NOT NULL DEFAULT 8,
  agent_cpu_millicores integer CONSTRAINT node_resource_envelopes_agent_cpu_millicores_not_null NOT NULL,
  agent_memory_bytes bigint CONSTRAINT node_resource_envelopes_agent_memory_bytes_not_null NOT NULL,
  agent_task_limit integer CONSTRAINT node_resource_envelopes_agent_task_limit_not_null NOT NULL,
  agent_file_descriptor_limit integer CONSTRAINT node_resource_envelopes_agent_file_descriptor_limit_not_null NOT NULL,
  supervisor_cpu_millicores integer CONSTRAINT node_resource_envelopes_supervisor_cpu_millicores_not_null NOT NULL,
  supervisor_memory_bytes bigint CONSTRAINT node_resource_envelopes_supervisor_memory_bytes_not_null NOT NULL,
  supervisor_task_limit integer CONSTRAINT node_resource_envelopes_supervisor_task_limit_not_null NOT NULL,
  supervisor_file_descriptor_limit integer CONSTRAINT node_resource_envelopes_supervisor_file_descriptor_lim_2b59cce8 NOT NULL,
  core_parent_cpu_millicores integer CONSTRAINT node_resource_envelopes_core_parent_cpu_millicores_not_null NOT NULL,
  core_parent_memory_bytes bigint CONSTRAINT node_resource_envelopes_core_parent_memory_bytes_not_null NOT NULL,
  core_parent_task_limit integer CONSTRAINT node_resource_envelopes_core_parent_task_limit_not_null NOT NULL,
  core_parent_file_descriptor_limit integer CONSTRAINT node_resource_envelopes_core_parent_file_descriptor_li_c963936c NOT NULL,
  aggregate_slot_file_descriptor_limit integer CONSTRAINT node_resource_envelopes_aggregate_slot_file_descriptor_e490d17e NOT NULL,
  aggregate_slot_tmpfs_bytes bigint CONSTRAINT node_resource_envelopes_aggregate_slot_tmpfs_bytes_not_null NOT NULL,
  aggregate_slot_tmpfs_inodes bigint CONSTRAINT node_resource_envelopes_aggregate_slot_tmpfs_inodes_not_null NOT NULL,
  detected_host_capacity_digest bytea CONSTRAINT node_resource_envelopes_detected_host_capacity_digest_not_null NOT NULL,
  issued_at timestamp with time zone CONSTRAINT node_resource_envelopes_issued_at_not_null NOT NULL,
  created_at timestamp with time zone CONSTRAINT node_resource_envelopes_created_at_not_null NOT NULL,
  CONSTRAINT node_resource_envelopes_pkey PRIMARY KEY (node_id, envelope_version),
  CONSTRAINT node_resource_envelopes_authority_epoch_range CHECK (authority_epoch BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_resource_envelopes_authority_sequence_range CHECK (authority_sequence BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_resource_envelopes_authority_fence_fk FOREIGN KEY (authority_operation_id, authority_epoch, authority_sequence) REFERENCES nodecontrol.control_plane_authority_fences(operation_id, authority_epoch, authority_sequence) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_resource_envelopes_node_fk FOREIGN KEY (node_id) REFERENCES nodecontrol.node_inventory(node_id) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_resource_envelopes_envelope_version_range CHECK (envelope_version BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_resource_envelopes_envelope_digest_length CHECK (octet_length(envelope_digest) = 32),
  CONSTRAINT node_resource_envelopes_canonical_package_length CHECK (octet_length(canonical_package) BETWEEN 1 AND 65536),
  CONSTRAINT node_resource_envelopes_deployment_key_id_length CHECK (octet_length(deployment_key_id) = 32),
  CONSTRAINT node_resource_envelopes_signature_length CHECK (octet_length(signature) = 64),
  CONSTRAINT node_resource_envelopes_max_slots_exact CHECK (max_slots = 8),
  CONSTRAINT node_resource_envelopes_agent_cpu_millicores_range CHECK (agent_cpu_millicores BETWEEN 100 AND 64000),
  CONSTRAINT node_resource_envelopes_supervisor_cpu_millicores_range CHECK (supervisor_cpu_millicores BETWEEN 100 AND 64000),
  CONSTRAINT node_resource_envelopes_core_parent_cpu_millicores_range CHECK (core_parent_cpu_millicores BETWEEN 100 AND 64000),
  CONSTRAINT node_resource_envelopes_agent_memory_bytes_range CHECK (agent_memory_bytes BETWEEN 67108864 AND 1099511627776),
  CONSTRAINT node_resource_envelopes_supervisor_memory_bytes_range CHECK (supervisor_memory_bytes BETWEEN 67108864 AND 1099511627776),
  CONSTRAINT node_resource_envelopes_core_parent_memory_bytes_range CHECK (core_parent_memory_bytes BETWEEN 67108864 AND 1099511627776),
  CONSTRAINT node_resource_envelopes_agent_task_limit_range CHECK (agent_task_limit BETWEEN 32 AND 4096),
  CONSTRAINT node_resource_envelopes_supervisor_task_limit_range CHECK (supervisor_task_limit BETWEEN 32 AND 4096),
  CONSTRAINT node_resource_envelopes_core_parent_task_limit_range CHECK (core_parent_task_limit BETWEEN 32 AND 4096),
  CONSTRAINT node_resource_envelopes_agent_file_descriptor_limit_range CHECK (agent_file_descriptor_limit BETWEEN 64 AND 1000000),
  CONSTRAINT node_resource_envelopes_supervisor_file_descriptor_limit_range CHECK (supervisor_file_descriptor_limit BETWEEN 64 AND 1000000),
  CONSTRAINT node_resource_envelopes_core_parent_file_descriptor_limit_range CHECK (core_parent_file_descriptor_limit BETWEEN 64 AND 1000000),
  CONSTRAINT node_resource_envelopes_aggregate_fd_range CHECK (aggregate_slot_file_descriptor_limit BETWEEN 64 AND 8000000),
  CONSTRAINT node_resource_envelopes_aggregate_tmpfs_bytes_range CHECK (aggregate_slot_tmpfs_bytes BETWEEN 1048576 AND 1099511627776),
  CONSTRAINT node_resource_envelopes_aggregate_tmpfs_inodes_range CHECK (aggregate_slot_tmpfs_inodes BETWEEN 1 AND 10000000),
  CONSTRAINT node_resource_envelopes_detected_host_capacity_digest_length CHECK (octet_length(detected_host_capacity_digest) = 32),
  CONSTRAINT node_resource_envelopes_issued_created_order CHECK (created_at >= issued_at),
  CONSTRAINT node_resource_envelopes_node_digest_key UNIQUE (node_id, envelope_digest)
);

CREATE TABLE nodecontrol.node_certificate_issuances (
  issuance_id uuid CONSTRAINT node_certificate_issuances_issuance_id_not_null NOT NULL,
  authority_operation_id uuid CONSTRAINT node_certificate_issuances_authority_operation_id_not_null NOT NULL,
  authority_epoch bigint CONSTRAINT node_certificate_issuances_authority_epoch_not_null NOT NULL,
  authority_sequence bigint CONSTRAINT node_certificate_issuances_authority_sequence_not_null NOT NULL,
  node_id uuid CONSTRAINT node_certificate_issuances_node_id_not_null NOT NULL,
  attempt_id uuid CONSTRAINT node_certificate_issuances_attempt_id_not_null NOT NULL,
  issuance_kind text COLLATE "C" CONSTRAINT node_certificate_issuances_issuance_kind_not_null NOT NULL,
  identity_epoch bigint CONSTRAINT node_certificate_issuances_identity_epoch_not_null NOT NULL,
  lineage_id uuid CONSTRAINT node_certificate_issuances_lineage_id_not_null NOT NULL,
  issuer_id text COLLATE "C" CONSTRAINT node_certificate_issuances_issuer_id_not_null NOT NULL,
  csr_sha256 bytea CONSTRAINT node_certificate_issuances_csr_sha256_not_null NOT NULL,
  public_key_sha256 bytea CONSTRAINT node_certificate_issuances_public_key_sha256_not_null NOT NULL,
  template_sha256 bytea CONSTRAINT node_certificate_issuances_template_sha256_not_null NOT NULL,
  request_digest bytea CONSTRAINT node_certificate_issuances_request_digest_not_null NOT NULL,
  status text COLLATE "C" CONSTRAINT node_certificate_issuances_status_not_null NOT NULL,
  serial_bytes bytea,
  leaf_der bytea,
  leaf_der_sha256 bytea,
  chain_der bytea,
  chain_der_sha256 bytea,
  not_before timestamp with time zone,
  not_after timestamp with time zone,
  failure_reason text COLLATE "C",
  created_at timestamp with time zone CONSTRAINT node_certificate_issuances_created_at_not_null NOT NULL,
  updated_at timestamp with time zone CONSTRAINT node_certificate_issuances_updated_at_not_null NOT NULL,
  terminal_at timestamp with time zone,
  retention_until timestamp with time zone,
  CONSTRAINT node_certificate_issuances_pkey PRIMARY KEY (issuance_id),
  CONSTRAINT node_certificate_issuances_authority_epoch_range CHECK (authority_epoch BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_certificate_issuances_authority_sequence_range CHECK (authority_sequence BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_certificate_issuances_authority_fence_fk FOREIGN KEY (authority_operation_id, authority_epoch, authority_sequence) REFERENCES nodecontrol.control_plane_authority_fences(operation_id, authority_epoch, authority_sequence) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_certificate_issuances_node_fk FOREIGN KEY (node_id) REFERENCES nodecontrol.node_inventory(node_id) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_certificate_issuances_node_attempt_key UNIQUE (node_id, attempt_id),
  CONSTRAINT node_certificate_issuances_authority_operation_key UNIQUE (authority_operation_id),
  CONSTRAINT node_certificate_issuances_issuance_kind_enum CHECK (issuance_kind IN ('initial','rotation','recovery')),
  CONSTRAINT node_certificate_issuances_identity_epoch_range CHECK (identity_epoch BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_certificate_issuances_issuer_id_format CHECK (octet_length(issuer_id) BETWEEN 1 AND 128 AND issuer_id ~ '^[A-Za-z0-9:_-]{1,128}$'),
  CONSTRAINT node_certificate_issuances_csr_sha256_length CHECK (octet_length(csr_sha256) = 32),
  CONSTRAINT node_certificate_issuances_public_key_sha256_length CHECK (octet_length(public_key_sha256) = 32),
  CONSTRAINT node_certificate_issuances_template_sha256_length CHECK (octet_length(template_sha256) = 32),
  CONSTRAINT node_certificate_issuances_request_digest_length CHECK (octet_length(request_digest) = 32),
  CONSTRAINT node_certificate_issuances_status_enum CHECK (status IN ('pending','active','rejected','superseded','failed')),
  CONSTRAINT node_certificate_issuances_serial_length CHECK (serial_bytes IS NULL OR octet_length(serial_bytes) BETWEEN 1 AND 20),
  CONSTRAINT node_certificate_issuances_leaf_der_length CHECK (leaf_der IS NULL OR octet_length(leaf_der) BETWEEN 1 AND 65536),
  CONSTRAINT node_certificate_issuances_leaf_der_sha256_length CHECK (leaf_der_sha256 IS NULL OR octet_length(leaf_der_sha256) = 32),
  CONSTRAINT node_certificate_issuances_chain_der_length CHECK (chain_der IS NULL OR octet_length(chain_der) BETWEEN 1 AND 262144),
  CONSTRAINT node_certificate_issuances_chain_der_sha256_length CHECK (chain_der_sha256 IS NULL OR octet_length(chain_der_sha256) = 32),
  CONSTRAINT node_certificate_issuances_result_all_or_none CHECK (num_nonnulls(serial_bytes,leaf_der,leaf_der_sha256,chain_der,chain_der_sha256,not_before,not_after) IN (0,7)),
  CONSTRAINT node_certificate_issuances_result_validity CHECK (not_before IS NULL OR not_after > not_before),
  CONSTRAINT node_certificate_issuances_failure_reason_enum CHECK (failure_reason IS NULL OR failure_reason IN ('ca_rejected','csr_invalid','template_invalid','provider_failed','superseded')),
  CONSTRAINT node_certificate_issuances_status_result CHECK ((status = 'pending' AND terminal_at IS NULL AND retention_until IS NULL) OR (status = 'active' AND serial_bytes IS NOT NULL AND terminal_at IS NOT NULL AND failure_reason IS NULL AND retention_until IS NOT NULL) OR (status IN ('rejected','superseded','failed') AND terminal_at IS NOT NULL AND failure_reason IS NOT NULL AND retention_until IS NOT NULL)),
  CONSTRAINT node_certificate_issuances_retention_window CHECK (retention_until IS NULL OR (terminal_at IS NOT NULL AND retention_until >= terminal_at + interval '30 days')),
  CONSTRAINT node_certificate_issuances_timestamp_order CHECK (updated_at >= created_at AND (terminal_at IS NULL OR terminal_at >= created_at))
);

CREATE TABLE nodecontrol.node_enrollment_grants (
  grant_id uuid CONSTRAINT node_enrollment_grants_grant_id_not_null NOT NULL,
  authority_operation_id uuid CONSTRAINT node_enrollment_grants_authority_operation_id_not_null NOT NULL,
  authority_epoch bigint CONSTRAINT node_enrollment_grants_authority_epoch_not_null NOT NULL,
  authority_sequence bigint CONSTRAINT node_enrollment_grants_authority_sequence_not_null NOT NULL,
  node_id uuid CONSTRAINT node_enrollment_grants_node_id_not_null NOT NULL,
  identity_epoch bigint CONSTRAINT node_enrollment_grants_identity_epoch_not_null NOT NULL,
  token_digest bytea CONSTRAINT node_enrollment_grants_token_digest_not_null NOT NULL,
  csr_digest bytea CONSTRAINT node_enrollment_grants_csr_digest_not_null NOT NULL,
  idempotency_digest bytea CONSTRAINT node_enrollment_grants_idempotency_digest_not_null NOT NULL,
  created_at timestamp with time zone CONSTRAINT node_enrollment_grants_created_at_not_null NOT NULL,
  expires_at timestamp with time zone CONSTRAINT node_enrollment_grants_expires_at_not_null NOT NULL,
  consumed_at timestamp with time zone,
  consumption_attempt_id uuid,
  consumption_request_digest bytea,
  claim_authority_operation_id uuid,
  claim_authority_epoch bigint,
  claim_authority_sequence bigint,
  result_issuance_id uuid,
  expired_at timestamp with time zone,
  invalidated_at timestamp with time zone,
  terminal_reason text COLLATE "C",
  terminal_at timestamp with time zone,
  retention_until timestamp with time zone,
  CONSTRAINT node_enrollment_grants_pkey PRIMARY KEY (grant_id),
  CONSTRAINT node_enrollment_grants_authority_epoch_range CHECK (authority_epoch BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_enrollment_grants_authority_sequence_range CHECK (authority_sequence BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_enrollment_grants_authority_fence_fk FOREIGN KEY (authority_operation_id, authority_epoch, authority_sequence) REFERENCES nodecontrol.control_plane_authority_fences(operation_id, authority_epoch, authority_sequence) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_enrollment_grants_node_fk FOREIGN KEY (node_id) REFERENCES nodecontrol.node_inventory(node_id) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_enrollment_grants_result_issuance_fk FOREIGN KEY (result_issuance_id) REFERENCES nodecontrol.node_certificate_issuances(issuance_id) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_enrollment_grants_claim_authority_fence_fk FOREIGN KEY (claim_authority_operation_id, claim_authority_epoch, claim_authority_sequence) REFERENCES nodecontrol.control_plane_authority_fences(operation_id, authority_epoch, authority_sequence) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_enrollment_grants_identity_epoch_range CHECK (identity_epoch BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_enrollment_grants_token_digest_length CHECK (octet_length(token_digest) = 32),
  CONSTRAINT node_enrollment_grants_csr_digest_length CHECK (octet_length(csr_digest) = 32),
  CONSTRAINT node_enrollment_grants_idempotency_digest_length CHECK (octet_length(idempotency_digest) = 32),
  CONSTRAINT node_enrollment_grants_consumption_request_digest_length CHECK (consumption_request_digest IS NULL OR octet_length(consumption_request_digest) = 32),
  CONSTRAINT node_enrollment_grants_token_digest_key UNIQUE (token_digest),
  CONSTRAINT node_enrollment_grants_expiry_window CHECK (expires_at > created_at AND expires_at <= created_at + interval '10 minutes'),
  CONSTRAINT node_enrollment_grants_consumption_all_or_none CHECK (num_nonnulls(consumed_at,consumption_attempt_id,consumption_request_digest,claim_authority_operation_id,claim_authority_epoch,claim_authority_sequence,result_issuance_id) IN (0,7)),
  CONSTRAINT node_enrollment_grants_claim_authority_epoch_range CHECK (claim_authority_epoch IS NULL OR claim_authority_epoch BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_enrollment_grants_claim_authority_sequence_range CHECK (claim_authority_sequence IS NULL OR claim_authority_sequence BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_enrollment_grants_exclusive_terminal_kind CHECK (num_nonnulls(consumed_at,expired_at,invalidated_at) <= 1),
  CONSTRAINT node_enrollment_grants_terminal_reason_enum CHECK (terminal_reason IS NULL OR terminal_reason IN ('consumed','expired','identity_epoch_advanced','operator_disabled','security_quarantine','superseded')),
  CONSTRAINT node_enrollment_grants_terminal_group CHECK (
    (consumed_at IS NULL AND expired_at IS NULL AND invalidated_at IS NULL AND terminal_reason IS NULL AND terminal_at IS NULL AND retention_until IS NULL)
    OR (consumed_at IS NOT NULL AND expired_at IS NULL AND invalidated_at IS NULL AND terminal_reason IS NOT NULL AND terminal_reason = 'consumed' AND terminal_at IS NOT NULL AND terminal_at = consumed_at AND retention_until IS NOT NULL)
    OR (consumed_at IS NULL AND expired_at IS NOT NULL AND invalidated_at IS NULL AND terminal_reason IS NOT NULL AND terminal_reason = 'expired' AND terminal_at IS NOT NULL AND terminal_at = expired_at AND retention_until IS NOT NULL)
    OR (consumed_at IS NULL AND expired_at IS NULL AND invalidated_at IS NOT NULL AND terminal_reason IS NOT NULL AND terminal_reason IN ('identity_epoch_advanced','operator_disabled','security_quarantine','superseded') AND terminal_at IS NOT NULL AND terminal_at = invalidated_at AND retention_until IS NOT NULL)
  ),
  CONSTRAINT node_enrollment_grants_terminal_timestamp CHECK (terminal_at IS NULL OR terminal_at >= created_at),
  CONSTRAINT node_enrollment_grants_retention_window CHECK ((retention_until IS NULL AND terminal_at IS NULL) OR (retention_until IS NOT NULL AND terminal_at IS NOT NULL AND retention_until >= terminal_at + interval '30 days'))
);

CREATE TABLE nodecontrol.node_certificates (
  certificate_id uuid CONSTRAINT node_certificates_certificate_id_not_null NOT NULL,
  issuance_id uuid CONSTRAINT node_certificates_issuance_id_not_null NOT NULL,
  authority_operation_id uuid CONSTRAINT node_certificates_authority_operation_id_not_null NOT NULL,
  authority_epoch bigint CONSTRAINT node_certificates_authority_epoch_not_null NOT NULL,
  authority_sequence bigint CONSTRAINT node_certificates_authority_sequence_not_null NOT NULL,
  node_id uuid CONSTRAINT node_certificates_node_id_not_null NOT NULL,
  identity_epoch bigint CONSTRAINT node_certificates_identity_epoch_not_null NOT NULL,
  lineage_id uuid CONSTRAINT node_certificates_lineage_id_not_null NOT NULL,
  issuer_id text COLLATE "C" CONSTRAINT node_certificates_issuer_id_not_null NOT NULL,
  serial_bytes bytea CONSTRAINT node_certificates_serial_bytes_not_null NOT NULL,
  leaf_der bytea CONSTRAINT node_certificates_leaf_der_not_null NOT NULL,
  leaf_der_sha256 bytea CONSTRAINT node_certificates_leaf_der_sha256_not_null NOT NULL,
  public_key_sha256 bytea CONSTRAINT node_certificates_public_key_sha256_not_null NOT NULL,
  chain_der_sha256 bytea CONSTRAINT node_certificates_chain_der_sha256_not_null NOT NULL,
  valid_from timestamp with time zone CONSTRAINT node_certificates_valid_from_not_null NOT NULL,
  valid_until timestamp with time zone CONSTRAINT node_certificates_valid_until_not_null NOT NULL,
  status text COLLATE "C" CONSTRAINT node_certificates_status_not_null NOT NULL,
  revoke_authority_operation_id uuid,
  revoke_authority_epoch bigint,
  revoke_authority_sequence bigint,
  revoked_at timestamp with time zone,
  revoke_reason text COLLATE "C",
  created_at timestamp with time zone CONSTRAINT node_certificates_created_at_not_null NOT NULL,
  updated_at timestamp with time zone CONSTRAINT node_certificates_updated_at_not_null NOT NULL,
  retention_until timestamp with time zone CONSTRAINT node_certificates_retention_until_not_null NOT NULL,
  CONSTRAINT node_certificates_pkey PRIMARY KEY (certificate_id),
  CONSTRAINT node_certificates_authority_epoch_range CHECK (authority_epoch BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_certificates_authority_sequence_range CHECK (authority_sequence BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_certificates_authority_fence_fk FOREIGN KEY (authority_operation_id, authority_epoch, authority_sequence) REFERENCES nodecontrol.control_plane_authority_fences(operation_id, authority_epoch, authority_sequence) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_certificates_issuance_key UNIQUE (issuance_id),
  CONSTRAINT node_certificates_issuance_fk FOREIGN KEY (issuance_id) REFERENCES nodecontrol.node_certificate_issuances(issuance_id) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_certificates_node_fk FOREIGN KEY (node_id) REFERENCES nodecontrol.node_inventory(node_id) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_certificates_revoke_authority_fence_fk FOREIGN KEY (revoke_authority_operation_id, revoke_authority_epoch, revoke_authority_sequence) REFERENCES nodecontrol.control_plane_authority_fences(operation_id, authority_epoch, authority_sequence) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_certificates_identity_epoch_range CHECK (identity_epoch BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_certificates_issuer_id_format CHECK (octet_length(issuer_id) BETWEEN 1 AND 128 AND issuer_id ~ '^[A-Za-z0-9:_-]{1,128}$'),
  CONSTRAINT node_certificates_serial_length CHECK (octet_length(serial_bytes) BETWEEN 1 AND 20),
  CONSTRAINT node_certificates_leaf_der_length CHECK (octet_length(leaf_der) BETWEEN 1 AND 65536),
  CONSTRAINT node_certificates_leaf_der_sha256_length CHECK (octet_length(leaf_der_sha256) = 32),
  CONSTRAINT node_certificates_public_key_sha256_length CHECK (octet_length(public_key_sha256) = 32),
  CONSTRAINT node_certificates_chain_der_sha256_length CHECK (octet_length(chain_der_sha256) = 32),
  CONSTRAINT node_certificates_validity_order CHECK (valid_until > valid_from),
  CONSTRAINT node_certificates_status_enum CHECK (status IN ('active','recovery_pending','recovery_limited','revoked')),
  CONSTRAINT node_certificates_revoke_all_or_none CHECK (num_nonnulls(revoke_authority_operation_id,revoke_authority_epoch,revoke_authority_sequence,revoked_at,revoke_reason) IN (0,5)),
  CONSTRAINT node_certificates_revoke_authority_epoch_range CHECK (revoke_authority_epoch IS NULL OR revoke_authority_epoch BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_certificates_revoke_authority_sequence_range CHECK (revoke_authority_sequence IS NULL OR revoke_authority_sequence BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_certificates_revoke_reason_enum CHECK (revoke_reason IS NULL OR revoke_reason IN ('scheduled','identity_epoch_advanced','operator_disabled','security_incident','superseded')),
  CONSTRAINT node_certificates_status_revoke_pair CHECK ((status = 'revoked') = (revoked_at IS NOT NULL)),
  CONSTRAINT node_certificates_timestamp_order CHECK (updated_at >= created_at AND (revoked_at IS NULL OR revoked_at >= created_at)),
  CONSTRAINT node_certificates_retention_window CHECK (retention_until >= valid_until + interval '30 days' AND (revoked_at IS NULL OR retention_until >= revoked_at + interval '30 days'))
);

CREATE TABLE nodecontrol.node_security_incidents (
  incident_id uuid CONSTRAINT node_security_incidents_incident_id_not_null NOT NULL,
  authority_operation_id uuid CONSTRAINT node_security_incidents_authority_operation_id_not_null NOT NULL,
  authority_epoch bigint CONSTRAINT node_security_incidents_authority_epoch_not_null NOT NULL,
  authority_sequence bigint CONSTRAINT node_security_incidents_authority_sequence_not_null NOT NULL,
  node_id uuid CONSTRAINT node_security_incidents_node_id_not_null NOT NULL,
  identity_epoch bigint CONSTRAINT node_security_incidents_identity_epoch_not_null NOT NULL,
  fault_subtype text COLLATE "C" CONSTRAINT node_security_incidents_fault_subtype_not_null NOT NULL,
  subtype_slot integer CONSTRAINT node_security_incidents_subtype_slot_not_null NOT NULL,
  status text COLLATE "C" CONSTRAINT node_security_incidents_status_not_null NOT NULL,
  first_evidence_digest bytea CONSTRAINT node_security_incidents_first_evidence_digest_not_null NOT NULL,
  last_evidence_digest bytea CONSTRAINT node_security_incidents_last_evidence_digest_not_null NOT NULL,
  occurrence_count bigint CONSTRAINT node_security_incidents_occurrence_count_not_null NOT NULL,
  trust_context_digest bytea CONSTRAINT node_security_incidents_trust_context_digest_not_null NOT NULL,
  first_occurred_at timestamp with time zone CONSTRAINT node_security_incidents_first_occurred_at_not_null NOT NULL,
  last_occurred_at timestamp with time zone CONSTRAINT node_security_incidents_last_occurred_at_not_null NOT NULL,
  resolution_authority_operation_id uuid,
  resolution_authority_epoch bigint,
  resolution_authority_sequence bigint,
  remediation_digest bytea,
  resolution_at timestamp with time zone,
  retention_until timestamp with time zone,
  CONSTRAINT node_security_incidents_pkey PRIMARY KEY (incident_id),
  CONSTRAINT node_security_incidents_authority_epoch_range CHECK (authority_epoch BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_security_incidents_authority_sequence_range CHECK (authority_sequence BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_security_incidents_authority_fence_fk FOREIGN KEY (authority_operation_id, authority_epoch, authority_sequence) REFERENCES nodecontrol.control_plane_authority_fences(operation_id, authority_epoch, authority_sequence) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_security_incidents_node_fk FOREIGN KEY (node_id) REFERENCES nodecontrol.node_inventory(node_id) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_security_incidents_resolution_authority_fence_fk FOREIGN KEY (resolution_authority_operation_id, resolution_authority_epoch, resolution_authority_sequence) REFERENCES nodecontrol.control_plane_authority_fences(operation_id, authority_epoch, authority_sequence) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_security_incidents_identity_epoch_range CHECK (identity_epoch BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_security_incidents_fault_subtype_enum CHECK (fault_subtype IN ('identity_compromise','online_signer_equivocation','metadata_rollback','root_rollback','root_equivocation','unverified_client_highwater_conflict','client_highwater_ahead','server_trust_bundle_conflict','trusted_time_rollback_or_unavailable','local_state_corruption_or_rollback','release_or_process_integrity','profile_binding_mismatch','incident_overflow')),
  CONSTRAINT node_security_incidents_status_enum CHECK (status IN ('open','resolution_pending_agent_ack','overflow','resolved')),
  CONSTRAINT node_security_incidents_subtype_slot_mapping CHECK ((fault_subtype = 'identity_compromise' AND subtype_slot = 1) OR (fault_subtype = 'online_signer_equivocation' AND subtype_slot = 2) OR (fault_subtype = 'metadata_rollback' AND subtype_slot = 3) OR (fault_subtype = 'root_rollback' AND subtype_slot = 4) OR (fault_subtype = 'root_equivocation' AND subtype_slot = 5) OR (fault_subtype = 'unverified_client_highwater_conflict' AND subtype_slot = 6) OR (fault_subtype = 'client_highwater_ahead' AND subtype_slot = 7) OR (fault_subtype = 'server_trust_bundle_conflict' AND subtype_slot = 8) OR (fault_subtype = 'trusted_time_rollback_or_unavailable' AND subtype_slot = 9) OR (fault_subtype = 'local_state_corruption_or_rollback' AND subtype_slot = 10) OR (fault_subtype = 'release_or_process_integrity' AND subtype_slot = 11) OR (fault_subtype = 'profile_binding_mismatch' AND subtype_slot = 12) OR (fault_subtype = 'incident_overflow' AND subtype_slot = 16)),
  CONSTRAINT node_security_incidents_overflow_status CHECK ((fault_subtype = 'incident_overflow' AND status IN ('overflow','resolution_pending_agent_ack','resolved')) OR (fault_subtype <> 'incident_overflow' AND status IN ('open','resolution_pending_agent_ack','resolved'))),
  CONSTRAINT node_security_incidents_first_evidence_digest_length CHECK (octet_length(first_evidence_digest) = 32),
  CONSTRAINT node_security_incidents_last_evidence_digest_length CHECK (octet_length(last_evidence_digest) = 32),
  CONSTRAINT node_security_incidents_trust_context_digest_length CHECK (octet_length(trust_context_digest) = 32),
  CONSTRAINT node_security_incidents_occurrence_count_range CHECK (occurrence_count BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_security_incidents_occurrence_time_order CHECK (last_occurred_at >= first_occurred_at),
  CONSTRAINT node_security_incidents_resolution_all_or_none CHECK (num_nonnulls(resolution_authority_operation_id,resolution_authority_epoch,resolution_authority_sequence,remediation_digest,resolution_at) IN (0,5)),
  CONSTRAINT node_security_incidents_resolution_authority_epoch_range CHECK (resolution_authority_epoch IS NULL OR resolution_authority_epoch BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_security_incidents_resolution_authority_sequence_range CHECK (resolution_authority_sequence IS NULL OR resolution_authority_sequence BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_security_incidents_remediation_digest_length CHECK (remediation_digest IS NULL OR octet_length(remediation_digest) = 32),
  CONSTRAINT node_security_incidents_resolution_status CHECK ((status IN ('open','overflow') AND resolution_at IS NULL AND retention_until IS NULL) OR (status = 'resolution_pending_agent_ack' AND resolution_at IS NOT NULL AND retention_until IS NULL) OR (status = 'resolved' AND resolution_at IS NOT NULL AND retention_until IS NOT NULL)),
  CONSTRAINT node_security_incidents_retention_window CHECK (retention_until IS NULL OR retention_until >= resolution_at + interval '180 days')
);

CREATE TABLE nodecontrol.node_security_fault_receipts (
  receipt_id uuid CONSTRAINT node_security_fault_receipts_receipt_id_not_null NOT NULL,
  authority_operation_id uuid CONSTRAINT node_security_fault_receipts_authority_operation_id_not_null NOT NULL,
  authority_epoch bigint CONSTRAINT node_security_fault_receipts_authority_epoch_not_null NOT NULL,
  authority_sequence bigint CONSTRAINT node_security_fault_receipts_authority_sequence_not_null NOT NULL,
  node_id uuid CONSTRAINT node_security_fault_receipts_node_id_not_null NOT NULL,
  identity_epoch bigint CONSTRAINT node_security_fault_receipts_identity_epoch_not_null NOT NULL,
  local_fault_id uuid CONSTRAINT node_security_fault_receipts_local_fault_id_not_null NOT NULL,
  request_digest bytea CONSTRAINT node_security_fault_receipts_request_digest_not_null NOT NULL,
  fault_subtype text COLLATE "C" CONSTRAINT node_security_fault_receipts_fault_subtype_not_null NOT NULL,
  evidence_digest bytea CONSTRAINT node_security_fault_receipts_evidence_digest_not_null NOT NULL,
  agent_boot_id uuid CONSTRAINT node_security_fault_receipts_agent_boot_id_not_null NOT NULL,
  incident_id uuid CONSTRAINT node_security_fault_receipts_incident_id_not_null NOT NULL,
  supervisor_boot_id uuid,
  supervisor_fault_id uuid,
  supervisor_evidence_digest bytea,
  local_binding_slot integer CONSTRAINT node_security_fault_receipts_local_binding_slot_not_null NOT NULL,
  supervisor_binding_slot integer,
  result text COLLATE "C" CONSTRAINT node_security_fault_receipts_result_not_null NOT NULL,
  delivery_status text COLLATE "C" CONSTRAINT node_security_fault_receipts_delivery_status_not_null NOT NULL,
  binding_status text COLLATE "C" CONSTRAINT node_security_fault_receipts_binding_status_not_null NOT NULL,
  delivered_at timestamp with time zone,
  cleared_at timestamp with time zone,
  clear_attestation_digest bytea,
  created_at timestamp with time zone CONSTRAINT node_security_fault_receipts_created_at_not_null NOT NULL,
  updated_at timestamp with time zone CONSTRAINT node_security_fault_receipts_updated_at_not_null NOT NULL,
  CONSTRAINT node_security_fault_receipts_pkey PRIMARY KEY (receipt_id),
  CONSTRAINT node_security_fault_receipts_authority_epoch_range CHECK (authority_epoch BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_security_fault_receipts_authority_sequence_range CHECK (authority_sequence BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_security_fault_receipts_authority_fence_fk FOREIGN KEY (authority_operation_id, authority_epoch, authority_sequence) REFERENCES nodecontrol.control_plane_authority_fences(operation_id, authority_epoch, authority_sequence) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_security_fault_receipts_node_fk FOREIGN KEY (node_id) REFERENCES nodecontrol.node_inventory(node_id) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_security_fault_receipts_incident_fk FOREIGN KEY (incident_id) REFERENCES nodecontrol.node_security_incidents(incident_id) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_security_fault_receipts_identity_epoch_range CHECK (identity_epoch BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_security_fault_receipts_request_digest_length CHECK (octet_length(request_digest) = 32),
  CONSTRAINT node_security_fault_receipts_evidence_digest_length CHECK (octet_length(evidence_digest) = 32),
  CONSTRAINT node_security_fault_receipts_fault_subtype_enum CHECK (fault_subtype IN ('identity_compromise','online_signer_equivocation','metadata_rollback','root_rollback','root_equivocation','unverified_client_highwater_conflict','client_highwater_ahead','server_trust_bundle_conflict','trusted_time_rollback_or_unavailable','local_state_corruption_or_rollback','release_or_process_integrity','profile_binding_mismatch','incident_overflow')),
  CONSTRAINT node_security_fault_receipts_supervisor_binding_all_or_none CHECK (num_nonnulls(supervisor_boot_id,supervisor_fault_id,supervisor_evidence_digest,supervisor_binding_slot) IN (0,4)),
  CONSTRAINT node_security_fault_receipts_supervisor_evidence_digest_length CHECK (supervisor_evidence_digest IS NULL OR octet_length(supervisor_evidence_digest) = 32),
  CONSTRAINT node_security_fault_receipts_local_binding_slot_range CHECK (local_binding_slot BETWEEN 1 AND 64),
  CONSTRAINT node_security_fault_receipts_supervisor_binding_slot_range CHECK (supervisor_binding_slot IS NULL OR supervisor_binding_slot BETWEEN 1 AND 16),
  CONSTRAINT node_security_fault_receipts_result_enum CHECK (result IN ('accepted')),
  CONSTRAINT node_security_fault_receipts_delivery_status_enum CHECK (delivery_status IN ('pending','deliverable')),
  CONSTRAINT node_security_fault_receipts_binding_status_enum CHECK (binding_status IN ('active','cleared')),
  CONSTRAINT node_security_fault_receipts_delivery_time_pair CHECK ((delivery_status = 'pending' AND delivered_at IS NULL) OR (delivery_status = 'deliverable' AND delivered_at IS NOT NULL)),
  CONSTRAINT node_security_fault_receipts_clear_group CHECK ((binding_status = 'active' AND cleared_at IS NULL AND clear_attestation_digest IS NULL) OR (binding_status = 'cleared' AND cleared_at IS NOT NULL AND clear_attestation_digest IS NOT NULL)),
  CONSTRAINT node_security_fault_receipts_clear_attestation_digest_length CHECK (clear_attestation_digest IS NULL OR octet_length(clear_attestation_digest) = 32),
  CONSTRAINT node_security_fault_receipts_timestamp_order CHECK (updated_at >= created_at AND (delivered_at IS NULL OR delivered_at >= created_at) AND (cleared_at IS NULL OR cleared_at >= created_at)),
  CONSTRAINT node_security_fault_receipts_local_fault_key UNIQUE (node_id, identity_epoch, local_fault_id),
  CONSTRAINT node_security_fault_receipts_request_digest_key UNIQUE (node_id, request_digest)
);

CREATE TABLE nodecontrol.node_recovery_sessions (
  recovery_id uuid CONSTRAINT node_recovery_sessions_recovery_id_not_null NOT NULL,
  authority_operation_id uuid CONSTRAINT node_recovery_sessions_authority_operation_id_not_null NOT NULL,
  authority_epoch bigint CONSTRAINT node_recovery_sessions_authority_epoch_not_null NOT NULL,
  authority_sequence bigint CONSTRAINT node_recovery_sessions_authority_sequence_not_null NOT NULL,
  node_id uuid CONSTRAINT node_recovery_sessions_node_id_not_null NOT NULL,
  identity_epoch bigint CONSTRAINT node_recovery_sessions_identity_epoch_not_null NOT NULL,
  reason text COLLATE "C" CONSTRAINT node_recovery_sessions_reason_not_null NOT NULL,
  version bigint CONSTRAINT node_recovery_sessions_version_not_null NOT NULL,
  status text COLLATE "C" CONSTRAINT node_recovery_sessions_status_not_null NOT NULL,
  incident_set_digest bytea CONSTRAINT node_recovery_sessions_incident_set_digest_not_null NOT NULL,
  resume_operator_state text COLLATE "C" CONSTRAINT node_recovery_sessions_resume_operator_state_not_null NOT NULL,
  recovery_certificate_id uuid,
  attestation_digest bytea,
  terminal_at timestamp with time zone,
  created_at timestamp with time zone CONSTRAINT node_recovery_sessions_created_at_not_null NOT NULL,
  updated_at timestamp with time zone CONSTRAINT node_recovery_sessions_updated_at_not_null NOT NULL,
  retention_until timestamp with time zone,
  CONSTRAINT node_recovery_sessions_pkey PRIMARY KEY (recovery_id),
  CONSTRAINT node_recovery_sessions_authority_epoch_range CHECK (authority_epoch BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_recovery_sessions_authority_sequence_range CHECK (authority_sequence BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_recovery_sessions_authority_fence_fk FOREIGN KEY (authority_operation_id, authority_epoch, authority_sequence) REFERENCES nodecontrol.control_plane_authority_fences(operation_id, authority_epoch, authority_sequence) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_recovery_sessions_node_fk FOREIGN KEY (node_id) REFERENCES nodecontrol.node_inventory(node_id) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_recovery_sessions_certificate_fk FOREIGN KEY (recovery_certificate_id) REFERENCES nodecontrol.node_certificates(certificate_id) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_recovery_sessions_identity_epoch_range CHECK (identity_epoch BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_recovery_sessions_reason_enum CHECK (reason IN ('identity_compromise','administrative_disable','retire','authority_restore','security_incident')),
  CONSTRAINT node_recovery_sessions_version_range CHECK (version BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_recovery_sessions_status_enum CHECK (status IN ('pending','completed','superseded')),
  CONSTRAINT node_recovery_sessions_incident_set_digest_length CHECK (octet_length(incident_set_digest) = 32),
  CONSTRAINT node_recovery_sessions_resume_operator_state_enum CHECK (resume_operator_state IN ('enabled','draining','disabled')),
  CONSTRAINT node_recovery_sessions_restore_resume_disabled CHECK (reason NOT IN ('retire','authority_restore') OR resume_operator_state = 'disabled'),
  CONSTRAINT node_recovery_sessions_attestation_digest_length CHECK (attestation_digest IS NULL OR octet_length(attestation_digest) = 32),
  CONSTRAINT node_recovery_sessions_terminal_group CHECK ((status = 'pending' AND terminal_at IS NULL AND retention_until IS NULL) OR (status IN ('completed','superseded') AND terminal_at IS NOT NULL AND retention_until IS NOT NULL)),
  CONSTRAINT node_recovery_sessions_timestamp_order CHECK (updated_at >= created_at AND (terminal_at IS NULL OR terminal_at >= created_at)),
  CONSTRAINT node_recovery_sessions_retention_window CHECK (retention_until IS NULL OR retention_until >= terminal_at + interval '180 days'),
  CONSTRAINT node_recovery_sessions_node_version_key UNIQUE (node_id, version)
);

CREATE TABLE nodecontrol.node_restore_reauthorization_approvals (
  approval_id uuid CONSTRAINT node_restore_reauthorization_approvals_approval_id_not_null NOT NULL,
  authority_operation_id uuid CONSTRAINT node_restore_reauthorization_approvals_authority_opera_6350300c NOT NULL,
  authority_epoch bigint CONSTRAINT node_restore_reauthorization_approvals_authority_epoch_not_null NOT NULL,
  authority_sequence bigint CONSTRAINT node_restore_reauthorization_approvals_authority_seque_3d56c216 NOT NULL,
  node_id uuid CONSTRAINT node_restore_reauthorization_approvals_node_id_not_null NOT NULL,
  recovery_id uuid CONSTRAINT node_restore_reauthorization_approvals_recovery_id_not_null NOT NULL,
  effect_digest bytea CONSTRAINT node_restore_reauthorization_approvals_effect_digest_not_null NOT NULL,
  scope_digest bytea CONSTRAINT node_restore_reauthorization_approvals_scope_digest_not_null NOT NULL,
  role text COLLATE "C" CONSTRAINT node_restore_reauthorization_approvals_role_not_null NOT NULL,
  operator_id text COLLATE "C" CONSTRAINT node_restore_reauthorization_approvals_operator_id_not_null NOT NULL,
  credential_digest bytea CONSTRAINT node_restore_reauthorization_approvals_credential_dige_7f5f14bf NOT NULL,
  leaf_der_sha256 bytea CONSTRAINT node_restore_reauthorization_approvals_leaf_der_sha256_not_null NOT NULL,
  operator_authority_epoch bigint CONSTRAINT node_restore_reauthorization_approvals_operator_author_55f751ab NOT NULL,
  operator_authority_sequence bigint CONSTRAINT node_restore_reauthorization_approvals_operator_author_1d815be3 NOT NULL,
  authorizer_version bigint CONSTRAINT node_restore_reauthorization_approvals_authorizer_vers_a269a2c3 NOT NULL,
  security_admin_binding_digest bytea CONSTRAINT node_restore_reauthorization_approvals_security_admin__9b45c6e2 NOT NULL,
  pop_scope text COLLATE "C" CONSTRAINT node_restore_reauthorization_approvals_pop_scope_not_null NOT NULL,
  evidence_completed_at timestamp with time zone CONSTRAINT node_restore_reauthorization_approvals_evidence_comple_424ad5ea NOT NULL,
  credential_expires_at timestamp with time zone CONSTRAINT node_restore_reauthorization_approvals_credential_expi_0abe35ff NOT NULL,
  created_at timestamp with time zone CONSTRAINT node_restore_reauthorization_approvals_created_at_not_null NOT NULL,
  expires_at timestamp with time zone CONSTRAINT node_restore_reauthorization_approvals_expires_at_not_null NOT NULL,
  status text COLLATE "C" CONSTRAINT node_restore_reauthorization_approvals_status_not_null NOT NULL,
  terminal_at timestamp with time zone,
  retention_until timestamp with time zone,
  CONSTRAINT node_restore_reauthorization_approvals_pkey PRIMARY KEY (approval_id),
  CONSTRAINT node_restore_reauthorization_approvals_authority_epoch_range CHECK (authority_epoch BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_restore_reauthorization_approvals_authority_sequence_range CHECK (authority_sequence BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_restore_reauthorization_approvals_authority_fence_fk FOREIGN KEY (authority_operation_id, authority_epoch, authority_sequence) REFERENCES nodecontrol.control_plane_authority_fences(operation_id, authority_epoch, authority_sequence) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_restore_reauthorization_approvals_node_fk FOREIGN KEY (node_id) REFERENCES nodecontrol.node_inventory(node_id) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_restore_reauthorization_approvals_recovery_fk FOREIGN KEY (recovery_id) REFERENCES nodecontrol.node_recovery_sessions(recovery_id) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_restore_reauthorization_approvals_effect_digest_length CHECK (octet_length(effect_digest) = 32),
  CONSTRAINT node_restore_reauthorization_approvals_scope_digest_length CHECK (octet_length(scope_digest) = 32),
  CONSTRAINT node_restore_reauthorization_approvals_role_enum CHECK (role IN ('proposal','approval')),
  CONSTRAINT node_restore_reauthorization_approvals_operator_id_format CHECK (octet_length(operator_id) BETWEEN 1 AND 128 AND operator_id ~ '^[A-Za-z0-9:_-]{1,128}$'),
  CONSTRAINT node_restore_reauthorization_approvals_credential_digest_length CHECK (octet_length(credential_digest) = 32),
  CONSTRAINT node_restore_reauthorization_approvals_leaf_der_sha256_length CHECK (octet_length(leaf_der_sha256) = 32),
  CONSTRAINT node_restore_reauthorization_approvals_operator_author_3ab23c85 CHECK (operator_authority_epoch BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_restore_reauthorization_approvals_operator_author_2f15ac9d CHECK (operator_authority_sequence BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_restore_reauthorization_approvals_authorizer_version_range CHECK (authorizer_version BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_restore_reauthorization_approvals_security_admin__8e32de15 CHECK (octet_length(security_admin_binding_digest) = 32),
  CONSTRAINT node_restore_reauthorization_approvals_pop_scope_format CHECK (octet_length(pop_scope) BETWEEN 1 AND 32 AND pop_scope ~ '^[a-z0-9](?:[a-z0-9-]{0,30}[a-z0-9])?$'),
  CONSTRAINT node_restore_reauthorization_approvals_expiry_bound CHECK (expires_at > created_at AND expires_at <= created_at + interval '15 minutes' AND expires_at <= credential_expires_at AND expires_at <= evidence_completed_at + interval '15 minutes'),
  CONSTRAINT node_restore_reauthorization_approvals_status_enum CHECK (status IN ('pending','consumed','superseded')),
  CONSTRAINT node_restore_reauthorization_approvals_terminal_group CHECK ((status = 'pending' AND terminal_at IS NULL AND retention_until IS NULL) OR (status IN ('consumed','superseded') AND terminal_at IS NOT NULL AND retention_until IS NOT NULL)),
  CONSTRAINT node_restore_reauthorization_approvals_retention_window CHECK (retention_until IS NULL OR retention_until >= terminal_at + interval '180 days')
);

CREATE TABLE nodecontrol.node_state_signing_intents (
  signing_id uuid CONSTRAINT node_state_signing_intents_signing_id_not_null NOT NULL,
  authority_operation_id uuid CONSTRAINT node_state_signing_intents_authority_operation_id_not_null NOT NULL,
  authority_epoch bigint CONSTRAINT node_state_signing_intents_authority_epoch_not_null NOT NULL,
  authority_sequence bigint CONSTRAINT node_state_signing_intents_authority_sequence_not_null NOT NULL,
  node_id uuid CONSTRAINT node_state_signing_intents_node_id_not_null NOT NULL,
  signing_kind text COLLATE "C" CONSTRAINT node_state_signing_intents_signing_kind_not_null NOT NULL,
  idempotency_key_digest bytea CONSTRAINT node_state_signing_intents_idempotency_key_digest_not_null NOT NULL,
  base_generation bigint CONSTRAINT node_state_signing_intents_base_generation_not_null NOT NULL,
  reserved_generation bigint CONSTRAINT node_state_signing_intents_reserved_generation_not_null NOT NULL,
  canonical_payload bytea CONSTRAINT node_state_signing_intents_canonical_payload_not_null NOT NULL,
  payload_digest bytea CONSTRAINT node_state_signing_intents_payload_digest_not_null NOT NULL,
  root_version bigint CONSTRAINT node_state_signing_intents_root_version_not_null NOT NULL,
  metadata_version bigint CONSTRAINT node_state_signing_intents_metadata_version_not_null NOT NULL,
  expected_key_id bytea CONSTRAINT node_state_signing_intents_expected_key_id_not_null NOT NULL,
  expected_public_key_digest bytea CONSTRAINT node_state_signing_intents_expected_public_key_digest_not_null NOT NULL,
  captured_inventory_version bigint CONSTRAINT node_state_signing_intents_captured_inventory_version_not_null NOT NULL,
  captured_identity_epoch bigint CONSTRAINT node_state_signing_intents_captured_identity_epoch_not_null NOT NULL,
  captured_security_version bigint CONSTRAINT node_state_signing_intents_captured_security_version_not_null NOT NULL,
  recovery_id uuid,
  recovery_reason text COLLATE "C",
  recovery_session_version bigint,
  recovery_session_status text COLLATE "C",
  recovery_incident_set_digest bytea,
  recovery_local_bindings_digest bytea,
  recovery_supervisor_bindings_digest bytea,
  recovery_remediation_digest bytea,
  recovery_required_action text COLLATE "C",
  signature bytea,
  signature_verified_at timestamp with time zone,
  activation_deadline timestamp with time zone CONSTRAINT node_state_signing_intents_activation_deadline_not_null NOT NULL,
  status text COLLATE "C" CONSTRAINT node_state_signing_intents_status_not_null NOT NULL,
  failure_reason text COLLATE "C",
  created_at timestamp with time zone CONSTRAINT node_state_signing_intents_created_at_not_null NOT NULL,
  updated_at timestamp with time zone CONSTRAINT node_state_signing_intents_updated_at_not_null NOT NULL,
  terminal_at timestamp with time zone,
  CONSTRAINT node_state_signing_intents_pkey PRIMARY KEY (signing_id),
  CONSTRAINT node_state_signing_intents_authority_epoch_range CHECK (authority_epoch BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_state_signing_intents_authority_sequence_range CHECK (authority_sequence BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_state_signing_intents_authority_fence_fk FOREIGN KEY (authority_operation_id, authority_epoch, authority_sequence) REFERENCES nodecontrol.control_plane_authority_fences(operation_id, authority_epoch, authority_sequence) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_state_signing_intents_node_fk FOREIGN KEY (node_id) REFERENCES nodecontrol.node_inventory(node_id) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_state_signing_intents_recovery_fk FOREIGN KEY (recovery_id) REFERENCES nodecontrol.node_recovery_sessions(recovery_id) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_state_signing_intents_signing_kind_enum CHECK (signing_kind IN ('desired','recovery')),
  CONSTRAINT node_state_signing_intents_idempotency_key_digest_length CHECK (octet_length(idempotency_key_digest) = 32),
  CONSTRAINT node_state_signing_intents_base_generation_range CHECK (base_generation BETWEEN 0 AND 9223372036854775807),
  CONSTRAINT node_state_signing_intents_reserved_generation_range CHECK (reserved_generation BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_state_signing_intents_generation_advance CHECK (reserved_generation > base_generation),
  CONSTRAINT node_state_signing_intents_canonical_payload_length CHECK (octet_length(canonical_payload) BETWEEN 1 AND 65536),
  CONSTRAINT node_state_signing_intents_payload_digest_length CHECK (octet_length(payload_digest) = 32),
  CONSTRAINT node_state_signing_intents_root_version_range CHECK (root_version BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_state_signing_intents_metadata_version_range CHECK (metadata_version BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_state_signing_intents_captured_inventory_version_range CHECK (captured_inventory_version BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_state_signing_intents_captured_security_version_range CHECK (captured_security_version BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_state_signing_intents_captured_identity_epoch_range CHECK (captured_identity_epoch BETWEEN 0 AND 9223372036854775807),
  CONSTRAINT node_state_signing_intents_expected_key_id_length CHECK (octet_length(expected_key_id) = 32),
  CONSTRAINT node_state_signing_intents_expected_public_key_digest_length CHECK (octet_length(expected_public_key_digest) = 32),
  CONSTRAINT node_state_signing_intents_recovery_reason_enum CHECK (recovery_reason IS NULL OR recovery_reason IN ('identity_compromise','administrative_disable','retire','authority_restore','security_incident')),
  CONSTRAINT node_state_signing_intents_recovery_session_version_range CHECK (recovery_session_version IS NULL OR recovery_session_version BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_state_signing_intents_recovery_session_status_enum CHECK (recovery_session_status IS NULL OR recovery_session_status IN ('pending','completed','superseded')),
  CONSTRAINT node_state_signing_intents_recovery_incident_set_digest_length CHECK (recovery_incident_set_digest IS NULL OR octet_length(recovery_incident_set_digest) = 32),
  CONSTRAINT node_state_signing_intents_recovery_local_bindings_dig_35bb2748 CHECK (recovery_local_bindings_digest IS NULL OR octet_length(recovery_local_bindings_digest) = 32),
  CONSTRAINT node_state_signing_intents_recovery_supervisor_binding_142fbdeb CHECK (recovery_supervisor_bindings_digest IS NULL OR octet_length(recovery_supervisor_bindings_digest) = 32),
  CONSTRAINT node_state_signing_intents_recovery_remediation_digest_length CHECK (recovery_remediation_digest IS NULL OR octet_length(recovery_remediation_digest) = 32),
  CONSTRAINT node_state_signing_intents_recovery_required_action_enum CHECK (recovery_required_action IS NULL OR recovery_required_action IN ('hold_stopped','submit_recovery_attestation','clear_security_latches')),
  CONSTRAINT node_state_signing_intents_kind_recovery_group CHECK ((signing_kind = 'desired' AND num_nonnulls(recovery_id,recovery_reason,recovery_session_version,recovery_session_status,recovery_incident_set_digest,recovery_local_bindings_digest,recovery_supervisor_bindings_digest,recovery_required_action,recovery_remediation_digest) = 0) OR (signing_kind = 'recovery' AND num_nonnulls(recovery_id,recovery_reason,recovery_session_version,recovery_session_status,recovery_incident_set_digest,recovery_local_bindings_digest,recovery_supervisor_bindings_digest,recovery_required_action) = 8)),
  CONSTRAINT node_state_signing_intents_remediation_action_pair CHECK ((recovery_required_action = 'clear_security_latches') = (recovery_remediation_digest IS NOT NULL)),
  CONSTRAINT node_state_signing_intents_signature_length CHECK (signature IS NULL OR octet_length(signature) = 64),
  CONSTRAINT node_state_signing_intents_signature_pair CHECK ((signature IS NULL) = (signature_verified_at IS NULL)),
  CONSTRAINT node_state_signing_intents_activation_deadline_order CHECK (activation_deadline > created_at),
  CONSTRAINT node_state_signing_intents_status_enum CHECK (status IN ('pending','active','failed','superseded')),
  CONSTRAINT node_state_signing_intents_failure_reason_enum CHECK (failure_reason IS NULL OR failure_reason IN ('deadline_expired','validation_failed','signer_failed','superseded','authority_aborted')),
  CONSTRAINT node_state_signing_intents_status_terminal_group CHECK ((status = 'pending' AND terminal_at IS NULL AND failure_reason IS NULL) OR (status = 'active' AND terminal_at IS NOT NULL AND failure_reason IS NULL AND signature IS NOT NULL) OR (status IN ('failed','superseded') AND terminal_at IS NOT NULL AND failure_reason IS NOT NULL)),
  CONSTRAINT node_state_signing_intents_timestamp_order CHECK (updated_at >= created_at AND (signature_verified_at IS NULL OR signature_verified_at >= created_at) AND (terminal_at IS NULL OR terminal_at >= created_at)),
  CONSTRAINT node_state_signing_intents_node_kind_generation_key UNIQUE (node_id, signing_kind, reserved_generation)
);

CREATE TABLE nodecontrol.node_root_metadata_publish_intents (
  publish_id uuid CONSTRAINT node_root_metadata_publish_intents_publish_id_not_null NOT NULL,
  authority_operation_id uuid CONSTRAINT node_root_metadata_publish_intents_authority_operation_fe8a9f18 NOT NULL,
  authority_epoch bigint CONSTRAINT node_root_metadata_publish_intents_authority_epoch_not_null NOT NULL,
  authority_sequence bigint CONSTRAINT node_root_metadata_publish_intents_authority_sequence_not_null NOT NULL,
  publish_kind text COLLATE "C" CONSTRAINT node_root_metadata_publish_intents_publish_kind_not_null NOT NULL,
  reason text COLLATE "C" CONSTRAINT node_root_metadata_publish_intents_reason_not_null NOT NULL,
  incident_id uuid,
  base_root_version bigint CONSTRAINT node_root_metadata_publish_intents_base_root_version_not_null NOT NULL,
  base_metadata_version bigint CONSTRAINT node_root_metadata_publish_intents_base_metadata_versi_6d4ab851 NOT NULL,
  reserved_version bigint CONSTRAINT node_root_metadata_publish_intents_reserved_version_not_null NOT NULL,
  canonical_payload bytea CONSTRAINT node_root_metadata_publish_intents_canonical_payload_not_null NOT NULL,
  payload_digest bytea CONSTRAINT node_root_metadata_publish_intents_payload_digest_not_null NOT NULL,
  key_set_digest bytea CONSTRAINT node_root_metadata_publish_intents_key_set_digest_not_null NOT NULL,
  current_key_ids bytea[] CONSTRAINT node_root_metadata_publish_intents_current_key_ids_not_null NOT NULL,
  new_key_ids bytea[],
  current_threshold integer CONSTRAINT node_root_metadata_publish_intents_current_threshold_not_null NOT NULL,
  new_threshold integer,
  activation_deadline timestamp with time zone CONSTRAINT node_root_metadata_publish_intents_activation_deadline_not_null NOT NULL,
  published_envelope bytea,
  published_envelope_digest bytea,
  status text COLLATE "C" CONSTRAINT node_root_metadata_publish_intents_status_not_null NOT NULL,
  failure_reason text COLLATE "C",
  created_at timestamp with time zone CONSTRAINT node_root_metadata_publish_intents_created_at_not_null NOT NULL,
  updated_at timestamp with time zone CONSTRAINT node_root_metadata_publish_intents_updated_at_not_null NOT NULL,
  terminal_at timestamp with time zone,
  CONSTRAINT node_root_metadata_publish_intents_pkey PRIMARY KEY (publish_id),
  CONSTRAINT node_root_metadata_publish_intents_authority_epoch_range CHECK (authority_epoch BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_root_metadata_publish_intents_authority_sequence_range CHECK (authority_sequence BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_root_metadata_publish_intents_authority_fence_fk FOREIGN KEY (authority_operation_id, authority_epoch, authority_sequence) REFERENCES nodecontrol.control_plane_authority_fences(operation_id, authority_epoch, authority_sequence) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_root_metadata_publish_intents_incident_fk FOREIGN KEY (incident_id) REFERENCES nodecontrol.node_security_incidents(incident_id) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_root_metadata_publish_intents_kind_version_key UNIQUE (publish_kind, reserved_version),
  CONSTRAINT node_root_metadata_publish_intents_publish_kind_enum CHECK (publish_kind IN ('root','metadata')),
  CONSTRAINT node_root_metadata_publish_intents_reason_enum CHECK (reason IN ('normal','root_rotation','emergency_revoke')),
  CONSTRAINT node_root_metadata_publish_intents_incident_reason_pair CHECK ((reason = 'emergency_revoke') = (incident_id IS NOT NULL)),
  CONSTRAINT node_root_metadata_publish_intents_base_root_version_range CHECK (base_root_version BETWEEN 0 AND 9223372036854775807),
  CONSTRAINT node_root_metadata_publish_intents_base_metadata_version_range CHECK (base_metadata_version BETWEEN 0 AND 9223372036854775807),
  CONSTRAINT node_root_metadata_publish_intents_reserved_version_range CHECK (reserved_version BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_root_metadata_publish_intents_version_continuity CHECK ((publish_kind = 'root' AND base_root_version = reserved_version - 1) OR (publish_kind = 'metadata' AND base_metadata_version = reserved_version - 1)),
  CONSTRAINT node_root_metadata_publish_intents_canonical_payload_length CHECK (octet_length(canonical_payload) BETWEEN 1 AND 65536),
  CONSTRAINT node_root_metadata_publish_intents_payload_digest_length CHECK (octet_length(payload_digest) = 32),
  CONSTRAINT node_root_metadata_publish_intents_key_set_digest_length CHECK (octet_length(key_set_digest) = 32),
  CONSTRAINT node_root_metadata_publish_intents_current_key_ids_valid CHECK (nodecontrol.bytea_array_is_sorted_unique_32(current_key_ids)),
  CONSTRAINT node_root_metadata_publish_intents_new_key_ids_valid CHECK (new_key_ids IS NULL OR nodecontrol.bytea_array_is_sorted_unique_32(new_key_ids)),
  CONSTRAINT node_root_metadata_publish_intents_current_threshold_range CHECK (current_threshold BETWEEN 1 AND 5),
  CONSTRAINT node_root_metadata_publish_intents_new_threshold_range CHECK (new_threshold IS NULL OR new_threshold BETWEEN 1 AND 5),
  CONSTRAINT node_root_metadata_publish_intents_threshold_key_coverage CHECK (current_threshold <= cardinality(current_key_ids) AND (new_threshold IS NULL OR new_threshold <= cardinality(new_key_ids))),
  CONSTRAINT node_root_metadata_publish_intents_root_threshold_pair CHECK ((reason = 'root_rotation') = (publish_kind = 'root' AND new_threshold IS NOT NULL AND new_key_ids IS NOT NULL)),
  CONSTRAINT node_root_metadata_publish_intents_activation_deadline_order CHECK (activation_deadline > created_at),
  CONSTRAINT node_root_metadata_publish_intents_published_envelope_length CHECK (published_envelope IS NULL OR octet_length(published_envelope) BETWEEN 1 AND 262144),
  CONSTRAINT node_root_metadata_publish_intents_published_envelope__81c85f22 CHECK (published_envelope_digest IS NULL OR octet_length(published_envelope_digest) = 32),
  CONSTRAINT node_root_metadata_publish_intents_published_envelope_pair CHECK ((published_envelope IS NULL) = (published_envelope_digest IS NULL)),
  CONSTRAINT node_root_metadata_publish_intents_status_enum CHECK (status IN ('pending','active','failed','superseded')),
  CONSTRAINT node_root_metadata_publish_intents_failure_reason_enum CHECK (failure_reason IS NULL OR failure_reason IN ('deadline_expired','validation_failed','threshold_failed','superseded','authority_aborted')),
  CONSTRAINT node_root_metadata_publish_intents_status_terminal_group CHECK ((status = 'pending' AND terminal_at IS NULL AND failure_reason IS NULL) OR (status = 'active' AND terminal_at IS NOT NULL AND failure_reason IS NULL AND published_envelope IS NOT NULL) OR (status IN ('failed','superseded') AND terminal_at IS NOT NULL AND failure_reason IS NOT NULL)),
  CONSTRAINT node_root_metadata_publish_intents_timestamp_order CHECK (updated_at >= created_at AND (terminal_at IS NULL OR terminal_at >= created_at))
);

CREATE TABLE nodecontrol.node_root_metadata_signature_shares (
  publish_id uuid CONSTRAINT node_root_metadata_signature_shares_publish_id_not_null NOT NULL,
  key_id bytea CONSTRAINT node_root_metadata_signature_shares_key_id_not_null NOT NULL,
  physical_key_id bytea CONSTRAINT node_root_metadata_signature_shares_physical_key_id_not_null NOT NULL,
  payload_digest bytea CONSTRAINT node_root_metadata_signature_shares_payload_digest_not_null NOT NULL,
  signature_role text COLLATE "C" CONSTRAINT node_root_metadata_signature_shares_signature_role_not_null NOT NULL,
  signature bytea CONSTRAINT node_root_metadata_signature_shares_signature_not_null NOT NULL,
  verified_at timestamp with time zone CONSTRAINT node_root_metadata_signature_shares_verified_at_not_null NOT NULL,
  CONSTRAINT node_root_metadata_signature_shares_pkey PRIMARY KEY (publish_id, key_id, signature_role),
  CONSTRAINT node_root_metadata_signature_shares_publish_fk FOREIGN KEY (publish_id) REFERENCES nodecontrol.node_root_metadata_publish_intents(publish_id) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_root_metadata_signature_shares_key_id_length CHECK (octet_length(key_id) = 32),
  CONSTRAINT node_root_metadata_signature_shares_physical_key_id_length CHECK (octet_length(physical_key_id) = 32),
  CONSTRAINT node_root_metadata_signature_shares_payload_digest_length CHECK (octet_length(payload_digest) = 32),
  CONSTRAINT node_root_metadata_signature_shares_signature_role_enum CHECK (signature_role IN ('current_root','new_root','metadata')),
  CONSTRAINT node_root_metadata_signature_shares_signature_length CHECK (octet_length(signature) = 64),
  CONSTRAINT node_root_metadata_signature_shares_physical_role_key UNIQUE (publish_id, physical_key_id, signature_role)
);

CREATE TABLE nodecontrol.node_desired_states (
  node_id uuid CONSTRAINT node_desired_states_node_id_not_null NOT NULL,
  generation bigint CONSTRAINT node_desired_states_generation_not_null NOT NULL,
  signing_id uuid CONSTRAINT node_desired_states_signing_id_not_null NOT NULL,
  authority_operation_id uuid CONSTRAINT node_desired_states_authority_operation_id_not_null NOT NULL,
  authority_epoch bigint CONSTRAINT node_desired_states_authority_epoch_not_null NOT NULL,
  authority_sequence bigint CONSTRAINT node_desired_states_authority_sequence_not_null NOT NULL,
  inventory_version bigint CONSTRAINT node_desired_states_inventory_version_not_null NOT NULL,
  resource_envelope_version bigint CONSTRAINT node_desired_states_resource_envelope_version_not_null NOT NULL,
  resource_envelope_digest bytea CONSTRAINT node_desired_states_resource_envelope_digest_not_null NOT NULL,
  root_version bigint CONSTRAINT node_desired_states_root_version_not_null NOT NULL,
  root_publish_id uuid CONSTRAINT node_desired_states_root_publish_id_not_null NOT NULL,
  metadata_version bigint CONSTRAINT node_desired_states_metadata_version_not_null NOT NULL,
  metadata_publish_id uuid CONSTRAINT node_desired_states_metadata_publish_id_not_null NOT NULL,
  signing_key_id bytea CONSTRAINT node_desired_states_signing_key_id_not_null NOT NULL,
  canonical_payload bytea CONSTRAINT node_desired_states_canonical_payload_not_null NOT NULL,
  payload_digest bytea CONSTRAINT node_desired_states_payload_digest_not_null NOT NULL,
  signature bytea CONSTRAINT node_desired_states_signature_not_null NOT NULL,
  issued_at timestamp with time zone CONSTRAINT node_desired_states_issued_at_not_null NOT NULL,
  effective_deadline timestamp with time zone CONSTRAINT node_desired_states_effective_deadline_not_null NOT NULL,
  valid_until timestamp with time zone CONSTRAINT node_desired_states_valid_until_not_null NOT NULL,
  reason text COLLATE "C" CONSTRAINT node_desired_states_reason_not_null NOT NULL,
  created_at timestamp with time zone CONSTRAINT node_desired_states_created_at_not_null NOT NULL,
  retention_until timestamp with time zone CONSTRAINT node_desired_states_retention_until_not_null NOT NULL,
  CONSTRAINT node_desired_states_pkey PRIMARY KEY (node_id, generation),
  CONSTRAINT node_desired_states_authority_epoch_range CHECK (authority_epoch BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_desired_states_authority_sequence_range CHECK (authority_sequence BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_desired_states_authority_fence_fk FOREIGN KEY (authority_operation_id, authority_epoch, authority_sequence) REFERENCES nodecontrol.control_plane_authority_fences(operation_id, authority_epoch, authority_sequence) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_desired_states_signing_key UNIQUE (signing_id),
  CONSTRAINT node_desired_states_node_fk FOREIGN KEY (node_id) REFERENCES nodecontrol.node_inventory(node_id) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_desired_states_signing_fk FOREIGN KEY (signing_id) REFERENCES nodecontrol.node_state_signing_intents(signing_id) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_desired_states_root_publish_fk FOREIGN KEY (root_publish_id) REFERENCES nodecontrol.node_root_metadata_publish_intents(publish_id) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_desired_states_metadata_publish_fk FOREIGN KEY (metadata_publish_id) REFERENCES nodecontrol.node_root_metadata_publish_intents(publish_id) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_desired_states_generation_range CHECK (generation BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_desired_states_inventory_version_range CHECK (inventory_version BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_desired_states_resource_envelope_version_range CHECK (resource_envelope_version BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_desired_states_root_version_range CHECK (root_version BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_desired_states_metadata_version_range CHECK (metadata_version BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_desired_states_resource_envelope_digest_length CHECK (octet_length(resource_envelope_digest) = 32),
  CONSTRAINT node_desired_states_signing_key_id_length CHECK (octet_length(signing_key_id) = 32),
  CONSTRAINT node_desired_states_canonical_payload_length CHECK (octet_length(canonical_payload) BETWEEN 1 AND 65536),
  CONSTRAINT node_desired_states_payload_digest_length CHECK (octet_length(payload_digest) = 32),
  CONSTRAINT node_desired_states_signature_length CHECK (octet_length(signature) = 64),
  CONSTRAINT node_desired_states_validity_window CHECK (issued_at < effective_deadline AND effective_deadline <= valid_until AND valid_until <= issued_at + interval '24 hours'),
  CONSTRAINT node_desired_states_reason_enum CHECK (reason IN ('provision','inventory_update','capacity_change','drain_maintenance','administrative_disable','retire','identity_compromise','host_remediation','security_recovery','authority_restore','release_update')),
  CONSTRAINT node_desired_states_created_order CHECK (created_at >= issued_at),
  CONSTRAINT node_desired_states_retention_window CHECK (retention_until >= created_at + interval '180 days')
);

CREATE TABLE nodecontrol.node_recovery_states (
  node_id uuid CONSTRAINT node_recovery_states_node_id_not_null NOT NULL,
  recovery_generation bigint CONSTRAINT node_recovery_states_recovery_generation_not_null NOT NULL,
  signing_id uuid CONSTRAINT node_recovery_states_signing_id_not_null NOT NULL,
  authority_operation_id uuid CONSTRAINT node_recovery_states_authority_operation_id_not_null NOT NULL,
  authority_epoch bigint CONSTRAINT node_recovery_states_authority_epoch_not_null NOT NULL,
  authority_sequence bigint CONSTRAINT node_recovery_states_authority_sequence_not_null NOT NULL,
  identity_epoch bigint CONSTRAINT node_recovery_states_identity_epoch_not_null NOT NULL,
  recovery_id uuid CONSTRAINT node_recovery_states_recovery_id_not_null NOT NULL,
  recovery_reason text COLLATE "C" CONSTRAINT node_recovery_states_recovery_reason_not_null NOT NULL,
  recovery_session_version bigint CONSTRAINT node_recovery_states_recovery_session_version_not_null NOT NULL,
  incident_set_digest bytea CONSTRAINT node_recovery_states_incident_set_digest_not_null NOT NULL,
  incident_count integer CONSTRAINT node_recovery_states_incident_count_not_null NOT NULL,
  local_fault_bindings_digest bytea CONSTRAINT node_recovery_states_local_fault_bindings_digest_not_null NOT NULL,
  local_fault_binding_count integer CONSTRAINT node_recovery_states_local_fault_binding_count_not_null NOT NULL,
  supervisor_fault_bindings_digest bytea CONSTRAINT node_recovery_states_supervisor_fault_bindings_digest_not_null NOT NULL,
  supervisor_fault_binding_count integer CONSTRAINT node_recovery_states_supervisor_fault_binding_count_not_null NOT NULL,
  remediation_digest bytea,
  root_version bigint CONSTRAINT node_recovery_states_root_version_not_null NOT NULL,
  metadata_version bigint CONSTRAINT node_recovery_states_metadata_version_not_null NOT NULL,
  recovery_action text COLLATE "C" CONSTRAINT node_recovery_states_recovery_action_not_null NOT NULL,
  all_slots_stopped boolean CONSTRAINT node_recovery_states_all_slots_stopped_not_null NOT NULL DEFAULT true,
  canonical_payload bytea CONSTRAINT node_recovery_states_canonical_payload_not_null NOT NULL,
  payload_digest bytea CONSTRAINT node_recovery_states_payload_digest_not_null NOT NULL,
  signing_key_id bytea CONSTRAINT node_recovery_states_signing_key_id_not_null NOT NULL,
  signature bytea CONSTRAINT node_recovery_states_signature_not_null NOT NULL,
  issued_at timestamp with time zone CONSTRAINT node_recovery_states_issued_at_not_null NOT NULL,
  valid_until timestamp with time zone CONSTRAINT node_recovery_states_valid_until_not_null NOT NULL,
  created_at timestamp with time zone CONSTRAINT node_recovery_states_created_at_not_null NOT NULL,
  retention_until timestamp with time zone CONSTRAINT node_recovery_states_retention_until_not_null NOT NULL,
  CONSTRAINT node_recovery_states_pkey PRIMARY KEY (node_id, recovery_generation),
  CONSTRAINT node_recovery_states_authority_epoch_range CHECK (authority_epoch BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_recovery_states_authority_sequence_range CHECK (authority_sequence BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_recovery_states_authority_fence_fk FOREIGN KEY (authority_operation_id, authority_epoch, authority_sequence) REFERENCES nodecontrol.control_plane_authority_fences(operation_id, authority_epoch, authority_sequence) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_recovery_states_signing_key UNIQUE (signing_id),
  CONSTRAINT node_recovery_states_node_fk FOREIGN KEY (node_id) REFERENCES nodecontrol.node_inventory(node_id) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_recovery_states_signing_fk FOREIGN KEY (signing_id) REFERENCES nodecontrol.node_state_signing_intents(signing_id) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_recovery_states_recovery_fk FOREIGN KEY (recovery_id) REFERENCES nodecontrol.node_recovery_sessions(recovery_id) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_recovery_states_recovery_generation_range CHECK (recovery_generation BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_recovery_states_identity_epoch_range CHECK (identity_epoch BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_recovery_states_recovery_session_version_range CHECK (recovery_session_version BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_recovery_states_root_version_range CHECK (root_version BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_recovery_states_metadata_version_range CHECK (metadata_version BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_recovery_states_recovery_reason_enum CHECK (recovery_reason IN ('identity_compromise','administrative_disable','retire','authority_restore','security_incident')),
  CONSTRAINT node_recovery_states_incident_set_digest_length CHECK (octet_length(incident_set_digest) = 32),
  CONSTRAINT node_recovery_states_incident_count_range CHECK (incident_count BETWEEN 0 AND 16),
  CONSTRAINT node_recovery_states_local_fault_bindings_digest_length CHECK (octet_length(local_fault_bindings_digest) = 32),
  CONSTRAINT node_recovery_states_local_fault_count_range CHECK (local_fault_binding_count BETWEEN 0 AND 64),
  CONSTRAINT node_recovery_states_supervisor_fault_bindings_digest_length CHECK (octet_length(supervisor_fault_bindings_digest) = 32),
  CONSTRAINT node_recovery_states_supervisor_fault_count_range CHECK (supervisor_fault_binding_count BETWEEN 0 AND 16),
  CONSTRAINT node_recovery_states_remediation_digest_length CHECK (remediation_digest IS NULL OR octet_length(remediation_digest) = 32),
  CONSTRAINT node_recovery_states_recovery_action_enum CHECK (recovery_action IN ('hold_stopped','submit_recovery_attestation','clear_security_latches')),
  CONSTRAINT node_recovery_states_remediation_action_pair CHECK ((recovery_action = 'clear_security_latches') = (remediation_digest IS NOT NULL)),
  CONSTRAINT node_recovery_states_empty_incident_reason CHECK (incident_count > 0 OR recovery_reason IN ('administrative_disable','retire','authority_restore')),
  CONSTRAINT node_recovery_states_all_slots_stopped_exact CHECK (all_slots_stopped = true),
  CONSTRAINT node_recovery_states_canonical_payload_length CHECK (octet_length(canonical_payload) BETWEEN 1 AND 65536),
  CONSTRAINT node_recovery_states_payload_digest_length CHECK (octet_length(payload_digest) = 32),
  CONSTRAINT node_recovery_states_signing_key_id_length CHECK (octet_length(signing_key_id) = 32),
  CONSTRAINT node_recovery_states_signature_length CHECK (octet_length(signature) = 64),
  CONSTRAINT node_recovery_states_validity_window CHECK (valid_until > issued_at AND valid_until <= issued_at + interval '15 minutes'),
  CONSTRAINT node_recovery_states_created_order CHECK (created_at >= issued_at),
  CONSTRAINT node_recovery_states_retention_window CHECK (retention_until >= created_at + interval '180 days')
);

CREATE TABLE nodecontrol.node_observed_states (
  node_id uuid CONSTRAINT node_observed_states_node_id_not_null NOT NULL,
  boot_id uuid CONSTRAINT node_observed_states_boot_id_not_null NOT NULL,
  sequence bigint CONSTRAINT node_observed_states_sequence_not_null NOT NULL,
  request_digest bytea CONSTRAINT node_observed_states_request_digest_not_null NOT NULL,
  observation_digest bytea CONSTRAINT node_observed_states_observation_digest_not_null NOT NULL,
  canonical_observation bytea CONSTRAINT node_observed_states_canonical_observation_not_null NOT NULL,
  sample_ended_at timestamp with time zone CONSTRAINT node_observed_states_sample_ended_at_not_null NOT NULL,
  arrived_at timestamp with time zone CONSTRAINT node_observed_states_arrived_at_not_null NOT NULL,
  seen_desired_generation bigint CONSTRAINT node_observed_states_seen_desired_generation_not_null NOT NULL DEFAULT 0,
  seen_desired_authority_epoch bigint,
  seen_desired_authority_sequence bigint,
  seen_desired_digest bytea,
  applied_desired_generation bigint CONSTRAINT node_observed_states_applied_desired_generation_not_null NOT NULL DEFAULT 0,
  applied_desired_authority_epoch bigint,
  applied_desired_authority_sequence bigint,
  applied_desired_digest bytea,
  seen_recovery_generation bigint CONSTRAINT node_observed_states_seen_recovery_generation_not_null NOT NULL DEFAULT 0,
  seen_recovery_authority_epoch bigint,
  seen_recovery_authority_sequence bigint,
  seen_recovery_digest bytea,
  applied_recovery_generation bigint CONSTRAINT node_observed_states_applied_recovery_generation_not_null NOT NULL DEFAULT 0,
  applied_recovery_authority_epoch bigint,
  applied_recovery_authority_sequence bigint,
  applied_recovery_digest bytea,
  inventory_version bigint CONSTRAINT node_observed_states_inventory_version_not_null NOT NULL,
  reducer_version bigint CONSTRAINT node_observed_states_reducer_version_not_null NOT NULL,
  reducer_input_digest bytea CONSTRAINT node_observed_states_reducer_input_digest_not_null NOT NULL,
  reducer_state_digest bytea CONSTRAINT node_observed_states_reducer_state_digest_not_null NOT NULL,
  health_state text COLLATE "C" CONSTRAINT node_observed_states_health_state_not_null NOT NULL,
  health_reason text COLLATE "C" CONSTRAINT node_observed_states_health_reason_not_null NOT NULL,
  capacity_accepting boolean CONSTRAINT node_observed_states_capacity_accepting_not_null NOT NULL,
  agent_accepting boolean CONSTRAINT node_observed_states_agent_accepting_not_null NOT NULL,
  final_accepting boolean CONSTRAINT node_observed_states_final_accepting_not_null NOT NULL,
  last_capacity_change_at timestamp with time zone,
  last_agent_change_at timestamp with time zone,
  last_final_change_at timestamp with time zone,
  created_at timestamp with time zone CONSTRAINT node_observed_states_created_at_not_null NOT NULL,
  updated_at timestamp with time zone CONSTRAINT node_observed_states_updated_at_not_null NOT NULL,
  CONSTRAINT node_observed_states_pkey PRIMARY KEY (node_id),
  CONSTRAINT node_observed_states_node_fk FOREIGN KEY (node_id) REFERENCES nodecontrol.node_inventory(node_id) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_observed_states_sequence_range CHECK (sequence BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_observed_states_request_digest_length CHECK (octet_length(request_digest) = 32),
  CONSTRAINT node_observed_states_observation_digest_length CHECK (octet_length(observation_digest) = 32),
  CONSTRAINT node_observed_states_canonical_observation_length CHECK (octet_length(canonical_observation) BETWEEN 1 AND 65536),
  CONSTRAINT node_observed_states_arrival_order CHECK (arrived_at >= sample_ended_at),
  CONSTRAINT node_observed_states_seen_desired_group CHECK ((seen_desired_generation = 0 AND seen_desired_authority_epoch IS NULL AND seen_desired_authority_sequence IS NULL AND seen_desired_digest IS NULL) OR (seen_desired_generation BETWEEN 1 AND 9223372036854775807 AND seen_desired_authority_epoch BETWEEN 1 AND 9223372036854775807 AND seen_desired_authority_sequence BETWEEN 1 AND 9223372036854775807 AND seen_desired_digest IS NOT NULL)),
  CONSTRAINT node_observed_states_seen_desired_digest_length CHECK (seen_desired_digest IS NULL OR octet_length(seen_desired_digest) = 32),
  CONSTRAINT node_observed_states_applied_desired_group CHECK ((applied_desired_generation = 0 AND applied_desired_authority_epoch IS NULL AND applied_desired_authority_sequence IS NULL AND applied_desired_digest IS NULL) OR (applied_desired_generation BETWEEN 1 AND 9223372036854775807 AND applied_desired_authority_epoch BETWEEN 1 AND 9223372036854775807 AND applied_desired_authority_sequence BETWEEN 1 AND 9223372036854775807 AND applied_desired_digest IS NOT NULL)),
  CONSTRAINT node_observed_states_applied_desired_digest_length CHECK (applied_desired_digest IS NULL OR octet_length(applied_desired_digest) = 32),
  CONSTRAINT node_observed_states_seen_recovery_group CHECK ((seen_recovery_generation = 0 AND seen_recovery_authority_epoch IS NULL AND seen_recovery_authority_sequence IS NULL AND seen_recovery_digest IS NULL) OR (seen_recovery_generation BETWEEN 1 AND 9223372036854775807 AND seen_recovery_authority_epoch BETWEEN 1 AND 9223372036854775807 AND seen_recovery_authority_sequence BETWEEN 1 AND 9223372036854775807 AND seen_recovery_digest IS NOT NULL)),
  CONSTRAINT node_observed_states_seen_recovery_digest_length CHECK (seen_recovery_digest IS NULL OR octet_length(seen_recovery_digest) = 32),
  CONSTRAINT node_observed_states_applied_recovery_group CHECK ((applied_recovery_generation = 0 AND applied_recovery_authority_epoch IS NULL AND applied_recovery_authority_sequence IS NULL AND applied_recovery_digest IS NULL) OR (applied_recovery_generation BETWEEN 1 AND 9223372036854775807 AND applied_recovery_authority_epoch BETWEEN 1 AND 9223372036854775807 AND applied_recovery_authority_sequence BETWEEN 1 AND 9223372036854775807 AND applied_recovery_digest IS NOT NULL)),
  CONSTRAINT node_observed_states_applied_recovery_digest_length CHECK (applied_recovery_digest IS NULL OR octet_length(applied_recovery_digest) = 32),
  CONSTRAINT node_observed_states_inventory_version_range CHECK (inventory_version BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_observed_states_reducer_version_range CHECK (reducer_version BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_observed_states_reducer_input_digest_length CHECK (octet_length(reducer_input_digest) = 32),
  CONSTRAINT node_observed_states_reducer_state_digest_length CHECK (octet_length(reducer_state_digest) = 32),
  CONSTRAINT node_observed_states_health_state_enum CHECK (health_state IN ('unknown','healthy','degraded','offline','quarantined')),
  CONSTRAINT node_observed_states_health_reason_enum CHECK (health_reason IN ('none','warming','capacity_pressure','required_slot_failed','heartbeat_timeout','security_quarantine','operator_disabled')),
  CONSTRAINT node_observed_states_final_accepting_requires_agent CHECK (NOT final_accepting OR agent_accepting),
  CONSTRAINT node_observed_states_timestamp_order CHECK (updated_at >= created_at)
);

CREATE TABLE nodecontrol.node_operator_audit (
  audit_id uuid CONSTRAINT node_operator_audit_audit_id_not_null NOT NULL,
  command_id uuid CONSTRAINT node_operator_audit_command_id_not_null NOT NULL,
  authority_operation_id uuid,
  authority_epoch bigint,
  authority_sequence bigint,
  operator_id text COLLATE "C" CONSTRAINT node_operator_audit_operator_id_not_null NOT NULL,
  credential_digest bytea CONSTRAINT node_operator_audit_credential_digest_not_null NOT NULL,
  role text COLLATE "C" CONSTRAINT node_operator_audit_role_not_null NOT NULL,
  action text COLLATE "C" CONSTRAINT node_operator_audit_action_not_null NOT NULL,
  target_kind text COLLATE "C" CONSTRAINT node_operator_audit_target_kind_not_null NOT NULL,
  target_id text COLLATE "C" CONSTRAINT node_operator_audit_target_id_not_null NOT NULL,
  reason text COLLATE "C" CONSTRAINT node_operator_audit_reason_not_null NOT NULL,
  result text COLLATE "C" CONSTRAINT node_operator_audit_result_not_null NOT NULL,
  before_version bigint,
  after_version bigint,
  occurred_at timestamp with time zone CONSTRAINT node_operator_audit_occurred_at_not_null NOT NULL,
  retention_until timestamp with time zone CONSTRAINT node_operator_audit_retention_until_not_null NOT NULL,
  CONSTRAINT node_operator_audit_pkey PRIMARY KEY (audit_id),
  CONSTRAINT node_operator_audit_command_id_key UNIQUE (command_id),
  CONSTRAINT node_operator_audit_authority_operation_key UNIQUE (authority_operation_id),
  CONSTRAINT node_operator_audit_authority_fence_fk FOREIGN KEY (authority_operation_id, authority_epoch, authority_sequence) REFERENCES nodecontrol.control_plane_authority_fences(operation_id, authority_epoch, authority_sequence) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_operator_audit_authority_all_or_none CHECK (num_nonnulls(authority_operation_id,authority_epoch,authority_sequence) IN (0,3)),
  CONSTRAINT node_operator_audit_authority_epoch_range CHECK (authority_epoch IS NULL OR authority_epoch BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_operator_audit_authority_sequence_range CHECK (authority_sequence IS NULL OR authority_sequence BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_operator_audit_operator_id_format CHECK (octet_length(operator_id) BETWEEN 1 AND 128 AND operator_id ~ '^[A-Za-z0-9:_-]{1,128}$'),
  CONSTRAINT node_operator_audit_credential_digest_length CHECK (octet_length(credential_digest) = 32),
  CONSTRAINT node_operator_audit_role_enum CHECK (role IN ('inventory_reader','inventory_writer','node_security_admin')),
  CONSTRAINT node_operator_audit_action_enum CHECK (action IN ('create_pop','update_inventory','register_envelope','drain_node','disable_node','restore_reauthorize','clear_security','retire_node')),
  CONSTRAINT node_operator_audit_target_kind_enum CHECK (target_kind IN ('pop','node','failure_domain','capacity_profile','recovery_session','security_incident')),
  CONSTRAINT node_operator_audit_target_id_format CHECK (octet_length(target_id) BETWEEN 1 AND 128 AND target_id ~ '^[A-Za-z0-9:_-]{1,128}$'),
  CONSTRAINT node_operator_audit_reason_enum CHECK (reason IN ('provision','inventory_update','capacity_change','drain_maintenance','administrative_disable','retire','identity_compromise','host_remediation','security_recovery','authority_restore','release_update')),
  CONSTRAINT node_operator_audit_result_enum CHECK (result IN ('accepted','rejected','conflict')),
  CONSTRAINT node_operator_audit_version_pair CHECK ((before_version IS NULL) = (after_version IS NULL)),
  CONSTRAINT node_operator_audit_version_range CHECK (before_version IS NULL OR (before_version BETWEEN 0 AND 9223372036854775807 AND after_version BETWEEN 1 AND 9223372036854775807)),
  CONSTRAINT node_operator_audit_retention_window CHECK (retention_until >= occurred_at + interval '180 days')
);

CREATE TABLE nodecontrol.node_state_transitions (
  transition_id uuid CONSTRAINT node_state_transitions_transition_id_not_null NOT NULL,
  node_id uuid CONSTRAINT node_state_transitions_node_id_not_null NOT NULL,
  authority_operation_id uuid,
  authority_epoch bigint,
  authority_sequence bigint,
  dimension text COLLATE "C" CONSTRAINT node_state_transitions_dimension_not_null NOT NULL,
  from_state text COLLATE "C" CONSTRAINT node_state_transitions_from_state_not_null NOT NULL,
  to_state text COLLATE "C" CONSTRAINT node_state_transitions_to_state_not_null NOT NULL,
  reason text COLLATE "C" CONSTRAINT node_state_transitions_reason_not_null NOT NULL,
  observation_boot_id uuid,
  observation_sequence bigint,
  incident_id uuid,
  audit_id uuid,
  aggregate_version bigint CONSTRAINT node_state_transitions_aggregate_version_not_null NOT NULL,
  occurred_at timestamp with time zone CONSTRAINT node_state_transitions_occurred_at_not_null NOT NULL,
  retention_until timestamp with time zone CONSTRAINT node_state_transitions_retention_until_not_null NOT NULL,
  CONSTRAINT node_state_transitions_pkey PRIMARY KEY (transition_id),
  CONSTRAINT node_state_transitions_node_fk FOREIGN KEY (node_id) REFERENCES nodecontrol.node_inventory(node_id) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_state_transitions_authority_fence_fk FOREIGN KEY (authority_operation_id, authority_epoch, authority_sequence) REFERENCES nodecontrol.control_plane_authority_fences(operation_id, authority_epoch, authority_sequence) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_state_transitions_incident_fk FOREIGN KEY (incident_id) REFERENCES nodecontrol.node_security_incidents(incident_id) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_state_transitions_audit_fk FOREIGN KEY (audit_id) REFERENCES nodecontrol.node_operator_audit(audit_id) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT node_state_transitions_authority_all_or_none CHECK (num_nonnulls(authority_operation_id,authority_epoch,authority_sequence) IN (0,3)),
  CONSTRAINT node_state_transitions_authority_epoch_range CHECK (authority_epoch IS NULL OR authority_epoch BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_state_transitions_authority_sequence_range CHECK (authority_sequence IS NULL OR authority_sequence BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_state_transitions_dimension_enum CHECK (dimension IN ('operator','security','identity','health','certificate')),
  CONSTRAINT node_state_transitions_from_state_format CHECK (octet_length(from_state) BETWEEN 1 AND 32 AND from_state ~ '^[a-z][a-z0-9_]{0,31}$'),
  CONSTRAINT node_state_transitions_to_state_format CHECK (octet_length(to_state) BETWEEN 1 AND 32 AND to_state ~ '^[a-z][a-z0-9_]{0,31}$'),
  CONSTRAINT node_state_transitions_dimension_state_values CHECK ((dimension = 'operator' AND from_state IN ('provisioning','enabled','draining','disabled') AND to_state IN ('provisioning','enabled','draining','disabled')) OR (dimension = 'security' AND from_state IN ('normal','quarantined') AND to_state IN ('normal','quarantined')) OR (dimension = 'identity' AND from_state IN ('never_enrolled','active','recovery_pending','recovery_limited','revoked') AND to_state IN ('never_enrolled','active','recovery_pending','recovery_limited','revoked')) OR (dimension = 'health' AND from_state IN ('unknown','healthy','degraded','offline','quarantined') AND to_state IN ('unknown','healthy','degraded','offline','quarantined')) OR (dimension = 'certificate' AND from_state IN ('active','recovery_pending','recovery_limited','revoked') AND to_state IN ('active','recovery_pending','recovery_limited','revoked'))),
  CONSTRAINT node_state_transitions_state_changed CHECK (from_state <> to_state),
  CONSTRAINT node_state_transitions_reason_enum CHECK (reason IN ('provision','inventory_update','capacity_change','drain_maintenance','administrative_disable','retire','identity_compromise','host_remediation','security_recovery','authority_restore','release_update','observation_update','certificate_lifecycle','incident_capacity_exceeded')),
  CONSTRAINT node_state_transitions_observation_pair CHECK ((observation_boot_id IS NULL) = (observation_sequence IS NULL)),
  CONSTRAINT node_state_transitions_observation_sequence_range CHECK (observation_sequence IS NULL OR observation_sequence BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_state_transitions_aggregate_version_range CHECK (aggregate_version BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT node_state_transitions_retention_window CHECK (retention_until >= occurred_at + interval '180 days')
);

CREATE TABLE nodecontrol.control_plane_trust_bundle_high_waters (
  purpose text COLLATE "C" CONSTRAINT control_plane_trust_bundle_high_waters_purpose_not_null NOT NULL,
  listener_kind text COLLATE "C" CONSTRAINT control_plane_trust_bundle_high_waters_listener_kind_not_null NOT NULL,
  trust_domain text COLLATE "C" CONSTRAINT control_plane_trust_bundle_high_waters_trust_domain_not_null NOT NULL,
  authority_operation_id uuid CONSTRAINT control_plane_trust_bundle_high_waters_authority_opera_db522a07 NOT NULL,
  authority_epoch bigint CONSTRAINT control_plane_trust_bundle_high_waters_authority_epoch_not_null NOT NULL,
  authority_sequence bigint CONSTRAINT control_plane_trust_bundle_high_waters_authority_seque_d3d6601d NOT NULL,
  bundle_version bigint CONSTRAINT control_plane_trust_bundle_high_waters_bundle_version_not_null NOT NULL,
  bundle_digest bytea CONSTRAINT control_plane_trust_bundle_high_waters_bundle_digest_not_null NOT NULL,
  cumulative_set_digest bytea CONSTRAINT control_plane_trust_bundle_high_waters_cumulative_set__b8532caf NOT NULL,
  cumulative_set_count integer CONSTRAINT control_plane_trust_bundle_high_waters_cumulative_set__ebf19554 NOT NULL,
  updated_at timestamp with time zone CONSTRAINT control_plane_trust_bundle_high_waters_updated_at_not_null NOT NULL,
  CONSTRAINT control_plane_trust_bundle_high_waters_pkey PRIMARY KEY (purpose, listener_kind, trust_domain),
  CONSTRAINT control_plane_trust_bundle_high_waters_authority_epoch_range CHECK (authority_epoch BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT control_plane_trust_bundle_high_waters_authority_sequence_range CHECK (authority_sequence BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT control_plane_trust_bundle_high_waters_authority_fence_fk FOREIGN KEY (authority_operation_id, authority_epoch, authority_sequence) REFERENCES nodecontrol.control_plane_authority_fences(operation_id, authority_epoch, authority_sequence) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT control_plane_trust_bundle_high_waters_purpose_enum CHECK (purpose IN ('bootstrap_server','agent_server','operator_server','node_client','operator_client')),
  CONSTRAINT control_plane_trust_bundle_high_waters_listener_kind_enum CHECK (listener_kind IN ('bootstrap','agent','operator','outbound_node','outbound_operator')),
  CONSTRAINT control_plane_trust_bundle_high_waters_purpose_listener_mapping CHECK ((purpose = 'bootstrap_server' AND listener_kind = 'bootstrap') OR (purpose = 'agent_server' AND listener_kind = 'agent') OR (purpose = 'operator_server' AND listener_kind = 'operator') OR (purpose = 'node_client' AND listener_kind = 'outbound_node') OR (purpose = 'operator_client' AND listener_kind = 'outbound_operator')),
  CONSTRAINT control_plane_trust_bundle_high_waters_trust_domain_format CHECK (octet_length(trust_domain) BETWEEN 1 AND 253 AND trust_domain = lower(trust_domain) AND trust_domain ~ '^(?=.{1,253}$)(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?)(?:\.(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?))*$'),
  CONSTRAINT control_plane_trust_bundle_high_waters_bundle_version_range CHECK (bundle_version BETWEEN 1 AND 9223372036854775807),
  CONSTRAINT control_plane_trust_bundle_high_waters_bundle_digest_length CHECK (octet_length(bundle_digest) = 32),
  CONSTRAINT control_plane_trust_bundle_high_waters_cumulative_set__82918e98 CHECK (octet_length(cumulative_set_digest) = 32),
  CONSTRAINT control_plane_trust_bundle_high_waters_cumulative_set__4ff0c7bc CHECK (cumulative_set_count BETWEEN 0 AND 128)
);

-- +goose StatementBegin
CREATE FUNCTION nodecontrol.reject_row_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
  outbox_reference boolean := false;
BEGIN
  IF TG_OP = 'DELETE' AND TG_TABLE_NAME = 'node_desired_states' THEN
    PERFORM 1
    FROM nodecontrol.node_inventory inventory
    WHERE inventory.node_id = OLD.node_id
    FOR UPDATE;
    IF NOT FOUND THEN
      RAISE EXCEPTION 'desired state node does not exist' USING ERRCODE = '23503';
    END IF;
    IF OLD.retention_until > transaction_timestamp()
       OR EXISTS (
         SELECT 1 FROM nodecontrol.node_inventory inventory
         WHERE inventory.node_id = OLD.node_id
           AND inventory.active_desired_generation = OLD.generation
       ) THEN
      RAISE EXCEPTION 'desired state is retained or actively referenced' USING ERRCODE = '23514';
    END IF;
    IF to_regclass('public.transactional_outbox') IS NOT NULL THEN
      EXECUTE 'LOCK TABLE public.transactional_outbox IN SHARE ROW EXCLUSIVE MODE';
      EXECUTE 'SELECT EXISTS (SELECT 1 FROM public.transactional_outbox WHERE aggregate_type = $1 AND aggregate_id = $2 AND aggregate_version = $3)'
      INTO outbox_reference
      USING 'node', OLD.node_id, OLD.generation;
    END IF;
    IF outbox_reference THEN
      RAISE EXCEPTION 'desired state is referenced by the transactional outbox' USING ERRCODE = '23514';
    END IF;
    RETURN OLD;
  END IF;
  IF TG_OP = 'DELETE' AND TG_TABLE_NAME = 'node_operator_audit' THEN
    IF OLD.retention_until > transaction_timestamp()
       OR EXISTS (
         SELECT 1 FROM nodecontrol.node_state_transitions transition
         WHERE transition.audit_id = OLD.audit_id
       ) THEN
      RAISE EXCEPTION 'operator audit is retained or referenced' USING ERRCODE = '23514';
    END IF;
    IF to_regclass('public.transactional_outbox') IS NOT NULL THEN
      EXECUTE 'LOCK TABLE public.transactional_outbox IN SHARE ROW EXCLUSIVE MODE';
      EXECUTE 'SELECT EXISTS (SELECT 1 FROM public.transactional_outbox WHERE aggregate_type = $1 AND aggregate_id = $2)'
      INTO outbox_reference
      USING 'operator_action', OLD.audit_id;
    END IF;
    IF outbox_reference THEN
      RAISE EXCEPTION 'operator audit is referenced by the transactional outbox' USING ERRCODE = '23514';
    END IF;
    RETURN OLD;
  END IF;
  RAISE EXCEPTION '% rows are immutable', TG_TABLE_NAME USING ERRCODE = '23514';
END
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION nodecontrol.enforce_authority_fence_update()
RETURNS trigger
LANGUAGE plpgsql
AS $$
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
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION nodecontrol.enforce_capacity_profile_immutability()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM nodecontrol.node_process_slots
    WHERE capacity_profile_id = OLD.profile_id
      AND capacity_profile_version = OLD.version
  ) THEN
    RAISE EXCEPTION 'referenced capacity profile is immutable' USING ERRCODE = '23514';
  END IF;
  IF TG_OP = 'DELETE' THEN
    RETURN OLD;
  END IF;
  RETURN NEW;
END
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION nodecontrol.enforce_process_slot_cap()
RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
  existing_count integer;
  profile_adapter text;
BEGIN
  PERFORM 1 FROM nodecontrol.node_inventory WHERE node_id = NEW.node_id FOR UPDATE;
  IF NOT FOUND THEN
    RAISE EXCEPTION 'unknown node for process slot' USING ERRCODE = '23503';
  END IF;

  SELECT adapter INTO profile_adapter
  FROM nodecontrol.node_capacity_profiles
  WHERE profile_id = NEW.capacity_profile_id
    AND version = NEW.capacity_profile_version
  FOR SHARE;
  IF profile_adapter IS NULL OR profile_adapter <> NEW.adapter THEN
    RAISE EXCEPTION 'slot adapter does not match capacity profile' USING ERRCODE = '23514';
  END IF;

  SELECT count(*) INTO existing_count
  FROM nodecontrol.node_process_slots
  WHERE node_id = NEW.node_id
    AND (TG_OP = 'INSERT' OR (node_id, slot_id) IS DISTINCT FROM (OLD.node_id, OLD.slot_id));
  IF existing_count >= 8 THEN
    RAISE EXCEPTION 'a node may have at most eight process slots' USING ERRCODE = '23514';
  END IF;
  RETURN NEW;
END
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION nodecontrol.enforce_inventory_pointers()
RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
  pointer_recovery_session nodecontrol.node_recovery_sessions%ROWTYPE;
  pointer_incident_digest bytea;
  pointer_incident_count integer;
  pointer_local_digest bytea;
  pointer_local_count integer;
  pointer_supervisor_digest bytea;
  pointer_supervisor_count integer;
  pointer_remediation_digest bytea;
  pointer_required_action text;
  pointer_pending_resolution_count integer;
  pointer_distinct_remediation_count integer;
BEGIN
  IF NEW.resource_envelope_version IS NOT NULL AND NOT EXISTS (
    SELECT 1
    FROM nodecontrol.node_resource_envelopes envelope
    JOIN nodecontrol.control_plane_authority_fences fence
      ON (fence.operation_id, fence.authority_epoch, fence.authority_sequence) =
         (envelope.authority_operation_id, envelope.authority_epoch, envelope.authority_sequence)
    WHERE envelope.node_id = NEW.node_id
      AND envelope.envelope_version = NEW.resource_envelope_version
       AND envelope.envelope_digest = NEW.resource_envelope_digest
       AND fence.provider_status = 'committed'
       AND fence.visibility_state = 'active'
       AND fence.effect_kind = 'resource_envelope_activate'
       AND fence.scope_kind = 'node'
  ) THEN
    RAISE EXCEPTION 'resource envelope pointer is not fence-finalized' USING ERRCODE = '23514';
  END IF;

  IF NEW.active_desired_generation IS NOT NULL AND NOT EXISTS (
    SELECT 1
    FROM nodecontrol.node_desired_states state
    JOIN nodecontrol.node_state_signing_intents intent ON intent.signing_id = state.signing_id
    JOIN nodecontrol.control_plane_authority_fences fence
      ON (fence.operation_id, fence.authority_epoch, fence.authority_sequence) =
         (state.authority_operation_id, state.authority_epoch, state.authority_sequence)
    WHERE state.node_id = NEW.node_id
      AND state.generation = NEW.active_desired_generation
       AND intent.signing_kind = 'desired'
       AND intent.status = 'active'
       AND (intent.authority_operation_id,intent.authority_epoch,intent.authority_sequence) =
           (state.authority_operation_id,state.authority_epoch,state.authority_sequence)
       AND intent.node_id = state.node_id
       AND intent.reserved_generation = state.generation
       AND intent.canonical_payload = state.canonical_payload
       AND intent.payload_digest = state.payload_digest
       AND intent.signature = state.signature
       AND intent.root_version = state.root_version
       AND intent.metadata_version = state.metadata_version
       AND intent.expected_key_id = state.signing_key_id
       AND intent.captured_inventory_version = state.inventory_version
       AND intent.captured_inventory_version = NEW.inventory_version
       AND intent.captured_identity_epoch = NEW.identity_epoch
       AND intent.captured_security_version = NEW.security_version
       AND state.resource_envelope_version = NEW.resource_envelope_version
       AND state.resource_envelope_digest = NEW.resource_envelope_digest
       AND state.root_publish_id = NEW.active_root_publish_id
       AND state.root_version = NEW.active_root_version
       AND state.metadata_publish_id = NEW.active_metadata_publish_id
       AND state.metadata_version = NEW.active_metadata_version
       AND fence.provider_status = 'committed'
       AND fence.visibility_state = 'active'
       AND fence.effect_kind = 'desired_activate'
       AND fence.scope_kind = 'node'
  ) THEN
    RAISE EXCEPTION 'desired pointer is not fence-finalized' USING ERRCODE = '23514';
  END IF;

  IF NEW.active_recovery_generation IS NOT NULL THEN
    SELECT recovery.* INTO pointer_recovery_session
    FROM nodecontrol.node_recovery_states state
    JOIN nodecontrol.node_recovery_sessions recovery ON recovery.recovery_id = state.recovery_id
    WHERE state.node_id = NEW.node_id
      AND state.recovery_generation = NEW.active_recovery_generation
    FOR SHARE OF state,recovery;
    IF NOT FOUND THEN
      RAISE EXCEPTION 'recovery pointer is not fence-finalized' USING ERRCODE = '23514';
    END IF;

    SELECT
      pg_catalog.sha256(COALESCE(
        pg_catalog.string_agg(pg_catalog.uuid_send(incident.incident_id),''::bytea ORDER BY incident.incident_id),
        ''::bytea
      )),
      count(*)::integer,
      count(*) FILTER (WHERE incident.status = 'resolution_pending_agent_ack')::integer,
      count(DISTINCT pg_catalog.encode(incident.remediation_digest,'hex'))
        FILTER (WHERE incident.status = 'resolution_pending_agent_ack')::integer,
      CASE
        WHEN count(*) FILTER (WHERE incident.status = 'resolution_pending_agent_ack') > 0
             AND count(DISTINCT pg_catalog.encode(incident.remediation_digest,'hex'))
                 FILTER (WHERE incident.status = 'resolution_pending_agent_ack') = 1
        THEN pg_catalog.decode(
          min(pg_catalog.encode(incident.remediation_digest,'hex'))
            FILTER (WHERE incident.status = 'resolution_pending_agent_ack'),
          'hex'
        )
        ELSE NULL
      END
    INTO pointer_incident_digest,pointer_incident_count,pointer_pending_resolution_count,
         pointer_distinct_remediation_count,pointer_remediation_digest
    FROM nodecontrol.node_security_incidents incident
    WHERE incident.node_id = NEW.node_id
      AND incident.identity_epoch = NEW.identity_epoch
      AND incident.status IN ('open','resolution_pending_agent_ack','overflow');

    SELECT
      pg_catalog.sha256(COALESCE(
        pg_catalog.string_agg(
          pg_catalog.uuid_send(receipt.local_fault_id) || pg_catalog.uuid_send(receipt.incident_id),
          ''::bytea ORDER BY receipt.local_fault_id
        ),
        ''::bytea
      )),
      count(*)::integer
    INTO pointer_local_digest,pointer_local_count
    FROM nodecontrol.node_security_fault_receipts receipt
    JOIN nodecontrol.node_security_incidents incident ON incident.incident_id = receipt.incident_id
    WHERE receipt.node_id = NEW.node_id
      AND receipt.identity_epoch = NEW.identity_epoch
      AND receipt.binding_status = 'active'
      AND incident.status IN ('open','resolution_pending_agent_ack','overflow');

    SELECT
      pg_catalog.sha256(COALESCE(
        pg_catalog.string_agg(
          pg_catalog.uuid_send(receipt.supervisor_boot_id) || pg_catalog.uuid_send(receipt.supervisor_fault_id) || receipt.supervisor_evidence_digest ||
          pg_catalog.uuid_send(receipt.incident_id),
          ''::bytea ORDER BY receipt.supervisor_boot_id,receipt.supervisor_fault_id
        ),
        ''::bytea
      )),
      count(*)::integer
    INTO pointer_supervisor_digest,pointer_supervisor_count
    FROM nodecontrol.node_security_fault_receipts receipt
    JOIN nodecontrol.node_security_incidents incident ON incident.incident_id = receipt.incident_id
    WHERE receipt.node_id = NEW.node_id
      AND receipt.identity_epoch = NEW.identity_epoch
      AND receipt.binding_status = 'active'
      AND receipt.supervisor_fault_id IS NOT NULL
      AND incident.status IN ('open','resolution_pending_agent_ack','overflow');

    pointer_required_action := CASE
      WHEN pointer_pending_resolution_count > 0 THEN 'clear_security_latches'
      WHEN pointer_recovery_session.recovery_certificate_id IS NOT NULL
           AND pointer_recovery_session.attestation_digest IS NULL THEN 'submit_recovery_attestation'
      ELSE 'hold_stopped'
    END;
  END IF;

  IF NEW.active_recovery_generation IS NOT NULL AND NOT EXISTS (
    SELECT 1
    FROM nodecontrol.node_recovery_states state
    JOIN nodecontrol.node_state_signing_intents intent ON intent.signing_id = state.signing_id
    JOIN nodecontrol.node_recovery_sessions recovery ON recovery.recovery_id = state.recovery_id
    JOIN nodecontrol.control_plane_authority_fences fence
      ON (fence.operation_id, fence.authority_epoch, fence.authority_sequence) =
         (state.authority_operation_id, state.authority_epoch, state.authority_sequence)
    WHERE state.node_id = NEW.node_id
      AND state.recovery_generation = NEW.active_recovery_generation
       AND intent.signing_kind = 'recovery'
       AND intent.status = 'active'
       AND (intent.authority_operation_id,intent.authority_epoch,intent.authority_sequence) =
           (state.authority_operation_id,state.authority_epoch,state.authority_sequence)
       AND intent.node_id = state.node_id
       AND intent.reserved_generation = state.recovery_generation
       AND intent.canonical_payload = state.canonical_payload
       AND intent.payload_digest = state.payload_digest
       AND intent.signature = state.signature
       AND intent.root_version = state.root_version
       AND intent.metadata_version = state.metadata_version
       AND intent.expected_key_id = state.signing_key_id
       AND intent.captured_inventory_version = NEW.inventory_version
       AND intent.captured_identity_epoch = state.identity_epoch
       AND intent.captured_identity_epoch = NEW.identity_epoch
       AND intent.captured_security_version = NEW.security_version
       AND state.root_version = NEW.active_root_version
       AND state.metadata_version = NEW.active_metadata_version
       AND intent.recovery_id = state.recovery_id
       AND intent.recovery_reason = state.recovery_reason
       AND intent.recovery_session_version = state.recovery_session_version
       AND intent.recovery_session_status = 'pending'
       AND recovery.node_id = state.node_id
       AND recovery.identity_epoch = state.identity_epoch
       AND recovery.reason = state.recovery_reason
       AND recovery.version = state.recovery_session_version
       AND recovery.status = 'pending'
       AND recovery.incident_set_digest = pointer_incident_digest
       AND state.incident_set_digest = pointer_incident_digest
       AND state.incident_count = pointer_incident_count
       AND intent.recovery_incident_set_digest = pointer_incident_digest
       AND state.local_fault_bindings_digest = pointer_local_digest
       AND state.local_fault_binding_count = pointer_local_count
       AND intent.recovery_local_bindings_digest = pointer_local_digest
       AND state.supervisor_fault_bindings_digest = pointer_supervisor_digest
       AND state.supervisor_fault_binding_count = pointer_supervisor_count
       AND intent.recovery_supervisor_bindings_digest = pointer_supervisor_digest
       AND pointer_distinct_remediation_count <= 1
       AND state.remediation_digest IS NOT DISTINCT FROM pointer_remediation_digest
       AND intent.recovery_remediation_digest IS NOT DISTINCT FROM pointer_remediation_digest
       AND state.recovery_action = pointer_required_action
       AND intent.recovery_required_action = pointer_required_action
       AND NEW.operator_state = 'disabled'
       AND NEW.security_state = 'quarantined'
       AND NEW.identity_state IN ('recovery_pending','recovery_limited')
       AND fence.provider_status = 'committed'
       AND fence.visibility_state = 'active'
       AND fence.effect_kind = 'recovery_activate'
       AND fence.scope_kind = 'node'
  ) THEN
    RAISE EXCEPTION 'recovery pointer is not fence-finalized' USING ERRCODE = '23514';
  END IF;

  IF NEW.active_root_publish_id IS NOT NULL AND NOT EXISTS (
    SELECT 1
    FROM nodecontrol.node_root_metadata_publish_intents intent
    JOIN nodecontrol.control_plane_authority_fences fence
      ON (fence.operation_id, fence.authority_epoch, fence.authority_sequence) =
         (intent.authority_operation_id, intent.authority_epoch, intent.authority_sequence)
    WHERE intent.publish_id = NEW.active_root_publish_id
      AND intent.publish_kind = 'root'
      AND intent.reserved_version = NEW.active_root_version
       AND intent.status = 'active'
       AND fence.provider_status = 'committed'
       AND fence.visibility_state = 'active'
       AND fence.effect_kind = 'root_publish'
       AND fence.scope_kind = 'global_node_trust'
  ) THEN
    RAISE EXCEPTION 'root pointer is not fence-finalized' USING ERRCODE = '23514';
  END IF;

  IF NEW.active_metadata_publish_id IS NOT NULL AND NOT EXISTS (
    SELECT 1
    FROM nodecontrol.node_root_metadata_publish_intents intent
    JOIN nodecontrol.control_plane_authority_fences fence
      ON (fence.operation_id, fence.authority_epoch, fence.authority_sequence) =
         (intent.authority_operation_id, intent.authority_epoch, intent.authority_sequence)
    WHERE intent.publish_id = NEW.active_metadata_publish_id
      AND intent.publish_kind = 'metadata'
      AND intent.reserved_version = NEW.active_metadata_version
       AND intent.status = 'active'
       AND fence.provider_status = 'committed'
       AND fence.visibility_state = 'active'
       AND fence.effect_kind = 'metadata_publish'
       AND fence.scope_kind = 'global_node_trust'
  ) THEN
    RAISE EXCEPTION 'metadata pointer is not fence-finalized' USING ERRCODE = '23514';
  END IF;

  IF NEW.last_authority_operation_id IS NOT NULL AND NOT EXISTS (
    SELECT 1
    FROM nodecontrol.control_plane_authority_fences fence
    WHERE (fence.operation_id, fence.authority_epoch, fence.authority_sequence) =
          (NEW.last_authority_operation_id, NEW.last_authority_epoch, NEW.last_authority_sequence)
       AND fence.provider_status = 'committed'
       AND fence.visibility_state = 'active'
       AND fence.scope_kind = 'node'
  ) THEN
    RAISE EXCEPTION 'inventory authority pointer is not fence-finalized' USING ERRCODE = '23514';
  END IF;
  RETURN NULL;
END
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION nodecontrol.enforce_certificate_issuance_workflow()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
    IF NEW.status <> 'pending' THEN
      RAISE EXCEPTION 'certificate issuances must start pending' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
  END IF;
  IF TG_OP = 'DELETE' THEN
    IF OLD.status = 'pending' OR OLD.retention_until IS NULL OR OLD.retention_until > transaction_timestamp()
       OR EXISTS (SELECT 1 FROM nodecontrol.node_certificates WHERE issuance_id = OLD.issuance_id)
       OR EXISTS (SELECT 1 FROM nodecontrol.node_enrollment_grants WHERE result_issuance_id = OLD.issuance_id) THEN
      RAISE EXCEPTION 'certificate issuance is retained or referenced' USING ERRCODE = '23514';
    END IF;
    RETURN OLD;
  END IF;
  IF OLD.status <> 'pending' THEN
    RAISE EXCEPTION 'terminal certificate issuance is immutable' USING ERRCODE = '23514';
  END IF;
  IF (NEW.issuance_id, NEW.authority_operation_id, NEW.authority_epoch, NEW.authority_sequence,
      NEW.node_id, NEW.attempt_id, NEW.issuance_kind, NEW.identity_epoch, NEW.lineage_id,
      NEW.issuer_id, NEW.csr_sha256, NEW.public_key_sha256, NEW.template_sha256, NEW.request_digest,
      NEW.created_at)
     IS DISTINCT FROM
     (OLD.issuance_id, OLD.authority_operation_id, OLD.authority_epoch, OLD.authority_sequence,
      OLD.node_id, OLD.attempt_id, OLD.issuance_kind, OLD.identity_epoch, OLD.lineage_id,
      OLD.issuer_id, OLD.csr_sha256, OLD.public_key_sha256, OLD.template_sha256, OLD.request_digest,
      OLD.created_at) THEN
    RAISE EXCEPTION 'certificate issuance inputs are immutable' USING ERRCODE = '23514';
  END IF;
  IF OLD.serial_bytes IS NOT NULL
     AND (NEW.serial_bytes,NEW.leaf_der,NEW.leaf_der_sha256,NEW.chain_der,NEW.chain_der_sha256,NEW.not_before,NEW.not_after)
         IS DISTINCT FROM
         (OLD.serial_bytes,OLD.leaf_der,OLD.leaf_der_sha256,OLD.chain_der,OLD.chain_der_sha256,OLD.not_before,OLD.not_after) THEN
    RAISE EXCEPTION 'certificate issuance result is immutable' USING ERRCODE = '23514';
  END IF;
  RETURN NEW;
END
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION nodecontrol.enforce_enrollment_grant_workflow()
RETURNS trigger
LANGUAGE plpgsql
AS $$
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
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION nodecontrol.enforce_certificate_workflow()
RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
  issuance nodecontrol.node_certificate_issuances%ROWTYPE;
BEGIN
  IF TG_OP = 'INSERT' THEN
    SELECT * INTO issuance
    FROM nodecontrol.node_certificate_issuances
    WHERE issuance_id = NEW.issuance_id
    FOR SHARE;
    IF NOT FOUND
       OR issuance.status <> 'active'
       OR (NEW.authority_operation_id,NEW.authority_epoch,NEW.authority_sequence,
           NEW.node_id,NEW.identity_epoch,NEW.lineage_id,NEW.issuer_id,NEW.serial_bytes,NEW.leaf_der,
           NEW.leaf_der_sha256,NEW.public_key_sha256,NEW.chain_der_sha256,NEW.valid_from,NEW.valid_until)
          IS DISTINCT FROM
          (issuance.authority_operation_id,issuance.authority_epoch,issuance.authority_sequence,
           issuance.node_id,issuance.identity_epoch,issuance.lineage_id,issuance.issuer_id,issuance.serial_bytes,
           issuance.leaf_der,issuance.leaf_der_sha256,issuance.public_key_sha256,issuance.chain_der_sha256,
           issuance.not_before,issuance.not_after)
       OR (issuance.issuance_kind = 'recovery' AND NEW.status <> 'recovery_pending')
       OR (issuance.issuance_kind <> 'recovery' AND NEW.status <> 'active')
       OR NOT EXISTS (
         SELECT 1
         FROM nodecontrol.control_plane_authority_fences fence
         WHERE (fence.operation_id,fence.authority_epoch,fence.authority_sequence) =
               (NEW.authority_operation_id,NEW.authority_epoch,NEW.authority_sequence)
           AND fence.provider_status = 'committed'
           AND fence.visibility_state = 'active'
           AND fence.effect_kind = 'certificate_activate'
           AND fence.scope_kind = 'node'
       ) THEN
      RAISE EXCEPTION 'certificate must exactly activate its finalized issuance result' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
  END IF;
  IF TG_OP = 'DELETE' THEN
    IF OLD.retention_until > transaction_timestamp() THEN
      RAISE EXCEPTION 'certificate retention has not elapsed' USING ERRCODE = '23514';
    END IF;
    RETURN OLD;
  END IF;
  IF OLD.status = 'revoked' THEN
    RAISE EXCEPTION 'revoked certificate is immutable' USING ERRCODE = '23514';
  END IF;
  IF (NEW.certificate_id,NEW.issuance_id,NEW.authority_operation_id,NEW.authority_epoch,NEW.authority_sequence,
      NEW.node_id,NEW.identity_epoch,NEW.lineage_id,NEW.issuer_id,NEW.serial_bytes,NEW.leaf_der,
      NEW.leaf_der_sha256,NEW.public_key_sha256,NEW.chain_der_sha256,NEW.valid_from,NEW.valid_until,NEW.created_at)
     IS DISTINCT FROM
     (OLD.certificate_id,OLD.issuance_id,OLD.authority_operation_id,OLD.authority_epoch,OLD.authority_sequence,
      OLD.node_id,OLD.identity_epoch,OLD.lineage_id,OLD.issuer_id,OLD.serial_bytes,OLD.leaf_der,
      OLD.leaf_der_sha256,OLD.public_key_sha256,OLD.chain_der_sha256,OLD.valid_from,OLD.valid_until,OLD.created_at) THEN
    RAISE EXCEPTION 'certificate identity is immutable' USING ERRCODE = '23514';
  END IF;
  IF NEW.status IS DISTINCT FROM OLD.status AND NOT (
    (OLD.status = 'recovery_pending' AND NEW.status IN ('recovery_limited','active','revoked')) OR
    (OLD.status = 'recovery_limited' AND NEW.status IN ('active','revoked')) OR
    (OLD.status = 'active' AND NEW.status = 'revoked')
  ) THEN
    RAISE EXCEPTION 'illegal certificate status transition' USING ERRCODE = '23514';
  END IF;
  RETURN NEW;
END
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION nodecontrol.enforce_security_incident_workflow()
RETURNS trigger
LANGUAGE plpgsql
AS $$
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
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION nodecontrol.enforce_security_fault_receipt()
RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
  local_count integer;
  supervisor_count integer;
BEGIN
  IF TG_OP <> 'DELETE' AND pg_catalog.current_setting('transaction_isolation') <> 'read committed' THEN
    RAISE EXCEPTION 'nodecontrol security fault receipt mutation requires read committed isolation' USING ERRCODE = '25001';
  END IF;
  IF TG_OP = 'DELETE' THEN
    RAISE EXCEPTION 'security fault receipts are immutable' USING ERRCODE = '23514';
  END IF;
  IF TG_OP = 'INSERT' THEN
    PERFORM 1 FROM nodecontrol.node_inventory WHERE node_id = NEW.node_id FOR UPDATE;
    SELECT count(*) INTO local_count
    FROM nodecontrol.node_security_fault_receipts
    WHERE node_id = NEW.node_id AND binding_status = 'active';
    IF local_count >= 64 THEN
      RAISE EXCEPTION 'local security-fault binding cap reached' USING ERRCODE = '23514';
    END IF;
    IF NEW.supervisor_fault_id IS NOT NULL THEN
      SELECT count(*) INTO supervisor_count
      FROM nodecontrol.node_security_fault_receipts
      WHERE node_id = NEW.node_id AND binding_status = 'active' AND supervisor_fault_id IS NOT NULL;
      IF supervisor_count >= 16 THEN
        RAISE EXCEPTION 'supervisor security-fault binding cap reached' USING ERRCODE = '23514';
      END IF;
    END IF;
    IF NEW.delivery_status <> 'pending' OR NEW.binding_status <> 'active'
       OR NEW.delivered_at IS NOT NULL OR NEW.cleared_at IS NOT NULL
       OR NEW.clear_attestation_digest IS NOT NULL THEN
      RAISE EXCEPTION 'security fault receipts must start pending and actively bound' USING ERRCODE = '23514';
    END IF;
  ELSE
    IF (NEW.receipt_id,NEW.authority_operation_id,NEW.authority_epoch,NEW.authority_sequence,NEW.node_id,
        NEW.identity_epoch,NEW.local_fault_id,NEW.request_digest,NEW.fault_subtype,NEW.evidence_digest,
        NEW.agent_boot_id,NEW.incident_id,NEW.supervisor_boot_id,NEW.supervisor_fault_id,
        NEW.supervisor_evidence_digest,NEW.local_binding_slot,NEW.supervisor_binding_slot,NEW.result,NEW.created_at)
       IS DISTINCT FROM
       (OLD.receipt_id,OLD.authority_operation_id,OLD.authority_epoch,OLD.authority_sequence,OLD.node_id,
        OLD.identity_epoch,OLD.local_fault_id,OLD.request_digest,OLD.fault_subtype,OLD.evidence_digest,
        OLD.agent_boot_id,OLD.incident_id,OLD.supervisor_boot_id,OLD.supervisor_fault_id,
        OLD.supervisor_evidence_digest,OLD.local_binding_slot,OLD.supervisor_binding_slot,OLD.result,OLD.created_at) THEN
      RAISE EXCEPTION 'security fault receipt binding is immutable' USING ERRCODE = '23514';
    END IF;
    IF NEW.delivery_status IS DISTINCT FROM OLD.delivery_status
       AND NOT (OLD.delivery_status = 'pending' AND NEW.delivery_status = 'deliverable') THEN
      RAISE EXCEPTION 'illegal receipt delivery transition' USING ERRCODE = '23514';
    END IF;
    IF NEW.binding_status IS DISTINCT FROM OLD.binding_status
       AND NOT (OLD.binding_status = 'active' AND NEW.binding_status = 'cleared') THEN
      RAISE EXCEPTION 'illegal receipt binding transition' USING ERRCODE = '23514';
    END IF;
    IF OLD.binding_status = 'cleared' THEN
      RAISE EXCEPTION 'cleared security fault receipt is immutable' USING ERRCODE = '23514';
    END IF;
    IF OLD.delivered_at IS NOT NULL AND NEW.delivered_at IS DISTINCT FROM OLD.delivered_at THEN
      RAISE EXCEPTION 'receipt delivery metadata is immutable' USING ERRCODE = '23514';
    END IF;
    IF OLD.cleared_at IS NOT NULL
       AND (NEW.cleared_at,NEW.clear_attestation_digest) IS DISTINCT FROM
           (OLD.cleared_at,OLD.clear_attestation_digest) THEN
      RAISE EXCEPTION 'receipt clear metadata is immutable' USING ERRCODE = '23514';
    END IF;
  END IF;
  IF NOT EXISTS (
    SELECT 1
    FROM nodecontrol.node_security_incidents incident
    JOIN nodecontrol.control_plane_authority_fences fence
      ON (fence.operation_id,fence.authority_epoch,fence.authority_sequence) =
         (NEW.authority_operation_id,NEW.authority_epoch,NEW.authority_sequence)
    WHERE incident.incident_id = NEW.incident_id
      AND (incident.node_id,incident.identity_epoch,incident.fault_subtype,
           incident.authority_operation_id,incident.authority_epoch,incident.authority_sequence) =
          (NEW.node_id,NEW.identity_epoch,NEW.fault_subtype,
           NEW.authority_operation_id,NEW.authority_epoch,NEW.authority_sequence)
      AND incident.status IN ('open','resolution_pending_agent_ack','overflow')
      AND fence.provider_status = 'committed'
      AND fence.visibility_state = 'active'
      AND fence.effect_kind = 'security_incident_open'
      AND fence.scope_kind = 'node'
  ) THEN
    RAISE EXCEPTION 'security fault receipt must bind its exact finalized incident' USING ERRCODE = '23514';
  END IF;
  RETURN NEW;
END
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION nodecontrol.enforce_recovery_session_workflow()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
    IF NEW.status <> 'pending' THEN
      RAISE EXCEPTION 'recovery sessions must start pending' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
  END IF;
  IF TG_OP = 'DELETE' THEN
    IF OLD.retention_until IS NULL OR OLD.retention_until > transaction_timestamp() THEN
      RAISE EXCEPTION 'recovery session retention has not elapsed' USING ERRCODE = '23514';
    END IF;
    RETURN OLD;
  END IF;
  IF OLD.status <> 'pending' THEN
    RAISE EXCEPTION 'terminal recovery session is immutable' USING ERRCODE = '23514';
  END IF;
  IF (NEW.recovery_id,NEW.authority_operation_id,NEW.authority_epoch,NEW.authority_sequence,
      NEW.node_id,NEW.identity_epoch,NEW.reason,NEW.version,NEW.incident_set_digest,
      NEW.resume_operator_state,NEW.created_at)
     IS DISTINCT FROM
     (OLD.recovery_id,OLD.authority_operation_id,OLD.authority_epoch,OLD.authority_sequence,
      OLD.node_id,OLD.identity_epoch,OLD.reason,OLD.version,OLD.incident_set_digest,
      OLD.resume_operator_state,OLD.created_at) THEN
    RAISE EXCEPTION 'recovery session inputs are immutable' USING ERRCODE = '23514';
  END IF;
  IF NEW.status IS DISTINCT FROM OLD.status AND NEW.status NOT IN ('completed','superseded') THEN
    RAISE EXCEPTION 'illegal recovery session status transition' USING ERRCODE = '23514';
  END IF;
  RETURN NEW;
END
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION nodecontrol.enforce_restore_approval_workflow()
RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
  pair_count integer;
BEGIN
  IF TG_WHEN = 'AFTER' THEN
    IF NEW.status = 'consumed' THEN
      SELECT count(*) INTO pair_count
      FROM nodecontrol.node_restore_reauthorization_approvals pair
      WHERE (pair.authority_operation_id,pair.authority_epoch,pair.authority_sequence,
             pair.node_id,pair.recovery_id,pair.effect_digest,pair.scope_digest,
             pair.operator_authority_epoch,pair.operator_authority_sequence,pair.authorizer_version,
             pair.security_admin_binding_digest,pair.pop_scope,pair.evidence_completed_at) =
            (NEW.authority_operation_id,NEW.authority_epoch,NEW.authority_sequence,
             NEW.node_id,NEW.recovery_id,NEW.effect_digest,NEW.scope_digest,
             NEW.operator_authority_epoch,NEW.operator_authority_sequence,NEW.authorizer_version,
             NEW.security_admin_binding_digest,NEW.pop_scope,NEW.evidence_completed_at)
        AND pair.status = 'consumed'
        AND pair.role IN ('proposal','approval')
        AND pair.expires_at >= transaction_timestamp()
        AND pair.credential_expires_at >= transaction_timestamp()
        AND pair.evidence_completed_at + interval '15 minutes' >= transaction_timestamp();
      IF pair_count <> 2 THEN
        RAISE EXCEPTION 'restore activation must atomically consume its exact proposal and approval pair' USING ERRCODE = '23514';
      END IF;
    END IF;
    RETURN NULL;
  END IF;

  IF TG_OP = 'INSERT' THEN
    IF NEW.status <> 'pending' THEN
      RAISE EXCEPTION 'restore approvals must start pending' USING ERRCODE = '23514';
    END IF;
    IF NEW.expires_at <= transaction_timestamp()
       OR NEW.credential_expires_at <= transaction_timestamp()
       OR NEW.evidence_completed_at + interval '15 minutes' <= transaction_timestamp() THEN
      RAISE EXCEPTION 'restore approval evidence is expired' USING ERRCODE = '23514';
    END IF;
    IF NOT EXISTS (
      SELECT 1
      FROM nodecontrol.node_recovery_sessions recovery
      WHERE recovery.recovery_id = NEW.recovery_id
        AND recovery.node_id = NEW.node_id
        AND recovery.status = 'completed'
    ) THEN
      RAISE EXCEPTION 'restore approval must bind a completed recovery for the same node' USING ERRCODE = '23514';
    END IF;
    IF NOT EXISTS (
      SELECT 1
      FROM nodecontrol.control_plane_authority_fences fence
      WHERE (fence.operation_id,fence.authority_epoch,fence.authority_sequence) =
            (NEW.authority_operation_id,NEW.authority_epoch,NEW.authority_sequence)
        AND fence.effect_digest = NEW.effect_digest
        AND fence.scope_digest = NEW.scope_digest
        AND fence.effect_kind = 'operator_transition'
        AND fence.scope_kind = 'node'
        AND fence.provider_status = 'committed'
        AND fence.visibility_state = 'active'
    ) THEN
      RAISE EXCEPTION 'restore approval must bind its exact committed operator-transition fence' USING ERRCODE = '23514';
    END IF;
    IF EXISTS (
      SELECT 1
      FROM nodecontrol.node_restore_reauthorization_approvals counterpart
      WHERE counterpart.node_id = NEW.node_id
        AND counterpart.recovery_id = NEW.recovery_id
        AND counterpart.role <> NEW.role
        AND counterpart.status = 'pending'
        AND (counterpart.authority_operation_id,counterpart.authority_epoch,counterpart.authority_sequence,
             counterpart.effect_digest,counterpart.scope_digest,counterpart.operator_authority_epoch,
             counterpart.operator_authority_sequence,counterpart.authorizer_version,
             counterpart.security_admin_binding_digest,counterpart.pop_scope,counterpart.evidence_completed_at)
            IS DISTINCT FROM
            (NEW.authority_operation_id,NEW.authority_epoch,NEW.authority_sequence,
             NEW.effect_digest,NEW.scope_digest,NEW.operator_authority_epoch,
             NEW.operator_authority_sequence,NEW.authorizer_version,
             NEW.security_admin_binding_digest,NEW.pop_scope,NEW.evidence_completed_at)
    ) THEN
      RAISE EXCEPTION 'restore proposal and approval evidence must match exactly' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
  END IF;
  IF TG_OP = 'DELETE' THEN
    IF OLD.retention_until IS NULL OR OLD.retention_until > transaction_timestamp() THEN
      RAISE EXCEPTION 'restore approval retention has not elapsed' USING ERRCODE = '23514';
    END IF;
    RETURN OLD;
  END IF;
  IF OLD.status <> 'pending' THEN
    RAISE EXCEPTION 'terminal restore approval is immutable' USING ERRCODE = '23514';
  END IF;
  IF (NEW.approval_id,NEW.authority_operation_id,NEW.authority_epoch,NEW.authority_sequence,
      NEW.node_id,NEW.recovery_id,NEW.effect_digest,NEW.scope_digest,NEW.role,NEW.operator_id,
      NEW.credential_digest,NEW.leaf_der_sha256,NEW.operator_authority_epoch,NEW.operator_authority_sequence,
      NEW.authorizer_version,NEW.security_admin_binding_digest,NEW.pop_scope,NEW.evidence_completed_at,
      NEW.credential_expires_at,NEW.created_at,NEW.expires_at)
     IS DISTINCT FROM
     (OLD.approval_id,OLD.authority_operation_id,OLD.authority_epoch,OLD.authority_sequence,
      OLD.node_id,OLD.recovery_id,OLD.effect_digest,OLD.scope_digest,OLD.role,OLD.operator_id,
      OLD.credential_digest,OLD.leaf_der_sha256,OLD.operator_authority_epoch,OLD.operator_authority_sequence,
      OLD.authorizer_version,OLD.security_admin_binding_digest,OLD.pop_scope,OLD.evidence_completed_at,
      OLD.credential_expires_at,OLD.created_at,OLD.expires_at) THEN
    RAISE EXCEPTION 'restore approval inputs are immutable' USING ERRCODE = '23514';
  END IF;
  IF NEW.status IS DISTINCT FROM OLD.status AND NEW.status NOT IN ('consumed','superseded') THEN
    RAISE EXCEPTION 'illegal restore approval status transition' USING ERRCODE = '23514';
  END IF;
  IF NEW.status = 'consumed'
     AND (OLD.expires_at <= transaction_timestamp()
       OR OLD.credential_expires_at <= transaction_timestamp()
       OR OLD.evidence_completed_at + interval '15 minutes' <= transaction_timestamp()) THEN
    RAISE EXCEPTION 'expired restore approval cannot be consumed' USING ERRCODE = '23514';
  END IF;
  RETURN NEW;
END
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION nodecontrol.enforce_signing_intent_workflow()
RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
  recovery_session nodecontrol.node_recovery_sessions%ROWTYPE;
  authoritative_incident_digest bytea;
  authoritative_local_digest bytea;
  authoritative_supervisor_digest bytea;
  authoritative_remediation_digest bytea;
  authoritative_required_action text;
  pending_resolution_count integer;
  distinct_remediation_count integer;
BEGIN
  IF TG_OP = 'INSERT' THEN
    PERFORM pg_catalog.pg_advisory_xact_lock(lock_key)
    FROM (
      SELECT DISTINCT pg_catalog.hashtextextended(encode(identity_value,'hex'),0) AS lock_key
      FROM unnest(ARRAY[NEW.expected_key_id,NEW.expected_public_key_digest]) AS identity_value
      ORDER BY lock_key
    ) locks;
    IF EXISTS (
      SELECT 1
      FROM nodecontrol.node_root_metadata_signature_shares share
      WHERE share.key_id IN (NEW.expected_key_id,NEW.expected_public_key_digest)
         OR share.physical_key_id IN (NEW.expected_key_id,NEW.expected_public_key_digest)
    ) THEN
      RAISE EXCEPTION 'root-share identity cannot be used by an online state signer' USING ERRCODE = '23514';
    END IF;
    IF NEW.status <> 'pending' THEN
      RAISE EXCEPTION 'state signing intents must start pending' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
  END IF;
  IF TG_OP = 'DELETE' THEN
    RAISE EXCEPTION 'state signing intents are retained and immutable' USING ERRCODE = '23514';
  END IF;
  IF OLD.status <> 'pending' THEN
    RAISE EXCEPTION 'terminal state signing intent is immutable' USING ERRCODE = '23514';
  END IF;
  IF (NEW.signing_id,NEW.authority_operation_id,NEW.authority_epoch,NEW.authority_sequence,
      NEW.node_id,NEW.signing_kind,NEW.idempotency_key_digest,NEW.base_generation,NEW.reserved_generation,
      NEW.canonical_payload,NEW.payload_digest,NEW.root_version,NEW.metadata_version,NEW.expected_key_id,
      NEW.expected_public_key_digest,NEW.captured_inventory_version,NEW.captured_identity_epoch,
      NEW.captured_security_version,NEW.recovery_id,NEW.recovery_reason,NEW.recovery_session_version,
      NEW.recovery_session_status,NEW.recovery_incident_set_digest,NEW.recovery_local_bindings_digest,
      NEW.recovery_supervisor_bindings_digest,NEW.recovery_remediation_digest,NEW.recovery_required_action,
      NEW.activation_deadline,NEW.created_at)
     IS DISTINCT FROM
     (OLD.signing_id,OLD.authority_operation_id,OLD.authority_epoch,OLD.authority_sequence,
      OLD.node_id,OLD.signing_kind,OLD.idempotency_key_digest,OLD.base_generation,OLD.reserved_generation,
      OLD.canonical_payload,OLD.payload_digest,OLD.root_version,OLD.metadata_version,OLD.expected_key_id,
      OLD.expected_public_key_digest,OLD.captured_inventory_version,OLD.captured_identity_epoch,
      OLD.captured_security_version,OLD.recovery_id,OLD.recovery_reason,OLD.recovery_session_version,
      OLD.recovery_session_status,OLD.recovery_incident_set_digest,OLD.recovery_local_bindings_digest,
      OLD.recovery_supervisor_bindings_digest,OLD.recovery_remediation_digest,OLD.recovery_required_action,
      OLD.activation_deadline,OLD.created_at) THEN
    RAISE EXCEPTION 'state signing intent inputs are immutable' USING ERRCODE = '23514';
  END IF;
  IF OLD.signature IS NOT NULL
     AND (NEW.signature,NEW.signature_verified_at) IS DISTINCT FROM (OLD.signature,OLD.signature_verified_at) THEN
    RAISE EXCEPTION 'verified state signature is immutable' USING ERRCODE = '23514';
  END IF;
  IF NEW.status IS DISTINCT FROM OLD.status AND NEW.status NOT IN ('active','failed','superseded') THEN
    RAISE EXCEPTION 'illegal state signing intent transition' USING ERRCODE = '23514';
  END IF;
  IF NEW.status = 'active' AND OLD.status = 'pending' THEN
    IF transaction_timestamp() > NEW.activation_deadline THEN
      RAISE EXCEPTION 'state signing activation deadline expired' USING ERRCODE = '23514';
    END IF;
    IF NOT EXISTS (
      SELECT 1 FROM nodecontrol.control_plane_authority_fences fence
      WHERE (fence.operation_id,fence.authority_epoch,fence.authority_sequence) =
            (NEW.authority_operation_id,NEW.authority_epoch,NEW.authority_sequence)
        AND fence.provider_status = 'committed'
        AND fence.visibility_state = 'active'
        AND fence.scope_kind = 'node'
        AND fence.effect_kind = CASE NEW.signing_kind
              WHEN 'desired' THEN 'desired_activate'
              ELSE 'recovery_activate'
            END
    ) THEN
      RAISE EXCEPTION 'state signing activation requires its committed exact fence' USING ERRCODE = '23514';
    END IF;
    IF NOT EXISTS (
      SELECT 1
      FROM nodecontrol.node_inventory inventory
      JOIN nodecontrol.node_root_metadata_publish_intents root_intent
        ON root_intent.publish_id = inventory.active_root_publish_id
       AND root_intent.publish_kind = 'root'
       AND root_intent.reserved_version = inventory.active_root_version
       AND root_intent.status = 'active'
      JOIN nodecontrol.control_plane_authority_fences root_fence
        ON (root_fence.operation_id,root_fence.authority_epoch,root_fence.authority_sequence) =
           (root_intent.authority_operation_id,root_intent.authority_epoch,root_intent.authority_sequence)
       AND root_fence.provider_status = 'committed'
       AND root_fence.visibility_state = 'active'
       AND root_fence.effect_kind = 'root_publish'
       AND root_fence.scope_kind = 'global_node_trust'
      JOIN nodecontrol.node_root_metadata_publish_intents metadata_intent
        ON metadata_intent.publish_id = inventory.active_metadata_publish_id
       AND metadata_intent.publish_kind = 'metadata'
       AND metadata_intent.reserved_version = inventory.active_metadata_version
       AND metadata_intent.base_root_version = inventory.active_root_version
       AND metadata_intent.status = 'active'
      JOIN nodecontrol.control_plane_authority_fences metadata_fence
        ON (metadata_fence.operation_id,metadata_fence.authority_epoch,metadata_fence.authority_sequence) =
           (metadata_intent.authority_operation_id,metadata_intent.authority_epoch,metadata_intent.authority_sequence)
       AND metadata_fence.provider_status = 'committed'
       AND metadata_fence.visibility_state = 'active'
       AND metadata_fence.effect_kind = 'metadata_publish'
       AND metadata_fence.scope_kind = 'global_node_trust'
      WHERE inventory.node_id = NEW.node_id
        AND inventory.inventory_version = NEW.captured_inventory_version
        AND inventory.identity_epoch = NEW.captured_identity_epoch
        AND inventory.security_version = NEW.captured_security_version
        AND inventory.active_root_version = NEW.root_version
        AND inventory.active_metadata_version = NEW.metadata_version
       AND NEW.base_generation = CASE NEW.signing_kind
              WHEN 'desired' THEN COALESCE(inventory.active_desired_generation,0)
              ELSE COALESCE(inventory.active_recovery_generation,0)
            END
        AND (NEW.signing_kind <> 'recovery' OR (
          inventory.operator_state = 'disabled'
          AND inventory.security_state = 'quarantined'
          AND inventory.identity_state IN ('recovery_pending','recovery_limited')
        ))
      FOR SHARE OF inventory,root_intent,metadata_intent
    ) THEN
      RAISE EXCEPTION 'state signing activation captured node or trust values are stale' USING ERRCODE = '23514';
    END IF;
    IF NEW.signing_kind = 'recovery' THEN
      SELECT recovery.* INTO recovery_session
      FROM nodecontrol.node_recovery_sessions recovery
      WHERE recovery.recovery_id = NEW.recovery_id
      FOR SHARE;
      IF NOT FOUND
         OR recovery_session.node_id IS DISTINCT FROM NEW.node_id
         OR recovery_session.identity_epoch IS DISTINCT FROM NEW.captured_identity_epoch
         OR recovery_session.reason IS DISTINCT FROM NEW.recovery_reason
         OR recovery_session.version IS DISTINCT FROM NEW.recovery_session_version
         OR recovery_session.status IS DISTINCT FROM 'pending'
         OR NEW.recovery_session_status IS DISTINCT FROM 'pending' THEN
        RAISE EXCEPTION 'state signing activation captured recovery session is stale' USING ERRCODE = '23514';
      END IF;

      SELECT
        pg_catalog.sha256(COALESCE(
          pg_catalog.string_agg(pg_catalog.uuid_send(incident.incident_id),''::bytea ORDER BY incident.incident_id),
          ''::bytea
        )),
        count(*) FILTER (WHERE incident.status = 'resolution_pending_agent_ack'),
        count(DISTINCT pg_catalog.encode(incident.remediation_digest,'hex'))
          FILTER (WHERE incident.status = 'resolution_pending_agent_ack'),
        CASE
          WHEN count(*) FILTER (WHERE incident.status = 'resolution_pending_agent_ack') > 0
               AND count(DISTINCT pg_catalog.encode(incident.remediation_digest,'hex'))
                   FILTER (WHERE incident.status = 'resolution_pending_agent_ack') = 1
          THEN pg_catalog.decode(
            min(pg_catalog.encode(incident.remediation_digest,'hex'))
              FILTER (WHERE incident.status = 'resolution_pending_agent_ack'),
            'hex'
          )
          ELSE NULL
        END
      INTO authoritative_incident_digest,pending_resolution_count,distinct_remediation_count,
           authoritative_remediation_digest
      FROM nodecontrol.node_security_incidents incident
      WHERE incident.node_id = NEW.node_id
        AND incident.identity_epoch = NEW.captured_identity_epoch
        AND incident.status IN ('open','resolution_pending_agent_ack','overflow');

      SELECT pg_catalog.sha256(COALESCE(
        pg_catalog.string_agg(
          pg_catalog.uuid_send(receipt.local_fault_id) || pg_catalog.uuid_send(receipt.incident_id),
          ''::bytea ORDER BY receipt.local_fault_id
        ),
        ''::bytea
      ))
      INTO authoritative_local_digest
      FROM nodecontrol.node_security_fault_receipts receipt
      JOIN nodecontrol.node_security_incidents incident ON incident.incident_id = receipt.incident_id
      WHERE receipt.node_id = NEW.node_id
        AND receipt.identity_epoch = NEW.captured_identity_epoch
        AND receipt.binding_status = 'active'
        AND incident.status IN ('open','resolution_pending_agent_ack','overflow');

      SELECT pg_catalog.sha256(COALESCE(
        pg_catalog.string_agg(
          pg_catalog.uuid_send(receipt.supervisor_boot_id) || pg_catalog.uuid_send(receipt.supervisor_fault_id) || receipt.supervisor_evidence_digest ||
          pg_catalog.uuid_send(receipt.incident_id),
          ''::bytea ORDER BY receipt.supervisor_boot_id,receipt.supervisor_fault_id
        ),
        ''::bytea
      ))
      INTO authoritative_supervisor_digest
      FROM nodecontrol.node_security_fault_receipts receipt
      JOIN nodecontrol.node_security_incidents incident ON incident.incident_id = receipt.incident_id
      WHERE receipt.node_id = NEW.node_id
        AND receipt.identity_epoch = NEW.captured_identity_epoch
        AND receipt.binding_status = 'active'
        AND receipt.supervisor_fault_id IS NOT NULL
        AND incident.status IN ('open','resolution_pending_agent_ack','overflow');

      authoritative_required_action := CASE
        WHEN pending_resolution_count > 0 THEN 'clear_security_latches'
        WHEN recovery_session.recovery_certificate_id IS NOT NULL
             AND recovery_session.attestation_digest IS NULL THEN 'submit_recovery_attestation'
        ELSE 'hold_stopped'
      END;
      IF recovery_session.incident_set_digest IS DISTINCT FROM authoritative_incident_digest
         OR NEW.recovery_incident_set_digest IS DISTINCT FROM authoritative_incident_digest
         OR NEW.recovery_local_bindings_digest IS DISTINCT FROM authoritative_local_digest
         OR NEW.recovery_supervisor_bindings_digest IS DISTINCT FROM authoritative_supervisor_digest
         OR distinct_remediation_count > 1
         OR NEW.recovery_remediation_digest IS DISTINCT FROM authoritative_remediation_digest
         OR NEW.recovery_required_action IS DISTINCT FROM authoritative_required_action THEN
        RAISE EXCEPTION 'state signing activation captured recovery bindings are stale' USING ERRCODE = '23514';
      END IF;
    END IF;
  END IF;
  RETURN NEW;
END
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION nodecontrol.enforce_root_publish_workflow()
RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
  pending_count integer;
  current_share_count integer;
  new_share_count integer;
  prior_version bigint;
  superseded_count integer;
BEGIN
  IF TG_OP = 'INSERT' THEN
    IF pg_catalog.current_setting('transaction_isolation') <> 'read committed' THEN
      RAISE EXCEPTION 'nodecontrol root/metadata publish mutation requires read committed isolation' USING ERRCODE = '25001';
    END IF;
    IF NEW.status <> 'pending' THEN
      RAISE EXCEPTION 'root/metadata publish intents must start pending' USING ERRCODE = '23514';
    END IF;
    PERFORM pg_catalog.pg_advisory_xact_lock(
      pg_catalog.hashtextextended('nodecontrol:root-metadata-pending-cap',0)
    );
    SELECT count(*) INTO pending_count
    FROM nodecontrol.node_root_metadata_publish_intents
    WHERE status = 'pending';
    IF pending_count >= 8 THEN
      RAISE EXCEPTION 'at most eight root/metadata publishes may be pending' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
  END IF;
  IF TG_OP = 'DELETE' THEN
    RAISE EXCEPTION 'root/metadata publish intents are retained' USING ERRCODE = '23514';
  END IF;
  IF (NEW.publish_id,NEW.authority_operation_id,NEW.authority_epoch,NEW.authority_sequence,
      NEW.publish_kind,NEW.reason,NEW.incident_id,NEW.base_root_version,NEW.base_metadata_version,
      NEW.reserved_version,NEW.canonical_payload,NEW.payload_digest,NEW.key_set_digest,
      NEW.current_key_ids,NEW.new_key_ids,NEW.current_threshold,NEW.new_threshold,NEW.activation_deadline,NEW.created_at)
     IS DISTINCT FROM
     (OLD.publish_id,OLD.authority_operation_id,OLD.authority_epoch,OLD.authority_sequence,
      OLD.publish_kind,OLD.reason,OLD.incident_id,OLD.base_root_version,OLD.base_metadata_version,
      OLD.reserved_version,OLD.canonical_payload,OLD.payload_digest,OLD.key_set_digest,
      OLD.current_key_ids,OLD.new_key_ids,OLD.current_threshold,OLD.new_threshold,OLD.activation_deadline,OLD.created_at) THEN
    RAISE EXCEPTION 'root/metadata publish inputs are immutable' USING ERRCODE = '23514';
  END IF;
  IF OLD.status = 'pending' THEN
    IF NEW.status IS DISTINCT FROM OLD.status THEN
      IF NEW.status NOT IN ('active','failed','superseded') THEN
        RAISE EXCEPTION 'illegal root/metadata publish transition' USING ERRCODE = '23514';
      END IF;
      IF NEW.status = 'active' THEN
        IF transaction_timestamp() > NEW.activation_deadline THEN
          RAISE EXCEPTION 'root/metadata publish activation deadline expired' USING ERRCODE = '23514';
        END IF;
        IF NOT EXISTS (
          SELECT 1
          FROM nodecontrol.control_plane_authority_fences fence
          WHERE (fence.operation_id,fence.authority_epoch,fence.authority_sequence) =
                (NEW.authority_operation_id,NEW.authority_epoch,NEW.authority_sequence)
            AND fence.provider_status = 'committed'
            AND fence.visibility_state = 'active'
            AND fence.scope_kind = 'global_node_trust'
            AND fence.effect_kind = CASE NEW.publish_kind
                  WHEN 'root' THEN 'root_publish'
                  ELSE 'metadata_publish'
                END
        ) THEN
          RAISE EXCEPTION 'root/metadata publish activation requires its committed authority fence' USING ERRCODE = '23514';
        END IF;
        IF NEW.reason = 'emergency_revoke' AND NOT EXISTS (
          SELECT 1
          FROM nodecontrol.node_security_incidents incident
          WHERE incident.incident_id = NEW.incident_id
            AND incident.status IN ('open','resolution_pending_agent_ack','overflow')
        ) THEN
          RAISE EXCEPTION 'emergency root/metadata publish requires its exact open incident' USING ERRCODE = '23514';
        END IF;
        SELECT
          count(*) FILTER (WHERE signature_role = CASE NEW.publish_kind WHEN 'root' THEN 'current_root' ELSE 'metadata' END),
          count(*) FILTER (WHERE signature_role = 'new_root')
        INTO current_share_count, new_share_count
        FROM nodecontrol.node_root_metadata_signature_shares
        WHERE publish_id = NEW.publish_id;
        IF current_share_count < NEW.current_threshold
           OR (NEW.reason = 'root_rotation' AND new_share_count < NEW.new_threshold) THEN
          RAISE EXCEPTION 'root/metadata publish lacks its captured signature threshold' USING ERRCODE = '23514';
        END IF;
        prior_version := CASE NEW.publish_kind
          WHEN 'root' THEN NEW.base_root_version
          ELSE NEW.base_metadata_version
        END;
        IF prior_version > 0 THEN
          UPDATE nodecontrol.node_root_metadata_publish_intents prior_intent
          SET status = 'superseded',
              failure_reason = 'superseded',
              terminal_at = NEW.terminal_at,
              updated_at = NEW.updated_at
          WHERE prior_intent.publish_kind = NEW.publish_kind
            AND prior_intent.reserved_version = prior_version
            AND prior_intent.status = 'active';
          GET DIAGNOSTICS superseded_count = ROW_COUNT;
          IF superseded_count <> 1 THEN
            RAISE EXCEPTION 'next root/metadata activation must atomically supersede exactly one active predecessor' USING ERRCODE = '23514';
          END IF;
        END IF;
      END IF;
    END IF;
    RETURN NEW;
  END IF;
  IF OLD.status = 'active' AND NEW.status = 'superseded' THEN
    IF pg_catalog.pg_trigger_depth() = 2 AND EXISTS (
      SELECT 1
      FROM nodecontrol.node_root_metadata_publish_intents next_intent
      JOIN nodecontrol.control_plane_authority_fences fence
        ON (fence.operation_id,fence.authority_epoch,fence.authority_sequence) =
           (next_intent.authority_operation_id,next_intent.authority_epoch,next_intent.authority_sequence)
      WHERE next_intent.publish_kind = OLD.publish_kind
        AND next_intent.status = 'pending'
        AND next_intent.reserved_version = OLD.reserved_version + 1
        AND ((OLD.publish_kind = 'root' AND next_intent.base_root_version = OLD.reserved_version)
          OR (OLD.publish_kind = 'metadata' AND next_intent.base_metadata_version = OLD.reserved_version))
        AND fence.provider_status = 'committed'
        AND fence.visibility_state = 'active'
        AND fence.effect_kind = CASE OLD.publish_kind WHEN 'root' THEN 'root_publish' ELSE 'metadata_publish' END
        AND fence.scope_kind = 'global_node_trust'
    ) THEN
      RETURN NEW;
    END IF;
    RAISE EXCEPTION 'active root/metadata publish may be superseded only by its nested next activation' USING ERRCODE = '23514';
  END IF;
  RAISE EXCEPTION 'terminal root/metadata publish intent is immutable' USING ERRCODE = '23514';
END
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION nodecontrol.enforce_root_share_binding()
RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
  intent nodecontrol.node_root_metadata_publish_intents%ROWTYPE;
BEGIN
  IF TG_OP <> 'INSERT' THEN
    RAISE EXCEPTION 'root/metadata signature shares are immutable' USING ERRCODE = '23514';
  END IF;
  PERFORM pg_catalog.pg_advisory_xact_lock(lock_key)
  FROM (
    SELECT DISTINCT pg_catalog.hashtextextended(encode(identity_value,'hex'),0) AS lock_key
    FROM unnest(ARRAY[NEW.key_id,NEW.physical_key_id]) AS identity_value
    ORDER BY lock_key
  ) locks;
  SELECT * INTO intent
  FROM nodecontrol.node_root_metadata_publish_intents
  WHERE publish_id = NEW.publish_id
  FOR SHARE;
  IF NOT FOUND OR intent.status <> 'pending' OR intent.payload_digest <> NEW.payload_digest THEN
    RAISE EXCEPTION 'signature share does not bind the pending intent payload' USING ERRCODE = '23514';
  END IF;
  IF (intent.publish_kind = 'metadata' AND NEW.signature_role <> 'metadata')
     OR (intent.publish_kind = 'root' AND NEW.signature_role = 'metadata')
     OR (intent.reason <> 'root_rotation' AND NEW.signature_role = 'new_root') THEN
    RAISE EXCEPTION 'signature share role does not match publish intent' USING ERRCODE = '23514';
  END IF;
  IF (NEW.signature_role IN ('current_root','metadata') AND NOT (NEW.key_id = ANY(intent.current_key_ids)))
     OR (NEW.signature_role = 'new_root' AND NOT (NEW.key_id = ANY(intent.new_key_ids))) THEN
    RAISE EXCEPTION 'signature share key is outside the captured threshold key set' USING ERRCODE = '23514';
  END IF;
  IF EXISTS (
    SELECT 1
    FROM nodecontrol.node_state_signing_intents signing_intent
    WHERE signing_intent.expected_key_id IN (NEW.key_id,NEW.physical_key_id)
       OR signing_intent.expected_public_key_digest IN (NEW.key_id,NEW.physical_key_id)
  ) THEN
    RAISE EXCEPTION 'online state signer identity cannot provide threshold shares' USING ERRCODE = '23514';
  END IF;
  IF NEW.key_id = NEW.physical_key_id THEN
    RAISE EXCEPTION 'logical and physical threshold identities must be distinct' USING ERRCODE = '23514';
  END IF;
  RETURN NEW;
END
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION nodecontrol.enforce_observed_state()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  IF TG_OP = 'DELETE' THEN
    RAISE EXCEPTION 'observed state high-water rows are append-only' USING ERRCODE = '23514';
  END IF;
  IF NEW.final_accepting AND NOT NEW.agent_accepting THEN
    RAISE EXCEPTION 'final accepting requires agent accepting' USING ERRCODE = '23514';
  END IF;
  IF TG_OP = 'INSERT' THEN
    IF NEW.sequence <> 1 THEN
      RAISE EXCEPTION 'a new observation boot starts at sequence one' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
  END IF;
  IF NEW.node_id IS DISTINCT FROM OLD.node_id OR NEW.created_at IS DISTINCT FROM OLD.created_at THEN
    RAISE EXCEPTION 'observation identity is immutable' USING ERRCODE = '23514';
  END IF;
  IF NEW.boot_id = OLD.boot_id THEN
    IF NEW.sequence < OLD.sequence OR NEW.sample_ended_at < OLD.sample_ended_at THEN
      RAISE EXCEPTION 'observation sequence or sample time moved backward' USING ERRCODE = '23514';
    END IF;
    IF NEW.sequence = OLD.sequence AND NEW IS DISTINCT FROM OLD THEN
      RAISE EXCEPTION 'same-sequence observation must be an exact retry' USING ERRCODE = '23514';
    END IF;
  ELSIF NEW.sequence <> 1 OR NEW.sample_ended_at < OLD.sample_ended_at THEN
    RAISE EXCEPTION 'new observation boot must start at one without time rollback' USING ERRCODE = '23514';
  END IF;
  RETURN NEW;
END
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION nodecontrol.enforce_trust_bundle_high_water()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  IF TG_OP = 'DELETE' THEN
    RAISE EXCEPTION 'trust bundle high-water rows are append-only' USING ERRCODE = '23514';
  END IF;
  IF NOT EXISTS (
    SELECT 1 FROM nodecontrol.control_plane_authority_fences fence
    WHERE (fence.operation_id,fence.authority_epoch,fence.authority_sequence) =
          (NEW.authority_operation_id,NEW.authority_epoch,NEW.authority_sequence)
      AND fence.provider_status = 'committed'
      AND fence.visibility_state = 'active'
      AND fence.effect_kind = 'trust_bundle_publish'
      AND (
        (NEW.purpose IN ('bootstrap_server','agent_server','node_client')
          AND fence.scope_kind = 'global_node_trust')
        OR (NEW.purpose IN ('operator_server','operator_client')
          AND fence.scope_kind = 'global_operator_trust')
      )
  ) THEN
    RAISE EXCEPTION 'trust bundle activation requires a committed purpose-matched global scope fence' USING ERRCODE = '23514';
  END IF;
  IF TG_OP = 'INSERT' THEN
    RETURN NEW;
  END IF;
  IF (NEW.purpose,NEW.listener_kind,NEW.trust_domain)
     IS DISTINCT FROM (OLD.purpose,OLD.listener_kind,OLD.trust_domain) THEN
    RAISE EXCEPTION 'trust bundle high-water identity is immutable' USING ERRCODE = '23514';
  END IF;
  IF (NEW.authority_operation_id,NEW.authority_epoch,NEW.authority_sequence,NEW.bundle_version) =
     (OLD.authority_operation_id,OLD.authority_epoch,OLD.authority_sequence,OLD.bundle_version) THEN
    IF NEW IS DISTINCT FROM OLD THEN
      RAISE EXCEPTION 'same trust bundle high-water value forked' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
  END IF;
  IF (NEW.authority_epoch,NEW.authority_sequence) <=
     (OLD.authority_epoch,OLD.authority_sequence)
     OR NEW.bundle_version <= OLD.bundle_version
     OR NEW.bundle_digest = OLD.bundle_digest
     OR NEW.updated_at <= OLD.updated_at
     OR NEW.cumulative_set_count < OLD.cumulative_set_count
     OR (NEW.cumulative_set_count = OLD.cumulative_set_count
         AND NEW.cumulative_set_digest <> OLD.cumulative_set_digest)
     OR (NEW.cumulative_set_count > OLD.cumulative_set_count
         AND NEW.cumulative_set_digest = OLD.cumulative_set_digest) THEN
    RAISE EXCEPTION 'trust bundle high-water rollback, fork, or non-append transition' USING ERRCODE = '23514';
  END IF;
  RETURN NEW;
END
$$;
-- +goose StatementEnd

CREATE INDEX authority_checkpoint_node
ON nodecontrol.control_plane_authority_fences USING btree (authority_epoch, scope_digest, authority_sequence DESC)
WHERE scope_kind = 'node' AND provider_status = 'committed' AND visibility_state = 'active';

CREATE INDEX authority_checkpoint_global_node_trust
ON nodecontrol.control_plane_authority_fences USING btree (authority_epoch, authority_sequence DESC)
WHERE scope_kind = 'global_node_trust' AND provider_status = 'committed' AND visibility_state = 'active';

CREATE UNIQUE INDEX node_one_unconsumed_grant_per_epoch
ON nodecontrol.node_enrollment_grants USING btree (node_id, identity_epoch)
WHERE consumed_at IS NULL AND expired_at IS NULL AND invalidated_at IS NULL;

CREATE UNIQUE INDEX node_certificate_issuer_serial_unique
ON nodecontrol.node_certificates USING btree (issuer_id, serial_bytes);

CREATE UNIQUE INDEX node_certificate_leaf_digest_unique
ON nodecontrol.node_certificates USING btree (leaf_der_sha256);

CREATE INDEX node_certificate_active_authorization
ON nodecontrol.node_certificates USING btree (node_id, identity_epoch, public_key_sha256, status, issuer_id, serial_bytes) INCLUDE (certificate_id, lineage_id, leaf_der_sha256, valid_from, valid_until, authority_operation_id, authority_epoch, authority_sequence)
WHERE status IN ('active','recovery_pending','recovery_limited');

CREATE UNIQUE INDEX node_open_incident_per_subtype
ON nodecontrol.node_security_incidents USING btree (node_id, fault_subtype)
WHERE status IN ('open','resolution_pending_agent_ack','overflow');

CREATE UNIQUE INDEX node_open_incident_per_slot
ON nodecontrol.node_security_incidents USING btree (node_id, subtype_slot)
WHERE status IN ('open','resolution_pending_agent_ack','overflow');

CREATE UNIQUE INDEX node_security_fault_receipts_supervisor_fault_unique
ON nodecontrol.node_security_fault_receipts USING btree (node_id, supervisor_boot_id, supervisor_fault_id)
WHERE supervisor_fault_id IS NOT NULL;

CREATE UNIQUE INDEX node_security_fault_receipts_local_slot_unique
ON nodecontrol.node_security_fault_receipts USING btree (node_id, local_binding_slot)
WHERE binding_status = 'active';

CREATE UNIQUE INDEX node_security_fault_receipts_supervisor_slot_unique
ON nodecontrol.node_security_fault_receipts USING btree (node_id, supervisor_binding_slot)
WHERE binding_status = 'active' AND supervisor_binding_slot IS NOT NULL;

CREATE UNIQUE INDEX node_one_pending_recovery_session
ON nodecontrol.node_recovery_sessions USING btree (node_id)
WHERE status = 'pending';

CREATE UNIQUE INDEX node_restore_approval_operator_unique
ON nodecontrol.node_restore_reauthorization_approvals USING btree (node_id, recovery_id, effect_digest, operator_id);

CREATE UNIQUE INDEX node_restore_approval_role_unique
ON nodecontrol.node_restore_reauthorization_approvals USING btree (node_id, recovery_id, effect_digest, role);

CREATE UNIQUE INDEX node_one_nonterminal_signing_intent_per_kind
ON nodecontrol.node_state_signing_intents USING btree (node_id, signing_kind)
WHERE status = 'pending';

CREATE UNIQUE INDEX node_root_publish_one_per_base_pointer
ON nodecontrol.node_root_metadata_publish_intents USING btree (base_root_version, base_metadata_version)
WHERE status = 'pending';

CREATE TRIGGER control_plane_authority_fences_enforce_update
BEFORE INSERT OR UPDATE OR DELETE ON nodecontrol.control_plane_authority_fences
FOR EACH ROW EXECUTE FUNCTION nodecontrol.enforce_authority_fence_update();

CREATE TRIGGER node_capacity_profiles_immutable_after_reference
BEFORE UPDATE OR DELETE ON nodecontrol.node_capacity_profiles
FOR EACH ROW EXECUTE FUNCTION nodecontrol.enforce_capacity_profile_immutability();

CREATE CONSTRAINT TRIGGER node_inventory_validate_pointers
AFTER INSERT OR UPDATE ON nodecontrol.node_inventory
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION nodecontrol.enforce_inventory_pointers();

CREATE TRIGGER node_process_slots_enforce_cap
BEFORE INSERT OR UPDATE ON nodecontrol.node_process_slots
FOR EACH ROW EXECUTE FUNCTION nodecontrol.enforce_process_slot_cap();

CREATE TRIGGER node_resource_envelopes_immutable
BEFORE UPDATE OR DELETE ON nodecontrol.node_resource_envelopes
FOR EACH ROW EXECUTE FUNCTION nodecontrol.reject_row_mutation();

CREATE TRIGGER node_certificate_issuances_workflow
BEFORE INSERT OR UPDATE OR DELETE ON nodecontrol.node_certificate_issuances
FOR EACH ROW EXECUTE FUNCTION nodecontrol.enforce_certificate_issuance_workflow();

CREATE TRIGGER node_enrollment_grants_workflow
BEFORE INSERT OR UPDATE OR DELETE ON nodecontrol.node_enrollment_grants
FOR EACH ROW EXECUTE FUNCTION nodecontrol.enforce_enrollment_grant_workflow();

CREATE TRIGGER node_certificates_workflow
BEFORE INSERT OR UPDATE OR DELETE ON nodecontrol.node_certificates
FOR EACH ROW EXECUTE FUNCTION nodecontrol.enforce_certificate_workflow();

CREATE TRIGGER node_security_incidents_workflow
BEFORE INSERT OR UPDATE OR DELETE ON nodecontrol.node_security_incidents
FOR EACH ROW EXECUTE FUNCTION nodecontrol.enforce_security_incident_workflow();

CREATE TRIGGER node_security_fault_receipts_enforce
BEFORE INSERT OR UPDATE OR DELETE ON nodecontrol.node_security_fault_receipts
FOR EACH ROW EXECUTE FUNCTION nodecontrol.enforce_security_fault_receipt();

CREATE TRIGGER node_recovery_sessions_workflow
BEFORE INSERT OR UPDATE OR DELETE ON nodecontrol.node_recovery_sessions
FOR EACH ROW EXECUTE FUNCTION nodecontrol.enforce_recovery_session_workflow();

CREATE TRIGGER node_restore_approvals_workflow
BEFORE INSERT OR UPDATE OR DELETE ON nodecontrol.node_restore_reauthorization_approvals
FOR EACH ROW EXECUTE FUNCTION nodecontrol.enforce_restore_approval_workflow();

CREATE CONSTRAINT TRIGGER node_restore_approvals_pair
AFTER UPDATE ON nodecontrol.node_restore_reauthorization_approvals
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION nodecontrol.enforce_restore_approval_workflow();

CREATE TRIGGER node_state_signing_intents_workflow
BEFORE INSERT OR UPDATE OR DELETE ON nodecontrol.node_state_signing_intents
FOR EACH ROW EXECUTE FUNCTION nodecontrol.enforce_signing_intent_workflow();

CREATE TRIGGER node_root_metadata_publish_intents_workflow
BEFORE INSERT OR UPDATE OR DELETE ON nodecontrol.node_root_metadata_publish_intents
FOR EACH ROW EXECUTE FUNCTION nodecontrol.enforce_root_publish_workflow();

CREATE TRIGGER node_root_metadata_signature_shares_binding
BEFORE INSERT OR UPDATE OR DELETE ON nodecontrol.node_root_metadata_signature_shares
FOR EACH ROW EXECUTE FUNCTION nodecontrol.enforce_root_share_binding();

CREATE TRIGGER node_desired_states_immutable
BEFORE UPDATE OR DELETE ON nodecontrol.node_desired_states
FOR EACH ROW EXECUTE FUNCTION nodecontrol.reject_row_mutation();

CREATE TRIGGER node_recovery_states_immutable
BEFORE UPDATE OR DELETE ON nodecontrol.node_recovery_states
FOR EACH ROW EXECUTE FUNCTION nodecontrol.reject_row_mutation();

CREATE TRIGGER node_observed_states_monotonic
BEFORE INSERT OR UPDATE OR DELETE ON nodecontrol.node_observed_states
FOR EACH ROW EXECUTE FUNCTION nodecontrol.enforce_observed_state();

CREATE TRIGGER node_operator_audit_immutable
BEFORE UPDATE OR DELETE ON nodecontrol.node_operator_audit
FOR EACH ROW EXECUTE FUNCTION nodecontrol.reject_row_mutation();

CREATE TRIGGER node_state_transitions_immutable
BEFORE UPDATE OR DELETE ON nodecontrol.node_state_transitions
FOR EACH ROW EXECUTE FUNCTION nodecontrol.reject_row_mutation();

CREATE TRIGGER control_plane_trust_bundle_high_waters_monotonic
BEFORE INSERT OR UPDATE OR DELETE ON nodecontrol.control_plane_trust_bundle_high_waters
FOR EACH ROW EXECUTE FUNCTION nodecontrol.enforce_trust_bundle_high_water();

-- +goose Down
DROP TABLE nodecontrol.control_plane_trust_bundle_high_waters;
DROP TABLE nodecontrol.node_state_transitions;
DROP TABLE nodecontrol.node_operator_audit;
DROP TABLE nodecontrol.node_observed_states;
DROP TABLE nodecontrol.node_recovery_states;
DROP TABLE nodecontrol.node_desired_states;
DROP TABLE nodecontrol.node_root_metadata_signature_shares;
DROP TABLE nodecontrol.node_root_metadata_publish_intents;
DROP TABLE nodecontrol.node_state_signing_intents;
DROP TABLE nodecontrol.node_restore_reauthorization_approvals;
DROP TABLE nodecontrol.node_recovery_sessions;
DROP TABLE nodecontrol.node_security_fault_receipts;
DROP TABLE nodecontrol.node_security_incidents;
DROP TABLE nodecontrol.node_certificates;
DROP TABLE nodecontrol.node_enrollment_grants;
DROP TABLE nodecontrol.node_certificate_issuances;
DROP TABLE nodecontrol.node_resource_envelopes;
DROP TABLE nodecontrol.node_process_slots;
DROP TABLE nodecontrol.node_endpoints;
DROP TABLE nodecontrol.node_failure_domain_membership;
DROP TABLE nodecontrol.node_inventory;
DROP TABLE nodecontrol.node_capacity_profiles;
DROP TABLE nodecontrol.node_failure_domains;
DROP TABLE nodecontrol.node_pops;
DROP TABLE nodecontrol.control_plane_authority_fences;
DROP FUNCTION nodecontrol.enforce_trust_bundle_high_water();
DROP FUNCTION nodecontrol.enforce_observed_state();
DROP FUNCTION nodecontrol.enforce_root_share_binding();
DROP FUNCTION nodecontrol.enforce_root_publish_workflow();
DROP FUNCTION nodecontrol.enforce_signing_intent_workflow();
DROP FUNCTION nodecontrol.enforce_restore_approval_workflow();
DROP FUNCTION nodecontrol.enforce_recovery_session_workflow();
DROP FUNCTION nodecontrol.enforce_security_fault_receipt();
DROP FUNCTION nodecontrol.enforce_security_incident_workflow();
DROP FUNCTION nodecontrol.enforce_certificate_workflow();
DROP FUNCTION nodecontrol.enforce_enrollment_grant_workflow();
DROP FUNCTION nodecontrol.enforce_certificate_issuance_workflow();
DROP FUNCTION nodecontrol.enforce_inventory_pointers();
DROP FUNCTION nodecontrol.enforce_process_slot_cap();
DROP FUNCTION nodecontrol.enforce_capacity_profile_immutability();
DROP FUNCTION nodecontrol.enforce_authority_fence_update();
DROP FUNCTION nodecontrol.reject_row_mutation();
DROP FUNCTION nodecontrol.bytea_array_is_sorted_unique_32(bytea[]);
DROP FUNCTION nodecontrol.text_array_is_sorted_unique(text[]);
DROP SCHEMA nodecontrol;
