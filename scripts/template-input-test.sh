#!/usr/bin/env bash
# Sourced by template-test.sh to prove domains that resemble Go imports still build.

fct_test_import_domains() {
	local template_src="${1}" scratch="${2}" root="${3}" valid_domain copier
	shift 3
	copier=("$@")
	# A domain that names an import of the template, or init, still gives a repository that builds.
	for valid_domain in extension fmt init; do
		"${copier[@]}" copy --quiet --vcs-ref v0.0.1 --defaults --data "domain=${valid_domain}" --data "module=example.com/ncly-${valid_domain}" "${template_src}" "${scratch}/ncly-${valid_domain}"
		go -C "${scratch}/ncly-${valid_domain}" mod edit -replace "github.com/pascalandy/nicely=${root}"
		go -C "${scratch}/ncly-${valid_domain}" mod tidy
		go -C "${scratch}/ncly-${valid_domain}" vet ./... || fct_fail "the repository that the template renders for the domain ${valid_domain} does not build"
	done
}
