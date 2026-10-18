# Retry safety

Exit 75 is a promise that the same invocation is safe to repeat. It requires all of the following:

1. The failure is temporary
2. No paid request has been started, including a request whose response was lost
3. Every effect already produced is either harmless to repeat or covered by a demonstrated idempotent operation

Apply this rule to the whole invocation, including successful batch items and writes that happened before a lock failed. A free operation is not necessarily safe to repeat. Unknown effects, non-repeatable writes, and any started paid request exclude 75. Such runtime failures exit 1 and retain evidence for inspection or a supported explicit recovery action. Signals retain 130 and 143.

In a recorded operation, the record settles conditions 2 and 3 for the steps that this invocation started: exit 75 requires every step whose intent this invocation persisted to declare the `none` or `repeatable` effect, as [Record format](record-format.md) defines. Such a step counts even when it then failed, because it may have made part of its effect, such as a reserved result folder before a refused upload. A step that an earlier attempt completed, and that this invocation only reuses, does not count, because a repeat reuses it again. The same call with the same `--run-id` after exit 75 therefore continues the run or runs it again, as [Run IDs](run-ids.md) defines.

An idempotence claim includes destinations, configuration, external actions, and the result of a second invocation. Merely obtaining the same final file contents does not prove that no duplicate action occurred. Each writing command states what a repeated invocation does.
