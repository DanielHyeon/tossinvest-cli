# Branch Test Map: `ParseRateBudgetReset`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | empty/whitespace raw is absent | header/reset parser tests | existing | yes |
| B2 | malformed, negative, or zero observed-at is unparsed | implausible/parser table | unsafe manual construction possible | yes |
| B3 | exactly 1,000,000,000 is epoch | `TestParseRateBudgetResetUsesExactThresholdAndPlausibilityBoundaries` | scheduler duplicated kind rules | yes |
| B4 | delta conversion rejects overflow/wrap evidence | official parser table + scheduler wrapping case | huge seconds could wrap to +1s in scheduler | yes |
| B5 | inclusive -1m/+24h accepted; one second outside rejected | official parser table | scheduler omitted plausibility bounds | yes |
| B6 | final plausibility rejection clears the derived instant but preserves raw | official parser/implausible tables | downstream accepted implausible constructed budgets | yes |
