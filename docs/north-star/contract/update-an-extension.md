# Update an extension

This part says how `just update` carries an [extension repository](extension-repository.md) to another template version.

Each template recipe has one command line.
The SDK program owns `check`, `finish`, `contract`, and `update`, as well as install and uninstall.
The update recipe downloads the target Nicely module, then runs its `sdk/extensionrepo/cmd/extensionrepo` program with `go -C`.
This uses the target's updater and dependency replacements, even when the repository pins an older SDK.
`go run package@version` cannot run a module that has `replace` directives.

An older recipe cannot change during its own run.
For a repository rendered before this recipe, the bootstrap runs the target program directly.
Replace VERSION with the target tag, such as `v0.24.0`, from the extension repository:

```sh
go -C "$(go mod download -json github.com/pascalandy/nicely@VERSION | sed -n 's/^[[:space:]]*"Dir": "\(.*\)",$/\1/p')" run ./sdk/extensionrepo/cmd/extensionrepo update VERSION "$PWD"
```

The updater accepts only `vX.Y.Z` tags and an absolute repository path.
It first requires a committed, clean Git tree.
Copier keeps the author's `go.mod` and applies the target template.
The updater checks for conflict markers, `.rej` files, and unmerged index entries before changing dependencies.
On a conflict, it prints each file and the bootstrap command with `--continue` appended.

Resolve each conflict and mark its file with `git add`.
Remove each `.rej` file, then run the printed command.
Continuation requires the target tag in `.copier-answers.yml` and no remaining conflicts.
It skips Copier, moves the Nicely pin, reconciles the previous and target module replacements, and runs `go mod tidy`.
Both update paths finish with `just check`.
The updater preserves unrelated dependencies and replacements.
It never resets, stashes, or commits.
If Copier fails, its output gives restoration commands and the bootstrap command to retry.
