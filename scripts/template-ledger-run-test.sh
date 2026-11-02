#!/usr/bin/env bash
# Sourced by template-ledger-test.sh. fct_test_ledger calls fct_test_ledger_run
# at the end of round 3, with its repository and its fixture variables. It
# proves the record of the first failing run of round 1, and that ## Open
# lines names each open finding (#420). Markdown backticks in the fixture
# text are literal, and fct_readiness reads ledger_earlier and ledger_run.
# shellcheck disable=SC2016,SC2034,SC2154

fct_test_ledger_run() {
	git -C "${finish}" checkout -q readiness.md spec.md
	# spec.md and the ledger agree on the open work.
	fct_readiness "${head}" 'Ready with stated limits' "${advise}${declined}"$'\n\n'"${warn}"$'\n- **State**: `open`\n- **Spec**: `long-lines`' 'agent B'
	fct_finish 1
	fct_finish_said 'spec.md: ## Open lines does not name the open finding R2-F1 of readiness.md'
	fct_edit "${finish}/spec.md" 'None.' $'| ID | Command | Part | Case | Code |\n|---|---|---|---|---|\n| `long-lines` | inspect | Failures | R2-F1: a line over 1 MB is cut | `RUNTIME` |'
	fct_finish 1
	fct_finish_said 'spec.md: ## Open lines does not name the open finding' absent
	git -C "${finish}" checkout -q spec.md

	# Round 4 stops at a limit with R4-F1 open and named under ## Open lines.
	# The user accepts it, so its row moves from ## Open lines to **Limits.**,
	# and the reviewer gives it the state limit with no new round.
	local before stop saved="${ledger_earlier}" r4=$'### R4-F1 Long lines are cut\n\n- **Severity**: `warn`'
	before="$(git -C "${finish}" rev-parse HEAD)" ledger_earlier="${round1} ${round2} ${head}"
	git -C "${finish}" checkout -q readiness.md
	fct_edit "${finish}/spec.md" 'None.' $'| ID | Command | Part | Case | Code |\n|---|---|---|---|---|\n| `long-lines` | inspect | Failures | R4-F1: a line over 1 MB is cut | `RUNTIME` |'
	git -C "${finish}" commit -qm 'round 4 stops with an open finding' spec.md
	stop="$(git -C "${finish}" rev-parse HEAD)"
	fct_readiness "${stop}" 'Needs work' "${valid}"$'\n\n'"${r4}"$'\n- **State**: `open`\n- **Spec**: `long-lines`' 'agent B'
	git -C "${finish}" add readiness.md
	git -C "${finish}" commit -qm 'readiness review of round 4'
	fct_finish 1
	fct_finish_said 'the finding R4-F1 is open'
	fct_edit "${finish}/spec.md" '| `long-lines` | inspect | Failures | R4-F1: a line over 1 MB is cut | `RUNTIME` |' ''
	fct_edit "${finish}/spec.md" '**Limits.** None.' $'**Limits.**\n\n| ID | State | Limit | Accepted |\n|---|---|---|---|\n| `long-lines` | `limit` | A line over 1 MB is cut | The user, 2026-10-31, https://example.com/pull/1 |'
	git -C "${finish}" commit -qm 'the user accepts the limit long-lines' spec.md
	fct_readiness "${stop}" 'Ready with stated limits' "${valid}"$'\n\n'"${r4}"$'\n- **State**: `limit`\n- **Spec**: `long-lines`' 'agent B'
	fct_finish 0
	# Any other change of spec.md still needs a review of its commit.
	fct_edit "${finish}/spec.md" '| `bad-input` | The file is missing or is no regular file, or `--timeout` is no duration | `USAGE_INVALID` |' '| `bad-input` | The file is missing | `USAGE_INVALID` |'
	fct_finish 1
	fct_finish_said 'and these files changed since: spec.md'
	git -C "${finish}" reset -q --hard "${before}"
	ledger_earlier="${saved}"

	# Round 1 records the run of its first scenario: the command, a failing
	# exit status, and the assertion that failed (#420).
	ledger_run='' fct_readiness "${head}" Ready "${valid}" 'agent B'
	fct_finish 1
	fct_finish_said 'readiness.md: it holds no part ## First failing run'
	ledger_run=$'- **Command**: `go test ./...`\n- **Exit**: 0\n- **Assertion**:' fct_readiness "${head}" Ready "${valid}" 'agent B'
	fct_finish 1
	fct_finish_said 'the first run of the scenario exited 0'
	fct_finish_said 'names no assertion that failed'
	ledger_run=$'- **Exit**: one' fct_readiness "${head}" Ready "${valid}" 'agent B'
	fct_finish 1
	fct_finish_said 'names no command'
	fct_finish_said 'the exit status "one" of the first run'
	# A record of another commit, such as of an earlier build, or in fenced code, records no run of this build.
	ledger_run=$'- **Commit**: '"${intake}"$'\n- **Command**: `go test ./...`\n- **Exit**: 1\n- **Assertion**: `stdout 3`' fct_readiness "${head}" Ready "${valid}" 'agent B'
	fct_finish 1
	fct_finish_said "names the commit \"${intake}\", which is not the Scenario commit of ## Order of round 1"
	ledger_run=$'```\n- **Commit**: '"$(fct_step "${finish}" scenario)"$'\n- **Command**: `go test ./...`\n- **Exit**: 1\n- **Assertion**: `stdout 3`\n```' fct_readiness "${head}" Ready "${valid}" 'agent B'
	fct_finish 1
	fct_finish_said 'names no command'
	# A longer fence holds a shorter one, and a value can end with a period.
	ledger_run=$'````\n```\n````\n\n- **Commit**: '"$(fct_step "${finish}" scenario)"$'.\n- **Command:** `go test ./...`\n- **Exit**: 1.\n- **Assertion**: `stdout 3` found no match.' fct_readiness "${head}" Ready "${valid}" 'agent B'
	fct_finish 0
	git -C "${finish}" checkout -q readiness.md
}
