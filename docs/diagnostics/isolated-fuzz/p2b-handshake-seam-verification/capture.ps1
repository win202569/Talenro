param([Parameter(Mandatory)][ValidateSet('Prepare','Run')][string]$Action,
[ValidateSet('full','final')][string]$Slot,[string]$FrozenSHA256)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$root=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'));Set-Location -LiteralPath $root
$packet=Join-Path $root '.superpowers/sdd/2026-10-19-p2b-handshake-seam-verification'
$previous=Join-Path $root '.superpowers/sdd/2026-10-18-p2b-inbox-tree-copy-repair'
$legacy=Join-Path $root 'docs/diagnostics/isolated-fuzz/p2b-inbox-tree-copy-repair'
$manifest=Join-Path $packet 'frozen.json';$index=Join-Path $packet 'invocations.json'
$engine='C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe'
$engineHash='362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139'
$sdk=Join-Path ([IO.Path]::GetDirectoryName($engine)) 'System.Management.Automation.dll'
$sdkHash='2D8AF026E589C5603629BFC3E7C57265B4167B12ECA3AE549C6A6C3CA7E4967B'
function Assert-Hash($Path,$Hash){
    if($Hash -cnotmatch '^[A-F0-9]{64}$' -or (Get-FileHash -LiteralPath $Path).Hash -cne $Hash){throw ('Pin mismatch: '+$Path)}
}
function Save-Json($Path,$Value){[IO.File]::WriteAllText($Path,($Value|ConvertTo-Json -Depth 35),[Text.UTF8Encoding]::new($false))}
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
Assert-Hash $engine $engineHash;Assert-Hash $sdk $sdkHash
if($Action -ceq 'Prepare'){
    if(Test-Path -LiteralPath $packet){throw 'Fresh seam packet required'}
    Assert-Hash (Join-Path $previous 'frozen.json') 'DB3FFAFC5D8A9E9501B3C12FFBFBF3735CB5B52D8414CCB020EC71ED52E58418'
    Assert-Hash (Join-Path $previous 'product-seal.json') 'DB47535AF5D567C23D94DAC03E491E9E6D3A9C3BC4C92708AC0B44416AA305C8'
    Assert-Hash (Join-Path $previous 'invocations.json') '447B896F231E8B7DD714887D919B12700A0AA0F214D3F55F8F45EB3E1DE693F6'
    Assert-Hash (Join-Path $legacy 'inbox.psm1') '47BABFEDFDA73C850EA49CC74E3B6348919EAC69E7AFBFD3DCA907B4988C02FC'
    $priorRows=@(Get-Content -LiteralPath (Join-Path $previous 'invocations.json') -Raw|ConvertFrom-Json)
    if($priorRows.Count -ne 5 -or @($priorRows|Where-Object Outcome -cne 'EXPECTED_WITHIN_SCOPE').Count -ne 0){throw 'Five prior exact rows required'}
    foreach($r in $priorRows){
        foreach($stream in @('stdout','stderr')){Assert-Hash (Join-Path $previous ($r.Slot+'.'+$stream+'.bin')) $r.($stream+'SHA256')}
        Assert-Hash (Join-Path $previous ($r.Slot+'.trace.json')) $r.TraceSHA256
        if($null -ne $r.OriginalTraceSHA256){Assert-Hash (Join-Path $previous ($r.Slot+'.original.trace.json')) $r.OriginalTraceSHA256}
    }
    $old=Get-Content -LiteralPath (Join-Path $previous 'frozen.json') -Raw|ConvertFrom-Json
    if($old.Pins.Count -ne 1326){throw 'Inherited inventory'}
    $pins=@($old.Pins);foreach($pin in $pins){Assert-Hash $pin.VerificationPath $pin.SHA256}
    $priorFiles=@(Get-ChildItem -LiteralPath $previous -File -Force|Sort-Object Name)
    $priorLeaf=@(Get-ChildItem -LiteralPath $legacy -File -Force|Sort-Object Name)
    if($priorFiles.Count -ne 26 -or $priorLeaf.Count -ne 13){throw 'Historical inventory'}
    $extra=@($priorFiles|ForEach-Object{$_.FullName})+@($priorLeaf|ForEach-Object{$_.FullName})
    foreach($path in @('docs/roadmap/2026-10-18-p2b-inbox-tree-copy-repair.md','docs/roadmap/2026-10-18-p2b-inbox-evidence-correction.md','docs/roadmap/2026-10-18-p2b-inbox-limited-acceptance-proposal.md','docs/roadmap/2026-10-18-p2b-inbox-limited-acceptance-approval.md')){$extra+=,(Join-Path $root $path)}
    $inputs=@(Get-ChildItem -LiteralPath $PSScriptRoot -File -Force|Sort-Object Name)
    $expectedNames=@('.gitattributes','capture.ps1','expected-output.json','handshake-seam.tests.ps1','README.md')
    if($inputs.Count -ne 5 -or @($inputs|Where-Object{$_.Name -cnotin $expectedNames}).Count -ne 0){throw 'Five exact prepared inputs required'}
    foreach($file in $inputs){if([IO.File]::ReadAllText($file.FullName).Contains([string][char]13)){throw 'New inputs require physical LF'};$extra+=,$file.FullName}
    $extra+=@($engine,$sdk)
    foreach($path in $extra){$pins+=,[pscustomobject]@{VerificationPath=$path;SHA256=(Get-FileHash -LiteralPath $path).Hash;Role='seam-frozen-input'}}
    $spec=Join-Path $root 'docs/superpowers/specs/2026-10-19-p2b-handshake-contract-design.md'
    Assert-Hash $spec '9C37A7AD0156DDB7732503E78F919FAA887B0708D3F91CFA5494C9BD367309CC'
    $plan=Join-Path $root 'docs/superpowers/plans/2026-10-19-p2b-handshake-seam-verification.md'
    $specSnapshot=Join-Path $packet 'approved-spec.snapshot.md';$planSnapshot=Join-Path $packet 'approved-plan.snapshot.md'
    $pins+=,[pscustomobject]@{VerificationPath=$specSnapshot;SHA256=(Get-FileHash -LiteralPath $spec).Hash;Role='approved-spec-snapshot'}
    $pins+=,[pscustomobject]@{VerificationPath=$planSnapshot;SHA256=(Get-FileHash -LiteralPath $plan).Hash;Role='approved-plan-snapshot'}
    if($extra.Count -ne 50 -or $pins.Count -ne 1378){throw '1378 mapped inventory required'}
    $pins|ForEach-Object{[Console]::Out.WriteLine(($_|ConvertTo-Json -Compress))}
    $null=New-Item -ItemType Directory -Path $packet
    [IO.File]::WriteAllBytes($specSnapshot,[IO.File]::ReadAllBytes($spec))
    [IO.File]::WriteAllBytes($planSnapshot,[IO.File]::ReadAllBytes($plan))
    Save-Json $manifest ([pscustomobject][ordered]@{Schema=1;BaseCommit='c1b87ccc604b488dae3f48e8325b1238cd3e1df3';NewBudgetApproved=$true;Pins=$pins;InheritedPins=1326;AdditionalPins=52;NewExtraPaths=$extra;PreviousPacketFiles=@($priorFiles.Name);PreviousPortableFiles=@($priorLeaf.Name);Order=@('full','final');MaxHosts=2;MaxReviewers=1;MaxMs=120000;ExitWaitMs=2000;DrainMs=2000;Context=('CONTEXT: '+$identity.Name+' admin=False');OldTreeHostsSpent=5;OldTreeReviewersSpent=1;OriginalHostsSpent=2;OriginalRemainingFrozen=3;DiagnosticHostsSpent=1;RealGateImplemented=$false;RuntimeApproved=$false;FormalAcceptance=$false})
    [pscustomobject]@{FrozenSHA256=(Get-FileHash -LiteralPath $manifest).Hash;Pins=$pins.Count;PreparedInputs=$inputs.Count}|ConvertTo-Json;return
}
Assert-Hash $manifest $FrozenSHA256
$f=Get-Content -LiteralPath $manifest -Raw|ConvertFrom-Json
foreach($pin in $f.Pins){Assert-Hash $pin.VerificationPath $pin.SHA256}
$rows=@();if(Test-Path -LiteralPath $index){$rows=@(Get-Content -LiteralPath $index -Raw|ConvertFrom-Json)}
foreach($r in $rows){
    if($r.Outcome -cne 'EXPECTED_WITHIN_SCOPE'){throw 'Prior nonexact/incomplete freezes remaining slots and review'}
    foreach($stream in @('stdout','stderr')){Assert-Hash (Join-Path $packet ($r.Slot+'.'+$stream+'.bin')) $r.($stream+'SHA256')}
    Assert-Hash (Join-Path $packet ($r.Slot+'.trace.json')) $r.TraceSHA256
}
if([string]::IsNullOrEmpty($Slot) -or $rows.Count -ge 2 -or $Slot -cne $f.Order[$rows.Count]){throw 'Ordered NEW2 only; no retry'}
$row=[pscustomobject][ordered]@{Number=($rows.Count+1);Slot=$Slot;AttemptConsumed=$true;Launched=$false;PID=$null;ElapsedMs=$null;TimedOut=$false;ExitCode=$null;OwnedExit=$false;StreamsComplete=$false;PostPinsIntact=$false;ExactExpected=$false;CrossRunEqual=$false;Outcome='INCOMPLETE';stdoutSHA256=$null;stderrSHA256=$null;TraceSHA256=$null;TraceDifferences=$null;RealGateImplemented=$false;RuntimeApproved=$false;FormalAcceptance=$false}
$rows+=,$row;Save-Json $index $rows
$trace=Join-Path $packet ($Slot+'.trace.json')
$p=[Diagnostics.Process]::new();$p.StartInfo=[Diagnostics.ProcessStartInfo]::new()
$p.StartInfo.FileName=$engine;$p.StartInfo.UseShellExecute=$false;$p.StartInfo.CreateNoWindow=$true;$p.StartInfo.RedirectStandardOutput=$true;$p.StartInfo.RedirectStandardError=$true
foreach($arg in @('-NoProfile','-NonInteractive','-File',(Join-Path $PSScriptRoot 'handshake-seam.tests.ps1'),'-TracePath',$trace)){$p.StartInfo.ArgumentList.Add($arg)}
$stdout=[IO.File]::Create((Join-Path $packet ($Slot+'.stdout.bin')));$stderr=[IO.File]::Create((Join-Path $packet ($Slot+'.stderr.bin')))
$timer=[Diagnostics.Stopwatch]::StartNew()
try{
    if(-not $p.Start()){throw 'Owned target did not start'}
    $row.Launched=$true;$row.PID=$p.Id;Save-Json $index $rows
    $ot=$p.StandardOutput.BaseStream.CopyToAsync($stdout);$et=$p.StandardError.BaseStream.CopyToAsync($stderr)
    if(-not $p.WaitForExit($f.MaxMs)){$row.TimedOut=$true;$p.Kill()}
    $row.OwnedExit=$p.WaitForExit($f.ExitWaitMs);if($row.OwnedExit){$row.ExitCode=$p.ExitCode}
    $row.StreamsComplete=[Threading.Tasks.Task]::WaitAll([Threading.Tasks.Task[]]@($ot,$et),$f.DrainMs)
}finally{$timer.Stop();$row.ElapsedMs=$timer.ElapsedMilliseconds;$stdout.Dispose();$stderr.Dispose();$p.Dispose();Save-Json $index $rows}
Assert-Hash $manifest $FrozenSHA256;Assert-Hash $engine $engineHash;Assert-Hash $sdk $sdkHash
foreach($pin in $f.Pins){Assert-Hash $pin.VerificationPath $pin.SHA256}
$row.PostPinsIntact=$true
foreach($stream in @('stdout','stderr')){$row.($stream+'SHA256')=(Get-FileHash -LiteralPath (Join-Path $packet ($Slot+'.'+$stream+'.bin'))).Hash}
if(Test-Path -LiteralPath $trace){$row.TraceSHA256=(Get-FileHash -LiteralPath $trace).Hash}
Save-Json $index $rows
$raw=[Text.UTF8Encoding]::new($false,$true).GetString([IO.File]::ReadAllBytes((Join-Path $packet ($Slot+'.stdout.bin'))))
$lf=[string][char]10;$crlf=[string][char]13+[char]10
$normal=$raw.Replace($crlf,$lf);if($normal.EndsWith($lf)){$normal=$normal.Substring(0,$normal.Length-1)}
$actual=@($normal.Split([char]10))
$oraclePath=Join-Path $PSScriptRoot 'expected-output.json';$oracle=Get-Content -LiteralPath $oraclePath -Raw|ConvertFrom-Json
$expected=@($oracle.Output|ForEach-Object{if($_ -ceq '{Context}'){$f.Context}else{$_}})
$same=$actual.Count -eq $expected.Count;if($same){for($i=0;$i -lt $actual.Count;$i++){if($actual[$i] -cne $expected[$i]){$same=$false;break}}}
$traceOK=$false;$script:diffs=[Collections.Generic.List[string]]::new()
if(Test-Path -LiteralPath $trace){
    $actualDoc=[System.Text.Json.JsonDocument]::Parse([string](Get-Content -LiteralPath $trace -Raw),[System.Text.Json.JsonDocumentOptions]::new())
    $oracleDoc=[System.Text.Json.JsonDocument]::Parse([string](Get-Content -LiteralPath $oraclePath -Raw),[System.Text.Json.JsonDocumentOptions]::new())
    try{Compare-ExactJson $actualDoc.RootElement $oracleDoc.RootElement.GetProperty('TraceExpected') 'trace';$traceOK=$script:diffs.Count -eq 0}finally{$actualDoc.Dispose();$oracleDoc.Dispose()}
}
$row.TraceDifferences=@($script:diffs)
$row.ExactExpected=$same -and $traceOK
if($Slot -ceq 'final'){$row.CrossRunEqual=$row.stdoutSHA256 -ceq $rows[0].stdoutSHA256 -and $row.TraceSHA256 -ceq $rows[0].TraceSHA256}
else{$row.CrossRunEqual=$true}
if(-not $row.TimedOut -and $row.OwnedExit -and $row.StreamsComplete -and $row.PostPinsIntact -and $row.ExactExpected -and $row.CrossRunEqual -and $row.ExitCode -eq 0 -and (Get-Item -LiteralPath (Join-Path $packet ($Slot+'.stderr.bin'))).Length -eq 0){$row.Outcome='EXPECTED_WITHIN_SCOPE'}
Save-Json $index $rows;$row|ConvertTo-Json -Depth 5;[Console]::Out.Write($raw)
if($row.Outcome -cne 'EXPECTED_WITHIN_SCOPE'){throw 'Unexpected result; freeze remaining slots/review, no repair/retry'}
