# Version pins

The repository is one Go module. Its `go.mod` pins one Nicely version, as [Compatibility](../compatibility.md) defines. It also repeats the `replace` of Lip Gloss from Nicely's `go.mod`, because a `replace` never reaches another module. That fork fixes a `tmux info` call as the package loads, until a release of Lip Gloss ships the fix. The template writes `go.mod` only at creation, so the repository owns it with its own dependencies.
