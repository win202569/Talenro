# Definitions-only excerpt through BC-Bind; original frozen test unchanged.
$script:typeMethod=[object].GetMethod('GetType',[type[]]@())
$script:invokeMethod=[Reflection.MethodBase].GetMethod('Invoke',[type[]]@([object],[object[]]))
$script:referenceMethod=[object].GetMethod('ReferenceEquals',[type[]]@([object],[object]))
function BC-Type($Value){if($null -eq $Value){return $null};$argsLocal=[object[]]@($Value.PSObject.BaseObject,$null);$script:invokeMethod.Invoke($script:typeMethod,$argsLocal)}
function BC-Ref($Left,$Right){$argsLocal=[object[]]@($Left,$Right);[bool]$script:referenceMethod.Invoke($null,$argsLocal)}
function BC-Need([bool]$Value){if(-not $Value){throw 'Assertion'}}
function BC-Case([string]$Name,[scriptblock]$Body){
    if($Group -cne 'all' -and $Name -cne 'ApiRequired' -and
       -not ($Group -ceq 'reserved-tail' -and $Name -ceq 'ReservedZero') -and
       -not ($Group -ceq 'sequence-equality' -and $Name -ceq 'SequenceExact')){return}
    $script:case=$Name;& $Body;[Console]::Out.WriteLine('PASS: P2BCodec'+$Name)
}
function BC-Fresh($Value){
    if($null -eq $Value){return $null}
    $type=BC-Type $Value;$base=$Value.PSObject.BaseObject
    if($type -eq [byte[]]){return ,([byte[]]$base.Clone())}
    if($type -eq [string]){return ,([string]::new($base.ToCharArray()))}
    $storage=[Array]::CreateInstance($type,1);$storage.SetValue($base,0);return ,($storage.GetValue(0))
}
function BC-Object($Fields){
    $result=[pscustomobject][ordered]@{}
    foreach($name in $Fields.Keys){$result.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new($name,(BC-Fresh $Fields[$name])))}
    return $result
}
function BC-Equal($Actual,$Expected,[int]$Depth=0){
    if($Depth -gt 4){return $false}
    if($null -eq $Expected){return ($null -eq $Actual)}
    $type=BC-Type $Expected
    if((BC-Type $Actual) -ne $type){return $false}
    if($type -eq [Management.Automation.PSCustomObject]){
        $got=@($Actual.PSObject.Properties);$want=@($Expected.PSObject.Properties)
        if($got.Count -ne $want.Count){return $false}
        foreach($p in $want){$found=@($got|Where-Object{[StringComparer]::Ordinal.Equals($_.Name,$p.Name)})
            if($found.Count -ne 1 -or $found[0].MemberType -ne [Management.Automation.PSMemberTypes]::NoteProperty -or -not (BC-Equal $found[0].Value $p.Value ($Depth+1))){return $false}
        };return $true
    }
    if($type -eq [byte[]]){
        if($Actual.PSObject.BaseObject.GetLength(0) -ne $Expected.PSObject.BaseObject.GetLength(0)){return $false}
        for($i=0;$i -lt $Expected.PSObject.BaseObject.GetLength(0);$i++){if($Actual[$i] -ne $Expected[$i]){return $false}}
        return $true
    }
    if($type -eq [string]){return [StringComparer]::Ordinal.Equals($Actual.PSObject.BaseObject,$Expected.PSObject.BaseObject)}
    return ($Actual.PSObject.BaseObject -ceq $Expected.PSObject.BaseObject)
}
function BC-Bytes([int]$Kind=1){return ,([Convert]::FromHexString($script:fixtures[$Kind-1].FrameHex))}
function BC-Message([int]$Kind=1){
    $f=$script:fixtures[$Kind-1]
    BC-Object ([ordered]@{Version=[uint16]1;Kind=[byte]$f.Kind;Sender=[byte]$f.Sender;Receiver=[byte]$f.Receiver
        BindingDigest=[Convert]::FromHexString($f.BindingHex);SessionId=[Convert]::FromHexString($f.SessionHex)
        Sequence=[uint64]::Parse($f.Sequence,[Globalization.CultureInfo]::InvariantCulture)
        OriginTick=[long]::Parse($f.OriginTick);Frequency=[long]::Parse($f.Frequency);DeadlineTick=[long]::Parse($f.DeadlineTick);ObservedTick=[long]::Parse($f.ObservedTick)
        Payload=[Convert]::FromHexString($f.PayloadHex)})
}
function BC-Expected($Message){
    $stop=if($Message.OriginTick -eq 10 -and $Message.Frequency -eq 1001){[long]113623}elseif($Message.OriginTick -eq 10 -and $Message.Frequency -eq 1000){[long]113510}elseif($Message.OriginTick -eq 0 -and $Message.Frequency -eq 1000){[long]113500}else{throw 'No independent expected time fixture'}
    BC-Object ([ordered]@{Schema=[int]1;Sender=$Message.Sender;Receiver=$Message.Receiver;BindingDigest=$Message.BindingDigest;SessionId=$Message.SessionId
        NextSequence=$Message.Sequence;OriginTick=$Message.OriginTick;Frequency=$Message.Frequency;DeadlineTick=$Message.DeadlineTick;StopAtTick=$stop
        ArmSequence=[uint64]3;StopSequence=[uint64]2;LastNonHeartbeatSequence=[uint64]6})
}
function BC-Put($Owner,[string]$Name,$Value){$Owner.PSObject.Properties[$Name].Value=$Value}
function BC-UInt([byte[]]$Bytes,[int]$Offset,[int]$Width,[uint64]$Value){
    $hex=$Value.ToString(('X'+($Width*2)),[Globalization.CultureInfo]::InvariantCulture)
    for($j=0;$j -lt $Width;$j++){$Bytes[$Offset+$j]=[Convert]::ToByte($hex.Substring(($Width-1-$j)*2,2),16)}
}
function BC-Result([string]$Reason,$Bytes=$null,$Message=$null,$Next=$null){
    $good=$Reason -cin @('encoded','decoded','binding-matched')
    [pscustomobject][ordered]@{Schema=[int]1;Mode='p2b-codec-only';Accepted=[bool]$good;Reason=$Reason;Bytes=$Bytes;Message=$Message;NextSequence=$Next;SourceVerified=$false;LiveActionsPerformed=$false;RuntimeApproved=$false}
}
function BC-Encode($Message){$r=ConvertTo-P2BControlFrame $Message;BC-Need (BC-Equal $r (BC-Result encoded -Bytes (BC-Bytes ([int]$Message.Kind))))}
function BC-RejectEncode($Message){BC-Need (BC-Equal (ConvertTo-P2BControlFrame $Message) (BC-Result invalid-input))}
function BC-RejectDecode($Bytes,[string]$Reason='invalid-frame'){BC-Need (BC-Equal (ConvertFrom-P2BControlFrame $Bytes) (BC-Result $Reason))}
function BC-Bind($Message,$Expected,[string]$Reason='binding-matched',$Next=[uint64]8){
    $want=if($Reason -ceq 'binding-matched'){BC-Result $Reason -Next $Next}else{BC-Result $Reason}
    BC-Need (BC-Equal (Test-P2BControlBinding $Message $Expected) $want)
}
