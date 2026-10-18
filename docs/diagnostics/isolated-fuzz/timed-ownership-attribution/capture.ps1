# One approved bounded candidate STAGE ATTRIBUTION host. No retry or later suites.
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$root=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))
Set-Location -LiteralPath $root
$packet=Join-Path $root '.superpowers/sdd/2026-10-18-p1-timed-ownership-attribution'
if(Test-Path -LiteralPath $packet){throw 'One attribution host only; no retry/overwrite'}
$principal=[Security.Principal.WindowsPrincipal]::new([Security.Principal.WindowsIdentity]::GetCurrent())
if($principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)){throw 'Ordinary token required'}
$engine='C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe'
$engineHash='362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139'
function Hash($Path){(Get-FileHash -LiteralPath $Path).Hash}
function NeedHash($Path,$Expected){if((Hash $Path) -cne $Expected){throw ('Pin drift: '+$Path)}}
function Save($Path,$Value){[IO.File]::WriteAllText($Path,($Value|ConvertTo-Json -Depth 15),[Text.UTF8Encoding]::new($false))}
$prior='.superpowers/sdd/2026-10-18-p1-strict-timing-json'
NeedHash ($prior+'/frozen.json') '90F5E19FE954E37C81AB60FB9B224F3DB0462A737FAAD3790A9F057A3858FDF7'
$old=Get-Content ($prior+'/frozen.json') -Raw|ConvertFrom-Json
if($old.Pins.Count -ne 607){throw 'Inherited inventory'}
$pins=@($old.Pins)
$anchors=@{
    'frozen.json'='90F5E19FE954E37C81AB60FB9B224F3DB0462A737FAAD3790A9F057A3858FDF7'
    'invocation.json'='D37361FE9BF3C73B2290848AB94AF456655810B0F13A0446D429EC68C5CF5A3C'
    'stdout.txt'='C6F3D25AF65AAC206F018FB0A67155B746D11A5916E486F39114245BCBF8830F'
    'stderr.txt'='E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855'
}
foreach($name in $anchors.Keys){NeedHash ($prior+'/'+$name) $anchors[$name];$pins+=,[pscustomobject]@{Path=($prior+'/'+$name);SHA256=$anchors[$name]}}
$subject=Join-Path $PSScriptRoot '../stage-timing-ownership'
$target=Join-Path $subject 'owned-return.tests.ps1'
$checker=Join-Path $PSScriptRoot '../strict-timing-json/inspect-capture.ps1'
$candidate=Join-Path $PSScriptRoot '../scalar-owned-return-candidate/offline-supervision-policy.ps1'
NeedHash $candidate '1971F808DAB0C0A520FC708BF240AF4367D21EE811FE9AEB8664E605E7E8BCE8'
NeedHash $checker '7E140A1BA753E1CBA836B36753FB3B741F615CB5941E9C1FD4990F9AF476B4D4'
NeedHash (Join-Path $subject 'owned-return.tests.ps1') '986CDE7605B88500C35B70D5C4152B1D63370BAC3301B5BA514AC3B2C34AE5BA'
NeedHash (Join-Path $subject 'output-inventory.json') '7514F6C3348CF02EE59769F1EA895883E3A92F4C35373F7F7D963686156561D3'
foreach($dir in @($PSScriptRoot)){foreach($file in @(Get-ChildItem -LiteralPath $dir -File -Force)){$pins+=,[pscustomobject]@{Path=$file.FullName;SHA256=(Hash $file.FullName)}}}

foreach($pin in $pins){NeedHash $pin.Path $pin.SHA256};NeedHash $engine $engineHash
$null=New-Item -Path $packet -ItemType Directory
$manifest=Join-Path $packet 'frozen.json';$index=Join-Path $packet 'invocation.json'
Save $manifest ([pscustomobject]@{Schema=1;Pins=$pins;ExpectedCases=39;TimingRecords=1340;BaseLines=153;MaxHosts=1;MaxMs=30000;ExitMs=2000;DrainMs=2000;DecisionBound=80;DirectScalarBound=29;DirectMetadataBound=1;Target=$target;Checker=$checker;Candidate=$candidate;EngineSHA256=$engineHash})
$manifestHash=Hash $manifest
function CheckPins{NeedHash $manifest $manifestHash;NeedHash $engine $engineHash;foreach($pin in $pins){NeedHash $pin.Path $pin.SHA256}}
$r=[ordered]@{Schema=1;AttemptConsumed=$true;Kind='CANDIDATE_STAGE_ATTRIBUTION';StartedUTC=[DateTime]::UtcNow.ToString('o');PID=$null;TimedOut=$false;ExitCode=$null;ElapsedMs=$null;OwnedExited=$false;StreamsComplete=$false;PinsIntact=$false;ControllerError=$null;Summary=$null;CheckerError=$null;TextOutcome=$null;TimingRecords=0;BaseLines=0;LastBoundary=$null;OpenAnnouncement=$null;Outcome='EXECUTION_INCOMPLETE';FormalAcceptance=$false}
Save $index $r
$psi=[Diagnostics.ProcessStartInfo]::new();$psi.FileName=$engine;$psi.UseShellExecute=$false;$psi.CreateNoWindow=$true;$psi.WorkingDirectory=$root
$psi.RedirectStandardOutput=$true;$psi.RedirectStandardError=$true
$psi.Arguments='-NoProfile -NonInteractive -File "'+$target+'"'
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
    $summary=@($lines|Where-Object{$_ -clike 'RESULT: *'});if($summary.Count -eq 1){$r.Summary=$summary[0]}
    if($r.StreamsComplete){
        try{
            # Saved text only; same ordinary controller, no second process/import.
            $analysis=& $checker -CapturePath (Join-Path $packet 'stdout.txt') -Mode Prefix|ConvertFrom-Json
            Save (Join-Path $packet 'analysis.json') $analysis
            $r.AnalysisSHA256=Hash (Join-Path $packet 'analysis.json')
            $r.TextOutcome=$analysis.Outcome;$r.TimingRecords=$analysis.TimingRecords;$r.BaseLines=$analysis.BaseLines
            $r.LastBoundary=$analysis.LastBoundary;$r.OpenAnnouncement=$analysis.OpenAnnouncement
        }catch{$r.CheckerError=$_.Exception.Message}
    }
    $valid=(-not $r.TimedOut -and $r.ExitCode -eq 0 -and $r.ElapsedMs -le 34000 -and $r.OwnedExited -and $r.StreamsComplete -and $r.PinsIntact -and $r.StderrBytes -eq 0 -and $null -eq $r.ControllerError -and $null -eq $r.CheckerError)
    if($valid -and $r.TextOutcome -ceq 'COMPLETE_TEXT_ONLY' -and $r.Summary -ceq 'RESULT: ISOLATED_WITHIN_MATRIX cases=39 calls=80 scalarCalls=29 metadataCalls=1'){$r.Outcome='PASSED_WITHIN_MATRIX'}
    Save $index $r
}
[pscustomobject]$r|ConvertTo-Json -Compress
if($r.Outcome -cne 'PASSED_WITHIN_MATRIX'){throw 'Attribution incomplete/failed; one host consumed; no retry or later suites'}
