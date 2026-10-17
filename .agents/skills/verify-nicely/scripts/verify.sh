#!/usr/bin/env bash
# verify.sh drives the real ncly in a scratch folder, as SKILL.md describes.
# It runs on macOS and Linux: bash 3.2, and only the flags that BSD and GNU
# tools share.
#
#   verify.sh start                     build ncly, prepare a run, print its folder
#   verify.sh prepare <run> auth [--mac-keychain]
#                                       install the fixture that declares the test service
#   verify.sh prepare <run> agent       run just install in the scratch home, with a stub codex
#   verify.sh step <run> <feature> <ncly args...>
#                                       run ncly in the run's scratch folder and
#                                       record the call in the feature's transcript
#   verify.sh tmux <run> <feature> [--answer <prompt> <keys>]... <ncly args...>
#                                       run ncly on a terminal inside tmux, whose
#                                       server does not answer, type each answer once
#                                       its prompt shows, and record it the same way
#   verify.sh quote <run> [<feature>]   print transcripts with placeholders for home paths
#   verify.sh cleanup <run>             remove the scratch folder, keep the evidence
set -Eeuo pipefail

script_dir="$(cd "$(dirname "$0")" && pwd -P)"
marker=".verify-nicely-run"

die() {
	echo "verify-nicely: $*" >&2
	exit 1
}

log() { echo "verify-nicely: $*" >&2; }

# usage prints the commands of the header: on stdout for help, on stderr with exit 2 otherwise.
usage() {
	if [[ "${1:-}" == help ]]; then
		sed -n '6,18p' "$0" | sed 's/^# \{0,1\}//'
		exit 0
	fi
	sed -n '6,18p' "$0" | sed 's/^# \{0,1\}//' >&2
	exit 2
}

# run_folder checks that $1 is a run that start created, and prints its absolute path.
run_folder() {
	[[ -n "${1:-}" ]] || usage
	[[ -f "$1/$marker" ]] || die "$1 is not a run that verify.sh start created"
	(cd "$1" && pwd -P)
}

# check_feature accepts the name of a feature file: lowercase letters, digits, and hyphens.
check_feature() {
	case "$1" in
	"" | *[!a-z0-9-]* | [!a-z]*) die "a feature is named in lowercase letters, digits, and hyphens, such as hello" ;;
	esac
}

