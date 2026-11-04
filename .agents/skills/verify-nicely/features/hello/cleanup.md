# Handler cleanup after a resistant worker

The fixture command `hello staged <worker.sh>` runs a worker with private staging, then removes staging when the worker returns.
The scenario [staged_cleanup.txtar](../../../../../testdata/extensions/hello/testdata/script/staged_cleanup.txtar) supplies the worker and an independent stage check.
Copy its `worker.sh`, `grandchild.sh`, `check.sh`, input and previous output into the run's scratch folder.
Keep that folder as the working directory for every call below.

Drive `hello staged worker.sh --json` through `just verify step`, with `CLEANUP_WORKER=stubborn`.
After `worker.pid` appears, send SIGINT to the real host, then repeat with SIGTERM in a fresh scratch subfolder.
A process group around the Verify step can deliver the signal while the host reaches workers in their own sessions.
The transcript must show exit 130 with `INTERRUPTED`, or 143 with `TERMINATED`, and retain the `stage` data key.
Both calls must finish within the 15-second cleanup budget plus the 2-second stream drain and scheduling tolerance.
Check that `cleanup.done` exists, no `.ncly-stage-*` remains, both worker PIDs are gone, and input and previous output match their copies.
Also drive the cooperative worker with `CLEANUP_WORKER=cooperative`.
The scenario covers the direct-extension control with the same SDK and no host.
The [stubborn control](../../../../../testdata/script/extension_cancel_stubborn.txtar) proves the final handler deadline and the force-stop of a cleanup child.
Only the direct handler receives the reservation; see [Cancellation](../../../../../docs/north-star/contract/programs-that-ncly-runs.md).

For the repeated production-budget check, run `just verify stress-cleanup "$run" 10 20000` from the same clean committed checkout that built the run.
It runs the real public host ten times per signal, with 20,000 files and partial output in each private stage.
The driver counts the files before signalling the host through the worker's ancestry.
Every call records the signal answer, preserved files, stage absence and stopped worker PIDs.
It records deletion and handler cleanup times, plus the margin inside the 2-second reservation, in `evidence/cleanup-stress.txt` and `.json`.
Run Cleanup and quote `cleanup-stress` afterward. A failed run or nonpositive cleanup margin fails the driver.
