# One bounded OFFLINE batch only. No product/test import or candidate call.
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$root=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))
Set-Location -LiteralPath $root
$packet=Join-Path $root '.superpowers/sdd/2026-10-18-p1-inline-ownership-observer-corrected'
if(Test-Path -LiteralPath $packet){throw 'One offline batch only; no retry/overwrite'}
$principal=[Security.Principal.WindowsPrincipal]::new([Security.Principal.WindowsIdentity]::GetCurrent())
if($principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)){throw 'Ordinary token required'}
$engine='C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe'
$engineHash='362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139'
function Hash($Path){(Get-FileHash -LiteralPath $Path).Hash}
function NeedHash($Path,$Expected){if((Hash $Path) -cne $Expected){throw ('Pin drift: '+$Path)}}
function Save($Path,$Value){[IO.File]::WriteAllText($Path,($Value|ConvertTo-Json -Depth 15),[Text.UTF8Encoding]::new($false))}
$prior='.superpowers/sdd/2026-10-18-p1-timed-ownership-attribution'
NeedHash ($prior+'/frozen.json') 'E0622E5FE6994998E45E503E9DD51E05F52A5E8A3FA33A6FD4158DF3F6A55317'
$old=Get-Content ($prior+'/frozen.json') -Raw|ConvertFrom-Json
if($old.Pins.Count -ne 615){throw 'Inherited inventory'}
$pins=@($old.Pins)
$anchors=@{
    'frozen.json'='E0622E5FE6994998E45E503E9DD51E05F52A5E8A3FA33A6FD4158DF3F6A55317'
    'invocation.json'='461456CF7E66A6A8D2F4BB4AA6232844897A43A5AC0C15E3541BFA0EB2D457FA'
    'stdout.txt'='EC63374C80CCB78765CCB78268919C02032DE0FE3B0EE0EEEA00852008A3A34D'
    'analysis.json'='A1A950609234B95F66364C7BD738FCF435B08B2B89FFCB656E3943AFCB94141C'
    'stderr.txt'='E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855'
}
foreach($name in $anchors.Keys){NeedHash ($prior+'/'+$name) $anchors[$name];$pins+=,[pscustomobject]@{Path=($prior+'/'+$name);SHA256=$anchors[$name]}}
$subject=Join-Path $PSScriptRoot '../stage-timing-ownership'
NeedHash (Join-Path $subject 'inspect-capture.ps1') 'B3C6FE888A6303204B47FC23B9D22F790A207EFCA6956A71021EE738CC19C36C'
NeedHash (Join-Path $subject 'owned-return.tests.ps1') '986CDE7605B88500C35B70D5C4152B1D63370BAC3301B5BA514AC3B2C34AE5BA'
NeedHash (Join-Path $subject 'output-inventory.json') '7514F6C3348CF02EE59769F1EA895883E3A92F4C35373F7F7D963686156561D3'
foreach($dir in @($PSScriptRoot)){foreach($file in @(Get-ChildItem -LiteralPath $dir -File -Force)){$pins+=,[pscustomobject]@{Path=$file.FullName;SHA256=(Hash $file.FullName)}}}
foreach($pin in $pins){NeedHash $pin.Path $pin.SHA256};NeedHash $engine $engineHash
$null=New-Item -Path $packet -ItemType Directory
$manifest=Join-Path $packet 'frozen.json';$index=Join-Path $packet 'invocation.json'
Save $manifest ([pscustomobject]@{Schema=1;Pins=$pins;ExpectedCases=45;MaxOfflineHosts=1;MaxMs=30000;ExitMs=2000;DrainMs=2000;ProductCalls=0;CandidateImports=0;EngineSHA256=$engineHash})
$manifestHash=Hash $manifest
function CheckPins{NeedHash $manifest $manifestHash;NeedHash $engine $engineHash;foreach($pin in $pins){NeedHash $pin.Path $pin.SHA256}}
$r=[ordered]@{Schema=1;AttemptConsumed=$true;Kind='OFFLINE_ONLY';StartedUTC=[DateTime]::UtcNow.ToString('o');PID=$null;TimedOut=$false;ExitCode=$null;ElapsedMs=$null;OwnedExited=$false;StreamsComplete=$false;PinsIntact=$false;ControllerError=$null;Summary=$null;ProductCalls=0;CandidateImports=0;FormalAcceptance=$false}
Save $index $r
$psi=[Diagnostics.ProcessStartInfo]::new();$psi.FileName=$engine;$psi.UseShellExecute=$false;$psi.CreateNoWindow=$true;$psi.WorkingDirectory=$root
$psi.RedirectStandardOutput=$true;$psi.RedirectStandardError=$true
$psi.Arguments='-NoProfile -NonInteractive -File "'+(Join-Path $PSScriptRoot 'offline.tests.ps1')+'" -Packet "'+$packet+'"'
$p=[Diagnostics.Process]::new();$p.StartInfo=$psi
$o=[IO.File]::Open((Join-Path $packet 'stdout.txt'),[IO.FileMode]::CreateNew)
$e=[IO.File]::Open((Join-Path $packet 'stderr.txt'),[IO.FileMode]::CreateNew)
$ot=$null;$et=$null;$watch=[Diagnostics.Stopwatch]::StartNew()
try{
    if(-not $p.Start()){throw 'Launch failed'};$r.PID=$p.Id;Save $index $r
    $ot=$p.StandardOutput.BaseStream.CopyToAsync($o);$et=$p.StandardError.BaseStream.CopyToAsync($e)
    if(-not $p.WaitForExit([Math]::Max(0,30000-[int]$watch.ElapsedMilliseconds))){$r.TimedOut=$true}
}catch{$r.ControllerError=$_.Exception.Message}
finally{
    try{
        if($null -ne $r.PID){if(-not $p.HasExited){$p.Kill();if(-not $p.WaitForExit(2000)){throw 'Owned exit unknown'}};$r.OwnedExited=$true;$r.ExitCode=$p.ExitCode}
        if($null -eq $ot -or $null -eq $et -or -not [Threading.Tasks.Task]::WaitAll([Threading.Tasks.Task[]]@($ot,$et),2000)){throw 'Incomplete drain'}
        $null=$ot.GetAwaiter().GetResult();$null=$et.GetAwaiter().GetResult();$r.StreamsComplete=$true
    }catch{$r.ControllerError=([string]$r.ControllerError+' | '+$_.Exception.Message)}
    $watch.Stop();$r.ElapsedMs=$watch.ElapsedMilliseconds;$o.Dispose();$e.Dispose();$p.Dispose()
    try{CheckPins;$r.PinsIntact=$true}catch{$r.ControllerError=([string]$r.ControllerError+' | '+$_.Exception.Message)}
    $r.StdoutSHA256=Hash (Join-Path $packet 'stdout.txt');$r.StderrSHA256=Hash (Join-Path $packet 'stderr.txt')
    $r.StderrBytes=(Get-Item (Join-Path $packet 'stderr.txt')).Length
    $lines=@(Get-Content (Join-Path $packet 'stdout.txt'))
    $summary=@($lines|Where-Object{$_ -cmatch '^RESULT: OBSERVER_EQUIVALENCE cases=45 passed=[0-9]+ productCalls=0$'})
    if($summary.Count -eq 1){$r.Summary=$summary[0]}
    Save $index $r
}
[pscustomobject]$r|ConvertTo-Json -Compress
if($r.TimedOut -or -not $r.OwnedExited -or -not $r.StreamsComplete -or -not $r.PinsIntact -or $r.StderrBytes -ne 0 -or $null -ne $r.ControllerError -or $r.ExitCode -ne 0 -or $r.Summary -cne 'RESULT: OBSERVER_EQUIVALENCE cases=45 passed=45 productCalls=0'){throw 'Offline batch failed/incomplete; no repair or retry performed'}
