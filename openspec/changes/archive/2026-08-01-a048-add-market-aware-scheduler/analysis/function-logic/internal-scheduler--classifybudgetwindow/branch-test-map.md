# Branch Test Map: `classifyBudgetWindow`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | missing trusted reset is initial | initial budget tests | existing | yes |
| B2 | kind mismatch is conflict | `TestResetWindowConflictingProvenanceFailsClosed` | existing | yes |
| B3 | kind switch selects epoch or delta semantics | epoch/delta tests | existing | yes |
| B4 | epoch case uses exact identity | `TestEpochResetWindowIdentityRemainsExact` | existing | yes |
| B5 | exact epoch anchor is same window | epoch identity test | existing | yes |
| B6 | later epoch after old boundary is next | generation tests | manual path could advance downstream | yes |
| B7 | delta case uses fixed tolerance | drift tests | existing | yes |
| B8 | delta after anchor+tolerance with later reset is next | `TestDeltaResetStartsNewGenerationOnlyAfterPriorBoundary` | existing | yes |
| B9 | inclusive ±1s delta bounds are same; extreme difference conflicts without overflow | boundary/MinInt-safe tests | `absDuration(MinInt)` overflowed | yes |
