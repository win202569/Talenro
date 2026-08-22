#!/usr/bin/env bash
set -euo pipefail

script_source=${BASH_SOURCE[0]}
case "${script_source}" in
  [A-Za-z]:\\*) script_source=${script_source//\\//} ;;
esac
case "${script_source}" in
  */*) script_parent=${script_source%/*} ;;
  *) script_parent=. ;;
esac
if ! script_dir="$(cd -- "${script_parent}" 2>/dev/null && pwd -P)" ||
   ! repo_root="$(cd -- "${script_dir}/.." 2>/dev/null && pwd -P)"; then
  printf '%s\n' 'smoke: repository resolution failed with exit code 2.' >&2
  exit 2
fi
declare -a compose=()
compose_project=''
compose_override=''
nats_config=''
compose_touched=0
api_pid=''
mirror_a_pid=''
mirror_b_pid=''
build_dir=''
api_binary=''
mirror_binary=''
conformance_binary=''
control_stdout=''
control_stderr=''
mirror_a_stdout=''
mirror_a_stderr=''
mirror_b_stdout=''
mirror_b_stderr=''
quiet_exit=0
quiet_output=''
run_counter=0
fault_service=''
fault_container_id=''
fault_network_id=''

invoke_quiet() {
  local seconds=$1
  shift
  local stdout_path stderr_path cleanup_exit=0
  ((run_counter += 1))
  printf -v stdout_path '%s/command-%03d.stdout' "${build_dir}" "${run_counter}"
  printf -v stderr_path '%s/command-%03d.stderr' "${build_dir}" "${run_counter}"
  if timeout --kill-after=5s "${seconds}s" "$@" >"${stdout_path}" 2>"${stderr_path}"; then
    quiet_exit=0
  else
    quiet_exit=$?
  fi
  quiet_output=''
  rm -f -- "${stdout_path}" "${stderr_path}" >/dev/null 2>&1 || cleanup_exit=1
  if ((quiet_exit == 0 && cleanup_exit != 0)); then
    quiet_exit=1
  fi
}

run_quiet() {
  local stage=$1
  local seconds=$2
  shift 2
  invoke_quiet "${seconds}" "$@"
  if (( quiet_exit != 0 )); then
    printf '%s failed with exit code %d.\n' "${stage}" "${quiet_exit}" >&2
    return "${quiet_exit}"
  fi
}

run_quiet_in_directory() {
  local stage=$1
  local directory=$2
  local seconds=$3
  shift 3
  invoke_quiet "${seconds}" bash -c 'cd -- "$1" && shift && exec "$@"' bash "${directory}" "$@"
  if (( quiet_exit != 0 )); then
    printf '%s failed with exit code %d.\n' "${stage}" "${quiet_exit}" >&2
    return "${quiet_exit}"
  fi
}

capture_quiet() {
  local seconds=$1
  shift
  local stdout_path stderr_path size cleanup_exit=0
  ((run_counter += 1))
  printf -v stdout_path '%s/command-%03d.stdout' "${build_dir}" "${run_counter}"
  printf -v stderr_path '%s/command-%03d.stderr' "${build_dir}" "${run_counter}"
  if timeout --kill-after=5s "${seconds}s" "$@" >"${stdout_path}" 2>"${stderr_path}"; then
    quiet_exit=0
  else
    quiet_exit=$?
  fi
  if ((quiet_exit == 0)); then
    size=$(wc -c <"${stdout_path}")
    if ((size > 4096)); then
      quiet_exit=1
      quiet_output=''
    else
      quiet_output=$(<"${stdout_path}")
      while [[ "${quiet_output}" == *$'\n' || "${quiet_output}" == *$'\r' ]]; do
        quiet_output=${quiet_output%?}
      done
    fi
  else
    quiet_output=''
  fi
  rm -f -- "${stdout_path}" "${stderr_path}" >/dev/null 2>&1 || cleanup_exit=1
  if ((quiet_exit == 0 && cleanup_exit != 0)); then
    quiet_exit=1
  fi
}

create_build_directory() {
  local requested_root=${TMPDIR:-/tmp}
  local canonical_root candidate

  if ! canonical_root="$(cd -- "${requested_root}" 2>/dev/null && pwd -P)"; then
    printf '%s\n' 'smoke: build directory failed with exit code 1.' >&2
    return 1
  fi
  set +e
  candidate="$(mktemp -d "${canonical_root}/talenro-smoke-c11.XXXXXX" 2>/dev/null)"
  quiet_exit=$?
  set -e
  if (( quiet_exit != 0 )); then
    printf 'smoke: build directory failed with exit code %d.\n' "${quiet_exit}" >&2
    return "${quiet_exit}"
  fi
  if ! build_dir="$(cd -- "${candidate}" 2>/dev/null && pwd -P)" ||
     [[ "${build_dir}" != "${canonical_root}"/talenro-smoke-c11.* ]]; then
    rmdir -- "${candidate}" >/dev/null 2>&1 || true
    build_dir=''
    printf '%s\n' 'smoke: build directory validation failed with exit code 1.' >&2
    return 1
  fi
  api_binary="${build_dir}/control-api"
  mirror_binary="${build_dir}/bundle-mirror"
  conformance_binary="${build_dir}/trust-conformance"
  control_stdout="${build_dir}/control-api.stdout.sink"
  control_stderr="${build_dir}/control-api.stderr.sink"
  mirror_a_stdout="${build_dir}/mirror-a.stdout.sink"
  mirror_a_stderr="${build_dir}/mirror-a.stderr.sink"
  mirror_b_stdout="${build_dir}/mirror-b.stdout.sink"
  mirror_b_stderr="${build_dir}/mirror-b.stderr.sink"
  compose_override="${build_dir}/compose.ephemeral.yaml"
  nats_config="${build_dir}/nats.conf"
  compose_project="talenro-c11-smoke-$(printf '%08x%04x' "$$" "${RANDOM}")"
  if [[ ! "${compose_project}" =~ ^talenro-c11-smoke-[0-9a-f]{12}$ ]]; then
    printf '%s\n' 'smoke: dependency ownership failed with exit code 1.' >&2
    return 1
  fi
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
    '}' >"${nats_config}"
  compose=(docker compose --project-name "${compose_project}" \
    -f "${compose_override}")
}

remove_build_directory() {
  local canonical_root expected_prefix file
  [[ -n "${build_dir}" ]] || return 0
  if ! canonical_root="$(cd -- "${TMPDIR:-/tmp}" 2>/dev/null && pwd -P)"; then
    return 1
  fi
  expected_prefix="${canonical_root}/talenro-smoke-c11."
  if [[ "${build_dir}" != "${expected_prefix}"* || "${api_binary}" != "${build_dir}/control-api" ||
        "${mirror_binary}" != "${build_dir}/bundle-mirror" || "${conformance_binary}" != "${build_dir}/trust-conformance" ||
        "${compose_override}" != "${build_dir}/compose.ephemeral.yaml" || "${nats_config}" != "${build_dir}/nats.conf" ]]; then
    return 1
  fi
  for file in "${api_binary}" "${mirror_binary}" "${conformance_binary}" \
    "${control_stdout}" "${control_stderr}" "${mirror_a_stdout}" "${mirror_a_stderr}" "${mirror_b_stdout}" "${mirror_b_stderr}" \
    "${compose_override}" "${nats_config}"; do
    [[ "${file}" == "${build_dir}/"* ]] || return 1
    if [[ -e "${file}" ]] && ! rm -f -- "${file}" >/dev/null 2>&1; then
      return 1
    fi
  done
  if ! rmdir -- "${build_dir}" >/dev/null 2>&1; then
    return 1
  fi
  build_dir=''
  api_binary=''
  mirror_binary=''
  conformance_binary=''
  compose_override=''
  nats_config=''
}

load_environment() {
  local env_file="${repo_root}/.env.example"
  local line name value existing
  local -a required=(
    TALENRO_HTTP_ADDRESS
    TALENRO_METRICS_ADDRESS
    TALENRO_ALLOW_PUBLIC_METRICS
    TALENRO_DATABASE_URL
    TALENRO_REDIS_ADDRESS
    TALENRO_NATS_URL
    TALENRO_ALLOW_PUBLIC_HTTP
    TALENRO_PROFILE
    TALENRO_PUBLIC_BASE_URL
    TALENRO_PRIMARY_BUNDLE_BASE_URL
    TALENRO_MIRROR_A_BASE_URL
    TALENRO_MIRROR_B_BASE_URL
    TALENRO_WEBAUTHN_RP_ID
    TALENRO_WEBAUTHN_ORIGINS
    TALENRO_EMAIL_VERIFICATION_MODE
    TALENRO_SIGNER_PROVIDER
    TALENRO_FIELD_PROTECTOR_PROVIDER
    TALENRO_EMAIL_PROVIDER
    TALENRO_ERROR_REPORTER_PROVIDER
    TALENRO_REQUEST_DEADLINE
    TALENRO_REDIS_TIMEOUT
    TALENRO_SIGNER_TIMEOUT
    TALENRO_ERROR_REPORT_TIMEOUT
    TALENRO_REDIS_DOWN_AFTER_FAILURES
    TALENRO_REDIS_RECOVER_AFTER_SUCCESSES
    TALENRO_OUTBOX_DEGRADED_BACKLOG
    TALENRO_OUTBOX_DOWN_BACKLOG
    TALENRO_OUTBOX_DEGRADED_AGE
    TALENRO_OUTBOX_DOWN_AGE
    TALENRO_ERROR_REPORT_QUEUE
    TALENRO_ERROR_REPORT_BATCH
    TALENRO_CLOCK_SKEW
    TALENRO_LOGIN_RATE_LIMIT
    TALENRO_LOGIN_RATE_WINDOW
    TALENRO_DELIVERY_RATE_LIMIT
    TALENRO_DELIVERY_RATE_WINDOW
    TALENRO_CHALLENGE_RATE_LIMIT
    TALENRO_CHALLENGE_RATE_WINDOW
    TALENRO_SENSITIVE_LOOKUP_KEY_B64
    TALENRO_SENSITIVE_ENCRYPTION_KEY_B64
    TALENRO_LOCAL_ROOT_SIGNING_SEED_B64
    TALENRO_LOCAL_CONFIG_SIGNING_SEED_B64
  )
  local -A allowed=()
  local -A parsed=()

  for name in "${required[@]}"; do
    allowed["${name}"]=1
  done

  if [[ ! -f "${env_file}" ]]; then
    printf '%s\n' 'smoke: environment failed with exit code 2.' >&2
    return 2
  fi

  while IFS= read -r line || [[ -n "${line}" ]]; do
    line=${line%$'\r'}
    if [[ -z "${line}" || "${line}" == \#* ]]; then
      continue
    fi
    if [[ ! "${line}" =~ ^([A-Z][A-Z0-9_]*)=([A-Za-z0-9._:/?@=+-]+)$ ]]; then
      printf '%s\n' 'smoke: environment failed with exit code 2.' >&2
      return 2
    fi
    name=${BASH_REMATCH[1]}
    value=${BASH_REMATCH[2]}
    if [[ -z "${allowed[${name}]+present}" || -n "${parsed[${name}]+present}" ]]; then
      printf '%s\n' 'smoke: environment failed with exit code 2.' >&2
      return 2
    fi
    parsed["${name}"]=${value}
  done < "${env_file}"

  for name in "${required[@]}"; do
    if [[ -z "${parsed[${name}]+present}" ]]; then
      printf '%s\n' 'smoke: environment failed with exit code 2.' >&2
      return 2
    fi
  done

  while IFS= read -r existing; do
    case "${existing}" in
      TALENRO_*|C11_*|COMPOSE_*) unset "${existing}" ;;
    esac
  done < <(compgen -v || true)
  for name in "${required[@]}"; do
    printf -v "${name}" '%s' "${parsed[${name}]}"
    export "${name}"
  done
}

assert_compose_ownership() {
  local service=$1
  local container_id owner owned_service status
  case "${service}" in
    postgres|redis|nats) ;;
    *)
      printf '%s\n' 'smoke: dependency ownership failed with exit code 1.' >&2
      return 1
      ;;
  esac
  if capture_quiet 30 "${compose[@]}" ps --quiet --no-trunc "${service}"; then
    container_id=${quiet_output}
  else
    status=${quiet_exit}
    printf 'smoke: dependency ownership failed with exit code %d.\n' "${status}" >&2
    return "${status}"
  fi
  if [[ ! "${container_id}" =~ ^[0-9a-f]{64}$ ]]; then
    printf '%s\n' 'smoke: dependency ownership failed with exit code 1.' >&2
    return 1
  fi
  if capture_quiet 30 docker inspect --type container --format '{{ index .Config.Labels "com.docker.compose.project" }}' "${container_id}"; then
    owner=${quiet_output}
  else
    status=${quiet_exit}
    printf 'smoke: dependency ownership failed with exit code %d.\n' "${status}" >&2
    return "${status}"
  fi
  if [[ "${owner}" != "${compose_project}" ]]; then
    printf '%s\n' 'smoke: dependency ownership failed with exit code 1.' >&2
    return 1
  fi
  if capture_quiet 30 docker inspect --type container --format '{{ index .Config.Labels "com.docker.compose.service" }}' "${container_id}"; then
    owned_service=${quiet_output}
  else
    status=${quiet_exit}
    printf 'smoke: dependency ownership failed with exit code %d.\n' "${status}" >&2
    return "${status}"
  fi
  if [[ "${owned_service}" != "${service}" ]]; then
    printf '%s\n' 'smoke: dependency ownership failed with exit code 1.' >&2
    return 1
  fi
}

resolve_network_fault_target() {
  local service=$1
  local stage container_id project_label service_label network_id network_project_label network_name_label status
  case "${service}" in
    postgres) stage='smoke: postgres network ownership' ;;
    redis) stage='smoke: redis network ownership' ;;
    nats) stage='smoke: nats network ownership' ;;
    *)
      printf '%s\n' 'smoke: network ownership failed with exit code 1.' >&2
      return 1
      ;;
  esac
  if [[ ! "${compose_project}" =~ ^talenro-c11-smoke-[0-9a-f]{12}$ ]]; then
    printf '%s failed with exit code 1.\n' "${stage}" >&2
    return 1
  fi
  if capture_quiet 30 "${compose[@]}" ps --quiet --no-trunc "${service}"; then
    container_id=${quiet_output}
  else
    status=${quiet_exit}
    printf '%s failed with exit code %d.\n' "${stage}" "${status}" >&2
    return "${status}"
  fi
  if [[ ! "${container_id}" =~ ^[0-9a-f]{64}$ ]]; then
    printf '%s failed with exit code 1.\n' "${stage}" >&2
    return 1
  fi
  if capture_quiet 30 docker inspect --type container --format '{{ index .Config.Labels "com.docker.compose.project" }}' "${container_id}"; then
    project_label=${quiet_output}
  else
    status=${quiet_exit}
    printf '%s failed with exit code %d.\n' "${stage}" "${status}" >&2
    return "${status}"
  fi
  if capture_quiet 30 docker inspect --type container --format '{{ index .Config.Labels "com.docker.compose.service" }}' "${container_id}"; then
    service_label=${quiet_output}
  else
    status=${quiet_exit}
    printf '%s failed with exit code %d.\n' "${stage}" "${status}" >&2
    return "${status}"
  fi
  if [[ "${project_label}" != "${compose_project}" || "${service_label}" != "${service}" ]]; then
    printf '%s failed with exit code 1.\n' "${stage}" >&2
    return 1
  fi
  if capture_quiet 30 docker network ls --quiet --no-trunc \
      --filter "label=com.docker.compose.project=${compose_project}"; then
    network_id=${quiet_output}
  else
    status=${quiet_exit}
    printf '%s failed with exit code %d.\n' "${stage}" "${status}" >&2
    return "${status}"
  fi
  if [[ ! "${network_id}" =~ ^[0-9a-f]{64}$ ]]; then
    printf '%s failed with exit code 1.\n' "${stage}" >&2
    return 1
  fi
  if capture_quiet 30 docker network inspect --format '{{ index .Labels "com.docker.compose.project" }}' "${network_id}"; then
    network_project_label=${quiet_output}
  else
    status=${quiet_exit}
    printf '%s failed with exit code %d.\n' "${stage}" "${status}" >&2
    return "${status}"
  fi
  if capture_quiet 30 docker network inspect --format '{{ index .Labels "com.docker.compose.network" }}' "${network_id}"; then
    network_name_label=${quiet_output}
  else
    status=${quiet_exit}
    printf '%s failed with exit code %d.\n' "${stage}" "${status}" >&2
    return "${status}"
  fi
  if [[ "${network_project_label}" != "${compose_project}" || "${network_name_label}" != default ]]; then
    printf '%s failed with exit code 1.\n' "${stage}" >&2
    return 1
  fi
  fault_service=${service}
  fault_container_id=${container_id}
  fault_network_id=${network_id}
}

disconnect_compose_network() {
  local service=$1
  local stage
  resolve_network_fault_target "${service}"
  case "${service}" in
    postgres) stage='smoke: postgres network disconnect' ;;
    redis) stage='smoke: redis network disconnect' ;;
    nats) stage='smoke: nats network disconnect' ;;
    *)
      printf '%s\n' 'smoke: network disconnect failed with exit code 1.' >&2
      return 1
      ;;
  esac
  run_quiet "${stage}" 60 docker network disconnect "${fault_network_id}" "${fault_container_id}"
}

connect_compose_network() {
  local service=$1
  local stage
  case "${service}" in
    postgres) stage='smoke: postgres network connect' ;;
    redis) stage='smoke: redis network connect' ;;
    nats) stage='smoke: nats network connect' ;;
    *)
      printf '%s\n' 'smoke: network connect failed with exit code 1.' >&2
      return 1
      ;;
  esac
  if [[ "${fault_service}" != "${service}" || ! "${fault_container_id}" =~ ^[0-9a-f]{64}$ || ! "${fault_network_id}" =~ ^[0-9a-f]{64}$ ]]; then
    printf '%s failed with exit code 1.\n' "${stage}" >&2
    return 1
  fi
  run_quiet "${stage}" 60 docker network connect "${fault_network_id}" "${fault_container_id}"
  fault_service=''
  fault_container_id=''
  fault_network_id=''
}

compose_port() {
  local service=$1
  local container_port=$2
  local endpoint status port
  if capture_quiet 30 "${compose[@]}" port "${service}" "${container_port}"; then
    endpoint=${quiet_output}
  else
    status=${quiet_exit}
    printf 'smoke: dependency endpoints failed with exit code %d.\n' "${status}" >&2
    return "${status}"
  fi
  if [[ ! "${endpoint}" =~ ^127\.0\.0\.1:([1-9][0-9]{0,4})$ ]]; then
    printf '%s\n' 'smoke: dependency endpoints failed with exit code 1.' >&2
    return 1
  fi
  port=${BASH_REMATCH[1]}
  if ((10#${port} > 65535)); then
    printf '%s\n' 'smoke: dependency endpoints failed with exit code 1.' >&2
    return 1
  fi
  quiet_output=${port}
}

now_milliseconds() {
  local whole fraction
  if [[ -z "${EPOCHREALTIME:-}" ]]; then
    return 70
  fi
  whole=${EPOCHREALTIME%%.*}
  fraction=${EPOCHREALTIME#*.}000
  now_ms=$((10#${whole} * 1000 + 10#${fraction:0:3}))
}

get_status() {
  local url=$1
  local timeout_ms=$2
  local timeout
  printf -v timeout '%d.%03d' "$((timeout_ms / 1000))" "$((timeout_ms % 1000))"

  set +e
  http_status="$(curl --silent --output /dev/null --write-out '%{http_code}' \
    --no-keepalive --connect-timeout "${timeout}" --max-time "${timeout}" -- "${url}" 2>/dev/null)"
  curl_exit=$?
  set -e
}

wait_status() {
  local url=$1
  local expected=$2
  local seconds=$3
  local stage=$4
  local process_pid=$5
  local process_stage=$6
  local deadline remaining sleep_ms sleep_value process_exit

  if ! now_milliseconds; then
    printf '%s\n' 'smoke: clock failed with exit code 70.' >&2
    return 70
  fi
  deadline=$((now_ms + seconds * 1000))

  while true; do
    if [[ -n "${process_pid}" ]] && ! kill -0 "${process_pid}" 2>/dev/null; then
      set +e
      wait "${process_pid}" >/dev/null 2>&1
      process_exit=$?
      set -e
      if (( process_exit == 0 )); then
        process_exit=1
      fi
      printf '%s failed with exit code %d.\n' "${process_stage}" "${process_exit}" >&2
      return "${process_exit}"
    fi

    now_milliseconds
    remaining=$((deadline - now_ms))
    if (( remaining <= 0 )); then
      printf '%s failed with exit code 1.\n' "${stage}" >&2
      return 1
    fi
    if (( remaining > 3000 )); then
      request_timeout=3000
    else
      request_timeout=${remaining}
    fi
    get_status "${url}" "${request_timeout}"
    if (( curl_exit == 127 )); then
      printf '%s\n' 'smoke: HTTP client failed with exit code 127.' >&2
      return 127
    fi
    if (( curl_exit == 0 )) && [[ "${http_status}" == "${expected}" ]]; then
      return 0
    fi

    now_milliseconds
    remaining=$((deadline - now_ms))
    if (( remaining <= 0 )); then
      continue
    fi
    if (( remaining > 250 )); then
      sleep_ms=250
    else
      sleep_ms=${remaining}
    fi
    printf -v sleep_value '0.%03d' "${sleep_ms}"
    sleep "${sleep_value}"
  done
}

terminate_pid() {
  local child_pid=$1
  local attempt deadline remaining sleep_ms sleep_value
  if [[ -z "${child_pid}" ]] || ! kill -0 "${child_pid}" 2>/dev/null; then
    if [[ -n "${child_pid}" ]]; then
      wait "${child_pid}" >/dev/null 2>&1 || true
    fi
    return 0
  fi

  kill -TERM "${child_pid}" 2>/dev/null || true
  if now_milliseconds; then
    deadline=$((now_ms + 4750))
    while kill -0 "${child_pid}" 2>/dev/null; do
      if ! now_milliseconds; then
        break
      fi
      remaining=$((deadline - now_ms))
      if (( remaining <= 0 )); then
        break
      fi
      if (( remaining > 250 )); then
        sleep_ms=250
      else
        sleep_ms=${remaining}
      fi
      printf -v sleep_value '%d.%03d' "$((sleep_ms / 1000))" "$((sleep_ms % 1000))"
      sleep "${sleep_value}"
    done
    if ! kill -0 "${child_pid}" 2>/dev/null; then
      wait "${child_pid}" >/dev/null 2>&1 || true
      return 0
    fi
  fi

  kill -KILL "${child_pid}" 2>/dev/null || true
  for ((attempt = 0; attempt < 4; attempt++)); do
    if ! kill -0 "${child_pid}" 2>/dev/null; then
      wait "${child_pid}" >/dev/null 2>&1 || true
      return 0
    fi
    sleep 0.25
  done
  return 1
}

scan_artifacts() {
  local path line forbidden
  for path in "${control_stdout}" "${control_stderr}" "${mirror_a_stdout}" "${mirror_a_stderr}" "${mirror_b_stdout}" "${mirror_b_stderr}"; do
    [[ -f "${path}" ]] || return 1
    while IFS= read -r line || [[ -n "${line}" ]]; do
      for forbidden in "${TALENRO_DATABASE_URL}" "${TALENRO_SENSITIVE_LOOKUP_KEY_B64}" \
        "${TALENRO_SENSITIVE_ENCRYPTION_KEY_B64}" "${TALENRO_LOCAL_ROOT_SIGNING_SEED_B64}" \
        "${TALENRO_LOCAL_CONFIG_SIGNING_SEED_B64}"; do
        [[ -z "${forbidden}" || "${line}" != *"${forbidden}"* ]] || return 1
      done
    done < "${path}"
  done
}

cleanup() {
  local original_exit=$?
  local cleanup_exit=0
  local cleanup_stage='smoke: cleanup'
  trap - EXIT INT TERM
  set +e

  terminate_pid "${mirror_b_pid}"
  if (( $? != 0 && cleanup_exit == 0 )); then
    cleanup_exit=1
    cleanup_stage='smoke: mirror B cleanup'
  fi
  terminate_pid "${mirror_a_pid}"
  if (( $? != 0 && cleanup_exit == 0 )); then
    cleanup_exit=1
    cleanup_stage='smoke: mirror A cleanup'
  fi
  terminate_pid "${api_pid}"
  if (( $? != 0 && cleanup_exit == 0 )); then
    cleanup_exit=1
    cleanup_stage='smoke: control API cleanup'
  fi
  if [[ -n "${build_dir}" ]] && ! scan_artifacts && (( cleanup_exit == 0 )); then
    cleanup_exit=1
    cleanup_stage='smoke: artifact privacy'
  fi
  if (( compose_touched )); then
    invoke_quiet 60 "${compose[@]}" down --remove-orphans --timeout 20
    if (( quiet_exit != 0 )); then
      cleanup_exit=${quiet_exit}
      cleanup_stage='smoke: compose down'
    fi
  fi
  if ! remove_build_directory; then
    if (( cleanup_exit == 0 )); then
      cleanup_exit=1
      cleanup_stage='smoke: build cleanup'
    fi
  fi

  quiet_output=''
  if (( original_exit != 0 )); then
    exit "${original_exit}"
  fi
  if (( cleanup_exit != 0 )); then
    printf '%s failed with exit code %d.\n' "${cleanup_stage}" "${cleanup_exit}" >&2
    exit "${cleanup_exit}"
  fi
  printf '%s\n' 'smoke: C1.1 acceptance passed.'
  exit 0
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

command -v timeout >/dev/null 2>&1 || {
  printf '%s\n' 'smoke: hard deadline tool failed with exit code 127.' >&2
  exit 127
}
load_environment
if [[ ! "${TALENRO_HTTP_ADDRESS}" =~ ^127\.0\.0\.1:[1-9][0-9]{0,4}$ ||
      ! "${TALENRO_METRICS_ADDRESS}" =~ ^127\.0\.0\.1:[1-9][0-9]{0,4}$ ||
      ! "${TALENRO_REDIS_ADDRESS}" =~ ^127\.0\.0\.1:[1-9][0-9]{0,4}$ ||
      ! "${TALENRO_NATS_URL}" =~ ^nats://127\.0\.0\.1:[1-9][0-9]{0,4}$ ||
      ! "${TALENRO_DATABASE_URL}" =~ ^postgres://[^@/]+@127\.0\.0\.1:[1-9][0-9]{0,4}/[^?]+\?sslmode=disable$ ||
      "${TALENRO_ALLOW_PUBLIC_METRICS}" != false ||
      "${TALENRO_ALLOW_PUBLIC_HTTP}" != false ||
      "${TALENRO_PROFILE}" != local ||
      "${TALENRO_PUBLIC_BASE_URL}" != http://localhost:8080 ||
      "${TALENRO_PRIMARY_BUNDLE_BASE_URL}" != http://localhost:8080 ||
      "${TALENRO_MIRROR_A_BASE_URL}" != http://localhost:8081 ||
      "${TALENRO_MIRROR_B_BASE_URL}" != http://localhost:8082 ||
      "${TALENRO_WEBAUTHN_RP_ID}" != localhost ||
      "${TALENRO_WEBAUTHN_ORIGINS}" != http://localhost:8080 ||
      "${TALENRO_EMAIL_VERIFICATION_MODE}" != disabled ||
      "${TALENRO_SIGNER_PROVIDER}" != local ||
      "${TALENRO_FIELD_PROTECTOR_PROVIDER}" != local ||
      "${TALENRO_EMAIL_PROVIDER}" != local ||
      "${TALENRO_ERROR_REPORTER_PROVIDER}" != discard ||
      "${TALENRO_REQUEST_DEADLINE}" != 5s ||
      "${TALENRO_REDIS_TIMEOUT}" != 250ms ||
      "${TALENRO_SIGNER_TIMEOUT}" != 2s ||
      "${TALENRO_ERROR_REPORT_TIMEOUT}" != 1s ||
      "${TALENRO_REDIS_DOWN_AFTER_FAILURES}" != 3 ||
      "${TALENRO_REDIS_RECOVER_AFTER_SUCCESSES}" != 2 ||
      "${TALENRO_OUTBOX_DEGRADED_BACKLOG}" != 1000 ||
      "${TALENRO_OUTBOX_DOWN_BACKLOG}" != 10000 ||
      "${TALENRO_OUTBOX_DEGRADED_AGE}" != 60s ||
      "${TALENRO_OUTBOX_DOWN_AGE}" != 300s ||
      "${TALENRO_ERROR_REPORT_QUEUE}" != 100 ||
      "${TALENRO_ERROR_REPORT_BATCH}" != 20 ||
      "${TALENRO_CLOCK_SKEW}" != 120s ||
      "${TALENRO_LOGIN_RATE_LIMIT}" != 10 ||
      "${TALENRO_LOGIN_RATE_WINDOW}" != 15m ||
      "${TALENRO_DELIVERY_RATE_LIMIT}" != 5 ||
      "${TALENRO_DELIVERY_RATE_WINDOW}" != 1h ||
      "${TALENRO_CHALLENGE_RATE_LIMIT}" != 20 ||
      "${TALENRO_CHALLENGE_RATE_WINDOW}" != 5m ||
      ! "${TALENRO_SENSITIVE_LOOKUP_KEY_B64}" =~ ^[A-Za-z0-9_-]{43}$ ||
      ! "${TALENRO_SENSITIVE_ENCRYPTION_KEY_B64}" =~ ^[A-Za-z0-9_-]{43}$ ||
      ! "${TALENRO_LOCAL_ROOT_SIGNING_SEED_B64}" =~ ^[A-Za-z0-9_-]{43}$ ||
      ! "${TALENRO_LOCAL_CONFIG_SIGNING_SEED_B64}" =~ ^[A-Za-z0-9_-]{43}$ ]]; then
  printf '%s\n' 'smoke: local environment failed with exit code 2.' >&2
  exit 2
fi
http_base="http://${TALENRO_HTTP_ADDRESS}"

create_build_directory

compose_touched=1
run_quiet 'smoke: compose up' 180 "${compose[@]}" up -d --wait --wait-timeout 120 postgres redis nats
for service in postgres redis nats; do
  assert_compose_ownership "${service}"
done
compose_port postgres 5432
postgres_port=${quiet_output}
compose_port redis 6379
redis_port=${quiet_output}
compose_port nats 4222
nats_port=${quiet_output}
export TALENRO_DATABASE_URL="postgres://talenro:talenro_dev@127.0.0.1:${postgres_port}/talenro?sslmode=disable"
export TALENRO_REDIS_ADDRESS="127.0.0.1:${redis_port}"
export TALENRO_NATS_URL="nats://127.0.0.1:${nats_port}"
export C11_E2E_COMPOSE_PROJECT="${compose_project}"
export C11_E2E_DATABASE_URL="${TALENRO_DATABASE_URL}"
export C11_E2E_REDIS_ADDRESS="${TALENRO_REDIS_ADDRESS}"
export C11_E2E_NATS_URL="${TALENRO_NATS_URL}"
run_quiet 'smoke: migrations' 120 go tool goose -dir "${repo_root}/db/migrations" postgres "${TALENRO_DATABASE_URL}" up
run_quiet_in_directory 'smoke: control API build' "${repo_root}" 600 go build -o "${api_binary}" "${repo_root}/cmd/control-api"
run_quiet_in_directory 'smoke: mirror build' "${repo_root}" 600 go build -o "${mirror_binary}" "${repo_root}/cmd/bundle-mirror"
run_quiet_in_directory 'smoke: conformance build' "${repo_root}" 600 go build -o "${conformance_binary}" "${repo_root}/cmd/trust-conformance"

if ! pushd "${repo_root}" >/dev/null; then
  printf '%s\n' 'smoke: process start failed with exit code 1.' >&2
  exit 1
fi
"${api_binary}" >"${control_stdout}" 2>"${control_stderr}" &
api_pid=$!
env -i TALENRO_DATABASE_URL="${TALENRO_DATABASE_URL}" TALENRO_HTTP_ADDRESS=127.0.0.1:8081 \
  "${mirror_binary}" >"${mirror_a_stdout}" 2>"${mirror_a_stderr}" &
mirror_a_pid=$!
env -i TALENRO_DATABASE_URL="${TALENRO_DATABASE_URL}" TALENRO_HTTP_ADDRESS=127.0.0.1:8082 \
  "${mirror_binary}" >"${mirror_b_stdout}" 2>"${mirror_b_stderr}" &
mirror_b_pid=$!
popd >/dev/null
wait_status "${http_base}/livez" 200 10 'smoke: initial liveness' "${api_pid}" 'smoke: control API'
wait_status "${http_base}/readyz" 200 10 'smoke: initial readiness' "${api_pid}" 'smoke: control API'
wait_status 'http://127.0.0.1:8081/' 404 10 'smoke: mirror A readiness' "${mirror_a_pid}" 'smoke: mirror A'
wait_status 'http://127.0.0.1:8082/' 404 10 'smoke: mirror B readiness' "${mirror_b_pid}" 'smoke: mirror B'

run_quiet_in_directory 'smoke: C1.1 happy path' "${repo_root}" 600 env \
  C11_E2E_EXTERNAL_RUNTIME=1 \
  C11_E2E_PRIMARY_URL=http://127.0.0.1:8080 C11_E2E_MIRROR_A_URL=http://127.0.0.1:8081 \
  C11_E2E_MIRROR_B_URL=http://127.0.0.1:8082 C11_E2E_PRIMARY_ORIGIN=http://localhost:8080 \
  C11_E2E_MIRROR_A_ORIGIN=http://localhost:8081 C11_E2E_MIRROR_B_ORIGIN=http://localhost:8082 \
  C11_CONFORMANCE_BINARY="${conformance_binary}" \
  go test -tags=e2e ./internal/e2e -run '^TestC11HappyPath$' -count=1 -timeout 10m

disconnect_compose_network postgres
wait_status "${http_base}/readyz" 503 5 'smoke: postgres readiness failure' "${api_pid}" 'smoke: control API'
wait_status "${http_base}/livez" 200 2 'smoke: postgres liveness' "${api_pid}" 'smoke: control API'

connect_compose_network postgres
wait_status "${http_base}/readyz" 200 10 'smoke: postgres recovery' "${api_pid}" 'smoke: control API'

disconnect_compose_network redis
wait_status "${http_base}/readyz" 503 5 'smoke: redis readiness failure' "${api_pid}" 'smoke: control API'
wait_status "${http_base}/livez" 200 2 'smoke: redis liveness' "${api_pid}" 'smoke: control API'
connect_compose_network redis
wait_status "${http_base}/readyz" 200 10 'smoke: redis recovery' "${api_pid}" 'smoke: control API'

disconnect_compose_network nats
wait_status "${http_base}/livez" 200 2 'smoke: nats liveness' "${api_pid}" 'smoke: control API'
connect_compose_network nats
wait_status "${http_base}/readyz" 200 10 'smoke: nats recovery' "${api_pid}" 'smoke: control API'
