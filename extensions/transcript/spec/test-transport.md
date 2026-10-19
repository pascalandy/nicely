# Test transport

The test transport lets the scenarios of transcript run [the engine](protocol.md#the-engine) without YouTube or Deepgram.
Only the test binary of `cmd/ncly` uses it.

`TestMain` in `cmd/ncly/script_test.go` registers `transcript-external`, as it registers `agent-external`.
It sets a package variable of the transcript extension, as it sets `record.Boundary`.
That variable points `uv` at a launcher file in `extensions/transcript/python/testing/`.
Release builds never embed that folder, and no build tag exists.
`uv` runs `--with-requirements <hash>/transcript.py <launcher> <hash>`, so the launcher imports the extracted engine.
The test binary passes the launcher as `NCLY_TEST_LAUNCHER`, and refuses to start the extension without it or the uv cache.
The launcher patches the downloader to copy a fixture audio file.
It patches `httpx.post` with an `httpx.MockTransport` that reads the whole upload, logs one line per attempt, and returns a fixture response.
It appends to `$WORK/starts.log`, `downloads.log`, and `uploads.log`, and never logs the key.
`NCLY_TEST_TITLE` sets the title of the video, and `NCLY_TEST_DEEPGRAM_STATUS` the HTTP status of the fixture response.
`NCLY_TEST_DOWNLOAD=temporary` or `failed` fails the download, and `NCLY_TEST_COOKIES=skipped` adds the cookie warning.
`NCLY_TEST_DEEPGRAM_BODY=garbage` answers 200 with no JSON. An empty variable counts as unset.
`NCLY_TEST_PAUSE=upload` writes the process IDs of Python and uv, then answers the upload once `$WORK/release` exists.
`NCLY_TEST_BOUNDARY` takes `<boundary>@<n>` to kill at the n-th time of a boundary, such as the intent of a second candidate.

The real-uv scenarios use a test-owned uv cache outside `$WORK`, which only the test binary sets.
The cache warms once with network, like Go modules, then the scenarios run offline.
The test binary keeps it in `nicely-test/uv` in the user cache folder, and passes it as `NCLY_TEST_UV_CACHE`.
It gives the scenarios the first `uv` and Python 3.12 in `PATH` that answer under a scratch `HOME`.
A shim that needs the real `HOME` never counts.
"Empty cache" in M13-T2 means Nicely's extraction cache; SP1 and Verify on the real binary prove the cold uv cache.

Spike SP2 ran on 2026-10-18 on Omarchy with uv 0.13.1, offline through `unshare -rn`.
Its two-file PEP 723 script imports its sibling, and its uv cache was warm.
The production line failed with exit 2, because `--upgrade-package yt-dlp` asks the package index for yt-dlp.
With `--offline`, the same line ran from the warm cache, so the test binary adds `--offline`.
After one warm run, the `python` recipe also passed offline, with or without Node on `PATH`.
Without Node, pyright's PyPI wrapper fetched Node 26.12.0 during the warm run.

## Fresh session

The test binary of `cmd/ncly` picks its command from its name, as `testscript.Main` does.
Linked as `ncly`, it runs core with the test keychain, but gives no extension the test transport.
Core finds transcript as `ncly-transcript` in the data folder, as in a release.
Only its `transcript-external` command carries the transport, with `NCLY_TEST_UV_CACHE`, `NCLY_TEST_LAUNCHER`, and `WORK` for the logs.

So the fresh session of M16-T2 gets a folder outside any checkout, which holds:

- An `ncly` that runs the test binary with a scratch home, its own XDG folders, and a fake `DEEPGRAM_API_KEY`
- That home's `ncly-transcript`, beside its manifest, catalogs, and skill. It sets the three variables, then runs `transcript-external`
- A copy of the launcher and its fixtures, and links to the real `uv` and Python
- A grant of `deepgram` to transcript

Nothing that the session runs reaches YouTube, Deepgram, or the OS keychain.
No source or spec lies in its folder, though the session can still read files elsewhere.
Its upload log is `$WORK/uploads.log`, and its evidence is fixture-backed.
