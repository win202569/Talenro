$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'tree-timeout-probe.ps1')
function Check-TreeProbe([bool]$Value,[string]$Name) {if(-not $Value){throw "ASSERTION FAILED: $Name"};Write-Output "PASS: $Name"}
function Reject-TreeProbe([scriptblock]$Action,[string]$Name) {$refused=$false;try{& $Action|Out-Null}catch{$refused=$true};Check-TreeProbe $refused $Name}
$workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..'))
$scope=Join-Path $workspace '.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator'
$root=Join-Path $scope ('tree-plan-tests-'+[guid]::NewGuid().ToString('N'))
$target=Join-Path $root 'tree-timeout-probe-901'
$plan=Get-TreeTimeoutProbePlan $target
Check-TreeProbe ($plan.LimitSeconds -eq 2) 'probe retains two-second whole-process bound'
Check-TreeProbe (-not (Test-Path -LiteralPath $root)) 'planning creates no run root and launches no fixture'
Reject-TreeProbe {Get-TreeTimeoutProbePlan (Join-Path $workspace 'outside')} 'outside owned root refused'
Reject-TreeProbe {Get-TreeTimeoutProbePlan (Join-Path $root 'envelope-observe-901')} 'original sampling identity cannot be used as diagnostic probe'
$tokens=$null;$errors=$null;[Management.Automation.Language.Parser]::ParseInput($plan.ParentScript,[ref]$tokens,[ref]$errors)|Out-Null
Check-TreeProbe (@($errors).Count -eq 0) 'generated parent script parses without executing it'
New-Item -ItemType Directory -Path $target -Force|Out-Null
Reject-TreeProbe {Get-TreeTimeoutProbePlan $target} 'existing probe identity refused without retry'
Reject-TreeProbe {Invoke-TreeTimeoutProbe $target} 'execution refuses existing root before process launch'
$receipt=[pscustomobject]@{TimedOut=$true;TotalProcesses=1;CleanupVerified=$true;ActiveAfterCleanup=0;MembersAfterCleanup=@();TerminationRequested=$true}
$evidence=Get-TreeTimeoutProbeEvidence $target $receipt
Check-TreeProbe (-not $evidence.Acceptance -and -not $evidence.OriginalTestPassed -and $evidence.Markers.Count -eq 0) 'missing markers remain missing and cannot promote a clean timeout to acceptance'
[IO.File]::WriteAllText((Join-Path $target 'parent-entry.utc'),'2026-10-15T00:00:00Z')
[IO.File]::WriteAllText((Join-Path $target 'parent-before-start.utc'),'2026-10-15T00:00:01Z')
$evidence=Get-TreeTimeoutProbeEvidence $target $receipt
Check-TreeProbe (($evidence.Markers -join '|') -ceq 'parent-entry.utc|parent-before-start.utc' -and -not $evidence.OriginalTestPassed) 'only actual stage files are reported without inferring a child launch'
