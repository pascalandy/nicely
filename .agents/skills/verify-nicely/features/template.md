# template

`template` is the extension template of the checkout, which an agent renders with the command of the `nicely-extensions` skill. `just verify prepare "$run" template` renders it as the extension `probe`, with the Copier that the template pins, and installs it with `just install` of the rendered repository. A user then finds `probe` in `ncly --help`, runs `ncly probe inspect`, and reads its declaration in `ncly describe`, its skill, and its check in `ncly doctor`.

## Sub-features

- `template-command` gives the command that creates an extension repository, in the `nicely-extensions` skill
- `template-install-command` gives the commands that install an extension repository at a release tag, in the `nicely-extensions` skill
- `template-remove-command` gives the steps that remove an extension, in the `nicely-extensions` skill
- `template-install` renders the template, and installs the rendered repository with `just install` in the scratch data folder
- `template-release` gives the release steps in the `AGENTS.md` of the rendered repository, and the commands that install a release tag in its `README.md`
- `template-discovery` lists `probe` under the extensions of `ncly --help`
- `template-inspect` gives the size and the line count of a file, as JSON and as text, through the worker `inspect.py`, which `python3` runs through `sdk/wrap`
- `template-usage` refuses a missing file with `USAGE_INVALID`, before the worker starts
- `template-refusal` refuses a file with a NUL byte, which the worker accepts, with `PROBE_NOT_TEXT`
- `template-describe` describes `probe inspect`, with its program `python3` and no effect
- `template-skill` serves the skill installed beside `probe`
- `template-doctor` reports the `python3` that `probe inspect` needs

## How to get to it (user POV)

- `ncly skill view nicely`, which names `ncly skill view nicely-extensions`
- The command of the `nicely-extensions` skill, `uv tool run copier@<version> copy ...`, then `just install` in the new repository
- `ncly --help` and `ncly probe --help`
- `ncly probe inspect <file>`
- `ncly describe probe inspect`
- `ncly skill view probe`
- `ncly doctor probe`

## Driving it with verify.sh

Preconditions:

- A new run from `just verify`, whose Doctor step passed
- `uv`, `just`, and a `python3` of 3.9 or later, which answers under the scratch home

- **Template command.** Run `just verify step "$run" template skill view nicely`. Exit 0, and it names `ncly skill view nicely-extensions`. Then run `just verify step "$run" template skill view nicely-extensions`. Exit 0. The skill holds the section `## Build an extension`, and in its code block the line `uv tool run copier@<version> copy --vcs-ref VERSION --defaults --data domain=DOMAIN --data name=NAME --data module=github.com/me/ncly-DOMAIN-NAME gh:pascalandy/nicely ncly-DOMAIN-NAME`. `<version>` is the one of the line `copier :=` in `template/justfile`.
- **Install command.** In the output of `skill view nicely-extensions`, the skill holds the section `## Install an extension`. It quotes the install rule, in a line that starts with `> A user installs a fixed version with`. Its code block holds the lines `git clone --branch TAG https://github.com/OWNER/REPO` and `cd REPO && just install`, and the text names `ncly describe ROOT --json`.
- **Remove command.** In the output of `skill view nicely-extensions`, the skill holds the section `## Remove an extension`. It names `ncly doctor extensions --json`, the four files `ncly-SLUG`, `ncly-SLUG.toml`, `ncly-SLUG.locales/`, and `ncly-SLUG.skill/`, `just uninstall`, `NOT_FOUND`, and `GRANT_UNUSED`.
- **Render and install.** Run `just verify prepare "$run" template`. It exits 0. Its log names `just install` of `probe`, the line `installed <run>/scratch/home/.local/share/nicely/extensions/ncly-probe`, and the `python3` that it puts first in the PATH of the wrapper. Then run `ls "$run/scratch/home/.local/share/nicely/extensions"`: it lists `ncly-probe`, `ncly-probe.locales`, `ncly-probe.skill`, and `ncly-probe.toml`, beside the four files of `hello`. Report the output of `prepare` and of `ls` beside the quote.
- **Release and install text.** Run `grep -n -A12 '^## Release$' "$run/scratch/ncly-probe/AGENTS.md"` and `grep -n -A3 '^git clone' "$run/scratch/ncly-probe/README.md"`. Both exit 0. `AGENTS.md` names the tag form `vX.Y.Z` and the command `git tag v0.1.0`. `README.md` holds the lines `git clone --branch TAG https://example.com/ncly-probe ncly-probe`, `cd ncly-probe`, `just install`, and `ncly describe probe --json`. Report both outputs beside the quote.
- **Discovery.** Run `just verify step "$run" template --help`. Exit 0, and the `EXTENSIONS` section lists `probe:  Inspect files` beside `hello` and `skill`.
- **Help.** Run `just verify step "$run" template probe --help`. Exit 0, and `COMMANDS` lists `inspect:  Give the size and the line count of a file`.
- **Inspect as JSON.** Run `just verify step "$run" template probe inspect README.md --json`. Exit 0, stderr is empty, and stdout is `{"ok":true,"contract_version":1,"lines":<n>,"size_bytes":<bytes>}`. `<n>` is the number that `wc -l < README.md` prints, plus 1 when the last line of `README.md` has no newline. `<bytes>` is the number that `wc -c < README.md` prints. Report the output of both `wc` calls beside the quote.
- **Inspect as text.** Run `just verify step "$run" template probe inspect README.md`. Exit 0, and stdout is `README.md: <n> lines, <size>`. `<size>` is `<bytes>` in decimal units, with at most one decimal above 1,000 bytes, such as `5.4 kB` for 5372 bytes and `8 kB` for 7990 bytes.
- **Missing file.** Run `just verify step "$run" template probe inspect nothere.md`. Exit 2, stdout is empty, and stderr starts with `error[USAGE_INVALID]: nothere.md is no file that probe can inspect.`, with the hint `ncly probe inspect notes.md`.
- **Refusal.** Run `printf '\211PNG\r\n\032\n\000\000' > "$run/image.png"`, then `just verify step "$run" template probe inspect "$run/image.png"`. Exit 2, stdout is empty, and stderr starts with `error[PROBE_NOT_TEXT]: <run>/image.png is no text file, so its lines mean nothing.`, with a hint that names `notes.md`.
- **Describe.** Run `just verify step "$run" template describe probe inspect --json`. Exit 0. The one entry of `commands` has the path `probe inspect` and the source `extension`. Its `effects` are all `false`, and `requires.bins` holds `python3` with the version `3.9`.
- **Skill.** Run `just verify step "$run" template skill view probe`. Exit 0. Its first line names the skill folder `<run>/scratch/home/.local/share/nicely/extensions/ncly-probe.skill`, and the file of the skill `probe` follows, from its frontmatter `name: probe`.
- **Doctor.** Run `just verify step "$run" template doctor probe --json`. Exit 0, and the check `probe.python3` is a `pass` whose message names the version of `python3` and its path under `<run>/scratch/tools`.

## Gotchas

- `prepare` renders the working tree of the checkout, uncommitted changes included, as `evidence/commit.txt` marks with `dirty`. The template of a change has no release tag, so `prepare` tags a scratch copy of `copier.yml` and `template/` as `v0.0.0`
- `prepare` needs the network once, when `uv` holds no Copier of the pinned version yet
- A shim of mise or pyenv finds no Python under the scratch home, so `prepare` links the first `python3` of `PATH` that answers there. When none answers, it stops and names the install command
- `just template`, outside `ncly`, proves the rest of the template: `just check`, `just contract`, and `just update` of the rendered repository. Verify drives what a user runs through `ncly`
