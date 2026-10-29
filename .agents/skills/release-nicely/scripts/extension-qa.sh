#!/usr/bin/env bash
# extension-qa.sh tests an installed extension with the ncly that brew-qa.sh installed on a Mac (COR-101).
# It installs video-convert from its highest tag into scratch XDG folders, then runs describe, doctor,
# a dry run, and one conversion of a clip that ffmpeg makes, and checks the output with ffprobe.
# Usage: bash extension-qa.sh vX.Y.Z. The config, extensions, and keys of the host stay as they are.
set -Eeuo pipefail
want="${1:?usage: extension-qa.sh vX.Y.Z}"
repo=https://github.com/pascalandy/ncly-video-convert
fail() {
	printf 'extension-qa: %s\n' "$*" >&2
	exit 1
}
pass() {
	printf 'extension-qa: %s\n' "$*"
}

# Homebrew's folders come first, as in brew-qa.sh, so ncly is the one that Homebrew installed.
PATH="/opt/homebrew/bin:/opt/homebrew/sbin:${PATH}"
ncly="$(command -v ncly)" || fail "ncly is missing: run brew-qa.sh first"
# A shim of mise fails once the XDG folders move, so each tool is found first. PATH then keeps
# only their folders and those of the system.
path="${ncly%/*}"
for tool in go just git jq HandBrakeCLI ffmpeg ffprobe; do
	found="$(mise which "${tool}" 2>/dev/null || command -v "${tool}")" || fail "${tool} is missing: install it, then run extension-qa.sh again"
	path="${path}:${found%/*}"
done
export PATH="${path}:/usr/bin:/bin:/usr/sbin:/sbin"
pass "found ncly, go, just, git, jq, HandBrakeCLI, ffmpeg, and ffprobe"

got="$(ncly --version)"
[[ "${got}" == "ncly ${want}" ]] || fail "ncly --version printed \"${got}\", want \"ncly ${want}\""
pass "${ncly} prints ${got}"

scratch="$(mktemp -d "${TMPDIR:-/tmp}/ncly-extension-qa.XXXXXX")"
trap 'rm -rf "${scratch}"' EXIT
# On Linux, Go takes its build cache from XDG_CACHE_HOME, so GOCACHE keeps the cache of the host.
GOCACHE="$(go env GOCACHE)"
export GOCACHE NCLY_NO_INPUT=1
export XDG_DATA_HOME="${scratch}/data" XDG_CONFIG_HOME="${scratch}/config" XDG_STATE_HOME="${scratch}/state" XDG_CACHE_HOME="${scratch}/cache"

tags="$(git ls-remote --tags --refs --sort=-v:refname "${repo}")" || fail "git ls-remote cannot list the tags of ${repo}"
tag="$(sed -n 's|^.*refs/tags/\(v[0-9]*\.[0-9]*\.[0-9]*\)$|\1|p' <<<"${tags}" | sed -n 1p)"
[[ -n "${tag}" ]] || fail "${repo} has no tag vX.Y.Z"
clone="${scratch}/ncly-video-convert"
git -c advice.detachedHead=false clone --quiet --branch "${tag}" "${repo}" "${clone}" || fail "git clone of ${repo} at ${tag} failed"
commit="$(git -C "${clone}" rev-parse HEAD)"
pass "cloned video-convert ${tag} at ${commit}"

(cd "${clone}" && just install) >"${scratch}/install.log" 2>&1 || {
	cat "${scratch}/install.log" >&2
	fail "just install of video-convert ${tag} failed"
}
[[ -x "${XDG_DATA_HOME}/nicely/extensions/ncly-video-convert" ]] || fail "just install wrote no ncly-video-convert"
pass "just install put ncly-video-convert in the scratch extensions folder"

# fct_ncly runs ncly with --json, and fails unless it exits 0 and answers "ok":true.
fct_ncly() {
	local status=0
	ncly "$@" --json >"${scratch}/answer" 2>"${scratch}/stderr" || status=$?
	if [[ "${status}" != 0 ]] || ! jq -e '.ok == true' "${scratch}/answer" >/dev/null 2>&1; then
		cat "${scratch}/answer" "${scratch}/stderr" >&2
		fail "ncly $* --json exited ${status} without \"ok\":true"
	fi
}

fct_ncly describe video convert
jq -e '.commands | any(.path == "video convert" and .source == "extension")' "${scratch}/answer" >/dev/null ||
	fail "ncly describe names no video convert from an extension: $(cat "${scratch}/answer")"
pass "ncly describe video convert finds the extension"

fct_ncly doctor video-convert
pass "ncly doctor video-convert passes: $(jq -r '[.checks[].id] | join(", ")' "${scratch}/answer")"

work="${scratch}/work"
clip="${work}/clip.mp4"
mkdir "${work}"
ffmpeg -nostdin -v error -f lavfi -i testsrc=size=320x240:rate=30 -f lavfi -i sine=frequency=440:sample_rate=48000 \
	-t 2 -c:v libx264 -pix_fmt yuv420p -c:a aac "${clip}" || fail "ffmpeg cannot make ${clip}"
pass "ffmpeg made a clip of 2 seconds"

fct_ncly video convert "${clip}" --dry-run
jq -e '.dry_run == true and (.output | endswith("/clip.converted.mp4"))' "${scratch}/answer" >/dev/null ||
	fail "the dry run names no clip.converted.mp4: $(cat "${scratch}/answer")"
[[ ! -e "${work}/clip.converted.mp4" ]] || fail "the dry run wrote ${work}/clip.converted.mp4"
pass "the dry run answers ok and writes nothing"

fct_ncly video convert "${clip}"
output="$(jq -r '.output' "${scratch}/answer")"
[[ "${output##*/}" == clip.converted.mp4 && -f "${work}/clip.converted.mp4" ]] || fail "the conversion wrote no clip.converted.mp4 beside the clip: ${output}"
size="$(wc -c <"${work}/clip.converted.mp4" | tr -d ' ')"
[[ "$(jq -r '.size_bytes' "${scratch}/answer")" == "${size}" ]] || fail "size_bytes differs from the ${size} bytes of ${output}"
pass "ncly video convert wrote clip.converted.mp4, ${size} bytes"

streams="$(ffprobe -v error -select_streams v -show_entries stream=codec_type -of csv=p=0 "${output}")"
[[ "${streams}" == *video* ]] || fail "ffprobe finds no video stream in ${output}"
duration="$(ffprobe -v error -show_entries format=duration -of csv=p=0 "${output}")"
awk -v d="${duration}" 'BEGIN { exit !(d >= 1.9 && d <= 2.1) }' || fail "ffprobe gives ${output} a duration of ${duration} s, want 2 s"
pass "ffprobe finds a video stream and a duration of ${duration} s"

if [[ "$(uname -s)" == Darwin ]]; then os="macOS $(sw_vers -productVersion)"; else os="$(uname -sr)"; fi
pass "video-convert ${tag} at ${commit} installs and converts with ${got} on ${os}"
