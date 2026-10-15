. (Join-Path $PSScriptRoot 'run.ps1')
function Get-ChildReadinessRoot([string]$RunRoot) {
    $workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..'))
    $scope=Join-Path $workspace '.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator'
    $root=Get-IsolatedFullPath $RunRoot
    if(-not (Test-IsolatedWithin $root $scope) -or [IO.Path]::GetFileName($root) -cnotmatch '^child-readiness-[0-9]{3}$'){throw 'Owned child-readiness identity required'}
    return $root
}
function Get-ChildReadinessPlan([string]$RunRoot) {
    $root=Get-ChildReadinessRoot $RunRoot
    if(Test-Path -LiteralPath $root){throw 'New readiness identity required; no retry'}
    $shell=Join-Path ([Environment]::GetFolderPath('System')) 'WindowsPowerShell/v1.0/powershell.exe'
    if(-not (Test-Path -LiteralPath $shell -PathType Leaf)){throw 'Existing fixture runtime missing'}
    $nonce=[guid]::NewGuid().ToString('N')
    $path=Join-Path $root 'child-ready.json'
    $child=@'
$readyPayload=[pscustomobject]@{Nonce='__NONCE__';ProcessId=$PID;UTC=[DateTime]::UtcNow.ToString('o')}
$readyStream=[IO.File]::Open('__PATH__',[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::Read)
try {
    $readyWriter=New-Object IO.StreamWriter($readyStream)
    try {$readyWriter.WriteLine(($readyPayload|ConvertTo-Json -Compress));$readyWriter.Flush()}finally{$readyWriter.Dispose()}
} finally {$readyStream.Dispose()}
[Threading.Thread]::Sleep(20000)
'@
    $child=$child.Replace('__NONCE__',$nonce).Replace('__PATH__',$path.Replace("'","''"))
    $encoded=[Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($child))
    $parent="Start-Process -FilePath '"+$shell.Replace("'","''")+"' -ArgumentList '-NoProfile -EncodedCommand "+$encoded+"' -WindowStyle Hidden;[Threading.Thread]::Sleep(20000)"
    [pscustomobject]@{RunRoot=$root;Nonce=$nonce;ReadinessPath=$path;Executable=$shell;LimitSeconds=2;ChildScript=$child;ParentScript=$parent;Arguments=@('-NoProfile','-EncodedCommand',[Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($parent)))}
}
function Get-ChildReadinessEvidence([object]$Plan,[object]$Receipt) {
    $root=Get-ChildReadinessRoot $Plan.RunRoot
    $path=Get-IsolatedFullPath (Join-Path $root 'child-ready.json')
    if($Plan.LimitSeconds -ne 2 -or $Plan.Nonce -cnotmatch '^[0-9a-f]{32}$' -or $path -ine $Plan.ReadinessPath){throw 'Invalid readiness plan'}
    $observed=Test-Path -LiteralPath $path -PathType Leaf
    $ready=$false;$clean=$false;$reason='Readiness record absent'
    try {
        $clean=$Receipt.CleanupVerified -is [bool] -and $Receipt.CleanupVerified -and $Receipt.ActiveAfterCleanup -eq 0 -and $null -ne $Receipt.MembersAfterCleanup -and @($Receipt.MembersAfterCleanup).Count -eq 0
    } catch {$clean=$false}
    if($observed) {
        if([IsolatedFuzzFileInfo]::Links($path) -ne 1 -or (Get-Item -LiteralPath $path).Length -gt 4096){throw 'Linked or oversized readiness record refused'}
        try {
            $record=[IO.File]::ReadAllText($path)|ConvertFrom-Json
            $childId=[UInt64]0
            if($record.Nonce -cne $Plan.Nonce -or -not [UInt64]::TryParse([string]$record.ProcessId,[ref]$childId) -or $childId -eq 0 -or $childId -eq $Receipt.RootId){throw 'Wrong nonce or child ID'}
            $created=[DateTimeOffset]$record.UTC;$start=[DateTimeOffset]$Receipt.StartUTC
            if($created -lt $start -or $created -gt $start.AddSeconds(2)){throw 'Readiness timestamp outside original budget'}
            if($Receipt.MemberEvidenceTruncated -isnot [bool] -or $Receipt.MemberEvidenceTruncated){throw 'Incomplete member evidence'}
            $members=@($Receipt.MemberEvidence|Where-Object {$null -ne $_ -and $_.Id -eq $childId})
            if($members.Count -ne 1){throw 'Unique intended child not observed'}
            $member=$members[0]
            if($member.Owned -isnot [bool] -or -not $member.Owned -or $member.State -cne 'running' -or $member.QueryError -ne 0 -or $member.ExitCodeRead -isnot [bool] -or -not $member.ExitCodeRead -or $member.ExitCode -ne 259 -or $created -gt [DateTimeOffset]$member.CoreObservedUTC){throw 'Intended child lacks owned-running evidence after readiness timestamp'}
            $ready=$true;$reason='Unique nonce/child ID matched owned-running evidence'
        } catch {$reason=$_.Exception.Message}
    }
    $timedOut=$Receipt.TimedOut -is [bool] -and $Receipt.TimedOut
    [pscustomobject]@{RunRoot=$root;DiagnosticOnly=$true;Acceptance=$false;OriginalTestPassed=$false;ReadinessRecordObserved=$observed;ChildReady=$ready;ReadinessReason=$reason;TimedOut=$timedOut;CleanupVerified=$clean;FixturePassed=($ready -and $timedOut -and $clean)}
}
function Invoke-ChildReadinessFixture([string]$RunRoot) {
    $plan=Get-ChildReadinessPlan $RunRoot
    New-Item -ItemType Directory -Path $plan.RunRoot -ErrorAction Stop|Out-Null
    $plan|ConvertTo-Json -Depth 4|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'plan.json')
    try {
        $receipt=Invoke-RecordedProcess $plan.Executable $plan.Arguments @{} (Join-Path $plan.RunRoot 'process') $plan.LimitSeconds
        $evidence=Get-ChildReadinessEvidence $plan $receipt
        [pscustomobject]@{Evidence=$evidence;Receipt=$receipt}|ConvertTo-Json -Depth 8|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'result.json')
        if(-not $evidence.CleanupVerified){throw 'Owned-job cleanup not verified; stop without retry'}
        return $evidence
    } catch {
        [pscustomobject]@{UTC=[DateTime]::UtcNow.ToString('o');Error=$_.Exception.Message;DiagnosticOnly=$true;Acceptance=$false;OriginalTestPassed=$false}|ConvertTo-Json|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'stopped.json')
        throw
    }
}
