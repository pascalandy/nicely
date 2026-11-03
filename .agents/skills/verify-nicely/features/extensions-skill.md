# extensions-skill

`extensions-skill` is the `nicely-extensions` skill that `ncly` bundles. An agent reads it to create an extension repository from the template, to install an extension at a release tag, and to remove one. [template.md](template.md) drives the repository that its command renders.

## Sub-features

- `template-command` gives the command that creates an extension repository, in the `nicely-extensions` skill, with the VERSION of a development build
- `template-install-command` gives the commands that install an extension repository at a release tag, in the `nicely-extensions` skill
- `template-remove-command` gives the steps that remove an extension, in the `nicely-extensions` skill

## How to get to it (user POV)

- `ncly skill view nicely`, which names `ncly skill view nicely-extensions`
- `ncly skill view nicely-extensions`

## Driving it with verify.sh

Preconditions:

- A new run from `just verify`, whose Doctor step passed

- **Template command.** Run `just verify step "$run" extensions-skill skill view nicely`. Exit 0, and it names `ncly skill view nicely-extensions`. Then run `just verify step "$run" extensions-skill skill view nicely-extensions`. Exit 0. The skill holds the section `## Build an extension`, and in its code block the line `uv tool run copier@<version> copy --vcs-ref VERSION --defaults --data domain=DOMAIN --data name=NAME --data module=github.com/me/ncly-DOMAIN-NAME gh:pascalandy/nicely ncly-DOMAIN-NAME`. `<version>` is the one of the line `copier :=` in `template/justfile`. The section also holds `For a pseudo-version, use its commit as VERSION`, and a line that starts with ``Then build each command with the `nicely-extension` skill of the new repository``.
- **Install command.** In the output of `skill view nicely-extensions`, the skill holds the section `## Install an extension`. It quotes the install rule, in a line that starts with `> A user installs a fixed version with`. Its code block holds the lines `git clone --branch TAG https://github.com/OWNER/REPO` and `cd REPO && just install`, and the text names `ncly describe ROOT --json`.
- **Remove command.** In the output of `skill view nicely-extensions`, the skill holds the section `## Remove an extension`. It names `ncly doctor extensions --json`, the four files `ncly-SLUG`, `ncly-SLUG.toml`, `ncly-SLUG.locales/`, and `ncly-SLUG.skill/`, `just uninstall`, `NOT_FOUND`, and `GRANT_UNUSED`.

## Gotchas

- The skill comes with the `ncly` of the run, so this feature needs no `prepare`
