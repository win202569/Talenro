param([Parameter(Mandatory)][ValidateSet('Prepare','Run','Integrate')][string]$Action,
[ValidateSet('green','final')][string]$Slot,[string]$FrozenSHA256)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$root=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'));Set-Location -LiteralPath $root
$packet=Join-Path $root '.superpowers/sdd/2026-10-19-p2b-controller-causality-reverification'
$previous=Join-Path $root '.superpowers/sdd/2026-10-19-p2b-error-causality-characterization'
$previousLeaf=Join-Path $root 'docs/diagnostics/isolated-fuzz/p2b-error-causality-characterization'
$oldLeaf=Join-Path $root 'docs/diagnostics/isolated-fuzz/p2b-handshake-seam-verification'
$helperLeaf=Join-Path $root 'docs/diagnostics/isolated-fuzz/p2b-controller-exception-containment'
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
function Check-Json($ActualPath,$OraclePath,[string]$Property){
    $script:diffs=[Collections.Generic.List[string]]::new()
    $ad=[System.Text.Json.JsonDocument]::Parse([IO.File]::ReadAllText($ActualPath))
    $od=[System.Text.Json.JsonDocument]::Parse([IO.File]::ReadAllText($OraclePath))
    try{Compare-ExactJson $ad.RootElement $od.RootElement.GetProperty($Property) 'trace'}finally{$ad.Dispose();$od.Dispose()}
}
$identity=[Security.Principal.WindowsIdentity]::GetCurrent()
if(([Security.Principal.WindowsPrincipal]::new($identity)).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)){throw 'Ordinary parent required'}
Assert-Hash $engine '362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139'
Assert-Hash $sdk '2D8AF026E589C5603629BFC3E7C57265B4167B12ECA3AE549C6A6C3CA7E4967B'
$manifestPath=Join-Path $packet 'frozen.json';$index=Join-Path $packet 'hosts.json'
$oraclePath=Join-Path $PSScriptRoot 'expected-output.json'
if($Action -ceq 'Prepare'){
    if(Test-Path -LiteralPath $packet){
        $existing=@(Get-ChildItem -LiteralPath $packet -Force -Recurse)
        if($existing.Count -ne 1 -or $existing[0].Name -cne 'progress.md' -or $existing[0].PSIsContainer){throw 'Fresh packet with optional ledger only'}
    }
    Assert-Hash (Join-Path $previous 'frozen.json') '6B31CB6ABDD0C610CF955732207D3DC478B9052418269539181003481411428E'
    $old=Get-Content -LiteralPath (Join-Path $previous 'frozen.json') -Raw|ConvertFrom-Json
    if($old.Pins.Count -ne 1434){throw 'Inherited inventory'}
    $pins=@($old.Pins);foreach($pin in $pins){Assert-Hash $pin.VerificationPath $pin.SHA256}
    $previousFiles=@(Get-ChildItem -LiteralPath $previous -File -Recurse|Sort-Object FullName)
    $previousPortable=@(Get-ChildItem -LiteralPath $previousLeaf -File -Force|Sort-Object Name)
    if($previousFiles.Count -ne 7 -or $previousPortable.Count -ne 6){throw 'Historical counts'}
    $oldSummary=Get-Content -LiteralPath (Join-Path $previousLeaf 'portable-summary.json') -Raw|ConvertFrom-Json
    if($oldSummary.Status -cne 'CHARACTERIZED_WITHIN_SCOPE_NO_ACCEPTANCE' -or $oldSummary.ArtifactInventory.Count -ne 7){throw 'Diagnostic identity'}
    foreach($a in $oldSummary.ArtifactInventory){Assert-Hash (Join-Path $previous $a.RelativePath) $a.SHA256}
    $extra=@($previousFiles.FullName)+@($previousPortable.FullName)
    $extra+=,(Join-Path $root 'docs/roadmap/2026-10-19-p2b-error-causality-characterization.md')
    $inputs=@(Get-ChildItem -LiteralPath $PSScriptRoot -File -Force|Sort-Object Name)
    $names=@('.gitattributes','capture.ps1','containment.tests.ps1','expected-output.json','README.md')
    if($inputs.Count -ne 5 -or @($inputs|Where-Object{$_.Name -cnotin $names}).Count -ne 0){throw 'Five exact inputs'}
    foreach($file in $inputs){if([IO.File]::ReadAllText($file.FullName).Contains([string][char]13)){throw 'LF required'};$extra+=,$file.FullName}
    foreach($path in $extra){$pins+=,[pscustomobject]@{VerificationPath=$path;SHA256=(Get-FileHash -LiteralPath $path).Hash;Role='reverification-input'}}
    $plan=Join-Path $root 'docs/superpowers/plans/2026-10-19-p2b-controller-causality-reverification.md'
    Assert-Hash $plan '5264D0FBE0A26258E4DDC1F9A54E2C06786E1CFEAD4A141B8458CDDB77BD2FAC'
    $snapshot=Join-Path $packet 'approved-plan.snapshot.md'
    $pins+=,[pscustomobject]@{VerificationPath=$snapshot;SHA256=(Get-FileHash -LiteralPath $plan).Hash;Role='approved-plan-snapshot'}
    if($extra.Count -ne 19 -or $pins.Count -ne 1454){throw '1454 mapped inventory'}
    foreach($pin in $pins){[Console]::Out.WriteLine(($pin|ConvertTo-Json -Compress))}
    if(-not(Test-Path -LiteralPath $packet)){$null=New-Item -ItemType Directory -Path $packet}
    [IO.File]::WriteAllBytes($snapshot,[IO.File]::ReadAllBytes($plan))
    Save-Json $manifestPath ([pscustomobject][ordered]@{Schema=1;Mode='p2b-controller-causality-frozen';BaseCommit='62f7ad23058fc13dbea1387a41cbe6fab5da00b8';NewBudgetApproved=$true;Pins=$pins;InheritedPins=1434;AdditionalPins=20;MaxHosts=2;MaxChildren=4;MaxReviewers=1;MaxMs=120000;ExitWaitMs=2000;DrainMs=2000;Context=('CONTEXT: '+$identity.Name+' admin=False');RealGateImplemented=$false;RuntimeApproved=$false;FormalAcceptance=$false})
    [pscustomobject]@{FrozenSHA256=(Get-FileHash -LiteralPath $manifestPath).Hash;Pins=1454;Inputs=5}|ConvertTo-Json
    return
}
if($FrozenSHA256 -cnotmatch '^[A-F0-9]{64}$'){throw 'Explicit frozen hash required'}
Assert-Hash $manifestPath $FrozenSHA256
$f=Get-Content -LiteralPath $manifestPath -Raw|ConvertFrom-Json
if($f.Pins.Count -ne 1454 -or $f.MaxHosts -ne 2 -or $f.MaxChildren -ne 4 -or $f.MaxReviewers -ne 1 -or -not $f.NewBudgetApproved){throw 'Budget identity'}
foreach($pin in $f.Pins){Assert-Hash $pin.VerificationPath $pin.SHA256}
$hosts=@();if(Test-Path -LiteralPath $index){$hosts=@(Get-Content -LiteralPath $index -Raw|ConvertFrom-Json)}
$childPath=Join-Path $packet 'children.json'
if($Action -ceq 'Integrate'){
    if($hosts.Count -ne 2 -or $hosts[0].Outcome -cne 'EXPECTED_GREEN' -or $hosts[1].Slot -cne 'final' -or $hosts[1].Outcome -cne 'INCOMPLETE' -or -not $hosts[1].AttemptConsumed){throw 'Internal final guard'}
    $children=@(Get-Content -LiteralPath $childPath -Raw|ConvertFrom-Json)
    if($children.Count -ne 3 -or @($children|Where-Object{$_.Outcome -cne 'EXPECTED' -or -not $_.OwnedExit -or $_.Rescue}).Count -ne 0){throw 'Completed GREEN children required'}
    $row=[pscustomobject][ordered]@{Number=4;Case='final-seam';AttemptConsumed=$true;Started=$false;PID=$null;OwnedExit=$false;ExitCode=$null;Rescue=$false;Outcome='INCOMPLETE'}
    $children+=,$row;Save-Json $childPath $children
    . (Join-Path $helperLeaf 'owned-run.ps1')
    $p=[Diagnostics.Process]::new();$p.StartInfo=[Diagnostics.ProcessStartInfo]::new()
    $p.StartInfo.FileName=$engine;$p.StartInfo.UseShellExecute=$false;$p.StartInfo.CreateNoWindow=$true
    $p.StartInfo.RedirectStandardOutput=$true;$p.StartInfo.RedirectStandardError=$true
    $trace=Join-Path $packet 'final.seam.trace.json'
    foreach($a in @('-NoProfile','-NonInteractive','-File',(Join-Path $oldLeaf 'handshake-seam.tests.ps1'),'-TracePath',$trace)){$p.StartInfo.ArgumentList.Add($a)}
    $r=$null;$hasExited=$false
    try{
        $r=Invoke-P2BOwnedRun -Process $p -StdoutPath (Join-Path $packet 'final.seam.stdout.bin') -StderrPath (Join-Path $packet 'final.seam.stderr.bin') -ReceiptPath (Join-Path $packet 'final.seam.receipt.json') -MaxMs 110000
        $row.Started=$r.Started
        if($r.Started){$row.PID=$p.Id;$hasExited=$p.HasExited;$row.OwnedExit=$hasExited -and $r.OwnedExit;$row.ExitCode=$r.ExitCode}
    }finally{
        $associated=$false;try{$row.PID=$p.Id;$associated=$true;$row.Started=$true}catch{}
        try{if($associated){if(-not $p.HasExited){$row.Rescue=$true;$p.Kill()};$row.OwnedExit=$p.WaitForExit(2000);if($row.OwnedExit){$row.ExitCode=$p.ExitCode}}}
        finally{$p.Dispose();$children[3]=$row;Save-Json $childPath $children}
    }
    foreach($pin in $f.Pins){Assert-Hash $pin.VerificationPath $pin.SHA256}
    $stdoutHash=(Get-FileHash -LiteralPath (Join-Path $packet 'final.seam.stdout.bin')).Hash
    $stderrHash=(Get-FileHash -LiteralPath (Join-Path $packet 'final.seam.stderr.bin')).Hash
    $traceHash=(Get-FileHash -LiteralPath $trace).Hash
    $raw=[Text.UTF8Encoding]::new($false,$true).GetString([IO.File]::ReadAllBytes((Join-Path $packet 'final.seam.stdout.bin')))
    $seamOracle=Join-Path $oldLeaf 'expected-output.json'
    $oldOracle=Get-Content -LiteralPath $seamOracle -Raw|ConvertFrom-Json
    $lf=[string][char]10;$normal=$raw.Replace(([string][char]13+[char]10),$lf);if($normal.EndsWith($lf)){$normal=$normal.Substring(0,$normal.Length-1)}
    $actual=@($normal.Split([char]10));$expected=@($oldOracle.Output|ForEach-Object{if($_ -ceq '{Context}'){$f.Context}else{$_}})
    $same=$actual.Count -eq $expected.Count
    if($same){for($i=0;$i -lt $actual.Count;$i++){if($actual[$i] -cne $expected[$i]){$same=$false;break}}}
    Check-Json $trace $seamOracle 'TraceExpected'
    $ok=$r.Started -and $null -eq $r.PrimaryError -and $r.CleanupErrors.Count -eq 0 -and -not $r.CleanupOverflow -and -not $r.TimedOut -and -not $r.TerminationRequested -and $r.OwnedExit -and $hasExited -and $r.StreamsComplete -and $r.ExitCode -eq 0 -and -not $row.Rescue -and $same -and $script:diffs.Count -eq 0 -and $stdoutHash -ceq 'F693A64A1402A0AF882E30D62F6EB79601F5F380974900DD982B7ED94641A0FE' -and $traceHash -ceq '4751D03D28FB8342AFFC56BF0E8343F0E0B1D5E548A7DF7C1B164F719C10A17E' -and (Get-Item -LiteralPath (Join-Path $packet 'final.seam.stderr.bin')).Length -eq 0
    if($ok){$row.Outcome='FINAL_EXACT'};$children[3]=$row;Save-Json $childPath $children
    Save-Json (Join-Path $packet 'final.capture.json') ([pscustomobject][ordered]@{Exact=$ok;Started=$r.Started;OwnedExit=$r.OwnedExit;ActualHasExited=$hasExited;StreamsComplete=$r.StreamsComplete;TimedOut=$r.TimedOut;TerminationRequested=$r.TerminationRequested;PrimaryPresent=$null -ne $r.PrimaryError;CleanupCount=$r.CleanupErrors.Count;CleanupOverflow=$r.CleanupOverflow;ExitCode=$r.ExitCode;Rescue=$row.Rescue;PostPinsIntact=$true;StdoutSHA256=$stdoutHash;StderrSHA256=$stderrHash;TraceSHA256=$traceHash;TraceDifferences=$script:diffs.ToArray();CompileCount=0;ProductChanged=$false;SourceVerified=$false;LiveActionsPerformed=$false;RuntimeApproved=$false;FormalAcceptance=$false;RealGateImplemented=$false})
    [Console]::Out.Write($raw)
    if(-not $ok){throw 'Final nonexact; freeze review, no retry'}
    return
}
if($Action -cne 'Run' -or $Slot -cnotin @('green','final')){throw 'Run slot required'}
if(($Slot -ceq 'green' -and $hosts.Count -ne 0) -or ($Slot -ceq 'final' -and ($hosts.Count -ne 1 -or $hosts[0].Outcome -cne 'EXPECTED_GREEN'))){throw 'Ordered one-shot host budget'}
if($Slot -ceq 'green' -and (Test-Path -LiteralPath $childPath)){throw 'Existing children prohibit GREEN'}
if($Slot -ceq 'final'){
    $children=@(Get-Content -LiteralPath $childPath -Raw|ConvertFrom-Json)
    if($children.Count -ne 3 -or @($children|Where-Object{$_.Outcome -cne 'EXPECTED' -or -not $_.OwnedExit -or $_.Rescue}).Count -ne 0){throw 'GREEN children required'}
}
$row=[pscustomobject][ordered]@{Number=($hosts.Count+1);Slot=$Slot;AttemptConsumed=$true;Started=$false;PID=$null;TimedOut=$false;TerminationRequested=$false;OwnedExit=$false;StreamsComplete=$false;ExitCode=$null;ElapsedMs=$null;ParentError=$null;CleanupErrors=@();PostPinsIntact=$false;ExactExpected=$false;DetailsValid=$false;stdoutSHA256=$null;stderrSHA256=$null;TraceSHA256=$null;DetailsSHA256=$null;TraceDifferences=@();Outcome='INCOMPLETE';RealGateImplemented=$false;RuntimeApproved=$false;FormalAcceptance=$false}
$hosts+=,$row;Save-Json $index $hosts
$trace=Join-Path $packet 'green.trace.json';$detailsPath=Join-Path $packet 'green.details.json'
$p=[Diagnostics.Process]::new();$p.StartInfo=[Diagnostics.ProcessStartInfo]::new()
$p.StartInfo.FileName=$engine;$p.StartInfo.UseShellExecute=$false;$p.StartInfo.CreateNoWindow=$true
$p.StartInfo.RedirectStandardOutput=$true;$p.StartInfo.RedirectStandardError=$true
$argsList=if($Slot -ceq 'green'){@('-NoProfile','-NonInteractive','-File',(Join-Path $PSScriptRoot 'containment.tests.ps1'),'-TracePath',$trace,'-DetailsPath',$detailsPath)}else{@('-NoProfile','-NonInteractive','-File',$PSCommandPath,'-Action','Integrate','-FrozenSHA256',$FrozenSHA256)}
$p.StartInfo.Arguments=(@($argsList|ForEach-Object{'"'+$_+'"'}) -join ' ')
$stdout=$null;$stderr=$null;$ot=$null;$et=$null
$cleanup=[Collections.Generic.List[string]]::new()
function Add-Cleanup($Record){if($cleanup.Count -lt 8){$message=$Record.ToString();if($message.Length -gt 1024){$message=$message.Substring(0,1024)};$cleanup.Add($message)}}
$timer=[Diagnostics.Stopwatch]::StartNew()
try{
    $stdout=[IO.File]::Create((Join-Path $packet ($Slot+'.stdout.bin')))
    $stderr=[IO.File]::Create((Join-Path $packet ($Slot+'.stderr.bin')))
    if(-not $p.Start()){throw 'Host did not start'}
    $row.Started=$true;$row.PID=$p.Id
    $ot=$p.StandardOutput.BaseStream.CopyToAsync($stdout);$et=$p.StandardError.BaseStream.CopyToAsync($stderr)
    Save-Json $index $hosts
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
    Save-Json $index $hosts
}
Assert-Hash $manifestPath $FrozenSHA256
foreach($pin in $f.Pins){Assert-Hash $pin.VerificationPath $pin.SHA256}
$row.PostPinsIntact=$true
foreach($s in @('stdout','stderr')){$row.($s+'SHA256')=(Get-FileHash -LiteralPath (Join-Path $packet ($Slot+'.'+$s+'.bin'))).Hash}
$raw=[Text.UTF8Encoding]::new($false,$true).GetString([IO.File]::ReadAllBytes((Join-Path $packet ($Slot+'.stdout.bin'))))
$lf=[string][char]10;$normal=$raw.Replace(([string][char]13+[char]10),$lf);if($normal.EndsWith($lf)){$normal=$normal.Substring(0,$normal.Length-1)}
$oracle=Get-Content -LiteralPath $oraclePath -Raw|ConvertFrom-Json
$actual=@($normal.Split([char]10));$same=$false
if($Slot -ceq 'green'){
    $expected=@($oracle.GreenOutput)
    $same=$actual.Count -eq $expected.Count
    if($same){for($i=0;$i -lt $actual.Count;$i++){if($actual[$i] -cne $expected[$i]){$same=$false;break}}}
    if(Test-Path -LiteralPath $trace){$row.TraceSHA256=(Get-FileHash -LiteralPath $trace).Hash;Check-Json $trace $oraclePath 'TraceExpected';$row.TraceDifferences=$script:diffs.ToArray();$row.ExactExpected=$same -and $script:diffs.Count -eq 0}
    if(Test-Path -LiteralPath $detailsPath){
        $row.DetailsSHA256=(Get-FileHash -LiteralPath $detailsPath).Hash
        $d=Get-Content -LiteralPath $detailsPath -Raw|ConvertFrom-Json
        $valid=$d.Schema -eq 1 -and $d.Mode -ceq 'p2b-controller-causality-details' -and $d.ErrorCount -eq 2 -and $d.RecordsNewestFirst.Count -eq 2 -and -not $d.RecordsTruncated
        if($valid){foreach($record in $d.RecordsNewestFirst){
            if($record.RecordType -cne 'System.Management.Automation.ErrorRecord' -or $record.ChainTruncated -or $record.Chain.Count -lt 1 -or $record.Chain.Count -gt 8){$valid=$false;continue}
            foreach($node in $record.Chain){if($node.Message.Length -gt 1024 -or [string]::IsNullOrEmpty($node.Type)){$valid=$false}}
        }}
        $row.DetailsValid=$valid
    }
    $children=@(Get-Content -LiteralPath $childPath -Raw|ConvertFrom-Json)
    $childrenOK=$children.Count -eq 3 -and @($children|Where-Object{$_.Outcome -cne 'EXPECTED' -or -not $_.OwnedExit -or $_.Rescue}).Count -eq 0
}else{
    $trace=Join-Path $packet 'final.capture.json'
    if(Test-Path -LiteralPath $trace){$row.TraceSHA256=(Get-FileHash -LiteralPath $trace).Hash;Check-Json $trace $oraclePath 'FinalExpected';$row.TraceDifferences=$script:diffs.ToArray();$row.ExactExpected=$script:diffs.Count -eq 0 -and $row.stdoutSHA256 -ceq 'F693A64A1402A0AF882E30D62F6EB79601F5F380974900DD982B7ED94641A0FE'}
    $row.DetailsValid=$row.ExactExpected
    $children=@(Get-Content -LiteralPath $childPath -Raw|ConvertFrom-Json)
    $childrenOK=$children.Count -eq 4 -and $children[3].Outcome -ceq 'FINAL_EXACT' -and $children[3].OwnedExit -and -not $children[3].Rescue
}
if($row.ExactExpected -and $row.DetailsValid -and $childrenOK -and $row.Started -and -not $row.TimedOut -and -not $row.TerminationRequested -and $row.OwnedExit -and $row.StreamsComplete -and $row.ExitCode -eq 0 -and $null -eq $row.ParentError -and $row.CleanupErrors.Count -eq 0 -and (Get-Item -LiteralPath (Join-Path $packet ($Slot+'.stderr.bin'))).Length -eq 0){$row.Outcome=if($Slot -ceq 'green'){'EXPECTED_GREEN'}else{'EXPECTED_FINAL'}}
Save-Json $index $hosts
$row|ConvertTo-Json -Depth 12
[Console]::Out.Write($raw)
if($row.Outcome -cnotin @('EXPECTED_GREEN','EXPECTED_FINAL')){throw 'Nonexact verification; remaining budget frozen; no retry'}
