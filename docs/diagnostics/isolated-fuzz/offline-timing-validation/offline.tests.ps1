param([Parameter(Mandatory)][string]$Packet)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$subject=Join-Path $PSScriptRoot '../stage-timing-ownership'
$checker=Join-Path $subject 'inspect-capture.ps1'
$inventory=Get-Content (Join-Path $subject 'output-inventory.json') -Raw|ConvertFrom-Json
$script:passed=0;$script:failed=0;$script:caseNumber=0
function Need([bool]$Condition,[string]$Detail){if(-not $Condition){throw $Detail}}
function Case([string]$Name,[scriptblock]$Body){
    $script:caseNumber++
    try{& $Body;$script:passed++;Write-Output ('PASS: '+$Name)}
    catch{$script:failed++;Write-Output ('FAIL: '+$Name+' detail='+[regex]::Replace($_.Exception.Message,'[\r\n]+',' '))}
}
function Store-Text([string]$Name,[string[]]$Lines){
    $path=Join-Path $Packet ($Name+'.synthetic.txt')
    if(Test-Path -LiteralPath $path){throw 'No fixture overwrite'}
    [IO.File]::WriteAllLines($path,$Lines,[Text.UTF8Encoding]::new($false))
    return $path
}
function Inspect([string]$Name,[string[]]$Lines,[string]$Mode='Prefix'){
    $path=Store-Text $Name $Lines
    $result=& $checker -CapturePath $path -Mode $Mode
    return ($result|ConvertFrom-Json)
}
function Reject([string]$Name,[string[]]$Lines,[string]$Mode='Prefix'){
    $path=Store-Text $Name $Lines
    $rejected=$false
    try{$null=& $checker -CapturePath $path -Mode $Mode}catch{$rejected=$true}
    Need $rejected 'invalid capture was accepted'
}
# Independently constructed protocol text; no candidate/test execution.
$records=[Collections.Generic.List[object]]::new()
$stageTicks=@{};$stageCounts=@{};$body=0L
$decisionA=0;$decisionC=0;$scalarA=0;$scalarC=0;$metaA=0;$metaC=0
foreach($item in $inventory.StagePlan){
    $parts=$item.Split('|');$caseName=$parts[0];$stage=$parts[1];$pair=$records.Count+1
    if(-not $stageTicks.ContainsKey($stage)){$stageTicks[$stage]=0L;$stageCounts[$stage]=0}
    if($stage -ceq 'decision'){$decisionA++}
    foreach($boundary in @('enter','leave')){
        if($boundary -ceq 'leave'){
            $stageTicks[$stage]+=10;$stageCounts[$stage]++;$body+=10
            if($stage -ceq 'decision'){$decisionC++}
            if($stage -cin @('scalar-null','scalar-A','scalar-B')){$scalarA++;$scalarC++}
            if($stage -ceq 'metadata'){$metaA++;$metaC++}
        }
        $seq=$records.Count+1
        $records.Add([pscustomobject][ordered]@{
            Schema=1;Sequence=$seq;Case=$caseName;Stage=$stage;Boundary=$boundary;Pair=$pair
            Frequency=1000;ElapsedTicks=($seq*100);DurationTicks=$(if($boundary -ceq 'enter'){$null}else{10})
            StageTicks=$stageTicks[$stage];StageCompleted=$stageCounts[$stage]
            BodyTicksTotal=$body;EmitTicksBefore=(($seq-1)*5)
            DecisionAttempted=$decisionA;DecisionCompleted=$decisionC
            ScalarAttempted=$scalarA;ScalarCompleted=$scalarC;MetadataAttempted=$metaA;MetadataCompleted=$metaC
        })
    }
}
$time=[string[]]@($records|ForEach-Object{'TIME: '+($_|ConvertTo-Json -Compress)})
$base=[string[]]@($inventory.BaseOutput.focused|ForEach-Object{
    if($_ -cmatch '^\{Fact:([A-Za-z0-9-]+)\}$'){
        $name=$Matches[1];$spec=@($inventory.BaseOutput.Facts|Where-Object{$_.Case -ceq $name})[0]
        $raw=$spec.StoredTypes[0];$retained=$raw -cne 'System.String'
        'FACT: '+([pscustomobject][ordered]@{
            Case=$name;StoredType=$raw;BaseType=$spec.BaseType;HeldDescriptorRetained=$true;HeldPayloadSame=$true;HeldMutationRetained=$true
            RereadDescriptorRetained=$retained;RereadMutationRetained=$(if($retained){$true}else{$null})
            FieldBaseUnchanged=$true;CrossOwnerClean=$true;IndependentBase=$true
        }|ConvertTo-Json -Compress)
    }else{$_}
})
function Changed([int]$Index,[string]$Field,$Value){
    $copy=$records[$Index]|ConvertTo-Json -Compress|ConvertFrom-Json
    $copy.$Field=$Value
    return ('TIME: '+($copy|ConvertTo-Json -Compress))
}
Case complete-synthetic-text {
    $r=Inspect complete ($time+$base) Complete
    Need ($r.Outcome -ceq 'COMPLETE_TEXT_ONLY' -and $r.TimingRecords -eq 1340 -and $r.BaseLines -eq 153) 'complete text classification'
    Need (-not $r.RuntimeVerified -and -not $r.ProcessReceiptValidated -and -not $r.FormalAcceptance) 'synthetic text cannot grant acceptance'
}
Case closed-prefix {$r=Inspect closed $time[0..1];Need ($r.Outcome -ceq 'PREFIX_ONLY' -and $null -eq $r.OpenAnnouncement) 'closed prefix classification'}
Case open-prefix {$r=Inspect open @($time[0]);Need ($r.Outcome -ceq 'PREFIX_ONLY' -and $null -ne $r.OpenAnnouncement) 'open announcement classification'}
Case empty-prefix {$r=Inspect empty @();Need ($r.Outcome -ceq 'PREFIX_ONLY' -and $r.TimingRecords -eq 0) 'empty is not complete'}
Case incomplete-complete-mode {Reject incomplete $time[0..1] Complete}
Case lowercase-complete-mode {Reject lowercase $time[0..1] complete}
Case missing-entry {Reject no-entry @($time[1])}
Case missing-middle-exit {Reject no-exit @($time[0],$time[2])}
Case duplicate-entry {Reject dup-entry @($time[0],$time[0])}
Case reordered {Reject reorder @($time[1],$time[0])}
Case negative-elapsed {Reject neg-clock @((Changed 0 ElapsedTicks -1))}
Case frequency-changes {Reject freq-change @($time[0],(Changed 1 Frequency 999))}
Case zero-frequency {Reject freq-zero @((Changed 0 Frequency 0))}
Case negative-duration {Reject neg-duration @($time[0],(Changed 1 DurationTicks -1))}
Case wrong-stage-total {Reject stage-total @($time[0],(Changed 1 StageTicks 11))}
Case wrong-body-total {Reject body-total @($time[0],(Changed 1 BodyTicksTotal 11))}
Case emission-overlap {Reject emit-overlap @((Changed 0 EmitTicksBefore 101))}
Case completed-exceeds-attempted {Reject counter-excess @((Changed 0 DecisionCompleted 1))}
Case decision-bound {Reject decision-max @((Changed 0 DecisionAttempted 81))}
Case scalar-bound {Reject scalar-max @((Changed 0 ScalarAttempted 30))}
Case metadata-bound {Reject metadata-max @((Changed 0 MetadataAttempted 2))}
Case floating-counter {Reject float @((Changed 0 DecisionAttempted 0.5))}
Case malformed-json {Reject malformed @('TIME: {')}
Case truncated-json {Reject truncated @($time[0].Substring(0,$time[0].Length-2))}
Case extra-json-field {Reject extra @($time[0].Replace('{"Schema":1,','{"Unexpected":0,"Schema":1,'))}
Case wrong-stage-name {Reject stage-name @((Changed 0 Stage 'other'))}
Case missing-ownership-fact {Reject no-fact ($time+@($base|Where-Object{-not $_.StartsWith('FACT: ')})) Complete}
Case false-ownership-fact {
    $b=[string[]]$base.Clone();$idx=0;while(-not $b[$idx].StartsWith('FACT: ')){$idx++}
    $b[$idx]=$b[$idx].Replace('"CrossOwnerClean":true','"CrossOwnerClean":false')
    Reject false-fact ($time+$b) Complete
}
Case counter-decreases {Reject counter-back @((Changed 0 DecisionAttempted 1),$time[1])}
Case wrong-stage-completed {Reject stage-count @($time[0],(Changed 1 StageCompleted 2))}
Case base-output-reordered {Reject base-order ($time+@($base[1],$base[0])+$base[2..152]) Complete}
Case timing-only-not-complete {Reject timing-only $time Complete}
Case missing-final-base-line {Reject tail-missing ($time+$base[0..151]) Complete}
Case duplicate-json-property {Reject dup-property @($time[0].Replace('{"Schema":1,','{"Schema":0,"Schema":1,'))}
# Parse as data. Execute ONLY these three logger definitions, never the script.
$sourcePath=Join-Path $subject 'owned-return.tests.ps1'
$tokens=$null;$errors=$null;$ast=[Management.Automation.Language.Parser]::ParseFile($sourcePath,[ref]$tokens,[ref]$errors)
Need ($errors.Count -eq 0) 'logger source parse'
$definitions=@($ast.FindAll({param($n)$n -is [Management.Automation.Language.FunctionDefinitionAst] -and $n.Name -cin @('T-Emit','T-Begin','T-End')},$false))
Need ($definitions.Count -eq 3) 'exact helper allowlist'
$module=New-Module -ScriptBlock {
    param($Definitions)
    $script:case='startup';$script:calls=0;$script:scalarCalls=0;$script:metadataCalls=0
    $script:TOrigin=[Diagnostics.Stopwatch]::GetTimestamp();$script:TFrequency=[Diagnostics.Stopwatch]::Frequency
    $script:TSequence=0;$script:TActive='';$script:TCase='';$script:TPair=0;$script:TStart=0L
    $script:TBodyTotal=0L;$script:TEmitTotal=0L;$script:TStageTicks=@{};$script:TStageCounts=@{}
    $script:TDecisionCompleted=0;$script:TScalarCompleted=0;$script:TMetadataCompleted=0
    . ([scriptblock]::Create($Definitions))
} -ArgumentList (($definitions|ForEach-Object{$_.Extent.Text}) -join [Environment]::NewLine)
Case logger-pipeline-and-prefix {
    $writer=[IO.StringWriter]::new([Globalization.CultureInfo]::InvariantCulture);$priorWriter=[Console]::Out
    try{
        [Console]::SetOut($writer)
        $payload=[object]::new()
        $actual=@(& $module {param($Payload) T-Begin module-load;$Payload;T-End module-load} $payload)
    }finally{[Console]::SetOut($priorWriter)}
    Need ($actual.Count -eq 1 -and [object]::ReferenceEquals($actual[0],$payload)) 'logger polluted/replaced output'
    $r=Inspect logger-prefix @($writer.ToString() -split '\r?\n'|Where-Object{$_ -cne ''})
    Need ($r.Outcome -ceq 'PREFIX_ONLY' -and $r.TimingRecords -eq 2 -and $r.CompletedBodyTicks -ge 0) 'actual logger prefix invalid'
    Need ($r.RecordedEmitTicksBeforeLast -gt 0 -and -not $r.LastEmissionDurationKnown) 'emission accounting'
    $writer.Dispose()
}
Case logger-rejects-nesting {
    $rejected=& $module {T-Begin scratch;$ok=$false;try{T-Begin nested}catch{$ok=$_.Exception.Message -ceq 'TIMING_INVALID: overlapping stages'};T-End scratch;return $ok}
    Need ($rejected -eq $true) 'nested stage accepted'
}
Case logger-rejects-wrong-end {
    $rejected=& $module {T-Begin scratch;$ok=$false;try{T-End wrong}catch{$ok=$_.Exception.Message -ceq 'TIMING_INVALID: stage mismatch'};T-End scratch;return $ok}
    Need ($rejected -eq $true) 'wrong stage end accepted'
}
Write-Output ('RESULT: OFFLINE_FIXTURES cases='+$script:caseNumber+' passed='+$script:passed+' failed='+$script:failed+' productCalls=0')
if($script:failed -gt 0){exit 1}
