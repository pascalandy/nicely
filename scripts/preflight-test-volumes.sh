# shellcheck shell=bash disable=SC2034,SC2154
# The volume cases of preflight-test.sh, which sources this file. fct_main of the test sets
# root, and its helpers write the stubs and run preflight.sh. See #366.

# fct_macos writes the stubs of a macOS host. GOCACHE and the fixture folders sit on the
# external volume /Volumes/ext, GOMODCACHE on /Volumes/net, which diskutil cannot read, and the
# checkout on the internal volume /. stat reports each file locked with uchg.
fct_macos() {
	fct_stub stubs uname 'echo Darwin'
	fct_stub stubs stat 'echo uchg'
	fct_stub stubs go "case \"\$*\" in
\"env GOCACHE\") echo ${root}/ext/go-build ;;
\"env GOMODCACHE\") echo ${root}/net/go-mod ;;
esac"
	fct_stub stubs df "case \"\$2\" in
${root}/ext/*|${root}/fixtures/*) mount=/Volumes/ext ;;
${root}/net/*) mount=/Volumes/net ;;
*) mount=/ ;;
esac
printf 'Filesystem 512-blocks Used Available Capacity Mounted on\n/dev/disk9s1 100 50 50 50%% %s\n' \"\$mount\""
	fct_stub stubs diskutil "case \"\$*\" in
\"info -plist /\") printf '<key>Internal</key>\n<true/>\n' ;;
\"info -plist /Volumes/ext\") printf '<key>Internal</key>\n<false/>\n' ;;
*) exit 1 ;;
esac"
}

fct_volume_cases() {
	test_case="on macOS, files on internal volumes pass"
	fct_present
	fct_macos
	fct_stub stubs df "printf 'Filesystem 512-blocks Used Available Capacity Mounted on\n/dev/disk3s5 100 50 50 50%% /\n'"
	fct_run 0
	fct_said "ok    on internal volumes: checkout, GOCACHE, GOMODCACHE"
	fct_said "preflight: 0 missing, 0 warnings"

	test_case="on macOS, a Go cache or a fixture on an external volume warns"
	fct_present
	fct_macos
	fct_fixture far
	fct_run 0 "${root}/fixtures/far"
	fct_said "ok    fixture folder ${root}/fixtures/far: checksums OK, originals locked"
	fct_said "warn  on an external volume: GOCACHE at /Volumes/ext, GOMODCACHE at /Volumes/net, fixture ${root}/fixtures/far at /Volumes/ext: through cua do shell"
	fct_hint "warn  on an external volume"
	fct_unsaid "checkout at"
	fct_said "preflight: 0 missing, 1 warnings"

	test_case="on Linux, preflight names no volume"
	fct_present
	fct_macos
	fct_stub stubs uname 'echo Linux'
	fct_run 0
	fct_unsaid "volume"
	fct_said "preflight: 0 missing, 0 warnings"
}
