package contracts_test

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	nodeagentv1 "talenro.local/platform/gen/go/talenro/nodeagent/v1"
	nodebootstrapv1 "talenro.local/platform/gen/go/talenro/nodebootstrap/v1"
	nodeoperatorv1 "talenro.local/platform/gen/go/talenro/nodeoperator/v1"
)

var (
	_ nodebootstrapv1.ClientInterface              = (*nodebootstrapv1.Client)(nil)
	_ nodebootstrapv1.ClientWithResponsesInterface = (*nodebootstrapv1.ClientWithResponses)(nil)
	_ nodeagentv1.ClientInterface                  = (*nodeagentv1.Client)(nil)
	_ nodeagentv1.ClientWithResponsesInterface     = (*nodeagentv1.ClientWithResponses)(nil)
	_ nodeoperatorv1.ClientInterface               = (*nodeoperatorv1.Client)(nil)
	_ nodeoperatorv1.ClientWithResponsesInterface  = (*nodeoperatorv1.ClientWithResponses)(nil)
	_                                              = nodebootstrapv1.HandlerWithOptions
	_                                              = nodeagentv1.HandlerWithOptions
	_                                              = nodeoperatorv1.HandlerWithOptions
	_ nodebootstrapv1.ServerInterface
	_ nodeagentv1.ServerInterface
	_ nodeoperatorv1.ServerInterface
	_ nodeagentv1.ConflictArtifactKindV1        = (nodeagentv1.SignedConflictArtifactV1{}).Kind
	_ nodeagentv1.TrustConflictEvidenceResultV1 = (nodeagentv1.TrustConflictEvidenceAckV1{}).Result
	_ *nodeagentv1.CanonicalUUID                = (nodeagentv1.PollNodeDesiredStateResponse409Headers{}).TalenroTrustConflictIncidentID
	_ *nodeagentv1.CanonicalUUID                = (nodeagentv1.PollNodeRecoveryStateResponse409Headers{}).TalenroTrustConflictIncidentID
)

type operationExpectation struct {
	method          string
	path            string
	operationID     string
	requestSchema   string
	successSchemas  map[string]string
	requestHeaders  []string
	responseHasETag bool
	wireBytes       int
	decodedBytes    int
}

type nodeSpecExpectation struct {
	name        string
	title       string
	auth        string
	get         func() (*openapi3.T, error)
	operations  []operationExpectation
	schemaNames []string
}

func TestNodeOpenAPISpecsAreIsolatedAndValid(t *testing.T) {
	tests := []struct {
		name       string
		get        func() (*openapi3.T, error)
		wantPath   string
		forbidPath string
	}{
		{name: "bootstrap", get: nodebootstrapv1.GetSwagger, wantPath: "/v1/node-enrollments/claim", forbidPath: "/v1/node-agent/desired-state:poll"},
		{name: "agent", get: nodeagentv1.GetSwagger, wantPath: "/v1/node-agent/desired-state:poll", forbidPath: "/v1/operator/nodes"},
		{name: "operator", get: nodeoperatorv1.GetSwagger, wantPath: "/v1/operator/nodes", forbidPath: "/v1/node-enrollments/claim"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			spec, err := test.get()
			if err != nil {
				t.Fatal(err)
			}
			if err := spec.Validate(context.Background()); err != nil {
				t.Fatal(err)
			}
			if spec.Paths.Find(test.wantPath) == nil || spec.Paths.Find(test.forbidPath) != nil {
				t.Fatalf("listener path isolation failed")
			}
		})
	}
}

func TestNodeOpenAPIOperationMatrix(t *testing.T) {
	for _, test := range nodeSpecExpectations() {
		t.Run(test.name, func(t *testing.T) {
			spec := loadAndValidateNodeSpec(t, test.get)
			if spec.OpenAPI != "3.0.3" {
				t.Errorf("OpenAPI version = %q, want 3.0.3", spec.OpenAPI)
			}
			if spec.Info == nil || spec.Info.Title != test.title || spec.Info.Version != "1.0.0" {
				t.Errorf("info = %#v, want title %q version 1.0.0", spec.Info, test.title)
			}
			if len(spec.Security) != 0 {
				t.Errorf("top-level security = %#v, want none", spec.Security)
			}
			if spec.Components != nil && len(spec.Components.SecuritySchemes) != 0 {
				t.Errorf("security schemes = %v, want none; peer identity belongs to TLS middleware", mapKeys(spec.Components.SecuritySchemes))
			}

			wantPaths := make(map[string]int)
			seenOperationIDs := make(map[string]bool)
			for _, operationTest := range test.operations {
				wantPaths[operationTest.path]++
				item := spec.Paths.Find(operationTest.path)
				if item == nil {
					t.Fatalf("missing path %s", operationTest.path)
				}
				operation := operationForMethod(item, operationTest.method)
				if operation == nil {
					t.Fatalf("missing %s %s", operationTest.method, operationTest.path)
				}
				if operation.OperationID != operationTest.operationID {
					t.Errorf("%s %s operation ID = %q, want %q", operationTest.method, operationTest.path, operation.OperationID, operationTest.operationID)
				}
				if seenOperationIDs[operation.OperationID] {
					t.Errorf("operation ID %q appears more than once", operation.OperationID)
				}
				seenOperationIDs[operation.OperationID] = true
				assertRuntimeTLSContract(t, test.auth, operationTest, item, operation)
				assertRequestContract(t, operationTest, operation)
				assertSuccessContract(t, operationTest, operation)
			}

			if spec.Paths.Len() != len(wantPaths) {
				t.Errorf("path count = %d, want %d", spec.Paths.Len(), len(wantPaths))
			}
			for path, wantOperations := range wantPaths {
				if got := pathItemOperationCount(spec.Paths.Find(path)); got != wantOperations {
					t.Errorf("%s operation count = %d, want %d", path, got, wantOperations)
				}
			}
			assertExactStringSet(t, "component schemas", schemaNames(spec), test.schemaNames)
		})
	}
}

