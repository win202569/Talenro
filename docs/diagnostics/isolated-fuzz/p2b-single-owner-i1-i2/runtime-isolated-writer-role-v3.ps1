param(
 [Parameter(Mandatory)][ValidateSet('Compile')][string]$Role,
 [ValidatePattern('^S0[1-9]$')][string]$Scenario,
 [Parameter(Mandatory)][ValidatePattern('^[A-F0-9]{64}$')][string]$InputSHA256,
 [Parameter(Mandatory)][ValidatePattern('^[A-F0-9]{64}$')][string]$IntentSHA256
)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$script:HTRepo='D:\Projects\Talenro\.worktrees\c12-b01-task9-coordinator'
$relative='.superpowers/sdd/2026-11-05-runtime-isolated-writer-v3'
$base=Join-Path $script:HTRepo $relative
$part=if($Role -ceq 'Compile'){'compile'}else{"scenarios/$Scenario"}
$folder=Join-Path $base $part
function Save-IW([string]$Name,$Value){
 $bytes=[Text.UTF8Encoding]::new($false).GetBytes((ConvertTo-Json -InputObject $Value -Depth 80))
 $limit=if($Name.EndsWith('-map.json')){2097152}else{262144}
 if($bytes.Length -gt $limit){throw 'IW record bound'}
 $stream=[IO.File]::Open((Join-Path $folder $Name),[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::None)
 try{$stream.Write($bytes,0,$bytes.Length);$stream.Flush()}finally{$stream.Dispose()}
}
function Map-IW([string]$Name){
 $map=Assert-HTInputs $binding;Save-IW $Name $map
 return [ordered]@{Path="$relative/$part/$Name";Count=$map.Count;Length=(Get-Item (Join-Path $folder $Name)).Length;SHA256=(Get-FileHash (Join-Path $folder $Name)).Hash}
}
$row=[ordered]@{Schema=1;Role=$Role;Scenario=$Scenario;InputSHA256=$InputSHA256;IntentSHA256=$IntentSHA256;Before=$null;After=$null;Report=$null;CompileAttempted=$false;CompileSucceeded=$false;AssemblySHA256=$null;AssemblyLength=$null;ExitCode=1;Error=$null;Status='NOT_ACCEPTED'}
$binding=$null;$claimed=$false
try{
 if((Get-Location).Path -cne $script:HTRepo){throw 'IW fixed root'}
 if((Get-FileHash (Join-Path $base input.json)).Hash -cne $InputSHA256 -or (Get-FileHash (Join-Path $folder intent.json)).Hash -cne $IntentSHA256){throw 'IW input/intent identity'}
 $binding=Get-Content (Join-Path $base input.json) -Raw|ConvertFrom-Json -AsHashtable
 $intent=Get-Content (Join-Path $folder intent.json) -Raw|ConvertFrom-Json -AsHashtable
 $expected=if($Role -ceq 'Compile'){32}else{32+[int]$Scenario.Substring(1)}
 if($intent.InputSHA256 -cne $InputSHA256 -or $intent.Computation -ne $expected -or -not $intent.AttemptConsumed -or $intent.Token -cnotmatch '^[0-9a-f-]{36}$'){throw 'IW claim identity'}
 if($Role -cne 'Compile' -and $intent.Scenario -cne $Scenario){throw 'IW scenario identity'}
 $claimName=$Role.ToLowerInvariant()+'.claim.json'
 Save-IW $claimName ([ordered]@{Role=$Role;Computation=$expected;Token=$intent.Token;InputSHA256=$InputSHA256;IntentSHA256=$IntentSHA256});$claimed=$true
 $identity=[Security.Principal.WindowsIdentity]::GetCurrent()
 $principal=[Security.Principal.WindowsPrincipal]::new($identity)
 if($PSVersionTable.PSVersion.ToString() -cne '7.6.5' -or $PSVersionTable.PSEdition -cne 'Core' -or -not [Environment]::Is64BitProcess -or $identity.Name -cne 'DESKTOP-H4SVCLU\Lenovo' -or $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)){throw 'IW runtime context'}
 if((Get-FileHash ([Environment]::ProcessPath)).Hash -cne $binding.EngineSHA256 -or (Get-FileHash (Join-Path $PSHOME System.Management.Automation.dll)).Hash -cne $binding.SMASHA256){throw 'IW runtime identity'}
 foreach($p in $binding.HelperSHA256.Keys){if((Get-FileHash (Join-Path $script:HTRepo $p)).Hash -cne $binding.HelperSHA256[$p]){throw "IW helper drift $p"}}
 $common='docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/runtime-copy-transport-common.ps1'
 if((Get-FileHash (Join-Path $script:HTRepo $common)).Hash -cne $binding.HistoricalEvidenceSHA256[$common]){throw 'IW common identity'}
 . (Join-Path $script:HTRepo $common)
 if($Role -cne 'Producer'){$row.Before=Map-IW $(if($Role -ceq 'Compile'){'worker.before-map.json'}else{'writer.before-map.json'})}
 $adapter=Join-Path $script:HTRepo $binding.AdapterPath
 if((Get-FileHash $adapter).Hash -cne $binding.AdapterSHA256 -or (Get-Item $adapter).Length -ne 50176){throw 'IW adapter identity'}
 $null=[Reflection.Assembly]::LoadFrom($adapter)
 $dll=Join-Path $base isolated-writer.dll
 if($Role -ceq 'Compile'){
  if(Test-Path $dll){throw 'IW assembly already exists'}
  Save-IW compile.intent.json ([ordered]@{Compilation=6;Computation=32;AttemptConsumed=$true;Sources=$binding.CompilationSources;InputSHA256=$InputSHA256;ReferenceSHA256=$binding.AdapterSHA256;Command='Add-Type -LiteralPath $sources -ReferencedAssemblies $references -OutputAssembly $dll -ErrorAction Stop'})
  $row.CompileAttempted=$true
  $sources=@($binding.CompilationSources|ForEach-Object {Join-Path $script:HTRepo $_})
  $references=@(Get-ChildItem -LiteralPath (Join-Path $PSHOME ref) -Filter '*.dll'|Sort-Object Name|ForEach-Object {$_.FullName})+@($adapter)
  Add-Type -LiteralPath $sources -ReferencedAssemblies $references -OutputAssembly $dll -ErrorAction Stop
  $row.CompileSucceeded=$true;$row.AssemblyLength=(Get-Item $dll).Length;$row.AssemblySHA256=(Get-FileHash $dll).Hash
  if($row.AssemblyLength -gt 1048576){throw 'IW assembly bound'}
  Save-IW compile.result.json ([ordered]@{Compilation=6;Succeeded=$true;SHA256=$row.AssemblySHA256;Length=$row.AssemblyLength;ReferenceSHA256=$binding.AdapterSHA256})
 }else{
  $compiled=Get-Content (Join-Path $base compile/compile.result.json) -Raw|ConvertFrom-Json -AsHashtable
  if(-not $compiled.Succeeded -or (Get-FileHash $dll).Hash -cne $compiled.SHA256){throw 'IW assembly mismatch'}
 }
 $null=[Reflection.Assembly]::LoadFrom($dll)
 if($Role -ceq 'Compile'){
  $task=[Talenro.Diagnostics.IsolatedWriterV1.ProtocolTests]::RunAsync()
 }elseif($Role -ceq 'Writer'){
  $options=[Talenro.Diagnostics.IsolatedWriterV1.WriterOptions]::new();$options.Scenario=$Scenario;$options.Token=$intent.Token;$options.InputSHA256=$InputSHA256;$options.Directory=Join-Path $folder data
  $task=[Talenro.Diagnostics.IsolatedWriterV1.Writer]::RunAsync($options)
 }else{
  $options=[Talenro.Diagnostics.IsolatedWriterV1.ProducerOptions]::new();$options.Scenario=$Scenario
  $task=[Talenro.Diagnostics.IsolatedWriterV1.Producer]::RunAsync($options)
 }
 # Role work may deliberately remain pending; the owning parent supplies the deadline.
 while(-not $task.IsCompleted){[Threading.Thread]::Sleep(1)}
 $value=$task.GetAwaiter().GetResult()
 if($Role -ceq 'Compile'){
  $row.Report=$value
  if(-not $value.Passed -or $value.Cases.Count -ne 18){throw 'IW unit gate failed'}
  $row.ExitCode=0
 }else{$row.ExitCode=[int]$value}
}catch{
 $row.Error=[ordered]@{Message=$_.Exception.Message;Stack=$_.ScriptStackTrace;Detail=$_.ToString()};$row.ExitCode=1
 if($Role -ceq 'Compile' -and $row.CompileAttempted -and -not(Test-Path (Join-Path $folder compile.result.json))){Save-IW compile.result.json ([ordered]@{Compilation=6;Succeeded=$false;Error=$row.Error})}
}finally{
 if($null -ne $row.Before){
  try{$row.After=Map-IW $(if($Role -ceq 'Compile'){'worker.after-map.json'}else{'writer.after-map.json'})}catch{$row.Error=[ordered]@{Message=$_.Exception.Message;Stack=$_.ScriptStackTrace};$row.ExitCode=1}
 }
}
if($claimed){Save-IW ($Role.ToLowerInvariant()+'.role-result.json') $row}
if($Role -ceq 'Compile'){
 if($row.ExitCode -ne 0){Save-IW stop.json ([ordered]@{Status='COMPILE_OR_UNIT_STOPPED';Error=$row.Error;NoRetry=$true})}
 [Console]::Out.WriteLine($(if($row.ExitCode -eq 0){'ISOLATED_WRITER_UNITS_VERIFIED'}else{'ISOLATED_WRITER_UNITS_FAILED'}))
}
exit $row.ExitCode
