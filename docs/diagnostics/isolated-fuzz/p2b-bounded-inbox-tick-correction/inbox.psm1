# Fixed dependency and private conventional state; no live adapter.
Set-StrictMode -Version Latest
$dependency=Join-Path $PSScriptRoot '../p2b-cold-cache-repair/control-frame.ps1'
if((Microsoft.PowerShell.Utility\Get-FileHash -LiteralPath $dependency -Algorithm SHA256).Hash -cne '9A532E17D33CAB1FEE9D8BB24C41D5A4B867E42F5EBAB0EE3C3C29C9CDC1093C'){throw 'Fixed codec identity mismatch'}
. $dependency
$script:metadata=New-P2BCMetadata
$script:registry=[Runtime.CompilerServices.ConditionalWeakTable[object,object]]::new()
function Copy-IBTree($Value){
    if($null -eq $Value){return $null}
    $type=Get-P2BCRawType $Value $script:metadata
    if($type -eq [Management.Automation.PSCustomObject]){
        $result=[pscustomobject][ordered]@{}
        foreach($p in $Value.PSObject.Properties){$v=Copy-IBTree $p.Value;$result.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new($p.Name,$v))}
        return $result
    }
    if($type -eq [object[]]){
        $result=[object[]]::new($Value.GetLength(0))
        for($i=0;$i -lt $result.GetLength(0);$i++){$result[$i]=Copy-IBTree $Value[$i]}
        return ,$result
    }
    return ,(Copy-P2BCScalar $Value $script:metadata)
}
function New-IBResult([bool]$Accepted,[string]$Reason,$Token,$Endpoint,[int]$AcceptedBytes,$Frame,$Snapshot){
    $values=[ordered]@{Schema=[int]1;Mode='p2b-inbox-only';Accepted=$Accepted;Reason=$Reason;Inbox=$Token;EndpointId=$Endpoint;AcceptedBytes=$AcceptedBytes;Frame=$Frame;Snapshot=$Snapshot;SourceVerified=$false;LiveActionsPerformed=$false;RuntimeApproved=$false}
    $out=[pscustomobject][ordered]@{}
    foreach($name in $values.Keys){
        $value=$null
        # Token is intentionally an identity, not an owned data snapshot.
        if($name -ceq 'Inbox'){$value=$Token}else{$value=Copy-IBTree $values[$name]}
        $out.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new($name,$value))
    }
    return $out
}
function Get-IBState($Token){
    if($null -eq $Token){return $null}
    $current=$Token;$type=Get-P2BCRawType $current $script:metadata
    for($i=0;$type -eq [Management.Automation.PSObject] -and $i -lt 8;$i++){
        $current=(Read-P2BCField $script:metadata.WrapperBaseField $current $script:metadata)[0]
        if($null -eq $current){return $null}
        $type=Get-P2BCRawType $current $script:metadata
    }
    if($type -ne [object]){return $null}
    $state=[object]$null
    if(-not $script:registry.TryGetValue($current,[ref]$state)){return $null}
    return $state
}
function Get-IBEndpointId($Value){
    Assert-P2BCPlain $Value ([byte]) $script:metadata
    $copy=Copy-P2BCScalar $Value $script:metadata
    if($copy -lt 1 -or $copy -gt 4){throw 'P2BC_INVALID'}
    return $copy
}
function Stop-IBState($State,[string]$Reason,$Endpoint,[string]$Stage,$Sequence){
    if(-not $State.Stopped){$State.FirstFailure=[pscustomobject][ordered]@{Reason=$Reason;EndpointId=$Endpoint;Stage=$Stage;Sequence=$Sequence};$State.Stopped=$true}
}
function New-IBEndpoint {
    $slots=[object[]]::new(4);for($i=0;$i -lt 4;$i++){$slots[$i]=[byte[]]::new(256)}
    [pscustomobject]@{Slots=$slots;Head=[int]0;Count=[int]0;NextSequence=[uint64]1;LastNonHeartbeatSequence=[uint64]0;Partial=[byte[]]::new(256);PartialLength=[int]0;Heartbeat=[byte[]]::new(256);HasHeartbeat=$false}
}
function Get-IBBinding([object]$Value,[object]$Metadata){
    if($null -eq $Value -or $Metadata.TypeMethod.Invoke($Value.PSObject.BaseObject,$null) -ne [Management.Automation.PSCustomObject]){throw 'P2BC_INVALID'}
    $shape=[ordered]@{Schema=[int];BindingDigest=[byte[]];SessionId=[byte[]];OriginTick=[long];Frequency=[long];DeadlineTick=[long]}
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
    if($snapshot.Schema -ne 1){throw 'P2BC_INVALID'}
    $null=Get-P2BCDeclaredBounds $snapshot.OriginTick $snapshot.Frequency $snapshot.DeadlineTick
    return $snapshot
}

