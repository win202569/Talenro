$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
Add-Type -AssemblyName System.Net.Http
$script:SmokeSinkDirectory = $null
$script:SmokeRunCounter = 0
$script:SmokeMirrorIO = [System.Collections.Generic.Dictionary[int, object]]::new()
$null = Add-Type -TypeDefinition @'
using System;
using System.Collections.Generic;
using System.Runtime.InteropServices;

public static class TalenroSmokeEnvironment {
    [DllImport("kernel32.dll", CharSet = CharSet.Unicode, SetLastError = true)]
    private static extern IntPtr GetEnvironmentStringsW();

    [DllImport("kernel32.dll", CharSet = CharSet.Unicode, SetLastError = true)]
    private static extern bool FreeEnvironmentStringsW(IntPtr environmentBlock);

    public static string[] NamesWithPrefix(string prefix) {
        var names = new List<string>();
        IntPtr block = GetEnvironmentStringsW();
        if (block == IntPtr.Zero) {
            throw new InvalidOperationException("environment block unavailable");
        }
        try {
            long offset = 0;
            while (true) {
                string entry = Marshal.PtrToStringUni(new IntPtr(block.ToInt64() + offset));
                if (String.IsNullOrEmpty(entry)) {
                    break;
                }
                offset += (entry.Length + 1) * 2L;
                int separator = entry.IndexOf('=');
                if (separator > 0) {
                    string name = entry.Substring(0, separator);
                    if (name.StartsWith(prefix, StringComparison.OrdinalIgnoreCase)) {
                        names.Add(name);
                    }
                }
            }
        } finally {
            FreeEnvironmentStringsW(block);
        }
        return names.ToArray();
    }
}
'@

function Stop-Smoke {
  param(
    [Parameter(Mandatory)]
    [int]$ExitCode,

    [Parameter(Mandatory)]
    [string]$Stage
  )

  if ($ExitCode -eq 0) {
    $ExitCode = 1
  }
  $failure = [System.Exception]::new('sanitized smoke failure')
  $failure.Data['ExitCode'] = $ExitCode
  $failure.Data['Stage'] = $Stage
  throw $failure
}

function Get-SmokeFailure {
  param([Parameter(Mandatory)]$ErrorRecord)

  $result = @{ ExitCode = 1; Stage = 'smoke: internal' }
  if ($ErrorRecord.Exception.Data.Contains('ExitCode')) {
    $candidateExitCode = $ErrorRecord.Exception.Data['ExitCode']
    if ($null -ne $candidateExitCode) {
      try {
        $convertedExitCode = [int]$candidateExitCode
        if ($convertedExitCode -ne 0) {
          $result.ExitCode = $convertedExitCode
        }
      } catch {
        $result.ExitCode = 1
      }
    }
  }
  if ($ErrorRecord.Exception.Data.Contains('Stage')) {
    $result.Stage = [string]$ErrorRecord.Exception.Data['Stage']
  }
  return $result
}

function Get-SmokeCompletedProcessExitCode {
  param(
    [Parameter(Mandatory)][System.Diagnostics.Process]$Process,
    [Parameter(Mandatory)][string]$Stage
  )

  try {
    if (-not $Process.HasExited) {
      Stop-Smoke -ExitCode 1 -Stage $Stage
    }

    # Complete the state/stream update after a successful timed wait before
    # reading ExitCode. Callers acquire Process.Handle immediately at launch.
    $Process.WaitForExit()
    $Process.Refresh()
    $rawExitCode = $Process.ExitCode
    if ($null -eq $rawExitCode) {
      Stop-Smoke -ExitCode 1 -Stage $Stage
    }
    return [int]$rawExitCode
  } catch {
    if ($_.Exception.Data.Contains('Stage')) {
      throw
    }
    Stop-Smoke -ExitCode 1 -Stage $Stage
  }
}

