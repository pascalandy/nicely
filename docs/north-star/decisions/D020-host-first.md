# D020 Build the extension host before any domain

Decided 2026-10-05, revised 2026-10-09. The extension host comes first, in M01, then the SDK, then `skill` as the first bundled extension. `describe`, `doctor`, `auth`, and grants follow. `agent` comes before transcript, with its harnesses, and brings run records. Taps follow. Transcript then arrives through the official tap, in four milestones: transcription, summary, resume, and the rest. The real Python program joins the first complete transcript path.

**Why.** A domain built before the host would be rewritten as an extension. Each milestone adds one capability on top of proven layers. `agent run` is a one-step paid command, so it proves records before transcript's two steps and its Python protocol. Early integration still exposes ownership and protocol mismatches while they are cheap to fix.

**Rejected.** The first version of this entry, a day-one milestone with doctor, auth, skill, transcript, profiles, records, and resume in eleven tasks plus separate acceptance lists. Agents picked single tasks and left the rest, and nobody could follow what remained.