func TestNodeOpenAPISchemaContracts(t *testing.T) {
	bootstrap := loadAndValidateNodeSpec(t, nodebootstrapv1.GetSwagger)
	agent := loadAndValidateNodeSpec(t, nodeagentv1.GetSwagger)
	operator := loadAndValidateNodeSpec(t, nodeoperatorv1.GetSwagger)

	for name, spec := range map[string]*openapi3.T{"bootstrap": bootstrap, "agent": agent, "operator": operator} {
		t.Run(name+" canonical primitives", func(t *testing.T) {
			if got := mustSchema(t, spec, "CanonicalUUID").Pattern; got != `^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$` {
				t.Errorf("CanonicalUUID pattern = %q", got)
			}
			if got := mustSchema(t, spec, "DigestHex32").Pattern; got != `^[0-9a-f]{64}$` {
				t.Errorf("DigestHex32 pattern = %q", got)
			}
			base64URL32 := mustSchema(t, spec, "Base64URL32")
			if base64URL32.MinLength != 43 || base64URL32.MaxLength == nil || *base64URL32.MaxLength != 43 || base64URL32.Pattern != `^[A-Za-z0-9_-]{43}$` {
				t.Errorf("Base64URL32 contract = %#v", base64URL32)
			}
		})
	}

	for name, spec := range map[string]*openapi3.T{"bootstrap": bootstrap, "agent": agent, "operator": operator} {
		t.Run(name+" public error", func(t *testing.T) {
			assertObjectShape(t, spec, "PublicErrorV1", []string{"code", "request_id"}, []string{"code", "request_id", "reason", "retry_after_ms"})
			assertEnum(t, spec, "PublicErrorV1", "code", []string{"invalid_request", "unauthenticated", "forbidden", "not_found", "conflict", "rate_limited", "dependency_unavailable", "internal"})
			assertEnum(t, spec, "PublicErrorV1", "reason", []string{"malformed", "precondition_failed", "stale_version", "reenroll_required", "incident_capacity_exceeded", "authority_unavailable", "credential_invalid", "scope_changed", "operation_in_progress", "deadline_expired", "unsupported_capability"})
			assertPublicErrorBounds(t, spec)
		})
	}

	assertObjectShape(t, bootstrap, "ClaimNodeEnrollmentRequestV1",
		[]string{"attempt_id", "node_id", "enrollment_grant", "csr_der", "time_nonce"},
		[]string{"attempt_id", "node_id", "enrollment_grant", "csr_der", "time_nonce"})
	assertObjectShape(t, bootstrap, "CertificateAuthorizationV1",
		[]string{"leaf_der", "issuer_chain", "receipt", "node_state_trust_metadata", "time_attestation"},
		[]string{"leaf_der", "issuer_chain", "receipt", "node_state_trust_metadata", "time_attestation"})
	assertCSRDER(t, bootstrap)
	assertCSRDER(t, agent)

	agentShapes := map[string][]string{
		"RotateNodeCertificateRequestV1": {"attempt_id", "csr_der", "time_nonce"},
		"DesiredPollRequestV1":           {"boot_id", "agent_build_digest", "capability_schema_digest", "high_water", "time_nonce"},
		"RecoveryPollRequestV1":          {"boot_id", "recovery_id", "sorted_known_incident_ids", "high_water", "time_nonce"},
		"NodeStatePollResponseV1":        {"root_chain", "metadata", "desired_state", "time_attestation"},
		"NodeRecoveryPollResponseV1":     {"root_chain", "metadata", "recovery_state", "time_attestation"},
		"NodeObservationV1":              {"node_id", "identity_epoch", "boot_id", "sequence", "observed_at", "agent_build_digest", "capability_schema_digest", "seen_generation", "seen_authority_sequence", "seen_digest", "applied_generation", "applied_authority_sequence", "applied_digest", "sorted_slot_facts", "reducer_version", "reducer_digest", "reducer_output"},
		"ObservationAckV1":               {"boot_id", "sequence", "report_digest", "result"},
		"SecurityFaultReportV1":          {"operation_id", "local_fault_id", "subtype", "evidence_digest", "identity_epoch", "boot_id", "request_digest", "supervisor_fault"},
		"SecurityFaultReceiptV1":         {"local_fault_id", "incident_id", "request_digest", "authority_epoch", "authority_sequence", "result"},
		"TrustConflictEvidenceRequestV1": {"incident_id", "artifacts"},
		"SignedConflictArtifactV1":       {"kind", "canonical_bytes"},
		"TrustConflictEvidenceAckV1":     {"incident_id", "result"},
		"RecoveryAttestationV1":          {"recovery_id", "nonce", "recovery_snapshot_digest", "recovery_generation", "certificate_id", "proof_of_possession", "agent_build_digest", "supervisor_build_digest", "agent_guard_digest", "supervisor_guard_digest", "latch_guard_digest", "trusted_time_evidence_digest", "all_slots_stopped"},
		"RecoveryAttestationAckV1":       {"recovery_id", "attestation_digest", "result"},
	}
	for name, properties := range agentShapes {
		required := properties
		switch name {
		case "DesiredPollRequestV1":
			required = []string{"boot_id", "agent_build_digest", "capability_schema_digest", "high_water"}
		case "NodeStatePollResponseV1":
			required = nil
		case "NodeRecoveryPollResponseV1":
			required = []string{"recovery_state", "time_attestation"}
		case "SecurityFaultReportV1":
			required = []string{"operation_id", "local_fault_id", "subtype", "evidence_digest", "identity_epoch", "boot_id", "request_digest"}
		}
		assertObjectShape(t, agent, name, required, properties)
	}
	assertObjectShape(t, agent, "NodeActorHighWaterV1",
		[]string{"control_plane_authority_epoch", "node_authority_checkpoint", "server_ca_bundle", "root", "metadata"},
		[]string{"control_plane_authority_epoch", "node_authority_checkpoint", "server_ca_bundle", "root", "metadata", "desired_seen", "desired_applied", "recovery_seen"})
	artifacts := mustSchema(t, agent, "TrustConflictEvidenceRequestV1").Properties["artifacts"].Value
	if artifacts.MinItems != 1 || artifacts.MaxItems == nil || *artifacts.MaxItems != 2 || artifacts.Items == nil || artifacts.Items.Ref != "#/components/schemas/SignedConflictArtifactV1" {
		t.Errorf("trust-conflict artifacts bounds/ref = %#v, want 1..2 SignedConflictArtifactV1", artifacts)
	}
	slotFacts := mustSchema(t, agent, "NodeObservationV1").Properties["sorted_slot_facts"].Value
	if slotFacts.MinItems != 0 || slotFacts.MaxItems == nil || *slotFacts.MaxItems != 8 {
		t.Errorf("sorted_slot_facts bounds = %#v, want 0..8", slotFacts)
	}

	assertEnumRegistry(t, agent, "FaultSubtypeV1", []string{"identity_compromise", "online_signer_equivocation", "metadata_rollback", "root_rollback", "root_equivocation", "unverified_client_highwater_conflict", "client_highwater_ahead", "server_trust_bundle_conflict", "trusted_time_rollback_or_unavailable", "local_state_corruption_or_rollback", "release_or_process_integrity", "profile_binding_mismatch", "incident_overflow"})
	assertEnumRegistry(t, agent, "AdapterV1", []string{"fixture", "xray", "sing_box"})
	assertEnumRegistry(t, agent, "RecoveryReasonV1", []string{"identity_compromise", "administrative_disable", "retire", "authority_restore", "security_incident"})
	assertEnumRegistry(t, agent, "RecoveryActionV1", []string{"hold_stopped", "submit_recovery_attestation", "clear_security_latches"})
	assertEnumRegistry(t, agent, "OperationResultV1", []string{"pending", "active", "accepted", "completed", "superseded", "rejected"})
	assertEnumRegistry(t, operator, "OperatorReasonCodeV1", []string{"provision", "inventory_update", "capacity_change", "drain_maintenance", "administrative_disable", "retire", "identity_compromise", "host_remediation", "security_recovery", "authority_restore", "release_update"})
	assertEnumRegistry(t, operator, "DesiredReasonV1", []string{"initial", "operator_update", "drain", "resume", "lease_refresh", "clear_slot_quarantine", "restore_reauthorize"})
	assertEnumRegistry(t, operator, "ProtocolCapabilityV1", []string{"fixture_loopback", "xray_loopback", "sing_box_loopback"})
	assertEnumRegistry(t, operator, "AdapterV1", []string{"fixture", "xray", "sing_box"})
	assertEnumRegistry(t, operator, "OperatorStateV1", []string{"provisioning", "enabled", "draining", "disabled"})
	assertEnumRegistry(t, operator, "SecurityStateV1", []string{"normal", "quarantined"})
	assertEnumRegistry(t, operator, "IdentityStateV1", []string{"never_enrolled", "active", "recovery_pending", "recovery_limited", "revoked"})
	assertEnumRegistry(t, operator, "FaultSubtypeV1", []string{"identity_compromise", "online_signer_equivocation", "metadata_rollback", "root_rollback", "root_equivocation", "unverified_client_highwater_conflict", "client_highwater_ahead", "server_trust_bundle_conflict", "trusted_time_rollback_or_unavailable", "local_state_corruption_or_rollback", "release_or_process_integrity", "profile_binding_mismatch", "incident_overflow"})
	assertEnumRegistry(t, operator, "RestorePhaseV1", []string{"proposal", "approval"})
	assertEnumRegistry(t, operator, "OperationResultV1", []string{"pending", "active", "accepted", "completed", "superseded", "rejected"})
	assertNodeOpenAPIRegistryBindings(t, bootstrap, agent, operator)

	operatorItemShapes := map[string][]string{
		"NodePOPV1":         {"pop_code", "iso_country", "region", "operator_state", "version", "created_at", "updated_at"},
		"FailureDomainV1":   {"failure_domain_id", "domain_type", "stable_id", "version", "created_at", "updated_at"},
		"NodeEndpointV1":    {"endpoint_id", "node_id", "address", "port", "transport", "protocol_capability", "operator_state", "inventory_version", "created_at", "updated_at"},
		"NodeProcessSlotV1": {"node_id", "slot_id", "adapter", "capacity_profile_id", "capacity_profile_version", "required", "operator_state", "inventory_version", "created_at", "updated_at"},
	}
	for name, properties := range operatorItemShapes {
		assertObjectShape(t, operator, name, properties, properties)
	}
	nodeRequired := []string{"node_id", "pop_code", "operator_state", "security_state", "identity_state", "identity_epoch", "inventory_version", "security_version", "created_at", "updated_at"}
	nodeProperties := append(append([]string{}, nodeRequired...), "resume_operator_state", "pending_operator_transition", "pending_transition_signing_id", "lineage_id", "resource_envelope_version", "resource_envelope_digest", "active_desired_generation", "next_desired_generation", "active_recovery_generation", "next_recovery_generation", "active_root_publish_id", "active_root_version", "active_metadata_publish_id", "active_metadata_version", "last_authority_operation_id", "last_authority_epoch", "last_authority_sequence")
	assertObjectShape(t, operator, "NodeV1", nodeRequired, nodeProperties)
	nodeSchema := mustSchema(t, operator, "NodeV1")
	for _, property := range nodeProperties[len(nodeRequired):] {
		if !nodeSchema.Properties[property].Value.Nullable {
			t.Errorf("NodeV1.%s must be explicitly nullable", property)
		}
	}
	if len(nodeSchema.AllOf) != 5 {
		t.Errorf("NodeV1 all-or-none constraint count = %d, want 5", len(nodeSchema.AllOf))
	}

	operatorRequestShapes := map[string][]string{
		"CreateNodePOPV1":                {"command_id", "pop_code", "iso_country", "region", "operator_state", "reason_code"},
		"UpdateNodePOPV1":                {"command_id", "iso_country", "region", "operator_state", "reason_code"},
		"CreateFailureDomainV1":          {"command_id", "failure_domain_id", "domain_type", "stable_id", "reason_code"},
		"UpdateFailureDomainV1":          {"command_id", "domain_type", "stable_id", "reason_code"},
		"CreateNodeV1":                   {"command_id", "node_id", "pop_code", "operator_state", "reason_code"},
		"UpdateNodeV1":                   {"command_id", "pop_code", "operator_state", "reason_code"},
		"CreateNodeEndpointV1":           {"command_id", "endpoint_id", "address", "port", "transport", "protocol_capability", "operator_state", "reason_code"},
		"UpdateNodeEndpointV1":           {"command_id", "address", "port", "transport", "protocol_capability", "operator_state", "reason_code"},
		"CreateNodeProcessSlotV1":        {"command_id", "slot_id", "adapter", "capacity_profile_id", "capacity_profile_version", "required", "operator_state", "reason_code"},
		"UpdateNodeProcessSlotV1":        {"command_id", "adapter", "capacity_profile_id", "capacity_profile_version", "required", "operator_state", "reason_code"},
		"CreateEnrollmentGrantV1":        {"command_id", "csr_der_sha256", "reason_code"},
		"PutNodeDesiredStateV1":          {"command_id", "reason_code", "requested_valid_until", "inventory_version", "resource_envelope_version", "resource_envelope_digest", "sorted_processes"},
		"DrainNodeV1":                    {"command_id", "reason_code"},
		"DisableNodeV1":                  {"command_id", "reason_code"},
		"ReenrollNodeV1":                 {"command_id", "recovery_id", "csr_der_sha256", "reason_code"},
		"CompleteReenrollmentV1":         {"command_id", "recovery_id", "certificate_id", "recovery_attestation_digest", "host_remediation_evidence"},
		"RegisterHostSecurityIncidentV1": {"command_id", "subtype", "host_remediation_evidence"},
		"RegisterResourceEnvelopeV1":     {"command_id", "envelope_package", "host_remediation_evidence"},
		"ClearSecurityQuarantineV1":      {"command_id", "incident_id", "host_remediation_evidence", "recovery_snapshot_digest", "recovery_attestation_digest"},
		"ResumeAfterSecurityV1":          {"command_id", "recovery_id", "target_operator_state", "recovery_attestation_digest"},
		"ReauthorizeAfterRestoreV1":      {"command_id", "recovery_id", "phase", "proposal_id", "target_operator_state", "effect_digest", "host_remediation_evidence"},
		"SigningOperationV1":             {"signing_id", "status", "reserved_generation"},
		"RecoveryOperationV1":            {"recovery_id", "status"},
		"ResourceEnvelopeActivationV1":   {"version", "authority_sequence", "digest"},
		"RestoreReauthorizationV1":       {"proposal_id", "approval_id", "status", "expires_at"},
		"RecoveryEnrollmentGrantV1":      {"recovery_id", "grant_id", "csr_der_sha256", "expires_at", "status", "enrollment_grant"},
	}
	for name, properties := range operatorRequestShapes {
		required := properties
		switch name {
		case "ReauthorizeAfterRestoreV1":
			required = []string{"command_id", "recovery_id", "phase", "target_operator_state", "effect_digest", "host_remediation_evidence"}
		case "RestoreReauthorizationV1":
			required = []string{"proposal_id", "status", "expires_at"}
		case "RecoveryEnrollmentGrantV1":
			required = []string{"recovery_id", "grant_id", "csr_der_sha256", "expires_at", "status"}
		}
		assertObjectShape(t, operator, name, required, properties)
	}
	if len(mustSchema(t, operator, "ReauthorizeAfterRestoreV1").OneOf) != 2 {
		t.Error("ReauthorizeAfterRestoreV1 must have proposal and approval branches")
	}
	assertOperatorHeaderContracts(t, operator)
	assertOperatorLists(t, operator)
	for name, schemaRef := range operator.Components.Schemas {
		assertClosedRequiredObjects(t, "#/components/schemas/"+name, schemaRef, map[*openapi3.Schema]bool{})
		assertNoForbiddenOperatorProperties(t, "#/components/schemas/"+name, schemaRef, map[*openapi3.Schema]bool{})
	}
}

