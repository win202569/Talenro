# Cold-cache regression oracle. No product helper supplies construction expectations.
function CT-Init {
    $flags=[Reflection.BindingFlags]'NonPublic,Static,DeclaredOnly'
    $script:ctTableField=[Management.Automation.PSObject].GetField('s_instanceMembersResurrectionTable',$flags)
    Need ($null -ne $script:ctTableField -and $script:ctTableField.IsStatic -and $script:ctTableField.IsInitOnly)
    $script:ctTry=$script:ctTableField.FieldType.GetMethod('TryGetValue')
    $script:ctLocal=[Management.Automation.PSObject].GetProperty('StoreTypeNameAndInstanceMembersLocally',[Reflection.BindingFlags]'NonPublic,Instance,DeclaredOnly')
    $script:ctSet=[Reflection.PropertyInfo].GetMethod('SetValue',[type[]]@([object],[object],[object[]]))
}
function CT-Local([object[]]$Source,[bool]$Local){
    $a=[object[]]::new(3);$a[0]=$Source[0];$a[1]=$Local;$a[2]=$null
    $null=$script:ctSet.Invoke($script:ctLocal,$a)
}
function CT-Key([object[]]$Source){
    $a=[object[]]::new(1);$a[0]=$Source[0]
    $out=[object[]]::new(1);$out[0]=$script:keyMethod.Invoke($null,$a)
    return ,$out
}
function CT-Lookup($Key){
    $a=[object[]]::new(2);$a[0]=$Key;$a[1]=$null
    $table=$script:ctTableField.GetValue($null)
    $out=[object[]]::new(2);$out[0]=[bool]$script:ctTry.Invoke($table,$a);$out[1]=$a[1]
    return ,$out
}
function CT-TagObservation($Snapshot){
    $hits=@($Snapshot.Descriptors|Where-Object{$_.Name -ceq 'CacheAttachedProbe'})
    if($hits.Count -eq 0){return $null}
    Need ($hits.Count -eq 1)
    $e=$hits[0]
    return [pscustomobject]@{ID=$e.ID;Name=$e.Name;CLRType=$e.CLRType;Hidden=$e.Hidden;Reserved=$e.Reserved;Instance=$e.Instance}
}
function CT-Case([string]$Kind,[string]$Scenario,[string]$Variant,[bool]$Local,[bool]$Public){
    $record.Stage='ColdCache:'+ $Kind+':'+$Scenario+':'+$Variant+':'+$Local+':'+$Public
    $own=$Scenario -ceq 'own-dirty'
    $dirty=$Scenario -cin @('shared-dirty','own-dirty')
    $shared= -not $Local -and $Kind -cne 'String'
    $peerViews=$Scenario -cne 'absent-clean'
    $expectFound=$peerViews -and ($own -or $shared)
    $expectAccepted=-not $dirty
    $row=[pscustomobject][ordered]@{N=($record.CacheCases.Count+1);Kind=$Kind;Scenario=$Scenario;Variant=$Variant;Local=$Local;Public=$Public;Phase='Construct';ExpectedFound=$expectFound;ExpectedAccepted=$expectAccepted;ExpectedBaselineMiss=($Scenario -ceq 'shared-dirty');TagID=0;PeerBefore=$null;SubjectBefore=$null;SubjectAfter=$null;PeerAfter=$null;KeyIDs=$null;AuthoritativeTableID=$null;Preconditions=$null;PreconditionsPassed=$false;TagBefore=$null;TagAfter=$null;Result=$null;CountersBefore=$null;CountersAfter=$null;DirectCacheRetained=$null;Matched=$null;Completed=$false}
    $record.CacheCases+=,$row
    [P2BDescriptorFixture.CounterNote]::Reset()
    $peer=New-TestSource $Kind
    if($Public){
        # Valid literal frame base is placed in actual wrappers without adapting it.
        $a=[object[]]::new(1);$a[0]=BC-Bytes 2
        $peer[0]=$script:objectCtor.Invoke($a)
    }
    CT-Local $peer $Local
    if($peerViews){BT-Views $peer 0}
    $tag=$null
    if($Variant -cne 'clean'){
        if($Variant -clike '*subclass'){$tag=[P2BDescriptorFixture.CounterNote]::new('CacheAttachedProbe',[int]42)}
        else{$tag=[Management.Automation.PSNoteProperty]::new('CacheAttachedProbe',[int]42)}
        if($Variant -clike 'hidden-*'){$script:hiddenField.SetValue($tag,$true)}
        $raw=(Read-Field $script:tableField $peer[0])[0]
        $method=$script:rawField.FieldType.GetMethod('Add',[type[]]@([Management.Automation.PSMemberInfo]))
        $a=[object[]]::new(1);$a[0]=$tag;$null=$method.Invoke($raw,$a)
        $row.TagID=Ref-ID $tag
    }
    $subject=[object[]]::new(1)
    if($own){$subject[0]=$peer[0]}
    else{
        $a=[object[]]::new(1);$a[0]=(Read-Field $script:baseField $peer[0])[0]
        $subject[0]=$script:objectCtor.Invoke($a)
        CT-Local $subject $Local
    }
    # Fixed fields/key oracle only: do not use subject.PSObject/InstanceMembers
    # or any SDK view factory on the COLD subject before the API call.
    $initialCache=(Read-Field $script:tableField $subject[0])[0]
    $peerKey=(CT-Key $peer)[0];$subjectKey=(CT-Key $subject)[0]
    $lookup=CT-Lookup $subjectKey
    $row.PeerBefore=Snapshot $peer[0];$row.SubjectBefore=Snapshot $subject[0]
    $row.TagBefore=CT-TagObservation $row.PeerBefore
    $row.KeyIDs=[object[]]@((Ref-ID $peerKey),(Ref-ID $subjectKey))
    $row.AuthoritativeTableID=Ref-ID $lookup[1]
    $required=Raw-Type ((Read-Field $script:baseField $subject[0])[0])
    $tagOK=$Variant -ceq 'clean'
    if($Variant -cne 'clean'){
        $want=if($Variant -clike '*subclass'){'P2BDescriptorFixture.CounterNote'}else{'System.Management.Automation.PSNoteProperty'}
        $tagOK=$null -ne $row.TagBefore -and $row.TagBefore.ID -eq $row.TagID -and $row.TagBefore.CLRType -ceq $want -and $row.TagBefore.Hidden -eq ($Variant -clike 'hidden-*') -and -not $row.TagBefore.Reserved -and $row.TagBefore.Instance
    }
    $cacheAfterObserving=(Read-Field $script:tableField $subject[0])[0]
    $row.CountersBefore=[P2BDescriptorFixture.CounterNote]::Total
    $row.Preconditions=[pscustomobject]@{
        ExactReceiver=((Raw-Type $subject[0]) -eq [Management.Automation.PSObject])
        CacheStateExpected=(($null -eq $initialCache) -eq (-not $own))
        ObservationDidNotMaterialize=(Same $initialCache $cacheAfterObserving)
        KeySharingExpected=((Same $peerKey $subjectKey) -eq ($own -or $shared))
        TablePresenceExpected=([bool]$lookup[0] -eq $expectFound)
        TableReferenceExpected=($(if($expectFound){Same $lookup[1] ((Read-Field $script:tableField $peer[0])[0])}else{$null -eq $lookup[1]}))
        TagIdentityClassFlags=$tagOK
        FixtureCallbacksZero=($row.CountersBefore -eq 0)
    }
    $row.Phase='PreconditionsSaved'
    foreach($condition in $row.Preconditions.PSObject.Properties){Need ([bool]$condition.Value)}
    $row.PreconditionsPassed=$true
    $row.Phase='Product'
    $record.ProductCalled=$true
    if($Public){$row.Result=SP-Result (ConvertFrom-P2BControlFrame $subject[0])}
    else{
        $accepted=$true;$reason='plain'
        try{Assert-P2BCPlain $subject[0] $required (New-P2BCMetadata)}
        catch{if($_.Exception.Message -ceq 'P2BC_INVALID'){$accepted=$false;$reason='invalid-input'}else{throw}}
        $row.Result=[pscustomobject]@{Accepted=$accepted;Reason=$reason;SourceVerified=$false;LiveActionsPerformed=$false;RuntimeApproved=$false}
    }
    $row.CountersAfter=[P2BDescriptorFixture.CounterNote]::Total
    $row.SubjectAfter=Snapshot $subject[0];$row.PeerAfter=Snapshot $peer[0]
    $row.TagAfter=CT-TagObservation $row.PeerAfter
    if(-not $Public){$row.DirectCacheRetained=Same $initialCache ((Read-Field $script:tableField $subject[0])[0])}
    $row.Matched=$row.Result.Accepted -eq $expectAccepted
    $row.Phase='PostconditionsSaved'
    Need ($row.CountersAfter -eq 0 -and -not $row.Result.SourceVerified -and -not $row.Result.LiveActionsPerformed -and -not $row.Result.RuntimeApproved)
    if(-not $Public){Need $row.DirectCacheRetained}
    if($Variant -cne 'clean'){Need ($row.TagAfter.ID -eq $row.TagID -and $row.TagAfter.Hidden -eq $row.TagBefore.Hidden -and $row.TagAfter.CLRType -ceq $row.TagBefore.CLRType)}
    if($row.Result.Accepted){Need ($row.Result.Reason -ceq $(if($Public){'decoded'}else{'plain'}))}
    else{Need ($row.Result.Reason -ceq 'invalid-input')}
    if(-not $row.ExpectedBaselineMiss){Need $row.Matched}
    $row.Phase='Completed';$row.Completed=$true
}
function CT-Run {
    CT-Init
    foreach($kind in @('Byte','UInt16','UInt32','UInt64','Int32','Int64','Boolean','String','Bytes')){
        foreach($local in @($false,$true)){CT-Case $kind 'absent-clean' 'clean' $local $false}
        if($kind -cne 'String'){
            CT-Case $kind 'shared-clean' 'clean' $false $false
            foreach($variant in @('visible-note','hidden-note','visible-subclass','hidden-subclass')){CT-Case $kind 'shared-dirty' $variant $false $false}
        }
        CT-Case $kind 'separate-peer' 'hidden-note' $true $false
        CT-Case $kind 'own-dirty' 'hidden-note' $true $false
    }
    CT-Case 'String' 'separate-peer' 'hidden-note' $false $false
    CT-Case 'String' 'own-dirty' 'hidden-note' $false $false
    CT-Case 'Bytes' 'absent-clean' 'clean' $false $true
    CT-Case 'Bytes' 'shared-clean' 'clean' $false $true
    foreach($variant in @('visible-note','hidden-note','visible-subclass','hidden-subclass')){CT-Case 'Bytes' 'shared-dirty' $variant $false $true}
    Need ($record.CacheCases.Count -eq 84)
    $record.CacheTestsCompleted=$true
    $record.Stage='UnmaterializedSharedDecorationRefusal'
    foreach($case in $record.CacheCases){if(-not $case.Matched){throw 'UnmaterializedSharedDecorationRefusal'}}
}
