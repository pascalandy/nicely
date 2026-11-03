# template

`template` is the extension template of the checkout, which an agent renders with the command of the `nicely-extensions` skill. `just verify prepare "$run" template` renders it as the extension `probe`, with the Copier that the template pins, and installs it with `just install` of the rendered repository. A user then finds `probe` in `ncly --help`, runs `ncly probe inspect`, and reads its declaration in `ncly describe`, its skill, and its check in `ncly doctor`.
[extensions-skill.md](extensions-skill.md) covers the command of the `nicely-extensions` skill that renders it.

## Sub-features

- `template-loop` sends each command of the rendered repository to its `nicely-extension` skill, which builds it in rounds until `just finish` passes
- `template-install` renders the template, and installs the rendered repository with `just install` in the scratch data folder
- `template-rule-index` counts the rules of the rule index in `just check` of the rendered repository
- `template-publication` refuses in `just check` a command that writes a user file with no row in `## Publication` of `spec.md`
- `template-acceptance` reads the answer of `probe inspect` in its acceptance case, and refuses a count of 30 lines for 3
- `template-release` gives the release steps in the `AGENTS.md` of the rendered repository, and the commands that install a release tag in its `README.md`
- `template-discovery` lists `probe` under the extensions of `ncly --help`
- `template-inspect` gives the size and the line count of a file, as JSON and as text, through the worker `inspect.py`, which `python3` runs through `sdk/wrap`
- `template-dry-run` gives the same size and line count without changing the file
- `template-usage` refuses a missing file with `USAGE_INVALID`, before the worker starts
- `template-refusal` refuses a file with a NUL byte, which the worker accepts, with `PROBE_NOT_TEXT`
- `template-describe` describes `probe inspect`, with its program `python3` and no effect
- `template-skill` serves the skill installed beside `probe`
- `template-doctor` reports the `python3` that `probe inspect` needs

## How to get to it (user POV)

- The command of the `nicely-extensions` skill, `uv tool run copier@<version> copy ...`, then `just install` in the new repository
- `ncly --help` and `ncly probe --help`
- `ncly probe inspect <file>`
- `ncly probe inspect <file> --dry-run --json`
- `ncly describe probe inspect`
- `ncly skill view probe`
- `ncly doctor probe`

## Driving it with verify.sh

Preconditions:

- A new run from `just verify`, whose Doctor step passed
- `uv`, `just`, and a `python3` of 3.9 or later, which answers under the scratch home

