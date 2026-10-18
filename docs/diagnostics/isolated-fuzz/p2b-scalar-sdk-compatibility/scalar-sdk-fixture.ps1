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
function SC-ProveValue($Value,[string]$Label,[int]$History){
    $proof=[pscustomobject][ordered]@{Label=$Label;History=$History;ValueType=$null;OwnerType=$null;Entries=@();SDKCount=$null;ApplicationCount=$null;Counters=$null;Validated=$false}
    $record.Proofs+=,$proof;$record.Stage='SDKProof:'+ $Label
    $proof.ValueType=(BC-Type $Value).FullName
    for($n=0;$n -lt $History;$n++){DG-FullHistory $Value}
    $raw=DG-RawMembers $Value;$owner=(SC-Owner $Value)[0]
    $proof.OwnerType=(DG-RawType $owner).FullName
    foreach($entry in $raw){$proof.Entries+=,(SC-Entry $entry $owner)}
    # All observations above survive a failure in the independent fixture checker.
    $applications=DG-ApplicationMembers $Value $raw
    $proof.SDKCount=$script:memberSetEvidence.SDKCount;$proof.ApplicationCount=$applications.Count
    $proof.Counters=[P2BDescriptorFixture.CounterNote]::Total
    BC-Need ($applications.Count -eq 0 -and $proof.Counters -eq 0)
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
