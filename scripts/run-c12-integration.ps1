[CmdletBinding(DefaultParameterSetName = 'Focused')]
param(
  [Parameter(Mandatory = $true, ParameterSetName = 'Focused')]
  [string]$Profile,

  [Parameter(Mandatory = $true, ParameterSetName = 'Focused')]
  [string]$Packages,

  [Parameter(ParameterSetName = 'Focused')]
  [string]$Run = '',

  [Parameter(ParameterSetName = 'Focused')]
  [switch]$Race,

  [Parameter(Mandatory = $true, ParameterSetName = 'Suite')]
  [string]$Suite,

  [Parameter(Mandatory = $true, ParameterSetName = 'Focused')]
  [Parameter(Mandatory = $true, ParameterSetName = 'Suite')]
  [string]$Timeout
)

Set-StrictMode -Version 2.0
$ErrorActionPreference = 'Stop'

$script:c12AllowedPackages = @{
  './internal/testinfra' = $true
  './internal/store' = $true
  './internal/nodecontrol/contracts' = $true
  './internal/nodecontrol/authority' = $true
  './internal/nodecontrol/serving' = $true
  './internal/readiness' = $true
}

function ConvertFrom-C12Duration {
  param(
    [Parameter(Mandatory = $true)]
    [string]$Value
  )

  if ($Value -notmatch '^([1-9][0-9]*)(s|m)$') {
    throw "timeout '$Value' is not a finite positive seconds/minutes duration"
  }
  [long]$quantity = 0
  if (-not [long]::TryParse($Matches[1], [ref]$quantity)) {
    throw "timeout '$Value' is outside the supported range"
  }
  if ($Matches[2] -eq 'm') {
    if ($quantity -gt 1440) {
      throw "timeout '$Value' is outside the supported range"
    }
    return [TimeSpan]::FromMinutes($quantity)
  }
  if ($quantity -gt 86400) {
    throw "timeout '$Value' is outside the supported range"
  }
  return [TimeSpan]::FromSeconds($quantity)
}

function New-C12RandomSuffix {
  $bytes = New-Object byte[] 16
  $random = [System.Security.Cryptography.RandomNumberGenerator]::Create()
  try {
    $random.GetBytes($bytes)
  }
  finally {
    $random.Dispose()
  }
  return (($bytes | ForEach-Object { $_.ToString('x2') }) -join '')
}

function New-C12DatabasePassword {
  param(
    [Parameter(Mandatory = $true)]
    [string]$RunSuffix
  )

  $hash = [System.Security.Cryptography.SHA256]::Create()
  try {
    $bytes = [System.Text.Encoding]::UTF8.GetBytes("TALENRO-C12-POSTGRES-PASSWORD-V1`0$RunSuffix")
    return (($hash.ComputeHash($bytes) | ForEach-Object { $_.ToString('x2') }) -join '')
  }
  finally {
    $hash.Dispose()
  }
}

function Invoke-C12Docker {
  param(
    [Parameter(Mandatory = $true)]
    [string[]]$Arguments,

    [Parameter(Mandatory = $true)]
    [string]$Stage,

    [switch]$AllowFailure
  )

  $dockerArgs = $Arguments
  $priorNativeErrorActionPreference = $ErrorActionPreference
  try {
    $ErrorActionPreference = 'Continue'
    $output = @(& docker @dockerArgs 2>&1)
    $dockerExitCode = $LASTEXITCODE
  }
  finally {
    $ErrorActionPreference = $priorNativeErrorActionPreference
  }
  if (-not $AllowFailure -and $dockerExitCode -ne 0) {
    throw "$Stage failed with exit code $dockerExitCode"
  }
  return [pscustomobject]@{
    ExitCode = $dockerExitCode
    Output = @($output | ForEach-Object { $_.ToString() })
  }
}

