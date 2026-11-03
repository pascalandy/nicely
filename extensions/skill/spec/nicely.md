# The nicely skill

`ncly` bundles four skills that teach an agent to operate Nicely. `nicely` holds what every call needs, and names when to read each other skill:

| Skill | Teaches | Read |
|---|---|---|
| `nicely` | The variables, `ncly describe`, dry run first, codes, exit codes, `next`, doctor, the shell timeout, and extension skills | Before any `ncly` call |
| `nicely-auth` | `ncly auth`, and the rule to relay a missing key or grant to the human | For a key or a grant |
| `nicely-runs` | Run IDs, `ncly run`, and recovery after a lost answer | Before a recorded command |
| `nicely-extensions` | The install at a tag, the template command, and the removal | To install, build, or remove one |

A bundled skill has no folder for a relative link, so each topic is a skill of its own. One line in an agent's instructions points to `nicely`: `When asked to use nicely, run ncly skill view nicely`. The first example of `ncly --help` names it too. A skill of `[skill] paths` that takes a bundled name wins over it, with `SKILL_SHADOWED`.

The skills teach only what the release that ships them can do. Each item that adds a core or skill command adds its lines to the skill of its topic. `nicely-extensions` quotes the install rule of [Release and install](../../../docs/north-star/contract/extensions.md#release-and-install). Explicit resume remains in [COR-621](../../../docs/backlog/COR-621-resume-sdk.md#later). A skill writes each command in a code span, such as `ncly skill list`, so that a test reads it. A test fails when a bundled skill names an `ncly` command, or an error or warning code, that this `ncly` lacks. Each holds at most 60 lines, and `nicely` at most 2,000 bytes, because an agent reads it before every call.

Every external extension ships a skill named after its slug.
An agent reads `ncly skill view <slug>` for an extension root that it finds through describe. The `skill` extension ships no skill named `skill`: the bundled skills teach core and the skill domain together.
