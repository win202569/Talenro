# Supported offline entry. Catalog data describes dependencies, never authority.
[CmdletBinding(PositionalBinding=$false)]
param([string]$Action='Inspect',[string]$Profile='')
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
if($Action -cnotin @('Inspect','RunApproved')){throw 'OD: unsupported action'}
if(($Profile -cnotin @('entry-migration-core','entry-migration-owned','entry-migration-live')) -and -not ($Action -ceq 'Inspect' -and $Profile -ceq '')){throw 'OD: unsupported profile'}
$repo=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))
$private=Join-Path $PSScriptRoot 'private/offline-diagnostics'
if($Action -ceq 'RunApproved'){
    # Deliberately independent of catalog content.
    switch -CaseSensitive ($Profile){
        'entry-migration-core' {$selected='entry-migration-core'}
        'entry-migration-owned' {$selected='entry-migration-owned'}
        'entry-migration-live' {$selected='entry-migration-live'}
        default {throw 'OD: unsupported profile'}
    }
    $kernel=Join-Path $private session.psm1
    if(-not [IO.File]::Exists($kernel)){throw 'OD: backend unavailable'}
    Import-Module -Name $kernel -ErrorAction Stop
    Invoke-ODProfile -Profile $selected
    return
}
$catalog=Get-Content -LiteralPath (Join-Path $private catalog.json) -Raw | ConvertFrom-Json
$dependencies=[Collections.Generic.List[object]]::new()
foreach($item in $catalog.Dependencies){
    $path=if([IO.Path]::IsPathRooted($item.Path)){[IO.Path]::GetFullPath($item.Path)}else{[IO.Path]::GetFullPath((Join-Path $repo $item.Path))}
    $exists=[IO.File]::Exists($path)
    $hash=if($exists){(Get-FileHash -LiteralPath $path -Algorithm SHA256).Hash}else{$null}
    $dependencies.Add([pscustomobject][ordered]@{Role=$item.Role;Path=$path;Exists=$exists;SHA256=$hash;MatchesPinned=($exists -and $hash -ceq $item.SHA256)})
}
$freeze=[IO.Path]::GetFullPath((Join-Path $repo '.superpowers/sdd/2026-10-31-offline-diagnostic-runtime-migration/runtime-green.frozen.json'))
$freezeExists=[IO.File]::Exists($freeze)
$freezeHash=if($freezeExists){(Get-FileHash -LiteralPath $freeze -Algorithm SHA256).Hash}else{$null}
$dependencies.Add([pscustomobject][ordered]@{Role='FreezeManifest';Path=$freeze;Exists=$freezeExists;SHA256=$freezeHash;MatchesPinned=$null})
$manifest=if($freezeExists){Get-Content -LiteralPath $freeze -Raw | ConvertFrom-Json}else{$null}
foreach($name in @('session.psm1','sinks.ps1','owned-step.ps1','sequence.ps1','profiles.ps1','observations.ps1')){
    $path=[IO.Path]::GetFullPath((Join-Path $private $name))
    $exists=[IO.File]::Exists($path)
    $hash=if($exists){(Get-FileHash -LiteralPath $path -Algorithm SHA256).Hash}else{$null}
    $matches=$false
    if($exists -and $null -ne $manifest){
        $pins=@($manifest.Pins | Where-Object {$_.VerificationPath -ieq $path})
        $matches=$pins.Count -gt 0
        foreach($pin in $pins){if($pin.ExpectedState -cne 'Present' -or $pin.SHA256 -cne $hash){$matches=$false}}
    }
    $dependencies.Add([pscustomobject][ordered]@{Role=('Impl:'+$name);Path=$path;Exists=$exists;SHA256=$hash;MatchesPinned=$matches})
}
[pscustomobject][ordered]@{
    Schema=1;Action='Inspect';Profiles=@($catalog.Profiles);Dependencies=$dependencies.ToArray()
    History=[pscustomobject][ordered]@{Path=[IO.Path]::GetFullPath((Join-Path $repo $catalog.History.Path));State=$catalog.History.State;Role=$catalog.History.Role}
    Authority='None';SourceVerified=$false;LiveActionsPerformed=$false;RuntimeApproved=$false;FormalAcceptance=$false;RealGateImplemented=$false
}
