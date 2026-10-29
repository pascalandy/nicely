#!/usr/bin/env bash
# preflight.sh checks this host for a UAT, for just preflight, and changes
# nothing. It prints one line for each prerequisite, and a hint under each one
# that is missing or that warns. It exits 1 when one is missing. Each argument
# is a fixture folder, whose SHA256SUMS and locks it checks, as
# docs/qa/fixtures.md says. It runs on macOS and Linux, with bash 3.2.
# scripts/preflight-test.sh proves it. See #356.
set -Euo pipefail

root="$(cd "${BASH_SOURCE[0]%/*}/.." && pwd -P)"
os="$(uname -s)"
missing=0
warnings=0

pass() { printf 'ok    %s\n' "$1"; }

warn() {
	printf 'warn  %s\n      hint: %s\n' "$1" "$2"
	warnings=$((warnings + 1))
}

fail() {
	printf 'FAIL  %s\n      hint: %s\n' "$1" "$2"
	missing=$((missing + 1))
}

# install_hint prints how to install a package: brew on macOS, then the first of pacman, apt,
# dnf, and brew on Linux, as doctor orders them. $2 names the package for pacman when it differs.
install_hint() {
	local manager
	if [[ "$os" == Darwin ]]; then
		printf 'brew install %s\n' "$1"
		return
	fi
	for manager in pacman apt dnf brew; do
		command -v "$manager" >/dev/null || continue
		case "$manager" in
		pacman) printf 'sudo pacman -S %s\n' "${2:-$1}" ;;
		apt) printf 'sudo apt install %s\n' "$1" ;;
		dnf) printf 'sudo dnf install %s\n' "$1" ;;
		brew) printf 'brew install %s\n' "$1" ;;
		esac
		return
	done
	printf 'install %s\n' "$1"
}

check_tmux() {
	local version
	if command -v tmux >/dev/null && version="$(tmux -V 2>/dev/null)"; then
		pass "$version at $(command -v tmux)"
	else
		fail "tmux is missing or its version probe failed: the terminal steps of verify-nicely need it" "$(install_hint tmux)"
	fi
}

# check_python reads the minimum Python from the header of transcript.py, as dev-preferences.md
# says.
check_python() {
	local version minimum hint
	minimum="$(sed -n 's/^# requires-python = ">=\([0-9.]*\)"$/\1/p' "$root/extensions/transcript/python/transcript.py")"
	hint="run under uv: uv run --no-project --python $minimum just preflight"
	if ! command -v python3 >/dev/null; then
		fail "python3 is missing: the transcript scenarios need Python $minimum or later" "$hint"
		return
	fi
	if version="$(python3 -c 'import sys; print("%d.%d.%d" % sys.version_info[:3]); sys.exit(sys.version_info < tuple(map(int, sys.argv[1].split("."))))' "$minimum")"; then
		pass "python3 $version at $(command -v python3)"
	else
		fail "python3 $version at $(command -v python3): Python $minimum or later required" "$hint"
	fi
}

# check_bash asks the first bash of PATH, which runs the scripts of just check.
check_bash() {
	local version
	version="$(bash -c 'printf "%s\n" "$BASH_VERSION"')"
	if [[ "${version%%.*}" -ge 5 ]]; then
		pass "bash $version at $(command -v bash)"
	else
		# Remove this warning once #296 is fixed.
		warn "bash $version at $(command -v bash): under bash 3.2, the completion scenarios of just test fail, see #296" "$(install_hint bash)"
	fi
}

# check_modules lists, without the network, the packages that just check builds in each Go
# module. A package whose module the cache lacks fails, as on the first check of a clone.
check_modules() {
	local dir
	local -a patterns
	if ! command -v go >/dev/null; then
		fail "go is missing" "install Go from https://go.dev/dl/"
		return
	fi
	for dir in . tools testdata/extensions/standalone testdata/extensions/standalone-named; do
		case "$dir" in
		.) patterns=(./... ./testdata/extensions/...) ;;
		tools) patterns=(tool ./cmd/...) ;;
		*) patterns=(./...) ;;
		esac
		if (cd "$root/$dir" && GOPROXY=off GOFLAGS=-mod=readonly go list -deps -test "${patterns[@]}" >/dev/null 2>&1); then
			pass "Go modules of $dir are in the module cache"
		else
			fail "Go modules of $dir are missing from the module cache" "run go mod download in $dir, with the network"
		fi
	done
}

check_codex() {
	local version
	if command -v codex >/dev/null && version="$(codex --version 2>/dev/null)"; then
		pass "${version%%$'\n'*} at $(command -v codex)"
	else
		fail "codex is missing or its version probe failed: the agent extension and its demos need it" "$(install_hint codex openai-codex)"
	fi
}

# check_ncly finds each ncly in PATH. The first one runs, and its version names the commit that
# it was built from, or a release tag. It warns when that is not HEAD, and when it hides another.
check_ncly() {
	local dir seen=":" first version head tags i
	local -a found=()
	while IFS= read -r dir; do
		[[ -n "$dir" && -x "$dir/ncly" && ! -d "$dir/ncly" && "$seen" != *":$dir:"* ]] || continue
		seen="$seen$dir:"
		found[${#found[@]}]="$dir/ncly"
	done < <(printf '%s\n' "$PATH" | tr ':' '\n')
	if [[ ${#found[@]} -eq 0 ]]; then
		fail "ncly is missing from PATH" "just install, or install the release that the UAT tests"
		return
	fi
	first="${found[0]}"
	version="$("$first" --version 2>/dev/null)"
	head="$(git -C "$root" rev-parse --short=12 HEAD)"
	tags="$(git -C "$root" tag --points-at HEAD | tr '\n' ' ')"
	case "$version" in
	*"-$head" | *"-$head+dirty" | "ncly ${tags%% *}") pass "$version at $first, built from HEAD $head" ;;
	*) warn "$version at $first is not built from HEAD $head" "just install, unless the UAT tests a release" ;;
	esac
	for ((i = 1; i < ${#found[@]}; i++)); do
		warn "$first hides ${found[$i]}, $("${found[$i]}" --version 2>/dev/null)" "remove the ncly that the UAT does not test, or reorder PATH"
	done
}

# locked reports whether a file is read-only, and also locked with uchg on macOS.
locked() {
	[[ -f "$1" && ! -w "$1" ]] && { [[ "$os" != Darwin ]] || [[ "$(stat -f %Sf "$1")" == *uchg* ]]; }
}

# check_fixture checks a fixture folder: each checksum of its SHA256SUMS, then that each
# original and SHA256SUMS are locked.
check_fixture() {
	local folder="$1" file sum
	if [[ ! -f "$folder/SHA256SUMS" ]]; then
		fail "fixture folder $folder holds no SHA256SUMS" "mount the fixture volume, whose path the fleet skill gives"
		return
	fi
	sum="sha256sum"
	command -v sha256sum >/dev/null || sum="shasum -a 256"
	if ! (cd "$folder" && $sum -c SHA256SUMS >/dev/null 2>&1); then
		fail "fixture folder $folder: a checksum differs from SHA256SUMS" "stop the UAT and tell Pascal: an original changed"
		return
	fi
	while read -r _ file; do
		if ! locked "$folder/${file#\*}"; then
			fail "fixture $folder/${file#\*} is not locked" "ask Pascal to lock it, as docs/qa/fixtures.md says"
			return
		fi
	done <"$folder/SHA256SUMS"
	if ! locked "$folder/SHA256SUMS"; then
		fail "$folder/SHA256SUMS is not locked" "ask Pascal to lock it, as docs/qa/fixtures.md says"
		return
	fi
	pass "fixture folder $folder: checksums OK, originals locked"
}

check_tmux
check_python
check_bash
check_modules
check_codex
check_ncly
for folder in "$@"; do check_fixture "$folder"; done
printf 'preflight: %s missing, %s warnings\n' "$missing" "$warnings"
[[ "$missing" == 0 ]]
