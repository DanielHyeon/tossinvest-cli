# Function Logic Map: `Journal.RecordStrategyDispatched`

- Source: `internal/journal/strategy_lineage.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| receipt | exact account/decision/risk/client/quantity/revision/state | committed strategy plan | reject stale/forged receipt |
| execution refs | non-empty core mutation and broker IDs | execgw Outcome | rollback on collision |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | receipt/ref invalid | none | error | terminal CAS tests |
| B2 | exact PLANNED or IN_DOUBT CAS succeeds | state revision increments | continue | direct/recovery tests |
| B3 | CAS misses but exact prior DISPATCHED exists | no state mutation | idempotent continuation | retry test |
| B4 | stale/different binding or execution ref collision | rollback | error | collision/CAS tests |
| Success | links exact | append MUTATION_ATTEMPT and BROKER_ORDER, commit | nil | recovery trace tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `verifyStrategyReceiptBindingTx` then state-only SQL CAS | exact immutable binding plus state/revision transition without naming immutable identity columns in UPDATE | one immediate transaction | AST |
| v14 unique indexes | one core attempt/order per strategy attempt | collision fails closed | schema tests |

## State mutations and fallbacks

- DISPATCHED never downgrades. IN_DOUBT can promote only with a current exact receipt.

## Safety conclusion

- Safe edit boundary: strategy terminal state and exact execution links only.
- High-risk impact: yes, records whether exposure may exist.
