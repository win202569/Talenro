param([Parameter(Mandatory)][ValidateSet('Prepare','SealProduct','Run')][string]$Action,
    [ValidateSet('red','focused','full')][string]$Slot,[string]$FrozenSHA256,[string]$ProductSealSHA256)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$root=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'));Set-Location -LiteralPath $root
$packet=Join-Path $root '.superpowers/sdd/2026-10-18-p2b-scalar-retained-repair'
$previous=Join-Path $root '.superpowers/sdd/2026-10-18-p2b-scalar-raw-continuation'
$manifest=Join-Path $packet 'frozen.json';$index=Join-Path $packet 'invocations.json';$seal=Join-Path $packet 'product-seal.json'
$subject=Join-Path $PSScriptRoot 'control-frame.ps1';$baselineCopy=Join-Path $packet 'baseline-product.ps1'
$baselineHash='B468B56426A084E472E97F651E23BE83C985014E3B2A8972785DD0D10A76D039'
$engine='C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe'
$engineHash='362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139'
$sdk=Join-Path ([IO.Path]::GetDirectoryName($engine)) 'System.Management.Automation.dll'
$sdkHash='2D8AF026E589C5603629BFC3E7C57265B4167B12ECA3AE549C6A6C3CA7E4967B'
function Assert-Hash($Path,$Hash){if($Hash -cnotmatch '^[A-F0-9]{64}$' -or (Get-FileHash -LiteralPath $Path).Hash -cne $Hash){throw ('Pin mismatch: '+$Path)}}
function Save-Json($Path,$Value){[IO.File]::WriteAllText($Path,($Value|ConvertTo-Json -Depth 24),[Text.UTF8Encoding]::new($false))}
$identity=[Security.Principal.WindowsIdentity]::GetCurrent()
if(([Security.Principal.WindowsPrincipal]::new($identity)).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)){throw 'Ordinary parent required'}
Assert-Hash $engine $engineHash;Assert-Hash $sdk $sdkHash
if($Action -ceq 'Prepare'){
    if(Test-Path -LiteralPath $packet){throw 'No replay'}
    $previous=Join-Path $root '.superpowers/sdd/2026-10-18-p2b-receiver-retention'
    Assert-Hash (Join-Path $previous 'frozen.json') '9049F39F4791ED297FD65EDFF8DE4995BBFEF330F690C71A4D19482E64EFF052'
    Assert-Hash (Join-Path $previous 'receipt.json') '52FCF3BD967087A06043561E9949710CBC7C5B7D7A0656FF3620BEF2BC2D5159'
    $old=Get-Content -LiteralPath (Join-Path $previous 'frozen.json') -Raw|ConvertFrom-Json
    if($old.Pins.Count -ne 1134){throw 'Inherited inventory'}
    foreach($pin in $old.Pins){Assert-Hash $pin.VerificationPath $pin.SHA256}
    $prior=Get-Content -LiteralPath (Join-Path $previous 'receipt.json') -Raw|ConvertFrom-Json
    if($prior.Consumed -ne 1 -or $prior.Outcome -cne 'RETENTION_VERIFIED'){throw 'Preserve prior test-only scope'}
    foreach($stream in @('stdout','stderr')){Assert-Hash (Join-Path $previous ($stream+'.bin')) $prior.($stream+'SHA256')}
    Assert-Hash $subject $baselineHash
    $pins=@($old.Pins)
    foreach($file in Get-ChildItem -LiteralPath $previous -File -Force){$pins+=,[pscustomobject]@{VerificationPath=$file.FullName;SHA256=(Get-FileHash -LiteralPath $file.FullName).Hash}}
    foreach($path in @((Join-Path $root 'docs/roadmap/2026-10-18-p2b-receiver-retention.md'),(Join-Path $root 'docs/diagnostics/isolated-fuzz/p2b-receiver-retention/portable-summary.json'))){
        $pins+=,[pscustomobject]@{VerificationPath=$path;SHA256=(Get-FileHash -LiteralPath $path).Hash}
    }
    $files=@(Get-ChildItem -LiteralPath $PSScriptRoot -File -Force|Where-Object{$_.Name -cne 'control-frame.ps1'})
    if($files.Count -ne 13){throw 'Thirteen immutable inputs required'}
    foreach($file in $files){
        if([IO.File]::ReadAllText($file.FullName).Contains("`r")){throw 'Physical LF required'}
        $pins+=,[pscustomobject]@{VerificationPath=$file.FullName;SHA256=(Get-FileHash -LiteralPath $file.FullName).Hash}
    }
    $pins+=,[pscustomobject]@{VerificationPath=$baselineCopy;SHA256=$baselineHash;Role='baseline-snapshot'}
    if($pins.Count -ne 1154){throw '1154 mapped pins required'}
    $oracle=Get-Content -LiteralPath (Join-Path $PSScriptRoot 'expected-output.json') -Raw|ConvertFrom-Json
    $expected=@($oracle.full|ForEach-Object{if($_ -ceq '{Context}'){'CONTEXT: '+$identity.Name+' admin=False'}else{$_}})
    if($expected.Count -ne 36){throw 'Full oracle inventory'}
    $null=New-Item -ItemType Directory -Path $packet
    [IO.File]::WriteAllText($baselineCopy,[IO.File]::ReadAllText($subject),[Text.UTF8Encoding]::new($false));Assert-Hash $baselineCopy $baselineHash
    Save-Json $manifest ([pscustomobject]@{Schema=1;BaseCommit='34bdafe520c8b42879aad964c66a593d66d65ba6';Pins=$pins;FullExpected=$expected;Order=@('red','focused','full');MaxHosts=3;MaxReviewers=1;MaxMs=60000;ExitWaitMs=2000;DrainMs=2000;FormalAcceptance=$false})
    (Get-FileHash -LiteralPath $manifest).Hash;return
}
Assert-Hash $manifest $FrozenSHA256;$f=Get-Content -LiteralPath $manifest -Raw|ConvertFrom-Json
foreach($pin in $f.Pins){Assert-Hash $pin.VerificationPath $pin.SHA256}
$rows=@();if(Test-Path -LiteralPath $index){$rows=@(Get-Content -LiteralPath $index -Raw|ConvertFrom-Json)}
foreach($row in $rows){
    if($row.Outcome -cne 'EXPECTED_WITHIN_SCOPE'){throw 'Unexpected prior result freezes remaining work'}
    foreach($stream in @('stdout','stderr')){Assert-Hash (Join-Path $packet ($row.Slot+'.'+$stream+'.bin')) $row.($stream+'SHA256')}
    if($null -ne $row.TraceSHA256){Assert-Hash (Join-Path $packet ($row.Slot+'.trace.json')) $row.TraceSHA256}
}
if($Action -ceq 'SealProduct'){
    if($rows.Count -ne 1 -or $rows[0].Slot -cne 'red' -or (Test-Path -LiteralPath $seal)){throw 'Seal once after exact RED'}
    $hash=(Get-FileHash -LiteralPath $subject).Hash;if($hash -ceq $baselineHash){throw 'Repair required'}
    Save-Json $seal ([pscustomobject]@{ModuleSHA256=$hash;FrozenSHA256=$FrozenSHA256;RuntimeApproved=$false})
    (Get-FileHash -LiteralPath $seal).Hash;return
}
if([string]::IsNullOrEmpty($Slot) -or $rows.Count -ge 3 -or $Slot -cne $f.Order[$rows.Count]){throw 'Ordered three-host allowance; no retries'}
$subjectHash=$baselineHash
if($Slot -cne 'red'){
    Assert-Hash $seal $ProductSealSHA256;$s=Get-Content -LiteralPath $seal -Raw|ConvertFrom-Json
    if($s.FrozenSHA256 -cne $FrozenSHA256){throw 'Seal binding'}
    $subjectHash=$s.ModuleSHA256
}
Assert-Hash $subject $subjectHash
$target=Join-Path $PSScriptRoot 'target.ps1'
$row=[pscustomobject][ordered]@{Number=($rows.Count+1);Slot=$Slot;AttemptConsumed=$true;Launched=$false;TargetSHA256=(Get-FileHash -LiteralPath $target).Hash;ModuleSHA256=$subjectHash;PID=$null;ElapsedMs=$null;TimedOut=$false;ExitCode=$null;OwnedExit=$false;StreamsComplete=$false;PostPinsIntact=$false;ExactExpected=$false;Outcome='INCOMPLETE';stdoutSHA256=$null;stderrSHA256=$null;TraceSHA256=$null;FormalAcceptance=$false}
$rows+=,$row;Save-Json $index $rows
$p=[Diagnostics.Process]::new();$p.StartInfo=[Diagnostics.ProcessStartInfo]::new()
$p.StartInfo.FileName=$engine;$p.StartInfo.UseShellExecute=$false;$p.StartInfo.CreateNoWindow=$true;$p.StartInfo.RedirectStandardOutput=$true;$p.StartInfo.RedirectStandardError=$true
foreach($arg in @('-NoProfile','-NonInteractive','-File',$target,'-Slot',$Slot)){$p.StartInfo.ArgumentList.Add($arg)}
if($Slot -ceq 'full'){foreach($arg in @('-TracePath',(Join-Path $packet 'full.trace.json'))){$p.StartInfo.ArgumentList.Add($arg)}}
$stdout=[IO.File]::Create((Join-Path $packet ($Slot+'.stdout.bin')));$stderr=[IO.File]::Create((Join-Path $packet ($Slot+'.stderr.bin')))
$timer=[Diagnostics.Stopwatch]::StartNew()
try{
    $null=$p.Start();$row.Launched=$true;$row.PID=$p.Id;Save-Json $index $rows
    $outTask=$p.StandardOutput.BaseStream.CopyToAsync($stdout);$errTask=$p.StandardError.BaseStream.CopyToAsync($stderr)
    if(-not $p.WaitForExit($f.MaxMs)){$row.TimedOut=$true;$p.Kill()}
    $row.OwnedExit=$p.WaitForExit($f.ExitWaitMs);if($row.OwnedExit){$row.ExitCode=$p.ExitCode}
    $row.StreamsComplete=[Threading.Tasks.Task]::WaitAll([Threading.Tasks.Task[]]@($outTask,$errTask),$f.DrainMs)
}finally{$timer.Stop();$row.ElapsedMs=$timer.ElapsedMilliseconds;$stdout.Dispose();$stderr.Dispose();$p.Dispose();Save-Json $index $rows}
Assert-Hash $manifest $FrozenSHA256;Assert-Hash $subject $subjectHash;Assert-Hash $engine $engineHash;Assert-Hash $sdk $sdkHash
foreach($pin in $f.Pins){Assert-Hash $pin.VerificationPath $pin.SHA256};$row.PostPinsIntact=$true
foreach($stream in @('stdout','stderr')){$row.($stream+'SHA256')=(Get-FileHash -LiteralPath (Join-Path $packet ($Slot+'.'+$stream+'.bin'))).Hash}
Save-Json $index $rows
$raw=[Text.UTF8Encoding]::new($false,$true).GetString([IO.File]::ReadAllBytes((Join-Path $packet ($Slot+'.stdout.bin'))))
$wantExit=if($Slot -ceq 'red'){1}else{0}
if($Slot -ceq 'full'){
    $normalized=$raw.Replace("`r`n","`n");if($normalized.EndsWith("`n")){$normalized=$normalized.Substring(0,$normalized.Length-1)}
    $lines=@($normalized.Split("`n"));$same=$lines.Count -eq $f.FullExpected.Count
    if($same){for($i=0;$i -lt $lines.Count;$i++){if($lines[$i] -cne $f.FullExpected[$i]){$same=$false;break}}}
    $trace=Join-Path $packet 'full.trace.json';$t=Get-Content -LiteralPath $trace -Raw|ConvertFrom-Json
    $row.TraceSHA256=(Get-FileHash -LiteralPath $trace).Hash
    $row.ExactExpected=$same -and $null -eq $t.Error -and $t.Counts.Total -eq 0 -and -not $t.RuntimeApproved
}else{
    $d=$raw|ConvertFrom-Json
    $common=$d.Schema -eq 1 -and $d.Mode -ceq 'p2b-scalar-retained-repair-tests' -and $d.HostVersion -ceq '7.6.5' -and $d.Core -and $d.X64 -and -not $d.Admin -and $d.Captured -and $d.FixtureObservable -and $d.CompilePerformed -and $d.ProductImported -and $d.ProductCalled -and -not $d.FullSuiteRun -and -not $d.RuntimeApproved -and $d.Rows.Count -eq 35
    foreach($case in $d.Rows){
        $common=$common -and $case.BridgeCaptureRetained -and $case.BridgeBaseMatches -and $case.Phase -ceq 'Completed' -and $case.BaseRetained -and $case.InputDescriptorRetained -and $case.CountersBefore.Total -eq 0 -and $case.CountersAfterAPI.Total -eq 0 -and $case.CountersAfterInventory.Total -eq 0 -and -not $case.Result.SourceVerified -and -not $case.Result.LiveActionsPerformed -and -not $case.Result.RuntimeApproved
        foreach($condition in $case.Preconditions.PSObject.Properties){$common=$common -and [bool]$condition.Value}
        foreach($condition in $case.BridgeVerdict.PSObject.Properties){$common=$common -and [bool]$condition.Value}
        if($case.Variant -cne 'clean'){$common=$common -and $case.CreatedAndRetainedSame -and $case.Before.Tag.SameDescriptor -and $case.After.Tag.SameDescriptor}
    }
    $common=$common -and $d.ProofsCompleted -and $d.Proofs.Count -eq 27 -and $d.HistoryCases.Count -eq 14
    foreach($proof in $d.Proofs){
        $common=$common -and $proof.Validated -and $proof.Retained -and $proof.Counters -eq 0
        foreach($condition in $proof.Verdict.PSObject.Properties){$common=$common -and [bool]$condition.Value}
    }
    foreach($history in $d.HistoryCases){$common=$common -and $history.Validated -and $history.Counters -eq 0 -and $history.Result.Accepted}
    if($Slot -ceq 'focused'){
        $common=$common -and $d.Guards.Count -eq 45 -and $d.KeyControls.Count -eq 54
        foreach($guard in $d.Guards){$common=$common -and $guard.Validated -and $guard.Counters -eq 0 -and $guard.Rejected -eq ($guard.Variant -cne 'clean')}
        foreach($control in $d.KeyControls){$common=$common -and $control.Validated -and $control.ProductChecked -and $control.KeySame -eq $control.ExpectedShared -and $control.TableSame -eq $control.ExpectedShared}
        $common=$common -and $d.Negatives.Count -eq 182
        foreach($negative in $d.Negatives){$common=$common -and $negative.Validated -and $negative.MutationProven -and $negative.Counters -eq 0 -and -not $negative.Result.Accepted -and $negative.Result.Reason -ceq 'invalid-input'}
    }else{$common=$common -and $d.Negatives.Count -eq 0 -and $d.Guards.Count -eq 0 -and $d.KeyControls.Count -eq 0}
    $hiddenAccepted=@($d.Rows|Where-Object{$_.Variant -clike 'hidden-*' -and $_.Result.Accepted}).Count
    if($Slot -ceq 'red'){
        $row.ExactExpected=$common -and -not $d.Passed -and $d.Stage -ceq 'HiddenDecorationRefusal' -and $null -ne $d.Error -and $d.Error.Chain[0].Message -ceq 'HiddenDecorationRefusal' -and $hiddenAccepted -eq 14
    }else{
        $row.ExactExpected=$common -and $d.Passed -and $d.Stage -ceq 'Passed' -and $null -eq $d.Error -and $hiddenAccepted -eq 0 -and @($d.Rows|Where-Object{$_.Variant -cne 'clean' -and ($_.Result.Accepted -or $_.Result.Reason -cne 'invalid-input')}).Count -eq 0
    }
}
if(-not $row.TimedOut -and $row.OwnedExit -and $row.StreamsComplete -and $row.PostPinsIntact -and $row.ExactExpected -and $row.ExitCode -eq $wantExit -and (Get-Item -LiteralPath (Join-Path $packet ($Slot+'.stderr.bin'))).Length -eq 0){$row.Outcome='EXPECTED_WITHIN_SCOPE'}
Save-Json $index $rows;$row|ConvertTo-Json
if($Slot -ceq 'full'){[Console]::Out.Write($raw)}
if($row.Outcome -cne 'EXPECTED_WITHIN_SCOPE'){throw 'Unexpected result; stop,no fix/retry/reviewer'}
