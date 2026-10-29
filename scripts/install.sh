#!/usr/bin/env bash
# install.sh builds ncly into ~/.local/bin and the first-party extensions into Nicely's data
# folder, for just install. It runs on macOS and Linux, with bash 3.2.
set -Eeuo pipefail
cd "${BASH_SOURCE[0]%/*}/.."
die() {
	echo "install: $*" >&2
	exit 1
}
[[ "${HOME:-}" == /* ]] || die "HOME must be an absolute path"
bin="$HOME/.local/bin"
# An XDG variable counts only when it holds an absolute path, as Configuration in contract.md says.
data="${XDG_DATA_HOME:-}"
[[ "$data" == /* ]] || data="$HOME/.local/share"
extensions="$data/nicely/extensions"
mkdir -p "$bin" "$extensions"
# Every file is built in a staging folder beside its destination first, so a failed build
# keeps the previous install, and each move is a rename on the same file system.
stage_bin="$(mktemp -d "$bin/.ncly-install.XXXXXX")"
stage="$(mktemp -d "$extensions/.install.XXXXXX")"
trap 'rm -rf "$stage_bin" "$stage"' EXIT
go build -o "$stage_bin/ncly" ./cmd/ncly
names=()
for main in extensions/*/cmd/ncly-*/; do
	main="${main%/}"
	name="${main##*/}"
	source="${main%/cmd/*}"
	go build -o "$stage/$name" "./$main"
	cp "$source/manifest.toml" "$stage/$name.toml"
	cp -R "$source/locales" "$stage/$name.locales"
	if [[ -d "$source/skill" ]]; then cp -R "$source/skill" "$stage/$name.skill"; fi
	names+=("$name")
done
mv -f "$stage_bin/ncly" "$bin/ncly"
echo "installed $bin/ncly"
# bash 3.2, as macOS ships it, rejects an empty array under set -u without this form.
for name in ${names[@]+"${names[@]}"}; do
	rm -rf "${extensions:?}/$name.locales" "${extensions:?}/$name.skill"
	for file in "$name.toml" "$name.locales" "$name.skill" "$name"; do
		if [[ -e "$stage/$file" ]]; then mv -f "$stage/$file" "$extensions/$file"; fi
	done
	echo "installed $extensions/$name"
done
case ":$PATH:" in
*":$bin:"*) ;;
*) echo "install: add $bin to PATH to run ncly" >&2 ;;
esac
