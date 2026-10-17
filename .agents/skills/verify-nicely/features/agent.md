# agent

`agent` is the first-party extension that runs one trusted task through Codex. A user installs it from a checkout with `just install`, which puts `ncly` in `~/.local/bin/` and `ncly-agent` with its manifest, catalogs, and skill in the `extensions` folder of Nicely's data folder. The user then finds `agent` in `ncly --help`, reads its skill, checks Codex with `ncly doctor agent`, and checks a task with `ncly agent run --tools --dry-run`, which shows the invocation of Codex and starts no task. This feature never starts a real Codex: `just verify prepare "$run" agent` puts a counted stub `codex` first in the wrapper's `PATH`.

## Sub-features

- `agent-install` installs `ncly` and `ncly-agent` with `just install` in the scratch home, and the wrapper then runs the installed `ncly`
- `agent-discovery` lists `agent` under the extensions of `ncly --help`, from the scratch data folder
- `agent-help` shows the help of `ncly agent` and `ncly agent run` from the installed manifest
- `agent-doctor` passes `agent.codex` for a recent Codex, fails it with `PREREQ_MISSING` and the install command for an older one, and only probes the version
- `agent-skill` serves the skill installed beside the extension
- `agent-dry-run` prepares a task, shows the resolved settings and the invocation of Codex, starts no task, and records nothing
- `agent-tools` refuses a call without `--tools` with `CAPABILITY_UNSUPPORTED`, even in a dry run
- `agent-depth` gives Codex `NCLY_AGENT_DEPTH` plus one, and refuses a call at `[agent] max_depth` with `AGENT_DEPTH_LIMIT` before it probes Codex

## How to get to it (user POV)

- `just install`, from a checkout
- `ncly --help`, `ncly agent --help`, and `ncly agent run --help`
- `ncly doctor agent` and `ncly doctor extensions`
- `ncly skill view agent`
- `ncly agent run <task> --tools --model <id> --dry-run`, as text and as JSON

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
- **Dry run.** Run `just verify step "$run" agent agent run 'Review this branch' --tools --model test-model --dry-run --json`. Exit 0, and stdout is one answer with `"dry_run":true`, `"harness_version":"0.163.0"`, `"model":"test-model"`, `"effort":"medium"`, and `cwd`, the folder of the checkout. `invocation` starts with `<run>/scratch/stub/codex` and `exec`, holds `--sandbox`, `workspace-write`, `--cd`, and the folder of the checkout, and ends with `-`. `environment` is `{"NCLY_AGENT_DEPTH":"1"}`. The one item of `plan` has the `output` `<run>/scratch/home/.local/state/nicely/runs`, and `pending_checks` holds `agent.login`.
- **Dry run as text.** Run `just verify step "$run" agent agent run 'Review this branch' --tools --model test-model --dry-run`. Exit 0. stdout starts with `Would run Codex 0.163.0 with this command, and give it the task on its input:`, then the invocation on one line, then `model: test-model`, `effort: medium`, `folder:` and the folder of the checkout, `environment: NCLY_AGENT_DEPTH=1`, and `pending  agent.login`.
- **No record.** Run `ls "$run/scratch/home/.local/state/nicely/runs"`. It fails, because a dry run records nothing. Report its output beside the quote.
- **Without --tools.** Run `just verify step "$run" agent agent run 'Review this branch' --model test-model --dry-run --json`. Exit 78, stdout is empty, and stderr holds the code `CAPABILITY_UNSUPPORTED` and the message `Codex cannot run a task with its tools turned off.`
- **Depth limit.** Run `NCLY_AGENT_DEPTH=2 just verify step "$run" agent agent run 'Review this branch' --tools --model test-model --dry-run --json`. Exit 2, stdout is empty, and stderr holds the code `AGENT_DEPTH_LIMIT` and the message `This call already runs inside 2 agents, and max_depth in [agent] allows 2.` This call adds no line to `evidence/codex-calls.txt`.
- **One level deeper.** Run `NCLY_AGENT_DEPTH=1 just verify step "$run" agent agent run 'Review this branch' --tools --model test-model --dry-run --json`. Exit 0, and `environment` is `{"NCLY_AGENT_DEPTH":"2"}`.
- **No execution yet.** Run `just verify step "$run" agent agent run 'Review this branch' --tools --model test-model --json`. Exit 78, and stderr holds `CAPABILITY_UNSUPPORTED` and the message `This build of the agent extension prepares a Codex task, but cannot run it yet.`
- **No task.** After the cleanup, run `just verify quote "$run" codex-calls`. Each line is `codex --version`, one per `ncly doctor agent` call and one per `ncly agent run` call other than the one at the depth limit, and `test ! -e "$run/evidence/codex-task.txt"` succeeds, because no other call reached the stub.

## Gotchas

- Run `prepare` before any `agent` step: before it, the wrapper runs the build of `just verify`, without `ncly-agent`, and `PATH` may hold a real Codex
- `prepare` runs `just install` with the scratch home, so it never writes the real `~/.local/bin/ncly`. It keeps the caller's Go caches, which live under the real home
- `CODEX_STUB_VERSION` and `NCLY_AGENT_DEPTH` reach `ncly` and the stub through the wrapper, which keeps the caller's variables beside `HOME`, the XDG folders, and `PATH`
