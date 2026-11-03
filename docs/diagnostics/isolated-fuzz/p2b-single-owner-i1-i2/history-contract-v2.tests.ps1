# Definitions only. The approved 20 groups exercise real read-only production checks.
function HC-Need([bool]$Condition,[string]$Label){if(-not $Condition){throw ('History contract regression: '+$Label)}}
function HC-Equal($Left,$Right){return ((ConvertTo-Json -InputObject $Left -Depth 100 -Compress) -ceq (ConvertTo-Json -InputObject $Right -Depth 100 -Compress))}
function HC-Copy($Value){return ConvertFrom-Json -AsHashtable -InputObject (ConvertTo-Json -InputObject $Value -Depth 100)}
function HC-Error($Record){if($null -eq $Record){return $null};return [ordered]@{Type=$Record.Exception.GetType().FullName;Message=$Record.Exception.Message;ErrorId=$Record.FullyQualifiedErrorId;Position=$Record.InvocationInfo.PositionMessage;Stack=$Record.ScriptStackTrace}}
function HC-Tree([string]$Text){$tokens=$null;$errors=$null;$tree=[Management.Automation.Language.Parser]::ParseInput($Text,[ref]$tokens,[ref]$errors);HC-Need ($errors.Count -eq 0) 'parse';return $tree}
function HC-Definition($Tree,[string]$Name){
    $nodes=@($Tree.FindAll({param($n)$n -is [Management.Automation.Language.FunctionDefinitionAst] -and $n.Name -ceq $Name},$true))
    HC-Need ($nodes.Count -eq 1) ('unique definition '+$Name);return $nodes[0]
}
function HC-Calls($Tree,[string]$Name){return @($Tree.FindAll({param($n)$n -is [Management.Automation.Language.CommandAst] -and $n.GetCommandName() -ceq $Name},$true))}
function HC-Reject([scriptblock]$Body,[string]$Message){
    $caught=$null;$output=@();try{$output=@(& $Body)}catch{$caught=$_}
    HC-Need ($null -ne $caught -and $output.Count -eq 0) 'silent rejection required'
    HC-Need ($caught.Exception.GetType().FullName -ceq 'System.Management.Automation.RuntimeException' -and $caught.Exception.Message -ceq $Message -and $caught.FullyQualifiedErrorId -ceq $Message) ('precise rejection '+$Message)
}
function HC-Case([string]$Name,[scriptblock]$Body,$State){
    if($State.Stopped){$State.Groups.Add([ordered]@{Name=$Name;Status='NotRun'});return}
    try{$output=@(& $Body);HC-Need ($output.Count -eq 0) ('silent group '+$Name);$State.Groups.Add([ordered]@{Name=$Name;Status='PASS'})}
    catch{$State.RootError=$_;$State.Stopped=$true;$State.Groups.Add([ordered]@{Name=$Name;Status='FAIL'})}
}
function Invoke-ODHistoryContractV2Regression {
    param([Parameter(Mandatory)][ValidateSet('Green')][string]$Phase)
    Set-StrictMode -Version Latest;$ErrorActionPreference='Stop'
    $repo=(Get-Location).Path;$l='docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2'
    $p='.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner';$packet=Join-Path $repo $p
    $names=@('NewEpochBinding','RealHistory','ExternalIdentity','RepositoryFixture','ApprovalIdentity','CandidateIdentity','PriorReceiptIdentity','ParentLineage','WriterIdentity','WrongHashRejected','InvalidPathRejected','UnknownEpochRejected','DisjointNamespaces','ConsumedAttempt','IntentOwnership','ActualInputCall','ActualSlotFlow','SourceWindow','OriginalGuardsUnchanged','OriginalSuiteUnchanged')
    $state=[pscustomobject]@{Groups=[Collections.Generic.List[object]]::new();Stopped=$false;RootError=$null}
    $setup=$null;$recording=$null;$before=[ordered]@{};$after=[ordered]@{};$historyBefore=[ordered]@{};$historyAfter=[ordered]@{}
    $binding=$null;$sources=$null;$testHash=$null;$inputHash=$null;$phaseHash=$null;$code=2
    try{
        $testPath=Join-Path $repo ($l+'/history-contract-v2.tests.ps1');$inputPath=Join-Path $packet 'history-contract.input.json'
        $testHash=(Get-FileHash -LiteralPath $testPath).Hash;$inputHash=(Get-FileHash -LiteralPath $inputPath).Hash
        $binding=Get-Content -LiteralPath $inputPath -Raw|ConvertFrom-Json -AsHashtable
        $amendment=Get-Content (Join-Path $packet 'history-contract-v2.input.json') -Raw|ConvertFrom-Json -AsHashtable
        HC-Need ($amendment.TestSHA256 -ceq $testHash -and $binding.TestSHA256 -ceq (Get-FileHash (Join-Path $repo ($l+'/history-contract.tests.ps1'))).Hash) 'versioned and original test identity'
        $sources=$binding.BaselineSources
        if($Phase -ceq 'Green'){
            $candidatePath=Join-Path $packet 'history-contract-candidate.input.json';$phaseHash=(Get-FileHash $candidatePath).Hash
            $candidate=Get-Content $candidatePath -Raw|ConvertFrom-Json -AsHashtable
            HC-Need ($candidate.InputSHA256 -ceq $inputHash) 'candidate identity';$sources=$candidate.CandidateSources
        }
        $intent=Get-Content (Join-Path $packet ('history-contract-v2-'+$Phase.ToLowerInvariant()+'.intent.json')) -Raw|ConvertFrom-Json -AsHashtable
        HC-Need ($intent.AttemptConsumed -eq $true -and $intent.Computation -eq 24 -and $intent.Ceiling -eq 24 -and $intent.TestSHA256 -ceq $testHash -and $intent.InputSHA256 -ceq $inputHash -and $intent.PhaseInputSHA256 -ceq $phaseHash) 'actual finite intent'
        foreach($path in $sources.Keys){$before[$path]=(Get-FileHash -LiteralPath $path).Hash;HC-Need ($before[$path] -ceq $sources[$path]) ('source '+$path)}
        foreach($path in $binding.HistoricalFiles.Keys){$historyBefore[$path]=(Get-FileHash -LiteralPath $path).Hash;HC-Need ($historyBefore[$path] -ceq $binding.HistoricalFiles[$path]) ('history '+$path)}
        $texts=@{};$trees=@{};$oldTrees=@{}
        foreach($file in @('driver.ps1','capture.ps1','source-model.ps1','observation-contract.tests.ps1')){
            $texts[$file]=[IO.File]::ReadAllText((Join-Path $repo ($l+'/'+$file)));$trees[$file]=HC-Tree $texts[$file]
            $oldTrees[$file]=HC-Tree ([Text.Encoding]::UTF8.GetString([Convert]::FromBase64String($binding.BaselineSourceBytes[$l+'/'+$file])))
        }
        # Missing new epoch, not a missing function/setup error, is the sole expected RED.
        $d=HC-Definition $trees['driver.ps1'] Get-ODAttemptSlot
        . ([scriptblock]::Create(([string][char]10 * ($d.Extent.StartLineNumber-1))+$d.Extent.Text))
        HC-Case $names[0] {HC-Need ((Get-ODAttemptSlot contract-green history-contract) -ceq 'history-contract-host') 'new physical slot'} $state
        if(-not $state.Stopped){
            $allow=@('Get-ODEpochManifestName','Driver-Hash','Driver-CheckHash','Driver-Need','Driver-EqualJSON','Driver-Keys','Driver-VerifyHistory','Driver-ValidateHistoryInputs','Assert-ODHistoryLineage','Driver-RequireFinal','Get-ODWriterManifestPath','Driver-FalseFlags')
            foreach($name in $allow){
                $file=if($name -cin @('Get-ODEpochManifestName','Assert-ODHistoryLineage')){'source-model.ps1'}else{'driver.ps1'}
                $d=HC-Definition $trees[$file] $name
                . ([scriptblock]::Create(([string][char]10 * ($d.Extent.StartLineNumber-1))+$d.Extent.Text))
            }
            $approval=Get-Content (Join-Path $packet 'history-contract.approval.json') -Raw|ConvertFrom-Json -AsHashtable
            $fixture=$binding.FixtureBindings
            $parent=Get-Content (Join-Path $repo $fixture.Parent.Path) -Raw|ConvertFrom-Json -AsHashtable
            $inputs=$null
            HC-Case $names[1] {
                $outputs=@(Driver-VerifyHistory $repo $binding.OriginalHistory);HC-Need ($outputs.Count -eq 0 -and $binding.OriginalHistory.Count -eq 2307) 'real complete history'
                $value=Driver-ValidateHistoryInputs $repo $approval $binding $candidate
                HC-Need ((HC-Equal @($value.Keys) @('Parent','WriterArtifacts','CandidateSources','PriorBinding','PathRepair')) -and (HC-Equal $value.Parent $parent) -and (HC-Equal $value.CandidateSources $candidate.CandidateSources)) 'real returned identities'
                HC-Need ((HC-Equal $value.WriterArtifacts $fixture.WriterArtifacts) -and (HC-Equal $value.PriorBinding $fixture.PriorBinding) -and (HC-Equal $value.PathRepair $fixture.PathRepair)) 'distinct historical provenance'
            } $state
            HC-Case $names[2] {HC-Need ($fixture.ExternalHistory.Count -eq 3) 'three external identities';$o=@(Driver-VerifyHistory $repo $fixture.ExternalHistory);HC-Need ($o.Count -eq 0) 'silent external history'} $state
            HC-Case $names[3] {
                HC-Need ($fixture.RepositoryFixture.Map -ceq 'BaselineSources' -and $fixture.RepositoryFixture.Path -ceq ($l+'/expected-output.json')) 'explicit live fixture map'
                $map=[ordered]@{};$map[$fixture.RepositoryFixture.Path]=$binding.BaselineSources[$fixture.RepositoryFixture.Path]
                HC-Need ($map[$fixture.RepositoryFixture.Path] -ceq '0FA169E9C7D69EB4D53AFF6C20378ADE4DB43E6767880F2991F26A7B87C69BBF') 'independent fixture hash'
                $o=@(Driver-VerifyHistory $repo $map);HC-Need ($o.Count -eq 0) 'silent relative fixture'
            } $state
            # One-field mutations must be rejected by the real input validator.
            HC-Case $names[4] {
                foreach($field in @('Schema','BaseCommit','UserReply','AddedAuxiliary')){
                    $bad=HC-Copy $approval;$bad[$field]=if($field -cin @('Schema','AddedAuxiliary')){99}else{'wrong'}
                    HC-Reject {Driver-ValidateHistoryInputs $repo $bad $binding $candidate} 'OD: history approval mismatch'
                }
            } $state
            HC-Case $names[5] {
                $key=$l+'/driver.ps1';$bad=HC-Copy $candidate;$bad.CandidateSources[$key]='0'*64
                HC-Reject {Driver-ValidateHistoryInputs $repo $approval $binding $bad} 'OD: history candidate mismatch'
                $bad=HC-Copy $candidate;$bad.CandidateSources.Remove($key)
                HC-Reject {Driver-ValidateHistoryInputs $repo $approval $binding $bad} 'OD: history candidate mismatch'
            } $state
            HC-Case $names[6] {
                $bad=HC-Copy $binding;$key=@($bad.FixtureBindings.PriorBinding.Keys)[0];$bad.FixtureBindings.PriorBinding[$key]='0'*64
                HC-Reject {Driver-ValidateHistoryInputs $repo $approval $bad $candidate} 'OD: history prior receipt mismatch'
            } $state
            HC-Case $names[7] {
                $bad=HC-Copy $binding;$bad.FixtureBindings.Parent.SHA256=$fixture.WriterSHA256
                HC-Reject {Driver-ValidateHistoryInputs $repo $approval $bad $candidate} 'OD: history parent mismatch'
            } $state
            HC-Case $names[8] {
                $bad=HC-Copy $binding;$bad.FixtureBindings.WriterSHA256='0'*64
                HC-Reject {Driver-ValidateHistoryInputs $repo $approval $bad $candidate} 'OD: history writer mismatch'
            } $state
            HC-Case $names[9] {
                $map=[ordered]@{};$map[$fixture.RepositoryFixture.Path]='0'*64
                HC-Reject {Driver-VerifyHistory $repo $map} ('Input drift/missing: '+[IO.Path]::GetFullPath((Join-Path $repo $fixture.RepositoryFixture.Path)))
            } $state
            HC-Case $names[10] {
                foreach($path in @('C:/not-approved-history.bin','../outside-history.bin')){
                    $map=[ordered]@{};$map[$path]='0'*64;HC-Reject {Driver-VerifyHistory $repo $map} 'OD: historical path binding mismatch'
                }
            } $state
            HC-Case $names[11] {
                foreach($epoch in @('','unknown','History-Contract')){
                    HC-Reject {Get-ODAttemptSlot contract-green $epoch} 'OD: invalid explicit epoch'
                    HC-Reject {Get-ODEpochManifestName $epoch} 'OD: invalid explicit epoch'
                }
                foreach($grant in @('writer-green','core','lifecycle','continuation')){HC-Reject {Get-ODAttemptSlot $grant history-contract} 'OD: invalid explicit grant'}
                HC-Need ((Get-ODEpochManifestName history-contract) -ceq 'history-contract.frozen.json') 'literal new manifest'
            } $state
            HC-Case $names[12] {
                $slots=@('history-contract','history-contract-host','history-contract-binding-red','history-contract-binding-green')
                $paths=@(foreach($slot in $slots){foreach($suffix in @('intent.json','final.json','tool-record.json')){Join-Path $packet ($slot+'.'+$suffix)}})
                HC-Need (@($paths|Select-Object -Unique).Count -eq 12) 'disjoint attempts'
                HC-Need (-not (Test-Path (Join-Path $packet 'history-contract.frozen.json')) -and -not (Test-Path (Join-Path $packet 'history-contract-host.intent.json'))) 'pre-freeze real absence'
            } $state
            HC-Case $names[13] {
                # Evaluate only the actual read-only occupancy expression, never the gate.
                $gate=HC-Definition $trees['driver.ps1'] Driver-HistoryGate
                $calls=@(HC-Calls $gate Test-Path|Where-Object {$_.Extent.Text -ceq 'Test-Path -LiteralPath $occupiedPath'})
                HC-Need ($calls.Count -eq 1) 'actual occupancy decision'
                $expression=[scriptblock]::Create($calls[0].Extent.Text)
                foreach($rel in @('pin-binding.stop.json','extent-contract.stop.json','snapshots/repair-green')){$occupiedPath=Join-Path $packet $rel;HC-Need ([bool](& $expression)) 'existing file/directory consumed'}
                foreach($rel in @('history-contract.intent.json','history-contract-host.intent.json','snapshots/history-contract')){$occupiedPath=Join-Path $packet $rel;HC-Need (-not [bool](& $expression)) 'fresh destination absent'}
            } $state
            HC-Case $names[14] {
                # Structural, not proof that writes or completed-Prepare authentication ran.
                $acquire=HC-Definition $trees['driver.ps1'] Driver-AcquireHistoryAttempt
                HC-Need ($acquire.Extent.Text.Contains('[IO.FileMode]::CreateNew') -and $acquire.Extent.Text.Contains('[guid]::NewGuid()')) 'exclusive private acquisition'
                HC-Need ((HC-Equal @($acquire.Parameters|ForEach-Object {$_.Name.VariablePath.UserPath}) @('Action','Grant'))) 'no caller path/token'
                $gate=HC-Definition $trees['driver.ps1'] Driver-HistoryGate
                HC-Need ($gate.Extent.Text.Contains('$owned.Token -ceq $Ownership.Token') -and $gate.Extent.Text.Contains('$Ownership.SHA256')) 'actual owned-byte/token comparison'
                $acquireCalls=@(HC-Calls $trees['driver.ps1'] Driver-AcquireHistoryAttempt)
                HC-Need ($acquireCalls.Count -eq 1) 'one private call site'
                $engineChecks=@(HC-Calls $trees['driver.ps1'] Driver-CheckHash|Where-Object {$_.Extent.Text -match '^Driver-CheckHash \$engine '})
                HC-Need ($engineChecks.Count -eq 1 -and $acquireCalls[0] -is [Management.Automation.Language.CommandAst] -and $engineChecks[0] -is [Management.Automation.Language.CommandAst] -and $acquireCalls[0].Extent.StartOffset -lt $engineChecks[0].Extent.StartOffset) 'acquire before engine validation'
            } $state
            HC-Case $names[15] {
                $gate=HC-Definition $trees['driver.ps1'] Driver-HistoryGate;$calls=@(HC-Calls $gate Driver-ValidateHistoryInputs)
                HC-Need ($calls.Count -eq 1) 'one actual shared call'
                $a=$approval;$b=$binding;$c=$candidate
                $actual=& ([scriptblock]::Create($calls[0].Extent.Text))
                $expected=Driver-ValidateHistoryInputs $repo $approval $binding $candidate
                HC-Need ((HC-Equal $actual $expected) -and (HC-Equal $actual.Parent $parent)) 'actual gate call results'
            } $state
            HC-Case $names[16] {
                foreach($name in @('Driver-UnitEvidence','Driver-Recorder','Driver-RequireFinal')){
                    $d=HC-Definition $trees['driver.ps1'] $name
                    $calls=@(HC-Calls $d Get-ODAttemptSlot);HC-Need ($calls.Count -gt 0) ('physical consumer '+$name)
                    foreach($call in $calls){HC-Need ($call.Extent.Text -match '\$Epoch\b') ('explicit epoch '+$name)}
                }
                $evidence=HC-Definition $trees['driver.ps1'] Driver-Evidence
                foreach($name in @('Driver-UnitEvidence','Driver-Recorder')){
                    $calls=@(HC-Calls $evidence $name);HC-Need ($calls.Count -gt 0) ('evidence consumer '+$name)
                    foreach($call in $calls){HC-Need ($call.Extent.Text -match '\$Epoch\b') ('evidence explicit epoch '+$name)}
                }
                HC-Need ($texts['capture.ps1'].Contains("'history-contract-host'")) 'literal capture slot'
                foreach($item in @(@('capture.ps1','Invoke-ODContractTests'),@('observation-contract.tests.ps1','New-ODIntentUnitContext'),@('source-model.ps1','New-ODFixedContext'))){
                    HC-Need (@(HC-Calls $trees[$item[0]] $item[1]|Where-Object {$_.Extent.Text -match '\$Epoch\b'}).Count -ge 1) ('epoch flow '+$item[1])
                }
                $unit=HC-Definition $trees['source-model.ps1'] New-ODIntentUnitContext
                HC-Need (@(HC-Calls $unit Assert-ODHistoryLineage).Count -eq 1) 'lineage actually consumed'
            } $state
            HC-Case $names[17] {
                $changed=@($sources.Keys|Where-Object {$sources[$_] -cne $binding.BaselineSources[$_]})
                HC-Need ($sources.Count -eq 31 -and $changed.Count -eq 4 -and @(Compare-Object $changed @($binding.BaselineSourceBytes.Keys)).Count -eq 0) 'four-source window; other27 unchanged'
            } $state
            HC-Case $names[18] {
                foreach($pair in @(@('driver.ps1','Driver-VerifyHistory'),@('source-model.ps1','Get-ODFailureExtent'),@('driver.ps1','Driver-Recorder'),@('driver.ps1','Driver-FalseFlags'))){
                    HC-Need ((HC-Definition $trees[$pair[0]] $pair[1]).Extent.Text -ceq (HC-Definition $oldTrees[$pair[0]] $pair[1]).Extent.Text) ('unchanged guard '+$pair[1])
                }
                $old=[Text.Encoding]::UTF8.GetString([Convert]::FromBase64String($binding.BaselineSourceBytes[$l+'/driver.ps1']))
                $start='$process=[Diagnostics.Process]::new()';$nl=[string][char]10
                $oldTail=$old.Substring($old.IndexOf($start));$oldTail=$oldTail.Substring(0,$oldTail.LastIndexOf($nl+'}catch{'))
                $newTail=$texts['driver.ps1'].Substring($texts['driver.ps1'].IndexOf($start));$newTail=$newTail.Substring(0,$newTail.LastIndexOf($nl+'}catch{'))
                HC-Need ($oldTail -ceq $newTail) 'owned process unchanged'
            } $state
            HC-Case $names[19] {
                $old=[Text.Encoding]::UTF8.GetString([Convert]::FromBase64String($binding.BaselineSourceBytes[$l+'/observation-contract.tests.ps1']))
                $bindingLine="    if("+ '$Epoch' +" -cnotin @('legacy','extent-contract','history-contract')){throw 'OD: invalid explicit epoch'}"+[string][char]10
                HC-Need ($texts['observation-contract.tests.ps1'].Replace($bindingLine,'') -ceq $old) 'all49suite cases unchanged outside entry binding'
                HC-Need ((HC-Definition $trees['driver.ps1'] Driver-Recorder).Extent.Text -ceq (HC-Definition $oldTrees['driver.ps1'] Driver-Recorder).Extent.Text) 'original3recorder groups unchanged'
            } $state
        }
        $code=if($state.Stopped){1}else{0}
    }catch{$setup=$_;$code=2}
    foreach($name in $names){if($name -cnotin @($state.Groups|ForEach-Object Name)){$state.Groups.Add([ordered]@{Name=$name;Status='NotRun'})}}
    try{
        if($null -ne $sources){foreach($path in $sources.Keys){$after[$path]=(Get-FileHash -LiteralPath $path).Hash;HC-Need ($after[$path] -ceq $before[$path]) ('post source '+$path)}}
        if($null -ne $binding){foreach($path in $binding.HistoricalFiles.Keys){$historyAfter[$path]=(Get-FileHash -LiteralPath $path).Hash;HC-Need ($historyAfter[$path] -ceq $historyBefore[$path]) ('post history '+$path)}}
        HC-Need ((Get-FileHash $testPath).Hash -ceq $testHash -and (Get-FileHash $inputPath).Hash -ceq $inputHash) 'post immutable test/input'
    }catch{$recording=$_;$code=2}
    return [ordered]@{Schema=1;Phase=$Phase;TestSHA256=$testHash;InputSHA256=$inputHash;PhaseInputSHA256=$phaseHash;SourceSHA256=$before;PostSourceSHA256=$after;HistoricalSHA256=$historyBefore;PostHistoricalSHA256=$historyAfter;Groups=$state.Groups.ToArray();PASS=@($state.Groups|Where-Object Status -CEQ PASS).Count;FAIL=@($state.Groups|Where-Object Status -CEQ FAIL).Count;NotRun=@($state.Groups|Where-Object Status -CEQ NotRun).Count;ExitCode=$code;RootError=(HC-Error $state.RootError);SetupError=(HC-Error $setup);RecordingError=(HC-Error $recording)}
}
