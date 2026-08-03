# Branch Test Map: `loadRiskBucketState`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | active owner reads still delegate with released rows excluded | full journal risk-bucket replay suite | existing | GREEN |
| sibling | released receipt recomputes the same digest, then rejects late-fill drift | `TestRiskBucketLateFillCannotBindReopenedOwner`, `TestRiskBucketOwnerReleasedReplayRequiresExactSealedReceiptAndEvent` | released replay returned early without digest validation | GREEN |
