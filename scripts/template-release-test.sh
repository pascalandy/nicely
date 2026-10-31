#!/usr/bin/env bash
# Sourced by template-test.sh. Only local bare remotes receive release tags.

# shellcheck source=scripts/template-release-helpers.sh
source "${BASH_SOURCE[0]%/*}/template-release-helpers.sh"

fct_test_release() {
	local repo="${1}" scratch="${2}" slug="${3}" extension_root="${4}" sha previous receipt dir status out_file out_bytes err_file err_bytes i examples command
	local remote="${2}/release-origin.git" fixture="${2}/release-fixture/input.md"
	fct_has "${repo}/AGENTS.md" '## Release'
	fct_has "${repo}/README.md" "ncly describe ${extension_root} --json"
	fct_has "${repo}/README.md" "git clone --branch TAG https://example.com/ncly-${slug} ncly-${slug}"
	[[ "$(just --justfile "${repo}/justfile" --show release | grep -c '^    ')" == 1 ]] || fct_fail 'just release must have one command'
	just --justfile "${repo}/justfile" --show release | grep -q '^    go run ./internal/export release ' || fct_fail 'just release must call the SDK program'
	git init --bare -q "${remote}"
	git -C "${repo}" remote add origin "${remote}"
	git -C "${repo}" push -q origin HEAD:main
	git -C "${remote}" symbolic-ref HEAD refs/heads/main
	git config --file "${GIT_CONFIG_GLOBAL}" "url.${remote}.insteadOf" "https://example.com/ncly-${slug}"
	mkdir -p "${scratch}/release-bin" "${scratch}/release-fixture"
	printf '# Release input\nTwo lines\n' >"${fixture}"
	cat >"${scratch}/release-bin/gh" <<'SCRIPT'
#!/usr/bin/env bash
set -Eeuo pipefail
[[ "$#" == 4 && "$1" == api && "$2" == --paginate && "$3" == --slurp && "$4" == "repos/{owner}/{repo}/commits/$(git rev-parse HEAD)/check-runs" ]]
printf '[{"total_count":1,"check_runs":[{"name":"local-stub","head_sha":"%s","conclusion":"success"}]}]\n' "$(git rev-parse HEAD)"
SCRIPT
	chmod 755 "${scratch}/release-bin/gh"
	status=0
	fct_release_attempt "${repo}" "${scratch}" v0.1.0 || status=$?
	if [[ "${status}" == 0 ]] || ! grep -qF 'just signoff' "${scratch}/release.out"; then
		fct_fail 'release accepts a SHA without a receipt'
	fi
	[[ -z "$(git -C "${remote}" tag -l)" ]] || fct_fail 'receipt refusal pushed a tag'

	# Invalid receipt exits cannot authorize a release.
	fct_release_receipt "${repo}" "${scratch}"
	sha="$(git -C "${repo}" rev-parse HEAD)"
	receipt="${scratch}/release-cache/nicely/signoff/test-${sha}/receipt.json"
	fct_release_refused_receipts "${repo}" "${scratch}" "${receipt}"

	printf '\nDirty change\n' >>"${repo}/README.md"
	status=0
	fct_release_attempt "${repo}" "${scratch}" v0.1.0 || status=$?
	if [[ "${status}" == 0 ]] || ! grep -qF 'clean HEAD' "${scratch}/release.out"; then
		fct_fail 'release accepts a dirty tree'
	fi
	git -C "${repo}" restore README.md
	status=0
	fct_release_attempt "${repo}" "${scratch}" v0.1.0 missing-fixture.md || status=$?
	if [[ "${status}" == 0 ]] || ! grep -qF 'acceptance exited' "${scratch}/release.out"; then
		fct_fail 'release accepts failed acceptance'
	fi
	[[ -z "$(git -C "${remote}" tag -l)" ]] || fct_fail 'acceptance refusal pushed a tag'

	# Economy selects modes even when the first example already sets them.
	examples="$(grep -m1 '^examples = ' "${repo}/manifest.toml")"
	command="$(sed -n 's/^examples = \["ncly \(.*\) notes.md".*$/\1/p' "${repo}/manifest.toml" | head -n 1)"
	fct_edit "${repo}/manifest.toml" "${examples}" "examples = [\"ncly ${command} 'notes.md' --dry-run --json\", \"ncly ${command} notes.md\"]"
	git -C "${repo}" add manifest.toml

	# D129: exceeding a skill budget is a reported reservation, never a release block.
	for ((i = 0; i < 100; i++)); do printf 'Release budget sample.\n' >>"${repo}/skill/SKILL.md"; done
	git -C "${repo}" add skill/SKILL.md
	printf '\nUnpushed change\n' >>"${repo}/README.md"
	git -C "${repo}" add README.md
	git -C "${repo}" commit -qm 'unpushed release input'
	fct_release_receipt "${repo}" "${scratch}"
	status=0
	fct_release_attempt "${repo}" "${scratch}" v0.1.0 || status=$?
	if [[ "${status}" == 0 ]] || ! grep -qF 'pushed default branch' "${scratch}/release.out"; then
		fct_fail 'release accepts an unpushed SHA'
	fi
	git -C "${repo}" push -q origin HEAD:main
	fct_release_attempt "${repo}" "${scratch}" v0.1.0 || fct_fail 'first local release failed'
	receipt="$(sed -n 's/^release: receipt in //p' "${scratch}/release.out")"
	PATH="/usr/bin:${PATH}" jq -r '.economy[] | select(.command[1:3] != ["skill", "view"] and .command[1] != "describe" and .command[1] != "doctor") | "default Economy: \(.command | join(" ")), exit \(.exit)"' "${receipt}"
	PATH="/usr/bin:${PATH}" jq -e '[.economy[] | select(.command | index("notes.md"))] | length == 4 and all(.[]; .exit == 0)' "${receipt}" >/dev/null || fct_fail 'default release Economy did not run all four extension samples successfully'
	previous="$(git -C "${repo}" rev-parse HEAD)"

	# D095: a new tag in prose cannot repair an install block pinned to the previous tag.
	fct_edit "${repo}/README.md" "git clone --branch TAG https://example.com/ncly-${slug} ncly-${slug}" "git clone --branch v0.1.0 https://example.com/ncly-${slug} ncly-${slug}"
	printf '\nThe current release is v0.1.1.\n' >>"${repo}/README.md"
	git -C "${repo}" add README.md
	git -C "${repo}" commit -qm 'pin the README install to the previous tag'
	git -C "${repo}" push -q origin HEAD:main
	fct_release_receipt "${repo}" "${scratch}"
	status=0
	fct_release_attempt "${repo}" "${scratch}" v0.1.1 "${fixture}" || status=$?
	if [[ "${status}" == 0 ]] || ! grep -qF "README install cloned SHA ${previous}, want tagged SHA" "${scratch}/release.out"; then
		fct_fail 'D095: release accepts a README install of the previous tag'
	fi
	printf 'RED D095: previous-tag README rejected after publication, exit %s\n' "${status}"

	fct_edit "${repo}/README.md" "git clone --branch v0.1.0 https://example.com/ncly-${slug} ncly-${slug}" "git clone --branch TAG https://example.com/ncly-${slug} ncly-${slug}"
	git -C "${repo}" add README.md
	git -C "${repo}" commit -qm 'install the selected release tag'
	git -C "${repo}" push -q origin HEAD:main
	fct_release_receipt "${repo}" "${scratch}"
	sha="$(git -C "${repo}" rev-parse HEAD)"
	fct_release_attempt "${repo}" "${scratch}" v0.1.2 release-fixture/input.md || fct_fail 'correct README release failed'
	[[ "$(tail -n 1 "${scratch}/release.out")" == "Released v0.1.2 ${sha}" ]] || fct_fail 'release ends without the tag and SHA'
	receipt="$(sed -n 's/^release: receipt in //p' "${scratch}/release.out")"
	dir="${receipt%/*}"
	PATH="/usr/bin:${PATH}" jq -e '.ok == true' "${dir}/economy-6-json.stdout" >/dev/null || fct_fail 'Economy JSON mode returned no answer'
	PATH="/usr/bin:${PATH}" jq -e --argjson bytes "$(wc -c <"${fixture}" | tr -d ' ')" '.size_bytes == $bytes' "${dir}/economy-6-json.stdout" >/dev/null || fct_fail 'Economy did not replace the example input with the supplied fixture'
	if PATH="/usr/bin:${PATH}" jq -e . "${dir}/economy-6-text.stdout" >/dev/null 2>&1; then fct_fail 'Economy text mode returned JSON'; fi
	PATH="/usr/bin:${PATH}" jq -e --arg sha "${sha}" '.tag == "v0.1.2" and .sha == $sha and .cloned_sha == $sha and .pushed and (.failure == null) and .signoff.sha == $sha and .hosted_checks[0].check_runs[0].head_sha == $sha and any(.reservations[]; contains("exceeds a budget")) and (.economy | length == 14) and all(.economy[]; .exit == 0 and (.elapsed_ms | type) == "number") and ([.steps[] | select(.name == "not-found")][0].exit == 2)' "${receipt}" >/dev/null || fct_fail 'release receipt lacks published proof or Economy rows'
	while IFS=$'\t' read -r out_file out_bytes err_file err_bytes; do
		[[ "$(wc -c <"${dir}/${out_file}" | tr -d ' ')" == "${out_bytes}" ]] || fct_fail 'Economy stdout count differs from the actual output'
		[[ "$(wc -c <"${dir}/${err_file}" | tr -d ' ')" == "${err_bytes}" ]] || fct_fail 'Economy stderr count differs from the actual output'
	done < <(PATH="/usr/bin:${PATH}" jq -r '.economy[] | [.stdout, .stdout_bytes, .stderr, .stderr_bytes] | @tsv' "${receipt}")
	grep -qF '"code":"NOT_FOUND"' "${dir}/not-found.stderr" || fct_fail 'uninstall gives no NOT_FOUND'
	git -C "${remote}" for-each-ref --format='%(contents)' refs/tags/v0.1.2 >"${scratch}/tag-message.json"
	PATH="/usr/bin:${PATH}" jq -e --arg sha "${sha}" '.sha == $sha and .signoff.exit == 0 and ([.checks[].exit] | all(. == 0))' "${scratch}/tag-message.json" >/dev/null || fct_fail 'annotated tag lacks its successful signoff receipt'
	[[ -z "$(git -C "${repo}" status --porcelain)" ]] || fct_fail 'release wrote into tracked files'
	printf 'GREEN release receipt: correct README, Economy, and NOT_FOUND, exit 0\n'
}

# A Go module's /v2 suffix is not part of its published repository URL.
fct_test_release_url() {
	local template_src="${1}" scratch="${2}" copier
	shift 2
	copier=("$@")
	"${copier[@]}" copy --quiet --vcs-ref v0.0.1 --defaults --data domain=probe --data module=example.com/ncly-probe/v2 "${template_src}" "${scratch}/ncly-probe-v2"
	fct_has "${scratch}/ncly-probe-v2/README.md" 'git clone --branch TAG https://example.com/ncly-probe ncly-probe'
	if grep -qF 'example.com/ncly-probe/v2' "${scratch}/ncly-probe-v2/README.md"; then
		fct_fail 'README.md gives the module path example.com/ncly-probe/v2 as a repository URL'
	fi
}
