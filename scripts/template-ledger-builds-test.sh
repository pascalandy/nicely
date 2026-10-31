#!/usr/bin/env bash
# Sourced by template-test.sh. fct_test_ledger calls fct_test_ledger_builds
# at the end of round 3, with its repository and its fixture variables. It
# proves the forms of readiness.md of #416, and the start of each build (#417).
# Markdown backticks in the fixture text are literal, and fct_readiness reads
# ledger_earlier.
# shellcheck disable=SC2016,SC2034,SC2154

fct_test_ledger_builds() {
	local base forged kept head2 head3 i limit=$'**Limits.**\n\n| ID | State | Limit | Accepted |\n|---|---|---|---|\n| `long-lines` | `limit` | A line over 1 MB is cut | The user, 2026-10-31, https://example.com/pull/1 |'
	git -C "${finish}" checkout -q readiness.md spec.md
	base="$(git -C "${finish}" rev-parse HEAD)"
	# After a limit that the user accepted, or a review of the form of #409, a
	# round with the Finish pass ends the build, and the next build starts.
	for forged in ledger legacy; do
		if [[ "${forged}" == ledger ]]; then
			kept="${advise}${declined}"$'\n\n'"${warn}"$'\n- **State**: `limit`\n- **Spec**: `long-lines`'
			fct_readiness "${head}" 'Ready with stated limits' "${kept}" 'agent B'
		else
			kept="${valid}"
			printf 'Commit: %s\nReviewer: agent B\nVerdict: Ready with stated limits\n\n## Findings\n\n### Long lines are cut\n\n- **Severity**: `warn`\n- **Limit**: `long-lines`\n' "${head}" >"${finish}/readiness.md"
		fi
		git -C "${finish}" add readiness.md
		git -C "${finish}" commit -qm "a review with an accepted limit: ${forged}"
		fct_edit "${finish}/spec.md" '**Limits.** None.' "${limit}"
		git -C "${finish}" commit -qam 'the user accepts the limit long-lines'
		fct_steps "${finish}"
		printf 'Handler of the build after the limit.\n' >>"${finish}/README.md"
		git -C "${finish}" commit -qam 'handler of the build after the limit'
		ledger_earlier="${round1} ${round2} ${head}:pass"
		fct_readiness "$(git -C "${finish}" rev-parse HEAD)" 'Ready with stated limits' "${kept}" 'agent B'
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

	# A new build, such as of another command, starts at its own intake after
	# a round with the Finish pass. The ledger keeps the findings of each build.
	git -C "${finish}" checkout -q readiness.md
	fct_steps "${finish}"
	printf 'Handler of the next build.\n' >>"${finish}/README.md"
	git -C "${finish}" commit -qam 'handler of the next build'
	head2="$(git -C "${finish}" rev-parse HEAD)" ledger_earlier="${round1} ${round2} ${head}:pass"
	fct_readiness "${head2}" Ready "${valid}" 'agent B'
	fct_finish 0
	git -C "${finish}" add readiness.md
	git -C "${finish}" commit -qm 'readiness review of the next build'
	base="$(git -C "${finish}" rev-parse HEAD)"
	# Ten builds that each end with a review that passes still pass.
	for i in 3 4 5 6 7 8 9 10; do
		ledger_earlier="${ledger_earlier} $(git -C "${finish}" log --format=%H -n 1 HEAD~1):pass"
		fct_steps "${finish}"
		printf 'Handler of build %s.\n' "${i}" >>"${finish}/README.md"
		git -C "${finish}" commit -qam "handler of build ${i}"
		fct_readiness "$(git -C "${finish}" rev-parse HEAD)" Ready "${valid}" 'agent B'
		git -C "${finish}" add readiness.md
		git -C "${finish}" commit -qm "readiness review of build ${i}"
	done
	fct_finish 0
	git -C "${finish}" reset -q --hard "${base}"

	# Only a review that passes ends a build: a verdict line that says Ready,
	# a state with no commit, or a review that drops a finding ends none.
	fct_steps "${finish}"
	printf 'Handler of a build that needs work.\n' >>"${finish}/README.md"
	git -C "${finish}" commit -qam 'handler of a build that needs work'
	head3="$(git -C "${finish}" rev-parse HEAD)" ledger_earlier="${round1} ${round2} ${head}:pass ${head2}:pass"
	fct_readiness "${head3}" 'Needs work' "${valid}"$'\n\n### Blank lines are dropped\n\n- **Severity**: `block`' 'agent B'
	git -C "${finish}" add readiness.md
	git -C "${finish}" commit -qm 'readiness review of the build that needs work'
	base="$(git -C "${finish}" rev-parse HEAD)"
	for forged in verdict state none; do
		case "${forged}" in
		verdict) fct_edit "${finish}/readiness.md" 'Verdict: Needs work' 'Verdict: Ready' ;;
		state)
			fct_edit "${finish}/readiness.md" 'Verdict: Needs work' 'Verdict: Ready'
			fct_edit "${finish}/readiness.md" '- **Severity**: `block`' $'- **Severity**: `block`\n- **State**: `fixed`'
			;;
		none) fct_readiness "${head3}" Ready 'None.' 'agent B' ;;
		esac
		git -C "${finish}" commit -qam "a review that ends the build: ${forged}"
		fct_steps "${finish}"
		printf 'Handler of the build after it.\n' >>"${finish}/README.md"
		git -C "${finish}" commit -qam 'handler of the build after it'
		ledger_earlier="${round1} ${round2} ${head}:pass ${head2}:pass ${head3}:pass"
		fct_readiness "$(git -C "${finish}" rev-parse HEAD)" Ready "${valid}"$'\n\n### R5-F1 Blank lines are dropped\n\n- **Severity**: `block`\n- **State**: `fixed`\n- **Commit**: '"${head3}"$'\n- **Spec**: `bad-input`' 'agent B'
		fct_finish 1
		fct_finish_said 'round 5 says pass, but its review in'
		git -C "${finish}" reset -q --hard "${base}"
	done
}
