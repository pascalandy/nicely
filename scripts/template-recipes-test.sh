#!/usr/bin/env bash
# Sourced by template-test.sh to prove the files and the recipes that a
# rendered repository adds to the extension.

# shellcheck source=scripts/template-install-test.sh
source "${BASH_SOURCE[0]%/*}/template-install-test.sh"

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

# fct_test_analyzers proves that just check runs the analyzers of the pinned
# SDK on the code of the author. Its probe repeats the line that 3e6a4eff
# removed from extensions/agent/execute.go.
fct_test_analyzers() {
	local repo="${1}" scratch="${2}" domain="${3}" probe="${1}/analyzer_probe.go"
	printf 'package %s\n\nimport (\n\t"context"\n\t"io"\n\n\t"github.com/pascalandy/nicely/sdk/runner"\n)\n\nfunc probe(ctx context.Context, path string) error {\n\t_, err := runner.Run(ctx, runner.Program{Path: path, Stdout: io.Discard})\n\treturn err\n}\n' "${domain}" >"${probe}"
	if (cd "${repo}" && just check) >"${scratch}/analyzers.out" 2>&1; then
		fct_fail 'just check accepts io.Discard as the Stdout of a runner.Program'
	fi
	grep -qF 'analyzer_probe.go:11:63: runner.Program.Stdout: use nil or omit the field instead of io.Discard' "${scratch}/analyzers.out" ||
		fct_fail "just check fails without the diagnostic of runnerdiscard: $(cat "${scratch}/analyzers.out")"
	rm "${probe}"
}

# fct_test_gates proves the gates that just check of the rendered repository
# $1 adds: the analyzers, and ## Publication of its command $4.
fct_test_gates() {
	fct_test_analyzers "${1}" "${2}" "${3}"
	fct_test_publication "${1}" "${2}" "${4}"
}

# fct_test_publication proves that just check reads ## Publication of
# spec.md through the kit. The rendered repository $1, whose command is $3,
# declares that this command writes a user file, and gives it no row. Its
# effects in manifest.toml say so, as the builds of #425 did.
fct_test_publication() {
	local repo="${1}" scratch="${2}" command="${3}"
	cp "${repo}/manifest.toml" "${scratch}/publication-manifest.toml"
	fct_edit "${repo}/manifest.toml" "effects = {}" "effects = { user_writes = true }"
	if (cd "${repo}" && just check) >"${scratch}/publication.out" 2>&1; then
		fct_fail 'just check accepts a command that writes a user file with no row in ## Publication'
	fi
	grep -qF "spec.md: \`ncly ${command}\` may write a user file, because its effects set user_writes or are absent, and ## Publication has no row for it." "${scratch}/publication.out" ||
		fct_fail "just check fails without the problem of ## Publication: $(cat "${scratch}/publication.out")"
	cp "${scratch}/publication-manifest.toml" "${repo}/manifest.toml"
	fct_has "${repo}/spec.md" "## Publication"
}

# fct_test_example_notices requires notices from plain just check in the rendered repository $1, whose slug is $3.
fct_test_example_notices() {
	local repo="${1}" scratch="${2}" status=0 line examples count
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
	# awk counts the rule index apart from the SDK, which prints measure 1 of D068 for owner C.
	count="$(awk -F'|' '/^\| Rule \| Enforced by \|$/ { rows = 1; next }
		rows && /^\|---/ { next }
		rows && /^\|/ { n++; if ($3 ~ /^ Prose only/) only++; if ($3 ~ /^ Prose on purpose/) purpose++; next }
		{ rows = 0 }
		END { printf "%d of %d rules are Prose only, and %d Prose on purpose", only, n, purpose }' "${repo}/AGENTS/rule-index.md" "${repo}"/AGENTS/rule-index/*.md)"
	[[ "${count}" != "0 of 0 "* ]] || fct_fail 'the rendered rule index holds no rule'
	grep -qxF "AGENTS/rule-index.md: ${count}" "${scratch}/examples-check.err" ||
		fct_fail "plain just check does not print: AGENTS/rule-index.md: ${count}"
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

# fct_test_acceptance runs just acceptance in the rendered repository $1 on
# the input that it makes, also under settings of the caller that the owned
# run drops. Then it runs it from another folder on a read-only fixture,
# given by a relative path, which keeps its content.
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
	# At the parent of 09809608, a scratch HOME kept the other settings of the caller, such as this NCLY_CONFIG.
	printf '[invalid\n' >"${scratch}/caller-config.toml"
	if ! (cd "${repo}" && NCLY_CONFIG="${scratch}/caller-config.toml" NCLY_LANG=xx-YY just acceptance) >"${scratch}/acceptance.out" 2>&1; then
		cat "${scratch}/acceptance.out" >&2
		fct_fail 'just acceptance runs with the NCLY_CONFIG and NCLY_LANG of the caller'
	fi
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

# fct_test_recipe_lines requires one command line in each recipe of the
# justfile $1, so a copied recipe only wires a command to the SDK or an owned
# script.
fct_test_recipe_lines() {
	awk '
		/^[^[:space:]#].*:[[:space:]]*$/ { recipe = $1; sub(/:$/, "", recipe); lines = 0; reported = 0 }
		/^[[:space:]]+#!/ || /^[[:space:]]+[^[:space:]#]/ {
			if ((++lines > 1 || $1 ~ /^#!/) && !reported++) {
				print "template-test.sh: recipe " recipe " must have one command line" > "/dev/stderr"
				failed = 1
			}
		}
		END { exit failed }
	' "${1}"
}