func TestNodeOpenAPITrustConflictEnums(t *testing.T) {
	agent := loadAndValidateNodeSpec(t, nodeagentv1.GetSwagger)
	assertEnumRegistryInOrder(t, agent, "ConflictArtifactKindV1", []string{
		"desired_state",
		"recovery_state",
		"node_state_trust_metadata",
		"node_state_root_set",
		"server_ca_trust_bundle",
	})
	assertEnumRegistryInOrder(t, agent, "TrustConflictEvidenceResultV1", []string{"accepted", "escalated"})

	kind := mustSchema(t, agent, "SignedConflictArtifactV1").Properties["kind"]
	if got := schemaReference(kind); got != "#/components/schemas/ConflictArtifactKindV1" {
		t.Errorf("SignedConflictArtifactV1.kind reference = %q", got)
	}
	result := mustSchema(t, agent, "TrustConflictEvidenceAckV1").Properties["result"]
	if got := schemaReference(result); got != "#/components/schemas/TrustConflictEvidenceResultV1" {
		t.Errorf("TrustConflictEvidenceAckV1.result reference = %q", got)
	}
}

func TestNodeOpenAPIGeneratedTrustConflictTypes(t *testing.T) {
	artifactKinds := []struct {
		value nodeagentv1.ConflictArtifactKindV1
		want  string
	}{
		{value: nodeagentv1.DesiredState, want: "desired_state"},
		{value: nodeagentv1.RecoveryState, want: "recovery_state"},
		{value: nodeagentv1.NodeStateTrustMetadata, want: "node_state_trust_metadata"},
		{value: nodeagentv1.NodeStateRootSet, want: "node_state_root_set"},
		{value: nodeagentv1.ServerCaTrustBundle, want: "server_ca_trust_bundle"},
	}
	for _, test := range artifactKinds {
		if !test.value.Valid() || string(test.value) != test.want {
			t.Errorf("generated conflict artifact kind = %q, valid=%t; want %q", test.value, test.value.Valid(), test.want)
		}
	}
	for _, invalid := range []nodeagentv1.ConflictArtifactKindV1{
		"server_ca_bundle",
		"node_state_root",
		"node_state_metadata",
		"node_desired_state",
		"node_recovery_state",
		"Desired_State",
	} {
		if invalid.Valid() {
			t.Errorf("obsolete or case-variant conflict artifact kind %q is valid", invalid)
		}
	}

	results := []struct {
		value nodeagentv1.TrustConflictEvidenceResultV1
		want  string
	}{
		{value: nodeagentv1.TrustConflictEvidenceResultV1Accepted, want: "accepted"},
		{value: nodeagentv1.TrustConflictEvidenceResultV1Escalated, want: "escalated"},
	}
	for _, test := range results {
		if !test.value.Valid() || string(test.value) != test.want {
			t.Errorf("generated trust-conflict result = %q, valid=%t; want %q", test.value, test.value.Valid(), test.want)
		}
	}
	for _, invalid := range []nodeagentv1.TrustConflictEvidenceResultV1{
		"pending", "active", "completed", "superseded", "rejected", "Accepted",
	} {
		if invalid.Valid() {
			t.Errorf("non-route trust-conflict result %q is valid", invalid)
		}
	}

	responses := []struct {
		name  string
		value any
	}{
		{name: "desired", value: nodeagentv1.PollNodeDesiredStateResponse{}},
		{name: "recovery", value: nodeagentv1.PollNodeRecoveryStateResponse{}},
	}
	for _, response := range responses {
		typeOfResponse := reflect.TypeOf(response.value)
		if _, ok := typeOfResponse.FieldByName("JSON409"); ok {
			t.Errorf("%s poll generated response exposes JSON409", response.name)
		}
		if _, ok := typeOfResponse.FieldByName("Headers409"); !ok {
			t.Errorf("%s poll generated response lacks typed Headers409", response.name)
		}
	}
}

func TestNodeOpenAPIPollConflictContract(t *testing.T) {
	agent := loadAndValidateNodeSpec(t, nodeagentv1.GetSwagger)
	const headerName = "Talenro-Trust-Conflict-Incident-ID"

	component := agent.Components.Responses["NodePollConflict"]
	if component == nil || component.Value == nil {
		t.Fatal("missing shared NodePollConflict response")
	}
	assertNodePollConflictResponse(t, component.Value, headerName)

	for _, path := range []string{"/v1/node-agent/desired-state:poll", "/v1/node-agent/recovery-state:poll"} {
		operation := agent.Paths.Find(path).Post
		conflict := operation.Responses.Value("409")
		if conflict == nil || conflict.Ref != "#/components/responses/NodePollConflict" {
			t.Fatalf("%s 409 response = %#v, want shared NodePollConflict", path, conflict)
		}
		assertNodePollConflictResponse(t, conflict.Value, headerName)
		if got := effectiveParameterNames(agent.Paths.Find(path), operation, "header"); len(got) != 0 {
			t.Errorf("%s accepts request headers %v, want none", path, got)
		}
		for _, status := range []string{"200", "204"} {
			response := operation.Responses.Value(status)
			if response != nil && response.Value != nil && response.Value.Headers[headerName] != nil {
				t.Errorf("%s %s unexpectedly exposes %s", path, status, headerName)
			}
		}
	}

	for _, expectation := range nodeSpecExpectations()[1].operations {
		if expectation.operationID == "pollNodeDesiredState" || expectation.operationID == "pollNodeRecoveryState" {
			continue
		}
		item := agent.Paths.Find(expectation.path)
		operation := operationForMethod(item, expectation.method)
		conflict := operation.Responses.Value("409")
		if conflict != nil && conflict.Value != nil && conflict.Value.Headers[headerName] != nil {
			t.Errorf("%s 409 unexpectedly exposes %s", expectation.operationID, headerName)
		}
	}
}

func TestNodeOpenAPISourceAndEmbeddedInventoriesMatch(t *testing.T) {
	tests := []struct {
		name string
		file string
		get  func() (*openapi3.T, error)
	}{
		{name: "bootstrap", file: "node-bootstrap-api.v1.yaml", get: nodebootstrapv1.GetSwagger},
		{name: "agent", file: "node-agent-api.v1.yaml", get: nodeagentv1.GetSwagger},
		{name: "operator", file: "node-operator-api.v1.yaml", get: nodeoperatorv1.GetSwagger},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source := loadNodeOpenAPISource(t, test.file)
			embedded := loadAndValidateNodeSpec(t, test.get)
			assertExactStringSet(t, "source/embedded schemas", schemaNames(source), schemaNames(embedded))
			assertExactStringSet(t, "source/embedded responses", mapKeys(source.Components.Responses), mapKeys(embedded.Components.Responses))
			assertExactStringSet(t, "source/embedded paths", mapKeys(source.Paths.Map()), mapKeys(embedded.Paths.Map()))
			assertExactStringSet(t, "source/embedded operations", operationInventory(source), operationInventory(embedded))
		})
	}
}

func TestNodeOpenAPIGeneratedClientAndServerSurfaces(t *testing.T) {
	clients := []struct {
		name string
		new  func(string) error
	}{
		{name: "bootstrap", new: func(server string) error { _, err := nodebootstrapv1.NewClientWithResponses(server); return err }},
		{name: "agent", new: func(server string) error { _, err := nodeagentv1.NewClientWithResponses(server); return err }},
		{name: "operator", new: func(server string) error { _, err := nodeoperatorv1.NewClientWithResponses(server); return err }},
	}
	for _, test := range clients {
		t.Run(test.name, func(t *testing.T) {
			if err := test.new("https://node-control.invalid"); err != nil {
				t.Fatal(err)
			}
		})
	}
	_ = nodebootstrapv1.ClaimNodeEnrollmentResponse{}
	_ = nodeagentv1.PollNodeDesiredStateResponse{}
	_ = nodeoperatorv1.CreateNodeResponse{}
}

