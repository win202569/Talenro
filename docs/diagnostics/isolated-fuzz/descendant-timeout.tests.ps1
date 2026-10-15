$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'child-readiness.ps1')
$helper=Join-Path $PSScriptRoot 'descendant-timeout.ps1'
if(Test-Path -LiteralPath $helper){. $helper}
function Check-Descendant([bool]$Value,[string]$Name) {if(-not $Value){throw "ASSERTION FAILED: $Name"};Write-Output "PASS: $Name"}
function Reject-Descendant([scriptblock]$Action,[string]$Name) {$rejected=$false;try{& $Action|Out-Null}catch{$rejected=$true};Check-Descendant $rejected $Name}
Check-Descendant ($null -ne (Get-Command Get-DescendantTimeoutResult -CommandType Function -ErrorAction SilentlyContinue)) 'revised regression exposes its real shared evaluator'
$workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..'))
$fixture=Join-Path $workspace ('.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator/descendant-unit-'+[guid]::NewGuid().ToString('N'))
$plan=Get-ChildReadinessPlan (Join-Path $fixture 'file-child-readiness-901') $true
New-Item -ItemType Directory -Path $plan.RunRoot -Force|Out-Null
function Write-DescendantMarker([string]$Nonce=$plan.Nonce,[long]$ProcessId=18,[string]$UTC='2026-10-15T00:00:01Z') {[IO.File]::WriteAllText($plan.ReadinessPath,([pscustomobject]@{Nonce=$Nonce;ProcessId=$ProcessId;UTC=$UTC}|ConvertTo-Json -Compress))}
$receipt=[pscustomobject]@{RootId=17;StartUTC='2026-10-15T00:00:00Z';TotalProcesses=2;TimedOut=$true;CleanupVerified=$true;ActiveAfterCleanup=0;MembersAfterCleanup=@();MemberEvidenceTruncated=$false;MemberEvidence=@([pscustomobject]@{Id=18;Owned=$true;State='running';QueryError=0;ExitCodeRead=$true;ExitCode=259;CoreObservedUTC='2026-10-15T00:00:02Z'})}
Write-DescendantMarker
$result=Get-DescendantTimeoutResult $plan $receipt
Check-Descendant ($result.RegressionPassed -and $result.FixtureEvidence.ChildReady -and $result.TotalProcesses -eq 2) 'ready owned child with all original numeric timeout cleanup conditions passes only this case'
Check-Descendant ($result.DiagnosticOnly -and -not $result.Acceptance -and -not $result.OriginalTestPassed) 'single regression result cannot claim full original acceptance'
foreach($change in @(@('TotalProcesses',1),@('TotalProcesses',0),@('TotalProcesses','two'),@('TimedOut',$false),@('CleanupVerified',$false),@('ActiveAfterCleanup',1),@('MembersAfterCleanup',@(18)),@('MemberEvidenceTruncated',$true))) {
    $copy=$receipt|ConvertTo-Json -Depth 6|ConvertFrom-Json;$copy.($change[0])=$change[1]
    Check-Descendant (-not (Get-DescendantTimeoutResult $plan $copy).RegressionPassed) ('original/strict condition still required: '+$change[0]+'='+[string]$change[1])
}
Write-DescendantMarker ('0'*32)
Check-Descendant (-not (Get-DescendantTimeoutResult $plan $receipt).RegressionPassed) 'aggregate count and cleanup cannot substitute for correct child nonce'
Write-DescendantMarker $plan.Nonce 17
Check-Descendant (-not (Get-DescendantTimeoutResult $plan $receipt).RegressionPassed) 'parent cannot stand in for target child'
Write-DescendantMarker $plan.Nonce 18 '2026-10-15T00:00:02.001Z'
Check-Descendant (-not (Get-DescendantTimeoutResult $plan $receipt).RegressionPassed) 'late readiness cannot relax original two-second window'
Write-DescendantMarker
$copy=$receipt|ConvertTo-Json -Depth 6|ConvertFrom-Json;$copy.MemberEvidence[0].Owned=$false
Check-Descendant (-not (Get-DescendantTimeoutResult $plan $copy).RegressionPassed) 'matching marker still requires native ownership proof'
$missing=Get-ChildReadinessPlan (Join-Path $fixture 'file-child-readiness-902') $true
Check-Descendant (-not (Get-DescendantTimeoutResult $missing $receipt).RegressionPassed) 'missing target marker remains unproved'
$legacy=Get-ChildReadinessPlan (Join-Path $fixture 'child-readiness-901')
Reject-Descendant {Get-DescendantTimeoutResult $legacy $receipt} 'revised case explicitly requires file mode'
Reject-Descendant {Invoke-DescendantTimeoutRegression $plan.RunRoot} 'existing case identity refused before any native launch'
$tokens=$null;$errors=$null;$ast=[Management.Automation.Language.Parser]::ParseFile((Join-Path $PSScriptRoot 'run.tests.ps1'),[ref]$tokens,[ref]$errors)
Check-Descendant (@($errors).Count -eq 0 -and @($ast.ParamBlock.Parameters|Where-Object {$_.Name.VariablePath.UserPath -ceq 'DescendantTimeoutOnly'}).Count -eq 1) 'runner exposes a parsed single-case switch without executing the suite'
