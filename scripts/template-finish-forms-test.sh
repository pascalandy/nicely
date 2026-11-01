#!/usr/bin/env bash
# Sourced by template-finish-test.sh. fct_test_finish calls
# fct_test_finish_forms with its repository and its fixture variables. It
# proves that just finish fails on each finding that it cannot grade, and on
# each form that grades outside the line - **Severity**: of a finding.
# Markdown backticks in the fixture text are literal.
# shellcheck disable=SC2016,SC2154

fct_test_finish_forms() {
	local number
	local hidden=$'### R1-F2 The count could name the longest line\n\n- **Severity**: `advise`\n\n| Output keeps the author | `warn` |\n| Output keeps the serial number | block |\n\n```\n### Output keeps the place\n\n- **Severity**: `warn`\n```\n\n### R1-F3 Output keeps the name\n\n- Severity: warn\n- **Severity**: `advise`\n\n### R1-F4 Output keeps the date\n\nSeverity: warn.\n\n- **Severity**: `advise`\n\n### Severity: warn\n\n- **Severity**: `advise`\n- **Spec**: `block`'
	local unread=$'- **Severity**: `warn`\n\n### No severity\n\n### Two severities\n\n- **Severity**: `advise`\n- **Severity**: `warn`\n\n### Unknown severity\n\n- **Severity**: minor\n\n### A fenced heading\n\n````\n# a comment\n```\n````\n\n- **Severity**: `block`\n\n### An open fence\n\n```\n- **Severity**: `advise`'
	# A finding that finish cannot grade fails it.
	fct_readiness "${sha}" 'Ready with stated limits' ''
	fct_finish 1
	fct_finish_said 'readiness.md: it holds no part ## Findings'
	fct_readiness "${sha}" 'Ready with stated limits' $'## Verdict\n\nReady with stated limits.'
	fct_finish 1
	fct_finish_said 'readiness.md:7: ## Findings holds no ### finding and no None.'
	fct_readiness "${sha}" 'Ready with stated limits' $'| Finding | Severity |\n|---|---|\n| Output is lost | `block` |\n\n'"${advise}"
	fct_finish 1
	fct_finish_said 'readiness.md:9: this text of ## Findings is outside every ### finding'
	fct_readiness "${sha}" 'Ready with stated limits' "${unread}"
	fct_finish 1
	fct_finish_said 'readiness.md:9: this severity has no finding of its own'
	fct_finish_said 'readiness.md:11: this finding has no severity'
	fct_finish_said 'readiness.md:16: this finding already has a severity'
	fct_finish_said 'readiness.md:20: the severity "minor" is not block, warn, or advise'
	fct_finish_said 'readiness.md:22: the verdict is Ready with stated limits, but this finding has the severity block'
	fct_finish_said 'readiness.md:33: this fenced code has no end'

	# A severity outside the line - **Severity**: of a finding fails it, even after a valid finding (#413).
	fct_readiness "${sha}" 'Ready with stated limits' "${hidden}"
	fct_finish 1
	fct_finish_said 'readiness.md:13: this table names the severity warn'
	fct_finish_said 'readiness.md:14: this table names the severity block'
	fct_finish_said 'readiness.md:19: this fenced code names the severity `warn`'
	for number in 24 29 33; do
		fct_finish_said "readiness.md:${number}: this line grades with the severity warn"
	done
	fct_finish_said 'readiness.md:36: this line names the severity `block` in code'

	# A grade in a free form, such as in notes, fails too (#422).
	fct_readiness "${sha}" 'Ready with stated limits' "${warn}"$'\n- **State**: `limit`\n- **Spec**: `large-file`\n\n'"${advise}"$'\n\n## Notes\n\n| R1-F1 | really a warn |\n\nGrade: block. The name in the output is a block for me.\n\nThis issue is warn-level and should block the release.'
	fct_finish 1
	fct_finish_said 'readiness.md:21: this table names the severity warn'
	fct_finish_said 'readiness.md:23: this line grades with the severity block outside the line - **Severity**: of a finding. Fix: grade each finding only in its line - **Severity**:, and rephrase this line without the word grade'
	fct_finish_said 'readiness.md:25: this line grades with the severity warn outside the line - **Severity**: of a finding. Fix: grade each finding only in its line - **Severity**:, and rephrase this line without the form warn-level'
}
