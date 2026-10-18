#!/usr/bin/env bash
# check-layers-test.sh proves just check-layers and just signoff-layers on a
# scratch stack, for just test. It runs on macOS and Linux, with bash 3.2.
#
# Each case clones a bare origin into a new caller and runs the caller's
# justfile: a copy of the real one, where only check is a stub. gh is a stub
# too, so no case reaches the network. To prove that a case catches a broken
# guard, copy the justfile and scripts/ to a scratch folder. Break the guard in
# the copy of check-layers.sh. Run the copy of this test. See #58 and #71.
set -Eeuo pipefail
trap 'printf "%s: line %s failed: %s\n" "${BASH_SOURCE[0]##*/}" "${LINENO}" "${BASH_COMMAND}" >&2' ERR
# A hook or a caller can export GIT_DIR. The scratch repositories must see their own.
while IFS= read -r git_variable; do unset "${git_variable}"; done < <(git rev-parse --local-env-vars)

LAYERS_TEST_ROOT=""
LAYERS_TEST_GROUP=""
test_case="setup"

# shellcheck disable=SC2317,SC2329
fct_cleanup() {
	if [[ -n "${LAYERS_TEST_GROUP}" ]]; then
		kill -KILL -- "-${LAYERS_TEST_GROUP}" 2>/dev/null || true
	fi
	if [[ -n "${LAYERS_TEST_ROOT}" ]]; then
		rm -rf "${LAYERS_TEST_ROOT}"
	fi
}

fct_fail() {
	printf 'check-layers-test: %s: %s\n' "${test_case}" "$*" >&2
	exit 1
}

# fct_wait_for <what> <command...> runs the command until it succeeds, for at most 30 seconds.
fct_wait_for() {
	local what="${1}" tries=0
	shift
	until "$@"; do
		tries=$((tries + 1))
		((tries < 300)) || fct_fail "timed out waiting for ${what}"
		sleep 0.1
	done
}

# shellcheck disable=SC2329
fct_gone() {
	! kill -0 "${1}" 2>/dev/null
}

# fct_layer <branch> <base> <file> <content> commits the file on a new branch of the seed, pushes it, and sets layer_sha.
fct_layer() {
	local seed="${root}/seed"
	git -C "${seed}" switch --quiet --create "${1}" "${2}"
	printf '%s\n' "${4}" >"${seed}/${3}"
	git -C "${seed}" add "${3}"
	git -C "${seed}" commit --quiet --message "${1}"
	git -C "${seed}" push --quiet origin "${1}"
	layer_sha="$(git -C "${seed}" rev-parse HEAD)"
}

# fct_caller <name> clones origin into a new caller, edits a tracked file, adds an untracked one, and records its state.
fct_caller() {
	caller="${root}/${1}"
	git clone --quiet "${root}/origin.git" "${caller}"
	git -C "${caller}" config fetch.prune false
	rm -f "${root}/signed" "${root}/signoff-fails"
	printf 'caller change\n' >>"${caller}/README.txt"
	printf 'caller file\n' >"${caller}/untracked.txt"
	fct_snapshot >"${root}/caller.before"
}

# fct_snapshot prints what a run must not change in the caller: its branch, HEAD, index, files, and worktrees.
fct_snapshot() {
	git -C "${caller}" symbolic-ref HEAD
	git -C "${caller}" rev-parse HEAD
	git -C "${caller}" ls-files --stage
	git -C "${caller}" status --porcelain --untracked-files=all
	git -C "${caller}" diff
	cat "${caller}/untracked.txt"
	git -C "${caller}" worktree list --porcelain
}

fct_caller_unchanged() {
	fct_snapshot >"${root}/caller.after"
	if ! cmp -s "${root}/caller.before" "${root}/caller.after"; then
		diff "${root}/caller.before" "${root}/caller.after" >&2 || true
		fct_fail "the run changed the caller's checkout"
	fi
}

