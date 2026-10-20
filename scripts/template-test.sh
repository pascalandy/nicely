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

fct_main() {
	local root pin scratch src repo data extensions status copier
	root="$(cd "${BASH_SOURCE[0]%/*}/.." && pwd -P)"
	# A Git hook or a caller can export GIT_DIR. The scratch repositories must see their own.
	while IFS= read -r variable; do unset "${variable}"; done < <(git rev-parse --local-env-vars)
	# just update runs the Copier that the template pins, and so does this test.
	pin="$(sed -n 's/^copier := "\(.*\)"$/\1/p' "${root}/template/justfile")"
	[[ "${pin}" == "uv tool run copier@"* ]] || fct_fail "template/justfile pins no Copier"
	grep -qF -- "${pin}" "${root}/extensions/skill/nicely/SKILL.md" || fct_fail "the nicely skill does not give ${pin}, which template/justfile pins"
	scratch="$(mktemp -d "${TMPDIR:-/tmp}/ncly-template-test.XXXXXXXX")"
	readonly NCLY_TEMPLATE_TEST_ROOT="${scratch}"
	trap 'rm -rf "${NCLY_TEMPLATE_TEST_ROOT}"' EXIT
	src="${scratch}/source" repo="${scratch}/ncly-probe" data="${scratch}/data"
	extensions="${data}/nicely/extensions"
	# Git, in Copier too, takes its identity from this file, without the signing of the user.
	printf '[user]\n\tname = template-test\n\temail = template-test@example.com\n[commit]\n\tgpgsign = false\n[tag]\n\tgpgsign = false\n[init]\n\tdefaultBranch = main\n' >"${scratch}/gitconfig"
	export GIT_CONFIG_GLOBAL="${scratch}/gitconfig" GIT_CONFIG_NOSYSTEM=1

	# The template source holds the working tree of this checkout, changes included, tagged as a version.
	mkdir "${src}"
	cp "${root}/copier.yml" "${src}/"
	cp -R "${root}/template" "${src}/template"
	git -C "${src}" init -q
	fct_revision v0.0.1
	read -r -a copier <<<"${pin}"
	"${copier[@]}" copy --quiet --vcs-ref v0.0.1 --defaults --data domain=probe --data module=example.com/ncly-probe "${src}" "${repo}"

	# The author's go.mod points at this checkout, so the rendered repository tests the SDK and ncly of this change.
	go -C "${repo}" mod edit -replace "github.com/pascalandy/nicely=${root}"
	fct_has "${repo}/go.mod" "$(grep '^replace charm.land/lipgloss/v2 ' "${root}/go.mod")"
	go -C "${repo}" mod tidy
	git -C "${repo}" init -q
	fct_commit "rendered from v0.0.1"
	(cd "${repo}" && just check)
	(cd "${repo}" && XDG_DATA_HOME="${data}" just install)
	for file in ncly-probe ncly-probe.toml ncly-probe.locales/en.toml ncly-probe.skill/SKILL.md; do
		[[ -f "${extensions}/${file}" ]] || fct_fail "just install wrote no ${extensions}/${file}"
	done
	# The installed extension runs under ncly. /usr/bin comes first, so the shim of a version manager never answers for python3.
	XDG_DATA_HOME="${data}" PATH="/usr/bin:${PATH}" go -C "${repo}" tool ncly probe inspect "${repo}/go.mod" --json >"${scratch}/inspect.out"
	grep -qF '"ok":true' "${scratch}/inspect.out" || fct_fail "the installed ncly probe inspect failed: $(cat "${scratch}/inspect.out")"
	[[ "$(cd "${repo}" && just contract)" == "${root}/docs/north-star/contract" ]] || fct_fail "just contract does not print ${root}/docs/north-star/contract"

	# The author edits two files, and the next template revision changes one of them elsewhere.
	fct_edit "${repo}/vision.md" "" "Author's note: kept through each update."
	fct_edit "${repo}/AGENTS.md" "# AGENTS.md" "# AGENTS.md of the author"
	fct_commit "customized"
	fct_edit "${src}/template/AGENTS.md.jinja" "" "This line comes from template v0.0.2."
	fct_revision v0.0.2
	fct_update v0.0.2 || {
		cat "${scratch}/update.out" >&2
		fct_fail "just update v0.0.2 failed on a customized repository"
	}
	fct_has "${repo}/vision.md" "Author's note: kept through each update."
	fct_has "${repo}/AGENTS.md" "# AGENTS.md of the author"
	fct_has "${repo}/AGENTS.md" "This line comes from template v0.0.2."
	fct_has "${repo}/.copier-answers.yml" "_commit: v0.0.2"
	fct_has "${repo}/go.mod" "require github.com/pascalandy/nicely v0.0.2"
	fct_has "${repo}/go.mod" "replace github.com/pascalandy/nicely => ${root}"
	fct_commit "updated to v0.0.2"

	# The author and template v0.0.3 change the same line. The stub go logs each call, so a go get after the conflict shows.
	fct_edit "${repo}/AGENTS.md" "This line comes from template v0.0.2." "The author rewrote this line."
	fct_commit "rewritten"
	fct_edit "${src}/template/AGENTS.md.jinja" "This line comes from template v0.0.2." "Template v0.0.3 rewrote this line."
	fct_revision v0.0.3
	mkdir "${scratch}/bin"
	printf '#!/bin/sh\necho "go $*" >>"%s/go.log"\nexit 1\n' "${scratch}" >"${scratch}/bin/go"
	chmod 755 "${scratch}/bin/go"
	status=0
	PATH="${scratch}/bin:${PATH}" fct_update v0.0.3 || status=$?
	[[ "${status}" == 1 ]] || fct_fail "just update v0.0.3 exited with ${status} on a conflict, want 1"
	[[ ! -e "${scratch}/go.log" ]] || fct_fail "just update ran go after a conflict: $(cat "${scratch}/go.log")"
	fct_has "${repo}/AGENTS.md" "<<<<<<< before updating"
	fct_said "  AGENTS.md"
	fct_said "  go get github.com/pascalandy/nicely@v0.0.3"
	fct_said "  go mod tidy"
	fct_said "  just check"

	# On a Git tree with changes, just update stops before Copier.
	status=0
	fct_update v0.0.3 || status=$?
	[[ "${status}" == 1 ]] || fct_fail "just update exited with ${status} on a Git tree with changes, want 1"
	fct_said "update: the update needs a clean Git tree. Commit your changes, then run just update v0.0.3 again"

	# When Copier fails, just update says how to return to the commit before the update.
	git -C "${repo}" reset -q --hard
	status=0
	fct_update v9.9.9 || status=$?
	[[ "${status}" == 1 ]] || fct_fail "just update v9.9.9 exited with ${status} for a version that does not exist, want 1"
	fct_said "  git restore --source=$(git -C "${repo}" rev-parse HEAD) --staged --worktree -- ."
	fct_said "  just update v9.9.9"
	printf 'the template renders a repository that passes its checks, installs, and takes an update\n'
}

fct_main
exit 0
