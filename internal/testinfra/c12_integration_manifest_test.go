package testinfra_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
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
	c12ManifestIDPattern    = regexp.MustCompile(`^[a-z][a-z0-9-]{0,63}$`)
	c12PackagePattern       = regexp.MustCompile(`^\./(?:[A-Za-z0-9_.-]+/)*[A-Za-z0-9_.-]+$`)
	c12TestNamePattern      = regexp.MustCompile(`^Test[A-Za-z0-9_]+$`)
	c12TimeoutPattern       = regexp.MustCompile(`^[1-9][0-9]*(?:s|m)$`)
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

func TestC12SelectedIntegrationManifest(t *testing.T) {
	if *c12SelectedSuite == "" && *c12SelectedManifests == "" && *c12SelectedSuiteTimeout == "" {
		raw, sources := validC12ManifestFixture(t)
		if err := validateC12IntegrationManifests([][]byte{raw}, sources, "batch01", 120*time.Minute); err != nil {
			t.Fatal(err)
		}
		return
	}
	if *c12SelectedSuite == "" || *c12SelectedManifests == "" || *c12SelectedSuiteTimeout == "" {
		t.Fatal("selected C12 manifest validation requires suite, manifests, and suite timeout together")
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
	sources, err := loadTrackedC12IntegrationSources(repositoryRoot, packages)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateC12IntegrationManifests(rawManifests, sources, *c12SelectedSuite, suiteTimeout); err != nil {
		t.Fatal(err)
	}
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
		"TestC12SelectedIntegrationManifest",
	}
	for _, literal := range required {
		if !strings.Contains(source, literal) {
			t.Errorf("runner lacks required closed-boundary literal %q", literal)
		}
	}
	for _, forbidden := range []string{"Invoke-Expression", "cmd /c", "powershell -Command", ".Arguments"} {
		if strings.Contains(strings.ToLower(source), strings.ToLower(forbidden)) {
			t.Errorf("runner contains forbidden command construction %q", forbidden)
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
}

func TestC12BaseRunnerCapturesNativeFailuresBeforeCleanup(t *testing.T) {
	fakeBin := t.TempDir()
	fakeDocker := filepath.Join(fakeBin, "docker.cmd")
	const fakeDockerSource = `@echo off
if "%1"=="run" (
  >&2 echo C12_NATIVE_STDERR_CANARY
  exit /b 73
)
if "%1"=="container" if "%2"=="inspect" (
  >&2 echo Error: No such container
  exit /b 1
)
>&2 echo unexpected docker invocation
exit /b 74
`
	if err := os.WriteFile(fakeDocker, []byte(fakeDockerSource), 0o600); err != nil {
		t.Fatal(err)
	}

	output, exitCode := runC12PowerShell(t, map[string]string{
		"Path": fakeBin + string(os.PathListSeparator) + os.Getenv("Path"),
	}, "-Profile", "base", "-Packages", "./internal/testinfra", "-Run", "^TestC12DependenciesAreIsolatedAndBaseMigrated$", "-Timeout", "3m")
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
		{name: "future profile", args: []string{"-Profile", "authority-v7", "-Packages", "./internal/store", "-Run", "^TestAlpha$", "-Timeout", "3m"}, wantErr: "only the base profile"},
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

func TestC12Batch01SuiteFailsClosedWhileCanonicalManifestIsAbsent(t *testing.T) {
	output, exitCode := runC12PowerShell(t, nil, "-Suite", "batch01", "-Timeout", "120m")
	if exitCode == 0 || !strings.Contains(output, "canonical Batch 01 manifest is absent") {
		t.Fatalf("exit=%d output=%q, want absent canonical manifest rejection", exitCode, output)
	}
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

func runC12PowerShell(t *testing.T, extraEnvironment map[string]string, arguments ...string) (string, int) {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
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
		environment = append(environment, item)
	}
	for name, value := range extraEnvironment {
		environment = append(environment, name+"="+value)
	}
	command.Env = environment
	output, commandErr := command.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("C12 PowerShell validation did not finish within 20 seconds: %v\n%s", ctx.Err(), output)
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

func loadTrackedC12IntegrationSources(repositoryRoot string, packages map[string]struct{}) ([]c12IntegrationSource, error) {
	command := exec.Command("git", "ls-files", "--", "*_test.go")
	command.Dir = repositoryRoot
	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("list tracked integration sources: %w", err)
	}
	paths := strings.Fields(strings.ReplaceAll(string(output), "\r\n", "\n"))
	sort.Strings(paths)
	sources := make([]c12IntegrationSource, 0, len(paths))
	for _, sourcePath := range paths {
		normalized := filepath.ToSlash(sourcePath)
		packageName := "./" + path.Dir(normalized)
		if _, selected := packages[packageName]; !selected {
			continue
		}
		body, readErr := os.ReadFile(filepath.Join(repositoryRoot, filepath.FromSlash(normalized)))
		if readErr != nil {
			return nil, fmt.Errorf("read tracked integration source %s: %w", normalized, readErr)
		}
		sources = append(sources, c12IntegrationSource{Path: normalized, Tracked: true, Body: body})
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
