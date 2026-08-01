# Branch Test Map: `dashboardSQL`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | happy path preserves bounded indexed query plan and projects persisted metric observation time without calculation-time freshness | million-row query-plan test, `TestDashboardNewestSourceAtDoesNotLaunderFreshnessWithCalculationTime` | observation timestamp absent at `948e721` | PASS |
