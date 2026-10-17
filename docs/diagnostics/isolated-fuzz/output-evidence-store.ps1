# Offline local file persistence only. No process collector or runtime helper.
. (Join-Path $PSScriptRoot 'output-evidence-package.ps1')
. (Join-Path $PSScriptRoot 'isolation.ps1')
$script:OfflineEvidenceStoreRoot=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../.superpowers/diagnostics/offline-output-evidence-store'))

function Assert-OfflineStoreDirectory([string]$Directory){
    if([string]::IsNullOrWhiteSpace($Directory) -or $Directory -notmatch '^[A-Za-z]:[\\/]'){throw 'Absolute local evidence directory required'}
    $full=Get-IsolatedFullPath $Directory
    if(-not $full.Equals($Directory.Replace('/','\'),[StringComparison]::OrdinalIgnoreCase)){throw 'Noncanonical evidence path rejected'}
    $root=Get-IsolatedFullPath $script:OfflineEvidenceStoreRoot
    if(-not (Test-Path -LiteralPath $root -PathType Container)){throw 'Dedicated evidence root must already exist'}
    if(-not ([IO.Path]::GetDirectoryName($full)).Equals($root,[StringComparison]::OrdinalIgnoreCase) -or [IO.Path]::GetFileName($full) -cnotmatch '^[A-Za-z0-9][A-Za-z0-9_-]{0,79}\z'){throw 'Evidence destination outside direct-child scope'}
    return $full
}
function Open-OfflineStoreInput([string]$Path,[long]$Maximum){
    $null=Get-IsolatedFullPath $Path
    $stream=[IO.File]::Open($Path,[IO.FileMode]::Open,[IO.FileAccess]::Read,[IO.FileShare]::Read)
    try{
        $info=[IsolatedFuzzFileInfo]::Snapshot($stream.SafeFileHandle)
        if($info.Links -ne 1 -or ($info.Attributes -band [uint32][IO.FileAttributes]::ReparsePoint) -ne 0){throw 'Linked evidence input rejected'}
        if($stream.Length -gt $Maximum){throw 'Evidence file exceeds bound'}
        return $stream
    }catch{$stream.Dispose();throw}
}
function Read-OfflineStoreBytes([IO.BinaryReader]$Reader,[int]$Count){
    if($Count -lt 0 -or $Count -gt ($Reader.BaseStream.Length-$Reader.BaseStream.Position)){throw 'Truncated evidence field'}
    $bytes=$Reader.ReadBytes($Count)
    if($bytes.Length -ne $Count){throw 'Truncated evidence bytes'}
    return ,$bytes
}
function Read-OfflineStoreField([IO.BinaryReader]$Reader,[int]$Maximum){
    $count=$Reader.ReadInt32()
    if($count -lt 0 -or $count -gt $Maximum){throw 'Evidence field exceeds bound'}
    return ,(Read-OfflineStoreBytes $Reader $count)
}
function Get-OfflineStoreHash([byte[]]$Bytes){
    $sha=[Security.Cryptography.SHA256]::Create()
    try{return ,$sha.ComputeHash($Bytes)}finally{$sha.Dispose()}
}
function Read-OfflineOutputEvidenceStore([string]$Directory){
    $result=[pscustomobject]@{Verified=$false;Reason='unverified';Package=$null;OriginVerified=$false;Acceptance=$false;RuntimeApproved=$false}
    $payload=$null;$marker=$null;$pr=$null;$mr=$null;$memory=$null;$reader=$null
    try{
        $dir=Assert-OfflineStoreDirectory $Directory
        if(-not (Test-Path -LiteralPath $dir -PathType Container)){throw 'Evidence directory missing'}
        if(Test-Path -LiteralPath (Join-Path $dir 'stopped')){throw 'Stopped evidence cannot complete'}
        # Both bindings deny writes/deletes for the duration of verification.
        $marker=Open-OfflineStoreInput (Join-Path $dir 'complete.bin') 48
        $payload=Open-OfflineStoreInput (Join-Path $dir 'evidence.bin') 3145728
        if($marker.Length -ne 48){throw 'Completion marker must be exactly 48 bytes'}
        $mr=New-Object IO.BinaryReader($marker)
        if([Text.Encoding]::ASCII.GetString((Read-OfflineStoreBytes $mr 8)) -cne 'TLROEC01'){throw 'Unknown completion marker'}
        $length=$mr.ReadInt64();$hash=Read-OfflineStoreBytes $mr 32
        if($length -ne $payload.Length){throw 'Completion length mismatch'}
        $pr=New-Object IO.BinaryReader($payload)
        $bytes=Read-OfflineStoreBytes $pr ([int]$payload.Length)
        if([Convert]::ToBase64String((Get-OfflineStoreHash $bytes)) -cne [Convert]::ToBase64String($hash)){throw 'Completion hash mismatch'}
        $memory=New-Object IO.MemoryStream(,$bytes);$reader=New-Object IO.BinaryReader($memory)
        if([Text.Encoding]::ASCII.GetString((Read-OfflineStoreBytes $reader 8)) -cne 'TLROEV01'){throw 'Unknown evidence payload'}
        $exitCode=$reader.ReadInt32();$encoding=$reader.ReadInt32()
        if($encoding -notin @(1,2)){throw 'Unknown evidence encoding'}
        $utf8=New-Object Text.UTF8Encoding($false,$true)
        $origin=$utf8.GetString((Read-OfflineStoreField $reader 4096))
        $stdout=Read-OfflineStoreField $reader 1048576
        $stderr=Read-OfflineStoreField $reader 1048576
        if($memory.Position -ne $memory.Length){throw 'Trailing payload bytes rejected'}
        $name=if($encoding -eq 1){'utf-8'}else{'utf-16le'}
        $package=New-OfflineOutputEvidencePackage -StdoutBytes $stdout -StderrBytes $stderr -ExitCode $exitCode -EncodingName $name -OriginDeclaration $origin
        $check=Test-OfflineOutputEvidencePackage $package
        if(-not $check.Verified){throw 'Reconstructed package did not verify'}
        if(Test-Path -LiteralPath (Join-Path $dir 'stopped')){throw 'Stopped evidence cannot complete'}
        $result.Package=$package;$result.Verified=$true;$result.Reason='integrity-verified-only'
    }catch{$result.Reason=$_.Exception.Message}
    finally{foreach($resource in @($reader,$memory,$pr,$mr,$payload,$marker)){if($null -ne $resource){$resource.Dispose()}}}
    return $result
}
function Write-OfflineStoreNewFile([string]$Path,[byte[]]$Bytes){
    $stream=[IO.File]::Open($Path,[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::None)
    try{
        $info=[IsolatedFuzzFileInfo]::Snapshot($stream.SafeFileHandle)
        if($info.Links -ne 1 -or ($info.Attributes -band [uint32][IO.FileAttributes]::ReparsePoint) -ne 0){throw 'Linked evidence output rejected'}
        $stream.Write($Bytes,0,$Bytes.Length);$stream.Flush()
    }finally{$stream.Dispose()}
}
function Write-OfflineOutputEvidenceStore([string]$Directory,$Package){
    $dir=Assert-OfflineStoreDirectory $Directory
    if(Test-Path -LiteralPath $dir){throw 'Evidence destination already exists'}
    if(-not (Test-OfflineOutputEvidencePackage $Package).Verified){throw 'Consistent evidence package required'}
    # Snapshot caller values into a primitive payload before creating any directory.
    $stdout=Get-OfflineEvidenceStreamBytes $Package.Stdout;$stderr=Get-OfflineEvidenceStreamBytes $Package.Stderr
    $utf8=New-Object Text.UTF8Encoding($false,$true);$origin=$utf8.GetBytes($Package.OriginDeclaration)
    if($origin.Length -gt 4096){throw 'Origin UTF8 exceeds bound'}
    $encoding=if($Package.EncodingName -ceq 'utf-8'){1}else{2}
    $memory=New-Object IO.MemoryStream;$writer=New-Object IO.BinaryWriter($memory)
    try{
        $writer.Write([Text.Encoding]::ASCII.GetBytes('TLROEV01'))
        $writer.Write([int]$Package.ExitCode);$writer.Write([int]$encoding)
        foreach($field in @($origin,$stdout,$stderr)){$writer.Write([int]$field.Length);$writer.Write([byte[]]$field)}
        $writer.Flush();$bytes=$memory.ToArray()
    }finally{$writer.Dispose();$memory.Dispose()}
    if($bytes.Length -gt 3145728){throw 'Payload exceeds bound'}
    $created=$false;$binding=$null;$reader=$null
    try{
        # New-Item without Force fails on an existing directory; no reuse.
        New-Item -ItemType Directory -Path $dir -ErrorAction Stop|Out-Null;$created=$true
        $null=Assert-OfflineStoreDirectory $dir
        Write-OfflineStoreNewFile (Join-Path $dir 'evidence.bin') $bytes
        $binding=Open-OfflineStoreInput (Join-Path $dir 'evidence.bin') 3145728
        $reader=New-Object IO.BinaryReader($binding)
        $stored=Read-OfflineStoreBytes $reader ([int]$binding.Length)
        if($stored.Length -ne $bytes.Length -or [Convert]::ToBase64String((Get-OfflineStoreHash $stored)) -cne [Convert]::ToBase64String((Get-OfflineStoreHash $bytes))){throw 'Written payload did not reread consistently'}
        $markerMemory=New-Object IO.MemoryStream;$markerWriter=New-Object IO.BinaryWriter($markerMemory)
        try{
            $markerWriter.Write([Text.Encoding]::ASCII.GetBytes('TLROEC01'));$markerWriter.Write([long]$stored.Length)
            $markerWriter.Write((Get-OfflineStoreHash $stored));$markerWriter.Flush();$markerBytes=$markerMemory.ToArray()
        }finally{$markerWriter.Dispose();$markerMemory.Dispose()}
        # Completion is published only after the closed payload rereads correctly.
        Write-OfflineStoreNewFile (Join-Path $dir 'complete.bin') $markerBytes
        $reader.Dispose();$reader=$null;$binding.Dispose();$binding=$null
        $result=Read-OfflineOutputEvidenceStore $dir
        if(-not $result.Verified){throw ('Final evidence reread failed: '+$result.Reason)}
        return $result
    }catch{
        if($created){try{Write-OfflineStoreNewFile (Join-Path $dir 'stopped') ([byte[]]@(1))}catch{}}
        throw
    }finally{if($null -ne $reader){$reader.Dispose()};if($null -ne $binding){$binding.Dispose()}}
}
