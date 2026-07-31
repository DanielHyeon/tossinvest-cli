# Branch Test Map: `Journal.recordExitJudgementTx`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | missing position | validation tests | yes | yes |
| B2 | invalid provenance | provenance tests | yes | yes |
| B3 | legacy suppression claim | `TestLegacyJudgementRejectsEveryArmSuppressionReason` | yes | yes |
| B4 | proposal present | proposal tests | yes | yes |
| B5 | invalid proposal | proposal tests | yes | yes |
| B6 | invalid proposal provenance | provenance tests | yes | yes |
| B7 | proposal/judgement provenance mismatch | provenance tests | yes | yes |
| B8 | recomputed snapshot present | snapshot persistence tests | yes | yes |
| B9 | invalid exact snapshot | recovery forgery tests | yes | yes |
| B10 | transaction begin failure | fault suite | yes | yes |
| B11 | current-state read failure | fault suite | yes | yes |
| B12 | completed state | completed-state test | yes | yes |
| B13 | generation mismatch | generation test | yes | yes |
| B14 | duplicate decision lookup | duplicate test | yes | yes |
| B15 | duplicate decision found | duplicate test | yes | yes |
| B16 | duplicate lookup error | storage fault suite | yes | yes |
| B17 | legacy monotonicity path | monotonicity tests | yes | yes |
| B18 | descending high water | monotonicity tests | yes | yes |
| B19 | descending protection | monotonicity tests | yes | yes |
| B20 | omitted level fallback | judgement tests | yes | yes |
| B21 | saved effective exists | recovery persistence tests | yes | yes |
| B22 | recomputed selection path | recovery tests | yes | yes |
| B23 | saved line available | stale recovery test | yes | yes |
| B24 | ambiguous recovery | quarantine test | yes | yes |
| B25 | quarantine write failure | fault suite | yes | yes |
| B26 | quarantine commit failure | fault suite | yes | yes |
| B27 | saved-monotone selected | `TestSavedMonotoneRecoveryCannotArmRecomputedOrder` | yes | yes |
| B28 | recomputed selected | first-evaluation test | yes | yes |
| B29 | saved candidate clears proposal/reason | saved no-arm test | yes | yes |
| B30 | effective snapshot encode | output digest tests | yes | yes |
| B31 | encode failure | forgery tests | yes | yes |
| B32 | state update failure | fault suite | yes | yes |
| B33 | after-state fault | fault suite | yes | yes |
| B34 | proposal arm | arm tests | yes | yes |
| B35 | arm failure | fault suite | yes | yes |
| B36 | after-arm fault | fault suite | yes | yes |
| B37 | evaluation event | persistence tests | yes | yes |
| B38 | event append failure | fault suite | yes | yes |
| B39 | after-event fault | fault suite | yes | yes |
| B40 | commit failure | crash/fault suite | yes | yes |
| B41 | post-commit result outcome | durable-result tests | yes | yes |
| B42 | saved-monotone result | saved no-arm test | yes | yes |
| B43 | armed result | proposal/E2E tests | yes | yes |
| B44 | working-order suppressed result | typed arm-suppression tests | yes | yes |
