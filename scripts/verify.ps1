$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

function Stop-Script {
  param(
    [Parameter(Mandatory)]
    [int]$ExitCode,

    [Parameter(Mandatory)]
    [string]$Stage
  )

  $failure = [System.Exception]::new('sanitized script failure')
  $failure.Data['ExitCode'] = $ExitCode
  $failure.Data['Stage'] = $Stage
  throw $failure
}

function Invoke-External {
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
    Stop-Script -ExitCode 127 -Stage $Stage
  }

  $previousErrorAction = $ErrorActionPreference
  $ErrorActionPreference = 'Continue'
  try {
    $global:LASTEXITCODE = 0
    $output = @(& $commands[0].Source @ArgumentList 2>&1 | ForEach-Object { $_.ToString() })
    $exitCode = $LASTEXITCODE
  } finally {
    $ErrorActionPreference = $previousErrorAction
  }
  if ($exitCode -ne 0) {
    Stop-Script -ExitCode $exitCode -Stage $Stage
  }

  return $output
}

function Assert-NoGeneratedDrift {
  $generatedPaths = @('api', 'gen', 'internal/store')

  $null = Invoke-External -Stage 'verify: generated worktree drift' -FilePath 'git' -ArgumentList (
    @('diff', '--quiet', '--exit-code', '--') + $generatedPaths
  )
  $null = Invoke-External -Stage 'verify: generated index drift' -FilePath 'git' -ArgumentList (
    @('diff', '--cached', '--quiet', '--exit-code', '--') + $generatedPaths
  )
  $untracked = @(Invoke-External -Stage 'verify: generated inventory' -FilePath 'git' -ArgumentList (
    @('ls-files', '--others', '--exclude-standard', '--') + $generatedPaths
  ))
  if ($untracked.Count -ne 0) {
    Stop-Script -ExitCode 1 -Stage 'verify: generated inventory'
  }
}

function Assert-FormatterClean {
  $output = @(Invoke-External -Stage 'verify: gofmt and goimports' -FilePath 'go' -ArgumentList @(
    'tool', 'golangci-lint', 'fmt', '--diff'
  ))
  if ($output.Count -ne 0) {
    Stop-Script -ExitCode 1 -Stage 'verify: gofmt and goimports drift'
  }
}

function Invoke-Main {
  $repoRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..')).Path
  $powerShell = Join-Path $PSHOME 'powershell.exe'

  Push-Location -LiteralPath $repoRoot
  try {
    Write-Output 'verify: pinned tools'
    $null = Invoke-External -Stage 'verify: pinned tools' -FilePath $powerShell -ArgumentList @(
      '-NoProfile',
      '-ExecutionPolicy',
      'Bypass',
      '-File',
      (Join-Path $PSScriptRoot 'check-tools.ps1')
    )
    Write-Output 'verify: generation'
    $null = Invoke-External -Stage 'verify: generation' -FilePath $powerShell -ArgumentList @(
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
    $null = Invoke-External -Stage 'verify: go vet' -FilePath 'go' -ArgumentList @('vet', './...')
    Write-Output 'verify: golangci-lint'
    $null = Invoke-External -Stage 'verify: golangci-lint' -FilePath 'go' -ArgumentList @(
      'tool', 'golangci-lint', 'run', './...'
    )
    Write-Output 'verify: race tests'
    $null = Invoke-External -Stage 'verify: race tests' -FilePath 'go' -ArgumentList @(
      'test', '-race', '-count=1', './...'
    )
    Write-Output 'verify: compose config'
    $null = Invoke-External -Stage 'verify: compose config' -FilePath 'docker' -ArgumentList @(
      'compose',
      '-f',
      (Join-Path $repoRoot 'deploy/dev/compose.yaml'),
      'config',
      '--quiet'
    )
  } finally {
    Pop-Location
  }
}

try {
  Invoke-Main
} catch {
  $exitCode = 1
  $stage = 'verify: internal'
  if ($_.Exception.Data.Contains('ExitCode')) {
    $exitCode = [int]$_.Exception.Data['ExitCode']
  }
  if ($_.Exception.Data.Contains('Stage')) {
    $stage = [string]$_.Exception.Data['Stage']
  }
  [Console]::Error.WriteLine("${stage} failed with exit code ${exitCode}.")
  exit $exitCode
}
