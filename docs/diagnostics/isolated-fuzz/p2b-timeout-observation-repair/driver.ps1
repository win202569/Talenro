param(
    [Parameter(Mandatory)][ValidateSet('Prepare','RunOne')][string]$Action,
    [Parameter(Mandatory)][ValidateSet('prefix-red','prefix-green','recording-red','recording-green','owned-green','sequence-green')][string]$Grant,
    [ValidatePattern('^[A-F0-9]{64}$')][string]$FrozenSHA256,
    [ValidatePattern('^[A-F0-9]{64}$')][string]$PreviousFrozenSHA256
)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$repo=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))
$packet=Join-Path $repo '.superpowers/sdd/2026-10-19-p2b-timeout-observation-repair'
$engine='C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe'
$inputs=@('.gitattributes','README.md','expected-output.json','session.psm1','sinks.ps1','owned-step.ps1','sequence.ps1','core.tests.ps1','owned.tests.ps1','sequence.tests.ps1','capture.ps1','driver.ps1','prefix.tests.ps1','recording.tests.ps1')
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
function Driver-Verify($Manifest){foreach($pin in $Manifest.Pins){Driver-CheckHash $pin.VerificationPath $pin.SHA256}}
function Driver-Static {
    if(@(Get-ChildItem -LiteralPath $PSScriptRoot -File).Count -ne 14){throw 'Expected exactly14 input files before first run'}
    foreach($name in $inputs){
        $path=Join-Path $PSScriptRoot $name
        $bytes=[IO.File]::ReadAllBytes($path)
        if($bytes -contains 13){throw ('Non-LF source: '+$name)}
        if($name.EndsWith('.ps1') -or $name.EndsWith('.psm1')){
            $tokens=$null;$errors=$null
            $null=[Management.Automation.Language.Parser]::ParseFile($path,[ref]$tokens,[ref]$errors)
            if($errors.Count -ne 0){throw ('Static parse: '+$name+' '+($errors.Message -join '; '))}
        }
    }
    $null=Get-Content -LiteralPath (Join-Path $PSScriptRoot expected-output.json) -Raw | ConvertFrom-Json -AsHashtable
}
Driver-CheckHash $engine '362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139'
Driver-CheckHash (Join-Path ([IO.Path]::GetDirectoryName($engine)) System.Management.Automation.dll) '2D8AF026E589C5603629BFC3E7C57265B4167B12ECA3AE549C6A6C3CA7E4967B'

