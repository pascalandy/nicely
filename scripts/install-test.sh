#!/usr/bin/env bash
# install-test.sh proves just install for just test: with names, it builds only
# ncly and the named first-party extensions, and without a name, all of them,
# into an absolute XDG_BIN_HOME and XDG_DATA_HOME. An unknown name fails with
# exit 1 before any file is written. It runs on macOS and Linux, with bash 3.2.
# See #375.
set -Eeuo pipefail
trap 'printf "%s: line %s failed: %s\n" "${BASH_SOURCE[0]##*/}" "${LINENO}" "${BASH_COMMAND}" >&2' ERR

INSTALL_TEST_ROOT=""
test_case="setup"

# shellcheck disable=SC2317,SC2329
fct_cleanup() {
	if [[ -n "${INSTALL_TEST_ROOT}" ]]; then
		rm -rf "${INSTALL_TEST_ROOT}"
	fi
}

fct_fail() {
	printf 'install-test: %s: %s\n' "${test_case}" "$*" >&2
	exit 1
}

# fct_install <expected exit> <case> [<name>...] runs just install with the names, under the
# bin and data folders of a new folder for the case.
fct_install() {
	local expected="${1}" status=0
	test_case="${2}"
	shift 2
	case_dir="${root}/${test_case// /-}"
	mkdir "${case_dir}"
	XDG_BIN_HOME="${case_dir}/bin" XDG_DATA_HOME="${case_dir}/data" \
		just --justfile "${repo_root}/justfile" --working-directory "${repo_root}" install "$@" >"${root}/output" 2>&1 </dev/null || status=$?
	if [[ "${status}" != "${expected}" ]]; then
		cat "${root}/output" >&2
		fct_fail "just install $* exited ${status}, not ${expected}"
	fi
}

# fct_files [<path>...] requires the folder of the case to hold these paths, and no other,
# down to the entries of the extensions folder, so a staging folder that remains fails too.
fct_files() {
	(cd "${case_dir}" && find . -mindepth 1 -maxdepth 4 | sed 's|^\./||' | LC_ALL=C sort) >"${root}/files"
	if (($# > 0)); then printf '%s\n' "$@"; fi | LC_ALL=C sort >"${root}/files.expected"
	if ! cmp -s "${root}/files.expected" "${root}/files"; then
		diff "${root}/files.expected" "${root}/files" >&2 || true
		fct_fail "the install wrote other files"
	fi
}

# fct_agent requires the install of the agent extension in the extensions folder of the case.
fct_agent() {
	local extensions="${case_dir}/data/nicely/extensions"
	[[ -x "${extensions}/ncly-agent" ]] || fct_fail "ncly-agent is no executable"
	[[ -f "${extensions}/ncly-agent.locales/en.toml" ]] || fct_fail "ncly-agent.locales holds no en.toml"
	[[ -f "${extensions}/ncly-agent.skill/SKILL.md" ]] || fct_fail "ncly-agent.skill holds no SKILL.md"
}

fct_main() {
	local extensions name main leftover
	local -a agent=(data data/nicely data/nicely/extensions data/nicely/extensions/ncly-agent
		data/nicely/extensions/ncly-agent.locales data/nicely/extensions/ncly-agent.skill data/nicely/extensions/ncly-agent.toml)
	repo_root="$(cd "${BASH_SOURCE[0]%/*}/.." && pwd -P)"
	root="$(mktemp -d "${TMPDIR:-/tmp}/ncly-install-test.XXXXXXXX")"
	INSTALL_TEST_ROOT="${root}"
	trap fct_cleanup EXIT

	# Two names install those two, and nothing else.
	fct_install 0 "ncly and agent" ncly agent
	fct_files bin bin/ncly "${agent[@]}"
	fct_agent
	"${case_dir}/bin/ncly" --version >/dev/null || fct_fail "the installed ncly does not run"

	# An extension alone creates no bin folder.
	fct_install 0 "agent alone" agent
	fct_files "${agent[@]}"
	fct_agent

	# An unknown name names the valid ones and writes nothing, not even a folder.
	fct_install 1 "unknown name" ncly agent unknown
	grep -qF 'unknown name "unknown"' "${root}/output" || fct_fail "the error names no unknown name: $(cat "${root}/output")"
	grep -qE 'ncly .*agent' "${root}/output" || fct_fail "the error lists no valid name: $(cat "${root}/output")"
	fct_files

	# Without a name, ncly and every first-party extension.
	fct_install 0 "every name"
	[[ -x "${case_dir}/bin/ncly" ]] || fct_fail "the install wrote no bin/ncly"
	extensions="${case_dir}/data/nicely/extensions"
	for main in "${repo_root}"/extensions/*/cmd/ncly-*/; do
		main="${main%/}"
		name="${main##*/}"
		[[ -x "${extensions}/${name}" && -f "${extensions}/${name}.toml" ]] || fct_fail "the install wrote no ${name}"
	done
	fct_agent
	leftover="$(find "${case_dir}" -name '.*install.*')"
	[[ -z "${leftover}" ]] || fct_fail "a staging folder remains: ${leftover}"
	printf 'install-test: just install builds only the names that it gets, or all of them\n'
}

fct_main
