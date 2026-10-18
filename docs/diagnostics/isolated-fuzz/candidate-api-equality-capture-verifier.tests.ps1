# Offline verifier tests only: never imports or launches the policy/product.
$ErrorActionPreference='Stop'
function Need([bool]$Value,[string]$Case){if(-not $Value){throw ('ASSERTION FAILED: '+$Case)}}
$module=Join-Path $PSScriptRoot 'candidate-api-equality-capture-verifier.ps1'
if(Test-Path $module){. $module}
$inventory=Get-Content (Join-Path $PSScriptRoot 'integer-candidate-api-equality/expected-output.json') -Raw|ConvertFrom-Json
$ticks=@(100,110,120,140,150,160,190,200,205,210,220,230);$k=0
$clean=@(foreach($rule in $inventory.Lines){
    switch -CaseSensitive ($rule.Kind){
        Context {'CONTEXT: SYNTHETIC\offline admin=False'}
        Host {'HOST: 7.6.5 Core x64=True'}
        Literal {$rule.Value}
        Boundary {('BOUNDARY: '+$rule.Name+' '+$rule.Edge+' ticks='+$ticks[$k]);$k++}
        Subphase {('SUBPHASE: '+$rule.Name+' '+$rule.Edge+' ticks='+$ticks[$k]);$k++}
    }
})
function Receipt { [pscustomobject]@{StopwatchFrequency=[long]1000;ParentAdministratorRole=$false;ChildLaunched=$true;StreamsSaved=$true;TimedOut=$false;ExitCode=0;ElapsedMs=1000;CPUMs=1500;CleanupStatus='owned-host-exited';CaptureValid=$false;ExpectedResult=$false} }
function Verify($Lines=$clean,$Receipt=(Receipt),[string]$ErrorText=''){
    Need ($null -ne (Get-Command Test-CandidateAPIEqualityCapture -ErrorAction SilentlyContinue)) 'VerifierAvailable'
    Test-CandidateAPIEqualityCapture -Stdout (($Lines -join "`n")+"`n") -Stderr $ErrorText -Receipt $Receipt
}
$count=0
function Case([string]$Name,[scriptblock]$Body){& $Body;$script:count++;Write-Output ('PASS: SplitVerifier/'+$Name)}
Case NormalPS7NotLegacyClassifier {
    $r=Verify
    Need ($r.CaptureValid -and $r.Complete -and $r.Status -ceq 'complete-diagnostic') 'NormalPS7'
    Need ($r.LineCount -eq 95 -and $r.PassCount -eq 64 -and $r.MarkerCount -eq 12 -and $r.ClosedPairs.Count -eq 6) 'CompleteCounts'
    Need (-not $r.FormalProjectAcceptance) 'NoAcceptance'
    Need ($r.ClosedPairs[0].ElapsedMs -eq 10 -and $r.ClosedPairs[1].ElapsedMs -eq 20 -and $r.ClosedPairs[2].ElapsedMs -eq 30 -and $r.ClosedPairs[3].ElapsedMs -eq 5 -and $r.ClosedPairs[4].ElapsedMs -eq 60 -and $r.ClosedPairs[5].ElapsedMs -eq 10) 'LiteralPairDurations'
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
Case RemovedMarker {$v=@($clean[0..65]+$clean[67..94]);$r=Verify $v;Need (-not $r.CaptureValid -and -not $r.Complete) 'RemovedMarker'}
Case DuplicateMarker {$v=@($clean[0..65]+$clean[65..94]);$r=Verify $v;Need (-not $r.CaptureValid) 'DuplicateMarker'}
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
Case MissingTailNotComplete {$r=Verify $clean[0..93];Need ($r.CaptureValid -and -not $r.Complete) 'MissingTail'}
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
Case NoFinalNewline {Need (-not (Test-CandidateAPIEqualityCapture -Stdout ($clean -join "`n") -Stderr '' -Receipt (Receipt)).CaptureValid) 'UnterminatedTail'}
Case CRLFWorks {$r=Test-CandidateAPIEqualityCapture -Stdout (($clean -join "`r`n")+"`r`n") -Stderr '' -Receipt (Receipt);Need $r.Complete 'CRLF'}

Case InnerOrderRejects {$v=[string[]]$clean.Clone();$v[70]=$clean[71];$v[71]=$clean[70];Need (-not (Verify $v).CaptureValid) 'InnerOrder'}
Case InnerBeforeParentTick {$v=[string[]]$clean.Clone();$v[70]='SUBPHASE: Evaluate4096 START ticks=149';Need (-not (Verify $v).CaptureValid) 'ContainmentStart'}
Case ParentEndsBeforeChild {$v=[string[]]$clean.Clone();$v[74]='BOUNDARY: Check4096 END ticks=204';Need (-not (Verify $v).CaptureValid) 'ContainmentEnd'}
Case SiblingBackwards {$v=[string[]]$clean.Clone();$v[72]='SUBPHASE: Assert4096 START ticks=189';Need (-not (Verify $v).CaptureValid) 'SiblingTicks'}
Case PrematureParentEnd {$v=[string[]]$clean.Clone();$v[71]='BOUNDARY: Check4096 END ticks=190';Need (-not (Verify $v).CaptureValid) 'ParentStack'}
Case DuplicateInnerEnd {$v=[string[]]$clean.Clone();$v[72]=$clean[71];Need (-not (Verify $v).CaptureValid) 'DuplicateInner'}
Case InnerMalformedTick {$v=[string[]]$clean.Clone();$v[70]='SUBPHASE: Evaluate4096 START ticks=9223372036854775808';Need (-not (Verify $v).CaptureValid) 'InnerOverflow'}
Case ParentStillOpenAfterEval {
    $r=Verify $clean[0..71]
    Need ($r.CaptureValid -and -not $r.Complete -and $r.ClosedPairs.Count -eq 3 -and $r.OpenScopes.Count -eq 1 -and $r.OpenScopes[0].Name -ceq 'Check4096') 'OpenParent'
}
Case AssertStillOpen {
    $r=Verify $clean[0..72]
    Need ($r.CaptureValid -and -not $r.Complete -and $r.ClosedPairs.Count -eq 3 -and $r.OpenScopes.Count -eq 2 -and $r.OpenScopes[1].Name -ceq 'Assert4096') 'OpenAssertion'
}
Case ParentOpenAfterAssert {
    $r=Verify $clean[0..73]
    Need ($r.CaptureValid -and -not $r.Complete -and $r.ClosedPairs.Count -eq 4 -and $r.OpenScopes.Count -eq 1) 'ParentUnfinished'
}
Case ZeroSpanIsValid {$v=[string[]]$clean.Clone();$v[71]='SUBPHASE: Evaluate4096 END ticks=160';$r=Verify $v;Need ($r.Complete -and $r.ClosedPairs[2].ElapsedMs -eq 0) 'ZeroSpan'}
Case Legacy91Rejected {$v=@($clean[0..69]+$clean[74..94]);Need (-not (Verify $v).CaptureValid) 'Legacy91'}
foreach($value in @([double]::NaN,[double]::PositiveInfinity,-1,'1500')){
    Case ('BadCPU/'+[string]$value) {$q=Receipt;$q.CPUMs=$value;Need (-not (Verify -Receipt $q).CaptureValid) 'CPU'}
}

Write-Output ('TOTAL: '+$count+' offline cases; no product execution or acceptance.')
