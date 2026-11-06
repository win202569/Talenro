param([Parameter(Mandatory)][ValidateSet('Red','Green')][string]$Phase)
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'
$repo = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))
$root = Join-Path $repo ".superpowers/sdd/2026-11-06-c12-cache-wait-v1/$Phase"
# Bootstrap is independent of the implementation being tested. Never reuse a phase.
for ($p = [IO.DirectoryInfo]::new($root); $null -ne $p; $p = $p.Parent) {
    if ($p.Exists -and ($p.Attributes -band [IO.FileAttributes]::ReparsePoint)) { throw 'REPARSE' }
}
if (Test-Path -LiteralPath $root) { throw 'OCCUPIED_PHASE' }
[void][IO.Directory]::CreateDirectory($root)
$claim = [IO.File]::Open((Join-Path $root 'claim.json'), 'CreateNew', 'Write', 'None')
try { $bytes = [Text.Encoding]::UTF8.GetBytes('{"Phase":"' + $Phase + '","Budget":"one invocation"}'); $claim.Write($bytes) } finally { $claim.Dispose() }
$results = [Collections.Generic.List[object]]::new()
function Assert-True($Condition, $Message) { if (-not $Condition) { throw $Message } }
function Reject($Body, $Pattern) {
    $caught = $null
    try { & $Body } catch { $caught = $_.Exception.Message }
    Assert-True ($null -ne $caught -and $caught -match $Pattern) "Expected rejection: $Pattern; actual: $caught"
}
function Case($Name, $Body) { & $Body; $results.Add(@{ Name=$Name; Status='PASS' }) }
try {
    $common = Join-Path $PSScriptRoot 'runtime-cache-wait-common.ps1'
    if (-not (Test-Path -LiteralPath $common)) { throw 'MISSING_IMPLEMENTATION: runtime-cache-wait-common.ps1' }
    . $common
    $identity = @{ Schema=1; SessionId=('a' * 32); Role='parent'; PID=123; CreationUtcTicks=638000000000000000L; ExecutablePath=[IO.Path]::GetFullPath('C:/fixture/pwsh.exe'); ExecutableSHA256=('A' * 64) }
    # Removing any of these comparisons must admit the corresponding wrong identity.
    foreach ($field in @('PID','CreationUtcTicks','ExecutablePath','ExecutableSHA256','SessionId')) {
        Case "identity-$field" {
            $wrong = $identity.Clone()
            switch ($field) {
                PID { $wrong.PID=124 }
                CreationUtcTicks { $wrong.CreationUtcTicks=638000000000000001L }
                ExecutablePath { $wrong.ExecutablePath=[IO.Path]::GetFullPath('C:/fixture/other.exe') }
                ExecutableSHA256 { $wrong.ExecutableSHA256='B' * 64 }
                SessionId { $wrong.SessionId='b' * 32 }
            }
            Assert-C12WaitIdentity $identity $identity
            Reject { Assert-C12WaitIdentity $identity $wrong } 'IDENTITY'
        }
    }
    Case 'occupied-record-preserved' {
        $null = Write-C12WaitRecord $root 'sentinel.json' @{ Value='original' }
        $before = [IO.File]::ReadAllBytes((Join-Path $root 'sentinel.json'))
        Reject { Write-C12WaitRecord $root 'sentinel.json' @{ Value='replacement' } } 'OCCUPIED'
        Assert-True ([Convert]::ToBase64String($before) -ceq [Convert]::ToBase64String([IO.File]::ReadAllBytes((Join-Path $root 'sentinel.json')))) 'Sentinel changed'
    }
    Case 'partial-temp-never-consumed' {
        [IO.File]::WriteAllText((Join-Path $root 'partial.json.tmp'), '{')
        Assert-True ($null -eq (Read-C12WaitRecord $root 'partial.json')) 'Accepted incomplete publication'
        Reject { Read-C12WaitRecord $root 'partial.json.tmp' } 'LEAF'
    }
    Case 'missing-metrics-remain-null' {
        $s = ConvertTo-C12WaitMetrics @{}
        Assert-True ($null -eq $s.CPUms -and $null -eq $s.ReadBytes -and 'CPUms' -in $s.UnavailableFields -and 'ReadBytes' -in $s.UnavailableFields) 'Missing counter became zero'
    }
    Case 'counter-reset-invalidates-delta' {
        $d = Get-C12WaitDelta @{ ReadBytes=100; CPUms=20 } @{ ReadBytes=99; CPUms=25 }
        Assert-True ($null -eq $d.ReadBytes -and $d.CPUms -eq 5 -and 'ReadBytes' -in $d.InvalidFields) 'Invalid delta accepted'
    }
    Case 'record-size-bound' {
        Reject { Write-C12WaitRecord $root 'large.json' @{ Value=('x' * 65536) } } 'RECORD_LIMIT'
        Assert-True (-not (Test-Path (Join-Path $root 'large.json'))) 'Oversized record published'
    }
    Case 'total-size-bound' {
        $quota = New-C12WaitRoot (Join-Path $root 'quota')
        $f = [IO.File]::Open((Join-Path $quota 'retained.tmp'), 'CreateNew', 'Write', 'None')
        try { $f.SetLength(4194304) } finally { $f.Dispose() }
        Reject { Write-C12WaitRecord $quota 'overflow.json' @{ Value=1 } } 'TOTAL_LIMIT'
    }
    Case 'no-ready-no-launch' {
        Assert-True ((Get-C12WaitDecision 0 120000 $false $false) -eq 'WAIT') 'Launch without ready'
        Assert-True ((Get-C12WaitDecision 0 120000 $false $true) -eq 'START') 'Ready ignored'
    }
    Case 'deadline-independent-of-collector' {
        foreach ($ready in @($false,$true)) {
            Assert-True ((Get-C12WaitDecision 120000 120000 $false $ready) -eq 'STOP') 'Deadline extended'
            Assert-True ((Get-C12WaitDecision 120001 120000 $false $ready) -eq 'STOP') 'Late ready restarted work'
        }
        Assert-True ((Get-C12WaitDecision 100 120000 $true $false) -eq 'EXITED') 'Exit not observed'
    }
    Case 'occupied-root-rejected' { Reject { New-C12WaitRoot $root } 'OCCUPIED' }
    $pwsh = (Get-Process -Id $PID).Path
    Case 'owned-normal-exit' {
        $p = [Diagnostics.Process]::new()
        $p.StartInfo = New-C12WaitProcessInfo $pwsh @('-NoProfile','-NonInteractive','-Command','exit 0') $repo
        $clock = [Diagnostics.Stopwatch]::StartNew()
        try {
            Assert-True ($p.Start()) 'Fixture start failed'
            $r = Wait-C12WaitOwnedProcess $p $clock 5000
            Assert-True ($r.ExitConfirmed -and -not $r.TimedOut -and $r.ExitCode -eq 0) 'Normal lifecycle failed'
        } finally { Stop-C12WaitOwnedProcess $p; $p.Dispose() }
    }
    Case 'owned-one-second-cutoff' {
        $p = [Diagnostics.Process]::new()
        $p.StartInfo = New-C12WaitProcessInfo $pwsh @('-NoProfile','-NonInteractive','-Command','Start-Sleep -Seconds 30') $repo
        $clock = [Diagnostics.Stopwatch]::StartNew()
        try {
            Assert-True ($p.Start()) 'Fixture start failed'
            $r = Wait-C12WaitOwnedProcess $p $clock 1000
            Assert-True ($r.ExitConfirmed -and $r.TimedOut -and $r.ElapsedMs -lt 7000) 'Watchdog/cleanup failed'
        } finally { Stop-C12WaitOwnedProcess $p; $p.Dispose() }
    }
    if ($Phase -eq 'Red') { throw 'UNEXPECTED_RED_PASS' }
    Assert-True ($results.Count -eq 16) 'Case count mismatch'
    $outcome = @{ Status='WAIT_DIAGNOSTIC_TESTS_VERIFIED'; Phase=$Phase; Cases=$results.ToArray(); Count=$results.Count }
    [IO.File]::WriteAllText((Join-Path $root 'result.json'), ($outcome | ConvertTo-Json -Depth 10))
    $outcome | ConvertTo-Json -Depth 10
} catch {
    $outcome = @{ Status='STOP'; Phase=$Phase; Error=$_.Exception.Message; Cases=$results.ToArray(); Count=$results.Count }
    [IO.File]::WriteAllText((Join-Path $root 'result.json'), ($outcome | ConvertTo-Json -Depth 10))
    throw
}
