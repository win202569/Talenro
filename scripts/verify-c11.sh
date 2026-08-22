#!/usr/bin/env bash
set -euo pipefail

script_source=${BASH_SOURCE[0]}
case "${script_source}" in
  */*) script_parent=${script_source%/*} ;;
  *) script_parent=. ;;
esac
if ! script_dir="$(cd -- "${script_parent}" 2>/dev/null && pwd -P)" ||
   ! repo_root="$(cd -- "${script_dir}/.." 2>/dev/null && pwd -P)"; then
  printf '%s\n' 'verify-c11: repository resolution failed with exit code 2.' >&2
  exit 2
fi

run_dir=''
run_counter=0
compose_touched=0
compose_project=''
compose_override=''
conformance_binary=''
declare -a compose=()
captured_output=''

create_run_directory() {
  local requested_root=${TMPDIR:-/tmp}
  local canonical_root candidate
  if ! canonical_root="$(cd -- "${requested_root}" 2>/dev/null && pwd -P)"; then
    printf '%s\n' 'verify-c11: runtime ownership failed with exit code 1.' >&2
    return 1
  fi
  if ! candidate="$(mktemp -d "${canonical_root}/talenro-verify-c11.XXXXXXXX" 2>/dev/null)" ||
     ! run_dir="$(cd -- "${candidate}" 2>/dev/null && pwd -P)" ||
     [[ "${run_dir}" != "${canonical_root}"/talenro-verify-c11.* ]]; then
    [[ -z "${candidate:-}" ]] || rmdir -- "${candidate}" >/dev/null 2>&1 || true
    printf '%s\n' 'verify-c11: runtime ownership failed with exit code 1.' >&2
    return 1
  fi
}

invoke_quiet() {
  local seconds=$1
  shift
  local stdout_path stderr_path status cleanup_status=0
  ((run_counter += 1))
  printf -v stdout_path '%s/command-%03d.stdout' "${run_dir}" "${run_counter}"
  printf -v stderr_path '%s/command-%03d.stderr' "${run_dir}" "${run_counter}"
  if timeout --kill-after=5s "${seconds}s" "$@" >"${stdout_path}" 2>"${stderr_path}"; then
    status=0
  else
    status=$?
  fi
  captured_output=''
  rm -f -- "${stdout_path}" "${stderr_path}" >/dev/null 2>&1 || cleanup_status=1
  if ((status != 0)); then
    return "${status}"
  fi
  if ((cleanup_status != 0)); then
    return 1
  fi
  return "${status}"
}

capture_quiet() {
  local seconds=$1
  shift
  local stdout_path stderr_path status size cleanup_status=0
  ((run_counter += 1))
  printf -v stdout_path '%s/command-%03d.stdout' "${run_dir}" "${run_counter}"
  printf -v stderr_path '%s/command-%03d.stderr' "${run_dir}" "${run_counter}"
  if timeout --kill-after=5s "${seconds}s" "$@" >"${stdout_path}" 2>"${stderr_path}"; then
    status=0
  else
    status=$?
  fi
  if ((status == 0)); then
    size=$(wc -c <"${stdout_path}")
    if ((size > 4194304)); then
      status=1
    else
      captured_output=$(<"${stdout_path}")
      while [[ "${captured_output}" == *$'\n' || "${captured_output}" == *$'\r' ]]; do
        captured_output=${captured_output%?}
      done
    fi
  fi
  rm -f -- "${stdout_path}" "${stderr_path}" >/dev/null 2>&1 || cleanup_status=1
  if ((status != 0)); then
    return "${status}"
  fi
  if ((cleanup_status != 0)); then
    return 1
  fi
  return "${status}"
}

run_quiet() {
  local stage=$1
  local seconds=$2
  shift 2
  local status
  printf 'verify-c11: %s\n' "${stage}"
  if invoke_quiet "${seconds}" "$@"; then
    return 0
  else
    status=$?
  fi
  printf 'verify-c11: %s failed with exit code %d.\n' "${stage}" "${status}" >&2
  return "${status}"
}

generated_snapshot() {
  local destination=$1
  local find_output sorted_output
  local status
  case "${destination}" in
    "${run_dir}/generated.before"|"${run_dir}/generated.after") ;;
    *)
      printf '%s\n' 'verify-c11: generated inventory failed with exit code 1.' >&2
      return 1
      ;;
  esac
  find_output=${destination}.find
  sorted_output=${destination}.sorted
  if timeout --kill-after=5s 60s bash -c '
    set -euo pipefail
    root=$1
    output=$2
    find_output=$3
    sorted_output=$4
    cleanup_snapshot_files() {
      local original_exit=$?
      trap - EXIT
      if ! rm -f -- "${find_output}" "${sorted_output}" >/dev/null 2>&1 &&
         ((original_exit == 0)); then
        original_exit=1
      fi
      exit "${original_exit}"
    }
    trap cleanup_snapshot_files EXIT
    for generated_root in api gen internal/store; do
      [[ -d "${root}/${generated_root}" ]]
    done
    : >"${output}"
    if find "${root}/api" "${root}/gen" "${root}/internal/store" -mindepth 1 -print0 >"${find_output}"; then
      :
    else
      status=$?
      exit "${status}"
    fi
    if sort -z -- "${find_output}" >"${sorted_output}"; then
      :
    else
      status=$?
      exit "${status}"
    fi
    while IFS= read -r -d "" path; do
      relative=${path#"${root}/"}
      if [[ -L "${path}" ]]; then
        exit 1
      fi
      if [[ -d "${path}" ]]; then
        printf "directory\\0%s\\0" "${relative}" >>"${output}"
        continue
      fi
      [[ -f "${path}" ]] || exit 1
      digest=$(git -C "${root}" hash-object --no-filters -- "${path}")
      printf "worktree\\0%s\\0%s\\0" "${relative}" "${digest}" >>"${output}"
    done <"${sorted_output}"
    printf "index\\0" >>"${output}"
    git -C "${root}" ls-files --stage -z -- api gen internal/store >>"${output}"
  ' bash "${repo_root}" "${destination}" "${find_output}" "${sorted_output}" >/dev/null 2>&1; then
    return 0
  else
    status=$?
  fi
  printf 'verify-c11: generated inventory failed with exit code %d.\n' "${status}" >&2
  return "${status}"
}

clear_ambient_environment() {
  local name
  while IFS= read -r name; do
    case "${name}" in
      TALENRO_*|C11_*|COMPOSE_*) unset "${name}" ;;
    esac
  done < <(compgen -v)
}

create_compose_override() {
  compose_override=${run_dir}/compose.ephemeral.yaml
  printf '%s\n' \
    'services:' \
    '  postgres:' \
    '    image: postgres:18.4-alpine3.23' \
    '    environment:' \
    '      POSTGRES_DB: talenro' \
    '      POSTGRES_USER: talenro' \
    '      POSTGRES_PASSWORD: talenro_dev' \
    '    ports:' \
    '      - target: 5432' \
    '        published: "0"' \
    '        host_ip: 127.0.0.1' \
    '        protocol: tcp' \
    '    healthcheck:' \
    '      test: ["CMD-SHELL", "pg_isready -U talenro -d talenro"]' \
    '      interval: 2s' \
    '      timeout: 2s' \
    '      retries: 20' \
    '    volumes:' \
    '      - type: tmpfs' \
    '        target: /var/lib/postgresql/18/docker' \
    '  redis:' \
    '    image: redis:8.8.1-alpine3.23' \
    '    command: ["redis-server", "--appendonly", "yes", "--save", "60", "1"]' \
    '    ports:' \
    '      - target: 6379' \
    '        published: "0"' \
    '        host_ip: 127.0.0.1' \
    '        protocol: tcp' \
    '    healthcheck:' \
    '      test: ["CMD", "redis-cli", "ping"]' \
    '      interval: 2s' \
    '      timeout: 2s' \
    '      retries: 20' \
    '    volumes:' \
    '      - type: tmpfs' \
    '        target: /data' \
    '  nats:' \
    '    image: nats:2.14.3-alpine3.22' \
    '    command: ["-c", "/etc/nats/nats.conf"]' \
    '    ports:' \
    '      - target: 4222' \
    '        published: "0"' \
    '        host_ip: 127.0.0.1' \
    '        protocol: tcp' \
    '    healthcheck:' \
    '      test: ["CMD", "wget", "-q", "-O", "-", "http://127.0.0.1:8222/healthz?js-enabled-only=true"]' \
    '      interval: 2s' \
    '      timeout: 2s' \
    '      retries: 20' \
    '    volumes:' \
    '      - type: bind' \
    '        source: ./nats.conf' \
    '        target: /etc/nats/nats.conf' \
    '        read_only: true' \
    '      - type: tmpfs' \
    '        target: /data' >"${compose_override}"
  printf '%s\n' \
    'server_name: talenro-c11' \
    'http: 8222' \
    '' \
    'jetstream {' \
    '  store_dir: "/data/jetstream"' \
    '  max_mem_store: 256MB' \
    '  max_file_store: 1GB' \
    '}' >"${run_dir}/nats.conf"
}

assert_compose_ownership() {
  local service=$1
  local container_id owner status
  if capture_quiet 30 "${compose[@]}" ps -q "${service}"; then
    :
  else
    status=$?
    printf 'verify-c11: dependency ownership failed with exit code %d.\n' "${status}" >&2
    return "${status}"
  fi
  container_id=${captured_output}
  if [[ ! "${container_id}" =~ ^[0-9a-f]{12,64}$ ]]; then
    printf '%s\n' 'verify-c11: dependency ownership failed with exit code 1.' >&2
    return 1
  fi
  if capture_quiet 30 docker inspect --format '{{ index .Config.Labels "com.docker.compose.project" }}' "${container_id}"; then
    :
  else
    status=$?
    printf 'verify-c11: dependency ownership failed with exit code %d.\n' "${status}" >&2
    return "${status}"
  fi
  owner=${captured_output}
  if [[ "${owner}" != "${compose_project}" ]]; then
    printf '%s\n' 'verify-c11: dependency ownership failed with exit code 1.' >&2
    return 1
  fi
}

compose_port() {
  local service=$1
  local container_port=$2
  local status endpoint port
  if capture_quiet 30 "${compose[@]}" port "${service}" "${container_port}"; then
    :
  else
    status=$?
    printf 'verify-c11: dependency endpoints failed with exit code %d.\n' "${status}" >&2
    return "${status}"
  fi
  endpoint=${captured_output}
  if [[ ! "${endpoint}" =~ ^127\.0\.0\.1:([1-9][0-9]{0,4})$ ]]; then
    printf '%s\n' 'verify-c11: dependency endpoints failed with exit code 1.' >&2
    return 1
  fi
  port=${BASH_REMATCH[1]}
  if ((10#${port} > 65535)); then
    printf '%s\n' 'verify-c11: dependency endpoints failed with exit code 1.' >&2
    return 1
  fi
  captured_output=${port}
}

start_dependencies() {
  local status postgres_port redis_port nats_port service
  compose_project="talenro-c11-verify-$(printf '%08x%04x' "$$" "${RANDOM}")"
  if [[ ! "${compose_project}" =~ ^talenro-c11-verify-[0-9a-f]{12}$ ]]; then
    printf '%s\n' 'verify-c11: dependency ownership failed with exit code 1.' >&2
    return 1
  fi
  compose=(docker compose --project-name "${compose_project}" \
    -f "${compose_override}")
  compose_touched=1
  if invoke_quiet 180 "${compose[@]}" up -d --wait --wait-timeout 120 postgres redis nats; then
    :
  else
    status=$?
    printf 'verify-c11: dependency startup failed with exit code %d.\n' "${status}" >&2
    return "${status}"
  fi
  for service in postgres redis nats; do
    assert_compose_ownership "${service}"
  done
  compose_port postgres 5432
  postgres_port=${captured_output}
  compose_port redis 6379
  redis_port=${captured_output}
  compose_port nats 4222
  nats_port=${captured_output}

  export TALENRO_DATABASE_URL="postgres://talenro:talenro_dev@127.0.0.1:${postgres_port}/talenro?sslmode=disable"
  export TALENRO_REDIS_ADDRESS="127.0.0.1:${redis_port}"
  export TALENRO_NATS_URL="nats://127.0.0.1:${nats_port}"
  export C11_E2E_COMPOSE_PROJECT="${compose_project}"
  export C11_E2E_DATABASE_URL="${TALENRO_DATABASE_URL}"
  export C11_E2E_REDIS_ADDRESS="${TALENRO_REDIS_ADDRESS}"
  export C11_E2E_NATS_URL="${TALENRO_NATS_URL}"
}

stop_dependencies() {
  local status
  ((compose_touched)) || return 0
  if invoke_quiet 60 "${compose[@]}" down --remove-orphans --timeout 20; then
    compose_touched=0
    return 0
  else
    status=$?
  fi
  printf 'verify-c11: dependency cleanup failed with exit code %d.\n' "${status}" >&2
  return "${status}"
}

validate_conformance_binary() {
  local expected
  expected=${run_dir}/trust-conformance
  case "${OSTYPE:-}" in
    msys*|cygwin*) expected=${expected}.exe ;;
  esac
  if [[ "${conformance_binary}" != "${expected}" || "${conformance_binary}" != /* ||
        ! -f "${conformance_binary}" || -L "${conformance_binary}" ||
        ! -s "${conformance_binary}" || ! -x "${conformance_binary}" ]]; then
    printf '%s\n' 'verify-c11: conformance binary ownership failed with exit code 1.' >&2
    return 1
  fi
}

run_e2e_tests() {
  local name previous='' previous_set=0 status
  for name in C11_E2E_COMPOSE_PROJECT C11_E2E_DATABASE_URL C11_E2E_REDIS_ADDRESS C11_E2E_NATS_URL; do
    if [[ -z "${!name:-}" ]]; then
      printf '%s\n' 'verify-c11: e2e environment failed with exit code 1.' >&2
      return 1
    fi
  done
  if [[ ${C11_CONFORMANCE_BINARY+x} ]]; then
    previous=${C11_CONFORMANCE_BINARY}
    previous_set=1
  fi
  export C11_CONFORMANCE_BINARY=${conformance_binary}
  if run_quiet 'e2e tests' 600 go test -tags=e2e ./internal/e2e -count=1 -timeout 10m; then
    status=0
  else
    status=$?
  fi
  if ((previous_set)); then
    export C11_CONFORMANCE_BINARY=${previous}
  else
    unset C11_CONFORMANCE_BINARY
  fi
  return "${status}"
}

remove_run_directory() {
  local canonical_root expected file
  [[ -n "${run_dir}" ]] || return 0
  canonical_root=$(cd -- "${TMPDIR:-/tmp}" 2>/dev/null && pwd -P) || return 1
  expected=${canonical_root}/talenro-verify-c11.
  [[ "${run_dir}" == "${expected}"* ]] || return 1
  for file in "${run_dir}"/* "${run_dir}"/.[!.]* "${run_dir}"/..?*; do
    [[ -e "${file}" ]] || continue
    [[ -f "${file}" && "${file}" == "${run_dir}/"* ]] || return 1
    rm -f -- "${file}" >/dev/null 2>&1 || return 1
  done
  rmdir -- "${run_dir}" >/dev/null 2>&1 || return 1
  run_dir=''
}

cleanup() {
  local original_exit=$?
  local cleanup_exit=0
  local status
  trap - EXIT INT TERM
  set +e
  stop_dependencies
  status=$?
  if (( status != 0 )); then
    cleanup_exit=${status}
  fi
  remove_run_directory
  if (( $? != 0 && cleanup_exit == 0 )); then
    cleanup_exit=1
    printf '%s\n' 'verify-c11: runtime cleanup failed with exit code 1.' >&2
  fi
  if ((original_exit != 0)); then
    exit "${original_exit}"
  fi
  exit "${cleanup_exit}"
}

command -v timeout >/dev/null 2>&1 || {
  printf '%s\n' 'verify-c11: hard deadline tool failed with exit code 127.' >&2
  exit 127
}
create_run_directory
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
create_compose_override
clear_ambient_environment
export CGO_ENABLED=0
cd -- "${repo_root}"

run_quiet 'check tools' 900 "${script_dir}/check-tools.sh"
generated_before=${run_dir}/generated.before
generated_after=${run_dir}/generated.after
generated_snapshot "${generated_before}"
run_quiet 'generate' 600 "${script_dir}/generate.sh"
generated_snapshot "${generated_after}"
printf '%s\n' 'verify-c11: generated diff'
if ! cmp -s -- "${generated_before}" "${generated_after}"; then
  printf '%s\n' 'verify-c11: generated diff failed with exit code 1.' >&2
  exit 1
fi

run_quiet 'unit tests' 600 go test ./... -count=1 -timeout 10m
run_quiet 'fuzz FuzzDecode' 120 go test -run '^$' -fuzz '^FuzzDecode$' -fuzztime=10s -timeout 30s ./internal/strictjson
run_quiet 'fuzz FuzzEnvelopeSeal' 120 go test -run '^$' -fuzz '^FuzzEnvelopeSeal$' -fuzztime=10s -timeout 30s ./internal/trust
run_quiet 'fuzz FuzzVerifyEnvelope' 120 go test -run '^$' -fuzz '^FuzzVerifyEnvelope$' -fuzztime=10s -timeout 30s ./internal/trustclient
CGO_ENABLED=1 run_quiet 'race tests' 600 go test -race ./... -count=1 -timeout 10m
run_quiet 'go vet' 600 go vet ./...
run_quiet 'golangci-lint' 600 go tool golangci-lint run ./...

start_dependencies
run_quiet 'migrations up' 120 go tool goose -dir "${repo_root}/db/migrations" postgres "${TALENRO_DATABASE_URL}" up
run_quiet 'migrations down-to 1' 120 go tool goose -dir "${repo_root}/db/migrations" postgres "${TALENRO_DATABASE_URL}" down-to 1
run_quiet 'migrations restore' 120 go tool goose -dir "${repo_root}/db/migrations" postgres "${TALENRO_DATABASE_URL}" up
run_quiet 'integration tests' 600 go test -tags=integration ./... -count=1 -timeout 10m

stop_dependencies
clear_ambient_environment
start_dependencies
run_quiet 'e2e migrations up' 120 go tool goose -dir "${repo_root}/db/migrations" postgres "${TALENRO_DATABASE_URL}" up

conformance_binary=${run_dir}/trust-conformance
case "${OSTYPE:-}" in
  msys*|cygwin*) conformance_binary=${conformance_binary}.exe ;;
esac
run_quiet 'conformance build' 600 go build -o "${conformance_binary}" ./cmd/trust-conformance
validate_conformance_binary
run_quiet 'ephemeral compose config' 60 "${compose[@]}" config --quiet
run_quiet 'checked-in compose config' 60 docker compose -f deploy/dev/compose.yaml config --quiet
run_e2e_tests

stop_dependencies
clear_ambient_environment
run_quiet 'Bash smoke' 4200 bash "${script_dir}/smoke.sh"
if command -v powershell >/dev/null 2>&1; then
  run_quiet 'PowerShell smoke' 4200 powershell -NoProfile -ExecutionPolicy Bypass -File "${script_dir}/smoke.ps1"
fi