# clean_path prints PATH without the folders that hold an ncly-<domain> file, so
# discovery finds only the extensions that the run installs.
clean_path() {
	local kept="" dir file found
	local IFS=:
	for dir in $PATH; do
		# The wrapper quotes PATH in single quotes, so a folder that holds one stays out.
		[[ "$dir" == /* && "$dir" != *"'"* ]] || continue
		found=""
		for file in "$dir"/ncly-*; do
			[[ -e "$file" ]] && found=1 && break
		done
		if [[ -n "$found" ]]; then
			log "left $dir out of PATH, because it holds an ncly extension"
			continue
		fi
		kept="${kept:+$kept:}$dir"
	done
	printf '%s\n' "$kept"
}

# install_hello builds the Go fixture hello and installs it in the data folder,
# with its manifest, catalogs, and skill beside it, as Discovery in contract.md lays out.
install_hello() {
	local repo="$1" extensions="$2" fixture="$1/testdata/extensions/hello"
	mkdir -p "$extensions"
	(cd "$repo" && go build -o "$extensions/ncly-hello" ./testdata/extensions/hello/cmd/ncly-hello)
	cp "$fixture/manifest.toml" "$extensions/ncly-hello.toml"
	cp -R "$fixture/locales" "$extensions/ncly-hello.locales"
	cp -R "$fixture/skill" "$extensions/ncly-hello.skill"
}

# write_wrapper writes the scratch folder's ncly, which sets HOME, the XDG
# folders, and PATH, unsets NCLY_CONFIG, which would point the config
# elsewhere, then runs the built ncly.
write_wrapper() {
	local wrapper="$1" binary="$2" home="$3" path="$4"
	cat >"$wrapper" <<EOF
#!/bin/sh
# Written by verify-nicely: runs the ncly of this run in its scratch folder only.
HOME='$home'
XDG_CONFIG_HOME='$home/.config'
XDG_DATA_HOME='$home/.local/share'
XDG_STATE_HOME='$home/.local/state'
XDG_CACHE_HOME='$home/.cache'
PATH='$path'
export HOME XDG_CONFIG_HOME XDG_DATA_HOME XDG_STATE_HOME XDG_CACHE_HOME PATH
unset NCLY_CONFIG
exec '$binary' "\$@"
EOF
	chmod +x "$wrapper"
}

start() {
	local repo root run stamp home path
	command -v go >/dev/null || die "go is missing: install it from https://go.dev/dl/"
	repo="$(git -C "$script_dir" rev-parse --show-toplevel)" || die "run it from a checkout of nicely"
	# Not /tmp: Linux mounts it as tmpfs and macOS purges it, and the evidence must outlive the run.
	root="${XDG_CACHE_HOME:-$HOME/.cache}/nicely/verify"
	stamp="$(date +%Y%m%d-%H%M%S)-$$"
	run="$root/$stamp"
	home="$run/scratch/home"
	case "$run" in *"'"*) die "the run folder $run holds a quote, which the wrapper cannot hold" ;; esac
	mkdir -p "$run/build" "$run/evidence" "$run/scratch/bin" "$home/.config" "$home/.local/share" "$home/.local/state" "$home/.cache"
	: >"$run/$marker"
	log "building ncly from $repo"
	(cd "$repo" && go build -o "$run/build/ncly" ./cmd/ncly)
	install_hello "$repo" "$home/.local/share/nicely/extensions"
	path="$(clean_path)"
	write_wrapper "$run/scratch/bin/ncly" "$run/build/ncly" "$home" "$path"
	git -C "$repo" rev-parse --short=12 HEAD >"$run/evidence/commit.txt"
	if [[ -n "$(git -C "$repo" status --porcelain)" ]]; then
		echo "dirty" >>"$run/evidence/commit.txt"
	fi
	log "ncly: $run/scratch/bin/ncly"
	log "evidence: $run/evidence"
	printf '%s\n' "$run"
}

# test_service is the service of the auth feature, whose key is a fake one.
test_service="verifynicely"

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

step() {
	local run feature transcript out err code
	run="$(run_folder "${1:-}")"
	check_feature "${2:-}"
	feature="$2"
	shift 2
	[[ -d "$run/scratch" ]] || die "$run was cleaned up: start a new run"
	transcript="$run/evidence/$feature.txt"
	out="$run/scratch/step.out"
	err="$run/scratch/step.err"
	code=0
	"$run/scratch/bin/ncly" "$@" >"$out" 2>"$err" || code=$?
	# A caller may run step where errexit is off, as cleanup does, and the
	# condition turns it off in the group too, so each read fails the group
	# itself. A call that the transcript misses stops here, and its outputs
	# stay in scratch.
	if ! {
		printf '$ ncly'
		printf ' %q' "$@"
		printf '\n--- stdout\n'
		cat "$out" || exit 1
		printf -- '--- stderr\n'
		cat "$err" || exit 1
		printf -- '--- exit %s\n\n' "$code"
	} | tee -a "$transcript"; then
		die "could not record the call in $transcript: its outputs stay in $out and $err"
	fi
	rm -f "$out" "$err"
	return "$code"
}

# The private tmux servers of tmux_step, which stop_servers resumes and kills.
pane_server="" stuck_server="" pane_socket="" stuck_socket="" stuck_pid=""

# stop_servers resumes the held server, kills both, and removes their socket
# files, which tmux leaves behind. It runs on every exit of tmux_step.
stop_servers() {
	[[ -z "$stuck_pid" ]] || kill -CONT "$stuck_pid" 2>/dev/null || true
	[[ -z "$pane_server" ]] || tmux -L "$pane_server" kill-server 2>/dev/null || true
	[[ -z "$stuck_server" ]] || tmux -L "$stuck_server" kill-server 2>/dev/null || true
	[[ -z "$pane_socket" ]] || rm -f "$pane_socket"
	[[ -z "$stuck_socket" ]] || rm -f "$stuck_socket"
}

# tmux_step runs ncly in a pane of a private tmux server, so that its stdin and
# stdout are a terminal, with TMUX naming a second private server that SIGSTOP
# holds, so that a call to tmux would wait. A program that asks tmux about
# colors as it loads, as Lip Gloss v2.0.6 does, then never answers. It records
# what the pane shows and the exit code, or that no exit came within the limit,
# and exits 1 then. Both servers live under their own -L socket names, and
# stop_servers resumes and kills them on exit, Ctrl-C and SIGTERM included, so
# the user's tmux sessions stay untouched. Each --answer waits until the pane
# shows its prompt, types its keys, records the pane, then presses Enter, or
# presses Ctrl-C when the keys are ^C. The transcript never holds the typed
# keys, only what the pane shows, so a typed key that never shows stays out.
tmux_step() {
	local run feature transcript work code="" limit=5 start i missed="" typed
	local -a prompts=() keys=()
	run="$(run_folder "${1:-}")"
	check_feature "${2:-}"
	feature="$2"
	shift 2
	while [[ "${1:-}" == --answer ]]; do
		[[ $# -ge 3 ]] || die "--answer needs a prompt and the keys to type"
		prompts[${#prompts[@]}]="$2"
		keys[${#keys[@]}]="$3"
		shift 3
	done
	# Each answer gets the time of a call of its own.
	limit=$((limit * (1 + ${#prompts[@]})))
	[[ -d "$run/scratch" ]] || die "$run was cleaned up: start a new run"
	command -v tmux >/dev/null || die "tmux is missing: install it, such as with brew install tmux"
	# The wrapper's PATH leaves out each folder that holds an ncly extension.
	# When it leaves tmux out too, ncly cannot reach tmux, and any build would
	# pass, so the step stops.
	(PATH="$(sed -n "s/^PATH='\(.*\)'\$/\1/p" "$run/scratch/bin/ncly")" && command -v tmux >/dev/null) ||
		die "tmux lies in a folder that the wrapper leaves out of PATH, because it holds an ncly extension: ncly could not reach tmux"
	transcript="$run/evidence/$feature.txt"
	work="$run/scratch/tmux"
	# The caller's own tmux stays out of both servers.
	unset TMUX
	rm -rf "$work"
	mkdir -p "$work"
	trap stop_servers EXIT
	trap 'exit 130' INT
	trap 'exit 143' TERM
	pane_server="nicely-verify-${run##*/}"
	stuck_server="$pane_server-stuck"
	tmux -L "$stuck_server" -f /dev/null new-session -d -s stuck
	stuck_socket="$(tmux -L "$stuck_server" display-message -p '#{socket_path}')"
	stuck_pid="$(tmux -L "$stuck_server" display-message -p '#{pid}')"
	# The pane waits for the file go, runs ncly, writes its exit code, then
	# stays open until kill-server, so the capture holds only what ncly showed.
	# ncly gets only TERM, which tmux sets, and TMUX, beside what the wrapper
	# sets, because a variable of the caller, such as the NO_COLOR=1 that agents
	# set or WT_SESSION, can skip the detection that asks tmux. The variables
	# that reach the session bus of the user's keychain on Linux, which the
	# pane gets from the caller, detect nothing, so they pass too.
	# shellcheck disable=SC2016 # the pane's sh expands these
	tmux -L "$pane_server" -f /dev/null new-session -d -s verify -x 200 -y 50 /bin/sh -c '
		while [ ! -e "$1" ]; do sleep 0.1; done
		code=$2 tmux=$3
		shift 3
		env -i TERM="$TERM" TMUX="$tmux" \
			${DBUS_SESSION_BUS_ADDRESS:+"DBUS_SESSION_BUS_ADDRESS=$DBUS_SESSION_BUS_ADDRESS"} \
			${XDG_RUNTIME_DIR:+"XDG_RUNTIME_DIR=$XDG_RUNTIME_DIR"} "$@"
		echo "$?" >"$code"
		exec sleep 3600' sh "$work/go" "$work/code" "$stuck_socket,$stuck_pid,0" "$run/scratch/bin/ncly" "$@"
	pane_socket="$(tmux -L "$pane_server" display-message -p '#{socket_path}')"
	# A form that erases itself from the top of the pane clears the whole
	# screen, which tmux 3.3 and later would scroll into the history that the
	# capture reads, as if the form stayed. An older tmux lacks the option.
	tmux -L "$pane_server" set-option -g scroll-on-clear off 2>/dev/null || true
	kill -STOP "$stuck_pid"
	start=$SECONDS
	: >"$work/go"
	: >"$work/answers"
	for ((i = 0; i < ${#prompts[@]}; i++)); do
		until tmux -L "$pane_server" capture-pane -p -J -t verify | grep -F -q -- "${prompts[$i]}"; do
			if [[ -s "$work/code" ]] || ((SECONDS - start > limit)); then
				missed="${prompts[$i]}"
				break 2
			fi
			sleep 0.1
		done
		if [[ "${keys[$i]}" == '^C' ]]; then
			tmux -L "$pane_server" send-keys -t verify C-c
			printf -- '--- terminal, once it showed %s, then Ctrl-C\n' "${prompts[$i]}" >>"$work/answers"
			continue
		fi
		tmux -L "$pane_server" capture-pane -p -J -t verify >"$work/before"
		tmux -L "$pane_server" send-keys -t verify -l -- "${keys[$i]}"
		# The pane shows the typing once ncly draws it, within the limit.
		while typed="$(tmux -L "$pane_server" capture-pane -p -J -t verify)" &&
			[[ "$typed" == "$(cat "$work/before")" ]] && ((SECONDS - start <= limit)); do
			sleep 0.1
		done
		{
			printf -- '--- terminal, once it showed %s, after the typing, before Enter\n' "${prompts[$i]}"
			printf '%s\n' "$typed" | awk 'NF { for (; blank > 0; blank--) print ""; print; next } { blank++ }'
		} >>"$work/answers"
		tmux -L "$pane_server" send-keys -t verify Enter
	done
	# SECONDS counts whole seconds, so the wait lasts between limit and limit + 1 seconds.
	while [[ ! -s "$work/code" ]] && ((SECONDS - start <= limit)); do sleep 0.1; done
	[[ ! -s "$work/code" ]] || code="$(cat "$work/code")"
	# -S - reaches the lines that scrolled into the history; the trailing blank lines go.
	tmux -L "$pane_server" capture-pane -p -J -S - -E - -t verify |
		awk 'NF { for (; blank > 0; blank--) print ""; print; next } { blank++ }' >"$work/shown" ||
		die "could not read the pane of tmux"
	if ! {
		printf '$ ncly'
		printf ' %q' "$@"
		printf '\n'
		cat "$work/answers" || exit 1
		[[ -z "$missed" ]] || printf -- '--- the prompt %s never showed\n' "$missed"
		printf -- '--- terminal, inside tmux whose server does not answer\n'
		cat "$work/shown" || exit 1
		if [[ -n "$code" ]]; then
			printf -- '--- exit %s, before the limit of %ss\n\n' "$code" "$limit"
		else
			printf -- '--- no exit within the limit of %ss\n\n' "$limit"
		fi
	} | tee -a "$transcript"; then
		die "could not record the call in $transcript: what the pane showed stays in $work/shown"
	fi
	[[ -n "$code" && -z "$missed" ]] || return 1
	return "$code"
}

quote() {
	local run file
	run="$(run_folder "${1:-}")"
	if [[ -n "${2:-}" ]]; then
		check_feature "$2"
		set -- "$run/evidence/$2.txt"
	else
		set -- "$run/evidence"/*.txt
	fi
	for file in "$@"; do
		[[ -f "$file" ]] || die "no transcript at $file"
		echo "==> evidence/${file##*/}"
		# Literal replacements: the run folder first, because it sits inside HOME.
		RUN="$run" HOME_DIR="$HOME" awk '
			function swap(s, from, to,   out, i) {
				if (from == "") return s
				out = ""
				while ((i = index(s, from)) > 0) {
					out = out substr(s, 1, i - 1) to
					s = substr(s, i + length(from))
				}
				return out s
			}
			{ print swap(swap($0, ENVIRON["RUN"], "<run>"), ENVIRON["HOME_DIR"], "/Users/me") }
		' "$file"
	done
}

# logout_fake_key logs the fake key of the auth feature out of the real
# keychain, then checks that status no longer finds it. Both calls go to the
# transcript evidence/cleanup.txt. It fails when the key may remain.
logout_fake_key() {
	local run="$1" code=0 answer
	# The key must come from the keychain, never from the caller's variable.
	unset VERIFYNICELY_API_KEY
	answer="$(step "$run" cleanup auth logout "$test_service" --force --json)" || code=$?
	# 0 removed the key, and 2 with NOT_FOUND found none to remove.
	case "$code:$answer" in
	0:* | 2:*'"errors":[{"code":"NOT_FOUND"'*) ;;
	*)
		log "ncly auth logout $test_service exited with $code: see $run/evidence/cleanup.txt"
		return 1
		;;
	esac
	code=0
	answer="$(step "$run" cleanup auth status --json)" || code=$?
	# Only a keychain that answers without the key proves that it is gone:
	# exit 78 with AUTH_MISSING. Another code, such as the 1 of a step that
	# could not record the call, proves nothing.
	case "$code:$answer" in
	78:*'"keychain":"available"'*"{\"service\":\"$test_service\",\"configured\":false,\"source\":\"none\",\"code\":\"AUTH_MISSING\""*) ;;
	*)
		log "ncly auth status still finds a key for $test_service, or no keychain answers: see $run/evidence/cleanup.txt"
		return 1
		;;
	esac
	log "logged the fake key of $test_service out of the keychain"
}

cleanup() {
	local run
	run="$(run_folder "${1:-}")"
	# The auth feature leaves its fake key in the real keychain, after a failed
	# run too. The scratch folder stays until the key is gone, so that a later
	# cleanup can try again.
	if [[ -f "$run/auth-prepared" && -d "$run/scratch" ]]; then
		logout_fake_key "$run" || die "kept the scratch folder of $run: run cleanup again once the keychain answers"
	fi
	rm -rf "$run/scratch" "$run/build"
	[[ ! -e "$run/scratch" && ! -e "$run/build" ]] || die "could not remove the scratch folder of $run"
	log "removed the scratch folder; the evidence stays in $run/evidence"
	ls "$run/evidence"
}

command="${1:-start}"
[[ $# -gt 0 ]] && shift
case "$command" in
start) start ;;
prepare) prepare "$@" ;;
step) step "$@" ;;
tmux) tmux_step "$@" ;;
quote) quote "$@" ;;
cleanup) cleanup "$@" ;;
-h | --help | help) usage help ;;
*) usage ;;
esac
