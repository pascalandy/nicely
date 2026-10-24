# Named test kit fixture

This repository proves the test kit for a named extension with one executable root.
It uses only public SDK packages and the host from `go tool ncly`.

## ncly video convert

Return the input file name and the SHA-256 digest of the example service's key.
This fixture performs no conversion, file write, or network request.
Without a grant, the host returns `KEY_NOT_GRANTED` and names `video-convert` in the next command.
After `grant example`, the host passes the kit's sentinel key to the handler.
The direct executable uses that sentinel without a host grant.

The scenarios also prove scratch installation, slug-named sidecars, skill lookup, and missing-argument failures.
The skill may refer to the sibling `video subtitle` without declaring or installing it.
