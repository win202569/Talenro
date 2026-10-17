# Pure synthetic manifest/trace contract. No live adapter or persistence path.
. (Join-Path $PSScriptRoot 'output-capture-simulation.ps1')
function Assert-ContractPassiveProperties($Value){
    if($null -eq $Value){return}
    foreach($p in $Value.PSObject.Properties){
        if($p.MemberType -in @([Management.Automation.PSMemberTypes]::ScriptProperty,[Management.Automation.PSMemberTypes]::CodeProperty)){throw 'Executable property rejected'}
        # Only CLR properties may supply intrinsic bounds;ETS notes/aliases can clip bytes.
        if(([StringComparer]::OrdinalIgnoreCase.Equals($p.Name,'Length') -or [StringComparer]::OrdinalIgnoreCase.Equals($p.Name,'Rank')) -and $p.MemberType -ne [Management.Automation.PSMemberTypes]::Property){throw 'Intrinsic bound shadow rejected'}
    }
}
function Assert-ContractShape($Value,[string[]]$Names){
    $t=Get-OfflineCaptureValueType $Value
    if($null -eq $t -or $t.FullName -cne 'System.Management.Automation.PSCustomObject'){throw 'Plain object required'}
    $props=@($Value.PSObject.Properties)
    if($props.Count -ne $Names.Count){throw 'Wrong field count'}
    foreach($p in $props){if($p.MemberType -ne [Management.Automation.PSMemberTypes]::NoteProperty -or $p.Name -cnotin $Names){throw 'Exact passive fields required'}}
}
function Assert-ContractString($Value,[int]$Minimum,[int]$Maximum){
    if((Get-OfflineCaptureValueType $Value) -ne [string]){throw 'String required'}
    Assert-ContractPassiveProperties $Value
    if($Value.Length -lt $Minimum -or $Value.Length -gt $Maximum){throw 'String exceeds bound'}
}
function Assert-ContractText($Value){
    Assert-ContractString $Value 1 1024
    $utf8=New-Object Text.UTF8Encoding($false,$true)
    if($utf8.GetByteCount($Value) -gt 4096){throw 'UTF8 exceeds bound'}
}
function Get-ContractUTC($Value){
    Assert-ContractString $Value 28 28
    $parsed=[DateTimeOffset]::MinValue
    if(-not [DateTimeOffset]::TryParseExact($Value,"yyyy-MM-dd'T'HH:mm:ss.fffffff'Z'",[Globalization.CultureInfo]::InvariantCulture,[Globalization.DateTimeStyles]::AssumeUniversal,[ref]$parsed)){throw 'Exact UTC required'}
    return $parsed
}
function Copy-ContractBytes($Value,[int]$Maximum){
    if((Get-OfflineCaptureValueType $Value) -ne [byte[]]){throw 'Raw bytes required'}
    Assert-ContractPassiveProperties $Value
    if($Value.Length -gt $Maximum){throw 'Byte bound exceeded'}
    $copy=New-Object byte[] $Value.Length;[Array]::Copy($Value,$copy,$Value.Length)
    return ,$copy
}
function Assert-ContractArray($Value){
    if((Get-OfflineCaptureValueType $Value) -ne [object[]]){throw 'Exact Object array required'}
    Assert-ContractPassiveProperties $Value
    if($Value.Rank -ne 1 -or $Value.Length -gt 4096){throw 'Event count exceeded'}
}
function Get-ContractPacket($Packet){
    Assert-ContractShape $Packet @('Schema','SourceMode','SessionLabel','ExecutableDeclaration','ArgumentBytes','EncodingName','OriginDeclaration','StartedUTC','EndedUTC','Events')
    Assert-OfflineEvidenceInteger $Packet.Schema 1 1
    Assert-ContractString $Packet.SourceMode 9 9
    if($Packet.SourceMode -cne 'synthetic'){throw 'Synthetic only'}
    Assert-ContractString $Packet.SessionLabel 1 80
    if($Packet.SessionLabel -cnotmatch '^[A-Za-z0-9][A-Za-z0-9_-]{0,79}\z'){throw 'Invalid session label'}
    Assert-ContractText $Packet.ExecutableDeclaration;Assert-ContractText $Packet.OriginDeclaration
    Assert-ContractString $Packet.EncodingName 1 8
    if($Packet.EncodingName -cnotin @('utf-8','utf-16le')){throw 'Explicit encoding required'}
    $null=Get-ContractUTC $Packet.StartedUTC;$null=Get-ContractUTC $Packet.EndedUTC
    $argsCopy=Copy-ContractBytes $Packet.ArgumentBytes 65536
    if(($argsCopy.Length%2) -ne 0 -or ($argsCopy.Length -ge 2 -and $argsCopy[0] -eq 255 -and $argsCopy[1] -eq 254)){throw 'UTF16LE arguments without BOM required'}
    $utf16=New-Object Text.UnicodeEncoding($false,$false,$true);$null=$utf16.GetString($argsCopy)
    Assert-ContractArray $Packet.Events
    $events=New-Object 'System.Collections.Generic.List[object]';$total=[long]0
    foreach($e in $Packet.Events){
        $type=Get-OfflineCaptureValueType $e
        if($null -eq $type -or $type.FullName -cne 'System.Management.Automation.PSCustomObject'){throw 'Plain event required'}
        foreach($prop in $e.PSObject.Properties){if($prop.MemberType -ne [Management.Automation.PSMemberTypes]::NoteProperty){throw 'Passive event required'}}
        foreach($prop in $e.PSObject.Properties){Assert-ContractPassiveProperties $prop.Value}
        Assert-OfflineCaptureEvent $e
        switch($e.Kind){
            chunk{
                $total+=$e.Bytes.Length;if($total -gt 2097152){throw 'Whole trace bound exceeded'}
                $events.Add([pscustomobject]@{Kind='chunk';ElapsedMs=[long]$e.ElapsedMs;Channel=[string]$e.Channel;Bytes=(Copy-ContractBytes $e.Bytes 65536)})
            }
            eof{$events.Add([pscustomobject]@{Kind='eof';ElapsedMs=[long]$e.ElapsedMs;Channel=[string]$e.Channel})}
            exit{$events.Add([pscustomobject]@{Kind='exit';ElapsedMs=[long]$e.ElapsedMs;ExitCode=[int]$e.ExitCode})}
            tick{$events.Add([pscustomobject]@{Kind='tick';ElapsedMs=[long]$e.ElapsedMs})}
        }
    }
    return [pscustomobject]@{Schema=1;SourceMode='synthetic';SessionLabel=[string]$Packet.SessionLabel;ExecutableDeclaration=[string]$Packet.ExecutableDeclaration;ArgumentBytes=$argsCopy;EncodingName=[string]$Packet.EncodingName;OriginDeclaration=[string]$Packet.OriginDeclaration;StartedUTC=[string]$Packet.StartedUTC;EndedUTC=[string]$Packet.EndedUTC;Events=[object[]]$events.ToArray().Clone()}
}
function Get-ContractBindingHash($Packet){
    $memory=New-Object IO.MemoryStream;$writer=New-Object IO.BinaryWriter($memory)
    try{
        $writer.Write([Text.Encoding]::ASCII.GetBytes('TLRCAP01'));$writer.Write([int]1)
        $encoding=if($Packet.EncodingName -ceq 'utf-8'){1}else{2};$writer.Write([int]$encoding)
        $utf8=New-Object Text.UTF8Encoding($false,$true)
        foreach($s in @($Packet.SessionLabel,$Packet.ExecutableDeclaration,$Packet.OriginDeclaration,$Packet.StartedUTC,$Packet.EndedUTC)){
            $bytes=$utf8.GetBytes($s);$writer.Write([int]$bytes.Length);$writer.Write([byte[]]$bytes)
        }
        $writer.Write([int]$Packet.ArgumentBytes.Length);$writer.Write([byte[]]$Packet.ArgumentBytes);$writer.Write([int]$Packet.Events.Length)
        foreach($e in $Packet.Events){
            $kind=switch($e.Kind){chunk{1} eof{2} exit{3} tick{4}}
            $writer.Write([byte]$kind);$writer.Write([long]$e.ElapsedMs)
            if($e.Kind -cin @('chunk','eof')){$channel=if($e.Channel -ceq 'stdout'){1}else{2};$writer.Write([byte]$channel)}
            if($e.Kind -ceq 'chunk'){$writer.Write([int]$e.Bytes.Length);$writer.Write([byte[]]$e.Bytes)}
            if($e.Kind -ceq 'exit'){$writer.Write([int]$e.ExitCode)}
        }
        $writer.Flush();return (Get-OfflineOutputByteInfo $memory.ToArray()).SHA256
    }finally{$writer.Dispose();$memory.Dispose()}
}
function New-ContractStoredEvents($Events){
    $stored=New-Object 'System.Collections.Generic.List[object]'
    foreach($e in $Events){
        switch($e.Kind){
            chunk{$stored.Add([pscustomobject]@{Kind='chunk';ElapsedMs=$e.ElapsedMs;Channel=$e.Channel;Data=(New-OfflineEvidenceStream $e.Bytes)})}
            eof{$stored.Add([pscustomobject]@{Kind='eof';ElapsedMs=$e.ElapsedMs;Channel=$e.Channel})}
            exit{$stored.Add([pscustomobject]@{Kind='exit';ElapsedMs=$e.ElapsedMs;ExitCode=$e.ExitCode})}
            tick{$stored.Add([pscustomobject]@{Kind='tick';ElapsedMs=$e.ElapsedMs})}
        }
    }
    # List<T>.ToArray may share its empty array;the returned record must own it.
    return ,([object[]]$stored.ToArray().Clone())
}
function New-ContractValidatedRecord($Packet){
    $lifecycle=Get-OfflineCaptureSimulation -Events $Packet.Events -EncodingName $Packet.EncodingName -OriginDeclaration $Packet.OriginDeclaration
    $intent=switch($lifecycle.Status){complete{'none'} incomplete{'none'} timed-out{'would-request-owned-stop'} refused{'would-refuse-and-stop-if-owned'}}
    [pscustomobject]@{Schema=1;Policy='synthetic-capture-contract-v1';SourceMode='synthetic';SessionLabel=$Packet.SessionLabel;ExecutableDeclaration=$Packet.ExecutableDeclaration;Arguments=(New-OfflineEvidenceStream $Packet.ArgumentBytes);EncodingName=$Packet.EncodingName;OriginDeclaration=$Packet.OriginDeclaration;StartedUTC=$Packet.StartedUTC;EndedUTC=$Packet.EndedUTC;UtcOrderConsistent=((Get-ContractUTC $Packet.EndedUTC) -ge (Get-ContractUTC $Packet.StartedUTC));Events=(New-ContractStoredEvents $Packet.Events);Lifecycle=$lifecycle;StopIntent=$intent;CleanupStatus='unknown';LiveActionsPerformed=$false;BindingSHA256=(Get-ContractBindingHash $Packet);OriginVerified=$false;Acceptance=$false;RuntimeApproved=$false}
}
function New-SyntheticCaptureRecord([object]$Packet){
    try{$clean=Get-ContractPacket $Packet;return (New-ContractValidatedRecord $clean)}catch{
        return [pscustomobject]@{Schema=1;SourceMode='synthetic';Created=$false;Reason='invalid-packet';OriginVerified=$false;Acceptance=$false;RuntimeApproved=$false}
    }
}
function Read-ContractSnapshot($Value,[int]$Maximum){
    Assert-ContractShape $Value @('Base64','ByteLength','SHA256')
    Assert-OfflineEvidenceInteger $Value.ByteLength 0 $Maximum
    $maxEncoded=[int](4*[Math]::Ceiling($Maximum/3.0))
    Assert-ContractString $Value.Base64 0 $maxEncoded;Assert-ContractString $Value.SHA256 64 64
    if($Value.Base64 -cnotmatch '^[A-Za-z0-9+/]*={0,2}\z' -or $Value.SHA256 -cnotmatch '^[A-F0-9]{64}\z'){throw 'Invalid snapshot'}
    $bytes=[Convert]::FromBase64String($Value.Base64)
    if($bytes.Length -gt $Maximum -or $bytes.Length -ne $Value.ByteLength -or [Convert]::ToBase64String($bytes) -cne $Value.Base64 -or (Get-OfflineOutputByteInfo $bytes).SHA256 -cne $Value.SHA256){throw 'Snapshot mismatch'}
    return ,$bytes
}
function Read-ContractStoredEvents($Events){
    Assert-ContractArray $Events
    $raw=New-Object 'System.Collections.Generic.List[object]';$total=[long]0
    foreach($e in $Events){
        $type=Get-OfflineCaptureValueType $e
        if($null -eq $type -or $type.FullName -cne 'System.Management.Automation.PSCustomObject'){throw 'Plain stored event required'}
        foreach($p in $e.PSObject.Properties){if($p.MemberType -ne [Management.Automation.PSMemberTypes]::NoteProperty){throw 'Passive stored event required'}}
        Assert-ContractString $e.Kind 1 5
        switch -CaseSensitive ($e.Kind){
            chunk{
                Assert-ContractShape $e @('Kind','ElapsedMs','Channel','Data');$bytes=Read-ContractSnapshot $e.Data 65536
                $total+=$bytes.Length;if($total -gt 2097152){throw 'Trace bound exceeded'}
                $raw.Add([pscustomobject]@{Kind='chunk';ElapsedMs=$e.ElapsedMs;Channel=$e.Channel;Bytes=$bytes})
            }
            eof{Assert-ContractShape $e @('Kind','ElapsedMs','Channel');$raw.Add([pscustomobject]@{Kind='eof';ElapsedMs=$e.ElapsedMs;Channel=$e.Channel})}
            exit{Assert-ContractShape $e @('Kind','ElapsedMs','ExitCode');$raw.Add([pscustomobject]@{Kind='exit';ElapsedMs=$e.ElapsedMs;ExitCode=$e.ExitCode})}
            tick{Assert-ContractShape $e @('Kind','ElapsedMs');$raw.Add([pscustomobject]@{Kind='tick';ElapsedMs=$e.ElapsedMs})}
            default{throw 'Unknown stored event'}
        }
    }
    return ,([object[]]$raw.ToArray())
}
function Assert-ContractEquivalent($Actual,$Expected,[int]$Depth=0){
    if($Depth -gt 8){throw 'Nested bound exceeded'}
    if($null -eq $Expected){if($null -ne $Actual){throw 'Expected null'};return}
    $type=Get-OfflineCaptureValueType $Expected
    if($type -eq [string]){Assert-ContractString $Actual $Expected.Length $Expected.Length;if($Actual -cne $Expected){throw 'String mismatch'};return}
    if($type -eq [bool]){if((Get-OfflineCaptureValueType $Actual) -ne [bool] -or $Actual -ne $Expected){throw 'Boolean mismatch'};return}
    if($type -in @([int],[long])){Assert-OfflineEvidenceInteger $Actual ([long]::MinValue) ([long]::MaxValue);if($Actual -ne $Expected){throw 'Integer mismatch'};return}
    if($type -eq [object[]]){
        Assert-ContractArray $Actual;if($Actual.Length -ne $Expected.Length){throw 'Array length mismatch'}
        for($i=0;$i -lt $Expected.Length;$i++){Assert-ContractEquivalent $Actual[$i] $Expected[$i] ($Depth+1)};return
    }
    if($type.FullName -cne 'System.Management.Automation.PSCustomObject'){throw 'Unsupported value'}
    $names=@($Expected.PSObject.Properties|ForEach-Object{$_.Name});Assert-ContractShape $Actual $names
    foreach($name in $names){Assert-ContractEquivalent $Actual.$name $Expected.$name ($Depth+1)}
}
function Test-SyntheticCaptureRecord([object]$Record){
    $v=[pscustomobject]@{Verified=$false;Reason='invalid-record';OriginVerified=$false;Acceptance=$false;RuntimeApproved=$false}
    try{
        Assert-ContractShape $Record @('Schema','Policy','SourceMode','SessionLabel','ExecutableDeclaration','Arguments','EncodingName','OriginDeclaration','StartedUTC','EndedUTC','UtcOrderConsistent','Events','Lifecycle','StopIntent','CleanupStatus','LiveActionsPerformed','BindingSHA256','OriginVerified','Acceptance','RuntimeApproved')
        Assert-OfflineEvidenceInteger $Record.Schema 1 1
        Assert-ContractString $Record.Policy 29 29
        if($Record.Policy -cne 'synthetic-capture-contract-v1'){throw 'Unknown policy'}
        foreach($flag in @('LiveActionsPerformed','OriginVerified','Acceptance','RuntimeApproved')){if((Get-OfflineCaptureValueType $Record.$flag) -ne [bool] -or $Record.$flag){throw 'No action or authority allowed'}}
        $arguments=Read-ContractSnapshot $Record.Arguments 65536;$events=Read-ContractStoredEvents $Record.Events
        $packet=[pscustomobject]@{Schema=1;SourceMode=$Record.SourceMode;SessionLabel=$Record.SessionLabel;ExecutableDeclaration=$Record.ExecutableDeclaration;ArgumentBytes=$arguments;EncodingName=$Record.EncodingName;OriginDeclaration=$Record.OriginDeclaration;StartedUTC=$Record.StartedUTC;EndedUTC=$Record.EndedUTC;Events=$events}
        $clean=Get-ContractPacket $packet;$expected=New-ContractValidatedRecord $clean
        Assert-ContractString $Record.BindingSHA256 64 64
        if($Record.BindingSHA256 -cnotmatch '^[A-F0-9]{64}\z' -or $Record.BindingSHA256 -cne $expected.BindingSHA256){$v.Reason='binding-mismatch';throw 'Binding mismatch'}
        $v.Reason='derived-state-mismatch';Assert-ContractEquivalent $Record $expected
        if(-not $expected.UtcOrderConsistent){$v.Reason='clock-order-inconsistent';throw 'Clock reversal'}
        $v.Verified=$true;$v.Reason='consistent-synthetic-only'
    }catch{}
    return $v
}
