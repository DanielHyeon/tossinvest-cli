# Function Logic Map: `effectiveStop`

- Source: `internal/weeklyvaluelane/evaluate.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| candidate | sealed digest/version, observed <= evaluated <= freshUntil | structural-stop port | STOP_INVALID |
| saved stop | positive canonical minor units | persisted common state | STOP_INVALID |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | candidate seal/freshness invalid | none | STOP_INVALID | stale/tamper test |
| B2 | candidate price invalid | none | STOP_INVALID | price test |
| B3 | no saved stop | none | candidate | happy path |
| B4 | saved stop tighter | none | saved stop | non-retreat test |
| B5 | candidate tighter | none | candidate | tighten test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| parseUnsigned | canonical price validation | fail closed | CodeGraph + AST |

## State mutations and fallbacks

- No mutation; monotonic maximum of saved and sealed candidate.

## Safety conclusion

- Safe edit boundary: pure stop selection only.
- High-risk impact: yes; stale/tamper/non-retreat tests required.