function Driver-Need([bool]$Condition,[string]$Message){if(-not $Condition){throw $Message}}
function Driver-EqualJSON($Actual,$Expected){
    return ((ConvertTo-Json -InputObject $Actual -Depth 50 -Compress) -ceq (ConvertTo-Json -InputObject $Expected -Depth 50 -Compress))
}
function Driver-Evidence($Evidence,$Profile,$Oracle,[string]$CurrentGrant){
    $e=$Evidence
    Driver-Need ($e -is [Collections.IDictionary]) 'Evidence root not object'
    Driver-Need (Driver-EqualJSON @($e.Keys) $Oracle.EvidenceRootNames) 'Evidence root ordered fields mismatch'
    Driver-Need ($e.Schema -eq 1 -and $e.Grant -ceq $CurrentGrant -and $e.Status -ceq $Profile.Status -and -not $e.Overflow) 'Evidence status/binding mismatch'
    Driver-Need (Driver-EqualJSON $e.Groups $Profile.Groups) 'Complete ordered group trace mismatch'
    Driver-Need ($e.Observations -is [array] -and $e.Observations.Count -le 32) 'Observation bounds'
    foreach($flag in @('SourceVerified','LiveActionsPerformed','RuntimeApproved','FormalAcceptance','RealGateImplemented')){
        Driver-Need ($e[$flag] -is [bool] -and -not $e[$flag]) ('Capability flag mismatch '+$flag)
    }
    if($CurrentGrant -ceq 'recording-green'){
        $r=$Oracle.Recording
        Driver-Need ($e.CurrentGroup -ceq $r.CurrentGroup -and (Driver-EqualJSON $e.NotRun $r.NotRun)) 'Recording group/NotRun mismatch'
        Driver-Need ($e.Observations.Count -eq 1 -and $e.Observations[0].Name -ceq 'Read') 'Recording actual observation missing'
        Driver-Need (Driver-EqualJSON $e.Probe ([ordered]@{Ready=$true;PrimarySame=$true;StartCalls=0;AckIssued=0})) 'Recording identity or budget mismatch'
        $path=Join-Path $packet ($CurrentGrant+'/recording/ledger.json')
        $message=$r.PrimaryMessageTemplate.Replace('{LedgerPath}',[IO.Path]::GetFullPath($path))
        Driver-Need ($e.PrimaryError.Type -ceq $r.PrimaryType -and $e.PrimaryError.ErrorId -ceq $r.PrimaryErrorId -and $e.PrimaryError.Message -ceq $message) 'Recording actual primary mismatch'
        Driver-Need ($e.SecondaryErrors.Count -eq 1 -and $e.SecondaryErrors[0].Type -ceq $r.SecondaryType -and $e.SecondaryErrors[0].ErrorId -ceq $r.SecondaryErrorId -and $e.SecondaryErrors[0].Message -ceq $r.SecondaryMessage) 'Recording assertion secondary mismatch'
        $observed=$e.Observations[0].Value
        Driver-Need (-not $observed.Result.Ok -and $observed.View.FailureLatched -and $observed.View.StartCalls -eq 0 -and $observed.View.AckIssued -eq 0 -and $observed.View.Sinks.Ledger.State -ceq 'Failed' -and $observed.View.Sinks.Ledger.LogicalIOAttempts -eq 1) 'Recording lifecycle/IO mismatch'
        Driver-Need ((Driver-EqualJSON $observed.Result.Error $e.PrimaryError) -and (Driver-EqualJSON $observed.View.PrimaryError $e.PrimaryError)) 'Recording primary projection changed'
    }else{
        Driver-Need ($null -eq $e.PrimaryError -and $e.SecondaryErrors.Count -eq 0 -and $null -eq $e.CurrentGroup -and $e.NotRun.Count -eq 0 -and $null -eq $e.Probe) 'Unexpected capture failure or incomplete groups'
    }
    $started=[Collections.Generic.List[object]]::new()
    foreach($entry in $e.Observations){
        Driver-Need ((Driver-EqualJSON @($entry.Keys) @('Name','Value'))) 'Observation root mismatch'
        if($entry.Name -ceq 'Step' -and $null -ne $entry.Value.Result.Observation -and $entry.Value.Result.Observation.Started){
            $o=$entry.Value.Result.Observation
            Driver-Need ($o.Association -ceq 'Associated' -and $o.PID -gt 0 -and $o.OwnedExit -and $o.StreamsComplete) 'Direct owned child lifecycle incomplete'
            $started.Add($o)
        }
    }
    Driver-Need ($started.Count -eq $Profile.Children) 'Exact direct child count mismatch'
    if($started.Count -gt 0){Driver-Need (@($started.PID | Select-Object -Unique).Count -eq $started.Count) 'Duplicate direct child observation'}
    return $started.Count
}