function Invoke-C12Go {
  param(
    [Parameter(Mandatory = $true)]
    [string[]]$Arguments,

    [Parameter(Mandatory = $true)]
    [string]$Stage,

    [switch]$AllowFailure
  )

  $goArgs = $Arguments
  $priorNativeErrorActionPreference = $ErrorActionPreference
  try {
    $ErrorActionPreference = 'Continue'
    $output = @(& go @goArgs 2>&1)
    $goExitCode = $LASTEXITCODE
  }
  finally {
    $ErrorActionPreference = $priorNativeErrorActionPreference
  }
  if (-not $AllowFailure -and $goExitCode -ne 0) {
    throw "$Stage failed with exit code $goExitCode"
  }
  return [pscustomobject]@{
    ExitCode = $goExitCode
    Output = @($output | ForEach-Object { $_.ToString() })
  }
}

function Get-C12SingleOutputLine {
  param(
    [Parameter(Mandatory = $true)]
    [object]$Result,

    [Parameter(Mandatory = $true)]
    [string]$Stage
  )

  $lines = @($Result.Output | Where-Object { -not [string]::IsNullOrWhiteSpace($_) })
  if ($Result.ExitCode -ne 0 -or $lines.Count -ne 1) {
    throw "$Stage did not return one bounded line"
  }
  return [string]$lines[0]
}

function Assert-C12ContainerIdentity {
  param(
    [Parameter(Mandatory = $true)]
    [object]$Resource,

    [Parameter(Mandatory = $true)]
    [string]$RunSuffix
  )

  if ([string]::IsNullOrEmpty([string]$Resource.ID)) {
    throw "container $($Resource.Kind) has no captured ID"
  }
  $dockerArgs = @(
    'container', 'inspect', '--format',
    '{{.Id}}|{{.Name}}|{{ index .Config.Labels `talenro.c12.run` }}',
    [string]$Resource.Name
  )
  $result = Invoke-C12Docker -Arguments $dockerArgs -Stage "inspect $($Resource.Kind) identity"
  $identity = Get-C12SingleOutputLine -Result $result -Stage "inspect $($Resource.Kind) identity"
  $parts = $identity -split '\|', 3
  if ($parts.Count -ne 3 -or
      $parts[0] -cne [string]$Resource.ID -or
      $parts[1] -cne "/$($Resource.Name)" -or
      $parts[2] -cne $RunSuffix) {
    throw "container $($Resource.Kind) identity mismatch"
  }
}

function Set-C12MappedPort {
  param(
    [Parameter(Mandatory = $true)]
    [object]$Resource
  )

  $dockerArgs = @('container', 'port', [string]$Resource.ID, "$($Resource.ContainerPort)/tcp")
  $result = Invoke-C12Docker -Arguments $dockerArgs -Stage "inspect $($Resource.Kind) mapped port"
  $mapping = Get-C12SingleOutputLine -Result $result -Stage "inspect $($Resource.Kind) mapped port"
  if ($mapping -notmatch '^127\.0\.0\.1:([1-9][0-9]{0,4})$') {
    throw "container $($Resource.Kind) has a non-loopback or malformed mapped port"
  }
  $port = [int]$Matches[1]
  if ($port -gt 65535) {
    throw "container $($Resource.Kind) has an invalid mapped port"
  }
  $Resource.Port = $port
}

