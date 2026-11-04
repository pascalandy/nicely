# Cancellation

On Ctrl-C or SIGTERM, `ncly` sends the signal to every program it started and to every descendant that it can still reach. The host keeps a 15-second cleanup bound.
After 13 seconds, it force-stops descendants, leaving a nominal 2-second reservation for the extension handler.
Force-stop and exit overhead use part of that reservation.
Staging cleanup must fit the remaining time.
The force-stop targets all observed descendants, including programs that the handler started for cleanup.
The reservation protects only the direct handler; the runner cannot distinguish cleanup children from workers.
A handler must account for stopped cleanup programs before reporting its signal outcome.
Keep child diagnostics separate from the structured answer.
At 15 seconds, it force-stops every remaining process, including an extension that does not finish cleanup.
An unfinished handler can leave staging behind.
The host then reports `INTERRUPTED` with exit 130, or `TERMINATED` with exit 143, without the handler's result data.
A Python program still has its own 10 seconds to stop its children.
The shared runner exposes this reservation as `Program.CleanupReserve`; its default is zero.
The reservation stays inside `Cleanup`, and values outside zero through `Cleanup` are clamped to that range.
A direct extension keeps the ordinary worker budget, because it has no outer host deadline. A list command that completion runs gets SIGKILL at once instead, as [Completion values](../completion-values.md) defines. Before each signal, `ncly` finds descendants through their parent process, so a descendant that left the process group of its parent, as the children of transcript's Python program do, still receives the signal. On Linux, `ncly` registers as a child subreaper, so it also reaches a descendant whose parent already exited. macOS has no subreaper, so a descendant whose parent exited before the signal is the one exception there, and it keeps running. `ncly` keeps every finished output and removes only its own temporary files. It exits 130 with `INTERRUPTED` or 143 with `TERMINATED`, and under `--json` the answer keeps the data keys that apply, such as `output_dir` or the `results` of finished items. A second signal during this cleanup is ignored. A SIGKILL sent to `ncly` itself leaves no time to clean up.

On one macOS host, each probe removed 20,000 staging files. With 100,000 files on that host, both signals reached the 15-second cutoff with staging left.
[Issue #473](https://github.com/pascalandy/nicely/issues/473) records these measurements and signal answers, and tracks recovery of incomplete cleanup.
