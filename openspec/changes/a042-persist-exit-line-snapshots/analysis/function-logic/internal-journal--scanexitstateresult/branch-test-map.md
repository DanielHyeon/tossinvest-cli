# Branch Test Map: `scanExitStateResult`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | true all-NULL legacy tuple | `TestKnownLegacyIdentityIsResolvedInMemoryWithoutBackfill` | yes | yes |
| B2 | one v10 column survives while status is NULL | `TestLegacyDetectionRequiresEveryV10ColumnToBeNull` | yes | yes |
| B3 | each SEED output column appears alone | `TestSeedRejectsEveryPartialV10OutputColumn` | yes | yes |
| B4 | incomplete/forged EVALUATED tuple | corruption and forged-line tests | yes | yes |
| B5 | complete snapshot reopen | whole-candidate persistence and SIGKILL tests | yes | yes |
| B6 | nullable rung decoding | ladder persistence tests | yes | yes |
| B7 | nullable policy/generation/status decoding | migration/legacy tests | yes | yes |
| B8 | detect any v10 evidence | all-NULL and single-column tables | yes | yes |
| B9 | known legacy identity resolution | no-backfill test | yes | yes |
| B10 | ambiguous legacy identity stays unknown | legacy adoption tests | yes | yes |
| B11 | missing status with evidence | single-column evidence test | yes | yes |
| B12 | incomplete policy tuple | corruption matrix | yes | yes |
| B13 | invalid policy identity | corruption matrix | yes | yes |
| B14 | SEED output-evidence aggregate | per-column SEED table | yes | yes |
| B15 | SEED clean no-evaluation path | opening/read tests | yes | yes |
| B16 | EVALUATED completeness predicate | partial tuple tests | yes | yes |
| B17 | effective JSON semantic decode | forged digest/line tests | yes | yes |
| B18 | policy/position/generation identity equality | identity contract tests | yes | yes |
| B19 | snapshot/decision/observation identity equality | identity contract tests | yes | yes |
| B20 | flattened output/state equality | flattened corruption tests | yes | yes |
| B21 | source/time equality and final attach | persistence/reopen tests | yes | yes |
