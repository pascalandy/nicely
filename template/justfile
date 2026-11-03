set shell := ["bash", "-euo", "pipefail", "-c"]

# The Copier used to create the repository, also pinned by the target SDK updater
copier := "uv tool run copier@9.19.1"

# List the recipes
default:
    @just --list

# Run every repository check through the pinned SDK
check:
    go run github.com/pascalandy/nicely/sdk/extensionrepo/cmd/extensionrepo check

# Fail until the loop is done: no open line, accepted limits, and a passing readiness verdict for this commit
finish:
    go run github.com/pascalandy/nicely/sdk/extensionrepo/cmd/extensionrepo finish

# Write ## Intake of vision.md from intake.toml, and readiness.md from review.toml
records:
    go run github.com/pascalandy/nicely/sdk/extensionrepo/cmd/extensionrepo records

# Check the TLA+ models of verification/tlc.json with the pinned TLC, then write the record that just check reads. A local jar needs the pinned SHA-256 too
verify-models jar="":
    go run github.com/pascalandy/nicely/sdk/extensionrepo/cmd/extensionrepo verify-models {{ if jar == "" { "" } else { quote(join(invocation_directory(), jar)) } }}

# Run one scenario of testdata/script/ by its name, such as just scenario inspect
scenario name:
    go run github.com/pascalandy/nicely/sdk/extensionrepo/cmd/extensionrepo scenario {{ quote(name) }}

# Check the committed SHA in a detached scratch worktree, then sign that exact SHA
signoff:
    go run ./internal/export signoff

# Release the signed, pushed SHA and prove the published README install in scratch folders
release version fixture="":
    go run ./internal/export release {{ quote(version) }} {{ if fixture == "" { "" } else { quote(join(invocation_directory(), fixture)) } }}

# Build the extension and install its sidecars before its executable
install:
    go run ./internal/export install

# Remove the installed extension, with its executable first
uninstall:
    go run ./internal/export uninstall

# Run the command of ## Acceptance on a copy of the fixture or on a made input, in an owned run
acceptance fixture="":
    go run ./internal/export acceptance {{ if fixture == "" { "" } else { quote(join(invocation_directory(), fixture)) } }}

# Run the target version's updater, with the template before the SDK, on a clean Git tree
update version:
    go -C "$(go mod download -json github.com/pascalandy/nicely@{{ quote(version) }} | sed -n 's/^[[:space:]]*"Dir": "\(.*\)",$/\1/p')" run ./sdk/extensionrepo/cmd/extensionrepo update {{ quote(version) }} {{ quote(justfile_directory()) }}

# Print the folder of the pinned contract
contract:
    go run github.com/pascalandy/nicely/sdk/extensionrepo/cmd/extensionrepo contract
