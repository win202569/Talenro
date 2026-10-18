# Bridge observations use actual PSObject handles; product helpers are not the oracle.
function BT-Init {
    $script:fieldReadMethod=[Reflection.FieldInfo].GetMethod('GetValue',[type[]]@([object]))
    $script:referenceIds=[Collections.Generic.Dictionary[object,int]]::new([Collections.Generic.ReferenceEqualityComparer]::Instance)
    $script:idLookup=$script:referenceIds.GetType().GetMethod('TryGetValue',[type[]]@([object],[int].MakeByRefType()))
    $script:idAdd=$script:referenceIds.GetType().GetMethod('Add',[type[]]@([object],[int]))
    $flags=[Reflection.BindingFlags]'NonPublic,Instance,DeclaredOnly'
    $sdk=[Management.Automation.PSObject].Assembly;$script:shimType=$sdk.GetType('System.Management.Automation.PSInternalMemberSet',$true)
    $script:reserved=$sdk.GetType('System.Management.Automation.ReservedNameMembers',$true)
    $script:objectCtor=[Management.Automation.PSObject].GetConstructor([type[]]@([object]))
    $script:asObject=[Management.Automation.PSObject].GetMethod('AsPSObject',[type[]]@([object]))
    $script:baseField=[Management.Automation.PSObject].GetField('_immediateBaseObject',$flags)
    $script:tableField=[Management.Automation.PSObject].GetField('_instanceMembers',$flags)
    $script:flagsField=[Management.Automation.PSObject].GetField('_flags',$flags)
    $script:keyMethod=[Management.Automation.PSObject].GetMethod('GetKeyForResurrectionTables',[Reflection.BindingFlags]'NonPublic,Static,DeclaredOnly')
    $script:psobjectFactory=$script:reserved.GetMethod('GeneratePSObjectMemberSet',[Reflection.BindingFlags]'NonPublic,Static,DeclaredOnly')
    $script:extendedFactory=$script:reserved.GetMethod('GeneratePSExtendedMemberSet',[Reflection.BindingFlags]'NonPublic,Static,DeclaredOnly')
    $script:rawField=[Management.Automation.PSMemberSet].GetField('internalMembers',$flags)
    $script:ownerField=[Management.Automation.PSMemberInfo].GetField('instance',$flags)
    $script:nameField=[Management.Automation.PSMemberInfo].GetField('name',$flags)
    $script:hiddenField=[Management.Automation.PSMemberInfo].GetField('<IsHidden>k__BackingField',$flags)
    $script:reservedField=[Management.Automation.PSMemberInfo].GetField('<IsReservedMember>k__BackingField',$flags)
    $script:instanceField=[Management.Automation.PSMemberInfo].GetField('<IsInstance>k__BackingField',$flags)
    $script:shimOwnerField=$script:shimType.GetField('_psObject',$flags)
    $script:constructorOwnerField=[Management.Automation.PSMemberSet].GetField('_constructorPSObject',$flags)
}
function BT-Views([object[]]$Source,[int]$History){
    $null=Generate 'psobject' $Source[0];$null=Generate 'psextended' $Source[0]
    for($n=0;$n -lt $History;$n++){
        foreach($name in @('GeneratePSBaseMemberSet','GeneratePSAdaptedMemberSet','GeneratePSTypeNames')){
            $method=$script:reserved.GetMethod($name,[Reflection.BindingFlags]'NonPublic,Static,DeclaredOnly')
            $null=$method.Invoke($null,$Source)
        }
    }
}
function BT-FromValue($Value){
    $shim=$Value.PSObject
    Need ((BC-Type $shim) -eq $script:shimType)
    $source=[object[]]::new(1);$source[0]=(Read-Field $script:shimOwnerField $shim.PSObject.BaseObject)[0]
    return ,$source
}
function BT-InitialProof {
    [P2BDescriptorFixture.CounterNote]::Reset()
    foreach($kind in @('Byte','UInt16','UInt32','UInt64','Int32','Int64','Boolean','String','Bytes')){
        foreach($history in @(0,1,2)){
            $record.Stage='RetainedProof:'+ $kind+':'+$history
            $source=New-TestSource $kind;BT-Views $source $history
            $before=Snapshot $source[0];$handle=Capture-P2BTestReceiver $source;$after=Observe-P2BTestReceiver $handle
            $verdict=Get-P2BRetentionVerdict $before $after 0 $false $false
            $row=[pscustomobject]@{Kind=$kind;History=$history;Before=$before;After=$after;Verdict=$verdict;Retained=(Same $source[0] $handle[0]);Counters=[P2BDescriptorFixture.CounterNote]::Total;Validated=$false}
            $record.Proofs+=,$row
            Need ($row.Retained -and (All-P2BRetentionConditions $verdict) -and $row.Counters -eq 0 -and $after.Descriptors.Count -eq $(if($history -eq 0){2}else{5}))
            foreach($entry in $after.Descriptors){
                Need ($entry.Hidden -and $entry.Reserved -and $entry.Instance -and $entry.Owner.KeyID -eq $after.Wrapper.KeyID -and $entry.Owner.TableID -eq $after.Wrapper.TableID)
                if($entry.CLRType -ceq 'System.Management.Automation.PSInternalMemberSet'){Need ($entry.InternalOwner.WrapperID -eq $entry.Owner.WrapperID)}
                elseif($entry.Name -ceq 'psextended'){Need ($entry.ConstructorOwner.WrapperID -eq $entry.Owner.WrapperID -and $entry.ExtendedRawTableID -eq $after.Wrapper.TableID)}
                elseif($entry.Name -cne 'pstypenames'){throw 'Unexpected SDK proof'}
            }
            $row.Validated=$true
        }
    }
    Need ($record.Proofs.Count -eq 27);$record.ProofsCompleted=$true
}
function BT-HistoryCases {
    foreach($history in @(1,2)){
        foreach($mode in @('encode-scalar','encode-array','binding-message-scalar','binding-message-array','binding-expected-scalar','binding-expected-array','decode-array')){
            $record.Stage='History:'+ $history+':'+$mode
            $m=BC-Message 2;$e=BC-Expected $m
            if($mode -ceq 'decode-array'){$owner=BC-Object ([ordered]@{Data=(BC-Bytes 2)});$name='Data'}
            elseif($mode -clike 'binding-expected-*'){$owner=$e;$name=if($mode -clike '*-scalar'){'Schema'}else{'BindingDigest'}}
            else{$owner=$m;$name=if($mode -clike '*-scalar'){'OriginTick'}else{'Payload'}}
            $source=BT-FromValue $owner.$name;BT-Views $source $history
            $handle=Capture-P2BTestReceiver $source
            $row=[pscustomobject]@{History=$history;Mode=$mode;Before=(Observe-P2BTestReceiver $handle);After=$null;Verdict=$null;Result=$null;Counters=$null;Validated=$false}
            $record.HistoryCases+=,$row
            [P2BDescriptorFixture.CounterNote]::Reset();$record.ProductCalled=$true
            $r=if($mode -ceq 'decode-array'){ConvertFrom-P2BControlFrame $owner.Data}elseif($mode -clike 'encode-*'){ConvertTo-P2BControlFrame $m}else{Test-P2BControlBinding $m $e}
            $row.Result=SP-Result $r;$row.Counters=[P2BDescriptorFixture.CounterNote]::Total;$row.After=Observe-P2BTestReceiver $handle
            $row.Verdict=Get-P2BRetentionVerdict $row.Before $row.After 0 $false $false
            $want=if($mode -ceq 'decode-array'){'decoded'}elseif($mode -clike 'encode-*'){'encoded'}else{'binding-matched'}
            Need ($r.Accepted -and $r.Reason -ceq $want -and $row.Counters -eq 0 -and (All-P2BRetentionConditions $row.Verdict) -and -not $r.SourceVerified -and -not $r.LiveActionsPerformed -and -not $r.RuntimeApproved)
            $row.Validated=$true
        }
    }
}
function BT-GuardControls {
    $meta=New-P2BCMetadata
    foreach($kind in @('Byte','UInt16','UInt32','UInt64','Int32','Int64','Boolean','String','Bytes')){
        foreach($variant in @('clean','visible-note','hidden-note','visible-subclass','hidden-subclass')){
            $record.Stage='Guard:'+ $kind+':'+$variant
            $source=New-TestSource $kind;BT-Views $source 1
            $tag=$null
            if($variant -cne 'clean'){
                if($variant -clike '*subclass'){$tag=[P2BDescriptorFixture.CounterNote]::new('AttachedProbe',[int]42)}
                else{$tag=[Management.Automation.PSNoteProperty]::new('AttachedProbe',[int]42)}
                if($variant -clike 'hidden-*'){$script:hiddenField.SetValue($tag,$true)}
                $raw=(Read-Field $script:tableField $source[0])[0]
                $method=$script:rawField.FieldType.GetMethod('Add',[type[]]@([Management.Automation.PSMemberInfo]))
                $a=[object[]]::new(1);$a[0]=$tag;$null=$method.Invoke($raw,$a)
            }
            $base=(Read-Field $script:baseField $source[0])[0];$required=Raw-Type $base
            $row=[pscustomobject]@{Kind=$kind;Variant=$variant;Before=(Snapshot $source[0]);After=$null;Rejected=$false;CopyType=$null;CopyBaseDistinct=$null;Counters=$null;Validated=$false}
            $record.Guards+=,$row
            [P2BDescriptorFixture.CounterNote]::Reset()
            try{Assert-P2BCPlain $source[0] $required $meta}catch{if($_.Exception.Message -ceq 'P2BC_INVALID'){$row.Rejected=$true}else{throw}}
            if($variant -ceq 'clean'){
                $copy=Copy-P2BCScalar $source[0] $meta
                $row.CopyType=(BC-Type $copy).FullName;$row.CopyBaseDistinct=-not (Same $copy.PSObject.BaseObject $base)
                Need ($row.CopyType -ceq $required.FullName -and $row.CopyBaseDistinct)
            }
            $row.After=Snapshot $source[0];$row.Counters=[P2BDescriptorFixture.CounterNote]::Total
            Need ($row.Rejected -eq ($variant -cne 'clean') -and $row.Counters -eq 0)
            $row.Validated=$true
        }
    }
    Need ($record.Guards.Count -eq 45)
}

