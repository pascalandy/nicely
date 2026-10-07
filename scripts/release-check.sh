#!/usr/bin/env bash
set -Eeuo pipefail

RELEASE_SCRATCH=""

# shellcheck disable=SC2329
fct_finish() {
	local status=$?
	if [[ -n "${RELEASE_SCRATCH}" ]]; then
		printf 'Release rehearsal files retained at %s\n' "${RELEASE_SCRATCH}" >&2
	fi
	return "${status}"
}

fct_clean_environment() {
	env -i "${RELEASE_ENV[@]}" "$@"
}

fct_same_sources() {
	local root="${1}" tree="${2}" file
	git -C "${root}" ls-files -z >"${RELEASE_SCRATCH}/tracked.after"
	cmp "${RELEASE_SCRATCH}/tracked.before" "${RELEASE_SCRATCH}/tracked.after"
	git -C "${root}" ls-files --others --exclude-standard -z >"${RELEASE_SCRATCH}/untracked"
	if [[ -s "${RELEASE_SCRATCH}/untracked" ]]; then
		printf 'Stage or remove untracked source files before release-check\n' >&2
		return 1
	fi
	while IFS= read -r -d '' file; do
		if [[ -L "${root}/${file}" || ! -f "${root}/${file}" ]]; then
			printf 'Tracked source is missing or a symbolic link: %s\n' "${file}" >&2
			return 1
		fi
		cmp "${root}/${file}" "${tree}/${file}"
		if [[ -x "${root}/${file}" && ! -x "${tree}/${file}" ]] || [[ ! -x "${root}/${file}" && -x "${tree}/${file}" ]]; then
			printf 'Source mode changed during capture: %s\n' "${file}" >&2
			return 1
		fi
	done <"${RELEASE_SCRATCH}/tracked.before"
}

