package testinfra_test

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

const c12IntegrationManifestSchema = "talenro-c12-integration-manifest/v1"

type c12IntegrationManifest struct {
	Schema string                `json:"schema"`
	Groups []c12IntegrationGroup `json:"groups"`
}

type c12IntegrationGroup struct {
	ID      string   `json:"id"`
	Package string   `json:"package"`
	Profile string   `json:"profile"`
	Tests   []string `json:"tests"`
	Timeout string   `json:"timeout"`
}

type c12IntegrationSource struct {
	Path    string
	Tracked bool
	Body    []byte
}

var (
	c12SelectedSuite        = flag.String("c12-suite", "", "selected closed C12 integration suite")
	c12SelectedManifests    = flag.String("c12-manifests", "", "pipe-delimited selected C12 integration manifests")
	c12SelectedSuiteTimeout = flag.String("c12-suite-timeout", "", "selected C12 suite global timeout")
	c12CandidateTree        = flag.String("c12-candidate-tree", "", "immutable staged candidate tree validated by the C12 runner")
	c12FocusedPackages      = flag.String("c12-focused-packages", "", "pipe-delimited focused C12 packages to resolve")
	c12FocusedTests         = flag.String("c12-focused-tests", "", "pipe-delimited focused C12 literal tests to resolve")
	c12ManifestIDPattern    = regexp.MustCompile(`^[a-z][a-z0-9-]{0,63}$`)
	c12PackagePattern       = regexp.MustCompile(`^\./(?:[A-Za-z0-9_.-]+/)*[A-Za-z0-9_.-]+$`)
	c12TestNamePattern      = regexp.MustCompile(`^Test[A-Za-z0-9_]+$`)
	c12TimeoutPattern       = regexp.MustCompile(`^[1-9][0-9]*(?:s|m)$`)
	c12Task4AllowedPackages = map[string]struct{}{
		"./internal/testinfra":             {},
		"./internal/store":                 {},
		"./internal/nodecontrol/contracts": {},
		"./internal/nodecontrol/authority": {},
		"./internal/nodecontrol/serving":   {},
		"./internal/readiness":             {},
	}
)

func TestC12IntegrationManifestValidatorUsesGoParserForExactCoverage(t *testing.T) {
	raw, sources := validC12ManifestFixture(t)
	if err := validateC12IntegrationManifests([][]byte{raw}, sources, "batch01", 120*time.Minute); err != nil {
		t.Fatal(err)
	}
}

func TestC12IntegrationManifestValidatorRejectsClosedSetMutations(t *testing.T) {
	raw, sources := validC12ManifestFixture(t)

	tests := []struct {
		name    string
		mutate  func(t *testing.T, manifest c12IntegrationManifest, sources []c12IntegrationSource) (c12IntegrationManifest, []c12IntegrationSource)
		wantErr string
	}{
		{
			name: "missing test",
			mutate: func(_ *testing.T, manifest c12IntegrationManifest, sources []c12IntegrationSource) (c12IntegrationManifest, []c12IntegrationSource) {
				manifest.Groups[0].Tests = manifest.Groups[0].Tests[:1]
				return manifest, sources
			},
			wantErr: "missing",
		},
		{
			name: "extra test",
			mutate: func(_ *testing.T, manifest c12IntegrationManifest, sources []c12IntegrationSource) (c12IntegrationManifest, []c12IntegrationSource) {
				manifest.Groups[0].Tests = append(manifest.Groups[0].Tests, "TestDoesNotExist")
				return manifest, sources
			},
			wantErr: "extra",
		},
		{
			name: "duplicate test",
			mutate: func(_ *testing.T, manifest c12IntegrationManifest, sources []c12IntegrationSource) (c12IntegrationManifest, []c12IntegrationSource) {
				manifest.Groups[0].Tests = []string{"TestAlpha", "TestAlpha", "TestBeta"}
				return manifest, sources
			},
			wantErr: "duplicate",
		},
		{
			name: "unsorted tests",
			mutate: func(_ *testing.T, manifest c12IntegrationManifest, sources []c12IntegrationSource) (c12IntegrationManifest, []c12IntegrationSource) {
				manifest.Groups[0].Tests = []string{"TestBeta", "TestAlpha"}
				return manifest, sources
			},
			wantErr: "sorted",
		},
		{
			name: "recursive package",
			mutate: func(_ *testing.T, manifest c12IntegrationManifest, sources []c12IntegrationSource) (c12IntegrationManifest, []c12IntegrationSource) {
				manifest.Groups[0].Package = "./..."
				return manifest, sources
			},
			wantErr: "explicit package",
		},
		{
			name: "package outside Task 4 allowlist",
			mutate: func(_ *testing.T, manifest c12IntegrationManifest, sources []c12IntegrationSource) (c12IntegrationManifest, []c12IntegrationSource) {
				manifest.Groups[0].Package = "./cmd/control-api"
				return manifest, sources
			},
			wantErr: "Task 4 allowed package set",
		},
		{
			name: "unsupported profile",
			mutate: func(_ *testing.T, manifest c12IntegrationManifest, sources []c12IntegrationSource) (c12IntegrationManifest, []c12IntegrationSource) {
				manifest.Groups[0].Profile = "production"
				return manifest, sources
			},
			wantErr: "profile",
		},
		{
			name: "oversized group timeout",
			mutate: func(_ *testing.T, manifest c12IntegrationManifest, sources []c12IntegrationSource) (c12IntegrationManifest, []c12IntegrationSource) {
				manifest.Groups[0].Timeout = "31m"
				return manifest, sources
			},
			wantErr: "30m",
		},
		{
			name: "untracked source",
			mutate: func(_ *testing.T, manifest c12IntegrationManifest, sources []c12IntegrationSource) (c12IntegrationManifest, []c12IntegrationSource) {
				sources[0].Tracked = false
				return manifest, sources
			},
			wantErr: "extra",
		},
		{
			name: "nonliteral first line",
			mutate: func(_ *testing.T, manifest c12IntegrationManifest, sources []c12IntegrationSource) (c12IntegrationManifest, []c12IntegrationSource) {
				sources[0].Body = []byte("package store_test\n\nfunc TestAlpha(t *testing.T) {}\nfunc TestBeta(t *testing.T) {}\n")
				return manifest, sources
			},
			wantErr: "extra",
		},
		{
			name: "skip call",
			mutate: func(_ *testing.T, manifest c12IntegrationManifest, sources []c12IntegrationSource) (c12IntegrationManifest, []c12IntegrationSource) {
				sources[0].Body = []byte("//go:build integration\n\npackage store_test\n\nimport \"testing\"\n\nfunc TestAlpha(t *testing.T) { t.Skip(\"no\") }\nfunc TestBeta(t *testing.T) {}\n")
				return manifest, sources
			},
			wantErr: "Skip",
		},
	}

	var manifest c12IntegrationManifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatal(err)
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manifestCopy := cloneC12IntegrationManifest(manifest)
			sourceCopy := cloneC12IntegrationSources(sources)
			manifestCopy, sourceCopy = test.mutate(t, manifestCopy, sourceCopy)
			mutated, err := json.Marshal(manifestCopy)
			if err != nil {
				t.Fatal(err)
			}
			err = validateC12IntegrationManifests([][]byte{mutated}, sourceCopy, "batch01", 120*time.Minute)
			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("validation error = %v, want fragment %q", err, test.wantErr)
			}
		})
	}
}

func TestC12IntegrationManifestValidatorRejectsUnknownJSONField(t *testing.T) {
	raw, sources := validC12ManifestFixture(t)
	mutated := bytesReplaceOnce(t, raw, []byte(`"schema":`), []byte(`"unknown":true,"schema":`))
	err := validateC12IntegrationManifests([][]byte{mutated}, sources, "batch01", 120*time.Minute)
	if err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("validation error = %v, want unknown field rejection", err)
	}
}

func TestC12FocusedTestMapUsesPackageLocalParserCoverage(t *testing.T) {
	packages := []string{
		"./internal/nodecontrol/contracts",
		"./internal/nodecontrol/authority",
		"./internal/nodecontrol/serving",
	}
	sources := []c12IntegrationSource{
		{Path: "internal/nodecontrol/contracts/contracts_integration_test.go", Tracked: true, Body: []byte("//go:build integration\n\npackage contracts_test\n\nimport \"testing\"\nfunc TestContractBoundary(t *testing.T) {}\n")},
		{Path: "internal/nodecontrol/authority/authority_integration_test.go", Tracked: true, Body: []byte("//go:build integration\n\npackage authority_test\n\nimport \"testing\"\nfunc TestAuthorityBoundary(t *testing.T) {}\n")},
		{Path: "internal/nodecontrol/serving/serving_integration_test.go", Tracked: true, Body: []byte("//go:build integration\n\npackage serving_test\n\nimport \"testing\"\nfunc TestServingBoundary(t *testing.T) {}\n")},
	}
	requested := []string{"TestAuthorityBoundary", "TestContractBoundary", "TestServingBoundary"}
	resolved, err := resolveC12FocusedTestMap(sources, packages, requested)
	if err != nil {
		t.Fatal(err)
	}
	for packageName, want := range map[string]string{
		"./internal/nodecontrol/contracts": "TestContractBoundary",
		"./internal/nodecontrol/authority": "TestAuthorityBoundary",
		"./internal/nodecontrol/serving":   "TestServingBoundary",
	} {
		if got := resolved[packageName]; len(got) != 1 || got[0] != want {
			t.Errorf("focused map[%s] = %v, want [%s]", packageName, got, want)
		}
	}

	if _, err := resolveC12FocusedTestMap(sources, packages, append(requested, "TestMissing")); err == nil || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("missing requested test error = %v", err)
	}
	ambiguous := append(cloneC12IntegrationSources(sources), c12IntegrationSource{
		Path: "internal/nodecontrol/contracts/shared_integration_test.go", Tracked: true,
		Body: []byte("//go:build integration\n\npackage contracts_test\n\nimport \"testing\"\nfunc TestAuthorityBoundary(t *testing.T) {}\n"),
	})
	if _, err := resolveC12FocusedTestMap(ambiguous, packages, requested); err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("ambiguous requested test error = %v", err)
	}
}

func TestC12SelectedIntegrationManifest(t *testing.T) {
	if *c12SelectedSuite == "" && *c12SelectedManifests == "" && *c12SelectedSuiteTimeout == "" && *c12CandidateTree == "" {
		raw, sources := validC12ManifestFixture(t)
		if err := validateC12IntegrationManifests([][]byte{raw}, sources, "batch01", 120*time.Minute); err != nil {
			t.Fatal(err)
		}
		return
	}
	if *c12SelectedSuite == "" || *c12SelectedManifests == "" || *c12SelectedSuiteTimeout == "" || *c12CandidateTree == "" {
		t.Fatal("selected C12 manifest validation requires suite, manifests, suite timeout, and candidate tree together")
	}
	if matched, _ := regexp.MatchString(`^[0-9a-f]{40}$`, *c12CandidateTree); !matched {
		t.Fatal("selected C12 candidate tree is not an exact Git tree identity")
	}
	suiteTimeout, err := time.ParseDuration(*c12SelectedSuiteTimeout)
	if err != nil {
		t.Fatal("parse selected C12 suite timeout:", err)
	}
	manifestPaths := strings.Split(*c12SelectedManifests, "|")
	if len(manifestPaths) == 0 {
		t.Fatal("selected C12 manifest list is empty")
	}
	repositoryRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	rawManifests := make([][]byte, 0, len(manifestPaths))
	for _, manifestPath := range manifestPaths {
		if manifestPath == "" || filepath.IsAbs(manifestPath) || strings.Contains(filepath.ToSlash(manifestPath), "../") {
			t.Fatalf("selected C12 manifest path %q is not a fixed repository-relative path", manifestPath)
		}
		raw, readErr := os.ReadFile(filepath.Join(repositoryRoot, filepath.FromSlash(manifestPath)))
		if readErr != nil {
			t.Fatalf("read selected C12 manifest %s: %v", manifestPath, readErr)
		}
		rawManifests = append(rawManifests, raw)
	}
	packages, err := c12ManifestPackages(rawManifests)
	if err != nil {
		t.Fatal(err)
	}
	sources, err := loadCandidateC12IntegrationSources(repositoryRoot, packages)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateC12IntegrationManifests(rawManifests, sources, *c12SelectedSuite, suiteTimeout); err != nil {
		t.Fatal(err)
	}
}

func TestC12ResolveFocusedIntegrationTests(t *testing.T) {
	if *c12FocusedPackages == "" && *c12FocusedTests == "" {
		return
	}
	if *c12FocusedPackages == "" || *c12FocusedTests == "" {
		t.Fatal("focused C12 resolver requires packages and tests together")
	}
	packages := strings.Split(*c12FocusedPackages, "|")
	requested := strings.Split(*c12FocusedTests, "|")
	repositoryRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	packageSet := make(map[string]struct{}, len(packages))
	for _, packageName := range packages {
		packageSet[packageName] = struct{}{}
	}
	sources, err := loadCandidateC12IntegrationSources(repositoryRoot, packageSet)
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := resolveC12FocusedTestMap(sources, packages, requested)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(resolved)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("C12_FOCUSED_MAP:%s\n", raw)
}

func TestC12BaseRunnerSourceIsClosed(t *testing.T) {
	raw, err := os.ReadFile("../../scripts/run-c12-integration.ps1")
	if err != nil {
		t.Fatalf("read C12 integration runner: %v", err)
	}
	source := string(raw)
	required := []string{
		"postgres:18.4-alpine3.23",
		"redis:8.8.1-alpine3.23",
		"nats:2.14.3-alpine3.22",
		"127.0.0.1::5432",
		"127.0.0.1::6379",
		"127.0.0.1::4222",
		"& docker @dockerArgs",
		"& go @goArgs",
		"$LASTEXITCODE",
		"finally",
		"goose",
		"up-to",
		"-tags=integration",
		"-p=1",
		"-json",
		"talenro-c12-trusted-validator/v1",
		"Invoke-C12TrustedValidator",
	}
	for _, literal := range required {
		if !strings.Contains(source, literal) {
			t.Errorf("runner lacks required closed-boundary literal %q", literal)
		}
	}
	for _, forbidden := range []string{"Invoke-Expression", "cmd /c", "powershell -Command", "ProcessStartInfo"} {
		if strings.Contains(strings.ToLower(source), strings.ToLower(forbidden)) {
			t.Errorf("runner contains forbidden command construction %q", forbidden)
		}
	}
}

func TestC12RunnerAuthorityProfilesAreClosed(t *testing.T) {
	source := string(readC12RunnerSource(t))
	for _, required := range []string{
		"$script:c12AllowedProfiles = @('base', 'authority-v7', 'authority-v7-pitr')",
		"'base' = [TimeSpan]::FromMinutes(3)",
		"'authority-v7' = [TimeSpan]::FromMinutes(3)",
		"'authority-v7-pitr' = [TimeSpan]::FromMinutes(8)",
		"TestPrepareC12AuthorityV7Database",
		"TestC12DependenciesAreIsolatedAndAuthorityV7Migrated",
		"TestC12AuthorityPITRProfile",
		"TestC12AuthorityPITROwnershipWALFailureSeam",
	} {
		if !strings.Contains(source, required) {
			t.Errorf("authority runner profile registry lacks %q", required)
		}
	}
	if strings.Contains(source, "authority-v7-raw") || strings.Contains(source, "authority_v7.sql") {
		t.Fatal("runner exposes a raw v7 migration profile")
	}
}

func TestC12PreparedAuthorityExecutablesAreClosed(t *testing.T) {
	source := string(readC12RunnerSource(t))
	for _, required := range []string{
		"$script:c12PreparedAuthorityRoles = @('trusted-validator', 'authority-initializer-json')",
		"$script:c12GoToolchainVersion = 'go1.26.5'",
		"function New-C12PreparedArtifactRoot",
		"function Resolve-C12ClosedGoToolchain",
		"function Invoke-C12ClosedGoBuild",
		"function New-C12PreparedTrustedValidator",
		"function New-C12PreparedAuthorityInitializer",
		"function Invoke-C12PreparedAuthorityRole",
		"$validatorBuildArguments = @('build', '-o', $temporaryExecutable, $sourcePath)",
		"$initializerBuildArguments = @('test', '-c', '-tags=integration', '-p=1', '-o', $temporaryTestExecutable, './internal/testinfra')",
		"$test2JSONBuildArguments = @('build', '-o', $temporaryTest2JSONExecutable, '.')",
		"'GOFLAGS', 'GOWORK', 'GOENV', 'GOTOOLCHAIN'",
		"'GOCACHE', 'GOTMPDIR', 'GOMODCACHE', 'GOPROXY'",
	} {
		if !strings.Contains(source, required) {
			t.Errorf("prepared authority executable contract lacks %q", required)
		}
	}
	for _, forbidden := range []string{
		"[ValidateSet('trusted-validator', 'authority-initializer-json',",
		"'direct-executable'",
		"'trusted-validator-probe'",
	} {
		if strings.Contains(source, forbidden) {
			t.Errorf("prepared authority executable contract exposes %q", forbidden)
		}
	}
}

func TestC12InitializerPreparationIsCompileOnly(t *testing.T) {
	source := string(readC12RunnerSource(t))
	preparation := c12PowerShellFunction(t, source, "New-C12PreparedAuthorityInitializer")
	for _, required := range []string{
		"$initializerBuildArguments = @('test', '-c', '-tags=integration', '-p=1', '-o', $temporaryTestExecutable, './internal/testinfra')",
		"Resolve-C12Test2JSONExecutable",
		"New-C12SealedExecutableReceipt",
	} {
		if !strings.Contains(preparation, required) {
			t.Errorf("compile-only initializer preparation lacks %q", required)
		}
	}
	for _, forbidden := range []string{
		"'-run', '^$'",
		"TALENRO_C12_AUTHORITY_V7_INIT_NONCE",
		"TALENRO_DATABASE_URL",
		"Invoke-C12PreparedAuthorityRole",
		"& $temporaryTestExecutable",
	} {
		if strings.Contains(preparation, forbidden) {
			t.Errorf("compile-only initializer preparation can execute or receive candidate authority through %q", forbidden)
		}
	}
	test2JSONPreparation := c12PowerShellFunction(t, source, "Resolve-C12Test2JSONExecutable")
	if !strings.Contains(test2JSONPreparation, "$test2JSONBuildArguments = @('build', '-o', $temporaryTest2JSONExecutable, '.')") {
		t.Error("compile-only initializer preparation does not setup-build the exact sealed test2json executable")
	}
}

func TestC12PreparedArtifactIdentityAndCleanup(t *testing.T) {
	source := string(readC12RunnerSource(t))
	for _, required := range []string{
		"public sealed class C12SealedExecutable",
		"VolumeSerialNumber",
		"FileIndexHigh",
		"FileIndexLow",
		"NumberOfLinks",
		"FileShare.Read",
		"GetSecurityDescriptorSddlForm",
		"function New-C12SealedExecutableReceipt",
		"function Assert-C12SealedExecutableReceipt",
		"ExecutableSHA256",
		"ExecutableLength",
		"ExecutableVolumeSerial",
		"ExecutableFileIndex",
		"ExecutableLinkCount",
		"ExecutableOwner",
		"ExecutableDACL",
		"ExecutableReparse",
		"ReceiptSeal",
		"receipt replay",
		"Remove-C12PreparedArtifactRoot",
	} {
		if !strings.Contains(source, required) {
			t.Errorf("prepared artifact identity/cleanup contract lacks %q", required)
		}
	}
}

func TestC12PreparedNativeJobGateIsClosed(t *testing.T) {
	source := string(readC12RunnerSource(t))
	worker := c12PowerShellAssignment(t, source, "$script:c12PreparedNativeWorkerScript = {")
	ordered := []string{
		"AssignProcessToJobObject",
		"Assert-C12SealedExecutableReceipt",
		"JOB_MEMBER_READY",
		"WaitOne",
		"FORMAL_RELEASE",
		"EXEC_BEGIN",
		"EXEC_END",
	}
	previous := -1
	for _, marker := range ordered {
		index := strings.Index(worker, marker)
		if index < 0 {
			t.Errorf("prepared native worker lacks %q", marker)
			continue
		}
		if index <= previous {
			t.Errorf("prepared native worker phase %q is out of order", marker)
		}
		previous = index
	}
	for _, required := range []string{"CreateKillOnClose", "EventWaitHandle", "exact-once", "[TimeSpan]::FromMinutes(2)", "[C12NativeJob]::Close"} {
		if !strings.Contains(source, required) {
			t.Errorf("prepared native Job gate contract lacks %q", required)
		}
	}
}

func TestC12PreparedArtifactLifecycleIsCoherent(t *testing.T) {
	source := string(readC12RunnerSource(t))
	for _, required := range []string{
		"$script:c12PreparedArtifactRoot",
		"$script:c12PreparedReceipts",
		"ArtifactRootIdentity",
		"ReceiptIdentity",
		"Remove-C12PreparedArtifactRoot",
		"Close-C12PreparedNativeWorker",
		"function Close-C12AllPreparedArtifacts",
		"Close-C12AllPreparedArtifacts -Deadline ([DateTime]::UtcNow.Add($script:c12GroupCleanupBudget))",
	} {
		if !strings.Contains(source, required) {
			t.Errorf("prepared artifact lifecycle lacks %q", required)
		}
	}
	for _, forbidden := range []string{"$script:c12GoPrewarmPurposes", "talenro-c12-validator-$(New-C12RandomSuffix)"} {
		if strings.Contains(source, forbidden) {
			t.Errorf("prepared artifact lifecycle retains incoherent per-call prewarm state %q", forbidden)
		}
	}
	group := c12PowerShellFunction(t, source, "Invoke-C12Group")
	prepare := strings.Index(group, "New-C12PreparedAuthorityInitializer")
	container := strings.Index(group, "Start-C12Container")
	baseMigration := strings.Index(group, "ordinary Goose base migration")
	formal := strings.Index(group, "Invoke-C12AuthorityInitializer")
	if prepare < 0 || container < 0 || prepare >= container {
		t.Error("authority initializer is not prepared before container/resource startup")
	}
	if baseMigration < 0 || formal < 0 || formal <= baseMigration {
		t.Error("authority initializer formal release does not follow the ordinary base migration")
	}
}

func TestC12FormalAuthorityStagesDoNotInvokeGo(t *testing.T) {
	source := string(readC12RunnerSource(t))
	validator := c12PowerShellFunction(t, source, "Invoke-C12TrustedValidator")
	initializer := c12PowerShellFunction(t, source, "Invoke-C12AuthorityInitializer")
	for name, body := range map[string]string{"trusted validator": validator, "authority initializer": initializer} {
		if !strings.Contains(body, "Invoke-C12PreparedAuthorityRole") {
			t.Errorf("formal %s does not release a sealed prepared role", name)
		}
		for _, forbidden := range []string{"Invoke-C12Go", "-Executable 'go'", "'go', 'run'", "'go', 'test'", "'go', 'build'", "'go', 'tool'"} {
			if strings.Contains(body, forbidden) {
				t.Errorf("formal %s re-enters the Go toolchain through %q", name, forbidden)
			}
		}
	}
	for _, required := range []string{
		"Invoke-C12PreparedAuthorityRole -Prepared $preparedValidator -Role 'trusted-validator' -Timeout ([TimeSpan]::FromMinutes(2))",
		"Invoke-C12PreparedAuthorityRole -Prepared $preparedInitializer -Role 'authority-initializer-json' -Timeout ([TimeSpan]::FromMinutes(2))",
		"Assert-C12GoJSONResult",
	} {
		if !strings.Contains(source, required) {
			t.Errorf("formal authority handoff lacks %q", required)
		}
	}
}

func TestC12PreparedArtifactExecutionAndTamperConverge(t *testing.T) {
	for _, mode := range []string{"success", "initializer-capabilities", "byte-tamper", "file-id-splice", "reparse", "hard-link", "dacl", "wrong-role", "replay"} {
		t.Run(mode, func(t *testing.T) {
			output, exitCode := runC12PreparedArtifactHarness(t, mode)
			if exitCode != 0 {
				t.Fatalf("prepared artifact %s harness exit=%d output=%q", mode, exitCode, output)
			}
		})
	}
}

// A rejected authenticated frame must never release a payload and must converge
// the exact retained process and Job, including replay of a previously valid frame.
func TestC12PreparedProtocolRejectsMalformedFrames(t *testing.T) {
	for _, mode := range []string{"duplicate-sequence", "reordered-frame", "foreign-pid", "wrong-nonce", "oversized-frame", "previous-digest", "bad-hmac", "bad-hmac-cleanup", "invalid-utf8", "noncanonical-json"} {
		t.Run(mode, func(t *testing.T) {
			output, code := runC12PreparedArtifactHarness(t, "protocol-"+mode)
			if code != 0 {
				t.Fatalf("protocol rejection %s exit=%d output=%q", mode, code, output)
			}
		})
	}
}

// A prepared worker may remain suspended longer than its preparation allowance
// plus two minutes. Only the authenticated invocation deadline can bound release.
func TestC12PreparedInvocationDeadlineIsReleaseRelative(t *testing.T) {
	for _, mode := range []string{"delayed-preparation-release", "deadline-wire-tamper", "deadline-malformed", "deadline-expired"} {
		t.Run(mode, func(t *testing.T) {
			output, code := runC12PreparedArtifactHarness(t, mode)
			if code != 0 {
				t.Fatalf("invocation deadline %s exit=%d output=%q", mode, code, output)
			}
		})
	}
}

func TestC12RunnerAuthorityCapabilitiesAreChildScoped(t *testing.T) {
	source := string(readC12RunnerSource(t))
	for _, required := range []string{
		"TALENRO_C12_AUTHORITY_V7_INIT_NONCE", "TALENRO_C12_AUTHORITY_V7_RUN_SUFFIX",
		"TALENRO_C12_AUTHORITY_V7_PROFILE", "disposable_fixture",
		"Clear-C12InheritedCapabilities", "GIT_*", "finally",
	} {
		if !strings.Contains(source, required) {
			t.Errorf("authority child capability boundary lacks %q", required)
		}
	}
	for _, forbidden := range []string{"-Environment $env:", "Get-ChildItem Env: |", "Start-Process -Environment"} {
		if strings.Contains(source, forbidden) {
			t.Errorf("authority runner exposes arbitrary inherited environment via %q", forbidden)
		}
	}
}

func TestC12RunnerPITRWALAndCleanupAreClosed(t *testing.T) {
	source := string(readC12RunnerSource(t))
	for _, required := range []string{
		"postgres:18.4-alpine3.23", "test_decoding", "TALENRO_C12_AUTHORITY_PITR_NONCE",
		"TALENRO_C12_AUTHORITY_PITR_DESCRIPTOR", "TALENRO_C12_AUTHORITY_PITR_WAL_PATH",
		"TALENRO_C12_AUTHORITY_PITR_HMAC_KEY", "HMACSHA256", "8192", "4096",
		"BOOTSTRAP", "INTENT", "ACTUAL", "RECOVERED_ACTUAL", "NOT_FOUND", "TRANSITION", "CLEAN_INTENT", "CLEAN_RESULT",
		"[TimeSpan]::FromSeconds(75)", "talenro.c12.run",
	} {
		if !strings.Contains(source, required) {
			t.Errorf("PITR ownership WAL/cleanup boundary lacks %q", required)
		}
	}
	for _, forbidden := range []string{"docker system prune", "container prune", "volume prune", "network prune", "--filter name=", "Get-ChildItem -Recurse"} {
		if strings.Contains(strings.ToLower(source), strings.ToLower(forbidden)) {
			t.Errorf("PITR cleanup contains broad discovery/removal surface %q", forbidden)
		}
	}
}

func TestC12RunnerPITRFailureSeamsAreClosed(t *testing.T) {
	source := string(readC12RunnerSource(t))
	want := []string{
		"after-intent-before-create", "after-create-before-actual", "after-actual-before-return",
		"after-clean-intent-before-remove", "after-remove-before-clean-result",
	}
	for _, seam := range want {
		if got := strings.Count(source, "'"+seam+"'"); got == 0 {
			t.Errorf("PITR failure seam %q is absent", seam)
		}
	}
	for _, required := range []string{"authority-v7-pitr", "TestC12AuthorityPITROwnershipWALFailureSeam", "foreign canary"} {
		if !strings.Contains(strings.ToLower(source), strings.ToLower(required)) {
			t.Errorf("PITR seam closure lacks %q", required)
		}
	}
}

// TestC12PreparedProcessIsSuspendedUntilVerifiedJobMembership catches the
// production escape window created by Start-Job: the worker begins executing
// before the controller has assigned and verified its native Job membership.
func TestC12PreparedProcessIsSuspendedUntilVerifiedJobMembership(t *testing.T) {
	snapshot := c12PowerShellExecutableAST(t)
	worker := c12ReachablePowerShellNodes(t, snapshot, "New-C12PreparedNativeWorker")
	// Embedded C# is one PowerShell AST node, so call ordering is proved by
	// the native sentinel below, not by four separate PowerShell text nodes.
	c12RequirePowerShellASTTerms(t, "suspended prepared worker", worker, []string{
		`(?i)CreateProcessW`, `(?i)AssignProcessToJobObject`, `(?i)IsProcessInJob`, `(?i)ResumeThread`,
	})
	c12RequirePowerShellASTTerms(t, "suspended prepared worker", worker, []string{
		`CREATE_SUSPENDED`, `CREATE_NO_WINDOW`, `CREATE_UNICODE_ENVIRONMENT`, `(?i)bInheritHandles\s*[:=]\s*\$?false`,
		`(?i)TerminateProcess`, `(?i)TerminateJobObject`, `JOB_OBJECT_MSG_ACTIVE_PROCESS_ZERO`,
		`(?i)GetQueuedCompletionStatus`, `(?i)QueryInformationJobObject`, `ActiveProcesses`,
	})
	c12ForbidPowerShellCommand(t, "suspended prepared worker", worker, "Start-Job")
	for _, mode := range []string{"prepared-suspended-membership", "prepared-assignment-failure", "prepared-pid-mismatch", "prepared-invalid-thread-handle", "prepared-cleanup-retry", "prepared-job-collision", "prepared-exit-race", "prepared-exit-pending", "prepared-exit-pending-retry"} {
		t.Run(mode, func(t *testing.T) {
			output, exitCode := runC12PreparedArtifactHarness(t, mode)
			if exitCode != 0 {
				t.Fatalf("production suspended-worker sentinel harness exit=%d output=%q", exitCode, output)
			}
		})
	}
}

// TestC12PreparedPipeProtocolAndProjectionAreClosed freezes the six private
// frames and the minimal projection. It rejects the current full-receipt,
// release-before-capability, byte-stream handoff.
func TestC12PreparedPipeProtocolAndProjectionAreClosed(t *testing.T) {
	c12AssertLiteralPreparedFrames(t)
	snapshot := c12PowerShellExecutableAST(t)
	nodes := c12ReachablePowerShellNodes(t, snapshot, "$script:c12PreparedNativeWorkerScript", "Invoke-C12PreparedAuthorityRole")
	c12RequirePowerShellASTTerms(t, "prepared private pipe", nodes, []string{
		`PIPE_ACCESS_DUPLEX`, `FILE_FLAG_FIRST_PIPE_INSTANCE`, `PIPE_TYPE_MESSAGE`, `PIPE_READMODE_MESSAGE`,
		`PIPE_WAIT`, `PIPE_REJECT_REMOTE_CLIENTS`, `(?i)nMaxInstances\s*[:=]\s*1`,
		`(?i)GetNamedPipeClientProcessId`, `(?i)GetNamedPipeServerProcessId`,
	})
	c12RequirePowerShellASTSequence(t, "prepared private protocol", nodes, []string{
		`HELLO`, `VERIFICATION_PROJECTION`, `JOB_MEMBER_READY`, `CAPABILITIES`, `CAPABILITIES_INSTALLED`, `GATE_WAITING`,
	})
	// Protocol and release are distinct sibling calls; behavioral trace assertions
	// below prove their ordering across that boundary.
	c12RequirePowerShellASTSequence(t, "formal gate release", c12ReachablePowerShellNodes(t, snapshot, "Release-C12PreparedWorker"), []string{
		`^StartNew$`, `FORMAL_RELEASE`, `^Set$`,
	})
	projectionNodes := c12ReachablePowerShellNodes(t, snapshot, "$script:c12PreparedNativeWorkerScript", "New-C12VerificationProjection")
	for _, forbidden := range []string{`Receipt\s*=`, `ReceiptFields`, `PayloadFields`, `ReceiptKeyHex`, `Graph`, `RootHandle`, `RefCount`} {
		c12ForbidPowerShellASTTerm(t, "minimal worker projection", projectionNodes, forbidden)
	}
	c12ForbidPowerShellCommand(t, "prepared pipe path", nodes, "Start-Job")
	output, exitCode := runC12PreparedArtifactHarness(t, "prepared-pipe-six-frame")
	if exitCode != 0 {
		t.Fatalf("production six-frame pipe harness exit=%d output=%q", exitCode, output)
	}
}

// TestC12PreparedGoGraphsBindEverySelectedInput freezes the three setup-only
// graph commands, finite bounds, and every selected Go/toolchain content set.
func TestC12PreparedGoGraphsBindEverySelectedInput(t *testing.T) {
	t.Run("discovery-policy", func(t *testing.T) {
		snapshot := c12PowerShellExecutableAST(t)
		nodes := c12ReachablePowerShellNodes(t, snapshot,
			"New-C12PreparedTrustedValidator", "New-C12PreparedAuthorityInitializer", "Resolve-C12Test2JSONExecutable")
		c12RequirePowerShellASTTerms(t, "prepared Go graph closure", nodes, []string{
			`(?i)\blist\b.*-deps.*-json`, `(?i)-test.*-tags=integration`, `cmd/test2json`,
			`\b8192\b`, `\b131072\b`, `\b67108864\b`, `\b16777216\b`, `GOSUMDB`, `-mod=readonly`, `(?i)mod\s+verify`,
		})
		graphFields := []string{
			"schema", "purpose", "go_executable_path", "go_executable_identity", "go_executable_sha256", "go_version",
			"goroot_path", "goroot_identity", "gomodcache_path", "gomodcache_identity", "module_root_path", "module_root_identity",
			"goos", "goarch", "cgo_enabled", "go_mod_sha256", "go_sum_sha256", "candidate_tree_digest", "build_argv",
			"package_count", "file_count", "packages", "import_path", "for_test", "origin", "module_path", "module_version",
			"module_sum", "module_replace_or_null", "directory_identity", "imports", "deps", "test_imports", "xtest_imports",
			"embed_patterns", "test_embed_patterns", "xtest_embed_patterns", "files", "set", "origin_relative_path", "length", "sha256",
		}
		for _, field := range graphFields {
			c12RequirePowerShellASTTerms(t, "prepared Go graph registry", nodes, []string{regexp.QuoteMeta(field)})
		}
		for _, selectedSet := range []string{
			"GoFiles", "CgoFiles", "CFiles", "CXXFiles", "MFiles", "HFiles", "FFiles", "SFiles", "SwigFiles",
			"SwigCXXFiles", "SysoFiles", "EmbedPatterns", "EmbedFiles", "TestGoFiles", "TestEmbedPatterns",
			"TestEmbedFiles", "XTestGoFiles", "XTestEmbedPatterns", "XTestEmbedFiles",
		} {
			c12RequirePowerShellASTTerms(t, "prepared Go selected sets", nodes, []string{regexp.QuoteMeta(selectedSet)})
		}
	})
	t.Run("receipts-and-mutations", func(t *testing.T) {
		output, exitCode := runC12PreparedGoGraphHarness(t)
		if exitCode != 0 {
			t.Fatalf("production prepared graph/receipt harness exit=%d output=%q", exitCode, output)
		}
	})
}

func TestC12GoGraphReceiptPurposesRequireIndependentExactLabels(t *testing.T) {
	runner := readC12RunnerSource(t)
	index := bytes.Index(runner, []byte("$script:c12RepositoryRoot = (Resolve-Path"))
	if index < 0 {
		t.Fatal("runner lacks main-program marker")
	}
	const appendix = `
# Isolate the production purpose gate from later filesystem sealing. A valid
# purpose set must reach the deliberately malformed source identity afterward.
$ownership = [pscustomobject]@{}
$ownership | Add-Member -MemberType ScriptMethod -Name VerifyExactPath -Value {}
$artifact = [pscustomobject]@{ Closed=$false; Profile='base'; Ownership=$ownership }
$cases = @(
  @{ name='concatenated'; purpose='authority-initializer'; labels=@('authority-initializer|test2json'); pass=$false },
  @{ name='duplicate'; purpose='authority-initializer'; labels=@('authority-initializer','authority-initializer'); pass=$false },
  @{ name='missing'; purpose='authority-initializer'; labels=@('authority-initializer'); pass=$false },
  @{ name='empty'; purpose='authority-initializer'; labels=@(); pass=$false },
  @{ name='case-changed'; purpose='authority-initializer'; labels=@('authority-initializer','TEST2JSON'); pass=$false },
  @{ name='valid'; purpose='authority-initializer'; labels=@('authority-initializer','test2json'); pass=$true },
  @{ name='valid-reversed'; purpose='authority-initializer'; labels=@('test2json','authority-initializer'); pass=$true },
  @{ name='validator-valid'; purpose='trusted-validator'; labels=@('trusted-validator'); pass=$true },
  @{ name='validator-duplicate'; purpose='trusted-validator'; labels=@('trusted-validator','trusted-validator'); pass=$false }
)
$failures = [Collections.Generic.List[string]]::new()
foreach ($case in $cases) {
  $graphs = @($case.labels | ForEach-Object { [pscustomobject]@{ Purpose=$_ } })
  $message = ''
  try {
    $null = New-C12SealedExecutableReceipt -ArtifactRoot $artifact -Role 'authority-initializer-json' -Profile 'base' -Purpose $case.purpose -SourceIdentity 'fixture' -SourceDigest 'intentionally-invalid' -CandidateTreeIdentity 'intentionally-invalid' -BuildArguments @('build') -GoToolchain ([pscustomobject]@{}) -ExecutablePath 'unused.exe' -Arguments @('fixture') -WorkingDirectory 'unused' -GoGraphReceipts $graphs
  } catch { $message = $_.Exception.Message }
  $expected = if ($case.pass) { 'prepared receipt closed identity is malformed' } else { 'prepared receipt requires its independent purpose-labelled Go graphs' }
  if ($message -cne $expected) { $failures.Add($case.name + ': ' + $message) }
  else { Write-Output ('C12_GRAPH_PURPOSE_OK:' + $case.name) }
}
if ($failures.Count) { throw ('purpose gate rejected/accepted the wrong set: ' + ($failures -join '; ')) }
`
	root := t.TempDir()
	path := filepath.Join(root, "scripts", "run-c12-integration.ps1")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(append([]byte(nil), runner[:index]...), appendix...), 0o600); err != nil {
		t.Fatal(err)
	}
	output, code := runC12PowerShellAtRootWithTimeout(t, root, 2*time.Minute, nil, "-Profile", "base", "-Packages", "./internal/testinfra", "-Timeout", "3m")
	if code != 0 {
		t.Fatalf("purpose gate exit=%d output=%q", code, output)
	}
	for _, name := range []string{"concatenated", "duplicate", "missing", "empty", "case-changed", "valid", "valid-reversed", "validator-valid", "validator-duplicate"} {
		if !strings.Contains(output, "C12_GRAPH_PURPOSE_OK:"+name) {
			t.Errorf("missing purpose gate case %s: %s", name, output)
		}
	}
}

