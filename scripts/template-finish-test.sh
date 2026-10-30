#!/usr/bin/env bash
# Sourced by template-test.sh to prove just finish, the exit of the loop of
# the nicely-extension skill (#405). It works on a copy of the rendered
# repository, so the other tests keep their state.
# Markdown backticks in the fixture text are literal.
# shellcheck disable=SC2016,SC2154 # fct_test_finish sets finish and scratch for its helpers.

# fct_finish runs just finish in the copy, and requires the exit code $1.
fct_finish() {
	local status=0
	(cd "${finish}" && just finish) >"${scratch}/finish.out" 2>&1 || status=$?
	if [[ "${status}" != "${1}" ]]; then
		cat "${scratch}/finish.out" >&2
		fct_fail "just finish exited ${status}, want ${1}"
	fi
}

# fct_finish_said requires the text $1 in the output of the last just finish,
# or its absence when $2 is "absent".
fct_finish_said() {
	local found=0
	grep -qF -- "${1}" "${scratch}/finish.out" || found=$?
	if [[ "${2:-}" == absent && "${found}" == 0 || "${2:-}" != absent && "${found}" != 0 ]]; then
		cat "${scratch}/finish.out" >&2
		fct_fail "just finish: want ${2:-present} \"${1}\""
	fi
}

# fct_readiness writes readiness.md for the commit $1, with the verdict $2, the
# findings $3 under ## Findings, and the reviewer $4. An empty $3 leaves out
# ## Findings, and an empty $4 leaves out the line Reviewer:.
fct_readiness() {
	local reviewer="${4-an agent that wrote none of the code}"
	{
		printf 'Commit: %s\n' "${1}"
		[[ -z "${reviewer}" ]] || printf 'Reviewer: %s\n' "${reviewer}"
		printf 'Verdict: %s\n\n# Readiness review\n' "${2}"
		[[ -z "${3}" ]] || printf '\n## Findings\n\n%s\n' "${3}"
	} >"${finish}/readiness.md"
}

fct_test_finish() {
	local finish="${2}/finish" scratch="${2}" sha file number
	local free='this limit has no spec ID, no state limit, and no acceptance'
	local unfilled='this line of the form is not filled'
	local warn=$'### Large files are slow\n\n- **Severity**: `warn`'
	local advise=$'### The count could name the longest line\n\n- **Severity**: `advise`'
	local unread=$'- **Severity**: `warn`\n\n### No severity\n\n### Two severities\n\n- **Severity**: `advise`\n- **Severity**: `warn`\n\n### Unknown severity\n\n- **Severity**: minor\n\n### A fenced heading\n\n````\n# a comment\n```\n````\n\n- **Severity**: `block`\n\n### An open fence\n\n```\n- **Severity**: `advise`'
	cp -R "${1}" "${finish}"
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
	fct_readiness "${sha}" Ready "${advise}"
	git -C "${finish}" add readiness.md
	git -C "${finish}" commit -qm 'independent readiness review'
	fct_finish 0
	fct_finish_said 'finish: no line is open'

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
	fct_readiness "${sha}" 'Ready with stated limits' "${warn}"$'\n- **Limit**: `small-file`'
	fct_finish 1
	fct_finish_said 'readiness.md:9: this warn names no limit that the user accepted'

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

	fct_readiness "${sha}" 'Ready with stated limits' "${warn}"$'\n- **Limit**: `large-file`\n\n'"${advise}"
	git -C "${finish}" add readiness.md
	git -C "${finish}" commit -qm 'readiness review'
	fct_finish 0
	fct_finish_said 'finish: no line is open'

	# A change after the review needs a new review.
	printf 'A change after the review.\n' >>"${finish}/README.md"
	fct_finish 1
	fct_finish_said "readiness.md: the review is of ${sha}, and these files changed since: README.md"
}
