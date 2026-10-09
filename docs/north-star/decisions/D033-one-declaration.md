# D033 Describe a command once for humans and agents

Decided 2026-10-05. One [command declaration](../contract.md#command-descriptions) supplies help, completion, targeted JSON discovery, and doctor prerequisites. It describes inputs, results, effects, and supported capabilities. Adapters and extensions declare only guarantees they can enforce.

**Why.** An agent needs the relevant command's contract without loading every domain or probing by trial and error. Required flags, flag groups, and typed result schemas make that contract sufficient to build and parse a call. Shared declarations keep help, preflight checks, and execution from making different promises. Discovery describes possible effects and requirements; shared preparation selects those of the invocation. M00's declaration grows with its first consumers, rather than pretending every field already exists. An extension's manifest holds the same declarations.

**Rejected.** Separate manually maintained command catalogs for agents. Dumping the whole command tree for every lookup. Treating an unknown capability as supported.
