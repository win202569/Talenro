package authority

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/gowebpki/jcs"
	"talenro.local/platform/internal/nodecontrol/contracts"
)

func TestNodeControlV7SealedGrantTypes(t *testing.T) {
	wantFiles := map[string]string{
		"VerifiedAuthorityV7UpGrant":           "v7_migration_grant.go",
		"VerifiedAuthorityV7DownGrant":         "v7_migration_grant.go",
		"VerifiedAuthorityV7DownAuthorization": "v7_down_authorization.go",
		"VerifiedFreshRestoreImportAdmission":  "v7_staging_import.go",
	}
	for typeName, path := range wantFiles {
		file := task8ParseAuthorityFile(t, path)
		structure := task8FindStruct(t, file, typeName)
		if len(structure.Fields.List) != 1 {
			t.Fatalf("%s field count = %d, want one unexported pointer", typeName, len(structure.Fields.List))
		}
		field := structure.Fields.List[0]
		if len(field.Names) != 1 || unicode.IsUpper(rune(field.Names[0].Name[0])) {
			t.Fatalf("%s field must be one unexported name", typeName)
		}
		if _, ok := field.Type.(*ast.StarExpr); !ok {
			t.Fatalf("%s field is %T, want pointer to sealed immutable state", typeName, field.Type)
		}
	}

	for _, sourcePath := range []string{"v7_migration_grant.go", "v7_down_authorization.go", "v7_staging_import.go"} {
		raw := string(task8ReadAuthorityFile(t, sourcePath))
		for _, forbidden := range []string{"unsafe.", "reflect.", "func NewVerified", "func NewAuthorityV7UpGrant", "func NewAuthorityV7DownGrant"} {
			if strings.Contains(raw, forbidden) {
				t.Errorf("%s exposes forbidden opaque bypass %q", sourcePath, forbidden)
			}
		}
	}

	// These literal private registries catch the production break where the
	// verifier accepts canonical preimages but the opaque value discards them
	// before persistence. They deliberately inspect declarations, not comments.
	wantPrivateViews := map[string][]string{
		"authorityV7DownPersistenceView": {
			"Facts", "ProviderRetirementSetBodyJCS", "ProviderRetirementEvidenceBundleJCS",
		},
		"authorityV7DownAuthorizationPersistenceView": {
			"Facts", "AuthorizationBodyJCS", "AuthorizationEnvelopeJCS",
		},
		"freshRestoreImportAdmissionPersistenceView": {
			"Facts", "StagingImportCapabilityBodyJCS", "StagingImportCapabilityEnvelopeJCS",
			"StagingImportCapabilityEvidenceBundleJCS", "ManifestBodyJCS", "ManifestEnvelopeJCS",
			"ManifestEvidenceBundleJCS", "StagingExclusionLeaseBodyJCS", "StagingExclusionLeaseEnvelopeJCS",
			"StagingExclusionLeaseEvidenceBundleJCS", "AcquisitionLockedProviderHeadBodyJCS",
			"AcquisitionLockedProviderHeadEnvelopeJCS", "CurrentDatabaseIncarnationProofBodyJCS",
			"CurrentDatabaseIncarnationProofEnvelopeJCS", "CurrentDatabaseIncarnationProofEvidenceBundleJCS",
		},
	}
	files := map[string]*ast.File{
		"v7_migration_grant.go":    task8ParseAuthorityFile(t, "v7_migration_grant.go"),
		"v7_down_authorization.go": task8ParseAuthorityFile(t, "v7_down_authorization.go"),
		"v7_staging_import.go":     task8ParseAuthorityFile(t, "v7_staging_import.go"),
	}
	for typeName, wantFields := range wantPrivateViews {
		var found *ast.StructType
		for _, file := range files {
			found = task8MaybeFindStruct(file, typeName)
			if found != nil {
				break
			}
		}
		if found == nil {
			t.Errorf("sealed opaque state discards required private view %s", typeName)
			continue
		}
		if got := task8StructFieldNames(t, typeName, found); !reflect.DeepEqual(got, wantFields) {
			t.Errorf("%s fields = %v, want exact private preimage registry %v", typeName, got, wantFields)
		}
		for index, field := range found.Fields.List {
			if index == 0 {
				if _, ok := field.Type.(*ast.SelectorExpr); !ok {
					t.Errorf("%s Facts field type = %T, want exact contracts facts value", typeName, field.Type)
				}
				continue
			}
			array, ok := field.Type.(*ast.ArrayType)
			if !ok || array.Len != nil {
				t.Errorf("%s preimage field %s type = %T, want []byte", typeName, wantFields[index], field.Type)
				continue
			}
			identifier, byteElement := array.Elt.(*ast.Ident)
			if !byteElement || identifier.Name != "byte" {
				t.Errorf("%s preimage field %s element type = %T, want byte", typeName, wantFields[index], array.Elt)
			}
		}
	}
	task8AssertStrictAuthorityV7Verifier(t)
	task8AssertFactoryProbeOverlayAuthentication(t)
	task8RunIntegrationFactoryProbe(t)
	fixtureFile := task8ParseAuthorityFile(t, "v7_opaque_integration_fixture.go")
	task8AssertIntegrationFactoriesSnapshotCallerInputsFirst(t, fixtureFile)
	for functionName, wantParameters := range map[string][]string{
		"NewDisposableAuthorityV7DownGrantForIntegration": {
			"facts", "providerRetirementSetBodyJCS", "providerRetirementEvidenceBundleJCS",
		},
		"NewAuthorityV7DownAuthorizationForIntegration": {"facts", "authorizationBodyJCS"},
		"NewFreshRestoreImportAdmissionForIntegration": {
			"facts", "stagingImportCapabilityBodyJCS", "stagingImportCapabilityEnvelopeJCS",
			"stagingImportCapabilityEvidenceBundleJCS", "manifestBodyJCS", "manifestEnvelopeJCS",
			"manifestEvidenceBundleJCS", "stagingExclusionLeaseBodyJCS", "stagingExclusionLeaseEnvelopeJCS",
			"stagingExclusionLeaseEvidenceBundleJCS", "acquisitionLockedProviderHeadBodyJCS",
			"acquisitionLockedProviderHeadEnvelopeJCS", "currentDatabaseIncarnationProofBodyJCS",
			"currentDatabaseIncarnationProofEnvelopeJCS", "currentDatabaseIncarnationProofEvidenceBundleJCS",
		},
	} {
		if got := task8FunctionParameterNames(t, fixtureFile, functionName); !reflect.DeepEqual(got, wantParameters) {
			t.Errorf("%s parameters = %v, want revised clone-safe signature %v", functionName, got, wantParameters)
		}
	}
}

func task8AssertFactoryProbeOverlayAuthentication(t *testing.T) {
	t.Helper()
	repositoryRoot, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(repositoryRoot, ".superpowers", "sdd", "task-8-corrective-implementation-plan", "task-2-overlay-gate", "overlay.json")
	raw, err := os.ReadFile(overlayPath)
	if err != nil {
		t.Fatal(err)
	}
	approved, err := task8AuthenticateApprovedFactoryProbeOverlay(repositoryRoot, raw)
	if err != nil {
		t.Fatalf("approved overlay authentication: %v", err)
	}
	malicious := task8FactoryProbeOverlay{Replace: make(map[string]string, len(approved.Replace)+1)}
	for source, replacement := range approved.Replace {
		malicious.Replace[source] = replacement
	}
	malicious.Replace[filepath.ToSlash(filepath.Join(repositoryRoot, "internal", "nodecontrol", "authority", "v7_migration_grant.go"))] = filepath.ToSlash(overlayPath)
	maliciousRaw, err := json.Marshal(malicious)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := task8AuthenticateApprovedFactoryProbeOverlay(repositoryRoot, maliciousRaw); err == nil {
		t.Fatal("approved overlay authentication accepted a Task 2 product replacement")
	}
	probePath := filepath.Join(repositoryRoot, ".superpowers", "sdd", ".t8", "overlay-auth-test", "v7_factory_probe_integration_test.go")
	merged := task8FactoryProbeOverlay{Replace: make(map[string]string, len(approved.Replace)+1)}
	for source, replacement := range approved.Replace {
		merged.Replace[source] = replacement
	}
	merged.Replace[filepath.ToSlash(filepath.Join(repositoryRoot, "internal", "nodecontrol", "authority", "v7_factory_probe_integration_test.go"))] = filepath.ToSlash(probePath)
	mergedRaw, err := json.Marshal(merged)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := task8AuthenticateMergedFactoryProbeOverlay(repositoryRoot, probePath, mergedRaw); err != nil {
		t.Fatalf("merged overlay authentication: %v", err)
	}
	merged.Replace[filepath.ToSlash(filepath.Join(repositoryRoot, "internal", "nodecontrol", "authority", "v7_opaque_integration_fixture.go"))] = filepath.ToSlash(probePath)
	mergedRaw, err = json.Marshal(merged)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := task8AuthenticateMergedFactoryProbeOverlay(repositoryRoot, probePath, mergedRaw); err == nil {
		t.Fatal("merged overlay authentication accepted a factory replacement")
	}
}

func TestTask8ChildGoCommandBindsAuthenticatedOverlayAndWindowsTarget(t *testing.T) {
	repositoryRoot, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	approvedOverlayPath := filepath.Join(repositoryRoot, ".superpowers", "sdd", "task-8-corrective-implementation-plan", "task-2-overlay-gate", "overlay.json")

	for _, test := range []struct {
		name string
		want []string
	}{
		{
			name: "semantic probe environment",
			want: []string{"windows", "amd64", "-overlay=" + approvedOverlayPath},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			command := task8ChildGoCommand(t, t.TempDir(), "env", "GOOS", "GOARCH", "GOFLAGS")
			output, err := command.CombinedOutput()
			if err != nil {
				t.Fatalf("inspect child Go environment: %v\n%s", err, output)
			}
			if got := strings.Fields(string(output)); !reflect.DeepEqual(got, test.want) {
				t.Fatalf("child Go environment = %q, want %q", got, test.want)
			}
		})
	}
}