fct_guards() {
	local guards="${1}" command
	mkdir -p "${guards}"
	cat >"${guards}/git" <<'SCRIPT'
#!/usr/bin/env bash
set -Eeuo pipefail
original_args=("$@")
if [[ $# -ge 4 && "${1}" == -c && "${2}" == log.showSignature=false && "${3}" == -c && "${4}" == column.ui=never ]]; then
	shift 4
fi
case "${1:-}" in
rev-parse|show|log|describe|rev-list|status|archive|version|--version|ls-files) exec "${NCLY_REAL_GIT}" "${original_args[@]}" ;;
tag) case "${2:-}" in -l|--list|--points-at) exec "${NCLY_REAL_GIT}" "${original_args[@]}" ;; esac ;;
ls-remote) if [[ $# -eq 2 && "${2}" == --get-url ]]; then exec "${NCLY_REAL_GIT}" "${original_args[@]}"; fi ;;
config) if [[ $# -eq 2 && "${2}" == gpg.program ]]; then exec "${NCLY_REAL_GIT}" "${original_args[@]}"; fi ;;
esac
printf 'git command blocked: %s\n' "${1:-<missing>}" >>"${NCLY_BLOCKED_LOG}"
printf 'release-check blocked a guarded git command\n' >&2
exit 1
SCRIPT
	cat >"${guards}/deny" <<'SCRIPT'
#!/usr/bin/env bash
set -Eeuo pipefail
printf '%s blocked\n' "${0##*/}" >>"${NCLY_BLOCKED_LOG}"
printf 'release-check blocked %s\n' "${0##*/}" >&2
exit 1
SCRIPT
	chmod 755 "${guards}/git" "${guards}/deny"
	for command in ssh gh curl wget scp sftp rsync; do
		ln -s deny "${guards}/${command}"
	done
}

fct_execute_this() {
	local tag="${1}" root tree file parent go_path go_cache go_modcache go_toolchain git_path variable
	root="$(git rev-parse --show-toplevel)"
	RELEASE_SCRATCH="$(mktemp -d "${TMPDIR:-/tmp}/ncly-release-check.XXXXXXXX")"
	tree="${RELEASE_SCRATCH}/tree"
	mkdir -p "${tree}" "${RELEASE_SCRATCH}/home" "${RELEASE_SCRATCH}/bin"
	git -C "${root}" ls-files -z >"${RELEASE_SCRATCH}/tracked.before"
	git -C "${root}" ls-files --others --exclude-standard -z >"${RELEASE_SCRATCH}/untracked"
	if [[ -s "${RELEASE_SCRATCH}/untracked" ]]; then
		printf 'Stage or remove untracked source files before release-check\n' >&2
		return 1
	fi
	while IFS= read -r -d '' file; do
		if [[ -L "${root}/${file}" || ! -f "${root}/${file}" ]]; then
			printf 'Tracked source is missing or a symbolic link: %s\n' "${file}" >&2
			return 1
		fi
		parent="${file%/*}"
		if [[ "${parent}" != "${file}" ]]; then
			mkdir -p "${tree}/${parent}"
		fi
		cp -p "${root}/${file}" "${tree}/${file}"
	done <"${RELEASE_SCRATCH}/tracked.before"
	fct_same_sources "${root}" "${tree}"
	(cd "${root}" && just check)
	fct_same_sources "${root}" "${tree}"
	go_path="$(go env GOPATH)"
	go_cache="$(go env GOCACHE)"
	go_modcache="$(go env GOMODCACHE)"
	go_toolchain="$(go -C "${root}/tools" env GOVERSION)"
	git_path="$(command -v git)"
	RELEASE_ENV=("PATH=${PATH}" "HOME=${RELEASE_SCRATCH}/home" "XDG_CONFIG_HOME=${RELEASE_SCRATCH}/home/config" "XDG_CACHE_HOME=${RELEASE_SCRATCH}/home/cache" "XDG_DATA_HOME=${RELEASE_SCRATCH}/home/data" "XDG_STATE_HOME=${RELEASE_SCRATCH}/home/state" "GOPATH=${go_path}" "GOCACHE=${go_cache}" "GOMODCACHE=${go_modcache}" "GOENV=off" "GOFLAGS=-mod=readonly" "GOTOOLCHAIN=${go_toolchain}" "CGO_ENABLED=0" "GIT_CONFIG_GLOBAL=/dev/null" "GIT_CONFIG_NOSYSTEM=1" "GORELEASER_CURRENT_TAG=${tag}" "NCLY_RELEASE_VERSION=${tag#v}" "AUR_SSH_KEY=" "NCLY_REAL_GIT=${git_path}" "NCLY_BLOCKED_LOG=${RELEASE_SCRATCH}/blocked-publishers.log")
	for variable in HTTP_PROXY HTTPS_PROXY ALL_PROXY NO_PROXY SSL_CERT_FILE SSL_CERT_DIR; do
		if [[ -n "${!variable:-}" ]]; then
			RELEASE_ENV+=("${variable}=${!variable}")
		fi
	done
	fct_clean_environment "${git_path}" -C "${tree}" -c init.templateDir= init --initial-branch=main --quiet
	fct_clean_environment "${git_path}" -C "${tree}" -c core.hooksPath=/dev/null add --all
	fct_clean_environment "${git_path}" -C "${tree}" -c core.hooksPath=/dev/null -c user.name='Release rehearsal' -c user.email=release@example.com commit --quiet -m 'Capture release-check inputs'
	fct_clean_environment "${git_path}" -C "${tree}" remote add origin https://github.com/pascalandy/nicely.git
	fct_guards "${RELEASE_SCRATCH}/guards"
	RELEASE_ENV[0]="PATH=${RELEASE_SCRATCH}/guards:${PATH}"
	cd "${tree}"
	fct_clean_environment go tool -modfile=tools/go.mod govulncheck ./...
	fct_clean_environment go tool -modfile=tools/go.mod goreleaser check
	fct_clean_environment go run -C tools ./cmd/release-artifacts assets "${tag}" "${tree}"
	fct_clean_environment go tool -modfile=tools/go.mod goreleaser release --snapshot --clean
	fct_clean_environment go run -C tools ./cmd/release-artifacts verify "${tag}" "${tree}" "${tree}/dist"
	fct_clean_environment go run -C tools ./cmd/release-artifacts formula "${tag}" "${tree}/dist"
	fct_clean_environment bash scripts/verify-aur.sh "${tag}" "${tree}/dist"
	fct_clean_environment ruby -c "${tree}/dist/Formula/ncly.rb"
	fct_same_sources "${root}" "${tree}"
	if [[ -s "${RELEASE_SCRATCH}/blocked-publishers.log" ]]; then
		printf 'A guarded command was attempted during release-check\n' >&2
		return 1
	fi
	printf 'Release rehearsal artifacts: %s\n' "${tree}/dist"
}

fct_main() {
	if [[ $# -ne 1 || ! "${1}" =~ ^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$ ]]; then
		printf 'Usage: just release-check vX.Y.Z\n' >&2
		exit 2
	fi
	trap 'fct_finish' EXIT
	trap 'exit 130' INT
	trap 'exit 143' TERM
	fct_execute_this "${1}"
}

fct_main "$@"
exit 0
