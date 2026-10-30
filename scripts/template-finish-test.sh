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

fct_test_finish() {
	local finish="${2}/finish" scratch="${2}" sha
	local free='this limit has no spec ID, no state limit, and no acceptance'
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

	fct_edit "${finish}/spec.md" 'None.' $'| ID | Command | Part | Case | Code |\n|---|---|---|---|---|\n| `large-file` | inspect | Failures | A file over 1 GB | `RUNTIME` |'
	fct_finish 1
	fct_finish_said 'this line is open'

	fct_edit "${finish}/spec.md" '| `large-file` | inspect | Failures | A file over 1 GB | `RUNTIME` |' ''
	fct_edit "${finish}/spec.md" '**Limits.** A file over 1 GB takes more than 30 s to count.' $'**Limits.**\n\n| ID | State | Limit | Accepted |\n|---|---|---|---|\n| `large-file` | `limit` | A file over 1 GB takes more than 30 s to count | The user, 2026-10-30, https://example.com/pull/1 |'
	fct_edit "${finish}/spec.md" '| `large-file` | `limit` | A file over 1 GB takes more than 30 s to count | The user, 2026-10-30, https://example.com/pull/1 |' '| `large-file` | `open` | A file over 1 GB takes more than 30 s to count | The user, 2026-10-30, https://example.com/pull/1 |'
	fct_finish 1
	fct_finish_said 'the limit `large-file` has the state "open"'
	fct_edit "${finish}/spec.md" '| `large-file` | `open` | A file over 1 GB takes more than 30 s to count | The user, 2026-10-30, https://example.com/pull/1 |' '| `large-file` | `limit` | A file over 1 GB takes more than 30 s to count | The user, 2026-10-30, https://example.com/pull/1 |'
	git -C "${finish}" commit -qam 'accepted limit'
	sha="$(git -C "${finish}" rev-parse HEAD)"
	printf '# Readiness review\n\nCommit: %s\nVerdict: Needs work\n' "${sha}" >"${finish}/readiness.md"
	fct_finish 1
	fct_finish_said 'readiness.md: the verdict is Needs work'
	fct_finish_said "${free}" absent

	printf '# Readiness review\n\nCommit: %s\nVerdict: Ready with stated limits\n' "${sha}" >"${finish}/readiness.md"
	git -C "${finish}" add readiness.md
	git -C "${finish}" commit -qm 'readiness review'
	fct_finish 0
	fct_finish_said 'finish: no line is open'

	# A change after the review needs a new review.
	printf 'A change after the review.\n' >>"${finish}/README.md"
	fct_finish 1
	fct_finish_said "readiness.md: the review is of ${sha}, and these files changed since: README.md"
}
