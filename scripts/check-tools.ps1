$ErrorActionPreference = 'Stop'

function Invoke-ToolVersion {
  param(
    [string]$Tool,
    [string[]]$Arguments
  )

  $output = @(& go tool $Tool @Arguments 2>&1)
  $exitCode = $LASTEXITCODE
  if ($exitCode -ne 0) {
    throw "$Tool failed with exit code ${exitCode}: $($output -join "`n")"
  }

  return $output
}

$checks = @(
  [pscustomobject]@{ Tool = 'buf'; Arguments = @('--version'); Pattern = '^1\.72\.0$' }
  [pscustomobject]@{ Tool = 'protoc-gen-go'; Arguments = @('--version'); Pattern = '^protoc-gen-go v1\.36\.11$' }
  [pscustomobject]@{ Tool = 'oapi-codegen'; Arguments = @('--version'); Pattern = '^v2\.8\.0$' }
  [pscustomobject]@{ Tool = 'sqlc'; Arguments = @('version'); Pattern = '^v1\.31\.1$' }
  [pscustomobject]@{ Tool = 'goose'; Arguments = @('-version'); Pattern = '^goose version: v3\.27\.1$' }
  [pscustomobject]@{ Tool = 'golangci-lint'; Arguments = @('version'); Pattern = '^golangci-lint has version 2\.12\.2 built with .+$' }
)

foreach ($check in $checks) {
  $lines = @(Invoke-ToolVersion -Tool $check.Tool -Arguments $check.Arguments | ForEach-Object { $_.ToString() })
  if ($check.Tool -eq 'protoc-gen-go') {
    $lines = @($lines | ForEach-Object { $_ -replace '^protoc-gen-go\.exe ', 'protoc-gen-go ' })
  }

  $matches = @($lines | Where-Object { $_ -match $check.Pattern })
  if ($matches.Count -ne 1) {
    throw "$($check.Tool) version mismatch: $($lines -join "`n")"
  }

  Write-Output $matches[0]
}
