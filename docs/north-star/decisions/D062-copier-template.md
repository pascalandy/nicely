# D062 Ship the extension template in the Nicely repository, tied to the SDK by one version

Decided 2026-10-20 in [#273](https://github.com/pascalandy/nicely/issues/273). The extension template is a Copier template in the Nicely repository. A `copier.yml` at the root points at its folder, so each Nicely tag is also a template version. An extension repository pins one Nicely version in its `go.mod`. That version selects the SDK, the contract that `just contract` prints, and the `ncly` that its tests run. Nicely's `just check` renders the template and runs the checks of the rendered repository. [Extension repository](../contract/extension-repository.md) owns the layout and the recipes.

**Scope update, 2026-10-24.** [#340](https://github.com/pascalandy/nicely/issues/340) found no release steps in a render. The template now writes a `README.md` with the install commands, and the release steps in `AGENTS.md`. A release is a tag `vX.Y.Z`, the form of Nicely's tags and of Go module versions. The install keeps `git clone --branch <tag>`, and names the warning that Git 2.56 prints for an annotated tag. No clone option removes it, and a clone followed by a checkout would add a second command.

**Scope update, 2026-10-29.** [#373](https://github.com/pascalandy/nicely/issues/373) found that a push can skip the pre-push hook of Lefthook, so nothing protected `main`. The template now ships a GitHub workflow that runs `just check` for each push to `main` and each pull request. A public repository runs it on standard GitHub runners for free.

**Why.** Copier records the template version and updates the copied files, so `just update` carries an author's repository forward. Coupled breaking changes to the SDK and template land together, for the reason of [D041](D041-first-party-extensions.md). Compatible additions may land in stages. One version number leaves no pair of versions to mismatch.

**Rejected.** A template embedded in `ncly`, which adds a core command. A GitHub template repository, whose copy starts an unrelated history that cannot take updates. A template in its own repository, which needs two pull requests for each SDK change.
