# D033 Describe a command once for humans and agents

Decided 2026-10-05. One [command declaration](../contract.md#command-descriptions) supplies help, completion, discovery, and doctor prerequisites. Declare only enforceable capabilities and add fields with their first consumers.

**Why.** Agents can build and parse a call without trial execution or unrelated domains. Shared declarations prevent conflicting promises. Discovery reports possible effects; preparation selects the invocation's effects.

**Rejected.** Separate agent catalogs, full-tree dumps for targeted lookups, and unknown capabilities treated as supported.
