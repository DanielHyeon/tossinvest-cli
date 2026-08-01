# Branch Test Map: `dedupeKey`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | empty broker id is not a safe dedupe identity | existing missing-ID order behavior | reviewed | yes |
| B2 | exact scoped identities overlap; whitespace/market/day reuse does not | `TestOpenClosedDedupeUsesExactScopedOrderIdentity` | trimmed bare ID hid rows | yes |
| B3 | equal malformed identities dedupe, different malformed time stays distinct | `TestOpenClosedDedupeUsesExactScopedOrderIdentity` | trimmed bare ID hid rows | yes |
