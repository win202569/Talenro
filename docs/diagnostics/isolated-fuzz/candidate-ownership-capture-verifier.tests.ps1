param([string]$VerifierPath=(Join-Path $PSScriptRoot 'candidate-ownership-capture-verifier.ps1'),[string]$VerifierCommand='Test-CandidateOwnershipCapture')
# Offline verifier tests only: never imports or launches the policy/product.
$ErrorActionPreference='Stop'
function Need([bool]$Value,[string]$Case){if(-not $Value){throw ('ASSERTION FAILED: '+$Case)}}
$module=$VerifierPath
if(Test-Path $module){. $module}
$inventory=Get-Content (Join-Path $PSScriptRoot 'integer-candidate-three-phase/expected-output.json') -Raw|ConvertFrom-Json
$ticks=@(100,110,120,140,150,160,170,180,190,220,230,240,250,260,265,270,280,290);$k=0
$clean=@(foreach($rule in $inventory.Lines){
    switch -CaseSensitive ($rule.Kind){
        Context {'CONTEXT: SYNTHETIC\offline admin=False'}
        Host {'HOST: 7.6.5 Core x64=True'}
        Literal {$rule.Value}
        Boundary {('BOUNDARY: '+$rule.Name+' '+$rule.Edge+' ticks='+$ticks[$k]);$k++}
        Subphase {('SUBPHASE: '+$rule.Name+' '+$rule.Edge+' ticks='+$ticks[$k]);$k++}
        ProductPhase {('PRODUCTPHASE: '+$rule.Name+' '+$rule.Edge+' ticks='+$ticks[$k]);$k++}
    }
})
# Literal ownership records added independently of the new inventory/classifier.
$clean=@($clean[0..94]+'OWNERSHIP: SequentialIsolation START ticks=300'+'OWNERSHIP: SequentialIsolation END ticks=350'+$clean[95]+'OWNERSHIP: PlainControls START ticks=360'+'OWNERSHIP: PlainControls END ticks=390'+$clean[96..100])
function Receipt { [pscustomobject]@{StopwatchFrequency=[long]1000;ParentAdministratorRole=$false;ChildLaunched=$true;StreamsSaved=$true;TimedOut=$false;ExitCode=0;ElapsedMs=1000;CPUMs=1500;CleanupStatus='owned-host-exited';CaptureValid=$false;ExpectedResult=$false} }
function Verify($Lines=$clean,$Receipt=(Receipt),[string]$ErrorText=''){
    Need ($null -ne (Get-Command $VerifierCommand -ErrorAction SilentlyContinue)) 'VerifierAvailable'
    & $VerifierCommand -Stdout (($Lines -join "`n")+"`n") -Stderr $ErrorText -Receipt $Receipt
}
$count=0
function Case([string]$Name,[scriptblock]$Body){& $Body;$script:count++;Write-Output ('PASS: OwnershipVerifier/'+$Name)}
Case NormalPS7NotLegacyClassifier {
    $r=Verify
    Need ($r.CaptureValid -and $r.Complete -and $r.Status -ceq 'complete-diagnostic') 'NormalPS7'
    Need ($r.LineCount -eq 105 -and $r.PassCount -eq 64 -and $r.MarkerCount -eq 22 -and $r.ClosedPairs.Count -eq 11) 'CompleteCounts'
    Need (-not $r.FormalProjectAcceptance) 'NoAcceptance'
    Need ($r.ClosedPairs[0].ElapsedMs -eq 10 -and $r.ClosedPairs[1].ElapsedMs -eq 20 -and $r.ClosedPairs[2].ElapsedMs -eq 10 -and $r.ClosedPairs[3].ElapsedMs -eq 30 -and $r.ClosedPairs[4].ElapsedMs -eq 10 -and $r.ClosedPairs[5].ElapsedMs -eq 90 -and $r.ClosedPairs[6].ElapsedMs -eq 5 -and $r.ClosedPairs[7].ElapsedMs -eq 120 -and $r.ClosedPairs[8].ElapsedMs -eq 10) 'LiteralPairDurations'
}
# Each mutation below catches a missing admission/refusal check,not implementation text.
$mutations=@(
    @{Name='WrongHost';Index=1;Value='HOST: 5.1.26100 Desktop x64=True'},
    @{Name='HostRegexLiteralDot';Index=1;Value='HOST: 7x6.5 Core x64=True'},
    @{Name='HostSuffix';Index=1;Value='HOST: 7.6.5 Core x64=True trailing'},
    @{Name='ChildAdministrator';Index=0;Value='CONTEXT: SYNTHETIC\offline admin=True'},
    @{Name='LaterAdministrator';Index=32;Value='CONTEXT: SYNTHETIC\offline admin=True'},
    @{Name='ChangedAssertion';Index=5;Value='PASS: wrong'},
    @{Name='MarkerName';Index=65;Value='BOUNDARY: Unknown START ticks=100'},
    @{Name='MarkerEdge';Index=65;Value='BOUNDARY: InvalidInputs END ticks=100'},
    @{Name='NegativeTick';Index=65;Value='BOUNDARY: InvalidInputs START ticks=-1'},
    @{Name='OverflowTick';Index=65;Value='BOUNDARY: InvalidInputs START ticks=9223372036854775808'},
    @{Name='NonIntegerTick';Index=65;Value='BOUNDARY: InvalidInputs START ticks=1.5'},
    @{Name='BackwardPair';Index=66;Value='BOUNDARY: InvalidInputs END ticks=99'},
    @{Name='BackwardAcrossPairs';Index=67;Value='BOUNDARY: Build4096 START ticks=109'}
)
foreach($m in $mutations){Case $m.Name {$v=[string[]]$clean.Clone();$v[$m.Index]=$m.Value;$r=Verify $v;Need (-not $r.CaptureValid -and -not $r.Complete) $m.Name}}
Case RemovedMarker {$v=@($clean[0..65]+$clean[67..104]);$r=Verify $v;Need (-not $r.CaptureValid -and -not $r.Complete) 'RemovedMarker'}
Case DuplicateMarker {$v=@($clean[0..65]+$clean[65..104]);$r=Verify $v;Need (-not $r.CaptureValid) 'DuplicateMarker'}
Case ReorderedMarker {$v=[string[]]$clean.Clone();$v[65]=$clean[66];$v[66]=$clean[65];Need (-not (Verify $v).CaptureValid) 'ReorderedMarker'}
Case StderrRejects {Need (-not (Verify -ErrorText 'error').CaptureValid) 'Stderr'}
Case ExtraOutput {Need (-not (Verify @($clean+'extra')).CaptureValid) 'ExtraOutput'}
Case PartialMarkerRejects {$v=[string[]]$clean[0..69];$v[69]='BOUNDARY: Check4096 START tick';Need (-not (Verify $v).CaptureValid) 'PartialMarker'}
Case OpenPairPrefix {
    $receipt=Receipt;$receipt.TimedOut=$true;$receipt.ExitCode=-1;$receipt.ElapsedMs=30090
    $r=Verify $clean[0..70] $receipt
    Need ($r.CaptureValid -and -not $r.Complete -and $r.Status -ceq 'valid-prefix') 'PrefixNotComplete'
    Need ($r.ClosedPairs.Count -eq 2 -and $r.OpenScopes.Count -eq 2 -and $r.OpenScopes[0].Name -ceq 'Check4096' -and $r.OpenScopes[0].StartTicks -eq 150 -and $r.OpenScopes[1].Name -ceq 'Evaluate4096' -and $r.OpenScopes[1].StartTicks -eq 160) 'NoInventedEnd'
}
Case MissingTailNotComplete {$r=Verify $clean[0..103];Need ($r.CaptureValid -and -not $r.Complete) 'MissingTail'}
Case TimeoutNotComplete {$q=Receipt;$q.TimedOut=$true;Need (-not (Verify -Receipt $q).Complete) 'Timeout'}
Case ExitFailureNotComplete {$q=Receipt;$q.ExitCode=1;Need (-not (Verify -Receipt $q).Complete) 'Exit'}
Case OverBudgetNotComplete {$q=Receipt;$q.ElapsedMs=30001;Need (-not (Verify -Receipt $q).Complete) 'Budget'}
foreach($field in @('StopwatchFrequency','ElapsedMs','CPUMs','ParentAdministratorRole','ChildLaunched','StreamsSaved','CleanupStatus','ExitCode','TimedOut')){
    Case ('MissingReceipt/'+$field) {$q=Receipt;$q.PSObject.Properties.Remove($field);Need (-not (Verify -Receipt $q).CaptureValid) ('Missing/'+$field)}
}
foreach($value in @([long]0,[long]-1,1.5,'1000')){
    Case ('BadFrequency/'+[string]$value) {$q=Receipt;$q.StopwatchFrequency=$value;Need (-not (Verify -Receipt $q).CaptureValid) 'Frequency'}
}
foreach($value in @([double]::NaN,[double]::PositiveInfinity,-1,'1000')){
    Case ('BadElapsed/'+[string]$value) {$q=Receipt;$q.ElapsedMs=$value;Need (-not (Verify -Receipt $q).CaptureValid) 'Elapsed'}
}
Case ParentAdministrator {$q=Receipt;$q.ParentAdministratorRole=$true;Need (-not (Verify -Receipt $q).CaptureValid) 'Parent'}
Case StringBoolean {$q=Receipt;$q.TimedOut='False';Need (-not (Verify -Receipt $q).CaptureValid) 'Boolean'}
Case CleanupUnknown {$q=Receipt;$q.CleanupStatus='unknown';Need (-not (Verify -Receipt $q).CaptureValid) 'Cleanup'}
Case NoFinalNewline {Need (-not (& $VerifierCommand -Stdout ($clean -join "`n") -Stderr '' -Receipt (Receipt)).CaptureValid) 'UnterminatedTail'}
Case CRLFWorks {$r=& $VerifierCommand -Stdout (($clean -join "`r`n")+"`r`n") -Stderr '' -Receipt (Receipt);Need $r.Complete 'CRLF'}


