param(
    [Parameter(Mandatory)][ValidateSet('core-red','core-green','owned-red','owned-green','sequence-green')][string]$Grant,
    [Parameter(Mandatory)][ValidatePattern('^[A-F0-9]{64}$')][string]$FrozenSHA256
)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$repo=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))
$packet=Join-Path $repo '.superpowers/sdd/2026-10-19-p2b-live-completion-supervisor'
$script:Groups=[Collections.Generic.List[object]]::new()
$script:Details=[Collections.Generic.List[object]]::new()
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
    $bytes=[Text.UTF8Encoding]::new($false).GetBytes((ConvertTo-Json -InputObject $Value -Depth 50))
    $stream=$null;$primary=$null
    try{
        $stream=[IO.File]::Open($Path,[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::None)
        $stream.Write($bytes,0,$bytes.Length);$stream.Flush()
    }catch{$primary=$_}
    finally{if($null -ne $stream){try{$stream.Dispose()}catch{if($null -eq $primary){$primary=$_}}}}
    if($null -ne $primary){throw $primary}
}
try{
    if($PSVersionTable.PSVersion.ToString() -cne '7.6.5' -or $PSVersionTable.PSEdition -cne 'Core' -or -not [Environment]::Is64BitProcess){throw 'Unexpected PS7 runtime'}
    $identity=[Security.Principal.WindowsIdentity]::GetCurrent()
    $admin=([Security.Principal.WindowsPrincipal]::new($identity)).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
    if($identity.Name -cne 'DESKTOP-H4SVCLU\Lenovo' -or $admin){throw 'Unexpected fixed offline context'}
    Import-Module (Join-Path $PSScriptRoot session.psm1) -ErrorAction Stop
    . (Join-Path $PSScriptRoot core.tests.ps1)
    . (Join-Path $PSScriptRoot owned.tests.ps1)
    . (Join-Path $PSScriptRoot sequence.tests.ps1)
    switch -CaseSensitive ($Grant){
        'core-red' {Invoke-LSCoreTests}
        'core-green' {Invoke-LSCoreTests}
        'owned-red' {Invoke-LSOwnedTests}
        'owned-green' {Invoke-LSOwnedTests}
        'sequence-green' {Invoke-LSSequenceTests}
    }
    [Console]::Out.Flush()
    $trace=[ordered]@{Groups=$script:Groups.ToArray()}
    Capture-Once (Join-Path $packet ($Grant+'.trace.json')) $trace
    Capture-Once (Join-Path $packet ($Grant+'.details.json')) (Capture-Value $script:Details.ToArray())
    $failed=@($script:Groups | Where-Object {-not $_.Passed}).Count
    if($failed -gt 0){exit 1}
    exit 0
}catch{
    # One independent best-effort error output; no alternate file, no retry.
    try{[Console]::Error.WriteLine($_.ToString()+"`n"+$_.ScriptStackTrace);[Console]::Error.Flush()}catch{}
    exit 2
}
