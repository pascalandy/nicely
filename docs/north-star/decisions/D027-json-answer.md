# D027 Answer in one JSON line with `ok` and `errors`

Decided 2026-10-05. With `--json`, use one JSON object per line with `ok` and, on failure, `errors`. [Output](../contract.md#output) defines exceptions, report streams, batch results, and leading-error order.

**Why.** `errors` follows Pascal's script-output convention. `ok` survives merged streams, and one line permits `tail -n1 | jq`. Reports stay on stdout even when checks fail, as established report CLIs do. A batch keeps item causes separate from the overall retry verdict.

**Rejected.** Flat or singular errors without `ok`, JSON by default, and reports on stderr. The first observed failure cannot authorize a whole-batch retry.
