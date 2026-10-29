# shellcheck shell=bash
# tmux.sh holds the tmux step of verify.sh, which sources it:
#
#   verify.sh tmux <run> <feature> [--answer <prompt> <keys>]... <ncly args...>
#                                       run ncly on a terminal inside tmux, whose
#                                       server does not answer, type each answer once
#                                       its prompt shows, and record it the same way

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
	local run feature transcript work code="" limit=5 start i missed="" typed at
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
	at="$(stamp)"
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
		printf '\n--- at %s\n' "$at"
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
