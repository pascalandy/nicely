# Deferred verification work

This part of [COR-104](../COR-104-verification-map.md) keeps its deferred ideas and the condition that brings back each one.

## Unprivileged scratch runs

Tests, acceptance, and reviews run in a new scratch folder, with no root, mount, or system service.
Two template rules stay `Prose only` until then.

Revisit condition: a review stalls on a recursive delete or mounts a file system again, as in #425.

## Process lifetime in the runner

`sdk/runner` makes the output of each program an explicit choice, and the COR-104-T2 rule on `io.Discard` goes.

Revisit condition: a new correction of exit, signal, drain, or cleanup, class K04 of [N1][n1].

[n1]: https://github.com/pascalandy/nicely-research/blob/main/synthesis/verification-map/analysis/N1-correct.md
