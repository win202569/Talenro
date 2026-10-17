$ErrorActionPreference='Stop'
function Check-Store([bool]$Value,[string]$Name){if(-not $Value){throw ('ASSERTION FAILED: '+$Name)};Write-Output ('PASS: '+$Name)}
$helper=Join-Path $PSScriptRoot 'output-evidence-store.ps1'
if(Test-Path -LiteralPath $helper){. $helper}
Check-Store ($null -ne (Get-Command Write-OfflineOutputEvidenceStore -CommandType Function -ErrorAction SilentlyContinue)) 'OfflineEvidenceStoreRequired'
$root=Join-Path ([IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..'))) '.superpowers/diagnostics/offline-output-evidence-store'
$null=Get-IsolatedFullPath $root
if(-not (Test-Path -LiteralPath $root)){New-Item -ItemType Directory -Path $root -ErrorAction Stop|Out-Null}
$script:storePrefix='fixture-'+[Guid]::NewGuid().ToString('N');$script:storeCounter=0
function New-StorePath{$script:storeCounter++;return (Join-Path $root ($script:storePrefix+'-'+$script:storeCounter))}
function New-StorePackage{New-OfflineOutputEvidencePackage -StdoutBytes ([byte[]]@(97,98,99)) -StderrBytes ([byte[]]@(0,255,13,10)) -ExitCode ([int]-1) -EncodingName 'utf-8' -OriginDeclaration 'synthetic-only'}
function Put-StoreFixture([string]$Path,[byte[]]$Payload,[byte[]]$Marker){
    New-Item -ItemType Directory -Path $Path -ErrorAction Stop|Out-Null
    [IO.File]::WriteAllBytes((Join-Path $Path 'evidence.bin'),$Payload)
    [IO.File]::WriteAllBytes((Join-Path $Path 'complete.bin'),$Marker)
}
function Make-StoreMarker([byte[]]$Payload){
    $ms=New-Object IO.MemoryStream;$bw=New-Object IO.BinaryWriter($ms)
    $sha=[Security.Cryptography.SHA256]::Create()
    try{$bw.Write([Text.Encoding]::ASCII.GetBytes('TLROEC01'));$bw.Write([long]$Payload.Length);$bw.Write($sha.ComputeHash($Payload));$bw.Flush();return ,$ms.ToArray()}finally{$sha.Dispose();$bw.Dispose();$ms.Dispose()}
}
function Check-StoreRefusal([string]$Path,[string]$Name){
    $r=Read-OfflineOutputEvidenceStore $Path
    Check-Store (-not $r.Verified -and $null -eq $r.Package -and -not $r.OriginVerified -and -not $r.Acceptance -and -not $r.RuntimeApproved) $Name
}
$path=New-StorePath;$p=New-StorePackage;$r=Write-OfflineOutputEvidenceStore -Directory $path -Package $p
Check-Store ($r.Verified -and $r.Package.Stdout.Base64 -ceq 'YWJj' -and $r.Package.Stderr.Base64 -ceq 'AP8NCg==' -and $r.Package.ExitCode -eq -1 -and $r.Package.OriginDeclaration -ceq 'synthetic-only' -and $r.Package.Classification -ceq 'refused') 'LosslessFileRoundtripIncludingRefusedOutput'
Check-Store (-not $r.OriginVerified -and -not $r.Acceptance -and -not $r.RuntimeApproved -and -not $r.Package.OriginVerified) 'StoreNeverAuthenticatesOrApproves'
$bytes=[IO.File]::ReadAllBytes((Join-Path $path 'evidence.bin'));$marker=[IO.File]::ReadAllBytes((Join-Path $path 'complete.bin'))
# Independently hand-encoded layout: magic,exit=-1,encoding=1,origin length14,origin,stdout length3/abc,stderr length4/raw bytes.
$expected='544C524F45563031FFFFFFFF010000000E00000073796E7468657469632D6F6E6C79030000006162630400000000FF0D0A'
Check-Store (([BitConverter]::ToString($bytes).Replace('-','')) -ceq $expected) 'PrimitiveLayoutMatchesIndependentLiteral'
Check-Store ($marker.Length -eq 48 -and [Text.Encoding]::ASCII.GetString($marker,0,8) -ceq 'TLROEC01' -and [BitConverter]::ToInt64($marker,8) -eq $bytes.Length) 'CompletionMarkerExactLayout'
$p.Stdout.Base64='changed';Check-Store ((Read-OfflineOutputEvidenceStore $path).Package.Stdout.Base64 -ceq 'YWJj') 'StoredBytesIndependentOfCallerObject'
$rejected=$false;try{Write-OfflineOutputEvidenceStore $path (New-StorePackage)|Out-Null}catch{$rejected=$true}
Check-Store ($rejected -and [Convert]::ToBase64String([IO.File]::ReadAllBytes((Join-Path $path 'evidence.bin'))) -ceq [Convert]::ToBase64String($bytes) -and -not (Test-Path -LiteralPath (Join-Path $path 'stopped'))) 'ExistingDestinationNotOverwrittenOrMarked'
foreach($enc in @('utf-8','utf-16le')){
    $q=New-OfflineOutputEvidencePackage -StdoutBytes ([byte[]]@()) -StderrBytes ([byte[]]@()) -ExitCode 0 -EncodingName $enc -OriginDeclaration 'empty'
    $r=Write-OfflineOutputEvidenceStore (New-StorePath) $q
    Check-Store ($r.Verified -and $r.Package.EncodingName -ceq $enc -and $r.Package.Classification -ceq 'empty' -and $r.Package.Stdout.ByteLength -eq 0) ('EmptyEncodingRoundtrip-'+$enc)
}
$xml='#< CLIXML'+"`n"+'<Objs Version="1.1.0.1" xmlns="http://schemas.microsoft.com/powershell/2004/04"><Obj S="progress" RefId="0"><TN RefId="0"><T>System.Management.Automation.PSCustomObject</T><T>System.Object</T></TN><MS><I64 N="SourceId">1</I64><PR N="Record"><AV>fixture</AV><AI>0</AI><Nil /><PI>-1</PI><PC>-1</PC><T>Completed</T><SR>-1</SR><SD> </SD></PR></MS></Obj></Objs>'
$q=New-OfflineOutputEvidencePackage -StdoutBytes ([byte[]]@()) -StderrBytes ([Text.Encoding]::UTF8.GetBytes($xml)) -ExitCode 0 -EncodingName 'utf-8' -OriginDeclaration 'progress'
$r=Write-OfflineOutputEvidenceStore (New-StorePath) $q
Check-Store ($r.Verified -and $r.Package.Classification -ceq 'progress-only' -and $r.Package.ProgressCount -eq 1) 'ClassificationRebuiltFromFileBytes'
# Catch byte loss/transcoding or ignoring the explicit UTF16 encoding on reread.
$utf16Stderr=[Text.Encoding]::Unicode.GetBytes($xml)
$q=New-OfflineOutputEvidencePackage -StdoutBytes ([byte[]]@(65,0,45,78)) -StderrBytes $utf16Stderr -ExitCode 0 -EncodingName 'utf-16le' -OriginDeclaration 'nonempty-utf16'
$r=Write-OfflineOutputEvidenceStore (New-StorePath) $q
Check-Store ($r.Verified -and $r.Package.EncodingName -ceq 'utf-16le' -and $r.Package.Stdout.Base64 -ceq 'QQAtTg==' -and $r.Package.Stdout.ByteLength -eq 4 -and $r.Package.Stderr.Base64 -ceq [Convert]::ToBase64String($utf16Stderr) -and $r.Package.Stderr.ByteLength -eq ($xml.Length*2)) 'NonemptyUtf16ExactBytesSurvivePersistence'
Check-Store ($r.Verified -and $r.Package.Classification -ceq 'progress-only' -and $r.Package.ProgressCount -eq 1) 'NonemptyUtf16ProgressClassificationRebuilt'
foreach($kind in @('missing-payload','missing-marker','stopped','truncated-marker','marker-tail','marker-magic','marker-length','marker-hash','payload-hash')){
    $d=New-StorePath;New-Item -ItemType Directory -Path $d -ErrorAction Stop|Out-Null
    $b=[byte[]]$bytes.Clone();$m=[byte[]]$marker.Clone()
    switch($kind){
        stopped{[IO.File]::WriteAllBytes((Join-Path $d 'stopped'),[byte[]]@(1))}
        truncated-marker{$m=$m[0..46]} marker-tail{$m=[byte[]]($m+0)}
        marker-magic{$m[0]=0} marker-length{$m[8]=0} marker-hash{$m[16]=$m[16] -bxor 1}
        payload-hash{$b[0]=0}
    }
    if($kind -ne 'missing-payload'){[IO.File]::WriteAllBytes((Join-Path $d 'evidence.bin'),$b)}
    if($kind -ne 'missing-marker'){[IO.File]::WriteAllBytes((Join-Path $d 'complete.bin'),$m)}
    Check-StoreRefusal $d ('IncompleteOrMismatchRefused-'+$kind)
}
# Correctly recomputed hashes must not let malformed primitive fields through.
# Paired control: identical stream fields with one valid origin character.
$originControl=[byte[]]($bytes[0..15]+[byte[]]@(1,0,0,0,120)+$bytes[34..48])
$controlPath=New-StorePath;Put-StoreFixture $controlPath $originControl (Make-StoreMarker $originControl)
$control=Read-OfflineOutputEvidenceStore $controlPath
Check-Store ($control.Verified -and $control.Package.OriginDeclaration -ceq 'x' -and $control.Package.Stdout.Base64 -ceq 'YWJj' -and $control.Package.Stderr.Base64 -ceq 'AP8NCg==' -and $control.Package.ExitCode -eq -1) 'SingleCharacterOriginControlOtherwiseValid'
foreach($kind in @('magic','encoding','negative-origin','huge-origin','empty-origin','invalid-utf8','origin-char-bound','negative-stdout','huge-stdout','huge-stderr','truncated','tail','oversized-payload')){
    $b=[byte[]]$bytes.Clone()
    switch($kind){
        magic{$b[0]=0} encoding{$b[12]=3}
        negative-origin{[Array]::Copy([BitConverter]::GetBytes([int]-1),0,$b,16,4)}
        huge-origin{[Array]::Copy([BitConverter]::GetBytes([int]4097),0,$b,16,4)}
        empty-origin{$b=[byte[]]($bytes[0..15]+[byte[]]@(0,0,0,0)+$bytes[34..48])} invalid-utf8{$b[20]=255}
        origin-char-bound{$b=[byte[]]($b[0..15]+[BitConverter]::GetBytes([int]1025)+([Text.Encoding]::UTF8.GetBytes(('x'*1025)))+$b[34..48])}
        negative-stdout{[Array]::Copy([BitConverter]::GetBytes([int]-1),0,$b,34,4)}
        huge-stdout{[Array]::Copy([BitConverter]::GetBytes([int]1048577),0,$b,34,4)}
        huge-stderr{[Array]::Copy([BitConverter]::GetBytes([int]1048577),0,$b,41,4)}
        truncated{$b=$b[0..47]} tail{$b=[byte[]]($b+0)}
        oversized-payload{$b=New-Object byte[] 3145729}
    }
    $d=New-StorePath;Put-StoreFixture $d $b (Make-StoreMarker $b)
    Check-StoreRefusal $d ('ValidHashMalformedPayloadRefused-'+$kind)
    if($kind -eq 'empty-origin'){
        $emptyOriginResult=Read-OfflineOutputEvidenceStore $d
        Check-Store (-not $emptyOriginResult.Verified -and $emptyOriginResult.Reason -ceq 'Bounded origin declaration required') 'EmptyOriginRefusedByOriginGuardNotMalformedFollowingField'
    }
}
$q=New-OfflineOutputEvidencePackage -StdoutBytes (New-Object byte[] 1048576) -StderrBytes (New-Object byte[] 1048576) -ExitCode 0 -EncodingName 'utf-8' -OriginDeclaration ('x'*1024)
$r=Write-OfflineOutputEvidenceStore (New-StorePath) $q
Check-Store ($r.Verified -and $r.Package.Stdout.ByteLength -eq 1048576 -and $r.Package.Stderr.ByteLength -eq 1048576 -and $r.Package.OriginDeclaration.Length -eq 1024) 'BothStreamsAndOriginExactBound'
$bad=New-StorePackage;$bad.Acceptance=$true;$d=New-StorePath;$rejected=$false
try{Write-OfflineOutputEvidenceStore $d $bad|Out-Null}catch{$rejected=$true}
Check-Store ($rejected -and -not (Test-Path -LiteralPath $d)) 'BadPackageRejectedBeforeDirectoryCreation'
$bad=New-StorePackage;$bad.OriginDeclaration=[string][char]0xD800;$d=New-StorePath;$rejected=$false
try{Write-OfflineOutputEvidenceStore $d $bad|Out-Null}catch{$rejected=$true}
Check-Store ($rejected -and -not (Test-Path -LiteralPath $d)) 'InvalidOriginUtf16CannotSilentlyReplaceBytes'
foreach($unsafe in @((Join-Path $root '../escape'),'\\localhost\share\fixture',($path+':stream'),(Join-Path $path 'nested'),($path+'.'))){
    Check-StoreRefusal $unsafe 'UnsafeReadPathRefused'
    $rejected=$false;try{Write-OfflineOutputEvidenceStore $unsafe (New-StorePackage)|Out-Null}catch{$rejected=$true}
    Check-Store $rejected 'UnsafeWritePathRefused'
}
foreach($leaf in @('evidence.bin','complete.bin')){
    $d=New-StorePath;Put-StoreFixture $d $bytes $marker
    $link=Join-Path $root ($script:storePrefix+'-hardlink-'+$leaf)
    New-Item -ItemType HardLink -Path $link -Target (Join-Path $d $leaf) -ErrorAction Stop|Out-Null
    Check-StoreRefusal $d ('MultipleLinksRefused-'+$leaf)
}
$junction=New-StorePath
New-Item -ItemType Junction -Path $junction -Target $path -ErrorAction Stop|Out-Null
Check-StoreRefusal $junction 'ReparseDirectoryRefused'
$rejected=$false;try{Write-OfflineOutputEvidenceStore $junction (New-StorePackage)|Out-Null}catch{$rejected=$true}
Check-Store $rejected 'ReparseWriteRefused'
$d=New-StorePath;Put-StoreFixture $d $bytes $marker
$held=[IO.File]::Open((Join-Path $d 'evidence.bin'),[IO.FileMode]::Open,[IO.FileAccess]::Write,[IO.FileShare]::None)
try{Check-StoreRefusal $d 'ExclusiveWriterPreventsReadVerification'}finally{$held.Dispose()}
Write-Output 'LIMIT: synthetic local fixtures retained;no live collection,origin trust,power-loss guarantee or acceptance.'
