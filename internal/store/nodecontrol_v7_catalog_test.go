package store_test

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

var task8BaseTables = []string{
	"control_plane_authority_fences",
	"control_plane_trust_bundle_high_waters",
	"node_capacity_profiles",
	"node_certificate_issuances",
	"node_certificates",
	"node_desired_states",
	"node_endpoints",
	"node_enrollment_grants",
	"node_failure_domain_membership",
	"node_failure_domains",
	"node_inventory",
	"node_observed_states",
	"node_operator_audit",
	"node_pops",
	"node_process_slots",
	"node_recovery_sessions",
	"node_recovery_states",
	"node_resource_envelopes",
	"node_restore_reauthorization_approvals",
	"node_root_metadata_publish_intents",
	"node_root_metadata_signature_shares",
	"node_security_fault_receipts",
	"node_security_incidents",
	"node_state_signing_intents",
	"node_state_transitions",
}

var task8V7Tables = []string{
	"control_plane_authority_epoch_transition_applications",
	"control_plane_authority_epoch_transition_cancellations",
	"control_plane_authority_epoch_transition_intents",
	"control_plane_authority_epoch_transition_recovery_applications",
	"control_plane_authority_epoch_transition_recovery_intents",
	"control_plane_authority_epoch_transition_recovery_prefix_decisi",
	"control_plane_authority_epoch_transition_resolutions",
	"control_plane_authority_epoch_transition_terminal_applications",
	"control_plane_authority_fresh_restore_import_applications",
	"control_plane_authority_fresh_restore_requirements",
	"control_plane_authority_indeterminate_source_seals",
	"control_plane_authority_legacy_database_source_retirements",
	"control_plane_authority_legacy_source_seals",
	"control_plane_authority_protocol_activation_completions",
	"control_plane_authority_protocol_activation_releases",
	"control_plane_authority_protocol_activations",
	"control_plane_authority_protocol_downgrade_authorizations",
	"control_plane_authority_protocol_migration_latches",
	"control_plane_authority_protocol_upgrade_attempts",
	"control_plane_authority_protocol_upgrade_intents",
	"control_plane_authority_runtime_rebind_results",
	"control_plane_authority_runtime_registration_results",
	"control_plane_authority_staging_import_capabilities",
	"control_plane_authority_staging_import_capability_recovery_appl",
	"control_plane_authority_staging_import_capability_recovery_inte",
	"control_plane_authority_staging_import_capability_revocation_ap",
}

func TestNodeControlV7CatalogHasExact51Tables(t *testing.T) {
	raw := task8ReadFile(t, "../../db/schema/nodecontrol.v1.yaml")
	var document struct {
		Tables []struct {
			Name                  string  `yaml:"name"`
			IntroducedInMigration uint64  `yaml:"introduced_in_migration"`
			RetiredInMigration    *uint64 `yaml:"retired_in_migration"`
		} `yaml:"tables"`
	}
	if err := yaml.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	got := make([]string, 0, len(document.Tables))
	seen := make(map[string]bool, len(document.Tables))
	for _, table := range document.Tables {
		if table.IntroducedInMigration == 0 {
			t.Errorf("table %s lacks introduced_in_migration", table.Name)
		}
		if seen[table.Name] {
			t.Errorf("duplicate raw manifest table %s", table.Name)
		}
		seen[table.Name] = true
		got = append(got, table.Name)
	}
	want := append(append([]string(nil), task8BaseTables...), task8V7Tables...)
	sort.Strings(got)
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("raw final manifest tables = %v, want independent exact 25+26 registry %v", got, want)
	}
	if len(got) != 51 {
		t.Fatalf("raw final manifest table count = %d, want 51", len(got))
	}
}

func TestNodeControlV7ManifestHasLifecycleAwareFunctionCatalog(t *testing.T) {
	raw := task8ReadFile(t, "../../db/schema/nodecontrol.v1.yaml")
	var document struct {
		Functions []struct {
			Name                  string  `yaml:"name"`
			Arguments             string  `yaml:"arguments"`
			IntroducedInMigration uint64  `yaml:"introduced_in_migration"`
			RetiredInMigration    *uint64 `yaml:"retired_in_migration"`
			Language              string  `yaml:"language"`
			Volatility            string  `yaml:"volatility"`
			SecurityDefiner       bool    `yaml:"security_definer"`
			DefinitionSHA256      string  `yaml:"definition_sha256"`
		} `yaml:"functions"`
	}
	if err := yaml.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	wantNames := strings.Fields(`
begin_staging_import bytea_array_is_sorted_unique_32 enforce_authority_fence_update
enforce_capacity_profile_immutability enforce_certificate_issuance_workflow
enforce_certificate_workflow enforce_enrollment_grant_workflow enforce_inventory_pointers
enforce_observed_state enforce_process_slot_cap enforce_recovery_session_workflow
enforce_restore_approval_workflow enforce_root_publish_workflow enforce_root_share_binding
enforce_security_fault_receipt enforce_security_incident_workflow enforce_signing_intent_workflow
enforce_trust_bundle_high_water reject_row_mutation text_array_is_sorted_unique
v7_assert_activation_barrier v7_assert_source_writable v7_authority_proof_group_valid
v7_acquire_source_freeze_for_seal v7_consume_down_guard v7_guard_authority_proof_transition v7_insert_downgrade_authorization
v7_reject_immutable_mutation v7_require_role v7_source_is_frozen v7_text_array_is_sorted_unique`)
	gotNameSet := make(map[string]struct{}, len(document.Functions))
	replacementFunctionLifecycles := map[string][]task8LifecycleName{
		"enforce_authority_fence_update":     nil,
		"enforce_enrollment_grant_workflow":  nil,
		"enforce_security_incident_workflow": nil,
	}
	replacementV7SecurityDefiner := map[string]bool{
		"enforce_authority_fence_update":     true,
		"enforce_enrollment_grant_workflow":  false,
		"enforce_security_incident_workflow": false,
	}
	for _, function := range document.Functions {
		if function.IntroducedInMigration == 0 || function.RetiredInMigration != nil && *function.RetiredInMigration <= function.IntroducedInMigration {
			t.Fatalf("function %s has invalid lifecycle %d/%v", function.Name, function.IntroducedInMigration, function.RetiredInMigration)
		}
		if function.Language == "" || function.Volatility == "" || len(function.DefinitionSHA256) != 64 || function.DefinitionSHA256 == strings.Repeat("0", 64) {
			t.Fatalf("function %s omits exact catalog metadata", function.Name)
		}
		if function.Name == "v7_source_is_frozen" && function.Arguments == "" &&
			(function.IntroducedInMigration != 7 || function.RetiredInMigration != nil ||
				function.Language != "sql" || function.Volatility != "volatile" || !function.SecurityDefiner) {
			t.Fatalf("v7_source_is_frozen manifest metadata is stale: %+v", function)
		}
		if function.Name == "v7_acquire_source_freeze_for_seal" && function.Arguments == "" &&
			(function.IntroducedInMigration != 7 || function.RetiredInMigration != nil ||
				function.Language != "plpgsql" || function.Volatility != "volatile" || !function.SecurityDefiner) {
			t.Fatalf("v7_acquire_source_freeze_for_seal manifest metadata is stale: %+v", function)
		}
		if wantSecurityDefiner, ok := replacementV7SecurityDefiner[function.Name]; ok &&
			function.Arguments == "" && function.IntroducedInMigration == 7 &&
			(function.RetiredInMigration != nil || function.Language != "plpgsql" || function.Volatility != "volatile" ||
				function.SecurityDefiner != wantSecurityDefiner) {
			t.Fatalf("%s V7 replacement manifest metadata is stale: %+v", function.Name, function)
		}
		gotNameSet[function.Name] = struct{}{}
		if lifecycles, ok := replacementFunctionLifecycles[function.Name]; ok && function.Arguments == "" {
			replacementFunctionLifecycles[function.Name] = append(lifecycles, task8LifecycleName{
				name: function.Name, introduced: function.IntroducedInMigration, retired: function.RetiredInMigration,
			})
		}
	}
	gotNames := make([]string, 0, len(gotNameSet))
	for name := range gotNameSet {
		gotNames = append(gotNames, name)
	}
	sort.Strings(gotNames)
	sort.Strings(wantNames)
	if !reflect.DeepEqual(gotNames, wantNames) {
		t.Fatalf("manifest function names = %v, want independent exact registry %v", gotNames, wantNames)
	}
	if len(document.Functions) != 34 || len(gotNameSet) != 31 {
		t.Fatalf("manifest function catalog raw/unique counts=%d/%d, want 34/31", len(document.Functions), len(gotNameSet))
	}
	for name, lifecycles := range replacementFunctionLifecycles {
		sort.Slice(lifecycles, func(left, right int) bool {
			return lifecycles[left].introduced < lifecycles[right].introduced
		})
		if len(lifecycles) != 2 || lifecycles[0].introduced != 6 ||
			lifecycles[0].retired == nil || *lifecycles[0].retired != 7 ||
			lifecycles[1].introduced != 7 || lifecycles[1].retired != nil {
			t.Fatalf("%s manifest lifecycle = %#v, want exact 6-retired-7 then 7-active", name, lifecycles)
		}
	}
}

func TestNodeControlV7ManifestLifecycleViewsAreNonOverlapping(t *testing.T) {
	raw := task8ReadFile(t, "../../db/schema/nodecontrol.v1.yaml")
	manifest, err := decodeNodeControlManifest(raw)
	if err != nil {
		t.Fatal(err)
	}
	base := base25Manifest(manifest)
	final := final51Manifest(manifest)
	if len(base.Tables) != 25 || len(final.Tables) != 51 || len(base.Functions) != 19 || len(final.Functions) != 31 {
		t.Fatalf("manifest lifecycle views tables/functions = base:%d/%d final:%d/%d, want 25/19 and 51/31",
			len(base.Tables), len(base.Functions), len(final.Tables), len(final.Functions))
	}
	for _, table := range manifest.Tables {
		if table.IntroducedInMigration == 0 || table.PrimaryKey.IntroducedInMigration == 0 {
			t.Fatalf("table %s lacks explicit table/primary-key lifecycle", table.Name)
		}
		for _, version := range []uint64{6, 7} {
			requireTask8ActiveNamesUnique(t, table.Name+" columns", version, task8ColumnLifecycle(table.Columns))
			requireTask8ActiveNamesUnique(t, table.Name+" enums", version, task8EnumLifecycle(table.Enums))
			requireTask8ActiveNamesUnique(t, table.Name+" constraints", version, task8ConstraintLifecycle(table.Constraints))
			requireTask8ActiveNamesUnique(t, table.Name+" indexes", version, task8IndexLifecycle(table.Indexes))
			requireTask8ActiveNamesUnique(t, table.Name+" triggers", version, task8TriggerLifecycle(table.Triggers))
		}
	}
	for _, version := range []uint64{6, 7} {
		requireTask8ActiveNamesUnique(t, "functions", version, task8FunctionLifecycle(manifest.Functions))
	}
}

