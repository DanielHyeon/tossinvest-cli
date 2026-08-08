# Function Logic Map: `recordPolicySave`

- Source: `cmd/tossctl/operatingsettings.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| before/after | four trading-policy booleans | config policy | audit unavailable returns without changing policy |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | audit log unavailable | none | return | existing audit tests |
| B2-B3 | iterate fields; skip unchanged values | append one line per change | record errors intentionally ignored | existing policy audit tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `openAuditLog` | obtain operational audit sink | nil means no sink | CodeGraph + AST |
| `log.Record` | persist exact toggle delta | best-effort after config save | AST |

## State mutations and fallbacks

- This change only applies gofmt alignment; executable logic is unchanged.

## Safety conclusion

- Safe edit boundary: mechanical formatting only.
- High-risk impact: yes — trading-policy audit logic, verified unchanged by tests/AST.
