# shellcheck shell=bash
# shellcheck disable=SC2154 # verify.sh sets script_dir before prepare.sh sources this file
# prepare-template.sh holds the fixture of the template feature, which prepare.sh sources:
#
#   verify.sh prepare <run> template    render the extension template as probe, install it in the
#                                       scratch home, with the real python3 first in PATH

# prepare_template renders the extension template of the checkout as the extension probe, with
# the Copier that the template pins, as an agent does with the command of the nicely-extensions
# skill. The template of a change has no release tag, so the source is a scratch Git repository of
# copier.yml and template/, as the checkout holds them, tagged v0.0.0. The go.mod of probe then
# builds against the checkout, and just install puts probe in the scratch data folder. The real
# python3, 3.9 or later, goes first in the wrapper's PATH, because a shim needs the real home.
prepare_template() {
	local run="$1" home="$1/scratch/home" tools="$1/scratch/tools" src="$1/scratch/template-source"
	local repo="$1/scratch/ncly-probe" checkout pin path gocache gomodcache gopath goenv
	local -a copier
	[[ -d "$run/scratch" ]] || die "$run was cleaned up: start a new run"
	checkout="$(git -C "$script_dir" rev-parse --show-toplevel)" || die "run it from a checkout of nicely"
	command -v just >/dev/null || die "just is missing: install it, such as with brew install just"
	command -v uv >/dev/null || die "uv is missing: install it, such as with brew install uv"
	pin="$(sed -n 's/^copier := "\(.*\)"$/\1/p' "$checkout/template/justfile")"
	[[ -n "$pin" ]] || die "$checkout/template/justfile pins no Copier"
	mkdir -p "$src" "$tools"
	cp "$checkout/copier.yml" "$src/"
	cp -R "$checkout/template" "$src/template"
	if ! { git -C "$src" init -q && git -C "$src" add -A &&
		git -C "$src" -c user.name=verify-nicely -c user.email=verify-nicely@example.com -c commit.gpgsign=false commit -qm template &&
		git -C "$src" -c tag.gpgsign=false tag v0.0.0; }; then
		die "could not tag the template source in $src"
	fi
	read -r -a copier <<<"$pin"
	log "running ${copier[*]} copy into $repo"
	"${copier[@]}" copy --quiet --vcs-ref v0.0.0 --defaults --data domain=probe --data module=example.com/ncly-probe "$src" "$repo" ||
		die "Copier could not render the template"
	if ! { go -C "$repo" mod edit -replace "github.com/pascalandy/nicely=$checkout" && go -C "$repo" mod tidy; }; then
		die "go mod tidy failed in $repo"
	fi
	# The build keeps the caller's Go caches and settings, which live under the real home.
	gocache="$(go env GOCACHE)"
	gomodcache="$(go env GOMODCACHE)"
	gopath="$(go env GOPATH)"
	goenv="$(go env GOENV)"
	log "running just install of probe with HOME=$home"
	(cd "$repo" && env HOME="$home" XDG_CONFIG_HOME="$home/.config" XDG_DATA_HOME="$home/.local/share" \
		XDG_STATE_HOME="$home/.local/state" XDG_CACHE_HOME="$home/.cache" \
		GOCACHE="$gocache" GOMODCACHE="$gomodcache" GOPATH="$gopath" GOENV="$goenv" just install) ||
		die "just install of probe failed"
	link_host_program python3 "$home" "$tools" -c 'import sys; sys.exit(sys.version_info < (3, 9))'
	path="$(wrapper_path "$run")"
	write_wrapper "$run/scratch/bin/ncly" "$run/build/ncly" "$home" "$tools:$path"
	: >"$run/template-prepared"
	log "probe lives in $home/.local/share/nicely/extensions, and $repo holds its repository"
	log "the wrapper now runs $run/build/ncly, with the python3 of $tools first in PATH"
}