func TestC12GoGraphReceiptBindingPreservesEveryBodyByte(t *testing.T) {
	runner := readC12RunnerSource(t)
	index := bytes.Index(runner, []byte("$script:c12RepositoryRoot = (Resolve-Path"))
	if index < 0 {
		t.Fatal("runner lacks main-program marker")
	}
	const appendix = `
$receipt = [pscustomobject]@{}
foreach ($field in $script:c12PreparedReceiptPayloadFields) { $receipt | Add-Member -NotePropertyName $field -NotePropertyValue 'fixture' }
$graph = [pscustomobject]@{ BodyBytes=[byte[]](0..255); Purpose='trusted-validator'; Digest='fixture'; Files=@('selected-file') }
$receipt.GoGraphReceipts = @($graph)
$first = Get-C12PreparedReceiptPayload $receipt
$graphLine = ([Text.Encoding]::UTF8.GetString($first) -split [char]10 | Where-Object { $_.StartsWith('GoGraphReceipts=') })
$graphJSON = [Text.Encoding]::UTF8.GetString([Convert]::FromBase64String($graphLine.Substring('GoGraphReceipts='.Length)))
$expectedBytes = '"BodyBytes":{"content":"' + [Convert]::ToBase64String($graph.BodyBytes) + '","encoding":"base64","length":256}'
if (-not $graphJSON.Contains($expectedBytes)) { throw 'graph BodyBytes is not bound as full deterministic base64 with exact length' }
$key = [byte[]](1..32)
$mac = [Security.Cryptography.HMACSHA256]::new($key)
try {
  $baseline = [Convert]::ToBase64String($mac.ComputeHash($first))
  $graph.BodyBytes = [byte[]](0..255)
  if ([Convert]::ToBase64String($mac.ComputeHash((Get-C12PreparedReceiptPayload $receipt))) -cne $baseline) { throw 'identical body bytes changed receipt binding' }
  foreach ($offset in 0..255) {
    $graph.BodyBytes[$offset] = $graph.BodyBytes[$offset] -bxor 1
    if ([Convert]::ToBase64String($mac.ComputeHash((Get-C12PreparedReceiptPayload $receipt))) -ceq $baseline) { throw "body byte mutation $offset preserved receipt binding" }
    $graph.BodyBytes[$offset] = $graph.BodyBytes[$offset] -bxor 1
  }
  $graph.BodyBytes = [byte[]](0..254)
  if ([Convert]::ToBase64String($mac.ComputeHash((Get-C12PreparedReceiptPayload $receipt))) -ceq $baseline) { throw 'body length mutation preserved receipt binding' }
  $graph.BodyBytes = [byte[]](0..255)
  $graph.Files = @('changed-file')
  if ([Convert]::ToBase64String($mac.ComputeHash((Get-C12PreparedReceiptPayload $receipt))) -ceq $baseline) { throw 'graph metadata mutation preserved receipt binding' }
} finally { $mac.Dispose() }
Write-Output 'C12_GRAPH_BYTE_BINDING_OK'
`
	root := t.TempDir()
	path := filepath.Join(root, "scripts", "run-c12-integration.ps1")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(append([]byte(nil), runner[:index]...), appendix...), 0o600); err != nil {
		t.Fatal(err)
	}
	output, code := runC12PowerShellAtRootWithTimeout(t, root, 3*time.Minute, nil, "-Profile", "base", "-Packages", "./internal/testinfra", "-Timeout", "3m")
	if code != 0 || !strings.Contains(output, "C12_GRAPH_BYTE_BINDING_OK") {
		t.Fatalf("byte binding exit=%d output=%q", code, output)
	}
}

func TestC12GoGraphVerificationDefaultsToProfileDeadline(t *testing.T) {
	runner := readC12RunnerSource(t)
	index := bytes.Index(runner, []byte("$script:c12RepositoryRoot = (Resolve-Path"))
	if index < 0 {
		t.Fatal("runner lacks main-program marker")
	}
	const appendix = `
$body = '{"toolchain":{"go_executable_path":"unused","go_executable_sha256":"unused","go_version":"go1.26.5","goroot_path":"unused","gomodcache_path":"unused"},"environment":{"goos":"windows","goarch":"amd64"},"purpose":"trusted-validator","selected_packages":["fixture.go"],"build_argv":["build","fixture.go"],"working_directory":"unused"}'
$receipt = [pscustomobject]@{ BodyBytes = [Text.Encoding]::UTF8.GetBytes($body) }
function New-C12GoGraphReceipt {
  param($Purpose,$Packages,[DateTime]$Deadline,$ArtifactRoot,$GoToolchain,$BuildArgv,$WorkingDirectory)
  $script:observedDeadline = $Deadline
  return $receipt
}
$before = [DateTime]::UtcNow
Assert-C12GoGraphReceipt -Receipt $receipt -ArtifactRoot ([pscustomobject]@{ Profile='base' }) -Deadline ([DateTime]::MaxValue)
$after = [DateTime]::UtcNow
if ($script:observedDeadline.Kind -ne [DateTimeKind]::Utc -or $script:observedDeadline -lt $before.AddMinutes(3) -or $script:observedDeadline -gt $after.AddMinutes(3)) {
  throw 'default graph verification deadline is not bounded by the closed profile in UTC'
}
$explicit = [DateTime]::UtcNow.AddSeconds(30)
Assert-C12GoGraphReceipt -Receipt $receipt -ArtifactRoot ([pscustomobject]@{ Profile='base' }) -Deadline $explicit
if ($script:observedDeadline -ne $explicit) { throw 'explicit graph verification deadline was changed' }
Write-Output 'C12_GRAPH_DEFAULT_DEADLINE_OK'
`
	root := t.TempDir()
	path := filepath.Join(root, "scripts", "run-c12-integration.ps1")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(append([]byte(nil), runner[:index]...), appendix...), 0o600); err != nil {
		t.Fatal(err)
	}
	output, code := runC12PowerShellAtRootWithTimeout(t, root, 2*time.Minute, nil, "-Profile", "base", "-Packages", "./internal/testinfra", "-Timeout", "3m")
	if code != 0 || !strings.Contains(output, "C12_GRAPH_DEFAULT_DEADLINE_OK") {
		t.Fatalf("default graph deadline exit=%d output=%q", code, output)
	}
}

func TestC12PreparedArtifactDirectLeafLedgerIsClosed(t *testing.T) {
	snapshot := c12PowerShellExecutableAST(t)
	nodes := c12ReachablePowerShellNodes(t, snapshot, "New-C12PreparedArtifactRoot", "Remove-C12PreparedArtifactRoot")
	c12RequirePowerShellASTTerms(t, "prepared direct-leaf ledger", nodes, []string{
		"Expected", "CreateAttempted", "Bound", "Removed", "Absent", "NeverAttempted", "exact_file",
		"owned_ephemeral_subtree", "go-cache", "go-tmp", "creation phase", "unknown direct sibling",
		"NumberOfLinks", "RefCount", "Lifecycle", "InspectDirectory",
	})
	for _, forbidden := range []string{"Remove-C12BoundedDirectory", "-Recurse", "Remove-Item -LiteralPath $ArtifactRoot.Root"} {
		c12ForbidPowerShellASTTerm(t, "prepared direct-leaf cleanup", nodes, regexp.QuoteMeta(forbidden))
	}
	output, exitCode := runC12CleanupStateHarness(t, "prepared-direct-ledger")
	if exitCode != 0 {
		t.Fatalf("production direct-leaf retained-state harness exit=%d output=%q", exitCode, output)
	}
}

func TestC12PITRRunRootHasExactFiveLeafLedger(t *testing.T) {
	snapshot := c12PowerShellExecutableAST(t)
	nodes := c12ReachablePowerShellNodes(t, snapshot, "New-C12PITRRunRoot")
	wantLeaves := []string{"controller-ownership-v1.wal", "ownership.wal", "tlsgen.go", "server.crt", "server.key"}
	for _, leaf := range wantLeaves {
		c12RequirePowerShellASTExactLiteralCount(t, "PITR five-leaf ledger", nodes, leaf, 1)
	}
	c12RequirePowerShellASTTerms(t, "PITR five-leaf ledger", nodes, []string{"exact_file", "CreationClosed", "NeverAttempted", "Remove", "unknown sibling", "direct"})
	output, exitCode := runC12CleanupStateHarness(t, "pitr-five-leaf")
	if exitCode != 0 {
		t.Fatalf("production PITR five-leaf harness exit=%d output=%q", exitCode, output)
	}
}

func TestC12CleanupRetryStateRetainsExactOwnership(t *testing.T) {
	snapshot := c12PowerShellExecutableAST(t)
	nodes := c12ReachablePowerShellNodes(t, snapshot, "Remove-C12PreparedArtifactRoot", "Remove-C12PITRBaseResources")
	c12RequirePowerShellASTTerms(t, "cleanup retry ownership state", nodes, []string{
		"CreationOpen", "CreationClosed", "NeverAttempted", "CreateAttempted", "Created", "Verified", "CleanIntent",
		"Removed", "Absent", "75", "RetryState", "RootHandle", "ProcessHandle", "WALHandle", "Ledger", "top-level finalizer",
	})
	c12ForbidPowerShellASTTerm(t, "cleanup retry ownership state", nodes, `(?i)\.CreateAttempted\s*=\s*\$true`)
	output, exitCode := runC12CleanupStateHarness(t, "cleanup-retained-production")
	if exitCode != 0 {
		t.Fatalf("production retained cleanup state harness exit=%d output=%q", exitCode, output)
	}
}

func TestC12ControllerOwnershipWALUsesExactBytes(t *testing.T) {
	c12AssertLiteralControllerWAL(t)
	snapshot := c12PowerShellExecutableAST(t)
	nodes := c12ReachablePowerShellNodes(t, snapshot, "New-C12PITRRunRoot", "Remove-C12PITRBaseResources")
	c12RequirePowerShellASTTerms(t, "controller ownership WAL", nodes, []string{
		"talenro-c12-authority-pitr-controller-ownership-wal/v1", "controller-ownership-v1.wal", "WriteThrough",
		"FlushFileBuffers", "BOOTSTRAP", "INTENT", "ACTUAL", "NOT_FOUND", "CLEAN_INTENT", "CLEAN_RESULT",
		"talenro.c12.controller-wal.registry.v1", "talenro.c12.controller-wal.payload.v1",
		"talenro.c12.controller-wal.record.v1", "talenro.c12.controller-wal.hmac.v1", "record_digest", "hmac_sha256",
		"yyyy-MM-ddTHH:mm:ss.fffffffZ", "previous_record_digest", "payload_digest",
	})
	for _, forbidden := range []string{"ConvertTo-Json", "ConvertFrom-Json"} {
		c12ForbidPowerShellCommand(t, "controller WAL exact-byte path", nodes, forbidden)
	}
	output, exitCode := runC12CleanupStateHarness(t, "controller-wal")
	if exitCode != 0 {
		t.Fatalf("production controller-WAL verifier harness exit=%d output=%q", exitCode, output)
	}
}

func TestC12PITRDockerReceiptAndAbsenceAreExact(t *testing.T) {
	snapshot := c12PowerShellExecutableAST(t)
	nodes := c12ReachablePowerShellNodes(t, snapshot, "Invoke-C12Docker", "Remove-C12PITRBaseResources")
	c12RequirePowerShellASTTerms(t, "Docker receipt/exact absence", nodes, []string{
		"postgres:18.4-alpine3.23", "sha256:996d0920e4ff9df1fc19dacb904492f3c1ec0ec1cc338f0ad7123be7731c5f5e",
		"docker_endpoint_identity_digest", "talenro.c12.docker-endpoint.v1", "context_name", "endpoint", "engine_id",
		"server_version", "os_type", "architecture", "DOCKER_HOST", "DOCKER_CONTEXT", "DOCKER_CONFIG",
		"DOCKER_TLS_VERIFY", "DOCKER_CERT_PATH", "DOCKER_API_VERSION", "talenro.c12.docker-exact-absence.v1",
		"Error response from daemon: No such container: ", ": no such volume", "\\[\"container\",\"inspect\"",
		"\\[\"volume\",\"inspect\"", "5b5d0a",
	})
	c12ForbidPowerShellASTTerm(t, "Docker raw inspect", nodes, `--format\s+['\"]\{\{\.(?:Name|Id)\}\}`)
	output, exitCode := runC12CleanupStateHarness(t, "docker-exact-production")
	if exitCode != 0 {
		t.Fatalf("production Docker receipt/absence harness exit=%d output=%q", exitCode, output)
	}
}

type c12PowerShellASTNode struct {
	Kind      string   `json:"kind"`
	Parent    string   `json:"parent"`
	Text      string   `json:"text"`
	Command   string   `json:"command"`
	Arguments []string `json:"arguments"`
	Start     int      `json:"start"`
	Scope     string   `json:"-"`
	Path      string   `json:"-"`
	Ordinal   int      `json:"-"`
}

type c12PowerShellASTScope struct {
	Name  string                 `json:"name"`
	Nodes []c12PowerShellASTNode `json:"nodes"`
}

type c12PowerShellASTDocument struct {
	Scopes []c12PowerShellASTScope `json:"scopes"`
}

var (
	c12PowerShellASTOnce  sync.Once
	c12PowerShellASTCache map[string][]c12PowerShellASTNode
	c12PowerShellASTError error
)

// c12PowerShellExecutableAST delegates syntax classification to the Windows
// PowerShell parser. Only AST nodes inside one exact function/worker
// ScriptBlock are returned; comments, sibling functions and nested dead
// function declarations cannot satisfy a corrective assertion.
func c12PowerShellExecutableAST(t *testing.T) map[string][]c12PowerShellASTNode {
	t.Helper()
	c12PowerShellASTOnce.Do(func() {
		runner, err := filepath.Abs("../../scripts/run-c12-integration.ps1")
		if err != nil {
			c12PowerShellASTError = err
			return
		}
		c12PowerShellASTCache, c12PowerShellASTError = c12ParsePowerShellExecutableAST(t, runner)
		if c12PowerShellASTError == nil {
			c12PowerShellASTError = c12CheckPowerShellASTInertMarkerFixture(t)
		}
	})
	if c12PowerShellASTError != nil {
		t.Fatal(c12PowerShellASTError)
	}
	return c12PowerShellASTCache
}

func c12ParsePowerShellExecutableAST(t *testing.T, scriptPath string) (map[string][]c12PowerShellASTNode, error) {
	t.Helper()
	probe := filepath.Join(t.TempDir(), "c12-ast-probe.ps1")
	if err := os.WriteFile(probe, []byte(c12PowerShellASTProbe), 0o600); err != nil {
		return nil, err
	}
	powershellPath := filepath.Join(os.Getenv("SystemRoot"), "System32", "WindowsPowerShell", "v1.0", "powershell.exe")
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, powershellPath, "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", probe, "-ScriptPath", scriptPath)
	output, runErr := command.CombinedOutput()
	if ctx.Err() != nil {
		return nil, fmt.Errorf("PowerShell AST probe timeout: %w", ctx.Err())
	}
	if runErr != nil {
		return nil, fmt.Errorf("PowerShell AST probe: %w: %s", runErr, output)
	}
	var document c12PowerShellASTDocument
	if err := json.Unmarshal(output, &document); err != nil {
		return nil, fmt.Errorf("decode PowerShell AST probe: %w: %s", err, output)
	}
	result := make(map[string][]c12PowerShellASTNode, len(document.Scopes))
	for _, scope := range document.Scopes {
		if _, duplicate := result[scope.Name]; duplicate {
			return nil, fmt.Errorf("PowerShell AST scope %s is not unique", scope.Name)
		}
		sort.Slice(scope.Nodes, func(i, j int) bool { return scope.Nodes[i].Start < scope.Nodes[j].Start })
		for index := range scope.Nodes {
			scope.Nodes[index].Scope = scope.Name
			scope.Nodes[index].Ordinal = index
		}
		result[scope.Name] = scope.Nodes
	}
	return result, nil
}

func c12CheckPowerShellASTInertMarkerFixture(t *testing.T) error {
	t.Helper()
	const fixture = `function Invoke-C12ASTGenuine {
  Write-Output 'C12_GENUINE_MARKER'
  Write-Output 'C12_SEQUENCE_FIRST'
  Write-Output 'C12_SEQUENCE_SECOND'
}
function Invoke-C12ASTUncalledSibling {
  Write-Output 'C12_UNCALLED_SIBLING_MARKER'
}
function Invoke-C12ASTSplitFirst {
  Write-Output 'C12_SEQUENCE_FIRST'
}
function Invoke-C12ASTSplitSecond {
  Write-Output 'C12_SEQUENCE_SECOND'
}
function Invoke-C12ASTRoot {
  $assignment = 'C12_ASSIGNMENT_MARKER'
  $table = @{ value = 'C12_HASHTABLE_MARKER' }
  if ($false) { Write-Output 'C12_FALSE_BRANCH_MARKER' }
  Invoke-C12ASTGenuine
}
function Invoke-C12ASTSplitRoot {
  Invoke-C12ASTSplitFirst
  Invoke-C12ASTSplitSecond
}`
	path := filepath.Join(t.TempDir(), "c12-ast-inert-fixture.ps1")
	if err := os.WriteFile(path, []byte(fixture), 0o600); err != nil {
		return err
	}
	document, err := c12ParsePowerShellExecutableAST(t, path)
	if err != nil {
		return err
	}
	nodes, err := c12PowerShellCallPaths(document, "Invoke-C12ASTRoot")
	if err != nil {
		return err
	}
	evidence := c12PowerShellNodeEvidence(nodes)
	if !strings.Contains(evidence, "C12_GENUINE_MARKER") {
		return fmt.Errorf("PowerShell AST executable-path fixture lost the genuinely called marker: %s", evidence)
	}
	for _, inert := range []string{"C12_ASSIGNMENT_MARKER", "C12_HASHTABLE_MARKER", "C12_UNCALLED_SIBLING_MARKER", "C12_FALSE_BRANCH_MARKER"} {
		if strings.Contains(evidence, inert) {
			return fmt.Errorf("PowerShell AST executable-path fixture admitted inert marker %s: %s", inert, evidence)
		}
	}
	sequence := []*regexp.Regexp{regexp.MustCompile(`C12_SEQUENCE_FIRST`), regexp.MustCompile(`C12_SEQUENCE_SECOND`)}
	if !c12PowerShellAnyLeafPathHasSequence(nodes, sequence) {
		return fmt.Errorf("PowerShell AST executable-path fixture lost a genuine single-chain sequence: %s", evidence)
	}
	splitNodes, err := c12PowerShellCallPaths(document, "Invoke-C12ASTSplitRoot")
	if err != nil {
		return err
	}
	if c12PowerShellAnyLeafPathHasSequence(splitNodes, sequence) {
		return fmt.Errorf("PowerShell AST executable-path fixture merged two invoked sibling call chains: %s", c12PowerShellNodeEvidence(splitNodes))
	}
	return nil
}

func c12ReachablePowerShellNodes(t *testing.T, document map[string][]c12PowerShellASTNode, roots ...string) []c12PowerShellASTNode {
	t.Helper()
	var result []c12PowerShellASTNode
	for _, root := range roots {
		nodes, err := c12PowerShellCallPaths(document, root)
		if err != nil {
			t.Fatal(err)
		}
		result = append(result, nodes...)
	}
	return result
}

func c12PowerShellCallPaths(document map[string][]c12PowerShellASTNode, root string) ([]c12PowerShellASTNode, error) {
	type leafPath struct {
		name  string
		nodes []c12PowerShellASTNode
	}
	resolveCall := func(scope string, node c12PowerShellASTNode) string {
		if node.Kind != "CommandAst" || node.Command == "" {
			return ""
		}
		candidates := []string{scope + "::" + node.Command, node.Command}
		if separator := strings.LastIndex(scope, "::"); separator >= 0 {
			candidates = append([]string{scope[:separator] + "::" + node.Command}, candidates...)
		}
		for _, candidate := range candidates {
			if _, callable := document[candidate]; callable {
				return candidate
			}
		}
		return ""
	}
	var walk func(string, map[string]bool) ([]leafPath, error)
	walk = func(scope string, active map[string]bool) ([]leafPath, error) {
		nodes, exists := document[scope]
		if !exists {
			return nil, fmt.Errorf("production runner lacks executable AST scope %s", scope)
		}
		if active[scope] {
			return nil, nil
		}
		nextActive := make(map[string]bool, len(active)+1)
		for key, value := range active {
			nextActive[key] = value
		}
		nextActive[scope] = true
		var paths []leafPath
		for callIndex, node := range nodes {
			candidate := resolveCall(scope, node)
			if candidate == "" || nextActive[candidate] {
				continue
			}
			children, err := walk(candidate, nextActive)
			if err != nil {
				return nil, err
			}
			for _, child := range children {
				projection := make([]c12PowerShellASTNode, 0, len(nodes)+len(child.nodes))
				projection = append(projection, nodes[:callIndex+1]...)
				projection = append(projection, child.nodes...)
				projection = append(projection, nodes[callIndex+1:]...)
				paths = append(paths, leafPath{
					name:  fmt.Sprintf("%s>%s@%d", scope, child.name, node.Ordinal),
					nodes: projection,
				})
			}
		}
		if len(paths) == 0 {
			paths = append(paths, leafPath{name: scope, nodes: append([]c12PowerShellASTNode(nil), nodes...)})
		}
		return paths, nil
	}
	paths, err := walk(root, map[string]bool{})
	if err != nil {
		return nil, err
	}
	var result []c12PowerShellASTNode
	for pathIndex, path := range paths {
		pathID := fmt.Sprintf("%s#%d", path.name, pathIndex)
		for _, sourceNode := range path.nodes {
			node := sourceNode
			node.Path = pathID
			result = append(result, node)
		}
	}
	return result, nil
}

func c12PowerShellNodeEvidence(nodes []c12PowerShellASTNode) string {
	var evidence strings.Builder
	for _, node := range nodes {
		evidence.WriteString(node.Path)
		evidence.WriteByte('\x00')
		evidence.WriteString(node.Command)
		for _, argument := range node.Arguments {
			evidence.WriteByte('\x00')
			evidence.WriteString(argument)
		}
		evidence.WriteByte('\n')
	}
	return evidence.String()
}

func c12PowerShellNodeMatches(node c12PowerShellASTNode, expression *regexp.Regexp) bool {
	if expression.MatchString(node.Command) {
		return true
	}
	for _, argument := range node.Arguments {
		if expression.MatchString(argument) {
			return true
		}
	}
	return false
}

func c12RequirePowerShellASTTerms(t *testing.T, label string, nodes []c12PowerShellASTNode, patterns []string) {
	t.Helper()
	for _, pathNodes := range c12PowerShellLeafPaths(nodes) {
		complete := true
		for _, pattern := range patterns {
			expression := regexp.MustCompile(pattern)
			found := false
			for _, node := range pathNodes {
				if c12PowerShellNodeMatches(node, expression) {
					found = true
					break
				}
			}
			if !found {
				complete = false
				break
			}
		}
		if complete {
			return
		}
	}
	t.Errorf("%s has no single reachable executable call path containing every term %q", label, patterns)
}

func c12RequirePowerShellASTSequence(t *testing.T, label string, nodes []c12PowerShellASTNode, patterns []string) {
	t.Helper()
	expressions := make([]*regexp.Regexp, len(patterns))
	for index, pattern := range patterns {
		expressions[index] = regexp.MustCompile(pattern)
	}
	if c12PowerShellAnyLeafPathHasSequence(nodes, expressions) {
		return
	}
	t.Errorf("%s has no single reachable executable call path containing ordered edges %q", label, patterns)
}

func c12PowerShellAnyLeafPathHasSequence(nodes []c12PowerShellASTNode, expressions []*regexp.Regexp) bool {
	for _, pathNodes := range c12PowerShellLeafPaths(nodes) {
		last := -1
		complete := true
		for _, expression := range expressions {
			found := -1
			for index, node := range pathNodes {
				if index > last && c12PowerShellNodeMatches(node, expression) {
					found = index
					break
				}
			}
			if found < 0 {
				complete = false
				break
			}
			last = found
		}
		if complete {
			return true
		}
	}
	return false
}

func c12PowerShellLeafPaths(nodes []c12PowerShellASTNode) map[string][]c12PowerShellASTNode {
	paths := make(map[string][]c12PowerShellASTNode)
	for _, node := range nodes {
		paths[node.Path] = append(paths[node.Path], node)
	}
	return paths
}

func c12ForbidPowerShellCommand(t *testing.T, label string, nodes []c12PowerShellASTNode, command string) {
	t.Helper()
	for _, node := range nodes {
		if strings.EqualFold(node.Command, command) {
			t.Errorf("%s contains forbidden CommandAst %s", label, command)
		}
	}
}

func c12ForbidPowerShellASTTerm(t *testing.T, label string, nodes []c12PowerShellASTNode, pattern string) {
	t.Helper()
	expression := regexp.MustCompile(pattern)
	for _, node := range nodes {
		if c12PowerShellNodeMatches(node, expression) {
			t.Errorf("%s contains forbidden executable AST term %q in %s", label, pattern, node.Kind)
			return
		}
	}
}

func c12RequirePowerShellASTExactLiteralCount(t *testing.T, label string, nodes []c12PowerShellASTNode, literal string, want int) {
	t.Helper()
	for _, pathNodes := range c12PowerShellLeafPaths(nodes) {
		got := 0
		for _, node := range pathNodes {
			for _, argument := range node.Arguments {
				if argument == literal {
					got++
				}
			}
		}
		if got == want {
			return
		}
	}
	t.Errorf("%s has no single reachable executable call path with literal %q count %d", label, literal, want)
}

func c12AssertLiteralPreparedFrames(t *testing.T) {
	t.Helper()
	payloads := []string{
		`{"schema":"prepared/v1","type":"HELLO","run":"r","bootstrap_nonce":"b","worker_nonce":"w","worker_pid":41,"expected_controller_pid":42}`,
		`{"schema":"prepared/v1","type":"VERIFICATION_PROJECTION","role":"trusted-validator","run":"r","bootstrap_nonce":"b","worker_nonce":"w","gate_name":"g","formal_argv":["validator"],"input_count":0,"inputs":[],"terminal_marker":"PASS"}`,
		`{"schema":"prepared/v1","type":"JOB_MEMBER_READY","run":"r","worker_nonce":"w","projection_digest":"00","opened_inputs_digest":"11"}`,
		`{"schema":"prepared/v1","type":"CAPABILITIES","run":"r","capability_count":0,"capabilities":[],"capability_digest":"22"}`,
		`{"schema":"prepared/v1","type":"CAPABILITIES_INSTALLED","run":"r","capability_digest":"22"}`,
		`{"schema":"prepared/v1","type":"GATE_WAITING","run":"r","gate_name":"g","capability_digest":"22"}`,
	}
	prefixes := [][4]byte{{0x87, 0, 0, 0}, {0xe8, 0, 0, 0}, {0x84, 0, 0, 0}, {0x78, 0, 0, 0}, {0x5b, 0, 0, 0}, {0x61, 0, 0, 0}}
	var wire []byte
	for index, payload := range payloads {
		if !json.Valid([]byte(payload)) || strings.ContainsAny(payload, "\r\n") || bytes.HasPrefix([]byte(payload), []byte{0xef, 0xbb, 0xbf}) {
			t.Fatalf("literal prepared frame %d is not compact BOM-free UTF-8 JSON", index)
		}
		wire = append(wire, prefixes[index][:]...)
		wire = append(wire, payload...)
	}
	want, _ := hex.DecodeString("b29408e581e7e4f8d4ed41c67118dbd9cbcfa8fd2473cbb842893536f0025348")
	got := sha256.Sum256(wire)
	if !bytes.Equal(got[:], want) {
		t.Fatalf("literal six-frame wire SHA-256 = %x, want b29408e581e7e4f8d4ed41c67118dbd9cbcfa8fd2473cbb842893536f0025348", got)
	}
}

const c12LiteralControllerWALBootstrap = `{"schema":"talenro-c12-authority-pitr-controller-ownership-wal/v1","version":1,"run":"11111111111111111111111111111111","profile":"authority-v7-pitr","nonce_digest":"2222222222222222222222222222222222222222222222222222222222222222","docker_executable_digest":"3333333333333333333333333333333333333333333333333333333333333333","docker_endpoint_identity_digest":"4444444444444444444444444444444444444444444444444444444444444444","sequence":0,"previous_record_digest":null,"event":"BOOTSTRAP","timestamp_utc":"2026-08-29T17:00:00.0000000Z","payload":{"registry":[],"registry_digest":"f3dc0dc654e27a1376da941217e388da59821b25502a30532149c0a2e536221d"},"payload_digest":"c74747ab1c8a1c04fbb7ae900c465c7c29962094e8062dab6c7e9e4460bfe96f","record_digest":"dc132aaf295a65ca5e1d21c8854c54be80c836cd8b79b02129f86e29ddce3016","hmac_sha256":"827f50e451eeb685ab497f38e9968042e7249aabc64e4df60fb0467b2f4fb8a2"}` + "\n"

func c12AssertLiteralControllerWAL(t *testing.T) {
	t.Helper()
	line := c12LiteralControllerWALBootstrap
	if len(line) != 897 || line[len(line)-1] != '\n' || strings.Contains(line, "\r") || bytes.HasPrefix([]byte(line), []byte{0xef, 0xbb, 0xbf}) {
		t.Fatalf("literal BOOTSTRAP WAL line grammar/length mismatch: %d bytes", len(line))
	}
	const recordDigest = "dc132aaf295a65ca5e1d21c8854c54be80c836cd8b79b02129f86e29ddce3016"
	key := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31}
	rawDigest, err := hex.DecodeString(recordDigest)
	if err != nil {
		t.Fatal(err)
	}
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte("talenro.c12.controller-wal.hmac.v1"))
	mac.Write([]byte{0})
	mac.Write(rawDigest)
	if got := hex.EncodeToString(mac.Sum(nil)); got != "827f50e451eeb685ab497f38e9968042e7249aabc64e4df60fb0467b2f4fb8a2" {
		t.Fatalf("literal BOOTSTRAP HMAC = %s", got)
	}
}

const c12PowerShellASTProbe = `param([Parameter(Mandatory)][string]$ScriptPath)
$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
$tokens = $null
$parseErrors = $null
$ast = [System.Management.Automation.Language.Parser]::ParseFile($ScriptPath, [ref]$tokens, [ref]$parseErrors)
if ($parseErrors.Count -ne 0) { throw ('runner parse errors: ' + ($parseErrors -join '; ')) }

function Get-C12DirectNodes {
  param([System.Management.Automation.Language.Ast]$Body)
  $nested = @($Body.FindAll({ param($n) $n -is [System.Management.Automation.Language.FunctionDefinitionAst] }, $true))
  $nodes = @($Body.FindAll({
    param($n)
    $n -is [System.Management.Automation.Language.CommandAst] -or
    $n -is [System.Management.Automation.Language.InvokeMemberExpressionAst]
  }, $true))
  $result = @()
  foreach ($node in $nodes) {
    $insideNested = $false
    foreach ($definition in $nested) {
      if ($node.Extent.StartOffset -ge $definition.Extent.StartOffset -and $node.Extent.EndOffset -le $definition.Extent.EndOffset) { $insideNested = $true; break }
    }
    if ($insideNested) { continue }
    $insideLiteralFalseBranch = $false
    $ancestor = $node.Parent
    while ($null -ne $ancestor) {
      if ($ancestor -is [System.Management.Automation.Language.IfStatementAst]) {
        foreach ($clause in $ancestor.Clauses) {
          $condition = ([string]$clause.Item1.Extent.Text).Trim()
          if ($condition -match '^\(?\s*\$false\s*\)?$') { $insideLiteralFalseBranch = $true; break }
        }
      }
      if ($insideLiteralFalseBranch) { break }
      $ancestor = $ancestor.Parent
    }
    if ($insideLiteralFalseBranch) { continue }
    $command = ''
    $arguments = [Collections.Generic.List[string]]::new()
    if ($node -is [System.Management.Automation.Language.CommandAst]) {
      $command = [string]$node.GetCommandName()
      for ($index = 1; $index -lt $node.CommandElements.Count; $index++) {
        $element = $node.CommandElements[$index]
        if ($element -is [System.Management.Automation.Language.StringConstantExpressionAst] -or $element -is [System.Management.Automation.Language.ExpandableStringExpressionAst]) {
          $arguments.Add([string]$element.Value)
        } else {
          $arguments.Add([string]$element.Extent.Text)
        }
      }
    } else {
      if ($node.Member -is [System.Management.Automation.Language.StringConstantExpressionAst]) {
        $command = [string]$node.Member.Value
      } else {
        $command = [string]$node.Member.Extent.Text
      }
      foreach ($argument in $node.Arguments) {
        if ($argument -is [System.Management.Automation.Language.StringConstantExpressionAst] -or $argument -is [System.Management.Automation.Language.ExpandableStringExpressionAst]) {
          $arguments.Add([string]$argument.Value)
        } else {
          $arguments.Add([string]$argument.Extent.Text)
        }
      }
    }
    $result += [pscustomobject]@{
      kind = $node.GetType().Name
      parent = if ($null -eq $node.Parent) { '' } else { $node.Parent.GetType().Name }
      text = [string]$node.Extent.Text
      command = $command
      arguments = @($arguments)
      start = [int]$node.Extent.StartOffset
    }
  }
  return @($result | Sort-Object start,kind,text)
}

$scopes = @()
$definitions = @($ast.FindAll({ param($n) $n -is [System.Management.Automation.Language.FunctionDefinitionAst] }, $true))
foreach ($definition in $definitions) {
  $owners = [Collections.Generic.List[string]]::new()
  $parent = $definition.Parent
  while ($null -ne $parent) {
    if ($parent -is [System.Management.Automation.Language.FunctionDefinitionAst]) { $owners.Insert(0, [string]$parent.Name) }
    if ($parent -is [System.Management.Automation.Language.AssignmentStatementAst] -and $parent.Left.Extent.Text -ceq '$script:c12PreparedNativeWorkerScript') {
      $owners.Insert(0, '$script:c12PreparedNativeWorkerScript')
    }
    $parent = $parent.Parent
  }
  $owners.Add([string]$definition.Name)
  $scopes += [pscustomobject]@{ name = ($owners -join '::'); nodes = @(Get-C12DirectNodes -Body $definition.Body) }
}
$workerAssignments = @($ast.FindAll({
  param($n)
  $n -is [System.Management.Automation.Language.AssignmentStatementAst] -and
    $n.Left.Extent.Text -ceq '$script:c12PreparedNativeWorkerScript'
}, $true))
if ($workerAssignments.Count -eq 1) {
  $scopes += [pscustomobject]@{ name = '$script:c12PreparedNativeWorkerScript'; nodes = @(Get-C12DirectNodes -Body $workerAssignments[0].Right) }
}
[pscustomobject]@{ scopes = @($scopes) } | ConvertTo-Json -Depth 8 -Compress
`

