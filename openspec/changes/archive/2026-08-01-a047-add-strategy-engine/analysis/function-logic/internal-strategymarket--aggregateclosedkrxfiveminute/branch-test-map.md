# Branch Test Map: `AggregateClosedKRXFiveMinute`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | branchless happy-path sentinel (AST has no conditional): every legacy raw-slice call returns `RefusalSource` and a zero `VerifiedBar` | `TestLegacyRawSliceCannotMintVerifiedBar` | yes | yes |
| Scenario | opaque official-page identity/refusal checks are implemented by `SealOfficialClosedKRXFiveMinuteFor`, not this legacy function | `TestAggregateClosedKRXFiveMinuteFailsClosed` | yes | yes |
| Scenario | exact official-page decimal aggregation is implemented by the unexported aggregator, not this legacy function | `TestAggregateClosedKRXFiveMinutePreservesExactDecimals` | yes | yes |