func task8ChildGoCommand(t *testing.T, directory string, arguments ...string) *exec.Cmd {
	t.Helper()
	repositoryRoot, err := filepath.Abs(filepath.Join("..", "..", ".."))
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
	if _, err := task8AuthenticateApprovedFactoryProbeOverlay(repositoryRoot, raw); err != nil {
		t.Fatalf("authenticate approved child overlay: %v", err)
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

type task8FactoryProbeOverlay struct {
	Replace map[string]string `json:"Replace"`
}

func task8DecodeFactoryProbeOverlay(raw []byte) (task8FactoryProbeOverlay, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var overlay task8FactoryProbeOverlay
	if err := decoder.Decode(&overlay); err != nil {
		return task8FactoryProbeOverlay{}, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return task8FactoryProbeOverlay{}, errors.New("overlay has a trailing JSON value")
		}
		return task8FactoryProbeOverlay{}, err
	}
	if overlay.Replace == nil {
		return task8FactoryProbeOverlay{}, errors.New("overlay Replace is nil")
	}
	return overlay, nil
}

func task8ApprovedFactoryProbeReplacementSet(repositoryRoot string) map[string]string {
	overlayRoot := filepath.Join(repositoryRoot, ".superpowers", "sdd", "task-8-corrective-implementation-plan", "task-2-overlay-gate")
	return map[string]string{
		filepath.ToSlash(filepath.Join(repositoryRoot, "internal", "nodecontrol", "authority", "postgres_repository.go")):   filepath.ToSlash(filepath.Join(overlayRoot, "postgres_repository.go")),
		filepath.ToSlash(filepath.Join(repositoryRoot, "db", "migrations", "00007_nodecontrol_authority_abort_serving.go")): filepath.ToSlash(filepath.Join(overlayRoot, "00007_nodecontrol_authority_abort_serving.go")),
	}
}

func task8RequireFactoryProbeFileHash(path, wantHex string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	got := sha256.Sum256(raw)
	if hex.EncodeToString(got[:]) != strings.ToLower(wantHex) {
		return fmt.Errorf("%s SHA-256 = %X, want %s", path, got, wantHex)
	}
	return nil
}

func task8AuthenticateFactoryProbeGuardedFiles(repositoryRoot string) error {
	for path, wantHex := range map[string]string{
		filepath.Join(repositoryRoot, "internal", "nodecontrol", "authority", "postgres_repository.go"):   "4F759A5B17648A75DF8930939282D51D63F1EC4D8563C254BFE8755321B9AFF3",
		filepath.Join(repositoryRoot, "db", "migrations", "00007_nodecontrol_authority_abort_serving.go"): "31352FB84C3CCB69ECD1C999D69D13C4C83D216A9DCA1D015BEBFD902E3E2B1E",
	} {
		if err := task8RequireFactoryProbeFileHash(path, wantHex); err != nil {
			return fmt.Errorf("guarded caller: %w", err)
		}
	}
	for path, wantHex := range map[string]string{
		filepath.Join(repositoryRoot, ".superpowers", "sdd", "task-8-corrective-implementation-plan", "task-2-overlay-gate", "postgres_repository.go"):                       "8F3447CA7D6E32C41FF80057750A6172DB1C2207673BCADB8CDFD9752AE5B5DB",
		filepath.Join(repositoryRoot, ".superpowers", "sdd", "task-8-corrective-implementation-plan", "task-2-overlay-gate", "00007_nodecontrol_authority_abort_serving.go"): "C01D113F656E3C064536A3BC475A1CE307579766824210984F8E7711D8B182C0",
	} {
		if err := task8RequireFactoryProbeFileHash(path, wantHex); err != nil {
			return fmt.Errorf("approved caller replacement: %w", err)
		}
	}
	return nil
}

func task8AuthenticateApprovedFactoryProbeOverlay(repositoryRoot string, raw []byte) (task8FactoryProbeOverlay, error) {
	wantOverlayHash, err := hex.DecodeString("39ED8367879A0A5D25F2AD6F4C6A2AC3B5F5CE9E3A77EEA3382C7C52311DF0E2")
	if err != nil {
		return task8FactoryProbeOverlay{}, err
	}
	gotOverlayHash := sha256.Sum256(raw)
	if !bytes.Equal(gotOverlayHash[:], wantOverlayHash) {
		return task8FactoryProbeOverlay{}, fmt.Errorf("approved overlay SHA-256 = %X", gotOverlayHash)
	}
	overlay, err := task8DecodeFactoryProbeOverlay(raw)
	if err != nil {
		return task8FactoryProbeOverlay{}, err
	}
	if want := task8ApprovedFactoryProbeReplacementSet(repositoryRoot); !reflect.DeepEqual(overlay.Replace, want) {
		return task8FactoryProbeOverlay{}, fmt.Errorf("approved overlay replacements = %v, want exact %v", overlay.Replace, want)
	}
	if err := task8AuthenticateFactoryProbeGuardedFiles(repositoryRoot); err != nil {
		return task8FactoryProbeOverlay{}, err
	}
	return overlay, nil
}

func task8AuthenticateMergedFactoryProbeOverlay(repositoryRoot, probeSourcePath string, raw []byte) (task8FactoryProbeOverlay, error) {
	overlay, err := task8DecodeFactoryProbeOverlay(raw)
	if err != nil {
		return task8FactoryProbeOverlay{}, err
	}
	probeSourcePath, err = filepath.Abs(probeSourcePath)
	if err != nil {
		return task8FactoryProbeOverlay{}, err
	}
	probeRoot := filepath.Join(repositoryRoot, ".superpowers", "sdd", ".t8")
	relativeProbe, err := filepath.Rel(probeRoot, probeSourcePath)
	if err != nil || filepath.IsAbs(relativeProbe) || relativeProbe == ".." || strings.HasPrefix(relativeProbe, ".."+string(filepath.Separator)) || filepath.Base(probeSourcePath) != "v7_factory_probe_integration_test.go" {
		return task8FactoryProbeOverlay{}, errors.New("factory probe replacement is outside the approved temporary probe root")
	}
	want := task8ApprovedFactoryProbeReplacementSet(repositoryRoot)
	want[filepath.ToSlash(filepath.Join(repositoryRoot, "internal", "nodecontrol", "authority", "v7_factory_probe_integration_test.go"))] = filepath.ToSlash(probeSourcePath)
	if !reflect.DeepEqual(overlay.Replace, want) {
		return task8FactoryProbeOverlay{}, fmt.Errorf("merged overlay replacements = %v, want exact %v", overlay.Replace, want)
	}
	if err := task8AuthenticateFactoryProbeGuardedFiles(repositoryRoot); err != nil {
		return task8FactoryProbeOverlay{}, err
	}
	return overlay, nil
}

func TestApprovedMigrationBackingBindsAuthorizedStateToConsume(t *testing.T) {
	repositoryRoot, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	backingPath := filepath.Join(repositoryRoot, ".superpowers", "sdd", "task-8-corrective-implementation-plan", "task-2-overlay-gate", "00007_nodecontrol_authority_abort_serving.go")
	raw, err := os.ReadFile(backingPath)
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	for _, required := range []string{
		`SELECT txid_current()::text, transaction_timestamp()`,
		`authorityV7CanonicalBody("pristine-downgrade-authorized-state.v1", authorizedState)`,
		`FROM nodecontrol.v7_consume_down_guard($1,$2,$3,$4,$5,$6)`,
		`authorizedStateJCS, authorizedStateDigest[:]`,
	} {
		if !strings.Contains(source, required) {
			t.Errorf("approved migration backing lacks authorized-state consume binding %q", required)
		}
	}
	if strings.Contains(source, `SELECT nodecontrol.v7_consume_down_guard($1,$2,$3,$4)`) {
		t.Error("approved migration backing retains the obsolete four-argument consume call")
	}
}

func TestNodeControlV7OpaqueCrossPackageUse(t *testing.T) {
	for _, path := range []string{"v7_migration_grant.go", "v7_down_authorization.go", "v7_staging_import.go"} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("Task 8 opaque implementation %s is missing: %v", path, err)
		}
	}
	repositoryRoot, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	probeParent := filepath.Join(repositoryRoot, ".superpowers", "sdd", ".t8")
	if err := os.MkdirAll(probeParent, 0o700); err != nil {
		t.Fatal(err)
	}
	probeDirectory, err := os.MkdirTemp(probeParent, "opaque-probe-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(probeDirectory) })
	probe := `package opaqueprobe
import (
 "errors"
 "testing"
 authority "talenro.local/platform/internal/nodecontrol/authority"
)
func TestOpaqueBoundary(t *testing.T) {
 var up authority.VerifiedAuthorityV7UpGrant
 upCopy := up
 if _, err := authority.ConsumeVerifiedAuthorityV7UpGrant(upCopy); !errors.Is(err, authority.ErrInvalidArgument) { t.Fatalf("zero Up error = %v", err) }
 var down authority.VerifiedAuthorityV7DownGrant
 if _, err := authority.ConsumeVerifiedAuthorityV7DownGrant(down); !errors.Is(err, authority.ErrInvalidArgument) { t.Fatalf("zero Down error = %v", err) }
 var authorization authority.VerifiedAuthorityV7DownAuthorization
 if _, err := authority.ConsumeVerifiedAuthorityV7DownAuthorization(authorization); !errors.Is(err, authority.ErrInvalidArgument) { t.Fatalf("zero authorization error = %v", err) }
 var admission authority.VerifiedFreshRestoreImportAdmission
 if _, err := authority.ViewVerifiedFreshRestoreImportAdmission(admission); !errors.Is(err, authority.ErrInvalidArgument) { t.Fatalf("zero admission error = %v", err) }
}
`
	if err := os.WriteFile(filepath.Join(probeDirectory, "opaque_test.go"), []byte(probe), 0o600); err != nil {
		t.Fatal(err)
	}
	command := task8ChildGoCommand(t, probeDirectory, "test", ".", "-run", "^TestOpaqueBoundary$", "-count=1")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("cross-package opaque semantic probe failed: %v\n%s", err, output)
	}

	for _, mutation := range []struct {
		name   string
		mutate func(*contracts.AuthorityV7DownAuthorizationFactsV1)
	}{
		{
			name: "duplicate envelope key",
			mutate: func(value *contracts.AuthorityV7DownAuthorizationFactsV1) {
				value.AuthorizationEnvelopeJCS = []byte(`{"body":"task8","body":"duplicate"}`)
			},
		},
		{
			name: "noncanonical envelope bytes",
			mutate: func(value *contracts.AuthorityV7DownAuthorizationFactsV1) {
				value.AuthorizationEnvelopeJCS = []byte(`{ "body":"task8"}`)
			},
		},
		{
			name: "body digest mismatch",
			mutate: func(value *contracts.AuthorityV7DownAuthorizationFactsV1) {
				value.AuthorizationDigest = contracts.Digest{0xde, 0xad, 0xbe, 0xef}
			},
		},
	} {
		t.Run("rejects "+mutation.name, func(t *testing.T) {
			view := task8AuthorizationPersistenceView(time.Now().UTC())
			mutation.mutate(&view.Facts)
			if _, err := newVerifiedAuthorityV7DownAuthorization(view); !errors.Is(err, ErrInvalidArgument) {
				t.Fatalf("constructor error = %v, want ErrInvalidArgument", err)
			}
		})
	}

	t.Run("constructors clone caller-owned mutable facts", func(t *testing.T) {
		downFacts := task8DownFacts(time.Now().UTC())
		downView := task8DownPersistenceView(downFacts)
		down, err := newVerifiedAuthorityV7DownGrant(downView)
		if err != nil {
			t.Fatal(err)
		}
		downFacts.StableTableInventory[0].TableName = "mutated.after.construction"
		for _, arm := range task8DownPrivateByteArms(&downView) {
			arm[0] ^= 0xff
		}
		gotDown, err := ConsumeVerifiedAuthorityV7DownGrant(down)
		if err != nil || gotDown.Facts.StableTableInventory[0].TableName != "nodecontrol.node_pops" || !task8AllByteArmsStartWith(task8DownPrivateByteArms(&gotDown), '{') {
			t.Fatalf("sealed Down view aliased caller input: table=%q body=%#x error=%v", gotDown.Facts.StableTableInventory[0].TableName, gotDown.ProviderRetirementSetBodyJCS[0], err)
		}

		admissionFacts := task8AdmissionFacts(time.Now().UTC())
		admissionView := task8AdmissionPersistenceView(admissionFacts)
		admission, err := newVerifiedFreshRestoreImportAdmission(admissionView)
		if err != nil {
			t.Fatal(err)
		}
		admissionFacts.ManifestTopology.Objects[0].Payload[0] ^= 0xff
		for _, arm := range task8AdmissionPrivateByteArms(&admissionView) {
			arm[0] ^= 0xff
		}
		view, err := ViewVerifiedFreshRestoreImportAdmission(admission)
		if err != nil || view.ManifestTopology.Objects[0].Payload[0] != '{' || !task8AllByteArmsStartWith(task8AdmissionPrivateByteArms(&admission.state.view), '{') {
			t.Fatalf("sealed staging facts aliased caller input: first=%#x error=%v", view.ManifestTopology.Objects[0].Payload[0], err)
		}
	})

	upFacts := task8DisposableUpFacts(time.Now().UTC())
	up, err := newVerifiedAuthorityV7UpGrant(upFacts)
	if err != nil {
		t.Fatal(err)
	}
	copies := make([]VerifiedAuthorityV7UpGrant, 32)
	for index := range copies {
		copies[index] = up
	}
	start := make(chan struct{})
	results := make(chan error, len(copies))
	var wait sync.WaitGroup
	for _, copyOfGrant := range copies {
		wait.Add(1)
		go func(grant VerifiedAuthorityV7UpGrant) {
			defer wait.Done()
			<-start
			_, consumeErr := ConsumeVerifiedAuthorityV7UpGrant(grant)
			results <- consumeErr
		}(copyOfGrant)
	}
	close(start)
	wait.Wait()
	close(results)
	successes, conflicts := 0, 0
	for consumeErr := range results {
		switch {
		case consumeErr == nil:
			successes++
		case errors.Is(consumeErr, ErrConflict):
			conflicts++
		default:
			t.Fatalf("32-way consume error = %v", consumeErr)
		}
	}
	if successes != 1 || conflicts != 31 {
		t.Fatalf("32-way copied grant results = success:%d conflict:%d, want 1/31", successes, conflicts)
	}
	task8AssertCopiedDownConsume(t, task8DownFacts(time.Now().UTC()))
	task8AssertCopiedAuthorizationConsume(t, time.Now().UTC())
	task8AssertCopiedAdmissionConsume(t, task8AdmissionFacts(time.Now().UTC()))

	expiredFacts := task8DisposableUpFacts(time.Now().UTC().Add(-10 * time.Minute))
	expiredFacts.ExpiresAt = time.Now().UTC().Add(-time.Minute)
	expired, err := newVerifiedAuthorityV7UpGrant(expiredFacts)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ConsumeVerifiedAuthorityV7UpGrant(expired); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("expired first consume = %v, want ErrInvalidArgument", err)
	}
	if _, err := ConsumeVerifiedAuthorityV7UpGrant(expired); !errors.Is(err, ErrConflict) {
		t.Fatalf("expired burned retry = %v, want ErrConflict", err)
	}

	downFacts := task8DownFacts(time.Now().UTC())
	down, err := newVerifiedAuthorityV7DownGrant(task8DownPersistenceView(downFacts))
	if err != nil {
		t.Fatal(err)
	}
	returnedDown, err := ConsumeVerifiedAuthorityV7DownGrant(down)
	if err != nil {
		t.Fatal(err)
	}
	returnedDown.Facts.StableTableInventory[0].TableName = "mutated.table"
	returnedDown.Facts.ProviderRetirementSet.RetiredMembers[0].DatabaseName = "mutated"
	returnedDown.ProviderRetirementSetBodyJCS[0] ^= 0xff
	returnedDown.ProviderRetirementEvidenceBundleJCS[0] ^= 0xff
	if down.state.view.Facts.StableTableInventory[0].TableName == "mutated.table" || down.state.view.Facts.ProviderRetirementSet.RetiredMembers[0].DatabaseName == "mutated" ||
		down.state.view.ProviderRetirementSetBodyJCS[0] == returnedDown.ProviderRetirementSetBodyJCS[0] ||
		down.state.view.ProviderRetirementEvidenceBundleJCS[0] == returnedDown.ProviderRetirementEvidenceBundleJCS[0] {
		t.Fatal("Down consume returned aliases into sealed state")
	}

	authorizationView := task8AuthorizationPersistenceView(time.Now().UTC())
	authorization, err := newVerifiedAuthorityV7DownAuthorization(authorizationView)
	if err != nil {
		t.Fatal(err)
	}
	for _, arm := range task8AuthorizationPrivateByteArms(&authorizationView) {
		arm[0] ^= 0xff
	}
	if !task8AllByteArmsStartWith(task8AuthorizationPrivateByteArms(&authorization.state.view), '{') {
		t.Fatal("Down authorization sealed state aliases caller-owned bytes")
	}
	authorizationCopy := authorization
	returnedAuthorization, err := ConsumeVerifiedAuthorityV7DownAuthorization(authorizationCopy)
	if err != nil {
		t.Fatal(err)
	}
	for _, arm := range task8AuthorizationPrivateByteArms(&returnedAuthorization) {
		arm[0] ^= 0xff
	}
	if !task8AllByteArmsStartWith(task8AuthorizationPrivateByteArms(&authorization.state.view), '{') {
		t.Fatal("Down authorization consume returned an envelope alias")
	}
	if _, err := ConsumeVerifiedAuthorityV7DownAuthorization(authorization); !errors.Is(err, ErrConflict) {
		t.Fatalf("copied authorization retry = %v, want ErrConflict", err)
	}

	admissionFacts := task8AdmissionFacts(time.Now().UTC())
	admission, err := newVerifiedFreshRestoreImportAdmission(task8AdmissionPersistenceView(admissionFacts))
	if err != nil {
		t.Fatal(err)
	}
	firstView, err := ViewVerifiedFreshRestoreImportAdmission(admission)
	if err != nil {
		t.Fatal(err)
	}
	firstView.ManifestTopology.Objects[0].Payload[0] ^= 0xff
	secondView, err := ViewVerifiedFreshRestoreImportAdmission(admission)
	if err != nil {
		t.Fatal(err)
	}
	if secondView.ManifestTopology.Objects[0].Payload[0] == firstView.ManifestTopology.Objects[0].Payload[0] {
		t.Fatal("fresh-import repeatable view returned a payload alias")
	}
	consumedAdmission, err := consumeVerifiedFreshRestoreImportAdmission(admission)
	if err != nil {
		t.Fatal(err)
	}
	for _, arm := range task8AdmissionPrivateByteArms(&consumedAdmission) {
		arm[0] ^= 0xff
	}
	if !task8AllByteArmsStartWith(task8AdmissionPrivateByteArms(&admission.state.view), '{') {
		t.Fatal("fresh-import consume returned a private preimage alias")
	}
	if _, err := consumeVerifiedFreshRestoreImportAdmission(admission); !errors.Is(err, ErrConflict) {
		t.Fatalf("fresh-import copied consume retry = %v, want ErrConflict", err)
	}

	nilProjection := cloneFreshImportTopologyProjection(contracts.FreshImportTopologyProjectionV1{Objects: nil})
	if nilProjection.Objects != nil {
		t.Fatal("projection clone changed nil objects into canonical empty")
	}
	emptyProjection := cloneFreshImportTopologyProjection(contracts.FreshImportTopologyProjectionV1{Objects: []contracts.FreshImportTopologyProjectionObjectV1{}})
	if emptyProjection.Objects == nil || len(emptyProjection.Objects) != 0 {
		t.Fatal("projection clone lost present non-nil canonical empty objects")
	}
	originalProjection := contracts.FreshImportTopologyProjectionV1{Objects: []contracts.FreshImportTopologyProjectionObjectV1{{NormalizedPayload: []byte(`{"pop_code":"tbs-1"}`)}}}
	payloadProjection := cloneFreshImportTopologyProjection(originalProjection)
	payloadProjection.Objects[0].NormalizedPayload[0] ^= 0xff
	if originalProjection.Objects[0].NormalizedPayload[0] != '{' {
		t.Fatal("projection clone payload aliases caller-owned bytes")
	}
}

func TestNodeControlV7ProductionBuildOmitsFixtureFactories(t *testing.T) {
	expectations := map[string][]string{
		"v7_opaque_integration_fixture.go": {
			"NewDisposableAuthorityV7UpGrantForIntegration", "NewDisposableAuthorityV7DownGrantForIntegration",
			"NewAuthorityV7DownAuthorizationForIntegration", "NewFreshRestoreImportAdmissionForIntegration",
		},
		filepath.Join("..", "..", "testinfra", "c12_authority_pitr_integration.go"): {"C12AuthorityPITRController", "OpenC12AuthorityPITR"},
	}
	for path, symbols := range expectations {
		raw := task8ReadAuthorityFile(t, path)
		firstLine := strings.SplitN(strings.ReplaceAll(string(raw), "\r\n", "\n"), "\n", 2)[0]
		if firstLine != "//go:build integration" {
			t.Errorf("%s first line = %q, want exact integration tag", path, firstLine)
		}
		for _, symbol := range symbols {
			if !strings.Contains(string(raw), symbol) {
				t.Errorf("%s lacks frozen integration symbol %s", path, symbol)
			}
		}
	}

	ordinary, err := build.Default.ImportDir(".", build.IgnoreVendor)
	if err != nil {
		t.Fatal(err)
	}
	if containsString(ordinary.GoFiles, "v7_opaque_integration_fixture.go") {
		t.Fatal("ordinary authority production build includes integration fixture factories")
	}
	testinfra, err := build.Default.ImportDir(filepath.Join("..", "..", "testinfra"), build.IgnoreVendor)
	if err != nil {
		t.Fatal(err)
	}
	if containsString(testinfra.GoFiles, "c12_authority_pitr_integration.go") {
		t.Fatal("ordinary testinfra production build includes PITR control helper")
	}
}

func TestStoredFencePersistedOutcomeProjection(t *testing.T) {
	file := task8ParseAuthorityFile(t, "repository.go")
	wantFields := map[string][]string{
		"AbortClaim": {"Reason", "ClaimedAt"},
		"PersistedAuthorityEffectOutcome": {
			"CommitmentJCS", "CommitmentDigest", "Receipt", "ProviderHeadJCS", "ProviderHeadDigest",
			"CheckpointAnchorJCS", "CheckpointAnchorDigest", "Reason", "AttestationExpiresAt", "ActivationDeadline",
			"ExpectedProviderIdentityDigest", "EvidenceJCS", "EvidenceDigest", "ResolutionJCS", "ResolutionDigest",
		},
		"StoredFence":        {"Record", "AbortClaim", "PersistedOutcome"},
		"ClaimV1Reservation": {"Reservation", "ProtocolActivationID"},
	}
	for name, want := range wantFields {
		structure := task8FindStruct(t, file, name)
		got := make([]string, 0, len(structure.Fields.List))
		for _, field := range structure.Fields.List {
			if len(field.Names) != 1 {
				t.Fatalf("%s contains anonymous/grouped fields", name)
			}
			got = append(got, field.Names[0].Name)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s fields = %v, want exact persisted projection %v", name, got, want)
		}
	}
	pending := task8FindStruct(t, file, "PendingFence")
	if !task8StructHasField(pending, "AbortClaim") {
		t.Error("PendingFence lacks the protected-compatible AbortClaim projection")
	}
	wantMethods := []string{"RecordPendingClaimV1", "ClaimAbort", "GetStoredFence", "Lock"}
	gotEmbeddings, gotMethods := task8InterfaceShape(t, file, "AuthorityV7Repository")
	if !reflect.DeepEqual(gotEmbeddings, []string{"Repository"}) {
		t.Errorf("AuthorityV7Repository anonymous embeddings = %v, want exactly [Repository]", gotEmbeddings)
	}
	sort.Strings(gotMethods)
	sort.Strings(wantMethods)
	if !reflect.DeepEqual(gotMethods, wantMethods) {
		t.Errorf("AuthorityV7Repository methods = %v, want exact extension %v", gotMethods, wantMethods)
	}
	for _, interfaceName := range []string{"Repository", "AuthorityV7Repository"} {
		if dependencies := task8InterfaceProviderOrCallbackDependencies(t, file, interfaceName); len(dependencies) != 0 {
			t.Errorf("%s exposes forbidden Provider/callback dependency at repository boundary: %v", interfaceName, dependencies)
		}
	}

	repositoryRoot, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	probeParent := filepath.Join(repositoryRoot, ".superpowers", "sdd", ".t8")
	if err := os.MkdirAll(probeParent, 0o700); err != nil {
		t.Fatal(err)
	}
	probeDirectory, err := os.MkdirTemp(probeParent, "repository-embedding-probe-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(probeDirectory) })
	probe := `package repositoryembeddingprobe
import authority "talenro.local/platform/internal/nodecontrol/authority"

func authorityV7RepositoryAsBase(value authority.AuthorityV7Repository) authority.Repository {
	return value
}
`
	if err := os.WriteFile(filepath.Join(probeDirectory, "repository_embedding_test.go"), []byte(probe), 0o600); err != nil {
		t.Fatal(err)
	}
	command := task8ChildGoCommand(t, probeDirectory, "test", ".", "-run", "^$", "-count=1")
	if output, err := command.CombinedOutput(); err != nil {
		t.Errorf("AuthorityV7Repository cannot be directly returned as Repository: %v\n%s", err, output)
	}

	// Exercise the production clone used by both GetStoredFence and Lock with
	// every reference-bearing arm populated. A fresh clone after exhaustive
	// caller mutation must remain byte/value-identical to an independent literal
	// fixture.
	stored := task8FullStoredFenceCopyFixture()
	first := cloneStoredFence(stored)
	first.AbortClaim.Reason = AbortSuperseded
	first.Record.BoundEffectDigest[0] ^= 0xff
	first.Record.BoundDatabasePoint.RequiredLSN = "f/ffffffff"
	first.Record.TerminalReceipt.EffectDigest[0] ^= 0xff
	first.Record.TerminalReceipt.DatabasePoint.RequiredLSN = "e/eeeeeeee"
	*first.Record.TerminalReceipt.AbortReason = AbortValidationFailed
	first.PersistedOutcome.CommitmentJCS[0] ^= 0xff
	first.PersistedOutcome.ProviderHeadJCS[0] ^= 0xff
	first.PersistedOutcome.CheckpointAnchorJCS[0] ^= 0xff
	first.PersistedOutcome.EvidenceJCS[0] ^= 0xff
	first.PersistedOutcome.ResolutionJCS[0] ^= 0xff
	first.PersistedOutcome.Receipt.DatabasePoint.RequiredLSN = "d/dddddddd"
	if got, want := cloneStoredFence(stored), task8FullStoredFenceCopyFixture(); !reflect.DeepEqual(got, want) {
		t.Fatalf("production StoredFence clone aliases caller mutation\ngot=%#v\nwant=%#v", got, want)
	}
}

func task8InterfaceProviderOrCallbackDependencies(t *testing.T, file *ast.File, name string) []string {
	t.Helper()
	var found *ast.InterfaceType
	for _, declaration := range file.Decls {
		generic, ok := declaration.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, specification := range generic.Specs {
			typeSpec, ok := specification.(*ast.TypeSpec)
			if !ok || typeSpec.Name.Name != name {
				continue
			}
			found, _ = typeSpec.Type.(*ast.InterfaceType)
		}
	}
	if found == nil {
		t.Errorf("repository source lacks interface %s", name)
		return nil
	}
	var result []string
	for _, field := range found.Methods.List {
		methodName := "anonymous"
		if len(field.Names) == 1 {
			methodName = field.Names[0].Name
		}
		ast.Inspect(field.Type, func(node ast.Node) bool {
			switch value := node.(type) {
			case *ast.FuncType:
				if node != field.Type {
					result = append(result, methodName+": callback")
					return false
				}
			case *ast.Ident:
				if value.Name == "Provider" {
					result = append(result, methodName+": Provider")
				}
			}
			return true
		})
	}
	return result
}

func task8FullStoredFenceCopyFixture() StoredFence {
	digest := func(value byte) contracts.Digest {
		var result contracts.Digest
		for index := range result {
			result[index] = value
		}
		return result
	}
	effectDigest := digest(0x31)
	receiptEffect := digest(0x32)
	abortReason := AbortProviderDependencyFailed
	point := DatabasePoint{SystemID: 17, Timeline: 3, RequiredLSN: "0/1700000"}
	receiptPoint := DatabasePoint{SystemID: 19, Timeline: 5, RequiredLSN: "0/1900000"}
	reservation := Reservation{
		OperationID: uuid.MustParse("74000000-0000-4000-8000-000000000001"), Kind: EffectDesiredActivate,
		ScopeKind: ScopeNode, ScopeDigest: digest(0x21), Epoch: 74, Sequence: 1, ReservationDigest: digest(0x22),
	}
	receipt := Receipt{
		Reservation: reservation, EffectDigest: &receiptEffect, DatabasePoint: &receiptPoint,
		Status: StatusAborted, AbortReason: &abortReason, ReceiptDigest: digest(0x33),
	}
	return StoredFence{
		Record:     Record{Reservation: reservation, BoundEffectDigest: &effectDigest, BoundDatabasePoint: &point, TerminalReceipt: &receipt},
		AbortClaim: &AbortClaim{Reason: AbortProviderDependencyFailed, ClaimedAt: time.Date(2026, time.August, 29, 16, 0, 0, 0, time.UTC)},
		PersistedOutcome: &PersistedAuthorityEffectOutcome{
			CommitmentJCS: []byte(`{"commitment":"literal"}`), CommitmentDigest: digest(0x41), Receipt: receipt,
			ProviderHeadJCS: []byte(`{"head":"literal"}`), ProviderHeadDigest: digest(0x42),
			CheckpointAnchorJCS: []byte(`{"anchor":"literal"}`), CheckpointAnchorDigest: digest(0x43),
			Reason: AuthorityEffectReason("literal_reason"), AttestationExpiresAt: time.Date(2026, time.August, 29, 16, 1, 0, 0, time.UTC),
			ActivationDeadline: time.Date(2026, time.August, 29, 16, 2, 0, 0, time.UTC), ExpectedProviderIdentityDigest: digest(0x44),
			EvidenceJCS: []byte(`{"evidence":"literal"}`), EvidenceDigest: digest(0x45),
			ResolutionJCS: []byte(`{"resolution":"literal"}`), ResolutionDigest: digest(0x46),
		},
	}
}

func task8ReadAuthorityFile(t *testing.T, path string) []byte {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return raw
}

func task8ParseAuthorityFile(t *testing.T, path string) *ast.File {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), path, task8ReadAuthorityFile(t, path), parser.AllErrors)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return file
}

