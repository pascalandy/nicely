# shellcheck shell=bash
# shellcheck disable=SC2154 # verify.sh sets script_dir and test_service before it sources this file
# prepare.sh holds the prepare step of verify.sh, which sources it: the fixture of each
# feature that needs one beside hello.
#
#   verify.sh prepare <run> auth [--mac-keychain]
#                                       install the fixture that declares the test service
#   verify.sh prepare <run> agent       run just install in the scratch home, with a stub codex
#   verify.sh prepare <run> transcript  run just install in the scratch home, with the real uv and
#                                       Python first in PATH

# prepare installs the fixture of a feature beside hello. The auth fixture is
# the extension verifykeys, whose one command declares the test service, so
# that ncly auth reads that service's entry in the real keychain. Run through
# ncly, it answers the SHA-256 of the key that it receives, never the key, or
# the problem that NCLY_MISSING_KEYS names for the service, such as
# KEY_NOT_GRANTED before a human grants the key to it.
prepare() {
	local run extensions
	run="$(run_folder "${1:-}")"
	case "${2:-}" in
	agent)
		prepare_agent "$run"
		return
		;;
	transcript)
		prepare_transcript "$run"
		return
		;;
	template)
		# shellcheck source=/dev/null
		. "$script_dir/prepare-template.sh"
		prepare_template "$run"
		return
		;;
	auth) ;;
	*) die "only the auth, agent, transcript, and template features have a fixture" ;;
	esac
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

# install_checkout installs ncly and the first-party extensions with just install, in the
# scratch home, as a user installs them from a checkout.
install_checkout() {
	local run="$1" repo home gocache gomodcache gopath goenv
	[[ -d "$run/scratch" ]] || die "$run was cleaned up: start a new run"
	repo="$(git -C "$script_dir" rev-parse --show-toplevel)" || die "run it from a checkout of nicely"
	command -v just >/dev/null || die "just is missing: install it, such as with brew install just"
	home="$run/scratch/home"
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
}

# wrapper_path prints the PATH of the run's wrapper.
wrapper_path() {
	sed -n "s/^PATH='\(.*\)'\$/\1/p" "$1/scratch/bin/ncly"
}

# prepare_agent installs ncly and the first-party extensions with install_checkout. It puts a
# stub codex first in the wrapper's PATH, so no real Codex ever starts, and points the wrapper
# at the installed ncly. The stub appends each call to evidence/codex-calls.txt, which outlives
# the cleanup, and answers --version with codex-cli and $CODEX_STUB_VERSION, 0.163.0 by
# default. It appends any other call, which starts a task, to evidence/codex-task.txt, and the
# task that it reads on stdin to evidence/codex-input.txt, waits $CODEX_STUB_WAIT seconds when
# it is set, then writes "OK from the stub" in the file of --output-last-message.
prepare_agent() {
	local run="$1" home="$1/scratch/home" stub="$1/scratch/stub" path
	install_checkout "$run"
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
	path="$(wrapper_path "$run")"
	case "$stub" in *"'"*) die "the stub folder $stub holds a quote, which the wrapper cannot hold" ;; esac
	write_wrapper "$run/scratch/bin/ncly" "$home/.local/bin/ncly" "$home" "$stub:$path"
	: >"$run/agent-prepared"
	log "the wrapper now runs $home/.local/bin/ncly, with the stub codex of $stub first in PATH"
}

# prepare_transcript installs ncly and the first-party extensions with install_checkout. It
# links the real uv and Python 3.12 or later into scratch/tools, first in the wrapper's PATH, and
# points the wrapper at the installed ncly. The engine then runs through them, with the empty uv
# cache of the scratch home, which it fills from the network. The wrapper also sets the fake key
# of deepgram, whatever the caller holds. Core reads that variable before the keychain, so no
# call of the run, a tmux one included, ever reaches a real key in the keychain.
prepare_transcript() {
	local run="$1" home="$1/scratch/home" tools="$1/scratch/tools" wrapper="$1/scratch/bin/ncly" path
	install_checkout "$run"
	mkdir -p "$tools"
	link_host_program uv "$home" "$tools" --version
	link_host_program python3 "$home" "$tools" -c 'import sys; sys.exit(sys.version_info < (3, 12))'
	path="$(wrapper_path "$run")"
	write_wrapper "$wrapper" "$home/.local/bin/ncly" "$home" "$tools:$path"
	# The last line of the wrapper is its exec, so the key goes just before it.
	if ! {
		sed '$d' "$wrapper"
		printf '%s\n' "DEEPGRAM_API_KEY='$fake_deepgram_key'" 'export DEEPGRAM_API_KEY'
		tail -n 1 "$wrapper"
	} >"$wrapper.new" || ! mv "$wrapper.new" "$wrapper" || ! chmod +x "$wrapper"; then
		die "could not set the fake key of deepgram in $wrapper"
	fi
	grep -q "^DEEPGRAM_API_KEY='$fake_deepgram_key'\$" "$wrapper" || die "$wrapper holds no fake key of deepgram"
	: >"$run/transcript-prepared"
	log "the wrapper now runs $home/.local/bin/ncly, with the uv and python3 of $tools first in PATH"
	log "every ncly call of this run gets DEEPGRAM_API_KEY=$fake_deepgram_key, a fake key that Deepgram refuses"
}

# fake_deepgram_key is the key of deepgram in each transcript run: no human's key, which
# Deepgram refuses with 401, so nothing bills.
fake_deepgram_key="fake-deepgram-key-for-verify"

# link_host_program links into the folder $3 the first program of PATH named $1 that answers the
# rest of the arguments with exit 0 under the home $2, as in the run. A shim that needs the real
# home, such as one of mise, never counts.
link_host_program() {
	local name="$1" home="$2" tools="$3" dir
	shift 3
	local IFS=:
	for dir in $PATH; do
		[[ -f "$dir/$name" && -x "$dir/$name" ]] || continue
		if env -i HOME="$home" PATH="$dir" "$dir/$name" "$@" >/dev/null 2>&1; then
			ln -sf "$dir/$name" "$tools/$name"
			log "$name: $dir/$name"
			return
		fi
	done
	die "no $name in PATH answers under the scratch home: install it, such as with brew install $name"
}
