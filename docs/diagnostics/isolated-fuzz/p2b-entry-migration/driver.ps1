param(
    [Parameter(Mandatory)][ValidateSet('Prepare','RunOne')][string]$Action,
    [Parameter(Mandatory)][ValidateSet('entry-red','entry-green','runtime-red','runtime-green','lifecycle','continuation')][string]$Grant,
    [ValidatePattern('^[A-F0-9]{64}$')][string]$FrozenSHA256,
    [ValidatePattern('^[A-F0-9]{64}$')][string]$PreviousFrozenSHA256
)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$repo=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))
$packet=Join-Path $repo '.superpowers/sdd/2026-10-19-offline-diagnostic-entry-migration'
$engine='C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe'
$inputPaths=@('scripts/.gitattributes','docs/diagnostics/isolated-fuzz/p2b-entry-migration/.gitattributes','docs/diagnostics/isolated-fuzz/p2b-entry-migration/README.md','docs/diagnostics/isolated-fuzz/p2b-entry-migration/expected-output.json','docs/diagnostics/isolated-fuzz/p2b-entry-migration/test-helpers.ps1','docs/diagnostics/isolated-fuzz/p2b-entry-migration/entry.tests.ps1','docs/diagnostics/isolated-fuzz/p2b-entry-migration/core.tests.ps1','docs/diagnostics/isolated-fuzz/p2b-entry-migration/prefix.tests.ps1','docs/diagnostics/isolated-fuzz/p2b-entry-migration/owned.tests.ps1','docs/diagnostics/isolated-fuzz/p2b-entry-migration/live.tests.ps1','docs/diagnostics/isolated-fuzz/p2b-entry-migration/capture.ps1','docs/diagnostics/isolated-fuzz/p2b-entry-migration/driver.ps1','scripts/diagnose-offline.ps1','scripts/private/offline-diagnostics/catalog.json','scripts/private/offline-diagnostics/session.psm1','scripts/private/offline-diagnostics/sinks.ps1','scripts/private/offline-diagnostics/owned-step.ps1','scripts/private/offline-diagnostics/sequence.ps1','scripts/private/offline-diagnostics/profiles.ps1','scripts/private/offline-diagnostics/observations.ps1')
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
    foreach($pin in $Manifest.Pins){
        if($pin.ExpectedState -ceq 'Absent'){
            if(Test-Path -LiteralPath $pin.VerificationPath){throw ('Expected absent: '+$pin.VerificationPath)}
            if($null -ne $pin.SHA256){throw 'Absent hash must be null'}
        }elseif($pin.ExpectedState -ceq 'Present'){Driver-CheckHash $pin.VerificationPath $pin.SHA256}
        else{throw 'Unknown pin state'}
    }
}
function Driver-Static {
    if(@(Get-ChildItem -LiteralPath $PSScriptRoot -File).Count -ne 11){throw 'Expected eleven frozen leaf inputs'}
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
    Driver-Need ($pins.Count -eq 1 -and $pins[0].ExpectedState -ceq 'Present') 'Unique operational source pin required'
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
        Driver-ErrorLocation $r.Error @('scripts/private/offline-diagnostics/sinks.ps1')
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
function Driver-Evidence($Evidence,$Profile,$Oracle,[string]$CurrentGrant,$Manifest,[string]$KnownFreeze){
    $e=$Evidence
    Driver-Keys $e $Oracle.EvidenceRootNames Evidence
    Driver-Need ($e.Schema -eq 1 -and $e.Grant -ceq $CurrentGrant -and $e.Status -ceq $Profile.Status -and -not $e.Overflow -and $null -eq $e.PrimaryError -and $e.SecondaryErrors.Count -eq 0 -and $null -eq $e.CurrentGroup -and $null -eq $e.Probe) 'Capture status/controller fault mismatch'
    Driver-Need ((Driver-EqualJSON $e.Groups $Profile.Groups) -and (Driver-EqualJSON $e.NotRun $Profile.NotRun)) 'Ordered groups/NotRun mismatch'
    Driver-FalseFlags $e
    Driver-Need ($e.Observations -is [array] -and $e.Observations.Count -le 32) 'Capture observation bounds'
    $profileReturns=@($e.Observations | Where-Object {$_.Name -ceq 'ProfileReturn'})
    $expectedReturns=if($CurrentGrant -cin @('runtime-green','lifecycle','continuation')){1}else{0}
    Driver-Need ($profileReturns.Count -eq $expectedReturns) 'Exactly one canonical profile return'
    foreach($row in $e.Observations){Driver-Keys $row @('Name','Value') Observation}
    if($CurrentGrant -ceq 'entry-red'){
        Driver-Need ($e.Observations.Count -eq 0) 'Missing entry RED cannot invoke code'
    }
    if($CurrentGrant -cin @('entry-green','lifecycle','continuation')){
        $inspects=@($e.Observations | Where-Object {$_.Name -cin @('InspectDefault','InspectExplicit')})
        Driver-Need ($inspects.Count -eq 2) 'Two actual Inspect calls required'
        $key=if($CurrentGrant -ceq 'entry-green'){'Missing'}else{'Present'}
        $expected=Driver-Expand $Oracle.InspectTemplates[$key] $Manifest $KnownFreeze
        foreach($row in $inspects){
            Driver-Need ((Driver-EqualJSON $row.Value.Report $expected) -and -not $row.Value.RuntimeLoaded -and (Driver-EqualJSON $row.Value.Before $row.Value.After) -and (Driver-EqualJSON $row.Value.Roots $Profile.ExistingProfiles)) 'Actual read-only Inspect mismatch'
        }
        $rejections=@($e.Observations | Where-Object {$_.Name -ceq 'RejectedRequest'})
        Driver-Need ($rejections.Count -eq 14) 'All actual forbidden requests required'
        for($i=0;$i -lt 14;$i++){
            $actual=$rejections[$i].Value;$want=$Oracle.Rejections[$i]
            Driver-ErrorLocation $actual.Error @('scripts/diagnose-offline.ps1','docs/diagnostics/isolated-fuzz/p2b-entry-migration/entry.tests.ps1')
            Driver-Need ($actual.Name -ceq $want.Name -and -not $actual.RuntimeLoaded -and $actual.Error.Type -ceq $want.Type -and $actual.Error.ErrorId -cin $want.Ids -and $actual.Error.Message -ceq $want.Message) 'Fixed forbidden-request actual signature'
        }
        $tree=@($e.Observations | Where-Object {$_.Name -ceq 'EntryTree'})
        Driver-Need ($tree.Count -eq 1 -and (Driver-EqualJSON $tree[0].Value.Before $tree[0].Value.After) -and (Driver-EqualJSON $tree[0].Value.Roots $Profile.ExistingProfiles)) 'Inspect/rejections changed managed profile tree'
    }
    if($CurrentGrant -ceq 'runtime-red'){
        Driver-Need ($e.Observations.Count -eq 1 -and $e.Observations[0].Name -ceq 'EntryInvocation') 'Missing backend RED actual invocation required'
        $v=$e.Observations[0].Value
        Driver-ErrorLocation $v.Error @('scripts/diagnose-offline.ps1')
        Driver-Need ($v.EntryPath -ceq [IO.Path]::GetFullPath((Join-Path $repo 'scripts/diagnose-offline.ps1')) -and $v.Profile -ceq 'entry-migration-core' -and -not $v.RuntimeLoaded -and $v.Error.Type -ceq 'System.Management.Automation.RuntimeException' -and $v.Error.ErrorId -ceq 'OD: backend unavailable' -and $v.Error.Message -ceq 'OD: backend unavailable') 'Exact actual missing backend signature'
    }
    $started=[Collections.Generic.List[object]]::new()
    if($expectedReturns -eq 1){
        $r=$profileReturns[0].Value
        Driver-Keys $r $Oracle.ProfileRootNames ProfileReport
        Driver-Need ($r.Schema -eq 1 -and $r.Profile -ceq $Profile.Profile -and $r.Status -ceq 'COMPLETE' -and $r.Closed -and $null -eq $r.PrimaryError -and $r.SecondaryErrors.Count -eq 0 -and -not $r.Overflow -and $r.FrozenSHA256 -ceq $KnownFreeze) 'Private runtime report status mismatch'
        Driver-FalseFlags $r
        Driver-Need ($r.EntryPath -ceq [IO.Path]::GetFullPath((Join-Path $repo 'scripts/diagnose-offline.ps1')) -and $r.KernelPath -ceq [IO.Path]::GetFullPath((Join-Path $repo 'scripts/private/offline-diagnostics/session.psm1')) -and $r.EntrySHA256 -ceq (Driver-PinHash $Manifest 'scripts/diagnose-offline.ps1') -and $r.KernelSHA256 -ceq (Driver-PinHash $Manifest 'scripts/private/offline-diagnostics/session.psm1')) 'Actual source identity differs from independently frozen pins'
        $offset=if($CurrentGrant -ceq 'runtime-green'){0}else{6}
        $count=$Profile.Groups.Count-$offset-2
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
        Driver-ErrorLocation $v.Error @('scripts/private/offline-diagnostics/profiles.ps1','scripts/private/offline-diagnostics/session.psm1')
        Driver-Need ($v.SameModule -and (Driver-EqualJSON $v.Before $v.After) -and $v.Error.Type -ceq 'System.Management.Automation.RuntimeException' -and $v.Error.ErrorId -ceq 'OD: profile already invoked' -and $v.Error.Message -ceq 'OD: profile already invoked') 'Actual reentry guard mismatch'
    }
    Driver-Need ($started.Count -eq $Profile.Children) 'Exact direct child count mismatch'
    if($started.Count -gt 0){Driver-Need (@($started.PID | Select-Object -Unique).Count -eq $started.Count) 'Distinct owned direct targets required'}
    return $started.Count
}
if($Action -ceq 'Prepare'){
    Driver-Need ($Grant -cin @('entry-red','entry-green','runtime-green')) 'Only three planned Prepare phases'
    Driver-Static
    $manifestPath=Join-Path $packet ($Grant+'.frozen.json')
    Driver-Need (-not (Test-Path -LiteralPath $manifestPath)) 'Prepare attempted; no retry'
    if($Grant -ceq 'entry-red'){
        $existing=@(Get-ChildItem -LiteralPath $packet -Force)
        Driver-Need (@($existing | Where-Object {$_.Name -cne 'progress.md' -or $_.PSIsContainer}).Count -eq 0) 'New packet must be pristine except ledger'
        $spec=Join-Path $repo 'docs/superpowers/specs/2026-10-19-offline-diagnostic-entry-migration-design.md'
        $plan=Join-Path $repo 'docs/superpowers/plans/2026-10-19-offline-diagnostic-entry-migration.md'
        Driver-CheckHash $spec 'DA27DA1C6BD3578512FB60A7F38B6635EAC9ACEF6570534F8B868BD07D088D9A'
        Driver-CheckHash $plan '062AAC17317B5F8FBE5E29A88A67ABF6E53D5BD4BC6CA8151AC04DE783F81205'
        Driver-Once (Join-Path $packet approved-spec.snapshot.md) ([IO.File]::ReadAllBytes($spec))
        Driver-Once (Join-Path $packet approved-plan.snapshot.md) ([IO.File]::ReadAllBytes($plan))
        $oldPacket=Join-Path $repo '.superpowers/sdd/2026-10-19-p2b-prestart-cause-repair'
        $oldManifest=Join-Path $oldPacket classification-green.frozen.json
        Driver-CheckHash $oldManifest '01B99203F29B6073843756E347054F8489F5825C325B6F4084DF9C8018A07123'
        $old=Get-Content -LiteralPath $oldManifest -Raw | ConvertFrom-Json
        Driver-Need ($old.Pins.Count -eq 1930) 'Inherited mapping count'
        $pins=[Collections.Generic.List[object]]::new()
        foreach($pin in $old.Pins){$pins.Add([pscustomobject][ordered]@{VerificationPath=[IO.Path]::GetFullPath($pin.VerificationPath);SHA256=$pin.SHA256;Role=$pin.Role;ExpectedState='Present'})}
        $raw=@(Get-ChildItem -LiteralPath $oldPacket -Recurse -File | Sort-Object FullName)
        $portable=@(Get-ChildItem -LiteralPath (Join-Path $repo 'docs/diagnostics/isolated-fuzz/p2b-prestart-cause-repair') -File | Sort-Object Name)
        Driver-Need ($raw.Count -eq 92 -and $portable.Count -eq 16) 'Historical originals complete'
        $paths=@($raw.FullName)+@($portable.FullName)+@((Join-Path $repo 'docs/roadmap/2026-10-19-p2b-prestart-cause-repair.md'),(Join-Path $repo 'docs/roadmap/2026-10-19-p2b-c1-c2-closure-map.md'),$spec,$plan,(Join-Path $packet approved-spec.snapshot.md),(Join-Path $packet approved-plan.snapshot.md))
        foreach($path in $paths){$pins.Add([pscustomobject][ordered]@{VerificationPath=[IO.Path]::GetFullPath($path);SHA256=(Driver-Hash $path);Role='entry-migration-frozen-dependency';ExpectedState='Present'})}
        foreach($relative in $inputPaths){
            $path=[IO.Path]::GetFullPath((Join-Path $repo $relative))
            $present=[IO.File]::Exists($path)
            $pins.Add([pscustomobject][ordered]@{VerificationPath=$path;SHA256=$(if($present){Driver-Hash $path}else{$null});Role='entry-migration-planned-input';ExpectedState=$(if($present){'Present'}else{'Absent'})})
        }
        Driver-Need ($pins.Count -eq 2064 -and @($pins | Where-Object ExpectedState -CEQ Absent).Count -eq 8) 'Initial2064mapping/8absent requirement'
        $seen=@{}
        foreach($pin in $pins){$key=[IO.Path]::GetFullPath($pin.VerificationPath);$identity=$pin.ExpectedState+':'+$pin.SHA256;if($seen.ContainsKey($key) -and $seen[$key] -cne $identity){throw ('Conflicting mapping '+$key)};$seen[$key]=$identity}
        $manifest=[pscustomobject][ordered]@{Schema=2;Phase=$Grant;BaseCommit='a382faa4ec932f17ad6122ea41abb56ab261d65e';PlanSHA256='062AAC17317B5F8FBE5E29A88A67ABF6E53D5BD4BC6CA8151AC04DE783F81205';SpecSHA256='DA27DA1C6BD3578512FB60A7F38B6635EAC9ACEF6570534F8B868BD07D088D9A';InheritedPins=1930;Pins=$pins.ToArray();MaxHosts=6;MaxChildren=7;MaxReviewers=1;ChangedInputs=[string[]]@();SourceVerified=$false;LiveActionsPerformed=$false;RuntimeApproved=$false;FormalAcceptance=$false;RealGateImplemented=$false}
    }else{
        $previous=if($Grant -ceq 'entry-green'){'entry-red'}else{'entry-green'}
        Driver-Need (-not [string]::IsNullOrEmpty($PreviousFrozenSHA256)) 'Previous freeze identity required'
        $previousPath=Join-Path $packet ($previous+'.frozen.json')
        Driver-CheckHash $previousPath $PreviousFrozenSHA256
        $manifest=Get-Content -LiteralPath $previousPath -Raw | ConvertFrom-Json
        $allowed=if($Grant -ceq 'entry-green'){@('scripts/diagnose-offline.ps1','scripts/private/offline-diagnostics/catalog.json')}else{@('scripts/private/offline-diagnostics/session.psm1','scripts/private/offline-diagnostics/sinks.ps1','scripts/private/offline-diagnostics/owned-step.ps1','scripts/private/offline-diagnostics/sequence.ps1','scripts/private/offline-diagnostics/profiles.ps1','scripts/private/offline-diagnostics/observations.ps1')}
        $changed=[Collections.Generic.List[string]]::new()
        foreach($pin in $manifest.Pins){
            $relative=[IO.Path]::GetRelativePath($repo,$pin.VerificationPath).Replace('\','/')
            if($relative -cin $allowed){
                Driver-Need ($pin.ExpectedState -ceq 'Absent' -and [IO.File]::Exists($pin.VerificationPath)) 'Only planned absent-to-present addition permitted'
                $pin.SHA256=Driver-Hash $pin.VerificationPath;$pin.ExpectedState='Present';$changed.Add($relative)
            }elseif($pin.ExpectedState -ceq 'Absent'){Driver-Need (-not (Test-Path -LiteralPath $pin.VerificationPath)) 'Unexpected new input'}
            else{Driver-CheckHash $pin.VerificationPath $pin.SHA256}
        }
        Driver-Need (($changed.Count -eq $allowed.Count) -and ((@($changed | Sort-Object) -join ',') -ceq (@($allowed | Sort-Object) -join ','))) 'Exact addition window mismatch'
        $manifest.Phase=$Grant;$manifest.ChangedInputs=$changed.ToArray()
    }
    Driver-Verify $manifest
    $snapshot=Join-Path $packet ('snapshots/'+$Grant)
    Driver-Need (-not (Test-Path -LiteralPath $snapshot)) 'Snapshot already exists'
    $null=[IO.Directory]::CreateDirectory($snapshot);$snapshotCount=0
    foreach($relative in $inputPaths){
        $path=Join-Path $repo $relative
        if([IO.File]::Exists($path)){
            $target=Join-Path $snapshot $relative
            $null=[IO.Directory]::CreateDirectory([IO.Path]::GetDirectoryName($target))
            Driver-Once $target ([IO.File]::ReadAllBytes($path));$snapshotCount++
        }
    }
    $wanted=switch -CaseSensitive ($Grant){'entry-red'{12}'entry-green'{14}'runtime-green'{20}}
    Driver-Need ($snapshotCount -eq $wanted) 'Exact present source snapshots'
    Driver-JSON $manifestPath $manifest
    [pscustomobject][ordered]@{Prepared=$Grant;FrozenSHA256=(Driver-Hash $manifestPath);Pins=$manifest.Pins.Count;Snapshots=$snapshotCount;Absent=@($manifest.Pins | Where-Object ExpectedState -CEQ Absent).Count;ChangedInputs=$manifest.ChangedInputs} | ConvertTo-Json
    return
}
Driver-Need (-not [string]::IsNullOrEmpty($FrozenSHA256)) 'Missing frozen identity'
$phase=switch -CaseSensitive ($Grant){'entry-red'{'entry-red'}'entry-green'{'entry-green'}'runtime-red'{'entry-green'}default{'runtime-green'}}
$manifestPath=Join-Path $packet ($phase+'.frozen.json')
Driver-CheckHash $manifestPath $FrozenSHA256
$manifest=Get-Content -LiteralPath $manifestPath -Raw | ConvertFrom-Json
Driver-Need ($manifest.Schema -eq 2 -and $manifest.Pins.Count -eq 2064) 'Exact2064mapping state-aware manifest'
Driver-Verify $manifest
$oracle=Get-Content -LiteralPath (Join-Path $PSScriptRoot expected-output.json) -Raw | ConvertFrom-Json -AsHashtable
$intentPath=Join-Path $packet ($Grant+'.intent.json');$finalPath=Join-Path $packet ($Grant+'.final.json')
foreach($suffix in @('intent.json','final.json','stdout.bin','stderr.bin','evidence.json')){Driver-Need (-not (Test-Path -LiteralPath (Join-Path $packet ($Grant+'.'+$suffix)))) 'Grant attempted/unknown; no retry'}
Driver-JSON $intentPath ([ordered]@{Grant=$Grant;AttemptConsumed=$true;FrozenSHA256=$FrozenSHA256;AuthorizedHost=1;AuthorizedChildren=$oracle.Grants[$Grant].Children})
$row=[ordered]@{Grant=$Grant;Started=$false;PID=$null;OwnedExit=$false;StreamsComplete=$false;ExitCode=$null;TimedOut=$false;TerminationRequested=$false;PrimaryError=$null;SecondaryErrors=@();Overflow=$false;PostPinsIntact=$false;Exact=$false;StdoutSHA256=$null;StderrSHA256=$null;EvidenceSHA256=$null;EvidenceValid=$false;ChildrenObserved=0;EvidenceState='Unknown';ExpectedNegative=($Grant -cin @('entry-red','runtime-red'));ExpectedOutputSHA256=$null;Rescue=$false;SourceVerified=$false;LiveActionsPerformed=$false;RuntimeApproved=$false;FormalAcceptance=$false;RealGateImplemented=$false}
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
