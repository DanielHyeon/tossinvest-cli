# Function Logic Map: `normalizeReconcileMarket`

- Source: `internal/journal/reconcile_states.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| symbol/market | empty global or uppercase KR/US with non-empty symbol | reconcile v24 contract | invalid request |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | market outside empty/KR/US | none | invalid request | validation test |
| B2 | market set without symbol | none | invalid request | validation test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| strings normalization | canonicalize caller input | deterministic | AST |

## State mutations and fallbacks

- Pure validation; no mutation or fallback.

## Safety conclusion

- Safe edit boundary: exact market enum and account-wide invariant.
- High-risk impact: yes — invalid scopes fail closed.
