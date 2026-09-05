package contracts_test

import (
	"crypto/sha256"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"talenro.local/platform/internal/nodecontrol/contracts"
)

// TestAuthorityV7BoundaryDTOsCompileAndRemainClosed deliberately inspects the
// public boundary declaration instead of importing symbols that do not exist at
// the RED point. This keeps the initial failure semantic while independently
// freezing the seven public DTO registries.
func TestAuthorityV7BoundaryDTOsCompileAndRemainClosed(t *testing.T) {
	const boundaryPath = "authority_v7_boundary.go"
	raw, err := os.ReadFile(boundaryPath)
	if err != nil {
		t.Fatalf("Task 8 boundary is missing: %v", err)
	}
	file, err := parser.ParseFile(token.NewFileSet(), boundaryPath, raw, parser.AllErrors)
	if err != nil {
		t.Fatalf("parse Task 8 boundary: %v", err)
	}

	for _, spec := range file.Imports {
		pathValue, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(pathValue, "/authority") || strings.Contains(pathValue, "/store") || strings.Contains(pathValue, "/migrations") {
			t.Fatalf("boundary imports forbidden downstream package %q", pathValue)
		}
	}
	if strings.Contains(string(raw), "map[string]any") || strings.Contains(string(raw), "map[string]interface{}") {
		t.Fatal("boundary uses a permissive map payload")
	}

	want := map[string][]string{
		"AuthorityV7UpMigrationFactsV1": {
			"MigrationLatch", "MigrationLatchDigest", "UpgradeIntentOrNull", "UpgradeIntentDigestOrNull",
			"AuthorityProtocolProfile", "LocalRuntimeIsolationDigest", "TransactionNonce", "ExpiresAt",
		},
		"AuthorityV7DownMigrationFactsV1": {
			"MigrationLatch", "MigrationLatchDigest", "AuthorityProtocolProfile", "CurrentCatalogDigest", "ManifestID",
			"ManifestDigest", "StableTableCount", "StableTableInventory", "ProviderRetirementSet",
			"ProviderRetirementSetDigest", "EnvironmentAnchorSetDigest", "AuthorizationID", "AuthorizationScope",
			"TransactionNonce", "ExpiresAt",
		},
		"AuthorityV7DownAuthorizationRequestV1": {
			"AuthorizationID", "MigrationLatch", "MigrationLatchDigest", "CurrentCatalogDigest", "PristineInventory",
			"PristineInventoryDigest", "ProviderRetirementSet", "ProviderRetirementSetDigest", "EnvironmentAnchorSetDigest",
			"ActualDatabaseTransactionID", "TransactionNonce", "AuthorizationScope",
		},
		"AuthorityV7DownAuthorizationFactsV1": {
			"Request", "Authorization", "AuthorizationDigest", "AuthorizationEnvelopeJCS", "SignerRole", "SignerKeyID",
			"SignaturePolicyVersion", "TrustRootDigest", "SignatureAlgorithm",
		},
		"FreshRestoreImportProjectionInputV1": {
			"ManifestTopology", "StagingImportCapability", "StagingExclusionLease", "CurrentProviderHeadDigest",
			"ProviderPhase", "ServingLeaseAbsentDigest", "StagingExclusionState", "DatabaseRouteClosedDigest",
			"CurrentDatabaseIdentityDigest", "DatabaseTimelineLineageChainDigest", "CurrentDatabaseIncarnationRegistrationDigest",
			"LatestDatabaseAuthorityRebindResultDigestOrNull", "RuntimeRebindChainDigest", "RuntimeInstanceBindingDigest",
			"PreImportInventoryDigest", "NormalizedCatalogDigest", "AdmissionCheckedAt",
		},
		"FreshImportTopologyProjectionV1": {
			"ProjectionVersion", "TargetActivationID", "TargetDeploymentID", "TargetDatabaseIdentityDigest",
			"NormalizedCatalogDigest", "ObjectCount", "Objects",
		},
		"FreshRestoreImportApplicationV1": {
			"SingleUseApplyID", "StagingImportCapabilityDigest", "StagingImportCapabilityRecoveryIntentDigestOrNull",
			"StagingImportCapabilityRecoveryApplicationDigestOrNull", "ManifestDigest", "TargetActivationID",
			"CurrentDatabaseIdentityDigest", "DatabaseTimelineLineageChainDigest", "TargetDatabaseIncarnationRegistrationDigest",
			"RuntimeRebindChainDigest", "RuntimeInstanceBindingDigest", "StagingExclusionLeaseDigest",
			"AcquisitionLockedProviderHeadDigest", "DatabaseRouteClosedDigest", "PreImportInventoryDigest",
			"PostImportInventoryDigest", "ImportedObjectCount", "CompleteNodeSetDigest", "ForbiddenStateZeroDigest",
			"DatabaseTransactionID", "TransactionSnapshotDigest", "AppliedAt",
		},
	}

	got := task8StructFields(t, file)
	methods := task8MethodNames(file)
	for typeName, wantFields := range want {
		gotFields, exists := got[typeName]
		if !exists {
			t.Errorf("missing boundary DTO %s", typeName)
			continue
		}
		if !reflect.DeepEqual(gotFields, wantFields) {
			t.Errorf("%s fields = %v, want exact registry %v", typeName, gotFields, wantFields)
		}
		if !methods[typeName+".Validate"] && !methods["*"+typeName+".Validate"] {
			t.Errorf("%s lacks Validate() error", typeName)
		}
	}

	closedValues := []string{
		"production", "disposable_fixture", "legacy_v6", "claim_v1", "locked", "pending", "down_00007_only",
		"pre_authorization", "pop", "failure_domain_definition", "capacity_profile_definition", "node_reconstruction_seed",
		"fresh_v7_staging_closed", "held",
	}
	for _, value := range closedValues {
		if !strings.Contains(string(raw), strconv.Quote(value)) {
			t.Errorf("boundary lacks closed literal %q", value)
		}
	}
}

