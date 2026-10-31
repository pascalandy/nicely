#!/usr/bin/env bash
# Shared calls and receipt fixtures for the local release proof.

fct_release_attempt() {
	local repo="${1}" scratch="${2}" tag="${3}" status=0
	shift 3
	(cd "${scratch}" && PATH="${scratch}/release-bin:${PATH}" XDG_CACHE_HOME="${scratch}/release-cache" just --justfile "${repo}/justfile" release "${tag}" "$@") >"${scratch}/release.out" 2>&1 || status=$?
	printf 'release proof: just release %s, exit %s\n' "${tag}" "${status}"
	cat "${scratch}/release.out"
	return "${status}"
}

# Receipts are fixture facts. The release under test must read T9's format.
fct_release_receipt() {
	local repo="${1}" scratch="${2}" sha
	sha="$(git -C "${repo}" rev-parse HEAD)"
	mkdir -p "${scratch}/release-cache/nicely/signoff/test-${sha}"
	PATH="/usr/bin:${PATH}" jq -n --arg sha "${sha}" '{sha:$sha,os:"test",tools:{git:"test"},checks:[{command:["just","finish"],exit:0},{command:["just","check"],exit:0}],signoff:{command:["gh","signoff","--commit",$sha],exit:0}}' >"${scratch}/release-cache/nicely/signoff/test-${sha}/receipt.json"
}

# fct_release_refused_receipts requires that each unsuccessful form of the
# receipt $3 of the repository $1 refuses the release, then restores it.
fct_release_refused_receipts() {
	local repo="${1}" scratch="${2}" receipt="${3}" previous status
	cp "${receipt}" "${scratch}/valid-signoff.json"
	for previous in '.checks[0].exit = 1' '.signoff.exit = 1' '.sha = "previous"' 'del(.checks[0].exit)' '.signoff.exit = null'; do
		PATH="/usr/bin:${PATH}" jq "${previous}" "${scratch}/valid-signoff.json" >"${receipt}"
		status=0
		fct_release_attempt "${repo}" "${scratch}" v0.1.0 || status=$?
		if [[ "${status}" == 0 ]] || ! grep -qF 'just signoff' "${scratch}/release.out"; then
			fct_fail 'release accepts an unsuccessful receipt'
		fi
	done
	# A receipt without just finish, as before #415 or written by hand, cannot authorize a release (#416).
	PATH="/usr/bin:${PATH}" jq '.checks |= map(select(.command[-1] != "finish"))' "${scratch}/valid-signoff.json" >"${receipt}"
	status=0
	fct_release_attempt "${repo}" "${scratch}" v0.1.0 || status=$?
	if [[ "${status}" == 0 ]] || ! grep -qF 'holds no successful just finish' "${scratch}/release.out"; then
		fct_fail 'release accepts a receipt without just finish'
	fi
	cp "${scratch}/valid-signoff.json" "${receipt}"
}