# fct_run <expected exit> <command...> runs the command in the caller, and sets run to the folder of its logs.
fct_run() {
	local expected="${1}" status=0
	shift
	: >"${root}/checks"
	: >"${root}/gh-calls"
	(cd "${caller}" && "$@") >"${root}/stdout" 2>"${root}/stderr" </dev/null || status=$?
	if [[ "${status}" != "${expected}" ]]; then
		cat "${root}/stdout" "${root}/stderr" >&2
		fct_fail "expected exit ${expected}, got ${status}"
	fi
	run="$(sed -n 's/^[a-z-]*-layers: logs in //p' "${root}/stderr")"
}

fct_same() {
	if ! cmp -s "${1}" "${2}"; then
		diff "${1}" "${2}" >&2 || true
		fct_fail "${3}"
	fi
}

# fct_expect_checks <sha>... requires one stub check per SHA, in that order, each in the run's worktree and without files of an earlier layer.
fct_expect_checks() {
	local sha
	: >"${root}/checks.expected"
	for sha in "$@"; do printf '%s %s/worktree clean\n' "${sha}" "${run}" >>"${root}/checks.expected"; done
	fct_same "${root}/checks.expected" "${root}/checks" "the checks did not run on the expected commits"
}

# fct_expect_verdicts <branch> <sha> <verdict>... requires these verdicts, in that order, on the terminal and in summary.txt.
fct_expect_verdicts() {
	: >"${root}/stdout.expected"
	: >"${root}/summary.expected"
	while (($# > 0)); do
		printf '%s %s %s\n' "${1}" "$(git -C "${caller}" rev-parse --short "${2}")" "${3}" >>"${root}/stdout.expected"
		printf '%s %s %s\n' "${1}" "${2}" "${3}" >>"${root}/summary.expected"
		shift 3
	done
	fct_same "${root}/stdout.expected" "${root}/stdout" "check-layers printed other verdicts"
	fct_same "${root}/summary.expected" "${run}/summary.txt" "summary.txt holds other verdicts"
}

# fct_expect_gh <call>... requires exactly these calls to the gh stub, in that order.
fct_expect_gh() {
	: >"${root}/gh-calls.expected"
	if (($# > 0)); then printf '%s\n' "$@" >"${root}/gh-calls.expected"; fi
	fct_same "${root}/gh-calls.expected" "${root}/gh-calls" "gh got other calls"
}

# fct_expect_signed_summary requires one sign-off per line of summary.txt, with its SHA, in the same order, each SHA an ancestor of the next.
fct_expect_signed_summary() {
	local branch sha verdict previous=""
	: >"${root}/gh-calls.expected"
	while read -r branch sha verdict; do
		[[ "${verdict}" == ok ]] || fct_fail "summary.txt gives ${branch} the verdict ${verdict}"
		printf 'signoff --commit %s\n' "${sha}" >>"${root}/gh-calls.expected"
		if [[ -n "${previous}" ]] && ! git -C "${caller}" merge-base --is-ancestor "${previous}" "${sha}"; then
			fct_fail "${sha} of ${branch} does not contain ${previous}"
		fi
		previous="${sha}"
	done <"${run}/summary.txt"
	fct_same "${root}/gh-calls.expected" "${root}/gh-calls" "the sign-offs do not match summary.txt"
}

fct_expect_no_worktree() {
	[[ -n "${run}" && ! -e "${run}/worktree" ]] || fct_fail "the worktree of the run remains"
}

fct_setup() {
	local repo_root
	repo_root="$(cd "${BASH_SOURCE[0]%/*}/.." && pwd -P)"
	root="$(mktemp -d "${TMPDIR:-/tmp}/ncly-layers-test.XXXXXXXX")"
	LAYERS_TEST_ROOT="${root}"
	# macOS reaches TMPDIR through the /var link, and git prints the resolved path.
	root="$(cd "${root}" && pwd -P)"
	mkdir -p "${root}/bin" "${root}/home" "${root}/seed/scripts"
	cat >"${root}/gitconfig" <<'CONFIG'
[user]
	name = Test
	email = test@example.com
[commit]
	gpgsign = false
[core]
	hooksPath = /dev/null
[init]
	defaultBranch = main
[advice]
	detachedHead = false
CONFIG
	export GIT_CONFIG_GLOBAL="${root}/gitconfig" GIT_CONFIG_NOSYSTEM=1 HOME="${root}/home" XDG_CACHE_HOME="${root}/cache"
	export NCLY_LAYERS_TEST="${root}" PATH="${root}/bin:${PATH}"
	cat >"${root}/check.sh" <<'SCRIPT'
#!/usr/bin/env bash
set -Eeuo pipefail
# A case can move a tracking ref of the caller once, during the first check.
if [[ -e "${NCLY_LAYERS_TEST}/move" ]]; then
	read -r ref target <"${NCLY_LAYERS_TEST}/move"
	git update-ref "${ref}" "${target}"
	rm "${NCLY_LAYERS_TEST}/move"
fi
state=clean
if [[ -e left/over ]]; then state=dirty; fi
printf '%s %s %s\n' "$(git rev-parse HEAD)" "$(pwd -P)" "${state}" >>"${NCLY_LAYERS_TEST}/checks"
mkdir -p left
printf 'left by a layer\n' >left/over
# A case can advance a branch on origin once, during the first check.
if [[ -e "${NCLY_LAYERS_TEST}/advance" ]]; then
	read -r branch target <"${NCLY_LAYERS_TEST}/advance"
	git push --quiet origin "${target}:refs/heads/${branch}"
	rm "${NCLY_LAYERS_TEST}/advance"
fi
# A case can move the worktree's HEAD once, during the first check.
if [[ -e "${NCLY_LAYERS_TEST}/detach" ]]; then
	git checkout --quiet --detach HEAD~1
	rm "${NCLY_LAYERS_TEST}/detach"
fi
if [[ -e "${NCLY_LAYERS_TEST}/block" ]]; then
	pwd -P >"${NCLY_LAYERS_TEST}/started.tmp"
	mv "${NCLY_LAYERS_TEST}/started.tmp" "${NCLY_LAYERS_TEST}/started"
	exec sleep 600
fi
if [[ -f fail.txt && "$(cat fail.txt)" == fail ]]; then
	echo 'stub check failed on purpose'
	exit 1
fi
echo 'stub check passed'
SCRIPT
	cat >"${root}/bin/gh" <<'SCRIPT'
#!/usr/bin/env bash
set -Eeuo pipefail
printf '%s\n' "$*" >>"${NCLY_LAYERS_TEST}/gh-calls"
if [[ "$*" == "stack view --json" ]]; then
	cat "${NCLY_LAYERS_TEST}/stack.json"
	exit 0
fi
# signed holds the green statuses, and signoff-fails the SHAs whose sign-off fails.
if [[ $# -eq 3 && "${1}" == signoff && "${2}" == --commit ]]; then
	if grep -qx "${3}" "${NCLY_LAYERS_TEST}/signoff-fails" 2>/dev/null; then
		echo "gh stub: the sign-off of ${3} failed on purpose" >&2
		exit 1
	fi
	printf '%s\n' "${3}" >>"${NCLY_LAYERS_TEST}/signed"
	echo "gh stub: signed off on ${3}"
	exit 0
fi
if [[ $# -eq 4 && "${1} ${2} ${3}" == "signoff status --commit" ]]; then
	grep -qx "${4}" "${NCLY_LAYERS_TEST}/signed" 2>/dev/null
	echo "gh stub: signoff is green on ${4}"
	exit 0
fi
echo "gh stub: no answer for: $*" >&2
exit 1
SCRIPT
	chmod 755 "${root}/check.sh" "${root}/bin/gh"
	git init --quiet --bare "${root}/origin.git"
	git init --quiet "${root}/seed"
	git -C "${root}/seed" remote add origin "${root}/origin.git"
	cp "${repo_root}/justfile" "${root}/seed/justfile"
	cat >>"${root}/seed/justfile" <<'JUST'

set allow-duplicate-recipes

# The stub check of scripts/check-layers-test.sh replaces the real one, and only lists a dependency, as the real one does
check: layers-test-check

layers-test-check:
    @bash "${NCLY_LAYERS_TEST}/check.sh"
JUST
	cp "${repo_root}/scripts/check-layers.sh" "${root}/seed/scripts/check-layers.sh"
	printf 'scratch stack\n' >"${root}/seed/README.txt"
	git -C "${root}/seed" add .
	git -C "${root}/seed" commit --quiet --message main
	git -C "${root}/seed" push --quiet origin main
	fct_layer one main one.txt one
	sha_one="${layer_sha}"
	fct_layer two one two.txt two
	sha_two="${layer_sha}"
	fct_layer three two three.txt three
	sha_three="${layer_sha}"
	fct_layer next two next.txt next
	sha_next="${layer_sha}"
	fct_layer orphan main orphan.txt orphan
	sha_orphan="${layer_sha}"
	fct_layer above orphan above.txt above
	sha_above="${layer_sha}"
	fct_layer bad main fail.txt fail
	sha_bad="${layer_sha}"
	fct_layer fixed bad fail.txt pass
	sha_fixed="${layer_sha}"
	fct_layer gone main gone.txt gone
	git -C "${root}/seed" switch --quiet --create nojust main
	git -C "${root}/seed" rm --quiet justfile
	git -C "${root}/seed" commit --quiet --message nojust
	git -C "${root}/seed" push --quiet origin nojust
	sha_nojust="$(git -C "${root}/seed" rev-parse HEAD)"
	fct_layer moving main moving.txt first
}

fct_case_two_layers() {
	test_case="two layers"
	fct_caller two-layers
	fct_run 0 just check-layers one two
	fct_expect_checks "${sha_one}" "${sha_two}"
	fct_expect_verdicts one "${sha_one}" ok two "${sha_two}" ok
	fct_expect_gh
	[[ "$(ls "${run}")" == "01-one.log"$'\n'"02-two.log"$'\n'"summary.txt" ]] || fct_fail "expected one log per layer and summary.txt"
	grep -q 'stub check passed' "${run}/01-one.log" || fct_fail "the log of one lacks the check output"
	grep -q 'stub check passed' "${run}/02-two.log" || fct_fail "the log of two lacks the check output"
	fct_expect_no_worktree
	fct_caller_unchanged
}

fct_case_failing_check() {
	test_case="failing check"
	fct_caller failing-check
	fct_run 1 just check-layers bad fixed
	fct_expect_checks "${sha_bad}" "${sha_fixed}"
	fct_expect_verdicts bad "${sha_bad}" FAIL fixed "${sha_fixed}" ok
	fct_expect_gh
	grep -q 'stub check failed on purpose' "${run}/01-bad.log" || fct_fail "the log of bad lacks the check output"
	fct_expect_no_worktree
	fct_caller_unchanged
}

fct_case_coherent_stack() {
	test_case="coherent stack"
	fct_caller coherent-stack
	fct_run 0 just check-layers one two three
	fct_expect_checks "${sha_one}" "${sha_two}" "${sha_three}"
	fct_expect_verdicts one "${sha_one}" ok two "${sha_two}" ok three "${sha_three}" ok
	fct_expect_no_worktree
	fct_caller_unchanged
}

fct_case_missing_parent() {
	test_case="layer without its parent"
	fct_caller missing-parent
	fct_run 1 just check-layers one orphan above
	fct_expect_checks "${sha_one}" "${sha_orphan}" "${sha_above}"
	fct_expect_verdicts one "${sha_one}" ok orphan "${sha_orphan}" FAIL above "${sha_above}" ok
	fct_expect_gh
	grep -q "origin/orphan ${sha_orphan} does not contain origin/one ${sha_one}" "${run}/02-orphan.log" || fct_fail "the log of orphan does not name both layers"
	grep -q 'stub check passed' "${run}/02-orphan.log" || fct_fail "the check of orphan must pass, so that only the ancestry fails it"
	fct_expect_no_worktree
	fct_caller_unchanged
}

fct_case_moved_ref() {
	test_case="tracking ref moved during the run"
	fct_caller moved-ref
	printf 'refs/remotes/origin/two %s\n' "${sha_bad}" >"${root}/move"
	fct_run 0 just check-layers one two three
	[[ ! -e "${root}/move" && "$(git -C "${caller}" rev-parse refs/remotes/origin/two)" == "${sha_bad}" ]] || fct_fail "the stub check did not move origin/two"
	fct_expect_checks "${sha_one}" "${sha_two}" "${sha_three}"
	fct_expect_verdicts one "${sha_one}" ok two "${sha_two}" ok three "${sha_three}" ok
	fct_caller_unchanged
}

fct_case_moved_head() {
	test_case="HEAD moved during the check"
	fct_caller moved-head
	: >"${root}/detach"
	fct_run 1 just check-layers one two
	[[ ! -e "${root}/detach" ]] || fct_fail "the stub check did not move HEAD"
	fct_expect_checks "${sha_one}" "${sha_two}"
	fct_expect_verdicts one "${sha_one}" FAIL two "${sha_two}" ok
	fct_expect_gh
	grep -q 'stub check passed' "${run}/01-one.log" || fct_fail "the check of one must pass, so that only the moved HEAD fails it"
	grep -q 'HEAD moved off origin/one during the check' "${run}/01-one.log" || fct_fail "the log of one does not name the moved HEAD"
	fct_expect_no_worktree
	fct_caller_unchanged
}

fct_case_deleted_branch() {
	test_case="deleted branch"
	fct_caller deleted-branch
	git -C "${root}/seed" push --quiet origin --delete gone
	fct_run 1 just check-layers one gone
	git -C "${caller}" rev-parse --quiet --verify refs/remotes/origin/gone >/dev/null || fct_fail "the caller lost its stale origin/gone"
	grep -q 'git fetch of the layers from origin failed' "${root}/stderr" || fct_fail "the fetch did not fail"
	[[ ! -s "${root}/checks" && -z "${run}" ]] || fct_fail "a check ran"
	fct_caller_unchanged
}

fct_case_missing_justfile() {
	test_case="layer without a justfile"
	fct_caller missing-justfile
	# A justfile above the run's folder, which just would find for a layer that has none.
	mkdir -p "${root}/cache"
	cat >"${root}/cache/justfile" <<'JUST'
check:
    @echo decoy >>"${NCLY_LAYERS_TEST}/decoy"
JUST
	fct_run 1 just check-layers nojust
	[[ ! -e "${root}/decoy" ]] || fct_fail "the justfile above the worktree ran"
	fct_expect_checks
	fct_expect_verdicts nojust "${sha_nojust}" FAIL
	fct_caller_unchanged
	test_case="layer without a justfile for signoff-layers"
	fct_run 1 just signoff-layers nojust
	rm "${root}/cache/justfile"
	[[ ! -e "${root}/decoy" ]] || fct_fail "the justfile above the worktree ran"
	fct_expect_verdicts nojust "${sha_nojust}" FAIL
	fct_expect_gh
	fct_caller_unchanged
}

fct_case_main_only_refspec() {
	local sha_moved
	test_case="main-only refspec"
	fct_caller main-only-refspec
	git -C "${caller}" config remote.origin.fetch '+refs/heads/main:refs/remotes/origin/main'
	git -C "${root}/seed" switch --quiet moving
	printf 'second\n' >"${root}/seed/moving.txt"
	git -C "${root}/seed" commit --quiet --all --message moved
	git -C "${root}/seed" push --quiet origin moving
	sha_moved="$(git -C "${root}/seed" rev-parse HEAD)"
	fct_run 0 just check-layers moving
	fct_expect_checks "${sha_moved}"
	fct_expect_verdicts moving "${sha_moved}" ok
	fct_caller_unchanged
}

fct_case_git_environment() {
	test_case="inherited GIT_DIR and GIT_WORK_TREE"
	fct_caller git-environment
	fct_run 0 env "GIT_DIR=${caller}/.git" "GIT_WORK_TREE=${caller}" just check-layers one two
	fct_expect_checks "${sha_one}" "${sha_two}"
	fct_caller_unchanged
}

fct_case_just_environment() {
	test_case="inherited JUST_ variables"
	fct_caller just-environment
	fct_run 0 env "JUST_JUSTFILE=${caller}/justfile" "JUST_WORKING_DIRECTORY=${caller}" JUST_NO_DEPS=true just check-layers one two
	fct_expect_checks "${sha_one}" "${sha_two}"
	fct_caller_unchanged
	test_case="inherited JUST_ variables for signoff-layers"
	fct_run 0 env "JUST_JUSTFILE=${caller}/justfile" "JUST_WORKING_DIRECTORY=${caller}" JUST_NO_DEPS=true just signoff-layers one two
	fct_expect_checks "${sha_one}" "${sha_two}"
	fct_expect_signed_summary
	fct_caller_unchanged
}

fct_case_discovery() {
	test_case="discovery"
	fct_caller discovery
	cat >"${root}/stack.json" <<'JSON'
{"trunk": "main", "branches": [
  {"name": "landed", "isMerged": true},
  {"name": "one", "isMerged": false},
  {"name": "two", "isMerged": false}
]}
JSON
	fct_run 0 just check-layers
	fct_expect_gh "stack view --json"
	fct_expect_checks "${sha_one}" "${sha_two}"
	fct_caller_unchanged
	test_case="discovery for signoff-layers"
	fct_run 0 just signoff-layers
	fct_expect_gh "stack view --json" "signoff --commit ${sha_one}" "signoff --commit ${sha_two}"
	fct_expect_checks "${sha_one}" "${sha_two}"
	fct_caller_unchanged
}

fct_case_signoff_stack() {
	test_case="signoff-layers on three layers"
	fct_caller signoff-stack
	fct_run 0 just signoff-layers one two three
	[[ "${run}" == "${root}/cache/nicely/signoff-layers/"* ]] || fct_fail "the logs are outside the signoff-layers folder: ${run}"
	fct_expect_checks "${sha_one}" "${sha_two}" "${sha_three}"
	fct_expect_verdicts one "${sha_one}" ok two "${sha_two}" ok three "${sha_three}" ok
	fct_expect_signed_summary
	grep -q "signoff-layers: signed off ${sha_two} after its checks passed" "${run}/02-two.log" || fct_fail "the log of two does not name its sign-off"
	fct_expect_no_worktree
	fct_caller_unchanged
	# A green status from an earlier run proves nothing about this run, so the layers are checked and signed again.
	test_case="signoff-layers over green statuses"
	(cd "${caller}" && gh signoff status --commit "${sha_two}") >/dev/null || fct_fail "the first run left no green status"
	fct_run 0 just signoff-layers one two three
	fct_expect_checks "${sha_one}" "${sha_two}" "${sha_three}"
	fct_expect_signed_summary
	fct_caller_unchanged
}

fct_case_signoff_failures() {
	test_case="signoff-layers after a failed check"
	fct_caller signoff-failures
	fct_run 1 just signoff-layers bad fixed
	fct_expect_checks "${sha_bad}"
	fct_expect_verdicts bad "${sha_bad}" FAIL
	fct_expect_gh
	fct_caller_unchanged
	test_case="signoff-layers after a failed ancestry check"
	fct_run 1 just signoff-layers one orphan above
	fct_expect_checks "${sha_one}" "${sha_orphan}"
	fct_expect_verdicts one "${sha_one}" ok orphan "${sha_orphan}" FAIL
	fct_expect_gh "signoff --commit ${sha_one}"
	fct_caller_unchanged
	test_case="signoff-layers after a moved HEAD"
	: >"${root}/detach"
	fct_run 1 just signoff-layers one two
	fct_expect_checks "${sha_one}"
	fct_expect_verdicts one "${sha_one}" FAIL
	fct_expect_gh
	fct_caller_unchanged
	test_case="signoff-layers after a failed sign-off"
	printf '%s\n' "${sha_two}" >"${root}/signoff-fails"
	fct_run 1 just signoff-layers one two three
	fct_expect_checks "${sha_one}" "${sha_two}"
	fct_expect_verdicts one "${sha_one}" ok two "${sha_two}" FAIL
	fct_expect_gh "signoff --commit ${sha_one}" "signoff --commit ${sha_two}"
	grep -q 'stub check passed' "${run}/02-two.log" || fct_fail "the log of two lacks the check that passed"
	grep -q "the checks passed, but gh signoff --commit ${sha_two} failed" "${run}/02-two.log" || fct_fail "the log of two does not name the failed sign-off"
	fct_expect_no_worktree
	fct_caller_unchanged
}

fct_case_signoff_advanced_origin() {
	test_case="signoff-layers while origin advances"
	fct_caller signoff-advanced
	printf 'two %s\n' "${sha_next}" >"${root}/advance"
	fct_run 0 just signoff-layers one two three
	[[ ! -e "${root}/advance" ]] || fct_fail "the stub check did not advance two"
	[[ "$(git --git-dir="${root}/origin.git" rev-parse refs/heads/two)" == "${sha_next}" ]] || fct_fail "two did not advance on origin"
	[[ "$(git -C "${caller}" rev-parse refs/remotes/origin/two)" == "${sha_next}" ]] || fct_fail "origin/two did not advance in the caller"
	fct_expect_checks "${sha_one}" "${sha_two}" "${sha_three}"
	fct_expect_verdicts one "${sha_one}" ok two "${sha_two}" ok three "${sha_three}" ok
	fct_expect_signed_summary
	git -C "${root}/seed" push --quiet --force origin "${sha_two}:refs/heads/two"
	fct_caller_unchanged
}

# fct_case_signal <signal> <exit> sends the signal to check-layers once the stub check runs in the worktree.
fct_case_signal() {
	local signal="${1}" expected="${2}" pid status=0 worktree
	test_case="SIG${signal}"
	fct_caller "signal-${signal}"
	: >"${root}/checks"
	: >"${root}/block"
	rm -f "${root}/started"
	# Job control gives check-layers its own process group, so the signal spares this test.
	set -m
	(cd "${caller}" && exec just check-layers one two) >"${root}/stdout" 2>"${root}/stderr" </dev/null &
	pid=$!
	set +m
	LAYERS_TEST_GROUP="${pid}"
	fct_wait_for "the stub check to start" test -e "${root}/started"
	worktree="$(cat "${root}/started")"
	[[ "${worktree}" == "${root}/cache/nicely/check-layers/"*/worktree ]] || fct_fail "the stub check ran outside the worktree: ${worktree}"
	kill "-${signal}" -- "-${pid}"
	fct_wait_for "check-layers to exit after SIG${signal}" fct_gone "${pid}"
	wait "${pid}" || status=$?
	LAYERS_TEST_GROUP=""
	rm "${root}/block"
	[[ "${status}" == "${expected}" ]] || fct_fail "expected exit ${expected}, got ${status}"
	# Without the TERM trap, bash can still run the EXIT trap, then die by the signal, so the worktree alone proves nothing.
	# just reports an exit code only when the script ended through its trap, and a signal otherwise.
	grep -q "with exit code ${expected}\$" "${root}/stderr" || fct_fail "check-layers did not end through its SIG${signal} trap"
	[[ ! -e "${worktree}" ]] || fct_fail "the worktree remains"
	[[ ! -s "${root}/stdout" && ! -s "${worktree%/worktree}/summary.txt" ]] || fct_fail "check-layers gave a verdict for the interrupted layer"
	[[ $(wc -l <"${root}/checks") -eq 1 ]] || fct_fail "a check started after the signal"
	fct_caller_unchanged
}

fct_main() {
	trap fct_cleanup EXIT
	trap 'exit 130' INT
	trap 'exit 143' TERM
	fct_setup
	fct_case_two_layers
	fct_case_failing_check
	fct_case_coherent_stack
	fct_case_missing_parent
	fct_case_moved_ref
	fct_case_moved_head
	fct_case_deleted_branch
	fct_case_missing_justfile
	fct_case_main_only_refspec
	fct_case_git_environment
	fct_case_just_environment
	fct_case_discovery
	fct_case_signoff_stack
	fct_case_signoff_failures
	fct_case_signoff_advanced_origin
	fct_case_signal INT 130
	fct_case_signal TERM 143
	printf 'check-layers and signoff-layers check each layer as origin holds it, on the layer below, sign only what passed, and leave the caller unchanged\n'
}

fct_main
exit 0