function Start-C12Container {
  param(
    [Parameter(Mandatory = $true)]
    [object]$Resource,

    [Parameter(Mandatory = $true)]
    [string]$RunSuffix,

    [Parameter(Mandatory = $true)]
    [string]$DatabaseName,

    [Parameter(Mandatory = $true)]
    [string]$DatabasePassword
  )

  switch ($Resource.Kind) {
    'postgres' {
      $dockerArgs = @(
        'run', '--detach', '--name', [string]$Resource.Name,
        '--label', "talenro.c12.run=$RunSuffix",
        '--publish', '127.0.0.1::5432',
        '--env', 'POSTGRES_USER=talenro',
        '--env', "POSTGRES_PASSWORD=$DatabasePassword",
        '--env', "POSTGRES_DB=$DatabaseName",
        '--health-cmd', "pg_isready -U talenro -d $DatabaseName",
        '--health-interval', '1s', '--health-timeout', '2s', '--health-retries', '60',
        'postgres:18.4-alpine3.23'
      )
    }
    'redis' {
      $dockerArgs = @(
        'run', '--detach', '--name', [string]$Resource.Name,
        '--label', "talenro.c12.run=$RunSuffix",
        '--publish', '127.0.0.1::6379',
        'redis:8.8.1-alpine3.23'
      )
    }
    'nats' {
      $dockerArgs = @(
        'run', '--detach', '--name', [string]$Resource.Name,
        '--label', "talenro.c12.run=$RunSuffix",
        '--publish', '127.0.0.1::4222',
        'nats:2.14.3-alpine3.22', '-js', '--name', [string]$Resource.Name
      )
    }
    default {
      throw "unknown C12 dependency kind $($Resource.Kind)"
    }
  }

  $result = Invoke-C12Docker -Arguments $dockerArgs -Stage "start $($Resource.Kind) container"
  $containerID = Get-C12SingleOutputLine -Result $result -Stage "start $($Resource.Kind) container"
  if ($containerID -notmatch '^[0-9a-f]{64}$') {
    throw "start $($Resource.Kind) did not return an exact container ID"
  }
  $Resource.ID = $containerID
  Assert-C12ContainerIdentity -Resource $Resource -RunSuffix $RunSuffix
  Set-C12MappedPort -Resource $Resource
}

function Test-C12TCPProtocol {
  param(
    [Parameter(Mandatory = $true)]
    [int]$Port,

    [Parameter(Mandatory = $true)]
    [string]$Request,

    [Parameter(Mandatory = $true)]
    [string]$Expected
  )

  $client = New-Object System.Net.Sockets.TcpClient
  try {
    $connection = $client.BeginConnect('127.0.0.1', $Port, $null, $null)
    if (-not $connection.AsyncWaitHandle.WaitOne(1000)) {
      return $false
    }
    $client.EndConnect($connection)
    $stream = $client.GetStream()
    $stream.ReadTimeout = 1000
    $stream.WriteTimeout = 1000
    $requestBytes = [System.Text.Encoding]::ASCII.GetBytes($Request)
    $stream.Write($requestBytes, 0, $requestBytes.Length)
    $stream.Flush()
    $buffer = New-Object byte[] 4096
    $response = New-Object System.Text.StringBuilder
    for ($attempt = 0; $attempt -lt 3; $attempt++) {
      $read = $stream.Read($buffer, 0, $buffer.Length)
      if ($read -le 0) {
        break
      }
      [void]$response.Append([System.Text.Encoding]::ASCII.GetString($buffer, 0, $read))
      if ($response.ToString().Contains($Expected)) {
        return $true
      }
    }
    return $false
  }
  catch {
    return $false
  }
  finally {
    $client.Close()
  }
}

function Wait-C12Dependencies {
  param(
    [Parameter(Mandatory = $true)]
    [object[]]$Resources
  )

  $postgres = @($Resources | Where-Object { $_.Kind -eq 'postgres' })[0]
  $redis = @($Resources | Where-Object { $_.Kind -eq 'redis' })[0]
  $nats = @($Resources | Where-Object { $_.Kind -eq 'nats' })[0]
  $deadline = [DateTime]::UtcNow.AddSeconds(60)
  while ([DateTime]::UtcNow -lt $deadline) {
    $dockerArgs = @('container', 'inspect', '--format', '{{.State.Health.Status}}', [string]$postgres.ID)
    $healthResult = Invoke-C12Docker -Arguments $dockerArgs -Stage 'inspect PostgreSQL health' -AllowFailure
    $postgresReady = $false
    if ($healthResult.ExitCode -eq 0) {
      $healthLines = @($healthResult.Output | Where-Object { -not [string]::IsNullOrWhiteSpace($_) })
      $postgresReady = $healthLines.Count -eq 1 -and [string]$healthLines[0] -ceq 'healthy'
    }
    $redisReady = Test-C12TCPProtocol -Port $redis.Port -Request "*1`r`n`$4`r`nPING`r`n" -Expected "+PONG`r`n"
    $natsReady = Test-C12TCPProtocol -Port $nats.Port -Request "CONNECT {`"verbose`":false,`"pedantic`":false}`r`nPING`r`n" -Expected "PONG`r`n"
    if ($postgresReady -and $redisReady -and $natsReady) {
      return
    }
    Start-Sleep -Milliseconds 250
  }
  throw 'C12 dependencies did not pass health and protocol probes within 60 seconds'
}

