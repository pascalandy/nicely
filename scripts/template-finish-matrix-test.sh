#!/usr/bin/env bash
# Sourced by template-finish-test.sh. fct_test_finish calls
# fct_test_finish_matrix on a repository that passes just finish. It proves
# that finish fails when a family of ## Case matrix in spec.md has no test,
# when a refusal has no pair, and when a line of Failures or Acceptance is in
# no family (#423). Markdown backticks in the fixture text are literal.
# shellcheck disable=SC2016,SC2154

fct_test_finish_matrix() {
	local row='| The input holds text | Input: Of the wrong type | `lines` | `not-text` |'
	fct_edit "${finish}/spec.md" "${row}" '| The input holds text | Input: Of the wrong type | `lines` | None |'
	fct_finish 1
	fct_finish_said 'the line `not-text` is in no family of ## Case matrix'
	fct_edit "${finish}/spec.md" '| The input holds text | Input: Of the wrong type | `lines` | None |' '| The input holds text | Input: Of the wrong type | None | `not-text` |'
	fct_finish 1
	fct_finish_said 'the family "The input holds text" names no spec ID of an input that passes'
	fct_edit "${finish}/spec.md" '| The input holds text | Input: Of the wrong type | None | `not-text` |' '| The input holds text | Input: Of the wrong type | `lines`, `utf-16` | `not-text` |'
	fct_finish 1
	fct_finish_said 'the family "The input holds text" names `utf-16`, which is no line of Failures or Acceptance and no accepted limit'
	fct_edit "${finish}/spec.md" '| The input holds text | Input: Of the wrong type | `lines`, `utf-16` | `not-text` |' '| The input holds text | Input: Of the wrong type | `lines` |'
	fct_finish 1
	fct_finish_said 'this row of ## Case matrix has no cell for each column'
	git -C "${finish}" checkout -q spec.md
	awk '/^## Case matrix$/ { skip = 1; next } /^## / { skip = 0 } !skip' "${finish}/spec.md" >"${scratch}/spec.md"
	mv "${scratch}/spec.md" "${finish}/spec.md"
	fct_finish 1
	fct_finish_said 'spec.md: it holds no table under ## Case matrix'
	git -C "${finish}" checkout -q spec.md
	fct_finish 0
}
