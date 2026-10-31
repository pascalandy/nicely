#!/usr/bin/env bash
# Shared calls of the proof of just finish in template-finish-test.sh.
# shellcheck disable=SC2154 # fct_test_finish sets finish and scratch for these helpers.
# Markdown backticks in the fixture text are literal.
# shellcheck disable=SC2016

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
# ## Findings, and an empty $4 leaves out the line Reviewer:. The rounds of the
# ledger reviewed the heads of ledger_earlier, then $1.
fct_readiness() {
	local reviewer="${4-an agent that wrote none of the code}"
	{
		printf 'Commit: %s\n' "${1}"
		[[ -z "${reviewer}" ]] || printf 'Reviewer: %s\n' "${reviewer}"
		printf 'Verdict: %s\n\n# Readiness review\n' "${2}"
		[[ -z "${3}" ]] || printf '\n## Findings\n\n%s\n' "${3}"
		fct_ledger "${finish}" "${ledger_earlier:-} ${1}"
	} >"${finish}/readiness.md"
}

# fct_steps commits the first three steps of round 1 in the repository $1,
# as empty commits, so the ledger can name their order (#417).
fct_steps() {
	local step
	for step in intake spec scenario; do
		git -C "${1}" commit -q --allow-empty -m "step: ${step}"
	done
}

# fct_step prints the newest commit of the step $2 in the repository $1.
fct_step() {
	git -C "${1}" log --format=%H -n 1 -F --grep="step: ${2}" HEAD
}

# fct_ledger prints the ledger of the repository $1 for the rounds that
# reviewed the heads $2, in order. The newest commits of fct_steps give the
# order of round 1, and its handler is in the first head (#417).
fct_ledger() {
	local heads intake i finish_cell
	read -r -a heads <<<"${2}"
	intake="$(fct_step "${1}" intake)"
	printf '\n## Rounds\n\n| Round | Head | Change review | Change commits | Readiness review | Finish |\n|---|---|---|---|---|---|\n'
	for ((i = 0; i < ${#heads[@]}; i++)); do
		finish_cell='`fail`'
		[[ "$((i + 1))" != "${#heads[@]}" ]] || finish_cell='`pending`'
		printf '| %s | %s | agent A | %s..%s | agent B | %s |\n' "$((i + 1))" "${heads[i]}" "${intake}" "${heads[i]}" "${finish_cell}"
	done
	printf '\n## Order of round 1\n\n| Step | Commit |\n|---|---|\n'
	printf '| %s | %s |\n' Intake "${intake}" Spec "$(fct_step "${1}" spec)" Scenario "$(fct_step "${1}" scenario)" Handler "${heads[0]:-}"
}