# Test-only fixed-field observations, independent of product metadata helpers.
function SC-Entry($Entry,$Owner){
    $flags=[Reflection.BindingFlags]'NonPublic,Instance,DeclaredOnly'
    $t=[Management.Automation.PSMemberInfo];$base=$Entry.PSObject.BaseObject
    $type=BC-Type $Entry;$name=$t.GetField('name',$flags).GetValue($base)
    $shim=[Management.Automation.PSObject].Assembly.GetType('System.Management.Automation.PSInternalMemberSet',$true)
    $internalOwner=$null;$constructorOwner=$null;$getter=$null;$setter=$null
    if($type -eq $shim){$internalOwner=BC-Ref $shim.GetField('_psObject',$flags).GetValue($base) $Owner}
    if($type -eq [Management.Automation.PSMemberSet]){$constructorOwner=BC-Ref ([Management.Automation.PSMemberSet].GetField('_constructorPSObject',$flags).GetValue($base)) $Owner}
    if($type -eq [Management.Automation.PSCodeProperty]){
        $sdk=[Management.Automation.PSObject].Assembly.GetType('System.Management.Automation.Language.CachedReflectionInfo',$true).GetField('ReservedNameMembers_PSTypeNames',[Reflection.BindingFlags]'NonPublic,Static,DeclaredOnly').GetValue($null)
        $getter=BC-Ref ([Management.Automation.PSCodeProperty].GetField('<GetterCodeReference>k__BackingField',$flags).GetValue($base)) $sdk
        $setter=$null -eq [Management.Automation.PSCodeProperty].GetField('<SetterCodeReference>k__BackingField',$flags).GetValue($base)
    }
    [pscustomobject]@{Name=$name;CLRType=$type.FullName;Hidden=$t.GetField('<IsHidden>k__BackingField',$flags).GetValue($base);Reserved=$t.GetField('<IsReservedMember>k__BackingField',$flags).GetValue($base);Instance=$t.GetField('<IsInstance>k__BackingField',$flags).GetValue($base);OwnerSame=(BC-Ref $t.GetField('instance',$flags).GetValue($base) $Owner);InternalOwnerSame=$internalOwner;ConstructorOwnerSame=$constructorOwner;GetterSame=$getter;SetterNull=$setter}
}
function SC-Owner($Value){
    $shim=$Value.PSObject
    $type=[Management.Automation.PSObject].Assembly.GetType('System.Management.Automation.PSInternalMemberSet',$true)
    BC-Need ((BC-Type $shim) -eq $type)
    # Return a containing array, not a PSObject through the pipeline itself.
    $arguments=[object[]]::new(1)
    $arguments[0]=$type.GetField('_psObject',[Reflection.BindingFlags]'NonPublic,Instance,DeclaredOnly').GetValue($shim.PSObject.BaseObject)
    return ,$arguments
}
function SC-Negatives {
    $flags=[Reflection.BindingFlags]'NonPublic,Instance,DeclaredOnly'
    $baseType=[Management.Automation.PSMemberInfo]
    $shimType=[Management.Automation.PSObject].Assembly.GetType('System.Management.Automation.PSInternalMemberSet',$true)
    foreach($mode in @('encode-scalar','encode-array','binding-message-scalar','binding-message-array','binding-expected-scalar','binding-expected-array','decode-array')){
        foreach($sdkName in @('psobject','psextended','psbase','psadapted','pstypenames')){
            $mutations=@('owner','hidden','reserved','instance')
            if($sdkName -cin @('psobject','psbase','psadapted')){$mutations+=,'internal-owner'}
            if($sdkName -ceq 'psextended'){$mutations+=,'constructor-owner'}
            if($sdkName -ceq 'pstypenames'){$mutations+=@('getter','setter')}
            foreach($mutation in $mutations){
                $record.Stage='Negative:'+ $mode+':'+$sdkName+':'+$mutation
                $row=[pscustomobject]@{Mode=$mode;Name=$sdkName;Mutation=$mutation;Before=$null;AfterMutation=$null;MutationProven=$false;Result=$null;Counters=$null;Validated=$false}
                $record.Negatives+=,$row
                $m=BC-Message 2;$e=BC-Expected $m
                if($mode -ceq 'decode-array'){$owner=BC-Object ([ordered]@{Data=(BC-Bytes 2)});$fieldName='Data'}
                elseif($mode -clike 'binding-expected-*'){$owner=$e;$fieldName=if($mode -clike '*-scalar'){'Schema'}else{'BindingDigest'}}
                else{$owner=$m;$fieldName=if($mode -clike '*-scalar'){'OriginTick'}else{'Payload'}}
                $value=$owner.$fieldName;DG-FullHistory $value
                $descriptor=DG-GetDescriptor $value $sdkName;$base=$descriptor.PSObject.BaseObject
                $originalOwner=(SC-Owner $value)[0];$foreign=BC-Object ([ordered]@{Fresh=[long]735});$foreignOwner=(SC-Owner $foreign.Fresh)[0]
                $row.Before=SC-Entry $descriptor $originalOwner
                $field=$null;$newValue=$null
                if($mutation -ceq 'owner'){$field=$baseType.GetField('instance',$flags);$newValue=$foreignOwner}
                elseif($mutation -ceq 'internal-owner'){$field=$shimType.GetField('_psObject',$flags);$newValue=$foreignOwner}
                elseif($mutation -ceq 'constructor-owner'){$field=[Management.Automation.PSMemberSet].GetField('_constructorPSObject',$flags);$newValue=$foreignOwner}
                elseif($mutation -cin @('getter','setter')){
                    $field=[Management.Automation.PSCodeProperty].GetField($(if($mutation -ceq 'getter'){'<GetterCodeReference>k__BackingField'}else{'<SetterCodeReference>k__BackingField'}),$flags)
                    $newValue=[object].GetMethod('GetType',[type[]]@())
                }else{
                    $field=$baseType.GetField($(switch($mutation){hidden{'<IsHidden>k__BackingField'} reserved{'<IsReservedMember>k__BackingField'} instance{'<IsInstance>k__BackingField'}}),$flags);$newValue=$false
                }
                DG-SetField $field $base $newValue
                $row.AfterMutation=SC-Entry $descriptor $originalOwner
                $row.MutationProven=if($mutation -cin @('hidden','reserved','instance')){-not [bool]$field.GetValue($base)}else{DG-Same $field.GetValue($base) $newValue}
                BC-Need $row.MutationProven
                [P2BDescriptorFixture.CounterNote]::Reset()
                $r=if($mode -ceq 'decode-array'){ConvertFrom-P2BControlFrame $owner.Data}elseif($mode -clike 'encode-*'){ConvertTo-P2BControlFrame $m}else{Test-P2BControlBinding $m $e}
                $row.Result=SP-Result $r;$row.Counters=[P2BDescriptorFixture.CounterNote]::Total
                DG-Rejected $r;BC-Need ($row.Counters -eq 0)
                $row.Validated=$true
            }
        }
    }
    BC-Need ($record.Negatives.Count -eq 182)
}
function SC-KeyTable($Wrapper){
    BC-Need ((DG-RawType $Wrapper) -eq [Management.Automation.PSObject])
    $flags=[Reflection.BindingFlags]'NonPublic,Instance,DeclaredOnly'
    $keyMethod=[Management.Automation.PSObject].GetMethod('GetKeyForResurrectionTables',[Reflection.BindingFlags]'NonPublic,Static,DeclaredOnly')
    $a=[object[]]::new(1);$a[0]=$Wrapper
    $result=[object[]]::new(2);$result[0]=$keyMethod.Invoke($null,$a)
    $read=[Reflection.FieldInfo].GetMethod('GetValue',[type[]]@([object]))
    $field=[Management.Automation.PSObject].GetField('_instanceMembers',$flags)
    $result[1]=$read.Invoke($field,$a)
    return ,$result
}
function SC-KeyControls([bool]$VerifyProduct){
    $flags=[Reflection.BindingFlags]'NonPublic,Instance,DeclaredOnly'
    $pso=[Management.Automation.PSObject]
    $sdk=$pso.Assembly.GetType('System.Management.Automation.ReservedNameMembers',$true)
    $ctor=$pso.GetConstructor([type[]]@([object]))
    $makeShim=$sdk.GetMethod('GeneratePSObjectMemberSet',[Reflection.BindingFlags]'NonPublic,Static,DeclaredOnly')
    $makeExtended=$sdk.GetMethod('GeneratePSExtendedMemberSet',[Reflection.BindingFlags]'NonPublic,Static,DeclaredOnly')
    $local=$pso.GetProperty('StoreTypeNameAndInstanceMembersLocally',[Reflection.BindingFlags]'NonPublic,Instance,DeclaredOnly')
    BC-Need ($null -ne $local -and $local.CanWrite)
    $samples=BC-Object ([ordered]@{Byte=[byte]7;UInt16=[uint16]1;UInt32=[uint32]1;UInt64=[uint64]1;Int32=[int]1;Int64=[long]1;Boolean=$false;String='probe';Bytes=[byte[]]@(1,2)})
    foreach($property in $samples.PSObject.Properties){
        foreach($localMode in @($false,$true)){
            foreach($scenario in @('same-wrapper','same-base-distinct','equal-content-independent')){
                $record.Stage='KeyControl:'+ $property.Name+':'+$localMode+':'+$scenario
                $row=[pscustomobject]@{Type=$property.Name;Local=$localMode;Scenario=$scenario;WrapperSame=$null;BaseSame=$null;KeySame=$null;TableSame=$null;ExpectedShared=$null;ProductChecked=$false;Validated=$false}
                $record.KeyControls+=,$row
                $leftBase=(BC-Fresh $property.Value).PSObject.BaseObject
                $rightBase=$leftBase
                if($scenario -ceq 'equal-content-independent'){$rightBase=(BC-Fresh $property.Value).PSObject.BaseObject}
                $a=[object[]]::new(1);$a[0]=$leftBase
                $left=$ctor.Invoke($a);$right=$left
                if($scenario -cne 'same-wrapper'){$a[0]=$rightBase;$right=$ctor.Invoke($a)}
                # Reflection binder is intentionally given an argument array to preserve wrapper receiver.
                $set=[Reflection.PropertyInfo].GetMethod('SetValue',[type[]]@([object],[object],[object[]]))
                foreach($wrapper in @($left,$right)){
                    $argsLocal=[object[]]::new(3);$argsLocal[0]=$wrapper;$argsLocal[1]=$localMode;$argsLocal[2]=$null
                    $null=$set.Invoke($local,$argsLocal)
                }
                $a[0]=$left;$null=$makeShim.Invoke($null,$a);$null=$makeExtended.Invoke($null,$a)
                $a[0]=$right;$null=$makeShim.Invoke($null,$a);$null=$makeExtended.Invoke($null,$a)
                $l=SC-KeyTable $left;$r=SC-KeyTable $right
                $row.WrapperSame=BC-Ref $left $right;$row.BaseSame=BC-Ref $leftBase $rightBase
                $row.KeySame=BC-Ref $l[0] $r[0];$row.TableSame=BC-Ref $l[1] $r[1]
                $row.ExpectedShared=($scenario -ceq 'same-wrapper' -or ($scenario -ceq 'same-base-distinct' -and -not $localMode -and $property.Name -cne 'String'))
                BC-Need ($row.WrapperSame -eq ($scenario -ceq 'same-wrapper') -and $row.BaseSame -eq ($scenario -cne 'equal-content-independent') -and $row.KeySame -eq $row.ExpectedShared -and $row.TableSame -eq $row.ExpectedShared)
                if($VerifyProduct){
                    $metadata=New-P2BCMetadata
                    # Invoke product ownership reader directly on actual wrapper, not through coercing APIs.
                    $actual=Get-P2BCStorage $left $metadata
                    BC-Need ((BC-Ref $actual[0] $l[0]) -and (BC-Ref $actual[1] $l[1]))
                    $actual=Get-P2BCStorage $right $metadata
                    BC-Need ((BC-Ref $actual[0] $r[0]) -and (BC-Ref $actual[1] $r[1]))
                    $row.ProductChecked=$true
                }
                $row.Validated=$true
            }
        }
    }
    BC-Need ($record.KeyControls.Count -eq 54)
}
