param([Parameter(Mandatory)][ValidatePattern('^[a-f0-9]{32}$')][string]$SessionId,[Parameter(Mandatory)][int]$ParentPID,[Parameter(Mandatory)][long]$ParentCreationUtcTicks)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'runtime-cache-wait-common.ps1')
$repo=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))
$root=Join-Path $repo '.superpowers/sdd/2026-11-06-c12-cache-wait-v2/Live'
Assert-C12WaitPath $root
$worker=$null; $failure=$null; $count=0; $exitConfirmed=$false; $stderr=$null; $endReason=$null
function Invoke-BoundQuery($Identity,[int]$Sequence,[string]$Phase) {
    $request=@{ SessionId=$SessionId; Sequence=$Sequence; Phase=$Phase; Identity=$Identity }
    $null=Write-C12WaitRecord $root ("request-{0:D2}.json" -f $Sequence) $request
    $json=$request | ConvertTo-Json -Depth 20 -Compress
    if ([Text.Encoding]::UTF8.GetByteCount($json) -gt 65536) { throw 'REQUEST_LIMIT' }
    $queryClock=[Diagnostics.Stopwatch]::StartNew()
    $worker.StandardInput.WriteLine($json); $worker.StandardInput.Flush()
    $pending=$worker.StandardOutput.ReadLineAsync()
    if (-not $pending.Wait([Math]::Max(0,5000-[int]$queryClock.ElapsedMilliseconds))) { throw 'QUERY_TIMEOUT' }
    if ($null -eq $pending.Result) { throw 'WORKER_EOF' }
    if ([Text.Encoding]::UTF8.GetByteCount($pending.Result) -gt 65536) { throw 'RESPONSE_LIMIT' }
    $answer=$pending.Result | ConvertFrom-Json -AsHashtable -Depth 20
    $null=Write-C12WaitRecord $root ("response-{0:D2}.json" -f $Sequence) $answer
    if ($answer.SessionId -cne $SessionId -or $answer.Sequence -ne $Sequence -or $answer.Phase -cne $Phase -or $answer.Status -notin @('OBSERVED','UNAVAILABLE')) { throw 'RESPONSE_BINDING_OR_STATUS' }
    Assert-C12WaitIdentity $Identity $answer.Identity
    return $answer
}
try {
    $claim=[IO.File]::Open((Join-Path $root 'collector.claim'), 'CreateNew','Write','None'); $claim.Dispose()
    $parent=Read-C12WaitRecord $root 'parent.json'
    if ($null -eq $parent -or $parent.PID -ne $ParentPID -or $parent.CreationUtcTicks -ne $ParentCreationUtcTicks -or $parent.SessionId -cne $SessionId) { throw 'PARENT_BINDING' }
    $pwsh=[Diagnostics.Process]::GetCurrentProcess().MainModule.FileName
    $worker=[Diagnostics.Process]::new()
    $worker.StartInfo=New-C12WaitProcessInfo $pwsh @('-NoProfile','-NonInteractive','-File',(Join-Path $PSScriptRoot 'runtime-cache-wait-worker.ps1'),'-SessionId',$SessionId,'-ParentPID',"$ParentPID",'-ParentCreationUtcTicks',"$ParentCreationUtcTicks") $repo
    $worker.StartInfo.RedirectStandardInput=$true
    if (-not $worker.Start()) { throw 'WORKER_START_FAILED' }
    $stderr=$worker.StandardError.ReadToEndAsync()
    $null=Write-C12WaitRecord $root 'worker.json' @{ SessionId=$SessionId; PID=$worker.Id; CreationUtcTicks=$worker.StartTime.ToUniversalTime().Ticks }
    $preflight=Invoke-BoundQuery $parent 0 'preflight'
    foreach ($key in @('ReadOperations','ReadBytes','WriteOperations','WriteBytes')) { if ($null -eq $preflight[$key]) { throw 'PREFLIGHT_IO_UNAVAILABLE' } }
    $null=Write-C12WaitRecord $root 'ready.json' @{ SessionId=$SessionId; Status='READY'; Identity=$parent }
    $targetClock=[Diagnostics.Stopwatch]::StartNew(); $target=$null
    while ($targetClock.ElapsedMilliseconds -lt 15000) {
        if ($null -ne (Read-C12WaitRecord $root 'executor-result.json')) { throw 'EXECUTOR_FINISHED_BEFORE_TARGET' }
        $target=Read-C12WaitRecord $root 'target.json'
        if ($null -ne $target) { break }
        Start-Sleep -Milliseconds 100
    }
    if ($null -eq $target -or $targetClock.ElapsedMilliseconds -ge 15000 -or $target.SessionId -cne $SessionId -or $target.DeadlineMs -ne 120000) { throw 'TARGET_TIMEOUT_OR_BINDING' }
    $previous=$null
    $start=Get-C12WaitStartUtc $target.StartedUtc
    $next=5000.0
    while ($count -lt 24) {
        if ($null -ne (Read-C12WaitRecord $root 'executor-result.json')) { $endReason='EXECUTOR_FINISHED'; break }
        $elapsed=([DateTime]::UtcNow-$start).TotalMilliseconds
        if ($elapsed -ge 120000) { $endReason='DEADLINE'; break }
        if ($elapsed -lt $next) { Start-Sleep -Milliseconds 100; continue }
        $answer=Invoke-BoundQuery $target.Identity ($count+1) 'sample'
        $null=Write-C12WaitRecord $root ("sample-{0:D2}.json" -f ($count+1)) @{ SessionId=$SessionId; Raw=$answer; Delta=(Get-C12WaitDelta $previous $answer); ObservedElapsedMs=([DateTime]::UtcNow-$start).TotalMilliseconds }
        $previous=$answer; $count++
        # No catch-up bursts when observation itself is slow.
        $next=([DateTime]::UtcNow-$start).TotalMilliseconds+5000
    }
} catch { $failure=$_.Exception.Message }
finally {
    if ($null -ne $worker) {
        try {
            $worker.StandardInput.Close()
            Stop-C12WaitOwnedProcess $worker
            $exitConfirmed=$worker.HasExited
        } catch { $failure='WORKER_EXIT_UNCONFIRMED: ' + $_.Exception.Message }
        $worker.Dispose()
    } else { $exitConfirmed=$true }
    $workerError=$null
    if ($null -ne $stderr -and $stderr.IsCompletedSuccessfully) {
        if ([Text.Encoding]::UTF8.GetByteCount($stderr.Result) -le 8192) { $workerError=$stderr.Result } else { $workerError='STDERR_LIMIT' }
    }
    $receipt=@{ SessionId=$SessionId; Samples=$count; Failure=$failure; WorkerExitConfirmed=$exitConfirmed; WorkerStderr=$workerError; EndReason=$(if ($null -ne $failure) { 'ERROR' } elseif ($count -ge 24) { 'SAMPLE_LIMIT' } else { $endReason }); Status=$(if ($null -eq $failure) { 'COLLECTION_ENDED' } else { 'STOPPED' }) }
    $null=Write-C12WaitRecord $root 'collector-stop.json' $receipt
    $receipt | ConvertTo-Json -Depth 10
}
if ($null -ne $failure -or -not $exitConfirmed) { exit 1 }
