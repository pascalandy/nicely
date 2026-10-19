# transcript

`transcript` is the first-party extension that transcribes one YouTube video with Deepgram. A user installs it from a checkout with `just install`. It puts `ncly-transcript` with its manifest, catalogs, and skill in the `extensions` folder of Nicely's data folder. The user then finds `transcript` in `ncly --help`, reads its skill, checks `uv` and Python with `ncly doctor transcript`, and previews a call with `--dry-run`. The dry run extracts the embedded Python engine into Nicely's cache and runs it through the real `uv`. `uv` fills its empty cache from the network. It sends nothing to YouTube or Deepgram. `just verify prepare "$run" transcript` links the real `uv` and Python into the run and puts them first in the wrapper's `PATH`.

## Sub-features

- `transcript-install` installs `ncly` and `ncly-transcript` with `just install` in the scratch home, and the wrapper then runs the installed `ncly`
- `transcript-discovery` lists `transcript` under the extensions of `ncly --help`, from the scratch data folder
- `transcript-help` shows the help of `ncly transcript run youtube` from the installed manifest
- `transcript-doctor` passes `transcript.uv` and `transcript.python3`, the programs that the manifest requires
- `transcript-skill` serves the skill installed beside the extension
- `transcript-refusals` refuses a URL that names no single video and a missing output folder with `USAGE_INVALID`, before the engine is extracted
- `transcript-dry-run` runs the engine through `uv` from an empty uv cache and answers the plan as JSON and as text. It creates no record and no result folder

## How to get to it (user POV)

- `just install`, from a checkout
- `ncly --help` and `ncly transcript run youtube --help`
- `ncly doctor transcript`
- `ncly skill view transcript`
- `ncly transcript run youtube --url <url> --output <folder> --dry-run`, as text and as JSON

## Driving it with verify.sh

Preconditions:

- A new run from `just verify`, whose Doctor step passed
- `just` in `PATH`
- A `uv` and a Python 3.12 or later that answer under a scratch home, such as those of the system or of Homebrew
- The network reaches PyPI, since `uv` fills the empty cache of the scratch home

In `just verify step "$run" transcript <args>`, `transcript` names this feature, and `<args>` are the arguments of `ncly`. Quote each URL, since `?` is a pattern of the shell.

- **Install.** Run `just verify prepare "$run" transcript`. It exits 0. Its lines name `<run>/scratch/home/.local/bin/ncly` and `<run>/scratch/home/.local/share/nicely/extensions/ncly-transcript` as installed, then the `uv` and `python3` that it linked into `<run>/scratch/tools`. Report these lines beside the quote.
- **Discovery.** Run `just verify step "$run" transcript --help`. Exit 0, and the `EXTENSIONS` section lists `transcript:  Transcribe videos with Deepgram`.
- **Help.** Run `just verify step "$run" transcript transcript run youtube --help`. Exit 0. It shows `ncly transcript run youtube --url <url> [flags]` and the flags `--dry-run`, `--output`, `--run-id`, `--timeout`, and `--url`, and `--timeout` says `570s by default`. Its effects line reads `changes your files, uses the network, may bill`.
- **Programs.** Run `just verify step "$run" transcript doctor transcript --json`. Exit 0, and the checks `transcript.python3` and `transcript.uv` both `pass`, with messages that name the links of `<run>/scratch/tools`.
- **Skill.** Run `just verify step "$run" transcript skill view transcript`. Exit 0, and its first line names the folder `<run>/scratch/home/.local/share/nicely/extensions/ncly-transcript.skill`, followed by the skill `transcript`.
- **Output folder.** Run `mkdir "$run/scratch/home/videos"`, which the next steps give to `--output`.
- **Playlist.** Run `just verify step "$run" transcript transcript run youtube --url 'https://www.youtube.com/playlist?list=PL123' --output "$run/scratch/home/videos" --dry-run --json`. Exit 2, stdout is empty, and stderr holds the code `USAGE_INVALID` and the message `https://www.youtube.com/playlist?list=PL123 is no URL of one YouTube video.`
- **Missing folder.** Run `just verify step "$run" transcript transcript run youtube --url 'https://www.youtube.com/watch?v=EIEc43CxIvY' --output "$run/scratch/home/missing" --dry-run --json`. Exit 2, stdout is empty, and stderr holds the code `USAGE_INVALID` and the message `<run>/scratch/home/missing is no existing folder.`
- **Nothing extracted.** Run `ls "$run/scratch/home/.cache/nicely"`. It fails, because each refusal stopped before the engine was extracted. Report its output beside the quote.
- **Dry run.** Run `just verify step "$run" transcript transcript run youtube --url 'https://www.youtube.com/watch?v=EIEc43CxIvY' --output "$run/scratch/home/videos" --dry-run --json`. Exit 0, and stdout is one answer with `"dry_run":true`. The one item of `plan` has the `key` of the URL and the `output` `<run>/scratch/home/videos`. Its steps are `[{"id":"transcribe","action":"run"}]`, and its effects `network`, `paid`, and `user_writes` are all `true`. The one pending check is `auth.deepgram`, with the message `Authentication is checked only during execution.`
- **Engine.** Run `ls "$run/scratch/home/.cache/nicely/python"/*/`, then `ls "$run/scratch/home/.cache/nicely/uv"`. The first lists `deepgram.py`, `outputs.py`, `protocol.py`, `runtime.py`, `transcript.py`, `youtube.py`, and `ytdlp_arc.py`, in one folder named after a hash of 64 hexadecimal digits. The second lists the cache that `uv` filled for the dry run, such as `archive-v0` and `wheels-v7`. Report both outputs beside the quote.
- **No record.** Run `ls "$run/scratch/home/.local/state/nicely/runs"`, then `ls -A "$run/scratch/home/videos"`. The first fails and the second prints nothing, because a dry run records nothing and creates no result folder. Report both outputs beside the quote.
- **Dry run as text.** Run the call of Dry run without `--json`. Exit 0, and stdout holds two lines: `Would transcribe https://www.youtube.com/watch?v=EIEc43CxIvY into a new folder in <run>/scratch/home/videos.`, then `pending  auth.deepgram  Authentication is checked only during execution.`

## Gotchas

- Run `prepare` before any `transcript` step. Before it, the wrapper runs the build of `just verify`, without `ncly-transcript`. Its `PATH` may also start with a shim of `uv` that fails under the scratch home
- Give every call of `ncly transcript run youtube` here `--dry-run`: a call without it needs a Deepgram key, and only a fake one may ever reach this feature
- The dry run reaches PyPI through `uv` only. Nicely's cache in the scratch home holds the engine and the uv cache, and the cleanup removes both
- The wrapper keeps the caller's variables, so a `UV_OFFLINE` or `UV_INDEX_URL` of the caller reaches `uv`