func task8FunctionParameterNames(t *testing.T, file *ast.File, name string) []string {
	t.Helper()
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Name.Name != name {
			continue
		}
		var result []string
		for _, field := range function.Type.Params.List {
			for _, parameter := range field.Names {
				result = append(result, parameter.Name)
			}
		}
		return result
	}
	t.Errorf("integration fixture lacks function %s", name)
	return nil
}

func task8AssertIntegrationFactoriesSnapshotCallerInputsFirst(t *testing.T, file *ast.File) {
	t.Helper()
	expectedClones := map[string]string{
		"NewDisposableAuthorityV7DownGrantForIntegration": "cloneAuthorityV7DownPersistenceView",
		"NewAuthorityV7DownAuthorizationForIntegration":   "cloneAuthorityV7DownAuthorizationIntegrationInput",
		"NewFreshRestoreImportAdmissionForIntegration":    "cloneFreshRestoreImportAdmissionPersistenceView",
	}
	for functionName, cloneName := range expectedClones {
		var function *ast.FuncDecl
		for _, declaration := range file.Decls {
			candidate, ok := declaration.(*ast.FuncDecl)
			if ok && candidate.Name.Name == functionName {
				function = candidate
				break
			}
		}
		if function == nil || function.Body == nil || len(function.Body.List) == 0 {
			t.Fatalf("integration fixture lacks body for %s", functionName)
		}
		parameters := make(map[string]int)
		for _, field := range function.Type.Params.List {
			for _, parameter := range field.Names {
				parameters[parameter.Name] = 0
			}
		}
		assignment, ok := function.Body.List[0].(*ast.AssignStmt)
		if !ok || assignment.Tok != token.DEFINE || len(assignment.Lhs) != 1 || len(assignment.Rhs) != 1 {
			t.Errorf("%s first statement = %T, want complete private snapshot assignment", functionName, function.Body.List[0])
			continue
		}
		snapshotName, snapshotOK := assignment.Lhs[0].(*ast.Ident)
		call, callOK := assignment.Rhs[0].(*ast.CallExpr)
		clone, cloneOK := func() (*ast.Ident, bool) {
			if !callOK {
				return nil, false
			}
			identifier, ok := call.Fun.(*ast.Ident)
			return identifier, ok
		}()
		if !snapshotOK || snapshotName.Name != "snapshot" || !cloneOK || clone.Name != cloneName {
			t.Errorf("%s first statement does not deep-clone the complete private snapshot with %s", functionName, cloneName)
			continue
		}
		ast.Inspect(assignment, func(node ast.Node) bool {
			identifier, ok := node.(*ast.Ident)
			if ok {
				if _, isParameter := parameters[identifier.Name]; isParameter {
					parameters[identifier.Name]++
				}
			}
			return true
		})
		for parameter, count := range parameters {
			if count != 1 {
				t.Errorf("%s snapshot references caller parameter %s %d times, want exactly once", functionName, parameter, count)
			}
		}
		for _, statement := range function.Body.List[1:] {
			ast.Inspect(statement, func(node ast.Node) bool {
				identifier, ok := node.(*ast.Ident)
				if ok {
					if _, callerOwned := parameters[identifier.Name]; callerOwned {
						t.Errorf("%s reads caller parameter %s after its private snapshot", functionName, identifier.Name)
					}
				}
				return true
			})
		}
	}
}

func task8FindStruct(t *testing.T, file *ast.File, name string) *ast.StructType {
	t.Helper()
	for _, declaration := range file.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok || general.Tok != token.TYPE {
			continue
		}
		for _, rawSpec := range general.Specs {
			spec := rawSpec.(*ast.TypeSpec)
			if spec.Name.Name == name {
				structure, ok := spec.Type.(*ast.StructType)
				if !ok {
					t.Fatalf("%s is %T, want struct", name, spec.Type)
				}
				return structure
			}
		}
	}
	t.Fatalf("missing struct %s", name)
	return nil
}

func task8MaybeFindStruct(file *ast.File, name string) *ast.StructType {
	for _, declaration := range file.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok || general.Tok != token.TYPE {
			continue
		}
		for _, rawSpec := range general.Specs {
			spec := rawSpec.(*ast.TypeSpec)
			if spec.Name.Name != name {
				continue
			}
			structure, _ := spec.Type.(*ast.StructType)
			return structure
		}
	}
	return nil
}

func task8StructFieldNames(t *testing.T, typeName string, structure *ast.StructType) []string {
	t.Helper()
	fields := make([]string, 0, len(structure.Fields.List))
	for _, field := range structure.Fields.List {
		if len(field.Names) != 1 {
			t.Fatalf("%s contains an anonymous or grouped field", typeName)
		}
		fields = append(fields, field.Names[0].Name)
	}
	return fields
}

func task8StructHasField(structure *ast.StructType, name string) bool {
	for _, field := range structure.Fields.List {
		if len(field.Names) == 1 && field.Names[0].Name == name {
			return true
		}
	}
	return false
}

func task8InterfaceShape(t *testing.T, file *ast.File, name string) ([]string, []string) {
	t.Helper()
	for _, declaration := range file.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok || general.Tok != token.TYPE {
			continue
		}
		for _, rawSpec := range general.Specs {
			spec := rawSpec.(*ast.TypeSpec)
			if spec.Name.Name != name {
				continue
			}
			iface, ok := spec.Type.(*ast.InterfaceType)
			if !ok {
				t.Fatalf("%s is %T, want interface", name, spec.Type)
			}
			var embeddings, methods []string
			for _, field := range iface.Methods.List {
				if len(field.Names) == 1 {
					methods = append(methods, field.Names[0].Name)
					continue
				}
				if len(field.Names) == 0 {
					if identifier, ok := field.Type.(*ast.Ident); ok {
						embeddings = append(embeddings, identifier.Name)
					} else {
						t.Fatalf("%s contains unsupported anonymous interface element %T", name, field.Type)
					}
				}
			}
			return embeddings, methods
		}
	}
	t.Fatalf("missing interface %s", name)
	return nil, nil
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func task8Digest(label string) contracts.Digest {
	return contracts.Digest(sha256.Sum256([]byte(label)))
}

func task8UUID(label string) uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(label))
}

func task8DisposableUpFacts(now time.Time) contracts.AuthorityV7UpMigrationFactsV1 {
	latch := contracts.AuthorityV7MigrationLatchFactsV1{
		InstallationID:         task8UUID("installation"),
		InstallationKind:       contracts.AuthorityV7InstallationKindDisposableFixture,
		MigrationVersion:       7,
		DatabaseIdentityDigest: task8Digest("database-identity"),
		UpCatalogDigest:        task8Digest("up-catalog"),
		DownState:              contracts.AuthorityV7MigrationDownStateLocked,
		InstalledAt:            now.Add(-time.Minute),
	}
	return contracts.AuthorityV7UpMigrationFactsV1{
		MigrationLatch:              latch,
		MigrationLatchDigest:        task8Digest("migration-latch"),
		AuthorityProtocolProfile:    contracts.AuthorityV7ProtocolProfileLegacyV6,
		LocalRuntimeIsolationDigest: task8Digest("local-runtime-isolation"),
		TransactionNonce:            task8Digest("up-transaction-nonce"),
		ExpiresAt:                   now.Add(5 * time.Minute),
	}
}

func task8RetirementSet(now time.Time) contracts.AuthorityV7ProviderRetirementSetFactsV1 {
	return contracts.AuthorityV7ProviderRetirementSetFactsV1{
		InstallationID:                                 task8UUID("installation"),
		ReleaseScope:                                   "release/task8",
		EnvironmentInventoryDigest:                     task8Digest("environment-inventory"),
		EnvironmentInventoryAnchorSetDigest:            task8Digest("environment-anchor-set"),
		EnvironmentInventoryMembershipRetirementDigest: task8Digest("membership-retirement"),
		ProviderCount:                                  1,
		Retirements: []contracts.AuthorityV7ProviderRetirementFactsV1{{
			ProviderIdentityDigest:                    task8Digest("provider-identity"),
			ProviderEndpointIdentityDigest:            task8Digest("provider-endpoint"),
			ProviderProtocolDowngradeRetirementDigest: task8Digest("provider-retirement"),
		}},
		RetiredMemberCount: 1,
		RetiredMembers: []contracts.AuthorityV7RetiredEnvironmentMemberFactsV1{{
			EnvironmentRecordDigest:        task8Digest("environment-record"),
			EnvironmentAttestationDigest:   task8Digest("environment-attestation"),
			DeploymentID:                   task8UUID("deployment"),
			PostgresSystemID:               11,
			Timeline:                       1,
			DatabaseOID:                    17,
			DatabaseName:                   "task8db",
			DatabaseIdentityDigest:         task8Digest("database-identity"),
			EnvironmentInstanceGeneration:  1,
			ProviderIdentityDigest:         task8Digest("provider-identity"),
			ProviderEndpointIdentityDigest: task8Digest("provider-endpoint"),
			ProviderNamespace:              "task8/fixture",
			ProviderProfile:                "legacy_v6",
		}},
		CreatedAt: now.Add(-time.Minute),
	}
}

func task8StableInventory() []contracts.AuthorityV7StableTableInventoryItemV1 {
	return []contracts.AuthorityV7StableTableInventoryItemV1{{TableName: "nodecontrol.node_pops", Classification: "stable", RowCount: 0, ContentDigest: task8Digest("stable-node-pops")}}
}

func task8DownFacts(now time.Time) contracts.AuthorityV7DownMigrationFactsV1 {
	up := task8DisposableUpFacts(now)
	return contracts.AuthorityV7DownMigrationFactsV1{
		MigrationLatch:              up.MigrationLatch,
		MigrationLatchDigest:        up.MigrationLatchDigest,
		AuthorityProtocolProfile:    contracts.AuthorityV7ProtocolProfileLegacyV6,
		CurrentCatalogDigest:        task8Digest("current-catalog"),
		ManifestID:                  task8UUID("manifest"),
		ManifestDigest:              task8Digest("manifest"),
		StableTableCount:            1,
		StableTableInventory:        task8StableInventory(),
		ProviderRetirementSet:       task8RetirementSet(now),
		ProviderRetirementSetDigest: task8Digest("provider-retirement-set"),
		EnvironmentAnchorSetDigest:  task8Digest("environment-anchor-set"),
		AuthorizationID:             task8UUID("authorization"),
		AuthorizationScope:          contracts.AuthorityV7DowngradeAuthorizationScopeDown00007Only,
		TransactionNonce:            task8Digest("down-transaction-nonce"),
		ExpiresAt:                   now.Add(5 * time.Minute),
	}
}

func task8DownPersistenceView(facts contracts.AuthorityV7DownMigrationFactsV1) authorityV7DownPersistenceView {
	return authorityV7DownPersistenceView{
		Facts:                               facts,
		ProviderRetirementSetBodyJCS:        []byte(`{"fixture":"provider-retirement-set"}`),
		ProviderRetirementEvidenceBundleJCS: []byte(`{"fixture":"provider-retirement-evidence"}`),
	}
}

func task8AuthorizationFacts(now time.Time) contracts.AuthorityV7DownAuthorizationFactsV1 {
	down := task8DownFacts(now)
	pristine := contracts.AuthorityV7PristineDowngradeInventoryFactsV1{
		InstallationID:              down.MigrationLatch.InstallationID,
		InstallationKind:            contracts.AuthorityV7InstallationKindDisposableFixture,
		DatabaseIdentityDigest:      down.MigrationLatch.DatabaseIdentityDigest,
		MigrationLatchDigest:        down.MigrationLatchDigest,
		CurrentCatalogDigest:        down.CurrentCatalogDigest,
		ManifestID:                  down.ManifestID,
		ManifestDigest:              down.ManifestDigest,
		StableTableCount:            1,
		StableTableInventory:        task8StableInventory(),
		ControlTableCount:           2,
		NonControlProtocolRowCount:  0,
		MigrationLatchCount:         1,
		DowngradeAuthorizationCount: 0,
		DatabaseTransactionID:       77,
		TransactionNonce:            down.TransactionNonce,
		Stage:                       contracts.AuthorityV7DowngradeStagePreAuthorization,
		ObservedAt:                  now.Add(-30 * time.Second),
	}
	request := contracts.AuthorityV7DownAuthorizationRequestV1{
		AuthorizationID:             down.AuthorizationID,
		MigrationLatch:              down.MigrationLatch,
		MigrationLatchDigest:        down.MigrationLatchDigest,
		CurrentCatalogDigest:        down.CurrentCatalogDigest,
		PristineInventory:           pristine,
		PristineInventoryDigest:     task8Digest("pristine-inventory"),
		ProviderRetirementSet:       down.ProviderRetirementSet,
		ProviderRetirementSetDigest: down.ProviderRetirementSetDigest,
		EnvironmentAnchorSetDigest:  down.EnvironmentAnchorSetDigest,
		ActualDatabaseTransactionID: pristine.DatabaseTransactionID,
		TransactionNonce:            down.TransactionNonce,
		AuthorizationScope:          down.AuthorizationScope,
	}
	authorization := contracts.AuthorityV7ProtocolDowngradeAuthorizationFactsV1{
		AuthorizationID:                              request.AuthorizationID,
		InstallationID:                               request.MigrationLatch.InstallationID,
		MigrationLatchDigest:                         request.MigrationLatchDigest,
		DatabaseIdentityDigest:                       request.MigrationLatch.DatabaseIdentityDigest,
		MigrationVersion:                             7,
		CurrentCatalogDigest:                         request.CurrentCatalogDigest,
		PristineDowngradeInventoryDigest:             request.PristineInventoryDigest,
		ProviderProtocolDowngradeRetirementSetDigest: request.ProviderRetirementSetDigest,
		EnvironmentInventoryAnchorSetDigest:          request.EnvironmentAnchorSetDigest,
		DatabaseTransactionID:                        request.ActualDatabaseTransactionID,
		TransactionNonce:                             request.TransactionNonce,
		AuthorizationScope:                           request.AuthorizationScope,
		IssuedAt:                                     now.Add(-time.Second),
		ExpiresAt:                                    now.Add(time.Minute),
	}
	return contracts.AuthorityV7DownAuthorizationFactsV1{
		Request:                  request,
		Authorization:            authorization,
		AuthorizationDigest:      task8Digest("authorization-body"),
		AuthorizationEnvelopeJCS: []byte(`{"body":"task8"}`),
		SignerRole:               "authority_protocol_downgrade_authorizer",
		SignerKeyID:              "task8-key",
		SignaturePolicyVersion:   1,
		TrustRootDigest:          task8Digest("trust-root"),
		SignatureAlgorithm:       "ed25519",
	}
}

func task8AuthorizationPersistenceView(now time.Time) authorityV7DownAuthorizationPersistenceView {
	facts := task8AuthorizationFacts(now)
	body := task8CanonicalAuthorizationBody(facts.Authorization)
	facts.AuthorizationDigest = authorityV7DomainDigest("authority-protocol-downgrade-authorization.v1", body)
	return authorityV7DownAuthorizationPersistenceView{
		Facts:                    facts,
		AuthorizationBodyJCS:     body,
		AuthorizationEnvelopeJCS: append([]byte(nil), facts.AuthorizationEnvelopeJCS...),
	}
}

func task8CanonicalAuthorizationBody(value contracts.AuthorityV7ProtocolDowngradeAuthorizationFactsV1) []byte {
	raw, err := json.Marshal(map[string]string{
		"authorization_id":                                  value.AuthorizationID.String(),
		"installation_id":                                   value.InstallationID.String(),
		"migration_latch_digest":                            hex.EncodeToString(value.MigrationLatchDigest[:]),
		"database_identity_digest":                          hex.EncodeToString(value.DatabaseIdentityDigest[:]),
		"migration_version":                                 strconv.FormatUint(value.MigrationVersion, 10),
		"current_catalog_digest":                            hex.EncodeToString(value.CurrentCatalogDigest[:]),
		"pristine_downgrade_inventory_digest":               hex.EncodeToString(value.PristineDowngradeInventoryDigest[:]),
		"provider_protocol_downgrade_retirement_set_digest": hex.EncodeToString(value.ProviderProtocolDowngradeRetirementSetDigest[:]),
		"environment_inventory_anchor_set_digest":           hex.EncodeToString(value.EnvironmentInventoryAnchorSetDigest[:]),
		"database_transaction_id":                           strconv.FormatUint(value.DatabaseTransactionID, 10),
		"transaction_nonce":                                 hex.EncodeToString(value.TransactionNonce[:]),
		"authorization_scope":                               string(value.AuthorizationScope),
		"issued_at":                                         value.IssuedAt.UTC().Format(time.RFC3339Nano),
		"expires_at":                                        value.ExpiresAt.UTC().Format(time.RFC3339Nano),
	})
	if err != nil {
		panic(err)
	}
	canonical, err := jcs.Transform(raw)
	if err != nil {
		panic(err)
	}
	return canonical
}

