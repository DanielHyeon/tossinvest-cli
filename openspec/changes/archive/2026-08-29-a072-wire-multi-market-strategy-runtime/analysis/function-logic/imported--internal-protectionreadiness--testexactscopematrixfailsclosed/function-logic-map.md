# Function Logic Map: `TestExactScopeMatrixFailsClosed`

- Source: `internal/protectionreadiness/attestation_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| signed KR fixture plus one mutated runtime field | one exact field differs from attested scope | signed attestation fixture | assert typed scope mismatch and UNWIRED |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | each account/profile/market/order/session/quantity/trigger/replace/tool/build/evidence/broker field substitution | test-only fixture mutation | `RefusalScopeMismatch` | table row subtest |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Assess` | project sealed market verdict | no retry; pure evaluation | CodeGraph + AST |

## State mutations and fallbacks

- No production mutation; each subtest creates a fresh signed fixture and mutates only its runtime scope.

## Safety conclusion

- Safe edit boundary: exact-scope matrix expectations only
- High-risk impact: no (test)
- High-risk impact: no (test)