func TestAuthorityV7BoundaryValidationRejectsAmbiguousAndCrossBoundPayloads(t *testing.T) {
	now := time.Now().UTC()
	latch := contracts.AuthorityV7MigrationLatchFactsV1{
		InstallationID:         boundaryUUID("installation"),
		InstallationKind:       contracts.AuthorityV7InstallationKindDisposableFixture,
		MigrationVersion:       7,
		DatabaseIdentityDigest: boundaryDigest("database"),
		UpCatalogDigest:        boundaryDigest("catalog"),
		DownState:              contracts.AuthorityV7MigrationDownStateLocked,
		InstalledAt:            now.Add(-time.Minute),
	}
	validUp := contracts.AuthorityV7UpMigrationFactsV1{
		MigrationLatch:              latch,
		MigrationLatchDigest:        boundaryDigest("latch"),
		AuthorityProtocolProfile:    contracts.AuthorityV7ProtocolProfileLegacyV6,
		LocalRuntimeIsolationDigest: boundaryDigest("isolation"),
		TransactionNonce:            boundaryDigest("nonce"),
		ExpiresAt:                   now.Add(time.Minute),
	}
	if err := validUp.Validate(); err != nil {
		t.Fatalf("valid disposable Up: %v", err)
	}
	productionWithoutIntent := validUp
	productionWithoutIntent.MigrationLatch.InstallationKind = contracts.AuthorityV7InstallationKindProduction
	productionWithoutIntent.AuthorityProtocolProfile = contracts.AuthorityV7ProtocolProfileClaimV1
	if err := productionWithoutIntent.Validate(); !errors.Is(err, contracts.ErrInvalidAuthorityValue) {
		t.Fatalf("production Up without intent error = %v", err)
	}

	validObject := contracts.FreshRestoreImportManifestObjectV1{
		ObjectType:    contracts.FreshRestoreImportObjectCapacityProfileDefinition,
		CanonicalKey:  "standard.profile/7",
		Payload:       []byte(`{"profile_id":"standard.profile","version":7,"adapter":"fixture","egress_limit_bps":1000000,"connection_limit":1,"handshake_limit_per_second":1,"cpu_quota_millicores":100,"cpu_limit_basis_points":1,"memory_limit_bytes":67108864,"task_limit":32,"file_descriptor_limit":64,"queue_limit":1,"packet_loss_limit_basis_points":1,"required_metrics":["cpu_basis_points"]}`),
		PayloadDigest: boundaryDigest("capacity-payload"),
	}
	if err := validObject.Validate(); err != nil {
		t.Fatalf("valid textual-profile object: %v", err)
	}
	mutations := []struct {
		name    string
		payload string
	}{
		{name: "extra", payload: `{"profile_id":"standard.profile","version":7,"adapter":"fixture","egress_limit_bps":1000000,"connection_limit":1,"handshake_limit_per_second":1,"cpu_quota_millicores":100,"cpu_limit_basis_points":1,"memory_limit_bytes":67108864,"task_limit":32,"file_descriptor_limit":64,"queue_limit":1,"packet_loss_limit_basis_points":1,"required_metrics":["cpu_basis_points"],"extra":true}`},
		{name: "duplicate", payload: `{"profile_id":"standard.profile","profile_id":"other","version":7,"adapter":"fixture","egress_limit_bps":1000000,"connection_limit":1,"handshake_limit_per_second":1,"cpu_quota_millicores":100,"cpu_limit_basis_points":1,"memory_limit_bytes":67108864,"task_limit":32,"file_descriptor_limit":64,"queue_limit":1,"packet_loss_limit_basis_points":1,"required_metrics":["cpu_basis_points"]}`},
		{name: "null array", payload: `{"profile_id":"standard.profile","version":7,"adapter":"fixture","egress_limit_bps":1000000,"connection_limit":1,"handshake_limit_per_second":1,"cpu_quota_millicores":100,"cpu_limit_basis_points":1,"memory_limit_bytes":67108864,"task_limit":32,"file_descriptor_limit":64,"queue_limit":1,"packet_loss_limit_basis_points":1,"required_metrics":null}`},
	}
	for _, mutation := range mutations {
		t.Run(mutation.name, func(t *testing.T) {
			candidate := validObject
			candidate.Payload = []byte(mutation.payload)
			if err := candidate.Validate(); !errors.Is(err, contracts.ErrInvalidAuthorityValue) {
				t.Fatalf("error = %v, want ErrInvalidAuthorityValue", err)
			}
		})
	}

	normalizedNode := contracts.FreshImportTopologyProjectionObjectV1{
		ObjectType:        contracts.FreshRestoreImportObjectNodeReconstructionSeed,
		CanonicalKey:      boundaryUUID("node").String(),
		NormalizedPayload: []byte(`{"node_id":"` + boundaryUUID("node").String() + `","pop_code":"tbs-1","operator_state":"disabled","security_state":"quarantined","identity_state":"unauthorized","health_state":"unknown","identity_epoch":"0","inventory_version":"1","security_version":"1","next_desired_generation":"1","next_recovery_generation":"1","resume_operator_state_or_null":null,"pending_operator_transition_or_null":null,"active_pointer_set":[],"authority_anchor_set":[]}`),
	}
	projection := contracts.FreshImportTopologyProjectionV1{
		ProjectionVersion:            1,
		TargetActivationID:           boundaryUUID("activation"),
		TargetDeploymentID:           boundaryUUID("deployment"),
		TargetDatabaseIdentityDigest: boundaryDigest("target-database"),
		NormalizedCatalogDigest:      boundaryDigest("normalized-catalog"),
		ObjectCount:                  1,
		Objects:                      []contracts.FreshImportTopologyProjectionObjectV1{normalizedNode},
	}
	if err := projection.Validate(); err != nil {
		t.Fatalf("valid normalized projection: %v", err)
	}
	projection.Objects[0].NormalizedPayload = []byte(`null`)
	if err := projection.Validate(); !errors.Is(err, contracts.ErrInvalidAuthorityValue) {
		t.Fatalf("null normalized payload error = %v", err)
	}
}

