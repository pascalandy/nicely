#!/usr/bin/env bash
# Sourced by template-test.sh for file edits and updater assertions.
# shellcheck disable=SC2154 # fct_repository sets repo and scratch in template-test.sh.

# fct_edit replaces the line $2 of the file $1 with $3, or appends $3 when $2
# is empty. sed -i differs between macOS and Linux, so the test avoids it.
fct_edit() {
	local file="${1}" old="${2}" new="${3}" line found=""
	while IFS= read -r line || [[ -n "${line}" ]]; do
		if [[ -n "${old}" && "${line}" == "${old}" ]]; then
			line="${new}" found=1
		fi
		printf '%s\n' "${line}"
	done <"${file}" >"${file}.new"
	if [[ -z "${old}" ]]; then
		printf '%s\n' "${new}" >>"${file}.new"
	elif [[ -z "${found}" ]]; then
		fct_fail "${file} holds no line \"${old}\""
	fi
	mv "${file}.new" "${file}"
}

# fct_has fails unless the file $1 holds the line $2.
fct_has() {
	grep -qxF -- "${2}" "${1}" || fct_fail "${1} lacks the line \"${2}\""
}

# fct_update runs just update $1 in the rendered repository. It keeps the
# output in update.out and returns the exit code of just.
fct_update() {
	local status=0
	(cd "${repo}" && just update "${1}") >"${scratch}/update.out" 2>&1 || status=$?
	return "${status}"
}

# fct_said fails unless the last just update printed the line $1.
fct_said() {
	if ! grep -qxF -- "${1}" "${scratch}/update.out"; then
		cat "${scratch}/update.out" >&2
		fct_fail "just update did not print \"${1}\""
	fi
}
