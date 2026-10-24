# D048 Describe an extension in a manifest and catalogs beside it

Decided 2026-10-10. Keep a static manifest and separate catalogs beside each executable, named by [Extension identity](../contract/extension-identity.md). The manifest uses catalog IDs, shared command declarations, and declared error codes.

**Why.** Help and discovery must not execute code that can hang. Separate catalogs let translators edit text without command definitions. Catalog files already identify supported languages. Code declarations let core check exits and choose the leading error.

**Rejected.** Literal English or catalog tables in the manifest, executable metadata probes, and a duplicate `languages` key.
