#!/usr/bin/env bash
# Prove isolated signoff through core's recipe, with real Git and just.
# gh is always a stub. A dirty fix must not make a broken commit pass.
set -Eeuo pipefail
trap 'printf "%s: line %s failed: %s\n" "${BASH_SOURCE[0]##*/}" "${LINENO}" "${BASH_COMMAND}" >&2' ERR

root=""
test_case="setup"

fct_cleanup() {
	# Keep artifacts when trash is absent, such as on a hosted runner.
	if command -v trash >/dev/null; then trash "${root}"; fi
}

fct_fail() {
	cat "${root}/output" >&2
	printf 'signoff-test: %s: %s\n' "${test_case}" "$*" >&2
	exit 1
}

fct_setup() {
	local repo_root variable
	for variable in $(compgen -e); do
		case "${variable}" in GIT_* | JUST_*) unset "${variable}" ;; esac
	done
	repo_root="$(cd "${BASH_SOURCE[0]%/*}/.." && pwd -P)"
	root="$(mktemp -d "${TMPDIR:-/tmp}/ncly-signoff-test.XXXXXXXX")"
	mkdir -p "${root}/bin" "${root}/caller/scripts" "${root}/caller/signoff-fixture" "${root}/cache"
	export NCLY_SIGNOFF_TEST="${root}" PATH="${root}/bin:${PATH}" XDG_CACHE_HOME="${root}/cache"
	cat >"${root}/bin/gh" <<'SCRIPT'
#!/usr/bin/env bash
set -Eeuo pipefail
case "$*" in
--version)
	if git symbolic-ref --quiet HEAD >/dev/null; then
		echo 'version probe is not isolated' >&2
		exit 1
	fi
	echo 'gh version test'
	exit 0
	;;
signoff | signoff\ --commit\ *) ;;
*) echo "gh stub refuses: $*" >&2; exit 1 ;;
esac
if [[ "$*" == signoff\ --commit\ * ]]; then
	jq -e --arg sha "${3}" '.sha == $sha and .checks[0].exit == 0 and .signoff == null' ../receipt.json >/dev/null
fi
printf '%s\n' "$*" >>"${NCLY_SIGNOFF_TEST}/signed"
git rev-parse HEAD >>"${NCLY_SIGNOFF_TEST}/signed-head"
[[ ! -e "${NCLY_SIGNOFF_TEST}/gh-fails" ]]
SCRIPT
	chmod 755 "${root}/bin/gh"
	cp "${repo_root}/justfile" "${root}/caller/justfile"
	cp "${repo_root}/go.mod" "${repo_root}/go.sum" "${root}/caller/"
	cp -R "${repo_root}/sdk" "${root}/caller/sdk"
	cp "${repo_root}/scripts/signoff.sh" "${root}/caller/scripts/signoff.sh"
	cat >>"${root}/caller/justfile" <<'JUST'

set allow-duplicate-recipes

check: signoff-test-check

signoff-test-check:
    @git rev-parse HEAD >>"${NCLY_SIGNOFF_TEST}/checked"
    @test -z "${GIT_DIR-}${GIT_WORK_TREE-}${GIT_SIGNOFF_TEST-}${JUST_SIGNOFF_TEST-}${JUST_NO_DEPS-}"
    @test "$(cat fix.txt)" = fixed || { echo COMMITTED-FAILURE >&2; exit 1; }
    @go test -count=1 ./signoff-fixture
    @go env GOPRIVATE >"${NCLY_SIGNOFF_TEST}/goprivate"
    @if [[ -e "${NCLY_SIGNOFF_TEST}/move-head" ]]; then git checkout --quiet --detach HEAD~; fi
    @if [[ -e "${NCLY_SIGNOFF_TEST}/change-tracked" ]]; then echo changed >fix.txt; fi
    @if [[ -e "${NCLY_SIGNOFF_TEST}/move-caller" ]]; then git -C "${NCLY_SIGNOFF_TEST}/caller" -c commit.gpgsign=false commit --quiet --allow-empty --message 'caller moved'; fi
JUST
	printf 'package fixture\n\nfunc fixed() bool { return false }\n' >"${root}/caller/signoff-fixture/fix.go"
	cat >"${root}/caller/signoff-fixture/fix_test.go" <<'GO'
package fixture

import "testing"

func TestFix(t *testing.T) {
	if !fixed() {
		t.Fatal("COMMITTED-FAILURE")
	}
}
GO
	git -C "${root}/caller" init -q
	git -C "${root}/caller" config user.name 'Signoff test'
	git -C "${root}/caller" config user.email 'test@example.com'
	git -C "${root}/caller" config core.hooksPath /dev/null
	git -C "${root}/caller" config commit.gpgsign false
	printf 'broken\n' >"${root}/caller/fix.txt"
	git -C "${root}/caller" add .
	git -C "${root}/caller" commit -qm 'broken commit'
	sha="$(git -C "${root}/caller" rev-parse HEAD)"
	: >"${root}/output"
	: >"${root}/signed"
	: >"${root}/signed-head"
}

# Run core's recipe and check the observable result, including the checked SHA.
fct_run() {
	local outcome="${1}" status=0
	shift
	: >"${root}/checked"
	: >"${root}/signed"
	: >"${root}/signed-head"
	(cd "${root}/caller" && "$@") >"${root}/output" 2>&1 </dev/null || status=$?
	if [[ "${outcome}" == passes && "${status}" != 0 || "${outcome}" == fails && "${status}" == 0 ]]; then
		fct_fail "$* exited ${status}, expected ${outcome}"
	fi
	[[ "$(cat "${root}/checked")" == "${sha}" ]] || fct_fail 'check did not run once on the frozen SHA'
}

fct_main() {
	fct_setup
	trap fct_cleanup EXIT
	test_case="dirty fix"
	printf 'fixed\n' >"${root}/caller/fix.txt"
	printf 'package fixture\n\nfunc fixed() bool { return true }\n' >"${root}/caller/signoff-fixture/fix.go"
	(cd "${root}/caller" && just check) >"${root}/output" 2>&1
	fct_run fails just signoff
	grep -qF COMMITTED-FAILURE "${root}/output" || fct_fail 'the committed defect did not fail the check'
	[[ ! -s "${root}/signed" ]] || fct_fail 'signed the unfixed commit'
	fct_receipt 1 null

	git -C "${root}/caller" add fix.txt
	git -C "${root}/caller" commit -qm 'committed fix'
	sha="$(git -C "${root}/caller" rev-parse HEAD)"
	test_case="dirty Go fix without overlay"
	fct_run fails just signoff
	grep -qF COMMITTED-FAILURE "${root}/output" || fct_fail 'the committed Go defect did not fail the check'
	[[ ! -s "${root}/signed" ]] || fct_fail 'signed the unfixed Go commit'
	fct_receipt 1 null
	jq -n --arg fixed "${root}/caller/signoff-fixture/fix.go" '{Replace:{"signoff-fixture/fix.go":$fixed}}' >"${root}/overlay.json"
	printf 'GOFLAGS=-overlay=%s/overlay.json\n' "${root}" >"${root}/go-env"
	local form
	for form in GOFLAGS GOENV; do
		test_case="dirty Go fix through ${form}"
		case "${form}" in
		GOFLAGS) fct_run fails env GOFLAGS="-overlay=${root}/overlay.json" just signoff ;;
		GOENV) fct_run fails env GOENV="${root}/go-env" just signoff ;;
		esac
		grep -qF COMMITTED-FAILURE "${root}/output" || fct_fail 'the overlay bypassed the committed Go defect'
		[[ ! -s "${root}/signed" ]] || fct_fail 'signed the overlay-fixed commit'
		fct_receipt 1 null
	done
	git -C "${root}/caller" add signoff-fixture/fix.go
	git -C "${root}/caller" commit -qm 'committed Go fix'
	sha="$(git -C "${root}/caller" rev-parse HEAD)"
	printf 'dirty caller\n' >"${root}/caller/fix.txt"
	printf 'untracked caller\n' >"${root}/caller/untracked.txt"
	git -C "${root}/caller" diff >"${root}/before"
	mkdir "${root}/decoy"
	git -C "${root}/decoy" init -q
	printf 'check:\n    @exit 1\n' >"${root}/decoy/justfile"
	printf 'GOPRIVATE=example.com/private\n' >"${root}/go-env-private"
	for form in normal no-deps-env no-deps-flag caller-env caller-go-env; do
		test_case="passing commit, ${form}"
		case "${form}" in
		normal) fct_run passes just signoff ;;
		no-deps-env) fct_run passes env JUST_NO_DEPS=true just signoff ;;
		no-deps-flag) fct_run passes just --no-deps signoff ;;
		caller-env)
			fct_run passes env GIT_DIR="${root}/decoy/.git" GIT_WORK_TREE="${root}/decoy" GIT_SIGNOFF_TEST=leak \
				JUST_SIGNOFF_TEST=leak JUST_NO_DEPS=true JUST_JUSTFILE="${root}/decoy/justfile" JUST_WORKING_DIRECTORY="${root}/decoy" \
				just --justfile "${root}/caller/justfile" --working-directory "${root}/caller" signoff
			;;
		caller-go-env)
			fct_run passes env GOENV="${root}/go-env-private" just signoff
			[[ "$(cat "${root}/goprivate")" == example.com/private ]] || fct_fail 'signoff dropped a persisted Go setting'
			;;
		esac
		[[ "$(cat "${root}/signed")" == "signoff --commit ${sha}" ]] || fct_fail 'gh did not receive the frozen full SHA'
		[[ "$(cat "${root}/signed-head")" == "${sha}" ]] || fct_fail 'gh ran outside the frozen checkout'
		grep -qxF "Signed off on ${sha}" "${root}/output" || fct_fail 'missing landing-script output'
		fct_receipt 0 0
	done
	git -C "${root}/caller" diff >"${root}/after"
	cmp -s "${root}/before" "${root}/after" || fct_fail 'signoff changed the caller'
	[[ "$(cat "${root}/caller/untracked.txt")" == 'untracked caller' ]] || fct_fail 'signoff removed caller files'
	[[ "$(git -C "${root}/caller" worktree list --porcelain | grep -c '^worktree ')" == 1 ]] || fct_fail 'scratch worktree remains'

	test_case="caller HEAD moves during the check"
	: >"${root}/move-caller"
	fct_run passes just signoff
	[[ "$(cat "${root}/signed")" == "signoff --commit ${sha}" ]] || fct_fail 'signed the moved caller HEAD'
	fct_receipt 0 0
	[[ "$(git -C "${root}/caller" rev-parse HEAD)" != "${sha}" ]] || fct_fail 'caller HEAD did not move'
	sha="$(git -C "${root}/caller" rev-parse HEAD)"
	mv "${root}/move-caller" "${root}/caller-moved"

	test_case="gh failure"
	: >"${root}/gh-fails"
	fct_run fails just signoff
	if grep -q '^Signed off on ' "${root}/output"; then fct_fail 'reported success after gh failed'; fi
	fct_receipt 0 1
	mv "${root}/gh-fails" "${root}/gh-failed"
	for form in move-head change-tracked; do
		test_case="${form} during the check"
		: >"${root}/${form}"
		fct_run fails just signoff
		[[ ! -s "${root}/signed" ]] || fct_fail 'signed a changed checkout'
		grep -q 'during the check' "${root}/output" || fct_fail 'failed before the checkout guard'
		mv "${root}/${form}" "${root}/${form}-done"
	done
	printf 'signoff: dirty fixes fail, frozen SHAs pass, caller files stay, and failed gates sign nothing\n'
}

fct_receipt() {
	local receipt
	receipt="$(sed -n 's/^signoff: receipt in //p' "${root}/output")"
	[[ "${receipt}" == "${root}/cache/nicely/signoff/"*/receipt.json ]] || fct_fail 'receipt is not outside the repository'
	jq -e --arg sha "${sha}" --argjson check "${1}" --argjson signoff "${2}" '
        .sha == $sha and (.os == "linux" or .os == "darwin") and
        ([.tools.git, .tools.go, .tools.just, .tools.gh] | all(type == "string" and length > 0)) and
        (.checks | length == 1) and .checks[0].command[0] == "just" and .checks[0].exit == $check and
        (if $signoff == null then .signoff == null
         else .signoff.command == ["gh", "signoff", "--commit", $sha] and .signoff.exit == $signoff end)
    ' "${receipt}" >/dev/null || fct_fail 'receipt does not describe the actual run'
	[[ -s "${receipt%/*}/check.log" ]] || fct_fail 'receipt has no check log'
}

fct_main
