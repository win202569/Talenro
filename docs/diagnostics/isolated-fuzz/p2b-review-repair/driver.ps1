param(
    [Parameter(Mandatory)][ValidateSet('Prepare','RunOne')][string]$Action,
    [Parameter(Mandatory)][ValidateSet('seed','writer-red','contract-red','writer-green','contract-green','repair-green','core','lifecycle','continuation')][string]$Grant,
    [ValidatePattern('^[A-F0-9]{64}$')][string]$FrozenSHA256,
    [ValidatePattern('^[A-F0-9]{64}$')][string]$PreviousFrozenSHA256
)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$repo=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))
$packet=Join-Path $repo '.superpowers/sdd/2026-10-31-offline-diagnostic-review-repair'
$engine='C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe'
$inputPaths=@('docs/diagnostics/isolated-fuzz/p2b-review-repair/.gitattributes','docs/diagnostics/isolated-fuzz/p2b-review-repair/README.md','docs/diagnostics/isolated-fuzz/p2b-review-repair/expected-output.json','docs/diagnostics/isolated-fuzz/p2b-review-repair/test-helpers.ps1','docs/diagnostics/isolated-fuzz/p2b-review-repair/entry.tests.ps1','docs/diagnostics/isolated-fuzz/p2b-review-repair/core.tests.ps1','docs/diagnostics/isolated-fuzz/p2b-review-repair/prefix.tests.ps1','docs/diagnostics/isolated-fuzz/p2b-review-repair/owned.tests.ps1','docs/diagnostics/isolated-fuzz/p2b-review-repair/live.tests.ps1','docs/diagnostics/isolated-fuzz/p2b-review-repair/capture.ps1','docs/diagnostics/isolated-fuzz/p2b-review-repair/driver.ps1','scripts/offline-diagnostics-v2/diagnose-offline.ps1','scripts/offline-diagnostics-v2/private/catalog.json','scripts/offline-diagnostics-v2/private/session.psm1','scripts/offline-diagnostics-v2/private/sinks.ps1','scripts/offline-diagnostics-v2/private/owned-step.ps1','scripts/offline-diagnostics-v2/private/sequence.ps1','scripts/offline-diagnostics-v2/private/profiles.ps1','scripts/offline-diagnostics-v2/private/observations.ps1','scripts/offline-diagnostics-v2/.gitattributes','scripts/offline-diagnostics-v2/private/write-boundary.ps1','docs/diagnostics/isolated-fuzz/p2b-review-repair/observation-contract.ps1','docs/diagnostics/isolated-fuzz/p2b-review-repair/fault-stream.cs','docs/diagnostics/isolated-fuzz/p2b-review-repair/record-boundary.tests.ps1','docs/diagnostics/isolated-fuzz/p2b-review-repair/observation-contract.tests.ps1')
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
    foreach($name in @('.gitattributes','README.md','expected-output.json','test-helpers.ps1','entry.tests.ps1','core.tests.ps1','prefix.tests.ps1','owned.tests.ps1','live.tests.ps1','capture.ps1','driver.ps1','record-boundary.tests.ps1','observation-contract.tests.ps1','fault-stream.cs','observation-contract.ps1')){
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
        Driver-ErrorLocation $r.Error @('scripts/offline-diagnostics-v2/private/sinks.ps1')
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
                    Driver-UnitError $errorRecord $errorRecord.ErrorId $message 'scripts/offline-diagnostics-v2/private/write-boundary.ps1'
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
                Driver-ErrorLocation $v.Error @('docs/diagnostics/isolated-fuzz/p2b-review-repair/observation-contract.ps1')
            }else{Driver-Need ($null -eq $v.Error) 'No unexpected unit exception'}
        }
    }
}

