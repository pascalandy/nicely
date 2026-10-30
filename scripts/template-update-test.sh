#!/usr/bin/env bash
# Sourced by template-test.sh to exercise target-version updates and the old-recipe bootstrap.

fct_copy_source() {
	git -C "${root}" ls-files --cached --others --exclude-standard -z |
		tar -C "${root}" -cf - --null -T - | tar -C "${1}" -xf -
}

# Publish a scratch tag to the file proxy, without touching the user's module cache.
fct_proxy_revision() {
	local src="${1}" version="${2}" proxy="${3}"
	git -C "${src}" show "${version}:go.mod" >"${proxy}/${version}.mod"
	printf '{"Version":"%s","Time":"2000-01-01T00:00:00Z"}\n' "${version}" >"${proxy}/${version}.info"
	git -C "${src}" archive --format=zip --prefix="github.com/pascalandy/nicely@${version}/" "${version}" -- . ':!tools' ':!testdata/extensions/standalone' ':!testdata/extensions/standalone-named' >"${proxy}/${version}.zip"
	printf '%s\n' "${version}" >>"${proxy}/list"
}

# Run the bootstrap exactly as the bundled skill documents it.
fct_bootstrap() {
	local version="${1}" mode="${2:-}" command status=0
	command="$(sed -n '/^go -C .* update VERSION /p' "${root}/extensions/skill/nicely-extensions/SKILL.md")"
	[[ -n "${command}" && "${command}" != *$'\n'* ]] || fct_fail 'the skill must document one bootstrap command'
	command="${command//VERSION/${version}}"
	(cd "${repo}" && bash -c "${command}${mode:+ --continue}") >"${scratch}/update.out" 2>&1 || status=$?
	return "${status}"
}

