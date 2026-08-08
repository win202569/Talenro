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

$repoRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..')).Path

Push-Location -LiteralPath $repoRoot
try {
  Write-Output 'generate: buf lint'
  Invoke-External -FilePath 'go' -ArgumentList @('tool', 'buf', 'lint')
  Write-Output 'generate: protobuf'
  Invoke-External -FilePath 'go' -ArgumentList @('tool', 'buf', 'generate')
  Write-Output 'generate: OpenAPI'
  Invoke-External -FilePath 'go' -ArgumentList @(
    'tool',
    'oapi-codegen',
    '--config',
    'api/openapi/oapi-codegen.yaml',
    'api/openapi/control-api.v1.yaml'
  )
  Write-Output 'generate: SQL'
  Invoke-External -FilePath 'go' -ArgumentList @('tool', 'sqlc', 'generate')
  Write-Output 'generate: gofmt'
  Invoke-External -FilePath 'gofmt' -ArgumentList @('-w', 'gen/go', 'internal/store')
} finally {
  Pop-Location
}
