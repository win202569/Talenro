# DEFINITIONS ONLY. Await complete immutable fixtures/helpers before RED25.
# No source imports, filesystem writes or test execution at module load.
function RT-Need([bool]$Condition,[string]$Label){
    if(-not $Condition){throw ('Runtime binding regression: '+$Label)}
}
function RT-Equal($Left,$Right){
    return ((ConvertTo-Json -InputObject $Left -Depth 100 -Compress) -ceq (ConvertTo-Json -InputObject $Right -Depth 100 -Compress))
}
function RT-Copy($Value){
    return ConvertFrom-Json -AsHashtable -InputObject (ConvertTo-Json -InputObject $Value -Depth 100)
}
function RT-Error($Record){
    if($null -eq $Record){return $null}
    return [ordered]@{Type=$Record.Exception.GetType().FullName;Message=$Record.Exception.Message;ErrorId=$Record.FullyQualifiedErrorId;Position=$Record.InvocationInfo.PositionMessage;Stack=$Record.ScriptStackTrace}
}
function RT-Tree([string]$Text){
    $tokens=$null;$errors=$null
    $tree=[Management.Automation.Language.Parser]::ParseInput($Text,[ref]$tokens,[ref]$errors)
    RT-Need ($errors.Count -eq 0) 'source parse'
    return $tree
}
function RT-Definition($Tree,[string]$Name){
    $nodes=@($Tree.FindAll({param($n)
        $n -is [Management.Automation.Language.FunctionDefinitionAst] -and $n.Name -ceq $Name
    },$true))
    RT-Need ($nodes.Count -eq 1) ('unique definition '+$Name)
    return $nodes[0]
}
function RT-Calls($Tree,[string]$Name){
    # Deliberately emit individual nodes, never a comma-wrapped array.
    foreach($node in $Tree.FindAll({param($n)
        $n -is [Management.Automation.Language.CommandAst] -and $n.GetCommandName() -ceq $Name
    },$true)){$node}
}
function RT-SingleCall($Tree,[string]$Name){
    $calls=@(RT-Calls $Tree $Name)
    RT-Need ($calls.Count -eq 1) ('unique call '+$Name)
    RT-Need ($calls[0] -is [Management.Automation.Language.CommandAst]) ('scalar command '+$Name)
    RT-Need ($calls[0].Extent.StartOffset -is [int]) ('integer call offset '+$Name)
    return $calls[0]
}
function RT-ReadOnlyDefinition($Tree,[string]$Name){
    $allowed=@('Get-ODAttemptSlot','Get-ODEpochManifestName','Driver-Hash',
        'Driver-CheckHash','Driver-Need','Driver-EqualJSON','Driver-Keys',
        'Driver-VerifyHistory','Driver-ValidateRuntimeInputs','Driver-RequireFinal',
        'Get-ODWriterManifestPath','Driver-FalseFlags','Assert-ODRuntimeLineage')
    RT-Need ($Name -cin $allowed) ('extraction allowlist '+$Name)
    $definition=RT-Definition $Tree $Name
    # Return only a scriptblock; the caller dot-sources it in its local scope.
    # Gate, receipt orchestration, acquisition and Prepare are never extractable.
    return [scriptblock]::Create(([string][char]10 * ($definition.Extent.StartLineNumber-1))+$definition.Extent.Text)
}
function RT-Reject([scriptblock]$Body,[string]$Message){
    $caught=$null;$emitted=@()
    try{$emitted=@(& $Body)}catch{$caught=$_}
    RT-Need ($null -ne $caught -and $emitted.Count -eq 0) 'silent validation rejection'
    RT-Need ($caught.Exception.GetType().FullName -ceq 'System.Management.Automation.RuntimeException' -and $caught.Exception.Message -ceq $Message -and $caught.FullyQualifiedErrorId -ceq $Message) ('precise rejection '+$Message)
}
function RT-Case([string]$Name,[scriptblock]$Body,$State){
    if($State.Stopped){$State.Groups.Add([ordered]@{Name=$Name;Status='NotRun'});return}
    try{
        $emitted=@(& $Body)
        RT-Need ($emitted.Count -eq 0) ('silent group '+$Name)
        $State.Groups.Add([ordered]@{Name=$Name;Status='PASS'})
    }catch{
        $State.RootError=$_;$State.Stopped=$true
        $State.Groups.Add([ordered]@{Name=$Name;Status='FAIL'})
    }
}
function Get-RTGroupNames {
    @('NewEpochBinding','Green24Identity','HistoricalEpoch','LiveCandidate',
      'ApprovalIdentity','PriorFailurePreserved','ExternalHistory','RelativeHistory',
      'ParentLineage','WriterIdentity','InvalidReceiptRejected','InvalidCandidateRejected',
      'EpochRejected','DisjointSlots','ScalarCallSelection','AttemptOwnership',
      'ActualGateCall','SourceWindow','OriginalGuardsUnchanged','OriginalSuiteUnchanged')
}
function Invoke-ODRuntimeBindingRegression {
    param([Parameter(Mandatory)][ValidateSet('Red','Green')][string]$Phase)
    Set-StrictMode -Version Latest;$ErrorActionPreference='Stop'
    $repo=(Get-Location).Path;$l='docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2'
    $p='.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner';$packet=Join-Path $repo $p
    $names=@(Get-RTGroupNames)
    $state=[pscustomobject]@{Groups=[Collections.Generic.List[object]]::new();Stopped=$false;RootError=$null}
    $setup=$null;$recording=$null;$before=[ordered]@{};$after=[ordered]@{};$historyBefore=[ordered]@{};$historyAfter=[ordered]@{}
    $binding=$null;$sources=$null;$testHash=$null;$inputHash=$null;$phaseHash=$null;$code=2
    try{
        $testPath=Join-Path $repo ($l+'/history-contract-runtime.tests.ps1');$inputPath=Join-Path $packet 'history-contract-runtime.input.json'
        $testHash=(Get-FileHash -LiteralPath $testPath).Hash;$inputHash=(Get-FileHash -LiteralPath $inputPath).Hash
        $binding=Get-Content -LiteralPath $inputPath -Raw|ConvertFrom-Json -AsHashtable
        RT-Need ($binding.TestSHA256 -ceq $testHash) 'immutable test identity'
        $sources=$binding.BaselineSources
        if($Phase -ceq 'Green'){
            $candidatePath=Join-Path $packet 'history-contract-runtime.candidate.input.json';$phaseHash=(Get-FileHash $candidatePath).Hash
            $candidate=Get-Content $candidatePath -Raw|ConvertFrom-Json -AsHashtable
            RT-Need ($candidate.InputSHA256 -ceq $inputHash) 'candidate identity';$sources=$candidate.CandidateSources
        }
        $intent=Get-Content (Join-Path $packet ('history-contract-runtime-binding-'+$Phase.ToLowerInvariant()+'.intent.json')) -Raw|ConvertFrom-Json -AsHashtable
        RT-Need ($intent.AttemptConsumed -eq $true -and $intent.Computation -eq $(if($Phase -ceq 'Red'){25}else{26}) -and $intent.Ceiling -eq 26 -and $intent.TestSHA256 -ceq $testHash -and $intent.InputSHA256 -ceq $inputHash -and $intent.PhaseInputSHA256 -ceq $phaseHash) 'actual finite intent'
        foreach($path in $sources.Keys){$before[$path]=(Get-FileHash -LiteralPath $path).Hash;RT-Need ($before[$path] -ceq $sources[$path]) ('source '+$path)}
        foreach($path in $binding.HistoricalFiles.Keys){$historyBefore[$path]=(Get-FileHash -LiteralPath $path).Hash;RT-Need ($historyBefore[$path] -ceq $binding.HistoricalFiles[$path]) ('history '+$path)}
        $texts=@{};$trees=@{};$oldTrees=@{}
        foreach($file in @('driver.ps1','capture.ps1','source-model.ps1','observation-contract.tests.ps1')){
            $texts[$file]=[IO.File]::ReadAllText((Join-Path $repo ($l+'/'+$file)));$trees[$file]=RT-Tree $texts[$file]
            $oldTrees[$file]=RT-Tree ([Text.Encoding]::UTF8.GetString([Convert]::FromBase64String($binding.BaselineSourceBytes[$l+'/'+$file])))
        }
        # Missing new epoch, not a missing function/setup error, is the sole expected RED.
        $d=RT-Definition $trees['driver.ps1'] Get-ODAttemptSlot
        . ([scriptblock]::Create(([string][char]10 * ($d.Extent.StartLineNumber-1))+$d.Extent.Text))
        RT-Case $names[0] {RT-Need ((Get-ODAttemptSlot contract-green history-contract-runtime) -ceq 'history-contract-runtime-host') 'new physical slot'} $state
        if(-not $state.Stopped){
            foreach($name in @('Get-ODEpochManifestName','Driver-Hash','Driver-CheckHash','Driver-Need','Driver-EqualJSON','Driver-Keys','Driver-VerifyHistory','Driver-ValidateRuntimeInputs','Assert-ODRuntimeLineage','Driver-RequireFinal','Get-ODWriterManifestPath','Driver-FalseFlags')){
                $file=if($name -cin @('Get-ODEpochManifestName','Assert-ODRuntimeLineage')){'source-model.ps1'}else{'driver.ps1'}
                . (RT-ReadOnlyDefinition $trees[$file] $name)
            }
            $approval=Get-Content (Join-Path $packet 'history-contract-runtime.approval.json') -Raw|ConvertFrom-Json -AsHashtable
            $fixture=$binding.FixtureBindings
            $parent=Get-Content (Join-Path $repo $fixture.Parent.Path) -Raw|ConvertFrom-Json -AsHashtable
            RT-Case $names[1] {
                $value=Driver-ValidateRuntimeInputs $repo $approval $binding $candidate
                RT-Need ((RT-Equal @($value.Keys) @('Parent','WriterArtifacts','HistoricalGreen24','CandidateSources')) -and (RT-Equal $value.Parent $parent) -and (RT-Equal $value.HistoricalGreen24 $binding.Green24Artifacts.SHA256)) 'authenticated actual GREEN24'
                $r=$binding.Green24Artifacts.Records.Final
                RT-Need ($r.PASS -eq 20 -and $r.FAIL -eq 0 -and $r.NotRun -eq 0 -and $r.ExitCode -eq 0 -and $null -eq $r.RootError -and $null -eq $r.SetupError -and $null -eq $r.RecordingError) 'saved complete pass'
                RT-Need ($binding.Green24Artifacts.SHA256[$p+'/history-contract-v2-green.final.json'] -ceq 'AB86CCC7042089B0D705E9FB5933B4ADD92045514B5B976068315693F7CCFC8A') 'literal original receipt'
            } $state
            RT-Case $names[2] {
                $r=$binding.Green24Artifacts.Records.Final
                RT-Need ((RT-Equal $r.SourceSHA256 $binding.BaselineSources) -and (RT-Equal $r.PostSourceSHA256 $binding.BaselineSources)) 'old receipt retains old source epoch'
                foreach($path in $binding.BaselineSourceBytes.Keys){
                    $bytes=[Convert]::FromBase64String($binding.BaselineSourceBytes[$path])
                    RT-Need ([Convert]::ToHexString([Security.Cryptography.SHA256]::HashData($bytes)) -ceq $r.SourceSHA256[$path]) 'actual preserved baseline bytes'
                }
                RT-Need (-not (RT-Equal $r.SourceSHA256 $candidate.CandidateSources)) 'old success is not new candidate proof'
            } $state
            RT-Case $names[3] {
                $v=Driver-ValidateRuntimeInputs $repo $approval $binding $candidate
                RT-Need ((RT-Equal $v.CandidateSources $candidate.CandidateSources) -and $v.CandidateSources.Count -eq 31) 'actual live source map'
            } $state
            RT-Case $names[4] {
                foreach($field in @('Schema','UserReply','AddedAuxiliary')){
                    $bad=RT-Copy $approval;$bad[$field]=if($field -ceq 'UserReply'){'wrong'}else{99}
                    RT-Reject {Driver-ValidateRuntimeInputs $repo $bad $binding $candidate} 'OD: runtime approval mismatch'
                }
            } $state
            RT-Case $names[5] {
                RT-Need ($fixture.Stops.Count -eq 13) 'all13old stops'
                foreach($path in $fixture.Stops.Keys){RT-Need ((Get-FileHash -LiteralPath $path).Hash -ceq $fixture.Stops[$path]) 'old stop retained'}
                $r=Get-Content (Join-Path $packet 'history-contract-binding-green.final.json') -Raw|ConvertFrom-Json -AsHashtable
                RT-Need ($r.PASS -eq 14 -and $r.FAIL -eq 1 -and $r.NotRun -eq 5 -and $r.RootError.ErrorId -ceq 'ComparisonFailure') 'GREEN23 remains failed'
            } $state
            RT-Case $names[6] {
                RT-Need ($fixture.ExternalHistory.Count -eq 3) 'three actual externals'
                $o=@(Driver-VerifyHistory $repo $fixture.ExternalHistory);RT-Need ($o.Count -eq 0) 'silent external verification'
                $bad=[ordered]@{'C:/not-approved-runtime-history.bin'=('0'*64)}
                RT-Reject {Driver-VerifyHistory $repo $bad} 'OD: historical path binding mismatch'
            } $state
            RT-Case $names[7] {
                $path=$l+'/expected-output.json';$map=[ordered]@{};$map[$path]=$binding.BaselineSources[$path]
                RT-Need ($map[$path] -ceq '0FA169E9C7D69EB4D53AFF6C20378ADE4DB43E6767880F2991F26A7B87C69BBF') 'literal relative fixture'
                $o=@(Driver-VerifyHistory $repo $map);RT-Need ($o.Count -eq 0) 'silent relative verification'
                $bad=[ordered]@{'../outside-runtime-history.bin'=('0'*64)}
                RT-Reject {Driver-VerifyHistory $repo $bad} 'OD: historical path binding mismatch'
            } $state
            RT-Case $names[8] {
                RT-Need ((Get-FileHash (Join-Path $repo $fixture.Parent.Path)).Hash -ceq 'A9949DD46C7F928CFE7DB33F092AF0497E6F89556E8F359A8A55CB4484BB1556') 'real repair parent'
                $bad=RT-Copy $binding;$bad.FixtureBindings.Parent.SHA256=$fixture.WriterSHA256
                RT-Reject {Driver-ValidateRuntimeInputs $repo $approval $bad $candidate} 'OD: runtime parent mismatch'
            } $state
            RT-Case $names[9] {
                RT-Need ($parent.PreviousFrozenSHA256 -ceq '182B85886304BFE3CBFB248DDA70F2EB99DF3E2158F574DBE299519DE7CE0461') 'actual parent writer link'
                $bad=RT-Copy $binding;$bad.FixtureBindings.WriterSHA256='0'*64
                RT-Reject {Driver-ValidateRuntimeInputs $repo $approval $bad $candidate} 'OD: runtime writer mismatch'
            } $state
            RT-Case $names[10] {
                foreach($kind in @('Computation','SourceEpoch','PassCount','ToolIdentity')){
                    $bad=RT-Copy $binding
                    switch($kind){
                        Computation {$bad.Green24Artifacts.Records.Intent.Computation=23}
                        SourceEpoch {$bad.Green24Artifacts.Records.Final.SourceSHA256[$l+'/driver.ps1']='0'*64}
                        PassCount {$bad.Green24Artifacts.Records.Final.PASS=19}
                        ToolIdentity {$bad.Green24Artifacts.Records.Tool.Input.cmd='not the recorded command'}
                    }
                    RT-Reject {Driver-ValidateRuntimeInputs $repo $approval $bad $candidate} 'OD: runtime receipt mismatch'
                }
            } $state
            RT-Case $names[11] {
                $key=$l+'/driver.ps1';$bad=RT-Copy $candidate;$bad.CandidateSources[$key]='0'*64
                RT-Reject {Driver-ValidateRuntimeInputs $repo $approval $binding $bad} 'OD: runtime candidate mismatch'
                $bad=RT-Copy $candidate;$bad.CandidateSources.Remove($key)
                RT-Reject {Driver-ValidateRuntimeInputs $repo $approval $binding $bad} 'OD: runtime candidate mismatch'
            } $state
            RT-Case $names[12] {
                foreach($epoch in @('','unknown','History-Contract-Runtime')){
                    RT-Reject {Get-ODAttemptSlot contract-green $epoch} 'OD: invalid explicit epoch'
                    RT-Reject {Get-ODEpochManifestName $epoch} 'OD: invalid explicit epoch'
                }
                foreach($grant in @('writer-green','core','lifecycle','continuation')){RT-Reject {Get-ODAttemptSlot $grant history-contract-runtime} 'OD: invalid explicit grant'}
                RT-Need ((Get-ODEpochManifestName history-contract-runtime) -ceq 'history-contract-runtime.frozen.json') 'literal runtime manifest'
            } $state
            RT-Case $names[13] {
                $slots=@('history-contract-runtime','history-contract-runtime-host','history-contract-runtime-binding-red','history-contract-runtime-binding-green')
                $paths=@(foreach($slot in $slots){foreach($suffix in @('intent.json','final.json','tool-record.json')){Join-Path $packet ($slot+'.'+$suffix)}})
                RT-Need (@($paths|Select-Object -Unique).Count -eq 12) 'disjoint namespace'
                RT-Need ((Test-Path (Join-Path $packet 'history-contract.stop.json')) -and -not (Test-Path (Join-Path $packet 'history-contract-runtime-host.intent.json'))) 'old occupied/new absent'
            } $state
            RT-Case $names[14] {
                $gate=RT-Definition $trees['driver.ps1'] Driver-RuntimeGate
                $null=RT-SingleCall $gate Driver-ValidateRuntimeInputs
                $acquisition=RT-SingleCall $trees['driver.ps1'] Driver-AcquireRuntimeAttempt
                $engine=@(RT-Calls $trees['driver.ps1'] Driver-CheckHash|Where-Object {$_.Extent.Text -match '^Driver-CheckHash \$engine '})
                RT-Need ($engine.Count -eq 1 -and $engine[0] -is [Management.Automation.Language.CommandAst] -and $engine[0].Extent.StartOffset -is [int] -and $acquisition.Extent.StartOffset -lt $engine[0].Extent.StartOffset) 'scalar acquisition before engine checks'
            } $state
            RT-Case $names[15] {
                $acquire=RT-Definition $trees['driver.ps1'] Driver-AcquireRuntimeAttempt
                RT-Need ($acquire.Extent.Text.Contains('[IO.FileMode]::CreateNew') -and $acquire.Extent.Text.Contains('[guid]::NewGuid()')) 'exclusive ownership'
                RT-Need (RT-Equal @($acquire.Parameters|ForEach-Object {$_.Name.VariablePath.UserPath}) @('Action','Grant')) 'no public path/token'
                $gate=RT-Definition $trees['driver.ps1'] Driver-RuntimeGate
                RT-Need ($gate.Extent.Text.Contains('$owned.Token -ceq $Ownership.Token') -and $gate.Extent.Text.Contains('[object]::ReferenceEquals($Ownership,$runtimeOwnership)')) 'local-only ownership'
                $calls=@(RT-Calls $gate Test-Path|Where-Object {$_.Extent.Text -ceq 'Test-Path -LiteralPath $occupiedPath'})
                RT-Need ($calls.Count -eq 1) 'actual occupancy expression'
                $expression=[scriptblock]::Create($calls[0].Extent.Text)
                foreach($rel in @('history-contract.stop.json','snapshots/repair-green')){$occupiedPath=Join-Path $packet $rel;RT-Need ([bool](& $expression)) 'occupied file/directory rejected'}
                $occupiedPath=Join-Path $packet 'history-contract-runtime.intent.json'
                RT-Need (-not [bool](& $expression)) 'fresh path absent'
            } $state
            RT-Case $names[16] {
                $gate=RT-Definition $trees['driver.ps1'] Driver-RuntimeGate
                $call=RT-SingleCall $gate Driver-ValidateRuntimeInputs
                $a=$approval;$b=$binding;$c=$candidate
                $v=& ([scriptblock]::Create($call.Extent.Text))
                RT-Need ((RT-Equal $v.Parent $parent) -and (RT-Equal $v.CandidateSources $candidate.CandidateSources) -and (RT-Equal $v.HistoricalGreen24 $binding.Green24Artifacts.SHA256)) 'real gate validation results'
                $unit=RT-Definition $trees['source-model.ps1'] New-ODIntentUnitContext
                $null=RT-SingleCall $unit Assert-ODRuntimeLineage
                RT-Need ($texts['capture.ps1'].Contains("'history-contract-runtime-host'")) 'literal capture slot'
                foreach($item in @(@('capture.ps1','Invoke-ODContractTests'),@('observation-contract.tests.ps1','New-ODIntentUnitContext'))){
                    $calls=@(RT-Calls $trees[$item[0]] $item[1]);RT-Need ($calls.Count -eq 1 -and $calls[0].Extent.Text -match '\$Epoch\b') 'explicit runtime epoch flow'
                }
            } $state
            RT-Case $names[17] {
                $changed=@($sources.Keys|Where-Object {$sources[$_] -cne $binding.BaselineSources[$_]})
                RT-Need ($sources.Count -eq 31 -and $changed.Count -eq 4 -and @(Compare-Object $changed @($binding.BaselineSourceBytes.Keys)).Count -eq 0) 'four-source window; other27 unchanged'
            } $state
            RT-Case $names[18] {
                foreach($pair in @(@('driver.ps1','Driver-VerifyHistory'),@('source-model.ps1','Get-ODFailureExtent'),@('driver.ps1','Driver-Recorder'),@('driver.ps1','Driver-FalseFlags'))){
                    RT-Need ((RT-Definition $trees[$pair[0]] $pair[1]).Extent.Text -ceq (RT-Definition $oldTrees[$pair[0]] $pair[1]).Extent.Text) ('unchanged guard '+$pair[1])
                }
                $old=[Text.Encoding]::UTF8.GetString([Convert]::FromBase64String($binding.BaselineSourceBytes[$l+'/driver.ps1']))
                $start='$process=[Diagnostics.Process]::new()';$nl=[string][char]10
                $oldTail=$old.Substring($old.IndexOf($start));$oldTail=$oldTail.Substring(0,$oldTail.LastIndexOf($nl+'}catch{'))
                $newTail=$texts['driver.ps1'].Substring($texts['driver.ps1'].IndexOf($start));$newTail=$newTail.Substring(0,$newTail.LastIndexOf($nl+'}catch{'))
                RT-Need ($oldTail -ceq $newTail) 'owned process unchanged'
            } $state
            RT-Case $names[19] {
                $old=[Text.Encoding]::UTF8.GetString([Convert]::FromBase64String($binding.BaselineSourceBytes[$l+'/observation-contract.tests.ps1']))
                $oldLine="    if("+ '$Epoch' +" -cnotin @('legacy','extent-contract','history-contract')){throw 'OD: invalid explicit epoch'}"
                $newLine="    if("+ '$Epoch' +" -cnotin @('legacy','extent-contract','history-contract','history-contract-runtime')){throw 'OD: invalid explicit epoch'}"
                RT-Need ($texts['observation-contract.tests.ps1'].Replace($newLine,$oldLine) -ceq $old) 'all49suite cases unchanged outside entry binding'
                RT-Need ((RT-Definition $trees['driver.ps1'] Driver-Recorder).Extent.Text -ceq (RT-Definition $oldTrees['driver.ps1'] Driver-Recorder).Extent.Text) 'original3recorder groups unchanged'
            } $state
        }
        $code=if($state.Stopped){1}else{0}
    }catch{$setup=$_;$code=2}
    foreach($name in $names){if($name -cnotin @($state.Groups|ForEach-Object Name)){$state.Groups.Add([ordered]@{Name=$name;Status='NotRun'})}}
    try{
        if($null -ne $sources){foreach($path in $sources.Keys){$after[$path]=(Get-FileHash -LiteralPath $path).Hash;RT-Need ($after[$path] -ceq $before[$path]) ('post source '+$path)}}
        if($null -ne $binding){foreach($path in $binding.HistoricalFiles.Keys){$historyAfter[$path]=(Get-FileHash -LiteralPath $path).Hash;RT-Need ($historyAfter[$path] -ceq $historyBefore[$path]) ('post history '+$path)}}
        RT-Need ((Get-FileHash $testPath).Hash -ceq $testHash -and (Get-FileHash $inputPath).Hash -ceq $inputHash) 'post immutable test/input'
    }catch{$recording=$_;$code=2}
    return [ordered]@{Schema=1;Phase=$Phase;TestSHA256=$testHash;InputSHA256=$inputHash;PhaseInputSHA256=$phaseHash;SourceSHA256=$before;PostSourceSHA256=$after;HistoricalSHA256=$historyBefore;PostHistoricalSHA256=$historyAfter;Groups=$state.Groups.ToArray();PASS=@($state.Groups|Where-Object Status -CEQ PASS).Count;FAIL=@($state.Groups|Where-Object Status -CEQ FAIL).Count;NotRun=@($state.Groups|Where-Object Status -CEQ NotRun).Count;ExitCode=$code;RootError=(RT-Error $state.RootError);SetupError=(RT-Error $setup);RecordingError=(RT-Error $recording)}
}
