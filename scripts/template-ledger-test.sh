#!/usr/bin/env bash
# Sourced by template-test.sh to prove the ledger of readiness.md, which
# just finish reads (#417), and the forms of readiness.md of #416. It works on
# a copy of the rendered repository, so the other tests keep their state.
# Markdown backticks in the fixture text are literal.
# shellcheck disable=SC2016

fct_test_ledger() {
	local finish="${2}/ledger" scratch="${2}" file first round2 head stray valid
	local advise=$'### R1-F1 The count could name the longest line\n\n- **Severity**: `advise`'
	local declined=$'\n- **State**: `declined`\n- **Reason**: the answer stays small'
	local warn=$'### R2-F1 Long lines are cut\n\n- **Severity**: `warn`'
	local lowered=$'### R2-F1 Long lines are cut\n\n- **Severity**: `advise`'
	cp -R "${1}" "${finish}"
	for file in vision.md spec.md qa.md; do
		grep -v '^Fill:' "${finish}/${file}" >"${scratch}/form.md"
		mv "${scratch}/form.md" "${finish}/${file}"
	done
	git -C "${finish}" commit -qam 'fill the forms'
	first="$(git -C "${finish}" rev-list --max-parents=0 HEAD)"

	# Round 1 reports an advise, and round 2 declines it and reports a warn.
	# Their committed reviews are the history of round 3.
	fct_readiness "$(git -C "${finish}" rev-parse HEAD)" 'Needs work' "${advise}"
	git -C "${finish}" add readiness.md
	git -C "${finish}" commit -qm 'readiness review of round 1'
	printf 'Round 2.\n' >>"${finish}/README.md"
	git -C "${finish}" commit -qam 'round 2'
	round2="$(git -C "${finish}" rev-parse HEAD)"
	fct_readiness "${round2}" 'Needs work' "${advise}${declined}"$'\n\n'"${warn}" 'agent B' 2
	git -C "${finish}" add readiness.md
	git -C "${finish}" commit -qm 'readiness review of round 2'
	printf 'Round 3.\n' >>"${finish}/README.md"
	git -C "${finish}" commit -qam 'round 3'
	head="$(git -C "${finish}" rev-parse HEAD)"

	# A warn of round 2 that the ledger of round 3 does not record fails finish.
	fct_readiness "${head}" Ready "${advise}${declined}" 'agent B' 3
	fct_finish 1
	fct_finish_said 'readiness.md: the finding R2-F1 of the review in'
	# A finding keeps its highest severity, so a warn never closes as a declined advise.
	fct_readiness "${head}" Ready "${advise}${declined}"$'\n\n'"${lowered}${declined}" 'agent B' 3
	fct_finish 1
	fct_finish_said 'the finding R2-F1 had the severity warn'
	fct_readiness "${head}" Ready "${advise}${declined}"$'\n\n'"${warn}${declined}" 'agent B' 3
	fct_finish 1
	fct_finish_said 'the finding R2-F1 has the severity warn, and only an advise can be declined'
	fct_readiness "${head}" Ready "${advise}${declined}"$'\n\n'"${warn}" 'agent B' 3
	fct_finish 1
	fct_finish_said 'the finding R2-F1 of round 2 has no state'
	# None. under ## Open lines does not close a finding that the ledger records as open.
	fct_readiness "${head}" 'Ready with stated limits' "${advise}${declined}"$'\n\n'"${warn}"$'\n- **State**: `open`\n- **Spec**: `long-lines`' 'agent B' 3
	fct_finish 1
	fct_finish_said 'the finding R2-F1 is open, even if ## Open lines holds None.'
	fct_readiness "${head}" Ready "${advise}${declined}"$'\n\n'"${warn}"$'\n- **State**: `fixed`\n- **Spec**: `long-lines`' 'agent B' 3
	fct_finish 1
	fct_finish_said 'this fixed finding names no commit of HEAD'
	fct_finish_said 'this fixed finding names no line of Failures or Acceptance in spec.md'
	valid="${advise}${declined}"$'\n\n'"${warn}"$'\n- **State**: `fixed`\n- **Commit**: '"${head}"$'\n- **Spec**: `bad-input`'
	fct_readiness "${head}" Ready "${valid}" 'agent B' 3
	git -C "${finish}" add readiness.md
	git -C "${finish}" commit -qm 'readiness review of round 3'
	fct_finish 0

	# A build has at most 5 rounds, and finish tells the agent to stop.
	fct_readiness "${head}" Ready "${valid}" 'agent B' 6
	fct_finish 1
	fct_finish_said 'readiness.md: the ledger holds 6 rounds, and a build has at most 5'
	fct_finish_said 'this build used its 5 rounds. Stop, and report each open line and each open finding to the user'
	fct_readiness "${head}" Ready "${valid}" 'agent B' 3
	stray="$(git -C "${finish}" commit-tree -m stray "${head}^{tree}")"
	fct_edit "${finish}/readiness.md" "| 3 | ${head} | agent A | ${first}..${head} | agent B | \`pending\` |" "| 3 | ${head} | agent A | ${stray}..${head} | agent B | \`pending\` |"
	fct_finish 1
	fct_finish_said "the commit \"${stray}\" is not an ancestor of HEAD"
	fct_readiness "${head}" Ready "${valid}" 'agent B' 3
	fct_edit "${finish}/readiness.md" "| 3 | ${head} | agent A | ${first}..${head} | agent B | \`pending\` |" "| 3 | ${round2} | agent A | ${first}..${head} | agent B | \`pending\` |"
	fct_finish 1
	fct_finish_said "the last round names the head ${round2}, and the line Commit: names ${head}"
	# Round 1 commits its first scenario before its handler.
	fct_readiness "${head}" Ready "${valid}" 'agent B' 3
	fct_edit "${finish}/readiness.md" "| Scenario | ${first} |" "| Scenario | ${head} |"
	fct_finish 1
	fct_finish_said 'the handler and the first scenario are in one commit'
	printf 'Commit: %s\nReviewer: agent B\nVerdict: Ready\n\n## Findings\n\n%s\n' "${head}" "${valid}" >"${finish}/readiness.md"
	fct_finish 1
	fct_finish_said 'readiness.md: it holds no round under ## Rounds'

	# The word severity in prose, or a table of severities, fails with its cause (#416).
	fct_readiness "${head}" Ready "${advise}"$'\n- **Evidence**: I kept the severity low'"${declined}"$'\n\n| Finding | Grade |\n|---|---|\n| R1-F1 | advise |' 'agent B' 3
	fct_finish 1
	fct_finish_said 'this line names the word severity outside the line - **Severity**: of a finding. Fix: grade each finding only in its line - **Severity**:, and rephrase this line without the word severity'
	fct_finish_said 'this table names the severity advise'
	# A warn in prose beside an advise, or in a table of a part after ## Findings, fails (#416).
	fct_readiness "${head}" Ready "${valid}"$'\n\nThis is a warn.\n\n## Notes\n\n| Long lines kept | `warn` |' 'agent B' 3
	fct_finish 1
	fct_finish_said 'this line names the severity warn outside the line - **Severity**: of a finding'
	fct_finish_said 'this table names the severity warn'
	# Finish reads the header in the first three lines, after a byte order mark (#416).
	fct_readiness "${head}" Ready "${valid}" 'agent B' 3
	printf '# Readiness review\n' | cat - "${finish}/readiness.md" >"${scratch}/readiness.md"
	mv "${scratch}/readiness.md" "${finish}/readiness.md"
	fct_finish 1
	fct_finish_said 'readiness.md: its first three lines hold no line Commit: <sha> or Verdict: <verdict>'
	git -C "${finish}" checkout -q readiness.md
	printf '\357\273\277' | cat - "${finish}/readiness.md" >"${scratch}/readiness.md"
	mv "${scratch}/readiness.md" "${finish}/readiness.md"
	fct_finish 0

	# A new build, such as of another command, starts at its own intake.
	# The findings of the build before it are no history of the new build.
	git -C "${finish}" checkout -q readiness.md
	printf 'Intake of the next build.\n' >>"${finish}/README.md"
	git -C "${finish}" commit -qam 'intake of the next build'
	first="$(git -C "${finish}" rev-parse HEAD)"
	printf 'Handler of the next build.\n' >>"${finish}/README.md"
	git -C "${finish}" commit -qam 'handler of the next build'
	fct_readiness "$(git -C "${finish}" rev-parse HEAD)" Ready "${advise}${declined}" 'agent B' 1 "${first}"
	fct_finish 0
}
