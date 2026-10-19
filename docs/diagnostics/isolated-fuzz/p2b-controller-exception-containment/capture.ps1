param([Parameter(Mandatory)][ValidateSet('Prepare','Run')][string]$Action,
[ValidateSet('final')][string]$Slot,[string]$FrozenSHA256)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$root=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'));Set-Location -LiteralPath $root
$packet=Join-Path $root '.superpowers/sdd/2026-10-19-p2b-controller-exception-containment'
$oldPacket=Join-Path $root '.superpowers/sdd/2026-10-19-p2b-handshake-seam-verification'
$oldLeaf=Join-Path $root 'docs/diagnostics/isolated-fuzz/p2b-handshake-seam-verification'
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
if(([Security.Principal.WindowsPrincipal]::new($identity)).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)){throw 'Ordinary controller required'}
Assert-Hash $engine '362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139'
Assert-Hash $sdk '2D8AF026E589C5603629BFC3E7C57265B4167B12ECA3AE549C6A6C3CA7E4967B'
if($Action -ceq 'Prepare'){
    if(Test-Path -LiteralPath $packet){throw 'Fresh containment packet required'}
    Assert-Hash (Join-Path $oldPacket 'frozen.json') 'D7476546E4F2491E31D9A4B5A76B2A1458DCF3172CA07EF6D874E5E6970ACD86'
    Assert-Hash (Join-Path $oldPacket 'invocations.json') 'F631E13950558B50F2B3A7FA3404FC2F462132B2B743D34796826456163DB21E'
    $old=Get-Content -LiteralPath (Join-Path $oldPacket 'frozen.json') -Raw|ConvertFrom-Json
    if($old.Pins.Count -ne 1378){throw 'Inherited count'}
    $pins=@($old.Pins);foreach($pin in $pins){Assert-Hash $pin.VerificationPath $pin.SHA256}
    $rows=@(Get-Content -LiteralPath (Join-Path $oldPacket 'invocations.json') -Raw|ConvertFrom-Json)
    if($rows.Count -ne 2 -or @($rows|Where-Object Outcome -cne 'EXPECTED_WITHIN_SCOPE').Count -ne 0){throw 'Old receipt identity'}
    $extra=@((Join-Path $oldPacket 'frozen.json'),(Join-Path $oldPacket 'invocations.json'))
    foreach($r in $rows){foreach($v in @(@('stdout.bin','stdoutSHA256'),@('stderr.bin','stderrSHA256'),@('trace.json','TraceSHA256'))){$path=Join-Path $oldPacket ($r.Slot+'.'+$v[0]);Assert-Hash $path $r.($v[1]);$extra+=,$path}}
    foreach($n in @('portable-summary.json','review.md','review-disposition.md')){$extra+=,(Join-Path $oldLeaf $n)}
    $extra+=,(Join-Path $root 'docs/roadmap/2026-10-19-p2b-handshake-seam-verification.md')
    $inputs=@(Get-ChildItem -LiteralPath $PSScriptRoot -File -Force|Sort-Object Name)
    $names=@('.gitattributes','capture.ps1','containment.tests.ps1','expected-output.json','fixture-child.ps1','owned-run.ps1','README.md')
    if($inputs.Count -ne 7 -or @($inputs|Where-Object{$_.Name -cnotin $names}).Count -ne 0){throw 'Seven inputs required'}
    foreach($inputFile in $inputs){if([IO.File]::ReadAllText($inputFile.FullName).Contains([string][char]13)){throw 'New inputs need LF'};$extra+=,$inputFile.FullName}
    foreach($path in $extra){$pins+=,[pscustomobject]@{VerificationPath=$path;SHA256=(Get-FileHash -LiteralPath $path).Hash;Role='containment-frozen-input'}}
    $plan=Join-Path $root 'docs/superpowers/plans/2026-10-19-p2b-controller-exception-containment.md'
    $snapshot=Join-Path $packet 'approved-plan.snapshot.md'
    $pins+=,[pscustomobject]@{VerificationPath=$snapshot;SHA256=(Get-FileHash -LiteralPath $plan).Hash;Role='approved-plan-snapshot'}
    if($extra.Count -ne 19 -or $pins.Count -ne 1398){throw '1398 inventory'}
    foreach($pin in $pins){[Console]::Out.WriteLine(($pin|ConvertTo-Json -Compress))}
    $null=New-Item -ItemType Directory -Path $packet
    [IO.File]::WriteAllBytes($snapshot,[IO.File]::ReadAllBytes($plan))
    $manifest=[pscustomobject][ordered]@{Schema=1;Phase='RED';BaseCommit='19e3880d5133ecdb75bed9981ceff25ceb0547ce';NewBudgetApproved=$true;Pins=$pins;MaxHosts=3;MaxChildren=5;MaxReviewers=1;Context=('CONTEXT: '+$identity.Name+' admin=False');RealGateImplemented=$false;RuntimeApproved=$false;FormalAcceptance=$false}
    Save-Json (Join-Path $packet 'red.frozen.json') $manifest
    [IO.File]::WriteAllBytes((Join-Path $packet 'red.owned-run.ps1'),[IO.File]::ReadAllBytes((Join-Path $PSScriptRoot 'owned-run.ps1')))
    [pscustomobject]@{FrozenSHA256=(Get-FileHash -LiteralPath (Join-Path $packet 'red.frozen.json')).Hash;Pins=$pins.Count;Inputs=7}|ConvertTo-Json
    return
}
if($Slot -cne 'final'){throw 'Only final integration slot'}
$manifestPath=Join-Path $packet 'green.frozen.json';Assert-Hash $manifestPath $FrozenSHA256
$f=Get-Content -LiteralPath $manifestPath -Raw|ConvertFrom-Json
if($f.Phase -cne 'GREEN' -or $f.Pins.Count -ne 1398){throw 'GREEN identity'}
foreach($pin in $f.Pins){Assert-Hash $pin.VerificationPath $pin.SHA256}
$hosts=@(Get-Content -LiteralPath (Join-Path $packet 'hosts.json') -Raw|ConvertFrom-Json)
if($hosts.Count -ne 3 -or $hosts[2].Slot -cne 'FINAL' -or $hosts[2].Outcome -cne 'INCOMPLETE' -or $hosts[0].Outcome -cne 'EXPECTED_RED_RESCUED' -or $hosts[1].Outcome -cne 'EXPECTED_GREEN'){throw 'Final host budget'}
$childPath=Join-Path $packet 'children.json';$children=@(Get-Content -LiteralPath $childPath -Raw|ConvertFrom-Json)
if($children.Count -ne 4 -or $children[0].Outcome -cne 'EXPECTED_RED_RESCUED' -or $children[1].Outcome -cne 'FAULT_CONTAINED' -or $children[2].Outcome -cne 'NORMAL_EXIT' -or $children[3].Outcome -cne 'TIMEOUT_CONTAINED'){throw 'Final child budget'}
$row=[pscustomobject][ordered]@{Number=5;Case='final-seam';AttemptConsumed=$true;Started=$false;PID=$null;OwnedExit=$false;Rescue=$false;Outcome='INCOMPLETE'}
$children+=,$row;Save-Json $childPath $children
. (Join-Path $PSScriptRoot 'owned-run.ps1')
$p=[Diagnostics.Process]::new();$p.StartInfo=[Diagnostics.ProcessStartInfo]::new()
$p.StartInfo.FileName=$engine;$p.StartInfo.UseShellExecute=$false;$p.StartInfo.CreateNoWindow=$true;$p.StartInfo.RedirectStandardOutput=$true;$p.StartInfo.RedirectStandardError=$true
$trace=Join-Path $packet 'final.seam.trace.json'
foreach($a in @('-NoProfile','-NonInteractive','-File',(Join-Path $oldLeaf 'handshake-seam.tests.ps1'),'-TracePath',$trace)){$p.StartInfo.ArgumentList.Add($a)}
$r=$null;$hasExited=$false
try{
    $r=Invoke-P2BOwnedRun -Process $p -StdoutPath (Join-Path $packet 'final.seam.stdout.bin') -StderrPath (Join-Path $packet 'final.seam.stderr.bin') -ReceiptPath (Join-Path $packet 'final.seam.receipt.json')
    $row.Started=$r.Started
    if($r.Started){$row.PID=$p.Id;$hasExited=$p.HasExited;$row.OwnedExit=$hasExited -and $r.OwnedExit}
}finally{
    if($null -ne $r -and $r.Started -and -not $p.HasExited){$row.Rescue=$true;$p.Kill();$row.OwnedExit=$p.WaitForExit(2000)}
    $p.Dispose();$children[4]=$row;Save-Json $childPath $children
}
foreach($pin in $f.Pins){Assert-Hash $pin.VerificationPath $pin.SHA256}
$stdoutHash=(Get-FileHash -LiteralPath (Join-Path $packet 'final.seam.stdout.bin')).Hash
$stderrHash=(Get-FileHash -LiteralPath (Join-Path $packet 'final.seam.stderr.bin')).Hash
$traceHash=(Get-FileHash -LiteralPath $trace).Hash
$raw=[Text.UTF8Encoding]::new($false,$true).GetString([IO.File]::ReadAllBytes((Join-Path $packet 'final.seam.stdout.bin')))
$oraclePath=Join-Path $oldLeaf 'expected-output.json';$oracle=Get-Content -LiteralPath $oraclePath -Raw|ConvertFrom-Json
$lf=[string][char]10;$normal=$raw.Replace(([string][char]13+[char]10),$lf);if($normal.EndsWith($lf)){$normal=$normal.Substring(0,$normal.Length-1)}
$actual=@($normal.Split([char]10));$expected=@($oracle.Output|ForEach-Object{if($_ -ceq '{Context}'){$f.Context}else{$_}})
$same=$actual.Count -eq $expected.Count;if($same){for($i=0;$i -lt $actual.Count;$i++){if($actual[$i] -cne $expected[$i]){$same=$false;break}}}
$script:diffs=[Collections.Generic.List[string]]::new()
$ad=[System.Text.Json.JsonDocument]::Parse([IO.File]::ReadAllText($trace));$od=[System.Text.Json.JsonDocument]::Parse([IO.File]::ReadAllText($oraclePath))
try{Compare-ExactJson $ad.RootElement $od.RootElement.GetProperty('TraceExpected') 'trace'}finally{$ad.Dispose();$od.Dispose()}
$ok=$r.Started -and $null -eq $r.PrimaryError -and $r.CleanupErrors.Count -eq 0 -and -not $r.TimedOut -and -not $r.TerminationRequested -and $r.OwnedExit -and $hasExited -and $r.StreamsComplete -and $r.ExitCode -eq 0 -and -not $row.Rescue -and $same -and $script:diffs.Count -eq 0 -and $stdoutHash -ceq 'F693A64A1402A0AF882E30D62F6EB79601F5F380974900DD982B7ED94641A0FE' -and $traceHash -ceq '4751D03D28FB8342AFFC56BF0E8343F0E0B1D5E548A7DF7C1B164F719C10A17E' -and (Get-Item -LiteralPath (Join-Path $packet 'final.seam.stderr.bin')).Length -eq 0
if($ok){$row.Outcome='FINAL_EXACT'};$children[4]=$row;Save-Json $childPath $children
Save-Json (Join-Path $packet 'final.capture.json') ([pscustomobject][ordered]@{Exact=$ok;Started=$r.Started;OwnedExit=$r.OwnedExit;ActualHasExited=$hasExited;StreamsComplete=$r.StreamsComplete;TerminationRequested=$r.TerminationRequested;ExitCode=$r.ExitCode;Rescue=$row.Rescue;PostPinsIntact=$true;StdoutSHA256=$stdoutHash;StderrSHA256=$stderrHash;TraceSHA256=$traceHash;TraceDifferences=$script:diffs.ToArray();CompileCount=0;ProductChanged=$false;RealGateImplemented=$false;RuntimeApproved=$false;FormalAcceptance=$false})
[Console]::Out.Write($raw)
if(-not $ok){throw 'Final nonexact: freeze review; no retry'}
