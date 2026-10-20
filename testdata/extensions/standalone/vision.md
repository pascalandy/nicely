# standalone vision

`standalone` is the fixture extension repository of Nicely's test kit. It proves that an extension outside Nicely tests itself with `go test ./...`.

- It wraps `tally-tool`, which the scenarios replace with the fake child of the test kit. No real program has that name
- It refuses every real service: its key belongs to the test service `example`, and its scenarios hold only sentinel keys
