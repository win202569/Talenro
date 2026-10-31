param(
    [Parameter(Mandatory)][ValidateSet('writer-red','contract-red','writer-green','contract-green','core','lifecycle','continuation')][string]$Grant,
    [Parameter(Mandatory)][ValidatePattern('^[A-F0-9]{64}$')][string]$FrozenSHA256
)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$repo=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))
$packet=Join-Path $repo '.superpowers/sdd/2026-10-31-offline-integrated-review-repair'
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
        return [ordered]@{Type=$Value.Exception.GetType().FullName;Message=$Value.Exception.Message;ErrorId=$Value.FullyQualifiedErrorId;Position=$Value.InvocationInfo.PositionMessage;Stack=$Value.ScriptStackTrace}
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
    & $script:Publisher {param($Path,$Bytes)Write-ODCaptureRecord $Path $Bytes} $Path $bytes
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
    Capture-Once (Join-Path $packet ($Grant+'.evidence.json')) (Capture-Value $value)
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
    $script:CaptureRoot=Join-Path $packet ('capture/'+$Grant)
    if(Test-Path -LiteralPath $script:CaptureRoot){throw 'Capture attempted; no retry'}
    $null=[IO.Directory]::CreateDirectory($script:CaptureRoot)
    $journal=[IO.File]::Open((Join-Path $script:CaptureRoot journal.jsonl),[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::Read)
    Write-ODJournal ([Text.Encoding]::UTF8.GetBytes('{"Stage":"begin"}'+"`n"))
    . (Join-Path $repo 'scripts/offline-diagnostics-v3/private/write-boundary.ps1')
    $oracle=Get-Content -LiteralPath (Join-Path $PSScriptRoot expected-output.json) -Raw | ConvertFrom-Json -AsHashtable
    . (Join-Path $PSScriptRoot recording.tests.ps1)
    $script:RecorderGuard=Start-ODRecorderGuards
    $script:Publisher=Import-Module (Join-Path $PSScriptRoot publisher.psm1) -PassThru -ErrorAction Stop
    . (Join-Path $PSScriptRoot test-helpers.ps1)
    . (Join-Path $PSScriptRoot entry.tests.ps1)
    switch -CaseSensitive ($Grant){
        'writer-red' {. (Join-Path $PSScriptRoot record-boundary.tests.ps1);Invoke-ODWriterTests}
        'writer-green' {. (Join-Path $PSScriptRoot record-boundary.tests.ps1);Invoke-ODWriterTests}
        'contract-red' {. (Join-Path $PSScriptRoot observation-contract.tests.ps1);Invoke-ODContractTests}
        'contract-green' {. (Join-Path $PSScriptRoot observation-contract.tests.ps1);Invoke-ODContractTests}
        'core' {Invoke-ODEntryContract $Grant;Invoke-ODRuntimeContract $Grant}
        'lifecycle' {Invoke-ODEntryContract $Grant;Invoke-ODRuntimeContract $Grant}
        'continuation' {Invoke-ODEntryContract $Grant;Invoke-ODRuntimeContract $Grant}
    }
    Complete-ODRecorderGuards $script:RecorderGuard $script:Publisher $script:CaseModules
    Write-ODJournal ([Text.Encoding]::UTF8.GetBytes('{"Stage":"tests-complete"}'+"`n"))
    $script:CurrentGroup=$null
    $exitCode=if(@($script:Groups | Where-Object {-not $_.Passed}).Count -gt 0){1}else{0}
}catch{$script:Aborted=$true;$exitCode=2;Capture-Failure $_;Journal-ODFailure $_}
finally{try{[Console]::Out.Flush()}catch{Capture-Failure $_;$exitCode=2;Journal-ODFailure $_}}
if($null -ne $script:HostPrimary){Capture-ErrorSummary}
try{Capture-Evidence;Write-ODJournal ([Text.Encoding]::UTF8.GetBytes('{"Stage":"published"}'+"`n"))}catch{Capture-Failure $_;$exitCode=2;Journal-ODFailure $_;Capture-ErrorSummary}
finally{if($null -ne $journal){try{$journal.Dispose()}catch{$exitCode=2}}}
exit $exitCode
