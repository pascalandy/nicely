#!/usr/bin/env bash
# template-test.sh renders the extension template of this checkout with Copier,
# as D062 decides, and proves the rendered repository. It passes go mod tidy,
# just check, just install, just uninstall, and just acceptance into scratch
# folders. A clone at a release tag installs with the commands of its
# README.md. just update then carries a customized repository to the next
# template revision, and keeps the edits and the dependencies of its author. On
# a conflict, it stops before dependency changes. An extension of v0.23.4, before
# just finish, updates and passes just signoff after its first build. The test runs on macOS and Linux,
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

fct_main() {
	local root pin scratch copier template_src proxy cache_proxy
	root="$(cd "${BASH_SOURCE[0]%/*}/.." && pwd -P)"
	# shellcheck source=scripts/template-test-helpers.sh
	source "${root}/scripts/template-test-helpers.sh"
	# shellcheck source=scripts/template-repository-test.sh
	source "${root}/scripts/template-repository-test.sh"
	# shellcheck source=scripts/template-update-test.sh
	source "${root}/scripts/template-update-test.sh"
	# shellcheck source=scripts/template-adoption-test.sh
	source "${root}/scripts/template-adoption-test.sh"
	# shellcheck source=scripts/template-input-test.sh
	source "${root}/scripts/template-input-test.sh"
	# shellcheck source=scripts/template-release-test.sh
	source "${root}/scripts/template-release-test.sh"
	# shellcheck source=scripts/template-recipes-test.sh
	source "${root}/scripts/template-recipes-test.sh"
	# shellcheck source=scripts/template-acceptance-test.sh
	source "${root}/scripts/template-acceptance-test.sh"
	fct_test_recipe_lines "${root}/template/justfile"
	# shellcheck source=scripts/template-signoff-test.sh
	source "${root}/scripts/template-signoff-test.sh"
	# shellcheck source=scripts/template-finish-test.sh
	source "${root}/scripts/template-finish-test.sh"
	# shellcheck source=scripts/template-ledger-test.sh
	source "${root}/scripts/template-ledger-test.sh"
	# shellcheck source=scripts/template-ledger-builds-test.sh
	source "${root}/scripts/template-ledger-builds-test.sh"
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
	# Update drops GIT_CONFIG_GLOBAL too. Its children need the same private
	# identity through HOME, while tool caches keep their existing locations.
	export GOCACHE="${GOCACHE:-$(go env GOCACHE)}" UV_CACHE_DIR="${UV_CACHE_DIR:-$(uv cache dir)}"
	export HOME="${scratch}/home" XDG_CONFIG_HOME="${scratch}/home/.config"
	mkdir -p "${HOME}"
	cp "${scratch}/gitconfig" "${HOME}/.gitconfig"
	export GOPROXY="file://${scratch}/proxy,${cache_proxy},${GOPROXY:-https://proxy.golang.org,direct}"
	export GOMODCACHE="${scratch}/modcache" GONOSUMDB=github.com/pascalandy/nicely
	# Go normally makes cached module folders read-only, which prevents scratch cleanup.
	export GOFLAGS="${GOFLAGS:-} -modcacherw"
	read -r -a copier <<<"${pin}"
	fct_repository "${scratch}/named" video convert video-convert 'video convert' --data name=convert
	fct_repository "${scratch}/whole" probe '' probe 'probe inspect'
	fct_test_bootstrap "${scratch}/bootstrap" "${copier[@]}"
	fct_test_adoption "${scratch}/adoption" "${copier[@]}"
	template_src="${scratch}/whole/source"
	fct_test_import_domains "${template_src}" "${scratch}" "${root}" "${copier[@]}"
	fct_test_release_url "${template_src}" "${scratch}" "${copier[@]}"
	fct_test_refused_inputs "${template_src}" "${scratch}" "${copier[@]}"
	fct_test_development_build "${template_src}" "${scratch}" "${copier[@]}"
}

fct_main
exit 0
