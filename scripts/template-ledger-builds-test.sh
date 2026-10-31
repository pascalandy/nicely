#!/usr/bin/env bash
# Sourced by template-test.sh. fct_test_ledger calls fct_test_ledger_builds
# at the end of round 3, with its repository and its fixture variables. It
# proves the forms of readiness.md of #416, and the start of each build (#417).
# Markdown backticks in the fixture text are literal, and fct_readiness reads
# ledger_earlier.
# shellcheck disable=SC2016,SC2034,SC2154

fct_test_ledger_builds() {
	local base forged limit=$'**Limits.**\n\n| ID | State | Limit | Accepted |\n|---|---|---|---|\n| `long-lines` | `limit` | A line over 1 MB is cut | The user, 2026-10-31, https://example.com/pull/1 |'
	git -C "${finish}" checkout -q readiness.md spec.md
	base="$(git -C "${finish}" rev-parse HEAD)"
	# After a limit that the user accepted, or a review of the form of #409,
	# a second build starts with a history of its own.
	for forged in ledger legacy; do
		if [[ "${forged}" == ledger ]]; then
			fct_readiness "${head}" 'Ready with stated limits' "${advise}${declined}"$'\n\n'"${warn}"$'\n- **State**: `limit`\n- **Spec**: `long-lines`' 'agent B'
		else
			printf 'Commit: %s\nReviewer: agent B\nVerdict: Ready with stated limits\n\n## Findings\n\n### Long lines are cut\n\n- **Severity**: `warn`\n- **Limit**: `long-lines`\n' "${head}" >"${finish}/readiness.md"
		fi
		git -C "${finish}" add readiness.md
		git -C "${finish}" commit -qm "a review with an accepted limit: ${forged}"
		fct_edit "${finish}/spec.md" '**Limits.** None.' "${limit}"
		git -C "${finish}" commit -qam 'the user accepts the limit long-lines'
		fct_steps "${finish}"
		printf 'Handler of the build after the limit.\n' >>"${finish}/README.md"
		git -C "${finish}" commit -qam 'handler of the build after the limit'
		ledger_earlier=""
		fct_readiness "$(git -C "${finish}" rev-parse HEAD)" 'Ready with stated limits' 'None.' 'agent B'
		fct_finish 0
		ledger_earlier="${round1} ${round2}"
		git -C "${finish}" reset -q --hard "${base}"
	done

	# The word severity in prose, or a table of severities, fails with its cause (#416).
	fct_readiness "${head}" Ready "${advise}"$'\n- **Evidence**: I kept the severity low'"${declined}"$'\n\n| Finding | Grade |\n|---|---|\n| R1-F1 | advise |' 'agent B'
	fct_finish 1
	fct_finish_said 'this line names the word severity outside the line - **Severity**: of a finding. Fix: grade each finding only in its line - **Severity**:, and rephrase this line without the word severity'
	fct_finish_said 'this table names the severity advise'
	# A warn in prose beside an advise, or in a table of a part after ## Findings, fails (#416).
	fct_readiness "${head}" Ready "${valid}"$'\n\nThis is a warn.\n\n## Notes\n\n| Long lines kept | `warn` |' 'agent B'
	fct_finish 1
	fct_finish_said 'this line names the severity warn outside the line - **Severity**: of a finding'
	fct_finish_said 'this table names the severity warn'
	# Finish reads the header in the first three lines, after a byte order mark (#416).
	fct_readiness "${head}" Ready "${valid}" 'agent B'
	printf '# Readiness review\n' | cat - "${finish}/readiness.md" >"${scratch}/readiness.md"
	mv "${scratch}/readiness.md" "${finish}/readiness.md"
	fct_finish 1
	fct_finish_said 'readiness.md: its first three lines hold no line Commit: <sha> or Verdict: <verdict>'
	git -C "${finish}" checkout -q readiness.md
	printf '\357\273\277' | cat - "${finish}/readiness.md" >"${scratch}/readiness.md"
	mv "${scratch}/readiness.md" "${finish}/readiness.md"
	fct_finish 0

	# A new build, such as of another command, starts at its own intake,
	# after a review with a passing verdict. The findings of the build before
	# it are no history of the new build.
	git -C "${finish}" checkout -q readiness.md
	fct_steps "${finish}"
	printf 'Handler of the next build.\n' >>"${finish}/README.md"
	git -C "${finish}" commit -qam 'handler of the next build'
	ledger_earlier=""
	fct_readiness "$(git -C "${finish}" rev-parse HEAD)" Ready "${advise}${declined}" 'agent B'
	fct_finish 0

	# A review that needs work ends no build, even when its verdict line says
	# Ready. A finding without an ID keeps its title in later reviews.
	fct_readiness "$(git -C "${finish}" rev-parse HEAD)" 'Needs work' "${warn/R2-F1 /}"
	git -C "${finish}" add readiness.md
	git -C "${finish}" commit -qm 'readiness review of the next build'
	base="$(git -C "${finish}" rev-parse HEAD)"
	for forged in verdict state none; do
		case "${forged}" in
		verdict) fct_edit "${finish}/readiness.md" 'Verdict: Needs work' 'Verdict: Ready' ;;
		state)
			fct_edit "${finish}/readiness.md" 'Verdict: Needs work' 'Verdict: Ready'
			fct_edit "${finish}/readiness.md" '- **Severity**: `warn`' $'- **Severity**: `warn`\n- **State**: `fixed`'
			;;
		none) fct_readiness "$(git -C "${finish}" rev-parse HEAD~1)" Ready 'None.' 'agent B' ;;
		esac
		git -C "${finish}" commit -qam "a review that ends the build: ${forged}"
		fct_steps "${finish}"
		printf 'Handler of a third build.\n' >>"${finish}/README.md"
		git -C "${finish}" commit -qam 'handler of a third build'
		fct_readiness "$(git -C "${finish}" rev-parse HEAD)" Ready 'None.' 'agent B'
		fct_finish 1
		fct_finish_said 'readiness.md: the finding titled "Long lines are cut" of the review in'
		git -C "${finish}" reset -q --hard "${base}"
	done
}
