$ErrorActionPreference='Stop'
# Real file-level reparse fixtures only. No production edits, mocks or cleanup.
$identity=[Security.Principal.WindowsIdentity]::GetCurrent()
if(-not ([Security.Principal.WindowsPrincipal]::new($identity)).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)){throw 'Approved administrator fixture context required'}
. (Join-Path $PSScriptRoot 'synthetic-capture-store.ps1')
function Check-Reparse([bool]$Value,[string]$Name){if(-not $Value){throw ('ASSERTION FAILED: '+$Name)};Write-Output ('PASS: '+$Name)}
Check-Reparse ($null -ne (Get-Command Write-OfflineSyntheticCaptureStore -CommandType Function) -and $null -ne (Get-Command Read-OfflineSyntheticCaptureStore -CommandType Function)) 'FileReparseRequired'
Write-Output ('HOST: '+$PSVersionTable.PSVersion.ToString()+' '+$PSVersionTable.PSEdition+' x64='+[Environment]::Is64BitProcess)
function Assert-ReparseResult($Value,[bool]$Verified,[string]$Reason){
    $names=@('Schema','Verified','Reason','Record','OriginVerified','Acceptance','RuntimeApproved')
    if(@($Value.PSObject.Properties).Count -ne 7 -or @($Value.PSObject.Properties|Where-Object {$_.Name -cnotin $names -or $_.MemberType -ne 'NoteProperty'}).Count -ne 0 -or $Value.Schema -ne 1 -or $Value.Verified -isnot [bool] -or $Value.Verified -ne $Verified -or $Value.Reason -cne $Reason){throw 'Unexpected file-reparse result'}
    foreach($flag in @('OriginVerified','Acceptance','RuntimeApproved')){if($Value.$flag -isnot [bool] -or $Value.$flag){throw 'File-reparse result granted authority'}}
    if((-not $Verified -and $null -ne $Value.Record) -or ($Verified -and $null -eq $Value.Record)){throw 'File-reparse record publication mismatch'}
}
function Test-ReparseBytes([byte[]]$A,[byte[]]$B){return [string]::Equals([Convert]::ToBase64String($A),[Convert]::ToBase64String($B),[StringComparison]::Ordinal)}
function New-ReparseCopy([string]$Path,[byte[]]$Bytes){$stream=[IO.File]::Open($Path,[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::None);try{$stream.Write($Bytes,0,$Bytes.Length);$stream.Flush()}finally{$stream.Dispose()}}
$root=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../.superpowers/diagnostics/offline-synthetic-capture-store'))
$null=Get-IsolatedFullPath $root
$prefix='reparse-'+[Guid]::NewGuid().ToString('N');$source=Join-Path $root ($prefix+'-source')
$packet=[pscustomobject]@{Schema=1;SourceMode='synthetic';SessionLabel='s';ExecutableDeclaration='x';ArgumentBytes=[byte[]]@();EncodingName='utf-8';OriginDeclaration='o';StartedUTC='2026-10-17T00:00:00.0000000Z';EndedUTC='2026-10-17T00:00:00.0000000Z';Events=[object[]]@()}
$written=Write-OfflineSyntheticCaptureStore $source (New-SyntheticCaptureRecord $packet)
Assert-ReparseResult $written $true 'consistent-synthetic-archive-only'
$control=Read-OfflineSyntheticCaptureStore $source;Assert-ReparseResult $control $true 'consistent-synthetic-archive-only'
$original=@{}
foreach($leaf in @('record.bin','complete.bin')){
    $path=Join-Path $source $leaf;$item=Get-Item -LiteralPath $path -Force
    if($item.PSIsContainer -or ($item.Attributes -band [IO.FileAttributes]::ReparsePoint) -ne 0){throw 'Normal source must be an ordinary file'}
    $original[$leaf]=[IO.File]::ReadAllBytes($path)
}
Check-Reparse ($control.Record.SessionLabel -ceq 's' -and @($control.Record.PSObject.Properties).Count -eq 20) 'FileReparseNormalControl'
foreach($leaf in @('record.bin','complete.bin')){
    $d=Join-Path $root ($prefix+'-'+$leaf.Replace('.','-'))
    $null=Get-IsolatedFullPath $d
    if(Test-Path -LiteralPath $d){throw 'Fresh file-reparse fixture collision'}
    New-Item -ItemType Directory -Path $d -ErrorAction Stop|Out-Null
    $other=if($leaf -ceq 'record.bin'){'complete.bin'}else{'record.bin'}
    New-ReparseCopy (Join-Path $d $other) $original[$other]
    $link=Join-Path $d $leaf;$target=Join-Path $source $leaf
    New-Item -ItemType SymbolicLink -Path $link -Target $target -ErrorAction Stop|Out-Null
    $item=Get-Item -LiteralPath $link -Force;$targets=@($item.Target)
    $linkedBytes=[IO.File]::ReadAllBytes($link)
    Check-Reparse (-not $item.PSIsContainer -and ($item.Attributes -band [IO.FileAttributes]::ReparsePoint) -ne 0 -and $item.LinkType -ceq 'SymbolicLink' -and $targets.Count -eq 1 -and [string]::Equals([IO.Path]::GetFullPath($targets[0]),$target,[StringComparison]::OrdinalIgnoreCase) -and (Test-ReparseBytes $linkedBytes $original[$leaf]) -and (Test-ReparseBytes ([IO.File]::ReadAllBytes((Join-Path $d $other))) $original[$other])) ('FileReparseFixture-'+$leaf)
    # Removing reparse-path protection would expose an otherwise valid archive.
    $refused=Read-OfflineSyntheticCaptureStore $d
    Assert-ReparseResult $refused $false 'io-error'
    $sourceIntact=$true
    foreach($name in @('record.bin','complete.bin')){if(-not (Test-ReparseBytes ([IO.File]::ReadAllBytes((Join-Path $source $name))) $original[$name])){$sourceIntact=$false}}
    $again=Read-OfflineSyntheticCaptureStore $source;Assert-ReparseResult $again $true 'consistent-synthetic-archive-only'
    Check-Reparse ($sourceIntact -and -not (Test-Path -LiteralPath (Join-Path $d 'stopped')) -and -not (Test-Path -LiteralPath (Join-Path $source 'stopped'))) ('FileReparseReaderRefused-'+$leaf)
}
Check-Reparse $true 'FileReparseComplete'
Write-Output 'LIMIT: real synthetic file-reparse refusal only;admin context,no cleanup or acceptance.'
