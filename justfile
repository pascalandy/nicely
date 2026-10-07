set shell := ["bash", "-euo", "pipefail", "-c"]

go-tools := 'GOTOOLCHAIN="$(go -C tools env GOVERSION)" go tool -modfile=tools/go.mod'
golangci := go-tools + " golangci-lint"

# List the recipes
default:
    @just --list

# Run every check that signoff requires
[group('checks')]
check: fmt-check lint tidy-check test release-lint gitleaks-rules gitleaks

# Run the Go tests, the testscript scenarios, and the release script checks
[group('checks')]
test *args:
    #!/usr/bin/env bash
    set -Eeuo pipefail
    # A hook or a caller can export GIT_DIR. Fixture commits must see a normal repository.
    while IFS= read -r git_variable; do unset "${git_variable}"; done < <(git rev-parse --local-env-vars)
    go test ./... {{ args }}
    (cd tools && go test ./cmd/... {{ args }})
    bash scripts/release-check-test.sh
    bash scripts/test-hook-env.sh

# Lint the Go code
[group('checks')]
lint:
    {{ golangci }} run ./...
    cd tools && go tool golangci-lint run ./cmd/...

# Fail when a Go file needs formatting
[group('checks')]
fmt-check:
    {{ golangci }} fmt --diff
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
