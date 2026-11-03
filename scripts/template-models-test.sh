#!/usr/bin/env bash
# Sourced by template-recipes-test.sh to prove just verify-models, and the
# record that just check reads, in a rendered repository. A fake mise stands
# for Java and TLC, so the test needs neither and no network.

# fct_test_models runs just verify-models in the rendered repository $1, with
# a local jar of the pinned digest, then with changed bytes, as the URL of
# TLC v1.8.0 served on 2026-11-02. Then just check reads only the record.
fct_test_models() {
	local repo="${1}" scratch="${2}/models" status=0 pin
	mkdir -p "${scratch}/bin" "${scratch}/tlc" "${repo}/verification/m"
	cat >"${scratch}/bin/mise" <<'EOF'
#!/bin/sh
printf '%s\n' "$*" >>"$FAKE_MISE_LOG"
config=
previous=
for arg in "$@"; do
	[ "$previous" = -config ] && config=$arg
	previous=$arg
done
[ -n "$config" ] || exit 0
name=$(basename "$config" .cfg)
cat "$FAKE_TLC/$name.out"
exit "$(cat "$FAKE_TLC/$name.exit")"
EOF
	chmod +x "${scratch}/bin/mise"
	printf 'Model checking completed. No error has been found.\n9 states generated, 7 distinct states found, 0 states left on queue.\nThe depth of the complete state graph search is 3.\n' >"${scratch}/tlc/Pass.out"
	printf 'Error: Invariant Safe is violated.\n4 states generated, 4 distinct states found, 2 states left on queue.\nThe depth of the complete state graph search is 2.\n' >"${scratch}/tlc/Bad.out"
	printf '0\n' >"${scratch}/tlc/Pass.exit"
	printf '12\n' >"${scratch}/tlc/Bad.exit"
	printf 'the pinned jar\n' >"${scratch}/pinned.jar"
	printf 'the changed jar\n' >"${scratch}/changed.jar"
	if command -v sha256sum >/dev/null; then
		pin="$(sha256sum "${scratch}/pinned.jar")"
	else
		pin="$(shasum -a 256 "${scratch}/pinned.jar")"
	fi
	pin="${pin%% *}"
	printf -- '---- MODULE Model ----\n====\n' >"${repo}/verification/m/Model.tla"
	printf 'SPECIFICATION Spec\n' >"${repo}/verification/m/Pass.cfg"
	printf 'SPECIFICATION Spec\n' >"${repo}/verification/m/Bad.cfg"
	printf '{"java": "java@temurin-21.0.13+9.0.LTS", "tlc_url": "https://example.com/tla2tools.jar", "tlc_sha256": "%s", "checks": [
  {"name": "Pass", "model": "verification/m/Model.tla", "config": "verification/m/Pass.cfg"},
  {"name": "Bad", "model": "verification/m/Model.tla", "config": "verification/m/Bad.cfg", "violation": "Safe"}]}\n' "${pin}" >"${repo}/verification/tlc.json"
	# The jar is relative to the folder of the caller, as for just acceptance.
	if ! (cd "${scratch}" && PATH="${scratch}/bin:${PATH}" FAKE_MISE_LOG="${scratch}/mise.log" FAKE_TLC="${scratch}/tlc" just --justfile "${repo}/justfile" verify-models pinned.jar) >"${scratch}/run.out" 2>&1; then
		cat "${scratch}/run.out" >&2
		fct_fail 'just verify-models fails with the pinned jar'
	fi
	grep -qF 'verify-models: wrote verification/tlc-record.json.' "${scratch}/run.out" || fct_fail "just verify-models wrote no record: $(cat "${scratch}/run.out")"
	cp "${repo}/verification/tlc-record.json" "${scratch}/record.json"
	rm "${scratch}/mise.log"
	(cd "${repo}" && PATH="${scratch}/bin:${PATH}" FAKE_MISE_LOG="${scratch}/mise.log" FAKE_TLC="${scratch}/tlc" just verify-models "${scratch}/changed.jar") >"${scratch}/run.out" 2>&1 || status=$?
	[[ "${status}" != 0 ]] || fct_fail 'just verify-models accepts a jar with other bytes'
	grep -qF 'Changed bytes never prove a model' "${scratch}/run.out" || fct_fail "just verify-models fails without the digest: $(cat "${scratch}/run.out")"
	[[ ! -e "${scratch}/mise.log" ]] || fct_fail 'just verify-models starts Java before the digest check'
	cmp -s "${repo}/verification/tlc-record.json" "${scratch}/record.json" || fct_fail 'a refused jar changes the record'
	# just check needs no Java: the PATH of the caller holds no mise here.
	printf -- '---- MODULE Model ----\nX == 1\n====\n' >"${repo}/verification/m/Model.tla"
	status=0
	(cd "${repo}" && just check) >"${scratch}/check.out" 2>&1 || status=$?
	[[ "${status}" != 0 ]] || fct_fail 'just check accepts the record of an earlier model'
	grep -qF 'verification/tlc-record.json: the record is of other inputs: verification/m/Model.tla changed.' "${scratch}/check.out" ||
		fct_fail "just check fails without the stale record: $(cat "${scratch}/check.out")"
	# With a fresh record, just check goes on to gofmt, which this file fails.
	printf -- '---- MODULE Model ----\n====\n' >"${repo}/verification/m/Model.tla"
	mkdir "${repo}/models_probe"
	printf 'package probe\n\nvar  x = 1\n' >"${repo}/models_probe/probe.go"
	status=0
	(cd "${repo}" && just check) >"${scratch}/check.out" 2>&1 || status=$?
	if [[ "${status}" == 0 ]] || ! grep -qF 'run gofmt -w .' "${scratch}/check.out" || grep -qF 'tlc-record.json' "${scratch}/check.out"; then
		fct_fail "just check does not accept a fresh record: $(cat "${scratch}/check.out")"
	fi
	rm -r "${repo}/models_probe" "${repo}/verification"
}
