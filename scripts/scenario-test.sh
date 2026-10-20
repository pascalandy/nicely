#!/usr/bin/env bash
# scenario-test.sh proves that just scenario refuses a selection that would run
# no scenario or a wrong count, before go test starts. See #95 and #207.
set -Eeuo pipefail
trap 'printf "%s: line %s failed: %s\n" "${BASH_SOURCE[0]##*/}" "${LINENO}" "${BASH_COMMAND}" >&2' ERR

# fct_refuses runs scenario.sh with the arguments after message. It requires
# exit 2, message on stderr, an empty stdout, and no call of go.
fct_refuses() {
	local message="${1}" status=0
	shift
	bash "${script}" "$@" >"${test_root}/stdout" 2>"${test_root}/stderr" || status=$?
	if [[ "${status}" != 2 || -s "${test_root}/stdout" || -e "${test_root}/calls" ]] || ! grep -qF -- "${message}" "${test_root}/stderr"; then
		cat "${test_root}/stdout" "${test_root}/stderr" >&2
		[[ ! -e "${test_root}/calls" ]] || cat "${test_root}/calls" >&2
		printf 'scenario.sh %s: expected exit 2 and "%s" before any go test, got exit %s\n' "${*:-with no argument}" "${message}" "${status}" >&2
		exit 1
	fi
}

fct_main() {
	local script test_root
	script="$(cd "${BASH_SOURCE[0]%/*}" && pwd -P)/scenario.sh"
	test_root="$(mktemp -d "${TMPDIR:-/tmp}/ncly-scenario-test.XXXXXXXX")"
	readonly NCLY_SCENARIO_TEST_ROOT="${test_root}"
	trap 'rm -rf "${NCLY_SCENARIO_TEST_ROOT}"' EXIT
	mkdir "${test_root}/bin"
	# The stub go records each call, so a guard that is gone starts no real go test.
	cat >"${test_root}/bin/go" <<'SCRIPT'
#!/usr/bin/env bash
printf 'go %s\n' "$*" >>"${NCLY_TEST_CALLS}"
SCRIPT
	chmod 755 "${test_root}/bin/go"
	export PATH="${test_root}/bin:${PATH}" NCLY_TEST_CALLS="${test_root}/calls"
	# Without a name, bash 5 runs no scenario and exits 0, the false green of #91.
	fct_refuses 'name at least one scenario'
	fct_refuses 'no scenario named "no_such_scenario"' no_such_scenario
	fct_refuses 'no scenario named "core/no_such_scenario"' core/no_such_scenario
	fct_refuses '--count takes a whole number above 0, not "0"' extension_cancel --count 0
	printf 'just scenario refuses each selection that would run no scenario or a wrong count\n'
}

fct_main
exit 0
