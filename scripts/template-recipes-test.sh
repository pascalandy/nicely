#!/usr/bin/env bash
# Sourced by template-test.sh to prove the files and the recipes that a
# rendered repository adds to the extension.

# fct_test_files checks the files of the rendered repository $1 that no
# recipe runs, and that its forms keep no Jinja. The slug is $2.
fct_test_files() {
	local repo="${1}" slug="${2}" file
	fct_has "${repo}/.github/workflows/check.yml" "        run: just check"
	fct_has "${repo}/.github/workflows/check.yml" "        os: \${{ github.event.repository.private && fromJSON('[\"ubuntu-24.04\"]') || fromJSON('[\"ubuntu-24.04\",\"macos-15\"]') }}"
	fct_has "${repo}/.github/workflows/check.yml" "    runs-on: \${{ matrix.os }}"
	fct_has "${repo}/qa.md" "# ${slug} QA"
	for file in qa.md acceptance.sh; do
		if grep -nE '\{\{|\}\}|\{%|%\}|\{#|#\}' "${repo}/${file}"; then
			fct_fail "${file} keeps Jinja after the render"
		fi
	done
}

# fct_test_uninstall removes the install that just install made in the data
# folder $3 for the rendered repository $1. ncly then finds no root $4 of the
# slug $2, and a second just uninstall finds nothing to remove.
fct_test_uninstall() {
	local repo="${1}" slug="${2}" extensions="${3}/nicely/extensions" extension_root="${4}" scratch="${5}" root_words file status=0
	read -r -a root_words <<<"${extension_root}"
	(cd "${repo}" && just uninstall) >"${scratch}/uninstall.out"
	[[ "$(head -n 1 "${scratch}/uninstall.out")" == "removed ${extensions}/ncly-${slug}" ]] ||
		fct_fail "just uninstall does not remove the executable first: $(cat "${scratch}/uninstall.out")"
	for file in "ncly-${slug}" "ncly-${slug}.toml" "ncly-${slug}.locales" "ncly-${slug}.skill"; do
		fct_has "${scratch}/uninstall.out" "removed ${extensions}/${file}"
	done
	[[ -z "$(ls -A "${extensions}")" ]] || fct_fail "just uninstall leaves $(ls -A "${extensions}") in ${extensions}"
	PATH="/usr/bin:${PATH}" "$(command -v go)" -C "${repo}" tool ncly describe "${root_words[@]}" --json >/dev/null 2>"${scratch}/describe.err" || status=$?
	if [[ "${status}" != 2 ]] || ! grep -qF '"code":"NOT_FOUND"' "${scratch}/describe.err"; then
		fct_fail "ncly describe ${extension_root} exits ${status} after just uninstall, want 2 with NOT_FOUND: $(cat "${scratch}/describe.err")"
	fi
	(cd "${repo}" && just uninstall) >"${scratch}/uninstall.out"
	fct_has "${scratch}/uninstall.out" "uninstall: nothing to remove, since ${extensions} holds no ncly-${slug}"
}

# fct_test_acceptance runs just acceptance in the rendered repository $1 on
# the input that it makes. Then it runs it from another folder on a read-only
# fixture, given by a relative path, which keeps its content.
fct_test_acceptance() {
	local repo="${1}" scratch="${2}" fixture="${2}/fixtures/notes.md"
	# The author fills acceptance.sh, so it starts free of findings.
	shellcheck "${repo}/acceptance.sh"
	shfmt -d "${repo}/acceptance.sh"
	if ! (cd "${repo}" && just acceptance) >"${scratch}/acceptance.out" 2>&1; then
		cat "${scratch}/acceptance.out" >&2
		fct_fail 'just acceptance fails on the input that it makes'
	fi
	grep -q '^acceptance: ncly .* passed on .*/work/notes\.md$' "${scratch}/acceptance.out" || fct_fail "just acceptance reports no pass: $(cat "${scratch}/acceptance.out")"
	mkdir -p "${scratch}/fixtures"
	printf 'one\ntwo\n' >"${fixture}"
	chmod 444 "${fixture}"
	if ! (cd "${scratch}/fixtures" && just --justfile "${repo}/justfile" acceptance notes.md) >"${scratch}/acceptance.out" 2>&1; then
		cat "${scratch}/acceptance.out" >&2
		fct_fail 'just acceptance fails on a fixture'
	fi
	grep -qF '"size_bytes":8' "${scratch}/acceptance.out" || fct_fail "just acceptance did not run on the fixture: $(cat "${scratch}/acceptance.out")"
	[[ "$(cat "${fixture}")" == $'one\ntwo' ]] || fct_fail 'just acceptance changes the fixture'
}
