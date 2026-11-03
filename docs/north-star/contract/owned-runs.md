# Owned runs

This part says how the SDK builds the environment of a verification run, and how an action returns its result.
[D068](../decisions/D068-verification-map.md) puts this at level 1, with core as its owner.

## The constructor

`extensionrepo.NewOwnedRun` builds the environment and the folder of each run of the test kit, signoff, release, and acceptance.
It is the one function of `sdk/extensionrepo` and `sdk/extensiontest` that reads the environment of the caller as a whole.
The [analyzer](analyzers.md) `callerenv` refuses `os.Environ` and `syscall.Environ` everywhere else in these packages.

A run has one of two profiles:

- A host run keeps the tools and accounts of the caller, such as `PATH`, `HOME`, and the login of `gh`. Signoff, `go test` of `just check`, and the Git steps of a release use it
- A scratch run keeps no variable of the caller. The kit, acceptance, and the install check of a release use it

A host run drops this caller state:

- `JUST_*`, from a recipe of the caller, and `GIT_*`, from a Git hook
- The Nicely settings `NCLY_CONFIG`, `NCLY_LANG`, `NCLY_JSON`, `NCLY_NO_INPUT`, and `NCLY_VERBOSE`
- `NCLY_TEST_SCENARIO` and `NCLY_TEST_NOTICES` of an outer run
- `GOFLAGS` and `GOWORK`, which can select other source

A scratch run puts `HOME`, `TMPDIR`, and the XDG folders in its folder, and points D-Bus at no bus.
Its `PATH` starts with the folders of the run, then holds each absolute folder of the caller without an `ncly-*` file.
Replacing `HOME` alone is no isolation: at the parent of 09809608, a scratch `HOME` kept an invalid `NCLY_CONFIG` of the caller.

Each run masks `GOFLAGS` with a blank value, which also masks a value that `go env -w` keeps.
Each run also sets `GOWORK=off`.
The profile names the variables of the caller that the run keeps anyway, and the values that it sets:

| Run | Profile | Keeps | Sets |
|---|---|---|---|
| Kit scenario and lifecycle check | Scratch | `GOCOVERDIR`, `GORACE` | Sentinel keys |
| `go test` of `just check` and `just scenario` | Host | `GOFLAGS`, `GOWORK` | `NCLY_TEST_SCENARIO` for one scenario |
| Signoff | Host | Nothing more | Nothing more |
| Git and gh steps of a release | Host | `GIT_CONFIG_GLOBAL`, `GIT_CONFIG_NOSYSTEM` | Nothing more |
| Install check of a release | Scratch | `GIT_CONFIG_GLOBAL`, `GIT_CONFIG_NOSYSTEM` | Go settings, `NCLY_NO_INPUT=1`, `LC_ALL=C` |
| `just acceptance` | Scratch | Nothing | Go settings |

The Go settings are the caches, the module folder, and the module source of the caller.
They are `GOCACHE`, `GOMODCACHE`, `GOPATH`, `GOPROXY`, `GOPRIVATE`, `GONOPROXY`, `GONOSUMDB`, `GOSUMDB`, `GOINSECURE`, and `GOTOOLCHAIN`.
`go env` gives them before the scratch run starts.
A scratch run that uses programs of the caller links them into its `bin` folder first, through `mise which` when it answers.
A shim of a version manager reads its configuration under `HOME`, which the run moves.

`just acceptance` calls `go run ./internal/export acceptance [fixture]`.
That action runs `acceptance.sh` in its scratch run, with the programs that the manifest declares.

## The result

Each action of the export program returns one result to its caller.
The result carries the answer for stdout, the notices for stderr, and the exit code of the child that decided the action.
One writer delivers it after the action ends:

1. It writes the answer, such as `Signed off on <sha>`, to stdout.
2. It writes the notices, then any refusal of the action, to stderr.
3. It returns the exit code of the child, or 1 for a refusal.

A stream that refuses the answer or the notices fails the action with exit 1, because the caller never read them.
A host run with a folder gives its programs `NCLY_TEST_NOTICES`, so `extensionrepo.Notice` records each notice there.
The result reads them from the run.
At the parent of e35bd0bb, check ran `go test` outside that file, and a passing package hid its notices.
