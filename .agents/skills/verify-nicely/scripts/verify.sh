#!/usr/bin/env bash
# verify.sh drives the real ncly in a scratch folder, as SKILL.md describes.
# It runs on macOS and Linux: bash 3.2, and only the flags that BSD and GNU
# tools share.
#
#   verify.sh start                     build ncly, prepare a run, print its folder
#   verify.sh prepare <run> <feature> [--mac-keychain]
#                                       install the fixture of a feature, as prepare.sh lists them
#   verify.sh step <run> <feature> <ncly args...>
#                                       run ncly in the run's scratch folder and
#                                       record the call in the feature's transcript
#   verify.sh tmux <run> <feature> [--answer <prompt> <keys>]... <ncly args...>
#                                       run ncly on a terminal inside tmux, whose
#                                       server does not answer, type each answer once
#                                       its prompt shows, and record it the same way
#   verify.sh quote <run> [<feature>]   print transcripts with placeholders for home paths
#   verify.sh cleanup <run>             remove the scratch folder, keep the evidence
#   verify.sh prune --keep <n>          remove the runs before the newest <n>, once cleaned up
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

# runs_root prints the folder of the runs. Not /tmp: Linux mounts it as tmpfs and macOS purges
# it, and the evidence must outlive the run.
runs_root() { printf '%s\n' "${XDG_CACHE_HOME:-$HOME/.cache}/nicely/verify"; }

# stamp prints the time in ISO 8601 with its offset, such as 2026-10-29T09:12:03-04:00. BSD
# date lacks %:z, so the colon goes in here.
stamp() {
	local t
	t="$(date +%Y-%m-%dT%H:%M:%S%z)"
	printf '%s:%s\n' "${t%??}" "${t#"${t%??}"}"
}

start() {
	local repo run stamp home path
	command -v go >/dev/null || die "go is missing: install it from https://go.dev/dl/"
	repo="$(git -C "$script_dir" rev-parse --show-toplevel)" || die "run it from a checkout of nicely"
	stamp="$(date +%Y%m%d-%H%M%S)-$$"
	run="$(runs_root)/$stamp"
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
# shellcheck disable=SC2034 # prepare-auth.sh and cleanup.sh read it
test_service="verifynicely"

# prepare.sh holds the prepare step and the fixture of each feature that needs one.
# shellcheck source=/dev/null
. "$script_dir/prepare.sh"
# cleanup.sh holds cleanup, quote, and prune.
# shellcheck source=/dev/null
. "$script_dir/cleanup.sh"
# tmux.sh holds the tmux step.
# shellcheck source=/dev/null
. "$script_dir/tmux.sh"

step() {
	local run feature transcript out err code at
	run="$(run_folder "${1:-}")"
	check_feature "${2:-}"
	feature="$2"
	shift 2
	[[ -d "$run/scratch" ]] || die "$run was cleaned up: start a new run"
	transcript="$run/evidence/$feature.txt"
	out="$run/scratch/step.out"
	err="$run/scratch/step.err"
	code=0
	at="$(stamp)"
	"$run/scratch/bin/ncly" "$@" >"$out" 2>"$err" || code=$?
	# A caller may run step where errexit is off, as cleanup does, and the
	# condition turns it off in the group too, so each read fails the group
	# itself. A call that the transcript misses stops here, and its outputs
	# stay in scratch.
	if ! {
		printf '$ ncly'
		printf ' %q' "$@"
		printf '\n--- at %s\n--- stdout\n' "$at"
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

command="${1:-start}"
[[ $# -gt 0 ]] && shift
case "$command" in
start) start ;;
prepare) prepare "$@" ;;
step) step "$@" ;;
tmux) tmux_step "$@" ;;
quote) quote "$@" ;;
cleanup) cleanup "$@" ;;
prune) prune "$@" ;;
-h | --help | help) usage help ;;
*) usage ;;
esac