func TestNodeControlV7TrustedSourceSealerContract(t *testing.T) {
	rawUp := string(task8ReadFile(t, "../../db/migrations/assets/nodecontrol_authority_v7_up.sql"))
	up := strings.ToLower(task8StripSQLComments(rawUp))
	body := task8SQLFunction(t, up, "nodecontrol.v7_acquire_source_freeze_for_seal")
	normalized := strings.Join(strings.Fields(body), " ")

	requiredInOrder := []string{
		"perform nodecontrol.v7_require_role('nodecontrol_upgrade_executor');",
		"if pg_catalog.current_setting('transaction_isolation') is distinct from 'read committed' then",
		"raise exception 'authority v7 source sealing requires read committed' using errcode = '25001';",
		"perform pg_catalog.pg_advisory_xact_lock( pg_catalog.hashtextextended('nodecontrol:v7-source-freeze', 0) );",
	}
	last := -1
	for _, required := range requiredInOrder {
		index := strings.Index(normalized, required)
		if index <= last {
			t.Fatalf("trusted source sealer order for %q = %d after %d", required, index, last)
		}
		last = index
	}
	for _, required := range []string{
		"returns void", "language plpgsql", "volatile", "security definer",
		"set search_path = pg_catalog, nodecontrol",
	} {
		if got := strings.Count(normalized, required); got != 1 {
			t.Errorf("trusted source sealer metadata %q count=%d, want 1", required, got)
		}
	}
	for _, forbidden := range []string{"insert into", "execute ", "format(", "regclass", "jsonb"} {
		if strings.Contains(normalized, forbidden) {
			t.Errorf("trusted source sealer contains forbidden dynamic/payload token %q", forbidden)
		}
	}

	for _, required := range []string{
		"revoke execute on function nodecontrol.v7_acquire_source_freeze_for_seal() from public;",
		"alter function nodecontrol.v7_acquire_source_freeze_for_seal() owner to nodecontrol_upgrade_executor;",
		"grant execute on function nodecontrol.v7_acquire_source_freeze_for_seal() to current_user;",
	} {
		if got := strings.Count(up, required); got != 1 {
			t.Errorf("trusted source sealer ACL/owner %q count=%d, want 1", required, got)
		}
	}

	down := strings.ToLower(task8StripSQLComments(string(task8ReadFile(t, "../../db/migrations/assets/nodecontrol_authority_v7_down.sql"))))
	for _, required := range []string{
		"revoke execute on function nodecontrol.v7_acquire_source_freeze_for_seal() from current_user;",
		"drop function nodecontrol.v7_acquire_source_freeze_for_seal();",
	} {
		if got := strings.Count(down, required); got != 1 {
			t.Errorf("trusted source sealer Down %q count=%d, want 1", required, got)
		}
	}

	queries := strings.ToLower(string(task8ReadFile(t, "../../db/queries/nodecontrol_authority.sql")))
	const acquireQuery = "-- name: acquireauthorityv7sourcefreezeforseal :exec\nselect nodecontrol.v7_acquire_source_freeze_for_seal();"
	if got := strings.Count(strings.ReplaceAll(queries, "\r\n", "\n"), acquireQuery); got != 1 {
		t.Fatalf("trusted source sealer sqlc query count=%d, want exact 1", got)
	}
}

func TestNodeControlV7ManifestGuardTriggerTopology(t *testing.T) {
	raw := task8ReadFile(t, "../../db/schema/nodecontrol.v1.yaml")
	manifest, err := decodeNodeControlManifest(raw)
	if err != nil {
		t.Fatal(err)
	}
	manifest = manifestAtMigration(manifest, 7)
	wantSourceTables := make(map[string]bool, len(task8BaseTables)-1)
	for _, table := range task8BaseTables {
		if table != "control_plane_authority_fences" {
			wantSourceTables[table] = true
		}
	}
	wantProofTables := map[string]string{
		"node_resource_envelopes":            "ncv7_node_resource_envelopes_proof_guard",
		"node_certificate_issuances":         "ncv7_node_certificate_issuances_proof_guard",
		"node_enrollment_grants":             "ncv7_node_enrollment_grants_proof_guard",
		"node_certificates":                  "ncv7_node_certificates_proof_guard",
		"node_security_incidents":            "ncv7_node_security_incidents_proof_guard",
		"node_state_signing_intents":         "ncv7_node_state_signing_intents_proof_guard",
		"node_root_metadata_publish_intents": "ncv7_node_root_metadata_publish_intents_proof_guard",
		"node_state_transitions":             "ncv7_node_state_transitions_proof_guard",
	}
	wantMarkerTables := map[string]string{
		"control_plane_authority_legacy_database_source_retirements": "ncv7_dr_immutable_ud",
		"control_plane_authority_indeterminate_source_seals":         "ncv7_is_immutable_ud",
		"control_plane_authority_legacy_source_seals":                "ncv7_ls_immutable_ud",
		"control_plane_authority_fresh_restore_requirements":         "ncv7_fr_immutable_ud",
	}
	proofSeen := make(map[string]int, len(wantProofTables))
	markerSeen := make(map[string]int, len(wantMarkerTables))
	sourceCount := 0
	proofCount := 0
	closureCount := 0
	markerCount := 0
	for _, table := range manifest.Tables {
		tableSourceCount := 0
		for _, trigger := range table.Triggers {
			if trigger.Function == "v7_assert_source_writable()" {
				sourceCount++
				tableSourceCount++
				wantName := "ncv7_00_fence_source_lock"
				if table.Name != "control_plane_authority_fences" {
					if !wantSourceTables[table.Name] {
						t.Errorf("manifest source guard appears on non-source table %s", table.Name)
					}
					if trigger.Name != "ncv7_00_source_lock" && trigger.Name != "ncv7_01_source_guard" {
						t.Errorf("manifest source guard %s.%s has an unregistered name", table.Name, trigger.Name)
					}
				} else if trigger.Name != wantName {
					t.Errorf("manifest fence source guard name=%s, want %s", trigger.Name, wantName)
				}
				if trigger.IntroducedInMigration != 7 || trigger.RetiredInMigration != nil || trigger.Timing != "BEFORE" ||
					!reflect.DeepEqual(trigger.Events, []string{"INSERT", "UPDATE", "DELETE"}) || trigger.WhenSQL != nil {
					t.Errorf("manifest source guard %s.%s has inexact lifecycle/timing/events/WHEN: %+v", table.Name, trigger.Name, trigger)
				}
			}
			if trigger.Name == "ncv7_fence_owner_closure" {
				closureCount++
				if table.Name != "control_plane_authority_fences" || trigger.Function != "v7_assert_activation_barrier()" ||
					trigger.IntroducedInMigration != 7 || trigger.RetiredInMigration != nil || trigger.Timing != "AFTER" ||
					!reflect.DeepEqual(trigger.Events, []string{"UPDATE"}) || trigger.WhenSQL != nil {
					t.Errorf("manifest deferred fence closure is inexact: table=%s trigger=%+v", table.Name, trigger)
				}
			}
			if strings.HasPrefix(trigger.Name, "ncv7_") && strings.HasSuffix(trigger.Name, "_proof_guard") {
				proofCount++
				wantName, registered := wantProofTables[table.Name]
				proofSeen[table.Name]++
				if !registered || trigger.Name != wantName || trigger.Function != "v7_guard_authority_proof_transition()" ||
					trigger.IntroducedInMigration != 7 || trigger.RetiredInMigration != nil || trigger.Timing != "BEFORE" ||
					!reflect.DeepEqual(trigger.Events, []string{"INSERT", "UPDATE", "DELETE"}) || trigger.WhenSQL != nil {
					t.Errorf("manifest proof guard %s.%s is inexact: %+v", table.Name, trigger.Name, trigger)
				}
			}
			if wantName, markerTable := wantMarkerTables[table.Name]; markerTable && trigger.Name == wantName {
				markerCount++
				markerSeen[table.Name]++
				if trigger.Function != "v7_reject_immutable_mutation()" || trigger.IntroducedInMigration != 7 || trigger.RetiredInMigration != nil ||
					trigger.Timing != "BEFORE" || !reflect.DeepEqual(trigger.Events, []string{"INSERT", "UPDATE", "DELETE"}) ||
					trigger.WhenSQL != nil {
					t.Errorf("manifest freeze marker guard %s.%s is inexact: %+v", table.Name, trigger.Name, trigger)
				}
			}
		}
		switch {
		case wantSourceTables[table.Name] && tableSourceCount != 2:
			t.Errorf("manifest source table %s guard count=%d, want statement+row pair", table.Name, tableSourceCount)
		case table.Name == "control_plane_authority_fences" && tableSourceCount != 1:
			t.Errorf("manifest fence source guard count=%d, want 1", tableSourceCount)
		case !wantSourceTables[table.Name] && table.Name != "control_plane_authority_fences" && tableSourceCount != 0:
			t.Errorf("manifest non-source table %s source guard count=%d, want 0", table.Name, tableSourceCount)
		}
	}
	for table, name := range wantProofTables {
		if proofSeen[table] != 1 {
			t.Errorf("manifest proof guard %s.%s count=%d, want 1", table, name, proofSeen[table])
		}
	}
	for table, name := range wantMarkerTables {
		if markerSeen[table] != 1 {
			t.Errorf("manifest freeze marker guard %s.%s count=%d, want 1", table, name, markerSeen[table])
		}
	}
	if sourceCount != 49 || proofCount != 8 || closureCount != 1 || markerCount != 4 {
		t.Fatalf("manifest source/proof/closure/marker trigger counts=%d/%d/%d/%d, want exact 49/8/1/4",
			sourceCount, proofCount, closureCount, markerCount)
	}
}

func TestNodeControlV7ManifestFenceClaimTemporalConstraint(t *testing.T) {
	raw := task8ReadFile(t, "../../db/schema/nodecontrol.v1.yaml")
	manifest, err := decodeNodeControlManifest(raw)
	if err != nil {
		t.Fatal(err)
	}
	manifest = manifestAtMigration(manifest, 7)
	wantDefinition := "CHECK (authority_protocol_profile = 'legacy_v6'::text AND abort_claimed_at IS NULL AND protocol_activation_id IS NULL AND (provider_status = 'aborted'::text) = (abort_reason IS NOT NULL) OR authority_protocol_profile = 'claim_v1'::text AND protocol_activation_id IS NOT NULL AND (abort_claimed_at IS NULL AND abort_reason IS NULL OR abort_claimed_at IS NOT NULL AND abort_reason IS NOT NULL) AND (provider_status <> 'aborted'::text OR abort_claimed_at IS NOT NULL) AND (abort_claimed_at IS NULL OR effect_digest IS NULL) AND (abort_claimed_at IS NULL OR abort_claimed_at >= reserved_at) AND (terminal_at IS NULL OR abort_claimed_at IS NULL OR terminal_at >= abort_claimed_at))"
	wantColumns := []string{
		"authority_protocol_profile", "abort_claimed_at", "protocol_activation_id",
		"provider_status", "abort_reason", "effect_digest", "reserved_at", "terminal_at",
	}
	for _, table := range manifest.Tables {
		if table.Name != "control_plane_authority_fences" {
			continue
		}
		for _, constraint := range table.Constraints {
			if constraint.Name != "ncv7_fence_claim_ck" {
				continue
			}
			if constraint.Kind != "check" || constraint.DefinitionSQL != wantDefinition ||
				!reflect.DeepEqual(constraint.Columns, wantColumns) {
				t.Fatalf("manifest fence claim constraint is stale: %+v", constraint)
			}
			return
		}
		t.Fatal("manifest omits ncv7_fence_claim_ck")
	}
	t.Fatal("manifest omits control_plane_authority_fences")
}

type task8LifecycleName struct {
	name       string
	introduced uint64
	retired    *uint64
}

