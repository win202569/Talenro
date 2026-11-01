[CmdletBinding(PositionalBinding=$false)]
param(
    [Parameter(Mandatory)][ValidateSet('Prepare','RunOne')][string]$Action,
    [Parameter(Mandatory)][ValidateSet('seed','writer-green','contract-green','repair-green','core','lifecycle','continuation')][string]$Grant,
    [ValidatePattern('^[A-F0-9]{64}$')][string]$FrozenSHA256,
    [ValidatePattern('^[A-F0-9]{64}$')][string]$PreviousFrozenSHA256
)
try{
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$repo=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))
$packet=Join-Path $repo '.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner'
$engine='C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe'
$inputPaths=@('docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/.gitattributes','docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/README.md','docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/expected-output.json','docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/test-helpers.ps1','docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/entry.tests.ps1','docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/core.tests.ps1','docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/prefix.tests.ps1','docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/owned.tests.ps1','docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/live.tests.ps1','docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/capture.ps1','docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/driver.ps1','scripts/offline-diagnostics-v4/diagnose-offline.ps1','scripts/offline-diagnostics-v4/private/catalog.json','scripts/offline-diagnostics-v4/private/session.psm1','scripts/offline-diagnostics-v4/private/sinks.ps1','scripts/offline-diagnostics-v4/private/owned-step.ps1','scripts/offline-diagnostics-v4/private/sequence.ps1','scripts/offline-diagnostics-v4/private/profiles.ps1','scripts/offline-diagnostics-v4/private/observations.ps1','scripts/offline-diagnostics-v4/.gitattributes','scripts/offline-diagnostics-v4/private/write-boundary.ps1','docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/observation-contract.ps1','docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/fault-stream.cs','docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/record-boundary.tests.ps1','docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/observation-contract.tests.ps1','docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/case-module.ps1','docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/publisher.psm1','docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/recording.tests.ps1')
$inputPaths+=@('scripts/offline-diagnostics-v4/private/record-owner.ps1','docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/expected-observations.json','docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/source-model.ps1')
function Driver-Hash([string]$Path){(Get-FileHash -LiteralPath $Path -Algorithm SHA256).Hash}
function Driver-CheckHash([string]$Path,[string]$Expected){if((Driver-Hash $Path) -cne $Expected){throw ('Input drift/missing: '+$Path)}}
function Driver-Once([string]$Path,[byte[]]$Bytes){
    $stream=$null;$primary=$null
    try{
        $stream=[IO.File]::Open($Path,[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::None)
        $stream.Write($Bytes,0,$Bytes.Length);$stream.Flush()
    }catch{$primary=$_}
    finally{if($null -ne $stream){try{$stream.Dispose()}catch{if($null -eq $primary){$primary=$_}}}}
    if($null -ne $primary){throw $primary}
}
function Driver-JSON([string]$Path,$Value){Driver-Once $Path ([Text.UTF8Encoding]::new($false).GetBytes((ConvertTo-Json -InputObject $Value -Depth 60)))}
function Driver-Verify($Manifest){
    $seen=@{}
    foreach($pin in $Manifest.Pins){
        $key=[IO.Path]::GetFullPath($pin.VerificationPath);$identity=$pin.ExpectedState+':'+$pin.SHA256
        if($seen.ContainsKey($key) -and $seen[$key] -cne $identity){throw ('Conflicting mapping '+$key)}
        $seen[$key]=$identity
        if($pin.ExpectedState -ceq 'Absent'){
            if(Test-Path -LiteralPath $pin.VerificationPath){throw ('Expected absent: '+$pin.VerificationPath)}
            if($null -ne $pin.SHA256){throw 'Absent hash must be null'}
        }elseif($pin.ExpectedState -ceq 'Present'){Driver-CheckHash $pin.VerificationPath $pin.SHA256}
        else{throw 'Unknown pin state'}
    }
}
function Driver-Static {
    Driver-Need ($inputPaths.Count -eq 31 -and @($inputPaths|Select-Object -Unique).Count -eq 31) 'Exact thirty-one source identities'
    foreach($name in @('case-module.ps1','publisher.psm1','recording.tests.ps1','.gitattributes','README.md','expected-output.json','test-helpers.ps1','entry.tests.ps1','core.tests.ps1','prefix.tests.ps1','owned.tests.ps1','live.tests.ps1','capture.ps1','driver.ps1','record-boundary.tests.ps1','observation-contract.tests.ps1','fault-stream.cs','observation-contract.ps1')){
        if(-not [IO.File]::Exists((Join-Path $PSScriptRoot $name))){throw ('Missing explicit frozen source '+$name)}
    }
    foreach($relative in $inputPaths){
        $path=[IO.Path]::GetFullPath((Join-Path $repo $relative))
        if(-not [IO.File]::Exists($path)){throw ('Missing explicit frozen source '+$relative)}
        if([IO.File]::ReadAllBytes($path) -contains 13){throw ('Non-LF source: '+$relative)}
        if($path.EndsWith('.ps1') -or $path.EndsWith('.psm1')){
            $tokens=$null;$errors=$null
            $null=[Management.Automation.Language.Parser]::ParseFile($path,[ref]$tokens,[ref]$errors)
            if($errors.Count -ne 0){throw ('Static parse: '+$relative+' '+($errors.Message -join '; '))}
        }
        if($path.EndsWith('.json')){$null=Get-Content -LiteralPath $path -Raw | ConvertFrom-Json -AsHashtable}
    }
}
Driver-CheckHash $engine '362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139'
Driver-CheckHash (Join-Path ([IO.Path]::GetDirectoryName($engine)) System.Management.Automation.dll) '2D8AF026E589C5603629BFC3E7C57265B4167B12ECA3AE549C6A6C3CA7E4967B'
function Driver-Need([bool]$Condition,[string]$Message){if(-not $Condition){throw $Message}}
function Driver-EqualJSON($Actual,$Expected){
    return ((ConvertTo-Json -InputObject $Actual -Depth 50 -Compress) -ceq (ConvertTo-Json -InputObject $Expected -Depth 50 -Compress))
}

function Driver-PinHash($Manifest,[string]$Relative){
    $path=[IO.Path]::GetFullPath((Join-Path $repo $Relative))
    $pins=@($Manifest.Pins | Where-Object {$_.VerificationPath -ieq $path})
    Driver-Need ($pins.Count -ge 1) 'Operational source pin required'
    foreach($pin in $pins){Driver-Need ($pin.ExpectedState -ceq 'Present' -and $pin.SHA256 -ceq $pins[0].SHA256) 'Conflicting operational source pin'}
    return $pins[0].SHA256
}
function Driver-Expand($Value,$Manifest,[string]$KnownFreeze){
    if($null -eq $Value){return $null}
    if($Value -is [Collections.IDictionary]){$r=[ordered]@{};foreach($key in $Value.Keys){$r[$key]=Driver-Expand $Value[$key] $Manifest $KnownFreeze};return $r}
    if($Value -is [array]){$rows=[Collections.Generic.List[object]]::new();foreach($item in $Value){$rows.Add((Driver-Expand $item $Manifest $KnownFreeze))};return ,($rows.ToArray())}
    if($Value -is [string]){
        if($Value.StartsWith('@repo:')){return [IO.Path]::GetFullPath((Join-Path $repo $Value.Substring(6)))}
        if($Value.StartsWith('@pin:')){return Driver-PinHash $Manifest $Value.Substring(5)}
        if($Value -ceq '@frozen'){return $KnownFreeze}
    }
    return $Value
}
function Driver-FalseFlags($Value){
    foreach($flag in @('SourceVerified','LiveActionsPerformed','RuntimeApproved','FormalAcceptance','RealGateImplemented')){
        Driver-Need ($Value[$flag] -is [bool] -and -not $Value[$flag]) ('False capability flag required '+$flag)
    }
}
function Driver-Keys($Value,[string[]]$Names,[string]$Label){
    Driver-Need ($Value -is [Collections.IDictionary] -and (Driver-EqualJSON @($Value.Keys) $Names)) ('Ordered complete root mismatch '+$Label)
}
function Driver-JSONKeys($Element){
    if($Element.ValueKind -eq [Text.Json.JsonValueKind]::Object){
        $names=[Collections.Generic.HashSet[string]]::new([StringComparer]::Ordinal)
        foreach($property in $Element.EnumerateObject()){
            Driver-Need ($names.Add($property.Name)) 'Duplicate JSON member refused'
            Driver-JSONKeys $property.Value
        }
    }elseif($Element.ValueKind -eq [Text.Json.JsonValueKind]::Array){foreach($item in $Element.EnumerateArray()){Driver-JSONKeys $item}}
}
function Driver-ReadEvidence([string]$Path){
    $bytes=[IO.File]::ReadAllBytes($Path)
    Driver-Need ($bytes.Length -gt 0 -and $bytes.Length -le 1048576) 'Evidence byte bound'
    $text=[Text.UTF8Encoding]::new($false,$true).GetString($bytes)
    Driver-Need (-not $text.StartsWith([string][char]0xFEFF,[StringComparison]::Ordinal)) 'Evidence BOM refused'
    $options=[Text.Json.JsonDocumentOptions]::new();$options.MaxDepth=96
    $options.AllowTrailingCommas=$false;$options.CommentHandling=[Text.Json.JsonCommentHandling]::Disallow
    $document=[Text.Json.JsonDocument]::Parse($text,$options)
    try{Driver-Need ($document.RootElement.ValueKind -eq [Text.Json.JsonValueKind]::Object) 'One JSON root object';Driver-JSONKeys $document.RootElement}
    finally{$document.Dispose()}
    return ConvertFrom-Json -InputObject $text -AsHashtable -Depth 96
}
function Driver-ErrorLocation($Record,[string[]]$Sources){
    Driver-Need ($Record.Position -is [string] -and $Record.Stack -is [string] -and $Record.Stack.Length -gt 0) 'Actual error source/stack required'
    $matched=$false
    foreach($relative in $Sources){
        $source=[IO.Path]::GetFullPath((Join-Path $repo $relative))
        $match=[regex]::Match($Record.Position,([regex]::Escape($source)+':([0-9]+)'),[Text.RegularExpressions.RegexOptions]::IgnoreCase)
        if($match.Success){$line=[int]$match.Groups[1].Value;$matched=$line -gt 0 -and $line -le @([IO.File]::ReadAllLines($source)).Count;if($matched){break}}
    }
    Driver-Need $matched 'Actual error must bind a frozen source and valid line'
}
function Driver-Step($Step,$Oracle,[string]$ProfileId){
    Driver-Keys $Step @('View','Result') StepObservation
    $r=$Step.Result;$v=$Step.View
    Driver-Keys $r @('Accepted','Rejection','AckPresent','Observation','Error') StepResult
    Driver-Need ($r.Accepted -is [bool] -and $r.AckPresent -is [bool]) 'Actual root booleans'
    if($null -ne $v){
        Driver-Keys $v @('Status','FailureLatched','NextOrdinal','Reserved','StartCalls','AckIssued','AckConsumed','Root','Sinks','PrimaryError','SecondaryErrors','Overflow') SessionView
        $case=[IO.Path]::GetFileName($v.Root)
        Driver-Need ($Oracle.Cases.Contains($case) -and $ProfileId -cin $Oracle.Cases[$case].Grants) 'Fixed profile/case binding'
        Driver-Need ($v.Root -ceq [IO.Path]::GetFullPath((Join-Path $packet ('profiles/'+$ProfileId+'/'+$case)))) 'Fixed managed case root'
        Driver-Need ($v.SecondaryErrors.Count -le 8 -and $v.StartCalls -le $Oracle.Cases[$case].Budget) 'Session evidence bounds'
    }else{Driver-Need (-not $r.Accepted -and $null -eq $r.Observation -and $null -eq $r.Error) 'Unknown handle cannot create a step';return $null}
    if($null -ne $r.Error){
        Driver-ErrorLocation $r.Error @('scripts/offline-diagnostics-v4/private/sinks.ps1')
        Driver-Need ($case -cin @('preconsume','prestart','receipt-fault','sequence-output-fault')) 'Unexpected actual step error'
        Driver-Need ($r.Error.Type -ceq 'System.Management.Automation.MethodInvocationException' -and $v.FailureLatched -and $v.AckIssued -eq 0 -and -not $r.AckPresent -and (Driver-EqualJSON $r.Error $v.PrimaryError)) 'Actual negative cause/binding mismatch'
        if($case -ceq 'sequence-output-fault'){
            $expected='Exception calling "Write" with "1" argument(s): "Cannot write to a closed TextWriter.'+[Environment]::NewLine+'Object name: ''StreamWriter''."'
            Driver-Need ($r.Error.ErrorId -ceq 'ObjectDisposedException' -and $r.Error.Message -ceq $expected -and $v.Sinks.ConsoleOut.State -ceq 'Failed' -and $v.Sinks.ConsoleOut.LogicalIOAttempts -eq 1) 'Actual output-fault signature'
        }else{
            $filename=switch -CaseSensitive ($case){'preconsume'{'start.json'}'prestart'{'stdout-0.bin'}'receipt-fault'{'receipt-0.json'}}
            $method=if($case -ceq 'prestart'){'Create'}else{'WriteAllText'}
            $argc=if($case -ceq 'prestart'){1}else{3}
            $expected=('Exception calling "{0}" with "{1}" argument(s): "Access to the path ''{2}'' is denied."' -f $method,$argc,(Join-Path $v.Root $filename))
            Driver-Need ($r.Error.ErrorId -ceq 'UnauthorizedAccessException' -and $r.Error.Message -ceq $expected) 'Actual managed file-fault signature'
        }
    }
    $o=$r.Observation
    if($null -eq $o){return $null}
    Driver-Keys $o @('Started','Association','PID','OwnedExit','ExitCode','StreamsComplete','TimedOut','TerminationRequested','ElapsedMs','KillCalls','WaitCalls','SeamExact','Stdout') OwnedObservation
    foreach($key in @('Started','OwnedExit','StreamsComplete','TimedOut','TerminationRequested','SeamExact')){Driver-Need ($o[$key] -is [bool]) 'Lifecycle boolean type'}
    Driver-Need ($o.ElapsedMs -ge 0) 'Negative actual duration'
    if(-not $o.Started){
        Driver-Need ($case -ceq 'prestart' -and $o.Association -ceq 'Unassociated' -and $null -eq $o.PID -and $o.KillCalls -eq 0 -and $o.WaitCalls -eq 0 -and $v.StartCalls -eq 0 -and $null -ne $r.Error) 'Prestart exact nonassociation'
        return $null
    }
    Driver-Need ($o.Association -ceq 'Associated' -and $o.PID -gt 0 -and $o.OwnedExit -and $o.StreamsComplete) 'Actual direct lifecycle incomplete'
    $max=if($o.SeamExact){114000}elseif($case -ceq 'timeout'){4200}else{6000}
    Driver-Need ($o.ElapsedMs -le $max) 'Actual duration beyond unchanged bound'
    if($case -ceq 'timeout'){
        Driver-Need ($o.TimedOut -and $o.TerminationRequested -and $o.ExitCode -eq -1 -and $null -eq $r.Error) 'Timeout tuple mismatch'
    }elseif($case -ceq 'receipt-fault'){
        Driver-Need (-not $o.TimedOut -and $o.TerminationRequested -and $o.ExitCode -eq -1 -and $v.Sinks.Receipt0.LogicalIOAttempts -eq 1 -and $v.SecondaryErrors.Count -eq 0) 'Receipt no-retry tuple mismatch'
    }else{Driver-Need (-not $o.TimedOut -and -not $o.TerminationRequested -and $o.ExitCode -eq 0) 'Normal direct lifecycle mismatch'}
    return $o
}
function Driver-UnitError($Record,[string]$ExpectedId,[string]$ExpectedMessage,[string]$Source){
    Driver-Keys $Record @('Type','Message','ErrorId','Position','Stack') UnitError
    Driver-Need ($Record.Type -ceq 'System.Management.Automation.MethodInvocationException' -and $Record.ErrorId -ceq $ExpectedId -and $Record.Message -ceq $ExpectedMessage) 'Fixed actual unit error'
    Driver-ErrorLocation $Record @($Source)
    $path=[IO.Path]::GetFullPath((Join-Path $repo $Source))
    $location=[regex]::Match($Record.Position,('^At '+[regex]::Escape($path)+':([0-9]+) char:([0-9]+)\r?\n'),[Text.RegularExpressions.RegexOptions]::CultureInvariant)
    Driver-Need $location.Success 'Actual writer error statement position'
    $operation=if($ExpectedId -ceq 'IOException'){'Open'}elseif($ExpectedId -ceq 'FileNotFoundException'){'Dispose'}elseif($ExpectedMessage.StartsWith('Exception calling "Write"',[StringComparison]::Ordinal)){'Write'}else{'Flush'}
    $tokens=$null;$issues=$null
    $ast=[Management.Automation.Language.Parser]::ParseFile($path,[ref]$tokens,[ref]$issues)
    Driver-Need ($issues.Count -eq 0) 'Frozen writer AST'
    $calls=@($ast.FindAll({param($node)$node -is [Management.Automation.Language.InvokeMemberExpressionAst] -and $node.Member.Value -ceq $operation},$true))
    $matched=$false
    foreach($call in $calls){$extent=Get-ODFailureExtent $call;if($extent.StartLineNumber -eq [int]$location.Groups[1].Value -and $extent.StartColumnNumber -eq [int]$location.Groups[2].Value){$matched=$true}}
    Driver-Need $matched 'Writer cause binds actual operation AST, not any valid line'
}
function Driver-UnitEvidence($Evidence,$Oracle,[string]$CurrentGrant,$Manifest,[string]$KnownFreeze){
    $expected=$Oracle.Grants[$CurrentGrant]
    Driver-Need ($Evidence.NotRun -is [array] -and $Evidence.NotRun.Count -eq 0 -and $Evidence.Status -ceq 'COMPLETE') 'Complete unit execution'
    if($CurrentGrant -ceq 'writer-green'){
        Driver-Need ($Evidence.Observations.Count -eq 11) 'Ten record cases and one owner aggregate'
        for($i=0;$i -lt 10;$i++){
            $row=$Evidence.Observations[$i];$v=$row.Value;$f=$expected.Groups[$i+1].Facts
            Driver-Need ($row.Name -ceq 'RecordCase') 'Record observation order'
            Driver-Keys $v @('Boundary','Case','Outcome','Consumed','Primary','Secondary','OpenCalls','WriteCalls','FlushCalls','DisposeCalls','BytesExact','ExistingUnchanged','FactoryIsolated','Overflow') RecordCase
            Driver-Keys $v.Outcome @('Succeeded','PrimaryError','SecondaryErrors') PrivateOutcomeProjection
            Driver-Need ($v.Boundary -ceq $f.Boundary -and $v.Case -ceq $f.Case -and $v.Secondary -is [array] -and $v.Secondary.Count -eq $f.SecondaryCount -and $v.Overflow -is [bool] -and -not $v.Overflow -and $v.Consumed -is [bool] -and $v.Consumed) 'Record identity and single consumption'
            Driver-Need ($v.Outcome.Succeeded -is [bool] -and $v.Outcome.Succeeded -eq $f.Succeeded -and $v.Outcome.SecondaryErrors -is [array] -and (Driver-EqualJSON $v.Outcome.PrimaryError $v.Primary) -and (Driver-EqualJSON $v.Outcome.SecondaryErrors $v.Secondary)) 'Live identity assertions retain equal projections'
            foreach($key in @('OpenCalls','WriteCalls','FlushCalls','DisposeCalls','BytesExact','ExistingUnchanged','FactoryIsolated')){Driver-Need ($v[$key] -ceq $f[$key] -and $v[$key].GetType() -eq $f[$key].GetType()) ('Record exact field '+$key)}
            $root=Join-Path $packet ('unit/writer-green/'+$v.Boundary+'/'+$v.Case)
            $op=if($v.Case.StartsWith('Write',[StringComparison]::Ordinal)){'Write'}else{'Flush'};$argc=if($op -ceq 'Write'){3}else{0}
            $opMessage=('Exception calling "{0}" with "{1}" argument(s): "Cannot access a closed file."' -f $op,$argc)
            $disposeMessage=('Exception calling "Dispose" with "0" argument(s): "Could not find file ''{0}''."' -f (Join-Path $root absent.txt))
            $openMessage=('Exception calling "Open" with "4" argument(s): "The file ''{0}'' already exists."' -f (Join-Path $root record.bin))
            if($v.Outcome.Succeeded){Driver-Need ($null -eq $v.Primary -and $v.Secondary.Count -eq 0) 'Success has no errors'}
            else{
                Driver-Need ($null -ne $v.Primary) 'Actual first unit cause required'
                $wanted=if($v.Case -cin @('WriteDisposeFirstCause','FlushDisposeFirstCause')){'ObjectDisposedException'}elseif($v.Case -ceq 'DisposeOnly'){'FileNotFoundException'}else{'IOException'}
                $primaryMessage=switch -CaseSensitive ($wanted){'ObjectDisposedException'{$opMessage}'FileNotFoundException'{$disposeMessage}'IOException'{$openMessage}}
                Driver-UnitError $v.Primary $wanted $primaryMessage 'scripts/offline-diagnostics-v4/private/write-boundary.ps1'
                foreach($secondary in $v.Secondary){Driver-UnitError $secondary 'FileNotFoundException' $disposeMessage 'scripts/offline-diagnostics-v4/private/write-boundary.ps1'}
            }
        }
        $owner=$Evidence.Observations[10]
        Driver-Need ($owner.Name -ceq 'OwnerControls' -and $owner.Value -is [array] -and $owner.Value.Count -eq 12 -and $Oracle.OwnerControlNames.Count -eq 12) 'All twelve owner controls'
        $assertions=@(1,2,1,1,9,2,14,1,1,1,1,1)
        $secondaryCounts=@(0,1,1,2,8,0,0,0,0,1,1,1)
        for($i=0;$i -lt 12;$i++){
            $v=$owner.Value[$i]
            Driver-Keys $v @('Name','Assertions','ContractOnly','Primary','Secondary','Overflow','IntentReport') OwnerControl
            Driver-Need ($v.Name -ceq $Oracle.OwnerControlNames[$i] -and $v.Assertions -is [long] -and $v.Assertions -eq $assertions[$i] -and $v.ContractOnly -is [bool] -and $v.ContractOnly -eq ($i -eq 2) -and $v.Secondary -is [array] -and $v.Secondary.Count -eq $secondaryCounts[$i] -and $v.Overflow -is [bool] -and $v.Overflow -eq ($i -eq 4)) 'Literal owner control evidence'
            Driver-Need (($null -eq $v.Primary) -eq ($i -eq 8)) 'Owner primary presence'
            if($i -eq 7){
                $r=$v.IntentReport
                Driver-Keys $r $Oracle.ProfileRootNames IntentFailureReport
                Driver-Need ($r.Schema -eq 2 -and $r.Profile -ceq 'entry-migration-core' -and $r.Status -ceq 'INCOMPLETE' -and $r.Groups -is [array] -and $r.Groups.Count -eq 0 -and $r.Observations -is [array] -and $r.Observations.Count -eq 0 -and $r.Closed -is [bool] -and $r.Closed -and $r.Overflow -is [bool] -and -not $r.Overflow -and $r.SecondaryErrors -is [array] -and $r.SecondaryErrors.Count -eq 0 -and (Driver-EqualJSON $r.PrimaryError $v.Primary)) 'Known intent failure no work/report references'
                Driver-Need ($r.FrozenSHA256 -ceq $KnownFreeze -and $r.EntryPath -ceq [IO.Path]::GetFullPath((Join-Path $repo 'scripts/offline-diagnostics-v4/diagnose-offline.ps1')) -and $r.EntrySHA256 -ceq (Driver-PinHash $Manifest 'scripts/offline-diagnostics-v4/diagnose-offline.ps1') -and $r.KernelPath -ceq [IO.Path]::GetFullPath((Join-Path $repo 'scripts/offline-diagnostics-v4/private/session.psm1')) -and $r.KernelSHA256 -ceq (Driver-PinHash $Manifest 'scripts/offline-diagnostics-v4/private/session.psm1')) 'Unit intent bound to current writer freeze, not future repair'
                Driver-FalseFlags $r
                $message=('Exception calling "Open" with "4" argument(s): "The file ''{0}'' already exists."' -f (Join-Path $packet 'unit/writer-green/owner-controls/intent-existing.bin'))
                Driver-UnitError $r.PrimaryError 'IOException' $message 'scripts/offline-diagnostics-v4/private/write-boundary.ps1'
            }else{Driver-Need ($null -eq $v.IntentReport) 'No invented intent report on other controls'}
        }
    }else{
        Driver-Need ($CurrentGrant -ceq 'contract-green' -and $Evidence.Observations.Count -eq 1 -and $Evidence.Observations[0].Name -ceq 'SchemaSuite') 'One bounded contract aggregate'
        $cases=$Evidence.Observations[0].Value
        Driver-Need ($cases -is [array] -and $cases.Count -eq 49 -and $Oracle.ContractGroupNames.Count -eq 49) 'All forty-nine actual contract cases'
        $ids=@('entry-migration-core','entry-migration-owned','entry-migration-live')
        for($i=0;$i -lt 49;$i++){
            $v=$cases[$i]
            Driver-Keys $v @('Name','Profile','Kind','ExpectedRejected','ActualRejected','Error') SchemaCase
            $reject=($i -ge 4 -and $i -le 23) -or ($i -ge 28 -and $i -le 45)
            $kind=if($i -eq 0){'Availability'}elseif($i -le 3){'Baseline'}elseif($i -le 23){'Mutation'}elseif($i -le 26){'Schema2'}elseif($i -eq 27){'IntentFailure'}elseif($i -le 45){'Schema2Mutation'}else{'Static'}
            $profileId=if($i -eq 0 -or $i -ge 46){''}elseif($i -le 3){$ids[$i-1]}elseif($i -ge 24 -and $i -le 26){$ids[$i-24]}elseif($i -eq 19){$ids[2]}elseif($i -cin @(20,21)){$ids[1]}else{$ids[0]}
            Driver-Need ($v.Name -ceq $Oracle.ContractGroupNames[$i] -and $v.Kind -ceq $kind -and $v.Profile -ceq $profileId -and $v.ExpectedRejected -is [bool] -and $v.ExpectedRejected -eq $reject -and $v.ActualRejected -is [bool] -and $v.ActualRejected -eq $reject) 'Fixed complete contract case disposition'
            if($v.ActualRejected){
                Driver-Keys $v.Error @('Type','Message','ErrorId','Position','Stack') ContractError
                Driver-Need ($v.Error.Type -ceq 'System.Management.Automation.RuntimeException' -and $v.Error.Message -ceq 'OD: observation contract mismatch' -and $v.Error.ErrorId -ceq 'OD: observation contract mismatch') 'Only exact contract error is rejection'
                Driver-ErrorLocation $v.Error @('docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/observation-contract.ps1')
            }else{Driver-Need ($null -eq $v.Error) 'No unexpected unit exception'}
        }
    }
}

. (Join-Path $PSScriptRoot observation-contract.ps1)
. (Join-Path $PSScriptRoot source-model.ps1)

function Driver-Recorder($Evidence,$Oracle,[string]$CurrentGrant,$Manifest){
    $records=@($Evidence.Observations | Where-Object {$_.Name -ceq 'Recorder'})
    Driver-Need ($records.Count -eq 1 -and $Evidence.Observations[-1].Name -ceq 'Recorder') 'One final recorder observation'
    $r=$records[0].Value
    Driver-Keys $r $Oracle.Recorder.RootKeys Recorder
    $writerPath=[IO.Path]::GetFullPath((Join-Path $repo 'scripts/offline-diagnostics-v4/private/write-boundary.ps1'))
    Driver-Keys $r.Publisher @('Name','Path','WriterFile','FactoryFile','WriterSHA256','DefaultFactory','CollectorStateAbsent') PublisherBinding
    Driver-Need ($r.Publisher.Name -ceq 'publisher' -and $r.Publisher.Path -ceq [IO.Path]::GetFullPath((Join-Path $PSScriptRoot publisher.psm1)) -and $r.Publisher.WriterFile -ceq $writerPath -and $r.Publisher.FactoryFile -ceq $writerPath -and $r.Publisher.WriterSHA256 -ceq (Driver-PinHash $Manifest 'scripts/offline-diagnostics-v4/private/write-boundary.ps1')) 'Actual production publisher binding'
    foreach($key in @('DefaultFactory','CollectorStateAbsent')){Driver-Need ($r.Publisher[$key] -is [bool] -and $r.Publisher[$key]) 'Publisher default factory and absent collector'}
    $writerGrant=$CurrentGrant -ceq 'writer-green'
    $expectedCases=if($writerGrant){12}else{0}
    $caseInventory=@($Oracle.WriterCases)+@([ordered]@{Boundary='Owner';Case='Controls'},[ordered]@{Boundary='Capture';Case='FinalWriteControl'})
    Driver-Need ($r.Cases -is [array] -and $r.Cases.Count -eq $expectedCases) 'Exact case module inventory'
    $stages=[Collections.Generic.List[object]]::new()
    $stages.Add([ordered]@{Stage='Before';Case=''})
    if($writerGrant){foreach($c in $Oracle.WriterCases){foreach($stage in @('Created','Replaced','Finally')){$stages.Add([ordered]@{Stage=$stage;Case=$c.Name})}}}
    $stages.Add([ordered]@{Stage='After';Case=''})
    Driver-Need ($r.Stages -is [array] -and $r.Stages.Count -eq $stages.Count) 'Complete recorder stages'
    for($i=0;$i -lt $stages.Count;$i++){
        $actual=$r.Stages[$i]
        Driver-Keys $actual @('Stage','Case','Commands','CollectorSame') RecorderStage
        Driver-Need ($actual.Stage -ceq $stages[$i].Stage -and $actual.Case -ceq $stages[$i].Case -and $actual.Commands -is [array] -and $actual.Commands.Count -eq 6 -and $actual.CollectorSame -is [bool] -and $actual.CollectorSame) 'Recorder exact stage and collector identity/order'
        $names=@('New-ODRecordStream','Write-ODCaptureRecord','Write-ODProfileIntentRecord','Register-ODRecordOutcome','Import-ODProfileFailure','Capture-Failure')
        for($j=0;$j -lt 6;$j++){
            $command=$actual.Commands[$j];Driver-Keys $command $Oracle.Recorder.CommandKeys RecorderCommand
            Driver-Need ($command.Name -ceq $names[$j] -and $command.SameOriginal -is [bool] -and $command.SameOriginal -and $command.ModuleName -is [string] -and $command.ModuleName -ceq '') 'Actual outer references isolated'
        }
    }
    $moduleNames=[Collections.Generic.List[string]]::new()
    for($i=0;$i -lt $expectedCases;$i++){
        $c=$r.Cases[$i];$case=$caseInventory[$i]
        $primaryNull=if($i -lt 10){$case.Case -ceq 'SuccessOnce'}else{$true}
        $secondaryCount=if($i -lt 10){$Oracle.Grants[$CurrentGrant].Groups[$i+1].Facts.SecondaryCount}else{0}
        Driver-Keys $c @('Boundary','Case','ModuleName','Exports','State') RecorderCase
        Driver-Need ($c.Boundary -ceq $case.Boundary -and $c.Case -ceq $case.Case -and $c.ModuleName -is [string] -and $c.ModuleName -cmatch '^__DynamicModule_[a-f0-9-]{36}$') 'Case module role identity'
        $moduleNames.Add($c.ModuleName)
        Driver-Keys $c.Exports $Oracle.Recorder.ExportsKeys NoExports
        foreach($key in $Oracle.Recorder.ExportsKeys){Driver-Need ($c.Exports[$key] -is [long] -and $c.Exports[$key] -eq 0) 'Case exports must be zero'}
        Driver-Keys $c.State @('Initialized','PrimaryNull','SecondaryCount','Overflow') CaseState
        Driver-Need ($c.State.Initialized -is [bool] -and $c.State.Initialized -and $c.State.PrimaryNull -is [bool] -and $c.State.PrimaryNull -eq $primaryNull -and $c.State.SecondaryCount -eq $secondaryCount -and $c.State.Overflow -is [bool] -and -not $c.State.Overflow) 'Explicit isolated case state'
    }
    Driver-Need (@($moduleNames.ToArray()|Select-Object -Unique).Count -eq $expectedCases) 'Distinct private case modules'
    Driver-Keys $r.Outcome @('Succeeded','PrimaryError','SecondaryErrors') PublisherProbeOutcome
    Driver-Need ($r.Outcome.Succeeded -is [bool] -and $r.Outcome.Succeeded -and $null -eq $r.Outcome.PrimaryError -and $r.Outcome.SecondaryErrors -is [array] -and $r.Outcome.SecondaryErrors.Count -eq 0 -and $r.Consumed -is [bool] -and $r.Consumed) 'Actual default probe outcome consumed once'
    $probePath=Join-Path $packet ('capture/'+$CurrentGrant+'/publisher-probe.json')
    $probe=[Text.Encoding]::UTF8.GetBytes($Oracle.Recorder.ProbeText)
    Driver-Need ($r.ProbePath -ceq $probePath -and $r.ProbeBytes -ceq [Convert]::ToBase64String($probe) -and [Linq.Enumerable]::SequenceEqual([byte[]][IO.File]::ReadAllBytes($probePath),[byte[]]$probe)) 'Actual production writer probe bytes'
    $journal=Join-Path $packet ('capture/'+$CurrentGrant+'/journal.jsonl')
    $expectedJournal='{"Stage":"begin"}'+[char]10+'{"Stage":"tests-complete"}'+[char]10+'{"Stage":"published"}'+[char]10
    Driver-Need ((Get-Item -LiteralPath $journal).Length -le 1048576 -and [Text.UTF8Encoding]::new($false,$true).GetString([IO.File]::ReadAllBytes($journal)) -ceq $expectedJournal) 'Complete independent journal required'
}

function Driver-Evidence($Evidence,$Profile,$Oracle,[string]$CurrentGrant,$Manifest,[string]$KnownFreeze){
    $e=$Evidence
    Driver-Keys $e $Oracle.EvidenceRootNames Evidence
    # Validate every returned representation before deciding success/failure.
    # INCOMPLETE with no rows is a known failure representation, never a pass.
    # The actual parent has no historical/unit-context fallback.
    if($CurrentGrant -cin @('core','lifecycle','continuation')){
        $returns=@($e.Observations|Where-Object Name -CEQ ProfileReturn)
        Driver-Need ($returns.Count -eq 1) 'one complete profile representation required'
        $context=New-ODRuntimeContext $repo $Profile.Profile $Manifest $KnownFreeze
        Assert-ODProfileReport $returns[0].Value $Profile.Profile $context
        if($returns[0].Value.Status -ceq 'COMPLETE'){
            Assert-ODProfileObservations $returns[0].Value.Observations $Profile.Profile $context
        }
        Driver-Need ($returns[0].Value.Status -ceq 'COMPLETE') 'valid failure representation remains unsuccessful'
    }
    Driver-Need ($e.Schema -eq 1 -and $e.Grant -ceq $CurrentGrant -and $e.Status -ceq $Profile.Status -and -not $e.Overflow -and $null -eq $e.PrimaryError -and $e.SecondaryErrors.Count -eq 0 -and $null -eq $e.CurrentGroup -and $null -eq $e.Probe) 'Capture status/controller fault mismatch'
    Driver-Need ((Driver-EqualJSON $e.Groups $Profile.Groups) -and (Driver-EqualJSON $e.NotRun $Profile.NotRun)) 'Ordered groups/NotRun mismatch'
    Driver-FalseFlags $e
    Driver-Need ($e.Observations -is [array] -and $e.Observations.Count -le 32) 'Capture observation bounds'
    Driver-Recorder $e $Oracle $CurrentGrant $Manifest
    if($CurrentGrant -cin @('writer-green','contract-green')){$unit=$e.Clone();$unit.Observations=@($e.Observations | Where-Object {$_.Name -cne 'Recorder'});Driver-UnitEvidence $unit $Oracle $CurrentGrant $Manifest $KnownFreeze;return 0}
    $profileReturns=@($e.Observations | Where-Object {$_.Name -ceq 'ProfileReturn'})
    $expectedReturns=if($CurrentGrant -cin @('core','lifecycle','continuation')){1}else{0}
    Driver-Need ($profileReturns.Count -eq $expectedReturns) 'Exactly one canonical profile return'
    foreach($row in $e.Observations){Driver-Keys $row @('Name','Value') Observation}
    if($CurrentGrant -cin @('core','lifecycle','continuation')){
        $inspects=@($e.Observations | Where-Object {$_.Name -cin @('InspectDefault','InspectExplicit')})
        Driver-Need ($inspects.Count -eq 2) 'Two actual Inspect calls required'
        $key='Present'
        $expected=Driver-Expand $Oracle.InspectTemplates[$key] $Manifest $KnownFreeze
        foreach($row in $inspects){
            Driver-Need ((Driver-EqualJSON $row.Value.Report $expected) -and -not $row.Value.RuntimeLoaded -and (Driver-EqualJSON $row.Value.Before $row.Value.After) -and (Driver-EqualJSON $row.Value.Roots $Profile.ExistingProfiles)) 'Actual read-only Inspect mismatch'
        }
        $rejections=@($e.Observations | Where-Object {$_.Name -ceq 'RejectedRequest'})
        Driver-Need ($rejections.Count -eq 14) 'All actual forbidden requests required'
        for($i=0;$i -lt 14;$i++){
            $actual=$rejections[$i].Value;$want=$Oracle.Rejections[$i]
            Driver-ErrorLocation $actual.Error @('scripts/offline-diagnostics-v4/diagnose-offline.ps1','docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/entry.tests.ps1')
            Driver-Need ($actual.Name -ceq $want.Name -and -not $actual.RuntimeLoaded -and $actual.Error.Type -ceq $want.Type -and $actual.Error.ErrorId -cin $want.Ids -and $actual.Error.Message -ceq $want.Message) 'Fixed forbidden-request actual signature'
        }
        $tree=@($e.Observations | Where-Object {$_.Name -ceq 'EntryTree'})
        Driver-Need ($tree.Count -eq 1 -and (Driver-EqualJSON $tree[0].Value.Before $tree[0].Value.After) -and (Driver-EqualJSON $tree[0].Value.Roots $Profile.ExistingProfiles)) 'Inspect/rejections changed managed profile tree'
    }
    $started=[Collections.Generic.List[object]]::new()
    if($expectedReturns -eq 1){
        $r=$profileReturns[0].Value
        Driver-Keys $r $Oracle.ProfileRootNames ProfileReport
        Driver-Need ($r.Schema -eq 2 -and $r.Profile -ceq $Profile.Profile -and $r.Status -ceq 'COMPLETE' -and $r.Closed -and $null -eq $r.PrimaryError -and $r.SecondaryErrors.Count -eq 0 -and -not $r.Overflow -and $r.FrozenSHA256 -ceq $KnownFreeze) 'Private runtime report status mismatch'
        Driver-FalseFlags $r
        Driver-Need ($r.EntryPath -ceq [IO.Path]::GetFullPath((Join-Path $repo 'scripts/offline-diagnostics-v4/diagnose-offline.ps1')) -and $r.KernelPath -ceq [IO.Path]::GetFullPath((Join-Path $repo 'scripts/offline-diagnostics-v4/private/session.psm1')) -and $r.EntrySHA256 -ceq (Driver-PinHash $Manifest 'scripts/offline-diagnostics-v4/diagnose-offline.ps1') -and $r.KernelSHA256 -ceq (Driver-PinHash $Manifest 'scripts/offline-diagnostics-v4/private/session.psm1')) 'Actual source identity differs from independently frozen pins'
        $offset=6
        $count=$Profile.Groups.Count-$offset-5
        $expectedGroups=@($Profile.Groups | Select-Object -Skip $offset -First $count)
        Driver-Need (Driver-EqualJSON $r.Groups $expectedGroups) 'Profile ordered group projection mismatch'
        Driver-Need ($r.Observations -is [array] -and $r.Observations.Count -le 32) 'Profile observation bounds'
        foreach($detail in $r.Observations){
            Driver-Keys $detail @('Name','Value') ProfileObservation
            if($detail.Name -ceq 'Step'){$owned=Driver-Step $detail.Value $Oracle $Profile.Profile;if($null -ne $owned){$started.Add($owned)}}
        }
        $reentry=@($e.Observations | Where-Object {$_.Name -ceq 'ProfileReentry'})
        Driver-Need ($reentry.Count -eq 1) 'Single actual public reentry required'
        $v=$reentry[0].Value
        Driver-ErrorLocation $v.Error @('scripts/offline-diagnostics-v4/private/profiles.ps1','scripts/offline-diagnostics-v4/private/session.psm1')
        Driver-Need ($v.SameModule -and (Driver-EqualJSON $v.Before $v.After) -and $v.Error.Type -ceq 'System.Management.Automation.RuntimeException' -and $v.Error.ErrorId -ceq 'OD: profile already invoked' -and $v.Error.Message -ceq 'OD: profile already invoked') 'Actual reentry guard mismatch'
    }
    Driver-Need ($started.Count -eq $Profile.Children) 'Exact direct child count mismatch'
    if($started.Count -gt 0){Driver-Need (@($started.PID | Select-Object -Unique).Count -eq $started.Count) 'Distinct owned direct targets required'}
    return $started.Count
}
# New preparation uses named immutable inputs, never a historical adjustment
# replay, a repository glob, or the mutable authoring ledger as a runtime pin.
Driver-Need (-not (Test-Path -LiteralPath (Join-Path $packet stop.json))) 'Execution stopped; no refreeze'
function Driver-AddPin($Map,$Rows,[string]$Path,[string]$Hash,[string]$Role){
    $canonical=[IO.Path]::GetFullPath($Path)
    Driver-Need ($Hash -cmatch '^[A-F0-9]{64}$') 'Present pin requires complete SHA256'
    if($Map.ContainsKey($canonical)){
        $prior=$Map[$canonical]
        Driver-Need ($prior.ExpectedState -ceq 'Present' -and $prior.SHA256 -ceq $Hash) 'Conflicting canonical source identity'
        return
    }
    $pin=[pscustomobject][ordered]@{VerificationPath=$canonical;SHA256=$Hash;ExpectedState='Present';Role=$Role}
    $Map.Add($canonical,$pin);$Rows.Add($pin)
}
function Driver-RequireAvailability([string]$Name,[string]$SeedSHA256,[int]$Total,[string[]]$Names){
    $record=Get-Content -LiteralPath (Join-Path $packet ($Name+'.final.json')) -Raw|ConvertFrom-Json -AsHashtable
    Driver-Keys $record @('Schema','Grant','FrozenSHA256','ToolExitCode','Report') AvailabilityReceipt
    Driver-Need ($record.Schema -eq 1 -and $record.Grant -ceq $Name -and $record.FrozenSHA256 -ceq $SeedSHA256 -and $record.ToolExitCode -eq 1) 'Actual availability computation binding'
    $report=$record.Report
    Driver-Keys $report @('Schema','PASS','FAIL','NotRun','ExitCode','Groups') AvailabilityReport
    foreach($key in @('Schema','PASS','FAIL','NotRun','ExitCode')){Driver-Need ($report[$key] -is [long]) 'Availability JSON integer kinds'}
    Driver-Need ($report.Schema -eq 1 -and $report.PASS -eq 0 -and $report.FAIL -eq 1 -and $report.NotRun -eq ($Total-1) -and $report.ExitCode -eq 1 -and $report.Groups -is [array] -and $report.Groups.Count -eq $Total -and $Names.Count -eq $Total) 'Exact missing-API result'
    for($i=0;$i -lt $Total;$i++){
        Driver-Keys $report.Groups[$i] @('Name','Status') AvailabilityGroup
        Driver-Need ($report.Groups[$i].Name -is [string] -and $report.Groups[$i].Status -is [string] -and $report.Groups[$i].Name -ceq $Names[$i] -and $report.Groups[$i].Status -ceq $(if($i -eq 0){'FAIL'}else{'NotRun'})) 'Exact availability group order'
    }
}
function Driver-RequireFinal([string]$Name){
    $record=Get-Content -LiteralPath (Join-Path $packet ($Name+'.final.json')) -Raw|ConvertFrom-Json -AsHashtable
    foreach($key in @('Started','OwnedExit','StreamsComplete','PostPinsIntact','Exact','EvidenceValid')){Driver-Need ($record[$key] -is [bool] -and $record[$key]) 'Prior actual host must be exact'}
    foreach($key in @('TimedOut','TerminationRequested','Overflow','Rescue')){Driver-Need ($record[$key] -is [bool] -and -not $record[$key]) 'Prior host cannot require rescue'}
    Driver-Need ($record.Grant -ceq $Name -and $record.ExitCode -eq 0 -and $null -eq $record.PrimaryError -and $record.SecondaryErrors -is [array] -and $record.SecondaryErrors.Count -eq 0) 'Prior host disposition'
    $phase=if($Name -ceq 'writer-green'){'writer-green'}else{'repair-green'}
    Driver-Need ($record.FrozenSHA256 -ceq (Driver-Hash (Join-Path $packet ($phase+'.frozen.json')))) 'Prior actual host source epoch'
    Driver-FalseFlags $record
    Driver-Need ($record.EvidenceSHA256 -ceq (Driver-Hash (Join-Path $packet ($Name+'.evidence.json'))) -and $record.StdoutSHA256 -ceq (Driver-Hash (Join-Path $packet ($Name+'.stdout.bin'))) -and $record.StderrSHA256 -ceq (Driver-Hash (Join-Path $packet ($Name+'.stderr.bin')))) 'Prior original stream/evidence identity'
}
if($Action -ceq 'Prepare'){
    Driver-Need ($Grant -cin @('seed','writer-green','repair-green')) 'Only three approved source freezes'
    Driver-Static
    $manifestPath=Join-Path $packet ($Grant+'.frozen.json')
    Driver-Need (-not (Test-Path -LiteralPath $manifestPath)) 'Prepare attempted; no retry'
    if($Grant -ceq 'seed'){
        Driver-Need ([string]::IsNullOrEmpty($PreviousFrozenSHA256)) 'Seed has no previous new freeze'
        foreach($name in @('writer-green','contract-green','core','lifecycle','continuation')){
            foreach($suffix in @('intent.json','final.json','stdout.bin','stderr.bin','evidence.json')){Driver-Need (-not (Test-Path -LiteralPath (Join-Path $packet ($name+'.'+$suffix)))) 'Runtime output cannot precede first freeze'}
        }
        foreach($relative in @('scripts/offline-diagnostics-v4/private/write-boundary.ps1','scripts/offline-diagnostics-v4/private/record-owner.ps1','docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/observation-contract.ps1')){
            $tokens=$null;$issues=$null
            $ast=[Management.Automation.Language.Parser]::ParseFile((Join-Path $repo $relative),[ref]$tokens,[ref]$issues)
            Driver-Need ($issues.Count -eq 0 -and @($ast.FindAll({param($node)$node -is [Management.Automation.Language.FunctionDefinitionAst]},$true)).Count -eq 0) 'Availability seed is definition-free'
        }
        $spec=Join-Path $repo 'docs/superpowers/specs/2026-11-01-offline-i1-i2-single-owner-design.md'
        $plan=Join-Path $repo 'docs/superpowers/plans/2026-11-01-offline-i1-i2-single-owner.md'
        Driver-CheckHash $spec 'A1173B1781FD01D9FB0195FA3BF4EF07D702902C0562355CB79EE9CF08164951'
        Driver-CheckHash $plan '4448A8EDC17572625EDA26950137C17E6E66F0A5790C4E7E09A093A5FBDEF3CC'
        $oldManifest=Join-Path $repo '.superpowers/sdd/2026-10-31-offline-integrated-review-repair/seed.frozen.json'
        Driver-CheckHash $oldManifest '9411278BDB6636F38C1D643C647C754194B0D67249958342BF26FC49D6DFDDC1'
        $old=Get-Content -LiteralPath $oldManifest -Raw|ConvertFrom-Json
        Driver-Need ($old.Schema -eq 2 -and $old.Phase -ceq 'seed' -and $old.Pins.Count -eq 2482) 'Actual inherited seed mapping'
        Driver-Verify $old
        $pins=[Collections.Generic.List[object]]::new()
        $map=[Collections.Generic.Dictionary[string,object]]::new([StringComparer]::OrdinalIgnoreCase)
        foreach($pin in $old.Pins){
            Driver-Need ($pin.ExpectedState -ceq 'Present') 'Inherited Present remains Present'
            Driver-AddPin $map $pins $pin.VerificationPath $pin.SHA256 'inherited-v3-seed'
        }
        $inheritedUnique=$pins.Count
        foreach($relative in $inputPaths){$path=Join-Path $repo $relative;Driver-AddPin $map $pins $path (Driver-Hash $path) 'single-owner-source'}
        foreach($path in @($spec,$plan,$oldManifest)){Driver-AddPin $map $pins $path (Driver-Hash $path) 'immutable-provenance'}
        $consents=@('approval.json','recovery-approval.json','prefreeze-authoring-consent.json')
        foreach($name in $consents){$path=Join-Path $packet $name;Driver-AddPin $map $pins $path (Driver-Hash $path) 'explicit-consent'}
        # Existing approved document snapshots came from the original consent
        # recording step. Verify rather than overwrite or recopy them.
        Driver-CheckHash (Join-Path $packet approved-spec.snapshot.md) (Driver-Hash $spec)
        Driver-CheckHash (Join-Path $packet approved-plan.snapshot.md) (Driver-Hash $plan)
        foreach($name in $consents){Driver-Once (Join-Path $packet ('approved-'+$name+'.snapshot.json')) ([IO.File]::ReadAllBytes((Join-Path $packet $name)))}
        $manifest=[pscustomobject][ordered]@{
            Schema=2;Phase=$Grant;BaseCommit='7d4570bd0389b00a8296b2e3d3a9ce0090a7359a'
            PlanSHA256=(Driver-Hash $plan);DesignSHA256=(Driver-Hash $spec);InheritedManifestSHA256=(Driver-Hash $oldManifest)
            InheritedPins=2482;InheritedUniquePins=$inheritedUnique;PinCount=$pins.Count;SourceCount=31;SourcePaths=[string[]]$inputPaths
            Pins=$pins.ToArray();ChangedInputs=[string[]]@();PreviousFrozenSHA256=$null
            MaxAvailabilityREDs=2;MaxHosts=5;MaxChildren=7;MaxCompilations=1;MaxReviewers=1
            SourceVerified=$false;LiveActionsPerformed=$false;RuntimeApproved=$false;FormalAcceptance=$false;RealGateImplemented=$false
        }
    }else{
        Driver-Need (-not [string]::IsNullOrEmpty($PreviousFrozenSHA256)) 'Previous identity required'
        $previous=if($Grant -ceq 'writer-green'){'seed'}else{'writer-green'}
        $previousPath=Join-Path $packet ($previous+'.frozen.json');Driver-CheckHash $previousPath $PreviousFrozenSHA256
        $manifest=Get-Content -LiteralPath $previousPath -Raw|ConvertFrom-Json
        Driver-Need ($manifest.Phase -ceq $previous -and $manifest.Pins.Count -eq $manifest.PinCount -and $manifest.SourceCount -eq 31 -and (Driver-EqualJSON $manifest.SourcePaths $inputPaths)) 'Prior source ownership and manifest'
        if($Grant -ceq 'writer-green'){
            $oracle=Get-Content -LiteralPath (Join-Path $PSScriptRoot expected-output.json) -Raw|ConvertFrom-Json -AsHashtable
            Driver-RequireAvailability 'writer-availability-red' $PreviousFrozenSHA256 26 $oracle.WriterGroupNames
            $contractNames=@($oracle.ContractGroupNames)+@($oracle.Recorder.Names)
            Driver-RequireAvailability 'contract-availability-red' $PreviousFrozenSHA256 52 $contractNames
        }else{Driver-RequireFinal 'writer-green'}
        [string[]]$allowed=if($Grant -ceq 'writer-green'){@('scripts/offline-diagnostics-v4/private/write-boundary.ps1','scripts/offline-diagnostics-v4/private/record-owner.ps1')}else{@('docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/observation-contract.ps1')}
        $changed=[Collections.Generic.List[string]]::new()
        foreach($pin in $manifest.Pins){
            $relative=[IO.Path]::GetRelativePath($repo,$pin.VerificationPath).Replace('\','/')
            if($relative -cin $allowed){
                $hash=Driver-Hash $pin.VerificationPath
                Driver-Need ($pin.ExpectedState -ceq 'Present' -and $pin.SHA256 -cne $hash) 'Every specified source must change'
                $pin.SHA256=$hash;$changed.Add($relative)
            }else{Driver-CheckHash $pin.VerificationPath $pin.SHA256}
        }
        Driver-Need ($changed.Count -eq $allowed.Count -and @(Compare-Object $changed.ToArray() $allowed).Count -eq 0) 'Exact approved source window'
        $manifest.Phase=$Grant;$manifest.ChangedInputs=$changed.ToArray();$manifest.PreviousFrozenSHA256=$PreviousFrozenSHA256
    }
    Driver-Need ($manifest.Pins.Count -eq $manifest.PinCount -and $manifest.InheritedPins -eq 2482 -and $manifest.InheritedUniquePins -gt 0 -and $manifest.InheritedUniquePins -le $manifest.InheritedPins -and $manifest.PinCount -ge $manifest.InheritedUniquePins -and $manifest.SourceCount -eq 31) 'Disclosed original-row and canonical-unique mapping counts'
    Driver-Verify $manifest
    $base=Join-Path $packet ('snapshots/'+$Grant);Driver-Need (-not (Test-Path -LiteralPath $base)) 'Snapshot attempted'
    foreach($relative in $inputPaths){
        $target=Join-Path $base $relative;$null=[IO.Directory]::CreateDirectory([IO.Path]::GetDirectoryName($target))
        Driver-Once $target ([IO.File]::ReadAllBytes((Join-Path $repo $relative)))
        Driver-CheckHash $target (Driver-Hash (Join-Path $repo $relative))
    }
    Driver-JSON $manifestPath $manifest
    [ordered]@{Prepared=$Grant;FrozenSHA256=(Driver-Hash $manifestPath);Pins=$manifest.PinCount;InheritedPins=2482;InheritedUniquePins=$manifest.InheritedUniquePins;Snapshots=31;ChangedInputs=$manifest.ChangedInputs}|ConvertTo-Json
    return
}
Driver-Need (-not (Test-Path -LiteralPath (Join-Path $packet stop.json))) 'Execution stopped; no retry'
Driver-Need (-not [string]::IsNullOrEmpty($FrozenSHA256)) 'Missing frozen identity'
Driver-Need ($Grant -cin @('writer-green','contract-green','core','lifecycle','continuation')) 'Only five approved capture hosts'
$phase=if($Grant -ceq 'writer-green'){'writer-green'}else{'repair-green'}
$manifestPath=Join-Path $packet ($phase+'.frozen.json')
Driver-CheckHash $manifestPath $FrozenSHA256
$manifest=Get-Content -LiteralPath $manifestPath -Raw | ConvertFrom-Json
Driver-Need ($manifest.Schema -eq 2 -and $manifest.Phase -ceq $phase -and $manifest.Pins.Count -eq $manifest.PinCount -and $manifest.InheritedPins -eq 2482 -and $manifest.InheritedUniquePins -gt 0 -and $manifest.InheritedUniquePins -le $manifest.InheritedPins -and $manifest.PinCount -ge $manifest.InheritedUniquePins -and $manifest.SourceCount -eq 31 -and (Driver-EqualJSON $manifest.SourcePaths $inputPaths) -and $manifest.MaxHosts -eq 5 -and $manifest.MaxChildren -eq 7 -and $manifest.MaxCompilations -eq 1 -and $manifest.MaxReviewers -eq 1) 'Exact new source ownership and finite budget'
Driver-Verify $manifest
Driver-FalseFlags $manifest
Driver-Need ($manifest.MaxAvailabilityREDs -eq 2) 'Only two availability computations'
switch -CaseSensitive ($Grant){
    'contract-green' {Driver-RequireFinal 'writer-green'}
    'core' {Driver-RequireFinal 'contract-green'}
    'lifecycle' {Driver-RequireFinal 'core'}
    'continuation' {Driver-RequireFinal 'lifecycle'}
}
$oracle=Get-Content -LiteralPath (Join-Path $PSScriptRoot expected-output.json) -Raw | ConvertFrom-Json -AsHashtable
$intentPath=Join-Path $packet ($Grant+'.intent.json');$finalPath=Join-Path $packet ($Grant+'.final.json')
foreach($suffix in @('intent.json','final.json','stdout.bin','stderr.bin','evidence.json')){Driver-Need (-not (Test-Path -LiteralPath (Join-Path $packet ($Grant+'.'+$suffix)))) 'Grant attempted/unknown; no retry'}
Driver-JSON $intentPath ([ordered]@{Grant=$Grant;AttemptConsumed=$true;FrozenSHA256=$FrozenSHA256;AuthorizedHost=1;AuthorizedChildren=$oracle.Grants[$Grant].Children})
$row=[ordered]@{Grant=$Grant;Started=$false;PID=$null;OwnedExit=$false;StreamsComplete=$false;ExitCode=$null;TimedOut=$false;TerminationRequested=$false;PrimaryError=$null;SecondaryErrors=@();Overflow=$false;PostPinsIntact=$false;Exact=$false;StdoutSHA256=$null;StderrSHA256=$null;EvidenceSHA256=$null;EvidenceValid=$false;ChildrenObserved=0;EvidenceState='Unknown';ExpectedNegative=$false;ExpectedOutputSHA256=$null;Rescue=$false;SourceVerified=$false;LiveActionsPerformed=$false;RuntimeApproved=$false;FormalAcceptance=$false;RealGateImplemented=$false}
$row.FrozenSHA256=$FrozenSHA256
$cleanup=[Collections.Generic.List[string]]::new()
function Driver-Error($Record){
    if($null -eq $row.PrimaryError){$row.PrimaryError=$Record.ToString()+"`n"+$Record.ScriptStackTrace}
    elseif($cleanup.Count -lt 8){$cleanup.Add($Record.ToString())}else{$row.Overflow=$true}
}
$process=[Diagnostics.Process]::new();$out=$null;$err=$null;$ot=$null;$et=$null
$timer=[Diagnostics.Stopwatch]::StartNew()
try{
    $out=[IO.File]::Open((Join-Path $packet ($Grant+'.stdout.bin')),[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::Read)
    $err=[IO.File]::Open((Join-Path $packet ($Grant+'.stderr.bin')),[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::Read)
    $process.StartInfo=[Diagnostics.ProcessStartInfo]::new()
    $process.StartInfo.FileName=$engine;$process.StartInfo.UseShellExecute=$false;$process.StartInfo.CreateNoWindow=$true
    $process.StartInfo.RedirectStandardOutput=$true;$process.StartInfo.RedirectStandardError=$true;$process.StartInfo.WorkingDirectory=$repo
    foreach($arg in @('-NoLogo','-NoProfile','-NonInteractive','-File',(Join-Path $PSScriptRoot capture.ps1),'-Grant',$Grant,'-FrozenSHA256',$FrozenSHA256)){$process.StartInfo.ArgumentList.Add($arg)}
    if(-not $process.Start()){throw 'Capture host did not start'}
    $row.Started=$true;$row.PID=$process.Id
    $ot=$process.StandardOutput.BaseStream.CopyToAsync($out)
    $et=$process.StandardError.BaseStream.CopyToAsync($err)
    $max=if($Grant -ceq 'continuation'){150000}else{120000}
    if(-not $process.WaitForExit([Math]::Max(0,($max-[int]$timer.ElapsedMilliseconds)))){$row.TimedOut=$true}
}catch{Driver-Error $_}
finally{
    # Id reflection resolves only this exact object, even if Start threw after
    # association. Never look up processes by PID/name or kill a process tree.
    $associated=$false
    try{
        $property=[Diagnostics.Process].GetProperty('Id',([Reflection.BindingFlags]::Public -bor [Reflection.BindingFlags]::Instance))
        $id=$property.GetValue($process)
        if($id -is [int] -and $id -gt 0){$associated=$true;$row.PID=$id}
    }catch{if($row.Started){Driver-Error $_}}
    if($associated){
        $exited=$false
        try{$exited=$process.HasExited}catch{Driver-Error $_}
        if(-not $exited){$row.TerminationRequested=$true;try{$process.Kill()}catch{Driver-Error $_}}
        try{$row.OwnedExit=$process.WaitForExit(2000);if($row.OwnedExit){$row.ExitCode=$process.ExitCode}}catch{Driver-Error $_}
    }
    if($null -ne $ot -and $null -ne $et){try{$row.StreamsComplete=[Threading.Tasks.Task]::WaitAll([Threading.Tasks.Task[]]@($ot,$et),2000)}catch{Driver-Error $_}}
    foreach($stream in @($out,$err)){if($null -ne $stream){try{$stream.Dispose()}catch{Driver-Error $_}}}
    try{$process.Dispose()}catch{Driver-Error $_}
    $timer.Stop()
}


if($null -eq $row.PrimaryError -and $row.Started -and $row.OwnedExit -and $row.StreamsComplete -and -not $row.TimedOut -and -not $row.TerminationRequested){
    try{
        Driver-CheckHash $manifestPath $FrozenSHA256;Driver-Verify $manifest;$row.PostPinsIntact=$true
        $profile=$oracle.Grants[$Grant]
        $stdoutPath=Join-Path $packet ($Grant+'.stdout.bin');$stderrPath=Join-Path $packet ($Grant+'.stderr.bin');$evidencePath=Join-Path $packet ($Grant+'.evidence.json')
        $stdout=[Text.UTF8Encoding]::new($false,$true).GetString([IO.File]::ReadAllBytes($stdoutPath))
        $stderr=[Text.UTF8Encoding]::new($false,$true).GetString([IO.File]::ReadAllBytes($stderrPath))
        $nl=[Environment]::NewLine
        $expected=($profile.Output -join $nl)+$nl
        $expectedError=if($profile.ErrorOutput.Count -gt 0){($profile.ErrorOutput -join $nl)+$nl}else{''}
        $row.ExpectedOutputSHA256=[Convert]::ToHexString([Security.Cryptography.SHA256]::HashData([Text.Encoding]::UTF8.GetBytes($expected)))
        $row.StdoutSHA256=Driver-Hash $stdoutPath;$row.StderrSHA256=Driver-Hash $stderrPath
        if($stdout -cne $expected -or $stderr -cne $expectedError){throw 'Exact original stdout/stderr mismatch'}
        Driver-Need ((Get-Item -LiteralPath $evidencePath).Length -le 1048576) 'Evidence byte bound'
        $actual=Driver-ReadEvidence $evidencePath
        $row.ChildrenObserved=Driver-Evidence $actual $profile $oracle $Grant $manifest $FrozenSHA256
        $row.EvidenceSHA256=Driver-Hash $evidencePath;$row.EvidenceValid=$true;$row.EvidenceState='RequiredExact'
        Driver-Need ($row.ExitCode -eq $profile.ExitCode) 'Unexpected capture exit'
        $row.Exact=$true
    }catch{Driver-Error $_}
}
$row.SecondaryErrors=$cleanup.ToArray()
Driver-JSON $finalPath $row
[pscustomobject]$row | ConvertTo-Json -Depth 8
if(-not $row.Exact){throw 'Grant nonexact/unknown: freeze all remaining grants and review'}

}catch{
    $actual=$_
    $stopRoot=Join-Path ([IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))) '.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner'
    if((Test-Path -LiteralPath $stopRoot) -and -not (Test-Path -LiteralPath (Join-Path $stopRoot stop.json))){
        Driver-JSON (Join-Path $stopRoot stop.json) ([ordered]@{Action=$Action;Grant=$Grant;Message=$actual.Exception.Message;ErrorId=$actual.FullyQualifiedErrorId;Position=$actual.InvocationInfo.PositionMessage;Stack=$actual.ScriptStackTrace;Remainder='FROZEN'})
    }
    throw
}
