param(
    [Parameter(Mandatory)][ValidateSet('entry-red','entry-green','runtime-red','runtime-green','lifecycle','continuation')][string]$Grant,
    [Parameter(Mandatory)][ValidatePattern('^[A-F0-9]{64}$')][string]$FrozenSHA256
)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$repo=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))
$packet=Join-Path $repo '.superpowers/sdd/2026-10-19-offline-diagnostic-entry-migration'
$script:Groups=[Collections.Generic.List[object]]::new()
$script:Details=[Collections.Generic.List[object]]::new()
$script:ActiveSessions=[Collections.Generic.List[object]]::new()
$script:HostPrimary=$null;$script:HostSecondary=[Collections.Generic.List[Management.Automation.ErrorRecord]]::new()
$script:HostOverflow=$false;$script:CurrentGroup=$null
$script:ErrorOutputAttempted=$false;$script:Aborted=$false
$oracle=Get-Content -LiteralPath (Join-Path $PSScriptRoot expected-output.json) -Raw | ConvertFrom-Json -AsHashtable
$exitCode=2
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
try{
    if($PSVersionTable.PSVersion.ToString() -cne '7.6.5' -or $PSVersionTable.PSEdition -cne 'Core' -or -not [Environment]::Is64BitProcess){throw 'Unexpected PS7 runtime'}
    $identity=[Security.Principal.WindowsIdentity]::GetCurrent()
    $admin=([Security.Principal.WindowsPrincipal]::new($identity)).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
    if($identity.Name -cne 'DESKTOP-H4SVCLU\Lenovo' -or $admin){throw 'Unexpected fixed offline context'}
    . (Join-Path $PSScriptRoot test-helpers.ps1)
    . (Join-Path $PSScriptRoot entry.tests.ps1)
    switch -CaseSensitive ($Grant){
        'entry-red' {Invoke-ODEntryContract $Grant}
        'entry-green' {Invoke-ODEntryContract $Grant}
        'runtime-red' {Invoke-ODRuntimeContract $Grant}
        'runtime-green' {Invoke-ODRuntimeContract $Grant}
        'lifecycle' {Invoke-ODEntryContract $Grant;Invoke-ODRuntimeContract $Grant}
        'continuation' {Invoke-ODEntryContract $Grant;Invoke-ODRuntimeContract $Grant}
    }
    $script:CurrentGroup=$null
    $exitCode=if(@($script:Groups | Where-Object {-not $_.Passed}).Count -gt 0){1}else{0}
}catch{$script:Aborted=$true;$exitCode=2;Capture-Failure $_}
finally{try{[Console]::Out.Flush()}catch{Capture-Failure $_;$exitCode=2}}
if($null -ne $script:HostPrimary){Capture-ErrorSummary}
try{Capture-Evidence}catch{Capture-Failure $_;$exitCode=2;Capture-ErrorSummary}
exit $exitCode
