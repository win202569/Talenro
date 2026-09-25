#!/usr/bin/env bash
case "${OSTYPE-}" in
  msys*|cygwin*|win32*) printf '%s\n' 'verify-devtools: Windows requires PowerShell 7.6.5.' >&2; exit 1 ;;
esac
devtools_platform=$(uname -s 2>/dev/null) || {
  printf '%s\n' 'verify-devtools: platform failed with exit code 1.' >&2
  exit 1
}
case "${devtools_platform}" in
  MINGW*|MSYS*|CYGWIN*) printf '%s\n' 'verify-devtools: Windows requires PowerShell 7.6.5.' >&2; exit 1 ;;
esac
set -euo pipefail

stage=module
deadline=$((SECONDS + 900))
run_dir=''
finish() {
  local status=$1
  if [[ -n "${run_dir}" ]]; then
    if ! rm -f -- "${run_dir}/stdout" "${run_dir}/stderr" 2>/dev/null || ! rmdir -- "${run_dir}" 2>/dev/null; then
      stage=cleanup; status=1
    fi
  fi
  if ((status == 0)); then printf '%s\n' 'verify-devtools: passed'; else
    printf 'verify-devtools: %s failed with exit code %s.\n' "${stage}" "${status}" >&2
  fi
  exit "${status}"
}
capture() {
  local remaining=$((deadline - SECONDS)) status
  ((remaining > 0)) || finish 124
  if timeout --kill-after=5s "${remaining}s" "$@" >"${run_dir}/stdout" 2>"${run_dir}/stderr"; then status=0; else status=$?; fi
  ((status == 0)) || finish "${status}"
  [[ $(wc -c <"${run_dir}/stdout") -le 4194304 && $(wc -c <"${run_dir}/stderr") -le 4194304 ]] || finish 1
  captured=$(<"${run_dir}/stdout")
  captured=${captured%$'\r'}
}
cache_key() {
  local value=$1 character i
  escaped=''
  for ((i=0; i<${#value}; i++)); do
    character=${value:i:1}
    case "${character}" in [A-Z]) escaped+="!${character,,}" ;; *) escaped+="${character}" ;; esac
  done
}
script_source=${BASH_SOURCE[0]}
case "${script_source}" in
  */*) script_parent=${script_source%/*} ;;
  *) script_parent=. ;;
esac
if ! repo_root="$(cd -- "${script_parent}/.." 2>/dev/null && pwd -P)" ||
   [[ ! -f "${repo_root}/tools/devtools/go.mod" || ! -f "${repo_root}/tools/devtools/go.sum" ]]; then
  printf 'verify-devtools: %s failed with exit code 1.\n' "${stage}" >&2
  exit 1
fi
stage=environment
[[ $# == 0 && ! ${GOFLAGS:-} =~ (^|[[:space:]])-(modfile|overlay)(=|[[:space:]]|$) ]] || finish 1
windows=0
case "${OSTYPE}" in msys*|cygwin*) windows=1; export PATH="/usr/bin:/mingw64/bin:/bin:${PATH}" ;; esac
if ((windows)); then
  profile=$(cygpath -u "${USERPROFILE}") || finish 1
  module_cache="${profile}/go/pkg/mod"
  go_path="${module_cache}/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/go.exe"
else
  module_cache="${HOME}/go/pkg/mod"
  go_path=$(command -v go) || { stage=toolchain; finish 127; }
fi
for variable in ${!GO@}; do unset "${variable}"; done
export GOWORK=off GOENV=off GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off GOAUTH=off GOVCS=all:off GOFLAGS=-mod=readonly
if ((windows)); then
  export GOMODCACHE="$(cygpath -w "${module_cache}")" GOPATH="$(cygpath -w "${profile}/go")"
else export GOMODCACHE="${module_cache}" GOPATH="${HOME}/go"; fi
stage=toolchain
[[ -f "${go_path}" && -x "${go_path}" ]] || finish 127
for utility in timeout mktemp wc rm rmdir; do command -v "${utility}" >/dev/null || finish 127; done
run_dir=$(mktemp -d "${TMPDIR:-/tmp}/talenro-devtools.XXXXXXXX" 2>/dev/null) || finish 1
trap 'finish 124' TERM INT HUP
cd -- "${repo_root}/tools/devtools" 2>/dev/null || finish 1
capture "${go_path}" version
if ((windows)); then [[ ${captured} == 'go version go1.26.5 windows/amd64' ]] || finish 1
else [[ ${captured} =~ ^go\ version\ go1\.26\.5\ [a-z0-9]+/[a-z0-9]+$ ]] || finish 1; fi
stage=dependencies
capture "${go_path}" list -m -f '{{if not .Main}}{{.Path}}|{{.Version}}|{{.Dir}}{{end}}' all
modules=${captured}
while IFS= read -r line; do
  line=${line%$'\r'}
  [[ -n ${line} ]] || continue
  IFS='|' read -r path version directory extra <<<"${line}"
  [[ ${path} != go && ${path} != toolchain ]] || continue
  [[ -z ${extra} && ${path} =~ ^[A-Za-z0-9._~+!/-]+$ && ${version} =~ ^v[A-Za-z0-9._+!-]+$ && ! ${path} =~ (^|/)\.\.(/|$) ]] || finish 1
  cache_key "${path}"; key=${escaped}
  cache_key "${version}"; version_key=${escaped}
  expected="${module_cache}/${key}@${version_key}"
  [[ -n ${directory} ]] || finish 1
  if ((windows)); then directory=$(cygpath -u "${directory}") || finish 1; fi
  [[ ${directory} == "${expected}" && -d ${expected} ]] || finish 1
  zip="${module_cache}/cache/download/${key}/@v/${version_key}.zip"
  [[ -f ${zip} && -f ${zip}hash ]] || finish 1
done <<<"${modules}"
stage=integrity
capture "${go_path}" mod verify
[[ ${captured} == 'all modules verified' ]] || finish 1
finish 0
