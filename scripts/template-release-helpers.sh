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
	PATH="/usr/bin:${PATH}" jq -n --arg sha "${sha}" '{sha:$sha,os:"test",tools:{git:"test"},checks:[{command:["just","check"],exit:0}],signoff:{command:["gh","signoff","--commit",$sha],exit:0}}' >"${scratch}/release-cache/nicely/signoff/test-${sha}/receipt.json"
}
