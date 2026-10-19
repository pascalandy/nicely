set shell := ["bash", "-euo", "pipefail", "-c"]

# A shebang recipe writes its script here. The Codex sandbox makes just's default, $XDG_RUNTIME_DIR/just, read-only.
# A nested just reads this setting again, but not the --tempdir of its caller. JUST_TEMPDIR and --tempdir still win. See #79.
set tempdir := "/tmp"

go-tools := 'GOTOOLCHAIN="$(go -C tools env GOVERSION)" go tool -modfile=tools/go.mod'
golangci := go-tools + " golangci-lint"

# List the recipes
default:
    @just --list

# Print the next card of the plan, or the step that unblocks it
[group('plan')]
next:
    @go -C tools run ./cmd/plan next "{{ justfile_directory() }}"

# Show every milestone and the cards of the current one
[group('plan')]
status:
    @go -C tools run ./cmd/plan status "{{ justfile_directory() }}"

# Build ncly into ~/.local/bin and the first-party extensions into Nicely's data folder
install:
    #!/usr/bin/env bash
    set -Eeuo pipefail
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

# Run every check that signoff requires
[group('checks')]
check: fmt-check lint tidy-check test race python release-lint gitleaks-rules gitleaks (limits "--base" "origin/main")

# Run just check on each layer of a stack, bottom to top, as origin holds it
[group('checks')]
[positional-arguments]
check-layers *branches:
    @exec bash scripts/check-layers.sh "$@"

# Build ncly into a scratch folder for the verify-nicely skill, or run one of its steps
[group('checks')]
[positional-arguments]
verify *args:
    @bash .agents/skills/verify-nicely/scripts/verify.sh "$@"

# Run testscript scenarios by exact name, each N times: just scenario <name>|<group>/<name>... [--count N]
[group('checks')]
[positional-arguments]
scenario *args:
    @bash scripts/scenario.sh "$@"

# Run the Go tests, the testscript scenarios, and the tests of the scripts
[group('checks')]
test *args:
    #!/usr/bin/env bash
    set -Eeuo pipefail
    # A hook or a caller can export GIT_DIR. Fixture commits must see a normal repository.
    while IFS= read -r git_variable; do unset "${git_variable}"; done < <(git rev-parse --local-env-vars)
    # ./... skips testdata/, so the Go fixture extensions are named.
    go test ./... ./testdata/extensions/... {{ args }}
    # Go's test cache never rechecks a file outside the tools module, such as the docs or the release workflow, so these tests run uncached.
    (cd tools && go test -count=1 ./cmd/... {{ args }})
    bash scripts/release-check-test.sh
    bash scripts/release-verify-test.sh
    bash scripts/test-hook-env.sh
    bash scripts/check-layers-test.sh
    bash scripts/signoff-test.sh

# Run the Go tests of sdk and internal under the race detector, without the scenarios
[group('checks')]
race:
    go test -race ./sdk/... ./internal/...

# Check the Python engine of transcript with ruff, pyright, and pytest through uv
[group('checks')]
python:
    #!/usr/bin/env bash
    set -Eeuo pipefail
    # pyright resolves a sibling module only from the engine folder. Its pyrightconfig.json makes
    # that folder the root of pyright, which otherwise takes the first parent with a pyproject.toml,
    # such as a home folder, and checks all of it.
    cd extensions/transcript/python
    uv tool run ruff@0.16.9 check --isolated
    uv tool run ruff@0.16.9 format --isolated --check
    requirements="$(mktemp)"
    trap 'rm -f "$requirements"' EXIT
    uv export --quiet --no-hashes --script transcript.py > "$requirements"
    # The header pins the runtime packages; this recipe pins only the tools. pyright reads the tests, so it needs pytest.
    uv tool run --with-requirements "$requirements" --with pytest==9.1.1 pyright@1.1.414
    uv tool run --with-requirements "$requirements" --with pytest-xdist==3.8.0 pytest@9.1.1 -n auto

