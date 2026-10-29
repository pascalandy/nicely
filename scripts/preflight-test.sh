#!/usr/bin/env bash
# preflight-test.sh proves just preflight for just test: it exits 0 once every
# prerequisite is present, warnings included, and 1 with a hint for each one
# that is missing. It runs on macOS and Linux, with bash 3.2.
#
# preflight.sh runs with a PATH of stubs and of links to the few real tools
# that it reads with, such as uname, so no tool of the host decides a case.
# Each case writes the stubs that it needs. See #356.
set -Eeuo pipefail
trap 'printf "%s: line %s failed: %s\n" "${BASH_SOURCE[0]##*/}" "${LINENO}" "${BASH_COMMAND}" >&2' ERR

PREFLIGHT_TEST_ROOT=""
test_case="setup"
head_commit="abcdef012345"

# shellcheck disable=SC2317,SC2329
fct_cleanup() {
	if [[ -n "${PREFLIGHT_TEST_ROOT}" ]]; then
		if [[ "$(uname -s)" == Darwin ]]; then chflags -R nouchg "${PREFLIGHT_TEST_ROOT}" 2>/dev/null || true; fi
		chmod -R u+w "${PREFLIGHT_TEST_ROOT}" 2>/dev/null || true
		rm -rf "${PREFLIGHT_TEST_ROOT}"
	fi
}

fct_fail() {
	printf 'preflight-test: %s: %s\n' "${test_case}" "$*" >&2
	exit 1
}

# fct_stub <folder> <name> <shell lines> writes an executable stub.
fct_stub() {
	mkdir -p "${root}/${1}"
	printf '#!/bin/sh\n%s\n' "${3}" >"${root}/${1}/${2}"
	chmod 755 "${root}/${1}/${2}"
}

# fct_present writes a stub for each prerequisite, each one present and current.
fct_present() {
	rm -rf "${root}/stubs" "${root}/second"
	fct_stub stubs tmux 'echo "tmux 3.5a"'
	fct_stub stubs python3 'echo 3.12.7'
	fct_stub stubs bash "echo '5.2.37(1)-release'"
	fct_stub stubs go 'exit 0'
	fct_stub stubs codex 'echo "codex-cli 0.165.0"'
	fct_stub stubs git "case \"\$*\" in *rev-parse*) echo ${head_commit} ;; esac"
	fct_stub stubs ncly "echo 'ncly v0.23.3-0.20261029120000-${head_commit}'"
}

