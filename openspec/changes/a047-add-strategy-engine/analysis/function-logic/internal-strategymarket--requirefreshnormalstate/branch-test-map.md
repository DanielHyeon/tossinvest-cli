# Branch Test Map: `RequireFreshNormalState`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil/error/stale/HALT | `TestAuthoritativeSymbolStateMustBePresentFreshAndNormal` | existing | yes |
| B2 | wrong symbol/caller-claimed source | same test rows | yes | yes |
| B3 | exact official identity pass | same test pass case | yes | yes |
| B4 | non-NORMAL state refusal | same HALT row | existing | yes |
