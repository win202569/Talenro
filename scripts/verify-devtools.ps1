$devtoolsRuntime = $PSVersionTable
if ($devtoolsRuntime.PSEdition -ne 'Core' -or
    [string]$devtoolsRuntime.PSVersion -cne '7.6.5' -or
    $devtoolsRuntime.ContainsKey('PSVersionPreReleaseLabel') -or
    [Environment]::OSVersion.Platform -ne [PlatformID]::Win32NT) {
  [Console]::Error.WriteLine('verify-devtools: PowerShell 7.6.5 required.')
  exit 1
}

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

function Invoke-DevtoolsCommand {
  param([string]$GoPath, [string[]]$Arguments, [string]$Directory, [DateTime]$Deadline)
  $remaining = [Math]::Floor(($Deadline - [DateTime]::UtcNow).TotalMilliseconds)
  if ($remaining -le 0) { $script:exitCode = 124; throw 'deadline' }
  $process = [System.Diagnostics.Process]::new()
  $process.StartInfo.FileName = $GoPath
  $process.StartInfo.WorkingDirectory = $Directory
  $process.StartInfo.UseShellExecute = $false
  $process.StartInfo.CreateNoWindow = $true
  $process.StartInfo.RedirectStandardOutput = $true
  $process.StartInfo.RedirectStandardError = $true
  foreach ($argument in $Arguments) { $process.StartInfo.ArgumentList.Add($argument) }
  $captures = @([System.IO.MemoryStream]::new(), [System.IO.MemoryStream]::new())
  $started = $false
  try {
    $started = $process.Start()
    if (-not $started) { throw 'process start failed' }
    $null = $process.Handle
    $streams = @($process.StandardOutput.BaseStream, $process.StandardError.BaseStream)
    $buffers = @([byte[]]::new(65536), [byte[]]::new(65536))
    $reads = @($streams[0].ReadAsync($buffers[0], 0, 65536), $streams[1].ReadAsync($buffers[1], 0, 65536))
    $finished = @($false, $false)
    $capturedBytes = 0L
    while (-not ($finished[0] -and $finished[1] -and $process.HasExited)) {
      if ([DateTime]::UtcNow -ge $Deadline) { $script:exitCode = 124; throw 'deadline' }
      for ($i = 0; $i -lt 2; $i++) {
        if ($finished[$i] -or -not $reads[$i].IsCompleted) { continue }
        $count = $reads[$i].GetAwaiter().GetResult()
        if ($count -eq 0) { $finished[$i] = $true; continue }
        # Cap retained stdout + stderr while the command runs, not after exit.
        if ($capturedBytes + $count -gt 4194304) { throw 'output limit' }
        $captures[$i].Write($buffers[$i], 0, $count)
        $capturedBytes += $count
        $reads[$i] = $streams[$i].ReadAsync($buffers[$i], 0, 65536)
      }
      [System.Threading.Thread]::Sleep(1)
    }
    $process.WaitForExit()
    $process.Refresh()
    if ($null -eq $process.ExitCode) { throw 'unknown exit' }
    if ($process.ExitCode -ne 0) { $script:exitCode = $process.ExitCode; throw 'command failed' }
    return [System.Text.Encoding]::UTF8.GetString($captures[0].ToArray()).TrimEnd("`r", "`n")
  } finally {
    try {
      if ($started -and -not $process.HasExited) {
        $process.Kill($true)
        if (-not $process.WaitForExit(1000)) { $script:exitCode = 1; throw 'cleanup failed' }
      }
    } finally {
      $process.Dispose()
      foreach ($capture in $captures) { $capture.Dispose() }
    }
  }
}

function ConvertTo-DevtoolsCacheKey {
  param([string]$Value)
  return [regex]::Replace($Value, '[A-Z]', { param($match) '!' + $match.Value.ToLowerInvariant() })
}

