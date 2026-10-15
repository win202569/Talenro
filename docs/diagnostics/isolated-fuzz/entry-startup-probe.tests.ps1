$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'entry-startup-probe.ps1')
function Check-EntryProbe([bool]$Value,[string]$Name) {if(-not $Value){throw "ASSERTION FAILED: $Name"};Write-Output "PASS: $Name"}
function Reject-EntryProbe([scriptblock]$Action,[string]$Name) {$refused=$false;try{& $Action|Out-Null}catch{$refused=$true};Check-EntryProbe $refused $Name}
$workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..'))
$scope=Join-Path $workspace '.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator'
$fixture=Join-Path $scope ('entry-plan-tests-'+[guid]::NewGuid().ToString('N'))
$target=Join-Path $fixture 'entry-startup-probe-901'
$plan=Get-EntryStartupProbePlan $target
Check-EntryProbe ($plan.LimitSeconds -eq 2) 'entry probe retains two-second whole-process budget'
Check-EntryProbe (-not (Test-Path -LiteralPath $fixture)) 'planning creates no receipt directory or observed process'
Reject-EntryProbe {Get-EntryStartupProbePlan (Join-Path $workspace 'outside')} 'outside evidence scope refused'
Reject-EntryProbe {Get-EntryStartupProbePlan (Join-Path $fixture 'tree-timeout-probe-901')} 'consumed tree probe identity cannot become entry probe'
Check-EntryProbe ([Text.Encoding]::Unicode.GetString([Convert]::FromBase64String($plan.Arguments[2])) -ceq $plan.EntryScript) 'encoded argument roundtrip preserves complete entry program'
New-Item -ItemType Directory -Path (Join-Path $target 'process') -Force|Out-Null
Reject-EntryProbe {Get-EntryStartupProbePlan $target} 'existing entry identity refused'
Reject-EntryProbe {Invoke-EntryStartupProbe $target} 'execution refuses existing identity before process creation'
# Run the generated program in this unit-test host, never a timed child/probe.
# Read the real buffered stream before Dispose can flush it on our behalf.
$stream=New-Object IO.MemoryStream
$writer=New-Object IO.StreamWriter($stream)
$previous=[Console]::Out;$captured=$null
try {
    [Console]::SetOut($writer)
    & ([scriptblock]::Create($plan.EntryScript))
    $captured=[Text.Encoding]::UTF8.GetString($stream.ToArray())
} finally {[Console]::SetOut($previous);$writer.Dispose();$stream.Dispose()}
Check-EntryProbe ($captured.Contains('TALENRO_ENTRY_PROBE_ENTERED')) 'entry program flushes marker into real buffered stream before disposal'
Check-EntryProbe (Test-Path -LiteralPath (Join-Path $target 'script-entry.utc')) 'same-host program writes its separate timestamp marker'
$receipt=[pscustomobject]@{TimedOut=$true;TotalProcesses=1;CleanupVerified=$true;ActiveAfterCleanup=0;MembersAfterCleanup=@()}
[IO.File]::WriteAllText((Join-Path $target 'process/stdout.txt'),'TALENRO_ENTRY_PROBE_ENTERED-extra')
$e=Get-EntryStartupProbeEvidence $target $receipt
Check-EntryProbe (-not $e.StdoutEntryObserved -and $e.FileMarkerObserved -and -not $e.OriginalTestPassed -and -not $e.Acceptance) 'file-only evidence and suffix text cannot invent exact flushed stdout entry'
[IO.File]::WriteAllText((Join-Path $target 'process/stdout.txt'),"TALENRO_ENTRY_PROBE_ENTERED`r`n")
$e=Get-EntryStartupProbeEvidence $target $receipt
Check-EntryProbe ($e.StdoutEntryObserved -and $e.FileMarkerObserved -and -not $e.OriginalTestPassed -and -not $e.Acceptance) 'both markers remain diagnostic and never pass original test'
$empty=Join-Path $fixture 'entry-startup-probe-902'
New-Item -ItemType Directory -Path (Join-Path $empty 'process') -Force|Out-Null
$e=Get-EntryStartupProbeEvidence $empty $receipt
Check-EntryProbe (-not $e.StdoutEntryObserved -and -not $e.FileMarkerObserved -and -not $e.Acceptance) 'absent output and file remain absent without startup inference'
