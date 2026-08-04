#!/usr/bin/env bash
set -euo pipefail

run_tool() {
  local tool=$1
  shift

  local output
  if output="$(go tool "$tool" "$@" 2>&1)"; then
    printf '%s\n' "$output"
    return 0
  fi

  local exit_code=$?
  printf '%s\n' "$output" >&2
  printf '%s failed with exit code %d\n' "$tool" "$exit_code" >&2
  return "$exit_code"
}

validate_exact_line() {
  local tool=$1
  local pattern=$2
  local output=$3
  local count
  count="$(printf '%s\n' "$output" | grep -Ec "$pattern" || true)"

  if [[ "$count" != '1' ]]; then
    printf '%s version mismatch: %s\n' "$tool" "$output" >&2
    return 1
  fi

  printf '%s\n' "$output" | grep -E "$pattern"
}

buf_output="$(run_tool buf --version)"
validate_exact_line buf '^1\.72\.0$' "$buf_output"

protoc_output="$(run_tool protoc-gen-go --version)"
protoc_output="$(printf '%s\n' "$protoc_output" | sed 's/^protoc-gen-go\.exe /protoc-gen-go /')"
validate_exact_line protoc-gen-go '^protoc-gen-go v1\.36\.11$' "$protoc_output"

oapi_output="$(run_tool oapi-codegen --version)"
validate_exact_line oapi-codegen '^v2\.8\.0$' "$oapi_output"

sqlc_output="$(run_tool sqlc version)"
validate_exact_line sqlc '^v1\.31\.1$' "$sqlc_output"

goose_output="$(run_tool goose -version)"
validate_exact_line goose '^goose version: v3\.27\.1$' "$goose_output"

lint_output="$(run_tool golangci-lint version)"
validate_exact_line golangci-lint '^golangci-lint has version 2\.12\.2 built with .+$' "$lint_output"
