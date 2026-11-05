param(
 [Parameter(Mandatory)][ValidatePattern('^[A-F0-9]{64}$')][string]$InputSHA256
)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$script:HTRepo='D:\Projects\Talenro\.worktrees\c12-b01-task9-coordinator'
$relative='.superpowers/sdd/2026-11-05-runtime-staged-integrity-v1'
$base=Join-Path $script:HTRepo $relative
$engine='C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe'
$rolePath=Join-Path $PSScriptRoot runtime-staged-integrity-role.ps1
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
 foreach($name in @('preparation.stop.json','execution.stop.json','gate.claim.json','intent.json','data','child.claim.json','writer.claim.json','producer.claim.json')){Need-IW (-not(Test-Path (Join-Path $base $name))) "occupied $name"}
 Verify-IW
 ${function:Assert-HTInputs}=$script:iwHashFunction;${function:Get-HTHash}=$script:iwGetHashFunction
 Need-IW ((Get-FileHash (Join-Path $script:HTRepo $binding.AdapterPath)).Hash -ceq $binding.AdapterSHA256) 'adapter identity'
 Need-IW ((Get-FileHash (Join-Path $script:HTRepo $binding.AssemblyPath)).Hash -ceq $binding.AssemblySHA256) 'assembly identity'
 $compiled=Get-Content (Join-Path $script:HTRepo $binding.CompiledResultPath) -Raw|ConvertFrom-Json -AsHashtable
 Need-IW ($compiled.Passed -and $compiled.OwnedExit -and $compiled.ExitCode -eq 0 -and -not $compiled.PendingResources) 'prior compiler'
 Save-IW gate.claim.json ([ordered]@{InputSHA256=$InputSHA256;AttemptConsumed=$true;Computation=44;MaxStarts=2;MaxCompilations=0})
 $intent=[ordered]@{Schema=1;Computation=44;Ceiling=44;Scenario='StagedIntegrity44';MaxStarts=2;MaxProcesses=2;AttemptConsumed=$true;Token=[guid]::NewGuid().ToString('D');InputSHA256=$InputSHA256}
 Save-IW intent.json $intent
 $before=Map-IW supervisor.before-map.json
 $null=New-Item -ItemType Directory -Path (Join-Path $folder data)
 $null=[Reflection.Assembly]::LoadFrom((Join-Path $script:HTRepo $binding.AdapterPath))
 $null=[Reflection.Assembly]::LoadFrom((Join-Path $script:HTRepo $binding.AssemblyPath))
 Journal-IW 1 begin StagedIntegrity44
 $options=[Talenro.Diagnostics.IsolatedWriterV1.ScenarioOptions]::new()
 $options.Scenario='StagedIntegrity44';$options.Token=$intent.Token;$options.InputSHA256=$InputSHA256;$options.IntentSHA256=(Get-FileHash (Join-Path $folder intent.json)).Hash;$options.Engine=$engine;$options.RoleScript=$rolePath;$options.Root=$script:HTRepo
 $task=[Talenro.Diagnostics.IsolatedWriterV1.Supervisor]::RunAsync($options)
 while(-not $task.IsCompleted){[Threading.Thread]::Sleep(1)}
 $retained=$task.GetAwaiter().GetResult()
 # Preserve the terminal outcome before any post-run evidence check can fail.
 Save-IW supervision.json $retained
 Journal-IW 2 end StagedIntegrity44
 $terminal=$retained.NextAllowed -and -not $retained.PendingResources -and -not $retained.ExitUnconfirmed
 Need-IW $terminal 'pending resource or unknown exit; no further inspection'
 $audit=[ordered]@{Schema=1;Scenario='StagedIntegrity44';InputSHA256=$InputSHA256;IntentSHA256=$options.IntentSHA256;Before=$before;After=$null;State='FAILED';ElapsedMilliseconds=$null;AbsentPaths=$null;Error=$null}
 $auditClock=[Diagnostics.Stopwatch]::StartNew()
 try{
  $after=Map-IW supervisor.after-map.json
  $audit.After=$after;$audit.AbsentPaths=@($binding.AbsentPaths);$audit.State='VERIFIED'
 }catch{
  $audit.After=$null;$audit.State='FAILED';$audit.Error=$_.Exception.Message
 }finally{$auditClock.Stop();$audit.ElapsedMilliseconds=$auditClock.Elapsed.TotalMilliseconds}
 Save-IW integrity-review.json $audit
 Need-IW ($audit.State -ceq 'VERIFIED') 'post-exit integrity audit failed'
 Save-IW result.json ([ordered]@{Computation=44;Scenario='StagedIntegrity44';Before=$before;After=$after;Result=$retained;BusinessVerified=$false;Status='NOT_ACCEPTED'})
 Need-IW ($retained.SupervisionOutcome -ceq 'SUCCESS' -and $retained.Writer.OwnedExit -and $retained.Producer.OwnedExit -and $retained.Writer.ExitCode -eq 0 -and $retained.Producer.ExitCode -eq 0) 'transport did not succeed'
 [Console]::Out.WriteLine('BUSINESS_PREFIX_TRANSPORT_COMPLETED_REVIEW_REQUIRED')
}catch{
 $errorRecord=[ordered]@{Status='STOPPED';Message=$_.Exception.Message;Stack=$_.ScriptStackTrace;NoRetry=$true;InputSHA256=$InputSHA256}
 if(-not(Test-Path (Join-Path $base execution.stop.json))){Save-IW execution.stop.json $errorRecord}
 $errorRecord|ConvertTo-Json -Depth 10
 exit 1
}
