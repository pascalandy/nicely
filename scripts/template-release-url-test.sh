#!/usr/bin/env bash
# Sourced by template-release-test.sh. It proves the repository URL that a rendered
# README.md gives for a module path with a major version suffix.

# A Go module's /v2 suffix is not part of its published repository URL.
fct_test_release_url() {
	local template_src="${1}" scratch="${2}" copier
	shift 2
	copier=("$@")
	"${copier[@]}" copy --quiet --vcs-ref v0.0.1 --defaults --data domain=probe --data module=example.com/ncly-probe/v2 "${template_src}" "${scratch}/ncly-probe-v2"
	fct_has "${scratch}/ncly-probe-v2/README.md" 'git clone --branch TAG https://example.com/ncly-probe ncly-probe'
	if grep -qF 'example.com/ncly-probe/v2' "${scratch}/ncly-probe-v2/README.md"; then
		fct_fail 'README.md gives the module path example.com/ncly-probe/v2 as a repository URL'
	fi
}
