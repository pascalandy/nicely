#!/usr/bin/env bash
# check-layers.sh runs just check on each layer of a stack, bottom to top, as
# origin holds it, for just check-layers. A layer fails when it does not
# contain the layer below. With --signoff, for just signoff-layers, it signs
# off each SHA that passed, and stops at the first failure. It runs on macOS
# and Linux.
#
#   just check-layers [branch...]
#   just signoff-layers [branch...]
#
# Without branches, it reads the layers of the current stack from gh stack.
# The run's folder keeps one log per layer, and summary.txt, which gives the
# branch, the full SHA checked, and the verdict of each finished layer.
# scripts/check-layers-test.sh proves it on a scratch stack. See #58.
set -Eeuo pipefail
# A hook or a caller can export GIT_DIR, which would point git -C at the caller's repository.
while IFS= read -r git_variable; do unset "${git_variable}"; done < <(git rev-parse --local-env-vars)
# A caller can export JUST_JUSTFILE or JUST_NO_DEPS, which would run the nested just check on the caller, or run nothing.
for just_variable in $(compgen -e); do
	if [[ "$just_variable" == JUST_* ]]; then unset "$just_variable"; fi
done

name="check-layers"
signoff=false
if [[ "${1:-}" == --signoff ]]; then
	name="signoff-layers"
	signoff=true
	shift
fi

die() {
	echo "$name: $*" >&2
	exit 1
}

branches=("$@")
if (($# == 0)); then
	stack="$(gh stack view --json)" || die "run it on a branch of the stack, or pass the branches bottom to top"
	while IFS= read -r branch; do branches+=("$branch"); done < <(jq -r '.branches[] | select(.isMerged | not) | .name' <<<"$stack")
fi
((${#branches[@]} > 0)) || die "no layer to check"
# One refspec per layer: the configured fetch may skip a layer or keep a deleted one, and a missing branch fails here.
refspecs=()
for branch in "${branches[@]}"; do refspecs+=("+refs/heads/$branch:refs/remotes/origin/$branch"); done
git fetch --quiet origin "${refspecs[@]}" || die "git fetch of the layers from origin failed"
# Freeze each layer's SHA once: a tracking ref that moves during the run must not change the commit checked.
shas=()
for branch in "${branches[@]}"; do shas+=("$(git rev-parse "refs/remotes/origin/$branch")"); done
# Not /tmp: Linux mounts it as tmpfs and macOS purges it, and the logs must outlive the run.
run="${XDG_CACHE_HOME:-$HOME/.cache}/nicely/$name/$(date +%Y%m%d-%H%M%S)-$$"
worktree="$run/worktree"
summary="$run/summary.txt"
mkdir -p "$run"
trap 'git worktree remove --force "$worktree" 2>/dev/null || true' EXIT
# Ctrl-C and SIGTERM end the run after the current command, then the EXIT trap removes the worktree.
# Without these traps, Ctrl-C moves on to the next layer, and SIGTERM can leave the worktree behind.
# Each trap ignores a repeat, such as the SIGTERM that just passes on, which would skip the EXIT trap's removal (#316).
trap 'trap "" INT TERM; exit 130' INT
trap 'trap "" INT TERM; exit 143' TERM
git worktree add --quiet --detach "$worktree" "${shas[0]}"
echo "$name: logs in $run" >&2
status=0
for i in "${!branches[@]}"; do
	branch="${branches[$i]}"
	sha="${shas[$i]}"
	log="$run/$(printf '%02d' $((i + 1)))-${branch//\//-}.log"
	: >"$log"
	verdict=ok
	# A layer restacked on an old parent can pass its own check, so the check cannot clear this failure.
	if ((i > 0)) && ! git merge-base --is-ancestor "${shas[i - 1]}" "$sha"; then
		echo "$name: origin/$branch $sha does not contain origin/${branches[i - 1]} ${shas[i - 1]}" >>"$log"
		verdict=FAIL
	fi
	git -C "$worktree" checkout --quiet --force --detach "$sha"
	git -C "$worktree" clean -ffdxq
	# The layer's own justfile, or a FAIL: without it, just would search the folders above the worktree.
	(cd "$worktree" && just --justfile "$worktree/justfile" --working-directory "$worktree" check) >>"$log" 2>&1 </dev/null || verdict=FAIL
	if [[ "$(git -C "$worktree" rev-parse HEAD)" != "$sha" ]]; then
		echo "$name: HEAD moved off origin/$branch during the check" >>"$log"
		verdict=FAIL
	fi
	# Sign the frozen SHA, never the branch: origin can move during the run, and only this SHA was checked.
	if $signoff && [[ "$verdict" == ok ]]; then
		if gh signoff --commit "$sha" >>"$log" 2>&1 </dev/null; then
			echo "$name: signed off $sha after its checks passed" >>"$log"
		else
			echo "$name: the checks passed, but gh signoff --commit $sha failed" >>"$log"
			verdict=FAIL
		fi
	fi
	[[ "$verdict" == ok ]] || status=1
	# Written once the verdict is final, so a layer that a signal interrupts gets no line.
	echo "$branch $sha $verdict" >>"$summary"
	echo "$branch $(git rev-parse --short "$sha") $verdict"
	# A layer above a failed one is never signed off.
	if $signoff && [[ "$verdict" != ok ]]; then break; fi
done
exit "$status"
