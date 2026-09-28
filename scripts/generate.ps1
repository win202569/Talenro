$devtoolsRuntime = $PSVersionTable
if ($devtoolsRuntime.PSEdition -ne 'Core' -or
    [string]$devtoolsRuntime.PSVersion -cne '7.6.5' -or
    $devtoolsRuntime.ContainsKey('PSVersionPreReleaseLabel') -or
    [Environment]::OSVersion.Platform -ne [PlatformID]::Win32NT) {
  [Console]::Error.WriteLine('generate: PowerShell 7.6.5 required.')
  exit 1
}

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
  $devtoolsMod = Join-Path $repoRoot 'tools/devtools/go.mod'
  $devtoolsPowerShell = [Environment]::ProcessPath
  if ([string]::IsNullOrWhiteSpace($devtoolsPowerShell) -or
      -not [IO.Path]::IsPathFullyQualified($devtoolsPowerShell) -or
      [IO.Path]::GetFileName($devtoolsPowerShell) -ine 'pwsh.exe' -or
      -not [IO.File]::Exists($devtoolsPowerShell)) {
    throw 'devtools host resolution failed'
  }
  $verifyArguments = @('-NoProfile', '-NonInteractive', '-File', (Join-Path $PSScriptRoot 'verify-devtools.ps1'))

  Push-Location -LiteralPath $repoRoot
  try {
    Invoke-External -Stage 'generate: devtools verification' -FilePath $devtoolsPowerShell -ArgumentList $verifyArguments
    Write-Output 'generate: buf lint'
    Invoke-External -Stage 'generate: buf lint' -FilePath 'go' -ArgumentList @('tool', "-modfile=$devtoolsMod", 'buf', 'lint')
    Write-Output 'generate: protobuf'
    Invoke-External -Stage 'generate: protobuf' -FilePath 'go' -ArgumentList @('tool', "-modfile=$devtoolsMod", 'buf', 'generate')
    Write-Output 'generate: OpenAPI'
    Invoke-External -Stage 'generate: OpenAPI' -FilePath 'go' -ArgumentList @(
      'tool',
      "-modfile=$devtoolsMod",
      'oapi-codegen',
      '--config',
      'api/openapi/oapi-codegen.yaml',
      'api/openapi/control-api.v1.yaml'
    )
    Invoke-External -Stage 'generate: node bootstrap OpenAPI' -FilePath 'go' -ArgumentList @(
      'tool',
      "-modfile=$devtoolsMod",
      'oapi-codegen',
      '--config',
      'api/openapi/node-bootstrap-oapi-codegen.yaml',
      'api/openapi/node-bootstrap-api.v1.yaml'
    )
    Invoke-External -Stage 'generate: node agent OpenAPI' -FilePath 'go' -ArgumentList @(
      'tool',
      "-modfile=$devtoolsMod",
      'oapi-codegen',
      '--config',
      'api/openapi/node-agent-oapi-codegen.yaml',
      'api/openapi/node-agent-api.v1.yaml'
    )
    Invoke-External -Stage 'generate: node operator OpenAPI' -FilePath 'go' -ArgumentList @(
      'tool',
      "-modfile=$devtoolsMod",
      'oapi-codegen',
      '--config',
      'api/openapi/node-operator-oapi-codegen.yaml',
      'api/openapi/node-operator-api.v1.yaml'
    )
    Write-Output 'generate: SQL'
    Invoke-External -Stage 'generate: SQL' -FilePath 'go' -ArgumentList @('tool', "-modfile=$devtoolsMod", 'sqlc', 'generate')
    Write-Output 'generate: gofmt'
    Invoke-External -Stage 'generate: gofmt' -FilePath 'gofmt' -ArgumentList @('-w', 'gen/go', 'internal/store')
  } finally {
    Pop-Location
  }
}

try {
  . (Join-Path $PSScriptRoot 'private/devtools-process.ps1')
  Initialize-DevtoolsProcessOwnership
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
