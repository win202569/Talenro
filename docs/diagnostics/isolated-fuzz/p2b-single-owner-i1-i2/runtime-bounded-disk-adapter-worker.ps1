param(
 [Parameter(Mandatory)][ValidatePattern('^[A-F0-9]{64}$')][string]$InputSHA256,
 [Parameter(Mandatory)][ValidatePattern('^[A-F0-9]{64}$')][string]$IntentSHA256
)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$script:HTRepo='D:\Projects\Talenro\.worktrees\c12-b01-task9-coordinator'
if((Get-Location).Path -cne $script:HTRepo){throw 'BA fixed root'}
$d=Join-Path $script:HTRepo '.superpowers/sdd/2026-11-04-runtime-bounded-disk-adapter-v1'
function Save-BA([string]$Name,$Value){
 $bytes=[Text.UTF8Encoding]::new($false).GetBytes((ConvertTo-Json -InputObject $Value -Depth 100))
 $limit=if($Name -cin @('worker.before-map.json','worker.after-map.json')){2097152}else{262144}
 if($bytes.Length -gt $limit){throw 'BA receipt bound'}
 $s=[IO.File]::Open((Join-Path $d $Name),[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::None)
 try{$s.Write($bytes,0,$bytes.Length);$s.Flush()}finally{$s.Dispose()}
}
function Save-BAMap([string]$Name,$Map){
 Save-BA $Name $Map
 [ordered]@{Path=('.superpowers/sdd/2026-11-04-runtime-bounded-disk-adapter-v1/'+$Name);Count=$Map.Count;Length=(Get-Item (Join-Path $d $Name)).Length;SHA256=(Get-FileHash (Join-Path $d $Name)).Hash}
}
function Error-BA($r){
 $message=$r.Exception.Message;$stack=$r.ScriptStackTrace
 if($null -eq $stack){$stack=''}
 [ordered]@{Type=$r.Exception.GetType().FullName;Message=$message.Substring(0,[Math]::Min(4096,$message.Length));Stack=$stack.Substring(0,[Math]::Min(4096,$stack.Length));Truncated=($message.Length -gt 4096 -or $stack.Length -gt 4096)}
}
function Stage-BA([int]$Seq,[string]$Kind){
 $bytes=[Text.UTF8Encoding]::new($false).GetBytes((ConvertTo-Json -Compress ([ordered]@{Schema=1;Seq=$Seq;Stage='compile';Kind=$Kind;Case='';ElapsedTicks=[Diagnostics.Stopwatch]::GetTimestamp();Frequency=[Diagnostics.Stopwatch]::Frequency}))+[char]10)
 $mode=if($Seq -eq 1){[IO.FileMode]::CreateNew}else{[IO.FileMode]::Append}
 $s=[IO.File]::Open((Join-Path $d stages.jsonl),$mode,[IO.FileAccess]::Write,[IO.FileShare]::Read)
 try{if($s.Length+$bytes.Length -gt 262144){throw 'BA journal bound'};$s.Write($bytes,0,$bytes.Length);$s.Flush()}finally{$s.Dispose()}
}
$row=[ordered]@{Schema=1;Status='ADAPTER_FAILED';InputSHA256=$InputSHA256;IntentSHA256=$IntentSHA256;Before=$null;After=$null;Report=$null;Error=$null;CompileAttempted=$false;CompileSucceeded=$false;AssemblySHA256=$null;AssemblyLength=$null;ExitCode=1;SourceVerified=$false;LiveActionsPerformed=$false;RuntimeApproved=$false;FormalAcceptance=$false;RealGateImplemented=$false}
try{
 foreach($n in @('worker.claim.json','compile.intent.json','compile.result.json','adapter.dll','stages.jsonl','verification-result.json','cases')){if(Test-Path (Join-Path $d $n)){throw 'BA occupied'}}
 if((Get-FileHash (Join-Path $d input.json)).Hash -cne $InputSHA256 -or (Get-FileHash (Join-Path $d intent.json)).Hash -cne $IntentSHA256){throw 'BA identity'}
 $binding=Get-Content (Join-Path $d input.json) -Raw|ConvertFrom-Json -AsHashtable
 $intent=Get-Content (Join-Path $d intent.json) -Raw|ConvertFrom-Json -AsHashtable
 if($intent.Computation -ne 30 -or $intent.Ceiling -ne 30 -or $intent.MaxProcesses -ne 1 -or -not $intent.AttemptConsumed -or $intent.InputSHA256 -cne $InputSHA256){throw 'BA intent'}
 Save-BA worker.claim.json ([ordered]@{Computation=30;InputSHA256=$InputSHA256;IntentSHA256=$IntentSHA256;Token=$intent.Token})
 . (Join-Path $PSScriptRoot runtime-copy-transport-common.ps1)
 $row.Before=Save-BAMap worker.before-map.json (Assert-HTInputs $binding)
 $identity=[Security.Principal.WindowsIdentity]::GetCurrent()
 $principal=[Security.Principal.WindowsPrincipal]::new($identity)
 $engine=[Environment]::ProcessPath
 if($PSVersionTable.PSVersion.ToString() -cne '7.6.5' -or $PSVersionTable.PSEdition -cne 'Core' -or -not [Environment]::Is64BitProcess -or $identity.Name -cne 'DESKTOP-H4SVCLU\Lenovo' -or $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)){throw 'BA runtime context'}
 if((Get-FileHash $engine).Hash -cne $binding.EngineSHA256 -or (Get-FileHash (Join-Path $PSHOME System.Management.Automation.dll)).Hash -cne $binding.SMASHA256){throw 'BA runtime hashes'}
 if('Talenro.Diagnostics.BoundedDiskV1.BoundedFileSink' -as [type]){throw 'BA existing type'}
 Save-BA compile.intent.json ([ordered]@{Computation=30;Compilation=4;AttemptConsumed=$true;Sources=$binding.CompilationSources;InputSHA256=$InputSHA256})
 $row.CompileAttempted=$true
 Stage-BA 1 begin
 $sourcePaths=@($binding.CompilationSources|ForEach-Object {Join-Path $script:HTRepo $_})
 $dll=Join-Path $d adapter.dll
 Add-Type -LiteralPath $sourcePaths -OutputAssembly $dll -ErrorAction Stop
 $row.CompileSucceeded=$true
 $row.AssemblyLength=(Get-Item $dll).Length
 if($row.AssemblyLength -gt 16777216){throw 'BA assembly size'}
 $row.AssemblySHA256=(Get-FileHash $dll).Hash
 Save-BA compile.result.json ([ordered]@{Compilation=4;Succeeded=$true;Length=$row.AssemblyLength;SHA256=$row.AssemblySHA256;InputSHA256=$InputSHA256})
 Stage-BA 2 end
 Add-Type -LiteralPath $dll -ErrorAction Stop
 $task=[Talenro.Diagnostics.BoundedDiskV1.Verification]::RunAsync((Join-Path $d cases))
 $report=$task.GetAwaiter().GetResult()
 $row.Report=$report
 if($report.Status -ceq 'ADAPTER_VERIFIED' -and $report.CaseResults.Count -eq 24 -and @($report.CaseResults|Where-Object {-not $_.Passed}).Count -eq 0 -and $report.PendingOperations -eq 0 -and $report.UnavailablePaths.Count -eq 0 -and $null -eq $report.Error){
  $row.Status='ADAPTER_VERIFIED';$row.ExitCode=0
 }
}catch{
 $row.Error=Error-BA $_;$row.Status='ADAPTER_FAILED';$row.ExitCode=1
 if($row.CompileAttempted -and -not(Test-Path (Join-Path $d compile.result.json))){
  try{Save-BA compile.result.json ([ordered]@{Compilation=4;Succeeded=$false;Error=$row.Error})}catch{}
 }
}finally{
 if($null -ne $row.Before){
  try{$row.After=Save-BAMap worker.after-map.json (Assert-HTInputs $binding)}catch{$row.Error=Error-BA $_;$row.Status='ADAPTER_FAILED';$row.ExitCode=1}
 }
}
Save-BA verification-result.json $row
if($row.Status -cne 'ADAPTER_VERIFIED'){Save-BA worker.stop.json ([ordered]@{Status=$row.Status;Error=$row.Error;NoRetry=$true})}
[Console]::Out.WriteLine($row.Status)
exit $row.ExitCode

