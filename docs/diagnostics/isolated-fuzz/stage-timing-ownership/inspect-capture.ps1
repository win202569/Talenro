param(
    [Parameter(Mandatory)][string]$CapturePath,
    [ValidateSet('Prefix','Complete')][string]$Mode='Prefix'
)
# Offline saved-text inspection only. Never imports or launches the subject.
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$inventory=Get-Content -LiteralPath (Join-Path $PSScriptRoot 'output-inventory.json') -Raw|ConvertFrom-Json
function Need([bool]$Condition,[string]$Detail){if(-not $Condition){throw ('INVALID_CAPTURE: '+$Detail)}}
function Integer($Value){return ($Value -is [int] -or $Value -is [long])}
$lines=@([IO.File]::ReadAllLines((Resolve-Path -LiteralPath $CapturePath).Path,[Text.UTF8Encoding]::new($false,$true))|Where-Object{$_ -cne ''})
Need ($lines.Count -le 1493) line-bound
$timing=@($lines|Where-Object{$_.StartsWith('TIME: ',[StringComparison]::Ordinal)})
$base=@($lines|Where-Object{-not $_.StartsWith('TIME: ',[StringComparison]::Ordinal)})
Need ($timing.Count -le 1340 -and $base.Count -le 153) inventory-bounds
$stages=@{};$counts=@{};$body=0L;$elapsed=0L;$emit=0L;$frequency=$null;$open=$null;$last=$null
$counterNames=@('DecisionAttempted','DecisionCompleted','ScalarAttempted','ScalarCompleted','MetadataAttempted','MetadataCompleted')
$previousCounters=@{};foreach($n in $counterNames){$previousCounters[$n]=0L}
for($i=0;$i -lt $timing.Count;$i++){
    $r=$timing[$i].Substring(6)|ConvertFrom-Json
    Need ((@($r.PSObject.Properties.Name) -join '|') -ceq ($inventory.TimingProperties -join '|')) timing-schema
    $plan=([string]$inventory.StagePlan[[int][Math]::Floor($i/2)]).Split('|')
    $boundary=if($i%2 -eq 0){'enter'}else{'leave'}
    $pair=2*[int][Math]::Floor($i/2)+1
    Need ($r.Schema -eq 1 -and (Integer $r.Schema) -and (Integer $r.Sequence) -and $r.Sequence -eq $i+1) sequence
    Need ($r.Case -ceq $plan[0] -and $r.Stage -ceq $plan[1] -and $r.Boundary -ceq $boundary -and (Integer $r.Pair) -and $r.Pair -eq $pair) ordered-stage
    foreach($n in @('Frequency','ElapsedTicks','StageTicks','StageCompleted','BodyTicksTotal','EmitTicksBefore')+$counterNames){
        Need ((Integer $r.$n) -and $r.$n -ge 0) integer-fields
    }
    Need ($r.Frequency -gt 0) positive-frequency
    if($null -eq $frequency){$frequency=$r.Frequency}else{Need ($r.Frequency -eq $frequency) fixed-frequency}
    Need ($r.ElapsedTicks -ge $elapsed -and $r.EmitTicksBefore -ge $emit) monotonic-clock-emission
    foreach($n in $counterNames){Need ($r.$n -ge $previousCounters[$n]) monotonic-counters;$previousCounters[$n]=$r.$n}
    Need ($r.DecisionCompleted -le $r.DecisionAttempted -and $r.DecisionAttempted -le 80) decision-bound
    Need ($r.ScalarCompleted -le $r.ScalarAttempted -and $r.ScalarAttempted -le 29) scalar-bound
    Need ($r.MetadataCompleted -le $r.MetadataAttempted -and $r.MetadataAttempted -le 1) metadata-bound
    if(-not $stages.ContainsKey($r.Stage)){$stages[$r.Stage]=0L;$counts[$r.Stage]=0L}
    if($boundary -ceq 'enter'){
        Need ($null -eq $r.DurationTicks) enter-no-duration
        $open=$r
    }else{
        Need ($null -ne $open -and (Integer $r.DurationTicks) -and $r.DurationTicks -ge 0) leave-duration
        Need ($r.DurationTicks -le $r.ElapsedTicks-$open.ElapsedTicks) bounded-body
        $stages[$r.Stage]+=$r.DurationTicks;$counts[$r.Stage]++;$body+=$r.DurationTicks
        $open=$null
    }
    Need ($r.StageTicks -eq $stages[$r.Stage] -and $r.StageCompleted -eq $counts[$r.Stage] -and $r.BodyTicksTotal -eq $body) cumulative-stage-accounting
    Need ($r.BodyTicksTotal+$r.EmitTicksBefore -le $r.ElapsedTicks) exclusive-time-accounting
    $last=$r;$elapsed=$r.ElapsedTicks;$emit=$r.EmitTicksBefore
}
for($i=0;$i -lt $base.Count;$i++){
    $expected=[string]$inventory.BaseOutput.focused[$i]
    if($expected -cmatch '^\{Fact:([A-Za-z0-9-]+)\}$'){
        $caseName=$Matches[1]
        Need ($base[$i].StartsWith('FACT: ',[StringComparison]::Ordinal)) fact-prefix
        $fact=$base[$i].Substring(6)|ConvertFrom-Json
        Need ((@($fact.PSObject.Properties.Name) -join '|') -ceq ($inventory.BaseOutput.FactProperties -join '|')) fact-schema
        $want=@($inventory.BaseOutput.Facts|Where-Object{$_.Case -ceq $caseName})
        Need ($want.Count -eq 1 -and $fact.Case -ceq $caseName -and $fact.BaseType -ceq $want[0].BaseType -and $fact.StoredType -cin $want[0].StoredTypes) fact-case-storage
        foreach($n in @('HeldDescriptorRetained','HeldPayloadSame','HeldMutationRetained','FieldBaseUnchanged','CrossOwnerClean','IndependentBase')){Need ($fact.$n -is [bool] -and $fact.$n) required-true-fact}
        Need ($fact.RereadDescriptorRetained -is [bool]) reread-type
        Need ($fact.StoredType -ceq 'System.String' -or $fact.RereadDescriptorRetained) required-retention
        if($fact.RereadDescriptorRetained){Need ($fact.RereadMutationRetained -is [bool] -and $fact.RereadMutationRetained) reread-mutation}
        else{Need ($null -eq $fact.RereadMutationRetained) absent-reread-mutation}
    }else{Need ($base[$i] -ceq $expected) base-ordered-prefix}
}
$complete=$timing.Count -eq 1340 -and $base.Count -eq 153 -and $null -eq $open
if($complete){
    Need ($last.DecisionAttempted -eq 80 -and $last.DecisionCompleted -eq 80 -and $last.ScalarAttempted -eq 29 -and $last.ScalarCompleted -eq 29 -and $last.MetadataAttempted -eq 1 -and $last.MetadataCompleted -eq 1) complete-call-counts
}
if($Mode -eq 'Complete'){Need $complete complete-inventory-required}
# Valid prefix is NEVER promoted to a completed run or test acceptance.
[pscustomobject]@{
    Schema=1;Outcome=$(if($complete){'COMPLETE_TEXT_ONLY'}else{'PREFIX_ONLY'})
    TimingRecords=$timing.Count;BaseLines=$base.Count;LastBoundary=$last;OpenAnnouncement=$open
    StageTicks=$stages;CompletedBodyTicks=$body;RecordedEmitTicksBeforeLast=$emit
    UnclassifiedTicksBeforeLast=($elapsed-$body-$emit)
    Frequency=$frequency;LastEmissionDurationKnown=$false;ActualInterruptedBodyCountKnown=$false
    ProcessReceiptValidated=$false;RuntimeVerified=$false;FormalAcceptance=$false
}|ConvertTo-Json -Depth 8
