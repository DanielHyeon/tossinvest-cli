# Branch Test Map: `Store.Collect`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | validation failure writes nothing | model/store validation tests | existing | existing |
| B2 | first collection appends complete set | `TestCollectAppendsExistingObservationsAndLatestMeasurement` | existing | existing |
| B3 | exact replay skips all immutable rows | `TestCollectExactReplayIsIdempotentAcrossRestartAndConcurrency` | yes | yes |
| B4 | divergent bytes fail closed | `TestCollectDivergentReplayFailsClosed` | yes | yes |
| B5 | crash phases expose zero or complete transaction | `TestPerformanceMigrationAndAppendSIGKILLPhasesAreAllOrNone` | yes | yes |
| B6 | transaction begin failure returns | DB/context error contract | no — defensive | yes |
| B7 | trade compare conflict rolls back | divergent replay test | yes | yes |
| B8 | observation compare conflict rolls back | divergent replay test | yes | yes |
| B9 | snapshot compare conflict rolls back | divergent snapshot test | yes | yes |
| B10 | commit/SIGKILL is all-or-none | phase SIGKILL test | yes | yes |
