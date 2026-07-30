# Branch Test Map: `inspectOpen`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | File metadata cannot be read | existing invalid-file tests | baseline | pass |
| B2 | Non-regular file is refused | existing candidate-type tests | baseline | pass |
| B3 | File without an executable bit is refused | existing permission test | baseline | pass |
| B4 | Rewind before hashing fails | descriptor failure test | baseline | pass |
| B5 | Reading candidate bytes for SHA-256 fails | descriptor failure test | baseline | pass |
| B6 | Go build information is absent or unreadable | existing foreign-binary test | baseline | pass |
| B7 | Go module path does not match tossctl | existing wrong-module test | baseline | pass |
| B8 | Go command path does not match tossctl | existing wrong-command test | baseline | pass |
| B9 | Every Go build setting is collected without executing the candidate | `TestInspectReportsVCSBuildSettings` | VCS identity unavailable | pass |
| B10 | GOOS/GOARCH differs from the running platform | existing wrong-platform test | baseline | pass |
