# D059 Let a human type each answer, and show no button

Decided 2026-10-16. No screen of Nicely shows a button, such as Yes and No. A confirmation is a prompt that ends with `[y/N]`, and the human types the answer. A form takes typed values, and a choice among values is a list that the keyboard moves through. The rule covers every confirmation, including the grant of a key, and every screen that an extension builds from `sdk/tui`. [Modes](../contract.md#modes) holds the look.

**Why.** Pascal set this principle when he picked the looks of [M07](../../milestones/M07-forms.md). The mockup showed what buttons cost: Huh's confirm buttons mark the focused one by color alone, so under `NO_COLOR` nothing shows which answer Enter picks, on the step that hands a key to an executable. `[y/N]` writes the default as text, which survives no color and a screen reader, and Huh's accessible mode prints the same text. The question and its answer stay on screen as a record. It is also the plain prompt of [M06](../../milestones/M06-auth.md), which a human already knows.

**Rejected.** Huh's confirm field, with Yes and No buttons and No focused by default. The facts of the confirmation as a table above the same buttons. Both answer with one key, but the state of a button lives in its color.