. (Join-Path $PSScriptRoot observation-contract.ps1)
function Driver-Evidence($Evidence,$Profile,$Oracle,[string]$CurrentGrant,$Manifest,[string]$KnownFreeze){
    $e=$Evidence
    Driver-Keys $e $Oracle.EvidenceRootNames Evidence
    Driver-Need ($e.Schema -eq 1 -and $e.Grant -ceq $CurrentGrant -and $e.Status -ceq $Profile.Status -and -not $e.Overflow -and $null -eq $e.PrimaryError -and $e.SecondaryErrors.Count -eq 0 -and $null -eq $e.CurrentGroup -and $null -eq $e.Probe) 'Capture status/controller fault mismatch'
    Driver-Need ((Driver-EqualJSON $e.Groups $Profile.Groups) -and (Driver-EqualJSON $e.NotRun $Profile.NotRun)) 'Ordered groups/NotRun mismatch'
    Driver-FalseFlags $e
    Driver-Need ($e.Observations -is [array] -and $e.Observations.Count -le 32) 'Capture observation bounds'
    if($CurrentGrant -cin @('writer-red','writer-green','contract-red','contract-green')){Driver-UnitEvidence $e $Oracle $CurrentGrant;return 0}
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
            Driver-ErrorLocation $actual.Error @('scripts/offline-diagnostics-v2/diagnose-offline.ps1','docs/diagnostics/isolated-fuzz/p2b-review-repair/entry.tests.ps1')
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
        Driver-Need ($r.EntryPath -ceq [IO.Path]::GetFullPath((Join-Path $repo 'scripts/offline-diagnostics-v2/diagnose-offline.ps1')) -and $r.KernelPath -ceq [IO.Path]::GetFullPath((Join-Path $repo 'scripts/offline-diagnostics-v2/private/session.psm1')) -and $r.EntrySHA256 -ceq (Driver-PinHash $Manifest 'scripts/offline-diagnostics-v2/diagnose-offline.ps1') -and $r.KernelSHA256 -ceq (Driver-PinHash $Manifest 'scripts/offline-diagnostics-v2/private/session.psm1')) 'Actual source identity differs from independently frozen pins'
        $offset=6
        $count=$Profile.Groups.Count-$offset-2
        $expectedGroups=@($Profile.Groups | Select-Object -Skip $offset -First $count)
        Driver-Need (Driver-EqualJSON $r.Groups $expectedGroups) 'Profile ordered group projection mismatch'
        Driver-Need ($r.Observations -is [array] -and $r.Observations.Count -le 32) 'Profile observation bounds'
        Assert-ODProfileObservations $r.Observations $Profile.Profile ([ordered]@{Repo=$repo;PacketRoot=$packet;ProfileRoot=[IO.Path]::GetFullPath((Join-Path $packet ('profiles/'+$Profile.Profile)));Oracle=$Oracle;Sources=[ordered]@{Sinks=(Join-Path $repo 'scripts/offline-diagnostics-v2/private/sinks.ps1');Core=(Join-Path $repo 'docs/diagnostics/isolated-fuzz/p2b-review-repair/core.tests.ps1')}})
        foreach($detail in $r.Observations){
            Driver-Keys $detail @('Name','Value') ProfileObservation
            if($detail.Name -ceq 'Step'){$owned=Driver-Step $detail.Value $Oracle $Profile.Profile;if($null -ne $owned){$started.Add($owned)}}
        }
        $reentry=@($e.Observations | Where-Object {$_.Name -ceq 'ProfileReentry'})
        Driver-Need ($reentry.Count -eq 1) 'Single actual public reentry required'
        $v=$reentry[0].Value
        Driver-ErrorLocation $v.Error @('scripts/offline-diagnostics-v2/private/profiles.ps1','scripts/offline-diagnostics-v2/private/session.psm1')
        Driver-Need ($v.SameModule -and (Driver-EqualJSON $v.Before $v.After) -and $v.Error.Type -ceq 'System.Management.Automation.RuntimeException' -and $v.Error.ErrorId -ceq 'OD: profile already invoked' -and $v.Error.Message -ceq 'OD: profile already invoked') 'Actual reentry guard mismatch'
    }
    Driver-Need ($started.Count -eq $Profile.Children) 'Exact direct child count mismatch'
    if($started.Count -gt 0){Driver-Need (@($started.PID | Select-Object -Unique).Count -eq $started.Count) 'Distinct owned direct targets required'}
    return $started.Count
}
if($Action -ceq 'Prepare'){
    Driver-Need ($Grant -cin @('seed','writer-green','repair-green')) 'Only three approved Prepare phases'
    Driver-Static
    $manifestPath=Join-Path $packet ($Grant+'.frozen.json')
    Driver-Need (-not (Test-Path -LiteralPath $manifestPath)) 'Prepare attempted; no retry'
    if($Grant -ceq 'seed'){
        $existing=@(Get-ChildItem -LiteralPath $packet -Force)
        Driver-Need (@($existing|Where-Object {$_.Name -cne 'progress.md' -or $_.PSIsContainer}).Count -eq 0) 'New packet ledger only'
        $spec=Join-Path $repo 'docs/superpowers/specs/2026-10-19-offline-diagnostic-entry-migration-design.md'
        $plan=Join-Path $repo 'docs/superpowers/plans/2026-10-31-offline-diagnostic-review-repair.md'
        Driver-CheckHash $spec 'DA27DA1C6BD3578512FB60A7F38B6635EAC9ACEF6570534F8B868BD07D088D9A'
        Driver-CheckHash $plan 'DE28CCC3233B99F703F931FA5DAB013A88EF874BDBC6D322744CC2730D562C62'
        $oldPacket=Join-Path $repo '.superpowers/sdd/2026-10-31-offline-diagnostic-runtime-migration'
        $oldManifest=Join-Path $oldPacket runtime-green.frozen.json
        Driver-CheckHash $oldManifest 'C3E24036ED56FFC89C8742418342D02A9B447B8C8643B0F8F1860D0DCAEF3175'
        $old=Get-Content -LiteralPath $oldManifest -Raw|ConvertFrom-Json
        Driver-Need ($old.Pins.Count -eq 2129) 'Inherited2129'
        $pins=[Collections.Generic.List[object]]::new()
        foreach($pin in $old.Pins){
            $path=$pin.PSObject.Properties['VerificationPath'];$hash=$pin.PSObject.Properties['SHA256'];$state=$pin.PSObject.Properties['ExpectedState'];$role=$pin.PSObject.Properties['Role']
            Driver-Need ($null -ne $path -and $null -ne $hash -and $null -ne $state) 'Required pin identity'
            Driver-Need ($path.Value -is [string] -and -not [string]::IsNullOrWhiteSpace($path.Value) -and $hash.Value -is [string] -and $hash.Value -cmatch '^[A-F0-9]{64}$' -and $state.Value -ceq 'Present') 'Valid inherited Present identity'
            $pins.Add([pscustomobject][ordered]@{VerificationPath=[IO.Path]::GetFullPath($path.Value);SHA256=$hash.Value;ExpectedState=$state.Value;Role=$(if($null -eq $role){$null}else{$role.Value})})
        }
        Driver-Verify ([pscustomobject]@{Pins=$pins.ToArray()})
        $portablePath=Join-Path $repo 'docs/diagnostics/isolated-fuzz/p2b-entry-runtime-migration/portable-final.json'
        Driver-CheckHash $portablePath '01889C7A479EFD418D19BF2D5B2E8B279077139006E1FB5E923706C4A99A300F'
        $portable=Get-Content -LiteralPath $portablePath -Raw|ConvertFrom-Json
        Driver-Need ($portable.Artifacts.Count -eq 142 -and $portable.Directories.Count -eq 49) 'Historical portable inventory'
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
        $leaf=@(Get-ChildItem -LiteralPath (Join-Path $repo 'docs/diagnostics/isolated-fuzz/p2b-entry-runtime-migration') -File|Sort-Object Name)
        Driver-Need ($raw.Count -eq 142 -and $leaf.Count -eq 15 -and $inputPaths.Count -eq 25) 'Explicit inventory'
        $reports=@('docs/roadmap/2026-10-31-offline-diagnostic-runtime-migration.md','docs/roadmap/2026-10-31-offline-diagnostic-runtime-review.md','docs/runbooks/offline-diagnostics.md')
        $hashes=@('F68CF2B1B093B412F14E2FC56007A437EC9A754E8425642F603B526390B2C798','D6F3DB4A607BE202C88C6D7664050DEB2346C6FB435EBCF8ED41D4D813A6A2BD','1526710D06AA99C76A5595C5074AA7CBD917C36EE5D2A6DC0EB9ADB1D13B9B72')
        for($i=0;$i -lt 3;$i++){Driver-CheckHash (Join-Path $repo $reports[$i]) $hashes[$i]}
        $paths=@($raw.FullName)+@($leaf.FullName)+@($reports|ForEach-Object {Join-Path $repo $_})+@($plan,(Join-Path $packet approved-spec.snapshot.md),(Join-Path $packet approved-plan.snapshot.md))+@($inputPaths|ForEach-Object {Join-Path $repo $_})
        Driver-Need ($paths.Count -eq 188) 'Exact188 new dependencies'
        foreach($path in $paths){$pins.Add([pscustomobject][ordered]@{VerificationPath=[IO.Path]::GetFullPath($path);SHA256=(Driver-Hash $path);ExpectedState='Present';Role='review-repair-dependency'})}
        $manifest=[pscustomobject][ordered]@{Schema=2;Phase=$Grant;BaseCommit='e0ea493506b6af9417954ee437a3a8293b430e71';PlanSHA256='DE28CCC3233B99F703F931FA5DAB013A88EF874BDBC6D322744CC2730D562C62';InheritedPins=2129;Pins=$pins.ToArray();ChangedInputs=[string[]]@();MaxHosts=7;MaxChildren=7;MaxCompilations=2;MaxReviewers=1;SourceVerified=$false;LiveActionsPerformed=$false;RuntimeApproved=$false;FormalAcceptance=$false;RealGateImplemented=$false}
    }else{
        Driver-Need (-not [string]::IsNullOrEmpty($PreviousFrozenSHA256)) 'Previous identity required'
        $previous=if($Grant -ceq 'writer-green'){'seed'}else{'writer-green'}
        $previousPath=Join-Path $packet ($previous+'.frozen.json');Driver-CheckHash $previousPath $PreviousFrozenSHA256
        $manifest=Get-Content -LiteralPath $previousPath -Raw|ConvertFrom-Json
        $allowed=if($Grant -ceq 'writer-green'){'scripts/offline-diagnostics-v2/private/write-boundary.ps1'}else{'docs/diagnostics/isolated-fuzz/p2b-review-repair/observation-contract.ps1'}
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
    Driver-Need ($manifest.Pins.Count -eq 2317) 'Exact2317'
    Driver-Verify $manifest
    $base=Join-Path $packet ('snapshots/'+$Grant);Driver-Need (-not (Test-Path -LiteralPath $base)) 'Snapshot attempted'
    foreach($relative in $inputPaths){$target=Join-Path $base $relative;$null=[IO.Directory]::CreateDirectory([IO.Path]::GetDirectoryName($target));Driver-Once $target ([IO.File]::ReadAllBytes((Join-Path $repo $relative)))}
    Driver-JSON $manifestPath $manifest
    [ordered]@{Prepared=$Grant;FrozenSHA256=(Driver-Hash $manifestPath);Pins=2317;Snapshots=25;ChangedInputs=$manifest.ChangedInputs}|ConvertTo-Json
    return
}
Driver-Need (-not [string]::IsNullOrEmpty($FrozenSHA256)) 'Missing frozen identity'
Driver-Need ($Grant -cin @('writer-red','contract-red','writer-green','contract-green','core','lifecycle','continuation')) 'Approved RunOne only'
$phase=switch -CaseSensitive ($Grant){'writer-red'{'seed'}'contract-red'{'seed'}'writer-green'{'writer-green'}default{'repair-green'}}
$manifestPath=Join-Path $packet ($phase+'.frozen.json')
Driver-CheckHash $manifestPath $FrozenSHA256
$manifest=Get-Content -LiteralPath $manifestPath -Raw | ConvertFrom-Json
Driver-Need ($manifest.Schema -eq 2 -and $manifest.Pins.Count -eq 2317) 'Exact2317mapping state-aware manifest'
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