func requireTask8ActiveNamesUnique(t *testing.T, label string, version uint64, values []task8LifecycleName) {
	t.Helper()
	seen := make(map[string]bool)
	for _, value := range values {
		if value.introduced == 0 || value.retired != nil && *value.retired <= value.introduced {
			t.Fatalf("%s object %s has invalid lifecycle %d/%v", label, value.name, value.introduced, value.retired)
		}
		if !manifestLifecycleActive(value.introduced, value.retired, version) {
			continue
		}
		if seen[value.name] {
			t.Fatalf("%s has overlapping active duplicate %s at migration %d", label, value.name, version)
		}
		seen[value.name] = true
	}
}

func task8ColumnLifecycle(values []nodeControlColumnSpec) []task8LifecycleName {
	result := make([]task8LifecycleName, len(values))
	for index, value := range values {
		result[index] = task8LifecycleName{value.Name, value.IntroducedInMigration, value.RetiredInMigration}
	}
	return result
}
func task8EnumLifecycle(values []nodeControlEnumSpec) []task8LifecycleName {
	result := make([]task8LifecycleName, len(values))
	for index, value := range values {
		result[index] = task8LifecycleName{value.Name, value.IntroducedInMigration, value.RetiredInMigration}
	}
	return result
}
func task8ConstraintLifecycle(values []nodeControlConstraintSpec) []task8LifecycleName {
	result := make([]task8LifecycleName, len(values))
	for index, value := range values {
		result[index] = task8LifecycleName{value.Name, value.IntroducedInMigration, value.RetiredInMigration}
	}
	return result
}
func task8IndexLifecycle(values []nodeControlIndexSpec) []task8LifecycleName {
	result := make([]task8LifecycleName, len(values))
	for index, value := range values {
		result[index] = task8LifecycleName{value.Name, value.IntroducedInMigration, value.RetiredInMigration}
	}
	return result
}
func task8TriggerLifecycle(values []nodeControlTriggerSpec) []task8LifecycleName {
	result := make([]task8LifecycleName, len(values))
	for index, value := range values {
		result[index] = task8LifecycleName{value.Name, value.IntroducedInMigration, value.RetiredInMigration}
	}
	return result
}

func task8FunctionLifecycle(values []nodeControlFunctionSpec) []task8LifecycleName {
	result := make([]task8LifecycleName, len(values))
	for index, value := range values {
		result[index] = task8LifecycleName{value.Name + "(" + value.Arguments + ")", value.IntroducedInMigration, value.RetiredInMigration}
	}
	return result
}

func TestNodeControlV7RegisteredMigrationOnly(t *testing.T) {
	source := strings.ToLower(string(task8ReadFile(t, "../../db/migrations/00007_nodecontrol_authority_abort_serving.go")))
	for _, required := range []string{
		"package migrations", "goose.newgomigration", "nodecontrolauthorityv7migration", "go:embed assets/nodecontrol_authority_v7_up.sql",
		"go:embed assets/nodecontrol_authority_v7_down.sql", "authorityv7migrationcontext", "withauthorityv7migrationcontext",
	} {
		if !strings.Contains(source, required) {
			t.Errorf("registered migration shell lacks %q", required)
		}
	}
	for _, forbidden := range []string{"addmigration(", "addnamedmigration(", "init()", "func nodecontrolauthorityv7upsql(", "func nodecontrolauthorityv7downsql("} {
		if strings.Contains(source, forbidden) {
			t.Errorf("registered migration shell contains forbidden global/raw path %q", forbidden)
		}
	}
}

func task8ChildGoCommand(t *testing.T, directory string, arguments ...string) *exec.Cmd {
	t.Helper()
	repositoryRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	approvedOverlayPath, err := filepath.Abs(filepath.Join(repositoryRoot, ".superpowers", "sdd", "task-8-corrective-implementation-plan", "task-2-overlay-gate", "overlay.json"))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(approvedOverlayPath)
	if err != nil {
		t.Fatal(err)
	}
	if got := sha256.Sum256(raw); fmt.Sprintf("%X", got) != "39ED8367879A0A5D25F2AD6F4C6A2AC3B5F5CE9E3A77EEA3382C7C52311DF0E2" {
		t.Fatalf("approved child overlay SHA-256 = %X", got)
	}

	environment := make([]string, 0, len(os.Environ())+6)
	for _, item := range os.Environ() {
		name, _, _ := strings.Cut(item, "=")
		switch {
		case strings.EqualFold(name, "GOOS"), strings.EqualFold(name, "GOARCH"), strings.EqualFold(name, "CGO_ENABLED"), strings.EqualFold(name, "GOFLAGS"):
			continue
		}
		environment = append(environment, item)
	}
	environment = append(environment,
		"GOOS=windows",
		"GOARCH=amd64",
		"CGO_ENABLED=0",
		"GOFLAGS=-overlay="+approvedOverlayPath,
		"GOPROXY=off",
		"GOWORK=off",
	)
	command := exec.Command("go", arguments...)
	command.Dir = directory
	command.Env = environment
	return command
}

func TestNodeControlV7ProviderScopedMigration(t *testing.T) {
	if _, err := os.Stat("../../db/migrations/00007_nodecontrol_authority_abort_serving.go"); err != nil {
		t.Fatalf("provider-scoped migration is missing: %v", err)
	}
	repositoryRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	probeParent := filepath.Join(repositoryRoot, ".superpowers", "sdd", ".t8")
	if err := os.MkdirAll(probeParent, 0o700); err != nil {
		t.Fatal(err)
	}
	probeDirectory, err := os.MkdirTemp(probeParent, "provider-probe-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(probeDirectory) })
	probe := fmt.Sprintf(`package providerprobe
import (
 "database/sql"
 "os"
 "path/filepath"
 "strings"
 "testing"
 _ "github.com/jackc/pgx/v5/stdlib"
 "github.com/pressly/goose/v3"
 migrations "talenro.local/platform/db/migrations"
)
func TestProviderScope(t *testing.T) {
 db, err := sql.Open("pgx", "postgres://task8.invalid/never_connect")
 if err != nil { t.Fatal(err) }
 defer db.Close()
 fsys := os.DirFS(filepath.Join(%q, "db", "migrations"))
 _, ordinaryErr := goose.NewProvider(goose.DialectPostgres, db, fsys)
 if ordinaryErr == nil || !strings.Contains(ordinaryErr.Error(), "00007") { t.Fatalf("ordinary provider error = %%v, want unregistered 00007", ordinaryErr) }
 dedicated, err := goose.NewProvider(goose.DialectPostgres, db, fsys, goose.WithDisableGlobalRegistry(true), goose.WithGoMigrations(migrations.NodeControlAuthorityV7Migration()))
 if err != nil { t.Fatal(err) }
 sources := dedicated.ListSources()
 if len(sources) != 7 || sources[6].Version != 7 || sources[6].Type != goose.TypeGo { t.Fatalf("dedicated sources = %%#v, want exact versions 1..7 with Go v7", sources) }
 _, afterErr := goose.NewProvider(goose.DialectPostgres, db, fsys)
 if afterErr == nil || !strings.Contains(afterErr.Error(), "00007") { t.Fatalf("dedicated provider mutated global registry: %%v", afterErr) }
}
`, repositoryRoot)
	if err := os.WriteFile(filepath.Join(probeDirectory, "provider_test.go"), []byte(probe), 0o600); err != nil {
		t.Fatal(err)
	}
	command := task8ChildGoCommand(t, probeDirectory, "test", ".", "-run", "^TestProviderScope$", "-count=1")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("provider-scoped semantic probe failed: %v\n%s", err, output)
	}
}

func TestSQLCV7SchemaInputOrder(t *testing.T) {
	var config struct {
		SQL []struct {
			Schema any `yaml:"schema"`
		} `yaml:"sql"`
	}
	if err := yaml.Unmarshal(task8ReadFile(t, "../../sqlc.yaml"), &config); err != nil {
		t.Fatal(err)
	}
	if len(config.SQL) != 1 {
		t.Fatalf("sqlc PostgreSQL unit count = %d, want 1", len(config.SQL))
	}
	want := []any{"db/migrations", "db/migrations/assets/nodecontrol_authority_v7_up.sql"}
	if !reflect.DeepEqual(config.SQL[0].Schema, want) {
		t.Fatalf("sqlc schema input = %#v, want exact ordered %#v", config.SQL[0].Schema, want)
	}
}

func TestNodeInventoryV7FreshImportShape(t *testing.T) {
	up := strings.ToLower(string(task8ReadFile(t, "../../db/migrations/assets/nodecontrol_authority_v7_up.sql")))
	down := strings.ToLower(string(task8ReadFile(t, "../../db/migrations/assets/nodecontrol_authority_v7_down.sql")))
	for _, required := range []string{
		"identity_state = 'unauthorized'", "operator_state = 'disabled'", "security_state = 'quarantined'", "identity_epoch = 0",
		"lineage_id is null", "resume_operator_state is null", "next_desired_generation = 1", "next_recovery_generation = 1",
		"begin_staging_import", "nodecontrol_staging_importer",
	} {
		if !strings.Contains(up, required) {
			t.Errorf("fresh-import unauthorized shape lacks %q", required)
		}
	}
	if strings.Contains(down, "identity_state = 'unauthorized'") || strings.Contains(down, "identity_state in ('never_enrolled', 'active', 'recovery_pending', 'recovery_limited', 'revoked', 'unauthorized')") {
		t.Fatal("Down asset retains response-only unauthorized in the base inventory check")
	}
}

