# shellcheck shell=bash
# shellcheck disable=SC2154 # verify.sh sets script_dir and test_service before it sources this file
# prepare.sh holds the prepare step of verify.sh, which sources it: the fixture of each
# feature that needs one beside hello.
#
#   verify.sh prepare <run> auth [--mac-keychain]
#                                       install the fixture that declares the test service
#   verify.sh prepare <run> agent       run just install in the scratch home, with a stub codex

# prepare installs the fixture of a feature beside hello. The auth fixture is
# the extension verifykeys, whose one command declares the test service, so
# that ncly auth reads that service's entry in the real keychain. Run through
# ncly, it answers the SHA-256 of the key that it receives, never the key, or
# the problem that NCLY_MISSING_KEYS names for the service, such as
# KEY_NOT_GRANTED before a human grants the key to it.
prepare() {
	local run extensions
	run="$(run_folder "${1:-}")"
	if [[ "${2:-}" == agent ]]; then
		prepare_agent "$run"
		return
	fi
	[[ "${2:-}" == auth ]] || die "only the auth and agent features have a fixture"
	if [[ "$(uname -s)" == Darwin && "${3:-}" != --mac-keychain ]]; then
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

# prepare_agent installs ncly and the first-party extensions with just install, in the scratch
# home, as a user installs them from a checkout. It puts a stub codex first in the wrapper's
# PATH, so no real Codex ever starts, and points the wrapper at the installed ncly. The stub
# appends each call to evidence/codex-calls.txt, which outlives the cleanup, and answers
# --version with codex-cli and $CODEX_STUB_VERSION, 0.163.0 by default. It appends any other
# call, which starts a task, to evidence/codex-task.txt, and the task that it reads on stdin to
# evidence/codex-input.txt, waits $CODEX_STUB_WAIT seconds when it is set, then writes "OK from
# the stub" in the file of --output-last-message.
prepare_agent() {
	local run="$1" repo home stub path gocache gomodcache gopath goenv
	[[ -d "$run/scratch" ]] || die "$run was cleaned up: start a new run"
	repo="$(git -C "$script_dir" rev-parse --show-toplevel)" || die "run it from a checkout of nicely"
	command -v just >/dev/null || die "just is missing: install it, such as with brew install just"
	home="$run/scratch/home"
	stub="$run/scratch/stub"
	mkdir -p "$stub"
	cat >"$stub/codex" <<'STUB'
#!/bin/sh
# Written by verify-nicely: a counted stub of codex, which never sends a request.
evidence="$(dirname "$0")/../../evidence"
echo "codex $*" >>"$evidence/codex-calls.txt"
if [ "$1" = --version ]; then
	echo "codex-cli ${CODEX_STUB_VERSION:-0.163.0}"
	exit 0
fi
echo "codex $*" >>"$evidence/codex-task.txt"
cat >>"$evidence/codex-input.txt"
echo >>"$evidence/codex-input.txt"
out=
while [ $# -gt 0 ]; do
	if [ "$1" = --output-last-message ]; then out="$2"; fi
	shift
done
if [ -n "$CODEX_STUB_WAIT" ]; then sleep "$CODEX_STUB_WAIT"; fi
echo '{"type":"turn.completed"}'
printf 'OK from the stub' >"$out"
STUB
	chmod +x "$stub/codex"
	# The build keeps the caller's Go caches and settings, which live under the real home.
	gocache="$(go env GOCACHE)"
	gomodcache="$(go env GOMODCACHE)"
	gopath="$(go env GOPATH)"
	goenv="$(go env GOENV)"
	log "running just install with HOME=$home"
	(cd "$repo" && env HOME="$home" XDG_CONFIG_HOME="$home/.config" XDG_DATA_HOME="$home/.local/share" \
		XDG_STATE_HOME="$home/.local/state" XDG_CACHE_HOME="$home/.cache" \
		GOCACHE="$gocache" GOMODCACHE="$gomodcache" GOPATH="$gopath" GOENV="$goenv" just install) ||
		die "just install failed"
	[[ -x "$home/.local/bin/ncly" ]] || die "just install wrote no $home/.local/bin/ncly"
	path="$(sed -n "s/^PATH='\(.*\)'\$/\1/p" "$run/scratch/bin/ncly")"
	case "$stub" in *"'"*) die "the stub folder $stub holds a quote, which the wrapper cannot hold" ;; esac
	write_wrapper "$run/scratch/bin/ncly" "$home/.local/bin/ncly" "$home" "$stub:$path"
	: >"$run/agent-prepared"
	log "the wrapper now runs $home/.local/bin/ncly, with the stub codex of $stub first in PATH"
}
