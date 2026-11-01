#!/usr/bin/env bash
# Sourced by template-test.sh to prove just finish, the exit of the loop of
# the nicely-extension skill (#405). It works on a copy of the rendered
# repository, so the other tests keep their state.
# Markdown backticks in the fixture text are literal.
# shellcheck disable=SC2016

# shellcheck source=scripts/template-finish-helpers.sh
source "${BASH_SOURCE[0]%/*}/template-finish-helpers.sh"
# shellcheck source=scripts/template-finish-forms-test.sh
source "${BASH_SOURCE[0]%/*}/template-finish-forms-test.sh"
# shellcheck source=scripts/template-finish-matrix-test.sh
source "${BASH_SOURCE[0]%/*}/template-finish-matrix-test.sh"

fct_test_finish() {
	local finish="${2}/finish" scratch="${2}" sha file number ledger_earlier=""
	local free='this limit has no spec ID, no state limit, and no acceptance'
	local unfilled='this line of the form is not filled'
	local warn=$'### R1-F1 Large files are slow\n\n- **Severity**: `warn`'
	local advise=$'### R1-F2 The count could name the longest line\n\n- **Severity**: `advise`'
	cp -R "${1}" "${finish}"
	fct_steps "${finish}"
	[[ "$(just --justfile "${finish}/justfile" --show finish | sed '/^#/d')" == $'finish:\n    go run github.com/pascalandy/nicely/sdk/extensionrepo/cmd/extensionrepo finish' ]] ||
		fct_fail 'just finish must call the SDK program in one line'

	# The state of trial 1: no open line, a limit that nobody accepted, and no readiness review.
	fct_edit "${finish}/spec.md" '**Limits.** None.' '**Limits.** A file over 1 GB takes more than 30 s to count.'
	git -C "${finish}" commit -qam 'state of trial 1'
	fct_finish 1
	fct_finish_said "${free}"
	fct_finish_said 'readiness.md: no readiness verdict is recorded'
	fct_finish_said 'this line is open' absent
	# A new repository keeps each Fill: line of its forms.
	for file in vision.md spec.md qa.md; do
		grep -n '^Fill:' "${finish}/${file}" >"${scratch}/fill.lines" || fct_fail "${file} of the template holds no Fill: line"
		while IFS=: read -r number _; do
			fct_finish_said "${file}:${number}: ${unfilled}"
		done <"${scratch}/fill.lines"
		grep -v '^Fill:' "${finish}/${file}" >"${scratch}/form.md"
		mv "${scratch}/form.md" "${finish}/${file}"
	done
	# An indented Fill: line is not filled either (#412).
	printf '  Fill: an indented line.\n' >>"${finish}/vision.md"
	fct_finish 1
	fct_finish_said "vision.md:$(grep -c '' "${finish}/vision.md"): ${unfilled}"
	fct_edit "${finish}/vision.md" '  Fill: an indented line.' ''

	fct_edit "${finish}/spec.md" 'None.' $'| ID | Command | Part | Case | Code |\n|---|---|---|---|---|\n| `large-file` | inspect | Failures | A file over 1 GB | `RUNTIME` |'
	fct_finish 1
	fct_finish_said 'this line is open'
	fct_finish_said "${unfilled}" absent

	fct_edit "${finish}/spec.md" '| `large-file` | inspect | Failures | A file over 1 GB | `RUNTIME` |' ''

	# The state of trial 2: the author of the code wrote the review, with no
	# line Reviewer:, and wrote the verdict Ready beside a warn.
	fct_edit "${finish}/spec.md" '**Limits.** A file over 1 GB takes more than 30 s to count.' '**Limits.** None.'
	git -C "${finish}" commit -qam 'state of trial 2'
	sha="$(git -C "${finish}" rev-parse HEAD)"
	fct_readiness "${sha}" Ready "${warn}" ''
	fct_finish 1
	fct_finish_said 'readiness.md: it names no reviewer'
	fct_finish_said 'readiness.md:8: the verdict is Ready, but this finding has the severity warn'
	# An advise of this round with no state is open, so the author declines it (#423).
	fct_readiness "${sha}" Ready "${advise}"$'\n- **State**: `declined`\n- **Reason**: the answer stays small'
	git -C "${finish}" add readiness.md
	git -C "${finish}" commit -qm 'independent readiness review'
	ledger_earlier="${sha}"
	fct_finish 0
	fct_finish_said 'finish: no line or finding is open'

	fct_edit "${finish}/spec.md" '**Limits.** None.' $'**Limits.**\n\n| ID | State | Limit | Accepted |\n|---|---|---|---|\n| `large-file` | `open` | A file over 1 GB takes more than 30 s to count | The user, 2026-10-30, https://example.com/pull/1 |'
	fct_finish 1
	fct_finish_said 'the limit `large-file` has the state "open"'
	fct_edit "${finish}/spec.md" '| `large-file` | `open` | A file over 1 GB takes more than 30 s to count | The user, 2026-10-30, https://example.com/pull/1 |' '| `large-file` | `limit` | A file over 1 GB takes more than 30 s to count | The user, 2026-10-30, https://example.com/pull/1 |'
	git -C "${finish}" commit -qam 'accepted limit'
	sha="$(git -C "${finish}" rev-parse HEAD)"
	number="$(grep -n '^| `large-file` | `limit` |' "${finish}/spec.md" | cut -d: -f1)"
	fct_readiness "${sha}" 'Needs work' 'None.'
	fct_finish 1
	fct_finish_said 'readiness.md: the verdict is Needs work'
	fct_finish_said "${free}" absent

	# The verdict follows from the findings and the limits.
	fct_readiness "${sha}" Ready 'None.'
	fct_finish 1
	fct_finish_said "spec.md:${number}: the verdict is Ready, but spec.md states the limit \`large-file\`"
	fct_readiness "${sha}" 'Ready with stated limits' "${warn}"$'\n\n### Output is lost\n\n- **Severity**: `block`'
	fct_finish 1
	fct_finish_said 'readiness.md:9: this warn names no limit that the user accepted'
	fct_finish_said 'readiness.md:13: the verdict is Ready with stated limits, but this finding has the severity block'
	fct_readiness "${sha}" 'Ready with stated limits' "${warn}"$'\n- **State**: `limit`\n- **Spec**: `small-file`'
	fct_finish 1
	fct_finish_said 'readiness.md:9: this warn names no limit that the user accepted'

	# Finish refuses a verdict or a commit that it does not know.
	fct_readiness "${sha}" 'Ready with limits' 'None.'
	fct_finish 1
	fct_finish_said 'readiness.md: the verdict "Ready with limits" is not Ready, Ready with stated limits, or Needs work'
	fct_readiness 0123456789abcdef0123456789abcdef01234567 'Ready with stated limits' 'None.'
	fct_finish 1
	fct_finish_said 'readiness.md: Git knows no commit 0123456789abcdef0123456789abcdef01234567'

	fct_test_finish_forms

	# Finish reads Commit:, Reviewer:, and Verdict: only in the first lines, never in fenced code (#413).
	fct_readiness "${sha}" 'Ready with stated limits' $'```\nReviewer: an agent that wrote none of the code\n```\n\nNone.' ''
	fct_finish 1
	fct_finish_said 'readiness.md: it names no reviewer'
	fct_readiness '' 'Ready with stated limits' $'```\nCommit: '"${sha}"$'\n```\n\nNone.'
	fct_finish 1
	fct_finish_said 'readiness.md: its first three lines hold no line Commit: <sha> or Verdict: <verdict>'

	fct_readiness "${sha}" 'Ready with stated limits' "${warn}"$'\n- **State**: `limit`\n- **Spec**: `large-file`\n\n'"${advise}"$'\n- **State**: `declined`\n- **Reason**: the answer stays small\n\n## Severity assessment\n\nEach finding names its severity.\n\n## Notes\n\n| R1-F2 | the answer stays small |\n\nLarge files block the terminal, and inspect may warn.'
	git -C "${finish}" add readiness.md
	git -C "${finish}" commit -qm 'readiness review'
	fct_finish 0
	fct_finish_said 'finish: no line or finding is open'
	fct_test_finish_matrix

	# A change after the review needs a new review.
	printf 'A change after the review.\n' >>"${finish}/README.md"
	fct_finish 1
	fct_finish_said "readiness.md: the review is of ${sha}, and these files changed since: README.md"
	git -C "${finish}" commit -qam 'a commit after the review'
	fct_finish 1
	fct_finish_said "readiness.md: the review is of ${sha}, and these files changed since: README.md"

	# A missing form fails finish, which still reports the rest (#412).
	git -C "${finish}" rm -q qa.md spec.md
	fct_finish 1
	fct_finish_said 'qa.md: the form is missing'
	fct_finish_said 'spec.md: the form is missing'
	fct_finish_said 'readiness.md:9: this warn names no limit that the user accepted'
	fct_finish_said 'no such file or directory' absent
}
