# Branch Test Map: `normalizeStrategyDecision`

| Branch | Scenario | Test | RED | GREEN |
|---|---|---|---|---|
| B1 | identifier whitespace is normalized | existing plan binding tests | existing | existing |
| B2 | zero timestamp uses journal time | production issuance test | existing | existing |
| B3 | supplied timestamp is UTC | replay tests | existing | existing |
| B4 | payload leading/trailing whitespace is not silently rewritten | `TestStrategyProductionIssuanceRejectsNonCanonicalDecisionPayload` | payload was trimmed and accepted | pass |
