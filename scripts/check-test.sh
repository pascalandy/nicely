#!/usr/bin/env bash
# check-test.sh proves just check for just test: by default it runs the checks
# in order and stops at the first failure with its code, and with --keep-going
# it runs every check, then sums up each exit code. It runs on macOS and
# Linux, with bash 3.2.
#
# The caller's justfile is a copy of the real one, where each check that
# scripts/check.sh runs is a stub. A stub writes its name and arguments to the
# log, then exits with the code in code-<name>, 0 by default. See #357.
set -Eeuo pipefail
trap 'printf "%s: line %s failed: %s\n" "${BASH_SOURCE[0]##*/}" "${LINENO}" "${BASH_COMMAND}" >&2' ERR

CHECK_TEST_ROOT=""
test_case="setup"

# shellcheck disable=SC2317,SC2329
fct_cleanup() {
	if [[ -n "${CHECK_TEST_ROOT}" ]]; then
		rm -rf "${CHECK_TEST_ROOT}"
	fi
}

fct_fail() {
	printf 'check-test: %s: %s\n' "${test_case}" "$*" >&2
	exit 1
}

fct_setup() {
	local repo_root line check name
	local -a checks=()
	repo_root="$(cd "${BASH_SOURCE[0]%/*}/.." && pwd -P)"
	root="$(mktemp -d "${TMPDIR:-/tmp}/ncly-check-test.XXXXXXXX")"
	CHECK_TEST_ROOT="${root}"
	mkdir -p "${root}/caller/scripts"
	export NCLY_CHECK_TEST="${root}"
	cp "${repo_root}/justfile" "${root}/caller/justfile"
	cp "${repo_root}/scripts/check.sh" "${root}/caller/scripts/check.sh"
	# The list of checks comes from check.sh itself, so a new check needs no change here.
	line="$(grep '^checks=(' "${repo_root}/scripts/check.sh")"
	eval "${line}"
	[[ ${#checks[@]} -gt 0 ]] || fct_fail "check.sh lists no checks"
	printf '\nset allow-duplicate-recipes\n' >>"${root}/caller/justfile"
	for check in "${checks[@]}"; do
		name="${check%% *}"
		# shellcheck disable=SC2016 # the stub recipe expands these
		printf '\n[positional-arguments]\n%s *args:\n    @echo "%s${*:+ $*}" >>"${NCLY_CHECK_TEST}/log"; exit "$(cat "${NCLY_CHECK_TEST}/code-%s" 2>/dev/null || echo 0)"\n' \
			"${name}" "${name}" "${name}" >>"${root}/caller/justfile"
	done
	printf '%s\n' "${checks[@]}" >"${root}/log.all"
}

# fct_run <expected exit> <just check arguments>... runs just check in the caller.
fct_run() {
	local expected="${1}" status=0
	shift
	: >"${root}/log"
	(cd "${root}/caller" && just check "$@") >"${root}/output" 2>&1 </dev/null || status=$?
	if [[ "${status}" != "${expected}" ]]; then
		cat "${root}/output" >&2
		fct_fail "just check $* exited ${status}, not ${expected}"
	fi
}

# fct_log <lines of log.all> requires the log to hold the first lines of log.all.
fct_log() {
	# BSD head refuses -n 0, so awk takes the first lines.
	awk -v n="${1}" 'NR <= n' "${root}/log.all" >"${root}/log.expected"
	if ! cmp -s "${root}/log.expected" "${root}/log"; then
		diff "${root}/log.expected" "${root}/log" >&2 || true
		fct_fail "ran other checks"
	fi
}

fct_said() {
	grep -F -q -- "${1}" "${root}/output" || {
		cat "${root}/output" >&2
		fct_fail "the output lacks: ${1}"
	}
}

fct_main() {
	local all position
	trap fct_cleanup EXIT
	fct_setup
	all="$(wc -l <"${root}/log.all" | tr -d ' ')"
	position="$(grep -n '^test$' "${root}/log.all" | cut -d: -f1)"

	test_case="every check passes"
	fct_run 0
	fct_log "${all}"
	grep -q '^limits --base origin/main$' "${root}/log" || fct_fail "limits lost its arguments"

	test_case="test fails"
	echo 3 >"${root}/code-test"
	fct_run 3
	fct_log "${position}"

	test_case="test and gitleaks fail, with --keep-going"
	echo 1 >"${root}/code-gitleaks"
	fct_run 1 --keep-going
	fct_log "${all}"
	fct_said "test             exit 3"
	fct_said "gitleaks         exit 1"
	fct_said "fmt-check        exit 0"
	fct_said "2 of ${all} checks failed"

	test_case="every check passes, with --keep-going"
	rm -f "${root}/code-test" "${root}/code-gitleaks"
	fct_run 0 --keep-going
	fct_log "${all}"
	fct_said "every check passed"

	test_case="JUST_NO_DEPS skips no check"
	(
		export JUST_NO_DEPS=true
		fct_run 0
	)
	fct_log "${all}"

	test_case="unknown flag"
	fct_run 2 --keep-goin
	fct_log 0
	fct_said "usage: just check [--keep-going]"
	printf 'just check stops at the first failure with its code, and --keep-going runs every check, then sums them up\n'
}

fct_main
exit 0
