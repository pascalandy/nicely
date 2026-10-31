#!/usr/bin/env bash
# Sourced by template-test.sh to prove the files and the recipes that a
# rendered repository adds to the extension.

# fct_test_files checks the files of the rendered repository $1 that no
# recipe runs, and that its forms keep no Jinja. The slug is $2.
fct_test_files() {
	local repo="${1}" slug="${2}" file recipe
	fct_has "${repo}/.github/workflows/check.yml" "        run: just check"
	fct_has "${repo}/.github/workflows/check.yml" "        os: \${{ github.event.repository.private && fromJSON('[\"ubuntu-24.04\"]') || fromJSON('[\"ubuntu-24.04\",\"macos-15\"]') }}"
	fct_has "${repo}/.github/workflows/check.yml" "    runs-on: \${{ matrix.os }}"
	fct_has "${repo}/qa.md" "# ${slug} QA"
	for recipe in install uninstall signoff; do
		[[ "$(just --justfile "${repo}/justfile" --show "${recipe}" | sed '/^#/d')" == "${recipe}:"$'\n'"    go run ./internal/export ${recipe}" ]] ||
			fct_fail "just ${recipe} must call the SDK program in one line"
	done
	for file in qa.md acceptance.sh; do
		if grep -nE '\{\{|\}\}|\{%|%\}|\{#|#\}' "${repo}/${file}"; then
			fct_fail "${file} keeps Jinja after the render"
		fi
	done
}

# fct_test_example_notices requires notices from plain just check in the rendered repository $1, whose slug is $3.
fct_test_example_notices() {
	local repo="${1}" scratch="${2}" status=0 line examples
	local skipped='skipped: command is outside this manifest, so its paid effects are unknown'
	local code reason='Requires a controlled runtime; https://example.com/issues/1'
	code="$(printf '%s_WARNING' "${3}" | tr '[:lower:]' '[:upper:]' | tr '-' '_')"
	cp "${repo}/manifest.toml" "${scratch}/examples-manifest.toml"
	{
		printf 'warnings = [{ code = "%s", untested = "%s" }]\n' "${code}" "${reason}"
		cat "${scratch}/examples-manifest.toml"
	} >"${repo}/manifest.toml"
	# The template's command offers a dry run, so an example of a core command gives the skip.
	examples="$(grep -m1 '^examples = \[' "${repo}/manifest.toml")"
	fct_edit "${repo}/manifest.toml" "${examples}" "${examples%]}, \"ncly describe\"]"
	if ! (cd "${repo}" && just check) >"${scratch}/examples-check.out" 2>"${scratch}/examples-check.err"; then
		cat "${scratch}/examples-check.out" "${scratch}/examples-check.err" >&2
		fct_fail 'just check fails when an example is skipped'
	fi
	cat "${scratch}/examples-check.out"
	cat "${scratch}/examples-check.err" >&2
	if ! grep -qF "warning ${code} untested" "${scratch}/examples-check.err" ||
		! grep -qF "${reason}" "${scratch}/examples-check.err"; then
		fct_fail 'plain just check hides the declared-code reservation on stderr'
	fi
	if ! (cd "${repo}" && go run ./internal/export test -json) >"${scratch}/examples-test.jsonl" 2>"${scratch}/examples-test.err"; then
		cat "${scratch}/examples-test.err" >&2
		fct_fail 'the JSON test wrapper fails'
	fi
	[[ -s "${scratch}/examples-test.jsonl" ]] || fct_fail 'the JSON test wrapper prints no test events'
	while IFS= read -r line || [[ -n "${line}" ]]; do
		PATH="/usr/bin:${PATH}" jq -e . <<<"${line}" >/dev/null || fct_fail 'the test wrapper prints non-JSON output on stdout'
	done <"${scratch}/examples-test.jsonl"
	grep -qF "${skipped}" "${scratch}/examples-test.err" ||
		fct_fail 'the JSON test wrapper hides the skip reason on stderr'
	grep -qF "${skipped}" "${scratch}/examples-check.err" ||
		fct_fail 'plain just check hides the skip reason on stderr'
	cp "${repo}/testdata/script/refusal.txtar" "${scratch}/examples-refusal.txtar"
	{
		printf 'notice-regression\n'
		cat "${scratch}/examples-refusal.txtar"
	} >"${repo}/testdata/script/refusal.txtar"
	(cd "${repo}" && just check) >"${scratch}/examples-failed-check.out" 2>&1 || status=$?
	[[ "${status}" != 0 ]] || fct_fail 'the notice wrapper hides a failed scenario'
	grep -qF 'notice-regression' "${scratch}/examples-failed-check.out" ||
		fct_fail 'just check failed before the broken scenario'
	grep -qF "${skipped}" "${scratch}/examples-failed-check.out" ||
		fct_fail 'a failed just check hides the skip reason'
	mv "${scratch}/examples-refusal.txtar" "${repo}/testdata/script/refusal.txtar"
	mv "${scratch}/examples-manifest.toml" "${repo}/manifest.toml"
}

