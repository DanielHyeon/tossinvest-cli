# Function Logic Map: `validAcquireRequest`

- Source: `internal/strategyrouter/quota.go`
- Current-base source SHA-256: `f76c9e4fa25fc664efcc918ab5e1f588051dd5119149298778cf3d0a6d4a26c9`
- Signature: `validAcquireRequest(params=1, results=1)`
- Source range: `200:1`–`203:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `201:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| — | — | — | no branch node |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| validQuotaKey | 201:9 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validMarket | 201:39 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validHorizon | 201:70 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| boundedNonEmpty | 202:3 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| boundedNonEmpty | 202:73 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| request.ObservedAt.IsZero | 202:139 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
