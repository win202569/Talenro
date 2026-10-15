. (Join-Path $PSScriptRoot 'run.ps1')
function Get-TreeTimeoutProbePlan([string]$RunRoot,[bool]$SynchronousStdout=$false) {
    $workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..'))
    $scope=Join-Path $workspace '.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator'
    $root=Get-IsolatedFullPath $RunRoot
    $identityPattern=if($SynchronousStdout){'^full-parent-probe-[0-9]{3}$'}else{'^tree-timeout-probe-[0-9]{3}$'}
    if(-not (Test-IsolatedWithin $root $scope) -or [IO.Path]::GetFileName($root) -cnotmatch $identityPattern -or (Test-Path -LiteralPath $root)){throw 'New owned diagnostic identity required; no retry'}
    $shell=Join-Path ([Environment]::GetFolderPath('System')) 'WindowsPowerShell/v1.0/powershell.exe'
    if(-not (Test-Path -LiteralPath $shell -PathType Leaf)){throw 'Existing Windows PowerShell fixture runtime missing'}
    $child="[IO.File]::WriteAllText('"+(Join-Path $root 'child-entry.utc').Replace("'","''")+"',[DateTime]::UtcNow.ToString('o'));Start-Sleep -Seconds 20"
    $childEncoded=[Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($child))
    $parent=@'
__STDOUT_ENTRY__
[IO.File]::WriteAllText('__ENTRY__',[DateTime]::UtcNow.ToString('o'))
[IO.File]::WriteAllText('__BEFORE__',[DateTime]::UtcNow.ToString('o'))
__STDOUT_BEFORE__
Start-Process -FilePath '__SHELL__' -ArgumentList '-NoProfile -EncodedCommand __CHILD__' -WindowStyle Hidden
$launchOK=$?
__STDOUT_AFTER__
[IO.File]::WriteAllText('__AFTER__',([pscustomobject]@{UTC=[DateTime]::UtcNow.ToString('o');StartProcessSucceeded=$launchOK}|ConvertTo-Json -Compress))
Start-Sleep -Seconds 20
'@
    foreach($stage in @('ENTRY','BEFORE','AFTER')) {
        $label=if($stage -eq 'ENTRY'){'ENTRY'}else{$stage+'_START'}
        $statement=if($SynchronousStdout){"[Console]::Out.WriteLine('TALENRO_FULL_PARENT_$label');[Console]::Out.Flush()"}else{''}
        $parent=$parent.Replace('__STDOUT_'+$stage+'__',$statement)
    }
    $parent=$parent.Replace('__ENTRY__',(Join-Path $root 'parent-entry.utc').Replace("'","''")).Replace('__BEFORE__',(Join-Path $root 'parent-before-start.utc').Replace("'","''")).Replace('__SHELL__',$shell.Replace("'","''")).Replace('__CHILD__',$childEncoded).Replace('__AFTER__',(Join-Path $root 'parent-after-start.json').Replace("'","''"))
    [pscustomobject]@{RunRoot=$root;Executable=$shell;LimitSeconds=2;ParentScript=$parent;Arguments=@('-NoProfile','-EncodedCommand',[Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($parent)))}
}
function Get-TreeTimeoutProbeEvidence([string]$RunRoot,[object]$Receipt) {
    $workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..'))
    $scope=Join-Path $workspace '.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator'
    $root=Get-IsolatedFullPath $RunRoot
    if(-not (Test-IsolatedWithin $root $scope)){throw 'Evidence root outside owned scope'}
    $markers=@();$data=@()
    foreach($name in @('parent-entry.utc','parent-before-start.utc','parent-after-start.json','child-entry.utc')) {
        $file=Get-IsolatedFullPath (Join-Path $root $name)
        if(Test-Path -LiteralPath $file -PathType Leaf) {
            if([IsolatedFuzzFileInfo]::Links($file) -ne 1){throw 'Linked stage marker refused'}
            $markers+=$name;$data+=[pscustomobject]@{Name=$name;Text=[IO.File]::ReadAllText($file)}
        }
    }
    $stdoutStages=@()
    $stdout=Get-IsolatedFullPath (Join-Path $root 'process/stdout.txt')
    if(Test-Path -LiteralPath $stdout -PathType Leaf) {
        if([IsolatedFuzzFileInfo]::Links($stdout) -ne 1){throw 'Linked stdout evidence refused'}
        $allowed=@('TALENRO_FULL_PARENT_ENTRY','TALENRO_FULL_PARENT_BEFORE_START','TALENRO_FULL_PARENT_AFTER_START')
        $stdoutStages=@([IO.File]::ReadAllLines($stdout)|Where-Object {$allowed -ccontains $_})
    }
    [pscustomobject]@{Acceptance=$false;OriginalTestPassed=$false;DiagnosticOnly=$true;Markers=$markers;MarkerData=$data;StdoutStages=$stdoutStages;Receipt=$Receipt}
}
function Invoke-TreeTimeoutProbe([string]$RunRoot,[bool]$SynchronousStdout=$false) {
    $plan=Get-TreeTimeoutProbePlan $RunRoot $SynchronousStdout
    New-Item -ItemType Directory -Path $plan.RunRoot -ErrorAction Stop|Out-Null
    $plan|ConvertTo-Json -Depth 4|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'plan.json')
    try {
        $receipt=Invoke-RecordedProcess $plan.Executable $plan.Arguments @{} (Join-Path $plan.RunRoot 'process') $plan.LimitSeconds
        $evidence=Get-TreeTimeoutProbeEvidence $plan.RunRoot $receipt
        $evidence|ConvertTo-Json -Depth 8|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'result.json')
        if(-not $receipt.CleanupVerified -or $receipt.ActiveAfterCleanup -ne 0 -or $receipt.MembersAfterCleanup.Count -ne 0){throw 'Probe owned-job cleanup not verified; stop without retry'}
        return [pscustomobject]@{RunRoot=$plan.RunRoot;Acceptance=$false;OriginalTestPassed=$false;Markers=$evidence.Markers;StdoutStages=$evidence.StdoutStages;TimedOut=$receipt.TimedOut;TotalProcesses=$receipt.TotalProcesses;CleanupVerified=$receipt.CleanupVerified}
    } catch {
        [pscustomobject]@{UTC=[DateTime]::UtcNow.ToString('o');Error=$_.Exception.Message;Acceptance=$false}|ConvertTo-Json|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'stopped.json')
        throw
    }
}
