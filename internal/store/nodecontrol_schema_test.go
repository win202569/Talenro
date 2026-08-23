package store_test

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

var nodeControlAuthorityTables = []string{
	"node_pops",
	"node_failure_domains",
	"node_inventory",
	"node_failure_domain_membership",
	"node_endpoints",
	"node_process_slots",
	"node_capacity_profiles",
	"node_resource_envelopes",
	"node_enrollment_grants",
	"node_certificate_issuances",
	"node_certificates",
	"node_state_signing_intents",
	"node_root_metadata_publish_intents",
	"node_root_metadata_signature_shares",
	"node_desired_states",
	"node_recovery_states",
	"node_observed_states",
	"node_state_transitions",
	"node_security_incidents",
	"node_security_fault_receipts",
	"node_recovery_sessions",
	"node_restore_reauthorization_approvals",
	"node_operator_audit",
	"control_plane_trust_bundle_high_waters",
	"control_plane_authority_fences",
}

// nodeControlTableCreationOrder is dependency ordered. The manifest's review
// order is intentionally independent from the physical creation order.
var nodeControlTableCreationOrder = []string{
	"control_plane_authority_fences",
	"node_pops",
	"node_failure_domains",
	"node_capacity_profiles",
	"node_inventory",
	"node_failure_domain_membership",
	"node_endpoints",
	"node_process_slots",
	"node_resource_envelopes",
	"node_certificate_issuances",
	"node_enrollment_grants",
	"node_certificates",
	"node_security_incidents",
	"node_security_fault_receipts",
	"node_recovery_sessions",
	"node_restore_reauthorization_approvals",
	"node_state_signing_intents",
	"node_root_metadata_publish_intents",
	"node_root_metadata_signature_shares",
	"node_desired_states",
	"node_recovery_states",
	"node_observed_states",
	"node_operator_audit",
	"node_state_transitions",
	"control_plane_trust_bundle_high_waters",
}

func TestNodeControlMigrationContainsEveryAuthorityTable(t *testing.T) {
	body, err := os.ReadFile("../../db/migrations/00006_nodecontrol.sql")
	if err != nil {
		t.Fatal(err)
	}
	for _, table := range nodeControlAuthorityTables {
		if !bytes.Contains(body, []byte("CREATE TABLE nodecontrol."+table)) {
			t.Fatalf("missing table %s", table)
		}
	}
}

func TestNodeControlMigrationIsExplicitlyReversible(t *testing.T) {
	body, err := os.ReadFile("../../db/migrations/00006_nodecontrol.sql")
	if err != nil {
		t.Fatal(err)
	}
	migration := string(body)
	if strings.Contains(strings.ToUpper(migration), "CASCADE") {
		t.Fatal("nodecontrol migration must not use CASCADE")
	}
	upMarker := strings.Index(migration, "-- +goose Up")
	downMarker := strings.Index(migration, "-- +goose Down")
	if upMarker < 0 || downMarker <= upMarker {
		t.Fatal("nodecontrol migration must contain ordered goose Up and Down sections")
	}
	down := migration[downMarker:]
	previous := -1
	for index := len(nodeControlTableCreationOrder) - 1; index >= 0; index-- {
		needle := "DROP TABLE nodecontrol." + nodeControlTableCreationOrder[index]
		position := strings.Index(down, needle)
		if position < 0 {
			t.Fatalf("Down section is missing %q", needle)
		}
		if position <= previous {
			t.Fatalf("Down section does not drop %s in exact reverse creation order", nodeControlTableCreationOrder[index])
		}
		previous = position
	}
	if schemaDrop := strings.Index(down, "DROP SCHEMA nodecontrol;"); schemaDrop <= previous {
		t.Fatal("Down section must drop nodecontrol schema after every table")
	}
}
