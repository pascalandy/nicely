# template

`template` is the extension template of the checkout, which an agent renders with the command of the `nicely` skill. `just verify prepare "$run" template` renders it as the extension `probe`, with the Copier that the template pins, and installs it with `just install` of the rendered repository. A user then finds `probe` in `ncly --help`, runs `ncly probe inspect`, and reads its declaration in `ncly describe`, its skill, and its check in `ncly doctor`.

## Sub-features

- `template-command` gives the command that creates an extension repository, in the `nicely` skill
- `template-install` renders the template, and installs the rendered repository with `just install` in the scratch data folder
- `template-discovery` lists `probe` under the extensions of `ncly --help`
- `template-inspect` gives the size and the line count of a file, as JSON and as text, through the worker `inspect.py`, which `python3` runs through `sdk/wrap`
- `template-usage` refuses a missing file with `USAGE_INVALID`, before the worker starts
- `template-describe` describes `probe inspect`, with its program `python3` and no effect
- `template-skill` serves the skill installed beside `probe`
- `template-doctor` reports the `python3` that `probe inspect` needs

## How to get to it (user POV)

- `ncly skill view nicely`
- The command of the `nicely` skill, `uv tool run copier@<version> copy ...`, then `just install` in the new repository
- `ncly --help` and `ncly probe --help`
- `ncly probe inspect <file>`
- `ncly describe probe inspect`
- `ncly skill view probe`
- `ncly doctor probe`

## Driving it with verify.sh

Preconditions:

- A new run from `just verify`, whose Doctor step passed
- `uv`, `just`, and a `python3` of 3.9 or later, which answers under the scratch home

- **Template command.** Run `just verify step "$run" template skill view nicely`. Exit 0. The skill holds the section `## Build an extension`, and in its code block the line `uv tool run copier@<version> copy --vcs-ref VERSION --defaults --data domain=DOMAIN --data module=github.com/me/ncly-DOMAIN gh:pascalandy/nicely ncly-DOMAIN`. `<version>` is the one of the line `copier :=` in `template/justfile`.
- **Render and install.** Run `just verify prepare "$run" template`. It exits 0. Its log names `just install` of `probe`, the line `installed <run>/scratch/home/.local/share/nicely/extensions/ncly-probe`, and the `python3` that it puts first in the PATH of the wrapper. Then run `ls "$run/scratch/home/.local/share/nicely/extensions"`: it lists `ncly-probe`, `ncly-probe.locales`, `ncly-probe.skill`, and `ncly-probe.toml`, beside the four files of `hello`. Report the output of `prepare` and of `ls` beside the quote.
- **Discovery.** Run `just verify step "$run" template --help`. Exit 0, and the `EXTENSIONS` section lists `probe:  Inspect files` beside `hello` and `skill`.
- **Help.** Run `just verify step "$run" template probe --help`. Exit 0, and `COMMANDS` lists `inspect:  Give the size and the line count of a file`.
- **Inspect as JSON.** Run `just verify step "$run" template probe inspect README.md --json`. Exit 0, stderr is empty, and stdout is `{"ok":true,"contract_version":1,"lines":<n>,"size_bytes":<bytes>}`. `<n>` is the number that `wc -l < README.md` prints, plus 1 when the last line of `README.md` has no newline. `<bytes>` is the number that `wc -c < README.md` prints. Report the output of both `wc` calls beside the quote.
- **Inspect as text.** Run `just verify step "$run" template probe inspect README.md`. Exit 0, and stdout is `README.md: <n> lines, <size>`. `<size>` is `<bytes>` in decimal units, with one decimal above 1,000 bytes, such as `5.4 kB` for 5372 bytes.
- **Missing file.** Run `just verify step "$run" template probe inspect nothere.md`. Exit 2, stdout is empty, and stderr starts with `error[USAGE_INVALID]: nothere.md is no file that probe can inspect.`, with the hint `ncly probe inspect notes.md`.
- **Describe.** Run `just verify step "$run" template describe probe inspect --json`. Exit 0. The one entry of `commands` has the path `probe inspect` and the source `extension`. Its `effects` are all `false`, and `requires.bins` holds `python3` with the version `3.9`.
- **Skill.** Run `just verify step "$run" template skill view probe`. Exit 0. Its first line names the skill folder `<run>/scratch/home/.local/share/nicely/extensions/ncly-probe.skill`, and the file of the skill `probe` follows, from its frontmatter `name: probe`.
- **Doctor.** Run `just verify step "$run" template doctor probe --json`. Exit 0, and the check `probe.python3` is a `pass` whose message names the version of `python3` and its path under `<run>/scratch/tools`.

## Gotchas

- `prepare` renders the working tree of the checkout, uncommitted changes included, as `evidence/commit.txt` marks with `dirty`. The template of a change has no release tag, so `prepare` tags a scratch copy of `copier.yml` and `template/` as `v0.0.0`
- `prepare` needs the network once, when `uv` holds no Copier of the pinned version yet
- A shim of mise or pyenv finds no Python under the scratch home, so `prepare` links the first `python3` of `PATH` that answers there. When none answers, it stops and names the install command
- `just template`, outside `ncly`, proves the rest of the template: `just check`, `just contract`, and `just update` of the rendered repository. Verify drives what a user runs through `ncly`