func readC12RunnerSource(t *testing.T) []byte {
	t.Helper()
	raw, err := os.ReadFile("../../scripts/run-c12-integration.ps1")
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func c12PowerShellFunction(t *testing.T, source, name string) string {
	t.Helper()
	marker := "function " + name + " {"
	start := strings.Index(source, marker)
	if start < 0 {
		t.Fatalf("runner lacks PowerShell function %s", name)
	}
	rest := source[start+len(marker):]
	end := strings.Index(rest, "\nfunction ")
	if end < 0 {
		return source[start:]
	}
	return source[start : start+len(marker)+end]
}

func c12PowerShellAssignment(t *testing.T, source, marker string) string {
	t.Helper()
	start := strings.Index(source, marker)
	if start < 0 {
		t.Fatalf("runner lacks PowerShell assignment %s", marker)
	}
	rest := source[start+len(marker):]
	end := strings.Index(rest, "\nfunction ")
	if end < 0 {
		return source[start:]
	}
	return source[start : start+len(marker)+end]
}

func TestC12Task4AllowedPackagesStaySynchronizedWithRunner(t *testing.T) {
	raw, err := os.ReadFile("../../scripts/run-c12-integration.ps1")
	if err != nil {
		t.Fatal(err)
	}
	wantImportPaths := map[string]string{
		"./internal/testinfra":             "talenro.local/platform/internal/testinfra",
		"./internal/store":                 "talenro.local/platform/internal/store",
		"./internal/nodecontrol/contracts": "talenro.local/platform/internal/nodecontrol/contracts",
		"./internal/nodecontrol/authority": "talenro.local/platform/internal/nodecontrol/authority",
		"./internal/nodecontrol/serving":   "talenro.local/platform/internal/nodecontrol/serving",
		"./internal/readiness":             "talenro.local/platform/internal/readiness",
	}
	wantAuthorityLines := []string{
		`$script:c12AllowedPackages = [System.Collections.Generic.Dictionary[string,string]]::new([System.StringComparer]::Ordinal)`,
		`$script:c12AllowedPackages.Add('./internal/testinfra', 'talenro.local/platform/internal/testinfra')`,
		`$script:c12AllowedPackages.Add('./internal/store', 'talenro.local/platform/internal/store')`,
		`$script:c12AllowedPackages.Add('./internal/nodecontrol/contracts', 'talenro.local/platform/internal/nodecontrol/contracts')`,
		`$script:c12AllowedPackages.Add('./internal/nodecontrol/authority', 'talenro.local/platform/internal/nodecontrol/authority')`,
		`$script:c12AllowedPackages.Add('./internal/nodecontrol/serving', 'talenro.local/platform/internal/nodecontrol/serving')`,
		`$script:c12AllowedPackages.Add('./internal/readiness', 'talenro.local/platform/internal/readiness')`,
		`    if (-not $script:c12AllowedPackages.ContainsKey($package)) {`,
		`    Assert-C12GoJSONResult -Result $testResult -Package ([string]$script:c12AllowedPackages[$Package]) -ExpectedTests $ExpectedTests`,
		`    if (-not $script:c12AllowedPackages.ContainsKey($package)) {`,
	}
	var authorityLines []string
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSuffix(line, "\r")
		if strings.Contains(line, "$script:c12AllowedPackages") {
			authorityLines = append(authorityLines, line)
		}
	}
	if len(authorityLines) != len(wantAuthorityLines) {
		t.Fatalf("runner package authority line count = %d, want %d; lines=%q", len(authorityLines), len(wantAuthorityLines), authorityLines)
	}
	for index, want := range wantAuthorityLines {
		if authorityLines[index] != want {
			t.Fatalf("runner package authority line %d = %q, want %q", index+1, authorityLines[index], want)
		}
	}
	matches := regexp.MustCompile(`(?m)^\$script:c12AllowedPackages\.Add\('(\./[A-Za-z0-9_./-]+)', '(talenro\.local/platform/[A-Za-z0-9_./-]+)'\)$`).FindAllStringSubmatch(string(raw), -1)
	runnerPackages := make(map[string]string, len(matches))
	for _, match := range matches {
		runnerPackages[match[1]] = match[2]
	}
	if len(runnerPackages) != len(c12Task4AllowedPackages) {
		t.Fatalf("runner allowed package count = %d, validator count = %d", len(runnerPackages), len(c12Task4AllowedPackages))
	}
	for packageName := range c12Task4AllowedPackages {
		importPath, exists := runnerPackages[packageName]
		if !exists {
			t.Errorf("runner and validator allowed package sets differ at %s", packageName)
		} else if importPath != wantImportPaths[packageName] {
			t.Errorf("runner import path for %s = %q, want %q", packageName, importPath, wantImportPaths[packageName])
		}
	}
}

func TestC12BaseRunnerUsesPowerShellSafeDockerLabelTemplate(t *testing.T) {
	raw, err := os.ReadFile("../../scripts/run-c12-integration.ps1")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	const safeTemplate = "{{ index .Config.Labels `talenro.c12.run` }}"
	if got := strings.Count(source, safeTemplate); got != 2 {
		t.Fatalf("PowerShell-safe Docker label template count = %d, want 2 identity checks", got)
	}
	if strings.Contains(source, `{{index .Config.Labels "talenro.c12.run"}}`) {
		t.Fatal("Docker label template uses double quotes that Windows PowerShell strips before native invocation")
	}
	const safeRoleTemplate = "{{ index .Config.Labels `talenro.c12.role` }}"
	if got := strings.Count(source, safeRoleTemplate); got != 2 {
		t.Fatalf("PowerShell-safe Docker role-label template count = %d, want 2 identity checks", got)
	}
}

func TestC12BaseRunnerCapturesNativeFailuresBeforeCleanup(t *testing.T) {
	fakeBin := t.TempDir()
	fakeDocker := filepath.Join(fakeBin, "docker.exe")
	const fakeDockerSource = `package main
import("fmt";"os")
func main(){
 a:=os.Args[1:]
 if len(a)>=2&&a[0]=="image"&&a[1]=="inspect"{fmt.Println("sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa");return}
 if len(a)>=1&&a[0]=="run"{fmt.Fprintln(os.Stderr,"C12_NATIVE_STDERR_CANARY");os.Exit(73)}
 if len(a)>=2&&a[0]=="container"&&a[1]=="inspect"{fmt.Fprintln(os.Stderr,"Error: No such container");os.Exit(1)}
 fmt.Fprintln(os.Stderr,"unexpected docker invocation");os.Exit(74)
}`
	buildFakeGoExecutable(t, fakeDocker, fakeDockerSource)

	runner, err := os.ReadFile("../../scripts/run-c12-integration.ps1")
	if err != nil {
		t.Fatal(err)
	}
	marker := []byte("$script:c12RepositoryRoot = (Resolve-Path")
	index := bytes.Index(runner, marker)
	if index < 0 {
		t.Fatal("runner lacks main-program marker")
	}
	repositoryRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	repositoryPayload := base64.StdEncoding.EncodeToString([]byte(repositoryRoot))
	appendix := fmt.Sprintf(`
$script:c12RepositoryRoot = [Text.Encoding]::UTF8.GetString([Convert]::FromBase64String('%s'))
$script:c12SuiteDeadline = [DateTime]::UtcNow.AddMinutes(3)
try {
  Invoke-C12Group -GroupID 'base-native-failure-fixture' -GroupProfile 'base' -Package './internal/testinfra' -RunPattern '^TestC12DependenciesAreIsolatedAndBaseMigrated$' -GroupTimeout '3m' -ExpectedTests @('TestC12DependenciesAreIsolatedAndBaseMigrated')
  exit 0
}
catch { [Console]::Error.WriteLine($_.Exception.Message); exit 1 }
`, repositoryPayload)
	harnessRoot := t.TempDir()
	harness := filepath.Join(harnessRoot, "scripts", "run-c12-integration.ps1")
	if err := os.MkdirAll(filepath.Dir(harness), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(harness, append(append([]byte(nil), runner[:index]...), []byte(appendix)...), 0o600); err != nil {
		t.Fatal(err)
	}
	output, exitCode := runC12PowerShellAtRootWithTimeout(t, harnessRoot, 90*time.Second, map[string]string{
		"Path": fakeBin + string(os.PathListSeparator) + os.Getenv("Path"),
	}, "-Profile", "base", "-Packages", "./internal/testinfra", "-Timeout", "3m")
	if exitCode == 0 || !strings.Contains(output, "start postgres container failed with exit code 73") {
		t.Fatalf("exit=%d output=%q, want captured native exit failure", exitCode, output)
	}
	if strings.Contains(output, "C12_NATIVE_STDERR_CANARY") {
		t.Fatalf("runner leaked native stderr: %q", output)
	}
	if strings.Contains(output, "C12 cleanup failure") {
		t.Fatalf("runner treated expected absent unstarted resources as cleanup failures: %q", output)
	}
}

func TestC12BaseRunnerDoesNotAdoptSameNameContainerAfterFailedCreate(t *testing.T) {
	fakeBin := t.TempDir()
	invocationLog := filepath.Join(t.TempDir(), "docker.log")
	const fakeDockerSource = `package main
import("fmt";"os";"strings")
func main(){
 args:=os.Args[1:]; joined:=strings.Join(args," ")
 f,_:=os.OpenFile(os.Getenv("C12_DOCKER_LOG"),os.O_CREATE|os.O_APPEND|os.O_WRONLY,0600); fmt.Fprintln(f,joined); f.Close()
 if len(args)>1 && args[0]=="image" && args[1]=="inspect" { fmt.Println("sha256:"+strings.Repeat("a",64)); return }
 if args[0]=="run" { os.Exit(73) }
 if len(args)>1 && args[0]=="container" && args[1]=="inspect" { fmt.Println(strings.Repeat("f",64)+"|/attacker|attacker"); return }
 os.Exit(74)
}
`
	buildFakeGoExecutable(t, filepath.Join(fakeBin, "docker.exe"), fakeDockerSource)
	output, exitCode := runC12PowerShell(t, map[string]string{
		"C12_DOCKER_LOG": invocationLog,
		"Path":           fakeBin + string(os.PathListSeparator) + os.Getenv("Path"),
	}, "-Profile", "base", "-Packages", "./internal/testinfra", "-Run", "^TestC12DependenciesAreIsolatedAndBaseMigrated$", "-Timeout", "3m")
	if exitCode == 0 {
		t.Fatalf("exit=%d output=%q, want failed create", exitCode, output)
	}
	logged, err := os.ReadFile(invocationLog)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(logged), "container inspect") || strings.Contains(string(logged), "container stop") || strings.Contains(string(logged), "container rm") {
		t.Fatalf("failed create caused name-based adoption or cleanup: %q", logged)
	}
}

func TestC12BaseRunnerRejectsMutableDockerIdentityDimensions(t *testing.T) {
	const fakeDockerSource = `package main
import("fmt";"os";"strings")
var ids=map[string]string{"postgres":strings.Repeat("1",64),"redis":strings.Repeat("2",64),"nats":strings.Repeat("3",64)}
var refs=map[string]string{"postgres":"postgres:18.4-alpine3.23","redis":"redis:8.8.1-alpine3.23","nats":"nats:2.14.3-alpine3.22"}
var imageIDs=map[string]string{"postgres":"sha256:"+strings.Repeat("a",64),"redis":"sha256:"+strings.Repeat("b",64),"nats":"sha256:"+strings.Repeat("c",64)}
func roleOf(v string)string{for _,r:=range []string{"postgres","redis","nats"}{if strings.Contains(v,r){return r}};if len(v)>0{switch v[0]{case '1':return "postgres";case '2':return "redis";case '3':return "nats"}};return ""}
func main(){
 args:=os.Args[1:]; joined:=strings.Join(args," "); last:=args[len(args)-1]
 if len(args)>1&&args[0]=="image"&&args[1]=="inspect"{fmt.Println(imageIDs[roleOf(last)]);return}
 if args[0]=="run"{role:=roleOf(joined);name:="";for i:=range args{if args[i]=="--name"&&i+1<len(args){name=args[i+1]}};os.WriteFile(os.Getenv("C12_NAME_STATE"),[]byte(name),0600);fmt.Println(ids[role]);return}
 if len(args)>1&&args[0]=="container"&&args[1]=="inspect"{
  nameRaw,_:=os.ReadFile(os.Getenv("C12_NAME_STATE"));name:=string(nameRaw);role:=roleOf(name);suffix:=strings.TrimSuffix(strings.TrimPrefix(name,"talenro-c12-"),"-"+role);mode:=os.Getenv("C12_IDENTITY_MODE");reportedRole:=role;reportedRef:=refs[role];reportedImageID:=imageIDs[role]
  if mode=="role"{reportedRole="attacker"};if mode=="reference"{reportedRef="attacker:latest"};if mode=="image-id"{reportedImageID="sha256:"+strings.Repeat("f",64)}
  fmt.Printf("%s|/%s|%s|%s|%s|%s\n",ids[role],name,suffix,reportedRole,reportedRef,reportedImageID);return
 }
 os.Exit(74)
}
`
	for _, test := range []struct{ mode, want string }{{"role", "role mismatch"}, {"reference", "image reference mismatch"}, {"image-id", "immutable image ID mismatch"}} {
		t.Run(test.mode, func(t *testing.T) {
			fakeBin := t.TempDir()
			buildFakeGoExecutable(t, filepath.Join(fakeBin, "docker.exe"), fakeDockerSource)
			output, exitCode := runC12PowerShell(t, map[string]string{
				"C12_IDENTITY_MODE": test.mode,
				"C12_NAME_STATE":    filepath.Join(t.TempDir(), "name"),
				"Path":              fakeBin + string(os.PathListSeparator) + os.Getenv("Path"),
			}, "-Profile", "base", "-Packages", "./internal/testinfra", "-Run", "^TestC12DependenciesAreIsolatedAndBaseMigrated$", "-Timeout", "3m")
			if exitCode == 0 || !strings.Contains(output, test.want) {
				t.Fatalf("exit=%d output=%q, want %q", exitCode, output, test.want)
			}
		})
	}
}

func TestC12BaseRunnerWatchdogKillsNativeDescendantsAndContinuesCleanup(t *testing.T) {
	fakeBin := t.TempDir()
	childPID := filepath.Join(t.TempDir(), "child.pid")
	delayedSentinel := filepath.Join(t.TempDir(), "delayed.txt")
	cleanupLog := filepath.Join(t.TempDir(), "cleanup.log")
	const fakeDockerSource = `package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

var ids = map[string]string{
	"postgres": strings.Repeat("1", 64),
	"redis": strings.Repeat("2", 64),
	"nats": strings.Repeat("3", 64),
}

func roleOf(value string) string {
	for _, role := range []string{"postgres", "redis", "nats"} {
		if strings.Contains(value, role) { return role }
	}
	if len(value) > 0 { switch value[0] { case '1': return "postgres"; case '2': return "redis"; case '3': return "nats" } }
	return ""
}

func main() {
	args := os.Args[1:]
	joined := strings.Join(args, " ")
	last := args[len(args)-1]
	if len(args) >= 2 && args[0] == "image" && args[1] == "inspect" {
		imageIDs := map[string]string{"postgres":"sha256:"+strings.Repeat("a",64),"redis":"sha256:"+strings.Repeat("b",64),"nats":"sha256:"+strings.Repeat("c",64)}
		fmt.Println(imageIDs[roleOf(last)])
		return
	}
	if args[0] == "run" {
		role := roleOf(joined)
		name := ""
		for index := range args { if args[index] == "--name" && index+1 < len(args) { name = args[index+1] } }
		os.WriteFile(os.Getenv("C12_NAME_"+strings.ToUpper(role)), []byte(name), 0600)
		fmt.Println(ids[role])
		return
	}
	if len(args) >= 2 && args[0] == "container" && args[1] == "inspect" {
		if len(args) >= 4 && args[3] == "{{.Id}}" { os.Exit(1) }
		role := roleOf(last)
		nameRaw, _ := os.ReadFile(os.Getenv("C12_NAME_"+strings.ToUpper(role)))
		name := string(nameRaw)
		suffix := strings.TrimSuffix(strings.TrimPrefix(name, "talenro-c12-"), "-"+role)
		if strings.Contains(joined, "talenro.c12.role") {
			refs := map[string]string{"postgres":"postgres:18.4-alpine3.23","redis":"redis:8.8.1-alpine3.23","nats":"nats:2.14.3-alpine3.22"}
			imageIDs := map[string]string{"postgres":"sha256:"+strings.Repeat("a",64),"redis":"sha256:"+strings.Repeat("b",64),"nats":"sha256:"+strings.Repeat("c",64)}
			fmt.Printf("%s|/%s|%s|%s|%s|%s\n", ids[role], name, suffix, role, refs[role], imageIDs[role])
		} else {
			fmt.Printf("%s|/%s|%s\n", ids[role], name, suffix)
		}
		return
	}
	if len(args) >= 2 && args[0] == "container" && args[1] == "port" {
		containerID := args[len(args)-2]
		switch containerID[0] { case '1': fmt.Println("127.0.0.1:15432"); case '2': fmt.Println("127.0.0.1:16379"); case '3': fmt.Println("malformed-nats-port") }
		return
	}
	if len(args) >= 2 && args[0] == "container" && args[1] == "stop" {
		if last[0] == '3' {
			child := exec.Command("powershell.exe", "-NoProfile", "-Command", "[IO.File]::WriteAllText($env:C12_CHILD_PID,[string]$PID); Start-Sleep -Seconds 6; [IO.File]::WriteAllText($env:C12_DELAYED_SENTINEL,'late')")
			child.Env = os.Environ()
			if err := child.Start(); err != nil { os.Exit(75) }
			time.Sleep(8*time.Second)
			return
		}
		file, err := os.OpenFile(os.Getenv("C12_CLEANUP_LOG"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil { os.Exit(76) }
		fmt.Fprintln(file, last)
		file.Close()
		return
	}
	if len(args) >= 2 && args[0] == "container" && args[1] == "rm" { return }
	os.Exit(74)
}
`
	buildFakeGoExecutable(t, filepath.Join(fakeBin, "docker.exe"), fakeDockerSource)

	// Exercise the production native watchdog and exact cleanup directly. Formal
	// prepared-validator release belongs to the separate authenticated-pipe tests.
	runner, err := os.ReadFile("../../scripts/run-c12-integration.ps1")
	if err != nil {
		t.Fatal(err)
	}
	mainIndex := bytes.Index(runner, []byte("$script:c12RepositoryRoot = (Resolve-Path"))
	if mainIndex < 0 {
		t.Fatal("runner lacks main-program marker")
	}
	const cleanupAppendix = `
$script:c12RepositoryRoot = (Get-Location).Path
$runSuffix = '11111111111111111111111111111111'
$failures = @()
foreach ($kind in @('nats', 'redis', 'postgres')) {
  $identity = switch ($kind) { 'nats' { '3' }; 'redis' { '2' }; 'postgres' { '1' } }
  $image = switch ($kind) { 'nats' { 'nats:2.14.3-alpine3.22' }; 'redis' { 'redis:8.8.1-alpine3.23' }; 'postgres' { 'postgres:18.4-alpine3.23' } }
  $imageByte = switch ($kind) { 'nats' { 'c' }; 'redis' { 'b' }; 'postgres' { 'a' } }
  $resource = [pscustomobject]@{ Kind = $kind; ID = ($identity * 64); Name = "talenro-c12-$runSuffix-$kind"; ImageRef = $image; ImageID = ('sha256:' + ($imageByte * 64)) }
  [IO.File]::WriteAllText([Environment]::GetEnvironmentVariable('C12_NAME_' + $kind.ToUpperInvariant()), $resource.Name)
  try { Remove-C12Container -Resource $resource -RunSuffix $runSuffix -Deadline ([DateTime]::UtcNow.AddSeconds(15)) }
  catch { $failures += $_.Exception.Message; [Console]::Error.WriteLine($_.Exception.Message) }
}
if ($failures.Count -eq 1 -and $failures[0] -match 'stop exact nats container timed out') { Write-Output 'C12_NATIVE_CLEANUP_WATCHDOG_OK' }
exit 1
`
	harnessRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(harnessRoot, "scripts"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(harnessRoot, "scripts", "run-c12-integration.ps1"), append(append([]byte(nil), runner[:mainIndex]...), []byte(cleanupAppendix)...), 0o600); err != nil {
		t.Fatal(err)
	}
	started := time.Now()
	output, exitCode := runC12PowerShellAtRootWithTimeout(t, harnessRoot, 30*time.Second, map[string]string{
		"C12_CHILD_PID":        childPID,
		"C12_CLEANUP_LOG":      cleanupLog,
		"C12_DELAYED_SENTINEL": delayedSentinel,
		"C12_NAME_POSTGRES":    filepath.Join(t.TempDir(), "postgres.name"),
		"C12_NAME_REDIS":       filepath.Join(t.TempDir(), "redis.name"),
		"C12_NAME_NATS":        filepath.Join(t.TempDir(), "nats.name"),
		"Path":                 fakeBin + string(os.PathListSeparator) + os.Getenv("Path"),
	}, "-Profile", "base", "-Packages", "./internal/testinfra", "-Run", "^TestC12DependenciesAreIsolatedAndBaseMigrated$", "-Timeout", "3m")
	elapsed := time.Since(started)
	if exitCode == 0 || elapsed > 25*time.Second {
		t.Fatalf("exit=%d elapsed=%s output=%q, want watchdog failure within twenty-five seconds", exitCode, elapsed, output)
	}
	if !strings.Contains(output, "C12_NATIVE_CLEANUP_WATCHDOG_OK") {
		t.Fatalf("production cleanup did not preserve exactly the expected native timeout: %q", output)
	}
	deadline := time.Now().Add(7 * time.Second)
	for time.Now().Before(deadline) && !fileExists(childPID) {
		time.Sleep(25 * time.Millisecond)
	}
	pidRaw, err := os.ReadFile(childPID)
	if err != nil {
		t.Fatalf("hanging native descendant did not record its exact PID: %v; output=%q", err, output)
	}
	remaining := time.Until(deadline)
	if remaining > 0 {
		time.Sleep(remaining)
	}
	if fileExists(delayedSentinel) {
		t.Fatalf("timed-out native descendant survived long enough to write its delayed sentinel; output=%q elapsed=%s", output, elapsed)
	}
	pid := strings.TrimSpace(string(pidRaw))
	check := exec.Command("powershell", "-NoProfile", "-Command", "Get-Process -Id "+pid+" -ErrorAction Stop | Out-Null")
	if err := check.Run(); err == nil {
		t.Fatalf("timed-out native descendant PID %s still exists", pid)
	}
	cleanup, err := os.ReadFile(cleanupLog)
	if err != nil {
		t.Fatalf("cleanup did not continue after one timed-out resource: %v", err)
	}
	if !strings.Contains(string(cleanup), strings.Repeat("2", 64)) || !strings.Contains(string(cleanup), strings.Repeat("1", 64)) {
		t.Fatalf("cleanup continuation log = %q, want later exact redis and postgres IDs; output=%q", cleanup, output)
	}
}

func TestC12BaseRunnerUsesFreshBoundedDeadlineForExactContainerCleanup(t *testing.T) {
	output, exitCode := runC12ExpiredGroupCleanupHarness(t)
	if exitCode != 0 {
		t.Fatalf("expired-group cleanup harness exit=%d output=%q", exitCode, output)
	}
}

func TestC12PreparedArtifactCleanupRetainsOwnershipAfterDeadline(t *testing.T) {
	output, exitCode := runC12CleanupStateHarness(t, "prepared-deadline")
	if exitCode != 0 {
		t.Fatalf("prepared artifact cleanup retry harness exit=%d output=%q", exitCode, output)
	}
}

func TestC12PITRGroupCleanupSkipsCandidatesBeforeRunRootAcquisition(t *testing.T) {
	output, exitCode := runC12CleanupStateHarness(t, "group-before-run-root")
	if exitCode != 0 {
		t.Fatalf("pre-run-root PITR cleanup harness exit=%d output=%q", exitCode, output)
	}
	if strings.Contains(output, "C12 cleanup failure [cleanup-before-run-root]: PITR candidates:") {
		t.Fatalf("pre-run-root failure incorrectly entered PITR candidate cleanup: %q", output)
	}
}

func TestC12PITRBaseCleanupSkipsNeverAttemptedVolumes(t *testing.T) {
	output, exitCode := runC12CleanupStateHarness(t, "never-attempted-volume")
	if exitCode != 0 {
		t.Fatalf("never-attempted PITR volume cleanup harness exit=%d output=%q", exitCode, output)
	}
}

func TestC12PITRBaseCleanupReinspectsAttemptedUnconfirmedVolume(t *testing.T) {
	output, exitCode := runC12CleanupStateHarness(t, "attempted-unconfirmed-volume")
	if exitCode != 0 {
		t.Fatalf("attempted-unconfirmed PITR volume cleanup harness exit=%d output=%q", exitCode, output)
	}
}

func TestC12BaseRunnerBoundsDependencyProbesToOneDeadline(t *testing.T) {
	redisListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer redisListener.Close()
	natsListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer natsListener.Close()

	release := make(chan struct{})
	defer close(release)
	redisAccepted := make(chan struct{}, 1)
	natsAccepted := make(chan struct{}, 1)
	serveStalledProbe := func(listener net.Listener, accepted chan<- struct{}) {
		connection, acceptErr := listener.Accept()
		if acceptErr != nil {
			return
		}
		defer connection.Close()
		if accepted != nil {
			accepted <- struct{}{}
		}
		_ = connection.SetReadDeadline(time.Now().Add(5 * time.Second))
		buffer := make([]byte, 4096)
		_, _ = connection.Read(buffer)
		<-release
	}
	go serveStalledProbe(redisListener, redisAccepted)
	go serveStalledProbe(natsListener, natsAccepted)

	output, exitCode := runC12DependencyDeadlineHarness(
		t,
		redisListener.Addr().(*net.TCPAddr).Port,
		natsListener.Addr().(*net.TCPAddr).Port,
	)
	if exitCode != 0 {
		t.Fatalf("dependency deadline harness exit=%d output=%q", exitCode, output)
	}
	select {
	case <-redisAccepted:
	default:
		t.Fatalf("dependency deadline harness never reached the stalled Redis protocol probe: %q", output)
	}
	select {
	case <-natsAccepted:
		t.Fatalf("dependency deadline harness granted NATS a new wait window after Redis exhausted the shared deadline: %q", output)
	default:
	}
}

func TestC12BaseRunnerCleanupBindsExactOwnedDirectoryIdentity(t *testing.T) {
	output, exitCode := runC12OwnedDirectoryIdentityHarness(t)
	if exitCode != 0 {
		t.Fatalf("owned directory identity harness exit=%d output=%q", exitCode, output)
	}
}

func TestC12BaseRunnerCreatesOwnedDirectoryFromAtomicNativeHandle(t *testing.T) {
	output, exitCode := runC12AtomicOwnedDirectoryCreationHarness(t)
	if exitCode != 0 {
		t.Fatalf("atomic owned-directory creation harness exit=%d output=%q", exitCode, output)
	}
}

func TestC12BaseRunnerContainedChildDoesNotSpawnWhenNamedJobContainmentFails(t *testing.T) {
	for _, mode := range []string{"open", "assignment"} {
		t.Run(mode, func(t *testing.T) {
			sentinel := filepath.Join(t.TempDir(), "native-spawned")
			output, exitCode := runC12ContainmentFailureHarness(t, mode, sentinel)
			if exitCode != 0 {
				t.Fatalf("containment failure harness exit=%d output=%q", exitCode, output)
			}
			time.Sleep(500 * time.Millisecond)
			if fileExists(sentinel) {
				t.Fatalf("native executable spawned after named Job %s failure", mode)
			}
		})
	}
}

func TestC12BaseRunnerContainedChildJoinsNamedJobBeforeAnyCompilerOrTargetProcess(t *testing.T) {
	sentinel := filepath.Join(t.TempDir(), "native-spawned")
	output, exitCode := runC12PreMembershipProcessHarness(t, sentinel)
	if exitCode != 0 {
		t.Fatalf("pre-membership process harness exit=%d output=%q", exitCode, output)
	}
	if fileExists(sentinel) {
		t.Fatalf("target process ran despite the test-only active-process limit: %q", output)
	}
}

func TestC12BaseRunnerRejectsNamedJobCollision(t *testing.T) {
	output, exitCode := runC12NamedJobCollisionHarness(t)
	if exitCode != 0 {
		t.Fatalf("named Job collision harness exit=%d output=%q", exitCode, output)
	}
}

func TestC12BaseRunnerDeclaresExactGroupAndSuiteDeadlines(t *testing.T) {
	raw, err := os.ReadFile("../../scripts/run-c12-integration.ps1")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	for _, literal := range []string{
		"$script:c12SuiteDeadline = [DateTime]::UtcNow.AddMinutes(120)",
		"[DateTime]::UtcNow.Add($groupDuration).AddMinutes(3)",
		"Invoke-C12Native",
		"Wait-Job -Job $job -Timeout",
		"JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE",
		"CreateJobObject(IntPtr.Zero, name)",
		"$typeBuilder.DefinePInvokeMethod(",
		"'OpenJobObject', 'kernel32.dll'",
		"'GetCurrentProcess', 'kernel32.dll'",
		"$assignMethod.Invoke",
	} {
		if !strings.Contains(source, literal) {
			t.Errorf("runner lacks enforced deadline literal %q", literal)
		}
	}
	for _, forbidden := range []string{"host.pid", "PIDTemporaryPath", "ReleasePath", "$releasePath", "$jobHostID", "C12NativeJob]::Assign", "c12NativeJobMemberSource", "ClientSource"} {
		if strings.Contains(source, forbidden) {
			t.Errorf("runner retains legacy filesystem/PID containment artifact %q", forbidden)
		}
	}
}

func TestC12BaseRunnerEnforcesBoundedCanonicalSubtests(t *testing.T) {
	const packageName = "talenro.local/platform/internal/store"
	parent := "TestParent"
	base := []string{
		c12GoJSONEvent(t, "run", packageName, parent),
		c12GoJSONEvent(t, "pass", packageName, parent),
		c12GoJSONEvent(t, "pass", packageName, ""),
	}
	manyDescendants := []string{c12GoJSONEvent(t, "run", packageName, parent)}
	for index := 0; index < 257; index++ {
		manyDescendants = append(manyDescendants, c12GoJSONEvent(t, "pass", packageName, fmt.Sprintf("%s/child-%03d", parent, index)))
	}
	manyDescendants = append(manyDescendants, c12GoJSONEvent(t, "pass", packageName, parent), c12GoJSONEvent(t, "pass", packageName, ""))
	tests := []struct {
		name    string
		events  []string
		wantErr string
	}{
		{name: "valid descendant", events: []string{
			c12GoJSONEvent(t, "run", packageName, parent),
			c12GoJSONEvent(t, "run", packageName, parent+"/valid-1"),
			c12GoJSONEvent(t, "pass", packageName, parent+"/valid-1"),
			c12GoJSONEvent(t, "pass", packageName, parent),
			c12GoJSONEvent(t, "pass", packageName, ""),
		}},
		{name: "orphan descendant", events: append([]string{c12GoJSONEvent(t, "pass", packageName, "TestOther/orphan")}, base...), wantErr: "orphan descendant"},
		{name: "malformed suffix", events: append([]string{c12GoJSONEvent(t, "pass", packageName, parent+"/bad space")}, base...), wantErr: "malformed"},
		{name: "oversized suffix segment", events: append([]string{c12GoJSONEvent(t, "pass", packageName, parent+"/"+strings.Repeat("a", 65))}, base...), wantErr: "malformed suffix"},
		{name: "excessive depth", events: append([]string{c12GoJSONEvent(t, "pass", packageName, parent+"/a/b/c/d/e")}, base...), wantErr: "depth"},
		{name: "excessive descendant count", events: manyDescendants, wantErr: "bounded descendant count"},
		{name: "duplicate terminal pass", events: append([]string{
			c12GoJSONEvent(t, "pass", packageName, parent+"/child"),
			c12GoJSONEvent(t, "pass", packageName, parent+"/child"),
		}, base...), wantErr: "duplicate terminal pass"},
		{name: "missing terminal pass", events: append([]string{c12GoJSONEvent(t, "run", packageName, parent+"/child")}, base...), wantErr: "missing terminal pass"},
		{name: "skipped descendant", events: append([]string{c12GoJSONEvent(t, "skip", packageName, parent+"/child")}, base...), wantErr: "skipped"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output, exitCode := runC12GoJSONAssertionHarness(t, packageName, []string{parent}, test.events)
			if test.wantErr == "" {
				if exitCode != 0 {
					t.Fatalf("exit=%d output=%q, want valid descendant acceptance", exitCode, output)
				}
				return
			}
			if exitCode == 0 || !strings.Contains(output, test.wantErr) {
				t.Fatalf("exit=%d output=%q, want rejection containing %q", exitCode, output, test.wantErr)
			}
		})
	}
}

func TestC12BaseRunnerRejectsUntrustedInputsBeforeDocker(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{name: "recursive package", args: []string{"-Profile", "base", "-Packages", "./...", "-Run", "^TestAlpha$", "-Timeout", "3m"}, wantErr: "invalid -Packages"},
		{name: "empty package segment", args: []string{"-Profile", "base", "-Packages", "./internal/store||./internal/testinfra", "-Run", "^TestAlpha$", "-Timeout", "3m"}, wantErr: "invalid -Packages"},
		{name: "duplicate package", args: []string{"-Profile", "base", "-Packages", "./internal/store|./internal/store", "-Run", "^TestAlpha$", "-Timeout", "3m"}, wantErr: "duplicate package"},
		{name: "package metacharacter", args: []string{"-Profile", "base", "-Packages", "./internal/store;whoami", "-Run", "^TestAlpha$", "-Timeout", "3m"}, wantErr: "invalid -Packages"},
		{name: "unknown package", args: []string{"-Profile", "base", "-Packages", "./cmd/control-api", "-Run", "^TestAlpha$", "-Timeout", "3m"}, wantErr: "unknown package"},
		{name: "future profile", args: []string{"-Profile", "authority-v8", "-Packages", "./internal/store", "-Run", "^TestAlpha$", "-Timeout", "3m"}, wantErr: "unsupported closed profile"},
		{name: "unanchored run", args: []string{"-Profile", "base", "-Packages", "./internal/store", "-Run", "TestAlpha", "-Timeout", "3m"}, wantErr: "anchored"},
		{name: "oversized focused timeout", args: []string{"-Profile", "base", "-Packages", "./internal/store", "-Run", "^TestAlpha$", "-Timeout", "31m"}, wantErr: "30m"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output, exitCode := runC12PowerShell(t, nil, test.args...)
			if exitCode == 0 || !strings.Contains(output, test.wantErr) {
				t.Fatalf("exit=%d output=%q, want failure containing %q", exitCode, output, test.wantErr)
			}
		})
	}
}

func TestC12BaseRunnerRejectsCaseVariantAllowedPackageBeforeGroup(t *testing.T) {
	const packageName = "./INTERNAL/STORE"
	output, exitCode := runC12PackageMembershipHarness(t, packageName)
	if exitCode == 0 || !strings.Contains(output, "unknown package "+packageName) {
		t.Fatalf("exit=%d output=%q, want exact-case unknown package rejection", exitCode, output)
	}
	if strings.Contains(output, "C12_GROUP_CALLED") {
		t.Fatalf("case-variant package reached the group boundary: %q", output)
	}
}

func TestC12BaseRunnerRejectsInheritedDependenciesWithoutEchoingThem(t *testing.T) {
	const canary = "postgres://secret-canary@127.0.0.1:5432/userdb"
	output, exitCode := runC12PowerShell(t, map[string]string{"TALENRO_DATABASE_URL": canary},
		"-Profile", "base", "-Packages", "./internal/store", "-Run", "^TestAlpha$", "-Timeout", "3m")
	if exitCode == 0 || !strings.Contains(output, "inherited dependency") {
		t.Fatalf("exit=%d output=%q, want inherited dependency rejection", exitCode, output)
	}
	if strings.Contains(output, canary) || strings.Contains(output, "secret-canary") {
		t.Fatalf("runner leaked inherited dependency canary: %q", output)
	}
}

func TestC12BaseRunnerRejectsEveryInheritedGitVariableWithoutEchoingIt(t *testing.T) {
	tests := []struct {
		name, variable, value string
	}{
		{name: "directory", variable: "GIT_DIR", value: `C:\secret-git-dir`},
		{name: "work tree", variable: "GIT_WORK_TREE", value: `C:\secret-work-tree`},
		{name: "common directory", variable: "GIT_COMMON_DIR", value: `C:\secret-common-dir`},
		{name: "lowercase namespace", variable: "git_namespace", value: "secret-namespace"},
		{name: "config count", variable: "GIT_CONFIG_COUNT", value: "1"},
		{name: "config key", variable: "GIT_CONFIG_KEY_0", value: "secret.key"},
		{name: "config value", variable: "GIT_CONFIG_VALUE_0", value: "secret-value"},
		{name: "index", variable: "GIT_INDEX_FILE", value: `C:\secret-index`},
		{name: "object directory", variable: "GIT_OBJECT_DIRECTORY", value: `C:\secret-objects`},
		{name: "alternate objects", variable: "GIT_ALTERNATE_OBJECT_DIRECTORIES", value: `C:\secret-alternate`},
		{name: "empty value", variable: "GIT_CEILING_DIRECTORIES", value: ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output, exitCode := runC12PowerShell(t, map[string]string{test.variable: test.value},
				"-Profile", "base", "-Packages", "./...", "-Run", "^TestAlpha$", "-Timeout", "3m")
			if exitCode == 0 || !strings.Contains(output, "inherited GIT_* variable is forbidden") {
				t.Fatalf("exit=%d output=%q, want inherited GIT_* rejection for %s", exitCode, output, test.variable)
			}
			if test.value != "" && strings.Contains(output, test.value) {
				t.Fatalf("runner leaked inherited %s canary: %q", test.variable, output)
			}
		})
	}
}

