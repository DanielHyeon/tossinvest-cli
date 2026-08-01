# Branch Test Map: `Console.orders`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | known OPEN list enters the pending path | existing OPEN-list tests | existing | yes |
| B2 | every OPEN row renders and receives exact-id evidence lookup | existing OPEN-list tests + `TestOrdersJoinExitEvidenceOnlyByAttemptIntentLineage` | yes | yes |
| B3 | non-empty OPEN broker identities enter the exact composite duplicate set | `TestOpenClosedDedupeUsesExactScopedOrderIdentity` | bare trimmed id hid distinct rows | yes |
| B4 | known CLOSED list enters the finished path | existing CLOSED-list tests | existing | yes |
| B5 | every CLOSED row is considered | existing CLOSED-list tests | existing | yes |
| B6 | only the exact scoped CLOSED identity already present in OPEN is skipped | existing partial-filled duplicate test + scoped collision table | market/day/whitespace reuse was hidden | yes |
| B7 | known conditional list enters the conditional path | existing conditional tests | existing | yes |
| B8 | every conditional row remains visible | existing conditional tests | existing | yes |
| B9 | enrich only post-filter rendered rows | filtered visible evidence tests | pre-filter scope overflow | yes |
| B10 | conditional row uses conditional origin semantics | existing conditional tests | yes | yes |
| B11 | plain row uses exact scoped origin/evidence | `TestOpaqueBrokerOrderIDsKeepDistinctOriginAndExitEvidence` | trimmed IDs cross-linked | yes |

The new leaf `attachOrderExitEvidence` covers linked, ambiguous, corrupt/legacy, and absent evidence. `TestOrdersJoinExitEvidenceOnlyByAttemptIntentLineage` supplies the positive exact lineage and same-symbol/time negative case; journal read-model tests cover ambiguity.
