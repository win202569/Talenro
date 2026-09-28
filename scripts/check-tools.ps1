$devtoolsRuntime = $PSVersionTable
if ($devtoolsRuntime.PSEdition -ne 'Core' -or
    [string]$devtoolsRuntime.PSVersion -cne '7.6.5' -or
    $devtoolsRuntime.ContainsKey('PSVersionPreReleaseLabel') -or
    [Environment]::OSVersion.Platform -ne [PlatformID]::Win32NT) {
  [Console]::Error.WriteLine('check-tools: PowerShell 7.6.5 required.')
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
    $toolArguments = @('tool')
    if ($Tool -ne 'goose') {
      $toolArguments += "-modfile=$script:devtoolsMod"
    }
    $toolArguments += $Tool
    $toolArguments += $Arguments
    $output = @(& $commands[0].Source @toolArguments 2>&1 | ForEach-Object { $_.ToString() })
    $exitCode = $LASTEXITCODE
  } finally {
    $ErrorActionPreference = $previousErrorAction
  }
  if ($exitCode -ne 0) {
    Stop-Script -ExitCode $exitCode -Stage "check-tools: ${Tool} version"
  }

  return $output
}

function Invoke-DevtoolsVerification {
  $devtoolsPowerShell = [Environment]::ProcessPath
  if ([string]::IsNullOrWhiteSpace($devtoolsPowerShell) -or
      -not [IO.Path]::IsPathFullyQualified($devtoolsPowerShell) -or
      [IO.Path]::GetFileName($devtoolsPowerShell) -ine 'pwsh.exe' -or
      -not [IO.File]::Exists($devtoolsPowerShell)) {
    throw 'devtools host resolution failed'
  }
  $verifyArguments = @('-NoProfile', '-NonInteractive', '-File', (Join-Path $PSScriptRoot 'verify-devtools.ps1'))
  $previousErrorAction = $ErrorActionPreference
  $ErrorActionPreference = 'Continue'
  try {
    $global:LASTEXITCODE = 0
    $null = @(& $devtoolsPowerShell @verifyArguments 2>&1 | ForEach-Object { $_.ToString() })
    $exitCode = $LASTEXITCODE
  } finally {
    $ErrorActionPreference = $previousErrorAction
  }
  if ($exitCode -ne 0) {
    Stop-Script -ExitCode $exitCode -Stage 'check-tools: devtools verification'
  }
}

function Invoke-Main {
  $repoRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..')).Path
  $script:devtoolsMod = Join-Path $repoRoot 'tools/devtools/go.mod'
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
    Invoke-DevtoolsVerification
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

$devtoolsExecutionEnvironment = $null
try {
  . (Join-Path $PSScriptRoot 'private/devtools-process.ps1')
  Initialize-DevtoolsProcessOwnership
  $devtoolsExecutionEnvironment = Initialize-DevtoolsExecutionEnvironment
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
} finally {
  if ($null -ne $devtoolsExecutionEnvironment) { Restore-DevtoolsExecutionEnvironment -Saved $devtoolsExecutionEnvironment }
}
