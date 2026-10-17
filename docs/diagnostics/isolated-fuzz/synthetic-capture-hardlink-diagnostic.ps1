# Exploratory one-attempt fixture diagnostic,not an archive acceptance test.
# Separate finite authorization required. No elevation,alternate link API,
# ACL/ownership/security change,cleanup,product import or original suite replay.
$ErrorActionPreference='Stop'
[Console]::OutputEncoding=New-Object Text.UTF8Encoding($false)
[Console]::InputEncoding=New-Object Text.UTF8Encoding($false)
function Write-LinkRow($Row){[Console]::Out.WriteLine('LINK '+(ConvertTo-Json -InputObject $Row -Depth 8 -Compress));[Console]::Out.Flush()}
$projectRoot=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..'))
$fixtureRoot=Join-Path $projectRoot '.superpowers/diagnostics/offline-synthetic-capture-store'
if($PSVersionTable.PSVersion.Major -ne 7){throw 'PS7 required'}
if(-not (Test-Path -LiteralPath $fixtureRoot -PathType Container) -or ((Get-Item -LiteralPath $fixtureRoot -Force).Attributes -band [IO.FileAttributes]::ReparsePoint) -ne 0){throw 'Ordinary existing fixture root required'}
$directory=[IO.Path]::GetFullPath((Join-Path $fixtureRoot ('hardlink-probe-'+[Guid]::NewGuid().ToString('N'))))
if(-not $directory.StartsWith(($fixtureRoot+[IO.Path]::DirectorySeparatorChar),[StringComparison]::OrdinalIgnoreCase) -or (Test-Path -LiteralPath $directory)){throw 'New scoped directory required'}
$source=Join-Path $directory 'source.bin';$link=Join-Path $directory 'link.bin'
$identity=[Security.Principal.WindowsIdentity]::GetCurrent();$principal=[Security.Principal.WindowsPrincipal]::new($identity)
[Console]::Out.WriteLine('HOST: '+$PSVersionTable.PSVersion.ToString()+' '+$PSVersionTable.PSEdition+' x64='+[Environment]::Is64BitProcess);[Console]::Out.Flush()
Write-LinkRow ([ordered]@{Event='setup-start';Directory=$directory;AdministratorRole=$principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)})
$null=New-Item -ItemType Directory -Path $directory -ErrorAction Stop
$stream=[IO.File]::Open($source,[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::None)
try{$payload=[byte[]]@(84,76,82,49);$stream.Write($payload,0,$payload.Length);$stream.Flush()}finally{$stream.Dispose()}
if((Get-Item -LiteralPath $source).Length -ne 4 -or (Test-Path -LiteralPath $link)){throw 'Unexpected pre-attempt state'}
Write-LinkRow ([ordered]@{Event='setup-end';Source=$source;SourceBytes=4;DestinationExists=$false})
Write-LinkRow ([ordered]@{Event='attempt-start';Attempts=1;Operation='New-Item HardLink';Destination=$link})
$outcome='created';$failureInfo=$null;$linkType=$null
try{
    $created=New-Item -ItemType HardLink -Path $link -Target $source -ErrorAction Stop
    $linkType=[string]$created.LinkType
}catch{
    $failure=$_;$chain=New-Object 'System.Collections.Generic.List[object]';$exception=$failure.Exception;$permissionDenied=$false
    for($depth=0;$null -ne $exception -and $depth -lt 8;$depth++){
        $nativeCode=$null
        if($exception -is [ComponentModel.Win32Exception]){$nativeCode=$exception.NativeErrorCode;if($nativeCode -eq 5){$permissionDenied=$true}}
        $hr=[int]$exception.HResult;$hex=[BitConverter]::ToUInt32([BitConverter]::GetBytes($hr),0).ToString('X8')
        if($exception -is [UnauthorizedAccessException] -or $exception -is [Security.SecurityException] -or $hex -ceq '80070005'){$permissionDenied=$true}
        $chain.Add([pscustomobject]@{ExceptionType=$exception.GetType().FullName;HResult=$hr;HResultHex=('0x'+$hex);NativeErrorCode=$nativeCode;Message=$exception.Message})
        $exception=$exception.InnerException
    }
    $category=[string]$failure.CategoryInfo.Category
    if($category -cin @('PermissionDenied','SecurityError')){$permissionDenied=$true}
    $outcome=if($permissionDenied){'refused'}else{'unexpected-error'}
    $failureInfo=[ordered]@{FullyQualifiedErrorId=$failure.FullyQualifiedErrorId;Category=$category;Exceptions=[object[]]$chain.ToArray();ChainTruncated=($null -ne $exception)}
}
$sourceBytes=[IO.File]::ReadAllBytes($source)
if($sourceBytes.Length -ne 4 -or [Convert]::ToBase64String($sourceBytes) -cne 'VExSMQ=='){throw 'Ordinary source changed'}
$destinationExists=Test-Path -LiteralPath $link
if($outcome -ceq 'created'){
    if(-not $destinationExists -or $linkType -cne 'HardLink' -or [Convert]::ToBase64String([IO.File]::ReadAllBytes($link)) -cne 'VExSMQ=='){throw 'Unexpected created link state'}
}elseif($destinationExists){throw 'Unexpected destination after failure'}
Write-LinkRow ([ordered]@{Event='result';Attempts=1;Outcome=$outcome;SourceBytes=4;DestinationExists=$destinationExists;ReportedLinkType=$linkType;Error=$failureInfo})
if($outcome -ceq 'unexpected-error'){throw 'Unrelated hardlink error;stop without retry'}
[Console]::Out.WriteLine('DIAGNOSTIC COMPLETE: one hardlink attempt only;not archive acceptance.');[Console]::Out.Flush()
