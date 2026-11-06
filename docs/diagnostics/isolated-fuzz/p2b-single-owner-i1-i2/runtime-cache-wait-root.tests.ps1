param([Parameter(Mandatory)][ValidateSet('RootRed','RootGreen')][string]$Phase)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'runtime-cache-wait-common.ps1')
$repo=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))
$root=New-C12WaitRoot (Join-Path $repo ".superpowers/sdd/2026-11-06-c12-cache-wait-v2/$Phase")
$expected=[IO.Path]::GetFullPath((Join-Path $repo '.superpowers/sdd/2026-11-06-c12-cache-wait-v2/Live'))
$results=[Collections.Generic.List[object]]::new()
function Check($Condition,$Message) { if (-not $Condition) { throw $Message } }
function Case($Name,$Body) { try { & $Body; $results.Add(@{Name=$Name;Status='PASS'}) } catch { $results.Add(@{Name=$Name;Status='FAIL';Error=$_.Exception.Message}) } }
function Resolve-RoleRoot([string]$Role) {
    $tokens=$null; $errors=$null
    $ast=[Management.Automation.Language.Parser]::ParseFile((Join-Path $PSScriptRoot "runtime-cache-wait-$Role.ps1"),[ref]$tokens,[ref]$errors)
    if ($errors.Count) { throw 'ROLE_PARSE_ERROR' }
    $assignments=@($ast.FindAll({param($n) $n -is [Management.Automation.Language.AssignmentStatementAst] -and $n.Left.Extent.Text -ceq '$root'},$true))
    if ($assignments.Count -ne 1) { throw 'ROOT_ASSIGNMENT_COUNT' }
    $joins=@($assignments[0].Right.FindAll({param($n) $n -is [Management.Automation.Language.CommandAst] -and $n.GetCommandName() -ceq 'Join-Path'},$true))
    if ($joins.Count -ne 1) { throw 'ROOT_JOIN_COUNT' }
    # Evaluate only the actual Join-Path expression, never New-C12WaitRoot/launches.
    return (& ([scriptblock]::Create($joins[0].Extent.Text)))
}
Case 'executor-resolves-fresh-v2' { Check ((Resolve-RoleRoot 'executor') -ceq $expected) 'Executor still targets old evidence root' }
Case 'collector-resolves-fresh-v2' { Check ((Resolve-RoleRoot 'collector') -ceq $expected) 'Collector still targets old evidence root' }
Case 'occupied-root-preserves-sentinel' {
    $occupied=New-C12WaitRoot (Join-Path $root 'occupied')
    $record=Write-C12WaitRecord $occupied 'sentinel.json' @{Value='preserve-original'}
    $rejected=$false
    try { $null=New-C12WaitRoot $occupied } catch { $rejected=$_.Exception.Message -match 'OCCUPIED' }
    Check ($rejected -and (Get-FileHash -LiteralPath $record.Path).Hash -ceq $record.SHA256) 'Occupied root accepted or sentinel changed'
}
Case 'fresh-root-exclusive-claim' {
    $fresh=New-C12WaitRoot (Join-Path $root 'fresh')
    Check (Test-Path -LiteralPath (Join-Path $fresh 'exclusive.claim')) 'Exclusive claim missing'
    $rejected=$false
    try { $null=New-C12WaitRoot $fresh } catch { $rejected=$_.Exception.Message -match 'OCCUPIED' }
    Check $rejected 'Second claim accepted'
}
$failed=@($results | Where-Object Status -eq 'FAIL').Count
$receipt=@{Phase=$Phase;Status=$(if ($failed) {'ROOT_GATE_FAILED'} else {'ROOT_GATE_VERIFIED'});Count=$results.Count;Failed=$failed;Cases=$results.ToArray();ChildProcessesStarted=0;LiveRootCreated=(Test-Path -LiteralPath $expected)}
$null=Write-C12WaitRecord $root 'result.json' $receipt
$receipt | ConvertTo-Json -Depth 10
if ($failed) { exit 1 }
