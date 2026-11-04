param(
    [Parameter(Mandatory)][ValidateSet('writer-green','contract-green','core','lifecycle','continuation')][string]$Grant,
    [Parameter(Mandatory)][ValidatePattern('^[A-F0-9]{64}$')][string]$FrozenSHA256,
    [string]$Epoch='legacy'
,
[Parameter(Mandatory)][string]$InputSHA256,
[Parameter(Mandatory)][string]$IntentSHA256
)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$script:HTRepo='D:\Projects\Talenro\.worktrees\c12-b01-task9-coordinator'
$script:HTD=Join-Path $script:HTRepo '.superpowers/sdd/2026-11-04-runtime-copy-transport-v1'
if((Get-Location).Path -cne $script:HTRepo -or $Grant -cne 'contract-green' -or $Epoch -cne 'history-contract-runtime' -or $FrozenSHA256 -cne '186FE2C2B44A31A98EA5406109AC888A625FD1B0061C5338D0382268015D9BBA'){throw 'HT: fixed invocation'}
if((Get-FileHash (Join-Path $script:HTD input.json)).Hash -cne $InputSHA256 -or (Get-FileHash (Join-Path $script:HTD intent.json)).Hash -cne $IntentSHA256){throw 'HT: bootstrap identity'}
$script:HTBinding=Get-Content (Join-Path $script:HTD input.json) -Raw|ConvertFrom-Json -AsHashtable
foreach($htPath in $script:HTBinding.HelperSHA256.Keys){if((Get-FileHash (Join-Path $script:HTRepo $htPath)).Hash -cne $script:HTBinding.HelperSHA256[$htPath]){throw 'HT: helper identity'}}
. (Join-Path $PSScriptRoot runtime-copy-transport-common.ps1)
Initialize-HT $InputSHA256 $IntentSHA256
Write-HTEvent setup begin ''

Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$repo=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))
$packet=Join-Path $repo '.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner'
if($Epoch -cnotin @('legacy','extent-contract','history-contract','history-contract-runtime') -or ($Epoch -cin @('extent-contract','history-contract','history-contract-runtime') -and $Grant -cne 'contract-green')){throw 'OD: invalid capture epoch'}
$attemptSlot=if($Epoch -ceq 'history-contract-runtime'){'history-contract-runtime-host'}elseif($Epoch -ceq 'history-contract'){'history-contract-host'}elseif($Epoch -ceq 'extent-contract'){'extent-contract-green'}elseif($Grant -ceq 'writer-green'){'writer-projection-green'}else{$Grant}
$script:Groups=[Collections.Generic.List[object]]::new()
$script:Details=[Collections.Generic.List[object]]::new()
$script:ActiveSessions=[Collections.Generic.List[object]]::new()
$script:HostPrimary=$null;$script:HostSecondary=[Collections.Generic.List[Management.Automation.ErrorRecord]]::new()
$script:HostOverflow=$false;$script:CurrentGroup=$null
$script:ErrorOutputAttempted=$false;$script:Aborted=$false
$exitCode=2
$script:Publisher=$null;$script:RecorderGuard=$null
$script:CaseModules=[Collections.Generic.List[object]]::new()
$script:EvidenceAttempted=$false
$journal=$null;$journalState=[pscustomobject]@{Failed=$false;Bytes=0}
function Capture-Failure([Management.Automation.ErrorRecord]$Record){
    if($null -eq $Record){return}
    if($null -eq $script:HostPrimary){$script:HostPrimary=$Record;return}
    if([object]::ReferenceEquals($script:HostPrimary,$Record)){return}
    foreach($prior in $script:HostSecondary){if([object]::ReferenceEquals($prior,$Record)){return}}
    if($script:HostSecondary.Count -lt 8){$script:HostSecondary.Add($Record)}else{$script:HostOverflow=$true}
}
function Capture-Value($Value){
    if($null -eq $Value){return $null}
    if($Value -is [Management.Automation.ErrorRecord]){
        return [ordered]@{Type=$Value.Exception.GetType().FullName;Message=$Value.Exception.Message;ErrorId=$Value.FullyQualifiedErrorId;Position=$(if($null -eq $Value.InvocationInfo){$null}else{$Value.InvocationInfo.PositionMessage});Stack=$Value.ScriptStackTrace}
    }
    if($Value -is [Collections.IDictionary]){
        $result=[ordered]@{}
        foreach($key in $Value.Keys){$result[$key]=Capture-Value $Value[$key]}
        return $result
    }
    if($Value -is [array]){
        $result=[Collections.Generic.List[object]]::new()
        foreach($item in $Value){$result.Add((Capture-Value $item))}
        return ,($result.ToArray())
    }
    if($Value -is [Management.Automation.PSCustomObject]){
        $result=[ordered]@{}
        foreach($property in $Value.PSObject.Properties){$result[$property.Name]=Capture-Value $property.Value}
        return $result
    }
    if($Value -is [string] -or $Value -is [ValueType]){return $Value}
    throw 'Unsupported evidence value'
}
function Capture-Once([string]$Path,$Value){
    # One serialization and one CreateNew/write/flush/close. No retry branch.
    $bytes=[Text.UTF8Encoding]::new($false).GetBytes((ConvertTo-Json -InputObject $Value -Depth 50))
    if($bytes.Length -gt 1048576){throw 'Evidence byte limit exceeded'}
    if($script:EvidenceAttempted){throw 'Evidence attempt consumed'}
    $script:EvidenceAttempted=$true
    $outcome=& $script:Publisher {param($Path,$Bytes)Write-ODCaptureRecord $Path $Bytes} $Path $bytes
    $consumption=[pscustomobject]@{Consumed=$false}
    Register-ODRecordOutcome $outcome $consumption
    if(-not $outcome.Succeeded){$script:Aborted=$true}
    return $outcome.Succeeded
}

