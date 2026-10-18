# Controlled SDK identity observation only; no product import or fixture compile.
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$OutputEncoding=[Text.UTF8Encoding]::new($false)
$identity=[Security.Principal.WindowsIdentity]::GetCurrent()
$admin=([Security.Principal.WindowsPrincipal]::new($identity)).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
$record=[ordered]@{Schema=1;Mode='p2b-scalar-receiver-diagnostic';HostVersion=$PSVersionTable.PSVersion.ToString();Core=($PSVersionTable.PSEdition -ceq 'Core');X64=[Environment]::Is64BitProcess;Admin=$admin;Stage='Context';Captured=$false;Cases=@();References=@();Error=$null;ProductImported=$false;ProductCalled=$false;FixtureCompiled=$false;RuntimeApproved=$false}
$script:references=[object[]]::new(2048);$script:referenceCount=0
$script:typeMethod=[object].GetMethod('GetType',[type[]]@())
$script:invokeMethod=[Reflection.MethodBase].GetMethod('Invoke',[type[]]@([object],[object[]]))
$script:referenceMethod=[object].GetMethod('ReferenceEquals',[type[]]@([object],[object]))
$script:fieldReadMethod=[Reflection.FieldInfo].GetMethod('GetValue',[type[]]@([object]))
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
    for($n=0;$n -lt $script:referenceCount;$n++){if(Same $Value $script:references[$n]){return $n+1}}
    Need ($script:referenceCount -lt $script:references.Length)
    $script:references[$script:referenceCount]=$Value
    $script:referenceCount++
    $record.References+=,[pscustomobject]@{ID=$script:referenceCount;CLRType=(Raw-Type $Value).FullName}
    return $script:referenceCount
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
function New-Base([string]$Kind){
    $result=[object[]]::new(1)
    switch($Kind){
        'byte'{$a=[Array]::CreateInstance([byte],1);$a.SetValue([byte]7,0);$result[0]=$a.GetValue(0)}
        'bytes'{$result[0]=[byte[]]@(7,9)}
        'string'{$result[0]=[string]::new([char[]]@([char]112,[char]114,[char]111,[char]98,[char]101))}
        'custom'{
            $wrapper=$script:emptyCtor.Invoke([object[]]::new(0))
            $result[0]=(Read-Field $script:baseField $wrapper)[0]
        }
    }
    return ,$result
}
function New-Wrapper($Base,[string]$Kind){
    $result=[object[]]::new(1)
    if($Kind -ceq 'custom'){$result[0]=$script:emptyCtor.Invoke([object[]]::new(0))}
    else{$argsLocal=[object[]]::new(1);$argsLocal[0]=$Base;$result[0]=$script:objectCtor.Invoke($argsLocal)}
    return ,$result
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

# Original helper read order; traces are stored before returning, not reconstructed.
function Legacy-Raw($Value,$Phase){
    $set=$Value.psextended;$base=$set.PSObject.BaseObject
    $raw=(Read-Field $script:rawField $base)[0]
    $owner=(Read-Field $script:constructorOwnerField $base)[0]
    $Phase.LegacyRawOwner=Wrapper-State $owner
    $Phase.LegacyRawID=Ref-ID $raw
    $Phase.LegacySetID=Ref-ID $base
    $Phase.LegacyRawNames=@()
    foreach($entry in $raw){$Phase.LegacyRawNames+=,((Read-Field $script:nameField $entry)[0])}
    return ,$raw
}
function Legacy-Owner($Value,$Phase){
    $shim=$Value.PSObject
    $owner=(Read-Field $script:shimOwnerField $shim.PSObject.BaseObject)[0]
    $Phase.LegacyShimOwner=Wrapper-State $owner
    $a=[object[]]::new(1);$a[0]=$owner
    return ,$a
}
function Capture-Handle($Value){
    # SDK AsPSObject preserves an existing actual wrapper; otherwise creates one
    # at this explicit boundary. Both possibilities are recorded, never conflated.
    $handle=[object[]]::new(3)
    $handle[1]=$Value
    $a=[object[]]::new(1);$a[0]=$Value
    $handle[0]=$script:asObject.Invoke($null,$a)
    $handle[2]=(Read-Field $script:baseField $handle[0])[0]
    return ,$handle
}
function Pinned-Read([object[]]$Handle){
    # Never re-access .PSObject/.psextended on the scalar inside this helper.
    return Snapshot $Handle[0]
}
function New-HolderValue([string]$Kind){
    $base=(New-Base $Kind)[0]
    # Mirror the exact-note-property field path without importing old helpers.
    $holder=[pscustomobject][ordered]@{}
    $holder.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new('Data',$base))
    return $holder
}
try{
    Need (-not $admin -and $record.HostVersion -ceq '7.6.5' -and $record.Core -and $record.X64)
    $record.Stage='Metadata'
    $flags=[Reflection.BindingFlags]'NonPublic,Instance,DeclaredOnly'
    $sdk=[Management.Automation.PSObject].Assembly
    $script:shimType=$sdk.GetType('System.Management.Automation.PSInternalMemberSet',$true)
    $reserved=$sdk.GetType('System.Management.Automation.ReservedNameMembers',$true)
    $script:objectCtor=[Management.Automation.PSObject].GetConstructor([type[]]@([object]))
    $script:emptyCtor=[Management.Automation.PSObject].GetConstructor([type[]]@())
    $script:asObject=[Management.Automation.PSObject].GetMethod('AsPSObject',[type[]]@([object]))
    $script:baseField=[Management.Automation.PSObject].GetField('_immediateBaseObject',$flags)
    $script:tableField=[Management.Automation.PSObject].GetField('_instanceMembers',$flags)
    $script:flagsField=[Management.Automation.PSObject].GetField('_flags',$flags)
    $script:keyMethod=[Management.Automation.PSObject].GetMethod('GetKeyForResurrectionTables',[Reflection.BindingFlags]'NonPublic,Static,DeclaredOnly')
    $script:psobjectFactory=$reserved.GetMethod('GeneratePSObjectMemberSet',[Reflection.BindingFlags]'NonPublic,Static,DeclaredOnly')
    $script:extendedFactory=$reserved.GetMethod('GeneratePSExtendedMemberSet',[Reflection.BindingFlags]'NonPublic,Static,DeclaredOnly')
    $script:rawField=[Management.Automation.PSMemberSet].GetField('internalMembers',$flags)
    $script:ownerField=[Management.Automation.PSMemberInfo].GetField('instance',$flags)
    $script:nameField=[Management.Automation.PSMemberInfo].GetField('name',$flags)
    $script:hiddenField=[Management.Automation.PSMemberInfo].GetField('<IsHidden>k__BackingField',$flags)
    $script:reservedField=[Management.Automation.PSMemberInfo].GetField('<IsReservedMember>k__BackingField',$flags)
    $script:instanceField=[Management.Automation.PSMemberInfo].GetField('<IsInstance>k__BackingField',$flags)
    $script:shimOwnerField=$script:shimType.GetField('_psObject',$flags)
    $script:constructorOwnerField=[Management.Automation.PSMemberSet].GetField('_constructorPSObject',$flags)
    foreach($kind in @('string','byte','bytes')){
        foreach($source in @('variable','note-field')){
            $record.Stage=$kind+':'+$source
            $row=[pscustomobject][ordered]@{Number=($record.Cases.Count+1);Kind=$kind;Source=$source;InputWasWrapper=$null;CapturedInputRetained=$null;CapturedBaseRetained=$null;HandleID=$null;PinnedWrapperID=$null;Phases=@();Controls=@();Complete=$false}
            $record.Cases+=,$row
            # Separate trusted values for old/new paths; do not compare values
            # from different paths by reference or assert identical SDK history.
            $legacy=(New-Base $kind)[0];$next=(New-Base $kind)[0]
            if($source -ceq 'note-field'){
                $legacyHolder=New-HolderValue $kind;$nextHolder=New-HolderValue $kind
                $legacy=$legacyHolder.Data;$next=$nextHolder.Data
            }
            $handle=Capture-Handle $next
            $row.InputWasWrapper=((Raw-Type $handle[1]) -eq [Management.Automation.PSObject])
            $row.CapturedInputRetained=Same $handle[0] $handle[1]
            $row.CapturedBaseRetained=Same ((Read-Field $script:baseField $handle[0])[0]) $handle[2]
            $row.HandleID=Ref-ID $handle;$row.PinnedWrapperID=Ref-ID $handle[0]
            for($repeat=0;$repeat -lt 3;$repeat++){
                $phase=[pscustomobject][ordered]@{Repeat=$repeat;LegacyRawOwner=$null;LegacyShimOwner=$null;LegacyRawID=$null;LegacySetID=$null;LegacyRawNames=@();LegacyRawReturnedSame=$null;LegacyRawMatchesShimTable=$null;LegacyKeysMatch=$null;LegacyOwnerWrappersMatch=$null;Before=$null;Calls=@();After=$null;ReadAgain=$null;FixedPassed=$false}
                $row.Phases+=,$phase
                $raw=Legacy-Raw $legacy $phase;$owner=(Legacy-Owner $legacy $phase)[0]
                $phase.LegacyRawReturnedSame=(Ref-ID $raw) -eq $phase.LegacyRawID
                $phase.LegacyRawMatchesShimTable=$phase.LegacyRawID -eq $phase.LegacyShimOwner.TableID
                $phase.LegacyKeysMatch=$phase.LegacyRawOwner.KeyID -eq $phase.LegacyShimOwner.KeyID
                $phase.LegacyOwnerWrappersMatch=$phase.LegacyRawOwner.WrapperID -eq $phase.LegacyShimOwner.WrapperID
                $phase.Before=Pinned-Read $handle
                $phase.Calls=@((Generate 'psobject' $handle[0]),(Generate 'psextended' $handle[0]))
                $phase.After=Pinned-Read $handle
                $phase.ReadAgain=Pinned-Read $handle
                # Save all operands above BEFORE judging any relation.
                $a=$phase.After;$b=$phase.ReadAgain
                $fixed=$a.Wrapper.WrapperID -eq $row.PinnedWrapperID -and $b.Wrapper.WrapperID -eq $row.PinnedWrapperID -and $null -ne $a.Wrapper.TableID -and $a.Wrapper.TableID -eq $b.Wrapper.TableID -and $a.Descriptors.Count -eq 2 -and $b.Descriptors.Count -eq 2
                foreach($entry in $a.Descriptors){
                    $fixed=$fixed -and $entry.Hidden -and $entry.Reserved -and $entry.Instance -and $entry.Owner.KeyID -eq $a.Wrapper.KeyID -and $entry.Owner.TableID -eq $a.Wrapper.TableID
                    if($entry.Name -ceq 'psobject'){$fixed=$fixed -and $entry.CLRType -ceq 'System.Management.Automation.PSInternalMemberSet' -and $entry.InternalOwner.WrapperID -eq $entry.Owner.WrapperID}
                    elseif($entry.Name -ceq 'psextended'){$fixed=$fixed -and $entry.CLRType -ceq 'System.Management.Automation.PSMemberSet' -and $entry.ConstructorOwner.WrapperID -eq $entry.Owner.WrapperID -and $entry.ExtendedRawTableID -eq $a.Wrapper.TableID}
                    else{$fixed=$false}
                }
                $phase.FixedPassed=$fixed
                Need ($phase.LegacyRawReturnedSame -and $row.CapturedBaseRetained -and $fixed)
                if($row.InputWasWrapper){Need $row.CapturedInputRetained}
            }
            foreach($control in @('same-base-distinct-wrapper','equal-content-independent')){
                $base=$handle[2]
                if($control -ceq 'equal-content-independent'){$base=(New-Base $kind)[0]}
                $foreign=(New-Wrapper $base $kind)[0]
                $calls=@((Generate 'psobject' $foreign),(Generate 'psextended' $foreign))
                $state=Snapshot $foreign;$subject=Pinned-Read $handle
                $expected=($kind -cne 'string' -and $control -ceq 'same-base-distinct-wrapper')
                $entry=[pscustomobject]@{Scenario=$control;Calls=$calls;Subject=$subject;Foreign=$state;BaseSame=(Same $base $handle[2]);WrapperDifferent=(-not (Same $foreign $handle[0]));ExpectedKeyTableSame=$expected;ObservedKeySame=($state.Wrapper.KeyID -eq $subject.Wrapper.KeyID);ObservedTableSame=($state.Wrapper.TableID -eq $subject.Wrapper.TableID);Validated=$false}
                $row.Controls+=,$entry
                Need ($entry.WrapperDifferent -and $entry.BaseSame -eq ($control -ceq 'same-base-distinct-wrapper') -and $entry.ObservedKeySame -eq $expected -and $entry.ObservedTableSame -eq $expected)
                $entry.Validated=$true
            }
            $row.Complete=$true
        }
    }
    Need ($record.Cases.Count -eq 6)
    $record.Stage='Captured';$record.Captured=$true
}catch{
    $caught=$_;$chain=[Collections.Generic.List[object]]::new();$errorItem=$caught.Exception
    for($n=0;$null -ne $errorItem -and $n -lt 8;$n++){$chain.Add([pscustomobject]@{Type=$errorItem.GetType().FullName;Message=$errorItem.Message});$errorItem=$errorItem.InnerException}
    $record.Error=[pscustomobject]@{Chain=$chain.ToArray();Stack=$caught.ScriptStackTrace;Position=$caught.InvocationInfo.PositionMessage}
    $record.Captured=$true
}
$record|ConvertTo-Json -Depth 24
if($null -eq $record.Error -and $record.Stage -ceq 'Captured'){exit 0}else{exit 1}
