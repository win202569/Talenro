#!/usr/bin/env bash
set -euo pipefail

go tool buf --version | grep -F '1.72.0'
go tool protoc-gen-go --version | grep -F 'v1.36.11'
go tool oapi-codegen --version | grep -F 'v2.8.0'
go tool sqlc version | grep -F 'v1.31.1'
go tool goose -version | grep -F 'v3.27.1'
go tool golangci-lint version | grep -F '2.12.2'
