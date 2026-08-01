# Function Logic Map: `RequireNoPosition`

- Source: `internal/strategymarket/bars.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| source/request | non-nil canonical exact market/symbol | strategy lane | unavailable refusal |
| reading | exact identity, allowlisted official position source, fresh timestamp | source response | unavailable/stale refusal |
| exposure | canonical exact quantity `0`, no open orders | official account read | blocked refusal |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | source/request absent | none | unavailable | proof test |
| B2 | error/identity/source/time absent | one read only | unavailable | wrong-symbol/caller-source rows |
| B3 | stale/future | exact local time | stale | stale row |
| B4 | quantity invalid/nonzero or orders exist | exact decimal parse | blocked | exposure rows |
| B5 | valid zero exposure | none | opaque proof | pass row |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ReadPosition` | one authoritative read | typed refusal; no retry/fallback | AST |
| `exactDecimal` | reject non-canonical quantity assertion | pure | AST |

## State mutations and fallbacks

- Read-only; exact request identity is checked after the source call.

## Safety conclusion

- Safe edit boundary: caller-provided authority strings cannot mint a proof.
- High-risk impact: yes — duplicate-position/order prevention.
