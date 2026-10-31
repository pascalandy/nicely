#!/usr/bin/env bash
# Shared calls of the proof of just finish in template-finish-test.sh.
# shellcheck disable=SC2154 # fct_test_finish sets finish and scratch for these helpers.

# fct_finish runs just finish in the copy, and requires the exit code $1.
fct_finish() {
	local status=0
	(cd "${finish}" && just finish) >"${scratch}/finish.out" 2>&1 || status=$?
	if [[ "${status}" != "${1}" ]]; then
		cat "${scratch}/finish.out" >&2
		fct_fail "just finish exited ${status}, want ${1}"
	fi
}

# fct_finish_said requires the text $1 in the output of the last just finish,
# or its absence when $2 is "absent".
fct_finish_said() {
	local found=0
	grep -qF -- "${1}" "${scratch}/finish.out" || found=$?
	if [[ "${2:-}" == absent && "${found}" == 0 || "${2:-}" != absent && "${found}" != 0 ]]; then
		cat "${scratch}/finish.out" >&2
		fct_fail "just finish: want ${2:-present} \"${1}\""
	fi
}

# fct_readiness writes readiness.md for the commit $1, with the verdict $2, the
# findings $3 under ## Findings, and the reviewer $4. An empty $3 leaves out
# ## Findings, and an empty $4 leaves out the line Reviewer:.
fct_readiness() {
	local reviewer="${4-an agent that wrote none of the code}"
	{
		printf 'Commit: %s\n' "${1}"
		[[ -z "${reviewer}" ]] || printf 'Reviewer: %s\n' "${reviewer}"
		printf 'Verdict: %s\n\n# Readiness review\n' "${2}"
		[[ -z "${3}" ]] || printf '\n## Findings\n\n%s\n' "${3}"
	} >"${finish}/readiness.md"
}
