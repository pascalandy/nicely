#!/usr/bin/env bash
# Run the shared isolated-checkout signoff program. scripts/signoff-test.sh proves it.
set -Eeuo pipefail
root="$(cd "${BASH_SOURCE[0]%/*}/.." && pwd -P)"
cd "$root"
exec go run ./sdk/extensionrepo/cmd/extensionrepo signoff
