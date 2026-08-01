# Branch Test Map: `validResetSemantics`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | exact official delta/epoch derivation is accepted | reset semantics boundary table | existing happy path | yes |
| B2 | wrapping delta raw, exact threshold mislabel, >24h delta, implausible epoch, and derived-instant mismatch are rejected | reset semantics boundary table | duplicated parser accepted some forged budgets | yes |
