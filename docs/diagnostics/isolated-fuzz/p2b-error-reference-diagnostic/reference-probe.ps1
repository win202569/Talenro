# Fixed one-host reference measurement. Not an integration or recovery entry.
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$repo=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))
$packet=Join-Path $repo '.superpowers/sdd/2026-11-01-offline-error-reference-diagnostic'
$probePath=$PSCommandPath
$script:RetainedCases=[Collections.Generic.List[object]]::new()
$journal=$null;$journalFailed=$false;$journalBytes=0;$failure=$null
function Ref-Need([bool]$Condition,[string]$Message){if(-not $Condition){throw ('REF: '+$Message)}}
function Ref-Hash([string]$Path){(Get-FileHash -LiteralPath $Path -Algorithm SHA256).Hash}
function Ref-Project([Management.Automation.ErrorRecord]$Record){
    [ordered]@{Type=$Record.Exception.GetType().FullName;Message=$Record.Exception.Message;ErrorId=$Record.FullyQualifiedErrorId;Position=$Record.InvocationInfo.PositionMessage;Stack=$Record.ScriptStackTrace}
}
function Invoke-ODReferenceFault([string]$MissingPath){
    $script:CaseState.FaultCalls++
    [IO.File]::ReadAllText($MissingPath)
}
function Observe-ODReference([string]$Label,[Management.Automation.ErrorRecord]$Record){
    $state=$script:CaseState
    Ref-Need ($state.Records.Count -lt 2) 'two observation bound'
    $projection=Ref-Project $Record
    $expected='Exception calling "ReadAllText" with "1" argument(s): "Could not find file '''+$state.Missing+'''."'
    Ref-Need ($projection.Type -ceq 'System.Management.Automation.MethodInvocationException' -and $projection.ErrorId -ceq 'FileNotFoundException' -and $projection.Message -ceq $expected) 'actual fixed missing-file signature'
    Ref-Need ($projection.Position.Contains($probePath+':') -and $projection.Stack.Contains('Invoke-ODReferenceFault')) 'actual fixed fault source'
    $state.Records.Add($Record)
    $state.Exceptions.Add($Record.Exception)
    $state.Projections.Add($projection)
    $before=$(if($null -eq $script:HostPrimary){0}else{1})+$script:HostSecondary.Count
    $state.Events.Add([ordered]@{Label=$Label;BeforeCount=$before;AfterCount=$null;ObservationIndex=($state.Records.Count-1)})
}
function Ref-CollectObserved {
    $state=$script:CaseState
    Capture-Failure $state.Records[$state.Records.Count-1]
    $state.Events[$state.Events.Count-1].AfterCount=$(if($null -eq $script:HostPrimary){0}else{1})+$script:HostSecondary.Count
}
function Ref-Index([object]$Record){
    for($i=0;$i -lt $script:CaseState.Records.Count;$i++){
        if([object]::ReferenceEquals($Record,$script:CaseState.Records[$i])){return $i}
    }
    throw 'REF: collected error was not observed by reference'
}
function Ref-Case([string]$Name){
    $root=Join-Path $packet ('cases/'+$Name)
    Ref-Need (-not (Test-Path -LiteralPath $root)) 'case already attempted'
    $null=[IO.Directory]::CreateDirectory($root)
    $missing=Join-Path $root absent.txt
    Ref-Need (-not (Test-Path -LiteralPath $missing)) 'fixed fault file must be absent'
    $script:HostPrimary=$null
    $script:HostSecondary=[Collections.Generic.List[Management.Automation.ErrorRecord]]::new()
    $script:HostOverflow=$false
    $script:CaseState=[pscustomobject]@{Name=$Name;Missing=$missing;FaultCalls=0;Records=[Collections.Generic.List[object]]::new();Exceptions=[Collections.Generic.List[object]]::new();Projections=[Collections.Generic.List[object]]::new();Events=[Collections.Generic.List[object]]::new();Primary=$null;Secondary=$null}
    try{
        switch -CaseSensitive ($Name){
            'DirectOuter' {Invoke-ODReferenceFault $missing}
            'CatchBare' {
                try{Invoke-ODReferenceFault $missing}
                catch{$saved=$_;Observe-ODReference Inner $saved;Ref-CollectObserved;throw}
            }
            'CatchRetained' {
                try{Invoke-ODReferenceFault $missing}
                catch{$saved=$_;Observe-ODReference Inner $saved;Ref-CollectObserved;throw $saved}
            }
            'FinallyBare' {
                try{}finally{
                    try{Invoke-ODReferenceFault $missing}
                    catch{$saved=$_;Observe-ODReference Inner $saved;Ref-CollectObserved;throw}
                }
            }
            'FinallyRetained' {
                try{}finally{
                    try{Invoke-ODReferenceFault $missing}
                    catch{$saved=$_;Observe-ODReference Inner $saved;Ref-CollectObserved;throw $saved}
                }
            }
            'IndependentSameText' {
                foreach($label in @('First','Second')){
                    try{Invoke-ODReferenceFault $missing}
                    catch{$saved=$_;Observe-ODReference $label $saved;Ref-CollectObserved}
                }
            }
            default{throw 'REF: fixed cases only'}
        }
    }catch{$outer=$_;Observe-ODReference Outer $outer;Ref-CollectObserved}
    $state=$script:CaseState
    $expectedFaults=if($Name -ceq 'IndependentSameText'){2}else{1}
    $expectedLabels=if($Name -ceq 'IndependentSameText'){@('First','Second')}elseif($Name -ceq 'DirectOuter'){@('Outer')}else{@('Inner','Outer')}
    Ref-Need ($state.FaultCalls -eq $expectedFaults -and (($state.Events.Label -join ',') -ceq ($expectedLabels -join ','))) 'complete fixed events/fault calls'
    $state.Primary=$script:HostPrimary;$state.Secondary=$script:HostSecondary.ToArray()
    $count=$(if($null -eq $state.Primary){0}else{1})+$state.Secondary.Count
    Ref-Need ($count -ge 1 -and $count -le 2 -and -not $script:HostOverflow) 'bounded actual collector'
    $primaryIndex=Ref-Index $state.Primary
    $secondaryIndices=@(foreach($record in $state.Secondary){Ref-Index $record})
    if($Name -ceq 'DirectOuter'){Ref-Need ($count -eq 1 -and $primaryIndex -eq 0 -and $secondaryIndices.Count -eq 0) 'outer-only control'}
    if($Name -ceq 'IndependentSameText'){Ref-Need ($count -eq 2 -and $primaryIndex -eq 0 -and $secondaryIndices.Count -eq 1 -and $secondaryIndices[0] -eq 1) 'distinct actual faults cannot merge'}
    $recordMatrix=[Collections.Generic.List[object]]::new()
    $exceptionMatrix=[Collections.Generic.List[object]]::new()
    for($i=0;$i -lt $state.Records.Count;$i++){
        $recordRow=[Collections.Generic.List[bool]]::new();$exceptionRow=[Collections.Generic.List[bool]]::new()
        for($j=0;$j -lt $state.Records.Count;$j++){
            $recordRow.Add([object]::ReferenceEquals($state.Records[$i],$state.Records[$j]))
            $exceptionRow.Add([object]::ReferenceEquals($state.Exceptions[$i],$state.Exceptions[$j]))
        }
        $recordMatrix.Add($recordRow.ToArray());$exceptionMatrix.Add($exceptionRow.ToArray())
    }
    for($i=0;$i -lt $state.Records.Count;$i++){
        Ref-Need ($recordMatrix[$i][$i] -and $exceptionMatrix[$i][$i]) 'actual identity diagonals'
        for($j=0;$j -lt $state.Records.Count;$j++){Ref-Need ($recordMatrix[$i][$j] -eq $recordMatrix[$j][$i] -and $exceptionMatrix[$i][$j] -eq $exceptionMatrix[$j][$i]) 'symmetric actual reference matrices'}
    }
    $script:RetainedCases.Add($state)
    [ordered]@{Schema=1;Case=$Name;FaultCalls=$state.FaultCalls;Events=$state.Events.ToArray();RecordSame=$recordMatrix.ToArray();ExceptionSame=$exceptionMatrix.ToArray();Collector=[ordered]@{PrimaryIndex=$primaryIndex;SecondaryIndices=$secondaryIndices;Overflow=$script:HostOverflow};ErrorProjections=$state.Projections.ToArray()}
}
try{
    Ref-Need ($PSVersionTable.PSVersion.ToString() -ceq '7.6.5' -and $PSVersionTable.PSEdition -ceq 'Core' -and [Environment]::Is64BitProcess) 'pinned PS7 context'
    $bindingPath=Join-Path $packet binding.json
    $intent=Get-Content -LiteralPath (Join-Path $packet attempt.intent.json) -Raw|ConvertFrom-Json
    Ref-Need ($intent.AttemptConsumed -is [bool] -and $intent.AttemptConsumed -and $intent.BindingSHA256 -ceq (Ref-Hash $bindingPath) -and $intent.DiagnosticHosts -eq 1 -and $intent.Targets -eq 0 -and $intent.Compilations -eq 0 -and $intent.Reviews -eq 0) 'actual one-shot diagnostic intent'
    $binding=Get-Content -LiteralPath $bindingPath -Raw|ConvertFrom-Json
    Ref-Need ($binding.Inputs.Count -eq 8 -and $binding.Identity -ceq 'offline-error-reference-diagnostic-2026-11-01') 'eight fixed diagnostic inputs'
    foreach($pin in $binding.Inputs){Ref-Need ((Ref-Hash $pin.Path) -ceq $pin.SHA256) 'diagnostic input drift'}
    $expectedPaths=@(
        [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '.gitattributes')),
        $probePath,
        (Join-Path $repo 'docs/diagnostics/isolated-fuzz/p2b-integrated-review-repair/capture.ps1'),
        (Join-Path $repo 'docs/roadmap/2026-11-01-offline-error-reference-readonly-analysis.md'),
        (Join-Path $repo 'docs/superpowers/plans/2026-11-01-offline-error-reference-diagnostic.md'),
        'C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe',
        'C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/System.Management.Automation.dll',
        (Join-Path $packet approval.json)
    )
    Ref-Need (@(Compare-Object @($expectedPaths|ForEach-Object {[IO.Path]::GetFullPath($_)}|Sort-Object) @($binding.Inputs.Path|ForEach-Object {[IO.Path]::GetFullPath($_)}|Sort-Object)).Count -eq 0) 'fixed diagnostic file set'
    $journal=[IO.File]::Open((Join-Path $packet observations.jsonl),[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::Read)
    $capture=Join-Path $repo 'docs/diagnostics/isolated-fuzz/p2b-integrated-review-repair/capture.ps1'
    Ref-Need ((Ref-Hash $capture) -ceq 'F8BC1C86C20A5C015C45F91312474CBC9048F6E481C6F0A4B281B969E429F8F2') 'frozen collector'
    $tokens=$null;$errors=$null
    $ast=[Management.Automation.Language.Parser]::ParseFile($capture,[ref]$tokens,[ref]$errors)
    $collectors=@($ast.FindAll({param($n)$n -is [Management.Automation.Language.FunctionDefinitionAst] -and $n.Name -ceq 'Capture-Failure'},$false))
    Ref-Need ($errors.Count -eq 0 -and $collectors.Count -eq 1) 'unique pure collector definition'
    . ([scriptblock]::Create($collectors[0].Extent.Text))
    foreach($case in @('DirectOuter','CatchBare','CatchRetained','FinallyBare','FinallyRetained','IndependentSameText')){
        $row=Ref-Case $case
        $bytes=[Text.UTF8Encoding]::new($false).GetBytes((ConvertTo-Json -InputObject $row -Depth 20 -Compress)+"`n")
        Ref-Need ($journalBytes+$bytes.Length -le 1048576) 'diagnostic journal byte bound'
        try{$journal.Write($bytes,0,$bytes.Length);$journal.Flush();$journalBytes+=$bytes.Length}catch{$journalFailed=$true;throw}
    }
    $final=[ordered]@{Schema=1;Status='COMPLETE';Cases=6;FaultCalls=7;SourceVerified=$false;LiveActionsPerformed=$false;RuntimeApproved=$false;FormalAcceptance=$false;RealGateImplemented=$false}
    $bytes=[Text.UTF8Encoding]::new($false).GetBytes((ConvertTo-Json -InputObject $final -Compress)+"`n")
    Ref-Need (-not $journalFailed -and $journalBytes+$bytes.Length -le 1048576) 'healthy final journal'
    try{$journal.Write($bytes,0,$bytes.Length);$journal.Flush()}catch{$journalFailed=$true;throw}
}catch{$failure=$_}
finally{
    if($null -ne $journal){try{$journal.Dispose()}catch{$journalFailed=$true;if($null -eq $failure){$failure=$_}}}
}
if($null -ne $failure){try{[Console]::Error.WriteLine($failure.ToString()+"`n"+$failure.ScriptStackTrace)}catch{};exit 1}
if($journalFailed){exit 1}
exit 0
