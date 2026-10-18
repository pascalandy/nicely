# D011 Check locally by default, test live on request, and install nothing

Decided 2026-10-05, revised 2026-10-13. Doctor checks locally by default. `--live` uses free endpoints only. Doctor never installs; a human applies the fix in its hint. [Doctor](../core-spec.md#ncly-doctor) and [Required programs](../core-spec.md#required-programs) own the checks and install hints.

**Why.** A read-only report needs no dry run, confirmation, or sudo terminal. Static manifests provide checks without starting extensions. Only an extension knows which conditional programs its config selects.

**Rejected.** Installation inside doctor adds a writing mode for a command the human can already run. Generated install commands cannot assume identical package names or one package manager per OS. An unused conditional program must not fail the report.
