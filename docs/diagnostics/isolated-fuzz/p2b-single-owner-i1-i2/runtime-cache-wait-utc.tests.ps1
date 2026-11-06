param([Parameter(Mandatory)][ValidateSet('Red','Green','V2Gate')][string]$Phase)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'runtime-cache-wait-common.ps1')
$repo=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))
$root=New-C12WaitRoot (Join-Path $repo ".superpowers/sdd/2026-11-06-c12-cache-wait-utc-fix-v1/$Phase")
$tokens=$null; $errors=$null
$ast=[Management.Automation.Language.Parser]::ParseFile((Join-Path $PSScriptRoot 'runtime-cache-wait-collector.ps1'),[ref]$tokens,[ref]$errors)
if ($errors.Count -ne 0) { throw 'Collector syntax errors' }
# Execute the actual start conversion, not a copied implementation or the live script.
# AST selection isolates this pure expression from process/CIM/filesystem actions.
$assignments=@($ast.FindAll({param($node) $node -is [Management.Automation.Language.AssignmentStatementAst] -and $node.Left.Extent.Text -ceq '$start'},$true))
if ($assignments.Count -ne 1) { throw 'Expected one start conversion' }
$convert=[scriptblock]::Create($assignments[0].Extent.Text + '; $start')
$results=[Collections.Generic.List[object]]::new()
function Case([string]$Name,[scriptblock]$Body) {
    try { & $Body; $results.Add(@{Name=$Name;Status='PASS'}) }
    catch { $results.Add(@{Name=$Name;Status='FAIL';Error=$_.Exception.Message}) }
}
function Check($Condition,[string]$Message) { if (-not $Condition) { throw $Message } }
function Convert-Start($Value) { $target=@{StartedUtc=$Value}; & $convert }
function Reject-Start($Value) {
    $rejected=$false
    try { $null=Convert-Start $Value } catch { $rejected=$true }
    Check $rejected 'Ambiguous/non-UTC start accepted'
}
$originalCulture=[Threading.Thread]::CurrentThread.CurrentCulture
try {
    # The exact saved live timestamp; literal expected ticks independent of conversion.
    $null=Write-C12WaitRecord $root 'target.json' @{StartedUtc='2026-11-06T18:28:02.0520103Z'}
    $targetRecord=Read-C12WaitRecord $root 'target.json'
    foreach ($culture in @('en-US','tr-TR')) {
        Case "json-utc-ticks-$culture" {
            [Threading.Thread]::CurrentThread.CurrentCulture=[Globalization.CultureInfo]::GetCultureInfo($culture)
            $start=Convert-Start $targetRecord.StartedUtc
            Check ($start.Kind -eq [DateTimeKind]::Utc -and $start.Ticks -eq 639295864820520103L) 'UTC instant/subsecond ticks changed'
        }
    }
    Case 'five-second-sample-remains-before-deadline' {
        $start=Convert-Start $targetRecord.StartedUtc
        $now=[DateTime]::new(639295864870520103L,[DateTimeKind]::Utc)
        $elapsed=($now-$start).TotalMilliseconds
        Check ($elapsed -eq 5000 -and (Get-C12WaitDecision $elapsed 120000 $false $true) -eq 'START') 'Premature deadline'
    }
    Case 'deadline-exactly-at-120-seconds' {
        $start=Convert-Start $targetRecord.StartedUtc
        $now=[DateTime]::new(639295866020520103L,[DateTimeKind]::Utc)
        Check ((Get-C12WaitDecision ($now-$start).TotalMilliseconds 120000 $false $true) -eq 'STOP') 'Deadline extended'
    }
    Case 'reject-unspecified-kind' { Reject-Start ([DateTime]::new(639295864820520103L,[DateTimeKind]::Unspecified)) }
    Case 'reject-local-kind' { Reject-Start ([DateTime]::new(639295864820520103L,[DateTimeKind]::Local)) }
    Case 'reject-unparsed-string' { Reject-Start '2026-11-06T18:28:02.0520103' }
    Case 'reject-null' { Reject-Start $null }
} finally { [Threading.Thread]::CurrentThread.CurrentCulture=$originalCulture }
$failed=@($results | Where-Object Status -eq 'FAIL').Count
$receipt=@{Phase=$Phase;Status=$(if ($failed -eq 0) {'UTC_OFFLINE_REGRESSION_VERIFIED'} else {'FAILED'});Count=$results.Count;Failed=$failed;Cases=$results.ToArray();ChildProcessesStarted=0;GoStarted=$false;LiveCollectorStarted=$false}
$null=Write-C12WaitRecord $root 'result.json' $receipt
$receipt | ConvertTo-Json -Depth 10
if ($failed -gt 0) { exit 1 }
