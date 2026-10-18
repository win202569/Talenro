# Pure declared protocol data only. Definitions have no load-time actions.
# Helpers are internal conventions, not an isolation or security boundary.
function New-P2BCMetadata {
    [pscustomobject]@{
        TypeMethod=[object].GetMethod('GetType',[type[]]@())
        InvokeMethod=[Reflection.MethodBase].GetMethod('Invoke',[type[]]@([object],[object[]]))
        ReferenceMethod=[object].GetMethod('ReferenceEquals',[type[]]@([object],[object]))
        MembersProperty=[Management.Automation.PSMemberSet].GetProperty('Members')
        RawMembersField=[Management.Automation.PSMemberSet].GetField('internalMembers',[Reflection.BindingFlags]'NonPublic,Instance')
        ShimType=[Management.Automation.PSObject].Assembly.GetType('System.Management.Automation.PSInternalMemberSet',$true)
        NameField=[Management.Automation.PSMemberInfo].GetField('name',[Reflection.BindingFlags]'NonPublic,Instance')
        ReservedField=[Management.Automation.PSMemberInfo].GetField('<IsReservedMember>k__BackingField',[Reflection.BindingFlags]'NonPublic,Instance')
        HiddenField=[Management.Automation.PSMemberInfo].GetField('<IsHidden>k__BackingField',[Reflection.BindingFlags]'NonPublic,Instance')
        InstanceField=[Management.Automation.PSMemberInfo].GetField('<IsInstance>k__BackingField',[Reflection.BindingFlags]'NonPublic,Instance')
        OwnerField=[Management.Automation.PSMemberInfo].GetField('instance',[Reflection.BindingFlags]'NonPublic,Instance,DeclaredOnly')
        ShimOwnerField=[Management.Automation.PSObject].Assembly.GetType('System.Management.Automation.PSInternalMemberSet',$true).GetField('_psObject',[Reflection.BindingFlags]'NonPublic,Instance,DeclaredOnly')
        ConstructorOwnerField=[Management.Automation.PSMemberSet].GetField('_constructorPSObject',[Reflection.BindingFlags]'NonPublic,Instance,DeclaredOnly')
        GetterField=[Management.Automation.PSCodeProperty].GetField('<GetterCodeReference>k__BackingField',[Reflection.BindingFlags]'NonPublic,Instance,DeclaredOnly')
        SetterField=[Management.Automation.PSCodeProperty].GetField('<SetterCodeReference>k__BackingField',[Reflection.BindingFlags]'NonPublic,Instance,DeclaredOnly')
        SDKGetterField=[Management.Automation.PSObject].Assembly.GetType('System.Management.Automation.Language.CachedReflectionInfo',$true).GetField('ReservedNameMembers_PSTypeNames',[Reflection.BindingFlags]'NonPublic,Static,DeclaredOnly')
        FieldReadMethod=[Reflection.FieldInfo].GetMethod('GetValue',[type[]]@([object]))
        WrapperBaseField=[Management.Automation.PSObject].GetField('_immediateBaseObject',[Reflection.BindingFlags]'NonPublic,Instance,DeclaredOnly')
        WrapperMembersField=[Management.Automation.PSObject].GetField('_instanceMembers',[Reflection.BindingFlags]'NonPublic,Instance,DeclaredOnly')
        WrapperFlagsField=[Management.Automation.PSObject].GetField('_flags',[Reflection.BindingFlags]'NonPublic,Instance,DeclaredOnly')
        ArrayLength=[Array].GetMethod('GetLength',[type[]]@([int]))
        ArrayClone=[Array].GetMethod('Clone',[type[]]@())
        StringChars=[string].GetMethod('ToCharArray',[type[]]@())
        StringConstructor=[string].GetConstructor([type[]]@([char[]]))
    }
}
function Test-P2BCReference($Left,$Right,[object]$Metadata){
    # Keep actual PSObject owners in an argument array: direct CLR binding can
    # unwrap them to PSCustomObject bases and erase their distinct identities.
    $arguments=[object[]]@($Left,$Right)
    return [bool]$Metadata.ReferenceMethod.Invoke($null,$arguments)
}
function Get-P2BCRawType($Value,[object]$Metadata){
    $arguments=[object[]]::new(2);$arguments[0]=$Value;$arguments[1]=$null
    return $Metadata.InvokeMethod.Invoke($Metadata.TypeMethod,$arguments)
}
function Read-P2BCField([Reflection.FieldInfo]$Field,$Target,[object]$Metadata){
    if($null -eq $Field){throw 'P2BC_INVALID'}
    $arguments=[object[]]::new(1);$arguments[0]=$Target
    $result=[object[]]::new(1);$result[0]=$Metadata.FieldReadMethod.Invoke($Field,$arguments)
    return ,$result
}
function Get-P2BCStorage($Wrapper,[object]$Metadata){
    # Actual PSObject references must never cross a coercing CLR argument binder.
    if($null -eq $Wrapper -or (Get-P2BCRawType $Wrapper $Metadata) -ne [Management.Automation.PSObject]){throw 'P2BC_INVALID'}
    foreach($field in @($Metadata.WrapperBaseField,$Metadata.WrapperMembersField,$Metadata.WrapperFlagsField)){
        if($null -eq $field -or $field.DeclaringType -ne [Management.Automation.PSObject]){throw 'P2BC_INVALID'}
    }
    $flags=[int]((Read-P2BCField $Metadata.WrapperFlagsField $Wrapper $Metadata)[0])
    $current=$Wrapper
    for($depth=0;$depth -lt 8;$depth++){
        $next=(Read-P2BCField $Metadata.WrapperBaseField $current $Metadata)[0]
        if($null -eq $next){throw 'P2BC_INVALID'}
        $type=Get-P2BCRawType $next $Metadata
        if($type -ne [Management.Automation.PSObject]){break}
        $current=$next
    }
    if($type -eq [Management.Automation.PSObject]){throw 'P2BC_INVALID'}
    $result=[object[]]::new(3)
    # SDK uses the innermost wrapper for string/custom bases or outer local flag.
    # Branch assignments intentionally avoid pipeline-unwrapping a PSObject key.
    if($type -eq [string] -or $type -eq [Management.Automation.PSCustomObject] -or ($flags -band 16) -ne 0){$result[0]=$current}
    else{$result[0]=$next}
    $result[1]=(Read-P2BCField $Metadata.WrapperMembersField $Wrapper $Metadata)[0]
    $result[2]=$next
    if($null -ne $result[1] -and (Get-P2BCRawType $result[1] $Metadata) -ne $Metadata.RawMembersField.FieldType){throw 'P2BC_INVALID'}
    return ,$result
}
function Assert-P2BCPlain([object]$Value,[type]$Required,[object]$Metadata){
    if($null -eq $Value){throw 'P2BC_INVALID'}
    $inputType=Get-P2BCRawType $Value $Metadata
    $wrapper=$null;$shimBase=$null
    if($inputType -eq [Management.Automation.PSObject]){$wrapper=$Value}
    else{
        if($inputType -ne $Required){throw 'P2BC_INVALID'}
        # Acquire ONE shim only. Never reacquire psextended on a scalar/string:
        # separate adaptation can create a different actual wrapper/table.
        $shim=$Value.PSObject;$shimBase=$shim.PSObject.BaseObject
        if($Metadata.TypeMethod.Invoke($shimBase,$null) -ne $Metadata.ShimType){throw 'P2BC_INVALID'}
        $wrapper=(Read-P2BCField $Metadata.ShimOwnerField $shimBase $Metadata)[0]
    }
    $storage=Get-P2BCStorage $wrapper $Metadata
    if((Get-P2BCRawType $storage[2] $Metadata) -ne $Required){throw 'P2BC_INVALID'}
    if($inputType -ne [Management.Automation.PSObject] -and -not (Test-P2BCReference $storage[2] $Value $Metadata)){throw 'P2BC_INVALID'}
    $raw=$storage[1]
    # A clean explicit wrapper may have no instance table yet; no views are
    # generated by the validator. Partial legitimate reserved-view history is OK.
    if($null -eq $raw){return}
    $seen=[Collections.Generic.HashSet[string]]::new([StringComparer]::Ordinal)
    foreach($member in $raw){
        $base=$member.PSObject.BaseObject;$type=$Metadata.TypeMethod.Invoke($base,$null)
        $name=$Metadata.NameField.GetValue($base)
        if($null -eq $name -or -not $seen.Add($name)){throw 'P2BC_INVALID'}
        if($name -cnotin @('psobject','psextended','psbase','psadapted','pstypenames')){throw 'P2BC_INVALID'}
        $wantType=if($name -ceq 'psextended'){[Management.Automation.PSMemberSet]}elseif($name -ceq 'pstypenames'){[Management.Automation.PSCodeProperty]}else{$Metadata.ShimType}
        if($type -ne $wantType -or -not $Metadata.ReservedField.GetValue($base) -or -not $Metadata.HiddenField.GetValue($base) -or -not $Metadata.InstanceField.GetValue($base)){throw 'P2BC_INVALID'}
        $owner=(Read-P2BCField $Metadata.OwnerField $base $Metadata)[0]
        $ownStorage=Get-P2BCStorage $owner $Metadata
        if(-not (Test-P2BCReference $ownStorage[0] $storage[0] $Metadata) -or -not (Test-P2BCReference $ownStorage[1] $raw $Metadata)){throw 'P2BC_INVALID'}
        if($name -ceq 'psobject' -and $null -ne $shimBase -and -not (Test-P2BCReference $base $shimBase $Metadata)){throw 'P2BC_INVALID'}
        if($type -eq $Metadata.ShimType){
            if(-not (Test-P2BCReference ((Read-P2BCField $Metadata.ShimOwnerField $base $Metadata)[0]) $owner $Metadata)){throw 'P2BC_INVALID'}
        }elseif($type -eq [Management.Automation.PSMemberSet]){
            if(-not (Test-P2BCReference ((Read-P2BCField $Metadata.ConstructorOwnerField $base $Metadata)[0]) $owner $Metadata) -or -not (Test-P2BCReference $Metadata.RawMembersField.GetValue($base) $raw $Metadata)){throw 'P2BC_INVALID'}
        }else{
            if(-not $Metadata.SDKGetterField.IsStatic -or -not $Metadata.SDKGetterField.IsInitOnly){throw 'P2BC_INVALID'}
            $getter=$Metadata.SDKGetterField.GetValue($null)
            if($null -eq $getter -or -not (Test-P2BCReference $Metadata.GetterField.GetValue($base) $getter $Metadata) -or $null -ne $Metadata.SetterField.GetValue($base)){throw 'P2BC_INVALID'}
        }
    }
}
function Copy-P2BCScalar([object]$Value,[object]$Metadata){
    if($null -eq $Value){return $null}
    $base=$Value.PSObject.BaseObject;$type=$Metadata.TypeMethod.Invoke($base,$null)
    if($type -eq [byte[]]){$fresh=$Metadata.ArrayClone.Invoke($base,$null)}
    elseif($type -eq [string]){
        $chars=$Metadata.StringChars.Invoke($base,$null)
        $argsLocal=[object[]]::new(1);$argsLocal[0]=$chars;$fresh=$Metadata.StringConstructor.Invoke($argsLocal)
    }elseif($type -in @([byte],[uint16],[uint32],[uint64],[int],[long],[bool])){
        $storage=[Array]::CreateInstance($type,1);$storage.SetValue($base,0);$fresh=$storage.GetValue(0)
    }else{throw 'Unexpected scalar'}
    if($Metadata.TypeMethod.Invoke($fresh.PSObject.BaseObject,$null) -ne $type -or [object]::ReferenceEquals($base,$fresh.PSObject.BaseObject)){throw 'Scalar ownership'}
    try{Assert-P2BCPlain $fresh $type $Metadata}
    catch{throw [InvalidOperationException]::new('New scalar decorated',$_.Exception)}
    return ,$fresh
}
function Copy-P2BCMessage([object]$Message,[object]$Metadata){
    $owned=[pscustomobject][ordered]@{}
    foreach($name in @('Version','Kind','Sender','Receiver','BindingDigest','SessionId','Sequence','OriginTick','Frequency','DeadlineTick','ObservedTick','Payload')){
        $scalar=Copy-P2BCScalar $Message.PSObject.Properties[$name].Value $Metadata
        $owned.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new($name,$scalar))
    }
    return $owned
}
function New-P2BCResult([bool]$Accepted,[string]$Reason,[object]$Bytes,[object]$Message,[object]$NextSequence,[object]$Metadata){
    $values=[ordered]@{Schema=[int]1;Mode='p2b-codec-only';Accepted=$Accepted;Reason=$Reason;Bytes=$Bytes;Message=$Message;NextSequence=$NextSequence;SourceVerified=$false;LiveActionsPerformed=$false;RuntimeApproved=$false}
    $owned=[pscustomobject][ordered]@{}
    foreach($name in $values.Keys){
        $scalar=if($name -ceq 'Message' -and $null -ne $Message){Copy-P2BCMessage $Message $Metadata}else{Copy-P2BCScalar $values[$name] $Metadata}
        $owned.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new($name,$scalar))
    }
    return $owned
}
function Get-P2BCByteSnapshot([object]$Value,[int]$Minimum,[int]$Maximum,[object]$Metadata){
    Assert-P2BCPlain $Value ([byte[]]) $Metadata
    $length=[int]$Metadata.ArrayLength.Invoke($Value.PSObject.BaseObject,[object[]]@([int]0))
    if($length -lt $Minimum -or $length -gt $Maximum){throw 'P2BC_INVALID'}
    return ,(Copy-P2BCScalar $Value $Metadata)
}
function Get-P2BCSnapshot([object]$Value,[bool]$Expected,[object]$Metadata){
    if($null -eq $Value -or $Metadata.TypeMethod.Invoke($Value.PSObject.BaseObject,$null) -ne [Management.Automation.PSCustomObject]){throw 'P2BC_INVALID'}
    $shape=if($Expected){[ordered]@{Schema=[int];Sender=[byte];Receiver=[byte];BindingDigest=[byte[]];SessionId=[byte[]];NextSequence=[uint64];OriginTick=[long];Frequency=[long];DeadlineTick=[long];StopAtTick=[long];ArmSequence=[uint64];StopSequence=[uint64];LastNonHeartbeatSequence=[uint64]}}
    else{[ordered]@{Version=[uint16];Kind=[byte];Sender=[byte];Receiver=[byte];BindingDigest=[byte[]];SessionId=[byte[]];Sequence=[uint64];OriginTick=[long];Frequency=[long];DeadlineTick=[long];ObservedTick=[long];Payload=[byte[]]}}
    $names=[Collections.Generic.HashSet[string]]::new([StringComparer]::Ordinal)
    foreach($name in $shape.Keys){$null=$names.Add($name)}
    # Public Properties/Members views may invoke virtual descriptor members.
    # Use the fixed SDK field only after proving both receiver identities.
    $shim=$Value.PSObject;$extended=$Value.psextended
    $shimBase=$shim.PSObject.BaseObject;$extendedBase=$extended.PSObject.BaseObject
    if($Metadata.TypeMethod.Invoke($shimBase,$null) -ne $Metadata.ShimType -or
       $Metadata.TypeMethod.Invoke($extendedBase,$null) -ne [Management.Automation.PSMemberSet]){throw 'P2BC_INVALID'}
    foreach($field in @($Metadata.OwnerField,$Metadata.ShimOwnerField,$Metadata.ConstructorOwnerField,$Metadata.GetterField,$Metadata.SetterField,$Metadata.SDKGetterField)){
        if($null -eq $field){throw 'P2BC_INVALID'}
    }
    $owner=$Metadata.ShimOwnerField.GetValue($shimBase)
    $ownerArguments=[object[]]@($owner,$null)
    if($null -eq $owner -or $Metadata.InvokeMethod.Invoke($Metadata.TypeMethod,$ownerArguments) -ne [Management.Automation.PSObject] -or
       -not (Test-P2BCReference $Metadata.ConstructorOwnerField.GetValue($extendedBase) $owner $Metadata)){throw 'P2BC_INVALID'}
    if(-not $Metadata.SDKGetterField.IsStatic -or -not $Metadata.SDKGetterField.IsInitOnly){throw 'P2BC_INVALID'}
    $sdkGetter=$Metadata.SDKGetterField.GetValue($null)
    if($null -eq $sdkGetter){throw 'P2BC_INVALID'}
    $rawMembers=$Metadata.RawMembersField.GetValue($extendedBase)
    if($Metadata.TypeMethod.Invoke($rawMembers.PSObject.BaseObject,$null) -ne $Metadata.RawMembersField.FieldType){throw 'P2BC_INVALID'}
    $properties=[Collections.Generic.Dictionary[string,object]]::new([StringComparer]::Ordinal)
    $seen=[Collections.Generic.HashSet[string]]::new([StringComparer]::Ordinal)
    $sdkCount=0
    foreach($member in $rawMembers){
        $base=$member.PSObject.BaseObject;$type=$Metadata.TypeMethod.Invoke($base,$null)
        $name=$Metadata.NameField.GetValue($base)
        $reserved=$Metadata.ReservedField.GetValue($base);$hidden=$Metadata.HiddenField.GetValue($base)
        $instance=$Metadata.InstanceField.GetValue($base)
        if($null -eq $name -or -not $seen.Add($name)){throw 'P2BC_INVALID'}
        if($name -cin @('psobject','psextended','psbase','psadapted','pstypenames')){
            # Structural SDK compatibility, not unforgeable factory provenance.
            $wantType=if($name -ceq 'psextended'){[Management.Automation.PSMemberSet]}elseif($name -ceq 'pstypenames'){[Management.Automation.PSCodeProperty]}else{$Metadata.ShimType}
            if($type -ne $wantType -or -not $reserved -or -not $hidden -or -not $instance -or
               -not (Test-P2BCReference $Metadata.OwnerField.GetValue($base) $owner $Metadata)){throw 'P2BC_INVALID'}
            if($name -ceq 'psobject' -and -not (Test-P2BCReference $base $shimBase $Metadata)){throw 'P2BC_INVALID'}
            if($name -ceq 'psextended' -and -not (Test-P2BCReference $base $extendedBase $Metadata)){throw 'P2BC_INVALID'}
            if($type -eq $Metadata.ShimType -and -not (Test-P2BCReference $Metadata.ShimOwnerField.GetValue($base) $owner $Metadata)){throw 'P2BC_INVALID'}
            if($name -ceq 'pstypenames'){
                if(-not (Test-P2BCReference $Metadata.GetterField.GetValue($base) $sdkGetter $Metadata) -or
                   $null -ne $Metadata.SetterField.GetValue($base)){throw 'P2BC_INVALID'}
            }
            $sdkCount++
        }else{
            if($type -ne [Management.Automation.PSNoteProperty] -or -not $names.Contains($name) -or
               $reserved -or $hidden -or -not $instance){throw 'P2BC_INVALID'}
            $properties.Add($name,$member)
        }
    }
    if($sdkCount -lt 2 -or $sdkCount -gt 5 -or -not $seen.Contains('psobject') -or
       -not $seen.Contains('psextended') -or $properties.Count -ne $shape.Count){throw 'P2BC_INVALID'}
    # Finish ALL exact-type/member checks before copying or interpreting any field.
    foreach($name in $shape.Keys){Assert-P2BCPlain $properties[$name].Value $shape[$name] $Metadata}
    $snapshot=[pscustomobject][ordered]@{}
    foreach($name in $shape.Keys){
        $raw=$properties[$name].Value
        if($shape[$name] -eq [byte[]]){
            $min=if($name -ceq 'BindingDigest'){32}elseif($name -ceq 'SessionId'){16}else{0}
            $max=if($name -ceq 'Payload'){112}else{$min}
            $scalar=Get-P2BCByteSnapshot $raw $min $max $Metadata
        }else{$scalar=Copy-P2BCScalar $raw $Metadata}
        $snapshot.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new($name,$scalar))
    }
    return $snapshot
}
function Read-P2BCUInt([byte[]]$Bytes,[int]$Offset,[int]$Width){
    if($Width -notin @(2,4,8) -or $Offset -lt 0 -or $Offset+$Width -gt $Bytes.GetLength(0)){throw 'P2BC_INVALID'}
    $value=[Numerics.BigInteger]::Zero
    for($i=$Width-1;$i -ge 0;$i--){$value=[Numerics.BigInteger]::Add([Numerics.BigInteger]::Multiply($value,[Numerics.BigInteger]256),[Numerics.BigInteger]$Bytes[$Offset+$i])}
    if($value -lt [Numerics.BigInteger]::Zero -or $value -gt [Numerics.BigInteger]([uint64]::MaxValue)){throw 'P2BC_INVALID'}
    return [uint64]$value
}
function Write-P2BCUInt([byte[]]$Bytes,[int]$Offset,[int]$Width,[uint64]$Value){
    if($Width -notin @(2,4,8) -or $Offset -lt 0 -or $Offset+$Width -gt $Bytes.GetLength(0)){throw 'P2BC_INVALID'}
    $n=[Numerics.BigInteger]$Value;$limit=[Numerics.BigInteger]::Pow([Numerics.BigInteger]256,$Width)
    if($n -ge $limit){throw 'P2BC_INVALID'}
    for($i=0;$i -lt $Width;$i++){$Bytes[$Offset+$i]=[byte][Numerics.BigInteger]::Remainder($n,[Numerics.BigInteger]256);$n=[Numerics.BigInteger]::Divide($n,[Numerics.BigInteger]256)}
}
function Assert-P2BCZeroBytes([byte[]]$Bytes,[int]$Start,[int]$Count){
    if($Start -lt 0 -or $Count -lt 0 -or $Start -gt $Bytes.GetLength(0) -or $Count -gt $Bytes.GetLength(0)-$Start){throw 'P2BC_INVALID'}
    for($i=0;$i -lt $Count;$i++){if($Bytes[$Start+$i] -ne 0){throw 'P2BC_INVALID'}}
}
function Get-P2BCDelta([long]$Frequency,[int]$Milliseconds,[bool]$Ceiling){
    $f=[Numerics.BigInteger]$Frequency;$ms=[Numerics.BigInteger]$Milliseconds;$thousand=[Numerics.BigInteger]1000
    $q=[Numerics.BigInteger]::Divide($f,$thousand);$r=[Numerics.BigInteger]::Remainder($f,$thousand)
    $whole=[Numerics.BigInteger]::Multiply($q,$ms);$fraction=[Numerics.BigInteger]::Multiply($r,$ms);$max=[Numerics.BigInteger]([long]::MaxValue)
    if($whole -lt [Numerics.BigInteger]::Zero -or $whole -gt $max -or $fraction -lt [Numerics.BigInteger]::Zero -or $fraction -gt $max){throw 'P2BC_INVALID'}
    $n=[Numerics.BigInteger]::Add($whole,[Numerics.BigInteger]::Divide($fraction,$thousand))
    if($Ceiling -and [Numerics.BigInteger]::Remainder($fraction,$thousand) -ne [Numerics.BigInteger]::Zero){$n=[Numerics.BigInteger]::Add($n,[Numerics.BigInteger]::One)}
    if($n -lt [Numerics.BigInteger]::Zero -or $n -gt $max){throw 'P2BC_INVALID'}
    return [long]$n
}
function Get-P2BCDeclaredBounds([long]$Origin,[long]$Frequency,[long]$Deadline){
    if($Origin -lt 0 -or $Frequency -lt 1000){throw 'P2BC_INVALID'}
    $max=[Numerics.BigInteger]([long]::MaxValue)
    $d=[Numerics.BigInteger]::Add([Numerics.BigInteger]$Origin,[Numerics.BigInteger](Get-P2BCDelta $Frequency 120000 $false))
    $exec=[Numerics.BigInteger]::Add([Numerics.BigInteger]$Origin,[Numerics.BigInteger](Get-P2BCDelta $Frequency 113500 $false))
    $reserve=[Numerics.BigInteger]::Subtract($d,[Numerics.BigInteger](Get-P2BCDelta $Frequency 6500 $true))
    foreach($n in @($d,$exec,$reserve)){if($n -lt [Numerics.BigInteger]::Zero -or $n -gt $max){throw 'P2BC_INVALID'}}
    if($d -ne [Numerics.BigInteger]$Deadline){throw 'P2BC_INVALID'}
    [pscustomobject]@{Deadline=[long]$d;StopAt=[Math]::Min([long]$exec,[long]$reserve)}
}
function Assert-P2BCDirection([object]$Message){
    $pair=([int]$Message.Sender).ToString([Globalization.CultureInfo]::InvariantCulture)+':'+([int]$Message.Receiver).ToString([Globalization.CultureInfo]::InvariantCulture)
    $allowed=switch([int]$Message.Kind){
        {$_ -in @(1,2,8,12)}{@('1:2','2:1','2:3','3:2');break}
        3{@('3:2');break} 4{@('2:1');break}
        {$_ -in @(5,6,7)}{@('1:2');break}
        9{@('1:2','2:1','2:3');break} 10{@('1:2','2:1','3:2');break}
        11{@('3:2','2:1');break} default{throw 'P2BC_INVALID'}
    }
    if($pair -cnotin $allowed){throw 'P2BC_INVALID'}
}
function Assert-P2BCPayload([object]$Message,[object]$Bounds){
    $p=$Message.Payload;$len=$p.GetLength(0);$lengths=@(0,16,16,32,24,16,16,24,16,24,16,32,24)
    if($len -ne $lengths[[int]$Message.Kind]){throw 'P2BC_INVALID'}
    Assert-P2BCZeroBytes $p 0 8
    switch([int]$Message.Kind){
        1{if((Read-P2BCUInt $p 8 4) -ne 1 -or (Read-P2BCUInt $p 12 4) -ne 256){throw 'P2BC_INVALID'}}
        2{if((Read-P2BCUInt $p 8 4) -ne 7){throw 'P2BC_INVALID'};Assert-P2BCZeroBytes $p 12 4}
        3{$want=@(7,1048576,1048576,2097152,4096,65536);for($i=0;$i -lt 6;$i++){if((Read-P2BCUInt $p (8+4*$i) 4) -ne $want[$i]){throw 'P2BC_INVALID'}}}
        4{if((Read-P2BCUInt $p 8 8) -ne $Message.Sequence -or (Read-P2BCUInt $p 16 8) -ne [uint64]$Bounds.StopAt){throw 'P2BC_INVALID'}}
        {$_ -in @(5,6,7)}{
            if((Read-P2BCUInt $p 8 8) -eq 0){throw 'P2BC_INVALID'}
            if($_ -eq 7 -and (Read-P2BCUInt $p 16 4) -notin @(1,2,3)){throw 'P2BC_INVALID'}
        }
        8{if((Read-P2BCUInt $p 8 8) -ge $Message.Sequence){throw 'P2BC_INVALID'}}
        9{if((Read-P2BCUInt $p 8 4) -notin (1..13)){throw 'P2BC_INVALID'};Assert-P2BCZeroBytes $p 12 4}
        10{if((Read-P2BCUInt $p 8 8) -eq 0){throw 'P2BC_INVALID'}}
        11{
            $kind=Read-P2BCUInt $p 8 4;$state=Read-P2BCUInt $p 12 4;$subject=Read-P2BCUInt $p 16 4;$status=Read-P2BCUInt $p 20 4;$value=Read-P2BCUInt $p 24 8
            if($kind -notin (1..4) -or $state -notin (1..10) -or $status -notin (1..3)){throw 'P2BC_INVALID'}
            if($Message.Sender -eq 3 -and $kind -ne 1){throw 'P2BC_INVALID'}
            switch([int]$kind){
                1{if($subject -ne $Message.Sender -or $status -ne 1 -or $value -ne 0){throw 'P2BC_INVALID'}}
                2{if($subject -notin @(3,4) -or ($status -eq 1 -and $value -gt [uint64][uint32]::MaxValue) -or ($status -ne 1 -and $value -ne 0)){throw 'P2BC_INVALID'}}
                3{if($subject -ne 5 -or ($status -eq 1 -and $value -gt 2) -or ($status -ne 1 -and $value -ne 0)){throw 'P2BC_INVALID'}}
                4{if($subject -ne 4 -or ($status -eq 1 -and $value -notin @(1,2)) -or ($status -ne 1 -and $value -ne 3)){throw 'P2BC_INVALID'}}
            }
        }
        12{
            if((Read-P2BCUInt $p 8 4) -notin (1..13) -or (Read-P2BCUInt $p 12 4) -notin (0..10)){throw 'P2BC_INVALID'}
            Assert-P2BCZeroBytes $p 20 4
        }
    }
}
function Assert-P2BCMessage([object]$Message){
    if($Message.Version -ne 1 -or $Message.Kind -lt 1 -or $Message.Kind -gt 12 -or $Message.Sequence -eq 0 -or $Message.ObservedTick -lt $Message.OriginTick){throw 'P2BC_INVALID'}
    $bounds=Get-P2BCDeclaredBounds $Message.OriginTick $Message.Frequency $Message.DeadlineTick
    Assert-P2BCDirection $Message;Assert-P2BCPayload $Message $bounds
}
function Assert-P2BCExpected([object]$Expected){
    if($Expected.Schema -ne 1 -or $Expected.NextSequence -eq 0){throw 'P2BC_INVALID'}
    $pair=([int]$Expected.Sender).ToString([Globalization.CultureInfo]::InvariantCulture)+':'+([int]$Expected.Receiver).ToString([Globalization.CultureInfo]::InvariantCulture)
    if($pair -cnotin @('1:2','2:1','2:3','3:2')){throw 'P2BC_INVALID'}
    $bounds=Get-P2BCDeclaredBounds $Expected.OriginTick $Expected.Frequency $Expected.DeadlineTick
    if($Expected.StopAtTick -ne $bounds.StopAt){throw 'P2BC_INVALID'}
}
function Test-P2BCBytesEqual([byte[]]$Left,[byte[]]$Right){
    if($Left.GetLength(0) -ne $Right.GetLength(0)){return $false}
    for($i=0;$i -lt $Left.GetLength(0);$i++){if($Left[$i] -ne $Right[$i]){return $false}}
    return $true
}
function ConvertTo-P2BControlFrame([object]$Message){
    Set-StrictMode -Version Latest
    $metadata=New-P2BCMetadata
    try{
        $m=Get-P2BCSnapshot $Message $false $metadata;Assert-P2BCMessage $m
        $bytes=[byte[]]::new(256);$bytes[0]=80;$bytes[1]=50;$bytes[2]=66;$bytes[3]=67
        Write-P2BCUInt $bytes 4 2 ([uint64]$m.Version);$bytes[6]=$m.Kind;$bytes[7]=$m.Sender;$bytes[8]=$m.Receiver
        [Array]::Copy($m.BindingDigest,0,$bytes,16,32);[Array]::Copy($m.SessionId,0,$bytes,48,16)
        Write-P2BCUInt $bytes 64 8 $m.Sequence
        foreach($row in @(@(72,'OriginTick'),@(80,'Frequency'),@(88,'DeadlineTick'),@(96,'ObservedTick'))){Write-P2BCUInt $bytes $row[0] 8 ([uint64]$m.($row[1]))}
        Write-P2BCUInt $bytes 104 2 ([uint64]$m.Payload.GetLength(0));[Array]::Copy($m.Payload,0,$bytes,112,$m.Payload.GetLength(0))
        return New-P2BCResult $true 'encoded' $bytes $null $null $metadata
    }catch{$reason=if($_.Exception.Message -ceq 'P2BC_INVALID'){'invalid-input'}else{'internal-error'};return New-P2BCResult $false $reason $null $null $null $metadata}
}
function ConvertFrom-P2BControlFrame([object]$Bytes){
    Set-StrictMode -Version Latest
    $metadata=New-P2BCMetadata;$content=$false
    try{
        # Input type/decorations are distinct from malformed frame contents.
        Assert-P2BCPlain $Bytes ([byte[]]) $metadata
        $content=$true;$snapshot=Get-P2BCByteSnapshot $Bytes 256 256 $metadata
        if($snapshot[0] -ne 80 -or $snapshot[1] -ne 50 -or $snapshot[2] -ne 66 -or $snapshot[3] -ne 67){throw 'P2BC_INVALID'}
        Assert-P2BCZeroBytes $snapshot 9 7
        Assert-P2BCZeroBytes $snapshot 106 6
        Assert-P2BCZeroBytes $snapshot 224 32
        $length=Read-P2BCUInt $snapshot 104 2;if($length -gt 112){throw 'P2BC_INVALID'}
        Assert-P2BCZeroBytes $snapshot (112+[int]$length) (112-[int]$length)
        $ticks=@();foreach($offset in @(72,80,88,96)){$n=Read-P2BCUInt $snapshot $offset 8;if($n -gt [uint64][long]::MaxValue){throw 'P2BC_INVALID'};$ticks+=,[long]$n}
        $digest=[byte[]]::new(32);$session=[byte[]]::new(16);$payload=[byte[]]::new([int]$length)
        [Array]::Copy($snapshot,16,$digest,0,32);[Array]::Copy($snapshot,48,$session,0,16);[Array]::Copy($snapshot,112,$payload,0,[int]$length)
        $m=[pscustomobject][ordered]@{Version=[uint16](Read-P2BCUInt $snapshot 4 2);Kind=[byte]$snapshot[6];Sender=[byte]$snapshot[7];Receiver=[byte]$snapshot[8];BindingDigest=$digest;SessionId=$session;Sequence=(Read-P2BCUInt $snapshot 64 8);OriginTick=$ticks[0];Frequency=$ticks[1];DeadlineTick=$ticks[2];ObservedTick=$ticks[3];Payload=$payload}
        Assert-P2BCMessage $m
        return New-P2BCResult $true 'decoded' $null $m $null $metadata
    }catch{
        $reason=if($_.Exception.Message -cne 'P2BC_INVALID'){'internal-error'}elseif($content){'invalid-frame'}else{'invalid-input'}
        return New-P2BCResult $false $reason $null $null $null $metadata
    }
}
function Test-P2BControlBinding([object]$Message,[object]$Expected){
    Set-StrictMode -Version Latest
    $metadata=New-P2BCMetadata
    try{
        $message=Get-P2BCSnapshot $Message $false $metadata;$expected=Get-P2BCSnapshot $Expected $true $metadata
        Assert-P2BCMessage $message;Assert-P2BCExpected $expected
        if($message.Sender -ne $expected.Sender -or $message.Receiver -ne $expected.Receiver -or
           -not (Test-P2BCBytesEqual $message.BindingDigest $expected.BindingDigest) -or -not (Test-P2BCBytesEqual $message.SessionId $expected.SessionId) -or
           $message.OriginTick -ne $expected.OriginTick -or $message.Frequency -ne $expected.Frequency -or $message.DeadlineTick -ne $expected.DeadlineTick){
            return New-P2BCResult $false 'binding-mismatch' $null $null $null $metadata
        }
        if($message.Sequence -ne $expected.NextSequence){return New-P2BCResult $false 'sequence-mismatch' $null $null $null $metadata}
        if($message.Sequence -eq [uint64]::MaxValue){return New-P2BCResult $false 'sequence-exhausted' $null $null $null $metadata}
        $kind=[int]$message.Kind;$reference=if($kind -in @(5,6,7,8,10)){Read-P2BCUInt $message.Payload 8 8}else{[uint64]0}
        if(($kind -in @(5,6,7) -and ($expected.ArmSequence -eq 0 -or $reference -ne $expected.ArmSequence)) -or
           ($kind -eq 10 -and ($expected.StopSequence -eq 0 -or $reference -ne $expected.StopSequence)) -or
           ($kind -eq 8 -and $reference -ne $expected.LastNonHeartbeatSequence)){
            return New-P2BCResult $false 'reference-mismatch' $null $null $null $metadata
        }
        if($kind -eq 4 -and $message.ObservedTick -ge $expected.StopAtTick){return New-P2BCResult $false 'expired-declaration' $null $null $null $metadata}
        $next=[uint64][Numerics.BigInteger]::Add([Numerics.BigInteger]$message.Sequence,[Numerics.BigInteger]::One)
        return New-P2BCResult $true 'binding-matched' $null $null $next $metadata
    }catch{$reason=if($_.Exception.Message -ceq 'P2BC_INVALID'){'invalid-input'}else{'internal-error'};return New-P2BCResult $false $reason $null $null $null $metadata}
}