func TestNodeControlV7RoleAndGuardRegistry(t *testing.T) {
	rawUp := string(task8ReadFile(t, "../../db/migrations/assets/nodecontrol_authority_v7_up.sql"))
	up := strings.ToLower(rawUp)
	roleStatement := strings.ToLower(task8AuthorityV7RolePreflightStatement(t, rawUp))
	postflightStatement := strings.ToLower(task8AuthorityV7PostflightStatement(t, rawUp))
	roles := []string{"nodecontrol_upgrade_executor", "nodecontrol_migration_downgrader", "nodecontrol_staging_importer"}
	roleExecutable := strings.Join(strings.Fields(task8StripSQLComments(roleStatement)), " ")
	postflightExecutable := strings.Join(strings.Fields(task8StripSQLComments(postflightStatement)), " ")
	const absentRegistry = "select count(*) into role_count from pg_catalog.pg_roles where rolname in ( 'nodecontrol_upgrade_executor', 'nodecontrol_migration_downgrader', 'nodecontrol_staging_importer' );"
	registryIndex := strings.Index(roleExecutable, absentRegistry)
	guardIndex := strings.Index(roleExecutable, "if role_count <> 0 then")
	rejectionIndex := strings.Index(roleExecutable, "raise exception 'authority v7 capability role names must all be absent before creation' using errcode = '42710';")
	createIndex := strings.Index(roleExecutable, "create role ")
	if registryIndex < 0 || guardIndex <= registryIndex || rejectionIndex <= guardIndex || createIndex <= rejectionIndex {
		t.Fatalf("v7 capability-role preflight order = registry:%d guard:%d rejection:%d create:%d", registryIndex, guardIndex, rejectionIndex, createIndex)
	}
	if got := strings.Count(roleExecutable, "if role_count <> 0 then"); got != 1 {
		t.Fatalf("v7 capability-role absent guard count = %d, want 1", got)
	}
	if got := strings.Count(roleExecutable, "create role "); got != len(roles) {
		t.Fatalf("v7 capability-role CREATE count = %d, want %d", got, len(roles))
	}
	for _, role := range roles {
		declaration := "create role " + role + " nologin nosuperuser nocreatedb nocreaterole inherit noreplication nobypassrls connection limit -1 password null;"
		if strings.Count(roleExecutable, declaration) != 1 {
			t.Errorf("v7 ACL lacks exact capability-role declaration %q", declaration)
		}
	}
	for _, forbidden := range []string{"if not exists", "alter role", "set role", "pg_has_role", "grant nodecontrol_"} {
		if strings.Contains(roleExecutable, forbidden) {
			t.Errorf("v7 capability-role preflight contains adoption path %q", forbidden)
		}
	}
	helpers := []string{
		"v7_text_array_is_sorted_unique", "v7_reject_immutable_mutation", "v7_require_role", "v7_acquire_source_freeze_for_seal", "v7_source_is_frozen",
		"v7_assert_source_writable", "v7_assert_activation_barrier", "v7_authority_proof_group_valid",
		"v7_guard_authority_proof_transition", "v7_insert_downgrade_authorization", "v7_consume_down_guard", "begin_staging_import",
	}
	for _, helper := range helpers {
		if !strings.Contains(up, "nodecontrol."+helper) {
			t.Errorf("v7 helper registry lacks nodecontrol.%s", helper)
		}
	}
	for _, required := range []string{"security definer", "set search_path = pg_catalog, nodecontrol", "revoke execute", "from public"} {
		if !strings.Contains(up, required) {
			t.Errorf("v7 restricted-function ACL lacks %q", required)
		}
	}
	if strings.Contains(up, "pg_catalog.coalesce(") {
		t.Fatal("v7 asset schema-qualifies the non-function COALESCE grammar construct")
	}
	if matches := regexp.MustCompile(`(?i)\bAS\s+authorization\b`).FindAllStringIndex(up, -1); len(matches) != 0 {
		t.Fatalf("v7 asset uses reserved AUTHORIZATION as a SQL alias %d times", len(matches))
	}
	consume := task8SQLFunction(t, up, "nodecontrol.v7_consume_down_guard")
	forbiddenCountDeclarations := regexp.MustCompile(`(?m)^\s*v_forbidden_count\s+bigint;\s*$`).FindAllStringIndex(consume, -1)
	forbiddenCountTokens := regexp.MustCompile(`\bv_forbidden_count\b`).FindAllStringIndex(consume, -1)
	if len(forbiddenCountDeclarations) != 1 || len(forbiddenCountTokens) != 13 {
		t.Fatalf("v7 consume forbidden-count variable registry = declarations:%d tokens:%d, want 1/13", len(forbiddenCountDeclarations), len(forbiddenCountTokens))
	}

	insert := task8SQLFunction(t, up, "nodecontrol.v7_insert_downgrade_authorization")
	requireRole := task8SQLFunction(t, up, "nodecontrol.v7_require_role")
	acquireSourceFreeze := task8SQLFunction(t, up, "nodecontrol.v7_acquire_source_freeze_for_seal")
	beginStaging := task8SQLFunction(t, up, "nodecontrol.begin_staging_import")
	task8AssertV7RequireRoleStructuralOracle(t, requireRole)
	t.Run("round5 exact role and ACL closure", func(t *testing.T) {
		for _, stale := range []string{
			"nodecontrol_authority_v7_upgrade_executor",
			"nodecontrol_authority_v7_protocol_downgrader",
			"nodecontrol_authority_v7_inventory_staging_writer",
		} {
			if strings.Contains(up, stale) {
				t.Errorf("v7 asset retains stale capability-role identifier %q", stale)
			}
		}
		for name, body := range map[string]string{"insert": insert, "consume": consume} {
			if got := strings.Count(body, "perform nodecontrol.v7_require_role('nodecontrol_migration_downgrader');"); got != 2 {
				t.Errorf("%s capability closure invocation count = %d, want exact pre/post-lock 2", name, got)
			}
		}
		if got := strings.Count(beginStaging, "perform nodecontrol.v7_require_role('nodecontrol_staging_importer');"); got != 1 {
			t.Errorf("begin_staging_import capability closure invocation count = %d, want 1", got)
		}
		if got := strings.Count(acquireSourceFreeze, "perform nodecontrol.v7_require_role('nodecontrol_upgrade_executor');"); got != 1 {
			t.Errorf("v7_acquire_source_freeze_for_seal capability closure invocation count = %d, want 1", got)
		}
		for _, required := range []string{
			"rolinherit", "rolconnlimit", "rolvaliduntil", "pg_catalog.pg_db_role_setting",
			"pg_catalog.pg_auth_members", "pg_catalog.pg_shdepend", "pg_catalog.pg_default_acl",
			"pg_catalog.pg_init_privs", "pg_catalog.aclexplode", "except all",
		} {
			if !strings.Contains(requireRole, required) {
				t.Errorf("v7_require_role exact closure lacks %q", required)
			}
		}
		for label, proof := range map[string]struct {
			source string
			want   []string
		}{
			"role preflight": {
				source: roleExecutable,
				want: []string{
					"from pg_catalog.pg_authid as role_catalog",
					"role_catalog.rolpassword is null",
					"password_count <> 3",
				},
			},
			"postflight": {
				source: postflightExecutable,
				want: []string{
					"join pg_catalog.pg_authid as authorization_catalog on authorization_catalog.oid = role_catalog.oid",
					"authorization_catalog.rolpassword is null",
					"if role_count <> 3 or drift_count <> 0 then",
				},
			},
		} {
			for _, required := range proof.want {
				if got := strings.Count(proof.source, required); got != 1 {
					t.Errorf("v7 %s password-null proof %q count=%d, want 1", label, required, got)
				}
			}
		}
		executableUp := strings.ToLower(task8StripSQLCommentsAndStrings(rawUp))
		if got := len(regexp.MustCompile(`\b(?:from|join)\s+pg_catalog\.pg_authid\b`).FindAllStringIndex(executableUp, -1)); got != 2 {
			t.Errorf("v7 executable pg_authid reader count=%d, want exact creation preflight and final postflight 2", got)
		}
		if got := len(regexp.MustCompile(`\brolpassword\b`).FindAllStringIndex(executableUp, -1)); got != 2 {
			t.Errorf("v7 executable rolpassword reference count=%d, want exact creation preflight and final postflight 2", got)
		}
		columnACLExplodes := regexp.MustCompile(`(?i)pg_catalog\.aclexplode\(\s*attribute\.attacl\s*\)`).FindAllStringIndex(up, -1)
		columnACLGuards := regexp.MustCompile(`(?i)where\s+attribute\.attacl\s+is\s+not\s+null`).FindAllStringIndex(up, -1)
		if len(columnACLExplodes) != 5 || len(columnACLGuards) != 5 {
			t.Errorf("column ACL one-dimensional scan registry = explode:%d guard:%d, want exact 5/5", len(columnACLExplodes), len(columnACLGuards))
		}
		for _, forbidden := range []string{"coalesce(attribute.attacl", "pg_catalog.acldefault('c'::\"char\""} {
			if strings.Contains(up, forbidden) {
				t.Errorf("column ACL scan can synthesize a zero-dimensional ACL array via %q", forbidden)
			}
		}
	})

	t.Run("round5 multi namespace evidence exact cover", func(t *testing.T) {
		if strings.Contains(insert, "1 + 3 * jsonb_array_length(v_retirement_set->'retirements')") {
			t.Error("retirement evidence still uses the single-namespace 1+3P formula")
		}
		for _, required := range []string{
			"v_provider_namespace_count", "v_namespace_evidence", "v_namespace_states_jcs",
			"encode(convert_to(v_provider_namespace, 'UTF8'), 'hex')",
			"1 + 2 * jsonb_array_length(v_retirement_set->'retirements') + v_provider_namespace_count",
		} {
			if !strings.Contains(insert, strings.ToLower(required)) {
				t.Errorf("multi-namespace evidence closure lacks %q", required)
			}
		}
		if strings.Contains(insert, "jsonb_array_length(v_evidence_body->'namespace_states') <> 1") {
			t.Error("provider retirement response still requires exactly one namespace state")
		}
	})

	t.Run("round5 provider pair JSON extraction precedence", func(t *testing.T) {
		safePair := regexp.MustCompile(`(?m)\((?:v_evidence_body|value)->>'provider_identity_digest'\)\s*\|\|\s*':'\s*\|\|\s*\((?:v_evidence_body|value)->>'provider_endpoint_identity_digest'\)`)
		unsafePair := regexp.MustCompile(`(?m)(?:v_evidence_body|value)->>'provider_identity_digest'\s*\|\|\s*':'\s*\|\|\s*(?:v_evidence_body|value)->>'provider_endpoint_identity_digest'`)
		if got := len(safePair.FindAllStringIndex(insert, -1)); got != 5 {
			t.Errorf("parenthesized provider-pair extraction registry = %d, want exact 5", got)
		}
		if got := len(unsafePair.FindAllStringIndex(insert, -1)); got != 0 {
			t.Errorf("unparenthesized provider-pair extraction registry = %d, want 0", got)
		}
	})

	t.Run("round5 signature metadata closure", func(t *testing.T) {
		if got := strings.Count(insert, "'ecdsa-p256-sha256'"); got < 2 {
			t.Errorf("P-256 signature algorithm registry count = %d, want top-level and nested", got)
		}
		if got := strings.Count(insert, "^[1-9][0-9]*$"); got < 2 {
			t.Errorf("canonical positive signature policy registry count = %d, want top-level and nested", got)
		}
		if got := strings.Count(insert, "^[a-za-z0-9_-]{85}[aqgw]$"); got < 2 {
			t.Errorf("canonical raw-64 signature registry count = %d, want top-level and nested", got)
		}
		for _, stale := range []string{
			"signature_algorithm' is distinct from 'ed25519'",
			"signature_policy_version' is distinct from '1'",
		} {
			if strings.Contains(insert, stale) {
				t.Errorf("signature metadata retains single-value restriction %q", stale)
			}
		}
	})

	t.Run("round5 response request digest binding", func(t *testing.T) {
		for _, required := range []string{
			"talenro.c12.retire-environment-inventory-membership-request.v1",
			"talenro.c12.retire-provider-protocol-for-downgrade-request.v1",
			"v_membership_request_body_jcs", "v_provider_request_body_jcs",
		} {
			if !strings.Contains(insert, required) {
				t.Errorf("response request-digest binding lacks %q", required)
			}
		}
	})

	t.Run("round5 historical evidence validity", func(t *testing.T) {
		for _, required := range []string{"v_validation_now timestamptz;", "v_validation_now := clock_timestamp();"} {
			if !strings.Contains(insert, required) {
				t.Errorf("historical evidence validation lacks %q", required)
			}
		}
		for _, stale := range []string{
			"(v_evidence_body->>'expires_at')::timestamptz <= clock_timestamp()",
			"(v_evidence_body->>'issued_at')::timestamptz > clock_timestamp()",
			"(v_evidence_body->>'observed_at')::timestamptz > clock_timestamp()",
		} {
			if strings.Contains(insert, stale) {
				t.Errorf("historical evidence incorrectly requires current freshness: %s", stale)
			}
		}
		for _, required := range []string{
			"v_membership_retired_at", "v_admin_issued_at", "v_admin_expires_at",
			"v_history_observed_at", "v_history_expires_at", "v_final_retired_at",
		} {
			if !strings.Contains(insert, required) {
				t.Errorf("historical retired-at cross-binding lacks %q", required)
			}
		}
	})

	t.Run("round5 retired member six-key ordering", func(t *testing.T) {
		for _, stale := range []string{
			"(previous->>'postgres_system_id')::numeric,\n           (previous->>'timeline')::numeric,\n           (previous->>'database_oid')::numeric,",
			"(value->>'postgres_system_id')::numeric,\n           (value->>'timeline')::numeric,\n           (value->>'database_oid')::numeric,",
		} {
			if strings.Contains(insert, stale) {
				t.Error("retired_members ordering incorrectly includes timeline")
			}
		}
		for _, required := range []string{
			"(previous->>'postgres_system_id')::numeric,\n           (previous->>'database_oid')::numeric,",
			"(value->>'postgres_system_id')::numeric,\n           (value->>'database_oid')::numeric,",
		} {
			if !strings.Contains(insert, required) {
				t.Error("retired_members ordering lacks the exact six-key adjacency")
			}
		}
	})
}

