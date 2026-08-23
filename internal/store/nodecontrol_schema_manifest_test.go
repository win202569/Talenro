package store_test

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

type nodeControlManifest struct {
	Version int                    `yaml:"version"`
	Schema  string                 `yaml:"schema"`
	Tables  []nodeControlTableSpec `yaml:"tables"`
}

type nodeControlTableSpec struct {
	Name        string                      `yaml:"name"`
	Columns     []nodeControlColumnSpec     `yaml:"columns"`
	PrimaryKey  nodeControlPrimaryKeySpec   `yaml:"primary_key"`
	Enums       []nodeControlEnumSpec       `yaml:"enums"`
	Constraints []nodeControlConstraintSpec `yaml:"constraints"`
	Indexes     []nodeControlIndexSpec      `yaml:"indexes"`
	Triggers    []nodeControlTriggerSpec    `yaml:"triggers"`
}

type nodeControlColumnSpec struct {
	Name       string   `yaml:"name"`
	SQLType    string   `yaml:"sql_type"`
	Nullable   bool     `yaml:"nullable"`
	DefaultSQL *string  `yaml:"default_sql"`
	Collation  *string  `yaml:"collation"`
	CheckSQL   []string `yaml:"check_sql"`
}

type nodeControlPrimaryKeySpec struct {
	Name              string   `yaml:"name"`
	Columns           []string `yaml:"columns"`
	Deferrable        bool     `yaml:"deferrable"`
	InitiallyDeferred bool     `yaml:"initially_deferred"`
}

type nodeControlEnumSpec struct {
	Name   string   `yaml:"name"`
	Column string   `yaml:"column"`
	Values []string `yaml:"values"`
}

type nodeControlConstraintSpec struct {
	Name              string   `yaml:"name"`
	Kind              string   `yaml:"kind"`
	DefinitionSQL     string   `yaml:"definition_sql"`
	Columns           []string `yaml:"columns"`
	ReferencedTable   *string  `yaml:"referenced_table"`
	ReferencedColumns []string `yaml:"referenced_columns"`
	OnUpdate          *string  `yaml:"on_update"`
	OnDelete          *string  `yaml:"on_delete"`
	Deferrable        bool     `yaml:"deferrable"`
	InitiallyDeferred bool     `yaml:"initially_deferred"`
}

type nodeControlIndexSpec struct {
	Name      string   `yaml:"name"`
	Unique    bool     `yaml:"unique"`
	Method    string   `yaml:"method"`
	Keys      []string `yaml:"keys"`
	Include   []string `yaml:"include"`
	Predicate *string  `yaml:"predicate"`
}

type nodeControlTriggerSpec struct {
	Name     string   `yaml:"name"`
	Timing   string   `yaml:"timing"`
	Events   []string `yaml:"events"`
	Function string   `yaml:"function"`
	WhenSQL  *string  `yaml:"when_sql"`
}

var compactManifestToken = regexp.MustCompile(`(?i)(^|[^a-z0-9_])(V|N|D32\??|B64K|bounded|closed|paired|optional|authority_group|retention|same as)([^a-z0-9_]|$)`)

