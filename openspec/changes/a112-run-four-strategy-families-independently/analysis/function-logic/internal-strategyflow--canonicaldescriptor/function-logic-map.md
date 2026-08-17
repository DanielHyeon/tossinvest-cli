# Function Logic Map: `canonicalDescriptor`

- Source: `internal/strategyflow/registry.go`
- Current-base source SHA-256: `b188fe1dd7dfc1bc2f76b9905b8d461a6747156d5d3ed3e208f740dc79129e54`
- Signature: `canonicalDescriptor(params=1, results=2)`
- Source range: `109:1`–`117:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `113:4, 116:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | range | 111:2 | planned targeted RED before any edit; not run by L0 |
| B2 | if | 112:3 | planned targeted RED before any edit; not run by L0 |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| descriptorFor | 110:11 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| keyFor | 112:6 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| keyFor | 112:28 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
