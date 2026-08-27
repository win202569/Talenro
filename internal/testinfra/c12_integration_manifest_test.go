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
	"net"
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

func TestC12BaseRunnerCleanupBindsExactOwnedDirectoryIdentity(t *testing.T) {
	output, exitCode := runC12OwnedDirectoryIdentityHarness(t)
	if exitCode != 0 {
		t.Fatalf("owned directory identity harness exit=%d output=%q", exitCode, output)
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
	files := map[string][]byte{
		"scripts/run-c12-integration.ps1":                             runner,
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
 encoder:=json.NewEncoder(os.Stdout);encoder.Encode(map[string]string{"Action":"run","Package":stage,"Test":testName});encoder.Encode(map[string]string{"Action":"pass","Package":stage,"Test":testName});encoder.Encode(map[string]string{"Action":"pass","Package":stage})
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
 if len(args)>1&&args[0]=="container"&&args[1]=="port"{role:=roleOf(args[len(args)-2]);port:=0;if role=="postgres"{port=15432}else if role=="redis"{port,_=strconv.Atoi(os.Getenv("C12_REDIS_PORT"))}else{port,_=strconv.Atoi(os.Getenv("C12_NATS_PORT"))};fmt.Printf("127.0.0.1:%d\n",port);return}
 if len(args)>1&&args[0]=="container"&&args[1]=="stop"{return};if len(args)>1&&args[0]=="container"&&args[1]=="rm"{os.WriteFile(state(roleOf(last),".removed"),[]byte("removed"),0600);return};os.Exit(74)
}`
	buildFakeGoExecutable(t, filepath.Join(fakeBin, "go.exe"), fakeGoSource)
	buildFakeGoExecutable(t, filepath.Join(fakeBin, "docker.exe"), fakeDockerSource)
	redisPort := startC12ProtocolServer(t, "+PONG\r\n")
	natsPort := startC12ProtocolServer(t, "PONG\r\n")
	dockerState := t.TempDir()
	output, exitCode := runC12PowerShellAtRootWithTimeout(t, repository, 90*time.Second, map[string]string{
		"C12_DOCKER_STATE":            dockerState,
		"C12_FAKE_GO_LOG":             goLog,
		"C12_LATER_MUTATION_SENTINEL": laterMutationSentinel,
		"C12_NATS_PORT":               fmt.Sprint(natsPort),
		"C12_REDIS_PORT":              fmt.Sprint(redisPort),
		"C12_SNAPSHOT_REAL_GO":        realGo,
		"C12_SNAPSHOT_MUTATION":       mutationMode,
		"Path":                        fakeBin + string(os.PathListSeparator) + os.Getenv("Path"),
	}, "-Suite", "batch01", "-Timeout", "120m")
	return c12SnapshotSuiteRun{repository: repository, objectsRoot: objectsRoot, objectsBefore: objectsBefore, goLog: goLog, laterMutationSentinel: laterMutationSentinel, output: output, exitCode: exitCode}
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
