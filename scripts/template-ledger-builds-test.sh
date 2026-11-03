#!/usr/bin/env bash
# Sourced by template-test.sh. fct_test_ledger calls fct_test_ledger_builds
# at the end of round 3, with its repository and its fixture variables. It
# proves the start of each build (#417), also after a Markdown readiness.md
# of a template before typed records (COR-104-T4).
# Markdown backticks in the fixture text are literal, and fct_readiness reads
# ledger_earlier.
# shellcheck disable=SC2016,SC2034,SC2154

fct_test_ledger_builds() {
	local base passed forged kept head2 head3 i limit=$'**Limits.**\n\n| ID | State | Limit | Accepted |\n|---|---|---|---|\n| `long-lines` | `limit` | A line over 1 MB is cut | The user, 2026-10-31, https://example.com/pull/1 |'
	local blank
	blank="$(fct_finding R5-F1 block open bad-input)"
	git -C "${finish}" checkout -q review.toml readiness.md spec.md
	base="$(git -C "${finish}" rev-parse HEAD)"
	# After a limit that the user accepted, a round with the finish pass ends
	# the build, and the next build starts.
	kept="${advise}"$'\n\n'"$(fct_finding R2-F1 warn limit long-lines)"
	fct_readiness "${head}" 'Ready with stated limits' "${kept}" 'agent B'
	fct_review "${finish}" 'a review with an accepted limit'
	fct_edit "${finish}/spec.md" '**Limits.** None.' "${limit}"
	git -C "${finish}" commit -qam 'the user accepts the limit long-lines'
	fct_steps "${finish}"
	printf 'Handler of the build after the limit.\n' >>"${finish}/README.md"
	git -C "${finish}" commit -qam 'handler of the build after the limit'
	ledger_earlier="${round1} ${round2} ${head}:pass"
	fct_readiness "$(git -C "${finish}" rev-parse HEAD)" 'Ready with stated limits' "${kept}" 'agent B'
	fct_finish 0
	ledger_earlier="${round1} ${round2}"
	git -C "${finish}" reset -q --hard "${base}"

	# A new build, such as of another command, starts at its own intake after
	# a round with the finish pass. The ledger keeps the findings of each build.
	fct_steps "${finish}"
	printf 'Handler of the next build.\n' >>"${finish}/README.md"
	git -C "${finish}" commit -qam 'handler of the next build'
	head2="$(git -C "${finish}" rev-parse HEAD)" ledger_earlier="${round1} ${round2} ${head}:pass"
	fct_readiness "${head2}" Ready "${valid}" 'agent B'
	fct_finish 0
	fct_review "${finish}" 'readiness review of the next build'
	base="$(git -C "${finish}" rev-parse HEAD)" passed="${base}"
	# Ten builds that each end with a review that passes still pass.
	for i in 3 4 5 6 7 8 9 10; do
		ledger_earlier="${ledger_earlier} $(git -C "${finish}" log --format=%H -n 1 HEAD~1):pass"
		fct_steps "${finish}"
		printf 'Handler of build %s.\n' "${i}" >>"${finish}/README.md"
		git -C "${finish}" commit -qam "handler of build ${i}"
		fct_readiness "$(git -C "${finish}" rev-parse HEAD)" Ready "${valid}" 'agent B'
		fct_review "${finish}" "readiness review of build ${i}"
	done
	fct_finish 0
	git -C "${finish}" reset -q --hard "${base}"

	# Only a review that passes ends a build: a verdict that says Ready, a
	# state with no commit, a review that drops a finding, or a stale view ends none.
	fct_steps "${finish}"
	printf 'Handler of a build that needs work.\n' >>"${finish}/README.md"
	git -C "${finish}" commit -qam 'handler of a build that needs work'
	head3="$(git -C "${finish}" rev-parse HEAD)" ledger_earlier="${round1} ${round2} ${head}:pass ${head2}:pass"
	fct_readiness "${head3}" 'Needs work' "${valid}"$'\n\n'"${blank}" 'agent B'
	fct_review "${finish}" 'readiness review of the build that needs work'
	base="$(git -C "${finish}" rev-parse HEAD)"
	for forged in verdict state none view; do
		case "${forged}" in
		verdict) fct_edit "${finish}/review.toml" 'verdict = "Needs work"' 'verdict = "Ready"' ;;
		state)
			fct_edit "${finish}/review.toml" 'verdict = "Needs work"' 'verdict = "Ready"'
			fct_edit "${finish}/review.toml" 'state = "open"' 'state = "fixed"'
			;;
		none) fct_readiness "${head3}" Ready '' 'agent B' ;;
		view) fct_readiness "${head3}" Ready "${valid}"$'\n\n'"$(fct_finding R5-F1 block fixed bad-input "commit = \"${head3}\"")" 'agent B' ;;
		esac
		fct_records
		[[ "${forged}" != view ]] || fct_edit "${finish}/readiness.md" 'Verdict: Ready' 'Verdict: Needs work'
		fct_review "${finish}" "a review that ends the build: ${forged}"
		fct_steps "${finish}"
		printf 'Handler of the build after it.\n' >>"${finish}/README.md"
		git -C "${finish}" commit -qam 'handler of the build after it'
		ledger_earlier="${round1} ${round2} ${head}:pass ${head2}:pass ${head3}:pass"
		fct_readiness "$(git -C "${finish}" rev-parse HEAD)" Ready "${valid}"$'\n\n'"$(fct_finding R5-F1 block fixed bad-input "commit = \"${head3}\"")" 'agent B'
		fct_finish 1
		fct_finish_said 'review.toml: round 5 says pass, but just finish fails on its review in'
		git -C "${finish}" reset -q --hard "${base}"
	done

	git -C "${finish}" reset -q --hard "${passed}"
	# A review that ended a build passes as just finish read it then. A limit
	# that a later build accepts changes nothing for it, and an open line of
	# its time keeps it from ending the build.
	for forged in limit open; do
		if [[ "${forged}" == limit ]]; then
			fct_edit "${finish}/spec.md" '**Limits.** None.' "${limit}"
			git -C "${finish}" commit -qam 'a later build accepts the limit long-lines'
		else
			fct_edit "${finish}/spec.md" 'None.' $'| ID | Command | Part | Case | Code |\n|---|---|---|---|---|\n| `slow` | inspect | Failures | A file that the program reads slowly | `RUNTIME` |'
			git -C "${finish}" commit -qam 'an open line'
			head3="$(git -C "${finish}" rev-parse HEAD)" ledger_earlier="${round1} ${round2} ${head}:pass ${head2}:pass"
			fct_readiness "${head3}" Ready "${valid}" 'agent B'
			fct_review "${finish}" 'a review that misses the open line'
			ledger_earlier="${ledger_earlier} ${head3}:pass"
		fi
		fct_steps "${finish}"
		printf 'Handler of the build after it.\n' >>"${finish}/README.md"
		git -C "${finish}" commit -qam 'handler of the build after it'
		[[ "${forged}" == open ]] || ledger_earlier="${round1} ${round2} ${head}:pass ${head2}:pass"
		fct_readiness "$(git -C "${finish}" rev-parse HEAD)" 'Ready with stated limits' "${valid}" 'agent B'
		if [[ "${forged}" == limit ]]; then
			fct_finish 0
		else
			fct_finish 1
			fct_finish_said 'review.toml: round 5 says pass, but just finish fails on its review in'
		fi
		git -C "${finish}" reset -q --hard "${passed}"
	done

	# An extension with a Markdown readiness.md starts a new build of records
	# when it updates, and keeps the Markdown reviews as history.
	printf 'Commit: %s\nReviewer: agent B\nVerdict: Ready\n\n## Findings\n\n### Grade: block\n' "${head2}" >"${finish}/readiness.md"
	git -C "${finish}" rm -q review.toml
	git -C "${finish}" commit -qam 'a Markdown review of a template before typed records'
	ledger_earlier=""
	fct_readiness "${head2}" Ready "${advise}" 'agent B'
	fct_finish 1
	fct_finish_said 'review.toml: the intake of this build comes before the last Markdown readiness.md, in'
	git -C "${finish}" checkout -q readiness.md
	fct_steps "${finish}"
	printf 'Handler of the first build of records.\n' >>"${finish}/README.md"
	git -C "${finish}" commit -qam 'handler of the first build of records'
	fct_readiness "$(git -C "${finish}" rev-parse HEAD)" Ready "${advise}" 'agent B'
	fct_finish 0
	git -C "${finish}" reset -q --hard "${passed}"
}