func task8AuthorityV7RolePreflightStatement(t *testing.T, source string) string {
	t.Helper()
	segments := strings.Split(source, "-- talenro:statement")
	if len(segments) < 2 {
		t.Fatal("authority-v7 Up asset lacks a statement-delimited role preflight")
	}
	if executablePreamble := strings.TrimSpace(task8StripSQLComments(segments[0])); executablePreamble != "" {
		t.Fatalf("authority-v7 Up asset has executable SQL before its role preflight: %q", executablePreamble)
	}
	statement := strings.TrimSpace(segments[1])
	if !strings.HasPrefix(strings.ToLower(statement), "do $roles$") {
		t.Fatalf("authority-v7 first executable statement is not the role preflight: %.80q", statement)
	}
	return statement
}

func task8AuthorityV7PostflightStatement(t *testing.T, source string) string {
	t.Helper()
	lower := strings.ToLower(source)
	const openMarker = "do $postflight$"
	const closeMarker = "$postflight$;"
	if got := strings.Count(lower, openMarker); got != 1 {
		t.Fatalf("authority-v7 postflight opening marker count=%d, want 1", got)
	}
	if got := strings.Count(lower, closeMarker); got != 1 {
		t.Fatalf("authority-v7 postflight closing marker count=%d, want 1", got)
	}
	start := strings.Index(lower, openMarker)
	closeRelative := strings.Index(lower[start+len(openMarker):], closeMarker)
	if closeRelative < 0 {
		t.Fatal("authority-v7 postflight is not closed")
	}
	end := start + len(openMarker) + closeRelative + len(closeMarker)
	return source[start:end]
}

func TestNodeControlV7ProofLifecycleDDL(t *testing.T) {
	rawUp := string(task8ReadFile(t, "../../db/migrations/assets/nodecontrol_authority_v7_up.sql"))
	rawDown := string(task8ReadFile(t, "../../db/migrations/assets/nodecontrol_authority_v7_down.sql"))
	up := strings.ToLower(task8StripSQLComments(rawUp))
	down := strings.ToLower(task8StripSQLComments(rawDown))
	validator := task8SQLFunction(t, up, "nodecontrol.v7_authority_proof_group_valid")
	guard := task8SQLFunction(t, up, "nodecontrol.v7_guard_authority_proof_transition")
	sourceGuard := task8SQLFunction(t, up, "nodecontrol.v7_assert_source_writable")
	fenceGuard := task8SQLFunction(t, up, "nodecontrol.enforce_authority_fence_update")
	incidentGuard := task8SQLFunction(t, up, "nodecontrol.enforce_security_incident_workflow")

	for _, required := range []string{
		"select coalesce(",
		"effect_reason in ('none','failed','superseded','activation_deadline_expired','validation_rejected')",
		"effect_reason <> 'none'",
		"effect_reason = 'superseded'",
		"effect_reason <> 'superseded'",
		"activation_deadline < attestation_expires_at",
	} {
		if strings.Count(validator, required) != 1 {
			t.Errorf("proof shape validator lifecycle token %q count=%d, want 1", required, strings.Count(validator, required))
		}
	}
	if opens, closes := strings.Count(validator, "("), strings.Count(validator, ")"); opens != closes {
		t.Fatalf("proof shape validator parenthesis balance=%d/%d, want exact equality", opens, closes)
	}
	for _, required := range []string{
		"if tg_op = 'insert' then",
		"if tg_op = 'delete' then",
		"return old",
		"authority proof delete requires elapsed owner retention",
		"domain-prepared authority owner",
		"commitment-only authority proof",
		"authority proof commitment is immutable",
		"terminal authority proof requires its exact committed fence digest",
		"authority proof/fence scope digest mismatch",
		"authority proof mutation requires read committed isolation",
		"serial_bytes",
		"signature_verified_at",
		"published_envelope_digest",
		"authority_effect_disposition",
		"only one authority proof group may advance per row mutation",
		"row_data->>'signing_kind'",
		"new_row - allowed_columns",
	} {
		if !strings.Contains(guard, required) {
			t.Errorf("proof transition guard lacks %q", required)
		}
	}
	if strings.Contains(guard, "intent_kind") {
		t.Fatal("proof transition guard reads nonexistent intent_kind instead of signing_kind")
	}
	for _, required := range []string{
		"if tg_level = 'statement' then",
		"return null",
		"pg_advisory_xact_lock_shared",
		"nodecontrol:v7-source-freeze",
	} {
		if !strings.Contains(sourceGuard, required) {
			t.Errorf("v7 source guard statement preflight lacks %q", required)
		}
	}
	if strings.Contains(sourceGuard, "new.identity_state") {
		t.Fatal("v7 source guard directly dereferences identity_state on heterogeneous trigger rows")
	}
	if count := strings.Count(sourceGuard, "pg_catalog.to_jsonb(new)->>'identity_state'"); count != 2 {
		t.Fatalf("v7 source guard safe identity_state projection count=%d, want 2", count)
	}
	for _, required := range []string{
		"new claim-v1 authority fence must begin as an exact unbound reservation",
		"new legacy_v6 authority fences are forbidden after v7 activation",
		"authority fence bind requires exactly one commitment-only owner",
		"authority fence effect digest does not match owner commitment",
		"authority fence bind and finalize cannot occur in one proof transition",
		"authority fence claim or abort requires no proof owner",
		"authority fence mutation requires read committed isolation",
	} {
		if !strings.Contains(fenceGuard, required) {
			t.Errorf("v7 authority fence guard lacks %q", required)
		}
	}
	for _, required := range []string{
		"claim_v1", "reserved", "fence_pending", "security incident opening requires its exact reserved claim-v1 prepared proof",
	} {
		if !strings.Contains(incidentGuard, required) {
			t.Errorf("v7 security incident workflow lacks %q", required)
		}
	}
	if strings.Count(up, "create or replace function nodecontrol.enforce_authority_fence_update()") != 1 ||
		strings.Count(down, "create or replace function nodecontrol.enforce_authority_fence_update()") != 1 {
		t.Fatal("authority fence workflow replacement/restoration must occur exactly once in Up and Down")
	}
	if strings.Count(up, "create or replace function nodecontrol.enforce_security_incident_workflow()") != 1 ||
		strings.Count(down, "create or replace function nodecontrol.enforce_security_incident_workflow()") != 1 {
		t.Fatal("security incident workflow replacement/restoration must occur exactly once in Up and Down")
	}
	for _, required := range []string{
		"create constraint trigger ncv7_fence_owner_closure",
		"after update on nodecontrol.control_plane_authority_fences",
		"deferrable initially deferred",
		"execute function nodecontrol.v7_assert_activation_barrier()",
	} {
		if !strings.Contains(strings.Join(strings.Fields(up), " "), required) {
			t.Errorf("v7 deferred fence closure lacks %q", required)
		}
	}
	if !strings.Contains(down, "drop trigger ncv7_fence_owner_closure on nodecontrol.control_plane_authority_fences") {
		t.Fatal("v7 Down does not remove the deferred fence owner closure")
	}
	normalizedUp := strings.Join(strings.Fields(up), " ")
	normalizedDown := strings.Join(strings.Fields(down), " ")
	for _, required := range []string{
		"create trigger ncv7_00_source_lock before insert or update or delete on nodecontrol.%i for each statement execute function nodecontrol.v7_assert_source_writable()",
		"create trigger ncv7_01_source_guard before insert or update or delete on nodecontrol.%i for each row execute function nodecontrol.v7_assert_source_writable()",
		"create trigger ncv7_00_fence_source_lock before insert or update or delete on nodecontrol.control_plane_authority_fences for each statement execute function nodecontrol.v7_assert_source_writable()",
	} {
		if !strings.Contains(normalizedUp, required) {
			t.Errorf("v7 source lock topology lacks %q", required)
		}
	}
	for _, required := range []string{
		"drop trigger ncv7_00_source_lock on nodecontrol.%i",
		"drop trigger ncv7_01_source_guard on nodecontrol.%i",
		"drop trigger ncv7_00_fence_source_lock on nodecontrol.control_plane_authority_fences",
	} {
		if !strings.Contains(normalizedDown, required) {
			t.Errorf("v7 Down source lock topology lacks %q", required)
		}
	}
	wantSourceTables := make([]string, 0, len(task8BaseTables)-1)
	for _, table := range task8BaseTables {
		if table != "control_plane_authority_fences" {
			wantSourceTables = append(wantSourceTables, table)
		}
	}
	sort.Strings(wantSourceTables)
	sourceLoopPattern := regexp.MustCompile(`(?s)do\s+\$source_guards\$.*?foreach\s+relation_name\s+in\s+array\s+array\[(.*?)\]\s+loop`)
	quotedRelationPattern := regexp.MustCompile(`'([^']+)'`)
	for _, source := range []struct {
		label string
		sql   string
	}{{"Up", up}, {"Down", down}} {
		match := sourceLoopPattern.FindStringSubmatch(source.sql)
		if len(match) != 2 {
			t.Fatalf("v7 %s source guard loop is not uniquely parseable", source.label)
		}
		matches := quotedRelationPattern.FindAllStringSubmatch(match[1], -1)
		got := make([]string, 0, len(matches))
		seen := make(map[string]bool, len(matches))
		for _, relation := range matches {
			if seen[relation[1]] {
				t.Errorf("v7 %s source guard loop duplicates %s", source.label, relation[1])
			}
			seen[relation[1]] = true
			got = append(got, relation[1])
		}
		sort.Strings(got)
		if !reflect.DeepEqual(got, wantSourceTables) {
			t.Errorf("v7 %s source guard loop=%v, want exact 24-table base-source registry %v", source.label, got, wantSourceTables)
		}
	}
	if !strings.Contains(strings.Join(strings.Fields(up), " "), "add column authority_protocol_profile text collate \"c\" not null default 'legacy_v6'") ||
		!strings.Contains(strings.Join(strings.Fields(up), " "), "alter column authority_protocol_profile drop default") {
		t.Fatal("v7 fence profile backfill must use ADD COLUMN NOT NULL DEFAULT followed by DROP DEFAULT")
	}
	if strings.Contains(up, "update nodecontrol.control_plane_authority_fences set authority_protocol_profile = 'legacy_v6'") {
		t.Fatal("v7 fence profile backfill performs an ordinary UPDATE through the legacy trigger")
	}
	if strings.Contains(up, "'fresh_restore'") || !strings.Contains(up, "'fresh_restore_target'") {
		t.Fatal("v7 activation mode registry must use fresh_restore_target, never fresh_restore")
	}
	for _, required := range []string{
		"authority_effect_disposition is null",
		"authority_effect_disposition in ('applied','not_applied')",
	} {
		if !strings.Contains(up, required) {
			t.Errorf("v7 state transition three-phase disposition constraint lacks %q", required)
		}
	}
	if got := strings.Count(up, "before insert or update or delete on nodecontrol.%i for each row execute function nodecontrol.v7_guard_authority_proof_transition()"); got != 1 {
		t.Fatalf("proof trigger dynamic BEFORE INSERT OR UPDATE OR DELETE registry count=%d, want 1", got)
	}
	for _, relation := range []string{"node_resource_envelopes"} {
		replacement := "drop trigger " + relation + "_immutable on nodecontrol." + relation + "; create trigger " + relation + "_immutable before delete on nodecontrol." + relation
		if !strings.Contains(strings.Join(strings.Fields(up), " "), replacement) {
			t.Errorf("v7 Up lacks exact DELETE-only replacement for %s immutable trigger", relation)
		}
		restore := "drop trigger " + relation + "_immutable on nodecontrol." + relation + "; create trigger " + relation + "_immutable before update or delete on nodecontrol." + relation
		if !strings.Contains(strings.Join(strings.Fields(down), " "), restore) {
			t.Errorf("v7 Down lacks exact base restoration for %s immutable trigger", relation)
		}
	}
	if normalizedUp := strings.Join(strings.Fields(up), " "); !strings.Contains(normalizedUp, "drop trigger node_state_transitions_immutable on nodecontrol.node_state_transitions;") ||
		strings.Contains(normalizedUp, "create trigger node_state_transitions_immutable") {
		t.Error("v7 Up must remove the v6 node_state_transitions immutable trigger and leave DELETE authority solely to the v7 proof guard")
	}
	restore := "create trigger node_state_transitions_immutable before update or delete on nodecontrol.node_state_transitions"
	if !strings.Contains(strings.Join(strings.Fields(down), " "), restore) {
		t.Error("v7 Down lacks exact base restoration for node_state_transitions immutable trigger")
	}
	if strings.Contains(strings.Join(strings.Fields(down), " "), "drop trigger node_state_transitions_immutable on nodecontrol.node_state_transitions") {
		t.Error("v7 Down must not drop the node_state_transitions immutable trigger removed by v7 Up before recreating its v6 shape")
	}

	for _, constraint := range []string{
		"ncv7_grant_claim_authority_tuple_ck", "ncv7_grant_consumption_result_ck", "ncv7_grant_claim_lifecycle_ck",
		"ncv7_certificate_revoke_authority_tuple_ck", "ncv7_certificate_revoke_result_ck", "ncv7_certificate_revoke_lifecycle_ck",
		"ncv7_incident_resolution_authority_tuple_ck", "ncv7_incident_resolution_result_ck", "ncv7_incident_resolution_lifecycle_ck",
	} {
		if strings.Count(up, constraint) != 1 || strings.Count(down, constraint) != 1 {
			t.Errorf("split lifecycle constraint %s Up/Down count=%d/%d, want 1/1", constraint, strings.Count(up, constraint), strings.Count(down, constraint))
		}
	}
	for _, baseConstraint := range []string{
		"add constraint node_enrollment_grants_consumption_all_or_none check (num_nonnulls(consumed_at,consumption_attempt_id,consumption_request_digest,claim_authority_operation_id,claim_authority_epoch,claim_authority_sequence,result_issuance_id) in (0,7))",
		"add constraint node_certificates_revoke_all_or_none check (num_nonnulls(revoke_authority_operation_id,revoke_authority_epoch,revoke_authority_sequence,revoked_at,revoke_reason) in (0,5))",
		"add constraint node_security_incidents_resolution_all_or_none check (num_nonnulls(resolution_authority_operation_id,resolution_authority_epoch,resolution_authority_sequence,remediation_digest,resolution_at) in (0,5))",
	} {
		if !strings.Contains(down, baseConstraint) {
			t.Errorf("v7 Down lacks exact base constraint restoration %q", baseConstraint)
		}
	}
	if strings.Count(up, "create or replace function nodecontrol.enforce_enrollment_grant_workflow()") != 1 ||
		strings.Count(down, "create or replace function nodecontrol.enforce_enrollment_grant_workflow()") != 1 {
		t.Fatal("enrollment workflow replacement/restoration must occur exactly once in Up and Down")
	}
	if !strings.Contains(up, "old.result_issuance_id,old.expired_at") || strings.Contains(up, "old.claim_authority_operation_id,old.claim_authority_epoch,old.claim_authority_sequence,\n                  old.result_issuance_id") {
		t.Fatal("v7 enrollment workflow still treats a prepared claim binding as a terminal result")
	}
	if !strings.Contains(down, "old.claim_authority_operation_id,old.claim_authority_epoch,old.claim_authority_sequence") {
		t.Fatal("v7 Down does not restore the base enrollment terminal tuple")
	}
}

