#!/usr/bin/env bash
set -Eeuo pipefail

# Verifies a published release in its own temporary folder, and never changes the checkout or
# GitHub. The release-artifacts verifier of the tag checks the assets. The formula of the
# Homebrew tap gets its own result, since the Homebrew tap can move on to a later release.
# Exits 0 only when the assets pass and the Homebrew tap matches them, 1 otherwise.

readonly NCLY_REPOSITORY="pascalandy/nicely"
readonly NCLY_SOURCE_URL="https://github.com/pascalandy/nicely.git"
readonly NCLY_TAP_URL="https://github.com/pascalandy/homebrew-tap.git"
VERIFY_SCRATCH=""

# shellcheck disable=SC2317,SC2329
fct_finish() {
	local status=$?
	if [[ -n "${VERIFY_SCRATCH}" ]]; then
		printf 'Release verification files retained at %s\n' "${VERIFY_SCRATCH}" >&2
	fi
	return "${status}"
}

fct_git() {
	GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 git -c init.templateDir= -c advice.detachedHead=false "$@"
}

fct_clean_environment() {
	env -i "${VERIFY_ENV[@]}" "$@"
}

# Prints the Go toolchain that setup-go installs from go-version-file: the toolchain line
# first, then the go line.
fct_toolchain() {
	local mod="${1}" version
	version="$(sed -nE 's/^toolchain go(1\.[0-9]+(\.[0-9]+|rc[0-9]+)?).*/\1/p' "${mod}")"
	if [[ -z "${version}" ]]; then
		version="$(sed -nE 's/^go ([0-9]+(\.[0-9]+)*).*/\1/p' "${mod}")"
	fi
	if [[ -z "${version}" ]]; then
		printf 'tools/go.mod names no Go version\n' >&2
		return 1
	fi
	printf 'go%s\n' "${version}"
}

# Prints a line for each binary archive that another Go release built. Such a binary differs
# from its rebuild even when nobody changed it. A toolchain names a non-default experiment only
# for the targets where it differs, such as go1.27.2-X:nodwarf5 on Linux, so only the release
# part of each version counts.
fct_compiler_gap() {
	local dist="${1}" used="${2}" archive name folder built
	for archive in "${dist}"/ncly_*.tar.gz; do
		name="${archive##*/}"
		if [[ "${name}" == *_source.tar.gz ]]; then
			continue
		fi
		folder="${VERIFY_SCRATCH}/compiler/${name%.tar.gz}"
		mkdir -p "${folder}"
		if tar -xzf "${archive}" -C "${folder}" ncly 2>/dev/null && built="$(fct_clean_environment go version "${folder}/ncly" 2>/dev/null)"; then
			built="${built##*: }"
			if [[ "${built%%[ -]*}" != "${used%%[ -]*}" ]]; then
				printf 'Compiler: %s was built with %s, and this check rebuilt it with %s. Its rebuild can differ without any change to the asset.\n' "${name}" "${built}" "${used}"
			fi
		fi
	done
}

