param([Parameter(Mandatory)][ValidatePattern('^[a-f0-9]{32}$')][string]$SessionId)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'runtime-cache-wait-common.ps1')
$repo=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))
$root=New-C12WaitRoot (Join-Path $repo '.superpowers/sdd/2026-11-06-c12-cache-wait-v2/Live')
$go='C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/go.exe'
$goHash='D6A71E6DC9806CB3F1FCB817D562175C5A7B1505FC3F6638299648E0B9F0F74F'
$child=$null; $started=$false; $failure=$null; $result=$null
$before=Get-C12WaitModuleHashes $repo
try {
    if ((Get-FileHash -LiteralPath $go -Algorithm SHA256).Hash -cne $goHash) { throw 'PIN_MISMATCH' }
    $parent=Get-C12WaitIdentity ([Diagnostics.Process]::GetCurrentProcess()) $SessionId 'parent'
    $null=Write-C12WaitRecord $root 'parent.json' $parent
    # Only allowlisted launch coordinates, never command lines or environment values.
    @{ SessionId=$SessionId; ParentPID=$parent.PID; ParentCreationUtcTicks=$parent.CreationUtcTicks; Root=$root } | ConvertTo-Json -Compress
    $handshake=[Diagnostics.Stopwatch]::StartNew()
    $ready=$null
    while ($handshake.ElapsedMilliseconds -lt 60000) {
        if ($null -ne (Read-C12WaitRecord $root 'collector-stop.json')) { throw 'PREFLIGHT_STOPPED' }
        $ready=Read-C12WaitRecord $root 'ready.json'
        if ($null -ne $ready) { break }
        Start-Sleep -Milliseconds 100
    }
    if ($null -eq $ready -or $handshake.ElapsedMilliseconds -ge 60000) { throw 'PREFLIGHT_TIMEOUT' }
    if ($ready.SessionId -cne $SessionId -or $ready.Status -cne 'READY') { throw 'READY_BINDING' }
    Assert-C12WaitIdentity $parent $ready.Identity
    $child=[Diagnostics.Process]::new()
    $child.StartInfo=New-C12WaitProcessInfo $go @('mod','verify') $repo
    foreach ($key in @($child.StartInfo.Environment.Keys)) { if ($key -match '^(GO|GIT_)') { [void]$child.StartInfo.Environment.Remove($key) } }
    $envValues=@{ GOENV='off'; GOWORK='off'; GOTOOLCHAIN='local'; GOPATH='C:/Users/Lenovo/go'; GOMODCACHE='C:/Users/Lenovo/go/pkg/mod'; GOPROXY='off'; GOSUMDB='off'; GOAUTH='off'; GOVCS='all:off'; GOFLAGS='-mod=readonly'; GOMAXPROCS='2' }
    foreach ($key in $envValues.Keys) { $child.StartInfo.Environment[$key]=$envValues[$key] }
    $clock=[Diagnostics.Stopwatch]::StartNew()
    if (-not $child.Start()) { throw 'GO_START_FAILED' }
    $started=$true
    $identity=Get-C12WaitIdentity $child $SessionId 'go'
    if ($identity.ExecutableSHA256 -cne $goHash -or $identity.ExecutablePath -cne [IO.Path]::GetFullPath($go)) { throw 'GO_IDENTITY' }
    $null=Write-C12WaitRecord $root 'target.json' @{ SessionId=$SessionId; Identity=$identity; StartedUtc=$child.StartTime.ToUniversalTime().ToString('O'); DeadlineMs=120000 }
    $result=Wait-C12WaitOwnedProcess $child $clock 120000
} catch { $failure=$_.Exception.Message }
finally {
    $exitConfirmed=-not $started
    if ($null -ne $child) {
        try { Stop-C12WaitOwnedProcess $child; if ($started) { $exitConfirmed=$child.HasExited } } catch { $failure='EXIT_UNCONFIRMED: ' + $_.Exception.Message; $exitConfirmed=$false }
        $child.Dispose()
    }
    $after=Get-C12WaitModuleHashes $repo
    $unchanged=$true
    foreach ($key in $before.Keys) { if ($before[$key] -cne $after[$key]) { $unchanged=$false } }
    $receipt=@{ SessionId=$SessionId; Status=$(if ($started) { 'ATTEMPTED' } else { 'NOT_STARTED' }); Failure=$failure; GoStarted=$started; ExitConfirmed=$exitConfirmed; Result=$result; Before=$before; After=$after; ModuleFilesUnchanged=$unchanged; Devtools='NOT_RUN'; Acceptance='NOT_ACCEPTED' }
    $null=Write-C12WaitRecord $root 'executor-result.json' $receipt
    $receipt | ConvertTo-Json -Depth 15
}
if ($null -ne $failure -or -not $started -or -not $unchanged -or -not $exitConfirmed -or $result.TimedOut -or $result.ExitCode -ne 0) { exit 1 }
