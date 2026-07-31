# Branch Test Map: `addResetDelta`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | invalid duration seconds fail closed | parser boundary tests | duplicated scheduler multiply wrapped | yes |
| B2 | add/sub mismatch fails closed | extreme arithmetic coverage | unsafe add could forge a short reset | yes |
| B3 | valid whole-second delta preserves observed nanoseconds | reset parser and existing delta tests | existing | yes |
