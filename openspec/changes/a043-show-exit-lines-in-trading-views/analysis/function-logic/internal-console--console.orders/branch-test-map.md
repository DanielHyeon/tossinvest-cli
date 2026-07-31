# Branch Test Map: `Console.orders`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | known OPEN list enters the pending path | existing OPEN-list tests | existing | yes |
| B2 | every OPEN row renders and receives exact-id evidence lookup | existing OPEN-list tests + `TestOrdersJoinExitEvidenceOnlyByAttemptIntentLineage` | yes | yes |
| B3 | non-empty OPEN broker IDs enter the duplicate set | existing duplicate/order group tests | existing | yes |
| B4 | known CLOSED list enters the finished path | existing CLOSED-list tests | existing | yes |
| B5 | every CLOSED row is considered | existing CLOSED-list tests | existing | yes |
| B6 | CLOSED row already present in OPEN is skipped | existing partial-filled duplicate test | existing | yes |
| B7 | known conditional list enters the conditional path | existing conditional tests | existing | yes |
| B8 | every conditional row remains visible | existing conditional tests | existing | yes |
| B9 | only explicit triggered/conditional IDs are considered | existing conditional tests + exact lineage fixture | yes | yes |
| B10 | exact conditional candidate link stops the candidate scan | exact-id lineage behavior covered by journal unit tests; console unlinked fixture covers miss | yes | yes |

The new leaf `attachOrderExitEvidence` covers linked, ambiguous, corrupt/legacy, and absent evidence. `TestOrdersJoinExitEvidenceOnlyByAttemptIntentLineage` supplies the positive exact lineage and same-symbol/time negative case; journal read-model tests cover ambiguity.
