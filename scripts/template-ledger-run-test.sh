#!/usr/bin/env bash
# Sourced by template-ledger-test.sh. fct_test_ledger calls fct_test_ledger_run
# at the end of round 3, with its repository and its fixture variables. It
# proves the record of the first failing run of round 1, and that ## Open
# lines holds the line of each open finding (#420, #422, #423). Markdown
# backticks in the fixture text are literal, and fct_readiness reads
# ledger_earlier and ledger_run.
# shellcheck disable=SC2016,SC2034,SC2154

fct_test_ledger_run() {
	local scenario run
	git -C "${finish}" checkout -q review.toml readiness.md spec.md
	scenario="$(fct_step "${finish}" scenario)"
	# spec.md and the ledger agree on the open work.
	fct_readiness "${head}" 'Ready with stated limits' "${advise}"$'\n\n'"${warn}" 'agent B'
	fct_finish 1
	fct_finish_said 'spec.md: ## Open lines holds no row `long-lines`, which the open finding R2-F1 names'
	fct_edit "${finish}/spec.md" 'None.' $'| ID | Command | Part | Case | Code |\n|---|---|---|---|---|\n| `long-lines` | inspect | Failures | A line over 1 MB is cut | `RUNTIME` |'
	fct_finish 1
	fct_finish_said 'spec.md: ## Open lines holds no row' absent
	fct_finish_said 'review.toml: the finding R2-F1 is open'
	git -C "${finish}" checkout -q spec.md

	# Round 4 stops at a limit with R4-F1 open and its row under ## Open lines.
	# The user accepts it, so its row moves from ## Open lines to **Limits.**,
	# and the reviewer gives it the state limit with no new round.
	local before stop saved="${ledger_earlier}"
	before="$(git -C "${finish}" rev-parse HEAD)" ledger_earlier="${round1} ${round2} ${head}"
	git -C "${finish}" checkout -q review.toml readiness.md
	fct_edit "${finish}/spec.md" 'None.' $'| ID | Command | Part | Case | Code |\n|---|---|---|---|---|\n| `long-lines` | inspect | Failures | A line over 1 MB is cut | `RUNTIME` |'
	git -C "${finish}" commit -qm 'round 4 stops with an open finding' spec.md
	stop="$(git -C "${finish}" rev-parse HEAD)"
	fct_readiness "${stop}" 'Needs work' "${valid}"$'\n\n'"$(fct_finding R4-F1 warn open long-lines)" 'agent B'
	fct_review "${finish}" 'readiness review of round 4'
	fct_finish 1
	fct_finish_said 'the finding R4-F1 is open'
	fct_edit "${finish}/spec.md" '| `long-lines` | inspect | Failures | A line over 1 MB is cut | `RUNTIME` |' ''
	fct_edit "${finish}/spec.md" '**Limits.** None.' $'**Limits.**\n\n| ID | State | Limit | Accepted |\n|---|---|---|---|\n| `long-lines` | `limit` | A line over 1 MB is cut | The user, 2026-10-31, https://example.com/pull/1 |'
	git -C "${finish}" commit -qm 'the user accepts the limit long-lines' spec.md
	fct_readiness "${stop}" 'Ready with stated limits' "${valid}"$'\n\n'"$(fct_finding R4-F1 warn limit long-lines)" 'agent B'
	fct_finish 0
	# Any other change of spec.md still needs a review of its commit.
	fct_edit "${finish}/spec.md" '| `bad-input` | The file is missing or is no regular file, or `--timeout` is no duration | `USAGE_INVALID` |' '| `bad-input` | The file is missing | `USAGE_INVALID` |'
	fct_finish 1
	fct_finish_said 'and these files changed since: spec.md'
	git -C "${finish}" reset -q --hard "${before}"
	ledger_earlier="${saved}"

	# Round 1 records the run of its first scenario: the scenario, a failing
	# exit status, and the assertion that failed (#420).
	ledger_run='' fct_readiness "${head}" Ready "${valid}" 'agent B'
	fct_finish 1
	fct_finish_said 'review.toml: it records no first failing run'
	ledger_run=$'[first_run]\nscenario = "inspect"\nexit = 0' fct_readiness "${head}" Ready "${valid}" 'agent B'
	fct_finish 1
	fct_finish_said 'the first run of the scenario exited 0'
	fct_finish_said 'names no assertion that failed'
	ledger_run=$'[first_run]\nexit = "one"' fct_readiness "${head}" Ready "${valid}" 'agent B'
	fct_finish 1
	fct_finish_said 'review.toml: it is no valid record'
	# A record of another commit, such as of an earlier build, records no run of this build.
	run=$'\nscenario = "inspect"\nexit = 1\nassertion = \'\'\'stdout \'"lines":3\' found no match\'\'\''
	ledger_run=$'[first_run]\ncommit = "'"${intake}"'"'"${run}" fct_readiness "${head}" Ready "${valid}" 'agent B'
	fct_finish 1
	fct_finish_said "review.toml: the first run names the commit \"${intake}\", which is not the Scenario commit of [order]"
	ledger_run=$'[first_run]\ncommit = "'"${scenario}"'"'"${run}" fct_readiness "${head}" Ready "${valid}" 'agent B'
	fct_finish 0
	# The assertion copies a line of the scenario, so a severity in it grades nothing (#422).
	ledger_run=$'[first_run]\ncommit = "'"${scenario}"$'"\nscenario = "inspect"\nexit = 1\nassertion = \'\'\'stdout \'"lines":3\' found no match, not a warn\'\'\'' fct_readiness "${head}" Ready "${valid}" 'agent B'
	fct_finish 0

	# The scenario is the name of a scenario of the Scenario commit, and the
	# assertion holds a line of that scenario. So signoff can run it again (#422).
	ledger_run=$'[first_run]\ncommit = "'"${scenario}"$'"\nscenario = "inspect; echo invented"\nexit = 1\nassertion = "stdout 3"' fct_readiness "${head}" Ready "${valid}" 'agent B'
	fct_finish 1
	fct_finish_said 'review.toml: the first run names the scenario "inspect; echo invented", which is no name of a scenario'
	ledger_run=$'[first_run]\ncommit = "'"${scenario}"$'"\nscenario = "missing"\nexit = 1\nassertion = "stdout 3"' fct_readiness "${head}" Ready "${valid}" 'agent B'
	fct_finish 1
	fct_finish_said 'review.toml: the first run runs the scenario testdata/script/missing.txtar, which the Scenario commit'
	ledger_run=$'[first_run]\ncommit = "'"${scenario}"$'"\nscenario = "inspect"\nexit = 1\nassertion = "the count was wrong"' fct_readiness "${head}" Ready "${valid}" 'agent B'
	fct_finish 1
	fct_finish_said 'review.toml: the first run names the assertion "the count was wrong", which holds no line of testdata/script/inspect.txtar'
	git -C "${finish}" checkout -q review.toml readiness.md
}
