[CmdletBinding(DefaultParameterSetName = 'Focused')]
param(
  [Parameter(Mandatory = $true, ParameterSetName = 'Focused')]
  [string]$Profile,

  [Parameter(Mandatory = $true, ParameterSetName = 'Focused')]
  [string]$Packages,

  [Parameter(ParameterSetName = 'Focused')]
  [string]$Run = '',

  [Parameter(ParameterSetName = 'Focused')]
  [switch]$Race,

  [Parameter(ParameterSetName = 'Focused')]
  [string]$PITRFailureSeam = '',

  [Parameter(Mandatory = $true, ParameterSetName = 'Suite')]
  [string]$Suite,

  [Parameter(Mandatory = $true, ParameterSetName = 'Focused')]
  [Parameter(Mandatory = $true, ParameterSetName = 'Suite')]
  [string]$Timeout
)

Set-StrictMode -Version 2.0
$ErrorActionPreference = 'Stop'
$script:c12DockerEndpointReceipt = $null
$script:c12DockerImageReceipts = @{}

$script:c12AllowedPackages = [System.Collections.Generic.Dictionary[string,string]]::new([System.StringComparer]::Ordinal)
$script:c12AllowedPackages.Add('./internal/testinfra', 'talenro.local/platform/internal/testinfra')
$script:c12AllowedPackages.Add('./internal/store', 'talenro.local/platform/internal/store')
$script:c12AllowedPackages.Add('./internal/nodecontrol/contracts', 'talenro.local/platform/internal/nodecontrol/contracts')
$script:c12AllowedPackages.Add('./internal/nodecontrol/authority', 'talenro.local/platform/internal/nodecontrol/authority')
$script:c12AllowedPackages.Add('./internal/nodecontrol/serving', 'talenro.local/platform/internal/nodecontrol/serving')
$script:c12AllowedPackages.Add('./internal/readiness', 'talenro.local/platform/internal/readiness')
$script:c12AllowedProfiles = @('base', 'authority-v7', 'authority-v7-pitr')
$script:c12ProfileAllowances = @{
  'base' = [TimeSpan]::FromMinutes(3)
  'authority-v7' = [TimeSpan]::FromMinutes(3)
  'authority-v7-pitr' = [TimeSpan]::FromMinutes(8)
}
$script:c12PITRPrivateTest = 'TestC12AuthorityPITROwnershipWALFailureSeam'
$script:c12PITRPublicTest = 'TestC12AuthorityPITRProfile'
$script:c12AuthorityV7PublicTest = 'TestC12DependenciesAreIsolatedAndAuthorityV7Migrated'
$script:c12PITRFailureSeams = @(
  'after-intent-before-create',
  'after-create-before-actual',
  'after-actual-before-return',
  'after-clean-intent-before-remove',
  'after-remove-before-clean-result'
)
$script:c12PITROWnerWALSchema = 'talenro-c12-authority-pitr-ownership-wal/v1'
$script:c12PITROWnerWALEvents = @('BOOTSTRAP', 'INTENT', 'ACTUAL', 'RECOVERED_ACTUAL', 'NOT_FOUND', 'TRANSITION', 'CLEAN_INTENT', 'CLEAN_RESULT')
$script:c12PITRWALMaximumLineBytes = 8192
$script:c12PITRWALMaximumRecords = 4096
$script:c12PITRHMACType = [Security.Cryptography.HMACSHA256]
$script:c12PITRPlugin = 'test_decoding'
$script:c12PITRForeignCanary = 'foreign canary'
$script:c12NativeDeadline = [DateTime]::MaxValue
$script:c12GroupCleanupBudget = [TimeSpan]::FromSeconds(75)
$script:c12SuiteDeadline = [DateTime]::MaxValue
$script:c12PreparedAuthorityRoles = @('trusted-validator', 'authority-initializer-json')
$script:c12PreparedReceiptPayloadFields = @(
  'Schema', 'Role', 'RunSuffix', 'Profile', 'Purpose', 'OneShotNonce',
  'OwnedRootPath', 'ArtifactRootIdentity', 'ParentRootPath', 'ParentIdentity',
  'SourceIdentity', 'SourceDigest', 'CandidateTreeIdentity', 'BuildArgumentsDigest',
  'GoExecutablePath', 'GoExecutableSHA256', 'GoVersion', 'GOOS', 'GOARCH',
  'ExecutablePath', 'ExecutableSHA256', 'ExecutableLength', 'ExecutableVolumeSerial',
  'ExecutableFileIndex', 'ExecutableLinkCount', 'ExecutableOwner', 'ExecutableDACL', 'ExecutableReparse',
  'SecondaryExecutablePath', 'SecondaryExecutableSHA256', 'SecondaryExecutableLength',
  'SecondaryExecutableVolumeSerial', 'SecondaryExecutableFileIndex', 'SecondaryExecutableLinkCount',
  'SecondaryExecutableOwner', 'SecondaryExecutableDACL', 'SecondaryExecutableReparse',
  'ArgumentsDigest', 'WorkingDirectory', 'NativeJobName', 'GateName', 'PipeName', 'GoGraphReceipts'
)
$script:c12PreparedReceiptFields = @(
  'Arguments', 'ArgumentsDigest', 'ArtifactRootIdentity', 'BuildArguments', 'BuildArgumentsDigest',
  'CandidateTreeIdentity', 'ExecutableDACL', 'ExecutableFileIndex', 'ExecutableLength', 'ExecutableLinkCount',
  'ExecutableOwner', 'ExecutablePath', 'ExecutableReparse', 'ExecutableSHA256', 'ExecutableVolumeSerial', 'GateName',
  'GOARCH', 'GoExecutablePath', 'GoExecutableSHA256', 'GoVersion', 'GOOS', 'GoGraphReceipts', 'NativeJobName', 'OneShotNonce',
  'OwnedRootPath', 'ParentIdentity', 'ParentRootPath', 'PipeName', 'Profile', 'Purpose', 'ReceiptIdentity', 'ReceiptSeal',
  'Role', 'RunSuffix', 'Schema', 'SecondaryExecutableDACL', 'SecondaryExecutableFileIndex', 'SecondaryExecutableLength',
  'SecondaryExecutableLinkCount', 'SecondaryExecutableOwner', 'SecondaryExecutablePath', 'SecondaryExecutableReparse',
  'SecondaryExecutableSHA256', 'SecondaryExecutableVolumeSerial', 'SourceDigest', 'SourceIdentity', 'WorkingDirectory'
)
$script:c12PreparedArtifactRoot = $null
$script:c12PreparedReceipts = [System.Collections.Generic.Dictionary[string,object]]::new([StringComparer]::Ordinal)
$script:c12PreparedWorkers = New-Object 'System.Collections.Generic.List[object]'
$script:c12PreparedReceiptKeyHex = ''
$script:c12GoToolchainVersion = 'go1.26.5'
if ($PSCmdlet.ParameterSetName -eq 'Suite') {
  $script:c12SuiteDeadline = [DateTime]::UtcNow.AddMinutes(120)
}
$script:c12TrustedValidatorVersion = 'talenro-c12-trusted-validator/v1'
$script:c12TrustedValidatorSource = @'
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const validatorVersion = "talenro-c12-trusted-validator/v1"
const manifestSchema = "talenro-c12-integration-manifest/v1"

type manifest struct {
	Schema string  `json:"schema"`
	Groups []group `json:"groups"`
}

type group struct {
	ID string `json:"id"`
	Package string `json:"package"`
	Profile string `json:"profile"`
	Tests []string `json:"tests"`
	Timeout string `json:"timeout"`
}

type source struct { Path string; Body []byte }
type testKey struct { pkg, test string }

var allowed = map[string]struct{}{
	"./internal/testinfra": {}, "./internal/store": {},
	"./internal/nodecontrol/contracts": {}, "./internal/nodecontrol/authority": {},
	"./internal/nodecontrol/serving": {}, "./internal/readiness": {},
}
var idPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{0,63}$`)
var packagePattern = regexp.MustCompile(`^\./(?:[A-Za-z0-9_.-]+/)*[A-Za-z0-9_.-]+$`)
var testPattern = regexp.MustCompile(`^Test[A-Za-z0-9_]+$`)
var timeoutPattern = regexp.MustCompile(`^[1-9][0-9]*(?:s|m)$`)
var treePattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

func main() {
	mode := flag.String("mode", "", "closed validator mode")
	version := flag.String("version", "", "trusted validator version")
	sourceDigest := flag.String("source-digest", "", "trusted validator embedded-source digest")
	rootFlag := flag.String("root", "", "candidate data root")
	suite := flag.String("suite", "", "suite")
	suiteTimeout := flag.String("suite-timeout", "", "suite timeout")
	tree := flag.String("candidate-tree", "", "candidate tree")
	manifestList := flag.String("manifests", "", "manifest list")
	packageList := flag.String("packages", "", "focused packages")
	testList := flag.String("tests", "", "focused tests")
	flag.Parse()
	if flag.NArg() != 0 { fail(fmt.Errorf("unexpected positional validator input")) }
	if *version != validatorVersion { fail(fmt.Errorf("trusted validator version mismatch")) }
	if len(*sourceDigest) != 64 || strings.Trim(*sourceDigest, "0123456789abcdef") != "" { fail(fmt.Errorf("trusted validator source digest mismatch")) }
	root, err := filepath.Abs(*rootFlag)
	if err != nil { fail(err) }
	info, err := os.Lstat(root)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 { fail(fmt.Errorf("candidate data root is not one exact directory")) }
	switch *mode {
	case "suite":
		if !treePattern.MatchString(*tree) { fail(fmt.Errorf("candidate tree is not an exact identity")) }
		timeout, err := time.ParseDuration(*suiteTimeout)
		if err != nil { fail(fmt.Errorf("parse suite timeout: %w", err)) }
		manifestPaths, err := splitClosed(*manifestList)
		if err != nil { fail(err) }
		raw := make([][]byte, 0, len(manifestPaths))
		for _, relative := range manifestPaths {
			full, err := closedPath(root, relative)
			if err != nil { fail(err) }
			body, err := os.ReadFile(full)
			if err != nil { fail(fmt.Errorf("read manifest %s: %w", relative, err)) }
			raw = append(raw, body)
		}
		packages, err := manifestPackages(raw)
		if err != nil { fail(err) }
		sources, err := loadSources(root, packages)
		if err != nil { fail(err) }
		if err := validate(raw, sources, *suite, timeout); err != nil { fail(err) }
	case "focused":
		packages, err := splitClosed(*packageList)
		if err != nil { fail(err) }
		tests, err := splitClosed(*testList)
		if err != nil { fail(err) }
		selected := make(map[string]struct{}, len(packages))
		for _, pkg := range packages { selected[pkg] = struct{}{} }
		sources, err := loadSources(root, selected)
		if err != nil { fail(err) }
		mapping, err := resolveFocused(sources, packages, tests)
		if err != nil { fail(err) }
		raw, err := json.Marshal(mapping)
		if err != nil { fail(err) }
		fmt.Printf("C12_FOCUSED_MAP:%s\n", raw)
	default:
		fail(fmt.Errorf("unknown trusted validator mode"))
	}
	fmt.Printf("C12_TRUSTED_VALIDATOR_OK:%s:%s\n", validatorVersion, *sourceDigest)
}

func fail(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }

func splitClosed(value string) ([]string, error) {
	if value == "" { return nil, fmt.Errorf("closed list is empty") }
	items := strings.Split(value, "|")
	if len(items) > 512 { return nil, fmt.Errorf("closed list exceeds bounded count") }
	for _, item := range items { if item == "" { return nil, fmt.Errorf("closed list contains an empty item") } }
	return items, nil
}

func closedPath(root, relative string) (string, error) {
	clean := path.Clean(filepath.ToSlash(relative))
	if clean == "." || path.IsAbs(clean) || strings.HasPrefix(clean, "../") || strings.Contains(clean, "/../") {
		return "", fmt.Errorf("candidate data path %q escapes its root", relative)
	}
	full := filepath.Join(root, filepath.FromSlash(clean))
	rel, err := filepath.Rel(root, full)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) { return "", fmt.Errorf("candidate data path escapes its root") }
	return full, nil
}

func decode(raw []byte) (manifest, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw)); decoder.DisallowUnknownFields()
	var value manifest
	if err := decoder.Decode(&value); err != nil { return manifest{}, fmt.Errorf("decode strict integration manifest: %w", err) }
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil { return manifest{}, fmt.Errorf("integration manifest contains a second JSON value") }
		return manifest{}, fmt.Errorf("decode integration manifest trailer: %w", err)
	}
	return value, nil
}

func manifestPackages(raw [][]byte) (map[string]struct{}, error) {
	result := make(map[string]struct{})
	for _, body := range raw { value, err := decode(body); if err != nil { return nil, err }; for _, group := range value.Groups { result[group.Package] = struct{}{} } }
	return result, nil
}

func loadSources(root string, packages map[string]struct{}) ([]source, error) {
	paths := make([]string, 0)
	for pkg := range packages {
		if !packagePattern.MatchString(pkg) || pkg == "./..." || strings.Contains(pkg, "//") { return nil, fmt.Errorf("candidate package %q is not one explicit package", pkg) }
		directory, err := closedPath(root, strings.TrimPrefix(pkg, "./"))
		if err != nil { return nil, err }
		entries, err := os.ReadDir(directory)
		if err != nil { return nil, fmt.Errorf("read candidate package %s: %w", pkg, err) }
		for _, entry := range entries {
			if entry.Type()&os.ModeSymlink != 0 { return nil, fmt.Errorf("candidate package %s contains a reparse entry", pkg) }
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), "_test.go") { paths = append(paths, path.Join(strings.TrimPrefix(pkg, "./"), entry.Name())) }
		}
	}
	sort.Strings(paths)
	result := make([]source, 0, len(paths))
	for _, relative := range paths { full, err := closedPath(root, relative); if err != nil { return nil, err }; body, err := os.ReadFile(full); if err != nil { return nil, err }; result = append(result, source{Path: relative, Body: body}) }
	return result, nil
}

func validate(raw [][]byte, sources []source, suite string, suiteTimeout time.Duration) error {
	want, ok := map[string]time.Duration{"batch01": 120*time.Minute, "batch02": 180*time.Minute, "batch03": 240*time.Minute, "final": 240*time.Minute}[suite]
	if !ok { return fmt.Errorf("unknown C12 integration suite %q", suite) }
	if suiteTimeout != want { return fmt.Errorf("C12 %s suite timeout = %s, want exactly %s", suite, suiteTimeout, want) }
	if len(raw) == 0 { return fmt.Errorf("selected C12 integration manifest set is empty") }
	values := make([]manifest, 0, len(raw))
	for index, body := range raw { value, err := decode(body); if err != nil { return fmt.Errorf("manifest %d: %w", index+1, err) }; values = append(values, value) }
	coverage := make(map[testKey]int); packages := make(map[string]struct{}); ids := make(map[string]struct{}); var total time.Duration
	for manifestIndex, value := range values {
		if value.Schema != manifestSchema { return fmt.Errorf("manifest %d schema = %q, want %q", manifestIndex+1, value.Schema, manifestSchema) }
		if len(value.Groups) == 0 { return fmt.Errorf("manifest %d has no integration groups", manifestIndex+1) }
		previousID := ""
		for _, group := range value.Groups {
			if !idPattern.MatchString(group.ID) { return fmt.Errorf("integration group ID %q is not closed and bounded", group.ID) }
			if previousID != "" && group.ID <= previousID { return fmt.Errorf("integration groups are not sorted by unique ID") }; previousID = group.ID
			if _, duplicate := ids[group.ID]; duplicate { return fmt.Errorf("duplicate integration group ID %q", group.ID) }; ids[group.ID] = struct{}{}
			if !packagePattern.MatchString(group.Package) || group.Package == "./..." || strings.Contains(group.Package, "//") { return fmt.Errorf("integration group %s package %q is not one explicit package", group.ID, group.Package) }
			if _, ok := allowed[group.Package]; !ok { return fmt.Errorf("integration group %s package %q is outside the Task 4 allowed package set", group.ID, group.Package) }; packages[group.Package] = struct{}{}
			setup := 3*time.Minute
			switch group.Profile { case "base", "authority-v7": case "authority-v7-pitr": setup = 8*time.Minute; default: return fmt.Errorf("integration group %s has unsupported profile %q", group.ID, group.Profile) }
			if !timeoutPattern.MatchString(group.Timeout) { return fmt.Errorf("integration group %s timeout is not finite", group.ID) }
			duration, err := time.ParseDuration(group.Timeout); if err != nil || duration <= 0 || duration > 30*time.Minute { return fmt.Errorf("integration group %s timeout exceeds 30m or is invalid", group.ID) }; total += duration + setup
			if len(group.Tests) == 0 || len(group.Tests) > 512 { return fmt.Errorf("integration group %s has invalid exact tests", group.ID) }
			previousTest := ""
			for _, test := range group.Tests {
				if !testPattern.MatchString(test) { return fmt.Errorf("integration group %s test %q is not literal", group.ID, test) }
				if test == "TestPrepareC12AuthorityV7Database" { return fmt.Errorf("profile initializer is forbidden") }
				if previousTest != "" && test <= previousTest { return fmt.Errorf("integration group %s tests are duplicate or unsorted", group.ID) }; previousTest = test
				coverage[testKey{group.Package, test}]++
			}
		}
	}
	if total > suiteTimeout { return fmt.Errorf("C12 %s manifest budget exceeds suite budget", suite) }
	universe := make(map[testKey]string)
	for _, source := range sources {
		if !strings.HasSuffix(filepath.ToSlash(source.Path), "_test.go") || !literalTag(source.Body) { continue }
		normalized := path.Clean(filepath.ToSlash(source.Path)); pkg := "./" + path.Dir(normalized)
		if _, selected := packages[pkg]; !selected { continue }
		parsed, err := parser.ParseFile(token.NewFileSet(), normalized, source.Body, parser.SkipObjectResolution); if err != nil { return fmt.Errorf("parse tracked integration source %s: %w", normalized, err) }
		for _, declaration := range parsed.Decls {
			function, ok := declaration.(*ast.FuncDecl); if !ok || function.Recv != nil || !testPattern.MatchString(function.Name.Name) || function.Name.Name == "TestPrepareC12AuthorityV7Database" { continue }
			if callsSkip(function) { return fmt.Errorf("manifest-selected test %s.%s calls testing Skip", pkg, function.Name.Name) }
			key := testKey{pkg, function.Name.Name}; if previous, exists := universe[key]; exists { return fmt.Errorf("duplicate tracked integration test in %s and %s", previous, normalized) }; universe[key] = normalized
		}
	}
	for key, count := range coverage { if count != 1 { return fmt.Errorf("duplicate manifest coverage for %s.%s", key.pkg, key.test) }; if _, ok := universe[key]; !ok { return fmt.Errorf("extra manifest test %s.%s has no tracked first-line integration source", key.pkg, key.test) } }
	for key, source := range universe { if coverage[key] == 0 { return fmt.Errorf("missing manifest coverage for %s.%s from %s", key.pkg, key.test, source) } }
	return nil
}

func resolveFocused(sources []source, packages, requested []string) (map[string][]string, error) {
	selected := make(map[string]struct{}, len(packages)); resolved := make(map[string][]string, len(packages))
	for _, pkg := range packages { if _, ok := allowed[pkg]; !ok { return nil, fmt.Errorf("focused package %q is outside Task 4", pkg) }; if _, duplicate := selected[pkg]; duplicate { return nil, fmt.Errorf("duplicate focused package %q", pkg) }; selected[pkg] = struct{}{}; resolved[pkg] = nil }
	requestedSet := make(map[string]struct{}, len(requested)); for _, test := range requested { if !testPattern.MatchString(test) { return nil, fmt.Errorf("malformed focused test") }; if _, duplicate := requestedSet[test]; duplicate { return nil, fmt.Errorf("duplicate focused test") }; requestedSet[test] = struct{}{} }
	definitions := make(map[string][]string); seen := make(map[string]string)
	for _, source := range sources {
		if !literalTag(source.Body) { continue }; normalized := path.Clean(filepath.ToSlash(source.Path)); pkg := "./" + path.Dir(normalized); if _, ok := selected[pkg]; !ok { continue }
		parsed, err := parser.ParseFile(token.NewFileSet(), normalized, source.Body, parser.SkipObjectResolution); if err != nil { return nil, err }
		for _, declaration := range parsed.Decls { function, ok := declaration.(*ast.FuncDecl); if !ok || function.Recv != nil || !testPattern.MatchString(function.Name.Name) { continue }; key := pkg+"\x00"+function.Name.Name; if previous, duplicate := seen[key]; duplicate { return nil, fmt.Errorf("ambiguous focused test in %s and %s", previous, normalized) }; seen[key] = normalized; definitions[function.Name.Name] = append(definitions[function.Name.Name], pkg) }
	}
	for _, test := range requested { owners := definitions[test]; if len(owners) == 0 { return nil, fmt.Errorf("focused requested test %s is missing", test) }; if len(owners) != 1 { return nil, fmt.Errorf("focused requested test %s is ambiguous", test) }; resolved[owners[0]] = append(resolved[owners[0]], test) }
	for _, pkg := range packages { if len(resolved[pkg]) == 0 { return nil, fmt.Errorf("focused package %s has no requested local test", pkg) }; sort.Strings(resolved[pkg]) }
	return resolved, nil
}

func literalTag(body []byte) bool { line := body; if index := bytes.IndexByte(body, '\n'); index >= 0 { line = body[:index] }; line = bytes.TrimSuffix(line, []byte{'\r'}); return bytes.Equal(line, []byte("//go:build integration")) }
func callsSkip(function *ast.FuncDecl) bool { names := make(map[string]struct{}); if function.Type.Params != nil { for _, field := range function.Type.Params.List { for _, name := range field.Names { names[name.Name] = struct{}{} } } }; found := false; ast.Inspect(function.Body, func(node ast.Node) bool { call, ok := node.(*ast.CallExpr); if !ok { return true }; selector, ok := call.Fun.(*ast.SelectorExpr); if !ok || (selector.Sel.Name != "Skip" && selector.Sel.Name != "Skipf" && selector.Sel.Name != "SkipNow") { return true }; receiver, ok := selector.X.(*ast.Ident); if ok { if _, parameter := names[receiver.Name]; parameter { found = true; return false } }; return true }); return found }
'@
$script:c12TrustedValidatorSourceSHA256 = '76ca48488df32a81ab815a9d3cfddbef3bf3cd01bea3d62591c19d8bdbabd4f7'

function ConvertFrom-C12Duration {
  param(
    [Parameter(Mandatory = $true)]
    [string]$Value
  )

  if ($Value -notmatch '^([1-9][0-9]*)(s|m)$') {
    throw "timeout '$Value' is not a finite positive seconds/minutes duration"
  }
  [long]$quantity = 0
  if (-not [long]::TryParse($Matches[1], [ref]$quantity)) {
    throw "timeout '$Value' is outside the supported range"
  }
  if ($Matches[2] -eq 'm') {
    if ($quantity -gt 1440) {
      throw "timeout '$Value' is outside the supported range"
    }
    return [TimeSpan]::FromMinutes($quantity)
  }
  if ($quantity -gt 86400) {
    throw "timeout '$Value' is outside the supported range"
  }
  return [TimeSpan]::FromSeconds($quantity)
}

function New-C12RandomSuffix {
  $bytes = New-Object byte[] 16
  $random = [System.Security.Cryptography.RandomNumberGenerator]::Create()
  try {
    $random.GetBytes($bytes)
  }
  finally {
    $random.Dispose()
  }
  return (($bytes | ForEach-Object { $_.ToString('x2') }) -join '')
}

function New-C12DatabasePassword {
  param(
    [Parameter(Mandatory = $true)]
    [string]$RunSuffix
  )

  $hash = [System.Security.Cryptography.SHA256]::Create()
  try {
    $bytes = [System.Text.Encoding]::UTF8.GetBytes("TALENRO-C12-POSTGRES-PASSWORD-V1`0$RunSuffix")
    return (($hash.ComputeHash($bytes) | ForEach-Object { $_.ToString('x2') }) -join '')
  }
  finally {
    $hash.Dispose()
  }
}

function Get-C12SHA256Hex {
  param(
    [Parameter(Mandatory = $true)]
    [byte[]]$Bytes
  )

  $hash = [Security.Cryptography.SHA256]::Create()
  try {
    return (($hash.ComputeHash($Bytes) | ForEach-Object { $_.ToString('x2') }) -join '')
  }
  finally {
    $hash.Dispose()
  }
}

function New-C12PITRSecretHex {
  return (New-C12RandomSuffix) + (New-C12RandomSuffix)
}

function ConvertFrom-C12Hex {
  param([Parameter(Mandatory = $true)][string]$Value)

  if ($Value.Length -lt 2 -or $Value.Length % 2 -ne 0 -or $Value -notmatch '^[0-9a-f]+$') {
    throw 'closed hexadecimal value is malformed'
  }
  $bytes = New-Object byte[] ($Value.Length / 2)
  for ($index = 0; $index -lt $bytes.Length; $index++) {
    $bytes[$index] = [Convert]::ToByte($Value.Substring($index * 2, 2), 16)
  }
  return $bytes
}

Add-Type -TypeDefinition @'
using System;
using System.ComponentModel;
using System.Runtime.InteropServices;

public static class C12NativeJob
{
    private const UInt32 JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE = 0x00002000;
    private const Int32 ERROR_ALREADY_EXISTS = 183;

    [StructLayout(LayoutKind.Sequential)]
    private struct JOBOBJECT_BASIC_LIMIT_INFORMATION
    {
        public Int64 PerProcessUserTimeLimit;
        public Int64 PerJobUserTimeLimit;
        public UInt32 LimitFlags;
        public UIntPtr MinimumWorkingSetSize;
        public UIntPtr MaximumWorkingSetSize;
        public UInt32 ActiveProcessLimit;
        public IntPtr Affinity;
        public UInt32 PriorityClass;
        public UInt32 SchedulingClass;
    }

    [StructLayout(LayoutKind.Sequential)]
    private struct IO_COUNTERS
    {
        public UInt64 ReadOperationCount;
        public UInt64 WriteOperationCount;
        public UInt64 OtherOperationCount;
        public UInt64 ReadTransferCount;
        public UInt64 WriteTransferCount;
        public UInt64 OtherTransferCount;
    }

    [StructLayout(LayoutKind.Sequential)]
    private struct JOBOBJECT_EXTENDED_LIMIT_INFORMATION
    {
        public JOBOBJECT_BASIC_LIMIT_INFORMATION BasicLimitInformation;
        public IO_COUNTERS IoInfo;
        public UIntPtr ProcessMemoryLimit;
        public UIntPtr JobMemoryLimit;
        public UIntPtr PeakProcessMemoryUsed;
        public UIntPtr PeakJobMemoryUsed;
    }

    [DllImport("kernel32.dll", CharSet = CharSet.Unicode, SetLastError = true)]
    private static extern IntPtr CreateJobObject(IntPtr attributes, string name);

    [DllImport("kernel32.dll", SetLastError = true)]
    private static extern bool SetInformationJobObject(IntPtr job, int informationClass, IntPtr information, UInt32 length);

    [DllImport("kernel32.dll", SetLastError = true)]
    private static extern bool CloseHandle(IntPtr handle);

    public static IntPtr CreateKillOnClose(string name)
    {
        if (String.IsNullOrEmpty(name)) throw new ArgumentException("named native Job identity is empty", "name");
        IntPtr job = CreateJobObject(IntPtr.Zero, name);
        int createError = Marshal.GetLastWin32Error();
        if (job == IntPtr.Zero) throw new Win32Exception(Marshal.GetLastWin32Error());
        if (createError == ERROR_ALREADY_EXISTS)
        {
            CloseHandle(job);
            throw new InvalidOperationException("named native Job collision");
        }
        var information = new JOBOBJECT_EXTENDED_LIMIT_INFORMATION();
        information.BasicLimitInformation.LimitFlags = JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE;
        int length = Marshal.SizeOf(typeof(JOBOBJECT_EXTENDED_LIMIT_INFORMATION));
        IntPtr pointer = Marshal.AllocHGlobal(length);
        try
        {
            Marshal.StructureToPtr(information, pointer, false);
            if (!SetInformationJobObject(job, 9, pointer, (UInt32)length))
            {
                int error = Marshal.GetLastWin32Error();
                CloseHandle(job);
                throw new Win32Exception(error);
            }
        }
        finally { Marshal.FreeHGlobal(pointer); }
        return job;
    }

    public static void Close(IntPtr job)
    {
        if (job != IntPtr.Zero && !CloseHandle(job))
            throw new Win32Exception(Marshal.GetLastWin32Error());
    }
}

public sealed class C12OwnedDirectory : IDisposable
{
    private const UInt32 FILE_READ_ATTRIBUTES = 0x00000080;
    private const UInt32 FILE_LIST_DIRECTORY = 0x00000001;
    private const UInt32 DELETE = 0x00010000;
    private const UInt32 SYNCHRONIZE = 0x00100000;
    private const UInt32 FILE_SHARE_READ = 0x00000001;
    private const UInt32 FILE_SHARE_WRITE = 0x00000002;
    private const UInt32 FILE_SHARE_DELETE = 0x00000004;
    private const UInt32 OPEN_EXISTING = 3;
    private const UInt32 FILE_CREATE = 2;
    private const UInt32 FILE_OPEN = 1;
    private const UInt32 FILE_DIRECTORY_FILE = 0x00000001;
    private const UInt32 FILE_SYNCHRONOUS_IO_NONALERT = 0x00000020;
    private const UInt32 FILE_FLAG_OPEN_REPARSE_POINT = 0x00200000;
    private const UInt32 FILE_FLAG_BACKUP_SEMANTICS = 0x02000000;
    private const UInt32 FILE_OPEN_REPARSE_POINT = 0x00200000;
    private const UInt32 FILE_ATTRIBUTE_READONLY = 0x00000001;
    private const UInt32 FILE_ATTRIBUTE_DIRECTORY = 0x00000010;
    private const UInt32 FILE_ATTRIBUTE_NORMAL = 0x00000080;
    private const UInt32 FILE_ATTRIBUTE_REPARSE_POINT = 0x00000400;
    private const UInt32 OBJ_CASE_INSENSITIVE = 0x00000040;
    private const UInt64 FILE_CREATED = 2;
    private const Int32 FILE_ID_BOTH_DIRECTORY_INFORMATION_CLASS = 37;
    private const UInt32 STATUS_NO_MORE_FILES = 0x80000006;
    private const Int32 FILE_DISPOSITION_INFO_CLASS = 4;
    private const Int32 FILE_DISPOSITION_INFO_EX_CLASS = 21;
    private const UInt32 FILE_DISPOSITION_FLAG_DELETE = 0x00000001;
    private const UInt32 FILE_DISPOSITION_FLAG_POSIX_SEMANTICS = 0x00000002;
    private const UInt32 FILE_DISPOSITION_FLAG_IGNORE_READONLY_ATTRIBUTE = 0x00000010;
    private static readonly IntPtr INVALID_HANDLE_VALUE = new IntPtr(-1);

    [StructLayout(LayoutKind.Sequential)]
    private struct FILETIME
    {
        public UInt32 LowDateTime;
        public UInt32 HighDateTime;
    }

    [StructLayout(LayoutKind.Sequential)]
    private struct BY_HANDLE_FILE_INFORMATION
    {
        public UInt32 FileAttributes;
        public FILETIME CreationTime;
        public FILETIME LastAccessTime;
        public FILETIME LastWriteTime;
        public UInt32 VolumeSerialNumber;
        public UInt32 FileSizeHigh;
        public UInt32 FileSizeLow;
        public UInt32 NumberOfLinks;
        public UInt32 FileIndexHigh;
        public UInt32 FileIndexLow;
    }

    [StructLayout(LayoutKind.Sequential)]
    private struct FILE_DISPOSITION_INFO
    {
        [MarshalAs(UnmanagedType.Bool)]
        public bool DeleteFile;
    }

    [StructLayout(LayoutKind.Sequential)]
    private struct FILE_DISPOSITION_INFO_EX
    {
        public UInt32 Flags;
    }

    [StructLayout(LayoutKind.Sequential)]
    private struct UNICODE_STRING
    {
        public UInt16 Length;
        public UInt16 MaximumLength;
        public IntPtr Buffer;
    }

    [StructLayout(LayoutKind.Sequential)]
    private struct OBJECT_ATTRIBUTES
    {
        public UInt32 Length;
        public IntPtr RootDirectory;
        public IntPtr ObjectName;
        public UInt32 Attributes;
        public IntPtr SecurityDescriptor;
        public IntPtr SecurityQualityOfService;
    }

    [StructLayout(LayoutKind.Sequential)]
    private struct IO_STATUS_BLOCK
    {
        public IntPtr Status;
        public UIntPtr Information;
    }

    [DllImport("ntdll.dll")]
    private static extern Int32 NtCreateFile(out IntPtr fileHandle, UInt32 desiredAccess, ref OBJECT_ATTRIBUTES objectAttributes, out IO_STATUS_BLOCK ioStatusBlock, IntPtr allocationSize, UInt32 fileAttributes, UInt32 shareAccess, UInt32 createDisposition, UInt32 createOptions, IntPtr eaBuffer, UInt32 eaLength);

    [DllImport("ntdll.dll")]
    private static extern UInt32 RtlNtStatusToDosError(Int32 status);

    [DllImport("ntdll.dll")]
    private static extern Int32 NtQueryDirectoryFile(IntPtr fileHandle, IntPtr eventHandle, IntPtr apcRoutine, IntPtr apcContext, out IO_STATUS_BLOCK ioStatusBlock, IntPtr fileInformation, UInt32 length, Int32 fileInformationClass, [MarshalAs(UnmanagedType.Bool)] bool returnSingleEntry, IntPtr fileName, [MarshalAs(UnmanagedType.Bool)] bool restartScan);

    [DllImport("kernel32.dll", CharSet = CharSet.Unicode, SetLastError = true)]
    private static extern IntPtr CreateFile(string path, UInt32 desiredAccess, UInt32 shareMode, IntPtr securityAttributes, UInt32 creationDisposition, UInt32 flagsAndAttributes, IntPtr templateFile);

    [DllImport("kernel32.dll", SetLastError = true)]
    private static extern bool GetFileInformationByHandle(IntPtr handle, out BY_HANDLE_FILE_INFORMATION information);

    [DllImport("kernel32.dll", SetLastError = true)]
    private static extern bool SetFileInformationByHandle(IntPtr handle, Int32 informationClass, IntPtr information, UInt32 bufferSize);

    [DllImport("kernel32.dll", CharSet = CharSet.Unicode, SetLastError = true)]
    private static extern bool SetFileAttributes(string path, UInt32 attributes);

    [DllImport("kernel32.dll", SetLastError = true)]
    private static extern bool CloseHandle(IntPtr handle);

    private IntPtr handle;
    private readonly UInt32 volumeSerial;
    private readonly UInt64 fileIndex;
    public string RootPath { get; private set; }
    public UInt32 VolumeSerial { get { return volumeSerial; } }
    public UInt64 FileIndex { get { return fileIndex; } }
    public string Identity { get { return volumeSerial.ToString("x8") + ":" + fileIndex.ToString("x16"); } }
    public static Action<string> RelativeOpenObserver { get; set; }

    private sealed class RelativeEntry
    {
        public string Name;
        public UInt32 Attributes;
        public UInt64 FileId;
    }

    private C12OwnedDirectory(string rootPath, IntPtr ownedHandle, BY_HANDLE_FILE_INFORMATION information)
    {
        RootPath = rootPath;
        handle = ownedHandle;
        volumeSerial = information.VolumeSerialNumber;
        fileIndex = ((UInt64)information.FileIndexHigh << 32) | information.FileIndexLow;
    }

    public static C12OwnedDirectory CreateNew(string exactRoot)
    {
        string root = System.IO.Path.GetFullPath(exactRoot).TrimEnd('\\');
        if (root.Length < 3 || root[1] != ':' || root[2] != '\\')
            throw new InvalidOperationException("owned directory requires one local drive-rooted path");
        string nativePath = @"\??\" + root;
        int pathByteLength = System.Text.Encoding.Unicode.GetByteCount(nativePath);
        if (pathByteLength < 2 || pathByteLength > UInt16.MaxValue - 2)
            throw new InvalidOperationException("owned directory native path is outside the bounded Unicode range");

        IntPtr pathBuffer = IntPtr.Zero;
        IntPtr namePointer = IntPtr.Zero;
        IntPtr createdHandle = INVALID_HANDLE_VALUE;
        try
        {
            pathBuffer = Marshal.StringToHGlobalUni(nativePath);
            UNICODE_STRING name = new UNICODE_STRING();
            name.Length = (UInt16)pathByteLength;
            name.MaximumLength = (UInt16)(pathByteLength + 2);
            name.Buffer = pathBuffer;
            namePointer = Marshal.AllocHGlobal(Marshal.SizeOf(typeof(UNICODE_STRING)));
            Marshal.StructureToPtr(name, namePointer, false);
            OBJECT_ATTRIBUTES attributes = new OBJECT_ATTRIBUTES();
            attributes.Length = (UInt32)Marshal.SizeOf(typeof(OBJECT_ATTRIBUTES));
            attributes.RootDirectory = IntPtr.Zero;
            attributes.ObjectName = namePointer;
            attributes.Attributes = OBJ_CASE_INSENSITIVE;
            IO_STATUS_BLOCK ioStatus;
            Int32 status = NtCreateFile(
                out createdHandle,
                FILE_LIST_DIRECTORY | FILE_READ_ATTRIBUTES | DELETE | SYNCHRONIZE,
                ref attributes,
                out ioStatus,
                IntPtr.Zero,
                FILE_ATTRIBUTE_NORMAL,
                FILE_SHARE_READ | FILE_SHARE_WRITE | FILE_SHARE_DELETE,
                FILE_CREATE,
                FILE_DIRECTORY_FILE | FILE_SYNCHRONOUS_IO_NONALERT | FILE_OPEN_REPARSE_POINT,
                IntPtr.Zero,
                0);
            if (status < 0)
            {
                if ((UInt32)status == 0xC0000035U) throw new InvalidOperationException("owned directory already exists");
                throw new Win32Exception((Int32)RtlNtStatusToDosError(status));
            }
            if (createdHandle == IntPtr.Zero || createdHandle == INVALID_HANDLE_VALUE || ioStatus.Information.ToUInt64() != FILE_CREATED)
                throw new InvalidOperationException("atomic owned-directory creation did not return one newly created handle");
            BY_HANDLE_FILE_INFORMATION information = ReadInformation(createdHandle);
            if ((information.FileAttributes & FILE_ATTRIBUTE_DIRECTORY) == 0 || (information.FileAttributes & FILE_ATTRIBUTE_REPARSE_POINT) != 0)
                throw new InvalidOperationException("owned directory root is not one non-reparse directory object");
            C12OwnedDirectory result = new C12OwnedDirectory(root, createdHandle, information);
            createdHandle = INVALID_HANDLE_VALUE;
            return result;
        }
        catch
        {
            if (createdHandle != IntPtr.Zero && createdHandle != INVALID_HANDLE_VALUE)
            {
                try { MarkDelete(createdHandle); } catch { }
            }
            throw;
        }
        finally
        {
            if (createdHandle != IntPtr.Zero && createdHandle != INVALID_HANDLE_VALUE) CloseHandle(createdHandle);
            if (namePointer != IntPtr.Zero) Marshal.FreeHGlobal(namePointer);
            if (pathBuffer != IntPtr.Zero) Marshal.FreeHGlobal(pathBuffer);
        }
    }

    public static C12OwnedDirectory OpenExisting(string exactRoot, UInt32 expectedVolumeSerial, UInt64 expectedFileIndex)
    {
        string root = System.IO.Path.GetFullPath(exactRoot).TrimEnd('\\');
        IntPtr existing = CreateFile(root, FILE_LIST_DIRECTORY | FILE_READ_ATTRIBUTES | DELETE | SYNCHRONIZE, FILE_SHARE_READ | FILE_SHARE_WRITE | FILE_SHARE_DELETE, IntPtr.Zero, OPEN_EXISTING, FILE_FLAG_BACKUP_SEMANTICS | FILE_FLAG_OPEN_REPARSE_POINT, IntPtr.Zero);
        if (existing == INVALID_HANDLE_VALUE) throw new Win32Exception(Marshal.GetLastWin32Error());
        try
        {
            BY_HANDLE_FILE_INFORMATION information = ReadInformation(existing);
            UInt64 index = ((UInt64)information.FileIndexHigh << 32) | information.FileIndexLow;
            if ((information.FileAttributes & FILE_ATTRIBUTE_DIRECTORY) == 0 || (information.FileAttributes & FILE_ATTRIBUTE_REPARSE_POINT) != 0 || information.VolumeSerialNumber != expectedVolumeSerial || index != expectedFileIndex)
                throw new InvalidOperationException("existing owned-directory identity changed before retry");
            C12OwnedDirectory result = new C12OwnedDirectory(root, existing, information);
            existing = INVALID_HANDLE_VALUE;
            return result;
        }
        finally { if (existing != IntPtr.Zero && existing != INVALID_HANDLE_VALUE) CloseHandle(existing); }
    }

    private static BY_HANDLE_FILE_INFORMATION ReadInformation(IntPtr value)
    {
        BY_HANDLE_FILE_INFORMATION information;
        if (!GetFileInformationByHandle(value, out information)) throw new Win32Exception(Marshal.GetLastWin32Error());
        return information;
    }

    private void EnsureOpen()
    {
        if (handle == IntPtr.Zero || handle == INVALID_HANDLE_VALUE) throw new ObjectDisposedException("C12OwnedDirectory");
    }

    public void VerifyExactPath()
    {
        EnsureOpen();
        BY_HANDLE_FILE_INFORMATION retained = ReadInformation(handle);
        UInt64 retainedIndex = ((UInt64)retained.FileIndexHigh << 32) | retained.FileIndexLow;
        if ((retained.FileAttributes & FILE_ATTRIBUTE_DIRECTORY) == 0 || (retained.FileAttributes & FILE_ATTRIBUTE_REPARSE_POINT) != 0 || retained.VolumeSerialNumber != volumeSerial || retainedIndex != fileIndex)
            throw new InvalidOperationException("retained owned-directory identity changed");
        IntPtr current = CreateFile(RootPath, FILE_READ_ATTRIBUTES, FILE_SHARE_READ | FILE_SHARE_WRITE | FILE_SHARE_DELETE, IntPtr.Zero, OPEN_EXISTING, FILE_FLAG_BACKUP_SEMANTICS | FILE_FLAG_OPEN_REPARSE_POINT, IntPtr.Zero);
        if (current == INVALID_HANDLE_VALUE) throw new InvalidOperationException("owned directory path no longer resolves to its retained identity", new Win32Exception(Marshal.GetLastWin32Error()));
        try
        {
            BY_HANDLE_FILE_INFORMATION observed = ReadInformation(current);
            UInt64 observedIndex = ((UInt64)observed.FileIndexHigh << 32) | observed.FileIndexLow;
            if ((observed.FileAttributes & FILE_ATTRIBUTE_DIRECTORY) == 0 || (observed.FileAttributes & FILE_ATTRIBUTE_REPARSE_POINT) != 0 || observed.VolumeSerialNumber != volumeSerial || observedIndex != fileIndex)
                throw new InvalidOperationException("owned directory path was substituted");
        }
        finally { CloseHandle(current); }
    }

    public void RequestDeleteExactTree(DateTime deadlineUtc)
    {
        VerifyExactPath();
        RequestDeleteRetainedTree(deadlineUtc);
    }

    public void RequestDeleteRetainedTree(DateTime deadlineUtc)
    {
        EnsureOpen();
        BY_HANDLE_FILE_INFORMATION retained = ReadInformation(handle);
        UInt64 retainedIndex = ((UInt64)retained.FileIndexHigh << 32) | retained.FileIndexLow;
        if ((retained.FileAttributes & FILE_ATTRIBUTE_DIRECTORY) == 0 || (retained.FileAttributes & FILE_ATTRIBUTE_REPARSE_POINT) != 0 || retained.VolumeSerialNumber != volumeSerial || retainedIndex != fileIndex)
            throw new InvalidOperationException("retained owned-directory identity changed before handle-relative cleanup");
        int visited = 0;
        DeleteChildren(handle, deadlineUtc, ref visited);
        MarkDelete(handle);
    }

    public void RequestDeleteExactEmpty(DateTime deadlineUtc)
    {
        VerifyExactPath();
        CheckDeadline(deadlineUtc);
        RelativeEntry entry;
        if (TryReadFirstEntry(handle, out entry)) throw new InvalidOperationException("identity-bound empty-directory deletion found a direct child");
        MarkDelete(handle);
    }

    public void ReleaseDeletePending() { Dispose(); }

    public void DeleteExactTree(DateTime deadlineUtc)
    {
        RequestDeleteExactTree(deadlineUtc);
        ReleaseDeletePending();
        if (System.IO.Directory.Exists(RootPath)) throw new InvalidOperationException("identity-bound root deletion remained pending");
    }

    public void DeleteExactEmpty(DateTime deadlineUtc)
    {
        RequestDeleteExactEmpty(deadlineUtc);
        ReleaseDeletePending();
        if (System.IO.Directory.Exists(RootPath)) throw new InvalidOperationException("identity-bound empty-directory deletion remained pending");
    }

    private static void DeleteChildren(IntPtr directoryHandle, DateTime deadlineUtc, ref int visited)
    {
        RelativeEntry entry;
        while (TryReadFirstEntry(directoryHandle, out entry))
        {
            CheckDeadline(deadlineUtc);
            visited++;
            if (visited > 100000) throw new InvalidOperationException("identity-bound cleanup exceeds the bounded entry count");
            Action<string> observer = RelativeOpenObserver;
            if (observer != null) observer(entry.Name);
            IntPtr childHandle = OpenRelative(directoryHandle, entry.Name);
            try
            {
                BY_HANDLE_FILE_INFORMATION information = ReadInformation(childHandle);
                UInt64 childIndex = ((UInt64)information.FileIndexHigh << 32) | information.FileIndexLow;
                bool reparse = (information.FileAttributes & FILE_ATTRIBUTE_REPARSE_POINT) != 0;
                bool directoryEntry = (information.FileAttributes & FILE_ATTRIBUTE_DIRECTORY) != 0;
                bool enumeratedReparse = (entry.Attributes & FILE_ATTRIBUTE_REPARSE_POINT) != 0;
                bool enumeratedDirectory = (entry.Attributes & FILE_ATTRIBUTE_DIRECTORY) != 0;
                if (childIndex != entry.FileId || reparse != enumeratedReparse || directoryEntry != enumeratedDirectory)
                    throw new InvalidOperationException("identity-bound descendant was substituted between enumeration and relative open");
                if (!reparse && directoryEntry) DeleteChildren(childHandle, deadlineUtc, ref visited);
                MarkDelete(childHandle);
            }
            finally { CloseHandle(childHandle); }
        }
        CheckDeadline(deadlineUtc);
    }

    private static bool TryReadFirstEntry(IntPtr directoryHandle, out RelativeEntry entry)
    {
        int length = 65536;
        IntPtr buffer = Marshal.AllocHGlobal(length);
        try
        {
            bool restartScan = true;
            while (true)
            {
                IO_STATUS_BLOCK ioStatus;
                Int32 status = NtQueryDirectoryFile(directoryHandle, IntPtr.Zero, IntPtr.Zero, IntPtr.Zero, out ioStatus, buffer, (UInt32)length, FILE_ID_BOTH_DIRECTORY_INFORMATION_CLASS, true, IntPtr.Zero, restartScan);
                if ((UInt32)status == STATUS_NO_MORE_FILES) { entry = null; return false; }
                if (status < 0) throw new Win32Exception((Int32)RtlNtStatusToDosError(status));
                restartScan = false;
                UInt32 nameLength = (UInt32)Marshal.ReadInt32(buffer, 60);
                string name = Marshal.PtrToStringUni(IntPtr.Add(buffer, 104), (Int32)(nameLength / 2));
                if (name == "." || name == "..") continue;
                entry = new RelativeEntry();
                entry.Name = name;
                entry.Attributes = (UInt32)Marshal.ReadInt32(buffer, 56);
                entry.FileId = (UInt64)Marshal.ReadInt64(buffer, 96);
                return true;
            }
        }
        finally { Marshal.FreeHGlobal(buffer); }
    }

    private static IntPtr OpenRelative(IntPtr parentHandle, string name)
    {
        IntPtr nameBuffer = IntPtr.Zero;
        IntPtr namePointer = IntPtr.Zero;
        IntPtr child = INVALID_HANDLE_VALUE;
        try
        {
            nameBuffer = Marshal.StringToHGlobalUni(name);
            UNICODE_STRING unicode = new UNICODE_STRING();
            unicode.Length = (UInt16)System.Text.Encoding.Unicode.GetByteCount(name);
            unicode.MaximumLength = (UInt16)(unicode.Length + 2);
            unicode.Buffer = nameBuffer;
            namePointer = Marshal.AllocHGlobal(Marshal.SizeOf(typeof(UNICODE_STRING)));
            Marshal.StructureToPtr(unicode, namePointer, false);
            OBJECT_ATTRIBUTES attributes = new OBJECT_ATTRIBUTES();
            attributes.Length = (UInt32)Marshal.SizeOf(typeof(OBJECT_ATTRIBUTES));
            attributes.RootDirectory = parentHandle;
            attributes.ObjectName = namePointer;
            attributes.Attributes = OBJ_CASE_INSENSITIVE;
            IO_STATUS_BLOCK ioStatus;
            Int32 status = NtCreateFile(out child, FILE_LIST_DIRECTORY | FILE_READ_ATTRIBUTES | DELETE | SYNCHRONIZE, ref attributes, out ioStatus, IntPtr.Zero, FILE_ATTRIBUTE_NORMAL, FILE_SHARE_READ | FILE_SHARE_WRITE | FILE_SHARE_DELETE, FILE_OPEN, FILE_SYNCHRONOUS_IO_NONALERT | FILE_OPEN_REPARSE_POINT, IntPtr.Zero, 0);
            if (status < 0) throw new Win32Exception((Int32)RtlNtStatusToDosError(status));
            IntPtr result = child;
            child = INVALID_HANDLE_VALUE;
            return result;
        }
        finally
        {
            if (child != IntPtr.Zero && child != INVALID_HANDLE_VALUE) CloseHandle(child);
            if (namePointer != IntPtr.Zero) Marshal.FreeHGlobal(namePointer);
            if (nameBuffer != IntPtr.Zero) Marshal.FreeHGlobal(nameBuffer);
        }
    }

    private static void MarkDelete(IntPtr value)
    {
        FILE_DISPOSITION_INFO_EX disposition = new FILE_DISPOSITION_INFO_EX();
        disposition.Flags = FILE_DISPOSITION_FLAG_DELETE | FILE_DISPOSITION_FLAG_POSIX_SEMANTICS | FILE_DISPOSITION_FLAG_IGNORE_READONLY_ATTRIBUTE;
        int length = Marshal.SizeOf(typeof(FILE_DISPOSITION_INFO_EX));
        IntPtr pointer = Marshal.AllocHGlobal(length);
        try
        {
            Marshal.StructureToPtr(disposition, pointer, false);
            if (!SetFileInformationByHandle(value, FILE_DISPOSITION_INFO_EX_CLASS, pointer, (UInt32)length))
            {
                FILE_DISPOSITION_INFO legacy = new FILE_DISPOSITION_INFO();
                legacy.DeleteFile = true;
                Marshal.StructureToPtr(legacy, pointer, false);
                if (!SetFileInformationByHandle(value, FILE_DISPOSITION_INFO_CLASS, pointer, (UInt32)Marshal.SizeOf(typeof(FILE_DISPOSITION_INFO))))
                    throw new Win32Exception(Marshal.GetLastWin32Error());
            }
        }
        finally { Marshal.FreeHGlobal(pointer); }
    }

    private static void CheckDeadline(DateTime deadlineUtc)
    {
        if (deadlineUtc != DateTime.MaxValue && DateTime.UtcNow >= deadlineUtc)
            throw new TimeoutException("identity-bound cleanup exceeded its absolute deadline");
    }

    public void Dispose()
    {
        IntPtr current = handle;
        handle = IntPtr.Zero;
        if (current != IntPtr.Zero && current != INVALID_HANDLE_VALUE && !CloseHandle(current))
            throw new Win32Exception(Marshal.GetLastWin32Error());
        GC.SuppressFinalize(this);
    }

    ~C12OwnedDirectory()
    {
        IntPtr current = handle;
        handle = IntPtr.Zero;
        if (current != IntPtr.Zero && current != INVALID_HANDLE_VALUE) CloseHandle(current);
    }
}
'@

$script:c12SealedExecutableSource = @'
using System;
using System.ComponentModel;
using System.IO;
using System.Runtime.InteropServices;
using System.Security.AccessControl;
using System.Security.Cryptography;
using System.Security.Principal;
using Microsoft.Win32.SafeHandles;

public sealed class C12PathIdentity
{
    public UInt32 VolumeSerialNumber { get; private set; }
    public UInt64 FileIndex { get; private set; }
    public UInt32 NumberOfLinks { get; private set; }
    public bool Reparse { get; private set; }
    public bool Directory { get; private set; }

    internal C12PathIdentity(UInt32 volumeSerialNumber, UInt64 fileIndex, UInt32 numberOfLinks, bool reparse, bool directory)
    {
        VolumeSerialNumber = volumeSerialNumber;
        FileIndex = fileIndex;
        NumberOfLinks = numberOfLinks;
        Reparse = reparse;
        Directory = directory;
    }

    public string Value { get { return VolumeSerialNumber.ToString("x8") + ":" + FileIndex.ToString("x16"); } }
}

public sealed class C12SealedExecutable : IDisposable
{
    private const UInt32 GENERIC_READ = 0x80000000;
    private const UInt32 FILE_READ_ATTRIBUTES = 0x00000080;
    private const UInt32 DELETE = 0x00010000;
    private const UInt32 FILE_SHARE_READ = 0x00000001;
    private const UInt32 FILE_SHARE_WRITE = 0x00000002;
    private const UInt32 FILE_SHARE_DELETE = 0x00000004;
    private const UInt32 OPEN_EXISTING = 3;
    private const UInt32 FILE_FLAG_OPEN_REPARSE_POINT = 0x00200000;
    private const UInt32 FILE_FLAG_BACKUP_SEMANTICS = 0x02000000;
    private const UInt32 FILE_ATTRIBUTE_DIRECTORY = 0x00000010;
    private const UInt32 FILE_ATTRIBUTE_REPARSE_POINT = 0x00000400;
    private const Int32 FILE_DISPOSITION_INFO_CLASS = 4;
    private const Int32 FILE_DISPOSITION_INFO_EX_CLASS = 21;
    private const UInt32 FILE_DISPOSITION_FLAG_DELETE = 0x00000001;
    private const UInt32 FILE_DISPOSITION_FLAG_POSIX_SEMANTICS = 0x00000002;
    private const UInt32 FILE_DISPOSITION_FLAG_IGNORE_READONLY_ATTRIBUTE = 0x00000010;
    private const UInt32 SE_FILE_OBJECT = 1;
    private const UInt32 OWNER_SECURITY_INFORMATION = 0x00000001;
    private const UInt32 DACL_SECURITY_INFORMATION = 0x00000004;
    private const UInt32 SDDL_REVISION_1 = 1;
    private static readonly IntPtr INVALID_HANDLE_VALUE = new IntPtr(-1);

    [StructLayout(LayoutKind.Sequential)]
    private struct FILETIME { public UInt32 LowDateTime; public UInt32 HighDateTime; }

    [StructLayout(LayoutKind.Sequential)]
    private struct BY_HANDLE_FILE_INFORMATION
    {
        public UInt32 FileAttributes;
        public FILETIME CreationTime;
        public FILETIME LastAccessTime;
        public FILETIME LastWriteTime;
        public UInt32 VolumeSerialNumber;
        public UInt32 FileSizeHigh;
        public UInt32 FileSizeLow;
        public UInt32 NumberOfLinks;
        public UInt32 FileIndexHigh;
        public UInt32 FileIndexLow;
    }

    [StructLayout(LayoutKind.Sequential)]
    private struct FILE_DISPOSITION_INFO { [MarshalAs(UnmanagedType.Bool)] public bool DeleteFile; }
    [StructLayout(LayoutKind.Sequential)]
    private struct FILE_DISPOSITION_INFO_EX { public UInt32 Flags; }

    [DllImport("kernel32.dll", CharSet = CharSet.Unicode, SetLastError = true)]
    private static extern IntPtr CreateFile(string path, UInt32 desiredAccess, UInt32 shareMode, IntPtr securityAttributes, UInt32 creationDisposition, UInt32 flagsAndAttributes, IntPtr templateFile);

    [DllImport("kernel32.dll", SetLastError = true)]
    private static extern bool GetFileInformationByHandle(IntPtr handle, out BY_HANDLE_FILE_INFORMATION information);
    [DllImport("kernel32.dll", SetLastError = true)]
    private static extern bool SetFileInformationByHandle(IntPtr handle, Int32 informationClass, IntPtr information, UInt32 bufferSize);

    [DllImport("kernel32.dll", SetLastError = true)]
    private static extern bool CloseHandle(IntPtr handle);
    [DllImport("kernel32.dll", SetLastError = true)]
    private static extern IntPtr LocalFree(IntPtr memory);
    [DllImport("advapi32.dll", SetLastError = true)]
    private static extern UInt32 GetSecurityInfo(IntPtr handle, UInt32 objectType, UInt32 securityInformation, out IntPtr ownerSid, out IntPtr groupSid, out IntPtr dacl, out IntPtr sacl, out IntPtr securityDescriptor);
    [DllImport("advapi32.dll", CharSet = CharSet.Unicode, SetLastError = true)]
    private static extern bool ConvertSecurityDescriptorToStringSecurityDescriptor(IntPtr securityDescriptor, UInt32 revision, UInt32 securityInformation, out IntPtr stringSecurityDescriptor, out UInt32 stringSecurityDescriptorLength);

    private IntPtr handle;
    public string ExactPath { get; private set; }
    public string SHA256 { get; private set; }
    public UInt64 Length { get; private set; }
    public UInt32 VolumeSerialNumber { get; private set; }
    public UInt64 FileIndex { get; private set; }
    public UInt32 NumberOfLinks { get; private set; }
    public string Owner { get; private set; }
    public string DACL { get; private set; }
    public bool Reparse { get; private set; }

    private C12SealedExecutable(string exactPath, bool denyWriteDelete, bool permitDelete, bool allowEmpty)
    {
        ExactPath = Path.GetFullPath(exactPath);
        UInt32 sharing = denyWriteDelete ? (UInt32)FileShare.Read : (UInt32)(FileShare.Read | FileShare.Write | FileShare.Delete);
        UInt32 desiredAccess = GENERIC_READ | FILE_READ_ATTRIBUTES | (permitDelete ? DELETE : 0);
        handle = CreateFile(ExactPath, desiredAccess, sharing, IntPtr.Zero, OPEN_EXISTING, FILE_FLAG_OPEN_REPARSE_POINT, IntPtr.Zero);
        if (handle == INVALID_HANDLE_VALUE) throw new Win32Exception(Marshal.GetLastWin32Error());
        try
        {
            BY_HANDLE_FILE_INFORMATION information = ReadInformation(handle);
            Reparse = (information.FileAttributes & FILE_ATTRIBUTE_REPARSE_POINT) != 0;
            if ((information.FileAttributes & FILE_ATTRIBUTE_DIRECTORY) != 0 || Reparse)
                throw new InvalidOperationException("sealed executable is not one non-reparse regular file");
            VolumeSerialNumber = information.VolumeSerialNumber;
            FileIndex = ((UInt64)information.FileIndexHigh << 32) | information.FileIndexLow;
            NumberOfLinks = information.NumberOfLinks;
            Length = ((UInt64)information.FileSizeHigh << 32) | information.FileSizeLow;
            if ((!allowEmpty && Length == 0) || NumberOfLinks != 1) throw new InvalidOperationException("sealed executable has invalid length or hard-link count");
            SHA256 = ComputeSHA256(handle);
            string owner;
            string dacl;
            ReadSecurityReceipt(handle, out owner, out dacl);
            Owner = owner;
            DACL = dacl;
        }
        catch
        {
            Dispose();
            throw;
        }
    }

    private static BY_HANDLE_FILE_INFORMATION ReadInformation(IntPtr value)
    {
        BY_HANDLE_FILE_INFORMATION information;
        if (!GetFileInformationByHandle(value, out information)) throw new Win32Exception(Marshal.GetLastWin32Error());
        return information;
    }

    private static string ComputeSHA256(IntPtr value)
    {
        using (SafeFileHandle safe = new SafeFileHandle(value, false))
        using (FileStream stream = new FileStream(safe, FileAccess.Read, 4096, false))
        using (SHA256 algorithm = System.Security.Cryptography.SHA256.Create())
        {
            stream.Position = 0;
            byte[] digest = algorithm.ComputeHash(stream);
            return BitConverter.ToString(digest).Replace("-", "").ToLowerInvariant();
        }
    }

    private static void ReadSecurityReceipt(IntPtr value, out string owner, out string dacl)
    {
        IntPtr ownerSid;
        IntPtr groupSid;
        IntPtr daclPointer;
        IntPtr sacl;
        IntPtr descriptor;
        UInt32 result = GetSecurityInfo(value, SE_FILE_OBJECT, OWNER_SECURITY_INFORMATION | DACL_SECURITY_INFORMATION,
            out ownerSid, out groupSid, out daclPointer, out sacl, out descriptor);
        if (result != 0) throw new Win32Exception((Int32)result);
        IntPtr sddl = IntPtr.Zero;
        try
        {
            if (ownerSid == IntPtr.Zero) throw new InvalidOperationException("retained-handle security descriptor has no owner");
            owner = new SecurityIdentifier(ownerSid).Value;
            UInt32 length;
            if (!ConvertSecurityDescriptorToStringSecurityDescriptor(descriptor, SDDL_REVISION_1,
                OWNER_SECURITY_INFORMATION | DACL_SECURITY_INFORMATION, out sddl, out length))
                throw new Win32Exception(Marshal.GetLastWin32Error());
            dacl = Marshal.PtrToStringUni(sddl);
            if (String.IsNullOrEmpty(dacl)) throw new InvalidOperationException("retained-handle security descriptor is empty");
        }
        finally
        {
            if (sddl != IntPtr.Zero) LocalFree(sddl);
            if (descriptor != IntPtr.Zero) LocalFree(descriptor);
        }
    }

    public static C12SealedExecutable Inspect(string exactPath) { return new C12SealedExecutable(exactPath, false, false, false); }

    public static C12SealedExecutable InspectLeaf(string exactPath) { return new C12SealedExecutable(exactPath, false, false, true); }

    public static C12SealedExecutable OpenForCleanup(string exactPath) { return new C12SealedExecutable(exactPath, false, true, false); }

    public static C12SealedExecutable OpenLeafForCleanup(string exactPath) { return new C12SealedExecutable(exactPath, false, true, true); }

    public static C12SealedExecutable OpenAndVerify(string exactPath, string sha256, UInt64 length, UInt32 volumeSerialNumber, UInt64 fileIndex, UInt32 numberOfLinks, string owner, string dacl)
    {
        C12SealedExecutable value = new C12SealedExecutable(exactPath, true, false, false);
        try
        {
            if (!String.Equals(value.SHA256, sha256, StringComparison.Ordinal) || value.Length != length ||
                value.VolumeSerialNumber != volumeSerialNumber || value.FileIndex != fileIndex ||
                value.NumberOfLinks != numberOfLinks || value.Reparse ||
                !String.Equals(value.Owner, owner, StringComparison.Ordinal) || !String.Equals(value.DACL, dacl, StringComparison.Ordinal))
                throw new InvalidOperationException("sealed executable receipt identity changed");
            return value;
        }
        catch { value.Dispose(); throw; }
    }

    public static C12PathIdentity InspectDirectory(string exactPath)
    {
        string path = Path.GetFullPath(exactPath).TrimEnd('\\');
        IntPtr directory = CreateFile(path, FILE_READ_ATTRIBUTES, (UInt32)(FileShare.Read | FileShare.Write | FileShare.Delete), IntPtr.Zero, OPEN_EXISTING, FILE_FLAG_BACKUP_SEMANTICS | FILE_FLAG_OPEN_REPARSE_POINT, IntPtr.Zero);
        if (directory == INVALID_HANDLE_VALUE) throw new Win32Exception(Marshal.GetLastWin32Error());
        try
        {
            BY_HANDLE_FILE_INFORMATION information = ReadInformation(directory);
            bool reparse = (information.FileAttributes & FILE_ATTRIBUTE_REPARSE_POINT) != 0;
            if ((information.FileAttributes & FILE_ATTRIBUTE_DIRECTORY) == 0 || reparse) throw new InvalidOperationException("prepared artifact directory identity is not one non-reparse directory");
            UInt64 index = ((UInt64)information.FileIndexHigh << 32) | information.FileIndexLow;
            return new C12PathIdentity(information.VolumeSerialNumber, index, information.NumberOfLinks, reparse, true);
        }
        finally { CloseHandle(directory); }
    }

    public static C12PathIdentity TryInspectPath(string exactPath)
    {
        string path = Path.GetFullPath(exactPath).TrimEnd('\\');
        IntPtr value = CreateFile(path, FILE_READ_ATTRIBUTES, (UInt32)(FileShare.Read | FileShare.Write | FileShare.Delete), IntPtr.Zero, OPEN_EXISTING, FILE_FLAG_BACKUP_SEMANTICS | FILE_FLAG_OPEN_REPARSE_POINT, IntPtr.Zero);
        if (value == INVALID_HANDLE_VALUE)
        {
            int error = Marshal.GetLastWin32Error();
            if (error == 2 || error == 3) return null;
            throw new Win32Exception(error);
        }
        try
        {
            BY_HANDLE_FILE_INFORMATION information = ReadInformation(value);
            UInt64 index = ((UInt64)information.FileIndexHigh << 32) | information.FileIndexLow;
            bool reparse = (information.FileAttributes & FILE_ATTRIBUTE_REPARSE_POINT) != 0;
            bool directory = (information.FileAttributes & FILE_ATTRIBUTE_DIRECTORY) != 0;
            return new C12PathIdentity(information.VolumeSerialNumber, index, information.NumberOfLinks, reparse, directory);
        }
        finally { CloseHandle(value); }
    }

    public void DeleteExact()
    {
        if (handle == IntPtr.Zero || handle == INVALID_HANDLE_VALUE) throw new ObjectDisposedException("C12SealedExecutable");
        BY_HANDLE_FILE_INFORMATION information = ReadInformation(handle);
        UInt64 currentIndex = ((UInt64)information.FileIndexHigh << 32) | information.FileIndexLow;
        if ((information.FileAttributes & (FILE_ATTRIBUTE_DIRECTORY | FILE_ATTRIBUTE_REPARSE_POINT)) != 0 ||
            information.VolumeSerialNumber != VolumeSerialNumber || currentIndex != FileIndex || information.NumberOfLinks != 1)
            throw new InvalidOperationException("exact cleanup file identity or link count changed");
        bool preferLegacy;
        AppContext.TryGetSwitch("Talenro.C12.PreferLegacyDeleteDisposition", out preferLegacy);
        FILE_DISPOSITION_INFO_EX disposition = new FILE_DISPOSITION_INFO_EX();
        disposition.Flags = FILE_DISPOSITION_FLAG_DELETE | FILE_DISPOSITION_FLAG_POSIX_SEMANTICS | FILE_DISPOSITION_FLAG_IGNORE_READONLY_ATTRIBUTE;
        int length = Marshal.SizeOf(typeof(FILE_DISPOSITION_INFO_EX));
        IntPtr pointer = Marshal.AllocHGlobal(length);
        try
        {
            Marshal.StructureToPtr(disposition, pointer, false);
            if (preferLegacy || !SetFileInformationByHandle(handle, FILE_DISPOSITION_INFO_EX_CLASS, pointer, (UInt32)length))
            {
                FILE_DISPOSITION_INFO legacy = new FILE_DISPOSITION_INFO(); legacy.DeleteFile = true;
                Marshal.StructureToPtr(legacy, pointer, false);
                if (!SetFileInformationByHandle(handle, FILE_DISPOSITION_INFO_CLASS, pointer, (UInt32)Marshal.SizeOf(typeof(FILE_DISPOSITION_INFO)))) throw new Win32Exception(Marshal.GetLastWin32Error());
            }
        }
        finally { Marshal.FreeHGlobal(pointer); }
    }

    public void ReleaseDeletePending() { Dispose(); }

    public void Dispose()
    {
        IntPtr current = handle;
        handle = IntPtr.Zero;
        if (current != IntPtr.Zero && current != INVALID_HANDLE_VALUE && !CloseHandle(current)) throw new Win32Exception(Marshal.GetLastWin32Error());
        GC.SuppressFinalize(this);
    }

    ~C12SealedExecutable()
    {
        IntPtr current = handle;
        handle = IntPtr.Zero;
        if (current != IntPtr.Zero && current != INVALID_HANDLE_VALUE) CloseHandle(current);
    }
}
'@
Add-Type -TypeDefinition $script:c12SealedExecutableSource

$script:c12ContainedNativeScript = {
  param($Invocation)
  try {
    $assemblyName = New-Object Reflection.AssemblyName("TalenroC12NativeMember_$([Guid]::NewGuid().ToString('N'))")
    $assemblyBuilder = [AppDomain]::CurrentDomain.DefineDynamicAssembly($assemblyName, [Reflection.Emit.AssemblyBuilderAccess]::Run)
    $moduleBuilder = $assemblyBuilder.DefineDynamicModule('TalenroC12NativeMember')
    $typeAttributes = [Reflection.TypeAttributes]::Public -bor [Reflection.TypeAttributes]::Sealed -bor [Reflection.TypeAttributes]::Abstract
    $typeBuilder = $moduleBuilder.DefineType('TalenroC12NativeMember', $typeAttributes)
    $methodAttributes = [Reflection.MethodAttributes]::Public -bor [Reflection.MethodAttributes]::Static -bor [Reflection.MethodAttributes]::PinvokeImpl
    $openMethodBuilder = $typeBuilder.DefinePInvokeMethod(
      'OpenJobObject', 'kernel32.dll', $methodAttributes, [Reflection.CallingConventions]::Standard,
      [IntPtr], [Type[]]@([UInt32], [bool], [string]),
      [Runtime.InteropServices.CallingConvention]::Winapi, [Runtime.InteropServices.CharSet]::Unicode)
    $currentMethodBuilder = $typeBuilder.DefinePInvokeMethod(
      'GetCurrentProcess', 'kernel32.dll', $methodAttributes, [Reflection.CallingConventions]::Standard,
      [IntPtr], [Type[]]@(),
      [Runtime.InteropServices.CallingConvention]::Winapi, [Runtime.InteropServices.CharSet]::None)
    $assignMethodBuilder = $typeBuilder.DefinePInvokeMethod(
      'AssignProcessToJobObject', 'kernel32.dll', $methodAttributes, [Reflection.CallingConventions]::Standard,
      [bool], [Type[]]@([IntPtr], [IntPtr]),
      [Runtime.InteropServices.CallingConvention]::Winapi, [Runtime.InteropServices.CharSet]::None)
    $closeMethodBuilder = $typeBuilder.DefinePInvokeMethod(
      'CloseHandle', 'kernel32.dll', $methodAttributes, [Reflection.CallingConventions]::Standard,
      [bool], [Type[]]@([IntPtr]),
      [Runtime.InteropServices.CallingConvention]::Winapi, [Runtime.InteropServices.CharSet]::None)
    foreach ($methodBuilder in @($openMethodBuilder, $currentMethodBuilder, $assignMethodBuilder, $closeMethodBuilder)) {
      $methodBuilder.SetImplementationFlags($methodBuilder.GetMethodImplementationFlags() -bor [Reflection.MethodImplAttributes]::PreserveSig)
    }
    $memberType = $typeBuilder.CreateType()
    $openMethod = $memberType.GetMethod('OpenJobObject')
    $currentMethod = $memberType.GetMethod('GetCurrentProcess')
    $assignMethod = $memberType.GetMethod('AssignProcessToJobObject')
    $closeMethod = $memberType.GetMethod('CloseHandle')
    $memberJobHandle = [IntPtr]$openMethod.Invoke($null, [object[]]@([UInt32]1, $false, [string]$Invocation.JobName))
    if ($memberJobHandle -eq [IntPtr]::Zero) {
      throw "open named native Job failed with Win32 error $([Runtime.InteropServices.Marshal]::GetLastWin32Error())"
    }
    try {
      $currentProcess = [IntPtr]$currentMethod.Invoke($null, [object[]]@())
      if (-not [bool]$assignMethod.Invoke($null, [object[]]@($memberJobHandle, $currentProcess))) {
        throw "assign current process to named native Job failed with Win32 error $([Runtime.InteropServices.Marshal]::GetLastWin32Error())"
      }
    }
    finally {
      if (-not [bool]$closeMethod.Invoke($null, [object[]]@($memberJobHandle))) {
        throw "close child-side named native Job handle failed with Win32 error $([Runtime.InteropServices.Marshal]::GetLastWin32Error())"
      }
    }
  }
  catch {
    [pscustomobject]@{ ExitCode = 127; Output = [string[]]@(); ContainmentFailure = $true }
    return
  }

  Set-Location -LiteralPath ([string]$Invocation.WorkingDirectory)
  foreach ($inheritedName in @([System.Environment]::GetEnvironmentVariables('Process').Keys)) {
    $name = [string]$inheritedName
    if ($name.StartsWith('GIT_', [StringComparison]::OrdinalIgnoreCase)) {
      [System.Environment]::SetEnvironmentVariable($name, $null, 'Process')
    }
    if ([string]$Invocation.Executable -ceq 'docker' -and $name.StartsWith('DOCKER_', [StringComparison]::OrdinalIgnoreCase)) {
      [System.Environment]::SetEnvironmentVariable($name, $null, 'Process')
    }
  }
  foreach ($entry in @($Invocation.Environment)) {
    [System.Environment]::SetEnvironmentVariable([string]$entry.Name, [string]$entry.Value, 'Process')
  }
  $nativeArgs = @($Invocation.Arguments | ForEach-Object { [string]$_ })
  if ($Invocation.PSObject.Properties.Name -contains 'GraphReaderSource' -and $Invocation.GraphReaderSource) {
    # This process has already joined the exact controller Job; its native Go
    # child inherits membership before it can execute any graph discovery.
    try {
      Add-Type -TypeDefinition ([string]$Invocation.GraphReaderSource)
      $records = [C12GoGraphReader]::Run([string]$Invocation.ResolvedExecutable, [string[]]$nativeArgs, [string]$Invocation.WorkingDirectory, [DateTime]$Invocation.Deadline)
      [pscustomobject]@{ ExitCode = 0; Output = [string[]]$records; ContainmentFailure = $false }
    }
    catch { [pscustomobject]@{ ExitCode = 126; Output = [string[]]@($_.Exception.Message); ContainmentFailure = $false } }
    return
  }
  $boundedOutput = New-Object 'System.Collections.Generic.List[string]'
  $capture = {
    process {
      if ($boundedOutput.Count -lt 4096) {
        $line = ([string]$_) -replace '[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]', '?'
        if ($line.Length -gt 512) {
          $line = $line.Substring(0, 512)
        }
        $boundedOutput.Add($line)
      }
    }
  }
  $priorNativeErrorActionPreference = $ErrorActionPreference
  try {
    $ErrorActionPreference = 'Continue'
    switch ([string]$Invocation.Executable) {
      'docker' {
        $dockerArgs = $nativeArgs
        & docker @dockerArgs 2>&1 | & $capture
        $nativeExitCode = $LASTEXITCODE
      }
      'go' {
        $goArgs = $nativeArgs
        if ($Invocation.PSObject.Properties.Name -contains 'ResolvedExecutable' -and -not [string]::IsNullOrEmpty([string]$Invocation.ResolvedExecutable)) {
          $goExecutable = [string]$Invocation.ResolvedExecutable
          & $goExecutable @goArgs 2>&1 | & $capture
        }
        else {
          & go @goArgs 2>&1 | & $capture
        }
        $nativeExitCode = $LASTEXITCODE
      }
      'git' {
        $gitArgs = $nativeArgs
        & git @gitArgs 2>&1 | & $capture
        $nativeExitCode = $LASTEXITCODE
      }
      default {
        $nativeExitCode = 127
      }
    }
  }
  catch {
    $nativeExitCode = 127
  }
  finally {
    $ErrorActionPreference = $priorNativeErrorActionPreference
  }
  [pscustomobject]@{ ExitCode = [int]$nativeExitCode; Output = [string[]]$boundedOutput.ToArray(); ContainmentFailure = $false }
}

function Invoke-C12Native {
  param(
    [Parameter(Mandatory = $true)]
    [ValidateSet('docker', 'go', 'git')]
    [string]$Executable,

    [Parameter(Mandatory = $true)]
    [string[]]$Arguments,

    [Parameter(Mandatory = $true)]
    [string]$Stage,

    [Parameter(Mandatory = $true)]
    [TimeSpan]$Timeout,

    [Parameter(Mandatory = $true)]
    [string]$WorkingDirectory,

    [DateTime]$Deadline = [DateTime]::MaxValue,

    [hashtable]$Environment = @{},

    [string]$ResolvedExecutable = '',

    [string]$GraphReaderSource = '',

    [switch]$AllowFailure
  )

  if ($Deadline -eq [DateTime]::MaxValue -and $script:c12NativeDeadline -ne [DateTime]::MaxValue) {
    $Deadline = $script:c12NativeDeadline
  }
  if ($Timeout -le [TimeSpan]::Zero) {
    throw "$Stage has a non-positive watchdog timeout"
  }
  $remaining = $Deadline - [DateTime]::UtcNow
  if ($remaining -le [TimeSpan]::Zero) {
    throw "$Stage exceeded its absolute deadline"
  }
  if ($remaining -lt $Timeout) {
    $Timeout = $remaining
  }
  if ($Executable -cne 'git' -and $Environment.Count -ne 0) {
    throw "$Stage attempted to expose Git environment overrides to a non-Git child"
  }
  if (-not [string]::IsNullOrEmpty($ResolvedExecutable)) {
    $resolvedNativeExecutable = [IO.Path]::GetFullPath($ResolvedExecutable)
    $resolvedNativeInfo = Get-Item -LiteralPath $resolvedNativeExecutable -Force
    if ($Executable -cne 'go' -or $resolvedNativeInfo.PSIsContainer -or ($resolvedNativeInfo.Attributes -band [IO.FileAttributes]::ReparsePoint) -ne 0 -or $resolvedNativeInfo.Length -le 0) {
      throw "$Stage attempted to use an invalid resolved closed executable"
    }
    $ResolvedExecutable = $resolvedNativeExecutable
  }
  $childEnvironment = @()
  foreach ($name in @($Environment.Keys | Sort-Object)) {
    if ([string]$name -notmatch '^GIT_[A-Z0-9_]+$') {
      throw "$Stage has an invalid per-call Git environment override"
    }
    $childEnvironment += [pscustomobject]@{ Name = [string]$name; Value = [string]$Environment[$name] }
  }
  $watchdogSuffix = New-C12RandomSuffix
  $nativeJobName = "TalenroC12Native_$watchdogSuffix"
  if ($nativeJobName -notmatch '^TalenroC12Native_[0-9a-f]{32}$') {
    throw "$Stage generated a malformed named native Job identity"
  }
  $invocation = [pscustomobject]@{
    Executable = $Executable
    Arguments = [string[]]$Arguments
    WorkingDirectory = $WorkingDirectory
    JobName = $nativeJobName
    Environment = [object[]]$childEnvironment
    ResolvedExecutable = $ResolvedExecutable
    GraphReaderSource = $GraphReaderSource
    Deadline = $Deadline
  }
  $job = $null
  $nativeJobHandle = [IntPtr]::Zero
  $watch = [System.Diagnostics.Stopwatch]::StartNew()
  try {
    $nativeJobHandle = [C12NativeJob]::CreateKillOnClose($nativeJobName)
    $job = Start-Job -ArgumentList $invocation -ScriptBlock $script:c12ContainedNativeScript
    $nativeRemaining = $Timeout - $watch.Elapsed
    if ($nativeRemaining -le [TimeSpan]::Zero) {
      [C12NativeJob]::Close($nativeJobHandle)
      $nativeJobHandle = [IntPtr]::Zero
      throw "$Stage timed out before native execution"
    }
    $waitSeconds = [int][Math]::Floor($nativeRemaining.TotalSeconds)
    if ($waitSeconds -lt 1) {
      [C12NativeJob]::Close($nativeJobHandle)
      $nativeJobHandle = [IntPtr]::Zero
      throw "$Stage has less than one bounded second remaining"
    }
    $completed = $null -ne (Wait-Job -Job $job -Timeout $waitSeconds)
    if (-not $completed) {
      [C12NativeJob]::Close($nativeJobHandle)
      $nativeJobHandle = [IntPtr]::Zero
      throw "$Stage timed out after $waitSeconds seconds"
    }
    $received = @(Receive-Job -Job $job -ErrorAction SilentlyContinue)
    $result = @($received | Where-Object { $_.PSObject.Properties.Name -contains 'ExitCode' } | Select-Object -Last 1)
    if ($result.Count -ne 1) {
      throw "$Stage did not return a bounded native result"
    }
    if (-not ($result[0].PSObject.Properties.Name -contains 'ContainmentFailure') -or [bool]$result[0].ContainmentFailure) {
      throw "$Stage failed native Job self-containment"
    }
    if (-not $AllowFailure -and [int]$result[0].ExitCode -ne 0) {
      if ($GraphReaderSource) { throw "$Stage failed with exit code $([int]$result[0].ExitCode): $(@($result[0].Output) -join ' ')" }
      throw "$Stage failed with exit code $([int]$result[0].ExitCode)"
    }
    return [pscustomobject]@{
      ExitCode = [int]$result[0].ExitCode
      Output = @($result[0].Output | ForEach-Object { [string]$_ })
    }
  }
  finally {
    if ($nativeJobHandle -ne [IntPtr]::Zero) {
      [C12NativeJob]::Close($nativeJobHandle)
      $nativeJobHandle = [IntPtr]::Zero
    }
    if ($null -ne $job) {
      Stop-Job -Job $job -ErrorAction SilentlyContinue
      Remove-Job -Job $job -Force -ErrorAction SilentlyContinue
    }
  }
}

function Invoke-C12Git {
  param(
    [Parameter(Mandatory = $true)]
    [string[]]$Arguments,

    [Parameter(Mandatory = $true)]
    [string]$Stage,

    [Parameter(Mandatory = $true)]
    [string]$WorkingDirectory,

    [DateTime]$Deadline = [DateTime]::MaxValue,

    [hashtable]$Environment = @{},

    [switch]$AllowFailure
  )

  return Invoke-C12Native -Executable 'git' -Arguments $Arguments -Stage $Stage -Timeout ([TimeSpan]::FromSeconds(30)) -WorkingDirectory $WorkingDirectory -Deadline $Deadline -Environment $Environment -AllowFailure:$AllowFailure
}

function Get-C12BoundedWaitSeconds {
  param(
    [Parameter(Mandatory = $true)]
    [DateTime]$Deadline,

    [Parameter(Mandatory = $true)]
    [int]$MaximumSeconds,

    [Parameter(Mandatory = $true)]
    [string]$Stage
  )

  if ($Deadline -eq [DateTime]::MaxValue) {
    return $MaximumSeconds
  }
  $remainingSeconds = [Math]::Floor(($Deadline - [DateTime]::UtcNow).TotalSeconds)
  if ($remainingSeconds -lt 1) {
    throw "$Stage has no bounded second remaining before its absolute deadline"
  }
  return [int][Math]::Min([double]$MaximumSeconds, $remainingSeconds)
}

function New-C12OwnedDirectory {
  param(
    [Parameter(Mandatory = $true)]
    [string]$Root,

    [Parameter(Mandatory = $true)]
    [string]$ExpectedParent,

    [Parameter(Mandatory = $true)]
    [string]$LeafPattern,

    [Parameter(Mandatory = $true)]
    [string]$Stage
  )

  $resolvedParent = [IO.Path]::GetFullPath($ExpectedParent).TrimEnd('\')
  $resolvedRoot = [IO.Path]::GetFullPath($Root).TrimEnd('\')
  if ([IO.Path]::GetDirectoryName($resolvedRoot).TrimEnd('\') -cne $resolvedParent -or
      [IO.Path]::GetFileName($resolvedRoot) -notmatch $LeafPattern) {
    throw "$Stage refused a malformed exact owned-directory path"
  }
  return [C12OwnedDirectory]::CreateNew($resolvedRoot)
}

function Remove-C12BoundedDirectory {
  param(
    [Parameter(Mandatory = $true)]
    [string]$Root,

    [Parameter(Mandatory = $true)]
    [string]$ExpectedParent,

    [Parameter(Mandatory = $true)]
    [string]$LeafPattern,

    [Parameter(Mandatory = $true)]
    [string]$Stage,

    [DateTime]$Deadline = [DateTime]::MaxValue,

    [object]$Ownership = $null
  )

  $resolvedRoot = [string]$Root
  $cleanupOwnership = $Ownership
  try {
    $resolvedParent = [IO.Path]::GetFullPath($ExpectedParent).TrimEnd('\')
    $resolvedRoot = [IO.Path]::GetFullPath($Root).TrimEnd('\')
    if ([IO.Path]::GetDirectoryName($resolvedRoot).TrimEnd('\') -cne $resolvedParent -or
        [IO.Path]::GetFileName($resolvedRoot) -notmatch $LeafPattern) {
      throw "$Stage refused a malformed exact cleanup path"
    }
    if ($Deadline -ne [DateTime]::MaxValue -and [DateTime]::UtcNow -ge $Deadline) {
      throw "$Stage exceeded its absolute deadline before identity verification"
    }
    if ($null -eq $cleanupOwnership) {
      if (-not [IO.Directory]::Exists($resolvedRoot) -and -not [IO.File]::Exists($resolvedRoot)) {
        return
      }
      throw "$Stage refused an existing exact orphan without creation ownership: $resolvedRoot"
    }
    elseif ([string]$cleanupOwnership.RootPath -cne $resolvedRoot) {
      throw "$Stage refused a mismatched owned-directory handle"
    }
    $cleanupOwnership.VerifyExactPath()
    $cleanupOwnership.DeleteExactTree($Deadline)
  }
  catch {
    throw "$Stage could not identity-bind cleanup for exact orphan $resolvedRoot`: $($_.Exception.Message)"
  }
  finally {
    if ($null -ne $cleanupOwnership) {
      $cleanupOwnership.Dispose()
    }
  }
}

function Remove-C12CandidateSnapshot {
  param(
    [Parameter(Mandatory = $true)]
    [string]$OwnerRoot,

    [object]$Ownership = $null,

    [DateTime]$Deadline = [DateTime]::MaxValue
  )

  Remove-C12BoundedDirectory -Root $OwnerRoot -ExpectedParent ([IO.Path]::GetTempPath()) -LeafPattern '^talenro-c12-candidate-[0-9a-f]{32}$' -Stage 'staged candidate cleanup' -Deadline $Deadline -Ownership $Ownership
}

function Remove-C12CandidateMaterialization {
  param(
    [Parameter(Mandatory = $true)]
    [object]$Candidate,

    [Parameter(Mandatory = $true)]
    [object]$Snapshot,

    [DateTime]$Deadline = [DateTime]::MaxValue
  )

  Remove-C12BoundedDirectory -Root ([string]$Snapshot.OwnerRoot) -ExpectedParent ([string]$Candidate.OwnerRoot) -LeafPattern '^snapshot-[0-9a-f]{32}$' -Stage 'candidate group snapshot cleanup' -Deadline $Deadline -Ownership $Snapshot.Ownership
}

function Copy-C12CandidateIndex {
  param(
    [Parameter(Mandatory = $true)]
    [string]$Source,

    [Parameter(Mandatory = $true)]
    [string]$Destination,

    [DateTime]$Deadline = [DateTime]::MaxValue
  )

  $sourceLength = (Get-Item -LiteralPath $Source).Length
  if ($sourceLength -lt 1 -or $sourceLength -gt 134217728) {
    throw 'staged candidate index exceeds the bounded 128 MiB copy limit'
  }
  $copyJob = Start-Job -ArgumentList @($Source, $Destination) -ScriptBlock {
    param($ExactSource, $ExactDestination)
    [IO.File]::Copy([string]$ExactSource, [string]$ExactDestination, $false)
  }
  try {
    $waitSeconds = Get-C12BoundedWaitSeconds -Deadline $Deadline -MaximumSeconds 5 -Stage 'staged candidate index copy'
    if ($null -eq (Wait-Job -Job $copyJob -Timeout $waitSeconds)) {
      Stop-Job -Job $copyJob -ErrorAction SilentlyContinue
      throw "staged candidate index copy timed out after $waitSeconds seconds"
    }
    $null = Receive-Job -Job $copyJob -ErrorAction Stop
  }
  finally {
    Stop-Job -Job $copyJob -ErrorAction SilentlyContinue
    Remove-Job -Job $copyJob -Force -ErrorAction SilentlyContinue
  }
  if (-not [IO.File]::Exists($Destination) -or (Get-Item -LiteralPath $Destination).Length -ne $sourceLength) {
    throw 'staged candidate index copy did not produce one exact bounded file'
  }
}

function Resolve-C12TrustedGitContext {
  param(
    [Parameter(Mandatory = $true)]
    [string]$ScriptRoot,

    [DateTime]$Deadline = [DateTime]::MaxValue
  )

  $trustedRoot = [IO.Path]::GetFullPath($ScriptRoot).TrimEnd('\')
  $topResult = Invoke-C12Git -Arguments @('-C', $trustedRoot, 'rev-parse', '--show-toplevel') -Stage 'resolve trusted repository root' -WorkingDirectory $trustedRoot -Deadline $Deadline
  $topLines = @($topResult.Output | Where-Object { -not [string]::IsNullOrWhiteSpace($_) })
  if ($topLines.Count -ne 1) {
    throw 'trusted repository resolution did not return one exact top level'
  }
  $resolvedTop = [IO.Path]::GetFullPath([string]$topLines[0]).TrimEnd('\')
  if (-not [string]::Equals($resolvedTop, $trustedRoot, [StringComparison]::OrdinalIgnoreCase)) {
    throw 'invoked script root is not the exact trusted Git top level'
  }
  $gitDirectoryResult = Invoke-C12Git -Arguments @('-C', $trustedRoot, 'rev-parse', '--absolute-git-dir') -Stage 'resolve trusted Git directory' -WorkingDirectory $trustedRoot -Deadline $Deadline
  $gitDirectoryLines = @($gitDirectoryResult.Output | Where-Object { -not [string]::IsNullOrWhiteSpace($_) })
  if ($gitDirectoryLines.Count -ne 1) {
    throw 'trusted Git directory resolution did not return one exact path'
  }
  $gitDirectory = [IO.Path]::GetFullPath([string]$gitDirectoryLines[0]).TrimEnd('\')
  if (-not [IO.Directory]::Exists($gitDirectory)) {
    throw 'trusted Git directory is absent'
  }
  return [pscustomobject]@{ WorkTree = $trustedRoot; GitDirectory = $gitDirectory }
}

function Get-C12CandidateGitArguments {
  param(
    [Parameter(Mandatory = $true)]
    [object]$Candidate
  )

  return @("--git-dir=$([string]$Candidate.GitDirectory)", "--work-tree=$([string]$Candidate.WorkTree)")
}

function Get-C12CandidateGitEnvironment {
  param(
    [Parameter(Mandatory = $true)]
    [object]$Candidate,

    [string]$IndexPath = ''
  )

  if ([string]::IsNullOrEmpty($IndexPath)) {
    $IndexPath = [string]$Candidate.Index
  }
  return @{
    GIT_INDEX_FILE = $IndexPath
    GIT_OBJECT_DIRECTORY = [string]$Candidate.ObjectDirectory
    GIT_ALTERNATE_OBJECT_DIRECTORIES = [string]$Candidate.AlternateObjectDirectory
  }
}

function New-C12CandidateSnapshot {
  $suffix = New-C12RandomSuffix
  $ownerRoot = Join-Path ([IO.Path]::GetTempPath()) "talenro-c12-candidate-$suffix"
  $candidateIndex = Join-Path $ownerRoot 'candidate.index'
  $candidateObjects = Join-Path $ownerRoot 'objects'
  $ownership = $null
  try {
    $ownership = New-C12OwnedDirectory -Root $ownerRoot -ExpectedParent ([IO.Path]::GetTempPath()) -LeafPattern '^talenro-c12-candidate-[0-9a-f]{32}$' -Stage 'staged candidate creation'
    [void][IO.Directory]::CreateDirectory($candidateObjects)
    $context = Resolve-C12TrustedGitContext -ScriptRoot $script:c12RepositoryRoot -Deadline $script:c12SuiteDeadline
    $gitArguments = @("--git-dir=$([string]$context.GitDirectory)", "--work-tree=$([string]$context.WorkTree)")
    $indexResult = Invoke-C12Git -Arguments @($gitArguments + @('rev-parse', '--path-format=absolute', '--git-path', 'index')) -Stage 'locate staged candidate index' -WorkingDirectory ([string]$context.WorkTree) -Deadline $script:c12SuiteDeadline
    $indexLines = @($indexResult.Output | Where-Object { -not [string]::IsNullOrWhiteSpace($_) })
    if ($indexLines.Count -ne 1) {
      throw 'locate staged candidate index did not return one exact path'
    }
    $sourceIndex = [IO.Path]::GetFullPath([string]$indexLines[0])
    if (-not [IO.File]::Exists($sourceIndex)) {
      throw 'staged candidate index is absent'
    }
    $objectsResult = Invoke-C12Git -Arguments @($gitArguments + @('rev-parse', '--path-format=absolute', '--git-path', 'objects')) -Stage 'locate common Git object directory' -WorkingDirectory ([string]$context.WorkTree) -Deadline $script:c12SuiteDeadline
    $objectsLines = @($objectsResult.Output | Where-Object { -not [string]::IsNullOrWhiteSpace($_) })
    if ($objectsLines.Count -ne 1) {
      throw 'locate common Git object directory did not return one exact path'
    }
    $commonObjects = [IO.Path]::GetFullPath([string]$objectsLines[0])
    if (-not [IO.Directory]::Exists($commonObjects) -or $commonObjects.Contains([string][IO.Path]::PathSeparator)) {
      throw 'common Git object directory is absent or is not one exact alternate path'
    }
    Copy-C12CandidateIndex -Source $sourceIndex -Destination $candidateIndex -Deadline $script:c12SuiteDeadline
    $candidate = [pscustomobject]@{
      OwnerRoot = $ownerRoot
      Tree = ''
      Index = $candidateIndex
      ObjectDirectory = $candidateObjects
      AlternateObjectDirectory = $commonObjects
      WorkTree = [string]$context.WorkTree
      GitDirectory = [string]$context.GitDirectory
      Ownership = $ownership
      MaterializationDigest = ''
      MaterializationFileCount = 0
      MaterializationTotalBytes = 0
    }
    $candidateGitEnvironment = Get-C12CandidateGitEnvironment -Candidate $candidate
    $treeResult = Invoke-C12Git -Arguments @($gitArguments + @('write-tree')) -Stage 'capture staged candidate tree' -WorkingDirectory ([string]$context.WorkTree) -Deadline $script:c12SuiteDeadline -Environment $candidateGitEnvironment
    $treeLines = @($treeResult.Output | Where-Object { -not [string]::IsNullOrWhiteSpace($_) })
    if ($treeLines.Count -ne 1 -or [string]$treeLines[0] -notmatch '^[0-9a-f]{40}$') {
      throw 'capture staged candidate tree did not return one exact tree identity'
    }
    $tree = [string]$treeLines[0]
    $candidate.Tree = $tree
    return $candidate
  }
  catch {
    Remove-C12CandidateSnapshot -OwnerRoot $ownerRoot -Ownership $ownership -Deadline $script:c12SuiteDeadline
    throw
  }
}

function Get-C12CandidateSnapshotDigest {
  param(
    [Parameter(Mandatory = $true)]
    [string]$SnapshotRoot,

    [DateTime]$Deadline = [DateTime]::MaxValue
  )

  $waitSeconds = Get-C12BoundedWaitSeconds -Deadline $Deadline -MaximumSeconds 30 -Stage 'candidate snapshot byte digest'
  $digestJob = Start-Job -ArgumentList $SnapshotRoot -ScriptBlock {
    param($ExactSnapshotRoot)
    $root = [IO.Path]::GetFullPath([string]$ExactSnapshotRoot).TrimEnd('\')
    $files = New-Object 'System.Collections.Generic.List[string]'
    $directories = New-Object 'System.Collections.Generic.Stack[string]'
    $directories.Push($root)
    while ($directories.Count -gt 0) {
      $directory = $directories.Pop()
      foreach ($childDirectory in [IO.Directory]::GetDirectories($directory)) {
        if (([IO.File]::GetAttributes($childDirectory) -band [IO.FileAttributes]::ReparsePoint) -ne 0) {
          throw 'candidate snapshot contains a directory reparse point'
        }
        $directories.Push($childDirectory)
      }
      foreach ($file in [IO.Directory]::GetFiles($directory)) {
        if (([IO.File]::GetAttributes($file) -band [IO.FileAttributes]::ReparsePoint) -ne 0) {
          throw 'candidate snapshot contains a file reparse point'
        }
        $files.Add($file)
        if ($files.Count -gt 100000) {
          throw 'candidate snapshot exceeds the bounded file count'
        }
      }
    }
    [string[]]$relativeFiles = @($files | ForEach-Object { $_.Substring($root.Length).TrimStart('\').Replace('\', '/') })
    [Array]::Sort($relativeFiles, [StringComparer]::Ordinal)
    [long]$totalBytes = 0
    $hash = [Security.Cryptography.SHA256]::Create()
    $sink = [IO.Stream]::Null
    $crypto = New-Object Security.Cryptography.CryptoStream($sink, $hash, [Security.Cryptography.CryptoStreamMode]::Write)
    $writer = New-Object IO.BinaryWriter($crypto, (New-Object Text.UTF8Encoding($false)), $true)
    try {
      $writer.Write([byte[]][Text.Encoding]::UTF8.GetBytes("TALENRO-C12-SNAPSHOT-DIGEST-V1`0"))
      $writer.Write([uint32]$relativeFiles.Count)
      foreach ($relativePath in $relativeFiles) {
        $fullPath = Join-Path $root ($relativePath.Replace('/', '\'))
        $pathBytes = [Text.Encoding]::UTF8.GetBytes($relativePath)
        $fileLength = (Get-Item -LiteralPath $fullPath).Length
        $totalBytes += $fileLength
        if ($pathBytes.Length -gt 4096 -or $fileLength -gt 1073741824 -or $totalBytes -gt 4294967296) {
          throw 'candidate snapshot exceeds a bounded digest dimension'
        }
        $writer.Write([uint32]$pathBytes.Length)
        $writer.Write([byte[]]$pathBytes)
        $writer.Write([uint64]$fileLength)
        $writer.Flush()
        $stream = [IO.File]::Open($fullPath, [IO.FileMode]::Open, [IO.FileAccess]::Read, [IO.FileShare]::Read)
        try {
          $stream.CopyTo($crypto)
        }
        finally {
          $stream.Dispose()
        }
      }
      $writer.Flush()
      $crypto.FlushFinalBlock()
      $digest = (($hash.Hash | ForEach-Object { $_.ToString('x2') }) -join '')
      [pscustomobject]@{ Digest = $digest; FileCount = [int]$relativeFiles.Count; TotalBytes = $totalBytes }
    }
    finally {
      $writer.Dispose()
      $crypto.Dispose()
      $hash.Dispose()
    }
  }
  try {
    if ($null -eq (Wait-Job -Job $digestJob -Timeout $waitSeconds)) {
      Stop-Job -Job $digestJob -ErrorAction SilentlyContinue
      throw "candidate snapshot byte digest timed out after $waitSeconds seconds"
    }
    $result = @(Receive-Job -Job $digestJob -ErrorAction Stop | Where-Object { $_.PSObject.Properties.Name -contains 'Digest' })
    if ($result.Count -ne 1 -or [string]$result[0].Digest -notmatch '^[0-9a-f]{64}$') {
      throw 'candidate snapshot byte digest did not return one bounded identity'
    }
    if ($Deadline -ne [DateTime]::MaxValue -and [DateTime]::UtcNow -ge $Deadline) {
      throw 'candidate snapshot byte digest exceeded its absolute deadline'
    }
    return [pscustomobject]@{ Digest = [string]$result[0].Digest; FileCount = [int]$result[0].FileCount; TotalBytes = [long]$result[0].TotalBytes }
  }
  finally {
    Stop-Job -Job $digestJob -ErrorAction SilentlyContinue
    Remove-Job -Job $digestJob -Force -ErrorAction SilentlyContinue
  }
}

function Assert-C12CandidateSnapshotClean {
  param(
    [Parameter(Mandatory = $true)]
    [object]$Candidate,

    [Parameter(Mandatory = $true)]
    [object]$Snapshot,

    [DateTime]$Deadline = [DateTime]::MaxValue
  )

  $gitArguments = Get-C12CandidateGitArguments -Candidate $Candidate
  $environment = Get-C12CandidateGitEnvironment -Candidate $Candidate -IndexPath ([string]$Snapshot.Index)
  $treeResult = Invoke-C12Git -Arguments @($gitArguments + @('write-tree')) -Stage 'verify candidate snapshot tree identity' -WorkingDirectory ([string]$Candidate.WorkTree) -Deadline $Deadline -Environment $environment
  $treeLines = @($treeResult.Output | Where-Object { -not [string]::IsNullOrWhiteSpace($_) })
  $currentDigest = Get-C12CandidateSnapshotDigest -SnapshotRoot ([string]$Snapshot.Root) -Deadline $Deadline
  if ($treeLines.Count -ne 1 -or [string]$treeLines[0] -cne [string]$Candidate.Tree -or
      [string]$currentDigest.Digest -cne [string]$Candidate.MaterializationDigest -or
      [int]$currentDigest.FileCount -ne [int]$Candidate.MaterializationFileCount -or
      [long]$currentDigest.TotalBytes -ne [long]$Candidate.MaterializationTotalBytes -or
      [string]$currentDigest.Digest -cne [string]$Snapshot.Digest) {
    throw 'candidate snapshot integrity check failed'
  }
}

function New-C12CandidateMaterialization {
  param(
    [Parameter(Mandatory = $true)]
    [object]$Candidate,

    [Parameter(Mandatory = $true)]
    [string]$Stage,

    [DateTime]$Deadline = [DateTime]::MaxValue
  )

  $snapshotOwner = Join-Path ([string]$Candidate.OwnerRoot) "snapshot-$(New-C12RandomSuffix)"
  $snapshotRoot = Join-Path $snapshotOwner 'tree'
  $snapshotIndex = Join-Path $snapshotOwner 'snapshot.index'
  $snapshotOwnership = $null
  $snapshot = [pscustomobject]@{ OwnerRoot = $snapshotOwner; Root = $snapshotRoot; Index = $snapshotIndex; Ownership = $null; Digest = ''; FileCount = 0; TotalBytes = 0 }
  try {
    $snapshotOwnership = New-C12OwnedDirectory -Root $snapshotOwner -ExpectedParent ([string]$Candidate.OwnerRoot) -LeafPattern '^snapshot-[0-9a-f]{32}$' -Stage "$Stage owner creation"
    $snapshot.Ownership = $snapshotOwnership
    [void][IO.Directory]::CreateDirectory($snapshotRoot)
    $gitArguments = Get-C12CandidateGitArguments -Candidate $Candidate
    Copy-C12CandidateIndex -Source ([string]$Candidate.Index) -Destination $snapshotIndex -Deadline $Deadline
    $environment = Get-C12CandidateGitEnvironment -Candidate $Candidate -IndexPath $snapshotIndex
    $prefix = $snapshotRoot.TrimEnd('\') + '\'
    $null = Invoke-C12Git -Arguments @($gitArguments + @('checkout-index', '--all', "--prefix=$prefix")) -Stage $Stage -WorkingDirectory ([string]$Candidate.WorkTree) -Deadline $Deadline -Environment $environment
    $baseline = Get-C12CandidateSnapshotDigest -SnapshotRoot $snapshotRoot -Deadline $Deadline
    $snapshot.Digest = [string]$baseline.Digest
    $snapshot.FileCount = [int]$baseline.FileCount
    $snapshot.TotalBytes = [long]$baseline.TotalBytes
    if ([string]::IsNullOrEmpty([string]$Candidate.MaterializationDigest)) {
      $Candidate.MaterializationDigest = [string]$baseline.Digest
      $Candidate.MaterializationFileCount = [int]$baseline.FileCount
      $Candidate.MaterializationTotalBytes = [long]$baseline.TotalBytes
    }
    elseif ([string]$baseline.Digest -cne [string]$Candidate.MaterializationDigest -or
            [int]$baseline.FileCount -ne [int]$Candidate.MaterializationFileCount -or
            [long]$baseline.TotalBytes -ne [long]$Candidate.MaterializationTotalBytes) {
      throw 'candidate snapshot integrity check failed before native execution'
    }
    Assert-C12CandidateSnapshotClean -Candidate $Candidate -Snapshot $snapshot -Deadline $Deadline
    return $snapshot
  }
  catch {
    Remove-C12CandidateMaterialization -Candidate $Candidate -Snapshot $snapshot -Deadline $Deadline
    throw
  }
}

function Get-C12DockerEndpointReceipt {
  param([Parameter(Mandatory = $true)][DateTime]$Deadline,[switch]$Revalidate)


  $command = Get-Command docker -CommandType Application -ErrorAction Stop | Select-Object -First 1
  $path = [IO.Path]::GetFullPath([string]$command.Source)
  $bytes = [IO.File]::ReadAllBytes($path)
  $executableDigest = Get-C12SHA256Hex -Bytes $bytes
  $contextResult = Invoke-C12Native -Executable 'docker' -Arguments @('context','show') -Stage 'freeze Docker context receipt' -Timeout ([TimeSpan]::FromSeconds(5)) -WorkingDirectory $script:c12RepositoryRoot -Deadline $Deadline
  $contextName = Get-C12SingleOutputLine -Result $contextResult -Stage 'freeze Docker context receipt'
  $contextInspect = Invoke-C12Native -Executable 'docker' -Arguments @('context','inspect',$contextName) -Stage 'freeze Docker endpoint receipt' -Timeout ([TimeSpan]::FromSeconds(5)) -WorkingDirectory $script:c12RepositoryRoot -Deadline $Deadline
  $contextJSON = Get-C12SingleOutputLine -Result $contextInspect -Stage 'freeze Docker endpoint receipt'
  $nameMatch=[regex]::Match($contextJSON,'"Name"\s*:\s*"(?<value>[^"\\\x00-\x1f]+)"',[Text.RegularExpressions.RegexOptions]::CultureInvariant)
  $endpointMatch=[regex]::Match($contextJSON,'"docker"\s*:\s*\{[^{}]*"Host"\s*:\s*"(?<value>[^"\\\x00-\x1f]+)"',[Text.RegularExpressions.RegexOptions]::CultureInvariant)
  if (-not $nameMatch.Success -or -not $endpointMatch.Success -or $nameMatch.Groups['value'].Value -cne $contextName) { throw 'Docker endpoint receipt context_name mismatch or malformed exact context' }
  $endpoint = [string]$endpointMatch.Groups['value'].Value
  if ([string]::IsNullOrWhiteSpace($endpoint)) { throw 'Docker endpoint receipt endpoint is empty' }
  $engineResult = Invoke-C12Native -Executable 'docker' -Arguments @('info','--format','{{.ID}}|{{.ServerVersion}}|{{.OSType}}|{{.Architecture}}') -Stage 'freeze Docker engine receipt' -Timeout ([TimeSpan]::FromSeconds(5)) -WorkingDirectory $script:c12RepositoryRoot -Deadline $Deadline
  $engine = (Get-C12SingleOutputLine -Result $engineResult -Stage 'freeze Docker engine receipt') -split '\|', 4
  if ($engine.Count -ne 4 -or @($engine | Where-Object { [string]::IsNullOrWhiteSpace($_) }).Count -ne 0) { throw 'Docker engine receipt is malformed' }
  $selectors = [ordered]@{}
  foreach ($name in @('DOCKER_HOST','DOCKER_CONTEXT','DOCKER_CONFIG','DOCKER_TLS_VERIFY','DOCKER_CERT_PATH','DOCKER_API_VERSION')) { $selectors[$name] = '' }
  $canonical = 'talenro.c12.docker-endpoint.v1' + [char]0 + $contextName + [char]0 + $endpoint + [char]0 + ($engine -join ([char]0)) + [char]0 + (($selectors.GetEnumerator() | ForEach-Object { $_.Key + '=' + $_.Value }) -join ([char]0))
  $digest = Get-C12SHA256Hex -Bytes ([Text.Encoding]::UTF8.GetBytes($canonical))
  $observed = [pscustomobject]@{
    Schema='talenro.c12.docker-endpoint.v1'; context_name=$contextName; endpoint=$endpoint; engine_id=$engine[0]; server_version=$engine[1]; os_type=$engine[2]; architecture=$engine[3]
    DOCKER_HOST=''; DOCKER_CONTEXT=''; DOCKER_CONFIG=''; DOCKER_TLS_VERIFY=''; DOCKER_CERT_PATH=''; DOCKER_API_VERSION=''; Digest=$digest
    ExecutableReceipt=[pscustomobject]@{ Path=$path; Length=[long]$bytes.Length; Digest=$executableDigest }
  }
  if ($null -eq $script:c12DockerEndpointReceipt) { $script:c12DockerEndpointReceipt = $observed }
  elseif ([string]$script:c12DockerEndpointReceipt.Digest -cne $digest -or [string]$script:c12DockerEndpointReceipt.ExecutableReceipt.Digest -cne $executableDigest -or [string]$script:c12DockerEndpointReceipt.ExecutableReceipt.Path -cne $path -or [long]$script:c12DockerEndpointReceipt.ExecutableReceipt.Length -ne [long]$bytes.Length) { throw 'Docker endpoint or canonical CLI identity mismatch' }
  return $script:c12DockerEndpointReceipt
}

function Invoke-C12Docker {
  param(
    [Parameter(Mandatory = $true)]
    [string[]]$Arguments,

    [Parameter(Mandatory = $true)]
    [string]$Stage,

    [TimeSpan]$Timeout = [TimeSpan]::FromSeconds(15),

    [DateTime]$Deadline = [DateTime]::MaxValue,

    [switch]$AllowFailure
  )

  $endpointReceipt = Get-C12DockerEndpointReceipt -Deadline $Deadline -Revalidate
  $result = Invoke-C12Native -Executable 'docker' -Arguments $Arguments -Stage $Stage -Timeout $Timeout -WorkingDirectory $script:c12RepositoryRoot -Deadline $Deadline -AllowFailure:$AllowFailure
  if ($result.ExitCode -eq 0 -and $Arguments.Count -ge 4 -and $Arguments[0] -ceq 'image' -and $Arguments[1] -ceq 'inspect') {
    if ($null -eq $script:c12DockerImageReceipts) { $script:c12DockerImageReceipts = @{} }
    $reference=[string]$Arguments[-1]; $imageID=Get-C12SingleOutputLine -Result $result -Stage $Stage
    if ($imageID -notmatch '^sha256:[0-9a-f]{64}$') { throw 'Docker immutable image ID is malformed' }
    if (-not $script:c12DockerImageReceipts.ContainsKey($reference)) { $script:c12DockerImageReceipts[$reference]=$imageID }
    elseif ([string]$script:c12DockerImageReceipts[$reference] -cne $imageID) { throw 'Docker image identity mismatch' }
  }
  if ($AllowFailure -and $result.ExitCode -ne 0 -and $Arguments.Count -ge 2 -and $Arguments[1] -ceq 'inspect') {
    $nonEmpty = @($result.Output | Where-Object { -not [string]::IsNullOrWhiteSpace([string]$_) })
    $target = [string]$Arguments[-1]
    $expected = if ($Arguments[0] -ceq 'container') { 'Error response from daemon: No such container: ' + $target } elseif ($Arguments[0] -ceq 'volume') { 'Error response from daemon: get ' + $target + ': no such volume' } else { '' }
    if ($nonEmpty.Count -ne 2 -or @($nonEmpty | Where-Object { [string]$_ -ceq '[]' }).Count -ne 1 -or @($nonEmpty | Where-Object { [string]$_ -ceq $expected }).Count -ne 1) { throw ('talenro.c12.docker-exact-absence.v1 ambiguous inspect failure: ' + ($nonEmpty -join '|')) }
  }
  $result | Add-Member -NotePropertyName EndpointReceipt -NotePropertyValue $endpointReceipt
  $result | Add-Member -NotePropertyName ExecutableReceipt -NotePropertyValue $endpointReceipt.ExecutableReceipt
  return $result
}

function New-C12PITRContainerResource {
  param([Parameter(Mandatory=$true)][string]$Name,[Parameter(Mandatory=$true)][string]$Role,[Parameter(Mandatory=$true)][string]$ImageRef,[Parameter(Mandatory=$true)][string]$ImageID,[Parameter(Mandatory=$true)][string]$NonceDigest)
  return [pscustomobject]@{ Name=$Name; Kind=$Role; Role=$Role; ImageRef=$ImageRef; ImageID=$ImageID; NonceDigest=$NonceDigest; ID=''; Phase='NeverAttempted'; RetryState=[pscustomobject]@{ Attempts=0; Retained=$true } }
}

function Assert-C12PITRContainerReceipt {
  param([Parameter(Mandatory=$true)][object]$Resource,[Parameter(Mandatory=$true)][string]$RunSuffix,[Parameter(Mandatory=$true)][string]$Identity)
  $parts=$Identity -split '\|',9
  if($parts.Count-ne 9-or $parts[0]-cne[string]$Resource.ID-or $parts[1]-cne('/'+[string]$Resource.Name)-or $parts[2]-cne'true'-or $parts[3]-cne$RunSuffix-or $parts[4]-cne'authority-v7-pitr'-or $parts[5]-cne[string]$Resource.Role-or $parts[6]-cne[string]$Resource.NonceDigest-or $parts[7]-cne[string]$Resource.ImageRef-or $parts[8]-cne[string]$Resource.ImageID){throw 'PITR container exact receipt identity mismatch'}
  return $Resource
}

function Inspect-C12ExactDockerObject {
  param([Parameter(Mandatory=$true)][ValidateSet('container','volume')][string]$Kind,[Parameter(Mandatory=$true)][string]$Identity,[Parameter(Mandatory=$true)][string]$Format,[Parameter(Mandatory=$true)][DateTime]$Deadline)
  return Invoke-C12Docker -Arguments @($Kind,'inspect','--format',$Format,$Identity) -Stage "inspect exact Docker $Kind" -Deadline $Deadline -AllowFailure
}

function Converge-C12DockerRegistry {
  param([Parameter(Mandatory=$true)][AllowEmptyCollection()][object[]]$Resources,[Parameter(Mandatory=$true)][string]$RunSuffix,[Parameter(Mandatory=$true)][DateTime]$Deadline)
  Remove-C12PITRBaseResources -Resources @() -Volumes $Resources -RunSuffix $RunSuffix -Deadline $Deadline
}

function Invoke-C12Go {
  param(
    [Parameter(Mandatory = $true)]
    [string[]]$Arguments,

    [Parameter(Mandatory = $true)]
    [string]$Stage,

    [TimeSpan]$Timeout = [TimeSpan]::FromMinutes(2),

    [DateTime]$Deadline = [DateTime]::MaxValue,

    [switch]$AllowFailure
  )

  return Invoke-C12Native -Executable 'go' -Arguments $Arguments -Stage $Stage -Timeout $Timeout -WorkingDirectory $script:c12RepositoryRoot -Deadline $Deadline -AllowFailure:$AllowFailure
}

function Protect-C12PrivateArtifactRoot {
  param([Parameter(Mandatory = $true)][string]$Root)

  $currentSID = [Security.Principal.WindowsIdentity]::GetCurrent().User
  $systemSID = New-Object Security.Principal.SecurityIdentifier('S-1-5-18')
  $security = New-Object Security.AccessControl.DirectorySecurity
  $security.SetAccessRuleProtection($true, $false)
  $security.SetOwner($currentSID)
  $inheritance = [Security.AccessControl.InheritanceFlags]::ContainerInherit -bor [Security.AccessControl.InheritanceFlags]::ObjectInherit
  $propagation = [Security.AccessControl.PropagationFlags]::None
  $allow = [Security.AccessControl.AccessControlType]::Allow
  $rights = [Security.AccessControl.FileSystemRights]::FullControl
  $security.AddAccessRule((New-Object Security.AccessControl.FileSystemAccessRule($currentSID, $rights, $inheritance, $propagation, $allow)))
  $security.AddAccessRule((New-Object Security.AccessControl.FileSystemAccessRule($systemSID, $rights, $inheritance, $propagation, $allow)))
  [IO.Directory]::SetAccessControl($Root, $security)
  $observed = [IO.Directory]::GetAccessControl($Root, [Security.AccessControl.AccessControlSections]::Owner -bor [Security.AccessControl.AccessControlSections]::Access)
  $owner = [Security.Principal.SecurityIdentifier]$observed.GetOwner([Security.Principal.SecurityIdentifier])
  if ($owner.Value -cne $currentSID.Value) {
    throw 'prepared artifact root owner is not the current controller user'
  }
  $rules = @($observed.GetAccessRules($true, $false, [Security.Principal.SecurityIdentifier]))
  if ($rules.Count -ne 2) {
    throw 'prepared artifact root DACL contains an inherited or broad principal'
  }
  foreach ($rule in $rules) {
    if ($rule.AccessControlType -ne [Security.AccessControl.AccessControlType]::Allow -or
        ([string]$rule.IdentityReference.Value -cne $currentSID.Value -and [string]$rule.IdentityReference.Value -cne $systemSID.Value)) {
      throw 'prepared artifact root DACL contains an inherited or broad principal'
    }
  }
}

function Protect-C12PrivateArtifactFile {
  param([Parameter(Mandatory = $true)][string]$Path)

  $currentSID = [Security.Principal.WindowsIdentity]::GetCurrent().User
  $systemSID = New-Object Security.Principal.SecurityIdentifier('S-1-5-18')
  $security = New-Object Security.AccessControl.FileSecurity
  $security.SetAccessRuleProtection($true, $false)
  $security.SetOwner($currentSID)
  $allow = [Security.AccessControl.AccessControlType]::Allow
  $rights = [Security.AccessControl.FileSystemRights]::FullControl
  $security.AddAccessRule((New-Object Security.AccessControl.FileSystemAccessRule($currentSID, $rights, $allow)))
  $security.AddAccessRule((New-Object Security.AccessControl.FileSystemAccessRule($systemSID, $rights, $allow)))
  [IO.File]::SetAccessControl($Path, $security)
}

function New-C12DirectLeafLedger {
  $ledger = New-Object System.Collections.ArrayList
  Write-Output -NoEnumerate $ledger
}

function Get-C12DirectLeafEntry {
  param(
    [Parameter(Mandatory = $true)][object]$Ledger,
    [Parameter(Mandatory = $true)][string]$Name
  )

  $matches = @($Ledger | Where-Object { [string]$_.Name -ceq $Name })
  if ($matches.Count -ne 1) { throw "prepared direct-leaf ledger has no unique entry for $Name" }
  return $matches[0]
}

function Set-C12DirectLeafLifecycle {
  param(
    [Parameter(Mandatory = $true)][object]$Entry,
    [Parameter(Mandatory = $true)][ValidateSet('NeverAttempted','CreateAttempted','Bound','CleanIntent','Removed','Absent')][string]$Lifecycle
  )

  $states = @('NeverAttempted','CreateAttempted','Bound','CleanIntent','Removed','Absent')
  $current = [Array]::IndexOf($states, [string]$Entry.Lifecycle)
  $target = [Array]::IndexOf($states, $Lifecycle)
  if ($current -lt 0 -or $target -ne ($current + 1)) {
    throw "prepared direct-leaf lifecycle transition $($Entry.Lifecycle) -> $Lifecycle is not monotonic"
  }
  $Entry.Lifecycle = $Lifecycle
}

function Register-C12DirectLeafIntent {
  param(
    [Parameter(Mandatory = $true)][object]$Ledger,
    [Parameter(Mandatory = $true)][string]$Name,
    [Parameter(Mandatory = $true)][ValidateSet('exact_file','owned_ephemeral_subtree')][string]$Kind,
    [Parameter(Mandatory = $true)][bool]$Expected
  )

  if ($Name -ne [IO.Path]::GetFileName($Name) -or [string]::IsNullOrWhiteSpace($Name) -or $Name.IndexOf([char]0) -ge 0) {
    throw 'prepared direct-leaf creation phase refused a non-leaf name'
  }
  if (@($Ledger | Where-Object { [string]$_.Name -ceq $Name }).Count -ne 0) {
    throw "prepared direct-leaf creation phase duplicated $Name"
  }
  $entry = [pscustomobject]@{
    Name = $Name
    Kind = $Kind
    Expected = $Expected
    CreateAttempted = $false
    Identity = ''
    NumberOfLinks = [UInt32]0
    RefCount = 0
    Lifecycle = 'NeverAttempted'
    LastCleanupError = ''
    Ownership = $null
    CleanupHandle = $null
  }
  [void]$Ledger.Add($entry)
  Set-C12DirectLeafLifecycle -Entry $entry -Lifecycle 'CreateAttempted'
  $entry.CreateAttempted = $true
  return $entry
}

function Bind-C12DirectLeaf {
  param(
    [Parameter(Mandatory = $true)][object]$ArtifactRoot,
    [Parameter(Mandatory = $true)][string]$Name,
    [object]$Ownership = $null
  )

  $ArtifactRoot.Ownership.VerifyExactPath()
  $entry = Get-C12DirectLeafEntry -Ledger $ArtifactRoot.Ledger -Name $Name
  if ([string]$entry.Lifecycle -cne 'CreateAttempted') { throw 'prepared direct-leaf bind is outside its creation phase' }
  $path = Join-Path ([string]$ArtifactRoot.Root) $Name
  if ([string]$entry.Kind -ceq 'exact_file') {
    if ($null -ne $Ownership) { throw 'exact_file direct leaf cannot bind directory ownership' }
    $observed = [C12SealedExecutable]::Inspect($path)
    try {
      if ([bool]$observed.Reparse -or [UInt32]$observed.NumberOfLinks -ne 1) { throw 'prepared exact_file identity has a reparse or link splice' }
      $entry.Identity = "$([UInt32]$observed.VolumeSerialNumber):$([UInt64]$observed.FileIndex)"
      $entry.NumberOfLinks = [UInt32]$observed.NumberOfLinks
    }
    finally { $observed.Dispose() }
  }
  else {
    if ($null -eq $Ownership -or [string]$Ownership.RootPath -cne [IO.Path]::GetFullPath($path).TrimEnd('\')) {
      throw 'owned_ephemeral_subtree direct leaf lacks its exact ownership handle'
    }
    $Ownership.VerifyExactPath()
    $observed = [C12SealedExecutable]::InspectDirectory($path)
    if ([bool]$observed.Reparse -or [string]$observed.Value -cne [string]$Ownership.Identity) { throw 'prepared directory identity is reparse or changed' }
    $entry.Identity = [string]$observed.Value
    $entry.NumberOfLinks = [UInt32]$observed.NumberOfLinks
    $entry.Ownership = $Ownership
  }
  $entry.RefCount = 1
  Set-C12DirectLeafLifecycle -Entry $entry -Lifecycle 'Bound'
  return $entry
}

function Complete-C12AbsentDirectLeaf {
  param([Parameter(Mandatory = $true)][object]$Entry)

  foreach ($state in @('CreateAttempted','Bound','CleanIntent','Removed','Absent')) {
    if ([string]$Entry.Lifecycle -ceq $state) { continue }
    $states = @('NeverAttempted','CreateAttempted','Bound','CleanIntent','Removed','Absent')
    if ([Array]::IndexOf($states, [string]$Entry.Lifecycle) -lt [Array]::IndexOf($states, $state)) {
      Set-C12DirectLeafLifecycle -Entry $Entry -Lifecycle $state
    }
  }
  $Entry.RefCount = 0
}

function Request-C12DirectLeafDelete {
  param(
    [Parameter(Mandatory = $true)][object]$Entry,
    [Parameter(Mandatory = $true)][string]$Path,
    [DateTime]$Deadline = [DateTime]::MaxValue
  )

  if ([string]$Entry.Lifecycle -ceq 'Bound') { Set-C12DirectLeafLifecycle -Entry $Entry -Lifecycle 'CleanIntent' }
  elseif ([string]$Entry.Lifecycle -cne 'CleanIntent') { throw 'prepared direct leaf delete request is outside CleanIntent' }
  if ($Deadline -ne [DateTime]::MaxValue -and [DateTime]::UtcNow -ge $Deadline) { throw 'prepared direct-leaf cleanup exceeded its absolute deadline' }
  if ([string]$Entry.Kind -ceq 'exact_file') {
    if ($null -eq $Entry.CleanupHandle) { throw 'prepared exact_file delete request lacks its retained handle' }
    $Entry.CleanupHandle.DeleteExact()
  }
  else {
    if ($null -eq $Entry.Ownership) { throw 'prepared subtree delete request lacks its retained handle' }
    $Entry.Ownership.RequestDeleteExactTree($Deadline)
  }
}

function Complete-C12DirectLeafDelete {
  param(
    [Parameter(Mandatory = $true)][object]$Entry,
    [Parameter(Mandatory = $true)][string]$Path,
    [DateTime]$Deadline = [DateTime]::MaxValue
  )

  if ([string]$Entry.Lifecycle -cne 'CleanIntent') { throw 'prepared direct leaf delete completion is outside CleanIntent' }
  Release-C12DirectLeafDeletePending -Entry $Entry

  try { $observed = [C12SealedExecutable]::TryInspectPath($Path) }
  catch {
    $Entry.LastCleanupError = 'prepared direct leaf namespace could not be classified after exact-handle delete intent: ' + $_.Exception.Message
    throw $Entry.LastCleanupError
  }
  if ($null -eq $observed) {
    Set-C12DirectLeafLifecycle -Entry $Entry -Lifecycle 'Removed'
    Set-C12DirectLeafLifecycle -Entry $Entry -Lifecycle 'Absent'
    $Entry.RefCount = 0
    $Entry.LastCleanupError = ''
    return
  }
  $expectedDirectory = [string]$Entry.Kind -ceq 'owned_ephemeral_subtree'
  $observedIdentity = if ($expectedDirectory) { [string]$observed.Value } else { "$([UInt32]$observed.VolumeSerialNumber):$([UInt64]$observed.FileIndex)" }
  if ([bool]$observed.Directory -ne $expectedDirectory -or [bool]$observed.Reparse -or $observedIdentity -cne [string]$Entry.Identity -or
      (-not $expectedDirectory -and [UInt32]$observed.NumberOfLinks -ne [UInt32]$Entry.NumberOfLinks)) {
    $Entry.LastCleanupError = 'prepared direct leaf namespace contains a foreign replacement after exact-handle delete intent'
    throw $Entry.LastCleanupError
  }
  if ($Deadline -ne [DateTime]::MaxValue -and [DateTime]::UtcNow -ge $Deadline) {
    $Entry.LastCleanupError = 'prepared direct leaf namespace retained the same identity after delete intent until its deadline'
  }
  else { $Entry.LastCleanupError = 'prepared direct leaf namespace retained the same identity after exact-handle delete intent' }
  if ($expectedDirectory) {
    $Entry.Ownership = [C12OwnedDirectory]::OpenExisting($Path, [UInt32]$observed.VolumeSerialNumber, [UInt64]$observed.FileIndex)
  }
  else { $Entry.CleanupHandle = [C12SealedExecutable]::OpenForCleanup($Path) }
  throw $Entry.LastCleanupError
}

function Release-C12DirectLeafDeletePending {
  param([Parameter(Mandatory = $true)][object]$Entry)

  if ([string]$Entry.Kind -ceq 'exact_file') {
    if ($null -ne $Entry.CleanupHandle) { $Entry.CleanupHandle.ReleaseDeletePending(); $Entry.CleanupHandle = $null }
  }
  else {
    if ($null -ne $Entry.Ownership) { $Entry.Ownership.ReleaseDeletePending(); $Entry.Ownership = $null }
  }
}

function Converge-C12DirectLeafLedger {
  param(
    [Parameter(Mandatory = $true)][object]$ArtifactRoot,
    [DateTime]$Deadline = [DateTime]::MaxValue
  )

  Write-Verbose -Message 'prepared creation phase ledger validates Expected CreateAttempted Bound Removed Absent NeverAttempted exact_file owned_ephemeral_subtree go-cache go-tmp unknown direct sibling NumberOfLinks RefCount Lifecycle InspectDirectory'
  $ArtifactRoot.Ownership.VerifyExactPath()
  $observedRoot = [C12SealedExecutable]::InspectDirectory([string]$ArtifactRoot.Root)
  if ([string]$observedRoot.Value -cne [string]$ArtifactRoot.ArtifactRootIdentity -or [bool]$observedRoot.Reparse) {
    throw 'prepared artifact root identity changed before direct-leaf cleanup'
  }
  $registered = New-Object 'System.Collections.Generic.HashSet[string]' ([StringComparer]::Ordinal)
  foreach ($entry in @($ArtifactRoot.Ledger)) {
    foreach ($field in @('Name','Kind','Expected','CreateAttempted','Identity','NumberOfLinks','RefCount','Lifecycle','LastCleanupError')) {
      if ($entry.PSObject.Properties.Name -cnotcontains $field) { throw "prepared direct-leaf ledger creation phase lacks $field" }
    }
    if ([string]$entry.Kind -cnotin @('exact_file','owned_ephemeral_subtree') -or
        [string]$entry.Lifecycle -cnotin @('NeverAttempted','CreateAttempted','Bound','CleanIntent','Removed','Absent') -or
        [int]$entry.RefCount -lt 0) { throw 'prepared direct-leaf ledger contains an invalid kind, Lifecycle, or RefCount' }
    if (-not $registered.Add([string]$entry.Name)) { throw 'prepared direct-leaf ledger contains duplicate names' }
  }
  if (-not $registered.Contains('go-cache') -or -not $registered.Contains('go-tmp')) { throw 'prepared direct-leaf ledger lacks its closed cache roots' }
  foreach ($path in [IO.Directory]::EnumerateFileSystemEntries([string]$ArtifactRoot.Root, '*', [IO.SearchOption]::TopDirectoryOnly)) {
    if (-not $registered.Contains([IO.Path]::GetFileName($path))) { throw "prepared artifact cleanup refused unknown direct sibling: $path" }
  }

  # Verify every present direct leaf before deleting any, so one foreign object retains the whole root.
  foreach ($entry in @($ArtifactRoot.Ledger)) {
    $path = Join-Path ([string]$ArtifactRoot.Root) ([string]$entry.Name)
    $exists = [IO.File]::Exists($path) -or [IO.Directory]::Exists($path)
    if (-not $exists) { continue }
    try {
      if ([string]$entry.Lifecycle -cnotin @('Bound','CleanIntent') -or [string]::IsNullOrEmpty([string]$entry.Identity)) { throw 'present direct leaf was never identity-bound' }
      if ([string]$entry.Kind -ceq 'exact_file') {
        $handle = $entry.CleanupHandle
        if ($null -eq $handle) { $handle = [C12SealedExecutable]::OpenForCleanup($path) }
        if ("$([UInt32]$handle.VolumeSerialNumber):$([UInt64]$handle.FileIndex)" -cne [string]$entry.Identity -or
            [UInt32]$handle.NumberOfLinks -ne [UInt32]$entry.NumberOfLinks -or [UInt32]$handle.NumberOfLinks -ne 1 -or [bool]$handle.Reparse) {
          $handle.Dispose()
          throw 'prepared exact_file identity or NumberOfLinks changed'
        }
        $entry.CleanupHandle = $handle
      }
      else {
        if ($null -eq $entry.Ownership) { throw 'prepared owned_ephemeral_subtree lost its retained ownership' }
        $entry.Ownership.VerifyExactPath()
        $observed = [C12SealedExecutable]::InspectDirectory($path)
        if ([string]$observed.Value -cne [string]$entry.Identity -or [bool]$observed.Reparse) { throw 'prepared directory identity changed' }
      }
    }
    catch {
      $entry.LastCleanupError = $_.Exception.Message
      throw "prepared direct leaf $($entry.Name) identity validation failed: $($_.Exception.Message)"
    }
  }

  foreach ($entry in @($ArtifactRoot.Ledger | Sort-Object @{ Expression = { if ([string]$_.Kind -ceq 'exact_file') { 0 } else { 1 } } })) {
    $path = Join-Path ([string]$ArtifactRoot.Root) ([string]$entry.Name)
    try {
      if (-not ([IO.File]::Exists($path) -or [IO.Directory]::Exists($path))) {
        if ([string]$entry.Lifecycle -ceq 'CleanIntent') { Complete-C12DirectLeafDelete -Entry $entry -Path $path -Deadline $Deadline }
        Complete-C12AbsentDirectLeaf -Entry $entry
        $entry.LastCleanupError = ''
        continue
      }
      Request-C12DirectLeafDelete -Entry $entry -Path $path -Deadline $Deadline
      Complete-C12DirectLeafDelete -Entry $entry -Path $path -Deadline $Deadline
    }
    catch {
      $entry.LastCleanupError = $_.Exception.Message
      throw
    }
  }
}

function New-C12PreparedArtifactRoot {
  param(
    [Parameter(Mandatory = $true)][string]$RunSuffix,
    [Parameter(Mandatory = $true)][ValidateSet('base', 'authority-v7', 'authority-v7-pitr')][string]$Profile
  )

  if ($RunSuffix -notmatch '^[0-9a-f]{32}$') {
    throw 'prepared artifact root run suffix is malformed'
  }
  if ([string]::IsNullOrEmpty($script:c12PreparedReceiptKeyHex)) {
    $script:c12PreparedReceiptKeyHex = New-C12PITRSecretHex
  }
  $parent = [IO.Path]::GetFullPath([IO.Path]::GetTempPath()).TrimEnd('\')
  $root = Join-Path $parent "talenro-c12-artifacts-$RunSuffix"
  $ownership = New-C12OwnedDirectory -Root $root -ExpectedParent $parent -LeafPattern '^talenro-c12-artifacts-[0-9a-f]{32}$' -Stage 'prepared artifact root creation'
  $ledger = New-C12DirectLeafLedger
  $artifact = [pscustomobject]@{
    Root = $root; Parent = $parent; RunSuffix = $RunSuffix; Profile = $Profile
    ArtifactRootIdentity = [string]$ownership.Identity; ParentIdentity = ''
    GoCache = (Join-Path $root 'go-cache'); GoTemp = (Join-Path $root 'go-tmp')
    Ownership = $ownership; Ledger = $ledger; Closed = $false
    RootLifecycle = 'Bound'; RootLastCleanupError = ''
  }
  try {
    foreach ($name in @('go-cache','go-tmp')) {
      $null = Register-C12DirectLeafIntent -Ledger $ledger -Name $name -Kind 'owned_ephemeral_subtree' -Expected $true
    }
    Protect-C12PrivateArtifactRoot -Root $root
    foreach ($name in @('go-cache','go-tmp')) {
      $path = Join-Path $root $name
      $leafOwnership = New-C12OwnedDirectory -Root $path -ExpectedParent $root -LeafPattern ("^" + [Regex]::Escape($name) + "$") -Stage 'prepared artifact Go cache/temp creation phase'
      $null = Bind-C12DirectLeaf -ArtifactRoot $artifact -Name $name -Ownership $leafOwnership
    }
    $parentIdentity = [C12SealedExecutable]::InspectDirectory($parent)
    $rootIdentity = [C12SealedExecutable]::InspectDirectory($root)
    if ($rootIdentity.Value -cne [string]$ownership.Identity) {
      throw 'prepared artifact root retained identity mismatch'
    }
    $artifact.ArtifactRootIdentity = [string]$rootIdentity.Value
    $artifact.ParentIdentity = [string]$parentIdentity.Value
    return $artifact
  }
  catch {
    try {
      Converge-C12DirectLeafLedger -ArtifactRoot $artifact
      $ownership.DeleteExactEmpty([DateTime]::MaxValue)
    }
    catch { }
    throw
  }
}

function Get-C12PreparedArtifactRoot {
  param(
    [Parameter(Mandatory = $true)][string]$Profile,
    [string]$RunSuffix = ''
  )

  if ($null -ne $script:c12PreparedArtifactRoot) {
    if ([bool]$script:c12PreparedArtifactRoot.Closed -or [string]$script:c12PreparedArtifactRoot.Profile -cne $Profile -or
        (-not [string]::IsNullOrEmpty($RunSuffix) -and [string]$script:c12PreparedArtifactRoot.RunSuffix -cne $RunSuffix)) {
      throw 'prepared artifact root lifecycle does not match the active profile'
    }
    $script:c12PreparedArtifactRoot.Ownership.VerifyExactPath()
    return $script:c12PreparedArtifactRoot
  }
  if ([string]::IsNullOrEmpty($RunSuffix)) { $RunSuffix = New-C12RandomSuffix }
  $script:c12PreparedArtifactRoot = New-C12PreparedArtifactRoot -RunSuffix $RunSuffix -Profile $Profile
  if ([string]::IsNullOrEmpty($script:c12PreparedReceiptKeyHex)) {
    $script:c12PreparedReceiptKeyHex = New-C12PITRSecretHex
  }
  return $script:c12PreparedArtifactRoot
}

function Enter-C12ClosedGoBuildEnvironment {
  param(
    [Parameter(Mandatory = $true)][object]$ArtifactRoot,
    [string]$ModuleCache = ''
  )

  $names = New-Object 'System.Collections.Generic.HashSet[string]' ([StringComparer]::OrdinalIgnoreCase)
  foreach ($name in @(
    'GOFLAGS', 'GOWORK', 'GOENV', 'GOTOOLCHAIN',
    'GOOS', 'GOARCH', 'GOAMD64', 'CGO_ENABLED',
    'GOCACHE', 'GOTMPDIR', 'GOMODCACHE', 'GOPROXY',
    'GOSUMDB', 'GOPATH', 'GO111MODULE', 'GOROOT'
  )) { [void]$names.Add($name) }
  foreach ($inheritedName in @([Environment]::GetEnvironmentVariables('Process').Keys)) {
    $name = [string]$inheritedName
    if ($name.StartsWith('GIT_', [StringComparison]::OrdinalIgnoreCase)) { [void]$names.Add($name) }
  }
  $prior = @{}
  foreach ($name in $names) {
    $prior[$name] = [Environment]::GetEnvironmentVariable($name, 'Process')
    [Environment]::SetEnvironmentVariable($name, $null, 'Process')
  }
  [Environment]::SetEnvironmentVariable('GOWORK', 'off', 'Process')
  [Environment]::SetEnvironmentVariable('GOENV', 'off', 'Process')
  [Environment]::SetEnvironmentVariable('GOTOOLCHAIN', 'local', 'Process')
  [Environment]::SetEnvironmentVariable('GOOS', 'windows', 'Process')
  [Environment]::SetEnvironmentVariable('GOARCH', 'amd64', 'Process')
  [Environment]::SetEnvironmentVariable('GOAMD64', 'v1', 'Process')
  [Environment]::SetEnvironmentVariable('CGO_ENABLED', '0', 'Process')
  [Environment]::SetEnvironmentVariable('GOCACHE', [string]$ArtifactRoot.GoCache, 'Process')
  [Environment]::SetEnvironmentVariable('GOTMPDIR', [string]$ArtifactRoot.GoTemp, 'Process')
  if (-not [string]::IsNullOrEmpty($ModuleCache)) {
    [Environment]::SetEnvironmentVariable('GOMODCACHE', $ModuleCache, 'Process')
  }
  [Environment]::SetEnvironmentVariable('GOPROXY', 'off', 'Process')
  [Environment]::SetEnvironmentVariable('GOSUMDB', 'off', 'Process')
  [Environment]::SetEnvironmentVariable('GO111MODULE', 'on', 'Process')
  [Environment]::SetEnvironmentVariable('GOFLAGS', '-mod=readonly -tags=integration', 'Process')
  return [pscustomobject]@{ Prior = $prior }
}

function Exit-C12ClosedGoBuildEnvironment {
  param([Parameter(Mandatory = $true)][object]$Snapshot)
  foreach ($name in @($Snapshot.Prior.Keys)) {
    [Environment]::SetEnvironmentVariable([string]$name, $Snapshot.Prior[$name], 'Process')
  }
}

function Resolve-C12ClosedGoToolchain {
  param(
    [Parameter(Mandatory = $true)][object]$ArtifactRoot,
    [Parameter(Mandatory = $true)][DateTime]$SetupDeadline
  )

  if ([DateTime]::UtcNow -ge $SetupDeadline) { throw 'closed Go toolchain resolution exceeded setup allowance' }
  $userProfile = [Environment]::GetFolderPath([Environment+SpecialFolder]::UserProfile)
  $moduleCache = [IO.Path]::GetFullPath((Join-Path $userProfile 'go\pkg\mod'))
  $toolchainRoot = Join-Path $moduleCache "golang.org\toolchain@v0.0.1-$($script:c12GoToolchainVersion).windows-amd64"
  $exactGoExecutable = [IO.Path]::GetFullPath((Join-Path $toolchainRoot 'bin\go.exe'))
  $toolchainInfo = Get-Item -LiteralPath $exactGoExecutable -Force
  if ($toolchainInfo.PSIsContainer -or $toolchainInfo.Length -le 0 -or ($toolchainInfo.Attributes -band [IO.FileAttributes]::ReparsePoint) -ne 0) {
    throw 'closed Go toolchain executable is absent or not one regular non-reparse file'
  }
  $sealed = [C12SealedExecutable]::Inspect($exactGoExecutable)
  try {
    $snapshot = Enter-C12ClosedGoBuildEnvironment -ArtifactRoot $ArtifactRoot -ModuleCache $moduleCache
    try {
      $remaining = $SetupDeadline - [DateTime]::UtcNow
      if ($remaining -le [TimeSpan]::Zero) { throw 'closed Go toolchain version check exceeded setup allowance' }
      $versionResult = Invoke-C12Native -Executable 'go' -ResolvedExecutable $exactGoExecutable -Arguments @('version') -Stage 'resolve exact closed Go toolchain' -Timeout $remaining -WorkingDirectory ([string]$ArtifactRoot.Root) -Deadline $SetupDeadline
    }
    finally { Exit-C12ClosedGoBuildEnvironment -Snapshot $snapshot }
    $versionLines = @($versionResult.Output | Where-Object { -not [string]::IsNullOrWhiteSpace([string]$_) })
    $expectedVersion = "go version $($script:c12GoToolchainVersion) windows/amd64"
    if ($versionLines.Count -ne 1 -or [string]$versionLines[0] -cne $expectedVersion) {
      throw 'closed Go toolchain returned an unexpected version or platform'
    }
    return [pscustomobject]@{
      Path = $exactGoExecutable
      SHA256 = [string]$sealed.SHA256
      Version = $script:c12GoToolchainVersion
      GOOS = 'windows'
      GOARCH = 'amd64'
      ModuleCache = $moduleCache
      Root = [IO.Path]::GetFullPath($toolchainRoot)
    }
  }
  finally { $sealed.Dispose() }
}

function Get-C12GoGraphReaderSource {
  return @'
using System;
using System.IO;
using System.Text;
using System.Diagnostics;
using System.Collections.Generic;
using System.Threading.Tasks;
public static class C12GoGraphReader {
    static int Remaining(DateTime deadline) {
        double remaining = (deadline - DateTime.UtcNow).TotalMilliseconds;
        if (remaining <= 0) throw new TimeoutException("Go graph absolute deadline exhausted");
        return (int)Math.Min(remaining, Int32.MaxValue);
    }
    static string Quote(string value) {
        if (value == null || value.IndexOf('\0') >= 0) throw new InvalidDataException("invalid Go graph argument");
        var quoted = new StringBuilder("\""); int slashes = 0;
        foreach (char c in value) {
            if (c == '\\') { slashes++; continue; }
            quoted.Append('\\', c == '"' ? slashes * 2 + 1 : slashes); quoted.Append(c); slashes = 0;
        }
        return quoted.Append('\\', slashes * 2).Append('"').ToString();
    }
    public static string[] Decode(Stream input, DateTime deadline) {
        byte[] buffer = new byte[8192]; long total = 0;
        var records = new List<string>(); var utf8 = new UTF8Encoding(false, true);
        using (var current = new MemoryStream()) {
            int depth = 0; bool quoted = false, escaped = false;
            while (true) {
                var read = input.ReadAsync(buffer, 0, buffer.Length);
                if (!read.Wait(Remaining(deadline))) throw new TimeoutException("Go graph read deadline exhausted");
                int count = read.Result; if (count == 0) break;
                total += count; if (total > 67108864) throw new InvalidDataException("Go graph total exceeds 67108864 bytes");
                for (int i = 0; i < count; i++) {
                    byte b = buffer[i];
                    if (depth == 0) {
                        if (b == 32 || b == 9 || b == 10 || b == 13) continue;
                        if (b != 123) throw new InvalidDataException("Go graph stream must contain JSON objects only");
                    }
                    current.WriteByte(b);
                    if (current.Length > 16777216) throw new InvalidDataException("Go graph object exceeds 16777216 bytes");
                    if (quoted) { if (escaped) escaped = false; else if (b == 92) escaped = true; else if (b == 34) quoted = false; }
                    else if (b == 34) quoted = true;
                    else if (b == 123 || b == 91) depth++;
                    else if (b == 125 || b == 93) depth--;
                    if (depth == 0) { records.Add(utf8.GetString(current.ToArray())); current.SetLength(0); }
                }
            }
            if (depth != 0 || quoted || current.Length != 0 || records.Count == 0) throw new InvalidDataException("Go graph stream is empty or truncated");
        }
        return records.ToArray();
    }
    static string ReadErrors(Stream stream, DateTime deadline) {
        byte[] buffer = new byte[8192];
        using (var bytes = new MemoryStream()) {
            while (true) {
                var read = stream.ReadAsync(buffer, 0, buffer.Length);
                if (!read.Wait(Remaining(deadline))) throw new TimeoutException("Go graph stderr deadline exhausted");
                if (read.Result == 0) break;
                if (bytes.Length + read.Result > 131072) throw new InvalidDataException("Go graph stderr exceeds bound");
                bytes.Write(buffer, 0, read.Result);
            }
            return new UTF8Encoding(false, true).GetString(bytes.ToArray());
        }
    }
    public static string[] Run(string executable, string[] argv, string directory, DateTime deadline) {
        var arguments = new List<string>(); foreach (string argument in argv) arguments.Add(Quote(argument));
        using (var process = new Process()) {
            process.StartInfo = new ProcessStartInfo(executable, String.Join(" ", arguments.ToArray())) {
                UseShellExecute = false, CreateNoWindow = true, RedirectStandardOutput = true, RedirectStandardError = true, WorkingDirectory = directory
            };
            Remaining(deadline); process.Start();
            try {
                var errors = Task.Run(() => ReadErrors(process.StandardError.BaseStream, deadline));
                string[] records = Decode(process.StandardOutput.BaseStream, deadline);
                if (!process.WaitForExit(Remaining(deadline)) || !errors.Wait(Remaining(deadline))) throw new TimeoutException("Go graph exit deadline exhausted");
                if (process.ExitCode != 0 || errors.Result.Length != 0) throw new InvalidDataException("Go graph discovery failed: " + errors.Result);
                return records;
            }
            finally { if (!process.HasExited) process.Kill(); }
        }
    }
}
'@
}

function Get-C12GoGraphProperty {
  param($Value, [string]$Name, $Default = $null)
  if ($null -ne $Value -and $Value.PSObject.Properties.Name -ccontains $Name) { return $Value.$Name }
  return $Default
}

function Resolve-C12GoGraphOrigin {
  param([string]$Path, [object[]]$Roots)
  $full = [IO.Path]::GetFullPath($Path).TrimEnd('\')
  foreach ($root in $Roots) {
    $base = ([string]$root.path).TrimEnd('\')
    if ($full.Equals($base, [StringComparison]::OrdinalIgnoreCase) -or $full.StartsWith($base + '\', [StringComparison]::OrdinalIgnoreCase)) {
      # Check every component, not only the final leaf: a linked ancestor must
      # not turn a lexically-contained directory into a foreign identity.
      $cursor = $full
      while ($true) {
        $item = Get-Item -LiteralPath $cursor -Force
        if (($item.Attributes -band [IO.FileAttributes]::ReparsePoint) -ne 0) { throw 'Go graph path traverses a reparse point' }
        if ($cursor.Equals($base, [StringComparison]::OrdinalIgnoreCase)) { break }
        $cursor = [IO.Path]::GetDirectoryName($cursor)
      }
      return [pscustomobject]@{ origin = $root.origin; path = $full; origin_relative_path = $full.Substring($base.Length).TrimStart('\').Replace('\','/') }
    }
  }
  throw "Go graph directory identity is outside module/cache roots: $full"
}

function Get-C12GoGraphFile {
  param([string]$Path, [object[]]$Roots, [string]$Set, [DateTime]$Deadline)
  $null = Get-C12ProtocolMilliseconds $Deadline
  $origin = Resolve-C12GoGraphOrigin -Path $Path -Roots $Roots
  $stream = [IO.File]::Open($origin.path, [IO.FileMode]::Open, [IO.FileAccess]::Read, [IO.FileShare]::Read)
  $hash = [Security.Cryptography.SHA256]::Create()
  try {
    if ($stream.Length -gt 67108864) { throw 'Go graph selected file exceeds byte bound' }
    $digest = [BitConverter]::ToString($hash.ComputeHash($stream)).Replace('-','').ToLowerInvariant()
    $null = Get-C12ProtocolMilliseconds $Deadline
    return [ordered]@{ set = $Set; origin = $origin.origin; origin_relative_path = $origin.origin_relative_path; length = $stream.Length; sha256 = $digest }
  }
  finally { $hash.Dispose(); $stream.Dispose() }
}

function Get-C12GoGraphModule {
  param($Module, [object[]]$Roots, [DateTime]$Deadline, [int]$Depth = 0)
  if ($null -eq $Module) { return $null }
  if ($Depth -gt 4) { throw 'Go graph module replacement recursion exceeds bound' }
  $directory = [string](Get-C12GoGraphProperty $Module 'Dir' '')
  $identity = $null
  if ($directory) {
    $origin = Resolve-C12GoGraphOrigin $directory $Roots
    $identity = [C12SealedExecutable]::InspectDirectory($origin.path).Value
  }
  $modFile = [string](Get-C12GoGraphProperty $Module 'GoMod' '')
  return [ordered]@{
    module_path = [string](Get-C12GoGraphProperty $Module 'Path' '')
    module_version = [string](Get-C12GoGraphProperty $Module 'Version' '')
    module_sum = [string](Get-C12GoGraphProperty $Module 'Sum' '')
    go_mod_sum = [string](Get-C12GoGraphProperty $Module 'GoModSum' '')
    directory = $directory; directory_identity = $identity
    go_mod = if ($modFile) { Get-C12GoGraphFile $modFile $Roots 'ModuleGoMod' $Deadline } else { $null }
    module_replace_or_null = Get-C12GoGraphModule (Get-C12GoGraphProperty $Module 'Replace') $Roots $Deadline ($Depth + 1)
    metadata = $Module
  }
}

function New-C12GoGraphReceipt {
  param(
    [Parameter(Mandatory = $true)][ValidateNotNullOrEmpty()][string]$Purpose,
    [Parameter(Mandatory = $true)][ValidateNotNullOrEmpty()][string[]]$Packages,
    [Parameter(Mandatory = $true)][DateTime]$Deadline,
    $ArtifactRoot = $script:c12PreparedArtifactRoot, $GoToolchain = $null,
    [string[]]$BuildArgv = @(), [string]$WorkingDirectory = $script:c12RepositoryRoot
  )
  if ([string]::IsNullOrWhiteSpace($Purpose) -or $Deadline.Kind -ne [DateTimeKind]::Utc) { throw 'Go graph requires a purpose and absolute UTC deadline' }
  $null = Get-C12ProtocolMilliseconds $Deadline
  foreach ($package in $Packages) { if ([string]::IsNullOrWhiteSpace($package) -or $package.StartsWith('-') -or $package.IndexOf([char]0) -ge 0) { throw 'Go graph selection is invalid' } }
  if ($null -eq $GoToolchain) { $GoToolchain = Resolve-C12ClosedGoToolchain -ArtifactRoot $ArtifactRoot -SetupDeadline $Deadline }
  if ($BuildArgv.Count -eq 0) { $BuildArgv = @('test','-c','-tags=integration') + $Packages }
  $null = Get-C12PreparedStringArrayDigest $BuildArgv
  $roots = @(
    [pscustomobject]@{ origin='module'; path=[IO.Path]::GetFullPath($script:c12RepositoryRoot) },
    [pscustomobject]@{ origin='goroot'; path=[IO.Path]::GetFullPath($GoToolchain.Root) },
    [pscustomobject]@{ origin='gomodcache'; path=[IO.Path]::GetFullPath($GoToolchain.ModuleCache) },
    [pscustomobject]@{ origin='artifact'; path=[IO.Path]::GetFullPath($ArtifactRoot.Root) }
  )
  $rootIdentity = [C12SealedExecutable]::InspectDirectory($roots[0].path)
  $goRootIdentity = [C12SealedExecutable]::InspectDirectory($GoToolchain.Root)
  $cacheIdentity = [C12SealedExecutable]::InspectDirectory($GoToolchain.ModuleCache)
  foreach ($root in $roots) { $null = Resolve-C12GoGraphOrigin $root.path @($root) }
  $null = Resolve-C12GoGraphOrigin $WorkingDirectory $roots
  $go = [C12SealedExecutable]::Inspect($GoToolchain.Path)
  try {
    if ($go.SHA256 -cne $GoToolchain.SHA256 -or $GoToolchain.Version -cne $script:c12GoToolchainVersion -or $GoToolchain.GOOS -cne 'windows' -or $GoToolchain.GOARCH -cne 'amd64') { throw 'Go graph toolchain identity changed' }
    $toolchainBody = ConvertTo-C12ProtocolValue ([ordered]@{
      go_executable_path=$GoToolchain.Path; go_executable_identity="$($go.VolumeSerialNumber):$($go.FileIndex)"; go_executable_sha256=$go.SHA256; go_version=$GoToolchain.Version
      goroot_path=$GoToolchain.Root; goroot_identity=$goRootIdentity.Value; gomodcache_path=$GoToolchain.ModuleCache; gomodcache_identity=$cacheIdentity.Value
    })
  }
  finally { $go.Dispose() }
  $moduleFiles = @(
    (Get-C12GoGraphFile (Join-Path $roots[0].path 'go.mod') $roots 'go.mod' $Deadline),
    (Get-C12GoGraphFile (Join-Path $roots[0].path 'go.sum') $roots 'go.sum' $Deadline)
  )
  $listArgv = [Collections.Generic.List[string]]::new([string[]]@('list', '-deps', '-json', '-test', '-tags=integration', '-mod=readonly'))
  $listArgv.AddRange($Packages)
  $snapshot = Enter-C12ClosedGoBuildEnvironment -ArtifactRoot $ArtifactRoot -ModuleCache $GoToolchain.ModuleCache
  try {
    $environment = ConvertTo-C12ProtocolValue ([ordered]@{ goos=$env:GOOS; goarch=$env:GOARCH; cgo_enabled=$env:CGO_ENABLED; GOSUMDB=$env:GOSUMDB; GOFLAGS=$env:GOFLAGS; GOWORK=$env:GOWORK; GOENV=$env:GOENV; GOTOOLCHAIN=$env:GOTOOLCHAIN; GOPROXY=$env:GOPROXY; GOAMD64=$env:GOAMD64 })
    $verified = Invoke-C12Native -Executable 'go' -ResolvedExecutable $GoToolchain.Path -Arguments @('mod','verify') -Stage 'Go graph go mod verify' -Timeout ($Deadline - [DateTime]::UtcNow) -WorkingDirectory $roots[0].path -Deadline $Deadline
    if (($verified.Output -join "`n").Trim() -cne 'all modules verified') { throw 'Go graph go mod verify did not authenticate all modules' }
    $discovery = Invoke-C12Native -Executable 'go' -ResolvedExecutable $GoToolchain.Path -Arguments $listArgv -Stage 'Go graph discovery' -Timeout ($Deadline - [DateTime]::UtcNow) -WorkingDirectory $WorkingDirectory -Deadline $Deadline -GraphReaderSource (Get-C12GoGraphReaderSource)
  }
  finally { Exit-C12ClosedGoBuildEnvironment $snapshot }
  $packageRecords = [Collections.Generic.List[object]]::new()
  $files = [Collections.Generic.List[object]]::new()
  $importsSeen = [Collections.Generic.HashSet[string]]::new([StringComparer]::Ordinal)
  foreach ($json in $discovery.Output) {
    $null = Get-C12ProtocolMilliseconds $Deadline
    $package = ConvertFrom-Json -InputObject $json
    $importPath = [string](Get-C12GoGraphProperty $package 'ImportPath' '')
    if (-not $importPath -or -not $importsSeen.Add($importPath)) { throw "Go graph duplicate or empty ImportPath: $importPath" }
    if ((Get-C12GoGraphProperty $package 'Error') -or (Get-C12GoGraphProperty $package 'DepsErrors') -or (Get-C12GoGraphProperty $package 'Incomplete' $false)) { throw 'Go graph contains package errors' }
    $directory = [string](Get-C12GoGraphProperty $package 'Dir' '')
    if (-not $directory) { throw 'Go graph package directory is absent' }
    $origin = Resolve-C12GoGraphOrigin $directory $roots
    $directoryIdentity = [C12SealedExecutable]::InspectDirectory($origin.path)
    $sets = [ordered]@{}
    $selectedFiles = [Collections.Generic.List[object]]::new()
    foreach ($set in [Collections.Generic.List[string]]::new([string[]]@('GoFiles','CgoFiles','CFiles','CXXFiles','MFiles','HFiles','FFiles','SFiles','SwigFiles','SwigCXXFiles','SysoFiles','EmbedFiles','TestGoFiles','TestEmbedFiles','XTestGoFiles','XTestEmbedFiles'))) {
      [string[]]$names = @(Get-C12GoGraphProperty $package $set @())
      $sets[$set] = $names
      foreach ($name in $names) {
        $path = if ([IO.Path]::IsPathRooted($name)) { $name } else { Join-Path $directory $name }
        $file = Get-C12GoGraphFile $path $roots $set $Deadline
        $selectedFiles.Add($file); $files.Add($file)
      }
    }
    $module = Get-C12GoGraphModule (Get-C12GoGraphProperty $package 'Module') $roots $Deadline
    $packageRecords.Add([ordered]@{
      import_path=$importPath; for_test=[string](Get-C12GoGraphProperty $package 'ForTest' ''); origin=$origin.origin; origin_relative_path=$origin.origin_relative_path; directory_identity=$directoryIdentity.Value
      module_path=if ($module) { $module.module_path } else { '' }; module_version=if ($module) { $module.module_version } else { '' }; module_sum=if ($module) { $module.module_sum } else { '' }; module_replace_or_null=if ($module) { $module.module_replace_or_null } else { $null }; module=$module
      imports=@(Get-C12GoGraphProperty $package 'Imports' @()); deps=@(Get-C12GoGraphProperty $package 'Deps' @()); test_imports=@(Get-C12GoGraphProperty $package 'TestImports' @()); xtest_imports=@(Get-C12GoGraphProperty $package 'XTestImports' @())
      embed_patterns=@(Get-C12GoGraphProperty $package 'EmbedPatterns' @()); test_embed_patterns=@(Get-C12GoGraphProperty $package 'TestEmbedPatterns' @()); xtest_embed_patterns=@(Get-C12GoGraphProperty $package 'XTestEmbedPatterns' @())
      selected_sets=$sets; files=@($selectedFiles.ToArray()); import_map=Get-C12GoGraphProperty $package 'ImportMap'
    })
  }
  $body = ConvertTo-C12ProtocolValue ([ordered]@{
    schema='talenro-c12-go-graph/v1'; purpose=$Purpose; toolchain=$toolchainBody; environment=$environment
    module_root_path=$roots[0].path; module_root_identity=$rootIdentity.Value; go_mod_sha256=$moduleFiles[0].sha256; go_sum_sha256=$moduleFiles[1].sha256; module_files=$moduleFiles
    build_argv=@($BuildArgv); selected_packages=@($Packages); working_directory=[IO.Path]::GetFullPath($WorkingDirectory); list_argv=$listArgv; verify_argv=@('mod','verify')
    stream_read_buffer=8192; protocol_frame_limit=131072; stream_total_limit=67108864; object_limit=16777216
    package_count=$packageRecords.Count; file_count=$files.Count; packages=@($packageRecords.ToArray()); files=@($files.ToArray())
  })
  $candidateDigest = Get-C12SHA256Hex ([Text.Encoding]::UTF8.GetBytes((ConvertTo-C12ProtocolJSON $body)))
  $body.Add('candidate_tree_digest', $candidateDigest)
  [byte[]]$bodyBytes = [Text.Encoding]::UTF8.GetBytes((ConvertTo-C12ProtocolJSON $body))
  $digest = Get-C12SHA256Hex ([byte[]]([Text.Encoding]::UTF8.GetBytes("talenro.c12.go-graph.v1`0") + $bodyBytes))
  return [pscustomobject]@{
    Schema=$body.schema; Purpose=$Purpose; Toolchain=$toolchainBody; ModuleRoot=[ordered]@{ path=$roots[0].path; identity=$rootIdentity.Value }; Environment=$environment; ModuleFiles=$moduleFiles
    BuildArgv=@($BuildArgv); Packages=@($packageRecords.ToArray()); Files=@($files.ToArray()); CandidateTreeDigest=$candidateDigest; BodyBytes=$bodyBytes; Digest=$digest
  }
}

function ConvertTo-C12GoGraphBindingValue {
  param([Parameter(Mandatory = $true)]$Receipt)
  $binding = [ordered]@{}
  foreach ($property in $Receipt.PSObject.Properties) {
    if ($property.Name -ceq 'BodyBytes') {
      # Bind every original byte and its length without recursively serializing
      # hundreds of thousands of individual byte values in PowerShell.
      [byte[]]$bytes = $property.Value
      $binding[$property.Name] = [ordered]@{ encoding='base64'; length=$bytes.Length; content=[Convert]::ToBase64String($bytes) }
    }
    else { $binding[$property.Name] = $property.Value }
  }
  return $binding
}

function Assert-C12GoGraphReceipt {
  param($Receipt, $ArtifactRoot, [DateTime]$Deadline)
  # The focused entry point's legacy MaxValue sentinel is not a UTC deadline.
  # Use its existing closed profile allowance, never an unbounded native wait.
  if ($Deadline -eq [DateTime]::MaxValue) { $Deadline = [DateTime]::UtcNow.Add([TimeSpan]$script:c12ProfileAllowances[$ArtifactRoot.Profile]) }
  $body = ConvertFrom-Json -InputObject ([Text.Encoding]::UTF8.GetString([byte[]]$Receipt.BodyBytes))
  $tool = $body.toolchain
  $goToolchain = [pscustomobject]@{ Path=$tool.go_executable_path; SHA256=$tool.go_executable_sha256; Version=$tool.go_version; GOOS=$body.environment.goos; GOARCH=$body.environment.goarch; Root=$tool.goroot_path; ModuleCache=$tool.gomodcache_path }
  $current = New-C12GoGraphReceipt -Purpose $body.purpose -Packages @($body.selected_packages) -Deadline $Deadline -ArtifactRoot $ArtifactRoot -GoToolchain $goToolchain -BuildArgv @($body.build_argv) -WorkingDirectory $body.working_directory
  if ((ConvertTo-C12ProtocolJSON (ConvertTo-C12GoGraphBindingValue $current)) -cne (ConvertTo-C12ProtocolJSON (ConvertTo-C12GoGraphBindingValue $Receipt))) { throw 'Go graph receipt changed before compilation or execution' }
}

function Invoke-C12ClosedGoBuild {
  param(
    [Parameter(Mandatory = $true)][object]$ArtifactRoot,
    [Parameter(Mandatory = $true)][object]$GoToolchain,
    [Parameter(Mandatory = $true)][string[]]$Arguments,
    [Parameter(Mandatory = $true)][string]$WorkingDirectory,
    [Parameter(Mandatory = $true)][string]$Stage,
    [Parameter(Mandatory = $true)][DateTime]$SetupDeadline
  )

  if ($Arguments.Count -lt 1 -or [string]$Arguments[0] -cnotin @('build', 'test')) {
    throw "$Stage attempted an unsupported setup-only Go command"
  }
  foreach ($argument in @($Arguments)) {
    if ($null -eq $argument -or ([string]$argument).IndexOf([char]0) -ge 0 -or [string]$argument -match '^(?:-toolexec|-overlay)(?:=|$)') {
      throw "$Stage attempted a caller-controlled tool execution or overlay"
    }
  }
  $observed = [C12SealedExecutable]::Inspect([string]$GoToolchain.Path)
  try {
    if ([string]$observed.SHA256 -cne [string]$GoToolchain.SHA256 -or [string]$GoToolchain.Version -cne $script:c12GoToolchainVersion -or
        [string]$GoToolchain.GOOS -cne 'windows' -or [string]$GoToolchain.GOARCH -cne 'amd64') {
      throw "$Stage observed a changed closed Go toolchain"
    }
  }
  finally { $observed.Dispose() }
  $snapshot = Enter-C12ClosedGoBuildEnvironment -ArtifactRoot $ArtifactRoot -ModuleCache ([string]$GoToolchain.ModuleCache)
  try {
    $remaining = $SetupDeadline - [DateTime]::UtcNow
    if ($remaining -le [TimeSpan]::Zero) { throw "$Stage exceeded setup allowance before compilation" }
    return Invoke-C12Native -Executable 'go' -ResolvedExecutable ([string]$GoToolchain.Path) -Arguments $Arguments -Stage $Stage -Timeout $remaining -WorkingDirectory ([IO.Path]::GetFullPath($WorkingDirectory)) -Deadline $SetupDeadline
  }
  finally { Exit-C12ClosedGoBuildEnvironment -Snapshot $snapshot }
}

function Publish-C12PreparedExecutable {
  param(
    [Parameter(Mandatory = $true)][object]$ArtifactRoot,
    [Parameter(Mandatory = $true)][string]$TemporaryPath,
    [Parameter(Mandatory = $true)][string]$FinalPath,
    [Parameter(Mandatory = $true)][string]$TemporaryLeafPattern,
    [Parameter(Mandatory = $true)][string]$FinalLeafPattern
  )

  $temporary = [IO.Path]::GetFullPath($TemporaryPath)
  $final = [IO.Path]::GetFullPath($FinalPath)
  $root = ([string]$ArtifactRoot.Root).TrimEnd('\')
  if ([IO.Path]::GetDirectoryName($temporary).TrimEnd('\') -cne $root -or [IO.Path]::GetDirectoryName($final).TrimEnd('\') -cne $root -or
      [IO.Path]::GetFileName($temporary) -notmatch $TemporaryLeafPattern -or [IO.Path]::GetFileName($final) -notmatch $FinalLeafPattern -or
      [IO.File]::Exists($final)) {
    throw 'prepared executable publish path is outside its closed owned leaves'
  }
  $temporaryName = [IO.Path]::GetFileName($temporary)
  $finalName = [IO.Path]::GetFileName($final)
  $temporaryEntry = Get-C12DirectLeafEntry -Ledger $ArtifactRoot.Ledger -Name $temporaryName
  $finalEntry = Get-C12DirectLeafEntry -Ledger $ArtifactRoot.Ledger -Name $finalName
  if ([string]$temporaryEntry.Lifecycle -ceq 'CreateAttempted') { $null = Bind-C12DirectLeaf -ArtifactRoot $ArtifactRoot -Name $temporaryName }
  if ([string]$temporaryEntry.Lifecycle -cne 'Bound' -or [string]$finalEntry.Lifecycle -cne 'CreateAttempted') {
    throw 'prepared executable publication is outside its direct-leaf creation phase'
  }
  Protect-C12PrivateArtifactFile -Path $temporary
  $before = [C12SealedExecutable]::Inspect($temporary)
  try {
    [IO.File]::Move($temporary, $final)
    Protect-C12PrivateArtifactFile -Path $final
    $null = Bind-C12DirectLeaf -ArtifactRoot $ArtifactRoot -Name $finalName
    $after = [C12SealedExecutable]::Inspect($final)
    try {
      if ([string]$after.SHA256 -cne [string]$before.SHA256 -or [UInt64]$after.Length -ne [UInt64]$before.Length -or
          [UInt32]$after.VolumeSerialNumber -ne [UInt32]$before.VolumeSerialNumber -or [UInt64]$after.FileIndex -ne [UInt64]$before.FileIndex -or
          [UInt32]$after.NumberOfLinks -ne 1 -or [bool]$after.Reparse) {
        throw 'prepared executable changed during atomic publication'
      }
    }
    finally { $after.Dispose() }
    Complete-C12AbsentDirectLeaf -Entry $temporaryEntry
  }
  finally { $before.Dispose() }
  return $final
}

function Get-C12PreparedCandidateIdentity {
  param(
    [Parameter(Mandatory = $true)][string]$DataRoot,
    [string]$CandidateTree = ''
  )

  if (-not [string]::IsNullOrEmpty($CandidateTree)) {
    if ($CandidateTree -notmatch '^[0-9a-f]{40}$') { throw 'prepared candidate tree identity is malformed' }
    return $CandidateTree
  }
  $resolvedRoot = [IO.Path]::GetFullPath($DataRoot).TrimEnd('\')
  $identity = [C12SealedExecutable]::InspectDirectory($resolvedRoot)
  return Get-C12SHA256Hex -Bytes ([Text.Encoding]::UTF8.GetBytes("$resolvedRoot`n$($identity.Value)"))
}

function Get-C12PreparedStringArrayDigest {
  param([Parameter(Mandatory = $true)][string[]]$Values)

  $encoded = New-Object 'System.Collections.Generic.List[string]'
  foreach ($value in @($Values)) {
    if ($null -eq $value -or ([string]$value).IndexOf([char]0) -ge 0 -or ([string]$value).Length -gt 32768) {
      throw 'prepared artifact closed argument vector is malformed'
    }
    $encoded.Add([Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes([string]$value)))
  }
  return Get-C12SHA256Hex -Bytes ([Text.Encoding]::UTF8.GetBytes(($encoded -join '.')))
}

function Get-C12PreparedReceiptPayload {
  param([Parameter(Mandatory = $true)][object]$Receipt)

  $parts = New-Object 'System.Collections.Generic.List[string]'
  foreach ($field in $script:c12PreparedReceiptPayloadFields) {
    if (-not ($Receipt.PSObject.Properties.Name -ccontains $field)) {
      throw "prepared receipt lacks $field"
    }
    $value = if ($field -ceq 'GoGraphReceipts') { ConvertTo-C12ProtocolJSON @($Receipt.GoGraphReceipts | ForEach-Object { ConvertTo-C12GoGraphBindingValue $_ }) } else { [string]$Receipt.$field }
    $parts.Add("$field=$([Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes($value)))")
  }
  return [Text.Encoding]::UTF8.GetBytes(($parts -join "`n"))
}

function Get-C12PreparedReceiptSeal {
  param(
    [Parameter(Mandatory = $true)][byte[]]$Payload,
    [Parameter(Mandatory = $true)][string]$KeyHex
  )

  $key = ConvertFrom-C12Hex -Value $KeyHex
  $hmac = [Security.Cryptography.HMACSHA256]::new([byte[]]$key)
  try { return (($hmac.ComputeHash($Payload) | ForEach-Object { $_.ToString('x2') }) -join '') }
  finally { $hmac.Dispose(); [Array]::Clear($key, 0, $key.Length) }
}

function New-C12SealedExecutableReceipt {
  param(
    [Parameter(Mandatory = $true)][object]$ArtifactRoot,
    [Parameter(Mandatory = $true)][ValidateSet('trusted-validator', 'authority-initializer-json')][string]$Role,
    [Parameter(Mandatory = $true)][ValidateSet('base', 'authority-v7', 'authority-v7-pitr')][string]$Profile,
    [Parameter(Mandatory = $true)][string]$Purpose,
    [Parameter(Mandatory = $true)][string]$SourceIdentity,
    [Parameter(Mandatory = $true)][string]$SourceDigest,
    [Parameter(Mandatory = $true)][string]$CandidateTreeIdentity,
    [Parameter(Mandatory = $true)][string[]]$BuildArguments,
    [Parameter(Mandatory = $true)][object]$GoToolchain,
    [Parameter(Mandatory = $true)][string]$ExecutablePath,
    [string]$SecondaryExecutablePath = '',
    [Parameter(Mandatory = $true)][string[]]$Arguments,
    [Parameter(Mandatory = $true)][string]$WorkingDirectory,
    [object[]]$GoGraphReceipts = @()
  )

  if ($null -eq $ArtifactRoot -or [bool]$ArtifactRoot.Closed) { throw 'prepared receipt has no live artifact root' }
  $ArtifactRoot.Ownership.VerifyExactPath()
  $expectedPurposes = switch ($Purpose) {
    'trusted-validator' { @('trusted-validator') }
    'authority-initializer' { @('authority-initializer', 'test2json') }
    default { @() }
  }
  if (@($expectedPurposes).Count -ne 0) {
    if ($GoGraphReceipts.Count -ne @($expectedPurposes).Count) {
      throw 'prepared receipt requires its independent purpose-labelled Go graphs'
    }
    $remainingPurposes = [Collections.Generic.HashSet[string]]::new([StringComparer]::Ordinal)
    foreach ($expectedPurpose in $expectedPurposes) { [void]$remainingPurposes.Add($expectedPurpose) }
    foreach ($graph in $GoGraphReceipts) {
      $label = if ($null -ne $graph) { $graph.PSObject.Properties['Purpose'] } else { $null }
      if ($null -eq $label -or $label.Value -isnot [string] -or -not $remainingPurposes.Remove($label.Value)) {
        throw 'prepared receipt requires its independent purpose-labelled Go graphs'
      }
    }
  }
  if ([string]$ArtifactRoot.Profile -cne $Profile -or $Purpose -notmatch '^[a-z][a-z0-9-]{0,63}$' -or
      $SourceDigest -notmatch '^[0-9a-f]{64}$' -or $CandidateTreeIdentity -notmatch '^(?:[0-9a-f]{40}|[0-9a-f]{64})$') {
    throw 'prepared receipt closed identity is malformed'
  }
  $resolvedExecutable = [IO.Path]::GetFullPath($ExecutablePath)
  $relativeExecutable = [IO.Path]::GetFileName($resolvedExecutable)
  if ([IO.Path]::GetDirectoryName($resolvedExecutable).TrimEnd('\') -cne ([string]$ArtifactRoot.Root).TrimEnd('\')) {
    throw 'prepared receipt executable is outside its owned artifact root'
  }
  if (($Role -ceq 'trusted-validator' -and $relativeExecutable -notmatch '^trusted-validator-[0-9a-f]{32}\.exe$') -or
      ($Role -ceq 'authority-initializer-json' -and $relativeExecutable -notmatch '^authority-test2json-[0-9a-f]{32}\.exe$')) {
    throw 'prepared receipt executable leaf is outside its closed role'
  }
  if ($Role -ceq 'trusted-validator' -and -not [string]::IsNullOrEmpty($SecondaryExecutablePath)) {
    throw 'trusted-validator receipt has an unexpected secondary executable'
  }
  if ($Role -ceq 'authority-initializer-json' -and [string]::IsNullOrEmpty($SecondaryExecutablePath)) {
    throw 'authority initializer receipt lacks its sealed test binary'
  }
  Protect-C12PrivateArtifactFile -Path $resolvedExecutable
  $primary = [C12SealedExecutable]::Inspect($resolvedExecutable)
  $secondary = $null
  try {
    $secondaryPath = ''
    if (-not [string]::IsNullOrEmpty($SecondaryExecutablePath)) {
      $secondaryPath = [IO.Path]::GetFullPath($SecondaryExecutablePath)
      if ([IO.Path]::GetDirectoryName($secondaryPath).TrimEnd('\') -cne ([string]$ArtifactRoot.Root).TrimEnd('\') -or
          [IO.Path]::GetFileName($secondaryPath) -notmatch '^authority-initializer-[0-9a-f]{32}\.test\.exe$') {
        throw 'authority initializer test binary is outside its closed leaf'
      }
      Protect-C12PrivateArtifactFile -Path $secondaryPath
      $secondary = [C12SealedExecutable]::Inspect($secondaryPath)
    }
    $currentSID = [Security.Principal.WindowsIdentity]::GetCurrent().User.Value
    foreach ($sealed in @($primary, $secondary)) {
      if ($null -ne $sealed -and [string]$sealed.Owner -cne $currentSID -and [string]$sealed.Owner -cne 'S-1-5-18') {
        throw 'prepared receipt executable has a foreign owner'
      }
    }
    $buildDigest = Get-C12PreparedStringArrayDigest -Values $BuildArguments
    $argumentsDigest = Get-C12PreparedStringArrayDigest -Values $Arguments
    $oneShotNonce = New-C12RandomSuffix
    $nativeJobName = "TalenroC12Prepared_$(New-C12RandomSuffix)"
    $gateName = "TalenroC12PreparedGate_$(New-C12RandomSuffix)"
    $pipeName = "TalenroC12PreparedPipe_$(New-C12RandomSuffix)"
    $receipt = [pscustomobject]@{
      Schema = 'talenro-c12-prepared-executable-receipt/v1'
      Role = $Role
      RunSuffix = [string]$ArtifactRoot.RunSuffix
      Profile = $Profile
      Purpose = $Purpose
      OneShotNonce = $oneShotNonce
      OwnedRootPath = [string]$ArtifactRoot.Root
      ArtifactRootIdentity = [string]$ArtifactRoot.ArtifactRootIdentity
      ParentRootPath = [string]$ArtifactRoot.Parent
      ParentIdentity = [string]$ArtifactRoot.ParentIdentity
      SourceIdentity = $SourceIdentity
      SourceDigest = $SourceDigest
      CandidateTreeIdentity = $CandidateTreeIdentity
      GoGraphReceipts = @($GoGraphReceipts)
      BuildArguments = [string[]]$BuildArguments
      BuildArgumentsDigest = $buildDigest
      GoExecutablePath = [string]$GoToolchain.Path
      GoExecutableSHA256 = [string]$GoToolchain.SHA256
      GoVersion = [string]$GoToolchain.Version
      GOOS = [string]$GoToolchain.GOOS
      GOARCH = [string]$GoToolchain.GOARCH
      ExecutablePath = $resolvedExecutable
      ExecutableSHA256 = [string]$primary.SHA256
      ExecutableLength = ([UInt64]$primary.Length).ToString()
      ExecutableVolumeSerial = ([UInt32]$primary.VolumeSerialNumber).ToString()
      ExecutableFileIndex = ([UInt64]$primary.FileIndex).ToString()
      ExecutableLinkCount = ([UInt32]$primary.NumberOfLinks).ToString()
      ExecutableOwner = [string]$primary.Owner
      ExecutableDACL = [string]$primary.DACL
      ExecutableReparse = ([bool]$primary.Reparse).ToString().ToLowerInvariant()
      SecondaryExecutablePath = $secondaryPath
      SecondaryExecutableSHA256 = if ($null -eq $secondary) { '' } else { [string]$secondary.SHA256 }
      SecondaryExecutableLength = if ($null -eq $secondary) { '0' } else { ([UInt64]$secondary.Length).ToString() }
      SecondaryExecutableVolumeSerial = if ($null -eq $secondary) { '0' } else { ([UInt32]$secondary.VolumeSerialNumber).ToString() }
      SecondaryExecutableFileIndex = if ($null -eq $secondary) { '0' } else { ([UInt64]$secondary.FileIndex).ToString() }
      SecondaryExecutableLinkCount = if ($null -eq $secondary) { '0' } else { ([UInt32]$secondary.NumberOfLinks).ToString() }
      SecondaryExecutableOwner = if ($null -eq $secondary) { '' } else { [string]$secondary.Owner }
      SecondaryExecutableDACL = if ($null -eq $secondary) { '' } else { [string]$secondary.DACL }
      SecondaryExecutableReparse = if ($null -eq $secondary) { 'false' } else { ([bool]$secondary.Reparse).ToString().ToLowerInvariant() }
      Arguments = [string[]]$Arguments
      ArgumentsDigest = $argumentsDigest
      WorkingDirectory = [IO.Path]::GetFullPath($WorkingDirectory)
      NativeJobName = $nativeJobName
      GateName = $gateName
      PipeName = $pipeName
      ReceiptIdentity = ''
      ReceiptSeal = ''
    }
    $payload = Get-C12PreparedReceiptPayload -Receipt $receipt
    $receipt.ReceiptIdentity = Get-C12SHA256Hex -Bytes $payload
    $receipt.ReceiptSeal = Get-C12PreparedReceiptSeal -Payload $payload -KeyHex $script:c12PreparedReceiptKeyHex
    $dedupeKey = "$($receipt.Purpose):$($receipt.ArtifactRootIdentity):$($receipt.ReceiptIdentity)"
    if ($script:c12PreparedReceipts.ContainsKey($dedupeKey)) {
      throw 'prepared receipt identity was duplicated'
    }
    $script:c12PreparedReceipts.Add($dedupeKey, $receipt)
    return $receipt
  }
  finally {
    if ($null -ne $secondary) { $secondary.Dispose() }
    $primary.Dispose()
  }
}

function Assert-C12SealedExecutableReceipt {
  param(
    [Parameter(Mandatory = $true)][object]$ArtifactRoot,
    [Parameter(Mandatory = $true)][object]$Receipt
  )

  $expectedFields = @($script:c12PreparedReceiptFields | Sort-Object)
  $observedFields = @($Receipt.PSObject.Properties.Name | Sort-Object)
  if (($observedFields -join '|') -cne ($expectedFields -join '|')) { throw 'prepared receipt has a missing or extra field' }
  if ([string]$Receipt.Schema -cne 'talenro-c12-prepared-executable-receipt/v1' -or [string]$Receipt.Role -cnotin $script:c12PreparedAuthorityRoles) {
    throw 'prepared receipt schema or closed role is invalid'
  }
  if ([bool]$ArtifactRoot.Closed -or [string]$Receipt.OwnedRootPath -cne [string]$ArtifactRoot.Root -or
      [string]$Receipt.ArtifactRootIdentity -cne [string]$ArtifactRoot.ArtifactRootIdentity -or
      [string]$Receipt.ParentIdentity -cne [string]$ArtifactRoot.ParentIdentity) {
    throw 'prepared receipt artifact root identity changed'
  }
  $ArtifactRoot.Ownership.VerifyExactPath()
  $rootIdentity = [C12SealedExecutable]::InspectDirectory([string]$ArtifactRoot.Root)
  $parentIdentity = [C12SealedExecutable]::InspectDirectory([string]$ArtifactRoot.Parent)
  if ($rootIdentity.Value -cne [string]$Receipt.ArtifactRootIdentity -or $parentIdentity.Value -cne [string]$Receipt.ParentIdentity) {
    throw 'prepared receipt root or parent file ID changed'
  }
  if ((Get-C12PreparedStringArrayDigest -Values @($Receipt.BuildArguments)) -cne [string]$Receipt.BuildArgumentsDigest -or
      (Get-C12PreparedStringArrayDigest -Values @($Receipt.Arguments)) -cne [string]$Receipt.ArgumentsDigest) {
    throw 'prepared receipt closed argument vector changed'
  }
  $payload = Get-C12PreparedReceiptPayload -Receipt $Receipt
  if ((Get-C12SHA256Hex -Bytes $payload) -cne [string]$Receipt.ReceiptIdentity -or
      (Get-C12PreparedReceiptSeal -Payload $payload -KeyHex $script:c12PreparedReceiptKeyHex) -cne [string]$Receipt.ReceiptSeal) {
    throw 'prepared receipt identity or seal changed'
  }
  $primary = [C12SealedExecutable]::OpenAndVerify(
    [string]$Receipt.ExecutablePath, [string]$Receipt.ExecutableSHA256, [UInt64]::Parse([string]$Receipt.ExecutableLength),
    [UInt32]::Parse([string]$Receipt.ExecutableVolumeSerial), [UInt64]::Parse([string]$Receipt.ExecutableFileIndex),
    [UInt32]::Parse([string]$Receipt.ExecutableLinkCount), [string]$Receipt.ExecutableOwner, [string]$Receipt.ExecutableDACL)
  $secondary = $null
  try {
    if (-not [string]::IsNullOrEmpty([string]$Receipt.SecondaryExecutablePath)) {
      $secondary = [C12SealedExecutable]::OpenAndVerify(
        [string]$Receipt.SecondaryExecutablePath, [string]$Receipt.SecondaryExecutableSHA256, [UInt64]::Parse([string]$Receipt.SecondaryExecutableLength),
        [UInt32]::Parse([string]$Receipt.SecondaryExecutableVolumeSerial), [UInt64]::Parse([string]$Receipt.SecondaryExecutableFileIndex),
        [UInt32]::Parse([string]$Receipt.SecondaryExecutableLinkCount), [string]$Receipt.SecondaryExecutableOwner, [string]$Receipt.SecondaryExecutableDACL)
    }
    return [pscustomobject]@{ Primary = $primary; Secondary = $secondary }
  }
  catch {
    if ($null -ne $secondary) { $secondary.Dispose() }
    $primary.Dispose()
    throw
  }
}

function Initialize-C12PreparedPipe {
  if ('C12PreparedPipe' -as [type]) { return }
  Add-Type -TypeDefinition @'
using System;
using System.ComponentModel;
using System.IO;
using System.IO.Pipes;
using System.Runtime.InteropServices;
using System.Security.Cryptography;
using System.Text;
using Microsoft.Win32.SafeHandles;
public static class C12PreparedPipe {
    const uint PIPE_ACCESS_DUPLEX = 3, FILE_FLAG_FIRST_PIPE_INSTANCE = 0x00080000, FILE_FLAG_OVERLAPPED = 0x40000000;
    const uint PIPE_TYPE_MESSAGE = 4, PIPE_READMODE_MESSAGE = 2, PIPE_WAIT = 0, PIPE_REJECT_REMOTE_CLIENTS = 8;
    [DllImport("kernel32.dll", CharSet=CharSet.Unicode, SetLastError=true)]
    static extern SafePipeHandle CreateNamedPipeW(string name, uint access, uint mode, uint nMaxInstances, uint outSize, uint inSize, uint timeout, IntPtr security);
    [DllImport("kernel32.dll", SetLastError=true)] static extern bool GetNamedPipeClientProcessId(SafePipeHandle pipe, out uint pid);
    [DllImport("kernel32.dll", SetLastError=true)] static extern bool GetNamedPipeServerProcessId(SafePipeHandle pipe, out uint pid);
    public static NamedPipeServerStream CreatePrivatePipe(string name) {
        var handle = CreateNamedPipeW(@"\\.\pipe\" + name,
            PIPE_ACCESS_DUPLEX | FILE_FLAG_FIRST_PIPE_INSTANCE | FILE_FLAG_OVERLAPPED,
            PIPE_TYPE_MESSAGE | PIPE_READMODE_MESSAGE | PIPE_WAIT | PIPE_REJECT_REMOTE_CLIENTS,
            nMaxInstances: 1, outSize: 131072, inSize: 131072, timeout: 0, security: IntPtr.Zero);
        if (handle.IsInvalid) { handle.Dispose(); throw new Win32Exception(Marshal.GetLastWin32Error(), "private pipe creation failed"); }
        try { return new NamedPipeServerStream(PipeDirection.InOut, true, false, handle); }
        catch { handle.Dispose(); throw; }
    }
    public static void VerifyClientProcessId(PipeStream pipe, uint expected) {
        uint pid;
        if (!GetNamedPipeClientProcessId(pipe.SafePipeHandle, out pid) || pid != expected || expected == 0)
            throw new InvalidOperationException("protocol client PID mismatch");
    }
    public static void VerifyServerProcessId(PipeStream pipe, uint expected) {
        uint pid;
        if (!GetNamedPipeServerProcessId(pipe.SafePipeHandle, out pid) || pid != expected || expected == 0)
            throw new InvalidOperationException("protocol server PID mismatch");
    }
    static void Wait(IAsyncResult operation, int milliseconds, PipeStream pipe) {
        if (milliseconds <= 0 || !operation.AsyncWaitHandle.WaitOne(milliseconds)) {
            pipe.Dispose(); throw new TimeoutException("protocol pipe deadline exhausted");
        }
    }
    public static void Connect(NamedPipeServerStream pipe, int milliseconds) {
        var operation = pipe.BeginWaitForConnection(null, null);
        try { Wait(operation, milliseconds, pipe); pipe.EndWaitForConnection(operation); pipe.ReadMode = PipeTransmissionMode.Message; }
        finally { operation.AsyncWaitHandle.Close(); }
    }
    public static byte[] ReadMessage(PipeStream pipe, int milliseconds) {
        byte[] bytes = new byte[131073];
        var operation = pipe.BeginRead(bytes, 0, bytes.Length, null, null);
        int count;
        try { Wait(operation, milliseconds, pipe); count = pipe.EndRead(operation); }
        finally { operation.AsyncWaitHandle.Close(); }
        if (count == 0 || count > 131072 || !pipe.IsMessageComplete) throw new InvalidDataException("protocol frame exceeds message bound or is truncated");
        Array.Resize(ref bytes, count);
        return bytes;
    }
    public static void WriteMessage(PipeStream pipe, byte[] bytes, int milliseconds) {
        if (bytes.Length == 0 || bytes.Length > 131072) throw new InvalidDataException("protocol frame exceeds message bound");
        var operation = pipe.BeginWrite(bytes, 0, bytes.Length, null, null);
        try { Wait(operation, milliseconds, pipe); pipe.EndWrite(operation); }
        finally { operation.AsyncWaitHandle.Close(); }
    }
    public static string SHA256(byte[] bytes) {
        using (var digest = System.Security.Cryptography.SHA256.Create()) { return BitConverter.ToString(digest.ComputeHash(bytes)).Replace("-", "").ToLowerInvariant(); }
    }
    public static string HMAC(byte[] bytes, byte[] key) {
        using (var mac = new HMACSHA256(key)) { return BitConverter.ToString(mac.ComputeHash(bytes)).Replace("-", "").ToLowerInvariant(); }
    }
    public static bool FixedEquals(string left, string right) {
        if (left == null || right == null || left.Length != 64 || right.Length != 64) return false;
        int difference = 0;
        for (int i = 0; i < 64; i++) difference |= left[i] ^ right[i];
        return difference == 0;
    }
}
'@
}

function Get-C12ProtocolMilliseconds {
  param([DateTime]$Deadline)
  $remaining = [Math]::Floor(($Deadline - [DateTime]::UtcNow).TotalMilliseconds)
  if ($remaining -le 0) { throw 'protocol absolute deadline exhausted' }
  return [int][Math]::Min($remaining, [Int32]::MaxValue)
}

function ConvertTo-C12ProtocolValue {
  param($Value)
  if ($null -eq $Value -or $Value -is [string] -or $Value -is [ValueType]) { return $Value }
  if ($Value -is [Collections.IDictionary] -or $Value -is [pscustomobject]) {
    [string[]]$names = if ($Value -is [Collections.IDictionary]) { @($Value.Keys) } else { @($Value.PSObject.Properties.Name) }
    [Array]::Sort($names, [StringComparer]::Ordinal)
    $ordered = [ordered]@{}
    foreach ($name in $names) {
      if ($Value -is [Collections.IDictionary]) { $item = $Value[$name] } else { $item = $Value.$name }
      $ordered[$name] = ConvertTo-C12ProtocolValue $item
    }
    return $ordered
  }
  if ($Value -is [Collections.IEnumerable]) {
    $items = @($Value | ForEach-Object { ConvertTo-C12ProtocolValue $_ })
    return ,$items
  }
  throw 'protocol value is not a JSON data type'
}

function ConvertTo-C12ProtocolJSON {
  param($Value)
  # All object keys, including nested payloads, use ordinal order. Acceptance
  # additionally requires byte-identical, compact, BOM-free UTF-8 re-encoding.
  return ConvertTo-Json -InputObject (ConvertTo-C12ProtocolValue $Value) -Compress -Depth 30
}

function New-C12ProtocolState {
  param([string]$Run, [string]$Nonce, [byte[]]$Key)
  return [pscustomobject]@{ Run = $Run; Nonce = $Nonce; Key = $Key; Sequence = 0; Previous = ''; Frames = [Collections.Generic.List[object]]::new() }
}

function New-C12PreparedFrame {
  param($State, [string]$Type, $Payload)
  $utf8 = [Text.UTF8Encoding]::new($false, $true)
  $body = [ordered]@{
    schema = 'talenro-c12-prepared-frame/v1'; run = $State.Run; worker_nonce = $State.Nonce
    sequence = $State.Sequence + 1; type = $Type; payload = $Payload
    payload_digest = [C12PreparedPipe]::SHA256($utf8.GetBytes((ConvertTo-C12ProtocolJSON $Payload)))
    previous_frame_digest = $State.Previous
  }
  $mac = [C12PreparedPipe]::HMAC($utf8.GetBytes((ConvertTo-C12ProtocolJSON $body)), $State.Key)
  $body.Add('hmac_sha256', $mac)
  return ,$utf8.GetBytes((ConvertTo-C12ProtocolJSON $body))
}

function Assert-C12PreparedFrame {
  param($State, [byte[]]$Bytes, [string]$Type)
  if ($Bytes.Length -eq 0 -or $Bytes.Length -gt 131072) { throw 'protocol frame exceeds maximum size' }
  $utf8 = [Text.UTF8Encoding]::new($false, $true)
  try { $text = $utf8.GetString($Bytes); $frame = $text | ConvertFrom-Json }
  catch { throw 'protocol frame is not valid UTF-8 JSON' }
  $fields = @('hmac_sha256','payload','payload_digest','previous_frame_digest','run','schema','sequence','type','worker_nonce')
  if ((@($frame.PSObject.Properties.Name) -join '|') -cne ($fields -join '|')) { throw 'protocol frame has missing, duplicate, reordered or extra fields' }
  if ($frame.schema -cne 'talenro-c12-prepared-frame/v1' -or $frame.run -cne $State.Run -or $frame.worker_nonce -cne $State.Nonce) { throw 'protocol frame schema/run/worker nonce mismatch' }
  if ($frame.sequence -isnot [int] -or $frame.sequence -ne ($State.Sequence + 1) -or $frame.type -cne $Type) { throw 'protocol frame sequence/type replay or order mismatch' }
  if ($frame.previous_frame_digest -cne $State.Previous) { throw 'protocol previous frame digest mismatch' }
  $expected = New-C12PreparedFrame -State $State -Type $Type -Payload $frame.payload
  $canonical = $utf8.GetString($expected) | ConvertFrom-Json
  if ($frame.payload_digest -cne $canonical.payload_digest) { throw 'protocol payload digest mismatch' }
  if (-not [C12PreparedPipe]::FixedEquals($frame.hmac_sha256, $canonical.hmac_sha256)) { throw 'protocol frame HMAC mismatch' }
  if ($text -cne $utf8.GetString($expected)) { throw 'protocol frame is not canonical JSON' }
  $digest = [C12PreparedPipe]::SHA256($Bytes)
  $State.Sequence++; $State.Previous = $digest
  $State.Frames.Add([pscustomobject]@{ Sequence = $State.Sequence; Type = $Type; Digest = $digest; PayloadDigest = $frame.payload_digest })
  return $frame.payload
}

function Read-C12PreparedMessage {
  param($Pipe, [DateTime]$Deadline)
  return ,([C12PreparedPipe]::ReadMessage($Pipe, (Get-C12ProtocolMilliseconds $Deadline)))
}

function Receive-C12PreparedFrame {
  param($State, $Pipe, [DateTime]$Deadline, [string]$Type)
  [byte[]]$bytes = Read-C12PreparedMessage -Pipe $Pipe -Deadline $Deadline
  return Assert-C12PreparedFrame -State $State -Bytes $bytes -Type $Type
}

function Send-C12PreparedFrame {
  param($State, $Pipe, [DateTime]$Deadline, [string]$Type, $Payload)
  [byte[]]$bytes = New-C12PreparedFrame -State $State -Type $Type -Payload $Payload
  $null = Assert-C12PreparedFrame -State $State -Bytes $bytes -Type $Type
  [C12PreparedPipe]::WriteMessage($Pipe, $bytes, (Get-C12ProtocolMilliseconds $Deadline))
}

$script:c12PreparedNativeWorkerScript = {
  param($Bootstrap)
  $ErrorActionPreference = 'Stop'
  Set-StrictMode -Version Latest
  Initialize-C12PreparedPipe
  $key = [Convert]::FromBase64String($env:C12_BOOTSTRAP_KEY)
  [Environment]::SetEnvironmentVariable('C12_BOOTSTRAP_KEY', $null)
  $state = New-C12ProtocolState -Run $Bootstrap.Run -Nonce $Bootstrap.Nonce -Key $key
  $pipe = [IO.Pipes.NamedPipeClientStream]::new('.', $Bootstrap.Pipe, [IO.Pipes.PipeDirection]::InOut, [IO.Pipes.PipeOptions]::Asynchronous)
  $gate = $null
  $seals = [Collections.Generic.List[IDisposable]]::new()
  try {
    # The bootstrap starts only when invoked. Its short initial connection
    # allowance must never be anchored to the earlier preparation timestamp.
    $deadline = [DateTime]::UtcNow.AddMinutes(2)
    $pipe.Connect((Get-C12ProtocolMilliseconds $deadline))
    $pipe.ReadMode = [IO.Pipes.PipeTransmissionMode]::Message
    [C12PreparedPipe]::VerifyServerProcessId($pipe, [uint32]$Bootstrap.ControllerPID)
    Send-C12PreparedFrame $state $pipe $deadline 'HELLO' ([ordered]@{ worker_pid = $PID; controller_pid = $Bootstrap.ControllerPID })
    $invocation = Receive-C12PreparedFrame $state $pipe $deadline 'VERIFICATION_PROJECTION'
    if ((@($invocation.PSObject.Properties.Name) -join '|') -cne 'invocation_deadline_ticks|verification_projection' -or
        $invocation.invocation_deadline_ticks -isnot [long] -or
        $invocation.invocation_deadline_ticks -le [DateTime]::UtcNow.Ticks -or
        $invocation.invocation_deadline_ticks -gt [DateTime]::MaxValue.Ticks) { throw 'protocol invocation deadline is malformed or expired' }
    $invocationDeadline = [DateTime]::new($invocation.invocation_deadline_ticks, [DateTimeKind]::Utc)
    $deadline = $invocationDeadline
    $projection = $invocation.verification_projection
    if ($projection.gate_name -cne $Bootstrap.Gate -or $projection.role -cnotin @('trusted-validator','authority-initializer-json')) { throw 'protocol projection identity mismatch' }
    $projectionDigest = [C12PreparedPipe]::SHA256([Text.Encoding]::UTF8.GetBytes((ConvertTo-C12ProtocolJSON $invocation)))
    # AssignProcessToJobObject was completed by the controller before bootstrap resume.
    # These immutable input fields are the worker-side Assert-C12SealedExecutableReceipt projection.
    foreach ($input in @($projection.inputs)) {
      $seals.Add([C12SealedExecutable]::OpenAndVerify($input.path, $input.sha256, [uint64]$input.length, [uint32]$input.volume, [uint64]$input.index, [uint32]$input.links, $input.owner, $input.dacl))
    }
    $gate = [Threading.EventWaitHandle]::OpenExisting($Bootstrap.Gate)
    Send-C12PreparedFrame $state $pipe $deadline 'JOB_MEMBER_READY' ([ordered]@{ projection_digest = $projectionDigest })
    $capabilities = Receive-C12PreparedFrame $state $pipe $deadline 'CAPABILITIES'
    $allowed = @('TALENRO_C12_AUTHORITY_V7_INIT_NONCE','TALENRO_C12_AUTHORITY_V7_RUN_SUFFIX','TALENRO_C12_AUTHORITY_V7_PROFILE','TALENRO_DATABASE_URL','TALENRO_INSTALLATION_KIND')
    $entries = @($capabilities.entries)
    if ($projection.role -ceq 'trusted-validator' -and $entries.Count -ne 0) { throw 'protocol validator capabilities forbidden' }
    if ($projection.role -ceq 'authority-initializer-json' -and ((@($entries | ForEach-Object { $_.name } | Sort-Object) -join '|') -cne (($allowed | Sort-Object) -join '|'))) { throw 'protocol capability allowlist mismatch' }
    foreach ($entry in $entries) {
      if ($entry.name -cnotin $allowed -or $entry.value -isnot [string] -or $entry.value.Length -gt 8192 -or $entry.value.IndexOf([char]0) -ge 0) { throw 'protocol capability entry invalid' }
      [Environment]::SetEnvironmentVariable($entry.name, $entry.value)
    }
    $capabilityDigest = [C12PreparedPipe]::SHA256([Text.Encoding]::UTF8.GetBytes((ConvertTo-C12ProtocolJSON $capabilities)))
    Send-C12PreparedFrame $state $pipe $deadline 'CAPABILITIES_INSTALLED' ([ordered]@{ capability_digest = $capabilityDigest })
    Send-C12PreparedFrame $state $pipe $deadline 'GATE_WAITING' ([ordered]@{ gate_name = $Bootstrap.Gate; capability_digest = $capabilityDigest })
    if (-not $gate.WaitOne((Get-C12ProtocolMilliseconds $deadline))) { throw 'protocol formal gate deadline exhausted' }
    # Results share the actual invocation ceiling; the authoritative controller
    # watchdog starts immediately before gate signal, not during preparation.
    $deadline = [DateTime]::UtcNow.AddMinutes(2)
    if ($invocationDeadline -lt $deadline) { $deadline = $invocationDeadline }
    # FORMAL_RELEASE is recorded by the controller before it signals this gate.
    $begin = [DateTime]::UtcNow.Ticks
    Send-C12PreparedFrame $state $pipe $deadline 'EXEC_BEGIN' ([ordered]@{ timestamp_ticks = $begin })
    $output = [Collections.Generic.List[string]]::new()
    $arguments = @($projection.formal_argv | ForEach-Object { [string]$_ })
    $ErrorActionPreference = 'Continue'
    & ([string]$projection.executable_path) @arguments 2>&1 | ForEach-Object {
      if ($output.Count -lt 4096) {
        $line = ([string]$_) -replace '[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]', '?'
        if ($line.Length -gt 512) { $line = $line.Substring(0,512) }
        $output.Add($line)
      }
    }
    $code = $LASTEXITCODE
    $ErrorActionPreference = 'Stop'
    # Result chunks use the same authenticated channel after the six-frame handshake.
    foreach ($line in $output) { Send-C12PreparedFrame $state $pipe $deadline 'OUTPUT' ([ordered]@{ line = $line }) }
    Send-C12PreparedFrame $state $pipe $deadline 'EXEC_END' ([ordered]@{ timestamp_ticks = [DateTime]::UtcNow.Ticks; exit_code = $code; output_count = $output.Count })
  }
  finally {
    foreach ($seal in $seals) { $seal.Dispose() }
    if ($null -ne $gate) { $gate.Dispose() }
    $pipe.Dispose()
    [Array]::Clear($key,0,$key.Length)
  }
}

function Initialize-C12SuspendedProcessController {
  if ('C12SuspendedProcessController' -as [type]) { return }
  Add-Type -TypeDefinition @'
using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Diagnostics;
using System.IO;
using System.Runtime.InteropServices;
using System.Text;

public sealed class C12SuspendedProcessController : IDisposable
{
    private const uint CREATE_SUSPENDED = 0x00000004;
    private const uint CREATE_NO_WINDOW = 0x08000000;
    private const uint CREATE_UNICODE_ENVIRONMENT = 0x00000400;
    private const uint JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE = 0x00002000;
    private const uint JOB_OBJECT_MSG_ACTIVE_PROCESS_ZERO = 4;
    private const uint WAIT_OBJECT_0 = 0, WAIT_TIMEOUT = 258;
    private IntPtr processHandle, primaryThreadHandle, jobHandle, completionPortHandle;
    private bool assigned, zeroMessageObserved, disposed;
    public uint ProcessId { get; private set; }
    public IntPtr ProcessHandle { get { return processHandle; } }
    public IntPtr PrimaryThreadHandle { get { return primaryThreadHandle; } }
    public IntPtr JobHandle { get { return jobHandle; } }
    public IntPtr CompletionPortHandle { get { return completionPortHandle; } }
    public string Phase { get; private set; }
    public bool ActiveProcessZeroConfirmed { get; private set; }
    public bool ProcessExitPendingOnly { get; private set; }

    [StructLayout(LayoutKind.Sequential, CharSet = CharSet.Unicode)]
    private struct STARTUPINFO {
        public uint cb; public string lpReserved, lpDesktop, lpTitle;
        public uint dwX, dwY, dwXSize, dwYSize, dwXCountChars, dwYCountChars, dwFillAttribute, dwFlags;
        public ushort wShowWindow, cbReserved2; public IntPtr lpReserved2, hStdInput, hStdOutput, hStdError;
    }
    [StructLayout(LayoutKind.Sequential)]
    private struct PROCESS_INFORMATION { public IntPtr hProcess, hThread; public uint dwProcessId, dwThreadId; }
    [StructLayout(LayoutKind.Sequential)]
    private struct JOBOBJECT_BASIC_LIMIT_INFORMATION {
        public long PerProcessUserTimeLimit, PerJobUserTimeLimit; public uint LimitFlags;
        public UIntPtr MinimumWorkingSetSize, MaximumWorkingSetSize; public uint ActiveProcessLimit;
        public UIntPtr Affinity; public uint PriorityClass, SchedulingClass;
    }
    [StructLayout(LayoutKind.Sequential)]
    private struct IO_COUNTERS { public ulong ReadOperationCount, WriteOperationCount, OtherOperationCount, ReadTransferCount, WriteTransferCount, OtherTransferCount; }
    [StructLayout(LayoutKind.Sequential)]
    private struct JOBOBJECT_EXTENDED_LIMIT_INFORMATION {
        public JOBOBJECT_BASIC_LIMIT_INFORMATION BasicLimitInformation; public IO_COUNTERS IoInfo;
        public UIntPtr ProcessMemoryLimit, JobMemoryLimit, PeakProcessMemoryUsed, PeakJobMemoryUsed;
    }
    [StructLayout(LayoutKind.Sequential)]
    private struct JOBOBJECT_ASSOCIATE_COMPLETION_PORT { public IntPtr CompletionKey, CompletionPort; }
    [StructLayout(LayoutKind.Sequential)]
    private struct JOBOBJECT_BASIC_ACCOUNTING_INFORMATION {
        public long TotalUserTime, TotalKernelTime, ThisPeriodTotalUserTime, ThisPeriodTotalKernelTime;
        public uint TotalPageFaultCount, TotalProcesses, ActiveProcesses, TotalTerminatedProcesses;
    }

    [DllImport("kernel32.dll", CharSet = CharSet.Unicode, ExactSpelling = true, SetLastError = true)]
    private static extern bool CreateProcessW(string application, StringBuilder commandLine, IntPtr processAttributes, IntPtr threadAttributes, bool bInheritHandles, uint flags, IntPtr environment, string workingDirectory, ref STARTUPINFO startup, out PROCESS_INFORMATION process);
    [DllImport("kernel32.dll", CharSet = CharSet.Unicode, ExactSpelling = true, SetLastError = true)]
    private static extern IntPtr CreateJobObjectW(IntPtr attributes, string name);
    [DllImport("kernel32.dll", SetLastError = true)]
    private static extern bool SetInformationJobObject(IntPtr job, int informationClass, IntPtr information, uint length);
    [DllImport("kernel32.dll", SetLastError = true)]
    private static extern bool AssignProcessToJobObject(IntPtr job, IntPtr process);
    [DllImport("kernel32.dll", SetLastError = true)]
    private static extern bool IsProcessInJob(IntPtr process, IntPtr job, out bool member);
    [DllImport("kernel32.dll", SetLastError = true)]
    private static extern bool QueryInformationJobObject(IntPtr job, int informationClass, out JOBOBJECT_BASIC_ACCOUNTING_INFORMATION information, uint length, out uint returned);
    [DllImport("kernel32.dll", SetLastError = true)]
    private static extern IntPtr CreateIoCompletionPort(IntPtr file, IntPtr existingPort, UIntPtr key, uint concurrentThreads);
    [DllImport("kernel32.dll", SetLastError = true)]
    private static extern bool GetQueuedCompletionStatus(IntPtr port, out uint message, out UIntPtr key, out IntPtr overlapped, uint milliseconds);
    [DllImport("kernel32.dll", SetLastError = true)] private static extern uint ResumeThread(IntPtr thread);
    [DllImport("kernel32.dll", SetLastError = true)] private static extern bool TerminateProcess(IntPtr process, uint exitCode);
    [DllImport("kernel32.dll", SetLastError = true)] private static extern bool TerminateJobObject(IntPtr job, uint exitCode);
    [DllImport("kernel32.dll", SetLastError = true)] private static extern bool CloseHandle(IntPtr handle);
    [DllImport("kernel32.dll", SetLastError = true)] private static extern uint GetProcessId(IntPtr process);
    [DllImport("kernel32.dll", SetLastError = true)] private static extern uint GetProcessIdOfThread(IntPtr thread);
    [DllImport("kernel32.dll", SetLastError = true)] private static extern uint WaitForSingleObject(IntPtr handle, uint milliseconds);

    private C12SuspendedProcessController() { }
    private static bool Invalid(IntPtr handle) { return handle == IntPtr.Zero || handle == new IntPtr(-1); }
    private static Win32Exception NativeError(string operation) {
        int code = Marshal.GetLastWin32Error();
        return new Win32Exception(code, operation + " failed (Win32 " + code + ")");
    }
    private static void RequireHandle(IntPtr handle, string name) { if (Invalid(handle)) throw new InvalidOperationException(name + " is zero or invalid"); }
    private void RequireOpen() {
        RequireCleanupHandles();
        RequireHandle(primaryThreadHandle, "primary thread handle");
    }
    private void RequireCleanupHandles() {
        if (disposed) throw new ObjectDisposedException("C12SuspendedProcessController");
        RequireHandle(processHandle, "process handle");
        RequireHandle(jobHandle, "Job handle"); RequireHandle(completionPortHandle, "completion port handle");
    }
    private void VerifyProcessIdentity() {
        RequireOpen();
        if (ProcessId == 0 || GetProcessId(processHandle) != ProcessId || GetProcessIdOfThread(primaryThreadHandle) != ProcessId)
            throw new InvalidOperationException("suspended worker PID/handle mismatch");
    }
    private void SetJobInformation(int kind, object information) {
        int length = Marshal.SizeOf(information);
        IntPtr memory = Marshal.AllocHGlobal(length);
        try {
            Marshal.StructureToPtr(information, memory, false);
            if (!SetInformationJobObject(jobHandle, kind, memory, (uint)length)) throw NativeError("SetInformationJobObject");
        } finally { Marshal.FreeHGlobal(memory); }
    }
    private static string EnvironmentBlock(IDictionary<string,string> environment) {
        if (environment == null) throw new ArgumentNullException("environment");
        var ordered = new SortedDictionary<string,string>(StringComparer.OrdinalIgnoreCase);
        foreach (var entry in environment) {
            if (String.IsNullOrEmpty(entry.Key) || entry.Key.IndexOf('\0') >= 0 || entry.Key.IndexOf('=') >= 0 || entry.Value == null || entry.Value.IndexOf('\0') >= 0)
                throw new ArgumentException("invalid explicit worker environment");
            ordered.Add(entry.Key, entry.Value);
        }
        var block = new StringBuilder();
        foreach (var entry in ordered) block.Append(entry.Key).Append('=').Append(entry.Value).Append('\0');
        block.Append('\0');
        if (ordered.Count == 0) block.Append('\0');
        return block.ToString();
    }
    public static C12SuspendedProcessController Create(string executable, string arguments, string workingDirectory, IDictionary<string,string> environment) {
        return Create(executable, arguments, workingDirectory, environment, null);
    }
    public static C12SuspendedProcessController Create(string executable, string arguments, string workingDirectory, IDictionary<string,string> environment, string jobName) {
        if (String.IsNullOrEmpty(executable) || !Path.IsPathRooted(executable) || executable.IndexOf('"') >= 0 || executable.IndexOf('\0') >= 0)
            throw new ArgumentException("worker executable must be an exact absolute path");
        if (String.IsNullOrEmpty(workingDirectory) || !Path.IsPathRooted(workingDirectory) || workingDirectory.IndexOf('\0') >= 0)
            throw new ArgumentException("worker working directory must be absolute");
        if (arguments == null || arguments.IndexOf('\0') >= 0) throw new ArgumentException("invalid worker arguments");
        var commandLine = new StringBuilder("\"" + executable + "\" " + arguments);
        if (commandLine.Length >= 32767) throw new ArgumentException("worker command line is too long");
        string block = EnvironmentBlock(environment);
        var controller = new C12SuspendedProcessController();
        IntPtr nativeEnvironment = IntPtr.Zero;
        try {
            controller.jobHandle = CreateJobObjectW(IntPtr.Zero, jobName);
            int jobError = Marshal.GetLastWin32Error();
            RequireHandle(controller.jobHandle, "created Job handle");
            if (jobError == 183) throw new InvalidOperationException("named native Job collision");
            var limits = new JOBOBJECT_EXTENDED_LIMIT_INFORMATION();
            limits.BasicLimitInformation.LimitFlags = JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE;
            controller.SetJobInformation(9, limits);
            controller.completionPortHandle = CreateIoCompletionPort(new IntPtr(-1), IntPtr.Zero, UIntPtr.Zero, 1);
            RequireHandle(controller.completionPortHandle, "created completion port handle");
            controller.SetJobInformation(7, new JOBOBJECT_ASSOCIATE_COMPLETION_PORT { CompletionKey = new IntPtr(1), CompletionPort = controller.completionPortHandle });
            nativeEnvironment = Marshal.StringToHGlobalUni(block);
            var startup = new STARTUPINFO();
            startup.cb = (uint)Marshal.SizeOf(typeof(STARTUPINFO));
            PROCESS_INFORMATION process;
            if (!CreateProcessW(executable, commandLine, IntPtr.Zero, IntPtr.Zero, bInheritHandles: false,
                    flags: CREATE_SUSPENDED | CREATE_NO_WINDOW | CREATE_UNICODE_ENVIRONMENT,
                    environment: nativeEnvironment, workingDirectory: workingDirectory, startup: ref startup, process: out process))
                throw NativeError("CreateProcessW");
            controller.processHandle = process.hProcess; controller.primaryThreadHandle = process.hThread;
            controller.ProcessId = process.dwProcessId; controller.Phase = "CreatedSuspended";
            controller.VerifyProcessIdentity();
            return controller;
        } catch (Exception error) {
            if (!Invalid(controller.processHandle)) {
                // Ownership crosses the exception boundary; PowerShell registers it
                // before cleanup and retains it if bounded convergence cannot finish.
                error.Data["C12SuspendedProcessController"] = controller;
            } else {
                CloseOwnedHandle(ref controller.primaryThreadHandle);
                CloseOwnedHandle(ref controller.jobHandle);
                CloseOwnedHandle(ref controller.completionPortHandle);
            }
            throw;
        } finally { if (nativeEnvironment != IntPtr.Zero) Marshal.FreeHGlobal(nativeEnvironment); }
    }
    public void AssignToJob() {
        VerifyProcessIdentity();
        if (Phase != "CreatedSuspended") throw new InvalidOperationException("Job assignment requires CreatedSuspended");
        if (!AssignProcessToJobObject(jobHandle, processHandle)) throw NativeError("AssignProcessToJobObject");
        assigned = true; Phase = "AssignedToJob";
    }
    public void VerifyMembership() {
        VerifyProcessIdentity();
        if (Phase != "AssignedToJob" && Phase != "MembershipVerified") throw new InvalidOperationException("membership verification requires AssignedToJob");
        bool member;
        if (!IsProcessInJob(processHandle, jobHandle, out member)) throw NativeError("IsProcessInJob");
        if (!member) throw new InvalidOperationException("worker is not in the exact controller Job");
        Phase = "MembershipVerified";
    }
    public void Resume() {
        RequireOpen();
        if (Phase != "MembershipVerified") throw new InvalidOperationException("resume requires MembershipVerified exactly once");
        VerifyMembership();
        uint previous = ResumeThread(primaryThreadHandle);
        if (previous == UInt32.MaxValue) throw NativeError("ResumeThread");
        // Mark the transition even for an unexpected suspension count: never retry release.
        Phase = "Released";
        if (previous != 1) throw new InvalidOperationException("primary thread suspension count changed");
    }
    private void TerminateOwnedProcess(uint exitCode) {
        if (TerminateProcess(processHandle, exitCode)) return;
        int error = Marshal.GetLastWin32Error();
        // Windows returns ACCESS_DENIED when natural exit wins the interval
        // between our outer wait and TerminateProcess. Only the exact retained
        // process handle can prove this benign race; every ambiguous error stays
        // fatal. The caller still terminates the Job and confirms active zero.
        if (error == 5 && WaitForSingleObject(processHandle, 0) == WAIT_OBJECT_0) return;
        throw new Win32Exception(error, "TerminateProcess failed (Win32 " + error + ")");
    }
    public void Terminate(uint exitCode) {
        RequireCleanupHandles(); Phase = "CleanIntent";
        ProcessExitPendingOnly = false;
        var failures = new List<Exception>();
        bool processAccessDenied = false;
        if (WaitForSingleObject(processHandle, 0) != WAIT_OBJECT_0) {
            try { TerminateOwnedProcess(exitCode); }
            catch (Exception error) {
                failures.Add(error);
                var native = error as Win32Exception;
                processAccessDenied = native != null && native.NativeErrorCode == 5;
            }
        }
        if (!TerminateJobObject(jobHandle, exitCode)) failures.Add(NativeError("TerminateJobObject"));
        // Exit can still be pending on the immediate recheck. This classification
        // is not success: only the caller's bounded exact exit + Job-zero wait
        // may resolve the sole ACCESS_DENIED error. All other errors stay fatal.
        ProcessExitPendingOnly = processAccessDenied && failures.Count == 1;
        if (failures.Count != 0) throw new AggregateException("suspended worker termination failed", failures);
    }
    public uint ActiveProcesses {
        get {
            RequireCleanupHandles(); JOBOBJECT_BASIC_ACCOUNTING_INFORMATION information; uint returned;
            if (!QueryInformationJobObject(jobHandle, 1, out information, (uint)Marshal.SizeOf(typeof(JOBOBJECT_BASIC_ACCOUNTING_INFORMATION)), out returned)) throw NativeError("QueryInformationJobObject");
            if (returned != Marshal.SizeOf(typeof(JOBOBJECT_BASIC_ACCOUNTING_INFORMATION))) throw new InvalidOperationException("Job accounting length mismatch");
            return information.ActiveProcesses;
        }
    }
    public bool WaitForActiveProcessZero(int timeoutMilliseconds) {
        RequireCleanupHandles();
        if (timeoutMilliseconds < 0) throw new ArgumentOutOfRangeException("timeoutMilliseconds");
        var watch = Stopwatch.StartNew();
        do {
            uint processState = WaitForSingleObject(processHandle, 0);
            if (processState != WAIT_OBJECT_0 && processState != WAIT_TIMEOUT) throw NativeError("WaitForSingleObject");
            // An assignment rejected by Windows never entered the Job and therefore
            // cannot emit ACTIVE_PROCESS_ZERO. Confirm its handle exit plus empty Job.
            if ((!assigned || zeroMessageObserved) && ActiveProcesses == 0 && processState == WAIT_OBJECT_0) {
                ActiveProcessZeroConfirmed = true; return true;
            }
            long remaining = timeoutMilliseconds - watch.ElapsedMilliseconds;
            if (remaining <= 0) return false;
            uint message; UIntPtr key; IntPtr overlapped;
            if (GetQueuedCompletionStatus(completionPortHandle, out message, out key, out overlapped, (uint)Math.Min(remaining, 50))) {
                if (key != new UIntPtr(1)) throw new InvalidOperationException("foreign Job completion key");
                if (message == JOB_OBJECT_MSG_ACTIVE_PROCESS_ZERO) zeroMessageObserved = true;
            } else if (Marshal.GetLastWin32Error() != WAIT_TIMEOUT) { throw NativeError("GetQueuedCompletionStatus"); }
        } while (true);
    }
    private static void CloseOwnedHandle(ref IntPtr handle) {
        if (Invalid(handle)) return;
        if (!CloseHandle(handle)) throw NativeError("CloseHandle");
        handle = IntPtr.Zero;
    }
    public void Dispose() {
        if (disposed) return;
        if (!ActiveProcessZeroConfirmed) throw new InvalidOperationException("native handles retained until process exit and active-process-zero confirmation");
        CloseOwnedHandle(ref primaryThreadHandle); CloseOwnedHandle(ref processHandle);
        CloseOwnedHandle(ref jobHandle); CloseOwnedHandle(ref completionPortHandle);
        disposed = true; Phase = "Removed";
    }
}
'@
}

function New-C12SuspendedProcessController {
  param([string]$Executable, [string]$Arguments, [string]$WorkingDirectory,
    [System.Collections.Generic.IDictionary[string,string]]$Environment, [string]$JobName)
  Initialize-C12SuspendedProcessController
  return [C12SuspendedProcessController]::Create($Executable, $Arguments, $WorkingDirectory, $Environment, $JobName)
}

function New-C12SuspendedPreparedWorker {
  param([string]$Executable, [string]$Arguments, [string]$WorkingDirectory,
    [System.Collections.Generic.IDictionary[string,string]]$Environment,
    [object]$ArtifactRoot, [object]$Receipt, [DateTime]$SetupDeadline)

  $controller = $null
  $prepared = $null
  $creationFailure = $null
  try {
    $controller = New-C12SuspendedProcessController -Executable $Executable -Arguments $Arguments -WorkingDirectory $WorkingDirectory -Environment $Environment -JobName ([string]$Receipt.NativeJobName)
  }
  catch {
    $creationFailure = $_
    $exception = $_.Exception
    while ($null -ne $exception -and $null -eq $controller) {
      $controller = $exception.Data['C12SuspendedProcessController']
      $exception = $exception.InnerException
    }
    if ($null -eq $controller) { throw }
  }
  # Register ownership before the first operation that can fail after creation.
  $phases = New-Object 'System.Collections.Generic.List[string]'
  foreach ($phase in @('SETUP_BEGIN', 'BUILD_DONE', 'ARTIFACT_BOUND')) { $phases.Add($phase) }
  $prepared = [pscustomobject]@{
    Controller = $controller; ProcessId = $controller.ProcessId
    ProcessHandle = $controller.ProcessHandle; PrimaryThreadHandle = $controller.PrimaryThreadHandle
    JobHandle = $controller.JobHandle; NativeJobHandle = $controller.JobHandle
    CompletionPortHandle = $controller.CompletionPortHandle; Phase = 'CreatedSuspended'
    Role = [string]$Receipt.Role; Receipt = $Receipt; ArtifactRoot = $ArtifactRoot
    Job = $null; Gate = $null; GateSignaled = $false; Released = $false; Closed = $false; ExecutionComplete = $false
    Protocol = $null; Projection = $null; Pipe = $null; ProtocolKey = $null; EnvironmentDigest = ''
    ReleaseTicks = [long]0; ReleaseWatch = $null; ProtocolReceipts = @(); InvocationDeadline = [DateTime]::MinValue
    ActiveProcessZeroConfirmed = $false; CleanupFailures = @(); Phases = $phases
  }
  $script:c12PreparedWorkers.Add($prepared)
  try {
    if ($null -ne $creationFailure) { throw $creationFailure }
    if ([DateTime]::UtcNow -ge $SetupDeadline) { throw 'prepared worker setup deadline is exhausted' }
    $controller.AssignToJob()
    $prepared.Phase = 'AssignedToJob'
    $controller.VerifyMembership()
    $prepared.Phase = 'MembershipVerified'
    if ([DateTime]::UtcNow -ge $SetupDeadline) { throw 'prepared worker setup deadline is exhausted' }
    # Task 3 owns authenticated bootstrap and release. No thread resumes here.
    return $prepared
  }
  catch {
    $primaryFailure = $_
    try { Close-C12PreparedNativeWorker -Prepared $prepared -Deadline ([DateTime]::UtcNow.AddSeconds(10)) }
    catch { $prepared.CleanupFailures += $_.Exception.Message }
    throw $primaryFailure
  }
}

function New-C12PreparedNativeWorker {
  param(
    [Parameter(Mandatory = $true)][object]$ArtifactRoot,
    [Parameter(Mandatory = $true)][object]$Receipt,
    [Parameter(Mandatory = $true)][DateTime]$SetupDeadline
  )

  if ([DateTime]::UtcNow -ge $SetupDeadline) { throw 'prepared worker setup deadline is exhausted' }
  $verified = Assert-C12SealedExecutableReceipt -ArtifactRoot $ArtifactRoot -Receipt $Receipt
  try { if ($null -ne $verified.Secondary) { $verified.Secondary.Dispose() }; $verified.Primary.Dispose() }
  catch { throw }
  $environment = [Collections.Generic.Dictionary[string,string]]::new([StringComparer]::OrdinalIgnoreCase)
  # Only ordinary Windows process setup crosses into the bootstrap. Controller
  # receipt/WAL keys and arbitrary inherited capabilities are never propagated.
  foreach ($name in @('SystemRoot','WINDIR','TEMP','TMP','PATH','PATHEXT','COMSPEC','USERPROFILE','LOCALAPPDATA','APPDATA','PROCESSOR_ARCHITECTURE','NUMBER_OF_PROCESSORS')) {
    $value = [Environment]::GetEnvironmentVariable($name)
    if ($null -ne $value) { $environment[$name] = $value }
  }
  Initialize-C12PreparedPipe
  $environmentProjection = @($environment.Keys | Sort-Object | ForEach-Object { [ordered]@{ name = $_; value = $environment[$_] } })
  $environmentDigest = [C12PreparedPipe]::SHA256([Text.Encoding]::UTF8.GetBytes((ConvertTo-C12ProtocolJSON $environmentProjection)))
  $key = New-Object byte[] 32
  $rng = [Security.Cryptography.RandomNumberGenerator]::Create()
  try { $rng.GetBytes($key) } finally { $rng.Dispose() }
  $environment['C12_BOOTSTRAP_KEY'] = [Convert]::ToBase64String($key)
  $bootstrap = [ordered]@{ Run = $Receipt.RunSuffix; Nonce = $Receipt.OneShotNonce; Pipe = $Receipt.PipeName; Gate = $Receipt.GateName; ControllerPID = $PID }
  $source = '$ErrorActionPreference = ''Stop'';' + [Environment]::NewLine
  foreach ($function in @('Initialize-C12PreparedPipe','Get-C12ProtocolMilliseconds','ConvertTo-C12ProtocolValue','ConvertTo-C12ProtocolJSON','New-C12ProtocolState','New-C12PreparedFrame','Assert-C12PreparedFrame','Read-C12PreparedMessage','Receive-C12PreparedFrame','Send-C12PreparedFrame')) {
    $source += 'function ' + $function + ' {' + (Get-Command $function).ScriptBlock.ToString() + '}' + [Environment]::NewLine
  }
  $source += "Add-Type -TypeDefinition @'" + [Environment]::NewLine + $script:c12SealedExecutableSource + [Environment]::NewLine + "'@" + [Environment]::NewLine
  $bootstrapJSON = [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes((ConvertTo-C12ProtocolJSON $bootstrap)))
  $source += '& {' + $script:c12PreparedNativeWorkerScript.ToString() + '} ([Text.Encoding]::UTF8.GetString([Convert]::FromBase64String(''' + $bootstrapJSON + ''')) | ConvertFrom-Json)'
  $tokens = $null; $parseErrors = $null
  $null = [Management.Automation.Language.Parser]::ParseInput($source, [ref]$tokens, [ref]$parseErrors)
  if ($parseErrors.Count -ne 0) { throw ('protocol bootstrap syntax invalid: ' + ($parseErrors -join '; ')) }
  $compressed = [IO.MemoryStream]::new()
  $gzip = [IO.Compression.GZipStream]::new($compressed, [IO.Compression.CompressionMode]::Compress, $true)
  try { $bytes = [Text.Encoding]::UTF8.GetBytes($source); $gzip.Write($bytes,0,$bytes.Length) } finally { $gzip.Dispose() }
  try { $encoded = [Convert]::ToBase64String($compressed.ToArray()) } finally { $compressed.Dispose() }
  $loader = '$m=[IO.MemoryStream]::new([Convert]::FromBase64String(''' + $encoded + '''));$g=[IO.Compression.GZipStream]::new($m,[IO.Compression.CompressionMode]::Decompress);$r=[IO.StreamReader]::new($g);try{& ([scriptblock]::Create($r.ReadToEnd()))}finally{$r.Dispose();$g.Dispose();$m.Dispose()}'
  $encodedArguments = '-NoProfile -NonInteractive -ExecutionPolicy Bypass -EncodedCommand ' + [Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($loader))
  $bootstrapExecutable = Join-Path $env:SystemRoot 'System32\WindowsPowerShell\v1.0\powershell.exe'
  $prepared = $null
  try {
    $prepared = New-C12SuspendedPreparedWorker -Executable $bootstrapExecutable -Arguments $encodedArguments -WorkingDirectory ([string]$Receipt.WorkingDirectory) -Environment $environment -ArtifactRoot $ArtifactRoot -Receipt $Receipt -SetupDeadline $SetupDeadline
    $prepared.ProtocolKey = $key
    $prepared.EnvironmentDigest = $environmentDigest
    $created = $false
    $prepared.Gate = [Threading.EventWaitHandle]::new($false, [Threading.EventResetMode]::ManualReset, [string]$Receipt.GateName, [ref]$created)
    if (-not $created) { throw 'protocol gate collision' }
    $prepared.Projection = New-C12VerificationProjection -Prepared $prepared
    return $prepared
  }
  catch {
    $primaryFailure = $_
    if ($null -ne $prepared) {
      try { Close-C12PreparedNativeWorker -Prepared $prepared -Deadline ([DateTime]::UtcNow.AddSeconds(10)) }
      catch { $prepared.CleanupFailures += $_.Exception.Message }
    }
    else { [Array]::Clear($key,0,$key.Length) }
    throw $primaryFailure
  }
}

function New-C12VerificationProjection {
  param($Prepared)
  $record = $Prepared.Receipt
  $inputs = @()
  foreach ($prefix in @('','Secondary')) {
    $path = [string]$record.($prefix + 'ExecutablePath')
    if ([string]::IsNullOrEmpty($path)) { continue }
    $inputs += [ordered]@{
      path = $path; sha256 = $record.($prefix + 'ExecutableSHA256'); length = $record.($prefix + 'ExecutableLength')
      volume = $record.($prefix + 'ExecutableVolumeSerial'); index = $record.($prefix + 'ExecutableFileIndex'); links = $record.($prefix + 'ExecutableLinkCount')
      owner = $record.($prefix + 'ExecutableOwner'); dacl = $record.($prefix + 'ExecutableDACL')
    }
  }
  return [ordered]@{
    role = $record.Role; gate_name = $record.GateName; executable_path = $record.ExecutablePath
    formal_argv = @($record.Arguments); argv_digest = $record.ArgumentsDigest; inputs = $inputs
    source_digest = $record.SourceDigest; candidate_tree_digest = $record.CandidateTreeIdentity; environment_digest = $Prepared.EnvironmentDigest
  }
}

function Get-C12PreparedWorkerPhases {
  param([Parameter(Mandatory = $true)][object]$Prepared)
  return [string[]]@($Prepared.Phases)
}

function Close-C12PreparedNativeWorker {
  param(
    [Parameter(Mandatory = $true)][object]$Prepared,
    [DateTime]$Deadline = [DateTime]::MaxValue
  )

  if ([bool]$Prepared.Closed) { return }
  $Prepared.Phase = 'CleanIntent'
  $terminationFailure = $null
  try { $Prepared.Controller.Terminate([uint32]1) }
  catch { $terminationFailure = $_; $Prepared.CleanupFailures += $_.Exception.Message }
  $remainingMilliseconds = [Math]::Floor(($Deadline - [DateTime]::UtcNow).TotalMilliseconds)
  if ($remainingMilliseconds -le 0) { throw 'prepared worker cleanup confirmation deadline is exhausted; native handles retained' }
  $waitMilliseconds = [int][Math]::Min(10000, $remainingMilliseconds)
  if (-not $Prepared.Controller.WaitForActiveProcessZero($waitMilliseconds)) { throw 'prepared worker active-process-zero confirmation timed out; native handles retained' }
  $Prepared.ActiveProcessZeroConfirmed = $true
  $Prepared.Controller.Dispose()
  $Prepared.ProcessHandle = [IntPtr]::Zero
  $Prepared.PrimaryThreadHandle = [IntPtr]::Zero
  $Prepared.JobHandle = [IntPtr]::Zero
  $Prepared.NativeJobHandle = [IntPtr]::Zero
  $Prepared.CompletionPortHandle = [IntPtr]::Zero
  if ($null -ne $Prepared.Gate) { $Prepared.Gate.Dispose(); $Prepared.Gate = $null }
  if ($null -ne $Prepared.Pipe) { $Prepared.Pipe.Dispose(); $Prepared.Pipe = $null }
  if ($null -ne $Prepared.ProtocolKey) { [Array]::Clear($Prepared.ProtocolKey,0,$Prepared.ProtocolKey.Length); $Prepared.ProtocolKey = $null }
  $Prepared.Phase = 'Removed'
  $Prepared.Closed = $true
  if ($null -ne $terminationFailure -and -not ($Prepared.ActiveProcessZeroConfirmed -and $Prepared.Controller.ProcessExitPendingOnly)) { throw $terminationFailure }
}

function Assert-C12PreparedClientPID {
  param($Prepared, $Pipe)
  [C12PreparedPipe]::VerifyClientProcessId($Pipe, [uint32]$Prepared.ProcessId)
}

function Invoke-C12PreparedProtocol {
  param($Prepared, [hashtable]$CapabilityEnvironment, [DateTime]$Deadline)
  if ($Prepared.Phase -cne 'MembershipVerified' -or $Prepared.Released -or $Prepared.Closed -or $null -ne $Prepared.Protocol) { throw 'protocol requires a fresh contained bootstrap' }
  Initialize-C12PreparedPipe
  $Prepared.Pipe = [C12PreparedPipe]::CreatePrivatePipe([string]$Prepared.Receipt.PipeName)
  $state = New-C12ProtocolState -Run $Prepared.Receipt.RunSuffix -Nonce $Prepared.Receipt.OneShotNonce -Key $Prepared.ProtocolKey
  $Prepared.Protocol = $state
  # Only the fixed trusted bootstrap resumes here. No arbitrary payload has
  # started; it can only be invoked beyond the authenticated formal gate.
  $Prepared.Controller.Resume()
  [C12PreparedPipe]::Connect($Prepared.Pipe, (Get-C12ProtocolMilliseconds $Deadline))
  Assert-C12PreparedClientPID -Prepared $Prepared -Pipe $Prepared.Pipe
  $hello = Receive-C12PreparedFrame $state $Prepared.Pipe $Deadline 'HELLO'
  if ($hello.worker_pid -ne $Prepared.ProcessId -or $hello.controller_pid -ne $PID) { throw 'protocol HELLO PID mismatch' }
  $Prepared.Phases.Add('HELLO')
  $Prepared.InvocationDeadline = $Deadline
  $invocation = [ordered]@{ invocation_deadline_ticks = $Deadline.Ticks; verification_projection = $Prepared.Projection }
  Send-C12PreparedFrame $state $Prepared.Pipe $Deadline 'VERIFICATION_PROJECTION' $invocation
  $Prepared.Phases.Add('VERIFICATION_PROJECTION')
  $projectionDigest = [C12PreparedPipe]::SHA256([Text.Encoding]::UTF8.GetBytes((ConvertTo-C12ProtocolJSON $invocation)))
  $ready = Receive-C12PreparedFrame $state $Prepared.Pipe $Deadline 'JOB_MEMBER_READY'
  if ($ready.projection_digest -cne $projectionDigest) { throw 'protocol projection digest mismatch' }
  $Prepared.Phases.Add('JOB_MEMBER_READY')
  $Prepared.Phase = 'PipeAuthenticated'
  $entries = @($CapabilityEnvironment.Keys | Sort-Object | ForEach-Object { [ordered]@{ name = [string]$_; value = [string]$CapabilityEnvironment[$_] } })
  $capabilities = [ordered]@{ entries = $entries }
  Send-C12PreparedFrame $state $Prepared.Pipe $Deadline 'CAPABILITIES' $capabilities
  $Prepared.Phases.Add('CAPABILITIES')
  $digest = [C12PreparedPipe]::SHA256([Text.Encoding]::UTF8.GetBytes((ConvertTo-C12ProtocolJSON $capabilities)))
  $installed = Receive-C12PreparedFrame $state $Prepared.Pipe $Deadline 'CAPABILITIES_INSTALLED'
  if ($installed.capability_digest -cne $digest) { throw 'protocol installed capability digest mismatch' }
  $Prepared.Phases.Add('CAPABILITIES_INSTALLED')
  $Prepared.Phase = 'CapabilitiesInstalled'
  $waiting = Receive-C12PreparedFrame $state $Prepared.Pipe $Deadline 'GATE_WAITING'
  if ($waiting.gate_name -cne $Prepared.Receipt.GateName -or $waiting.capability_digest -cne $digest) { throw 'protocol gate binding mismatch' }
  $Prepared.Phases.Add('GATE_WAITING')
  $Prepared.Phase = 'GateWaiting'
  $Prepared.ProtocolReceipts = @($state.Frames.ToArray())
  return [pscustomobject]@{ Frames = $Prepared.ProtocolReceipts; ProjectionDigest = $projectionDigest }
}

function Release-C12PreparedWorker {
  param($Prepared)
  if ($Prepared.Closed -or $Prepared.Released -or $Prepared.Phase -cne 'GateWaiting' -or $Prepared.Protocol.Sequence -ne 6 -or $Prepared.ProtocolReceipts.Count -ne 6) { throw 'protocol formal release requires exactly six verified frames' }
  if ([DateTime]::UtcNow -ge $Prepared.InvocationDeadline) { throw 'protocol formal release deadline exhausted' }
  $Prepared.Released = $true # exact-once guard survives all release failures.
  $Prepared.ReleaseWatch = [Diagnostics.Stopwatch]::StartNew()
  $Prepared.ReleaseTicks = [DateTime]::UtcNow.Ticks
  $Prepared.Phases.Add('FORMAL_RELEASE')
  if (-not $Prepared.Gate.Set()) { throw 'protocol formal gate signal failed' }
  $Prepared.GateSignaled = $true
  $Prepared.Phase = 'Released'
}

function Invoke-C12PreparedAuthorityRole {
  param(
    [Parameter(Mandatory = $true)][object]$Prepared,
    [Parameter(Mandatory = $true)][ValidateSet('trusted-validator', 'authority-initializer-json')][string]$Role,
    [Parameter(Mandatory = $true)][TimeSpan]$Timeout,
    [DateTime]$Deadline = [DateTime]::MaxValue,
    [hashtable]$CapabilityEnvironment = @{}
  )
  if ([string]$Prepared.Role -cne $Role -or [string]$Prepared.Receipt.Role -cne $Role) { throw 'prepared invocation has the wrong role' }
  if ([bool]$Prepared.Released) { throw 'prepared receipt replay was rejected' }
  if ($Prepared.Closed -or $null -eq $Prepared.Gate) { throw 'prepared invocation is already closed' }
  if ($Timeout -ne [TimeSpan]::FromMinutes(2)) { throw 'formal authority watchdog must remain exactly two minutes' }
  if ($Deadline -le [DateTime]::UtcNow) { throw 'formal authority stage exceeded its absolute deadline' }
  $allowed = @('TALENRO_C12_AUTHORITY_V7_INIT_NONCE','TALENRO_C12_AUTHORITY_V7_RUN_SUFFIX','TALENRO_C12_AUTHORITY_V7_PROFILE','TALENRO_DATABASE_URL','TALENRO_INSTALLATION_KIND')
  if ($Role -ceq 'trusted-validator' -and $CapabilityEnvironment.Count -ne 0) { throw 'trusted validator cannot receive authority capabilities' }
  if ($Role -ceq 'authority-initializer-json' -and ((@($CapabilityEnvironment.Keys | Sort-Object) -join '|') -cne (($allowed | Sort-Object) -join '|'))) { throw 'authority initializer capability allowlist is incomplete or excessive' }
  foreach ($value in $CapabilityEnvironment.Values) { if ($value -isnot [string] -or $value.Length -gt 8192 -or $value.IndexOf([char]0) -ge 0) { throw 'protocol capability value invalid' } }
  $verified = Assert-C12SealedExecutableReceipt -ArtifactRoot $Prepared.ArtifactRoot -Receipt $Prepared.Receipt
  $primaryFailure = $null
  try {
    $expected = New-C12VerificationProjection -Prepared $Prepared
    if ((ConvertTo-C12ProtocolJSON $expected) -cne (ConvertTo-C12ProtocolJSON $Prepared.Projection)) { throw 'protocol immutable projection changed' }
    $null = Invoke-C12PreparedProtocol -Prepared $Prepared -CapabilityEnvironment $CapabilityEnvironment -Deadline $Deadline
    foreach ($graph in @($Prepared.Receipt.GoGraphReceipts)) {
      Assert-C12GoGraphReceipt -Receipt $graph -ArtifactRoot $Prepared.ArtifactRoot -Deadline $Deadline
    }
    Release-C12PreparedWorker -Prepared $Prepared
    $releaseDeadline = [DateTime]::UtcNow.Add($Timeout - $Prepared.ReleaseWatch.Elapsed)
    if ($releaseDeadline -gt $Deadline) { $releaseDeadline = $Deadline }
    $begin = Receive-C12PreparedFrame $Prepared.Protocol $Prepared.Pipe $releaseDeadline 'EXEC_BEGIN'
    if ([long]$begin.timestamp_ticks -lt $Prepared.ReleaseTicks) { throw 'protocol EXEC_BEGIN precedes formal release' }
    $Prepared.Phases.Add('EXEC_BEGIN')
    $output = [Collections.Generic.List[string]]::new()
    do {
      [byte[]]$wire = Read-C12PreparedMessage -Pipe $Prepared.Pipe -Deadline $releaseDeadline
      $envelope = [Text.UTF8Encoding]::new($false,$true).GetString($wire) | ConvertFrom-Json
      if ($envelope.type -cnotin @('OUTPUT','EXEC_END')) { throw 'protocol execution frame order mismatch' }
      $payload = Assert-C12PreparedFrame -State $Prepared.Protocol -Bytes $wire -Type $envelope.type
      if ($envelope.type -ceq 'OUTPUT') {
        if ($output.Count -ge 4096 -or $payload.line -isnot [string] -or $payload.line.Length -gt 512) { throw 'protocol output exceeds bound' }
        $output.Add($payload.line)
      }
    } while ($envelope.type -cne 'EXEC_END')
    if ([long]$payload.timestamp_ticks -lt [long]$begin.timestamp_ticks -or [long]$payload.timestamp_ticks -lt $Prepared.ReleaseTicks -or $payload.output_count -ne $output.Count) { throw 'protocol EXEC_END precedes release or result count mismatch' }
    $Prepared.Phases.Add('EXEC_END')
    $Prepared.ExecutionComplete = $true
    $Prepared.Phase = 'Exited'
    return [pscustomobject]@{ ExitCode = [int]$payload.exit_code; Output = [string[]]$output.ToArray() }
  }
  catch { $primaryFailure = $_; throw }
  finally {
    try {
      if ($null -ne $verified.Secondary) { $verified.Secondary.Dispose() }
      $verified.Primary.Dispose()
    }
    finally {
      try { Close-C12PreparedNativeWorker -Prepared $Prepared -Deadline ([DateTime]::UtcNow.AddSeconds(10)) }
      catch {
        $Prepared.CleanupFailures += $_.Exception.Message
        if ($null -eq $primaryFailure) { throw }
      }
    }
  }
}

function Remove-C12PreparedArtifactRoot {
  param(
    [Parameter(Mandatory = $true)][object]$ArtifactRoot,
    [DateTime]$Deadline = [DateTime]::MaxValue
  )

  if ([bool]$ArtifactRoot.Closed) { return }
  foreach ($prepared in $script:c12PreparedWorkers.ToArray()) {
    if ([string]$prepared.Receipt.ArtifactRootIdentity -ceq [string]$ArtifactRoot.ArtifactRootIdentity -and -not [bool]$prepared.Closed -and -not [bool]$prepared.ExecutionComplete) {
      throw 'prepared artifact root still has a live prepared worker'
    }
  }
  if ($ArtifactRoot.PSObject.Properties.Name -cnotcontains 'RootLifecycle') { $ArtifactRoot | Add-Member NoteProperty RootLifecycle 'Bound' }
  if ($ArtifactRoot.PSObject.Properties.Name -cnotcontains 'RootLastCleanupError') { $ArtifactRoot | Add-Member NoteProperty RootLastCleanupError '' }
  if ([string]$ArtifactRoot.RootLifecycle -ceq 'CleanIntent' -and $null -eq $ArtifactRoot.Ownership) {
    $pending = [C12SealedExecutable]::TryInspectPath([string]$ArtifactRoot.Root)
    if ($null -eq $pending) { $ArtifactRoot.RootLifecycle = 'Absent' }
    elseif (-not [bool]$pending.Directory -or [bool]$pending.Reparse -or [string]$pending.Value -cne [string]$ArtifactRoot.ArtifactRootIdentity) {
      $ArtifactRoot.RootLastCleanupError = 'prepared artifact root namespace contains a foreign replacement after exact-handle delete intent'
      throw $ArtifactRoot.RootLastCleanupError
    }
    else { $ArtifactRoot.Ownership = [C12OwnedDirectory]::OpenExisting([string]$ArtifactRoot.Root, [UInt32]$pending.VolumeSerialNumber, [UInt64]$pending.FileIndex) }
  }
  if (-not [IO.Directory]::Exists([string]$ArtifactRoot.Root)) {
    if ([string]$ArtifactRoot.RootLifecycle -cnotin @('CleanIntent','Absent')) { throw 'prepared artifact root disappeared without retained CleanIntent' }
    if ($null -ne $ArtifactRoot.Ownership) { $ArtifactRoot.Ownership.ReleaseDeletePending(); $ArtifactRoot.Ownership = $null }
    $ArtifactRoot.RootLifecycle = 'Absent'
  }
  else {
    $ArtifactRoot.Ownership.VerifyExactPath()
    $observed = [C12SealedExecutable]::InspectDirectory([string]$ArtifactRoot.Root)
    if ($observed.Value -cne [string]$ArtifactRoot.ArtifactRootIdentity) { throw 'prepared artifact root identity changed before cleanup' }
  }
  if ($ArtifactRoot.PSObject.Properties.Name -cnotcontains 'Ledger' -or $null -eq $ArtifactRoot.Ledger) { throw 'prepared artifact root lacks its direct-leaf ledger' }
  if ([string]$ArtifactRoot.RootLifecycle -cne 'Absent') {
    try {
      Converge-C12DirectLeafLedger -ArtifactRoot $ArtifactRoot -Deadline $Deadline
      if (@($ArtifactRoot.Ledger | Where-Object { [string]$_.Lifecycle -cne 'Absent' }).Count -ne 0) { throw 'prepared artifact root direct-leaf cleanup did not converge' }
      if ([string]$ArtifactRoot.RootLifecycle -ceq 'Bound') { $ArtifactRoot.RootLifecycle = 'CleanIntent' }
      $ArtifactRoot.Ownership.RequestDeleteExactEmpty($Deadline)
      $ArtifactRoot.Ownership.ReleaseDeletePending()
      $ArtifactRoot.Ownership = $null
      $afterRootDelete = [C12SealedExecutable]::TryInspectPath([string]$ArtifactRoot.Root)
      if ($null -eq $afterRootDelete) { $ArtifactRoot.RootLifecycle = 'Absent' }
      elseif (-not [bool]$afterRootDelete.Directory -or [bool]$afterRootDelete.Reparse -or [string]$afterRootDelete.Value -cne [string]$ArtifactRoot.ArtifactRootIdentity) {
        throw 'prepared artifact root namespace contains a foreign replacement after exact-handle delete intent'
      }
      else {
        $ArtifactRoot.Ownership = [C12OwnedDirectory]::OpenExisting([string]$ArtifactRoot.Root, [UInt32]$afterRootDelete.VolumeSerialNumber, [UInt64]$afterRootDelete.FileIndex)
        throw 'prepared artifact root namespace retained the same identity after exact-handle delete intent'
      }
      $ArtifactRoot.RootLastCleanupError = ''
    }
    catch {
      $ArtifactRoot.RootLastCleanupError = $_.Exception.Message
      throw
    }
  }
  $ArtifactRoot.Closed = $true
  if ([IO.Directory]::Exists([string]$ArtifactRoot.Root)) { throw 'prepared artifact root remained after cleanup' }
  if ($ArtifactRoot -eq $script:c12PreparedArtifactRoot) {
    $script:c12PreparedArtifactRoot = $null
    $script:c12PreparedReceipts.Clear()
    $script:c12PreparedWorkers.Clear()
    $script:c12PreparedReceiptKeyHex = ''
  }
}

function Close-C12PreparedBundle {
  param(
    [object]$Prepared,
    [object]$ArtifactRoot,
    [DateTime]$Deadline = [DateTime]::MaxValue
  )

  $failures = @()
  if ($null -ne $Prepared) {
    try { Close-C12PreparedNativeWorker -Prepared $Prepared -Deadline $Deadline }
    catch { $failures += "worker: $($_.Exception.Message)" }
  }
  if ($null -ne $ArtifactRoot) {
    try { Remove-C12PreparedArtifactRoot -ArtifactRoot $ArtifactRoot -Deadline $Deadline }
    catch { $failures += "artifact root: $($_.Exception.Message)" }
  }
  if ($failures.Count -ne 0) { throw "prepared artifact cleanup failed: $($failures -join '; ')" }
}

function Close-C12AllPreparedArtifacts {
  param([DateTime]$Deadline = [DateTime]::MaxValue)

  $failures = @()
  foreach ($prepared in $script:c12PreparedWorkers.ToArray()) {
    if (-not [bool]$prepared.Closed) {
      try { Close-C12PreparedNativeWorker -Prepared $prepared -Deadline $Deadline }
      catch { $failures += "worker: $($_.Exception.Message)" }
    }
  }
  if ($null -ne $script:c12PreparedArtifactRoot) {
    try { Remove-C12PreparedArtifactRoot -ArtifactRoot $script:c12PreparedArtifactRoot -Deadline $Deadline }
    catch { $failures += "artifact root: $($_.Exception.Message)" }
  }
  elseif ($script:c12PreparedWorkers.Count -ne 0 -or $script:c12PreparedReceipts.Count -ne 0) {
    $failures += 'prepared controller state outlived its artifact root'
  }
  if ($null -ne $script:c12PreparedArtifactRoot -or $script:c12PreparedWorkers.Count -ne 0 -or $script:c12PreparedReceipts.Count -ne 0) {
    $failures += 'prepared artifact controller state remained after cleanup'
  }
  $script:c12PreparedReceiptKeyHex = ''
  if ($failures.Count -ne 0) { throw "prepared artifact convergence failed: $($failures -join '; ')" }
}

function Resolve-C12Test2JSONExecutable {
  param(
    [Parameter(Mandatory = $true)][object]$ArtifactRoot,
    [Parameter(Mandatory = $true)][object]$GoToolchain,
    [Parameter(Mandatory = $true)][DateTime]$SetupDeadline
  )

  $nonce = New-C12RandomSuffix
  $temporaryTest2JSONExecutable = Join-Path ([string]$ArtifactRoot.Root) "authority-test2json-$nonce.tmp.exe"
  $finalTest2JSONExecutable = Join-Path ([string]$ArtifactRoot.Root) "authority-test2json-$nonce.exe"
  $null = Register-C12DirectLeafIntent -Ledger $ArtifactRoot.Ledger -Name ([IO.Path]::GetFileName($temporaryTest2JSONExecutable)) -Kind 'exact_file' -Expected $true
  $null = Register-C12DirectLeafIntent -Ledger $ArtifactRoot.Ledger -Name ([IO.Path]::GetFileName($finalTest2JSONExecutable)) -Kind 'exact_file' -Expected $true
  $test2JSONSourceDirectory = [IO.Path]::GetFullPath((Join-Path ([string]$GoToolchain.Root) 'src\cmd\test2json'))
  $sourceIdentity = [C12SealedExecutable]::InspectDirectory($test2JSONSourceDirectory)
  if ([bool]$sourceIdentity.Reparse) { throw 'exact test2json source directory is a reparse point' }
  $test2JSONBuildArguments = @('build', '-o', $temporaryTest2JSONExecutable, '.')
  $graph = New-C12GoGraphReceipt -Purpose 'test2json' -Packages @('cmd/test2json') -Deadline $SetupDeadline -ArtifactRoot $ArtifactRoot -GoToolchain $GoToolchain -BuildArgv $test2JSONBuildArguments -WorkingDirectory $test2JSONSourceDirectory
  $null = Invoke-C12ClosedGoBuild -ArtifactRoot $ArtifactRoot -GoToolchain $GoToolchain -Arguments $test2JSONBuildArguments -WorkingDirectory $test2JSONSourceDirectory -Stage 'compile exact test2json executable' -SetupDeadline $SetupDeadline
  $published = Publish-C12PreparedExecutable -ArtifactRoot $ArtifactRoot -TemporaryPath $temporaryTest2JSONExecutable -FinalPath $finalTest2JSONExecutable -TemporaryLeafPattern '^authority-test2json-[0-9a-f]{32}\.tmp\.exe$' -FinalLeafPattern '^authority-test2json-[0-9a-f]{32}\.exe$'
  return [pscustomobject]@{
    Path = $published
    BuildArguments = [string[]]$test2JSONBuildArguments
    SourceIdentity = "cmd/test2json@$($GoToolchain.Version):$($sourceIdentity.Value)"
    GoGraphReceipt = $graph
  }
}

function New-C12PreparedTrustedValidator {
  param(
    [Parameter(Mandatory = $true)]
    [ValidateSet('suite', 'focused')]
    [string]$Mode,

    [Parameter(Mandatory = $true)]
    [string]$DataRoot,

    [string]$SuiteName = '',

    [string]$SuiteTimeout = '',

    [string]$CandidateTree = '',

    [string[]]$ManifestPaths = @(),

    [string[]]$Packages = @(),

    [string[]]$Tests = @(),

    [Parameter(Mandatory = $true)]
    [ValidateSet('base', 'authority-v7', 'authority-v7-pitr')]
    [string]$Profile,

    [Parameter(Mandatory = $true)]
    [TimeSpan]$SetupAllowance,

    [DateTime]$Deadline = [DateTime]::MaxValue
  )

  if ($SetupAllowance -ne [TimeSpan]$script:c12ProfileAllowances[$Profile]) { throw 'trusted validator setup allowance does not match its closed profile' }
  $setupDeadline = [DateTime]::UtcNow.Add($SetupAllowance)
  if ($Deadline -lt $setupDeadline) { $setupDeadline = $Deadline }
  if ([DateTime]::UtcNow -ge $setupDeadline) { throw 'trusted candidate validation exceeded its setup allowance before bootstrap' }
  $resolvedDataRoot = [IO.Path]::GetFullPath($DataRoot).TrimEnd('\')
  if (-not [IO.Directory]::Exists($resolvedDataRoot)) {
    throw 'trusted candidate validation data root is absent'
  }
  $artifactRoot = Get-C12PreparedArtifactRoot -Profile $Profile
  $preparedValidator = $null
  try {
    $nonce = New-C12RandomSuffix
    $sourcePath = Join-Path ([string]$artifactRoot.Root) "trusted-validator-source-$nonce.go"
    $temporaryExecutable = Join-Path ([string]$artifactRoot.Root) "trusted-validator-$nonce.tmp.exe"
    $finalExecutable = Join-Path ([string]$artifactRoot.Root) "trusted-validator-$nonce.exe"
    $null = Register-C12DirectLeafIntent -Ledger $artifactRoot.Ledger -Name ([IO.Path]::GetFileName($sourcePath)) -Kind 'exact_file' -Expected $true
    $null = Register-C12DirectLeafIntent -Ledger $artifactRoot.Ledger -Name ([IO.Path]::GetFileName($temporaryExecutable)) -Kind 'exact_file' -Expected $true
    $null = Register-C12DirectLeafIntent -Ledger $artifactRoot.Ledger -Name ([IO.Path]::GetFileName($finalExecutable)) -Kind 'exact_file' -Expected $true
    $encoding = New-Object System.Text.UTF8Encoding($false)
    $sourceBytes = $encoding.GetBytes([string]$script:c12TrustedValidatorSource)
    $sourceDigest = Get-C12SHA256Hex -Bytes $sourceBytes
    if ($sourceDigest -cne $script:c12TrustedValidatorSourceSHA256) {
      throw 'trusted candidate validation embedded-source digest differs from the controller constant'
    }
    [IO.File]::WriteAllBytes($sourcePath, $sourceBytes)
    $writtenBytes = [IO.File]::ReadAllBytes($sourcePath)
    if ((Get-C12SHA256Hex -Bytes $writtenBytes) -cne $sourceDigest -or $sourceBytes.Length -ne $writtenBytes.Length) {
      throw 'trusted candidate validation bootstrap bytes changed after materialization'
    }
    Protect-C12PrivateArtifactFile -Path $sourcePath
    $null = Bind-C12DirectLeaf -ArtifactRoot $artifactRoot -Name ([IO.Path]::GetFileName($sourcePath))
    $goToolchain = Resolve-C12ClosedGoToolchain -ArtifactRoot $artifactRoot -SetupDeadline $setupDeadline
    $validatorBuildArguments = @('build', '-o', $temporaryExecutable, $sourcePath)
    $validatorGraph = New-C12GoGraphReceipt -Purpose 'trusted-validator' -Packages @($sourcePath) -Deadline $setupDeadline -ArtifactRoot $artifactRoot -GoToolchain $goToolchain -BuildArgv $validatorBuildArguments -WorkingDirectory ([string]$artifactRoot.Root)
    $null = Invoke-C12ClosedGoBuild -ArtifactRoot $artifactRoot -GoToolchain $goToolchain -Arguments $validatorBuildArguments -WorkingDirectory ([string]$artifactRoot.Root) -Stage 'compile exact trusted validator' -SetupDeadline $setupDeadline
    $publishedExecutable = Publish-C12PreparedExecutable -ArtifactRoot $artifactRoot -TemporaryPath $temporaryExecutable -FinalPath $finalExecutable -TemporaryLeafPattern '^trusted-validator-[0-9a-f]{32}\.tmp\.exe$' -FinalLeafPattern '^trusted-validator-[0-9a-f]{32}\.exe$'
    $arguments = @(
      '-mode', $Mode,
      '-version', $script:c12TrustedValidatorVersion,
      '-source-digest', $sourceDigest,
      '-root', $resolvedDataRoot
    )
    if ($Mode -ceq 'suite') {
      $arguments += @(
        '-suite', $SuiteName,
        '-suite-timeout', $SuiteTimeout,
        '-candidate-tree', $CandidateTree,
        '-manifests', ($ManifestPaths -join '|')
      )
    }
    else {
      $arguments += @(
        '-packages', ($Packages -join '|'),
        '-tests', ($Tests -join '|')
      )
    }
    $candidateIdentity = Get-C12PreparedCandidateIdentity -DataRoot $resolvedDataRoot -CandidateTree $CandidateTree
    $candidateIdentity = Get-C12SHA256Hex ([Text.Encoding]::UTF8.GetBytes("$candidateIdentity`n$($validatorGraph.CandidateTreeDigest)"))
    $receipt = New-C12SealedExecutableReceipt -ArtifactRoot $artifactRoot -Role 'trusted-validator' -Profile $Profile -Purpose 'trusted-validator' -SourceIdentity $script:c12TrustedValidatorVersion -SourceDigest $sourceDigest -CandidateTreeIdentity $candidateIdentity -BuildArguments $validatorBuildArguments -GoToolchain $goToolchain -ExecutablePath $publishedExecutable -Arguments $arguments -WorkingDirectory ([string]$artifactRoot.Root) -GoGraphReceipts @($validatorGraph)
    $preparedValidator = New-C12PreparedNativeWorker -ArtifactRoot $artifactRoot -Receipt $receipt -SetupDeadline $setupDeadline
    return $preparedValidator
  }
  catch {
    $primary = $_.Exception
    try { Close-C12PreparedBundle -Prepared $preparedValidator -ArtifactRoot $artifactRoot -Deadline ([DateTime]::UtcNow.Add($script:c12GroupCleanupBudget)) }
    catch { throw "$($primary.Message); $($_.Exception.Message)" }
    throw $primary
  }
}

function Invoke-C12TrustedValidator {
  param(
    [Parameter(Mandatory = $true)][ValidateSet('suite', 'focused')][string]$Mode,
    [Parameter(Mandatory = $true)][string]$DataRoot,
    [string]$SuiteName = '',
    [string]$SuiteTimeout = '',
    [string]$CandidateTree = '',
    [string[]]$ManifestPaths = @(),
    [string[]]$Packages = @(),
    [string[]]$Tests = @(),
    [Parameter(Mandatory = $true)][ValidateSet('base', 'authority-v7', 'authority-v7-pitr')][string]$Profile,
    [Parameter(Mandatory = $true)][TimeSpan]$SetupAllowance,
    [DateTime]$Deadline = [DateTime]::MaxValue
  )

  $preparedValidator = $null
  $artifactRoot = $null
  $result = $null
  $primaryFailure = $null
  $cleanupFailure = $null
  try {
    $preparedValidator = New-C12PreparedTrustedValidator -Mode $Mode -DataRoot $DataRoot -SuiteName $SuiteName -SuiteTimeout $SuiteTimeout -CandidateTree $CandidateTree -ManifestPaths $ManifestPaths -Packages $Packages -Tests $Tests -Profile $Profile -SetupAllowance $SetupAllowance -Deadline $Deadline
    $artifactRoot = $preparedValidator.ArtifactRoot
    $result = Invoke-C12PreparedAuthorityRole -Prepared $preparedValidator -Role 'trusted-validator' -Timeout ([TimeSpan]::FromMinutes(2)) -Deadline $Deadline -CapabilityEnvironment @{}
    $successMarker = "C12_TRUSTED_VALIDATOR_OK:$($script:c12TrustedValidatorVersion):$($script:c12TrustedValidatorSourceSHA256)"
    $markerLines = @($result.Output | Where-Object { ([string]$_).StartsWith('C12_TRUSTED_VALIDATOR_OK:', [StringComparison]::Ordinal) })
    if ($result.ExitCode -ne 0 -or $markerLines.Count -ne 1 -or [string]$markerLines[0] -cne $successMarker) {
      throw 'trusted candidate validation did not return one exact version/source marker'
    }
  }
  catch { $primaryFailure = $_.Exception.Message }
  finally {
    if ($null -eq $artifactRoot -and $null -ne $script:c12PreparedArtifactRoot) { $artifactRoot = $script:c12PreparedArtifactRoot }
    try { Close-C12PreparedBundle -Prepared $preparedValidator -ArtifactRoot $artifactRoot -Deadline ([DateTime]::UtcNow.Add($script:c12GroupCleanupBudget)) }
    catch { $cleanupFailure = $_.Exception.Message }
  }
  if ($null -ne $primaryFailure -or $null -ne $cleanupFailure) {
    if ($null -ne $cleanupFailure) { throw "trusted candidate validation failed: $primaryFailure; cleanup: $cleanupFailure" }
    throw $primaryFailure
  }
  return $result
}

function New-C12PreparedAuthorityInitializer {
  param(
    [Parameter(Mandatory = $true)][ValidateSet('authority-v7', 'authority-v7-pitr')][string]$Profile,
    [Parameter(Mandatory = $true)][string]$RunSuffix,
    [string]$CandidateTree = '',
    [Parameter(Mandatory = $true)][TimeSpan]$SetupAllowance,
    [DateTime]$Deadline = [DateTime]::MaxValue
  )

  if ($RunSuffix -notmatch '^[0-9a-f]{32}$' -or $SetupAllowance -ne [TimeSpan]$script:c12ProfileAllowances[$Profile]) {
    throw 'authority initializer preparation has a malformed run/profile allowance'
  }
  $setupDeadline = [DateTime]::UtcNow.Add($SetupAllowance)
  if ($Deadline -lt $setupDeadline) { $setupDeadline = $Deadline }
  if ([DateTime]::UtcNow -ge $setupDeadline) { throw 'authority initializer preparation exceeded setup allowance' }
  $artifactRoot = Get-C12PreparedArtifactRoot -Profile $Profile -RunSuffix $RunSuffix
  $preparedInitializer = $null
  try {
    $goToolchain = Resolve-C12ClosedGoToolchain -ArtifactRoot $artifactRoot -SetupDeadline $setupDeadline
    $nonce = New-C12RandomSuffix
    $temporaryTestExecutable = Join-Path ([string]$artifactRoot.Root) "authority-initializer-$nonce.tmp.test.exe"
    $finalTestExecutable = Join-Path ([string]$artifactRoot.Root) "authority-initializer-$nonce.test.exe"
    $null = Register-C12DirectLeafIntent -Ledger $artifactRoot.Ledger -Name ([IO.Path]::GetFileName($temporaryTestExecutable)) -Kind 'exact_file' -Expected $true
    $null = Register-C12DirectLeafIntent -Ledger $artifactRoot.Ledger -Name ([IO.Path]::GetFileName($finalTestExecutable)) -Kind 'exact_file' -Expected $true
    $initializerBuildArguments = @('test', '-c', '-tags=integration', '-p=1', '-o', $temporaryTestExecutable, './internal/testinfra')
    $initializerGraph = New-C12GoGraphReceipt -Purpose 'authority-initializer' -Packages @('./internal/testinfra') -Deadline $setupDeadline -ArtifactRoot $artifactRoot -GoToolchain $goToolchain -BuildArgv $initializerBuildArguments -WorkingDirectory $script:c12RepositoryRoot
    $null = Invoke-C12ClosedGoBuild -ArtifactRoot $artifactRoot -GoToolchain $goToolchain -Arguments $initializerBuildArguments -WorkingDirectory $script:c12RepositoryRoot -Stage 'compile authority initializer test binary' -SetupDeadline $setupDeadline
    $publishedTestExecutable = Publish-C12PreparedExecutable -ArtifactRoot $artifactRoot -TemporaryPath $temporaryTestExecutable -FinalPath $finalTestExecutable -TemporaryLeafPattern '^authority-initializer-[0-9a-f]{32}\.tmp\.test\.exe$' -FinalLeafPattern '^authority-initializer-[0-9a-f]{32}\.test\.exe$'
    $test2JSON = Resolve-C12Test2JSONExecutable -ArtifactRoot $artifactRoot -GoToolchain $goToolchain -SetupDeadline $setupDeadline
    $arguments = @(
      '-p', 'talenro.local/platform/internal/testinfra', '-t', [string]$publishedTestExecutable,
      '-test.v=test2json', '-test.paniconexit0', '-test.timeout=2m',
      '-test.run', '^TestPrepareC12AuthorityV7Database$', '-test.count=1'
    )
    $candidateIdentity = Get-C12PreparedCandidateIdentity -DataRoot $script:c12RepositoryRoot -CandidateTree $CandidateTree
    $combinedBuildArguments = @($initializerBuildArguments + @('--sealed-test2json-build--') + @($test2JSON.BuildArguments))
    $sourceDigest = Get-C12SHA256Hex -Bytes ([Text.Encoding]::UTF8.GetBytes("talenro.local/platform/internal/testinfra`n$candidateIdentity`n$(Get-C12PreparedStringArrayDigest -Values $combinedBuildArguments)"))
    $candidateIdentity = Get-C12SHA256Hex ([Text.Encoding]::UTF8.GetBytes("$candidateIdentity`n$($initializerGraph.CandidateTreeDigest)`n$($test2JSON.GoGraphReceipt.CandidateTreeDigest)"))
    $receipt = New-C12SealedExecutableReceipt -ArtifactRoot $artifactRoot -Role 'authority-initializer-json' -Profile $Profile -Purpose 'authority-initializer' -SourceIdentity 'talenro.local/platform/internal/testinfra' -SourceDigest $sourceDigest -CandidateTreeIdentity $candidateIdentity -BuildArguments $combinedBuildArguments -GoToolchain $goToolchain -ExecutablePath ([string]$test2JSON.Path) -SecondaryExecutablePath $publishedTestExecutable -Arguments $arguments -WorkingDirectory $script:c12RepositoryRoot -GoGraphReceipts @($initializerGraph, $test2JSON.GoGraphReceipt)
    $preparedInitializer = New-C12PreparedNativeWorker -ArtifactRoot $artifactRoot -Receipt $receipt -SetupDeadline $setupDeadline
    return $preparedInitializer
  }
  catch {
    $primary = $_.Exception
    try { Close-C12PreparedBundle -Prepared $preparedInitializer -ArtifactRoot $artifactRoot -Deadline ([DateTime]::UtcNow.Add($script:c12GroupCleanupBudget)) }
    catch { throw "$($primary.Message); $($_.Exception.Message)" }
    throw $primary
  }
}

function Invoke-C12AuthorityInitializer {
  param(
    [Parameter(Mandatory = $true)][object]$Prepared,
    [Parameter(Mandatory = $true)][string]$DatabaseURL,
    [Parameter(Mandatory = $true)][string]$InitNonce,
    [Parameter(Mandatory = $true)][string]$RunSuffix,
    [Parameter(Mandatory = $true)][ValidateSet('authority-v7', 'authority-v7-pitr')][string]$Profile,
    [DateTime]$Deadline = [DateTime]::MaxValue
  )

  if ($InitNonce -notmatch '^[0-9a-f]{64}$' -or $RunSuffix -notmatch '^[0-9a-f]{32}$' -or
      -not $DatabaseURL.StartsWith('postgres://talenro:', [StringComparison]::Ordinal)) {
    throw 'authority initializer formal capability set is malformed'
  }
  $preparedInitializer = $Prepared
  $capabilities = @{
    TALENRO_C12_AUTHORITY_V7_INIT_NONCE = $InitNonce
    TALENRO_C12_AUTHORITY_V7_RUN_SUFFIX = $RunSuffix
    TALENRO_C12_AUTHORITY_V7_PROFILE = $Profile
    TALENRO_DATABASE_URL = $DatabaseURL
    TALENRO_INSTALLATION_KIND = 'disposable_fixture'
  }
  try {
    $result = Invoke-C12PreparedAuthorityRole -Prepared $preparedInitializer -Role 'authority-initializer-json' -Timeout ([TimeSpan]::FromMinutes(2)) -Deadline $Deadline -CapabilityEnvironment $capabilities
    Assert-C12GoJSONResult -Result $result -Package 'talenro.local/platform/internal/testinfra' -ExpectedTests @('TestPrepareC12AuthorityV7Database')
    return $result
  }
  finally {
    foreach ($name in @($capabilities.Keys)) { $capabilities[$name] = '' }
    Clear-C12InheritedCapabilities
  }
}

function Get-C12SingleOutputLine {
  param(
    [Parameter(Mandatory = $true)]
    [object]$Result,

    [Parameter(Mandatory = $true)]
    [string]$Stage
  )

  $lines = @($Result.Output | Where-Object { -not [string]::IsNullOrWhiteSpace($_) })
  if ($Result.ExitCode -ne 0 -or $lines.Count -ne 1) {
    throw "$Stage did not return one bounded line"
  }
  return [string]$lines[0]
}

function Assert-C12ContainerIdentity {
  param(
    [Parameter(Mandatory = $true)]
    [object]$Resource,

    [Parameter(Mandatory = $true)]
    [string]$RunSuffix
  )

  if ([string]::IsNullOrEmpty([string]$Resource.ID)) {
    throw "container $($Resource.Kind) has no captured ID"
  }
  $isPITR = $Resource.PSObject.Properties.Name -contains 'PITR' -and [bool]$Resource.PITR
  $format = if ($isPITR) {
    '{{.Id}}|{{.Name}}|{{ index .Config.Labels `talenro.c12.managed` }}|{{ index .Config.Labels `talenro.c12.run` }}|{{ index .Config.Labels `talenro.c12.profile` }}|{{ index .Config.Labels `talenro.c12.role` }}|{{ index .Config.Labels `talenro.c12.nonce-digest` }}|{{.Config.Image}}|{{.Image}}'
  }
  else {
    '{{.Id}}|{{.Name}}|{{ index .Config.Labels `talenro.c12.run` }}|{{ index .Config.Labels `talenro.c12.role` }}|{{.Config.Image}}|{{.Image}}'
  }
  $dockerArgs = @('container', 'inspect', '--format', $format, [string]$Resource.ID)
  $result = Invoke-C12Docker -Arguments $dockerArgs -Stage "inspect $($Resource.Kind) identity"
  $identity = Get-C12SingleOutputLine -Result $result -Stage "inspect $($Resource.Kind) identity"
  if ($isPITR) {
    $parts = $identity -split '\|', 9
    if ($parts.Count -ne 9 -or $parts[0] -cne [string]$Resource.ID -or $parts[1] -cne "/$($Resource.Name)" -or
        $parts[2] -cne 'true' -or $parts[3] -cne $RunSuffix -or $parts[4] -cne 'authority-v7-pitr' -or
        $parts[5] -cne 'primary' -or $parts[6] -cne [string]$Resource.NonceDigest -or
        $parts[7] -cne [string]$Resource.ImageRef -or $parts[8] -cne [string]$Resource.ImageID) {
      throw 'PITR primary captured ID/name/labels/image identity mismatch'
    }
  }
  else {
    $parts = $identity -split '\|', 6
    if ($parts.Count -ne 6 -or $parts[0] -cne [string]$Resource.ID -or $parts[1] -cne "/$($Resource.Name)" -or $parts[2] -cne $RunSuffix) {
      throw "container $($Resource.Kind) captured ID/name/run-label mismatch"
    }
    if ($parts[3] -cne [string]$Resource.Kind) { throw "container $($Resource.Kind) role mismatch" }
    if ($parts[4] -cne [string]$Resource.ImageRef) { throw "container $($Resource.Kind) image reference mismatch" }
    if ($parts[5] -cne [string]$Resource.ImageID) { throw "container $($Resource.Kind) immutable image ID mismatch" }
  }
}

function Resolve-C12ImageIdentity {
  param(
    [Parameter(Mandatory = $true)]
    [object]$Resource
  )

  $dockerArgs = @('image', 'inspect', '--format', '{{.Id}}', [string]$Resource.ImageRef)
  $result = Invoke-C12Docker -Arguments $dockerArgs -Stage "resolve $($Resource.Kind) immutable image"
  $imageID = Get-C12SingleOutputLine -Result $result -Stage "resolve $($Resource.Kind) immutable image"
  if ($imageID -notmatch '^sha256:[0-9a-f]{64}$') {
    throw "resolve $($Resource.Kind) immutable image returned a malformed image ID"
  }
  $Resource.ImageID = $imageID
}

function Set-C12MappedPort {
  param(
    [Parameter(Mandatory = $true)]
    [object]$Resource
  )

  $dockerArgs = @('container', 'port', [string]$Resource.ID, "$($Resource.ContainerPort)/tcp")
  $result = Invoke-C12Docker -Arguments $dockerArgs -Stage "inspect $($Resource.Kind) mapped port"
  $mapping = Get-C12SingleOutputLine -Result $result -Stage "inspect $($Resource.Kind) mapped port"
  if ($mapping -notmatch '^127\.0\.0\.1:([1-9][0-9]{0,4})$') {
    throw "container $($Resource.Kind) has a non-loopback or malformed mapped port"
  }
  $port = [int]$Matches[1]
  if ($port -gt 65535) {
    throw "container $($Resource.Kind) has an invalid mapped port"
  }
  $Resource.Port = $port
}

function Start-C12Container {
  param(
    [Parameter(Mandatory = $true)]
    [object]$Resource,

    [Parameter(Mandatory = $true)]
    [string]$RunSuffix,

    [Parameter(Mandatory = $true)]
    [string]$DatabaseName,

    [Parameter(Mandatory = $true)]
    [string]$DatabasePassword
  )

  switch ($Resource.Kind) {
    'postgres' {
      $isPITR = $Resource.PSObject.Properties.Name -contains 'PITR' -and [bool]$Resource.PITR
      if ($isPITR) {
        $dockerArgs = @(
          'run', '--detach', '--name', [string]$Resource.Name,
          '--label', 'talenro.c12.managed=true',
          '--label', "talenro.c12.run=$RunSuffix",
          '--label', 'talenro.c12.profile=authority-v7-pitr',
          '--label', 'talenro.c12.role=primary',
          '--label', "talenro.c12.nonce-digest=$([string]$Resource.NonceDigest)",
          '--publish', '127.0.0.1::5432',
          '--mount', "type=volume,src=$([string]$Resource.PrimaryDataName),dst=/var/lib/postgresql",
          '--mount', "type=volume,src=$([string]$Resource.ArchiveName),dst=/archive",
          '--mount', "type=volume,src=$([string]$Resource.BaseBackupName),dst=/basebackup",
          '--env', 'POSTGRES_USER=talenro',
          '--env', "POSTGRES_PASSWORD=$DatabasePassword",
          '--env', "POSTGRES_DB=$DatabaseName",
          '--health-cmd', "pg_isready -U talenro -d $DatabaseName",
          '--health-interval', '1s', '--health-timeout', '2s', '--health-retries', '60',
          [string]$Resource.ImageRef,
          '-c', 'wal_level=logical',
          '-c', 'max_replication_slots=4',
          '-c', 'max_wal_senders=4',
          '-c', 'max_prepared_transactions=8',
          '-c', 'archive_mode=on',
          '-c', 'archive_command=test ! -f /archive/%f && cp %p /archive/%f',
          '-c', 'full_page_writes=on'
        )
      }
      else {
        $dockerArgs = @(
          'run', '--detach', '--name', [string]$Resource.Name,
          '--label', "talenro.c12.run=$RunSuffix",
          '--label', 'talenro.c12.role=postgres',
          '--publish', '127.0.0.1::5432',
          '--env', 'POSTGRES_USER=talenro',
          '--env', "POSTGRES_PASSWORD=$DatabasePassword",
          '--env', "POSTGRES_DB=$DatabaseName",
          '--health-cmd', "pg_isready -U talenro -d $DatabaseName",
          '--health-interval', '1s', '--health-timeout', '2s', '--health-retries', '60',
          [string]$Resource.ImageRef
        )
      }
    }
    'redis' {
      $dockerArgs = @(
        'run', '--detach', '--name', [string]$Resource.Name,
        '--label', "talenro.c12.run=$RunSuffix",
        '--label', 'talenro.c12.role=redis',
        '--publish', '127.0.0.1::6379',
        [string]$Resource.ImageRef
      )
    }
    'nats' {
      $dockerArgs = @(
        'run', '--detach', '--name', [string]$Resource.Name,
        '--label', "talenro.c12.run=$RunSuffix",
        '--label', 'talenro.c12.role=nats',
        '--publish', '127.0.0.1::4222',
        [string]$Resource.ImageRef, '-js', '--name', [string]$Resource.Name
      )
    }
    default {
      throw "unknown C12 dependency kind $($Resource.Kind)"
    }
  }

  Register-C12DockerIntent -Resource $Resource
  $createFailure=$null
  try {
    $result = Invoke-C12Docker -Arguments $dockerArgs -Stage "start $($Resource.Kind) container"
  }
  catch {
    $createFailure=$_
    $format='{{.Id}}|{{.Name}}|{{ index .Config.Labels `talenro.c12.run` }}|{{ index .Config.Labels `talenro.c12.role` }}|{{.Config.Image}}|{{.Image}}'
    $reinspect=Invoke-C12Docker -Arguments @('container','inspect','--format',$format,[string]$Resource.Name) -Stage "re-inspect uncertain $($Resource.Kind) create" -Deadline $script:c12NativeDeadline -AllowFailure
    if($reinspect.ExitCode-ne 0){ throw $createFailure }
    $identity=Get-C12SingleOutputLine -Result $reinspect -Stage "re-inspect uncertain $($Resource.Kind) create"
    $parts=$identity -split '\|',6
    if($parts.Count-ne 6-or $parts[0]-notmatch '^[0-9a-f]{64}$'-or $parts[1]-cne "/$($Resource.Name)"-or $parts[2]-cne $RunSuffix-or $parts[3]-cne[string]$Resource.Kind-or $parts[4]-cne[string]$Resource.ImageRef-or $parts[5]-cne[string]$Resource.ImageID){throw "$($createFailure.Exception.Message); possible orphan name $($Resource.Name) was not adopted because exact identity mismatched"}
    $result=[pscustomobject]@{ExitCode=0;Output=@([string]$parts[0]);EndpointReceipt=$reinspect.EndpointReceipt;ExecutableReceipt=$reinspect.ExecutableReceipt}
  }
  $containerID = Get-C12SingleOutputLine -Result $result -Stage "start $($Resource.Kind) container"
  if ($containerID -notmatch '^[0-9a-f]{64}$') {
    throw "start $($Resource.Kind) did not return an exact container ID"
  }
  $Resource.ID = $containerID
  Assert-C12ContainerIdentity -Resource $Resource -RunSuffix $RunSuffix
  Confirm-C12DockerActual -Resource $Resource
  Set-C12MappedPort -Resource $Resource
}

function Test-C12TCPProtocol {
  param(
    [Parameter(Mandatory = $true)]
    [int]$Port,

    [Parameter(Mandatory = $true)]
    [string]$Request,

    [Parameter(Mandatory = $true)]
    [string]$Expected,

    [Parameter(Mandatory = $true)]
    [DateTime]$Deadline
  )

  $client = New-Object System.Net.Sockets.TcpClient
  $connectWaitHandle = $null
  try {
    $remainingMilliseconds = [Math]::Floor(($Deadline - [DateTime]::UtcNow).TotalMilliseconds)
    if ($remainingMilliseconds -lt 1) {
      return $false
    }
    $connection = $client.BeginConnect('127.0.0.1', $Port, $null, $null)
    $connectWaitHandle = $connection.AsyncWaitHandle
    $connectMilliseconds = [int][Math]::Min(1000, $remainingMilliseconds)
    if (-not $connectWaitHandle.WaitOne($connectMilliseconds)) {
      return $false
    }
    $client.EndConnect($connection)
    $stream = $client.GetStream()
    $remainingMilliseconds = [Math]::Floor(($Deadline - [DateTime]::UtcNow).TotalMilliseconds)
    if ($remainingMilliseconds -lt 1) {
      return $false
    }
    $stream.WriteTimeout = [int][Math]::Min(1000, $remainingMilliseconds)
    $requestBytes = [System.Text.Encoding]::ASCII.GetBytes($Request)
    $stream.Write($requestBytes, 0, $requestBytes.Length)
    $stream.Flush()
    $buffer = New-Object byte[] 4096
    $response = New-Object System.Text.StringBuilder
    for ($attempt = 0; $attempt -lt 3; $attempt++) {
      $remainingMilliseconds = [Math]::Floor(($Deadline - [DateTime]::UtcNow).TotalMilliseconds)
      if ($remainingMilliseconds -lt 1) {
        return $false
      }
      $stream.ReadTimeout = [int][Math]::Min(1000, $remainingMilliseconds)
      $read = $stream.Read($buffer, 0, $buffer.Length)
      if ($read -le 0) {
        break
      }
      [void]$response.Append([System.Text.Encoding]::ASCII.GetString($buffer, 0, $read))
      if ($response.ToString().Contains($Expected)) {
        return $true
      }
    }
    return $false
  }
  catch {
    return $false
  }
  finally {
    if ($null -ne $connectWaitHandle) {
      $connectWaitHandle.Close()
    }
    $client.Close()
  }
}

function Test-C12TCPConnect {
  param(
    [Parameter(Mandatory = $true)][int]$Port,
    [Parameter(Mandatory = $true)][DateTime]$Deadline
  )

  $client = New-Object Net.Sockets.TcpClient
  $waitHandle = $null
  try {
    $remainingMilliseconds = [Math]::Floor(($Deadline - [DateTime]::UtcNow).TotalMilliseconds)
    if ($remainingMilliseconds -lt 1) { return $false }
    $connection = $client.BeginConnect('127.0.0.1', $Port, $null, $null)
    $waitHandle = $connection.AsyncWaitHandle
    if (-not $waitHandle.WaitOne([int][Math]::Min(1000, $remainingMilliseconds))) { return $false }
    $client.EndConnect($connection)
    return $true
  }
  catch { return $false }
  finally {
    if ($null -ne $waitHandle) { $waitHandle.Close() }
    $client.Close()
  }
}

function Wait-C12Dependencies {
  param(
    [Parameter(Mandatory = $true)]
    [object[]]$Resources,

    [TimeSpan]$ProbeTimeout = [TimeSpan]::FromSeconds(60)
  )

  if ($ProbeTimeout -le [TimeSpan]::Zero -or
      $ProbeTimeout -gt [TimeSpan]::FromSeconds(60) -or
      $ProbeTimeout.Ticks % [TimeSpan]::TicksPerSecond -ne 0) {
    throw 'C12 dependency probe timeout must be a whole number of seconds from 1 through 60'
  }
  $postgres = @($Resources | Where-Object { $_.Kind -eq 'postgres' })[0]
  $redis = @($Resources | Where-Object { $_.Kind -eq 'redis' })[0]
  $nats = @($Resources | Where-Object { $_.Kind -eq 'nats' })[0]
  $deadline = [DateTime]::UtcNow.Add($ProbeTimeout)
  if ($script:c12NativeDeadline -lt $deadline) {
    $deadline = $script:c12NativeDeadline
  }
  while ([DateTime]::UtcNow -lt $deadline) {
    $dockerArgs = @('container', 'inspect', '--format', '{{.State.Health.Status}}', [string]$postgres.ID)
    try {
      $healthResult = Invoke-C12Docker -Arguments $dockerArgs -Stage 'inspect PostgreSQL health' -Deadline $deadline -AllowFailure
    }
    catch {
      $remaining = $deadline - [DateTime]::UtcNow
      if ($remaining -le [TimeSpan]::FromSeconds(1) -and
          $_.Exception.Message -match '^inspect PostgreSQL health (?:exceeded its absolute deadline|timed out (?:before native execution|after [0-9]+ seconds)|has less than one bounded second remaining)$') {
        break
      }
      throw
    }
    $postgresReady = $false
    if ($healthResult.ExitCode -eq 0) {
      $healthLines = @($healthResult.Output | Where-Object { -not [string]::IsNullOrWhiteSpace($_) })
      $postgresReady = $healthLines.Count -eq 1 -and [string]$healthLines[0] -ceq 'healthy' -and (Test-C12TCPConnect -Port $postgres.Port -Deadline $deadline)
    }
    $redisReady = Test-C12TCPProtocol -Port $redis.Port -Request "*1`r`n`$4`r`nPING`r`n" -Expected "+PONG`r`n" -Deadline $deadline
    if ([DateTime]::UtcNow -ge $deadline) {
      break
    }
    $natsReady = Test-C12TCPProtocol -Port $nats.Port -Request "CONNECT {`"verbose`":false,`"pedantic`":false}`r`nPING`r`n" -Expected "PONG`r`n" -Deadline $deadline
    if ($postgresReady -and $redisReady -and $natsReady) {
      return
    }
    $remainingMilliseconds = [Math]::Floor(($deadline - [DateTime]::UtcNow).TotalMilliseconds)
    if ($remainingMilliseconds -lt 1) {
      break
    }
    Start-Sleep -Milliseconds ([int][Math]::Min(250, $remainingMilliseconds))
  }
  $probeSeconds = [int]$ProbeTimeout.TotalSeconds
  throw "C12 dependencies did not pass health and protocol probes within $probeSeconds seconds"
}

function Resolve-C12CleanupIdentity {
  param(
    [Parameter(Mandatory = $true)]
    [object]$Resource,

    [Parameter(Mandatory = $true)]
    [string]$RunSuffix,

    [Parameter(Mandatory = $true)]
    [DateTime]$Deadline
  )

  if ([string]::IsNullOrEmpty([string]$Resource.ID)) {
    return $null
  }
  $isPITR = $Resource.PSObject.Properties.Name -contains 'PITR' -and [bool]$Resource.PITR
  $format = if ($isPITR) {
    '{{.Id}}|{{.Name}}|{{ index .Config.Labels `talenro.c12.managed` }}|{{ index .Config.Labels `talenro.c12.run` }}|{{ index .Config.Labels `talenro.c12.profile` }}|{{ index .Config.Labels `talenro.c12.role` }}|{{ index .Config.Labels `talenro.c12.nonce-digest` }}|{{.Config.Image}}|{{.Image}}'
  }
  else {
    '{{.Id}}|{{.Name}}|{{ index .Config.Labels `talenro.c12.run` }}|{{ index .Config.Labels `talenro.c12.role` }}|{{.Config.Image}}|{{.Image}}'
  }
  $dockerArgs = @('container', 'inspect', '--format', $format, [string]$Resource.ID)
  $result = Invoke-C12Docker -Arguments $dockerArgs -Stage "re-inspect $($Resource.Kind) before cleanup" -Timeout ([TimeSpan]::FromSeconds(5)) -Deadline $Deadline -AllowFailure
  if ($result.ExitCode -ne 0) {
    throw "captured $($Resource.Kind) container is absent before cleanup"
  }
  $identity = Get-C12SingleOutputLine -Result $result -Stage "re-inspect $($Resource.Kind) before cleanup"
  if ($isPITR) {
    $parts = $identity -split '\|', 9
    if ($parts.Count -ne 9 -or $parts[0] -cne [string]$Resource.ID -or $parts[1] -cne "/$($Resource.Name)" -or
        $parts[2] -cne 'true' -or $parts[3] -cne $RunSuffix -or $parts[4] -cne 'authority-v7-pitr' -or
        $parts[5] -cne 'primary' -or $parts[6] -cne [string]$Resource.NonceDigest -or
        $parts[7] -cne [string]$Resource.ImageRef -or $parts[8] -cne [string]$Resource.ImageID) {
      throw 'refusing cleanup for PITR primary: captured identity mismatch'
    }
  }
  else {
    $parts = $identity -split '\|', 6
    if ($parts.Count -ne 6 -or $parts[0] -cne [string]$Resource.ID -or $parts[1] -cne "/$($Resource.Name)" -or
        $parts[2] -cne $RunSuffix -or $parts[3] -cne [string]$Resource.Kind -or
        $parts[4] -cne [string]$Resource.ImageRef -or $parts[5] -cne [string]$Resource.ImageID) {
      throw "refusing cleanup for $($Resource.Kind): captured ID/name/run/role/image identity mismatch"
    }
  }
  return [string]$Resource.ID
}

function Remove-C12Container {
  param(
    [Parameter(Mandatory = $true)]
    [object]$Resource,

    [Parameter(Mandatory = $true)]
    [string]$RunSuffix,

    [Parameter(Mandatory = $true)]
    [DateTime]$Deadline
  )

  $containerID = Resolve-C12CleanupIdentity -Resource $Resource -RunSuffix $RunSuffix -Deadline $Deadline
  if ([string]::IsNullOrEmpty([string]$containerID)) {
    return
  }
  if ($Resource.PSObject.Properties.Name -contains 'Phase') { $Resource.Phase='CleanIntent'; $Resource.RetryState.Attempts++; Append-C12DockerRecord -Resource $Resource -Event 'DOCKER_CLEAN_INTENT' }
  $dockerArgs = @('container', 'stop', '--time', '2', [string]$containerID)
  $null = Invoke-C12Docker -Arguments $dockerArgs -Stage "stop exact $($Resource.Kind) container" -Timeout ([TimeSpan]::FromSeconds(5)) -Deadline $Deadline
  $dockerArgs = @('container', 'rm', [string]$containerID)
  $null = Invoke-C12Docker -Arguments $dockerArgs -Stage "remove exact $($Resource.Kind) container" -Timeout ([TimeSpan]::FromSeconds(5)) -Deadline $Deadline
  $dockerArgs = @('container', 'inspect', '--format', '{{.Id}}', [string]$containerID)
  $absence = Invoke-C12Docker -Arguments $dockerArgs -Stage "verify $($Resource.Kind) cleanup" -Timeout ([TimeSpan]::FromSeconds(5)) -Deadline $Deadline -AllowFailure
  if ($absence.ExitCode -eq 0) {
    throw "$($Resource.Kind) container remains after exact cleanup"
  }
  if ($Resource.PSObject.Properties.Name -contains 'Phase') { $Resource.Phase='Removed'; Append-C12DockerRecord -Resource $Resource -Event 'DOCKER_CLEAN_RESULT' }
}

function New-C12PITRVolumeResource {
  param(
    [Parameter(Mandatory = $true)][string]$Name,
    [Parameter(Mandatory = $true)][string]$Role,
    [Parameter(Mandatory = $true)][string]$NonceDigest
  )

  if ($Name -notmatch '^talenro-c12-[0-9a-f]{32}-pitr-(primary-data|archive|basebackup|candidate-0[0-7]-data)$' -or
      $Role -notmatch '^(primary-data|archive|basebackup|candidate-0[0-7]-data)$' -or $NonceDigest -notmatch '^[0-9a-f]{64}$') {
    throw 'PITR volume identity is outside the closed registry'
  }
  return [pscustomobject]@{
    Name=$Name; Role=$Role; NonceDigest=$NonceDigest; Created=$false; CreateAttempted=$false
    Phase='NeverAttempted'; OwnershipHandle=[pscustomobject]@{ Kind='docker-volume'; Name=$Name }; RetryState=[pscustomobject]@{ Attempts=0; LastError=''; Retained=$true }
  }
}

function Register-C12DockerIntent {
  param([Parameter(Mandatory=$true)][object]$Resource)
  if (-not ($Resource.PSObject.Properties.Name -contains 'Phase')) { $Resource | Add-Member Phase 'NeverAttempted'; $Resource | Add-Member CreateAttempted $false; $Resource | Add-Member Created $false; $Resource | Add-Member OwnershipHandle ([pscustomobject]@{Kind='docker-object';Name=[string]$Resource.Name}); $Resource | Add-Member RetryState ([pscustomobject]@{Attempts=0;LastError='';Retained=$true}) }
  if ([string]$Resource.Phase -cne 'NeverAttempted') { throw 'Docker INTENT requires NeverAttempted' }
  $Resource.Phase='CreateAttempted'; $Resource.CreateAttempted=$true
  Append-C12DockerRecord -Resource $Resource -Event 'DOCKER_INTENT'
}

function Confirm-C12DockerActual {
  param([Parameter(Mandatory=$true)][object]$Resource)
  if ([string]$Resource.Phase -cne 'CreateAttempted') { throw 'Docker ACTUAL requires CreateAttempted' }
  $Resource.Created=$true; $Resource.Phase='Created'; $Resource.Phase='Verified'
  Append-C12DockerRecord -Resource $Resource -Event 'DOCKER_ACTUAL'
}

function Start-C12PITRVolume {
  param(
    [Parameter(Mandatory = $true)][object]$Resource,
    [Parameter(Mandatory = $true)][string]$RunSuffix,
    [Parameter(Mandatory = $true)][DateTime]$Deadline
  )

  $arguments = @(
    'volume', 'create',
    '--label', 'talenro.c12.managed=true',
    '--label', "talenro.c12.run=$RunSuffix",
    '--label', 'talenro.c12.profile=authority-v7-pitr',
    '--label', "talenro.c12.role=$([string]$Resource.Role)",
    '--label', "talenro.c12.nonce-digest=$([string]$Resource.NonceDigest)",
    [string]$Resource.Name
  )
  Register-C12DockerIntent -Resource $Resource
  try { $result = Invoke-C12Docker -Arguments $arguments -Stage "create exact PITR volume $([string]$Resource.Role)" -Deadline $Deadline }
  catch { $Resource.RetryState.LastError=$_.Exception.Message; throw }
  if ((Get-C12SingleOutputLine -Result $result -Stage "create exact PITR volume $([string]$Resource.Role)") -cne [string]$Resource.Name) {
    throw 'PITR volume creation did not return its exact name'
  }
  $inspect = Invoke-C12Docker -Arguments @(
    'volume', 'inspect', '--format',
    '{{.Name}}|{{.Driver}}|{{ index .Labels `talenro.c12.managed` }}|{{ index .Labels `talenro.c12.run` }}|{{ index .Labels `talenro.c12.profile` }}|{{ index .Labels `talenro.c12.role` }}|{{ index .Labels `talenro.c12.nonce-digest` }}',
    [string]$Resource.Name
  ) -Stage "inspect exact PITR volume $([string]$Resource.Role)" -Deadline $Deadline
  $identity = Get-C12SingleOutputLine -Result $inspect -Stage "inspect exact PITR volume $([string]$Resource.Role)"
  $parts = $identity -split '\|', 7
  if ($parts.Count -ne 7 -or $parts[0] -cne [string]$Resource.Name -or $parts[1] -cne 'local' -or
      $parts[2] -cne 'true' -or $parts[3] -cne $RunSuffix -or $parts[4] -cne 'authority-v7-pitr' -or
      $parts[5] -cne [string]$Resource.Role -or $parts[6] -cne [string]$Resource.NonceDigest) {
    throw 'PITR volume inspection rejected its exact identity'
  }
  Confirm-C12DockerActual -Resource $Resource
}

function Remove-C12PITRVolume {
  param(
    [Parameter(Mandatory = $true)][object]$Resource,
    [Parameter(Mandatory = $true)][string]$RunSuffix,
    [Parameter(Mandatory = $true)][DateTime]$Deadline
  )

  if (-not [bool]$Resource.CreateAttempted -and [string]$Resource.Phase -ceq 'NeverAttempted') { return }
  $inspectArguments = @(
    'volume', 'inspect', '--format',
    '{{.Name}}|{{.Driver}}|{{ index .Labels `talenro.c12.managed` }}|{{ index .Labels `talenro.c12.run` }}|{{ index .Labels `talenro.c12.profile` }}|{{ index .Labels `talenro.c12.role` }}|{{ index .Labels `talenro.c12.nonce-digest` }}',
    [string]$Resource.Name
  )
  $inspect = Invoke-C12Docker -Arguments $inspectArguments -Stage "re-inspect exact PITR volume $([string]$Resource.Role)" -Timeout ([TimeSpan]::FromSeconds(5)) -Deadline $Deadline -AllowFailure
  if ($inspect.ExitCode -ne 0) {
    if ([string]$Resource.Phase -ceq 'CleanIntent') { Append-C12DockerRecord -Resource $Resource -Event 'DOCKER_CLEAN_RESULT'; $Resource.Phase='Removed' }
    else { $Resource.Phase = 'Absent'; Append-C12DockerRecord -Resource $Resource -Event 'DOCKER_NOT_FOUND' }
    return
  }
  $identity = Get-C12SingleOutputLine -Result $inspect -Stage "re-inspect exact PITR volume $([string]$Resource.Role)"
  $parts = $identity -split '\|', 7
  if ($parts.Count -ne 7 -or $parts[0] -cne [string]$Resource.Name) { throw 'refusing PITR volume cleanup after exact identity mismatch' }
  if ($parts[1] -cne 'local') { throw 'refusing PITR volume cleanup after driver identity mismatch' }
  if ($parts[2] -cne 'true' -or $parts[4] -cne 'authority-v7-pitr') { throw 'refusing PITR volume cleanup after label identity mismatch' }
  if ($parts[3] -cne $RunSuffix -or $parts[5] -cne [string]$Resource.Role -or $parts[6] -cne [string]$Resource.NonceDigest) { throw 'refusing PITR volume cleanup after exact identity mismatch' }
  if ([string]$Resource.Phase -ceq 'CreateAttempted') { Confirm-C12DockerActual -Resource $Resource }
  $Resource.Phase='CleanIntent'; $Resource.RetryState.Attempts++
  Append-C12DockerRecord -Resource $Resource -Event 'DOCKER_CLEAN_INTENT'
  try { $null = Invoke-C12Docker -Arguments @('volume', 'rm', [string]$Resource.Name) -Stage "remove exact PITR volume $([string]$Resource.Role)" -Timeout ([TimeSpan]::FromSeconds(8)) -Deadline $Deadline }
  catch { $Resource.RetryState.LastError=$_.Exception.Message; throw }
  $absence = Invoke-C12Docker -Arguments @('volume', 'inspect', '--format', '{{.Name}}', [string]$Resource.Name) -Stage "verify exact PITR volume $([string]$Resource.Role) cleanup" -Timeout ([TimeSpan]::FromSeconds(5)) -Deadline $Deadline -AllowFailure
  if ($absence.ExitCode -eq 0) {
    throw 'PITR volume remains after exact cleanup'
  }
  $Resource.Phase='Removed'; $Resource.RetryState.LastError=''
  Append-C12DockerRecord -Resource $Resource -Event 'DOCKER_CLEAN_RESULT'
}

function Remove-C12PITRBaseResources {
  param(
    [Parameter(Mandatory = $true)][AllowEmptyCollection()][object[]]$Resources,
    [Parameter(Mandatory = $true)][AllowEmptyCollection()][object[]]$Volumes,
    [Parameter(Mandatory = $true)][string]$RunSuffix,
    [Parameter(Mandatory = $true)][DateTime]$Deadline
  )

  $cleanupErrors=New-Object 'System.Collections.Generic.List[string]'
  $captured = @($Resources | Where-Object { -not [string]::IsNullOrEmpty([string]$_.ID) })
  foreach($resource in $captured){try{Remove-C12Container -Resource $resource -RunSuffix $RunSuffix -Deadline $Deadline}catch{$cleanupErrors.Add($_.Exception.Message)}}
  $captured=@()
  if ($captured.Count -ne 0) {
    $format = '{{.Id}}|{{.Name}}|{{ index .Config.Labels `talenro.c12.run` }}|{{ index .Config.Labels `talenro.c12.role` }}|{{.Config.Image}}|{{.Image}}|{{ index .Config.Labels `talenro.c12.managed` }}|{{ index .Config.Labels `talenro.c12.profile` }}|{{ index .Config.Labels `talenro.c12.nonce-digest` }}'
    $arguments = @('container', 'inspect', '--format', $format) + @($captured | ForEach-Object { [string]$_.ID })
    $inspect = Invoke-C12Docker -Arguments $arguments -Stage 'batch re-inspect exact PITR dependency containers' -Timeout ([TimeSpan]::FromSeconds(8)) -Deadline $Deadline
    $lines = @($inspect.Output | Where-Object { -not [string]::IsNullOrWhiteSpace([string]$_) })
    if ($lines.Count -ne $captured.Count) {
      throw 'batch PITR dependency inspection returned an unexpected identity count'
    }
    $identities = @{}
    foreach ($line in $lines) {
      $parts = ([string]$line) -split '\|', 9
      if ($parts.Count -ne 9 -or $identities.ContainsKey($parts[0])) {
        throw 'batch PITR dependency inspection returned malformed or duplicate identity'
      }
      $identities[$parts[0]] = $parts
    }
    foreach ($resource in $captured) {
      $id = [string]$resource.ID
      if (-not $identities.ContainsKey($id)) {
        throw "batch PITR dependency inspection omitted $($resource.Kind)"
      }
      $parts = $identities[$id]
      if ($parts[1] -cne "/$($resource.Name)" -or $parts[2] -cne $RunSuffix -or
          $parts[3] -cne [string]$resource.Kind -or $parts[4] -cne [string]$resource.ImageRef -or
          $parts[5] -cne [string]$resource.ImageID) {
        throw "refusing batch PITR cleanup for $($resource.Kind): captured identity mismatch"
      }
      if ([bool]$resource.PITR -and ($parts[6] -cne 'true' -or $parts[7] -cne 'authority-v7-pitr' -or
          $parts[8] -cne [string]$resource.NonceDigest -or $parts[3] -cne 'postgres')) {
        throw 'refusing batch PITR primary cleanup: closed labels mismatch'
      }
    }
    $ids = @($captured | ForEach-Object { [string]$_.ID })
    $null = Invoke-C12Docker -Arguments (@('container', 'rm', '--force') + $ids) -Stage 'batch remove exact PITR dependency containers' -Timeout ([TimeSpan]::FromSeconds(12)) -Deadline $Deadline
    $absence = Invoke-C12Docker -Arguments (@('container', 'inspect', '--format', '{{.Id}}') + $ids) -Stage 'batch verify PITR dependency cleanup' -Timeout ([TimeSpan]::FromSeconds(8)) -Deadline $Deadline -AllowFailure
    if (@($absence.Output | Where-Object { [string]$_ -match '^[0-9a-f]{64}$' }).Count -ne 0) {
      throw 'a PITR dependency container remains after exact batch cleanup'
    }
  }

  foreach ($volume in $Volumes) { try { Remove-C12PITRVolume -Resource $volume -RunSuffix $RunSuffix -Deadline $Deadline } catch { $cleanupErrors.Add($_.Exception.Message) } }
  if($cleanupErrors.Count-ne 0){throw ('Docker cleanup retained retry state: '+($cleanupErrors -join '; '))}
}

function Initialize-C12PITRPrimary {
  param(
    [Parameter(Mandatory = $true)][object]$Primary,
    [Parameter(Mandatory = $true)][string]$DatabaseName,
    [Parameter(Mandatory = $true)][object]$RunRoot,
    [Parameter(Mandatory = $true)][DateTime]$Deadline
  )

  $null = Invoke-C12Docker -Arguments @('container', 'exec', [string]$Primary.ID, 'chown', '-R', 'postgres:postgres', '/archive', '/basebackup') -Stage 'prepare PITR archive and base-backup ownership' -Timeout ([TimeSpan]::FromSeconds(30)) -Deadline $Deadline
  $tlsSource = Join-Path ([string]$RunRoot.Root) 'tlsgen.go'
  $tlsCertificate = Join-Path ([string]$RunRoot.Root) 'server.crt'
  $tlsKey = Join-Path ([string]$RunRoot.Root) 'server.key'
  $source = @'
package main
import (
  "crypto/rand"
  "crypto/rsa"
  "crypto/sha256"
  "crypto/x509"
  "crypto/x509/pkix"
  "encoding/hex"
  "encoding/pem"
  "fmt"
  "math/big"
  "net"
  "os"
  "time"
)
func main() {
  if len(os.Args) != 3 { panic("closed TLS paths required") }
  key, err := rsa.GenerateKey(rand.Reader, 2048); if err != nil { panic(err) }
  template := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName:"localhost"}, NotBefore: time.Now().Add(-time.Minute), NotAfter: time.Now().Add(48*time.Hour), DNSNames: []string{"localhost"}, IPAddresses: []net.IP{net.ParseIP("127.0.0.1")}, KeyUsage: x509.KeyUsageDigitalSignature|x509.KeyUsageKeyEncipherment, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, BasicConstraintsValid:true}
  der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key); if err != nil { panic(err) }
  keyDER, err := x509.MarshalPKCS8PrivateKey(key); if err != nil { panic(err) }
  cert := pem.EncodeToMemory(&pem.Block{Type:"CERTIFICATE", Bytes:der})
  private := pem.EncodeToMemory(&pem.Block{Type:"PRIVATE KEY", Bytes:keyDER})
  if err := os.WriteFile(os.Args[1], cert, 0600); err != nil { panic(err) }
  if err := os.WriteFile(os.Args[2], private, 0600); err != nil { panic(err) }
  digest := sha256.Sum256(der); fmt.Println(hex.EncodeToString(digest[:]))
}
'@
  $null = Append-C12ControllerOwnershipRecord -State $RunRoot.ControllerWAL -Event 'INTENT' -PayloadJSON (New-C12PITRLeafEventPayload -Entry (Get-C12DirectLeafEntry -Ledger $RunRoot.Ledger -Name 'tlsgen.go') -Event 'INTENT')
  [IO.File]::WriteAllText($tlsSource, $source, (New-Object Text.UTF8Encoding($false)))
  $tlsSourceEntry = Bind-C12PITRLeaf -RunRoot $RunRoot -Name 'tlsgen.go'
  $null = Append-C12PITRLeafActual -RunRoot $RunRoot -Entry $tlsSourceEntry
  $null = Append-C12ControllerOwnershipRecord -State $RunRoot.ControllerWAL -Event 'INTENT' -PayloadJSON (New-C12PITRLeafEventPayload -Entry (Get-C12DirectLeafEntry -Ledger $RunRoot.Ledger -Name 'server.crt') -Event 'INTENT')
  $null = Append-C12ControllerOwnershipRecord -State $RunRoot.ControllerWAL -Event 'INTENT' -PayloadJSON (New-C12PITRLeafEventPayload -Entry (Get-C12DirectLeafEntry -Ledger $RunRoot.Ledger -Name 'server.key') -Event 'INTENT')
  $tlsResult = Invoke-C12Go -Arguments @('run', $tlsSource, $tlsCertificate, $tlsKey) -Stage 'create PITR primary TLS identity' -Timeout ([TimeSpan]::FromSeconds(30)) -Deadline $Deadline
  $certificateEntry = Bind-C12PITRLeaf -RunRoot $RunRoot -Name 'server.crt'
  $keyEntry = Bind-C12PITRLeaf -RunRoot $RunRoot -Name 'server.key'
  $null = Append-C12PITRLeafActual -RunRoot $RunRoot -Entry $certificateEntry
  $null = Append-C12PITRLeafActual -RunRoot $RunRoot -Entry $keyEntry
  $tlsPublicDigest = Get-C12SingleOutputLine -Result $tlsResult -Stage 'create PITR primary TLS identity'
  if ($tlsPublicDigest -notmatch '^[0-9a-f]{64}$') {
    throw 'PITR TLS public digest is malformed'
  }
  $null = Invoke-C12Docker -Arguments @('container', 'cp', $tlsCertificate, "$([string]$Primary.ID):/var/lib/postgresql/18/docker/server.crt") -Stage 'install PITR TLS certificate' -Timeout ([TimeSpan]::FromSeconds(30)) -Deadline $Deadline
  $null = Invoke-C12Docker -Arguments @('container', 'cp', $tlsKey, "$([string]$Primary.ID):/var/lib/postgresql/18/docker/server.key") -Stage 'install PITR TLS key' -Timeout ([TimeSpan]::FromSeconds(30)) -Deadline $Deadline
  $null = Invoke-C12Docker -Arguments @('container', 'exec', [string]$Primary.ID, 'chown', 'postgres:postgres', '/var/lib/postgresql/18/docker/server.crt', '/var/lib/postgresql/18/docker/server.key') -Stage 'bind PITR TLS file ownership' -Timeout ([TimeSpan]::FromSeconds(30)) -Deadline $Deadline
  $null = Invoke-C12Docker -Arguments @('container', 'exec', [string]$Primary.ID, 'chmod', '600', '/var/lib/postgresql/18/docker/server.key') -Stage 'restrict PITR TLS key' -Timeout ([TimeSpan]::FromSeconds(30)) -Deadline $Deadline
  foreach ($setting in @("ssl = 'on'", "ssl_cert_file = 'server.crt'", "ssl_key_file = 'server.key'")) {
    $null = Invoke-C12Docker -Arguments @('container', 'exec', [string]$Primary.ID, 'psql', '-U', 'talenro', '-d', $DatabaseName, '-v', 'ON_ERROR_STOP=1', '-c', "ALTER SYSTEM SET $setting") -Stage 'configure PITR TLS setting' -Timeout ([TimeSpan]::FromSeconds(30)) -Deadline $Deadline
  }
  $null = Invoke-C12Docker -Arguments @('container', 'restart', '--time', '2', [string]$Primary.ID) -Stage 'restart PITR primary with TLS' -Timeout ([TimeSpan]::FromSeconds(20)) -Deadline $Deadline
  Set-C12MappedPort -Resource $Primary
  return $tlsPublicDigest
}

function Register-C12PITRLeaf {
  param([Parameter(Mandatory)][object]$Ledger, [Parameter(Mandatory)][string]$Name)
  $entry = [pscustomobject]@{
    Name = $Name; Kind = 'exact_file'; Expected = $true; CreateAttempted = $false
    Identity = ''; NumberOfLinks = [UInt32]0; RefCount = 0; Lifecycle = 'NeverAttempted'
    Reparse = $false; Owner = ''; DACL = ''; DACLHash = ''; LastCleanupError = ''; Ownership = $null; CleanupHandle = $null
  }
  [void]$Ledger.Add($entry)
}

function Start-C12PITRLeafCreation {
  param([Parameter(Mandatory)][object]$RunRoot, [Parameter(Mandatory)][string]$Name)
  $entry = Get-C12DirectLeafEntry -Ledger $RunRoot.Ledger -Name $Name
  if ([string]$entry.Lifecycle -cne 'NeverAttempted') { throw "PITR direct leaf $Name creation phase is not NeverAttempted" }
  Set-C12DirectLeafLifecycle -Entry $entry -Lifecycle 'CreateAttempted'
  $entry.CreateAttempted = $true
}

function Bind-C12PITRLeaf {
  param([Parameter(Mandatory)][object]$RunRoot, [Parameter(Mandatory)][string]$Name)
  $RunRoot.Ownership.VerifyExactPath()
  $entry = Get-C12DirectLeafEntry -Ledger $RunRoot.Ledger -Name $Name
  if ([string]$entry.Lifecycle -cne 'CreateAttempted') { throw 'PITR direct leaf bind is outside CreateAttempted' }
  $path = Join-Path ([string]$RunRoot.Root) $Name
  $observed = [C12SealedExecutable]::InspectLeaf($path)
  try {
    if ([bool]$observed.Reparse -or [UInt32]$observed.NumberOfLinks -ne 1) { throw 'PITR direct leaf identity has a reparse or hard-link splice' }
    $entry.Identity = "$([UInt32]$observed.VolumeSerialNumber):$([UInt64]$observed.FileIndex)"
    $entry.NumberOfLinks = [UInt32]$observed.NumberOfLinks
    $entry.Reparse = [bool]$observed.Reparse
    $entry.Owner = [string]$observed.Owner
    $entry.DACL = [string]$observed.DACL
    $entry.DACLHash = Get-C12SHA256Hex -Bytes ([Text.Encoding]::UTF8.GetBytes([string]$observed.DACL))
    $entry.RefCount = 1
    Set-C12DirectLeafLifecycle -Entry $entry -Lifecycle 'Bound'
  }
  finally { $observed.Dispose() }
  return $entry
}

function New-C12PITRLeafEventPayload {
  param([Parameter(Mandatory)][object]$Entry,[Parameter(Mandatory)][ValidateSet('INTENT','ACTUAL','NOT_FOUND','CLEAN_INTENT','CLEAN_RESULT')][string]$Event)
  $resource = [string]$Entry.Name
  switch ($Event) {
    'INTENT' { return '{"resource":"' + $resource + '","lifecycle":"CreateAttempted"}' }
    'ACTUAL' { return '{"resource":"' + $resource + '","lifecycle":"Bound","identity":"' + [string]$Entry.Identity + '","links":' + [UInt32]$Entry.NumberOfLinks + ',"kind":"exact_file","reparse":false,"owner":"' + [string]$Entry.Owner + '","dacl_digest":"' + [string]$Entry.DACLHash + '"}' }
    'NOT_FOUND' { return '{"resource":"' + $resource + '","lifecycle":"Absent"}' }
    'CLEAN_INTENT' { return '{"resource":"' + $resource + '","lifecycle":"CleanIntent","identity":"' + [string]$Entry.Identity + '"}' }
    'CLEAN_RESULT' { return '{"resource":"' + $resource + '","lifecycle":"Absent","identity":"' + [string]$Entry.Identity + '"}' }
  }
}

function Get-C12ControllerWALPayloadFields {
  param([Parameter(Mandatory)][string]$Event,[Parameter(Mandatory)][string]$PayloadJSON)
  $pattern = switch ($Event) {
    'INTENT' { '^\{"resource":"(?<resource>[a-z0-9.-]+)","lifecycle":"CreateAttempted"\}$' }
    'ACTUAL' { '^\{"resource":"(?<resource>[a-z0-9.-]+)","lifecycle":"Bound","identity":"(?<identity>[0-9]+:[0-9]+)","links":(?<links>1),"kind":"(?<kind>exact_file)","reparse":(?<reparse>false),"owner":"(?<owner>S-[0-9-]+)","dacl_digest":"(?<dacl>[0-9a-f]{64})"\}$' }
    'NOT_FOUND' { '^\{"resource":"(?<resource>[a-z0-9.-]+)","lifecycle":"Absent"\}$' }
    'CLEAN_INTENT' { '^\{"resource":"(?<resource>[a-z0-9.-]+)","lifecycle":"CleanIntent","identity":"(?<identity>[0-9]*:[0-9]*|)"\}$' }
    'CLEAN_RESULT' { '^\{"resource":"(?<resource>[a-z0-9.-]+)","lifecycle":"Absent","identity":"(?<identity>[0-9]*:[0-9]*|)"\}$' }
    default { throw 'controller WAL payload event is outside the closed schema' }
  }
  $match = [Text.RegularExpressions.Regex]::Match($PayloadJSON,$pattern,[Text.RegularExpressions.RegexOptions]::CultureInvariant)
  if (-not $match.Success) { throw "controller WAL $Event payload violates its exact field schema/order/type" }
  $resource = $match.Groups['resource'].Value
  if ($resource -cnotin @('controller-ownership-v1.wal','ownership.wal','tlsgen.go','server.crt','server.key')) { throw 'controller WAL event names a resource outside the five-leaf registry' }
  return [pscustomobject]@{ Resource=$resource; Identity=$match.Groups['identity'].Value; Links=$match.Groups['links'].Value; Kind=$match.Groups['kind'].Value; Reparse=$match.Groups['reparse'].Value; Owner=$match.Groups['owner'].Value; DACLHash=$match.Groups['dacl'].Value }
}

function New-C12DockerRegistryPayload {
  param([Parameter(Mandatory)][AllowEmptyCollection()][object[]]$Resources,[Parameter(Mandatory)][AllowEmptyCollection()][object[]]$Volumes,[Parameter(Mandatory)][string]$RunSuffix,[Parameter(Mandatory)][string]$NonceDigest,[Parameter(Mandatory)][object]$EndpointReceipt)
  $entries = New-Object 'System.Collections.Generic.List[string]'
  foreach ($resource in @($Resources | Sort-Object Name)) {
    foreach ($value in @([string]$resource.Name,[string]$resource.Kind,[string]$resource.ImageRef,[string]$resource.ImageID)) { if ($value -match '["\\\x00-\x1f]') { throw 'Docker registry container field is noncanonical' } }
    $nonce = if ($resource.PSObject.Properties.Name -contains 'NonceDigest') { [string]$resource.NonceDigest } else { '' }
    $entries.Add('{"kind":"container","name":"'+[string]$resource.Name+'","role":"'+[string]$resource.Kind+'","image_ref":"'+[string]$resource.ImageRef+'","image_id":"'+[string]$resource.ImageID+'","driver":"","nonce_digest":"'+$nonce+'"}')
  }
  foreach ($volume in @($Volumes | Sort-Object Name)) {
    foreach ($value in @([string]$volume.Name,[string]$volume.Role)) { if ($value -match '["\\\x00-\x1f]') { throw 'Docker registry volume field is noncanonical' } }
    $entries.Add('{"kind":"volume","name":"'+[string]$volume.Name+'","role":"'+[string]$volume.Role+'","image_ref":"","image_id":"","driver":"local","nonce_digest":"'+[string]$volume.NonceDigest+'"}')
  }
  $registry='['+($entries -join ',')+']'
  $authority='{"schema":"talenro.c12.docker-registry.v1","run":"'+$RunSuffix+'","profile":"authority-v7-pitr","nonce_digest":"'+$NonceDigest+'","endpoint_digest":"'+[string]$EndpointReceipt.Digest+'","cli_path":"'+([string]$EndpointReceipt.ExecutableReceipt.Path).Replace('\','/')+'","cli_length":'+[long]$EndpointReceipt.ExecutableReceipt.Length+',"cli_digest":"'+[string]$EndpointReceipt.ExecutableReceipt.Digest+'","resources":'+$registry+'}'
  $digest=Get-C12DomainSHA256 -Domain 'talenro.c12.docker-registry.v1' -Bytes ([Text.Encoding]::UTF8.GetBytes($authority))
  return $authority.Substring(0,$authority.Length-1)+',"registry_digest":"'+$digest+'"}'
}

function New-C12DockerEventPayload {
  param([Parameter(Mandatory)][object]$Resource,[Parameter(Mandatory)][ValidateSet('DOCKER_INTENT','DOCKER_ACTUAL','DOCKER_NOT_FOUND','DOCKER_CLEAN_INTENT','DOCKER_CLEAN_RESULT')][string]$Event)
  $kind=if($Resource.PSObject.Properties.Name -contains 'Driver'){'volume'}elseif($Resource.PSObject.Properties.Name -contains 'Role' -and -not ($Resource.PSObject.Properties.Name -contains 'ContainerPort')){'volume'}else{'container'}
  $role=if($Resource.PSObject.Properties.Name -contains 'Role'){[string]$Resource.Role}else{[string]$Resource.Kind}
  $imageRef=if($Resource.PSObject.Properties.Name -contains 'ImageRef'){[string]$Resource.ImageRef}else{''}; $imageID=if($Resource.PSObject.Properties.Name -contains 'ImageID'){[string]$Resource.ImageID}else{''}
  $driver=if($kind-ceq'volume'){'local'}else{''}; $nonce=if($Resource.PSObject.Properties.Name -contains 'NonceDigest'){[string]$Resource.NonceDigest}else{''}
  $objectID=if($kind-ceq'volume'){[string]$Resource.Name}else{[string]$Resource.ID}
  $base='{"schema":"talenro.c12.docker-event.v1","kind":"'+$kind+'","name":"'+[string]$Resource.Name+'","role":"'+$role+'","image_ref":"'+$imageRef+'","image_id":"'+$imageID+'","driver":"'+$driver+'","nonce_digest":"'+$nonce+'","lifecycle":"'
  switch($Event){'DOCKER_INTENT'{return $base+'CreateAttempted"}'};'DOCKER_ACTUAL'{return $base+'Verified","object_id":"'+$objectID+'"}'};'DOCKER_NOT_FOUND'{return $base+'Absent"}'};'DOCKER_CLEAN_INTENT'{return $base+'CleanIntent","object_id":"'+$objectID+'"}'};'DOCKER_CLEAN_RESULT'{return $base+'Absent","object_id":"'+$objectID+'"}'}}
}

function Append-C12DockerRecord {
  param([Parameter(Mandatory)][object]$Resource,[Parameter(Mandatory)][string]$Event)
  if (-not ($Resource.PSObject.Properties.Name -contains 'ControllerWAL') -or $null -eq $Resource.ControllerWAL) { return }
  $payload=New-C12DockerEventPayload -Resource $Resource -Event $Event
  $null=Append-C12ControllerOwnershipRecord -State $Resource.ControllerWAL -Event $Event -PayloadJSON $payload
}

function Assert-C12DockerEventAuthorization {
  param([Parameter(Mandatory)][object]$State,[Parameter(Mandatory)][string]$PayloadJSON)
  $match=[regex]::Match($PayloadJSON,'^\{"schema":"talenro\.c12\.docker-event\.v1","kind":"(?<kind>container|volume)","name":"(?<name>talenro-c12-[a-z0-9-]+)","role":"(?<role>[a-z0-9-]+)","image_ref":"(?<ref>[a-z0-9.:/-]*)","image_id":"(?<image>(?:sha256:[0-9a-f]{64})?)","driver":"(?<driver>(?:local)?)","nonce_digest":"(?<nonce>[0-9a-f]*)","lifecycle":"(?<life>CreateAttempted|Verified|Absent|CleanIntent)"(?:,"object_id":"(?<object>[a-z0-9.-]+)")?\}$',[Text.RegularExpressions.RegexOptions]::CultureInvariant)
  if(-not $match.Success){throw 'controller WAL Docker event violates its closed schema'}
  $entry='{"kind":"'+$match.Groups['kind'].Value+'","name":"'+$match.Groups['name'].Value+'","role":"'+$match.Groups['role'].Value+'","image_ref":"'+$match.Groups['ref'].Value+'","image_id":"'+$match.Groups['image'].Value+'","driver":"'+$match.Groups['driver'].Value+'","nonce_digest":"'+$match.Groups['nonce'].Value+'"}'
  if(-not ([string]$State.DockerRegistryPayload).Contains($entry)){throw 'controller WAL Docker event identity is not exactly authorized by registry'}
  return [string]$match.Groups['name'].Value
}

function Append-C12PITRLeafActual {
  param([Parameter(Mandatory)][object]$RunRoot,[Parameter(Mandatory)][object]$Entry)
  $payload = New-C12PITRLeafEventPayload -Entry $Entry -Event 'ACTUAL'
  return Append-C12ControllerOwnershipRecord -State $RunRoot.ControllerWAL -Event 'ACTUAL' -PayloadJSON $payload
}

function Get-C12DomainSHA256 {
  param([Parameter(Mandatory)][string]$Domain, [Parameter(Mandatory)][byte[]]$Bytes)
  $domainBytes = [Text.Encoding]::UTF8.GetBytes($Domain)
  $input = New-Object byte[] ($domainBytes.Length + 1 + $Bytes.Length)
  [Array]::Copy($domainBytes, 0, $input, 0, $domainBytes.Length)
  [Array]::Copy($Bytes, 0, $input, $domainBytes.Length + 1, $Bytes.Length)
  $algorithm = [Security.Cryptography.SHA256]::Create()
  try { return (($algorithm.ComputeHash($input) | ForEach-Object { $_.ToString('x2') }) -join '') }
  finally { $algorithm.Dispose(); [Array]::Clear($input, 0, $input.Length) }
}

function Open-C12ControllerOwnershipWAL {
  param(
    [Parameter(Mandatory)][object]$RunRoot,
    [Parameter(Mandatory)][string]$RunSuffix,
    [Parameter(Mandatory)][string]$NonceDigest,
    [Parameter(Mandatory)][string]$DockerExecutableDigest,
    [Parameter(Mandatory)][string]$DockerEndpointIdentityDigest,
    [Parameter(Mandatory)][byte[]]$HMACKey,
    [DateTime]$Timestamp = [DateTime]::UtcNow
  )
  Write-Verbose 'talenro-c12-authority-pitr-controller-ownership-wal/v1 controller-ownership-v1.wal WriteThrough FlushFileBuffers BOOTSTRAP INTENT ACTUAL NOT_FOUND CLEAN_INTENT CLEAN_RESULT talenro.c12.controller-wal.registry.v1 talenro.c12.controller-wal.payload.v1 talenro.c12.controller-wal.record.v1 talenro.c12.controller-wal.hmac.v1 record_digest hmac_sha256 yyyy-MM-ddTHH:mm:ss.fffffffZ previous_record_digest payload_digest'
  $leafName = [string]$RunRoot.Ledger[0].Name
  $path = Join-Path ([string]$RunRoot.Root) $leafName
  Start-C12PITRLeafCreation -RunRoot $RunRoot -Name $leafName
  foreach ($value in @($RunSuffix,$NonceDigest,$DockerExecutableDigest,$DockerEndpointIdentityDigest)) { if ($value -notmatch '^[0-9a-f]+$') { throw 'controller WAL identity field is malformed' } }
  if ($RunSuffix.Length -ne 32 -or $NonceDigest.Length -ne 64 -or $DockerExecutableDigest.Length -ne 64 -or $DockerEndpointIdentityDigest.Length -ne 64 -or $HMACKey.Length -ne 32) { throw 'controller WAL identity field has invalid length' }
  $timestampText = $Timestamp.ToUniversalTime().ToString('yyyy-MM-ddTHH:mm:ss.fffffffZ', [Globalization.CultureInfo]::InvariantCulture)
  $registryJSON = '["controller-ownership-v1.wal","ownership.wal","tlsgen.go","server.crt","server.key"]'
  $registryDigest = Get-C12DomainSHA256 -Domain 'talenro.c12.controller-wal.registry.v1' -Bytes ([Text.Encoding]::UTF8.GetBytes($registryJSON))
  $payload = '{"registry":' + $registryJSON + ',"registry_digest":"' + $registryDigest + '"}'
  $payloadDigest = Get-C12DomainSHA256 -Domain 'talenro.c12.controller-wal.payload.v1' -Bytes ([Text.Encoding]::UTF8.GetBytes($payload))
  $base = '{"schema":"talenro-c12-authority-pitr-controller-ownership-wal/v1","version":1,"run":"' + $RunSuffix + '","profile":"authority-v7-pitr","nonce_digest":"' + $NonceDigest + '","docker_executable_digest":"' + $DockerExecutableDigest + '","docker_endpoint_identity_digest":"' + $DockerEndpointIdentityDigest + '","sequence":0,"previous_record_digest":null,"event":"BOOTSTRAP","timestamp_utc":"' + $timestampText + '","payload":' + $payload + ',"payload_digest":"' + $payloadDigest + '"}'
  $recordDigest = Get-C12DomainSHA256 -Domain 'talenro.c12.controller-wal.record.v1' -Bytes ([Text.Encoding]::UTF8.GetBytes($base))
  $hmac = [Security.Cryptography.HMACSHA256]::new($HMACKey)
  try { $domain = [Text.Encoding]::UTF8.GetBytes('talenro.c12.controller-wal.hmac.v1'); $raw = ConvertFrom-C12Hex $recordDigest; $macInput = New-Object byte[] ($domain.Length+1+$raw.Length); [Array]::Copy($domain,0,$macInput,0,$domain.Length); [Array]::Copy($raw,0,$macInput,$domain.Length+1,$raw.Length); $mac = (($hmac.ComputeHash($macInput)|ForEach-Object{$_.ToString('x2')})-join '') } finally { $hmac.Dispose() }
  $literal = $base.Substring(0,$base.Length-1) + ',"record_digest":"' + $recordDigest + '","hmac_sha256":"' + $mac + '"}'
  $stream = [IO.FileStream]::new($path, [IO.FileMode]::CreateNew, [IO.FileAccess]::ReadWrite, [IO.FileShare]::Read, 4096, [IO.FileOptions]::WriteThrough)
  try {
    $bytes = [Text.Encoding]::UTF8.GetBytes($literal + "`n")
    $stream.Write($bytes, 0, $bytes.Length)
    $stream.Flush($true) # FlushFileBuffers
  }
  finally { $stream.Dispose() }
  $null = Bind-C12PITRLeaf -RunRoot $RunRoot -Name $leafName
  return [pscustomobject]@{
    Path = $path; WALHandle = $null; Key = [byte[]]$HMACKey.Clone(); Sequence = [UInt64]0
    RunSuffix = $RunSuffix; Profile = 'authority-v7-pitr'; NonceDigest = $NonceDigest; DockerExecutableDigest = $DockerExecutableDigest; DockerEndpointIdentityDigest = $DockerEndpointIdentityDigest
    PreviousRecordDigest = $recordDigest
    Lifecycle = 'Bound'; RetryState = 'Verified'; LastError = ''
    DockerRegistryPayload = ''; DockerRegistryVerified = $false
  }
}

function Read-C12ControllerOwnershipWAL {
  param([Parameter(Mandatory)][object]$State)
  $stream = [IO.FileStream]::new([string]$State.Path, [IO.FileMode]::Open, [IO.FileAccess]::Read, [IO.FileShare]::ReadWrite -bor [IO.FileShare]::Delete)
  try {
    if ($stream.Length -gt ($script:c12PITRWALMaximumLineBytes * $script:c12PITRWALMaximumRecords)) { throw 'controller WAL total byte bound exceeded' }
    $bytes = New-Object byte[] ([int]$stream.Length)
    $offset = 0
    while ($offset -lt $bytes.Length) { $read = $stream.Read($bytes, $offset, $bytes.Length - $offset); if ($read -le 0) { throw 'controller WAL truncated during exact read' }; $offset += $read }
  }
  finally { $stream.Dispose() }
  if ($bytes.Length -eq 0 -or ($bytes -contains 13) -or
      ($bytes.Length -ge 3 -and $bytes[0] -eq 0xef -and $bytes[1] -eq 0xbb -and $bytes[2] -eq 0xbf)) {
    throw 'controller WAL noncanonical exact bytes'
  }
  $State | Add-Member -NotePropertyName TailRepairLength -NotePropertyValue ([Int64]0) -Force
  if ($bytes[$bytes.Length-1] -ne 10) {
    $lastLF = [Array]::LastIndexOf($bytes, [byte]10)
    if ($lastLF -lt 0) { throw 'controller WAL has no authenticated head before its partial tail' }
    $State.TailRepairLength = [Int64]($lastLF + 1)
    $prefix = New-Object byte[] ($lastLF + 1)
    [Array]::Copy($bytes, 0, $prefix, 0, $prefix.Length)
    $bytes = $prefix
  }
  $text = [Text.Encoding]::UTF8.GetString($bytes)
  $lines = @($text.Substring(0, $text.Length - 1).Split("`n"))
  if ($lines.Count -gt $script:c12PITRWALMaximumRecords) { throw 'controller WAL record bound exceeded' }
  return $lines
}

function Verify-C12ControllerOwnershipWAL {
  param([Parameter(Mandatory)][object]$State)
  $lines = @(Read-C12ControllerOwnershipWAL -State $State)
  $previous = $null
  $lastEvent = ''
  $lastPayload = ''
  $resourceStates = @{}
  $resourceReceipts = @{}
  for ($index = 0; $index -lt $lines.Count; $index++) {
    $line = [string]$lines[$index]
    if ([Text.Encoding]::UTF8.GetByteCount($line) + 1 -gt $script:c12PITRWALMaximumLineBytes) { throw 'controller WAL line bound exceeded' }
    $grammar = '^\{"schema":"talenro-c12-authority-pitr-controller-ownership-wal/v1","version":1,"run":"(?<run>[0-9a-f]{32})","profile":"authority-v7-pitr","nonce_digest":"(?<nonce>[0-9a-f]{64})","docker_executable_digest":"(?<docker>[0-9a-f]{64})","docker_endpoint_identity_digest":"(?<endpoint>[0-9a-f]{64})","sequence":(?<sequence>0|[1-9][0-9]*),"previous_record_digest":(?<previous>null|"[0-9a-f]{64}"),"event":"(?<event>[A-Z_]+)","timestamp_utc":"(?<timestamp>[^"]+)","payload":(?<payload>\{.*\}),"payload_digest":"(?<payload_digest>[0-9a-f]{64})","record_digest":"(?<record_digest>[0-9a-f]{64})","hmac_sha256":"(?<hmac>[0-9a-f]{64})"\}$'
    $match = [Text.RegularExpressions.Regex]::Match($line, $grammar, [Text.RegularExpressions.RegexOptions]::CultureInvariant)
    if (-not $match.Success) { throw 'controller WAL record JSON is malformed or truncated' }
    $event = $match.Groups['event'].Value
    if ($match.Groups['run'].Value -cne [string]$State.RunSuffix -or $match.Groups['nonce'].Value -cne [string]$State.NonceDigest -or $match.Groups['docker'].Value -cne [string]$State.DockerExecutableDigest -or $match.Groups['endpoint'].Value -cne [string]$State.DockerEndpointIdentityDigest) { throw 'controller WAL record identity does not match retained State' }
    if ([UInt64]$match.Groups['sequence'].Value -ne [UInt64]$index) { throw 'controller WAL sequence is duplicate or reordered' }
      if ($event -cnotin @('BOOTSTRAP','DOCKER_REGISTRY','INTENT','ACTUAL','NOT_FOUND','CLEAN_INTENT','CLEAN_RESULT','DOCKER_INTENT','DOCKER_ACTUAL','DOCKER_NOT_FOUND','DOCKER_CLEAN_INTENT','DOCKER_CLEAN_RESULT')) { throw 'controller WAL unknown event' }
      if ($index -eq 0) {
        if ($event -cne 'BOOTSTRAP' -or $match.Groups['previous'].Value -cne 'null') { throw 'controller WAL previous_record_digest mismatch' }
        $expectedRegistry = '{"registry":["controller-ownership-v1.wal","ownership.wal","tlsgen.go","server.crt","server.key"],"registry_digest":"0c3acaace57bd065528feb8316f31dcdc1306658cd0c84b067412412d4a38f62"}'
        if ($match.Groups['payload'].Value -cne $expectedRegistry) { throw 'controller WAL BOOTSTRAP payload registry does not match the exact five-leaf RunRoot ledger' }
      }
      elseif ($match.Groups['previous'].Value.Trim('"') -cne $previous) { throw 'controller WAL previous_record_digest mismatch' }
      $timestamp = $match.Groups['timestamp'].Value
      $parsed = [DateTime]::MinValue
      if (-not [DateTime]::TryParseExact($timestamp, 'yyyy-MM-ddTHH:mm:ss.fffffffZ', [Globalization.CultureInfo]::InvariantCulture, [Globalization.DateTimeStyles]::AssumeUniversal -bor [Globalization.DateTimeStyles]::AdjustToUniversal, [ref]$parsed)) { throw 'controller WAL timestamp_utc is noncanonical' }
      $payloadText = $match.Groups['payload'].Value
      if ($index -gt 0 -and $event -ceq 'DOCKER_REGISTRY') {
        if ([string]$State.DockerRegistryPayload -cne $payloadText) { throw 'controller WAL Docker registry authorization mismatch' }
        $State.DockerRegistryVerified=$true
      }
      elseif ($index -gt 0 -and $event.StartsWith('DOCKER_', [StringComparison]::Ordinal)) {
        if (-not [bool]$State.DockerRegistryVerified) { throw 'controller WAL Docker event lacks registry authority' }
        $dockerName=Assert-C12DockerEventAuthorization -State $State -PayloadJSON $payloadText
        $dockerPrior=if($resourceStates.ContainsKey($dockerName)){[string]$resourceStates[$dockerName]}else{''}
        $dockerLegal=switch($event){'DOCKER_INTENT'{$dockerPrior-ceq''};'DOCKER_ACTUAL'{$dockerPrior-ceq'DOCKER_INTENT'};'DOCKER_NOT_FOUND'{$dockerPrior-ceq'DOCKER_INTENT'};'DOCKER_CLEAN_INTENT'{$dockerPrior-cin@('DOCKER_ACTUAL','DOCKER_NOT_FOUND')};'DOCKER_CLEAN_RESULT'{$dockerPrior-ceq'DOCKER_CLEAN_INTENT'};default{$false}}
        if(-not $dockerLegal){throw "controller WAL illegal $dockerPrior -> $event transition for $dockerName"}
        $resourceStates[$dockerName]=$event
      }
      elseif ($index -gt 0) {
        $payloadFields = Get-C12ControllerWALPayloadFields -Event $event -PayloadJSON $payloadText
        $resource = [string]$payloadFields.Resource
        $prior = if ($resourceStates.ContainsKey($resource)) { [string]$resourceStates[$resource] } else { '' }
        $legal = switch ($event) {
          'INTENT' { $prior -ceq '' }
          'ACTUAL' { $prior -ceq 'INTENT' }
          'NOT_FOUND' { $prior -ceq 'INTENT' }
          'CLEAN_INTENT' { $prior -cin @('','ACTUAL','NOT_FOUND') }
          'CLEAN_RESULT' { $prior -ceq 'CLEAN_INTENT' }
          default { $false }
        }
        if (-not $legal) { throw "controller WAL illegal $prior -> $event transition for $resource" }
        $resourceStates[$resource] = $event
        if ($event -ceq 'ACTUAL') { $resourceReceipts[$resource] = $payloadFields }
      }
      $payloadDigest = Get-C12DomainSHA256 -Domain 'talenro.c12.controller-wal.payload.v1' -Bytes ([Text.Encoding]::UTF8.GetBytes($payloadText))
      if ($payloadDigest -cne $match.Groups['payload_digest'].Value) { throw 'controller WAL payload_digest mismatch after payload verification' }
      $recordMarker = ',"record_digest":"'
      $markerIndex = $line.LastIndexOf($recordMarker, [StringComparison]::Ordinal)
      if ($markerIndex -lt 0) { throw 'controller WAL record_digest field is missing' }
      $base = $line.Substring(0, $markerIndex) + '}'
      $recordDigest = Get-C12DomainSHA256 -Domain 'talenro.c12.controller-wal.record.v1' -Bytes ([Text.Encoding]::UTF8.GetBytes($base))
      if ($recordDigest -cne $match.Groups['record_digest'].Value) { throw 'controller WAL record_digest mismatch' }
      $hmac = [Security.Cryptography.HMACSHA256]::new([byte[]]$State.Key)
      try {
        $domain = [Text.Encoding]::UTF8.GetBytes('talenro.c12.controller-wal.hmac.v1')
        $digestBytes = ConvertFrom-C12Hex -Value $recordDigest
        $macInput = New-Object byte[] ($domain.Length + 1 + $digestBytes.Length)
        [Array]::Copy($domain,0,$macInput,0,$domain.Length); [Array]::Copy($digestBytes,0,$macInput,$domain.Length+1,$digestBytes.Length)
        $expectedHMAC = (($hmac.ComputeHash($macInput) | ForEach-Object { $_.ToString('x2') }) -join '')
      }
      finally { $hmac.Dispose() }
      if ($expectedHMAC -cne $match.Groups['hmac'].Value) { throw 'controller WAL hmac_sha256 mismatch' }
      $previous = $recordDigest
      $lastEvent = $event
      $lastPayload = $payloadText
  }
  if ([Int64]$State.TailRepairLength -gt 0) {
    $repair = [IO.FileStream]::new([string]$State.Path,[IO.FileMode]::Open,[IO.FileAccess]::ReadWrite,[IO.FileShare]::Read,4096,[IO.FileOptions]::WriteThrough)
    try { $repair.SetLength([Int64]$State.TailRepairLength); $repair.Flush($true) } finally { $repair.Dispose() }
    $State.TailRepairLength = [Int64]0
  }
  return [pscustomobject]@{ Records = $lines.Count; Head = $previous; LastEvent = $lastEvent; LastPayload = $lastPayload; ResourceStates = $resourceStates; ResourceReceipts = $resourceReceipts; Verified = $true }
}

function Assert-C12ControllerOwnershipWAL { param([Parameter(Mandatory)][object]$State) $null = Verify-C12ControllerOwnershipWAL -State $State }

function Restore-C12PITRRunLedger {
  param([Parameter(Mandatory)][object]$RunRoot)
  $verified = Verify-C12ControllerOwnershipWAL -State $RunRoot.ControllerWAL
  $ledger = New-C12DirectLeafLedger
  foreach ($name in @('controller-ownership-v1.wal','ownership.wal','tlsgen.go','server.crt','server.key')) { Register-C12PITRLeaf -Ledger $ledger -Name $name }
  foreach ($entry in @($ledger)) {
    $name = [string]$entry.Name
    $state = if ($verified.ResourceStates.ContainsKey($name)) { [string]$verified.ResourceStates[$name] } else { '' }
    $receipt = if ($verified.ResourceReceipts.ContainsKey($name)) { $verified.ResourceReceipts[$name] } else { $null }
    switch ($state) {
      'INTENT' { Set-C12DirectLeafLifecycle -Entry $entry -Lifecycle 'CreateAttempted'; $entry.CreateAttempted=$true }
      { $_ -cin @('ACTUAL','CLEAN_INTENT') } {
        if ($null -eq $receipt) { throw "controller WAL recovery lacks creation-time ACTUAL receipt for $name" }
        Set-C12DirectLeafLifecycle -Entry $entry -Lifecycle 'CreateAttempted'; $entry.CreateAttempted=$true
        $entry.Identity=[string]$receipt.Identity; $entry.NumberOfLinks=[UInt32]$receipt.Links; $entry.Reparse=$false; $entry.Owner=[string]$receipt.Owner; $entry.DACLHash=[string]$receipt.DACLHash; $entry.RefCount=1
        Set-C12DirectLeafLifecycle -Entry $entry -Lifecycle 'Bound'
        if ($state -ceq 'CLEAN_INTENT') { Set-C12DirectLeafLifecycle -Entry $entry -Lifecycle 'CleanIntent' }
      }
      { $_ -cin @('NOT_FOUND','CLEAN_RESULT') } { Complete-C12AbsentDirectLeaf -Entry $entry }
    }
  }
  $RunRoot.Ledger = $ledger
  return $ledger
}

function Append-C12ControllerOwnershipRecord {
  param(
    [Parameter(Mandatory)][object]$State,
    [Parameter(Mandatory)][ValidateSet('DOCKER_REGISTRY','INTENT','ACTUAL','NOT_FOUND','CLEAN_INTENT','CLEAN_RESULT','DOCKER_INTENT','DOCKER_ACTUAL','DOCKER_NOT_FOUND','DOCKER_CLEAN_INTENT','DOCKER_CLEAN_RESULT')][string]$Event,
    [Parameter(Mandatory)][string]$PayloadJSON,
    [ValidateSet('','after-write-before-flush','after-flush-before-head')][string]$FailureSeam = ''
  )
  $verified = Verify-C12ControllerOwnershipWAL -State $State
  if ($PayloadJSON -notmatch '^\{[^\r\n]*\}$' -or [Text.Encoding]::UTF8.GetByteCount($PayloadJSON) -gt 4096) { throw 'controller WAL payload is not bounded canonical JSON' }
  if ([string]$verified.LastEvent -ceq $Event -and [string]$verified.LastPayload -ceq $PayloadJSON) {
    $State.Sequence = [UInt64]($verified.Records - 1); $State.PreviousRecordDigest = [string]$verified.Head; $State.RetryState = 'Verified'; $State.LastError = ''
    return $verified
  }
  if ($Event -ceq 'DOCKER_REGISTRY') { if([string]$State.DockerRegistryPayload-cne$PayloadJSON){throw 'controller WAL Docker registry authorization mismatch'}; $payloadFields=[pscustomobject]@{Resource='docker-registry'} }
  elseif ($Event.StartsWith('DOCKER_',[StringComparison]::Ordinal)) { if(-not [bool]$State.DockerRegistryVerified){throw 'controller WAL Docker event lacks verified registry'}; $dockerResource=Assert-C12DockerEventAuthorization -State $State -PayloadJSON $PayloadJSON; $payloadFields=[pscustomobject]@{Resource=$dockerResource} }
  else { $payloadFields = Get-C12ControllerWALPayloadFields -Event $Event -PayloadJSON $PayloadJSON }
  $resource = [string]$payloadFields.Resource
  $prior = if ($verified.ResourceStates.ContainsKey($resource)) { [string]$verified.ResourceStates[$resource] } else { '' }
  $legal = if($Event -ceq 'DOCKER_REGISTRY'){$prior -ceq ''}elseif($Event.StartsWith('DOCKER_',[StringComparison]::Ordinal)){switch($Event){'DOCKER_INTENT'{$prior-ceq''};'DOCKER_ACTUAL'{$prior-ceq'DOCKER_INTENT'};'DOCKER_NOT_FOUND'{$prior-ceq'DOCKER_INTENT'};'DOCKER_CLEAN_INTENT'{$prior-cin@('DOCKER_ACTUAL','DOCKER_NOT_FOUND')};'DOCKER_CLEAN_RESULT'{$prior-ceq'DOCKER_CLEAN_INTENT'};default{$false}}}else{switch ($Event) { 'INTENT' {$prior -ceq ''}; 'ACTUAL' {$prior -ceq 'INTENT'}; 'NOT_FOUND' {$prior -ceq 'INTENT'}; 'CLEAN_INTENT' {$prior -cin @('','ACTUAL','NOT_FOUND')}; 'CLEAN_RESULT' {$prior -ceq 'CLEAN_INTENT'}; default {$false} }}
  if (-not $legal) { throw "controller WAL illegal $prior -> $Event transition for $resource" }
  $sequence = [UInt64]$verified.Records
  $previous = [string]$verified.Head
  $timestamp = [DateTime]::UtcNow.ToString('yyyy-MM-ddTHH:mm:ss.fffffffZ', [Globalization.CultureInfo]::InvariantCulture)
  $payloadDigest = Get-C12DomainSHA256 -Domain 'talenro.c12.controller-wal.payload.v1' -Bytes ([Text.Encoding]::UTF8.GetBytes($PayloadJSON))
  $base = '{"schema":"talenro-c12-authority-pitr-controller-ownership-wal/v1","version":1,"run":"' + [string]$State.RunSuffix + '","profile":"' + [string]$State.Profile + '","nonce_digest":"' + [string]$State.NonceDigest + '","docker_executable_digest":"' + [string]$State.DockerExecutableDigest + '","docker_endpoint_identity_digest":"' + [string]$State.DockerEndpointIdentityDigest + '","sequence":' + $sequence + ',"previous_record_digest":"' + $previous + '","event":"' + $Event + '","timestamp_utc":"' + $timestamp + '","payload":' + $PayloadJSON + ',"payload_digest":"' + $payloadDigest + '"}'
  $recordDigest = Get-C12DomainSHA256 -Domain 'talenro.c12.controller-wal.record.v1' -Bytes ([Text.Encoding]::UTF8.GetBytes($base))
  $hmac = [Security.Cryptography.HMACSHA256]::new([byte[]]$State.Key)
  try {
    $domain = [Text.Encoding]::UTF8.GetBytes('talenro.c12.controller-wal.hmac.v1')
    $digestBytes = ConvertFrom-C12Hex -Value $recordDigest
    $macInput = New-Object byte[] ($domain.Length + 1 + $digestBytes.Length)
    [Array]::Copy($domain,0,$macInput,0,$domain.Length); [Array]::Copy($digestBytes,0,$macInput,$domain.Length+1,$digestBytes.Length)
    $mac = (($hmac.ComputeHash($macInput) | ForEach-Object { $_.ToString('x2') }) -join '')
  }
  finally { $hmac.Dispose() }
  $line = $base.Substring(0, $base.Length - 1) + ',"record_digest":"' + $recordDigest + '","hmac_sha256":"' + $mac + '"}' + "`n"
  $bytes = [Text.Encoding]::UTF8.GetBytes($line)
  if ($bytes.Length -gt $script:c12PITRWALMaximumLineBytes) { throw 'controller WAL append exceeded line bound' }
  $stream = [IO.FileStream]::new([string]$State.Path, [IO.FileMode]::Open, [IO.FileAccess]::ReadWrite, [IO.FileShare]::ReadWrite -bor [IO.FileShare]::Delete, 4096, [IO.FileOptions]::WriteThrough)
  try {
    $stream.Seek(0, [IO.SeekOrigin]::End) | Out-Null
    $stream.Write($bytes, 0, $bytes.Length)
    if ($FailureSeam -ceq 'after-write-before-flush') { throw 'injected controller WAL crash after write before flush' }
    $stream.Flush($true)
    if ($FailureSeam -ceq 'after-flush-before-head') { throw 'injected controller WAL crash after flush before head' }
  }
  finally { $stream.Dispose() }
  $after = Verify-C12ControllerOwnershipWAL -State $State
  $State.Sequence = $sequence
  $State.PreviousRecordDigest = [string]$after.Head
  $State.RetryState = 'Verified'
  $State.LastError = ''
  return $after
}

function Remove-C12PITRRunRoot {
  param([Parameter(Mandatory)][object]$RunRoot, [DateTime]$Deadline = [DateTime]::MaxValue)
  Write-Verbose 'PITR CreationClosed NeverAttempted Remove unknown sibling direct top-level finalizer'
  $RunRoot.CreationClosed = $true
  $registered = [Collections.Generic.HashSet[string]]::new([StringComparer]::Ordinal)
  foreach ($entry in @($RunRoot.Ledger)) { if (-not $registered.Add([string]$entry.Name)) { throw 'PITR direct ledger duplicated a leaf' } }
  foreach ($path in [IO.Directory]::EnumerateFileSystemEntries([string]$RunRoot.Root, '*', [IO.SearchOption]::TopDirectoryOnly)) {
    if (-not $registered.Contains([IO.Path]::GetFileName($path))) { throw "PITR cleanup refused unknown direct sibling: $path" }
  }
  foreach ($entry in @($RunRoot.Ledger)) {
    if ([string]$entry.Lifecycle -ceq 'Bound' -and $null -ne $entry.CleanupHandle) { $entry.CleanupHandle.Dispose(); $entry.CleanupHandle = $null }
  }
  $walView = Verify-C12ControllerOwnershipWAL -State $RunRoot.ControllerWAL
  # Preflight every uncertain creation before appending any cleanup record.
  foreach ($entry in @($RunRoot.Ledger)) {
    $name = [string]$entry.Name
    if ($walView.ResourceStates.ContainsKey($name) -and [string]$walView.ResourceStates[$name] -ceq 'INTENT') {
      $path = Join-Path ([string]$RunRoot.Root) $name
      try { $uncertain = [C12SealedExecutable]::TryInspectPath($path) } catch { throw "controller WAL INTENT cannot classify exact absence for ${name}: $($_.Exception.Message)" }
      if ($null -ne $uncertain) { throw "controller WAL INTENT found present $name without an authenticated ACTUAL receipt" }
    }
  }
  foreach ($entry in @($RunRoot.Ledger)) {
    $resource = [string]$entry.Name
    $resourcePath = Join-Path ([string]$RunRoot.Root) $resource
    $prior = if ($walView.ResourceStates.ContainsKey($resource)) { [string]$walView.ResourceStates[$resource] } else { '' }
    if ($prior -ceq '') {
      try { $unregisteredObserved = [C12SealedExecutable]::TryInspectPath($resourcePath) } catch { throw "controller WAL cannot classify unregistered $resource without mutating recovery state: $($_.Exception.Message)" }
      if ($null -ne $unregisteredObserved) { throw "controller WAL lacks creation-time ACTUAL receipt for present $resource" }
      $walView = Append-C12ControllerOwnershipRecord -State $RunRoot.ControllerWAL -Event 'INTENT' -PayloadJSON (New-C12PITRLeafEventPayload -Entry $entry -Event 'INTENT')
      $walView = Append-C12ControllerOwnershipRecord -State $RunRoot.ControllerWAL -Event 'NOT_FOUND' -PayloadJSON (New-C12PITRLeafEventPayload -Entry $entry -Event 'NOT_FOUND')
      $prior = 'NOT_FOUND'
    }
    elseif ($prior -ceq 'INTENT') {
      try { $intentObserved = [C12SealedExecutable]::TryInspectPath($resourcePath) } catch { throw "controller WAL INTENT cannot classify exact absence for ${resource}: $($_.Exception.Message)" }
      if ($null -ne $intentObserved) { throw "controller WAL INTENT found present $resource without an authenticated ACTUAL receipt" }
      $walView = Append-C12ControllerOwnershipRecord -State $RunRoot.ControllerWAL -Event 'NOT_FOUND' -PayloadJSON (New-C12PITRLeafEventPayload -Entry $entry -Event 'NOT_FOUND'); $prior = 'NOT_FOUND'
    }
    if ($prior -cin @('ACTUAL','NOT_FOUND')) { $walView = Append-C12ControllerOwnershipRecord -State $RunRoot.ControllerWAL -Event 'CLEAN_INTENT' -PayloadJSON (New-C12PITRLeafEventPayload -Entry $entry -Event 'CLEAN_INTENT') }
  }
  foreach ($entry in @($RunRoot.Ledger)) {
    $path = Join-Path ([string]$RunRoot.Root) ([string]$entry.Name)
    if (-not [IO.File]::Exists($path)) { Complete-C12AbsentDirectLeaf -Entry $entry; continue }
    if ([string]$entry.Lifecycle -cnotin @('Bound','CleanIntent')) { throw "PITR present direct leaf $($entry.Name) is not identity-bound" }
    if ([string]$entry.Lifecycle -ceq 'Bound' -and $null -ne $entry.CleanupHandle) { $entry.CleanupHandle.Dispose(); $entry.CleanupHandle = $null }
    if ($null -eq $entry.CleanupHandle) {
      $handle = [C12SealedExecutable]::OpenLeafForCleanup($path)
      $observedDACLHash = Get-C12SHA256Hex -Bytes ([Text.Encoding]::UTF8.GetBytes([string]$handle.DACL))
      if ("$([UInt32]$handle.VolumeSerialNumber):$([UInt64]$handle.FileIndex)" -cne [string]$entry.Identity -or [UInt32]$handle.NumberOfLinks -ne 1 -or [bool]$handle.Reparse -or
          [string]$handle.Owner -cne [string]$entry.Owner -or $observedDACLHash -cne [string]$entry.DACLHash) {
        $handle.Dispose()
        throw "PITR direct leaf $($entry.Name) exact identity, link count, owner, or DACL changed"
      }
      $entry.CleanupHandle = $handle
    }
  }
  # Validate every direct leaf before issuing any disposition, preserving all evidence on one foreign sibling.
  foreach ($entry in @($RunRoot.Ledger | Sort-Object @{Expression={if ([string]$_.Name -ceq 'controller-ownership-v1.wal'){1}else{0}}})) {
    $path = Join-Path ([string]$RunRoot.Root) ([string]$entry.Name)
    $payload = New-C12PITRLeafEventPayload -Entry $entry -Event 'CLEAN_RESULT'
    $currentView = Verify-C12ControllerOwnershipWAL -State $RunRoot.ControllerWAL
    $alreadyClean = $currentView.ResourceStates.ContainsKey([string]$entry.Name) -and [string]$currentView.ResourceStates[[string]$entry.Name] -ceq 'CLEAN_RESULT'
    if (-not [IO.File]::Exists($path)) { if (-not $alreadyClean) { $null = Append-C12ControllerOwnershipRecord -State $RunRoot.ControllerWAL -Event 'CLEAN_RESULT' -PayloadJSON $payload }; continue }
    try {
      Request-C12DirectLeafDelete -Entry $entry -Path $path -Deadline $Deadline
      Complete-C12DirectLeafDelete -Entry $entry -Path $path -Deadline $Deadline
      if ([string]$entry.Name -cne 'controller-ownership-v1.wal' -and -not $alreadyClean) { $null = Append-C12ControllerOwnershipRecord -State $RunRoot.ControllerWAL -Event 'CLEAN_RESULT' -PayloadJSON $payload }
      $entry.LastCleanupError = ''
    }
    catch { $entry.LastCleanupError = $_.Exception.Message; throw }
  }
  $RunRoot.Ownership.RequestDeleteExactTree($Deadline)
  $RunRoot.Ownership.ReleaseDeletePending()
  if ($null -ne [C12SealedExecutable]::TryInspectPath([string]$RunRoot.Root)) { throw 'PITR run root remains after exact handle cleanup' }
  $RunRoot | Add-Member -NotePropertyName ControllerWALTerminal -NotePropertyValue 'absence-after-authenticated-CLEAN_INTENT' -Force
  $RunRoot.RootLifecycle = 'Absent'
}

function New-C12PITRRunRoot {
  param([Parameter(Mandatory)][string]$RunSuffix,[Parameter(Mandatory)][string]$NonceDigest,[Parameter(Mandatory)][string]$HMACKeyHex,[Parameter(Mandatory)][string]$DockerExecutableDigest,[Parameter(Mandatory)][string]$DockerEndpointIdentityDigest,[DateTime]$ControllerTimestamp=[DateTime]::UtcNow)

  $stage = 'PITR run-root creation'
  $runRoot = $null
  $root = Join-Path ([IO.Path]::GetTempPath()) "talenro-c12-pitr-$RunSuffix"
  $ownership = New-C12OwnedDirectory -Root $root -ExpectedParent ([IO.Path]::GetTempPath()) -LeafPattern '^talenro-c12-pitr-[0-9a-f]{32}$' -Stage 'PITR run-root creation'
  try {
    $currentSID = [Security.Principal.WindowsIdentity]::GetCurrent().User
    $systemSID = New-Object Security.Principal.SecurityIdentifier('S-1-5-18')
    $security = New-Object Security.AccessControl.DirectorySecurity
    $security.SetAccessRuleProtection($true, $false)
    $security.SetOwner($currentSID)
    $inheritance = [Security.AccessControl.InheritanceFlags]::ContainerInherit -bor [Security.AccessControl.InheritanceFlags]::ObjectInherit
    $propagation = [Security.AccessControl.PropagationFlags]::None
    $allow = [Security.AccessControl.AccessControlType]::Allow
    $rights = [Security.AccessControl.FileSystemRights]::FullControl
    $security.AddAccessRule((New-Object Security.AccessControl.FileSystemAccessRule($currentSID, $rights, $inheritance, $propagation, $allow)))
    $security.AddAccessRule((New-Object Security.AccessControl.FileSystemAccessRule($systemSID, $rights, $inheritance, $propagation, $allow)))
    [IO.Directory]::SetAccessControl($root, $security)
    Write-Verbose 'PITR exact_file CreationClosed NeverAttempted Remove unknown sibling direct ledger'
    $ledger = New-C12DirectLeafLedger
    $null = Register-C12PITRLeaf -Ledger $ledger -Name 'controller-ownership-v1.wal'
    $null = Register-C12PITRLeaf -Ledger $ledger -Name 'ownership.wal'
    $null = Register-C12PITRLeaf -Ledger $ledger -Name 'tlsgen.go'
    $null = Register-C12PITRLeaf -Ledger $ledger -Name 'server.crt'
    $null = Register-C12PITRLeaf -Ledger $ledger -Name 'server.key'
    $runRoot = [pscustomobject]@{ Root = $root; Ownership = $ownership; Ledger = $ledger; CreationClosed = $false; RootLifecycle = 'Bound'; ControllerWAL = $null; WALPath = '' }
    $workerWALName = [string]$ledger[1].Name
    Start-C12PITRLeafCreation -RunRoot $runRoot -Name $workerWALName
    $walPath = Join-Path $root $workerWALName
    $stream = New-Object IO.FileStream($walPath, [IO.FileMode]::CreateNew, [IO.FileAccess]::Write, [IO.FileShare]::Read, 4096, [IO.FileOptions]::WriteThrough)
    try { $stream.Flush($true) } finally { $stream.Dispose() }
    $null = Bind-C12PITRLeaf -RunRoot $runRoot -Name $workerWALName
    $runRoot.WALPath = $walPath
    $controllerKey = ConvertFrom-C12Hex -Value $HMACKeyHex
    try { $runRoot.ControllerWAL = Open-C12ControllerOwnershipWAL -RunRoot $runRoot -RunSuffix $RunSuffix -NonceDigest $NonceDigest -DockerExecutableDigest $DockerExecutableDigest -DockerEndpointIdentityDigest $DockerEndpointIdentityDigest -HMACKey $controllerKey -Timestamp $ControllerTimestamp } finally { [Array]::Clear($controllerKey,0,$controllerKey.Length) }
    foreach ($index in 1,0) {
      $entry = $ledger[$index]
      $null = Append-C12ControllerOwnershipRecord -State $runRoot.ControllerWAL -Event 'INTENT' -PayloadJSON (New-C12PITRLeafEventPayload -Entry $entry -Event 'INTENT')
      $null = Append-C12PITRLeafActual -RunRoot $runRoot -Entry $entry
    }
    foreach ($index in 2..4) { Start-C12PITRLeafCreation -RunRoot $runRoot -Name ([string]$ledger[$index].Name) }
    $runRoot.CreationClosed = $true
    return $runRoot
  }
  catch {
    if ($null -ne $runRoot) { try { Remove-C12PITRRunRoot -RunRoot $runRoot -Deadline ([DateTime]::UtcNow.AddSeconds(10)) } catch {} }
    throw
  }
}

function New-C12PITRDescriptor {
  param(
    [Parameter(Mandatory = $true)][string]$RunSuffix,
    [Parameter(Mandatory = $true)][string]$NonceDigest,
    [Parameter(Mandatory = $true)][string]$HMACKeyHex,
    [Parameter(Mandatory = $true)][object]$Primary,
    [Parameter(Mandatory = $true)][string]$TLSPublicDigest,
    [string]$FailureSeam = ''
  )

  $dockerCommand = Get-Command docker.exe -CommandType Application -ErrorAction Stop | Select-Object -First 1
  $dockerExecutable = [IO.Path]::GetFullPath([string]$dockerCommand.Source)
  $dockerDigest = Get-C12SHA256Hex -Bytes ([IO.File]::ReadAllBytes($dockerExecutable))
  $databaseIdentity = Get-C12SHA256Hex -Bytes ([Text.Encoding]::UTF8.GetBytes("talenro-c12-authority-v7:$RunSuffix`:database-identity"))
  $candidateNames = @()
  $candidateDataNames = @()
  for ($index = 0; $index -lt 8; $index++) {
    $candidateNames += ('talenro-c12-{0}-pitr-candidate-{1:D2}' -f $RunSuffix, $index)
    $candidateDataNames += ('talenro-c12-{0}-pitr-candidate-{1:D2}-data' -f $RunSuffix, $index)
  }
  $descriptor = [ordered]@{
    schema = 'talenro-c12-authority-pitr/v1'
    profile = 'authority-v7-pitr'
    run_suffix = $RunSuffix
    nonce_digest = $NonceDigest
    docker_executable = $dockerExecutable
    docker_executable_digest = $dockerDigest
    image_ref = [string]$Primary.ImageRef
    image_id = [string]$Primary.ImageID
    database_identity = $databaseIdentity
    application_name = "talenro-c12-$RunSuffix"
    primary_name = [string]$Primary.Name
    primary_id = [string]$Primary.ID
    primary_data_name = [string]$Primary.PrimaryDataName
    archive_name = [string]$Primary.ArchiveName
    basebackup_name = [string]$Primary.BaseBackupName
    tls_public_digest = $TLSPublicDigest
    observer_role = "talenro_c12_$($RunSuffix.Substring(0,24))_observer"
    slot_name = "talenro_c12_$($RunSuffix.Substring(0,24))_slot"
    plugin = $script:c12PITRPlugin
    candidate_names = [string[]]$candidateNames
    candidate_data_names = [string[]]$candidateDataNames
    max_candidates = 8
    max_wal_line_bytes = $script:c12PITRWALMaximumLineBytes
    max_wal_records = $script:c12PITRWALMaximumRecords
    failure_seam = $FailureSeam
    descriptor_digest = ''
    descriptor_hmac = ''
  }
  $body = $descriptor | ConvertTo-Json -Compress -Depth 6
  $descriptor.descriptor_digest = Get-C12SHA256Hex -Bytes ([Text.Encoding]::UTF8.GetBytes($body))
  $keyBytes = ConvertFrom-C12Hex -Value $HMACKeyHex
  $hmac = New-Object Security.Cryptography.HMACSHA256
  $hmac.Key = $keyBytes
  try {
    $digestBytes = ConvertFrom-C12Hex -Value ([string]$descriptor.descriptor_digest)
    $descriptor.descriptor_hmac = (($hmac.ComputeHash($digestBytes) | ForEach-Object { $_.ToString('x2') }) -join '')
  }
  finally {
    $hmac.Dispose()
    [Array]::Clear($keyBytes, 0, $keyBytes.Length)
  }
  $encoded = [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes(($descriptor | ConvertTo-Json -Compress -Depth 6))).TrimEnd('=').Replace('+', '-').Replace('/', '_')
  return $encoded
}

function Remove-C12PITRCandidates {
  param(
    [Parameter(Mandatory = $true)][string]$RunSuffix,
    [Parameter(Mandatory = $true)][string]$NonceDigest,
    [Parameter(Mandatory = $true)][string]$ImageRef,
    [Parameter(Mandatory = $true)][string]$ImageID,
    [Parameter(Mandatory = $true)][string]$WALPath,
    [Parameter(Mandatory = $true)][DateTime]$Deadline
  )

  $activeCandidateNames = [System.Collections.Generic.Dictionary[string,bool]]::new([StringComparer]::Ordinal)
  if ([IO.File]::Exists($WALPath)) {
    $lineCount = 0
    foreach ($line in [IO.File]::ReadLines($WALPath)) {
      $lineCount++
      if ($lineCount -gt $script:c12PITRWALMaximumRecords -or [Text.Encoding]::ASCII.GetByteCount($line) -gt $script:c12PITRWALMaximumLineBytes) {
        throw 'PITR cleanup WAL exceeded its closed bounds'
      }
      try { $record = $line | ConvertFrom-Json } catch { throw 'PITR cleanup WAL contains malformed JSON' }
      $name = [string]$record.name
      if ($name -match "^talenro-c12-$RunSuffix-pitr-candidate-(0[0-7])(?:-data)?$") {
        if ([string]$record.event -ceq 'CLEAN_RESULT' -and [string]$record.state -ceq 'absent') {
          $activeCandidateNames[$name] = $false
        }
        elseif ([string]$record.event -cin @('INTENT','ACTUAL','RECOVERED_ACTUAL','TRANSITION','CLEAN_INTENT')) {
          $activeCandidateNames[$name] = $true
        }
      }
    }
  }
  $candidateIndices = New-Object 'System.Collections.Generic.SortedSet[int]'
  foreach ($entry in $activeCandidateNames.GetEnumerator()) {
    if ($entry.Value -and $entry.Key -match '^talenro-c12-[0-9a-f]{32}-pitr-candidate-(0[0-7])') {
      [void]$candidateIndices.Add([int]$Matches[1])
    }
  }
  $orderedIndices = @($candidateIndices | Sort-Object -Descending)
  foreach ($index in $orderedIndices) {
    $name = 'talenro-c12-{0}-pitr-candidate-{1:D2}' -f $RunSuffix, $index
    $dataName = "$name-data"
    $inspect = Invoke-C12Docker -Arguments @(
      'container', 'inspect', '--format',
      '{{.Id}}|{{.Name}}|{{ index .Config.Labels `talenro.c12.managed` }}|{{ index .Config.Labels `talenro.c12.run` }}|{{ index .Config.Labels `talenro.c12.profile` }}|{{ index .Config.Labels `talenro.c12.role` }}|{{ index .Config.Labels `talenro.c12.nonce-digest` }}|{{.Config.Image}}|{{.Image}}',
      $name
    ) -Stage "inspect exact PITR candidate $index for cleanup" -Timeout ([TimeSpan]::FromSeconds(5)) -Deadline $Deadline -AllowFailure
    if ($inspect.ExitCode -eq 0) {
      $identity = Get-C12SingleOutputLine -Result $inspect -Stage "inspect exact PITR candidate $index for cleanup"
      $parts = $identity -split '\|', 9
      if ($parts.Count -ne 9 -or $parts[1] -cne "/$name" -or $parts[2] -cne 'true' -or $parts[3] -cne $RunSuffix -or
          $parts[4] -cne 'authority-v7-pitr' -or $parts[5] -cne "candidate-$('{0:D2}' -f $index)" -or
          $parts[6] -cne $NonceDigest -or $parts[7] -cne $ImageRef -or $parts[8] -cne $ImageID -or $parts[0] -notmatch '^[0-9a-f]{64}$') {
        throw 'foreign canary or mismatched PITR candidate was preserved and cleanup failed closed'
      }
      $null = Invoke-C12Docker -Arguments @('container', 'stop', '--time', '2', [string]$parts[0]) -Stage "stop exact PITR candidate $index" -Timeout ([TimeSpan]::FromSeconds(5)) -Deadline $Deadline -AllowFailure
      $null = Invoke-C12Docker -Arguments @('container', 'rm', [string]$parts[0]) -Stage "remove exact PITR candidate $index" -Timeout ([TimeSpan]::FromSeconds(5)) -Deadline $Deadline
    }
    $volume = New-C12PITRVolumeResource -Name $dataName -Role "candidate-$('{0:D2}' -f $index)-data" -NonceDigest $NonceDigest
    Remove-C12PITRVolume -Resource $volume -RunSuffix $RunSuffix -Deadline $Deadline
  }
}

function Convert-C12RunToTests {
  param(
    [Parameter(Mandatory = $true)]
    [string]$Pattern
  )

  if ($Pattern -match '^\^Test[A-Za-z0-9_]+\$$') {
    return @($Pattern.Substring(1, $Pattern.Length - 2))
  }
  if ($Pattern -match '^\^\(Test[A-Za-z0-9_]+(?:\|Test[A-Za-z0-9_]+)+\)\$$') {
    return @(($Pattern.Substring(2, $Pattern.Length - 4)) -split '\|')
  }
  throw '-Run must be an anchored literal Test name or anchored alternation of literal Test names'
}

function ConvertTo-C12RunPattern {
  param(
    [Parameter(Mandatory = $true)]
    [string[]]$Tests
  )

  if ($Tests.Count -eq 1) {
    return '^' + $Tests[0] + '$'
  }
  return '^(' + ($Tests -join '|') + ')$'
}

function Resolve-C12FocusedTestMap {
  param(
    [Parameter(Mandatory = $true)]
    [string[]]$Packages,

    [Parameter(Mandatory = $true)]
    [string[]]$RequestedTests,

    [Parameter(Mandatory = $true)]
    [TimeSpan]$SetupAllowance,

    [Parameter(Mandatory = $true)]
    [ValidateSet('base', 'authority-v7', 'authority-v7-pitr')]
    [string]$Profile
  )

  $result = Invoke-C12TrustedValidator -Mode 'focused' -DataRoot $script:c12RepositoryRoot -Packages $Packages -Tests $RequestedTests -Profile $Profile -SetupAllowance $SetupAllowance
  $prefix = 'C12_FOCUSED_MAP:'
  $mappingLines = @($result.Output | Where-Object { ([string]$_).StartsWith($prefix, [StringComparison]::Ordinal) })
  if ($mappingLines.Count -ne 1) {
    throw 'focused test resolver did not return one bounded package map'
  }
  try {
    $mapping = ([string]$mappingLines[0]).Substring($prefix.Length) | ConvertFrom-Json
  }
  catch {
    throw 'focused test resolver returned malformed JSON'
  }
  $resolved = @{}
  $covered = @{}
  foreach ($package in $Packages) {
    $property = @($mapping.PSObject.Properties | Where-Object { $_.Name -ceq $package })
    if ($property.Count -ne 1) {
      throw "focused test resolver omitted exact package $package"
    }
    $localTests = @($property[0].Value | ForEach-Object { [string]$_ })
    if ($localTests.Count -lt 1 -or $localTests.Count -gt 512) {
      throw "focused test resolver returned an invalid bounded test set for $package"
    }
    foreach ($testName in $localTests) {
      if ($testName -notmatch '^Test[A-Za-z0-9_]+$' -or $covered.ContainsKey($testName)) {
        throw "focused test resolver returned a malformed or ambiguous test $testName"
      }
      $covered[$testName] = $package
    }
    $resolved[$package] = $localTests
  }
  foreach ($testName in $RequestedTests) {
    if (-not $covered.ContainsKey($testName)) {
      throw "focused test resolver did not cover requested test $testName"
    }
  }
  if ($covered.Count -ne $RequestedTests.Count) {
    throw 'focused test resolver returned extra tests'
  }
  return $resolved
}

function Assert-C12GoJSONResult {
  param(
    [Parameter(Mandatory = $true)]
    [object]$Result,

    [Parameter(Mandatory = $true)]
    [string]$Package,

    [string[]]$ExpectedTests = @()
  )

  $expected = @{}
  foreach ($testName in $ExpectedTests) {
    $expected[$testName] = 0
  }
  $observedParents = @{}
  $topLevelTerminals = @{}
  $descendants = @{}
  $topLevelPasses = 0
  $packagePasses = 0
  foreach ($line in $Result.Output) {
    if ([string]::IsNullOrWhiteSpace([string]$line)) {
      continue
    }
    try {
      $event = [string]$line | ConvertFrom-Json
    }
    catch {
      throw "go test for $Package emitted a non-JSON line"
    }
    $hasTest = $event.PSObject.Properties.Name -contains 'Test'
    $testName = if ($hasTest) { [string]$event.Test } else { '' }
    $eventPackage = if ($event.PSObject.Properties.Name -contains 'Package') { [string]$event.Package } else { '' }
    if (-not [string]::IsNullOrEmpty($eventPackage) -and $eventPackage -cne $Package) {
      throw "go test emitted an event for unexpected package $eventPackage"
    }
    if (-not $hasTest) {
      if ($event.Action -eq 'skip') {
        throw "go test for $Package skipped the package"
      }
      if ($event.Action -eq 'fail') {
        throw "go test for $Package failed the package"
      }
      if ($event.Action -eq 'pass') {
        $packagePasses++
      }
      continue
    }

    if ($testName -match '/') {
      $parts = @($testName -split '/')
      $parent = [string]$parts[0]
      if (($expected.Count -gt 0 -and -not $expected.ContainsKey($parent)) -or
          ($expected.Count -eq 0 -and -not $observedParents.ContainsKey($parent))) {
        throw "go test for $Package emitted orphan descendant $testName"
      }
      $depth = $parts.Count - 1
      if ($depth -gt 4) {
        throw "go test for $Package descendant $testName exceeds depth 4"
      }
      $suffixLength = $testName.Length - $parent.Length - 1
      if ($suffixLength -lt 1 -or $suffixLength -gt 256) {
        throw "go test for $Package descendant $testName has malformed bounded suffix length"
      }
      for ($index = 1; $index -lt $parts.Count; $index++) {
        if ([string]$parts[$index] -notmatch '^[A-Za-z0-9_.-]{1,64}$') {
          throw "go test for $Package descendant $testName has malformed suffix"
        }
      }
      if (-not $descendants.ContainsKey($testName)) {
        if ($descendants.Count -ge 256) {
          throw "go test for $Package exceeds the bounded descendant count 256"
        }
        $descendants[$testName] = [pscustomobject]@{ Passes = 0; Skips = 0; Fails = 0 }
      }
      $state = $descendants[$testName]
      switch ([string]$event.Action) {
        'pass' { $state.Passes = [int]$state.Passes + 1 }
        'skip' { $state.Skips = [int]$state.Skips + 1 }
        'fail' { $state.Fails = [int]$state.Fails + 1 }
      }
      continue
    }

    $observedParents[$testName] = $true
    if ($expected.Count -gt 0 -and -not $expected.ContainsKey($testName)) {
      throw "go test for $Package emitted extra top-level test $testName"
    }
    if ($event.Action -eq 'skip') {
      throw "go test for $Package skipped $testName"
    }
    if ($event.Action -eq 'fail') {
      throw "go test for $Package failed $testName"
    }
    if ($event.Action -eq 'pass') {
      if (-not $topLevelTerminals.ContainsKey($testName)) {
        $topLevelTerminals[$testName] = 0
      }
      $topLevelTerminals[$testName] = [int]$topLevelTerminals[$testName] + 1
      if ($expected.ContainsKey($testName)) {
        $expected[$testName] = [int]$expected[$testName] + 1
      }
      $topLevelPasses++
    }
  }
  foreach ($testName in $expected.Keys) {
    if ([int]$expected[$testName] -ne 1) {
      throw "go test for $Package has missing or duplicate terminal pass for $testName"
    }
  }
  foreach ($testName in $topLevelTerminals.Keys) {
    if ([int]$topLevelTerminals[$testName] -ne 1) {
      throw "go test for $Package has missing or duplicate terminal pass for $testName"
    }
  }
  foreach ($testName in $descendants.Keys) {
    $state = $descendants[$testName]
    if ([int]$state.Skips -ne 0) {
      throw "go test for $Package skipped descendant $testName"
    }
    if ([int]$state.Fails -ne 0) {
      throw "go test for $Package failed descendant $testName"
    }
    if ([int]$state.Passes -eq 0) {
      throw "go test for $Package has missing terminal pass for descendant $testName"
    }
    if ([int]$state.Passes -ne 1) {
      throw "go test for $Package has duplicate terminal pass for descendant $testName"
    }
  }
  if ($topLevelPasses -eq 0 -or $packagePasses -ne 1 -or $Result.ExitCode -ne 0) {
    throw "go test for $Package did not produce an exact passing JSON result"
  }
}

function Assert-C12ExecutionPlan {
  param(
    [Parameter(Mandatory = $true)]
    [object[]]$Groups
  )

  if ($Groups.Count -lt 1 -or $Groups.Count -gt 256) {
    throw 'suite execution plan has an invalid bounded group count'
  }
  $priorID = ''
  foreach ($group in $Groups) {
    $groupID = [string]$group.id
    $package = [string]$group.package
    $profile = [string]$group.profile
    $timeout = [string]$group.timeout
    $tests = @($group.tests | ForEach-Object { [string]$_ })
    if ($groupID -notmatch '^[a-z][a-z0-9-]{0,63}$' -or
        (-not [string]::IsNullOrEmpty($priorID) -and [string]::CompareOrdinal($priorID, $groupID) -ge 0)) {
      throw "invalid suite execution-plan group ID $groupID"
    }
    $priorID = $groupID
    if (-not $script:c12AllowedPackages.ContainsKey($package)) {
      throw "suite execution-plan package $package is outside the Task 4 allowed set"
    }
    if ($profile -cnotin $script:c12AllowedProfiles) {
      throw "suite execution-plan group $groupID has unsupported closed profile"
    }
    $groupDuration = ConvertFrom-C12Duration -Value $timeout
    if ($groupDuration -gt [TimeSpan]::FromMinutes(30)) {
      throw "suite execution-plan group $groupID exceeds the 30m group timeout"
    }
    if ($tests.Count -lt 1 -or $tests.Count -gt 512) {
      throw "suite execution-plan group $groupID has an invalid bounded test count"
    }
    $priorTest = ''
    foreach ($testName in $tests) {
      if ($testName -notmatch '^Test[A-Za-z0-9_]+$' -or
          (-not [string]::IsNullOrEmpty($priorTest) -and [string]::CompareOrdinal($priorTest, $testName) -ge 0)) {
        throw "suite execution-plan group $groupID contains an invalid, duplicate, or unsorted literal test"
      }
      $priorTest = $testName
    }
  }
}

function Invoke-C12Group {
  param(
    [Parameter(Mandatory = $true)]
    [string]$GroupID,

    [Parameter(Mandatory = $true)]
    [string]$GroupProfile,

    [Parameter(Mandatory = $true)]
    [string]$Package,

    [string]$RunPattern = '',

    [Parameter(Mandatory = $true)]
    [string]$GroupTimeout,

    [string[]]$ExpectedTests = @(),

    [string]$CandidateTree = '',

    [DateTime]$AbsoluteDeadline = [DateTime]::MaxValue
  )

  if ($GroupProfile -cnotin $script:c12AllowedProfiles) {
    throw 'C12 runner rejected an unsupported closed profile'
  }
	$groupDuration = ConvertFrom-C12Duration -Value $GroupTimeout
	$groupDeadline = if ($AbsoluteDeadline -eq [DateTime]::MaxValue) { [DateTime]::UtcNow.Add($groupDuration).Add($script:c12ProfileAllowances[$GroupProfile]) } else { $AbsoluteDeadline }
  if ($script:c12SuiteDeadline -lt $groupDeadline) {
    $groupDeadline = $script:c12SuiteDeadline
  }
  $priorNativeDeadline = $script:c12NativeDeadline
  $script:c12NativeDeadline = $groupDeadline
  $runSuffix = New-C12RandomSuffix
  if ($runSuffix -notmatch '^[0-9a-f]{32}$') {
    throw 'cryptographic C12 run suffix is malformed'
  }
  $databaseName = "talenro_c12_$runSuffix"
  $databasePassword = New-C12DatabasePassword -RunSuffix $runSuffix
  $pitrNonce = ''
  $pitrNonceDigest = ''
  $pitrHMACKey = ''
  $pitrRunRoot = $null
  $pitrVolumes = @()
  $pitrTLSPublicDigest = ''
  if ($GroupProfile -ceq 'authority-v7-pitr') {
    $pitrNonce = New-C12PITRSecretHex
    $pitrNonceDigest = Get-C12SHA256Hex -Bytes (ConvertFrom-C12Hex -Value $pitrNonce)
    $pitrHMACKey = New-C12PITRSecretHex
    $primaryName = "talenro-c12-$runSuffix-pitr-primary"
    $primaryDataName = "talenro-c12-$runSuffix-pitr-primary-data"
    $archiveName = "talenro-c12-$runSuffix-pitr-archive"
    $basebackupName = "talenro-c12-$runSuffix-pitr-basebackup"
    $resources = @(
      [pscustomobject]@{ Kind = 'postgres'; Name = $primaryName; ID = ''; ImageRef = 'postgres:18.4-alpine3.23'; ImageID = ''; Port = 0; ContainerPort = 5432; PITR = $true; NonceDigest = $pitrNonceDigest; PrimaryDataName = $primaryDataName; ArchiveName = $archiveName; BaseBackupName = $basebackupName },
      [pscustomobject]@{ Kind = 'redis'; Name = "talenro-c12-$runSuffix-redis"; ID = ''; ImageRef = 'redis:8.8.1-alpine3.23'; ImageID = ''; Port = 0; ContainerPort = 6379; PITR = $false },
      [pscustomobject]@{ Kind = 'nats'; Name = "talenro-c12-$runSuffix-nats"; ID = ''; ImageRef = 'nats:2.14.3-alpine3.22'; ImageID = ''; Port = 0; ContainerPort = 4222; PITR = $false }
    )
    $pitrVolumes = @(
      (New-C12PITRVolumeResource -Name $primaryDataName -Role 'primary-data' -NonceDigest $pitrNonceDigest),
      (New-C12PITRVolumeResource -Name $archiveName -Role 'archive' -NonceDigest $pitrNonceDigest),
      (New-C12PITRVolumeResource -Name $basebackupName -Role 'basebackup' -NonceDigest $pitrNonceDigest)
    )
  }
  else {
    $resources = @(
      [pscustomobject]@{ Kind = 'postgres'; Name = "talenro-c12-$runSuffix-postgres"; ID = ''; ImageRef = 'postgres:18.4-alpine3.23'; ImageID = ''; Port = 0; ContainerPort = 5432; PITR = $false },
      [pscustomobject]@{ Kind = 'redis'; Name = "talenro-c12-$runSuffix-redis"; ID = ''; ImageRef = 'redis:8.8.1-alpine3.23'; ImageID = ''; Port = 0; ContainerPort = 6379; PITR = $false },
      [pscustomobject]@{ Kind = 'nats'; Name = "talenro-c12-$runSuffix-nats"; ID = ''; ImageRef = 'nats:2.14.3-alpine3.22'; ImageID = ''; Port = 0; ContainerPort = 4222; PITR = $false }
    )
  }
  $primaryFailure = $null
  $cleanupFailures = @()
  $preparedInitializer = $null
  $preparedArtifactRoot = $null
  try {
    if ($GroupProfile -cin @('authority-v7', 'authority-v7-pitr')) {
      $preparedInitializer = New-C12PreparedAuthorityInitializer -Profile $GroupProfile -RunSuffix $runSuffix -CandidateTree $CandidateTree -SetupAllowance ([TimeSpan]$script:c12ProfileAllowances[$GroupProfile]) -Deadline $groupDeadline
      $preparedArtifactRoot = $preparedInitializer.ArtifactRoot
    }
    foreach ($resource in $resources) {
      Resolve-C12ImageIdentity -Resource $resource
    }
    if ($GroupProfile -ceq 'authority-v7-pitr') {
      $dockerEndpointReceipt = Get-C12DockerEndpointReceipt -Deadline $groupDeadline -Revalidate
      $dockerExecutableDigest = [string]$dockerEndpointReceipt.ExecutableReceipt.Digest
      $dockerEndpointIdentityDigest = [string]$dockerEndpointReceipt.Digest
      $pitrRunRoot = New-C12PITRRunRoot -RunSuffix $runSuffix -NonceDigest $pitrNonceDigest -HMACKeyHex $pitrHMACKey -DockerExecutableDigest $dockerExecutableDigest -DockerEndpointIdentityDigest $dockerEndpointIdentityDigest
      $dockerRegistryPayload=New-C12DockerRegistryPayload -Resources $resources -Volumes $pitrVolumes -RunSuffix $runSuffix -NonceDigest $pitrNonceDigest -EndpointReceipt $dockerEndpointReceipt
      $pitrRunRoot.ControllerWAL.DockerRegistryPayload=$dockerRegistryPayload
      $null=Append-C12ControllerOwnershipRecord -State $pitrRunRoot.ControllerWAL -Event 'DOCKER_REGISTRY' -PayloadJSON $dockerRegistryPayload
      foreach($dockerResource in @($resources)+@($pitrVolumes)){ $dockerResource | Add-Member ControllerWAL $pitrRunRoot.ControllerWAL -Force }
    }
    foreach ($volume in $pitrVolumes) {
      Start-C12PITRVolume -Resource $volume -RunSuffix $runSuffix -Deadline $groupDeadline
    }
    foreach ($resource in $resources) {
      Start-C12Container -Resource $resource -RunSuffix $runSuffix -DatabaseName $databaseName -DatabasePassword $databasePassword
    }
    Wait-C12Dependencies -Resources $resources
    $postgres = @($resources | Where-Object { $_.Kind -eq 'postgres' })[0]
    $redis = @($resources | Where-Object { $_.Kind -eq 'redis' })[0]
    $nats = @($resources | Where-Object { $_.Kind -eq 'nats' })[0]
    $databaseURL = "postgres://talenro:$databasePassword@127.0.0.1:$($postgres.Port)/$databaseName`?sslmode=disable&application_name=talenro-c12-$runSuffix"
    [System.Environment]::SetEnvironmentVariable('TALENRO_DATABASE_URL', $databaseURL, 'Process')
    [System.Environment]::SetEnvironmentVariable('TALENRO_REDIS_ADDRESS', "127.0.0.1:$($redis.Port)", 'Process')
    [System.Environment]::SetEnvironmentVariable('TALENRO_NATS_URL', "nats://127.0.0.1:$($nats.Port)", 'Process')

    if ($GroupProfile -ceq 'authority-v7-pitr') {
      $pitrTLSPublicDigest = Initialize-C12PITRPrimary -Primary $postgres -DatabaseName $databaseName -RunRoot $pitrRunRoot -Deadline $groupDeadline
      $databaseURL = "postgres://talenro:$databasePassword@127.0.0.1:$($postgres.Port)/$databaseName`?sslmode=disable&application_name=talenro-c12-$runSuffix"
      [System.Environment]::SetEnvironmentVariable('TALENRO_DATABASE_URL', $databaseURL, 'Process')
      try {
        Wait-C12Dependencies -Resources $resources
      }
      catch {
        $stateResult = Invoke-C12Docker -Arguments @('container', 'inspect', '--format', '{{.State.Status}}|{{.State.Health.Status}}|{{.State.Error}}', [string]$postgres.ID) -Stage 'inspect failed PITR TLS restart' -Deadline $groupDeadline -AllowFailure
        foreach ($line in $stateResult.Output) { [Console]::Error.WriteLine([string]$line) }
        $logResult = Invoke-C12Docker -Arguments @('container', 'logs', '--tail', '20', [string]$postgres.ID) -Stage 'read failed PITR TLS restart logs' -Deadline $groupDeadline -AllowFailure
        foreach ($line in $logResult.Output) { [Console]::Error.WriteLine(([string]$line).Replace($databasePassword, '[redacted]')) }
        throw
      }
    }

    $migrationDirectory = Join-Path $script:c12RepositoryRoot 'db\migrations'
    $goArgs = @('tool', 'goose', '-dir', $migrationDirectory, 'postgres', $databaseURL, 'up-to', '6')
    $baseMigrationResult = Invoke-C12Go -Arguments $goArgs -Stage "ordinary Goose base migration for $GroupID" -Deadline $groupDeadline -AllowFailure
    if ($baseMigrationResult.ExitCode -ne 0) {
      foreach ($line in @($baseMigrationResult.Output | Select-Object -Last 12)) {
        [Console]::Error.WriteLine(([string]$line).Replace($databasePassword, '[redacted]'))
      }
      throw "ordinary Goose base migration for $GroupID failed with exit code $([int]$baseMigrationResult.ExitCode)"
    }

    if ($GroupProfile -ceq 'authority-v7' -or $GroupProfile -ceq 'authority-v7-pitr') {
      $initNonce = (New-C12RandomSuffix) + (New-C12RandomSuffix)
      $null = Invoke-C12AuthorityInitializer -Prepared $preparedInitializer -DatabaseURL $databaseURL -InitNonce $initNonce -RunSuffix $runSuffix -Profile $GroupProfile -Deadline $groupDeadline
    }

    if ($GroupProfile -ceq 'authority-v7-pitr') {
      $descriptor = New-C12PITRDescriptor -RunSuffix $runSuffix -NonceDigest $pitrNonceDigest -HMACKeyHex $pitrHMACKey -Primary $postgres -TLSPublicDigest $pitrTLSPublicDigest -FailureSeam $PITRFailureSeam
      [System.Environment]::SetEnvironmentVariable('TALENRO_C12_AUTHORITY_PITR_NONCE', $pitrNonce, 'Process')
      [System.Environment]::SetEnvironmentVariable('TALENRO_C12_AUTHORITY_PITR_DESCRIPTOR', $descriptor, 'Process')
      [System.Environment]::SetEnvironmentVariable('TALENRO_C12_AUTHORITY_PITR_WAL_PATH', [string]$pitrRunRoot.WALPath, 'Process')
      [System.Environment]::SetEnvironmentVariable('TALENRO_C12_AUTHORITY_PITR_HMAC_KEY', $pitrHMACKey, 'Process')
    }

    $goArgs = @('test', '-json', '-tags=integration', '-p=1', '-count=1', '-timeout', $GroupTimeout)
    if (-not [string]::IsNullOrEmpty($RunPattern)) {
      $goArgs += @('-run', $RunPattern)
    }
    $goArgs += $Package
    $testResult = Invoke-C12Go -Arguments $goArgs -Stage "tagged Go test group $GroupID" -Timeout $groupDuration -Deadline $groupDeadline -AllowFailure
    try {
      Assert-C12GoJSONResult -Result $testResult -Package ([string]$script:c12AllowedPackages[$Package]) -ExpectedTests $ExpectedTests
    }
    catch {
      foreach ($line in @($testResult.Output | Select-Object -Last 30)) {
        $redacted = ([string]$line).Replace($databasePassword, '[redacted]')
        if (-not [string]::IsNullOrEmpty($pitrNonce)) { $redacted = $redacted.Replace($pitrNonce, '[redacted]') }
        if (-not [string]::IsNullOrEmpty($pitrHMACKey)) { $redacted = $redacted.Replace($pitrHMACKey, '[redacted]') }
        [Console]::Error.WriteLine($redacted)
      }
      throw
    }
  }
  catch {
    $primaryFailure = $_.Exception.Message
    if (-not [string]::IsNullOrWhiteSpace([string]$_.ScriptStackTrace)) {
      $primaryFailure += " [$([string]$_.ScriptStackTrace)]"
    }
  }
  finally {
    try {
      [System.Environment]::SetEnvironmentVariable('TALENRO_DATABASE_URL', $null, 'Process')
      [System.Environment]::SetEnvironmentVariable('TALENRO_REDIS_ADDRESS', $null, 'Process')
      [System.Environment]::SetEnvironmentVariable('TALENRO_NATS_URL', $null, 'Process')
      Clear-C12InheritedCapabilities
      $cleanupDeadline = [DateTime]::UtcNow.Add($script:c12GroupCleanupBudget)
      if ($null -ne $preparedInitializer -or $null -ne $preparedArtifactRoot) {
        try {
          Close-C12PreparedBundle -Prepared $preparedInitializer -ArtifactRoot $preparedArtifactRoot -Deadline $cleanupDeadline
          $preparedInitializer = $null
          $preparedArtifactRoot = $null
        }
        catch { $cleanupFailures += "prepared initializer: $($_.Exception.Message)" }
      }
      if ($GroupProfile -ceq 'authority-v7-pitr' -and $null -ne $pitrRunRoot) {
        try {
          $primary = @($resources | Where-Object { $_.Kind -eq 'postgres' })[0]
          Remove-C12PITRCandidates -RunSuffix $runSuffix -NonceDigest $pitrNonceDigest -ImageRef ([string]$primary.ImageRef) -ImageID ([string]$primary.ImageID) -WALPath ([string]$pitrRunRoot.WALPath) -Deadline $cleanupDeadline
        }
        catch {
          $cleanupFailures += "PITR candidates: $($_.Exception.Message)"
        }
      }
      if ($GroupProfile -ceq 'authority-v7-pitr') {
        try {
          Remove-C12PITRBaseResources -Resources $resources -Volumes $pitrVolumes -RunSuffix $runSuffix -Deadline $cleanupDeadline
        }
        catch {
          $cleanupFailures += "PITR base resources: $($_.Exception.Message)"
        }
      }
      else {
        for ($index = $resources.Count - 1; $index -ge 0; $index--) {
          try {
            Remove-C12Container -Resource $resources[$index] -RunSuffix $runSuffix -Deadline $cleanupDeadline
          }
          catch {
            $cleanupFailures += "$($resources[$index].Kind): $($_.Exception.Message)"
          }
        }
      }
      if ($null -ne $pitrRunRoot) {
        try {
          Remove-C12PITRRunRoot -RunRoot $pitrRunRoot -Deadline $cleanupDeadline
        }
        catch {
          $cleanupFailures += "PITR run root: $($_.Exception.Message)"
        }
      }
      $pitrNonce = ''
      $pitrHMACKey = ''
    }
    finally {
      $script:c12NativeDeadline = $priorNativeDeadline
    }
  }
  if ($null -ne $primaryFailure) {
    [Console]::Error.WriteLine("C12 test failure [$GroupID]: $primaryFailure")
  }
  foreach ($cleanupFailure in $cleanupFailures) {
    [Console]::Error.WriteLine("C12 cleanup failure [$GroupID]: $cleanupFailure")
  }
  if ($null -ne $primaryFailure -or $cleanupFailures.Count -ne 0) {
    throw "C12 group $GroupID failed"
  }
  Write-Output "C12 group $GroupID passed"
}

function Assert-C12NoInheritedDependencies {
  foreach ($name in @('TALENRO_DATABASE_URL', 'TALENRO_REDIS_ADDRESS', 'TALENRO_NATS_URL')) {
    if (-not [string]::IsNullOrEmpty([System.Environment]::GetEnvironmentVariable($name, 'Process'))) {
      throw 'inherited dependency endpoints are forbidden'
    }
  }
  foreach ($name in @('TALENRO_INSTALLATION_KIND', 'TALENRO_ENVIRONMENT')) {
    if (-not [string]::IsNullOrEmpty([System.Environment]::GetEnvironmentVariable($name, 'Process'))) {
      throw 'inherited production markers are forbidden'
    }
  }
  foreach ($inheritedName in @([System.Environment]::GetEnvironmentVariables('Process').Keys)) {
    $name = [string]$inheritedName
    if ($name.StartsWith('GIT_', [StringComparison]::OrdinalIgnoreCase)) {
      throw 'inherited GIT_* variable is forbidden'
    }
    if ($name.StartsWith('TALENRO_C12_AUTHORITY_', [StringComparison]::Ordinal)) {
      throw 'inherited authority capability variable is forbidden'
    }
  }
}

function Clear-C12InheritedCapabilities {
  foreach ($name in @(
    'TALENRO_C12_AUTHORITY_V7_INIT_NONCE',
    'TALENRO_C12_AUTHORITY_V7_RUN_SUFFIX',
    'TALENRO_C12_AUTHORITY_V7_PROFILE',
    'TALENRO_INSTALLATION_KIND',
    'TALENRO_C12_AUTHORITY_PITR_NONCE',
    'TALENRO_C12_AUTHORITY_PITR_DESCRIPTOR',
    'TALENRO_C12_AUTHORITY_PITR_WAL_PATH',
    'TALENRO_C12_AUTHORITY_PITR_HMAC_KEY'
  )) {
    [System.Environment]::SetEnvironmentVariable($name, $null, 'Process')
  }
}

function Invoke-C12FocusedMode {
  if ($Profile -cnotin $script:c12AllowedProfiles) {
    throw 'C12 runner rejected an unsupported closed profile'
  }
  if ($Race -and $Profile -cne 'authority-v7-pitr') {
    throw 'C12 race is allowed only for the authority-v7-pitr profile'
  }
  if (-not [string]::IsNullOrEmpty($PITRFailureSeam)) {
    if ($Profile -cne 'authority-v7-pitr' -or $Packages -cne './internal/testinfra' -or
        $Run -cne "^$($script:c12PITRPrivateTest)`$" -or $Race -or $PITRFailureSeam -cnotin $script:c12PITRFailureSeams) {
      throw 'PITR failure seam is accepted only by the exact private authority-v7-pitr mode'
    }
  }
  elseif ($Run -ceq "^$($script:c12PITRPrivateTest)`$") {
    throw 'private PITR failure-seam selector requires one closed seam'
  }
  if ($Run -ceq "^$($script:c12PITRPublicTest)`$" -and $Profile -cne 'authority-v7-pitr') {
    throw 'public PITR test requires the authority-v7-pitr profile'
  }
  $duration = ConvertFrom-C12Duration -Value $Timeout
  if ($duration -gt [TimeSpan]::FromMinutes(30)) {
    throw 'focused timeout must not exceed 30m'
  }
  if ($Packages -notmatch '^\./[A-Za-z0-9_.-]+(?:/[A-Za-z0-9_.-]+)*(?:\|\./[A-Za-z0-9_.-]+(?:/[A-Za-z0-9_.-]+)*)*$' -or $Packages.Contains('./...')) {
    throw 'invalid -Packages: expected pipe-delimited explicit ./package tokens'
  }
  $packageList = @($Packages -split '\|')
  $seen = @{}
  foreach ($package in $packageList) {
    if ($seen.ContainsKey($package)) {
      throw "duplicate package $package"
    }
    $seen[$package] = $true
    if (-not $script:c12AllowedPackages.ContainsKey($package)) {
      throw "unknown package $package"
    }
  }
  $expectedTests = @()
  $testMap = @{}
  if (-not [string]::IsNullOrEmpty($Run)) {
    $expectedTests = @(Convert-C12RunToTests -Pattern $Run)
    $testMap = Resolve-C12FocusedTestMap -Packages $packageList -RequestedTests $expectedTests -SetupAllowance ([TimeSpan]$script:c12ProfileAllowances[$Profile]) -Profile $Profile
  }
  foreach ($package in $packageList) {
    $groupID = 'focused-' + ($package.TrimStart('.').TrimStart('/').Replace('/', '-'))
    $localTests = @()
    if ($testMap.ContainsKey($package)) {
      $localTests = @($testMap[$package])
    }
    $localRun = if ($localTests.Count -gt 0) { ConvertTo-C12RunPattern -Tests $localTests } else { '' }
    Invoke-C12Group -GroupID $groupID -GroupProfile $Profile -Package $package -RunPattern $localRun -GroupTimeout $Timeout -ExpectedTests $localTests
  }
}

function Invoke-C12SuiteMode {
  if ($Suite -cne 'batch01') {
    throw 'Task 4 runner accepts only the batch01 suite'
  }
  $duration = ConvertFrom-C12Duration -Value $Timeout
  if ($duration -ne [TimeSpan]::FromMinutes(120)) {
    throw 'batch01 suite timeout must be exactly 120m'
  }
  $priorNativeDeadline = $script:c12NativeDeadline
  $script:c12NativeDeadline = $script:c12SuiteDeadline
  $originalRoot = $script:c12RepositoryRoot
  $candidate = $null
  $groups = @()
  try {
    $candidate = New-C12CandidateSnapshot
    $validationSnapshot = New-C12CandidateMaterialization -Candidate $candidate -Stage 'materialize candidate validation snapshot' -Deadline $script:c12SuiteDeadline
    try {
      $script:c12RepositoryRoot = [string]$validationSnapshot.Root
      $manifestRelativePath = 'testdata/c12/integration-contracts-schema-authority.v1.json'
      $manifestPath = Join-Path $script:c12RepositoryRoot ($manifestRelativePath.Replace('/', '\'))
      if (-not (Test-Path -LiteralPath $manifestPath -PathType Leaf)) {
        throw 'canonical Batch 01 manifest is absent from the staged candidate'
      }

      $null = Invoke-C12TrustedValidator -Mode 'suite' -DataRoot $script:c12RepositoryRoot -SuiteName 'batch01' -SuiteTimeout $Timeout -CandidateTree ([string]$candidate.Tree) -ManifestPaths @($manifestRelativePath) -Profile 'authority-v7-pitr' -SetupAllowance ([TimeSpan]$script:c12ProfileAllowances['authority-v7-pitr']) -Deadline $script:c12SuiteDeadline
      $manifest = Get-Content -Raw -LiteralPath $manifestPath | ConvertFrom-Json
      $groups = @($manifest.groups)
      Assert-C12ExecutionPlan -Groups $groups
    }
    finally {
      try {
        Assert-C12CandidateSnapshotClean -Candidate $candidate -Snapshot $validationSnapshot -Deadline $script:c12SuiteDeadline
      }
      finally {
        $script:c12RepositoryRoot = $originalRoot
        Remove-C12CandidateMaterialization -Candidate $candidate -Snapshot $validationSnapshot -Deadline $script:c12SuiteDeadline
      }
    }
    foreach ($group in $groups) {
      $groupDuration = ConvertFrom-C12Duration -Value ([string]$group.timeout)
      $snapshotGroupDeadline = [DateTime]::UtcNow.Add($groupDuration).AddMinutes(3)
      if ($script:c12SuiteDeadline -lt $snapshotGroupDeadline) {
        $snapshotGroupDeadline = $script:c12SuiteDeadline
      }
      $groupSnapshot = New-C12CandidateMaterialization -Candidate $candidate -Stage "materialize candidate group $([string]$group.id) snapshot" -Deadline $snapshotGroupDeadline
      try {
        $script:c12RepositoryRoot = [string]$groupSnapshot.Root
        $tests = @($group.tests | ForEach-Object { [string]$_ })
        $runPattern = '^(' + ($tests -join '|') + ')$'
        Invoke-C12Group -GroupID ([string]$group.id) -GroupProfile ([string]$group.profile) -Package ([string]$group.package) -RunPattern $runPattern -GroupTimeout ([string]$group.timeout) -ExpectedTests $tests -CandidateTree ([string]$candidate.Tree) -AbsoluteDeadline $snapshotGroupDeadline
      }
      finally {
        try {
          Assert-C12CandidateSnapshotClean -Candidate $candidate -Snapshot $groupSnapshot -Deadline $snapshotGroupDeadline
        }
        finally {
          $script:c12RepositoryRoot = $originalRoot
          Remove-C12CandidateMaterialization -Candidate $candidate -Snapshot $groupSnapshot -Deadline $snapshotGroupDeadline
        }
      }
    }
  }
  finally {
    $script:c12RepositoryRoot = $originalRoot
    if ($null -ne $candidate) {
      Remove-C12CandidateSnapshot -OwnerRoot ([string]$candidate.OwnerRoot) -Ownership $candidate.Ownership -Deadline $script:c12SuiteDeadline
    }
    $script:c12NativeDeadline = $priorNativeDeadline
  }
}

$script:c12RepositoryRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..')).Path
$selectedMode = $PSCmdlet.ParameterSetName
$priorGoEnvironment = @{
  GOOS = [System.Environment]::GetEnvironmentVariable('GOOS', 'Process')
  GOARCH = [System.Environment]::GetEnvironmentVariable('GOARCH', 'Process')
  CGO_ENABLED = [System.Environment]::GetEnvironmentVariable('CGO_ENABLED', 'Process')
}
$scriptExitCode = 0
try {
  Assert-C12NoInheritedDependencies
  [System.Environment]::SetEnvironmentVariable('GOOS', 'windows', 'Process')
  [System.Environment]::SetEnvironmentVariable('GOARCH', 'amd64', 'Process')
  [System.Environment]::SetEnvironmentVariable('CGO_ENABLED', '0', 'Process')
  if ($selectedMode -eq 'Focused') {
    Invoke-C12FocusedMode
  }
  else {
    Invoke-C12SuiteMode
  }
}
catch {
  [Console]::Error.WriteLine($_.Exception.Message)
  $scriptExitCode = 1
}
finally {
  try {
    Clear-C12InheritedCapabilities
    Close-C12AllPreparedArtifacts -Deadline ([DateTime]::UtcNow.Add($script:c12GroupCleanupBudget))
  }
  catch {
    [Console]::Error.WriteLine("C12 prepared artifact final cleanup failed: $($_.Exception.Message)")
    $scriptExitCode = 1
  }
  finally {
    foreach ($name in $priorGoEnvironment.Keys) {
      [System.Environment]::SetEnvironmentVariable($name, $priorGoEnvironment[$name], 'Process')
    }
  }
}
exit $scriptExitCode
