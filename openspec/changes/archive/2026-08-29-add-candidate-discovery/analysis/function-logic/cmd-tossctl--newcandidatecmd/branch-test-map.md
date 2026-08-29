# Branch Test Map: `newCandidateCmd`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | (무분기) `tossctl candidate scan`/`watch`가 등록되고 둘 다 `mutating != true` | `TestTheDiscoveryCommandsDeclareThemselvesReadOnly` | yes (등록 전 `findCommandPath` FAIL) | yes |
