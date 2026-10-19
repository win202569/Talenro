param(
    [Parameter(Mandatory)][ValidateSet('prefix-red','prefix-green','recording-red','recording-green','owned-green','sequence-green')][string]$Grant,
    [Parameter(Mandatory)][ValidatePattern('^[A-F0-9]{64}$')][string]$FrozenSHA256
)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$repo=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))
$packet=Join-Path $repo '.superpowers/sdd/2026-10-19-p2b-timeout-observation-repair'
$script:Groups=[Collections.Generic.List[object]]::new()
$script:Details=[Collections.Generic.List[object]]::new()
$script:ActiveSessions=[Collections.Generic.List[object]]::new()
$script:HostPrimary=$null
$script:HostSecondary=[Collections.Generic.List[Management.Automation.ErrorRecord]]::new()
$script:HostOverflow=$false
$script:CurrentGroup=$null
$script:ProbeReady=$false;$script:ProbePrimarySame=$false;$script:ProbeObserved=$null
$script:KnownProbe=$false;$script:Aborted=$false;$script:ErrorOutputAttempted=$false
$script:LSModule=$null
$exitCode=2
$oracle=Get-Content -LiteralPath (Join-Path $PSScriptRoot expected-output.json) -Raw | ConvertFrom-Json -AsHashtable
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
    $stream=$null;$primary=$null
    try{
        $stream=[IO.File]::Open($Path,[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::None)
        $stream.Write($bytes,0,$bytes.Length);$stream.Flush()
    }catch{$primary=$_}
    finally{if($null -ne $stream){try{$stream.Dispose()}catch{if($null -eq $primary){$primary=$_}else{Capture-Failure $_}}}}
    if($null -ne $primary){throw $primary}
}
function Capture-ErrorSummary {
    if($script:ErrorOutputAttempted){return}
    $script:ErrorOutputAttempted=$true
    $message=if($script:KnownProbe){'EXPECTED: RecordingProbeAbort'}elseif($null -ne $script:HostPrimary){'ERROR: '+$script:HostPrimary.Exception.Message}else{'ERROR: unknown capture failure'}
    try{[Console]::Error.WriteLine($message);[Console]::Error.Flush()}catch{Capture-Failure $_;$script:KnownProbe=$false}
}
function Capture-Evidence {
    $completedNames=@($script:Groups | ForEach-Object {$_.Name})
    $notRun=if($script:KnownProbe){@('AfterProbe')}else{@($oracle.Grants[$Grant].Groups | Where-Object {$_.Name -cnotin $completedNames} | ForEach-Object {$_.Name})}
    $probe=$null
    if($Grant -cin @('recording-red','recording-green')){
        $view=if($script:ActiveSessions.Count -eq 1){Get-LSView $script:ActiveSessions[0]}else{$null}
        $probe=[ordered]@{Ready=$script:ProbeReady;PrimarySame=$script:ProbePrimarySame;StartCalls=$(if($null -ne $view){$view.StartCalls}else{$null});AckIssued=$(if($null -ne $view){$view.AckIssued}else{$null})}
    }
    $value=[ordered]@{
        Schema=1;Grant=$Grant;Status=$(if($null -eq $script:HostPrimary -and -not $script:Aborted){'COMPLETE'}else{'INCOMPLETE'})
        Groups=$script:Groups.ToArray();CurrentGroup=$script:CurrentGroup;NotRun=@($notRun);Observations=$script:Details.ToArray()
        PrimaryError=$script:HostPrimary;SecondaryErrors=$script:HostSecondary.ToArray();Overflow=$script:HostOverflow;Probe=$probe
        SourceVerified=$false;LiveActionsPerformed=$false;RuntimeApproved=$false;FormalAcceptance=$false;RealGateImplemented=$false
    }
    Capture-Once (Join-Path $packet ($Grant+'.evidence.json')) (Capture-Value $value)
}
try{
    if($PSVersionTable.PSVersion.ToString() -cne '7.6.5' -or $PSVersionTable.PSEdition -cne 'Core' -or -not [Environment]::Is64BitProcess){throw 'Unexpected PS7 runtime'}
    $identity=[Security.Principal.WindowsIdentity]::GetCurrent()
    $admin=([Security.Principal.WindowsPrincipal]::new($identity)).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
    if($identity.Name -cne 'DESKTOP-H4SVCLU\Lenovo' -or $admin){throw 'Unexpected fixed offline context'}
    $script:LSModule=Import-Module (Join-Path $PSScriptRoot session.psm1) -PassThru -ErrorAction Stop
    . (Join-Path $PSScriptRoot core.tests.ps1)
    . (Join-Path $PSScriptRoot prefix.tests.ps1)
    . (Join-Path $PSScriptRoot recording.tests.ps1)
    . (Join-Path $PSScriptRoot owned.tests.ps1)
    . (Join-Path $PSScriptRoot sequence.tests.ps1)
    switch -CaseSensitive ($Grant){
        'prefix-red' {Invoke-LSPrefixTests}
        'prefix-green' {Invoke-LSPrefixTests;Invoke-LSCoreTests}
        'recording-red' {Invoke-LSRecordingProbe}
        'recording-green' {Invoke-LSRecordingProbe}
        'owned-green' {Invoke-LSOwnedTests}
        'sequence-green' {Invoke-LSSequenceTests}
    }
    $script:CurrentGroup=$null
    $exitCode=if(@($script:Groups | Where-Object {-not $_.Passed}).Count -gt 0){1}else{0}
}catch{
    $script:Aborted=$true;$exitCode=2
    $script:KnownProbe=$Grant -cin @('recording-red','recording-green') -and $script:ProbeReady -and $script:ProbePrimarySame -and [object]::ReferenceEquals($script:HostPrimary,$script:ProbeObserved) -and $_.Exception.Message -ceq 'ASSERT: RecordingProbeAbort'
    Capture-Failure $_
}finally{
    # No resources depend on evidence health; all live handles are revoked.
    foreach($handle in $script:ActiveSessions){try{$null=Close-LSSession $handle Stop}catch{Capture-Failure $_;$script:KnownProbe=$false;$exitCode=2}}
    try{[Console]::Out.Flush()}catch{Capture-Failure $_;$script:KnownProbe=$false;$exitCode=2}
}
if($null -ne $script:HostPrimary){Capture-ErrorSummary}
try{
    # One shared publication boundary, including actual assertion aborts.
    # It runs after output and session closure, never as a success retry.
    Capture-Evidence
}catch{Capture-Failure $_;$script:KnownProbe=$false;$exitCode=2;Capture-ErrorSummary}
exit $exitCode
