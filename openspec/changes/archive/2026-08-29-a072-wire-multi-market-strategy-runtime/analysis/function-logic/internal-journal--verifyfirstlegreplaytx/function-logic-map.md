# Function Logic Map: `verifyFirstLegReplayTx`

- Source: `internal/journal/strategy_first_leg_atomic.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| prepared replay | exact stored token and all cross-family identities | immutable journal rows | mismatch refuses without repair |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | general companion join/count differs | read only | replay mismatch | existing divergent replay suite |
| B2 | weekly lane companion missing/divergent or unexpected on non-weekly lane | read only | replay mismatch | paired weekly replay test |
| success | all exact rows match | read only | nil | exact replay suite |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| SQL verification joins | prove immutable first-leg and optional weekly companion | no repair/fallback | CodeGraph + AST |

## State mutations and fallbacks

- Read-only verification inside the caller transaction. Missing weekly evidence is never reconstructed.

## Safety conclusion

- Safe edit boundary: extend exact replay proof; preserve no-repair semantics.
- High-risk impact: yes/no
