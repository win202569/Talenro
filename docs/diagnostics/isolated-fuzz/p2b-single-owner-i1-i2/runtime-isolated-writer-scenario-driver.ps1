param(
 [Parameter(Mandatory)][ValidateSet('Scenarios')][string]$Stage,
 [Parameter(Mandatory)][ValidatePattern('^[A-F0-9]{64}$')][string]$InputSHA256
)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$script:HTRepo='D:\Projects\Talenro\.worktrees\c12-b01-task9-coordinator'
$relative='.superpowers/sdd/2026-11-05-runtime-isolated-writer-scenarios-v1'
$base=Join-Path $script:HTRepo $relative
$engine='C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe'
$rolePath=Join-Path $PSScriptRoot runtime-isolated-writer-scenario-role.ps1
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
try{
 Need-IW (-not(Test-Path (Join-Path $base preparation.stop.json))) 'preparation stopped'
 Need-IW (-not(Test-Path (Join-Path $base execution.stop.json))) 'execution stopped'
 Need-IW (-not(Test-Path (Join-Path $base gate.claim.json))) 'gate consumed'
 for($i=1;$i -le 9;$i++){Need-IW (-not(Test-Path (Join-Path $base ('scenarios/S{0:d2}' -f $i)))) 'scenario consumed'}
 Save-IW gate.claim.json ([ordered]@{InputSHA256=$InputSHA256;AttemptConsumed=$true;Computations=@(33..41);MaxStarts=17;MaxCompilations=0})
 Verify-IW
 ${function:Assert-HTInputs}=$script:iwHashFunction;${function:Get-HTHash}=$script:iwGetHashFunction
 $compiled=Get-Content (Join-Path $script:HTRepo $binding.CompiledResultPath) -Raw|ConvertFrom-Json -AsHashtable
 Need-IW $compiled.Passed 'prior compiler not passed'
 Need-IW ((Get-FileHash (Join-Path $script:HTRepo $binding.AssemblyPath)).Hash -ceq $binding.AssemblySHA256) 'assembly identity'
 $null=[Reflection.Assembly]::LoadFrom((Join-Path $script:HTRepo $binding.AdapterPath))
 $null=[Reflection.Assembly]::LoadFrom((Join-Path $script:HTRepo $binding.AssemblyPath))
 $gate=[Collections.Generic.List[object]]::new()
 for($i=1;$i -le 9;$i++){
  $scenario='S{0:d2}' -f $i;$part="scenarios/$scenario";$folder=Join-Path $base $part
  $max=if($i -eq 3){1}else{2};$intent=Begin-IW (32+$i) $scenario $max
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
  # Evidence checks occur only after the supervisor has frozen the outcome.
  $writerMaps=@()
  foreach($name in @('writer.before-map.json','writer.after-map.json')){
   $mp=Join-Path $folder $name
   if(Test-Path $mp){
    $observed=Get-Content $mp -Raw|ConvertFrom-Json -AsHashtable
    $expectedMap=Get-Content (Join-Path $script:HTRepo $after.Path) -Raw|ConvertFrom-Json -AsHashtable
    if($observed.Count -ne $expectedMap.Count){$pass=$false}
    foreach($key in $expectedMap.Keys){if($observed[$key] -cne $expectedMap[$key]){$pass=$false}}
    $writerMaps+=,[ordered]@{Path="$relative/$part/$name";Available=$true;Count=$observed.Count;Length=(Get-Item $mp).Length;SHA256=(Get-FileHash $mp).Hash}
   }else{
    $allowed=$name -ceq 'writer.after-map.json' -and $result.Writer.KillRequested -and $result.Writer.OwnedExit
    if(-not $allowed){$pass=$false}
    $writerMaps+=,[ordered]@{Path="$relative/$part/$name";Available=$false;AllowedTerminationAbsence=$allowed}
   }
  }
  if($i -le 2 -and $result.SupervisionOutcome -ceq 'SUCCESS'){
   foreach($channel in @('stdout','stderr')){
    $expectedLength=if($i -eq 2){0}elseif($channel -ceq 'stdout'){8193}else{48}
    $capture=Join-Path $folder ("data/"+$channel+".bin")
    Need-IW ((Get-Item $capture).Length -eq $expectedLength) 'stable file length'
    $bytes=[IO.File]::ReadAllBytes($capture)
    for($j=0;$j -lt $bytes.Length;$j++){if($bytes[$j] -ne ($j%251)){$pass=$false}}
    $receipt=if($channel -ceq 'stdout'){$result.Stdout}else{$result.Stderr}
    if((Get-FileHash $capture).Hash -cne $receipt.SHA256){$pass=$false}
   }
  }
  $result.EvidencePersisted=$true
  Journal-IW 2 end $scenario
  Save-IW result.json ([ordered]@{Scenario=$scenario;Computation=(32+$i);Passed=$pass;Before=$before;After=$after;WriterMaps=$writerMaps;Result=$result})
  $gate.Add([ordered]@{Scenario=$scenario;Passed=$pass;ResultSHA256=(Get-FileHash (Join-Path $folder result.json)).Hash})
  $saved=$folder;$folder=$base;Save-IW ('gate-through-'+$scenario+'.json') $gate.ToArray();$folder=$saved
  Need-IW $pass "unexpected $scenario result"
 }
 $folder=$base;Save-IW gate.result.json ([ordered]@{Passed=$true;Cases=$gate.ToArray();Status='NOT_ACCEPTED'})
 [Console]::Out.WriteLine('ISOLATED_WRITER_SYNTHETIC_GATE_VERIFIED')
}catch{
 $errorRecord=[ordered]@{Stage=$Stage;Status='STOPPED';Message=$_.Exception.Message;Stack=$_.ScriptStackTrace;NoRetry=$true;Remaining='NOT_RUN';InputSHA256=$InputSHA256}
 $folder=$base
 if(-not(Test-Path (Join-Path $base execution.stop.json))){Save-IW execution.stop.json $errorRecord}
 $errorRecord|ConvertTo-Json -Depth 10
 exit 1
}
