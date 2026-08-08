$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
Add-Type -AssemblyName System.Net.Http

function Stop-Smoke {
  param(
    [Parameter(Mandatory)]
    [int]$ExitCode,

    [Parameter(Mandatory)]
    [string]$Stage
  )

  $failure = [System.Exception]::new('sanitized smoke failure')
  $failure.Data['ExitCode'] = $ExitCode
  $failure.Data['Stage'] = $Stage
  throw $failure
}

function Get-SmokeFailure {
  param([Parameter(Mandatory)]$ErrorRecord)

  $result = @{ ExitCode = 1; Stage = 'smoke: internal' }
  if ($ErrorRecord.Exception.Data.Contains('ExitCode')) {
    $result.ExitCode = [int]$ErrorRecord.Exception.Data['ExitCode']
  }
  if ($ErrorRecord.Exception.Data.Contains('Stage')) {
    $result.Stage = [string]$ErrorRecord.Exception.Data['Stage']
  }
  return $result
}

function Invoke-QuietExternal {
  param(
    [Parameter(Mandatory)]
    [string]$Stage,

    [Parameter(Mandatory)]
    [string]$FilePath,

    [Parameter(Mandatory)]
    [string[]]$ArgumentList
  )

  $commands = @(Get-Command -Name $FilePath -CommandType Application, ExternalScript -ErrorAction SilentlyContinue)
  if ($commands.Count -eq 0) {
    Stop-Smoke -ExitCode 127 -Stage $Stage
  }

  $previousErrorAction = $ErrorActionPreference
  $ErrorActionPreference = 'Continue'
  try {
    $global:LASTEXITCODE = 0
    $null = @(& $commands[0].Source @ArgumentList 2>&1 | ForEach-Object { $_.ToString() })
    $exitCode = $LASTEXITCODE
  } finally {
    $ErrorActionPreference = $previousErrorAction
  }

  if ($exitCode -ne 0) {
    Stop-Smoke -ExitCode $exitCode -Stage $Stage
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
    'TALENRO_ALLOW_PUBLIC_HTTP'
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

  foreach ($name in $requiredNames) {
    [System.Environment]::SetEnvironmentVariable($name, $null, 'Process')
    [System.Environment]::SetEnvironmentVariable($name, [string]$parsed[$name], 'Process')
  }

  return $parsed
}

function Start-ControlAPI {
  param([Parameter(Mandatory)][string]$WorkingDirectory)

  $commands = @(Get-Command -Name 'go' -CommandType Application -ErrorAction SilentlyContinue)
  if ($commands.Count -eq 0) {
    Stop-Smoke -ExitCode 127 -Stage 'smoke: control API start'
  }

  $argumentList = @('run', './cmd/control-api')
  foreach ($argument in $argumentList) {
    if ($argument -notmatch '^[A-Za-z0-9./-]+$') {
      Stop-Smoke -ExitCode 2 -Stage 'smoke: control API arguments'
    }
  }

  $startInfo = [System.Diagnostics.ProcessStartInfo]::new()
  $startInfo.FileName = $commands[0].Source
  $startInfo.Arguments = $argumentList -join ' '
  $startInfo.WorkingDirectory = $WorkingDirectory
  $startInfo.UseShellExecute = $false
  $startInfo.CreateNoWindow = $true
  $startInfo.RedirectStandardOutput = $true
  $startInfo.RedirectStandardError = $true

  $process = [System.Diagnostics.Process]::new()
  $process.StartInfo = $startInfo
  try {
    if (-not $process.Start()) {
      $process.Dispose()
      Stop-Smoke -ExitCode 127 -Stage 'smoke: control API start'
    }
    $stdoutDrain = $process.StandardOutput.ReadToEndAsync()
    $stderrDrain = $process.StandardError.ReadToEndAsync()
    $process | Add-Member -NotePropertyName SmokeStdoutDrain -NotePropertyValue $stdoutDrain
    $process | Add-Member -NotePropertyName SmokeStderrDrain -NotePropertyValue $stderrDrain
    return $process
  } catch {
    $process.Dispose()
    Stop-Smoke -ExitCode 127 -Stage 'smoke: control API start'
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
    [Parameter(Mandatory)][string]$Stage
  )

  $timer = [System.Diagnostics.Stopwatch]::StartNew()
  $limitMilliseconds = $Seconds * 1000
  while ($timer.ElapsedMilliseconds -lt $limitMilliseconds) {
    if ($Process.HasExited) {
      $processExitCode = $Process.ExitCode
      if ($processExitCode -eq 0) {
        $processExitCode = 1
      }
      Stop-Smoke -ExitCode $processExitCode -Stage 'smoke: control API'
    }

    $remaining = $limitMilliseconds - [int]$timer.ElapsedMilliseconds
    $requestTimeout = [Math]::Max(1, [Math]::Min(1000, $remaining))
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

$repoRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..')).Path
$composeArguments = @('compose', '-f', (Join-Path $repoRoot 'deploy\dev\compose.yaml'))
$composeTouched = $false
$controlAPI = $null
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
    [string]$localEnvironment['TALENRO_ALLOW_PUBLIC_HTTP'] -ne 'false'
  ) {
    Stop-Smoke -ExitCode 2 -Stage 'smoke: local environment'
  }
  $httpBase = 'http://' + [string]$localEnvironment['TALENRO_HTTP_ADDRESS']

  $handler = [System.Net.Http.HttpClientHandler]::new()
  $handler.UseProxy = $false
  $client = [System.Net.Http.HttpClient]::new($handler, $true)
  $client.Timeout = [System.Threading.Timeout]::InfiniteTimeSpan

  $composeTouched = $true
  Invoke-QuietExternal -Stage 'smoke: compose up' -FilePath 'docker' -ArgumentList (
    $composeArguments + @('up', '-d', '--wait')
  )
  Invoke-QuietExternal -Stage 'smoke: migrations' -FilePath 'go' -ArgumentList @(
    'tool',
    'goose',
    '-dir',
    (Join-Path $repoRoot 'db\migrations'),
    'postgres',
    [string]$localEnvironment['TALENRO_DATABASE_URL'],
    'up'
  )

  $controlAPI = Start-ControlAPI -WorkingDirectory $repoRoot
  Wait-HTTPStatus -Client $client -Process $controlAPI -URL ($httpBase + '/livez') -Expected 200 -Seconds 10 -Stage 'smoke: initial liveness'
  Wait-HTTPStatus -Client $client -Process $controlAPI -URL ($httpBase + '/readyz') -Expected 200 -Seconds 10 -Stage 'smoke: initial readiness'

  Invoke-QuietExternal -Stage 'smoke: postgres stop' -FilePath 'docker' -ArgumentList (
    $composeArguments + @('stop', 'postgres')
  )
  Wait-HTTPStatus -Client $client -Process $controlAPI -URL ($httpBase + '/readyz') -Expected 503 -Seconds 5 -Stage 'smoke: readiness failure'
  Wait-HTTPStatus -Client $client -Process $controlAPI -URL ($httpBase + '/livez') -Expected 200 -Seconds 2 -Stage 'smoke: failure liveness'

  Invoke-QuietExternal -Stage 'smoke: postgres start' -FilePath 'docker' -ArgumentList (
    $composeArguments + @('start', 'postgres')
  )
  Wait-HTTPStatus -Client $client -Process $controlAPI -URL ($httpBase + '/readyz') -Expected 200 -Seconds 10 -Stage 'smoke: readiness recovery'
} catch {
  $primaryFailure = Get-SmokeFailure -ErrorRecord $_
} finally {
  if ($null -ne $controlAPI) {
    try {
      if (-not $controlAPI.HasExited) {
        Stop-Process -Id $controlAPI.Id -Force -ErrorAction Stop
      }
      $controlAPI.WaitForExit()
      $null = $controlAPI.SmokeStdoutDrain.GetAwaiter().GetResult()
      $null = $controlAPI.SmokeStderrDrain.GetAwaiter().GetResult()
    } catch {
      if ($null -eq $cleanupFailure) {
        $cleanupFailure = @{ ExitCode = 1; Stage = 'smoke: control API cleanup' }
      }
    } finally {
      $controlAPI.Dispose()
    }
  }

  if ($null -ne $client) {
    $client.Dispose()
  }

  if ($composeTouched) {
    try {
      Invoke-QuietExternal -Stage 'smoke: compose down' -FilePath 'docker' -ArgumentList (
        $composeArguments + @('down')
      )
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
  [Console]::Error.WriteLine("$($failure.Stage) failed with exit code $($failure.ExitCode).")
  exit $failure.ExitCode
}

Write-Output 'smoke: foundation acceptance passed.'
