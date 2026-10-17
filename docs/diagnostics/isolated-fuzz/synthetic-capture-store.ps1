# Offline synthetic records only. No live adapter,command execution or migration.
. (Join-Path $PSScriptRoot 'capture-adapter-contract.ps1')
. (Join-Path $PSScriptRoot 'isolation.ps1')
$script:SyntheticStoreRoot=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../.superpowers/diagnostics/offline-synthetic-capture-store'))
# Trusted fixed layouts,not caller-provided schemas or wire type/field names.
$script:SyntheticStoreLayouts=@{
    record=@('Schema:i32:0','Policy:text:64','SourceMode:text:16','SessionLabel:text:80','ExecutableDeclaration:text:4096','Arguments:snapshot:65536','EncodingName:text:8','OriginDeclaration:text:4096','StartedUTC:text:28','EndedUTC:text:28','UtcOrderConsistent:bool:0','Events:events:4096','Lifecycle:lifecycle:0','StopIntent:text:64','CleanupStatus:text:32','LiveActionsPerformed:bool:0','BindingSHA256:digest:32','OriginVerified:bool:0','Acceptance:bool:0','RuntimeApproved:bool:0')
    lifecycle=@('Schema:i32:0','Policy:text:64','Status:text:16','Reason:text:64','AcceptedEventCount:i32:0','FailureEventIndex:nullable32:0','LastObservedElapsedMs:i64:0','StdoutEOF:bool:0','StderrEOF:bool:0','ExitObserved:bool:0','ExitCode:nullable32:0','Stdout:snapshot:1048576','Stderr:snapshot:1048576','Package:optionalpackage:0','OriginVerified:bool:0','Acceptance:bool:0','RuntimeApproved:bool:0')
    package=@('Schema:i32:0','Policy:text:64','EncodingName:text:8','ExitCode:i32:0','OriginDeclaration:text:4096','OriginVerified:bool:0','Stdout:snapshot:1048576','Stderr:snapshot:1048576','Classification:text:16','ProgressCount:i32:0','Acceptance:bool:0','RuntimeApproved:bool:0')
}
function Get-SyntheticStoreType($Value){if($null -eq $Value){return $null};return [object].GetMethod('GetType').Invoke($Value.PSObject.BaseObject,$null)}
function Test-SyntheticStoreExact($Left,$Right){return ($Left -is [string] -and $Right -is [string] -and [string]::Equals($Left,$Right,[StringComparison]::Ordinal))}
function Assert-SyntheticStorePassive($Value){
    if($null -eq $Value){return}
    foreach($p in $Value.PSObject.Properties){
        if($p.MemberType -in @([Management.Automation.PSMemberTypes]::ScriptProperty,[Management.Automation.PSMemberTypes]::CodeProperty)){throw 'Executable archive property rejected'}
        if(([StringComparer]::OrdinalIgnoreCase.Equals($p.Name,'Length') -or [StringComparer]::OrdinalIgnoreCase.Equals($p.Name,'Rank')) -and $p.MemberType -ne [Management.Automation.PSMemberTypes]::Property){throw 'Archive bound shadow rejected'}
    }
}
function Assert-SyntheticStoreShape($Value,[string[]]$Names){
    if((Get-SyntheticStoreType $Value) -ne [Management.Automation.PSCustomObject]){throw 'Plain archive object required'}
    $properties=@($Value.PSObject.Properties)
    if($properties.Count -ne $Names.Count){throw 'Archive field count mismatch'}
    foreach($p in $properties){
        $matched=$false;foreach($name in $Names){if(Test-SyntheticStoreExact $p.Name $name){$matched=$true;break}}
        if(-not $matched -or $p.MemberType -ne [Management.Automation.PSMemberTypes]::NoteProperty){throw 'Exact passive archive fields required'}
    }
}
function Assert-SyntheticStoreInteger($Value,[long]$Minimum,[long]$Maximum){
    $t=Get-SyntheticStoreType $Value;if($t -ne [int] -and $t -ne [long]){throw 'Archive integer required'}
    Assert-SyntheticStorePassive $Value
    if($Value -lt $Minimum -or $Value -gt $Maximum){throw 'Archive integer exceeds bound'}
}
function Assert-SyntheticStoreText($Value,[int]$Maximum){
    if((Get-SyntheticStoreType $Value) -ne [string]){throw 'Archive string required'}
    Assert-SyntheticStorePassive $Value
    if($Value.Length -gt $Maximum){throw 'Archive string exceeds bound'}
    $utf8=New-Object Text.UTF8Encoding($false,$true)
    if($utf8.GetByteCount($Value) -gt $Maximum){throw 'Archive UTF8 exceeds bound'}
}
function Get-SyntheticStoreHash([byte[]]$Bytes){$sha=[Security.Cryptography.SHA256]::Create();try{return ,$sha.ComputeHash($Bytes)}finally{$sha.Dispose()}}
function New-SyntheticStoreSnapshot([byte[]]$Bytes){[pscustomobject]@{Base64=[Convert]::ToBase64String($Bytes);ByteLength=$Bytes.Length;SHA256=[BitConverter]::ToString((Get-SyntheticStoreHash $Bytes)).Replace('-','')}}
function Get-SyntheticStoreSnapshotBytes($Value,[int]$Maximum){
    Assert-SyntheticStoreShape $Value @('Base64','ByteLength','SHA256')
    Assert-SyntheticStoreInteger $Value.ByteLength 0 $Maximum
    Assert-SyntheticStoreText $Value.Base64 ([int](4*[Math]::Ceiling($Maximum/3.0)))
    Assert-SyntheticStoreText $Value.SHA256 64
    if($Value.Base64 -cnotmatch '^[A-Za-z0-9+/]*={0,2}\z' -or $Value.SHA256 -cnotmatch '^[A-F0-9]{64}\z'){throw 'Canonical archive snapshot required'}
    $bytes=[Convert]::FromBase64String($Value.Base64)
    if($bytes.Length -ne $Value.ByteLength -or $bytes.Length -gt $Maximum -or -not (Test-SyntheticStoreExact ([Convert]::ToBase64String($bytes)) $Value.Base64) -or -not (Test-SyntheticStoreExact ([BitConverter]::ToString((Get-SyntheticStoreHash $bytes)).Replace('-','')) $Value.SHA256)){throw 'Archive snapshot mismatch'}
    return ,$bytes
}
function Assert-SyntheticStoreField([string]$Layout,[string]$Name,$Value){
    if($Name -eq 'Schema' -and $Value -ne 1){throw 'Unknown contract schema'}
    if($Name -eq 'Policy'){
        $expected=switch($Layout){record{'synthetic-capture-contract-v1'} lifecycle{'offline-capture-simulation-v1'} package{'offline-output-evidence-v1'}}
        if(-not (Test-SyntheticStoreExact $Value $expected)){throw 'Unknown archive policy'}
    }
    if($Name -eq 'SourceMode' -and -not (Test-SyntheticStoreExact $Value 'synthetic')){throw 'Synthetic source required'}
    if($Name -eq 'SessionLabel' -and ($Value -cnotmatch '^[A-Za-z0-9][A-Za-z0-9_-]{0,79}\z')){throw 'Invalid session declaration'}
    if($Name -in @('ExecutableDeclaration','OriginDeclaration') -and ($Value.Length -lt 1 -or $Value.Length -gt 1024)){throw 'Declaration exceeds bound'}
    if($Name -eq 'EncodingName' -and -not ((Test-SyntheticStoreExact $Value 'utf-8') -or (Test-SyntheticStoreExact $Value 'utf-16le'))){throw 'Unknown encoding declaration'}
    if($Name -in @('StartedUTC','EndedUTC')){
        $parsed=[DateTimeOffset]::MinValue
        if($Value.Length -ne 28 -or -not [DateTimeOffset]::TryParseExact($Value,"yyyy-MM-dd'T'HH:mm:ss.fffffff'Z'",[Globalization.CultureInfo]::InvariantCulture,[Globalization.DateTimeStyles]::AssumeUniversal,[ref]$parsed)){throw 'Exact declared UTC required'}
    }
    if($Name -eq 'CleanupStatus' -and -not (Test-SyntheticStoreExact $Value 'unknown')){throw 'Cleanup cannot be certified'}
    if($Name -in @('LiveActionsPerformed','OriginVerified','Acceptance','RuntimeApproved') -and $Value){throw 'Archive cannot grant authority'}
    if($Name -in @('AcceptedEventCount','ProgressCount') -and $Value -lt 0){throw 'Negative derived count'}
    if($Name -eq 'AcceptedEventCount' -and $Value -gt 4096){throw 'Derived count exceeds bound'}
    if($Name -eq 'FailureEventIndex' -and $null -ne $Value -and ($Value -lt 0 -or $Value -gt 4095)){throw 'Failure index exceeds bound'}
    if($Name -eq 'LastObservedElapsedMs' -and $Value -lt 0){throw 'Negative elapsed time'}
}
function Write-SyntheticStoreValue([IO.BinaryWriter]$Writer,[string]$Type,$Value,[int]$Maximum){
    switch($Type){
        i32{Assert-SyntheticStoreInteger $Value ([int]::MinValue) ([int]::MaxValue);$Writer.Write([int]$Value)}
        i64{Assert-SyntheticStoreInteger $Value 0 ([long]::MaxValue);$Writer.Write([long]$Value)}
        bool{if((Get-SyntheticStoreType $Value) -ne [bool]){throw 'Archive boolean required'};Assert-SyntheticStorePassive $Value;$Writer.Write([byte][int]$Value)}
        nullable32{if($null -eq $Value){$Writer.Write([byte]0)}else{$Writer.Write([byte]1);Write-SyntheticStoreValue $Writer 'i32' $Value 0}}
        text{Assert-SyntheticStoreText $Value $Maximum;$utf8=New-Object Text.UTF8Encoding($false,$true);$b=$utf8.GetBytes($Value);$Writer.Write([int]$b.Length);$Writer.Write([byte[]]$b)}
        snapshot{$b=Get-SyntheticStoreSnapshotBytes $Value $Maximum;$Writer.Write([int]$b.Length);$Writer.Write([byte[]]$b)}
        digest{Assert-SyntheticStoreText $Value 64;if($Value -cnotmatch '^[A-F0-9]{64}\z'){throw 'Exact binding digest required'};$b=New-Object byte[] 32;for($i=0;$i -lt 32;$i++){$b[$i]=[Convert]::ToByte($Value.Substring(2*$i,2),16)};$Writer.Write([byte[]]$b)}
        lifecycle{Write-SyntheticStoreObject $Writer $Value 'lifecycle'}
        optionalpackage{if($null -eq $Value){$Writer.Write([byte]0)}else{$Writer.Write([byte]1);Write-SyntheticStoreObject $Writer $Value 'package'}}
        events{Write-SyntheticStoreEvents $Writer $Value}
        default{throw 'Unknown internal archive field type'}
    }
}
function Write-SyntheticStoreObject([IO.BinaryWriter]$Writer,$Value,[string]$Layout){
    $fields=$script:SyntheticStoreLayouts[$Layout];$names=@($fields|ForEach-Object{($_ -split ':')[0]})
    Assert-SyntheticStoreShape $Value $names
    foreach($field in $fields){$parts=$field -split ':';$v=$Value.($parts[0]);Write-SyntheticStoreValue $Writer $parts[1] $v ([int]$parts[2]);Assert-SyntheticStoreField $Layout $parts[0] $v}
}
function Write-SyntheticStoreEvents([IO.BinaryWriter]$Writer,$Events){
    if((Get-SyntheticStoreType $Events) -ne [object[]]){throw 'Object event array required'}
    Assert-SyntheticStorePassive $Events
    if($Events.Rank -ne 1 -or $Events.Length -gt 4096){throw 'Event count exceeds bound'}
    $Writer.Write([int]$Events.Length);$total=[long]0
    foreach($e in $Events){
        # All descriptors checked before Kind or any scalar is read.
        if((Get-SyntheticStoreType $e) -ne [Management.Automation.PSCustomObject]){throw 'Plain stored event required'}
        foreach($p in $e.PSObject.Properties){if($p.MemberType -ne [Management.Automation.PSMemberTypes]::NoteProperty){throw 'Passive stored event required'}}
        Assert-SyntheticStoreText $e.Kind 5
        $kind=if(Test-SyntheticStoreExact $e.Kind 'chunk'){1}elseif(Test-SyntheticStoreExact $e.Kind 'eof'){2}elseif(Test-SyntheticStoreExact $e.Kind 'exit'){3}elseif(Test-SyntheticStoreExact $e.Kind 'tick'){4}else{throw 'Unknown event kind'}
        $names=switch($kind){1{@('Kind','ElapsedMs','Channel','Data')} 2{@('Kind','ElapsedMs','Channel')} 3{@('Kind','ElapsedMs','ExitCode')} 4{@('Kind','ElapsedMs')}}
        Assert-SyntheticStoreShape $e $names
        $Writer.Write([byte]$kind);Write-SyntheticStoreValue $Writer 'i64' $e.ElapsedMs 0
        if($kind -in @(1,2)){
            Assert-SyntheticStoreText $e.Channel 6
            $channel=if(Test-SyntheticStoreExact $e.Channel 'stdout'){1}elseif(Test-SyntheticStoreExact $e.Channel 'stderr'){2}else{throw 'Unknown event channel'}
            $Writer.Write([byte]$channel)
        }
        if($kind -eq 1){$b=Get-SyntheticStoreSnapshotBytes $e.Data 65536;$total+=$b.Length;if($total -gt 2097152){throw 'Whole archive trace exceeds bound'};$Writer.Write([int]$b.Length);$Writer.Write([byte[]]$b)}
        if($kind -eq 3){Write-SyntheticStoreValue $Writer 'i32' $e.ExitCode 0}
    }
}
function Read-SyntheticStoreExact([IO.BinaryReader]$Reader,[int]$Count){
    if($Count -lt 0 -or $Count -gt ($Reader.BaseStream.Length-$Reader.BaseStream.Position)){throw 'Truncated archive field'}
    $bytes=$Reader.ReadBytes($Count);if($bytes.Length -ne $Count){throw 'Truncated archive bytes'};return ,$bytes
}
function Read-SyntheticStoreBlob([IO.BinaryReader]$Reader,[int]$Maximum){$n=$Reader.ReadInt32();if($n -lt 0 -or $n -gt $Maximum){throw 'Archive field exceeds bound'};return ,(Read-SyntheticStoreExact $Reader $n)}
function Read-SyntheticStoreBool([IO.BinaryReader]$Reader){$b=$Reader.ReadByte();if($b -gt 1){throw 'Noncanonical boolean'};return ($b -eq 1)}
function Read-SyntheticStoreValue([IO.BinaryReader]$Reader,[string]$Type,[int]$Maximum){
    switch($Type){
        i32{return $Reader.ReadInt32()}
        i64{$v=$Reader.ReadInt64();if($v -lt 0){throw 'Negative elapsed time'};return $v}
        bool{return (Read-SyntheticStoreBool $Reader)}
        nullable32{if(Read-SyntheticStoreBool $Reader){return $Reader.ReadInt32()};return $null}
        text{$b=Read-SyntheticStoreBlob $Reader $Maximum;$utf8=New-Object Text.UTF8Encoding($false,$true);return $utf8.GetString($b)}
        snapshot{return (New-SyntheticStoreSnapshot (Read-SyntheticStoreBlob $Reader $Maximum))}
        digest{return [BitConverter]::ToString((Read-SyntheticStoreExact $Reader 32)).Replace('-','')}
        lifecycle{return (Read-SyntheticStoreObject $Reader 'lifecycle')}
        optionalpackage{if(Read-SyntheticStoreBool $Reader){return (Read-SyntheticStoreObject $Reader 'package')};return $null}
        events{return ,(Read-SyntheticStoreEvents $Reader)}
        default{throw 'Unknown internal archive field type'}
    }
}
function Read-SyntheticStoreObject([IO.BinaryReader]$Reader,[string]$Layout){
    $record=[ordered]@{}
    foreach($field in $script:SyntheticStoreLayouts[$Layout]){$parts=$field -split ':';$v=Read-SyntheticStoreValue $Reader $parts[1] ([int]$parts[2]);Assert-SyntheticStoreField $Layout $parts[0] $v;$record[$parts[0]]=$v}
    return [pscustomobject]$record
}
function Read-SyntheticStoreEvents([IO.BinaryReader]$Reader){
    $count=$Reader.ReadInt32();if($count -lt 0 -or $count -gt 4096){throw 'Event count exceeds bound'}
    $events=New-Object 'System.Collections.Generic.List[object]';$total=[long]0
    for($i=0;$i -lt $count;$i++){
        $kind=$Reader.ReadByte();if($kind -lt 1 -or $kind -gt 4){throw 'Unknown event kind'}
        $time=Read-SyntheticStoreValue $Reader 'i64' 0
        if($kind -in @(1,2)){$c=$Reader.ReadByte();if($c -ne 1 -and $c -ne 2){throw 'Unknown event channel'};$channel=if($c -eq 1){'stdout'}else{'stderr'}}
        switch($kind){
            1{$b=Read-SyntheticStoreBlob $Reader 65536;$total+=$b.Length;if($total -gt 2097152){throw 'Whole archive trace exceeds bound'};$events.Add([pscustomobject]@{Kind='chunk';ElapsedMs=$time;Channel=$channel;Data=(New-SyntheticStoreSnapshot $b)})}
            2{$events.Add([pscustomobject]@{Kind='eof';ElapsedMs=$time;Channel=$channel})}
            3{$events.Add([pscustomobject]@{Kind='exit';ElapsedMs=$time;ExitCode=$Reader.ReadInt32()})}
            4{$events.Add([pscustomobject]@{Kind='tick';ElapsedMs=$time})}
        }
    }
    return ,([object[]]$events.ToArray().Clone())
}
function Assert-SyntheticStorePath([string]$Directory){
    if([string]::IsNullOrWhiteSpace($Directory) -or $Directory -notmatch '^[A-Za-z]:[\\/]'){throw 'Absolute local archive directory required'}
    $full=Get-IsolatedFullPath $Directory
    if(-not [string]::Equals($full,$Directory.Replace('/','\'),[StringComparison]::OrdinalIgnoreCase)){throw 'Noncanonical archive path'}
    $root=Get-IsolatedFullPath $script:SyntheticStoreRoot
    if(-not [string]::Equals([IO.Path]::GetDirectoryName($full),$root,[StringComparison]::OrdinalIgnoreCase) -or [IO.Path]::GetFileName($full) -cnotmatch '^[A-Za-z0-9][A-Za-z0-9_-]{0,79}\z'){throw 'Archive destination outside direct-child scope'}
    return $full
}
function Open-SyntheticStoreInput([string]$Path,[long]$Maximum){
    $null=Get-IsolatedFullPath $Path;$s=[IO.File]::Open($Path,[IO.FileMode]::Open,[IO.FileAccess]::Read,[IO.FileShare]::Read)
    try{$info=[IsolatedFuzzFileInfo]::Snapshot($s.SafeFileHandle);if($info.Links -ne 1 -or ($info.Attributes -band [uint32][IO.FileAttributes]::ReparsePoint) -ne 0 -or $s.Length -gt $Maximum){throw 'Unsafe or oversized archive file'};return $s}catch{$s.Dispose();throw}
}
function Write-SyntheticStoreNewFile([string]$Path,[byte[]]$Bytes){
    $null=Get-IsolatedFullPath $Path;$s=[IO.File]::Open($Path,[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::None)
    try{$info=[IsolatedFuzzFileInfo]::Snapshot($s.SafeFileHandle);if($info.Links -ne 1 -or ($info.Attributes -band [uint32][IO.FileAttributes]::ReparsePoint) -ne 0){throw 'Unsafe archive output'};$s.Write($Bytes,0,$Bytes.Length);$s.Flush()}finally{$s.Dispose()}
}
function Read-OfflineSyntheticCaptureStore([string]$Directory){
    $result=[pscustomobject]@{Schema=1;Verified=$false;Reason='invalid-path';Record=$null;OriginVerified=$false;Acceptance=$false;RuntimeApproved=$false}
    $payload=$null;$marker=$null;$pr=$null;$mr=$null;$memory=$null;$reader=$null
    try{
        $dir=Assert-SyntheticStorePath $Directory
        $result.Reason='missing-archive';if(-not (Test-Path -LiteralPath $dir -PathType Container)){return $result}
        $result.Reason='incomplete-archive'
        if((Test-Path -LiteralPath (Join-Path $dir 'stopped')) -or -not (Test-Path -LiteralPath (Join-Path $dir 'record.bin') -PathType Leaf) -or -not (Test-Path -LiteralPath (Join-Path $dir 'complete.bin') -PathType Leaf)){return $result}
        $result.Reason='io-error';$marker=Open-SyntheticStoreInput (Join-Path $dir 'complete.bin') 52;$payload=Open-SyntheticStoreInput (Join-Path $dir 'record.bin') 8388608
        $result.Reason='incomplete-archive';if($marker.Length -ne 52){return $result}
        $mr=New-Object IO.BinaryReader($marker);$result.Reason='unsupported-format'
        if(-not (Test-SyntheticStoreExact ([Text.Encoding]::ASCII.GetString((Read-SyntheticStoreExact $mr 8))) 'TLRSYC01') -or $mr.ReadInt32() -ne 1){return $result}
        $length=$mr.ReadInt64();$digest=Read-SyntheticStoreExact $mr 32;$result.Reason='integrity-mismatch'
        if($length -lt 0 -or $length -gt 8388608 -or $length -ne $payload.Length){return $result}
        $pr=New-Object IO.BinaryReader($payload);$bytes=Read-SyntheticStoreExact $pr ([int]$payload.Length)
        if(-not (Test-SyntheticStoreExact ([Convert]::ToBase64String((Get-SyntheticStoreHash $bytes))) ([Convert]::ToBase64String($digest)))){return $result}
        $memory=New-Object IO.MemoryStream(,$bytes);$reader=New-Object IO.BinaryReader($memory);$result.Reason='unsupported-format'
        if(-not (Test-SyntheticStoreExact ([Text.Encoding]::ASCII.GetString((Read-SyntheticStoreExact $reader 8))) 'TLRSYN01') -or $reader.ReadInt32() -ne 1){return $result}
        $result.Reason='invalid-payload';$record=Read-SyntheticStoreObject $reader 'record'
        if($memory.Position -ne $memory.Length){return $result}
        $result.Reason='invalid-record';if(-not (Test-SyntheticCaptureRecord $record).Verified){return $result}
        $result.Reason='incomplete-archive';if(Test-Path -LiteralPath (Join-Path $dir 'stopped')){return $result}
        $result.Record=$record;$result.Verified=$true;$result.Reason='consistent-synthetic-archive-only'
    }catch{}finally{foreach($resource in @($reader,$memory,$pr,$mr,$payload,$marker)){if($null -ne $resource){$resource.Dispose()}}}
    return $result
}
function Write-OfflineSyntheticCaptureStore([string]$Directory,[object]$Record){
    $dir=Assert-SyntheticStorePath $Directory
    if(Test-Path -LiteralPath $dir){throw 'Archive destination already exists'}
    # Encode/validate every caller field before any filesystem creation.
    $memory=New-Object IO.MemoryStream;$writer=New-Object IO.BinaryWriter($memory)
    try{$writer.Write([Text.Encoding]::ASCII.GetBytes('TLRSYN01'));$writer.Write([int]1);Write-SyntheticStoreObject $writer $Record 'record';$writer.Flush();if($memory.Length -gt 8388608){throw 'Archive payload exceeds bound'};$bytes=$memory.ToArray()}finally{$writer.Dispose();$memory.Dispose()}
    if(-not (Test-SyntheticCaptureRecord $Record).Verified){throw 'Consistent synthetic record required'}
    $root=Get-IsolatedFullPath $script:SyntheticStoreRoot
    if(-not (Test-Path -LiteralPath $root)){New-Item -ItemType Directory -Path $root -ErrorAction Stop|Out-Null}
    $null=Get-IsolatedFullPath $root
    $created=$false;$binding=$null;$reader=$null
    try{
        New-Item -ItemType Directory -Path $dir -ErrorAction Stop|Out-Null;$created=$true;$null=Assert-SyntheticStorePath $dir
        Write-SyntheticStoreNewFile (Join-Path $dir 'record.bin') $bytes
        $binding=Open-SyntheticStoreInput (Join-Path $dir 'record.bin') 8388608;$reader=New-Object IO.BinaryReader($binding)
        $stored=Read-SyntheticStoreExact $reader ([int]$binding.Length)
        if($stored.Length -ne $bytes.Length -or -not (Test-SyntheticStoreExact ([Convert]::ToBase64String((Get-SyntheticStoreHash $stored))) ([Convert]::ToBase64String((Get-SyntheticStoreHash $bytes))))){throw 'Written archive reread mismatch'}
        $markerMemory=New-Object IO.MemoryStream;$markerWriter=New-Object IO.BinaryWriter($markerMemory)
        try{$markerWriter.Write([Text.Encoding]::ASCII.GetBytes('TLRSYC01'));$markerWriter.Write([int]1);$markerWriter.Write([long]$stored.Length);$markerWriter.Write([byte[]](Get-SyntheticStoreHash $stored));$markerWriter.Flush();$markerBytes=$markerMemory.ToArray()}finally{$markerWriter.Dispose();$markerMemory.Dispose()}
        Write-SyntheticStoreNewFile (Join-Path $dir 'complete.bin') $markerBytes
        $reader.Dispose();$reader=$null;$binding.Dispose();$binding=$null
        $result=Read-OfflineSyntheticCaptureStore $dir;if(-not $result.Verified){throw 'Final archive verification failed'};return $result
    }catch{if($created){try{Write-SyntheticStoreNewFile (Join-Path $dir 'stopped') ([byte[]]@(1))}catch{}};throw}finally{if($null -ne $reader){$reader.Dispose()};if($null -ne $binding){$binding.Dispose()}}
}
