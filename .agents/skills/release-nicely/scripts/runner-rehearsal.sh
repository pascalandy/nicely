#!/usr/bin/env bash
# runner-rehearsal.sh rehearses the run of release.yml before the tag, in ubuntu:24.04 with the
# runner's tools. It runs the workflow's own steps through "Rehearse without publishing" (#94).
# Usage: bash runner-rehearsal.sh vX.Y.Z [commit]. The commit defaults to HEAD and must be on
# GitHub. It needs Linux on x86_64 and podman, publishes nothing, and keeps its log.
set -Eeuo pipefail
tag="${1:?usage: runner-rehearsal.sh vX.Y.Z [commit]}"
fail() {
	printf 'runner-rehearsal: %s\n' "$*" >&2
	exit 1
}
[[ "$(uname -sm)" == "Linux x86_64" ]] || fail "it needs Linux on x86_64, as the runner has"
command -v podman >/dev/null || fail "it needs podman"
here="$(cd "${BASH_SOURCE[0]%/*}" && pwd -P)"
sha="$(git -C "${here}" rev-parse --verify "${2:-HEAD}^{commit}")"
log="${XDG_CACHE_HOME:-${HOME}/.cache}/nicely/runner-rehearsal/${tag}-${sha:0:12}-$(date +%Y%m%d-%H%M%S).log"
mkdir -p "${log%/*}"
printf 'runner-rehearsal: %s at %s, log in %s\n' "${tag}" "${sha}" "${log}" >&2
podman run --rm -v "${here}/runner-rehearsal-job.sh:/job.sh:ro" docker.io/library/ubuntu:24.04 \
	bash /job.sh "${sha}" "${tag}" </dev/null 2>&1 | tee "${log}"
