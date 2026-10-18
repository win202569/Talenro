function BC-Fresh($Value){
    if($null -eq $Value){return $null}
    $type=BC-Type $Value;$base=$Value.PSObject.BaseObject
    if($type -eq [byte[]]){return ,([byte[]]$base.Clone())}
    if($type -eq [string]){return ,([string]::new($base.ToCharArray()))}
    $storage=[Array]::CreateInstance($type,1);$storage.SetValue($base,0);return ,($storage.GetValue(0))
}
# Test-only oracle uses the pinned SDK key method, never a product helper.
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
function SC-Entry($Entry,$Owner){
    $flags=[Reflection.BindingFlags]'NonPublic,Instance,DeclaredOnly'
    $t=[Management.Automation.PSMemberInfo];$base=$Entry.PSObject.BaseObject
    $type=BC-Type $Entry;$name=$t.GetField('name',$flags).GetValue($base)
    $shim=[Management.Automation.PSObject].Assembly.GetType('System.Management.Automation.PSInternalMemberSet',$true)
    $own=$t.GetField('instance',$flags).GetValue($base)
    $subjectState=SC-KeyTable $Owner;$ownState=SC-KeyTable $own
    $internalOwner=$null;$constructorOwner=$null;$extendedTable=$null;$getter=$null;$setter=$null
    if($type -eq $shim){$internalOwner=BC-Ref $shim.GetField('_psObject',$flags).GetValue($base) $own}
    if($type -eq [Management.Automation.PSMemberSet]){
        $constructorOwner=BC-Ref ([Management.Automation.PSMemberSet].GetField('_constructorPSObject',$flags).GetValue($base)) $own
        $extendedTable=BC-Ref ([Management.Automation.PSMemberSet].GetField('internalMembers',$flags).GetValue($base)) $subjectState[1]
    }
    if($type -eq [Management.Automation.PSCodeProperty]){
        $sdk=[Management.Automation.PSObject].Assembly.GetType('System.Management.Automation.Language.CachedReflectionInfo',$true).GetField('ReservedNameMembers_PSTypeNames',[Reflection.BindingFlags]'NonPublic,Static,DeclaredOnly').GetValue($null)
        $getter=BC-Ref ([Management.Automation.PSCodeProperty].GetField('<GetterCodeReference>k__BackingField',$flags).GetValue($base)) $sdk
        $setter=$null -eq [Management.Automation.PSCodeProperty].GetField('<SetterCodeReference>k__BackingField',$flags).GetValue($base)
    }
    [pscustomobject]@{Name=$name;CLRType=$type.FullName;Hidden=$t.GetField('<IsHidden>k__BackingField',$flags).GetValue($base);Reserved=$t.GetField('<IsReservedMember>k__BackingField',$flags).GetValue($base);Instance=$t.GetField('<IsInstance>k__BackingField',$flags).GetValue($base);OwnerSame=(BC-Ref $own $Owner);OwnerKeySame=(BC-Ref $ownState[0] $subjectState[0]);OwnerTableSame=(BC-Ref $ownState[1] $subjectState[1]);InternalOwnerSame=$internalOwner;ConstructorOwnerSame=$constructorOwner;ExtendedTableSame=$extendedTable;GetterSame=$getter;SetterNull=$setter}
}
function SC-ValidateEntries($Entries,$Raw,$Owner){
    $state=SC-KeyTable $Owner
    BC-Need (BC-Ref $state[1] $Raw)
    $seen=[Collections.Generic.HashSet[string]]::new([StringComparer]::Ordinal)
    foreach($entry in $Entries){
        BC-Need ($seen.Add($entry.Name) -and $entry.Hidden -and $entry.Reserved -and $entry.Instance -and $entry.OwnerKeySame -and $entry.OwnerTableSame)
        switch -CaseSensitive ($entry.Name){
            {$_ -cin @('psobject','psbase','psadapted')}{BC-Need ($entry.CLRType -ceq 'System.Management.Automation.PSInternalMemberSet' -and $entry.InternalOwnerSame);break}
            'psextended'{BC-Need ($entry.CLRType -ceq 'System.Management.Automation.PSMemberSet' -and $entry.ConstructorOwnerSame -and $entry.ExtendedTableSame);break}
            'pstypenames'{BC-Need ($entry.CLRType -ceq 'System.Management.Automation.PSCodeProperty' -and $entry.GetterSame -and $entry.SetterNull);break}
            default{throw 'Unexpected scalar descriptor'}
        }
    }
    BC-Need ($seen.Contains('psobject') -and $seen.Contains('psextended'))
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
function SC-ProveValue($Value,[string]$Label,[int]$History){
    $proof=[pscustomobject][ordered]@{Label=$Label;History=$History;ValueType=$null;OwnerType=$null;Entries=@();SDKCount=$null;ApplicationCount=$null;Counters=$null;Validated=$false}
    $record.Proofs+=,$proof;$record.Stage='SDKProof:'+ $Label
    $proof.ValueType=(BC-Type $Value).FullName
    for($n=0;$n -lt $History;$n++){DG-FullHistory $Value}
    $raw=DG-RawMembers $Value;$owner=(SC-Owner $Value)[0]
    $proof.OwnerType=(DG-RawType $owner).FullName
    foreach($entry in $raw){$proof.Entries+=,(SC-Entry $entry $owner)}
    # All observations above survive a failure in the independent fixture checker.
    SC-ValidateEntries $proof.Entries $raw $owner
    $proof.SDKCount=$proof.Entries.Count;$proof.ApplicationCount=0
    $proof.Counters=[P2BDescriptorFixture.CounterNote]::Total
    BC-Need ($proof.ApplicationCount -eq 0 -and $proof.Counters -eq 0)
    BC-Need ($proof.SDKCount -ge 2 -and $proof.SDKCount -le 5)
    if($History -gt 0){BC-Need ($proof.SDKCount -eq 5)}
    $proof.Validated=$true
}
function SC-ProveLeaves($Value,[string]$Label){
    if($null -eq $Value){return}
    if((BC-Type $Value) -eq [Management.Automation.PSCustomObject]){
        foreach($p in $Value.PSObject.Properties){SC-ProveLeaves $p.Value ($Label+'.'+$p.Name)}
    }else{foreach($history in @(0,1,2)){SC-ProveValue $Value $Label $history}}
}
function SC-InitialProof {
    [P2BDescriptorFixture.CounterNote]::Reset()
    $samples=BC-Object ([ordered]@{Byte=[byte]7;UInt16=[uint16]1;UInt32=[uint32]1;UInt64=[uint64]1;Int32=[int]1;Int64=[long]1;Boolean=$false;String='probe';Bytes=[byte[]]@(1,2)})
    foreach($p in $samples.PSObject.Properties){foreach($history in @(0,1,2)){SC-ProveValue $p.Value ('literal.'+$p.Name) $history}}
    $record.Stage='OwnedOutputProof'
    $m=BC-Message 2;$e=BC-Expected $m;$record.ProductCalled=$true
    $a=ConvertTo-P2BControlFrame $m;$b=ConvertFrom-P2BControlFrame (BC-Bytes 2);$c=Test-P2BControlBinding $m $e
    BC-Need ($a.Accepted -and $a.Reason -ceq 'encoded' -and $b.Accepted -and $b.Reason -ceq 'decoded' -and $c.Accepted -and $c.Reason -ceq 'binding-matched')
    SC-ProveLeaves $a 'encoded';SC-ProveLeaves $b 'decoded';SC-ProveLeaves $c 'binding'
    BC-Need ($record.Proofs.Count -eq 132)
    $record.ProofsCompleted=$true
}
function SC-HistoryCases {
    foreach($history in @(1,2)){
        foreach($mode in @('encode-scalar','encode-array','binding-message-scalar','binding-message-array','binding-expected-scalar','binding-expected-array','decode-array')){
            $record.Stage='History:'+ $history+':'+$mode
            $row=[pscustomobject]@{History=$history;Mode=$mode;Validated=$false;Result=$null;Counters=$null}
            $record.HistoryCases+=,$row
            $m=BC-Message 2;$e=BC-Expected $m
            if($mode -ceq 'decode-array'){$owner=BC-Object ([ordered]@{Data=(BC-Bytes 2)});$name='Data'}
            elseif($mode -clike 'binding-expected-*'){$owner=$e;$name=if($mode -clike '*-scalar'){'Schema'}else{'BindingDigest'}}
            else{$owner=$m;$name=if($mode -clike '*-scalar'){'OriginTick'}else{'Payload'}}
            SC-ProveValue $owner.$name ('history.'+$mode) $history
            $record.Stage='HistoryAPI:'+ $history+':'+$mode
            [P2BDescriptorFixture.CounterNote]::Reset()
            $r=if($mode -ceq 'decode-array'){ConvertFrom-P2BControlFrame $owner.Data}elseif($mode -clike 'encode-*'){ConvertTo-P2BControlFrame $m}else{Test-P2BControlBinding $m $e}
            $row.Result=SP-Result $r;$row.Counters=[P2BDescriptorFixture.CounterNote]::Total
            $want=if($mode -ceq 'decode-array'){'decoded'}elseif($mode -clike 'encode-*'){'encoded'}else{'binding-matched'}
            BC-Need ($r.Accepted -and $r.Reason -ceq $want -and $row.Counters -eq 0 -and -not $r.SourceVerified -and -not $r.LiveActionsPerformed -and -not $r.RuntimeApproved)
            $row.Validated=$true
        }
    }
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

# Deliberate wrapper controls. Constructors/factories/key oracle are fixture-only.
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

function SC-KeyNegatives {
    $flags=[Reflection.BindingFlags]'NonPublic,Instance,DeclaredOnly'
    $pso=[Management.Automation.PSObject]
    $ctor=$pso.GetConstructor([type[]]@([object]))
    $sdk=$pso.Assembly.GetType('System.Management.Automation.ReservedNameMembers',$true)
    $factory=$sdk.GetMethod('GeneratePSExtendedMemberSet',[Reflection.BindingFlags]'NonPublic,Static,DeclaredOnly')
    $ownerField=[Management.Automation.PSMemberInfo].GetField('instance',$flags)
    $shim=$pso.Assembly.GetType('System.Management.Automation.PSInternalMemberSet',$true)
    $internalField=$shim.GetField('_psObject',$flags)
    $tableField=$pso.GetField('_instanceMembers',$flags)
    $samples=BC-Object ([ordered]@{Byte=[byte]7;UInt16=[uint16]1;UInt32=[uint32]1;UInt64=[uint64]1;Int32=[int]1;Int64=[long]1;Boolean=$false;String='probe';Bytes=[byte[]]@(1,2)})
    foreach($property in $samples.PSObject.Properties){
        $mutations=@('equal-content-foreign-key','foreign-table')
        if($property.Name -ceq 'String'){$mutations+=,'same-string-base-foreign-wrapper'}
        foreach($mutation in $mutations){
            $record.Stage='KeyNegative:'+ $property.Name+':'+$mutation
            $row=[pscustomobject]@{Type=$property.Name;Mutation=$mutation;Before=$null;After=$null;ForeignKeyDifferent=$null;ForeignTableDifferent=$null;MutationProven=$false;Rejected=$false;Counters=$null;Validated=$false}
            $record.KeyNegatives+=,$row
            $value=(BC-Fresh $property.Value).PSObject.BaseObject
            $raw=DG-RawMembers $value;$owner=(SC-Owner $value)[0]
            $descriptor=DG-GetDescriptor $value 'psobject'
            $foreignBase=(BC-Fresh $property.Value).PSObject.BaseObject
            if($mutation -ceq 'same-string-base-foreign-wrapper'){$foreignBase=$value.PSObject.BaseObject}
            $a=[object[]]::new(1);$a[0]=$foreignBase
            $foreign=$ctor.Invoke($a);$a[0]=$foreign;$null=$factory.Invoke($null,$a)
            $s=SC-KeyTable $owner;$t=SC-KeyTable $foreign
            $row.Before=SC-Entry $descriptor $owner
            $row.ForeignKeyDifferent=-not (BC-Ref $s[0] $t[0])
            $row.ForeignTableDifferent=-not (BC-Ref $s[1] $t[1])
            BC-Need ($row.ForeignKeyDifferent -and $row.ForeignTableDifferent)
            if($mutation -ceq 'foreign-table'){
                DG-SetField $tableField $owner $t[1]
                $actual=SC-KeyTable $owner
                $row.MutationProven=BC-Ref $actual[1] $t[1]
            }else{
                DG-SetField $ownerField $descriptor.PSObject.BaseObject $foreign
                DG-SetField $internalField $descriptor.PSObject.BaseObject $foreign
                $row.MutationProven=(BC-Ref $ownerField.GetValue($descriptor.PSObject.BaseObject) $foreign) -and (BC-Ref $internalField.GetValue($descriptor.PSObject.BaseObject) $foreign)
            }
            $row.After=SC-Entry $descriptor $owner
            BC-Need $row.MutationProven
            [P2BDescriptorFixture.CounterNote]::Reset()
            try{Assert-P2BCPlain $value (BC-Type $property.Value) (New-P2BCMetadata)}
            catch{if($_.Exception.Message -ceq 'P2BC_INVALID'){$row.Rejected=$true}else{throw}}
            $row.Counters=[P2BDescriptorFixture.CounterNote]::Total
            BC-Need ($row.Rejected -and $row.Counters -eq 0)
            $row.Validated=$true
        }
    }
    BC-Need ($record.KeyNegatives.Count -eq 19)
}