func TestNodeOpenAPIEffectiveParameterAllowlistsIncludePathDefinitions(t *testing.T) {
	spec := loadAndValidateNodeSpec(t, nodeoperatorv1.GetSwagger)
	item := spec.Paths.Find("/v1/operator/nodes/{node_id}/endpoints")
	item.Parameters = append(item.Parameters,
		&openapi3.ParameterRef{Value: &openapi3.Parameter{Name: "X-Forbidden", In: "header"}},
		&openapi3.ParameterRef{Value: &openapi3.Parameter{Name: "include", In: "query"}},
	)

	if got := effectiveParameterNames(item, item.Get, "header"); !sameStringSet(got, []string{"X-Forbidden"}) {
		t.Errorf("effective headers = %v, want path-level X-Forbidden", got)
	}
	if got := effectiveParameterNames(item, item.Get, "query"); !sameStringSet(got, []string{"page_size", "cursor", "transport", "protocol_capability", "operator_state", "include"}) {
		t.Errorf("effective query parameters = %v, want operation filters plus path-level include", got)
	}
}

func nodeSpecExpectations() []nodeSpecExpectation {
	return []nodeSpecExpectation{
		{
			name: "bootstrap", title: "Talenro Node Bootstrap API", auth: "bootstrap_server_tls", get: nodebootstrapv1.GetSwagger,
			operations: []operationExpectation{
				{method: http.MethodPost, path: "/v1/node-enrollments/claim", operationID: "claimNodeEnrollment", requestSchema: "ClaimNodeEnrollmentRequestV1", successSchemas: map[string]string{"200": "CertificateAuthorizationV1"}, wireBytes: 65536, decodedBytes: 65536},
			},
			schemaNames: []string{"Base64URL32", "Base64URL64", "Base64URLDER", "CSRDER", "CanonicalUUID", "CertificateAuthorizationReceiptV1", "CertificateAuthorizationV1", "ClaimNodeEnrollmentRequestV1", "DigestHex32", "NodeTimeAttestationV1", "PositiveInt64", "PublicErrorV1", "SignedNodeStateTrustMetadataV1"},
		},
		{
			name: "agent", title: "Talenro Node Agent API", auth: "node_mtls", get: nodeagentv1.GetSwagger,
			operations: []operationExpectation{
				{method: http.MethodPost, path: "/v1/node-agent/certificates/rotate", operationID: "rotateNodeCertificate", requestSchema: "RotateNodeCertificateRequestV1", successSchemas: map[string]string{"200": "CertificateAuthorizationV1"}, wireBytes: 65536, decodedBytes: 65536},
				{method: http.MethodPost, path: "/v1/node-agent/desired-state:poll", operationID: "pollNodeDesiredState", requestSchema: "DesiredPollRequestV1", successSchemas: map[string]string{"200": "NodeStatePollResponseV1", "204": ""}, wireBytes: 65536, decodedBytes: 65536},
				{method: http.MethodPost, path: "/v1/node-agent/recovery-state:poll", operationID: "pollNodeRecoveryState", requestSchema: "RecoveryPollRequestV1", successSchemas: map[string]string{"200": "NodeRecoveryPollResponseV1", "204": ""}, wireBytes: 65536, decodedBytes: 65536},
				{method: http.MethodPost, path: "/v1/node-agent/observations", operationID: "createNodeObservation", requestSchema: "NodeObservationV1", successSchemas: map[string]string{"200": "ObservationAckV1"}, wireBytes: 65536, decodedBytes: 65536},
				{method: http.MethodPost, path: "/v1/node-agent/security-faults", operationID: "createNodeSecurityFault", requestSchema: "SecurityFaultReportV1", successSchemas: map[string]string{"200": "SecurityFaultReceiptV1"}, wireBytes: 65536, decodedBytes: 65536},
				{method: http.MethodPost, path: "/v1/node-agent/trust-conflict-evidence", operationID: "createNodeTrustConflictEvidence", requestSchema: "TrustConflictEvidenceRequestV1", successSchemas: map[string]string{"200": "TrustConflictEvidenceAckV1"}, wireBytes: 1048576, decodedBytes: 1048576},
				{method: http.MethodPost, path: "/v1/node-agent/recovery-attestations", operationID: "createNodeRecoveryAttestation", requestSchema: "RecoveryAttestationV1", successSchemas: map[string]string{"200": "RecoveryAttestationAckV1"}, wireBytes: 65536, decodedBytes: 65536},
			},
			schemaNames: []string{"AdapterV1", "Base64URL32", "Base64URL64", "Base64URLDER", "CSRDER", "CanonicalUUID", "CertificateAuthorizationReceiptV1", "CertificateAuthorizationV1", "ConflictArtifactKindV1", "DesiredPollRequestV1", "DigestHex32", "FaultSubtypeV1", "NodeActorHighWaterV1", "NodeHighWaterV1", "NodeObservationV1", "NodeRecoveryPollResponseV1", "NodeSlotFactV1", "NodeStatePollResponseV1", "NodeTimeAttestationV1", "NonNegativeInt64", "ObservationAckV1", "OperationResultV1", "PositiveInt64", "PublicErrorV1", "RecoveryActionV1", "RecoveryAttestationAckV1", "RecoveryAttestationV1", "RecoveryPollRequestV1", "RecoveryReasonV1", "RotateNodeCertificateRequestV1", "SecurityFaultReceiptV1", "SecurityFaultReportV1", "SignedCanonicalArtifactV1", "SignedConflictArtifactV1", "SignedRecoveryStateV1", "SupervisorFaultV1", "TrustConflictEvidenceAckV1", "TrustConflictEvidenceRequestV1", "TrustConflictEvidenceResultV1"},
		},
		operatorExpectation(),
	}
}

