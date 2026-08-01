# Branch Test Map: `newestDashboardSourceTime`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | all trade-close/metric-observation timestamps participate; calculation time never enters the call | `TestDashboardNewestSourceAtDoesNotLaunderFreshnessWithCalculationTime` | helper absent at `948e721` | PASS |
| B2 | optional blank snapshot/metric times are ignored | blank optional timestamp test | helper absent at `948e721` | PASS |
| B3 | malformed persisted timestamp fails closed | corrupt source timestamp test | helper absent at `948e721` | PASS |
| B4 | newest timestamp wins and is normalized to UTC | ordering/timezone freshness test | helper absent at `948e721` | PASS |
