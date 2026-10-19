param(
    [Parameter(Mandatory)][ValidateSet('Prepare','RunOne')][string]$Action,
    [Parameter(Mandatory)][ValidateSet('core-red','core-green','owned-red','owned-green','sequence-green')][string]$Grant,
    [ValidatePattern('^[A-F0-9]{64}$')][string]$FrozenSHA256,
    [ValidatePattern('^[A-F0-9]{64}$')][string]$PreviousFrozenSHA256
)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$repo=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))
$packet=Join-Path $repo '.superpowers/sdd/2026-10-19-p2b-live-completion-supervisor'
$engine='C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe'
$inputs=@('.gitattributes','README.md','expected-output.json','session.psm1','sinks.ps1','owned-step.ps1','sequence.ps1','core.tests.ps1','owned.tests.ps1','sequence.tests.ps1','capture.ps1','driver.ps1')
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
    if(@(Get-ChildItem -LiteralPath $PSScriptRoot -File).Count -ne 12){throw 'Expected exactly12 input files before first run'}
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
if($Action -ceq 'Prepare'){
    if($Grant -ceq 'sequence-green'){throw 'Sequence must reuse owned-green freeze; no new Prepare'}
    Driver-Static
    $manifestPath=Join-Path $packet ($Grant+'.frozen.json')
    if(Test-Path -LiteralPath $manifestPath){throw 'Prepare already attempted; no overwrite'}
    if($Grant -ceq 'core-red'){
        $existing=@(Get-ChildItem -LiteralPath $packet -Force)
        if(@($existing | Where-Object {$_.Name -cne 'progress.md' -or $_.PSIsContainer}).Count -ne 0){throw 'Packet not pristine'}
        $spec=Join-Path $repo 'docs/superpowers/specs/2026-10-19-p2b-live-completion-supervisor-design.md'
        $plan=Join-Path $repo 'docs/superpowers/plans/2026-10-19-p2b-live-completion-supervisor.md'
        Driver-CheckHash $spec '6687E0196D1B8DF5BAEE3E603F5EA86018AE84292D9D38116FCF17A35F6FD49F'
        Driver-CheckHash $plan '43B98C0C0EE06AC6A378ED17AD534BD6200575FFDAA520A589413F759EDFE172'
        Driver-Once (Join-Path $packet approved-spec.snapshot.md) ([IO.File]::ReadAllBytes($spec))
        Driver-Once (Join-Path $packet approved-plan.snapshot.md) ([IO.File]::ReadAllBytes($plan))
        $oldPacket=Join-Path $repo '.superpowers/sdd/2026-10-19-p2b-completion-publication-repair'
        $oldManifest=Join-Path $oldPacket green.frozen.json
        Driver-CheckHash $oldManifest '3A763103C93625C522CF8BFF997F9280EA07F3278A6D3FCE6D221759368D4237'
        $old=Get-Content -LiteralPath $oldManifest -Raw | ConvertFrom-Json
        if($old.Pins.Count -ne 1599){throw 'Inherited pin count'}
        $pins=[Collections.Generic.List[object]]::new()
        foreach($pin in $old.Pins){
            if(-not [IO.Path]::IsPathRooted($pin.VerificationPath)){$pin.VerificationPath=[IO.Path]::GetFullPath((Join-Path $repo $pin.VerificationPath))}
            $pins.Add($pin)
        }
        $raw=@(Get-ChildItem -LiteralPath $oldPacket -Recurse -File | Sort-Object FullName)
        $portable=@(Get-ChildItem -LiteralPath (Join-Path $repo 'docs/diagnostics/isolated-fuzz/p2b-completion-publication-repair') -File | Sort-Object Name)
        if($raw.Count -ne 45 -or $portable.Count -ne 11){throw 'Old originals incomplete'}
        $paths=@($raw.FullName)+@($portable.FullName)+@((Join-Path $repo 'docs/roadmap/2026-10-19-p2b-completion-publication-repair.md'),(Join-Path $packet approved-spec.snapshot.md),(Join-Path $packet approved-plan.snapshot.md))
        foreach($name in $inputs){$paths+=Join-Path $PSScriptRoot $name}
        foreach($path in $paths){$pins.Add([pscustomobject][ordered]@{VerificationPath=[IO.Path]::GetFullPath($path);SHA256=(Driver-Hash $path);Role='supervisor-reverification-input'})}
        if($pins.Count -ne 1670){throw 'Expected1670 mappings'}
        # Inherited mappings intentionally repeat historical paths/roles. Same
        # canonical path with conflicting hashes is never accepted.
        $seen=@{}
        foreach($pin in $pins){
            $key=[IO.Path]::GetFullPath($pin.VerificationPath)
            if($seen.ContainsKey($key) -and $seen[$key] -cne $pin.SHA256){throw ('Conflicting mapping: '+$key)}
            $seen[$key]=$pin.SHA256
        }
        $manifest=[pscustomobject][ordered]@{Schema=1;Phase=$Grant;BaseCommit='7c3e85efe1a8f9e7377ca39faa8359436327ce2d';SpecSHA256='6687E0196D1B8DF5BAEE3E603F5EA86018AE84292D9D38116FCF17A35F6FD49F';PlanSHA256='43B98C0C0EE06AC6A378ED17AD534BD6200575FFDAA520A589413F759EDFE172';InheritedPins=1599;Pins=$pins.ToArray();MaxHosts=5;MaxChildren=8;MaxReviewers=1;ChangedInputs=[string[]]@();SourceVerified=$false;LiveActionsPerformed=$false;RuntimeApproved=$false;FormalAcceptance=$false;RealGateImplemented=$false}
    }else{
        $previous=switch -CaseSensitive ($Grant){'core-green'{'core-red'}'owned-red'{'core-green'}'owned-green'{'owned-red'}}
        if([string]::IsNullOrEmpty($PreviousFrozenSHA256)){throw 'Missing previous manifest identity'}
        $previousPath=Join-Path $packet ($previous+'.frozen.json')
        Driver-CheckHash $previousPath $PreviousFrozenSHA256
        $manifest=Get-Content -LiteralPath $previousPath -Raw | ConvertFrom-Json
        $allowed=@()
        switch -CaseSensitive ($Grant){'core-green'{$allowed=@('session.psm1','sinks.ps1')}'owned-red'{$allowed=@()}'owned-green'{$allowed=@('owned-step.ps1')}}
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
$phase=if($Grant -ceq 'sequence-green'){'owned-green'}else{$Grant}
$manifestPath=Join-Path $packet ($phase+'.frozen.json')
Driver-CheckHash $manifestPath $FrozenSHA256
$manifest=Get-Content -LiteralPath $manifestPath -Raw | ConvertFrom-Json
if($manifest.Pins.Count -ne 1670){throw 'Pin count'}
Driver-Verify $manifest
$oracle=Get-Content -LiteralPath (Join-Path $PSScriptRoot expected-output.json) -Raw | ConvertFrom-Json -AsHashtable
$intentPath=Join-Path $packet ($Grant+'.intent.json')
$finalPath=Join-Path $packet ($Grant+'.final.json')
foreach($suffix in @('intent.json','final.json','stdout.bin','stderr.bin','trace.json','details.json')){
    if(Test-Path -LiteralPath (Join-Path $packet ($Grant+'.'+$suffix))){throw 'Grant already attempted/unknown: no retry'}
}
# A single fixed grant is explicitly requested by the live executor. There is
# no read of prior Outcome, no continuation from final JSON, no auto-dispatch.
Driver-JSON $intentPath ([ordered]@{Grant=$Grant;AttemptConsumed=$true;FrozenSHA256=$FrozenSHA256;AuthorizedHost=1;AuthorizedChildren=$oracle.Grants[$Grant].Children})
$row=[ordered]@{Grant=$Grant;Started=$false;PID=$null;OwnedExit=$false;StreamsComplete=$false;ExitCode=$null;TimedOut=$false;TerminationRequested=$false;PrimaryError=$null;SecondaryErrors=@();Overflow=$false;PostPinsIntact=$false;Exact=$false;StdoutSHA256=$null;StderrSHA256=$null;TraceSHA256=$null;DetailsSHA256=$null;ExpectedOutputSHA256=$null;Rescue=$false;SourceVerified=$false;LiveActionsPerformed=$false;RuntimeApproved=$false;FormalAcceptance=$false;RealGateImplemented=$false}
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
# Transport finalization is independent of the under-test module. A transport
# failure quarantines its candidates: do not parse or reopen them for success.
if($null -eq $row.PrimaryError -and $row.Started -and $row.OwnedExit -and $row.StreamsComplete -and -not $row.TimedOut -and -not $row.TerminationRequested){
    try{
        Driver-CheckHash $manifestPath $FrozenSHA256
        Driver-Verify $manifest;$row.PostPinsIntact=$true
        $stdoutPath=Join-Path $packet ($Grant+'.stdout.bin');$stderrPath=Join-Path $packet ($Grant+'.stderr.bin')
        $tracePath=Join-Path $packet ($Grant+'.trace.json');$detailsPath=Join-Path $packet ($Grant+'.details.json')
        $stdout=[Text.UTF8Encoding]::new($false,$true).GetString([IO.File]::ReadAllBytes($stdoutPath))
        $expected=($oracle.Grants[$Grant].Output -join "`r`n")+"`r`n"
        $row.ExpectedOutputSHA256=[Convert]::ToHexString([Security.Cryptography.SHA256]::HashData([Text.Encoding]::UTF8.GetBytes($expected)))
        $row.StdoutSHA256=Driver-Hash $stdoutPath;$row.StderrSHA256=Driver-Hash $stderrPath
        if($stdout.Replace("`r`n","`n") -cne $expected.Replace("`r`n","`n") -or (Get-Item -LiteralPath $stderrPath).Length -ne 0){throw 'Exact stdout/stderr mismatch'}
        $actual=Get-Content -LiteralPath $tracePath -Raw | ConvertFrom-Json -AsHashtable
        $want=[ordered]@{Groups=$oracle.Grants[$Grant].Groups}
        if((ConvertTo-Json -InputObject $actual -Depth 50 -Compress) -cne (ConvertTo-Json -InputObject $want -Depth 50 -Compress)){throw 'Complete ordered trace mismatch'}
        $row.TraceSHA256=Driver-Hash $tracePath;$row.DetailsSHA256=Driver-Hash $detailsPath
        $null=Get-Content -LiteralPath $detailsPath -Raw | ConvertFrom-Json -AsHashtable
        if($row.ExitCode -ne $oracle.Grants[$Grant].ExitCode){throw 'Unexpected capture exit'}
        $row.Exact=$true
    }catch{Driver-Error $_}
}
$row.SecondaryErrors=$cleanup.ToArray()
# This file is written once, after exit/drains/closes/checks. Even complete
# success bytes left by a failing write confer no authority on a new call.
Driver-JSON $finalPath $row
[pscustomobject]$row | ConvertTo-Json -Depth 8
if(-not $row.Exact){throw 'Grant nonexact/unknown: freeze all remaining grants and final review'}
