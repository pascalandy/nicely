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
Every program it starts gets an explicit environment, with every `GIT_*` variable of the caller removed.
This includes Git started by Copier, so a hook or `git rebase --exec` cannot redirect the update to its repository.
Other caller settings remain available to the update programs.
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

## Adopt just finish

An extension rendered before `just finish`, such as one of template v0.23.4, adopts the loop through its update (#427).
After the update, `just signoff` and so `just release` need a build that `just finish` accepts.

1. The update brings new `Fill:` lines and new parts into the forms, such as `## Case matrix` of `spec.md`. Keep them on a conflict.
2. Commit the update on its own. That commit adds `intake.toml`, with the example facts of the template.
3. The first build of the commands that exist starts after the update commit. Its intake reads the current `spec.md`, and the wrapped programs at the revisions that `vision.md` names.
4. The spec keeps each line that a test claims, and opens each line that no test proves yet. The first scenario claims an open line, and it fails first.

`just finish` refuses an Intake commit that is not after the newest commit that added `intake.toml`, as [Finish gate](finish-gate.md) says.
So the example facts never stand in for an intake, in a new extension or in an updated one.
A Markdown `readiness.md` stays history, as [Build records](build-records.md#markdown-records) says.
The `nicely-extension` skill of the template teaches these steps.
