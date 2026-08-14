#!/usr/bin/env bash
set -euo pipefail

script_source=${BASH_SOURCE[0]}
case "${script_source}" in
  */*) script_parent=${script_source%/*} ;;
  *) script_parent=. ;;
esac
if ! script_dir="$(cd -- "${script_parent}" 2>/dev/null && pwd -P)" ||
   ! repo_root="$(cd -- "${script_dir}/.." 2>/dev/null && pwd -P)"; then
  printf '%s\n' 'smoke: repository resolution failed with exit code 2.' >&2
  exit 2
fi
compose=(docker compose -f "${repo_root}/deploy/dev/compose.yaml")
compose_touched=0
api_pid=''
build_dir=''
api_binary=''
quiet_exit=0
quiet_output=''

invoke_quiet() {
  set +e
  quiet_output="$("$@" 2>&1)"
  quiet_exit=$?
  set -e
}

run_quiet() {
  local stage=$1
  shift
  invoke_quiet "$@"
  if (( quiet_exit != 0 )); then
    printf '%s failed with exit code %d.\n' "${stage}" "${quiet_exit}" >&2
    return "${quiet_exit}"
  fi
}

run_quiet_in_directory() {
  local stage=$1
  local directory=$2
  shift 2
  set +e
  quiet_output="$(cd -- "${directory}" 2>/dev/null && "$@" 2>&1)"
  quiet_exit=$?
  set -e
  if (( quiet_exit != 0 )); then
    printf '%s failed with exit code %d.\n' "${stage}" "${quiet_exit}" >&2
    return "${quiet_exit}"
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
  candidate="$(mktemp -d "${canonical_root}/talenro-smoke-control-api.XXXXXX" 2>/dev/null)"
  quiet_exit=$?
  set -e
  if (( quiet_exit != 0 )); then
    printf 'smoke: build directory failed with exit code %d.\n' "${quiet_exit}" >&2
    return "${quiet_exit}"
  fi
  if ! build_dir="$(cd -- "${candidate}" 2>/dev/null && pwd -P)" ||
     [[ "${build_dir}" != "${canonical_root}"/talenro-smoke-control-api.* ]]; then
    rmdir -- "${candidate}" >/dev/null 2>&1 || true
    build_dir=''
    printf '%s\n' 'smoke: build directory validation failed with exit code 1.' >&2
    return 1
  fi
  api_binary="${build_dir}/control-api"
}

remove_build_directory() {
  local canonical_root expected_prefix
  [[ -n "${build_dir}" ]] || return 0
  if ! canonical_root="$(cd -- "${TMPDIR:-/tmp}" 2>/dev/null && pwd -P)"; then
    return 1
  fi
  expected_prefix="${canonical_root}/talenro-smoke-control-api."
  if [[ "${build_dir}" != "${expected_prefix}"* || "${api_binary}" != "${build_dir}/control-api" ]]; then
    return 1
  fi
  if [[ -e "${api_binary}" ]] && ! rm -f -- "${api_binary}" >/dev/null 2>&1; then
    return 1
  fi
  if ! rmdir -- "${build_dir}" >/dev/null 2>&1; then
    return 1
  fi
  build_dir=''
  api_binary=''
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
    unset "${existing}"
  done < <(compgen -v TALENRO_ || true)
  for name in "${required[@]}"; do
    printf -v "${name}" '%s' "${parsed[${name}]}"
    export "${name}"
  done
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
  local deadline remaining sleep_ms sleep_value process_exit

  if ! now_milliseconds; then
    printf '%s\n' 'smoke: clock failed with exit code 70.' >&2
    return 70
  fi
  deadline=$((now_ms + seconds * 1000))

  while true; do
    if [[ -n "${api_pid}" ]] && ! kill -0 "${api_pid}" 2>/dev/null; then
      set +e
      wait "${api_pid}" >/dev/null 2>&1
      process_exit=$?
      set -e
      api_pid=''
      if (( process_exit == 0 )); then
        process_exit=1
      fi
      printf 'smoke: control API failed with exit code %d.\n' "${process_exit}" >&2
      return "${process_exit}"
    fi

    now_milliseconds
    remaining=$((deadline - now_ms))
    if (( remaining <= 0 )); then
      printf '%s failed with exit code 1.\n' "${stage}" >&2
      return 1
    fi
    if (( remaining > 1000 )); then
      request_timeout=1000
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

terminate_child() {
  local attempt deadline remaining sleep_ms sleep_value
  if [[ -z "${api_pid}" ]] || ! kill -0 "${api_pid}" 2>/dev/null; then
    if [[ -n "${api_pid}" ]]; then
      wait "${api_pid}" >/dev/null 2>&1 || true
    fi
    api_pid=''
    return 0
  fi

  kill -TERM "${api_pid}" 2>/dev/null || true
  if now_milliseconds; then
    deadline=$((now_ms + 4750))
    while kill -0 "${api_pid}" 2>/dev/null; do
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
    if ! kill -0 "${api_pid}" 2>/dev/null; then
      wait "${api_pid}" >/dev/null 2>&1 || true
      api_pid=''
      return 0
    fi
  fi

  kill -KILL "${api_pid}" 2>/dev/null || true
  for ((attempt = 0; attempt < 4; attempt++)); do
    if ! kill -0 "${api_pid}" 2>/dev/null; then
      wait "${api_pid}" >/dev/null 2>&1 || true
      api_pid=''
      return 0
    fi
    sleep 0.25
  done
  return 1
}

cleanup() {
  local original_exit=$?
  local cleanup_exit=0
  local cleanup_stage='smoke: cleanup'
  trap - EXIT INT TERM
  set +e

  terminate_child
  if (( $? != 0 )); then
    cleanup_exit=1
    cleanup_stage='smoke: control API cleanup'
  fi
  if (( compose_touched )); then
    invoke_quiet "${compose[@]}" down
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
  printf '%s\n' 'smoke: foundation acceptance passed.'
  exit 0
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

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
run_quiet 'smoke: compose up' "${compose[@]}" up -d --wait
run_quiet 'smoke: migrations' go tool goose -dir "${repo_root}/db/migrations" postgres "${TALENRO_DATABASE_URL}" up
run_quiet_in_directory 'smoke: control API build' "${repo_root}" go build -o "${api_binary}" "${repo_root}/cmd/control-api"

if ! pushd "${repo_root}" >/dev/null; then
  printf '%s\n' 'smoke: control API start failed with exit code 1.' >&2
  exit 1
fi
"${api_binary}" >/dev/null 2>&1 &
api_pid=$!
popd >/dev/null
wait_status "${http_base}/livez" 200 10 'smoke: initial liveness'
wait_status "${http_base}/readyz" 200 10 'smoke: initial readiness'

run_quiet 'smoke: postgres stop' "${compose[@]}" stop postgres
wait_status "${http_base}/readyz" 503 5 'smoke: readiness failure'
wait_status "${http_base}/livez" 200 2 'smoke: failure liveness'

run_quiet 'smoke: postgres start' "${compose[@]}" start postgres
wait_status "${http_base}/readyz" 200 10 'smoke: readiness recovery'
