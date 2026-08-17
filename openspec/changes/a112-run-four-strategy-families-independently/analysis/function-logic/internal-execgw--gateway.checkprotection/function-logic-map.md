# Function Logic Map: `checkProtection`

- Source: `internal/execgw/protection.go`
- Current-base source SHA-256: `71e4923e1301555808b3c65b437d1d20906f9d633d8eef52ac676a1433cd8267`
- Signature: `Gateway.checkProtection(params=3, results=2)`
- Source range: `89:1`–`110:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `91:3, 94:3, 97:3, 101:3, 107:3, 109:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | if | 90:2 | planned targeted RED before any edit; not run by L0 |
| B2 | if | 93:2 | planned targeted RED before any edit; not run by L0 |
| B3 | if | 96:2 | planned targeted RED before any edit; not run by L0 |
| B4 | if | 100:2 | planned targeted RED before any edit; not run by L0 |
| B5 | if | 106:2 | planned targeted RED before any edit; not run by L0 |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| g.protectionCheckForTest | 94:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| protectionNotWired | 97:34 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| canonicalProtectionQuantity | 99:18 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| protectionNotWired | 101:34 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| g.protectionReadiness.Check | 103:25 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| g.clk.Now | 105:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| protectionNotWired | 107:34 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| refusal.Error | 107:59 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
