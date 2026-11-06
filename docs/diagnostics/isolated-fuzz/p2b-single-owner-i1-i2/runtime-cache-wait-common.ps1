# Definition-only helpers. No process enumeration, launch or filesystem writes on import.
function Get-C12WaitStartUtc($Value) {
    # Read-C12WaitRecord already deserializes the executor's Z timestamp to UTC.
    # Never coerce DateTime through a culture-dependent string: it loses Kind/ticks.
    if ($Value -isnot [DateTime] -or $Value.Kind -ne [DateTimeKind]::Utc) { throw 'START_UTC_REQUIRED' }
    return $Value
}
function Assert-C12WaitPath([string]$Path) {
    if (-not [IO.Path]::IsPathFullyQualified($Path)) { throw 'PATH_NOT_ABSOLUTE' }
    for ($item = [IO.DirectoryInfo]::new($Path); $null -ne $item; $item = $item.Parent) {
        if ((Test-Path -LiteralPath $item.FullName) -and ((Get-Item -LiteralPath $item.FullName -Force).Attributes -band [IO.FileAttributes]::ReparsePoint)) { throw 'REPARSE' }
    }
}
function New-C12WaitRoot([string]$Path) {
    $Path = [IO.Path]::GetFullPath($Path)
    Assert-C12WaitPath $Path
    if (Test-Path -LiteralPath $Path) { throw 'OCCUPIED_ROOT' }
    # Phase claim is exclusive even if another caller wins directory creation.
    [void][IO.Directory]::CreateDirectory($Path)
    $f = [IO.File]::Open((Join-Path $Path 'exclusive.claim'), 'CreateNew', 'Write', 'None')
    $f.Dispose()
    return $Path
}
function Assert-C12WaitLeaf([string]$Leaf) {
    if ($Leaf -cnotmatch '^[a-zA-Z0-9_-]+\.json$') { throw 'LEAF' }
}
function Write-C12WaitRecord([string]$Root, [string]$Leaf, $Value) {
    Assert-C12WaitPath $Root
    Assert-C12WaitLeaf $Leaf
    $path = Join-Path $Root $Leaf
    if (Test-Path -LiteralPath $path) { throw 'OCCUPIED_RECORD' }
    $bytes = [Text.UTF8Encoding]::new($false).GetBytes(($Value | ConvertTo-Json -Depth 30 -Compress))
    if ($bytes.Length -gt 65536) { throw 'RECORD_LIMIT' }
    [long]$total = 0
    foreach ($item in Get-ChildItem -LiteralPath $Root -Force -Recurse) {
        if ($item.Attributes -band [IO.FileAttributes]::ReparsePoint) { throw 'REPARSE' }
        if (-not $item.PSIsContainer) { $total += $item.Length }
    }
    if ($total + $bytes.Length -gt 4194304) { throw 'TOTAL_LIMIT' }
    $temp = Join-Path $Root ($Leaf + '.' + [guid]::NewGuid().ToString('N') + '.tmp')
    $stream = [IO.File]::Open($temp, 'CreateNew', 'Write', 'None')
    try { $stream.Write($bytes); $stream.Flush($true) } finally { $stream.Dispose() }
    [IO.File]::Move($temp, $path, $false)
    return @{ Path=$path; Length=$bytes.Length; SHA256=[Convert]::ToHexString([Security.Cryptography.SHA256]::HashData($bytes)) }
}
function Read-C12WaitRecord([string]$Root, [string]$Leaf) {
    Assert-C12WaitPath $Root
    Assert-C12WaitLeaf $Leaf
    $path = Join-Path $Root $Leaf
    if (-not (Test-Path -LiteralPath $path)) { return $null }
    $item = Get-Item -LiteralPath $path -Force
    if ($item.Attributes -band [IO.FileAttributes]::ReparsePoint) { throw 'REPARSE' }
    if ($item.Length -gt 65536) { throw 'RECORD_LIMIT' }
    return ([IO.File]::ReadAllText($path) | ConvertFrom-Json -AsHashtable -Depth 30)
}
function Assert-C12WaitIdentity($Expected, $Observed) {
    foreach ($value in @($Expected,$Observed)) {
        if ($null -eq $value -or $value.Schema -ne 1 -or $value.SessionId -cnotmatch '^[a-f0-9]{32}$' -or $value.Role -notin @('parent','go') -or $value.PID -isnot [ValueType] -or $value.PID -le 0 -or [decimal]$value.PID % 1 -ne 0 -or $value.CreationUtcTicks -isnot [ValueType] -or $value.CreationUtcTicks -le 0 -or [decimal]$value.CreationUtcTicks % 1 -ne 0 -or $value.ExecutableSHA256 -cnotmatch '^[A-F0-9]{64}$') { throw 'IDENTITY_SCHEMA' }
        if (-not [IO.Path]::IsPathFullyQualified($value.ExecutablePath) -or [IO.Path]::GetFullPath($value.ExecutablePath) -cne $value.ExecutablePath) { throw 'IDENTITY_PATH' }
    }
    foreach ($key in @('Schema','SessionId','Role','PID','CreationUtcTicks','ExecutablePath','ExecutableSHA256')) {
        if ($Expected[$key] -cne $Observed[$key]) { throw "IDENTITY_$key" }
    }
}
function Get-C12WaitIdentity([Diagnostics.Process]$Process, [string]$SessionId, [string]$Role) {
    $Process.Refresh()
    $path = [IO.Path]::GetFullPath($Process.MainModule.FileName)
    $value = @{ Schema=1; SessionId=$SessionId; Role=$Role; PID=$Process.Id; CreationUtcTicks=$Process.StartTime.ToUniversalTime().Ticks; ExecutablePath=$path; ExecutableSHA256=(Get-FileHash -LiteralPath $path -Algorithm SHA256).Hash }
    Assert-C12WaitIdentity $value $value
    return $value
}
function Get-C12WaitDecision($NowMs, $DeadlineMs, [bool]$Exited, [bool]$CollectorReady) {
    if ($Exited) { return 'EXITED' }
    if ($NowMs -ge $DeadlineMs) { return 'STOP' }
    if ($CollectorReady) { return 'START' }
    return 'WAIT'
}
function ConvertTo-C12WaitMetrics($Raw) {
    $result = @{}
    $missing = [Collections.Generic.List[string]]::new()
    foreach ($name in @('CPUms','WorkingSetBytes','PrivateBytes','ReadOperations','ReadBytes','WriteOperations','WriteBytes','ThreadStates','WaitReasons')) {
        $result[$name] = $null
        if ($Raw.ContainsKey($name) -and $null -ne $Raw[$name]) { $result[$name]=$Raw[$name] } else { $missing.Add($name) }
    }
    $result.UnavailableFields=$missing.ToArray()
    return $result
}
function Get-C12WaitDelta($Previous, $Current) {
    $result = @{}
    $invalid = [Collections.Generic.List[string]]::new()
    foreach ($name in @('CPUms','ReadOperations','ReadBytes','WriteOperations','WriteBytes')) {
        $result[$name]=$null
        if ($null -ne $Previous -and $Previous.ContainsKey($name) -and $Current.ContainsKey($name) -and $null -ne $Previous[$name] -and $null -ne $Current[$name] -and $Current[$name] -ge $Previous[$name]) {
            $result[$name]=[decimal]$Current[$name] - [decimal]$Previous[$name]
        } else { $invalid.Add($name) }
    }
    $result.InvalidFields=$invalid.ToArray()
    return $result
}
function New-C12WaitProcessInfo([string]$Executable, [string[]]$Arguments, [string]$WorkingDirectory) {
    $info = [Diagnostics.ProcessStartInfo]::new()
    $info.FileName=$Executable; $info.WorkingDirectory=$WorkingDirectory
    $info.UseShellExecute=$false; $info.CreateNoWindow=$true
    $info.RedirectStandardOutput=$true; $info.RedirectStandardError=$true
    foreach ($arg in $Arguments) { $info.ArgumentList.Add($arg) }
    return $info
}
function Stop-C12WaitOwnedProcess([Diagnostics.Process]$Process) {
    if ($null -eq $Process) { return }
    try { $null = $Process.Id } catch { return }
    if (-not $Process.HasExited) { $Process.Kill($true) }
    if (-not $Process.WaitForExit(5000)) { throw 'EXIT_UNCONFIRMED' }
}
function Wait-C12WaitOwnedProcess([Diagnostics.Process]$Process, [Diagnostics.Stopwatch]$Clock, [int]$DeadlineMs) {
    $out = $Process.StandardOutput.ReadToEndAsync()
    $err = $Process.StandardError.ReadToEndAsync()
    $remaining = [Math]::Max(0, $DeadlineMs - [int][Math]::Ceiling($Clock.Elapsed.TotalMilliseconds))
    $exited = $Process.WaitForExit($remaining)
    $timedOut = -not $exited
    if ($timedOut) { Stop-C12WaitOwnedProcess $Process }
    $confirmed = $Process.HasExited
    if (-not $confirmed) { throw 'EXIT_UNCONFIRMED' }
    if (-not $out.Wait(5000) -or -not $err.Wait(5000)) { throw 'OUTPUT_DRAIN_UNCONFIRMED' }
    if ([Text.Encoding]::UTF8.GetByteCount($out.Result) + [Text.Encoding]::UTF8.GetByteCount($err.Result) -gt 32768) { throw 'OUTPUT_LIMIT' }
    return @{ TimedOut=$timedOut; ExitConfirmed=$confirmed; ExitCode=$Process.ExitCode; ElapsedMs=$Clock.Elapsed.TotalMilliseconds; Stdout=$out.Result; Stderr=$err.Result }
}
function Get-C12WaitModuleHashes([string]$Repo) {
    $hashes=@{}
    foreach ($leaf in @('go.mod','go.sum','tools/devtools/go.mod','tools/devtools/go.sum')) { $hashes[$leaf]=(Get-FileHash -LiteralPath (Join-Path $Repo $leaf) -Algorithm SHA256).Hash }
    return $hashes
}