function Resolve-C12CleanupIdentity {
  param(
    [Parameter(Mandatory = $true)]
    [object]$Resource,

    [Parameter(Mandatory = $true)]
    [string]$RunSuffix
  )

  $dockerArgs = @(
    'container', 'inspect', '--format',
    '{{.Id}}|{{.Name}}|{{ index .Config.Labels `talenro.c12.run` }}',
    [string]$Resource.Name
  )
  $result = Invoke-C12Docker -Arguments $dockerArgs -Stage "re-inspect $($Resource.Kind) before cleanup" -AllowFailure
  if ($result.ExitCode -ne 0) {
    if (-not [string]::IsNullOrEmpty([string]$Resource.ID)) {
      throw "captured $($Resource.Kind) container is absent before cleanup"
    }
    return $null
  }
  $identity = Get-C12SingleOutputLine -Result $result -Stage "re-inspect $($Resource.Kind) before cleanup"
  $parts = $identity -split '\|', 3
  if ($parts.Count -ne 3 -or
      $parts[0] -notmatch '^[0-9a-f]{64}$' -or
      $parts[1] -cne "/$($Resource.Name)" -or
      $parts[2] -cne $RunSuffix) {
    throw "refusing cleanup for $($Resource.Kind): exact name/ID/run label mismatch"
  }
  if (-not [string]::IsNullOrEmpty([string]$Resource.ID) -and $parts[0] -cne [string]$Resource.ID) {
    throw "refusing cleanup for $($Resource.Kind): captured ID mismatch"
  }
  $Resource.ID = $parts[0]
  return $parts[0]
}

function Remove-C12Container {
  param(
    [Parameter(Mandatory = $true)]
    [object]$Resource,

    [Parameter(Mandatory = $true)]
    [string]$RunSuffix
  )

  $containerID = Resolve-C12CleanupIdentity -Resource $Resource -RunSuffix $RunSuffix
  if ([string]::IsNullOrEmpty([string]$containerID)) {
    return
  }
  $dockerArgs = @('container', 'stop', '--time', '10', [string]$containerID)
  $null = Invoke-C12Docker -Arguments $dockerArgs -Stage "stop exact $($Resource.Kind) container"
  $dockerArgs = @('container', 'rm', [string]$containerID)
  $null = Invoke-C12Docker -Arguments $dockerArgs -Stage "remove exact $($Resource.Kind) container"
  $dockerArgs = @('container', 'inspect', '--format', '{{.Id}}', [string]$Resource.Name)
  $absence = Invoke-C12Docker -Arguments $dockerArgs -Stage "verify $($Resource.Kind) cleanup" -AllowFailure
  if ($absence.ExitCode -eq 0) {
    throw "$($Resource.Kind) container remains after exact cleanup"
  }
}

function Convert-C12RunToTests {
  param(
    [Parameter(Mandatory = $true)]
    [string]$Pattern
  )

  if ($Pattern -match '^\^Test[A-Za-z0-9_]+\$$') {
    return @($Pattern.Substring(1, $Pattern.Length - 2))
  }
  if ($Pattern -match '^\^\(Test[A-Za-z0-9_]+(?:\|Test[A-Za-z0-9_]+)+\)\$$') {
    return @(($Pattern.Substring(2, $Pattern.Length - 4)) -split '\|')
  }
  throw '-Run must be an anchored literal Test name or anchored alternation of literal Test names'
}

