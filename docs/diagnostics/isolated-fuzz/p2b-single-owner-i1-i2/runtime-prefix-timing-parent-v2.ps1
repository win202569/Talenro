param([Parameter(Mandatory)][ValidatePattern('^[A-F0-9]{64}$')][string]$InputSHA256)
Set-StrictMode -Version Latest;$ErrorActionPreference='Stop'
$script:PTRepo='D:\Projects\Talenro\.worktrees\c12-b01-task9-coordinator'
if((Get-Location).Path -cne $script:PTRepo){throw 'PT: fixed root required'}
$d=Join-Path $script:PTRepo '.superpowers/sdd/2026-11-04-runtime-prefix-timing-v1'
function Save-PTParent([string]$Name,$Value){
    $bytes=[Text.UTF8Encoding]::new($false).GetBytes((ConvertTo-Json -InputObject $Value -Depth 100))
    $s=[IO.File]::Open((Join-Path $d $Name),[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::None)
    try{$s.Write($bytes,0,$bytes.Length);$s.Flush()}finally{$s.Dispose()}
}
if((Get-FileHash (Join-Path $d 'stop.json')).Hash -cne '997509D3F0A7FCF0C981D6428C7999384CA6F758E62BDBCC8D193C9BA64B0449'){throw 'PT: historical preparation stop identity'}
foreach($name in @('intent.json','child.claim.json','journal.jsonl','stdout.bin','stderr.bin','result.json','final.json','execution-v2.stop.json','preparation-v2.stop.json','tool-record.json')){
    if(Test-Path -LiteralPath (Join-Path $d $name)){throw 'PT: consumed or partial attempt'}
}
$actualHelpers=[ordered]@{}
foreach($name in @('common','child','parent-v2')){$path='docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/runtime-prefix-timing-'+$name+'.ps1';$actualHelpers[$path]=(Get-FileHash (Join-Path $script:PTRepo $path)).Hash}
Save-PTParent intent.json ([ordered]@{Schema=1;Computation=27;Ceiling=27;MaxProcesses=1;AttemptConsumed=$true;Token=[guid]::NewGuid().ToString('D');InputSHA256=$InputSHA256;HelperSHA256=$actualHelpers})
$row=[ordered]@{Schema=1;Computation=27;Started=$false;PID=$null;OwnedExit=$false;StreamsComplete=$false;ExitCode=$null;TimedOut=$false;TerminationRequested=$false;Overflow=$false;PrimaryError=$null;SecondaryErrors=@();PreflightMs=$null;ExecutionMs=$null;CleanupMs=$null;Before=$null;After=$null;InputSHA256=$InputSHA256;ChildStatus=$null;DiagnosticComplete=$false;SourceVerified=$false;LiveActionsPerformed=$false;RuntimeApproved=$false;FormalAcceptance=$false;RealGateImplemented=$false}
$preflight=[Diagnostics.Stopwatch]::StartNew();$timer=$null;$cleanupTimer=$null
$process=$null;$channels=[Collections.Generic.List[object]]::new();$errors=[Collections.Generic.List[object]]::new()
function Add-PTParentError($record){
    $value=[ordered]@{Type=$record.Exception.GetType().FullName;Message=$record.Exception.Message;ErrorId=$record.FullyQualifiedErrorId;Stack=$record.ScriptStackTrace}
    if($null -eq $row.PrimaryError){$row.PrimaryError=$value}elseif($errors.Count -lt 8){$errors.Add($value)}else{$row.Overflow=$true}
}
function Drain-PTChannels {
    foreach($channel in $channels){
        if($channel.Ended -or -not $channel.Task.IsCompleted){continue}
        $read=$channel.Task.GetAwaiter().GetResult()
        if($read -eq 0){$channel.Ended=$true;continue}
        $remaining=65536-$channel.Bytes
        $write=[Math]::Min($remaining,$read)
        if($write -gt 0){$channel.File.Write($channel.Buffer,0,$write);$channel.File.Flush();$channel.Bytes+=$write}
        if($read -gt $remaining){$row.Overflow=$true;throw 'PT: stream overflow'}
        $channel.Task=$channel.Stream.ReadAsync($channel.Buffer,0,8192)
    }
}
try{
    $inputPath=Join-Path $d 'input.json'
    if((Get-FileHash $inputPath).Hash -cne $InputSHA256){throw 'PT: input identity'}
    $script:PTBinding=Get-Content $inputPath -Raw|ConvertFrom-Json -AsHashtable
    foreach($path in $actualHelpers.Keys){if($script:PTBinding.HelperSHA256[$path] -cne $actualHelpers[$path]){throw 'PT: helper identity'}}
    . (Join-Path $PSScriptRoot 'runtime-prefix-timing-common.ps1')
    $row.Before=Assert-PTInputs $script:PTBinding
    if($script:PTBinding.Limits.ExecutionMs -ne 120000 -or $script:PTBinding.Limits.ExitMs -ne 2000 -or $script:PTBinding.Limits.DrainMs -ne 2000 -or $script:PTBinding.Limits.StreamBytes -ne 65536){throw 'PT: fixed limits'}
    $engine='C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe'
    if((Get-PTHash $engine) -cne '362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139' -or (Get-PTHash (Join-Path ([IO.Path]::GetDirectoryName($engine)) 'System.Management.Automation.dll')) -cne '2D8AF026E589C5603629BFC3E7C57265B4167B12ECA3AE549C6A6C3CA7E4967B'){throw 'PT: pinned runtime'}
    $intentHash=Get-PTHash (Join-Path $d 'intent.json')
    $preflight.Stop();$row.PreflightMs=$preflight.ElapsedMilliseconds
    $timer=[Diagnostics.Stopwatch]::StartNew()
    foreach($name in @('stdout','stderr')){
        $file=[IO.File]::Open((Join-Path $d ($name+'.bin')),[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::Read)
        $channels.Add([pscustomobject]@{Name=$name;File=$file;Stream=$null;Buffer=[byte[]]::new(8192);Task=$null;Bytes=0;Ended=$false})
    }
    $process=[Diagnostics.Process]::new();$process.StartInfo=[Diagnostics.ProcessStartInfo]::new()
    $process.StartInfo.FileName=$engine;$process.StartInfo.UseShellExecute=$false;$process.StartInfo.CreateNoWindow=$true
    $process.StartInfo.RedirectStandardOutput=$true;$process.StartInfo.RedirectStandardError=$true;$process.StartInfo.WorkingDirectory=$script:PTRepo
    foreach($arg in @('-NoLogo','-NoProfile','-NonInteractive','-File',(Join-Path $PSScriptRoot 'runtime-prefix-timing-child.ps1'),'-InputSHA256',$InputSHA256,'-IntentSHA256',$intentHash)){$process.StartInfo.ArgumentList.Add($arg)}
    if($timer.ElapsedMilliseconds -ge 120000){$row.TimedOut=$true;throw 'PT: setup timeout'}
    if(-not $process.Start()){throw 'PT: process did not start'}
    $row.Started=$true;$row.PID=$process.Id
    $channels[0].Stream=$process.StandardOutput.BaseStream;$channels[1].Stream=$process.StandardError.BaseStream
    foreach($channel in $channels){$channel.Task=$channel.Stream.ReadAsync($channel.Buffer,0,8192)}
    while($true){
        if($timer.ElapsedMilliseconds -ge 120000){$row.TimedOut=$true;break}
        Drain-PTChannels
        if($process.HasExited){break}
        $remaining=[Math]::Max(0,120000-[int]$timer.ElapsedMilliseconds)
        $null=$process.WaitForExit([Math]::Min(10,$remaining))
    }
}catch{Add-PTParentError $_}
finally{
    if($preflight.IsRunning){$preflight.Stop();$row.PreflightMs=$preflight.ElapsedMilliseconds}
    if($null -ne $timer){$timer.Stop();$row.ExecutionMs=$timer.ElapsedMilliseconds}
    $cleanupTimer=[Diagnostics.Stopwatch]::StartNew()
    $associated=$false
    if($null -ne $process){
        try{$property=[Diagnostics.Process].GetProperty('Id',([Reflection.BindingFlags]::Public -bor [Reflection.BindingFlags]::Instance));$id=$property.GetValue($process);if($id -is [int] -and $id -gt 0){$associated=$true;$row.PID=$id}}catch{if($row.Started){Add-PTParentError $_}}
        if($associated){
            $exited=$false
            try{$exited=$process.HasExited}catch{Add-PTParentError $_}
            if(-not $exited){$row.TerminationRequested=$true;try{$process.Kill()}catch{Add-PTParentError $_}}
            try{$row.OwnedExit=$process.WaitForExit(2000);if($row.OwnedExit){$row.ExitCode=$process.ExitCode}}catch{Add-PTParentError $_}
        }
    }
    $drain=[Diagnostics.Stopwatch]::StartNew()
    if($channels.Count -eq 2 -and $null -ne $channels[0].Task -and $null -ne $channels[1].Task){
        try{
            while($drain.ElapsedMilliseconds -lt 2000){
                Drain-PTChannels
                if($channels[0].Ended -and $channels[1].Ended){$row.StreamsComplete=$true;break}
                [Threading.Thread]::Sleep(1)
            }
        }catch{Add-PTParentError $_}
    }
    foreach($channel in $channels){
        if($null -ne $channel.Stream){try{$channel.Stream.Dispose()}catch{Add-PTParentError $_}}
        try{$channel.File.Dispose()}catch{Add-PTParentError $_}
    }
    if($null -ne $process){try{$process.Dispose()}catch{Add-PTParentError $_}}
    $cleanupTimer.Stop();$row.CleanupMs=$cleanupTimer.ElapsedMilliseconds
}
try{
    if($null -ne (Get-Variable PTBinding -Scope Script -ErrorAction Ignore)){$row.After=Assert-PTInputs $script:PTBinding}
    if(Test-Path (Join-Path $d 'result.json')){
        $child=Get-Content (Join-Path $d 'result.json') -Raw|ConvertFrom-Json -AsHashtable
        $row.ChildStatus=$child.Status
        $row.DiagnosticComplete=$row.Started -and $row.OwnedExit -and $row.StreamsComplete -and $row.ExitCode -eq 0 -and -not $row.TimedOut -and -not $row.TerminationRequested -and -not $row.Overflow -and $null -eq $row.PrimaryError -and $child.Status -ceq 'PREFIX_COMPLETED' -and $null -eq $child.Error -and $null -eq $child.RecordingError -and $child.UnclosedIntervals -eq 0 -and $child.InputSHA256 -ceq $InputSHA256
    }
}catch{Add-PTParentError $_;$row.DiagnosticComplete=$false}
$row.SecondaryErrors=$errors.ToArray()
Save-PTParent final.json $row
if(-not $row.DiagnosticComplete){Save-PTParent execution-v2.stop.json ([ordered]@{Computation=27;Started=$row.Started;TimedOut=$row.TimedOut;Overflow=$row.Overflow;PrimaryError=$row.PrimaryError;ChildStatus=$row.ChildStatus;Remainder='FROZEN'})}
[ordered]@{Computation=27;Started=$row.Started;DiagnosticComplete=$row.DiagnosticComplete;TimedOut=$row.TimedOut;ExitCode=$row.ExitCode;ChildStatus=$row.ChildStatus;PrimaryError=$row.PrimaryError}|ConvertTo-Json -Depth 12
exit $(if($row.DiagnosticComplete){0}else{1})
