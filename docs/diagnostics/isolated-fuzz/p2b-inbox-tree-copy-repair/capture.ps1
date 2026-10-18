param([Parameter(Mandatory)][ValidateSet('Prepare','SealProduct','Run')][string]$Action,
[ValidateSet('red','full','capacity','sequence','final')][string]$Slot,[string]$FrozenSHA256,[string]$ProductSealSHA256)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$root=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'));Set-Location -LiteralPath $root
$packet=Join-Path $root '.superpowers/sdd/2026-10-18-p2b-inbox-tree-copy-repair'
$previous=Join-Path $root '.superpowers/sdd/2026-10-18-p2b-first-snapshot-diagnostic'
$manifest=Join-Path $packet 'frozen.json';$index=Join-Path $packet 'invocations.json';$seal=Join-Path $packet 'product-seal.json'
$subject=Join-Path $PSScriptRoot 'inbox.psm1'
$engine='C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe'
$engineHash='362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139'
$sdk=Join-Path ([IO.Path]::GetDirectoryName($engine)) 'System.Management.Automation.dll'
$sdkHash='2D8AF026E589C5603629BFC3E7C57265B4167B12ECA3AE549C6A6C3CA7E4967B'
function Assert-Hash($Path,$Hash){if($Hash -cnotmatch '^[A-F0-9]{64}$' -or (Get-FileHash -LiteralPath $Path).Hash -cne $Hash){throw ('Pin mismatch: '+$Path)}}
function Save-Json($Path,$Value){[IO.File]::WriteAllText($Path,($Value|ConvertTo-Json -Depth 30),[Text.UTF8Encoding]::new($false))}
$identity=[Security.Principal.WindowsIdentity]::GetCurrent()
if(([Security.Principal.WindowsPrincipal]::new($identity)).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)){throw 'Ordinary parent required'}
Assert-Hash $engine $engineHash;Assert-Hash $sdk $sdkHash
if($Action -ceq 'Prepare'){
    if(Test-Path -LiteralPath $packet){throw 'Fresh repair packet required'}
    Assert-Hash $subject '43067AD0939CB56426568DDF435695383E6C21ABFD223B5A369BC6AE6F2451C8'
    Assert-Hash (Join-Path $previous 'frozen.json') '4A398AE49E88206C4D0CD9D24C27EA42999DC9D64CC84A66E1D773085CFD391B'
    Assert-Hash (Join-Path $previous 'invocation.json') '6183737559F63F52D6A51755D643473DF3B17323A55C548543812CD8FC236492'
    $prior=Get-Content -LiteralPath (Join-Path $previous 'invocation.json') -Raw|ConvertFrom-Json
    if($prior.Outcome -cne 'DIAGNOSTIC_CAPTURED'){throw 'Prior captured diagnostic required'}
    foreach($stream in @('stdout','stderr')){Assert-Hash (Join-Path $previous ($stream+'.bin')) $prior.($stream+'SHA256')}
    $old=Get-Content -LiteralPath (Join-Path $previous 'frozen.json') -Raw|ConvertFrom-Json
    if($old.Pins.Count -ne 1289){throw 'Inherited pin inventory'}
    $pins=@($old.Pins);foreach($pin in $pins){Assert-Hash $pin.VerificationPath $pin.SHA256}
    $priorFiles=@(Get-ChildItem -LiteralPath $previous -File -Force);if($priorFiles.Count -ne 4){throw 'Diagnostic packet inventory'}
    $extra=@($priorFiles|ForEach-Object{$_.FullName})
    foreach($entry in @(@('p2b-first-snapshot-diagnostic',5),@('p2b-bounded-inbox-tick-correction',11))){
        $leaf=@(Get-ChildItem -LiteralPath (Join-Path $root ('docs/diagnostics/isolated-fuzz/'+$entry[0])) -File -Force)
        if($leaf.Count -ne $entry[1]){throw 'Prior portable inventory'};$extra+=@($leaf|ForEach-Object{$_.FullName})
    }
    foreach($path in @('docs/roadmap/2026-10-18-p2b-first-snapshot-diagnostic.md','docs/roadmap/2026-10-18-p2b-bounded-inbox-tick-correction.md','docs/superpowers/specs/2026-10-18-p2b-bounded-inbox-design.md')){$extra+=,(Join-Path $root $path)}
    $inputs=@(Get-ChildItem -LiteralPath $PSScriptRoot -File -Force|Where-Object{$_.Name -cne 'inbox.psm1'});if($inputs.Count -ne 10){throw 'Ten prepared inputs required'}
    foreach($file in $inputs){if([IO.File]::ReadAllText($file.FullName).Contains([string][char]13)){throw 'Physical LF required'};$extra+=,$file.FullName}
    $extra+=@($engine,$sdk)
    foreach($path in $extra){$pins+=,[pscustomobject]@{VerificationPath=$path;SHA256=(Get-FileHash -LiteralPath $path).Hash;Role='tree-repair-frozen-input'}}
    $plan=Join-Path $root 'docs/superpowers/plans/2026-10-18-p2b-inbox-tree-copy-repair.md'
    $snapshot=Join-Path $packet 'approved-plan.snapshot.md'
    $baseline=Join-Path $packet 'baseline-product.psm1'
    $pins+=,[pscustomobject]@{VerificationPath=$snapshot;SHA256=(Get-FileHash -LiteralPath $plan).Hash;Role='approved-plan-snapshot'}
    $pins+=,[pscustomobject]@{VerificationPath=$baseline;SHA256=(Get-FileHash -LiteralPath $subject).Hash;Role='baseline-product-snapshot'}
    # Display every mapping before creating the fresh packet; never alter inherited paths.
    $pins|ForEach-Object{[Console]::Out.WriteLine(($_|ConvertTo-Json -Compress))}
    $null=New-Item -ItemType Directory -Path $packet
    [IO.File]::WriteAllBytes($snapshot,[IO.File]::ReadAllBytes($plan))
    [IO.File]::WriteAllBytes($baseline,[IO.File]::ReadAllBytes($subject))
    Save-Json $manifest ([pscustomobject]@{Schema=1;BaseCommit='131fee8584a1b308d28152abff3151da3ab41554';OriginalHostsSpent=2;OriginalRemainingFrozen=3;DiagnosticHostsSpent=1;NewBudgetApproved=$true;Pins=$pins;PreviousPacketFiles=@($priorFiles.Name);NewExtraPaths=$extra;Order=@('red','full','capacity','sequence','final');MaxHosts=5;MaxReviewers=1;MaxMs=120000;ExitWaitMs=2000;DrainMs=2000;Context=('CONTEXT: '+$identity.Name+' admin=False');FormalAcceptance=$false})
    [pscustomobject]@{FrozenSHA256=(Get-FileHash -LiteralPath $manifest).Hash;Pins=$pins.Count;PreviousPacketFiles=$priorFiles.Count;PreparedInputs=$inputs.Count}|ConvertTo-Json;return
}
Assert-Hash $manifest $FrozenSHA256;$f=Get-Content -LiteralPath $manifest -Raw|ConvertFrom-Json
foreach($pin in $f.Pins){Assert-Hash $pin.VerificationPath $pin.SHA256}
$rows=@();if(Test-Path -LiteralPath $index){$rows=@(Get-Content -LiteralPath $index -Raw|ConvertFrom-Json)}
foreach($r in $rows){
    if($r.Outcome -cne 'EXPECTED_WITHIN_SCOPE'){throw 'Prior nonexact/incomplete freezes remainder'}
    foreach($stream in @('stdout','stderr')){Assert-Hash (Join-Path $packet ($r.Slot+'.'+$stream+'.bin')) $r.($stream+'SHA256')}
    Assert-Hash (Join-Path $packet ($r.Slot+'.trace.json')) $r.TraceSHA256
    if($null -ne $r.OriginalTraceSHA256){Assert-Hash (Join-Path $packet ($r.Slot+'.original.trace.json')) $r.OriginalTraceSHA256}
}
if($Action -ceq 'SealProduct'){
    if($rows.Count -ne 1 -or $rows[0].Slot -cne 'red' -or (Test-Path -LiteralPath $seal)){throw 'Seal once after exact RED'}
    if((Get-FileHash -LiteralPath $subject).Hash -ceq '43067AD0939CB56426568DDF435695383E6C21ABFD223B5A369BC6AE6F2451C8'){throw 'New repaired product required'}
    $good=[IO.File]::ReadAllText($subject)
    $count='if($endpoint.Count -ge 4)';$seq='if($message.Sequence -ne $endpoint.NextSequence)'
    foreach($needle in @($count,$seq)){if(($good.Split([string[]]@($needle),[StringSplitOptions]::None)).Count -ne 2){throw 'Unique mutant site required'}}
    [IO.File]::WriteAllText((Join-Path $packet 'good-product.psm1'),$good,[Text.UTF8Encoding]::new($false))
    [IO.File]::WriteAllText((Join-Path $packet 'capacity-product.psm1'),$good.Replace($count,'if($endpoint.Count -ge 5)'),[Text.UTF8Encoding]::new($false))
    [IO.File]::WriteAllText((Join-Path $packet 'sequence-product.psm1'),$good.Replace($seq,'if($message.Sequence -lt $endpoint.NextSequence)'),[Text.UTF8Encoding]::new($false))
    Save-Json $seal ([pscustomobject]@{FrozenSHA256=$FrozenSHA256;ModuleSHA256=(Get-FileHash -LiteralPath $subject).Hash;GoodSnapshotSHA256=(Get-FileHash -LiteralPath (Join-Path $packet 'good-product.psm1')).Hash;CapacitySHA256=(Get-FileHash -LiteralPath (Join-Path $packet 'capacity-product.psm1')).Hash;SequenceSHA256=(Get-FileHash -LiteralPath (Join-Path $packet 'sequence-product.psm1')).Hash;CapacityDiff='Count -ge 4 -> Count -ge 5';SequenceDiff='Sequence -ne Next -> -lt Next';RuntimeApproved=$false})
    (Get-FileHash -LiteralPath $seal).Hash;return
}
if([string]::IsNullOrEmpty($Slot) -or $rows.Count -ge 5 -or $Slot -cne $f.Order[$rows.Count]){throw 'Ordered new5; no retries'}
$hash=$null
if($Slot -ceq 'red'){$hash='43067AD0939CB56426568DDF435695383E6C21ABFD223B5A369BC6AE6F2451C8';Assert-Hash $subject $hash}
else{
    Assert-Hash $seal $ProductSealSHA256;$s=Get-Content -LiteralPath $seal -Raw|ConvertFrom-Json
    if($s.FrozenSHA256 -cne $FrozenSHA256){throw 'Product/frozen seal mismatch'}
    $hash=if($Slot -ceq 'capacity'){$s.CapacitySHA256}elseif($Slot -ceq 'sequence'){$s.SequenceSHA256}else{$s.ModuleSHA256}
    Assert-Hash $subject $hash
    Assert-Hash (Join-Path $packet 'good-product.psm1') $s.GoodSnapshotSHA256
    Assert-Hash (Join-Path $packet 'capacity-product.psm1') $s.CapacitySHA256
    Assert-Hash (Join-Path $packet 'sequence-product.psm1') $s.SequenceSHA256
}
$row=[pscustomobject][ordered]@{Number=($rows.Count+1);Slot=$Slot;AttemptConsumed=$true;Launched=$false;ModuleSHA256=$hash;PID=$null;ElapsedMs=$null;TimedOut=$false;ExitCode=$null;OwnedExit=$false;StreamsComplete=$false;PostPinsIntact=$false;ExactExpected=$false;Outcome='INCOMPLETE';stdoutSHA256=$null;stderrSHA256=$null;TraceSHA256=$null;OriginalTraceSHA256=$null;FormalAcceptance=$false}
$rows+=,$row;Save-Json $index $rows
$target=Join-Path $PSScriptRoot 'target.ps1';$trace=Join-Path $packet ($Slot+'.trace.json');$originalTrace=Join-Path $packet ($Slot+'.original.trace.json')
$p=[Diagnostics.Process]::new();$p.StartInfo=[Diagnostics.ProcessStartInfo]::new()
$p.StartInfo.FileName=$engine;$p.StartInfo.UseShellExecute=$false;$p.StartInfo.CreateNoWindow=$true;$p.StartInfo.RedirectStandardOutput=$true;$p.StartInfo.RedirectStandardError=$true
foreach($a in @('-NoProfile','-NonInteractive','-File',$target,'-Slot',$Slot,'-TracePath',$trace,'-OriginalTracePath',$originalTrace)){$p.StartInfo.ArgumentList.Add($a)}
$stdout=[IO.File]::Create((Join-Path $packet ($Slot+'.stdout.bin')));$stderr=[IO.File]::Create((Join-Path $packet ($Slot+'.stderr.bin')))
$timer=[Diagnostics.Stopwatch]::StartNew()
try{
    $null=$p.Start();$row.Launched=$true;$row.PID=$p.Id;Save-Json $index $rows
    $ot=$p.StandardOutput.BaseStream.CopyToAsync($stdout);$et=$p.StandardError.BaseStream.CopyToAsync($stderr)
    if(-not $p.WaitForExit($f.MaxMs)){$row.TimedOut=$true;$p.Kill()}
    $row.OwnedExit=$p.WaitForExit($f.ExitWaitMs);if($row.OwnedExit){$row.ExitCode=$p.ExitCode}
    $row.StreamsComplete=[Threading.Tasks.Task]::WaitAll([Threading.Tasks.Task[]]@($ot,$et),$f.DrainMs)
}finally{$timer.Stop();$row.ElapsedMs=$timer.ElapsedMilliseconds;$stdout.Dispose();$stderr.Dispose();$p.Dispose();Save-Json $index $rows}
Assert-Hash $manifest $FrozenSHA256;Assert-Hash $engine $engineHash;Assert-Hash $sdk $sdkHash
foreach($pin in $f.Pins){Assert-Hash $pin.VerificationPath $pin.SHA256}
Assert-Hash $subject $hash
$row.PostPinsIntact=$true
foreach($stream in @('stdout','stderr')){$row.($stream+'SHA256')=(Get-FileHash -LiteralPath (Join-Path $packet ($Slot+'.'+$stream+'.bin'))).Hash}
if(Test-Path -LiteralPath $trace){$row.TraceSHA256=(Get-FileHash -LiteralPath $trace).Hash}
if(Test-Path -LiteralPath $originalTrace){$row.OriginalTraceSHA256=(Get-FileHash -LiteralPath $originalTrace).Hash}
Save-Json $index $rows
$raw=[Text.UTF8Encoding]::new($false,$true).GetString([IO.File]::ReadAllBytes((Join-Path $packet ($Slot+'.stdout.bin'))))
$lf=[string][char]10;$crlf=[string][char]13+[char]10
$normal=$raw.Replace($crlf,$lf);if($normal.EndsWith($lf)){$normal=$normal.Substring(0,$normal.Length-1)}
$actual=@($normal.Split([char]10))
$oracle=Get-Content -LiteralPath (Join-Path $PSScriptRoot 'expected-output.json') -Raw|ConvertFrom-Json
$key=if($Slot -ceq 'final'){'full'}else{$Slot}
$expected=@($oracle.$key|ForEach-Object{if($_ -ceq '{Context}'){$f.Context}else{$_}})
$same=$actual.Count -eq $expected.Count;if($same){for($i=0;$i -lt $actual.Count;$i++){if($actual[$i] -cne $expected[$i]){$same=$false;break}}}
$t=$null;if(Test-Path -LiteralPath $trace){$t=Get-Content -LiteralPath $trace -Raw|ConvertFrom-Json}
$wantExit=if($Slot -cin @('red','capacity','sequence')){1}else{0}
$traceOK=$null -ne $t -and -not $t.RuntimeApproved
if($Slot -cin @('full','final')){
    $u=$null;if(Test-Path -LiteralPath $originalTrace){$u=Get-Content -LiteralPath $originalTrace -Raw|ConvertFrom-Json}
    $traceOK=$traceOK -and $t.Passed -and $null -eq $t.Error -and $t.Groups.Count -eq 12 -and $null -ne $u -and $u.Passed -and $null -eq $u.Error -and -not $u.RuntimeApproved -and $u.CompileCount -eq 1 -and $u.Observable -and $u.Counters -eq 0 -and $u.Groups.Count -eq 32 -and $u.Synthetic.Count -eq 4 -and $u.Decorations.Count -eq 8
}elseif($Slot -ceq 'red'){
    $v=$t.FirstActual;$rep=$t.Representation
    $traceOK=$traceOK -and -not $t.Passed -and $t.Groups.Count -eq 0 -and $null -ne $t.Error -and $t.Error.Message -ceq 'ASSERTION FAILED: P2BTreeFirstSnapshot' -and $null -ne $rep -and $rep.RawType -ceq 'System.Management.Automation.PSObject' -and $rep.BaseType -ceq 'System.Management.Automation.PSCustomObject' -and $null -ne $v -and -not $v.Accepted -and $v.Reason -ceq 'internal-error' -and $null -eq $v.Inbox -and $null -eq $v.Snapshot -and $null -eq $v.Frame -and $null -eq $v.EndpointId -and $v.AcceptedBytes -eq 0 -and -not $v.SourceVerified -and -not $v.LiveActionsPerformed -and -not $v.RuntimeApproved -and -not (Test-Path -LiteralPath $originalTrace)
}else{
    $assertion=if($Slot -ceq 'capacity'){'P2BInboxEndpointCapacity'}else{'P2BInboxSequenceExact'}
    $traceOK=$traceOK -and -not $t.Passed -and $null -ne $t.Error -and $t.Error.Message -ceq ('ASSERTION FAILED: '+$assertion) -and $t.CompileCount -eq 0 -and -not (Test-Path -LiteralPath $originalTrace)
}
$row.ExactExpected=$same -and $traceOK
if(-not $row.TimedOut -and $row.OwnedExit -and $row.StreamsComplete -and $row.PostPinsIntact -and $row.ExactExpected -and $row.ExitCode -eq $wantExit -and (Get-Item -LiteralPath (Join-Path $packet ($Slot+'.stderr.bin'))).Length -eq 0){$row.Outcome='EXPECTED_WITHIN_SCOPE'}
Save-Json $index $rows;$row|ConvertTo-Json;[Console]::Out.Write($raw)
if($row.Outcome -cne 'EXPECTED_WITHIN_SCOPE'){throw 'Unexpected result; freeze remainder/review, no repair/retry'}
