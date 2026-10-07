#!/usr/bin/env bash
set -Eeuo pipefail

fct_fails_with() {
	local expected="${1}" status=0
	shift
	"$@" >"${test_root}/stdout" 2>"${test_root}/stderr" || status=$?
	if [[ "${status}" != "${expected}" ]]; then
		cat "${test_root}/stderr" >&2
		printf 'Expected exit %s, got %s\n' "${expected}" "${status}" >&2
		exit 1
	fi
}

fct_main() {
	local script test_root scratch
	script="$(cd "${BASH_SOURCE[0]%/*}" && pwd -P)/release-check.sh"
	test_root="$(mktemp -d "${TMPDIR:-/tmp}/ncly-release-test.XXXXXXXX")"
	readonly NCLY_RELEASE_TEST_ROOT="${test_root}"
	trap 'rm -rf "${NCLY_RELEASE_TEST_ROOT}"' EXIT
	mkdir -p "${test_root}/repo" "${test_root}/bin"
	cat >"${test_root}/bin/just" <<'SCRIPT'
#!/usr/bin/env bash
printf 'just\n' >>"${NCLY_TEST_CALLS}"
exit 1
SCRIPT
	cp "${test_root}/bin/just" "${test_root}/bin/go"
	chmod 755 "${test_root}/bin/just" "${test_root}/bin/go"
	export PATH="${test_root}/bin:${PATH}" NCLY_TEST_CALLS="${test_root}/calls" TMPDIR="${test_root}"
	cd "${test_root}/repo"
	git -c init.templateDir= init --quiet
	printf 'original\n' >source.txt
	git add source.txt
	git -c core.hooksPath=/dev/null -c user.name=Test -c user.email=test@example.com commit --quiet -m fixture
	fct_fails_with 2 bash "${script}" v01.2.3
	[[ ! -e "${test_root}/calls" ]]
	printf 'new\n' >untracked.txt
	fct_fails_with 1 bash "${script}" v0.0.1
	[[ ! -e "${test_root}/calls" ]]
	rm untracked.txt
	printf 'working change\n' >source.txt
	fct_fails_with 1 bash "${script}" v0.0.1
	[[ $(cat "${test_root}/calls") == just ]]
	scratch="$(sed -n 's/^Release rehearsal files retained at //p' "${test_root}/stderr")"
	cmp source.txt "${scratch}/tree/source.txt"
	[[ $(git show HEAD:source.txt) == original && -z $(git tag --list) ]]
	rm -rf "${scratch}"
	printf 'Release-check validation and working-source capture passed\n'
}

fct_main
exit 0
