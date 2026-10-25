#!/usr/bin/env bash
# Sourced by template-test.sh to prove the release steps and the install
# commands of a rendered repository, as "Release and install" in
# docs/north-star/contract/extensions.md says.

# fct_test_release releases the rendered repository $1 as the tag v0.1.0, then
# installs it as a user does with the commands of its README.md, from a clone
# at that tag. The install goes to a new data folder, where ncly describe then
# finds the root $4 of the slug $3, so no earlier install can answer.
fct_test_release() {
	local repo="${1}" scratch="${2}" slug="${3}" extension_root="${4}" root_words data="${2}/release-data"
	read -r -a root_words <<<"${extension_root}"
	fct_has "${repo}/AGENTS.md" "## Release"
	fct_has "${repo}/README.md" "git clone --branch TAG https://example.com/ncly-${slug} ncly-${slug}"
	fct_has "${repo}/README.md" "ncly describe ${extension_root} --json"
	# An annotated tag makes Git warn that the tag is not a commit, which README.md names.
	git -C "${repo}" tag -a v0.1.0 -m v0.1.0
	git clone --quiet --branch v0.1.0 "${repo}" "${scratch}/ncly-${slug}-v0.1.0" 2>"${scratch}/clone.err"
	(cd "${scratch}/ncly-${slug}-v0.1.0" && XDG_DATA_HOME="${data}" just install)
	[[ -x "${data}/nicely/extensions/ncly-${slug}" ]] || fct_fail "the install of the tag wrote no ${data}/nicely/extensions/ncly-${slug}"
	XDG_DATA_HOME="${data}" PATH="/usr/bin:${PATH}" "$(command -v go)" -C "${repo}" tool ncly describe "${root_words[@]}" --json >"${scratch}/describe.out"
	grep -qE "\"path\":\"${extension_root}\",\"summary\":\"[^\"]*\",\"source\":\"extension\"" "${scratch}/describe.out" ||
		fct_fail "ncly describe does not find ${extension_root} after the install of the tag: $(cat "${scratch}/describe.out")"
}

# fct_test_release_url renders the template source $1 for a module path with a
# major version suffix, such as /v2. Go adds that suffix to the module path, not
# to the repository URL, so the clone commands of README.md omit it.
fct_test_release_url() {
	local template_src="${1}" scratch="${2}" copier
	shift 2
	copier=("$@")
	"${copier[@]}" copy --quiet --vcs-ref v0.0.1 --defaults --data domain=probe --data module=example.com/ncly-probe/v2 "${template_src}" "${scratch}/ncly-probe-v2"
	fct_has "${scratch}/ncly-probe-v2/README.md" "git clone --branch TAG https://example.com/ncly-probe ncly-probe"
	if grep -qF 'example.com/ncly-probe/v2' "${scratch}/ncly-probe-v2/README.md"; then
		fct_fail "README.md gives the module path example.com/ncly-probe/v2 as a repository URL"
	fi
}
