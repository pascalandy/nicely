# Discovery

Core finds executables named `ncly-<slug>`, whose identity follows [Extension identity](extension-identity.md).
It searches the data folder's `extensions/` directory, then each absolute, nonempty `PATH` entry in order.
Files must be executable, or links to executable files.
Core never resolves a link when it locates sidecars.
The manifest, catalogs, and skill sit beside the discovered path.
Core skips invalid slugs and sidecar names, such as `ncly-hello.toml`.
A domain reserved by [core-spec.md](../core-spec.md#command-tree) is unavailable to every external extension, including named extensions.

## Selection and conflicts

The first executable found for each slug wins.
A later executable with the same slug stays hidden, even if the winner has an invalid manifest.
Different named extensions of one domain are selected independently.

A whole-domain extension keeps the domain and hides every named extension of that domain, regardless of search order.
An invalid whole-domain winner still hides those named extensions.
Core commands and bundled extensions retain their reserved domains.
Doctor reports every hidden executable with its winning source in `hidden_by`, under [ncly doctor](../core-spec.md#ncly-doctor).
Hidden extensions contribute no commands, skills, prerequisites, grants, or config keys to the installed inventory.
Core never starts a hidden extension.

## Reading declarations

For a call beneath an extension root, core reads only that selected extension's manifest and catalogs.
Domain help, root help, completion, and describe read the declarations needed for their listings.
[Usage](usage.md) defines root help sections.
Core never starts an extension to learn its commands.
[Extension identity](extension-identity.md#shared-domain-group) defines the group that core supplies for named extensions.

A missing, unreadable, or invalid manifest or `en` catalog leaves the executable selected but unavailable.
A call or describe lookup at or beneath that extension root answers `EXTENSION_INVALID`, with exit 78, without starting it.
Help, completion, and describe listings omit its declarations.
A broken named extension never blocks a valid sibling's commands or help.
A domain group lists only valid roots.
With no valid roots, the domain group's help and describe lookup answer `EXTENSION_INVALID`.
A broken whole-domain extension retains its existing domain-wide failure behavior.
