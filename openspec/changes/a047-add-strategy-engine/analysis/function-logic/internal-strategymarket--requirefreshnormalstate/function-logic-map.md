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
| B1 | exact AST `if` at source line 315: `if source == nil \|\| now.IsZero() {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B2 | exact AST `if` at source line 319: `if err != nil \|\| reading.Market != market \|\| reading.Symbol != symbol \|\|` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B3 | exact AST `if` at source line 324: `if age < 0 \|\| age > 30*time.Second {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B4 | exact AST `if` at source line 327: `if reading.State != StateNormal {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ReadSymbolState` | one authoritative read | propagates as typed refusal; no retry/fallback | AST |

## State mutations and fallbacks

- Read-only; no fallback to caller authority strings.

## Safety conclusion

- Safe edit boundary: exact response identity and enum source are mandatory.
- High-risk impact: yes — prevents trading on stale/wrong-symbol status.