if($Action -ceq 'Prepare'){
    if($Grant -cin @('owned-green','sequence-green')){throw 'Must reuse recording-green freeze; no new Prepare'}
    Driver-Static
    $manifestPath=Join-Path $packet ($Grant+'.frozen.json')
    if(Test-Path -LiteralPath $manifestPath){throw 'Prepare already attempted; no overwrite'}
    if($Grant -ceq 'prefix-red'){
        $existing=@(Get-ChildItem -LiteralPath $packet -Force)
        if(@($existing | Where-Object {$_.Name -cne 'progress.md' -or $_.PSIsContainer}).Count -ne 0){throw 'Packet not pristine'}
        $budget=Join-Path $repo 'docs/roadmap/2026-10-19-p2b-timeout-recording-budget.md'
        Driver-CheckHash $budget 'E0C4F018B54E41A8A16D2354A767D627A351D5E66A9A4C7A5DF272842A9389A9'
        Driver-Once (Join-Path $packet approved-budget.snapshot.md) ([IO.File]::ReadAllBytes($budget))
        $oldPacket=Join-Path $repo '.superpowers/sdd/2026-10-19-p2b-live-completion-supervisor'
        $oldManifest=Join-Path $oldPacket owned-green.frozen.json
        Driver-CheckHash $oldManifest 'CE5B389CC63A0C698504A35ED2C71FE8074D90CF8E0F114CB785D259EEFB522C'
        $old=Get-Content -LiteralPath $oldManifest -Raw | ConvertFrom-Json
        if($old.Pins.Count -ne 1670){throw 'Inherited pin count'}
        $pins=[Collections.Generic.List[object]]::new()
        foreach($pin in $old.Pins){
            if(-not [IO.Path]::IsPathRooted($pin.VerificationPath)){$pin.VerificationPath=[IO.Path]::GetFullPath((Join-Path $repo $pin.VerificationPath))}
            $pins.Add($pin)
        }
        $raw=@(Get-ChildItem -LiteralPath $oldPacket -Recurse -File | Sort-Object FullName)
        $portable=@(Get-ChildItem -LiteralPath (Join-Path $repo 'docs/diagnostics/isolated-fuzz/p2b-live-completion-supervisor') -File | Sort-Object Name)
        if($raw.Count -ne 96 -or $portable.Count -ne 13){throw 'Old originals incomplete'}
        $paths=@($raw.FullName)+@($portable.FullName)+@((Join-Path $repo 'docs/roadmap/2026-10-19-p2b-live-completion-supervisor.md'),(Join-Path $packet approved-budget.snapshot.md))
        foreach($name in $inputs){$paths+=Join-Path $PSScriptRoot $name}
        foreach($path in $paths){$pins.Add([pscustomobject][ordered]@{VerificationPath=[IO.Path]::GetFullPath($path);SHA256=(Driver-Hash $path);Role='timeout-recording-reverification-input'})}
        if($pins.Count -ne 1795){throw 'Expected1795 mappings'}
        $seen=@{}
        foreach($pin in $pins){
            $key=[IO.Path]::GetFullPath($pin.VerificationPath)
            if($seen.ContainsKey($key) -and $seen[$key] -cne $pin.SHA256){throw ('Conflicting mapping: '+$key)}
            $seen[$key]=$pin.SHA256
        }
        $manifest=[pscustomobject][ordered]@{Schema=1;Phase=$Grant;BaseCommit='0f6b454481fae13b2f3bb742a5b6928ab503f47f';BudgetSHA256='E0C4F018B54E41A8A16D2354A767D627A351D5E66A9A4C7A5DF272842A9389A9';InheritedPins=1670;Pins=$pins.ToArray();MaxHosts=6;MaxChildren=7;MaxReviewers=1;ChangedInputs=[string[]]@();SourceVerified=$false;LiveActionsPerformed=$false;RuntimeApproved=$false;FormalAcceptance=$false;RealGateImplemented=$false}
    }else{
        $previous=switch -CaseSensitive ($Grant){'prefix-green'{'prefix-red'}'recording-red'{'prefix-green'}'recording-green'{'recording-red'}}
        if([string]::IsNullOrEmpty($PreviousFrozenSHA256)){throw 'Missing previous manifest identity'}
        $previousPath=Join-Path $packet ($previous+'.frozen.json')
        Driver-CheckHash $previousPath $PreviousFrozenSHA256
        $manifest=Get-Content -LiteralPath $previousPath -Raw | ConvertFrom-Json
        $allowed=@()
        switch -CaseSensitive ($Grant){'prefix-green'{$allowed=@('sequence.ps1')}'recording-red'{$allowed=@()}'recording-green'{$allowed=@('capture.ps1')}}
        $changed=[Collections.Generic.List[string]]::new()
        foreach($pin in $manifest.Pins){
            $hash=Driver-Hash $pin.VerificationPath
            if($hash -cne $pin.SHA256){
                $name=[IO.Path]::GetFileName($pin.VerificationPath)
                if($name -cnotin $allowed -or [IO.Path]::GetDirectoryName($pin.VerificationPath) -ine $PSScriptRoot){throw ('Unapproved drift: '+$pin.VerificationPath)}
                $changed.Add($name);$pin.SHA256=$hash
            }
        }
        if($changed.Count -ne $allowed.Count -or (@($changed | Sort-Object) -join ',') -cne (@($allowed | Sort-Object) -join ',')){throw 'Exact planned change set mismatch'}
        $manifest.Phase=$Grant;$manifest.ChangedInputs=$changed.ToArray()
    }
    Driver-Verify $manifest
    $snap=Join-Path $packet ('snapshots/'+$Grant)
    if(Test-Path -LiteralPath $snap){throw 'Snapshot exists; no retry'}
    $null=[IO.Directory]::CreateDirectory($snap)
    foreach($name in $inputs){Driver-Once (Join-Path $snap $name) ([IO.File]::ReadAllBytes((Join-Path $PSScriptRoot $name)))}
    Driver-JSON $manifestPath $manifest
    [pscustomobject][ordered]@{Prepared=$Grant;FrozenSHA256=(Driver-Hash $manifestPath);Pins=$manifest.Pins.Count;ChangedInputs=$manifest.ChangedInputs} | ConvertTo-Json
    return
}
if([string]::IsNullOrEmpty($FrozenSHA256)){throw 'Missing frozen identity'}
$phase=if($Grant -cin @('owned-green','sequence-green')){'recording-green'}else{$Grant}
$manifestPath=Join-Path $packet ($phase+'.frozen.json')
Driver-CheckHash $manifestPath $FrozenSHA256
$manifest=Get-Content -LiteralPath $manifestPath -Raw | ConvertFrom-Json
if($manifest.Pins.Count -ne 1795){throw 'Pin count'}
Driver-Verify $manifest
$oracle=Get-Content -LiteralPath (Join-Path $PSScriptRoot expected-output.json) -Raw | ConvertFrom-Json -AsHashtable
$intentPath=Join-Path $packet ($Grant+'.intent.json')
$finalPath=Join-Path $packet ($Grant+'.final.json')
foreach($suffix in @('intent.json','final.json','stdout.bin','stderr.bin','evidence.json')){
    if(Test-Path -LiteralPath (Join-Path $packet ($Grant+'.'+$suffix))){throw 'Grant already attempted/unknown: no retry'}
}
# A single fixed grant is explicitly requested by the live executor. There is
# no read of prior Outcome, no continuation from final JSON, no auto-dispatch.
Driver-JSON $intentPath ([ordered]@{Grant=$Grant;AttemptConsumed=$true;FrozenSHA256=$FrozenSHA256;AuthorizedHost=1;AuthorizedChildren=$oracle.Grants[$Grant].Children})
$row=[ordered]@{Grant=$Grant;Started=$false;PID=$null;OwnedExit=$false;StreamsComplete=$false;ExitCode=$null;TimedOut=$false;TerminationRequested=$false;PrimaryError=$null;SecondaryErrors=@();Overflow=$false;PostPinsIntact=$false;Exact=$false;StdoutSHA256=$null;StderrSHA256=$null;EvidenceSHA256=$null;EvidenceValid=$false;ChildrenObserved=0;EvidenceState='Unknown';ExpectedNegative=($Grant -cin @('prefix-red','recording-red','recording-green'));ExpectedOutputSHA256=$null;Rescue=$false;SourceVerified=$false;LiveActionsPerformed=$false;RuntimeApproved=$false;FormalAcceptance=$false;RealGateImplemented=$false}
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
    $max=if($Grant -ceq 'sequence-green'){150000}else{120000}
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

