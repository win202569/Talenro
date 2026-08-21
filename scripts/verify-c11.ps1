$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$script:C11RunDirectory = $null
$script:C11RunCounter = 0
$script:C11ComposeArguments = @()
$script:C11ComposeTouched = $false

function Stop-C11Verification {
  param(
    [Parameter(Mandatory)][int]$ExitCode,
    [Parameter(Mandatory)][string]$Stage
  )

  if ($ExitCode -eq 0) {
    $ExitCode = 1
  }
  $failure = [System.Exception]::new('sanitized C1.1 verification failure')
  $failure.Data['ExitCode'] = $ExitCode
  $failure.Data['Stage'] = $Stage
  throw $failure
}

function Get-C11CompletedProcessExitCode {
  param(
    [Parameter(Mandatory)][System.Diagnostics.Process]$Process,
    [Parameter(Mandatory)][string]$Stage
  )

  try {
    if (-not $Process.HasExited) {
      Stop-C11Verification -ExitCode 1 -Stage $Stage
    }

    # The process handle is acquired immediately after launch. The process is
    # already known to have exited here, so this overload only completes the
    # state/stream update before ExitCode is read.
    $Process.WaitForExit()
    $Process.Refresh()
    $rawExitCode = $Process.ExitCode
    if ($null -eq $rawExitCode) {
      Stop-C11Verification -ExitCode 1 -Stage $Stage
    }
    return [int]$rawExitCode
  } catch {
    if ($_.Exception.Data.Contains('Stage')) {
      throw
    }
    Stop-C11Verification -ExitCode 1 -Stage $Stage
  }
}

