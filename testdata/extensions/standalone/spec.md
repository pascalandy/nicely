# standalone spec

`standalone` follows the [contract](../../../docs/north-star/contract.md) through the SDK. [Test kit](../../../docs/north-star/contract/test-kit.md) says what its test runs and checks.

## ncly standalone tally

Runs `tally-tool` without arguments, and answers `count`, the whole number that it prints. The number proves success. It has no effect of its own.

**Failures.**

| ID | Condition | Code |
|---|---|---|
| `tally-missing` | `tally-tool` is missing from `PATH`, which fails before anything starts | `PREREQ_MISSING` |
| `tally-fails` | An exit other than 0, or output that is no whole number | `RUNTIME` |

## ncly standalone fetch

Reads the key of the service `example` through `Call.Key`, and answers `sha256`, the SHA-256 of the key, never the key. The digest proves which key arrived.

**Failures.**

| ID | Condition | Code |
|---|---|---|
| `not-granted` | A key that no human granted | `KEY_NOT_GRANTED` |
| `key-missing` | A missing key | `AUTH_MISSING` |
| `no-keychain` | A keychain that does not answer | `KEYRING_UNAVAILABLE` |
