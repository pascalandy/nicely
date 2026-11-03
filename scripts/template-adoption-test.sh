#!/usr/bin/env bash
# Sourced by template-test.sh to prove that an extension of a template before
# just finish adopts it after just update (#427). It uses the update helpers
# of template-update-test.sh and the finish and signoff helpers.
# shellcheck disable=SC2154 # fct_main sets root in template-test.sh.

# fct_test_adoption carries an extension that works at v0.23.4, before just
# finish, to this checkout, and runs its first build as the nicely-extension
# skill says (#427). Its first build starts at the update commit, its intake
# reads its current spec.md, and just signoff passes. In that build, a fix
# that belongs to sdk/wrap waits under **Limits.** with its issue (#407).
fct_test_adoption() {
	# shellcheck disable=SC2034 # The helpers of template-finish-helpers.sh read finish and records_status.
	local scratch="${1}" src="${1}/source" repo="${1}/ncly-probe" version=v0.99.2 copier file update recovery limit status finish="${1}/ncly-probe" records_status
	shift
	copier=("$@")
	mkdir -p "${src}"
	git -C "${root}" archive v0.23.4 | tar -C "${src}" -xf -
	git -C "${src}" init -q
	fct_revision v0.23.4
	"${copier[@]}" copy --quiet --vcs-ref v0.23.4 --defaults --data domain=probe --data module=example.com/ncly-probe "${src}" "${repo}"
	git -C "${repo}" init -q
	fct_commit 'rendered from v0.23.4'
	# The author filled the forms of v0.23.4, as video-convert did.
	for file in vision.md spec.md qa.md; do
		grep -v '^Fill:' "${repo}/${file}" >"${scratch}/form.md"
		mv "${scratch}/form.md" "${repo}/${file}"
	done
	fct_commit 'a working extension at v0.23.4'

	fct_copy_source "${src}"
	fct_revision "${version}"
	status=0
	fct_update "${version}" || status=$?
	# The template brings its new Fill: lines into the filled spec.md. The author keeps them, for the first build.
	[[ "${status}" == 1 ]] || fct_fail "just update ${version} exits ${status}, want 1 on the conflict of the filled spec.md"
	fct_said '  spec.md'
	recovery="$(sed -n '/^  go -C /s/^  //p' "${scratch}/update.out")"
	awk '/^<<<<<<< /{ before = 1; next } /^=======$/{ before = 0; next } /^>>>>>>> /{ next } !before' "${repo}/spec.md" >"${scratch}/form.md"
	mv "${scratch}/form.md" "${repo}/spec.md"
	git -C "${repo}" add spec.md
	(cd "${repo}" && bash -c "${recovery}") >"${scratch}/update.out" 2>&1 || {
		cat "${scratch}/update.out" >&2
		fct_fail "the continuation of just update ${version} failed"
	}
	fct_commit "update to ${version}"
	update="$(git -C "${repo}" rev-parse HEAD)"

	# Step 1: the intake reads the current spec.md, at the update commit.
	grep -v '^# Fill:' "${repo}/intake.toml" >"${scratch}/form.md"
	mv "${scratch}/form.md" "${repo}/intake.toml"
	printf '\n[[facts]]\nkind = "sourced"\nstatement = "Failures: a missing file is USAGE_INVALID, and a timeout is RUNTIME, whose answer sdk/wrap writes"\nsource = "spec.md of this extension"\nrevision = "%s"\nread = "Failures of ncly probe inspect"\n' "${update}" >>"${repo}/intake.toml"
	(cd "${repo}" && just records) >"${scratch}/records.out" 2>&1 || {
		cat "${scratch}/records.out" >&2
		fct_fail 'just records failed after the update'
	}
	git -C "${repo}" add -A
	git -C "${repo}" commit -qm 'step: intake'
	# Step 2: the spec. The timeout answer names no input, and only sdk/wrap can fix it (#407).
	# The author opened an issue in Nicely, and the user accepted the limit.
	grep -v '^Fill:' "${repo}/spec.md" >"${scratch}/form.md"
	mv "${scratch}/form.md" "${repo}/spec.md"
	# shellcheck disable=SC2016 # Markdown backticks are literal.
	limit='| `timeout-input` | `limit` | A timeout answer does not name the input file, because `sdk/wrap` writes it: https://github.com/pascalandy/nicely/issues/407 | The user, 2026-11-03, https://example.com/pull/2 |'
	fct_edit "${repo}/spec.md" '**Limits.** None.' $'**Limits.**\n\n| ID | State | Limit | Accepted |\n|---|---|---|---|\n'"${limit}"
	git -C "${repo}" commit -qam 'step: spec'
	# Step 3: the first scenario fails, for a missing file that the answer names in plain words.
	printf '# A missing file names the cause in plain words.\n# spec: bad-input\n%s\n%s\n' \
		"$(grep -m 1 '^refuses USAGE_INVALID ncly .* missing.md --json$' "${repo}/testdata/script/inspect.txtar")" \
		"stderr 'missing.md is no file that probe can read[.]'" >"${repo}/testdata/script/first.txtar"
	git -C "${repo}" add -A
	git -C "${repo}" commit -qm 'step: scenario'
	# Step 4: the handler.
	for file in locales/en.toml testdata/script/inspect.txtar; do
		sed "s/ can inspect\([.[]\)/ can read\1/" "${repo}/${file}" >"${scratch}/form.md"
		mv "${scratch}/form.md" "${repo}/${file}"
	done
	git -C "${repo}" commit -qam 'handler'

	# The update commit holds the example intake of the template, so its intake starts no build.
	fct_readiness "$(git -C "${repo}" rev-parse HEAD)" 'Ready with stated limits' ''
	sed "s/^intake = .*/intake = \"${update}\"/" "${repo}/review.toml" >"${scratch}/form.md"
	mv "${scratch}/form.md" "${repo}/review.toml"
	fct_records
	fct_finish 1
	fct_finish_said "review.toml: the intake of this build is not after ${update:0:12}, which added intake.toml"

	fct_signoff_stub "${scratch}"
	fct_signoff_review "${repo}" first '' 'Ready with stated limits'
	if ! fct_signoff "${repo}" "${scratch}"; then
		cat "${scratch}/signoff.out" >&2
		fct_fail 'just signoff failed on the first build after the update'
	fi
	fct_has "${scratch}/signoff.out" "Signed off on $(git -C "${repo}" rev-parse HEAD)"
	printf 'an extension of v0.23.4 adopts just finish through its first build after just update %s\n' "${version}"
}