# Lint the Go code
[group('checks')]
lint:
    {{ golangci }} run ./... ./testdata/extensions/...
    cd tools && go tool golangci-lint run ./cmd/...

# Fail when a Go file needs formatting
[group('checks')]
fmt-check:
    {{ golangci }} fmt --diff ./... ./testdata/extensions/...
    cd tools && go tool golangci-lint fmt --diff

# Check the shell scripts, those of the skills too, and the dormant manual workflow
[group('checks')]
release-lint:
    shellcheck scripts/*.sh .agents/skills/*/scripts/*.sh
    shfmt -d scripts/*.sh .agents/skills/*/scripts/*.sh
    {{ go-tools }} actionlint

# Build and inspect a local release without publishing
[group('checks')]
release-check version:
    bash scripts/release-check.sh {{ quote(version) }}

# Write one version's GitHub release notes from the previous tag to HEAD
[group('release')]
[positional-arguments]
changelog version previous="":
    #!/usr/bin/env bash
    set -Eeuo pipefail
    version="${1}"
    [[ "${version}" =~ ^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$ ]] || { printf 'Usage: just changelog vX.Y.Z [previous-tag]\n' >&2; exit 2; }
    previous="${2:-}"
    if [[ -z "${previous}" ]]; then
        previous="$(git describe --tags --abbrev=0 --match 'v[0-9]*' --exclude "${version}" HEAD)"
    fi
    git rev-parse --verify "refs/tags/${previous}^{commit}" >/dev/null
    git merge-base --is-ancestor "refs/tags/${previous}" HEAD
    mkdir -p .release
    uv tool run git-cliff@2.14.2 --config cliff.toml --tag "${version}" --output .release/changelog.md "${previous}..HEAD"
    printf 'Changelog: .release/changelog.md (%s..HEAD)\n' "${previous}"

# Check a published release and the formula of the Homebrew tap, without changing either
[group('checks')]
release-verify version:
    bash scripts/release-verify.sh {{ quote(version) }}

# Fail when go.mod or tools/go.mod needs tidying
[group('checks')]
tidy-check:
    go mod tidy -diff
    cd tools && go mod tidy -diff

# Scan the whole history for keys, home folders, and private addresses
[group('checks')]
gitleaks:
    gitleaks git --no-banner --redact --log-level warn

# Refuse a change that grows a file past its size or adds a sentence over 25 words
[group('checks')]
limits *args:
    @go -C tools run ./cmd/plan limits "{{ justfile_directory() }}" {{ args }}

# Scan the staged changes, as the pre-commit hook does
[group('checks')]
gitleaks-staged:
    gitleaks git --staged --no-banner --redact --log-level warn --verbose --no-color

# Prove the leak rules: every line of leaks.txt is blocked, every line of clean.txt passes
[group('checks')]
gitleaks-rules:
    #!/usr/bin/env bash
    set -euo pipefail
    scan() { printf '%s\n' "$1" | gitleaks stdin --config .gitleaks.toml --no-banner --log-level error >/dev/null; }
    status=0
    while IFS= read -r line; do
        if scan "$line"; then echo "not blocked: $line" >&2; status=1; fi
    done < testdata/gitleaks/leaks.txt
    while IFS= read -r line; do
        if ! scan "$line"; then echo "blocked: $line" >&2; status=1; fi
    done < testdata/gitleaks/clean.txt
    exit "$status"

# Format the Go files in place, or only the given files
fmt *files:
    {{ golangci }} fmt {{ files }}

# Run the checks, then mark the pushed HEAD green on GitHub
[group('checks')]
signoff:
    @exec bash scripts/signoff.sh

# Sign off each layer of a stack after its checks pass, up to the first failure
[group('checks')]
[positional-arguments]
signoff-layers *branches:
    @exec bash scripts/check-layers.sh --signoff "$@"
