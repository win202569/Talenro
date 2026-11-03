# Definitions only. Extract actual production statements; never run driver top level.
function HP-Need([bool]$Good,[string]$Message){if(-not $Good){throw ('History path regression: '+$Message)}}
function HP-Error($Record){if($null -eq $Record){return $null};return [ordered]@{Type=$Record.Exception.GetType().FullName;Message=$Record.Exception.Message;ErrorId=$Record.FullyQualifiedErrorId;Position=$Record.InvocationInfo.PositionMessage;Stack=$Record.ScriptStackTrace;TargetObject=[string]$Record.TargetObject}}
function HP-Tree([string]$Path){$tokens=$null;$issues=$null;$tree=[Management.Automation.Language.Parser]::ParseFile($Path,[ref]$tokens,[ref]$issues);HP-Need ($issues.Count -eq 0) 'parse';return $tree}
function HP-Definition($Tree,[string]$Name){$defs=@($Tree.FindAll({param($n)$n -is [Management.Automation.Language.FunctionDefinitionAst] -and $n.Name -ceq $Name},$true));HP-Need ($defs.Count -eq 1) ('definition '+$Name);return $defs[0]}
function HP-Case([string]$Name,[scriptblock]$Body,$State){
    if($State.Stopped){$State.Groups.Add([ordered]@{Name=$Name;Status='NotRun'});return}
    try{$emitted=@(& $Body);HP-Need ($emitted.Count -eq 0) ('unexpected output '+$Name);$State.Groups.Add([ordered]@{Name=$Name;Status='PASS'})}
    catch{$State.RootError=$_;$State.Stopped=$true;$State.Groups.Add([ordered]@{Name=$Name;Status='FAIL'})}
}
function HP-Reject([scriptblock]$Body,[string]$Kind='Binding'){
    $caught=$null;$outcome=@();try{$outcome=@(& $Body)}catch{$caught=$_}
    HP-Need ($null -ne $caught -and $outcome.Count -eq 0) 'must reject'
    if($Kind -ceq 'Binding'){HP-Need ($caught.Exception.Message -ceq 'OD: historical path binding mismatch' -and $caught.FullyQualifiedErrorId -ceq 'OD: historical path binding mismatch') 'binding error'}
    elseif($Kind -ceq 'Hash'){HP-Need ($caught.Exception.Message.StartsWith('Input drift/missing: ')) 'real hash rejection'}
    else{HP-Need ($caught.FullyQualifiedErrorId -match 'PathNotFound|FileNotFound|FileReadError') 'real missing file rejection'}
}
function Invoke-ODHistoryPathRegression {
    param([Parameter(Mandatory)][ValidateSet('Red','Green')][string]$Phase)
    Set-StrictMode -Version Latest;$ErrorActionPreference='Stop'
    $repo=(Get-Location).Path;$p=Join-Path $repo '.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner'
    $l='docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2'
    $names=@('SavedExternalReadable','AllOriginalHistory','OtherExternalIdentities','RepositoryRelative','WrongHashRejected','MissingFileRejected','ExternalHashRejected','UnknownAbsoluteRejected','DriveRelativeRejected','RootRelativeRejected','NetworkDeviceRejected','TraversalRejected','ConflictingAliasesRejected','IdenticalAliasesAccepted','MalformedInputsRejected','SourceWindow','ActualGateWiring','OriginalGuardsUnchanged')
    $state=[pscustomobject]@{Groups=[Collections.Generic.List[object]]::new();RootError=$null;Stopped=$false}
    $before=[ordered]@{};$after=[ordered]@{};$hb=[ordered]@{};$ha=[ordered]@{}
    $setup=$null;$recording=$null;$binding=$null;$sources=$null;$testHash=$null;$inputHash=$null;$phaseHash=$null;$code=2
    try{
        $testPath=Join-Path $repo ($l+'/history-path-v3.tests.ps1');$inputPath=Join-Path $p history-path-v3.input.json
        $testHash=(Get-FileHash $testPath).Hash;$inputHash=(Get-FileHash $inputPath).Hash
        $binding=Get-Content $inputPath -Raw|ConvertFrom-Json -AsHashtable
        HP-Need ($testHash -ceq $binding.TestSHA256) 'test hash'
        HP-Need ($binding.BaselineSources.Contains($l+'/expected-output.json') -and $binding.BaselineSources[$l+'/expected-output.json'] -cmatch '^[A-F0-9]{64}$') 'relative fixture present in baseline sources'
        $sources=$binding.BaselineSources
        if($Phase -ceq 'Green'){
            $phasePath=Join-Path $p history-path-v3-candidate.input.json;$phaseHash=(Get-FileHash $phasePath).Hash
            $candidate=Get-Content $phasePath -Raw|ConvertFrom-Json -AsHashtable
            HP-Need ($candidate.InputSHA256 -ceq $inputHash -and $candidate.RedFinalSHA256 -ceq (Get-FileHash (Join-Path $p history-path-red.final.json)).Hash) 'candidate lineage'
            $sources=$candidate.CandidateSources
        }
        $intent=Get-Content (Join-Path $p ('history-path-v3-'+$Phase.ToLowerInvariant()+'.intent.json')) -Raw|ConvertFrom-Json -AsHashtable
        HP-Need ($intent.AttemptConsumed -eq $true -and $intent.Computation -eq $(if($Phase -ceq 'Red'){20}else{21}) -and $intent.Ceiling -eq 21 -and $intent.InputSHA256 -ceq $inputHash -and $intent.TestSHA256 -ceq $testHash -and $intent.PhaseInputSHA256 -ceq $phaseHash) 'exact consumed intent'
        foreach($path in $sources.Keys){$before[$path]=(Get-FileHash $path).Hash;HP-Need ($before[$path] -ceq $sources[$path]) ('source '+$path)}
        foreach($path in $binding.HistoricalFiles.Keys){$hb[$path]=(Get-FileHash $path).Hash;HP-Need ($hb[$path] -ceq $binding.HistoricalFiles[$path]) ('history '+$path)}
        $tree=HP-Tree (Join-Path $repo ($l+'/driver.ps1'));$gate=HP-Definition $tree Driver-ExtentContractGate
        foreach($name in @('Driver-Hash','Driver-CheckHash')){
            $d=HP-Definition $tree $name
            HP-Need ($d.Extent.Text -ceq $binding.Helpers[$name].Text) 'unchanged hash helper'
            . ([scriptblock]::Create(([string][char]10*($binding.Helpers[$name].Line-1))+$d.Extent.Text))
        }
        $external='C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe'
        $single=[ordered]@{};$single[$external]=$binding.OriginalHistory[$external]
        if($Phase -ceq 'Red'){
            $loops=@($gate.FindAll({param($n)$n -is [Management.Automation.Language.ForEachStatementAst] -and $n.Condition.Extent.Text -ceq '$b.HistoricalFiles.Keys'},$true))
            HP-Need ($loops.Count -eq 1 -and $loops[0].Extent.Text -ceq $binding.Loop.Text) 'actual original history loop'
            $statement=[scriptblock]::Create(([string][char]10*($binding.Loop.Line-1))+$loops[0].Extent.Text)
            $b=[ordered]@{HistoricalFiles=$single}
            HP-Case $names[0] {& $statement} $state
        }else{
            $definition=HP-Definition $tree Driver-VerifyHistory
            . ([scriptblock]::Create($definition.Extent.Text))
            HP-Case $names[0] {Driver-VerifyHistory $repo $single} $state
            HP-Case $names[1] {Driver-VerifyHistory $repo $binding.OriginalHistory} $state
            HP-Case $names[2] {$two=[ordered]@{};foreach($k in $binding.OriginalHistory.Keys){if($k -match '^[A-Za-z]:' -and $k -cne $external){$two[$k]=$binding.OriginalHistory[$k]}};HP-Need ($two.Count -eq 2) 'other two external identities';Driver-VerifyHistory $repo $two} $state
            $relative=$l+'/expected-output.json';$good=[ordered]@{};$good[$relative]=$binding.BaselineSources[$relative]
            HP-Case $names[3] {Driver-VerifyHistory $repo $good} $state
            HP-Case $names[4] {$bad=[ordered]@{};$bad[$relative]='0'*64;HP-Reject {Driver-VerifyHistory $repo $bad} Hash} $state
            HP-Case $names[5] {$bad=[ordered]@{'.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/history-path-absent-fixture.bin'=('0'*64)};HP-Reject {Driver-VerifyHistory $repo $bad} Missing} $state
            HP-Case $names[6] {$bad=[ordered]@{};$bad[$external]='0'*64;HP-Reject {Driver-VerifyHistory $repo $bad}} $state
            HP-Case $names[7] {HP-Reject {Driver-VerifyHistory $repo ([ordered]@{'C:/history-path-unapproved.exe'=('0'*64)})}} $state
            HP-Case $names[8] {HP-Reject {Driver-VerifyHistory $repo ([ordered]@{'C:relative.txt'=('0'*64)})}} $state
            HP-Case $names[9] {foreach($k in @('\relative.txt','/relative.txt')){$bad=[ordered]@{};$bad[$k]='0'*64;HP-Reject {Driver-VerifyHistory $repo $bad}}} $state
            HP-Case $names[10] {foreach($k in @('\\server\share\file','\\?\C:\file','\\.\device')){$bad=[ordered]@{};$bad[$k]='0'*64;HP-Reject {Driver-VerifyHistory $repo $bad}}} $state
            HP-Case $names[11] {HP-Reject {Driver-VerifyHistory $repo ([ordered]@{'../outside-history.txt'=('0'*64)})}} $state
            HP-Case $names[12] {$bad=[ordered]@{};$bad[$relative]=$good[$relative];$bad['./'+$relative]='0'*64;HP-Reject {Driver-VerifyHistory $repo $bad}} $state
            HP-Case $names[13] {$aliases=[ordered]@{};$aliases[$relative]=$good[$relative];$aliases['./'+$relative]=$good[$relative];Driver-VerifyHistory $repo $aliases} $state
            HP-Case $names[14] {
                HP-Reject {Driver-VerifyHistory $repo $null};HP-Reject {Driver-VerifyHistory $repo ([ordered]@{})}
                foreach($k in @('','bad./file','bad /file','*.txt','x:y',[string][char]0)){$bad=[ordered]@{};$bad[$k]='0'*64;HP-Reject {Driver-VerifyHistory $repo $bad}}
                foreach($hash in @('0',('a'*64),('Z'*64))){$bad=[ordered]@{};$bad[$relative]=$hash;HP-Reject {Driver-VerifyHistory $repo $bad}}
            } $state
            HP-Case $names[15] {$changed=@($sources.Keys|Where-Object {$sources[$_] -cne $binding.BaselineSources[$_]});HP-Need ($sources.Count -eq 31 -and $changed.Count -eq 1 -and $changed[0] -ceq ($l+'/driver.ps1')) 'driver-only edit'} $state
            HP-Case $names[16] {
                $calls=@($gate.FindAll({param($n)$n -is [Management.Automation.Language.CommandAst] -and $n.GetCommandName() -ceq 'Driver-VerifyHistory'},$true))
                HP-Need ($calls.Count -eq 1 -and $calls[0].Extent.Text -ceq 'Driver-VerifyHistory $repo $b.HistoricalFiles') 'actual call binding'
                $b=[ordered]@{HistoricalFiles=$binding.OriginalHistory}
                & ([scriptblock]::Create($calls[0].Extent.Text))
            } $state
            HP-Case $names[17] {
                $text=[IO.File]::ReadAllText((Join-Path $repo ($l+'/driver.ps1')))
                $restored=$text.Remove($definition.Extent.StartOffset,$definition.Extent.EndOffset-$definition.Extent.StartOffset+1)
                $restored=$restored.Replace('Driver-VerifyHistory $repo $b.HistoricalFiles',$binding.Loop.Text)
                $hash=[Convert]::ToHexString([Security.Cryptography.SHA256]::HashData([Text.UTF8Encoding]::new($false).GetBytes($restored)))
                HP-Need ($hash -ceq $binding.BaselineSources[$l+'/driver.ps1']) 'reverse restores exact original driver'
            } $state
        }
        $code=if($state.Stopped){1}else{0}
    }catch{$setup=$_;$code=2}
    foreach($name in $names){if($name -cnotin @($state.Groups|ForEach-Object Name)){$state.Groups.Add([ordered]@{Name=$name;Status='NotRun'})}}
    try{
        if($null -ne $sources){foreach($path in $sources.Keys){$after[$path]=(Get-FileHash $path).Hash;HP-Need ($after[$path] -ceq $before[$path]) ('post source '+$path)}}
        if($null -ne $binding){foreach($path in $binding.HistoricalFiles.Keys){$ha[$path]=(Get-FileHash $path).Hash;HP-Need ($ha[$path] -ceq $hb[$path]) ('post history '+$path)}}
        HP-Need ((Get-FileHash $testPath).Hash -ceq $testHash -and (Get-FileHash $inputPath).Hash -ceq $inputHash) 'post test/input'
        if($Phase -ceq 'Green'){HP-Need ((Get-FileHash $phasePath).Hash -ceq $phaseHash) 'post candidate input'}
    }catch{$recording=$_;$code=2}
    return [ordered]@{Schema=1;Phase=$Phase;TestSHA256=$testHash;InputSHA256=$inputHash;PhaseInputSHA256=$phaseHash;SourceSHA256=$before;PostSourceSHA256=$after;HistoricalSHA256=$hb;PostHistoricalSHA256=$ha;Groups=$state.Groups.ToArray();PASS=@($state.Groups|Where-Object Status -CEQ PASS).Count;FAIL=@($state.Groups|Where-Object Status -CEQ FAIL).Count;NotRun=@($state.Groups|Where-Object Status -CEQ NotRun).Count;ExitCode=$code;RootError=(HP-Error $state.RootError);SetupError=(HP-Error $setup);RecordingError=(HP-Error $recording)}
}
