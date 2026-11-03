#!/usr/bin/env bash
# Sourced by template-test.sh to prove just signoff of the rendered
# repository. fct_test_uninstall calls it after its install tests.
# Markdown backticks in the fixture text are literal.
# shellcheck disable=SC2016

# Run the rendered recipe with real Git and just. Only gh is a stub.
fct_test_signoff() {
	local repo="${1}" scratch="${2}" ready="${2}/signoff-ready" sha file status=0
	fct_signoff_stub "${scratch}"
	# Signoff runs just finish first, so it signs no commit before the loop of the skill is done (#413).
	fct_signoff "${repo}" "${scratch}" || status=$?
	if [[ "${status}" == 0 ]] || ! grep -qF 'this line of the form is not filled' "${scratch}/signoff.out"; then
		cat "${scratch}/signoff.out" >&2
		fct_fail 'just signoff signs a commit that just finish refuses'
	fi
	[[ ! -e "${scratch}/signed-sha" ]] || fct_fail 'just signoff called gh signoff before just finish passed'

	# Signoff runs the first failing run again at its Scenario commit (#422).
	# An invented command, a compile error, and a scenario that passes sign nothing.
	fct_signoff_ready "${repo}" "${scratch}" "${scratch}/signoff-echo" read
	fct_signoff_refuses "${scratch}" "${scratch}/signoff-echo" 'echo invented' 'names the scenario "echo invented", which is no name of a scenario'
	fct_signoff_ready "${repo}" "${scratch}" "${scratch}/signoff-compile" read 'package broken'$'\n''func {'
	fct_signoff_refuses "${scratch}" "${scratch}/signoff-compile" first 'before any line of the scenario, such as on a compile error'
	fct_signoff_ready "${repo}" "${scratch}" "${scratch}/signoff-passes" inspect
	fct_signoff_refuses "${scratch}" "${scratch}/signoff-passes" first 'passes at its Scenario commit'
	# The run must fail on the line that the record names, not on another line of the scenario.
	fct_signoff_ready "${repo}" "${scratch}" "${scratch}/signoff-line" read
	fct_signoff_refuses "${scratch}" "${scratch}/signoff-line" first 'and first_run of review.toml names the assertion' \
		"$(grep -m 1 '^refuses ' "${scratch}/signoff-line/testdata/script/first.txtar")"

	fct_signoff_ready "${repo}" "${scratch}" "${ready}" read
	fct_signoff_review "${ready}" first
	sha="$(git -C "${ready}" rev-parse HEAD)"
	if ! fct_signoff "${ready}" "${scratch}"; then
		cat "${scratch}/signoff.out" >&2
		fct_fail 'rendered just signoff failed'
	fi
	fct_has "${scratch}/signoff.out" "Signed off on ${sha}"
	grep -qF 'signoff: the first scenario fails again at' "${scratch}/signoff.out" || fct_fail 'just signoff did not run the first failing run again'
	fct_has "${scratch}/signed-sha" "${sha}"

	# just scenario runs one scenario, without the coverage of the others, and refuses a name of no scenario.
	(cd "${ready}" && just scenario first) >"${scratch}/scenario.out" 2>&1 || {
		cat "${scratch}/scenario.out" >&2
		fct_fail 'just scenario first fails on a scenario that passes'
	}
	if (cd "${ready}" && just scenario ../first) >"${scratch}/scenario.out" 2>&1 || ! grep -qF '"../first" is no name of a scenario' "${scratch}/scenario.out"; then
		cat "${scratch}/scenario.out" >&2
		fct_fail 'just scenario accepts a path as the name of a scenario'
	fi
}

