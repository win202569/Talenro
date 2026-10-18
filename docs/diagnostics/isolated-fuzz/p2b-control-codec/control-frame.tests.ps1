param([ValidateSet('all','reserved-tail','sequence-equality')][string]$Group='all')
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$script:case='ApiRequired';$script:callback=0;$script:metadataSequence=0
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
function BC-Leaf($Owner,[string]$Field){
    $descriptor=$Owner.PSObject.Properties[$Field];BC-Need ($null -ne $descriptor -and $null -ne $descriptor.Value)
    $view=$Owner.$Field.PSObject
    foreach($p in $view.Properties){BC-Need ($p.MemberType -ne [Management.Automation.PSMemberTypes]::NoteProperty)}
    [pscustomobject]@{Owner=$Owner;Field=$Field;Descriptor=$descriptor;Base=$Owner.$Field.PSObject.BaseObject;Type=(BC-Type $descriptor.Value);View=$view;Members=$view.Properties}
}
function BC-Leaves($Value){
    foreach($p in $Value.PSObject.Properties){if($null -eq $p.Value){continue}
        if((BC-Type $p.Value) -eq [Management.Automation.PSCustomObject]){BC-Leaves $p.Value}else{BC-Leaf $Value $p.Name}
    }
}
function BC-Mutate($Selected,$Others){
    $script:metadataSequence++;$name='P2BCMetadata'+$script:metadataSequence
    foreach($other in $Others){BC-Need (-not (BC-Ref $Selected.Base $other.Base))}
    $payload=[object[]]@('initial')
    $Selected.Members.Add([Management.Automation.PSNoteProperty]::new($name,$payload));$held=$Selected.Members[$name]
    BC-Need ($null -ne $held -and (BC-Ref $held.Value $payload))
    $read=$Selected.Owner.($Selected.Field).PSObject.Properties[$name];$retained=$null -ne $read
    if($Selected.Type -ne [string]){BC-Need $retained}
    $held.Value[0]='changed'
    BC-Need ((BC-Ref $Selected.View.Properties[$name] $held) -and $held.Value[0] -ceq 'changed' -and $payload[0] -ceq 'changed')
    $after=$Selected.Owner.($Selected.Field).PSObject.Properties[$name]
    BC-Need (($null -ne $after) -eq $retained)
    if($retained){BC-Need ((BC-Ref $after.Value $payload) -and $after.Value[0] -ceq 'changed')}
    BC-Need ((BC-Ref $Selected.Owner.PSObject.Properties[$Selected.Field] $Selected.Descriptor) -and (BC-Type $Selected.Descriptor.Value) -eq $Selected.Type -and (BC-Ref $Selected.Owner.($Selected.Field).PSObject.BaseObject $Selected.Base))
    foreach($other in $Others){BC-Need ($null -eq $other.Members[$name] -and $null -eq $other.Owner.($other.Field).PSObject.Properties[$name])}
}
try{
    $identity=[Security.Principal.WindowsIdentity]::GetCurrent()
    $principal=[Security.Principal.WindowsPrincipal]::new($identity);$admin=$principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
    [Console]::Out.WriteLine('CONTEXT: '+$identity.Name+' admin='+$admin)
    [Console]::Out.WriteLine('HOST: '+$PSVersionTable.PSVersion.ToString()+' '+$PSVersionTable.PSEdition+' x64='+[Environment]::Is64BitProcess)
    BC-Need (-not $admin -and $PSVersionTable.PSVersion.ToString() -ceq '7.6.5' -and $PSVersionTable.PSEdition -ceq 'Core' -and [Environment]::Is64BitProcess)
    $module=Join-Path $PSScriptRoot 'control-frame.ps1';if(Test-Path -LiteralPath $module){. $module}
    BC-Case ApiRequired {
        foreach($fn in @('ConvertTo-P2BControlFrame','ConvertFrom-P2BControlFrame','Test-P2BControlBinding')){BC-Need ($null -ne (Get-Command $fn -CommandType Function -ErrorAction SilentlyContinue))}
    }
    $script:fixtures=@(Get-Content -LiteralPath (Join-Path $PSScriptRoot 'fixtures.json') -Raw|ConvertFrom-Json)

    BC-Case HelloLiteral {
        $m=BC-Message;$bytes=BC-Bytes
        BC-Need ($m.Sequence -eq [uint64]::Parse('72623859790382856') -and $bytes[64] -eq 8 -and $bytes[71] -eq 1 -and $bytes[88] -eq 202 -and $bytes[89] -eq 212)
        BC-Encode $m
        BC-Need (BC-Equal (ConvertFrom-P2BControlFrame $bytes) (BC-Result decoded -Message $m))
    }
    BC-Case EncodeAllKinds {foreach($kind in 1..12){BC-Encode (BC-Message $kind)}}
    BC-Case DecodeAllKinds {foreach($kind in 1..12){BC-Need (BC-Equal (ConvertFrom-P2BControlFrame (BC-Bytes $kind)) (BC-Result decoded -Message (BC-Message $kind)))}}
    BC-Case RoundTripSecondary {foreach($kind in 1..12){$m=BC-Message $kind;$a=ConvertTo-P2BControlFrame $m;$b=ConvertFrom-P2BControlFrame $a.Bytes;BC-Need ($a.Accepted -and $b.Accepted -and (BC-Equal $b.Message $m))}}
    BC-Case ExactMessageShape {
        BC-RejectEncode $null;BC-RejectEncode @{}
        foreach($field in @('Version','Payload')){
            $m=BC-Message;$m.PSObject.Properties.Remove($field);BC-RejectEncode $m
            $m=BC-Message;$v=$m.PSObject.Properties[$field].Value;$m.PSObject.Properties.Remove($field);$m.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new($field.ToLowerInvariant(),$v));BC-RejectEncode $m
        }
        $m=BC-Message;$m.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new('Extra',1));BC-RejectEncode $m
        $m=BC-Message 2
        foreach($mode in @('null','missing','extra','case')){
            $e=BC-Expected $m
            switch($mode){'null'{$e=$null} 'missing'{$e.PSObject.Properties.Remove('StopAtTick')} 'extra'{$e.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new('Extra',1))} 'case'{$e.PSObject.Properties.Remove('Schema');$e.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new('schema',[int]1))}}
            BC-Bind $m $e invalid-input
        }
    }
    BC-Case ExactTypes {
        foreach($field in @('Version','Kind','Sender','Receiver','Sequence','OriginTick','Frequency','DeadlineTick','ObservedTick')){
            foreach($value in @([int]1,'1',[double]1,$true,$null)){
                $m=BC-Message;BC-Put $m $field $value;BC-RejectEncode $m
            }
        }
        $m=BC-Message 2
        foreach($field in @('Schema','NextSequence','StopAtTick','ArmSequence','StopSequence','LastNonHeartbeatSequence')){
            $e=BC-Expected $m;BC-Put $e $field '1';BC-Bind $m $e invalid-input
        }
    }
    BC-Case PassiveDescriptors {
        $m=BC-Message;$m.PSObject.Properties.Remove('ObservedTick')
        $m.PSObject.Properties.Add([Management.Automation.PSScriptProperty]::new('ObservedTick',{$script:callback++;[long]11}))
        BC-Need ($m.ObservedTick -eq 11 -and $script:callback -eq 1);$script:callback=0;BC-RejectEncode $m;BC-Need ($script:callback -eq 0)
        $m=BC-Message;$m.PSObject.Properties.Remove('ObservedTick');$m.PSObject.Properties.Add([Management.Automation.PSAliasProperty]::new('ObservedTick','OriginTick'))
        BC-Need ($m.ObservedTick -eq 10 -and $m.PSObject.Properties['ObservedTick'].MemberType -eq [Management.Automation.PSMemberTypes]::AliasProperty);BC-RejectEncode $m
        foreach($name in @('GetType','ToString','Probe')){
            $m=BC-Message;$m.PSObject.Methods.Add([Management.Automation.PSScriptMethod]::new($name,{$script:callback++;'observable'}))
            BC-Need ($m.$name() -ceq 'observable' -and $script:callback -eq 1);$script:callback=0;BC-RejectEncode $m;BC-Need ($script:callback -eq 0)
        }
        $m=BC-Message 2;$e=BC-Expected $m;$e.PSObject.Properties.Remove('Schema')
        $e.PSObject.Properties.Add([Management.Automation.PSScriptProperty]::new('Schema',{$script:callback++;[int]1}))
        BC-Need ($e.Schema -eq 1 -and $script:callback -eq 1);$script:callback=0;BC-Bind $m $e invalid-input;BC-Need ($script:callback -eq 0)
    }
    BC-Case DecoratedScalars {
        foreach($field in @('Version','Kind','Sequence','OriginTick','BindingDigest','SessionId','Payload')){
            $m=BC-Message;$leaf=BC-Leaf $m $field;BC-Mutate $leaf @();BC-RejectEncode $m
        }
        $m=BC-Message 2;$e=BC-Expected $m;$leaf=BC-Leaf $e Schema;BC-Mutate $leaf @();BC-Bind $m $e invalid-input
        $holder=BC-Object ([ordered]@{Data=(BC-Bytes)});$leaf=BC-Leaf $holder Data;BC-Mutate $leaf @();BC-RejectDecode $holder.Data invalid-input
    }
    BC-Case ByteArrayShape {
        foreach($row in @(@('BindingDigest',31),@('BindingDigest',33),@('SessionId',15),@('SessionId',17))){
            $m=BC-Message;BC-Put $m $row[0] ([byte[]]::new($row[1]));BC-RejectEncode $m
        }
        foreach($value in @([object[]]@(0,1),([Array]::CreateInstance([byte],[int[]]@(2,2))),$null)){
            $m=BC-Message;BC-Put $m Payload $value;BC-RejectEncode $m
        }
        foreach($name in @('Length','Rank')){
            $m=BC-Message;$members=$m.Payload.PSObject.Properties
            $members.Add([Management.Automation.PSNoteProperty]::new($name,[int]99))
            BC-Need ($m.Payload.PSObject.Properties[$name].MemberType -eq [Management.Automation.PSMemberTypes]::NoteProperty -and $m.Payload.$name -eq 99);BC-RejectEncode $m
        }
    }
    BC-Case FrameLength {foreach($size in @(0,255,257,512)){BC-RejectDecode ([byte[]]::new($size))}}
    BC-Case HeaderMagicVersion {
        foreach($index in 0..3){$b=BC-Bytes;$b[$index]=255;BC-RejectDecode $b}
        foreach($value in @(0,2)){$b=BC-Bytes;BC-UInt $b 4 2 $value;BC-RejectDecode $b}
    }
    BC-Case ReservedZero {
        foreach($index in @(255)+(9..15)+(106..111)+(224..254)){$b=BC-Bytes;$b[$index]=1;BC-RejectDecode $b}
    }
    BC-Case PayloadLength {
        foreach($kind in 1..12){
            $m=BC-Message $kind;$len=$m.Payload.GetLength(0)
            foreach($n in @($len-1,$len+1,113)){$b=BC-Bytes $kind;BC-UInt $b 104 2 $n;BC-RejectDecode $b}
            $b=BC-Bytes $kind;$b[112+$len]=1;BC-RejectDecode $b
        }
    }
    BC-Case PayloadConstants {
        foreach($row in @(@(1,8),@(1,12),@(2,8),@(2,12),@(3,8),@(3,12),@(3,16),@(3,20),@(3,24),@(3,28))){
            $m=BC-Message $row[0];BC-UInt $m.Payload $row[1] 4 99;BC-RejectEncode $m
            $b=BC-Bytes $row[0];BC-UInt $b (112+$row[1]) 4 99;BC-RejectDecode $b
        }
    }
    BC-Case MalformedPayloadRange {
        foreach($kind in 1..12){foreach($offset in @(0,7)){$m=BC-Message $kind;$m.Payload[$offset]=1;BC-RejectEncode $m}}
        foreach($cause in @(0,14)){$m=BC-Message 12;BC-UInt $m.Payload 8 4 $cause;BC-RejectEncode $m}
        $m=BC-Message 12;BC-UInt $m.Payload 12 4 11;BC-RejectEncode $m
        $m=BC-Message 11;BC-UInt $m.Payload 12 4 11;BC-RejectEncode $m
        foreach($v in @(0,4)){$m=BC-Message 7;BC-UInt $m.Payload 16 4 $v;BC-RejectEncode $m}
        foreach($v in @(1,2,3)){
            $m=BC-Message 7;BC-UInt $m.Payload 16 4 $v;$r=ConvertTo-P2BControlFrame $m
            BC-Need ($r.Accepted -and $r.RuntimeApproved -eq $false -and (BC-Equal (ConvertFrom-P2BControlFrame $r.Bytes) (BC-Result decoded -Message $m)))
        }
        $m=BC-Message 9;BC-UInt $m.Payload 12 4 1;BC-RejectEncode $m
        $m=BC-Message 12;BC-UInt $m.Payload 20 4 1;BC-RejectEncode $m
    }
    BC-Case DirectionRoles {
        foreach($pair in @(@(1,2),@(2,1),@(2,3),@(3,2))){
            foreach($kind in @(1,2,8,12)){$m=BC-Message $kind;BC-Put $m Sender ([byte]$pair[0]);BC-Put $m Receiver ([byte]$pair[1]);BC-Need (ConvertTo-P2BControlFrame $m).Accepted}
        }
        foreach($kind in 1..12){
            $m=BC-Message $kind;BC-Put $m Sender ([byte]4);BC-RejectEncode $m
            $m=BC-Message $kind;BC-Put $m Receiver ([byte]0);BC-RejectEncode $m
        }
        foreach($kind in @(4,5,6,7)){$m=BC-Message $kind;BC-Put $m Sender ([byte]3);BC-Put $m Receiver ([byte]2);BC-RejectEncode $m}
        $m=BC-Message 10;BC-Put $m Sender ([byte]2);BC-Put $m Receiver ([byte]3);BC-RejectEncode $m
    }
    BC-Case ObservationSemantics {
        foreach($obs in 1..4){
            foreach($status in 1..3){
                $m=BC-Message 11;BC-Put $m Sender ([byte]2);BC-Put $m Receiver ([byte]1)
                BC-UInt $m.Payload 8 4 $obs;BC-UInt $m.Payload 12 4 3
                $subject=if($obs -eq 1){2}elseif($obs -eq 2){3}elseif($obs -eq 3){5}else{4}
                BC-UInt $m.Payload 16 4 $subject;BC-UInt $m.Payload 20 4 $status
                $v=if($obs -eq 4){if($status -eq 1){1}else{3}}else{0};BC-UInt $m.Payload 24 8 $v
                if($obs -eq 1 -and $status -ne 1){BC-RejectEncode $m}else{
                    $r=ConvertTo-P2BControlFrame $m;BC-Need ($r.Accepted -and (BC-Equal (ConvertFrom-P2BControlFrame $r.Bytes) (BC-Result decoded -Message $m)))
                }
            }
        }
        $m=BC-Message 11;BC-UInt $m.Payload 8 4 3;BC-UInt $m.Payload 16 4 5;BC-RejectEncode $m
        $m=BC-Message 11;BC-Put $m Sender ([byte]2);BC-Put $m Receiver ([byte]1);BC-UInt $m.Payload 8 4 2;BC-UInt $m.Payload 20 4 3;BC-UInt $m.Payload 24 8 1;BC-RejectEncode $m
        $m=BC-Message 11;BC-Put $m Sender ([byte]2);BC-Put $m Receiver ([byte]1);BC-UInt $m.Payload 8 4 3;BC-UInt $m.Payload 16 4 5;BC-UInt $m.Payload 24 8 3;BC-RejectEncode $m
        $m=BC-Message 11;BC-UInt $m.Payload 20 4 0;BC-RejectEncode $m
    }
    BC-Case SequenceExact {
        $m=BC-Message 2;$e=BC-Expected $m;BC-Bind $m $e
        foreach($seq in @([uint64]6,[uint64]8)){$a=BC-Message 2;BC-Put $a Sequence $seq;BC-Bind $a $e sequence-mismatch}
        $m=BC-Message 2;BC-Put $m Sequence ([uint64]::Parse('9223372036854775808'))
        $b=BC-Bytes 2;BC-UInt $b 64 8 ([uint64]::Parse('9223372036854775808'))
        BC-Need (BC-Equal (ConvertTo-P2BControlFrame $m) (BC-Result encoded -Bytes $b))
        BC-Need (BC-Equal (ConvertFrom-P2BControlFrame $b) (BC-Result decoded -Message $m))
        BC-Bind $m (BC-Expected $m) binding-matched ([uint64]::Parse('9223372036854775809'))
    }
    BC-Case SequenceExhaustion {
        $m=BC-Message 2;BC-Put $m Sequence ([uint64]::MaxValue)
        $b=BC-Bytes 2;BC-UInt $b 64 8 ([uint64]::MaxValue)
        BC-Need (BC-Equal (ConvertTo-P2BControlFrame $m) (BC-Result encoded -Bytes $b))
        BC-Need (BC-Equal (ConvertFrom-P2BControlFrame $b) (BC-Result decoded -Message $m))
        BC-Bind $m (BC-Expected $m) sequence-exhausted
        BC-Put $m Sequence ([uint64]0);BC-RejectEncode $m
    }
    BC-Case ArmReferences {
        $m=BC-Message 4;$e=BC-Expected $m;BC-Bind $m $e
        BC-UInt $m.Payload 8 8 6;BC-RejectEncode $m
        foreach($kind in @(5,6,7)){
            $m=BC-Message $kind;$e=BC-Expected $m;BC-Bind $m $e
            foreach($v in @([uint64]0,[uint64]4)){$e=BC-Expected $m;BC-Put $e ArmSequence $v;BC-Bind $m $e reference-mismatch}
        }
    }
    BC-Case StopReferences {
        $m=BC-Message 10;$e=BC-Expected $m;BC-Bind $m $e
        foreach($v in @([uint64]0,[uint64]3)){$e=BC-Expected $m;BC-Put $e StopSequence $v;BC-Bind $m $e reference-mismatch}
        BC-Encode (BC-Message 9)
    }
    BC-Case HeartbeatReference {
        $m=BC-Message 8;BC-Bind $m (BC-Expected $m)
        foreach($v in @(7,8)){$m=BC-Message 8;BC-UInt $m.Payload 8 8 $v;BC-RejectEncode $m}
        $m=BC-Message 8;$e=BC-Expected $m;BC-Put $e LastNonHeartbeatSequence ([uint64]5);BC-Bind $m $e reference-mismatch
    }
    BC-Case Timebase {
        BC-Encode (BC-Message 2)
        $m=BC-Message 2;BC-Put $m OriginTick ([long]10);BC-Put $m Frequency ([long]1001);BC-Put $m DeadlineTick ([long]120130);BC-Put $m ObservedTick ([long]14)
        $b=BC-Bytes 2;BC-UInt $b 72 8 10;BC-UInt $b 80 8 1001;BC-UInt $b 88 8 120130;BC-UInt $b 96 8 14
        BC-Need (BC-Equal (ConvertTo-P2BControlFrame $m) (BC-Result encoded -Bytes $b));BC-Bind $m (BC-Expected $m)
        foreach($row in @(@('OriginTick',[long]-1),@('Frequency',[long]999),@('ObservedTick',[long]-1),@('DeadlineTick',[long]120001),@('Frequency',[long]::MaxValue),@('OriginTick',[long]::MaxValue))){
            $m=BC-Message 2;BC-Put $m $row[0] $row[1];BC-RejectEncode $m
        }
        $m=BC-Message 12;BC-Put $m ObservedTick ([long]120001);BC-Need (ConvertTo-P2BControlFrame $m).Accepted
    }
    BC-Case ArmEquality {
        foreach($tick in @([long]113499,[long]113500,[long]113501)){
            $m=BC-Message 4;BC-Put $m ObservedTick $tick;$e=BC-Expected $m
            if($tick -eq 113499){BC-Bind $m $e}else{BC-Bind $m $e expired-declaration}
        }
        foreach($kind in @(9,12)){$m=BC-Message $kind;BC-Put $m ObservedTick ([long]120001);BC-Need (ConvertTo-P2BControlFrame $m).Accepted}
    }
    BC-Case BindingMismatch {
        $m=BC-Message 2
        foreach($field in @('BindingDigest','SessionId')){
            for($i=0;$i -lt $m.$field.GetLength(0);$i++){$e=BC-Expected $m;$e.$field[$i]=[byte]255;BC-Bind $m $e binding-mismatch}
        }
        $e=BC-Expected $m;BC-Put $e Sender ([byte]2);BC-Put $e Receiver ([byte]1);BC-Bind $m $e binding-mismatch
        $other=BC-Message;$e=BC-Expected $other;BC-Put $e NextSequence ([uint64]7);BC-Bind $m $e binding-mismatch
        $other=BC-Message 2;BC-Put $other OriginTick ([long]10);BC-Put $other Frequency ([long]1001);BC-Put $other DeadlineTick ([long]120130);BC-Put $other ObservedTick ([long]14)
        $e=BC-Expected $other;BC-Bind $m $e binding-mismatch
        BC-Need (BC-Equal $m (BC-Message 2))
    }
    BC-Case FragmentRefusal {
        $b=BC-Bytes;BC-RejectDecode ([byte[]]$b[0..127]);BC-RejectDecode ([byte[]]$b[128..255])
        BC-Need (BC-Equal (ConvertFrom-P2BControlFrame (BC-Bytes)) (BC-Result decoded -Message (BC-Message)))
    }
    BC-Case InputMutation {
        $m=BC-Message 2;$a=ConvertTo-P2BControlFrame $m;$b=ConvertTo-P2BControlFrame $m
        $inputs=@(BC-Leaves $m);$outputs=@((BC-Leaves $a);(BC-Leaves $b))
        foreach($leaf in $inputs){BC-Mutate $leaf $outputs}
        $m.Payload[8]=0;BC-Put $m OriginTick ([long]1)
        BC-Need ((BC-Equal $a (BC-Result encoded -Bytes (BC-Bytes 2))) -and (BC-Equal $b (BC-Result encoded -Bytes (BC-Bytes 2))))
        $bytes=BC-Bytes 2;$a=ConvertFrom-P2BControlFrame $bytes;$b=ConvertFrom-P2BControlFrame $bytes;$bytes[0]=0
        BC-Need ((BC-Equal $a (BC-Result decoded -Message (BC-Message 2))) -and (BC-Equal $b (BC-Result decoded -Message (BC-Message 2))))
        $m=BC-Message 2;$e=BC-Expected $m;$r=Test-P2BControlBinding $m $e;BC-Put $e NextSequence ([uint64]99);$e.BindingDigest[0]=255
        BC-Need (BC-Equal $r (BC-Result binding-matched -Next ([uint64]8)))
        BC-Need (BC-Equal (ConvertTo-P2BControlFrame (BC-Message 2)) (BC-Result encoded -Bytes (BC-Bytes 2)))
    }
    BC-Case OutputOwnership {
        foreach($mode in @('encode','decode','binding')){
            $m=BC-Message 2;$e=BC-Expected $m
            switch($mode){'encode'{$a=ConvertTo-P2BControlFrame $m;$b=ConvertTo-P2BControlFrame $m} 'decode'{$a=ConvertFrom-P2BControlFrame (BC-Bytes 2);$b=ConvertFrom-P2BControlFrame (BC-Bytes 2)} 'binding'{$a=Test-P2BControlBinding $m $e;$b=Test-P2BControlBinding $m $e}}
            $la=@(BC-Leaves $a);$lb=@(BC-Leaves $b);$li=@((BC-Leaves $m);(BC-Leaves $e))
            BC-Need (-not (BC-Ref $a $b))
            foreach($leaf in $la){BC-Mutate $leaf @($li+$lb+@($la|Where-Object{-not (BC-Ref $_ $leaf)}))}
            BC-Need (BC-Equal $a $b)
            if($mode -ceq 'encode'){$a.Bytes[0]=0;BC-Need ($b.Bytes[0] -eq 80)}
            if($mode -ceq 'decode'){foreach($field in @('BindingDigest','SessionId','Payload')){$old=$b.Message.$field[0];$a.Message.$field[0]=255;BC-Need ($b.Message.$field[0] -eq $old)}}
        }
    }
    BC-Case SequentialIsolation {
        $a=ConvertFrom-P2BControlFrame (BC-Bytes 2);$bad=ConvertFrom-P2BControlFrame ([byte[]]::new(255));$b=ConvertFrom-P2BControlFrame (BC-Bytes 2)
        $la=@(BC-Leaves $a);$lb=@(BC-Leaves $b);$lf=@(BC-Leaves $bad)
        foreach($leaf in $lb){BC-Mutate $leaf @($la+$lf+@($lb|Where-Object{-not (BC-Ref $_ $leaf)}))}
        foreach($leaf in $la){BC-Mutate $leaf @($lb+$lf+@($la|Where-Object{-not (BC-Ref $_ $leaf)}))}
        BC-Need ((BC-Equal $a (BC-Result decoded -Message (BC-Message 2))) -and (BC-Equal $b (BC-Result decoded -Message (BC-Message 2))) -and (BC-Equal $bad (BC-Result invalid-frame)))
        $a.Message.Payload[8]=0;BC-Need ($b.Message.Payload[8] -eq 7)
        BC-Encode (BC-Message 2);$m=BC-Message 2;BC-Bind $m (BC-Expected $m)
    }
    BC-Case ConservativeFlags {
        $m=BC-Message 2
        $results=@((ConvertTo-P2BControlFrame $m),(ConvertFrom-P2BControlFrame (BC-Bytes 2)),(Test-P2BControlBinding $m (BC-Expected $m)),(ConvertTo-P2BControlFrame $null),(ConvertFrom-P2BControlFrame ([byte[]]::new(255))))
        foreach($r in $results){
            BC-Need (@($r.PSObject.Properties).Count -eq 10 -and (BC-Type $r.Schema) -eq [int] -and $r.Schema -eq 1 -and $r.Mode -ceq 'p2b-codec-only')
            foreach($field in @('Accepted','SourceVerified','LiveActionsPerformed','RuntimeApproved')){BC-Need ((BC-Type $r.$field) -eq [bool])}
            BC-Need (-not $r.SourceVerified -and -not $r.LiveActionsPerformed -and -not $r.RuntimeApproved)
            $nonnull=0;foreach($field in @('Bytes','Message','NextSequence')){if($null -ne $r.$field){$nonnull++}}
            BC-Need (($r.Accepted -and $nonnull -eq 1) -or (-not $r.Accepted -and $nonnull -eq 0))
        }
    }
    BC-Case Complete {BC-Need ($script:callback -eq 0)}
    [Console]::Out.WriteLine('LIMIT: protocol declarations only; no identity, capability or runtime approval')
    exit 0
}catch{[Console]::Out.WriteLine('ASSERTION FAILED: P2BCodec'+$script:case);exit 1}
