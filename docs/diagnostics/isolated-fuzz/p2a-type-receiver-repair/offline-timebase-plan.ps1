# Declared integers only. Function definitions have no load-time actions.
# Helpers are implementation details, not an isolation or authority boundary.
function New-TimebasePlanMetadata {
    [pscustomobject]@{
        TypeMethod=[object].GetMethod('GetType',[type[]]@())
        MembersProperty=[Management.Automation.PSMemberSet].GetProperty('Members')
        StringCharsMethod=[string].GetMethod('ToCharArray',[type[]]@())
        StringConstructor=[string].GetConstructor([type[]]@([char[]]))
    }
}
function Assert-TimebasePlanRequest([object]$Request,[object]$Metadata){
    if($null -eq $Request -or $Metadata.TypeMethod.Invoke($Request.PSObject.BaseObject,$null) -ne [Management.Automation.PSCustomObject]){throw 'P2A_INVALID'}
    $names=@('Schema','OriginTick','Frequency','ObservedTick','StopTick')
    $allowed=[Collections.Generic.HashSet[string]]::new([StringComparer]::Ordinal)
    foreach($name in $names){$null=$allowed.Add($name)}
    $properties=@($Request.PSObject.Properties)
    if($properties.Count -ne 5){throw 'P2A_INVALID'}
    # Finish descriptor checks BEFORE reading any value, including legal names.
    foreach($p in $properties){
        if(-not $allowed.Contains($p.Name) -or $p.MemberType -ne [Management.Automation.PSMemberTypes]::NoteProperty){throw 'P2A_INVALID'}
    }
    $extended=$Request.psextended
    foreach($member in $Metadata.MembersProperty.GetValue($extended.PSObject.BaseObject,$null)){
        if(-not $allowed.Contains($member.Name) -or $member.MemberType -ne [Management.Automation.PSMemberTypes]::NoteProperty){throw 'P2A_INVALID'}
    }
    $snapshot=[pscustomobject][ordered]@{}
    foreach($name in $names){
        $value=$Request.PSObject.Properties[$name].Value
        if($name -ceq 'StopTick' -and $null -eq $value){
            $snapshot.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new($name,$null));continue
        }
        if($null -eq $value){throw 'P2A_INVALID'}
        $type=$Metadata.TypeMethod.Invoke($value.PSObject.BaseObject,$null)
        $required=if($name -ceq 'Schema'){[int]}else{[long]}
        if($type -ne $required){throw 'P2A_INVALID'}
        $view=$value.psextended
        foreach($member in $Metadata.MembersProperty.GetValue($view.PSObject.BaseObject,$null)){throw 'P2A_INVALID'}
        $snapshot.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new($name,$value))
    }
    # All fields have exact passive types before value/relationship comparisons.
    if($snapshot.Schema -ne 1){throw 'P2A_INVALID'}
    if($snapshot.OriginTick -lt 0 -or $snapshot.Frequency -lt 1000 -or $snapshot.ObservedTick -lt $snapshot.OriginTick){throw 'P2A_RANGE'}
    if($null -ne $snapshot.StopTick -and ($snapshot.StopTick -lt $snapshot.OriginTick -or $snapshot.StopTick -gt $snapshot.ObservedTick)){throw 'P2A_RANGE'}
    return $snapshot
}
function Get-TimebasePlanDelta([long]$Frequency,[int]$Milliseconds,[bool]$Ceiling){
    if($Frequency -lt 1000 -or $Milliseconds -notin @(500,1000,5000,6000,6500,113500,120000)){throw 'P2A_RANGE'}
    $f=[Numerics.BigInteger]$Frequency;$ms=[Numerics.BigInteger]$Milliseconds
    $thousand=[Numerics.BigInteger]1000;$max=[Numerics.BigInteger]([long]::MaxValue)
    $q=[Numerics.BigInteger]::Divide($f,$thousand)
    $r=[Numerics.BigInteger]::Remainder($f,$thousand)
    $whole=[Numerics.BigInteger]::Multiply($q,$ms)
    $fraction=[Numerics.BigInteger]::Multiply($r,$ms)
    if($whole -lt [Numerics.BigInteger]::Zero -or $whole -gt $max -or $fraction -lt [Numerics.BigInteger]::Zero -or $fraction -gt $max){throw 'P2A_RANGE'}
    $result=[Numerics.BigInteger]::Add($whole,[Numerics.BigInteger]::Divide($fraction,$thousand))
    if($result -lt [Numerics.BigInteger]::Zero -or $result -gt $max){throw 'P2A_RANGE'}
    if($Ceiling -and [Numerics.BigInteger]::Remainder($fraction,$thousand) -ne [Numerics.BigInteger]::Zero){
        $result=[Numerics.BigInteger]::Add($result,[Numerics.BigInteger]::One)
        if($result -gt $max){throw 'P2A_RANGE'}
    }
    return [long]$result
}
function Add-TimebasePlanTicks([long]$Left,[long]$Right){
    if($Left -lt 0 -or $Right -lt 0){throw 'P2A_RANGE'}
    $sum=[Numerics.BigInteger]::Add([Numerics.BigInteger]$Left,[Numerics.BigInteger]$Right)
    if($sum -lt [Numerics.BigInteger]::Zero -or $sum -gt [Numerics.BigInteger]([long]::MaxValue)){throw 'P2A_RANGE'}
    return [long]$sum
}
function Subtract-TimebasePlanTicks([long]$Left,[long]$Right){
    if($Left -lt 0 -or $Right -lt 0){throw 'P2A_RANGE'}
    $difference=[Numerics.BigInteger]::Subtract([Numerics.BigInteger]$Left,[Numerics.BigInteger]$Right)
    if($difference -lt [Numerics.BigInteger]::Zero -or $difference -gt [Numerics.BigInteger]([long]::MaxValue)){throw 'P2A_RANGE'}
    return [long]$difference
}
function Copy-TimebasePlanScalar([object]$Value,[object]$Metadata){
    if($null -eq $Value){return $null}
    $base=$Value.PSObject.BaseObject;$type=$Metadata.TypeMethod.Invoke($base,$null)
    if($type -eq [int] -or $type -eq [long] -or $type -eq [bool]){
        $storage=[Array]::CreateInstance($type,1);$storage.SetValue($base,0);$fresh=$storage.GetValue(0)
    }elseif($type -eq [string]){
        if($base.Length -eq 0){throw 'Unexpected empty scalar'}
        $chars=$Metadata.StringCharsMethod.Invoke($base,$null)
        $parameters=[object[]]::new(1);$parameters[0]=$chars
        $fresh=$Metadata.StringConstructor.Invoke($parameters)
    }else{throw 'Unexpected scalar type'}
    if($Metadata.TypeMethod.Invoke($fresh.PSObject.BaseObject,$null) -ne $type -or [object]::ReferenceEquals($base,$fresh.PSObject.BaseObject)){throw 'Scalar storage not independent'}
    if($type -eq [string]){
        if(-not [StringComparer]::Ordinal.Equals($base,$fresh.PSObject.BaseObject)){throw 'Scalar value changed'}
    }elseif($base -cne $fresh.PSObject.BaseObject){throw 'Scalar value changed'}
    $view=$fresh.psextended
    foreach($member in $Metadata.MembersProperty.GetValue($view.PSObject.BaseObject,$null)){throw 'New scalar decorated'}
    return ,$fresh
}
function New-TimebasePlanResult([bool]$Created,[string]$Reason,[object]$Plan,[object]$Metadata){
    $source=[pscustomobject]@{Schema=[int]1;Kind='timebase-plan-only';Created=$Created;Reason=$Reason;Plan=$Plan;SourceVerified=$false;LiveActionsPerformed=$false;RuntimeApproved=$false}
    $owned=[pscustomobject][ordered]@{}
    foreach($name in @('Schema','Kind','Created','Reason','Plan','SourceVerified','LiveActionsPerformed','RuntimeApproved')){
        if($name -ceq 'Plan' -and $null -ne $Plan){
            $ownedPlan=[pscustomobject][ordered]@{}
            foreach($field in @('OriginTick','Frequency','DeadlineTick','StopAtTick','NaturalUntilTick','CleanupUntilTick','PublishUntilTick','ObservedDeadlineExpired')){
                $scalar=Copy-TimebasePlanScalar $Plan.PSObject.Properties[$field].Value $Metadata
                $ownedPlan.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new($field,$scalar))
            }
            $owned.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new($name,$ownedPlan))
        }else{
            $scalar=Copy-TimebasePlanScalar $source.PSObject.Properties[$name].Value $Metadata
            $owned.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new($name,$scalar))
        }
    }
    return $owned
}
function New-OfflineTimebasePlan([object]$Request){
    Set-StrictMode -Version Latest
    $metadata=New-TimebasePlanMetadata
    try{
        $snapshot=Assert-TimebasePlanRequest $Request $metadata
        $deadline=Add-TimebasePlanTicks $snapshot.OriginTick (Get-TimebasePlanDelta $snapshot.Frequency 120000 $false)
        $executionStop=Add-TimebasePlanTicks $snapshot.OriginTick (Get-TimebasePlanDelta $snapshot.Frequency 113500 $false)
        $reserveStop=Subtract-TimebasePlanTicks $deadline (Get-TimebasePlanDelta $snapshot.Frequency 6500 $true)
        $stopAt=[Math]::Min($executionStop,$reserveStop)
        $natural=$null;$cleanup=$null;$publish=$null
        if($null -ne $snapshot.StopTick){
            $s=[Math]::Min([long]$snapshot.StopTick,$deadline)
            # Compute/check BOTH branches before min; overflow cannot be hidden.
            $naturalCandidate=Add-TimebasePlanTicks $s (Get-TimebasePlanDelta $snapshot.Frequency 500 $false)
            $naturalBound=Subtract-TimebasePlanTicks $deadline (Get-TimebasePlanDelta $snapshot.Frequency 6000 $true)
            $natural=[Math]::Min($naturalCandidate,$naturalBound)
            $cleanupCandidate=Add-TimebasePlanTicks $natural (Get-TimebasePlanDelta $snapshot.Frequency 5000 $false)
            $cleanupBound=Subtract-TimebasePlanTicks $deadline (Get-TimebasePlanDelta $snapshot.Frequency 1000 $true)
            $cleanup=[Math]::Min($cleanupCandidate,$cleanupBound)
            $publishCandidate=Add-TimebasePlanTicks $cleanup (Get-TimebasePlanDelta $snapshot.Frequency 1000 $false)
            $publish=[Math]::Min($publishCandidate,$deadline)
        }
        $plan=[pscustomobject]@{
            OriginTick=$snapshot.OriginTick;Frequency=$snapshot.Frequency;DeadlineTick=$deadline;StopAtTick=$stopAt
            NaturalUntilTick=$natural;CleanupUntilTick=$cleanup;PublishUntilTick=$publish
            ObservedDeadlineExpired=($snapshot.ObservedTick -ge $deadline)
        }
        return New-TimebasePlanResult $true 'plan-created' $plan $metadata
    }catch{
        $reason=switch -CaseSensitive ($_.Exception.Message){'P2A_INVALID'{'invalid-input'} 'P2A_RANGE'{'range-refused'} default{'internal-error'}}
        return New-TimebasePlanResult $false $reason $null $metadata
    }
}
