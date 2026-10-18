# D059 Let a human type each answer, and show no button

Decided 2026-10-16. Every screen uses keyboard input without buttons. Confirmations use `[y/N]`; choices use keyboard-navigable lists. [Modes](../contract/modes.md) owns the look, including key grants.

**Why.** Pascal chose this in [M07](../../milestones/M07-forms.md). Huh's buttons showed focus by color alone, hiding the default under `NO_COLOR`. A text default survives color removal and screen readers, and the question and answer remain visible.

**Rejected.** Confirm buttons, even with No focused or a table of facts above them.