- **Render and install.** Run `just verify prepare "$run" template`. It exits 0. Its log names `just install` of `probe`, the line `installed <run>/scratch/home/.local/share/nicely/extensions/ncly-probe`, and the `python3` that it puts first in the PATH of the wrapper. Then run `ls "$run/scratch/home/.local/share/nicely/extensions"`: it lists `ncly-probe`, `ncly-probe.locales`, `ncly-probe.skill`, and `ncly-probe.toml`, beside the four files of `hello`. Report the output of `prepare` and of `ls` beside the quote.
- **Release and install text.** Run `grep -n -A12 '^## Release$' "$run/scratch/ncly-probe/AGENTS.md"` and `grep -n -A3 '^git clone' "$run/scratch/ncly-probe/README.md"`. Both exit 0. `AGENTS.md` names `just release vX.Y.Z` and the release receipt. `README.md` holds the lines `git clone --branch TAG https://example.com/ncly-probe ncly-probe`, `cd ncly-probe`, `just install`, and `ncly describe probe --json`. Report both outputs beside the quote.
- **Loop text.** Run `grep -n -A12 '^## Done when$' "$run/scratch/ncly-probe/AGENTS.md"`, then `grep -n '^## Open lines$\|^## Case matrix$' "$run/scratch/ncly-probe/spec.md"`. Both exit 0, and `spec.md` lists both parts. "Done when" names `just finish`, the hosted checks, 5 rounds, and a "full stop" request that runs only after it. Then run `grep -n '^##' "$run/scratch/ncly-probe/.agents/skills/nicely-extension/SKILL.md"`. It lists `## Intake`, `## Acceptance lines`, `## Test files`, `## Rounds`, and `## Readiness review`. Then run `grep -n '^##' "$run/scratch/ncly-probe/.agents/skills/nicely-extension/intake.md"`. It lists `## Intake of a script`, `## Intake of a program`, and `## Evidence of the intake`. Report the four outputs.
- **Finish gate.** Run `(cd "$run/scratch/ncly-probe" && just finish)`. Exit 1. Stderr names each `Fill:` line of `vision.md`, `spec.md`, `qa.md`, and `intake.toml` as `<file>:<line>: this line of the form is not filled`, and `review.toml: no readiness review is recorded`. Then run `just records` there. Exit 0, and it names `vision.md`. Report both outputs.
- **Rule index.** Run `(cd "$run/scratch/ncly-probe" && just check)`. Exit 0, and stderr holds `AGENTS/rule-index.md: <p> of <n> rules are Prose only, and <q> Prose on purpose`. Count the rows of `AGENTS/rule-index.md` and `AGENTS/rule-index/` with `grep`, and report both.
- **Publication.** Run `grep -n -A6 '^## Publication$' "$run/scratch/ncly-probe/spec.md"`. Exit 0. The part holds a line that starts with `None:`, names `ncly probe inspect`, and ends with `creates or replaces no user file.`. Then run `cp "$run/scratch/ncly-probe/manifest.toml" "$run/manifest.toml"`. In `$run/scratch/ncly-probe/manifest.toml`, change the line `effects = {}` to `effects = { user_writes = true }`. Run `(cd "$run/scratch/ncly-probe" && just check)`. Exit not 0. The output holds a line that names `ncly probe inspect`, `may write a user file`, and `## Publication has no row for its create mode`. Then run `cp "$run/manifest.toml" "$run/scratch/ncly-probe/manifest.toml"`. Report the outputs.
- **Acceptance case.** Run `grep -n '^reads ' "$run/scratch/ncly-probe/testdata/script/acceptance.txtar"`. Exit 0, and the line ends with `answer size_bytes lines`. Run `cp "$run/scratch/ncly-probe/worker/inspect.py" "$run/inspect.py"`. In `$run/scratch/ncly-probe/worker/inspect.py`, change `"lines": lines,` to `"lines": lines * 10,`. Run `(cd "$run/scratch/ncly-probe" && just scenario acceptance)`. Exit not 0, and the output holds `reads: answer gives` and `lines: 30`. Then run `cp "$run/inspect.py" "$run/scratch/ncly-probe/worker/inspect.py"`, and the same `just scenario acceptance`. Exit 0. Report the outputs.
- **Discovery.** Run `just verify step "$run" template --help`. Exit 0, and the `EXTENSIONS` section lists `probe:  Inspect files` beside `hello` and `skill`.
- **Help.** Run `just verify step "$run" template probe --help`. Exit 0, and `COMMANDS` lists `inspect:  Give the size and the line count of a file`.
- **Inspect as JSON.** Run `just verify step "$run" template probe inspect README.md --json`. Exit 0, stderr is empty, and stdout is `{"ok":true,"contract_version":1,"lines":<n>,"size_bytes":<bytes>}`. `<n>` is the number that `wc -l < README.md` prints, plus 1 when the last line of `README.md` has no newline. `<bytes>` is the number that `wc -c < README.md` prints. Report the output of both `wc` calls beside the quote.
- **Dry run.** Run `just verify step "$run" template probe inspect README.md --dry-run --json`. Exit 0, and stderr is empty. Stdout has `"ok":true`, `"contract_version":1`, and the same `lines` and `size_bytes` values as the run. Repeat both `wc` calls. Their results remain unchanged.
- **Inspect as text.** Run `just verify step "$run" template probe inspect README.md`. Exit 0, and stdout is `README.md: <n> lines, <size>`. `<size>` is `<bytes>` in decimal units, with at most one decimal above 1,000 bytes, such as `5.4 kB` for 5372 bytes and `8 kB` for 7990 bytes.
- **Missing file.** Run `just verify step "$run" template probe inspect nothere.md`. Exit 2, stdout is empty, and stderr starts with `error[USAGE_INVALID]: nothere.md is no file that probe can inspect.`, with the hint `ncly probe inspect notes.md`.
- **Refusal.** Run `printf '\211PNG\r\n\032\n\000\000' > "$run/image.png"`, then `just verify step "$run" template probe inspect "$run/image.png"`. Exit 2, stdout is empty, and stderr starts with `error[PROBE_NOT_TEXT]: <run>/image.png is no text file, so its lines mean nothing.`, with a hint that names `notes.md`.
- **Describe.** Run `just verify step "$run" template describe probe inspect --json`. Exit 0. The one entry of `commands` has the path `probe inspect` and the source `extension`. Its `effects` are all `false`, and `requires.bins` holds `python3` with the version `3.9`.
  Its `modes` has `"dry_run":true`; every other mode is `false`.
- **Skill.** Run `just verify step "$run" template skill view probe`. Exit 0. Its first line names the skill folder `<run>/scratch/home/.local/share/nicely/extensions/ncly-probe.skill`, and the file of the skill `probe` follows, from its frontmatter `name: probe`.
- **Doctor.** Run `just verify step "$run" template doctor probe --json`. Exit 0, and the check `probe.python3` is a `pass` whose message names the version of `python3` and its path under `<run>/scratch/tools`.

## Gotchas

- `prepare` renders the working tree of the checkout, uncommitted changes included, as `evidence/commit.txt` marks with `dirty`. The template of a change has no release tag, so `prepare` tags a scratch copy of `copier.yml` and `template/` as `v0.0.0`
- `prepare` needs the network once, when `uv` holds no Copier of the pinned version yet
- A shim of mise or pyenv finds no Python under the scratch home, so `prepare` links the first `python3` of `PATH` that answers there. When none answers, it stops and names the install command
- `just template`, outside `ncly`, proves the rest of the template: `just contract`, `just update`, and the checks of `just check` of the rendered repository. Verify drives what a user runs through `ncly`
