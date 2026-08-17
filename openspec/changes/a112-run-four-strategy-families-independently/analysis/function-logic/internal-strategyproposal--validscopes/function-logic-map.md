# Function Logic Map: `validScopes`

- Source: `internal/strategyproposal/production.go`
- Current-base source SHA-256: `6cc7474d631e24c1daee677743fdbcc942787e9ae6874ed318cd3550326803b3`
- Signature: `validScopes(params=2, results=1)`
- Source range: `468:1`–`480:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `470:3, 475:4, 479:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | if | 469:2 | planned targeted RED before any edit; not run by L0 |
| B2 | range | 473:2 | planned targeted RED before any edit; not run by L0 |
| B3 | if | 474:3 | planned targeted RED before any edit; not run by L0 |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| len | 469:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| len | 469:25 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strings.ToUpper | 474:44 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strings.TrimSpace | 474:60 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| laneMatchesMarket | 474:345 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