function New-P2BInbox([object]$Binding){
    try{
        $bindingCopy=Get-IBBinding $Binding $script:metadata
        $endpoints=[object[]]::new(4);for($i=0;$i -lt 4;$i++){$endpoints[$i]=New-IBEndpoint}
        $state=[pscustomobject]@{Stopped=$false;FirstFailure=$null;Binding=$bindingCopy;Endpoints=$endpoints}
        $token=[object]::new();$script:registry.Add($token,$state)
        return New-IBResult $true 'created' $token $null 0 $null $null
    }catch{
        $reason=if($_.Exception.Message -ceq 'P2BC_INVALID'){'invalid-input'}else{'internal-error'}
        return New-IBResult $false $reason $null $null 0 $null $null
    }
}
function Add-P2BInboxFragment([object]$Inbox,[object]$EndpointId,[object]$Bytes){
    $state=$null;$id=$null;$accepted=0;$sequence=$null;$stage='input'
    try{
        $state=Get-IBState $Inbox
        if($null -eq $state){return New-IBResult $false 'invalid-input' $null $null 0 $null $null}
        if($state.Stopped){return New-IBResult $false 'stopped' $null $null 0 $null $null}
        $id=Get-IBEndpointId $EndpointId;$endpoint=$state.Endpoints[[int]$id-1]
        $chunk=Get-P2BCByteSnapshot $Bytes 1 (256-$endpoint.PartialLength) $script:metadata
        $length=$chunk.GetLength(0);[Array]::Copy($chunk,0,$endpoint.Partial,$endpoint.PartialLength,$length)
        $endpoint.PartialLength+=$length;$accepted=$length
        if($endpoint.PartialLength -lt 256){return New-IBResult $true 'partial' $null $id $accepted $null $null}
        $stage='frame';$decoded=ConvertFrom-P2BControlFrame $endpoint.Partial
        if(-not $decoded.Accepted){
            $why=if($decoded.Reason -ceq 'internal-error'){'internal-error'}else{'invalid-frame'}
            Stop-IBState $state $why $id $stage $null
            return New-IBResult $false $why $null $id $accepted $null $null
        }
        $message=$decoded.Message;$sequence=$message.Sequence;$stage='binding'
        $pairs=@(@(1,2),@(2,1),@(2,3),@(3,2));$pair=$pairs[[int]$id-1];$b=$state.Binding
        if($message.Sender -ne $pair[0] -or $message.Receiver -ne $pair[1] -or
           -not (Test-P2BCBytesEqual $message.BindingDigest $b.BindingDigest) -or
           -not (Test-P2BCBytesEqual $message.SessionId $b.SessionId) -or
           $message.OriginTick -ne $b.OriginTick -or $message.Frequency -ne $b.Frequency -or $message.DeadlineTick -ne $b.DeadlineTick){
            Stop-IBState $state 'binding-mismatch' $id $stage $sequence
            return New-IBResult $false 'binding-mismatch' $null $id $accepted $null $null
        }
        $stage='sequence'
        if($message.Sequence -ne $endpoint.NextSequence){
            Stop-IBState $state 'sequence-mismatch' $id $stage $sequence
            return New-IBResult $false 'sequence-mismatch' $null $id $accepted $null $null
        }
        if($message.Sequence -eq [uint64]::MaxValue){
            Stop-IBState $state 'sequence-exhausted' $id $stage $sequence
            return New-IBResult $false 'sequence-exhausted' $null $id $accepted $null $null
        }
        $heartbeat=$message.Kind -eq 8
        if($heartbeat){
            $stage='reference';$last=Read-P2BCUInt $message.Payload 8 8
            if($last -ne $endpoint.LastNonHeartbeatSequence){
                Stop-IBState $state 'reference-mismatch' $id $stage $sequence
                return New-IBResult $false 'reference-mismatch' $null $id $accepted $null $null
            }
        }else{
            $stage='capacity'
            if($endpoint.Count -ge 4){
                Stop-IBState $state 'capacity' $id $stage $sequence
                return New-IBResult $false 'capacity' $null $id $accepted $null $null
            }
        }
        $stage='commit'
        if($heartbeat){
            [Array]::Copy($endpoint.Partial,$endpoint.Heartbeat,256);$endpoint.HasHeartbeat=$true;$reason='heartbeat-updated'
        }else{
            $tail=($endpoint.Head+$endpoint.Count)%4
            [Array]::Copy($endpoint.Partial,$endpoint.Slots[$tail],256);$endpoint.Count++
            $endpoint.LastNonHeartbeatSequence=$sequence;$reason='queued'
        }
        $endpoint.NextSequence=[uint64][Numerics.BigInteger]::Add([Numerics.BigInteger]$sequence,[Numerics.BigInteger]::One)
        [Array]::Clear($endpoint.Partial,0,256);$endpoint.PartialLength=0
        return New-IBResult $true $reason $null $id $accepted $null $null
    }catch{
        $reason=if($_.Exception.Message -ceq 'P2BC_INVALID'){'invalid-input'}else{'internal-error'}
        if($reason -ceq 'internal-error'){$stage='commit'}
        if($null -ne $state){Stop-IBState $state $reason $id $stage $sequence}
        return New-IBResult $false $reason $null $id $accepted $null $null
    }
}
function Close-P2BInboxInput([object]$Inbox,[object]$EndpointId){
    $state=$null;$id=$null
    try{
        $state=Get-IBState $Inbox
        if($null -eq $state){return New-IBResult $false 'invalid-input' $null $null 0 $null $null}
        if($state.Stopped){return New-IBResult $false 'stopped' $null $null 0 $null $null}
        $id=Get-IBEndpointId $EndpointId
        $reason=if($state.Endpoints[[int]$id-1].PartialLength -ne 0){'truncated-frame'}else{'peer-closed'}
        Stop-IBState $state $reason $id 'eof' $null
        return New-IBResult $false $reason $null $id 0 $null $null
    }catch{
        $reason=if($_.Exception.Message -ceq 'P2BC_INVALID'){'invalid-input'}else{'internal-error'}
        if($null -ne $state){Stop-IBState $state $reason $id 'input' $null}
        return New-IBResult $false $reason $null $id 0 $null $null
    }
}
function Receive-P2BInboxFrame([object]$Inbox,[object]$EndpointId){
    $state=$null;$id=$null;$stage='input'
    try{
        $state=Get-IBState $Inbox
        if($null -eq $state){return New-IBResult $false 'invalid-input' $null $null 0 $null $null}
        if($state.Stopped){return New-IBResult $false 'stopped' $null $null 0 $null $null}
        $id=Get-IBEndpointId $EndpointId;$e=$state.Endpoints[[int]$id-1]
        if($e.Count -eq 0){return New-IBResult $true 'empty' $null $id 0 $null $null}
        $stage='commit';$frame=[byte[]]::new(256);[Array]::Copy($e.Slots[$e.Head],$frame,256)
        [Array]::Clear($e.Slots[$e.Head],0,256);$e.Head=($e.Head+1)%4;$e.Count--
        return New-IBResult $true 'dequeued' $null $id 0 $frame $null
    }catch{
        $reason=if($_.Exception.Message -ceq 'P2BC_INVALID'){'invalid-input'}else{'internal-error'}
        if($null -ne $state){Stop-IBState $state $reason $id $stage $null}
        return New-IBResult $false $reason $null $id 0 $null $null
    }
}
function Get-P2BInboxSnapshot([object]$Inbox){
    $state=$null
    try{
        $state=Get-IBState $Inbox
        if($null -eq $state){return New-IBResult $false 'invalid-input' $null $null 0 $null $null}
        $endpoints=[object[]]::new(4)
        for($i=0;$i -lt 4;$i++){
            $e=$state.Endpoints[$i];$q=[object[]]::new($e.Count)
            for($j=0;$j -lt $e.Count;$j++){$q[$j]=[byte[]]::new(256);[Array]::Copy($e.Slots[($e.Head+$j)%4],$q[$j],256)}
            $partial=[byte[]]::new($e.PartialLength);[Array]::Copy($e.Partial,$partial,$e.PartialLength)
            $hb=$null;if($e.HasHeartbeat){$hb=[byte[]]::new(256);[Array]::Copy($e.Heartbeat,$hb,256)}
            $endpoints[$i]=[pscustomobject][ordered]@{EndpointId=[byte]($i+1);Count=[int]$e.Count;NextSequence=[uint64]$e.NextSequence;LastNonHeartbeatSequence=[uint64]$e.LastNonHeartbeatSequence;PartialBytes=$partial;Queue=$q;Heartbeat=$hb;PartialLength=[int]$e.PartialLength}
        }
        $snapshot=[pscustomobject][ordered]@{Stopped=[bool]$state.Stopped;FirstFailure=$state.FirstFailure;Binding=$state.Binding;Endpoints=$endpoints}
        return New-IBResult $true 'inspected' $null $null 0 $null $snapshot
    }catch{
        if($null -ne $state){Stop-IBState $state 'internal-error' $null 'commit' $null}
        return New-IBResult $false 'internal-error' $null $null 0 $null $null
    }
}
Export-ModuleMember -Function New-P2BInbox,Add-P2BInboxFragment,Close-P2BInboxInput,Receive-P2BInboxFrame,Get-P2BInboxSnapshot