fct_execute_this() {
	local tag="${1}" tree dist tap commit toolchain used verifier failure gap tap_commit formula announced variable
	VERIFY_SCRATCH="$(mktemp -d "${TMPDIR:-/tmp}/ncly-release-verify.XXXXXXXX")"
	tree="${VERIFY_SCRATCH}/tree"
	dist="${VERIFY_SCRATCH}/dist"
	tap="${VERIFY_SCRATCH}/tap"
	mkdir -p "${dist}" "${VERIFY_SCRATCH}/bin" "${VERIFY_SCRATCH}/home" "${VERIFY_SCRATCH}/tmp"
	# The full ref reads the tag, even when a branch bears its name.
	fct_git init --quiet "${tree}"
	fct_git -C "${tree}" fetch --quiet --depth 1 "${NCLY_SOURCE_URL}" "refs/tags/${tag}:refs/tags/${tag}"
	fct_git -C "${tree}" checkout --quiet --detach "refs/tags/${tag}"
	commit="$(fct_git -C "${tree}" rev-parse HEAD)"
	gh release download "${tag}" --repo "${NCLY_REPOSITORY}" --dir "${dist}"
	# The workflow builds with the Go version of tools/go.mod at the tag, so the rebuild uses it too.
	toolchain="$(fct_toolchain "${tree}/tools/go.mod")"
	VERIFY_ENV=("PATH=${PATH}" "HOME=${VERIFY_SCRATCH}/home" "XDG_CONFIG_HOME=${VERIFY_SCRATCH}/home/config" "XDG_CACHE_HOME=${VERIFY_SCRATCH}/home/cache" "XDG_DATA_HOME=${VERIFY_SCRATCH}/home/data" "XDG_STATE_HOME=${VERIFY_SCRATCH}/home/state" "TMPDIR=${VERIFY_SCRATCH}/tmp" "GOPATH=$(go env GOPATH)" "GOCACHE=$(go env GOCACHE)" "GOMODCACHE=$(go env GOMODCACHE)" "GOENV=off" "GOFLAGS=-mod=readonly" "GOTOOLCHAIN=${toolchain}" "CGO_ENABLED=0" "GIT_CONFIG_GLOBAL=/dev/null" "GIT_CONFIG_NOSYSTEM=1")
	for variable in HTTP_PROXY HTTPS_PROXY ALL_PROXY NO_PROXY SSL_CERT_FILE SSL_CERT_DIR; do
		if [[ -n "${!variable:-}" ]]; then
			VERIFY_ENV+=("${variable}=${!variable}")
		fi
	done
	cd "${tree}"
	used="$(fct_clean_environment go env GOVERSION)"
	# The verifier of the tag also wrote its assets and its formula, so it reads them back.
	verifier="${VERIFY_SCRATCH}/bin/release-artifacts"
	fct_clean_environment go build -C tools -o "${verifier}" ./cmd/release-artifacts
	fct_clean_environment "${verifier}" assets "${tag}" "${tree}"
	printf 'Tag: %s\nCommit: %s\nGo: %s, from tools/go.mod at %s\n' "${tag}" "${commit}" "${toolchain}" "${tag}" >"${VERIFY_SCRATCH}/report"
	if ! fct_clean_environment "${verifier}" verify "${tag}" "${tree}" "${dist}" 2>"${VERIFY_SCRATCH}/verify.err"; then
		cat "${VERIFY_SCRATCH}/verify.err" >&2
		failure="$(head -n 1 "${VERIFY_SCRATCH}/verify.err")"
		gap="$(fct_compiler_gap "${dist}" "${used}")"
		printf 'Assets: failed: %s\n' "${failure}" >>"${VERIFY_SCRATCH}/report"
		if [[ -n "${gap}" ]]; then
			printf '%s\nHomebrew tap: not compared, because the assets failed\nResult: assets not verified, compiler gap\n' "${gap}" >>"${VERIFY_SCRATCH}/report"
		else
			printf 'Homebrew tap: not compared, because the assets failed\nResult: assets failed\n' >>"${VERIFY_SCRATCH}/report"
		fi
		cat "${VERIFY_SCRATCH}/report"
		return 1
	fi
	printf 'Assets: verified\n' >>"${VERIFY_SCRATCH}/report"
	fct_clean_environment "${verifier}" formula "${tag}" "${dist}"
	# One clone reads the Homebrew tap, so the formula and the commit that the report names agree.
	fct_git clone --quiet --depth 1 --branch main --single-branch "${NCLY_TAP_URL}" "${tap}"
	tap_commit="$(fct_git -C "${tap}" rev-parse HEAD)"
	formula="${tap}/Formula/ncly.rb"
	printf 'Homebrew tap commit: %s\n' "${tap_commit}" >>"${VERIFY_SCRATCH}/report"
	if [[ ! -f "${formula}" ]]; then
		printf 'Homebrew tap: Formula/ncly.rb is missing\nResult: assets verified, Homebrew tap formula wrong\n' >>"${VERIFY_SCRATCH}/report"
	elif cmp -s "${dist}/Formula/ncly.rb" "${formula}"; then
		printf 'Homebrew tap: Formula/ncly.rb matches %s\nResult: assets verified, Homebrew tap matches\n' "${tag}" >>"${VERIFY_SCRATCH}/report"
		cat "${VERIFY_SCRATCH}/report"
		return 0
	else
		announced="$(sed -nE 's|^  url "https://github\.com/pascalandy/nicely/releases/download/([^/"]+)/.*|\1|p' "${formula}")"
		if [[ -n "${announced}" && "${announced}" != "${tag}" ]]; then
			printf 'Homebrew tap: Formula/ncly.rb announces %s\nResult: assets verified, Homebrew tap on another version\n' "${announced}" >>"${VERIFY_SCRATCH}/report"
		else
			diff -u "${dist}/Formula/ncly.rb" "${formula}" >&2 || true
			printf 'Homebrew tap: Formula/ncly.rb differs from the formula rendered from the %s assets\nResult: assets verified, Homebrew tap formula wrong\n' "${tag}" >>"${VERIFY_SCRATCH}/report"
		fi
	fi
	cat "${VERIFY_SCRATCH}/report"
	return 1
}

fct_main() {
	local variable
	if [[ $# -ne 1 || ! "${1}" =~ ^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$ ]]; then
		printf 'Usage: just release-verify vX.Y.Z\n' >&2
		exit 2
	fi
	# A git hook can export GIT_DIR, which would aim the clones at the caller's repository.
	while IFS= read -r variable; do unset "${variable}"; done < <(git rev-parse --local-env-vars)
	trap 'fct_finish' EXIT
	trap 'exit 130' INT
	trap 'exit 143' TERM
	fct_execute_this "${1}"
}

fct_main "$@"
exit 0