func TestC12Batch01SuiteFailsClosedWhileCanonicalManifestIsAbsent(t *testing.T) {
	repository := t.TempDir()
	runner, err := os.ReadFile("../../scripts/run-c12-integration.ps1")
	if err != nil {
		t.Fatal(err)
	}
	runnerPath := filepath.Join(repository, "scripts", "run-c12-integration.ps1")
	if err := os.MkdirAll(filepath.Dir(runnerPath), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(runnerPath, runner, 0o600); err != nil {
		t.Fatal(err)
	}
	for _, arguments := range [][]string{{"init", "--quiet"}, {"add", "--", "scripts/run-c12-integration.ps1"}} {
		command := exec.Command("git", arguments...)
		command.Dir = repository
		if output, commandErr := command.CombinedOutput(); commandErr != nil {
			t.Fatalf("git %v: %v\n%s", arguments, commandErr, output)
		}
	}

	output, exitCode := runC12PowerShellAtRootWithTimeout(t, repository, 30*time.Second, nil, "-Suite", "batch01", "-Timeout", "120m")
	if exitCode == 0 || !strings.Contains(output, "canonical Batch 01 manifest is absent") {
		t.Fatalf("exit=%d output=%q, want absent canonical manifest rejection", exitCode, output)
	}
}

func TestC12Batch01SuiteUsesFreshExactSnapshotPerGroup(t *testing.T) {
	run := runC12SnapshotSuiteFixture(t, "")
	if run.exitCode != 0 {
		t.Fatalf("exit=%d output=%q, want two groups from fresh exact snapshots", run.exitCode, run.output)
	}
	entries := readC12SnapshotGoLog(t, run.goLog)
	wantStages := map[string]bool{"validator": false, "./internal/testinfra": false, "./internal/store": false}
	roots := make(map[string]struct{})
	for _, entry := range entries {
		if _, wanted := wantStages[entry.Stage]; !wanted {
			continue
		}
		wantStages[entry.Stage] = true
		if len(entry.GitEnvironment) != 0 {
			t.Fatalf("%s child inherited candidate Git variables: %v", entry.Stage, entry.GitEnvironment)
		}
		root := filepath.Clean(entry.WorkingDirectory)
		if strings.HasPrefix(strings.ToLower(root), strings.ToLower(filepath.Clean(run.repository)+string(os.PathSeparator))) {
			t.Fatalf("%s ran from mutable fixture repository %q", entry.Stage, root)
		}
		roots[strings.ToLower(root)] = struct{}{}
		if entry.Stage == "validator" && !strings.Contains(entry.Arguments, "-candidate-tree") {
			t.Fatalf("validator did not receive captured tree identity: %q", entry.Arguments)
		}
	}
	for stage, observed := range wantStages {
		if !observed {
			t.Fatalf("missing fake Go execution stage %s; entries=%+v", stage, entries)
		}
	}
	if len(roots) != 3 {
		t.Fatalf("validator and two groups used %d distinct snapshot roots, want 3; entries=%+v", len(roots), entries)
	}
	if objectsAfter := relativeFileSet(t, run.objectsRoot); !equalStringSets(run.objectsBefore, objectsAfter) {
		t.Fatalf("candidate capture wrote to common Git objects: before=%v after=%v", run.objectsBefore, objectsAfter)
	}
	assertC12CandidateRootsRemoved(t, entries)
}

func TestC12Batch01SuiteRejectsTrackedAndUntrackedGroupMutationBeforeLaterCompilation(t *testing.T) {
	for _, mode := range []string{"tracked", "untracked"} {
		t.Run(mode, func(t *testing.T) {
			run := runC12SnapshotSuiteFixture(t, mode)
			if run.exitCode == 0 || !strings.Contains(run.output, "candidate snapshot integrity check failed") {
				t.Errorf("exit=%d output=%q, want %s group mutation rejected by post-run integrity check", run.exitCode, run.output, mode)
			}
			if fileExists(run.laterMutationSentinel) {
				t.Errorf("later package compiled bytes changed by the first group in %s mode", mode)
			}
			assertC12CandidateRootsRemoved(t, readC12SnapshotGoLog(t, run.goLog))
		})
	}
}

func TestC12Batch01SuiteRejectsTransientCandidateTestMainValidatorBypass(t *testing.T) {
	run := runC12TransientTestMainFixture(t)
	if run.exitCode == 0 {
		t.Fatalf("exit=0 output=%q; candidate TestMain bypassed trusted validation", run.output)
	}
	if !strings.Contains(run.output, "trusted candidate validation") {
		t.Fatalf("exit=%d output=%q, want trusted candidate validation rejection", run.exitCode, run.output)
	}
	for name, path := range map[string]string{
		"candidate TestMain": run.testMainLog,
		"Docker":             run.dockerLog,
		"later group":        run.groupLog,
	} {
		if fileExists(path) {
			body, _ := os.ReadFile(path)
			t.Errorf("%s executed before trusted rejection: %q", name, body)
		}
	}
	body, err := os.ReadFile(run.invalidSource)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.HasPrefix(body, []byte("//go:build integration")) {
		t.Fatalf("fixture source was not restored to its invalid staged bytes: %q", body)
	}
}

type c12TransientTestMainRun struct {
	testMainLog   string
	dockerLog     string
	groupLog      string
	invalidSource string
	output        string
	exitCode      int
}

func runC12TransientTestMainFixture(t *testing.T) c12TransientTestMainRun {
	t.Helper()
	repository := t.TempDir()
	runner, err := os.ReadFile("../../scripts/run-c12-integration.ps1")
	if err != nil {
		t.Fatal(err)
	}
	validator, err := os.ReadFile("c12_integration_manifest_test.go")
	if err != nil {
		t.Fatal(err)
	}
	testMainLog := filepath.Join(t.TempDir(), "testmain.log")
	dockerLog := filepath.Join(t.TempDir(), "docker.log")
	groupLog := filepath.Join(t.TempDir(), "group.log")
	invalidRelative := "internal/nodecontrol/contracts/triage_transient_integration_test.go"
	invalidSource := filepath.Join(repository, filepath.FromSlash(invalidRelative))
	files := map[string][]byte{
		"go.mod":                          []byte("module talenro.local/platform\n\ngo 1.25.0\n"),
		"scripts/run-c12-integration.ps1": runner,
		"internal/testinfra/c12_integration_manifest_test.go": validator,
		"internal/testinfra/triage_testmain_test.go": []byte(`package testinfra_test

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil { os.Exit(91) }
	target := filepath.Join(root, "internal", "nodecontrol", "contracts", "triage_transient_integration_test.go")
	original, err := os.ReadFile(target)
	if err != nil { os.Exit(92) }
	tagged := append([]byte("//go:build integration\n\n"), original...)
	if err := os.WriteFile(target, tagged, 0600); err != nil { os.Exit(93) }
	if logPath := os.Getenv("C12_TRIAGE_TESTMAIN_LOG"); logPath != "" {
		_ = os.WriteFile(logPath, []byte(target), 0600)
	}
	code := m.Run()
	if err := os.WriteFile(target, original, 0600); err != nil { os.Exit(94) }
	os.Exit(code)
}
`),
		invalidRelative: []byte(`package contracts_test

import (
	"bytes"
	"os"
	"testing"
)

func TestC12TriageTransientSource(t *testing.T) {
	body, err := os.ReadFile("triage_transient_integration_test.go")
	if err != nil { t.Fatal(err) }
	if bytes.HasPrefix(body, []byte("//go:build integration")) {
		t.Fatal("later group compiled transient validator bytes")
	}
	if err := os.WriteFile(os.Getenv("C12_TRIAGE_GROUP_LOG"), []byte("later group saw original untagged bytes"), 0600); err != nil {
		t.Fatal(err)
	}
}
`),
		"testdata/c12/integration-contracts-schema-authority.v1.json": []byte(`{"schema":"talenro-c12-integration-manifest/v1","groups":[{"id":"triage-transient","package":"./internal/nodecontrol/contracts","profile":"base","tests":["TestC12TriageTransientSource"],"timeout":"2m"}]}`),
	}
	tracked := make([]string, 0, len(files))
	for name, body := range files {
		fullPath := filepath.Join(repository, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, body, 0o600); err != nil {
			t.Fatal(err)
		}
		tracked = append(tracked, name)
	}
	sort.Strings(tracked)
	for _, arguments := range [][]string{{"init", "--quiet"}, append([]string{"add", "--"}, tracked...)} {
		command := exec.Command("git", arguments...)
		command.Dir = repository
		if output, commandErr := command.CombinedOutput(); commandErr != nil {
			t.Fatalf("git %v: %v\n%s", arguments, commandErr, output)
		}
	}

	realGo, err := exec.LookPath("go")
	if err != nil {
		t.Fatal(err)
	}
	realGo, err = filepath.Abs(realGo)
	if err != nil {
		t.Fatal(err)
	}
	fakeBin := t.TempDir()
	const goWrapperSource = `package main
import("encoding/json";"fmt";"os";"os/exec")
func has(args []string,want string)bool{for _,arg:=range args{if arg==want{return true}};return false}
func run(real string,args []string)([]byte,error){cmd:=exec.Command(real,args...);cmd.Env=os.Environ();return cmd.CombinedOutput()}
func exitFor(err error){if err==nil{return};if value,ok:=err.(*exec.ExitError);ok{os.Exit(value.ExitCode())};fmt.Fprintln(os.Stderr,err);os.Exit(98)}
func main(){
 args:=os.Args[1:];real:=os.Getenv("C12_TRIAGE_REAL_GO")
 if len(args)>1&&args[0]=="tool"&&args[1]=="goose"{return}
 if len(args)>0&&args[0]=="test"&&has(args,"-json"){
  plain:=make([]string,0,len(args)-1);for _,arg:=range args{if arg!="-json"{plain=append(plain,arg)}}
  output,err:=run(real,plain);if err!=nil{os.Stderr.Write(output);exitFor(err)}
  packageName:=args[len(args)-1];testName:="TestC12TriageTransientSource";encoder:=json.NewEncoder(os.Stdout)
  encoder.Encode(map[string]string{"Action":"run","Package":packageName,"Test":testName})
  encoder.Encode(map[string]string{"Action":"pass","Package":packageName,"Test":testName})
  encoder.Encode(map[string]string{"Action":"pass","Package":packageName});return
 }
 output,err:=run(real,args);os.Stdout.Write(output);exitFor(err)
}`
	const dockerSource = `package main
import("fmt";"os";"path/filepath";"strconv";"strings")
var ids=map[string]string{"postgres":strings.Repeat("1",64),"redis":strings.Repeat("2",64),"nats":strings.Repeat("3",64)}
var refs=map[string]string{"postgres":"postgres:18.4-alpine3.23","redis":"redis:8.8.1-alpine3.23","nats":"nats:2.14.3-alpine3.22"}
var images=map[string]string{"postgres":"sha256:"+strings.Repeat("a",64),"redis":"sha256:"+strings.Repeat("b",64),"nats":"sha256:"+strings.Repeat("c",64)}
func roleOf(value string)string{for _,role:=range []string{"postgres","redis","nats"}{if strings.Contains(value,role){return role}};if len(value)>0{switch value[0]{case '1':return "postgres";case '2':return "redis";case '3':return "nats"}};return ""}
func state(role,suffix string)string{return filepath.Join(os.Getenv("C12_TRIAGE_DOCKER_STATE"),role+suffix)}
func main(){args:=os.Args[1:];f,_:=os.OpenFile(os.Getenv("C12_TRIAGE_DOCKER_LOG"),os.O_CREATE|os.O_APPEND|os.O_WRONLY,0600);if f!=nil{fmt.Fprintln(f,strings.Join(args," "));f.Close()};joined:=strings.Join(args," ");last:=args[len(args)-1]
 if len(args)>1&&args[0]=="image"&&args[1]=="inspect"{fmt.Println(images[roleOf(last)]);return}
 if args[0]=="run"{role:=roleOf(joined);name:="";for i:=range args{if args[i]=="--name"&&i+1<len(args){name=args[i+1]}};os.WriteFile(state(role,".name"),[]byte(name),0600);os.Remove(state(role,".removed"));fmt.Println(ids[role]);return}
 if len(args)>1&&args[0]=="container"&&args[1]=="inspect"{role:=roleOf(last);if _,err:=os.Stat(state(role,".removed"));err==nil{os.Exit(1)};if strings.Contains(joined,"State.Health.Status"){fmt.Println("healthy");return};if strings.Contains(joined,"{{.Id}}")&&!strings.Contains(joined,"talenro.c12.role"){fmt.Println(ids[role]);return};nameRaw,_:=os.ReadFile(state(role,".name"));name:=string(nameRaw);suffix:=strings.TrimSuffix(strings.TrimPrefix(name,"talenro-c12-"),"-"+role);fmt.Printf("%s|/%s|%s|%s|%s|%s\n",ids[role],name,suffix,role,refs[role],images[role]);return}
 if len(args)>1&&args[0]=="container"&&args[1]=="port"{role:=roleOf(args[len(args)-2]);port:=15432;if role=="redis"{port,_=strconv.Atoi(os.Getenv("C12_TRIAGE_REDIS_PORT"))}else if role=="nats"{port,_=strconv.Atoi(os.Getenv("C12_TRIAGE_NATS_PORT"))};fmt.Printf("127.0.0.1:%d\n",port);return}
 if len(args)>1&&args[0]=="container"&&args[1]=="stop"{return};if len(args)>1&&args[0]=="container"&&args[1]=="rm"{os.WriteFile(state(roleOf(last),".removed"),[]byte("removed"),0600);return};os.Exit(74)
}`
	buildFakeGoExecutable(t, filepath.Join(fakeBin, "go.exe"), goWrapperSource)
	buildFakeGoExecutable(t, filepath.Join(fakeBin, "docker.exe"), dockerSource)
	redisPort := startC12ProtocolServer(t, "+PONG\r\n")
	natsPort := startC12ProtocolServer(t, "PONG\r\n")
	output, exitCode := runC12PowerShellAtRootWithTimeout(t, repository, 90*time.Second, map[string]string{
		"C12_TRIAGE_DOCKER_LOG":   dockerLog,
		"C12_TRIAGE_DOCKER_STATE": t.TempDir(),
		"C12_TRIAGE_GROUP_LOG":    groupLog,
		"C12_TRIAGE_NATS_PORT":    fmt.Sprint(natsPort),
		"C12_TRIAGE_REAL_GO":      realGo,
		"C12_TRIAGE_REDIS_PORT":   fmt.Sprint(redisPort),
		"C12_TRIAGE_TESTMAIN_LOG": testMainLog,
		"Path":                    fakeBin + string(os.PathListSeparator) + os.Getenv("Path"),
	}, "-Suite", "batch01", "-Timeout", "120m")
	return c12TransientTestMainRun{testMainLog: testMainLog, dockerLog: dockerLog, groupLog: groupLog, invalidSource: invalidSource, output: output, exitCode: exitCode}
}

type c12SnapshotSuiteRun struct {
	repository            string
	objectsRoot           string
	objectsBefore         map[string]struct{}
	goLog                 string
	laterMutationSentinel string
	output                string
	exitCode              int
}

type c12SnapshotGoLogEntry struct {
	Stage            string   `json:"stage"`
	WorkingDirectory string   `json:"working_directory"`
	Arguments        string   `json:"arguments"`
	GitEnvironment   []string `json:"git_environment"`
}

func runC12SnapshotSuiteFixture(t *testing.T, mutationMode string) c12SnapshotSuiteRun {
	t.Helper()
	repository := t.TempDir()
	runner, err := os.ReadFile("../../scripts/run-c12-integration.ps1")
	if err != nil {
		t.Fatal(err)
	}
	// Only the copied test runner seeds its fresh, owned cache. Production still
	// creates an empty cache, and every graph discovery and verification runs.
	cacheSeed := prepareC12SnapshotCacheSeed(t, runner)
	cacheHook := "    [void][IO.Directory]::CreateDirectory($goCache)"
	if bytes.Count(runner, []byte(cacheHook)) != 1 {
		t.Fatal("snapshot cache hook is not unique")
	}
	seedScript := cacheHook + "\n    $seed = '" + strings.ReplaceAll(cacheSeed, "'", "''") + "'\n" + `    foreach ($seedFile in [IO.Directory]::EnumerateFiles($seed, '*', [IO.SearchOption]::AllDirectories)) {
      $relative = $seedFile.Substring($seed.Length).TrimStart('\')
      $target = Join-Path $goCache $relative
      [void][IO.Directory]::CreateDirectory([IO.Path]::GetDirectoryName($target))
      [IO.File]::Copy($seedFile, $target, $false)
    }`
	runner = bytes.Replace(runner, []byte(cacheHook), []byte(seedScript), 1)
	files := map[string][]byte{
		"go.mod":                          []byte("module talenro.local/platform\n\ngo 1.26.0\n"),
		"go.sum":                          []byte{},
		"scripts/run-c12-integration.ps1": runner,
		"testdata/c12/integration-contracts-schema-authority.v1.json": []byte(`{"schema":"talenro-c12-integration-manifest/v1","groups":[{"id":"a-first","package":"./internal/testinfra","profile":"base","tests":["TestFirst"],"timeout":"2m"},{"id":"b-second","package":"./internal/store","profile":"base","tests":["TestSecond"],"timeout":"2m"}]}`),
		"internal/store/later.txt":                                    []byte("captured candidate bytes\n"),
		"internal/testinfra/first_integration_test.go":                []byte("//go:build integration\n\npackage testinfra_test\n\nimport \"testing\"\n\nfunc TestFirst(t *testing.T) {}\n"),
		"internal/store/second_integration_test.go":                   []byte("//go:build integration\n\npackage store_test\n\nimport \"testing\"\n\nfunc TestSecond(t *testing.T) {}\n"),
	}
	tracked := make([]string, 0, len(files))
	for name, body := range files {
		fullPath := filepath.Join(repository, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, body, 0o600); err != nil {
			t.Fatal(err)
		}
		tracked = append(tracked, name)
	}
	sort.Strings(tracked)
	commands := [][]string{{"init", "--quiet"}, append([]string{"add", "--"}, tracked...)}
	for _, arguments := range commands {
		command := exec.Command("git", arguments...)
		command.Dir = repository
		if output, commandErr := command.CombinedOutput(); commandErr != nil {
			t.Fatalf("git %v: %v\n%s", arguments, commandErr, output)
		}
	}
	objectsRoot := filepath.Join(repository, ".git", "objects")
	objectsBefore := relativeFileSet(t, objectsRoot)
	if err := os.Remove(filepath.Join(repository, "testdata", "c12", "integration-contracts-schema-authority.v1.json")); err != nil {
		t.Fatal(err)
	}

	fakeBin := t.TempDir()
	goLog := filepath.Join(t.TempDir(), "go.log")
	laterMutationSentinel := filepath.Join(t.TempDir(), "later-mutated")
	realGo, err := exec.LookPath("go")
	if err != nil {
		t.Fatal(err)
	}
	realGo, err = filepath.Abs(realGo)
	if err != nil {
		t.Fatal(err)
	}
	const fakeGoSource = `package main
import("encoding/json";"fmt";"os";"os/exec";"path/filepath";"sort";"strings")
type logEntry struct{Stage string ` + "`json:\"stage\"`" + `;WorkingDirectory string ` + "`json:\"working_directory\"`" + `;Arguments string ` + "`json:\"arguments\"`" + `;GitEnvironment []string ` + "`json:\"git_environment\"`" + `}
func has(args []string,want string)bool{for _,arg:=range args{if arg==want{return true}};return false}
func appendLog(entry logEntry){file,_:=os.OpenFile(os.Getenv("C12_FAKE_GO_LOG"),os.O_CREATE|os.O_APPEND|os.O_WRONLY,0600);if file!=nil{json.NewEncoder(file).Encode(entry);file.Close()}}
func exitFor(err error){if err==nil{return};if value,ok:=err.(*exec.ExitError);ok{os.Exit(value.ExitCode())};fmt.Fprintln(os.Stderr,err);os.Exit(98)}
func main(){
 args:=os.Args[1:];cwd,_:=os.Getwd();gitEnvironment:=[]string{};for _,item:=range os.Environ(){name:=strings.SplitN(item,"=",2)[0];if strings.HasPrefix(strings.ToUpper(name),"GIT_"){gitEnvironment=append(gitEnvironment,name)}};sort.Strings(gitEnvironment)
 stage:="other";if len(args)>0&&args[0]=="run"{stage="validator";for index,arg:=range args{if arg=="-root"&&index+1<len(args){cwd=args[index+1]}}}else if len(args)>0&&args[0]=="tool"{stage="goose"}else if len(args)>0&&args[0]=="test"&&has(args,"-json"){stage=args[len(args)-1]}
 appendLog(logEntry{stage,cwd,strings.Join(args," "),gitEnvironment});if len(gitEnvironment)>0{os.Exit(81)}
 if stage=="validator"{command:=exec.Command(os.Getenv("C12_SNAPSHOT_REAL_GO"),args...);command.Env=os.Environ();output,err:=command.CombinedOutput();os.Stdout.Write(output);exitFor(err);return}
 if stage=="goose"||stage=="other"{return}
 testName:="TestFirst";if stage=="./internal/testinfra"{switch os.Getenv("C12_SNAPSHOT_MUTATION"){case "tracked":os.WriteFile(filepath.Join(cwd,"internal","store","later.txt"),[]byte("mutated by first group\n"),0600);case "untracked":os.WriteFile(filepath.Join(cwd,"internal","store","untracked.go"),[]byte("package store\n"),0600)}}else{testName="TestSecond";body,_:=os.ReadFile(filepath.Join(cwd,"internal","store","later.txt"));_,extraErr:=os.Stat(filepath.Join(cwd,"internal","store","untracked.go"));if string(body)!="captured candidate bytes\n"||extraErr==nil{os.WriteFile(os.Getenv("C12_LATER_MUTATION_SENTINEL"),[]byte("later compiled mutation"),0600)}}
 packageName:=map[string]string{"./internal/testinfra":"talenro.local/platform/internal/testinfra","./internal/store":"talenro.local/platform/internal/store"}[stage];if packageName==""{os.Exit(82)}
 encoder:=json.NewEncoder(os.Stdout);encoder.Encode(map[string]string{"Action":"run","Package":packageName,"Test":testName});encoder.Encode(map[string]string{"Action":"pass","Package":packageName,"Test":testName});encoder.Encode(map[string]string{"Action":"pass","Package":packageName})
}`
	const fakeDockerSource = `package main
import("fmt";"os";"path/filepath";"strconv";"strings")
var ids=map[string]string{"postgres":strings.Repeat("1",64),"redis":strings.Repeat("2",64),"nats":strings.Repeat("3",64)}
var refs=map[string]string{"postgres":"postgres:18.4-alpine3.23","redis":"redis:8.8.1-alpine3.23","nats":"nats:2.14.3-alpine3.22"}
var imageIDs=map[string]string{"postgres":"sha256:"+strings.Repeat("a",64),"redis":"sha256:"+strings.Repeat("b",64),"nats":"sha256:"+strings.Repeat("c",64)}
func roleOf(value string)string{for _,role:=range []string{"postgres","redis","nats"}{if strings.Contains(value,role){return role}};if len(value)>0{switch value[0]{case '1':return "postgres";case '2':return "redis";case '3':return "nats"}};return ""}
func state(role,suffix string)string{return filepath.Join(os.Getenv("C12_DOCKER_STATE"),role+suffix)}
func main(){args:=os.Args[1:];joined:=strings.Join(args," ");last:=args[len(args)-1]
 if len(args)>1&&args[0]=="image"&&args[1]=="inspect"{fmt.Println(imageIDs[roleOf(last)]);return}
 if args[0]=="run"{role:=roleOf(joined);name:="";for i:=range args{if args[i]=="--name"&&i+1<len(args){name=args[i+1]}};os.WriteFile(state(role,".name"),[]byte(name),0600);os.Remove(state(role,".removed"));fmt.Println(ids[role]);return}
 if len(args)>1&&args[0]=="container"&&args[1]=="inspect"{role:=roleOf(last);if strings.Contains(joined,"State.Health.Status"){fmt.Println("healthy");return};if strings.Contains(joined,"{{.Id}}")&&!strings.Contains(joined,"talenro.c12.role"){if _,err:=os.Stat(state(role,".removed"));err==nil{os.Exit(1)};fmt.Println(ids[role]);return};nameRaw,_:=os.ReadFile(state(role,".name"));name:=string(nameRaw);suffix:=strings.TrimSuffix(strings.TrimPrefix(name,"talenro-c12-"),"-"+role);fmt.Printf("%s|/%s|%s|%s|%s|%s\n",ids[role],name,suffix,role,refs[role],imageIDs[role]);return}
 if len(args)>1&&args[0]=="container"&&args[1]=="port"{role:=roleOf(args[len(args)-2]);port:=0;if role=="postgres"{port,_=strconv.Atoi(os.Getenv("C12_POSTGRES_PORT"))}else if role=="redis"{port,_=strconv.Atoi(os.Getenv("C12_REDIS_PORT"))}else{port,_=strconv.Atoi(os.Getenv("C12_NATS_PORT"))};fmt.Printf("127.0.0.1:%d\n",port);return}
 if len(args)>1&&args[0]=="container"&&args[1]=="stop"{return};if len(args)>1&&args[0]=="container"&&args[1]=="rm"{os.WriteFile(state(roleOf(last),".removed"),[]byte("removed"),0600);return};os.Exit(74)
}`
	buildFakeGoExecutable(t, filepath.Join(fakeBin, "go.exe"), fakeGoSource)
	buildFakeGoExecutable(t, filepath.Join(fakeBin, "docker.exe"), fakeDockerSource)
	postgresPort := startC12ProtocolServer(t, "")
	redisPort := startC12ProtocolServer(t, "+PONG\r\n")
	natsPort := startC12ProtocolServer(t, "PONG\r\n")
	dockerState := t.TempDir()
	output, exitCode := runC12PowerShellAtRootWithTimeout(t, repository, 20*time.Minute, map[string]string{
		"C12_DOCKER_STATE":            dockerState,
		"C12_FAKE_GO_LOG":             goLog,
		"C12_LATER_MUTATION_SENTINEL": laterMutationSentinel,
		"C12_NATS_PORT":               fmt.Sprint(natsPort),
		"C12_POSTGRES_PORT":           fmt.Sprint(postgresPort),
		"C12_REDIS_PORT":              fmt.Sprint(redisPort),
		"C12_SNAPSHOT_REAL_GO":        realGo,
		"C12_SNAPSHOT_MUTATION":       mutationMode,
		"Path":                        fakeBin + string(os.PathListSeparator) + os.Getenv("Path"),
	}, "-Suite", "batch01", "-Timeout", "120m")
	return c12SnapshotSuiteRun{repository: repository, objectsRoot: objectsRoot, objectsBefore: objectsBefore, goLog: goLog, laterMutationSentinel: laterMutationSentinel, output: output, exitCode: exitCode}
}

func prepareC12SnapshotCacheSeed(t *testing.T, runner []byte) string {
	t.Helper()
	startMarker := []byte("$script:c12TrustedValidatorSource = @'\n")
	normalized := bytes.ReplaceAll(runner, []byte("\r\n"), []byte("\n"))
	_, tail, found := bytes.Cut(normalized, startMarker)
	if !found {
		t.Fatal("trusted validator source is absent")
	}
	source, _, found := bytes.Cut(tail, []byte("\n'@"))
	if !found {
		t.Fatal("trusted validator source terminator is absent")
	}
	// Use the authenticated child environment and existing trusted test cache;
	// the executable is only a prewarm output and is never used by the runner.
	prewarmRoot := t.TempDir()
	sourcePath := filepath.Join(prewarmRoot, "validator.go")
	if err := os.WriteFile(sourcePath, source, 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()
	command := task8ChildGoCommandContext(t, ctx, "", "build", "-tags=integration", "-o", filepath.Join(prewarmRoot, "prewarm.exe"), sourcePath)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("prewarm trusted snapshot cache: %v\n%s", err, output)
	}
	command = task8ChildGoCommandContext(t, ctx, "", "env", "GOCACHE")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("resolve trusted snapshot cache: %v\n%s", err, output)
	}
	cache := strings.TrimSpace(string(output))
	if !filepath.IsAbs(cache) {
		t.Fatalf("snapshot cache is not absolute: %q", cache)
	}
	return cache
}

func readC12SnapshotGoLog(t *testing.T, logPath string) []c12SnapshotGoLogEntry {
	t.Helper()
	file, err := os.Open(logPath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	var entries []c12SnapshotGoLogEntry
	decoder := json.NewDecoder(file)
	for {
		var entry c12SnapshotGoLogEntry
		if err := decoder.Decode(&entry); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			t.Fatal(err)
		}
		entries = append(entries, entry)
	}
	return entries
}

func assertC12CandidateRootsRemoved(t *testing.T, entries []c12SnapshotGoLogEntry) {
	t.Helper()
	candidateRoots := make(map[string]struct{})
	for _, entry := range entries {
		root := filepath.Clean(entry.WorkingDirectory)
		candidateRoot := filepath.Dir(filepath.Dir(root))
		if strings.HasPrefix(filepath.Base(candidateRoot), "talenro-c12-candidate-") {
			candidateRoots[candidateRoot] = struct{}{}
		}
	}
	if len(candidateRoots) != 1 {
		t.Fatalf("candidate log resolved %d exact owner roots, want 1: %+v", len(candidateRoots), entries)
	}
	for candidateRoot := range candidateRoots {
		if fileExists(candidateRoot) {
			t.Fatalf("exact candidate owner root remains after bounded cleanup: %s", candidateRoot)
		}
	}
}

func startC12ProtocolServer(t *testing.T, response string) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	go func() {
		for {
			connection, acceptErr := listener.Accept()
			if acceptErr != nil {
				return
			}
			_ = connection.SetDeadline(time.Now().Add(3 * time.Second))
			buffer := make([]byte, 4096)
			_, _ = connection.Read(buffer)
			_, _ = io.WriteString(connection, response)
			_ = connection.Close()
		}
	}()
	return listener.Addr().(*net.TCPAddr).Port
}

func validC12ManifestFixture(t *testing.T) ([]byte, []c12IntegrationSource) {
	t.Helper()
	manifest := c12IntegrationManifest{
		Schema: c12IntegrationManifestSchema,
		Groups: []c12IntegrationGroup{{
			ID:      "base-store",
			Package: "./internal/store",
			Profile: "base",
			Tests:   []string{"TestAlpha", "TestBeta"},
			Timeout: "2m",
		}},
	}
	raw, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	sources := []c12IntegrationSource{{
		Path:    "internal/store/base_integration_test.go",
		Tracked: true,
		Body: []byte("//go:build integration\n\npackage store_test\n\nimport \"testing\"\n\n" +
			"func TestAlpha(t *testing.T) {}\n" +
			"func TestBeta(t *testing.T) { t.Run(\"bounded\", func(t *testing.T) {}) }\n"),
	}}
	return raw, sources
}

func cloneC12IntegrationManifest(source c12IntegrationManifest) c12IntegrationManifest {
	clone := source
	clone.Groups = append([]c12IntegrationGroup(nil), source.Groups...)
	for index := range clone.Groups {
		clone.Groups[index].Tests = append([]string(nil), source.Groups[index].Tests...)
	}
	return clone
}

func cloneC12IntegrationSources(source []c12IntegrationSource) []c12IntegrationSource {
	clone := append([]c12IntegrationSource(nil), source...)
	for index := range clone {
		clone[index].Body = append([]byte(nil), source[index].Body...)
	}
	return clone
}

func bytesReplaceOnce(t *testing.T, source, old, replacement []byte) []byte {
	t.Helper()
	index := strings.Index(string(source), string(old))
	if index < 0 {
		t.Fatalf("fixture lacks %q", old)
	}
	result := make([]byte, 0, len(source)-len(old)+len(replacement))
	result = append(result, source[:index]...)
	result = append(result, replacement...)
	result = append(result, source[index+len(old):]...)
	return result
}

func fileExists(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}

