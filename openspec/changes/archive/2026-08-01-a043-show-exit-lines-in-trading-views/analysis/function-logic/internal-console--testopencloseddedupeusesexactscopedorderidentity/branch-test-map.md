# Branch Test Map: `TestOpenClosedDedupeUsesExactScopedOrderIdentity`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | each exact/scoped/fallback collision variant | this table | bare trimmed ID collapsed all variants | yes |
| B2 | retained row count | this table | four distinct cases failed RED | yes |
| B3 | closed tally follows retained rows | this table | bare-id merge reported zero | yes |
