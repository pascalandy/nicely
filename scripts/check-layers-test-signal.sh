# shellcheck shell=bash disable=SC2034,SC2154
# The Ctrl-C and SIGTERM cases of check-layers-test.sh, which sources this file and sets the
# variables that they read, such as root and caller.

# fct_case_signal <signal> <exit> sends the signal to check-layers once the stub check runs in the worktree.
fct_case_signal() {
	local signal="${1}" expected="${2}" pid status=0 worktree layers
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
	# Repeat the signal until check-layers exits, as just repeats a SIGTERM to its child (#316).
	layers="${worktree%/worktree}"
	while kill "-${signal}" "${layers##*-}" 2>/dev/null; do :; done
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