function Capture-ErrorSummary {
    if($script:ErrorOutputAttempted){return}
    $script:ErrorOutputAttempted=$true
    try{[Console]::Error.WriteLine('ERROR: '+$script:HostPrimary.Exception.Message);[Console]::Error.Flush()}catch{Capture-Failure $_}
}
function Capture-Evidence {
    $completed=@($script:Groups | ForEach-Object {$_.Name})
    $notRun=@($oracle.Grants[$Grant].NotRun)+@($oracle.Grants[$Grant].Groups | Where-Object {$_.Name -cnotin $completed} | ForEach-Object {$_.Name})
    $value=[ordered]@{
        Schema=1;Grant=$Grant;Status=$(if($null -eq $script:HostPrimary -and -not $script:Aborted -and $notRun.Count -eq 0){'COMPLETE'}else{'INCOMPLETE'})
        Groups=$script:Groups.ToArray();CurrentGroup=$script:CurrentGroup;NotRun=$notRun;Observations=$script:Details.ToArray()
        PrimaryError=$script:HostPrimary;SecondaryErrors=$script:HostSecondary.ToArray();Overflow=$script:HostOverflow;Probe=$null
        SourceVerified=$false;LiveActionsPerformed=$false;RuntimeApproved=$false;FormalAcceptance=$false;RealGateImplemented=$false
    }
    Capture-Once (Join-Path $packet ($attemptSlot+'.evidence.json')) (Capture-Value $value)
}
function Write-ODJournal([byte[]]$Bytes){
    if($journalState.Failed){throw 'Journal latched'}
    try{if($journalState.Bytes+$Bytes.Length -gt 1048576){throw 'Journal byte limit'};$journal.Write($Bytes,0,$Bytes.Length);$journal.Flush();$journalState.Bytes+=$Bytes.Length}
    catch{$journalState.Failed=$true;throw}
}
function Journal-ODFailure($Record){
    if($null -eq $journal -or $journalState.Failed){return}
    try{
        $stages=if($null -eq $script:RecorderGuard){@()}else{$script:RecorderGuard.Stages.ToArray()}
        $value=[ordered]@{Stage='unexpected-error';Error=(Capture-Value $Record);RecorderStages=$stages}
        Write-ODJournal ([Text.UTF8Encoding]::new($false).GetBytes((ConvertTo-Json -InputObject $value -Depth 30 -Compress)+"`n"))
    }catch{$journalState.Failed=$true}
}
try{
    if($PSVersionTable.PSVersion.ToString() -cne '7.6.5' -or $PSVersionTable.PSEdition -cne 'Core' -or -not [Environment]::Is64BitProcess){throw 'Unexpected PS7 runtime'}
    $identity=[Security.Principal.WindowsIdentity]::GetCurrent()
    $admin=([Security.Principal.WindowsPrincipal]::new($identity)).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
    if($identity.Name -cne 'DESKTOP-H4SVCLU\Lenovo' -or $admin){throw 'Unexpected fixed offline context'}
    $script:CaptureRoot=Join-Path $script:HTD capture
    if(Test-Path -LiteralPath $script:CaptureRoot){throw 'Capture attempted; no retry'}
    $null=[IO.Directory]::CreateDirectory($script:CaptureRoot)
    $journal=[IO.File]::Open((Join-Path $script:CaptureRoot journal.jsonl),[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::Read)
    Write-ODJournal ([Text.Encoding]::UTF8.GetBytes('{"Stage":"begin"}'+"`n"))
    Write-HTEvent 'import.write-boundary' begin ''
. (Join-Path $repo 'scripts/offline-diagnostics-v4/private/write-boundary.ps1')
Write-HTEvent 'import.write-boundary' end ''
    Write-HTEvent 'import.record-owner' begin ''
. (Join-Path $repo 'scripts/offline-diagnostics-v4/private/record-owner.ps1')
Write-HTEvent 'import.record-owner' end ''
    Write-HTEvent 'oracle' begin ''
$oracle=Get-Content -LiteralPath (Join-Path $PSScriptRoot expected-output.json) -Raw | ConvertFrom-Json -AsHashtable
Write-HTEvent 'oracle' end ''
    Write-HTEvent 'import.recording' begin ''
. (Join-Path $PSScriptRoot recording.tests.ps1)
Write-HTEvent 'import.recording' end ''
    Write-HTEvent 'guard.start' begin ''
$script:RecorderGuard=Start-ODRecorderGuards
Write-HTEvent 'guard.start' end ''
    Write-HTEvent 'publisher.import' begin ''
$script:Publisher=Import-Module (Join-Path $PSScriptRoot publisher.psm1) -PassThru -ErrorAction Stop
Write-HTEvent 'publisher.import' end ''
    . (Join-Path $PSScriptRoot runtime-copy-transport-test-helpers.ps1)
    Write-HTEvent 'import.entry' begin ''
. (Join-Path $PSScriptRoot entry.tests.ps1)
Write-HTEvent 'import.entry' end ''
    Write-HTEvent setup end ''
. (Join-Path $PSScriptRoot runtime-copy-transport-contract.tests.ps1)
Invoke-ODContractTests $Epoch
    if(-not $script:HTCutoff){throw 'HT: first wrapper did not return'}
$script:CurrentGroup=$null
$exitCode=0
}catch{
 $htCaught=$_
 $script:Aborted=$true;$exitCode=2
 Write-HTFailureEvent error.entry $htCaught
 Capture-Failure $htCaught
 Write-HTFailureEvent error.collected $htCaught
 Journal-ODFailure $htCaught
 Write-HTFailureEvent error.journaled $htCaught
}
finally{
 try{Write-HTFailureEvent terminal.flush.begin $null;[Console]::Out.Flush();Write-HTFailureEvent terminal.flush.end $null}
 catch{Capture-Failure $_;$exitCode=2;Journal-ODFailure $_}
}
if($null -ne $script:HostPrimary){Capture-ErrorSummary}
if($null -ne $journal){try{$journal.Dispose()}catch{Capture-Failure $_;$exitCode=2}}
$exitCode=Complete-HT $exitCode $InputSHA256 $IntentSHA256
exit $exitCode
