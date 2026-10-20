# Compatibility

Within a contract version, command names, flags, error codes, exit codes, and the meaning of existing JSON remain stable. Compatibility covers key names, types, units, formats, required fields, absence versus `null`, enum values, and array ordering where specified. Object key order and translated human messages are not machine contracts.

- Readers ignore unknown optional object fields. Adding an optional field must preserve the behavior of an existing consumer
- Enums are closed unless their definition explicitly permits new values and specifies how readers handle an unknown value
- A breaking change needs a new contract version, a [decision](../decisions/README.md), and a documented migration or continued support for the old contract before release. A decision alone does not make a change compatible
- Internal Go/Python messages, extension manifests, and stored run records declare their own format versions. Readers reject unsupported versions before work instead of interpreting them as a known format

`NCLY_VERSION` remains the software release version. `NCLY_CONTRACT_VERSION` carries the public contract version to child programs. A command or mode appears in discovery only when the installed version supports it.

An extension has three version facts:

- The contract version that it speaks, which its manifest declares as `contract_version`
- Its own version, which `ncly-<domain> --version` prints from the module that built it
- The Nicely version in its `go.mod`, which pins the SDK. From M18-T3, it also pins the `ncly` that its tests run, and from M18-T5 the extension template
