# Function Logic Map: `ExitObserver.record`

- Source: `internal/app/engine/exitloop.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| managed position + exact a041 snapshot + observation metadata | projected quantity/action/provenance are from same immutable snapshot | a041 snapshot/a042 ledger specs | journal error prevents submit; Guardian remains sole authority |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B4 | proposal/orderability; clear working symbol; delay handling; deterministic intent id | may cancel collision before arm; an uncleared order sets typed arm suppression without mutating snapshot | error/hold | existing liquidation collision + typed arm-suppression tests |
| B5-B7 | construct judgement; atomic record; nil proposal | state/event/arm commit | error or state-only success | journal atomic tests |
| B8-B9 | proposal exists; submit | Guardian issuance and official gateway unchanged | existing result handling | E2E/emergency tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `RecordExitJudgement` | persist snapshot/event/arm | atomic local transaction | CodeGraph + AST |
| `submit` | issue reduction after durable arm | existing Guardian + official gateway | CodeGraph + AST |

## State mutations and fallbacks

- Adds exact snapshot plus source/time to judgement; no read-path recomputation.
- A failed working-order clear withholds only the arm and records `working_order_not_cleared`; the immutable orderable snapshot still advances high-water/protection.
- Arm-before-submit and zero-quantity state-only behavior remain unchanged.

## Safety conclusion

- Safe edit boundary: data handoff into existing ordering.
- High-risk impact: yes; emergency and fault-ordering tests required.
