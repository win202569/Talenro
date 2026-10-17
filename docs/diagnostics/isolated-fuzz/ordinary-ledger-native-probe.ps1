# Explicit diagnostic probe preparation. Importing this file starts no workload.
function Assert-OrdinaryLedgerProbeFreshTypes {
    if('IsolatedFuzzProcess' -as [type]){throw 'Unverified preloaded native types refused'}
}
function Open-OrdinaryLedgerProbeFileBinding([string]$Path,[string]$ExpectedSHA256=''){
    if(-not [IO.Path]::IsPathRooted($Path)){throw 'Absolute binding path required'}
    $safe=[IO.Path]::GetFullPath($Path)
    if($safe.Length -lt 4 -or $safe.StartsWith('\\') -or $safe.Substring(2).Contains(':')){throw 'Invalid binding path'}
    $cursor=$safe
    while($cursor -and $cursor.Length -gt 3){
        $item=Get-Item -LiteralPath $cursor -Force -ErrorAction Stop
        if(($item.Attributes -band [IO.FileAttributes]::ReparsePoint) -ne 0){throw 'Reparse binding refused'}
        $cursor=Split-Path -Parent $cursor
    }
    $stream=[IO.File]::Open($safe,[IO.FileMode]::Open,[IO.FileAccess]::Read,[IO.FileShare]::Read)
    try{
        $sha=[Security.Cryptography.SHA256]::Create()
        try{$hash=([BitConverter]::ToString($sha.ComputeHash($stream))).Replace('-','')}finally{$sha.Dispose()}
        $stream.Position=0
        if($ExpectedSHA256 -cne '' -and $hash -cne $ExpectedSHA256){throw 'Locked binding hash mismatch'}
        [pscustomobject]@{Path=$safe;SHA256=$hash;Stream=$stream}
    }catch{$stream.Dispose();throw}
}
function Initialize-OrdinaryLedgerProbe([string]$OwnParsedText){
    Assert-OrdinaryLedgerProbeFreshTypes
    $locks=New-Object 'System.Collections.Generic.List[object]'
    try{
        foreach($name in (Get-OrdinaryLedgerProbeSourceNames)){$locks.Add((Open-OrdinaryLedgerProbeFileBinding (Join-Path $PSScriptRoot $name)))}
        $self=@($locks|Where-Object {[IO.Path]::GetFileName($_.Path) -ceq 'ordinary-ledger-native-probe.ps1'})[0]
        if([IO.File]::ReadAllText($self.Path) -cne $OwnParsedText){throw 'Parsed helper differs from locked source'}
        . (Join-Path $PSScriptRoot 'run.ps1')
        # Dot-sourced dependency functions must survive this initializer scope.
        foreach($name in @('Get-IsolatedFullPath','Test-IsolatedWithin','ConvertTo-RecordedProcessReceipt','Assert-ProcessCompletion','Test-OrdinaryLedgerEvidence')){
            $command=Get-Command $name -CommandType Function -ErrorAction Stop
            Set-Item -Path ('Function:script:'+ $name) -Value $command.ScriptBlock
        }
        $script:OrdinaryProbeLoadedSources=@(foreach($file in $locks){[pscustomobject]@{Path=[IO.Path]::GetFileName($file.Path);SHA256=$file.SHA256}})
        $script:OrdinaryProbeLoadedType='IsolatedFuzzProcess' -as [type]
    }finally{foreach($file in $locks){$file.Stream.Dispose()}}
}
function Assert-OrdinaryLedgerProbeLoadedSources {
    if($null -eq $script:OrdinaryProbeLoadedType -or ('IsolatedFuzzProcess' -as [type]) -ne $script:OrdinaryProbeLoadedType){throw 'Unverified loaded native type'}
    foreach($row in $script:OrdinaryProbeLoadedSources){
        if((Get-OrdinaryLedgerProbeFile (Join-Path $PSScriptRoot $row.Path)).SHA256 -cne $row.SHA256){throw 'Source drift after verified load'}
    }
}
function Open-OrdinaryLedgerProbeBindings($Plan){
    $locks=New-Object 'System.Collections.Generic.List[object]'
    try{
        foreach($row in $script:OrdinaryProbeLoadedSources){$locks.Add((Open-OrdinaryLedgerProbeFileBinding (Join-Path $PSScriptRoot $row.Path) $row.SHA256))}
        $locks.Add((Open-OrdinaryLedgerProbeFileBinding $Plan.Executable $Plan.ExecutableSHA256))
        return ,$locks
    }catch{foreach($file in $locks){$file.Stream.Dispose()};throw}
}
function Get-OrdinaryLedgerProbeScope {
    $workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..'))
    Get-IsolatedFullPath (Join-Path $workspace '.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator')
}
function Get-OrdinaryLedgerProbeFile([string]$Path){
    $safe=Get-IsolatedFullPath $Path
    if(-not(Test-Path -LiteralPath $safe -PathType Leaf) -or [IsolatedFuzzFileInfo]::Links($safe) -ne 1){throw 'Independent probe file required'}
    $item=Get-Item -LiteralPath $safe -ErrorAction Stop
    [pscustomobject]@{Path=$safe;SHA256=(Get-FileHash -LiteralPath $safe -Algorithm SHA256 -ErrorAction Stop).Hash;Length=$item.Length}
}
function Get-OrdinaryLedgerProbeSourceNames {
    @('ordinary-ledger-native-probe.ps1','run.ps1','isolation.ps1','provenance.ps1','process-native.cs','retained-lifetime.cs','compile-stage-trigger.cs','compile-stage-session.cs','ordinary-ledger.cs','ordinary-ledger-session.cs','ordinary-ledger-receipt.ps1')
}
function Get-OrdinaryLedgerProbeRoot([string]$RunRoot,[string]$Case){
    if($Case -cnotin @('Root','Child')){throw 'Fixed probe case required'}
    $scope=Get-IsolatedFullPath (Get-OrdinaryLedgerProbeScope)
    if(-not(Test-Path -LiteralPath $scope -PathType Container)){throw 'Existing owned probe scope required'}
    $root=Get-IsolatedFullPath $RunRoot;$suffix=if($Case -ceq 'Root'){'root'}else{'child'}
    if((Split-Path -Parent $root) -ine $scope -or [IO.Path]::GetFileName($root) -cne ('ledger-native-'+$suffix+'-001')){throw 'Fixed new probe identity required'}
    return $root
}
function Get-OrdinaryLedgerProbePlan([string]$RunRoot,[string]$Case='Root',[string]$Nonce='', [bool]$RequireNew=$true){
    Assert-OrdinaryLedgerProbeLoadedSources
    $root=Get-OrdinaryLedgerProbeRoot $RunRoot $Case
    $control=Get-IsolatedFullPath ($root+'-control')
    if($RequireNew -and ((Test-Path -LiteralPath $root) -or (Test-Path -LiteralPath $control))){throw 'Probe identity already consumed;no replay'}
    if($Nonce -ceq ''){$Nonce=[guid]::NewGuid().ToString('N')}
    if($Nonce -cnotmatch '^[a-f0-9]{32}$'){throw 'Invalid probe nonce'}
    $shell=(Get-OrdinaryLedgerProbeFile (Join-Path $env:SystemRoot 'System32/WindowsPowerShell/v1.0/powershell.exe'))
    $binding=@(foreach($name in (Get-OrdinaryLedgerProbeSourceNames)){
        $file=Get-OrdinaryLedgerProbeFile (Join-Path $PSScriptRoot $name)
        [pscustomobject]@{Path=$name;SHA256=$file.SHA256}
    })
    $common=@'
$ErrorActionPreference='Stop'
$fixtureRoot='__ROOT__'
$fixtureNonce='__NONCE__'
function Write-Stage([string]$Name){
    $row=[pscustomobject]@{Nonce=$fixtureNonce;ProcessId=$PID;UTC=[DateTime]::UtcNow.ToString('o')}
    $target=Join-Path $fixtureRoot $Name
    [IO.File]::WriteAllText(($target+'.tmp'),($row|ConvertTo-Json -Compress))
    [IO.File]::Move(($target+'.tmp'),$target)
}
'@
    $common=$common.Replace('__ROOT__',$root.Replace("'","''")).Replace('__NONCE__',$Nonce)
    $child='';$limit=5
    if($Case -ceq 'Root'){$parent=$common+"`r`nWrite-Stage 'root-exit.json'`r`n"}
    else{
        $limit=20
        $child=$common+@'

Write-Stage 'child-ready.json'
$clock=[Diagnostics.Stopwatch]::StartNew()
while(-not [IO.File]::Exists((Join-Path $fixtureRoot 'release.json'))){
    if($clock.ElapsedMilliseconds -ge 10000){throw 'Child release deadline exceeded'}
    Start-Sleep -Milliseconds 5
}
$release=[IO.File]::ReadAllText((Join-Path $fixtureRoot 'release.json'))|ConvertFrom-Json
if($release.Nonce -cne $fixtureNonce){throw 'Wrong release nonce'}
Start-Sleep -Milliseconds 150
Write-Stage 'child-exit.json'
'@
        $encodedChild=[Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($child))
        $parent=$common+@'

Start-Process -FilePath '__SHELL__' -ArgumentList '-NoProfile -NonInteractive -EncodedCommand __CHILD__' -WindowStyle Hidden
$clock=[Diagnostics.Stopwatch]::StartNew()
while(-not [IO.File]::Exists((Join-Path $fixtureRoot 'child-ready.json'))){
    if($clock.ElapsedMilliseconds -ge 10000){throw 'Parent readiness deadline exceeded'}
    Start-Sleep -Milliseconds 5
}
$ready=[IO.File]::ReadAllText((Join-Path $fixtureRoot 'child-ready.json'))|ConvertFrom-Json
if($ready.Nonce -cne $fixtureNonce -or $ready.ProcessId -le 0 -or $ready.ProcessId -eq $PID){throw 'Wrong child readiness identity'}
Write-Stage 'release.json'
Write-Stage 'parent-exit.json'
'@
        $parent=$parent.Replace('__SHELL__',$shell.Path.Replace("'","''")).Replace('__CHILD__',$encodedChild)
    }
    [pscustomobject]@{Schema=1;Policy='ordinary-ledger-native-probe-v1';Case=$Case;RunRoot=$root;ControlRoot=$control;Nonce=$Nonce;Executable=$shell.Path;ExecutableSHA256=$shell.SHA256;SourceBinding=$binding;LimitSeconds=$limit;ParentScript=$parent;ChildScript=$child;Arguments=@('-NoProfile','-NonInteractive','-EncodedCommand',[Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($parent)));Environment=@{};Acceptance=$false;NoRetry=$true;Coverage='sampled'}
}
function Get-OrdinaryLedgerProbePlanHash($Plan){
    $bytes=[Text.Encoding]::UTF8.GetBytes(($Plan|ConvertTo-Json -Depth 12 -Compress -WarningAction Stop))
    $sha=[Security.Cryptography.SHA256]::Create()
    try{([BitConverter]::ToString($sha.ComputeHash($bytes))).Replace('-','')}finally{$sha.Dispose()}
}
function Assert-OrdinaryLedgerProbePlan($Plan,[bool]$RequireNew=$true){
    $fixed=Get-OrdinaryLedgerProbePlan $Plan.RunRoot $Plan.Case $Plan.Nonce $RequireNew
    if(($Plan|ConvertTo-Json -Depth 12 -Compress -WarningAction Stop) -cne ($fixed|ConvertTo-Json -Depth 12 -Compress -WarningAction Stop)){throw 'Probe plan/source/executable drift'}
}
function Write-OrdinaryLedgerProbeJSON([string]$Path,$Value){
    $safe=Get-IsolatedFullPath $Path
    $bytes=[Text.Encoding]::UTF8.GetBytes(($Value|ConvertTo-Json -Depth 24 -WarningAction Stop))
    $stream=[IO.File]::Open($safe,[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::Read)
    try{$stream.Write($bytes,0,$bytes.Length)}finally{$stream.Dispose()}
}
function Read-OrdinaryLedgerProbeJSON([string]$Path,[long]$Limit=1048576){
    $file=Get-OrdinaryLedgerProbeFile $Path
    if($file.Length -gt $Limit){throw 'Probe JSON exceeds bound'}
    $value=[IO.File]::ReadAllText($file.Path)|ConvertFrom-Json
    if($null -eq $value -or $value -is [ValueType] -or $value -is [string] -or $value -is [Collections.IList]){throw 'Probe JSON object required'}
    return $value
}
function Get-OrdinaryLedgerProbeUtc($Value){
    if($Value -is [DateTime]){if($Value.Kind -ne [DateTimeKind]::Utc){throw 'UTC marker required'};return [DateTimeOffset]$Value}
    if($Value -isnot [string]){throw 'UTC marker type invalid'}
    $time=[DateTimeOffset]::MinValue
    if(-not [DateTimeOffset]::TryParse($Value,[Globalization.CultureInfo]::InvariantCulture,[Globalization.DateTimeStyles]::None,[ref]$time) -or $time.Offset -ne [TimeSpan]::Zero){throw 'UTC marker invalid'}
    return $time
}
function Get-OrdinaryLedgerProbeEvidence($Plan,$Receipt){
    $errors=@();$core=$false;$stages=$false;$childEntry=$null;$identity=$null
    try{
        if($null -eq $Receipt){throw 'Missing native receipt'}
        $evaluation=Test-OrdinaryLedgerEvidence $Receipt
        if(-not $evaluation.EvidenceComplete){throw ('Ledger evidence incomplete: '+($evaluation.Errors -join ';'))}
        if($Receipt.ExitCode.GetType().FullName -notin @('System.Int32','System.Int64','System.UInt32','System.UInt64') -or $Receipt.ExitCode -ne 0){throw 'Root exit incomplete'}
        if($Receipt.Executable -cne $Plan.Executable -or $Receipt.WorkingDirectory -cne $Plan.RunRoot -or ($Receipt.Arguments|ConvertTo-Json -Compress) -cne ($Plan.Arguments|ConvertTo-Json -Compress) -or ($Receipt.Environment|ConvertTo-Json -Compress) -cne '{}'){throw 'Receipt invocation binding mismatch'}
        $identity=$Receipt.RunIdentity;$core=$true
    }catch{$errors+=('core: '+$_.Exception.Message)}
    try{
        $begin=Get-OrdinaryLedgerProbeUtc $Receipt.StartUTC;$end=Get-OrdinaryLedgerProbeUtc $Receipt.EndUTC
        if($end -lt $begin){throw 'Reversed receipt interval'}
        $names=if($Plan.Case -ceq 'Root'){@('root-exit.json')}else{@('child-ready.json','release.json','parent-exit.json','child-exit.json')}
        $rows=@{}
        foreach($name in $names){
            $row=Read-OrdinaryLedgerProbeJSON (Join-Path $Plan.RunRoot $name) 4096
            if($row.Nonce -isnot [string] -or $row.Nonce -cne $Plan.Nonce -or $row.ProcessId.GetType().FullName -notin @('System.Int32','System.Int64','System.UInt32','System.UInt64') -or $row.ProcessId -le 0 -or $row.ProcessId -gt [UInt32]::MaxValue){throw 'Marker identity invalid'}
            $time=Get-OrdinaryLedgerProbeUtc $row.UTC
            if($time -lt $begin -or $time -gt $end){throw 'Marker outside receipt interval'}
            $rows[$name]=$row
        }
        if($Plan.Case -ceq 'Root'){
            if($rows['root-exit.json'].ProcessId -ne $Receipt.RootId){throw 'Root marker mismatch'}
        }else{
            $ready=$rows['child-ready.json'];$release=$rows['release.json'];$parent=$rows['parent-exit.json'];$child=$rows['child-exit.json']
            if($ready.ProcessId -eq $Receipt.RootId -or $ready.ProcessId -ne $child.ProcessId -or $release.ProcessId -ne $Receipt.RootId -or $parent.ProcessId -ne $Receipt.RootId){throw 'Child marker mismatch'}
            if((Get-OrdinaryLedgerProbeUtc $ready.UTC) -gt (Get-OrdinaryLedgerProbeUtc $release.UTC) -or (Get-OrdinaryLedgerProbeUtc $release.UTC) -gt (Get-OrdinaryLedgerProbeUtc $parent.UTC) -or (Get-OrdinaryLedgerProbeUtc $parent.UTC) -gt (Get-OrdinaryLedgerProbeUtc $child.UTC)){throw 'Marker ordering invalid'}
            $matches=@($Receipt.MemberIdentityLedger.Entries|Where-Object {$_.Pid -eq $ready.ProcessId})
            if(-not $core -or $matches.Count -ne 1 -or $matches[0].Initial.State -cne 'running' -or -not $matches[0].Final.SameHandleVerified -or $matches[0].Final.ExitCode -ne 0){throw 'Controlled child same-run exit evidence missing'}
            $childEntry=$matches[0].EntryId
        }
        $stages=$true
    }catch{$errors+=('markers: '+$_.Exception.Message)}
    [pscustomobject]@{Schema=1;Case=$Plan.Case;RunRoot=$Plan.RunRoot;RunIdentity=$identity;Coverage='sampled';ObservationClaim='same-run-owned-handle-natural-completion';ChildOutlivedRootVerified=$false;NoRetry=$true;Acceptance=$false;OriginalTestPassed=$false;EvidenceComplete=$core;StageEvidenceComplete=$stages;ChildEntryId=$childEntry;Errors=[string[]]$errors;ObservationPassed=($core -and $stages);FinalizationRequired=$true;Finalized=$false}
}
function Get-OrdinaryLedgerProbeManifest([string]$RunRoot){
    function Walk-Probe([string]$Directory){
        foreach($entry in @(Get-ChildItem -LiteralPath $Directory -Force -ErrorAction Stop)){
            $safe=Get-IsolatedFullPath $entry.FullName
            if($entry.PSIsContainer){Walk-Probe $safe}else{
                $relative=$safe.Substring($RunRoot.Length+1).Replace('\','/')
                if($relative -cin @('evidence-manifest.json','completed.json')){continue}
                $file=Get-OrdinaryLedgerProbeFile $safe
                [pscustomobject]@{Path=$relative;SHA256=$file.SHA256;Length=$file.Length}
            }
        }
    }
    @(Walk-Probe $RunRoot)|Sort-Object Path
}
function Get-OrdinaryLedgerProbeSaved([string]$RunRoot){
    $leaf=[IO.Path]::GetFileName($RunRoot)
    $case=if($leaf -ceq 'ledger-native-root-001'){'Root'}elseif($leaf -ceq 'ledger-native-child-001'){'Child'}else{throw 'Fixed saved identity required'}
    $root=Get-OrdinaryLedgerProbeRoot $RunRoot $case
    if(Test-Path -LiteralPath (Join-Path $root 'stopped.json')){throw 'Stopped probe is not complete'}
    $plan=Read-OrdinaryLedgerProbeJSON (Join-Path $root 'plan.json');Assert-OrdinaryLedgerProbePlan $plan $false
    if($plan.RunRoot -cne $root -or $plan.Case -cne $case){throw 'Saved plan root mismatch'}
    $planHash=Get-OrdinaryLedgerProbePlanHash $plan
    $consumed=Read-OrdinaryLedgerProbeJSON (Join-Path $plan.ControlRoot 'consumed.json')
    $expectedConsumed=[pscustomobject]@{Schema=1;Case=$case;PlanSHA256=$planHash;IdentityConsumed=$true;Acceptance=$false;NoRetry=$true}
    if(($consumed|ConvertTo-Json -Compress) -cne ($expectedConsumed|ConvertTo-Json -Compress)){throw 'Consumed identity mismatch'}
    $manifestPath=Join-Path $root 'evidence-manifest.json'
    $manifest=Read-OrdinaryLedgerProbeJSON $manifestPath
    $expectedManifest=[pscustomobject]@{Schema=1;Acceptance=$false;NoRetry=$true;Files=@(Get-OrdinaryLedgerProbeManifest $root)}
    if(($manifest|ConvertTo-Json -Depth 8 -Compress) -cne ($expectedManifest|ConvertTo-Json -Depth 8 -Compress)){throw 'Saved artifact manifest mismatch'}
    $completed=Read-OrdinaryLedgerProbeJSON (Join-Path $root 'completed.json')
    $expectedCompleted=[pscustomobject]@{Schema=1;Finalized=$true;ObservationPassed=$true;Acceptance=$false;NoRetry=$true;PlanSHA256=$planHash;ManifestSHA256=(Get-OrdinaryLedgerProbeFile $manifestPath).SHA256}
    if(($completed|ConvertTo-Json -Compress) -cne ($expectedCompleted|ConvertTo-Json -Compress)){throw 'Durable completion marker mismatch'}
    foreach($path in @('process/stdout.txt','process/stderr.txt')){Get-OrdinaryLedgerProbeFile (Join-Path $root $path)|Out-Null}
    $receipt=Read-OrdinaryLedgerProbeJSON (Join-Path $root 'process/receipt.json')
    $result=Get-OrdinaryLedgerProbeEvidence $plan $receipt
    $savedResult=Read-OrdinaryLedgerProbeJSON (Join-Path $root 'result.json')
    if(-not $result.ObservationPassed -or ($savedResult|ConvertTo-Json -Depth 8 -Compress) -cne ($result|ConvertTo-Json -Depth 8 -Compress)){throw 'Saved result disagrees with raw evidence'}
    $result.Finalized=$true;return $result
}
if(-not (Get-Command Invoke-OrdinaryLedgerProbeNative -CommandType Function -ErrorAction SilentlyContinue)){
function Invoke-OrdinaryLedgerProbeNative($Plan,[string]$Stdout,[string]$Stderr){
    if(-not [Environment]::Is64BitProcess -or $PSVersionTable.PSEdition -cne 'Desktop' -or $PSVersionTable.PSVersion.Major -ne 5){throw 'Fresh64-bit Windows PowerShell5.1 control host required'}
    $native=[IsolatedFuzzProcess]::RunWithOrdinaryIdentityLedger($Plan.Executable,[string[]]$Plan.Arguments,@{},$Plan.RunRoot,$Stdout,$Stderr,$Plan.LimitSeconds)
    $receipt=ConvertTo-RecordedProcessReceipt $native $Stdout $Stderr $Plan.Executable $Plan.Arguments @{}
    $receipt|Add-Member WorkingDirectory $Plan.RunRoot
    return $receipt
}
}
function Invoke-OrdinaryLedgerProbe($Plan,[switch]$RuntimeApproved,[string]$ApprovedPlanSHA256=''){
    # These caller safeguards do not grant approval or attest security readiness.
    # Stage B requires separately reviewed current safety evidence before calling.
    if(-not $RuntimeApproved){throw 'Stage B runtime approval required'}
    $bindings=Open-OrdinaryLedgerProbeBindings $Plan
    try{
    Assert-OrdinaryLedgerProbePlan $Plan $true
    $planHash=Get-OrdinaryLedgerProbePlanHash $Plan
    if($ApprovedPlanSHA256 -cnotmatch '^[A-F0-9]{64}$' -or $ApprovedPlanSHA256 -cne $planHash){throw 'Exact approved plan hash required'}
    if($Plan.Case -ceq 'Child'){
        $prior=Get-OrdinaryLedgerProbeSaved (Join-Path (Get-OrdinaryLedgerProbeScope) 'ledger-native-root-001')
        if(-not $prior.ObservationPassed -or -not $prior.Finalized){throw 'Root probe must complete before child'}
    }
    New-Item -ItemType Directory -Path $Plan.ControlRoot -ErrorAction Stop|Out-Null
    try{
        Write-OrdinaryLedgerProbeJSON (Join-Path $Plan.ControlRoot 'consumed.json') ([pscustomobject]@{Schema=1;Case=$Plan.Case;PlanSHA256=$planHash;IdentityConsumed=$true;Acceptance=$false;NoRetry=$true})
        New-Item -ItemType Directory -Path $Plan.RunRoot -ErrorAction Stop|Out-Null
        Write-OrdinaryLedgerProbeJSON (Join-Path $Plan.RunRoot 'plan.json') $Plan
        $process=Join-Path $Plan.RunRoot 'process';New-Item -ItemType Directory -Path $process -ErrorAction Stop|Out-Null
        $receipt=Invoke-OrdinaryLedgerProbeNative $Plan (Join-Path $process 'stdout.txt') (Join-Path $process 'stderr.txt')
        Write-OrdinaryLedgerProbeJSON (Join-Path $process 'receipt.json') $receipt
        $result=Get-OrdinaryLedgerProbeEvidence $Plan $receipt
        Write-OrdinaryLedgerProbeJSON (Join-Path $Plan.RunRoot 'result.json') $result
        if(-not $result.ObservationPassed){throw ('Probe evidence incomplete: '+($result.Errors -join ';'))}
        $manifestPath=Join-Path $Plan.RunRoot 'evidence-manifest.json'
        Write-OrdinaryLedgerProbeJSON $manifestPath ([pscustomobject]@{Schema=1;Acceptance=$false;NoRetry=$true;Files=@(Get-OrdinaryLedgerProbeManifest $Plan.RunRoot)})
        Write-OrdinaryLedgerProbeJSON (Join-Path $Plan.RunRoot 'completed.json') ([pscustomobject]@{Schema=1;Finalized=$true;ObservationPassed=$true;Acceptance=$false;NoRetry=$true;PlanSHA256=$planHash;ManifestSHA256=(Get-OrdinaryLedgerProbeFile $manifestPath).SHA256})
        Get-OrdinaryLedgerProbeSaved $Plan.RunRoot
    }catch{
        $primary=$_
        $stopRoot=if(Test-Path -LiteralPath $Plan.RunRoot -PathType Container){$Plan.RunRoot}else{$Plan.ControlRoot}
        try{Write-OrdinaryLedgerProbeJSON (Join-Path $stopRoot 'stopped.json') ([pscustomobject]@{Error=$primary.Exception.Message;CleanupVerified=$false;Acceptance=$false;NoRetry=$true})}catch{}
        throw $primary
    }
    }finally{foreach($file in $bindings){$file.Stream.Dispose()}}
}
Initialize-OrdinaryLedgerProbe $MyInvocation.MyCommand.ScriptBlock.Ast.Extent.Text