var expectedNodeControlColumns = map[string][]string{
	"node_pops":            {"pop_code", "iso_country", "region", "operator_state", "version", "created_at", "updated_at"},
	"node_failure_domains": {"failure_domain_id", "domain_type", "stable_id", "version", "created_at", "updated_at"},
	"node_inventory": {
		"node_id", "pop_code", "operator_state", "security_state", "identity_state", "resume_operator_state",
		"pending_operator_transition", "pending_transition_signing_id", "identity_epoch", "lineage_id", "inventory_version",
		"security_version", "resource_envelope_version", "resource_envelope_digest", "active_desired_generation",
		"next_desired_generation", "active_recovery_generation", "next_recovery_generation", "active_root_publish_id",
		"active_root_version", "active_metadata_publish_id", "active_metadata_version", "last_authority_operation_id",
		"last_authority_epoch", "last_authority_sequence", "created_at", "updated_at",
	},
	"node_failure_domain_membership": {"node_id", "failure_domain_id", "domain_type", "inventory_version", "created_at"},
	"node_endpoints":                 {"endpoint_id", "node_id", "address", "port", "transport", "protocol_capability", "operator_state", "inventory_version", "created_at", "updated_at"},
	"node_process_slots":             {"node_id", "slot_id", "adapter", "capacity_profile_id", "capacity_profile_version", "required", "operator_state", "inventory_version", "created_at", "updated_at"},
	"node_capacity_profiles": {
		"profile_id", "version", "adapter", "egress_limit_bps", "connection_limit", "handshake_limit_per_second",
		"cpu_quota_millicores", "cpu_limit_basis_points", "memory_limit_bytes", "task_limit", "file_descriptor_limit",
		"queue_limit", "packet_loss_limit_basis_points", "required_metrics", "created_at",
	},
	"node_resource_envelopes": {
		"node_id", "envelope_version", "authority_operation_id", "authority_epoch", "authority_sequence", "envelope_digest",
		"canonical_package", "deployment_key_id", "signature", "max_slots", "agent_cpu_millicores", "agent_memory_bytes",
		"agent_task_limit", "agent_file_descriptor_limit", "supervisor_cpu_millicores", "supervisor_memory_bytes",
		"supervisor_task_limit", "supervisor_file_descriptor_limit", "core_parent_cpu_millicores", "core_parent_memory_bytes",
		"core_parent_task_limit", "core_parent_file_descriptor_limit", "aggregate_slot_file_descriptor_limit",
		"aggregate_slot_tmpfs_bytes", "aggregate_slot_tmpfs_inodes", "detected_host_capacity_digest", "issued_at", "created_at",
	},
	"node_enrollment_grants": {
		"grant_id", "authority_operation_id", "authority_epoch", "authority_sequence", "node_id", "identity_epoch", "token_digest",
		"csr_digest", "idempotency_digest", "created_at", "expires_at", "consumed_at", "consumption_attempt_id",
		"consumption_request_digest", "claim_authority_operation_id", "claim_authority_epoch", "claim_authority_sequence",
		"result_issuance_id", "expired_at", "invalidated_at", "terminal_reason", "terminal_at", "retention_until",
	},
	"node_certificate_issuances": {
		"issuance_id", "authority_operation_id", "authority_epoch", "authority_sequence", "node_id", "attempt_id", "issuance_kind",
		"identity_epoch", "lineage_id", "issuer_id", "csr_sha256", "public_key_sha256", "template_sha256", "request_digest",
		"status", "serial_bytes", "leaf_der", "leaf_der_sha256", "chain_der", "chain_der_sha256", "not_before", "not_after",
		"failure_reason", "created_at", "updated_at", "terminal_at", "retention_until",
	},
	"node_certificates": {
		"certificate_id", "issuance_id", "authority_operation_id", "authority_epoch", "authority_sequence", "node_id",
		"identity_epoch", "lineage_id", "issuer_id", "serial_bytes", "leaf_der", "leaf_der_sha256", "public_key_sha256",
		"chain_der_sha256", "valid_from", "valid_until", "status", "revoke_authority_operation_id", "revoke_authority_epoch",
		"revoke_authority_sequence", "revoked_at", "revoke_reason", "created_at", "updated_at", "retention_until",
	},
	"node_state_signing_intents": {
		"signing_id", "authority_operation_id", "authority_epoch", "authority_sequence", "node_id", "signing_kind",
		"idempotency_key_digest", "base_generation", "reserved_generation", "canonical_payload", "payload_digest", "root_version",
		"metadata_version", "expected_key_id", "expected_public_key_digest", "captured_inventory_version", "captured_identity_epoch",
		"captured_security_version", "recovery_id", "recovery_reason", "recovery_session_version", "recovery_session_status",
		"recovery_incident_set_digest", "recovery_local_bindings_digest", "recovery_supervisor_bindings_digest",
		"recovery_remediation_digest", "recovery_required_action", "signature", "signature_verified_at", "activation_deadline",
		"status", "failure_reason", "created_at", "updated_at", "terminal_at",
	},
	"node_root_metadata_publish_intents": {
		"publish_id", "authority_operation_id", "authority_epoch", "authority_sequence", "publish_kind", "reason", "incident_id",
		"base_root_version", "base_metadata_version", "reserved_version", "canonical_payload", "payload_digest", "key_set_digest",
		"current_key_ids", "new_key_ids", "current_threshold", "new_threshold", "activation_deadline", "published_envelope", "published_envelope_digest", "status",
		"failure_reason", "created_at", "updated_at", "terminal_at",
	},
	"node_root_metadata_signature_shares": {"publish_id", "key_id", "physical_key_id", "payload_digest", "signature_role", "signature", "verified_at"},
	"node_desired_states": {
		"node_id", "generation", "signing_id", "authority_operation_id", "authority_epoch", "authority_sequence", "inventory_version",
		"resource_envelope_version", "resource_envelope_digest", "root_version", "root_publish_id", "metadata_version",
		"metadata_publish_id", "signing_key_id", "canonical_payload", "payload_digest", "signature", "issued_at", "effective_deadline",
		"valid_until", "reason", "created_at", "retention_until",
	},
	"node_recovery_states": {
		"node_id", "recovery_generation", "signing_id", "authority_operation_id", "authority_epoch", "authority_sequence",
		"identity_epoch", "recovery_id", "recovery_reason", "recovery_session_version", "incident_set_digest", "incident_count",
		"local_fault_bindings_digest", "local_fault_binding_count", "supervisor_fault_bindings_digest", "supervisor_fault_binding_count",
		"remediation_digest", "root_version", "metadata_version", "recovery_action", "all_slots_stopped", "canonical_payload",
		"payload_digest", "signing_key_id", "signature", "issued_at", "valid_until", "created_at", "retention_until",
	},
	"node_observed_states": {
		"node_id", "boot_id", "sequence", "request_digest", "observation_digest", "canonical_observation", "sample_ended_at", "arrived_at",
		"seen_desired_generation", "seen_desired_authority_epoch", "seen_desired_authority_sequence", "seen_desired_digest",
		"applied_desired_generation", "applied_desired_authority_epoch", "applied_desired_authority_sequence", "applied_desired_digest",
		"seen_recovery_generation", "seen_recovery_authority_epoch", "seen_recovery_authority_sequence", "seen_recovery_digest",
		"applied_recovery_generation", "applied_recovery_authority_epoch", "applied_recovery_authority_sequence", "applied_recovery_digest",
		"inventory_version", "reducer_version", "reducer_input_digest", "reducer_state_digest", "health_state", "health_reason",
		"capacity_accepting", "agent_accepting", "final_accepting", "last_capacity_change_at", "last_agent_change_at",
		"last_final_change_at", "created_at", "updated_at",
	},
	"node_state_transitions": {
		"transition_id", "node_id", "authority_operation_id", "authority_epoch", "authority_sequence", "dimension", "from_state",
		"to_state", "reason", "observation_boot_id", "observation_sequence", "incident_id", "audit_id", "aggregate_version",
		"occurred_at", "retention_until",
	},
	"node_security_incidents": {
		"incident_id", "authority_operation_id", "authority_epoch", "authority_sequence", "node_id", "identity_epoch", "fault_subtype",
		"subtype_slot", "status", "first_evidence_digest", "last_evidence_digest", "occurrence_count", "trust_context_digest",
		"first_occurred_at", "last_occurred_at", "resolution_authority_operation_id", "resolution_authority_epoch",
		"resolution_authority_sequence", "remediation_digest", "resolution_at", "retention_until",
	},
	"node_security_fault_receipts": {
		"receipt_id", "authority_operation_id", "authority_epoch", "authority_sequence", "node_id", "identity_epoch", "local_fault_id",
		"request_digest", "fault_subtype", "evidence_digest", "agent_boot_id", "incident_id", "supervisor_boot_id", "supervisor_fault_id",
		"supervisor_evidence_digest", "local_binding_slot", "supervisor_binding_slot", "result", "delivery_status", "binding_status",
		"delivered_at", "cleared_at", "clear_attestation_digest", "created_at", "updated_at",
	},
	"node_recovery_sessions": {
		"recovery_id", "authority_operation_id", "authority_epoch", "authority_sequence", "node_id", "identity_epoch", "reason", "version",
		"status", "incident_set_digest", "resume_operator_state", "recovery_certificate_id", "attestation_digest", "terminal_at",
		"created_at", "updated_at", "retention_until",
	},
	"node_restore_reauthorization_approvals": {
		"approval_id", "authority_operation_id", "authority_epoch", "authority_sequence", "node_id", "recovery_id", "effect_digest",
		"scope_digest", "role", "operator_id", "credential_digest", "leaf_der_sha256", "operator_authority_epoch",
		"operator_authority_sequence", "authorizer_version", "security_admin_binding_digest", "pop_scope", "evidence_completed_at",
		"credential_expires_at", "created_at", "expires_at", "status", "terminal_at", "retention_until",
	},
	"node_operator_audit": {
		"audit_id", "command_id", "authority_operation_id", "authority_epoch", "authority_sequence", "operator_id", "credential_digest",
		"role", "action", "target_kind", "target_id", "reason", "result", "before_version", "after_version", "occurred_at", "retention_until",
	},
	"control_plane_trust_bundle_high_waters": {
		"purpose", "listener_kind", "trust_domain", "authority_operation_id", "authority_epoch", "authority_sequence", "bundle_version",
		"bundle_digest", "cumulative_set_digest", "cumulative_set_count", "updated_at",
	},
	"control_plane_authority_fences": {
		"operation_id", "effect_kind", "scope_kind", "authority_epoch", "authority_sequence", "scope_digest",
		"provider_reservation_digest", "effect_digest", "provider_status", "provider_receipt_digest", "db_system_id", "db_timeline",
		"required_lsn", "abort_reason", "visibility_state", "reserved_at", "effect_bound_at", "terminal_at",
	},
}