func operatorExpectation() nodeSpecExpectation {
	const nodePath = "/v1/operator/nodes/{node_id}"
	bothHeaders := []string{"Idempotency-Key", "If-Match"}
	operations := []operationExpectation{
		{method: http.MethodPost, path: "/v1/operator/pops", operationID: "createNodePOP", requestSchema: "CreateNodePOPV1", successSchemas: map[string]string{"201": "NodePOPV1"}, requestHeaders: []string{"Idempotency-Key"}, responseHasETag: true, wireBytes: 65536, decodedBytes: 65536},
		{method: http.MethodGet, path: "/v1/operator/pops", operationID: "listNodePOPs", successSchemas: map[string]string{"200": "NodePOPListV1"}},
		{method: http.MethodGet, path: "/v1/operator/pops/{pop_code}", operationID: "getNodePOP", successSchemas: map[string]string{"200": "NodePOPV1"}, responseHasETag: true},
		{method: http.MethodPut, path: "/v1/operator/pops/{pop_code}", operationID: "updateNodePOP", requestSchema: "UpdateNodePOPV1", successSchemas: map[string]string{"200": "NodePOPV1"}, requestHeaders: bothHeaders, responseHasETag: true, wireBytes: 65536, decodedBytes: 65536},
		{method: http.MethodPost, path: "/v1/operator/failure-domains", operationID: "createNodeFailureDomain", requestSchema: "CreateFailureDomainV1", successSchemas: map[string]string{"201": "FailureDomainV1"}, requestHeaders: []string{"Idempotency-Key"}, responseHasETag: true, wireBytes: 65536, decodedBytes: 65536},
		{method: http.MethodGet, path: "/v1/operator/failure-domains", operationID: "listNodeFailureDomains", successSchemas: map[string]string{"200": "FailureDomainListV1"}},
		{method: http.MethodGet, path: "/v1/operator/failure-domains/{failure_domain_id}", operationID: "getNodeFailureDomain", successSchemas: map[string]string{"200": "FailureDomainV1"}, responseHasETag: true},
		{method: http.MethodPut, path: "/v1/operator/failure-domains/{failure_domain_id}", operationID: "updateNodeFailureDomain", requestSchema: "UpdateFailureDomainV1", successSchemas: map[string]string{"200": "FailureDomainV1"}, requestHeaders: bothHeaders, responseHasETag: true, wireBytes: 65536, decodedBytes: 65536},
		{method: http.MethodPost, path: "/v1/operator/nodes", operationID: "createNode", requestSchema: "CreateNodeV1", successSchemas: map[string]string{"201": "NodeV1"}, requestHeaders: []string{"Idempotency-Key"}, responseHasETag: true, wireBytes: 65536, decodedBytes: 65536},
		{method: http.MethodGet, path: "/v1/operator/nodes", operationID: "listNodes", successSchemas: map[string]string{"200": "NodeListV1"}},
		{method: http.MethodGet, path: nodePath, operationID: "getNode", successSchemas: map[string]string{"200": "NodeV1"}, responseHasETag: true},
		{method: http.MethodPut, path: nodePath, operationID: "updateNode", requestSchema: "UpdateNodeV1", successSchemas: map[string]string{"200": "NodeV1"}, requestHeaders: bothHeaders, responseHasETag: true, wireBytes: 65536, decodedBytes: 65536},
		{method: http.MethodPost, path: nodePath + "/endpoints", operationID: "createNodeEndpoint", requestSchema: "CreateNodeEndpointV1", successSchemas: map[string]string{"201": "NodeEndpointV1"}, requestHeaders: bothHeaders, responseHasETag: true, wireBytes: 65536, decodedBytes: 65536},
		{method: http.MethodGet, path: nodePath + "/endpoints", operationID: "listNodeEndpoints", successSchemas: map[string]string{"200": "NodeEndpointListV1"}, responseHasETag: true},
		{method: http.MethodGet, path: nodePath + "/endpoints/{endpoint_id}", operationID: "getNodeEndpoint", successSchemas: map[string]string{"200": "NodeEndpointV1"}, responseHasETag: true},
		{method: http.MethodPut, path: nodePath + "/endpoints/{endpoint_id}", operationID: "updateNodeEndpoint", requestSchema: "UpdateNodeEndpointV1", successSchemas: map[string]string{"200": "NodeEndpointV1"}, requestHeaders: bothHeaders, responseHasETag: true, wireBytes: 65536, decodedBytes: 65536},
		{method: http.MethodPost, path: nodePath + "/process-slots", operationID: "createNodeProcessSlot", requestSchema: "CreateNodeProcessSlotV1", successSchemas: map[string]string{"201": "NodeProcessSlotV1"}, requestHeaders: bothHeaders, responseHasETag: true, wireBytes: 65536, decodedBytes: 65536},
		{method: http.MethodGet, path: nodePath + "/process-slots", operationID: "listNodeProcessSlots", successSchemas: map[string]string{"200": "NodeProcessSlotListV1"}, responseHasETag: true},
		{method: http.MethodGet, path: nodePath + "/process-slots/{slot_id}", operationID: "getNodeProcessSlot", successSchemas: map[string]string{"200": "NodeProcessSlotV1"}, responseHasETag: true},
		{method: http.MethodPut, path: nodePath + "/process-slots/{slot_id}", operationID: "updateNodeProcessSlot", requestSchema: "UpdateNodeProcessSlotV1", successSchemas: map[string]string{"200": "NodeProcessSlotV1"}, requestHeaders: bothHeaders, responseHasETag: true, wireBytes: 65536, decodedBytes: 65536},
		{method: http.MethodPost, path: nodePath + "/enrollment-grants", operationID: "createNodeEnrollmentGrant", requestSchema: "CreateEnrollmentGrantV1", successSchemas: map[string]string{"201": "EnrollmentGrantV1"}, requestHeaders: bothHeaders, responseHasETag: true, wireBytes: 65536, decodedBytes: 65536},
		{method: http.MethodPut, path: nodePath + "/desired-state", operationID: "putNodeDesiredState", requestSchema: "PutNodeDesiredStateV1", successSchemas: map[string]string{"202": "SigningOperationV1"}, requestHeaders: bothHeaders, wireBytes: 65536, decodedBytes: 65536},
		{method: http.MethodPost, path: nodePath + "/actions/drain", operationID: "drainNode", requestSchema: "DrainNodeV1", successSchemas: map[string]string{"202": "SigningOperationV1"}, requestHeaders: bothHeaders, wireBytes: 65536, decodedBytes: 65536},
		{method: http.MethodPost, path: nodePath + "/actions/disable", operationID: "disableNode", requestSchema: "DisableNodeV1", successSchemas: map[string]string{"202": "RecoveryOperationV1"}, requestHeaders: bothHeaders, wireBytes: 65536, decodedBytes: 65536},
		{method: http.MethodPost, path: nodePath + "/actions/reenroll", operationID: "reenrollNode", requestSchema: "ReenrollNodeV1", successSchemas: map[string]string{"201": "RecoveryEnrollmentGrantV1"}, requestHeaders: bothHeaders, wireBytes: 65536, decodedBytes: 65536},
		{method: http.MethodPost, path: nodePath + "/actions/complete-reenrollment", operationID: "completeNodeReenrollment", requestSchema: "CompleteReenrollmentV1", successSchemas: map[string]string{"202": "RecoveryOperationV1"}, requestHeaders: bothHeaders, wireBytes: 65536, decodedBytes: 65536},
		{method: http.MethodPost, path: nodePath + "/actions/register-host-security-incident", operationID: "registerNodeHostSecurityIncident", requestSchema: "RegisterHostSecurityIncidentV1", successSchemas: map[string]string{"202": "RecoveryOperationV1"}, requestHeaders: bothHeaders, wireBytes: 65536, decodedBytes: 65536},
		{method: http.MethodPost, path: nodePath + "/actions/register-resource-envelope", operationID: "registerNodeResourceEnvelope", requestSchema: "RegisterResourceEnvelopeV1", successSchemas: map[string]string{"200": "ResourceEnvelopeActivationV1"}, requestHeaders: bothHeaders, wireBytes: 65536, decodedBytes: 65536},
		{method: http.MethodPost, path: nodePath + "/actions/clear-security-quarantine", operationID: "clearNodeSecurityQuarantine", requestSchema: "ClearSecurityQuarantineV1", successSchemas: map[string]string{"202": "RecoveryOperationV1"}, requestHeaders: bothHeaders, wireBytes: 65536, decodedBytes: 65536},
		{method: http.MethodPost, path: nodePath + "/actions/resume-after-security", operationID: "resumeNodeAfterSecurity", requestSchema: "ResumeAfterSecurityV1", successSchemas: map[string]string{"202": "SigningOperationV1"}, requestHeaders: bothHeaders, wireBytes: 65536, decodedBytes: 65536},
		{method: http.MethodPost, path: nodePath + "/actions/reauthorize-after-restore", operationID: "reauthorizeNodeAfterRestore", requestSchema: "ReauthorizeAfterRestoreV1", successSchemas: map[string]string{"202": "RestoreReauthorizationV1"}, requestHeaders: bothHeaders, wireBytes: 65536, decodedBytes: 65536},
	}
	return nodeSpecExpectation{
		name: "operator", title: "Talenro Node Operator API", auth: "operator_mtls", get: nodeoperatorv1.GetSwagger, operations: operations,
		schemaNames: []string{"AdapterV1", "Base64URL32", "CanonicalUUID", "CapacityProfileIDV1", "ClearSecurityQuarantineV1", "CompleteReenrollmentV1", "CreateEnrollmentGrantV1", "CreateFailureDomainV1", "CreateNodeEndpointV1", "CreateNodePOPV1", "CreateNodeProcessSlotV1", "CreateNodeV1", "DesiredProcessV1", "DesiredReasonV1", "DigestHex32", "DisableNodeV1", "DisableReasonV1", "DrainNodeV1", "EnrollmentGrantV1", "FailureDomainListV1", "FailureDomainTypeV1", "FailureDomainV1", "FaultSubtypeV1", "HostRemediationEvidenceV1", "IdentityStateV1", "NodeEndpointListV1", "NodeEndpointV1", "NodeListV1", "NodePOPListV1", "NodePOPV1", "NodeProcessSlotListV1", "NodeProcessSlotV1", "NodeV1", "NonNegativeInt64", "OperationResultV1", "OperatorListCursorV1", "OperatorReasonCodeV1", "OperatorStateV1", "POPCodeV1", "PositiveInt64", "ProcessLifecycleV1", "ProtocolCapabilityV1", "PublicErrorV1", "PutNodeDesiredStateV1", "ReauthorizeAfterRestoreV1", "RecoveryEnrollmentGrantV1", "RecoveryOperationV1", "ReenrollNodeV1", "RegionV1", "RegisterHostSecurityIncidentV1", "RegisterResourceEnvelopeV1", "ResourceEnvelopeActivationV1", "ResourceEnvelopePackageV1", "RestorePhaseV1", "RestoreReauthorizationV1", "RestoreTargetOperatorStateV1", "ResumeAfterSecurityV1", "ResumeOperatorStateV1", "SecurityStateV1", "SigningOperationV1", "SlotIDV1", "StableIDV1", "TransportV1", "UpdateFailureDomainV1", "UpdateNodeEndpointV1", "UpdateNodePOPV1", "UpdateNodeProcessSlotV1", "UpdateNodeV1"},
	}
}

func assertRuntimeTLSContract(t *testing.T, auth string, expectation operationExpectation, item *openapi3.PathItem, operation *openapi3.Operation) {
	t.Helper()
	if got := fmt.Sprint(operation.Extensions["x-talenro-listener-auth"]); got != auth {
		t.Errorf("%s auth extension = %q, want %q", expectation.operationID, got, auth)
	}
	if operation.Security == nil || len(*operation.Security) != 0 {
		t.Errorf("%s security = %#v, want explicit empty requirements", expectation.operationID, operation.Security)
	}
	if got := effectiveParameterNames(item, operation, "header"); !sameStringSet(got, expectation.requestHeaders) {
		t.Errorf("%s request headers = %v, want %v", expectation.operationID, got, expectation.requestHeaders)
	}
}

func assertRequestContract(t *testing.T, expectation operationExpectation, operation *openapi3.Operation) {
	t.Helper()
	if expectation.requestSchema == "" {
		if operation.RequestBody != nil {
			t.Errorf("%s unexpectedly has a request body", expectation.operationID)
		}
		for _, extension := range []string{"x-talenro-content-encoding", "x-talenro-max-wire-bytes", "x-talenro-max-decoded-bytes"} {
			if _, exists := operation.Extensions[extension]; exists {
				t.Errorf("%s unexpectedly has %s without a request body", expectation.operationID, extension)
			}
		}
		return
	}
	if operation.RequestBody == nil || operation.RequestBody.Value == nil || !operation.RequestBody.Value.Required {
		t.Fatalf("%s must have a required request body", expectation.operationID)
	}
	media := operation.RequestBody.Value.Content["application/json"]
	if media == nil || media.Schema == nil || media.Schema.Ref != "#/components/schemas/"+expectation.requestSchema {
		t.Errorf("%s request schema = %#v, want %s", expectation.operationID, media, expectation.requestSchema)
	}
	if got := fmt.Sprint(operation.Extensions["x-talenro-content-encoding"]); got != "identity" {
		t.Errorf("%s content encoding = %q, want identity", expectation.operationID, got)
	}
	if got, ok := integerExtension(operation.Extensions["x-talenro-max-wire-bytes"]); !ok || got != int64(expectation.wireBytes) {
		t.Errorf("%s wire cap = %v, want %d", expectation.operationID, operation.Extensions["x-talenro-max-wire-bytes"], expectation.wireBytes)
	}
	if got, ok := integerExtension(operation.Extensions["x-talenro-max-decoded-bytes"]); !ok || got != int64(expectation.decodedBytes) {
		t.Errorf("%s decoded cap = %v, want %d", expectation.operationID, operation.Extensions["x-talenro-max-decoded-bytes"], expectation.decodedBytes)
	}
}

