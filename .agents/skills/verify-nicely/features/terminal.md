# terminal

Inside tmux, `ncly` answers at once, even when the tmux server does not answer, because it asks neither tmux nor the terminal anything as it loads. `ncly` loads Huh, Bubbles, and Glamour, whose Lip Gloss v2.0.6 runs `tmux info` at load, so go.mod replaces it with a fork that waits for the first print. This feature runs the built `ncly` on a real terminal, in a pane of a private tmux server, while the `TMUX` variable names a second private server that SIGSTOP holds, so that a call to `tmux` would wait.

## Sub-features

- `terminal-version` prints the version at once inside tmux whose server does not answer
- `terminal-help` prints the root help at once inside tmux whose server does not answer

## How to get to it (user POV)

- `ncly --version` inside tmux
- `ncly --help` inside tmux

## Driving it with verify.sh

Preconditions:

- A new run from `just verify`, whose Doctor step passed
- `tmux -V` answers. The run needs no tmux session of its own

`just verify tmux "$run" <feature> <ncly args...>` runs the call through the wrapper, like `step`, but on a terminal: in a pane of the private server `nicely-verify-<run>`, with `TMUX` naming the private server `nicely-verify-<run>-stuck`, which it holds with SIGSTOP during the call. The pane runs `ncly` with only `TERM`, which tmux sets, and `TMUX`, beside what the wrapper sets, because a variable of the caller, such as `NO_COLOR=1`, which agents set, can skip the detection that asks tmux, so any caller proves the end state. `DBUS_SESSION_BUS_ADDRESS` and `XDG_RUNTIME_DIR` pass too, when the caller sets them, so `ncly` reaches the session bus of the keychain on Linux. They detect nothing. It records the time it started, what the pane showed, and the exit code in the transcript, or `no exit within the limit of 5s`, and then exits 1. It resumes and kills both servers on every exit.

Before the `ncly` arguments, each `--answer <prompt> <keys>` waits until the pane shows `<prompt>`, then types `<keys>`. It records what the pane shows after the typing, under `--- terminal, once it showed <prompt>, after the typing, before Enter`, then presses Enter. Keys of `^C` press Ctrl-C instead, recorded as `--- terminal, once it showed <prompt>, then Ctrl-C`. The transcript never holds the typed keys, only what the pane shows. Each answer adds 5 seconds to the limit. When a prompt never shows, the transcript says `--- the prompt <prompt> never showed`, and the step exits 1. [auth.md](auth.md) answers the masked field and the confirmations this way.

- **Version.** Run `just verify tmux "$run" terminal --version`. Exit 0, the terminal shows `ncly v<version>-<commit>`, as in the Doctor step, and the transcript ends with `--- exit 0, before the limit of 5s`.
- **Help.** Run `just verify tmux "$run" terminal --help`. Exit 0, the terminal shows the whole root help, from `Curated CLI toolbox for humans and AI agents.` through its `FLAGS` and `EXAMPLES`, and the transcript ends with `--- exit 0, before the limit of 5s`.
- **Servers gone.** After both calls, run `tmux -L "nicely-verify-${run##*/}" ls` and `tmux -L "nicely-verify-${run##*/}-stuck" ls`. Each exits 1 with `error connecting to` its socket, `No such file or directory`, because the step removes the socket files too: report both beside the quote.

## Gotchas

- A healthy tmux server answers `tmux info` at once, so only the held server tells a build that asks tmux at load apart: `no exit within the limit of 5s` means `ncly` waited on tmux, and the feature fails
- The step stops with exit 1 when tmux lies in a folder that holds an `ncly-<domain>` file, which the wrapper leaves out of PATH, because `ncly` could not reach tmux and any build would pass. Report it as unreachable, with that message
- The pane shows text without styles, because `capture-pane` drops them. `terminal.txtar` proves the styles on a terminal
- Never point `tmux -L` at another name, and never send SIGSTOP to another tmux server: the two private servers are the only ones the run may touch