$deadline = [DateTime]::UtcNow.AddSeconds(900)
$stage = 'module'
$exitCode = 1
$savedEnvironment = @{}
try {
  $repoRoot = Split-Path -Parent $PSScriptRoot
  $moduleRoot = Join-Path $repoRoot 'tools/devtools'
  if (-not [System.IO.File]::Exists((Join-Path $moduleRoot 'go.mod')) -or
      -not [System.IO.File]::Exists((Join-Path $moduleRoot 'go.sum'))) {
    throw 'missing module locks'
  }
  $stage = 'environment'
  if ($args.Count -ne 0 -or $env:GOFLAGS -match '(^|\s)-(modfile|overlay)(=|\s|$)') { throw 'module override' }
  foreach ($item in @(Get-ChildItem Env: | Where-Object { $_.Name -match '^GO' })) {
    $savedEnvironment[$item.Name] = $item.Value
    Remove-Item -LiteralPath ('Env:' + $item.Name)
  }
  $fixedEnvironment = @{
    GOWORK='off'; GOENV='off'; GOTOOLCHAIN='local'; GOPROXY='off'; GOSUMDB='off';
    GOAUTH='off'; GOVCS='all:off'; GOFLAGS='-mod=readonly';
    GOPATH=(Join-Path $env:USERPROFILE 'go'); GOMODCACHE=(Join-Path $env:USERPROFILE 'go/pkg/mod')
  }
  foreach ($key in $fixedEnvironment.Keys) { [Environment]::SetEnvironmentVariable($key, $fixedEnvironment[$key], 'Process') }
  $moduleCache = $fixedEnvironment.GOMODCACHE
  $goPath = Join-Path $moduleCache 'golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/go.exe'
  $stage = 'toolchain'
  if (-not [System.IO.File]::Exists($goPath)) { $exitCode = 127; throw 'missing go' }
  . (Join-Path $PSScriptRoot 'private/devtools-process.ps1')
  Initialize-DevtoolsProcessOwnership
  $version = Invoke-DevtoolsCommand -GoPath $goPath -Arguments @('version') -Directory $moduleRoot -Deadline $deadline
  if ($version -cne 'go version go1.26.5 windows/amd64') { throw 'wrong version' }
  $stage = 'dependencies'
  $modules = Invoke-DevtoolsCommand -GoPath $goPath -Arguments @('list','-m','-f','{{if not .Main}}{{.Path}}|{{.Version}}|{{.Dir}}{{end}}','all') -Directory $moduleRoot -Deadline $deadline
  foreach ($line in ($modules -split "`n")) {
    if ([string]::IsNullOrWhiteSpace($line)) { continue }
    $fields = $line.TrimEnd("`r") -split '\|'
    if ($fields.Count -ne 3) { throw 'malformed module' }
    if ($fields[0] -in @('go','toolchain')) { continue }
    if ($fields[0] -notmatch '^[A-Za-z0-9._~+!/-]+$' -or $fields[1] -notmatch '^v[A-Za-z0-9._+!-]+$' -or $fields[0] -match '(^|/)\.\.(/|$)') { throw 'malformed module identity' }
    $key = ConvertTo-DevtoolsCacheKey $fields[0]
    $versionKey = ConvertTo-DevtoolsCacheKey $fields[1]
    $expected = [System.IO.Path]::GetFullPath((Join-Path $moduleCache ($key + '@' + $versionKey)))
    if ([string]::IsNullOrEmpty($fields[2]) -or [System.IO.Path]::GetFullPath($fields[2]) -cne $expected -or -not [System.IO.Directory]::Exists($expected)) { throw 'missing module directory' }
    $zip = Join-Path $moduleCache ('cache/download/' + $key + '/@v/' + $versionKey + '.zip')
    if (-not [System.IO.File]::Exists($zip) -or -not [System.IO.File]::Exists($zip + 'hash')) { throw 'missing module archive' }
  }
  $stage = 'integrity'
  $verified = Invoke-DevtoolsCommand -GoPath $goPath -Arguments @('mod','verify') -Directory $moduleRoot -Deadline $deadline
  if ($verified -cne 'all modules verified') { throw 'unexpected verification result' }
  $exitCode = 0
} catch {
  # Never include the original exception or repository paths in public output.
} finally {
  try {
    if ($savedEnvironment.Count -gt 0 -or (Test-Path variable:fixedEnvironment)) {
      foreach ($item in @(Get-ChildItem Env: | Where-Object { $_.Name -match '^GO' })) { Remove-Item -LiteralPath ('Env:' + $item.Name) }
      foreach ($key in $savedEnvironment.Keys) { [Environment]::SetEnvironmentVariable($key, $savedEnvironment[$key], 'Process') }
    }
  } catch { $stage = 'cleanup'; $exitCode = 1 }
}
if ($exitCode -eq 0) { [Console]::Out.WriteLine('verify-devtools: passed') }
else { [Console]::Error.WriteLine("verify-devtools: $stage failed with exit code $exitCode.") }
exit $exitCode