func relativeFileSet(t *testing.T, root string) map[string]struct{} {
	t.Helper()
	result := make(map[string]struct{})
	if err := filepath.WalkDir(root, func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(root, name)
		if err != nil {
			return err
		}
		result[filepath.ToSlash(relative)] = struct{}{}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return result
}

func equalStringSets(left, right map[string]struct{}) bool {
	if len(left) != len(right) {
		return false
	}
	for item := range left {
		if _, exists := right[item]; !exists {
			return false
		}
	}
	return true
}

func c12GoJSONEvent(t *testing.T, action, packageName, testName string) string {
	t.Helper()
	event := map[string]string{"Action": action, "Package": packageName}
	if testName != "" {
		event["Test"] = testName
	}
	raw, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func runC12CleanupStateHarness(t *testing.T, mode string) (string, int) {
	t.Helper()
	extraEnvironment := map[string]string{}
	if mode == "cleanup-retained-production" || mode == "docker-exact-production" {
		fakeBin := t.TempDir()
		dockerState := t.TempDir()
		dockerLog := filepath.Join(t.TempDir(), "docker-transcript.log")
		const fakeDockerSource = `package main
import("fmt";"os";"strings")
func failIfDockerEnv(){for _,v:=range os.Environ(){n:=v;if i:=strings.IndexByte(v,'=');i>=0{n=v[:i]};if strings.HasPrefix(strings.ToUpper(n),"DOCKER_"){fmt.Fprintln(os.Stderr,"inherited "+n);os.Exit(91)}}}
func main(){
 args:=os.Args[1:];if len(args)==0{os.Exit(74)};f,_:=os.OpenFile(os.Getenv("C12_DOCKER_LOG"),os.O_CREATE|os.O_APPEND|os.O_WRONLY,0600);fmt.Fprintln(f,strings.Join(args,"\x00"));f.Close();failIfDockerEnv()
 state:=os.Getenv("C12_DOCKER_STATE");suffix:=os.Getenv("C12_RUN_SUFFIX");nonce:=os.Getenv("C12_NONCE_DIGEST");mode:=os.Getenv("C12_DOCKER_MODE");last:=args[len(args)-1]
 if len(args)>=2&&args[0]=="context"&&args[1]=="show"{fmt.Println("default");return}
 if len(args)>=2&&args[0]=="context"&&args[1]=="inspect"{host:="npipe:////./pipe/docker_engine";if mode=="endpoint"{host="tcp://attacker.invalid:2375"};fmt.Printf("[{\"Name\":\"default\",\"Endpoints\":{\"docker\":{\"Host\":%q,\"SkipTLSVerify\":false}}}]\n",host);return}
 if args[0]=="info"{fmt.Println("engine-1111111111111111|28.4.0|linux|amd64");return}
 if len(args)>=2&&args[0]=="image"&&args[1]=="inspect"{id:="sha256:996d0920e4ff9df1fc19dacb904492f3c1ec0ec1cc338f0ad7123be7731c5f5e";if mode=="image"{id="sha256:"+strings.Repeat("f",64)};fmt.Println(id);return}
 if len(args)>=2&&args[0]=="volume"&&args[1]=="create"{os.WriteFile(state,[]byte(last),0600);fmt.Println(last);return}
 if len(args)>=2&&args[0]=="volume"&&args[1]=="inspect"{
  if _,e:=os.Stat(state);e!=nil{if mode=="absence-stdout"{fmt.Println("{}") }else{fmt.Println("[]")};message:="Error response from daemon: get "+last+": no such volume";if mode=="absence-stderr"{message+="!"};fmt.Fprintln(os.Stderr,message);os.Exit(1)}
  driver:="local";managed:="true";if mode=="driver"{driver="attacker"};if mode=="label"{managed="false"}
  if strings.Contains(strings.Join(args," "),".Driver"){fmt.Printf("%s|%s|%s|%s|authority-v7-pitr|primary-data|%s\n",last,driver,managed,suffix,nonce)}else{fmt.Println(last)};return
 }
 if len(args)>=2&&args[0]=="volume"&&args[1]=="rm"{os.Remove(state);if mode=="fail-after-remove"{fmt.Fprintln(os.Stderr,"forced post-remove transport loss");os.Exit(73)};return}
 if len(args)>=2&&args[0]=="container"&&args[1]=="inspect"{fmt.Println("[]");fmt.Fprintln(os.Stderr,"Error response from daemon: No such container: "+last);os.Exit(1)}
 fmt.Fprintln(os.Stderr,"unexpected fake docker argv "+strings.Join(args," "));os.Exit(74)
}`
		buildFakeGoExecutable(t, filepath.Join(fakeBin, "docker.exe"), fakeDockerSource)
		extraEnvironment["Path"] = fakeBin + string(os.PathListSeparator) + os.Getenv("Path")
		extraEnvironment["C12_DOCKER_STATE"] = dockerState
		extraEnvironment["C12_DOCKER_LOG"] = dockerLog
		extraEnvironment["C12_RUN_SUFFIX"] = strings.Repeat("d", 32)
		extraEnvironment["C12_NONCE_DIGEST"] = strings.Repeat("e", 64)
	}
	runner, err := os.ReadFile("../../scripts/run-c12-integration.ps1")
	if err != nil {
		t.Fatal(err)
	}
	marker := []byte("$script:c12RepositoryRoot = (Resolve-Path")
	index := bytes.Index(runner, marker)
	if index < 0 {
		t.Fatal("runner lacks main-program marker")
	}
	fixturePayload := base64.StdEncoding.EncodeToString([]byte(t.TempDir()))
	literalWALPayload := base64.StdEncoding.EncodeToString([]byte(c12LiteralControllerWALBootstrap))
	appendix := fmt.Sprintf(`
$mode = '%s'
$fixtureRoot = [Text.Encoding]::UTF8.GetString([Convert]::FromBase64String('%s'))
$literalControllerWAL = [Convert]::FromBase64String('%s')

try {
  if ($mode -ceq 'docker-exact') { $mode = 'attempted-unconfirmed-volume' }
  switch ($mode) {
	'prepared-direct-ledger' {
	  $artifact = $null
	  $primaryFailure = $null
	  try {
	    $artifact = New-C12PreparedArtifactRoot -RunSuffix ([Guid]::NewGuid().ToString('N')) -Profile 'authority-v7-pitr'
	    if ($artifact.PSObject.Properties.Name -cnotcontains 'Ledger') { throw 'production prepared artifact root lacks a direct-leaf ledger' }
	    $ledgerNames = @($artifact.Ledger | ForEach-Object { [string]$_.Name })
	    if ($ledgerNames.Count -lt 2 -or $ledgerNames -cnotcontains 'go-cache' -or $ledgerNames -cnotcontains 'go-tmp') {
	      throw ('production prepared direct-leaf positive ledger mismatch: ' + ($ledgerNames -join '|'))
	    }
	    $unknown = Join-Path ([string]$artifact.Root) 'unknown-direct-sibling.bin'
	    [IO.File]::WriteAllBytes($unknown, [byte[]](1,2,3,4))
	    $rejected = $false
	    try { Remove-C12PreparedArtifactRoot -ArtifactRoot $artifact -Deadline ([DateTime]::UtcNow.AddSeconds(20)) }
	    catch { $rejected = $_.Exception.Message -match 'unknown.*direct|direct.*sibling' }
	    if (-not $rejected) { throw 'production prepared ledger accepted an unknown direct sibling' }
	    if (-not [IO.Directory]::Exists([string]$artifact.Root) -or -not [IO.File]::Exists($unknown)) { throw 'prepared ledger mutation deleted owned state on rejection' }
	    [IO.File]::Delete($unknown)

	    $hardLinkSource = Join-Path ([string]$artifact.Root) 'hard-link-source.bin'
	    $hardLinkSplice = Join-Path $fixtureRoot 'hard-link-splice.bin'
	    $null = Register-C12DirectLeafIntent -Ledger $artifact.Ledger -Name 'hard-link-source.bin' -Kind 'exact_file' -Expected $true
	    [IO.File]::WriteAllBytes($hardLinkSource, [byte[]](9,8,7,6))
	    $null = Bind-C12DirectLeaf -ArtifactRoot $artifact -Name 'hard-link-source.bin'
	    $hardLinkResult = & (Join-Path $env:SystemRoot 'System32\cmd.exe') /d /c mklink /H $hardLinkSplice $hardLinkSource 2>&1
	    if ($LASTEXITCODE -ne 0) { throw ('create hard-link splice failed: ' + (@($hardLinkResult) -join ' ')) }
	    $rejected = $false
	    try { Remove-C12PreparedArtifactRoot -ArtifactRoot $artifact -Deadline ([DateTime]::UtcNow.AddSeconds(20)) }
	    catch { $rejected = $_.Exception.Message -match 'link|identity' }
	    if (-not $rejected -or -not [IO.File]::Exists($hardLinkSource) -or -not [IO.File]::Exists($hardLinkSplice)) { throw 'prepared ledger did not retain a hard-link splice' }
	    [IO.File]::Delete($hardLinkSplice)

	    $reparse = Join-Path ([string]$artifact.Root) 'reparse-leaf'
	    $null = Register-C12DirectLeafIntent -Ledger $artifact.Ledger -Name 'reparse-leaf' -Kind 'owned_ephemeral_subtree' -Expected $true
	    $null = New-Item -ItemType Junction -Path $reparse -Target $fixtureRoot
	    $rejected = $false
	    try { Bind-C12DirectLeaf -ArtifactRoot $artifact -Name 'reparse-leaf' }
	    catch { $rejected = $_.Exception.Message -match 'reparse|identity|ownership' }
	    if (-not $rejected -or -not [IO.Directory]::Exists($reparse)) { throw 'prepared ledger accepted or deleted a reparse leaf' }
	    $cleanupRejected = $false
	    try { Remove-C12PreparedArtifactRoot -ArtifactRoot $artifact -Deadline ([DateTime]::UtcNow.AddSeconds(20)) }
	    catch { $cleanupRejected = $_.Exception.Message -match 'identity|bound|reparse' }
	    if (-not $cleanupRejected -or -not [IO.Directory]::Exists($reparse)) { throw 'prepared cleanup accepted or deleted a foreign reparse leaf' }
	    [IO.Directory]::Delete($reparse, $false)
	    $reparseEntry = @($artifact.Ledger | Where-Object { [string]$_.Name -ceq 'reparse-leaf' })[0]
	    $reparseEntry.Expected = $false

	    $replace = Join-Path ([string]$artifact.Root) 'replace-leaf'
	    $null = Register-C12DirectLeafIntent -Ledger $artifact.Ledger -Name 'replace-leaf' -Kind 'owned_ephemeral_subtree' -Expected $true
	    $replaceOwnership = New-C12OwnedDirectory -Root $replace -ExpectedParent ([string]$artifact.Root) -LeafPattern '^replace-leaf$' -Stage 'test replacement leaf creation'
	    $null = Bind-C12DirectLeaf -ArtifactRoot $artifact -Name 'replace-leaf' -Ownership $replaceOwnership
	    $replaceOwnership.Dispose()
	    [IO.Directory]::Delete($replace, $false)
	    [void][IO.Directory]::CreateDirectory($replace)
	    $rejected = $false
	    try { Remove-C12PreparedArtifactRoot -ArtifactRoot $artifact -Deadline ([DateTime]::UtcNow.AddSeconds(20)) }
	    catch { $rejected = $_.Exception.Message -match 'identity|substitut|changed|ownership|disposed' }
	    if (-not $rejected -or -not [IO.Directory]::Exists($replace)) { throw 'prepared ledger accepted or deleted a replaced directory identity' }
	    [IO.Directory]::Delete($replace, $false)
	    $replaceEntry = @($artifact.Ledger | Where-Object { [string]$_.Name -ceq 'replace-leaf' })[0]
	    $replaceEntry.Expected = $false

	    Remove-C12PreparedArtifactRoot -ArtifactRoot $artifact -Deadline ([DateTime]::UtcNow.AddSeconds(20))
	    if ([IO.Directory]::Exists([string]$artifact.Root)) { throw 'production prepared ledger positive cleanup retained its closed root' }
	    $artifact = $null
	    Write-Output 'C12_PREPARED_DIRECT_LEDGER_OK'
	    exit 0
	  }
	  catch { $primaryFailure = $_; throw }
	  finally {
	    if ($null -ne $artifact -and [IO.Directory]::Exists([string]$artifact.Root)) {
	      try { Remove-C12BoundedDirectory -Root ([string]$artifact.Root) -ExpectedParent ([string]$artifact.Parent) -LeafPattern '^talenro-c12-artifacts-[0-9a-f]{32}$' -Stage 'test prepared direct-ledger cleanup' -Deadline ([DateTime]::UtcNow.AddSeconds(20)) -Ownership $artifact.Ownership }
	      catch { if ($null -eq $primaryFailure) { throw } }
	    }
	  }
	}

	'pitr-five-leaf' {
	  $runRoot = $null
	  try {
	    $runRoot = New-C12PITRRunRoot -RunSuffix ([Guid]::NewGuid().ToString('N'))
	    if ($runRoot.PSObject.Properties.Name -cnotcontains 'Ledger') { throw 'production PITR run-root lacks retained five-leaf ledger' }
	    $names = @($runRoot.Ledger | ForEach-Object { [string]$_.Name })
	    $expected = @('controller-ownership-v1.wal','ownership.wal','tlsgen.go','server.crt','server.key')
	    if (($names -join '|') -cne ($expected -join '|')) { throw ('production PITR run-root ledger mismatch: ' + ($names -join '|')) }
	    if ($null -eq (Get-Command Remove-C12PITRRunRoot -CommandType Function -ErrorAction SilentlyContinue)) { throw 'production PITR run-root cleanup entry is absent' }
	    $sixth = Join-Path ([string]$runRoot.Root) 'sixth-unregistered-leaf.bin'
	    [IO.File]::WriteAllBytes($sixth, [byte[]](5,6,7,8))
	    $rejected = $false
	    try { Remove-C12PITRRunRoot -RunRoot $runRoot -Deadline ([DateTime]::UtcNow.AddSeconds(20)) }
	    catch { $rejected = $_.Exception.Message -match 'unknown.*direct|direct.*sibling|unregistered.*leaf' }
	    if (-not $rejected) { throw 'production PITR ledger accepted a sixth direct leaf' }
	    if (-not [IO.Directory]::Exists([string]$runRoot.Root) -or -not [IO.File]::Exists($sixth)) { throw 'PITR sixth-leaf rejection deleted retained state' }
	    [IO.File]::Delete($sixth)
	    Remove-C12PITRRunRoot -RunRoot $runRoot -Deadline ([DateTime]::UtcNow.AddSeconds(20))
	    $runRoot = $null
	    Write-Output 'C12_PITR_FIVE_LEAF_LEDGER_OK'
	    exit 0
	  }
	  finally {
	    if ($null -ne $runRoot -and [IO.Directory]::Exists([string]$runRoot.Root)) {
	      Remove-C12BoundedDirectory -Root ([string]$runRoot.Root) -ExpectedParent ([IO.Path]::GetTempPath()) -LeafPattern '^talenro-c12-pitr-[0-9a-f]{32}$' -Stage 'test PITR run-root cleanup' -Deadline ([DateTime]::UtcNow.AddSeconds(20)) -Ownership $runRoot.Ownership
	    }
	  }
	}

	'controller-wal' {
	  $runRoot = $null
	  try {
	    $runRoot = New-C12PITRRunRoot -RunSuffix ([Guid]::NewGuid().ToString('N'))
	    if ($null -eq (Get-Command Assert-C12ControllerOwnershipWAL -CommandType Function -ErrorAction SilentlyContinue)) { throw 'production controller-WAL verifier entry is absent' }
	    if ($runRoot.PSObject.Properties.Name -cnotcontains 'ControllerWAL' -or $runRoot.ControllerWAL.PSObject.Properties.Name -cnotcontains 'Path') { throw 'production PITR run-root lacks controller WAL state/path' }
	    $walPath = [string]$runRoot.ControllerWAL.Path
	    $observed = [IO.File]::ReadAllBytes($walPath)
	    if (-not [Linq.Enumerable]::SequenceEqual([byte[]]$observed, [byte[]]$literalControllerWAL)) { throw 'production BOOTSTRAP WAL bytes differ from the reviewed literal line' }
	    Assert-C12ControllerOwnershipWAL -State $runRoot.ControllerWAL
	    $baseline = [byte[]]$observed.Clone()
	    $baselineText = [Text.Encoding]::UTF8.GetString($baseline)
	    $mutations = [ordered]@{
	      payload = $baselineText.Replace('"registry":[]','"registry":[1]')
	      payload_digest = $baselineText.Replace('"payload_digest":"c747','"payload_digest":"0747')
	      record_digest = $baselineText.Replace('"record_digest":"dc13','"record_digest":"0c13')
	      hmac = $baselineText.Replace('"hmac_sha256":"827f','"hmac_sha256":"027f')
	      previous = $baselineText.Replace('"previous_record_digest":null','"previous_record_digest":"0000000000000000000000000000000000000000000000000000000000000000"')
	    }
	    foreach ($entry in $mutations.GetEnumerator()) {
	      [IO.File]::WriteAllBytes($walPath, [Text.Encoding]::UTF8.GetBytes([string]$entry.Value))
	      $rejected = $false
	      try { Assert-C12ControllerOwnershipWAL -State $runRoot.ControllerWAL }
	      catch { $rejected = $_.Exception.Message -match [string]$entry.Key }
	      if (-not $rejected) { throw ('production controller WAL verifier accepted or misclassified ' + [string]$entry.Key + ' mutation') }
	      [IO.File]::WriteAllBytes($walPath, $baseline)
	    }
	    Assert-C12ControllerOwnershipWAL -State $runRoot.ControllerWAL
	    Write-Output 'C12_CONTROLLER_WAL_LITERAL_OK'
	    exit 0
	  }
	  finally {
	    if ($null -ne $runRoot -and [IO.Directory]::Exists([string]$runRoot.Root)) {
	      Remove-C12BoundedDirectory -Root ([string]$runRoot.Root) -ExpectedParent ([IO.Path]::GetTempPath()) -LeafPattern '^talenro-c12-pitr-[0-9a-f]{32}$' -Stage 'test controller-WAL run-root cleanup' -Deadline ([DateTime]::UtcNow.AddSeconds(20)) -Ownership $runRoot.Ownership
	    }
	  }
	}

    'prepared-deadline' {
      $root = Join-Path $fixtureRoot "talenro-c12-artifacts-$([Guid]::NewGuid().ToString('N'))"
      $ownership = $null
      try {
        $ownership = New-C12OwnedDirectory -Root $root -ExpectedParent $fixtureRoot -LeafPattern '^talenro-c12-artifacts-[0-9a-f]{32}$' -Stage 'prepared retry fixture creation'
	    $ledger = New-C12DirectLeafLedger
        $artifact = [pscustomobject]@{
          Root = $root
          Parent = $fixtureRoot
          Ownership = $ownership
          ArtifactRootIdentity = [string]$ownership.Identity
          Ledger = $ledger
          Closed = $false
        }
	    $null = Register-C12DirectLeafIntent -Ledger $artifact.Ledger -Name 'go-cache' -Kind 'owned_ephemeral_subtree' -Expected $true
	    $null = Register-C12DirectLeafIntent -Ledger $artifact.Ledger -Name 'go-tmp' -Kind 'owned_ephemeral_subtree' -Expected $true
	    $retryPath = Join-Path $root 'retry-state.bin'
	    $retryEntry = Register-C12DirectLeafIntent -Ledger $artifact.Ledger -Name 'retry-state.bin' -Kind 'exact_file' -Expected $true
	    [IO.File]::WriteAllBytes($retryPath, [byte[]](1,3,3,7))
	    $null = Bind-C12DirectLeaf -ArtifactRoot $artifact -Name 'retry-state.bin'
        $script:c12PreparedArtifactRoot = $artifact
        $script:c12PreparedReceipts.Clear()
        $script:c12PreparedWorkers.Clear()
        $script:c12PreparedReceiptKeyHex = (('a' * 64) -join '')

        $expiredRejected = $false
        try {
          Remove-C12PreparedArtifactRoot -ArtifactRoot $artifact -Deadline ([DateTime]::UtcNow.AddSeconds(-1))
        }
        catch {
          $expiredRejected = $_.Exception.Message.Contains('exceeded its absolute deadline')
        }
        if (-not $expiredRejected) { throw 'prepared artifact cleanup accepted an expired deadline' }
        if ($null -eq $script:c12PreparedArtifactRoot) { throw 'failed prepared cleanup cleared its global root state' }
	    if ([string]$retryEntry.Lifecycle -cne 'CleanIntent' -or [string]::IsNullOrEmpty([string]$retryEntry.LastCleanupError) -or
	        $null -eq $retryEntry.CleanupHandle -or [string]::IsNullOrEmpty([string]$retryEntry.Identity)) {
	      throw 'expired prepared cleanup did not retain exact CleanIntent identity/error/handle state'
	    }
        $ownership.VerifyExactPath()

        Remove-C12PreparedArtifactRoot -ArtifactRoot $artifact -Deadline ([DateTime]::UtcNow.AddSeconds(15))
        $ownership = $null
        if (-not [bool]$artifact.Closed) { throw 'fresh prepared cleanup did not close its artifact state' }
	    if ([string]$retryEntry.Lifecycle -cne 'Absent' -or -not [string]::IsNullOrEmpty([string]$retryEntry.LastCleanupError)) { throw 'fresh prepared cleanup did not converge the retained ledger entry' }
        if ($null -ne $script:c12PreparedArtifactRoot -or
            $script:c12PreparedReceipts.Count -ne 0 -or
            $script:c12PreparedWorkers.Count -ne 0 -or
            -not [string]::IsNullOrEmpty($script:c12PreparedReceiptKeyHex)) {
          throw 'fresh prepared cleanup did not clear all global controller state'
        }
        if ([IO.Directory]::Exists($root) -or [IO.File]::Exists($root)) {
          throw 'fresh prepared cleanup left its exact root'
        }
      }
      finally {
        if ($null -ne $ownership) { $ownership.Dispose() }
        if ([IO.Directory]::Exists($root)) { [IO.Directory]::Delete($root, $false) }
      }
      Write-Output 'C12_PREPARED_DEADLINE_RETRY_OK'
      exit 0
    }

    'cleanup-retained-production' {
      $script:c12RepositoryRoot = $fixtureRoot
      $runSuffix = (('d' * 32) -join '')
      $nonceDigest = (('e' * 64) -join '')
      $volume = New-C12PITRVolumeResource -Name "talenro-c12-$runSuffix-pitr-primary-data" -Role 'primary-data' -NonceDigest $nonceDigest
      Start-C12PITRVolume -Resource $volume -RunSuffix $runSuffix -Deadline ([DateTime]::UtcNow.AddSeconds(30))
      if ($volume.PSObject.Properties.Name -cnotcontains 'Phase' -or
          $volume.PSObject.Properties.Name -cnotcontains 'OwnershipHandle' -or
          $volume.PSObject.Properties.Name -cnotcontains 'RetryState') {
        throw 'production PITR volume lacks retained phase/ownership/retry state after positive create+inspect'
      }
      if ([string]$volume.Phase -cne 'Verified') { throw ('production PITR volume positive phase = ' + [string]$volume.Phase) }
      $ownershipHandle = $volume.OwnershipHandle
      $retryState = $volume.RetryState
      [Environment]::SetEnvironmentVariable('C12_DOCKER_MODE', 'fail-after-remove', 'Process')
      $failed = $false
      try { Remove-C12PITRBaseResources -Resources @() -Volumes @($volume) -RunSuffix $runSuffix -Deadline ([DateTime]::UtcNow.AddSeconds(30)) }
      catch { $failed = $_.Exception.Message -match 'exit code 73|post-remove transport loss' }
      if (-not $failed) { throw 'production cleanup did not surface the post-remove/pre-result mutation' }
      if ([string]$volume.Phase -cne 'CleanIntent' -or $volume.OwnershipHandle -ne $ownershipHandle -or $volume.RetryState -ne $retryState) {
        throw 'failed cleanup did not retain the same CleanIntent ownership/retry object'
      }
      [Environment]::SetEnvironmentVariable('C12_DOCKER_MODE', $null, 'Process')
      Remove-C12PITRBaseResources -Resources @() -Volumes @($volume) -RunSuffix $runSuffix -Deadline ([DateTime]::UtcNow.AddSeconds(30))
      if ([string]$volume.Phase -cnotin @('Removed','Absent')) { throw ('fresh-deadline retry did not reach a terminal phase: ' + [string]$volume.Phase) }
      Write-Output 'C12_PRODUCTION_CLEANUP_RETAINED_STATE_OK'
      exit 0
    }

    'docker-exact-production' {
      $script:c12RepositoryRoot = $fixtureRoot
      foreach ($name in @('DOCKER_HOST','DOCKER_CONTEXT','DOCKER_CONFIG','DOCKER_TLS_VERIFY','DOCKER_CERT_PATH','DOCKER_API_VERSION')) {
        [Environment]::SetEnvironmentVariable($name, 'attacker-canary', 'Process')
      }
      try {
        $context = Invoke-C12Docker -Arguments @('context','show') -Stage 'freeze Docker context receipt' -Deadline ([DateTime]::UtcNow.AddSeconds(30))
        if ($context.ExitCode -ne 0 -or (@($context.Output) -join '') -cne 'default') { throw 'production Docker positive context transcript mismatch' }
        if ($context.PSObject.Properties.Name -cnotcontains 'ExecutableReceipt' -or $context.PSObject.Properties.Name -cnotcontains 'EndpointReceipt') {
          throw 'production Invoke-C12Docker lacks executable and endpoint receipts'
        }
        if ($null -eq (Get-Command New-C12PITRContainerResource -CommandType Function -ErrorAction SilentlyContinue) -or
            $null -eq (Get-Command Assert-C12PITRContainerReceipt -CommandType Function -ErrorAction SilentlyContinue)) {
          throw 'production runner lacks a constructible PITR container-resource/receipt boundary; container receipt coverage cannot be synthesized by this harness'
        }
        $endpointDigest = [string]$context.EndpointReceipt.Digest
        $executableDigest = [string]$context.ExecutableReceipt.Digest
        if ($endpointDigest -notmatch '^[0-9a-f]{64}$' -or $executableDigest -notmatch '^[0-9a-f]{64}$') { throw 'production Docker receipt digest is malformed' }
        $endpoint = Invoke-C12Docker -Arguments @('context','inspect','default') -Stage 'freeze Docker endpoint' -Deadline ([DateTime]::UtcNow.AddSeconds(30))
        $engine = Invoke-C12Docker -Arguments @('info','--format','{{.ID}}|{{.ServerVersion}}|{{.OSType}}|{{.Architecture}}') -Stage 'freeze Docker engine' -Deadline ([DateTime]::UtcNow.AddSeconds(30))
        $image = Invoke-C12Docker -Arguments @('image','inspect','--format','{{.Id}}','postgres:18.4-alpine3.23') -Stage 'freeze Docker image receipt' -Deadline ([DateTime]::UtcNow.AddSeconds(30))
        foreach ($result in @($endpoint,$engine,$image)) {
          if ([string]$result.EndpointReceipt.Digest -cne $endpointDigest -or [string]$result.ExecutableReceipt.Digest -cne $executableDigest) { throw 'Docker operation did not revalidate the same executable+endpoint receipt' }
        }
        if ((@($image.Output) -join '') -cne 'sha256:996d0920e4ff9df1fc19dacb904492f3c1ec0ec1cc338f0ad7123be7731c5f5e') { throw 'production Docker image positive receipt mismatch' }

        $runSuffix = (('d' * 32) -join '')
        $nonceDigest = (('e' * 64) -join '')
        $volume = New-C12PITRVolumeResource -Name "talenro-c12-$runSuffix-pitr-primary-data" -Role 'primary-data' -NonceDigest $nonceDigest
        Start-C12PITRVolume -Resource $volume -RunSuffix $runSuffix -Deadline ([DateTime]::UtcNow.AddSeconds(30))
        foreach ($mutation in @('endpoint','image','driver','label')) {
          [Environment]::SetEnvironmentVariable('C12_DOCKER_MODE', $mutation, 'Process')
          $rejected = $false
          try {
            if ($mutation -ceq 'image') {
              $null = Invoke-C12Docker -Arguments @('image','inspect','--format','{{.Id}}','postgres:18.4-alpine3.23') -Stage 'mutated Docker image receipt' -Deadline ([DateTime]::UtcNow.AddSeconds(20))
            } else {
              $null = Remove-C12PITRBaseResources -Resources @() -Volumes @($volume) -RunSuffix $runSuffix -Deadline ([DateTime]::UtcNow.AddSeconds(20))
            }
          }
          catch { $rejected = $_.Exception.Message -match $mutation }
          if (-not $rejected) { throw ('production Docker parser accepted or misclassified ' + $mutation + ' mutation') }
          if (-not [IO.File]::Exists($env:C12_DOCKER_STATE)) { throw ($mutation + ' mutation removed the exact owned volume') }
        }
        [Environment]::SetEnvironmentVariable('C12_DOCKER_MODE', $null, 'Process')
        Remove-C12PITRBaseResources -Resources @() -Volumes @($volume) -RunSuffix $runSuffix -Deadline ([DateTime]::UtcNow.AddSeconds(30))
        foreach ($mutation in @('absence-stdout','absence-stderr')) {
          [Environment]::SetEnvironmentVariable('C12_DOCKER_MODE', $mutation, 'Process')
          $rejected = $false
          try { $null = Invoke-C12Docker -Arguments @('volume','inspect','--format','{{.Name}}',[string]$volume.Name) -Stage 'mutated exact absence transcript' -AllowFailure -Deadline ([DateTime]::UtcNow.AddSeconds(20)) }
          catch { $rejected = $_.Exception.Message -match 'absence|ambiguous' }
          if (-not $rejected) { throw ('production Docker exact-absence parser accepted ' + $mutation) }
        }
        Write-Output 'C12_PRODUCTION_DOCKER_EXACT_OK'
        exit 0
      }
      finally {
        [Environment]::SetEnvironmentVariable('C12_DOCKER_MODE', $null, 'Process')
        foreach ($name in @('DOCKER_HOST','DOCKER_CONTEXT','DOCKER_CONFIG','DOCKER_TLS_VERIFY','DOCKER_CERT_PATH','DOCKER_API_VERSION')) { [Environment]::SetEnvironmentVariable($name, $null, 'Process') }
      }
    }

    'group-before-run-root' {
      $script:c12HarnessCandidateCalls = 0
      $script:c12HarnessBaseCalls = 0
      function New-C12PreparedAuthorityInitializer {
        throw 'forced initializer failure before PITR run-root acquisition'
      }
      function Remove-C12PITRCandidates {
        $script:c12HarnessCandidateCalls++
      }
      function Remove-C12PITRBaseResources {
        $script:c12HarnessBaseCalls++
      }
      $groupFailed = $false
      try {
        Invoke-C12Group -GroupID 'cleanup-before-run-root' -GroupProfile 'authority-v7-pitr' -Package './internal/testinfra' -GroupTimeout '3m' -AbsoluteDeadline ([DateTime]::UtcNow.AddMinutes(2))
      }
      catch {
        $groupFailed = $_.Exception.Message -ceq 'C12 group cleanup-before-run-root failed'
      }
      if (-not $groupFailed) { throw 'pre-run-root group did not preserve its primary failure' }
      if ($script:c12HarnessCandidateCalls -ne 0) { throw 'pre-run-root group invoked PITR candidate cleanup' }
      Write-Output ('C12_PRE_RUN_ROOT_CLEANUP_OK BASE_CALLS=' + $script:c12HarnessBaseCalls)
      exit 0
    }

    'never-attempted-volume' {
      $script:c12HarnessDockerCalls = 0
      function Invoke-C12Docker {
        param(
          [string[]]$Arguments,
          [string]$Stage,
          [TimeSpan]$Timeout = [TimeSpan]::FromSeconds(15),
          [DateTime]$Deadline = [DateTime]::MaxValue,
          [switch]$AllowFailure
        )
        $script:c12HarnessDockerCalls++
        throw 'never-attempted PITR volume reached Docker'
      }
      $runSuffix = (('b' * 32) -join '')
      $nonceDigest = (('c' * 64) -join '')
      $volume = New-C12PITRVolumeResource -Name "talenro-c12-$runSuffix-pitr-primary-data" -Role 'primary-data' -NonceDigest $nonceDigest
      if ($volume.PSObject.Properties.Name -cnotcontains 'CreateAttempted') {
        throw 'PITR volume state lacks CreateAttempted'
      }
      if ([bool]$volume.CreateAttempted) { throw 'new PITR volume is already marked attempted' }
      Remove-C12PITRBaseResources -Resources @() -Volumes @($volume) -RunSuffix $runSuffix -Deadline ([DateTime]::UtcNow.AddSeconds(15))
      if ($script:c12HarnessDockerCalls -ne 0) { throw 'never-attempted PITR volume performed Docker I/O' }
      Write-Output 'C12_NEVER_ATTEMPTED_VOLUME_SKIPPED'
      exit 0
    }

    'attempted-unconfirmed-volume' {
      $script:c12HarnessDockerMode = ''
      $script:c12HarnessRemoveCalls = 0
      $script:c12HarnessInspectCalls = 0
      $runSuffix = (('d' * 32) -join '')
      $nonceDigest = (('e' * 64) -join '')
      $volumeName = "talenro-c12-$runSuffix-pitr-primary-data"
      function Invoke-C12Docker {
        param(
          [string[]]$Arguments,
          [string]$Stage,
          [TimeSpan]$Timeout = [TimeSpan]::FromSeconds(15),
          [DateTime]$Deadline = [DateTime]::MaxValue,
          [switch]$AllowFailure
        )
        if ($Arguments.Count -ge 2 -and $Arguments[0] -ceq 'volume' -and $Arguments[1] -ceq 'inspect') {
          $script:c12HarnessInspectCalls++
          $format = if ($Arguments.Count -ge 4) { [string]$Arguments[3] } else { '' }
          if ($format -ceq '{{.Name}}') {
            return [pscustomobject]@{ ExitCode = 1; Output = @() }
          }
          switch ($script:c12HarnessDockerMode) {
            'absent' {
              return [pscustomobject]@{ ExitCode = 1; Output = @() }
            }
            'foreign' {
              $foreign = "$volumeName|local|true|$runSuffix|authority-v7-pitr|archive|$nonceDigest"
              return [pscustomobject]@{ ExitCode = 0; Output = @($foreign) }
            }
            'matching' {
              $matching = "$volumeName|local|true|$runSuffix|authority-v7-pitr|primary-data|$nonceDigest"
              return [pscustomobject]@{ ExitCode = 0; Output = @($matching) }
            }
          }
        }
        if ($Arguments.Count -ge 2 -and $Arguments[0] -ceq 'volume' -and $Arguments[1] -ceq 'rm') {
          $script:c12HarnessRemoveCalls++
          if ($script:c12HarnessDockerMode -cne 'matching' -or [string]$Arguments[-1] -cne $volumeName) {
            throw 'PITR cleanup removed a volume without matching identity'
          }
          return [pscustomobject]@{ ExitCode = 0; Output = @() }
        }
        throw ('unexpected PITR volume cleanup call: ' + ($Arguments -join ' '))
      }

      $volume = New-C12PITRVolumeResource -Name $volumeName -Role 'primary-data' -NonceDigest $nonceDigest
      if ($volume.PSObject.Properties.Name -cnotcontains 'CreateAttempted') {
        throw 'PITR volume state lacks CreateAttempted'
      }
      $volume.CreateAttempted = $true
      $volume.Created = $false

      $script:c12HarnessDockerMode = 'absent'
      Remove-C12PITRBaseResources -Resources @() -Volumes @($volume) -RunSuffix $runSuffix -Deadline ([DateTime]::UtcNow.AddSeconds(15))
      if ($script:c12HarnessRemoveCalls -ne 0) { throw 'absent attempted volume was removed' }

      $script:c12HarnessDockerMode = 'foreign'
      $foreignRejected = $false
      try {
        Remove-C12PITRBaseResources -Resources @() -Volumes @($volume) -RunSuffix $runSuffix -Deadline ([DateTime]::UtcNow.AddSeconds(15))
      }
      catch {
        $foreignRejected = $_.Exception.Message.Contains('identity mismatch')
      }
      if (-not $foreignRejected) { throw 'foreign attempted volume identity was accepted' }
      if ($script:c12HarnessRemoveCalls -ne 0) { throw 'foreign attempted volume was removed' }

      $script:c12HarnessDockerMode = 'matching'
      Remove-C12PITRBaseResources -Resources @() -Volumes @($volume) -RunSuffix $runSuffix -Deadline ([DateTime]::UtcNow.AddSeconds(15))
      if ($script:c12HarnessRemoveCalls -ne 1) { throw 'matching attempted volume was not removed exactly once' }
      if ($script:c12HarnessInspectCalls -lt 4) { throw 'attempted volume was not re-inspected and absence-verified' }
      Write-Output 'C12_ATTEMPTED_UNCONFIRMED_VOLUME_REINSPECTED'
      exit 0
    }
  }
  throw ('unknown cleanup-state harness mode ' + $mode)
}
catch {
  [Console]::Error.WriteLine($_.Exception.Message)
  exit 1
}
`, mode, fixturePayload, literalWALPayload)
	harness := filepath.Join(t.TempDir(), "assert-c12-cleanup-state.ps1")
	if err := os.WriteFile(harness, append(append([]byte(nil), runner[:index]...), []byte(appendix)...), 0o600); err != nil {
		t.Fatal(err)
	}
	harnessTimeout := 45 * time.Second
	if mode == "prepared-direct-ledger" {
		harnessTimeout = 2 * time.Minute
	}
	ctx, cancel := context.WithTimeout(context.Background(), harnessTimeout)
	defer cancel()
	powershellPath := filepath.Join(os.Getenv("SystemRoot"), "System32", "WindowsPowerShell", "v1.0", "powershell.exe")
	command := exec.CommandContext(ctx, powershellPath, "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", harness,
		"-Profile", "base", "-Packages", "./internal/store", "-Timeout", "3m")
	command.Env = append([]string(nil), os.Environ()...)
	for name, value := range extraEnvironment {
		command.Env = append(command.Env, name+"="+value)
	}
	output, commandErr := command.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("cleanup-state %s harness timed out: %v\n%s", mode, ctx.Err(), output)
	}
	if commandErr == nil {
		return string(output), 0
	}
	var exitError *exec.ExitError
	if !errors.As(commandErr, &exitError) {
		t.Fatalf("launch cleanup-state %s harness: %v", mode, commandErr)
	}
	return string(output), exitError.ExitCode()
}

func runC12ExpiredGroupCleanupHarness(t *testing.T) (string, int) {
	t.Helper()
	runner, err := os.ReadFile("../../scripts/run-c12-integration.ps1")
	if err != nil {
		t.Fatal(err)
	}
	marker := []byte("$script:c12RepositoryRoot = (Resolve-Path")
	index := bytes.Index(runner, marker)
	if index < 0 {
		t.Fatal("runner lacks main-program marker")
	}
	rootPayload := base64.StdEncoding.EncodeToString([]byte(t.TempDir()))
	appendix := fmt.Sprintf(`
$script:c12RepositoryRoot = [Text.Encoding]::UTF8.GetString([Convert]::FromBase64String('%s'))
$script:c12HarnessPriorDeadline = [DateTime]::UtcNow.AddMinutes(5)
$script:c12NativeDeadline = $script:c12HarnessPriorDeadline
$script:c12HarnessResources = @{}
$script:c12HarnessCalls = New-Object System.Collections.Generic.List[string]
$script:c12HarnessCleanupDeadlineTicks = [long]0
$script:c12HarnessIDs = @{
  postgres = (('1' * 64) -join '')
  redis = (('2' * 64) -join '')
  nats = (('3' * 64) -join '')
}
$script:c12HarnessImageIDs = @{
  postgres = ('sha256:' + (('a' * 64) -join ''))
  redis = ('sha256:' + (('b' * 64) -join ''))
  nats = ('sha256:' + (('c' * 64) -join ''))
}

function Resolve-C12ImageIdentity {
  param([object]$Resource)
  $Resource.ImageID = [string]$script:c12HarnessImageIDs[[string]$Resource.Kind]
}

function Start-C12Container {
  param([object]$Resource, [string]$RunSuffix, [string]$DatabaseName, [string]$DatabasePassword)
  $Resource.ID = [string]$script:c12HarnessIDs[[string]$Resource.Kind]
  $script:c12HarnessResources[[string]$Resource.ID] = [pscustomobject]@{
    ID = [string]$Resource.ID
    Name = [string]$Resource.Name
    RunSuffix = $RunSuffix
    Kind = [string]$Resource.Kind
    ImageRef = [string]$Resource.ImageRef
    ImageID = [string]$Resource.ImageID
  }
}

function Wait-C12Dependencies { param([object[]]$Resources) }

function Invoke-C12Go {
  param(
    [string[]]$Arguments,
    [string]$Stage,
    [TimeSpan]$Timeout = [TimeSpan]::FromMinutes(2),
    [DateTime]$Deadline = [DateTime]::MaxValue,
    [switch]$AllowFailure
  )
  throw 'forced group deadline exhaustion before cleanup'
}

function Invoke-C12Docker {
  param(
    [string[]]$Arguments,
    [string]$Stage,
    [TimeSpan]$Timeout = [TimeSpan]::FromSeconds(15),
    [DateTime]$Deadline = [DateTime]::MaxValue,
    [switch]$AllowFailure
  )
  if ($Deadline -eq [DateTime]::MaxValue -and $script:c12NativeDeadline -ne [DateTime]::MaxValue) {
    $Deadline = $script:c12NativeDeadline
  }
  if ($Deadline -le [DateTime]::UtcNow) {
    throw "$Stage inherited the expired group deadline"
  }
  if ($script:c12HarnessCleanupDeadlineTicks -eq 0) {
    $cleanupSeconds = ($Deadline - [DateTime]::UtcNow).TotalSeconds
    if ($cleanupSeconds -lt 70 -or $cleanupSeconds -gt 76) {
      throw "$Stage received cleanup grace $cleanupSeconds seconds, want approximately 75"
    }
    $script:c12HarnessCleanupDeadlineTicks = $Deadline.Ticks
  }
  elseif ($script:c12HarnessCleanupDeadlineTicks -ne $Deadline.Ticks) {
    throw "$Stage did not share the one cleanup deadline"
  }
  $script:c12HarnessCalls.Add(($Arguments -join ' '))
  $containerID = [string]$Arguments[-1]
  if (-not $script:c12HarnessResources.ContainsKey($containerID)) {
    throw "$Stage targeted an uncaptured container"
  }
  $resource = $script:c12HarnessResources[$containerID]
  if ($Arguments.Count -ge 2 -and $Arguments[0] -ceq 'container' -and $Arguments[1] -ceq 'inspect') {
    if ($Arguments -contains '{{.Id}}') {
      return [pscustomobject]@{ ExitCode = 1; Output = @() }
    }
    $identity = "$($resource.ID)|/$($resource.Name)|$($resource.RunSuffix)|$($resource.Kind)|$($resource.ImageRef)|$($resource.ImageID)"
    return [pscustomobject]@{ ExitCode = 0; Output = @($identity) }
  }
  return [pscustomobject]@{ ExitCode = 0; Output = @() }
}

try {
  $groupFailed = $false
  try {
    Invoke-C12Group -GroupID 'deadline-cleanup' -GroupProfile 'base' -Package './internal/store' -GroupTimeout '3m' -AbsoluteDeadline ([DateTime]::UtcNow.AddSeconds(-1))
  }
  catch {
    $groupFailed = $_.Exception.Message -ceq 'C12 group deadline-cleanup failed'
  }
  if (-not $groupFailed) { throw 'expired group did not preserve its primary failure' }
  if ($script:c12NativeDeadline.Ticks -ne $script:c12HarnessPriorDeadline.Ticks) { throw 'group cleanup did not restore the finite prior native deadline' }
  if ($script:c12HarnessCalls.Count -ne 12) { throw "exact cleanup call count=$($script:c12HarnessCalls.Count), want 12" }
  $expectedIDs = @($script:c12HarnessIDs.nats, $script:c12HarnessIDs.redis, $script:c12HarnessIDs.postgres)
  for ($resourceIndex = 0; $resourceIndex -lt $expectedIDs.Count; $resourceIndex++) {
    $offset = $resourceIndex * 4
    $containerID = [string]$expectedIDs[$resourceIndex]
    if (-not $script:c12HarnessCalls[$offset].StartsWith('container inspect --format ') -or
        -not $script:c12HarnessCalls[$offset].EndsWith(" $containerID")) {
      throw "cleanup identity call $resourceIndex was not bound to $containerID"
    }
    if ($script:c12HarnessCalls[$offset + 1] -cne "container stop --time 2 $containerID") {
      throw "cleanup stop call $resourceIndex was not bound to $containerID"
    }
    if ($script:c12HarnessCalls[$offset + 2] -cne "container rm $containerID") {
      throw "cleanup remove call $resourceIndex was not bound to $containerID"
    }
    if (-not $script:c12HarnessCalls[$offset + 3].StartsWith('container inspect --format {{.Id}} ') -or
        -not $script:c12HarnessCalls[$offset + 3].EndsWith(" $containerID")) {
      throw "cleanup absence call $resourceIndex was not bound to $containerID"
    }
  }
  exit 0
}
catch { [Console]::Error.WriteLine($_.Exception.Message); exit 1 }
`, rootPayload)
	harness := filepath.Join(t.TempDir(), "assert-c12-expired-group-cleanup.ps1")
	if err := os.WriteFile(harness, append(append([]byte(nil), runner[:index]...), []byte(appendix)...), 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", harness,
		"-Profile", "base", "-Packages", "./internal/store", "-Timeout", "3m")
	output, commandErr := command.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("expired-group cleanup harness timed out: %v\n%s", ctx.Err(), output)
	}
	if commandErr == nil {
		return string(output), 0
	}
	var exitError *exec.ExitError
	if !errors.As(commandErr, &exitError) {
		t.Fatalf("launch expired-group cleanup harness: %v", commandErr)
	}
	return string(output), exitError.ExitCode()
}