func task8AssertV7RequireRoleStructuralOracle(t *testing.T, source string) {
	t.Helper()
	normalized := strings.Join(strings.Fields(strings.ToLower(task8StripSQLComments(source))), " ")
	executable := strings.ToLower(task8StripSQLCommentsAndStrings(source))
	executableNormalized := strings.Join(strings.Fields(executable), " ")

	const primaryGate = "if required_role::text not in ('nodecontrol_upgrade_executor', 'nodecontrol_migration_downgrader', 'nodecontrol_staging_importer') or current_user is distinct from required_role::text then raise exception 'nodecontrol v7 role % is required', required_role using errcode = '42501'; end if;"
	const bootstrapLookup = "select oid into bootstrap_oid from pg_catalog.pg_roles where rolname = session_user and rolsuper;"
	const databaseOwnerFact = "select 1 from pg_catalog.pg_database where datname = current_database() and datdba = bootstrap_oid"
	const schemaOwnerFact = "select 1 from pg_catalog.pg_namespace where nspname = 'nodecontrol' and nspowner = bootstrap_oid"
	const bootstrapRejection = "raise exception 'authority v7 down requires the exact bootstrap owner' using errcode = '42501';"
	for label, required := range map[string]string{
		"literal current_user capability gate": primaryGate,
		"bootstrap superuser lookup":           bootstrapLookup,
		"current database ownership":           databaseOwnerFact,
		"nodecontrol schema ownership":         schemaOwnerFact,
		"exact bootstrap rejection":            bootstrapRejection,
	} {
		if got := strings.Count(normalized, required); got != 1 {
			t.Errorf("v7_require_role %s count=%d, want exact 1", label, got)
		}
	}
	primaryIndex := strings.Index(normalized, primaryGate)
	bootstrapIndex := strings.Index(normalized, bootstrapLookup)
	databaseIndex := strings.Index(normalized, databaseOwnerFact)
	schemaIndex := strings.Index(normalized, schemaOwnerFact)
	rejectionIndex := strings.Index(normalized, bootstrapRejection)
	if primaryIndex < 0 || bootstrapIndex <= primaryIndex || databaseIndex <= bootstrapIndex || schemaIndex <= bootstrapIndex || rejectionIndex <= databaseIndex || rejectionIndex <= schemaIndex {
		t.Errorf("v7_require_role authorization order = primary:%d bootstrap:%d database:%d schema:%d rejection:%d", primaryIndex, bootstrapIndex, databaseIndex, schemaIndex, rejectionIndex)
	}

	for label, forbidden := range map[string]*regexp.Regexp{
		"membership authorization":                 regexp.MustCompile(`\bpg_has_role\s*\(`),
		"session_user left capability comparison":  regexp.MustCompile(`\bsession_user\s*(?:=|<>|!=|is\s+(?:not\s+)?distinct\s+from)\s*required_role(?:\s*::\s*text)?\b`),
		"session_user right capability comparison": regexp.MustCompile(`\brequired_role(?:\s*::\s*text)?\s*(?:=|<>|!=|is\s+(?:not\s+)?distinct\s+from)\s*session_user\b`),
	} {
		if forbidden.MatchString(executable) {
			t.Errorf("v7_require_role contains forbidden %s", label)
		}
	}
	for _, forbidden := range []string{"pg_catalog.pg_authid", "rolpassword"} {
		if strings.Contains(executableNormalized, forbidden) {
			t.Errorf("v7_require_role reads forbidden runtime secret catalog token %q", forbidden)
		}
	}

	membershipReferences := regexp.MustCompile(`\bpg_catalog\.pg_auth_members\b`).FindAllStringIndex(executable, -1)
	membershipAddition := regexp.MustCompile(`(?s)drift_count\s*:=\s*drift_count\s*\+\s*\(\s*select\s+count\(\*\)\s+from\s+pg_catalog\.pg_auth_members\s+as\s+membership`).FindStringIndex(executable)
	membershipError := "raise exception 'authority v7 capability role setting or membership drift' using errcode = '55000';"
	membershipErrorIndex := strings.Index(normalized, membershipError)
	if len(membershipReferences) != 1 || membershipAddition == nil {
		t.Errorf("v7_require_role membership-drift query = references:%d addition:%v, want exact 1/true", len(membershipReferences), membershipAddition != nil)
	}
	for _, direction := range []string{"membership.member in", "membership.roleid in"} {
		if got := strings.Count(executableNormalized, direction); got != 1 {
			t.Errorf("v7_require_role bidirectional membership closure %q count=%d, want 1", direction, got)
		}
	}
	if got := strings.Count(normalized, membershipError); got != 1 {
		t.Errorf("v7_require_role membership drift rejection count=%d, want 1", got)
	}
	if membershipAddition != nil && membershipErrorIndex <= membershipAddition[0] {
		t.Errorf("v7_require_role membership drift rejection order = query:%d rejection:%d", membershipAddition[0], membershipErrorIndex)
	}
}