func assertSuccessContract(t *testing.T, expectation operationExpectation, operation *openapi3.Operation) {
	t.Helper()
	gotStatuses := make([]string, 0)
	for status := 100; status < 400; status++ {
		statusText := strconv.Itoa(status)
		if operation.Responses.Value(statusText) != nil {
			gotStatuses = append(gotStatuses, statusText)
		}
	}
	wantStatuses := make([]string, 0, len(expectation.successSchemas))
	for status := range expectation.successSchemas {
		wantStatuses = append(wantStatuses, status)
	}
	if !sameStringSet(gotStatuses, wantStatuses) {
		t.Errorf("%s success statuses = %v, want %v", expectation.operationID, gotStatuses, wantStatuses)
	}
	for status, schemaName := range expectation.successSchemas {
		responseRef := operation.Responses.Value(status)
		if responseRef == nil || responseRef.Value == nil {
			t.Errorf("%s missing success response %s", expectation.operationID, status)
			continue
		}
		if schemaName == "" {
			if len(responseRef.Value.Content) != 0 {
				t.Errorf("%s %s must have an empty response", expectation.operationID, status)
			}
			continue
		}
		media := responseRef.Value.Content["application/json"]
		if media == nil || media.Schema == nil || media.Schema.Ref != "#/components/schemas/"+schemaName {
			t.Errorf("%s %s schema = %#v, want %s", expectation.operationID, status, media, schemaName)
		}
		hasETag := responseRef.Value.Headers["ETag"] != nil
		if hasETag != expectation.responseHasETag {
			t.Errorf("%s %s ETag present = %t, want %t", expectation.operationID, status, hasETag, expectation.responseHasETag)
		}
	}
}

func assertOperatorLists(t *testing.T, spec *openapi3.T) {
	t.Helper()
	tests := []struct {
		operationID string
		path        string
		listSchema  string
		itemSchema  string
		queryNames  []string
	}{
		{operationID: "listNodePOPs", path: "/v1/operator/pops", listSchema: "NodePOPListV1", itemSchema: "NodePOPV1", queryNames: []string{"page_size", "cursor", "iso_country", "region", "operator_state"}},
		{operationID: "listNodeFailureDomains", path: "/v1/operator/failure-domains", listSchema: "FailureDomainListV1", itemSchema: "FailureDomainV1", queryNames: []string{"page_size", "cursor", "domain_type", "stable_id"}},
		{operationID: "listNodes", path: "/v1/operator/nodes", listSchema: "NodeListV1", itemSchema: "NodeV1", queryNames: []string{"page_size", "cursor", "pop_code", "operator_state", "security_state", "identity_state"}},
		{operationID: "listNodeEndpoints", path: "/v1/operator/nodes/{node_id}/endpoints", listSchema: "NodeEndpointListV1", itemSchema: "NodeEndpointV1", queryNames: []string{"page_size", "cursor", "transport", "protocol_capability", "operator_state"}},
		{operationID: "listNodeProcessSlots", path: "/v1/operator/nodes/{node_id}/process-slots", listSchema: "NodeProcessSlotListV1", itemSchema: "NodeProcessSlotV1", queryNames: []string{"page_size", "cursor", "adapter", "required", "operator_state"}},
	}
	for _, test := range tests {
		item := spec.Paths.Find(test.path)
		op := item.Get
		if op.OperationID != test.operationID {
			t.Fatalf("%s operation ID = %q", test.path, op.OperationID)
		}
		responseSchema := op.Responses.Value("200").Value.Content["application/json"].Schema
		if responseSchema.Ref != "#/components/schemas/"+test.listSchema {
			t.Errorf("%s response schema = %q, want %s", test.operationID, responseSchema.Ref, test.listSchema)
		}
		listSchema := mustSchema(t, spec, test.listSchema)
		items := listSchema.Properties["items"]
		if items == nil || items.Value == nil || items.Value.Items == nil || items.Value.Items.Ref != "#/components/schemas/"+test.itemSchema || items.Value.MinItems != 0 || items.Value.MaxItems == nil || *items.Value.MaxItems != 200 {
			t.Errorf("%s items = %#v, want exact %s ref with 0..200", test.listSchema, items, test.itemSchema)
		}
		if got := effectiveParameterNames(item, op, "query"); !sameStringSet(got, test.queryNames) {
			t.Errorf("%s query parameters = %v, want %v", test.operationID, got, test.queryNames)
		}
		pageSize := findEffectiveParameter(item, op, "query", "page_size")
		if pageSize == nil || pageSize.Schema == nil || pageSize.Schema.Value == nil || pageSize.Schema.Value.Min == nil || *pageSize.Schema.Value.Min != 1 || pageSize.Schema.Value.Max == nil || *pageSize.Schema.Value.Max != 200 || fmt.Sprint(pageSize.Schema.Value.Default) != "50" {
			t.Errorf("%s page_size bounds/default = %#v, want 1..200 default 50", test.operationID, pageSize)
		}
	}
}

func assertOperatorHeaderContracts(t *testing.T, spec *openapi3.T) {
	t.Helper()
	idempotency := spec.Components.Parameters["IdempotencyKey"]
	if idempotency == nil || idempotency.Value == nil || idempotency.Value.Schema == nil || idempotency.Value.Schema.Value == nil {
		t.Fatal("missing IdempotencyKey parameter")
	}
	idempotencySchema := idempotency.Value.Schema.Value
	if idempotency.Value.Name != "Idempotency-Key" || idempotency.Value.In != "header" || !idempotency.Value.Required || idempotencySchema.MinLength != 22 || idempotencySchema.MaxLength == nil || *idempotencySchema.MaxLength != 86 || idempotencySchema.Pattern != `^[A-Za-z0-9_-]+$` {
		t.Errorf("IdempotencyKey contract = %#v", idempotency.Value)
	}
	ifMatch := spec.Components.Parameters["IfMatch"]
	if ifMatch == nil || ifMatch.Value == nil || ifMatch.Value.Schema == nil || ifMatch.Value.Schema.Value == nil {
		t.Fatal("missing IfMatch parameter")
	}
	if ifMatch.Value.Name != "If-Match" || ifMatch.Value.In != "header" || !ifMatch.Value.Required || ifMatch.Value.Schema.Value.Pattern != `^"[1-9][0-9]{0,18}"$` {
		t.Errorf("IfMatch contract = %#v", ifMatch.Value)
	}
	cursor := mustSchema(t, spec, "OperatorListCursorV1")
	if cursor.MinLength != 16 || cursor.MaxLength == nil || *cursor.MaxLength != 2048 || cursor.Pattern != `^[A-Za-z0-9_-]+$` {
		t.Errorf("OperatorListCursorV1 contract = %#v", cursor)
	}
	strongETag := spec.Components.Headers["StrongETag"]
	if strongETag == nil || strongETag.Value == nil || strongETag.Value.Schema == nil || strongETag.Value.Schema.Value == nil || strongETag.Value.Schema.Value.Pattern != `^"[1-9][0-9]{0,18}"$` {
		t.Errorf("StrongETag contract = %#v", strongETag)
	}
}

func loadAndValidateNodeSpec(t *testing.T, get func() (*openapi3.T, error)) *openapi3.T {
	t.Helper()
	spec, err := get()
	if err != nil {
		t.Fatal(err)
	}
	if err := spec.Validate(context.Background()); err != nil {
		t.Fatal(err)
	}
	return spec
}

func loadNodeOpenAPISource(t *testing.T, file string) *openapi3.T {
	t.Helper()
	path := filepath.Join("..", "..", "..", "api", "openapi", file)
	spec, err := openapi3.NewLoader().LoadFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := spec.Validate(context.Background()); err != nil {
		t.Fatal(err)
	}
	return spec
}

func assertNodePollConflictResponse(t *testing.T, response *openapi3.Response, headerName string) {
	t.Helper()
	if response == nil {
		t.Fatal("missing NodePollConflict response value")
	}
	if len(response.Content) != 0 {
		t.Errorf("NodePollConflict content = %v, want no body", mapKeys(response.Content))
	}
	assertExactStringSet(t, "NodePollConflict headers", mapKeys(response.Headers), []string{headerName})
	header := response.Headers[headerName]
	if header == nil || header.Value == nil || header.Value.Schema == nil ||
		schemaReference(header.Value.Schema) != "#/components/schemas/CanonicalUUID" {
		t.Errorf("NodePollConflict %s = %#v, want optional CanonicalUUID", headerName, header)
		return
	}
	if header.Value.Required {
		t.Errorf("NodePollConflict %s must remain optional", headerName)
	}
}

func operationInventory(spec *openapi3.T) []string {
	operations := make([]string, 0)
	for path, item := range spec.Paths.Map() {
		for _, candidate := range []struct {
			method    string
			operation *openapi3.Operation
		}{
			{method: http.MethodConnect, operation: item.Connect},
			{method: http.MethodDelete, operation: item.Delete},
			{method: http.MethodGet, operation: item.Get},
			{method: http.MethodHead, operation: item.Head},
			{method: http.MethodOptions, operation: item.Options},
			{method: http.MethodPatch, operation: item.Patch},
			{method: http.MethodPost, operation: item.Post},
			{method: http.MethodPut, operation: item.Put},
			{method: http.MethodTrace, operation: item.Trace},
		} {
			if candidate.operation != nil {
				operations = append(operations, candidate.method+" "+path+" "+candidate.operation.OperationID)
			}
		}
	}
	return operations
}

func operationForMethod(item *openapi3.PathItem, method string) *openapi3.Operation {
	switch method {
	case http.MethodGet:
		return item.Get
	case http.MethodPost:
		return item.Post
	case http.MethodPut:
		return item.Put
	default:
		return nil
	}
}

