#!/usr/bin/env bash
set -Eeuo pipefail

fct_main() {
	local repo_root test_dir caller before after foreign
	repo_root="$(cd "${BASH_SOURCE[0]%/*}/.." && pwd -P)"
	test_dir="$(mktemp -d "${TMPDIR:-/tmp}/ncly-hook-test.XXXXXXXX")"
	readonly NCLY_HOOK_TEST_ROOT="${test_dir}"
	trap 'rm -rf "${NCLY_HOOK_TEST_ROOT}"' EXIT
	caller="${test_dir}/caller"
	mkdir -p "${caller}/tools" "${caller}/scripts" "${test_dir}/bin"
	git -c init.templateDir= init --quiet "${caller}"
	printf 'caller data\n' >"${caller}/sentinel.txt"
	git -C "${caller}" add sentinel.txt
	git -C "${caller}" -c core.hooksPath=/dev/null -c user.name=Test -c user.email=test@example.com commit --quiet -m caller
	cp "${repo_root}/justfile" "${caller}/justfile"
	printf '#!/usr/bin/env bash\nexit 0\n' >"${caller}/scripts/release-check-test.sh"
	cp "${caller}/scripts/release-check-test.sh" "${caller}/scripts/test-hook-env.sh"
	cat >"${test_dir}/bin/go" <<'SCRIPT'
#!/usr/bin/env bash
set -Eeuo pipefail
foreign="$(mktemp -d "${NCLY_HOOK_TEST_WORK}/foreign.XXXXXXXX")"
cd "${foreign}"
git -c init.templateDir= init --quiet
printf 'foreign data\n' >marker.txt
git add marker.txt
git -c core.hooksPath=/dev/null -c user.name=Test -c user.email=test@example.com commit --quiet -m foreign
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
