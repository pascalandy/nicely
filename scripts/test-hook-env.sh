#!/usr/bin/env bash
set -Eeuo pipefail
trap 'printf "%s: line %s failed: %s\n" "${BASH_SOURCE[0]##*/}" "${LINENO}" "${BASH_COMMAND}" >&2' ERR

fct_main() {
	local repo_root test_dir caller before after foreign
	repo_root="$(cd "${BASH_SOURCE[0]%/*}/.." && pwd -P)"
	test_dir="$(mktemp -d "${TMPDIR:-/tmp}/ncly-hook-test.XXXXXXXX")"
	readonly NCLY_HOOK_TEST_ROOT="${test_dir}"
	trap 'rm -rf "${NCLY_HOOK_TEST_ROOT}"' EXIT
	caller="${test_dir}/caller"
	mkdir -p "${caller}/tools" "${caller}/scripts" "${test_dir}/bin"
	printf '[commit]\n gpgsign = true\n[gpg]\n program = %s/missing-signer\n' "${test_dir}" >"${test_dir}/signing.cfg"
	export GIT_CONFIG_GLOBAL="${test_dir}/signing.cfg" GIT_CONFIG_NOSYSTEM=1
	git -c init.templateDir= init --quiet "${caller}"
	printf 'caller data\n' >"${caller}/sentinel.txt"
	git -C "${caller}" add sentinel.txt
	git -C "${caller}" -c core.hooksPath=/dev/null -c commit.gpgsign=false -c user.name=Test -c user.email=test@example.com commit --quiet -m caller
	cp "${repo_root}/justfile" "${caller}/justfile"
	cp "${repo_root}/scripts/release-check-test.sh" "${caller}/scripts/release-check-test.sh"
	cp "${repo_root}/scripts/release-check.sh" "${caller}/scripts/release-check.sh"
	# The nested just test runs each line of the recipe, with the stub go below. A script test gets a stub
	# when it needs the real go, a file that this fixture does not copy, or would run this file again.
	# The outer just test runs it for real. release-verify-test.sh needs the real go. check-layers-test.sh
	# and signoff-test.sh copy check-layers.sh and signoff.sh. scenario-test.sh runs scenario.sh, which this
	# fixture does not copy. test-hook-env.sh is this file.
	# release-check-test.sh runs here for real, since it stubs go itself. See #96 and #230.
	printf '#!/usr/bin/env bash\nexit 0\n' >"${caller}/scripts/test-hook-env.sh"
	printf '#!/usr/bin/env bash\nexit 0\n' >"${caller}/scripts/check-layers-test.sh"
	printf '#!/usr/bin/env bash\nexit 0\n' >"${caller}/scripts/release-verify-test.sh"
	printf '#!/usr/bin/env bash\nexit 0\n' >"${caller}/scripts/signoff-test.sh"
	printf '#!/usr/bin/env bash\nexit 0\n' >"${caller}/scripts/scenario-test.sh"
	cat >"${test_dir}/bin/go" <<'SCRIPT'
#!/usr/bin/env bash
set -Eeuo pipefail
foreign="$(mktemp -d "${NCLY_HOOK_TEST_WORK}/foreign.XXXXXXXX")"
cd "${foreign}"
git -c init.templateDir= init --quiet
printf 'foreign data\n' >marker.txt
git add marker.txt
git -c core.hooksPath=/dev/null -c commit.gpgsign=false -c user.name=Test -c user.email=test@example.com commit --quiet -m foreign
SCRIPT
	chmod 755 "${test_dir}/bin/go"
	before="$(git -C "${caller}" rev-parse HEAD)"
	git -C "${caller}" ls-files --stage >"${test_dir}/index.before"
	env "PATH=${test_dir}/bin:${PATH}" "NCLY_HOOK_TEST_WORK=${test_dir}" "GIT_DIR=${caller}/.git" "GIT_IMPLICIT_WORK_TREE=0" \
		just --justfile "${caller}/justfile" --working-directory "${caller}" test
	after="$(git --git-dir="${caller}/.git" rev-parse HEAD)"
	[[ "${before}" == "${after}" ]]
	git -C "${caller}" ls-files --stage >"${test_dir}/index.after"
	cmp "${test_dir}/index.before" "${test_dir}/index.after"
	[[ $(git -C "${caller}" config core.bare) == false ]]
	[[ $(cat "${caller}/sentinel.txt") == 'caller data' ]]
	for foreign in "${test_dir}"/foreign.*; do
		[[ -d "${foreign}/.git" ]]
		[[ $(git -C "${foreign}" show HEAD:marker.txt) == 'foreign data' ]]
	done
	printf 'Git hook environment cannot change the caller repository\n'
}

fct_main
exit 0
