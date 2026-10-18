# PREPARED ONLY / NOT RUNTIME VERIFIED. No execution allowance granted.
# View-aware bounded regression: scalar copying removed/shallow sharing must expose contamination.
# Observer controls prove fixture visibility, NOT product-mutant sensitivity.
# Removing owned-return copying must expose later descriptor contamination.
# A shallow result sharing a later-decorated caller scalar must fail this test.
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$script:case='startup';$script:calls=0
$script:typeMethod=[object].GetMethod('GetType')
$script:membersProperty=[Management.Automation.PSMemberSet].GetProperty('Members')
$script:metadataName='P1PlainOwnershipMetadata'
$script:referenceMethod=[object].GetMethod('ReferenceEquals',[type[]]@([object],[object]))
$script:typeMethod=[object].GetMethod('GetType',[type[]]@())
$script:invokeMethod=[Reflection.MethodBase].GetMethod('Invoke',[type[]]@([object],[object[]]))
$script:stringConstructor=[string].GetConstructor([type[]]@([char[]]))
$script:wrapperConstructor=[Management.Automation.PSObject].GetConstructor([type[]]@([object]))
$script:baseGetter=[Management.Automation.PSObject].GetProperty('BaseObject').GetGetMethod()
function R-Reference($Left,$Right){
    $arguments=[object[]]@($Left,$Right)
    return [bool]$script:referenceMethod.Invoke($null,$arguments)
}
function R-Type($Value){
    $arguments=[object[]]@($Value,$null)
    return $script:invokeMethod.Invoke($script:typeMethod,$arguments)
}
function R-Base($Wrapper){
    $arguments=[object[]]@($Wrapper,$null)
    return ,($script:invokeMethod.Invoke($script:baseGetter,$arguments))
}
function R-String {
    $parameters=[object[]]::new(1);$parameters[0]=[char[]]@('c','l','e','a','n')
    return ,($script:stringConstructor.Invoke($parameters))
}
function R-Wrapper($Base){return ,($script:wrapperConstructor.Invoke([object[]]@($Base)))}

