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

# fct_records runs just records in the copy, which writes the views of the
# build records (COR-104-T4). It keeps the output and the exit code in
# records.out and records_status, for fct_records_said.
fct_records() {
	records_status=0
	(cd "${finish}" && just records) >"${scratch}/records.out" 2>&1 || records_status=$?
}

# fct_records_said requires the exit code $1 of the last just records, and
# the text $2 in its output.
fct_records_said() {
	if [[ "${records_status}" != "${1}" ]] || ! grep -qF -- "${2}" "${scratch}/records.out"; then
		cat "${scratch}/records.out" >&2
		fct_fail "just records exited ${records_status}, want ${1} and \"${2}\""
	fi
}

# fct_readiness writes review.toml for the commit $1, with the verdict $2, the
# [[findings]] tables $3, and the reviewer $4, then runs just records. An
# empty $4 leaves out the key reviewer. The rounds of the ledger reviewed the
# heads of ledger_earlier, then $1.
fct_readiness() {
	local reviewer="${4-an agent that wrote none of the code}"
	{
		printf 'commit = "%s"\n' "${1}"
		[[ -z "${reviewer}" ]] || printf 'reviewer = "%s"\n' "${reviewer}"
		printf 'verdict = "%s"\n' "${2}"
		[[ -z "${3}" ]] || printf '\n%s\n' "${3}"
		fct_ledger "${finish}" "${ledger_earlier:-} ${1}"
	} >"${finish}/review.toml"
	fct_records
}

# fct_review commits the records of the readiness review and their view, in
# the repository $1, with the message $2.
fct_review() {
	git -C "${1}" add review.toml readiness.md
	git -C "${1}" commit -qm "${2}"
}

# fct_fill deletes each Fill: line of the forms of the repository $1, and each
# # Fill: line of intake.toml, in the scratch folder $2.
fct_fill() {
	local file
	for file in vision.md spec.md qa.md intake.toml; do
		grep -v '^Fill:\|^# Fill:' "${1}/${file}" >"${2}/form.md"
		mv "${2}/form.md" "${1}/${file}"
	done
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

# fct_finding prints a [[findings]] table with the ID $1, the severity $2, the
# state $3, and the spec ID $4, in the family of the spec ID bad-input. The
# other arguments are more lines of the table, such as commit = "<sha>".
fct_finding() {
	printf '[[findings]]\nid = "%s"\ntitle = "Finding %s"\nseverity = "%s"\nstate = "%s"\nspec = "%s"\nfamily = "The input is a regular file that exists"\n' "${1}" "${1}" "${2}" "${3}" "${4}"
	shift 4
	[[ "$#" == 0 ]] || printf '%s\n' "$@"
}

# fct_ledger prints the ledger of the repository $1 for the rounds that
# reviewed the heads $2, in order. A head written <sha>:pass gives its round
# the finish pass, which ends its build. The newest commits of fct_steps give
# the order of round 1 of the last build, and its handler is in the first
# head of that build (#417). ledger_run holds the table [first_run]. Unset,
# it records a run of the newest scenario step that failed on an assertion.
# Empty, it leaves the table out (#420).
fct_ledger() {
	local heads first i head finish_cell handler="" run
	run=$'[first_run]\ncommit = "'"$(fct_step "${1}" scenario)"$'"\nscenario = "inspect"\nexit = 1\nassertion = \'\'\'stdout \'"lines":3\' of testdata/script/inspect.txtar found no match\'\'\''
	read -r -a heads <<<"${2}"
	first="$(git -C "${1}" rev-list --max-parents=0 HEAD)"
	for ((i = 0; i < ${#heads[@]}; i++)); do
		head="${heads[i]%%:*}" finish_cell=fail
		[[ -n "${handler}" ]] || handler="${head}"
		[[ "${heads[i]}" != *:pass ]] || finish_cell=pass handler=""
		[[ "$((i + 1))" != "${#heads[@]}" ]] || finish_cell=pending
		printf '\n[[rounds]]\nhead = "%s"\nchange_review = "agent A"\nchange_commits = "%s..%s"\nreadiness_review = "agent B"\nfinish = "%s"\n' "${head}" "${first}" "${head}" "${finish_cell}"
	done
	printf '\n[order]\nintake = "%s"\nspec = "%s"\nscenario = "%s"\nhandler = "%s"\n' "$(fct_step "${1}" intake)" "$(fct_step "${1}" spec)" "$(fct_step "${1}" scenario)" "${handler}"
	[[ -z "${ledger_run-${run}}" ]] || printf '\n%s\n' "${ledger_run-${run}}"
}
