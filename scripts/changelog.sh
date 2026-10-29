#!/usr/bin/env bash
# changelog.sh writes the GitHub release notes of one version, from the
# previous tag to HEAD, for just changelog vX.Y.Z [previous-tag].
set -Eeuo pipefail
cd "$(cd "${BASH_SOURCE[0]%/*}/.." && pwd -P)"
version="${1:-}"
[[ "${version}" =~ ^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$ ]] || {
	printf 'Usage: just changelog vX.Y.Z [previous-tag]\n' >&2
	exit 2
}
previous="${2:-}"
if [[ -z "${previous}" ]]; then
	previous="$(git describe --tags --abbrev=0 --match 'v[0-9]*' --exclude "${version}" HEAD)"
fi
git rev-parse --verify "refs/tags/${previous}^{commit}" >/dev/null
git merge-base --is-ancestor "refs/tags/${previous}" HEAD
mkdir -p .release
uv tool run git-cliff@2.14.2 --config cliff.toml --tag "${version}" --output .release/changelog.md "${previous}..HEAD"
printf 'Changelog: .release/changelog.md (%s..HEAD)\n' "${previous}"
