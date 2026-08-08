#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "${script_dir}/.." && pwd)"

run_quiet() {
  local output
  local exit_code

  if output="$("$@" 2>&1)"; then
    return 0
  else
    exit_code=$?
  fi

  printf 'External command failed with exit code %d.\n' "${exit_code}" >&2
  return "${exit_code}"
}

cd "${repo_root}"

printf '%s\n' 'generate: buf lint'
run_quiet go tool buf lint
printf '%s\n' 'generate: protobuf'
run_quiet go tool buf generate
printf '%s\n' 'generate: OpenAPI'
run_quiet go tool oapi-codegen --config api/openapi/oapi-codegen.yaml api/openapi/control-api.v1.yaml
printf '%s\n' 'generate: SQL'
run_quiet go tool sqlc generate
printf '%s\n' 'generate: gofmt'
run_quiet gofmt -w gen/go internal/store
