param([Parameter(Mandatory)][ValidateSet('Prepare','Run')][string]$Action,[string]$FrozenSHA256)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$root=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'));Set-Location -LiteralPath $root
$packet=Join-Path $root '.superpowers/sdd/2026-10-19-p2b-error-causality-characterization'
$previous=Join-Path $root '.superpowers/sdd/2026-10-19-p2b-controller-exception-containment'
$previousLeaf=Join-Path $root 'docs/diagnostics/isolated-fuzz/p2b-controller-exception-containment'
$engine='C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe'
$sdk=Join-Path ([IO.Path]::GetDirectoryName($engine)) 'System.Management.Automation.dll'
function Assert-Hash($Path,$Hash){if((Get-FileHash -LiteralPath $Path).Hash -cne $Hash){throw ('Pin mismatch: '+$Path)}}
function Save-Json($Path,$Value){[IO.File]::WriteAllText($Path,($Value|ConvertTo-Json -Depth 30),[Text.UTF8Encoding]::new($false))}
function Compare-ExactJson([System.Text.Json.JsonElement]$Left,[System.Text.Json.JsonElement]$Right,[string]$Path){
    if($Left.ValueKind -ne $Right.ValueKind){$script:diffs.Add($Path+':kind');return}
    switch($Left.ValueKind.ToString()){
        Object {
            $a=@($Left.EnumerateObject());$b=@($Right.EnumerateObject())
            if($a.Count -ne $b.Count){$script:diffs.Add($Path+':property-count');return}
            for($i=0;$i -lt $a.Count;$i++){
                if($a[$i].Name -cne $b[$i].Name){$script:diffs.Add($Path+':property-order');continue}
                Compare-ExactJson $a[$i].Value $b[$i].Value ($Path+'.'+$a[$i].Name)
            }
        }
        Array {
            $a=@($Left.EnumerateArray());$b=@($Right.EnumerateArray())
            if($a.Count -ne $b.Count){$script:diffs.Add($Path+':array-count');return}
            for($i=0;$i -lt $a.Count;$i++){Compare-ExactJson $a[$i] $b[$i] ($Path+'['+$i+']')}
        }
        String {if($Left.GetString() -cne $Right.GetString()){$script:diffs.Add($Path+':string')}}
        default {if($Left.GetRawText() -cne $Right.GetRawText()){$script:diffs.Add($Path+':token')}}
    }
}


