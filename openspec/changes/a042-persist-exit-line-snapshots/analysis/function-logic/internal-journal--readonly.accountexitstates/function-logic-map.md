# Function Logic Map: `ReadOnly.accountExitStates`

- Source: `internal/journal/account_views.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| account | completed and open states | journal columns only | wrapped SQL/semantic error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B4 | query, scan, rows error, map result | none | error/map | account-view tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `scanExitState` | exact persisted state | no recomputation | CodeGraph + AST |

## State mutations and fallbacks

- New typed snapshot projection is added beside this legacy helper.

## Safety conclusion

- Safe edit boundary: unchanged helper.
- High-risk impact: yes.
