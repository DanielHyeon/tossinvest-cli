# Branch Test Map: `Journal.PlanCampaignLeg`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | retry/stale version | `TestCampaignCommandsExpectedVersionAndDeterministicRetry` | yes | yes |
| B2 | CLOSING Position | `TestExposureAdmissionReadsPositionAndPendingRiskReduction/position_closing` | yes | yes |
| B3 | unresolved SELL | `TestExposureAdmissionReadsPositionAndPendingRiskReduction/pending_sell_intent` | yes | yes |
| B4 | zero quantity | `TestCampaignPlanAndOrderCapRejectZeroQuantity` | yes | yes |
| B5 | request validation group | plan tests | yes | yes |
| B6 | request validation group | plan tests | yes | yes |
| B7 | transaction start | plan tests | yes | yes |
| B8 | retry lookup | deterministic retry test | yes | yes |
| B9 | retry read | deterministic retry test | yes | yes |
| B10 | retry commit | deterministic retry test | yes | yes |
| B11 | campaign header | plan tests | yes | yes |
| B12 | expected version | stale-version test | yes | yes |
| B13 | admission query | EXIT FIRST tests | yes | yes |
| B14 | admission result | EXIT FIRST tests | yes | yes |
| B15 | campaign block | transition tests | yes | yes |
| B16 | D4 transition | transition tests | yes | yes |
| B17 | leg count query | sequence tests | yes | yes |
| B18 | contiguous sequence | sequence tests | yes | yes |
| B19 | leg insert | plan tests | yes | yes |
| B20 | campaign update | plan tests | yes | yes |
| B21 | command append | replay tests | yes | yes |
| B22 | event append | replay tests | yes | yes |
| B23 | commit/final read | plan tests | yes | yes |
