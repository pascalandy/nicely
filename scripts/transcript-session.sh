#!/usr/bin/env bash
# transcript-session.sh prepares the folder of a fresh agent session for transcript, as Fresh
# session in extensions/transcript/spec/test-transport.md says. It runs on macOS and Linux, from
# a checkout of nicely.
#
#   bash scripts/transcript-session.sh <folder>
#
# <folder> must not exist, and must lie outside any git checkout. Give the session
# <folder>/bin first in PATH and <folder>/work as its working folder. Its ncly is the test
# binary of cmd/ncly, so nothing reaches YouTube, Deepgram, or the OS keychain. Each upload
# appends one line to <folder>/logs/uploads.log. The script seeds three runs for the session:
# team-sync completed and lost its answer, Deepgram refused the key of keynote, and the upload
# of interview has an unknown outcome. Its last line on stdout is the folder.
set -Eeuo pipefail

die() {
	echo "transcript-session: $*" >&2
	exit 1
}

log() { echo "transcript-session: $*" >&2; }

(($# == 1)) || die "usage: bash scripts/transcript-session.sh <folder>"
repo="$(cd "$(dirname "$0")/.." && pwd -P)"
case "$1" in
/*) dir="$1" ;;
*) dir="$PWD/$1" ;;
esac
[[ ! -e "$dir" && ! -L "$dir" ]] || die "$dir exists: give a new folder"
mkdir -p "$dir"
dir="$(cd "$dir" && pwd -P)"
case "$dir" in *"'"*) die "$dir holds a quote, which the scripts of the session cannot hold" ;; esac
if git -C "$dir" rev-parse --git-dir >/dev/null 2>&1; then
	die "$dir lies inside a git checkout, whose source the session could read"
fi

home="$dir/home"
build="$dir/build"
tools="$dir/tools"
data="$home/.local/share/nicely/extensions"
mkdir -p "$build" "$dir/bin" "$tools" "$dir/work/videos" "$dir/logs" "$dir/tmp" "$data" \
	"$home/.config" "$home/.local/state/nicely" "$home/.cache" "$dir/engine/testing" "$dir/engine/tests"

# The test binary keeps its uv cache in the user cache folder of Go's os.UserCacheDir.
if [[ "$(uname -s)" == Darwin ]]; then
	cache="$HOME/Library/Caches/nicely-test/uv"
else
	cache="${XDG_CACHE_HOME:-}"
	[[ "$cache" == /* ]] || cache="$HOME/.cache"
	cache="$cache/nicely-test/uv"
fi

log "building the test binary of cmd/ncly"
(cd "$repo" && go test -c -o "$build/ncly.test" ./cmd/ncly) || die "go test -c failed"
ln -s ncly.test "$build/ncly"
ln -s ncly.test "$build/transcript-external"
log "warming the uv cache of the test binary in $cache, once, with the network"
(cd "$repo" && go test ./cmd/ncly -run '^TestScripts$/^transcript$/^transcript_dry_run$' -count=1 >/dev/null) ||
	die "the scenario transcript_dry_run failed, so the uv cache may be cold"

# The launcher finds its fixtures beside it, in ../tests/fixtures, so the session needs no checkout.
cp "$repo/extensions/transcript/python/testing/launcher.py" "$dir/engine/testing/"
cp -R "$repo/extensions/transcript/python/tests/fixtures" "$dir/engine/tests/fixtures"

# link_host_program links into tools the first program of PATH named $1 that answers the rest
# of the arguments under the scratch home. A shim that needs the real home never counts.
link_host_program() {
	local name="$1" folder
	shift
	local IFS=:
	for folder in $PATH; do
		[[ -f "$folder/$name" && -x "$folder/$name" ]] || continue
		if env -i HOME="$home" PATH="$folder" "$folder/$name" "$@" >/dev/null 2>&1; then
			ln -s "$folder/$name" "$tools/$name"
			log "$name: $folder/$name"
			return
		fi
	done
	die "no $name in PATH answers under the scratch home"
}
link_host_program uv --version
link_host_program python3 -c 'import sys; sys.exit(sys.version_info < (3, 12))'

# The wrapper keeps only tools and the system folders in PATH, so no other ncly-<domain> counts.
for folder in /usr/bin /bin; do
	! compgen -G "$folder/ncly-*" >/dev/null || die "$folder holds an ncly extension"
done
cat >"$dir/bin/ncly" <<EOF
#!/bin/sh
# ncly of this session: the test binary of cmd/ncly, with its test keychain, a scratch home, and
# a fake key of deepgram. Nothing that it runs leaves the machine.
HOME='$home'
XDG_CONFIG_HOME='$home/.config'
XDG_DATA_HOME='$home/.local/share'
XDG_STATE_HOME='$home/.local/state'
XDG_CACHE_HOME='$home/.cache'
TMPDIR='$dir/tmp'
PATH='$tools:/usr/bin:/bin'
DEEPGRAM_API_KEY='fake-deepgram-key-for-the-session'
export HOME XDG_CONFIG_HOME XDG_DATA_HOME XDG_STATE_HOME XDG_CACHE_HOME TMPDIR PATH DEEPGRAM_API_KEY
unset NCLY_CONFIG
exec '$build/ncly' "\$@"
EOF
chmod +x "$dir/bin/ncly"

# transcript installs as just install puts it, with the test transport in its executable.
cp "$repo/extensions/transcript/manifest.toml" "$data/ncly-transcript.toml"
cp -R "$repo/extensions/transcript/locales" "$data/ncly-transcript.locales"
cp -R "$repo/extensions/transcript/skill" "$data/ncly-transcript.skill"
cat >"$data/ncly-transcript" <<EOF
#!/bin/sh
# transcript of this session: transcript-external of the test binary, with the test transport.
WORK='$dir/logs'
NCLY_TEST_UV_CACHE='$cache'
NCLY_TEST_LAUNCHER='$dir/engine/testing/launcher.py'
export WORK NCLY_TEST_UV_CACHE NCLY_TEST_LAUNCHER
exec '$build/transcript-external' "\$@"
EOF
chmod +x "$data/ncly-transcript"
printf '{"format_version":1,"grants":[{"service":"deepgram","domain":"transcript","path":"%s","granted_at":"%s"}]}\n' \
	"$data/ncly-transcript" "$(date -u +%Y-%m-%dT%H:%M:%SZ)" >"$home/.local/state/nicely/grants.json"

# seed runs one call of the session's ncly, whose answer the session never sees.
seed() {
	local id="$1" url="$2" title="$3" status="$4"
	set +e
	NCLY_NO_INPUT=1 NCLY_JSON=1 NCLY_TEST_TITLE="$title" NCLY_TEST_DEEPGRAM_STATUS="$status" \
		"$dir/bin/ncly" transcript run youtube --url "$url" --output "$dir/work/videos" --run-id "$id" >/dev/null 2>&1
	local code=$?
	set -e
	log "seeded $id: $url, Deepgram $status, exit $code"
}
seed team-sync https://youtu.be/TeamSync001 'Team sync' 200
seed keynote https://youtu.be/Keynote0001 'Keynote' 401
seed interview https://youtu.be/Interview01 'Interview' 503
[[ "$(wc -l <"$dir/logs/uploads.log")" -eq 3 ]] || die "the seeds made no three uploads"
echo "$dir"
