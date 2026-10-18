# Command descriptions

One declaration per command supplies help, completion, validation, and discovery. M00 implements the fields that its commands use. Each later command adds input constraints, result types, prerequisites, and configuration sources as it needs them. An extension's manifest holds the same declarations, so help, completion, validation, and discovery treat every command alike. Domain code owns the behavior. The declaration is not a workflow language.

Effects distinguish user writes, network access, and paid requests. Modes distinguish dry run, recorded execution, and resume. A missing declaration is not permission to assume a capability. A command may support recorded execution without supporting resume.

[ncly describe](../core-spec.md#ncly-describe) lists the fields of a declaration. Discovery reads declarations and static extension manifests. It starts no harness or extension and reads no key. Detailed discovery is scoped to one command, and summary listings omit full schemas and skill bodies. Resolved configuration identifies the source file or environment variable for non-secret values. Secret values never appear.
