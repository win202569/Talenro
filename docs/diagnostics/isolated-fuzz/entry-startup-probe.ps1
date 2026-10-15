. (Join-Path $PSScriptRoot 'run.ps1')
function Get-EntryStartupProbePlan([string]$RunRoot) {
    $workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..'))
    $scope=Join-Path $workspace '.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator'
    $root=Get-IsolatedFullPath $RunRoot
    if(-not (Test-IsolatedWithin $root $scope) -or [IO.Path]::GetFileName($root) -cnotmatch '^entry-startup-probe-[0-9]{3}$' -or (Test-Path -LiteralPath $root)){throw 'New owned entry diagnostic identity required; no retry'}
    $shell=Join-Path ([Environment]::GetFolderPath('System')) 'WindowsPowerShell/v1.0/powershell.exe'
    if(-not (Test-Path -LiteralPath $shell -PathType Leaf)){throw 'Existing Windows PowerShell fixture runtime missing'}
    $script="[Console]::Out.WriteLine('TALENRO_ENTRY_PROBE_ENTERED');[Console]::Out.Flush();[IO.File]::WriteAllText('"+(Join-Path $root 'script-entry.utc').Replace("'","''")+"',[DateTime]::UtcNow.ToString('o'))"
    [pscustomobject]@{RunRoot=$root;Executable=$shell;LimitSeconds=2;EntryScript=$script;Arguments=@('-NoProfile','-EncodedCommand',[Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($script)))}
}
function Get-EntryStartupProbeEvidence([string]$RunRoot,[object]$Receipt) {
    $workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..'))
    $scope=Join-Path $workspace '.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator'
    $root=Get-IsolatedFullPath $RunRoot
    if(-not (Test-IsolatedWithin $root $scope)){throw 'Entry evidence outside owned scope'}
    $stdout=Get-IsolatedFullPath (Join-Path $root 'process/stdout.txt')
    $marker=Get-IsolatedFullPath (Join-Path $root 'script-entry.utc')
    $entryObserved=$false;$markerObserved=$false;$markerText=$null
    if(Test-Path -LiteralPath $stdout -PathType Leaf) {
        if([IsolatedFuzzFileInfo]::Links($stdout) -ne 1){throw 'Linked output refused'}
        $entryObserved=@([IO.File]::ReadAllLines($stdout)|Where-Object {$_ -ceq 'TALENRO_ENTRY_PROBE_ENTERED'}).Count -gt 0
    }
    if(Test-Path -LiteralPath $marker -PathType Leaf) {
        if([IsolatedFuzzFileInfo]::Links($marker) -ne 1){throw 'Linked marker refused'}
        $markerObserved=$true;$markerText=[IO.File]::ReadAllText($marker)
    }
    [pscustomobject]@{Acceptance=$false;OriginalTestPassed=$false;DiagnosticOnly=$true;StdoutEntryObserved=$entryObserved;FileMarkerObserved=$markerObserved;FileMarkerText=$markerText;Receipt=$Receipt}
}
function Invoke-EntryStartupProbe([string]$RunRoot) {
    $plan=Get-EntryStartupProbePlan $RunRoot
    New-Item -ItemType Directory -Path $plan.RunRoot -ErrorAction Stop|Out-Null
    $plan|ConvertTo-Json -Depth 4|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'plan.json')
    try {
        $receipt=Invoke-RecordedProcess $plan.Executable $plan.Arguments @{} (Join-Path $plan.RunRoot 'process') $plan.LimitSeconds
        $result=Get-EntryStartupProbeEvidence $plan.RunRoot $receipt
        $result|ConvertTo-Json -Depth 8|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'result.json')
        if(-not $receipt.CleanupVerified -or $receipt.ActiveAfterCleanup -ne 0 -or $receipt.MembersAfterCleanup.Count -ne 0){throw 'Entry probe owned-job cleanup unverified; stop without retry'}
        return [pscustomobject]@{RunRoot=$plan.RunRoot;Acceptance=$false;OriginalTestPassed=$false;StdoutEntryObserved=$result.StdoutEntryObserved;FileMarkerObserved=$result.FileMarkerObserved;TimedOut=$receipt.TimedOut;ExitCode=$receipt.ExitCode;CleanupVerified=$receipt.CleanupVerified}
    } catch {
        [pscustomobject]@{UTC=[DateTime]::UtcNow.ToString('o');Error=$_.Exception.Message;Acceptance=$false;OriginalTestPassed=$false;DiagnosticOnly=$true}|ConvertTo-Json|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'stopped.json')
        throw
    }
}
