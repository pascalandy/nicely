#!/usr/bin/env bash
# verify.sh drives the real ncly in a scratch folder, as SKILL.md describes.
# It runs on macOS and Linux: bash 3.2, and only the flags that BSD and GNU
# tools share.
#
#   verify.sh start                     build ncly, prepare a run, print its folder
#   verify.sh prepare <run> auth [--mac-keychain]
#                                       install the fixture that declares the test service
#   verify.sh step <run> <feature> <ncly args...>
#                                       run ncly in the run's scratch folder and
#                                       record the call in the feature's transcript
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
		sed -n '6,13p' "$0" | sed 's/^# \{0,1\}//'
		exit 0
	fi
	sed -n '6,13p' "$0" | sed 's/^# \{0,1\}//' >&2
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
# that ncly auth reads that service's entry in the real keychain. It never runs.
prepare() {
	local run extensions
	run="$(run_folder "${1:-}")"
	[[ "${2:-}" == auth ]] || die "only the auth feature has a fixture"
	if [[ "$(uname -s)" == Darwin && "${3:-}" != --mac-keychain ]]; then
		die "on macOS, the keychain may ask for a password, so only a human runs the auth feature there, with --mac-keychain"
	fi
	[[ -d "$run/scratch" ]] || die "$run was cleaned up: start a new run"
	extensions="$run/scratch/home/.local/share/nicely/extensions"
	printf '%s\n' '#!/bin/sh' \
		"echo '{\"ok\":false,\"contract_version\":1,\"errors\":[{\"code\":\"RUNTIME\",\"message\":\"verifykeys only declares a service.\",\"hint\":\"ncly auth status\"}]}' >&2" \
		'exit 1' >"$extensions/ncly-verifykeys"
	chmod +x "$extensions/ncly-verifykeys"
	printf '%s\n' 'format_version = 1' 'contract_version = 1' 'domain = "verifykeys"' '' \
		'[[commands]]' 'path = "verifykeys"' 'summary = "verifykeys.summary"' \
		'examples = ["ncly auth status", "ncly auth status --json"]' \
		"requires = { keys = [\"$test_service\"] }" >"$extensions/ncly-verifykeys.toml"
	mkdir -p "$extensions/ncly-verifykeys.locales"
	printf '%s\n' '["verifykeys.summary"]' \
		'description = "Help page: summary of the verify-nicely fixture, which declares a test service."' \
		'other = "Declare the test service of verify-nicely"' >"$extensions/ncly-verifykeys.locales/en.toml"
	: >"$run/auth-prepared"
	log "verifykeys declares the service $test_service, whose variable is VERIFYNICELY_API_KEY"
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

cleanup() {
	local run
	run="$(run_folder "${1:-}")"
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
quote) quote "$@" ;;
cleanup) cleanup "$@" ;;
-h | --help | help) usage help ;;
*) usage ;;
esac