var expectedNodeControlEnums = map[string]map[string][]string{
	"node_pops":            {"node_pops_operator_state_enum": {"provisioning", "enabled", "draining", "disabled"}},
	"node_failure_domains": {"node_failure_domains_domain_type_enum": {"facility", "compute", "upstream"}},
	"node_inventory": {
		"node_inventory_operator_state_enum":              {"provisioning", "enabled", "draining", "disabled"},
		"node_inventory_security_state_enum":              {"normal", "quarantined"},
		"node_inventory_identity_state_enum":              {"never_enrolled", "active", "recovery_pending", "recovery_limited", "revoked"},
		"node_inventory_resume_operator_state_enum":       {"provisioning", "enabled", "draining"},
		"node_inventory_pending_operator_transition_enum": {"draining", "resume", "restore_reauthorize"},
	},
	"node_failure_domain_membership": {"node_failure_domain_membership_domain_type_enum": {"facility", "compute", "upstream"}},
	"node_endpoints": {
		"node_endpoints_transport_enum":           {"tcp", "udp"},
		"node_endpoints_protocol_capability_enum": {"bootstrap_v1", "agent_control_v1", "agent_observation_v1"},
		"node_endpoints_operator_state_enum":      {"provisioning", "enabled", "draining", "disabled"},
	},
	"node_process_slots": {
		"node_process_slots_adapter_enum":        {"fixture", "xray", "sing_box"},
		"node_process_slots_operator_state_enum": {"provisioning", "enabled", "draining", "disabled"},
	},
	"node_capacity_profiles": {"node_capacity_profiles_adapter_enum": {"fixture", "xray", "sing_box"}},
	"node_enrollment_grants": {"node_enrollment_grants_terminal_reason_enum": {"expired", "identity_epoch_advanced", "operator_disabled", "security_quarantine", "superseded"}},
	"node_certificate_issuances": {
		"node_certificate_issuances_issuance_kind_enum":  {"initial", "rotation", "recovery"},
		"node_certificate_issuances_status_enum":         {"pending", "active", "rejected", "superseded", "failed"},
		"node_certificate_issuances_failure_reason_enum": {"ca_rejected", "csr_invalid", "template_invalid", "provider_failed", "superseded"},
	},
	"node_certificates": {
		"node_certificates_status_enum":        {"active", "recovery_pending", "recovery_limited", "revoked"},
		"node_certificates_revoke_reason_enum": {"scheduled", "identity_epoch_advanced", "operator_disabled", "security_incident", "superseded"},
	},
	"node_state_signing_intents": {
		"node_state_signing_intents_signing_kind_enum":             {"desired", "recovery"},
		"node_state_signing_intents_recovery_reason_enum":          {"identity_compromise", "administrative_disable", "retire", "authority_restore", "security_incident"},
		"node_state_signing_intents_recovery_session_status_enum":  {"pending", "completed", "superseded"},
		"node_state_signing_intents_recovery_required_action_enum": {"hold_stopped", "submit_recovery_attestation", "clear_security_latches"},
		"node_state_signing_intents_status_enum":                   {"pending", "active", "failed", "superseded"},
		"node_state_signing_intents_failure_reason_enum":           {"deadline_expired", "validation_failed", "signer_failed", "superseded", "authority_aborted"},
	},
	"node_root_metadata_publish_intents": {
		"node_root_metadata_publish_intents_publish_kind_enum":   {"root", "metadata"},
		"node_root_metadata_publish_intents_reason_enum":         {"normal", "root_rotation", "emergency_revoke"},
		"node_root_metadata_publish_intents_status_enum":         {"pending", "active", "failed", "superseded"},
		"node_root_metadata_publish_intents_failure_reason_enum": {"deadline_expired", "validation_failed", "threshold_failed", "superseded", "authority_aborted"},
	},
	"node_root_metadata_signature_shares": {"node_root_metadata_signature_shares_signature_role_enum": {"current_root", "new_root", "metadata"}},
	"node_desired_states":                 {"node_desired_states_reason_enum": {"provision", "inventory_update", "capacity_change", "drain_maintenance", "administrative_disable", "retire", "identity_compromise", "host_remediation", "security_recovery", "authority_restore", "release_update"}},
	"node_recovery_states": {
		"node_recovery_states_recovery_reason_enum": {"identity_compromise", "administrative_disable", "retire", "authority_restore", "security_incident"},
		"node_recovery_states_recovery_action_enum": {"hold_stopped", "submit_recovery_attestation", "clear_security_latches"},
	},
	"node_observed_states": {
		"node_observed_states_health_state_enum":  {"unknown", "healthy", "degraded", "offline", "quarantined"},
		"node_observed_states_health_reason_enum": {"none", "warming", "capacity_pressure", "required_slot_failed", "heartbeat_timeout", "security_quarantine", "operator_disabled"},
	},
	"node_state_transitions": {
		"node_state_transitions_dimension_enum": {"operator", "security", "identity", "health", "certificate"},
		"node_state_transitions_reason_enum":    {"provision", "inventory_update", "capacity_change", "drain_maintenance", "administrative_disable", "retire", "identity_compromise", "host_remediation", "security_recovery", "authority_restore", "release_update", "observation_update", "certificate_lifecycle", "incident_capacity_exceeded"},
	},
	"node_security_incidents": {
		"node_security_incidents_fault_subtype_enum": {"identity_compromise", "online_signer_equivocation", "metadata_rollback", "root_rollback", "root_equivocation", "unverified_client_highwater_conflict", "client_highwater_ahead", "server_trust_bundle_conflict", "trusted_time_rollback_or_unavailable", "local_state_corruption_or_rollback", "release_or_process_integrity", "profile_binding_mismatch", "incident_overflow"},
		"node_security_incidents_status_enum":        {"open", "resolution_pending_agent_ack", "overflow", "resolved"},
	},
	"node_security_fault_receipts": {
		"node_security_fault_receipts_fault_subtype_enum":   {"identity_compromise", "online_signer_equivocation", "metadata_rollback", "root_rollback", "root_equivocation", "unverified_client_highwater_conflict", "client_highwater_ahead", "server_trust_bundle_conflict", "trusted_time_rollback_or_unavailable", "local_state_corruption_or_rollback", "release_or_process_integrity", "profile_binding_mismatch", "incident_overflow"},
		"node_security_fault_receipts_result_enum":          {"accepted"},
		"node_security_fault_receipts_delivery_status_enum": {"pending", "deliverable"},
		"node_security_fault_receipts_binding_status_enum":  {"active", "cleared"},
	},
	"node_recovery_sessions": {
		"node_recovery_sessions_reason_enum":                {"identity_compromise", "administrative_disable", "retire", "authority_restore", "security_incident"},
		"node_recovery_sessions_status_enum":                {"pending", "completed", "superseded"},
		"node_recovery_sessions_resume_operator_state_enum": {"provisioning", "enabled", "draining", "disabled"},
	},
	"node_restore_reauthorization_approvals": {
		"node_restore_reauthorization_approvals_role_enum":   {"proposal", "approval"},
		"node_restore_reauthorization_approvals_status_enum": {"pending", "consumed", "superseded"},
	},
	"node_operator_audit": {
		"node_operator_audit_role_enum":        {"inventory_reader", "inventory_writer", "node_security_admin"},
		"node_operator_audit_action_enum":      {"create_pop", "update_inventory", "register_envelope", "drain_node", "disable_node", "restore_reauthorize", "clear_security", "retire_node"},
		"node_operator_audit_target_kind_enum": {"pop", "node", "failure_domain", "capacity_profile", "recovery_session", "security_incident"},
		"node_operator_audit_reason_enum":      {"provision", "inventory_update", "capacity_change", "drain_maintenance", "administrative_disable", "retire", "identity_compromise", "host_remediation", "security_recovery", "authority_restore", "release_update"},
		"node_operator_audit_result_enum":      {"accepted", "rejected", "conflict"},
	},
	"control_plane_trust_bundle_high_waters": {
		"control_plane_trust_bundle_high_waters_purpose_enum":       {"bootstrap_server", "agent_server", "operator_server", "node_client", "operator_client"},
		"control_plane_trust_bundle_high_waters_listener_kind_enum": {"bootstrap", "agent", "operator", "outbound_node", "outbound_operator"},
	},
	"control_plane_authority_fences": {
		"control_plane_authority_fences_effect_kind_enum":      {"trust_bundle_publish", "root_publish", "metadata_publish", "grant_create", "grant_claim", "certificate_activate", "certificate_revoke", "identity_epoch_advance", "security_incident_open", "security_incident_resolve", "resource_envelope_activate", "desired_activate", "recovery_activate", "operator_transition", "operator_authorizer_change"},
		"control_plane_authority_fences_scope_kind_enum":       {"node", "global_node_trust", "global_operator_trust"},
		"control_plane_authority_fences_provider_status_enum":  {"reserved", "committed", "aborted"},
		"control_plane_authority_fences_abort_reason_enum":     {"validation_failed", "superseded", "provider_dependency_failed", "activation_deadline_expired"},
		"control_plane_authority_fences_visibility_state_enum": {"fence_pending", "active", "aborted"},
	},
}

