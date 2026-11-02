# Finite actual-source auxiliary; loading defines functions only.
function CM-Need([bool]$Good,[string]$Label){if(-not $Good){throw ('Call model regression: '+$Label)}}
function CM-Tree([string]$Path){
    $t=$null;$e=$null;$a=[Management.Automation.Language.Parser]::ParseFile($Path,[ref]$t,[ref]$e)
    CM-Need ($e.Count -eq 0) ('parse '+$Path);return $a
}
function CM-Definition($Tree,[string]$Name){
    $found=@($Tree.FindAll({param($n)$n -is [Management.Automation.Language.FunctionDefinitionAst] -and $n.Name -ceq $Name},$true))
    CM-Need ($found.Count -eq 1) ('unique function '+$Name);return $found[0]
}
function CM-Same($A,$B){return (ConvertTo-Json -InputObject $A -Depth 70 -Compress) -ceq (ConvertTo-Json -InputObject $B -Depth 70 -Compress)}
function CM-Copy($Value){return ConvertFrom-Json -AsHashtable -InputObject (ConvertTo-Json -InputObject $Value -Depth 70)}
function CM-Error($Record){
    if($null -eq $Record){return $null}
    $position=$null;if($null -ne $Record.InvocationInfo){$position=$Record.InvocationInfo.PositionMessage}
    return [ordered]@{Type=$Record.Exception.GetType().FullName;Message=$Record.Exception.Message;ErrorId=$Record.FullyQualifiedErrorId;Position=$position;Stack=$Record.ScriptStackTrace}
}
function CM-Case([string]$Name,[scriptblock]$Body,$State){
    if($State.Stopped){$State.Groups.Add([ordered]@{Name=$Name;Status='NotRun'});return}
    try{$out=@(& $Body);CM-Need ($out.Count -eq 0) ('unexpected output '+$Name);$State.Groups.Add([ordered]@{Name=$Name;Status='PASS'})}
    catch{if($null -eq $State.RootError){$State.RootError=$_};$State.Stopped=$true;$State.Groups.Add([ordered]@{Name=$Name;Status='FAIL'})}
}
function CM-Reject([scriptblock]$Body,[string]$Message='OD: observation contract mismatch'){
    $caught=$null;$out=@();try{$out=@(& $Body)}catch{$caught=$_}
    CM-Need ($null -ne $caught -and $out.Count -eq 0) 'negative must reject without output'
    if($caught.Exception.GetType().FullName -cne 'System.Management.Automation.RuntimeException' -or $caught.Exception.Message -cne $Message -or $caught.FullyQualifiedErrorId -cne $Message){throw $caught}
}
function Invoke-ODCallModelRegressionV2 {
    [CmdletBinding(PositionalBinding=$false)]
    param([Parameter(Mandatory)][ValidateSet('Red','Green')][string]$Phase)
    Set-StrictMode -Version Latest;$ErrorActionPreference='Stop'
    $state=[pscustomobject]@{Groups=[Collections.Generic.List[object]]::new();RootError=$null;RecordingError=$null;Stopped=$false}
    $names=@('ExactDescriptor','OtherControlsExcluded','NamedCallsUnchanged','UnknownVariableExcluded','ActualUnitStackAccepted','WrongSourceRejected','WrongFunctionRejected','WrongCallerLineRejected','BodyExtentRejected','MissingBindingRejected','BindingIdentityRejected','AmbiguousBindingRejected','RuntimeBindingAbsent','RuntimeBridgeRejected','FaultOperationPreserved','HistoricalEpoch','RecoveryGate','SourceWindow','FixedSlots','OriginalGuardsUnchanged')
    $testHash=$null;$inputHash=$null;$before=[ordered]@{};$after=[ordered]@{};$exitCode=2
    try{
        $leaf=Split-Path (Get-Command Invoke-ODCallModelRegressionV2 -CommandType Function).ScriptBlock.File
        $repo=[IO.Path]::GetFullPath((Join-Path $leaf '../../../..'))
        $packet=Join-Path $repo '.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner'
        $prefix='contract-call-model-v2-'+$Phase.ToLowerInvariant()
        $testPath=Join-Path $leaf contract-call-model-v2.tests.ps1;$inputPath=Join-Path $packet contract-call-model-v2.input.json
        $testHash=(Get-FileHash $testPath).Hash;$inputHash=(Get-FileHash $inputPath).Hash
        $binding=Get-Content -Raw $inputPath|ConvertFrom-Json -AsHashtable
        CM-Need ($binding.Schema -eq 1 -and $binding.TestSHA256 -ceq $testHash) 'frozen test identity'
        foreach($key in $binding.HistoricalFiles.Keys){CM-Need ((Get-FileHash -LiteralPath (Join-Path $repo $key)).Hash -ceq $binding.HistoricalFiles[$key]) ('historical '+$key)}
        $expected=$binding.BaselineSources;$candidateHash=$null
        if($Phase -ceq 'Green'){
            $candidatePath=Join-Path $packet contract-call-model-v2-candidate.input.json
            $candidateHash=(Get-FileHash $candidatePath).Hash
            $candidate=Get-Content -Raw $candidatePath|ConvertFrom-Json -AsHashtable
            CM-Need ($candidate.Schema -eq 1 -and $candidate.InputSHA256 -ceq $inputHash -and $candidate.ExpectedPinCount -eq 2276) 'candidate identity'
            $expected=$candidate.CandidateSources
        }
        foreach($key in $expected.Keys){$before[$key]=(Get-FileHash -LiteralPath (Join-Path $repo $key)).Hash;CM-Need ($before[$key] -ceq $expected[$key]) ('source '+$key)}
        $manifest=Get-Content -Raw (Join-Path $packet writer-projection-repair.frozen.json)|ConvertFrom-Json -AsHashtable
        # Read-only modeling of the unmodified record source against original pins.
        $recordRelative='docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/record-boundary.tests.ps1'
        $record=Join-Path $repo $recordRelative
        $modelTree=CM-Tree (Join-Path $leaf source-model.ps1)
        foreach($name in @('Get-ODPinnedSHA256','Get-ODSourceOwner','Get-ODFailureExtent','Read-ODSourceModel')){. ([scriptblock]::Create((CM-Definition $modelTree $name).Extent.Text))}
        $models=Read-ODSourceModel $repo @($recordRelative) $manifest
        CM-Case $names[0] {
            $descriptors=@();if($models[$record].Contains('DynamicCalls')){$descriptors=@($models[$record].DynamicCalls)}
            if($descriptors.Count -ne 1){throw 'OD: fixed owner-control descriptor missing'}
            # Independently locate the literal case block, not via production helper.
            $owner=CM-Definition (CM-Tree $record) Invoke-ODOwnerControls
            $cases=@($owner.FindAll({param($n)
                if($n -isnot [Management.Automation.Language.CommandAst] -or $n.CommandElements.Count -lt 2){return $false}
                $v=$n.CommandElements[0];$b=$n.CommandElements[1]
                return $v -is [Management.Automation.Language.VariableExpressionAst] -and $v.VariablePath.UserPath -ceq 'controls' -and $b -is [Management.Automation.Language.ScriptBlockExpressionAst] -and $null -ne $b.ScriptBlock.ParamBlock -and @($b.ScriptBlock.ParamBlock.Parameters).Count -eq 1 -and $b.ScriptBlock.ParamBlock.Parameters[0].Name.VariablePath.UserPath -ceq 'Name'
            },$true))
            CM-Need ($cases.Count -eq 1) 'independent unique case'
            $call=$cases[0];$body=$call.CommandElements[1].ScriptBlock.Extent
            $want=[ordered]@{Source=$record;SHA256=(Get-FileHash $record).Hash;Owner='Invoke-ODOwnerControls';Variable='controls';Line=$call.Extent.StartLineNumber;Column=$call.Extent.StartColumnNumber;BodyStartLine=$body.StartLineNumber;BodyStartColumn=$body.StartColumnNumber;BodyEndLine=$body.EndLineNumber;BodyEndColumn=$body.EndColumnNumber}
            CM-Need (CM-Same $descriptors[0] $want) 'AST descriptor exact'
        } $state
        if($Phase -ceq 'Red'){
            CM-Need ($null -ne $state.RootError -and $state.RootError.Exception.GetType().FullName -ceq 'System.Management.Automation.RuntimeException' -and $state.RootError.Exception.Message -ceq 'OD: fixed owner-control descriptor missing' -and $state.RootError.FullyQualifiedErrorId -ceq 'OD: fixed owner-control descriptor missing') 'exact RED cause'
            for($i=1;$i -lt $names.Count;$i++){CM-Case $names[$i] {} $state}
            $exitCode=1
        }elseif(-not $state.Stopped){
            $d=$models[$record].DynamicCalls[0]
            $oldModelTree=CM-Tree (Join-Path (Join-Path $packet snapshots/writer-projection-rebind) 'docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/source-model.ps1')
            $oldCalls=& { . ([scriptblock]::Create((CM-Definition $oldModelTree Read-ODSourceModel).Extent.Text));(Read-ODSourceModel $repo @($recordRelative) $manifest)[$record].Calls }
            . ([scriptblock]::Create((CM-Definition $modelTree Get-ODIntentCallBinding).Extent.Text))
            CM-Case $names[1] {
                $commands=@((CM-Tree $record).FindAll({param($n)$n -is [Management.Automation.Language.CommandAst] -and $n.CommandElements.Count -gt 0 -and $n.CommandElements[0] -is [Management.Automation.Language.VariableExpressionAst] -and $n.CommandElements[0].VariablePath.UserPath -ceq 'controls'},$true))
                CM-Need ($commands.Count -ge 3 -and @($commands|Where-Object {$_.Extent.StartLineNumber -eq $d.Line}).Count -eq 1) 'initializer and cleanup excluded'
                $empty=CM-Copy $models;$empty[$record].DynamicCalls=@();CM-Reject {Get-ODIntentCallBinding $empty $record} 'OD: unit call binding mismatch'
            } $state
            CM-Case $names[2] {CM-Need (CM-Same @($models[$record].Calls) @($oldCalls)) 'named calls unchanged'} $state
            CM-Case $names[3] {
                $unknown=@((CM-Tree $record).FindAll({param($n)$n -is [Management.Automation.Language.CommandAst] -and $n.CommandElements.Count -gt 0 -and $n.CommandElements[0] -is [Management.Automation.Language.VariableExpressionAst] -and $n.CommandElements[0].VariablePath.UserPath -cnotin @('controls','entry','script:OriginalFactory')},$true))
                CM-Need ($unknown.Count -gt 0) 'actual unknown variables present'
                foreach($u in $unknown){CM-Need (@($models[$record].DynamicCalls|Where-Object {$_.Line -eq $u.Extent.StartLineNumber -and $_.Column -eq $u.Extent.StartColumnNumber}).Count -eq 0) 'unknown variable unauthorized'}
            } $state
            # Explicit test-only context, independent of absent repair manifest.
            $context=$null;$actual=$null;$expect=$null
            CM-Case $names[4] {
                $paths=@($recordRelative,'docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/capture.ps1','scripts/offline-diagnostics-v4/private/write-boundary.ps1','scripts/offline-diagnostics-v4/private/profiles.ps1')
                $sm=Read-ODSourceModel $repo $paths $manifest
                $profile=Join-Path $repo $paths[3];$fn=CM-Definition (CM-Tree $profile) Invoke-ODProfile;$offset=$fn.Extent.StartLineNumber-1
                $relativeCalls=@(foreach($c in $sm[$profile].Calls){if($c.Caller -ceq 'Invoke-ODProfile'){[ordered]@{Caller=$c.Caller;Callee=$c.Callee;Line=($c.Line-$offset);Column=$c.Column}}})
                $sm['<No file>']=[ordered]@{SHA256=$sm[$profile].SHA256;Calls=$relativeCalls;Faults=@();DynamicCalls=@()}
                $writer=Join-Path $repo $paths[2]
                $context=[ordered]@{Sources=[ordered]@{Capture=(Join-Path $repo $paths[1])};SourceModel=$sm;ProfileRoot=(Join-Path $packet 'unit/writer-projection-green/owner-controls/intent-profile/entry-migration-core');UnitBindings=[ordered]@{ProfileFunction='Invoke-ODProfile';AnonymousSource='<No file>';FactorySource=$writer;FactoryFunction='New-ODRecordStream';OwnerControlCall=(Get-ODIntentCallBinding $sm $record)}}
                $expect=[ordered]@{Type='System.Management.Automation.MethodInvocationException';ErrorId='IOException';Message=('Exception calling "Open" with "4" argument(s): "The file ''{0}'' already exists."' -f (Join-Path $packet 'unit/writer-projection-green/owner-controls/intent-existing.bin'));Source=$writer;Function='New-ODRecordStream';Operation='Open'}
                $e=Get-Content -Raw (Join-Path $packet writer-projection-green.evidence.json)|ConvertFrom-Json -AsHashtable
                $owners=@($e.Observations|Where-Object Name -CEQ OwnerControls);CM-Need ($owners.Count -eq 1) 'saved owner row'
                $cases=@($owners[0].Value|Where-Object Name -CEQ IntentFailureNoWork);CM-Need ($cases.Count -eq 1) 'saved actual intent case'
                $actual=$cases[0].IntentReport.PrimaryError
                $contractTree=CM-Tree (Join-Path $leaf observation-contract.ps1)
                foreach($name in @('OC-Need','OC-Keys','OC-Integer','OC-Expand','OC-SourceError')){. ([scriptblock]::Create((CM-Definition $contractTree $name).Extent.Text))}
                OC-SourceError $actual $expect $context
                # Persist setup only after actual acceptance; later cases receive data, not modules.
                $state | Add-Member -NotePropertyName Data -NotePropertyValue ([pscustomobject]@{Context=$context;Actual=$actual;Expected=$expect;ContractTree=$contractTree})
            } $state
            if(-not $state.Stopped){
                $context=$state.Data.Context;$actual=$state.Data.Actual;$expect=$state.Data.Expected
                foreach($name in @('OC-Need','OC-Keys','OC-Integer','OC-Expand','OC-SourceError')){. ([scriptblock]::Create((CM-Definition $state.Data.ContractTree $name).Extent.Text))}
            }
            CM-Case $names[5] {$v=CM-Copy $actual;$v.Stack=$v.Stack.Replace(('at Invoke-ODOwnerControls, '+$record+': line '+$d.Line),('at Invoke-ODOwnerControls, '+$context.Sources.Capture+': line '+$d.Line));CM-Reject {OC-SourceError $v $expect $context}} $state
            CM-Case $names[6] {$v=CM-Copy $actual;$v.Stack=$v.Stack.Replace('at Invoke-ODOwnerControls,','at Invoke-ODWriterTests,');CM-Reject {OC-SourceError $v $expect $context}} $state
            CM-Case $names[7] {$v=CM-Copy $actual;$v.Stack=$v.Stack.Replace(('at Invoke-ODOwnerControls, '+$record+': line '+$d.Line),('at Invoke-ODOwnerControls, '+$record+': line '+($d.Line+1)));CM-Reject {OC-SourceError $v $expect $context}} $state
            CM-Case $names[8] {
                $v=CM-Copy $context;$v.SourceModel[$record].DynamicCalls[0].BodyEndLine=$d.BodyStartLine;$v.UnitBindings.OwnerControlCall=CM-Copy $v.SourceModel[$record].DynamicCalls[0]
                CM-Reject {OC-SourceError $actual $expect $v}
            } $state
            CM-Case $names[9] {$v=CM-Copy $context;$v.Remove('UnitBindings');CM-Reject {OC-SourceError $actual $expect $v}} $state
            CM-Case $names[10] {foreach($k in @('Source','SHA256','Owner','Variable','Line','Column','BodyStartLine','BodyStartColumn','BodyEndLine','BodyEndColumn')){$v=CM-Copy $context;if($v.UnitBindings.OwnerControlCall[$k] -is [string]){$v.UnitBindings.OwnerControlCall[$k]+='foreign'}else{$v.UnitBindings.OwnerControlCall[$k]++};CM-Reject {OC-SourceError $actual $expect $v}}} $state
            CM-Case $names[11] {$v=CM-Copy $context;$v.SourceModel[$record].DynamicCalls=@($v.SourceModel[$record].DynamicCalls[0],(CM-Copy $v.SourceModel[$record].DynamicCalls[0]));CM-Reject {OC-SourceError $actual $expect $v};CM-Reject {Get-ODIntentCallBinding $v.SourceModel $record} 'OD: unit call binding mismatch'} $state
            CM-Case $names[12] {foreach($n in @('New-ODRuntimeContext','New-ODFixedContext')){CM-Need ((CM-Definition $modelTree $n).Extent.Text -ceq (CM-Definition $oldModelTree $n).Extent.Text) ('runtime constructor unchanged '+$n)};CM-Need ((CM-Definition $modelTree New-ODIntentUnitContext).Extent.Text.Contains('Get-ODIntentCallBinding')) 'unit constructor binds helper'} $state
            CM-Case $names[13] {$v=CM-Copy $context;$v.UnitBindings.Remove('OwnerControlCall');CM-Reject {OC-SourceError $actual $expect $v}} $state
            CM-Case $names[14] {$v=CM-Copy $expect;$v.Operation='Read';CM-Reject {OC-SourceError $actual $v $context};$v=CM-Copy $actual;$v.Position=$v.Position.Replace('char:5','char:6');CM-Need ($v.Position -cne $actual.Position) 'position mutation real';CM-Reject {OC-SourceError $v $expect $context}} $state
            CM-Case $names[15] {
                foreach($p in $manifest.Pins){$rel=[IO.Path]::GetRelativePath($repo,$p.VerificationPath).Replace('\','/');$path=if($rel -cin $manifest.SourcePaths){Join-Path (Join-Path $packet snapshots/writer-projection-rebind) $rel}else{$p.VerificationPath};if($p.ExpectedState -ceq 'Present'){CM-Need ((Get-FileHash -LiteralPath $path).Hash -ceq $p.SHA256) ('historical epoch '+$rel)}else{CM-Need (-not (Test-Path -LiteralPath $path)) 'historical absence'}}
            } $state
            $driverTree=CM-Tree (Join-Path $leaf driver.ps1)
            CM-Case $names[16] {
                foreach($n in @('Driver-Need','Assert-ODPinBindingFacts')){. ([scriptblock]::Create((CM-Definition $driverTree $n).Extent.Text))}
                $facts=[ordered]@{Schema=1;ParentSHA256='182B85886304BFE3CBFB248DDA70F2EB99DF3E2158F574DBE299519DE7CE0461';HistoricalStopSHA256='8581A9DA9A4CA3E7F2367757BE7616992889B434F42B1C4F35922B37921633ED';ApprovalValid=$true;AttemptConsumed=$false;NewStopPresent=$false}
                Assert-ODPinBindingFacts $facts
                foreach($k in $facts.Keys){$bad=CM-Copy $facts;if($bad[$k] -is [bool]){$bad[$k]=-not $bad[$k]}elseif($bad[$k] -is [string]){$bad[$k]='invalid'}else{$bad[$k]=2};CM-Reject {Assert-ODPinBindingFacts $bad} 'OD: pin binding recovery mismatch'}
            } $state
            CM-Case $names[17] {
                $changed=@(foreach($rel in $manifest.SourcePaths){$old=Join-Path (Join-Path $packet snapshots/writer-projection-rebind) $rel;if((Get-FileHash (Join-Path $repo $rel)).Hash -cne (Get-FileHash $old).Hash){$rel}})
                CM-Need ($changed.Count -eq 3 -and @(Compare-Object $changed @($expected.Keys)).Count -eq 0) 'exact three-source window'
            } $state
            CM-Case $names[18] {
                . ([scriptblock]::Create((CM-Definition $driverTree Get-ODAttemptSlot).Extent.Text))
                foreach($n in @('contract-green','core','lifecycle','continuation')){CM-Need ((Get-ODAttemptSlot $n) -ceq $n) ('unchanged slot '+$n)}
                $gate=(CM-Definition $driverTree Driver-PinBindingGate).Extent.Text
                CM-Need ($gate.Contains('pin-binding-rebind') -and $gate.Contains('repair-green.frozen.json') -and $gate.Contains('pin-binding.stop.json')) 'fixed gate identities'
            } $state
            CM-Case $names[19] {
                $old=CM-Tree (Join-Path (Join-Path $packet snapshots/writer-projection-rebind) 'docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/driver.ps1')
                foreach($n in @('Driver-ErrorLocation','Driver-UnitError','Driver-FalseFlags','Driver-Step','Driver-Recorder','Driver-Evidence','Driver-ReadEvidence','Driver-Once')){CM-Need ((CM-Definition $driverTree $n).Extent.Text -ceq (CM-Definition $old $n).Extent.Text) ('guard unchanged '+$n)}
                $start='$cleanup=[Collections.Generic.List[string]]::new()';$currentText=$driverTree.Extent.Text;$oldText=$old.Extent.Text
                $a=$currentText.Substring($currentText.IndexOf($start));$b=$oldText.Substring($oldText.IndexOf($start))
                CM-Need ($a.Replace('pin-binding.stop.json','error-projection.stop.json') -ceq $b) 'owned process and timing unchanged'
            } $state
            $exitCode=if($state.Stopped){2}else{0}
        }
        foreach($key in $expected.Keys){$after[$key]=(Get-FileHash -LiteralPath (Join-Path $repo $key)).Hash}
        CM-Need (CM-Same $before $after) 'source bytes unchanged during computation'
        CM-Need ((Get-FileHash $testPath).Hash -ceq $testHash -and (Get-FileHash $inputPath).Hash -ceq $inputHash) 'test/input unchanged'
        if($Phase -ceq 'Green'){CM-Need ((Get-FileHash $candidatePath).Hash -ceq $candidateHash) 'candidate input unchanged'}
    }catch{if($null -eq $state.RootError){$state.RootError=$_}else{$state.RecordingError=$_};$state.Stopped=$true;$exitCode=2}
    foreach($name in $names){if($name -cnotin @($state.Groups|ForEach-Object Name)){$state.Groups.Add([ordered]@{Name=$name;Status='NotRun'})}}
    return [ordered]@{Schema=1;Phase=$Phase;TestSHA256=$testHash;InputSHA256=$inputHash;SourceSHA256=$before;PostSourceSHA256=$after;Groups=$state.Groups.ToArray();PASS=@($state.Groups|Where-Object Status -CEQ PASS).Count;FAIL=@($state.Groups|Where-Object Status -CEQ FAIL).Count;NotRun=@($state.Groups|Where-Object Status -CEQ NotRun).Count;ExitCode=$exitCode;RootError=(CM-Error $state.RootError);RecordingError=(CM-Error $state.RecordingError)}
}
