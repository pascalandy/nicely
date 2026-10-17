# agent

`agent` is the first-party extension that runs one trusted task through Codex. A user installs it from a checkout with `just install`, which puts `ncly` in `~/.local/bin/` and `ncly-agent` with its manifest, catalogs, and skill in the `extensions` folder of Nicely's data folder. The user then finds `agent` in `ncly --help`, reads its skill, and checks Codex with `ncly doctor agent`. This feature never starts a real Codex: `just verify prepare "$run" agent` puts a counted stub `codex` first in the wrapper's `PATH`.

## Sub-features

- `agent-install` installs `ncly` and `ncly-agent` with `just install` in the scratch home, and the wrapper then runs the installed `ncly`
- `agent-discovery` lists `agent` under the extensions of `ncly --help`, from the scratch data folder
- `agent-help` shows the help of `ncly agent` and `ncly agent run` from the installed manifest
- `agent-doctor` passes `agent.codex` for a recent Codex, fails it with `PREREQ_MISSING` and the install command for an older one, and only probes the version
- `agent-skill` serves the skill installed beside the extension

## How to get to it (user POV)

- `just install`, from a checkout
- `ncly --help`, `ncly agent --help`, and `ncly agent run --help`
- `ncly doctor agent` and `ncly doctor extensions`
- `ncly skill view agent`

## Driving it with verify.sh

Preconditions:

- A new run from `just verify`, whose Doctor step passed
- `just` in `PATH`

- **Install.** Run `just verify prepare "$run" agent`. It exits 0, and its last lines name `<run>/scratch/home/.local/bin/ncly` and `<run>/scratch/home/.local/share/nicely/extensions/ncly-agent` as installed. Report these lines beside the quote.
- **Installed build.** Run `just verify step "$run" agent --version`. Exit 0, and stdout names the build of the Doctor step, because the wrapper now runs the `ncly` that `just install` wrote.
- **Discovery.** Run `just verify step "$run" agent --help`. Exit 0, and the `EXTENSIONS` section lists `agent:  Run trusted tasks through Codex` beside `skill`.
- **Help.** Run `just verify step "$run" agent agent --help`, then `just verify step "$run" agent agent run --help`. Both exit 0. The first lists the command `run`. The second shows `ncly agent run [task] [flags]` and the flags `--cwd`, `--dry-run`, `--effort`, `--model`, `--task-file`, and `--tools`.
- **Source.** Run `just verify step "$run" agent doctor extensions --json`. Exit 0, and the check `extensions.agent` is a `pass` with the source `extension` and the path `<run>/scratch/home/.local/share/nicely/extensions/ncly-agent`.
- **Recent Codex.** Run `just verify step "$run" agent doctor agent --json`. Exit 0, and the one check `agent.codex` is a `pass` whose message names `codex 0.163.0` and the stub `<run>/scratch/stub/codex`.
- **Older Codex.** Run `CODEX_STUB_VERSION=0.162.0 just verify step "$run" agent doctor agent --json`. Exit 78, and `agent.codex` fails with `PREREQ_MISSING`, the message `codex 0.162.0 is older than 0.162.1.`, and a hint and `next` that hold the install command of the machine: `sudo pacman -S openai-codex` on Arch, `brew install --cask codex` on macOS, or `Install codex.` and no `next` without either package manager.
- **Skill.** Run `just verify step "$run" agent skill view agent`. Exit 0, and its first line names the folder `<run>/scratch/home/.local/share/nicely/extensions/ncly-agent.skill`, followed by the skill `agent`.
- **No task.** After the cleanup, run `just verify quote "$run" codex-calls`. Each line is `codex --version`, one per doctor call, and `<run>/evidence/codex-task.txt` does not exist.

## Gotchas

- Run `prepare` before any `agent` step: before it, the wrapper runs the build of `just verify`, without `ncly-agent`, and `PATH` may hold a real Codex
- `prepare` runs `just install` with the scratch home, so it never writes the real `~/.local/bin/ncly`. It keeps the caller's Go caches, which live under the real home
- `CODEX_STUB_VERSION` reaches the stub through the wrapper, which keeps the caller's variables beside `HOME`, the XDG folders, and `PATH`
