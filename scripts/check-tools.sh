#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "${script_dir}/.." && pwd)"

run_tool() {
  local tool=$1
  shift
  local output
  local exit_code

  set +e
  output="$(go tool "${tool}" "$@" 2>&1)"
  exit_code=$?
  set -e

  if [[ "${exit_code}" != 0 ]]; then
    printf 'check-tools: %s version failed with exit code %d.\n' "${tool}" "${exit_code}" >&2
    return "${exit_code}"
  fi

  printf '%s\n' "${output}"
}

validate_exact_line() {
  local tool=$1
  local pattern=$2
  local output=$3
  local count=0
  local line
  local match=''

  while IFS= read -r line; do
    if [[ "${line}" =~ ${pattern} ]]; then
      ((count += 1))
      match=${line}
    fi
  done <<<"${output}"

  if [[ "${count}" != '1' ]]; then
    printf 'check-tools: %s version mismatch failed with exit code 1.\n' "${tool}" >&2
    return 1
  fi

  printf '%s\n' "${match}"
}

cd "${repo_root}"

buf_output="$(run_tool buf --version)"
validate_exact_line buf '^1\.72\.0$' "${buf_output}"

protoc_output="$(run_tool protoc-gen-go --version)"
validate_exact_line protoc-gen-go '^protoc-gen-go(\.exe)? v1\.36\.11$' "${protoc_output}"

oapi_output="$(run_tool oapi-codegen --version)"
validate_exact_line oapi-codegen '^v2\.8\.0$' "${oapi_output}"

sqlc_output="$(run_tool sqlc version)"
validate_exact_line sqlc '^v1\.31\.1$' "${sqlc_output}"

goose_output="$(run_tool goose -version)"
validate_exact_line goose '^goose version: v3\.27\.1$' "${goose_output}"

lint_output="$(run_tool golangci-lint version)"
validate_exact_line golangci-lint '^golangci-lint has version 2\.12\.2 built with .+$' "${lint_output}"