func TestNodeControlV7DownRegistry(t *testing.T) {
	up := strings.ToLower(string(task8ReadFile(t, "../../db/migrations/assets/nodecontrol_authority_v7_up.sql")))
	down := strings.ToLower(string(task8ReadFile(t, "../../db/migrations/assets/nodecontrol_authority_v7_down.sql")))
	if strings.Contains(down, "cascade") {
		t.Fatal("v7 Down uses forbidden CASCADE")
	}
	for _, required := range []string{"v7_consume_down_guard", "drop role nodecontrol_upgrade_executor", "drop role nodecontrol_migration_downgrader", "drop role nodecontrol_staging_importer"} {
		if !strings.Contains(down, required) {
			t.Errorf("v7 Down registry lacks %q", required)
		}
	}

	authorizationTable := task8SQLCreateTable(t, task8StripSQLComments(up), "nodecontrol.control_plane_authority_protocol_downgrade_authorizations")
	if got := strings.Count(authorizationTable, "provider_retirement_set_body_jcs bytea not null"); got != 1 {
		t.Errorf("downgrade authorization provider-retirement body column count = %d, want exactly 1", got)
	}

	insertParameters := []string{
		"p_authorization_id uuid", "p_installation_id uuid", "p_migration_latch_digest bytea",
		"p_database_identity_digest bytea", "p_migration_version bigint", "p_current_catalog_digest bytea",
		"p_pristine_inventory_digest bytea", "p_provider_retirement_set_digest bytea", "p_environment_anchor_set_digest bytea",
		"p_database_transaction_id numeric", "p_transaction_nonce bytea", "p_authorization_scope text",
		"p_issued_at timestamptz", "p_expires_at timestamptz", "p_authorization_body_jcs bytea",
		"p_authorization_envelope_jcs bytea", "p_authorization_digest bytea", "p_pristine_inventory_body_jcs bytea",
		"p_provider_retirement_set_body_jcs bytea", "p_provider_retirement_evidence_jcs bytea",
	}
	consumeParameters := []string{
		"p_installation_id uuid", "p_migration_latch_digest bytea", "p_authorization_digest bytea",
		"p_transaction_nonce bytea", "p_authorized_state_body_jcs bytea", "p_authorized_state_digest bytea",
	}
	insertResults := []task8SQLResultColumn{
		{"pristine_downgrade_inventory_digest", "bytea"}, {"migration_latch_digest", "bytea"}, {"downgrade_authorization_digest", "bytea"},
		{"provider_protocol_downgrade_retirement_set_digest", "bytea"}, {"database_transaction_id", "numeric"}, {"transaction_nonce", "bytea"},
		{"migration_latch_count", "bigint"}, {"downgrade_authorization_count", "bigint"}, {"non_control_protocol_row_count", "bigint"}, {"stage", "text"}, {"observed_at", "timestamptz"},
	}
	consumeResults := []task8SQLResultColumn{
		{"pristine_downgrade_inventory_digest", "bytea"}, {"authorized_state_digest", "bytea"}, {"consumed_migration_latch_digest", "bytea"},
		{"consumed_downgrade_authorization_digest", "bytea"}, {"database_transaction_id", "numeric"}, {"transaction_nonce", "bytea"},
		{"post_consume_migration_latch_count", "bigint"}, {"post_consume_downgrade_authorization_count", "bigint"},
		{"non_control_protocol_row_count", "bigint"}, {"stage", "text"}, {"consumed_at", "timestamptz"},
	}
	insert := task8SQLFunction(t, up, "nodecontrol.v7_insert_downgrade_authorization")
	consume := task8SQLFunction(t, up, "nodecontrol.v7_consume_down_guard")
	if got := task8SQLFunctionParameters(t, insert, "nodecontrol.v7_insert_downgrade_authorization"); !reflect.DeepEqual(got, insertParameters) {
		t.Errorf("downgrade authorization input registry = %#v, want exact 20-input registry %#v", got, insertParameters)
	}
	if got := task8SQLFunctionResults(t, insert); !reflect.DeepEqual(got, insertResults) {
		t.Errorf("downgrade authorization result registry = %#v, want exact 11-column registry %#v", got, insertResults)
	}
	if got := task8SQLFunctionParameters(t, consume, "nodecontrol.v7_consume_down_guard"); !reflect.DeepEqual(got, consumeParameters) {
		t.Errorf("downgrade consume input registry = %#v, want exact 6-input registry %#v", got, consumeParameters)
	}
	if got := task8SQLFunctionResults(t, consume); !reflect.DeepEqual(got, consumeResults) {
		t.Errorf("downgrade consume result registry = %#v, want exact 11-column registry %#v", got, consumeResults)
	}

	// Extract every executable table/classification tuple from each function.
	// Exact slice equality rejects omissions, additions, reordering, duplicates,
	// and a tuple that exists only in comments or a different function.
	for name, functionBody := range map[string]string{"insert": insert, "consume": consume} {
		if got := task8SQLPristineRegistry(functionBody); !reflect.DeepEqual(got, task8PristineDownRegistry) {
			t.Errorf("%s pristine registry = %#v, want exact 49 literal tuples %#v", name, got, task8PristineDownRegistry)
		}
	}

	// Strip both comments and SQL string literals before checking the lock
	// sequence. This catches the existing production break where the only
	// ACCESS EXCLUSIVE words are commentary or dynamically inert text.
	lockOrder := []string{
		"nodecontrol.control_plane_authority_protocol_migration_latches",
		"nodecontrol.control_plane_authority_protocol_downgrade_authorizations",
	}
	for _, pair := range task8PristineDownRegistry {
		if pair.classification == "authority_v7_non_control" {
			lockOrder = append(lockOrder, pair.table)
		}
	}
	for _, pair := range task8PristineDownRegistry {
		if pair.classification == "base_v6" {
			lockOrder = append(lockOrder, pair.table)
		}
	}
	lockOrder = append(lockOrder, "public.goose_db_version")
	for name, functionBody := range map[string]string{"insert": insert, "consume": consume} {
		if got := task8SQLAccessExclusiveLocks(functionBody); !reflect.DeepEqual(got, lockOrder) {
			t.Errorf("%s executable ACCESS EXCLUSIVE lock registry = %#v, want exact ordered set %#v", name, got, lockOrder)
		}
	}

	for _, signature := range []string{
		"drop function nodecontrol.v7_acquire_source_freeze_for_seal()",
		"drop function nodecontrol.v7_insert_downgrade_authorization(uuid,uuid,bytea,bytea,bigint,bytea,bytea,bytea,bytea,numeric,bytea,text,timestamptz,timestamptz,bytea,bytea,bytea,bytea,bytea,bytea)",
		"drop function nodecontrol.v7_consume_down_guard(uuid,bytea,bytea,bytea,bytea,bytea)",
	} {
		if !strings.Contains(task8StripSQLCommentsAndStrings(down), signature) {
			t.Errorf("v7 Down lacks exact revised signature %q", signature)
		}
	}

	reverseDDL := task8StripSQLComments(down)
	wantReverseDDLOrder := []string{
		"drop function nodecontrol.begin_staging_import(bytea,bytea,bytea,bytea,bytea,bytea,bytea,jsonb)",
		"drop function nodecontrol.v7_consume_down_guard(uuid,bytea,bytea,bytea,bytea,bytea)",
		"drop function nodecontrol.v7_insert_downgrade_authorization(uuid,uuid,bytea,bytea,bigint,bytea,bytea,bytea,bytea,numeric,bytea,text,timestamptz,timestamptz,bytea,bytea,bytea,bytea,bytea,bytea)",
		"drop function nodecontrol.v7_acquire_source_freeze_for_seal()",
		"drop trigger ncv7_fence_owner_closure on nodecontrol.control_plane_authority_fences",
		"drop trigger ncv7_00_fence_source_lock on nodecontrol.control_plane_authority_fences",
		"execute format('drop trigger ncv7_01_source_guard on nodecontrol.%i', relation_name)",
		"execute format('drop trigger ncv7_00_source_lock on nodecontrol.%i', relation_name)",
		"array['node_root_metadata_publish_intents','node_state_signing_intents','node_resource_envelopes','node_security_incidents','node_state_transitions','node_certificates','node_certificate_issuances','node_enrollment_grants']",
		"drop constraint ncv7_fi_fk01",
		"drop constraint ncv7_sa_fk02, drop constraint ncv7_sa_fk01",
		"drop constraint ncv7_ra_fk02, drop constraint ncv7_ra_fk01",
		"drop constraint ncv7_pd_fk01",
		"drop constraint ncv7_et_fk03, drop constraint ncv7_et_fk02, drop constraint ncv7_et_fk01",
		"drop constraint ncv7_ec_fk02, drop constraint ncv7_ec_fk01",
		"drop constraint ncv7_er_fk02, drop constraint ncv7_er_fk01",
		"drop constraint ncv7_ea_fk01",
		"drop constraint ncv7_ei_fk01",
		"drop constraint ncv7_pr_fk02, drop constraint ncv7_pr_fk01",
		"drop constraint ncv7_pc_fk01",
		"drop constraint ncv7_pa_fk03, drop constraint ncv7_pa_fk02, drop constraint ncv7_pa_fk01",
		"drop constraint ncv7_ua_fk03, drop constraint ncv7_ua_fk02, drop constraint ncv7_ua_fk01",
		"drop constraint ncv7_rb_fk02, drop constraint ncv7_rb_fk01",
		"drop constraint ncv7_rr_fk01",
		"relation_names text[] := array['node_root_metadata_publish_intents','node_state_signing_intents','node_resource_envelopes','node_security_incidents','node_security_incidents','node_state_transitions','node_certificates','node_certificate_issuances','node_enrollment_grants','node_enrollment_grants']",
	}
	lastReverse := -1
	for _, needle := range wantReverseDDLOrder {
		index := strings.Index(reverseDDL, needle)
		if index < 0 {
			t.Errorf("v7 Down lacks exact reverse-DDL step %q", needle)
			continue
		}
		if index <= lastReverse {
			t.Errorf("v7 Down reverse-DDL order is not exact at %q", needle)
		}
		lastReverse = index
	}

	wantDropOrder := []string{"fresh_restore_import_applications", "staging_import_capability_revocation_applications", "staging_import_capability_recovery_applications", "staging_import_capability_recovery_intents", "staging_import_capabilities", "protocol_activation_releases", "protocol_activation_completions", "epoch_transition_recovery_applications", "epoch_transition_recovery_prefix_decisions", "epoch_transition_recovery_intents", "epoch_transition_terminal_applications", "epoch_transition_cancellations", "epoch_transition_resolutions", "epoch_transition_applications", "epoch_transition_intents", "runtime_rebind_results", "protocol_activations", "protocol_upgrade_attempts", "runtime_registration_results", "fresh_restore_requirements", "legacy_source_seals", "indeterminate_source_seals", "legacy_database_source_retirements", "protocol_upgrade_intents", "protocol_downgrade_authorizations", "protocol_migration_latches"}
	last := -1
	for _, suffix := range wantDropOrder {
		needle := "drop table nodecontrol.control_plane_authority_" + suffix
		index := strings.Index(down, needle)
		if index < 0 {
			t.Errorf("v7 Down lacks %s", needle)
			continue
		}
		if index <= last {
			t.Errorf("v7 Down drop order is not exact at %s", suffix)
		}
		last = index
	}
	for _, role := range []string{"nodecontrol_staging_importer", "nodecontrol_migration_downgrader", "nodecontrol_upgrade_executor"} {
		needle := "drop role " + role
		index := strings.Index(down, needle)
		if index <= last {
			t.Errorf("v7 Down role drop order is not exact at %s", role)
		}
		last = index
	}
}

