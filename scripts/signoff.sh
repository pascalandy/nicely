#!/usr/bin/env bash
# signoff.sh runs just check, then marks the pushed HEAD green with
# gh signoff, for just signoff. The check runs here, not as a dependency of
# the recipe, because JUST_NO_DEPS and --no-deps skip dependencies. It runs on
# macOS and Linux. scripts/signoff-test.sh proves it. See #213.
set -Eeuo pipefail
# A caller can export JUST_NO_DEPS or JUST_JUSTFILE, which would run nothing, or the check of another justfile.
for just_variable in $(compgen -e); do
	if [[ "$just_variable" == JUST_* ]]; then unset "$just_variable"; fi
done

root="$(cd "${BASH_SOURCE[0]%/*}/.." && pwd -P)"
cd "$root"
just --justfile "$root/justfile" --working-directory "$root" check
gh signoff
