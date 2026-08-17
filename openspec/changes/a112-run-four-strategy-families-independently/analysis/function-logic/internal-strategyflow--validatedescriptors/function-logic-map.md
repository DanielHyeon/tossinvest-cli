# Function Logic Map: `ValidateDescriptors`

- Source: `internal/strategyflow/registry.go`
- Current-base source SHA-256: `b188fe1dd7dfc1bc2f76b9905b8d461a6747156d5d3ed3e208f740dc79129e54`
- Signature: `ValidateDescriptors(params=1, results=1)`
- Source range: `25:1`–`42:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `27:3, 37:4, 41:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | if | 26:2 | planned targeted RED before any edit; not run by L0 |
| B2 | range | 30:2 | planned targeted RED before any edit; not run by L0 |
| B3 | range | 34:2 | planned targeted RED before any edit; not run by L0 |
| B4 | if | 36:3 | planned targeted RED before any edit; not run by L0 |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| len | 26:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| len | 26:25 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 27:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| len | 27:67 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| len | 27:85 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 29:14 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| len | 29:42 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 33:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| len | 33:32 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 37:11 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
