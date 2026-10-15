$ErrorActionPreference='Stop'
function Check-Readiness([bool]$Value,[string]$Name) {if(-not $Value){throw "ASSERTION FAILED: $Name"};Write-Output "PASS: $Name"}
function Reject-Readiness([scriptblock]$Action,[string]$Name) {$refused=$false;try{& $Action|Out-Null}catch{$refused=$true};Check-Readiness $refused $Name}
$helper=Join-Path $PSScriptRoot 'child-readiness.ps1'
if(Test-Path -LiteralPath $helper){. $helper}
Check-Readiness ($null -ne (Get-Command Get-ChildReadinessPlan -ErrorAction SilentlyContinue)) 'readiness fixture exposes a guarded non-launching planner'
$workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..'))
$scope=Join-Path $workspace '.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator'
$fixture=Join-Path $scope ('readiness-tests-'+[guid]::NewGuid().ToString('N'))
$target=Join-Path $fixture 'child-readiness-901'
$plan=Get-ChildReadinessPlan $target
Check-Readiness ($plan.LimitSeconds -eq 2 -and -not (Test-Path -LiteralPath $fixture)) 'planning preserves whole-process two-second bound without creating evidence'
Reject-Readiness {Get-ChildReadinessPlan (Join-Path $fixture 'full-parent-probe-901')} 'old probe identity refused'
Reject-Readiness {Get-ChildReadinessPlan (Join-Path $workspace 'child-readiness-901')} 'outside owned scope refused'
$other=Get-ChildReadinessPlan (Join-Path $fixture 'child-readiness-902')
Check-Readiness ($plan.Nonce -cmatch '^[0-9a-f]{32}$' -and $plan.Nonce -cne $other.Nonce) 'separate plans use separate readiness identities'
$tokens=$null;$errors=$null;$ast=[Management.Automation.Language.Parser]::ParseInput($plan.ChildScript,[ref]$tokens,[ref]$errors)
Check-Readiness (@($errors).Count -eq 0 -and [Text.Encoding]::Unicode.GetString([Convert]::FromBase64String($plan.Arguments[2])) -ceq $plan.ParentScript) 'child parses and parent encoded arguments roundtrip'
New-Item -ItemType Directory -Path $target -Force|Out-Null
Reject-Readiness {Get-ChildReadinessPlan $target} 'consumed readiness root cannot be planned again'
Reject-Readiness {Invoke-ChildReadinessFixture $target} 'duplicate execution rejected before process launch'
# Execute child preparation/write in this host only; never execute its final sleep.
$prefix=($ast.EndBlock.Statements|Select-Object -SkipLast 1|ForEach-Object {$_.Extent.Text}) -join "`n"
& ([scriptblock]::Create($prefix))
$actual=Get-Content -LiteralPath $plan.ReadinessPath -Raw|ConvertFrom-Json
Check-Readiness ($actual.Nonce -ceq $plan.Nonce -and $actual.ProcessId -eq $PID) 'generated child writes its own ID and plan nonce using real file IO'
Reject-Readiness {& ([scriptblock]::Create($prefix))} 'child readiness write cannot overwrite existing evidence'
$receipt=[pscustomobject]@{RootId=17;StartUTC='2026-10-15T00:00:00Z';TimedOut=$true;CleanupVerified=$true;ActiveAfterCleanup=0;MembersAfterCleanup=@();MemberEvidenceTruncated=$false;MemberEvidence=@([pscustomobject]@{Id=18;Owned=$true;State='running';QueryError=0;ExitCodeRead=$true;ExitCode=259;CoreObservedUTC='2026-10-15T00:00:02Z'})}
function Write-ReadinessInput([string]$Nonce=$plan.Nonce,[long]$ProcessId=18,[string]$UTC='2026-10-15T00:00:01Z') {[IO.File]::WriteAllText($plan.ReadinessPath,([pscustomobject]@{Nonce=$Nonce;ProcessId=$ProcessId;UTC=$UTC}|ConvertTo-Json -Compress))}
Write-ReadinessInput
$e=Get-ChildReadinessEvidence $plan $receipt
Check-Readiness ($e.ChildReady -and $e.CleanupVerified -and $e.FixturePassed -and -not $e.Acceptance -and -not $e.OriginalTestPassed) 'matching ready child and timeout cleanup pass only the isolated fixture evaluator'
Write-ReadinessInput ('0'*32)
$e=Get-ChildReadinessEvidence $plan $receipt
Check-Readiness (-not $e.ChildReady -and $e.CleanupVerified -and -not $e.FixturePassed) 'wrong nonce cannot borrow successful cleanup'
Write-ReadinessInput $plan.Nonce 17
Check-Readiness (-not (Get-ChildReadinessEvidence $plan $receipt).ChildReady) 'root cannot impersonate intended child'
Write-ReadinessInput $plan.Nonce 19
Check-Readiness (-not (Get-ChildReadinessEvidence $plan $receipt).ChildReady) 'unobserved process ID rejected regardless of aggregate count'
Write-ReadinessInput
foreach($change in @(@('Owned',$false),@('State','unknown'),@('QueryError',87),@('ExitCodeRead',$false),@('ExitCode',0))) {
    $copy=$receipt|ConvertTo-Json -Depth 5|ConvertFrom-Json
    $copy.MemberEvidence[0].($change[0])=$change[1]
    Check-Readiness (-not (Get-ChildReadinessEvidence $plan $copy).ChildReady) ('unproved child rejected: '+$change[0])
}
Write-ReadinessInput $plan.Nonce 18 '2026-10-15T00:00:02.001Z'
Check-Readiness (-not (Get-ChildReadinessEvidence $plan $receipt).ChildReady) 'late readiness cannot extend the two-second budget'
Write-ReadinessInput $plan.Nonce 18 '2026-10-14T23:59:59Z'
Check-Readiness (-not (Get-ChildReadinessEvidence $plan $receipt).ChildReady) 'pre-run readiness rejected'
Write-ReadinessInput
foreach($change in @(@('CleanupVerified',$false),@('ActiveAfterCleanup',1),@('MembersAfterCleanup',@(18)),@('TimedOut',$false),@('MemberEvidenceTruncated',$true))) {
    $copy=$receipt|ConvertTo-Json -Depth 5|ConvertFrom-Json
    $copy.($change[0])=$change[1]
    Check-Readiness (-not (Get-ChildReadinessEvidence $plan $copy).FixturePassed) ('fixture cannot pass invalid result: '+$change[0])
}
[IO.File]::WriteAllText($plan.ReadinessPath,'{"Nonce":')
Check-Readiness (-not (Get-ChildReadinessEvidence $plan $receipt).ChildReady) 'partial readiness record remains unproved'
$missingPlan=Get-ChildReadinessPlan (Join-Path $fixture 'child-readiness-903')
Check-Readiness (-not (Get-ChildReadinessEvidence $missingPlan $receipt).ChildReady) 'missing readiness record remains missing'
