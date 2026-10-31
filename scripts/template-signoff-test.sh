#!/usr/bin/env bash
# Sourced by template-test.sh to prove just signoff of the rendered
# repository. fct_test_uninstall calls it after its install tests.

# Run the rendered recipe with real Git and just. Only gh is a stub.
fct_test_signoff() {
	local repo="${1}" scratch="${2}" ready="${2}/signoff-ready" sha file status=0
	mkdir -p "${scratch}/signoff-bin"
	cat >"${scratch}/signoff-bin/gh" <<'SCRIPT'
#!/usr/bin/env bash
set -Eeuo pipefail
if [[ "$*" == --version ]]; then echo 'gh version test'; exit 0; fi
[[ "$#" == 3 && "$1" == signoff && "$2" == --commit ]]
[[ "$(git rev-parse HEAD)" == "$3" ]]
if git symbolic-ref --quiet HEAD; then exit 1; fi
printf '%s\n' "$3" >"${NCLY_TEMPLATE_SIGNOFF}"
SCRIPT
	chmod 755 "${scratch}/signoff-bin/gh"
	# Signoff runs just finish first, so it signs no commit before the loop of the skill is done (#413).
	fct_signoff "${repo}" "${scratch}" || status=$?
	if [[ "${status}" == 0 ]] || ! grep -qF 'this line of the form is not filled' "${scratch}/signoff.out"; then
		cat "${scratch}/signoff.out" >&2
		fct_fail 'just signoff signs a commit that just finish refuses'
	fi
	[[ ! -e "${scratch}/signed-sha" ]] || fct_fail 'just signoff called gh signoff before just finish passed'

	cp -R "${repo}" "${ready}"
	fct_steps "${ready}"
	for file in vision.md spec.md qa.md; do
		grep -v '^Fill:' "${ready}/${file}" >"${scratch}/form.md"
		mv "${scratch}/form.md" "${ready}/${file}"
	done
	git -C "${ready}" commit -qam 'fill the forms'
	finish="${ready}" fct_readiness "$(git -C "${ready}" rev-parse HEAD)" Ready None.
	git -C "${ready}" add readiness.md
	git -C "${ready}" commit -qm 'readiness review'
	sha="$(git -C "${ready}" rev-parse HEAD)"
	if ! fct_signoff "${ready}" "${scratch}"; then
		cat "${scratch}/signoff.out" >&2
		fct_fail 'rendered just signoff failed'
	fi
	fct_has "${scratch}/signoff.out" "Signed off on ${sha}"
	fct_has "${scratch}/signed-sha" "${sha}"
}

# fct_signoff runs just signoff in the repository $1, with the stub gh of the scratch folder $2.
fct_signoff() {
	(cd "${1}" && PATH="${2}/signoff-bin:${PATH}" XDG_CACHE_HOME="${2}/signoff-cache" \
		NCLY_TEMPLATE_SIGNOFF="${2}/signed-sha" just signoff) >"${2}/signoff.out" 2>&1
}
