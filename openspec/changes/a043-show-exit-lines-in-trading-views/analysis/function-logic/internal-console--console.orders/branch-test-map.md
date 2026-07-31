# Branch Test Map: `Console.orders`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | known OPEN list renders all pending rows | existing `Test...Open...` order tests | existing | existing |
| B2 | CLOSED duplicate is suppressed by exact broker id | existing duplicate/order group tests | existing | existing |
| B3 | conditional row remains unlinked without exact attempt chain | `TestOrdersJoinExitEvidenceOnlyByAttemptDecisionID` | no | no |
| B4 | truncation remains a lower bound | existing pagination tests | existing | existing |
| B5 | server link filters remain local | existing filter matrix | existing | existing |
| B6 | exact attempt decision id renders trigger snapshot/provenance | `TestOrdersJoinExitEvidenceOnlyByAttemptDecisionID` | no | no |
| B7 | same symbol/time but no exact decision chain stays `근거 미연결` | `TestOrdersNeverFuzzyJoinExitEvidence` | no | no |
| B8 | each conditional row remains visible and unlinked unless an exact broker-order id is proven | existing conditional tests + `TestOrdersNeverFuzzyJoinExitEvidence` | no | no |
