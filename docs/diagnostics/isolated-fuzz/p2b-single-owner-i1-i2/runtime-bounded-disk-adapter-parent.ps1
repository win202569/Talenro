param([Parameter(Mandatory)][ValidatePattern('^[A-F0-9]{64}$')][string]$InputSHA256)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$script:HTRepo='D:\Projects\Talenro\.worktrees\c12-b01-task9-coordinator'
if((Get-Location).Path -cne $script:HTRepo){throw 'CT: fixed root'}
$d=Join-Path $script:HTRepo '.superpowers/sdd/2026-11-04-runtime-bounded-disk-adapter-v1'
function Save-CTOnce([string]$Name,$Value){
 if($Name -cnotin @('intent.json','final.json','stop.json','parent.before-map.json','parent.after-map.json')){throw 'CT: record name'}
 $bytes=[Text.UTF8Encoding]::new($false).GetBytes((ConvertTo-Json -InputObject $Value -Depth 100))
 $limit=if($Name.EndsWith('-map.json')){2097152}else{262144}
 if($bytes.Length -gt $limit){throw 'BA record bound'}
 $s=[IO.File]::Open((Join-Path $d $Name),[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::None)
 try{$s.Write($bytes,0,$bytes.Length);$s.Flush()}finally{$s.Dispose()}
}
function Save-ParentMap([string]$Name,$Map){
 Save-CTOnce $Name $Map
 [ordered]@{Path=('.superpowers/sdd/2026-11-04-runtime-bounded-disk-adapter-v1/'+$Name);Count=$Map.Count;Length=(Get-Item (Join-Path $d $Name)).Length;SHA256=(Get-FileHash (Join-Path $d $Name)).Hash}
}
function Get-CTError($Record){
 $base=$Record.Exception.GetBaseException()
 $values=[ordered]@{Type=$Record.Exception.GetType().FullName;Message=$Record.Exception.Message;ErrorId=$Record.FullyQualifiedErrorId;Stack=$Record.ScriptStackTrace;BaseType=$base.GetType().FullName;BaseStack=$base.StackTrace}
 $truncated=$false
 foreach($key in @($values.Keys)){if($null -ne $values[$key] -and $values[$key].Length -gt 4096){$values[$key]=$values[$key].Substring(0,4096);$truncated=$true}}
 $values.Truncated=$truncated
 return $values
}
foreach($name in @('stop.json','intent.json','worker.claim.json','stages.jsonl','capture','verification-result.json','stdout.bin','stderr.bin','final.json','tool-record.json','adapter.dll','cases','compile.intent.json','compile.result.json','worker.stop.json','parent.before-map.json','parent.after-map.json','worker.before-map.json','worker.after-map.json')){
 if(Test-Path (Join-Path $d $name)){throw 'CT: consumed/partial namespace'}
}
$helpers=[ordered]@{}
foreach($name in @('runtime-bounded-disk-adapter.cs','runtime-bounded-disk-adapter.tests.cs','runtime-bounded-disk-adapter-worker.ps1','runtime-bounded-disk-adapter-parent.ps1')){
 $p='docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/'+$name
 $helpers[$p]=(Get-FileHash (Join-Path $script:HTRepo $p)).Hash
}
Save-CTOnce intent.json ([ordered]@{Schema=1;Computation=30;Ceiling=30;MaxProcesses=1;AttemptConsumed=$true;Token=[guid]::NewGuid().ToString('D');InputSHA256=$InputSHA256;HelperSHA256=$helpers})
$row=[ordered]@{Schema=1;Computation=30;Started=$false;PID=$null;OwnedExit=$false;ExitCode=$null;TimedOut=$false;TerminationRequested=$false;StreamsComplete=$false;Overflow=$false;ReleaseDeferred=$false;PrimaryError=$null;SecondaryErrors=@();PreflightMs=$null;ExecutionMs=$null;CleanupMs=$null;DrainMs=$null;Before=$null;After=$null;InputSHA256=$InputSHA256;Channels=@();ChildStatus=$null;DiagnosticComplete=$false;Status='INCOMPLETE';SourceVerified=$false;LiveActionsPerformed=$false;RuntimeApproved=$false;FormalAcceptance=$false;RealGateImplemented=$false}
$errors=[Collections.Generic.List[object]]::new()
$channels=[Collections.Generic.List[object]]::new()
$process=$null;$timer=$null;$preflight=[Diagnostics.Stopwatch]::StartNew()
$script:CTRetained=$null
function Add-CTError($Record){
 $value=Get-CTError $Record
 if($null -eq $row.PrimaryError){$row.PrimaryError=$value}elseif($errors.Count -lt 8){$errors.Add($value)}else{$row.Overflow=$true}
}
function Complete-CTChannel($Channel){
 if($null -eq $Channel.Task -or -not $Channel.Task.IsCompleted){throw 'CT: terminal task required'}
 $Channel.Record.TaskStatus=$Channel.Task.Status.ToString()
 try{$null=$Channel.Task.GetAwaiter().GetResult();$Channel.Record.CopySucceeded=$true}catch{
  $Channel.Record.CopyError=Get-CTError $_
  $base=$_.Exception.GetBaseException()
  $Channel.Record.CapacityFault=$base -is [NotSupportedException] -and $null -ne $base.StackTrace -and $base.StackTrace.Contains('System.IO.MemoryStream.EnsureCapacity')
 }
 try{
  $length=[int]$Channel.Destination.Position
  if($length -lt 0 -or $length -gt 65536){throw 'CT: destination position'}
  if($Channel.Name -cnotin @('stdout','stderr')){throw 'CT: channel name'}
  $path=Join-Path $d ($Channel.Name+'.bin')
  $file=[IO.File]::Open($path,[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::None)
  try{$file.Write($Channel.Backing,0,$length);$file.Flush()}finally{$file.Dispose()}
  $Channel.Record.OutputAvailable=$true;$Channel.Record.RetainedBytes=$length
  $Channel.Record.OutputSHA256=(Get-FileHash -LiteralPath $path).Hash
 }catch{Add-CTError $_}
 foreach($resource in @($Channel.Source,$Channel.Destination,$Channel.CancelSource)){
  if($null -ne $resource){try{$resource.Dispose()}catch{Add-CTError $_}}
 }
}
try{
 $inputPath=Join-Path $d input.json
 if((Get-FileHash $inputPath).Hash -cne $InputSHA256){throw 'CT: input identity'}
 $script:HTBinding=Get-Content $inputPath -Raw|ConvertFrom-Json -AsHashtable
 foreach($p in $helpers.Keys){if($script:HTBinding.HelperSHA256[$p] -cne $helpers[$p]){throw 'CT: helper identity'}}
 . (Join-Path $PSScriptRoot runtime-copy-transport-common.ps1)
 $row.Before=Save-ParentMap parent.before-map.json (Assert-HTInputs $script:HTBinding)
 if($script:HTBinding.Limits.ExecutionMs -ne 120000 -or $script:HTBinding.Limits.ExitMs -ne 2000 -or $script:HTBinding.Limits.DrainMs -ne 2000 -or $script:HTBinding.Limits.StreamBytes -ne 65536){throw 'CT: fixed limits'}
 $engine='C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe'
 if((Get-FileHash $engine).Hash -cne '362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139' -or (Get-FileHash (Join-Path ([IO.Path]::GetDirectoryName($engine)) System.Management.Automation.dll)).Hash -cne '2D8AF026E589C5603629BFC3E7C57265B4167B12ECA3AE549C6A6C3CA7E4967B'){throw 'CT: runtime identity'}
 $intentHash=(Get-FileHash (Join-Path $d intent.json)).Hash
 $preflight.Stop();$row.PreflightMs=$preflight.ElapsedMilliseconds
 $timer=[Diagnostics.Stopwatch]::StartNew()
 foreach($name in @('stdout','stderr')){
  $backing=[byte[]]::new(65536)
  $destination=[IO.MemoryStream]::new($backing,0,65536,$true,$false)
  $cancel=[Threading.CancellationTokenSource]::new()
  $record=[ordered]@{Name=$name;TaskCreated=$false;Terminal=$false;TaskStatus='NotCreated';CopySucceeded=$false;CopyError=$null;OutputAvailable=$false;RetainedBytes=$null;OutputSHA256=$null;Capacity=65536;CapacityFault=$false;ReleaseDeferred=$false}
  $channels.Add([pscustomobject]@{Name=$name;Source=$null;Destination=$destination;Backing=$backing;Task=$null;CancelSource=$cancel;Record=$record})
 }
 $stdout=$channels[0];$stderr=$channels[1]
 $process=[Diagnostics.Process]::new()
 $process.StartInfo=[Diagnostics.ProcessStartInfo]::new()
 $process.StartInfo.FileName=$engine;$process.StartInfo.UseShellExecute=$false;$process.StartInfo.CreateNoWindow=$true
 $process.StartInfo.RedirectStandardOutput=$true;$process.StartInfo.RedirectStandardError=$true;$process.StartInfo.WorkingDirectory=$script:HTRepo
 foreach($arg in @('-NoLogo','-NoProfile','-NonInteractive','-File',(Join-Path $PSScriptRoot runtime-bounded-disk-adapter-worker.ps1),'-InputSHA256',$InputSHA256,'-IntentSHA256',$intentHash)){$process.StartInfo.ArgumentList.Add($arg)}
 if($timer.ElapsedMilliseconds -ge 120000){$row.TimedOut=$true;throw 'CT: setup timeout'}
 if(-not $process.Start()){throw 'CT: start returned false'}
 $row.Started=$true;$row.PID=$process.Id
 $stdout.Source=$process.StandardOutput.BaseStream;$stderr.Source=$process.StandardError.BaseStream
 $stdout.Task=$stdout.Source.CopyToAsync($stdout.Destination,8192,$stdout.CancelSource.Token)
 $stdout.Record.TaskCreated=$true
 $stderr.Task=$stderr.Source.CopyToAsync($stderr.Destination,8192,$stderr.CancelSource.Token)
 $stderr.Record.TaskCreated=$true
 $remaining=[Math]::Max(0,120000-[int]$timer.ElapsedMilliseconds)
 if(-not $process.WaitForExit($remaining)){$row.TimedOut=$true}
}catch{Add-CTError $_}
finally{
 if($preflight.IsRunning){$preflight.Stop();$row.PreflightMs=$preflight.ElapsedMilliseconds}
 if($null -ne $timer){$timer.Stop();$row.ExecutionMs=$timer.ElapsedMilliseconds}
 $cleanup=[Diagnostics.Stopwatch]::StartNew()
 $associated=$false
 if($null -ne $process){
  try{
   $property=[Diagnostics.Process].GetProperty('Id',([Reflection.BindingFlags]::Public -bor [Reflection.BindingFlags]::Instance))
   $id=$property.GetValue($process)
   if($id -is [int] -and $id -gt 0){$associated=$true;$row.PID=$id}
  }catch{if($row.Started){Add-CTError $_}}
  if($associated){
   $exited=$false;try{$exited=$process.HasExited}catch{Add-CTError $_}
   if(-not $exited){$row.TerminationRequested=$true;try{$process.Kill()}catch{Add-CTError $_}}
   try{$row.OwnedExit=$process.WaitForExit(2000);if($row.OwnedExit){$row.ExitCode=$process.ExitCode}}catch{Add-CTError $_}
  }
 }
 $drain=[Diagnostics.Stopwatch]::StartNew()
 foreach($channel in $channels){
  if($null -ne $channel.Task -and -not $channel.Task.IsCompleted){try{$channel.CancelSource.CancelAfter(1500)}catch{Add-CTError $_}}
 }
 while($drain.ElapsedMilliseconds -lt 2000){
  $pending=@($channels|Where-Object {$null -ne $_.Task -and -not $_.Task.IsCompleted}).Count
  if($pending -eq 0){break}
  [Threading.Thread]::Sleep(1)
 }
 # Freeze availability before any extraction. No background publication afterward.
 foreach($channel in $channels){
  $channel.Record.Terminal=$null -ne $channel.Task -and $channel.Task.IsCompleted
  if($null -ne $channel.Task){$channel.Record.TaskStatus=$channel.Task.Status.ToString()}
 }
 $drain.Stop();$row.DrainMs=$drain.ElapsedMilliseconds
 foreach($channel in $channels){
  if($channel.Record.Terminal){
   try{Complete-CTChannel $channel}catch{Add-CTError $_}
  }else{
   if($null -ne $channel.Task){
    $channel.Record.ReleaseDeferred=$true;$row.ReleaseDeferred=$true
   }
  }
  if($null -eq $channel.Task){
   foreach($resource in @($channel.Source,$channel.Destination,$channel.CancelSource)){if($null -ne $resource){try{$resource.Dispose()}catch{Add-CTError $_}}}
  }
 }
 if($row.ReleaseDeferred){$script:CTRetained=[pscustomobject]@{Process=$process;Channels=$channels}}
 elseif($null -ne $process){try{$process.Dispose()}catch{Add-CTError $_}}
 $row.StreamsComplete=$channels.Count -eq 2 -and @($channels|Where-Object {-not $_.Record.CopySucceeded}).Count -eq 0
 $row.Overflow=$row.Overflow -or @($channels|Where-Object {$_.Record.CapacityFault}).Count -gt 0
 $row.Channels=@($channels|ForEach-Object {$_.Record})
 $cleanup.Stop();$row.CleanupMs=$cleanup.ElapsedMilliseconds
}
try{
 if($null -ne (Get-Variable HTBinding -Scope Script -ErrorAction Ignore)){$row.After=Save-ParentMap parent.after-map.json (Assert-HTInputs $script:HTBinding)}
 if($row.OwnedExit -and (Test-Path (Join-Path $d verification-result.json))){
  $child=Get-Content (Join-Path $d verification-result.json) -Raw|ConvertFrom-Json -AsHashtable
  $row.ChildStatus=$child.Status
  $outputExact=$false
  if($row.Channels.Count -eq 2 -and $row.Channels[0].OutputAvailable -and $row.Channels[1].OutputAvailable){
   $expected=[Text.Encoding]::UTF8.GetBytes('ADAPTER_VERIFIED'+[char]13+[char]10)
   $actual=[IO.File]::ReadAllBytes((Join-Path $d stdout.bin))
   $outputExact=[Linq.Enumerable]::SequenceEqual([byte[]]$actual,[byte[]]$expected) -and (Get-Item (Join-Path $d stderr.bin)).Length -eq 0
  }
  $matrix=$null -ne $child.Report -and $child.Report.CaseResults.Count -eq 24 -and @($child.Report.CaseResults|Where-Object {-not $_.Passed}).Count -eq 0 -and $child.Report.PendingOperations -eq 0 -and $child.Report.UnavailablePaths.Count -eq 0
  $row.DiagnosticComplete=$row.Started -and $row.OwnedExit -and $row.StreamsComplete -and $row.ExitCode -eq 0 -and $outputExact -and -not $row.TimedOut -and -not $row.TerminationRequested -and -not $row.Overflow -and -not $row.ReleaseDeferred -and $null -eq $row.PrimaryError -and $errors.Count -eq 0 -and $null -ne $row.After -and $child.Status -ceq 'ADAPTER_VERIFIED' -and $matrix -and $null -eq $child.Error -and $child.InputSHA256 -ceq $InputSHA256
 }
}catch{Add-CTError $_;$row.DiagnosticComplete=$false}
$row.SecondaryErrors=$errors.ToArray()
if($row.DiagnosticComplete){$row.Status='ADAPTER_VERIFIED'}elseif($row.TimedOut){$row.Status='TIMEOUT'}elseif(@($row.Channels|Where-Object {$null -ne $_.CopyError}).Count){$row.Status='COPY_FAILED'}
Save-CTOnce final.json $row
if(-not $row.DiagnosticComplete){Save-CTOnce stop.json ([ordered]@{Computation=30;Started=$row.Started;Status=$row.Status;PrimaryError=$row.PrimaryError;Remainder='FROZEN'})}
[ordered]@{Computation=30;Started=$row.Started;Status=$row.Status;DiagnosticComplete=$row.DiagnosticComplete;TimedOut=$row.TimedOut;ExitCode=$row.ExitCode;PrimaryError=$row.PrimaryError}|ConvertTo-Json -Depth 8
exit $(if($row.DiagnosticComplete){0}else{1})
