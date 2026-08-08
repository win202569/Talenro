#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "${script_dir}/.." && pwd)"
generated_paths=(api gen internal/store)

capture_quiet() {
  local target=$1
  local stage=$2
  shift 2
  local captured
  local exit_code

  set +e
  captured="$("$@" 2>&1)"
  exit_code=$?
  set -e

  if [[ "${exit_code}" != 0 ]]; then
    printf '%s failed with exit code %d.\n' "${stage}" "${exit_code}" >&2
    return "${exit_code}"
  fi

  printf -v "${target}" '%s' "${captured}"
}

run_quiet() {
  local stage=$1
  shift
  local output
  capture_quiet output "${stage}" "$@"
}

cd "${repo_root}"

printf '%s\n' 'verify: pinned tools'
run_quiet 'verify: pinned tools' "${script_dir}/check-tools.sh"
printf '%s\n' 'verify: generation'
run_quiet 'verify: generation' "${script_dir}/generate.sh"

printf '%s\n' 'verify: generated drift'
run_quiet 'verify: generated worktree drift' git diff --quiet --exit-code -- "${generated_paths[@]}"
run_quiet 'verify: generated index drift' git diff --cached --quiet --exit-code -- "${generated_paths[@]}"
untracked=''
capture_quiet untracked 'verify: generated inventory' git ls-files --others --exclude-standard -- "${generated_paths[@]}"
if [[ -n "${untracked}" ]]; then
  printf '%s\n' 'verify: generated inventory failed with exit code 1.' >&2
  exit 1
fi

printf '%s\n' 'verify: gofmt and goimports'
formatter_output=''
capture_quiet formatter_output 'verify: gofmt and goimports' go tool golangci-lint fmt --diff
if [[ -n "${formatter_output}" ]]; then
  printf '%s\n' 'verify: gofmt and goimports drift failed with exit code 1.' >&2
  exit 1
fi

printf '%s\n' 'verify: go vet'
run_quiet 'verify: go vet' go vet ./...
printf '%s\n' 'verify: golangci-lint'
run_quiet 'verify: golangci-lint' go tool golangci-lint run ./...
printf '%s\n' 'verify: race tests'
run_quiet 'verify: race tests' go test -race -count=1 ./...
printf '%s\n' 'verify: compose config'
run_quiet 'verify: compose config' docker compose -f "${repo_root}/deploy/dev/compose.yaml" config --quiet
