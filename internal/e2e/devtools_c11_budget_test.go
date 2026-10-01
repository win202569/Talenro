//go:build e2e

package e2e

import (
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

// Execute the real stage calls and forwarding function, observing the timeout
// delivered to the external-process boundary. Never launch C11 dependencies.
func TestDevtoolsC11StageBudgets(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-only PowerShell entry")
	}
	source, err := filepath.Abs(filepath.Join("..", "..", "scripts", "verify-c11.ps1"))
	if err != nil {
		t.Fatal(err)
	}
	harness := filepath.Join(t.TempDir(), "stage-budgets.ps1")
	if err := os.WriteFile(harness, []byte(devtoolsC11BudgetHarness), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, devtoolsPowerShell(t), "-NoProfile", "-NonInteractive", "-File", harness, "-ScriptPath", source)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("stage budget harness: %v: %s", err, output)
	}
	var got []struct {
		Stage   string
		Seconds int
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(string(output))), &got); err != nil {
		t.Fatalf("stage results: %v: %s", err, output)
	}
	want := []struct {
		Stage   string
		Seconds int
	}{
		{"verify-c11: devtools verification", 3660},
		{"verify-c11: check tools", 4200},
		{"verify-c11: generate", 4200},
		{"verify-c11: unit tests", 600},
	}
	if len(got) != len(want) {
		t.Fatalf("stage count: got %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("stage %d: got %+v, want %+v", i, got[i], want[i])
		}
	}
}

const devtoolsC11BudgetHarness = `param([string]$ScriptPath)
$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
$tokens = $null
$errors = $null
$ast = [System.Management.Automation.Language.Parser]::ParseFile($ScriptPath, [ref]$tokens, [ref]$errors)
if ($errors.Count) { throw 'source parse failed' }
$stage = @($ast.FindAll({ param($n) $n -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $n.Name -eq 'Invoke-C11Stage' }, $true))
$main = @($ast.FindAll({ param($n) $n -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $n.Name -eq 'Invoke-C11Main' }, $true))
if ($stage.Count -ne 1 -or $main.Count -ne 1) { throw 'function resolution failed' }
Invoke-Expression $stage[0].Extent.Text
$script:observed = [Collections.Generic.List[object]]::new()
function Invoke-C11External {
  param([string]$Stage, [string]$FilePath, [string[]]$ArgumentList, [string]$WorkingDirectory, [ValidateRange(1,4200)][int]$TimeoutSeconds = 600)
  $script:observed.Add([pscustomobject]@{Stage=$Stage; Seconds=$TimeoutSeconds})
}
$devtoolsPowerShell = [Environment]::ProcessPath
$scriptsDirectory = Split-Path $ScriptPath
$verifyArguments = @('-NoProfile','-NonInteractive','-File',(Join-Path (Split-Path $ScriptPath) 'verify-devtools.ps1'))
$calls = @($main[0].Body.FindAll({ param($n) $n -is [System.Management.Automation.Language.CommandAst] -and $n.GetCommandName() -eq 'Invoke-C11Stage' }, $true))
if ($calls.Count -lt 4) { throw 'missing stages' }
foreach ($call in $calls[0..3]) { $null = & ([scriptblock]::Create('$PSScriptRoot = $scriptsDirectory; ' + $call.Extent.Text)) }
ConvertTo-Json -InputObject @($script:observed.ToArray()) -Compress
`
