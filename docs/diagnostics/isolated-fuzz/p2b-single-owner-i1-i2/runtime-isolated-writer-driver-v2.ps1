param(
 [Parameter(Mandatory)][ValidateSet('Compile','Scenarios')][string]$Stage,
 [Parameter(Mandatory)][ValidatePattern('^[A-F0-9]{64}$')][string]$InputSHA256
)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$script:HTRepo='D:\Projects\Talenro\.worktrees\c12-b01-task9-coordinator'
$relative='.superpowers/sdd/2026-11-04-runtime-isolated-writer-v1'
$base=Join-Path $script:HTRepo $relative
$engine='C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe'
$rolePath=Join-Path $PSScriptRoot runtime-isolated-writer-role.ps1
$folder=$base;$part='';$binding=$null;$retained=$null
function Need-IW([bool]$Value,[string]$Why){if(-not $Value){throw "IW $Why"}}
function Save-IW([string]$Name,$Value){
 $bytes=[Text.UTF8Encoding]::new($false).GetBytes((ConvertTo-Json -InputObject $Value -Depth 80))
 $limit=if($Name.EndsWith('-map.json')){2097152}else{262144}
 Need-IW ($bytes.Length -le $limit) 'record bound'
 $stream=[IO.File]::Open((Join-Path $folder $Name),[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::None)
 try{$stream.Write($bytes,0,$bytes.Length);$stream.Flush()}finally{$stream.Dispose()}
}
function Map-IW([string]$Name){
 $map=Assert-HTInputs $binding;Save-IW $Name $map
 return [ordered]@{Path="$relative/$part/$Name";Count=$map.Count;Length=(Get-Item (Join-Path $folder $Name)).Length;SHA256=(Get-FileHash (Join-Path $folder $Name)).Hash}
}
function Journal-IW([int]$Seq,[string]$Kind,[string]$Case){
 $bytes=[Text.UTF8Encoding]::new($false).GetBytes((ConvertTo-Json -Compress ([ordered]@{Schema=1;Seq=$Seq;Kind=$Kind;Case=$Case;Ticks=[Diagnostics.Stopwatch]::GetTimestamp();Frequency=[Diagnostics.Stopwatch]::Frequency}))+[char]10)
 Need-IW ($Seq -le 128) 'journal records'
 $mode=if($Seq -eq 1){[IO.FileMode]::CreateNew}else{[IO.FileMode]::Append}
 $stream=[IO.File]::Open((Join-Path $folder journal.jsonl),$mode,[IO.FileAccess]::Write,[IO.FileShare]::Read)
 try{Need-IW ($stream.Length+$bytes.Length -le 65536) 'journal bytes';$stream.Write($bytes,0,$bytes.Length);$stream.Flush()}finally{$stream.Dispose()}
}
function Begin-IW([int]$Computation,[string]$Scenario,[int]$MaxStarts){
 Need-IW (-not(Test-Path $folder)) 'consumed directory'
 $null=New-Item -ItemType Directory -Path $folder
 $intent=[ordered]@{Schema=1;Computation=$Computation;Scenario=$Scenario;MaxStarts=$MaxStarts;AttemptConsumed=$true;Token=[guid]::NewGuid().ToString('D');InputSHA256=$InputSHA256}
 Save-IW intent.json $intent
 return $intent
}
function Verify-IW{
 Need-IW ((Get-Location).Path -ceq $script:HTRepo) 'fixed root'
 Need-IW ((Get-FileHash (Join-Path $base input.json)).Hash -ceq $InputSHA256) 'input identity'
 $script:binding=Get-Content (Join-Path $base input.json) -Raw|ConvertFrom-Json -AsHashtable
 Need-IW ((Get-FileHash $engine).Hash -ceq $binding.EngineSHA256) 'engine identity'
 Need-IW ((Get-FileHash (Join-Path ([IO.Path]::GetDirectoryName($engine)) System.Management.Automation.dll)).Hash -ceq $binding.SMASHA256) 'SMA identity'
 foreach($p in $binding.HelperSHA256.Keys){Need-IW ((Get-FileHash (Join-Path $script:HTRepo $p)).Hash -ceq $binding.HelperSHA256[$p]) "helper drift $p"}
 $common='docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/runtime-copy-transport-common.ps1'
 Need-IW ((Get-FileHash (Join-Path $script:HTRepo $common)).Hash -ceq $binding.HistoricalEvidenceSHA256[$common]) 'common identity'
 . (Join-Path $script:HTRepo $common)
 # Dot-source into this script scope so only definition-only hash helpers are used.
 $script:iwHashFunction=${function:Assert-HTInputs}
 $script:iwGetHashFunction=${function:Get-HTHash}
}
function Invoke-CompilerIW{
 $row=[ordered]@{Computation=31;InputSHA256=$InputSHA256;Started=$false;Associated=$false;PID=$null;OwnedExit=$false;ExitCode=$null;KillRequested=$false;PendingResources=$false;Before=$null;After=$null;ElapsedMs=$null;Error=$null;Channels=@();Passed=$false;EvidencePersisted=$false;Status='NOT_ACCEPTED'}
 $row.Before=Map-IW supervisor.before-map.json
 Journal-IW 1 begin Compile
 $p=[Diagnostics.Process]::new();$cts=[Threading.CancellationTokenSource]::new();$channels=[Collections.Generic.List[object]]::new();$clock=[Diagnostics.Stopwatch]::StartNew()
 try{
  $p.StartInfo=[Diagnostics.ProcessStartInfo]::new();$p.StartInfo.FileName=$engine;$p.StartInfo.WorkingDirectory=$script:HTRepo;$p.StartInfo.UseShellExecute=$false;$p.StartInfo.CreateNoWindow=$true;$p.StartInfo.RedirectStandardOutput=$true;$p.StartInfo.RedirectStandardError=$true
  foreach($arg in @('-NoLogo','-NoProfile','-NonInteractive','-File',$rolePath,'-Role','Compile','-InputSHA256',$InputSHA256,'-IntentSHA256',(Get-FileHash (Join-Path $folder intent.json)).Hash)){$p.StartInfo.ArgumentList.Add($arg)}
  Need-IW ($p.Start()) 'compiler start';$row.Started=$true
  foreach($name in @('stdout','stderr')){
   $backing=[byte[]]::new(65536);$memory=[IO.MemoryStream]::new($backing,0,65536,$true,$false)
   $source=if($name -ceq 'stdout'){$p.StandardOutput.BaseStream}else{$p.StandardError.BaseStream}
   $copy=$source.CopyToAsync($memory,8192,$cts.Token)
   $channels.Add([pscustomobject]@{Name=$name;Bytes=$backing;Memory=$memory;Source=$source;Task=$copy})
  }
  Need-IW ($p.WaitForExit([Math]::Max(0,120000-[int]$clock.ElapsedMilliseconds))) 'compiler execution timeout'
 }catch{$row.Error=$_.Exception.Message}
 finally{
  try{$row.PID=$p.Id;$row.Associated=$true}catch{}
  $exitClock=[Diagnostics.Stopwatch]::StartNew()
  if($row.Associated){
   try{if(-not $p.HasExited){$row.KillRequested=$true;$p.Kill()};$row.OwnedExit=$p.WaitForExit([Math]::Max(0,2000-[int]$exitClock.ElapsedMilliseconds));if($row.OwnedExit){$row.ExitCode=$p.ExitCode}}catch{$row.Error=$_.Exception.Message}
  }
  $drain=[Diagnostics.Stopwatch]::StartNew()
  while(@($channels|Where-Object {-not $_.Task.IsCompleted}).Count -gt 0 -and $drain.ElapsedMilliseconds -lt 2000){[Threading.Thread]::Sleep(1)}
  if(@($channels|Where-Object {-not $_.Task.IsCompleted}).Count -gt 0){$cts.Cancel();$row.PendingResources=$true}
  foreach($c in $channels){
   $capture=[ordered]@{Name=$c.Name;TaskStatus=$c.Task.Status.ToString();Available=$false;Length=$null;SHA256=$null;Error=$null}
   if($c.Task.IsCompleted){
    try{
     $null=$c.Task.GetAwaiter().GetResult();$length=[int]$c.Memory.Position
     $stream=[IO.File]::Open((Join-Path $folder ($c.Name+'.bin')),[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::None)
     try{$stream.Write($c.Bytes,0,$length);$stream.Flush()}finally{$stream.Dispose()}
     $capture.Available=$true;$capture.Length=$length;$capture.SHA256=(Get-FileHash (Join-Path $folder ($c.Name+'.bin'))).Hash
    }catch{$capture.Error=$_.Exception.Message;$row.Error=$capture.Error}
    $c.Source.Dispose();$c.Memory.Dispose()
   }
   $row.Channels+=,$capture
  }
  if($row.PendingResources -or ($row.Associated -and -not $row.OwnedExit)){$script:retained=@($p,$cts,$channels);$row.PendingResources=$true}else{$p.Dispose();$cts.Dispose()}
 }
 $row.ElapsedMs=$clock.ElapsedMilliseconds;$row.After=Map-IW supervisor.after-map.json
 # The role's full report is separate from the source-compile receipt.
 if(Test-Path (Join-Path $folder compile.role-result.json)){
  $worker=Get-Content (Join-Path $folder compile.role-result.json) -Raw|ConvertFrom-Json -AsHashtable
  $exact=$row.Channels.Count -eq 2 -and $row.Channels[0].Available -and $row.Channels[1].Available -and $row.Channels[1].Length -eq 0
  if($exact){$exact=[Text.Encoding]::UTF8.GetString([IO.File]::ReadAllBytes((Join-Path $folder stdout.bin))) -ceq "ISOLATED_WRITER_UNITS_VERIFIED`r`n"}
  $row.Passed=$row.Started -and $row.OwnedExit -and $row.ExitCode -eq 0 -and -not $row.KillRequested -and -not $row.PendingResources -and $null -eq $row.Error -and $exact -and $worker.Report.Passed -and $worker.Report.Cases.Count -eq 18 -and $null -eq $worker.Error
 }
 Journal-IW 2 end Compile;$row.EvidencePersisted=$true;Save-IW result.json $row
 Need-IW $row.Passed 'compiler/unit result unexpected'
 return $row
}
try{
 Need-IW ((Get-FileHash (Join-Path $base preparation.stop.json)).Hash -ceq '894DF77E82EFEEC69F233FBF85A0E439F0590C502529D49A2DBE07D1A3F5185B') 'original stop identity'
 Need-IW (-not(Test-Path (Join-Path $base preparation-v2.stop.json))) 'preparation stopped'
 Need-IW (-not(Test-Path (Join-Path $base execution-v2.stop.json))) 'execution stopped'
 if($Stage -ceq 'Compile'){
  $part='compile';$folder=Join-Path $base $part;$intent=Begin-IW 31 '' 1
 }else{
  Need-IW (-not(Test-Path (Join-Path $base gate.claim.json))) 'gate consumed'
  for($i=1;$i -le 9;$i++){Need-IW (-not(Test-Path (Join-Path $base ('scenarios/S{0:d2}' -f $i)))) 'scenario consumed'}
  $folder=$base;Save-IW gate.claim.json ([ordered]@{InputSHA256=$InputSHA256;AttemptConsumed=$true;Computations=@(32..40);MaxStarts=17})
 }
 Verify-IW
 ${function:Assert-HTInputs}=$script:iwHashFunction;${function:Get-HTHash}=$script:iwGetHashFunction
 if($Stage -ceq 'Compile'){$value=Invoke-CompilerIW;$value|ConvertTo-Json -Depth 10;exit 0}
 $compiled=Get-Content (Join-Path $base compile/result.json) -Raw|ConvertFrom-Json -AsHashtable
 Need-IW $compiled.Passed 'missing passed compiler'
 $checkpoint=Get-Content (Join-Path $base compiled-checkpoint.json) -Raw|ConvertFrom-Json -AsHashtable
 Need-IW ($checkpoint.RemoteEqualsLocal -and $checkpoint.InputSHA256 -ceq $InputSHA256) 'compiled checkpoint not remotely verified'
 $assembly=Get-Content (Join-Path $base compile/compile.result.json) -Raw|ConvertFrom-Json -AsHashtable
 Need-IW ((Get-FileHash (Join-Path $base isolated-writer.dll)).Hash -ceq $assembly.SHA256) 'compiled DLL hash'
 $null=[Reflection.Assembly]::LoadFrom((Join-Path $script:HTRepo $binding.AdapterPath))
 $null=[Reflection.Assembly]::LoadFrom((Join-Path $base isolated-writer.dll))
 $gate=[Collections.Generic.List[object]]::new()
 for($i=1;$i -le 9;$i++){
  $scenario='S{0:d2}' -f $i;$part="scenarios/$scenario";$folder=Join-Path $base $part
  $max=if($i -eq 3){1}else{2};$intent=Begin-IW (31+$i) $scenario $max
  $before=Map-IW supervisor.before-map.json;$null=New-Item -ItemType Directory -Path (Join-Path $folder data)
  Journal-IW 1 begin $scenario
  $options=[Talenro.Diagnostics.IsolatedWriterV1.ScenarioOptions]::new()
  $options.Scenario=$scenario;$options.Token=$intent.Token;$options.InputSHA256=$InputSHA256;$options.IntentSHA256=(Get-FileHash (Join-Path $folder intent.json)).Hash;$options.Engine=$engine;$options.RoleScript=$rolePath;$options.Root=$script:HTRepo
  $task=[Talenro.Diagnostics.IsolatedWriterV1.Supervisor]::RunAsync($options)
  while(-not $task.IsCompleted){[Threading.Thread]::Sleep(1)}
  $result=$task.GetAwaiter().GetResult()
  $after=Map-IW supervisor.after-map.json
  $pass=$result.NextAllowed -and -not $result.PendingResources -and -not $result.ExitUnconfirmed -and $result.Writer.OwnedExit -and ($i -eq 3 -or $result.Producer.OwnedExit)
  if($i -le 2){
   $a=if($i -eq 1){8193}else{0};$b=if($i -eq 1){48}else{0}
   $pass=$pass -and $result.SupervisionOutcome -ceq 'SUCCESS' -and $result.Stdout.Available -and $result.Stderr.Available -and $result.Stdout.Length -eq $a -and $result.Stderr.Length -eq $b
  }else{
   $reason=switch($i){3{'ReadyTimeout'}4{'DrainTimeout'}5{'DrainTimeout'}6{'WriterExited','TransportFailure'}7{'ProducerFailed'}8{'CapacityExceeded'}9{'MissingReceipt'}}
   $pass=$pass -and $result.SupervisionOutcome -ceq 'FAILED' -and $result.Reason -cin $reason -and -not $result.Stdout.Available -and -not $result.Stderr.Available -and $null -eq $result.Stdout.Length -and $null -eq $result.Stdout.SHA256 -and $null -eq $result.Stderr.Length -and $null -eq $result.Stderr.SHA256
   if($i -eq 3){$pass=$pass -and -not $result.Producer.StartAttempted -and $result.Writer.KillRequested}
   if($i -in @(4,5)){$pass=$pass -and $result.Writer.KillRequested -and $result.Producer.ExitCode -eq 0}
   if($i -eq 6){$pass=$pass -and $result.Writer.ExitCode -eq 23}
   if($i -eq 7){$pass=$pass -and $result.Producer.ExitCode -eq 23}
   if($i -eq 8){$pass=$pass -and $result.Relay.Read[0] -eq 65537 -and $result.Relay.Sent[0] -le 65536}
   if($i -eq 9){$pass=$pass -and $result.Writer.ExitCode -eq 0 -and $result.Producer.ExitCode -eq 0}
  }
  $result.EvidencePersisted=$true
  Journal-IW 2 end $scenario
  Save-IW result.json ([ordered]@{Scenario=$scenario;Computation=(31+$i);Passed=$pass;Before=$before;After=$after;Result=$result})
  $gate.Add([ordered]@{Scenario=$scenario;Passed=$pass;ResultSHA256=(Get-FileHash (Join-Path $folder result.json)).Hash})
  $saved=$folder;$folder=$base;Save-IW ('gate-through-'+$scenario+'.json') $gate.ToArray();$folder=$saved
  Need-IW $pass "unexpected $scenario result"
 }
 $folder=$base;Save-IW gate.result.json ([ordered]@{Passed=$true;Cases=$gate.ToArray();Status='NOT_ACCEPTED'})
 [Console]::Out.WriteLine('ISOLATED_WRITER_SYNTHETIC_GATE_VERIFIED')
}catch{
 $errorRecord=[ordered]@{Stage=$Stage;Status='STOPPED';Message=$_.Exception.Message;Stack=$_.ScriptStackTrace;NoRetry=$true;Remaining='NOT_RUN';InputSHA256=$InputSHA256}
 $folder=$base
 if(-not(Test-Path (Join-Path $base execution-v2.stop.json))){Save-IW execution-v2.stop.json $errorRecord}
 $errorRecord|ConvertTo-Json -Depth 10
 exit 1
}