function Assert-C12GoJSONResult {
  param(
    [Parameter(Mandatory = $true)]
    [object]$Result,

    [Parameter(Mandatory = $true)]
    [string]$Package,

    [string[]]$ExpectedTests = @()
  )

  $expected = @{}
  foreach ($testName in $ExpectedTests) {
    $expected[$testName] = 0
  }
  $topLevelPasses = 0
  $packagePassed = $false
  foreach ($line in $Result.Output) {
    if ([string]::IsNullOrWhiteSpace([string]$line)) {
      continue
    }
    try {
      $event = [string]$line | ConvertFrom-Json
    }
    catch {
      throw "go test for $Package emitted a non-JSON line"
    }
    $hasTest = $event.PSObject.Properties.Name -contains 'Test'
    $testName = if ($hasTest) { [string]$event.Test } else { '' }
    if ($event.Action -eq 'skip') {
      throw "go test for $Package skipped $testName"
    }
    if ($hasTest -and $testName -notmatch '/' -and $event.Action -in @('pass', 'fail')) {
      if ($event.Action -ne 'pass') {
        throw "go test for $Package failed $testName"
      }
      if ($expected.Count -gt 0 -and -not $expected.ContainsKey($testName)) {
        throw "go test for $Package emitted extra top-level test $testName"
      }
      if ($expected.ContainsKey($testName)) {
        $expected[$testName] = [int]$expected[$testName] + 1
      }
      $topLevelPasses++
    }
    if (-not $hasTest -and $event.Action -eq 'pass') {
      $packagePassed = $true
    }
  }
  foreach ($testName in $expected.Keys) {
    if ([int]$expected[$testName] -ne 1) {
      throw "go test for $Package has missing or duplicate terminal pass for $testName"
    }
  }
  if ($topLevelPasses -eq 0 -or -not $packagePassed -or $Result.ExitCode -ne 0) {
    throw "go test for $Package did not produce an exact passing JSON result"
  }
}