func task8FreshPayloadDigest(objectType contracts.FreshRestoreImportObjectTypeV1, payload []byte) contracts.Digest {
	hash := sha256.New()
	_, _ = hash.Write([]byte("talenro.c12.fresh-import-payload-" + string(objectType) + ".v1"))
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write(payload)
	var result contracts.Digest
	copy(result[:], hash.Sum(nil))
	return result
}

func task8AdmissionFacts(now time.Time) contracts.FreshRestoreImportProjectionInputV1 {
	manifestDigest := task8Digest("fresh-manifest")
	activationID := task8UUID("fresh-activation")
	registrationDigest := task8Digest("fresh-registration")
	lineageDigest := task8Digest("fresh-lineage")
	runtimeRebindDigest := task8Digest("fresh-runtime-rebind")
	runtimeBindingDigest := task8Digest("fresh-runtime-binding")
	databaseIdentityDigest := task8Digest("fresh-database-identity")
	routeDigest := task8Digest("fresh-route-closed")
	preInventoryDigest := task8Digest("fresh-pre-inventory")
	postInventoryDigest := task8Digest("fresh-post-inventory")
	servingAbsentDigest := task8Digest("fresh-serving-absent")
	applyID := task8UUID("fresh-apply")
	exclusionID := task8UUID("fresh-exclusion")
	capabilityDigest := task8Digest("fresh-capability")
	headDigest := task8Digest("fresh-held-head")
	return contracts.FreshRestoreImportProjectionInputV1{
		ManifestTopology: contracts.FreshRestoreImportManifestTopologyFactsV1{
			ManifestID:         task8UUID("fresh-manifest-id"),
			SingleUseApplyID:   applyID,
			ManifestDigest:     manifestDigest,
			TargetActivationID: activationID,
			TargetDeploymentID: task8UUID("fresh-deployment"),
			TargetDatabaseIncarnationRegistrationDigest: registrationDigest,
			TargetEpochEvidenceDigest:                   task8Digest("fresh-epoch-evidence"),
			IssuedAt:                                    now.Add(-time.Minute),
			ExpiresAt:                                   now.Add(10 * time.Minute),
			ObjectCount:                                 1,
			Objects: []contracts.FreshRestoreImportManifestObjectV1{{
				ObjectType:    contracts.FreshRestoreImportObjectPOP,
				CanonicalKey:  "tbs-1",
				Payload:       []byte(`{"iso_country":"GE","pop_code":"tbs-1","region":"tbilisi"}`),
				PayloadDigest: task8FreshPayloadDigest(contracts.FreshRestoreImportObjectPOP, []byte(`{"iso_country":"GE","pop_code":"tbs-1","region":"tbilisi"}`)),
			}},
			CompleteNodeSetDigest:             task8Digest("fresh-complete-node-set"),
			ForbiddenObjectClassSetDigest:     task8Digest("fresh-forbidden-object-class"),
			ExpectedPostImportInventoryDigest: postInventoryDigest,
		},
		StagingImportCapability: contracts.FreshRestoreStagingImportCapabilityFactsV1{
			CapabilityDigest:                            capabilityDigest,
			CapabilityID:                                task8UUID("fresh-capability-id"),
			SingleUseApplyID:                            applyID,
			ManifestDigest:                              manifestDigest,
			TargetActivationID:                          activationID,
			TargetDatabaseIdentityDigest:                databaseIdentityDigest,
			DatabaseTimelineLineageChainDigest:          lineageDigest,
			TargetDatabaseIncarnationRegistrationDigest: registrationDigest,
			RuntimeRebindChainDigest:                    runtimeRebindDigest,
			RuntimeInstanceBindingDigest:                runtimeBindingDigest,
			PlannedStagingExclusionID:                   exclusionID,
			ExpectedPreAcquireProviderHeadDigest:        task8Digest("fresh-pre-head"),
			PreAcquireDatabaseIncarnationProofDigest:    task8Digest("fresh-incarnation-proof"),
			ProviderPhase:                               contracts.FreshRestoreProviderPhaseStagingClosed,
			ProviderServingLeaseAbsentDigest:            servingAbsentDigest,
			DatabaseRouteClosedDigest:                   routeDigest,
			PreImportInventoryDigest:                    preInventoryDigest,
			AllowedObjectSetDigest:                      task8Digest("fresh-allowed-objects"),
			ExpectedPostImportInventoryDigest:           postInventoryDigest,
			TransactionNonce:                            task8Digest("fresh-capability-nonce"),
			IssuedAt:                                    now.Add(-time.Minute),
			ExpiresAt:                                   now.Add(4 * time.Minute),
		},
		StagingExclusionLease: contracts.FreshV7StagingExclusionLeaseFactsV1{
			LeaseDigest:                   task8Digest("fresh-lease"),
			ExclusionID:                   exclusionID,
			StagingImportCapabilityDigest: capabilityDigest,
			CapabilityRegistrationCommitChallengeDigest:   task8Digest("fresh-capability-challenge"),
			CapabilityRegistrationCommitAttestationDigest: task8Digest("fresh-capability-attestation"),
			TargetActivationID:                            activationID,
			TargetDatabaseIdentityDigest:                  databaseIdentityDigest,
			DatabaseTimelineLineageChainDigest:            lineageDigest,
			ExpectedProviderHeadDigest:                    task8Digest("fresh-pre-head"),
			PreAcquireDatabaseIncarnationProofDigest:      task8Digest("fresh-incarnation-proof"),
			TargetDatabaseIncarnationRegistrationDigest:   registrationDigest,
			RuntimeRebindChainDigest:                      runtimeRebindDigest,
			RuntimeInstanceBindingDigest:                  runtimeBindingDigest,
			RequestedAdmissionExpiresAt:                   now.Add(3 * time.Minute),
			RequestNonce:                                  task8Digest("fresh-request-nonce"),
			RequestDigest:                                 task8Digest("fresh-request"),
			ExclusionLeaseID:                              exclusionID,
			AcquisitionLockedProviderHeadDigest:           headDigest,
			ProviderServingLeaseAbsentDigest:              servingAbsentDigest,
			ProviderControlSequence:                       1,
			ExclusionState:                                contracts.FreshRestoreStagingExclusionHeld,
			IssuedAt:                                      now.Add(-time.Minute),
			AdmissionExpiresAt:                            now.Add(3 * time.Minute),
		},
		CurrentProviderHeadDigest:                    headDigest,
		ProviderPhase:                                contracts.FreshRestoreProviderPhaseStagingClosed,
		ServingLeaseAbsentDigest:                     servingAbsentDigest,
		StagingExclusionState:                        contracts.FreshRestoreStagingExclusionHeld,
		DatabaseRouteClosedDigest:                    routeDigest,
		CurrentDatabaseIdentityDigest:                databaseIdentityDigest,
		DatabaseTimelineLineageChainDigest:           lineageDigest,
		CurrentDatabaseIncarnationRegistrationDigest: registrationDigest,
		RuntimeRebindChainDigest:                     runtimeRebindDigest,
		RuntimeInstanceBindingDigest:                 runtimeBindingDigest,
		PreImportInventoryDigest:                     preInventoryDigest,
		NormalizedCatalogDigest:                      task8Digest("fresh-normalized-catalog"),
		AdmissionCheckedAt:                           now,
	}
}

func task8AdmissionPersistenceView(facts contracts.FreshRestoreImportProjectionInputV1) freshRestoreImportAdmissionPersistenceView {
	proofIssuedAt := facts.AdmissionCheckedAt
	proofExpiresAt := proofIssuedAt.Add(30 * time.Second)
	return freshRestoreImportAdmissionPersistenceView{
		Facts:                                    facts,
		StagingImportCapabilityBodyJCS:           []byte(`{"fixture":"capability-body"}`),
		StagingImportCapabilityEnvelopeJCS:       []byte(`{"fixture":"capability-envelope"}`),
		StagingImportCapabilityEvidenceBundleJCS: []byte(`{"fixture":"capability-bundle"}`),
		ManifestBodyJCS:                          []byte(`{"fixture":"manifest-body"}`),
		ManifestEnvelopeJCS:                      []byte(`{"fixture":"manifest-envelope"}`),
		ManifestEvidenceBundleJCS:                []byte(`{"fixture":"manifest-bundle"}`),
		StagingExclusionLeaseBodyJCS:             []byte(`{"fixture":"lease-body"}`),
		StagingExclusionLeaseEnvelopeJCS:         []byte(`{"fixture":"lease-envelope"}`),
		StagingExclusionLeaseEvidenceBundleJCS:   []byte(`{"fixture":"lease-bundle"}`),
		AcquisitionLockedProviderHeadBodyJCS:     []byte(`{"fixture":"provider-head-body"}`),
		AcquisitionLockedProviderHeadEnvelopeJCS: []byte(`{"fixture":"provider-head-envelope"}`),
		CurrentDatabaseIncarnationProofBodyJCS: task8CanonicalJSON(map[string]string{
			"issued_at":                         proofIssuedAt.UTC().Format(time.RFC3339Nano),
			"expires_at":                        proofExpiresAt.UTC().Format(time.RFC3339Nano),
			"attestor_runtime_lease_expires_at": proofExpiresAt.UTC().Format(time.RFC3339Nano),
		}),
		CurrentDatabaseIncarnationProofEnvelopeJCS:       []byte(`{"fixture":"incarnation-proof-envelope"}`),
		CurrentDatabaseIncarnationProofEvidenceBundleJCS: []byte(`{"fixture":"incarnation-proof-bundle"}`),
	}
}

func task8DownPrivateByteArms(value *authorityV7DownPersistenceView) [][]byte {
	return [][]byte{
		value.ProviderRetirementSetBodyJCS,
		value.ProviderRetirementEvidenceBundleJCS,
	}
}

func task8AuthorizationPrivateByteArms(value *authorityV7DownAuthorizationPersistenceView) [][]byte {
	return [][]byte{
		value.Facts.AuthorizationEnvelopeJCS,
		value.AuthorizationBodyJCS,
		value.AuthorizationEnvelopeJCS,
	}
}

func task8AdmissionPrivateByteArms(value *freshRestoreImportAdmissionPersistenceView) [][]byte {
	return [][]byte{
		value.StagingImportCapabilityBodyJCS,
		value.StagingImportCapabilityEnvelopeJCS,
		value.StagingImportCapabilityEvidenceBundleJCS,
		value.ManifestBodyJCS,
		value.ManifestEnvelopeJCS,
		value.ManifestEvidenceBundleJCS,
		value.StagingExclusionLeaseBodyJCS,
		value.StagingExclusionLeaseEnvelopeJCS,
		value.StagingExclusionLeaseEvidenceBundleJCS,
		value.AcquisitionLockedProviderHeadBodyJCS,
		value.AcquisitionLockedProviderHeadEnvelopeJCS,
		value.CurrentDatabaseIncarnationProofBodyJCS,
		value.CurrentDatabaseIncarnationProofEnvelopeJCS,
		value.CurrentDatabaseIncarnationProofEvidenceBundleJCS,
	}
}

func task8AllByteArmsStartWith(values [][]byte, expected byte) bool {
	for _, value := range values {
		if len(value) == 0 || value[0] != expected {
			return false
		}
	}
	return true
}

func task8FormatFields(fields []string) string {
	return fmt.Sprint(fields)
}

func task8AssertCopiedDownConsume(t *testing.T, facts contracts.AuthorityV7DownMigrationFactsV1) {
	t.Helper()
	grant, err := newVerifiedAuthorityV7DownGrant(task8DownPersistenceView(facts))
	if err != nil {
		t.Fatal(err)
	}
	copies := make([]VerifiedAuthorityV7DownGrant, 32)
	for index := range copies {
		copies[index] = grant
	}
	task8RunCopiedConsume(t, len(copies), func(index int) error {
		_, consumeErr := ConsumeVerifiedAuthorityV7DownGrant(copies[index])
		return consumeErr
	})
}

func task8AssertCopiedAuthorizationConsume(t *testing.T, now time.Time) {
	t.Helper()
	authorization, err := newVerifiedAuthorityV7DownAuthorization(task8AuthorizationPersistenceView(now))
	if err != nil {
		t.Fatal(err)
	}
	copies := make([]VerifiedAuthorityV7DownAuthorization, 32)
	for index := range copies {
		copies[index] = authorization
	}
	task8RunCopiedConsume(t, len(copies), func(index int) error {
		_, consumeErr := ConsumeVerifiedAuthorityV7DownAuthorization(copies[index])
		return consumeErr
	})
}

func task8AssertCopiedAdmissionConsume(t *testing.T, facts contracts.FreshRestoreImportProjectionInputV1) {
	t.Helper()
	admission, err := newVerifiedFreshRestoreImportAdmission(task8AdmissionPersistenceView(facts))
	if err != nil {
		t.Fatal(err)
	}
	copies := make([]VerifiedFreshRestoreImportAdmission, 32)
	for index := range copies {
		copies[index] = admission
	}
	task8RunCopiedConsume(t, len(copies), func(index int) error {
		_, consumeErr := consumeVerifiedFreshRestoreImportAdmission(copies[index])
		return consumeErr
	})
}

func task8RunCopiedConsume(t *testing.T, count int, consume func(int) error) {
	t.Helper()
	start := make(chan struct{})
	results := make(chan error, count)
	var wait sync.WaitGroup
	for index := 0; index < count; index++ {
		wait.Add(1)
		go func(copyIndex int) {
			defer wait.Done()
			<-start
			results <- consume(copyIndex)
		}(index)
	}
	close(start)
	wait.Wait()
	close(results)
	successes, conflicts := 0, 0
	for consumeErr := range results {
		switch {
		case consumeErr == nil:
			successes++
		case errors.Is(consumeErr, ErrConflict):
			conflicts++
		default:
			t.Fatalf("32-way copied consume error = %v", consumeErr)
		}
	}
	if successes != 1 || conflicts != count-1 {
		t.Fatalf("%d-way copied consume results = success:%d conflict:%d, want 1/%d", count, successes, conflicts, count-1)
	}
}

