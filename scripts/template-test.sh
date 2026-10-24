#!/usr/bin/env bash
# template-test.sh renders the extension template of this checkout with Copier,
# as D062 decides, and proves the rendered repository. It passes go mod tidy,
# just check, and just install into scratch folders. just update then carries
# a customized repository to the next template revision and keeps the edits of
# its author. On a conflict, it stops before any go command. The test runs on
# macOS and Linux, with the bash 3.2 of macOS, and uses no network once uv
# holds the pinned Copier.
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
}

# fct_commit commits every change of the rendered repository, as its author.
fct_commit() {
	git -C "${repo}" add -A
	git -C "${repo}" commit -qm "${1}"
}

# fct_edit replaces the line $2 of the file $1 with $3, or appends $3 when $2
# is empty. sed -i differs between macOS and Linux, so the test avoids it.
fct_edit() {
	local file="${1}" old="${2}" new="${3}" line found=""
	while IFS= read -r line || [[ -n "${line}" ]]; do
		if [[ -n "${old}" && "${line}" == "${old}" ]]; then
			line="${new}" found=1
		fi
		printf '%s\n' "${line}"
	done <"${file}" >"${file}.new"
	if [[ -z "${old}" ]]; then
		printf '%s\n' "${new}" >>"${file}.new"
	elif [[ -z "${found}" ]]; then
		fct_fail "${file} holds no line \"${old}\""
	fi
	mv "${file}.new" "${file}"
}

# fct_has fails unless the file $1 holds the line $2.
fct_has() {
	grep -qxF -- "${2}" "${1}" || fct_fail "${1} lacks the line \"${2}\""
}

# fct_update runs just update $1 in the rendered repository. It keeps the
# output in update.out and returns the exit code of just.
fct_update() {
	local status=0
	(cd "${repo}" && just update "${1}") >"${scratch}/update.out" 2>&1 || status=$?
	return "${status}"
}

# fct_said fails unless the last just update printed the line $1.
fct_said() {
	if ! grep -qxF -- "${1}" "${scratch}/update.out"; then
		cat "${scratch}/update.out" >&2
		fct_fail "just update did not print \"${1}\""
	fi
}

fct_repository() {
	local scratch="${1}" domain="${2}" name="${3}" slug="${4}" command="${5}" src repo data extensions lipgloss file
	shift 5
	src="${scratch}/source" repo="${scratch}/ncly-${slug}" data="${scratch}/data"
	extensions="${data}/nicely/extensions"
	mkdir -p "${scratch}"
	# The template source holds the working tree of this checkout, changes included, tagged as a version.
	mkdir "${src}"
	cp "${root}/copier.yml" "${src}/"
	cp -R "${root}/template" "${src}/template"
	git -C "${src}" init -q
	fct_revision v0.0.1
	"${copier[@]}" copy --quiet --vcs-ref v0.0.1 --defaults --data "domain=${domain}" --data "module=example.com/ncly-${slug}" "$@" "${src}" "${repo}"

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
	git -C "${repo}" init -q
	fct_commit "rendered from v0.0.1"
	(cd "${repo}" && just check)
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
	fct_test_update "${src}" "${repo}" "${scratch}" "${root}"
	if [[ -z "${name}" ]]; then
		[[ -z "$(sed -n '/^\[/q; /^name = /p' "${repo}/manifest.toml")" ]] || fct_fail 'update changes the whole-domain default'
	else
		fct_has "${repo}/manifest.toml" "name = \"${name}\""
	fi
	printf 'template %s passes checks, replaces its install, and preserves author edits through updates\n' "${slug}"
}

