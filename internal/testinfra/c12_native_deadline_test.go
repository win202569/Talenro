package testinfra_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestC12NativeWatchdogPreservesSubsecondBudget(t *testing.T) {
	c12RunHarnessInParallel(t)
	runner, err := os.ReadFile("../../scripts/run-c12-integration.ps1")
	if err != nil {
		t.Fatal(err)
	}
	marker := []byte("$script:c12RepositoryRoot = (Resolve-Path")
	if bytes.Count(runner, marker) != 1 {
		t.Fatal("runner main-program marker is not unique")
	}
	prefix, _, _ := bytes.Cut(runner, marker)
	// Replace only PowerShell's job transport. The production native watchdog
	// and real Windows Job handle still execute, without process-startup noise.
	const appendix = `
$script:c12DeadlineEvents = [Collections.Generic.List[string]]::new()
$script:c12DeadlineMode = ''

function Start-Job {
  param([object[]]$ArgumentList, [scriptblock]$ScriptBlock)
  $script:c12DeadlineEvents.Add('start')
  if ($script:c12DeadlineMode -ceq 'slow-start') { [Threading.Thread]::Sleep(200) }
  if ($script:c12DeadlineMode -ceq 'late-completion') {
    # A scheduling-delayed waiter may return true after its absolute deadline.
    # Force that result without running a real native child or relying on races.
    $finished = [pscustomobject]@{}
    $finished | Add-Member -MemberType ScriptMethod -Name WaitOne -Value {
      param([int]$Milliseconds)
      $script:c12DeadlineEvents.Add('wait')
      [Threading.Thread]::Sleep(200)
      return $true
    }
    $finished | Add-Member -MemberType ScriptMethod -Name Dispose -Value {}
    return [pscustomobject]@{ Finished = $finished }
  }
  $completed = $script:c12DeadlineMode -cne 'unfinished'
  return [pscustomobject]@{ Finished = [Threading.ManualResetEvent]::new($completed) }
}

function Receive-Job {
  param([object]$Job, [string]$ErrorAction)
  $script:c12DeadlineEvents.Add('receive')
  return [pscustomobject]@{ ExitCode = 0; Output = @('native-result'); ContainmentFailure = $false }
}

function Stop-Job {
  param([object]$Job, [string]$ErrorAction)
  $script:c12DeadlineEvents.Add('stop')
}

function Remove-Job {
  param([object]$Job, [switch]$Force, [string]$ErrorAction)
  $script:c12DeadlineEvents.Add('remove')
  $Job.Finished.Dispose()
}

try {
  $script:c12DeadlineMode = 'completed'
  $result = Invoke-C12Native -Executable 'docker' -Arguments @('version') -Stage 'subsecond completed job' -Timeout ([TimeSpan]::FromMilliseconds(500)) -WorkingDirectory (Get-Location).Path -Deadline ([DateTime]::UtcNow.AddSeconds(5))
  if ($result.ExitCode -ne 0 -or @($result.Output).Count -ne 1 -or $result.Output[0] -cne 'native-result') { throw 'completed subsecond job did not return its native result' }
  if (($script:c12DeadlineEvents -join ',') -cne 'start,receive,stop,remove') { throw ('completed subsecond job lifecycle: ' + ($script:c12DeadlineEvents -join ',')) }
  Write-Output 'C12_SUBSECOND_COMPLETED_OK'

  $script:c12DeadlineMode = 'unfinished'
  $script:c12DeadlineEvents.Clear()
  $failure = ''
  $watch = [Diagnostics.Stopwatch]::StartNew()
  try { $null = Invoke-C12Native -Executable 'docker' -Arguments @('version') -Stage 'unfinished bounded job' -Timeout ([TimeSpan]::FromMilliseconds(200)) -WorkingDirectory (Get-Location).Path -Deadline ([DateTime]::UtcNow.AddSeconds(5)) }
  catch { $failure = $_.Exception.Message }
  $watch.Stop()
  if ($failure -notlike 'unfinished bounded job timed out*') { throw ('unfinished job did not time out: ' + $failure) }
  # These literal bounds check that a 200 ms budget is actually waited, rather
  # than rejected immediately or rounded up to a whole second.
  if ($watch.Elapsed.TotalMilliseconds -lt 150 -or $watch.Elapsed.TotalMilliseconds -gt 600) { throw ('unfinished 200 ms job elapsed milliseconds: ' + $watch.Elapsed.TotalMilliseconds) }
  if (($script:c12DeadlineEvents -join ',') -cne 'start,stop,remove') { throw ('unfinished job received a result or skipped cleanup: ' + ($script:c12DeadlineEvents -join ',')) }
  Write-Output 'C12_SUBSECOND_UNFINISHED_OK'

  $script:c12DeadlineMode = 'slow-start'
  $script:c12DeadlineEvents.Clear()
  $failure = ''
  try { $null = Invoke-C12Native -Executable 'docker' -Arguments @('version') -Stage 'expired startup job' -Timeout ([TimeSpan]::FromMilliseconds(100)) -WorkingDirectory (Get-Location).Path -Deadline ([DateTime]::UtcNow.AddSeconds(5)) }
  catch { $failure = $_.Exception.Message }
  if ($failure -cne 'expired startup job timed out before native execution') { throw ('completed job after expired startup was accepted or misclassified: ' + $failure) }
  if (($script:c12DeadlineEvents -join ',') -cne 'start,stop,remove') { throw ('expired startup job received a result or skipped cleanup: ' + ($script:c12DeadlineEvents -join ',')) }
  Write-Output 'C12_SUBSECOND_EXPIRED_STARTUP_OK'

  $script:c12DeadlineMode = 'unfinished'
  $script:c12DeadlineEvents.Clear()
  $failure = ''
  $watch = [Diagnostics.Stopwatch]::StartNew()
  try { $null = Invoke-C12Native -Executable 'docker' -Arguments @('version') -Stage 'earlier absolute deadline' -Timeout ([TimeSpan]::FromSeconds(2)) -WorkingDirectory (Get-Location).Path -Deadline ([DateTime]::UtcNow.AddMilliseconds(200)) }
  catch { $failure = $_.Exception.Message }
  $watch.Stop()
  if ($failure -notlike 'earlier absolute deadline timed out*') { throw ('earlier absolute deadline did not time out: ' + $failure) }
  if ($watch.Elapsed.TotalMilliseconds -lt 150 -or $watch.Elapsed.TotalMilliseconds -gt 600) { throw ('200 ms absolute deadline with 2 second Timeout elapsed milliseconds: ' + $watch.Elapsed.TotalMilliseconds) }
  if (($script:c12DeadlineEvents -join ',') -cne 'start,stop,remove') { throw ('earlier absolute deadline received a result or skipped cleanup: ' + ($script:c12DeadlineEvents -join ',')) }
  Write-Output 'C12_SUBSECOND_ABSOLUTE_DEADLINE_OK'

  $script:c12DeadlineMode = 'late-completion'
  $script:c12DeadlineEvents.Clear()
  $failure = ''
  try { $null = Invoke-C12Native -Executable 'docker' -Arguments @('version') -Stage 'late completed job' -Timeout ([TimeSpan]::FromSeconds(2)) -WorkingDirectory (Get-Location).Path -Deadline ([DateTime]::UtcNow.AddMilliseconds(100)) }
  catch { $failure = $_.Exception.Message }
  if ($failure -notlike 'late completed job timed out*') { throw ('true completion after the absolute deadline was accepted: ' + $failure) }
  if (($script:c12DeadlineEvents -join ',') -cne 'start,wait,stop,remove') { throw ('late completed job received a result or skipped cleanup: ' + ($script:c12DeadlineEvents -join ',')) }
  Write-Output 'C12_SUBSECOND_LATE_COMPLETION_OK'
  exit 0
}
catch {
  [Console]::Error.WriteLine($_.Exception.Message)
  exit 1
}
`
	root := t.TempDir()
	script := filepath.Join(root, "scripts", "run-c12-integration.ps1")
	if err := os.MkdirAll(filepath.Dir(script), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(script, append(append([]byte(nil), prefix...), []byte(appendix)...), 0o600); err != nil {
		t.Fatal(err)
	}
	output, exitCode := runC12PowerShellAtRoot(t, root, nil, "-Profile", "base", "-Packages", "./internal/testinfra", "-Timeout", "3m")
	if exitCode != 0 {
		t.Fatalf("production subsecond native watchdog harness exit=%d output=%q", exitCode, output)
	}
	for _, marker := range []string{"C12_SUBSECOND_COMPLETED_OK", "C12_SUBSECOND_UNFINISHED_OK", "C12_SUBSECOND_EXPIRED_STARTUP_OK", "C12_SUBSECOND_ABSOLUTE_DEADLINE_OK", "C12_SUBSECOND_LATE_COMPLETION_OK"} {
		if !strings.Contains(output, marker) {
			t.Errorf("production subsecond native watchdog harness omitted %s: %q", marker, output)
		}
	}
}

func TestC12ReadinessNormalizesOnlyNativeDeadlineErrors(t *testing.T) {
	c12RunHarnessInParallel(t)
	runner, err := os.ReadFile("../../scripts/run-c12-integration.ps1")
	if err != nil {
		t.Fatal(err)
	}
	marker := []byte("$script:c12RepositoryRoot = (Resolve-Path")
	if bytes.Count(runner, marker) != 1 {
		t.Fatal("runner main-program marker is not unique")
	}
	prefix, _, _ := bytes.Cut(runner, marker)
	const appendix = `
$script:c12ReadinessFailure = ''
$script:c12ReadinessCalls = 0
function Invoke-C12Docker {
  param([string[]]$Arguments, [string]$Stage, [DateTime]$Deadline, [switch]$AllowFailure)
  $script:c12ReadinessCalls++
  if ($Stage -cne 'inspect PostgreSQL health' -or $Deadline.Ticks -ne $script:c12NativeDeadline.Ticks) { throw 'readiness lost its exact health stage or earlier enclosing deadline' }
  throw $script:c12ReadinessFailure
}

$resources = @(
  [pscustomobject]@{ Kind = 'postgres'; ID = ('1' * 64); Port = 1 },
  [pscustomobject]@{ Kind = 'redis'; ID = ('2' * 64); Port = 1 },
  [pscustomobject]@{ Kind = 'nats'; ID = ('3' * 64); Port = 1 }
)
$cases = @(
  [pscustomobject]@{ Name = 'bounded wait'; Message = 'inspect PostgreSQL health timed out after its bounded native wait'; Expected = 'C12 dependencies did not pass health and protocol probes within 1 seconds' },
  [pscustomobject]@{ Name = 'submillisecond remainder'; Message = 'inspect PostgreSQL health timed out with less than one bounded millisecond remaining'; Expected = 'C12 dependencies did not pass health and protocol probes within 1 seconds' },
  [pscustomobject]@{ Name = 'identity mismatch'; Message = 'inspect PostgreSQL health identity mismatch'; Expected = 'inspect PostgreSQL health identity mismatch' }
)
try {
  foreach ($case in $cases) {
    $script:c12ReadinessFailure = $case.Message
    $script:c12ReadinessCalls = 0
    $script:c12NativeDeadline = [DateTime]::UtcNow.AddMilliseconds(750)
    $failure = ''
    try { Wait-C12Dependencies -Resources $resources -ProbeTimeout ([TimeSpan]::FromSeconds(1)) }
    catch { $failure = $_.Exception.Message }
    if ($failure -cne $case.Expected) { throw ($case.Name + ' readiness error = ' + $failure + '; expected ' + $case.Expected) }
    if ($script:c12ReadinessCalls -ne 1) { throw ($case.Name + ' did not exercise exactly one health inspection') }
  }
  Write-Output 'C12_READINESS_NATIVE_DEADLINES_ONLY_OK'
  exit 0
}
catch {
  [Console]::Error.WriteLine($_.Exception.Message)
  exit 1
}
`
	root := t.TempDir()
	script := filepath.Join(root, "scripts", "run-c12-integration.ps1")
	if err := os.MkdirAll(filepath.Dir(script), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(script, append(append([]byte(nil), prefix...), []byte(appendix)...), 0o600); err != nil {
		t.Fatal(err)
	}
	output, exitCode := runC12PowerShellAtRoot(t, root, nil, "-Profile", "base", "-Packages", "./internal/testinfra", "-Timeout", "3m")
	if exitCode != 0 || !strings.Contains(output, "C12_READINESS_NATIVE_DEADLINES_ONLY_OK") {
		t.Fatalf("production readiness deadline normalization harness exit=%d output=%q", exitCode, output)
	}
}
