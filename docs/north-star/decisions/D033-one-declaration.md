# D033 Describe a command once for humans and agents

Decided 2026-10-05, revised 2026-10-21. One [command declaration](../contract/command-descriptions.md) supplies help, completion, discovery, and doctor prerequisites. It also supplies these facts:

- The grant check of a dry run
- The codes that `ncly describe` lists
- The accepted values in a usage error
- The defaults that a handler gets

No handler, spec, or skill copies a declared value, such as a default, the accepted values, or the list of codes. Declare only enforceable capabilities and add fields with their first consumers.

**Why.** Agents can build and parse a call without trial execution or unrelated domains. Shared declarations prevent conflicting promises. Discovery reports possible effects; preparation selects the invocation's effects. The agent drives of 2026-10-21 found surfaces that skip the declaration, such as a dry run that answered `ok` without a grant. They also found defaults copied by hand, which drift.

**Rejected.** Separate agent catalogs, full-tree dumps for targeted lookups, and unknown capabilities treated as supported. Copies kept in sync by review. A keychain read in a dry run, which [Keys](../contract/keys.md) forbids. A list of codes for each command, which needs a new field that drifts.