func runC12DependencyDeadlineHarness(t *testing.T, redisPort, natsPort int) (string, int) {
	t.Helper()
	runner, err := os.ReadFile("../../scripts/run-c12-integration.ps1")
	if err != nil {
		t.Fatal(err)
	}
	marker := []byte("$script:c12RepositoryRoot = (Resolve-Path")
	index := bytes.Index(runner, marker)
	if index < 0 {
		t.Fatal("runner lacks main-program marker")
	}
	appendix := fmt.Sprintf(`
$script:c12ObservedProbeDeadline = [DateTime]::MaxValue
$script:c12HarnessEnclosingDeadline = [DateTime]::UtcNow.AddSeconds(1)
$script:c12NativeDeadline = $script:c12HarnessEnclosingDeadline
function Invoke-C12Docker {
  param(
    [string[]]$Arguments,
    [string]$Stage,
    [TimeSpan]$Timeout = [TimeSpan]::FromSeconds(15),
    [DateTime]$Deadline = [DateTime]::MaxValue,
    [switch]$AllowFailure
  )
  if ($Deadline -eq [DateTime]::MaxValue) { throw 'PostgreSQL health inspect did not receive the shared readiness deadline' }
  if ($Deadline.Ticks -ne $script:c12HarnessEnclosingDeadline.Ticks) { throw 'PostgreSQL health inspect did not use the earlier enclosing deadline' }
  $script:c12ObservedProbeDeadline = $Deadline
  return [pscustomobject]@{ ExitCode = 0; Output = @('healthy') }
}
$resources = @(
  [pscustomobject]@{ Kind = 'postgres'; ID = (('1' * 64) -join ''); Port = 1 },
  [pscustomobject]@{ Kind = 'redis'; ID = (('2' * 64) -join ''); Port = %d },
  [pscustomobject]@{ Kind = 'nats'; ID = (('3' * 64) -join ''); Port = %d }
)
try {
  $failure = ''
  $probeWatch = [Diagnostics.Stopwatch]::StartNew()
  try { Wait-C12Dependencies -Resources $resources -ProbeTimeout ([TimeSpan]::FromSeconds(2)) }
  catch { $failure = $_.Exception.Message }
  $probeWatch.Stop()
  if (-not $failure.Contains('C12 dependencies did not pass health and protocol probes within 2 seconds')) {
    throw "unexpected readiness failure: $failure"
  }
  if ($probeWatch.Elapsed -lt [TimeSpan]::FromMilliseconds(500)) { throw 'readiness returned before exercising the stalled protocol probe' }
  if ($probeWatch.Elapsed -gt [TimeSpan]::FromSeconds(5)) { throw "readiness exceeded its bounded enclosing deadline: $($probeWatch.Elapsed)" }
  if ($script:c12ObservedProbeDeadline -eq [DateTime]::MaxValue) { throw 'readiness never inspected PostgreSQL with a deadline' }
  if ($script:c12NativeDeadline.Ticks -ne $script:c12HarnessEnclosingDeadline.Ticks) { throw 'readiness changed the enclosing native deadline' }
  exit 0
}
catch { [Console]::Error.WriteLine($_.Exception.Message); exit 1 }
`, redisPort, natsPort)
	harness := filepath.Join(t.TempDir(), "assert-c12-dependency-deadline.ps1")
	if err := os.WriteFile(harness, append(append([]byte(nil), runner[:index]...), []byte(appendix)...), 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", harness,
		"-Profile", "base", "-Packages", "./internal/store", "-Timeout", "3m")
	output, commandErr := command.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("dependency deadline harness timed out: %v\n%s", ctx.Err(), output)
	}
	if commandErr == nil {
		return string(output), 0
	}
	var exitError *exec.ExitError
	if !errors.As(commandErr, &exitError) {
		t.Fatalf("launch dependency deadline harness: %v", commandErr)
	}
	return string(output), exitError.ExitCode()
}

func runC12PreparedArtifactHarness(t *testing.T, mode string) (string, int) {
	t.Helper()
	runner, err := os.ReadFile("../../scripts/run-c12-integration.ps1")
	if err != nil {
		t.Fatal(err)
	}
	marker := []byte("$script:c12RepositoryRoot = (Resolve-Path")
	index := bytes.Index(runner, marker)
	if index < 0 {
		t.Fatal("runner lacks main-program marker")
	}
	fixtureRoot := t.TempDir()
	fixturePayload := base64.StdEncoding.EncodeToString([]byte(fixtureRoot))
	appendix := fmt.Sprintf(`
$mode = '%s'
$fixtureRoot = [Text.Encoding]::UTF8.GetString([Convert]::FromBase64String('%s'))
$runSuffix = [Guid]::NewGuid().ToString('N')
$artifact = $null
$prepared = $null
$artifactPath = ''
$executionMarker = Join-Path $fixtureRoot 'prepared-target-executed.txt'
$success = $false
try {
  $artifact = New-C12PreparedArtifactRoot -RunSuffix $runSuffix -Profile 'authority-v7-pitr'
  $artifactPath = [string]$artifact.Root
  $sourceExecutable = [IO.Path]::GetFullPath((Join-Path $env:SystemRoot 'System32\cmd.exe'))
  $executablePrefix = if ($mode -ceq 'initializer-capabilities') { 'authority-test2json' } else { 'trusted-validator' }
  $executablePath = Join-Path $artifactPath "$executablePrefix-$([Guid]::NewGuid().ToString('N')).exe"
  $null = Register-C12DirectLeafIntent -Ledger $artifact.Ledger -Name ([IO.Path]::GetFileName($executablePath)) -Kind 'exact_file' -Expected $true
  [IO.File]::Copy($sourceExecutable, $executablePath, $false)
  Protect-C12PrivateArtifactFile -Path $executablePath
  $null = Bind-C12DirectLeaf -ArtifactRoot $artifact -Name ([IO.Path]::GetFileName($executablePath))
  $sourceDigest = Get-C12SHA256Hex -Bytes ([Text.Encoding]::UTF8.GetBytes('prepared-artifact-fixture/v1'))
  $toolchain = [pscustomobject]@{
    Path = $sourceExecutable
    SHA256 = (Get-C12SHA256Hex -Bytes ([IO.File]::ReadAllBytes($sourceExecutable)))
    Version = 'prepared-artifact-fixture/v1'
    GOOS = 'windows'
    GOARCH = 'amd64'
  }
  $arguments = @('/d', '/c', ('echo C12_TARGET_EXECUTED>"{0}" & echo C12_PREPARED_FIXTURE_OK' -f $executionMarker))
  $invocationRole = 'trusted-validator'
  $invocationCapabilities = @{}
  $secondaryExecutablePath = ''
  if ($mode -ceq 'initializer-capabilities') {
    $invocationRole = 'authority-initializer-json'
    $secondaryExecutablePath = Join-Path $artifactPath "authority-initializer-$([Guid]::NewGuid().ToString('N')).test.exe"
    $null = Register-C12DirectLeafIntent -Ledger $artifact.Ledger -Name ([IO.Path]::GetFileName($secondaryExecutablePath)) -Kind 'exact_file' -Expected $true
    [IO.File]::Copy($sourceExecutable, $secondaryExecutablePath, $false)
    Protect-C12PrivateArtifactFile -Path $secondaryExecutablePath
    $null = Bind-C12DirectLeaf -ArtifactRoot $artifact -Name ([IO.Path]::GetFileName($secondaryExecutablePath))
    $invocationCapabilities = @{
      TALENRO_C12_AUTHORITY_V7_INIT_NONCE = 'fixture-nonce'
      TALENRO_C12_AUTHORITY_V7_RUN_SUFFIX = $runSuffix
      TALENRO_C12_AUTHORITY_V7_PROFILE = 'authority-v7-pitr'
      TALENRO_DATABASE_URL = 'fixture-database'
      TALENRO_INSTALLATION_KIND = 'disposable_fixture'
    }
    [Environment]::SetEnvironmentVariable('TALENRO_C12_AUTHORITY_PITR_HMAC_KEY','CONTROLLER_WAL_KEY_CANARY')
    $arguments[2] = ('echo C12_TARGET_EXECUTED>"{0}" & set TALENRO_C12_AUTHORITY_V7_INIT_NONCE & set TALENRO_C12_AUTHORITY_PITR_HMAC_KEY & set C12_BOOTSTRAP_KEY & echo C12_PREPARED_FIXTURE_OK' -f $executionMarker)
  }
  $receipt = New-C12SealedExecutableReceipt -ArtifactRoot $artifact -Role $invocationRole -Profile 'authority-v7-pitr' -Purpose 'prepared-artifact-fixture' -SourceIdentity 'prepared-artifact-fixture/v1' -SourceDigest $sourceDigest -CandidateTreeIdentity ('f' * 40) -BuildArguments @('fixture', 'copy') -GoToolchain $toolchain -ExecutablePath $executablePath -SecondaryExecutablePath $secondaryExecutablePath -Arguments $arguments -WorkingDirectory $fixtureRoot

  if ($mode.StartsWith('protocol-') -or $mode.StartsWith('deadline-')) {
    if ($null -eq (Get-Command Invoke-C12PreparedProtocol -ErrorAction SilentlyContinue)) { throw 'authenticated prepared protocol is absent' }
    if ((ConvertTo-C12ProtocolJSON ([ordered]@{entries=@()})) -cne '{"entries":[]}') { throw 'canonical JSON changed an empty array into null' }
    if ((ConvertTo-C12ProtocolJSON ([ordered]@{z=1;entries=@([ordered]@{value='y';name='x'})})) -cne '{"entries":[{"name":"x","value":"y"}],"z":1}') { throw 'canonical JSON lost singleton array or ordinal nested-key ordering' }
    $prepared = New-C12PreparedNativeWorker -ArtifactRoot $artifact -Receipt $receipt -SetupDeadline ([DateTime]::UtcNow.AddSeconds(45))
    if ($mode.StartsWith('deadline-')) {
      $script:deadlineSend = (Get-Command Send-C12PreparedFrame).ScriptBlock
      $script:deadlineMutationApplied = $false
      function Send-C12PreparedFrame {
        param($State, $Pipe, $Deadline, $Type, $Payload)
        if ($Type -cne 'VERIFICATION_PROJECTION') { & $script:deadlineSend @PSBoundParameters; return }
        if (@($Payload.Keys) -cnotcontains 'invocation_deadline_ticks') { throw 'authenticated invocation deadline is absent' }
        $script:deadlineMutationApplied = $true
        if ($mode -ceq 'deadline-wire-tamper') {
          [byte[]]$wire = New-C12PreparedFrame -State $State -Type $Type -Payload $Payload
          $null = Assert-C12PreparedFrame -State $State -Bytes $wire -Type $Type
          $frame = [Text.Encoding]::UTF8.GetString($wire) | ConvertFrom-Json
          $frame.payload.invocation_deadline_ticks = [DateTime]::UtcNow.AddHours(1).Ticks
          # Change the authenticated deadline on the wire without re-signing.
          [C12PreparedPipe]::WriteMessage($Pipe, [Text.Encoding]::UTF8.GetBytes((ConvertTo-C12ProtocolJSON $frame)), (Get-C12ProtocolMilliseconds $Deadline))
          return
        }
        if ($mode -ceq 'deadline-malformed') { $Payload.invocation_deadline_ticks = 'tomorrow' }
        else { $Payload.invocation_deadline_ticks = [DateTime]::UtcNow.AddSeconds(-1).Ticks }
        # A correctly signed but invalid deadline must also fail in the worker.
        & $script:deadlineSend @PSBoundParameters
      }
    }
    $script:protocolRead = (Get-Command Read-C12PreparedMessage).ScriptBlock
    $script:protocolReadCount = 0
    $script:protocolHello = $null
    function Read-C12PreparedMessage {
      param($Pipe, $Deadline)
      [byte[]]$wire = & $script:protocolRead @PSBoundParameters
      $script:protocolReadCount++
      if ($script:protocolReadCount -eq 1) { $script:protocolHello = $wire }
      if ($mode -ceq 'protocol-duplicate-sequence') {
        if ($script:protocolReadCount -eq 2) { return ,$script:protocolHello }
        return ,$wire
      }
      if ($script:protocolReadCount -ne 1) { return ,$wire }
      if ($mode -ceq 'protocol-oversized-frame') { return ,([byte[]]::new(131073)) }
      if ($mode -ceq 'protocol-invalid-utf8') { return ,([byte[]]@(0xc3, 0x28)) }
      if ($mode -ceq 'protocol-noncanonical-json') { return ,([Text.Encoding]::UTF8.GetBytes(' ' + [Text.Encoding]::UTF8.GetString($wire))) }
      $frame = [Text.Encoding]::UTF8.GetString($wire) | ConvertFrom-Json
      switch ($mode) {
        'protocol-reordered-frame' { $frame.type = 'GATE_WAITING' }
        'protocol-wrong-nonce' { $frame.worker_nonce = 'foreign-worker' }
        'protocol-previous-digest' { $frame.previous_frame_digest = ('f' * 64) }
        'protocol-bad-hmac' { $frame.hmac_sha256 = ('0' * 64) }
        'protocol-bad-hmac-cleanup' { $frame.hmac_sha256 = ('0' * 64) }
      }
      return ,([Text.Encoding]::UTF8.GetBytes(($frame | ConvertTo-Json -Compress -Depth 30)))
    }
    if ($mode -ceq 'protocol-foreign-pid') {
      function Assert-C12PreparedClientPID {
        param($Prepared, $Pipe)
        [C12PreparedPipe]::VerifyClientProcessId($Pipe, [uint32]($Prepared.ProcessId + 1))
      }
    }
    if ($mode -ceq 'protocol-bad-hmac-cleanup') {
      $script:protocolClose = (Get-Command Close-C12PreparedNativeWorker).ScriptBlock
      $script:protocolCleanupInjected = $false
      function Close-C12PreparedNativeWorker {
        param($Prepared, $Deadline)
        & $script:protocolClose @PSBoundParameters
        if (-not $script:protocolCleanupInjected) { $script:protocolCleanupInjected = $true; throw 'injected protocol cleanup failure' }
      }
    }
  $rejected = $false
  $expectedFailure = @{
    'protocol-duplicate-sequence' = 'sequence/type replay or order mismatch'
    'protocol-reordered-frame' = 'sequence/type replay or order mismatch'
    'protocol-foreign-pid' = 'client PID mismatch'
    'protocol-wrong-nonce' = 'schema/run/worker nonce mismatch'
    'protocol-oversized-frame' = 'frame exceeds maximum size'
    'protocol-previous-digest' = 'previous frame digest mismatch'
    'protocol-bad-hmac' = 'frame HMAC mismatch'
    'protocol-bad-hmac-cleanup' = 'frame HMAC mismatch'
    'protocol-invalid-utf8' = 'not valid UTF-8 JSON'
    'protocol-noncanonical-json' = 'not canonical JSON'
    'deadline-wire-tamper' = 'protocol frame exceeds message bound or is truncated'
    'deadline-malformed' = 'protocol frame exceeds message bound or is truncated'
    'deadline-expired' = 'protocol frame exceeds message bound or is truncated'
  }
  $rejectionDeadline = [DateTime]::UtcNow.AddSeconds(20)
  if ($mode.StartsWith('deadline-')) { $rejectionDeadline = [DateTime]::UtcNow.AddSeconds(60) }
  try { $null = Invoke-C12PreparedAuthorityRole -Prepared $prepared -Role 'trusted-validator' -Timeout ([TimeSpan]::FromMinutes(2)) -Deadline $rejectionDeadline }
  catch { if ($_.Exception.ToString().Contains($expectedFailure[$mode])) { $rejected = $true } else { throw } }
    if (-not $rejected) { throw "$mode did not reject its malformed protocol input" }
    if ($mode.StartsWith('deadline-') -and -not $script:deadlineMutationApplied) { throw 'deadline tamper was never sent to the real worker' }
    if ($mode -ceq 'protocol-bad-hmac-cleanup' -and $prepared.CleanupFailures -cnotcontains 'injected protocol cleanup failure') { throw 'protocol cleanup failure was not retained alongside the primary HMAC failure' }
    if ($prepared.Released -or $prepared.GateSignaled -or [IO.File]::Exists($executionMarker)) { throw "$mode released or executed the payload" }
    Close-C12PreparedNativeWorker -Prepared $prepared -Deadline ([DateTime]::UtcNow.AddSeconds(10))
    if (-not $prepared.ActiveProcessZeroConfirmed -or -not $prepared.Closed -or (Get-Process -Id $prepared.ProcessId -ErrorAction SilentlyContinue)) { throw "$mode did not converge the exact process and Job" }
    $success = $true
  }
  elseif ($mode -cin @('prepared-suspended-membership', 'prepared-assignment-failure', 'prepared-pid-mismatch', 'prepared-invalid-thread-handle', 'prepared-cleanup-retry', 'prepared-job-collision', 'prepared-exit-race', 'prepared-exit-pending', 'prepared-exit-pending-retry')) {
    if ($null -eq (Get-Command New-C12SuspendedPreparedWorker -ErrorAction SilentlyContinue)) {
      throw 'production suspended process factory is absent'
    }
    Add-Type -TypeDefinition @'
using System;
using System.ComponentModel;
using System.Reflection;
using System.Runtime.InteropServices;
public static class C12ContainmentProbe {
  [DllImport("kernel32.dll", SetLastError=true)] public static extern uint GetProcessId(IntPtr process);
  [DllImport("kernel32.dll", SetLastError=true)] public static extern uint GetProcessIdOfThread(IntPtr thread);
  [DllImport("kernel32.dll", SetLastError=true)] public static extern bool IsProcessInJob(IntPtr process, IntPtr job, out bool member);
  [DllImport("kernel32.dll", SetLastError=true)] static extern bool DuplicateHandle(IntPtr source, IntPtr handle, IntPtr target, out IntPtr duplicate, uint access, bool inherit, uint options);
  [DllImport("kernel32.dll")] static extern IntPtr GetCurrentProcess();
  [DllImport("kernel32.dll", SetLastError=true)] static extern bool CloseHandle(IntPtr handle);
  public static void DenyAssignment(object controller) {
    FieldInfo field = controller.GetType().GetField("jobHandle", BindingFlags.Instance | BindingFlags.NonPublic);
    if (field == null) throw new Exception("controller has no retained Job handle");
    IntPtr original = (IntPtr)field.GetValue(controller), restricted;
    // Keep QUERY, TERMINATE and SYNCHRONIZE, but remove ASSIGN_PROCESS.
    if (!DuplicateHandle(GetCurrentProcess(), original, GetCurrentProcess(), out restricted, 0x001F003E, false, 0))
      throw new Win32Exception(Marshal.GetLastWin32Error());
    field.SetValue(controller, restricted);
    if (!CloseHandle(original)) throw new Win32Exception(Marshal.GetLastWin32Error());
  }
  public static IntPtr RemoveThreadHandle(object controller) {
    FieldInfo field = controller.GetType().GetField("primaryThreadHandle", BindingFlags.Instance | BindingFlags.NonPublic);
    IntPtr original = (IntPtr)field.GetValue(controller);
    field.SetValue(controller, IntPtr.Zero);
    return original;
  }
  public static void DenyProcessTermination(object controller) {
    var field = controller.GetType().GetField("processHandle", BindingFlags.Instance | BindingFlags.NonPublic);
    IntPtr original = (IntPtr)field.GetValue(controller), restricted;
    // Retain exact-process query/synchronization, but no PROCESS_TERMINATE.
    // This deterministically produces the ambiguous Win32 5 while the suspended
    // process is unsignaled; only the real owned Job can then end the process.
    if (!DuplicateHandle(GetCurrentProcess(), original, GetCurrentProcess(), out restricted, 0x00101400, false, 0))
      throw new Win32Exception(Marshal.GetLastWin32Error());
    field.SetValue(controller, restricted);
    if (!CloseHandle(original)) throw new Win32Exception(Marshal.GetLastWin32Error());
  }
  public static void RestoreThreadHandle(object controller, IntPtr original) {
    controller.GetType().GetField("primaryThreadHandle", BindingFlags.Instance | BindingFlags.NonPublic).SetValue(controller, original);
  }
  public static void ChangeExpectedPID(object controller) {
    controller.GetType().GetProperty("ProcessId").GetSetMethod(true).Invoke(controller, new object[] { (uint)1 });
  }
  public static void TerminateExitedProcess(object controller) {
    // Reproduce the exact native boundary after the process exits between the
    // outer wait check and TerminateProcess. No forged handles or native mocks.
    var method = controller.GetType().GetMethod("TerminateOwnedProcess", BindingFlags.Instance | BindingFlags.NonPublic);
    if (method == null) throw new Exception("production exact-handle termination boundary is absent");
    method.Invoke(controller, new object[] { (uint)1 });
  }
  public static void CloseTestHandle(IntPtr handle) { if (!CloseHandle(handle)) throw new Win32Exception(Marshal.GetLastWin32Error()); }
}
'@
    if ($mode -ceq 'prepared-job-collision') {
      $collisionJob = [C12NativeJob]::CreateKillOnClose([string]$receipt.NativeJobName)
      try {
        $collisionRejected = $false
        try { $prepared = New-C12PreparedNativeWorker -ArtifactRoot $artifact -Receipt $receipt -SetupDeadline ([DateTime]::UtcNow.AddSeconds(20)) }
        catch { $collisionRejected = $_.Exception.ToString() -match 'collision' }
        if (-not $collisionRejected -or $script:c12PreparedWorkers.Count -ne 0) { throw 'prepared worker adopted a colliding native Job' }
        if ([IO.File]::Exists($executionMarker)) { throw 'colliding native Job started the worker' }
      }
      finally { [C12NativeJob]::Close($collisionJob) }
    }
    elseif ($mode -cin @('prepared-assignment-failure', 'prepared-pid-mismatch', 'prepared-invalid-thread-handle')) {
      $script:containmentRealFactory = (Get-Command New-C12SuspendedProcessController).ScriptBlock
      $script:containmentFailedController = $null
      function New-C12SuspendedProcessController {
        param($Executable, $Arguments, $WorkingDirectory, $Environment, $JobName)
        $controller = & $script:containmentRealFactory @PSBoundParameters
        $script:containmentFailedController = $controller
        $script:containmentFailedPID = $controller.ProcessId
        if ($mode -ceq 'prepared-assignment-failure') { [C12ContainmentProbe]::DenyAssignment($controller) }
        elseif ($mode -ceq 'prepared-pid-mismatch') { [C12ContainmentProbe]::ChangeExpectedPID($controller) }
        else { $script:containmentRemovedThread = [C12ContainmentProbe]::RemoveThreadHandle($controller) }
        return $controller
      }
      $assignmentRejected = $false
      try { $prepared = New-C12PreparedNativeWorker -ArtifactRoot $artifact -Receipt $receipt -SetupDeadline ([DateTime]::UtcNow.AddSeconds(20)) }
      catch { $assignmentRejected = $_.Exception.ToString() -match 'AssignProcessToJobObject|PID/handle mismatch|primary thread handle is zero or invalid' }
      if (-not $assignmentRejected -or $null -eq $script:containmentFailedController) { throw 'forced native assignment failure was not preserved' }
      $prepared = @($script:c12PreparedWorkers | Select-Object -Last 1)[0]
      if ($mode -ceq 'prepared-invalid-thread-handle') {
        if (-not $prepared.Closed) {
          # Restore test-owned state so even a RED run can clean its process.
          [C12ContainmentProbe]::RestoreThreadHandle($prepared.Controller, $script:containmentRemovedThread)
        }
        else { [C12ContainmentProbe]::CloseTestHandle($script:containmentRemovedThread) }
      }
      if (-not $prepared.Closed -or $prepared.Phase -cne 'Removed' -or -not $prepared.ActiveProcessZeroConfirmed) { throw 'assignment failure did not converge its retained worker record' }
      if (Get-Process -Id $script:containmentFailedPID -ErrorAction SilentlyContinue) { throw 'assignment-failed suspended worker PID survived cleanup' }
      if (-not $prepared.Controller.ActiveProcessZeroConfirmed) { throw 'assignment-failed Job did not confirm zero active processes' }
      if ([IO.File]::Exists($executionMarker)) { throw 'assignment-failed worker executed its sentinel' }
    }
    else {
      $containmentSetupDeadline = [DateTime]::UtcNow.AddSeconds(20)
      if ($mode.StartsWith('prepared-exit-')) { $containmentSetupDeadline = [DateTime]::UtcNow.AddSeconds(60) }
      $prepared = New-C12PreparedNativeWorker -ArtifactRoot $artifact -Receipt $receipt -SetupDeadline $containmentSetupDeadline
      foreach ($handle in @('ProcessHandle', 'PrimaryThreadHandle', 'JobHandle', 'CompletionPortHandle')) {
        if ($prepared.PSObject.Properties.Name -cnotcontains $handle -or [IntPtr]$prepared.$handle -eq [IntPtr]::Zero -or [IntPtr]$prepared.$handle -eq [IntPtr](-1)) { throw "prepared worker lacks retained $handle" }
      }
      if ($prepared.Phase -cne 'MembershipVerified' -or $prepared.Controller.Phase -cne 'MembershipVerified') { throw 'prepared worker did not stop at MembershipVerified' }
      if ([C12ContainmentProbe]::GetProcessId($prepared.ProcessHandle) -ne $prepared.ProcessId -or [C12ContainmentProbe]::GetProcessIdOfThread($prepared.PrimaryThreadHandle) -ne $prepared.ProcessId) { throw 'retained handles do not bind the worker PID' }
      $member = $false
      if (-not [C12ContainmentProbe]::IsProcessInJob($prepared.ProcessHandle, $prepared.JobHandle, [ref]$member) -or -not $member) { throw 'prepared process is not in the exact controller Job' }
      Start-Sleep -Milliseconds 300
      if ([IO.File]::Exists($executionMarker)) { throw 'production prepared worker executed before release' }
      if ($mode -cin @('prepared-exit-pending','prepared-exit-pending-retry')) {
        [C12ContainmentProbe]::DenyProcessTermination($prepared.Controller)
        $prepared.ProcessHandle = $prepared.Controller.ProcessHandle
      }
      if ($mode -cin @('prepared-cleanup-retry','prepared-exit-pending-retry')) {
        $retainedProcess = $prepared.ProcessHandle
        $retainedJob = $prepared.JobHandle
        $retainedPort = $prepared.CompletionPortHandle
        $retainedThread = $prepared.PrimaryThreadHandle
        $timedOut = $false
        try { Close-C12PreparedNativeWorker -Prepared $prepared -Deadline ([DateTime]::UtcNow.AddSeconds(-1)) }
        catch { $timedOut = $_.Exception.Message -match 'confirmation|deadline' }
        if (-not $timedOut -or $prepared.Closed -or $prepared.Phase -cne 'CleanIntent') { throw 'cleanup timeout erased retryable worker state' }
        if ($prepared.ProcessHandle -ne $retainedProcess -or $prepared.JobHandle -ne $retainedJob -or $prepared.CompletionPortHandle -ne $retainedPort -or $prepared.PrimaryThreadHandle -ne $retainedThread) { throw 'cleanup timeout disposed retained native handles' }
        if ($prepared.ActiveProcessZeroConfirmed) { throw 'expired cleanup asserted unobserved convergence' }
        if ($mode -ceq 'prepared-exit-pending-retry' -and -not $prepared.Controller.ProcessExitPendingOnly) { throw 'pending-exit failure was not retained until exact convergence' }
      }
      Close-C12PreparedNativeWorker -Prepared $prepared -Deadline ([DateTime]::UtcNow.AddSeconds(10))
      if (-not $prepared.Closed -or -not $prepared.ActiveProcessZeroConfirmed -or $prepared.Phase -cne 'Removed') { throw 'prepared worker cleanup did not converge' }
      if ($mode -ceq 'prepared-exit-pending' -and -not $prepared.Controller.ProcessExitPendingOnly) { throw 'pending-exit convergence branch was not exercised' }
      if (Get-Process -Id $prepared.ProcessId -ErrorAction SilentlyContinue) { throw 'contained worker PID survived cleanup' }

      if ($mode -cin @('prepared-suspended-membership','prepared-exit-race')) {
        # The primitive positive control is test-only. Production stops at MembershipVerified.
        $environment = [Collections.Generic.Dictionary[string,string]]::new([StringComparer]::OrdinalIgnoreCase)
        foreach ($entry in [Environment]::GetEnvironmentVariables().GetEnumerator()) { $environment[[string]$entry.Key] = [string]$entry.Value }
        $controller = [C12SuspendedProcessController]::Create($executablePath, ('/d /c echo C12_TARGET_EXECUTED>"' + $executionMarker + '"'), $fixtureRoot, $environment)
        try {
          $earlyDisposeRejected = $false
          try { $controller.Dispose() } catch { $earlyDisposeRejected = $true }
          if (-not $earlyDisposeRejected -or $controller.ProcessHandle -eq [IntPtr]::Zero) { throw 'live native controller discarded its handles before exit confirmation' }
          foreach ($phase in @('CreatedSuspended', 'AssignedToJob')) {
            if ($controller.Phase -cne $phase) { throw "native controller skipped $phase" }
            $resumeRejected = $false
            try { $controller.Resume() } catch { $resumeRejected = $true }
            if (-not $resumeRejected) { throw "native controller resumed in $phase" }
            Start-Sleep -Milliseconds 300
            if ([IO.File]::Exists($executionMarker)) { throw "worker executed while $phase" }
            if ($phase -ceq 'CreatedSuspended') { $controller.AssignToJob() }
          }
          $controller.VerifyMembership()
          if ([IO.File]::Exists($executionMarker)) { throw 'worker executed before membership verification' }
          $controller.Resume()
          if (-not $controller.WaitForActiveProcessZero(10000)) { throw 'positive-control Job did not reach active-process-zero' }
          if ($controller.ActiveProcesses -ne 0) { throw 'native Job reported active processes after confirmation' }
          if ($mode -ceq 'prepared-exit-race') {
            [C12ContainmentProbe]::TerminateExitedProcess($controller)
            if ($controller.ActiveProcesses -ne 0 -or -not $controller.ActiveProcessZeroConfirmed) { throw 'exited-process termination lost exact Job convergence' }
          }
          if (-not [IO.File]::Exists($executionMarker) -or ([IO.File]::ReadAllText($executionMarker)).Trim() -cne 'C12_TARGET_EXECUTED') { throw 'native resume positive control did not execute sentinel' }
          $duplicateResumeRejected = $false
          try { $controller.Resume() } catch { $duplicateResumeRejected = $true }
          if (-not $duplicateResumeRejected) { throw 'duplicate native resume was accepted' }
        }
        finally { $controller.Terminate(1); if ($controller.WaitForActiveProcessZero(10000)) { $controller.Dispose() } }
        $disposedRejected = $false
        try { $controller.VerifyMembership() } catch { $disposedRejected = $true }
        if (-not $disposedRejected) { throw 'disposed controller accepted membership verification' }
      }
    }
    $success = $true
  }
  else {

  switch ($mode) {
    'byte-tamper' {
      [IO.File]::AppendAllText($executablePath, 'tamper')
    }
    'file-id-splice' {
      [IO.File]::Move($executablePath, "$executablePath.original")
      [IO.File]::Copy($sourceExecutable, $executablePath, $false)
    }
    'reparse' {
      [IO.File]::Move($executablePath, "$executablePath.original")
      New-Item -ItemType Junction -Path $executablePath -Target $fixtureRoot | Out-Null
    }
    'hard-link' {
      New-Item -ItemType HardLink -Path (Join-Path $artifactPath 'receipt-hard-link.exe') -Target $executablePath | Out-Null
    }
    'dacl' {
      $security = [IO.File]::GetAccessControl($executablePath)
      $everyone = New-Object Security.Principal.SecurityIdentifier('S-1-1-0')
      $security.AddAccessRule((New-Object Security.AccessControl.FileSystemAccessRule($everyone, [Security.AccessControl.FileSystemRights]::Read, [Security.AccessControl.AccessControlType]::Allow)))
      [IO.File]::SetAccessControl($executablePath, $security)
    }
  }

  if ($mode -cin @('byte-tamper', 'file-id-splice', 'reparse', 'hard-link', 'dacl')) {
    $rejected = $false
    try { $prepared = New-C12PreparedNativeWorker -ArtifactRoot $artifact -Receipt $receipt -SetupDeadline ([DateTime]::UtcNow.AddSeconds(45)) }
    catch { $rejected = $true }
    if (-not $rejected) { throw "$mode mutation reached a prepared worker" }
    if ($null -ne $prepared -and (Get-C12PreparedWorkerPhases -Prepared $prepared) -ccontains 'EXEC_BEGIN') { throw "$mode mutation reached EXEC_BEGIN" }
    $success = $true
  }
  else {
    $preparationDeadline = [DateTime]::UtcNow.AddSeconds(45)
    $prepared = New-C12PreparedNativeWorker -ArtifactRoot $artifact -Receipt $receipt -SetupDeadline $preparationDeadline
    if ($mode -ceq 'delayed-preparation-release') {
      # Real elapsed-time regression: no fake clock, rewritten bootstrap bytes,
      # or skipped sleep. The old embedded deadline must expire while suspended.
      $staleDeadline = $preparationDeadline.AddMinutes(2)
      while ([DateTime]::UtcNow -le $staleDeadline.AddMilliseconds(250)) { Start-Sleep -Milliseconds 100 }
      if ($prepared.Released -or [IO.File]::Exists($executionMarker)) { throw 'delayed preparation released a suspended payload' }
      $script:delayedRelease = (Get-Command Release-C12PreparedWorker).ScriptBlock
      function Release-C12PreparedWorker {
        param($Prepared)
        Start-Sleep -Milliseconds 500
        & $script:delayedRelease @PSBoundParameters
      }
    }
    if ($mode -ceq 'wrong-role') {
      $rejected = $false
      try { $null = Invoke-C12PreparedAuthorityRole -Prepared $prepared -Role 'authority-initializer-json' -Timeout ([TimeSpan]::FromMinutes(2)) -Deadline ([DateTime]::UtcNow.AddMinutes(2)) -CapabilityEnvironment @{} }
      catch { $rejected = $_.Exception.Message.Contains('wrong role') }
      if (-not $rejected) { throw 'wrong-role prepared invocation was accepted' }
      if ((Get-C12PreparedWorkerPhases -Prepared $prepared) -ccontains 'EXEC_BEGIN') { throw 'wrong-role invocation reached EXEC_BEGIN' }
      $success = $true
    }
    else {
      $invocationDeadline = [DateTime]::UtcNow.AddMinutes(2)
      if ($mode -ceq 'delayed-preparation-release') { $invocationDeadline = [DateTime]::UtcNow.AddSeconds(60) }
      $result = Invoke-C12PreparedAuthorityRole -Prepared $prepared -Role $invocationRole -Timeout ([TimeSpan]::FromMinutes(2)) -Deadline $invocationDeadline -CapabilityEnvironment $invocationCapabilities
      if ($mode -ceq 'delayed-preparation-release' -and ($prepared.ReleaseTicks -le $staleDeadline.Ticks -or $prepared.InvocationDeadline.Ticks -ne $invocationDeadline.Ticks -or -not $prepared.ExecutionComplete)) { throw 'delayed invocation did not complete authenticated EXEC_END after the stale preparation deadline' }
      if ($mode -ceq 'initializer-capabilities') {
        if (@($result.Output | Where-Object { $_ -ceq 'TALENRO_C12_AUTHORITY_V7_INIT_NONCE=fixture-nonce' }).Count -ne 1) { throw 'initializer did not receive its authenticated capability environment' }
        if (@($result.Output | Where-Object { $_ -match 'CONTROLLER_WAL_KEY_CANARY|^C12_BOOTSTRAP_KEY=' }).Count -ne 0) { throw 'initializer inherited a controller or bootstrap key' }
      }
      if (@($result.Output | Where-Object { [string]$_ -ceq 'C12_PREPARED_FIXTURE_OK' }).Count -ne 1) { throw 'prepared fixture output marker mismatch' }
      if (-not [IO.File]::Exists($executionMarker) -or ([IO.File]::ReadAllText($executionMarker)).Trim() -cne 'C12_TARGET_EXECUTED') {
        throw 'prepared target positive-control execution marker mismatch'
      }
      $phases = @(Get-C12PreparedWorkerPhases -Prepared $prepared)
      $expectedProtocol = @('SETUP_BEGIN','BUILD_DONE','ARTIFACT_BOUND','HELLO','VERIFICATION_PROJECTION','JOB_MEMBER_READY','CAPABILITIES','CAPABILITIES_INSTALLED','GATE_WAITING','FORMAL_RELEASE','EXEC_BEGIN','EXEC_END')
      if (($phases -join '|') -cne ($expectedProtocol -join '|')) { throw "production prepared pipe phase order mismatch: $($phases -join '|')" }
      if ($prepared.ProtocolReceipts.Count -ne 6 -or -not $prepared.GateSignaled -or $prepared.ReleaseTicks -le 0 -or -not $prepared.ActiveProcessZeroConfirmed) { throw 'six-frame release lacks verified receipts, gate, time or exact Job convergence' }
      $projectionJSON = ConvertTo-C12ProtocolJSON $prepared.Projection
      foreach ($forbidden in @('ReceiptKeyHex','ReceiptSeal','ReceiptFields','PayloadFields','RootHandle','RefCount','CONTROLLER_WAL_KEY_CANARY')) {
        if ($projectionJSON.Contains($forbidden)) { throw "worker projection leaked $forbidden" }
      }
      if ($projectionJSON.Contains($script:c12PreparedReceiptKeyHex)) { throw 'worker projection leaked receipt signing key bytes' }
      if ([IO.File]::GetLastWriteTimeUtc($executionMarker).Ticks -lt $prepared.ReleaseTicks) { throw 'worker marker timestamp precedes formal release' }
      if ($mode -cin @('replay', 'prepared-pipe-six-frame')) {
        $rejected = $false
        try { $null = Invoke-C12PreparedAuthorityRole -Prepared $prepared -Role 'trusted-validator' -Timeout ([TimeSpan]::FromMinutes(2)) -Deadline ([DateTime]::UtcNow.AddMinutes(2)) -CapabilityEnvironment @{} }
        catch { $rejected = $_.Exception.Message.Contains('receipt replay') }
        if (-not $rejected) { throw 'prepared receipt replay was accepted' }
        if (@(Get-C12PreparedWorkerPhases -Prepared $prepared | Where-Object { $_ -ceq 'EXEC_BEGIN' }).Count -ne 1) { throw 'receipt replay started another executable' }
      }
      $success = $true
    }
  }
  }
}
catch {
  [Console]::Error.WriteLine($_.Exception.ToString())
  if ($null -ne $prepared) { [Console]::Error.WriteLine("C12 prepared diagnostic: Phase=$($prepared.Phase); ExecutionComplete=$($prepared.ExecutionComplete); GateSignaled=$($prepared.GateSignaled)") }
}
finally {
  if ($null -ne $prepared) {
    try { Close-C12PreparedNativeWorker -Prepared $prepared -Deadline ([DateTime]::UtcNow.AddSeconds(15)) }
    catch { [Console]::Error.WriteLine($_.Exception.Message); $success = $false }
  }
  if ($null -ne $artifact) {
    try { Remove-C12PreparedArtifactRoot -ArtifactRoot $artifact -Deadline ([DateTime]::UtcNow.AddSeconds(15)) }
    catch { [Console]::Error.WriteLine($_.Exception.Message); $success = $false }
  }
  if (-not [string]::IsNullOrEmpty($artifactPath) -and [IO.Directory]::Exists($artifactPath)) {
    [Console]::Error.WriteLine('prepared artifact root remained after cleanup')
    $success = $false
  }
}
if ($success) { Write-Output "C12_PREPARED_ARTIFACT_HARNESS_OK:$mode"; exit 0 }
exit 1
`, mode, fixturePayload)
	harness := filepath.Join(t.TempDir(), "assert-c12-prepared-artifact.ps1")
	if err := os.WriteFile(harness, append(append([]byte(nil), runner[:index]...), []byte(appendix)...), 0o600); err != nil {
		t.Fatal(err)
	}
	harnessTimeout := 180 * time.Second
	if strings.HasPrefix(mode, "deadline-") {
		harnessTimeout = 150 * time.Second
	}
	if strings.HasPrefix(mode, "prepared-exit-") {
		harnessTimeout = 150 * time.Second
	}
	if mode == "delayed-preparation-release" {
		harnessTimeout = 270 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), harnessTimeout)
	defer cancel()
	powershellPath := filepath.Join(os.Getenv("SystemRoot"), "System32", "WindowsPowerShell", "v1.0", "powershell.exe")
	command := exec.CommandContext(ctx, powershellPath, "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", harness,
		"-Profile", "base", "-Packages", "./internal/store", "-Timeout", "3m")
	output, commandErr := command.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("prepared artifact %s harness timed out: %v\n%s", mode, ctx.Err(), output)
	}
	if commandErr == nil {
		return string(output), 0
	}
	var exitError *exec.ExitError
	if !errors.As(commandErr, &exitError) {
		t.Fatalf("launch prepared artifact %s harness: %v", mode, commandErr)
	}
	return string(output), exitError.ExitCode()
}

