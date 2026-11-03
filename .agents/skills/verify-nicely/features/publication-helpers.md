# publication-helpers

`publication-helpers` are the scenario commands `keeps` and `without` of the test kit, which a rendered extension runs in `just check`. On Linux, the kit runs their calls under its guard. [template.md](template.md) renders and installs the extension `probe` that this feature uses.

## Sub-features

- `kit-keeps` runs a call of `probe inspect` with its exact exit in a scenario of the rendered repository
- `kit-without` refuses a hard link for one call on Linux, and fails a check whose call never tries one

## How to get to it (user POV)

- A scenario in `testdata/script/` of an extension repository, with the lines `keeps <exit> ncly <root> <command> <arg>...` and `[linux] without link keeps <exit> ...`
- `just check` in that repository

## Driving it with verify.sh

Preconditions:

- A new run from `just verify`, whose Doctor step passed
- The Render and install step of [template.md](template.md) passed in this run
- A Linux host, because only Linux runs the guard

- **Keeps and without.** Write `$run/scratch/ncly-probe/testdata/script/helpers.txtar` with the line `keeps 0 ncly probe inspect notes.txt`, then a line `-- notes.txt --` and a line `one`. Run `(cd "$run/scratch/ncly-probe" && just check)`. Exit 0. Then add the line `[linux] without link keeps 0 ncly probe inspect notes.txt` after the first line, and run `just check` there again. Exit not 0. The output holds `keeps: no call tried` and `link` in one line, because `inspect` publishes no file. Then delete `helpers.txtar`. Report both outputs.

## Gotchas

- On macOS, `without` fails and names the platform. Run this feature on Linux only
- Delete `helpers.txtar` before another step of `template.md` runs `just check`
