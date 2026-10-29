#!/usr/bin/env bash
# Sourced by template-test.sh to prove the recipes that a rendered repository
# gives its user beside just install.

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