# fct_signoff_ready copies the rendered repository $1 to $3, with the steps of
# round 1 and filled forms, in the scratch folder $2. Its first scenario
# expects a message that ends with "can $4". "read" fails until the handler
# changes the message, and "inspect" passes at once. A Go file $5 in the
# Scenario commit, removed by the handler, breaks the build of that commit.
fct_signoff_ready() {
	local ready="${3}" file refuses expected
	cp -R "${1}" "${ready}"
	git -C "${ready}" commit -q --allow-empty -m 'step: intake'
	git -C "${ready}" commit -q --allow-empty -m 'step: spec'
	refuses="$(grep -m 1 '^refuses USAGE_INVALID ncly .* missing.md --json$' "${ready}/testdata/script/inspect.txtar")"
	expected="$(grep -m 1 "^stderr 'missing.md is no file that .* can inspect\[.\]'$" "${ready}/testdata/script/inspect.txtar")"
	printf '# A missing file names the cause in plain words.\n# spec: bad-input\n%s\n%s\n' "${refuses}" "${expected/can inspect/can ${4}}" >"${ready}/testdata/script/first.txtar"
	[[ -z "${5:-}" ]] || printf '%s\n' "${5}" >"${ready}/broken.go"
	git -C "${ready}" add -A
	git -C "${ready}" commit -qm 'step: scenario'
	[[ -z "${5:-}" ]] || git -C "${ready}" rm -q broken.go
	if [[ "${4}" != inspect ]]; then
		for file in locales/en.toml testdata/script/inspect.txtar; do
			sed "s/ can inspect\([.[]\)/ can ${4}\1/" "${ready}/${file}" >"${2}/form.md"
			mv "${2}/form.md" "${ready}/${file}"
		done
	fi
	git -C "${ready}" commit -q --allow-empty -am 'handler'
	fct_fill "${ready}" "${2}"
	git -C "${ready}" commit -qam 'fill the forms'
}

# fct_signoff_review commits the readiness review of the head of $1, whose
# first failing run names the scenario $2, and the assertion $3 or the line
# stderr of the first scenario. Its verdict is $4, or Ready.
fct_signoff_review() {
	local expected="${3:-}"
	[[ -n "${expected}" ]] || expected="$(grep -m 1 '^stderr ' "${1}/testdata/script/first.txtar")"
	finish="${1}" ledger_run=$'[first_run]\ncommit = "'"$(fct_step "${1}" scenario)"$'"\nscenario = "'"${2}"$'"\nexit = 1\nassertion = \'\'\''"${expected}"$' found no match\'\'\'' \
		fct_readiness "$(git -C "${1}" rev-parse HEAD)" "${4:-Ready}" ''
	fct_review "${1}" 'readiness review'
}

# fct_signoff_refuses requires that just signoff of the repository $2, whose
# first failing run names the scenario $3 and the assertion $5, fails with the
# text $4 and signs nothing.
fct_signoff_refuses() {
	local status=0
	fct_signoff_review "${2}" "${3}" "${5:-}"
	fct_signoff "${2}" "${1}" || status=$?
	if [[ "${status}" == 0 ]] || ! grep -qF -- "${4}" "${1}/signoff.out"; then
		cat "${1}/signoff.out" >&2
		fct_fail "just signoff with the first run \"${3}\" must fail with \"${4}\""
	fi
	[[ ! -e "${1}/signed-sha" ]] || fct_fail "just signoff signed a commit whose first run is \"${3}\""
}

# fct_signoff_stub writes the stub gh of the scratch folder $1. It signs only
# the detached HEAD of the scratch worktree of signoff.
fct_signoff_stub() {
	mkdir -p "${1}/signoff-bin"
	cat >"${1}/signoff-bin/gh" <<'SCRIPT'
#!/usr/bin/env bash
set -Eeuo pipefail
if [[ "$*" == --version ]]; then echo 'gh version test'; exit 0; fi
[[ "$#" == 3 && "$1" == signoff && "$2" == --commit ]]
[[ "$(git rev-parse HEAD)" == "$3" ]]
if git symbolic-ref --quiet HEAD; then exit 1; fi
printf '%s\n' "$3" >"${NCLY_TEMPLATE_SIGNOFF}"
SCRIPT
	chmod 755 "${1}/signoff-bin/gh"
}

# fct_signoff runs just signoff in the repository $1, with the stub gh of the scratch folder $2.
fct_signoff() {
	(cd "${1}" && PATH="${2}/signoff-bin:${PATH}" XDG_CACHE_HOME="${2}/signoff-cache" \
		NCLY_TEMPLATE_SIGNOFF="${2}/signed-sha" just signoff) >"${2}/signoff.out" 2>&1
}
