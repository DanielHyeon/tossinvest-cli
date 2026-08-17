# Function Logic Map: `capabilitySeal`

- Source: `internal/strategyrouter/quota.go`
- Current-base source SHA-256: `f76c9e4fa25fc664efcc918ab5e1f588051dd5119149298778cf3d0a6d4a26c9`
- Signature: `capabilitySeal(params=1, results=1)`
- Source range: `96:1`–`105:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `104:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | range | 98:2 | planned targeted RED before any edit; not run by L0 |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| sha256.New | 97:7 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| string | 98:82 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| string | 98:109 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| string | 98:137 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| writeString | 99:3 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| writeUint64 | 101:2 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| copy | 103:2 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| h.Sum | 103:18 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
