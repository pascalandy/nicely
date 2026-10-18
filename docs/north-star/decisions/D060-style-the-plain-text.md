# D060 Style the text that a pipe gets, and let the terminal wrap it

Decided 2026-10-16. Style the same text that pipes receive, with 16 ANSI colors and weight only. Let the terminal wrap lines. [Output](../contract/output.md) owns exceptions and non-color meaning.

**Why.** One layout needs one translation and test. Terminal wrapping preserves copyable paths and commands. ANSI colors follow the user's theme.

**Rejected.** Separate terminal layouts, column wrapping that breaks paths, and fixed 256-color or true-color palettes that can fail on light themes.