function Invoke-C12Group {
  param(
    [Parameter(Mandatory = $true)]
    [string]$GroupID,

    [Parameter(Mandatory = $true)]
    [string]$GroupProfile,

    [Parameter(Mandatory = $true)]
    [string]$Package,

    [string]$RunPattern = '',

    [Parameter(Mandatory = $true)]
    [string]$GroupTimeout,

    [string[]]$ExpectedTests = @()
  )

  if ($GroupProfile -cne 'base') {
    throw 'Task 4 runner accepts only the base profile'
  }
  $runSuffix = New-C12RandomSuffix
  if ($runSuffix -notmatch '^[0-9a-f]{32}$') {
    throw 'cryptographic C12 run suffix is malformed'
  }
  $databaseName = "talenro_c12_$runSuffix"
  $databasePassword = New-C12DatabasePassword -RunSuffix $runSuffix
  $resources = @(
    [pscustomobject]@{ Kind = 'postgres'; Name = "talenro-c12-$runSuffix-postgres"; ID = ''; Port = 0; ContainerPort = 5432 },
    [pscustomobject]@{ Kind = 'redis'; Name = "talenro-c12-$runSuffix-redis"; ID = ''; Port = 0; ContainerPort = 6379 },
    [pscustomobject]@{ Kind = 'nats'; Name = "talenro-c12-$runSuffix-nats"; ID = ''; Port = 0; ContainerPort = 4222 }
  )
  $primaryFailure = $null
  $cleanupFailures = @()
  try {
    foreach ($resource in $resources) {
      Start-C12Container -Resource $resource -RunSuffix $runSuffix -DatabaseName $databaseName -DatabasePassword $databasePassword
    }
    Wait-C12Dependencies -Resources $resources
    $postgres = @($resources | Where-Object { $_.Kind -eq 'postgres' })[0]
    $redis = @($resources | Where-Object { $_.Kind -eq 'redis' })[0]
    $nats = @($resources | Where-Object { $_.Kind -eq 'nats' })[0]
    $databaseURL = "postgres://talenro:$databasePassword@127.0.0.1:$($postgres.Port)/$databaseName`?sslmode=disable&application_name=talenro-c12-$runSuffix"
    [System.Environment]::SetEnvironmentVariable('TALENRO_DATABASE_URL', $databaseURL, 'Process')
    [System.Environment]::SetEnvironmentVariable('TALENRO_REDIS_ADDRESS', "127.0.0.1:$($redis.Port)", 'Process')
    [System.Environment]::SetEnvironmentVariable('TALENRO_NATS_URL', "nats://127.0.0.1:$($nats.Port)", 'Process')

    $migrationDirectory = Join-Path $script:c12RepositoryRoot 'db\migrations'
    $goArgs = @('tool', 'goose', '-dir', $migrationDirectory, 'postgres', $databaseURL, 'up-to', '6')
    $null = Invoke-C12Go -Arguments $goArgs -Stage "ordinary Goose base migration for $GroupID"

    $goArgs = @('test', '-json', '-tags=integration', '-p=1', '-count=1', '-timeout', $GroupTimeout)
    if (-not [string]::IsNullOrEmpty($RunPattern)) {
      $goArgs += @('-run', $RunPattern)
    }
    $goArgs += $Package
    $testResult = Invoke-C12Go -Arguments $goArgs -Stage "tagged Go test group $GroupID" -AllowFailure
    Assert-C12GoJSONResult -Result $testResult -Package $Package -ExpectedTests $ExpectedTests
  }
  catch {
    $primaryFailure = $_.Exception.Message
  }
  finally {
    [System.Environment]::SetEnvironmentVariable('TALENRO_DATABASE_URL', $null, 'Process')
    [System.Environment]::SetEnvironmentVariable('TALENRO_REDIS_ADDRESS', $null, 'Process')
    [System.Environment]::SetEnvironmentVariable('TALENRO_NATS_URL', $null, 'Process')
    for ($index = $resources.Count - 1; $index -ge 0; $index--) {
      try {
        Remove-C12Container -Resource $resources[$index] -RunSuffix $runSuffix
      }
      catch {
        $cleanupFailures += "$($resources[$index].Kind): $($_.Exception.Message)"
      }
    }
  }
  if ($null -ne $primaryFailure) {
    [Console]::Error.WriteLine("C12 test failure [$GroupID]: $primaryFailure")
  }
  foreach ($cleanupFailure in $cleanupFailures) {
    [Console]::Error.WriteLine("C12 cleanup failure [$GroupID]: $cleanupFailure")
  }
  if ($null -ne $primaryFailure -or $cleanupFailures.Count -ne 0) {
    throw "C12 group $GroupID failed"
  }
  Write-Output "C12 group $GroupID passed"
}

function Assert-C12NoInheritedDependencies {
  foreach ($name in @('TALENRO_DATABASE_URL', 'TALENRO_REDIS_ADDRESS', 'TALENRO_NATS_URL')) {
    if (-not [string]::IsNullOrEmpty([System.Environment]::GetEnvironmentVariable($name, 'Process'))) {
      throw 'inherited dependency endpoints are forbidden'
    }
  }
  foreach ($name in @('TALENRO_INSTALLATION_KIND', 'TALENRO_ENVIRONMENT')) {
    if (-not [string]::IsNullOrEmpty([System.Environment]::GetEnvironmentVariable($name, 'Process'))) {
      throw 'inherited production markers are forbidden'
    }
  }
}

