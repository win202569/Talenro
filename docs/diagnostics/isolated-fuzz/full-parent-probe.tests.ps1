$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'tree-timeout-probe.ps1')
function Check-FullParent([bool]$Value,[string]$Name) {if(-not $Value){throw "ASSERTION FAILED: $Name"};Write-Output "PASS: $Name"}
function Reject-FullParent([scriptblock]$Action,[string]$Name) {$refused=$false;try{& $Action|Out-Null}catch{$refused=$true};Check-FullParent $refused $Name}
$workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..'))
$scope=Join-Path $workspace '.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator'
$fixture=Join-Path $scope ('full-parent-tests-'+[guid]::NewGuid().ToString('N'))
$target=Join-Path $fixture 'full-parent-probe-901'
$plan=$null;try{$plan=Get-TreeTimeoutProbePlan $target $true}catch{}
Check-FullParent ($null -ne $plan -and $plan.LimitSeconds -eq 2) 'explicit synchronous full-parent mode accepts only its new identity with original budget'
Check-FullParent (-not (Test-Path -LiteralPath $fixture)) 'planning launches no timed fixture and writes no directory'
Reject-FullParent {Get-TreeTimeoutProbePlan $target} 'default tree mode cannot execute full-parent identity'
Reject-FullParent {Get-TreeTimeoutProbePlan (Join-Path $fixture 'tree-timeout-probe-901') $true} 'synchronous mode cannot reuse tree diagnostic identity'
$tokens=$null;$errors=$null;$ast=[Management.Automation.Language.Parser]::ParseInput($plan.ParentScript,[ref]$tokens,[ref]$errors)
Check-FullParent (@($errors).Count -eq 0) 'generated complete parent parses without launching child'
Check-FullParent ([Text.Encoding]::Unicode.GetString([Convert]::FromBase64String($plan.Arguments[2])) -ceq $plan.ParentScript) 'complete encoded parent survives argument roundtrip'
# Execute only the two entry-output statements in this host, not the child code.
$prefix=($ast.EndBlock.Statements|Select-Object -First 2|ForEach-Object {$_.Extent.Text}) -join ';'
$stream=New-Object IO.MemoryStream;$writer=New-Object IO.StreamWriter($stream);$previous=[Console]::Out;$captured=$null
try {[Console]::SetOut($writer);& ([scriptblock]::Create($prefix));$captured=[Text.Encoding]::UTF8.GetString($stream.ToArray())}finally{[Console]::SetOut($previous);$writer.Dispose();$stream.Dispose()}
Check-FullParent ($captured.Contains('TALENRO_FULL_PARENT_ENTRY')) 'full parent flushes entry before first file operation or child launch'
New-Item -ItemType Directory -Path (Join-Path $target 'process') -Force|Out-Null
Reject-FullParent {Get-TreeTimeoutProbePlan $target $true} 'existing full-parent identity refused'
Reject-FullParent {Invoke-TreeTimeoutProbe $target $true} 'execution refuses duplicate before any timed launch'
$receipt=[pscustomobject]@{TimedOut=$true;TotalProcesses=1;CleanupVerified=$true;ActiveAfterCleanup=0;MembersAfterCleanup=@()}
$e=Get-TreeTimeoutProbeEvidence $target $receipt
Check-FullParent ($e.StdoutStages.Count -eq 0 -and -not $e.Acceptance -and -not $e.OriginalTestPassed) 'missing stdout stages are not invented or accepted'
[IO.File]::WriteAllText((Join-Path $target 'process/stdout.txt'),"TALENRO_FULL_PARENT_ENTRY`r`nTALENRO_FULL_PARENT_BEFORE_START`r`nTALENRO_FULL_PARENT_AFTER_START-extra`r`n")
$e=Get-TreeTimeoutProbeEvidence $target $receipt
Check-FullParent (($e.StdoutStages -join '|') -ceq 'TALENRO_FULL_PARENT_ENTRY|TALENRO_FULL_PARENT_BEFORE_START' -and $e.Markers.Count -eq 0 -and -not $e.OriginalTestPassed) 'only exact observed stdout stages retained without return or child inference'
[IO.File]::WriteAllText((Join-Path $target 'process/stdout.txt'),"TALENRO_FULL_PARENT_ENTRY`r`nTALENRO_FULL_PARENT_BEFORE_START`r`nTALENRO_FULL_PARENT_AFTER_START`r`n")
$e=Get-TreeTimeoutProbeEvidence $target $receipt
Check-FullParent ($e.StdoutStages.Count -eq 3 -and $e.Markers.Count -eq 0 -and -not $e.Acceptance -and -not $e.OriginalTestPassed) 'all parent stdout stages still cannot prove child entry or original-test success'