fct_test_update() {
	local src="${1}" repo="${2}" scratch="${3}" root="${4}" status dependency lipgloss
	fct_edit "${repo}/vision.md" "" "Author's note: kept through each update."
	fct_edit "${repo}/AGENTS.md" "# AGENTS.md" "# AGENTS.md of the author"
	if [[ -z "$(sed -n '/^\[/q; /^name = /p' "${repo}/manifest.toml")" ]]; then
		sed '/^name:/d' "${repo}/.copier-answers.yml" >"${repo}/answers.new"
		mv "${repo}/answers.new" "${repo}/.copier-answers.yml"
	fi
	# The author's direct dependency and unrelated replacement survive the update.
	mkdir "${repo}/internal/words"
	printf 'package words\n\nimport _ "golang.org/x/text/language"\n' >"${repo}/internal/words/words.go"
	go -C "${repo}" mod tidy
	dependency="$(grep -E '^[[:space:]]golang\.org/x/text v[^ ]+$' "${repo}/go.mod")" ||
		fct_fail 'go mod tidy does not require golang.org/x/text directly'
	lipgloss="$(grep '^replace charm.land/lipgloss/v2 ' "${root}/go.mod")" ||
		fct_fail 'Nicely no longer replaces Lip Gloss: revise this replacement case'
	go -C "${repo}" mod edit -replace "charm.land/lipgloss/v2=$(go -C "${repo}" list -m -f '{{.Dir}}' charm.land/lipgloss/v2)"
	go -C "${repo}" mod edit -replace example.com/author=example.com/author@v1.0.0
	mkdir "${scratch}/nicely-v0.0.1"
	{
		cat "${root}/go.mod"
		printf 'replace example.com/dropped => example.com/dropped v1.0.0\n'
	} >"${scratch}/nicely-v0.0.1/go.mod"
	printf 'replace github.com/pascalandy/nicely v0.0.1 => "%s"\nreplace example.com/dropped => example.com/dropped v1.0.0\n' "${scratch}/nicely-v0.0.1" >>"${repo}/go.mod"
	fct_commit customized
	fct_edit "${src}/template/AGENTS.md.jinja" "" "This line comes from template v0.0.2."
	fct_revision v0.0.2
	fct_update v0.0.2 || {
		cat "${scratch}/update.out" >&2
		fct_fail 'just update v0.0.2 failed'
	}
	fct_said 'go test -count=1 ./...'
	fct_has "${repo}/vision.md" "Author's note: kept through each update."
	fct_has "${repo}/AGENTS.md" '# AGENTS.md of the author'
	fct_has "${repo}/AGENTS.md" 'This line comes from template v0.0.2.'
	fct_has "${repo}/.copier-answers.yml" '_commit: v0.0.2'
	fct_has "${repo}/go.mod" "${dependency}"
	[[ "$(go -C "${repo}" list -m -f '{{.Version}}' github.com/pascalandy/nicely)" == v0.0.2 ]] || fct_fail 'the SDK pin did not move'
	fct_has "${repo}/go.mod" "replace github.com/pascalandy/nicely => ${root}"
	fct_has "${repo}/go.mod" "${lipgloss}"
	fct_has "${repo}/go.mod" 'replace example.com/author => example.com/author v1.0.0'
	if grep -qF example.com/dropped "${repo}/go.mod"; then fct_fail 'update kept a replacement that Nicely dropped'; fi
	fct_commit 'updated to v0.0.2'

	# Copier changes the same line as the author. No dependency file can change before resolution.
	fct_edit "${repo}/AGENTS.md" 'This line comes from template v0.0.2.' 'The author rewrote this line.'
	fct_commit rewritten
	fct_edit "${src}/template/AGENTS.md.jinja" 'This line comes from template v0.0.2.' 'Template v0.0.3 rewrote this line.'
	fct_revision v0.0.3
	cp "${repo}/go.mod" "${scratch}/before.mod"
	cp "${repo}/go.sum" "${scratch}/before.sum"
	status=0
	fct_update v0.0.3 || status=$?
	[[ "${status}" == 1 ]] || fct_fail "just update exits ${status} on a conflict, want 1"
	cmp "${scratch}/before.mod" "${repo}/go.mod" || fct_fail 'update changed go.mod on a conflict'
	cmp "${scratch}/before.sum" "${repo}/go.sum" || fct_fail 'update changed go.sum on a conflict'
	fct_has "${repo}/AGENTS.md" '<<<<<<< before updating'
	fct_said '  AGENTS.md'
	grep -q -- '--continue$' "${scratch}/update.out" || fct_fail 'update printed no continuation command'
	status=0
	fct_update v0.0.3 || status=$?
	[[ "${status}" == 1 ]] || fct_fail 'update accepted a dirty tree'
	fct_said 'update: the update needs a clean Git tree. Commit your changes, then run just update v0.0.3 again'
	# The scratch repository alone returns to its committed state for the Copier failure case.
	git -C "${repo}" reset -q --hard
	status=0
	# The target module exists, but this Copier source lacks its tag.
	git -C "${src}" tag -d v0.0.3 >/dev/null
	fct_update v0.0.3 || status=$?
	[[ "${status}" == 1 ]] || fct_fail 'update accepted a missing Copier tag'
	fct_said "  git restore --source=$(git -C "${repo}" rev-parse HEAD) --staged --worktree -- ."
	grep -q 'run ./sdk/extensionrepo/cmd/extensionrepo update v0.0.3 ' "${scratch}/update.out" || fct_fail 'Copier failure printed no bootstrap retry'
}

