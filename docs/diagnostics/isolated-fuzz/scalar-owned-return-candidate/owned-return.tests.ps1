# Bounded regression: independent returned scalar leaves; no input sanitation.
# Removing owned-return copying must expose later descriptor contamination.
# A shallow result sharing a later-decorated caller scalar must fail this test.
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$script:case='startup';$script:calls=0
$script:typeMethod=[object].GetMethod('GetType')
$script:membersProperty=[Management.Automation.PSMemberSet].GetProperty('Members')
$script:metadataName='P1PlainOwnershipMetadata'
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
    Get-OfflineSupervisionDecision $Scenario
}
function P1O-Leaf($Owner,[string]$Field,[string]$Label){[pscustomobject]@{Owner=$Owner;Field=$Field;Label=$Label}}
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
    $metadata=New-SupervisionPolicyOwnershipMetadata
    P1O-Need ($null -eq (Copy-SupervisionPolicyScalar $null $metadata)) null-copy
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
        $source=[pscustomobject]@{Value=(P1R-Fresh $v.Type $v.Value)}
        $a=[pscustomobject]@{Value=(Copy-SupervisionPolicyScalar $source.Value $metadata)}
        $b=[pscustomobject]@{Value=(Copy-SupervisionPolicyScalar $source.Value $metadata)}
        foreach($o in @($source,$a,$b)){
            P1O-Need ((P1O-Type $o.Value) -eq $v.Type -and $o.Value.PSObject.BaseObject -ceq $v.Value) scalar-type-value
            P1O-Plain $o.Value scalar-plain
        }
        P1O-Need (-not [object]::ReferenceEquals($source.Value.PSObject.BaseObject,$a.Value.PSObject.BaseObject) -and -not [object]::ReferenceEquals($source.Value.PSObject.BaseObject,$b.Value.PSObject.BaseObject) -and -not [object]::ReferenceEquals($a.Value.PSObject.BaseObject,$b.Value.PSObject.BaseObject)) independent-backing
        $payload=[object[]]@('initial');P1O-Attach $a Value $payload
        $descriptor=P1O-CheckAttachment $a Value $payload $v.Type $v.Value scalar-copy
        foreach($o in @($source,$b)){P1O-Need ($null -eq $o.Value.PSObject.Properties[$script:metadataName]) scalar-no-cross-owner}
        $descriptor.Value[0]='changed'
        $null=P1O-CheckAttachment $a Value $payload $v.Type $v.Value scalar-copy-after
        P1O-Need ($payload[0] -ceq 'changed') scalar-mutation-retained
        foreach($o in @($source,$b)){P1O-Need ($null -eq $o.Value.PSObject.Properties[$script:metadataName]) scalar-no-cross-after}
        Write-Output ('PASS: '+$script:case)
    }
}
function P1O-Attach($Owner,[string]$Field,$Payload){
    # Deliberately operate through the existing owner field; never wrap or reassign.
    $Owner.$Field.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new($script:metadataName,$Payload))
}
function P1O-CheckAttachment($Owner,[string]$Field,$Payload,[type]$Type,$Base,[string]$Label){
    $value=$Owner.$Field;$descriptor=$value.PSObject.Properties[$script:metadataName]
    if((P1O-Type $value) -ne $Type -or $value.PSObject.BaseObject -cne $Base){P1O-Fail FIXTURE_INVALID ('base-changed:'+ $Label)}
    if($null -eq $descriptor -or $descriptor.MemberType -ne [Management.Automation.PSMemberTypes]::NoteProperty){P1O-Fail FIXTURE_INVALID ('attachment-not-retained:'+ $Label)}
    if(-not [object]::ReferenceEquals($descriptor.Value,$Payload)){P1O-Fail FIXTURE_INVALID ('payload-not-retained:'+ $Label)}
    return ,$descriptor
}
function P1O-CheckOthers($Leaves,$Selected,[string]$Phase){
    foreach($leaf in $Leaves){
        if([object]::ReferenceEquals($leaf.Owner,$Selected.Owner) -and $leaf.Field -ceq $Selected.Field){continue}
        $p=$leaf.Owner.($leaf.Field).PSObject.Properties[$script:metadataName]
        if($null -ne $p){P1O-Fail OWNERSHIP_GAP_OBSERVED ('cross-owner:'+ $leaf.Label+':'+$Phase)}
    }
}
try{
    $identity=[Security.Principal.WindowsIdentity]::GetCurrent()
    $principal=[Security.Principal.WindowsPrincipal]::new($identity)
    P1O-Need (-not $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) ordinary-host
    P1O-Need ($PSVersionTable.PSEdition -ceq 'Core' -and $PSVersionTable.PSVersion.ToString() -ceq '7.6.5' -and [Environment]::Is64BitProcess) engine-profile
    Write-Output 'HOST: 7.6.5 Core x64=True admin=False'
    $module=Join-Path $PSScriptRoot 'offline-supervision-policy.ps1'
    P1O-Need ((Get-FileHash -LiteralPath $module).Hash -ceq '1971F808DAB0C0A520FC708BF240AF4367D21EE811FE9AEB8664E605E7E8BCE8') source-pin
    . $module
    P1R-ScalarFixtures
    $script:case='control-before';$before=P1O-Evaluate (P1O-Scenario)
    P1O-Need (P1O-Equal $before (P1O-Want)) full-literal-before
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
    $priorLeaves=[Collections.Generic.List[object]]::new()
    foreach($row in $rows){foreach($direction in @('input','A','B')){
        $script:case=$row.Name+'-'+$direction;$script:metadataName='P1Owned_'+$script:case;Write-Output ('BEGIN: '+$script:case)
        $s=P1O-Scenario $row.Sticky
        $seed=if($row.ContainsKey('Seed')){$row.Seed}elseif($row.Type -eq [string]){'clean'}elseif($row.Type -eq [bool]){$true}else{0}
        if($row.Type -eq [long]){$seed=[long]0}
        $s.Observations[$row.Index].($row.Input)=P1R-Fresh $row.Type $seed
        $inputOwner=$s.Observations[$row.Index];$base=$seed
        if((P1O-Type $inputOwner.($row.Input)) -ne $row.Type -or $inputOwner.($row.Input).PSObject.BaseObject -cne $base){P1O-Fail FIXTURE_INVALID input-type-value}
        P1O-Plain $inputOwner.($row.Input) input
        $control=[pscustomobject]@{Value=(P1R-Fresh $row.Type $seed)};P1O-Plain $control.Value fixture-control
        $a=P1O-Evaluate $s;$b=P1O-Evaluate $s
        $exitType=if($row.Type -eq [long]){[long]}else{[int]};$want=P1O-Want $row.Sticky $exitType
        P1O-Need ((P1O-Equal $a $want) -and (P1O-Equal $b $want)) full-literals
        P1O-Need (-not [object]::ReferenceEquals($a,$b) -and -not [object]::ReferenceEquals($a.Decision,$b.Decision)) fresh-containers
        $leaves=@(P1O-Leaves $s $a $b $row.Sticky)
        foreach($leaf in $leaves){P1O-Plain $leaf.Owner.($leaf.Field) $leaf.Label}
        $aDestination=$a.Decision;$bDestination=$b.Decision
        if($row.ContainsKey('Root')){$aDestination=$a;$bDestination=$b}
        $selected=switch -CaseSensitive ($direction){
            input {P1O-Leaf $inputOwner $row.Input input}
            A {P1O-Leaf $aDestination $row.Output A}
            B {P1O-Leaf $bDestination $row.Output B}
        }
        $payload=[object[]]@('initial');$controlPayload=[object[]]@('initial')
        P1O-Attach $control Value $controlPayload
        $controlDescriptor=P1O-CheckAttachment $control Value $controlPayload $row.Type $base fixture-control
        P1O-Attach $selected.Owner $selected.Field $payload
        $descriptor=P1O-CheckAttachment $selected.Owner $selected.Field $payload $row.Type $base selected
        P1O-Need (-not [object]::ReferenceEquals($payload,$controlPayload)) distinct-payloads
        P1O-CheckOthers $leaves $selected before-mutation
        foreach($old in $priorLeaves){if($null -ne $old.Owner.($old.Field).PSObject.Properties[$script:metadataName]){P1O-Fail OWNERSHIP_GAP_OBSERVED cross-case}}
        $descriptor.Value[0]='changed'
        $null=P1O-CheckAttachment $selected.Owner $selected.Field $payload $row.Type $base selected-after
        if($payload[0] -cne 'changed' -or $controlDescriptor.Value[0] -cne 'initial'){P1O-Fail FIXTURE_INVALID mutation-control}
        $controlDescriptor.Value[0]='control-changed'
        $null=P1O-CheckAttachment $control Value $controlPayload $row.Type $base fixture-control-after
        if($payload[0] -cne 'changed' -or $controlPayload[0] -cne 'control-changed'){P1O-Fail FIXTURE_INVALID mutation-retention}
        P1O-CheckOthers $leaves $selected after-mutation
        P1O-Need ((P1O-Equal $a $want) -and (P1O-Equal $b $want)) unchanged-base-literals
        # Retain only unselected owners; their future contamination is a breach.
        foreach($leaf in $leaves){if(-not ([object]::ReferenceEquals($leaf.Owner,$selected.Owner) -and $leaf.Field -ceq $selected.Field)){$priorLeaves.Add($leaf)}}
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
        $s=P1O-Scenario;$want=P1O-Want
        if($row.Kind -ceq 'prepared'){$s.Observations=[object[]]@();$want=P1R-PreparedWant}
        if($row.Kind -ceq 'refused'){$s.Observations[0].Kind='bogus';$want=P1R-RefusedWant}
        $a=P1O-Evaluate $s;$b=P1O-Evaluate $s
        P1O-Need ((P1O-Equal $a $want) -and (P1O-Equal $b $want)) representative-literals
        P1O-Need (-not [object]::ReferenceEquals($a,$b)) representative-result-containers
        if($null -ne $a.Decision){P1O-Need (-not [object]::ReferenceEquals($a.Decision,$b.Decision)) representative-decision-containers}
        $leaves=@(P1O-Leaves $s $a $b $false)
        foreach($leaf in $leaves){P1O-Plain $leaf.Owner.($leaf.Field) $leaf.Label}
        $owner=if($direction -ceq 'A'){$a}else{$b}
        if(-not $row.Root){$owner=$owner.Decision}
        $expectedOwner=if($row.Root){$want}else{$want.Decision}
        $type=P1O-Type $expectedOwner.($row.Field);$base=$expectedOwner.($row.Field).PSObject.BaseObject
        $control=[pscustomobject]@{Value=(P1R-Fresh $type $base)}
        $payload=[object[]]@('initial');$controlPayload=[object[]]@('initial')
        P1O-Attach $control Value $controlPayload
        $cd=P1O-CheckAttachment $control Value $controlPayload $type $base representative-control
        $selected=P1O-Leaf $owner $row.Field $direction
        P1O-Attach $owner $row.Field $payload
        $d=P1O-CheckAttachment $owner $row.Field $payload $type $base representative-selected
        P1O-CheckOthers $leaves $selected before-mutation
        $d.Value[0]='changed';$null=P1O-CheckAttachment $owner $row.Field $payload $type $base representative-after
        P1O-Need ($payload[0] -ceq 'changed' -and $cd.Value[0] -ceq 'initial') representative-mutation
        $cd.Value[0]='control-changed';$null=P1O-CheckAttachment $control Value $controlPayload $type $base representative-control-after
        P1O-Need ($controlPayload[0] -ceq 'control-changed' -and $payload[0] -ceq 'changed') representative-independent-mutation
        P1O-CheckOthers $leaves $selected after-mutation
        P1O-Need ((P1O-Equal $a $want) -and (P1O-Equal $b $want)) representative-values-unchanged
        Write-Output ('PASS: '+$script:case)
    }}
    $script:case='control-after';$after=P1O-Evaluate (P1O-Scenario)
    P1O-Need (P1O-Equal $after (P1O-Want)) full-literal-after
    Write-Output 'PASS: control-after'
    P1O-Need ($script:calls -eq 80) exact-call-count
    Write-Output 'RESULT: ISOLATED_WITHIN_MATRIX cases=39 calls=80'
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