// TestFreshImportTopologyProjectionAllowsCanonicalEmptyPreState catches the
// production break where the projection validator conflates the required
// pristine pre-import snapshot with the nonempty manifest/post-import shape.
func TestFreshImportTopologyProjectionAllowsCanonicalEmptyPreState(t *testing.T) {
	empty := contracts.FreshImportTopologyProjectionV1{
		ProjectionVersion:            1,
		TargetActivationID:           uuid.MustParse("10000000-0000-4000-8000-000000000001"),
		TargetDeploymentID:           uuid.MustParse("20000000-0000-4000-8000-000000000002"),
		TargetDatabaseIdentityDigest: contracts.Digest{0x11, 0x22, 0x33, 0x44},
		NormalizedCatalogDigest:      contracts.Digest{0x55, 0x66, 0x77, 0x88},
		ObjectCount:                  0,
		Objects:                      []contracts.FreshImportTopologyProjectionObjectV1{},
	}
	if err := empty.Validate(); err != nil {
		t.Fatalf("canonical object_count=0, objects=[] pre-state rejected: %v", err)
	}

	validObject := contracts.FreshImportTopologyProjectionObjectV1{
		ObjectType:        contracts.FreshRestoreImportObjectPOP,
		CanonicalKey:      "tbs-1",
		NormalizedPayload: []byte(`{"pop_code":"tbs-1","iso_country":"GE","region":"tbilisi","operator_state":"disabled","version":"1"}`),
	}
	cases := []struct {
		name        string
		mutate      func(*contracts.FreshImportTopologyProjectionV1)
		objectCount uint64
		objects     []contracts.FreshImportTopologyProjectionObjectV1
	}{
		{name: "nil objects is not canonical empty", objectCount: 0, objects: nil},
		{name: "nonzero count cannot describe empty objects", objectCount: 1, objects: []contracts.FreshImportTopologyProjectionObjectV1{}},
		{name: "count must equal object length", objectCount: 2, objects: []contracts.FreshImportTopologyProjectionObjectV1{validObject}},
		{name: "zero count cannot describe one object", objectCount: 0, objects: []contracts.FreshImportTopologyProjectionObjectV1{validObject}},
		{name: "projection version remains closed", objectCount: 0, objects: []contracts.FreshImportTopologyProjectionObjectV1{}, mutate: func(value *contracts.FreshImportTopologyProjectionV1) {
			value.ProjectionVersion = 2
		}},
		{name: "target activation remains required", objectCount: 0, objects: []contracts.FreshImportTopologyProjectionObjectV1{}, mutate: func(value *contracts.FreshImportTopologyProjectionV1) {
			value.TargetActivationID = uuid.Nil
		}},
		{name: "target deployment remains required", objectCount: 0, objects: []contracts.FreshImportTopologyProjectionObjectV1{}, mutate: func(value *contracts.FreshImportTopologyProjectionV1) {
			value.TargetDeploymentID = uuid.Nil
		}},
		{name: "database identity digest remains required", objectCount: 0, objects: []contracts.FreshImportTopologyProjectionObjectV1{}, mutate: func(value *contracts.FreshImportTopologyProjectionV1) {
			value.TargetDatabaseIdentityDigest = contracts.Digest{}
		}},
		{name: "catalog digest remains required", objectCount: 0, objects: []contracts.FreshImportTopologyProjectionObjectV1{}, mutate: func(value *contracts.FreshImportTopologyProjectionV1) {
			value.NormalizedCatalogDigest = contracts.Digest{}
		}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			candidate := empty
			candidate.ObjectCount = test.objectCount
			candidate.Objects = test.objects
			if test.mutate != nil {
				test.mutate(&candidate)
			}
			if err := candidate.Validate(); !errors.Is(err, contracts.ErrInvalidAuthorityValue) {
				t.Fatalf("Validate() error = %v, want ErrInvalidAuthorityValue", err)
			}
		})
	}

	nonemptyCases := []struct {
		name    string
		objects []contracts.FreshImportTopologyProjectionObjectV1
	}{
		{name: "invalid member", objects: []contracts.FreshImportTopologyProjectionObjectV1{{
			ObjectType: contracts.FreshRestoreImportObjectPOP, CanonicalKey: "tbs-1", NormalizedPayload: []byte(`null`),
		}}},
		{name: "duplicate member", objects: []contracts.FreshImportTopologyProjectionObjectV1{validObject, validObject}},
		{name: "unordered members", objects: []contracts.FreshImportTopologyProjectionObjectV1{
			{ObjectType: contracts.FreshRestoreImportObjectPOP, CanonicalKey: "z-pop", NormalizedPayload: []byte(`{"pop_code":"z-pop","iso_country":"GE","region":"tbilisi","operator_state":"disabled","version":"1"}`)},
			validObject,
		}},
	}
	for _, test := range nonemptyCases {
		t.Run(test.name, func(t *testing.T) {
			candidate := empty
			candidate.ObjectCount = uint64(len(test.objects))
			candidate.Objects = test.objects
			if err := candidate.Validate(); !errors.Is(err, contracts.ErrInvalidAuthorityValue) {
				t.Fatalf("Validate() error = %v, want ErrInvalidAuthorityValue", err)
			}
		})
	}
}

