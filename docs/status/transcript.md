# Transcript

This part of [Current build](../status.md) says what the `transcript` extension does.

`just install` also puts the `transcript` extension in `~/.local/share/nicely/extensions/`. `ncly doctor transcript` checks that `uv` and Python 3.12 or later are installed. Then `ncly transcript status` checks the cookies of the browser that `[transcript]` sets. Without a browser, it passes. With `browser = "arc"`, it warns with `TRANSCRIPT_BROWSER_COOKIES_SKIPPED` when no cookie file of Arc's default profile opens. It reads no key, starts no program, and uses no network.

`ncly transcript run youtube --url <url> --dry-run` checks the URL of one video, the output folder, and the `[transcript]` tables of the config. It then runs the transcript program through `uv` and shows its plan, with no request to YouTube or Deepgram. It reads no key and creates no record.

Without `--dry-run`, the call needs the key of `deepgram`, which a human grants to transcript. A key with a character that no request can carry, such as a line break, answers `AUTH_REJECTED` before any record. It records the run in `~/.local/state/nicely/runs/` before the download. It saves each candidate result folder in the record, then creates that folder, before the upload. The folder takes the title of the video, with `-2`, `-3`, and so on when the name is taken. Deepgram gets the audio once, and the folder receives `raw_transcript.txt`, `raw_sentences.txt`, and `raw_transcript.json`. The call checks, hashes, and saves them in the record before it answers `output_dir` and `transcript`.

A key that Deepgram refuses gives `AUTH_REJECTED` in the item, under `RUNTIME`. Each failure keeps its cause in the item. Only a temporary failure before the result folder exits 75. A Deepgram answer lost after the upload leaves the step `unknown`, and nothing retries the upload.

The same call with the same `--run-id` returns a completed run again, without `uv`, the program, Deepgram, or a key. It checks each file first, and a file that changed or went missing answers `RESUME_UNSAFE`. A config change never moves a saved result or a pending run. Once the audio may have reached Deepgram, the same call answers `RESUME_UNSAFE` too. Another URL or `--output` under that ID answers `USAGE_INVALID`. A dry run with the ID plans `reuse`, or refuses as the call would.

Ctrl-C exits 130 and SIGTERM 143, and neither leaves a `uv`, Python, or yt-dlp process. The answer keeps the result folder. A stop after the upload leaves the step `unknown`, so the same call starts nothing, and a completion before the stop stays complete. A second call under a busy run ID exits 75 and starts nothing. Two runs of one title write two folders.
