#!/usr/bin/env bash
# Sourced by template-recipes-test.sh to prove just install and just uninstall
# in a rendered repository.

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
	fct_test_install_order "${repo}" "${slug}" "${extensions}" "${scratch}"
	fct_test_signoff "${repo}" "${scratch}"
}

# fct_test_install_order blocks manifest publication in $3, then tries just install in $1.
# A failed sidecar publication must leave no executable of the slug $2.
fct_test_install_order() {
	local repo="${1}" slug="${2}" extensions="${3}" scratch="${4}" status=0
	mkdir "${extensions}/ncly-${slug}.toml"
	(cd "${repo}" && just install) >"${scratch}/blocked-install.out" 2>&1 || status=$?
	[[ "${status}" != 0 ]] || fct_fail 'just install accepts a directory in place of its manifest'
	grep -qF "${extensions}/ncly-${slug}.toml" "${scratch}/blocked-install.out" ||
		fct_fail "just install failed before manifest publication: $(cat "${scratch}/blocked-install.out")"
	[[ ! -e "${extensions}/ncly-${slug}" && ! -L "${extensions}/ncly-${slug}" ]] ||
		fct_fail 'just install published the executable before its sidecars'
	mv "${extensions}/ncly-${slug}.toml" "${scratch}/blocked-manifest"
}
