set shell := ["bash", "-euo", "pipefail", "-c"]

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

# Run every check that signoff requires
[group('checks')]
check: fmt-check lint tidy-check test release-lint gitleaks-rules gitleaks

# Run just check on each layer of a stack, bottom to top, as origin holds it
[group('checks')]
[positional-arguments]
check-layers *branches:
    #!/usr/bin/env bash
    set -Eeuo pipefail
    # A hook or a caller can export GIT_DIR, which would point git -C at the caller's repository.
    while IFS= read -r git_variable; do unset "${git_variable}"; done < <(git rev-parse --local-env-vars)
    die() {
        echo "check-layers: $*" >&2
        exit 1
    }
    branches=("$@")
    if (($# == 0)); then
        stack="$(gh stack view --json)" || die "run it on a branch of the stack, or pass the branches bottom to top"
        while IFS= read -r branch; do branches+=("$branch"); done < <(jq -r '.branches[] | select(.isMerged | not) | .name' <<<"$stack")
    fi
    ((${#branches[@]} > 0)) || die "no layer to check"
    # One refspec per layer: the configured fetch may skip a layer or keep a deleted one, and a missing branch fails here.
    refspecs=()
    for branch in "${branches[@]}"; do refspecs+=("+refs/heads/$branch:refs/remotes/origin/$branch"); done
    git fetch --quiet origin "${refspecs[@]}" || die "git fetch of the layers from origin failed"
    # Not /tmp: Linux mounts it as tmpfs and macOS purges it, and the logs must outlive the run.
    run="${XDG_CACHE_HOME:-$HOME/.cache}/nicely/check-layers/$(date +%Y%m%d-%H%M%S)-$$"
    worktree="$run/worktree"
    mkdir -p "$run"
    trap 'git worktree remove --force "$worktree" 2>/dev/null || true' EXIT
    # Ctrl-C and SIGTERM end the run after the current command, then the EXIT trap removes the worktree.
    # Without these traps, Ctrl-C moves on to the next layer, and SIGTERM can leave the worktree behind.
    trap 'exit 130' INT
    trap 'exit 143' TERM
    git worktree add --quiet --detach "$worktree" "refs/remotes/origin/${branches[0]}"
    echo "check-layers: logs in $run" >&2
    status=0
    for i in "${!branches[@]}"; do
        branch="${branches[$i]}"
        sha="$(git rev-parse "refs/remotes/origin/$branch")"
        log="$run/$(printf '%02d' $((i + 1)))-${branch//\//-}.log"
        git -C "$worktree" checkout --quiet --force --detach "$sha"
        git -C "$worktree" clean -ffdxq
        verdict=ok
        (cd "$worktree" && just check) >"$log" 2>&1 </dev/null || verdict=FAIL
        if [[ "$(git -C "$worktree" rev-parse HEAD)" != "$sha" ]]; then
            echo "check-layers: HEAD moved off origin/$branch during the check" >>"$log"
            verdict=FAIL
        fi
        [[ "$verdict" == ok ]] || status=1
        echo "$branch $(git rev-parse --short "$sha") $verdict"
    done
    exit "$status"

# Run the Go tests, the testscript scenarios, and the release script checks
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
    bash scripts/test-hook-env.sh

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

# Check the release scripts and the dormant manual workflow
[group('checks')]
release-lint:
    shellcheck scripts/*.sh
    shfmt -d scripts/*.sh
    {{ go-tools }} actionlint

# Build and inspect a local release without publishing
[group('checks')]
release-check version:
    bash scripts/release-check.sh {{ quote(version) }}

# Fail when go.mod or tools/go.mod needs tidying
[group('checks')]
tidy-check:
    go mod tidy -diff
    cd tools && go mod tidy -diff

# Scan the whole history for keys, home folders, and private addresses
[group('checks')]
gitleaks:
    gitleaks git --no-banner --redact --log-level warn

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
signoff: check
    gh signoff
