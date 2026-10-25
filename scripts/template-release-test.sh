#!/usr/bin/env bash
# Sourced by template-test.sh to prove the release steps and the install
# commands of a rendered repository, as "Release and install" in
# docs/north-star/contract/extensions.md says.

# fct_test_release releases the rendered repository $1 as the tag v0.1.0, then
# installs it as a user does with the commands of its README.md, from a clone
# at that tag. ncly describe then finds the root $4 of the slug $3.
fct_test_release() {
	local repo="${1}" scratch="${2}" slug="${3}" extension_root="${4}" root_words
	read -r -a root_words <<<"${extension_root}"
	fct_has "${repo}/AGENTS.md" "## Release"
	fct_has "${repo}/README.md" "git clone --branch TAG https://example.com/ncly-${slug} ncly-${slug}"
	fct_has "${repo}/README.md" "ncly describe ${extension_root} --json"
	# An annotated tag makes Git warn that the tag is not a commit, which README.md names.
	git -C "${repo}" tag -a v0.1.0 -m v0.1.0
	git clone --quiet --branch v0.1.0 "${repo}" "${scratch}/ncly-${slug}-v0.1.0" 2>"${scratch}/clone.err"
	(cd "${scratch}/ncly-${slug}-v0.1.0" && just install)
	PATH="/usr/bin:${PATH}" "$(command -v go)" -C "${repo}" tool ncly describe "${root_words[@]}" --json >"${scratch}/describe.out"
	grep -qE "\"path\":\"${extension_root}\",\"summary\":\"[^\"]*\",\"source\":\"extension\"" "${scratch}/describe.out" ||
		fct_fail "ncly describe does not find ${extension_root} after the install of the tag: $(cat "${scratch}/describe.out")"
}