func TestNodeControlManifestIsLiteralAndExact(t *testing.T) {
	manifest, raw := loadNodeControlManifest(t)
	if manifest.Version != 1 || manifest.Schema != "nodecontrol" {
		t.Fatalf("manifest identity = version %d schema %q, want version 1 schema nodecontrol", manifest.Version, manifest.Schema)
	}

	migration, err := os.ReadFile("../../db/migrations/00006_nodecontrol.sql")
	if err != nil {
		t.Fatal(err)
	}
	assertExactNamedSet(t, "manifest tables", tableNames(manifest.Tables), nodeControlAuthorityTables)
	for _, table := range manifest.Tables {
		assertExactNamedSet(t, table.Name+" columns", columnNames(table.Columns), expectedNodeControlColumns[table.Name])
		assertExactEnums(t, table)
		assertManifestTableObjects(t, migration, table)
	}
	assertTerminalIssuanceRetention(t, manifest)
	assertTerminalWorkflowsStartPending(t, manifest)
	if compactManifestToken.Match(raw) {
		t.Fatalf("manifest contains a compact or qualitative semantic token: %q", compactManifestToken.Find(raw))
	}
}

func assertTerminalWorkflowsStartPending(t *testing.T, manifest nodeControlManifest) {
	t.Helper()
	want := map[string]string{
		"node_certificate_issuances":              "node_certificate_issuances_workflow",
		"node_recovery_sessions":                 "node_recovery_sessions_workflow",
		"node_restore_reauthorization_approvals": "node_restore_approvals_workflow",
		"node_state_signing_intents":              "node_state_signing_intents_workflow",
		"node_root_metadata_publish_intents":      "node_root_metadata_publish_intents_workflow",
	}
	for _, table := range manifest.Tables {
		triggerName, required := want[table.Name]
		if !required {
			continue
		}
		found := false
		for _, trigger := range table.Triggers {
			if trigger.Name == triggerName {
				found = true
				if !containsString(trigger.Events, "INSERT") {
					t.Fatalf("%s must guard INSERT so terminal workflow rows cannot bypass pending", triggerName)
				}
			}
		}
		if !found {
			t.Fatalf("%s lacks workflow trigger %s", table.Name, triggerName)
		}
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func assertTerminalIssuanceRetention(t *testing.T, manifest nodeControlManifest) {
	t.Helper()
	for _, table := range manifest.Tables {
		if table.Name != "node_certificate_issuances" {
			continue
		}
		for _, constraint := range table.Constraints {
			if constraint.Name == "node_certificate_issuances_status_result" {
				if strings.Count(constraint.DefinitionSQL, "retention_until IS NOT NULL") < 2 {
					t.Fatal("terminal certificate issuance states must require an explicit retention deadline")
				}
				return
			}
		}
	}
	t.Fatal("manifest lacks node_certificate_issuances_status_result")
}

func TestNodeControlManifestDecoderRejectsAliasesMergesAndUnknownFields(t *testing.T) {
	_, raw := loadNodeControlManifest(t)
	unknown := bytes.Replace(raw, []byte("schema: nodecontrol"), []byte("schema: nodecontrol\nunknown_catalog_field: true"), 1)
	if _, err := decodeNodeControlManifest(unknown); err == nil {
		t.Fatal("KnownFields decoder accepted an unknown manifest field")
	}
	anchored := bytes.Replace(raw, []byte("version: 1"), []byte("version: &manifest_version 1"), 1)
	if _, err := decodeNodeControlManifest(anchored); err == nil {
		t.Fatal("manifest decoder accepted a YAML anchor")
	}
	merged := bytes.Replace(raw, []byte("schema: nodecontrol"), []byte("schema: nodecontrol\n<<: {unexpected: true}"), 1)
	if _, err := decodeNodeControlManifest(merged); err == nil {
		t.Fatal("manifest decoder accepted a YAML merge key")
	}
}

func loadNodeControlManifest(t *testing.T) (nodeControlManifest, []byte) {
	t.Helper()
	raw, err := os.ReadFile("../../db/schema/nodecontrol.v1.yaml")
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := decodeNodeControlManifest(raw)
	if err != nil {
		t.Fatal(err)
	}
	return manifest, raw
}

func decodeNodeControlManifest(raw []byte) (nodeControlManifest, error) {
	var document yaml.Node
	if err := yaml.Unmarshal(raw, &document); err != nil {
		return nodeControlManifest{}, fmt.Errorf("parse manifest YAML: %w", err)
	}
	if err := rejectYAMLIndirection(&document); err != nil {
		return nodeControlManifest{}, err
	}
	if err := requireManifestFields(&document); err != nil {
		return nodeControlManifest{}, err
	}
	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	decoder.KnownFields(true)
	var manifest nodeControlManifest
	if err := decoder.Decode(&manifest); err != nil {
		return nodeControlManifest{}, fmt.Errorf("decode manifest: %w", err)
	}
	return manifest, nil
}

func rejectYAMLIndirection(node *yaml.Node) error {
	if node.Kind == yaml.AliasNode || node.Anchor != "" || (node.Kind == yaml.ScalarNode && node.Value == "<<") {
		return fmt.Errorf("YAML anchors, aliases, and merge keys are forbidden")
	}
	for _, child := range node.Content {
		if err := rejectYAMLIndirection(child); err != nil {
			return err
		}
	}
	return nil
}

func requireManifestFields(document *yaml.Node) error {
	if len(document.Content) != 1 {
		return fmt.Errorf("manifest must contain one YAML document")
	}
	root := document.Content[0]
	if err := requireMappingKeys(root, "manifest", "version", "schema", "tables"); err != nil {
		return err
	}
	tables := mappingValue(root, "tables")
	if tables == nil || tables.Kind != yaml.SequenceNode {
		return fmt.Errorf("manifest tables must be a sequence")
	}
	for tableIndex, table := range tables.Content {
		path := fmt.Sprintf("tables[%d]", tableIndex)
		if err := requireMappingKeys(table, path, "name", "columns", "primary_key", "enums", "constraints", "indexes", "triggers"); err != nil {
			return err
		}
		if err := requireSequenceObjectKeys(mappingValue(table, "columns"), path+".columns", "name", "sql_type", "nullable", "default_sql", "collation", "check_sql"); err != nil {
			return err
		}
		if err := requireMappingKeys(mappingValue(table, "primary_key"), path+".primary_key", "name", "columns", "deferrable", "initially_deferred"); err != nil {
			return err
		}
		if err := requireSequenceObjectKeys(mappingValue(table, "enums"), path+".enums", "name", "column", "values"); err != nil {
			return err
		}
		if err := requireSequenceObjectKeys(mappingValue(table, "constraints"), path+".constraints", "name", "kind", "definition_sql", "columns", "referenced_table", "referenced_columns", "on_update", "on_delete", "deferrable", "initially_deferred"); err != nil {
			return err
		}
		if err := requireSequenceObjectKeys(mappingValue(table, "indexes"), path+".indexes", "name", "unique", "method", "keys", "include", "predicate"); err != nil {
			return err
		}
		if err := requireSequenceObjectKeys(mappingValue(table, "triggers"), path+".triggers", "name", "timing", "events", "function", "when_sql"); err != nil {
			return err
		}
	}
	return nil
}

func requireSequenceObjectKeys(sequence *yaml.Node, path string, keys ...string) error {
	if sequence == nil || sequence.Kind != yaml.SequenceNode {
		return fmt.Errorf("%s must be a sequence", path)
	}
	for index, object := range sequence.Content {
		if err := requireMappingKeys(object, fmt.Sprintf("%s[%d]", path, index), keys...); err != nil {
			return err
		}
	}
	return nil
}

func requireMappingKeys(mapping *yaml.Node, path string, keys ...string) error {
	if mapping == nil || mapping.Kind != yaml.MappingNode {
		return fmt.Errorf("%s must be a mapping", path)
	}
	present := make(map[string]bool, len(mapping.Content)/2)
	for index := 0; index < len(mapping.Content); index += 2 {
		present[mapping.Content[index].Value] = true
	}
	for _, key := range keys {
		if !present[key] {
			return fmt.Errorf("%s omits explicit field %s", path, key)
		}
	}
	return nil
}

func mappingValue(mapping *yaml.Node, key string) *yaml.Node {
	if mapping == nil || mapping.Kind != yaml.MappingNode {
		return nil
	}
	for index := 0; index < len(mapping.Content); index += 2 {
		if mapping.Content[index].Value == key {
			return mapping.Content[index+1]
		}
	}
	return nil
}

func assertManifestTableObjects(t *testing.T, migration []byte, table nodeControlTableSpec) {
	t.Helper()
	if table.PrimaryKey.Name == "" || len(table.PrimaryKey.Columns) == 0 {
		t.Fatalf("%s must declare a stable named primary key", table.Name)
	}
	assertUniqueNames(t, table.Name+" constraints", constraintNames(table.Constraints))
	assertUniqueNames(t, table.Name+" indexes", indexNames(table.Indexes))
	assertUniqueNames(t, table.Name+" triggers", triggerNames(table.Triggers))
	assertUniqueNames(t, table.Name+" enums", enumNames(table.Enums))
	if !bytes.Contains(migration, []byte("CONSTRAINT "+table.PrimaryKey.Name+" PRIMARY KEY")) {
		t.Fatalf("migration lacks primary key counterpart %s", table.PrimaryKey.Name)
	}
	columns := make(map[string]nodeControlColumnSpec, len(table.Columns))
	for _, column := range table.Columns {
		columns[column.Name] = column
		if column.Name != "required_lsn" && (strings.HasSuffix(column.Name, "_at") || strings.HasSuffix(column.Name, "_until") || strings.HasSuffix(column.Name, "_deadline")) && column.SQLType != "timestamp with time zone" {
			t.Fatalf("%s.%s has sql_type %q, want timestamp with time zone", table.Name, column.Name, column.SQLType)
		}
		if column.SQLType == "text" || strings.HasPrefix(column.SQLType, "character(") {
			if column.Collation == nil || *column.Collation != "C" || len(column.CheckSQL) == 0 {
				t.Fatalf("%s.%s must spell out C collation and at least one exact text check", table.Name, column.Name)
			}
		}
		if strings.HasSuffix(column.Name, "_digest") || strings.HasSuffix(column.Name, "_sha256") || column.Name == "deployment_key_id" || column.Name == "expected_key_id" || column.Name == "signing_key_id" || column.Name == "credential_digest" {
			joined := strings.Join(column.CheckSQL, " ")
			if !strings.Contains(joined, "octet_length("+column.Name+") = 32") && !strings.Contains(joined, column.Name+" IS NULL OR octet_length("+column.Name+") = 32") {
				t.Fatalf("%s.%s does not carry an exact 32-byte check", table.Name, column.Name)
			}
		}
		if column.Name == "signature" && !strings.Contains(strings.Join(column.CheckSQL, " "), "octet_length(signature) = 64") {
			t.Fatalf("%s.signature does not carry an exact 64-byte check", table.Name)
		}
	}
	for _, constraint := range table.Constraints {
		if constraint.Name == "" || constraint.DefinitionSQL == "" {
			t.Fatalf("%s contains an unnamed or nonliteral constraint", table.Name)
		}
		if !bytes.Contains(migration, []byte("CONSTRAINT "+constraint.Name+" ")) {
			t.Fatalf("migration lacks constraint counterpart %s", constraint.Name)
		}
		if constraint.Kind == "foreign_key" {
			if constraint.OnDelete == nil || constraint.OnUpdate == nil || *constraint.OnDelete != "NO ACTION" || *constraint.OnUpdate != "NO ACTION" {
				t.Fatalf("foreign key %s must explicitly use NO ACTION", constraint.Name)
			}
		}
	}
	for _, index := range table.Indexes {
		if index.Name == "" || index.Method == "" || len(index.Keys) == 0 {
			t.Fatalf("%s contains a nonliteral index", table.Name)
		}
		if !bytes.Contains(migration, []byte("INDEX "+index.Name)) {
			t.Fatalf("migration lacks index counterpart %s", index.Name)
		}
	}
	for _, trigger := range table.Triggers {
		if trigger.Name == "" || trigger.Function == "" || len(trigger.Events) == 0 {
			t.Fatalf("%s contains a nonliteral trigger", table.Name)
		}
		if !bytes.Contains(migration, []byte("TRIGGER "+trigger.Name)) {
			t.Fatalf("migration lacks trigger counterpart %s", trigger.Name)
		}
	}
	assertAuthorityGroups(t, table.Name, columns, table.Constraints)
}

func assertAuthorityGroups(t *testing.T, table string, columns map[string]nodeControlColumnSpec, constraints []nodeControlConstraintSpec) {
	t.Helper()
	for columnName, column := range columns {
		if !strings.HasSuffix(columnName, "authority_operation_id") {
			continue
		}
		prefix := strings.TrimSuffix(columnName, "operation_id")
		epochName := prefix + "epoch"
		sequenceName := prefix + "sequence"
		epoch, epochOK := columns[epochName]
		sequence, sequenceOK := columns[sequenceName]
		if !epochOK || !sequenceOK || epoch.Nullable != column.Nullable || sequence.Nullable != column.Nullable {
			t.Fatalf("%s.%s lacks an exact nullability-matched authority epoch/sequence group", table, columnName)
		}
		found := false
		for _, constraint := range constraints {
			if constraint.Kind == "foreign_key" && equalStrings(constraint.Columns, []string{columnName, epochName, sequenceName}) && constraint.ReferencedTable != nil && *constraint.ReferencedTable == "control_plane_authority_fences" && equalStrings(constraint.ReferencedColumns, []string{"operation_id", "authority_epoch", "authority_sequence"}) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("%s.%s lacks the exact composite authority fence foreign key", table, columnName)
		}
	}
}

func assertExactEnums(t *testing.T, table nodeControlTableSpec) {
	t.Helper()
	want := expectedNodeControlEnums[table.Name]
	columns := make(map[string]struct{}, len(table.Columns))
	for _, column := range table.Columns {
		columns[column.Name] = struct{}{}
	}
	constraints := make(map[string]nodeControlConstraintSpec, len(table.Constraints))
	for _, constraint := range table.Constraints {
		constraints[constraint.Name] = constraint
	}
	got := make(map[string][]string, len(table.Enums))
	for _, enum := range table.Enums {
		if _, exists := got[enum.Name]; exists {
			t.Fatalf("%s has duplicate enum %s", table.Name, enum.Name)
		}
		got[enum.Name] = enum.Values
		if _, exists := columns[enum.Column]; !exists {
			t.Fatalf("%s enum %s names unknown column %s", table.Name, enum.Name, enum.Column)
		}
		constraint, exists := constraints[enum.Name]
		if !exists || constraint.Kind != "check" || !containsString(constraint.Columns, enum.Column) {
			t.Fatalf("%s enum %s lacks its exact catalog check counterpart", table.Name, enum.Name)
		}
		for _, value := range enum.Values {
			if !strings.Contains(constraint.DefinitionSQL, "'"+strings.ReplaceAll(value, "'", "''")+"'") {
				t.Fatalf("%s enum %s value %q is absent from its catalog check", table.Name, enum.Name, value)
			}
		}
	}
	if len(got) != len(want) {
		t.Fatalf("%s enum count = %d, want %d", table.Name, len(got), len(want))
	}
	for name, values := range want {
		if !equalStrings(got[name], values) {
			t.Fatalf("%s enum %s values = %v, want %v", table.Name, name, got[name], values)
		}
	}
}

func assertExactNamedSet(t *testing.T, label string, got, want []string) {
	t.Helper()
	gotCopy := append([]string(nil), got...)
	wantCopy := append([]string(nil), want...)
	sort.Strings(gotCopy)
	sort.Strings(wantCopy)
	if !equalStrings(gotCopy, wantCopy) {
		t.Fatalf("%s = %v, want exactly %v", label, gotCopy, wantCopy)
	}
	assertUniqueNames(t, label, got)
}

func assertUniqueNames(t *testing.T, label string, names []string) {
	t.Helper()
	seen := make(map[string]struct{}, len(names))
	for _, name := range names {
		if name == "" {
			t.Fatalf("%s contains an empty name", label)
		}
		if _, duplicate := seen[name]; duplicate {
			t.Fatalf("%s contains duplicate name %s", label, name)
		}
		seen[name] = struct{}{}
	}
}

func tableNames(tables []nodeControlTableSpec) []string {
	names := make([]string, 0, len(tables))
	for _, table := range tables {
		names = append(names, table.Name)
	}
	return names
}

func columnNames(columns []nodeControlColumnSpec) []string {
	names := make([]string, 0, len(columns))
	for _, column := range columns {
		names = append(names, column.Name)
	}
	return names
}

func constraintNames(constraints []nodeControlConstraintSpec) []string {
	names := make([]string, 0, len(constraints))
	for _, constraint := range constraints {
		names = append(names, constraint.Name)
	}
	return names
}

func indexNames(indexes []nodeControlIndexSpec) []string {
	names := make([]string, 0, len(indexes))
	for _, index := range indexes {
		names = append(names, index.Name)
	}
	return names
}

func triggerNames(triggers []nodeControlTriggerSpec) []string {
	names := make([]string, 0, len(triggers))
	for _, trigger := range triggers {
		names = append(names, trigger.Name)
	}
	return names
}

func enumNames(enums []nodeControlEnumSpec) []string {
	names := make([]string, 0, len(enums))
	for _, enum := range enums {
		names = append(names, enum.Name)
	}
	return names
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
