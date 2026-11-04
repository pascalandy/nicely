#!/usr/bin/env bash
# A Linux verify run owns one D-Bus session and unlocked Secret Service.
# Its private environment never inherits the caller's bus or keyring control.
set -Eeuo pipefail

fail() {
	echo "verify-nicely: $*" >&2
	exit 1
}

require_programs() {
	local program
	for program in dbus-run-session gnome-keyring-daemon; do
		command -v "$program" >/dev/null || fail "$program is missing: required for the private Linux keyring"
	done
}

start_session() {
	local state="$1" home="$2" runtime self i
	require_programs
	mkdir -p "$state"
	# Unix socket paths are short, even when the run folder has a long name.
	runtime="$(mktemp -d "${TMPDIR:-/tmp}/nk.XXXXXXXX")"
	chmod 700 "$runtime" "$state"
	printf '%s\n' "$runtime" >"$state/runtime"
	: >"$runtime/.verify-nicely-keyring"
	self="$(cd "$(dirname "$0")" && pwd -P)/keyring.sh"
	(
		code=0
		env -i PATH="$PATH" HOME="$home" XDG_CONFIG_HOME="$home/.config" \
			XDG_DATA_HOME="$home/.local/share" XDG_STATE_HOME="$home/.local/state" \
			XDG_CACHE_HOME="$home/.cache" XDG_RUNTIME_DIR="$runtime" \
			dbus-run-session -- bash "$self" serve "$state" "$runtime" || code=$?
		# An abandoned run can lose its state before the private bus exits.
		if [[ -f "$runtime/.verify-nicely-keyring" ]]; then rm -rf "${runtime:?}"; fi
		if [[ -d "$state" ]]; then printf '%s\n' "$code" >"$state/stopped"; fi
	) >"$state/session.log" 2>&1 </dev/null &
	for ((i = 0; i < 100; i++)); do
		[[ ! -f "$state/stopped" ]] || break
		[[ ! -s "$state/address" ]] || return 0
		sleep 0.1
	done
	: >"$state/stop"
	fail "private keyring did not start; see $state/session.log; run cleanup before removing scratch"
}

serve() {
	local state="$1" runtime="$2" daemon i started=$SECONDS
	# Only a public fixture password protects this disposable keyring.
	printf 'verify-nicely-private\n' | gnome-keyring-daemon --foreground --unlock \
		--components=secrets --control-directory="$runtime/keyring" >"$state/daemon.log" 2>&1 &
	daemon=$!
	printf '%s\n' "$daemon" >"$state/daemon-pid"
	trap 'kill "$daemon" 2>/dev/null || true; wait "$daemon" 2>/dev/null || true' EXIT
	trap 'exit 1' TERM INT
	for ((i = 0; i < 100; i++)); do
		kill -0 "$daemon" 2>/dev/null || fail "private gnome-keyring-daemon stopped"
		if [[ -S "$runtime/keyring/control" ]]; then
			gnome-keyring-daemon --start --components=secrets --control-directory="$runtime/keyring" >/dev/null
			printf '%s\n' "$DBUS_SESSION_BUS_ADDRESS" >"$state/address"
			break
		fi
		sleep 0.1
	done
	[[ -s "$state/address" ]] || fail "private keyring control socket did not start"
	while [[ -d "$state" && -f "$runtime/.verify-nicely-keyring" && ! -f "$state/stop" ]] && ((SECONDS - started < 21600)); do
		kill -0 "$daemon" 2>/dev/null || fail "private gnome-keyring-daemon stopped"
		sleep 1
	done
	# Exit before the EXIT trap loses this function's local daemon PID.
	exit 0
}

stop_session() {
	local state="$1" runtime i
	[[ -d "$state" ]] || return 0
	: >"$state/stop"
	for ((i = 0; i < 100; i++)); do
		if [[ -f "$state/stopped" ]]; then
			runtime="$(cat "$state/runtime")"
			[[ -n "$runtime" ]] || fail "private keyring runtime path is missing"
			if [[ -e "$runtime" ]]; then
				[[ -f "$runtime/.verify-nicely-keyring" ]] || fail "private keyring runtime marker is missing"
				rm -rf "${runtime:?}"
			fi
			echo "private Linux keyring and D-Bus session stopped"
			return 0
		fi
		sleep 0.1
	done
	fail "private keyring did not stop; kept $state and its runtime"
}

case "${1:-}" in
check) require_programs ;;
start) start_session "$2" "$3" ;;
serve) serve "$2" "$3" ;;
stop) stop_session "$2" ;;
*) fail "usage: keyring.sh check|start <state> <home>|serve <state> <runtime>|stop <state>" ;;
esac
