# Extension identity

This part defines an extension's identity and command ownership, as [D064](../decisions/D064-shared-domains.md) decides.

## Identity and root

A manifest declares `domain` and an optional `name`.
Each part matches `[a-z][a-z0-9]*`.
An absent `name` identifies a whole-domain extension.
A present `name` is nonempty and names one job in that domain.
The domain names a family of work, such as `video`.
Neither part names the wrapped tool.
Pascal's extensions use the same rule as other authors' extensions.
No extension is the main extension of a domain.

The extension slug is `domain` when `name` is absent, or `<domain>-<name>` when present.
The executable name after `ncly-` must equal the slug declared by the manifest.
The extension root is the command path `domain` or `<domain> <name>`, respectively.
The two-part limit applies to identity, not to command paths beneath the root.
Thus `ncly-video-color-grade` is invalid, while an existing path such as `transcript run youtube` remains valid.

For example, `domain = "video"` and `name = "convert"` identify `video-convert`.
Its executable is `ncly-video-convert`.
Its root command can take an argument, as `ncly video convert clip.mov` does.
A later `video-subtitle` extension owns `ncly video subtitle ...` independently.

An extension declares its root and only commands beneath that root.
A named extension never declares the domain root.
Every parent beneath its extension root is also declared.
An extension adds commands and never replaces a command, under [principle 13](../principles.md) and [D044](../decisions/D044-additive-extensions.md).

## Identity uses

The slug identifies the extension wherever the domain previously identified one extension.
Command paths still use separate words.
Existing whole-domain identities and stored data remain unchanged.

| Use | Form for `video-convert` |
|---|---|
| Executable and its Go entry directory | `ncly-video-convert`, `cmd/ncly-video-convert/` |
| Exported sidecars | `ncly-video-convert.toml`, `ncly-video-convert.locales/`, `ncly-video-convert.skill/` |
| Skill frontmatter and lookup | `name: video-convert`, `ncly skill view video-convert` |
| Catalog IDs | `video-convert.` prefix, such as `video-convert.summary` |
| Own errors and warnings | Uppercase slug with hyphens replaced by underscores, then `_`, such as `VIDEO_CONVERT_FAILED` |
| Config table and declared keys | `[video-convert]`, `video-convert.preset` |
| Status command | Extension root followed by `status`, such as `ncly video convert status` |
| Doctor component and check IDs | `video-convert`, `video-convert.<check>` |
| Doctor discovery check | `extensions.video-convert`, with the existing JSON key `domain` holding `video-convert` |
| Auth grant and revoke argument | One slug word, as `ncly auth grant <service> video-convert` |
| Grant storage and auth JSON | Existing `domain` key holds `video-convert`, including status, answers, and dry-run plans |

The legacy JSON key `domain` remains an extension identity field in doctor and auth.
It does not become a family selector.
A grant still binds the service, extension identity, and discovered source, as [Grants](grants.md) defines.
A grant for `video-convert` gives `video-subtitle` no key.
An uninstalled config table may use any valid slug, under [Configuration](configuration.md).

The optional manifest `name` preserves `format_version = 1` and `contract_version = 1`.
An older host skips a two-part executable name under its existing discovery rule.
Named extensions require the Nicely release that implements this part.

## Shared domain group

Core supplies the domain group for selected named extensions, even when only one exists.
The group lists their valid roots in alphabetical order in help, completion, and describe.
Core supplies its translated summary with the domain word, using the existing group format.
The group has no handler, extension skill, requirements, or supported execution modes.
Its describe entry has `source: "core"` and no `skill` key.
An extension root keeps its own summary and source.
That root's describe entry carries `skill`, under [ncly describe](../core-spec.md#ncly-describe).
The bundled `skill` extension identifies that skill by the extension slug.

`ncly video` and `ncly video --help` print the domain group's help without starting any extension.
`ncly describe video` returns the group and the selected extensions' commands.
`ncly doctor video-convert` selects one extension, not the whole family.
The same slug rule applies to auth and doctor completion.

Different roots never overlap, so core attaches each extension's declared subtree beneath the group.
Core needs no merge of declarations at individual command paths.
[Discovery](discovery.md) defines selection and conflicts.
[Dispatch](dispatch.md) defines execution.
