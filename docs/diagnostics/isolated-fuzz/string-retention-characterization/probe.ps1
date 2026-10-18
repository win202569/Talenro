# Diagnostic only: no product/candidate imports or invocations.
# Distinguish ineffective field mutation from retained-view mutation.
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$script:case='startup'
function R-Need([bool]$Condition,[string]$Detail){if(-not $Condition){throw ('FIXTURE_INVALID|'+$script:case+'|'+$Detail)}}
# Fixed public SDK reflection. Argument arrays keep PSObject references from being
# debased as ordinary CLR object method arguments by PowerShell method binding.
$script:referenceMethod=[object].GetMethod('ReferenceEquals',[type[]]@([object],[object]))
$script:typeMethod=[object].GetMethod('GetType',[type[]]@())
$script:invokeMethod=[Reflection.MethodBase].GetMethod('Invoke',[type[]]@([object],[object[]]))
$script:stringConstructor=[string].GetConstructor([type[]]@([char[]]))
$script:wrapperConstructor=[Management.Automation.PSObject].GetConstructor([type[]]@([object]))
function R-Reference($Left,$Right){
    $arguments=[object[]]@($Left,$Right)
    return [bool]$script:referenceMethod.Invoke($null,$arguments)
}
function R-Type($Value){
    $arguments=[object[]]@($Value,$null)
    return $script:invokeMethod.Invoke($script:typeMethod,$arguments)
}
function R-String {
    $parameters=[object[]]::new(1);$parameters[0]=[char[]]@('c','l','e','a','n')
    return ,($script:stringConstructor.Invoke($parameters))
}
function R-Wrapper($Base){return ,($script:wrapperConstructor.Invoke([object[]]@($Base)))}
function R-Descriptor($Members,[string]$Name,$Payload){
    $p=$Members[$Name]
    if($null -eq $p){return $null}
    R-Need ($p.MemberType -eq [Management.Automation.PSMemberTypes]::NoteProperty -and $p.Name -ceq $Name) known-descriptor
    R-Need (R-Reference $p.Value $Payload) payload-identity
    return ,$p
}
try{
    $principal=[Security.Principal.WindowsPrincipal]::new([Security.Principal.WindowsIdentity]::GetCurrent())
    R-Need (-not $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) ordinary-host
    R-Need ($PSVersionTable.PSEdition -ceq 'Core' -and $PSVersionTable.PSVersion.ToString() -ceq '7.6.5' -and [Environment]::Is64BitProcess) fixed-engine
    Write-Output 'HOST: 7.6.5 Core x64=True admin=False productCalls=0'
    $specs=@(
        @{Name='held-string-view';Kind='held';Type=[string];Expected='clean';Control=$true},
        @{Name='raw-int32-field';Kind='int';Type=[int];Expected=[int]0;Control=$true},
        @{Name='raw-string-field';Kind='raw';Type=[string];Expected='clean';Control=$false},
        @{Name='wrapped-string-field';Kind='wrapped';Type=[string];Expected='clean';Control=$false}
    )
    foreach($spec in $specs){
        $script:case=$spec.Name
        $base=if($spec.Kind -ceq 'int'){$storage=[Array]::CreateInstance([int],1);$storage.GetValue(0)}else{R-String}
        if($spec.Kind -ceq 'held'){
            # Calibrate actual wrapper identity using two wrappers of ONE base,
            # not distinct strings whose bases would trivially compare unequal.
            $left=R-Wrapper $base;$right=R-Wrapper $base
            R-Need ((R-Type $left) -eq [Management.Automation.PSObject] -and (R-Type $right) -eq [Management.Automation.PSObject]) identity-wrapper-types
            R-Need ((R-Reference $left $left) -and -not (R-Reference $left $right)) identity-wrapper-calibration
            R-Need (R-Reference $left.BaseObject $right.BaseObject) identity-shared-base-calibration
        }
        $owner=[pscustomobject]@{Value=$null}
        if($spec.Kind -ceq 'wrapped'){$owner.Value=R-Wrapper $base}else{$owner.Value=$base}
        # Initial setup ends here. Never replace this owner or field again.
        $field=$owner.PSObject.Properties['Value']
        $storedBefore=R-Type $field.Value
        $beforeBase=$owner.Value.PSObject.BaseObject
        R-Need ((R-Type $beforeBase) -eq $spec.Type -and $beforeBase -ceq $spec.Expected) base-before
        $heldView=$owner.Value.PSObject;$heldMembers=$heldView.Properties
        $name='Retention_'+$spec.Name
        R-Need ($null -eq $heldMembers[$name]) initially-absent
        $payload=[object[]]@('initial')
        $heldMembers.Add([Management.Automation.PSNoteProperty]::new($name,$payload))
        $held=R-Descriptor $heldMembers $name $payload
        R-Need ($null -ne $held) held-descriptor-retained
        $readView=if($spec.Kind -ceq 'held'){$heldView}else{$owner.Value.PSObject}
        $readMembers=$readView.Properties
        $read=R-Descriptor $readMembers $name $payload
        $retained=$null -ne $read
        $viewSame=R-Reference $heldView $readView
        $membersSame=R-Reference $heldMembers $readMembers
        R-Need (-not ($membersSame -and -not $retained)) member-identity-consistency
        if($spec.Control){R-Need $retained positive-control-retention}
        $held.Value[0]='changed'
        $heldAfter=R-Descriptor $heldView.Properties $name $payload
        R-Need ($null -ne $heldAfter -and $heldAfter.Value[0] -ceq 'changed') held-mutation-retained
        $postView=if($spec.Kind -ceq 'held'){$heldView}else{$owner.Value.PSObject}
        $post=R-Descriptor $postView.Properties $name $payload
        R-Need (($null -ne $post) -eq $retained) consistent-retention-across-rereads
        if($retained){R-Need ($post.Value[0] -ceq 'changed') reread-mutation-retained}
        $fieldAfter=$owner.PSObject.Properties['Value'];$afterBase=$owner.Value.PSObject.BaseObject
        $baseSame=R-Reference $beforeBase $afterBase
        $fieldSame=R-Reference $field $fieldAfter
        $storedAfter=R-Type $fieldAfter.Value
        R-Need ($baseSame -and $fieldSame -and $storedBefore -eq $storedAfter -and (R-Type $afterBase) -eq $spec.Type -and $afterBase -ceq $spec.Expected) unchanged-field-base-type-value
        $facts=[pscustomobject][ordered]@{
            Case=$spec.Name;BaseType=$spec.Type.FullName;StoredTypeBefore=$storedBefore.FullName;StoredTypeAfter=$storedAfter.FullName
            ViewType=(R-Type $heldView).FullName;HeldViewSameOnReread=$viewSame;MemberCollectionSameOnReread=$membersSame
            BaseIdentitySame=$baseSame;FieldDescriptorSame=$fieldSame;HeldDescriptorRetained=$true;HeldPayloadSame=$true
            RereadDescriptorRetained=$retained;RereadPayloadSame=$(if($retained){$true}else{$null})
            HeldMutationRetained=$true;RereadMutationRetained=$(if($retained){$true}else{$null})
            PostViewSameAsReread=(R-Reference $readView $postView);Capability=$(if($retained){'RETAINED'}else{'NOT_RETAINED'})
            Control=$spec.Control;ProductCalls=0
        }
        # Serialize facts only, never the subject or wrappers being measured.
        Write-Output ('FACT: '+($facts|ConvertTo-Json -Compress))
    }
    Write-Output 'RESULT: DIAGNOSTIC_COMPLETE cases=4 productCalls=0 acceptance=False'
}catch{
    $message=$_.Exception.Message
    if($message -cmatch '^FIXTURE_INVALID\|[a-z0-9-]+\|[a-z0-9-]+$'){
        $parts=$message -split '\|',3
        Write-Output ('RESULT: FIXTURE_INVALID case='+$parts[1]+' detail='+$parts[2]+' productCalls=0')
        exit 1
    }
    Write-Error $_;exit 1
}
