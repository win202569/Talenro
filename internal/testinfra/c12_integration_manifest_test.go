package testinfra_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
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
		"TestC12SelectedIntegrationManifest",
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

func TestC12Task4AllowedPackagesStaySynchronizedWithRunner(t *testing.T) {
	raw, err := os.ReadFile("../../scripts/run-c12-integration.ps1")
	if err != nil {
		t.Fatal(err)
	}
	matches := regexp.MustCompile(`(?m)^  '(\./[A-Za-z0-9_./-]+)' = \$true$`).FindAllStringSubmatch(string(raw), -1)
	runnerPackages := make(map[string]struct{}, len(matches))
	for _, match := range matches {
		runnerPackages[match[1]] = struct{}{}
	}
	if len(runnerPackages) != len(c12Task4AllowedPackages) {
		t.Fatalf("runner allowed package count = %d, validator count = %d", len(runnerPackages), len(c12Task4AllowedPackages))
	}
	for packageName := range c12Task4AllowedPackages {
		if _, exists := runnerPackages[packageName]; !exists {
			t.Errorf("runner and validator allowed package sets differ at %s", packageName)
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
	fakeDocker := filepath.Join(fakeBin, "docker.cmd")
	const fakeDockerSource = `@echo off
if "%1"=="image" if "%2"=="inspect" (
  echo sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
  exit /b 0
)
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

	started := time.Now()
	output, exitCode := runC12PowerShellWithTimeout(t, 30*time.Second, map[string]string{
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

func TestC12BaseRunnerDeclaresExactGroupAndSuiteDeadlines(t *testing.T) {
	raw, err := os.ReadFile("../../scripts/run-c12-integration.ps1")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	for _, literal := range []string{
		"$script:c12SuiteDeadline = [DateTime]::UtcNow.AddMinutes(120)",
		"$groupDeadline = [DateTime]::UtcNow.Add($groupDuration).AddMinutes(3)",
		"Invoke-C12Native",
		"Wait-Job -Job $job -Timeout",
		"JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE",
		"[C12NativeJob]::Assign($nativeJobHandle, $jobHostID)",
		"[IO.File]::WriteAllText($releasePath, 'contained')",
	} {
		if !strings.Contains(source, literal) {
			t.Errorf("runner lacks enforced deadline literal %q", literal)
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

func TestC12BaseRunnerRejectsInheritedGitCandidateOverrides(t *testing.T) {
	const canary = `C:\secret-candidate-index`
	output, exitCode := runC12PowerShell(t, map[string]string{"GIT_INDEX_FILE": canary},
		"-Profile", "base", "-Packages", "./...", "-Run", "^TestAlpha$", "-Timeout", "3m")
	if exitCode == 0 || !strings.Contains(output, "inherited Git candidate overrides") {
		t.Fatalf("exit=%d output=%q, want inherited Git override rejection", exitCode, output)
	}
	if strings.Contains(output, canary) || strings.Contains(output, "secret-candidate-index") {
		t.Fatalf("runner leaked inherited Git override canary: %q", output)
	}
}

func TestC12Batch01SuiteFailsClosedWhileCanonicalManifestIsAbsent(t *testing.T) {
	output, exitCode := runC12PowerShell(t, nil, "-Suite", "batch01", "-Timeout", "120m")
	if exitCode == 0 || !strings.Contains(output, "canonical Batch 01 manifest is absent") {
		t.Fatalf("exit=%d output=%q, want absent canonical manifest rejection", exitCode, output)
	}
}

func TestC12Batch01SuiteUsesOneStagedCandidateSnapshot(t *testing.T) {
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
	manifestPath := filepath.Join(repository, "testdata", "c12", "integration-contracts-schema-authority.v1.json")
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0o700); err != nil {
		t.Fatal(err)
	}
	manifest := `{"schema":"talenro-c12-integration-manifest/v1","groups":[{"id":"base-testinfra","package":"./internal/testinfra","profile":"base","tests":["TestC12DependenciesAreIsolatedAndBaseMigrated"],"timeout":"2m"}]}`
	if err := os.WriteFile(manifestPath, []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, arguments := range [][]string{{"init", "--quiet"}, {"add", "--", "scripts/run-c12-integration.ps1", "testdata/c12/integration-contracts-schema-authority.v1.json"}} {
		command := exec.Command("git", arguments...)
		command.Dir = repository
		if output, commandErr := command.CombinedOutput(); commandErr != nil {
			t.Fatalf("git %v: %v\n%s", arguments, commandErr, output)
		}
	}
	objectsRoot := filepath.Join(repository, ".git", "objects")
	objectsBefore := relativeFileSet(t, objectsRoot)
	if err := os.Remove(manifestPath); err != nil {
		t.Fatal(err)
	}

	fakeBin := t.TempDir()
	goLog := filepath.Join(t.TempDir(), "go.log")
	if err := os.WriteFile(filepath.Join(fakeBin, "go.cmd"), []byte("@echo off\r\n>>\"%C12_FAKE_GO_LOG%\" echo %CD%^|%*\r\nexit /b 0\r\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fakeBin, "docker.cmd"), []byte("@echo off\r\nexit /b 73\r\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	output, exitCode := runC12PowerShellAtRoot(t, repository, map[string]string{
		"C12_FAKE_GO_LOG": goLog,
		"Path":            fakeBin + string(os.PathListSeparator) + os.Getenv("Path"),
	}, "-Suite", "batch01", "-Timeout", "120m")
	if exitCode == 0 {
		t.Fatalf("exit=%d output=%q, want bounded fake Docker failure", exitCode, output)
	}
	if strings.Contains(output, "canonical Batch 01 manifest is absent") {
		t.Fatalf("suite reread the mutable working manifest instead of its staged candidate: %q", output)
	}
	logged, err := os.ReadFile(goLog)
	if err != nil {
		t.Fatalf("validator did not run from candidate snapshot: %v; output=%q", err, output)
	}
	logText := filepath.Clean(strings.TrimSpace(string(logged)))
	if strings.HasPrefix(strings.ToLower(logText), strings.ToLower(filepath.Clean(repository)+string(os.PathSeparator))) ||
		!strings.Contains(logText, "-c12-candidate-tree") {
		t.Fatalf("validator invocation did not identify and consume a separate candidate snapshot: %q", logText)
	}
	if objectsAfter := relativeFileSet(t, objectsRoot); !equalStringSets(objectsBefore, objectsAfter) {
		t.Fatalf("candidate materialization wrote to common Git objects: before=%v after=%v", objectsBefore, objectsAfter)
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

func buildFakeGoExecutable(t *testing.T, executable, source string) {
	t.Helper()
	sourcePath := filepath.Join(t.TempDir(), "main.go")
	if err := os.WriteFile(sourcePath, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "build", "-o", executable, sourcePath)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build fake native executable: %v\n%s", err, output)
	}
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
