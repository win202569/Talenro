# Characterizes accepted-result ownership, not admission-time refusal.
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
    $script:calls++;P1O-Need ($script:calls -le 38) call-bound
    Get-OfflineSupervisionDecision $Scenario
}
function P1O-Leaf($Owner,[string]$Field,[string]$Label){[pscustomobject]@{Owner=$Owner;Field=$Field;Label=$Label}}
function P1O-Leaves($Scenario,$A,$B,[bool]$Sticky){
    if(-not $Sticky){
        P1O-Leaf $Scenario.Observations[2] ExitCode input.ExitCode
        P1O-Leaf $Scenario.Observations[2] ExitObserved input.ExitObserved
    }
    P1O-Leaf $Scenario.Observations[3] NaturalCompletion input.NaturalCompletion
    P1O-Leaf $Scenario.Observations[4] Classification input.Classification
    foreach($pair in @(@{Value=$A;Name='A'},@{Value=$B;Name='B'})){
        foreach($field in @('WouldExitCode','WouldExitObserved','WouldNaturalCompletion','WouldClassification')){
            if($null -ne $pair.Value.Decision.$field){P1O-Leaf $pair.Value.Decision $field ($pair.Name+'.'+$field)}
        }
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
    $module=Join-Path $PSScriptRoot '../integer-token-local-candidate/offline-supervision-policy.ps1'
    P1O-Need ((Get-FileHash -LiteralPath $module).Hash -ceq 'F9B4E408D9A4FF9C7A36EFD6E1F2472F17DACB3F76F463FF4973DAB4139CC82A') source-pin
    . $module
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
    $priorLeaves=[Collections.Generic.List[object]]::new()
    foreach($row in $rows){foreach($direction in @('input','A','B')){
        $script:case=$row.Name+'-'+$direction;Write-Output ('BEGIN: '+$script:case)
        $s=P1O-Scenario $row.Sticky
        $s.Observations[$row.Index].($row.Input)=P1O-Fresh $row.Type
        $inputOwner=$s.Observations[$row.Index];$base=if($row.Type -eq [string]){'clean'}elseif($row.Type -eq [bool]){$true}else{0}
        if((P1O-Type $inputOwner.($row.Input)) -ne $row.Type -or $inputOwner.($row.Input).PSObject.BaseObject -cne $base){P1O-Fail FIXTURE_INVALID input-type-value}
        P1O-Plain $inputOwner.($row.Input) input
        $control=[pscustomobject]@{Value=(P1O-Fresh $row.Type)};P1O-Plain $control.Value fixture-control
        $a=P1O-Evaluate $s;$b=P1O-Evaluate $s
        $exitType=if($row.Type -eq [long]){[long]}else{[int]};$want=P1O-Want $row.Sticky $exitType
        P1O-Need ((P1O-Equal $a $want) -and (P1O-Equal $b $want)) full-literals
        P1O-Need (-not [object]::ReferenceEquals($a,$b) -and -not [object]::ReferenceEquals($a.Decision,$b.Decision)) fresh-containers
        $leaves=@(P1O-Leaves $s $a $b $row.Sticky)
        foreach($leaf in $leaves){P1O-Plain $leaf.Owner.($leaf.Field) $leaf.Label}
        $selected=switch -CaseSensitive ($direction){
            input {P1O-Leaf $inputOwner $row.Input input}
            A {P1O-Leaf $a.Decision $row.Output A}
            B {P1O-Leaf $b.Decision $row.Output B}
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
    $script:case='control-after';$after=P1O-Evaluate (P1O-Scenario)
    P1O-Need (P1O-Equal $after (P1O-Want)) full-literal-after
    Write-Output 'PASS: control-after'
    P1O-Need ($script:calls -eq 38) exact-call-count
    Write-Output 'RESULT: ISOLATED_WITHIN_MATRIX cases=18 calls=38'
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