func pathItemOperationCount(item *openapi3.PathItem) int {
	count := 0
	for _, operation := range []*openapi3.Operation{item.Connect, item.Delete, item.Get, item.Head, item.Options, item.Patch, item.Post, item.Put, item.Trace} {
		if operation != nil {
			count++
		}
	}
	return count
}

func effectiveParameterNames(item *openapi3.PathItem, operation *openapi3.Operation, location string) []string {
	names := make([]string, 0)
	for _, parameter := range effectiveParameters(item, operation) {
		if parameter.In == location {
			names = append(names, parameter.Name)
		}
	}
	return names
}

func findEffectiveParameter(item *openapi3.PathItem, operation *openapi3.Operation, location, name string) *openapi3.Parameter {
	return effectiveParameters(item, operation)[location+"\x00"+name]
}

func effectiveParameters(item *openapi3.PathItem, operation *openapi3.Operation) map[string]*openapi3.Parameter {
	parameters := make(map[string]*openapi3.Parameter, len(item.Parameters)+len(operation.Parameters))
	add := func(refs openapi3.Parameters) {
		for _, ref := range refs {
			if ref.Value == nil {
				continue
			}
			parameters[ref.Value.In+"\x00"+ref.Value.Name] = ref.Value
		}
	}
	add(item.Parameters)
	add(operation.Parameters)
	return parameters
}

func schemaNames(spec *openapi3.T) []string {
	names := make([]string, 0, len(spec.Components.Schemas))
	for name := range spec.Components.Schemas {
		names = append(names, name)
	}
	return names
}

func mustSchema(t *testing.T, spec *openapi3.T, name string) *openapi3.Schema {
	t.Helper()
	ref := spec.Components.Schemas[name]
	if ref == nil || ref.Value == nil {
		t.Fatalf("missing schema %s", name)
	}
	return ref.Value
}

func assertObjectShape(t *testing.T, spec *openapi3.T, name string, required, properties []string) {
	t.Helper()
	schema := mustSchema(t, spec, name)
	if schema.Type == nil || !schema.Type.Is(openapi3.TypeObject) {
		t.Errorf("%s type = %v, want object", name, schema.Type)
	}
	if schema.AdditionalProperties.Has == nil || *schema.AdditionalProperties.Has {
		t.Errorf("%s must set additionalProperties: false", name)
	}
	assertExactStringSet(t, name+" required", schema.Required, required)
	gotProperties := make([]string, 0, len(schema.Properties))
	for property := range schema.Properties {
		gotProperties = append(gotProperties, property)
	}
	assertExactStringSet(t, name+" properties", gotProperties, properties)
}

func assertPublicErrorBounds(t *testing.T, spec *openapi3.T) {
	t.Helper()
	publicError := mustSchema(t, spec, "PublicErrorV1")
	requestID := publicError.Properties["request_id"].Value
	if requestID.MinLength != 16 || requestID.MaxLength == nil || *requestID.MaxLength != 64 || requestID.Pattern != `^[A-Za-z0-9_-]+$` {
		t.Errorf("PublicErrorV1.request_id bounds/pattern = %#v, want length 16..64 and base64url characters", requestID)
	}
	retryAfter := publicError.Properties["retry_after_ms"].Value
	if retryAfter.Min == nil || *retryAfter.Min != 1 || retryAfter.Max == nil || *retryAfter.Max != 10000 {
		t.Errorf("PublicErrorV1.retry_after_ms bounds = %#v, want 1..10000", retryAfter)
	}
}

func assertNodeOpenAPIRegistryBindings(t *testing.T, bootstrap, agent, operator *openapi3.T) {
	t.Helper()
	assertPropertiesReference(t, "bootstrap UUID", bootstrap, "CanonicalUUID", map[string][]string{
		"ClaimNodeEnrollmentRequestV1":      {"attempt_id", "node_id"},
		"CertificateAuthorizationReceiptV1": {"issuance_id", "attempt_id", "node_id", "lineage_id", "issuer_id"},
		"NodeTimeAttestationV1":             {"node_id"},
	})
	assertPropertiesReference(t, "bootstrap digest", bootstrap, "DigestHex32", map[string][]string{
		"CertificateAuthorizationReceiptV1": {"csr_der_sha256", "leaf_der_sha256", "public_key_sha256"},
		"SignedNodeStateTrustMetadataV1":    {"digest"},
		"NodeTimeAttestationV1":             {"key_id"},
	})

	assertPropertiesReference(t, "agent UUID", agent, "CanonicalUUID", map[string][]string{
		"RotateNodeCertificateRequestV1":    {"attempt_id"},
		"DesiredPollRequestV1":              {"boot_id"},
		"RecoveryPollRequestV1":             {"boot_id", "recovery_id"},
		"SignedRecoveryStateV1":             {"recovery_id"},
		"NodeObservationV1":                 {"node_id", "boot_id"},
		"ObservationAckV1":                  {"boot_id"},
		"SupervisorFaultV1":                 {"supervisor_fault_id"},
		"SecurityFaultReportV1":             {"operation_id", "local_fault_id", "boot_id"},
		"SecurityFaultReceiptV1":            {"local_fault_id", "incident_id"},
		"TrustConflictEvidenceRequestV1":    {"incident_id"},
		"TrustConflictEvidenceAckV1":        {"incident_id"},
		"RecoveryAttestationV1":             {"recovery_id", "certificate_id"},
		"RecoveryAttestationAckV1":          {"recovery_id"},
		"CertificateAuthorizationReceiptV1": {"issuance_id", "attempt_id", "node_id", "lineage_id", "issuer_id"},
		"NodeTimeAttestationV1":             {"node_id"},
	})
	assertArrayItemsReference(t, "agent UUID array", agent, "CanonicalUUID", map[string][]string{
		"RecoveryPollRequestV1": {"sorted_known_incident_ids"},
	})
	assertPropertiesReference(t, "agent digest", agent, "DigestHex32", map[string][]string{
		"NodeHighWaterV1":                   {"digest"},
		"DesiredPollRequestV1":              {"agent_build_digest", "capability_schema_digest"},
		"SignedCanonicalArtifactV1":         {"digest"},
		"SignedRecoveryStateV1":             {"digest"},
		"NodeSlotFactV1":                    {"metrics_digest"},
		"NodeObservationV1":                 {"agent_build_digest", "capability_schema_digest", "seen_digest", "applied_digest", "reducer_digest"},
		"ObservationAckV1":                  {"report_digest"},
		"SupervisorFaultV1":                 {"evidence_digest"},
		"SecurityFaultReportV1":             {"evidence_digest", "request_digest"},
		"SecurityFaultReceiptV1":            {"request_digest"},
		"RecoveryAttestationV1":             {"recovery_snapshot_digest", "agent_build_digest", "supervisor_build_digest", "agent_guard_digest", "supervisor_guard_digest", "latch_guard_digest", "trusted_time_evidence_digest"},
		"RecoveryAttestationAckV1":          {"attestation_digest"},
		"CertificateAuthorizationReceiptV1": {"csr_der_sha256", "leaf_der_sha256", "public_key_sha256"},
		"NodeTimeAttestationV1":             {"key_id"},
	})
	assertPropertiesReference(t, "agent recovery reason", agent, "RecoveryReasonV1", map[string][]string{
		"SignedRecoveryStateV1": {"recovery_reason"},
	})

	assertPropertiesReference(t, "operator UUID", operator, "CanonicalUUID", map[string][]string{
		"FailureDomainV1":                {"failure_domain_id"},
		"NodeV1":                         {"node_id", "pending_transition_signing_id", "lineage_id", "active_root_publish_id", "active_metadata_publish_id", "last_authority_operation_id"},
		"NodeEndpointV1":                 {"endpoint_id", "node_id"},
		"NodeProcessSlotV1":              {"node_id"},
		"CreateNodePOPV1":                {"command_id"},
		"UpdateNodePOPV1":                {"command_id"},
		"CreateFailureDomainV1":          {"command_id", "failure_domain_id"},
		"UpdateFailureDomainV1":          {"command_id"},
		"CreateNodeV1":                   {"command_id", "node_id"},
		"UpdateNodeV1":                   {"command_id"},
		"CreateNodeEndpointV1":           {"command_id", "endpoint_id"},
		"UpdateNodeEndpointV1":           {"command_id"},
		"CreateNodeProcessSlotV1":        {"command_id"},
		"UpdateNodeProcessSlotV1":        {"command_id"},
		"CreateEnrollmentGrantV1":        {"command_id"},
		"EnrollmentGrantV1":              {"grant_id", "node_id"},
		"PutNodeDesiredStateV1":          {"command_id"},
		"SigningOperationV1":             {"signing_id"},
		"RecoveryOperationV1":            {"recovery_id"},
		"DrainNodeV1":                    {"command_id"},
		"DisableNodeV1":                  {"command_id"},
		"ReenrollNodeV1":                 {"command_id", "recovery_id"},
		"HostRemediationEvidenceV1":      {"evidence_id"},
		"CompleteReenrollmentV1":         {"command_id", "recovery_id", "certificate_id"},
		"RegisterHostSecurityIncidentV1": {"command_id"},
		"RegisterResourceEnvelopeV1":     {"command_id"},
		"ClearSecurityQuarantineV1":      {"command_id", "incident_id"},
		"ResumeAfterSecurityV1":          {"command_id", "recovery_id"},
		"ReauthorizeAfterRestoreV1":      {"command_id", "recovery_id", "proposal_id"},
		"RecoveryEnrollmentGrantV1":      {"recovery_id", "grant_id"},
		"RestoreReauthorizationV1":       {"proposal_id", "approval_id"},
	})
	assertPropertiesReference(t, "operator digest", operator, "DigestHex32", map[string][]string{
		"NodeV1":                       {"resource_envelope_digest"},
		"CreateEnrollmentGrantV1":      {"csr_der_sha256"},
		"EnrollmentGrantV1":            {"csr_der_sha256"},
		"PutNodeDesiredStateV1":        {"resource_envelope_digest"},
		"ReenrollNodeV1":               {"csr_der_sha256"},
		"HostRemediationEvidenceV1":    {"digest"},
		"CompleteReenrollmentV1":       {"recovery_attestation_digest"},
		"ResourceEnvelopePackageV1":    {"digest"},
		"ResourceEnvelopeActivationV1": {"digest"},
		"ClearSecurityQuarantineV1":    {"recovery_snapshot_digest", "recovery_attestation_digest"},
		"ResumeAfterSecurityV1":        {"recovery_attestation_digest"},
		"ReauthorizeAfterRestoreV1":    {"effect_digest"},
		"RecoveryEnrollmentGrantV1":    {"csr_der_sha256"},
	})
	assertPropertiesReference(t, "operator reason", operator, "OperatorReasonCodeV1", map[string][]string{
		"CreateNodePOPV1":         {"reason_code"},
		"UpdateNodePOPV1":         {"reason_code"},
		"CreateFailureDomainV1":   {"reason_code"},
		"UpdateFailureDomainV1":   {"reason_code"},
		"CreateNodeV1":            {"reason_code"},
		"UpdateNodeV1":            {"reason_code"},
		"CreateNodeEndpointV1":    {"reason_code"},
		"UpdateNodeEndpointV1":    {"reason_code"},
		"CreateNodeProcessSlotV1": {"reason_code"},
		"UpdateNodeProcessSlotV1": {"reason_code"},
		"CreateEnrollmentGrantV1": {"reason_code"},
		"DrainNodeV1":             {"reason_code"},
		"ReenrollNodeV1":          {"reason_code"},
	})
	assertPropertiesReference(t, "operator desired reason", operator, "DesiredReasonV1", map[string][]string{
		"PutNodeDesiredStateV1": {"reason_code"},
	})
	assertPropertiesReference(t, "operator disable reason", operator, "DisableReasonV1", map[string][]string{
		"DisableNodeV1": {"reason_code"},
	})
}

