#!/usr/bin/env bash
# signoff-test.sh proves just signoff for just test: it runs the check once,
# then gh signoff, even under JUST_NO_DEPS=true or --no-deps, and it signs
# nothing after a failed check. It runs on macOS and Linux, with bash 3.2.
#
# The caller's justfile is a copy of the real one, where only check is a
# stub. The stub writes CHECK-RAN from a dependency of check, because
# JUST_NO_DEPS and --no-deps skip exactly those. gh is a stub too, so no case
# reaches the network. See #213.
set -Eeuo pipefail
trap 'printf "%s: line %s failed: %s\n" "${BASH_SOURCE[0]##*/}" "${LINENO}" "${BASH_COMMAND}" >&2' ERR

SIGNOFF_TEST_ROOT=""
test_case="setup"

# shellcheck disable=SC2317,SC2329
fct_cleanup() {
	if [[ -n "${SIGNOFF_TEST_ROOT}" ]]; then
		rm -rf "${SIGNOFF_TEST_ROOT}"
	fi
}

fct_fail() {
	printf 'signoff-test: %s: %s\n' "${test_case}" "$*" >&2
	exit 1
}

fct_setup() {
	local repo_root
	repo_root="$(cd "${BASH_SOURCE[0]%/*}/.." && pwd -P)"
	root="$(mktemp -d "${TMPDIR:-/tmp}/ncly-signoff-test.XXXXXXXX")"
	SIGNOFF_TEST_ROOT="${root}"
	mkdir -p "${root}/bin" "${root}/caller/scripts"
	export NCLY_SIGNOFF_TEST="${root}" PATH="${root}/bin:${PATH}"
	cat >"${root}/bin/gh" <<'SCRIPT'
#!/usr/bin/env bash
set -Eeuo pipefail
if [[ "$*" != signoff ]]; then
	echo "gh stub: no answer for: $*" >&2
	exit 1
fi
echo SIGNOFF-RAN >>"${NCLY_SIGNOFF_TEST}/log"
SCRIPT
	chmod 755 "${root}/bin/gh"
	cp "${repo_root}/justfile" "${root}/caller/justfile"
	cat >>"${root}/caller/justfile" <<'JUST'

set allow-duplicate-recipes

# The stub check of scripts/signoff-test.sh replaces the real one, and only lists a dependency, as the real one does
check: signoff-test-check

signoff-test-check:
    @echo CHECK-RAN >>"${NCLY_SIGNOFF_TEST}/log"
    @test ! -e "${NCLY_SIGNOFF_TEST}/fail"
JUST
	cp "${repo_root}/scripts/signoff.sh" "${root}/caller/scripts/signoff.sh"
}

# fct_expect <passes|fails> <log line>... runs the three forms of just signoff, and requires each to end as given and to write exactly these lines, in this order.
fct_expect() {
	local outcome="${1}" form status
	shift
	printf '%s\n' "$@" >"${root}/log.expected"
	for form in "just signoff" "env JUST_NO_DEPS=true just signoff" "just --no-deps signoff"; do
		: >"${root}/log"
		status=0
		# shellcheck disable=SC2086
		(cd "${root}/caller" && ${form}) >"${root}/output" 2>&1 </dev/null || status=$?
		if [[ "${outcome}" == passes && "${status}" != 0 || "${outcome}" == fails && "${status}" == 0 ]]; then
			cat "${root}/output" >&2
			fct_fail "${form} exited ${status}"
		fi
		if ! cmp -s "${root}/log.expected" "${root}/log"; then
			diff "${root}/log.expected" "${root}/log" >&2 || true
			fct_fail "${form} ran other steps"
		fi
	done
}

fct_main() {
	trap fct_cleanup EXIT
	fct_setup
	test_case="passing check"
	fct_expect passes CHECK-RAN SIGNOFF-RAN
	test_case="failing check"
	: >"${root}/fail"
	fct_expect fails CHECK-RAN
	printf 'just signoff runs the check once before gh signoff, under JUST_NO_DEPS and --no-deps too, and signs nothing after a failed check\n'
}

fct_main
exit 0
