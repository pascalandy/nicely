# D062 Ship the extension template in the Nicely repository, tied to the SDK by one version

Decided 2026-10-20 in [#273](https://github.com/pascalandy/nicely/issues/273). The extension template is a Copier template in the Nicely repository. A `copier.yml` at the root points at its folder, so each Nicely tag is also a template version. An extension repository pins one Nicely version in its `go.mod`. That version selects the SDK, the contract that `just contract` prints, and the `ncly` that its tests run. Nicely's `just check` renders the template and runs the checks of the rendered repository. [Extensions](../contract/extensions.md#extension-repository) owns the layout and the recipes.

**Revision, 2026-10-24.** [D064](D064-shared-domains.md) applies this rule to named extensions within a shared domain.

**Why.** Copier records the template version and updates the copied files, so `just update` carries an author's repository forward. One pull request changes the SDK and the template together, for the reason of [D041](D041-first-party-extensions.md). One version number leaves no pair of versions to mismatch.

**Rejected.** A template embedded in `ncly`, which adds a core command. A GitHub template repository, whose copy starts an unrelated history that cannot take updates. A template in its own repository, which needs two pull requests for each SDK change. A GitHub workflow in the template, since `just` and Lefthook run the checks.
