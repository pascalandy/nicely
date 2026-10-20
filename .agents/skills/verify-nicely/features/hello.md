# hello

`hello` is the Go fixture extension, installed by `just verify` as an external extension in the scratch data folder. A user finds it in `ncly --help`, runs its commands through `ncly hello`, reads its declarations in `ncly describe`, its skill in `ncly skill view hello`, and its source in `ncly doctor`. Other features, such as `auth`, build on this run.

## Sub-features

- `hello-discovery` lists `hello` under the extensions of `ncly --help`, from the scratch data folder
- `hello-run` greets a name, as text and as JSON
- `hello-warning` greets the world with the warning `HELLO_UNNAMED` and exit 0
- `hello-error` refuses with the extension's own code `HELLO_GRUMPY` and exit 1
- `hello-report` answers a failed report on stdout with exit 78
- `hello-describe` describes `hello run` with the source `extension`
- `hello-skill` serves the skill installed beside the extension
- `hello-form` asks for a missing skill name in the form for missing values on a real terminal, then prints `Next time:` and the full command, and fails with that command as its hint outside a terminal
- `hello-doctor` reports where `hello` runs from

## How to get to it (user POV)

- `ncly --help` and `ncly hello --help`
- `ncly hello run <name>`, `ncly hello check`
- `ncly describe hello run`
- `ncly skill view hello`, in a pipe and in a terminal
- `ncly skill view` without a name, in a terminal and in a pipe
- `ncly doctor extensions`

## Driving it with verify.sh

Preconditions:

- A new run from `just verify`, whose Doctor step passed

- **Discovery.** Run `just verify step "$run" hello --help`. Exit 0, and the `EXTENSIONS` section lists `hello:  Greet people` beside `skill`.
- **Help.** Run `just verify step "$run" hello hello --help`. Exit 0, and `COMMANDS` lists `check`, `pick`, `run`, `say`, and `wait`.
- **Greet as text.** Run `just verify step "$run" hello hello run Ada`. Exit 0, stdout is `Hello, Ada!`, and stderr is empty.
- **Greet as JSON.** Run `just verify step "$run" hello hello run Ada --json`. Exit 0, and stdout is `{"ok":true,"contract_version":1,"greeting":"Hello, Ada!"}`.
- **Warning.** Run `just verify step "$run" hello hello run --json`. Exit 0, `ok` is `true`, and `warnings` holds the warning `HELLO_UNNAMED`.
- **Own error.** Run `just verify step "$run" hello hello run --grumpy`. Exit 1, stdout is empty, and stderr starts with `error[HELLO_GRUMPY]: Not today.`
- **Failed report.** Run `just verify step "$run" hello hello check --json`. Exit 78, and stdout is `{"ok":false,"contract_version":1,"checks":[{"id":"hello.mood","status":"fail"}]}`.
- **Describe.** Run `just verify step "$run" hello describe hello run --json`. Exit 0, and the one entry of `commands` has the path `hello run` and the source `extension`.
- **Skill.** Run `just verify step "$run" hello skill view hello`. Exit 0, and its first line names the skill folder `<run>/scratch/home/.local/share/nicely/extensions/ncly-hello.skill`, followed by the file of the skill `hello` as it is, frontmatter and Markdown included, because only a terminal renders it.
- **Skill in a terminal.** Run `just verify tmux "$run" hello skill view hello`. Exit 0. The terminal shows `folder  <run>/scratch/home/.local/share/nicely/extensions/ncly-hello.skill`, an empty line, `HELLO`, then the sentence of the skill, indented by two spaces, without its backticks. It shows no `---` and no `name: hello`, because the view removes the frontmatter. tmux drops the styles, which `styles.txtar` proves.
- **Skill from the form.** Run `just verify tmux "$run" hello --answer 'The name of the skill' hello skill view`. Exit 0. Before Enter, the terminal shows the form for missing values: `┃ name`, `┃ The name of the skill, as ncly skill list gives it`, `┃ > hello`, then `ctrl+e complete • enter submit`. After Enter, the form is gone, and the terminal shows the skill as in the step above, then `Next time:` and `  ncly skill view hello`.
- **Missing name.** Run `just verify step "$run" hello skill view`. Exit 2, stdout is empty, and stderr starts with `error[USAGE_INVALID]: The argument <name> is missing.`, with the hint `ncly skill view <name>`.
- **Proof.** Run `just verify step "$run" hello doctor extensions --json`. Exit 0, and the check `extensions.hello` is a `pass` with the source `extension` and the path `<run>/scratch/home/.local/share/nicely/extensions/ncly-hello`.

## Gotchas

- `hello check` exits 78 on purpose: its report always fails, and that is the end state to check
- A `hello` found anywhere else than under `<run>/scratch/home` means the wrapper was bypassed: rerun the step through `just verify step`
- `hello pick` writes the file `picked` in the current folder, which is the checkout: leave it out of a run
- `hello wait` waits for a signal and never ends alone: leave it out of a run, because `step` waits for the command to end