# No positive IO after a transport failure: its candidate artifacts stay quarantined.
if($null -eq $row.PrimaryError -and $row.Started -and $row.OwnedExit -and $row.StreamsComplete -and -not $row.TimedOut -and -not $row.TerminationRequested){
    try{
        Driver-CheckHash $manifestPath $FrozenSHA256
        Driver-Verify $manifest;$row.PostPinsIntact=$true
        $profile=$oracle.Grants[$Grant]
        $stdoutPath=Join-Path $packet ($Grant+'.stdout.bin');$stderrPath=Join-Path $packet ($Grant+'.stderr.bin')
        $evidencePath=Join-Path $packet ($Grant+'.evidence.json')
        $stdout=[Text.UTF8Encoding]::new($false,$true).GetString([IO.File]::ReadAllBytes($stdoutPath))
        $stderr=[Text.UTF8Encoding]::new($false,$true).GetString([IO.File]::ReadAllBytes($stderrPath))
        $expected=($profile.Output -join "`r`n")+"`r`n"
        $expectedError=if($profile.Contains('ErrorOutput')){($profile.ErrorOutput -join "`r`n")+"`r`n"}else{''}
        $row.ExpectedOutputSHA256=[Convert]::ToHexString([Security.Cryptography.SHA256]::HashData([Text.Encoding]::UTF8.GetBytes($expected)))
        $row.StdoutSHA256=Driver-Hash $stdoutPath;$row.StderrSHA256=Driver-Hash $stderrPath
        if($stdout.Replace("`r`n","`n") -cne $expected.Replace("`r`n","`n") -or $stderr.Replace("`r`n","`n") -cne $expectedError.Replace("`r`n","`n")){throw 'Exact stdout/stderr mismatch'}
        if($profile.Evidence -ceq 'Absent'){
            Driver-Need ($Grant -ceq 'recording-red' -and -not (Test-Path -LiteralPath $evidencePath) -and $profile.Children -eq 0) 'Only fixed recording-red may lack evidence'
            $row.EvidenceState='ExpectedAbsent'
        }else{
            Driver-Need ((Get-Item -LiteralPath $evidencePath).Length -le 1048576) 'Evidence byte bound'
            $actual=Get-Content -LiteralPath $evidencePath -Raw | ConvertFrom-Json -AsHashtable
            $row.ChildrenObserved=Driver-Evidence $actual $profile $oracle $Grant
            $row.EvidenceSHA256=Driver-Hash $evidencePath
            $row.EvidenceValid=$true;$row.EvidenceState='RequiredExact'
        }
        if($row.ExitCode -ne $profile.ExitCode){throw 'Unexpected capture exit'}
        $row.Exact=$true
    }catch{Driver-Error $_}
}
$row.SecondaryErrors=$cleanup.ToArray()
# Each grant's intent and final are separate one-attempt files. Complete
# success-looking bytes left by a failed write confer no continuation authority.
Driver-JSON $finalPath $row
[pscustomobject]$row | ConvertTo-Json -Depth 8
if(-not $row.Exact){throw 'Grant nonexact/unknown: freeze all remaining grants and final review'}
