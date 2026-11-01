param(
    [Parameter(Mandatory)][ValidateSet('Prepare','RunOne')][string]$Action,
    [Parameter(Mandatory)][ValidateSet('seed','writer-red','contract-red','writer-green','contract-green','repair-green','core','lifecycle','continuation')][string]$Grant,
    [ValidatePattern('^[A-F0-9]{64}$')][string]$FrozenSHA256,
    [ValidatePattern('^[A-F0-9]{64}$')][string]$PreviousFrozenSHA256,
    [string]$PreparationAdjustment
)
try{
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$repo=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))
$packet=Join-Path $repo '.superpowers/sdd/2026-10-31-offline-integrated-review-repair'
$engine='C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe'
$inputPaths=@('docs/diagnostics/isolated-fuzz/p2b-integrated-review-repair/.gitattributes','docs/diagnostics/isolated-fuzz/p2b-integrated-review-repair/README.md','docs/diagnostics/isolated-fuzz/p2b-integrated-review-repair/expected-output.json','docs/diagnostics/isolated-fuzz/p2b-integrated-review-repair/test-helpers.ps1','docs/diagnostics/isolated-fuzz/p2b-integrated-review-repair/entry.tests.ps1','docs/diagnostics/isolated-fuzz/p2b-integrated-review-repair/core.tests.ps1','docs/diagnostics/isolated-fuzz/p2b-integrated-review-repair/prefix.tests.ps1','docs/diagnostics/isolated-fuzz/p2b-integrated-review-repair/owned.tests.ps1','docs/diagnostics/isolated-fuzz/p2b-integrated-review-repair/live.tests.ps1','docs/diagnostics/isolated-fuzz/p2b-integrated-review-repair/capture.ps1','docs/diagnostics/isolated-fuzz/p2b-integrated-review-repair/driver.ps1','scripts/offline-diagnostics-v3/diagnose-offline.ps1','scripts/offline-diagnostics-v3/private/catalog.json','scripts/offline-diagnostics-v3/private/session.psm1','scripts/offline-diagnostics-v3/private/sinks.ps1','scripts/offline-diagnostics-v3/private/owned-step.ps1','scripts/offline-diagnostics-v3/private/sequence.ps1','scripts/offline-diagnostics-v3/private/profiles.ps1','scripts/offline-diagnostics-v3/private/observations.ps1','scripts/offline-diagnostics-v3/.gitattributes','scripts/offline-diagnostics-v3/private/write-boundary.ps1','docs/diagnostics/isolated-fuzz/p2b-integrated-review-repair/observation-contract.ps1','docs/diagnostics/isolated-fuzz/p2b-integrated-review-repair/fault-stream.cs','docs/diagnostics/isolated-fuzz/p2b-integrated-review-repair/record-boundary.tests.ps1','docs/diagnostics/isolated-fuzz/p2b-integrated-review-repair/observation-contract.tests.ps1','docs/diagnostics/isolated-fuzz/p2b-integrated-review-repair/case-module.ps1','docs/diagnostics/isolated-fuzz/p2b-integrated-review-repair/publisher.psm1','docs/diagnostics/isolated-fuzz/p2b-integrated-review-repair/recording.tests.ps1')
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
    foreach($name in @('case-module.ps1','publisher.psm1','recording.tests.ps1','.gitattributes','README.md','expected-output.json','test-helpers.ps1','entry.tests.ps1','core.tests.ps1','prefix.tests.ps1','owned.tests.ps1','live.tests.ps1','capture.ps1','driver.ps1','record-boundary.tests.ps1','observation-contract.tests.ps1','fault-stream.cs','observation-contract.ps1')){
        if(-not [IO.File]::Exists((Join-Path $PSScriptRoot $name))){throw ('Missing explicit frozen source '+$name)}
    }
    foreach($relative in $inputPaths){
        $path=[IO.Path]::GetFullPath((Join-Path $repo $relative))
        if(-not [IO.File]::Exists($path)){continue}
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
        Driver-ErrorLocation $r.Error @('scripts/offline-diagnostics-v3/private/sinks.ps1')
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
}
function Driver-UnitEvidence($Evidence,$Oracle,[string]$CurrentGrant){
    $expected=$Oracle.Grants[$CurrentGrant]
    Driver-Need ($Evidence.NotRun.Count -eq 0 -and $Evidence.Status -ceq 'COMPLETE') 'Complete unit execution'
    if($CurrentGrant.StartsWith('writer-')){
        Driver-Need ($Evidence.Observations.Count -eq 10) 'Ten record cases'
        for($i=0;$i -lt 10;$i++){
            $row=$Evidence.Observations[$i];$v=$row.Value;$g=$expected.Groups[$i];$f=$g.Facts
            Driver-Need ($row.Name -ceq 'RecordCase') 'Record observation order'
            Driver-Keys $v @('Boundary','Case','Thrown','Primary','Secondary','OpenCalls','WriteCalls','FlushCalls','DisposeCalls','BytesExact','ExistingUnchanged','FactoryIsolated','Overflow') RecordCase
            Driver-Need ($v.Boundary -ceq $f.Boundary -and $v.Case -ceq $f.Case -and $v.Secondary.Count -eq $f.SecondaryCount -and -not $v.Overflow) 'Record identity/error bounds'
            foreach($key in @('OpenCalls','WriteCalls','FlushCalls','DisposeCalls','BytesExact','ExistingUnchanged','FactoryIsolated')){Driver-Need ($v[$key] -ceq $f[$key] -and $v[$key].GetType() -eq $f[$key].GetType()) ('Record exact field '+$key)}
            $root=Join-Path $packet ('unit/'+$CurrentGrant+'/'+$v.Boundary+'/'+$v.Case)
            $op=if($v.Case.StartsWith('Write')){'Write'}else{'Flush'};$argc=if($op -ceq 'Write'){3}else{0}
            $opMessage=('Exception calling "{0}" with "{1}" argument(s): "Cannot access a closed file."' -f $op,$argc)
            $disposeMessage=('Exception calling "Dispose" with "0" argument(s): "Could not find file ''{0}''."' -f (Join-Path $root absent.txt))
            $openMessage=('Exception calling "Open" with "4" argument(s): "The file ''{0}'' already exists."' -f (Join-Path $root record.bin))
            $records=@();if($null -ne $v.Thrown){$records+=,$v.Thrown};if($null -ne $v.Primary){$records+=,$v.Primary};$records+=@($v.Secondary)
            if($v.Case -ceq 'SuccessOnce'){Driver-Need ($records.Count -eq 0) 'Success no errors'}
            else{
                Driver-Need ($null -ne $v.Thrown -and $null -ne $v.Primary) 'Real unit causes present'
                foreach($errorRecord in $records){
                    $message=switch -CaseSensitive ($errorRecord.ErrorId){'ObjectDisposedException'{$opMessage}'FileNotFoundException'{$disposeMessage}'IOException'{$openMessage}default{throw 'Unknown unit error'}}
                    Driver-UnitError $errorRecord $errorRecord.ErrorId $message 'scripts/offline-diagnostics-v3/private/write-boundary.ps1'
                }
                if($v.Case.EndsWith('FirstCause')){
                    $red=$CurrentGrant -ceq 'writer-red'
                    $throwId=if($red -and $v.Boundary -ceq 'Intent'){'FileNotFoundException'}else{'ObjectDisposedException'}
                    $primaryId=if($red -and $v.Boundary -ceq 'Capture'){'FileNotFoundException'}else{'ObjectDisposedException'}
                    $secondaryId=if($red -and $v.Boundary -ceq 'Capture'){'ObjectDisposedException'}else{'FileNotFoundException'}
                    Driver-Need ($v.Thrown.ErrorId -ceq $throwId -and $v.Primary.ErrorId -ceq $primaryId -and $v.Secondary[0].ErrorId -ceq $secondaryId) 'Seed versus fixed cause ordering'
                }else{Driver-Need (Driver-EqualJSON $v.Thrown $v.Primary) 'Single cause same projection'}
            }
        }
    }else{
        Driver-Need ($Evidence.Observations.Count -eq 23) 'Twenty-three schema cases'
        for($i=0;$i -lt 23;$i++){
            $row=$Evidence.Observations[$i];$v=$row.Value;$f=$expected.Groups[$i].Facts
            Driver-Keys $v @('Index','Kind','Profile','ActualRejected','Error') SchemaCase
            Driver-Need ($row.Name -ceq 'SchemaCase' -and $v.Index -eq $f.Index -and $v.Kind -ceq $f.Kind -and $v.ActualRejected -is [bool] -and $v.ActualRejected -eq $f.ActualRejected) 'Fixed schema result'
            $profileIndex=if($i -lt 3){$i}elseif($f.Index -eq 16){2}elseif($f.Index -cin @(17,18)){1}else{0}
            $profileId=@('entry-migration-core','entry-migration-owned','entry-migration-live')[$profileIndex]
            Driver-Need ($v.Profile -ceq $profileId) 'Fixed unit profile data'
            if($v.ActualRejected){
                Driver-Keys $v.Error @('Type','Message','ErrorId','Position','Stack') ContractError
                Driver-Need ($v.Error.Type -ceq 'System.Management.Automation.RuntimeException' -and $v.Error.Message -ceq 'OD: observation contract mismatch' -and $v.Error.ErrorId -ceq 'OD: observation contract mismatch') 'Only exact contract error is rejection'
                Driver-ErrorLocation $v.Error @('docs/diagnostics/isolated-fuzz/p2b-integrated-review-repair/observation-contract.ps1')
            }else{Driver-Need ($null -eq $v.Error) 'No unexpected unit exception'}
        }
    }
}

