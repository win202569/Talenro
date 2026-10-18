param([Parameter(Mandatory)][ValidateSet('Prepare','SealProduct','Run')][string]$Action,
    [ValidateSet('focused','full')][string]$Slot,[string]$FrozenSHA256,[string]$ProductSealSHA256)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$root=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'));Set-Location -LiteralPath $root
$leaf=Join-Path $PSScriptRoot '../p2b-scalar-raw-repair'
$old=Join-Path $root '.superpowers/sdd/2026-10-18-p2b-scalar-raw-repair'
$packet=Join-Path $root '.superpowers/sdd/2026-10-18-p2b-scalar-raw-continuation'
$manifest=Join-Path $packet 'frozen.json';$index=Join-Path $packet 'invocations.json';$seal=Join-Path $packet 'product-seal.json'
$reconciled=Join-Path $packet 'red-reconciled.json'
$subject=Join-Path $leaf 'control-frame.ps1'
$baselineHash='B468B56426A084E472E97F651E23BE83C985014E3B2A8972785DD0D10A76D039'
$engine='C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe'
$engineHash='362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139'
$sdk=Join-Path ([IO.Path]::GetDirectoryName($engine)) 'System.Management.Automation.dll'
$sdkHash='2D8AF026E589C5603629BFC3E7C57265B4167B12ECA3AE549C6A6C3CA7E4967B'
function Need([bool]$Condition){if(-not $Condition){throw 'Continuation precondition'}}
function Assert-Hash($Path,$Hash){if($Hash -cnotmatch '^[A-F0-9]{64}$' -or (Get-FileHash -LiteralPath $Path).Hash -cne $Hash){throw ('Pin mismatch: '+$Path)}}
function Save-Json($Path,$Value){[IO.File]::WriteAllText($Path,($Value|ConvertTo-Json -Depth 24),[Text.UTF8Encoding]::new($false))}
function Assert-Matrix($d,[bool]$Baseline){
    Need ($d.Schema -eq 1 -and $d.Mode -ceq 'p2b-scalar-raw-repair-tests' -and $d.HostVersion -ceq '7.6.5' -and $d.Core -and $d.X64 -and -not $d.Admin)
    Need ($d.Captured -and $d.FixtureObservable -and $d.CompilePerformed -and $d.ProductImported -and $d.ProductCalled -and -not $d.FullSuiteRun -and -not $d.RuntimeApproved -and $d.Rows.Count -eq 35)
    $variants=@('clean','visible-note','hidden-note','visible-subclass','hidden-subclass')
    $modes=@('encode-scalar','encode-array','binding-message-scalar','binding-message-array','binding-expected-scalar','binding-expected-array','decode-array')
    for($n=0;$n -lt 35;$n++){
        $r=$d.Rows[$n];$v=$variants[[int][Math]::Floor($n/7)];$mode=$modes[$n%7]
        Need ($r.Number -eq $n+1 -and $r.Variant -ceq $v -and $r.Mode -ceq $mode -and $r.Phase -ceq 'Completed')
        Need ($r.BaseRetained -and $r.InputDescriptorRetained -and $r.CountersBefore.Total -eq 0 -and $r.CountersAfterAPI.Total -eq 0 -and $r.CountersAfterInventory.Total -eq 0)
        Need (-not $r.Result.SourceVerified -and -not $r.Result.LiveActionsPerformed -and -not $r.Result.RuntimeApproved)
        Need (@($r.Preconditions.PSObject.Properties).Count -eq 7)
        foreach($p in $r.Preconditions.PSObject.Properties){Need ([bool]$p.Value)}
        $scalar=$mode -clike '*-scalar'
        $wantType=if(-not $scalar){'System.Byte[]'}elseif($mode -ceq 'binding-expected-scalar'){'System.Int32'}else{'System.Int64'}
        Need ($r.Before.ValueCLRType -ceq $wantType -and $r.After.ValueCLRType -ceq $wantType)
        if($v -ceq 'clean'){
            Need ($r.Before.NonReservedCount -eq 0 -and $r.Before.TagCount -eq 0 -and $r.After.NonReservedCount -eq 0 -and $r.After.TagCount -eq 0)
        }else{
            $wantClass=if($v -clike '*subclass'){'P2BDescriptorFixture.CounterNote'}else{'System.Management.Automation.PSNoteProperty'}
            Need ($r.CreatedAndRetainedSame -and $r.ContractWouldRequireRefusal)
            foreach($snapshot in @($r.Before,$r.After)){
                Need ($snapshot.NonReservedCount -eq 1 -and $snapshot.TagCount -eq 1 -and $snapshot.Tag.SameDescriptor -and $snapshot.Tag.CLRType -ceq $wantClass -and -not $snapshot.Tag.Reserved -and $snapshot.Tag.Hidden -eq ($v -clike 'hidden-*'))
            }
        }
        $wantAccepted=$v -ceq 'clean' -or ($Baseline -and $v -clike 'hidden-*')
        $wantReason=if(-not $wantAccepted){'invalid-input'}elseif($mode -ceq 'decode-array'){'decoded'}elseif($mode -clike 'encode-*'){'encoded'}else{'binding-matched'}
        Need ($r.Result.Accepted -eq $wantAccepted -and $r.Result.Reason -ceq $wantReason)
        if(-not $wantAccepted){Need (-not $r.Result.BytesPresent -and -not $r.Result.MessagePresent -and $null -eq $r.Result.NextSequence)}
    }
    if($Baseline){
        Need (-not $d.Passed -and $d.Stage -ceq 'HiddenDecorationRefusal' -and $null -ne $d.Error -and $d.Error.FullyQualifiedErrorId -ceq 'HiddenDecorationRefusal' -and $d.Error.Chain.Count -eq 1 -and $d.Error.Chain[0].Message -ceq 'HiddenDecorationRefusal' -and $d.Error.Chain[0].Type -ceq 'System.Management.Automation.RuntimeException')
    }else{Need ($d.Passed -and $d.Stage -ceq 'Passed' -and $null -eq $d.Error)}
}
$identity=[Security.Principal.WindowsIdentity]::GetCurrent()
Need (-not ([Security.Principal.WindowsPrincipal]::new($identity)).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator))
Assert-Hash $engine $engineHash;Assert-Hash $sdk $sdkHash
if($Action -ceq 'Prepare'){
    Need (-not (Test-Path -LiteralPath $packet))
    Assert-Hash (Join-Path $old 'frozen.json') '6229E52C1AE3D07B2FDF4AE08C58C5381CF40350C7CF8D903695152A2EB3AEA0'
    Assert-Hash (Join-Path $old 'invocations.json') '7E7C5FA1D8A543D6262C734CEB465E2320DD320D87E05551423FA27E3B752751'
    $f=Get-Content -LiteralPath (Join-Path $old 'frozen.json') -Raw|ConvertFrom-Json
    Need ($f.Pins.Count -eq 1046 -and $f.MaxHosts -eq 3 -and $f.MaxReviewers -eq 1)
    foreach($pin in $f.Pins){Assert-Hash $pin.VerificationPath $pin.SHA256}
    $prior=@(Get-Content -LiteralPath (Join-Path $old 'invocations.json') -Raw|ConvertFrom-Json)
    Need ($prior.Count -eq 1);$r=$prior[0]
    Need ($r.Number -eq 1 -and $r.Slot -ceq 'red' -and $r.AttemptConsumed -and $r.Launched -and -not $r.TimedOut -and $r.ExitCode -eq 1 -and $r.OwnedExit -and $r.StreamsComplete -and $r.PostPinsIntact -and $r.Outcome -ceq 'INCOMPLETE' -and -not $r.FormalAcceptance)
    Assert-Hash (Join-Path $old 'red.stdout.bin') '78B6E38372B86EE436A1AF42331651AF3DBDD91734331D589A574CF57EDA7767'
    Assert-Hash (Join-Path $old 'red.stderr.bin') 'E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855'
    Need ($r.stdoutSHA256 -ceq '78B6E38372B86EE436A1AF42331651AF3DBDD91734331D589A574CF57EDA7767' -and $r.stderrSHA256 -ceq 'E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855')
    $bytes=[IO.File]::ReadAllBytes((Join-Path $old 'red.stdout.bin'))
    Need ($bytes.Length -eq 76414 -and $bytes[76226] -eq 161 -and $bytes[76227] -eq 173)
    $encoding=[Text.Encoding]::GetEncoding(936,[Text.EncoderFallback]::ExceptionFallback,[Text.DecoderFallback]::ExceptionFallback)
    $text=$encoding.GetString($bytes);Need ([Convert]::ToHexString($encoding.GetBytes($text)) -ceq [Convert]::ToHexString($bytes))
    $d=$text|ConvertFrom-Json;Assert-Matrix $d $true
    Assert-Hash $subject $baselineHash;Need (-not (Test-Path -LiteralPath (Join-Path $old 'product-seal.json')))
    $pins=@($f.Pins)
    foreach($file in Get-ChildItem -LiteralPath $old -File -Force){$pins+=,[pscustomobject]@{VerificationPath=$file.FullName;SHA256=(Get-FileHash -LiteralPath $file.FullName).Hash}}
    $inputs=@(Get-ChildItem -LiteralPath $PSScriptRoot -File -Force);Need ($inputs.Count -eq 4)
    foreach($file in $inputs){$pins+=,[pscustomobject]@{VerificationPath=$file.FullName;SHA256=(Get-FileHash -LiteralPath $file.FullName).Hash}}
    $report=Join-Path $root 'docs/roadmap/2026-10-18-p2b-scalar-raw-repair.md'
    $pins+=,[pscustomobject]@{VerificationPath=$report;SHA256=(Get-FileHash -LiteralPath $report).Hash}
    Need ($pins.Count -eq 1056)
    $null=New-Item -ItemType Directory -Path $packet
    Save-Json $reconciled ([pscustomobject]@{Schema=1;Outcome='RED_EXPECTED_RECONCILED';OldReceiptOutcome='INCOMPLETE';OldReceiptUnchanged=$true;Decoder='strict CP936';Rows=35;ExpectedHiddenAccepted=14;PreconditionsAndIdentityVerified=$true;Counters=0;OriginalConsumed=1;OriginalRemaining=2;NewTargetLaunched=$false;RuntimeApproved=$false})
    Save-Json $manifest ([pscustomobject]@{Schema=1;BaseCommit='b5f4fca07a2dad59f93651ff55695117aa651d89';Pins=$pins;ReconciledSHA256=(Get-FileHash -LiteralPath $reconciled).Hash;FullExpected=$f.FullExpected;Order=@('focused','full');OriginalBudget=3;AlreadyConsumed=1;MaxHosts=2;MaxReviewers=1;MaxMs=60000;ExitWaitMs=2000;DrainMs=2000;FormalAcceptance=$false})
    (Get-FileHash -LiteralPath $manifest).Hash;return
}
Assert-Hash $manifest $FrozenSHA256;$f=Get-Content -LiteralPath $manifest -Raw|ConvertFrom-Json
Assert-Hash $reconciled $f.ReconciledSHA256
foreach($pin in $f.Pins){Assert-Hash $pin.VerificationPath $pin.SHA256}
$rows=@();if(Test-Path -LiteralPath $index){$rows=@(Get-Content -LiteralPath $index -Raw|ConvertFrom-Json)}
foreach($row in $rows){
    Need ($row.Outcome -ceq 'EXPECTED_WITHIN_SCOPE')
    foreach($stream in @('stdout','stderr')){Assert-Hash (Join-Path $packet ($row.Slot+'.'+$stream+'.bin')) $row.($stream+'SHA256')}
    if($null -ne $row.TraceSHA256){Assert-Hash (Join-Path $packet ($row.Slot+'.trace.json')) $row.TraceSHA256}
}
if($Action -ceq 'SealProduct'){
    Need ($rows.Count -eq 0 -and -not (Test-Path -LiteralPath $seal))
    $hash=(Get-FileHash -LiteralPath $subject).Hash;Need ($hash -cne $baselineHash)
    Save-Json $seal ([pscustomobject]@{ModuleSHA256=$hash;FrozenSHA256=$FrozenSHA256;RuntimeApproved=$false})
    (Get-FileHash -LiteralPath $seal).Hash;return
}
Need (-not [string]::IsNullOrEmpty($Slot) -and $rows.Count -lt 2 -and $Slot -ceq $f.Order[$rows.Count])
Assert-Hash $seal $ProductSealSHA256;$s=Get-Content -LiteralPath $seal -Raw|ConvertFrom-Json
Need ($s.FrozenSHA256 -ceq $FrozenSHA256);Assert-Hash $subject $s.ModuleSHA256
$target=Join-Path $PSScriptRoot 'target.ps1'
$row=[pscustomobject][ordered]@{Number=($rows.Count+2);Slot=$Slot;AttemptConsumed=$true;Launched=$false;ModuleSHA256=$s.ModuleSHA256;TargetSHA256=(Get-FileHash -LiteralPath $target).Hash;PID=$null;ElapsedMs=$null;TimedOut=$false;ExitCode=$null;OwnedExit=$false;StreamsComplete=$false;PostPinsIntact=$false;ExactExpected=$false;Outcome='INCOMPLETE';stdoutSHA256=$null;stderrSHA256=$null;TraceSHA256=$null;FormalAcceptance=$false}
$rows+=,$row;Save-Json $index $rows
$p=[Diagnostics.Process]::new();$p.StartInfo=[Diagnostics.ProcessStartInfo]::new()
$p.StartInfo.FileName=$engine;$p.StartInfo.UseShellExecute=$false;$p.StartInfo.CreateNoWindow=$true;$p.StartInfo.RedirectStandardOutput=$true;$p.StartInfo.RedirectStandardError=$true
foreach($arg in @('-NoProfile','-NonInteractive','-File',$target,'-Slot',$Slot,'-TracePath',(Join-Path $packet ($Slot+'.trace.json')))){$p.StartInfo.ArgumentList.Add($arg)}
$stdout=[IO.File]::Create((Join-Path $packet ($Slot+'.stdout.bin')));$stderr=[IO.File]::Create((Join-Path $packet ($Slot+'.stderr.bin')))
$timer=[Diagnostics.Stopwatch]::StartNew()
try{
    $null=$p.Start();$row.Launched=$true;$row.PID=$p.Id;Save-Json $index $rows
    $outTask=$p.StandardOutput.BaseStream.CopyToAsync($stdout);$errTask=$p.StandardError.BaseStream.CopyToAsync($stderr)
    if(-not $p.WaitForExit($f.MaxMs)){$row.TimedOut=$true;$p.Kill()}
    $row.OwnedExit=$p.WaitForExit($f.ExitWaitMs);if($row.OwnedExit){$row.ExitCode=$p.ExitCode}
    $row.StreamsComplete=[Threading.Tasks.Task]::WaitAll([Threading.Tasks.Task[]]@($outTask,$errTask),$f.DrainMs)
}finally{$timer.Stop();$row.ElapsedMs=$timer.ElapsedMilliseconds;$stdout.Dispose();$stderr.Dispose();$p.Dispose();Save-Json $index $rows}
Assert-Hash $manifest $FrozenSHA256;Assert-Hash $subject $s.ModuleSHA256;Assert-Hash $engine $engineHash;Assert-Hash $sdk $sdkHash
foreach($pin in $f.Pins){Assert-Hash $pin.VerificationPath $pin.SHA256};$row.PostPinsIntact=$true
foreach($stream in @('stdout','stderr')){$row.($stream+'SHA256')=(Get-FileHash -LiteralPath (Join-Path $packet ($Slot+'.'+$stream+'.bin'))).Hash}
Save-Json $index $rows
$raw=[Text.UTF8Encoding]::new($false,$true).GetString([IO.File]::ReadAllBytes((Join-Path $packet ($Slot+'.stdout.bin'))))
if($Slot -ceq 'focused'){
    $d=$raw|ConvertFrom-Json;Assert-Matrix $d $false;$row.ExactExpected=$true
}else{
    $normalized=$raw.Replace("`r`n","`n");if($normalized.EndsWith("`n")){$normalized=$normalized.Substring(0,$normalized.Length-1)}
    $lines=@($normalized.Split("`n"));$same=$lines.Count -eq $f.FullExpected.Count
    if($same){for($n=0;$n -lt $lines.Count;$n++){if($lines[$n] -cne $f.FullExpected[$n]){$same=$false;break}}}
    $trace=Join-Path $packet 'full.trace.json';$t=Get-Content -LiteralPath $trace -Raw|ConvertFrom-Json
    $row.TraceSHA256=(Get-FileHash -LiteralPath $trace).Hash
    $row.ExactExpected=$same -and $null -eq $t.Error -and $t.Counts.Total -eq 0 -and -not $t.RuntimeApproved
}
if(-not $row.TimedOut -and $row.OwnedExit -and $row.StreamsComplete -and $row.PostPinsIntact -and $row.ExactExpected -and $row.ExitCode -eq 0 -and (Get-Item -LiteralPath (Join-Path $packet ($Slot+'.stderr.bin'))).Length -eq 0){$row.Outcome='EXPECTED_WITHIN_SCOPE'}
Save-Json $index $rows;$row|ConvertTo-Json
if($Slot -ceq 'full'){[Console]::Out.Write($raw)}
Need ($row.Outcome -ceq 'EXPECTED_WITHIN_SCOPE')
