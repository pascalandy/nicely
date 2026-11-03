#!/usr/bin/env bash
# Sourced by template-test.sh to render and check one extension repository.
# shellcheck disable=SC2154 # root and copier come from fct_main in template-test.sh.

# Every Nicely documentation link in a render uses its selected template tag.
fct_revision_links() {
	local links link
	links="$(grep -RhoE --exclude-dir=.git 'https://github\.com/pascalandy/nicely/blob/[^[:space:])>]+' "${1}")" || fct_fail 'the render has no Nicely documentation links'
	while IFS= read -r link; do
		[[ "${link}" == "https://github.com/pascalandy/nicely/blob/${2}/"* ]] || fct_fail "rendered link uses another revision: ${link}"
	done <<<"${links}"
}

fct_repository() {
	local scratch="${1}" domain="${2}" name="${3}" slug="${4}" command="${5}" src repo data extensions lipgloss file
	shift 5
	src="${scratch}/source" repo="${scratch}/ncly-${slug}" data="${scratch}/data"
	extensions="${data}/nicely/extensions"
	mkdir -p "${scratch}"
	# The module proxy and Copier source both hold this checkout, with its changes.
	mkdir "${src}"
	fct_copy_source "${src}"
	git -C "${src}" init -q
	fct_revision v0.0.1
	"${copier[@]}" copy --quiet --vcs-ref v0.0.1 --defaults --data "domain=${domain}" --data "module=example.com/ncly-${slug}" "$@" "${src}" "${repo}"
	fct_revision_links "${repo}" v0.0.1

	# The reading order must send an author to the real-input acceptance criteria.
	awk '
		/^## Read, in this order$/ { reading = 1; next }
		reading && /^## / { reading = 0 }
		reading && /^[0-9]+\. / && /## Acceptance/ { found = 1 }
		END { exit !found }
	' "${repo}/AGENTS.md" || fct_fail 'AGENTS.md reading order must name ## Acceptance'

	# The author's go.mod points at this checkout, so the rendered repository tests the SDK and ncly of this change.
	go -C "${repo}" mod edit -replace "github.com/pascalandy/nicely=${root}"
	# The rendered go.mod repeats the replace of Lip Gloss of Nicely's go.mod, or drops it with Nicely's.
	lipgloss="$(grep '^replace charm.land/lipgloss/v2 ' "${root}/go.mod" || true)"
	[[ "$(grep '^replace charm.land/lipgloss/v2 ' "${repo}/go.mod" || true)" == "${lipgloss}" ]] ||
		fct_fail "the go.mod of the template does not repeat \"${lipgloss}\" of Nicely's go.mod"
	go -C "${repo}" mod tidy
	fct_has "${repo}/manifest.toml" "domain = \"${domain}\""
	fct_has "${repo}/extension.go" "package ${domain}"
	if [[ -n "${name}" ]]; then
		fct_has "${repo}/manifest.toml" "name = \"${name}\""
	else
		[[ -z "$(sed -n '/^\[/q; /^name = /p' "${repo}/manifest.toml")" ]] || fct_fail 'the whole-domain manifest declares a name'
	fi
	[[ -f "${repo}/cmd/ncly-${slug}/main.go" ]] || fct_fail "the entry directory is not cmd/ncly-${slug}"
	fct_test_files "${repo}" "${slug}"
	git -C "${repo}" init -q
	fct_commit "rendered from v0.0.1"
	(cd "${repo}" && just check)
	fct_test_gates "${repo}" "${scratch}" "${domain}" "${command}"
	fct_test_acceptance_cases "${repo}" "${scratch}"
	fct_test_finish "${repo}" "${scratch}"
	fct_test_ledger "${repo}" "${scratch}"
	fct_test_example_notices "${repo}" "${scratch}" "${slug}"
	export XDG_CONFIG_HOME="${scratch}/config" XDG_STATE_HOME="${scratch}/state" XDG_DATA_HOME="${data}" NCLY_NO_INPUT=1
	(cd "${repo}" && just install)
	for file in "ncly-${slug}" "ncly-${slug}.toml" "ncly-${slug}.locales/en.toml" "ncly-${slug}.skill/SKILL.md"; do
		[[ -f "${extensions}/${file}" ]] || fct_fail "just install wrote no ${extensions}/${file}"
	done
	local command_words
	read -r -a command_words <<<"${command}"
	# The installed extension runs under ncly. /usr/bin comes first, so the shim of a version manager never answers for python3.
	# go stays the caller's: the /usr/bin/go of the release runner is older than go.mod allows.
	PATH="/usr/bin:${PATH}" "$(command -v go)" -C "${repo}" tool ncly "${command_words[@]}" "${repo}/go.mod" --json >"${scratch}/inspect.out"
	grep -qF '"ok":true' "${scratch}/inspect.out" || fct_fail "the installed command failed: $(cat "${scratch}/inspect.out")"
	PATH="/usr/bin:${PATH}" "$(command -v go)" -C "${repo}" tool ncly skill view "${slug}" >"${scratch}/skill.out"
	fct_has "${scratch}/skill.out" "name: ${slug}"
	grep -qF "ncly ${command}" "${scratch}/skill.out" || fct_fail 'the installed skill teaches the wrong command'
	[[ "$(cd "${repo}" && just contract)" == "${root}/docs/north-star/contract" ]] || fct_fail "just contract does not print ${root}/docs/north-star/contract"
	printf 'stale\n' >"${extensions}/ncly-${slug}.locales/stale.toml"
	printf 'stale\n' >"${extensions}/ncly-${slug}.skill/stale.md"
	printf 'stale\n' >"${extensions}/ncly-${slug}.toml"
	printf 'stale\n' >"${extensions}/ncly-${slug}"
	(cd "${repo}" && just install)
	[[ ! -e "${extensions}/ncly-${slug}.locales/stale.toml" && ! -e "${extensions}/ncly-${slug}.skill/stale.md" ]] || fct_fail 'install keeps stale sidecars'
	PATH="/usr/bin:${PATH}" "$(command -v go)" -C "${repo}" tool ncly "${command_words[@]}" "${repo}/go.mod" --json >"${scratch}/reinstalled.out"
	grep -qF '"ok":true' "${scratch}/reinstalled.out" || fct_fail 'the replacement install does not run'
	fct_test_uninstall "${repo}" "${slug}" "${data}" "${domain}${name:+ ${name}}" "${scratch}"
	fct_test_acceptance "${repo}" "${scratch}"
	fct_test_release "${repo}" "${scratch}" "${slug}" "${domain}${name:+ ${name}}"
	fct_test_update "${src}" "${repo}" "${scratch}" "${root}"
	if [[ -z "${name}" ]]; then
		[[ -z "$(sed -n '/^\[/q; /^name = /p' "${repo}/manifest.toml")" ]] || fct_fail 'update changes the whole-domain default'
	else
		fct_has "${repo}/manifest.toml" "name = \"${name}\""
	fi
	printf 'template %s passes checks, replaces its install, and preserves author edits through updates\n' "${slug}"
}