func boundaryDigest(label string) contracts.Digest {
	return contracts.Digest(sha256.Sum256([]byte(label)))
}

func boundaryUUID(label string) uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(label))
}

func task8StructFields(t *testing.T, file *ast.File) map[string][]string {
	t.Helper()
	result := make(map[string][]string)
	for _, declaration := range file.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok || general.Tok != token.TYPE {
			continue
		}
		for _, rawSpec := range general.Specs {
			spec := rawSpec.(*ast.TypeSpec)
			structure, ok := spec.Type.(*ast.StructType)
			if !ok {
				continue
			}
			fields := make([]string, 0, len(structure.Fields.List))
			for _, field := range structure.Fields.List {
				if len(field.Names) != 1 {
					t.Fatalf("%s contains anonymous or grouped field", spec.Name.Name)
				}
				name := field.Names[0].Name
				fields = append(fields, name)
				if field.Tag == nil {
					t.Errorf("%s.%s lacks a JSON tag", spec.Name.Name, name)
					continue
				}
				tagLiteral, err := strconv.Unquote(field.Tag.Value)
				if err != nil {
					t.Fatal(err)
				}
				wantTag := lowerSnake(name)
				if gotTag := reflect.StructTag(tagLiteral).Get("json"); gotTag != wantTag {
					t.Errorf("%s.%s JSON tag = %q, want %q", spec.Name.Name, name, gotTag, wantTag)
				}
			}
			result[spec.Name.Name] = fields
		}
	}
	return result
}