func task8AssertStrictAuthorityV7Verifier(t *testing.T) {
	t.Helper()
	t.Run("closed fixture signer-role registry", func(t *testing.T) {
		want := map[string]string{
			"environment-inventory-membership-retirement.v1":          "release_deployment_operator",
			"fresh-restore-import-manifest.v1":                        "fresh_restore_export_operator",
			"provider-protocol-downgrade-retirement-authorization.v1": "provider_downgrade_retirement_admin",
			"provider-protocol-history-zero-projection.v1":            "claim_v1_provider_history_auditor",
			"provider-protocol-downgrade-retirement.v1":               "claim_v1_provider",
			"claim-v1-provider-head.v1":                               "claim_v1_provider",
			"fresh-v7-staging-exclusion-lease.v1":                     "claim_v1_provider",
			"database-incarnation-proof.v1":                           "claim_v1_provider",
			"authority-protocol-downgrade-authorization.v1":           "authority_protocol_downgrade_authorizer",
			"staging-import-capability.v1":                            "fresh_restore_staging_import_authorizer",
		}
		for schema, role := range want {
			if got := authorityV7FixtureSignerRole(schema); got != role {
				t.Errorf("role for %s = %q, want %q", schema, got, role)
			}
		}
		if got := authorityV7FixtureSignerRole("unknown.v1"); got != "" {
			t.Fatalf("unknown schema role = %q, want closed rejection", got)
		}
	})

	t.Run("history-zero window starts at observed_at", func(t *testing.T) {
		now := time.Now().UTC()
		body := task8CanonicalJSON(map[string]string{
			"observed_at": now.Add(-time.Second).Format(time.RFC3339Nano),
			"expires_at":  now.Add(time.Minute).Format(time.RFC3339Nano),
		})
		expiresAt, err := validateAuthorityV7CanonicalWindowAt(body, "observed_at", 2*time.Minute, now)
		if err != nil || !expiresAt.Equal(now.Add(time.Minute)) {
			t.Fatalf("valid history-zero observed window = %v", err)
		}
		if _, err := validateAuthorityV7CanonicalWindowAt(body, "issued_at", 2*time.Minute, now); !errors.Is(err, ErrInvalidArgument) {
			t.Fatalf("missing issued_at window = %v, want ErrInvalidArgument", err)
		}
		startEqual := task8CanonicalJSON(map[string]string{
			"issued_at":  now.Format(time.RFC3339Nano),
			"expires_at": now.Add(time.Second).Format(time.RFC3339Nano),
		})
		if _, err := validateAuthorityV7CanonicalWindowAt(startEqual, "issued_at", 2*time.Minute, now); err != nil {
			t.Fatalf("issued-at equality boundary = %v", err)
		}
		future := task8CanonicalJSON(map[string]string{
			"issued_at":  now.Add(time.Nanosecond).Format(time.RFC3339Nano),
			"expires_at": now.Add(time.Second).Format(time.RFC3339Nano),
		})
		if _, err := validateAuthorityV7CanonicalWindowAt(future, "issued_at", 2*time.Minute, now); !errors.Is(err, ErrInvalidArgument) {
			t.Fatalf("future-issued equality boundary = %v, want ErrInvalidArgument", err)
		}
		expiryEqual := task8CanonicalJSON(map[string]string{
			"issued_at":  now.Add(-time.Second).Format(time.RFC3339Nano),
			"expires_at": now.Format(time.RFC3339Nano),
		})
		if _, err := validateAuthorityV7CanonicalWindowAt(expiryEqual, "issued_at", 2*time.Minute, now); !errors.Is(err, ErrInvalidArgument) {
			t.Fatalf("expiry equality boundary = %v, want ErrInvalidArgument", err)
		}
	})

	t.Run("strict structural JSON and JCS gate", func(t *testing.T) {
		spec := authorityV7StrictJSONSpec{Fields: []string{"nested", "outer"}}
		valid := []byte(`{"nested":{"value":"ok"},"outer":"ok"}`)
		if _, err := validateAuthorityV7StrictJSON(valid, spec); err != nil {
			t.Fatalf("strict positive = %v", err)
		}
		mutations := [][]byte{
			nil,
			bytes.Repeat([]byte{'x'}, authorityV7CanonicalArtifactMaxBytes+1),
			{0xff},
			append([]byte{0xef, 0xbb, 0xbf}, valid...),
			[]byte(`{"nested":{"value":"ok","value":"duplicate"},"outer":"ok"}`),
			[]byte(`{"nested":{"value":"ok"},"outer":"first","outer":"second"}`),
			[]byte(`{"nested":{"value":"ok"},"outer":"ok"} {}`),
			[]byte(`{"nested":{"value":"ok"},"other":"extra","outer":"ok"}`),
			[]byte(`{"nested":{"value":"ok"}}`),
			[]byte(`{"Nested":{"value":"ok"},"outer":"ok"}`),
			[]byte(`{"nested":null,"outer":"ok"}`),
			[]byte(` {"nested":{"value":"ok"},"outer":"ok"}`),
			[]byte(`[]`),
		}
		for index, mutation := range mutations {
			if _, err := validateAuthorityV7StrictJSON(mutation, spec); !errors.Is(err, ErrInvalidArgument) {
				t.Errorf("strict mutation %d = %v, want ErrInvalidArgument", index, err)
			}
		}
	})

	view := task8AuthorizationPersistenceView(time.Now().UTC())
	bodyExpectation := authorityV7CanonicalBodyExpectation{
		Schema:          "authority-protocol-downgrade-authorization.v1",
		Fields:          task8AuthorizationBodyFields(),
		ExpectedBodyJCS: view.AuthorizationBodyJCS,
		ExpectedDigest:  view.Facts.AuthorizationDigest,
	}
	policy := task8AuthorityV7TestPolicy(bodyExpectation.Schema, "authority_protocol_downgrade_authorizer")
	envelope := task8SignAuthorityV7Envelope(view.AuthorizationBodyJCS, bodyExpectation.ExpectedDigest, policy, false)
	envelopeExpectation := authorityV7EnvelopeExpectation{Body: bodyExpectation, EnvelopeJCS: envelope, Policy: policy}
	if err := verifyAuthorityV7SignedEnvelope(view.AuthorizationBodyJCS, envelope, envelopeExpectation); err != nil {
		t.Fatalf("strict signed-envelope positive = %v", err)
	}

	t.Run("signed envelope mutations", func(t *testing.T) {
		for _, mutation := range []struct {
			name   string
			mutate func(map[string]json.RawMessage)
		}{
			{"schema", task8SetJSONField("schema", `"other.v1"`)},
			{"body digest", task8SetJSONField("body_digest", `"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"`)},
			{"role", task8SetJSONField("signer_role", `"wrong_role"`)},
			{"key", task8SetJSONField("signer_key_id", `"wrong-key"`)},
			{"policy", task8SetJSONField("signature_policy_version", `"01"`)},
			{"root", task8SetJSONField("trust_root_digest", `"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"`)},
			{"algorithm", task8SetJSONField("signature_algorithm", `"ecdsa-p256-sha256"`)},
			{"signature", task8SetJSONField("signature", `"AA"`)},
			{"padded signature", func(object map[string]json.RawMessage) {
				text, _ := authorityV7JSONString(object["signature"])
				object["signature"] = task8JSONRaw(text + "=")
			}},
			{"body", task8SetJSONField("body", `{"unexpected":"body"}`)},
			{"null body", task8SetJSONField("body", `null`)},
			{"unknown", task8SetJSONField("unknown", `"extra"`)},
			{"missing", func(object map[string]json.RawMessage) { delete(object, "signer_role") }},
		} {
			t.Run(mutation.name, func(t *testing.T) {
				mutated := task8MutateCanonicalObject(envelope, mutation.mutate)
				if err := verifyAuthorityV7SignedEnvelope(view.AuthorizationBodyJCS, mutated, authorityV7EnvelopeExpectation{Body: bodyExpectation, EnvelopeJCS: mutated, Policy: policy}); !errors.Is(err, ErrInvalidArgument) {
					t.Fatalf("mutation error = %v, want ErrInvalidArgument", err)
				}
			})
		}
		for name, mutated := range map[string][]byte{
			"duplicate":              bytes.Replace(envelope, []byte(`"body_digest":`), []byte(`"body_digest":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","body_digest":`), 1),
			"noncanonical":           append([]byte{' '}, envelope...),
			"trailing":               append(cloneAuthorityV7Bytes(envelope), []byte(`{}`)...),
			"metadata includes body": task8SignAuthorityV7Envelope(view.AuthorizationBodyJCS, bodyExpectation.ExpectedDigest, policy, true),
		} {
			if err := verifyAuthorityV7SignedEnvelope(view.AuthorizationBodyJCS, mutated, authorityV7EnvelopeExpectation{Body: bodyExpectation, EnvelopeJCS: mutated, Policy: policy}); !errors.Is(err, ErrInvalidArgument) {
				t.Errorf("%s envelope = %v, want ErrInvalidArgument", name, err)
			}
		}
	})

	t.Run("exact evidence registry and nullable arms", func(t *testing.T) {
		databaseBody := []byte(`{"value":"database"}`)
		databaseExpectation := authorityV7CanonicalBodyExpectation{
			Schema:          "task8-database-body.v1",
			Fields:          []string{"value"},
			ExpectedBodyJCS: databaseBody,
			ExpectedDigest:  authorityV7DomainDigest("task8-database-body.v1", databaseBody),
		}
		members := task8SortedEvidenceMembers([]authorityV7EvidenceMemberExpectation{
			{EvidenceKind: "external_signed_envelope", Envelope: &envelopeExpectation},
			{EvidenceKind: "database_immutable_body", Body: &databaseExpectation},
		})
		bundle := task8AuthorityV7EvidenceBundle(bodyExpectation.Schema, bodyExpectation.ExpectedDigest, members)
		if err := verifyAuthorityV7EvidenceBundle(bundle, bodyExpectation.Schema, bodyExpectation.ExpectedDigest, members); err != nil {
			t.Fatalf("strict evidence positive = %v", err)
		}
		for _, mutation := range []struct {
			name   string
			mutate func(map[string]json.RawMessage)
		}{
			{"message schema", task8SetJSONField("message_schema", `"other.v1"`)},
			{"message digest", task8SetJSONField("message_body_digest", `"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"`)},
			{"count", task8SetJSONField("evidence_count", `"02"`)},
			{"missing member", func(object map[string]json.RawMessage) {
				var items []json.RawMessage
				_ = json.Unmarshal(object["evidence"], &items)
				object["evidence"] = task8CanonicalJSON(items[:1])
			}},
			{"duplicate member", func(object map[string]json.RawMessage) {
				var items []json.RawMessage
				_ = json.Unmarshal(object["evidence"], &items)
				object["evidence"] = task8CanonicalJSON([]json.RawMessage{items[0], items[0]})
			}},
			{"reordered", func(object map[string]json.RawMessage) {
				var items []json.RawMessage
				_ = json.Unmarshal(object["evidence"], &items)
				object["evidence"] = task8CanonicalJSON([]json.RawMessage{items[1], items[0]})
			}},
			{"wrong kind", task8MutateEvidenceItem("database_immutable_body", "evidence_kind", `"external_signed_envelope"`)},
			{"wrong schema", task8MutateFirstEvidenceItem("schema", `"other.v1"`)},
			{"wrong digest", task8MutateFirstEvidenceItem("body_digest", `"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"`)},
			{"both null", task8MutateEvidenceItem("external_signed_envelope", "canonical_envelope_or_null", `null`)},
			{"both present", task8MutateEvidenceItem("external_signed_envelope", "canonical_body_or_null", `{"value":"extra"}`)},
			{"extra top field", task8SetJSONField("extra", `"value"`)},
		} {
			t.Run(mutation.name, func(t *testing.T) {
				mutated := task8MutateCanonicalObject(bundle, mutation.mutate)
				if err := verifyAuthorityV7EvidenceBundle(mutated, bodyExpectation.Schema, bodyExpectation.ExpectedDigest, members); !errors.Is(err, ErrInvalidArgument) {
					t.Fatalf("bundle mutation = %v, want ErrInvalidArgument", err)
				}
			})
		}
		for name, mutated := range map[string][]byte{
			"noncanonical":  append([]byte{' '}, bundle...),
			"trailing":      append(cloneAuthorityV7Bytes(bundle), []byte(`{}`)...),
			"duplicate key": bytes.Replace(bundle, []byte(`"evidence_count":`), []byte(`"evidence_count":"2","evidence_count":`), 1),
		} {
			if err := verifyAuthorityV7EvidenceBundle(mutated, bodyExpectation.Schema, bodyExpectation.ExpectedDigest, members); !errors.Is(err, ErrInvalidArgument) {
				t.Errorf("%s bundle = %v, want ErrInvalidArgument", name, err)
			}
		}
	})
}

func task8AuthorizationBodyFields() []string {
	return []string{
		"authorization_id", "installation_id", "migration_latch_digest", "database_identity_digest", "migration_version", "current_catalog_digest",
		"pristine_downgrade_inventory_digest", "provider_protocol_downgrade_retirement_set_digest", "environment_inventory_anchor_set_digest",
		"database_transaction_id", "transaction_nonce", "authorization_scope", "issued_at", "expires_at",
	}
}

func task8AuthorityV7TestPolicy(schema, role string) authorityV7SignerPolicy {
	seed := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31}
	return authorityV7SignerPolicy{
		Schema:                 schema,
		SignerRole:             role,
		SignerKeyID:            "56475aa75463474c0285df5dbf2bcab73da651358839e9b77481b2eab107708c",
		SignaturePolicyVersion: "1",
		TrustRootDigest:        strings.Repeat("a1", 32),
		SignatureAlgorithm:     "ed25519",
		PublicKey:              ed25519.NewKeyFromSeed(seed).Public().(ed25519.PublicKey),
	}
}

func task8SignAuthorityV7Envelope(body []byte, digest contracts.Digest, policy authorityV7SignerPolicy, includeBody bool) []byte {
	metadata := map[string]any{
		"schema": policy.Schema, "body_digest": hex.EncodeToString(digest[:]), "signer_role": policy.SignerRole,
		"signer_key_id": policy.SignerKeyID, "signature_policy_version": policy.SignaturePolicyVersion,
		"trust_root_digest": policy.TrustRootDigest, "signature_algorithm": policy.SignatureAlgorithm,
	}
	if includeBody {
		metadata["body"] = json.RawMessage(body)
	}
	metadataJCS := task8CanonicalJSON(metadata)
	input := append([]byte("talenro.c12.signature-envelope.v1"), 0)
	input = append(input, metadataJCS...)
	seed := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31}
	signature := ed25519.Sign(ed25519.NewKeyFromSeed(seed), input)
	return task8CanonicalJSON(map[string]any{
		"schema": policy.Schema, "body": json.RawMessage(body), "body_digest": hex.EncodeToString(digest[:]),
		"signer_role": policy.SignerRole, "signer_key_id": policy.SignerKeyID, "signature_policy_version": policy.SignaturePolicyVersion,
		"trust_root_digest": policy.TrustRootDigest, "signature_algorithm": policy.SignatureAlgorithm,
		"signature": base64.RawURLEncoding.EncodeToString(signature),
	})
}

func task8AuthorityV7EvidenceBundle(messageSchema string, messageDigest contracts.Digest, members []authorityV7EvidenceMemberExpectation) []byte {
	items := make([]json.RawMessage, len(members))
	for index, member := range members {
		if member.Body != nil {
			items[index] = task8CanonicalJSON(map[string]any{
				"evidence_kind": "database_immutable_body", "schema": member.Body.Schema,
				"body_digest": hex.EncodeToString(member.Body.ExpectedDigest[:]), "canonical_body_or_null": json.RawMessage(member.Body.ExpectedBodyJCS),
				"canonical_envelope_or_null": nil,
			})
			continue
		}
		items[index] = task8CanonicalJSON(map[string]any{
			"evidence_kind": "external_signed_envelope", "schema": member.Envelope.Body.Schema,
			"body_digest": hex.EncodeToString(member.Envelope.Body.ExpectedDigest[:]), "canonical_body_or_null": nil,
			"canonical_envelope_or_null": json.RawMessage(member.Envelope.EnvelopeJCS),
		})
	}
	return task8CanonicalJSON(map[string]any{
		"message_schema": messageSchema, "message_body_digest": hex.EncodeToString(messageDigest[:]),
		"evidence_count": strconv.Itoa(len(items)), "evidence": items,
	})
}

func task8SortedEvidenceMembers(values []authorityV7EvidenceMemberExpectation) []authorityV7EvidenceMemberExpectation {
	result := append([]authorityV7EvidenceMemberExpectation(nil), values...)
	sort.Slice(result, func(left, right int) bool {
		leftSchema, leftDigest := task8EvidenceMemberKey(result[left])
		rightSchema, rightDigest := task8EvidenceMemberKey(result[right])
		return authorityV7CompareEvidenceKeys(leftDigest, leftSchema, rightDigest, rightSchema) < 0
	})
	return result
}

func task8EvidenceMemberKey(value authorityV7EvidenceMemberExpectation) (string, contracts.Digest) {
	if value.Body != nil {
		return value.Body.Schema, value.Body.ExpectedDigest
	}
	return value.Envelope.Body.Schema, value.Envelope.Body.ExpectedDigest
}

func task8MutateCanonicalObject(raw []byte, mutate func(map[string]json.RawMessage)) []byte {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		panic(err)
	}
	mutate(object)
	return task8CanonicalJSON(object)
}

func task8SetJSONField(field, raw string) func(map[string]json.RawMessage) {
	return func(object map[string]json.RawMessage) { object[field] = json.RawMessage(raw) }
}

func task8MutateFirstEvidenceItem(field, raw string) func(map[string]json.RawMessage) {
	return func(object map[string]json.RawMessage) {
		var items []json.RawMessage
		if err := json.Unmarshal(object["evidence"], &items); err != nil {
			panic(err)
		}
		items[0] = task8MutateCanonicalObject(items[0], task8SetJSONField(field, raw))
		object["evidence"] = task8CanonicalJSON(items)
	}
}

func task8MutateEvidenceItem(targetKind, field, raw string) func(map[string]json.RawMessage) {
	return func(object map[string]json.RawMessage) {
		var items []json.RawMessage
		if err := json.Unmarshal(object["evidence"], &items); err != nil {
			panic(err)
		}
		for index, item := range items {
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(item, &fields); err != nil {
				panic(err)
			}
			kind, _ := authorityV7JSONString(fields["evidence_kind"])
			if kind == targetKind {
				fields[field] = json.RawMessage(raw)
				items[index] = task8CanonicalJSON(fields)
				object["evidence"] = task8CanonicalJSON(items)
				return
			}
		}
		panic("target evidence kind not found")
	}
}

func task8MutateFactoryEnvelopeField(envelope []byte, field string, value any) []byte {
	return task8MutateCanonicalObject(envelope, func(object map[string]json.RawMessage) {
		object[field] = task8CanonicalJSON(value)
	})
}

func task8MutateFactoryBundleField(bundle []byte, field string, value any) []byte {
	return task8MutateCanonicalObject(bundle, func(object map[string]json.RawMessage) {
		object[field] = task8CanonicalJSON(value)
	})
}

func task8MutateFactoryBundleMemberField(bundle []byte, field string, value any) []byte {
	return task8MutateCanonicalObject(bundle, func(object map[string]json.RawMessage) {
		var evidence []json.RawMessage
		if err := json.Unmarshal(object["evidence"], &evidence); err != nil || len(evidence) == 0 {
			panic("factory bundle has no evidence")
		}
		evidence[0] = task8MutateCanonicalObject(evidence[0], func(member map[string]json.RawMessage) {
			member[field] = task8CanonicalJSON(value)
		})
		object["evidence"] = task8CanonicalJSON(evidence)
	})
}

func task8MutateFactoryBundleEnvelopeField(bundle []byte, field string, value any) []byte {
	return task8MutateCanonicalObject(bundle, func(object map[string]json.RawMessage) {
		var evidence []json.RawMessage
		if err := json.Unmarshal(object["evidence"], &evidence); err != nil {
			panic(err)
		}
		for index := range evidence {
			var member map[string]json.RawMessage
			if err := json.Unmarshal(evidence[index], &member); err != nil {
				panic(err)
			}
			var envelope any
			if raw := member["canonical_envelope_or_null"]; len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
				continue
			} else if err := json.Unmarshal(raw, &envelope); err != nil {
				panic(err)
			}
			var envelopeObject map[string]json.RawMessage
			if err := json.Unmarshal(member["canonical_envelope_or_null"], &envelopeObject); err != nil {
				panic(err)
			}
			envelopeObject[field] = task8CanonicalJSON(value)
			member["canonical_envelope_or_null"] = task8CanonicalJSON(envelopeObject)
			evidence[index] = task8CanonicalJSON(member)
			object["evidence"] = task8CanonicalJSON(evidence)
			return
		}
		panic("factory bundle has no signed-envelope evidence")
	})
}

func task8ReverseFactoryBundleEvidence(bundle []byte) []byte {
	return task8MutateCanonicalObject(bundle, func(object map[string]json.RawMessage) {
		var evidence []json.RawMessage
		if err := json.Unmarshal(object["evidence"], &evidence); err != nil || len(evidence) < 2 {
			panic("factory bundle needs at least two evidence members")
		}
		for left, right := 0, len(evidence)-1; left < right; left, right = left+1, right-1 {
			evidence[left], evidence[right] = evidence[right], evidence[left]
		}
		object["evidence"] = task8CanonicalJSON(evidence)
	})
}

type task8FactoryEnvelopeMutation struct {
	Name  string
	Field string
	Value any
}

func task8FactoryEnvelopeMutations() []task8FactoryEnvelopeMutation {
	return []task8FactoryEnvelopeMutation{
		{Name: "signature", Field: "signature", Value: base64.RawURLEncoding.EncodeToString(make([]byte, ed25519.SignatureSize))},
		{Name: "role", Field: "signer_role", Value: "unregistered_task8_role"},
		{Name: "key", Field: "signer_key_id", Value: strings.Repeat("b2", 32)},
		{Name: "policy", Field: "signature_policy_version", Value: "2"},
		{Name: "root", Field: "trust_root_digest", Value: strings.Repeat("b2", 32)},
		{Name: "algorithm", Field: "signature_algorithm", Value: "ed25519ph"},
	}
}

func task8CanonicalInvalidDownFactoryInput(index int, raw []byte) []byte {
	switch index {
	case 0:
		return task8MutateFactoryEnvelopeField(raw, "created_at", "2030-01-01T00:00:00Z")
	case 1:
		return task8MutateFactoryBundleField(raw, "message_body_digest", strings.Repeat("00", 32))
	default:
		panic("unknown Down factory input")
	}
}

func task8CanonicalInvalidStagingFactoryInput(index int, raw []byte) []byte {
	switch index {
	case 0:
		return task8MutateFactoryEnvelopeField(raw, "transaction_nonce", strings.Repeat("b2", 32))
	case 3:
		return task8MutateFactoryEnvelopeField(raw, "source_archive_point", "wrong/archive/point")
	case 6:
		return task8MutateFactoryEnvelopeField(raw, "request_nonce", strings.Repeat("b2", 32))
	case 9:
		return task8MutateFactoryEnvelopeField(raw, "current_credential_policy_digest", strings.Repeat("b2", 32))
	case 11:
		return task8MutateFactoryEnvelopeField(raw, "request_digest", strings.Repeat("b2", 32))
	case 1, 4, 7, 10, 12:
		return task8MutateFactoryEnvelopeField(raw, "signature", base64.RawURLEncoding.EncodeToString(make([]byte, ed25519.SignatureSize)))
	case 2:
		return task8MutateFactoryBundleField(raw, "message_body_digest", strings.Repeat("00", 32))
	case 5:
		return task8MutateFactoryBundleField(raw, "evidence_count", "2")
	case 8:
		return task8MutateFactoryBundleMemberField(raw, "body_digest", strings.Repeat("00", 32))
	case 13:
		return task8MutateFactoryBundleField(raw, "message_schema", "wrong-schema.v1")
	default:
		panic("unknown staging factory input")
	}
}

func task8JSONRaw(value string) json.RawMessage {
	raw, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return raw
}

func task8CanonicalJSON(value any) []byte {
	raw, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	canonical, err := jcs.Transform(raw)
	if err != nil {
		panic(err)
	}
	return canonical
}

type task8FactorySignedArtifact struct {
	Schema   string
	Digest   contracts.Digest
	BodyJCS  []byte
	Envelope []byte
}

