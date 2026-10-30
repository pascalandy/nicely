#!/usr/bin/env bash
# author-path.sh replays an author's first steps against a published Nicely tag.
# Usage: bash author-path.sh vX.Y.Z. Requires Go, Git, just, uv, and jq.
# Repositories and command logs stay in the reported scratch folder for review.
# Markdown backticks in patterns and fixture text are literal.
# shellcheck disable=SC2016
set -Eeuo pipefail
trap 'printf "author-path: unexpected failure at line %s; inspect %s\n" "${LINENO}" "${scratch:-the output}" >&2' ERR
# shellcheck source=.agents/skills/release-nicely/scripts/author-path-command.sh
source "$(dirname -- "${BASH_SOURCE[0]}")/author-path-command.sh"

fct_fail() {
	printf 'author-path: FAIL %s\n' "$*" >&2
	failures=$((failures + 1))
}

# Each invocation has its own log, including the exact command and exit code.
fct_run() {
	local label="${1}" folder="${2}" status=0
	shift 2
	printf 'author-path: run %s:' "${label}"
	printf ' %q' "$@"
	printf '\n'
	(cd "${folder}" && "$@") >"${scratch}/${label}.log" 2>&1 || status=$?
	{
		printf '%s\t%s\t' "${label}" "${status}"
		printf '%q ' "$@"
		printf '\n'
	} >>"${scratch}/commands.tsv"
	printf 'author-path: %s exit %s\n' "${label}" "${status}"
	return "${status}"
}

fct_execute_this() {
	local source pin repo skill_log status=0
	# This source is the published module, with no replacement from a checkout.
	fct_run module "${scratch}" go mod download -json "github.com/pascalandy/nicely@${tag}"
	source="$(jq -er '.Dir' "${scratch}/module.log")"
	pin="$(sed -n 's/^copier := "uv tool run copier@\([0-9.]*\)"$/\1/p' "${source}/template/justfile")"
	[[ -n "${pin}" ]] || {
		fct_fail 'the published template pins no Copier'
		return
	}
	copier=(uv tool run "copier@${pin}" copy --quiet --defaults --vcs-ref "${tag}" --data domain=probe)
	fct_run render "${scratch}" "${copier[@]}" --data module=example.com/ncly-probe gh:pascalandy/nicely "${scratch}/rendered"
	repo="${scratch}/rendered"
	cp "${repo}/AGENTS.md" "${scratch}/author-AGENTS.md"
	cat "${scratch}/author-AGENTS.md"
	if [[ -f "${repo}/README.md" ]]; then
		cp "${repo}/README.md" "${scratch}/author-README.md"
		cat "${scratch}/author-README.md"
		if ! grep -qF 'AGENTS.md' "${repo}/README.md"; then
			fct_fail 'README.md no longer routes authors to AGENTS.md'
		fi
	else
		fct_fail '#323: the published template gives the author no README.md'
	fi
	if ! grep -qF 'go mod tidy' "${repo}/AGENTS.md" || ! grep -qF 'just check' "${repo}/AGENTS.md"; then
		fct_fail 'the first-step documentation changed; update this replay'
		return
	fi
	fct_run tidy "${repo}" go mod tidy
	fct_run initial-check "${repo}" just check
	fct_run resolved-module "${repo}" go list -m -json github.com/pascalandy/nicely
	jq -e --arg tag "${tag}" '.Version == $tag and .Replace == null' "${scratch}/resolved-module.log" >/dev/null
	fct_run published-skill "${repo}" go tool ncly skill view nicely
	skill_log="${scratch}/published-skill.log"
	if grep -qF 'ncly skill view nicely-extensions' "${skill_log}"; then
		fct_run extension-skill "${repo}" go tool ncly skill view nicely-extensions
		skill_log="${scratch}/extension-skill.log"
	fi
	if grep -qE -- '--vcs-ref v[0-9]' "${skill_log}" "${source}/copier.yml"; then
		fct_fail '#323 #324: stale version in the published nicely skill or Copier command'
	elif ! grep -qF -- '--vcs-ref VERSION' "${skill_log}"; then
		fct_fail '#324: the nicely skill no longer gives the release VERSION to Copier'
	fi
	fct_run short-module "${scratch}" "${copier[@]}" --data module=probe gh:pascalandy/nicely "${scratch}/short-module" || status=$?
	if [[ "${status}" == 0 ]] || ! grep -qF 'Give the Go module path' "${scratch}/short-module.log"; then
		fct_fail '#323: Copier does not clearly refuse a short module path'
	fi
	# Separate copies keep each finding visible when another path fails first.
	cp -R "${repo}" "${scratch}/removal"
	repo="${scratch}/removal"
	fct_remove_example
	awk '/^## Change a command$/ { active=1; next } active && /^## / { exit } active && /^[0-9]+\. / { print }' \
		"${scratch}/author-AGENTS.md" >"${scratch}/change-steps.txt"
	cp -R "${scratch}/rendered" "${scratch}/command"
	repo="${scratch}/command"
	fct_command_order
}

fct_main() {
	local tag="${1:-}" tool scratch failures=0 variable found tool_path=''
	if [[ $# != 1 || ! "${tag}" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
		printf 'Usage: bash author-path.sh vX.Y.Z\n' >&2
		exit 2
	fi
	for tool in go git just uv jq; do
		command -v "${tool}" >/dev/null || {
			printf 'author-path: install %s, then retry\n' "${tool}" >&2
			exit 1
		}
		# Resolve version-manager shims before the XDG config moves.
		found="$(mise which "${tool}" 2>/dev/null || command -v "${tool}")"
		tool_path="${tool_path}${found%/*}:"
	done
	export PATH="${tool_path}/usr/bin:/bin:/usr/sbin:/sbin"
	while IFS= read -r variable; do unset "${variable}"; done < <(git rev-parse --local-env-vars)
	unset JUST_JUSTFILE JUST_WORKING_DIRECTORY JUST_UNSTABLE NCLY_JSON
	export GOWORK=off GOFLAGS='' GOENV=off GOPROXY=https://proxy.golang.org GONOPROXY=none NCLY_NO_INPUT=1
	scratch="$(mktemp -d "${TMPDIR:-/tmp}/ncly-author-path.XXXXXXXX")"
	printf 'author-path: evidence %s\n' "${scratch}"
	# Go's build cache stays shared; config, installed extensions, and state are isolated.
	export XDG_CONFIG_HOME="${scratch}/config" XDG_DATA_HOME="${scratch}/data" XDG_STATE_HOME="${scratch}/state"
	fct_execute_this
	if [[ "${failures}" != 0 ]]; then
		printf 'author-path: FAIL %s findings for %s; evidence %s\n' "${failures}" "${tag}" "${scratch}" >&2
		exit 1
	fi
	printf 'author-path: PASS %s; evidence %s\n' "${tag}" "${scratch}"
}

fct_main "$@"