fct_test_bootstrap() {
	local scratch="${1}" src="${1}/source" repo version=v0.99.0 status recovery copier
	shift
	copier=("$@")
	mkdir -p "${src}"
	# The old render uses the authentic v0.23.2 template, including its copied updater.
	git -C "${root}" archive v0.23.2 | tar -C "${src}" -xf -
	git -C "${src}" init -q
	fct_revision v0.23.2
	repo="${scratch}/untouched old render"
	"${copier[@]}" copy --quiet --vcs-ref v0.23.2 --defaults --data domain=probe --data module=example.com/ncly-probe "${src}" "${repo}"
	git -C "${repo}" init -q
	fct_commit 'untouched v0.23.2 render'
	# Only the source advances. Nothing in the rendered repository changes before bootstrap.
	fct_copy_source "${src}"
	printf '\nreplace example.com/target => example.com/target v1.0.0\n' >>"${src}/go.mod"
	fct_revision "${version}"
	fct_bootstrap "${version}" || {
		cat "${scratch}/update.out" >&2
		fct_fail 'the untouched v0.23.2 bootstrap failed'
	}
	fct_has "${repo}/.copier-answers.yml" "_commit: ${version}"
	fct_has "${repo}/go.mod" 'replace example.com/target => example.com/target v1.0.0'
	[[ "$(go -C "${repo}" list -m -f '{{.Version}}' github.com/pascalandy/nicely)" == "${version}" ]] || fct_fail 'bootstrap did not move the SDK pin'
	fct_said 'go test -count=1 ./...'
	printf 'untouched v0.23.2 bootstrap passes through target %s and reconciles replacements\n' "${version}"

	# A second old render plants a real Copier conflict before the same bootstrap.
	repo="${scratch}/conflicted old render"
	"${copier[@]}" copy --quiet --vcs-ref v0.23.2 --defaults --data domain=probe --data module=example.com/ncly-probe "${src}" "${repo}"
	fct_edit "${repo}/AGENTS.md" '# AGENTS.md' '# AGENTS.md of the author'
	git -C "${repo}" init -q
	fct_commit 'author edits on v0.23.2'
	fct_edit "${src}/template/AGENTS.md.jinja" '# AGENTS.md' '# AGENTS.md of the target'
	version=v0.99.1
	fct_revision "${version}"
	cp "${repo}/go.mod" "${scratch}/before.mod"
	status=0
	fct_bootstrap "${version}" || status=$?
	[[ "${status}" == 1 ]] || fct_fail "bootstrap exits ${status} on a Copier conflict, want 1"
	fct_said '  AGENTS.md'
	cmp "${scratch}/before.mod" "${repo}/go.mod" || fct_fail 'bootstrap changed dependencies before conflict resolution'
	recovery="$(sed -n '/^  go -C /s/^  //p' "${scratch}/update.out")"
	[[ "${recovery}" == *' --continue' ]] || fct_fail 'bootstrap printed no continuation command'
	# Markers still count after staging removes the unmerged index entries.
	git -C "${repo}" add AGENTS.md
	status=0
	(cd "${repo}" && bash -c "${recovery}") >"${scratch}/update.out" 2>&1 || status=$?
	[[ "${status}" == 1 ]] || fct_fail 'continuation accepted staged conflict markers'
	fct_said '  AGENTS.md'
	fct_edit "${repo}/AGENTS.md" '<<<<<<< before updating' ''
	fct_edit "${repo}/AGENTS.md" '=======' ''
	fct_edit "${repo}/AGENTS.md" '>>>>>>> after updating' ''
	git -C "${repo}" add AGENTS.md
	printf 'unresolved\n' >"${repo}/AGENTS.md.rej"
	status=0
	(cd "${repo}" && bash -c "${recovery}") >"${scratch}/update.out" 2>&1 || status=$?
	[[ "${status}" == 1 ]] || fct_fail 'continuation accepted a .rej file'
	fct_said '  AGENTS.md.rej'
	mv "${repo}/AGENTS.md.rej" "${scratch}/resolved.rej"
	(cd "${repo}" && bash -c "${recovery}") >"${scratch}/update.out" 2>&1 || {
		cat "${scratch}/update.out" >&2
		fct_fail 'the printed continuation command failed'
	}
	fct_has "${repo}/.copier-answers.yml" "_commit: ${version}"
	fct_has "${repo}/go.mod" 'replace example.com/target => example.com/target v1.0.0'
	[[ "$(go -C "${repo}" list -m -f '{{.Version}}' github.com/pascalandy/nicely)" == "${version}" ]] || fct_fail 'continuation did not move the SDK pin'
	fct_said 'go test -count=1 ./...'
	printf 'v0.23.2 bootstrap recovers from Copier conflict, staged markers, and .rej through target %s\n' "${version}"
}