func assertPropertiesReference(t *testing.T, label string, spec *openapi3.T, registry string, expected map[string][]string) {
	t.Helper()
	want := "#/components/schemas/" + registry
	for schemaName, properties := range expected {
		schema := mustSchema(t, spec, schemaName)
		for _, propertyName := range properties {
			property, ok := schema.Properties[propertyName]
			if !ok {
				t.Errorf("%s: %s.%s is missing", label, schemaName, propertyName)
				continue
			}
			if got := schemaReference(property); got != want {
				t.Errorf("%s: %s.%s reference = %q, want %q", label, schemaName, propertyName, got, want)
			}
		}
	}
}

func assertArrayItemsReference(t *testing.T, label string, spec *openapi3.T, registry string, expected map[string][]string) {
	t.Helper()
	want := "#/components/schemas/" + registry
	for schemaName, properties := range expected {
		schema := mustSchema(t, spec, schemaName)
		for _, propertyName := range properties {
			property, ok := schema.Properties[propertyName]
			if !ok || property.Value == nil || property.Value.Items == nil {
				t.Errorf("%s: %s.%s array items are missing", label, schemaName, propertyName)
				continue
			}
			if got := schemaReference(property.Value.Items); got != want {
				t.Errorf("%s: %s.%s item reference = %q, want %q", label, schemaName, propertyName, got, want)
			}
		}
	}
}

func schemaReference(ref *openapi3.SchemaRef) string {
	if ref == nil {
		return ""
	}
	if ref.Ref != "" {
		return ref.Ref
	}
	if ref.Value != nil && len(ref.Value.AllOf) == 1 {
		return ref.Value.AllOf[0].Ref
	}
	return ""
}

func assertCSRDER(t *testing.T, spec *openapi3.T) {
	t.Helper()
	schema := mustSchema(t, spec, "CSRDER")
	if schema.MinLength != 2 || schema.MaxLength == nil || *schema.MaxLength != 21846 || len(schema.OneOf) != 3 {
		t.Errorf("CSRDER bounds/branches = %#v, want 2..21846 and three raw-base64url branches", schema)
	}
	if fmt.Sprint(schema.Extensions["x-talenro-base64url-padding"]) != "forbidden" || fmt.Sprint(schema.Extensions["x-talenro-min-decoded-bytes"]) != "1" || fmt.Sprint(schema.Extensions["x-talenro-max-decoded-bytes"]) != "16384" {
		t.Errorf("CSRDER decoded encoding contract = %#v", schema.Extensions)
	}
	wantPatterns := []string{`^(?:[A-Za-z0-9_-]{4})+$`, `^(?:[A-Za-z0-9_-]{4})*[A-Za-z0-9_-]{2}$`, `^(?:[A-Za-z0-9_-]{4})*[A-Za-z0-9_-]{3}$`}
	gotPatterns := make([]string, 0, 3)
	for _, branch := range schema.OneOf {
		if branch.Value != nil {
			gotPatterns = append(gotPatterns, branch.Value.Pattern)
		}
	}
	assertExactStringSet(t, "CSRDER patterns", gotPatterns, wantPatterns)
}

func assertEnum(t *testing.T, spec *openapi3.T, schemaName, propertyName string, want []string) {
	t.Helper()
	schema := mustSchema(t, spec, schemaName).Properties[propertyName]
	if schema == nil || schema.Value == nil {
		t.Fatalf("missing %s.%s", schemaName, propertyName)
	}
	assertEnumValues(t, schemaName+"."+propertyName, schema.Value, want)
}

func assertEnumRegistry(t *testing.T, spec *openapi3.T, schemaName string, want []string) {
	t.Helper()
	assertEnumValues(t, schemaName, mustSchema(t, spec, schemaName), want)
}

func assertEnumRegistryInOrder(t *testing.T, spec *openapi3.T, schemaName string, want []string) {
	t.Helper()
	schema := mustSchema(t, spec, schemaName)
	got := make([]string, 0, len(schema.Enum))
	for _, value := range schema.Enum {
		got = append(got, fmt.Sprint(value))
	}
	if strings.Join(got, "\x00") != strings.Join(want, "\x00") {
		t.Errorf("%s enum order = %v, want %v", schemaName, got, want)
	}
}

func assertEnumValues(t *testing.T, label string, schema *openapi3.Schema, want []string) {
	t.Helper()
	got := make([]string, 0, len(schema.Enum))
	for _, value := range schema.Enum {
		got = append(got, fmt.Sprint(value))
	}
	assertExactStringSet(t, label+" enum", got, want)
}

func assertClosedRequiredObjects(t *testing.T, location string, ref *openapi3.SchemaRef, seen map[*openapi3.Schema]bool) {
	t.Helper()
	if ref == nil || ref.Value == nil || seen[ref.Value] {
		return
	}
	schema := ref.Value
	seen[schema] = true
	if schema.Type != nil && schema.Type.Is(openapi3.TypeObject) {
		if schema.AdditionalProperties.Has == nil || *schema.AdditionalProperties.Has {
			t.Errorf("%s must set additionalProperties: false", location)
		}
		if len(schema.Required) == 0 {
			t.Errorf("%s must have a nonempty required array", location)
		}
	}
	for name, property := range schema.Properties {
		assertClosedRequiredObjects(t, location+"/properties/"+name, property, seen)
	}
	for index, branch := range schema.AllOf {
		assertClosedRequiredObjects(t, fmt.Sprintf("%s/allOf/%d", location, index), branch, seen)
	}
	for index, branch := range schema.AnyOf {
		assertClosedRequiredObjects(t, fmt.Sprintf("%s/anyOf/%d", location, index), branch, seen)
	}
	for index, branch := range schema.OneOf {
		assertClosedRequiredObjects(t, fmt.Sprintf("%s/oneOf/%d", location, index), branch, seen)
	}
	if schema.Not != nil {
		assertClosedRequiredObjects(t, location+"/not", schema.Not, seen)
	}
	if schema.Items != nil {
		assertClosedRequiredObjects(t, location+"/items", schema.Items, seen)
	}
}

func assertNoForbiddenOperatorProperties(t *testing.T, location string, ref *openapi3.SchemaRef, seen map[*openapi3.Schema]bool) {
	t.Helper()
	if ref == nil || ref.Value == nil || seen[ref.Value] {
		return
	}
	schema := ref.Value
	seen[schema] = true
	for name, property := range schema.Properties {
		switch name {
		case "total", "count", "offset", "sort", "include":
			t.Errorf("%s exposes forbidden property %q", location, name)
		}
		assertNoForbiddenOperatorProperties(t, location+"/properties/"+name, property, seen)
	}
	if schema.Items != nil {
		if schema.Items.Ref == "" && schema.Items.Value != nil && schema.Items.Value.Type == nil && len(schema.Items.Value.Properties) == 0 {
			t.Errorf("%s has an anonymous free-form items schema", location)
		}
		assertNoForbiddenOperatorProperties(t, location+"/items", schema.Items, seen)
	}
}

func assertExactStringSet(t *testing.T, label string, got, want []string) {
	t.Helper()
	got = append([]string(nil), got...)
	want = append([]string(nil), want...)
	sort.Strings(got)
	sort.Strings(want)
	if strings.Join(got, "\x00") != strings.Join(want, "\x00") {
		t.Errorf("%s = %v, want %v", label, got, want)
	}
}

func sameStringSet(got, want []string) bool {
	got = append([]string(nil), got...)
	want = append([]string(nil), want...)
	sort.Strings(got)
	sort.Strings(want)
	return strings.Join(got, "\x00") == strings.Join(want, "\x00")
}

func integerExtension(value any) (int64, bool) {
	switch value := value.(type) {
	case float64:
		integer := int64(value)
		return integer, float64(integer) == value
	case int64:
		return value, true
	case int:
		return int64(value), true
	default:
		parsed, err := strconv.ParseInt(fmt.Sprint(value), 10, 64)
		return parsed, err == nil
	}
}

func mapKeys[T any](values map[string]T) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
