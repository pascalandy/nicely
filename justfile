set shell := ["bash", "-euo", "pipefail", "-c"]

golangci := "go tool -modfile=tools/go.mod golangci-lint"

# List the recipes
default:
    @just --list

# Run every check that signoff requires
[group('checks')]
check: fmt-check lint tidy-check test gitleaks-rules gitleaks

# Run the Go tests and the testscript scenarios
[group('checks')]
test *args:
    go test ./... {{ args }}

# Lint the Go code
[group('checks')]
lint:
    {{ golangci }} run ./...

# Fail when a Go file needs formatting
[group('checks')]
fmt-check:
    {{ golangci }} fmt --diff

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