fct_main() {
	local root pin scratch copier template_src invalid_domain invalid_name derived_key
	root="$(cd "${BASH_SOURCE[0]%/*}/.." && pwd -P)"
	# shellcheck source=scripts/template-update-test.sh
	source "${root}/scripts/template-update-test.sh"
	# shellcheck source=scripts/template-input-test.sh
	source "${root}/scripts/template-input-test.sh"
	# A Git hook or a caller can export GIT_DIR. The scratch repositories must see their own.
	while IFS= read -r variable; do unset "${variable}"; done < <(git rev-parse --local-env-vars)
	# just update runs the Copier that the template pins, and so does this test.
	pin="$(sed -n 's/^copier := "\(.*\)"$/\1/p' "${root}/template/justfile")"
	[[ "${pin}" == "uv tool run copier@"* ]] || fct_fail "template/justfile pins no Copier"
	grep -qF -- "${pin}" "${root}/extensions/skill/nicely/SKILL.md" || fct_fail "the nicely skill does not give ${pin}, which template/justfile pins"
	scratch="$(mktemp -d "${TMPDIR:-/tmp}/ncly-template-test.XXXXXXXX")"
	readonly NCLY_TEMPLATE_TEST_ROOT="${scratch}"
	trap 'rm -rf "${NCLY_TEMPLATE_TEST_ROOT}"' EXIT
	# Git, in Copier too, takes its identity from this file, without the signing of the user.
	printf '[user]\n\tname = template-test\n\temail = template-test@example.com\n[commit]\n\tgpgsign = false\n[tag]\n\tgpgsign = false\n[init]\n\tdefaultBranch = main\n' >"${scratch}/gitconfig"
	export GIT_CONFIG_GLOBAL="${scratch}/gitconfig" GIT_CONFIG_NOSYSTEM=1
	read -r -a copier <<<"${pin}"
	fct_repository "${scratch}/named" video convert video-convert 'video convert' --data name=convert
	fct_repository "${scratch}/whole" probe '' probe 'probe inspect'
	template_src="${scratch}/whole/source"
	fct_test_import_domains "${template_src}" "${scratch}" "${root}" "${copier[@]}"
	# Copier refuses main, which Go builds as a program that no package imports. It refuses nicely, whose skill the bundled one hides.
	for invalid_domain in main nicely $'probe\n'; do
		if "${copier[@]}" copy --quiet --vcs-ref v0.0.1 --defaults --data "domain=${invalid_domain}" --data module=example.com/ncly-probe "${template_src}" "${scratch}/invalid-domain" 2>"${scratch}/refusal.out"; then
			fct_fail "Copier accepts the domain ${invalid_domain}"
		fi
		grep -qE 'choose another domain|Give a domain' "${scratch}/refusal.out" || fct_fail "Copier refuses the domain ${invalid_domain} for another reason: $(cat "${scratch}/refusal.out")"
	done
	for invalid_name in Convert 2convert convert-more 'convert more' convert_more $'convert\n'; do
		if "${copier[@]}" copy --quiet --vcs-ref v0.0.1 --defaults --data domain=video --data "name=${invalid_name}" --data module=example.com/ncly-video-convert "${template_src}" "${scratch}/invalid" 2>"${scratch}/refusal.out"; then
			fct_fail "Copier accepts the name ${invalid_name}"
		fi
		grep -qF 'choose another name' "${scratch}/refusal.out" || fct_fail "Copier refuses ${invalid_name} for another reason: $(cat "${scratch}/refusal.out")"
	done
	for derived_key in slug root example_command; do
		if "${copier[@]}" copy --quiet --vcs-ref v0.0.1 --defaults --data domain=video --data name=convert --data "${derived_key}=wrong" --data module=example.com/ncly-video-convert "${template_src}" "${scratch}/invalid-derived" 2>"${scratch}/refusal.out"; then
			fct_fail "Copier accepts an override of ${derived_key}"
		fi
		grep -qF 'Pass only domain, name, and module' "${scratch}/refusal.out" || fct_fail "Copier refuses ${derived_key} for another reason: $(cat "${scratch}/refusal.out")"
	done
}

fct_main
exit 0