func task8MethodNames(file *ast.File) map[string]bool {
	result := make(map[string]bool)
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Recv == nil || len(function.Recv.List) != 1 {
			continue
		}
		receiver := ""
		switch value := function.Recv.List[0].Type.(type) {
		case *ast.Ident:
			receiver = value.Name
		case *ast.StarExpr:
			if identifier, ok := value.X.(*ast.Ident); ok {
				receiver = "*" + identifier.Name
			}
		}
		result[receiver+"."+function.Name.Name] = true
	}
	return result
}

func lowerSnake(value string) string {
	var result []byte
	for index := 0; index < len(value); index++ {
		character := value[index]
		if character >= 'A' && character <= 'Z' {
			previousIsLowerOrDigit := index > 0 && (value[index-1] >= 'a' && value[index-1] <= 'z' || value[index-1] >= '0' && value[index-1] <= '9')
			nextIsLower := index+1 < len(value) && value[index+1] >= 'a' && value[index+1] <= 'z'
			previousIsUpper := index > 0 && value[index-1] >= 'A' && value[index-1] <= 'Z'
			if index > 0 && (previousIsLowerOrDigit || previousIsUpper && nextIsLower) {
				result = append(result, '_')
			}
			result = append(result, character-'A'+'a')
			continue
		}
		result = append(result, character)
	}
	return string(result)
}

func sortedStrings(values []string) []string {
	copyOfValues := append([]string(nil), values...)
	sort.Strings(copyOfValues)
	return copyOfValues
}
