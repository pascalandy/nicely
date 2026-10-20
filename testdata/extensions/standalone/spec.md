# standalone spec

`standalone` follows the [contract](../../../docs/north-star/contract.md) through the SDK. [Test kit](../../../docs/north-star/contract/test-kit.md) says what its test runs and checks.

## ncly standalone tally

Runs `tally-tool` without arguments, and answers `count`, the whole number that it prints. The number proves success. It has no effect of its own.

- `tally-tool` missing from `PATH` fails with `PREREQ_MISSING`, before anything starts
- An exit other than 0, or output that is no whole number, fails with `RUNTIME`

## ncly standalone fetch

Reads the key of the service `example` through `Call.Key`, and answers `sha256`, the SHA-256 of the key, never the key. The digest proves which key arrived.

- A key that no human granted fails with `KEY_NOT_GRANTED`, and a missing key with `AUTH_MISSING`
- A keychain that does not answer fails with `KEYRING_UNAVAILABLE`
