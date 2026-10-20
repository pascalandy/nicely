# Commands that core calls

Core calls some commands of an installed extension itself, to serve one of its own commands, apart from any call that a caller makes. Completion calls a list command for the values that only an extension knows, as [Completion values](completion-values.md) defines. Doctor calls each status command, as [Status command](status-command.md) defines. Each of these two sections gives the conditions that the command's own declaration meets, the limit of the call, and how core reads the answer.

Every command that core calls follows these rules:

- Core dispatches it through the extension that its discovery found, as [Dispatch](dispatch.md) defines, with the canonical arguments and the [Answer check](answer-check.md). Its environment is that of [Programs that ncly runs](programs-that-ncly-runs.md), with `NCLY_JSON=1`
- With `NCLY_JSON=1`, it never asks, so it runs in its own process group, unlike an ordinary dispatch. From M18-T4, it starts in a new session, as [Programs that ncly runs](programs-that-ncly-runs.md) defines for every program without the terminal
- Both sections require `effects` with `user_writes` and `paid` both false, so the command neither writes a user file nor sends a paid request. At its limit, core therefore sends SIGKILL to the command and its descendants at once, without the 15 seconds of cleanup. Core also closes its pipes instead of starting the drain limit

Every other rule belongs to one of the two sections alone, such as the flags that doctor passes to a status command.
