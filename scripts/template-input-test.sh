#!/usr/bin/env bash
# Sourced by template-test.sh to prove the Copier inputs: domains that resemble Go
# imports still build, Copier refuses each invalid input for its own reason, and
# a development build pins a commit that Go accepts.

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

fct_test_refused_inputs() {
	local template_src="${1}" scratch="${2}" invalid_domain invalid_name derived_key copier
	shift 2
	copier=("$@")
	# Copier refuses main, which Go builds as a program that no package imports. It refuses nicely, whose skill the bundled one hides.
	for invalid_domain in main nicely $'probe\n'; do
		if "${copier[@]}" copy --quiet --vcs-ref v0.0.1 --defaults --data "domain=${invalid_domain}" --data module=example.com/ncly-probe "${template_src}" "${scratch}/invalid-domain" 2>"${scratch}/refusal.out"; then
			fct_fail "Copier accepts the domain ${invalid_domain}"
		fi
		grep -qE 'choose another domain|Give a domain' "${scratch}/refusal.out" || fct_fail "Copier refuses the domain ${invalid_domain} for another reason: $(cat "${scratch}/refusal.out")"
	done
	# A name does not help, since ncly bundles skills named nicely-<topic> too.
	if "${copier[@]}" copy --quiet --vcs-ref v0.0.1 --defaults --data domain=nicely --data name=auth --data module=example.com/ncly-nicely-auth "${template_src}" "${scratch}/invalid-domain" 2>"${scratch}/refusal.out"; then
		fct_fail "Copier accepts the domain nicely with the name auth"
	fi
	grep -qF 'choose another domain' "${scratch}/refusal.out" || fct_fail "Copier refuses nicely-auth for another reason: $(cat "${scratch}/refusal.out")"
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

fct_test_development_build() {
	local template_src="${1}" scratch="${2}" src="${2}/development" repo="${2}/ncly-development" commit pin copier
	shift 2
	copier=("$@")
	# A development build of ncly gives a commit after a tag as VERSION (#405).
	git clone -q "${template_src}" "${src}"
	git -C "${src}" commit -q --allow-empty -m 'development build'
	commit="$(git -C "${src}" rev-parse --short=12 HEAD)"
	"${copier[@]}" copy --quiet --vcs-ref "${commit}" --defaults --data domain=probe --data module=example.com/ncly-probe "${src}" "${repo}"
	grep -qE "^_commit: v[0-9.]+-[0-9]+-g${commit:0:7}[0-9a-f]*$" "${repo}/.copier-answers.yml" ||
		fct_fail "Copier gives no git describe of ${commit}: $(grep '^_commit:' "${repo}/.copier-answers.yml")"
	# Go refuses the describe form. go mod tidy turns the commit into a pseudo-version.
	pin="$(sed -n 's/^require github.com\/pascalandy\/nicely //p' "${repo}/go.mod")"
	[[ "${pin}" =~ ^[0-9a-f]{7,}$ && "${commit}" == "${pin:0:12}"* ]] ||
		fct_fail "the go.mod of a development build pins ${pin}, not the commit ${commit}"
}
