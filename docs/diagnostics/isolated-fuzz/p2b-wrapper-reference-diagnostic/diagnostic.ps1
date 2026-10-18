# Controlled SDK identity observation only; no product import or fixture compile.
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$OutputEncoding=[Text.UTF8Encoding]::new($false)
$identity=[Security.Principal.WindowsIdentity]::GetCurrent()
$admin=([Security.Principal.WindowsPrincipal]::new($identity)).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
$record=[ordered]@{Schema=1;Mode='p2b-wrapper-reference-diagnostic';HostVersion=$PSVersionTable.PSVersion.ToString();Core=($PSVersionTable.PSEdition -ceq 'Core');X64=[Environment]::Is64BitProcess;Admin=$admin;Stage='Context';Captured=$false;Cases=@();References=@();Error=$null;ProductImported=$false;ProductCalled=$false;FixtureCompiled=$false;RuntimeApproved=$false}
$script:references=[object[]]::new(512);$script:referenceCount=0
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
try{
    Need (-not $admin -and $record.HostVersion -ceq '7.6.5' -and $record.Core -and $record.X64)
    $record.Stage='Metadata'
    $flags=[Reflection.BindingFlags]'NonPublic,Instance,DeclaredOnly'
    $sdk=[Management.Automation.PSObject].Assembly
    $script:shimType=$sdk.GetType('System.Management.Automation.PSInternalMemberSet',$true)
    $reserved=$sdk.GetType('System.Management.Automation.ReservedNameMembers',$true)
    $script:objectCtor=[Management.Automation.PSObject].GetConstructor([type[]]@([object]))
    $script:emptyCtor=[Management.Automation.PSObject].GetConstructor([type[]]@())
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
    foreach($kind in @('byte','bytes','string','custom')){
        foreach($scenario in @('same-wrapper','same-base-distinct-wrappers','equal-content-independent')){
            $record.Stage=$kind+':'+$scenario
            $row=[pscustomobject][ordered]@{Number=($record.Cases.Count+1);Kind=$kind;Scenario=$scenario;Construction=$null;Phases=@();Complete=$false}
            $record.Cases+=,$row
            $baseLeft=(New-Base $kind)[0]
            $baseRight=$baseLeft
            if($scenario -ceq 'equal-content-independent'){$baseRight=(New-Base $kind)[0]}
            $left=(New-Wrapper $baseLeft $kind)[0]
            $right=$left
            if($scenario -cne 'same-wrapper'){$right=(New-Wrapper $baseRight $kind)[0]}
            $construction=[ordered]@{LeftExactWrapper=((Raw-Type $left) -eq [Management.Automation.PSObject]);RightExactWrapper=((Raw-Type $right) -eq [Management.Automation.PSObject]);WrapperIdentityAsConstructed=((Same $left $right) -eq ($scenario -ceq 'same-wrapper'));RequestedBaseLeftID=(Ref-ID $baseLeft);RequestedBaseRightID=(Ref-ID $baseRight);IndependentBaseRequired=($kind -cne 'custom' -and $scenario -ceq 'equal-content-independent');IndependentBaseObserved=(-not (Same $baseLeft $baseRight));ContentLiteral=if($kind -ceq 'byte'){'7'}elseif($kind -ceq 'bytes'){'7,9'}elseif($kind -ceq 'string'){'probe'}else{'empty custom objects; singleton base may be shared'}}
            $row.Construction=[pscustomobject]$construction
            Need ($construction.LeftExactWrapper -and $construction.RightExactWrapper -and $construction.WrapperIdentityAsConstructed)
            if($construction.IndependentBaseRequired){Need $construction.IndependentBaseObserved}
            $calls=@((Generate 'psobject' $left),(Generate 'psextended' $right))
            Observe-Phase $row 'split-generation' $left $right $calls
            $calls=@((Generate 'psextended' $left),(Generate 'psobject' $right))
            Observe-Phase $row 'complete-views' $left $right $calls
            $calls=@((Generate 'psobject' $left),(Generate 'psextended' $left),(Generate 'psobject' $right),(Generate 'psextended' $right))
            Observe-Phase $row 'reuse-views' $left $right $calls
            $row.Complete=$true
        }
    }
    Need ($record.Cases.Count -eq 12)
    $record.Stage='Captured';$record.Captured=$true
}catch{
    $caught=$_;$chain=[Collections.Generic.List[object]]::new();$errorItem=$caught.Exception
    for($n=0;$null -ne $errorItem -and $n -lt 8;$n++){$chain.Add([pscustomobject]@{Type=$errorItem.GetType().FullName;Message=$errorItem.Message});$errorItem=$errorItem.InnerException}
    $record.Error=[pscustomobject]@{Chain=$chain.ToArray();Stack=$caught.ScriptStackTrace;Position=$caught.InvocationInfo.PositionMessage}
    $record.Captured=$true
}
$record|ConvertTo-Json -Depth 20
if($null -eq $record.Error -and $record.Stage -ceq 'Captured'){exit 0}else{exit 1}