# fct_test_uninstall removes the install that just install made in the data
# folder $3 for the rendered repository $1. ncly then finds no root $4 of the
# slug $2, and a second just uninstall finds nothing to remove.
fct_test_uninstall() {
	local repo="${1}" slug="${2}" extensions="${3}/nicely/extensions" extension_root="${4}" scratch="${5}" root_words file status=0
	read -r -a root_words <<<"${extension_root}"
	(cd "${repo}" && just uninstall) >"${scratch}/uninstall.out"
	[[ "$(head -n 1 "${scratch}/uninstall.out")" == "removed ${extensions}/ncly-${slug}" ]] ||
		fct_fail "just uninstall does not remove the executable first: $(cat "${scratch}/uninstall.out")"
	for file in "ncly-${slug}" "ncly-${slug}.toml" "ncly-${slug}.locales" "ncly-${slug}.skill"; do
		fct_has "${scratch}/uninstall.out" "removed ${extensions}/${file}"
	done
	[[ -z "$(ls -A "${extensions}")" ]] || fct_fail "just uninstall leaves $(ls -A "${extensions}") in ${extensions}"
	PATH="/usr/bin:${PATH}" "$(command -v go)" -C "${repo}" tool ncly describe "${root_words[@]}" --json >/dev/null 2>"${scratch}/describe.err" || status=$?
	if [[ "${status}" != 2 ]] || ! grep -qF '"code":"NOT_FOUND"' "${scratch}/describe.err"; then
		fct_fail "ncly describe ${extension_root} exits ${status} after just uninstall, want 2 with NOT_FOUND: $(cat "${scratch}/describe.err")"
	fi
	(cd "${repo}" && just uninstall) >"${scratch}/uninstall.out"
	fct_has "${scratch}/uninstall.out" "uninstall: nothing to remove, since ${extensions} holds no ncly-${slug}"
	fct_test_install_order "${repo}" "${slug}" "${extensions}" "${scratch}"
	fct_test_signoff "${repo}" "${scratch}"
}

# fct_test_install_order blocks manifest publication in $3, then tries just install in $1.
# A failed sidecar publication must leave no executable of the slug $2.
fct_test_install_order() {
	local repo="${1}" slug="${2}" extensions="${3}" scratch="${4}" status=0
	mkdir "${extensions}/ncly-${slug}.toml"
	(cd "${repo}" && just install) >"${scratch}/blocked-install.out" 2>&1 || status=$?
	[[ "${status}" != 0 ]] || fct_fail 'just install accepts a directory in place of its manifest'
	grep -qF "${extensions}/ncly-${slug}.toml" "${scratch}/blocked-install.out" ||
		fct_fail "just install failed before manifest publication: $(cat "${scratch}/blocked-install.out")"
	[[ ! -e "${extensions}/ncly-${slug}" && ! -L "${extensions}/ncly-${slug}" ]] ||
		fct_fail 'just install published the executable before its sidecars'
	mv "${extensions}/ncly-${slug}.toml" "${scratch}/blocked-manifest"
}

# fct_test_acceptance runs just acceptance in the rendered repository $1 on
# the input that it makes. Then it runs it from another folder on a read-only
# fixture, given by a relative path, which keeps its content.
fct_test_acceptance() {
	local repo="${1}" scratch="${2}" fixture="${2}/fixtures/notes.md"
	# The author fills acceptance.sh, so it starts free of findings.
	shellcheck "${repo}/acceptance.sh"
	shfmt -d "${repo}/acceptance.sh"
	if ! (cd "${repo}" && just acceptance) >"${scratch}/acceptance.out" 2>&1; then
		cat "${scratch}/acceptance.out" >&2
		fct_fail 'just acceptance fails on the input that it makes'
	fi
	grep -q '^acceptance: ncly .* passed on .*/work/notes\.md$' "${scratch}/acceptance.out" || fct_fail "just acceptance reports no pass: $(cat "${scratch}/acceptance.out")"
	mkdir -p "${scratch}/fixtures"
	printf 'one\ntwo\n' >"${fixture}"
	chmod 444 "${fixture}"
	if ! (cd "${scratch}/fixtures" && just --justfile "${repo}/justfile" acceptance notes.md) >"${scratch}/acceptance.out" 2>&1; then
		cat "${scratch}/acceptance.out" >&2
		fct_fail 'just acceptance fails on a fixture'
	fi
	grep -qF '"size_bytes":8' "${scratch}/acceptance.out" || fct_fail "just acceptance did not run on the fixture: $(cat "${scratch}/acceptance.out")"
	[[ "$(cat "${fixture}")" == $'one\ntwo' ]] || fct_fail 'just acceptance changes the fixture'
}
