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
    $null = @(& $commands[0].Source @ArgumentList 2>&1 | ForEach-Object { $_.ToString() })
    $exitCode = $LASTEXITCODE
  } finally {
    $ErrorActionPreference = $previousErrorAction
  }
  if ($exitCode -ne 0) {
    Stop-Script -ExitCode $exitCode -Stage $Stage
  }
}

function Invoke-Main {
  $repoRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..')).Path

  Push-Location -LiteralPath $repoRoot
  try {
    Write-Output 'generate: buf lint'
    Invoke-External -Stage 'generate: buf lint' -FilePath 'go' -ArgumentList @('tool', 'buf', 'lint')
    Write-Output 'generate: protobuf'
    Invoke-External -Stage 'generate: protobuf' -FilePath 'go' -ArgumentList @('tool', 'buf', 'generate')
    Write-Output 'generate: OpenAPI'
    Invoke-External -Stage 'generate: OpenAPI' -FilePath 'go' -ArgumentList @(
      'tool',
      'oapi-codegen',
      '--config',
      'api/openapi/oapi-codegen.yaml',
      'api/openapi/control-api.v1.yaml'
    )
    Write-Output 'generate: SQL'
    Invoke-External -Stage 'generate: SQL' -FilePath 'go' -ArgumentList @('tool', 'sqlc', 'generate')
    Write-Output 'generate: gofmt'
    Invoke-External -Stage 'generate: gofmt' -FilePath 'gofmt' -ArgumentList @('-w', 'gen/go', 'internal/store')
  } finally {
    Pop-Location
  }
}

try {
  Invoke-Main
} catch {
  $exitCode = 1
  $stage = 'generate: internal'
  if ($_.Exception.Data.Contains('ExitCode')) {
    $exitCode = [int]$_.Exception.Data['ExitCode']
  }
  if ($_.Exception.Data.Contains('Stage')) {
    $stage = [string]$_.Exception.Data['Stage']
  }
  [Console]::Error.WriteLine("${stage} failed with exit code ${exitCode}.")
  exit $exitCode
}
