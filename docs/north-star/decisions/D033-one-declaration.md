# D033 Describe a command once for humans and agents

Decided 2026-10-05, revised 2026-10-21. One [command declaration](../contract/command-descriptions.md) supplies help, completion, discovery, and doctor prerequisites. It also supplies these facts:

- What a dry run checks for its declared programs and keys. For a key, it checks the grant, from M19-T1, and never the keychain or the key's value
- The codes that a call can answer, from M19-T2
- The accepted values and the closest names in a usage error, from M19-T3
- The defaults that a handler reads, from M19-T4

No handler, extension spec, or skill restates a declared fact. The test kit checks each link that a machine can check, for every extension, from M19-T5. Declare only enforceable capabilities and add fields with their first consumers.

**Why.** Agents can build and parse a call without trial execution or unrelated domains. Shared declarations prevent conflicting promises. Discovery reports possible effects; preparation selects the invocation's effects. The agent drives of 2026-10-21 found surfaces that skip the declaration, such as a dry run that answered `ok` without a key. They also found facts copied by hand that drift, as [M19](../../milestones/M19-one-declaration.md#scope) measures.

**Rejected.** Separate agent catalogs, full-tree dumps for targeted lookups, and unknown capabilities treated as supported. Copies kept in sync by review. A keychain read in a dry run, which [Keys](../contract/keys.md) forbids.