function ConvertTo-SmokeCommandLineArgument {
  param([Parameter(Mandatory)][AllowEmptyString()][string]$Argument)

  if ($Argument.Length -gt 0 -and $Argument -notmatch '[\s"]') {
    return $Argument
  }
  $builder = [System.Text.StringBuilder]::new()
  $null = $builder.Append('"')
  $backslashes = 0
  foreach ($character in $Argument.ToCharArray()) {
    if ($character -eq [char]92) {
      $backslashes++
      continue
    }
    if ($character -eq [char]34) {
      $null = $builder.Append(('\' * (($backslashes * 2) + 1)))
      $null = $builder.Append('"')
      $backslashes = 0
      continue
    }
    if ($backslashes -gt 0) {
      $null = $builder.Append(('\' * $backslashes))
      $backslashes = 0
    }
    $null = $builder.Append($character)
  }
  if ($backslashes -gt 0) {
    $null = $builder.Append(('\' * ($backslashes * 2)))
  }
  $null = $builder.Append('"')
  return $builder.ToString()
}

function Join-SmokeCommandLine {
  param([Parameter(Mandatory)][string[]]$ArgumentList)

  return (($ArgumentList | ForEach-Object {
    ConvertTo-SmokeCommandLineArgument -Argument $_
  }) -join ' ')
}

function Get-SmokeProcessLaunch {
  param(
    [Parameter(Mandatory)][string]$Source,
    [Parameter(Mandatory)][string[]]$ArgumentList,
    [Parameter(Mandatory)][string]$Stage
  )

  if ([System.IO.Path]::GetExtension($Source) -notin @('.cmd', '.bat')) {
    return @{
      FilePath = $Source
      ArgumentString = (Join-SmokeCommandLine -ArgumentList $ArgumentList)
    }
  }

  $commandInterpreter = Join-Path $env:SystemRoot 'System32\cmd.exe'
  if (-not [System.IO.File]::Exists($commandInterpreter)) {
    Stop-Smoke -ExitCode 127 -Stage $Stage
  }

  $commandParts = [System.Collections.Generic.List[string]]::new()
  foreach ($value in @($Source) + @($ArgumentList)) {
    if ($value -match '[\r\n"%!]') {
      Stop-Smoke -ExitCode 1 -Stage $Stage
    }
    if ($commandParts.Count -eq 0 -or $value.Length -eq 0 -or $value -match '[\s&|<>()^]') {
      $commandParts.Add('"' + $value + '"')
    } else {
      $commandParts.Add($value)
    }
  }

  return @{
    FilePath = $commandInterpreter
    ArgumentString = ('/d /q /v:off /s /c "' + ($commandParts -join ' ') + '"')
  }
}

function Remove-SmokeSinkFiles {
  param([Parameter(Mandatory)][string[]]$Paths)

  $succeeded = $true
  foreach ($path in $Paths) {
    try {
      [System.IO.File]::Delete($path)
    } catch {
      $succeeded = $false
    }
  }
  return $succeeded
}

function Stop-SmokeOwnedProcessTree {
  param(
    [Parameter(Mandatory)][System.Diagnostics.Process]$Process,
    [Parameter(Mandatory)][string]$SinkPrefix
  )

  if ($Process.HasExited) {
    return
  }
  $taskKill = Join-Path $env:SystemRoot 'System32\taskkill.exe'
  if ([System.IO.File]::Exists($taskKill)) {
    $killer = $null
    $killOut = $SinkPrefix + '.kill.stdout'
    $killErr = $SinkPrefix + '.kill.stderr'
    try {
      $killer = Start-Process -FilePath $taskKill -ArgumentList "/PID $($Process.Id) /T /F" -PassThru -WindowStyle Hidden -RedirectStandardOutput $killOut -RedirectStandardError $killErr
      if (-not $killer.WaitForExit(5000)) {
        Stop-Process -Id $killer.Id -Force -ErrorAction SilentlyContinue
        $null = $killer.WaitForExit(5000)
      }
    } catch {
      # The exact-PID fallback below remains authoritative.
    } finally {
      if ($null -ne $killer) {
        try {
          $killer.Dispose()
        } catch {
          # A timeout is already the authoritative nonzero command status.
        }
      }
      $null = Remove-SmokeSinkFiles -Paths @($killOut, $killErr)
    }
  }
  if (-not $Process.HasExited) {
    Stop-Process -Id $Process.Id -Force -ErrorAction Stop
  }
  if (-not $Process.WaitForExit(5000)) {
    Stop-Smoke -ExitCode 1 -Stage 'smoke: process cleanup'
  }
}

function Invoke-QuietExternal {
  param(
    [Parameter(Mandatory)]
    [string]$Stage,

    [Parameter(Mandatory)]
    [string]$FilePath,

    [Parameter(Mandatory)]
    [string[]]$ArgumentList,

    [string]$WorkingDirectory = '',
    [ValidateRange(1, 600)][int]$TimeoutSeconds = 600,
    [switch]$CaptureOutput
  )

  $commands = @(Get-Command -Name $FilePath -CommandType Application -ErrorAction SilentlyContinue)
  if ($commands.Count -eq 0) {
    Stop-Smoke -ExitCode 127 -Stage $Stage
  }

  if ($null -eq $script:SmokeSinkDirectory) {
    Stop-Smoke -ExitCode 1 -Stage 'smoke: runtime ownership'
  }
  $script:SmokeRunCounter++
  $sinkPrefix = Join-Path $script:SmokeSinkDirectory ("command-{0:D3}" -f $script:SmokeRunCounter)
  $stdoutPath = $sinkPrefix + '.stdout'
  $stderrPath = $sinkPrefix + '.stderr'
  $process = $null
  $exitCode = 1
  $captured = ''
  $cleanupFailed = $false
  try {
    $launch = Get-SmokeProcessLaunch -Source $commands[0].Source -ArgumentList $ArgumentList -Stage $Stage
    $parameters = @{
      FilePath = $launch.FilePath
      ArgumentList = $launch.ArgumentString
      PassThru = $true
      WindowStyle = 'Hidden'
      RedirectStandardOutput = $stdoutPath
      RedirectStandardError = $stderrPath
    }
    if ($WorkingDirectory.Length -ne 0) {
      $parameters['WorkingDirectory'] = $WorkingDirectory
    }
    $process = Start-Process @parameters
    # Retain the process association before a fast redirected child can exit;
    # otherwise Windows PowerShell may expose a null ExitCode.
    try {
      $null = $process.Handle
    } catch {
      Stop-Smoke -ExitCode 1 -Stage $Stage
    }
    if (-not $process.WaitForExit($TimeoutSeconds * 1000)) {
      $exitCode = 124
      try {
        Stop-SmokeOwnedProcessTree -Process $process -SinkPrefix $sinkPrefix
      } catch {
        $cleanupFailed = $true
      }
    } else {
      $exitCode = Get-SmokeCompletedProcessExitCode -Process $process -Stage $Stage
    }
    if ($exitCode -eq 0 -and $CaptureOutput) {
      if ([System.IO.FileInfo]::new($stdoutPath).Length -gt 4096) {
        Stop-Smoke -ExitCode 1 -Stage $Stage
      }
      $captured = [System.IO.File]::ReadAllText($stdoutPath, [System.Text.Encoding]::UTF8).Trim()
    }
  } catch {
    if ($_.Exception.Data.Contains('Stage')) {
      throw
    }
    Stop-Smoke -ExitCode 127 -Stage $Stage
  } finally {
    if ($null -ne $process) {
      try {
        $process.Dispose()
      } catch {
        $cleanupFailed = $true
      }
    }
    if (-not (Remove-SmokeSinkFiles -Paths @($stdoutPath, $stderrPath))) {
      $cleanupFailed = $true
    }
  }

  if ($exitCode -ne 0) {
    Stop-Smoke -ExitCode $exitCode -Stage $Stage
  }
  if ($cleanupFailed) {
    Stop-Smoke -ExitCode 1 -Stage $Stage
  }
  if ($CaptureOutput) {
    return $captured
  }
}

function Import-SmokeEnvironment {
  param([Parameter(Mandatory)][string]$Path)

  $requiredNames = @(
    'TALENRO_HTTP_ADDRESS',
    'TALENRO_METRICS_ADDRESS',
    'TALENRO_ALLOW_PUBLIC_METRICS',
    'TALENRO_DATABASE_URL',
    'TALENRO_REDIS_ADDRESS',
    'TALENRO_NATS_URL',
    'TALENRO_ALLOW_PUBLIC_HTTP',
    'TALENRO_PROFILE',
    'TALENRO_PUBLIC_BASE_URL',
    'TALENRO_PRIMARY_BUNDLE_BASE_URL',
    'TALENRO_MIRROR_A_BASE_URL',
    'TALENRO_MIRROR_B_BASE_URL',
    'TALENRO_WEBAUTHN_RP_ID',
    'TALENRO_WEBAUTHN_ORIGINS',
    'TALENRO_EMAIL_VERIFICATION_MODE',
    'TALENRO_SIGNER_PROVIDER',
    'TALENRO_FIELD_PROTECTOR_PROVIDER',
    'TALENRO_EMAIL_PROVIDER',
    'TALENRO_ERROR_REPORTER_PROVIDER',
    'TALENRO_REQUEST_DEADLINE',
    'TALENRO_REDIS_TIMEOUT',
    'TALENRO_SIGNER_TIMEOUT',
    'TALENRO_ERROR_REPORT_TIMEOUT',
    'TALENRO_REDIS_DOWN_AFTER_FAILURES',
    'TALENRO_REDIS_RECOVER_AFTER_SUCCESSES',
    'TALENRO_OUTBOX_DEGRADED_BACKLOG',
    'TALENRO_OUTBOX_DOWN_BACKLOG',
    'TALENRO_OUTBOX_DEGRADED_AGE',
    'TALENRO_OUTBOX_DOWN_AGE',
    'TALENRO_ERROR_REPORT_QUEUE',
    'TALENRO_ERROR_REPORT_BATCH',
    'TALENRO_CLOCK_SKEW',
    'TALENRO_LOGIN_RATE_LIMIT',
    'TALENRO_LOGIN_RATE_WINDOW',
    'TALENRO_DELIVERY_RATE_LIMIT',
    'TALENRO_DELIVERY_RATE_WINDOW',
    'TALENRO_CHALLENGE_RATE_LIMIT',
    'TALENRO_CHALLENGE_RATE_WINDOW',
    'TALENRO_SENSITIVE_LOOKUP_KEY_B64',
    'TALENRO_SENSITIVE_ENCRYPTION_KEY_B64',
    'TALENRO_LOCAL_ROOT_SIGNING_SEED_B64',
    'TALENRO_LOCAL_CONFIG_SIGNING_SEED_B64'
  )
  $allowed = [System.Collections.Generic.HashSet[string]]::new([System.StringComparer]::Ordinal)
  foreach ($name in $requiredNames) {
    $null = $allowed.Add($name)
  }
  $parsed = @{}

  try {
    $lines = [System.IO.File]::ReadAllLines($Path, [System.Text.Encoding]::UTF8)
  } catch {
    Stop-Smoke -ExitCode 2 -Stage 'smoke: environment'
  }

  foreach ($line in $lines) {
    if ($line.Length -eq 0 -or $line.StartsWith('#', [System.StringComparison]::Ordinal)) {
      continue
    }
    if ($line -notmatch '^(?<name>[A-Z][A-Z0-9_]*)=(?<value>[A-Za-z0-9._:/?@=+\-]+)$') {
      Stop-Smoke -ExitCode 2 -Stage 'smoke: environment'
    }

    $name = $Matches.name
    $value = $Matches.value
    if (-not $allowed.Contains($name) -or $parsed.ContainsKey($name)) {
      Stop-Smoke -ExitCode 2 -Stage 'smoke: environment'
    }
    $parsed[$name] = $value
  }

  foreach ($name in $requiredNames) {
    if (-not $parsed.ContainsKey($name)) {
      Stop-Smoke -ExitCode 2 -Stage 'smoke: environment'
    }
  }

  try {
    foreach ($prefix in @('TALENRO_', 'C11_', 'COMPOSE_')) {
      $ambientNames = @([TalenroSmokeEnvironment]::NamesWithPrefix($prefix))
      foreach ($ambientName in $ambientNames) {
        [System.Environment]::SetEnvironmentVariable($ambientName, $null, 'Process')
      }
    }
  } catch {
    Stop-Smoke -ExitCode 2 -Stage 'smoke: environment isolation'
  }
  foreach ($name in $requiredNames) {
    [System.Environment]::SetEnvironmentVariable($name, [string]$parsed[$name], 'Process')
  }

  return $parsed
}

function New-SmokeBuildDirectory {
  $tempRoot = [System.IO.Path]::GetFullPath([System.IO.Path]::GetTempPath()).TrimEnd('\')
  try {
    $leaf = 'talenro-smoke-c11-' + [System.Guid]::NewGuid().ToString('N')
    $directory = Join-Path $tempRoot $leaf
    $null = [System.IO.Directory]::CreateDirectory($directory)
    return $directory
  } catch {
    Stop-Smoke -ExitCode 1 -Stage 'smoke: build directory'
  }
}

function Start-SmokeProcess {
  param(
    [Parameter(Mandatory)][string]$FilePath,
    [Parameter(Mandatory)][string]$WorkingDirectory,
    [Parameter(Mandatory)][string]$StandardOutputPath,
    [Parameter(Mandatory)][string]$StandardErrorPath,
    [Parameter(Mandatory)][string]$Stage
  )

  try {
    $process = Start-Process -FilePath $FilePath -WorkingDirectory $WorkingDirectory -PassThru -WindowStyle Hidden -RedirectStandardOutput $StandardOutputPath -RedirectStandardError $StandardErrorPath
    try {
      $null = $process.Handle
    } catch {
      Stop-Smoke -ExitCode 1 -Stage $Stage
    }
    return $process
  } catch {
    if ($_.Exception.Data.Contains('Stage')) {
      throw
    }
    Stop-Smoke -ExitCode 127 -Stage $Stage
  }
}

function Start-SmokeMirrorProcess {
  param(
    [Parameter(Mandatory)][string]$FilePath,
    [Parameter(Mandatory)][string]$WorkingDirectory,
    [Parameter(Mandatory)][string]$StandardOutputPath,
    [Parameter(Mandatory)][string]$StandardErrorPath,
    [Parameter(Mandatory)][string]$DatabaseURL,
    [Parameter(Mandatory)][string]$HTTPAddress,
    [Parameter(Mandatory)][string]$Stage
  )

  $process = $null
  $standardOutput = $null
  $standardError = $null
  $startFailureCode = 127
  try {
    $standardOutput = [System.IO.FileStream]::new(
      $StandardOutputPath,
      [System.IO.FileMode]::Create,
      [System.IO.FileAccess]::Write,
      [System.IO.FileShare]::Read
    )
    $standardError = [System.IO.FileStream]::new(
      $StandardErrorPath,
      [System.IO.FileMode]::Create,
      [System.IO.FileAccess]::Write,
      [System.IO.FileShare]::Read
    )
    $startInfo = [System.Diagnostics.ProcessStartInfo]::new()
    $startInfo.FileName = $FilePath
    $startInfo.WorkingDirectory = $WorkingDirectory
    $startInfo.UseShellExecute = $false
    $startInfo.CreateNoWindow = $true
    $startInfo.RedirectStandardOutput = $true
    $startInfo.RedirectStandardError = $true
    $startInfo.EnvironmentVariables.Clear()
    foreach ($name in @('SystemRoot', 'WINDIR', 'TEMP', 'TMP')) {
      $value = [System.Environment]::GetEnvironmentVariable($name, 'Process')
      if (-not [string]::IsNullOrEmpty($value)) {
        $startInfo.EnvironmentVariables[$name] = $value
      }
    }
    $startInfo.EnvironmentVariables['TALENRO_DATABASE_URL'] = $DatabaseURL
    $startInfo.EnvironmentVariables['TALENRO_HTTP_ADDRESS'] = $HTTPAddress

    $process = [System.Diagnostics.Process]::new()
    $process.StartInfo = $startInfo
    if (-not $process.Start()) {
      throw [System.InvalidOperationException]::new('process start failed')
    }
    $startFailureCode = 1
    $null = $process.Handle
    $standardOutputTask = $process.StandardOutput.BaseStream.CopyToAsync($standardOutput)
    $standardErrorTask = $process.StandardError.BaseStream.CopyToAsync($standardError)
    $script:SmokeMirrorIO.Add([int]$process.Id, [pscustomobject]@{
      StandardOutputTask = $standardOutputTask
      StandardErrorTask = $standardErrorTask
      StandardOutput = $standardOutput
      StandardError = $standardError
    })
    return $process
  } catch {
    if ($null -ne $process) {
      try {
        if (-not $process.HasExited) {
          Stop-Process -Id $process.Id -Force -ErrorAction SilentlyContinue
          $null = $process.WaitForExit(5000)
        }
      } catch {
        # The fixed, sanitized start failure below remains authoritative.
      }
      try {
        $process.Dispose()
      } catch {
        # The fixed, sanitized start failure below remains authoritative.
      }
    }
    if ($null -ne $standardOutput) {
      try {
        $standardOutput.Dispose()
      } catch {
        # The fixed, sanitized start failure below remains authoritative.
      }
    }
    if ($null -ne $standardError) {
      try {
        $standardError.Dispose()
      } catch {
        # The fixed, sanitized start failure below remains authoritative.
      }
    }
    Stop-Smoke -ExitCode $startFailureCode -Stage $Stage
  }
}

function Complete-SmokeMirrorProcessIO {
  param(
    [Parameter(Mandatory)][System.Diagnostics.Process]$Process,
    [Parameter(Mandatory)][string]$Stage
  )

  $entry = $null
  if (-not $script:SmokeMirrorIO.TryGetValue([int]$Process.Id, [ref]$entry)) {
    return
  }
  $failed = $false
  foreach ($task in @($entry.StandardOutputTask, $entry.StandardErrorTask)) {
    try {
      if (-not $task.Wait(5000) -or $task.IsFaulted) {
        $failed = $true
      }
    } catch {
      $failed = $true
    }
  }
  foreach ($stream in @($entry.StandardOutput, $entry.StandardError)) {
    try {
      $stream.Flush()
    } catch {
      $failed = $true
    } finally {
      try {
        $stream.Dispose()
      } catch {
        $failed = $true
      }
    }
  }
  $null = $script:SmokeMirrorIO.Remove([int]$Process.Id)
  if ($failed) {
    Stop-Smoke -ExitCode 1 -Stage $Stage
  }
}

function New-SmokeComposeOverride {
  param([Parameter(Mandatory)][string]$Directory)

  $path = Join-Path $Directory 'compose.ephemeral.yaml'
  $natsConfigPath = Join-Path $Directory 'nats.conf'
  $body = @'
services:
  postgres:
    image: postgres:18.4-alpine3.23
    environment:
      POSTGRES_DB: talenro
      POSTGRES_USER: talenro
      POSTGRES_PASSWORD: talenro_dev
    ports:
      - target: 5432
        published: "0"
        host_ip: 127.0.0.1
        protocol: tcp
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U talenro -d talenro"]
      interval: 2s
      timeout: 2s
      retries: 20
    volumes:
      - type: tmpfs
        target: /var/lib/postgresql/18/docker
  redis:
    image: redis:8.8.1-alpine3.23
    command: ["redis-server", "--appendonly", "yes", "--save", "60", "1"]
    ports:
      - target: 6379
        published: "0"
        host_ip: 127.0.0.1
        protocol: tcp
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 2s
      timeout: 2s
      retries: 20
    volumes:
      - type: tmpfs
        target: /data
  nats:
    image: nats:2.14.3-alpine3.22
    command: ["-c", "/etc/nats/nats.conf"]
    ports:
      - target: 4222
        published: "0"
        host_ip: 127.0.0.1
        protocol: tcp
    healthcheck:
      test: ["CMD", "wget", "-q", "-O", "-", "http://127.0.0.1:8222/healthz?js-enabled-only=true"]
      interval: 2s
      timeout: 2s
      retries: 20
    volumes:
      - type: bind
        source: ./nats.conf
        target: /etc/nats/nats.conf
        read_only: true
      - type: tmpfs
        target: /data
'@
  $natsConfig = @'
server_name: talenro-c11
http: 8222

jetstream {
  store_dir: "/data/jetstream"
  max_mem_store: 256MB
  max_file_store: 1GB
}
'@
  try {
    [System.IO.File]::WriteAllText($natsConfigPath, $natsConfig, [System.Text.UTF8Encoding]::new($false))
    [System.IO.File]::WriteAllText($path, $body, [System.Text.UTF8Encoding]::new($false))
  } catch {
    Stop-Smoke -ExitCode 1 -Stage 'smoke: dependency ownership'
  }
  return $path
}

function Assert-SmokeComposeOwnership {
  param(
    [Parameter(Mandatory)][string[]]$ComposeArguments,
    [Parameter(Mandatory)][string]$Project,
    [Parameter(Mandatory)][ValidateSet('postgres', 'redis', 'nats')][string]$Service
  )

  $containerID = Invoke-QuietExternal -Stage 'smoke: dependency ownership' -FilePath 'docker' -ArgumentList (
    $ComposeArguments + @('ps', '--quiet', '--no-trunc', $Service)
  ) -TimeoutSeconds 30 -CaptureOutput
  if ($containerID -notmatch '^[0-9a-f]{64}$') {
    Stop-Smoke -ExitCode 1 -Stage 'smoke: dependency ownership'
  }
  $owner = Invoke-QuietExternal -Stage 'smoke: dependency ownership' -FilePath 'docker' -ArgumentList @(
    'inspect', '--type', 'container', '--format', '{{ index .Config.Labels `com.docker.compose.project` }}', $containerID
  ) -TimeoutSeconds 30 -CaptureOutput
  if (-not [string]::Equals($owner, $Project, [System.StringComparison]::Ordinal)) {
    Stop-Smoke -ExitCode 1 -Stage 'smoke: dependency ownership'
  }
  $ownedService = Invoke-QuietExternal -Stage 'smoke: dependency ownership' -FilePath 'docker' -ArgumentList @(
    'inspect', '--type', 'container', '--format', '{{ index .Config.Labels `com.docker.compose.service` }}', $containerID
  ) -TimeoutSeconds 30 -CaptureOutput
  if (-not [string]::Equals($ownedService, $Service, [System.StringComparison]::Ordinal)) {
    Stop-Smoke -ExitCode 1 -Stage 'smoke: dependency ownership'
  }
}

function Resolve-SmokeNetworkFaultTarget {
  param(
    [Parameter(Mandatory)][string[]]$ComposeArguments,
    [Parameter(Mandatory)][string]$Project,
    [Parameter(Mandatory)][ValidateSet('postgres', 'redis', 'nats')][string]$Service
  )

  $stage = switch ($Service) {
    'postgres' { 'smoke: postgres network ownership' }
    'redis' { 'smoke: redis network ownership' }
    'nats' { 'smoke: nats network ownership' }
  }
  if ($Project -notmatch '^talenro-c11-smoke-[0-9a-f]{12}$') {
    Stop-Smoke -ExitCode 1 -Stage $stage
  }
  $containerID = Invoke-QuietExternal -Stage $stage -FilePath 'docker' -ArgumentList (
    $ComposeArguments + @('ps', '--quiet', '--no-trunc', $Service)
  ) -TimeoutSeconds 30 -CaptureOutput
  if ($containerID -notmatch '^[0-9a-f]{64}$') {
    Stop-Smoke -ExitCode 1 -Stage $stage
  }
  $projectLabel = Invoke-QuietExternal -Stage $stage -FilePath 'docker' -ArgumentList @(
    'inspect', '--type', 'container', '--format', '{{ index .Config.Labels `com.docker.compose.project` }}', $containerID
  ) -TimeoutSeconds 30 -CaptureOutput
  $serviceLabel = Invoke-QuietExternal -Stage $stage -FilePath 'docker' -ArgumentList @(
    'inspect', '--type', 'container', '--format', '{{ index .Config.Labels `com.docker.compose.service` }}', $containerID
  ) -TimeoutSeconds 30 -CaptureOutput
  if (
    -not [string]::Equals($projectLabel, $Project, [System.StringComparison]::Ordinal) -or
    -not [string]::Equals($serviceLabel, $Service, [System.StringComparison]::Ordinal)
  ) {
    Stop-Smoke -ExitCode 1 -Stage $stage
  }

  $networkID = Invoke-QuietExternal -Stage $stage -FilePath 'docker' -ArgumentList @(
    'network', 'ls', '--quiet', '--no-trunc',
    '--filter', ('label=com.docker.compose.project=' + $Project)
  ) -TimeoutSeconds 30 -CaptureOutput
  if ($networkID -notmatch '^[0-9a-f]{64}$') {
    Stop-Smoke -ExitCode 1 -Stage $stage
  }
  $networkProjectLabel = Invoke-QuietExternal -Stage $stage -FilePath 'docker' -ArgumentList @(
    'network', 'inspect', '--format', '{{ index .Labels `com.docker.compose.project` }}', $networkID
  ) -TimeoutSeconds 30 -CaptureOutput
  $networkNameLabel = Invoke-QuietExternal -Stage $stage -FilePath 'docker' -ArgumentList @(
    'network', 'inspect', '--format', '{{ index .Labels `com.docker.compose.network` }}', $networkID
  ) -TimeoutSeconds 30 -CaptureOutput
  if (
    -not [string]::Equals($networkProjectLabel, $Project, [System.StringComparison]::Ordinal) -or
    -not [string]::Equals($networkNameLabel, 'default', [System.StringComparison]::Ordinal)
  ) {
    Stop-Smoke -ExitCode 1 -Stage $stage
  }

  return [pscustomobject]@{ Service = $Service; ContainerID = $containerID; NetworkID = $networkID }
}

function Disconnect-SmokeComposeNetwork {
  param(
    [Parameter(Mandatory)][string[]]$ComposeArguments,
    [Parameter(Mandatory)][string]$Project,
    [Parameter(Mandatory)][ValidateSet('postgres', 'redis', 'nats')][string]$Service
  )

  $stage = switch ($Service) {
    'postgres' { 'smoke: postgres network disconnect' }
    'redis' { 'smoke: redis network disconnect' }
    'nats' { 'smoke: nats network disconnect' }
  }
  $target = Resolve-SmokeNetworkFaultTarget -ComposeArguments $ComposeArguments -Project $Project -Service $Service
  Invoke-QuietExternal -Stage $stage -FilePath 'docker' -ArgumentList @(
    'network', 'disconnect', [string]$target.NetworkID, [string]$target.ContainerID
  ) -TimeoutSeconds 60
  return $target
}

function Connect-SmokeComposeNetwork {
  param([Parameter(Mandatory)]$Target)

  $service = [string]$Target.Service
  $stage = switch ($service) {
    'postgres' { 'smoke: postgres network connect' }
    'redis' { 'smoke: redis network connect' }
    'nats' { 'smoke: nats network connect' }
    default { Stop-Smoke -ExitCode 1 -Stage 'smoke: network connect ownership' }
  }
  $containerID = [string]$Target.ContainerID
  $networkID = [string]$Target.NetworkID
  if ($containerID -notmatch '^[0-9a-f]{64}$' -or $networkID -notmatch '^[0-9a-f]{64}$') {
    Stop-Smoke -ExitCode 1 -Stage $stage
  }
  Invoke-QuietExternal -Stage $stage -FilePath 'docker' -ArgumentList @(
    'network', 'connect', $networkID, $containerID
  ) -TimeoutSeconds 60
}

function Get-SmokeComposePort {
  param(
    [Parameter(Mandatory)][string[]]$ComposeArguments,
    [Parameter(Mandatory)][string]$Service,
    [Parameter(Mandatory)][int]$ContainerPort
  )

  $endpoint = Invoke-QuietExternal -Stage 'smoke: dependency endpoints' -FilePath 'docker' -ArgumentList (
    $ComposeArguments + @('port', $Service, [string]$ContainerPort)
  ) -TimeoutSeconds 30 -CaptureOutput
  if ($endpoint -notmatch '^127\.0\.0\.1:(?<port>[1-9][0-9]{0,4})$') {
    Stop-Smoke -ExitCode 1 -Stage 'smoke: dependency endpoints'
  }
  $port = [int]$Matches.port
  if ($port -gt 65535) {
    Stop-Smoke -ExitCode 1 -Stage 'smoke: dependency endpoints'
  }
  return $port
}

function Remove-SmokeBuildDirectory {
  param(
    [Parameter(Mandatory)][string]$Directory,
    [Parameter(Mandatory)][string[]]$Files
  )

  $tempRoot = [System.IO.Path]::GetFullPath([System.IO.Path]::GetTempPath()).TrimEnd('\')
  $fullDirectory = [System.IO.Path]::GetFullPath($Directory).TrimEnd('\')
  $leaf = [System.IO.Path]::GetFileName($fullDirectory)
  if (
    [System.IO.Path]::GetDirectoryName($fullDirectory) -ne $tempRoot -or
    $leaf -notmatch '^talenro-smoke-c11-[0-9a-f]{32}$'
  ) {
    Stop-Smoke -ExitCode 1 -Stage 'smoke: build cleanup validation'
  }

  try {
    foreach ($file in $Files) {
      if ([string]::IsNullOrEmpty($file)) {
        continue
      }
      $fullFile = [System.IO.Path]::GetFullPath($file)
      if ([System.IO.Path]::GetDirectoryName($fullFile) -ne $fullDirectory) {
        Stop-Smoke -ExitCode 1 -Stage 'smoke: build cleanup validation'
      }
      if ([System.IO.File]::Exists($fullFile)) {
        [System.IO.File]::Delete($fullFile)
      }
    }
    if ([System.IO.Directory]::Exists($fullDirectory)) {
      [System.IO.Directory]::Delete($fullDirectory, $false)
    }
  } catch {
    if ($_.Exception.Data.Contains('Stage')) {
      throw
    }
    Stop-Smoke -ExitCode 1 -Stage 'smoke: build cleanup'
  }
}

function Get-HTTPStatus {
  param(
    [Parameter(Mandatory)][System.Net.Http.HttpClient]$Client,
    [Parameter(Mandatory)][string]$URL,
    [Parameter(Mandatory)][int]$TimeoutMilliseconds
  )

  $cancellation = [System.Threading.CancellationTokenSource]::new($TimeoutMilliseconds)
  $response = $null
  try {
    $response = $Client.GetAsync(
      $URL,
      [System.Net.Http.HttpCompletionOption]::ResponseHeadersRead,
      $cancellation.Token
    ).GetAwaiter().GetResult()
    return [int]$response.StatusCode
  } finally {
    if ($null -ne $response) {
      $response.Dispose()
    }
    $cancellation.Dispose()
  }
}

function Wait-HTTPStatus {
  param(
    [Parameter(Mandatory)][System.Net.Http.HttpClient]$Client,
    [Parameter(Mandatory)][System.Diagnostics.Process]$Process,
    [Parameter(Mandatory)][string]$URL,
    [Parameter(Mandatory)][int]$Expected,
    [Parameter(Mandatory)][int]$Seconds,
    [Parameter(Mandatory)][string]$Stage,
    [Parameter(Mandatory)][string]$ProcessStage
  )

  $timer = [System.Diagnostics.Stopwatch]::StartNew()
  $limitMilliseconds = $Seconds * 1000
  while ($timer.ElapsedMilliseconds -lt $limitMilliseconds) {
    if ($Process.HasExited) {
      $processExitCode = Get-SmokeCompletedProcessExitCode -Process $Process -Stage $ProcessStage
      if ($processExitCode -eq 0) {
        $processExitCode = 1
      }
      Stop-Smoke -ExitCode $processExitCode -Stage $ProcessStage
    }

    $remaining = $limitMilliseconds - [int]$timer.ElapsedMilliseconds
    $requestTimeout = [Math]::Max(1, [Math]::Min(3000, $remaining))
    try {
      if ((Get-HTTPStatus -Client $Client -URL $URL -TimeoutMilliseconds $requestTimeout) -eq $Expected) {
        return
      }
    } catch {
      # Connection and timeout details are deliberately discarded.
    }

    $remaining = $limitMilliseconds - [int]$timer.ElapsedMilliseconds
    if ($remaining -gt 0) {
      Start-Sleep -Milliseconds ([Math]::Min(250, $remaining))
    }
  }

  Stop-Smoke -ExitCode 1 -Stage $Stage
}

function Assert-SanitizedArtifacts {
  param(
    [Parameter(Mandatory)][string[]]$Paths,
    [Parameter(Mandatory)][string[]]$Forbidden
  )

  foreach ($path in $Paths) {
    try {
      $body = [System.IO.File]::ReadAllText($path, [System.Text.Encoding]::UTF8)
    } catch {
      Stop-Smoke -ExitCode 1 -Stage 'smoke: artifact scan'
    }
    foreach ($value in $Forbidden) {
      if ($value.Length -ne 0 -and $body.IndexOf($value, [System.StringComparison]::Ordinal) -ge 0) {
        Stop-Smoke -ExitCode 1 -Stage 'smoke: artifact privacy'
      }
    }
  }
}

$repoRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..')).Path
$composeArguments = @()
$composeProject = $null
$composeOverride = $null
$natsConfig = $null
$composeTouched = $false
$controlAPI = $null
$mirrorA = $null
$mirrorB = $null
$buildDirectory = $null
$controlAPIBinary = $null
$mirrorBinary = $null
$conformanceBinary = $null
$artifactPaths = @()
$client = $null
$primaryFailure = $null
$cleanupFailure = $null

try {
  $localEnvironment = Import-SmokeEnvironment -Path (Join-Path $repoRoot '.env.example')
  if (
    [string]$localEnvironment['TALENRO_HTTP_ADDRESS'] -notmatch '^127\.0\.0\.1:[1-9][0-9]{0,4}$' -or
    [string]$localEnvironment['TALENRO_METRICS_ADDRESS'] -notmatch '^127\.0\.0\.1:[1-9][0-9]{0,4}$' -or
    [string]$localEnvironment['TALENRO_REDIS_ADDRESS'] -notmatch '^127\.0\.0\.1:[1-9][0-9]{0,4}$' -or
    [string]$localEnvironment['TALENRO_NATS_URL'] -notmatch '^nats://127\.0\.0\.1:[1-9][0-9]{0,4}$' -or
    [string]$localEnvironment['TALENRO_DATABASE_URL'] -notmatch '^postgres://[^@/]+@127\.0\.0\.1:[1-9][0-9]{0,4}/[^?]+\?sslmode=disable$' -or
    [string]$localEnvironment['TALENRO_ALLOW_PUBLIC_METRICS'] -ne 'false' -or
    [string]$localEnvironment['TALENRO_ALLOW_PUBLIC_HTTP'] -ne 'false' -or
    [string]$localEnvironment['TALENRO_PROFILE'] -ne 'local' -or
    [string]$localEnvironment['TALENRO_PUBLIC_BASE_URL'] -ne 'http://localhost:8080' -or
    [string]$localEnvironment['TALENRO_PRIMARY_BUNDLE_BASE_URL'] -ne 'http://localhost:8080' -or
    [string]$localEnvironment['TALENRO_MIRROR_A_BASE_URL'] -ne 'http://localhost:8081' -or
    [string]$localEnvironment['TALENRO_MIRROR_B_BASE_URL'] -ne 'http://localhost:8082' -or
    [string]$localEnvironment['TALENRO_WEBAUTHN_RP_ID'] -ne 'localhost' -or
    [string]$localEnvironment['TALENRO_WEBAUTHN_ORIGINS'] -ne 'http://localhost:8080' -or
    [string]$localEnvironment['TALENRO_EMAIL_VERIFICATION_MODE'] -ne 'disabled' -or
    [string]$localEnvironment['TALENRO_SIGNER_PROVIDER'] -ne 'local' -or
    [string]$localEnvironment['TALENRO_FIELD_PROTECTOR_PROVIDER'] -ne 'local' -or
    [string]$localEnvironment['TALENRO_EMAIL_PROVIDER'] -ne 'local' -or
    [string]$localEnvironment['TALENRO_ERROR_REPORTER_PROVIDER'] -ne 'discard' -or
    [string]$localEnvironment['TALENRO_REQUEST_DEADLINE'] -ne '5s' -or
    [string]$localEnvironment['TALENRO_REDIS_TIMEOUT'] -ne '250ms' -or
    [string]$localEnvironment['TALENRO_SIGNER_TIMEOUT'] -ne '2s' -or
    [string]$localEnvironment['TALENRO_ERROR_REPORT_TIMEOUT'] -ne '1s' -or
    [string]$localEnvironment['TALENRO_REDIS_DOWN_AFTER_FAILURES'] -ne '3' -or
    [string]$localEnvironment['TALENRO_REDIS_RECOVER_AFTER_SUCCESSES'] -ne '2' -or
    [string]$localEnvironment['TALENRO_OUTBOX_DEGRADED_BACKLOG'] -ne '1000' -or
    [string]$localEnvironment['TALENRO_OUTBOX_DOWN_BACKLOG'] -ne '10000' -or
    [string]$localEnvironment['TALENRO_OUTBOX_DEGRADED_AGE'] -ne '60s' -or
    [string]$localEnvironment['TALENRO_OUTBOX_DOWN_AGE'] -ne '300s' -or
    [string]$localEnvironment['TALENRO_ERROR_REPORT_QUEUE'] -ne '100' -or
    [string]$localEnvironment['TALENRO_ERROR_REPORT_BATCH'] -ne '20' -or
    [string]$localEnvironment['TALENRO_CLOCK_SKEW'] -ne '120s' -or
    [string]$localEnvironment['TALENRO_LOGIN_RATE_LIMIT'] -ne '10' -or
    [string]$localEnvironment['TALENRO_LOGIN_RATE_WINDOW'] -ne '15m' -or
    [string]$localEnvironment['TALENRO_DELIVERY_RATE_LIMIT'] -ne '5' -or
    [string]$localEnvironment['TALENRO_DELIVERY_RATE_WINDOW'] -ne '1h' -or
    [string]$localEnvironment['TALENRO_CHALLENGE_RATE_LIMIT'] -ne '20' -or
    [string]$localEnvironment['TALENRO_CHALLENGE_RATE_WINDOW'] -ne '5m' -or
    [string]$localEnvironment['TALENRO_SENSITIVE_LOOKUP_KEY_B64'] -notmatch '^[A-Za-z0-9_-]{43}$' -or
    [string]$localEnvironment['TALENRO_SENSITIVE_ENCRYPTION_KEY_B64'] -notmatch '^[A-Za-z0-9_-]{43}$' -or
    [string]$localEnvironment['TALENRO_LOCAL_ROOT_SIGNING_SEED_B64'] -notmatch '^[A-Za-z0-9_-]{43}$' -or
    [string]$localEnvironment['TALENRO_LOCAL_CONFIG_SIGNING_SEED_B64'] -notmatch '^[A-Za-z0-9_-]{43}$'
  ) {
    Stop-Smoke -ExitCode 2 -Stage 'smoke: local environment'
  }
  $httpBase = 'http://' + [string]$localEnvironment['TALENRO_HTTP_ADDRESS']

  $handler = [System.Net.Http.HttpClientHandler]::new()
  $handler.UseProxy = $false
  $client = [System.Net.Http.HttpClient]::new($handler, $true)
  $client.Timeout = [System.Threading.Timeout]::InfiniteTimeSpan

  $buildDirectory = New-SmokeBuildDirectory
  $script:SmokeSinkDirectory = $buildDirectory
  $composeProject = 'talenro-c11-smoke-' + [System.Guid]::NewGuid().ToString('N').Substring(0, 12)
  if ($composeProject -notmatch '^talenro-c11-smoke-[0-9a-f]{12}$') {
    Stop-Smoke -ExitCode 1 -Stage 'smoke: dependency ownership'
  }
  $composeOverride = Join-Path $buildDirectory 'compose.ephemeral.yaml'
  $natsConfig = Join-Path $buildDirectory 'nats.conf'
  $null = New-SmokeComposeOverride -Directory $buildDirectory
  $composeArguments = @(
    'compose', '--project-name', $composeProject,
    '-f', $composeOverride
  )
  $controlAPIBinary = Join-Path $buildDirectory 'control-api.exe'
  $mirrorBinary = Join-Path $buildDirectory 'bundle-mirror.exe'
  $conformanceBinary = Join-Path $buildDirectory 'trust-conformance.exe'
  $controlStdout = Join-Path $buildDirectory 'control-api.stdout.sink'
  $controlStderr = Join-Path $buildDirectory 'control-api.stderr.sink'
  $mirrorAStdout = Join-Path $buildDirectory 'mirror-a.stdout.sink'
  $mirrorAStderr = Join-Path $buildDirectory 'mirror-a.stderr.sink'
  $mirrorBStdout = Join-Path $buildDirectory 'mirror-b.stdout.sink'
  $mirrorBStderr = Join-Path $buildDirectory 'mirror-b.stderr.sink'
  $artifactPaths = @($controlStdout, $controlStderr, $mirrorAStdout, $mirrorAStderr, $mirrorBStdout, $mirrorBStderr)

  $composeTouched = $true
  Invoke-QuietExternal -Stage 'smoke: compose up' -FilePath 'docker' -ArgumentList (
    $composeArguments + @('up', '-d', '--wait', '--wait-timeout', '120', 'postgres', 'redis', 'nats')
  ) -TimeoutSeconds 180
  foreach ($service in @('postgres', 'redis', 'nats')) {
    Assert-SmokeComposeOwnership -ComposeArguments $composeArguments -Project $composeProject -Service $service
  }
  $postgresPort = Get-SmokeComposePort -ComposeArguments $composeArguments -Service 'postgres' -ContainerPort 5432
  $redisPort = Get-SmokeComposePort -ComposeArguments $composeArguments -Service 'redis' -ContainerPort 6379
  $natsPort = Get-SmokeComposePort -ComposeArguments $composeArguments -Service 'nats' -ContainerPort 4222
  $localEnvironment['TALENRO_DATABASE_URL'] = "postgres://talenro:talenro_dev@127.0.0.1:${postgresPort}/talenro?sslmode=disable"
  $localEnvironment['TALENRO_REDIS_ADDRESS'] = "127.0.0.1:${redisPort}"
  $localEnvironment['TALENRO_NATS_URL'] = "nats://127.0.0.1:${natsPort}"
  [System.Environment]::SetEnvironmentVariable('TALENRO_DATABASE_URL', [string]$localEnvironment['TALENRO_DATABASE_URL'], 'Process')
  [System.Environment]::SetEnvironmentVariable('TALENRO_REDIS_ADDRESS', [string]$localEnvironment['TALENRO_REDIS_ADDRESS'], 'Process')
  [System.Environment]::SetEnvironmentVariable('TALENRO_NATS_URL', [string]$localEnvironment['TALENRO_NATS_URL'], 'Process')
  [System.Environment]::SetEnvironmentVariable('C11_E2E_COMPOSE_PROJECT', $composeProject, 'Process')
  [System.Environment]::SetEnvironmentVariable('C11_E2E_DATABASE_URL', [string]$localEnvironment['TALENRO_DATABASE_URL'], 'Process')
  [System.Environment]::SetEnvironmentVariable('C11_E2E_REDIS_ADDRESS', [string]$localEnvironment['TALENRO_REDIS_ADDRESS'], 'Process')
  [System.Environment]::SetEnvironmentVariable('C11_E2E_NATS_URL', [string]$localEnvironment['TALENRO_NATS_URL'], 'Process')
  Invoke-QuietExternal -Stage 'smoke: migrations' -FilePath 'go' -ArgumentList @(
    'tool',
    'goose',
    '-dir',
    (Join-Path $repoRoot 'db\migrations'),
    'postgres',
    [string]$localEnvironment['TALENRO_DATABASE_URL'],
    'up'
  ) -TimeoutSeconds 120
  Invoke-QuietExternal -Stage 'smoke: control API build' -FilePath 'go' -WorkingDirectory $repoRoot -ArgumentList @(
    'build',
    '-o',
    $controlAPIBinary,
    (Join-Path $repoRoot 'cmd\control-api')
  )
  Invoke-QuietExternal -Stage 'smoke: mirror build' -FilePath 'go' -WorkingDirectory $repoRoot -ArgumentList @(
    'build', '-o', $mirrorBinary, (Join-Path $repoRoot 'cmd\bundle-mirror')
  )
  Invoke-QuietExternal -Stage 'smoke: conformance build' -FilePath 'go' -WorkingDirectory $repoRoot -ArgumentList @(
    'build', '-o', $conformanceBinary, (Join-Path $repoRoot 'cmd\trust-conformance')
  )

  $controlAPI = Start-SmokeProcess -FilePath $controlAPIBinary -WorkingDirectory $repoRoot -StandardOutputPath $controlStdout -StandardErrorPath $controlStderr -Stage 'smoke: control API start'
  $mirrorA = Start-SmokeMirrorProcess -FilePath $mirrorBinary -WorkingDirectory $repoRoot -StandardOutputPath $mirrorAStdout -StandardErrorPath $mirrorAStderr -DatabaseURL ([string]$localEnvironment['TALENRO_DATABASE_URL']) -HTTPAddress '127.0.0.1:8081' -Stage 'smoke: mirror A start'
  $mirrorB = Start-SmokeMirrorProcess -FilePath $mirrorBinary -WorkingDirectory $repoRoot -StandardOutputPath $mirrorBStdout -StandardErrorPath $mirrorBStderr -DatabaseURL ([string]$localEnvironment['TALENRO_DATABASE_URL']) -HTTPAddress '127.0.0.1:8082' -Stage 'smoke: mirror B start'

  Wait-HTTPStatus -Client $client -Process $controlAPI -URL ($httpBase + '/livez') -Expected 200 -Seconds 10 -Stage 'smoke: initial liveness' -ProcessStage 'smoke: control API'
  Wait-HTTPStatus -Client $client -Process $controlAPI -URL ($httpBase + '/readyz') -Expected 200 -Seconds 10 -Stage 'smoke: initial readiness' -ProcessStage 'smoke: control API'
  Wait-HTTPStatus -Client $client -Process $mirrorA -URL 'http://127.0.0.1:8081/' -Expected 404 -Seconds 10 -Stage 'smoke: mirror A readiness' -ProcessStage 'smoke: mirror A'
  Wait-HTTPStatus -Client $client -Process $mirrorB -URL 'http://127.0.0.1:8082/' -Expected 404 -Seconds 10 -Stage 'smoke: mirror B readiness' -ProcessStage 'smoke: mirror B'

  $previousExternalRuntime = [System.Environment]::GetEnvironmentVariable('C11_E2E_EXTERNAL_RUNTIME', 'Process')
  $previousPrimaryURL = [System.Environment]::GetEnvironmentVariable('C11_E2E_PRIMARY_URL', 'Process')
  $previousMirrorAURL = [System.Environment]::GetEnvironmentVariable('C11_E2E_MIRROR_A_URL', 'Process')
  $previousMirrorBURL = [System.Environment]::GetEnvironmentVariable('C11_E2E_MIRROR_B_URL', 'Process')
  $previousPrimaryOrigin = [System.Environment]::GetEnvironmentVariable('C11_E2E_PRIMARY_ORIGIN', 'Process')
  $previousMirrorAOrigin = [System.Environment]::GetEnvironmentVariable('C11_E2E_MIRROR_A_ORIGIN', 'Process')
  $previousMirrorBOrigin = [System.Environment]::GetEnvironmentVariable('C11_E2E_MIRROR_B_ORIGIN', 'Process')
  $previousConformanceBinary = [System.Environment]::GetEnvironmentVariable('C11_CONFORMANCE_BINARY', 'Process')
  $happyPathFailure = $null
  $environmentRestoreFailed = $false
  try {
    [System.Environment]::SetEnvironmentVariable('C11_E2E_EXTERNAL_RUNTIME', '1', 'Process')
    [System.Environment]::SetEnvironmentVariable('C11_E2E_PRIMARY_URL', 'http://127.0.0.1:8080', 'Process')
    [System.Environment]::SetEnvironmentVariable('C11_E2E_MIRROR_A_URL', 'http://127.0.0.1:8081', 'Process')
    [System.Environment]::SetEnvironmentVariable('C11_E2E_MIRROR_B_URL', 'http://127.0.0.1:8082', 'Process')
    [System.Environment]::SetEnvironmentVariable('C11_E2E_PRIMARY_ORIGIN', 'http://localhost:8080', 'Process')
    [System.Environment]::SetEnvironmentVariable('C11_E2E_MIRROR_A_ORIGIN', 'http://localhost:8081', 'Process')
    [System.Environment]::SetEnvironmentVariable('C11_E2E_MIRROR_B_ORIGIN', 'http://localhost:8082', 'Process')
    [System.Environment]::SetEnvironmentVariable('C11_CONFORMANCE_BINARY', $conformanceBinary, 'Process')
    Invoke-QuietExternal -Stage 'smoke: C1.1 happy path' -FilePath 'go' -WorkingDirectory $repoRoot -ArgumentList @(
      'test', '-tags=e2e', './internal/e2e', '-run', '^TestC11HappyPath$', '-count=1', '-timeout', '10m'
    )
  } catch {
    $happyPathFailure = $_.Exception
  } finally {
    foreach ($entry in @(
      @{ Name = 'C11_E2E_EXTERNAL_RUNTIME'; Value = $previousExternalRuntime },
      @{ Name = 'C11_E2E_PRIMARY_URL'; Value = $previousPrimaryURL },
      @{ Name = 'C11_E2E_MIRROR_A_URL'; Value = $previousMirrorAURL },
      @{ Name = 'C11_E2E_MIRROR_B_URL'; Value = $previousMirrorBURL },
      @{ Name = 'C11_E2E_PRIMARY_ORIGIN'; Value = $previousPrimaryOrigin },
      @{ Name = 'C11_E2E_MIRROR_A_ORIGIN'; Value = $previousMirrorAOrigin },
      @{ Name = 'C11_E2E_MIRROR_B_ORIGIN'; Value = $previousMirrorBOrigin },
      @{ Name = 'C11_CONFORMANCE_BINARY'; Value = $previousConformanceBinary }
    )) {
      try {
        [System.Environment]::SetEnvironmentVariable([string]$entry.Name, $entry.Value, 'Process')
      } catch {
        $environmentRestoreFailed = $true
      }
    }
  }
  if ($null -ne $happyPathFailure) {
    throw $happyPathFailure
  }
  if ($environmentRestoreFailed) {
    Stop-Smoke -ExitCode 1 -Stage 'smoke: e2e environment cleanup'
  }

  $postgresFaultTarget = Disconnect-SmokeComposeNetwork -ComposeArguments $composeArguments -Project $composeProject -Service 'postgres'
  Wait-HTTPStatus -Client $client -Process $controlAPI -URL ($httpBase + '/readyz') -Expected 503 -Seconds 5 -Stage 'smoke: postgres readiness failure' -ProcessStage 'smoke: control API'
  Wait-HTTPStatus -Client $client -Process $controlAPI -URL ($httpBase + '/livez') -Expected 200 -Seconds 2 -Stage 'smoke: postgres liveness' -ProcessStage 'smoke: control API'

  Connect-SmokeComposeNetwork -Target $postgresFaultTarget
  $postgresFaultTarget = $null
  Wait-HTTPStatus -Client $client -Process $controlAPI -URL ($httpBase + '/readyz') -Expected 200 -Seconds 10 -Stage 'smoke: postgres recovery' -ProcessStage 'smoke: control API'

  $redisFaultTarget = Disconnect-SmokeComposeNetwork -ComposeArguments $composeArguments -Project $composeProject -Service 'redis'
  Wait-HTTPStatus -Client $client -Process $controlAPI -URL ($httpBase + '/readyz') -Expected 503 -Seconds 5 -Stage 'smoke: redis readiness failure' -ProcessStage 'smoke: control API'
  Wait-HTTPStatus -Client $client -Process $controlAPI -URL ($httpBase + '/livez') -Expected 200 -Seconds 2 -Stage 'smoke: redis liveness' -ProcessStage 'smoke: control API'
  Connect-SmokeComposeNetwork -Target $redisFaultTarget
  $redisFaultTarget = $null
  Wait-HTTPStatus -Client $client -Process $controlAPI -URL ($httpBase + '/readyz') -Expected 200 -Seconds 10 -Stage 'smoke: redis recovery' -ProcessStage 'smoke: control API'

  $natsFaultTarget = Disconnect-SmokeComposeNetwork -ComposeArguments $composeArguments -Project $composeProject -Service 'nats'
  Wait-HTTPStatus -Client $client -Process $controlAPI -URL ($httpBase + '/livez') -Expected 200 -Seconds 2 -Stage 'smoke: nats liveness' -ProcessStage 'smoke: control API'
  Connect-SmokeComposeNetwork -Target $natsFaultTarget
  $natsFaultTarget = $null
  Wait-HTTPStatus -Client $client -Process $controlAPI -URL ($httpBase + '/readyz') -Expected 200 -Seconds 10 -Stage 'smoke: nats recovery' -ProcessStage 'smoke: control API'

} catch {
  $primaryFailure = Get-SmokeFailure -ErrorRecord $_
} finally {
  foreach ($ownedProcess in @(
    @{ Process = $mirrorB; Stage = 'smoke: mirror B cleanup' },
    @{ Process = $mirrorA; Stage = 'smoke: mirror A cleanup' },
    @{ Process = $controlAPI; Stage = 'smoke: control API cleanup' }
  )) {
    $process = $ownedProcess.Process
    if ($null -eq $process) {
      continue
    }
    try {
      if (-not $process.HasExited) {
        Stop-Process -Id $process.Id -Force -ErrorAction Stop
      }
      if (-not $process.WaitForExit(5000)) {
        Stop-Smoke -ExitCode 1 -Stage ([string]$ownedProcess.Stage)
      }
    } catch {
      if ($null -eq $cleanupFailure) {
        $cleanupFailure = @{ ExitCode = 1; Stage = [string]$ownedProcess.Stage }
      }
    }
    try {
      Complete-SmokeMirrorProcessIO -Process $process -Stage ([string]$ownedProcess.Stage)
    } catch {
      if ($null -eq $cleanupFailure) {
        $cleanupFailure = Get-SmokeFailure -ErrorRecord $_
      }
    }
    try {
      $process.Dispose()
    } catch {
      if ($null -eq $cleanupFailure) {
        $cleanupFailure = @{ ExitCode = 1; Stage = [string]$ownedProcess.Stage }
      }
    }
  }

  if ($artifactPaths.Count -ne 0) {
    try {
      Assert-SanitizedArtifacts -Paths $artifactPaths -Forbidden @(
        [string]$localEnvironment['TALENRO_DATABASE_URL'],
        [string]$localEnvironment['TALENRO_SENSITIVE_LOOKUP_KEY_B64'],
        [string]$localEnvironment['TALENRO_SENSITIVE_ENCRYPTION_KEY_B64'],
        [string]$localEnvironment['TALENRO_LOCAL_ROOT_SIGNING_SEED_B64'],
        [string]$localEnvironment['TALENRO_LOCAL_CONFIG_SIGNING_SEED_B64']
      )
    } catch {
      if ($null -eq $cleanupFailure) {
        $cleanupFailure = Get-SmokeFailure -ErrorRecord $_
      }
    }
  }

  if ($null -ne $client) {
    try {
      $client.Dispose()
    } catch {
      if ($null -eq $cleanupFailure) {
        $cleanupFailure = @{ ExitCode = 1; Stage = 'smoke: HTTP client cleanup' }
      }
    }
  }

  if ($composeTouched) {
    try {
      Invoke-QuietExternal -Stage 'smoke: compose down' -FilePath 'docker' -ArgumentList (
        $composeArguments + @('down', '--remove-orphans', '--timeout', '20')
      ) -TimeoutSeconds 60
    } catch {
      $cleanupFailure = Get-SmokeFailure -ErrorRecord $_
    }
  }

  if ($null -ne $buildDirectory) {
    try {
      Remove-SmokeBuildDirectory -Directory $buildDirectory -Files (@($controlAPIBinary, $mirrorBinary, $conformanceBinary, $composeOverride, $natsConfig) + $artifactPaths)
    } catch {
      if ($null -eq $cleanupFailure) {
        $cleanupFailure = Get-SmokeFailure -ErrorRecord $_
      }
    }
  }
}

$failure = $primaryFailure
if ($null -eq $failure) {
  $failure = $cleanupFailure
}
if ($null -ne $failure) {
  $finalExitCode = 1
  try {
    $candidateExitCode = $failure.ExitCode
    if ($null -ne $candidateExitCode) {
      $convertedExitCode = [int]$candidateExitCode
      if ($convertedExitCode -ne 0) {
        $finalExitCode = $convertedExitCode
      }
    }
  } catch {
    $finalExitCode = 1
  }
  [Console]::Error.WriteLine("$($failure.Stage) failed with exit code ${finalExitCode}.")
  exit $finalExitCode
}

Write-Output 'smoke: C1.1 acceptance passed.'
