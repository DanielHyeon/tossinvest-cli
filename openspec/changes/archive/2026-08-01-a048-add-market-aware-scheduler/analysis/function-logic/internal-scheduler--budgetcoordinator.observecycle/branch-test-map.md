# Branch Test Map: `BudgetCoordinator.ObserveCycle`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil/zero cycle refuses | observation-cycle binding test | missing API | yes |
| B2 | cross-key/coordinator/generation refuses without ingest | `TestObservationCycleIsOpaqueOneShotAndScopeBound` | no causal scope binding | yes |
| B3 | forged or replayed capability refuses | `TestObservationCycleIsOpaqueOneShotAndScopeBound` | no one-shot authority | yes |
| success | valid cycle ingests once and reconciles only its request-start watermark | held-response/manual/delta tests | wall timestamp was reconciliation authority | yes |
