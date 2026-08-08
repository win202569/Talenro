#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "${script_dir}/.." && pwd)"
generated_paths=(api gen internal/store)

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

printf '%s\n' 'verify: pinned tools'
run_quiet "${script_dir}/check-tools.sh"
printf '%s\n' 'verify: generation'
run_quiet "${script_dir}/generate.sh"

printf '%s\n' 'verify: generated drift'
git diff --quiet --exit-code -- "${generated_paths[@]}"
git diff --cached --quiet --exit-code -- "${generated_paths[@]}"
if [[ -n "$(git ls-files --others --exclude-standard -- "${generated_paths[@]}")" ]]; then
  printf '%s\n' 'Generated files are untracked in the verified paths.' >&2
  exit 1
fi

printf '%s\n' 'verify: gofmt and goimports'
if ! formatter_output="$(go tool golangci-lint fmt --diff 2>&1)"; then
  printf '%s\n' 'Formatter verification failed.' >&2
  exit 1
fi
if [[ -n "${formatter_output}" ]]; then
  printf '%s\n' 'gofmt or goimports reported formatting drift.' >&2
  exit 1
fi

printf '%s\n' 'verify: go vet'
run_quiet go vet ./...
printf '%s\n' 'verify: golangci-lint'
run_quiet go tool golangci-lint run ./...
printf '%s\n' 'verify: race tests'
run_quiet go test -race -count=1 ./...
printf '%s\n' 'verify: compose config'
run_quiet docker compose -f "${repo_root}/deploy/dev/compose.yaml" config --quiet
