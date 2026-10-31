param(
    [Parameter(Mandatory)][ValidateSet('Prepare','RunOne')][string]$Action,
    [Parameter(Mandatory)][ValidateSet('seed','isolation-red','isolation-green')][string]$Grant,
    [ValidatePattern('^[A-F0-9]{64}$')][string]$FrozenSHA256,
    [ValidatePattern('^[A-F0-9]{64}$')][string]$PreviousFrozenSHA256
)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$repo=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))
$packet=Join-Path $repo '.superpowers/sdd/2026-10-31-offline-record-isolation-repair'
$diagnostic=Join-Path $repo '.superpowers/sdd/2026-10-31-offline-record-chain-diagnosis'
$plan=Join-Path $repo 'docs/superpowers/plans/2026-10-31-offline-record-isolation-repair.md'
$spec=Join-Path $repo 'docs/superpowers/specs/2026-10-19-offline-diagnostic-entry-migration-design.md'
$engine='C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe'
$inputs=@('.gitattributes','README.md','expected-output.json','case-module.ps1','publisher.psm1','isolation.tests.ps1','capture.ps1','driver.ps1')
$writer=Join-Path $repo 'scripts/offline-diagnostics-v2/private/write-boundary.ps1'
function RID-Need([bool]$Value,[string]$Message){if(-not $Value){throw $Message}}
function RID-Hash([string]$Path){(Get-FileHash -LiteralPath $Path -Algorithm SHA256).Hash}
function RID-Once([string]$Path,[byte[]]$Bytes){
    $stream=$null;$first=$null
    try{$stream=[IO.File]::Open($Path,[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::None);$stream.Write($Bytes,0,$Bytes.Length);$stream.Flush()}catch{$first=$_}
    finally{if($null -ne $stream){try{$stream.Dispose()}catch{if($null -eq $first){$first=$_}else{throw}}}}
    if($null -ne $first){throw $first}
}
function RID-JSON([string]$Path,$Value){RID-Once $Path ([Text.UTF8Encoding]::new($false).GetBytes((ConvertTo-Json -InputObject $Value -Depth 45)))}
function RID-Verify($Pins){
    $seen=@{}
    foreach($pin in $Pins){
        $path=[IO.Path]::GetFullPath($pin.Path)
        if($seen.ContainsKey($path)){RID-Need ($seen[$path] -ceq $pin.SHA256) 'Conflicting duplicate pin'}else{$seen[$path]=$pin.SHA256}
        RID-Need ((RID-Hash $path) -ceq $pin.SHA256) ('Pin drift '+$path)
    }
}
function RID-Static {
    foreach($file in $inputs){
        $path=Join-Path $PSScriptRoot $file
        RID-Need ([IO.File]::Exists($path)) ('Missing input '+$file)
        RID-Need (-not ([IO.File]::ReadAllBytes($path) -contains 13)) ('Non-LF input '+$file)
        if($file.EndsWith('.ps1') -or $file.EndsWith('.psm1')){$tokens=$null;$parseErrors=$null;$null=[Management.Automation.Language.Parser]::ParseFile($path,[ref]$tokens,[ref]$parseErrors);RID-Need ($parseErrors.Count -eq 0) ('AST invalid '+$file)}
    }
    $null=Get-Content -LiteralPath (Join-Path $PSScriptRoot expected-output.json) -Raw|ConvertFrom-Json -AsHashtable
}
function RID-Keys($Value,[object[]]$Keys){RID-Need ($Value -is [Collections.IDictionary] -and ($Value.Keys -join ',') -ceq ($Keys -join ',')) ('Exact keys '+($Keys -join ','))}
function RID-Literal($Actual,$Expected){
    if($null -eq $Expected){RID-Need ($null -eq $Actual) 'Expected null';return}
    if($Expected -is [Collections.IDictionary]){RID-Keys $Actual @($Expected.Keys);foreach($k in $Expected.Keys){RID-Literal $Actual[$k] $Expected[$k]};return}
    if($Expected -is [array]){RID-Need ($Actual -is [array] -and $Actual.Count -eq $Expected.Count) 'Exact array';for($n=0;$n -lt $Expected.Count;$n++){RID-Literal $Actual[$n] $Expected[$n]};return}
    if($Expected -is [bool]){RID-Need ($Actual -is [bool] -and $Actual -eq $Expected) 'Exact boolean';return}
    if($Expected -is [string]){RID-Need ($Actual -is [string] -and $Actual -ceq $Expected) 'Exact string';return}
    RID-Need (($Actual -is [long] -or $Actual -is [int]) -and $Actual -eq $Expected) 'Exact integer'
}
function RID-ErrorContract($Actual,[string]$Path,[string]$Role,$Oracle){
    RID-Keys $Actual $Oracle.ErrorKeys
    RID-Literal $Actual.Type $Oracle.ErrorType;RID-Literal $Actual.ErrorId $Oracle.ErrorId
    RID-Literal $Actual.Message ($Oracle.ErrorMessage.Replace('{path}',$Path))
    $expectedSource=if($Role -ceq 'Case'){$writer}else{Join-Path $PSScriptRoot publisher.psm1}
    RID-Literal $Actual.Source $expectedSource
    $sourceLines=[IO.File]::ReadAllLines($expectedSource)
    $originLines=if($Role -ceq 'Case'){@(3,12)}else{@(for($n=0;$n -lt $sourceLines.Length;$n++){if($sourceLines[$n].Contains('[IO.File]::Open(')){$n+1}})}
    RID-Need (($Actual.Line -is [long]) -and $Actual.Line -cin $originLines -and $Actual.Column -is [long] -and $Actual.Column -gt 0 -and $Actual.Column -le $sourceLines[$Actual.Line-1].Length+1) 'Pinned error invocation site'
    RID-Need ($Actual.Position -is [string] -and $Actual.Position.StartsWith(('At '+$expectedSource+':'+$Actual.Line+' char:'+$Actual.Column),[StringComparison]::Ordinal)) 'Pinned Position header'
    RID-Need ($Actual.Position.Contains("`n+") -and $Actual.Stack -is [string] -and $Actual.Stack.Length -gt 0) 'Complete Position/Stack'
    $allowed=@{}
    $allowed[$writer]=@(3,8,12)
    foreach($file in @('isolation.tests.ps1','capture.ps1','publisher.psm1')){
        $full=Join-Path $PSScriptRoot $file;$lines=[IO.File]::ReadAllLines($full)
        $allowed[$full]=@(for($n=0;$n -lt $lines.Length;$n++){if($lines[$n].Contains('# RI-CALL')){$n+1}})
    }
    $frames=@($Actual.Stack -split '\r?\n');RID-Need ($frames.Count -ge 2 -and $frames.Count -le 12) 'Bounded stack'
    foreach($frame in $frames){
        RID-Need ($frame -cmatch '^at ([^,]+), (.+): line ([0-9]+)$') 'Stack frame format'
        $framePath=$Matches[2];$frameLine=[int]$Matches[3]
        RID-Need ($allowed.ContainsKey($framePath) -and $frameLine -cin $allowed[$framePath]) 'Frozen deterministic stack call site'
    }
}
function RID-PublisherView($View,[string]$Case,[string]$Root,[bool]$Healthy,$Oracle){
    RID-Keys $View $Oracle.PublisherKeys
    RID-Literal $View.Case $Case
    RID-Literal $View.Path (Join-Path $Root $(if($Healthy){'healthy.bin'}else{'existing-publisher.bin'}))
    RID-Literal $View.Status $(if($Healthy){'Completed'}else{'Failed'});RID-Literal $View.Consumed $true
    RID-Literal $View.Open 1
    foreach($key in @('Write','Flush','Dispose')){RID-Literal $View[$key] $(if($Healthy){1}else{0})}
    RID-Literal $View.Secondary @();RID-Literal $View.Overflow $false
    if($Healthy){RID-Literal $View.Primary $null}else{RID-ErrorContract $View.Primary $View.Path Publisher $Oracle}
}
function RID-Evidence($Evidence,[bool]$Green,[string]$Root,$Oracle){
    RID-Keys $Evidence $Oracle.RootKeys
    RID-Literal $Evidence.Schema 1;RID-Literal $Evidence.Grant $Grant;RID-Literal $Evidence.Status 'COMPLETE'
    RID-Literal $Evidence.NotRun @();RID-Literal $Evidence.PrimaryError $null;RID-Literal $Evidence.SecondaryErrors @()
    foreach($flag in @('Overflow','SourceVerified','LiveActionsPerformed','RuntimeApproved','FormalAcceptance','RealGateImplemented')){RID-Literal $Evidence[$flag] $false}
    RID-Need ($Evidence.Groups -is [array] -and $Evidence.Groups.Count -eq 9) 'Nine groups'
    $same=$Green
    $facts=@(
        [ordered]@{SameOriginal=$same},[ordered]@{SameOriginal=$same},[ordered]@{SameOriginal=$same},[ordered]@{SameOriginal=$same},
        [ordered]@{InnerSentinel=$true;OuterSentinel=$false;OuterPrimaryNull=$true;OuterSecondaryCount=0;OuterOverflow=$false},
        [ordered]@{InnerChanged=$true;Calls=1;OuterSame=$same},
        [ordered]@{SameError=$true;SecondaryCount=0;SignatureCorrect=$true},
        [ordered]@{Ok=$true;Rejected=$false;BytesExact=$true;Status='Completed';Open=1;Write=1;Flush=1;Dispose=1;Exports=0},
        [ordered]@{FirstOk=$false;FirstRejected=$false;SignatureCorrect=$true;SameError=$true;SecondOk=$false;SecondRejected=$true;SecondErrorNull=$true;IOAfterSecond=0;Status='Failed';Open=1;Write=0;Flush=0;Dispose=0}
    )
    $passes=if($Green){$Oracle.GreenPassed}else{$Oracle.RedPassed}
    for($i=0;$i -lt 9;$i++){RID-Keys $Evidence.Groups[$i] $Oracle.GroupKeys;RID-Literal $Evidence.Groups[$i].Name $Oracle.Names[$i];RID-Literal $Evidence.Groups[$i].Passed $passes[$i];RID-Literal $Evidence.Groups[$i].Facts $facts[$i]}
    $o=$Evidence.Observations;RID-Keys $o $Oracle.ObservationKeys
    $moduleName=$o.CaseModule.Name;$moduleGuid=$o.CaseModule.Guid
    RID-Keys $o.CaseModule $Oracle.ModuleKeys
    RID-Need ($moduleName -is [string] -and $moduleName -cmatch '^__DynamicModule_[a-f0-9-]{36}$' -and $moduleGuid -is [string] -and $moduleGuid -cmatch '^[a-f0-9-]{36}$') 'Actual dynamic module identity'
    RID-Literal $o.CaseModule.Functions $(if($Green){0}else{4})
    foreach($key in @('Aliases','Variables','Cmdlets')){RID-Literal $o.CaseModule[$key] 0}
    foreach($stage in @('CommandsBefore','CommandsCreated','CommandsReplaced')){
        RID-Need ($o[$stage] -is [array] -and $o[$stage].Count -eq 4) 'Four identities per stage'
        for($i=0;$i -lt 4;$i++){
            $c=$o[$stage][$i];RID-Keys $c $Oracle.CommandKeys;RID-Literal $c.Name $Oracle.Commands[$i]
            $outer=$stage -ceq 'CommandsBefore' -or $Green
            RID-Literal $c.SameOriginal $outer;RID-Literal $c.ModuleName $(if($outer){''}else{$moduleName});RID-Literal $c.ModuleGuid $(if($outer){$null}else{$moduleGuid})
        }
    }
    RID-Literal $o.OuterState ([ordered]@{PrimaryNull=$true;SecondaryCount=0;Overflow=$false;SentinelPresent=$false;SentinelValue=$null})
    RID-Literal $o.InnerState ([ordered]@{PrimaryNull=$false;SecondaryCount=0;Overflow=$false;SentinelPresent=$true;SentinelValue=$Oracle.Sentinel})
    RID-Literal $o.Factory ([ordered]@{Changed=$true;Calls=1})
    RID-ErrorContract $o.CaseThrown (Join-Path $Root case-existing.bin) Case $Oracle
    RID-Literal $o.CasePrimary $o.CaseThrown
    RID-Keys $o.HealthyPublisher @('Result','View','Bytes','Exports')
    RID-Literal $o.HealthyPublisher.Result ([ordered]@{Ok=$true;Rejected=$false;Error=$null})
    RID-PublisherView $o.HealthyPublisher.View Healthy $Root $true $Oracle
    RID-Literal $o.HealthyPublisher.Bytes ([Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes($Oracle.HealthyText)));RID-Literal $o.HealthyPublisher.Exports 0
    RID-Keys $o.FailedPublisher @('First','Second','View','IOAfterSecond','SameError')
    RID-Keys $o.FailedPublisher.First $Oracle.ResultKeys;RID-Literal $o.FailedPublisher.First.Ok $false;RID-Literal $o.FailedPublisher.First.Rejected $false
    RID-ErrorContract $o.FailedPublisher.First.Error (Join-Path $Root existing-publisher.bin) Publisher $Oracle
    RID-Literal $o.FailedPublisher.Second ([ordered]@{Ok=$false;Rejected=$true;Error=$null})
    RID-PublisherView $o.FailedPublisher.View ExistingFailure $Root $false $Oracle
    RID-Literal $o.FailedPublisher.View.Primary $o.FailedPublisher.First.Error
    RID-Literal $o.FailedPublisher.IOAfterSecond 0;RID-Literal $o.FailedPublisher.SameError $true
    foreach($file in @('case-existing.bin','existing-publisher.bin')){RID-Need ([Linq.Enumerable]::SequenceEqual([byte[]][IO.File]::ReadAllBytes((Join-Path $Root $file)),[byte[]][Text.Encoding]::ASCII.GetBytes($Oracle.PriorText))) 'Existing file preserved'}
    RID-Need ([Linq.Enumerable]::SequenceEqual([byte[]][IO.File]::ReadAllBytes((Join-Path $Root healthy.bin)),[byte[]][Text.Encoding]::UTF8.GetBytes($Oracle.HealthyText))) 'Healthy actual bytes'
}
function Invoke-RIDriver {
    RID-Need ((RID-Hash $plan) -ceq '790571C1F49DB0ACD903605AD1F9B4CAC5F31A59B02C232B62925324B35E3826') 'Approved plan bytes'
    RID-Need ((RID-Hash $engine) -ceq '362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139') 'Pinned engine'
    RID-Need ((RID-Hash (Join-Path ([IO.Path]::GetDirectoryName($engine)) System.Management.Automation.dll)) -ceq '2D8AF026E589C5603629BFC3E7C57265B4167B12ECA3AE549C6A6C3CA7E4967B') 'Pinned SMA'
    $identity=[Security.Principal.WindowsIdentity]::GetCurrent();RID-Need ($identity.Name -ceq 'DESKTOP-H4SVCLU\Lenovo' -and -not ([Security.Principal.WindowsPrincipal]::new($identity)).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) 'Fixed context'
    RID-Need (-not (Test-Path -LiteralPath (Join-Path $packet stop.json))) 'Execution stopped; no retry'
    RID-Static
    if($Action -ceq 'Prepare'){
        RID-Need ($Grant -cin @('seed','isolation-green')) 'Prepare phase'
        $target=Join-Path $packet ($Grant+'.frozen.json');RID-Need (-not (Test-Path -LiteralPath $target)) 'Freeze attempted; no retry'
        if($Grant -ceq 'seed'){
            $existing=@(Get-ChildItem -LiteralPath $packet -Force|Sort-Object Name);RID-Need (($existing.Name -join ',') -ceq 'progress.md,static.tool-result.json') 'Seed packet only explicit ledger/static record'
            RID-Need ((RID-Hash (Join-Path $repo '.superpowers/sdd/2026-10-31-offline-diagnostic-review-repair/seed.frozen.json')) -ceq '9091825652E469001E51D8A510B12E0F84013C38509D740C649942DC2EAAFF6B') 'Original seed'
            RID-Need ((RID-Hash (Join-Path $repo 'docs/diagnostics/isolated-fuzz/p2b-record-chain-diagnosis/portable-final.json')) -ceq 'F47946FC985FE6F80B98D20057E38944FC8AE46DEA0F03EFC4BB1E56CD0459DB') 'Diagnostic portable'
            $pins=[Collections.Generic.List[object]]::new()
            $inherited=Get-Content -LiteralPath (Join-Path $diagnostic pins.json) -Raw|ConvertFrom-Json;RID-Need ($inherited.Count -eq 2365) 'Inherited2365'
            foreach($pin in $inherited){$pins.Add($pin)};RID-Verify $pins
            $raw=@(Get-ChildItem -LiteralPath $diagnostic -Recurse -File|Sort-Object FullName);RID-Need ($raw.Count -eq 15) 'Diagnostic15originals'
            $archive=Get-Content -LiteralPath (Join-Path $repo 'docs/diagnostics/isolated-fuzz/p2b-record-chain-diagnosis/portable-final.json') -Raw|ConvertFrom-Json
            RID-Need ($archive.Artifacts.Count -eq 15 -and $archive.Directories.Count -eq 2) 'Diagnostic archive inventory'
            foreach($a in $archive.Artifacts){RID-Need (-not [IO.Path]::IsPathRooted($a.Path) -and $a.Path -notmatch '(^|[\\/])\.\.([\\/]|$)') 'Archive path';$b=[Convert]::FromBase64String($a.Data);$original=Join-Path $diagnostic $a.Path;RID-Need ($b.Length -eq $a.Length -and (RID-Hash $original) -ceq $a.SHA256 -and [Linq.Enumerable]::SequenceEqual([byte[]][IO.File]::ReadAllBytes($original),[byte[]]$b)) 'Diagnostic actual bytes'}
            RID-Once (Join-Path $packet approved-plan.snapshot.md) ([IO.File]::ReadAllBytes($plan))
            RID-Once (Join-Path $packet approved-spec.snapshot.md) ([IO.File]::ReadAllBytes($spec))
            $extra=@($raw.FullName)+@((Join-Path $repo 'docs/diagnostics/isolated-fuzz/p2b-record-chain-diagnosis/portable-final.json'),(Join-Path $repo 'docs/roadmap/2026-10-31-offline-record-chain-diagnosis.md'),(Join-Path $repo 'docs/roadmap/2026-10-31-offline-diagnostic-review-repair.md'),$plan,(Join-Path $packet approved-plan.snapshot.md),(Join-Path $packet approved-spec.snapshot.md))+@($inputs|ForEach-Object {Join-Path $PSScriptRoot $_})
            RID-Need ($extra.Count -eq 29) 'Exact29new pins'
            foreach($path in $extra){$pins.Add([pscustomobject][ordered]@{Path=$path;SHA256=(RID-Hash $path)})}
            $manifest=[ordered]@{Schema=1;Phase=$Grant;BaseCommit='e75af9d17ab57771f752a4d61f9ffa7717163710';PlanSHA256=(RID-Hash $plan);Pins=$pins.ToArray();ChangedInputs=@();MaxHosts=2;MaxTargets=0;MaxCompilations=0;MaxReviews=0}
        }else{
            RID-Need (-not [string]::IsNullOrEmpty($PreviousFrozenSHA256)) 'Previous actual SHA required'
            $previous=Join-Path $packet seed.frozen.json;RID-Need ((RID-Hash $previous) -ceq $PreviousFrozenSHA256) 'Previous manifest identity'
            $red=Get-Content -LiteralPath (Join-Path $packet isolation-red.final.json) -Raw|ConvertFrom-Json;RID-Need $red.Exact 'Exact RED required'
            $manifest=Get-Content -LiteralPath $previous -Raw|ConvertFrom-Json
            $builder=Join-Path $PSScriptRoot case-module.ps1;$changed=0
            foreach($pin in $manifest.Pins){if($pin.Path -ceq $builder){$hash=RID-Hash $builder;RID-Need ($hash -cne $pin.SHA256) 'Builder must change';$pin.SHA256=$hash;$changed++}else{RID-Need ((RID-Hash $pin.Path) -ceq $pin.SHA256) ('No other changes '+$pin.Path)}}
            RID-Need ($changed -eq 1) 'Only builder changes';$manifest.Phase=$Grant;$manifest.ChangedInputs=@('case-module.ps1')
        }
        RID-Need ($manifest.Pins.Count -eq 2394) 'Exact2394pins';RID-Verify $manifest.Pins
        $snapshot=Join-Path $packet ('snapshots/'+$Grant);RID-Need (-not (Test-Path -LiteralPath $snapshot)) 'Snapshot attempted'
        $null=[IO.Directory]::CreateDirectory($snapshot)
        foreach($file in $inputs){RID-Once (Join-Path $snapshot $file) ([IO.File]::ReadAllBytes((Join-Path $PSScriptRoot $file)))}
        RID-JSON $target $manifest
        [ordered]@{Prepared=$Grant;FrozenSHA256=(RID-Hash $target);Pins=2394;Snapshots=8;ChangedInputs=$manifest.ChangedInputs}|ConvertTo-Json;return
    }
    RID-Need ($Grant -cin @('isolation-red','isolation-green') -and -not [string]::IsNullOrEmpty($FrozenSHA256)) 'Actual RunOne manifest required'
    $phase=if($Grant -ceq 'isolation-red'){'seed'}else{'isolation-green'}
    $manifestPath=Join-Path $packet ($phase+'.frozen.json');RID-Need ((RID-Hash $manifestPath) -ceq $FrozenSHA256) 'Actual freeze identity'
    $manifest=Get-Content -LiteralPath $manifestPath -Raw|ConvertFrom-Json;RID-Need ($manifest.Pins.Count -eq 2394 -and $manifest.Phase -ceq $phase) 'Manifest phase/count';RID-Verify $manifest.Pins
    foreach($suffix in @('intent.json','runtime.json','final.json','stdout.bin','stderr.bin')){RID-Need (-not (Test-Path -LiteralPath (Join-Path $packet ($Grant+'.'+$suffix)))) 'Attempt exists; no retry'}
    RID-Need (-not (Test-Path -LiteralPath (Join-Path $packet $Grant))) 'Capture phase already exists'
    RID-JSON (Join-Path $packet ($Grant+'.intent.json')) ([ordered]@{Grant=$Grant;AttemptConsumed=$true;FrozenSHA256=$FrozenSHA256;Host=1;Targets=0;Compilations=0;Reviews=0})
    $row=[ordered]@{Grant=$Grant;Started=$false;PID=$null;OwnedExit=$false;StreamsComplete=$false;ExitCode=$null;TimedOut=$false;TerminationRequested=$false;Errors=@();PostPinsIntact=$false;Exact=$false;EvidenceValid=$false;EvidenceState='UNKNOWN';Reason=$null;SourceVerified=$false;LiveActionsPerformed=$false;RuntimeApproved=$false;FormalAcceptance=$false;RealGateImplemented=$false}
    $errors=[Collections.Generic.List[object]]::new()
    function RID-ProcessError($Record){$errors.Add([ordered]@{Message=$Record.Exception.Message;ErrorId=$Record.FullyQualifiedErrorId;Stack=$Record.ScriptStackTrace})}
    $process=[Diagnostics.Process]::new();$out=$null;$err=$null;$ot=$null;$et=$null;$timer=[Diagnostics.Stopwatch]::StartNew()
    try{
        $out=[IO.File]::Open((Join-Path $packet ($Grant+'.stdout.bin')),[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::Read)
        $err=[IO.File]::Open((Join-Path $packet ($Grant+'.stderr.bin')),[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::Read)
        $process.StartInfo=[Diagnostics.ProcessStartInfo]::new();$process.StartInfo.FileName=$engine;$process.StartInfo.UseShellExecute=$false;$process.StartInfo.CreateNoWindow=$true
        $process.StartInfo.RedirectStandardOutput=$true;$process.StartInfo.RedirectStandardError=$true;$process.StartInfo.WorkingDirectory=$repo
        foreach($arg in @('-NoLogo','-NoProfile','-NonInteractive','-File',(Join-Path $PSScriptRoot capture.ps1),'-Grant',$Grant,'-FrozenSHA256',$FrozenSHA256)){$process.StartInfo.ArgumentList.Add($arg)}
        if(-not $process.Start()){throw 'Capture host did not start'}
        $row.Started=$true;$row.PID=$process.Id;$ot=$process.StandardOutput.BaseStream.CopyToAsync($out);$et=$process.StandardError.BaseStream.CopyToAsync($err)
        if(-not $process.WaitForExit([Math]::Max(0,(120000-[int]$timer.ElapsedMilliseconds)))){$row.TimedOut=$true}
    }catch{RID-ProcessError $_}
    finally{
        $associated=$false
        try{$property=[Diagnostics.Process].GetProperty('Id',([Reflection.BindingFlags]::Public -bor [Reflection.BindingFlags]::Instance));$id=$property.GetValue($process);if($id -is [int] -and $id -gt 0){$associated=$true;$row.PID=$id}}catch{if($row.Started){RID-ProcessError $_}}
        if($associated){$exited=$false;try{$exited=$process.HasExited}catch{RID-ProcessError $_};if(-not $exited){$row.TerminationRequested=$true;try{$process.Kill()}catch{RID-ProcessError $_}};try{$row.OwnedExit=$process.WaitForExit(2000);if($row.OwnedExit){$row.ExitCode=$process.ExitCode}}catch{RID-ProcessError $_}}
        if($null -ne $ot -and $null -ne $et){try{$row.StreamsComplete=[Threading.Tasks.Task]::WaitAll([Threading.Tasks.Task[]]@($ot,$et),2000)}catch{RID-ProcessError $_}}
        foreach($stream in @($out,$err)){if($null -ne $stream){try{$stream.Dispose()}catch{RID-ProcessError $_}}}
        try{$process.Dispose()}catch{RID-ProcessError $_};$timer.Stop()
    }
    try{RID-Verify $manifest.Pins;$row.PostPinsIntact=$true}catch{RID-ProcessError $_}
    $row.Errors=$errors.ToArray();RID-JSON (Join-Path $packet ($Grant+'.runtime.json')) $row
    try{
        RID-Need ($row.Started -and $row.OwnedExit -and $row.StreamsComplete -and -not $row.TimedOut -and -not $row.TerminationRequested -and $errors.Count -eq 0 -and $row.PostPinsIntact) 'Runtime exit/drain/pins incomplete'
        $oracle=Get-Content -LiteralPath (Join-Path $PSScriptRoot expected-output.json) -Raw|ConvertFrom-Json -AsHashtable
        $green=$Grant -ceq 'isolation-green';$passes=if($green){$oracle.GreenPassed}else{$oracle.RedPassed};$expectedExit=if($green){$oracle.GreenExit}else{$oracle.RedExit}
        $lines=@(for($i=0;$i -lt 9;$i++){($(if($passes[$i]){'PASS '}else{'FAIL '}))+$oracle.Names[$i]})
        $stdout=[Text.UTF8Encoding]::new($false,$true).GetString([IO.File]::ReadAllBytes((Join-Path $packet ($Grant+'.stdout.bin'))))
        RID-Need ($stdout -ceq (($lines -join [Environment]::NewLine)+[Environment]::NewLine) -and (Get-Item -LiteralPath (Join-Path $packet ($Grant+'.stderr.bin'))).Length -eq 0 -and $row.ExitCode -eq $expectedExit) 'Exact raw/exit contract'
        $root=Join-Path $packet $Grant;$evidencePath=Join-Path $root evidence.json
        RID-Need ((Get-Item -LiteralPath $evidencePath).Length -gt 0 -and (Get-Item -LiteralPath $evidencePath).Length -le 1048576) 'Evidence size'
        $e=Get-Content -LiteralPath $evidencePath -Raw|ConvertFrom-Json -AsHashtable
        RID-Evidence $e $green $root $oracle
        $journal=[Text.UTF8Encoding]::new($false,$true).GetString([IO.File]::ReadAllBytes((Join-Path $root journal.jsonl)))
        RID-Need ($journal -ceq ('{"Stage":"begin"}'+"`n"+'{"Stage":"tests-complete"}'+"`n"+'{"Stage":"published"}'+"`n")) 'Complete journal'
        $row.Exact=$true;$row.EvidenceValid=$true;$row.EvidenceState='COMPLETE'
    }catch{$row.Reason=[ordered]@{Message=$_.Exception.Message;ErrorId=$_.FullyQualifiedErrorId;Position=$_.InvocationInfo.PositionMessage;Stack=$_.ScriptStackTrace}}
    RID-JSON (Join-Path $packet ($Grant+'.final.json')) $row
    [pscustomobject]$row|ConvertTo-Json -Depth 12
    RID-Need $row.Exact 'Nonexact/UNKNOWN; freeze remainder'
}
try{Invoke-RIDriver}catch{
    if((Test-Path -LiteralPath $packet) -and -not (Test-Path -LiteralPath (Join-Path $packet stop.json))){RID-JSON (Join-Path $packet stop.json) ([ordered]@{Action=$Action;Grant=$Grant;Message=$_.Exception.Message;ErrorId=$_.FullyQualifiedErrorId;Position=$_.InvocationInfo.PositionMessage;Stack=$_.ScriptStackTrace;Remainder='FROZEN'})}
    throw
}
