# transcript-run

`transcript-run` goes on in a run of [transcript](transcript.md), whose `prepare` put a fake key in the wrapper's `DEEPGRAM_API_KEY`. A human grants the key of `deepgram` in a terminal, and a run gets the fake key. That run downloads the real audio and uploads it, and Deepgram refuses the fake key with 401, so nothing bills. A last run stops with Ctrl-C during its download.

## Sub-features

- `transcript-grant` grants the key of `deepgram` to transcript after a typed `y` on a real terminal
- `transcript-fake-key` runs one transcription with a fake key: Deepgram refuses it, and the item answers `AUTH_REJECTED` under `RUNTIME`, with its result folder
- `transcript-record` records the run with its failed step, and the same call starts nothing
- `transcript-other-call` refuses another URL under the run ID of that run with `USAGE_INVALID`, before anything starts
- `transcript-stop` stops a run with Ctrl-C during the real download: exit 130, no process left, and the step pending

## How to get to it (user POV)

- `ncly auth grant deepgram transcript` in a terminal, through `just verify tmux`
- `ncly transcript run youtube --url <url> --output <folder> --run-id <id> --json`, then `ncly run view <id> --json`
- Ctrl-C in a terminal during `ncly transcript run youtube ... --verbose`, through `just verify tmux`

## Driving it with verify.sh

Preconditions:

- The run of [transcript](transcript.md), whose steps passed from Install to Dry run as text, so `<run>/scratch/home/videos` exists
- `tmux` in `PATH`
- The network reaches YouTube and Deepgram

In `just verify step "$run" transcript-run <args>`, `transcript-run` names this feature, and `<args>` are the arguments of `ncly`. Quote each URL, since `?` is a pattern of the shell.

- **Grant.** Run `just verify tmux "$run" transcript-run --answer '[y/N]' y auth grant deepgram transcript`. Exit 0. Before Enter, the terminal shows `? Grant the key of deepgram to transcript?`, then `  Path  ` and the path of `ncly-transcript` in `<run>/scratch/home/.local/share/nicely/extensions/`, then `  [y/N] y`. After Enter, it shows `Granted the key of deepgram to transcript.`
- **Fake key.** Run `just verify step "$run" transcript-run transcript run youtube --url 'https://www.youtube.com/watch?v=EIEc43CxIvY' --output "$run/scratch/home/videos" --run-id verify-fake --json`. Exit 1 within a minute, and stdout is empty. The first error of stderr is `RUNTIME`, and the answer holds `"run_id":"verify-fake"`. Its one item has `"ok":false`, `"transcript":null`, and an `output_dir` in `<run>/scratch/home/videos/`, named after the title of the video. The item's one error has the code `AUTH_REJECTED`, the message `Deepgram refused the key.`, and the hint `ncly auth login deepgram`.
- **Result folder.** Run `ls -A` on the `output_dir` of Fake key. It prints `meta.txt` alone, since Deepgram answered no transcript. Report its output beside the quote.
- **Record.** Run `just verify step "$run" transcript-run run view verify-fake --json`. Exit 0. The run has `"status":"failed"`, `"attempts":1`, and `"tools"` with `"transcript":"1.0.0"` and the version of yt-dlp. Its step `transcribe` has `"effect":"paid"`, `"status":"failed"`, no artifact, and the error `AUTH_REJECTED`. Its `output_dir` is the one of Fake key.
- **Same call.** Run the call of Fake key again. Exit 1 within a few seconds, and stderr holds the code `RESUME_UNSAFE`, `"run_id":"verify-fake"`, and the hint `ncly run view verify-fake`, because the audio reached Deepgram once.
- **Another URL.** Run the call of Fake key with `--url 'https://youtu.be/EIEc43CxIvY'`. Exit 2 within a few seconds, because that ID names the call of Fake key. stderr holds the code `USAGE_INVALID`, `"run_id":"verify-fake"`, and the message `The run verify-fake holds another command, URL, or --output.`
- **Ctrl-C.** Run `just verify tmux "$run" transcript-run --answer 'Starting the audio download' '^C' transcript run youtube --url 'https://www.youtube.com/watch?v=EIEc43CxIvY' --output "$run/scratch/home/videos" --run-id verify-stop --verbose`. Exit 130. The terminal shows `transcript: Starting the audio download...`. Then it shows `error[INTERRUPTED]: The transcription stopped before it ended.` with the hint `ncly run view verify-stop`, and last `error[INTERRUPTED]: Stopped by Ctrl-C. Finished work is kept.` Between them, `--verbose` may show diagnostics of the engine, such as a `KeyboardInterrupt` traceback.
- **No process.** Right after Ctrl-C, run `pgrep -af 'uv[ ]run|transcript[.]py|yt[-]dlp'`. It exits 1 and prints nothing, since no `uv`, Python, or yt-dlp of the call remains. The brackets keep the pattern from matching the shell that runs it. Report its output beside the quote.
- **Stopped record.** Run `just verify step "$run" transcript-run run view verify-stop --json`. Exit 0. The run has `"status":"interrupted"`. Its step `transcribe` has `"status":"pending"`, `"started_at":null`, and the error `INTERRUPTED`, and its item has `"output_dir":null`. Ctrl-C came during the download, before the intent.

## Gotchas

- Call `ncly` only through `just verify step` or `just verify tmux`, whose wrapper forces the fake key. Core reads `DEEPGRAM_API_KEY` before the keychain, so another call could bill a real key. Never run `ncly auth login deepgram`: it writes the real keychain, where it would replace a real key
- Only Fake key, Same call, Another URL, and Ctrl-C run without `--dry-run`. Fake key reaches YouTube and Deepgram, Ctrl-C reaches YouTube alone, and the other two reach neither
