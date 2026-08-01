# Function Logic Map: `Journal.recordStrategyTerminal`

- Source: `internal/journal/strategy_lineage.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| receipt | full exact plan binding and current revision/state | committed plan/startup enumeration | stale error |
| target/reason | REFUSED or IN_DOUBT, non-empty stable reason | dispatch/recovery classifier | invalid error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | invalid target/source/binding | none | error | CAS tests |
| B2 | exact CAS misses | none | stale revision error | stale terminal test |
| Success | exact CAS succeeds | state+revision update and append-only reason in one tx | nil | recovery/terminal tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `verifyStrategyReceiptBindingTx` then state-only SQL CAS | bind account/decision/risk/client/quantity before updating only state/revision | rollback on mismatch | AST |
| refusal insert | preserve terminal reason history | same transaction | schema triggers/tests |

## State mutations and fallbacks

- PLANNED may become REFUSED/IN_DOUBT; IN_DOUBT may become REFUSED only with current receipt.

## Safety conclusion

- Safe edit boundary: exact strategy attempt terminal transitions.
- High-risk impact: yes, terminal truth must be durable and monotonic.
