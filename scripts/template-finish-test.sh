#!/usr/bin/env bash
# Sourced by template-test.sh to prove just finish, the exit of the loop of
# the nicely-extension skill (#405). It works on a copy of the rendered
# repository, so the other tests keep their state. Finish reads the intake
# and the reviews only in their typed records (COR-104-T4).
# Markdown backticks in the fixture text are literal.
# shellcheck disable=SC2016

# shellcheck source=scripts/template-finish-helpers.sh
source "${BASH_SOURCE[0]%/*}/template-finish-helpers.sh"
# shellcheck source=scripts/template-finish-forms-test.sh
source "${BASH_SOURCE[0]%/*}/template-finish-forms-test.sh"
# shellcheck source=scripts/template-finish-matrix-test.sh
source "${BASH_SOURCE[0]%/*}/template-finish-matrix-test.sh"

fct_test_finish() {
	local finish="${2}/finish" scratch="${2}" sha file number records_status=0 ledger_earlier=""
	local free='this limit has no spec ID, no state limit, and no acceptance'
	local unfilled='this line of the form is not filled'
	local warn advise declined='reason = "the answer stays small"'
	warn="$(fct_finding R1-F1 warn open large-file)"
	advise="$(fct_finding R1-F2 advise declined lines "${declined}")"
	cp -R "${1}" "${finish}"
	fct_steps "${finish}"
	for file in finish records; do
		[[ "$(just --justfile "${finish}/justfile" --show "${file}" | sed '/^#/d')" == "${file}:"$'\n'"    go run github.com/pascalandy/nicely/sdk/extensionrepo/cmd/extensionrepo ${file}" ]] ||
			fct_fail "just ${file} must call the SDK program in one line"
	done

	# The state of trial 1: no open line, a limit that nobody accepted, and no readiness review.
	fct_edit "${finish}/spec.md" '**Limits.** None.' '**Limits.** A file over 1 GB takes more than 30 s to count.'
	git -C "${finish}" commit -qam 'state of trial 1'
	fct_finish 1
	fct_finish_said "${free}"
	fct_finish_said 'review.toml: no readiness review is recorded'
	fct_finish_said 'this line is open' absent
	# The view of the intake that the template renders is the one that just records writes.
	fct_finish_said 'is not the view of' absent
	# A new repository keeps each Fill: line of its forms, and of intake.toml.
	for file in vision.md spec.md qa.md intake.toml; do
		grep -n '^Fill:\|^# Fill:' "${finish}/${file}" >"${scratch}/fill.lines" || fct_fail "${file} of the template holds no Fill: line"
		while IFS=: read -r number _; do
			fct_finish_said "${file}:${number}: ${unfilled}"
		done <"${scratch}/fill.lines"
	done
	fct_fill "${finish}" "${scratch}"
	# An indented Fill: line is not filled either (#412).
	printf '  Fill: an indented line.\n' >>"${finish}/vision.md"
	fct_finish 1
	fct_finish_said "vision.md:$(grep -c '' "${finish}/vision.md"): ${unfilled}"
	fct_edit "${finish}/vision.md" '  Fill: an indented line.' ''

	fct_edit "${finish}/spec.md" 'None.' $'| ID | Command | Part | Case | Code |\n|---|---|---|---|---|\n| `large-file` | inspect | Failures | A file over 1 GB | `RUNTIME` |'
	fct_finish 1
	fct_finish_said 'this line is open'
	fct_finish_said "${unfilled}" absent

	fct_test_finish_intake

	# The state of trial 2: the author of the code wrote the review, with no
	# reviewer, and wrote the verdict Ready beside a warn.
	fct_edit "${finish}/spec.md" '**Limits.** A file over 1 GB takes more than 30 s to count.' '**Limits.** None.'
	git -C "${finish}" commit -qam 'state of trial 2'
	sha="$(git -C "${finish}" rev-parse HEAD)"
	fct_readiness "${sha}" Ready "${warn}" ''
	fct_finish 1
	fct_finish_said 'review.toml: it names no reviewer'
	fct_finish_said 'review.toml: the verdict is Ready, but the finding R1-F1 has the severity warn and is not fixed'
	fct_finish_said 'review.toml: the finding R1-F1 is open'
	# The open finding names its row of ## Open lines (#420).
	fct_finish_said 'spec.md: ## Open lines does not name' absent

	fct_edit "${finish}/spec.md" '| `large-file` | inspect | Failures | A file over 1 GB | `RUNTIME` |' ''
	fct_readiness "${sha}" Ready "${advise}"
	fct_finish 1
	fct_finish_said 'readiness.md: it is not the view of review.toml' absent
	fct_finish_said "review.toml: the review is of ${sha}, and these files changed since: spec.md"
	git -C "${finish}" commit -qam 'no open line'
	sha="$(git -C "${finish}" rev-parse HEAD)"
	fct_readiness "${sha}" Ready "${advise}"
	fct_review "${finish}" 'independent readiness review'
	ledger_earlier="${sha}"
	fct_finish 0
	fct_finish_said 'finish: no line or finding is open'
	fct_test_finish_forms
	# A view that differs from its record fails, so a reader never sees another state.
	fct_edit "${finish}/readiness.md" 'Verdict: Ready' 'Verdict: Needs work'
	fct_finish 1
	fct_finish_said 'readiness.md: it is not the view of review.toml. Fix: run just records, then commit both files'
	git -C "${finish}" checkout -q readiness.md

	fct_edit "${finish}/spec.md" '**Limits.** None.' $'**Limits.**\n\n| ID | State | Limit | Accepted |\n|---|---|---|---|\n| `large-file` | `open` | A file over 1 GB takes more than 30 s to count | The user, 2026-10-30, https://example.com/pull/1 |'
	fct_finish 1
	fct_finish_said 'the limit `large-file` has the state "open"'
	fct_edit "${finish}/spec.md" '| `large-file` | `open` | A file over 1 GB takes more than 30 s to count | The user, 2026-10-30, https://example.com/pull/1 |' '| `large-file` | `limit` | A file over 1 GB takes more than 30 s to count | The user, 2026-10-30, https://example.com/pull/1 |'
	git -C "${finish}" commit -qam 'accepted limit'
	sha="$(git -C "${finish}" rev-parse HEAD)"
	number="$(grep -n '^| `large-file` | `limit` |' "${finish}/spec.md" | cut -d: -f1)"
	fct_readiness "${sha}" 'Needs work' ''
	fct_finish 1
	fct_finish_said 'review.toml: the verdict is Needs work'
	fct_finish_said "${free}" absent

	# The verdict follows from the findings and the limits.
	fct_readiness "${sha}" Ready ''
	fct_finish 1
	fct_finish_said "spec.md:${number}: the verdict is Ready, but spec.md states the limit \`large-file\`"
	fct_readiness "${sha}" 'Ready with stated limits' "$(fct_finding R1-F1 warn limit small-file)"$'\n\n'"$(fct_finding R1-F3 block open large-file)"
	fct_finish 1
	fct_finish_said 'review.toml: the finding R1-F1 names no limit that the user accepted'
	fct_finish_said 'review.toml: the verdict is Ready with stated limits, but the finding R1-F3 has the severity block and is not fixed'

	# Finish refuses a verdict or a commit that it does not know.
	fct_readiness "${sha}" 'Ready with limits' ''
	fct_finish 1
	fct_finish_said 'review.toml: the verdict "Ready with limits" is not Ready, Ready with stated limits, or Needs work'
	fct_readiness 0123456789abcdef0123456789abcdef01234567 'Ready with stated limits' ''
	fct_finish 1
	fct_finish_said 'review.toml: Git knows no commit 0123456789abcdef0123456789abcdef01234567'

	# The plain words block, warn, and advise in prose grade nothing (#419).
	fct_readiness "${sha}" 'Ready with stated limits' "$(fct_finding R1-F1 warn limit large-file 'evidence = "Large files block the terminal, and inspect may warn. The severity is low for me."')"$'\n\n'"${advise}"
	fct_review "${finish}" 'readiness review'
	fct_finish 0
	fct_finish_said 'finish: no line or finding is open'
	fct_has "${finish}/readiness.md" '- **Evidence**: Large files block the terminal, and inspect may warn. The severity is low for me.'
	fct_test_finish_matrix

	# A change after the review needs a new review.
	printf 'A change after the review.\n' >>"${finish}/README.md"
	fct_finish 1
	fct_finish_said "review.toml: the review is of ${sha}, and these files changed since: README.md"
	git -C "${finish}" commit -qam 'a commit after the review'
	fct_finish 1
	fct_finish_said "review.toml: the review is of ${sha}, and these files changed since: README.md"

	# A missing form fails finish, which still reports the rest (#412).
	git -C "${finish}" rm -q qa.md spec.md intake.toml
	fct_finish 1
	fct_finish_said 'qa.md: the form is missing'
	fct_finish_said 'spec.md: the form is missing'
	fct_finish_said 'intake.toml: the form is missing'
	fct_finish_said 'review.toml: the finding R1-F1 names no limit that the user accepted'
	fct_finish_said 'no such file or directory' absent
}
