# Handler cleanup after a resistant worker

The fixture command `hello staged <worker.sh>` runs a worker with private staging, then removes staging when the worker returns.
The scenario [staged_cleanup.txtar](../../../../../testdata/extensions/hello/testdata/script/staged_cleanup.txtar) supplies the worker and an independent stage check.
Copy its `worker.sh`, `grandchild.sh`, `check.sh`, input and previous output into the run's scratch folder.
From the checkout root, set `repo="$PWD"` before entering that folder.
In that folder, drive the calls below that name `hello staged`.

Call the same Verify entry point directly, because `just` starts its recipe at the checkout root:

```bash
CLEANUP_WORKER=stubborn bash "$repo/.agents/skills/verify-nicely/scripts/verify.sh" step "$run" hello hello staged ./worker.sh --json
```
After `worker.pid` appears, send SIGINT to the real host, then repeat with SIGTERM in a fresh scratch subfolder.
A process group around the Verify step can deliver the signal while the host reaches workers in their own sessions.
The transcript must show exit 130 with `INTERRUPTED`, or 143 with `TERMINATED`, and retain the `stage` data key.
Both calls must finish within the 15-second cleanup budget plus the 2-second stream drain and scheduling tolerance.
Check that `cleanup.done` exists, no `.ncly-stage-*` remains, both worker PIDs are gone, and input and previous output match their copies.
Also drive the cooperative worker with `CLEANUP_WORKER=cooperative`.
The scenario covers the direct-extension control with the same SDK and no host.
The [stubborn control](../../../../../testdata/script/extension_cancel_stubborn.txtar) proves the final handler deadline and the force-stop of a cleanup child.
Only the direct handler receives the reservation; see [Cancellation](../../../../../docs/north-star/contract/programs-that-ncly-runs/cancellation.md).

From `$repo`, run `just verify stress-cleanup "$run" 10 20000` on the same clean committed checkout that built the run.
It runs the real public host ten times per signal, with 20,000 files and partial output in each private stage.
The driver counts the files before signalling the host through the worker's ancestry.
Every call records the signal answer, preserved files, stage absence and stopped worker PIDs.
It records deletion and handler cleanup times, plus the margin inside the 2-second reservation, in `evidence/cleanup-stress.txt` and `.json`.
That margin subtracts measured handler cleanup only. The signal-to-exit time also includes force-stop and exit overhead.
The difference `15000 - signal_ms` is the observed margin before the nominal final deadline.
Run Cleanup and quote `cleanup-stress` afterward. A failed run or nonpositive cleanup margin fails the driver.
