# shellcheck shell=bash
# shellcheck disable=SC2154 # verify.sh sets script_dir and test_service before prepare.sh sources this file
# prepare-auth.sh holds the fixture of the auth feature, which prepare.sh sources:
#
#   verify.sh prepare <run> auth [--mac-keychain]
#                                       install the fixture that declares the test service

# prepare_auth installs the extension verifykeys, whose one command declares the test service,
# so that ncly auth reads that service's entry in the real keychain. Run through ncly, it
# answers the SHA-256 of the key that it receives, never the key, or the problem that
# NCLY_MISSING_KEYS names for the service, such as KEY_NOT_GRANTED before a human grants the
# key to it.
prepare_auth() {
	local run="$1" extensions
	if [[ "$(uname -s)" == Darwin && "${2:-}" != --mac-keychain ]]; then
		die "on macOS, the keychain may ask for a password, so only a human runs the auth feature there, with --mac-keychain"
	fi
	[[ -d "$run/scratch" ]] || die "$run was cleaned up: start a new run"
	extensions="$run/scratch/home/.local/share/nicely/extensions"
	cat >"$extensions/ncly-verifykeys" <<'FIXTURE'
#!/bin/sh
# verifykeys answers the SHA-256 of the key of verifynicely that ncly passes it, never the key,
# or the problem that NCLY_MISSING_KEYS names for verifynicely.
fail() {
	printf '{"ok":false,"contract_version":1,"errors":[{"code":"%s","message":"%s","hint":"%s"%s}]}\n' "$1" "$2" "$3" "$4" >&2
	exit 78
}
missing="$(printf '%s\n' "${NCLY_MISSING_KEYS:-}" | tr ',' '\n' | sed -n 's/^verifynicely=//p')"
case "$missing" in
KEY_NOT_GRANTED)
	fail KEY_NOT_GRANTED "The key of verifynicely is not granted to verifykeys." "ncly auth grant verifynicely verifykeys" ',"next":["ncly","auth","grant","verifynicely","verifykeys"]'
	;;
KEYRING_UNAVAILABLE)
	fail KEYRING_UNAVAILABLE "The keychain does not answer, and VERIFYNICELY_API_KEY is empty." "Export VERIFYNICELY_API_KEY, or unlock the keychain." ''
	;;
esac
if [ -n "$missing" ] || [ -z "${VERIFYNICELY_API_KEY:-}" ]; then
	fail AUTH_MISSING "No key for verifynicely in VERIFYNICELY_API_KEY or the keychain." "ncly auth login verifynicely" ',"next":["ncly","auth","login","verifynicely"]'
fi
if command -v sha256sum >/dev/null 2>&1; then
	digest="$(printf '%s' "$VERIFYNICELY_API_KEY" | sha256sum | cut -d' ' -f1)"
else
	digest="$(printf '%s' "$VERIFYNICELY_API_KEY" | shasum -a 256 | cut -d' ' -f1)"
fi
printf '{"ok":true,"contract_version":1,"sha256":"%s"}\n' "$digest"
FIXTURE
	chmod +x "$extensions/ncly-verifykeys"
	printf '%s\n' 'format_version = 1' 'contract_version = 1' 'domain = "verifykeys"' '' \
		'[[commands]]' 'path = "verifykeys"' 'summary = "verifykeys.summary"' '' \
		'[[commands]]' 'path = "verifykeys digest"' 'summary = "verifykeys.digest.summary"' \
		'examples = ["ncly verifykeys digest --json", "ncly auth grant verifynicely verifykeys"]' \
		"requires = { keys = [\"$test_service\"] }" >"$extensions/ncly-verifykeys.toml"
	mkdir -p "$extensions/ncly-verifykeys.locales"
	printf '%s\n' '["verifykeys.summary"]' \
		'description = "Help page: summary of the verify-nicely fixture, which declares a test service."' \
		'other = "Declare the test service of verify-nicely"' '' \
		'["verifykeys.digest.summary"]' \
		'description = "Help page: summary of the command of the verify-nicely fixture that answers the SHA-256 of the key of the test service, never the key."' \
		'other = "Answer the SHA-256 of the key of the test service"' >"$extensions/ncly-verifykeys.locales/en.toml"
	: >"$run/auth-prepared"
	log "verifykeys declares the service $test_service, whose variable is VERIFYNICELY_API_KEY, and answers the SHA-256 of its key"
}
