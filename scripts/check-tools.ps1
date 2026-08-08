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

function Invoke-ToolVersion {
  param(
    [string]$Tool,
    [string[]]$Arguments
  )

  $commands = @(Get-Command -Name 'go' -CommandType Application, ExternalScript -ErrorAction SilentlyContinue)
  if ($commands.Count -eq 0) {
    Stop-Script -ExitCode 127 -Stage 'check-tools: executable resolution'
  }

  $previousErrorAction = $ErrorActionPreference
  $ErrorActionPreference = 'Continue'
  try {
    $global:LASTEXITCODE = 0
    $output = @(& $commands[0].Source tool $Tool @Arguments 2>&1 | ForEach-Object { $_.ToString() })
    $exitCode = $LASTEXITCODE
  } finally {
    $ErrorActionPreference = $previousErrorAction
  }
  if ($exitCode -ne 0) {
    Stop-Script -ExitCode $exitCode -Stage "check-tools: ${Tool} version"
  }

  return $output
}

function Invoke-Main {
  $repoRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..')).Path
  $checks = @(
    [pscustomobject]@{ Tool = 'buf'; Arguments = @('--version'); Pattern = '^1\.72\.0$' }
    [pscustomobject]@{ Tool = 'protoc-gen-go'; Arguments = @('--version'); Pattern = '^protoc-gen-go v1\.36\.11$' }
    [pscustomobject]@{ Tool = 'oapi-codegen'; Arguments = @('--version'); Pattern = '^v2\.8\.0$' }
    [pscustomobject]@{ Tool = 'sqlc'; Arguments = @('version'); Pattern = '^v1\.31\.1$' }
    [pscustomobject]@{ Tool = 'goose'; Arguments = @('-version'); Pattern = '^goose version: v3\.27\.1$' }
    [pscustomobject]@{ Tool = 'golangci-lint'; Arguments = @('version'); Pattern = '^golangci-lint has version 2\.12\.2 built with .+$' }
  )

  Push-Location -LiteralPath $repoRoot
  try {
    foreach ($check in $checks) {
      $lines = @(Invoke-ToolVersion -Tool $check.Tool -Arguments $check.Arguments)
      if ($check.Tool -eq 'protoc-gen-go') {
        $lines = @($lines | ForEach-Object { $_ -replace '^protoc-gen-go\.exe ', 'protoc-gen-go ' })
      }

      $matches = @($lines | Where-Object { $_ -match $check.Pattern })
      if ($matches.Count -ne 1) {
        Stop-Script -ExitCode 1 -Stage "check-tools: $($check.Tool) version mismatch"
      }
      Write-Output $matches[0]
    }
  } finally {
    Pop-Location
  }
}

try {
  Invoke-Main
} catch {
  $exitCode = 1
  $stage = 'check-tools: internal'
  if ($_.Exception.Data.Contains('ExitCode')) {
    $exitCode = [int]$_.Exception.Data['ExitCode']
  }
  if ($_.Exception.Data.Contains('Stage')) {
    $stage = [string]$_.Exception.Data['Stage']
  }
  [Console]::Error.WriteLine("${stage} failed with exit code ${exitCode}.")
  exit $exitCode
}