type task8FactoryEvidenceMember struct {
	Kind     string
	Schema   string
	Digest   contracts.Digest
	BodyJCS  []byte
	Envelope []byte
}

type task8DownFactoryFixture struct {
	Facts  contracts.AuthorityV7DownMigrationFactsV1
	Inputs [2][]byte
}

type task8DownEvidenceWindows struct {
	MembershipIssuedAt  time.Time
	MembershipExpiresAt time.Time
	AdminIssuedAt       time.Time
	AdminExpiresAt      time.Time
	HistoryObservedAt   time.Time
	HistoryExpiresAt    time.Time
}

type task8AuthorizationFactoryFixture struct {
	Facts contracts.AuthorityV7DownAuthorizationFactsV1
	Body  []byte
}

type task8StagingFactoryFixture struct {
	Facts  contracts.FreshRestoreImportProjectionInputV1
	Inputs [14][]byte
}

func task8FactoryDomainDigest(schema string, bodyJCS []byte) contracts.Digest {
	hash := sha256.New()
	_, _ = hash.Write([]byte("talenro.c12." + schema))
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write(bodyJCS)
	var result contracts.Digest
	copy(result[:], hash.Sum(nil))
	return result
}

func task8FactorySign(schema, role string, bodyJCS []byte) task8FactorySignedArtifact {
	digest := task8FactoryDomainDigest(schema, bodyJCS)
	policy := task8AuthorityV7TestPolicy(schema, role)
	return task8FactorySignedArtifact{
		Schema: schema, Digest: digest, BodyJCS: bodyJCS,
		Envelope: task8SignAuthorityV7Envelope(bodyJCS, digest, policy, false),
	}
}

func task8FactoryBundle(messageSchema string, messageDigest contracts.Digest, members ...task8FactoryEvidenceMember) []byte {
	sorted := append([]task8FactoryEvidenceMember(nil), members...)
	sort.Slice(sorted, func(left, right int) bool {
		if compared := bytes.Compare(sorted[left].Digest[:], sorted[right].Digest[:]); compared != 0 {
			return compared < 0
		}
		return sorted[left].Schema < sorted[right].Schema
	})
	items := make([]map[string]any, len(sorted))
	for index, member := range sorted {
		item := map[string]any{
			"evidence_kind": member.Kind, "schema": member.Schema, "body_digest": hex.EncodeToString(member.Digest[:]),
			"canonical_body_or_null": nil, "canonical_envelope_or_null": nil,
		}
		switch member.Kind {
		case "database_immutable_body":
			item["canonical_body_or_null"] = json.RawMessage(member.BodyJCS)
		case "external_signed_envelope":
			item["canonical_envelope_or_null"] = json.RawMessage(member.Envelope)
		default:
			panic("unknown task8 evidence kind")
		}
		items[index] = item
	}
	return task8CanonicalJSON(map[string]any{
		"message_schema": messageSchema, "message_body_digest": hex.EncodeToString(messageDigest[:]),
		"evidence_count": strconv.Itoa(len(items)), "evidence": items,
	})
}

func task8FactoryExternal(value task8FactorySignedArtifact) task8FactoryEvidenceMember {
	return task8FactoryEvidenceMember{Kind: "external_signed_envelope", Schema: value.Schema, Digest: value.Digest, Envelope: value.Envelope}
}

func task8FactoryDigest(value contracts.Digest) string { return hex.EncodeToString(value[:]) }

func task8FactoryTime(value time.Time) string { return value.UTC().Format(time.RFC3339Nano) }

func task8FactoryRetiredMemberWire(value contracts.AuthorityV7RetiredEnvironmentMemberFactsV1) map[string]any {
	return map[string]any{
		"environment_record_digest": task8FactoryDigest(value.EnvironmentRecordDigest), "environment_attestation_digest": task8FactoryDigest(value.EnvironmentAttestationDigest),
		"deployment_id": value.DeploymentID.String(), "postgres_system_id": strconv.FormatUint(value.PostgresSystemID, 10),
		"timeline": strconv.FormatUint(value.Timeline, 10), "database_oid": strconv.FormatUint(value.DatabaseOID, 10), "database_name": value.DatabaseName,
		"database_identity_digest": task8FactoryDigest(value.DatabaseIdentityDigest), "environment_instance_generation": strconv.FormatUint(value.EnvironmentInstanceGeneration, 10),
		"provider_identity_digest": task8FactoryDigest(value.ProviderIdentityDigest), "provider_endpoint_identity_digest": task8FactoryDigest(value.ProviderEndpointIdentityDigest),
		"provider_namespace": value.ProviderNamespace, "provider_profile": value.ProviderProfile,
	}
}

func task8FactoryRetirementSetBody(value contracts.AuthorityV7ProviderRetirementSetFactsV1) []byte {
	retirements := make([]map[string]any, len(value.Retirements))
	for index, retirement := range value.Retirements {
		retirements[index] = map[string]any{
			"provider_identity_digest":                      task8FactoryDigest(retirement.ProviderIdentityDigest),
			"provider_endpoint_identity_digest":             task8FactoryDigest(retirement.ProviderEndpointIdentityDigest),
			"provider_protocol_downgrade_retirement_digest": task8FactoryDigest(retirement.ProviderProtocolDowngradeRetirementDigest),
		}
	}
	members := make([]map[string]any, len(value.RetiredMembers))
	for index, member := range value.RetiredMembers {
		members[index] = task8FactoryRetiredMemberWire(member)
	}
	return task8CanonicalJSON(map[string]any{
		"installation_id": value.InstallationID.String(), "release_scope": value.ReleaseScope,
		"environment_inventory_digest":                       task8FactoryDigest(value.EnvironmentInventoryDigest),
		"environment_inventory_anchor_set_digest":            task8FactoryDigest(value.EnvironmentInventoryAnchorSetDigest),
		"environment_inventory_membership_retirement_digest": task8FactoryDigest(value.EnvironmentInventoryMembershipRetirementDigest),
		"provider_count":                                     strconv.FormatUint(value.ProviderCount, 10), "retirements": retirements,
		"retired_member_count": strconv.FormatUint(value.RetiredMemberCount, 10), "retired_members": members,
		"created_at": task8FactoryTime(value.CreatedAt),
	})
}

func task8BuildDownFactoryFixture(now time.Time) task8DownFactoryFixture {
	return task8BuildDownFactoryFixtureWithWindows(now, task8DownEvidenceWindows{
		MembershipIssuedAt:  now.Add(-time.Minute),
		MembershipExpiresAt: now.Add(time.Minute),
		AdminIssuedAt:       now.Add(-time.Minute),
		AdminExpiresAt:      now.Add(time.Minute),
		HistoryObservedAt:   now.Add(-time.Minute),
		HistoryExpiresAt:    now.Add(time.Minute),
	})
}

func task8BuildDownFactoryFixtureWithWindows(now time.Time, windows task8DownEvidenceWindows) task8DownFactoryFixture {
	facts := task8DownFacts(now)
	facts.ExpiresAt = windows.MembershipExpiresAt
	for _, candidate := range []time.Time{windows.AdminExpiresAt, windows.HistoryExpiresAt} {
		if candidate.Before(facts.ExpiresAt) {
			facts.ExpiresAt = candidate
		}
	}
	set := &facts.ProviderRetirementSet
	member := set.RetiredMembers[0]
	retirement := &set.Retirements[0]
	membershipID := task8UUID("factory-membership-retirement")
	retirementID := task8UUID("factory-provider-retirement")
	tombstoneID := task8UUID("factory-retirement-tombstone")
	memberSetDigest := task8FactoryDomainDigest(
		"environment-inventory-member-set.v1",
		task8CanonicalJSON([]string{task8FactoryDigest(member.EnvironmentRecordDigest)}),
	)
	membershipBody := task8CanonicalJSON(map[string]any{
		"membership_retirement_id": membershipID.String(), "installation_id": set.InstallationID.String(), "installation_kind": "disposable_fixture",
		"migration_latch_digest": task8FactoryDigest(facts.MigrationLatchDigest), "database_identity_digest": task8FactoryDigest(facts.MigrationLatch.DatabaseIdentityDigest),
		"release_scope": set.ReleaseScope, "final_environment_inventory_digest": task8FactoryDigest(set.EnvironmentInventoryDigest), "final_inventory_sequence": "1",
		"final_environment_inventory_anchor_set_digest": task8FactoryDigest(set.EnvironmentInventoryAnchorSetDigest),
		"environment_member_set_digest":                 task8FactoryDigest(memberSetDigest), "environment_count": strconv.FormatUint(set.RetiredMemberCount, 10),
		"authorization_nonce": task8FactoryDigest(task8Digest("factory-membership-authorization-nonce")), "request_nonce": task8FactoryDigest(task8Digest("factory-membership-request-nonce")),
		"issued_at": task8FactoryTime(windows.MembershipIssuedAt), "expires_at": task8FactoryTime(windows.MembershipExpiresAt),
		"request_digest": task8FactoryDigest(task8Digest("factory-membership-request")), "membership_retirement_tombstone_id": task8UUID("factory-membership-tombstone").String(),
		"phase": "inventory_membership_retired", "retired_at": task8FactoryTime(now.Add(-30 * time.Second)),
	})
	membership := task8FactorySign("environment-inventory-membership-retirement.v1", "release_deployment_operator", membershipBody)
	set.EnvironmentInventoryMembershipRetirementDigest = membership.Digest

	membersJCS := task8CanonicalJSON([]map[string]any{task8FactoryRetiredMemberWire(member)})
	retirementMemberSetDigest := task8FactoryDomainDigest("provider-protocol-retirement-member-set.v1", membersJCS)
	adminBody := task8CanonicalJSON(map[string]any{
		"authorization_id": task8UUID("factory-provider-retirement-authorization").String(), "retirement_id": retirementID.String(),
		"installation_id": set.InstallationID.String(), "installation_kind": "disposable_fixture", "migration_latch_digest": task8FactoryDigest(facts.MigrationLatchDigest),
		"database_identity_digest": task8FactoryDigest(facts.MigrationLatch.DatabaseIdentityDigest), "release_scope": set.ReleaseScope,
		"environment_inventory_digest": task8FactoryDigest(set.EnvironmentInventoryDigest), "environment_inventory_anchor_set_digest": task8FactoryDigest(set.EnvironmentInventoryAnchorSetDigest),
		"environment_inventory_membership_retirement_digest": task8FactoryDigest(set.EnvironmentInventoryMembershipRetirementDigest),
		"provider_identity_digest":                           task8FactoryDigest(retirement.ProviderIdentityDigest), "provider_endpoint_identity_digest": task8FactoryDigest(retirement.ProviderEndpointIdentityDigest),
		"retirement_member_set_digest": task8FactoryDigest(retirementMemberSetDigest), "member_count": "1", "authorization_scope": "permanent_disposable_namespace_retirement",
		"authorization_nonce": task8FactoryDigest(task8Digest("factory-provider-retirement-authorization-nonce")),
		"issued_at":           task8FactoryTime(windows.AdminIssuedAt), "expires_at": task8FactoryTime(windows.AdminExpiresAt),
	})
	admin := task8FactorySign("provider-protocol-downgrade-retirement-authorization.v1", "provider_downgrade_retirement_admin", adminBody)

	historyFields := map[string]any{
		"provider_identity_digest": task8FactoryDigest(retirement.ProviderIdentityDigest), "provider_endpoint_identity_digest": task8FactoryDigest(retirement.ProviderEndpointIdentityDigest),
		"namespace": member.ProviderNamespace, "observed_profiles": []string{"claim_v1", "legacy_v6"},
		"observed_at": task8FactoryTime(windows.HistoryObservedAt), "expires_at": task8FactoryTime(windows.HistoryExpiresAt),
	}
	for _, field := range []string{
		"unknown_profile_count", "unknown_record_count", "legacy_reservation_count", "legacy_terminal_count", "legacy_epoch_transition_count", "legacy_cutover_count",
		"claim_v1_credential_policy_count", "claim_v1_accepted_key_mutation_count", "incarnation_registration_count", "genesis_preparation_count",
		"genesis_completion_count", "genesis_release_preparation_count", "genesis_open_count", "claim_v1_reservation_count", "claim_v1_terminal_count",
		"claim_v1_epoch_transition_count", "claim_v1_epoch_recovery_count", "serving_lease_event_count", "runtime_rebind_count", "timeline_lineage_event_count",
		"staging_exclusion_count", "staging_recovery_count", "source_retirement_count", "genesis_authorizing_control_count", "other_mutation_count",
		"provider_history_high_water",
	} {
		historyFields[field] = "0"
	}
	history := task8FactorySign("provider-protocol-history-zero-projection.v1", "claim_v1_provider_history_auditor", task8CanonicalJSON(historyFields))

	retirementBody := task8CanonicalJSON(map[string]any{
		"retirement_id": retirementID.String(), "retirement_nonce": task8FactoryDigest(task8Digest("factory-provider-retirement-nonce")),
		"provider_downgrade_retirement_authorization_digest": task8FactoryDigest(admin.Digest),
		"environment_inventory_membership_retirement_digest": task8FactoryDigest(set.EnvironmentInventoryMembershipRetirementDigest),
		"installation_id": set.InstallationID.String(), "release_scope": set.ReleaseScope,
		"environment_inventory_digest": task8FactoryDigest(set.EnvironmentInventoryDigest), "environment_inventory_anchor_set_digest": task8FactoryDigest(set.EnvironmentInventoryAnchorSetDigest),
		"provider_identity_digest": task8FactoryDigest(retirement.ProviderIdentityDigest), "provider_endpoint_identity_digest": task8FactoryDigest(retirement.ProviderEndpointIdentityDigest),
		"member_count": "1", "members": json.RawMessage(membersJCS), "retirement_member_set_digest": task8FactoryDigest(retirementMemberSetDigest),
		"request_nonce": task8FactoryDigest(task8Digest("factory-provider-retirement-request-nonce")), "request_digest": task8FactoryDigest(task8Digest("factory-provider-retirement-request")),
		"provider_control_sequence": "1",
		"namespace_states": []map[string]any{{
			"namespace": member.ProviderNamespace, "environment_record_digests": []string{task8FactoryDigest(member.EnvironmentRecordDigest)},
			"database_identity_digests": []string{task8FactoryDigest(member.DatabaseIdentityDigest)}, "history_zero_projection_digest": task8FactoryDigest(history.Digest),
			"pre_retirement_control_sequence": "0", "provider_history_high_water": "0", "retirement_tombstone_id": tombstoneID.String(),
		}},
		"phase": "down_retired", "retired_at": task8FactoryTime(now.Add(-30 * time.Second)),
	})
	providerRetirement := task8FactorySign("provider-protocol-downgrade-retirement.v1", "claim_v1_provider", retirementBody)
	retirement.ProviderProtocolDowngradeRetirementDigest = providerRetirement.Digest

	setBody := task8FactoryRetirementSetBody(*set)
	facts.ProviderRetirementSetDigest = task8FactoryDomainDigest("provider-protocol-downgrade-retirement-set.v1", setBody)
	bundle := task8FactoryBundle(
		"provider-protocol-downgrade-retirement-set.v1", facts.ProviderRetirementSetDigest,
		task8FactoryExternal(membership), task8FactoryExternal(admin), task8FactoryExternal(history), task8FactoryExternal(providerRetirement),
	)
	return task8DownFactoryFixture{Facts: facts, Inputs: [2][]byte{setBody, bundle}}
}

func task8BuildAuthorizationFactoryFixture(now time.Time) task8AuthorizationFactoryFixture {
	facts := task8AuthorizationFacts(now)
	body := task8CanonicalAuthorizationBody(facts.Authorization)
	signed := task8FactorySign("authority-protocol-downgrade-authorization.v1", "authority_protocol_downgrade_authorizer", body)
	facts.AuthorizationDigest = signed.Digest
	facts.AuthorizationEnvelopeJCS = append([]byte(nil), signed.Envelope...)
	facts.SignerRole = "authority_protocol_downgrade_authorizer"
	facts.SignerKeyID = "56475aa75463474c0285df5dbf2bcab73da651358839e9b77481b2eab107708c"
	facts.SignaturePolicyVersion = 1
	root, err := hex.DecodeString(strings.Repeat("a1", 32))
	if err != nil {
		panic(err)
	}
	copy(facts.TrustRootDigest[:], root)
	facts.SignatureAlgorithm = "ed25519"
	return task8AuthorizationFactoryFixture{Facts: facts, Body: body}
}

type task8StagingStableProviderTuple struct {
	ProviderIdentityDigest        contracts.Digest
	ProviderEndpointDigest        contracts.Digest
	Namespace                     string
	IncarnationID                 uuid.UUID
	IncarnationPublicKeyDigest    contracts.Digest
	RuntimeInstanceID             uuid.UUID
	RuntimeInstanceGeneration     uint64
	GenesisCredentialPolicyDigest contracts.Digest
	CurrentCredentialPolicyDigest contracts.Digest
	CredentialPolicyChainDigest   contracts.Digest
	GenesisDatabaseIdentityDigest contracts.Digest
	GenesisRegistrationDigest     contracts.Digest
}

func task8FactoryStagingProofBody(
	label string,
	facts contracts.FreshRestoreImportProjectionInputV1,
	stable task8StagingStableProviderTuple,
	headDigest contracts.Digest,
	issuedAt, expiresAt time.Time,
) []byte {
	return task8CanonicalJSON(map[string]any{
		"proof_id": task8UUID(label + "-proof-id").String(), "request_digest": task8FactoryDigest(task8Digest(label + "-request")),
		"caller_challenge_nonce": task8FactoryDigest(task8Digest(label + "-caller-challenge")), "purpose": "staging_import",
		"provider_challenge_nonce": task8FactoryDigest(task8Digest(label + "-provider-challenge")), "provider_head_digest": task8FactoryDigest(headDigest),
		"provider_identity_digest": task8FactoryDigest(stable.ProviderIdentityDigest), "provider_endpoint_identity_digest": task8FactoryDigest(stable.ProviderEndpointDigest),
		"namespace": stable.Namespace, "current_database_identity_digest": task8FactoryDigest(facts.CurrentDatabaseIdentityDigest),
		"database_timeline_lineage_chain_digest":           task8FactoryDigest(facts.DatabaseTimelineLineageChainDigest),
		"current_database_incarnation_registration_digest": task8FactoryDigest(facts.CurrentDatabaseIncarnationRegistrationDigest),
		"runtime_rebind_chain_digest":                      task8FactoryDigest(facts.RuntimeRebindChainDigest), "incarnation_id": stable.IncarnationID.String(),
		"runtime_instance_binding_digest": task8FactoryDigest(facts.RuntimeInstanceBindingDigest), "runtime_instance_id": stable.RuntimeInstanceID.String(),
		"runtime_instance_generation":   strconv.FormatUint(stable.RuntimeInstanceGeneration, 10),
		"attestor_runtime_lease_digest": task8FactoryDigest(task8Digest(label + "-attestor-lease")), "attestor_runtime_lease_sequence": "1",
		"attestor_runtime_lease_expires_at":                      task8FactoryTime(expiresAt.Add(5 * time.Second)),
		"attestor_runtime_lease_descendant_chain_digest_or_null": nil, "serving_lease_id_or_null": nil,
		"serving_lease_generation_or_null": nil, "serving_lease_digest_or_null": nil, "serving_lease_expires_at_or_null": nil,
		"issued_at": task8FactoryTime(issuedAt), "expires_at": task8FactoryTime(expiresAt),
	})
}

