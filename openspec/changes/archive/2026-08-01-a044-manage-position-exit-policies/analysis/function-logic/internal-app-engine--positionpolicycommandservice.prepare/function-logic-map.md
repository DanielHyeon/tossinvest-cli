# Function Logic Map: `PositionPolicyCommandService.prepare`

- Source: `internal/app/engine/position_policy_command.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| request action | registered action/reason | server contract | invalid rejected |
| lifecycle provenance | external adoption required for release/readopt | journal state DTO | ineligible refused before price read |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B14 | time replacement, provenance, price/retry/stop/policy branches | no journal mutation | error or authoritative request | command service tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `PositionPolicy` | obtain typed provenance and eligibility | fail closed | AST |
| `Prices`/`Retrier` | derive READOPT t0 | bounded query semantics | AST |

## State mutations and fallbacks

- Client time and re-adoption observation are discarded; engine config/registry supply all derived values.

## Safety conclusion

- Safe edit boundary: release/readopt only external-adopted provenance; all re-adopt values engine-derived.
- High-risk impact: yes
