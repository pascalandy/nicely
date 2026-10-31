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
# ## Findings, and an empty $4 leaves out the line Reviewer:. The ledger
# holds $5 rounds, or 1, from the intake $6, and its last round reviewed $1.
fct_readiness() {
	local reviewer="${4-an agent that wrote none of the code}"
	{
		printf 'Commit: %s\n' "${1}"
		[[ -z "${reviewer}" ]] || printf 'Reviewer: %s\n' "${reviewer}"
		printf 'Verdict: %s\n\n# Readiness review\n' "${2}"
		[[ -z "${3}" ]] || printf '\n## Findings\n\n%s\n' "${3}"
		fct_ledger "${finish}" "${1}" "${5:-1}" "${6:-}"
	} >"${finish}/readiness.md"
}

# fct_ledger prints the ledger of a build of $3 rounds in the repository $1,
# whose last round reviewed $2. The build starts at the intake $4, or at the
# first commit. Earlier rounds name that intake, and round 1 commits its
# handler in $2 (#417).
fct_ledger() {
	local first="${4}" i
	[[ -n "${first}" ]] || first="$(git -C "${1}" rev-list --max-parents=0 HEAD)"
	printf '\n## Rounds\n\n| Round | Head | Change review | Change commits | Readiness review | Finish |\n|---|---|---|---|---|---|\n'
	for ((i = 1; i < ${3}; i++)); do
		printf '| %s | %s | agent A | %s..%s | agent B | `fail` |\n' "${i}" "${first}" "${first}" "${first}"
	done
	printf '| %s | %s | agent A | %s..%s | agent B | `pending` |\n' "${3}" "${2}" "${first}" "${2}"
	printf '\n## Order of round 1\n\n| Step | Commit |\n|---|---|\n'
	printf '| %s | %s |\n' Intake "${first}" Spec "${first}" Scenario "${first}" Handler "${2}"
}