$identity=[Security.Principal.WindowsIdentity]::GetCurrent()
if(([Security.Principal.WindowsPrincipal]::new($identity)).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)){throw 'Ordinary parent required'}
Assert-Hash $engine '362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139'
Assert-Hash $sdk '2D8AF026E589C5603629BFC3E7C57265B4167B12ECA3AE549C6A6C3CA7E4967B'
$manifestPath=Join-Path $packet 'frozen.json';$index=Join-Path $packet 'invocation.json'
if($Action -ceq 'Prepare'){
    if(Test-Path -LiteralPath $packet){throw 'Fresh characterization packet required'}
    Assert-Hash (Join-Path $previous 'green.frozen.json') '9AAF1DA823DC427F0989DBAEFD802FAFE70B08653177FE93289D038ADA377248'
    $old=Get-Content -LiteralPath (Join-Path $previous 'green.frozen.json') -Raw|ConvertFrom-Json
    if($old.Pins.Count -ne 1398){throw 'Inherited inventory'}
    $pins=@($old.Pins);foreach($pin in $pins){Assert-Hash $pin.VerificationPath $pin.SHA256}
    $previousFiles=@(Get-ChildItem -LiteralPath $previous -File -Recurse|Sort-Object FullName)
    $previousPortable=@(Get-ChildItem -LiteralPath $previousLeaf -File -Force|Sort-Object Name)
    if($previousFiles.Count -ne 19 -or $previousPortable.Count -ne 10){throw 'Historical counts'}
    $oldSummary=Get-Content -LiteralPath (Join-Path $previousLeaf 'portable-summary.json') -Raw|ConvertFrom-Json
    if($oldSummary.Status -cne 'GREEN_FAILED_FROZEN_UNFINISHED' -or $oldSummary.ArtifactInventory.Count -ne 19){throw 'Historical outcome identity'}
    foreach($a in $oldSummary.ArtifactInventory){Assert-Hash (Join-Path $previous $a.RelativePath) $a.SHA256}
    $extra=@($previousFiles.FullName)+@($previousPortable.FullName)
    $extra+=,(Join-Path $root 'docs/roadmap/2026-10-19-p2b-controller-exception-containment.md')
    $inputs=@(Get-ChildItem -LiteralPath $PSScriptRoot -File -Force|Sort-Object Name)
    $names=@('.gitattributes','capture.ps1','causality.tests.ps1','expected-output.json','README.md')
    if($inputs.Count -ne 5 -or @($inputs|Where-Object{$_.Name -cnotin $names}).Count -ne 0){throw 'Five exact inputs'}
    foreach($file in $inputs){
        if([IO.File]::ReadAllText($file.FullName).Contains([string][char]13)){throw 'New input LF required'}
        $extra+=,$file.FullName
    }
    foreach($path in $extra){$pins+=,[pscustomobject]@{VerificationPath=$path;SHA256=(Get-FileHash -LiteralPath $path).Hash;Role='causality-frozen-input'}}
    $plan=Join-Path $root 'docs/superpowers/plans/2026-10-19-p2b-error-causality-characterization.md'
    $snapshot=Join-Path $packet 'approved-plan.snapshot.md'
    $pins+=,[pscustomobject]@{VerificationPath=$snapshot;SHA256=(Get-FileHash -LiteralPath $plan).Hash;Role='approved-plan-snapshot'}
    if($extra.Count -ne 35 -or $pins.Count -ne 1434){throw '1434 mapped inventory'}
    foreach($pin in $pins){[Console]::Out.WriteLine(($pin|ConvertTo-Json -Compress))}
    $null=New-Item -ItemType Directory -Path $packet
    [IO.File]::WriteAllBytes($snapshot,[IO.File]::ReadAllBytes($plan))
    Save-Json $manifestPath ([pscustomobject][ordered]@{Schema=1;Mode='p2b-error-causality-frozen';BaseCommit='7e2184f67182a9759015bc7ccd92bf291e7bcec2';NewBudgetApproved=$true;Pins=$pins;InheritedPins=1398;AdditionalPins=36;MaxHosts=1;MaxChildren=0;MaxReviewers=0;MaxMs=120000;ExitWaitMs=2000;DrainMs=2000;Context=($identity.Name+' admin=False');RealGateImplemented=$false;RuntimeApproved=$false;FormalAcceptance=$false})
    [pscustomobject]@{FrozenSHA256=(Get-FileHash -LiteralPath $manifestPath).Hash;Pins=1434;Inputs=5}|ConvertTo-Json
    return
}
if($FrozenSHA256 -cnotmatch '^[A-F0-9]{64}$'){throw 'Explicit frozen hash required'}
Assert-Hash $manifestPath $FrozenSHA256
$f=Get-Content -LiteralPath $manifestPath -Raw|ConvertFrom-Json
if($f.Pins.Count -ne 1434 -or $f.MaxHosts -ne 1 -or $f.MaxChildren -ne 0 -or $f.MaxReviewers -ne 0 -or -not $f.NewBudgetApproved){throw 'Fixed budget identity'}
foreach($pin in $f.Pins){Assert-Hash $pin.VerificationPath $pin.SHA256}
if(Test-Path -LiteralPath $index){throw 'Host slot already consumed; no retry'}
$row=[pscustomobject][ordered]@{Number=1;AttemptConsumed=$true;Started=$false;PID=$null;TimedOut=$false;TerminationRequested=$false;OwnedExit=$false;StreamsComplete=$false;ExitCode=$null;ElapsedMs=$null;ParentError=$null;CleanupErrors=@();PostPinsIntact=$false;ExactExpected=$false;DetailsValid=$false;stdoutSHA256=$null;stderrSHA256=$null;TraceSHA256=$null;DetailsSHA256=$null;TraceDifferences=@();ChildrenAllowed=0;ReviewersAllowed=0;Outcome='INCOMPLETE';RealGateImplemented=$false;RuntimeApproved=$false;FormalAcceptance=$false}
Save-Json $index $row
$trace=Join-Path $packet 'diagnostic.trace.json';$detailsPath=Join-Path $packet 'diagnostic.details.json'
$p=[Diagnostics.Process]::new();$p.StartInfo=[Diagnostics.ProcessStartInfo]::new()
$p.StartInfo.FileName=$engine;$p.StartInfo.UseShellExecute=$false;$p.StartInfo.CreateNoWindow=$true
$p.StartInfo.RedirectStandardOutput=$true;$p.StartInfo.RedirectStandardError=$true
$argsList=@('-NoProfile','-NonInteractive','-File',(Join-Path $PSScriptRoot 'causality.tests.ps1'),'-TracePath',$trace,'-DetailsPath',$detailsPath)
$p.StartInfo.Arguments=(@($argsList|ForEach-Object{'"'+$_+'"'}) -join ' ')
$stdout=$null;$stderr=$null;$ot=$null;$et=$null
$cleanup=[Collections.Generic.List[string]]::new()
function Add-Cleanup($Record){if($cleanup.Count -lt 8){$message=$Record.ToString();if($message.Length -gt 1024){$message=$message.Substring(0,1024)};$cleanup.Add($message)}}
$timer=[Diagnostics.Stopwatch]::StartNew()
try{
    $stdout=[IO.File]::Create((Join-Path $packet 'diagnostic.stdout.bin'))
    $stderr=[IO.File]::Create((Join-Path $packet 'diagnostic.stderr.bin'))
    if(-not $p.Start()){throw 'Diagnostic did not start'}
    $row.Started=$true;$row.PID=$p.Id
    $ot=$p.StandardOutput.BaseStream.CopyToAsync($stdout);$et=$p.StandardError.BaseStream.CopyToAsync($stderr)
    Save-Json $index $row
    $remaining=[Math]::Max(0,($f.MaxMs-[int]$timer.ElapsedMilliseconds))
    if(-not $p.WaitForExit($remaining)){$row.TimedOut=$true}
}catch{$row.ParentError=$_.ToString()}
finally{
    if($row.Started){
        $exited=$false;try{$exited=$p.HasExited}catch{Add-Cleanup $_}
        if(-not $exited){$row.TerminationRequested=$true;try{$p.Kill()}catch{Add-Cleanup $_}}
        try{$row.OwnedExit=$p.WaitForExit($f.ExitWaitMs);if($row.OwnedExit){$row.ExitCode=$p.ExitCode}}catch{Add-Cleanup $_}
    }
    if($null -ne $ot -and $null -ne $et){try{$row.StreamsComplete=[Threading.Tasks.Task]::WaitAll([Threading.Tasks.Task[]]@($ot,$et),$f.DrainMs)}catch{Add-Cleanup $_}}
    if($null -ne $stdout){try{$stdout.Dispose()}catch{Add-Cleanup $_}}
    if($null -ne $stderr){try{$stderr.Dispose()}catch{Add-Cleanup $_}}
    try{$p.Dispose()}catch{Add-Cleanup $_}
    $timer.Stop();$row.ElapsedMs=$timer.ElapsedMilliseconds;$row.CleanupErrors=$cleanup.ToArray()
    Save-Json $index $row
}
Assert-Hash $manifestPath $FrozenSHA256
foreach($pin in $f.Pins){Assert-Hash $pin.VerificationPath $pin.SHA256}
$row.PostPinsIntact=$true
foreach($s in @('stdout','stderr')){$row.($s+'SHA256')=(Get-FileHash -LiteralPath (Join-Path $packet ('diagnostic.'+$s+'.bin'))).Hash}
if(Test-Path -LiteralPath $trace){$row.TraceSHA256=(Get-FileHash -LiteralPath $trace).Hash}
if(Test-Path -LiteralPath $detailsPath){$row.DetailsSHA256=(Get-FileHash -LiteralPath $detailsPath).Hash}
Save-Json $index $row
$raw=[Text.UTF8Encoding]::new($false,$true).GetString([IO.File]::ReadAllBytes((Join-Path $packet 'diagnostic.stdout.bin')))
$lf=[string][char]10;$normal=$raw.Replace(([string][char]13+[char]10),$lf);if($normal.EndsWith($lf)){$normal=$normal.Substring(0,$normal.Length-1)}
$oraclePath=Join-Path $PSScriptRoot 'expected-output.json';$oracle=Get-Content -LiteralPath $oraclePath -Raw|ConvertFrom-Json
$actual=@($normal.Split([char]10));$same=$actual.Count -eq $oracle.Output.Count
if($same){for($i=0;$i -lt $actual.Count;$i++){if($actual[$i] -cne $oracle.Output[$i]){$same=$false;break}}}
$script:diffs=[Collections.Generic.List[string]]::new();$traceOK=$false
if(Test-Path -LiteralPath $trace){
    $ad=[System.Text.Json.JsonDocument]::Parse([IO.File]::ReadAllText($trace));$od=[System.Text.Json.JsonDocument]::Parse([IO.File]::ReadAllText($oraclePath))
    try{Compare-ExactJson $ad.RootElement $od.RootElement.GetProperty('TraceExpected') 'trace';$traceOK=$script:diffs.Count -eq 0}finally{$ad.Dispose();$od.Dispose()}
}
$row.TraceDifferences=$script:diffs.ToArray();$row.ExactExpected=$same -and $traceOK
if(Test-Path -LiteralPath $detailsPath){
    $d=Get-Content -LiteralPath $detailsPath -Raw|ConvertFrom-Json
    $valid=$d.Schema -eq 1 -and $d.Mode -ceq 'p2b-error-causality-details' -and $d.Events.Count -eq 2
    if($valid){
        $stages=@('first-write','final-write')
        for($i=0;$i -lt 2;$i++){
            if($d.Events[$i].Stage -cne $stages[$i] -or $d.Events[$i].ErrorCount -lt 1){$valid=$false}
            foreach($kind in @('Catch','Anchor')){
                $record=$d.Events[$i].$kind
                if($record.RecordType -cne 'System.Management.Automation.ErrorRecord' -or $record.ChainTruncated -or $record.Chain.Count -lt 1 -or $record.Chain.Count -gt 8){$valid=$false;continue}
                foreach($node in $record.Chain){if($node.Message.Length -gt 1024 -or [string]::IsNullOrEmpty($node.Type)){$valid=$false}}
                $leafType=[Type]::GetType($record.Chain[-1].Type,$false)
                if($null -eq $leafType -or (-not [UnauthorizedAccessException].IsAssignableFrom($leafType) -and -not [IO.IOException].IsAssignableFrom($leafType))){$valid=$false}
            }
        }
    }
    $row.DetailsValid=$valid
}
if($row.ExactExpected -and $row.DetailsValid -and $row.Started -and -not $row.TimedOut -and -not $row.TerminationRequested -and $row.OwnedExit -and $row.StreamsComplete -and $row.ExitCode -eq 0 -and $null -eq $row.ParentError -and $row.CleanupErrors.Count -eq 0 -and (Get-Item -LiteralPath (Join-Path $packet 'diagnostic.stderr.bin')).Length -eq 0){$row.Outcome='EXPECTED_WITHIN_DIAGNOSTIC_SCOPE'}
Save-Json $index $row
$row|ConvertTo-Json -Depth 12
[Console]::Out.Write($raw)
if($row.Outcome -cne 'EXPECTED_WITHIN_DIAGNOSTIC_SCOPE'){throw 'Nonexact diagnostic; budget spent, stop without retry'}
