#!/usr/bin/env bash
# Sourced by template-test.sh to prove the ledger of readiness.md, which
# just finish reads (#417), and the forms of readiness.md of #416. It works on
# a copy of the rendered repository, so the other tests keep their state.
# Markdown backticks in the fixture text are literal, and fct_readiness reads
# ledger_earlier.
# shellcheck disable=SC2016,SC2034

fct_test_ledger() {
	local finish="${2}/ledger" scratch="${2}" file first intake round1 review1 round2 review2 head stray valid ledger_earlier=""
	local advise=$'### R1-F1 The count could name the longest line\n\n- **Severity**: `advise`'
	local declined=$'\n- **State**: `declined`\n- **Reason**: the answer stays small'
	local warn=$'### R2-F1 Long lines are cut\n\n- **Severity**: `warn`'
	local lowered=$'### R2-F1 Long lines are cut\n\n- **Severity**: `advise`'
	cp -R "${1}" "${finish}"
	fct_steps "${finish}"
	intake="$(fct_step "${finish}" intake)" first="$(git -C "${finish}" rev-list --max-parents=0 HEAD)"
	for file in vision.md spec.md qa.md; do
		grep -v '^Fill:' "${finish}/${file}" >"${scratch}/form.md"
		mv "${scratch}/form.md" "${finish}/${file}"
	done
	git -C "${finish}" commit -qam 'fill the forms'
	round1="$(git -C "${finish}" rev-parse HEAD)"

	# Round 1 reports an advise, and round 2 declines it and reports a warn.
	# Their committed reviews are the history of round 3.
	fct_readiness "${round1}" 'Needs work' "${advise}"
	git -C "${finish}" add readiness.md
	git -C "${finish}" commit -qm 'readiness review of round 1'
	review1="$(git -C "${finish}" rev-parse HEAD)" ledger_earlier="${round1}"
	printf 'Round 2.\n' >>"${finish}/README.md"
	git -C "${finish}" commit -qam 'round 2'
	round2="$(git -C "${finish}" rev-parse HEAD)"
	fct_readiness "${round2}" 'Needs work' "${advise}${declined}"$'\n\n'"${warn}" 'agent B'
	git -C "${finish}" add readiness.md
	git -C "${finish}" commit -qm 'readiness review of round 2'
	review2="$(git -C "${finish}" rev-parse HEAD)" ledger_earlier="${round1} ${round2}"
	printf 'Round 3.\n' >>"${finish}/README.md"
	git -C "${finish}" commit -qam 'round 3'
	head="$(git -C "${finish}" rev-parse HEAD)"

	# A warn of round 2 that the ledger of round 3 does not record fails finish.
	fct_readiness "${head}" Ready "${advise}${declined}" 'agent B'
	fct_finish 1
	fct_finish_said 'readiness.md: the finding R2-F1 of the review in'
	# A finding keeps its highest severity, so a warn never closes as a declined advise.
	fct_readiness "${head}" Ready "${advise}${declined}"$'\n\n'"${lowered}${declined}" 'agent B'
	fct_finish 1
	fct_finish_said 'the finding R2-F1 had the severity warn'
	fct_readiness "${head}" Ready "${advise}${declined}"$'\n\n'"${warn}${declined}" 'agent B'
	fct_finish 1
	fct_finish_said 'the finding R2-F1 has the severity warn, and only an advise can be declined'
	fct_readiness "${head}" Ready "${advise}${declined}"$'\n\n'"${warn}" 'agent B'
	fct_finish 1
	fct_finish_said 'the finding R2-F1 of round 2 has no state'
	# None. under ## Open lines does not close a finding that the ledger records as open.
	fct_readiness "${head}" 'Ready with stated limits' "${advise}${declined}"$'\n\n'"${warn}"$'\n- **State**: `open`\n- **Spec**: `long-lines`' 'agent B'
	fct_finish 1
	fct_finish_said 'the finding R2-F1 is open, even if ## Open lines holds None.'
	fct_readiness "${head}" Ready "${advise}${declined}"$'\n\n'"${warn}"$'\n- **State**: `fixed`\n- **Spec**: `long-lines`' 'agent B'
	fct_finish 1
	fct_finish_said 'this fixed finding names no commit of HEAD'
	fct_finish_said 'this fixed finding names no line of Failures or Acceptance in spec.md'
	valid="${advise}${declined}"$'\n\n'"${warn}"$'\n- **State**: `fixed`\n- **Commit**: '"${head}"$'\n- **Spec**: `bad-input`'
	fct_readiness "${head}" Ready "${valid}" 'agent B'
	git -C "${finish}" add readiness.md
	git -C "${finish}" commit -qm 'readiness review of round 3'
	fct_finish 0
	# The user accepts a limit with no new round: after the review, only the
	# rows under **Limits.** of spec.md change.
	fct_edit "${finish}/spec.md" '**Limits.** None.' $'**Limits.**\n\n| ID | State | Limit | Accepted |\n|---|---|---|---|\n| `long-lines` | `limit` | A line over 1 MB is cut | The user, 2026-10-31, https://example.com/pull/1 |'
	fct_readiness "${head}" 'Ready with stated limits' "${advise}${declined}"$'\n\n'"${warn}"$'\n- **State**: `limit`\n- **Spec**: `long-lines`' 'agent B'
	fct_finish 0
	fct_edit "${finish}/spec.md" '| `bad-input` | The file is missing or is no regular file, or `--timeout` is no duration | `USAGE_INVALID` |' '| `bad-input` | The file is missing | `USAGE_INVALID` |'
	fct_finish 1
	fct_finish_said 'and these files changed since: spec.md'
	git -C "${finish}" checkout -q spec.md readiness.md

	# Moving the intake after a review that needs work hides no finding and resets no round.
	ledger_earlier=""
	fct_readiness "${head}" Ready "${advise}${declined}" 'agent B'
	fct_edit "${finish}/readiness.md" "| Intake | ${intake} |" "| Intake | ${review1} |"
	fct_edit "${finish}/readiness.md" "| Spec | $(fct_step "${finish}" spec) |" "| Spec | ${round2} |"
	fct_edit "${finish}/readiness.md" "| Scenario | $(fct_step "${finish}" scenario) |" "| Scenario | ${review2} |"
	ledger_earlier="${round1} ${round2}"
	fct_finish 1
	fct_finish_said 'readiness.md: the finding R2-F1 of the review in'
	fct_finish_said 'readiness.md: round 1 of the ledger in'
	# A round never leaves the ledger, and no row hides behind a label.
	ledger_earlier="${round2}"
	fct_readiness "${head}" Ready "${valid}" 'agent B'
	ledger_earlier="${round1} ${round2}"
	fct_finish 1
	fct_finish_said 'readiness.md: round 1 of the ledger in'
	fct_finish_said "readiness.md: the review of ${round1:0:12}, committed in ${review1:0:12}, has no round under ## Rounds"
	# Each round before the last has a committed review of its head.
	ledger_earlier="${round1} ${round2} ${review2}"
	fct_readiness "${head}" Ready "${valid}" 'agent B'
	ledger_earlier="${round1} ${round2}"
	fct_finish 1
	fct_finish_said 'round 3 has no committed review of its head'
	fct_readiness "${head}" Ready "${valid}" 'agent B'
	fct_edit "${finish}/readiness.md" "| 3 | ${head} | agent A | ${first}..${head} | agent B | \`pending\` |" "| 3 | ${head} | agent A | ${first}..${head} | agent B | \`pending\` |"$'\n**Earlier build.**\n'"| 4 | ${head} | agent A | ${first}..${head} | agent B | \`pending\` |"
	fct_finish 1
	fct_finish_said 'this line of ## Rounds is no row of its table'
	# Each of the first four steps of round 1 has its own commit.
	fct_readiness "${head}" Ready "${valid}" 'agent B'
	fct_edit "${finish}/readiness.md" "| Spec | $(fct_step "${finish}" spec) |" "| Spec | ${intake} |"
	fct_finish 1
	fct_finish_said 'the step Spec is in the commit of the step before it'

	# A build has at most 5 rounds, and finish tells the agent to stop.
	ledger_earlier="${round1} ${round2} ${head} ${head} ${head}"
	fct_readiness "${head}" Ready "${valid}" 'agent B'
	ledger_earlier="${round1} ${round2}"
	fct_finish 1
	fct_finish_said 'readiness.md: build 1 of the ledger holds 6 rounds, and a build has at most 5'
	fct_finish_said 'this build used its 5 rounds. Stop, and report each open line and each open finding to the user'
	# A build that ended keeps its limit, and only a review that passes ends a build.
	ledger_earlier="${round1} ${round2} ${head} ${head} ${head} ${head}:pass"
	fct_readiness "${head}" Ready "${valid}" 'agent B'
	fct_finish 1
	fct_finish_said 'readiness.md: build 1 of the ledger holds 6 rounds, and a build has at most 5'
	fct_finish_said 'the intake of this build comes before the end of the build before it'
	ledger_earlier="${round1} ${round2}:pass"
	fct_readiness "${head}" Ready "${valid}" 'agent B'
	fct_finish 1
	fct_finish_said 'round 2 says pass, but its review in'
	ledger_earlier="${round1} ${round2}"
	fct_readiness "${head}" Ready "${valid}" 'agent B'
	stray="$(git -C "${finish}" commit-tree -m stray "${head}^{tree}")"
	fct_edit "${finish}/readiness.md" "| 3 | ${head} | agent A | ${first}..${head} | agent B | \`pending\` |" "| 3 | ${head} | agent A | ${stray}..${head} | agent B | \`pending\` |"
	fct_finish 1
	fct_finish_said "the commit \"${stray}\" is not an ancestor of HEAD"
	fct_readiness "${head}" Ready "${valid}" 'agent B'
	fct_edit "${finish}/readiness.md" "| 3 | ${head} | agent A | ${first}..${head} | agent B | \`pending\` |" "| 3 | ${round2} | agent A | ${first}..${head} | agent B | \`pending\` |"
	fct_finish 1
	fct_finish_said "the last round names the head ${round2}, and the line Commit: names ${head}"
	# Round 1 commits its first scenario before its handler.
	fct_readiness "${head}" Ready "${valid}" 'agent B'
	fct_edit "${finish}/readiness.md" "| Handler | ${round1} |" "| Handler | $(fct_step "${finish}" scenario) |"
	fct_finish 1
	fct_finish_said 'the handler and the first scenario are in one commit'
	printf 'Commit: %s\nReviewer: agent B\nVerdict: Ready\n\n## Findings\n\n%s\n' "${head}" "${valid}" >"${finish}/readiness.md"
	fct_finish 1
	fct_finish_said 'readiness.md: it holds no round under ## Rounds'

	fct_test_ledger_builds
}
