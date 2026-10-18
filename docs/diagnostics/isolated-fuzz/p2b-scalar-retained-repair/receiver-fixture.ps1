function Need([bool]$Condition){if(-not $Condition){throw 'Diagnostic construction precondition'}}
function Raw-Type($Value){
    $argsLocal=[object[]]::new(2);$argsLocal[0]=$Value;$argsLocal[1]=$null
    return $script:invokeMethod.Invoke($script:typeMethod,$argsLocal)
}
function Same($Left,$Right){
    $argsLocal=[object[]]::new(2);$argsLocal[0]=$Left;$argsLocal[1]=$Right
    return [bool]$script:referenceMethod.Invoke($null,$argsLocal)
}
function Ref-ID($Value){
    if($null -eq $Value){return $null}
    # Fixed CLR reference-equality comparer; object[] preserves PSObject keys
    # through reflection binding. Ordinal IDs are assigned, never hash codes.
    $argsLocal=[object[]]::new(2);$argsLocal[0]=$Value;$argsLocal[1]=0
    if($script:idLookup.Invoke($script:referenceIds,$argsLocal)){return [int]$argsLocal[1]}
    $id=$record.References.Count+1;$argsLocal[1]=$id
    $null=$script:idAdd.Invoke($script:referenceIds,$argsLocal)
    $record.References+=,[pscustomobject]@{ID=$id;CLRType=(Raw-Type $Value).FullName}
    return $id
}
function Read-Field($Field,$Target){
    Need ($null -ne $Field)
    $argsLocal=[object[]]::new(1);$argsLocal[0]=$Target
    $result=[object[]]::new(1);$result[0]=$script:fieldReadMethod.Invoke($Field,$argsLocal)
    return ,$result
}
function Call-Method($Method,$Target,[object[]]$Arguments){
    Need ($null -ne $Method)
    $result=[object[]]::new(1);$result[0]=$Method.Invoke($Target,$Arguments)
    return ,$result
}
function New-TestBase([string]$Kind){
    $result=[object[]]::new(1)
    if($Kind -ceq 'String'){$result[0]=[string]::new([char[]]@([char]112,[char]114,[char]111,[char]98,[char]101))}
    elseif($Kind -ceq 'Bytes'){$result[0]=[byte[]]@(7,9)}
    else{
        $type=switch -CaseSensitive ($Kind){Byte{[byte]} UInt16{[uint16]} UInt32{[uint32]} UInt64{[uint64]} Int32{[int]} Int64{[long]} Boolean{[bool]} default{throw 'Unknown test type'}}
        $a=[Array]::CreateInstance($type,1)
        if($Kind -cne 'Boolean'){$a.SetValue([Convert]::ChangeType(7,$type,[Globalization.CultureInfo]::InvariantCulture),0)}
        $result[0]=$a.GetValue(0)
    }
    return ,$result
}
function New-TestSource([string]$Kind){
    $a=New-TestBase $Kind
    $source=[object[]]::new(1);$source[0]=$script:objectCtor.Invoke($a)
    return ,$source
}
function Wrapper-State($Wrapper){
    if($null -eq $Wrapper){return $null}
    $type=Raw-Type $Wrapper
    $state=[ordered]@{WrapperID=(Ref-ID $Wrapper);CLRType=$type.FullName;ImmediateBaseID=$null;UltimateBaseID=$null;KeyID=$null;TableID=$null;Flags=$null}
    if($type -ne [Management.Automation.PSObject]){return [pscustomobject]$state}
    $base=(Read-Field $script:baseField $Wrapper)[0];$state.ImmediateBaseID=Ref-ID $base
    $current=$base
    for($n=0;(Raw-Type $current) -eq [Management.Automation.PSObject];$n++){
        Need ($n -lt 8);$current=(Read-Field $script:baseField $current)[0]
    }
    $state.UltimateBaseID=Ref-ID $current
    $argsLocal=[object[]]::new(1);$argsLocal[0]=$Wrapper
    $key=(Call-Method $script:keyMethod $null $argsLocal)[0]
    $state.KeyID=Ref-ID $key
    $state.TableID=Ref-ID ((Read-Field $script:tableField $Wrapper)[0])
    $state.Flags=[int]((Read-Field $script:flagsField $Wrapper)[0])
    return [pscustomobject]$state
}
function Snapshot($Wrapper){
    $state=Wrapper-State $Wrapper
    $out=[pscustomobject][ordered]@{Wrapper=$state;Descriptors=@()}
    $raw=(Read-Field $script:tableField $Wrapper)[0]
    if($null -eq $raw){return $out}
    Need ((Raw-Type $raw) -eq $script:rawField.FieldType)
    foreach($entry in $raw){
        $type=Raw-Type $entry
        $owner=(Read-Field $script:ownerField $entry)[0]
        $internal=$null;$constructor=$null;$rawTableID=$null
        if($type -eq $script:shimType){$internal=(Read-Field $script:shimOwnerField $entry)[0]}
        if($type -eq [Management.Automation.PSMemberSet]){
            $constructor=(Read-Field $script:constructorOwnerField $entry)[0]
            $rawTableID=Ref-ID ((Read-Field $script:rawField $entry)[0])
        }
        $out.Descriptors+=,[pscustomobject][ordered]@{ID=(Ref-ID $entry);Name=((Read-Field $script:nameField $entry)[0]);CLRType=$type.FullName;Hidden=((Read-Field $script:hiddenField $entry)[0]);Reserved=((Read-Field $script:reservedField $entry)[0]);Instance=((Read-Field $script:instanceField $entry)[0]);Owner=(Wrapper-State $owner);InternalOwner=(Wrapper-State $internal);ConstructorOwner=(Wrapper-State $constructor);ExtendedRawTableID=$rawTableID}
    }
    return $out
}
function Generate([string]$View,$Wrapper){
    $method=if($View -ceq 'psobject'){$script:psobjectFactory}else{$script:extendedFactory}
    $argsLocal=[object[]]::new(1);$argsLocal[0]=$Wrapper
    $descriptor=(Call-Method $method $null $argsLocal)[0]
    return [pscustomobject]@{View=$View;RequestedWrapperID=(Ref-ID $Wrapper);ReturnedDescriptorID=(Ref-ID $descriptor)}
}
function Observe-Phase($Row,[string]$Name,$Left,$Right,[object[]]$Calls){
    $phase=[pscustomobject][ordered]@{Name=$Name;Calls=$Calls;Left=$null;Right=$null;Relations=$null;Complete=$false}
    $Row.Phases+=,$phase
    $phase.Left=Snapshot $Left;$phase.Right=Snapshot $Right
    $l=$phase.Left.Wrapper;$r=$phase.Right.Wrapper
    $phase.Relations=[pscustomobject]@{WrapperSame=($l.WrapperID -eq $r.WrapperID);ImmediateBaseSame=($l.ImmediateBaseID -eq $r.ImmediateBaseID);UltimateBaseSame=($l.UltimateBaseID -eq $r.UltimateBaseID);KeySame=($l.KeyID -eq $r.KeyID);TableSame=($null -ne $l.TableID -and $l.TableID -eq $r.TableID)}
    $phase.Complete=$true
}


