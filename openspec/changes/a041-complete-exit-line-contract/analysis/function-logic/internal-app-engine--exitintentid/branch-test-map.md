# Branch Test Map: `exitIntentID`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | valid decision becomes pending/mutation intent | concurrent engine/journal test | yes | yes |
| B2 | malformed decision uses existing random fallback | existing observer tests | existing | yes |
