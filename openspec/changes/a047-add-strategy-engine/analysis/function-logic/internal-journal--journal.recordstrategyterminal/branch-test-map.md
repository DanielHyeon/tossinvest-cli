# Branch Test Map: `Journal.recordStrategyTerminal`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | PLANNED to IN_DOUBT | terminal CAS test | missing v14 state | pass |
| B2 | stale second terminal transition | `TestStrategyTerminalStateIsCASAndDurable` | last-write risk | pass |
| B3 | zero recovery to REFUSED with durable reason | pending/recovery test | reason absent | pass |
| B4 | IN_DOUBT to REFUSED after definitive zero/refusal | recovery classification | transition missing | pass |
| B5 | refusal-history insert failure rolls back CAS | schema trigger/rollback tests | missing | pass |
| B6 | exact state+reason commit | recovery terminal table | missing | pass |
| B7 | immutable receipt binding precheck fails | stale/forged CAS tests | identity in UPDATE predicate | pass |
| B8 | source state/revision changed after receipt | terminal stale test | last-write risk | pass |