func runC12PreparedGoGraphHarness(t *testing.T) (string, int) {
	t.Helper()
	fixtureRoot := t.TempDir()
	for _, directory := range []string{"internal/testinfra", "src/cmd/test2json"} {
		if err := os.MkdirAll(filepath.Join(fixtureRoot, filepath.FromSlash(directory)), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	for name, body := range map[string]string{
		"go.mod":                         "module example.invalid/c12graph\n\ngo 1.26\n",
		"go.sum":                         "",
		"main.go":                        "package main\nfunc main() {}\n",
		"internal/testinfra/fixture.go":  "package testinfra\n",
		"internal/testinfra/payload.txt": "selected embed bytes\n",
		"src/cmd/test2json/test2json.go": "package main\nfunc main() {}\n",
	} {
		if err := os.WriteFile(filepath.Join(fixtureRoot, filepath.FromSlash(name)), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	fakeGo := filepath.Join(t.TempDir(), "go.exe")
	buildFakeGoExecutable(t, fakeGo, `package main
import("fmt";"io";"os";"path/filepath";"strings";"encoding/json")
func main(){
 a:=os.Args[1:]; root:=os.Getenv("C12_FAKE_MODULE_ROOT")
 if len(a)==1&&a[0]=="version"{fmt.Println("go version go1.26.5 windows/amd64");return}
 if len(a)>=2&&a[0]=="mod"&&a[1]=="verify"{fmt.Println("all modules verified");return}
 if len(a)>0&&a[0]=="list"{
  joined:=strings.Join(a," "); purpose:="trusted-validator"; dir:=""; importPath:="command-line-arguments"; name:="main"; file:=""
  module:="null"; graphRoot:=filepath.ToSlash(root)
  if strings.Contains(joined,"cmd/test2json"){
   purpose="test2json";dir=filepath.ToSlash(filepath.Join(root,"src","cmd","test2json"));importPath="cmd/test2json";file="test2json.go"
  }else if strings.Contains(joined,"-test")&&strings.Contains(joined,"./internal/testinfra"){
   purpose="authority-initializer";dir=filepath.ToSlash(filepath.Join(root,"internal","testinfra"));importPath="example.invalid/c12graph/internal/testinfra";name="testinfra";file="fixture.go"
   module=fmt.Sprintf("{\"Path\":\"example.invalid/c12graph\",\"Main\":true,\"Dir\":%q,\"GoMod\":%q}",graphRoot,filepath.ToSlash(filepath.Join(root,"go.mod")))
  }else{
   for _,arg:=range a{if strings.HasSuffix(strings.ToLower(arg),".go"){file=filepath.Base(arg);dir=filepath.ToSlash(filepath.Dir(arg));break}}
   if file==""{fmt.Fprintln(os.Stderr,"validator graph omitted exact source");os.Exit(74)}
  }
  line:=fmt.Sprintf("{\"Dir\":%q,\"ImportPath\":%q,\"Name\":%q,\"Root\":%q,\"Module\":%s,\"GoFiles\":[%q],\"Imports\":[],\"Deps\":[],\"TestGoFiles\":[],\"TestImports\":[],\"XTestGoFiles\":[],\"XTestImports\":[],\"EmbedPatterns\":[],\"EmbedFiles\":[],\"TestEmbedPatterns\":[],\"TestEmbedFiles\":[],\"XTestEmbedPatterns\":[],\"XTestEmbedFiles\":[]}",dir,importPath,name,graphRoot,module,file)
  var graph map[string]interface{};if err:=json.Unmarshal([]byte(line),&graph);err!=nil{panic(err)}
  if purpose=="authority-initializer"{graph["EmbedPatterns"]=[]string{"payload.txt"};graph["EmbedFiles"]=[]string{"payload.txt"}}
  mutation:=os.Getenv("C12_FAKE_GO_MUTATION")
  if mutation=="package-import"{graph["Imports"]=[]string{"example.invalid/changed"}}
  if mutation=="module-replacement"{graph["Module"].(map[string]interface{})["Replace"]=map[string]interface{}{"Path":"example.invalid/replacement","Version":"v1.0.1","Dir":graphRoot}}
  encoded,err:=json.Marshal(graph);if err!=nil{panic(err)};line=string(encoded)
  fmt.Println(line);if mutation==purpose+":duplicate-import-path"{fmt.Println(line)};return
 }
 out:="";for i:=0;i+1<len(a);i++{if a[i]=="-o"{out=a[i+1]}}
 if out!=""{src,_:=os.Executable();in,e:=os.Open(src);if e!=nil{panic(e)};defer in.Close();o,e:=os.Create(out);if e!=nil{panic(e)};_,e=io.Copy(o,in);if e!=nil{panic(e)};if e=o.Close();e!=nil{panic(e)};return}
 fmt.Fprintln(os.Stderr,"unexpected fake go argv",strings.Join(a," "));os.Exit(73)
}`)
	runner, err := os.ReadFile("../../scripts/run-c12-integration.ps1")
	if err != nil {
		t.Fatal(err)
	}
	marker := []byte("$script:c12RepositoryRoot = (Resolve-Path")
	index := bytes.Index(runner, marker)
	if index < 0 {
		t.Fatal("runner lacks main-program marker")
	}
	fixturePayload := base64.StdEncoding.EncodeToString([]byte(fixtureRoot))
	fakePayload := base64.StdEncoding.EncodeToString([]byte(fakeGo))
	appendix := fmt.Sprintf(`
$fixtureRoot = [Text.Encoding]::UTF8.GetString([Convert]::FromBase64String('%s'))
$fakeGo = [Text.Encoding]::UTF8.GetString([Convert]::FromBase64String('%s'))
$script:c12RepositoryRoot = $fixtureRoot
[Environment]::SetEnvironmentVariable('C12_FAKE_MODULE_ROOT', $fixtureRoot, 'Process')
function Resolve-C12ClosedGoToolchain {
  param([object]$ArtifactRoot,[DateTime]$SetupDeadline)
  $seal = [C12SealedExecutable]::Inspect($fakeGo)
  try {
    return [pscustomobject]@{ Path=$fakeGo; SHA256=[string]$seal.SHA256; Version='go1.26.5'; GOOS='windows'; GOARCH='amd64'; ModuleCache=$fixtureRoot; Root=$fixtureRoot }
  }
  finally { $seal.Dispose() }
}
$validator = $null
$initializer = $null
$mutants = @()
$artifact = $null
$success = $false
try {
  $allowance = [TimeSpan]$script:c12ProfileAllowances['authority-v7-pitr']
  $deadline = [DateTime]::UtcNow.AddMinutes(20)
  $validator = New-C12PreparedTrustedValidator -Mode 'focused' -DataRoot $fixtureRoot -Packages @('./internal/testinfra') -Tests @('TestC12PreparedGoGraphsBindEverySelectedInput') -Profile 'authority-v7-pitr' -SetupAllowance $allowance -Deadline $deadline
  $artifact = $validator.ArtifactRoot
  $initializer = New-C12PreparedAuthorityInitializer -Profile 'authority-v7-pitr' -RunSuffix ([string]$artifact.RunSuffix) -SetupAllowance $allowance -Deadline $deadline
  foreach ($prepared in @($validator,$initializer)) {
    if ($prepared.Receipt.PSObject.Properties.Name -cnotcontains 'GoGraphReceipts') {
      throw 'production prepared receipt lacks purpose-labelled GoGraphReceipts'
    }
  }
  $graphReceipts = @($validator.Receipt.GoGraphReceipts) + @($initializer.Receipt.GoGraphReceipts)
  $purposes = @('trusted-validator','authority-initializer','test2json')
  if ($graphReceipts.Count -ne 3) { throw "production graph receipt count $($graphReceipts.Count), want exactly three" }
  foreach ($purpose in $purposes) {
    $matches = @($graphReceipts | Where-Object { [string]$_.Purpose -ceq $purpose })
    if ($matches.Count -ne 1) { throw "production graph purpose $purpose count $($matches.Count), want one" }
    $graph = $matches[0]
    if ([string]$graph.Schema -cne 'talenro-c12-go-graph/v1' -or ([byte[]]$graph.BodyBytes).Count -eq 0 -or [string]$graph.Digest -notmatch '^[0-9a-f]{64}$') {
      throw "production graph receipt $purpose lacks exact schema/body/digest"
    }
    Write-Output ('C12_GRAPH_POSITIVE:' + $purpose + ':' + [Convert]::ToBase64String([byte[]]$graph.BodyBytes) + ':' + [string]$graph.Digest)
  }
  $toolchain = Resolve-C12ClosedGoToolchain -ArtifactRoot $artifact -SetupDeadline $deadline
  $baseline = @($initializer.Receipt.GoGraphReceipts | Where-Object Purpose -CEQ 'authority-initializer')[0]
  foreach ($mutation in @('selected-source','embed-file','go.mod','go.sum','module-replacement','package-import','build-argument')) {
    $path = switch ($mutation) {
      'selected-source' { Join-Path $fixtureRoot 'internal/testinfra/fixture.go' }
      'embed-file' { Join-Path $fixtureRoot 'internal/testinfra/payload.txt' }
      'go.mod' { Join-Path $fixtureRoot 'go.mod' }
      'go.sum' { Join-Path $fixtureRoot 'go.sum' }
    }
    $original = $null
    try {
      if ($path) { $original = [IO.File]::ReadAllBytes($path); [IO.File]::WriteAllBytes($path, [byte[]](@($original) + @(32))) }
      [Environment]::SetEnvironmentVariable('C12_FAKE_GO_MUTATION', $mutation, 'Process')
      $argv = @($baseline.BuildArgv)
      if ($mutation -ceq 'build-argument') { $argv += '-trimpath' }
      $changed = New-C12GoGraphReceipt -Purpose 'authority-initializer' -Packages @('./internal/testinfra') -Deadline $deadline -ArtifactRoot $artifact -GoToolchain $toolchain -BuildArgv $argv -WorkingDirectory $fixtureRoot
      if ($changed.CandidateTreeDigest -ceq $baseline.CandidateTreeDigest) { throw "candidate-tree digest ignored $mutation" }
      if ($mutation -cne 'build-argument') {
        $rejected = $false
        try { Assert-C12GoGraphReceipt -Receipt $baseline -ArtifactRoot $artifact -Deadline $deadline }
        catch { $rejected = $_.Exception.Message -match 'graph.*changed' }
        if (-not $rejected) { throw "pre-execution verification accepted $mutation" }
      }
      Write-Output ('C12_GRAPH_MUTATION_OK:' + $mutation)
    }
    finally {
      if ($null -ne $original) { [IO.File]::WriteAllBytes($path, $original) }
      [Environment]::SetEnvironmentVariable('C12_FAKE_GO_MUTATION', $null, 'Process')
    }
  }
  foreach ($purpose in $purposes) {
    [Environment]::SetEnvironmentVariable('C12_FAKE_GO_MUTATION', ($purpose + ':duplicate-import-path'), 'Process')
    $rejected = $false
    try {
      if ($purpose -ceq 'trusted-validator') {
        $candidate = New-C12PreparedTrustedValidator -Mode 'focused' -DataRoot $fixtureRoot -Packages @('./internal/testinfra') -Tests @('TestC12PreparedGoGraphsBindEverySelectedInput') -Profile 'authority-v7-pitr' -SetupAllowance $allowance -Deadline $deadline
        $mutants += $candidate
      }
      elseif ($purpose -ceq 'authority-initializer') {
        $candidate = New-C12PreparedAuthorityInitializer -Profile 'authority-v7-pitr' -RunSuffix ([string]$artifact.RunSuffix) -SetupAllowance $allowance -Deadline $deadline
        $mutants += $candidate
      }
      else {
        $toolchain = Resolve-C12ClosedGoToolchain -ArtifactRoot $artifact -SetupDeadline $deadline
        $null = Resolve-C12Test2JSONExecutable -ArtifactRoot $artifact -GoToolchain $toolchain -SetupDeadline $deadline
      }
    }
    catch { $rejected = $_.Exception.Message -match 'duplicate.*ImportPath|ImportPath.*duplicate' }
    finally { [Environment]::SetEnvironmentVariable('C12_FAKE_GO_MUTATION', $null, 'Process') }
    if (-not $rejected) { throw "production $purpose graph parser accepted its duplicate ImportPath mutation" }
  }
  $success = $true
}
catch { [Console]::Error.WriteLine($_.Exception.Message) }
finally {
  [Environment]::SetEnvironmentVariable('C12_FAKE_GO_MUTATION', $null, 'Process')
  foreach ($prepared in @($mutants + @($initializer,$validator))) {
    if ($null -ne $prepared) { try { Close-C12PreparedNativeWorker -Prepared $prepared -Deadline ([DateTime]::UtcNow.AddSeconds(15)) } catch { [Console]::Error.WriteLine($_.Exception.Message); $success=$false } }
  }
  if ($null -ne $artifact) { try { Remove-C12PreparedArtifactRoot -ArtifactRoot $artifact -Deadline ([DateTime]::UtcNow.AddSeconds(15)) } catch { [Console]::Error.WriteLine($_.Exception.Message); $success=$false } }
}
if ($success) { Write-Output 'C12_PREPARED_GRAPH_HARNESS_OK'; exit 0 }
exit 1
`, fixturePayload, fakePayload)
	harness := filepath.Join(t.TempDir(), "assert-c12-prepared-go-graph.ps1")
	if err := os.WriteFile(harness, append(append([]byte(nil), runner[:index]...), []byte(appendix)...), 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Minute)
	defer cancel()
	powershellPath := filepath.Join(os.Getenv("SystemRoot"), "System32", "WindowsPowerShell", "v1.0", "powershell.exe")
	command := exec.CommandContext(ctx, powershellPath, "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", harness, "-Profile", "base", "-Packages", "./internal/store", "-Timeout", "3m")
	output, commandErr := command.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("prepared graph harness timed out: %v\n%s", ctx.Err(), output)
	}
	if commandErr == nil {
		markers := regexp.MustCompile(`(?m)^C12_GRAPH_POSITIVE:(trusted-validator|authority-initializer|test2json):([A-Za-z0-9+/=]+):([0-9a-f]{64})\r?$`).FindAllSubmatch(output, -1)
		if len(markers) != 3 {
			t.Fatalf("prepared graph positive marker count = %d, want three exact purposes; output=%q", len(markers), output)
		}
		seenPurpose := make(map[string]bool, 3)
		seenBody := make(map[string]bool, 3)
		for _, marker := range markers {
			purpose := string(marker[1])
			if seenPurpose[purpose] {
				t.Fatalf("production graph purpose %s was emitted more than once", purpose)
			}
			seenPurpose[purpose] = true
			graphBytes, err := base64.StdEncoding.DecodeString(string(marker[2]))
			if err != nil {
				t.Fatal(err)
			}
			bodyKey := string(graphBytes)
			if seenBody[bodyKey] {
				t.Fatalf("production graph purpose %s aliased another purpose's body", purpose)
			}
			seenBody[bodyKey] = true
			digestInput := append([]byte("talenro.c12.go-graph.v1\x00"), graphBytes...)
			got := sha256.Sum256(digestInput)
			if hex.EncodeToString(got[:]) != string(marker[3]) {
				t.Fatalf("production %s graph digest does not bind exact graph bytes: got %x want %s", purpose, got, marker[3])
			}
			var candidate map[string]json.RawMessage
			if err := json.Unmarshal(graphBytes, &candidate); err != nil {
				t.Fatal(err)
			}
			var candidateDigest string
			if err := json.Unmarshal(candidate["candidate_tree_digest"], &candidateDigest); err != nil {
				t.Fatal(err)
			}
			delete(candidate, "candidate_tree_digest")
			candidateBytes, err := json.Marshal(candidate)
			if err != nil {
				t.Fatal(err)
			}
			candidateHash := sha256.Sum256(candidateBytes)
			if hex.EncodeToString(candidateHash[:]) != candidateDigest {
				t.Fatalf("production %s candidate-tree digest does not bind its literal input registry", purpose)
			}
		}
		return string(output), 0
	}
	var exitError *exec.ExitError
	if !errors.As(commandErr, &exitError) {
		t.Fatalf("launch prepared graph harness: %v", commandErr)
	}
	return string(output), exitError.ExitCode()
}

func runC12NamedJobCollisionHarness(t *testing.T) (string, int) {
	t.Helper()
	runner, err := os.ReadFile("../../scripts/run-c12-integration.ps1")
	if err != nil {
		t.Fatal(err)
	}
	marker := []byte("$script:c12RepositoryRoot = (Resolve-Path")
	index := bytes.Index(runner, marker)
	if index < 0 {
		t.Fatal("runner lacks main-program marker")
	}
	jobName := fmt.Sprintf("TalenroC12Collision_%d", time.Now().UnixNano())
	appendix := fmt.Sprintf(`
$first = [IntPtr]::Zero
$second = [IntPtr]::Zero
try {
  $first = [C12NativeJob]::CreateKillOnClose('%s')
  try {
    $second = [C12NativeJob]::CreateKillOnClose('%s')
    throw 'named native Job collision was accepted'
  }
  catch {
    if ($_.Exception.Message -notmatch 'collision') { throw }
  }
  exit 0
}
catch { [Console]::Error.WriteLine($_.Exception.Message); exit 1 }
finally { [C12NativeJob]::Close($second); [C12NativeJob]::Close($first) }
`, jobName, jobName)
	harness := filepath.Join(t.TempDir(), "assert-c12-job-collision.ps1")
	if err := os.WriteFile(harness, append(append([]byte(nil), runner[:index]...), []byte(appendix)...), 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", harness,
		"-Profile", "base", "-Packages", "./internal/store", "-Timeout", "3m")
	output, commandErr := command.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("named Job collision harness timed out: %v\n%s", ctx.Err(), output)
	}
	if commandErr == nil {
		return string(output), 0
	}
	var exitError *exec.ExitError
	if !errors.As(commandErr, &exitError) {
		t.Fatalf("launch named Job collision harness: %v", commandErr)
	}
	return string(output), exitError.ExitCode()
}

func runC12OwnedDirectoryIdentityHarness(t *testing.T) (string, int) {
	t.Helper()
	runner, err := os.ReadFile("../../scripts/run-c12-integration.ps1")
	if err != nil {
		t.Fatal(err)
	}
	marker := []byte("$script:c12RepositoryRoot = (Resolve-Path")
	index := bytes.Index(runner, marker)
	if index < 0 {
		t.Fatal("runner lacks main-program marker")
	}
	rootPayload := base64.StdEncoding.EncodeToString([]byte(t.TempDir()))
	appendix := fmt.Sprintf(`
$testRoot = [Text.Encoding]::UTF8.GetString([Convert]::FromBase64String('%s'))
$canaryRoot = Join-Path $testRoot 'canary'
$canaryFile = Join-Path $canaryRoot 'readonly-canary.txt'
$ordinaryRoot = Join-Path $testRoot 'talenro-c12-validator-00000000000000000000000000000000'
$ordinaryFile = Join-Path $ordinaryRoot 'ordinary-readonly-canary.txt'
$replacementRoot = Join-Path $testRoot 'talenro-c12-validator-11111111111111111111111111111111'
$ownedRoot = Join-Path $testRoot 'talenro-c12-validator-22222222222222222222222222222222'
$movedRoot = Join-Path $testRoot 'moved-owned-root'
$deadlineRoot = Join-Path $testRoot 'talenro-c12-validator-33333333333333333333333333333333'
$movedDeadlineRoot = Join-Path $testRoot 'moved-deadline-root'
$ownership = $null
try {
  [IO.Directory]::CreateDirectory($canaryRoot) | Out-Null
  [IO.File]::WriteAllText($canaryFile, 'C12_DIRECTORY_IDENTITY_CANARY')
  [IO.File]::SetAttributes($canaryFile, [IO.FileAttributes]::ReadOnly)

  [IO.Directory]::CreateDirectory($ordinaryRoot) | Out-Null
  [IO.File]::WriteAllText($ordinaryFile, 'C12_NULL_OWNERSHIP_CANARY')
  [IO.File]::SetAttributes($ordinaryFile, [IO.FileAttributes]::ReadOnly)
  $ordinaryRejected = $false
  $ordinaryFailure = ''
  try {
    Remove-C12BoundedDirectory -Root $ordinaryRoot -ExpectedParent $testRoot -LeafPattern '^talenro-c12-validator-[0-9a-f]{32}$' -Stage 'ordinary null-ownership cleanup'
  }
  catch { $ordinaryRejected = $true; $ordinaryFailure = $_.Exception.Message }
  if (-not $ordinaryRejected) { throw 'ordinary directory was adopted without creation ownership' }
  if (-not $ordinaryFailure.Contains($ordinaryRoot)) { throw 'null-ownership failure omitted the exact ordinary orphan' }
  if (-not [IO.File]::Exists($ordinaryFile)) { throw 'null-ownership cleanup deleted the ordinary canary' }
  if ([IO.File]::ReadAllText($ordinaryFile) -cne 'C12_NULL_OWNERSHIP_CANARY') { throw 'null-ownership cleanup changed ordinary canary bytes' }
  if (([IO.File]::GetAttributes($ordinaryFile) -band [IO.FileAttributes]::ReadOnly) -eq 0) { throw 'null-ownership cleanup normalized the ordinary canary attribute' }

  New-Item -ItemType Junction -Path $replacementRoot -Target $canaryRoot -Force | Out-Null
  $replacementRejected = $false
  try {
    Remove-C12BoundedDirectory -Root $replacementRoot -ExpectedParent $testRoot -LeafPattern '^talenro-c12-validator-[0-9a-f]{32}$' -Stage 'replacement root cleanup'
  }
  catch { $replacementRejected = $true }
  if (-not $replacementRejected) { throw 'replacement root cleanup was not rejected before traversal' }
  if ([IO.File]::ReadAllText($canaryFile) -cne 'C12_DIRECTORY_IDENTITY_CANARY') { throw 'replacement cleanup changed canary bytes' }
  if (([IO.File]::GetAttributes($canaryFile) -band [IO.FileAttributes]::ReadOnly) -eq 0) { throw 'replacement cleanup normalized the canary attribute' }
  if ([IO.Directory]::Exists($replacementRoot)) { [IO.Directory]::Delete($replacementRoot, $false) }

  $ownership = New-C12OwnedDirectory -Root $ownedRoot -ExpectedParent $testRoot -LeafPattern '^talenro-c12-validator-[0-9a-f]{32}$' -Stage 'owned directory creation'
  [IO.File]::WriteAllText((Join-Path $ownedRoot 'owned.txt'), 'owned')
  $substitutionSucceeded = $false
  try {
    [IO.Directory]::Move($ownedRoot, $movedRoot)
    New-Item -ItemType Junction -Path $ownedRoot -Target $canaryRoot -Force | Out-Null
    $substitutionSucceeded = $true
  }
  catch { }
  if ($substitutionSucceeded) { throw 'retained owned-directory handle allowed root substitution' }
  Remove-C12BoundedDirectory -Root $ownedRoot -ExpectedParent $testRoot -LeafPattern '^talenro-c12-validator-[0-9a-f]{32}$' -Stage 'owned directory cleanup' -Ownership $ownership
  $ownership = $null
  if ([IO.Directory]::Exists($ownedRoot) -or [IO.Directory]::Exists($movedRoot)) { throw 'owned directory identity cleanup left its exact root' }
  if ([IO.File]::ReadAllText($canaryFile) -cne 'C12_DIRECTORY_IDENTITY_CANARY') { throw 'owned cleanup changed canary bytes' }
  if (([IO.File]::GetAttributes($canaryFile) -band [IO.FileAttributes]::ReadOnly) -eq 0) { throw 'owned cleanup normalized the canary attribute' }

  $ownership = New-C12OwnedDirectory -Root $deadlineRoot -ExpectedParent $testRoot -LeafPattern '^talenro-c12-validator-[0-9a-f]{32}$' -Stage 'deadline-owned directory creation'
  $deadlineRejected = $false
  try {
    Remove-C12BoundedDirectory -Root $deadlineRoot -ExpectedParent $testRoot -LeafPattern '^talenro-c12-validator-[0-9a-f]{32}$' -Stage 'deadline-owned directory cleanup' -Deadline ([DateTime]::UtcNow.AddSeconds(-1)) -Ownership $ownership
  }
  catch { $deadlineRejected = $true }
  $ownership = $null
  if (-not $deadlineRejected) { throw 'expired cleanup deadline was accepted' }
  [IO.Directory]::Move($deadlineRoot, $movedDeadlineRoot)
  [IO.Directory]::Delete($movedDeadlineRoot, $true)
  exit 0
}
catch { [Console]::Error.WriteLine($_.Exception.Message); exit 1 }
finally {
  if ($null -ne $ownership) { $ownership.Dispose() }
  if ([IO.Directory]::Exists($replacementRoot) -and (([IO.File]::GetAttributes($replacementRoot) -band [IO.FileAttributes]::ReparsePoint) -ne 0)) { [IO.Directory]::Delete($replacementRoot, $false) }
  if ([IO.File]::Exists($ordinaryFile)) { [IO.File]::SetAttributes($ordinaryFile, [IO.FileAttributes]::Normal) }
  if ([IO.Directory]::Exists($ordinaryRoot)) { [IO.Directory]::Delete($ordinaryRoot, $true) }
  if ([IO.Directory]::Exists($deadlineRoot)) { [IO.Directory]::Delete($deadlineRoot, $true) }
  if ([IO.Directory]::Exists($movedDeadlineRoot)) { [IO.Directory]::Delete($movedDeadlineRoot, $true) }
  if ([IO.File]::Exists($canaryFile)) { [IO.File]::SetAttributes($canaryFile, [IO.FileAttributes]::Normal) }
}
`, rootPayload)
	harness := filepath.Join(t.TempDir(), "assert-c12-owned-directory.ps1")
	if err := os.WriteFile(harness, append(append([]byte(nil), runner[:index]...), []byte(appendix)...), 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", harness,
		"-Profile", "base", "-Packages", "./internal/store", "-Timeout", "3m")
	output, commandErr := command.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("owned directory identity harness timed out: %v\n%s", ctx.Err(), output)
	}
	if commandErr == nil {
		return string(output), 0
	}
	var exitError *exec.ExitError
	if !errors.As(commandErr, &exitError) {
		t.Fatalf("launch owned directory identity harness: %v", commandErr)
	}
	return string(output), exitError.ExitCode()
}

func runC12AtomicOwnedDirectoryCreationHarness(t *testing.T) (string, int) {
	t.Helper()
	runner, err := os.ReadFile("../../scripts/run-c12-integration.ps1")
	if err != nil {
		t.Fatal(err)
	}
	marker := []byte("$script:c12RepositoryRoot = (Resolve-Path")
	index := bytes.Index(runner, marker)
	if index < 0 {
		t.Fatal("runner lacks main-program marker")
	}
	rootPayload := base64.StdEncoding.EncodeToString([]byte(t.TempDir()))
	appendix := fmt.Sprintf(`
Add-Type -TypeDefinition @'
using System;
using System.ComponentModel;
using System.IO;
using System.Runtime.InteropServices;
using System.Threading;

public sealed class C12CreateSubstitutionFixture : IDisposable
{
    private const UInt32 FILE_LIST_DIRECTORY = 0x00000001;
    private const UInt32 FILE_SHARE_READ = 0x00000001;
    private const UInt32 FILE_SHARE_WRITE = 0x00000002;
    private const UInt32 FILE_SHARE_DELETE = 0x00000004;
    private const UInt32 OPEN_EXISTING = 3;
    private const UInt32 FILE_FLAG_BACKUP_SEMANTICS = 0x02000000;
    private const UInt32 FILE_NOTIFY_CHANGE_DIR_NAME = 0x00000002;
    private static readonly IntPtr INVALID_HANDLE_VALUE = new IntPtr(-1);

    [DllImport("kernel32.dll", CharSet = CharSet.Unicode, SetLastError = true)]
    private static extern IntPtr CreateFile(string path, UInt32 desiredAccess, UInt32 shareMode, IntPtr securityAttributes, UInt32 creationDisposition, UInt32 flagsAndAttributes, IntPtr templateFile);
    [DllImport("kernel32.dll", SetLastError = true)]
    private static extern bool ReadDirectoryChangesW(IntPtr directory, IntPtr buffer, UInt32 length, bool watchSubtree, UInt32 filter, out UInt32 bytesReturned, IntPtr overlapped, IntPtr completionRoutine);
    [DllImport("kernel32.dll", SetLastError = true)]
    private static extern bool CloseHandle(IntPtr handle);

    private IntPtr parentHandle;
    private readonly string target;
    private readonly string attacker;
    private readonly ManualResetEvent armed = new ManualResetEvent(false);
    private readonly ManualResetEvent completed = new ManualResetEvent(false);
    private readonly Thread thread;
    public bool Substituted { get; private set; }
    public string Failure { get; private set; }

    private C12CreateSubstitutionFixture(IntPtr handle, string exactTarget, string exactAttacker)
    {
        parentHandle = handle;
        target = exactTarget;
        attacker = exactAttacker;
        Failure = String.Empty;
        thread = new Thread(Run);
        thread.IsBackground = true;
        thread.Priority = ThreadPriority.Highest;
        thread.Start();
    }

    public static C12CreateSubstitutionFixture Arm(string parent, string target, string attacker)
    {
        IntPtr handle = CreateFile(parent, FILE_LIST_DIRECTORY, FILE_SHARE_READ | FILE_SHARE_WRITE | FILE_SHARE_DELETE, IntPtr.Zero, OPEN_EXISTING, FILE_FLAG_BACKUP_SEMANTICS, IntPtr.Zero);
        if (handle == INVALID_HANDLE_VALUE) throw new Win32Exception(Marshal.GetLastWin32Error());
        return new C12CreateSubstitutionFixture(handle, target, attacker);
    }

    private void Run()
    {
        IntPtr buffer = Marshal.AllocHGlobal(4096);
        try
        {
            armed.Set();
            UInt32 returned;
            if (!ReadDirectoryChangesW(parentHandle, buffer, 4096, false, FILE_NOTIFY_CHANGE_DIR_NAME, out returned, IntPtr.Zero, IntPtr.Zero))
                throw new Win32Exception(Marshal.GetLastWin32Error());
            Directory.Delete(target, false);
            Directory.Move(attacker, target);
            Substituted = true;
        }
        catch (Exception exception) { Failure = exception.GetType().Name + ": " + exception.Message; }
        finally
        {
            Marshal.FreeHGlobal(buffer);
            completed.Set();
        }
    }

    public void WaitUntilArmed()
    {
        if (!armed.WaitOne(5000)) throw new TimeoutException("directory substitution fixture did not arm");
    }

    public void WaitForCompletion()
    {
        if (!completed.WaitOne(5000)) throw new TimeoutException("directory substitution fixture did not complete");
    }

    public void Dispose()
    {
        IntPtr handle = parentHandle;
        parentHandle = IntPtr.Zero;
        if (handle != IntPtr.Zero && handle != INVALID_HANDLE_VALUE) CloseHandle(handle);
        if (!completed.WaitOne(5000) && thread.IsAlive) thread.Join(1000);
        armed.Dispose();
        completed.Dispose();
    }
}
'@
$testRoot = [Text.Encoding]::UTF8.GetString([Convert]::FromBase64String('%s'))
$originalPriority = [Threading.Thread]::CurrentThread.Priority
try {
  for ($attempt = 1; $attempt -le 8; $attempt++) {
    $suffix = ([string]$attempt).PadLeft(32, '0')
    $ownedRoot = Join-Path $testRoot "talenro-c12-validator-$suffix"
    $attackerRoot = Join-Path $testRoot "attacker-$suffix"
    $attackerCanary = Join-Path $attackerRoot 'readonly-canary.txt'
    $fixture = $null
    $ownership = $null
    try {
      [IO.Directory]::CreateDirectory($attackerRoot) | Out-Null
      [IO.File]::WriteAllText($attackerCanary, 'C12_ATOMIC_CREATE_CANARY')
      [IO.File]::SetAttributes($attackerCanary, [IO.FileAttributes]::ReadOnly)
      $fixture = [C12CreateSubstitutionFixture]::Arm($testRoot, $ownedRoot, $attackerRoot)
      $fixture.WaitUntilArmed()
      Start-Sleep -Milliseconds 100
      [Threading.Thread]::CurrentThread.Priority = [Threading.ThreadPriority]::Lowest
      $createFailure = ''
      try {
        $ownership = New-C12OwnedDirectory -Root $ownedRoot -ExpectedParent $testRoot -LeafPattern '^talenro-c12-validator-[0-9a-f]{32}$' -Stage 'atomic owned-directory creation'
      }
      catch { $createFailure = $_.Exception.Message }
      finally { [Threading.Thread]::CurrentThread.Priority = $originalPriority }
      $fixture.WaitForCompletion()

      $canaryPath = if ([IO.File]::Exists((Join-Path $ownedRoot 'readonly-canary.txt'))) { Join-Path $ownedRoot 'readonly-canary.txt' } else { $attackerCanary }
      if (-not [IO.File]::Exists($canaryPath)) { throw 'create/open race deleted the attacker canary' }
      if ([IO.File]::ReadAllText($canaryPath) -cne 'C12_ATOMIC_CREATE_CANARY') { throw 'create/open race changed attacker canary bytes' }
      if (([IO.File]::GetAttributes($canaryPath) -band [IO.FileAttributes]::ReadOnly) -eq 0) { throw 'create/open race normalized attacker canary attributes' }
      if (-not [string]::IsNullOrEmpty($createFailure)) { throw "create/open race caused owned-directory failure: $createFailure" }
      if ($fixture.Substituted) { throw 'two-step creation adopted a substituted ordinary directory' }
    }
    finally {
      [Threading.Thread]::CurrentThread.Priority = $originalPriority
      if ($null -ne $ownership) { $ownership.Dispose() }
      if (-not [IO.Directory]::Exists($testRoot)) { throw "test root disappeared after ownership dispose at attempt $attempt (owned=$ownedRoot attacker=$attackerRoot)" }
      if ($null -ne $fixture) { $fixture.Dispose() }
      if (-not [IO.Directory]::Exists($testRoot)) { throw "test root disappeared after fixture dispose at attempt $attempt (owned=$ownedRoot attacker=$attackerRoot)" }
      foreach ($candidate in @($ownedRoot, $attackerRoot)) {
        $candidateCanary = Join-Path $candidate 'readonly-canary.txt'
        if ([IO.File]::Exists($candidateCanary)) { [IO.File]::SetAttributes($candidateCanary, [IO.FileAttributes]::Normal) }
        if ([IO.Directory]::Exists($candidate)) { [IO.Directory]::Delete($candidate, $true) }
        if (-not [IO.Directory]::Exists($testRoot)) { throw "test root disappeared while deleting $candidate at attempt $attempt" }
      }
    }
  }

  if (-not [IO.Directory]::Exists($testRoot)) { throw 'test root disappeared before collision verification' }
  $collisionRoot = Join-Path $testRoot 'talenro-c12-validator-ffffffffffffffffffffffffffffffff'
  $collisionCanary = Join-Path $collisionRoot 'c.txt'
  [IO.Directory]::CreateDirectory($collisionRoot) | Out-Null
  [IO.File]::WriteAllText($collisionCanary, 'C12_ATOMIC_COLLISION_CANARY')
  [IO.File]::SetAttributes($collisionCanary, [IO.FileAttributes]::ReadOnly)
  $collisionRejected = $false
  try {
    $unexpected = New-C12OwnedDirectory -Root $collisionRoot -ExpectedParent $testRoot -LeafPattern '^talenro-c12-validator-[0-9a-f]{32}$' -Stage 'atomic owned-directory collision'
    if ($null -ne $unexpected) { $unexpected.Dispose() }
  }
  catch { $collisionRejected = $true }
  if (-not $collisionRejected) { throw 'atomic creation accepted a pre-existing collision' }
  if ([IO.File]::ReadAllText($collisionCanary) -cne 'C12_ATOMIC_COLLISION_CANARY') { throw 'creation collision changed canary bytes' }
  if (([IO.File]::GetAttributes($collisionCanary) -band [IO.FileAttributes]::ReadOnly) -eq 0) { throw 'creation collision normalized canary attributes' }
  [IO.File]::SetAttributes($collisionCanary, [IO.FileAttributes]::Normal)
  [IO.Directory]::Delete($collisionRoot, $true)
  exit 0
}
catch { [Console]::Error.WriteLine($_.Exception.Message); exit 1 }
finally { [Threading.Thread]::CurrentThread.Priority = $originalPriority }
`, rootPayload)
	harness := filepath.Join(t.TempDir(), "assert-c12-atomic-owned-directory.ps1")
	if err := os.WriteFile(harness, append(append([]byte(nil), runner[:index]...), []byte(appendix)...), 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", harness,
		"-Profile", "base", "-Packages", "./internal/store", "-Timeout", "3m")
	output, commandErr := command.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("atomic owned-directory creation harness timed out: %v\n%s", ctx.Err(), output)
	}
	if commandErr == nil {
		return string(output), 0
	}
	var exitError *exec.ExitError
	if !errors.As(commandErr, &exitError) {
		t.Fatalf("launch atomic owned-directory creation harness: %v", commandErr)
	}
	return string(output), exitError.ExitCode()
}

