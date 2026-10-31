#!/usr/bin/env bash
# template-test.sh renders the extension template of this checkout with Copier,
# as D062 decides, and proves the rendered repository. It passes go mod tidy,
# just check, just install, just uninstall, and just acceptance into scratch
# folders. A clone at a release tag installs with the commands of its
# README.md. just update then carries a customized repository to the next
# template revision, and keeps the edits and the dependencies of its author. On
# a conflict, it stops before dependency changes. The test runs on macOS and Linux,
# with the bash 3.2 of macOS, and uses no network once uv holds the pinned
# Copier.
set -Eeuo pipefail
trap 'printf "%s: line %s failed: %s\n" "${BASH_SOURCE[0]##*/}" "${LINENO}" "${BASH_COMMAND}" >&2' ERR

fct_fail() {
	printf 'template-test.sh: %s\n' "$*" >&2
	exit 1
}

# fct_revision commits the template source as the template version $1.
fct_revision() {
	git -C "${src}" add -A
	git -C "${src}" commit -qm "template ${1}"
	git -C "${src}" tag "${1}"
	fct_proxy_revision "${src}" "${1}" "${proxy}"
}

# fct_commit commits every change of the rendered repository, as its author.
fct_commit() {
	git -C "${repo}" add -A
	git -C "${repo}" commit -qm "${1}"
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
	fct_test_finish "${repo}" "${scratch}"
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

fct_main() {
	local root pin scratch copier template_src proxy cache_proxy
	root="$(cd "${BASH_SOURCE[0]%/*}/.." && pwd -P)"
	# Copied recipes only wire commands to the SDK or an owned script.
	awk '
		/^[^[:space:]#].*:[[:space:]]*$/ { recipe = $1; sub(/:$/, "", recipe); lines = 0; reported = 0 }
		/^[[:space:]]+#!/ || /^[[:space:]]+[^[:space:]#]/ {
			if ((++lines > 1 || $1 ~ /^#!/) && !reported++) {
				print "template-test.sh: recipe " recipe " must have one command line" > "/dev/stderr"
				failed = 1
			}
		}
		END { exit failed }
	' "${root}/template/justfile"
	# shellcheck source=scripts/template-test-helpers.sh
	source "${root}/scripts/template-test-helpers.sh"
	# shellcheck source=scripts/template-update-test.sh
	source "${root}/scripts/template-update-test.sh"
	# shellcheck source=scripts/template-input-test.sh
	source "${root}/scripts/template-input-test.sh"
	# shellcheck source=scripts/template-release-test.sh
	source "${root}/scripts/template-release-test.sh"
	# shellcheck source=scripts/template-recipes-test.sh
	source "${root}/scripts/template-recipes-test.sh"
	# shellcheck source=scripts/template-signoff-test.sh
	source "${root}/scripts/template-signoff-test.sh"
	# shellcheck source=scripts/template-finish-test.sh
	source "${root}/scripts/template-finish-test.sh"
	# A Git hook or a caller can export GIT_DIR. The scratch repositories must see their own.
	while IFS= read -r variable; do unset "${variable}"; done < <(git rev-parse --local-env-vars)
	# just update runs the Copier that the template pins, and so does this test.
	pin="$(sed -n 's/^copier := "\(.*\)"$/\1/p' "${root}/template/justfile")"
	[[ "${pin}" == "uv tool run copier@"* ]] || fct_fail "template/justfile pins no Copier"
	grep -qF -- "${pin}" "${root}/extensions/skill/nicely-extensions/SKILL.md" || fct_fail "the nicely-extensions skill does not give ${pin}, which template/justfile pins"
	# The recipes print clean paths. On macOS, TMPDIR ends in a slash and sits under the /var link, so resolve the scratch root.
	scratch="$(cd "$(mktemp -d "${TMPDIR:-/tmp}/ncly-template-test.XXXXXXXX")" && pwd -P)"
	readonly NCLY_TEMPLATE_TEST_ROOT="${scratch}"
	trap 'rm -rf "${NCLY_TEMPLATE_TEST_ROOT}"' EXIT
	# Git, in Copier too, takes its identity from this file, without the signing of the user.
	printf '[user]\n\tname = template-test\n\temail = template-test@example.com\n[commit]\n\tgpgsign = false\n[tag]\n\tgpgsign = false\n[init]\n\tdefaultBranch = main\n' >"${scratch}/gitconfig"
	export GIT_CONFIG_GLOBAL="${scratch}/gitconfig" GIT_CONFIG_NOSYSTEM=1
	proxy="${scratch}/proxy/github.com/pascalandy/nicely/@v"
	mkdir -p "${proxy}"
	cache_proxy="file://$(go env GOMODCACHE)/cache/download"
	export GOPROXY="file://${scratch}/proxy,${cache_proxy},${GOPROXY:-https://proxy.golang.org,direct}"
	export GOMODCACHE="${scratch}/modcache" GONOSUMDB=github.com/pascalandy/nicely
	# Go normally makes cached module folders read-only, which prevents scratch cleanup.
	export GOFLAGS="${GOFLAGS:-} -modcacherw"
	read -r -a copier <<<"${pin}"
	fct_repository "${scratch}/named" video convert video-convert 'video convert' --data name=convert
	fct_repository "${scratch}/whole" probe '' probe 'probe inspect'
	fct_test_bootstrap "${scratch}/bootstrap" "${copier[@]}"
	template_src="${scratch}/whole/source"
	fct_test_import_domains "${template_src}" "${scratch}" "${root}" "${copier[@]}"
	fct_test_release_url "${template_src}" "${scratch}" "${copier[@]}"
	fct_test_refused_inputs "${template_src}" "${scratch}" "${copier[@]}"
	fct_test_development_build "${template_src}" "${scratch}" "${copier[@]}"
}

fct_main
exit 0
