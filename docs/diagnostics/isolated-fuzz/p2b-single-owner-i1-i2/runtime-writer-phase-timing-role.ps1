param(
 [Parameter(Mandatory)][ValidateSet('Writer','Producer')][string]$Role,
 [Parameter(Mandatory)][ValidateSet('WriterPhase43')][string]$Scenario,
 [Parameter(Mandatory)][ValidatePattern('^[A-F0-9]{64}$')][string]$InputSHA256,
 [Parameter(Mandatory)][ValidatePattern('^[A-F0-9]{64}$')][string]$IntentSHA256
)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$script:HTRepo='D:\Projects\Talenro\.worktrees\c12-b01-task9-coordinator'
$relative='.superpowers/sdd/2026-11-05-runtime-writer-phase-timing-v1'
$base=Join-Path $script:HTRepo $relative
$part='';$folder=$base
function Save-IW([string]$Name,$Value){
 $bytes=[Text.UTF8Encoding]::new($false).GetBytes((ConvertTo-Json -InputObject $Value -Depth 80))
 $limit=if($Name.EndsWith('-map.json')){2097152}else{262144}
 if($bytes.Length -gt $limit){throw 'IW scenario record bound'}
 $s=[IO.File]::Open((Join-Path $folder $Name),[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::None)
 try{$s.Write($bytes,0,$bytes.Length);$s.Flush()}finally{$s.Dispose()}
}
$script:WPTState=$null;$script:WPTClock=$null;$script:WPTStream=$null
function Emit-WPT([string]$Phase){
 if($null -eq $script:WPTState -or $script:WPTState.Failed -or $script:WPTState.Terminal){return}
 try{$null=Write-WPTRecord $script:WPTState $script:WPTStream $Phase $script:WPTClock.ElapsedTicks}catch{$script:WPTState.Failed=$true}
}
function Map-IW([string]$Name){
 $timed=($Role -ceq 'Writer' -and $Name -ceq 'writer.after-map.json')
 if($timed){Emit-WPT pins.begin}
 $map=Assert-HTInputs $binding
 if($timed){Emit-WPT pins.end;Emit-WPT map.save.begin}
 Save-IW $Name $map
 if($timed){Emit-WPT map.save.end}
 $metadata=[ordered]@{Path="$relative/$part/$Name";Count=$map.Count;Length=(Get-Item (Join-Path $folder $Name)).Length;SHA256=(Get-FileHash (Join-Path $folder $Name)).Hash}
 if($timed){Emit-WPT map.meta.end}
 return $metadata
}
$row=[ordered]@{Role=$Role;Scenario=$Scenario;InputSHA256=$InputSHA256;IntentSHA256=$IntentSHA256;Before=$null;After=$null;ExitCode=1;Error=$null;Status='NOT_ACCEPTED'}
$claimed=$false;$binding=$null
try{
 if((Get-Location).Path -cne $script:HTRepo){throw 'IW fixed root'}
 if((Get-FileHash (Join-Path $base input.json)).Hash -cne $InputSHA256 -or (Get-FileHash (Join-Path $folder intent.json)).Hash -cne $IntentSHA256){throw 'IW input/intent identity'}
 $binding=Get-Content (Join-Path $base input.json) -Raw|ConvertFrom-Json -AsHashtable
 $intent=Get-Content (Join-Path $folder intent.json) -Raw|ConvertFrom-Json -AsHashtable
 $expected=43
 if($intent.Computation -ne $expected -or $intent.Ceiling -ne 43 -or $intent.MaxProcesses -ne 2 -or $intent.MaxStarts -ne 2 -or $intent.Scenario -cne $Scenario -or $intent.InputSHA256 -cne $InputSHA256 -or -not $intent.AttemptConsumed -or $intent.Token -cnotmatch '^[0-9a-f-]{36}$'){throw 'IW intent'}
 Save-IW ($Role.ToLowerInvariant()+'.claim.json') ([ordered]@{Role=$Role;Computation=$expected;Token=$intent.Token;InputSHA256=$InputSHA256;IntentSHA256=$IntentSHA256});$claimed=$true
 $identity=[Security.Principal.WindowsIdentity]::GetCurrent();$principal=[Security.Principal.WindowsPrincipal]::new($identity)
 if($PSVersionTable.PSVersion.ToString() -cne '7.6.5' -or $PSVersionTable.PSEdition -cne 'Core' -or -not [Environment]::Is64BitProcess -or $identity.Name -cne 'DESKTOP-H4SVCLU\Lenovo' -or $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)){throw 'IW runtime context'}
 if((Get-FileHash ([Environment]::ProcessPath)).Hash -cne $binding.EngineSHA256 -or (Get-FileHash (Join-Path $PSHOME System.Management.Automation.dll)).Hash -cne $binding.SMASHA256){throw 'IW runtime identity'}
 foreach($p in $binding.HelperSHA256.Keys){if((Get-FileHash (Join-Path $script:HTRepo $p)).Hash -cne $binding.HelperSHA256[$p]){throw "IW helper drift $p"}}
 $common='docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/runtime-copy-transport-common.ps1'
 if((Get-FileHash (Join-Path $script:HTRepo $common)).Hash -cne $binding.HistoricalEvidenceSHA256[$common]){throw 'IW common identity'}
 . (Join-Path $script:HTRepo $common)
 if($Role -ceq 'Producer'){
  & (Join-Path $PSScriptRoot runtime-writer-phase-timing-capture.ps1) -Grant contract-green -FrozenSHA256 '186FE2C2B44A31A98EA5406109AC888A625FD1B0061C5338D0382268015D9BBA' -Epoch history-contract-runtime -InputSHA256 $InputSHA256 -IntentSHA256 $IntentSHA256
  exit $LASTEXITCODE
 }
 $row.Before=Map-IW writer.before-map.json
 $adapter=Join-Path $script:HTRepo $binding.AdapterPath;$dll=Join-Path $script:HTRepo $binding.AssemblyPath
 if((Get-FileHash $adapter).Hash -cne $binding.AdapterSHA256 -or (Get-FileHash $dll).Hash -cne $binding.AssemblySHA256){throw 'IW assembly identity'}
 $null=[Reflection.Assembly]::LoadFrom($adapter);$null=[Reflection.Assembly]::LoadFrom($dll)
 if($Role -ceq 'Writer'){
  $options=[Talenro.Diagnostics.IsolatedWriterV1.WriterOptions]::new();$options.Scenario=$Scenario;$options.Token=$intent.Token;$options.InputSHA256=$InputSHA256;$options.Directory=Join-Path $folder data
  . (Join-Path $PSScriptRoot runtime-writer-phase-timing.ps1)
  $script:WPTState=New-WPTState $intent.Token $InputSHA256 ([Diagnostics.Stopwatch]::Frequency)
  $script:WPTStream=[Console]::OpenStandardError()
  $script:WPTClock=[Diagnostics.Stopwatch]::StartNew()
  $task=[Talenro.Diagnostics.IsolatedWriterV1.Writer]::RunAsync($options)
 }
 while(-not $task.IsCompleted){[Threading.Thread]::Sleep(1)}
 $row.ExitCode=[int]$task.GetAwaiter().GetResult()
 Emit-WPT task.returned
}catch{$row.Error=[ordered]@{Message=$_.Exception.Message;Stack=$_.ScriptStackTrace};$row.ExitCode=1;Emit-WPT error}
finally{
 if($null -ne $row.Before){try{$row.After=Map-IW writer.after-map.json}catch{$row.Error=[ordered]@{Message=$_.Exception.Message;Stack=$_.ScriptStackTrace};$row.ExitCode=1;Emit-WPT error}}
}
if($null -ne $script:WPTState -and $script:WPTState.Failed){$row.ExitCode=1}
try{
 if($claimed){
  Emit-WPT role.save.begin
  if($null -ne $script:WPTState -and $script:WPTState.Failed){$row.ExitCode=1}
  Save-IW ($Role.ToLowerInvariant()+'.role-result.json') $row
  Emit-WPT role.save.end
 }
 Emit-WPT exit.ready
}catch{$row.ExitCode=1;Emit-WPT error}
if($null -ne $script:WPTState -and $script:WPTState.Failed){$row.ExitCode=1}
exit $row.ExitCode
