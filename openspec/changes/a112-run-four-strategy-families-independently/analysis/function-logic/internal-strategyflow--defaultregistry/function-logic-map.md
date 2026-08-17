# Function Logic Map: `defaultRegistry`

- Source: `internal/strategyflow/adapters.go`
- Current-base source SHA-256: `5fa16d5a0d6e2286f965c396e726fb3bf8cba48b32e4ec110359ff533def69d1`
- Signature: `defaultRegistry(params=0, results=1)`
- Source range: `13:1`–`22:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `14:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| — | — | — | no branch node |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| newRegistry | 14:9 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
