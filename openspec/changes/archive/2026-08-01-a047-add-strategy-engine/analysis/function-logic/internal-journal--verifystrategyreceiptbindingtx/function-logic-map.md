# Function Logic Map: `verifyStrategyReceiptBindingTx`

- Source: `internal/journal/strategy_lineage.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| receipt | attempt/account/decision/risk/client/quantity exact values | committed strategy plan or startup enumeration | mismatch error |
| transaction | active immediate journal transaction | terminal recorder | SQL error propagation |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | query error or any immutable binding differs | none | exact-binding error | terminal/recovery CAS tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| joined v14 SELECT | compare immutable attempt and decision-lineage quantity before state-only UPDATE | same transaction closes TOCTOU | AST + tests |

## State mutations and fallbacks

- Read-only helper; immutable identity columns never appear in an UPDATE statement.

## Safety conclusion

- Safe edit boundary: full receipt precondition for state transition.
- High-risk impact: yes, prevents forged or stale terminal transitions.