fct_setup() {
	local tool path
	repo_root="$(cd "${BASH_SOURCE[0]%/*}/.." && pwd -P)"
	root="$(mktemp -d "${TMPDIR:-/tmp}/ncly-preflight-test.XXXXXXXX")"
	PREFLIGHT_TEST_ROOT="${root}"
	mkdir -p "${root}/tools"
	for tool in uname tr head stat sha256sum shasum; do
		if path="$(command -v "${tool}")"; then ln -s "${path}" "${root}/tools/${tool}"; fi
	done
	fct_present
}

# fct_run <expected exit> <arguments>... runs preflight.sh with the stubs first in PATH.
fct_run() {
	local expected="${1}" status=0
	shift
	env -i HOME="${root}" PATH="${root}/stubs:${root}/second:${root}/tools" "${BASH}" "${repo_root}/scripts/preflight.sh" "$@" >"${root}/output" 2>&1 </dev/null || status=$?
	if [[ "${status}" != "${expected}" ]]; then
		cat "${root}/output" >&2
		fct_fail "preflight exited ${status}, not ${expected}"
	fi
}

# fct_said <line> requires a line of the output that starts with the given text.
fct_said() {
	grep -F -q -- "${1}" "${root}/output" || {
		cat "${root}/output" >&2
		fct_fail "the output lacks: ${1}"
	}
}

# fct_hint <text> requires the line after the one that holds <text> to be a hint.
fct_hint() {
	grep -F -A1 -- "${1}" "${root}/output" | grep -q '^      hint: .' || {
		cat "${root}/output" >&2
		fct_fail "no hint follows: ${1}"
	}
}

# fct_fixture <name> writes a fixture folder with one original, its SHA256SUMS, and both locked.
fct_fixture() {
	local folder="${root}/fixtures/${1}"
	mkdir -p "${folder}"
	printf 'a real video\n' >"${folder}/clip.mov"
	if command -v sha256sum >/dev/null; then
		(cd "${folder}" && sha256sum clip.mov >SHA256SUMS)
	else
		(cd "${folder}" && shasum -a 256 clip.mov >SHA256SUMS)
	fi
	chmod 444 "${folder}/clip.mov" "${folder}/SHA256SUMS"
	if [[ "$(uname -s)" == Darwin ]]; then chflags uchg "${folder}/clip.mov" "${folder}/SHA256SUMS"; fi
}

fct_main() {
	trap fct_cleanup EXIT
	fct_setup

	test_case="every prerequisite present"
	fct_run 0
	fct_said "ok    tmux 3.5a"
	fct_said "ok    python3 3.12.7"
	fct_said "ok    bash 5.2.37(1)-release"
	fct_said "ok    Go modules of tools are in the module cache"
	fct_said "ok    codex-cli 0.165.0"
	fct_said "built from HEAD ${head_commit}"
	fct_said "preflight: 0 missing, 0 warnings"

	test_case="tmux, codex, and ncly missing, Python too old"
	rm "${root}/stubs/tmux" "${root}/stubs/codex" "${root}/stubs/ncly"
	fct_stub stubs python3 'echo 3.9.6; exit 1'
	fct_run 1
	fct_said "FAIL  tmux is missing"
	fct_hint "FAIL  tmux is missing"
	fct_said "FAIL  python3 3.9.6 at ${root}/stubs/python3: Python 3.12 or later required"
	fct_hint "FAIL  python3 3.9.6"
	fct_said "FAIL  codex is missing"
	fct_hint "FAIL  codex is missing"
	fct_said "FAIL  ncly is missing from PATH"
	fct_said "preflight: 4 missing, 0 warnings"

	test_case="Go modules missing from the cache"
	fct_present
	fct_stub stubs go 'exit 1'
	fct_run 1
	fct_said "FAIL  Go modules of tools are missing from the module cache"
	fct_hint "FAIL  Go modules of tools"

	test_case="bash 3.2 warns"
	fct_present
	fct_stub stubs bash "echo '3.2.57(1)-release'"
	fct_run 0
	fct_said "warn  bash 3.2.57(1)-release"
	fct_said "see #296"
	fct_hint "warn  bash 3.2.57"

	test_case="an ncly hides another, and comes from another commit"
	fct_present
	fct_stub stubs ncly "echo 'ncly v0.23.2-0.20261028120000-0123456789ab'"
	fct_stub second ncly "echo 'ncly v0.23.2'"
	fct_run 0
	fct_said "warn  ncly v0.23.2-0.20261028120000-0123456789ab at ${root}/stubs/ncly is not built from HEAD ${head_commit}"
	fct_said "warn  ${root}/stubs/ncly hides ${root}/second/ncly, ncly v0.23.2"
	fct_hint "hides"
	fct_said "preflight: 0 missing, 2 warnings"

	test_case="fixture folders"
	fct_present
	fct_fixture good
	fct_fixture changed
	if [[ "$(uname -s)" == Darwin ]]; then chflags nouchg "${root}/fixtures/changed/clip.mov"; fi
	chmod 644 "${root}/fixtures/changed/clip.mov"
	printf 'another video\n' >"${root}/fixtures/changed/clip.mov"
	fct_fixture unlocked
	if [[ "$(uname -s)" == Darwin ]]; then chflags nouchg "${root}/fixtures/unlocked/clip.mov"; fi
	chmod 644 "${root}/fixtures/unlocked/clip.mov"
	fct_run 1 "${root}/fixtures/good" "${root}/fixtures/changed" "${root}/fixtures/unlocked" "${root}/fixtures/gone"
	fct_said "ok    fixture folder ${root}/fixtures/good: checksums OK, originals locked"
	fct_said "FAIL  fixture folder ${root}/fixtures/changed: a checksum differs from SHA256SUMS"
	fct_said "FAIL  fixture ${root}/fixtures/unlocked/clip.mov is not locked"
	fct_said "FAIL  fixture folder ${root}/fixtures/gone holds no SHA256SUMS"
	fct_said "preflight: 3 missing, 0 warnings"
	printf 'just preflight exits 0 once every prerequisite is present, and 1 with a hint for each one that is missing\n'
}

fct_main
exit 0