. (Join-Path $PSScriptRoot observation-contract.ps1)

function Driver-Recorder($Evidence,$Oracle,[string]$CurrentGrant,$Manifest){
    $records=@($Evidence.Observations | Where-Object {$_.Name -ceq 'Recorder'})
    Driver-Need ($records.Count -eq 1 -and $Evidence.Observations[-1].Name -ceq 'Recorder') 'One final recorder observation'
    $r=$records[0].Value
    Driver-Keys $r $Oracle.Recorder.RootKeys Recorder
    $writerPath=[IO.Path]::GetFullPath((Join-Path $repo 'scripts/offline-diagnostics-v3/private/write-boundary.ps1'))
    Driver-Keys $r.Publisher @('Name','Path','WriterFile','FactoryFile','WriterSHA256','DefaultFactory','PrimaryNull','SecondaryCount','Overflow') PublisherBinding
    Driver-Need ($r.Publisher.Name -ceq 'publisher' -and $r.Publisher.Path -ceq [IO.Path]::GetFullPath((Join-Path $PSScriptRoot publisher.psm1)) -and $r.Publisher.WriterFile -ceq $writerPath -and $r.Publisher.FactoryFile -ceq $writerPath -and $r.Publisher.WriterSHA256 -ceq (Driver-PinHash $Manifest 'scripts/offline-diagnostics-v3/private/write-boundary.ps1')) 'Actual production publisher binding'
    foreach($key in @('DefaultFactory','PrimaryNull')){Driver-Need ($r.Publisher[$key] -is [bool] -and $r.Publisher[$key]) 'Publisher default/clean state'}
    Driver-Need ($r.Publisher.SecondaryCount -eq 0 -and $r.Publisher.Overflow -is [bool] -and -not $r.Publisher.Overflow) 'Publisher no cleanup errors'
    $writerGrant=$CurrentGrant -cin @('writer-red','writer-green')
    $expectedCases=if($writerGrant){10}else{0}
    Driver-Need ($r.Cases -is [array] -and $r.Cases.Count -eq $expectedCases) 'Exact case module inventory'
    $stages=[Collections.Generic.List[object]]::new()
    $stages.Add([ordered]@{Stage='Before';Case=''})
    if($writerGrant){foreach($c in $Oracle.WriterCases){foreach($stage in @('Created','Replaced','Finally')){$stages.Add([ordered]@{Stage=$stage;Case=$c.Name})}}}
    $stages.Add([ordered]@{Stage='After';Case=''})
    Driver-Need ($r.Stages -is [array] -and $r.Stages.Count -eq $stages.Count) 'Complete recorder stages'
    for($i=0;$i -lt $stages.Count;$i++){
        $actual=$r.Stages[$i]
        Driver-Keys $actual @('Stage','Case','Commands') RecorderStage
        Driver-Need ($actual.Stage -ceq $stages[$i].Stage -and $actual.Case -ceq $stages[$i].Case -and $actual.Commands -is [array] -and $actual.Commands.Count -eq 4) 'Recorder exact stage identity/order'
        $names=@('New-ODRecordStream','Write-ODCaptureRecord','Write-ODProfileIntentRecord','Capture-Failure')
        for($j=0;$j -lt 4;$j++){
            $command=$actual.Commands[$j];Driver-Keys $command $Oracle.Recorder.CommandKeys RecorderCommand
            Driver-Need ($command.Name -ceq $names[$j] -and $command.SameOriginal -is [bool] -and $command.SameOriginal -and $command.ModuleName -is [string] -and $command.ModuleName -ceq '') 'Actual outer references isolated'
        }
    }
    $moduleNames=[Collections.Generic.List[string]]::new()
    for($i=0;$i -lt $expectedCases;$i++){
        $c=$r.Cases[$i];$case=$Oracle.WriterCases[$i];$f=$Oracle.Grants[$CurrentGrant].Groups[$i].Facts
        Driver-Keys $c @('Boundary','Case','ModuleName','Exports','State') RecorderCase
        Driver-Need ($c.Boundary -ceq $case.Boundary -and $c.Case -ceq $case.Case -and $c.ModuleName -is [string] -and $c.ModuleName -cmatch '^__DynamicModule_[a-f0-9-]{36}$') 'Case module role identity'
        $moduleNames.Add($c.ModuleName)
        Driver-Keys $c.Exports $Oracle.Recorder.ExportsKeys NoExports
        foreach($key in $Oracle.Recorder.ExportsKeys){Driver-Need ($c.Exports[$key] -is [long] -and $c.Exports[$key] -eq 0) 'Case exports must be zero'}
        Driver-Keys $c.State @('Initialized','PrimaryNull','SecondaryCount','Overflow') CaseState
        Driver-Need ($c.State.Initialized -is [bool] -and $c.State.Initialized -and $c.State.PrimaryNull -is [bool] -and $c.State.PrimaryNull -eq ($case.Case -ceq 'SuccessOnce') -and $c.State.SecondaryCount -eq $f.SecondaryCount -and $c.State.Overflow -is [bool] -and -not $c.State.Overflow) 'Explicit isolated case state'
    }
    Driver-Need (@($moduleNames.ToArray()|Select-Object -Unique).Count -eq $expectedCases) 'Distinct private case modules'
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
    Driver-Need ($e.Schema -eq 1 -and $e.Grant -ceq $CurrentGrant -and $e.Status -ceq $Profile.Status -and -not $e.Overflow -and $null -eq $e.PrimaryError -and $e.SecondaryErrors.Count -eq 0 -and $null -eq $e.CurrentGroup -and $null -eq $e.Probe) 'Capture status/controller fault mismatch'
    Driver-Need ((Driver-EqualJSON $e.Groups $Profile.Groups) -and (Driver-EqualJSON $e.NotRun $Profile.NotRun)) 'Ordered groups/NotRun mismatch'
    Driver-FalseFlags $e
    Driver-Need ($e.Observations -is [array] -and $e.Observations.Count -le 32) 'Capture observation bounds'
    Driver-Recorder $e $Oracle $CurrentGrant $Manifest
    if($CurrentGrant -cin @('writer-red','writer-green','contract-red','contract-green')){$unit=$e.Clone();$unit.Observations=@($e.Observations | Where-Object {$_.Name -cne 'Recorder'});Driver-UnitEvidence $unit $Oracle $CurrentGrant;return 0}
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
            Driver-ErrorLocation $actual.Error @('scripts/offline-diagnostics-v3/diagnose-offline.ps1','docs/diagnostics/isolated-fuzz/p2b-integrated-review-repair/entry.tests.ps1')
            Driver-Need ($actual.Name -ceq $want.Name -and -not $actual.RuntimeLoaded -and $actual.Error.Type -ceq $want.Type -and $actual.Error.ErrorId -cin $want.Ids -and $actual.Error.Message -ceq $want.Message) 'Fixed forbidden-request actual signature'
        }
        $tree=@($e.Observations | Where-Object {$_.Name -ceq 'EntryTree'})
        Driver-Need ($tree.Count -eq 1 -and (Driver-EqualJSON $tree[0].Value.Before $tree[0].Value.After) -and (Driver-EqualJSON $tree[0].Value.Roots $Profile.ExistingProfiles)) 'Inspect/rejections changed managed profile tree'
    }
    $started=[Collections.Generic.List[object]]::new()
    if($expectedReturns -eq 1){
        $r=$profileReturns[0].Value
        Driver-Keys $r $Oracle.ProfileRootNames ProfileReport
        Driver-Need ($r.Schema -eq 1 -and $r.Profile -ceq $Profile.Profile -and $r.Status -ceq 'COMPLETE' -and $r.Closed -and $null -eq $r.PrimaryError -and $r.SecondaryErrors.Count -eq 0 -and -not $r.Overflow -and $r.FrozenSHA256 -ceq $KnownFreeze) 'Private runtime report status mismatch'
        Driver-FalseFlags $r
        Driver-Need ($r.EntryPath -ceq [IO.Path]::GetFullPath((Join-Path $repo 'scripts/offline-diagnostics-v3/diagnose-offline.ps1')) -and $r.KernelPath -ceq [IO.Path]::GetFullPath((Join-Path $repo 'scripts/offline-diagnostics-v3/private/session.psm1')) -and $r.EntrySHA256 -ceq (Driver-PinHash $Manifest 'scripts/offline-diagnostics-v3/diagnose-offline.ps1') -and $r.KernelSHA256 -ceq (Driver-PinHash $Manifest 'scripts/offline-diagnostics-v3/private/session.psm1')) 'Actual source identity differs from independently frozen pins'
        $offset=6
        $count=$Profile.Groups.Count-$offset-5
        $expectedGroups=@($Profile.Groups | Select-Object -Skip $offset -First $count)
        Driver-Need (Driver-EqualJSON $r.Groups $expectedGroups) 'Profile ordered group projection mismatch'
        Driver-Need ($r.Observations -is [array] -and $r.Observations.Count -le 32) 'Profile observation bounds'
        Assert-ODProfileObservations $r.Observations $Profile.Profile ([ordered]@{Repo=$repo;PacketRoot=$packet;ProfileRoot=[IO.Path]::GetFullPath((Join-Path $packet ('profiles/'+$Profile.Profile)));Oracle=$Oracle;Sources=[ordered]@{Sinks=(Join-Path $repo 'scripts/offline-diagnostics-v3/private/sinks.ps1');Core=(Join-Path $repo 'docs/diagnostics/isolated-fuzz/p2b-integrated-review-repair/core.tests.ps1')}})
        foreach($detail in $r.Observations){
            Driver-Keys $detail @('Name','Value') ProfileObservation
            if($detail.Name -ceq 'Step'){$owned=Driver-Step $detail.Value $Oracle $Profile.Profile;if($null -ne $owned){$started.Add($owned)}}
        }
        $reentry=@($e.Observations | Where-Object {$_.Name -ceq 'ProfileReentry'})
        Driver-Need ($reentry.Count -eq 1) 'Single actual public reentry required'
        $v=$reentry[0].Value
        Driver-ErrorLocation $v.Error @('scripts/offline-diagnostics-v3/private/profiles.ps1','scripts/offline-diagnostics-v3/private/session.psm1')
        Driver-Need ($v.SameModule -and (Driver-EqualJSON $v.Before $v.After) -and $v.Error.Type -ceq 'System.Management.Automation.RuntimeException' -and $v.Error.ErrorId -ceq 'OD: profile already invoked' -and $v.Error.Message -ceq 'OD: profile already invoked') 'Actual reentry guard mismatch'
    }
    Driver-Need ($started.Count -eq $Profile.Children) 'Exact direct child count mismatch'
    if($started.Count -gt 0){Driver-Need (@($started.PID | Select-Object -Unique).Count -eq $started.Count) 'Distinct owned direct targets required'}
    return $started.Count
}
$preparationException=$false
if(-not [string]::IsNullOrEmpty($PreparationAdjustment)){
    Driver-Need ($PreparationAdjustment -ceq 'offline-prepare-recording-adjustment-2026-11-01-1' -and $Action -ceq 'Prepare' -and $Grant -ceq 'seed') 'Adjustment is preparation-only'
    $adjustment=Join-Path $repo '.superpowers/sdd/2026-11-01-offline-prepare-recording-adjustment'
    function Driver-AdjustmentPath([string]$Path){
        $cursor=[IO.Path]::GetFullPath($Path)
        Driver-Need ($cursor.StartsWith($repo+[IO.Path]::DirectorySeparatorChar,[StringComparison]::OrdinalIgnoreCase)) 'Adjustment fixed repository path'
        while($cursor -ine $repo){
            $item=Get-Item -LiteralPath $cursor -Force
            Driver-Need (($item.Attributes -band [IO.FileAttributes]::ReparsePoint) -eq 0) 'Adjustment linked path refused'
            $cursor=[IO.Path]::GetDirectoryName($cursor)
        }
    }
    $adjustmentPlan=Join-Path $repo 'docs/superpowers/plans/2026-11-01-offline-prepare-recording-adjustment.md'
    $adjustmentDesign=Join-Path $repo 'docs/superpowers/specs/2026-11-01-offline-prepare-recording-adjustment-design.md'
    foreach($path in @($packet,$adjustment,$adjustmentPlan,$adjustmentDesign,$PSCommandPath)){Driver-AdjustmentPath $path}
    Driver-CheckHash $adjustmentPlan '5F070621AD11402B4ABBFD71EFA86D39E847130694D6817CEDE6340649BBD2BD'
    Driver-CheckHash $adjustmentDesign '6F57E7DAA71C5B3E6E193B8C46BD643B524D8ACFE2A945EF166FB1F4A06298E5'
    $retained=[ordered]@{
        'progress.md'='55B652081DA45F3F891D14C1DE95D634A3BA5872B53F9C6D60ECF59C84061E63'
        'seed.prepare.tool-result.json'='C6EDB04CFD2C1E78BE02E497A48DBA4A8E4F74CB4E95EB83CCDAD700F8903719'
        'static.tool-result.json'='69620AFC7571CB0D83F10812A405D7044A5F0135FEF628B9F2D40512A68441EA'
        'stop.json'='0D4FE34A03FAD3FD1D45B8C4B850C5B6D8C8BF9F3C9283304503B7119FD7EDE2'
        'author-integrity.tool-result.json'='8F091601FAD7F9603D603F786458B7EB4EFAFD0952A65FF38A82FE2237994E02'
    }
    $existing=@(Get-ChildItem -LiteralPath $packet -Force)
    Driver-Need ($existing.Count -eq 5 -and @($existing|Where-Object {$_.PSIsContainer -or $_.Name -cnotin @($retained.Keys)}).Count -eq 0) 'Exact retained five-file inventory'
    foreach($name in $retained.Keys){$path=Join-Path $packet $name;Driver-AdjustmentPath $path;Driver-CheckHash $path $retained[$name]}
    $stoppedPortable=Join-Path $PSScriptRoot portable-stopped.json
    Driver-AdjustmentPath $stoppedPortable
    Driver-CheckHash $stoppedPortable '1D10ABA536D414B06BE67422EA3662EF68AF4926858A70EBBBCFAB38E1EE6ADF'
    $archive=Get-Content -LiteralPath $stoppedPortable -Raw|ConvertFrom-Json
    foreach($record in $archive.Artifacts){
        $path=Join-Path $packet $record.Path
        Driver-Need ($record.Path -cin @($retained.Keys) -and [Convert]::ToBase64String([IO.File]::ReadAllBytes($path)) -ceq $record.Base64) 'Retained original bytes'
    }
    foreach($name in @('approval.json','original-driver.bin','static.tool-result.json','attempt.intent.json')){Driver-AdjustmentPath (Join-Path $adjustment $name)}
    Driver-CheckHash (Join-Path $adjustment approval.json) '7480924FD798E39D35E26EC27E0D1BF4D0E43CB16CD1CE632AF73FC85B84AA79'
    Driver-CheckHash (Join-Path $adjustment original-driver.bin) '4F75FBDD23CD6A21C3567460A1CFC62661B5B00D8D5A77F8FE4300CFE1500EE9'
    $approval=Get-Content -LiteralPath (Join-Path $adjustment approval.json) -Raw|ConvertFrom-Json -AsHashtable
    Driver-Keys $approval @('Schema','Identity','DesignSHA256','PlanSHA256','OriginalDriverSHA256','ExecutionBaseline','ApprovalText','AdditionalPrepare','Hosts','Targets','Compilations','Reviews','RuntimeRemainder') AdjustmentApproval
    foreach($name in @('Schema','AdditionalPrepare')){Driver-Need ($approval[$name] -is [long] -and $approval[$name] -eq 1) 'Adjustment one preparation'}
    foreach($name in @('Hosts','Targets','Compilations','Reviews')){Driver-Need ($approval[$name] -is [long] -and $approval[$name] -eq 0) 'Adjustment zero runtime grants'}
    Driver-Need ($approval.Identity -ceq $PreparationAdjustment -and $approval.ApprovalText -ceq '是' -and $approval.ExecutionBaseline -ceq '73aa96037cf1908ec47cd1d789349b9e602a15ab' -and $approval.RuntimeRemainder -ceq 'FROZEN') 'Actual adjustment approval'
    $staticOriginal=(Get-Content -LiteralPath (Join-Path $packet static.tool-result.json) -Raw|ConvertFrom-Json).waits[-1].output|ConvertFrom-Json
    Driver-Need ($staticOriginal.Sources.Count -eq 28) 'Original seed source inventory'
    foreach($source in $staticOriginal.Sources){
        $path=Join-Path $repo $source.Path;Driver-AdjustmentPath $path
        if($source.Path -cne 'docs/diagnostics/isolated-fuzz/p2b-integrated-review-repair/driver.ps1'){Driver-CheckHash $path $source.SHA256}
    }
    $intentPath=Join-Path $adjustment attempt.intent.json
    $intent=Get-Content -LiteralPath $intentPath -Raw|ConvertFrom-Json -AsHashtable
    Driver-Keys $intent @('Schema','Identity','ApprovalSHA256','PlanSHA256','RevisedDriverSHA256','Action','Grant','AttemptConsumed','RuntimeRemainder') AdjustmentIntent
    Driver-Need ($intent.Schema -is [long] -and $intent.Schema -eq 1 -and $intent.Identity -ceq $PreparationAdjustment -and $intent.ApprovalSHA256 -ceq (Driver-Hash (Join-Path $adjustment approval.json)) -and $intent.PlanSHA256 -ceq (Driver-Hash $adjustmentPlan) -and $intent.RevisedDriverSHA256 -ceq (Driver-Hash $PSCommandPath) -and $intent.Action -ceq 'Prepare' -and $intent.Grant -ceq 'seed' -and $intent.AttemptConsumed -is [bool] -and $intent.AttemptConsumed -and $intent.RuntimeRemainder -ceq 'FROZEN') 'Fixed consumed adjustment intent'
    foreach($name in @('seed.frozen.json','snapshots','writer-red.intent.json')){Driver-Need (-not (Test-Path -LiteralPath (Join-Path $packet $name))) 'Adjustment preparation output already exists'}
    Driver-JSON (Join-Path $adjustment attempt.entered.json) ([ordered]@{Identity=$PreparationAdjustment;IntentSHA256=(Driver-Hash $intentPath);AttemptConsumed=$true})
    $preparationException=$true
}else{
    Driver-Need (-not (Test-Path -LiteralPath (Join-Path $packet stop.json))) 'Execution stopped; no refreeze'
}
if($Action -ceq 'Prepare'){
    Driver-Need ($Grant -cin @('seed','writer-green','repair-green')) 'Only three approved Prepare phases'
    Driver-Static
    $manifestPath=Join-Path $packet ($Grant+'.frozen.json')
    Driver-Need (-not (Test-Path -LiteralPath $manifestPath)) 'Prepare attempted; no retry'
    if($Grant -ceq 'seed'){
        $existing=@(Get-ChildItem -LiteralPath $packet -Force)
        if($preparationException){
            Driver-Need ($existing.Count -eq 5 -and @($existing|Where-Object {$_.PSIsContainer -or $_.Name -cnotin @($retained.Keys)}).Count -eq 0) 'Exact retained five-file inventory'
            foreach($name in $retained.Keys){Driver-AdjustmentPath (Join-Path $packet $name);Driver-CheckHash (Join-Path $packet $name) $retained[$name]}
        }else{
        Driver-Need (@($existing|Where-Object {$_.PSIsContainer -or $_.Name -cnotin @('progress.md','static.tool-result.json')}).Count -eq 0) 'Only explicit ledger/static receipts before freeze'
        }
        $spec=Join-Path $repo 'docs/superpowers/specs/2026-10-19-offline-diagnostic-entry-migration-design.md'
        $plan=Join-Path $repo 'docs/superpowers/plans/2026-10-31-offline-integrated-review-repair.md'
        Driver-CheckHash $spec 'DA27DA1C6BD3578512FB60A7F38B6635EAC9ACEF6570534F8B868BD07D088D9A'
        Driver-CheckHash $plan '67A052C56F0C57AD94A2E2E6B79A81AA6969A61FDAA89736FA271C1BF75E6689'
        $oldPacket=Join-Path $repo '.superpowers/sdd/2026-10-31-offline-record-isolation-repair'
        $oldManifest=Join-Path $oldPacket isolation-green.frozen.json
        Driver-CheckHash $oldManifest '93B1D18EE6C4FF944FEB45AA8CE8E91C7011BDD3C0374FFA9A6587C2AAF1D447'
        $old=Get-Content -LiteralPath $oldManifest -Raw|ConvertFrom-Json
        Driver-Need ($old.Pins.Count -eq 2394) 'Inherited2394'
        $pins=[Collections.Generic.List[object]]::new()
        foreach($pin in $old.Pins){
            Driver-Need ($pin.Path -is [string] -and $pin.SHA256 -is [string] -and $pin.SHA256 -cmatch '^[A-F0-9]{64}$') 'Actual inherited identity'
            $pins.Add([pscustomobject][ordered]@{VerificationPath=[IO.Path]::GetFullPath($pin.Path);SHA256=$pin.SHA256;ExpectedState='Present';Role='inherited-isolation'})
        }
        Driver-Verify ([pscustomobject]@{Pins=$pins.ToArray()})
        $portablePath=Join-Path $repo 'docs/diagnostics/isolated-fuzz/p2b-record-isolation-repair/portable-final.json'
        Driver-CheckHash $portablePath '83FE6E002281837B4E841F334A591B6D936EE2D90D107441529AF4D123DFFD2D'
        $portable=Get-Content -LiteralPath $portablePath -Raw|ConvertFrom-Json
        Driver-Need ($portable.Artifacts.Count -eq 47 -and $portable.Directories.Count -eq 6) 'Historical portable inventory'
        foreach($a in $portable.Artifacts){
            Driver-Need (-not [IO.Path]::IsPathRooted($a.Path) -and $a.Path -notmatch '(^|[\\/])\.\.([\\/]|$)' -and $a.Encoding -ceq 'base64') 'Portable fixed path'
            $original=Join-Path $oldPacket $a.Path;$bytes=[Convert]::FromBase64String($a.Data)
            Driver-Need ($bytes.Length -eq $a.Length -and [IO.File]::Exists($original)) 'Portable original length'
            Driver-CheckHash $original $a.SHA256
            Driver-Need ([Linq.Enumerable]::SequenceEqual([byte[]][IO.File]::ReadAllBytes($original),[byte[]]$bytes)) 'Portable original bytes'
        }
        Driver-Once (Join-Path $packet approved-spec.snapshot.md) ([IO.File]::ReadAllBytes($spec))
        Driver-Once (Join-Path $packet approved-plan.snapshot.md) ([IO.File]::ReadAllBytes($plan))
        $raw=@(Get-ChildItem -LiteralPath $oldPacket -Recurse -File|Sort-Object FullName)
        $leaf=@(Get-ChildItem -LiteralPath (Join-Path $repo 'docs/diagnostics/isolated-fuzz/p2b-record-isolation-repair') -File -Force|Sort-Object Name)
        Driver-Need ($raw.Count -eq 47 -and $leaf.Count -eq 9 -and $inputPaths.Count -eq 28) 'Explicit inventory'
        $report=Join-Path $repo 'docs/roadmap/2026-10-31-offline-record-isolation-repair.md'
        $paths=@($raw.FullName)+@($leaf.FullName)+@($report,$plan,(Join-Path $packet approved-spec.snapshot.md),(Join-Path $packet approved-plan.snapshot.md))+@($inputPaths|ForEach-Object {Join-Path $repo $_})
        Driver-Need ($paths.Count -eq 88) 'Exact88new dependencies'
        foreach($path in $paths){$pins.Add([pscustomobject][ordered]@{VerificationPath=[IO.Path]::GetFullPath($path);SHA256=(Driver-Hash $path);ExpectedState='Present';Role='integrated-review-repair-dependency'})}
        $manifest=[pscustomobject][ordered]@{Schema=2;Phase=$Grant;BaseCommit='9b6eaf19bdecfdee0919b62c996a90f9a7c43373';PlanSHA256='67A052C56F0C57AD94A2E2E6B79A81AA6969A61FDAA89736FA271C1BF75E6689';InheritedPins=2394;Pins=$pins.ToArray();ChangedInputs=[string[]]@();MaxHosts=7;MaxChildren=7;MaxCompilations=2;MaxReviewers=1;SourceVerified=$false;LiveActionsPerformed=$false;RuntimeApproved=$false;FormalAcceptance=$false;RealGateImplemented=$false}
    }else{
        Driver-Need (-not [string]::IsNullOrEmpty($PreviousFrozenSHA256)) 'Previous identity required'
        $previous=if($Grant -ceq 'writer-green'){'seed'}else{'writer-green'}
        $previousPath=Join-Path $packet ($previous+'.frozen.json');Driver-CheckHash $previousPath $PreviousFrozenSHA256
        $manifest=Get-Content -LiteralPath $previousPath -Raw|ConvertFrom-Json
        $allowed=if($Grant -ceq 'writer-green'){'scripts/offline-diagnostics-v3/private/write-boundary.ps1'}else{'docs/diagnostics/isolated-fuzz/p2b-integrated-review-repair/observation-contract.ps1'}
        $changed=[Collections.Generic.List[string]]::new()
        foreach($pin in $manifest.Pins){
            $relative=[IO.Path]::GetRelativePath($repo,$pin.VerificationPath).Replace('\','/')
            if($relative -ceq $allowed){
                $hash=Driver-Hash $pin.VerificationPath;Driver-Need ($pin.ExpectedState -ceq 'Present' -and $pin.SHA256 -cne $hash) 'One specified source must change'
                $pin.SHA256=$hash;$changed.Add($relative)
            }else{Driver-CheckHash $pin.VerificationPath $pin.SHA256}
        }
        Driver-Need ($changed.Count -eq 1) 'Exact one-source window'
        $manifest.Phase=$Grant;$manifest.ChangedInputs=$changed.ToArray()
    }
    Driver-Need ($manifest.Pins.Count -eq 2482) 'Exact2317'
    Driver-Verify $manifest
    $base=Join-Path $packet ('snapshots/'+$Grant);Driver-Need (-not (Test-Path -LiteralPath $base)) 'Snapshot attempted'
    foreach($relative in $inputPaths){$target=Join-Path $base $relative;$null=[IO.Directory]::CreateDirectory([IO.Path]::GetDirectoryName($target));Driver-Once $target ([IO.File]::ReadAllBytes((Join-Path $repo $relative)))}
    Driver-JSON $manifestPath $manifest
    [ordered]@{Prepared=$Grant;FrozenSHA256=(Driver-Hash $manifestPath);Pins=2482;Snapshots=28;ChangedInputs=$manifest.ChangedInputs}|ConvertTo-Json
    return
}
Driver-Need (-not (Test-Path -LiteralPath (Join-Path $packet stop.json))) 'Execution stopped; no retry'
Driver-Need (-not [string]::IsNullOrEmpty($FrozenSHA256)) 'Missing frozen identity'
Driver-Need ($Grant -cin @('writer-red','contract-red','writer-green','contract-green','core','lifecycle','continuation')) 'Approved RunOne only'
$phase=switch -CaseSensitive ($Grant){'writer-red'{'seed'}'contract-red'{'seed'}'writer-green'{'writer-green'}default{'repair-green'}}
$manifestPath=Join-Path $packet ($phase+'.frozen.json')
Driver-CheckHash $manifestPath $FrozenSHA256
$manifest=Get-Content -LiteralPath $manifestPath -Raw | ConvertFrom-Json
Driver-Need ($manifest.Schema -eq 2 -and $manifest.Pins.Count -eq 2482) 'Exact2317mapping state-aware manifest'
Driver-Verify $manifest
$oracle=Get-Content -LiteralPath (Join-Path $PSScriptRoot expected-output.json) -Raw | ConvertFrom-Json -AsHashtable
$intentPath=Join-Path $packet ($Grant+'.intent.json');$finalPath=Join-Path $packet ($Grant+'.final.json')
foreach($suffix in @('intent.json','final.json','stdout.bin','stderr.bin','evidence.json')){Driver-Need (-not (Test-Path -LiteralPath (Join-Path $packet ($Grant+'.'+$suffix)))) 'Grant attempted/unknown; no retry'}
Driver-JSON $intentPath ([ordered]@{Grant=$Grant;AttemptConsumed=$true;FrozenSHA256=$FrozenSHA256;AuthorizedHost=1;AuthorizedChildren=$oracle.Grants[$Grant].Children})
$row=[ordered]@{Grant=$Grant;Started=$false;PID=$null;OwnedExit=$false;StreamsComplete=$false;ExitCode=$null;TimedOut=$false;TerminationRequested=$false;PrimaryError=$null;SecondaryErrors=@();Overflow=$false;PostPinsIntact=$false;Exact=$false;StdoutSHA256=$null;StderrSHA256=$null;EvidenceSHA256=$null;EvidenceValid=$false;ChildrenObserved=0;EvidenceState='Unknown';ExpectedNegative=($Grant -cin @('writer-red','contract-red'));ExpectedOutputSHA256=$null;Rescue=$false;SourceVerified=$false;LiveActionsPerformed=$false;RuntimeApproved=$false;FormalAcceptance=$false;RealGateImplemented=$false}
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
        if($stdout.Replace([Environment]::NewLine,([string][char]10)) -cne $expected.Replace([Environment]::NewLine,([string][char]10)) -or $stderr.Replace([Environment]::NewLine,([string][char]10)) -cne $expectedError.Replace([Environment]::NewLine,([string][char]10))){throw 'Exact stdout/stderr mismatch'}
        Driver-Need ((Get-Item -LiteralPath $evidencePath).Length -le 1048576) 'Evidence byte bound'
        $actual=Get-Content -LiteralPath $evidencePath -Raw | ConvertFrom-Json -AsHashtable
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
    $stopRoot=Join-Path ([IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))) '.superpowers/sdd/2026-10-31-offline-integrated-review-repair'
    if((Test-Path -LiteralPath $stopRoot) -and -not (Test-Path -LiteralPath (Join-Path $stopRoot stop.json))){
        Driver-JSON (Join-Path $stopRoot stop.json) ([ordered]@{Action=$Action;Grant=$Grant;Message=$actual.Exception.Message;ErrorId=$actual.FullyQualifiedErrorId;Position=$actual.InvocationInfo.PositionMessage;Stack=$actual.ScriptStackTrace;Remainder='FROZEN'})
    }
    throw
}
