# Function Logic Map: `StrategyEntrySupervisor.Snapshot`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| supervisor/market | non-nil and exact `KR` or `US` | closed market enum | return `(zero,false)` |
| market runtime | one immutable market-owned record | constructor | return `(zero,false)` if absent |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | nil/invalid market | none | `(zero,false)` | invalid snapshot test |
| B2 | missing market runtime | none | `(zero,false)` | constructor invariant/static coverage |
| B3 | valid market | read lock only | immutable snapshot | dormant, fault and paired restart tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| none | snapshot copies locked in-memory state | no retry or mutation | AST |

## State mutations and fallbacks

- Snapshot is observation-only. New fields may expose first typed refusal and saturating restart metadata but grant no recovery or activation transition.

## Safety conclusion

- Safe edit boundary: copy-only additions from one market runtime.
- High-risk impact: no direct mutation; safety-relevant operational evidence.