func task8FactoryStagingHeadBody(
	facts contracts.FreshRestoreImportProjectionInputV1,
	stable task8StagingStableProviderTuple,
) []byte {
	return task8CanonicalJSON(map[string]any{
		"provider_identity_digest":          task8FactoryDigest(stable.ProviderIdentityDigest),
		"provider_endpoint_identity_digest": task8FactoryDigest(stable.ProviderEndpointDigest), "namespace": stable.Namespace,
		"protocol_profile": "claim_v1", "genesis_credential_policy_digest": task8FactoryDigest(stable.GenesisCredentialPolicyDigest),
		"current_credential_policy_digest": task8FactoryDigest(stable.CurrentCredentialPolicyDigest),
		"credential_policy_chain_digest":   task8FactoryDigest(stable.CredentialPolicyChainDigest),
		"activation_id":                    facts.ManifestTopology.TargetActivationID.String(), "mode_or_null": nil, "preparation_digest_or_null": nil,
		"activation_digest_or_null": nil, "provider_completion_digest_or_null": nil, "database_completion_digest_or_null": nil,
		"provider_release_preparation_digest_or_null": nil, "database_release_digest_or_null": nil, "open_digest_or_null": nil,
		"phase": "fresh_v7_staging_closed", "genesis_database_identity_digest": task8FactoryDigest(stable.GenesisDatabaseIdentityDigest),
		"current_database_identity_digest":                            task8FactoryDigest(facts.CurrentDatabaseIdentityDigest),
		"database_timeline_lineage_chain_digest":                      task8FactoryDigest(facts.DatabaseTimelineLineageChainDigest),
		"latest_database_timeline_lineage_attestation_digest_or_null": nil,
		"genesis_database_incarnation_registration_digest":            task8FactoryDigest(stable.GenesisRegistrationDigest),
		"current_database_incarnation_registration_digest":            task8FactoryDigest(facts.CurrentDatabaseIncarnationRegistrationDigest),
		"runtime_rebind_chain_digest":                                 task8FactoryDigest(facts.RuntimeRebindChainDigest), "incarnation_id": stable.IncarnationID.String(),
		"incarnation_public_key_digest":   task8FactoryDigest(stable.IncarnationPublicKeyDigest),
		"runtime_instance_binding_digest": task8FactoryDigest(facts.RuntimeInstanceBindingDigest), "runtime_instance_id": stable.RuntimeInstanceID.String(),
		"runtime_instance_generation": strconv.FormatUint(stable.RuntimeInstanceGeneration, 10),
		"serving_lease_id_or_null":    nil, "serving_lease_generation_or_null": nil, "serving_lease_digest_or_null": nil,
		"database_release_attestation_digest_or_null": nil, "epoch_evidence_digest_or_null": nil, "genesis_epoch_transition_root_digest_or_null": nil,
		"selected_genesis_epoch_or_null": nil, "current_epoch_or_null": nil, "epoch_transition_chain_digest_or_null": nil,
		"epoch_transition_terminal_chain_digest_or_null": nil, "pending_epoch_transition_digest_or_null": nil,
		"latest_epoch_transition_resolution_digest_or_null": nil, "epoch_transition_recovery_state_or_null": nil,
		"epoch_transition_recovery_request_digest_or_null":  nil,
		"staging_exclusion_id_or_null":                      facts.StagingExclusionLease.ExclusionID.String(),
		"staging_exclusion_acquire_request_digest_or_null":  task8FactoryDigest(facts.StagingExclusionLease.RequestDigest),
		"staging_exclusion_recovery_request_digest_or_null": nil, "staging_exclusion_state_or_null": "held",
		"latest_reserved_sequence": "0", "latest_committed_sequence": "0", "latest_reservation_digest_or_null": nil,
		"latest_committed_operation_id_or_null": nil, "latest_committed_receipt_digest_or_null": nil,
		"latest_committed_database_point_or_null": nil,
		"provider_control_sequence":               strconv.FormatUint(facts.StagingExclusionLease.ProviderControlSequence, 10), "provider_history_high_water": "0",
	})
}

func task8FactoryStagingCapabilityBody(value contracts.FreshRestoreStagingImportCapabilityFactsV1) []byte {
	return task8CanonicalJSON(map[string]any{
		"capability_id": value.CapabilityID.String(), "single_use_apply_id": value.SingleUseApplyID.String(), "manifest_digest": task8FactoryDigest(value.ManifestDigest),
		"target_activation_id": value.TargetActivationID.String(), "target_database_identity_digest": task8FactoryDigest(value.TargetDatabaseIdentityDigest),
		"database_timeline_lineage_chain_digest":          task8FactoryDigest(value.DatabaseTimelineLineageChainDigest),
		"target_database_incarnation_registration_digest": task8FactoryDigest(value.TargetDatabaseIncarnationRegistrationDigest),
		"runtime_rebind_chain_digest":                     task8FactoryDigest(value.RuntimeRebindChainDigest), "runtime_instance_binding_digest": task8FactoryDigest(value.RuntimeInstanceBindingDigest),
		"planned_staging_exclusion_id":                  value.PlannedStagingExclusionID.String(),
		"expected_pre_acquire_provider_head_digest":     task8FactoryDigest(value.ExpectedPreAcquireProviderHeadDigest),
		"pre_acquire_database_incarnation_proof_digest": task8FactoryDigest(value.PreAcquireDatabaseIncarnationProofDigest),
		"provider_phase":                                string(value.ProviderPhase), "provider_serving_lease_absent_digest": task8FactoryDigest(value.ProviderServingLeaseAbsentDigest),
		"database_route_closed_digest": task8FactoryDigest(value.DatabaseRouteClosedDigest), "pre_import_inventory_digest": task8FactoryDigest(value.PreImportInventoryDigest),
		"allowed_object_set_digest":             task8FactoryDigest(value.AllowedObjectSetDigest),
		"expected_post_import_inventory_digest": task8FactoryDigest(value.ExpectedPostImportInventoryDigest),
		"transaction_nonce":                     task8FactoryDigest(value.TransactionNonce), "issued_at": task8FactoryTime(value.IssuedAt), "expires_at": task8FactoryTime(value.ExpiresAt),
	})
}

func task8FactoryStagingLeaseBody(value contracts.FreshV7StagingExclusionLeaseFactsV1) []byte {
	return task8CanonicalJSON(map[string]any{
		"exclusion_id": value.ExclusionID.String(), "staging_import_capability_digest": task8FactoryDigest(value.StagingImportCapabilityDigest),
		"capability_registration_commit_challenge_digest":   task8FactoryDigest(value.CapabilityRegistrationCommitChallengeDigest),
		"capability_registration_commit_attestation_digest": task8FactoryDigest(value.CapabilityRegistrationCommitAttestationDigest),
		"target_activation_id":                              value.TargetActivationID.String(), "target_database_identity_digest": task8FactoryDigest(value.TargetDatabaseIdentityDigest),
		"database_timeline_lineage_chain_digest":          task8FactoryDigest(value.DatabaseTimelineLineageChainDigest),
		"expected_provider_head_digest":                   task8FactoryDigest(value.ExpectedProviderHeadDigest),
		"pre_acquire_database_incarnation_proof_digest":   task8FactoryDigest(value.PreAcquireDatabaseIncarnationProofDigest),
		"target_database_incarnation_registration_digest": task8FactoryDigest(value.TargetDatabaseIncarnationRegistrationDigest),
		"runtime_rebind_chain_digest":                     task8FactoryDigest(value.RuntimeRebindChainDigest), "runtime_instance_binding_digest": task8FactoryDigest(value.RuntimeInstanceBindingDigest),
		"requested_admission_expires_at": task8FactoryTime(value.RequestedAdmissionExpiresAt), "request_nonce": task8FactoryDigest(value.RequestNonce),
		"request_digest": task8FactoryDigest(value.RequestDigest), "exclusion_lease_id": value.ExclusionLeaseID.String(),
		"acquisition_locked_provider_head_digest": task8FactoryDigest(value.AcquisitionLockedProviderHeadDigest),
		"provider_serving_lease_absent_digest":    task8FactoryDigest(value.ProviderServingLeaseAbsentDigest),
		"provider_control_sequence":               strconv.FormatUint(value.ProviderControlSequence, 10), "exclusion_state": string(value.ExclusionState),
		"issued_at": task8FactoryTime(value.IssuedAt), "admission_expires_at": task8FactoryTime(value.AdmissionExpiresAt),
	})
}

func task8BuildStagingFactoryFixture(now time.Time) task8StagingFactoryFixture {
	facts := task8AdmissionFacts(now)
	facts.ManifestTopology.IssuedAt = now.Add(-time.Minute)
	facts.ManifestTopology.ExpiresAt = now.Add(10 * time.Minute)
	facts.StagingImportCapability.IssuedAt = now.Add(-10 * time.Second)
	facts.StagingImportCapability.ExpiresAt = now.Add(4 * time.Minute)
	facts.StagingExclusionLease.IssuedAt = now.Add(-5 * time.Second)
	facts.StagingExclusionLease.RequestedAdmissionExpiresAt = now.Add(3 * time.Minute)
	facts.StagingExclusionLease.AdmissionExpiresAt = now.Add(3 * time.Minute)
	facts.AdmissionCheckedAt = now

	manifestObject := facts.ManifestTopology.Objects[0]
	allowedSetBody := task8CanonicalJSON([]map[string]string{{
		"object_type": string(manifestObject.ObjectType), "canonical_key": manifestObject.CanonicalKey,
		"payload_digest": task8FactoryDigest(manifestObject.PayloadDigest),
	}})
	facts.StagingImportCapability.AllowedObjectSetDigest = task8FactoryDomainDigest("allowed-import-object-set.v1", allowedSetBody)
	facts.ManifestTopology.CompleteNodeSetDigest = task8FactoryDomainDigest("complete-node-set.v1", task8CanonicalJSON([]string{}))

	normalizedPayload := task8CanonicalJSON(map[string]string{
		"pop_code": "tbs-1", "iso_country": "GE", "region": "tbilisi", "operator_state": "disabled", "version": "1",
	})
	postProjectionBody := task8CanonicalJSON(map[string]any{
		"projection_version": "1", "target_activation_id": facts.ManifestTopology.TargetActivationID.String(),
		"target_deployment_id":            facts.ManifestTopology.TargetDeploymentID.String(),
		"target_database_identity_digest": task8FactoryDigest(facts.CurrentDatabaseIdentityDigest),
		"normalized_catalog_digest":       task8FactoryDigest(facts.NormalizedCatalogDigest), "object_count": "1",
		"objects": []map[string]any{{"object_type": string(manifestObject.ObjectType), "canonical_key": manifestObject.CanonicalKey, "normalized_payload": json.RawMessage(normalizedPayload)}},
	})
	postProjectionDigest := task8FactoryDomainDigest("fresh-import-topology-projection.v1", postProjectionBody)
	facts.ManifestTopology.ExpectedPostImportInventoryDigest = postProjectionDigest
	facts.StagingImportCapability.ExpectedPostImportInventoryDigest = postProjectionDigest
	preProjectionBody := task8CanonicalJSON(map[string]any{
		"projection_version": "1", "target_activation_id": facts.ManifestTopology.TargetActivationID.String(),
		"target_deployment_id":            facts.ManifestTopology.TargetDeploymentID.String(),
		"target_database_identity_digest": task8FactoryDigest(facts.CurrentDatabaseIdentityDigest),
		"normalized_catalog_digest":       task8FactoryDigest(facts.NormalizedCatalogDigest), "object_count": "0", "objects": []map[string]any{},
	})
	preProjectionDigest := task8FactoryDomainDigest("fresh-import-topology-projection.v1", preProjectionBody)
	facts.PreImportInventoryDigest = preProjectionDigest
	facts.StagingImportCapability.PreImportInventoryDigest = preProjectionDigest

	manifestBody := task8CanonicalJSON(map[string]any{
		"manifest_id": facts.ManifestTopology.ManifestID.String(), "single_use_apply_id": facts.ManifestTopology.SingleUseApplyID.String(),
		"fresh_restore_requirement_digest":            task8FactoryDigest(task8Digest("factory-source-requirement")),
		"source_membership_retirement_digest":         task8FactoryDigest(task8Digest("factory-source-membership-retirement")),
		"source_legacy_retirement_set_digest":         task8FactoryDigest(task8Digest("factory-source-legacy-retirement")),
		"source_post_seal_exact_cover_digest":         task8FactoryDigest(task8Digest("factory-source-post-seal-cover")),
		"source_archive_evidence_digest":              task8FactoryDigest(task8Digest("factory-source-archive")),
		"source_legacy_epoch_source_set_digest":       task8FactoryDigest(task8Digest("factory-source-legacy-epoch-set")),
		"source_legacy_epoch_maximum_evidence_digest": task8FactoryDigest(task8Digest("factory-source-legacy-epoch-maximum")),
		"source_snapshot_id":                          task8UUID("factory-source-snapshot").String(), "source_archive_point": "task8/factory/archive",
		"target_activation_id": facts.ManifestTopology.TargetActivationID.String(), "target_deployment_id": facts.ManifestTopology.TargetDeploymentID.String(),
		"target_database_incarnation_registration_digest": task8FactoryDigest(facts.ManifestTopology.TargetDatabaseIncarnationRegistrationDigest),
		"target_epoch_evidence_digest":                    task8FactoryDigest(facts.ManifestTopology.TargetEpochEvidenceDigest),
		"issued_at":                                       task8FactoryTime(facts.ManifestTopology.IssuedAt), "expires_at": task8FactoryTime(facts.ManifestTopology.ExpiresAt),
		"object_count": "1", "objects": []map[string]any{{
			"object_type": string(manifestObject.ObjectType), "canonical_key": manifestObject.CanonicalKey,
			"payload": json.RawMessage(manifestObject.Payload), "payload_digest": task8FactoryDigest(manifestObject.PayloadDigest),
		}},
		"complete_node_set_digest":              task8FactoryDigest(facts.ManifestTopology.CompleteNodeSetDigest),
		"forbidden_object_class_set_digest":     task8FactoryDigest(facts.ManifestTopology.ForbiddenObjectClassSetDigest),
		"expected_post_import_inventory_digest": task8FactoryDigest(facts.ManifestTopology.ExpectedPostImportInventoryDigest),
	})
	manifest := task8FactorySign("fresh-restore-import-manifest.v1", "fresh_restore_export_operator", manifestBody)
	facts.ManifestTopology.ManifestDigest = manifest.Digest
	facts.StagingImportCapability.ManifestDigest = manifest.Digest

	stable := task8StagingStableProviderTuple{
		ProviderIdentityDigest: task8Digest("factory-provider-identity"), ProviderEndpointDigest: task8Digest("factory-provider-endpoint"),
		Namespace: "task8/fixture", IncarnationID: task8UUID("factory-incarnation"),
		IncarnationPublicKeyDigest: task8Digest("factory-incarnation-public-key"), RuntimeInstanceID: task8UUID("factory-runtime-instance"),
		RuntimeInstanceGeneration: 1, GenesisCredentialPolicyDigest: task8Digest("factory-genesis-credential-policy"),
		CurrentCredentialPolicyDigest: task8Digest("factory-current-credential-policy"), CredentialPolicyChainDigest: task8Digest("factory-credential-policy-chain"),
		GenesisDatabaseIdentityDigest: task8Digest("factory-genesis-database-identity"), GenesisRegistrationDigest: task8Digest("factory-genesis-registration"),
	}
	preAcquireHeadDigest := task8Digest("factory-pre-acquire-provider-head")
	facts.StagingImportCapability.ExpectedPreAcquireProviderHeadDigest = preAcquireHeadDigest
	facts.StagingExclusionLease.ExpectedProviderHeadDigest = preAcquireHeadDigest
	preAcquireProof := task8FactorySign(
		"database-incarnation-proof.v1", "claim_v1_provider",
		task8FactoryStagingProofBody("factory-pre-acquire", facts, stable, preAcquireHeadDigest, now.Add(-15*time.Second), now.Add(10*time.Second)),
	)
	facts.StagingImportCapability.PreAcquireDatabaseIncarnationProofDigest = preAcquireProof.Digest
	facts.StagingExclusionLease.PreAcquireDatabaseIncarnationProofDigest = preAcquireProof.Digest

	capability := task8FactorySign(
		"staging-import-capability.v1", "fresh_restore_staging_import_authorizer",
		task8FactoryStagingCapabilityBody(facts.StagingImportCapability),
	)
	facts.StagingImportCapability.CapabilityDigest = capability.Digest
	facts.StagingExclusionLease.StagingImportCapabilityDigest = capability.Digest

	head := task8FactorySign("claim-v1-provider-head.v1", "claim_v1_provider", task8FactoryStagingHeadBody(facts, stable))
	facts.CurrentProviderHeadDigest = head.Digest
	facts.StagingExclusionLease.AcquisitionLockedProviderHeadDigest = head.Digest
	currentProof := task8FactorySign(
		"database-incarnation-proof.v1", "claim_v1_provider",
		task8FactoryStagingProofBody("factory-current", facts, stable, head.Digest, now.Add(-time.Second), now.Add(20*time.Second)),
	)
	lease := task8FactorySign("fresh-v7-staging-exclusion-lease.v1", "claim_v1_provider", task8FactoryStagingLeaseBody(facts.StagingExclusionLease))
	facts.StagingExclusionLease.LeaseDigest = lease.Digest

	capabilityBundle := task8FactoryBundle(
		capability.Schema, capability.Digest, task8FactoryExternal(preAcquireProof), task8FactoryExternal(manifest),
	)
	manifestBundle := task8FactoryBundle(manifest.Schema, manifest.Digest, task8FactoryEvidenceMember{
		Kind: "database_immutable_body", Schema: "fresh-import-topology-projection.v1", Digest: postProjectionDigest, BodyJCS: postProjectionBody,
	})
	leaseBundle := task8FactoryBundle(lease.Schema, lease.Digest, task8FactoryExternal(capability), task8FactoryExternal(head))
	currentProofBundle := task8FactoryBundle(currentProof.Schema, currentProof.Digest, task8FactoryExternal(head))

	return task8StagingFactoryFixture{Facts: facts, Inputs: [14][]byte{
		capability.BodyJCS, capability.Envelope, capabilityBundle,
		manifest.BodyJCS, manifest.Envelope, manifestBundle,
		lease.BodyJCS, lease.Envelope, leaseBundle,
		head.BodyJCS, head.Envelope,
		currentProof.BodyJCS, currentProof.Envelope, currentProofBundle,
	}}
}

