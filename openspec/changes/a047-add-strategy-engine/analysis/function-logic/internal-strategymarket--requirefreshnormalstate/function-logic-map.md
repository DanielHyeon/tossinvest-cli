# Function Logic Map: `RequireFreshNormalState`

- Source: `internal/strategymarket/bars.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| source/request | non-nil and exact market/symbol request | strategy lane | unavailable refusal |
| reading | exact request identity, allowlisted official source, timestamp | source response | unavailable/stale refusal |
| state | NORMAL | official symbol-state endpoint | blocked refusal |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | source/clock absent | none | unavailable | proof test |
| B2 | error/identity/source/timestamp invalid | one read only | unavailable | wrong-symbol/caller-source rows |
| B3 | age outside 0..30s | none | stale | stale row |
| B4 | state non-NORMAL | none | blocked | HALT row |
| B5 | valid | none | opaque proof | pass row |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ReadSymbolState` | one authoritative read | propagates as typed refusal; no retry/fallback | AST |

## State mutations and fallbacks

- Read-only; no fallback to caller authority strings.

## Safety conclusion

- Safe edit boundary: exact response identity and enum source are mandatory.
- High-risk impact: yes — prevents trading on stale/wrong-symbol status.
