# Function Logic Map: `Read`

- Source: `internal/app/engine/strategy_runtime_projection.go`
- Current-base source SHA-256: `6f8241d38004892cdf51bb71f597b368545e3d18e362ecd036e666462de42fce`
- Signature: `Context.Read(params=1, results=2)`
- Source range: `23:1`–`49:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `25:3, 31:3, 35:3, 48:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | if | 24:2 | planned targeted RED before any edit; not run by L0 |
| B2 | if | 30:2 | planned targeted RED before any edit; not run by L0 |
| B3 | if | 34:2 | planned targeted RED before any edit; not run by L0 |
| B4 | range | 37:2 | planned targeted RED before any edit; not run by L0 |
| B5 | if | 39:3 | planned targeted RED before any edit; not run by L0 |
| B6 | if | 43:3 | planned targeted RED before any edit; not run by L0 |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| errors.New | 25:41 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| c.strategyProjectionMu.RLock | 27:2 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| c.strategyProjectionMu.RUnlock | 29:2 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| errors.New | 31:41 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| store.Read | 33:19 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| supervisor.Snapshot | 38:17 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strategyprojection.Market | 42:23 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strategyprojection.WithMarketFailure | 44:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