func runC12PreMembershipProcessHarness(t *testing.T, sentinel string) (string, int) {
	t.Helper()
	runner, err := os.ReadFile("../../scripts/run-c12-integration.ps1")
	if err != nil {
		t.Fatal(err)
	}
	marker := []byte("$script:c12RepositoryRoot = (Resolve-Path")
	index := bytes.Index(runner, marker)
	if index < 0 {
		t.Fatal("runner lacks main-program marker")
	}
	fakeBin := t.TempDir()
	if err := os.WriteFile(filepath.Join(fakeBin, "docker.cmd"), []byte("@echo off\r\n>\"%C12_PREMEMBERSHIP_SENTINEL%\" echo spawned\r\nexit /b 0\r\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	gatePayload := base64.StdEncoding.EncodeToString([]byte(filepath.Join(t.TempDir(), "release")))
	workingPayload := base64.StdEncoding.EncodeToString([]byte(t.TempDir()))
	jobName := fmt.Sprintf("TalenroC12PreMember_%d", time.Now().UnixNano())
	appendix := fmt.Sprintf(`
Add-Type -TypeDefinition @'
using System;
using System.ComponentModel;
using System.Runtime.InteropServices;
public static class C12SingleProcessJobFixture {
  private const UInt32 JOB_OBJECT_LIMIT_ACTIVE_PROCESS = 0x00000008;
  private const UInt32 JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE = 0x00002000;
  private const UInt32 PROCESS_TERMINATE = 0x0001;
  private const UInt32 PROCESS_SET_QUOTA = 0x0100;
  [StructLayout(LayoutKind.Sequential)] private struct BASIC {
    public Int64 PerProcessUserTimeLimit; public Int64 PerJobUserTimeLimit; public UInt32 LimitFlags;
    public UIntPtr MinimumWorkingSetSize; public UIntPtr MaximumWorkingSetSize; public UInt32 ActiveProcessLimit;
    public IntPtr Affinity; public UInt32 PriorityClass; public UInt32 SchedulingClass;
  }
  [StructLayout(LayoutKind.Sequential)] private struct IO_COUNTERS { public UInt64 ReadOperationCount; public UInt64 WriteOperationCount; public UInt64 OtherOperationCount; public UInt64 ReadTransferCount; public UInt64 WriteTransferCount; public UInt64 OtherTransferCount; }
  [StructLayout(LayoutKind.Sequential)] private struct EXTENDED { public BASIC BasicLimitInformation; public IO_COUNTERS IoInfo; public UIntPtr ProcessMemoryLimit; public UIntPtr JobMemoryLimit; public UIntPtr PeakProcessMemoryUsed; public UIntPtr PeakJobMemoryUsed; }
  [DllImport("kernel32.dll", CharSet=CharSet.Unicode, SetLastError=true)] private static extern IntPtr CreateJobObject(IntPtr attributes, string name);
  [DllImport("kernel32.dll", SetLastError=true)] private static extern bool SetInformationJobObject(IntPtr job, int informationClass, IntPtr information, UInt32 length);
  [DllImport("kernel32.dll", SetLastError=true)] private static extern IntPtr OpenProcess(UInt32 access, bool inheritHandle, UInt32 processID);
  [DllImport("kernel32.dll", SetLastError=true)] private static extern bool AssignProcessToJobObject(IntPtr job, IntPtr process);
  [DllImport("kernel32.dll", SetLastError=true)] private static extern bool CloseHandle(IntPtr handle);
  public static IntPtr Create(string name) {
    IntPtr job=CreateJobObject(IntPtr.Zero,name); if(job==IntPtr.Zero) throw new Win32Exception(Marshal.GetLastWin32Error());
    EXTENDED information=new EXTENDED(); information.BasicLimitInformation.LimitFlags=JOB_OBJECT_LIMIT_ACTIVE_PROCESS|JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE; information.BasicLimitInformation.ActiveProcessLimit=1;
    int length=Marshal.SizeOf(typeof(EXTENDED)); IntPtr pointer=Marshal.AllocHGlobal(length);
    try { Marshal.StructureToPtr(information,pointer,false); if(!SetInformationJobObject(job,9,pointer,(UInt32)length)){int error=Marshal.GetLastWin32Error();CloseHandle(job);throw new Win32Exception(error);} }
    finally { Marshal.FreeHGlobal(pointer); }
    return job;
  }
  public static void Assign(IntPtr job, UInt32 processID) {
    IntPtr process=OpenProcess(PROCESS_TERMINATE|PROCESS_SET_QUOTA,false,processID); if(process==IntPtr.Zero) throw new Win32Exception(Marshal.GetLastWin32Error());
    try { if(!AssignProcessToJobObject(job,process)) throw new Win32Exception(Marshal.GetLastWin32Error()); }
    finally { CloseHandle(process); }
  }
  public static void Close(IntPtr handle) { if(handle!=IntPtr.Zero) CloseHandle(handle); }
}
'@
$gatePath = [Text.Encoding]::UTF8.GetString([Convert]::FromBase64String('%s'))
$workingRoot = [Text.Encoding]::UTF8.GetString([Convert]::FromBase64String('%s'))
$targetJobName = '%s'
$outerHandle = [IntPtr]::Zero
$targetHandle = [IntPtr]::Zero
$job = $null
try {
  $outerHandle = [C12SingleProcessJobFixture]::Create("TalenroC12Single_$([Guid]::NewGuid().ToString('N'))")
  $targetHandle = [C12NativeJob]::CreateKillOnClose($targetJobName)
  $invocation = [pscustomobject]@{ Executable='docker'; Arguments=[string[]]@('version'); WorkingDirectory=$workingRoot; JobName=$targetJobName; Environment=[object[]]@() }
  $containedText = $script:c12ContainedNativeScript.ToString()
  $job = Start-Job -ArgumentList @($invocation, $containedText, $gatePath) -ScriptBlock {
    param($Invocation, $ContainedText, $GatePath)
    [pscustomobject]@{ C12GatePID = [int]$PID }
    $limit = [DateTime]::UtcNow.AddSeconds(10)
    while (-not [IO.File]::Exists($GatePath)) {
      if ([DateTime]::UtcNow -ge $limit) { throw 'test gate timed out' }
      Start-Sleep -Milliseconds 10
    }
    $contained = [ScriptBlock]::Create($ContainedText)
    & $contained $Invocation
  }
  $hostPID = 0
  $limit = [DateTime]::UtcNow.AddSeconds(10)
  while ($hostPID -eq 0 -and [DateTime]::UtcNow -lt $limit) {
    foreach ($item in @(Receive-Job -Job $job -Keep -ErrorAction SilentlyContinue)) {
      if ($item.PSObject.Properties.Name -contains 'C12GatePID') { $hostPID = [int]$item.C12GatePID }
    }
    if ($hostPID -eq 0) { Start-Sleep -Milliseconds 10 }
  }
  if ($hostPID -eq 0) { throw 'test gate did not publish the exact child PID' }
  [C12SingleProcessJobFixture]::Assign($outerHandle, [uint32]$hostPID)
  [IO.File]::WriteAllText($gatePath, 'release')
  if ($null -eq (Wait-Job -Job $job -Timeout 15)) { throw 'pre-membership process harness timed out' }
  $received = @(Receive-Job -Job $job -ErrorAction SilentlyContinue)
  $result = @($received | Where-Object { $_.PSObject.Properties.Name -contains 'ContainmentFailure' } | Select-Object -Last 1)
  if ($result.Count -ne 1) { throw 'contained child returned no exact result' }
  if ([bool]$result[0].ContainmentFailure) { throw 'compiler/helper process was required before named Job membership' }
  if ([int]$result[0].ExitCode -ne 127) { throw 'test-only process limit did not block the post-membership target' }
  exit 0
}
catch { [Console]::Error.WriteLine($_.Exception.Message); exit 1 }
finally {
  if ($targetHandle -ne [IntPtr]::Zero) { [C12NativeJob]::Close($targetHandle) }
  if ($null -ne $job) { Stop-Job -Job $job -ErrorAction SilentlyContinue; Remove-Job -Job $job -Force -ErrorAction SilentlyContinue }
  [C12SingleProcessJobFixture]::Close($outerHandle)
}
`, gatePayload, workingPayload, jobName)
	harness := filepath.Join(t.TempDir(), "assert-c12-pre-membership.ps1")
	if err := os.WriteFile(harness, append(append([]byte(nil), runner[:index]...), []byte(appendix)...), 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", harness,
		"-Profile", "base", "-Packages", "./internal/store", "-Timeout", "3m")
	command.Env = append([]string{}, os.Environ()...)
	command.Env = append(command.Env,
		"C12_PREMEMBERSHIP_SENTINEL="+sentinel,
		"Path="+fakeBin+string(os.PathListSeparator)+os.Getenv("Path"))
	output, commandErr := command.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("pre-membership process harness timed out: %v\n%s", ctx.Err(), output)
	}
	if commandErr == nil {
		return string(output), 0
	}
	var exitError *exec.ExitError
	if !errors.As(commandErr, &exitError) {
		t.Fatalf("launch pre-membership process harness: %v", commandErr)
	}
	return string(output), exitError.ExitCode()
}

func runC12ContainmentFailureHarness(t *testing.T, mode, sentinel string) (string, int) {
	t.Helper()
	runner, err := os.ReadFile("../../scripts/run-c12-integration.ps1")
	if err != nil {
		t.Fatal(err)
	}
	marker := []byte("$script:c12RepositoryRoot = (Resolve-Path")
	index := bytes.Index(runner, marker)
	if index < 0 {
		t.Fatal("runner lacks main-program marker")
	}
	fakeBin := t.TempDir()
	if err := os.WriteFile(filepath.Join(fakeBin, "docker.cmd"), []byte("@echo off\r\n>\"%C12_CONTAINMENT_SENTINEL%\" echo spawned\r\nexit /b 0\r\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	legacyArtifacts := t.TempDir()
	if err := os.WriteFile(filepath.Join(legacyArtifacts, "host.pid"), []byte("4"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(legacyArtifacts, "release"), []byte("attacker"), 0o600); err != nil {
		t.Fatal(err)
	}
	jobName := fmt.Sprintf("TalenroC12Reject_%d", time.Now().UnixNano())
	setup := ""
	if mode == "assignment" {
		setup = "$rejectingHandle = [C12RejectingJobFixture]::Create($jobName)"
	}
	rootPayload := base64.StdEncoding.EncodeToString([]byte(legacyArtifacts))
	appendix := fmt.Sprintf(`
Add-Type -TypeDefinition @'
using System;
using System.ComponentModel;
using System.Runtime.InteropServices;
public static class C12RejectingJobFixture {
  private const UInt32 JOB_OBJECT_LIMIT_ACTIVE_PROCESS = 0x00000008;
  [StructLayout(LayoutKind.Sequential)] private struct BASIC {
    public Int64 PerProcessUserTimeLimit; public Int64 PerJobUserTimeLimit; public UInt32 LimitFlags;
    public UIntPtr MinimumWorkingSetSize; public UIntPtr MaximumWorkingSetSize; public UInt32 ActiveProcessLimit;
    public IntPtr Affinity; public UInt32 PriorityClass; public UInt32 SchedulingClass;
  }
  [DllImport("kernel32.dll", CharSet=CharSet.Unicode, SetLastError=true)] private static extern IntPtr CreateJobObject(IntPtr attributes, string name);
  [DllImport("kernel32.dll", SetLastError=true)] private static extern bool SetInformationJobObject(IntPtr job, int informationClass, IntPtr information, UInt32 length);
  [DllImport("kernel32.dll", SetLastError=true)] private static extern bool CloseHandle(IntPtr handle);
  public static IntPtr Create(string name) {
    IntPtr job=CreateJobObject(IntPtr.Zero,name); if(job==IntPtr.Zero) throw new Win32Exception(Marshal.GetLastWin32Error());
    BASIC information=new BASIC(); information.LimitFlags=JOB_OBJECT_LIMIT_ACTIVE_PROCESS; information.ActiveProcessLimit=0;
    int length=Marshal.SizeOf(typeof(BASIC)); IntPtr pointer=Marshal.AllocHGlobal(length);
    try { Marshal.StructureToPtr(information,pointer,false); if(!SetInformationJobObject(job,2,pointer,(UInt32)length)){int error=Marshal.GetLastWin32Error();CloseHandle(job);throw new Win32Exception(error);} }
    finally { Marshal.FreeHGlobal(pointer); }
    return job;
  }
  public static void Close(IntPtr handle) { if(handle!=IntPtr.Zero) CloseHandle(handle); }
}
'@
$workingRoot = [Text.Encoding]::UTF8.GetString([Convert]::FromBase64String('%s'))
$jobName = '%s'
$rejectingHandle = [IntPtr]::Zero
$job = $null
try {
  %s
  $invocation = [pscustomobject]@{ Executable='docker'; Arguments=[string[]]@('version'); WorkingDirectory=$workingRoot; JobName=$jobName; Environment=[object[]]@() }
  $job = Start-Job -ArgumentList $invocation -ScriptBlock $script:c12ContainedNativeScript
  if ($null -eq (Wait-Job -Job $job -Timeout 15)) { throw 'contained child harness timed out' }
  $received = @(Receive-Job -Job $job -ErrorAction SilentlyContinue)
  $result = @($received | Where-Object { $_.PSObject.Properties.Name -contains 'ContainmentFailure' } | Select-Object -Last 1)
  if ($result.Count -ne 1 -or -not [bool]$result[0].ContainmentFailure -or [int]$result[0].ExitCode -ne 127) { throw 'contained child did not report exact containment failure' }
  exit 0
}
catch { [Console]::Error.WriteLine($_.Exception.Message); exit 1 }
finally {
  if ($null -ne $job) { Stop-Job -Job $job -ErrorAction SilentlyContinue; Remove-Job -Job $job -Force -ErrorAction SilentlyContinue }
  [C12RejectingJobFixture]::Close($rejectingHandle)
}
`, rootPayload, jobName, setup)
	harness := filepath.Join(t.TempDir(), "assert-c12-containment.ps1")
	if err := os.WriteFile(harness, append(append([]byte(nil), runner[:index]...), []byte(appendix)...), 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", harness,
		"-Profile", "base", "-Packages", "./internal/store", "-Timeout", "3m")
	command.Env = append([]string{}, os.Environ()...)
	command.Env = append(command.Env,
		"C12_CONTAINMENT_SENTINEL="+sentinel,
		"Path="+fakeBin+string(os.PathListSeparator)+os.Getenv("Path"))
	output, commandErr := command.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("containment failure harness timed out: %v\n%s", ctx.Err(), output)
	}
	if commandErr == nil {
		return string(output), 0
	}
	var exitError *exec.ExitError
	if !errors.As(commandErr, &exitError) {
		t.Fatalf("launch containment failure harness: %v", commandErr)
	}
	return string(output), exitError.ExitCode()
}

func runC12GoJSONAssertionHarness(t *testing.T, packageName string, expectedTests, events []string) (string, int) {
	t.Helper()
	runner, err := os.ReadFile("../../scripts/run-c12-integration.ps1")
	if err != nil {
		t.Fatal(err)
	}
	marker := []byte("$script:c12RepositoryRoot = (Resolve-Path")
	index := bytes.Index(runner, marker)
	if index < 0 {
		t.Fatal("runner lacks main-program marker")
	}
	eventPayload := base64.StdEncoding.EncodeToString([]byte(strings.Join(events, "\n")))
	expectedRaw, err := json.Marshal(expectedTests)
	if err != nil {
		t.Fatal(err)
	}
	expectedPayload := base64.StdEncoding.EncodeToString(expectedRaw)
	appendix := fmt.Sprintf(`
$eventText = [Text.Encoding]::UTF8.GetString([Convert]::FromBase64String('%s'))
$expectedText = [Text.Encoding]::UTF8.GetString([Convert]::FromBase64String('%s'))
$result = [pscustomobject]@{ ExitCode = 0; Output = @($eventText -split [char]10) }
$expected = @($expectedText | ConvertFrom-Json | ForEach-Object { [string]$_ })
try {
  Assert-C12GoJSONResult -Result $result -Package '%s' -ExpectedTests $expected
  exit 0
}
catch {
  [Console]::Error.WriteLine($_.Exception.Message)
  exit 1
}
`, eventPayload, expectedPayload, packageName)
	harness := filepath.Join(t.TempDir(), "assert-c12-json.ps1")
	if err := os.WriteFile(harness, append(append([]byte(nil), runner[:index]...), []byte(appendix)...), 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", harness,
		"-Profile", "base", "-Packages", "./internal/store", "-Timeout", "3m")
	output, commandErr := command.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("C12 JSON assertion harness timed out: %v\n%s", ctx.Err(), output)
	}
	if commandErr == nil {
		return string(output), 0
	}
	var exitError *exec.ExitError
	if !errors.As(commandErr, &exitError) {
		t.Fatalf("launch C12 JSON assertion harness: %v", commandErr)
	}
	return string(output), exitError.ExitCode()
}

func runC12PackageMembershipHarness(t *testing.T, packageName string) (string, int) {
	t.Helper()
	runner, err := os.ReadFile("../../scripts/run-c12-integration.ps1")
	if err != nil {
		t.Fatal(err)
	}
	marker := []byte("$script:c12RepositoryRoot = (Resolve-Path")
	index := bytes.Index(runner, marker)
	if index < 0 {
		t.Fatal("runner lacks main-program marker")
	}
	stub := []byte(`function Invoke-C12Group {
  param(
    [string]$GroupID,
    [string]$GroupProfile,
    [string]$Package,
    [string]$RunPattern = '',
    [string]$GroupTimeout,
    [string[]]$ExpectedTests = @(),
    [DateTime]$AbsoluteDeadline = [DateTime]::MaxValue
  )
  Write-Output "C12_GROUP_CALLED:$Package"
}

`)
	harnessRoot := t.TempDir()
	harness := filepath.Join(harnessRoot, "scripts", "run-c12-integration.ps1")
	if err := os.MkdirAll(filepath.Dir(harness), 0o700); err != nil {
		t.Fatal(err)
	}
	body := append(append(append([]byte(nil), runner[:index]...), stub...), runner[index:]...)
	if err := os.WriteFile(harness, body, 0o600); err != nil {
		t.Fatal(err)
	}
	return runC12PowerShellAtRoot(t, harnessRoot, nil,
		"-Profile", "base", "-Packages", packageName, "-Timeout", "3m")
}

func buildFakeGoExecutable(t *testing.T, executable, source string) {
	t.Helper()
	sourcePath := filepath.Join(t.TempDir(), "main.go")
	if err := os.WriteFile(sourcePath, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	command := task8ChildGoCommandContext(t, ctx, "", "build", "-o", executable, sourcePath)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build fake native executable: %v\n%s", err, output)
	}
}

func TestBuildFakeGoExecutableBindsDeterministicChildEnvironment(t *testing.T) {
	t.Setenv("GOOS", "linux")
	t.Setenv("GOARCH", "386")
	t.Setenv("CGO_ENABLED", "1")
	t.Setenv("GOFLAGS", "-overlay=ambient-untrusted-overlay.json")
	executable := filepath.Join(t.TempDir(), "fake-go.exe")
	buildFakeGoExecutable(t, executable, "package main\nimport \"fmt\"\nfunc main() { fmt.Print(\"TASK8_FAKE_GO_OK\") }\n")
	output, err := exec.Command(executable).CombinedOutput()
	if err != nil {
		t.Fatalf("run fake Go executable: %v\n%s", err, output)
	}
	if got := string(output); got != "TASK8_FAKE_GO_OK" {
		t.Fatalf("fake Go executable output = %q, want %q", got, "TASK8_FAKE_GO_OK")
	}
}

func task8ChildGoCommand(t *testing.T, directory string, arguments ...string) *exec.Cmd {
	t.Helper()
	command := exec.Command("go", arguments...)
	command.Dir = directory
	command.Env = task8ChildGoEnvironment(t)
	return command
}

func task8ChildGoCommandContext(t *testing.T, ctx context.Context, directory string, arguments ...string) *exec.Cmd {
	t.Helper()
	command := exec.CommandContext(ctx, "go", arguments...)
	command.Dir = directory
	command.Env = task8ChildGoEnvironment(t)
	return command
}

func task8ChildGoEnvironment(t *testing.T) []string {
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
	return append(environment,
		"GOOS=windows",
		"GOARCH=amd64",
		"CGO_ENABLED=0",
		"GOFLAGS=-overlay="+approvedOverlayPath,
		"GOPROXY=off",
		"GOWORK=off",
	)
}

func runC12PowerShell(t *testing.T, extraEnvironment map[string]string, arguments ...string) (string, int) {
	t.Helper()
	return runC12PowerShellWithTimeout(t, 20*time.Second, extraEnvironment, arguments...)
}

func runC12PowerShellWithTimeout(t *testing.T, limit time.Duration, extraEnvironment map[string]string, arguments ...string) (string, int) {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return runC12PowerShellAtRootWithTimeout(t, root, limit, extraEnvironment, arguments...)
}

func runC12PowerShellAtRoot(t *testing.T, root string, extraEnvironment map[string]string, arguments ...string) (string, int) {
	t.Helper()
	return runC12PowerShellAtRootWithTimeout(t, root, 20*time.Second, extraEnvironment, arguments...)
}

func runC12PowerShellAtRootWithTimeout(t *testing.T, root string, limit time.Duration, extraEnvironment map[string]string, arguments ...string) (string, int) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), limit)
	defer cancel()
	commandArguments := []string{"-NoProfile", "-ExecutionPolicy", "Bypass", "-File", filepath.Join("scripts", "run-c12-integration.ps1")}
	commandArguments = append(commandArguments, arguments...)
	command := exec.CommandContext(ctx, "powershell", commandArguments...)
	command.Dir = root
	environment := make([]string, 0, len(os.Environ())+len(extraEnvironment))
	for _, item := range os.Environ() {
		name := item
		if index := strings.IndexByte(item, '='); index >= 0 {
			name = item[:index]
		}
		if name == "TALENRO_DATABASE_URL" || name == "TALENRO_REDIS_ADDRESS" || name == "TALENRO_NATS_URL" {
			continue
		}
		overridden := false
		for override := range extraEnvironment {
			if strings.EqualFold(name, override) {
				overridden = true
				break
			}
		}
		if overridden {
			continue
		}
		if strings.HasPrefix(strings.ToUpper(name), "GIT_") {
			continue
		}
		environment = append(environment, item)
	}
	for name, value := range extraEnvironment {
		environment = append(environment, name+"="+value)
	}
	command.Env = environment
	output, commandErr := command.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("C12 PowerShell validation did not finish within %s: %v\n%s", limit, ctx.Err(), output)
	}
	if commandErr == nil {
		return string(output), 0
	}
	var exitError *exec.ExitError
	if !errors.As(commandErr, &exitError) {
		t.Fatalf("launch C12 PowerShell runner: %v", commandErr)
	}
	return string(output), exitError.ExitCode()
}

func validateC12IntegrationManifests(rawManifests [][]byte, sources []c12IntegrationSource, suite string, suiteTimeout time.Duration) error {
	wantSuiteTimeout, exists := map[string]time.Duration{
		"batch01": 120 * time.Minute,
		"batch02": 180 * time.Minute,
		"batch03": 240 * time.Minute,
		"final":   240 * time.Minute,
	}[suite]
	if !exists {
		return fmt.Errorf("unknown C12 integration suite %q", suite)
	}
	if suiteTimeout != wantSuiteTimeout {
		return fmt.Errorf("C12 %s suite timeout = %s, want exactly %s", suite, suiteTimeout, wantSuiteTimeout)
	}
	if len(rawManifests) == 0 {
		return fmt.Errorf("selected C12 integration manifest set is empty")
	}

	manifests := make([]c12IntegrationManifest, 0, len(rawManifests))
	for index, raw := range rawManifests {
		manifest, err := decodeC12IntegrationManifest(raw)
		if err != nil {
			return fmt.Errorf("manifest %d: %w", index+1, err)
		}
		manifests = append(manifests, manifest)
	}

	type testKey struct {
		pkg, test string
	}
	coverage := make(map[testKey]int)
	packageSet := make(map[string]struct{})
	groupIDs := make(map[string]struct{})
	var totalBudget time.Duration
	for manifestIndex, manifest := range manifests {
		if manifest.Schema != c12IntegrationManifestSchema {
			return fmt.Errorf("manifest %d schema = %q, want %q", manifestIndex+1, manifest.Schema, c12IntegrationManifestSchema)
		}
		if len(manifest.Groups) == 0 {
			return fmt.Errorf("manifest %d has no integration groups", manifestIndex+1)
		}
		previousID := ""
		for _, group := range manifest.Groups {
			if !c12ManifestIDPattern.MatchString(group.ID) {
				return fmt.Errorf("integration group ID %q is not closed and bounded", group.ID)
			}
			if previousID != "" && group.ID <= previousID {
				return fmt.Errorf("integration groups are not sorted by unique ID: %q follows %q", group.ID, previousID)
			}
			previousID = group.ID
			if _, duplicate := groupIDs[group.ID]; duplicate {
				return fmt.Errorf("duplicate integration group ID %q", group.ID)
			}
			groupIDs[group.ID] = struct{}{}
			if !c12PackagePattern.MatchString(group.Package) || group.Package == "./..." || strings.Contains(group.Package, "//") {
				return fmt.Errorf("integration group %s package %q is not one explicit package", group.ID, group.Package)
			}
			if _, allowed := c12Task4AllowedPackages[group.Package]; !allowed {
				return fmt.Errorf("integration group %s package %q is outside the Task 4 allowed package set", group.ID, group.Package)
			}
			packageSet[group.Package] = struct{}{}
			setupBudget := 3 * time.Minute
			switch group.Profile {
			case "base", "authority-v7":
			case "authority-v7-pitr":
				setupBudget = 8 * time.Minute
			default:
				return fmt.Errorf("integration group %s has unsupported profile %q", group.ID, group.Profile)
			}
			if !c12TimeoutPattern.MatchString(group.Timeout) {
				return fmt.Errorf("integration group %s timeout %q is not a finite seconds/minutes duration", group.ID, group.Timeout)
			}
			groupTimeout, err := time.ParseDuration(group.Timeout)
			if err != nil || groupTimeout <= 0 {
				return fmt.Errorf("integration group %s timeout %q is invalid", group.ID, group.Timeout)
			}
			if groupTimeout > 30*time.Minute {
				return fmt.Errorf("integration group %s timeout %s exceeds 30m", group.ID, groupTimeout)
			}
			totalBudget += groupTimeout + setupBudget
			if len(group.Tests) == 0 {
				return fmt.Errorf("integration group %s has no exact tests", group.ID)
			}
			previousTest := ""
			for _, testName := range group.Tests {
				if !c12TestNamePattern.MatchString(testName) {
					return fmt.Errorf("integration group %s test %q is not a literal top-level test", group.ID, testName)
				}
				if testName == "TestPrepareC12AuthorityV7Database" {
					return fmt.Errorf("profile initializer %s is forbidden in integration manifests", testName)
				}
				if previousTest != "" && testName < previousTest {
					return fmt.Errorf("integration group %s tests are not sorted", group.ID)
				}
				if testName == previousTest {
					return fmt.Errorf("integration group %s contains duplicate test %s", group.ID, testName)
				}
				previousTest = testName
				coverage[testKey{pkg: group.Package, test: testName}]++
			}
		}
	}
	if totalBudget > suiteTimeout {
		return fmt.Errorf("C12 %s manifest budget %s exceeds suite budget %s", suite, totalBudget, suiteTimeout)
	}

	universe := make(map[testKey]string)
	for _, source := range sources {
		if !source.Tracked || !strings.HasSuffix(filepath.ToSlash(source.Path), "_test.go") || !hasLiteralC12IntegrationFirstLine(source.Body) {
			continue
		}
		normalizedPath := path.Clean(filepath.ToSlash(source.Path))
		if strings.HasPrefix(normalizedPath, "../") || path.IsAbs(normalizedPath) {
			return fmt.Errorf("tracked integration source path %q escapes the repository", source.Path)
		}
		packageName := "./" + path.Dir(normalizedPath)
		if _, selected := packageSet[packageName]; !selected {
			continue
		}
		fileSet := token.NewFileSet()
		parsed, err := parser.ParseFile(fileSet, normalizedPath, source.Body, parser.SkipObjectResolution)
		if err != nil {
			return fmt.Errorf("parse tracked integration source %s: %w", normalizedPath, err)
		}
		for _, declaration := range parsed.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || function.Recv != nil || !c12TestNamePattern.MatchString(function.Name.Name) {
				continue
			}
			if function.Name.Name == "TestPrepareC12AuthorityV7Database" {
				continue
			}
			if functionCallsTestingSkip(function) {
				return fmt.Errorf("manifest-selected test %s.%s calls testing Skip", packageName, function.Name.Name)
			}
			key := testKey{pkg: packageName, test: function.Name.Name}
			if previous, duplicate := universe[key]; duplicate {
				return fmt.Errorf("duplicate tracked integration test %s.%s in %s and %s", packageName, function.Name.Name, previous, normalizedPath)
			}
			universe[key] = normalizedPath
		}
	}
	for key, count := range coverage {
		if count > 1 {
			return fmt.Errorf("duplicate manifest coverage for %s.%s", key.pkg, key.test)
		}
		if _, exists := universe[key]; !exists {
			return fmt.Errorf("extra manifest test %s.%s has no tracked first-line integration source", key.pkg, key.test)
		}
	}
	for key, sourcePath := range universe {
		if coverage[key] == 0 {
			return fmt.Errorf("missing manifest coverage for %s.%s from %s", key.pkg, key.test, sourcePath)
		}
	}
	return nil
}

func resolveC12FocusedTestMap(sources []c12IntegrationSource, packages, requested []string) (map[string][]string, error) {
	selected := make(map[string]struct{}, len(packages))
	resolved := make(map[string][]string, len(packages))
	for _, packageName := range packages {
		if _, allowed := c12Task4AllowedPackages[packageName]; !allowed {
			return nil, fmt.Errorf("focused package %q is outside the Task 4 allowed package set", packageName)
		}
		if _, duplicate := selected[packageName]; duplicate {
			return nil, fmt.Errorf("duplicate focused package %q", packageName)
		}
		selected[packageName] = struct{}{}
		resolved[packageName] = nil
	}
	requestedSet := make(map[string]struct{}, len(requested))
	for _, testName := range requested {
		if !c12TestNamePattern.MatchString(testName) {
			return nil, fmt.Errorf("focused requested test %q is not a literal top-level test", testName)
		}
		if _, duplicate := requestedSet[testName]; duplicate {
			return nil, fmt.Errorf("duplicate focused requested test %s", testName)
		}
		requestedSet[testName] = struct{}{}
	}

	definitions := make(map[string][]string)
	seenDefinition := make(map[string]string)
	for _, source := range sources {
		if !source.Tracked || !hasLiteralC12IntegrationFirstLine(source.Body) {
			continue
		}
		normalized := path.Clean(filepath.ToSlash(source.Path))
		packageName := "./" + path.Dir(normalized)
		if _, wantedPackage := selected[packageName]; !wantedPackage {
			continue
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), normalized, source.Body, parser.SkipObjectResolution)
		if err != nil {
			return nil, fmt.Errorf("parse focused integration source %s: %w", normalized, err)
		}
		for _, declaration := range parsed.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || function.Recv != nil || !c12TestNamePattern.MatchString(function.Name.Name) {
				continue
			}
			key := packageName + "\x00" + function.Name.Name
			if previous, duplicate := seenDefinition[key]; duplicate {
				return nil, fmt.Errorf("ambiguous focused test %s in %s and %s", function.Name.Name, previous, normalized)
			}
			seenDefinition[key] = normalized
			definitions[function.Name.Name] = append(definitions[function.Name.Name], packageName)
		}
	}
	for _, testName := range requested {
		owners := definitions[testName]
		if len(owners) == 0 {
			return nil, fmt.Errorf("focused requested test %s is missing from selected integration packages", testName)
		}
		if len(owners) != 1 {
			return nil, fmt.Errorf("focused requested test %s is ambiguous across selected integration packages", testName)
		}
		resolved[owners[0]] = append(resolved[owners[0]], testName)
	}
	for _, packageName := range packages {
		if len(resolved[packageName]) == 0 {
			return nil, fmt.Errorf("focused package %s has no requested local integration test", packageName)
		}
		sort.Strings(resolved[packageName])
	}
	return resolved, nil
}

func decodeC12IntegrationManifest(raw []byte) (c12IntegrationManifest, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var manifest c12IntegrationManifest
	if err := decoder.Decode(&manifest); err != nil {
		return c12IntegrationManifest{}, fmt.Errorf("decode strict integration manifest: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return c12IntegrationManifest{}, fmt.Errorf("integration manifest contains a second JSON value")
		}
		return c12IntegrationManifest{}, fmt.Errorf("decode integration manifest trailer: %w", err)
	}
	return manifest, nil
}

func c12ManifestPackages(rawManifests [][]byte) (map[string]struct{}, error) {
	packages := make(map[string]struct{})
	for _, raw := range rawManifests {
		manifest, err := decodeC12IntegrationManifest(raw)
		if err != nil {
			return nil, err
		}
		for _, group := range manifest.Groups {
			packages[group.Package] = struct{}{}
		}
	}
	return packages, nil
}

func loadCandidateC12IntegrationSources(candidateRoot string, packages map[string]struct{}) ([]c12IntegrationSource, error) {
	paths := make([]string, 0)
	for packageName := range packages {
		if !c12PackagePattern.MatchString(packageName) || packageName == "./..." || strings.Contains(packageName, "//") {
			return nil, fmt.Errorf("candidate package %q is not one explicit package", packageName)
		}
		packageDirectory := filepath.Join(candidateRoot, filepath.FromSlash(strings.TrimPrefix(packageName, "./")))
		walkErr := filepath.WalkDir(packageDirectory, func(sourcePath string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				if sourcePath != packageDirectory {
					return fs.SkipDir
				}
				return nil
			}
			if strings.HasSuffix(entry.Name(), "_test.go") {
				relative, err := filepath.Rel(candidateRoot, sourcePath)
				if err != nil {
					return err
				}
				paths = append(paths, filepath.ToSlash(relative))
			}
			return nil
		})
		if walkErr != nil {
			return nil, fmt.Errorf("walk candidate integration package %s: %w", packageName, walkErr)
		}
	}
	sort.Strings(paths)
	sources := make([]c12IntegrationSource, 0, len(paths))
	for _, sourcePath := range paths {
		body, readErr := os.ReadFile(filepath.Join(candidateRoot, filepath.FromSlash(sourcePath)))
		if readErr != nil {
			return nil, fmt.Errorf("read candidate integration source %s: %w", sourcePath, readErr)
		}
		sources = append(sources, c12IntegrationSource{Path: sourcePath, Tracked: true, Body: body})
	}
	return sources, nil
}

func hasLiteralC12IntegrationFirstLine(body []byte) bool {
	line := body
	if newline := bytes.IndexByte(body, '\n'); newline >= 0 {
		line = body[:newline]
	}
	line = bytes.TrimSuffix(line, []byte{'\r'})
	return bytes.Equal(line, []byte("//go:build integration"))
}

func functionCallsTestingSkip(function *ast.FuncDecl) bool {
	parameterNames := make(map[string]struct{})
	if function.Type.Params != nil {
		for _, field := range function.Type.Params.List {
			for _, name := range field.Names {
				parameterNames[name.Name] = struct{}{}
			}
		}
	}
	found := false
	ast.Inspect(function.Body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || (selector.Sel.Name != "Skip" && selector.Sel.Name != "Skipf" && selector.Sel.Name != "SkipNow") {
			return true
		}
		receiver, ok := selector.X.(*ast.Ident)
		if !ok {
			return true
		}
		if _, isParameter := parameterNames[receiver.Name]; isParameter {
			found = true
			return false
		}
		return true
	})
	return found
}
