set shell := ["bash", "-euo", "pipefail", "-c"]

# A shebang recipe writes its script here. The Codex sandbox makes just's default, $XDG_RUNTIME_DIR/just, read-only.
# A nested just reads this setting again, but not the --tempdir of its caller. JUST_TEMPDIR and --tempdir still win. See #79.
set tempdir := "/tmp"

go-tools := 'GOTOOLCHAIN="$(go -C "' + justfile_directory() + '/tools" env GOVERSION)" go tool -modfile="' + justfile_directory() + '/tools/go.mod"'
golangci := go-tools + " golangci-lint"

# Each fixture extension repository of the test kit is a Go module of its own, which ./... skips.
standalone := "testdata/extensions/standalone testdata/extensions/standalone-named"

# List the recipes
default:
    @just --list

# Print the next card of the plan, or the step that unblocks it
[group('plan')]
next:
    @go -C tools run ./cmd/plan next "{{ justfile_directory() }}"

# Show the XT and COR tables from frontmatter
[group('plan')]
status:
    @go -C tools run ./cmd/plan status "{{ justfile_directory() }}"

# Build ncly into $XDG_BIN_HOME or ~/.local/bin and the first-party extensions into Nicely's data folder, or only the given names, such as ncly agent
[positional-arguments]
install *names:
    @exec bash scripts/install.sh "$@"

# Run every check that signoff requires, up to the first failure, or each one with --keep-going
[group('checks')]
[positional-arguments]
check *flags:
    @exec bash scripts/check.sh "$@"

# Run just check on each layer of a stack, bottom to top, as origin holds it
[group('checks')]
[positional-arguments]
check-layers *branches:
    @exec bash scripts/check-layers.sh "$@"

# Check this host for a UAT, and the given fixture folders, without changing anything
[group('checks')]
[positional-arguments]
preflight *fixtures:
    @exec bash scripts/preflight.sh "$@"

# Build ncly into a scratch folder for the verify-nicely skill, or run one of its steps
[group('checks')]
[positional-arguments]
[no-exit-message]
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
    # Its test kit builds ncly, whose source the test cache of the fixture never rechecks.
    for dir in {{ standalone }}; do go -C "$dir" test -count=1 ./... {{ args }}; done
    # Go's test cache never rechecks a file outside the tools module, such as the docs or the release workflow, so these tests run uncached.
    (cd tools && go test -count=1 ./cmd/... {{ args }})
    bash scripts/release-check-test.sh
    bash scripts/release-verify-test.sh
    bash scripts/test-hook-env.sh
    bash scripts/check-layers-test.sh
    bash scripts/signoff-test.sh
    bash scripts/check-test.sh
    bash scripts/preflight-test.sh
    bash scripts/scenario-test.sh
    bash scripts/install-test.sh

# Render the extension template and check the rendered repository
[group('checks')]
template:
    bash scripts/template-test.sh

# Run the Go tests of sdk and internal under the race detector, without the scenarios
[group('checks')]
race:
    go test -race ./sdk/... ./internal/...

# Check the Python of transcript and of the extension template with ruff, and transcript with pyright and pytest, through uv
[group('checks')]
python:
    #!/usr/bin/env bash
    set -Eeuo pipefail
    # pyright resolves a sibling module only from the engine folder. Its pyrightconfig.json makes
    # that folder the root of pyright, which otherwise takes the first parent with a pyproject.toml,
    # such as a home folder, and checks all of it.
    cd extensions/transcript/python
    uv tool run ruff@0.16.9 check --isolated . ../../../template/worker
    uv tool run ruff@0.16.9 format --isolated --check . ../../../template/worker
    requirements="$(mktemp)"
    trap 'rm -f "$requirements"' EXIT
    uv export --quiet --no-hashes --script transcript.py > "$requirements"
    # The header pins the runtime packages; this recipe pins only the tools. pyright reads the tests, so it needs pytest.
    # The nodejs extra brings its own Node.js. Without Node on PATH, the nodeenv fallback fails on the alpha versions of Node 27.
    uv tool run --with-requirements "$requirements" --with pytest==9.1.1 'pyright[nodejs]@1.1.414'
    uv tool run --with-requirements "$requirements" --with pytest-xdist==3.8.0 pytest@9.1.1 -n auto

# Lint the Go code
[group('checks')]
lint:
    {{ golangci }} run ./... ./testdata/extensions/...
    for dir in {{ standalone }}; do (cd "$dir" && {{ golangci }} run ./...); done
    cd tools && go tool golangci-lint run ./cmd/...

# Fail when a Go file needs formatting
[group('checks')]
fmt-check:
    {{ golangci }} fmt --diff ./... ./testdata/extensions/...
    for dir in {{ standalone }}; do (cd "$dir" && {{ golangci }} fmt --diff); done
    cd tools && go tool golangci-lint fmt --diff

# Check the shell scripts, those of the skills too, and the GitHub workflows, the template's too
[group('checks')]
release-lint:
    shellcheck scripts/*.sh .agents/skills/*/scripts/*.sh
    shfmt -d scripts/*.sh .agents/skills/*/scripts/*.sh
    {{ go-tools }} actionlint
    {{ go-tools }} actionlint template/.github/workflows/check.yml

# Build and inspect a local release without publishing
[group('checks')]
release-check version:
    bash scripts/release-check.sh {{ quote(version) }}

# Write one version's GitHub release notes from the previous tag to HEAD
[group('release')]
[positional-arguments]
changelog version previous="":
    @exec bash scripts/changelog.sh "$@"

# Check a published release and the formula of the Homebrew tap, without changing either
[group('checks')]
release-verify version:
    bash scripts/release-verify.sh {{ quote(version) }}

# Fail when go.mod or tools/go.mod needs tidying
[group('checks')]
tidy-check:
    go mod tidy -diff
    for dir in {{ standalone }}; do go -C "$dir" mod tidy -diff; done
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
    @exec bash scripts/gitleaks-rules.sh

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
