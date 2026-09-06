# Branch Test Map: `insertExactStrategyDecision`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | each partial/malformed reference shape is refused in Go, before SQL | `TestStrategyEvidenceGoGuardRefusesBeforeSQL` | guard deleted → refusal arrives from the SQLite trigger instead, `errors.Is` fails | PASS |
| B2 | a trigger injected on `strategy_decision_lineage` makes the INSERT fail and the whole first leg rolls back | `TestFirstLegAtomicAdmissionLateStatementFailuresRollbackEveryFamily` `strategy_first_leg_atomic_test.go:263` | existing coverage (row corrected 2026-09-07: the previously cited test injects failure on `strategy_attempt_lineage` and never reaches this branch — issues.md I5) | PASS |
| B3 | a replay with a different consumed snapshot reference is a decision-lineage collision | `TestStrategyEvidenceLineageReplayIsExact` | existing coverage | PASS |