function ConvertTo-C11CommandLineArgument {
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

function Join-C11CommandLine {
  param([Parameter(Mandatory)][string[]]$ArgumentList)

  return (($ArgumentList | ForEach-Object {
    ConvertTo-C11CommandLineArgument -Argument $_
  }) -join ' ')
}

function Get-C11ProcessLaunch {
  param(
    [Parameter(Mandatory)][string]$Source,
    [Parameter(Mandatory)][string[]]$ArgumentList,
    [Parameter(Mandatory)][string]$Stage
  )

  if ([System.IO.Path]::GetExtension($Source) -notin @('.cmd', '.bat')) {
    return @{
      FilePath = $Source
      ArgumentString = (Join-C11CommandLine -ArgumentList $ArgumentList)
    }
  }

  $commandInterpreter = Join-Path $env:SystemRoot 'System32\cmd.exe'
  if (-not [System.IO.File]::Exists($commandInterpreter)) {
    Stop-C11Verification -ExitCode 127 -Stage $Stage
  }

  $commandParts = [System.Collections.Generic.List[string]]::new()
  foreach ($value in @($Source) + @($ArgumentList)) {
    if ($value -match '[\r\n"%!]') {
      Stop-C11Verification -ExitCode 1 -Stage $Stage
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

function Remove-C11SinkFiles {
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

function Stop-C11OwnedProcessTree {
  param(
    [Parameter(Mandatory)][System.Diagnostics.Process]$Process,
    [Parameter(Mandatory)][string]$SinkPrefix
  )

  try {
    if ($Process.HasExited) {
      return
    }
  } catch {
    return
  }

  $taskKill = Join-Path $env:SystemRoot 'System32\taskkill.exe'
  if ([System.IO.File]::Exists($taskKill)) {
    $killOut = $SinkPrefix + '.kill.stdout'
    $killErr = $SinkPrefix + '.kill.stderr'
    $killer = $null
    try {
      $killer = Start-Process -FilePath $taskKill -ArgumentList (
        "/PID $($Process.Id) /T /F"
      ) -PassThru -WindowStyle Hidden -RedirectStandardOutput $killOut -RedirectStandardError $killErr
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
      $null = Remove-C11SinkFiles -Paths @($killOut, $killErr)
    }
  }

  try {
    if (-not $Process.HasExited) {
      Stop-Process -Id $Process.Id -Force -ErrorAction Stop
    }
    if (-not $Process.WaitForExit(5000)) {
      Stop-C11Verification -ExitCode 1 -Stage 'verify-c11: process cleanup'
    }
  } catch {
    if ($_.Exception.Data.Contains('Stage')) {
      throw
    }
    Stop-C11Verification -ExitCode 1 -Stage 'verify-c11: process cleanup'
  }
}

function Invoke-C11External {
  param(
    [Parameter(Mandatory)][string]$Stage,
    [Parameter(Mandatory)][string]$FilePath,
    [Parameter(Mandatory)][string[]]$ArgumentList,
    [string]$WorkingDirectory = '',
    [ValidateRange(1, 4200)][int]$TimeoutSeconds = 600,
    [switch]$CaptureOutput,
    [ValidateRange(1, 4194304)][int]$MaximumCaptureBytes = 4096
  )

  $commands = @(Get-Command -Name $FilePath -CommandType Application -ErrorAction SilentlyContinue)
  if ($commands.Count -eq 0) {
    Stop-C11Verification -ExitCode 127 -Stage $Stage
  }
  if ($null -eq $script:C11RunDirectory) {
    Stop-C11Verification -ExitCode 1 -Stage 'verify-c11: runtime ownership'
  }

  $script:C11RunCounter++
  $sinkPrefix = Join-Path $script:C11RunDirectory ("command-{0:D3}" -f $script:C11RunCounter)
  $stdoutPath = $sinkPrefix + '.stdout'
  $stderrPath = $sinkPrefix + '.stderr'
  $process = $null
  $exitCode = 1
  $captured = ''
  $cleanupFailed = $false

  try {
    $launch = Get-C11ProcessLaunch -Source $commands[0].Source -ArgumentList $ArgumentList -Stage $Stage
    $startParameters = @{
      FilePath = $launch.FilePath
      ArgumentList = $launch.ArgumentString
      PassThru = $true
      WindowStyle = 'Hidden'
      RedirectStandardOutput = $stdoutPath
      RedirectStandardError = $stderrPath
    }
    if ($WorkingDirectory.Length -ne 0) {
      $startParameters['WorkingDirectory'] = $WorkingDirectory
    }
    $process = Start-Process @startParameters
    # Windows PowerShell can lose the association needed for ExitCode if the
    # redirected child exits before Process.Handle has ever been acquired.
    try {
      $null = $process.Handle
    } catch {
      Stop-C11Verification -ExitCode 1 -Stage $Stage
    }
    if (-not $process.WaitForExit($TimeoutSeconds * 1000)) {
      $exitCode = 124
      try {
        Stop-C11OwnedProcessTree -Process $process -SinkPrefix $sinkPrefix
      } catch {
        $cleanupFailed = $true
      }
    } else {
      $exitCode = Get-C11CompletedProcessExitCode -Process $process -Stage $Stage
    }

    if ($exitCode -eq 0 -and $CaptureOutput) {
      $length = [System.IO.FileInfo]::new($stdoutPath).Length
      if ($length -gt $MaximumCaptureBytes) {
        Stop-C11Verification -ExitCode 1 -Stage $Stage
      }
      $captured = [System.IO.File]::ReadAllText($stdoutPath, [System.Text.Encoding]::UTF8).Trim()
    }
  } catch {
    if ($_.Exception.Data.Contains('Stage')) {
      throw
    }
    Stop-C11Verification -ExitCode 127 -Stage $Stage
  } finally {
    if ($null -ne $process) {
      try {
        $process.Dispose()
      } catch {
        $cleanupFailed = $true
      }
    }
    if (-not (Remove-C11SinkFiles -Paths @($stdoutPath, $stderrPath))) {
      $cleanupFailed = $true
    }
  }

  if ($exitCode -ne 0) {
    Stop-C11Verification -ExitCode $exitCode -Stage $Stage
  }
  if ($cleanupFailed) {
    Stop-C11Verification -ExitCode 1 -Stage $Stage
  }
  if ($CaptureOutput) {
    return $captured
  }
}

function Invoke-C11Stage {
  param(
    [Parameter(Mandatory)][string]$Stage,
    [Parameter(Mandatory)][string]$FilePath,
    [Parameter(Mandatory)][string[]]$ArgumentList,
    [string]$WorkingDirectory = '',
    [ValidateRange(1, 4200)][int]$TimeoutSeconds = 600
  )

  Write-Output "verify-c11: $Stage"
  Invoke-C11External -Stage "verify-c11: $Stage" -FilePath $FilePath -ArgumentList $ArgumentList -WorkingDirectory $WorkingDirectory -TimeoutSeconds $TimeoutSeconds
}

function New-C11RunDirectory {
  $tempRoot = [System.IO.Path]::GetFullPath([System.IO.Path]::GetTempPath()).TrimEnd('\')
  $leaf = 'talenro-verify-c11-' + [System.Guid]::NewGuid().ToString('N')
  $directory = Join-Path $tempRoot $leaf
  try {
    $null = [System.IO.Directory]::CreateDirectory($directory)
  } catch {
    Stop-C11Verification -ExitCode 1 -Stage 'verify-c11: runtime ownership'
  }
  $fullDirectory = [System.IO.Path]::GetFullPath($directory).TrimEnd('\')
  if (
    [System.IO.Path]::GetDirectoryName($fullDirectory) -ne $tempRoot -or
    [System.IO.Path]::GetFileName($fullDirectory) -notmatch '^talenro-verify-c11-[0-9a-f]{32}$'
  ) {
    try {
      if ([System.IO.Directory]::Exists($directory)) {
        [System.IO.Directory]::Delete($directory, $false)
      }
    } catch {
      # The fixed, sanitized ownership failure below remains authoritative.
    }
    Stop-C11Verification -ExitCode 1 -Stage 'verify-c11: runtime ownership'
  }
  return $fullDirectory
}

function Remove-C11RunDirectory {
  param([Parameter(Mandatory)][string]$Directory)

  $tempRoot = [System.IO.Path]::GetFullPath([System.IO.Path]::GetTempPath()).TrimEnd('\')
  $fullDirectory = [System.IO.Path]::GetFullPath($Directory).TrimEnd('\')
  if (
    [System.IO.Path]::GetDirectoryName($fullDirectory) -ne $tempRoot -or
    [System.IO.Path]::GetFileName($fullDirectory) -notmatch '^talenro-verify-c11-[0-9a-f]{32}$'
  ) {
    Stop-C11Verification -ExitCode 1 -Stage 'verify-c11: runtime cleanup validation'
  }

  try {
    foreach ($file in [System.IO.Directory]::GetFiles($fullDirectory)) {
      if ([System.IO.Path]::GetDirectoryName([System.IO.Path]::GetFullPath($file)) -ne $fullDirectory) {
        Stop-C11Verification -ExitCode 1 -Stage 'verify-c11: runtime cleanup validation'
      }
      [System.IO.File]::Delete($file)
    }
    if ([System.IO.Directory]::GetDirectories($fullDirectory).Length -ne 0) {
      Stop-C11Verification -ExitCode 1 -Stage 'verify-c11: runtime cleanup validation'
    }
    [System.IO.Directory]::Delete($fullDirectory, $false)
  } catch {
    if ($_.Exception.Data.Contains('Stage')) {
      throw
    }
    Stop-C11Verification -ExitCode 1 -Stage 'verify-c11: runtime cleanup'
  }
}

function Clear-C11AmbientEnvironment {
  $variables = [System.Environment]::GetEnvironmentVariables('Process')
  foreach ($name in $variables.Keys) {
    $variableName = [string]$name
    if ($variableName -match '^(TALENRO_|C11_|COMPOSE_)') {
      [System.Environment]::SetEnvironmentVariable($variableName, $null, 'Process')
    }
  }
}

function Get-C11FileSHA256 {
  param([Parameter(Mandatory)][string]$Path)

  $algorithm = [System.Security.Cryptography.SHA256]::Create()
  $stream = $null
  try {
    $stream = [System.IO.File]::OpenRead($Path)
    return ([System.BitConverter]::ToString($algorithm.ComputeHash($stream))).Replace('-', '').ToLowerInvariant()
  } finally {
    if ($null -ne $stream) {
      $stream.Dispose()
    }
    $algorithm.Dispose()
  }
}

function Get-C11GeneratedSnapshot {
  param([Parameter(Mandatory)][string]$RepoRoot)

  $entries = [System.Collections.Generic.List[string]]::new()
  foreach ($root in @('api', 'gen', 'internal/store')) {
    $absoluteRoot = Join-Path $RepoRoot $root
    if (-not [System.IO.Directory]::Exists($absoluteRoot)) {
      Stop-C11Verification -ExitCode 1 -Stage 'verify-c11: generated inventory'
    }
    foreach ($item in Get-ChildItem -LiteralPath $absoluteRoot -Force -Recurse) {
      if (($item.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) {
        Stop-C11Verification -ExitCode 1 -Stage 'verify-c11: generated inventory'
      }
      $relative = $item.FullName.Substring($RepoRoot.Length + 1).Replace('\', '/')
      if ($item.PSIsContainer) {
        $entries.Add("directory`t$relative")
        continue
      }
      if (-not [System.IO.File]::Exists($item.FullName)) {
        Stop-C11Verification -ExitCode 1 -Stage 'verify-c11: generated inventory'
      }
      $entries.Add("worktree`t$relative`t$(Get-C11FileSHA256 -Path $item.FullName)")
    }
  }
  $array = $entries.ToArray()
  [System.Array]::Sort($array, [System.StringComparer]::Ordinal)
  $index = Invoke-C11External -Stage 'verify-c11: generated inventory' -FilePath 'git' -ArgumentList @(
    'ls-files', '--stage', '--', 'api', 'gen', 'internal/store'
  ) -WorkingDirectory $RepoRoot -TimeoutSeconds 30 -CaptureOutput -MaximumCaptureBytes 4194304
  return [pscustomobject]@{ Worktree = $array; Index = $index }
}

function Test-C11GeneratedSnapshotEqual {
  param(
    [Parameter(Mandatory)]$Before,
    [Parameter(Mandatory)]$After
  )

  if (-not [string]::Equals([string]$Before.Index, [string]$After.Index, [System.StringComparison]::Ordinal)) {
    return $false
  }
  if ($Before.Worktree.Count -ne $After.Worktree.Count) {
    return $false
  }
  for ($index = 0; $index -lt $Before.Worktree.Count; $index++) {
    if (-not [string]::Equals($Before.Worktree[$index], $After.Worktree[$index], [System.StringComparison]::Ordinal)) {
      return $false
    }
  }
  return $true
}

function New-C11ComposeOverride {
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
    Stop-C11Verification -ExitCode 1 -Stage 'verify-c11: dependency ownership'
  }
  return $path
}

function Assert-C11ComposeOwnership {
  param(
    [Parameter(Mandatory)][string]$Project,
    [Parameter(Mandatory)][string]$Service
  )

  $containerID = Invoke-C11External -Stage 'verify-c11: dependency ownership' -FilePath 'docker' -ArgumentList (
    $script:C11ComposeArguments + @('ps', '-q', $Service)
  ) -TimeoutSeconds 30 -CaptureOutput
  if ($containerID -notmatch '^[0-9a-f]{12,64}$') {
    Stop-C11Verification -ExitCode 1 -Stage 'verify-c11: dependency ownership'
  }
  $owner = Invoke-C11External -Stage 'verify-c11: dependency ownership' -FilePath 'docker' -ArgumentList @(
    'inspect', '--format', '{{ index .Config.Labels `com.docker.compose.project` }}', $containerID
  ) -TimeoutSeconds 30 -CaptureOutput
  if (-not [string]::Equals($owner, $Project, [System.StringComparison]::Ordinal)) {
    Stop-C11Verification -ExitCode 1 -Stage 'verify-c11: dependency ownership'
  }
}

function Get-C11LoopbackPort {
  param([Parameter(Mandatory)][string]$Service, [Parameter(Mandatory)][int]$ContainerPort)

  $endpoint = Invoke-C11External -Stage 'verify-c11: dependency endpoints' -FilePath 'docker' -ArgumentList (
    $script:C11ComposeArguments + @('port', $Service, [string]$ContainerPort)
  ) -TimeoutSeconds 30 -CaptureOutput
  if ($endpoint -notmatch '^127\.0\.0\.1:(?<port>[1-9][0-9]{0,4})$') {
    Stop-C11Verification -ExitCode 1 -Stage 'verify-c11: dependency endpoints'
  }
  $port = [int]$Matches.port
  if ($port -gt 65535) {
    Stop-C11Verification -ExitCode 1 -Stage 'verify-c11: dependency endpoints'
  }
  return $port
}

function Initialize-C11Dependencies {
  param([Parameter(Mandatory)][string]$ComposeOverride)

  $project = 'talenro-c11-verify-' + [System.Guid]::NewGuid().ToString('N').Substring(0, 12)
  if ($project -notmatch '^talenro-c11-verify-[0-9a-f]{12}$') {
    Stop-C11Verification -ExitCode 1 -Stage 'verify-c11: dependency ownership'
  }
  $script:C11ComposeArguments = @(
    'compose', '--project-name', $project,
    '-f', $ComposeOverride
  )
  $script:C11ComposeTouched = $true
  Invoke-C11External -Stage 'verify-c11: dependency startup' -FilePath 'docker' -ArgumentList (
    $script:C11ComposeArguments + @('up', '-d', '--wait', '--wait-timeout', '120', 'postgres', 'redis', 'nats')
  ) -TimeoutSeconds 180

  foreach ($service in @('postgres', 'redis', 'nats')) {
    Assert-C11ComposeOwnership -Project $project -Service $service
  }
  $postgresPort = Get-C11LoopbackPort -Service 'postgres' -ContainerPort 5432
  $redisPort = Get-C11LoopbackPort -Service 'redis' -ContainerPort 6379
  $natsPort = Get-C11LoopbackPort -Service 'nats' -ContainerPort 4222
  return [pscustomobject]@{
    Project = $project
    DatabaseURL = "postgres://talenro:talenro_dev@127.0.0.1:${postgresPort}/talenro?sslmode=disable"
    RedisAddress = "127.0.0.1:${redisPort}"
    NATSURL = "nats://127.0.0.1:${natsPort}"
  }
}

function Close-C11Dependencies {
  if (-not $script:C11ComposeTouched) {
    return
  }
  Invoke-C11External -Stage 'verify-c11: dependency cleanup' -FilePath 'docker' -ArgumentList (
    $script:C11ComposeArguments + @('down', '--remove-orphans', '--timeout', '20')
  ) -TimeoutSeconds 60
  $script:C11ComposeTouched = $false
}

function Get-C11BashExecutable {
  $commands = @(Get-Command -Name 'bash' -CommandType Application -ErrorAction SilentlyContinue)
  if ($commands.Count -gt 0) {
    return $commands[0].Source
  }

  foreach ($candidate in @(
    'C:\Program Files\Git\bin\bash.exe',
    'C:\Program Files\Git\usr\bin\bash.exe'
  )) {
    if ([System.IO.File]::Exists($candidate)) {
      return $candidate
    }
  }
  return $null
}

function Assert-C11OwnedConformanceBinary {
  param([Parameter(Mandatory)][string]$Path)

  try {
    $fullPath = [System.IO.Path]::GetFullPath($Path)
    $ownedDirectory = [System.IO.Path]::GetFullPath($script:C11RunDirectory).TrimEnd('\')
    $expectedPath = [System.IO.Path]::GetFullPath((Join-Path $ownedDirectory 'trust-conformance.exe'))
    $attributes = [System.IO.File]::GetAttributes($fullPath)
    $length = [System.IO.FileInfo]::new($fullPath).Length
  } catch {
    Stop-C11Verification -ExitCode 1 -Stage 'verify-c11: conformance binary ownership'
  }

  if (
    -not [System.IO.Path]::IsPathRooted($Path) -or
    -not [string]::Equals($fullPath, $expectedPath, [System.StringComparison]::OrdinalIgnoreCase) -or
    -not [System.IO.File]::Exists($fullPath) -or
    ($attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0 -or
    $length -le 0
  ) {
    Stop-C11Verification -ExitCode 1 -Stage 'verify-c11: conformance binary ownership'
  }
  return $fullPath
}

function Invoke-C11E2ETests {
  param(
    [Parameter(Mandatory)][string]$RepoRoot,
    [Parameter(Mandatory)][string]$ConformanceBinary
  )

  foreach ($name in @(
    'C11_E2E_COMPOSE_PROJECT',
    'C11_E2E_DATABASE_URL',
    'C11_E2E_REDIS_ADDRESS',
    'C11_E2E_NATS_URL'
  )) {
    if ([string]::IsNullOrEmpty([System.Environment]::GetEnvironmentVariable($name, 'Process'))) {
      Stop-C11Verification -ExitCode 1 -Stage 'verify-c11: e2e environment'
    }
  }

  $previousConformanceBinary = [System.Environment]::GetEnvironmentVariable('C11_CONFORMANCE_BINARY', 'Process')
  $primaryFailure = $null
  $restoreFailed = $false
  try {
    [System.Environment]::SetEnvironmentVariable('C11_CONFORMANCE_BINARY', $ConformanceBinary, 'Process')
    Invoke-C11Stage -Stage 'e2e tests' -FilePath 'go' -WorkingDirectory $RepoRoot -ArgumentList @(
      'test', '-tags=e2e', './internal/e2e', '-count=1', '-timeout', '10m'
    )
  } catch {
    $primaryFailure = $_.Exception
  } finally {
    try {
      [System.Environment]::SetEnvironmentVariable('C11_CONFORMANCE_BINARY', $previousConformanceBinary, 'Process')
    } catch {
      $restoreFailed = $true
    }
  }
  if ($null -ne $primaryFailure) {
    throw $primaryFailure
  }
  if ($restoreFailed) {
    Stop-C11Verification -ExitCode 1 -Stage 'verify-c11: e2e environment cleanup'
  }
}

function Invoke-C11Main {
  $repoRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..')).Path
  $powerShell = (Get-Process -Id $PID).Path
  $migrationDirectory = Join-Path $repoRoot 'db\migrations'
  $primaryFailure = $null
  $cleanupFailure = $null
  $composeOverride = $null
  $locationPushed = $false

  try {
    $script:C11RunDirectory = New-C11RunDirectory
    $composeOverride = New-C11ComposeOverride -Directory $script:C11RunDirectory
    Clear-C11AmbientEnvironment
    Push-Location -LiteralPath $repoRoot
    $locationPushed = $true

    try {
      Invoke-C11Stage -Stage 'check tools' -FilePath $powerShell -ArgumentList @(
        '-NoProfile', '-ExecutionPolicy', 'Bypass', '-File', (Join-Path $PSScriptRoot 'check-tools.ps1')
      ) -TimeoutSeconds 900

      $generatedBefore = Get-C11GeneratedSnapshot -RepoRoot $repoRoot
      Invoke-C11Stage -Stage 'generate' -FilePath $powerShell -ArgumentList @(
        '-NoProfile', '-ExecutionPolicy', 'Bypass', '-File', (Join-Path $PSScriptRoot 'generate.ps1')
      ) -TimeoutSeconds 600
      $generatedAfter = Get-C11GeneratedSnapshot -RepoRoot $repoRoot
      Write-Output 'verify-c11: generated diff'
      if (-not (Test-C11GeneratedSnapshotEqual -Before $generatedBefore -After $generatedAfter)) {
        Stop-C11Verification -ExitCode 1 -Stage 'verify-c11: generated diff'
      }

      Invoke-C11Stage -Stage 'unit tests' -FilePath 'go' -ArgumentList @('test', './...', '-count=1', '-timeout', '10m')
      $fuzzTargets = @(
        [pscustomobject]@{ Package = './internal/strictjson'; Target = 'FuzzDecode' }
        [pscustomobject]@{ Package = './internal/trust'; Target = 'FuzzEnvelopeSeal' }
        [pscustomobject]@{ Package = './internal/trustclient'; Target = 'FuzzVerifyEnvelope' }
      )
      foreach ($fuzz in $fuzzTargets) {
        Invoke-C11Stage -Stage "fuzz $($fuzz.Target)" -FilePath 'go' -ArgumentList @(
          'test', '-run', '^$', '-fuzz', "^$($fuzz.Target)$", '-fuzztime=10s', '-timeout', '30s', $fuzz.Package
        ) -TimeoutSeconds 120
      }

      Invoke-C11Stage -Stage 'race tests' -FilePath 'go' -ArgumentList @('test', '-race', './...', '-count=1', '-timeout', '10m')
      Invoke-C11Stage -Stage 'go vet' -FilePath 'go' -ArgumentList @('vet', './...')
      Invoke-C11Stage -Stage 'golangci-lint' -FilePath 'go' -ArgumentList @('tool', 'golangci-lint', 'run', './...')

      $runtime = Initialize-C11Dependencies -ComposeOverride $composeOverride
      [System.Environment]::SetEnvironmentVariable('TALENRO_DATABASE_URL', $runtime.DatabaseURL, 'Process')
      [System.Environment]::SetEnvironmentVariable('TALENRO_REDIS_ADDRESS', $runtime.RedisAddress, 'Process')
      [System.Environment]::SetEnvironmentVariable('TALENRO_NATS_URL', $runtime.NATSURL, 'Process')
      [System.Environment]::SetEnvironmentVariable('C11_E2E_COMPOSE_PROJECT', $runtime.Project, 'Process')
      [System.Environment]::SetEnvironmentVariable('C11_E2E_DATABASE_URL', $runtime.DatabaseURL, 'Process')
      [System.Environment]::SetEnvironmentVariable('C11_E2E_REDIS_ADDRESS', $runtime.RedisAddress, 'Process')
      [System.Environment]::SetEnvironmentVariable('C11_E2E_NATS_URL', $runtime.NATSURL, 'Process')

      Invoke-C11Stage -Stage 'migrations up' -FilePath 'go' -ArgumentList @(
        'tool', 'goose', '-dir', $migrationDirectory, 'postgres', $runtime.DatabaseURL, 'up'
      ) -TimeoutSeconds 120
      Invoke-C11Stage -Stage 'migrations down-to 1' -FilePath 'go' -ArgumentList @(
        'tool', 'goose', '-dir', $migrationDirectory, 'postgres', $runtime.DatabaseURL, 'down-to', '1'
      ) -TimeoutSeconds 120
      Invoke-C11Stage -Stage 'migrations restore' -FilePath 'go' -ArgumentList @(
        'tool', 'goose', '-dir', $migrationDirectory, 'postgres', $runtime.DatabaseURL, 'up'
      ) -TimeoutSeconds 120
      Invoke-C11Stage -Stage 'integration tests' -FilePath 'go' -ArgumentList @(
        'test', '-tags=integration', './...', '-count=1', '-timeout', '10m'
      )

      $conformanceBinary = Join-Path $script:C11RunDirectory 'trust-conformance.exe'
      Invoke-C11Stage -Stage 'conformance build' -FilePath 'go' -WorkingDirectory $repoRoot -ArgumentList @(
        'build', '-o', $conformanceBinary, './cmd/trust-conformance'
      )
      $conformanceBinary = Assert-C11OwnedConformanceBinary -Path $conformanceBinary

      Invoke-C11Stage -Stage 'ephemeral compose config' -FilePath 'docker' -ArgumentList (
        $script:C11ComposeArguments + @('config', '--quiet')
      ) -TimeoutSeconds 60
      Invoke-C11Stage -Stage 'checked-in compose config' -FilePath 'docker' -ArgumentList @(
        'compose', '-f', 'deploy/dev/compose.yaml', 'config', '--quiet'
      ) -TimeoutSeconds 60
      Invoke-C11E2ETests -RepoRoot $repoRoot -ConformanceBinary $conformanceBinary

      Close-C11Dependencies
      Clear-C11AmbientEnvironment

      Invoke-C11Stage -Stage 'PowerShell smoke' -FilePath $powerShell -ArgumentList @(
        '-NoProfile', '-ExecutionPolicy', 'Bypass', '-File', (Join-Path $PSScriptRoot 'smoke.ps1')
      ) -TimeoutSeconds 4200
      $bashExecutable = Get-C11BashExecutable
      if ($null -ne $bashExecutable) {
        Invoke-C11Stage -Stage 'Bash smoke' -FilePath $bashExecutable -ArgumentList @(
          (Join-Path $PSScriptRoot 'smoke.sh')
        ) -TimeoutSeconds 4200
      }
    } catch {
      $primaryFailure = $_.Exception
    }
  } finally {
    try {
      Close-C11Dependencies
    } catch {
      $cleanupFailure = $_.Exception
    }
    if ($null -ne $script:C11RunDirectory -and [System.IO.Directory]::Exists($script:C11RunDirectory)) {
      try {
        Remove-C11RunDirectory -Directory $script:C11RunDirectory
      } catch {
        if ($null -eq $cleanupFailure) {
          $cleanupFailure = $_.Exception
        }
      }
    }
    if ($locationPushed) {
      try {
        Pop-Location
      } catch {
        if ($null -eq $cleanupFailure) {
          $cleanupFailure = $_.Exception
        }
      }
    }
  }

  if ($null -ne $primaryFailure) {
    throw $primaryFailure
  }
  if ($null -ne $cleanupFailure) {
    throw $cleanupFailure
  }
}

try {
  Invoke-C11Main
} catch {
  $exitCode = 1
  $stage = 'verify-c11: internal'
  if ($_.Exception.Data.Contains('ExitCode')) {
    $candidateExitCode = $_.Exception.Data['ExitCode']
    if ($null -ne $candidateExitCode) {
      try {
        $convertedExitCode = [int]$candidateExitCode
        if ($convertedExitCode -ne 0) {
          $exitCode = $convertedExitCode
        }
      } catch {
        $exitCode = 1
      }
    }
  }
  if ($_.Exception.Data.Contains('Stage')) {
    $stage = [string]$_.Exception.Data['Stage']
  }
  [Console]::Error.WriteLine("${stage} failed with exit code ${exitCode}.")
  exit $exitCode
}
