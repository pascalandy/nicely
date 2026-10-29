#!/usr/bin/env bash
# gitleaks-rules.sh proves the leak rules for just gitleaks-rules: every line
# of testdata/gitleaks/leaks.txt is blocked, and every line of clean.txt
# passes.
set -euo pipefail
cd "$(cd "${BASH_SOURCE[0]%/*}/.." && pwd -P)"
scan() { printf '%s\n' "$1" | gitleaks stdin --config .gitleaks.toml --no-banner --log-level error >/dev/null; }
status=0
while IFS= read -r line; do
	if scan "$line"; then
		echo "not blocked: $line" >&2
		status=1
	fi
done <testdata/gitleaks/leaks.txt
while IFS= read -r line; do
	if ! scan "$line"; then
		echo "blocked: $line" >&2
		status=1
	fi
done <testdata/gitleaks/clean.txt
exit "$status"
