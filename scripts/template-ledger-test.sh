#!/usr/bin/env bash
# Sourced by template-test.sh to prove the ledger of review.toml, which
# just finish reads (#417, COR-104-T4). It works on a copy of the rendered
# repository, so the other tests keep their state.
# Markdown backticks in the fixture text are literal, and fct_readiness reads
# ledger_earlier.
# shellcheck disable=SC2016,SC2034

# shellcheck source=scripts/template-ledger-run-test.sh
source "${BASH_SOURCE[0]%/*}/template-ledger-run-test.sh"

fct_test_ledger() {
	local finish="${2}/ledger" scratch="${2}" first intake round1 review1 round2 review2 head stray valid records_status=0 ledger_earlier=""
	local declined='reason = "the answer stays small"' advise warn
	advise="$(fct_finding R1-F1 advise declined lines "${declined}")"
	warn="$(fct_finding R2-F1 warn open long-lines)"
	cp -R "${1}" "${finish}"
	fct_steps "${finish}"
	intake="$(fct_step "${finish}" intake)" first="$(git -C "${finish}" rev-list --max-parents=0 HEAD)"
	fct_fill "${finish}" "${scratch}"
	git -C "${finish}" commit -qam 'fill the forms'
	round1="$(git -C "${finish}" rev-parse HEAD)"

	# Round 1 reports an advise, and round 2 declines it and reports a warn.
	# Their committed reviews are the history of round 3.
	fct_readiness "${round1}" 'Needs work' "$(fct_finding R1-F1 advise open lines)"
	fct_review "${finish}" 'readiness review of round 1'
	review1="$(git -C "${finish}" rev-parse HEAD)" ledger_earlier="${round1}"
	printf 'Round 2.\n' >>"${finish}/README.md"
	git -C "${finish}" commit -qam 'round 2'
	round2="$(git -C "${finish}" rev-parse HEAD)"
	fct_readiness "${round2}" 'Needs work' "${advise}"$'\n\n'"${warn}" 'agent B'
	fct_review "${finish}" 'readiness review of round 2'
	review2="$(git -C "${finish}" rev-parse HEAD)" ledger_earlier="${round1} ${round2}"
	printf 'Round 3.\n' >>"${finish}/README.md"
	git -C "${finish}" commit -qam 'round 3'
	head="$(git -C "${finish}" rev-parse HEAD)"

	# A warn of round 2 that the ledger of round 3 does not record fails finish.
	fct_readiness "${head}" Ready "${advise}" 'agent B'
	fct_finish 1
	fct_finish_said 'review.toml: the finding R2-F1 of the review in'
	# A finding keeps its highest severity, so a warn never closes as a declined advise.
	fct_readiness "${head}" Ready "${advise}"$'\n\n'"$(fct_finding R2-F1 advise declined long-lines "${declined}")" 'agent B'
	fct_finish 1
	fct_finish_said 'review.toml: the finding R2-F1 had the severity warn'
	fct_readiness "${head}" Ready "${advise}"$'\n\n'"$(fct_finding R2-F1 warn declined long-lines "${declined}")" 'agent B'
	fct_finish 1
	fct_finish_said 'the finding R2-F1 has the severity warn, and only an advise can be declined'
	# None. under ## Open lines does not close a finding that the ledger records as open.
	fct_readiness "${head}" 'Ready with stated limits' "${advise}"$'\n\n'"${warn}" 'agent B'
	fct_finish 1
	fct_finish_said 'review.toml: the finding R2-F1 is open, even if ## Open lines holds None.'
	fct_readiness "${head}" Ready "${advise}"$'\n\n'"$(fct_finding R2-F1 warn fixed long-lines)" 'agent B'
	fct_finish 1
	fct_finish_said 'review.toml: the fixed finding R2-F1 names no commit of HEAD'
	fct_finish_said 'review.toml: the fixed finding R2-F1 names no line of Failures or Acceptance in spec.md'
	valid="${advise}"$'\n\n'"$(fct_finding R2-F1 warn fixed bad-input "commit = \"${head}\"")"
	fct_readiness "${head}" Ready "${valid}" 'agent B'
	fct_review "${finish}" 'readiness review of round 3'
	fct_finish 0
	# The user accepts a limit with no new round: after the review, only the
	# rows under **Limits.** of spec.md change.
	fct_edit "${finish}/spec.md" '**Limits.** None.' $'**Limits.**\n\n| ID | State | Limit | Accepted |\n|---|---|---|---|\n| `long-lines` | `limit` | A line over 1 MB is cut | The user, 2026-10-31, https://example.com/pull/1 |'
	fct_readiness "${head}" 'Ready with stated limits' "${advise}"$'\n\n'"$(fct_finding R2-F1 warn limit long-lines)" 'agent B'
	fct_finish 0
	fct_edit "${finish}/spec.md" '| `bad-input` | The file is missing or is no regular file, or `--timeout` is no duration | `USAGE_INVALID` |' '| `bad-input` | The file is missing | `USAGE_INVALID` |'
	fct_finish 1
	fct_finish_said 'and these files changed since: spec.md'
	git -C "${finish}" checkout -q spec.md review.toml readiness.md

	# Moving the intake after a review that needs work hides no finding and resets no round.
	ledger_earlier=""
	fct_readiness "${head}" Ready "${advise}" 'agent B'
	fct_edit "${finish}/review.toml" "intake = \"${intake}\"" "intake = \"${review1}\""
	fct_edit "${finish}/review.toml" "spec = \"$(fct_step "${finish}" spec)\"" "spec = \"${round2}\""
	fct_edit "${finish}/review.toml" "scenario = \"$(fct_step "${finish}" scenario)\"" "scenario = \"${review2}\""
	fct_records
	fct_finish 1
	fct_finish_said 'review.toml: the finding R2-F1 of the review in'
	fct_finish_said 'review.toml: round 1 of the ledger in'
	# A round never leaves the ledger.
	ledger_earlier="${round2}"
	fct_readiness "${head}" Ready "${valid}" 'agent B'
	ledger_earlier="${round1} ${round2}"
	fct_finish 1
	fct_finish_said 'review.toml: round 1 of the ledger in'
	fct_finish_said "review.toml: the review of ${round1:0:12}, committed in ${review1:0:12}, has no round"
	# Each round before the last has a committed review of its head.
	ledger_earlier="${round1} ${round2} ${review2}"
	fct_readiness "${head}" Ready "${valid}" 'agent B'
	ledger_earlier="${round1} ${round2}"
	fct_finish 1
	fct_finish_said 'review.toml: round 3 has no committed review of its head'
	# Each of the first four steps of round 1 has its own commit.
	fct_readiness "${head}" Ready "${valid}" 'agent B'
	fct_edit "${finish}/review.toml" "spec = \"$(fct_step "${finish}" spec)\"" "spec = \"${intake}\""
	fct_records
	fct_finish 1
	fct_finish_said 'the step spec is in the commit of the step before it'

	# A build has at most 5 rounds, and finish tells the agent to stop.
	ledger_earlier="${round1} ${round2} ${head} ${head} ${head}"
	fct_readiness "${head}" Ready "${valid}" 'agent B'
	ledger_earlier="${round1} ${round2}"
	fct_finish 1
	fct_finish_said 'review.toml: build 1 of the ledger holds 6 rounds, and a build has at most 5'
	fct_finish_said 'this build used its 5 rounds. Stop, and report each open line and each open finding to the user'
	# A build that ended keeps its limit, and only a review that passes ends a build.
	ledger_earlier="${round1} ${round2} ${head} ${head} ${head} ${head}:pass"
	fct_readiness "${head}" Ready "${valid}" 'agent B'
	fct_finish 1
	fct_finish_said 'review.toml: build 1 of the ledger holds 6 rounds, and a build has at most 5'
	fct_finish_said 'the intake of this build comes before the end of the build before it'
	ledger_earlier="${round1} ${round2}:pass"
	fct_readiness "${head}" Ready "${valid}" 'agent B'
	fct_finish 1
	fct_finish_said 'review.toml: round 2 says pass, but just finish fails on its review in'
	ledger_earlier="${round1} ${round2}"
	fct_readiness "${head}" Ready "${valid}" 'agent B'
	stray="$(git -C "${finish}" commit-tree -m stray "${head}^{tree}")"
	fct_edit "${finish}/review.toml" "change_commits = \"${first}..${head}\"" "change_commits = \"${stray}..${head}\""
	fct_records
	fct_finish 1
	fct_finish_said "the commit \"${stray}\" is not an ancestor of HEAD"
	fct_readiness "${head}" Ready "${valid}" 'agent B'
	fct_edit "${finish}/review.toml" "head = \"${head}\"" "head = \"${round2}\""
	fct_records
	fct_finish 1
	fct_finish_said "review.toml: the last round names the head ${round2}, and the key commit names ${head}"
	# Round 1 commits its first scenario before its handler.
	fct_readiness "${head}" Ready "${valid}" 'agent B'
	fct_edit "${finish}/review.toml" "handler = \"${round1}\"" "handler = \"$(fct_step "${finish}" scenario)\""
	fct_records
	fct_finish 1
	fct_finish_said 'the handler and the first scenario are in one commit'
	printf 'commit = "%s"\nreviewer = "agent B"\nverdict = "Ready"\n\n%s\n' "${head}" "${valid}" >"${finish}/review.toml"
	fct_records
	fct_finish 1
	fct_finish_said 'review.toml: it holds no round'

	fct_test_ledger_run
	fct_test_ledger_builds
}
