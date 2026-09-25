//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// Resolves an actual runtime once per caller; never falls back to Windows
// PowerShell or a machine-private Codex installation path.
func devtoolsPowerShell(t *testing.T) string {
	t.Helper()
	path, err := exec.LookPath("pwsh.exe")
	if err != nil {
		t.Fatal("PowerShell 7.6.5 environment unavailable: ", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, path, "-NoProfile", "-NonInteractive", "-Command", "$PSVersionTable.PSVersion.ToString(); $PSVersionTable.PSEdition")
	output, err := cmd.CombinedOutput()
	if err != nil || strings.Join(strings.Fields(string(output)), " ") != "7.6.5 Core" {
		t.Fatalf("actual PowerShell 7.6.5 Core required: %v: %q", err, output)
	}
	return path
}

func copyDevtoolsProcessHelper(t *testing.T, root string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "scripts", "private", "devtools-process.ps1"))
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(root, "scripts", "private")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "devtools-process.ps1"), data, 0600); err != nil {
		t.Fatal(err)
	}
}

// A missing/late version guard must expose the old module-stage failure instead
// of the required early runtime rejection. All input scripts are private copies.
func testDevtoolsRuntime(t *testing.T) {
	t.Helper()
	if runtime.GOOS != "windows" {
		return
	}
	for _, tc := range []struct {
		name, injected   string
		legacy, accepted bool
	}{
		{name: "runtime-5.1-rejected", legacy: true},
		{name: "runtime-7.6.5-accepted", accepted: true},
		{name: "runtime-other-version-rejected", injected: "@{PSEdition='Core'; PSVersion='7.6.4'}"},
		{name: "runtime-prerelease-rejected", injected: "@{PSEdition='Core'; PSVersion='7.6.5-preview.1'}"},
		{name: "runtime-later-version-rejected", injected: "@{PSEdition='Core'; PSVersion='7.6.6'}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			content, err := os.ReadFile(filepath.Join("..", "..", "scripts", "verify-devtools.ps1"))
			if err != nil {
				t.Fatal(err)
			}
			if tc.injected != "" {
				const read = "$devtoolsRuntime = $PSVersionTable"
				if bytes.Count(content, []byte(read)) != 1 {
					t.Fatal("runtime-input fixture must replace one read")
				}
				content = bytes.Replace(content, []byte(read), []byte("$devtoolsRuntime = "+tc.injected), 1)
			}
			root := filepath.Join(t.TempDir(), "repository with spaces")
			private := filepath.Join(root, "scripts", "private")
			if err := os.MkdirAll(private, 0700); err != nil {
				t.Fatal(err)
			}
			entry := filepath.Join(root, "scripts", "verify-devtools.ps1")
			if tc.legacy {
				content = append([]byte(devtoolsRuntimeTracePrefix), content...)
				content = append(content, []byte(devtoolsRuntimeTraceSuffix)...)
			}
			if err := os.WriteFile(entry, content, 0600); err != nil {
				t.Fatal(err)
			}
			marker := filepath.Join(root, "helper-loaded")
			if !tc.accepted {
				if err := os.WriteFile(filepath.Join(private, "devtools-process.ps1"), []byte("[IO.File]::WriteAllText($env:DEVTOOLS_HELPER_MARKER, 'loaded')\nthrow 'private-helper-canary'\n"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			host := filepath.Join(os.Getenv("SystemRoot"), "System32", "WindowsPowerShell", "v1.0", "powershell.exe")
			if !tc.legacy {
				host = devtoolsPowerShell(t)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, host, "-NoProfile", "-NonInteractive", "-File", entry)
			tracePath := filepath.Join(root, "runtime-trace.json")
			cmd.Env = devtoolsFixtureEnv(map[string]string{"DEVTOOLS_HELPER_MARKER": marker, "DEVTOOLS_RUNTIME_TRACE": tracePath})
			output, err := cmd.CombinedOutput()
			want := "verify-devtools: PowerShell 7.6.5 required."
			if tc.accepted {
				want = "verify-devtools: module failed with exit code 1."
			}
			if ctx.Err() != nil || err == nil || cmd.ProcessState == nil || cmd.ProcessState.ExitCode() != 1 || strings.TrimSpace(string(output)) != want {
				t.Fatalf("runtime contract: want %q / exit 1, got %v / %q", want, err, output)
			}
			if _, err := os.Stat(marker); !os.IsNotExist(err) {
				t.Fatalf("rejected runtime loaded private helper: %v", err)
			}
			if tc.legacy {
				data, err := os.ReadFile(tracePath)
				if err != nil {
					t.Fatal("runtime process evidence missing: ", err)
				}
				var trace struct {
					ChildCount      int
					PositiveControl bool
				}
				if err := json.Unmarshal(data, &trace); err != nil {
					t.Fatal(err)
				}
				if trace.ChildCount != 0 || !trace.PositiveControl {
					t.Fatalf("legacy rejection launched children or observer was blind: %+v", trace)
				}
			}
		})
	}
}

// Test-copy instrumentation only: finally still executes when the real entry
// exits. Count children before starting the observer's deliberate control.
const devtoolsRuntimeTracePrefix = `
$ErrorActionPreference = 'Stop'
$runtimeTraceSource = 'devtools-runtime-' + [guid]::NewGuid().ToString('N')
Register-CimIndicationEvent -Namespace root/cimv2 -Query "SELECT * FROM Win32_ProcessStartTrace WHERE ParentProcessID = $PID" -SourceIdentifier $runtimeTraceSource | Out-Null
try {
`

const devtoolsRuntimeTraceSuffix = `
} finally {
  $runtimeControl = $null
  try {
    $null = Wait-Event -SourceIdentifier $runtimeTraceSource -Timeout 2
    $runtimeChildren = @(Get-Event -SourceIdentifier $runtimeTraceSource -ErrorAction SilentlyContinue)
    $runtimeChildren | Remove-Event
    $runtimeControl = Start-Process -FilePath (Get-Process -Id $PID).Path -ArgumentList @('-NoProfile','-NonInteractive','-Command','exit 0') -PassThru -WindowStyle Hidden
    $null = Wait-Event -SourceIdentifier $runtimeTraceSource -Timeout 3
    $runtimeSeen = @(Get-Event -SourceIdentifier $runtimeTraceSource -ErrorAction SilentlyContinue | Where-Object { $_.SourceEventArgs.NewEvent.ProcessID -eq $runtimeControl.Id }).Count -gt 0
    $runtimeRecord = @{ChildCount=$runtimeChildren.Count; PositiveControl=$runtimeSeen} | ConvertTo-Json -Compress
    [IO.File]::WriteAllText($env:DEVTOOLS_RUNTIME_TRACE, $runtimeRecord)
  } finally {
    if ($null -ne $runtimeControl) {
      if (-not $runtimeControl.HasExited) { $runtimeControl.Kill(); $runtimeControl.WaitForExit() }
      $runtimeControl.Dispose()
    }
    Unregister-Event -SourceIdentifier $runtimeTraceSource -ErrorAction SilentlyContinue
    Get-Event -SourceIdentifier $runtimeTraceSource -ErrorAction SilentlyContinue | Remove-Event -ErrorAction SilentlyContinue
  }
}
`

func testDevtoolsBashRejection(t *testing.T) {
	t.Helper()
	for _, interpreter := range []string{`C:\Program Files\Git\bin\bash.exe`, `C:\Program Files\Git\usr\bin\bash.exe`} {
		t.Run(filepath.Base(filepath.Dir(interpreter)), func(t *testing.T) {
			if _, err := os.Stat(interpreter); err != nil {
				t.Fatal("required Windows Bash refusal environment unavailable: ", err)
			}
			root := filepath.Join(t.TempDir(), "repository with spaces")
			if err := os.MkdirAll(filepath.Join(root, "scripts"), 0700); err != nil {
				t.Fatal(err)
			}
			content, err := os.ReadFile(filepath.Join("..", "..", "scripts", "verify-devtools.sh"))
			if err != nil {
				t.Fatal(err)
			}
			// Observe whether the first work-stage statement is reached. No tools
			// are permitted even when the module is absent.
			const begin = "stage=module"
			if bytes.Count(content, []byte(begin)) != 1 {
				t.Fatal("expected one work-stage start")
			}
			content = bytes.Replace(content, []byte(begin), []byte("printf reached > \"${DEVTOOLS_WORK_MARKER}\"\n"+begin), 1)
			entry := filepath.Join(root, "scripts", "verify-devtools.sh")
			if err := os.WriteFile(entry, content, 0600); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, interpreter, filepath.ToSlash(entry))
			cmd.Dir = root
			cmd.Env = devtoolsFixtureEnv(map[string]string{"DEVTOOLS_WORK_MARKER": "work-reached"})
			output, err := cmd.CombinedOutput()
			if ctx.Err() != nil || err == nil || cmd.ProcessState == nil || cmd.ProcessState.ExitCode() != 1 || strings.TrimSpace(string(output)) != "verify-devtools: Windows requires PowerShell 7.6.5." {
				t.Fatalf("expected early Windows rejection, got %v: %q", err, output)
			}
			if _, err := os.Stat(filepath.Join(root, "work-reached")); !os.IsNotExist(err) {
				t.Fatalf("unsupported entry reached work stage: %v", err)
			}
		})
	}
}
