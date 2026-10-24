set shell := ["bash", "-euo", "pipefail", "-c"]

# The identity of the extension, which manifest.toml declares
domain := `sed -n 's/^domain = "\([a-z][a-z0-9]*\)"$/\1/p' manifest.toml`
name := `sed -n '/^\[/q; s/^name = "\([a-z][a-z0-9]*\)"$/\1/p' manifest.toml`
slug := domain + if name == "" { "" } else { "-" + name }

# The Copier of just update, pinned and run through uv. No template task runs, so it needs no --trust
copier := "uv tool run copier@9.19.1"

# List the recipes
default:
    @just --list

# Run every check: go mod tidy -diff, gofmt, go vet, and go test, whose test kit installs the extension into scratch folders
check:
    go mod tidy -diff
    @test -z "$(gofmt -l .)" || { gofmt -l . >&2; echo 'check: format these files with gofmt -w .' >&2; exit 1; }
    go vet ./...
    go test -count=1 ./...

# Build ncly-<slug> and install it with its sidecars into Nicely's extensions folder, in place of its previous install
install:
    #!/usr/bin/env bash
    set -Eeuo pipefail
    die() {
        echo "install: $*" >&2
        exit 1
    }
    name="ncly-{{ slug }}"
    [[ "$name" != ncly- ]] || die "manifest.toml declares no domain"
    [[ "${HOME:-}" == /* ]] || die "HOME must be an absolute path"
    # An XDG variable counts only when it holds an absolute path, as Configuration in the contract says.
    data="${XDG_DATA_HOME:-}"
    [[ "$data" == /* ]] || data="$HOME/.local/share"
    extensions="$data/nicely/extensions"
    mkdir -p "$extensions"
    # One go build gives the executable and the program that exports its sidecars. Both land in a
    # staging folder beside the destination, so a failed build keeps the previous install.
    stage="$(mktemp -d "$extensions/.install.XXXXXX")"
    trap 'rm -rf "$stage"' EXIT
    go build -o "$stage/" "./cmd/$name" ./internal/export
    "$stage/export" "$stage"
    # The previous sidecars move into the staging folder, which goes at exit. Each move is a
    # rename on the same file system, and the executable moves last.
    for file in "$name.locales" "$name.skill"; do
        if [[ -e "$extensions/$file" ]]; then mv "$extensions/$file" "$stage/previous-$file"; fi
    done
    for file in "$name.toml" "$name.locales" "$name.skill" "$name"; do
        if [[ -e "$stage/$file" ]]; then mv -f "$stage/$file" "$extensions/$file"; fi
    done
    echo "installed $extensions/$name"

# Move the repository to another Nicely version, such as v0.19.0: the template first, then the SDK, on a clean Git tree
update version:
    #!/usr/bin/env bash
    set -Eeuo pipefail
    version={{ quote(version) }}
    if [[ ! "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        echo "Usage: just update vX.Y.Z, such as just update v0.19.0" >&2
        exit 2
    fi
    if ! head="$(git rev-parse --verify --quiet HEAD)" || [[ -n "$(git status --porcelain)" ]]; then
        echo "update: the update needs a clean Git tree. Commit your changes, then run just update $version again" >&2
        git status --short >&2 || true
        exit 1
    fi
    # The recipe never resets, stashes, or commits: the author decides what to keep.
    if ! {{ copier }} update --defaults --conflict inline --vcs-ref "$version"; then
        echo "update: Copier stopped, so files can be incomplete without a conflict marker." >&2
        echo "Return to the commit before the update, remove the files that it added, and run it again:" >&2
        echo "  git restore --source=$head --staged --worktree -- ." >&2
        echo "  git clean -d --dry-run" >&2
        echo "  git clean -d --force" >&2
        echo "  just update $version" >&2
        exit 1
    fi
    # Copier writes a conflict as markers, or as a .rej file, and can still exit 0. It marks a
    # conflicted file as unmerged in the index. Look for all three before go.mod changes.
    conflicts="$(
        git ls-files --unmerged | cut -f2
        find . -name '*.rej' -not -path './.git/*' | sed 's|^\./||'
        { git diff --name-only HEAD; git ls-files --others --exclude-standard; } | while IFS= read -r file; do
            if [[ -f "$file" ]] && grep -qE '^(<<<<<<<|>>>>>>>)( |$)' "$file"; then echo "$file"; fi
        done
    )"
    if [[ -n "$conflicts" ]]; then
        echo "update: the template conflicts with your changes in these files:" >&2
        printf '%s\n' "$conflicts" | sort -u | sed 's/^/  /' >&2
        echo "Resolve each conflict, mark its file with git add, and delete each .rej file." >&2
        echo "Then finish the update with these three commands:" >&2
        echo "  go get github.com/pascalandy/nicely@$version" >&2
        echo "  go mod tidy" >&2
        echo "  just check" >&2
        exit 1
    fi
    go get "github.com/pascalandy/nicely@$version"
    go mod tidy
    just check

# Print the folder of the contract at the Nicely version that go.mod pins, to read its rules offline
contract:
    #!/usr/bin/env bash
    set -Eeuo pipefail
    # go list gives the folder of a module only once the module cache holds it.
    go mod download github.com/pascalandy/nicely
    dir="$(go list -m -f '{{{{.Dir}}' github.com/pascalandy/nicely)/docs/north-star/contract"
    [[ -d "$dir" ]] || { echo "contract: $dir is missing from the Nicely version that go.mod pins" >&2; exit 1; }
    echo "$dir"
