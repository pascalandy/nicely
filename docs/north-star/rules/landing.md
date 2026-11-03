# Reviews, merges, stacks, UATs, and releases

This part of the [core rule index](README.md) holds D27 to D46 of `dev-preferences.md`.

| ID | Rule | Source | Enforced by |
|---|---|---|---|
| D27 | Open pull requests that are ready for review | [dev-preferences.md][dev] | Prose only |
| D28 | Get an independent review of the change, and quote its verdict in the pull request | [dev-preferences.md][dev] | Prose only |
| D29 | In brigade mode, the brigade decides each merge | [dev-preferences.md][dev] | Prose only |
| D30 | Otherwise merge only after signoff, with no open finding and with confidence, or when Pascal says so | [dev-preferences.md][dev] | Prose only |
| D31 | Report each merge, ask for no second confirmation, and stop only on a real blocker | [dev-preferences.md][dev] | Prose only |
| D32 | Stack dependent work: one pull request by default, and layers only for distinct reviews | [dev-preferences.md][dev] | Prose only |
| D33 | Fix each finding of a review on its branch, and reply with the commit that fixes it | [dev-preferences.md][dev] | Prose only |
| D34 | For one branch, rebase on main, push with lease, and sign off, then land the signed SHA | [dev-preferences.md][dev] | Prose only |
| D35 | Skip the hooks only for the documented checked pushes, and never land a branch name for the signed SHA | [dev-preferences.md][dev] | Prose only |
| D36 | After an unstack, no stack metadata stays | [dev-preferences.md][dev] | Prose only |
| D37 | Push each rebased layer with `--force-with-lease` | [dev-preferences.md][dev] | Prose only |
| D38 | The local, origin, and pull request SHAs of each layer are the same | [dev-preferences.md][dev] | Prose only; [check-layers.sh][layers] freezes only the origin SHAs |
| D39 | Sign off the layers bottom to top with their ancestry, and again for a new head | [dev-preferences.md][dev] | [check-layers.sh][layers], partial: a new run starts by hand |
| D40 | After a retarget, each pull request has main as its base, and names each Story with a closing keyword | [dev-preferences.md][dev] | Prose only |
| D41 | Land only the top SHA whose last signoff summary says ok | [dev-preferences.md][dev] | Prose only |
| D42 | Check that each pull request is MERGED, and never close it by hand instead | [dev-preferences.md][dev] | Prose only |
| D43 | Check that its pull request closed each Story | [dev-preferences.md][dev] | Prose only |
| D44 | Report a failed step of a landing as an incomplete merge | [dev-preferences.md][dev] | Prose only |
| D45 | Give each UAT its own issue with its task and procedure, and only link it from the Epic | [dev-preferences.md][dev] | Prose only |
| D46 | Release a shipped item or a fix of a released behavior unasked, with D043, D065, and release-nicely | [dev-preferences.md][dev] | Prose only |

[dev]: ../dev-preferences.md
[layers]: ../../../scripts/check-layers.sh