Case InnerOrderRejects {$v=[string[]]$clean.Clone();$v[70]=$clean[77];$v[77]=$clean[70];Need (-not (Verify $v).CaptureValid) 'InnerOrder'}
Case InnerBeforeParentTick {$v=[string[]]$clean.Clone();$v[70]='SUBPHASE: Evaluate4096 START ticks=149';Need (-not (Verify $v).CaptureValid) 'ContainmentStart'}
Case ParentEndsBeforeChild {$v=[string[]]$clean.Clone();$v[80]='BOUNDARY: Check4096 END ticks=264';Need (-not (Verify $v).CaptureValid) 'ContainmentEnd'}
Case SiblingBackwards {$v=[string[]]$clean.Clone();$v[78]='SUBPHASE: Assert4096 START ticks=249';Need (-not (Verify $v).CaptureValid) 'SiblingTicks'}
Case PrematureParentEnd {$v=[string[]]$clean.Clone();$v[77]='BOUNDARY: Check4096 END ticks=250';Need (-not (Verify $v).CaptureValid) 'ParentStack'}
Case DuplicateInnerEnd {$v=[string[]]$clean.Clone();$v[78]=$clean[77];Need (-not (Verify $v).CaptureValid) 'DuplicateInner'}
Case InnerMalformedTick {$v=[string[]]$clean.Clone();$v[70]='SUBPHASE: Evaluate4096 START ticks=9223372036854775808';Need (-not (Verify $v).CaptureValid) 'InnerOverflow'}
Case ParentStillOpenAfterEval {
    $r=Verify $clean[0..77]
    Need ($r.CaptureValid -and -not $r.Complete -and $r.ClosedPairs.Count -eq 6 -and $r.OpenScopes.Count -eq 1 -and $r.OpenScopes[0].Name -ceq 'Check4096') 'OpenParent'
}
Case AssertStillOpen {
    $r=Verify $clean[0..78]
    Need ($r.CaptureValid -and -not $r.Complete -and $r.ClosedPairs.Count -eq 6 -and $r.OpenScopes.Count -eq 2 -and $r.OpenScopes[1].Name -ceq 'Assert4096') 'OpenAssertion'
}
Case ParentOpenAfterAssert {
    $r=Verify $clean[0..79]
    Need ($r.CaptureValid -and -not $r.Complete -and $r.ClosedPairs.Count -eq 7 -and $r.OpenScopes.Count -eq 1) 'ParentUnfinished'
}
Case ZeroSpanIsValid {$v=[string[]]$clean.Clone();$v[72]='PRODUCTPHASE: RootSetup END ticks=170';$r=Verify $v;Need ($r.Complete -and $r.ClosedPairs[2].ElapsedMs -eq 0) 'ZeroSpan'}
Case Legacy95Rejected {$v=@($clean[0..70]+$clean[77..104]);Need (-not (Verify $v).CaptureValid) 'Legacy95'}
foreach($value in @([double]::NaN,[double]::PositiveInfinity,-1,'1500')){
    Case ('BadCPU/'+[string]$value) {$q=Receipt;$q.CPUMs=$value;Need (-not (Verify -Receipt $q).CaptureValid) 'CPU'}
}
Case RootOpenDepthThree {
    $r=Verify $clean[0..71]
    Need ($r.CaptureValid -and -not $r.Complete -and $r.ClosedPairs.Count -eq 2 -and $r.OpenScopes.Count -eq 3 -and $r.OpenScopes[2].Name -ceq 'RootSetup') 'OpenRoot'
}
Case ValidationOpenDepthThree {
    $q=Receipt;$q.TimedOut=$true;$q.ExitCode=-1;$q.ElapsedMs=30080
    $r=Verify $clean[0..73] $q
    Need ($r.CaptureValid -and -not $r.Complete -and $r.ClosedPairs.Count -eq 3 -and $r.OpenScopes.Count -eq 3 -and $r.OpenScopes[2].Name -ceq 'ValidationCopy') 'OpenValidation'
}
Case DecisionOpenDepthThree {
    $r=Verify $clean[0..75]
    Need ($r.CaptureValid -and -not $r.Complete -and $r.ClosedPairs.Count -eq 4 -and $r.OpenScopes.Count -eq 3 -and $r.OpenScopes[2].Name -ceq 'Decision') 'OpenDecision'
}
Case PhaseEndStillRequiresAPIEnd {
    $r=Verify $clean[0..76]
    Need ($r.CaptureValid -and -not $r.Complete -and $r.ClosedPairs.Count -eq 5 -and $r.OpenScopes.Count -eq 2) 'OpenAPI'
}
Case EarlyReturnIsInvalidNotComplete {
    $v=@($clean[0..71]+$clean[77..104]);$r=Verify $v
    Need (-not $r.CaptureValid -and -not $r.Complete -and $r.Status -ceq 'invalid-capture') 'MissingPhaseEndOnReturn'
}
Case CrossedGrandparentEnd {$v=[string[]]$clean.Clone();$v[72]='BOUNDARY: Check4096 END ticks=180';Need (-not (Verify $v).CaptureValid) 'CrossedGrandparent'}
Case PhaseOutsideParentTime {$v=[string[]]$clean.Clone();$v[71]='PRODUCTPHASE: RootSetup START ticks=159';Need (-not (Verify $v).CaptureValid) 'PhaseContainment'}
Case PhaseOverlapsSibling {$v=[string[]]$clean.Clone();$v[73]='PRODUCTPHASE: ValidationCopy START ticks=179';Need (-not (Verify $v).CaptureValid) 'PhaseOverlap'}
Case PhaseWrongName {$v=[string[]]$clean.Clone();$v[71]='PRODUCTPHASE: ValidationCopy START ticks=170';Need (-not (Verify $v).CaptureValid) 'PhaseName'}
Case PhaseOverflow {$v=[string[]]$clean.Clone();$v[71]='PRODUCTPHASE: RootSetup START ticks=9223372036854775808';Need (-not (Verify $v).CaptureValid) 'PhaseOverflow'}
Case PhaseMissingMarker {$v=@($clean[0..71]+$clean[73..104]);Need (-not (Verify $v).CaptureValid) 'PhaseMissing'}
Case PhaseDuplicateMarker {$v=@($clean[0..71]+$clean[71..104]);Need (-not (Verify $v).CaptureValid) 'PhaseDuplicate'}
Case Legacy91Rejected {$v=@($clean[0..69]+$clean[80..104]);Need (-not (Verify $v).CaptureValid) 'Legacy91'}
# Breaks caught: treating body END as PASS, fabricating body entry/end, accepting
# crossed ownership spans, or labelling a non-timeout exit as an actual timeout.
Case OwnershipDurations {
    $r=Verify
    Need ($r.ClosedPairs[9].Kind -ceq 'OwnershipCase' -and $r.ClosedPairs[9].Name -ceq 'SequentialIsolation' -and $r.ClosedPairs[9].Parent -ceq '' -and $r.ClosedPairs[9].ElapsedMs -eq 50) 'SequentialDuration'
    Need ($r.ClosedPairs[10].Name -ceq 'PlainControls' -and $r.ClosedPairs[10].ElapsedMs -eq 30) 'PlainDuration'
    Need ($r.OwnershipProgress.Count -eq 2 -and $r.OwnershipProgress[0].State -ceq 'passed' -and $r.OwnershipProgress[1].State -ceq 'passed' -and $r.Termination -ceq 'completed') 'BothPassed'
}
$frontiers=@(
    @{Last=94;Seq='unobserved';Plain='unobserved';Pairs=9;Open=0},
    @{Last=95;Seq='entered';Plain='unobserved';Pairs=9;Open=1},
    @{Last=96;Seq='ended-awaiting-pass';Plain='unobserved';Pairs=10;Open=0},
    @{Last=97;Seq='passed';Plain='unobserved';Pairs=10;Open=0},
    @{Last=98;Seq='passed';Plain='entered';Pairs=10;Open=1},
    @{Last=99;Seq='passed';Plain='ended-awaiting-pass';Pairs=11;Open=0},
    @{Last=100;Seq='passed';Plain='passed';Pairs=11;Open=0},
    @{Last=103;Seq='passed';Plain='passed';Pairs=11;Open=0}
)
foreach($f in $frontiers){Case ('OwnershipFrontier/'+$f.Last) {
    $q=Receipt;$q.TimedOut=$true;$q.ExitCode=-1;$q.ElapsedMs=30080
    $r=Verify $clean[0..$f.Last] $q
    Need ($r.CaptureValid -and -not $r.Complete -and -not $r.FormalProjectAcceptance -and $r.Termination -ceq 'timeout') 'ActualTimeoutOnly'
    Need ($r.OwnershipProgress[0].State -ceq $f.Seq -and $r.OwnershipProgress[1].State -ceq $f.Plain) 'DistinctFrontier'
    Need ($r.ClosedPairs.Count -eq $f.Pairs -and $r.OpenScopes.Count -eq $f.Open) 'NoInventedEnd'
}}
Case OpenBodyErrorExitNotTimeout {
    $q=Receipt;$q.ExitCode=1;$r=Verify $clean[0..95] $q
    Need ($r.CaptureValid -and -not $r.Complete -and $r.Termination -ceq 'incomplete-exit' -and $r.OwnershipProgress[0].State -ceq 'entered') 'NoAssumedTimeout'
}
Case EndBeforePassNotComplete { $r=Verify $clean[0..99];Need (-not $r.Complete -and $r.OwnershipProgress[1].State -ceq 'ended-awaiting-pass' -and $r.Termination -ceq 'incomplete-exit') 'EndIsNotPass' }
$ownershipMutations=@(
    @{Name='OwnershipWrongName';Index=95;Value='OWNERSHIP: PlainControls START ticks=300'},
    @{Name='OwnershipWrongKind';Index=95;Value='BOUNDARY: SequentialIsolation START ticks=300'},
    @{Name='OwnershipWrongEdge';Index=95;Value='OWNERSHIP: SequentialIsolation END ticks=300'},
    @{Name='OwnershipNegative';Index=95;Value='OWNERSHIP: SequentialIsolation START ticks=-1'},
    @{Name='OwnershipOverflow';Index=95;Value='OWNERSHIP: SequentialIsolation START ticks=9223372036854775808'},
    @{Name='OwnershipFraction';Index=95;Value='OWNERSHIP: SequentialIsolation START ticks=300.5'},
    @{Name='OwnershipBackward';Index=96;Value='OWNERSHIP: SequentialIsolation END ticks=299'},
    @{Name='OwnershipBeforeEarlierScopes';Index=95;Value='OWNERSHIP: SequentialIsolation START ticks=289'},
    @{Name='OwnershipOverlap';Index=98;Value='OWNERSHIP: PlainControls START ticks=349'},
    @{Name='OwnershipCrossed';Index=96;Value='OWNERSHIP: PlainControls END ticks=350'},
    @{Name='OwnershipMalformed';Index=98;Value='OWNERSHIP: PlainControls START tick'},
    @{Name='OwnershipSuffix';Index=99;Value='OWNERSHIP: PlainControls END ticks=390 trailing'},
    @{Name='OwnershipPassWrong';Index=97;Value='PASS: P1I1PlainControls'}
)
foreach($m in $ownershipMutations){Case $m.Name {
    $v=[string[]]$clean.Clone();$v[$m.Index]=$m.Value;$r=Verify $v
    Need (-not $r.CaptureValid -and -not $r.Complete -and $r.Termination -ceq 'invalid' -and $r.OwnershipProgress.Count -eq 0) 'RejectOwnershipMutation'
}}
Case MissingOwnershipStart { $v=@($clean[0..94]+$clean[96..104]);Need (-not (Verify $v).CaptureValid) 'MissingOwnershipStart' }
Case MissingOwnershipEnd { $v=@($clean[0..95]+$clean[97..104]);Need (-not (Verify $v).CaptureValid) 'MissingOwnershipEnd' }
Case DuplicateOwnership { $v=@($clean[0..95]+$clean[95..104]);Need (-not (Verify $v).CaptureValid) 'DuplicateOwnership' }
Case OwnershipBeforeAPIClosed { $v=[string[]]$clean.Clone();$v[77]=$clean[95];Need (-not (Verify $v).CaptureValid) 'EarlyOwnership' }
Case OwnershipNestedInOtherBody { $v=[string[]]$clean.Clone();$v[96]=$clean[98];Need (-not (Verify $v).CaptureValid) 'NoNestedOwnership' }
Case Legacy101Rejected { $v=@($clean[0..94]+$clean[97]+$clean[100..104]);Need (-not (Verify $v).CaptureValid) 'Legacy101' }
Case ZeroOwnershipSpan { $v=[string[]]$clean.Clone();$v[96]='OWNERSHIP: SequentialIsolation END ticks=300';$r=Verify $v;Need ($r.Complete -and $r.ClosedPairs[9].ElapsedMs -eq 0) 'OwnershipZero' }
Case LongTicksAndTieToEven {
    $v=[string[]]$clean.Clone();$v[95]='OWNERSHIP: SequentialIsolation START ticks=9223372036854770000';$v[96]='OWNERSHIP: SequentialIsolation END ticks=9223372036854770005';$v[98]='OWNERSHIP: PlainControls START ticks=9223372036854770010';$v[99]='OWNERSHIP: PlainControls END ticks=9223372036854770025'
    $q=Receipt;$q.StopwatchFrequency=[long]10000000;$r=Verify $v $q
    Need ($r.Complete -and $r.ClosedPairs[9].ElapsedMs -eq 0 -and $r.ClosedPairs[10].ElapsedMs -eq 0.002) 'Int64SubtractionBeforeConversion'
}
Case ExplicitErrorDoesNotCreateEnd { $r=Verify -Lines $clean[0..95] -ErrorText 'ASSERTION FAILED';Need (-not $r.CaptureValid -and $r.OpenScopes.Count -eq 0 -and $r.OwnershipProgress.Count -eq 0) 'NoPartialInferenceAfterError' }
Case FullOutputTimeoutNotAccepted { $q=Receipt;$q.TimedOut=$true;$q.ExitCode=-1;$r=Verify -Receipt $q;Need (-not $r.Complete -and $r.Termination -ceq 'timeout') 'FullOutputStillNeedsExit' }
Case EveryOrderedPrefixPreservesIncomplete {
    foreach($last in 1..103){
        $q=Receipt;$q.TimedOut=$true;$q.ExitCode=-1;$q.ElapsedMs=30080
        $r=Verify $clean[0..$last] $q
        Need ($r.CaptureValid -and -not $r.Complete -and -not $r.FormalProjectAcceptance -and $r.Termination -ceq 'timeout' -and $r.LineCount -eq ($last+1)) ('Prefix/'+$last)
    }
}
foreach($field in @('ParentAdministratorRole','ChildLaunched','StreamsSaved','TimedOut')){
    Case ('StringBoolean/'+$field) {$q=Receipt;$q.$field='False';Need (-not (Verify -Receipt $q).CaptureValid) ('TypedBoolean/'+$field)}
}
foreach($value in @('0',0.5,$null,$true)){
    Case ('BadExit/'+[string]$value) {$q=Receipt;$q.ExitCode=$value;Need (-not (Verify -Receipt $q).CaptureValid) 'TypedExit'}
}
Case ZeroElapsed {$q=Receipt;$q.ElapsedMs=0;Need (-not (Verify -Receipt $q).CaptureValid) 'ZeroElapsed'}
Case NotLaunched {$q=Receipt;$q.ChildLaunched=$false;Need (-not (Verify -Receipt $q).CaptureValid) 'NotLaunched'}
Case StreamsNotSaved {$q=Receipt;$q.StreamsSaved=$false;Need (-not (Verify -Receipt $q).CaptureValid) 'StreamsNotSaved'}
Write-Output ('TOTAL: '+$count+' offline cases; no product execution or acceptance.')
