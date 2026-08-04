$ErrorActionPreference = 'Stop'

$expected = @{
  buf = '1.72.0'
  protoc = 'protoc-gen-go v1.36.11'
  oapi = 'v2.8.0'
  sqlc = 'v1.31.1'
  goose = 'v3.27.1'
  lint = '2.12.2'
}

$actual = @{
  buf = (@(go tool buf --version) -join "`n")
  protoc = (@(go tool protoc-gen-go --version) -join "`n")
  oapi = (@(go tool oapi-codegen --version) -join "`n")
  sqlc = (@(go tool sqlc version) -join "`n")
  goose = (@(go tool goose -version) -join "`n")
  lint = (@(go tool golangci-lint version) -join "`n")
}

$actual.protoc = $actual.protoc -replace '^protoc-gen-go\.exe ', 'protoc-gen-go '

foreach ($name in $expected.Keys) {
  if ($actual[$name] -notmatch [regex]::Escape($expected[$name])) {
    throw "$name version mismatch: $($actual[$name])"
  }
}
