#!/usr/bin/env bash
# check.sh runs every check that signoff requires, in the order of checks, for
# just check. It stops at the first check that fails, with its exit code. With
# --keep-going, it runs every check, then prints each one with its exit code,
# and exits 1 when one failed. It runs on macOS and Linux, with bash 3.2.
# scripts/check-test.sh proves it. See #357.
set -Eeuo pipefail

checks=(fmt-check lint tidy-check test race template python release-lint gitleaks-rules gitleaks "limits --base origin/main")

keep_going=""
case "$*" in
"") ;;
--keep-going) keep_going=1 ;;
*)
	echo "usage: just check [--keep-going]" >&2
	exit 2
	;;
esac

root="$(cd "${BASH_SOURCE[0]%/*}/.." && pwd -P)"
# A Ctrl-C stops every check, not only the one that runs.
trap 'exit 130' INT
trap 'exit 143' TERM

failed=0
summary=""
for check in "${checks[@]}"; do
	read -r -a words <<<"$check"
	code=0
	just --justfile "$root/justfile" --working-directory "$root" "${words[@]}" || code=$?
	if [[ -z "$keep_going" && "$code" != 0 ]]; then
		exit "$code"
	fi
	[[ "$code" == 0 ]] || failed=$((failed + 1))
	summary+="$(printf '%-16s exit %s' "${words[0]}" "$code")"$'\n'
done
if [[ -n "$keep_going" ]]; then
	printf '\njust check --keep-going:\n%s' "$summary"
	if ((failed > 0)); then
		printf '%s of %s checks failed\n' "$failed" "${#checks[@]}"
		exit 1
	fi
	printf 'every check passed\n'
fi