func task8RunIntegrationFactoryProbe(t *testing.T) {
	t.Helper()
	repositoryRoot, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	parentOverlayPath := ""
	for _, token := range strings.Fields(os.Getenv("GOFLAGS")) {
		if strings.HasPrefix(token, "-overlay=") {
			parentOverlayPath = strings.TrimPrefix(token, "-overlay=")
		}
	}
	if parentOverlayPath == "" {
		t.Fatal("integration factory probe requires the approved caller overlay in GOFLAGS")
	}
	parentOverlayRaw, err := os.ReadFile(parentOverlayPath)
	if err != nil {
		t.Fatalf("read effective caller overlay: %v", err)
	}
	merged, err := task8AuthenticateApprovedFactoryProbeOverlay(repositoryRoot, parentOverlayRaw)
	if err != nil {
		t.Fatalf("authenticate effective caller overlay: %v", err)
	}
	probeParent := filepath.Join(repositoryRoot, ".superpowers", "sdd", ".t8")
	if err := os.MkdirAll(probeParent, 0o700); err != nil {
		t.Fatal(err)
	}
	probeDirectory, err := os.MkdirTemp(probeParent, "factory-probe-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(probeDirectory) })
	probeSourcePath := filepath.Join(probeDirectory, "v7_factory_probe_integration_test.go")
	probe := `package authority

import (
	"bytes"
	"errors"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

func task8CallStagingFactory(fixture task8StagingFactoryFixture) (VerifiedFreshRestoreImportAdmission, error) {
	i := fixture.Inputs
	return NewFreshRestoreImportAdmissionForIntegration(
		fixture.Facts, i[0], i[1], i[2], i[3], i[4], i[5], i[6], i[7], i[8], i[9], i[10], i[11], i[12], i[13],
	)
}

func task8CallWithConcurrentCallerMutation(t *testing.T, kind string, mutate func(), call func() error) error {
	t.Helper()
	previousHook := authorityV7IntegrationFactoryAfterSnapshot
	snapshotDone := make(chan struct{})
	firstMutationDone := make(chan struct{})
	stop := make(chan struct{})
	var mutator sync.WaitGroup
	authorityV7IntegrationFactoryAfterSnapshot = func(got string) {
		if got != kind { t.Fatalf("factory snapshot hook = %q, want %q", got, kind) }
		close(snapshotDone)
		<-firstMutationDone
	}
	defer func() { authorityV7IntegrationFactoryAfterSnapshot = previousHook }()
	mutator.Add(1)
	go func() {
		defer mutator.Done()
		<-snapshotDone
		mutate()
		close(firstMutationDone)
		for {
			select {
			case <-stop:
				return
			default:
				mutate()
				runtime.Gosched()
			}
		}
	}()
	err := call()
	close(stop)
	mutator.Wait()
	return err
}

func TestNodeControlV7IntegrationFactoryMintMutationCloneChild(t *testing.T) {
	t.Run("down", func(t *testing.T) {
		now := time.Now().UTC()
		fixture := task8BuildDownFactoryFixture(now)
		want := task8BuildDownFactoryFixture(now)
		var grant VerifiedAuthorityV7DownGrant
		err := task8CallWithConcurrentCallerMutation(t, "down", func() {
			if fixture.Facts.ProviderRetirementSet.RetiredMembers[0].DatabaseName == "caller-mutated-a" {
				fixture.Facts.ProviderRetirementSet.RetiredMembers[0].DatabaseName = "caller-mutated-b"
			} else {
				fixture.Facts.ProviderRetirementSet.RetiredMembers[0].DatabaseName = "caller-mutated-a"
			}
			for index := range fixture.Inputs { fixture.Inputs[index][0] ^= 0xff }
		}, func() error {
			var callErr error
			grant, callErr = NewDisposableAuthorityV7DownGrantForIntegration(fixture.Facts, fixture.Inputs[0], fixture.Inputs[1])
			return callErr
		})
		if err != nil { t.Fatalf("positive mint: %v", err) }
		got, err := ConsumeVerifiedAuthorityV7DownGrant(grant)
		if err != nil || !reflect.DeepEqual(got.Facts, want.Facts) { t.Fatalf("clone consume: err=%v facts_equal=%v", err, reflect.DeepEqual(got.Facts, want.Facts)) }
		for index := range want.Inputs { if !bytes.Equal([][]byte{got.ProviderRetirementSetBodyJCS, got.ProviderRetirementEvidenceBundleJCS}[index], want.Inputs[index]) { t.Fatalf("clone input %d changed", index) } }
		got.Facts.ProviderRetirementSet.RetiredMembers[0].DatabaseName = "returned-mutated"
		got.ProviderRetirementSetBodyJCS[0] ^= 0xff
		if grant.state.view.Facts.ProviderRetirementSet.RetiredMembers[0].DatabaseName != want.Facts.ProviderRetirementSet.RetiredMembers[0].DatabaseName || !bytes.Equal(grant.state.view.ProviderRetirementSetBodyJCS, want.Inputs[0]) { t.Fatal("returned Down view aliases sealed state") }

		windowNow := time.Now().UTC()
		futureEvidence := task8BuildDownFactoryFixtureWithWindows(windowNow, task8DownEvidenceWindows{
			MembershipIssuedAt: windowNow.Add(time.Minute), MembershipExpiresAt: windowNow.Add(2*time.Minute),
			AdminIssuedAt: windowNow.Add(-time.Minute), AdminExpiresAt: windowNow.Add(time.Minute),
			HistoryObservedAt: windowNow.Add(-time.Minute), HistoryExpiresAt: windowNow.Add(time.Minute),
		})
		if _, err := NewDisposableAuthorityV7DownGrantForIntegration(futureEvidence.Facts, futureEvidence.Inputs[0], futureEvidence.Inputs[1]); !errors.Is(err, ErrInvalidArgument) { t.Fatalf("future-issued supporting evidence = %v, want ErrInvalidArgument", err) }
		nonOverlappingEvidence := task8BuildDownFactoryFixtureWithWindows(windowNow, task8DownEvidenceWindows{
			MembershipIssuedAt: windowNow.Add(-time.Minute), MembershipExpiresAt: windowNow.Add(time.Minute),
			AdminIssuedAt: windowNow.Add(2*time.Minute), AdminExpiresAt: windowNow.Add(3*time.Minute),
			HistoryObservedAt: windowNow.Add(2*time.Minute), HistoryExpiresAt: windowNow.Add(3*time.Minute),
		})
		if _, err := NewDisposableAuthorityV7DownGrantForIntegration(nonOverlappingEvidence.Facts, nonOverlappingEvidence.Inputs[0], nonOverlappingEvidence.Inputs[1]); !errors.Is(err, ErrInvalidArgument) { t.Fatalf("non-overlapping supporting evidence = %v, want ErrInvalidArgument", err) }
		grantOutlivesEvidence := task8BuildDownFactoryFixture(windowNow)
		grantOutlivesEvidence.Facts.ExpiresAt = grantOutlivesEvidence.Facts.ExpiresAt.Add(time.Nanosecond)
		if _, err := NewDisposableAuthorityV7DownGrantForIntegration(grantOutlivesEvidence.Facts, grantOutlivesEvidence.Inputs[0], grantOutlivesEvidence.Inputs[1]); !errors.Is(err, ErrInvalidArgument) { t.Fatalf("grant outlives supporting evidence = %v, want ErrInvalidArgument", err) }

		candidate := task8BuildDownFactoryFixture(now)
		candidate.Facts.ProviderRetirementSetDigest[0] ^= 1
		if _, err := NewDisposableAuthorityV7DownGrantForIntegration(candidate.Facts, candidate.Inputs[0], candidate.Inputs[1]); !errors.Is(err, ErrInvalidArgument) { t.Fatalf("facts mutation = %v", err) }
		for input := range candidate.Inputs {
			candidate = task8BuildDownFactoryFixture(now)
			candidate.Inputs[input] = task8CanonicalInvalidDownFactoryInput(input, candidate.Inputs[input])
			if _, err := NewDisposableAuthorityV7DownGrantForIntegration(candidate.Facts, candidate.Inputs[0], candidate.Inputs[1]); !errors.Is(err, ErrInvalidArgument) { t.Fatalf("input %d mutation = %v", input, err) }
		}
		for _, mutation := range task8FactoryEnvelopeMutations() {
			candidate = task8BuildDownFactoryFixture(time.Now().UTC())
			candidate.Inputs[1] = task8MutateFactoryBundleEnvelopeField(candidate.Inputs[1], mutation.Field, mutation.Value)
			if _, err := NewDisposableAuthorityV7DownGrantForIntegration(candidate.Facts, candidate.Inputs[0], candidate.Inputs[1]); !errors.Is(err, ErrInvalidArgument) { t.Fatalf("envelope %s mutation = %v", mutation.Name, err) }
		}
		for name, mutate := range map[string]func([]byte) []byte{
			"message": func(raw []byte) []byte { return task8MutateFactoryBundleField(raw, "message_schema", "wrong-schema.v1") },
			"count": func(raw []byte) []byte { return task8MutateFactoryBundleField(raw, "evidence_count", "5") },
			"member": func(raw []byte) []byte { return task8MutateFactoryBundleMemberField(raw, "body_digest", strings.Repeat("00", 32)) },
			"order": task8ReverseFactoryBundleEvidence,
		} {
			candidate = task8BuildDownFactoryFixture(time.Now().UTC())
			candidate.Inputs[1] = mutate(candidate.Inputs[1])
			if _, err := NewDisposableAuthorityV7DownGrantForIntegration(candidate.Facts, candidate.Inputs[0], candidate.Inputs[1]); !errors.Is(err, ErrInvalidArgument) { t.Fatalf("bundle %s mutation = %v", name, err) }
		}
	})

	t.Run("authorization", func(t *testing.T) {
		now := time.Now().UTC()
		fixture := task8BuildAuthorizationFactoryFixture(now)
		want := task8BuildAuthorizationFactoryFixture(now)
		var authorization VerifiedAuthorityV7DownAuthorization
		err := task8CallWithConcurrentCallerMutation(t, "authorization", func() {
			fixture.Facts.AuthorizationEnvelopeJCS[0] ^= 0xff
			fixture.Body[0] ^= 0xff
		}, func() error {
			var callErr error
			authorization, callErr = NewAuthorityV7DownAuthorizationForIntegration(fixture.Facts, fixture.Body)
			return callErr
		})
		if err != nil { t.Fatalf("positive mint: %v", err) }
		got, err := ConsumeVerifiedAuthorityV7DownAuthorization(authorization)
		if err != nil || !reflect.DeepEqual(got.Facts, want.Facts) || !bytes.Equal(got.AuthorizationBodyJCS, want.Body) { t.Fatalf("clone consume: err=%v", err) }
		got.AuthorizationBodyJCS[0] ^= 0xff
		if !bytes.Equal(authorization.state.view.AuthorizationBodyJCS, want.Body) { t.Fatal("returned authorization view aliases sealed state") }

		futureIssued := task8BuildAuthorizationFactoryFixture(time.Now().UTC().Add(time.Minute))
		if _, err := NewAuthorityV7DownAuthorizationForIntegration(futureIssued.Facts, futureIssued.Body); !errors.Is(err, ErrInvalidArgument) { t.Fatalf("future-issued factory mint = %v, want ErrInvalidArgument", err) }

		candidate := task8BuildAuthorizationFactoryFixture(now)
		candidate.Facts.AuthorizationDigest[0] ^= 1
		if _, err := NewAuthorityV7DownAuthorizationForIntegration(candidate.Facts, candidate.Body); !errors.Is(err, ErrInvalidArgument) { t.Fatalf("facts mutation = %v", err) }
		candidate = task8BuildAuthorizationFactoryFixture(now)
		candidate.Body = task8MutateFactoryEnvelopeField(candidate.Body, "transaction_nonce", strings.Repeat("b2", 32))
		if _, err := NewAuthorityV7DownAuthorizationForIntegration(candidate.Facts, candidate.Body); !errors.Is(err, ErrInvalidArgument) { t.Fatalf("body mutation = %v", err) }
		for _, mutation := range task8FactoryEnvelopeMutations() {
			candidate = task8BuildAuthorizationFactoryFixture(time.Now().UTC())
			candidate.Facts.AuthorizationEnvelopeJCS = task8MutateFactoryEnvelopeField(candidate.Facts.AuthorizationEnvelopeJCS, mutation.Field, mutation.Value)
			if _, err := NewAuthorityV7DownAuthorizationForIntegration(candidate.Facts, candidate.Body); !errors.Is(err, ErrInvalidArgument) { t.Fatalf("envelope %s mutation = %v", mutation.Name, err) }
		}
	})

	t.Run("staging", func(t *testing.T) {
		now := time.Now().UTC()
		fixture := task8BuildStagingFactoryFixture(now)
		want := task8BuildStagingFactoryFixture(now)
		if fixture.Facts.StagingImportCapability.ExpectedPreAcquireProviderHeadDigest == fixture.Facts.CurrentProviderHeadDigest { t.Fatal("pre/current provider heads are not independent") }
		currentProofDigest := task8FactoryDomainDigest("database-incarnation-proof.v1", fixture.Inputs[11])
		if fixture.Facts.StagingImportCapability.PreAcquireDatabaseIncarnationProofDigest == currentProofDigest { t.Fatal("pre/current proof digests are not independent") }
		if err := fixture.Facts.Validate(); err != nil { t.Fatalf("positive facts Validate: %v", err) }
		if err := validateFreshRestoreManifestPayloadDigests(fixture.Facts.ManifestTopology); err != nil { t.Fatalf("positive payload digests: %v", err) }
		if err := validateFreshRestoreAdmissionCurrentAt(fixture.Facts, time.Now()); err != nil { t.Fatalf("positive admission currentness: %v", err) }
		var admission VerifiedFreshRestoreImportAdmission
		err := task8CallWithConcurrentCallerMutation(t, "staging", func() {
			fixture.Facts.ManifestTopology.Objects[0].Payload[0] ^= 0xff
			for index := range fixture.Inputs { fixture.Inputs[index][0] ^= 0xff }
		}, func() error {
			var callErr error
			admission, callErr = task8CallStagingFactory(fixture)
			return callErr
		})
		if err != nil { t.Fatalf("positive mint: %v", err) }
		publicView, err := ViewVerifiedFreshRestoreImportAdmission(admission)
		if err != nil || !reflect.DeepEqual(publicView, want.Facts) { t.Fatalf("public clone: err=%v equal=%v", err, reflect.DeepEqual(publicView, want.Facts)) }
		privateView, err := consumeVerifiedFreshRestoreImportAdmission(admission)
		if err != nil || !reflect.DeepEqual(privateView.Facts, want.Facts) { t.Fatalf("private clone: err=%v", err) }
		gotInputs := task8AdmissionPrivateByteArms(&privateView)
		for index := range want.Inputs { if !bytes.Equal(gotInputs[index], want.Inputs[index]) { t.Fatalf("clone input %d changed", index) } }
		privateView.ManifestBodyJCS[0] ^= 0xff
		if !bytes.Equal(admission.state.view.ManifestBodyJCS, want.Inputs[3]) { t.Fatal("returned staging view aliases sealed state") }

		candidate := task8BuildStagingFactoryFixture(now)
		candidate.Facts.NormalizedCatalogDigest[0] ^= 1
		if _, err := task8CallStagingFactory(candidate); !errors.Is(err, ErrInvalidArgument) { t.Fatalf("facts mutation = %v", err) }
		for input := range candidate.Inputs {
			candidate = task8BuildStagingFactoryFixture(now)
			candidate.Inputs[input] = task8CanonicalInvalidStagingFactoryInput(input, candidate.Inputs[input])
			if _, err := task8CallStagingFactory(candidate); !errors.Is(err, ErrInvalidArgument) { t.Fatalf("input %d mutation = %v", input, err) }
		}
		for _, mutation := range task8FactoryEnvelopeMutations() {
			candidate = task8BuildStagingFactoryFixture(time.Now().UTC())
			candidate.Inputs[1] = task8MutateFactoryEnvelopeField(candidate.Inputs[1], mutation.Field, mutation.Value)
			if _, err := task8CallStagingFactory(candidate); !errors.Is(err, ErrInvalidArgument) { t.Fatalf("envelope %s mutation = %v", mutation.Name, err) }
		}
		for _, bundleIndex := range []int{2, 5, 8, 13} {
			for name, mutate := range map[string]func([]byte) []byte{
				"message": func(raw []byte) []byte { return task8MutateFactoryBundleField(raw, "message_body_digest", strings.Repeat("00", 32)) },
				"count": func(raw []byte) []byte { return task8MutateFactoryBundleField(raw, "evidence_count", "99") },
				"member": func(raw []byte) []byte { return task8MutateFactoryBundleMemberField(raw, "body_digest", strings.Repeat("00", 32)) },
			} {
				candidate = task8BuildStagingFactoryFixture(time.Now().UTC())
				candidate.Inputs[bundleIndex] = mutate(candidate.Inputs[bundleIndex])
				if _, err := task8CallStagingFactory(candidate); !errors.Is(err, ErrInvalidArgument) { t.Fatalf("bundle %d %s mutation = %v", bundleIndex, name, err) }
			}
		}
		for _, bundleIndex := range []int{2, 8} {
			candidate = task8BuildStagingFactoryFixture(time.Now().UTC())
			candidate.Inputs[bundleIndex] = task8ReverseFactoryBundleEvidence(candidate.Inputs[bundleIndex])
			if _, err := task8CallStagingFactory(candidate); !errors.Is(err, ErrInvalidArgument) { t.Fatalf("bundle %d order mutation = %v", bundleIndex, err) }
		}
	})
}
`
	if err := os.WriteFile(probeSourcePath, []byte(probe), 0o600); err != nil {
		t.Fatal(err)
	}
	probeOnDisk, err := os.ReadFile(probeSourcePath)
	if err != nil || !bytes.Equal(probeOnDisk, []byte(probe)) {
		t.Fatalf("authenticate generated factory probe: read error=%v content_equal=%v", err, bytes.Equal(probeOnDisk, []byte(probe)))
	}
	virtualSourcePath := filepath.Join(repositoryRoot, "internal", "nodecontrol", "authority", "v7_factory_probe_integration_test.go")
	merged.Replace[filepath.ToSlash(virtualSourcePath)] = filepath.ToSlash(probeSourcePath)
	mergedRaw, err := json.Marshal(merged)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := task8AuthenticateMergedFactoryProbeOverlay(repositoryRoot, probeSourcePath, mergedRaw); err != nil {
		t.Fatalf("authenticate merged factory probe overlay: %v", err)
	}
	mergedOverlayPath := filepath.Join(probeDirectory, "overlay.json")
	if err := os.WriteFile(mergedOverlayPath, mergedRaw, 0o600); err != nil {
		t.Fatal(err)
	}
	mergedOnDisk, err := os.ReadFile(mergedOverlayPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := task8AuthenticateMergedFactoryProbeOverlay(repositoryRoot, probeSourcePath, mergedOnDisk); err != nil {
		t.Fatalf("re-authenticate on-disk merged factory probe overlay: %v", err)
	}
	commandArguments := []string{"test", "-tags=integration", "-overlay=" + mergedOverlayPath}
	if os.Getenv("CGO_ENABLED") == "1" {
		commandArguments = append(commandArguments, "-race")
	}
	commandArguments = append(commandArguments,
		"./internal/nodecontrol/authority", "-run", "^TestNodeControlV7IntegrationFactoryMintMutationCloneChild$", "-count=1", "-timeout=90s",
	)
	command := task8ChildGoCommand(t, repositoryRoot, commandArguments...)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("integration factory mint/mutation/clone probe failed: %v\n%s", err, output)
	}
	t.Logf("integration factory child selector PASS: %s", bytes.TrimSpace(output))
}
