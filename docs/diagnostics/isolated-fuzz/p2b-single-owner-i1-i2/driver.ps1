[CmdletBinding(PositionalBinding=$false)]
param(
    [Parameter(Mandatory)][ValidateSet('Prepare','RunOne')][string]$Action,
    [Parameter(Mandatory)][ValidateSet('seed','writer-green','contract-green','repair-green','controller-rebind','writer-projection-rebind','pin-binding-rebind','extent-contract-rebind','core','lifecycle','continuation')][string]$Grant,
    [ValidatePattern('^[A-F0-9]{64}$')][string]$FrozenSHA256,
    [ValidatePattern('^[A-F0-9]{64}$')][string]$PreviousFrozenSHA256,
    [string]$Epoch='legacy'
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
function Get-ODWriterManifestPath {
    Join-Path $packet 'writer-projection-repair.frozen.json'
}

function Get-ODAttemptSlot([string]$Grant,[string]$Epoch='legacy'){
    if($Epoch -ceq 'extent-contract'){if($Grant -ceq 'contract-green'){return 'extent-contract-green'};throw 'OD: invalid explicit grant'}
    if($Epoch -cne 'legacy'){throw 'OD: invalid explicit epoch'}
    if($Grant -ceq 'writer-green'){return 'writer-projection-green'}
    if($Grant -cin @('contract-green','core','lifecycle','continuation')){return $Grant}
    throw 'OD: error projection recovery mismatch'
}
function Assert-ODOwnerScopeFacts([Collections.IDictionary]$Facts){
    $keys='Schema|ParentSHA256|HistoricalStopSHA256|ApprovalValid|AttemptConsumed|NewStopPresent'
    if($null -eq $Facts -or (@($Facts.Keys) -join '|') -cne $keys){throw 'OD: owner scope recovery mismatch'}
    if($Facts.Schema -isnot [int] -or $Facts.Schema -ne 1){throw 'OD: owner scope recovery mismatch'}
    if($Facts.ParentSHA256 -isnot [string] -or $Facts.ParentSHA256 -cne 'D8D31B8BE9F38A35F12CF5951DC28B3723D759B96AF3BC3B3D8B3D4D95FAC57C' -or $Facts.HistoricalStopSHA256 -isnot [string] -or $Facts.HistoricalStopSHA256 -cne 'ECC2C74E5544CA0EC9A7D2C5A4377DA35F1A352F26CFEADBFEF35124C2A41607'){throw 'OD: owner scope recovery mismatch'}
    foreach($key in @('ApprovalValid','AttemptConsumed','NewStopPresent')){if($Facts[$key] -isnot [bool]){throw 'OD: owner scope recovery mismatch'}}
    if(-not $Facts.ApprovalValid -or $Facts.AttemptConsumed -or $Facts.NewStopPresent){throw 'OD: owner scope recovery mismatch'}
}

function Assert-ODControllerRecoveryFacts([Collections.IDictionary]$Facts){
    $keys='Schema|HistoricalWriterSHA256|HistoricalStopSHA256|ApprovalValid|ForbiddenOutputsPresent|NewStopPresent'
    if($null -eq $Facts -or (@($Facts.Keys) -join '|') -cne $keys){throw 'OD: controller recovery mismatch'}
    if($Facts.Schema -isnot [int] -or $Facts.Schema -ne 1){throw 'OD: controller recovery mismatch'}
    foreach($name in @('HistoricalWriterSHA256','HistoricalStopSHA256')){
        if($Facts[$name] -isnot [string]){throw 'OD: controller recovery mismatch'}
    }
    if($Facts.HistoricalWriterSHA256 -cne '9AC3EAD775DB2BE7D21DE3B1CD3FE46342B1B176BA6B07C03F082D0F63B0C548' -or $Facts.HistoricalStopSHA256 -cne 'F19687B1628DC1C48CC771A07DA4308D48E89B8D51921DDACDB6C4CBFAFDE9D2'){throw 'OD: controller recovery mismatch'}
    foreach($name in @('ApprovalValid','ForbiddenOutputsPresent','NewStopPresent')){
        if($Facts[$name] -isnot [bool]){throw 'OD: controller recovery mismatch'}
    }
    if(-not $Facts.ApprovalValid -or $Facts.ForbiddenOutputsPresent -or $Facts.NewStopPresent){throw 'OD: controller recovery mismatch'}
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
function Driver-UnitEvidence($Evidence,$Oracle,[string]$CurrentGrant,$Manifest,[string]$KnownFreeze,[string]$Epoch='legacy'){
    $expected=$Oracle.Grants[$CurrentGrant]
    $slot=Get-ODAttemptSlot $CurrentGrant $Epoch
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
            $root=Join-Path $packet ('unit/'+$slot+'/'+$v.Boundary+'/'+$v.Case)
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
                $message=('Exception calling "Open" with "4" argument(s): "The file ''{0}'' already exists."' -f (Join-Path $packet ('unit/'+(Get-ODAttemptSlot $CurrentGrant $Epoch)+'/owner-controls/intent-existing.bin')))
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

function Driver-Recorder($Evidence,$Oracle,[string]$CurrentGrant,$Manifest,[string]$Epoch='legacy'){
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
    $probePath=Join-Path $packet ('capture/'+(Get-ODAttemptSlot $CurrentGrant $Epoch)+'/publisher-probe.json')
    $probe=[Text.Encoding]::UTF8.GetBytes($Oracle.Recorder.ProbeText)
    Driver-Need ($r.ProbePath -ceq $probePath -and $r.ProbeBytes -ceq [Convert]::ToBase64String($probe) -and [Linq.Enumerable]::SequenceEqual([byte[]][IO.File]::ReadAllBytes($probePath),[byte[]]$probe)) 'Actual production writer probe bytes'
    $journal=Join-Path $packet ('capture/'+(Get-ODAttemptSlot $CurrentGrant $Epoch)+'/journal.jsonl')
    $expectedJournal='{"Stage":"begin"}'+[char]10+'{"Stage":"tests-complete"}'+[char]10+'{"Stage":"published"}'+[char]10
    Driver-Need ((Get-Item -LiteralPath $journal).Length -le 1048576 -and [Text.UTF8Encoding]::new($false,$true).GetString([IO.File]::ReadAllBytes($journal)) -ceq $expectedJournal) 'Complete independent journal required'
}

function Driver-Evidence($Evidence,$Profile,$Oracle,[string]$CurrentGrant,$Manifest,[string]$KnownFreeze,[string]$Epoch='legacy'){
    $e=$Evidence
    Driver-Keys $e $Oracle.EvidenceRootNames Evidence
    # Validate every returned representation before deciding success/failure.
    # INCOMPLETE with no rows is a known failure representation, never a pass.
    # The actual parent has no historical/unit-context fallback.
    if($CurrentGrant -cin @('core','lifecycle','continuation')){
        $returns=@($e.Observations|Where-Object Name -CEQ ProfileReturn)
        Driver-Need ($returns.Count -eq 1) 'one complete profile representation required'
        $context=New-ODRuntimeContext $repo $Profile.Profile $Manifest $KnownFreeze $Epoch
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
    Driver-Recorder $e $Oracle $CurrentGrant $Manifest $Epoch
    if($CurrentGrant -cin @('writer-green','contract-green')){$unit=$e.Clone();$unit.Observations=@($e.Observations | Where-Object {$_.Name -cne 'Recorder'});Driver-UnitEvidence $unit $Oracle $CurrentGrant $Manifest $KnownFreeze $Epoch;return 0}
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
# The historical prelaunch stop is retained, never removed or overwritten.
function Driver-ControllerRegression([string]$Phase,$Binding){
    $red=$Phase -ceq 'Red'
    $name=if($red){'controller-regression-red.final.json'}else{'controller-regression-parameter-green.final.json'}
    $receipt=Get-Content -LiteralPath (Join-Path $packet $name) -Raw|ConvertFrom-Json -AsHashtable
    Driver-Keys $receipt @('Schema','Phase','TestSHA256','InputSHA256','SourceSHA256','PASS','FAIL','NotRun','ExitCode','Groups','RootError') ControllerRegression
    $expectedTest=if($red){$Binding.OriginalTestSHA256}else{$Binding.TestSHA256}
    $expectedInput=if($red){$Binding.OriginalInputSHA256}else{'0E5D77EDB197CE1B067C617E67B114174DFF1CEA95494B3265A4FE19594F2AC6'}
    foreach($key in @('Schema','PASS','FAIL','NotRun','ExitCode')){Driver-Need ($receipt[$key] -is [long]) 'Regression integer types'}
    Driver-Need ($receipt.Schema -eq 1 -and $receipt.Phase -ceq $Phase -and $receipt.TestSHA256 -ceq $expectedTest -and $receipt.InputSHA256 -ceq $expectedInput -and $receipt.PASS -eq $(if($red){0}else{13}) -and $receipt.FAIL -eq $(if($red){1}else{0}) -and $receipt.NotRun -eq $(if($red){12}else{0}) -and $receipt.ExitCode -eq $(if($red){1}else{0})) 'Exact regression disposition'
    $names=@('ProductionManifestDictionary','FiveFalseFlags','TrueFlagRejected','StringFlagRejected','MissingFlagRejected','NullFlagRejected','RecoveryExact','MissingApproval','ChangedStop','WrongParent','ExistingAttempt','NewStop','WriterBindings')
    Driver-Need ($receipt.Groups -is [array] -and $receipt.Groups.Count -eq 13) 'Thirteen regression groups'
    for($i=0;$i -lt 13;$i++){
        Driver-Keys $receipt.Groups[$i] @('Name','Status') RegressionGroup
        $status=if($red){if($i -eq 0){'FAIL'}else{'NotRun'}}else{'PASS'}
        Driver-Need ($receipt.Groups[$i].Name -ceq $names[$i] -and $receipt.Groups[$i].Status -ceq $status) 'Regression order/status'
    }
    Driver-Keys $receipt.SourceSHA256 @($Binding.BaselineSources.Keys) RegressionSources
    foreach($relative in $Binding.BaselineSources.Keys){
        if($red){Driver-Need ($receipt.SourceSHA256[$relative] -ceq $Binding.BaselineSources[$relative]) 'Original RED source epoch'}
        else{Driver-CheckHash (Join-Path (Join-Path $packet snapshots/controller-rebind) $relative) $receipt.SourceSHA256[$relative]}
        Driver-CheckHash (Join-Path (Join-Path $packet snapshots/writer-green) $relative) $Binding.BaselineSources[$relative]
    }
    if($red){
        Driver-Keys $receipt.RootError @('ErrorId','Message','Position','Stack') ActualRegressionCause
        $stop=Get-Content -LiteralPath (Join-Path $packet stop.json) -Raw|ConvertFrom-Json -AsHashtable
        Driver-Need ($receipt.RootError.ErrorId -ceq 'CannotIndex' -and $receipt.RootError.Message -ceq $stop.Message -and $receipt.RootError.Position -is [string] -and $receipt.RootError.Stack -is [string]) 'Exact original regression cause'
    }else{Driver-Need ($null -eq $receipt.RootError) 'No GREEN cause'}
    return $receipt
}
function Driver-ControllerMap($Info){
    return [ordered]@{Schema=1;HistoricalWriterSHA256=$Info.Binding.HistoricalWriterSHA256;HistoricalStopSHA256=$Info.Binding.HistoricalStopSHA256;ApprovalSHA256=$Info.Binding.ApprovalSHA256;AmendmentPlanSHA256=$Info.Binding.PlanSHA256;AuditSHA256=$Info.Binding.AuditSHA256;RegressionTestSHA256=$Info.Binding.TestSHA256;RegressionRedSHA256=$Info.RedSHA256;RegressionGreenSHA256=$Info.GreenSHA256;ParameterApprovalSHA256=$Info.Binding.ParameterApprovalSHA256;HistoricalAuxiliaryStopSHA256=$Info.Binding.HistoricalAuxiliaryStopSHA256;OriginalRegressionTestSHA256=$Info.Binding.OriginalTestSHA256;OriginalRegressionInputSHA256=$Info.Binding.OriginalInputSHA256}
}
function Driver-ControllerGate {
    $inputPath=Join-Path $packet controller-regression-parameter.input.json
    Driver-CheckHash $inputPath '0E5D77EDB197CE1B067C617E67B114174DFF1CEA95494B3265A4FE19594F2AC6'
    $binding=Get-Content -LiteralPath $inputPath -Raw|ConvertFrom-Json -AsHashtable
    Driver-CheckHash (Join-Path $packet controller-regression.input.json) '954E1DDAB0DA47A937D037B2D251D5630AD520DC56F7E1E1C16246B5C4579344'
    Driver-CheckHash (Join-Path $packet controller-repair.stop.json) '16A1CE2B13BC49389E981EED646A7AFF78219BB85760E5B63362FDE839774027'
    $parameterApprovalPath=Join-Path $packet controller-parameter.approval.json
    Driver-CheckHash $parameterApprovalPath $binding.ParameterApprovalSHA256
    $parameterApproval=Get-Content -LiteralPath $parameterApprovalPath -Raw|ConvertFrom-Json -AsHashtable
    Driver-CheckHash (Join-Path $packet writer-green.frozen.json) '9AC3EAD775DB2BE7D21DE3B1CD3FE46342B1B176BA6B07C03F082D0F63B0C548'
    Driver-CheckHash (Join-Path $packet stop.json) 'F19687B1628DC1C48CC771A07DA4308D48E89B8D51921DDACDB6C4CBFAFDE9D2'
    Driver-CheckHash (Join-Path $packet seed.frozen.json) '4E4342C389C507F6148AE029140814FBCB4046AD62D1EFE3BFD451863C9DC1BA'
    Driver-CheckHash (Join-Path $repo 'docs/superpowers/plans/2026-11-01-offline-controller-recovery.md') 'EC4F8E00B0B46BCB2086263A06F2AC58DCF6A5ECF3B3F47E43F9CF33CA2A5691'
    Driver-CheckHash (Join-Path $repo 'docs/roadmap/2026-11-01-offline-controller-readonly-audit.md') 'EC4FF0038A10BF613A7F99D264953F81B6B4BB20B0B872023C74281E2537C2EB'
    Driver-CheckHash (Join-Path $PSScriptRoot controller-recovery.tests.ps1) '522452799A8DCE5D8ECCDFBE0FF5852D56E3C564FA6C462843145A5E450F5DB5'
    Driver-CheckHash (Join-Path $PSScriptRoot controller-recovery-parameter.tests.ps1) '45571C4256BDDEFD86822714949FEA92E0F989EA040DFB536C78341F1E3DA97E'
    $approvalPath=Join-Path $packet controller-repair.approval.json
    Driver-CheckHash $approvalPath $binding.ApprovalSHA256
    $approval=Get-Content -LiteralPath $approvalPath -Raw|ConvertFrom-Json -AsHashtable
    $approved=$approval.Schema -eq 1 -and $approval.UserReply -ceq '是' -and $approval.Method -ceq 'Native' -and $approval.PlanSHA256 -ceq $binding.PlanSHA256 -and $approval.AuditSHA256 -ceq $binding.AuditSHA256 -and $approval.HistoricalWriterSHA256 -ceq $binding.HistoricalWriterSHA256 -and $approval.HistoricalStopSHA256 -ceq $binding.HistoricalStopSHA256 -and (Driver-EqualJSON $approval.ExistingSourceWindow @('driver.ps1','record-boundary.tests.ps1','source-model.ps1')) -and $approval.AuxiliaryRegression -ceq 'controller-recovery.tests.ps1'
    foreach($pair in @(@('AdditionalSourceSnapshots',31),@('SourceSnapshotCeiling',124),@('RegressionComputations',2),@('CaptureHosts',5),@('Targets',7),@('Compilations',1),@('ConditionalRuntimeReviews',1))){$approved=$approved -and $approval[$pair[0]] -is [long] -and $approval[$pair[0]] -eq $pair[1]}
    $approved=$approved -and $approval.RepeatOldRED -is [bool] -and -not $approval.RepeatOldRED -and $approval.PreserveHistoricalEvidence -is [bool] -and $approval.PreserveHistoricalEvidence
    $approved=$approved -and $parameterApproval.Schema -eq 1 -and $parameterApproval.UserReply -ceq '是' -and $parameterApproval.BaseCommit -ceq 'f7c4160f74a86828ad8675b24ce06e9ee8b946e3' -and $parameterApproval.HistoricalAuxiliaryStopSHA256 -ceq $binding.HistoricalAuxiliaryStopSHA256 -and $parameterApproval.OriginalTestSHA256 -ceq $binding.OriginalTestSHA256 -and $parameterApproval.OriginalInputSHA256 -ceq $binding.OriginalInputSHA256 -and $parameterApproval.CorrectedAuxiliary -ceq 'controller-recovery-parameter.tests.ps1'
    $approved=$approved -and $parameterApproval.SupplementalGreenComputations -is [long] -and $parameterApproval.SupplementalGreenComputations -eq 1 -and $parameterApproval.AdditionalSnapshotBudget -is [long] -and $parameterApproval.AdditionalSnapshotBudget -eq 0 -and $parameterApproval.RepeatRED -is [bool] -and -not $parameterApproval.RepeatRED -and $parameterApproval.PreserveAllOriginals -is [bool] -and $parameterApproval.PreserveAllOriginals
    foreach($pair in @(@('CaptureHosts',5),@('Targets',7),@('Compiles',1),@('ConditionalRuntimeReviews',1))){$approved=$approved -and $parameterApproval[$pair[0]] -is [long] -and $parameterApproval[$pair[0]] -eq $pair[1]}
    $null=Driver-ControllerRegression Red $binding
    $null=Driver-ControllerRegression Green $binding
    $info=[pscustomobject]@{Binding=$binding;RedSHA256=(Driver-Hash (Join-Path $packet controller-regression-red.final.json));GreenSHA256=(Driver-Hash (Join-Path $packet controller-regression-parameter-green.final.json))}
    # Historical controller gate: no current-source regrading or old-attempt replay.
    $epochPath=Join-Path $packet writer-controller-repair.frozen.json
    $epoch=Get-Content -LiteralPath $epochPath -Raw|ConvertFrom-Json -AsHashtable
    Driver-Need ($epoch.Schema -eq 2 -and $epoch.Phase -ceq 'writer-green' -and $epoch.PreviousFrozenSHA256 -ceq $binding.HistoricalWriterSHA256 -and $epoch.PinCount -eq 2215 -and $epoch.SourceCount -eq 31 -and (Driver-EqualJSON $epoch.SourcePaths $inputPaths) -and (Driver-EqualJSON $epoch.ControllerRecovery (Driver-ControllerMap $info))) 'Exact controller source epoch'
    $intent=Get-Content -LiteralPath (Join-Path $packet controller-repair.intent.json) -Raw|ConvertFrom-Json -AsHashtable
    Driver-Need ($intent.Schema -eq 1 -and $intent.AttemptConsumed -is [bool] -and $intent.AttemptConsumed -and $intent.HistoricalWriterSHA256 -ceq $binding.HistoricalWriterSHA256 -and $intent.HistoricalStopSHA256 -ceq $binding.HistoricalStopSHA256 -and $intent.ApprovalSHA256 -ceq $binding.ApprovalSHA256 -and $intent.PlanSHA256 -ceq $binding.PlanSHA256 -and $intent.TestSHA256 -ceq $binding.TestSHA256 -and $intent.ParameterApprovalSHA256 -ceq $binding.ParameterApprovalSHA256) 'Unique actual source recovery intent'
    return $info
}
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
function Driver-RequireFinal([string]$Name,[string]$Epoch='legacy'){
    $slot=Get-ODAttemptSlot $Name $Epoch
    $record=Get-Content -LiteralPath (Join-Path $packet ($slot+'.final.json')) -Raw|ConvertFrom-Json -AsHashtable
    foreach($key in @('Started','OwnedExit','StreamsComplete','PostPinsIntact','Exact','EvidenceValid')){Driver-Need ($record[$key] -is [bool] -and $record[$key]) 'Prior actual host must be exact'}
    foreach($key in @('TimedOut','TerminationRequested','Overflow','Rescue')){Driver-Need ($record[$key] -is [bool] -and -not $record[$key]) 'Prior host cannot require rescue'}
    Driver-Need ($record.Grant -ceq $Name -and $record.AttemptSlot -ceq $slot -and $record.ExitCode -eq 0 -and $null -eq $record.PrimaryError -and $record.SecondaryErrors -is [array] -and $record.SecondaryErrors.Count -eq 0) 'Prior host disposition'
    $boundPath=if($Name -ceq 'writer-green'){Get-ODWriterManifestPath}else{Join-Path $packet (Get-ODEpochManifestName $Epoch)}
    Driver-Need ($record.FrozenSHA256 -ceq (Driver-Hash $boundPath)) 'Prior actual host source epoch'
    Driver-FalseFlags $record
    Driver-Need ($record.EvidenceSHA256 -ceq (Driver-Hash (Join-Path $packet ($slot+'.evidence.json'))) -and $record.StdoutSHA256 -ceq (Driver-Hash (Join-Path $packet ($slot+'.stdout.bin'))) -and $record.StderrSHA256 -ceq (Driver-Hash (Join-Path $packet ($slot+'.stderr.bin')))) 'Prior original stream/evidence identity'
}

function Driver-ScopeRegression([string]$Phase,$Binding){
    $red=$Phase -ceq 'Red'
    $prefix=if($red){'owner-scope-signature-red'}else{'owner-scope-verifier-scope-green'}
    $path=Join-Path $packet ($prefix+'.final.json')
    $r=Get-Content -LiteralPath $path -Raw|ConvertFrom-Json -AsHashtable
    Driver-Keys $r @('Schema','Phase','TestSHA256','InputSHA256','SourceSHA256','Groups','PASS','FAIL','NotRun','ExitCode','RootError') ScopeRegression
    foreach($key in @('Schema','PASS','FAIL','NotRun','ExitCode')){Driver-Need ($r[$key] -is [long]) 'Scope regression integer types'}
    Driver-Need ($r.Schema -eq 1 -and $r.Phase -ceq $Phase -and $r.TestSHA256 -ceq $Binding.TestSHA256 -and $r.InputSHA256 -ceq $(if($red){'511274F055A0107E4A76899F650FB8151F4AC47B344D7824EDC95937D69AE792'}else{'F08D1A3AEB9E8840024C4F2C80711DE7D411C90B22C08BD265A3FD1EDC28A7F0'}) -and $r.PASS -eq $(if($red){0}else{9}) -and $r.FAIL -eq $(if($red){1}else{0}) -and $r.NotRun -eq $(if($red){8}else{0}) -and $r.ExitCode -eq $(if($red){1}else{0})) 'Exact scope regression disposition'
    $names=@('FourFunctionsPersist','DefinitionsExact','ModuleIsolation','AttemptBindings','ConsumedAttempt','HistoricalEpoch','RecoveryGate','NewEpochBindings','OriginalContractsUnchanged')
    Driver-Need ($r.Groups -is [array] -and $r.Groups.Count -eq 9) 'Nine scope groups'
    for($i=0;$i -lt 9;$i++){
        Driver-Keys $r.Groups[$i] @('Name','Status') ScopeGroup
        $status=if($red){if($i -eq 0){'FAIL'}else{'NotRun'}}else{'PASS'}
        Driver-Need ($r.Groups[$i].Name -ceq $names[$i] -and $r.Groups[$i].Status -ceq $status) 'Scope group order/disposition'
    }
    Driver-Keys $r.SourceSHA256 @($Binding.BaselineSources.Keys) ScopeSources
    foreach($relative in $Binding.BaselineSources.Keys){
        Driver-CheckHash (Join-Path (Join-Path $packet snapshots/controller-rebind) $relative) $Binding.BaselineSources[$relative]
        if($red){Driver-Need ($r.SourceSHA256[$relative] -ceq $Binding.BaselineSources[$relative]) 'Scope RED parent epoch'}
        else{Driver-CheckHash (Join-Path $repo $relative) $r.SourceSHA256[$relative]}
    }
    $intent=Get-Content -LiteralPath (Join-Path $packet ($prefix+'.intent.json')) -Raw|ConvertFrom-Json -AsHashtable
    Driver-Need ($intent.Schema -eq 1 -and $intent.Phase -ceq $Phase -and $intent.AttemptConsumed -is [bool] -and $intent.AttemptConsumed -and $intent.TestSHA256 -ceq $r.TestSHA256 -and $intent.InputSHA256 -ceq $r.InputSHA256) 'Unique scope phase intent'
    if(-not $red){Driver-Need ($intent.Computation -is [long] -and $intent.Computation -eq 6 -and $intent.ComputationCeiling -is [long] -and $intent.ComputationCeiling -eq 6) 'Final bounded auxiliary computation'}
    if($red){
        Driver-Keys $r.RootError @('Type','ErrorId','Message','Position','Stack') ActualScopeCause
        Driver-Need ($r.RootError.Type -ceq 'System.Management.Automation.CommandNotFoundException' -and $r.RootError.ErrorId -ceq 'CommandNotFoundException,Microsoft.PowerShell.Commands.GetCommandCommand' -and $r.RootError.Message.Contains('Invoke-ODProfile') -and $r.RootError.Position -is [string] -and $r.RootError.Stack -is [string]) 'Exact query-boundary scope cause'
    }else{Driver-Need ($null -eq $r.RootError) 'GREEN no scope cause'}
    return Driver-Hash $path
}
function Driver-ScopeMap($Info){
    return [ordered]@{
        Schema=1;ParentSHA256='D8D31B8BE9F38A35F12CF5951DC28B3723D759B96AF3BC3B3D8B3D4D95FAC57C'
        HistoricalStopSHA256='ECC2C74E5544CA0EC9A7D2C5A4377DA35F1A352F26CFEADBFEF35124C2A41607'
        OriginalScopePlanSHA256=$Info.Original.ApprovedPlanSHA256;OriginalScopeDesignSHA256=$Info.Original.ApprovedDesignSHA256
        OriginalScopeApprovalSHA256=$Info.Binding.HistoricalFiles['owner-scope.approval.json']
        OriginalTestSHA256='BDA424DCC24805F6FFDDEB30DDB67F502CDE7FD34B4B32E1E760D156507C4DCE'
        OriginalInputSHA256=$Info.Binding.HistoricalFiles['owner-scope.input.json']
        HistoricalArtifacts=$Info.Binding.HistoricalFiles
        SignatureRecovery=[ordered]@{Schema=1;HistoricalAuxiliaryStopSHA256='D82E7A269C8982EB6AA7BDECCC779D8A3E674CC3F01EBE6484A4145B35CF3236';SupplementDesignSHA256=$Info.Binding.DesignSHA256;SupplementPlanSHA256=$Info.Binding.PlanSHA256;SupplementApprovalSHA256=$Info.Binding.ApprovalSHA256;TestSHA256=$Info.Binding.TestSHA256;InputSHA256='511274F055A0107E4A76899F650FB8151F4AC47B344D7824EDC95937D69AE792';RedFinalSHA256=$Info.RedSHA256;FailedGreenDiagnosticSHA256='53D5EADE9E505069DC92C0561F06F1D269AEC0B8A48A6821DC7AA60016FAEBDF'}
        VerifierRecovery=[ordered]@{Schema=1;HistoricalSignatureStopSHA256='1D8844C21D1F0559CCD705D4394519E30D31BC72524F9671D821F46FF4118F38';PlanSHA256=$Info.Verifier.Binding.PlanSHA256;DesignSHA256=$Info.Verifier.Binding.DesignSHA256;ApprovalSHA256=$Info.Verifier.Binding.HistoricalFiles['owner-scope-verifier.approval.json'];RegressionTestSHA256=$Info.Verifier.Binding.RegressionTestSHA256;RegressionInputSHA256=$Info.Verifier.Binding.RegressionInputSHA256;TestSHA256=$Info.Verifier.Binding.TestSHA256;InputSHA256='F08D1A3AEB9E8840024C4F2C80711DE7D411C90B22C08BD265A3FD1EDC28A7F0';VerifierRedFinalSHA256=$Info.Verifier.RedSHA256;VerifierGreenFinalSHA256=$Info.Verifier.GreenSHA256;ScopeGreenFinalSHA256=$Info.GreenSHA256;HistoricalFailedGreenToolSHA256='F7E6F4F6B95D984FC88FCF7A8BE3949593913EB4CAD45F919F2F6E04E825B204'}
        Budget=[ordered]@{HistoricalHostsConsumed=1;HistoricalCompilesConsumed=1;TotalHostCeiling=6;TotalCompilationCeiling=2;ScopeComputationCeiling=6;SourceSnapshotCeiling=155;Targets=7;ConditionalReviews=1}
    }
}
function Driver-VerifierRegression([string]$Phase,$Binding){
    $red=$Phase -ceq 'Red';$prefix='owner-scope-verifier-'+$Phase.ToLowerInvariant()
    $path=Join-Path $packet ($prefix+'.final.json')
    $r=Get-Content -LiteralPath $path -Raw|ConvertFrom-Json -AsHashtable
    Driver-Keys $r @('Schema','Phase','TestSHA256','InputSHA256','CodeSHA256','FixtureSHA256','Groups','PASS','FAIL','NotRun','ExitCode','RootError') VerifierRegression
    foreach($k in @('Schema','PASS','FAIL','NotRun','ExitCode')){Driver-Need ($r[$k] -is [long]) 'Verifier integer kinds'}
    Driver-Need ($r.Schema -eq 1 -and $r.Phase -ceq $Phase -and $r.TestSHA256 -ceq $Binding.RegressionTestSHA256 -and $r.InputSHA256 -ceq $Binding.RegressionInputSHA256 -and $r.CodeSHA256 -ceq $(if($red){$Binding.HistoricalTestSHA256}else{$Binding.TestSHA256}) -and $r.PASS -eq $(if($red){0}else{8}) -and $r.FAIL -eq $(if($red){1}else{0}) -and $r.NotRun -eq $(if($red){7}else{0}) -and $r.ExitCode -eq $(if($red){1}else{0})) 'Exact verifier disposition'
    Driver-Need (Driver-EqualJSON $r.FixtureSHA256 $Binding.FixtureSHA256) 'Six actual verifier fixture identities'
    $names=@('SameDefinitionsAccepted','ChangedBodyRejected','ChangedInlineParameterRejected','UnexpectedAstKindRejected','UnexpectedNameRejected','FailurePreservesOriginal','CompletedGroupsPreserved','LaterCasesNotExecuted')
    Driver-Need ($r.Groups -is [array] -and $r.Groups.Count -eq 8) 'Eight verifier groups'
    for($i=0;$i -lt 8;$i++){
        Driver-Keys $r.Groups[$i] @('Name','Status') VerifierGroup
        Driver-Need ($r.Groups[$i].Name -ceq $names[$i] -and $r.Groups[$i].Status -ceq $(if($red){if($i -eq 0){'FAIL'}else{'NotRun'}}else{'PASS'})) 'Verifier group order/disposition'
    }
    if($red){
        Driver-Keys $r.RootError @('Type','ErrorId','Message','Position','Stack','Comparison') ActualVerifierCause
        Driver-Need ($r.RootError.Type -ceq 'System.Management.Automation.RuntimeException' -and $r.RootError.ErrorId -ceq 'Owner scope regression: actual body unchanged Capture-Value' -and $r.RootError.Message -ceq $r.RootError.ErrorId -and $r.RootError.Position -is [string] -and $r.RootError.Stack -is [string]) 'Actual old checker rejection'
        Driver-Keys $r.RootError.Comparison @('ExpectedNodeType','LoadedNodeType','ExpectedText','LoadedText','OldExpectedOperand','OldLoadedOperand') ActualComparison
        $c=$r.RootError.Comparison
        Driver-Need ($c.ExpectedNodeType -ceq 'System.Management.Automation.Language.FunctionDefinitionAst' -and $c.LoadedNodeType -ceq $c.ExpectedNodeType -and $c.ExpectedText -is [string] -and $c.LoadedText -is [string] -and $c.OldExpectedOperand -is [string] -and $c.OldLoadedOperand -is [string] -and $c.OldExpectedOperand -cne $c.OldLoadedOperand) 'Actual unequal old operands'
    }else{Driver-Need ($null -eq $r.RootError) 'No verifier GREEN cause'}
    $intent=Get-Content -LiteralPath (Join-Path $packet ($prefix+'.intent.json')) -Raw|ConvertFrom-Json -AsHashtable
    Driver-Need ($intent.Schema -eq 1 -and $intent.Phase -ceq $Phase -and $intent.AttemptConsumed -is [bool] -and $intent.AttemptConsumed -and $intent.Computation -eq $(if($red){4}else{5}) -and $intent.ComputationCeiling -eq 6 -and $intent.TestSHA256 -ceq $r.TestSHA256 -and $intent.InputSHA256 -ceq $r.InputSHA256) 'Unique verifier phase intent'
    if(-not $red){Driver-Need ($intent.CodeSHA256 -ceq $r.CodeSHA256 -and $intent.FullInputSHA256 -ceq (Driver-Hash (Join-Path $packet owner-scope-verifier-scope.input.json))) 'Verifier implementation/input intent'}
    return Driver-Hash $path
}
function Driver-VerifierGate {
    $path=Join-Path $packet owner-scope-verifier-scope.input.json
    Driver-CheckHash $path 'F08D1A3AEB9E8840024C4F2C80711DE7D411C90B22C08BD265A3FD1EDC28A7F0'
    $b=Get-Content -LiteralPath $path -Raw|ConvertFrom-Json -AsHashtable
    Driver-Need ($b.Schema -eq 1 -and $b.AdditionalComputations -eq 3 -and $b.ComputationCeiling -eq 6 -and $b.TestSHA256 -ceq '52B237AF4BF7DB0AD0D45919B41860C8F39831E2B26E69BC840BE341E096FF4E' -and $b.RegressionTestSHA256 -ceq 'BD3DF80E75C62B7924BBC4AB399493AE9AD1F102117B3061D4B7DE0775AFD926' -and $b.RegressionInputSHA256 -ceq 'C76019BD52329DDDF5DEA2117EBAD44B77AF6E15E72B515CD5707F63A2561548' -and $b.HistoricalTestSHA256 -ceq 'FA11EF9E1B3B3AAA411C3B848525243094FAC318EE8AD38F1F6B3C63581C3513' -and $b.HistoricalInputSHA256 -ceq '511274F055A0107E4A76899F650FB8151F4AC47B344D7824EDC95937D69AE792' -and $b.HistoricalRedFinalSHA256 -ceq '307D8BC0C2D5E9C12F34EAA99DEE81A9FF0FED7190891660062425BC80E79C1A' -and $b.HistoricalFailedGreenSHA256 -ceq '53D5EADE9E505069DC92C0561F06F1D269AEC0B8A48A6821DC7AA60016FAEBDF') 'Exact verifier frozen binding'
    foreach($n in $b.HistoricalFiles.Keys){Driver-CheckHash (Join-Path $packet $n) $b.HistoricalFiles[$n]}
    Driver-CheckHash (Join-Path $packet owner-scope-signature.stop.json) '1D8844C21D1F0559CCD705D4394519E30D31BC72524F9671D821F46FF4118F38'
    Driver-CheckHash (Join-Path $PSScriptRoot owner-scope-verifier.tests.ps1) $b.TestSHA256
    Driver-CheckHash (Join-Path $PSScriptRoot owner-scope-verifier-regression.tests.ps1) $b.RegressionTestSHA256
    Driver-CheckHash (Join-Path $packet owner-scope-verifier-regression.input.json) $b.RegressionInputSHA256
    Driver-CheckHash (Join-Path $repo 'docs/superpowers/plans/2026-11-01-offline-owner-scope-verifier.md') $b.PlanSHA256
    Driver-CheckHash (Join-Path $repo 'docs/superpowers/specs/2026-11-01-offline-owner-scope-verifier-design.md') $b.DesignSHA256
    Driver-Need ($b.FixtureSHA256.Count -eq 6) 'Exact six verifier fixtures'
    foreach($n in $b.FixtureSHA256.Keys){Driver-CheckHash (Join-Path $repo $n) $b.FixtureSHA256[$n]}
    $a=Get-Content -LiteralPath (Join-Path $packet owner-scope-verifier.approval.json) -Raw|ConvertFrom-Json -AsHashtable
    Driver-Need ($a.Schema -eq 1 -and $a.Approved -is [bool] -and $a.Approved -and $a.UserReply -ceq '是' -and $a.Method -ceq 'Native' -and $a.BaseCommit -ceq '9128d0cd6dc6ca4c3a359de38b2887f49d773f93' -and $a.PlanSHA256 -ceq $b.PlanSHA256 -and $a.DesignSHA256 -ceq $b.DesignSHA256 -and (Driver-EqualJSON $a.FixtureSHA256 $b.FixtureSHA256) -and (Driver-EqualJSON $a.HistoricalScopeBaseline $b.BaselineSources)) 'Exact verifier approval'
    foreach($n in $a.HistoricalFiles.Keys){Driver-Need ($b.HistoricalFiles[$n] -ceq $a.HistoricalFiles[$n]) 'Historical approval binding'}
    foreach($n in $a.CandidateSources.Keys){if($n -cne 'docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/driver.ps1'){Driver-CheckHash (Join-Path $repo $n) $a.CandidateSources[$n]}}
    foreach($pair in @(@('AdditionalComputations',3),@('ComputationCeiling',6),@('ModulesPerComputation',2),@('AdditionalHosts',0),@('AdditionalCompiles',0),@('AdditionalTargets',0),@('AdditionalSnapshots',0),@('AdditionalReviewers',0))){Driver-Need ($a[$pair[0]] -is [long] -and $a[$pair[0]] -eq $pair[1]) 'Exact verifier finite allowance'}
    return [pscustomobject]@{Binding=$b;Approval=$a;RedSHA256=(Driver-VerifierRegression Red $b);GreenSHA256=(Driver-VerifierRegression Green $b)}
}

function Driver-OwnerScopeGate {
    $newStop=Test-Path -LiteralPath (Join-Path $packet owner-scope-verifier.stop.json)
    Driver-Need (-not $newStop) 'Scope verifier recovery stopped; no retry'
    $inputPath=Join-Path $packet owner-scope-signature.input.json
    Driver-CheckHash $inputPath '511274F055A0107E4A76899F650FB8151F4AC47B344D7824EDC95937D69AE792'
    $binding=Get-Content -LiteralPath $inputPath -Raw|ConvertFrom-Json -AsHashtable
    foreach($name in $binding.HistoricalFiles.Keys){Driver-CheckHash (Join-Path $packet $name) $binding.HistoricalFiles[$name]}
    Driver-CheckHash (Join-Path $packet writer-scope.stop.json) 'D82E7A269C8982EB6AA7BDECCC779D8A3E674CC3F01EBE6484A4145B35CF3236'
    Driver-CheckHash (Join-Path $packet controller-parameter.stop.json) 'ECC2C74E5544CA0EC9A7D2C5A4377DA35F1A352F26CFEADBFEF35124C2A41607'
    Driver-CheckHash (Join-Path $PSScriptRoot owner-scope.tests.ps1) 'BDA424DCC24805F6FFDDEB30DDB67F502CDE7FD34B4B32E1E760D156507C4DCE'
    Driver-CheckHash (Join-Path $PSScriptRoot owner-scope-signature.tests.ps1) 'FA11EF9E1B3B3AAA411C3B848525243094FAC318EE8AD38F1F6B3C63581C3513'
    Driver-CheckHash (Join-Path $repo 'docs/superpowers/plans/2026-11-01-offline-owner-scope-signature.md') $binding.PlanSHA256
    Driver-CheckHash (Join-Path $repo 'docs/superpowers/specs/2026-11-01-offline-owner-scope-signature-design.md') $binding.DesignSHA256
    Driver-CheckHash (Join-Path $packet owner-scope-signature.approval.json) $binding.ApprovalSHA256
    $approval=Get-Content -LiteralPath (Join-Path $packet owner-scope-signature.approval.json) -Raw|ConvertFrom-Json -AsHashtable
    $original=Get-Content -LiteralPath (Join-Path $packet owner-scope.approval.json) -Raw|ConvertFrom-Json -AsHashtable
    Driver-CheckHash (Join-Path $repo 'docs/superpowers/plans/2026-11-01-offline-owner-scope-repair.md') $original.ApprovedPlanSHA256
    Driver-CheckHash (Join-Path $repo 'docs/superpowers/specs/2026-11-01-offline-owner-scope-repair-design.md') $original.ApprovedDesignSHA256
    $approved=$binding.Schema -eq 1 -and $binding.TestSHA256 -ceq 'FA11EF9E1B3B3AAA411C3B848525243094FAC318EE8AD38F1F6B3C63581C3513' -and $approval.Schema -eq 1 -and $approval.Approved -is [bool] -and $approval.Approved -and $approval.UserReply -ceq '是' -and $approval.BaseCommit -ceq '576d2c25fcc3740caef01b3db84ffd01ef0147d2' -and $approval.PlanSHA256 -ceq $binding.PlanSHA256 -and $approval.DesignSHA256 -ceq $binding.DesignSHA256 -and (Driver-EqualJSON $approval.HistoricalFiles $binding.HistoricalFiles) -and (Driver-EqualJSON $approval.BaselineSources $binding.BaselineSources)
    foreach($pair in @(@('AdditionalRedComputations',1),@('ReleasedExistingGreenComputations',1),@('AdditionalHosts',0),@('AdditionalCompiles',0),@('AdditionalTargets',0),@('AdditionalSnapshots',0),@('AdditionalReviewers',0))){$approved=$approved -and $approval[$pair[0]] -is [long] -and $approval[$pair[0]] -eq $pair[1]}
    $approved=$approved -and $original.Schema -eq 1 -and $original.Approved -is [bool] -and $original.Approved -and $original.UserReply -ceq '是' -and $original.Method -ceq 'Native' -and $original.SnapshotCeiling -eq 155 -and $original.AdditionalBudget.ScopeComputations -eq 2 -and $original.AdditionalBudget.CaptureHosts -eq 1 -and $original.AdditionalBudget.FixtureCompilations -eq 1 -and $original.AdditionalBudget.SourceSnapshots -eq 31 -and $original.AdditionalBudget.Targets -eq 0 -and $original.AdditionalBudget.Reviewers -eq 0
    $controller=Driver-ControllerGate
    $verifier=Driver-VerifierGate
    Driver-Need (Driver-EqualJSON $binding.BaselineSources $verifier.Binding.BaselineSources) 'Historical and new source windows agree'
    $red=Driver-ScopeRegression Red $binding;$green=Driver-ScopeRegression Green $verifier.Binding
    $info=[pscustomobject]@{Binding=$binding;Original=$original;Controller=$controller;Verifier=$verifier;RedSHA256=$red;GreenSHA256=$green}
    $grants=@('writer-green','contract-green','core','lifecycle','continuation')
    $start=if($Action -ceq 'Prepare' -and $Grant -ceq 'writer-scope-rebind'){0}elseif($Action -ceq 'Prepare' -and $Grant -ceq 'repair-green'){1}elseif($Action -ceq 'RunOne'){[array]::IndexOf($grants,$Grant)}else{-1}
    Driver-Need ($start -ge 0) 'Only approved current scope stages'
    $consumed=$false
    for($i=$start;$i -lt $grants.Count;$i++){
        $slot=Get-ODAttemptSlot $grants[$i]
        foreach($suffix in @('intent.json','final.json','stdout.bin','stderr.bin','evidence.json')){if(Test-Path -LiteralPath (Join-Path $packet ($slot+'.'+$suffix))){$consumed=$true}}
        foreach($relative in @(('capture/'+$slot),('unit/'+$slot))){if(Test-Path -LiteralPath (Join-Path $packet $relative)){$consumed=$true}}
    }
    if($start -le 1 -and (Test-Path -LiteralPath (Join-Path $packet profiles))){$consumed=$true}
    if($start -eq 0 -or ($Action -ceq 'Prepare' -and $Grant -ceq 'repair-green')){foreach($relative in @('repair-green.frozen.json','snapshots/repair-green')){if(Test-Path -LiteralPath (Join-Path $packet $relative)){$consumed=$true}}}
    if($Action -ceq 'Prepare' -and $Grant -ceq 'writer-scope-rebind'){
        foreach($relative in @('writer-scope-repair.intent.json','writer-scope-repair.frozen.json','snapshots/writer-scope-rebind')){if(Test-Path -LiteralPath (Join-Path $packet $relative)){$consumed=$true}}
    }else{
        $epoch=Get-Content -LiteralPath (Get-ODWriterManifestPath) -Raw|ConvertFrom-Json -AsHashtable
        Driver-Need ($epoch.Schema -eq 2 -and $epoch.Phase -ceq 'writer-green' -and $epoch.PreviousFrozenSHA256 -ceq 'D8D31B8BE9F38A35F12CF5951DC28B3723D759B96AF3BC3B3D8B3D4D95FAC57C' -and $epoch.PinCount -eq 2236 -and $epoch.SourceCount -eq 31 -and (Driver-EqualJSON $epoch.SourcePaths $inputPaths) -and (Driver-EqualJSON $epoch.OwnerScopeRecovery (Driver-ScopeMap $info))) 'Exact scope source epoch'
        $intent=Get-Content -LiteralPath (Join-Path $packet writer-scope-repair.intent.json) -Raw|ConvertFrom-Json -AsHashtable
        Driver-Need ($intent.Schema -eq 1 -and $intent.AttemptConsumed -is [bool] -and $intent.AttemptConsumed -and $intent.ApprovalSHA256 -ceq $verifier.Binding.HistoricalFiles['owner-scope-verifier.approval.json'] -and $intent.GreenSHA256 -ceq $green -and $intent.ParentSHA256 -ceq $epoch.PreviousFrozenSHA256) 'Unique scope source intent'
    }
    Assert-ODOwnerScopeFacts ([ordered]@{Schema=1;ParentSHA256=(Driver-Hash (Join-Path $packet writer-controller-repair.frozen.json));HistoricalStopSHA256=(Driver-Hash (Join-Path $packet controller-parameter.stop.json));ApprovalValid=[bool]$approved;AttemptConsumed=$consumed;NewStopPresent=$newStop})
    return $info
}

function Assert-ODProjectionFacts([Collections.IDictionary]$Facts){
    if($null -eq $Facts -or (@($Facts.Keys) -join '|') -cne 'Schema|ParentSHA256|HistoricalStopSHA256|ApprovalValid|AttemptConsumed|NewStopPresent'){throw 'OD: error projection recovery mismatch'}
    if($Facts.Schema -isnot [int] -or $Facts.Schema -ne 1 -or $Facts.ParentSHA256 -isnot [string] -or $Facts.ParentSHA256 -cne '7F09A142E228FF16CD7228287DEBBB5EE60DEC9A91FBF1EF691F148C0A911587' -or $Facts.HistoricalStopSHA256 -isnot [string] -or $Facts.HistoricalStopSHA256 -cne '4B06ED585A7F63DC38F084D0718E653F17266E2065175E58CAD67EC5AEF6FD14'){throw 'OD: error projection recovery mismatch'}
    foreach($k in @('ApprovalValid','AttemptConsumed','NewStopPresent')){if($Facts[$k] -isnot [bool]){throw 'OD: error projection recovery mismatch'}}
    if(-not $Facts.ApprovalValid -or $Facts.AttemptConsumed -or $Facts.NewStopPresent){throw 'OD: error projection recovery mismatch'}
}
function Driver-ProjectionParent {
    $path=Join-Path $packet writer-scope-repair.frozen.json
    Driver-CheckHash $path '7F09A142E228FF16CD7228287DEBBB5EE60DEC9A91FBF1EF691F148C0A911587'
    $parent=Get-Content -Raw -LiteralPath $path|ConvertFrom-Json -AsHashtable
    Driver-Need ($parent.Schema -eq 2 -and $parent.PinCount -eq 2236 -and $parent.Pins.Count -eq 2236 -and $parent.SourceCount -eq 31 -and (Driver-EqualJSON $parent.SourcePaths $inputPaths)) 'Historical projection parent shape'
    foreach($pin in $parent.Pins){
        $relative=[IO.Path]::GetRelativePath($repo,$pin.VerificationPath).Replace('\','/')
        $actual=if($relative -cin $parent.SourcePaths){Join-Path (Join-Path $packet snapshots/writer-scope-rebind) $relative}else{$pin.VerificationPath}
        Driver-Need ($pin.ExpectedState -ceq 'Present') 'Historical present pins only'
        Driver-CheckHash $actual $pin.SHA256
    }
    $null=Driver-ControllerGate
    foreach($n in @('owner-scope-verifier-red.final.json','owner-scope-verifier-green.final.json')){
        $r=Get-Content -Raw -LiteralPath (Join-Path $packet $n)|ConvertFrom-Json -AsHashtable
        foreach($relative in $r.FixtureSHA256.Keys){
            $actual=if($relative -cin $parent.SourcePaths){Join-Path (Join-Path $packet snapshots/writer-scope-rebind) $relative}else{Join-Path $repo $relative}
            Driver-CheckHash $actual $r.FixtureSHA256[$relative]
        }
    }
    $r=Get-Content -Raw -LiteralPath (Join-Path $packet owner-scope-verifier-scope-green.final.json)|ConvertFrom-Json -AsHashtable
    Driver-Need ($r.PASS -eq 9 -and $r.FAIL -eq 0 -and $r.NotRun -eq 0 -and $r.ExitCode -eq 0 -and $null -eq $r.RootError) 'Historical successful scope disposition'
    foreach($relative in $r.SourceSHA256.Keys){Driver-CheckHash (Join-Path (Join-Path $packet snapshots/writer-scope-rebind) $relative) $r.SourceSHA256[$relative]}
    return $parent
}
function Driver-ProjectionReceipt([string]$Phase,$Binding){
    $red=$Phase -ceq 'Red';$scope=$Phase -ceq 'Scope'
    $prefix=if($scope){'error-projection-scope-green'}else{'error-projection-'+$Phase.ToLowerInvariant()}
    $path=Join-Path $packet ($prefix+'.final.json')
    $r=Get-Content -Raw -LiteralPath $path|ConvertFrom-Json -AsHashtable
    if($scope){Driver-Keys $r @('Schema','Phase','TestSHA256','InputSHA256','SourceSHA256','Groups','PASS','FAIL','NotRun','ExitCode','RootError') ProjectionScope}
    else{Driver-Keys $r @('Schema','Phase','TestSHA256','InputSHA256','CodeSHA256','Groups','PASS','FAIL','NotRun','ExitCode','RootError') ProjectionRegression}
    $test=if($scope){$Binding.TestSHA256}else{$Binding.RegressionTestSHA256}
    $input=if($scope){Driver-Hash (Join-Path $packet error-projection-scope.input.json)}else{$Binding.RegressionInputSHA256}
    $count=if($scope){9}else{10}
    foreach($k in @('Schema','PASS','FAIL','NotRun','ExitCode')){Driver-Need ($r[$k] -is [long]) 'Projection integer kinds'}
    Driver-Need ($r.Schema -eq 1 -and $r.Phase -ceq $(if($red){'Red'}else{'Green'}) -and $r.TestSHA256 -ceq $test -and $r.InputSHA256 -ceq $input -and $r.PASS -eq $(if($red){0}else{$count}) -and $r.FAIL -eq $(if($red){1}else{0}) -and $r.NotRun -eq $(if($red){9}else{0}) -and $r.ExitCode -eq $(if($red){1}else{0})) 'Projection exact phase disposition'
    $names=if($scope){@('FourFunctionsPersist','DefinitionsExact','ModuleIsolation','AttemptBindings','ConsumedAttempt','HistoricalEpoch','RecoveryGate','NewEpochBindings','OriginalContractsUnchanged')}else{@('NullInvocationInfoProjects','RealCaughtMetadataExact','NullStackProjects','SharedExceptionIdentity','NestedJsonRoundTrip','NullPositionRejected','MissingPositionRejected','NullStackRejected','MissingStackRejected','UnsupportedValueRejected')}
    Driver-Need ($r.Groups -is [array] -and $r.Groups.Count -eq $count) 'Projection exact group count'
    for($i=0;$i -lt $count;$i++){
        Driver-Keys $r.Groups[$i] @('Name','Status') ProjectionGroup
        Driver-Need ($r.Groups[$i].Name -ceq $names[$i] -and $r.Groups[$i].Status -ceq $(if($red){if($i -eq 0){'FAIL'}else{'NotRun'}}else{'PASS'})) 'Projection exact group order/status'
    }
    if($scope){
        Driver-Need (Driver-EqualJSON $r.SourceSHA256 $Binding.CandidateSources) 'Projection exact current source receipt'
        foreach($p in $r.SourceSHA256.Keys){Driver-CheckHash (Join-Path (Join-Path $packet snapshots/writer-projection-rebind) $p) $r.SourceSHA256[$p]}
    }else{
        $capture='docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/capture.ps1'
        Driver-Need ($r.CodeSHA256 -ceq $(if($red){$Binding.BaselineSources[$capture]}else{$Binding.CandidateSources[$capture]})) 'Projection code epoch'
    }
    if($red){
        Driver-Keys $r.RootError @('Type','Message','ErrorId','Position','Stack') ProjectionActualCause
        Driver-Need ($r.RootError.Type -ceq 'System.Management.Automation.PropertyNotFoundException' -and $r.RootError.ErrorId -ceq 'PropertyNotFoundStrict' -and $r.RootError.Message -ceq "The property 'PositionMessage' cannot be found on this object. Verify that the property exists." -and $r.RootError.Position -is [string] -and $r.RootError.Stack -is [string]) 'Projection original cause'
    }else{Driver-Need ($null -eq $r.RootError) 'Projection GREEN has no cause'}
    $intent=Get-Content -Raw -LiteralPath (Join-Path $packet ($prefix+'.intent.json'))|ConvertFrom-Json -AsHashtable
    Driver-Need ($intent.Schema -eq 1 -and $intent.AttemptConsumed -is [bool] -and $intent.AttemptConsumed -and $intent.Phase -ceq $r.Phase -and $intent.Computation -eq $(if($red){7}elseif($scope){9}else{8}) -and $intent.ComputationCeiling -eq 9 -and $intent.TestSHA256 -ceq $r.TestSHA256 -and $intent.InputSHA256 -ceq $r.InputSHA256) 'Projection unique phase intent'
    if(-not $red){Driver-Need ($intent.FullInputSHA256 -ceq (Driver-Hash (Join-Path $packet error-projection-scope.input.json))) 'Projection full input bound to phase'}
    return Driver-Hash $path
}
function Driver-ProjectionMap($Info){
    return [ordered]@{Schema=1;ParentSHA256='7F09A142E228FF16CD7228287DEBBB5EE60DEC9A91FBF1EF691F148C0A911587';HistoricalStopSHA256='4B06ED585A7F63DC38F084D0718E653F17266E2065175E58CAD67EC5AEF6FD14';ApprovalSHA256='26188CE28C596508E962A121E95B89064C5A25195460BF9B4D288A2C17D8632F';PlanSHA256=$Info.Binding.PlanSHA256;DesignSHA256=$Info.Binding.DesignSHA256;RegressionTestSHA256=$Info.Binding.RegressionTestSHA256;RegressionInputSHA256=$Info.Binding.RegressionInputSHA256;ScopeTestSHA256=$Info.Binding.TestSHA256;ScopeInputSHA256=(Driver-Hash (Join-Path $packet error-projection-scope.input.json));RedFinalSHA256=$Info.RedSHA256;GreenFinalSHA256=$Info.GreenSHA256;ScopeFinalSHA256=$Info.ScopeSHA256;HistoricalArtifacts=$Info.Approval.HistoricalFiles;Budget=[ordered]@{HistoricalHostsConsumed=2;HistoricalCompilesConsumed=2;AuxiliaryConsumed=9;AuxiliaryCeiling=9;HostCeiling=7;CompileCeiling=3;SnapshotCeiling=186;Targets=7;ConditionalReviews=1}}
}
function Driver-ProjectionGate {
    $newStop=Test-Path -LiteralPath (Join-Path $packet error-projection.stop.json)
    Driver-Need (-not $newStop) 'Projection recovery stopped; no retry'
    Driver-CheckHash (Join-Path $packet error-projection.approval.json) '26188CE28C596508E962A121E95B89064C5A25195460BF9B4D288A2C17D8632F'
    $a=Get-Content -Raw -LiteralPath (Join-Path $packet error-projection.approval.json)|ConvertFrom-Json -AsHashtable
    Driver-Need ($a.Schema -eq 1 -and $a.Approved -is [bool] -and $a.Approved -and $a.UserReply -ceq '是' -and $a.Method -ceq 'Native' -and $a.BaseCommit -ceq 'cf1e5ce0822febd98035955d99edf293ff917dc6') 'Projection explicit approval'
    foreach($pair in @(@('AdditionalAuxiliaryComputations',3),@('AdditionalHosts',1),@('AdditionalCompiles',1),@('AdditionalSnapshots',31),@('AdditionalTargets',0),@('AdditionalReviewers',0),@('ModulesPerComputation',2),@('AuxiliaryCeiling',9),@('HostCeiling',7),@('CompileCeiling',3),@('SnapshotCeiling',186))){Driver-Need ($a[$pair[0]] -is [long] -and $a[$pair[0]] -eq $pair[1]) 'Projection exact finite allowance'}
    foreach($n in $a.HistoricalFiles.Keys){Driver-CheckHash (Join-Path $packet $n) $a.HistoricalFiles[$n]}
    Driver-CheckHash (Join-Path $packet owner-scope-verifier.stop.json) '4B06ED585A7F63DC38F084D0718E653F17266E2065175E58CAD67EC5AEF6FD14'
    Driver-CheckHash (Join-Path $repo 'docs/superpowers/plans/2026-11-01-offline-error-projection.md') $a.PlanSHA256
    Driver-CheckHash (Join-Path $repo 'docs/superpowers/specs/2026-11-01-offline-error-projection-design.md') $a.DesignSHA256
    Driver-CheckHash (Join-Path $packet error-projection.input.json) 'FA99FE52EB06CEBE0BC403381E66B227D70D019C7716BA863963D6A66F9C0366'
    Driver-CheckHash (Join-Path $PSScriptRoot error-projection.tests.ps1) 'C27D74F4AC49DCE668747D5E212C986034F594C29047CC6E9BF3AD7E3AA8E995'
    Driver-CheckHash (Join-Path $PSScriptRoot error-projection-scope.tests.ps1) 'C084996264D3811F36B10CEF820B0D1D2FA05D0ABAA52D603BECBEFBC1AAE8AE'
    $b=Get-Content -Raw -LiteralPath (Join-Path $packet error-projection-scope.input.json)|ConvertFrom-Json -AsHashtable
    Driver-Need ($b.Schema -eq 1 -and $b.TestSHA256 -ceq 'C084996264D3811F36B10CEF820B0D1D2FA05D0ABAA52D603BECBEFBC1AAE8AE' -and $b.RegressionTestSHA256 -ceq 'C27D74F4AC49DCE668747D5E212C986034F594C29047CC6E9BF3AD7E3AA8E995' -and $b.RegressionInputSHA256 -ceq 'FA99FE52EB06CEBE0BC403381E66B227D70D019C7716BA863963D6A66F9C0366' -and $b.PlanSHA256 -ceq $a.PlanSHA256 -and $b.DesignSHA256 -ceq $a.DesignSHA256 -and $b.ApprovalSHA256 -ceq '26188CE28C596508E962A121E95B89064C5A25195460BF9B4D288A2C17D8632F' -and (Driver-EqualJSON $b.BaselineSources $a.BaselineSources)) 'Projection frozen inputs'
    Driver-Keys $b.CandidateSources @($a.BaselineSources.Keys) ProjectionCandidateWindow
    foreach($n in $b.HistoricalFiles.Keys){Driver-CheckHash (Join-Path $packet $n) $b.HistoricalFiles[$n]}
    $parent=Driver-ProjectionParent
    $info=[pscustomobject]@{Approval=$a;Binding=$b;Parent=$parent;RedSHA256=(Driver-ProjectionReceipt Red $b);GreenSHA256=(Driver-ProjectionReceipt Green $b);ScopeSHA256=(Driver-ProjectionReceipt Scope $b)}
    $grants=@('writer-green','contract-green','core','lifecycle','continuation')
    $start=if($Action -ceq 'Prepare' -and $Grant -ceq 'writer-projection-rebind'){0}elseif($Action -ceq 'Prepare' -and $Grant -ceq 'repair-green'){1}elseif($Action -ceq 'RunOne'){[array]::IndexOf($grants,$Grant)}else{-1}
    Driver-Need ($start -ge 0) 'Only approved projection stages'
    $consumed=$false
    for($i=$start;$i -lt $grants.Count;$i++){
        $slot=Get-ODAttemptSlot $grants[$i]
        foreach($suffix in @('intent.json','final.json','stdout.bin','stderr.bin','evidence.json')){if(Test-Path -LiteralPath (Join-Path $packet ($slot+'.'+$suffix))){$consumed=$true}}
        foreach($n in @(('capture/'+$slot),('unit/'+$slot))){if(Test-Path -LiteralPath (Join-Path $packet $n)){$consumed=$true}}
    }
    if($start -le 1 -and (Test-Path -LiteralPath (Join-Path $packet profiles))){$consumed=$true}
    if($start -eq 0 -or ($Action -ceq 'Prepare' -and $Grant -ceq 'repair-green')){foreach($n in @('repair-green.frozen.json','snapshots/repair-green')){if(Test-Path -LiteralPath (Join-Path $packet $n)){$consumed=$true}}}
    if($Action -ceq 'Prepare' -and $Grant -ceq 'writer-projection-rebind'){
        foreach($n in @('writer-projection-repair.intent.json','writer-projection-repair.frozen.json','snapshots/writer-projection-rebind')){if(Test-Path -LiteralPath (Join-Path $packet $n)){$consumed=$true}}
    }else{
        $epoch=Get-Content -Raw -LiteralPath (Get-ODWriterManifestPath)|ConvertFrom-Json -AsHashtable
        Driver-Need ($epoch.Schema -eq 2 -and $epoch.Phase -ceq 'writer-green' -and $epoch.PinCount -eq 2251 -and $epoch.Pins.Count -eq 2251 -and $epoch.PreviousFrozenSHA256 -ceq $a.ParentSHA256 -and (Driver-EqualJSON $epoch.SourcePaths $inputPaths) -and (Driver-EqualJSON $epoch.OwnerScopeRecovery $parent.OwnerScopeRecovery) -and (Driver-EqualJSON $epoch.ErrorProjectionRecovery (Driver-ProjectionMap $info))) 'Exact projection source epoch'
        $intent=Get-Content -Raw -LiteralPath (Join-Path $packet writer-projection-repair.intent.json)|ConvertFrom-Json -AsHashtable
        Driver-Need ($intent.Schema -eq 1 -and $intent.AttemptConsumed -is [bool] -and $intent.AttemptConsumed -and $intent.ParentSHA256 -ceq $a.ParentSHA256 -and $intent.ApprovalSHA256 -ceq '26188CE28C596508E962A121E95B89064C5A25195460BF9B4D288A2C17D8632F' -and $intent.GreenSHA256 -ceq $info.ScopeSHA256) 'Unique projection source intent'
    }
    Assert-ODProjectionFacts ([ordered]@{Schema=1;ParentSHA256=(Driver-Hash (Join-Path $packet writer-scope-repair.frozen.json));HistoricalStopSHA256=(Driver-Hash (Join-Path $packet owner-scope-verifier.stop.json));ApprovalValid=$true;AttemptConsumed=$consumed;NewStopPresent=$newStop})
    return $info
}


function Assert-ODPinBindingFacts([Collections.IDictionary]$Facts){
    if($null -eq $Facts -or (@($Facts.Keys) -join '|') -cne 'Schema|ParentSHA256|HistoricalStopSHA256|ApprovalValid|AttemptConsumed|NewStopPresent'){throw 'OD: pin binding recovery mismatch'}
    if($Facts.Schema -isnot [int] -or $Facts.Schema -ne 1 -or $Facts.ParentSHA256 -isnot [string] -or $Facts.ParentSHA256 -cne '182B85886304BFE3CBFB248DDA70F2EB99DF3E2158F574DBE299519DE7CE0461' -or $Facts.HistoricalStopSHA256 -isnot [string] -or $Facts.HistoricalStopSHA256 -cne '8581A9DA9A4CA3E7F2367757BE7616992889B434F42B1C4F35922B37921633ED'){throw 'OD: pin binding recovery mismatch'}
    foreach($k in @('ApprovalValid','AttemptConsumed','NewStopPresent')){if($Facts[$k] -isnot [bool]){throw 'OD: pin binding recovery mismatch'}}
    if(-not $Facts.ApprovalValid -or $Facts.AttemptConsumed -or $Facts.NewStopPresent){throw 'OD: pin binding recovery mismatch'}
}
function Driver-PinBindingReceipt([string]$Prefix,[int]$Number,$InputBinding,[string]$InputHash,$ExpectedSources){
    $red=$Prefix.EndsWith('-red',[StringComparison]::Ordinal)
    $focused=$Prefix.StartsWith('pin-binding-',[StringComparison]::Ordinal)
    $names=if($focused){@('ValidPinnedHash','MissingPinRejected','AbsentPinRejected','MalformedHashRejected','ConflictingPinsRejected','ActualHashMismatchRejected','IdenticalPinsAccepted')}else{@('ExactDescriptor','OtherControlsExcluded','NamedCallsUnchanged','UnknownVariableExcluded','ActualUnitStackAccepted','WrongSourceRejected','WrongFunctionRejected','WrongCallerLineRejected','BodyExtentRejected','MissingBindingRejected','BindingIdentityRejected','AmbiguousBindingRejected','RuntimeBindingAbsent','RuntimeBridgeRejected','FaultOperationPreserved','HistoricalEpoch','RecoveryGate','SourceWindow','FixedSlots','OriginalGuardsUnchanged')}
    $path=Join-Path $packet ($Prefix+'.final.json');$r=Get-Content -Raw $path|ConvertFrom-Json -AsHashtable
    Driver-Keys $r @('Schema','Phase','TestSHA256','InputSHA256','SourceSHA256','PostSourceSHA256','Groups','PASS','FAIL','NotRun','ExitCode','RootError','RecordingError') PinReceipt
    foreach($k in @('Schema','PASS','FAIL','NotRun','ExitCode')){Driver-Need ($r[$k] -is [long]) 'Pin receipt integer'}
    Driver-Need ($r.Schema -eq 1 -and $r.Phase -ceq $(if($red){'Red'}else{'Green'}) -and $r.TestSHA256 -ceq $InputBinding.TestSHA256 -and $r.InputSHA256 -ceq $InputHash -and $r.PASS -eq $(if($red){0}else{$names.Count}) -and $r.FAIL -eq $(if($red){1}else{0}) -and $r.NotRun -eq $(if($red){$names.Count-1}else{0}) -and $r.ExitCode -eq $(if($red){1}else{0}) -and $null -eq $r.RecordingError) 'Pin exact receipt'
    Driver-Need ((Driver-EqualJSON $r.SourceSHA256 $ExpectedSources) -and (Driver-EqualJSON $r.PostSourceSHA256 $ExpectedSources)) 'Pin actual source epoch'
    Driver-Need ($r.Groups -is [array] -and $r.Groups.Count -eq $names.Count) 'Pin receipt groups'
    for($i=0;$i -lt $names.Count;$i++){
        Driver-Keys $r.Groups[$i] @('Name','Status') PinGroup
        Driver-Need ($r.Groups[$i].Name -ceq $names[$i] -and $r.Groups[$i].Status -ceq $(if($red){if($i -eq 0){'FAIL'}else{'NotRun'}}else{'PASS'})) 'Pin exact ordered groups'
    }
    if($red){
        Driver-Keys $r.RootError @('Type','Message','ErrorId','Position','Stack') PinActualError
        $type=if($focused){'System.Management.Automation.PropertyNotFoundException'}else{'System.Management.Automation.RuntimeException'}
        $id=if($focused){'PropertyNotFoundStrict'}else{'OD: fixed owner-control descriptor missing'}
        $message=if($focused){"The property 'SHA256' cannot be found on this object. Verify that the property exists."}else{$id}
        Driver-Need ($r.RootError.Type -ceq $type -and $r.RootError.ErrorId -ceq $id -and $r.RootError.Message -ceq $message -and $r.RootError.Position -is [string] -and $r.RootError.Stack -is [string]) 'Pin expected actual RED'
    }else{Driver-Need ($null -eq $r.RootError) 'Pin GREEN no error'}
    $intent=Get-Content -Raw (Join-Path $packet ($Prefix+'.intent.json'))|ConvertFrom-Json -AsHashtable
    Driver-Keys $intent @('Schema','Phase','AttemptConsumed','Computation','ComputationCeiling','TestSHA256','InputSHA256','SourceSHA256') PinIntent
    foreach($k in @('Schema','Computation','ComputationCeiling')){Driver-Need ($intent[$k] -is [long]) 'Pin intent integer'}
    Driver-Need ($intent.Schema -eq 1 -and $intent.Phase -ceq $r.Phase -and $intent.AttemptConsumed -is [bool] -and $intent.AttemptConsumed -and $intent.Computation -eq $Number -and $intent.ComputationCeiling -eq 14 -and $intent.TestSHA256 -ceq $r.TestSHA256 -and $intent.InputSHA256 -ceq $r.InputSHA256 -and (Driver-EqualJSON $intent.SourceSHA256 $ExpectedSources)) 'Pin unique actual attempt'
    $tool=Get-Content -Raw (Join-Path $packet ($Prefix+'.tool-record.json'))|ConvertFrom-Json -AsHashtable
    Driver-Need ($tool.Result.exit_code -eq $r.ExitCode -and (Driver-EqualJSON (ConvertFrom-Json -InputObject $tool.Result.output -AsHashtable) $r)) 'Pin full actual tool output'
    return Driver-Hash $path
}
function Driver-PinBindingMap($Info){
    $artifacts=[ordered]@{};foreach($k in $Info.Approval.HistoricalFiles.Keys){$artifacts[$k]=$Info.Approval.HistoricalFiles[$k]}
    foreach($prefix in @('pin-binding-red','pin-binding-green','contract-call-model-v2-red','contract-call-model-v2-green')){
        foreach($suffix in @('intent.json','final.json','tool-record.json')){$rel='.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/'+$prefix+'.'+$suffix;$artifacts[$rel]=Driver-Hash (Join-Path $repo $rel)}
    }
    return [ordered]@{Schema=1;ParentSHA256=$Info.Approval.ParentSHA256;ApprovalSHA256=(Driver-Hash (Join-Path $packet pin-binding.approval.json));DesignSHA256=$Info.Approval.DesignSHA256;PlanSHA256=$Info.Approval.PlanSHA256;PinTestSHA256=$Info.PinInput.TestSHA256;PinInputSHA256=(Driver-Hash (Join-Path $packet pin-binding.input.json));IntermediateInputSHA256=(Driver-Hash (Join-Path $packet pin-binding-intermediate.input.json));PinRedFinalSHA256=$Info.PinRed;PinGreenFinalSHA256=$Info.PinGreen;ModelTestSHA256=$Info.ModelInput.TestSHA256;ModelInputSHA256=(Driver-Hash (Join-Path $packet contract-call-model-v2.input.json));CandidateInputSHA256=(Driver-Hash (Join-Path $packet contract-call-model-v2-candidate.input.json));ModelRedFinalSHA256=$Info.ModelRed;ModelGreenFinalSHA256=$Info.ModelGreen;HistoricalStopSHA256=$Info.Approval.HistoricalStopSHA256;HistoricalArtifacts=$artifacts;Consumed=[ordered]@{Auxiliary=14;Hosts=3;Compiles=3;Snapshots=155;Targets=0;Reviews=0};Ceilings=[ordered]@{Auxiliary=14;Hosts=7;Compiles=3;Snapshots=186}}
}
function Driver-PinBindingGate {
    Driver-Need (-not (Test-Path (Join-Path $packet pin-binding.stop.json))) 'Pin recovery stopped; no retry'
    Driver-CheckHash (Join-Path $packet pin-binding.approval.json) '2AFB70A25B1253422575BB041E8F45DCBB06DCDE4ADB94274E0BAEB312E29B6C'
    Driver-CheckHash (Join-Path $packet pin-binding.input.json) 'D85B147683327AB448E644BC054758EC21895289363AE1D5D2E28238E56346EF'
    Driver-CheckHash (Join-Path $packet pin-binding-intermediate.input.json) '956C568C4DE08FAE5EEF40557E15BA0AADC4AA29934CDACEE3A45F9046566B71'
    Driver-CheckHash (Join-Path $packet contract-call-model-v2.input.json) '1EE26AFC7B68ABD3AEEBB3EF1FF57DACF20086DFE414AB582D6E92C55778E487'
    $a=Get-Content -Raw (Join-Path $packet pin-binding.approval.json)|ConvertFrom-Json -AsHashtable
    Driver-Need ($a.Schema -eq 1 -and $a.UserReply -ceq '是' -and $a.Method -ceq 'Native' -and $a.BaseCommit -ceq 'd3e129ac720f129fe9c5f5d3d138e0f6cd4ea7c5') 'Pin explicit approval'
    foreach($pair in @(@('AddedAuxiliary',3),@('AuxiliaryCeiling',14),@('RemainingHosts',4),@('RemainingTargets',7),@('RemainingCompilations',0),@('RemainingSnapshots',31),@('RemainingReviewers',1))){Driver-Need ($a[$pair[0]] -is [long] -and $a[$pair[0]] -eq $pair[1]) 'Pin finite allowance'}
    Driver-CheckHash (Join-Path $repo 'docs/superpowers/specs/2026-11-02-offline-pin-binding-recovery-design.md') $a.DesignSHA256
    Driver-CheckHash (Join-Path $repo 'docs/superpowers/plans/2026-11-02-offline-pin-binding-recovery.md') $a.PlanSHA256
    foreach($rel in $a.HistoricalFiles.Keys){Driver-CheckHash (Join-Path $repo $rel) $a.HistoricalFiles[$rel]}
    Driver-CheckHash (Join-Path $packet error-projection.stop.json) 'AC592FFF3F7FE2671692352CF68C819967CC066527A937E0816FC5E9F07C71EB'
    Driver-CheckHash (Join-Path $packet contract-call-model.stop.json) $a.HistoricalStopSHA256
    Driver-CheckHash (Get-ODWriterManifestPath) $a.ParentSHA256
    $parent=Get-Content -Raw (Get-ODWriterManifestPath)|ConvertFrom-Json -AsHashtable
    Driver-Need ($parent.Schema -eq 2 -and $parent.PinCount -eq 2251 -and $parent.Pins.Count -eq 2251 -and $parent.SourceCount -eq 31 -and (Driver-EqualJSON $parent.SourcePaths $inputPaths)) 'Accepted writer parent'
    foreach($pin in $parent.Pins){
        $rel=[IO.Path]::GetRelativePath($repo,$pin.VerificationPath).Replace('\','/')
        $actual=if($rel -cin $inputPaths){Join-Path (Join-Path $packet snapshots/writer-projection-rebind) $rel}else{$pin.VerificationPath}
        Driver-Need ($pin.ExpectedState -ceq 'Present') 'Accepted writer pins present'
        Driver-CheckHash $actual $pin.SHA256
    }
    $older=Driver-ProjectionParent
    $projectionApproval=Get-Content -Raw (Join-Path $packet error-projection.approval.json)|ConvertFrom-Json -AsHashtable
    $projectionBinding=Get-Content -Raw (Join-Path $packet error-projection-scope.input.json)|ConvertFrom-Json -AsHashtable
    $historical=[pscustomobject]@{Approval=$projectionApproval;Binding=$projectionBinding;RedSHA256=(Driver-ProjectionReceipt Red $projectionBinding);GreenSHA256=(Driver-ProjectionReceipt Green $projectionBinding);ScopeSHA256=(Driver-ProjectionReceipt Scope $projectionBinding)}
    Driver-Need ((Driver-EqualJSON $parent.OwnerScopeRecovery $older.OwnerScopeRecovery) -and (Driver-EqualJSON $parent.ErrorProjectionRecovery (Driver-ProjectionMap $historical))) 'Preserved historical maps'
    Driver-RequireFinal writer-green
    $b=Get-Content -Raw (Join-Path $packet pin-binding.input.json)|ConvertFrom-Json -AsHashtable
    $mid=Get-Content -Raw (Join-Path $packet pin-binding-intermediate.input.json)|ConvertFrom-Json -AsHashtable
    $v=Get-Content -Raw (Join-Path $packet contract-call-model-v2.input.json)|ConvertFrom-Json -AsHashtable
    $c=Get-Content -Raw (Join-Path $packet contract-call-model-v2-candidate.input.json)|ConvertFrom-Json -AsHashtable
    foreach($inputBinding in @($b,$v)){
        Driver-Need ($inputBinding.Schema -eq 1 -and $inputBinding.ApprovalSHA256 -ceq (Driver-Hash (Join-Path $packet pin-binding.approval.json)) -and $inputBinding.ParentSHA256 -ceq $a.ParentSHA256 -and $inputBinding.HistoricalStopSHA256 -ceq $a.HistoricalStopSHA256 -and $inputBinding.DesignSHA256 -ceq $a.DesignSHA256 -and $inputBinding.PlanSHA256 -ceq $a.PlanSHA256) 'Pin input approval'
        foreach($rel in $inputBinding.HistoricalFiles.Keys){Driver-CheckHash (Join-Path $repo $rel) $inputBinding.HistoricalFiles[$rel]}
    }
    Driver-CheckHash (Join-Path $PSScriptRoot pin-binding.tests.ps1) $b.TestSHA256
    Driver-CheckHash (Join-Path $PSScriptRoot contract-call-model-v2.tests.ps1) $v.TestSHA256
    Driver-Need ($mid.Schema -eq 1 -and $mid.InputSHA256 -ceq (Driver-Hash (Join-Path $packet pin-binding.input.json)) -and (Driver-EqualJSON $b.BaselineSources $a.BaselineSources) -and (Driver-EqualJSON $v.BaselineSources $mid.SourceSHA256)) 'Rename-only epoch binding'
    Driver-Keys $c @('Schema','InputSHA256','CandidateSources','HistoricalFiles','ExpectedPinCount') PinCandidate
    Driver-Need ($c.Schema -eq 1 -and $c.InputSHA256 -ceq (Driver-Hash (Join-Path $packet contract-call-model-v2.input.json)) -and $c.ExpectedPinCount -eq 2276) 'Pin candidate identity'
    Driver-Keys $c.CandidateSources @($a.BaselineSources.Keys) PinCandidateWindow
    foreach($rel in $c.CandidateSources.Keys){Driver-CheckHash (Join-Path $repo $rel) $c.CandidateSources[$rel]}
    foreach($rel in $c.HistoricalFiles.Keys){Driver-CheckHash (Join-Path $repo $rel) $c.HistoricalFiles[$rel]}
    $info=[pscustomobject]@{Approval=$a;PinInput=$b;ModelInput=$v;Candidate=$c;Parent=$parent;PinRed=(Driver-PinBindingReceipt pin-binding-red 11 $b (Driver-Hash (Join-Path $packet pin-binding.input.json)) $b.BaselineSources);PinGreen=(Driver-PinBindingReceipt pin-binding-green 12 $b (Driver-Hash (Join-Path $packet pin-binding.input.json)) $mid.SourceSHA256);ModelRed=(Driver-PinBindingReceipt contract-call-model-v2-red 13 $v (Driver-Hash (Join-Path $packet contract-call-model-v2.input.json)) $v.BaselineSources);ModelGreen=(Driver-PinBindingReceipt contract-call-model-v2-green 14 $v (Driver-Hash (Join-Path $packet contract-call-model-v2.input.json)) $c.CandidateSources)}
    $grants=@('contract-green','core','lifecycle','continuation')
    $start=if($Action -ceq 'Prepare' -and $Grant -ceq 'pin-binding-rebind'){0}elseif($Action -ceq 'RunOne'){[array]::IndexOf($grants,$Grant)}else{-1}
    Driver-Need ($start -ge 0) 'Only unspent pin recovery stages'
    $consumed=$false
    for($i=$start;$i -lt $grants.Count;$i++){
        $slot=$grants[$i]
        foreach($suffix in @('intent.json','final.json','stdout.bin','stderr.bin','evidence.json')){if(Test-Path (Join-Path $packet ($slot+'.'+$suffix))){$consumed=$true}}
        foreach($rel in @(('capture/'+$slot),('unit/'+$slot))){if(Test-Path (Join-Path $packet $rel)){$consumed=$true}}
    }
    if($start -eq 0 -and (Test-Path (Join-Path $packet profiles))){$consumed=$true}
    if($Action -ceq 'Prepare'){
        foreach($rel in @('repair-green.frozen.json','repair-green.intent.json','snapshots/repair-green')){if(Test-Path (Join-Path $packet $rel)){$consumed=$true}}
    }else{
        $epoch=Get-Content -Raw (Join-Path $packet repair-green.frozen.json)|ConvertFrom-Json -AsHashtable
        Driver-Need ($epoch.PinCount -eq 2276 -and $epoch.PreviousFrozenSHA256 -ceq $a.ParentSHA256 -and (Driver-EqualJSON $epoch.PinBindingRecovery (Driver-PinBindingMap $info)) -and (Driver-EqualJSON $epoch.OwnerScopeRecovery $parent.OwnerScopeRecovery) -and (Driver-EqualJSON $epoch.ErrorProjectionRecovery $parent.ErrorProjectionRecovery)) 'Exact pin repair epoch'
        $intent=Get-Content -Raw (Join-Path $packet repair-green.intent.json)|ConvertFrom-Json -AsHashtable
        Driver-Need ($intent.Schema -eq 1 -and $intent.AttemptConsumed -is [bool] -and $intent.AttemptConsumed -and $intent.ParentSHA256 -ceq $a.ParentSHA256 -and $intent.ApprovalSHA256 -ceq (Driver-Hash (Join-Path $packet pin-binding.approval.json)) -and $intent.GreenSHA256 -ceq $info.ModelGreen) 'Unique pin source intent'
    }
    Assert-ODPinBindingFacts ([ordered]@{Schema=1;ParentSHA256=(Driver-Hash (Get-ODWriterManifestPath));HistoricalStopSHA256=(Driver-Hash (Join-Path $packet contract-call-model.stop.json));ApprovalValid=$true;AttemptConsumed=$consumed;NewStopPresent=(Test-Path (Join-Path $packet pin-binding.stop.json))})
    return $info
}
function Assert-ODExtentContractFacts([Collections.IDictionary]$Facts){
    if($null -eq $Facts -or (@($Facts.Keys) -join '|') -cne 'Schema|ApprovalValid|HistoryValid|CandidateValid|ReceiptsValid|AttemptConsumed|NewStopPresent' -or $Facts.Schema -ne 1){throw 'OD: extent contract recovery mismatch'}
    foreach($key in @('ApprovalValid','HistoryValid','CandidateValid','ReceiptsValid')){if($Facts[$key] -isnot [bool] -or -not $Facts[$key]){throw 'OD: extent contract recovery mismatch'}}
    foreach($key in @('AttemptConsumed','NewStopPresent')){if($Facts[$key] -isnot [bool] -or $Facts[$key]){throw 'OD: extent contract recovery mismatch'}}
}
function Driver-ExtentContractMap($Info){
    return [ordered]@{ParentSHA256=$Info.Input.ParentSHA256;WriterSHA256=$Info.Input.WriterSHA256;ApprovalSHA256=(Driver-Hash (Join-Path $packet extent-contract.approval.json));InputSHA256=(Driver-Hash (Join-Path $packet extent-contract.input.json));CandidateSHA256=(Driver-Hash (Join-Path $packet extent-contract-candidate.input.json));RedSHA256=$Info.RedSHA256;GreenSHA256=$Info.GreenSHA256;AddendumSHA256=$Info.Approval.AddendumSHA256;WriterArtifacts=$Info.WriterArtifacts}
}
function Driver-ExtentContractGate {
    Driver-Need ($Epoch -ceq 'extent-contract' -and (($Action -ceq 'Prepare' -and $Grant -ceq 'extent-contract-rebind') -or ($Action -ceq 'RunOne' -and $Grant -ceq 'contract-green'))) 'Only explicit extent contract route'
    Driver-Need (-not (Test-Path (Join-Path $packet extent-contract.stop.json))) 'Extent recovery stopped'
    Driver-CheckHash (Join-Path $packet extent-contract.approval.json) '8DF0E338BAF106C5B47FC11CDB5A222EAA8184F0B990F05F0A99464D4E21592D'
    Driver-CheckHash (Join-Path $packet extent-contract.input.json) '50A088C0BFE6E931BB5F1DB3194399DEF9F1AD15CC7C949EF4091EF2A93737E2'
    $a=Get-Content (Join-Path $packet extent-contract.approval.json) -Raw|ConvertFrom-Json -AsHashtable
    $b=Get-Content (Join-Path $packet extent-contract.input.json) -Raw|ConvertFrom-Json -AsHashtable
    Driver-Need ($a.Schema -eq 1 -and $a.UserReply -ceq '是' -and $a.NamingReply -ceq '是' -and $a.Method -ceq 'Native') 'Explicit execution and corrected identity approval'
    foreach($pair in @(@('AddedAuxiliary',2),@('ComputationCeiling',19),@('AddedHosts',1),@('AddedSnapshots',31),@('AddedTargets',0),@('AddedCompilations',0),@('AddedReviewers',0),@('ConsumedAuxiliary',17),@('ConsumedHosts',4),@('ConsumedCompilations',3),@('ConsumedSnapshots',186))){Driver-Need ($a[$pair[0]] -is [long] -and $a[$pair[0]] -eq $pair[1]) 'Exact extent allowances'}
    foreach($rel in $b.HistoricalFiles.Keys){Driver-CheckHash (Join-Path $repo $rel) $b.HistoricalFiles[$rel]}
    foreach($rel in $b.HelperSHA256.Keys){Driver-CheckHash (Join-Path $repo $rel) $b.HelperSHA256[$rel]}
    Driver-CheckHash (Join-Path $PSScriptRoot extent-contract.tests.ps1) $b.TestSHA256
    $parent=Get-Content (Join-Path $packet repair-green.frozen.json) -Raw|ConvertFrom-Json -AsHashtable
    Driver-Need ($parent.PinCount -eq 2276 -and $parent.SourceCount -eq 31 -and (Driver-EqualJSON $parent.SourcePaths $inputPaths)) 'Parent source inventory'
    Driver-CheckHash (Join-Path $packet repair-green.frozen.json) $b.ParentSHA256
    $c=Get-Content (Join-Path $packet extent-contract-candidate.input.json) -Raw|ConvertFrom-Json -AsHashtable
    Driver-Need ($c.Schema -eq 1 -and $c.InputSHA256 -ceq (Driver-Hash (Join-Path $packet extent-contract.input.json)) -and $c.ExpectedPinCount -eq 2306 -and (Driver-EqualJSON $c.AppendPaths $b.AppendPaths)) 'Exact candidate inputs'
    Driver-Keys $c.CandidateSources @($b.BaselineSources.Keys) CandidateSourceKeys
    $changed=@($c.CandidateSources.Keys|Where-Object {$c.CandidateSources[$_] -cne $b.BaselineSources[$_]})
    Driver-Need ($changed.Count -eq 4 -and @(Compare-Object $changed $b.AllowedSources).Count -eq 0) 'Only four bound source edits'
    foreach($rel in $c.CandidateSources.Keys){Driver-CheckHash (Join-Path $repo $rel) $c.CandidateSources[$rel]}
    $receiptHashes=@{}
    $names=@('BindingAvailable','ManifestAgreement','SlotAgreement','OldRouteBlocked','ApprovalRejected','SpentRejected','StopRejected','HashRejected','ParentRejected','WriterRejected','EpochRejected','PinConflictsRejected','SourceWindow','GuardsUnchanged','CommandExtentRegression','HistoricalContext','RuntimeNoFallback','UnitLineage','ReceiptIdentity','BudgetIdentity','CreateNewIdentity','OriginalSuiteUnchanged')
    foreach($phaseName in @('Red','Green')){
        $slotName='extent-contract-binding-'+$phaseName.ToLowerInvariant()
        $r=Get-Content (Join-Path $packet ($slotName+'.final.json')) -Raw|ConvertFrom-Json -AsHashtable
        $i=Get-Content (Join-Path $packet ($slotName+'.intent.json')) -Raw|ConvertFrom-Json -AsHashtable
        $expectedSources=if($phaseName -ceq 'Red'){$b.BaselineSources}else{$c.CandidateSources}
        $phaseHash=if($phaseName -ceq 'Red'){$null}else{Driver-Hash (Join-Path $packet extent-contract-candidate.input.json)}
        Driver-Need ($i.AttemptConsumed -is [bool] -and $i.AttemptConsumed -and $i.Computation -eq $(if($phaseName -ceq 'Red'){18}else{19}) -and $i.Ceiling -eq 19 -and $i.Phase -ceq $phaseName -and $i.TestSHA256 -ceq $b.TestSHA256 -and $i.InputSHA256 -ceq $c.InputSHA256 -and $i.PhaseInputSHA256 -ceq $phaseHash) 'Exact finite auxiliary intent'
        Driver-Need ($r.Schema -eq 1 -and $r.Phase -ceq $phaseName -and $r.TestSHA256 -ceq $b.TestSHA256 -and $r.InputSHA256 -ceq $c.InputSHA256 -and $r.PhaseInputSHA256 -ceq $phaseHash -and $null -eq $r.SetupError -and $null -eq $r.RecordingError -and (Driver-EqualJSON $r.SourceSHA256 $expectedSources) -and (Driver-EqualJSON $r.PostSourceSHA256 $expectedSources) -and (Driver-EqualJSON $r.HistoricalSHA256 $b.HistoricalFiles) -and (Driver-EqualJSON $r.PostHistoricalSHA256 $b.HistoricalFiles)) 'Actual auxiliary identities and unchanged inputs'
        Driver-Need ($r.Groups.Count -eq 22) 'All22auxiliary dispositions'
        for($j=0;$j -lt 22;$j++){
            $status=if($phaseName -ceq 'Green'){'PASS'}elseif($j -eq 0){'FAIL'}else{'NotRun'}
            Driver-Need ($r.Groups[$j].Name -ceq $names[$j] -and $r.Groups[$j].Status -ceq $status) 'Ordered exact auxiliary groups'
        }
        if($phaseName -ceq 'Red'){
            Driver-Need ($r.PASS -eq 0 -and $r.FAIL -eq 1 -and $r.NotRun -eq 21 -and $r.ExitCode -eq 1 -and $r.RootError.Type -ceq 'System.Management.Automation.RuntimeException' -and $r.RootError.Message -ceq 'Extent contract regression: missing Get-ODEpochManifestName' -and $r.RootError.ErrorId -ceq $r.RootError.Message) 'Expected missing binding RED18'
        }else{
            Driver-Need ($r.PASS -eq 22 -and $r.FAIL -eq 0 -and $r.NotRun -eq 0 -and $r.ExitCode -eq 0 -and $null -eq $r.RootError -and $r.CommandGroups.Count -eq 10 -and @($r.CommandGroups|Where-Object Status -CNE PASS).Count -eq 0) 'Exact GREEN19 and command invariants'
        }
        $receiptHashes[$phaseName]=Driver-Hash (Join-Path $packet ($slotName+'.final.json'))
        Driver-Need (Test-Path (Join-Path $packet ($slotName+'.tool-record.json'))) 'Full auxiliary tool record required'
    }
    Driver-Need ($c.RedFinalSHA256 -ceq $receiptHashes.Red) 'Candidate bound to actual RED18'
    Driver-RequireFinal writer-green legacy
    $writerArtifacts=[ordered]@{}
    foreach($name in @('writer-projection-repair.frozen.json','writer-projection-green.final.json','writer-projection-green.evidence.json','writer-projection-green.stdout.bin','writer-projection-green.stderr.bin')){$writerArtifacts[$name]=Driver-Hash (Join-Path $packet $name)}
    $info=[pscustomobject]@{Approval=$a;Input=$b;Candidate=$c;Parent=$parent;RedSHA256=$receiptHashes.Red;GreenSHA256=$receiptHashes.Green;WriterArtifacts=$writerArtifacts}
    $spent=$false
    foreach($suffix in @('intent.json','final.json','stdout.bin','stderr.bin','evidence.json','tool-record.json')){if(Test-Path (Join-Path $packet ('extent-contract-green.'+$suffix))){$spent=$true}}
    foreach($rel in @('capture/extent-contract-green','unit/extent-contract-green')){if(Test-Path (Join-Path $packet $rel)){$spent=$true}}
    if($Action -ceq 'Prepare'){
        foreach($rel in @('extent-contract.intent.json','extent-contract.frozen.json','snapshots/extent-contract')){if(Test-Path (Join-Path $packet $rel)){$spent=$true}}
    }else{
        $f=Get-Content (Join-Path $packet extent-contract.frozen.json) -Raw|ConvertFrom-Json -AsHashtable
        $i=Get-Content (Join-Path $packet extent-contract.intent.json) -Raw|ConvertFrom-Json -AsHashtable
        Driver-Need ($i.AttemptConsumed -is [bool] -and $i.AttemptConsumed -and $i.ParentSHA256 -ceq $b.ParentSHA256 -and $i.GreenSHA256 -ceq $receiptHashes.Green -and $i.ApprovalSHA256 -ceq $b.ApprovalSHA256) 'Unique completed Prepare'
        Assert-ODExtentLineage $f $parent $b.ParentSHA256 $b.WriterSHA256
        Driver-Need ((Driver-EqualJSON $f.ExtentContractRecovery (Driver-ExtentContractMap $info))) 'Frozen new provenance'
        foreach($key in @('OwnerScopeRecovery','ErrorProjectionRecovery','PinBindingRecovery')){Driver-Need (Driver-EqualJSON $f[$key] $parent[$key]) 'Preserved recovery maps'}
        foreach($rel in $inputPaths){Driver-CheckHash (Join-Path (Join-Path $packet snapshots/extent-contract) $rel) $c.CandidateSources[$rel]}
    }
    Assert-ODExtentContractFacts ([ordered]@{Schema=1;ApprovalValid=$true;HistoryValid=$true;CandidateValid=$true;ReceiptsValid=$true;AttemptConsumed=$spent;NewStopPresent=(Test-Path (Join-Path $packet extent-contract.stop.json))})
    return $info
}
function Driver-PrepareExtent($Info){
    Driver-Static
    Driver-Need ($PreviousFrozenSHA256 -ceq $Info.Input.ParentSHA256) 'Exact original repair parent'
    Driver-JSON (Join-Path $packet extent-contract.intent.json) ([ordered]@{Schema=1;AttemptConsumed=$true;ParentSHA256=$PreviousFrozenSHA256;ApprovalSHA256=$Info.Input.ApprovalSHA256;GreenSHA256=$Info.GreenSHA256})
    $manifest=Get-Content (Join-Path $packet repair-green.frozen.json) -Raw|ConvertFrom-Json -AsHashtable
    $pins=[Collections.Generic.List[object]]::new();$map=[Collections.Generic.Dictionary[string,object]]::new([StringComparer]::OrdinalIgnoreCase)
    foreach($pin in $manifest.Pins){
        $rel=[IO.Path]::GetRelativePath($repo,$pin.VerificationPath).Replace('\','/')
        $hash=if($rel -cin $Info.Input.AllowedSources){$Info.Candidate.CandidateSources[$rel]}else{$pin.SHA256}
        Driver-AddPin $map $pins $pin.VerificationPath $hash $pin.Role
    }
    foreach($rel in $Info.Input.AppendPaths){$path=Join-Path $repo $rel;Driver-Need (-not $map.ContainsKey([IO.Path]::GetFullPath($path))) 'Unique new provenance';Driver-AddPin $map $pins $path (Driver-Hash $path) 'extent-contract-provenance'}
    Driver-Need ($pins.Count -eq 2306) 'Exact30new pins'
    $manifest.Pins=$pins.ToArray();$manifest.PinCount=2306;$manifest.Phase='extent-contract';$manifest.PreviousFrozenSHA256=$PreviousFrozenSHA256
    $manifest.ChangedInputs=$Info.Input.AllowedSources;$manifest.MaxHosts=1;$manifest.MaxChildren=0;$manifest.MaxCompilations=0;$manifest.MaxReviewers=0
    $manifest.ExtentContractRecovery=Driver-ExtentContractMap $Info
    Driver-Verify $manifest
    $base=Join-Path $packet snapshots/extent-contract
    foreach($rel in $inputPaths){$target=Join-Path $base $rel;$null=[IO.Directory]::CreateDirectory([IO.Path]::GetDirectoryName($target));Driver-Once $target ([IO.File]::ReadAllBytes((Join-Path $repo $rel)));Driver-CheckHash $target $Info.Candidate.CandidateSources[$rel]}
    $manifestPath=Join-Path $packet extent-contract.frozen.json;Driver-JSON $manifestPath $manifest
    [ordered]@{Prepared=$Grant;Phase=$manifest.Phase;FrozenSHA256=(Driver-Hash $manifestPath);Pins=2306;Snapshots=31;ChangedInputs=$manifest.ChangedInputs}|ConvertTo-Json
}

Driver-Need ($Epoch -cin @('legacy','extent-contract')) 'Explicit known epoch'
if($Epoch -ceq 'extent-contract'){
    $extentInputs=Driver-ExtentContractGate
    if($Action -ceq 'Prepare'){Driver-PrepareExtent $extentInputs;return}
}else{$pinInputs=Driver-PinBindingGate}
if($Action -ceq 'Prepare'){
    Driver-Need ($Grant -ceq 'pin-binding-rebind') 'Only unique pin binding preparation'
    Driver-Static
    Driver-Need ($PreviousFrozenSHA256 -ceq $pinInputs.Approval.ParentSHA256) 'Exact writer parent required'
    $manifest=Get-Content -Raw (Get-ODWriterManifestPath)|ConvertFrom-Json
    $allowed=@($pinInputs.Approval.BaselineSources.Keys);$changed=[Collections.Generic.List[string]]::new()
    foreach($pin in $manifest.Pins){
        $rel=[IO.Path]::GetRelativePath($repo,$pin.VerificationPath).Replace('\','/')
        if($rel -cin $allowed){
            $hash=Driver-Hash $pin.VerificationPath
            Driver-Need ($pin.ExpectedState -ceq 'Present' -and $pin.SHA256 -cne $hash) 'Every approved source changes'
            $pin.SHA256=$hash;$changed.Add($rel)
        }else{Driver-CheckHash $pin.VerificationPath $pin.SHA256}
    }
    Driver-Need ($changed.Count -eq 3 -and @(Compare-Object $changed.ToArray() $allowed).Count -eq 0) 'Exact three-source window'
    $manifest.Phase='repair-green';$manifest.PreviousFrozenSHA256=$PreviousFrozenSHA256;$manifest.ChangedInputs=$changed.ToArray()
    $pins=[Collections.Generic.List[object]]::new();$map=[Collections.Generic.Dictionary[string,object]]::new([StringComparer]::OrdinalIgnoreCase)
    foreach($pin in $manifest.Pins){Driver-AddPin $map $pins $pin.VerificationPath $pin.SHA256 $pin.Role}
    $provenance=@(
        '.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/writer-projection-repair.frozen.json',
        '.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/error-projection.stop.json',
        '.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/contract-call-model.stop.json',
        'docs/superpowers/specs/2026-11-02-offline-contract-call-model-design.md',
        'docs/superpowers/plans/2026-11-02-offline-contract-call-model.md',
        '.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/contract-call-model.approval.json',
        'docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/contract-call-model.tests.ps1',
        '.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/contract-call-model.input.json',
        '.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/contract-call-model-red.final.json',
        '.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/contract-call-model-red.tool-record.json',
        'docs/superpowers/specs/2026-11-02-offline-pin-binding-recovery-design.md',
        'docs/superpowers/plans/2026-11-02-offline-pin-binding-recovery.md',
        '.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/pin-binding.approval.json',
        'docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/pin-binding.tests.ps1',
        '.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/pin-binding.input.json',
        '.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/pin-binding-intermediate.input.json',
        '.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/pin-binding-red.final.json',
        '.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/pin-binding-green.final.json',
        'docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/contract-call-model-v2.tests.ps1',
        '.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/contract-call-model-v2.input.json',
        '.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/contract-call-model-v2-candidate.input.json',
        '.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/contract-call-model-v2-red.final.json',
        '.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/contract-call-model-v2-green.final.json',
        '.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/writer-projection-green.final.json',
        '.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/error-projection-writer.tool-record.json'
    )
    foreach($rel in $provenance){$path=Join-Path $repo $rel;Driver-Need (-not $map.ContainsKey([IO.Path]::GetFullPath($path))) 'New unique provenance';Driver-AddPin $map $pins $path (Driver-Hash $path) 'pin-binding-provenance'}
    Driver-Need ($pins.Count -eq 2276) 'Exactly twenty-five new provenance pins'
    $manifest.Pins=$pins.ToArray();$manifest.PinCount=2276;$manifest.MaxHosts=4;$manifest.MaxChildren=7;$manifest.MaxCompilations=0;$manifest.MaxReviewers=1
    Add-Member -InputObject $manifest -NotePropertyName PinBindingRecovery -NotePropertyValue (Driver-PinBindingMap $pinInputs)
    Driver-Need ($manifest.InheritedPins -eq 2482 -and $manifest.InheritedUniquePins -eq 2172 -and $manifest.SourceCount -eq 31 -and (Driver-EqualJSON $manifest.SourcePaths $inputPaths)) 'Original source inventory'
    Driver-Verify $manifest
    Driver-JSON (Join-Path $packet repair-green.intent.json) ([ordered]@{Schema=1;AttemptConsumed=$true;ParentSHA256=$PreviousFrozenSHA256;ApprovalSHA256=(Driver-Hash (Join-Path $packet pin-binding.approval.json));GreenSHA256=$pinInputs.ModelGreen})
    $base=Join-Path $packet snapshots/repair-green;Driver-Need (-not (Test-Path $base)) 'Snapshot attempted'
    foreach($relative in $inputPaths){$target=Join-Path $base $relative;$null=[IO.Directory]::CreateDirectory([IO.Path]::GetDirectoryName($target));Driver-Once $target ([IO.File]::ReadAllBytes((Join-Path $repo $relative)));Driver-CheckHash $target (Driver-Hash (Join-Path $repo $relative))}
    $manifestPath=Join-Path $packet repair-green.frozen.json;Driver-JSON $manifestPath $manifest
    [ordered]@{Prepared=$Grant;Phase=$manifest.Phase;FrozenSHA256=(Driver-Hash $manifestPath);Pins=$manifest.PinCount;InheritedPins=2482;InheritedUniquePins=2172;Snapshots=31;ChangedInputs=$manifest.ChangedInputs}|ConvertTo-Json
    return
}
if($Epoch -ceq 'legacy'){Driver-Need (-not (Test-Path -LiteralPath (Join-Path $packet pin-binding.stop.json))) 'Pin execution stopped; no retry'}
Driver-Need (-not [string]::IsNullOrEmpty($FrozenSHA256)) 'Missing frozen identity'
Driver-Need ($Grant -cin @('contract-green','core','lifecycle','continuation')) 'Only four remaining approved capture hosts'
$phase=if($Epoch -ceq 'extent-contract'){'extent-contract'}elseif($Grant -ceq 'writer-green'){'writer-green'}else{'repair-green'}
$manifestPath=if($Grant -ceq 'writer-green'){Get-ODWriterManifestPath}else{Join-Path $packet (Get-ODEpochManifestName $Epoch)}
Driver-CheckHash $manifestPath $FrozenSHA256
$manifest=Get-Content -LiteralPath $manifestPath -Raw | ConvertFrom-Json -AsHashtable
Driver-Need ($manifest.Schema -eq 2 -and $manifest.Phase -ceq $phase -and $manifest.Pins.Count -eq $manifest.PinCount -and $manifest.InheritedPins -eq 2482 -and $manifest.InheritedUniquePins -gt 0 -and $manifest.InheritedUniquePins -le $manifest.InheritedPins -and $manifest.PinCount -ge $manifest.InheritedUniquePins -and $manifest.SourceCount -eq 31 -and (Driver-EqualJSON $manifest.SourcePaths $inputPaths) -and $manifest.MaxHosts -eq $(if($Epoch -ceq 'extent-contract'){1}else{4}) -and $manifest.MaxChildren -eq $(if($Epoch -ceq 'extent-contract'){0}else{7}) -and $manifest.MaxCompilations -eq 0 -and $manifest.MaxReviewers -eq $(if($Epoch -ceq 'extent-contract'){0}else{1})) 'Exact new source ownership and finite budget'
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
$slot=Get-ODAttemptSlot $Grant $Epoch
$intentPath=Join-Path $packet ($slot+'.intent.json');$finalPath=Join-Path $packet ($slot+'.final.json')
foreach($suffix in @('intent.json','final.json','stdout.bin','stderr.bin','evidence.json')){Driver-Need (-not (Test-Path -LiteralPath (Join-Path $packet ($slot+'.'+$suffix)))) 'Grant attempted/unknown; no retry'}
Driver-JSON $intentPath ([ordered]@{Grant=$Grant;AttemptSlot=$slot;AttemptConsumed=$true;FrozenSHA256=$FrozenSHA256;AuthorizedHost=1;AuthorizedChildren=$oracle.Grants[$Grant].Children})
$row=[ordered]@{Grant=$Grant;AttemptSlot=$slot;Started=$false;PID=$null;OwnedExit=$false;StreamsComplete=$false;ExitCode=$null;TimedOut=$false;TerminationRequested=$false;PrimaryError=$null;SecondaryErrors=@();Overflow=$false;PostPinsIntact=$false;Exact=$false;StdoutSHA256=$null;StderrSHA256=$null;EvidenceSHA256=$null;EvidenceValid=$false;ChildrenObserved=0;EvidenceState='Unknown';ExpectedNegative=$false;ExpectedOutputSHA256=$null;Rescue=$false;SourceVerified=$false;LiveActionsPerformed=$false;RuntimeApproved=$false;FormalAcceptance=$false;RealGateImplemented=$false}
$row.FrozenSHA256=$FrozenSHA256
$cleanup=[Collections.Generic.List[string]]::new()
function Driver-Error($Record){
    if($null -eq $row.PrimaryError){$row.PrimaryError=$Record.ToString()+"`n"+$Record.ScriptStackTrace}
    elseif($cleanup.Count -lt 8){$cleanup.Add($Record.ToString())}else{$row.Overflow=$true}
}
$process=[Diagnostics.Process]::new();$out=$null;$err=$null;$ot=$null;$et=$null
$timer=[Diagnostics.Stopwatch]::StartNew()
try{
    $out=[IO.File]::Open((Join-Path $packet ($slot+'.stdout.bin')),[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::Read)
    $err=[IO.File]::Open((Join-Path $packet ($slot+'.stderr.bin')),[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::Read)
    $process.StartInfo=[Diagnostics.ProcessStartInfo]::new()
    $process.StartInfo.FileName=$engine;$process.StartInfo.UseShellExecute=$false;$process.StartInfo.CreateNoWindow=$true
    $process.StartInfo.RedirectStandardOutput=$true;$process.StartInfo.RedirectStandardError=$true;$process.StartInfo.WorkingDirectory=$repo
    foreach($arg in @('-NoLogo','-NoProfile','-NonInteractive','-File',(Join-Path $PSScriptRoot capture.ps1),'-Grant',$Grant,'-FrozenSHA256',$FrozenSHA256,'-Epoch',$Epoch)){$process.StartInfo.ArgumentList.Add($arg)}
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
        $stdoutPath=Join-Path $packet ($slot+'.stdout.bin');$stderrPath=Join-Path $packet ($slot+'.stderr.bin');$evidencePath=Join-Path $packet ($slot+'.evidence.json')
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
        $row.ChildrenObserved=Driver-Evidence $actual $profile $oracle $Grant $manifest $FrozenSHA256 $Epoch
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
    $stopName=if($Epoch -ceq 'extent-contract'){'extent-contract.stop.json'}else{'pin-binding.stop.json'}
    if((Test-Path -LiteralPath $stopRoot) -and -not (Test-Path -LiteralPath (Join-Path $stopRoot $stopName))){
        Driver-JSON (Join-Path $stopRoot $stopName) ([ordered]@{Action=$Action;Grant=$Grant;Message=$actual.Exception.Message;ErrorId=$actual.FullyQualifiedErrorId;Position=$actual.InvocationInfo.PositionMessage;Stack=$actual.ScriptStackTrace;Remainder='FROZEN'})
    }
    throw
}
