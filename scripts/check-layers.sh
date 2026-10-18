#!/usr/bin/env bash
# check-layers.sh runs just check on each layer of a stack, bottom to top, as
# origin holds it, for just check-layers. It runs on macOS and Linux.
#
#   just check-layers [branch...]
#
# Without branches, it reads the layers of the current stack from gh stack.
# scripts/check-layers-test.sh proves it on a scratch stack. See #58.
set -Eeuo pipefail
# A hook or a caller can export GIT_DIR, which would point git -C at the caller's repository.
while IFS= read -r git_variable; do unset "${git_variable}"; done < <(git rev-parse --local-env-vars)
# A caller can export JUST_JUSTFILE or JUST_NO_DEPS, which would run the nested just check on the caller, or run nothing.
for just_variable in $(compgen -e); do
	if [[ "$just_variable" == JUST_* ]]; then unset "$just_variable"; fi
done

die() {
	echo "check-layers: $*" >&2
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
# Not /tmp: Linux mounts it as tmpfs and macOS purges it, and the logs must outlive the run.
run="${XDG_CACHE_HOME:-$HOME/.cache}/nicely/check-layers/$(date +%Y%m%d-%H%M%S)-$$"
worktree="$run/worktree"
mkdir -p "$run"
trap 'git worktree remove --force "$worktree" 2>/dev/null || true' EXIT
# Ctrl-C and SIGTERM end the run after the current command, then the EXIT trap removes the worktree.
# Without these traps, Ctrl-C moves on to the next layer, and SIGTERM can leave the worktree behind.
trap 'exit 130' INT
trap 'exit 143' TERM
git worktree add --quiet --detach "$worktree" "refs/remotes/origin/${branches[0]}"
echo "check-layers: logs in $run" >&2
status=0
for i in "${!branches[@]}"; do
	branch="${branches[$i]}"
	sha="$(git rev-parse "refs/remotes/origin/$branch")"
	log="$run/$(printf '%02d' $((i + 1)))-${branch//\//-}.log"
	git -C "$worktree" checkout --quiet --force --detach "$sha"
	git -C "$worktree" clean -ffdxq
	verdict=ok
	# The layer's own justfile, or a FAIL: without it, just would search the folders above the worktree.
	(cd "$worktree" && just --justfile "$worktree/justfile" --working-directory "$worktree" check) >"$log" 2>&1 </dev/null || verdict=FAIL
	if [[ "$(git -C "$worktree" rev-parse HEAD)" != "$sha" ]]; then
		echo "check-layers: HEAD moved off origin/$branch during the check" >>"$log"
		verdict=FAIL
	fi
	[[ "$verdict" == ok ]] || status=1
	echo "$branch $(git rev-parse --short "$sha") $verdict"
done
exit "$status"
