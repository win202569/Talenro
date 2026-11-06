# Offline characterization/regression: never dot-source the live collector.
param([ValidateSet('Offline','ReasonRed','ReasonGreen','V2Gate')][string]$Phase='Offline')
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'runtime-cache-wait-common.ps1')
$repo=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))
$suiteRoot=New-C12WaitRoot (Join-Path $repo ".superpowers/sdd/2026-11-06-c12-cache-wait-loop-v1/$Phase")
$tokens=$null; $errors=$null
$ast=[Management.Automation.Language.Parser]::ParseFile((Join-Path $PSScriptRoot 'runtime-cache-wait-collector.ps1'),[ref]$tokens,[ref]$errors)
if ($errors.Count -ne 0) { throw 'Collector syntax errors' }
$loops=@($ast.FindAll({param($n) $n -is [Management.Automation.Language.WhileStatementAst] -and $n.Condition.Extent.Text -ceq '$count -lt 24'},$true))
$receipts=@($ast.FindAll({param($n) $n -is [Management.Automation.Language.AssignmentStatementAst] -and $n.Left.Extent.Text -ceq '$receipt'},$true))
if ($loops.Count -ne 1 -or $receipts.Count -ne 1) { throw 'Expected one sampling loop and receipt builder' }
$loop=[scriptblock]::Create($loops[0].Extent.Text)
$builder=[scriptblock]::Create($receipts[0].Extent.Text + '; $receipt')
$results=[Collections.Generic.List[object]]::new()
function Check($Condition,$Message) { if (-not $Condition) { throw $Message } }
function Run-Scenario([string]$Name,[double]$Elapsed,[int]$InitialCount,[bool]$ExecutorDone,[bool]$QueryFails) {
    $root=New-C12WaitRoot (Join-Path $suiteRoot $Name)
    $SessionId='b' * 32
    $target=@{Identity=@{Fixture=$true}}
    $count=$InitialCount; $previous=$null; $next=5000.0
    $start=[DateTime]::UtcNow.AddMilliseconds(-$Elapsed)
    $state=@{Queries=0;Sleeps=0;Sequences=[Collections.Generic.List[int]]::new()}
    $failure=$null; $exitConfirmed=$true; $workerError=$null; $endReason=$null
    # Replace only external read/query/wait boundaries. Publication/deltas stay real.
    function Read-C12WaitRecord($Root,$Leaf) { if ($ExecutorDone) { return @{Status='ATTEMPTED'} }; return $null }
    function Start-Sleep([int]$Milliseconds) {
        $state.Sleeps++
        if ($state.Sleeps -gt 1) { throw 'Unexpected repeated wait' }
        Set-Variable -Scope 1 -Name start -Value ([DateTime]::UtcNow.AddMilliseconds(-5000))
    }
    function Invoke-BoundQuery($Identity,[int]$Sequence,[string]$Phase) {
        $state.Queries++; $state.Sequences.Add($Sequence)
        Check ($Phase -ceq 'sample') 'Wrong query phase'
        if ($QueryFails) { throw 'QUERY_TIMEOUT' }
        # Advance the supplied scenario beyond cutoff after this one response.
        Set-Variable -Scope 1 -Name start -Value ([DateTime]::UtcNow.AddMilliseconds(-120000))
        return @{SessionId=$SessionId;Sequence=$Sequence;Phase=$Phase;Status='OBSERVED';CPUms=10;ReadOperations=1;ReadBytes=100;WriteOperations=0;WriteBytes=0;WorkingSetBytes=1000;PrivateBytes=1000;ThreadStates=@{Wait=1};WaitReasons=@{UserRequest=1};UnavailableFields=@();Identity=$Identity;TimestampUtc='2026-11-06T00:00:00Z'}
    }
    try { . $loop } catch { $failure=$_.Exception.Message }
    $receipt=& $builder
    $null=Write-C12WaitRecord $root 'scenario-result.json' @{Receipt=$receipt;Queries=$state.Queries;Sleeps=$state.Sleeps;Sequences=$state.Sequences.ToArray();NextMs=$next;Synthetic=$true}
    return @{Receipt=$receipt;Queries=$state.Queries;Sleeps=$state.Sleeps;Count=$count;NextMs=$next;Root=$root;Sequences=$state.Sequences.ToArray()}
}
function Case($Name,$Body) { try { & $Body; $results.Add(@{Name=$Name;Status='PASS'}) } catch { $results.Add(@{Name=$Name;Status='FAIL';Error=$_.Exception.Message}) } }
Case 'wait-then-one-sample-no-catchup' {
    $r=Run-Scenario 'wait' 0 0 $false $false
    Check ($r.Sleeps -eq 1 -and $r.Queries -eq 1 -and $r.Count -eq 1 -and $r.NextMs -ge 125000 -and $r.Sequences[0] -eq 1) 'Scheduling/count/spacing failure'
    Check (Test-Path -LiteralPath (Join-Path $r.Root 'sample-01.json')) 'Sample not published'
}
Case 'deadline-does-not-query' { $r=Run-Scenario 'deadline' 120000 0 $false $false; Check ($r.Queries -eq 0 -and $r.Sleeps -eq 0) 'Query after deadline' }
Case 'executor-finished-does-not-query' { $r=Run-Scenario 'finished' 5000 0 $true $false; Check ($r.Queries -eq 0 -and $r.Sleeps -eq 0) 'Query after executor completion' }
Case 'sample-cap-does-not-query' { $r=Run-Scenario 'cap' 5000 24 $false $false; Check ($r.Queries -eq 0 -and $r.Count -eq 24) 'Sample cap exceeded' }
Case 'query-failure-is-stopped-without-retry' {
    $r=Run-Scenario 'query-error' 5000 0 $false $true
    Check ($r.Queries -eq 1 -and $r.Count -eq 0 -and $r.Receipt.Status -ceq 'STOPPED' -and $r.Receipt.Failure -ceq 'QUERY_TIMEOUT') 'Query failure hidden/retried'
}
Case 'deadline-reason-retained' {
    $r=Run-Scenario 'reason-deadline' 120000 0 $false $false
    Check ($r.Receipt.ContainsKey('EndReason') -and $r.Receipt.EndReason -ceq 'DEADLINE') 'Deadline reason missing'
}
Case 'executor-completion-reason-retained' {
    $r=Run-Scenario 'reason-executor' 5000 0 $true $false
    Check ($r.Receipt.ContainsKey('EndReason') -and $r.Receipt.EndReason -ceq 'EXECUTOR_FINISHED') 'Executor completion reason missing'
}
Case 'sample-limit-reason-retained' {
    $r=Run-Scenario 'reason-cap' 5000 24 $false $false
    Check ($r.Receipt.ContainsKey('EndReason') -and $r.Receipt.EndReason -ceq 'SAMPLE_LIMIT') 'Sample limit reason missing'
}
$failed=@($results | Where-Object Status -eq 'FAIL').Count
$result=@{Status=$(if ($failed) {'OFFLINE_COVERAGE_GAPS'} else {'OFFLINE_LOOP_VERIFIED'});Count=$results.Count;Failed=$failed;Cases=$results.ToArray();Synthetic=$true;ChildProcessesStarted=0;GoStarted=$false;LiveCollectorStarted=$false}
$null=Write-C12WaitRecord $suiteRoot 'result.json' $result
$result | ConvertTo-Json -Depth 10
if ($failed) { exit 1 }
