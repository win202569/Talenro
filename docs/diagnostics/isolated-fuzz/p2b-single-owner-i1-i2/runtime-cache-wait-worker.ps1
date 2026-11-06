param([Parameter(Mandatory)][string]$SessionId,[Parameter(Mandatory)][int]$ParentPID,[Parameter(Mandatory)][long]$ParentCreationUtcTicks)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'runtime-cache-wait-common.ps1')
$selfPath=[IO.Path]::GetFullPath([Diagnostics.Process]::GetCurrentProcess().MainModule.FileName)
$selfHash=(Get-FileHash -LiteralPath $selfPath -Algorithm SHA256).Hash
$goPath=[IO.Path]::GetFullPath('C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/go.exe')
$goHash='D6A71E6DC9806CB3F1FCB817D562175C5A7B1505FC3F6638299648E0B9F0F74F'
$lastSequence=-1; $preflightDone=$false; $boundGo=$null
while ($null -ne ($line=[Console]::ReadLine())) {
    if ([Text.Encoding]::UTF8.GetByteCount($line) -gt 65536) { throw 'REQUEST_LIMIT' }
    $request=$line | ConvertFrom-Json -AsHashtable -Depth 20
    $response=@{ SessionId=$SessionId; Sequence=$request.Sequence; Phase=$request.Phase; Status='IDENTITY'; TimestampUtc=[DateTime]::UtcNow.ToString('O') }
    $process=$null
    try {
        if ($request.SessionId -cne $SessionId -or $request.Sequence -ne ($lastSequence+1) -or $request.Sequence -gt 24) { throw 'REQUEST_BINDING' }
        $lastSequence=$request.Sequence
        $expected=$request.Identity
        Assert-C12WaitIdentity $expected $expected
        if ($expected.SessionId -cne $SessionId) { throw 'SESSION_BINDING' }
        if ($request.Phase -ceq 'preflight' -and -not $preflightDone -and $request.Sequence -eq 0) {
            if ($expected.Role -cne 'parent' -or $expected.PID -ne $ParentPID -or $expected.CreationUtcTicks -ne $ParentCreationUtcTicks -or $expected.ExecutablePath -cne $selfPath -or $expected.ExecutableSHA256 -cne $selfHash) { throw 'PARENT_BINDING' }
        } elseif ($request.Phase -ceq 'sample' -and $preflightDone) {
            if ($expected.Role -cne 'go' -or $expected.ExecutablePath -cne $goPath -or $expected.ExecutableSHA256 -cne $goHash) { throw 'GO_BINDING' }
            if ($null -ne $boundGo) { Assert-C12WaitIdentity $boundGo $expected }
        } else { throw 'PHASE_BINDING' }
        # Exact requested ID only. No process list, command line or environment query.
        $process=[Diagnostics.Process]::GetProcessById([int]$expected.PID)
        $observed=Get-C12WaitIdentity $process $SessionId $expected.Role
        Assert-C12WaitIdentity $expected $observed
        $cim=Get-CimInstance -ClassName Win32_Process -Filter "ProcessId = $($expected.PID)" -Property ProcessId,ParentProcessId,CreationDate,ExecutablePath,ReadOperationCount,ReadTransferCount,WriteOperationCount,WriteTransferCount -OperationTimeoutSec 5 -ErrorAction Stop
        if ($null -eq $cim -or $null -eq $cim.CreationDate -or $null -eq $cim.ExecutablePath) { throw 'CIM_IDENTITY_UNAVAILABLE' }
        # CIM dates have microsecond precision; .NET retains 100ns precision separately.
        $cimTicks=$cim.CreationDate.ToUniversalTime().Ticks
        if ($cim.ProcessId -ne $expected.PID -or [decimal]::Floor([decimal]$cimTicks/10) -ne [decimal]::Floor([decimal]$expected.CreationUtcTicks/10) -or [IO.Path]::GetFullPath($cim.ExecutablePath) -cne $expected.ExecutablePath) { throw 'CIM_IDENTITY' }
        if ($expected.Role -ceq 'go' -and $cim.ParentProcessId -ne $ParentPID) { throw 'GO_NOT_OWNED_BY_PARENT' }
        $raw=@{}
        try { $raw.CPUms=$process.TotalProcessorTime.TotalMilliseconds } catch { }
        try { $raw.WorkingSetBytes=$process.WorkingSet64 } catch { }
        try { $raw.PrivateBytes=$process.PrivateMemorySize64 } catch { }
        foreach ($pair in @(@('ReadOperations','ReadOperationCount'),@('ReadBytes','ReadTransferCount'),@('WriteOperations','WriteOperationCount'),@('WriteBytes','WriteTransferCount'))) {
            if ($null -ne $cim.($pair[1])) { $raw[$pair[0]]=[decimal]$cim.($pair[1]) }
        }
        try {
            $states=@{}; $waits=@{}; $waitUnavailable=$false
            foreach ($thread in $process.Threads) {
                $state=$thread.ThreadState.ToString()
                if (-not $states.ContainsKey($state)) { $states[$state]=0 }; $states[$state]++
                if ($state -ceq 'Wait') {
                    try { $reason=$thread.WaitReason.ToString(); if (-not $waits.ContainsKey($reason)) { $waits[$reason]=0 }; $waits[$reason]++ } catch { $waitUnavailable=$true }
                }
            }
            $raw.ThreadStates=$states
            if (-not $waitUnavailable) { $raw.WaitReasons=$waits }
        } catch { }
        Assert-C12WaitIdentity $expected (Get-C12WaitIdentity $process $SessionId $expected.Role)
        $metrics=ConvertTo-C12WaitMetrics $raw
        foreach ($key in $metrics.Keys) { $response[$key]=$metrics[$key] }
        $response.Identity=$observed
        $response.Status=$(if ($metrics.UnavailableFields.Count -eq 0) { 'OBSERVED' } else { 'UNAVAILABLE' })
        if ($request.Phase -ceq 'preflight') { $preflightDone=$true } else { $boundGo=$expected }
    } catch {
        $response.Error=$_.Exception.Message
        # Uncertain identity is never relabelled as a harmless exit.
        $response.Status='IDENTITY'
    } finally { if ($null -ne $process) { $process.Dispose() } }
    $json=$response | ConvertTo-Json -Depth 20 -Compress
    if ([Text.Encoding]::UTF8.GetByteCount($json) -gt 65536) { throw 'RESPONSE_LIMIT' }
    [Console]::WriteLine($json)
    if ($response.Status -ceq 'IDENTITY') { break }
}
