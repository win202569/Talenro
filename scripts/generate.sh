#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "${script_dir}/.." && pwd)"

run_quiet() {
  local stage=$1
  shift
  local output
  local exit_code

  set +e
  output="$("$@" 2>&1)"
  exit_code=$?
  set -e

  if [[ "${exit_code}" != 0 ]]; then
    printf '%s failed with exit code %d.\n' "${stage}" "${exit_code}" >&2
    return "${exit_code}"
  fi
}

cd "${repo_root}"

printf '%s\n' 'generate: buf lint'
run_quiet 'generate: buf lint' go tool buf lint
printf '%s\n' 'generate: protobuf'
run_quiet 'generate: protobuf' go tool buf generate
printf '%s\n' 'generate: OpenAPI'
run_quiet 'generate: OpenAPI' go tool oapi-codegen --config api/openapi/oapi-codegen.yaml api/openapi/control-api.v1.yaml
run_quiet 'generate: node bootstrap OpenAPI' go tool oapi-codegen --config api/openapi/node-bootstrap-oapi-codegen.yaml api/openapi/node-bootstrap-api.v1.yaml
run_quiet 'generate: node agent OpenAPI' go tool oapi-codegen --config api/openapi/node-agent-oapi-codegen.yaml api/openapi/node-agent-api.v1.yaml
run_quiet 'generate: node operator OpenAPI' go tool oapi-codegen --config api/openapi/node-operator-oapi-codegen.yaml api/openapi/node-operator-api.v1.yaml
printf '%s\n' 'generate: SQL'
run_quiet 'generate: SQL' go tool sqlc generate
printf '%s\n' 'generate: gofmt'
run_quiet 'generate: gofmt' gofmt -w gen/go internal/store
