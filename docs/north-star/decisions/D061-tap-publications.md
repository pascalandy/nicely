# D061 Install complete tap publications without a local build

Decided 2026-10-16. Pascal chose precompiled Go extensions for each supported system. Each archive carries its executable, manifest, catalogs, and skill together.
Only the official tap ships these archives in M12. Other taps carry executables that run as published, such as scripts.
The official tap stays in the Nicely repository, as [D041](D041-first-party-extensions.md) decides.
It follows the installed `ncly` release tag and needs no clone. Git taps use bare clones and `git archive` publications.

A tap ID includes its host and maps to local storage through its full SHA-256 hash. Config edits use `tomledit`.
Each tap selects one immutable publication through an atomic link replacement. Sync keeps the selected publication and the previous one.
Calls select that tree once, before they read a manifest, and hold no reader lock.
A publication folder belongs to Nicely as a whole, so for taps the record of its tap ID and revision replaces D035's per-file tracking. Sync and remove move old folders to the trash, so an edit stays recoverable.
Repeatable steps derive their work from disk state, as [ncly tap](../core-spec.md#ncly-tap) defines.
Grants keep the unresolved path through `current`, under [D058](D058-grants-by-path.md)'s existing path rule.

**Why.** [`gh extension install`](https://cli.github.com/manual/gh_extension_install) installs precompiled release artifacts or a repository's executable script. Nicely follows that distribution model. Its users need no compiler.
Keeping the official release aligned with core avoids an independent upgrade that core may reject. The manifest already checks the public contract.
[`tomledit`](https://github.com/creachadair/tomledit) edits TOML structure while retaining comments and whitespace.
Immutable publications extend [D035](D035-sync-ownership.md)'s complete-destination rule to the interval between manifest loading and process launch.

**Rejected.** Building on the user's machine. Installing first-party extensions through Homebrew. A second compatibility version beside `contract_version`. Replacing the executable, manifest, catalogs, and skill separately. Holding a reader lock that blocks sync throughout an extension's execution. Treating a matching domain or path alone as proof of tap ownership.