function Invoke-C12FocusedMode {
  if ($Profile -cne 'base') {
    throw 'Task 4 runner accepts only the base profile'
  }
  if ($Race) {
    throw 'Task 4 base profile does not support -Race while CGO is disabled'
  }
  $duration = ConvertFrom-C12Duration -Value $Timeout
  if ($duration -gt [TimeSpan]::FromMinutes(30)) {
    throw 'focused timeout must not exceed 30m'
  }
  if ($Packages -notmatch '^\./[A-Za-z0-9_.-]+(?:/[A-Za-z0-9_.-]+)*(?:\|\./[A-Za-z0-9_.-]+(?:/[A-Za-z0-9_.-]+)*)*$' -or $Packages.Contains('./...')) {
    throw 'invalid -Packages: expected pipe-delimited explicit ./package tokens'
  }
  $packageList = @($Packages -split '\|')
  $seen = @{}
  foreach ($package in $packageList) {
    if ($seen.ContainsKey($package)) {
      throw "duplicate package $package"
    }
    $seen[$package] = $true
    if (-not $script:c12AllowedPackages.ContainsKey($package)) {
      throw "unknown package $package"
    }
  }
  $expectedTests = @()
  if (-not [string]::IsNullOrEmpty($Run)) {
    $expectedTests = @(Convert-C12RunToTests -Pattern $Run)
  }
  foreach ($package in $packageList) {
    $groupID = 'focused-' + ($package.TrimStart('.').TrimStart('/').Replace('/', '-'))
    Invoke-C12Group -GroupID $groupID -GroupProfile 'base' -Package $package -RunPattern $Run -GroupTimeout $Timeout -ExpectedTests $expectedTests
  }
}

function Invoke-C12SuiteMode {
  if ($Suite -cne 'batch01') {
    throw 'Task 4 runner accepts only the batch01 suite'
  }
  $duration = ConvertFrom-C12Duration -Value $Timeout
  if ($duration -ne [TimeSpan]::FromMinutes(120)) {
    throw 'batch01 suite timeout must be exactly 120m'
  }
  $manifestRelativePath = 'testdata/c12/integration-contracts-schema-authority.v1.json'
  $manifestPath = Join-Path $script:c12RepositoryRoot ($manifestRelativePath.Replace('/', '\'))
  if (-not (Test-Path -LiteralPath $manifestPath -PathType Leaf)) {
    throw 'canonical Batch 01 manifest is absent'
  }

  $goArgs = @(
    'test', '-run', '^TestC12SelectedIntegrationManifest$', '-count=1', './internal/testinfra',
    '-args', '-c12-suite', 'batch01', '-c12-manifests', $manifestRelativePath,
    '-c12-suite-timeout', $Timeout
  )
  $null = Invoke-C12Go -Arguments $goArgs -Stage 'validate canonical Batch 01 integration manifest'
  $manifest = Get-Content -Raw -LiteralPath $manifestPath | ConvertFrom-Json
  foreach ($group in @($manifest.groups)) {
    $tests = @($group.tests | ForEach-Object { [string]$_ })
    $runPattern = '^(' + ($tests -join '|') + ')$'
    Invoke-C12Group -GroupID ([string]$group.id) -GroupProfile ([string]$group.profile) -Package ([string]$group.package) -RunPattern $runPattern -GroupTimeout ([string]$group.timeout) -ExpectedTests $tests
  }
}

$script:c12RepositoryRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..')).Path
$selectedMode = $PSCmdlet.ParameterSetName
$priorGoEnvironment = @{
  GOOS = [System.Environment]::GetEnvironmentVariable('GOOS', 'Process')
  GOARCH = [System.Environment]::GetEnvironmentVariable('GOARCH', 'Process')
  CGO_ENABLED = [System.Environment]::GetEnvironmentVariable('CGO_ENABLED', 'Process')
}
$scriptExitCode = 0
try {
  Assert-C12NoInheritedDependencies
  [System.Environment]::SetEnvironmentVariable('GOOS', 'windows', 'Process')
  [System.Environment]::SetEnvironmentVariable('GOARCH', 'amd64', 'Process')
  [System.Environment]::SetEnvironmentVariable('CGO_ENABLED', '0', 'Process')
  if ($selectedMode -eq 'Focused') {
    Invoke-C12FocusedMode
  }
  else {
    Invoke-C12SuiteMode
  }
}
catch {
  [Console]::Error.WriteLine($_.Exception.Message)
  $scriptExitCode = 1
}
finally {
  foreach ($name in $priorGoEnvironment.Keys) {
    [System.Environment]::SetEnvironmentVariable($name, $priorGoEnvironment[$name], 'Process')
  }
}
exit $scriptExitCode
