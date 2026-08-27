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

  [Parameter(Mandatory = $true, ParameterSetName = 'Suite')]
  [string]$Suite,

  [Parameter(Mandatory = $true, ParameterSetName = 'Focused')]
  [Parameter(Mandatory = $true, ParameterSetName = 'Suite')]
  [string]$Timeout
)

Set-StrictMode -Version 2.0
$ErrorActionPreference = 'Stop'

$script:c12AllowedPackages = @{
  './internal/testinfra' = $true
  './internal/store' = $true
  './internal/nodecontrol/contracts' = $true
  './internal/nodecontrol/authority' = $true
  './internal/nodecontrol/serving' = $true
  './internal/readiness' = $true
}
$script:c12NativeDeadline = [DateTime]::MaxValue
$script:c12SuiteDeadline = [DateTime]::MaxValue
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
	fmt.Printf("C12_TRUSTED_VALIDATOR_OK:%s\n", validatorVersion)
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
    private const UInt32 DELETE = 0x00010000;
    private const UInt32 SYNCHRONIZE = 0x00100000;
    private const UInt32 FILE_SHARE_READ = 0x00000001;
    private const UInt32 FILE_SHARE_WRITE = 0x00000002;
    private const UInt32 FILE_SHARE_DELETE = 0x00000004;
    private const UInt32 OPEN_EXISTING = 3;
    private const UInt32 FILE_CREATE = 2;
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
    private const Int32 FILE_DISPOSITION_INFO_CLASS = 4;
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
                FILE_READ_ATTRIBUTES | DELETE | SYNCHRONIZE,
                ref attributes,
                out ioStatus,
                IntPtr.Zero,
                FILE_ATTRIBUTE_NORMAL,
                FILE_SHARE_READ | FILE_SHARE_WRITE,
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

    public void DeleteExactTree(DateTime deadlineUtc)
    {
        VerifyExactPath();
        int visited = 0;
        DeleteChildren(RootPath, handle, deadlineUtc, ref visited);
        VerifyExactPath();
        MarkDelete(handle);
        Dispose();
        if (System.IO.Directory.Exists(RootPath)) throw new InvalidOperationException("identity-bound root deletion remained pending");
    }

    private static void DeleteChildren(string directory, IntPtr directoryHandle, DateTime deadlineUtc, ref int visited)
    {
        CheckDeadline(deadlineUtc);
        foreach (string child in System.IO.Directory.EnumerateFileSystemEntries(directory, "*", System.IO.SearchOption.TopDirectoryOnly))
        {
            CheckDeadline(deadlineUtc);
            visited++;
            if (visited > 100000) throw new InvalidOperationException("identity-bound cleanup exceeds the bounded entry count");
            IntPtr childHandle = CreateFile(child, FILE_READ_ATTRIBUTES | DELETE, FILE_SHARE_READ | FILE_SHARE_WRITE, IntPtr.Zero, OPEN_EXISTING, FILE_FLAG_BACKUP_SEMANTICS | FILE_FLAG_OPEN_REPARSE_POINT, IntPtr.Zero);
            if (childHandle == INVALID_HANDLE_VALUE) throw new InvalidOperationException("identity-bound cleanup could not lock one descendant", new Win32Exception(Marshal.GetLastWin32Error()));
            try
            {
                BY_HANDLE_FILE_INFORMATION information = ReadInformation(childHandle);
                bool reparse = (information.FileAttributes & FILE_ATTRIBUTE_REPARSE_POINT) != 0;
                bool directoryEntry = (information.FileAttributes & FILE_ATTRIBUTE_DIRECTORY) != 0;
                if (!reparse && directoryEntry) DeleteChildren(child, childHandle, deadlineUtc, ref visited);
                if (!reparse && !directoryEntry && (information.FileAttributes & FILE_ATTRIBUTE_READONLY) != 0)
                {
                    if (!SetFileAttributes(child, information.FileAttributes & ~FILE_ATTRIBUTE_READONLY))
                        throw new Win32Exception(Marshal.GetLastWin32Error());
                }
                MarkDelete(childHandle);
            }
            finally { CloseHandle(childHandle); }
        }
        CheckDeadline(deadlineUtc);
    }

    private static void MarkDelete(IntPtr value)
    {
        FILE_DISPOSITION_INFO disposition = new FILE_DISPOSITION_INFO();
        disposition.DeleteFile = true;
        int length = Marshal.SizeOf(typeof(FILE_DISPOSITION_INFO));
        IntPtr pointer = Marshal.AllocHGlobal(length);
        try
        {
            Marshal.StructureToPtr(disposition, pointer, false);
            if (!SetFileInformationByHandle(value, FILE_DISPOSITION_INFO_CLASS, pointer, (UInt32)length))
                throw new Win32Exception(Marshal.GetLastWin32Error());
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
  }
  foreach ($entry in @($Invocation.Environment)) {
    [System.Environment]::SetEnvironmentVariable([string]$entry.Name, [string]$entry.Value, 'Process')
  }
  $nativeArgs = @($Invocation.Arguments | ForEach-Object { [string]$_ })
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
        & go @goArgs 2>&1 | & $capture
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

  return Invoke-C12Native -Executable 'docker' -Arguments $Arguments -Stage $Stage -Timeout $Timeout -WorkingDirectory $script:c12RepositoryRoot -Deadline $Deadline -AllowFailure:$AllowFailure
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

function Invoke-C12TrustedValidator {
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

    [DateTime]$Deadline = [DateTime]::MaxValue
  )

  if ($Deadline -ne [DateTime]::MaxValue -and [DateTime]::UtcNow -ge $Deadline) {
    throw 'trusted candidate validation exceeded its absolute deadline before bootstrap'
  }
  $resolvedDataRoot = [IO.Path]::GetFullPath($DataRoot).TrimEnd('\')
  if (-not [IO.Directory]::Exists($resolvedDataRoot)) {
    throw 'trusted candidate validation data root is absent'
  }
  $validatorLeaf = "talenro-c12-validator-$(New-C12RandomSuffix)"
  $validatorRoot = Join-Path ([IO.Path]::GetTempPath()) $validatorLeaf
  if ($validatorLeaf -notmatch '^talenro-c12-validator-[0-9a-f]{32}$') {
    throw 'trusted candidate validation generated a malformed bootstrap identity'
  }
  $sourcePath = Join-Path $validatorRoot 'main.go'
  $validatorOwnership = $null
  try {
    $validatorOwnership = New-C12OwnedDirectory -Root $validatorRoot -ExpectedParent ([IO.Path]::GetTempPath()) -LeafPattern '^talenro-c12-validator-[0-9a-f]{32}$' -Stage 'trusted candidate validator creation'
    $encoding = New-Object System.Text.UTF8Encoding($false)
    $sourceBytes = $encoding.GetBytes([string]$script:c12TrustedValidatorSource)
    [IO.File]::WriteAllBytes($sourcePath, $sourceBytes)
    $writtenBytes = [IO.File]::ReadAllBytes($sourcePath)
    $expectedHash = [Security.Cryptography.SHA256]::Create()
    $actualHash = [Security.Cryptography.SHA256]::Create()
    try {
      $expectedDigest = [Convert]::ToBase64String($expectedHash.ComputeHash($sourceBytes))
      $actualDigest = [Convert]::ToBase64String($actualHash.ComputeHash($writtenBytes))
    }
    finally {
      $expectedHash.Dispose()
      $actualHash.Dispose()
    }
    if ($expectedDigest -cne $actualDigest -or $sourceBytes.Length -ne $writtenBytes.Length) {
      throw 'trusted candidate validation bootstrap bytes changed after materialization'
    }
    if ($Deadline -ne [DateTime]::MaxValue -and [DateTime]::UtcNow -ge $Deadline) {
      throw 'trusted candidate validation exceeded its absolute deadline after bootstrap'
    }

    $arguments = @(
      'run', $sourcePath,
      '-mode', $Mode,
      '-version', $script:c12TrustedValidatorVersion,
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
    $result = Invoke-C12Native -Executable 'go' -Arguments $arguments -Stage 'trusted candidate validation' -Timeout ([TimeSpan]::FromMinutes(2)) -WorkingDirectory $validatorRoot -Deadline $Deadline
    $successMarker = "C12_TRUSTED_VALIDATOR_OK:$($script:c12TrustedValidatorVersion)"
    $markers = @($result.Output | Where-Object { [string]$_ -ceq $successMarker })
    if ($markers.Count -ne 1) {
      throw 'trusted candidate validation did not return its exact version marker'
    }
    return $result
  }
  finally {
    Remove-C12BoundedDirectory -Root $validatorRoot -ExpectedParent ([IO.Path]::GetTempPath()) -LeafPattern '^talenro-c12-validator-[0-9a-f]{32}$' -Stage 'trusted candidate validator cleanup' -Deadline $Deadline -Ownership $validatorOwnership
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
  $dockerArgs = @(
    'container', 'inspect', '--format',
    '{{.Id}}|{{.Name}}|{{ index .Config.Labels `talenro.c12.run` }}|{{ index .Config.Labels `talenro.c12.role` }}|{{.Config.Image}}|{{.Image}}',
    [string]$Resource.ID
  )
  $result = Invoke-C12Docker -Arguments $dockerArgs -Stage "inspect $($Resource.Kind) identity"
  $identity = Get-C12SingleOutputLine -Result $result -Stage "inspect $($Resource.Kind) identity"
  $parts = $identity -split '\|', 6
  if ($parts.Count -ne 6 -or $parts[0] -cne [string]$Resource.ID -or $parts[1] -cne "/$($Resource.Name)" -or $parts[2] -cne $RunSuffix) {
    throw "container $($Resource.Kind) captured ID/name/run-label mismatch"
  }
  if ($parts[3] -cne [string]$Resource.Kind) {
    throw "container $($Resource.Kind) role mismatch"
  }
  if ($parts[4] -cne [string]$Resource.ImageRef) {
    throw "container $($Resource.Kind) image reference mismatch"
  }
  if ($parts[5] -cne [string]$Resource.ImageID) {
    throw "container $($Resource.Kind) immutable image ID mismatch"
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

  try {
    $result = Invoke-C12Docker -Arguments $dockerArgs -Stage "start $($Resource.Kind) container"
  }
  catch {
    throw "$($_.Exception.Message); possible orphan name $($Resource.Name) was not adopted because create returned no captured ID"
  }
  $containerID = Get-C12SingleOutputLine -Result $result -Stage "start $($Resource.Kind) container"
  if ($containerID -notmatch '^[0-9a-f]{64}$') {
    throw "start $($Resource.Kind) did not return an exact container ID"
  }
  $Resource.ID = $containerID
  Assert-C12ContainerIdentity -Resource $Resource -RunSuffix $RunSuffix
  Set-C12MappedPort -Resource $Resource
}

function Test-C12TCPProtocol {
  param(
    [Parameter(Mandatory = $true)]
    [int]$Port,

    [Parameter(Mandatory = $true)]
    [string]$Request,

    [Parameter(Mandatory = $true)]
    [string]$Expected
  )

  $client = New-Object System.Net.Sockets.TcpClient
  try {
    $connection = $client.BeginConnect('127.0.0.1', $Port, $null, $null)
    if (-not $connection.AsyncWaitHandle.WaitOne(1000)) {
      return $false
    }
    $client.EndConnect($connection)
    $stream = $client.GetStream()
    $stream.ReadTimeout = 1000
    $stream.WriteTimeout = 1000
    $requestBytes = [System.Text.Encoding]::ASCII.GetBytes($Request)
    $stream.Write($requestBytes, 0, $requestBytes.Length)
    $stream.Flush()
    $buffer = New-Object byte[] 4096
    $response = New-Object System.Text.StringBuilder
    for ($attempt = 0; $attempt -lt 3; $attempt++) {
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
    $client.Close()
  }
}

function Wait-C12Dependencies {
  param(
    [Parameter(Mandatory = $true)]
    [object[]]$Resources
  )

  $postgres = @($Resources | Where-Object { $_.Kind -eq 'postgres' })[0]
  $redis = @($Resources | Where-Object { $_.Kind -eq 'redis' })[0]
  $nats = @($Resources | Where-Object { $_.Kind -eq 'nats' })[0]
  $deadline = [DateTime]::UtcNow.AddSeconds(60)
  while ([DateTime]::UtcNow -lt $deadline) {
    $dockerArgs = @('container', 'inspect', '--format', '{{.State.Health.Status}}', [string]$postgres.ID)
    $healthResult = Invoke-C12Docker -Arguments $dockerArgs -Stage 'inspect PostgreSQL health' -AllowFailure
    $postgresReady = $false
    if ($healthResult.ExitCode -eq 0) {
      $healthLines = @($healthResult.Output | Where-Object { -not [string]::IsNullOrWhiteSpace($_) })
      $postgresReady = $healthLines.Count -eq 1 -and [string]$healthLines[0] -ceq 'healthy'
    }
    $redisReady = Test-C12TCPProtocol -Port $redis.Port -Request "*1`r`n`$4`r`nPING`r`n" -Expected "+PONG`r`n"
    $natsReady = Test-C12TCPProtocol -Port $nats.Port -Request "CONNECT {`"verbose`":false,`"pedantic`":false}`r`nPING`r`n" -Expected "PONG`r`n"
    if ($postgresReady -and $redisReady -and $natsReady) {
      return
    }
    Start-Sleep -Milliseconds 250
  }
  throw 'C12 dependencies did not pass health and protocol probes within 60 seconds'
}

function Resolve-C12CleanupIdentity {
  param(
    [Parameter(Mandatory = $true)]
    [object]$Resource,

    [Parameter(Mandatory = $true)]
    [string]$RunSuffix
  )

  if ([string]::IsNullOrEmpty([string]$Resource.ID)) {
    return $null
  }
  $dockerArgs = @(
    'container', 'inspect', '--format',
    '{{.Id}}|{{.Name}}|{{ index .Config.Labels `talenro.c12.run` }}|{{ index .Config.Labels `talenro.c12.role` }}|{{.Config.Image}}|{{.Image}}',
    [string]$Resource.ID
  )
  $result = Invoke-C12Docker -Arguments $dockerArgs -Stage "re-inspect $($Resource.Kind) before cleanup" -Timeout ([TimeSpan]::FromSeconds(5)) -AllowFailure
  if ($result.ExitCode -ne 0) {
    throw "captured $($Resource.Kind) container is absent before cleanup"
  }
  $identity = Get-C12SingleOutputLine -Result $result -Stage "re-inspect $($Resource.Kind) before cleanup"
  $parts = $identity -split '\|', 6
  if ($parts.Count -ne 6 -or
      $parts[0] -cne [string]$Resource.ID -or
      $parts[1] -cne "/$($Resource.Name)" -or
      $parts[2] -cne $RunSuffix -or
      $parts[3] -cne [string]$Resource.Kind -or
      $parts[4] -cne [string]$Resource.ImageRef -or
      $parts[5] -cne [string]$Resource.ImageID) {
    throw "refusing cleanup for $($Resource.Kind): captured ID/name/run/role/image identity mismatch"
  }
  return [string]$Resource.ID
}

function Remove-C12Container {
  param(
    [Parameter(Mandatory = $true)]
    [object]$Resource,

    [Parameter(Mandatory = $true)]
    [string]$RunSuffix
  )

  $containerID = Resolve-C12CleanupIdentity -Resource $Resource -RunSuffix $RunSuffix
  if ([string]::IsNullOrEmpty([string]$containerID)) {
    return
  }
  $dockerArgs = @('container', 'stop', '--time', '2', [string]$containerID)
  $null = Invoke-C12Docker -Arguments $dockerArgs -Stage "stop exact $($Resource.Kind) container" -Timeout ([TimeSpan]::FromSeconds(5))
  $dockerArgs = @('container', 'rm', [string]$containerID)
  $null = Invoke-C12Docker -Arguments $dockerArgs -Stage "remove exact $($Resource.Kind) container" -Timeout ([TimeSpan]::FromSeconds(5))
  $dockerArgs = @('container', 'inspect', '--format', '{{.Id}}', [string]$containerID)
  $absence = Invoke-C12Docker -Arguments $dockerArgs -Stage "verify $($Resource.Kind) cleanup" -Timeout ([TimeSpan]::FromSeconds(5)) -AllowFailure
  if ($absence.ExitCode -eq 0) {
    throw "$($Resource.Kind) container remains after exact cleanup"
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
    [string[]]$RequestedTests
  )

  $result = Invoke-C12TrustedValidator -Mode 'focused' -DataRoot $script:c12RepositoryRoot -Packages $Packages -Tests $RequestedTests
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
    if ($profile -cne 'base') {
      throw "suite execution-plan group $groupID has unsupported Task 4 profile"
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

    [DateTime]$AbsoluteDeadline = [DateTime]::MaxValue
  )

  if ($GroupProfile -cne 'base') {
    throw 'Task 4 runner accepts only the base profile'
  }
  $groupDuration = ConvertFrom-C12Duration -Value $GroupTimeout
  $groupDeadline = if ($AbsoluteDeadline -eq [DateTime]::MaxValue) { [DateTime]::UtcNow.Add($groupDuration).AddMinutes(3) } else { $AbsoluteDeadline }
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
  $resources = @(
    [pscustomobject]@{ Kind = 'postgres'; Name = "talenro-c12-$runSuffix-postgres"; ID = ''; ImageRef = 'postgres:18.4-alpine3.23'; ImageID = ''; Port = 0; ContainerPort = 5432 },
    [pscustomobject]@{ Kind = 'redis'; Name = "talenro-c12-$runSuffix-redis"; ID = ''; ImageRef = 'redis:8.8.1-alpine3.23'; ImageID = ''; Port = 0; ContainerPort = 6379 },
    [pscustomobject]@{ Kind = 'nats'; Name = "talenro-c12-$runSuffix-nats"; ID = ''; ImageRef = 'nats:2.14.3-alpine3.22'; ImageID = ''; Port = 0; ContainerPort = 4222 }
  )
  $primaryFailure = $null
  $cleanupFailures = @()
  try {
    foreach ($resource in $resources) {
      Resolve-C12ImageIdentity -Resource $resource
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

    $migrationDirectory = Join-Path $script:c12RepositoryRoot 'db\migrations'
    $goArgs = @('tool', 'goose', '-dir', $migrationDirectory, 'postgres', $databaseURL, 'up-to', '6')
    $null = Invoke-C12Go -Arguments $goArgs -Stage "ordinary Goose base migration for $GroupID" -Deadline $groupDeadline

    $goArgs = @('test', '-json', '-tags=integration', '-p=1', '-count=1', '-timeout', $GroupTimeout)
    if (-not [string]::IsNullOrEmpty($RunPattern)) {
      $goArgs += @('-run', $RunPattern)
    }
    $goArgs += $Package
    $testResult = Invoke-C12Go -Arguments $goArgs -Stage "tagged Go test group $GroupID" -Timeout $groupDuration -Deadline $groupDeadline -AllowFailure
    Assert-C12GoJSONResult -Result $testResult -Package $Package -ExpectedTests $ExpectedTests
  }
  catch {
    $primaryFailure = $_.Exception.Message
  }
  finally {
    [System.Environment]::SetEnvironmentVariable('TALENRO_DATABASE_URL', $null, 'Process')
    [System.Environment]::SetEnvironmentVariable('TALENRO_REDIS_ADDRESS', $null, 'Process')
    [System.Environment]::SetEnvironmentVariable('TALENRO_NATS_URL', $null, 'Process')
    for ($index = $resources.Count - 1; $index -ge 0; $index--) {
      try {
        Remove-C12Container -Resource $resources[$index] -RunSuffix $runSuffix
      }
      catch {
        $cleanupFailures += "$($resources[$index].Kind): $($_.Exception.Message)"
      }
    }
  }
  if ($null -ne $primaryFailure) {
    [Console]::Error.WriteLine("C12 test failure [$GroupID]: $primaryFailure")
  }
  foreach ($cleanupFailure in $cleanupFailures) {
    [Console]::Error.WriteLine("C12 cleanup failure [$GroupID]: $cleanupFailure")
  }
  $script:c12NativeDeadline = $priorNativeDeadline
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
  }
}

function Invoke-C12FocusedMode {
  if ($Profile -cne 'base') {
    throw 'Task 4 runner accepts only the base profile'
  }
  if ($Race) {
    throw 'Task 4 base profile does not support -Race while CGO is disabled'
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
    $testMap = Resolve-C12FocusedTestMap -Packages $packageList -RequestedTests $expectedTests
  }
  foreach ($package in $packageList) {
    $groupID = 'focused-' + ($package.TrimStart('.').TrimStart('/').Replace('/', '-'))
    $localTests = @()
    if ($testMap.ContainsKey($package)) {
      $localTests = @($testMap[$package])
    }
    $localRun = if ($localTests.Count -gt 0) { ConvertTo-C12RunPattern -Tests $localTests } else { '' }
    Invoke-C12Group -GroupID $groupID -GroupProfile 'base' -Package $package -RunPattern $localRun -GroupTimeout $Timeout -ExpectedTests $localTests
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

      $null = Invoke-C12TrustedValidator -Mode 'suite' -DataRoot $script:c12RepositoryRoot -SuiteName 'batch01' -SuiteTimeout $Timeout -CandidateTree ([string]$candidate.Tree) -ManifestPaths @($manifestRelativePath) -Deadline $script:c12SuiteDeadline
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
        Invoke-C12Group -GroupID ([string]$group.id) -GroupProfile ([string]$group.profile) -Package ([string]$group.package) -RunPattern $runPattern -GroupTimeout ([string]$group.timeout) -ExpectedTests $tests -AbsoluteDeadline $snapshotGroupDeadline
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
  foreach ($name in $priorGoEnvironment.Keys) {
    [System.Environment]::SetEnvironmentVariable($name, $priorGoEnvironment[$name], 'Process')
  }
}
exit $scriptExitCode
