#!/usr/bin/env bash
# install.sh builds ncly and the first-party extensions for just install, or only the names that
# it gets: ncly, or the folder of a first-party extension, such as agent. ncly goes into
# $XDG_BIN_HOME or ~/.local/bin, and each extension into Nicely's data folder. It runs on macOS and
# Linux, with bash 3.2. scripts/install-test.sh proves it. See #375.
set -Eeuo pipefail
cd "${BASH_SOURCE[0]%/*}/.."
die() {
	echo "install: $*" >&2
	exit 1
}
[[ "${HOME:-}" == /* ]] || die "HOME must be an absolute path"
# A name is ncly, or the folder of a first-party extension that has a main in cmd/ncly-<slug>/.
valid="ncly"
for main in extensions/*/cmd/ncly-*/; do
	source="${main%/cmd/*}"
	valid="$valid ${source##*/}"
done
for name in "$@"; do
	known=""
	for each in $valid; do
		if [[ "$name" == "$each" ]]; then known=1; fi
	done
	[[ -n "$known" ]] || die "unknown name \"$name\", choose from: $valid"
done
# chosen succeeds when the command line names $1, or names nothing.
picks=" $* "
chosen() {
	[[ "$picks" == "  " || "$picks" == *" $1 "* ]]
}
# An XDG variable counts only when it holds an absolute path, as Configuration in contract.md
# says. uv too reads XDG_BIN_HOME only when it holds an absolute path.
bin="${XDG_BIN_HOME:-}"
[[ "$bin" == /* ]] || bin="$HOME/.local/bin"
data="${XDG_DATA_HOME:-}"
[[ "$data" == /* ]] || data="$HOME/.local/share"
extensions="$data/nicely/extensions"
# Every file is built in a staging folder beside its destination first, so a failed build
# keeps the previous install, and each move is a rename on the same file system.
stage_bin=""
stage=""
trap 'rm -rf ${stage_bin:+"$stage_bin"} ${stage:+"$stage"}' EXIT
if chosen ncly; then
	mkdir -p "$bin"
	stage_bin="$(mktemp -d "$bin/.ncly-install.XXXXXX")"
	go build -o "$stage_bin/ncly" ./cmd/ncly
fi
names=()
for main in extensions/*/cmd/ncly-*/; do
	main="${main%/}"
	name="${main##*/}"
	source="${main%/cmd/*}"
	chosen "${source##*/}" || continue
	if [[ -z "$stage" ]]; then
		mkdir -p "$extensions"
		stage="$(mktemp -d "$extensions/.install.XXXXXX")"
	fi
	go build -o "$stage/$name" "./$main"
	cp "$source/manifest.toml" "$stage/$name.toml"
	cp -R "$source/locales" "$stage/$name.locales"
	if [[ -d "$source/skill" ]]; then cp -R "$source/skill" "$stage/$name.skill"; fi
	names+=("$name")
done
if [[ -n "$stage_bin" ]]; then
	mv -f "$stage_bin/ncly" "$bin/ncly"
	echo "installed $bin/ncly"
fi
# bash 3.2, as macOS ships it, rejects an empty array under set -u without this form.
for name in ${names[@]+"${names[@]}"}; do
	rm -rf "${extensions:?}/$name.locales" "${extensions:?}/$name.skill"
	for file in "$name.toml" "$name.locales" "$name.skill" "$name"; do
		if [[ -e "$stage/$file" ]]; then mv -f "$stage/$file" "$extensions/$file"; fi
	done
	echo "installed $extensions/$name"
done
if [[ -n "$stage_bin" ]]; then
	case ":$PATH:" in
	*":$bin:"*) ;;
	*) echo "install: add $bin to PATH to run ncly" >&2 ;;
	esac
fi
