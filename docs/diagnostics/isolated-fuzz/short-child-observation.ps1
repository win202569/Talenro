. (Join-Path $PSScriptRoot 'run.ps1')
function Get-ShortChildObservationRoot([string]$RunRoot) {
    $workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..'))
    $scope=Join-Path $workspace '.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator'
    $root=Get-IsolatedFullPath $RunRoot
    if(-not (Test-IsolatedWithin $root $scope) -or [IO.Path]::GetFileName($root) -cnotmatch '^short-child-observe-[0-9]{3}$'){throw 'Owned short-child identity required'}
    return $root
}
function Get-ShortChildObservationPlan([string]$RunRoot) {
    $root=Get-ShortChildObservationRoot $RunRoot
    if(Test-Path -LiteralPath $root){throw 'New short-child identity required; no replay'}
    $shell=Join-Path ([Environment]::GetFolderPath('System')) 'WindowsPowerShell/v1.0/powershell.exe'
    if(-not (Test-Path -LiteralPath $shell -PathType Leaf)){throw 'Existing Windows PowerShell required'}
    $nonce=[guid]::NewGuid().ToString('N')
    $common=@'
$ErrorActionPreference='Stop'
$fixtureRoot='__ROOT__'
$fixtureNonce='__NONCE__'
function Write-Stage([string]$Name) {
    $text=[pscustomobject]@{Nonce=$fixtureNonce;ProcessId=$PID;UTC=[DateTime]::UtcNow.ToString('o')}|ConvertTo-Json -Compress
    $target=Join-Path $fixtureRoot $Name
    [IO.File]::WriteAllText(($target+'.tmp'),$text)
    [IO.File]::Move(($target+'.tmp'),$target)
}
'@
    $common=$common.Replace('__ROOT__',$root.Replace("'","''")).Replace('__NONCE__',$nonce)
    $child=$common+@'

Write-Stage 'child-ready.json'
$clock=[Diagnostics.Stopwatch]::StartNew()
while(-not [IO.File]::Exists((Join-Path $fixtureRoot 'release.json'))) {
    if($clock.ElapsedMilliseconds -ge 10000){throw 'Child release deadline exceeded'}
    Start-Sleep -Milliseconds 5
}
$release=[IO.File]::ReadAllText((Join-Path $fixtureRoot 'release.json'))|ConvertFrom-Json
if($release.Nonce -cne $fixtureNonce){throw 'Wrong release nonce'}
Start-Sleep -Milliseconds 150
Write-Stage 'child-exit.json'
'@
    $encoded=[Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($child))
    $parent=$common+@'

Start-Process -FilePath '__SHELL__' -ArgumentList '-NoProfile -NonInteractive -EncodedCommand __CHILD__' -WindowStyle Hidden
$clock=[Diagnostics.Stopwatch]::StartNew()
while(-not [IO.File]::Exists((Join-Path $fixtureRoot 'child-ready.json'))) {
    if($clock.ElapsedMilliseconds -ge 10000){throw 'Parent readiness deadline exceeded'}
    Start-Sleep -Milliseconds 5
}
$ready=[IO.File]::ReadAllText((Join-Path $fixtureRoot 'child-ready.json'))|ConvertFrom-Json
if($ready.Nonce -cne $fixtureNonce -or $ready.ProcessId -le 0 -or $ready.ProcessId -eq $PID){throw 'Wrong child readiness identity'}
Write-Stage 'release.json'
Write-Stage 'parent-exit.json'
'@
    $parent=$parent.Replace('__SHELL__',$shell.Replace("'","''")).Replace('__CHILD__',$encoded)
    [pscustomobject]@{RunRoot=$root;Nonce=$nonce;Executable=$shell;LimitSeconds=20;ChildDelayMs=150;ReadyLimitMs=10000;ParentScript=$parent;ChildScript=$child;Arguments=@('-NoProfile','-NonInteractive','-EncodedCommand',[Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($parent)))}
}
function Get-ShortChildObservationEvidence([object]$Plan,[object]$Receipt) {
    $root=Get-ShortChildObservationRoot $Plan.RunRoot
    $stages=@{};$stageComplete=$false;$stageError=$null
    try {
        foreach($name in @('child-ready.json','release.json','parent-exit.json','child-exit.json')) {
            $path=Get-IsolatedFullPath (Join-Path $root $name)
            if(-not (Test-Path -LiteralPath $path -PathType Leaf) -or (Get-Item -LiteralPath $path).Length -gt 4096 -or [IsolatedFuzzFileInfo]::Links($path) -ne 1){throw ('Missing/unsafe stage: '+$name)}
            $row=[IO.File]::ReadAllText($path)|ConvertFrom-Json
            if($row.Nonce -cne $Plan.Nonce -or $row.ProcessId -le 0){throw ('Wrong stage identity: '+$name)}
            $time=[DateTimeOffset]$row.UTC
            if($time -lt [DateTimeOffset]$Receipt.StartUTC -or $time -gt [DateTimeOffset]$Receipt.EndUTC){throw 'Stage outside receipt interval'}
            $stages[$name]=$row
        }
        $ready=$stages['child-ready.json'];$release=$stages['release.json'];$parent=$stages['parent-exit.json'];$child=$stages['child-exit.json']
        if($ready.ProcessId -eq $Receipt.RootId -or $ready.ProcessId -ne $child.ProcessId -or $release.ProcessId -ne $Receipt.RootId -or $parent.ProcessId -ne $Receipt.RootId){throw 'Stage process identities disagree'}
        if([DateTimeOffset]$ready.UTC -gt [DateTimeOffset]$release.UTC -or [DateTimeOffset]$release.UTC -gt [DateTimeOffset]$parent.UTC -or [DateTimeOffset]$release.UTC -gt [DateTimeOffset]$child.UTC){throw 'Stage order invalid'}
        $stageComplete=$true
    } catch {$stageError=$_.Exception.Message}
    $matched=$false
    if($stageComplete){$rows=@($Receipt.MemberEvidence|Where-Object {$_.Id -eq $stages['child-ready.json'].ProcessId});$matched=$rows.Count -eq 1 -and $rows[0].Owned -is [bool] -and $rows[0].Owned -and $rows[0].QueryError -eq 0 -and $rows[0].ExitCodeRead}
    $queries=$false
    try {
        $a=$Receipt.InitialAccountingQuery;$m=@($Receipt.InitialMemberQueries)
        $queries=$null -ne $a -and $a.Succeeded -and $a.Error -eq 0 -and $a.ValuesRead -and $a.Frequency -gt 0 -and $a.EndTimestamp -ge $a.StartTimestamp -and $m.Count -gt 0 -and $m.Count -le 8
        $last=$a.EndTimestamp
        foreach($row in $m){if(-not $row.HeaderRead -or $row.Frequency -ne $a.Frequency -or $row.StartTimestamp -lt $last -or $row.EndTimestamp -lt $row.StartTimestamp -or (-not $row.Succeeded -and $row.Error -ne 234)){$queries=$false};$last=$row.EndTimestamp}
        if($m.Count -eq 0 -or -not $m[-1].Succeeded -or $m[-1].Error -ne 0){$queries=$false}
    } catch {$queries=$false}
    $completion=$false;$completionError=$null
    try{Assert-ProcessCompletion $Receipt;if($Receipt.ExitCode -ne 0){throw 'Parent exit code nonzero'};$completion=$true}catch{$completionError=$_.Exception.Message}
    [pscustomobject]@{RunRoot=$root;DiagnosticOnly=$true;Acceptance=$false;OriginalTestPassed=$false;StageEvidenceComplete=$stageComplete;StageError=$stageError;Stages=$stages;ChildMatchedOwnedEvidence=$matched;QueryEvidenceComplete=$queries;CompletionPassed=$completion;CompletionError=$completionError;ObservationPassed=($stageComplete -and $matched -and $queries -and $completion)}
}
function Invoke-ShortChildObservation([string]$RunRoot) {
    $plan=Get-ShortChildObservationPlan $RunRoot
    New-Item -ItemType Directory -Path $plan.RunRoot -ErrorAction Stop|Out-Null
    $plan|ConvertTo-Json -Depth 4|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'plan.json') -ErrorAction Stop
    try {
        $receipt=Invoke-RecordedProcess $plan.Executable $plan.Arguments @{} (Join-Path $plan.RunRoot 'process') $plan.LimitSeconds
        $result=Get-ShortChildObservationEvidence $plan $receipt
        $result|ConvertTo-Json -Depth 8|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'result.json') -ErrorAction Stop
        if(-not $result.ObservationPassed){throw 'Short-child observation incomplete or completion refused; stop without retry'}
        return $result
    } catch {
        [pscustomobject]@{UTC=[DateTime]::UtcNow.ToString('o');Error=$_.Exception.Message;NoRetry=$true;Acceptance=$false}|ConvertTo-Json|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'stopped.json') -ErrorAction Stop
        throw
    }
}