$script:priorViews=[Collections.Generic.List[object]]::new()
$script:scalarCalls=0;$script:metadataCalls=0
# Test-only instrumentation. No subject/result objects enter these helpers.
$script:TOrigin=[Diagnostics.Stopwatch]::GetTimestamp()
$script:TFrequency=[Diagnostics.Stopwatch]::Frequency
$script:TSequence=0;$script:TActive='';$script:TCase='';$script:TPair=0
$script:TStart=0L;$script:TBodyTotal=0L;$script:TEmitTotal=0L
$script:TStageTicks=@{};$script:TStageCounts=@{}
$script:TDecisionCompleted=0;$script:TScalarCompleted=0;$script:TMetadataCompleted=0
function T-Emit([string]$Boundary,$Duration){
    $emitStart=[Diagnostics.Stopwatch]::GetTimestamp()
    $script:TSequence++
    $record=[pscustomobject][ordered]@{
        Schema=1;Sequence=$script:TSequence;Case=$script:TCase;Stage=$script:TActive;Boundary=$Boundary;Pair=$script:TPair
        Frequency=$script:TFrequency;ElapsedTicks=($emitStart-$script:TOrigin);DurationTicks=$Duration
        StageTicks=$script:TStageTicks[$script:TActive];StageCompleted=$script:TStageCounts[$script:TActive]
        BodyTicksTotal=$script:TBodyTotal;EmitTicksBefore=$script:TEmitTotal
        DecisionAttempted=$script:calls;DecisionCompleted=$script:TDecisionCompleted
        ScalarAttempted=$script:scalarCalls;ScalarCompleted=$script:TScalarCompleted
        MetadataAttempted=$script:metadataCalls;MetadataCompleted=$script:TMetadataCompleted
    }
    [Console]::Out.WriteLine('TIME: '+($record|ConvertTo-Json -Compress))
    [Console]::Out.Flush()
    $script:TEmitTotal+=[Diagnostics.Stopwatch]::GetTimestamp()-$emitStart
}
function T-Begin([string]$Stage){
    if($script:TActive -cne ''){throw 'TIMING_INVALID: overlapping stages'}
    $script:TActive=$Stage;$script:TCase=$script:case;$script:TPair=$script:TSequence+1
    if(-not $script:TStageTicks.ContainsKey($Stage)){$script:TStageTicks[$Stage]=0L;$script:TStageCounts[$Stage]=0}
    T-Emit enter $null
    # Body starts AFTER entry serialization/output/flush, with no nested timing.
    $script:TStart=[Diagnostics.Stopwatch]::GetTimestamp()
}
function T-End([string]$Stage){
    $end=[Diagnostics.Stopwatch]::GetTimestamp()
    if($script:TActive -cne $Stage){throw 'TIMING_INVALID: stage mismatch'}
    $duration=$end-$script:TStart
    if($duration -lt 0){throw 'TIMING_INVALID: negative duration'}
    $script:TStageTicks[$Stage]+=$duration;$script:TStageCounts[$Stage]++
    $script:TBodyTotal+=$duration
    T-Emit leave $duration
    $script:TActive=''
}
function P1O-Fail([string]$Outcome,[string]$Detail){throw ($Outcome+'|'+$script:case+'|'+$Detail)}
function P1O-Need([bool]$Condition,[string]$Detail){if(-not $Condition){P1O-Fail EXECUTION_INCOMPLETE $Detail}}
function P1O-Type($Value){if($null -eq $Value){return $null};$script:typeMethod.Invoke($Value.PSObject.BaseObject,$null)}
function P1O-Members($Value){
    $view=$Value.psextended
    return ,($script:membersProperty.GetValue($view.PSObject.BaseObject,$null))
}
function P1O-Plain($Value,[string]$Label){
    if($null -eq $Value){return}
    foreach($member in (P1O-Members $Value)){P1O-Fail FIXTURE_INVALID ('not-initially-plain:'+ $Label)}
}
function P1O-Equal($Actual,$Expected,[int]$Depth=0){
    if($Depth -gt 4){return $false}
    if($null -eq $Expected){return ($null -eq $Actual)}
    if($Expected -is [pscustomobject]){
        if($Actual -isnot [pscustomobject]){return $false}
        $props=@($Actual.PSObject.Properties);$want=@($Expected.PSObject.Properties)
        if($props.Count -ne $want.Count){return $false}
        foreach($p in $want){
            $matches=@($props | Where-Object {[string]::Equals($_.Name,$p.Name,[StringComparison]::Ordinal)})
            if($matches.Count -ne 1 -or $matches[0].MemberType -ne [Management.Automation.PSMemberTypes]::NoteProperty){return $false}
            if(-not (P1O-Equal $matches[0].Value $p.Value ($Depth+1))){return $false}
        }
        return $true
    }
    if((P1O-Type $Actual) -ne (P1O-Type $Expected)){return $false}
    if($Expected -is [string]){return [string]::Equals($Actual.PSObject.BaseObject,$Expected.PSObject.BaseObject,[StringComparison]::Ordinal)}
    if($Expected -is [bool] -or $Expected -is [int] -or $Expected -is [long]){return ($Actual.PSObject.BaseObject -ceq $Expected.PSObject.BaseObject)}
    return $false
}
function P1O-Fresh([type]$Type){
    if($Type -eq [string]){return ,([string]::new([char[]]@('c','l','e','a','n')))}
    $storage=[Array]::CreateInstance($Type,1)
    if($Type -eq [bool]){$storage.SetValue($true,0)}
    return ,($storage.GetValue(0))
}
function P1O-Observation($Sequence,$At,[string]$Kind,[string]$Actor,[hashtable]$Payload){
    $o=[ordered]@{Sequence=$Sequence;SessionLabel='s';ClockLabel='synthetic-ms';AtMs=$At;Actor=$Actor;Kind=$Kind}
    foreach($key in $Payload.Keys){$o[$key]=$Payload[$key]};[pscustomobject]$o
}
function P1O-Scenario([bool]$Sticky=$false){
    $observations=[object[]]@(
        (P1O-Observation 0 0 arm W @{LauncherOwnsControl=$true;SupervisorOwnsControl=$true;ContainerBound=$true;HandleInheritanceBlocked=$true;ReadersReady=$true}),
        (P1O-Observation 1 1 release W @{InputBindingMatches=$true}),
        (P1O-Observation 2 2 progress C @{StdoutRetainedBytes=0;StderrRetainedBytes=0;DataEventCount=3;StdoutEOF=$true;StderrEOF=$true;ExitObserved=(P1O-Fresh ([bool]));ExitCode=(P1O-Fresh ([int]))}),
        (P1O-Observation 3 3 cleanup W @{Verified=$true;NaturalCompletion=(P1O-Fresh ([bool]));CountsReconciled=$true}),
        (P1O-Observation 4 4 persist W @{ReadbackMatches=$true;Classification=(P1O-Fresh ([string]))})
    )
    if($Sticky){$observations[2]=P1O-Observation 2 2 fault L @{Fault='supervisor-lost'}}
    [pscustomobject]@{Schema=1;Mode='policy-only';SessionLabel='s';ClockLabel='synthetic-ms';StartMs=0;Observations=$observations}
}
function P1O-Want([bool]$Sticky=$false,[type]$ExitType=[int]){
    # Independent complete literal oracles. No product helper or actual result used.
    $d=[pscustomobject]@{
        DeadlineMs=[long]120000;StopAtMs=[long]113500;NaturalUntilMs=[long]502;CleanupUntilMs=[long]5502;PublishUntilMs=[long]6502
        LastAtMs=[long]4;LastSequence=[long]4;StdoutRetainedBytes=[long]0;StderrRetainedBytes=[long]0;DataEventCount=[long]3
        ControlSummaryCount=[long]0;ControlSummaryByteMax=[long]0;ControlQueueDepth=[long]0
        WouldCaptureStatus='complete';WouldCleanupStatus='verified';WouldStopOwnedSession=$false;WouldNaturalCompletion=$true
        WouldEvidenceComplete=$true;WouldPositiveTechnicalCandidate=$true;WouldExitObserved=$true;WouldExitCode=[int]0;WouldClassification='clean'
    }
    $r=[pscustomobject]@{Schema=1;Mode='policy-only';EvidenceClass='synthetic';Evaluated=$true;Reason='trace-evaluated';State='Persisted';FailureObservationIndex=$null;Decision=$d;OriginVerified=$false;Acceptance=$false;RuntimeApproved=$false;LiveActionsPerformed=$false;CleanupStatus='unknown'}
    if($ExitType -eq [long]){$d.WouldExitCode=[long]0}
    if($Sticky){
        $r.State='Aborted';$r.Reason='supervisor-lost';$r.FailureObservationIndex=[int]2
        $d.DataEventCount=[long]0;$d.WouldCaptureStatus='incomplete';$d.WouldStopOwnedSession=$true
        $d.WouldEvidenceComplete=$false;$d.WouldPositiveTechnicalCandidate=$false;$d.WouldExitObserved=$false;$d.WouldExitCode=$null
    }
    return $r
}
function P1O-Evaluate($Scenario){
    $script:calls++;P1O-Need ($script:calls -le 80) call-bound
    T-Begin decision
    Get-OfflineSupervisionDecision $Scenario
    $script:TDecisionCompleted++
    T-End decision
}
function P1O-Leaf($Owner,[string]$Field,[string]$Label){
    $descriptor=$Owner.PSObject.Properties[$Field]
    P1O-Need ($null -ne $descriptor -and $descriptor.MemberType -eq [Management.Automation.PSMemberTypes]::NoteProperty) field-descriptor
    $stored=R-Type $descriptor.Value
    $base=$Owner.$Field.PSObject.BaseObject
    $type=R-Type $base
    if($type -notin @([string],[int],[long],[bool]) -or ($stored -ne $type -and $stored -ne [Management.Automation.PSObject])){P1O-Fail FIXTURE_INVALID unknown-storage}
    $view=$Owner.$Field.PSObject
    $members=$view.Properties
    foreach($p in $members){if($p.MemberType -eq [Management.Automation.PSMemberTypes]::NoteProperty){P1O-Fail FIXTURE_INVALID initially-decorated-view}}
    return [pscustomobject]@{Owner=$Owner;Field=$Field;Label=$Label;Descriptor=$descriptor;Stored=$stored;Type=$type;Base=$base;View=$view;Members=$members}
}
function V-Invariant($Leaf){
    $field=$Leaf.Owner.PSObject.Properties[$Leaf.Field]
    $base=$Leaf.Owner.($Leaf.Field).PSObject.BaseObject
    if(-not (R-Reference $field $Leaf.Descriptor) -or (R-Type $field.Value) -ne $Leaf.Stored -or (R-Type $base) -ne $Leaf.Type -or -not (R-Reference $base $Leaf.Base) -or $base -cne $Leaf.Base){P1O-Fail FIXTURE_INVALID field-base-changed}
}
function V-Selected($Leaves,$Owner,[string]$Field){
    $found=@($Leaves|Where-Object{(R-Reference $_.Owner $Owner) -and $_.Field -ceq $Field})
    P1O-Need ($found.Count -eq 1) unique-selected-context
    return $found[0]
}
function V-Descriptor($Members,$Payload){
    $p=$Members[$script:metadataName]
    if($null -eq $p){return $null}
    if($p.MemberType -ne [Management.Automation.PSMemberTypes]::NoteProperty -or $p.Name -cne $script:metadataName -or -not (R-Reference $p.Value $Payload)){P1O-Fail FIXTURE_INVALID descriptor-payload}
    return ,$p
}
function V-Others($Leaves,$Selected){
    foreach($leaf in $Leaves){
        $ownerArguments=[object[]]@($leaf.Owner,$Selected.Owner)
        if([bool]$script:referenceMethod.Invoke($null,$ownerArguments) -and $leaf.Field -ceq $Selected.Field){continue}
        $field=$leaf.Owner.PSObject.Properties[$leaf.Field]
        $base=$leaf.Owner.($leaf.Field).PSObject.BaseObject
        $fieldArguments=[object[]]@($field,$leaf.Descriptor)
        if(-not [bool]$script:referenceMethod.Invoke($null,$fieldArguments)){P1O-Fail FIXTURE_INVALID field-base-changed}
        $storedArguments=[object[]]@($field.Value,$null)
        if($script:invokeMethod.Invoke($script:typeMethod,$storedArguments) -ne $leaf.Stored){P1O-Fail FIXTURE_INVALID field-base-changed}
        $typeArguments=[object[]]@($base,$null)
        if($script:invokeMethod.Invoke($script:typeMethod,$typeArguments) -ne $leaf.Type){P1O-Fail FIXTURE_INVALID field-base-changed}
        $baseArguments=[object[]]@($base,$leaf.Base)
        if(-not [bool]$script:referenceMethod.Invoke($null,$baseArguments) -or $base -cne $leaf.Base){P1O-Fail FIXTURE_INVALID field-base-changed}
        if($null -ne $leaf.Members[$script:metadataName] -or $null -ne $leaf.Owner.($leaf.Field).PSObject.Properties[$script:metadataName]){P1O-Fail OWNERSHIP_GAP_OBSERVED ('cross-owner:'+$leaf.Label)}
    }
}
function V-Exercise($Leaves,$Selected,[bool]$Independent){
    T-Begin selected-before
    P1O-Need $Independent independent-backing
    $payload=[object[]]@('initial')
    $Selected.Members.Add([Management.Automation.PSNoteProperty]::new($script:metadataName,$payload))
    $held=V-Descriptor $Selected.Members $payload
    if($null -eq $held){P1O-Fail FIXTURE_INVALID held-attachment-not-retained}
    $read=V-Descriptor $Selected.Owner.($Selected.Field).PSObject.Properties $payload
    $retained=$null -ne $read
    if($Selected.Stored -ne [string] -and -not $retained){P1O-Fail FIXTURE_INVALID required-reread-retention}
    V-Invariant $Selected
    T-End selected-before
    T-Begin current-before
    V-Others $Leaves $Selected
    T-End current-before
    T-Begin prior-before
    V-Others $script:priorViews $Selected
    T-End prior-before
    T-Begin selected-after
    $held.Value[0]='changed'
    $heldAfter=V-Descriptor $Selected.View.Properties $payload
    if($null -eq $heldAfter -or -not (R-Reference $heldAfter $held) -or $heldAfter.Value[0] -cne 'changed' -or $payload[0] -cne 'changed'){P1O-Fail FIXTURE_INVALID held-mutation-not-retained}
    $after=V-Descriptor $Selected.Owner.($Selected.Field).PSObject.Properties $payload
    if(($null -ne $after) -ne $retained){P1O-Fail FIXTURE_INVALID inconsistent-reread}
    if($retained -and $after.Value[0] -cne 'changed'){P1O-Fail FIXTURE_INVALID reread-mutation-not-retained}
    V-Invariant $Selected
    T-End selected-after
    T-Begin current-after
    V-Others $Leaves $Selected
    T-End current-after
    T-Begin prior-after
    V-Others $script:priorViews $Selected
    T-End prior-after
    T-Begin record-retain
    $fact=[pscustomobject][ordered]@{
        Case=$script:case;StoredType=$Selected.Stored.FullName;BaseType=$Selected.Type.FullName
        HeldDescriptorRetained=$true;HeldPayloadSame=$true;HeldMutationRetained=$true
        RereadDescriptorRetained=$retained;RereadMutationRetained=$(if($retained){$true}else{$null})
        FieldBaseUnchanged=$true;CrossOwnerClean=$true;IndependentBase=$Independent
    }
    Write-Output ('FACT: '+($fact|ConvertTo-Json -Compress))
    foreach($leaf in $Leaves){if(-not ((R-Reference $leaf.Owner $Selected.Owner) -and $leaf.Field -ceq $Selected.Field)){$script:priorViews.Add($leaf)}}
    T-End record-retain
}
function V-Control([type]$Type,[string]$Name){
    $script:case=$Name;$script:metadataName='P1View_'+$Name
    T-Begin sharing-control
    $base=P1R-Fresh $Type $(if($Type -eq [string]){'clean'}else{[int]0})
    $shared=if($Type -eq [string]){R-Wrapper $base}else{$base}
    $left=[pscustomobject]@{Value=$shared};$right=[pscustomobject]@{Value=$shared}
    $fresh=P1R-Fresh $Type $(if($Type -eq [string]){'clean'}else{[int]0})
    if($Type -eq [string]){$fresh=R-Wrapper $fresh}
    $other=[pscustomobject]@{Value=$fresh}
    $l=P1O-Leaf $left Value left;$r=P1O-Leaf $right Value right;$o=P1O-Leaf $other Value other
    P1O-Need ((R-Reference $l.Base $r.Base) -and -not (R-Reference $l.Base $o.Base)) control-base-sharing
    if($Type -eq [string]){P1O-Need ($l.Stored -eq [Management.Automation.PSObject] -and $r.Stored -eq [Management.Automation.PSObject]) control-stored-wrapper}
    $payload=[object[]]@('initial')
    $l.Members.Add([Management.Automation.PSNoteProperty]::new($script:metadataName,$payload))
    foreach($c in @($l,$r)){
        $d=V-Descriptor $c.Members $payload;$read=V-Descriptor $c.Owner.Value.PSObject.Properties $payload
        P1O-Need ($null -ne $d -and $null -ne $read) known-sharing-visible
    }
    V-Others @($o) $l
    $payload[0]='changed'
    foreach($c in @($l,$r)){
        $d=V-Descriptor $c.Members $payload;$read=V-Descriptor $c.Owner.Value.PSObject.Properties $payload
        P1O-Need ($null -ne $d -and $null -ne $read -and $d.Value[0] -ceq 'changed' -and $read.Value[0] -ceq 'changed') shared-mutation-visible
        V-Invariant $c
    }
    V-Others @($o) $l
    T-End sharing-control
    Write-Output ('PASS: '+$Name)
}
function P1O-Leaves($Scenario,$A,$B,[bool]$Sticky){
    foreach($p in $Scenario.PSObject.Properties){
        if($p.Name -cne 'Observations' -and $null -ne $p.Value){P1O-Leaf $Scenario $p.Name ('input.'+$p.Name)}
    }
    for($i=0;$i -lt $Scenario.Observations.Length;$i++){
        foreach($p in $Scenario.Observations[$i].PSObject.Properties){
            if($null -ne $p.Value){P1O-Leaf $Scenario.Observations[$i] $p.Name ('input.'+$i+'.'+$p.Name)}
        }
    }
    foreach($pair in @(@{Value=$A;Name='A'},@{Value=$B;Name='B'})){
        foreach($p in $pair.Value.PSObject.Properties){
            if($p.Name -ceq 'Decision'){
                if($null -ne $p.Value){foreach($q in $p.Value.PSObject.Properties){if($null -ne $q.Value){P1O-Leaf $p.Value $q.Name ($pair.Name+'.Decision.'+$q.Name)}}}
            }elseif($null -ne $p.Value){P1O-Leaf $pair.Value $p.Name ($pair.Name+'.'+$p.Name)}
        }
    }
}
function P1R-Fresh([type]$Type,$Value){
    if($Type -eq [string]){
        $chars=[char[]]$Value.ToCharArray()
        return ,([string]::new($chars))
    }
    $storage=[Array]::CreateInstance($Type,1);$storage.SetValue($Value,0)
    return ,($storage.GetValue(0))
}
function P1R-PreparedWant {
    [pscustomobject]@{Schema=1;Mode='policy-only';EvidenceClass='synthetic';Evaluated=$true;Reason='trace-evaluated';State='Prepared';FailureObservationIndex=$null;Decision=[pscustomobject]@{
        DeadlineMs=[long]120000;StopAtMs=[long]113500;NaturalUntilMs=$null;CleanupUntilMs=$null;PublishUntilMs=$null;LastAtMs=[long]0;LastSequence=$null
        StdoutRetainedBytes=[long]0;StderrRetainedBytes=[long]0;DataEventCount=[long]0;ControlSummaryCount=[long]0;ControlSummaryByteMax=[long]0;ControlQueueDepth=[long]0
        WouldCaptureStatus='incomplete';WouldCleanupStatus='unknown';WouldStopOwnedSession=$false;WouldNaturalCompletion=$false;WouldEvidenceComplete=$false;WouldPositiveTechnicalCandidate=$false;WouldExitObserved=$false;WouldExitCode=$null;WouldClassification='unknown'
    };OriginVerified=$false;Acceptance=$false;RuntimeApproved=$false;LiveActionsPerformed=$false;CleanupStatus='unknown'}
}
function P1R-RefusedWant {
    [pscustomobject]@{Schema=1;Mode='policy-only';EvidenceClass='synthetic';Evaluated=$false;Reason='invalid-observation';State='Refused';FailureObservationIndex=[int]0;Decision=$null;OriginVerified=$false;Acceptance=$false;RuntimeApproved=$false;LiveActionsPerformed=$false;CleanupStatus='unknown'}
}
function P1R-ScalarFixtures {
    $script:case='scalar-null'
    T-Begin metadata
    $script:metadataCalls++;$metadata=New-SupervisionPolicyOwnershipMetadata
    $script:TMetadataCompleted++
    T-End metadata
    T-Begin scalar-null
    $script:scalarCalls++;P1O-Need ($null -eq (Copy-SupervisionPolicyScalar $null $metadata)) null-copy
    $script:TScalarCompleted++
    T-End scalar-null
    Write-Output 'PASS: scalar-null'
    $vectors=@(
        @{Name='int32-zero';Type=[int];Value=[int]0},@{Name='int32-min';Type=[int];Value=[int]::MinValue},@{Name='int32-max';Type=[int];Value=[int]::MaxValue},
        @{Name='int64-zero';Type=[long];Value=[long]0},@{Name='int64-min';Type=[long];Value=[long]::MinValue},@{Name='int64-max';Type=[long];Value=[long]::MaxValue},
        @{Name='bool-true';Type=[bool];Value=$true},@{Name='bool-false';Type=[bool];Value=$false},
        @{Name='string-clean';Type=[string];Value='clean'},@{Name='string-refused';Type=[string];Value='refused'},@{Name='string-unknown';Type=[string];Value='unknown'},
        @{Name='string-mode';Type=[string];Value='policy-only'},@{Name='string-fault';Type=[string];Value='supervisor-lost'},@{Name='string-error';Type=[string];Value='internal-error'}
    )
    foreach($v in $vectors){
        $script:case='scalar-'+$v.Name;$script:metadataName='P1Owned_'+$script:case
    T-Begin scalar-fixture
        $source=[pscustomobject]@{Value=(P1R-Fresh $v.Type $v.Value)}
    T-End scalar-fixture
    T-Begin scalar-A
        $script:scalarCalls++;$a=[pscustomobject]@{Value=(Copy-SupervisionPolicyScalar $source.Value $metadata)}
    $script:TScalarCompleted++
    T-End scalar-A
    T-Begin scalar-B
        $script:scalarCalls++;$b=[pscustomobject]@{Value=(Copy-SupervisionPolicyScalar $source.Value $metadata)}
    $script:TScalarCompleted++
    T-End scalar-B
    T-Begin scalar-oracle-capture
        foreach($o in @($source,$a,$b)){
            P1O-Need ((P1O-Type $o.Value) -eq $v.Type -and $o.Value.PSObject.BaseObject -ceq $v.Value) scalar-type-value
            P1O-Plain $o.Value scalar-plain
        }
        P1O-Need (-not [object]::ReferenceEquals($source.Value.PSObject.BaseObject,$a.Value.PSObject.BaseObject) -and -not [object]::ReferenceEquals($source.Value.PSObject.BaseObject,$b.Value.PSObject.BaseObject) -and -not [object]::ReferenceEquals($a.Value.PSObject.BaseObject,$b.Value.PSObject.BaseObject)) independent-backing
        $leaves=@((P1O-Leaf $source Value source),(P1O-Leaf $a Value A),(P1O-Leaf $b Value B))
    T-End scalar-oracle-capture
        V-Exercise $leaves $leaves[1] $true
        Write-Output ('PASS: '+$script:case)
    }
}
try{
    $identity=[Security.Principal.WindowsIdentity]::GetCurrent()
    $principal=[Security.Principal.WindowsPrincipal]::new($identity)
    P1O-Need (-not $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) ordinary-host
    P1O-Need ($PSVersionTable.PSEdition -ceq 'Core' -and $PSVersionTable.PSVersion.ToString() -ceq '7.6.5' -and [Environment]::Is64BitProcess) engine-profile
    Write-Output 'HOST: 7.6.5 Core x64=True admin=False'
    $module=Join-Path $PSScriptRoot '../scalar-owned-return-candidate/offline-supervision-policy.ps1'
    P1O-Need ((Get-FileHash -LiteralPath $module).Hash -ceq '1971F808DAB0C0A520FC708BF240AF4367D21EE811FE9AEB8664E605E7E8BCE8') source-pin
    T-Begin module-load
    . $module
    T-End module-load
    V-Control ([string]) control-shared-string
    V-Control ([int]) control-shared-int32
    P1R-ScalarFixtures
    $script:case='control-before';$before=P1O-Evaluate (P1O-Scenario)
    T-Begin oracle-before
    P1O-Need (P1O-Equal $before (P1O-Want)) full-literal-before
    T-End oracle-before
    Write-Output 'PASS: control-before'
    $rows=@(
        @{Name='normal-exit-int32';Sticky=$false;Index=2;Input='ExitCode';Output='WouldExitCode';Type=[int]},
        @{Name='normal-exit-int64';Sticky=$false;Index=2;Input='ExitCode';Output='WouldExitCode';Type=[long]},
        @{Name='normal-observed';Sticky=$false;Index=2;Input='ExitObserved';Output='WouldExitObserved';Type=[bool]},
        @{Name='normal-natural';Sticky=$false;Index=3;Input='NaturalCompletion';Output='WouldNaturalCompletion';Type=[bool]},
        @{Name='normal-classification';Sticky=$false;Index=4;Input='Classification';Output='WouldClassification';Type=[string]},
        @{Name='sticky-classification';Sticky=$true;Index=4;Input='Classification';Output='WouldClassification';Type=[string]}
    )
    $rows+=@{Name='sticky-reason';Sticky=$true;Index=2;Input='Fault';Output='Reason';Type=[string];Root=$true;Seed='supervisor-lost'}
    foreach($row in $rows){foreach($direction in @('input','A','B')){
        $script:case=$row.Name+'-'+$direction;$script:metadataName='P1Owned_'+$script:case;Write-Output ('BEGIN: '+$script:case)
    T-Begin fixture
        $s=P1O-Scenario $row.Sticky
        $seed=if($row.ContainsKey('Seed')){$row.Seed}elseif($row.Type -eq [string]){'clean'}elseif($row.Type -eq [bool]){$true}else{0}
        if($row.Type -eq [long]){$seed=[long]0}
        $fresh=P1R-Fresh $row.Type $seed
        if($row.Type -eq [string]){$fresh=R-Wrapper $fresh}
        $s.Observations[$row.Index].($row.Input)=$fresh
        $inputOwner=$s.Observations[$row.Index];$base=$seed
        if((P1O-Type $inputOwner.($row.Input)) -ne $row.Type -or $inputOwner.($row.Input).PSObject.BaseObject -cne $base){P1O-Fail FIXTURE_INVALID input-type-value}
        P1O-Plain $inputOwner.($row.Input) input
    T-End fixture
        $a=P1O-Evaluate $s;$b=P1O-Evaluate $s
    T-Begin oracle-before
        $exitType=if($row.Type -eq [long]){[long]}else{[int]};$want=P1O-Want $row.Sticky $exitType
        P1O-Need ((P1O-Equal $a $want) -and (P1O-Equal $b $want)) full-literals
        P1O-Need (-not [object]::ReferenceEquals($a,$b) -and -not [object]::ReferenceEquals($a.Decision,$b.Decision)) fresh-containers
    T-End oracle-before
    T-Begin capture-select
        $leaves=@(P1O-Leaves $s $a $b $row.Sticky)
        foreach($leaf in $leaves){P1O-Plain $leaf.Owner.($leaf.Field) $leaf.Label}
        $aDestination=$a.Decision;$bDestination=$b.Decision
        if($row.ContainsKey('Root')){$aDestination=$a;$bDestination=$b}
        $selected=switch -CaseSensitive ($direction){
            input {V-Selected $leaves $inputOwner $row.Input}
            A {V-Selected $leaves $aDestination $row.Output}
            B {V-Selected $leaves $bDestination $row.Output}
        }
        $iLeaf=V-Selected $leaves $inputOwner $row.Input
        $aLeaf=V-Selected $leaves $aDestination $row.Output
        $bLeaf=V-Selected $leaves $bDestination $row.Output
        P1O-Need ($iLeaf.Type -eq $row.Type -and $aLeaf.Type -eq $row.Type -and $bLeaf.Type -eq $row.Type) selected-types
        $independent=(-not (R-Reference $iLeaf.Base $aLeaf.Base) -and -not (R-Reference $iLeaf.Base $bLeaf.Base) -and -not (R-Reference $aLeaf.Base $bLeaf.Base))
    T-End capture-select
        V-Exercise $leaves $selected $independent
    T-Begin oracle-after
        P1O-Need ((P1O-Equal $a $want) -and (P1O-Equal $b $want)) unchanged-base-literals
    T-End oracle-after
        Write-Output ('PASS: '+$script:case)
    }}
    $representatives=@(
        @{Name='root-schema';Kind='normal';Root=$true;Field='Schema'},@{Name='root-mode';Kind='normal';Root=$true;Field='Mode'},
        @{Name='root-reason';Kind='normal';Root=$true;Field='Reason'},@{Name='root-origin';Kind='normal';Root=$true;Field='OriginVerified'},
        @{Name='decision-deadline';Kind='normal';Root=$false;Field='DeadlineMs'},@{Name='decision-events';Kind='normal';Root=$false;Field='DataEventCount'},
        @{Name='decision-stop';Kind='normal';Root=$false;Field='WouldStopOwnedSession'},@{Name='prepared-classification';Kind='prepared';Root=$false;Field='WouldClassification'},
        @{Name='refused-index';Kind='refused';Root=$true;Field='FailureObservationIndex'}
    )
    foreach($row in $representatives){foreach($direction in @('A','B')){
        $script:case=$row.Name+'-'+$direction;$script:metadataName='P1Owned_'+$script:case
        Write-Output ('BEGIN: '+$script:case)
    T-Begin fixture
        $s=P1O-Scenario;$want=P1O-Want
        if($row.Kind -ceq 'prepared'){$s.Observations=[object[]]@();$want=P1R-PreparedWant}
        if($row.Kind -ceq 'refused'){$s.Observations[0].Kind='bogus';$want=P1R-RefusedWant}
    T-End fixture
        $a=P1O-Evaluate $s;$b=P1O-Evaluate $s
    T-Begin oracle-before
        P1O-Need ((P1O-Equal $a $want) -and (P1O-Equal $b $want)) representative-literals
        P1O-Need (-not [object]::ReferenceEquals($a,$b)) representative-result-containers
        if($null -ne $a.Decision){P1O-Need (-not [object]::ReferenceEquals($a.Decision,$b.Decision)) representative-decision-containers}
    T-End oracle-before
    T-Begin capture-select
        $leaves=@(P1O-Leaves $s $a $b $false)
        foreach($leaf in $leaves){P1O-Plain $leaf.Owner.($leaf.Field) $leaf.Label}
        $owner=if($direction -ceq 'A'){$a}else{$b}
        if(-not $row.Root){$owner=$owner.Decision}
        $expectedOwner=if($row.Root){$want}else{$want.Decision}
        $type=P1O-Type $expectedOwner.($row.Field);$base=$expectedOwner.($row.Field).PSObject.BaseObject
        $selected=V-Selected $leaves $owner $row.Field
        $otherOwner=if($direction -ceq 'A'){$b}else{$a}
        if(-not $row.Root){$otherOwner=$otherOwner.Decision}
        $other=V-Selected $leaves $otherOwner $row.Field
    T-End capture-select
        V-Exercise $leaves $selected (-not (R-Reference $selected.Base $other.Base))
    T-Begin oracle-after
        P1O-Need ((P1O-Equal $a $want) -and (P1O-Equal $b $want)) representative-values-unchanged
    T-End oracle-after
        Write-Output ('PASS: '+$script:case)
    }}
    $script:case='control-after';$after=P1O-Evaluate (P1O-Scenario)
    T-Begin oracle-after
    P1O-Need (P1O-Equal $after (P1O-Want)) full-literal-after
    T-End oracle-after
    Write-Output 'PASS: control-after'
    P1O-Need ($script:calls -eq 80) exact-call-count
    P1O-Need ($script:scalarCalls -eq 29 -and $script:metadataCalls -eq 1) helper-call-count
    P1O-Need ($script:TActive -ceq '' -and $script:TDecisionCompleted -eq 80 -and $script:TScalarCompleted -eq 29 -and $script:TMetadataCompleted -eq 1) timing-completion-counts
    Write-Output 'RESULT: ISOLATED_WITHIN_MATRIX cases=39 calls=80 scalarCalls=29 metadataCalls=1'
    Write-Output 'LIMIT: stated PS7 vectors only; no I1 closure, repair, promotion or runtime acceptance.'
}catch{
    $message=$_.Exception.Message
    if($message -cmatch '^(FIXTURE_INVALID|OWNERSHIP_GAP_OBSERVED|EXECUTION_INCOMPLETE)\|[^|]+\|[^|]+$'){
        $parts=$message -split '\|',3
        Write-Output ('RESULT: '+$parts[0]+' case='+$parts[1]+' detail='+$parts[2]+' calls='+$script:calls)
        exit 1
    }
    Write-Error $_;exit 1
}
