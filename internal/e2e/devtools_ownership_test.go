//go:build e2e

package e2e

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// Loading the helper must be inert. Initializing it must put this exact host in
// its non-inherited private Job without launching a compiler. A child that also
// initializes the helper exercises nesting rather than merely creating a child.
func testDevtoolsOwnership(t *testing.T) {
	t.Helper()
	if runtime.GOOS != "windows" {
		return
	}
	t.Run("ownership-boundary", func(t *testing.T) {
		root := filepath.Join(t.TempDir(), "ownership fixture")
		copyDevtoolsProcessHelper(t, root)
		entry := filepath.Join(root, "scripts", "ownership-test.ps1")
		if err := os.WriteFile(entry, []byte(devtoolsOwnershipScript), 0600); err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, devtoolsPowerShell(t), "-NoProfile", "-NonInteractive", "-File", entry)
		cmd.WaitDelay = 500 * time.Millisecond
		output, err := cmd.CombinedOutput()
		if err != nil || strings.TrimSpace(string(output)) != "ownership-boundary: passed" {
			t.Fatalf("real ownership boundary: %v: %s", err, output)
		}
	})
}

const devtoolsOwnershipScript = `param([switch]$Nested)
$ErrorActionPreference = 'Stop'
$source = 'devtools-owned-init-' + [guid]::NewGuid().ToString('N')
$child = $null
try {
  if (-not $Nested) {
    Register-CimIndicationEvent -Namespace root/cimv2 -Query "SELECT * FROM Win32_ProcessStartTrace WHERE ParentProcessID = $PID" -SourceIdentifier $source | Out-Null
  }
  . (Join-Path $PSScriptRoot 'private/devtools-process.ps1')
  if ($null -ne ('DevtoolsProcessOwnership' -as [type])) { throw 'helper loading initialized a type' }
  Initialize-DevtoolsProcessOwnership
  Add-Type -TypeDefinition @'
using System;
using System.Runtime.InteropServices;
public static class DevtoolsOwnershipQuery {
  [DllImport("kernel32.dll", SetLastError=true)] public static extern bool IsProcessInJob(IntPtr process, IntPtr job, out bool member);
  [DllImport("kernel32.dll", SetLastError=true)] public static extern bool GetHandleInformation(IntPtr handle, out uint flags);
}
'@
  $job = [IntPtr][DevtoolsProcessOwnership].GetField('job', [Reflection.BindingFlags]'Static,NonPublic').GetValue($null)
  $member = $false
  $flags = [uint32]0
  if (-not [DevtoolsOwnershipQuery]::IsProcessInJob([IntPtr](-1), $job, [ref]$member) -or -not $member) { throw 'host outside its exact Job' }
  if (-not [DevtoolsOwnershipQuery]::GetHandleInformation($job, [ref]$flags) -or ($flags -band 1)) { throw 'Job handle inherited' }
  if ($Nested) { [Console]::Out.WriteLine('nested: passed'); exit 0 }
  $null = Wait-Event -SourceIdentifier $source -Timeout 2
  if (@(Get-Event -SourceIdentifier $source -ErrorAction SilentlyContinue).Count -ne 0) { throw 'initialization launched a child' }
  $child = [Diagnostics.Process]::new()
  $child.StartInfo.FileName = [Environment]::ProcessPath
  $child.StartInfo.UseShellExecute = $false
  $child.StartInfo.CreateNoWindow = $true
  $child.StartInfo.RedirectStandardOutput = $true
  $child.StartInfo.RedirectStandardError = $true
  foreach ($arg in @('-NoProfile','-NonInteractive','-File',$PSCommandPath,'-Nested')) { $child.StartInfo.ArgumentList.Add($arg) }
  $null = $child.Start()
  $null = $child.Handle
  $output = $child.StandardOutput.ReadToEndAsync()
  $errors = $child.StandardError.ReadToEndAsync()
  $null = Wait-Event -SourceIdentifier $source -Timeout 3
  $observed = @(Get-Event -SourceIdentifier $source -ErrorAction SilentlyContinue | Where-Object { $_.SourceEventArgs.NewEvent.ProcessID -eq $child.Id })
  if ($observed.Count -eq 0) { throw 'process-event positive control missing' }
  if (-not $child.WaitForExit(5000)) { throw 'nested Job timed out' }
  if ($child.ExitCode -ne 0 -or $output.GetAwaiter().GetResult().Trim() -cne 'nested: passed' -or $errors.GetAwaiter().GetResult().Length -ne 0) { throw 'nested Job initialization failed' }
  [Console]::Out.WriteLine('ownership-boundary: passed')
} finally {
  if ($null -ne $child) {
    if (-not $child.HasExited) { $child.Kill($true); $null = $child.WaitForExit(1000) }
    $child.Dispose()
  }
  if (-not $Nested) {
    Unregister-Event -SourceIdentifier $source -ErrorAction SilentlyContinue
    Get-Event -SourceIdentifier $source -ErrorAction SilentlyContinue | Remove-Event -ErrorAction SilentlyContinue
  }
}
`
