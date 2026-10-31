param(
    [Parameter(Mandatory)][ValidateSet('isolation-red','isolation-green')][string]$Grant,
    [Parameter(Mandatory)][ValidatePattern('^[A-F0-9]{64}$')][string]$FrozenSHA256
)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$riRepo=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))
$riPacket=Join-Path $riRepo '.superpowers/sdd/2026-10-31-offline-record-isolation-repair'
$riPhase=if($Grant -ceq 'isolation-red'){'seed'}else{'isolation-green'}
$riFrozen=Join-Path $riPacket ($riPhase+'.frozen.json')
if((Get-FileHash -LiteralPath $riFrozen).Hash -cne $FrozenSHA256){throw 'Frozen identity mismatch'}
if($PSVersionTable.PSVersion.ToString() -cne '7.6.5' -or $PSVersionTable.PSEdition -cne 'Core' -or -not [Environment]::Is64BitProcess){throw 'Pinned runtime required'}
$riIdentity=[Security.Principal.WindowsIdentity]::GetCurrent()
if($riIdentity.Name -cne 'DESKTOP-H4SVCLU\Lenovo' -or ([Security.Principal.WindowsPrincipal]::new($riIdentity)).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)){throw 'Pinned non-admin identity required'}
$riPhaseRoot=Join-Path $riPacket $Grant
if(Test-Path -LiteralPath $riPhaseRoot){throw 'Phase attempted; no retry'}
$null=[IO.Directory]::CreateDirectory($riPhaseRoot)
$script:RIGroups=[Collections.Generic.List[object]]::new()
$script:RIObservations=$null
$script:HostPrimary=$null;$script:HostSecondary=[Collections.Generic.List[Management.Automation.ErrorRecord]]::new();$script:HostOverflow=$false
$riWriter=Join-Path $riRepo 'scripts/offline-diagnostics-v2/private/write-boundary.ps1'
$riCapture=Join-Path $riRepo 'docs/diagnostics/isolated-fuzz/p2b-review-repair/capture.ps1'
$riTokens=$null;$riParseErrors=$null
$riAst=[Management.Automation.Language.Parser]::ParseFile($riCapture,[ref]$riTokens,[ref]$riParseErrors)
$riCollectors=@($riAst.FindAll({param($n)$n -is [Management.Automation.Language.FunctionDefinitionAst] -and $n.Name -ceq 'Capture-Failure'},$false))
if($riParseErrors.Count -ne 0 -or $riCollectors.Count -ne 1){throw 'Fixed collector source required'}
$riCollectorText=$riCollectors[0].Extent.Text
. $riWriter
. ([ScriptBlock]::Create($riCollectorText))
$riOracle=Get-Content -LiteralPath (Join-Path $PSScriptRoot expected-output.json) -Raw|ConvertFrom-Json -AsHashtable
function Add-RIGroup([string]$Name,[bool]$Passed,$Facts){
    $script:RIGroups.Add([ordered]@{Name=$Name;Passed=$Passed;Facts=$Facts})
    [Console]::Out.WriteLine(($(if($Passed){'PASS '}else{'FAIL '}))+$Name)
}
function Convert-RIValue($Value){
    if($null -eq $Value){return $null}
    if($Value -is [Management.Automation.ErrorRecord]){return [ordered]@{Type=$Value.Exception.GetType().FullName;Message=$Value.Exception.Message;ErrorId=$Value.FullyQualifiedErrorId;Position=$Value.InvocationInfo.PositionMessage;Stack=$Value.ScriptStackTrace;Source=$Value.InvocationInfo.ScriptName;Line=$Value.InvocationInfo.ScriptLineNumber;Column=$Value.InvocationInfo.OffsetInLine}}
    if($Value -is [Collections.IDictionary]){$r=[ordered]@{};foreach($k in $Value.Keys){$r[$k]=Convert-RIValue $Value[$k]};return $r}
    if($Value -is [array]){$r=[Collections.Generic.List[object]]::new();foreach($v in $Value){$r.Add((Convert-RIValue $v))};return ,($r.ToArray())}
    if($Value -is [Management.Automation.PSCustomObject]){$r=[ordered]@{};foreach($p in $Value.PSObject.Properties){$r[$p.Name]=Convert-RIValue $p.Value};return $r}
    if($Value -is [string] -or $Value -is [ValueType]){return $Value}
    throw 'Unsupported evidence value'
}
$riJournal=$null;$riJournalState=[pscustomobject]@{Failed=$false;Bytes=0};$riCode=2
function Write-RIJournal($Stream,$State,[byte[]]$Bytes){
    if($State.Failed){throw 'Journal latched'}
    try{if($State.Bytes+$Bytes.Length -gt 1048576){throw 'Journal byte limit'};$Stream.Write($Bytes,0,$Bytes.Length);$Stream.Flush();$State.Bytes+=$Bytes.Length}
    catch{$State.Failed=$true;throw}
}
try{
    $riJournal=[IO.File]::Open((Join-Path $riPhaseRoot journal.jsonl),[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::Read)
    $riPublisherModule=Import-Module (Join-Path $PSScriptRoot publisher.psm1) -PassThru -ErrorAction Stop
    $riPublishers=@{}
    foreach($riCaseName in @('HostEvidence','Healthy','ExistingFailure')){$riPublishers[$riCaseName]=& $riPublisherModule {param($G,$C)New-RIPublisher $G $C} $Grant $riCaseName}
    foreach($riFile in @('case-existing.bin','existing-publisher.bin')){
        $riStream=$null
        try{$riStream=[IO.File]::Open((Join-Path $riPhaseRoot $riFile),[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::None);$riPrior=[Text.Encoding]::ASCII.GetBytes('PRIOR');$riStream.Write($riPrior,0,$riPrior.Length);$riStream.Flush()}finally{if($null -ne $riStream){$riStream.Dispose()}}
    }
    Write-RIJournal $riJournal $riJournalState ([Text.Encoding]::UTF8.GetBytes('{"Stage":"begin"}'+"`n"))
    . (Join-Path $PSScriptRoot isolation.tests.ps1)
    Invoke-RIIsolationTests $riPublisherModule $riPublishers # RI-CALL
    Write-RIJournal $riJournal $riJournalState ([Text.Encoding]::UTF8.GetBytes('{"Stage":"tests-complete"}'+"`n"))
    [Console]::Out.Flush()
    $riEvidence=[ordered]@{Schema=1;Grant=$Grant;Status='COMPLETE';Groups=$script:RIGroups.ToArray();NotRun=@();Observations=$script:RIObservations;PrimaryError=$null;SecondaryErrors=@();Overflow=$false;SourceVerified=$false;LiveActionsPerformed=$false;RuntimeApproved=$false;FormalAcceptance=$false;RealGateImplemented=$false}
    $riEvidenceBytes=[Text.UTF8Encoding]::new($false).GetBytes((ConvertTo-Json -InputObject (Convert-RIValue $riEvidence) -Depth 35))
    if($riEvidenceBytes.Length -gt 1048576){throw 'Evidence byte limit'}
    $riPublished=& $riPublisherModule {param($Publisher,$Bytes)Write-RIPublisherBytes $Publisher $Bytes} $riPublishers.HostEvidence $riEvidenceBytes # RI-CALL
    if(-not $riPublished.Ok -or $riPublished.Rejected -or $null -ne $riPublished.Error){if($null -ne $riPublished.Error){throw $riPublished.Error}else{throw 'Evidence not published'}}
    Write-RIJournal $riJournal $riJournalState ([Text.Encoding]::UTF8.GetBytes('{"Stage":"published"}'+"`n"))
    $riCode=if(@($script:RIGroups|Where-Object {-not $_.Passed}).Count -eq 0){0}else{1}
}catch{
    $riError=$_
    if($null -ne $riJournal -and -not $riJournalState.Failed){
        try{$riJournalBytes=[Text.UTF8Encoding]::new($false).GetBytes((ConvertTo-Json -InputObject ([ordered]@{Stage='unexpected-error';Error=(Convert-RIValue $riError)}) -Depth 12 -Compress)+"`n");Write-RIJournal $riJournal $riJournalState $riJournalBytes}catch{$riJournalState.Failed=$true}
    }
}finally{if($null -ne $riJournal){try{$riJournal.Dispose()}catch{$riCode=2}}}
exit $riCode
