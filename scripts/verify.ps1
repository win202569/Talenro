$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

function Invoke-External {
  param(
    [Parameter(Mandatory)]
    [string]$FilePath,

    [Parameter(Mandatory)]
    [string[]]$ArgumentList
  )

  $command = Get-Command -Name $FilePath -CommandType Application, ExternalScript -ErrorAction Stop |
    Select-Object -First 1
  $previousErrorAction = $ErrorActionPreference
  $ErrorActionPreference = 'Continue'
  try {
    $null = @(& $command.Source @ArgumentList 2>&1 | ForEach-Object { $_.ToString() })
    $exitCode = $LASTEXITCODE
  } finally {
    $ErrorActionPreference = $previousErrorAction
  }
  if ($exitCode -ne 0) {
    throw "External command failed with exit code ${exitCode}."
  }
}

function Assert-NoGeneratedDrift {
  $generatedPaths = @('api', 'gen', 'internal/store')

  Invoke-External -FilePath 'git' -ArgumentList (@('diff', '--quiet', '--exit-code', '--') + $generatedPaths)
  Invoke-External -FilePath 'git' -ArgumentList (@('diff', '--cached', '--quiet', '--exit-code', '--') + $generatedPaths)

  $untracked = @(& git ls-files --others --exclude-standard -- @generatedPaths)
  $exitCode = $LASTEXITCODE
  if ($exitCode -ne 0) {
    throw "Generated-file inventory failed with exit code ${exitCode}."
  }
  if ($untracked.Count -ne 0) {
    throw 'Generated files are untracked in the verified paths.'
  }
}

function Assert-FormatterClean {
  $output = @(& go tool golangci-lint fmt --diff 2>&1)
  $exitCode = $LASTEXITCODE
  if ($exitCode -ne 0) {
    throw "Formatter verification failed with exit code ${exitCode}."
  }
  if ($output.Count -ne 0) {
    throw 'gofmt or goimports reported formatting drift.'
  }
}

$repoRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..')).Path
$powerShell = Join-Path $PSHOME 'powershell.exe'

Push-Location -LiteralPath $repoRoot
try {
  Write-Output 'verify: pinned tools'
  Invoke-External -FilePath $powerShell -ArgumentList @(
    '-NoProfile',
    '-ExecutionPolicy',
    'Bypass',
    '-File',
    (Join-Path $PSScriptRoot 'check-tools.ps1')
  )
  Write-Output 'verify: generation'
  Invoke-External -FilePath $powerShell -ArgumentList @(
    '-NoProfile',
    '-ExecutionPolicy',
    'Bypass',
    '-File',
    (Join-Path $PSScriptRoot 'generate.ps1')
  )
  Write-Output 'verify: generated drift'
  Assert-NoGeneratedDrift
  Write-Output 'verify: gofmt and goimports'
  Assert-FormatterClean
  Write-Output 'verify: go vet'
  Invoke-External -FilePath 'go' -ArgumentList @('vet', './...')
  Write-Output 'verify: golangci-lint'
  Invoke-External -FilePath 'go' -ArgumentList @('tool', 'golangci-lint', 'run', './...')
  Write-Output 'verify: race tests'
  Invoke-External -FilePath 'go' -ArgumentList @('test', '-race', '-count=1', './...')
  Write-Output 'verify: compose config'
  Invoke-External -FilePath 'docker' -ArgumentList @(
    'compose',
    '-f',
    (Join-Path $repoRoot 'deploy/dev/compose.yaml'),
    'config',
    '--quiet'
  )
} finally {
  Pop-Location
}
