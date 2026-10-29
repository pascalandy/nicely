#!/usr/bin/env bash
# Sourced by template-test.sh to exercise updates of a rendered repository.

fct_test_update() {
	local src="${1}" repo="${2}" scratch="${3}" root="${4}" status dependency lipgloss
	# The author edits two files, and the next template revision changes one of them elsewhere.
	fct_edit "${repo}/vision.md" "" "Author's note: kept through each update."
	fct_edit "${repo}/AGENTS.md" "# AGENTS.md" "# AGENTS.md of the author"
	if [[ -z "$(sed -n '/^\[/q; /^name = /p' "${repo}/manifest.toml")" ]]; then
		# Simulate a repository rendered before name existed.
		sed '/^name:/d' "${repo}/.copier-answers.yml" >"${repo}/answers.new"
		mv "${repo}/answers.new" "${repo}/.copier-answers.yml"
	fi
	# The author adds a direct dependency, as #347 U1 found, so the require line of the template becomes a block.
	# golang.org/x/text is in the module graph of Nicely, so go mod tidy needs no network.
	mkdir "${repo}/internal/words"
	printf 'package words\n\nimport _ "golang.org/x/text/language"\n' >"${repo}/internal/words/words.go"
	go -C "${repo}" mod tidy
	dependency="$(grep -E '^[[:space:]]golang\.org/x/text v[^ ]+$' "${repo}/go.mod")" ||
		fct_fail "go mod tidy does not require golang.org/x/text as a direct dependency"
	# The author's replace of Lip Gloss differs from Nicely's. It names the folder of the fork in the module cache, so it builds offline.
	lipgloss="$(grep '^replace charm.land/lipgloss/v2 ' "${root}/go.mod")" ||
		fct_fail "Nicely's go.mod replaces no Lip Gloss: remove this case with the replace of template/go.mod.jinja"
	go -C "${repo}" mod edit -replace "charm.land/lipgloss/v2=$(go -C "${repo}" list -m -f '{{.Dir}}' charm.land/lipgloss/v2)"
	# The go.mod of Nicely v0.0.1 also replaces example.com/dropped, which v0.0.2 no longer replaces. just update reads only that go.mod.
	mkdir "${scratch}/nicely-v0.0.1"
	{
		cat "${root}/go.mod"
		printf 'replace example.com/dropped => example.com/dropped v1.0.0\n'
	} >"${scratch}/nicely-v0.0.1/go.mod"
	# A TMPDIR of macOS ends with a slash, and // starts a comment in go.mod, so the path is quoted.
	printf 'replace github.com/pascalandy/nicely v0.0.1 => "%s"\nreplace example.com/dropped => example.com/dropped v1.0.0\n' "${scratch}/nicely-v0.0.1" >>"${repo}/go.mod"
	fct_commit "customized"
	fct_edit "${src}/template/AGENTS.md.jinja" "" "This line comes from template v0.0.2."
	fct_revision v0.0.2
	# just update ends with just check, so its exit code 0 proves the check of the rendered repository.
	fct_update v0.0.2 || {
		cat "${scratch}/update.out" >&2
		git -C "${repo}" status --short >&2
		fct_fail "just update v0.0.2 failed on a customized repository"
	}
	fct_said "go test -count=1 ./..."
	fct_has "${repo}/vision.md" "Author's note: kept through each update."
	fct_has "${repo}/AGENTS.md" "# AGENTS.md of the author"
	fct_has "${repo}/AGENTS.md" "This line comes from template v0.0.2."
	fct_has "${repo}/.copier-answers.yml" "_commit: v0.0.2"
	[[ -z "$(git -C "${repo}" diff --name-only --diff-filter=U)" ]] ||
		fct_fail "just update v0.0.2 left unmerged files: $(git -C "${repo}" diff --name-only --diff-filter=U)"
	fct_has "${repo}/go.mod" "${dependency}"
	[[ "$(go -C "${repo}" list -m -f '{{.Version}}' github.com/pascalandy/nicely)" == v0.0.2 ]] ||
		fct_fail "just update v0.0.2 did not move the Nicely pin of go.mod to v0.0.2"
	fct_has "${repo}/go.mod" "replace github.com/pascalandy/nicely => ${root}"
	fct_has "${repo}/go.mod" "${lipgloss}"
	if grep -qF example.com/dropped "${repo}/go.mod"; then fct_fail "just update kept the replace of example.com/dropped, which Nicely dropped"; fi
	fct_commit "updated to v0.0.2"

	# The author and template v0.0.3 change the same line. The stub go logs each call, so a go get after the conflict shows.
	fct_edit "${repo}/AGENTS.md" "This line comes from template v0.0.2." "The author rewrote this line."
	fct_commit "rewritten"
	fct_edit "${src}/template/AGENTS.md.jinja" "This line comes from template v0.0.2." "Template v0.0.3 rewrote this line."
	fct_revision v0.0.3
	mkdir "${scratch}/bin"
	printf '#!/bin/sh\necho "go $*" >>"%s/go.log"\nexit 1\n' "${scratch}" >"${scratch}/bin/go"
	chmod 755 "${scratch}/bin/go"
	status=0
	PATH="${scratch}/bin:${PATH}" fct_update v0.0.3 || status=$?
	[[ "${status}" == 1 ]] || fct_fail "just update v0.0.3 exited with ${status} on a conflict, want 1"
	[[ ! -e "${scratch}/go.log" ]] || fct_fail "just update ran go after a conflict: $(cat "${scratch}/go.log")"
	fct_has "${repo}/AGENTS.md" "<<<<<<< before updating"
	fct_said "  AGENTS.md"
	fct_said "  go get github.com/pascalandy/nicely@v0.0.3"
	fct_said "  go mod tidy"
	fct_said "  just check"

	# On a Git tree with changes, just update stops before Copier.
	status=0
	fct_update v0.0.3 || status=$?
	[[ "${status}" == 1 ]] || fct_fail "just update exited with ${status} on a Git tree with changes, want 1"
	fct_said "update: the update needs a clean Git tree. Commit your changes, then run just update v0.0.3 again"

	# When Copier fails, just update says how to return to the commit before the update.
	git -C "${repo}" reset -q --hard
	status=0
	fct_update v9.9.9 || status=$?
	[[ "${status}" == 1 ]] || fct_fail "just update v9.9.9 exited with ${status} for a version that does not exist, want 1"
	fct_said "  git restore --source=$(git -C "${repo}" rev-parse HEAD) --staged --worktree -- ."
	fct_said "  just update v9.9.9"
}
