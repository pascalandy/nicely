#!/usr/bin/env bash
# Sourced by template-test.sh to prove just signoff of the rendered
# repository. fct_test_uninstall calls it after its install tests.

# Run the rendered recipe with real Git and just. Only gh is a stub.
fct_test_signoff() {
	local repo="${1}" scratch="${2}" sha
	sha="$(git -C "${repo}" rev-parse HEAD)"
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
	if ! (cd "${repo}" && PATH="${scratch}/signoff-bin:${PATH}" XDG_CACHE_HOME="${scratch}/signoff-cache" \
		NCLY_TEMPLATE_SIGNOFF="${scratch}/signed-sha" just signoff) >"${scratch}/signoff.out" 2>&1; then
		cat "${scratch}/signoff.out" >&2
		fct_fail 'rendered just signoff failed'
	fi
	fct_has "${scratch}/signoff.out" "Signed off on ${sha}"
	fct_has "${scratch}/signed-sha" "${sha}"
}
