# Branch Test Map: `freezeTradeOutcomeTx`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` at line 131: `if !ok {`; invariant: missing/corrupt/alternate path is explicit | `TestTheOutcomeIsFrozenInTheClosingTransaction`, `TestABackfillRecoversTheGapAndRefusesToRewriteIt`, `TestAnAnalyticsFailureDoesNotRollBackTheClose` | not separately captured for this evidence refresh | package-targeted regression PASS before integration; rerun by gate |
