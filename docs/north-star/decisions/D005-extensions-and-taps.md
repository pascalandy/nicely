# D005 Add extensions the Git way and carry them in taps

Decided 2026-10-05, revised 2026-10-09. An executable named `ncly-<domain>` becomes `ncly <domain>` when its manifest declares a compatible protocol and its capabilities. A bundled extension is a Go package compiled into `ncly` with the same manifest, as [Extensions](../contract.md#extensions) defines. A tap is a Git repository that carries skills and extensions, like a Pi package. Core commands and bundled extensions always win over an external extension with the same name. Source precedence is explicit, and discovery reports the selected source. An extension receives the global flags as the variables in [contract.md](../contract.md#global-flags), and it prefixes its own error codes with its domain.

**Why.** A script can join Nicely in any language. Its manifest lets an agent inspect what it supports before running it. One private tap moves Pascal's tools to every machine. Git, `kubectl`, and cargo find plugins by executable name, while Nicely also needs to verify their shared output contract.

**Rejected.** A tap for skills only. Translating extension text from core, because core cannot know an extension's strings. Guessing capabilities from an executable's name or version. Accepting malformed protocol output as a successful result.
