#!/usr/bin/env bash
# Sourced by template-test.sh to exercise updates of a rendered repository.

fct_test_update() {
	local src="${1}" repo="${2}" scratch="${3}" root="${4}" status
	# The author edits two files, and the next template revision changes one of them elsewhere.
	fct_edit "${repo}/vision.md" "" "Author's note: kept through each update."
	fct_edit "${repo}/AGENTS.md" "# AGENTS.md" "# AGENTS.md of the author"
	fct_commit "customized"
	fct_edit "${src}/template/AGENTS.md.jinja" "" "This line comes from template v0.0.2."
	fct_revision v0.0.2
	fct_update v0.0.2 || {
		cat "${scratch}/update.out" >&2
		fct_fail "just update v0.0.2 failed on a customized repository"
	}
	fct_has "${repo}/vision.md" "Author's note: kept through each update."
	fct_has "${repo}/AGENTS.md" "# AGENTS.md of the author"
	fct_has "${repo}/AGENTS.md" "This line comes from template v0.0.2."
	fct_has "${repo}/.copier-answers.yml" "_commit: v0.0.2"
	fct_has "${repo}/go.mod" "require github.com/pascalandy/nicely v0.0.2"
	fct_has "${repo}/go.mod" "replace github.com/pascalandy/nicely => ${root}"
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
