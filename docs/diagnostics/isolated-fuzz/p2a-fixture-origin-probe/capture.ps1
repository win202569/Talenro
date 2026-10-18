param([Parameter(Mandatory)][ValidateSet('Prepare','Run')][string]$Action,[string]$FrozenSHA256)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$root=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))
Set-Location -LiteralPath $root
$packet=Join-Path $root '.superpowers/sdd/2026-10-18-p2a-fixture-origin-probe'
$manifest=Join-Path $packet 'frozen.json';$index=Join-Path $packet 'invocation.json'
$target=Join-Path $PSScriptRoot 'fixture-origin.ps1'
function Save-Json($Path,$Value){[IO.File]::WriteAllText($Path,($Value|ConvertTo-Json -Depth 24),[Text.UTF8Encoding]::new($false))}
function Assert-Hash($Path,[string]$SHA256){if([string]::IsNullOrEmpty($SHA256) -or (Get-FileHash -LiteralPath $Path -Algorithm SHA256).Hash -cne $SHA256){throw ('Pin drift: '+$Path)}}
$identity=[Security.Principal.WindowsIdentity]::GetCurrent()
$principal=[Security.Principal.WindowsPrincipal]::new($identity)
if($principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)){throw 'Ordinary controller token required'}
if($Action -ceq 'Prepare'){
    if(Test-Path -LiteralPath $packet){throw 'New packet only; no retry/overwrite'}
    $s=Get-Content -LiteralPath (Join-Path $PSScriptRoot 'source-pins.json') -Raw|ConvertFrom-Json
    Assert-Hash $s.ApprovedPlan $s.ApprovedPlanSHA256
    Assert-Hash $s.OriginalTest.Path $s.OriginalTest.SHA256
    Assert-Hash $s.Engine $s.EngineSHA256
    $anchors=@()
    foreach($a in $s.Anchors.PSObject.Properties){
        $path=Join-Path $s.InheritedPacket $a.Name;Assert-Hash $path $a.Value
        $anchors+=,[pscustomobject]@{Path=$path;VerificationPath=$path;Role='previous-packet-anchor';SHA256=$a.Value}
    }
    if($anchors.Count -ne 4){throw 'Anchor inventory'}
    $old=Get-Content -LiteralPath (Join-Path $s.InheritedPacket 'frozen.json') -Raw|ConvertFrom-Json
    if(@($old.Pins).Count -ne 790){throw 'Inherited pin inventory'}
    $pins=@($old.Pins)+$anchors
    $files=@(Get-ChildItem -LiteralPath $PSScriptRoot -File -Force)
    if($files.Count -ne 6){throw 'Six portable files required'}
    foreach($file in $files){
        if([IO.File]::ReadAllText($file.FullName).Contains("`r")){throw 'Physical LF required'}
        $pins+=,[pscustomobject]@{Path=$file.FullName;VerificationPath=$file.FullName;Role='fixture-origin-input';SHA256=(Get-FileHash -LiteralPath $file.FullName).Hash}
    }
    if($pins.Count -ne 800){throw 'Common pin inventory'}
    foreach($pin in $pins){Assert-Hash $pin.VerificationPath $pin.SHA256}
    $tokens=$null;$errors=$null
    $original=[Management.Automation.Language.Parser]::ParseFile((Join-Path $root $s.OriginalTest.Path),[ref]$tokens,[ref]$errors)
    if($errors.Count){throw 'Original parse'}
    $constructor=@($original.FindAll({param($n)$n -is [Management.Automation.Language.FunctionDefinitionAst] -and $n.Name -ceq 'New-TBRequest'},$true))
    if($constructor.Count -ne 1 -or $constructor[0].Extent.Text -cne $s.OriginalConstructor.Definition){throw 'Original constructor drift'}
    $code=[IO.File]::ReadAllText($target)
    if($code -cne ($s.ExactTransformation.AddBefore+$s.OriginalConstructor.Definition+$s.ExactTransformation.AddAfter)){throw 'Exact transformation drift'}
    $tokens=$null;$errors=$null
    $ast=[Management.Automation.Language.Parser]::ParseFile($target,[ref]$tokens,[ref]$errors)
    if($errors.Count){throw 'Probe parse'}
    $commands=@($ast.FindAll({param($n)$n -is [Management.Automation.Language.CommandAst]},$true))
    if(@($commands|Where-Object{$_.GetCommandName() -ceq 'New-TBRequest'}).Count -ne 5){throw 'Five constructors only'}
    $allowed=@('Set-StrictMode','New-TBRequest','New-OriginFreshInt32','Origin-Reference','Origin-Need','Origin-Plain','Get-OriginSchemaView','Origin-Matches','ConvertTo-Json')
    foreach($c in $commands){if($c.GetCommandName() -cnotin $allowed -or $c.InvocationOperator -ne [Management.Automation.Language.TokenKind]::Unknown){throw 'Non-allowlisted command/import/invocation'}}
    $o=Get-Content -LiteralPath (Join-Path $PSScriptRoot 'expected-output.json') -Raw|ConvertFrom-Json
    $f=[pscustomobject]@{Schema=1;BaseCommit=$s.BaseCommit;ApprovedPlan=$s.ApprovedPlan;ApprovedPlanSHA256=$s.ApprovedPlanSHA256;Pins=$pins;CommonCount=800;OriginalCommonCount=790;Engine=$s.Engine;EngineSHA256=$s.EngineSHA256;Target=$target;Context=('CONTEXT: '+$identity.Name+' admin=False');Oracle=$o;MaxHosts=1;MaxReviewers=0;MaxMs=30000;ExitWaitMs=2000;DrainMs=2000;MaxRecordBytes=65536;ProductImported=$false;ProductCalled=$false;RepairPerformed=$false}
    $null=New-Item -Path $packet -ItemType Directory
    Save-Json $manifest $f
    (Get-FileHash -LiteralPath $manifest).Hash
    return
}
Assert-Hash $manifest $FrozenSHA256
$f=Get-Content -LiteralPath $manifest -Raw|ConvertFrom-Json
function Check-Pins{Assert-Hash $manifest $FrozenSHA256;Assert-Hash $f.Engine $f.EngineSHA256;foreach($pin in $f.Pins){Assert-Hash $pin.VerificationPath $pin.SHA256}}
Check-Pins
if(Test-Path -LiteralPath $index){throw 'Single attempt already consumed; no retry'}
foreach($name in @('stdout.bin','stderr.bin')){if(Test-Path -LiteralPath (Join-Path $packet $name)){throw 'Existing capture forbidden'}}
$row=[ordered]@{
    Schema=1;AttemptConsumed=$true;MaxHosts=1;ChildLaunched=$false;PID=$null;ParentAdministrator=$false
    Target=$target;TargetSHA256=(Get-FileHash -LiteralPath $target).Hash;Engine=$f.Engine;EngineSHA256=$f.EngineSHA256
    StartedUTC=[DateTime]::UtcNow.ToString('o');TimedOut=$false;ExitCode=$null;ElapsedMs=$null
    CleanupStatus='unknown';StreamsComplete=$false;PostPinsIntact=$false;StderrBytes=$null;LineCount=0
    Outcome='INCOMPLETE';Observation=$null;ControllerError=$null;ProductImported=$false;ProductCalled=$false;RepairPerformed=$false
}
Save-Json $index ([pscustomobject]$row) # Persist consumption BEFORE launch.
$psi=[Diagnostics.ProcessStartInfo]::new();$psi.FileName=$f.Engine
$psi.UseShellExecute=$false;$psi.CreateNoWindow=$true;$psi.WorkingDirectory=$root
$psi.RedirectStandardOutput=$true;$psi.RedirectStandardError=$true
foreach($arg in @('-NoProfile','-NonInteractive','-File',$target)){$psi.ArgumentList.Add($arg)}
$p=[Diagnostics.Process]::new();$p.StartInfo=$psi
$of=$null;$ef=$null;$ot=$null;$et=$null;$watch=[Diagnostics.Stopwatch]::new()
try{
    $of=[IO.File]::Open((Join-Path $packet 'stdout.bin'),[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::Read)
    $ef=[IO.File]::Open((Join-Path $packet 'stderr.bin'),[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::Read)
    $watch.Start();if(-not $p.Start()){throw 'Launch failed'}
    $row.ChildLaunched=$true;$row.PID=$p.Id;Save-Json $index ([pscustomobject]$row)
    $ot=$p.StandardOutput.BaseStream.CopyToAsync($of);$et=$p.StandardError.BaseStream.CopyToAsync($ef)
    if(-not $p.WaitForExit([Math]::Max(0,30000-[int]$watch.ElapsedMilliseconds))){$row.TimedOut=$true}
}catch{$row.ControllerError=$_.Exception.Message}finally{
    if($row.ChildLaunched){try{
        if(-not $p.HasExited){$p.Kill();if(-not $p.WaitForExit(2000)){throw 'Owned exit unknown'}}
        $row.CleanupStatus='owned-host-exited';$row.ExitCode=$p.ExitCode
        if($null -eq $ot -or $null -eq $et){throw 'Missing stream tasks'}
        if(-not [Threading.Tasks.Task]::WaitAll([Threading.Tasks.Task[]]@($ot,$et),2000)){throw 'Bounded drain incomplete'}
        [void]$ot.GetAwaiter().GetResult();[void]$et.GetAwaiter().GetResult();$row.StreamsComplete=$true
    }catch{$row.ControllerError=([string]$row.ControllerError+' | '+$_.Exception.Message)}}
    $watch.Stop();$row.ElapsedMs=$watch.ElapsedMilliseconds
    if($null -ne $of){$of.Dispose()};if($null -ne $ef){$ef.Dispose()}
    foreach($stream in @('stdout','stderr')){$path=Join-Path $packet ($stream+'.bin');if(Test-Path -LiteralPath $path){$row[$stream+'SHA256']=(Get-FileHash -LiteralPath $path).Hash}}
    try{Check-Pins;$row.PostPinsIntact=$true}catch{$row.ControllerError=([string]$row.ControllerError+' | '+$_.Exception.Message)}
    $p.Dispose();Save-Json $index ([pscustomobject]$row)
}
if($row.StreamsComplete){try{
    $raw=[IO.File]::ReadAllText((Join-Path $packet 'stdout.bin'),[Text.UTF8Encoding]::new($false,$true))
    $canonical=$raw.Replace("`r`n","`n")
    if($canonical.Contains("`r") -or -not $canonical.EndsWith("`n")){throw 'Stream framing'}
    $lines=@($canonical.Substring(0,$canonical.Length-1) -split "`n");$row.LineCount=$lines.Count
    $row.StderrBytes=(Get-Item -LiteralPath (Join-Path $packet 'stderr.bin')).Length
    $valid=$row.ChildLaunched -and -not $row.TimedOut -and $row.ElapsedMs -le 34000 -and $row.CleanupStatus -ceq 'owned-host-exited' -and $row.PostPinsIntact -and $row.StderrBytes -eq 0 -and $null -eq $row.ControllerError
    $o=$f.Oracle
    if($lines.Count -ne 4 -or $lines[0] -cne $f.Context -or $lines[1] -cne $o.Host -or $lines[3] -cne $o.Limit -or -not $lines[2].StartsWith($o.RecordPrefix,[StringComparison]::Ordinal)){throw 'Four-line envelope required'}
    $json=$lines[2].Substring($o.RecordPrefix.Length)
    if([Text.UTF8Encoding]::new($false).GetByteCount($json) -gt 65536){throw 'Oversized observation'}
    $obs=$json|ConvertFrom-Json;$row.Observation=$obs
    $names=@($obs.PSObject.Properties.Name)
    if($names.Count -ne $o.RecordFields.Count -or @(Compare-Object ($names|Sort-Object) ($o.RecordFields|Sort-Object)).Count -ne 0){throw 'Record schema'}
    if($obs.Schema -ne 1){throw 'Record schema version'}
    foreach($field in $o.RequiredFalse){if($obs.$field -isnot [bool] -or $obs.$field -ne $false){throw 'Product/repair flag'}}
    if($obs.Outcome -ceq 'INCOMPLETE'){throw 'Probe reported incomplete'}
    if($obs.Stage -cne 'complete' -or $obs.ErrorCode -cne '' -or $obs.ConstructorCalls -ne 5){throw 'Completed fixture inventory'}
    foreach($label in $o.FixtureLabels){
        if($obs.Types.$label -isnot [string] -or $obs.Types.$label -cne $o.RequiredType -or
            ($obs.Values.$label -isnot [int] -and $obs.Values.$label -isnot [long]) -or $obs.Values.$label -ne 1){throw 'Exact numeric fixture facts'}
    }
    foreach($field in $o.RequiredTrue){if($obs.$field -isnot [bool] -or -not $obs.$field){throw ('Fixture/control invalid: '+$field)}}
    $chain=$true
    foreach($field in $o.ChainFields){if($obs.$field -isnot [bool]){throw 'Unknown sharing fact'};if(-not $obs.$field){$chain=$false}}
    if($valid -and $chain -and $obs.Outcome -ceq $o.Confirmed.Outcome -and $row.ExitCode -eq $o.Confirmed.ExitCode){$row.Outcome=$o.Confirmed.Outcome}
    elseif($valid -and -not $chain -and $obs.Outcome -ceq $o.NotConfirmed.Outcome -and $row.ExitCode -eq $o.NotConfirmed.ExitCode){$row.Outcome=$o.NotConfirmed.Outcome}
}catch{$row.ControllerError=([string]$row.ControllerError+' | '+$_.Exception.Message)}}
Save-Json $index ([pscustomobject]$row)
[pscustomobject]$row|ConvertTo-Json -Depth 12 -Compress
if($row.Outcome -ceq 'INCOMPLETE'){throw 'STOP: fixture diagnosis incomplete; allowance spent, no retry/repair'}
# Stop after either diagnostic result. No further host, product run or source fix.