function Capture-P2BTestReceiver([object[]]$Source){
    Need ($Source.Count -eq 1 -and (Raw-Type $Source[0]) -eq [Management.Automation.PSObject])
    $handle=[object[]]::new(3)
    $handle[0]=$script:asObject.Invoke($null,$Source)
    $handle[1]=(Read-Field $script:baseField $handle[0])[0]
    $handle[2]=(Read-Field $script:tableField $handle[0])[0]
    return ,$handle
}
function Relay-P2BTestReceiver([object[]]$Handle){
    Need ($Handle.Count -eq 3)
    return ,$Handle
}
function Observe-P2BTestReceiver([object[]]$Handle){return Snapshot $Handle[0]}
function Get-P2BRetentionVerdict($Before,$After,[int]$DescriptorID,[bool]$WantHidden,[bool]$WantSubclass){
    $sameDescriptors=$Before.Descriptors.Count -eq $After.Descriptors.Count
    $tagMatches=0;$tagCorrect=($DescriptorID -eq 0)
    for($i=0;$i -lt $After.Descriptors.Count;$i++){
        $entry=$After.Descriptors[$i]
        if($i -ge $Before.Descriptors.Count -or $entry.ID -ne $Before.Descriptors[$i].ID){$sameDescriptors=$false}
        if($entry.Name -ceq 'AttachedProbe'){
            $tagMatches++
            $want=if($WantSubclass){'P2BDescriptorFixture.CounterNote'}else{'System.Management.Automation.PSNoteProperty'}
            $tagCorrect=$entry.ID -eq $DescriptorID -and $entry.CLRType -ceq $want -and $entry.Hidden -eq $WantHidden -and -not $entry.Reserved -and $entry.Instance
        }
    }
    [pscustomobject]@{
        WrapperSame=($Before.Wrapper.WrapperID -eq $After.Wrapper.WrapperID)
        BaseSame=($Before.Wrapper.UltimateBaseID -eq $After.Wrapper.UltimateBaseID)
        KeySame=($Before.Wrapper.KeyID -eq $After.Wrapper.KeyID)
        TableSame=($Before.Wrapper.TableID -eq $After.Wrapper.TableID)
        DescriptorReferencesSame=$sameDescriptors
        TagCountCorrect=($tagMatches -eq $(if($DescriptorID -eq 0){0}else{1}))
        TagIdentityClassFlagsCorrect=$tagCorrect
    }
}
function All-P2BRetentionConditions($Verdict){
    foreach($property in $Verdict.PSObject.Properties){if(-not [bool]$property.Value){return $false}}
    return $true
}