var task8PristineDownRegistry = []struct {
	table          string
	classification string
}{
	{"nodecontrol.control_plane_authority_epoch_transition_applications", "authority_v7_non_control"},
	{"nodecontrol.control_plane_authority_epoch_transition_cancellations", "authority_v7_non_control"},
	{"nodecontrol.control_plane_authority_epoch_transition_intents", "authority_v7_non_control"},
	{"nodecontrol.control_plane_authority_epoch_transition_recovery_applications", "authority_v7_non_control"},
	{"nodecontrol.control_plane_authority_epoch_transition_recovery_intents", "authority_v7_non_control"},
	{"nodecontrol.control_plane_authority_epoch_transition_recovery_prefix_decisions", "authority_v7_non_control"},
	{"nodecontrol.control_plane_authority_epoch_transition_resolutions", "authority_v7_non_control"},
	{"nodecontrol.control_plane_authority_epoch_transition_terminal_applications", "authority_v7_non_control"},
	{"nodecontrol.control_plane_authority_fences", "base_v6"},
	{"nodecontrol.control_plane_authority_fresh_restore_import_applications", "authority_v7_non_control"},
	{"nodecontrol.control_plane_authority_fresh_restore_requirements", "authority_v7_non_control"},
	{"nodecontrol.control_plane_authority_indeterminate_source_seals", "authority_v7_non_control"},
	{"nodecontrol.control_plane_authority_legacy_database_source_retirements", "authority_v7_non_control"},
	{"nodecontrol.control_plane_authority_legacy_source_seals", "authority_v7_non_control"},
	{"nodecontrol.control_plane_authority_protocol_activation_completions", "authority_v7_non_control"},
	{"nodecontrol.control_plane_authority_protocol_activation_releases", "authority_v7_non_control"},
	{"nodecontrol.control_plane_authority_protocol_activations", "authority_v7_non_control"},
	{"nodecontrol.control_plane_authority_protocol_upgrade_attempts", "authority_v7_non_control"},
	{"nodecontrol.control_plane_authority_protocol_upgrade_intents", "authority_v7_non_control"},
	{"nodecontrol.control_plane_authority_runtime_rebind_results", "authority_v7_non_control"},
	{"nodecontrol.control_plane_authority_runtime_registration_results", "authority_v7_non_control"},
	{"nodecontrol.control_plane_authority_staging_import_capabilities", "authority_v7_non_control"},
	{"nodecontrol.control_plane_authority_staging_import_capability_recovery_applications", "authority_v7_non_control"},
	{"nodecontrol.control_plane_authority_staging_import_capability_recovery_intents", "authority_v7_non_control"},
	{"nodecontrol.control_plane_authority_staging_import_capability_revocation_applications", "authority_v7_non_control"},
	{"nodecontrol.control_plane_trust_bundle_high_waters", "base_v6"},
	{"nodecontrol.node_capacity_profiles", "base_v6"},
	{"nodecontrol.node_certificate_issuances", "base_v6"},
	{"nodecontrol.node_certificates", "base_v6"},
	{"nodecontrol.node_desired_states", "base_v6"},
	{"nodecontrol.node_endpoints", "base_v6"},
	{"nodecontrol.node_enrollment_grants", "base_v6"},
	{"nodecontrol.node_failure_domain_membership", "base_v6"},
	{"nodecontrol.node_failure_domains", "base_v6"},
	{"nodecontrol.node_inventory", "base_v6"},
	{"nodecontrol.node_observed_states", "base_v6"},
	{"nodecontrol.node_operator_audit", "base_v6"},
	{"nodecontrol.node_pops", "base_v6"},
	{"nodecontrol.node_process_slots", "base_v6"},
	{"nodecontrol.node_recovery_sessions", "base_v6"},
	{"nodecontrol.node_recovery_states", "base_v6"},
	{"nodecontrol.node_resource_envelopes", "base_v6"},
	{"nodecontrol.node_restore_reauthorization_approvals", "base_v6"},
	{"nodecontrol.node_root_metadata_publish_intents", "base_v6"},
	{"nodecontrol.node_root_metadata_signature_shares", "base_v6"},
	{"nodecontrol.node_security_fault_receipts", "base_v6"},
	{"nodecontrol.node_security_incidents", "base_v6"},
	{"nodecontrol.node_state_signing_intents", "base_v6"},
	{"nodecontrol.node_state_transitions", "base_v6"},
}

func task8SQLFunction(t *testing.T, source, name string) string {
	t.Helper()
	marker := "create function " + name + "("
	start := strings.Index(source, marker)
	if start < 0 {
		marker = "create or replace function " + name + "("
		start = strings.Index(source, marker)
	}
	if start < 0 {
		t.Fatalf("SQL lacks function %s", name)
	}
	rest := source[start:]
	end := strings.Index(rest, "$fn$;")
	if end < 0 {
		t.Fatalf("SQL function %s lacks terminal $fn$", name)
	}
	return rest[:end+len("$fn$;")]
}

func task8SQLFunctionParameters(t *testing.T, functionBody, name string) []string {
	t.Helper()
	marker := "create function " + name
	start := strings.Index(functionBody, marker)
	if start < 0 {
		t.Fatalf("SQL function body lacks declaration %s", name)
	}
	open := strings.Index(functionBody[start+len(marker):], "(")
	if open < 0 {
		t.Fatalf("SQL function %s lacks parameter list", name)
	}
	open += start + len(marker)
	close := task8MatchingSQLParen(t, functionBody, open)
	return task8SplitSQLRegistry(functionBody[open+1 : close])
}

type task8SQLResultColumn struct {
	name     string
	typeName string
}

func task8SQLFunctionResults(t *testing.T, functionBody string) []task8SQLResultColumn {
	t.Helper()
	marker := "returns table"
	start := strings.Index(functionBody, marker)
	if start < 0 {
		t.Errorf("SQL function lacks RETURNS TABLE")
		return nil
	}
	open := strings.Index(functionBody[start+len(marker):], "(")
	if open < 0 {
		t.Error("SQL RETURNS TABLE lacks result registry")
		return nil
	}
	open += start + len(marker)
	close := task8MatchingSQLParen(t, functionBody, open)
	entries := task8SplitSQLRegistry(functionBody[open+1 : close])
	result := make([]task8SQLResultColumn, 0, len(entries))
	for _, entry := range entries {
		fields := strings.Fields(entry)
		if len(fields) != 2 {
			t.Fatalf("SQL result entry %q must contain exactly name and type", entry)
		}
		result = append(result, task8SQLResultColumn{name: fields[0], typeName: fields[1]})
	}
	return result
}

func task8SQLCreateTable(t *testing.T, source, name string) string {
	t.Helper()
	marker := "create table " + name
	start := strings.Index(source, marker)
	if start < 0 {
		t.Fatalf("SQL lacks table %s", name)
	}
	openRelative := strings.Index(source[start+len(marker):], "(")
	if openRelative < 0 {
		t.Fatalf("SQL table %s lacks column registry", name)
	}
	open := start + len(marker) + openRelative
	close := task8MatchingSQLParen(t, source, open)
	return source[start : close+1]
}

func task8MatchingSQLParen(t *testing.T, source string, open int) int {
	t.Helper()
	depth := 0
	for index := open; index < len(source); index++ {
		switch source[index] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return index
			}
		}
	}
	t.Fatal("SQL registry has no matching closing parenthesis")
	return -1
}

func task8SplitSQLRegistry(source string) []string {
	parts := strings.Split(source, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		normalized := strings.Join(strings.Fields(part), " ")
		if normalized != "" {
			result = append(result, normalized)
		}
	}
	return result
}

func task8SQLPristineRegistry(functionBody string) []struct {
	table          string
	classification string
} {
	executable := task8StripSQLComments(functionBody)
	pattern := regexp.MustCompile(`\(\s*'([^']+)'\s*,\s*'(authority_v7_non_control|base_v6)'\s*\)`)
	matches := pattern.FindAllStringSubmatch(executable, -1)
	result := make([]struct {
		table          string
		classification string
	}, 0, len(matches))
	for _, match := range matches {
		result = append(result, struct {
			table          string
			classification string
		}{match[1], match[2]})
	}
	return result
}

func task8SQLAccessExclusiveLocks(functionBody string) []string {
	executable := task8SQLFunctionExecutableBody(functionBody)
	pattern := regexp.MustCompile(`(?is)\block\s+table\s+(.+?)\s+in\s+access\s+exclusive\s+mode\b`)
	matches := pattern.FindAllStringSubmatch(executable, -1)
	result := make([]string, 0, 52)
	for _, match := range matches {
		for _, relation := range strings.Split(match[1], ",") {
			relation = strings.Join(strings.Fields(relation), " ")
			if relation != "" {
				result = append(result, relation)
			}
		}
	}
	return result
}

func task8SQLFunctionExecutableBody(functionBody string) string {
	// The outer $fn$ delimiter is syntax, not a dynamic string. Preserve only
	// its body, then remove comments, single-quoted values, and every nested
	// dollar-quoted string such as $sql$LOCK TABLE ...$sql$.
	first := strings.Index(functionBody, "$fn$")
	last := strings.LastIndex(functionBody, "$fn$")
	if first >= 0 && last > first {
		functionBody = functionBody[first+len("$fn$") : last]
	}
	executable := task8StripSQLCommentsAndStrings(functionBody)
	dollarTag := regexp.MustCompile(`\$[a-z_][a-z0-9_]*\$|\$\$`)
	for {
		location := dollarTag.FindStringIndex(executable)
		if location == nil {
			break
		}
		tag := executable[location[0]:location[1]]
		end := strings.Index(executable[location[1]:], tag)
		if end < 0 {
			executable = executable[:location[0]]
			break
		}
		end += location[1] + len(tag)
		executable = executable[:location[0]] + strings.Repeat(" ", end-location[0]) + executable[end:]
	}
	return executable
}

func task8StripSQLComments(source string) string {
	var result strings.Builder
	for index := 0; index < len(source); {
		switch {
		case index+1 < len(source) && source[index:index+2] == "--":
			for index < len(source) && source[index] != '\n' {
				index++
			}
		case index+1 < len(source) && source[index:index+2] == "/*":
			index += 2
			for index+1 < len(source) && source[index:index+2] != "*/" {
				index++
			}
			if index+1 < len(source) {
				index += 2
			}
		default:
			result.WriteByte(source[index])
			index++
		}
	}
	return result.String()
}

func task8StripSQLCommentsAndStrings(source string) string {
	source = task8StripSQLComments(source)
	var result strings.Builder
	for index := 0; index < len(source); {
		if source[index] != '\'' {
			result.WriteByte(source[index])
			index++
			continue
		}
		result.WriteByte(' ')
		index++
		for index < len(source) {
			if source[index] != '\'' {
				index++
				continue
			}
			if index+1 < len(source) && source[index+1] == '\'' {
				index += 2
				continue
			}
			index++
			break
		}
	}
	return result.String()
}

func TestAuthorityEffectProofColumnRegistry(t *testing.T) {
	up := strings.ToLower(string(task8ReadFile(t, "../../db/migrations/assets/nodecontrol_authority_v7_up.sql")))
	groups := map[string][]string{
		"node_enrollment_grants":             {"create_", "claim_"},
		"node_certificate_issuances":         {"activation_"},
		"node_certificates":                  {"revoke_"},
		"node_state_transitions":             {""},
		"node_security_incidents":            {"open_", "resolve_"},
		"node_resource_envelopes":            {"activation_"},
		"node_state_signing_intents":         {"activation_"},
		"node_root_metadata_publish_intents": {"activation_"},
	}
	suffixes := []string{
		"authority_effect_commitment_jcs", "authority_effect_commitment_digest", "authority_provider_head_jcs",
		"authority_provider_head_digest", "authority_checkpoint_anchor_jcs", "authority_checkpoint_anchor_digest",
		"authority_effect_reason", "authority_attestation_expires_at", "authority_activation_deadline",
		"authority_expected_provider_identity_digest", "authority_activation_evidence_jcs", "authority_activation_evidence_digest",
		"authority_effect_resolution_jcs", "authority_effect_resolution_digest",
	}
	for table, prefixes := range groups {
		if !strings.Contains(up, "alter table nodecontrol."+table) {
			t.Errorf("proof registry lacks owner table %s", table)
		}
		for _, prefix := range prefixes {
			for _, suffix := range suffixes {
				if !strings.Contains(up, prefix+suffix) {
					t.Errorf("proof registry lacks %s.%s%s", table, prefix, suffix)
				}
			}
		}
	}
	for _, kind := range []string{"grant_create", "grant_claim", "certificate_activate", "certificate_revoke", "identity_epoch_advance", "operator_transition", "security_incident_open", "security_incident_resolve", "resource_envelope_activate", "desired_activate", "recovery_activate", "root_publish", "metadata_publish"} {
		if !strings.Contains(up, "'"+kind+"'") {
			t.Errorf("proof kind registry lacks %s", kind)
		}
	}
	if strings.Contains(up, "authority_effect_outcomes") || strings.Contains(up, "generic_outcome") {
		t.Fatal("v7 DDL creates a forbidden generic authority outcome relation")
	}
}

func task8ReadFile(t *testing.T, path string) []byte {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return raw
}
